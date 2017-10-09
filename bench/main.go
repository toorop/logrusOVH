package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/toorop/logrusOVH"
)

var (
	nbMsg = 100
)

func getToken() string {
	token := os.Getenv("OVH_LOGS_TOKEN")
	if token == "" {
		println("OVH_LOGS_TOKEN must be set in ENV for tests")
		os.Exit(0)
	}
	return token
}

func main() {
	msg := "sdmlkfklsdjfklsdjf j foief ioehfioezui euzifu ie fueizo fuief uoieu fiozeuiezufezifu  EIFU IEZUF IEZUFI IAJopzri gpo ozei oezi ozei opeziùeospjoiujrifu  fuoizeufi uziefu  qz^ozfuzopupirugri"
	GelfTCP(msg, logrusOVH.COMPRESSNONE)
	GelfTCP(msg, logrusOVH.COMPRESSGZIP)
	GelfTCP(msg, logrusOVH.COMPRESSZLIB)
	/*GelfTLS(msg, logrusOVH.COMPRESSNONE)
	GelfTLS(msg, logrusOVH.COMPRESSGZIP)
	GelfTLS(msg, logrusOVH.COMPRESSZLIB)*/
	GelfUDP(msg, logrusOVH.COMPRESSNONE)
	GelfUDP(msg, logrusOVH.COMPRESSGZIP)
	GelfUDP(msg, logrusOVH.COMPRESSZLIB)
	GelfCAPNPROTOTCP(msg)
}

func GelfTCP(msg string, compression logrusOVH.CompressAlgo) {
	var t time.Time
	hook, err := logrusOVH.NewOvhHook(getToken(), logrusOVH.GELFTCP)
	if err != nil {
		panic(err)
	}
	hook.SetCompression(compression)
	log := logrus.New()
	log.Out = ioutil.Discard
	log.Hooks.Add(hook)
	t = time.Now()
	for i := 0; i < nbMsg; i++ {
		log.WithFields(logrus.Fields{"msgid": "mymsgID", "intField": 1, "T": "TestGelfTCP"}).Error(msg)
	}
	fmt.Println("GELFTCP", compression.String(), time.Since(t))
}

func GelfTLS(msg string, compression logrusOVH.CompressAlgo) {
	var t time.Time
	hook, err := logrusOVH.NewOvhHook(getToken(), logrusOVH.GELFTLS)
	if err != nil {
		panic(err)
	}
	hook.SetCompression(compression)
	log := logrus.New()
	log.Out = ioutil.Discard
	log.Hooks.Add(hook)
	t = time.Now()
	for i := 0; i < nbMsg; i++ {
		log.WithFields(logrus.Fields{"msgid": "mymsgID", "intField": 1, "T": "TestGelfTCP"}).Error(msg)
	}
	fmt.Println("GELFTLS", compression.String(), time.Since(t))
}

func GelfUDP(msg string, compression logrusOVH.CompressAlgo) {
	var t time.Time
	hook, err := logrusOVH.NewOvhHook(getToken(), logrusOVH.GELFUDP)
	if err != nil {
		panic(err)
	}
	hook.SetCompression(compression)
	log := logrus.New()
	log.Out = ioutil.Discard
	log.Hooks.Add(hook)
	t = time.Now()
	for i := 0; i < nbMsg; i++ {
		log.WithFields(logrus.Fields{"msgid": "mymsgID", "intField": 1, "T": "TestGelfTCP"}).Error(msg)
	}
	fmt.Println("GelfUDP", compression.String(), time.Since(t))
}

func GelfCAPNPROTOTCP(msg string) {
	var t time.Time
	hook, err := logrusOVH.NewOvhHook(getToken(), logrusOVH.CAPNPROTOTCP)
	if err != nil {
		panic(err)
	}
	log := logrus.New()
	log.Out = ioutil.Discard
	log.Hooks.Add(hook)
	t = time.Now()
	for i := 0; i < nbMsg; i++ {
		log.WithFields(logrus.Fields{"msgid": "mymsgID", "intField": 1, "T": "TestGelfTCP"}).Error(msg)
	}
	fmt.Println("CAPNPROTOTCP", time.Since(t))
}
