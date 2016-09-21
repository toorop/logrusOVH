package logrusOVH

import "github.com/toorop/logrus"

// Protocol define available transfert proto
type Protocol uint8

const (
	// GELFUDP for Gelf + UDP
	GELFUDP Protocol = 1 + iota
	// GELFTCP for Gelf + TCP
	GELFTCP
	// GELFTLS for Gelf + TLS
	GELFTLS
	// CAPNPROTOUDP for Cap'n proto + UDP
	CAPNPROTOUDP
	// CAPNPROTOTCP for Cap'n proto + TCP
	CAPNPROTOTCP
	// CAPNPROTOTLS for Cap'n proto + TLS
	CAPNPROTOTLS
)

// TODO reverse map

// OvhHook represents an OVH PAAS Log
type OvhHook struct {
	async  bool
	token  string
	levels []logrus.Level
	proto  Protocol
}

// NewOvhHook returns a sync Hook
func NewOvhHook(ovhToken string, proto Protocol) (*OvhHook, error) {
	return newOvhHook(ovhToken, proto, false)
}

// NewAsyncOvhHook returns a async hook
func NewAsyncOvhHook(ovhToken string, proto Protocol) (*OvhHook, error) {
	return newOvhHook(ovhToken, proto, true)
}

// generic (ooops)
func newOvhHook(ovhToken string, proto Protocol, async bool) (*OvhHook, error) {
	hook := OvhHook{
		async:  async,
		token:  ovhToken,
		proto:  proto,
		levels: logrus.AllLevels,
	}
	return &hook, nil
}

// TODO SetLevels

// Fire is called when a log event is fired.
func (hook *OvhHook) Fire(logrusEntry *logrus.Entry) error {
	e := Entry{
		entry:    logrusEntry,
		ovhToken: hook.token,
	}
	return e.send(hook.proto)
}

// Levels returns the available logging levels (interface impl)
func (hook *OvhHook) Levels() []logrus.Level {
	return hook.levels
}
