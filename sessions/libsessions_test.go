// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package sessions

import (
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestLibSessionSGetSetCGRID(t *testing.T) {
	sEv := engine.NewMapEvent(map[string]any{
		utils.EVENT_NAME:       "TEST_EVENT",
		utils.ToR:              "*voice",
		utils.OriginID:         "12345",
		utils.Account:          "account1",
		utils.Subject:          "subject1",
		utils.Destination:      "+4986517174963",
		utils.Category:         "call",
		utils.Tenant:           "cgrates.org",
		utils.RequestType:      "*prepaid",
		utils.SetupTime:        "2015-11-09 14:21:24",
		utils.AnswerTime:       "2015-11-09 14:22:02",
		utils.Usage:            "1m23s",
		utils.LastUsed:         "21s",
		utils.PDD:              "300ms",
		utils.SUPPLIER:         "supplier1",
		utils.DISCONNECT_CAUSE: "NORMAL_DISCONNECT",
		utils.OriginHost:       "127.0.0.1",
	})
	//Empty CGRID in event
	cgrID := GetSetCGRID(sEv)
	if len(cgrID) == 0 {
		t.Errorf("Unexpected cgrID: %+v", cgrID)
	}
	//populate CGRID in event
	sEv[utils.CGRID] = "someRandomVal"
	cgrID = GetSetCGRID(sEv)
	if cgrID != "someRandomVal" {
		t.Errorf("Expecting: someRandomVal, received: %+v", cgrID)
	}
}

func TestLibSessionSgetSessionTTL(t *testing.T) {
	sEv := engine.NewMapEvent(map[string]any{
		utils.EVENT_NAME:       "TEST_EVENT",
		utils.ToR:              "*voice",
		utils.OriginID:         "12345",
		utils.Account:          "account1",
		utils.Subject:          "subject1",
		utils.Destination:      "+4986517174963",
		utils.Category:         "call",
		utils.Tenant:           "cgrates.org",
		utils.RequestType:      "*prepaid",
		utils.SetupTime:        "2015-11-09 14:21:24",
		utils.AnswerTime:       "2015-11-09 14:22:02",
		utils.Usage:            "1m23s",
		utils.LastUsed:         "21s",
		utils.PDD:              "300ms",
		utils.SUPPLIER:         "supplier1",
		utils.DISCONNECT_CAUSE: "NORMAL_DISCONNECT",
		utils.OriginHost:       "127.0.0.1",
	})

	sEv[utils.SessionTTL] = "notanumber"
	if _, err := getSessionTTL(&sEv, time.Duration(0), nil); err == nil {
		t.Errorf("Expecting: NOT_FOUND, received: %+v", err)
	}
	sEv[utils.SessionTTL] = 0
	if ttl, err := getSessionTTL(&sEv, time.Duration(0), nil); err != nil {
		t.Error(err)
	} else if ttl != 0 {
		t.Errorf("Expecting: %+v, received: %+v", 0, ttl)
	}
	sEv[utils.SessionTTL] = "2s"
	sEv[utils.SessionTTLMaxDelay] = "notanumber"
	if _, err := getSessionTTL(&sEv, time.Duration(0), nil); err == nil {
		t.Error("Expecting: invalid duration, received: ", err)
	}
	sEv[utils.SessionTTLMaxDelay] = 0

	//ttl is taken from event
	if ttl, err := getSessionTTL(&sEv, time.Duration(0), nil); err != nil {
		t.Error(err)
	} else if ttl != time.Duration(2*time.Second) {
		t.Errorf("Expecting: %+v, received: %+v",
			time.Duration(2*time.Second), ttl)
	}
	//remove ttl from event
	delete(sEv, utils.SessionTTL)
	if ttl, err := getSessionTTL(&sEv, time.Duration(4*time.Second), nil); err != nil {
		t.Error(err)
	} else if ttl != time.Duration(4*time.Second) {
		t.Errorf("Expecting: %+v, received: %+v",
			time.Duration(4*time.Second), ttl)
	}

	//add sessionTTLMaxDelay in event
	sEv[utils.SessionTTLMaxDelay] = "1s"
	if ttl, err := getSessionTTL(&sEv, time.Duration(4*time.Second), nil); err != nil {
		t.Error(err)
	} else if ttl <= time.Duration(4*time.Second) {
		t.Errorf("Unexpected ttl : %+v", ttl)
	}

	//remove sessionTTLMaxDelay from event
	delete(sEv, utils.SessionTTLMaxDelay)
	if ttl, err := getSessionTTL(&sEv, time.Duration(7*time.Second),
		utils.DurationPointer(time.Duration(2*time.Second))); err != nil {
		t.Error(err)
	} else if ttl <= time.Duration(7*time.Second) {
		t.Errorf("Unexpected ttl : %+v", ttl)
	}
}

func TestGetFlagIDs(t *testing.T) {
	//empty check
	rcv := getFlagIDs("")
	var eOut []string
	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expected %s , received: %s", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}
	//normal check
	rcv = getFlagIDs("*attributes:ATTR1;ATTR2")
	eOut = []string{"ATTR1", "ATTR2"}
	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expected %s , received: %s", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}
}
