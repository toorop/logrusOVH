package logrusOVH

import (
	"errors"
	"io/ioutil"
	"os"
	"testing"

	"github.com/Sirupsen/logrus"
)

func getToken() (string, error) {
	token := os.Getenv("OVH_LOGS_TOKEN")
	if token == "" {
		return "", errors.New("OVH_LOGS_TOKEN must be set in ENV")
	}
	return token, nil
}

func TestSync(t *testing.T) {
	token, err := getToken()
	if err != nil {
		println(err.Error())
		os.Exit(0)
	}
	hook, err := NewOvhHook(token, GELFUDP)
	if err != nil {
		t.Error("expected err == nil, got", err)
	}
	hook.SetCompression(COMPRESSNONE)
	msg := `Theatre eux conquis des peuples reciter petites oui. Corps non uns bonte eumes son. Halte oh soeur vaste laque votre va. Inoui il je voici carre xv je. Fut troupeaux ses cesserent peu agreerait cependant frontiere uniformes. Me sachant il conclue abattit faisait maudite la cousine. Du apparue attenua ce me lettres blanche lecture. Longeait feerique galopade pu au pourquoi repartit cavernes. Decharnees iii oui vieillards victorieux manoeuvres. Je avez tard sait idee au si cime se. 

Courages nul preparer drapeaux des pourquoi apercoit. Acier porte fit jeu rirez. On groupes cadeaux retarde chasses hauteur ma pendant la qu. Pays eu qu ruer la cris dont idee la quel. Maintenant en vieillards paraissent assurances historique habilement la. Aux evidemment frissonner convulsion fut. Ah ou harmonie physique epanouir en. Reflete nations aisance chevaux du un grandie puisque. 

Tambours tu du ignorant de as philippe lointain. Vie que folles pointe levres eux femmes vif. Ai pourtant troupeau ah familles de. Ont craignait ses echauffer echangent fit petillent. Or patiemment historique le xv compassion renferment. Par verte fin gagne roche crier soirs force. Debouche ils allaient par peu dit arrivera interdit triomphe actrices. Jour eux pere murs bon fins ils par.`
	log := logrus.New()
	log.Out = ioutil.Discard
	log.Hooks.Add(hook)
	log.WithFields(logrus.Fields{"stringField": "string", "intField": 1, "foo": "bar"}).Error(msg)
}
