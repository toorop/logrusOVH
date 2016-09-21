package logrusOVH

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"

	"net"
	"time"

	"os"

	"github.com/Sirupsen/logrus"
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

func (e Entry) send(proto Protocol, compression CompressAlgo) error {
	switch proto {
	case GELFTCP:
		// TODO buffered conn
		return e.sendGelfTCP(compression)
	case GELFUDP:
		return e.sendGelfUDP(compression)
	default:
		return fmt.Errorf("%v not implemented or not supported", proto)
	}
}

// GELFTCP
func (e Entry) sendGelfTCP(compression CompressAlgo) error {
	data, err := e.gelf(compression)
	if err != nil {
		return err
	}
	// get conn
	conn, err := getConn(GELFTCP)
	if err != nil {
		return err
	}
	defer conn.Close()
	data = append(data, 0)
	n, err := conn.Write(data)
	if err != nil {
		return err
	}
	if n != len(data) {
		return fmt.Errorf("entry not completely sent %d/%d", n, len(data))
	}
	return nil
}

// GELFUDP
func (e Entry) sendGelfUDP(compression CompressAlgo) error {
	data, err := e.gelf(compression)
	if err != nil {
		return err
	}

	// conn
	conn, err := getConn(GELFUDP)
	if err != nil {
		return err
	}
	defer conn.Close()

	//
	/*chanPipe := make(chan []byte)
	sendChuncked(data)*/
	if len(data) < UDP_CHUNK_MAX_SIZE {
		log.Println("NOT CHUNKED")
		n, err := conn.Write(data)
		if err != nil {
			return err
		}
		if n != len(data) {
			return fmt.Errorf("entry not completely sent %d/%d", n, len(data))
		}
	} else {
		log.Println("CHUNKED", len(data))

		// chunk buffer
		chunkBuf := bytes.NewBuffer(nil)
		// data buffer
		dataBuf := bytes.NewBuffer(data)

		// nb chunck
		nbChunks := int(math.Ceil(float64(len(data)/UDP_CHUNK_MAX_DATA_SIZE))) + 1
		log.Println("nbChunks", nbChunks)

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
						log.Println("EOF", dataBuf.Bytes())
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
			//log.Println("buf size", len(chunkBuf.Bytes()))
			fmt.Println(chunkBuf.Bytes())
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

	// remove trailing }
	out = out[0 : len(out)-1]

	// From logrus
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
	if compression != COMPRESSNONE {
		var w io.Writer
		b := bytes.NewBuffer(nil)
		switch compression {
		case COMPRESSGZIP:
			w = gzip.NewWriter(b)
		case COMPRESSZLIB:
			w = zlib.NewWriter(b)
		default:
			return []byte{}, fmt.Errorf("%v compression not supported", compression)
		}
		w.Write(out)
		out = b.Bytes()
	}

	return out, nil
}

// return a conn
func getConn(proto Protocol) (conn net.Conn, err error) {
	//var addr net.Addr
	switch proto {
	case GELFTCP:
		conn, err = net.DialTimeout("tcp", "laas.runabove.com:2202", 5*time.Second)
	case GELFUDP:
		conn, err = net.DialTimeout("udp", "laas.runabove.com:2202", 5*time.Second)
	default:
		err = fmt.Errorf("%v not implemented or not supported", proto)
	}
	return
}
