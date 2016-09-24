# OVH "paas logs" hook for logrus [![GoDoc](http://godoc.org/github.com/toorop/logrusOVH?status.svg)](http://godoc.org/github.com/toorop/logrusOVH) [![Go Report Card](https://goreportcard.com/badge/github.com/toorop/logrusOVH)](https://goreportcard.com/report/github.com/toorop/logrusOVH)

Use this hook to send your [Logrus](https://github.com/Sirupsen/logrus) logs to [OVH "paas logs"](https://www.runabove.com/paas-logs.xml)

## Installation

```go
go get github.com/toorop/logrusOVH
```

## Usage

###  GELF
```go
import (
  "github.com/Sirupsen/logrus"
  "github.com/toorop/logrusOVH"
)
hook, err := NewOvhHook("YOUR OVH TOKEN", GELFTCP)
if err != nil {
    panic( err)
}
hook.SetCompression(COMPRESSNONE)
log := logrus.New()
log.Out = ioutil.Discard
log.Hooks.Add(hook)
log.WithFields(logrus.Fields{"msgid": "mymsgID", "intField": 1, "T": "TestGelfTCP"}).Error(msg)
```

###  GELF + GZIP + UDP
```go
import (
    "github.com/Sirupsen/logrus"
    "github.com/toorop/logrusOVH"
)
hook, err := NewOvhHook("YOUR OVH TOKEN", GELFUDP)
if err != nil {
    panic( err)
}
hook.SetCompression(COMPRESSGZIP)
log := logrus.New()
log.Out = ioutil.Discard
log.Hooks.Add(hook)
log.WithFields(logrus.Fields{"msgid": "mymsgID", "intField": 1, "T": "GELF + GZIP + UDP"}).Error(msg)
```


### Cap'n Proto + TLS
```go
import (
  "github.com/Sirupsen/logrus"
  "github.com/toorop/logrusOVH"
)
hook, err := NewOvhHook("YOUR OVH TOKEN", CAPNPROTOTLS)
if err != nil {
    panic( err)
}
hook.SetCompression(COMPRESSNONE)
log := logrus.New()
log.Out = ioutil.Discard
log.Hooks.Add(hook)
log.WithFields(logrus.Fields{"msgid": "mymsgID", "intField": 1, "T": "TestGelfTCP"}).Error(msg)
```

## Available serialisations, transport
 * GELFTCP: Gelf serialisation & TCP transport
 * GELFUDP: Gelf serialisation & UDP transport
 * GELFTLS: Gelf serialisation & TCP/TLS transport
 * CAPNPROTOTCP: Cap'n proto serialisation & TCP transport
 * CAPNPROTOTLS: Cap'n proto serialisation & TCP/TLS transport

## Available compression
 * COMPRESSNONE: no compression (default)
 * COMPRESSGZIP: GZIP compression for GELF
 * COMPRESSZLIB ZLIB compression for GELF
 * COMPRESSPACKNPPACKED: cap'n proto packed (not working yet)