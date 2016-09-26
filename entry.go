package logrusOVH

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"crypto/rand"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"math"

	"zombiezen.com/go/capnproto2"

	"net"
	"time"

	"os"

	"github.com/Sirupsen/logrus"
	"github.com/satori/go.uuid"
)

// Entry represent the base entry
type Entry struct {
	entry    *logrus.Entry
	ovhToken string
}

type entryGelf struct {
	OvhToken     string  `json:"_X-OVH-TOKEN"`
	Version      string  `json:"version"`
	Host         string  `json:"host"`
	ShortMessage string  `json:"short_message"`
	FullMessage  string  `json:"full_message"`
	Timestamp    float64 `json:"timestamp"`
	Level        uint8   `json:"level"`
}

func (e Entry) send(proto Protocol, compression CompressAlgo) (err error) {
	var data []byte

	switch proto {
	case GELFTCP, GELFUDP, GELFTLS:
		if data, err = e.gelf(compression); err != nil {
			return
		}
	case CAPNPROTOTCP, CAPNPROTOTLS:
		if data, err = e.capnproto(compression == COMPRESSPACKNPPACKED); err != nil {
			return
		}
	default:
		return fmt.Errorf("%v not implemented or not supported", proto)
	}
	conn, err := getConn(proto)
	if err != nil {
		return err
	}
	defer conn.Close()
	switch proto {
	case GELFTCP, GELFTLS, CAPNPROTOTCP, CAPNPROTOTLS:
		n, err := conn.Write(data)
		if err != nil {
			return err
		}
		if n != len(data) {
			return fmt.Errorf("entry not completely sent %d/%d", n, len(data))
		}
	case GELFUDP:
		if len(data) < UDP_CHUNK_MAX_SIZE {
			n, err := conn.Write(data)
			if err != nil {
				return err
			}
			if n != len(data) {
				return fmt.Errorf("entry not completely sent %d/%d", n, len(data))
			}
		} else {
			// chunk buffer
			chunkBuf := bytes.NewBuffer(nil)
			// data buffer
			dataBuf := bytes.NewBuffer(data)

			// nb chunck
			nbChunks := int(math.Ceil(float64(len(data)/UDP_CHUNK_MAX_DATA_SIZE))) + 1

			// MSG ID
			msgID := make([]byte, 8)
			n, err := io.ReadFull(rand.Reader, msgID)
			if err != nil || n != 8 {
				return fmt.Errorf("unable to generate msgID, %v", err)
			}

			for i := 0; i < nbChunks; i++ {
				chunkBuf.Write(GELF_CHUNK_MAGIC_BYTES)
				chunkBuf.Write(msgID)
				chunkBuf.WriteByte(byte(i))
				chunkBuf.WriteByte(byte(nbChunks))
				for j := 0; j < UDP_CHUNK_MAX_DATA_SIZE; j++ {
					b, err := dataBuf.ReadByte()
					if err != nil {
						if err == io.EOF {
							//log.Println("EOF", dataBuf.Bytes())
							break
						}
						return fmt.Errorf("unable to read from dataBuff, %v", err)
					}
					err = chunkBuf.WriteByte(b)
					if err != nil {
						return fmt.Errorf("unable to write to chunk buffer %v", err)
					}
				}
				// write data
				n, err := conn.Write(chunkBuf.Bytes())
				if err != nil {
					return err
				}
				if n != len(chunkBuf.Bytes()) {
					return fmt.Errorf("entry not completely sent %d/%d", n, len(chunkBuf.Bytes()))
				}

				// reset chunk buffer
				chunkBuf.Reset()
			}
		}
	}
	return nil
}

// Serialize entry for Gelf Proto
func (e Entry) gelf(compression CompressAlgo) (out []byte, err error) {
	g := entryGelf{
		OvhToken:    e.ovhToken,
		Version:     "1.1",
		FullMessage: e.entry.Message,
		Timestamp:   float64(time.Now().UnixNano()/1000000) / 1000.,
		Level:       uint8(e.entry.Level),
	}

	// host
	g.Host, err = os.Hostname()
	if err != nil {
		g.Host = "undefined"
	}

	// short message
	if len(g.FullMessage) > 80 {
		g.ShortMessage = g.FullMessage[0:80] + "..."
	} else {
		g.ShortMessage = g.FullMessage
	}
	out, err = json.Marshal(g)
	if err != nil {
		return []byte{}, fmt.Errorf("Failed to marshal gelfEntry to JSON, %v", err)
	}

	// From logrus
	if len(e.entry.Data) > 0 {
		// remove trailing }
		out = out[0 : len(out)-1]
		data := make(logrus.Fields, len(e.entry.Data)+3)
		for k, v := range e.entry.Data {
			switch v := v.(type) {
			case error:
				// Otherwise errors are ignored by `encoding/json`
				// https://github.com/Sirupsen/logrus/issues/137
				data["_"+k] = v.Error()
			default:
				data["_"+k] = v
			}
		}
		serialized, err := json.Marshal(data)
		if err != nil {
			return []byte{}, fmt.Errorf("Failed to marshal e.entry.Data to JSON, %v", err)
		}
		out = append(out, 44)
		out = append(out, serialized[1:]...)
	}

	// Compress ?
	if compression != COMPRESSNONE {
		var b bytes.Buffer
		switch compression {
		case COMPRESSGZIP:
			w := gzip.NewWriter(&b)
			w.Write(out)
			w.Close()
		case COMPRESSZLIB:
			w := zlib.NewWriter(&b)
			w.Write(out)
			w.Close()
		default:
			return []byte{}, fmt.Errorf("%v compression not supported", compression)
		}
		out = b.Bytes()
	}

	return out, nil
}

//return, if exist, value assoaciated with key key
func (e Entry) getCapnpFieldValue(key, expectedType string, ignoreDataKey []string) (found bool, value interface{}) {
	if _, found := e.entry.Data[key]; found == true {
		switch e.entry.Data[key].(type) {
		case string:
			if expectedType == "string" {
				ignoreDataKey = append(ignoreDataKey, key)
			}
		case uint8, uint16, uint32, uint64, uint, int, int8, int16, int32, int64:
			if expectedType == "uint8" {
				ignoreDataKey = append(ignoreDataKey, key)
			}
		}
	}
	if found {
		return found, e.entry.Data[key]
	}
	return found, nil
}

// Serialize entry as cap'n proto message
func (e Entry) capnproto(packed bool) (out []byte, err error) {
	var found bool
	var value interface{}
	ignoreDataKey := []string{}

	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return
	}

	record, err := NewRootRecord(seg)
	if err != nil {
		return
	}

	// Timestamp: ts
	record.SetTs(float64(e.entry.Time.UnixNano()/1000000) / 1000.)

	// Hostname
	var hostname string
	if found, value = e.getCapnpFieldValue("hostname", "string", ignoreDataKey); found {
		hostname = value.(string)
	} else {
		hostname, err = os.Hostname()
		if err != nil {
			return
		}
	}
	if err = record.SetHostname(hostname); err != nil {
		return
	}

	// Facility
	var facility uint8
	if found, value = e.getCapnpFieldValue("facility", "uint8", ignoreDataKey); found {
		facility = value.(uint8)
	} else {
		facility = 13
	}
	record.SetFacility(facility)

	// Severity
	record.SetSeverity(uint8(e.entry.Level))

	// appname
	var appname string
	found, value = e.getCapnpFieldValue("appname", "string", ignoreDataKey)
	if found {
		appname = value.(string)
	} else {
		appname = "johndoeapp"
	}
	if err = record.SetAppname(appname); err != nil {
		return
	}

	// procid
	var procid string
	if found, value = e.getCapnpFieldValue("procid", "string", ignoreDataKey); found {
		procid = value.(string)
	} else {
		procid = fmt.Sprintf("%v", os.Getpid())
	}
	if err = record.SetProcid(procid); err != nil {
		return
	}

	// msgID
	var msgID string
	if found, value = e.getCapnpFieldValue("msgid", "string", ignoreDataKey); found {
		msgID = value.(string)
	} else {
		msgID = uuid.NewV4().String()
	}
	if err = record.SetMsgid(msgID); err != nil {
		return
	}

	// msg
	var recordMsg string
	if len(e.entry.Message) > 80 {
		recordMsg = e.entry.Message[:79] + "..."
	} else {
		recordMsg = e.entry.Message
	}
	if err = record.SetMsg(recordMsg); err != nil {
		return
	}

	// FullMessage
	if err = record.SetFullMsg(e.entry.Message); err != nil {
		return
	}

	// sdID
	var sdID string
	if found, value = e.getCapnpFieldValue("siid", "string", ignoreDataKey); found {
		sdID = value.(string)
	} else {
		sdID = "none"
	}
	if err = record.SetSdId(sdID); err != nil {
		return
	}

	// pairs (+1 -> token OVH)
	nbPairs := int32(len(e.entry.Data)-len(ignoreDataKey)) + 1
	pairList, err := NewPair_List(seg, nbPairs)
	if err != nil {
		return []byte{}, err
	}

	// OVH token
	pair, err := NewPair(seg)
	if err != nil {
		return []byte{}, err
	}
	if err = pair.SetKey("X-OVH-TOKEN"); err != nil {
		return
	}
	if err = pair.Value().SetString(e.ovhToken); err != nil {
		return
	}
	if err = pairList.Set(0, pair); err != nil {
		return
	}

	// loop
	i := 0
L:
	for key, value := range e.entry.Data {
		for _, ignoredKey := range ignoreDataKey {
			if key == ignoredKey {
				continue L
			}
		}
		i++
		pair, err := NewPair(seg)
		if err != nil {
			return []byte{}, err
		}
		pair.SetKey(key)
		switch value.(type) {
		case string:
			pair.Value().SetString(value.(string))
		case bool:
			pair.Value().SetBool(value.(bool))
		case float64, float32:
			pair.Value().SetF64(value.(float64))
		case int64, int, int8, int32:
			pair.Value().SetI64(int64(value.(int)))
		case uint64, uint, uint8, uint32:
			pair.Value().SetU64(value.(uint64))
		default:
			return []byte{}, fmt.Errorf("capnproto type not supported for entry field")
		}
		if err = pairList.Set(i, pair); err != nil {
			return []byte{}, err
		}
	}
	if err = record.SetPairs(pairList); err != nil {
		return
	}
	var b bytes.Buffer
	if packed {
		if err = capnp.NewPackedEncoder(&b).Encode(msg); err != nil {
			return
		}
	} else {
		if err = capnp.NewEncoder(&b).Encode(msg); err != nil {
			return
		}
	}
	return b.Bytes(), nil
}

// return a conn
func getConn(proto Protocol) (conn net.Conn, err error) {
	//var addr net.Addr
	switch proto {
	case GELFTCP:
		conn, err = net.DialTimeout("tcp", "laas.runabove.com:2202", 5*time.Second)
	case GELFTLS:
		conf := &tls.Config{}
		conn, err = tls.Dial("tcp", "laas.runabove.com:12202", conf)
		if err != nil {
			return nil, err
		}
		err = conn.SetDeadline(time.Now().Add(10 * time.Second))
	case GELFUDP:
		conn, err = net.DialTimeout("udp", "laas.runabove.com:2202", 5*time.Second)
	case CAPNPROTOTCP:
		conn, err = net.DialTimeout("tcp", "laas.runabove.com:2204", 5*time.Second)
	case CAPNPROTOTLS:
		conf := &tls.Config{}
		conn, err = tls.Dial("tcp", "laas.runabove.com:12204", conf)
		if err != nil {
			return nil, err
		}
		err = conn.SetDeadline(time.Now().Add(10 * time.Second))
	default:
		err = fmt.Errorf("%v not implemented or not supported", proto)
	}
	return
}
