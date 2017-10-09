// MUST READ
// do not use this with automatic testing
// set your OVH token in ENV var before running test
// export OVH_LOGS_TOKEN="YOU TOKEN"
//
// As we can check if logs are really sent to OVH, check your Graylog web console
// we you launch thoses test. If you see 9 new entries it's OK.

package logrusOVH

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/Sirupsen/logrus"
)

const endpoint = "gra1.logs.ovh.com"

var msg = `Theatre eux conquis des peuples reciter petites oui. Corps non uns bonte eumes son. Halte oh soeur vaste laque votre va. Inoui il je voici carre xv je. Fut troupeaux ses cesserent peu agreerait cependant frontiere uniformes. Me sachant il conclue abattit faisait maudite la cousine. Du apparue attenua ce me lettres blanche lecture. Longeait feerique galopade pu au pourquoi repartit cavernes. Decharnees iii oui vieillards victorieux manoeuvres. Je avez tard sait idee au si cime se.

Courages nul preparer drapeaux des pourquoi apercoit. Acier porte fit jeu rirez. On groupes cadeaux retarde chasses hauteur ma pendant la qu. Pays eu qu ruer la cris dont idee la quel. Maintenant en vieillards paraissent assurances historique habilement la. Aux evidemment frissonner convulsion fut. Ah ou harmonie physique epanouir en. Reflete nations aisance chevaux du un grandie puisque.
frontiere uniformes. Me sachant il conclue abattit faisait maudite la cousine. Du apparue attenua ce me lettres blanche lecture. Longeait feerique galopade pu au pourquoi repartit cavernes. Decharnees iii oui vieillards victorieux manoeuvres. Je avez tard sait idee au si cime se.
`

func expectERRisNil(err error, t *testing.T) {
	if err != nil {
		t.Error("expected err == nil, got", err)
	}
}

func expectERRisNotNil(err error, t *testing.T) {
	if err == nil {
		t.Error("expected err != nil, got", err)
	}
}

func getToken() string {
	token := os.Getenv("OVH_LOGS_TOKEN")
	if token == "" {
		println("OVH_LOGS_TOKEN must be set in ENV for tests")
		os.Exit(0)
	}
	return token
}

/*
func TestGelfTCPBasic(t *testing.T) {
	hook, err := NewOvhHook(endpoint, getToken(), GELFTCP)
	if err != nil {
		t.Error("expected err == nil, got", err)
	}
	hook.SetCompression(COMPRESSNONE)
	log := logrus.New()
	log.Out = ioutil.Discard
	log.Hooks.Add(hook)
	log.Error(msg)
}

func TestGelfTCP(t *testing.T) {
	hook, err := NewOvhHook(endpoint, getToken(), GELFTCP)
	if err != nil {
		t.Error("expected err == nil, got", err)
	}
	hook.SetCompression(COMPRESSNONE)
	log := logrus.New()
	log.Out = ioutil.Discard
	log.Hooks.Add(hook)
	log.WithFields(logrus.Fields{"msgid": "mymsgID", "intField": 1, "T": "TestGelfTCP"}).Error(msg)
}
*/

func TestGelfTCPDeflate(t *testing.T) {
	hook, err := NewOvhHook(endpoint, getToken(), GELFTCP)
	if err != nil {
		t.Error("expected err == nil, got", err)
	}
	hook.SetCompression(COMPRESSDEFLATE)
	log := logrus.New()
	log.Out = ioutil.Discard
	log.Hooks.Add(hook)
	log.WithFields(logrus.Fields{"msgid": "mymsgID", "intField": 1, "T": "TestGelfTCPDeflate"}).Error(msg)
}

/*
func TestGelfTLS(t *testing.T) {
	hook, err := NewOvhHook(endpoint, getToken(), GELFTLS)
	if err != nil {
		t.Error("expected err == nil, got", err)
	}
	hook.SetCompression(COMPRESSNONE)
	log := logrus.New()
	log.Out = ioutil.Discard
	log.Hooks.Add(hook)
	log.WithFields(logrus.Fields{"msgid": "mymsgID", "intField": 1, "T": "TestGelfTLS"}).Error(msg)
}*/

func TestGelfUDP(t *testing.T) {
	hook, err := NewOvhHook(endpoint, getToken(), GELFUDP)
	if err != nil {
		t.Error("expected err == nil, got", err)
	}
	hook.SetCompression(COMPRESSNONE)
	log := logrus.New()
	log.Out = ioutil.Discard
	log.Hooks.Add(hook)
	log.WithFields(logrus.Fields{"msgid": "mymsgID", "intField": 1, "T": "TestGelfUDP"}).Error(msg)
}

func TestGelfUDPDeflate(t *testing.T) {
	hook, err := NewOvhHook(endpoint, getToken(), GELFUDP)
	if err != nil {
		t.Error("expected err == nil, got", err)
	}
	hook.SetCompression(COMPRESSDEFLATE)
	log := logrus.New()
	log.Out = ioutil.Discard
	log.Hooks.Add(hook)
	log.WithFields(logrus.Fields{"msgid": "mymsgID", "intField": 1, "T": "TestGelfUDPDeflate"}).Error(msg)
}

/*
func TestCompressNotAllowed(t *testing.T) {
	hook, err := NewOvhHook(endpoint, getToken(), GELFTCP)
	expectERRisNil(err, t)
	expectERRisNotNil(hook.SetCompression(COMPRESSZLIB), t)
	expectERRisNil(hook.SetCompression(COMPRESSNONE), t)
}

func TestGelfTCPGzip(t *testing.T) {
	hook, err := NewOvhHook(endpoint, getToken(), GELFTCP)
	if err != nil {
		t.Error("expected err == nil, got", err)
	}
	hook.SetCompression(COMPRESSGZIP)
	log := logrus.New()
	log.Out = ioutil.Discard
	log.Hooks.Add(hook)
	log.WithFields(logrus.Fields{"msgid": "mymsgID", "intField": 1, "T": "TestGelfTCPGzip"}).Error(msg)
}

func TestGelfTCPzlib(t *testing.T) {
	hook, err := NewOvhHook(endpoint, getToken(), GELFTCP)
	if err != nil {
		t.Error("expected err == nil, got", err)
	}
	hook.SetCompression(COMPRESSZLIB)
	log := logrus.New()
	log.Out = ioutil.Discard
	log.Hooks.Add(hook)
	log.WithFields(logrus.Fields{"msgid": "mymsgID", "intField": 1, "T": "TestGelfTCPzlib"}).Error(msg)
}

func TestCapnprotoTCP(t *testing.T) {
	hook, err := NewOvhHook(endpoint, getToken(), GELFTCP)
	if err != nil {
		t.Error("expected err == nil, got", err)
	}
	hook.SetCompression(COMPRESSNONE)
	log := logrus.New()
	log.Out = ioutil.Discard
	log.Hooks.Add(hook)
	log.WithFields(logrus.Fields{"msgid": "mymsgID", "intField": 1, "T": "TestCapnprotoTCP"}).Error(msg)
}

func TestCapnprotoTLS(t *testing.T) {
	hook, err := NewOvhHook(endpoint, getToken(), GELFTCP)
	if err != nil {
		t.Error("expected err == nil, got", err)
	}
	hook.SetCompression(COMPRESSNONE)
	log := logrus.New()
	log.Out = ioutil.Discard
	log.Hooks.Add(hook)
	log.WithFields(logrus.Fields{"msgid": "mymsgID", "intField": 1, "T": "TestCapnprotoTLS"}).Error(msg)
}

func TestAsync(t *testing.T) {
	hook, err := NewOvhHook(endpoint, getToken(), GELFTCP)
	if err != nil {
		t.Error("expected err == nil, got", err)
	}
	hook.SetCompression(COMPRESSNONE)
	log := logrus.New()
	log.Out = ioutil.Discard
	log.Hooks.Add(hook)
	log.WithFields(logrus.Fields{"msgid": "mymsgID", "intField": 1, "T": "TestAsync"}).Error(msg)
	// wait for async - yes i know, it's crappy ;)
	time.Sleep(2 + time.Second)
}
*/
