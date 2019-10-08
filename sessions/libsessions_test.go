/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package sessions

import (
	"testing"
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestLibSessionSGetSetCGRID(t *testing.T) {
	sEv := engine.NewMapEvent(map[string]interface{}{
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
	sEv := engine.NewMapEvent(map[string]interface{}{
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
		utils.SessionTTL:       "2s",
	})

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
