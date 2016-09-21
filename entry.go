package logrusOVH

import (
	"encoding/json"
	"fmt"
	"net"
	"time"

	"os"

	"github.com/toorop/logrus"
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

func (e Entry) send(proto Protocol) error {
	switch proto {
	case GELFTCP:
		return e.sendGelfTCP()
	default:
		return fmt.Errorf("%v not implemented or not supported", proto)
	}
}

// GELFTCP
func (e Entry) sendGelfTCP() error {
	data, err := e.gelf()
	if err != nil {
		return err
	}
	// compression... or not

	// get conn
	conn, err := getConn(GELFTCP)
	if err != nil {
		return err
	}
	// TODO no defer
	defer conn.Close()
	data = append(data, 0)
	n, err := conn.Write(data)
	if err != nil {
		return err
	}
	if n != len(data) {
		return fmt.Errorf("entry not completeley sent %d/%d", n, len(data))
	}
	return nil
}

// Serialize entry for Gelf Proto
func (e Entry) gelf() (out []byte, err error) {
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
	return out, nil
}

// GelfSendTCP send entry to OVH paas Logs via TCP
// TODO: (e Entry) Send(proto Protocol)
/*func (e Entry) GelfSendTCP() error {

	b, err := e.gelf()
	if err != nil {
		return err
	}
	b = append(b, 0)

	var addr = "laas.runabove.com:2202"
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return err
	}
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.Write(b)

	//log.Println(err, " - ", n)
	return err
}*/

// return a conn
func getConn(proto Protocol) (conn net.Conn, err error) {
	//var addr net.Addr
	switch proto {
	case GELFTCP:
		conn, err = net.DialTimeout("tcp", "laas.runabove.com:2202", 5*time.Second)
	default:
		err = fmt.Errorf("%v not implemented or not supported", proto)
	}
	return
}
