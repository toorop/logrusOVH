package logrusOVH

import (
	"errors"
	"os"
	"testing"

	"github.com/toorop/logrus"
)

func getToken() (string, error) {
	token := os.Getenv("OVH_LOGS_TOKEN")
	if token == "" {
		return "", errors.New("OVH_LOGS_TOKEN must be set un ENV")
	}
	return token, nil
}

func TestSync(t *testing.T) {
	token, err := getToken()
	if err != nil {
		println(err.Error())
		os.Exit(0)
	}
	hook, err := NewOvhHook(token, GELFTCP)
	if err != nil {
		t.Error("expected err == nil, got", err)
	}
	hook.SetCompression(COMPRESSNONE)
	msg := "test message ààà ééé"
	log := logrus.New()
	//log.Out = ioutil.Discard
	log.Hooks.Add(hook)
	log.WithFields(logrus.Fields{"stringField": "string", "intField": 1, "booField": false, "foo": "bar"}).Error(msg)
}
