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
func (e Entry) GelfSendTCP() error {

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
}

// Serialize entry for cap'n proto
func (e Entry) capnproto() ([]byte, error) {
	return []byte{}, nil
}
