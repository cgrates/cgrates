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
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var cfg, _ = config.NewDefaultCGRConfig()
var err error

func TestSMGenericEventParseFields(t *testing.T) {
	smGev := SMGenericEvent{}
	smGev[utils.EVENT_NAME] = "TEST_EVENT"
	smGev[utils.TOR] = "*voice"
	smGev[utils.OriginID] = "12345"
	smGev[utils.Account] = "account1"
	smGev[utils.Subject] = "subject1"
	smGev[utils.Destination] = "+4986517174963"
	smGev[utils.Category] = "call"
	smGev[utils.Tenant] = "cgrates.org"
	smGev[utils.RequestType] = "*prepaid"
	smGev[utils.SetupTime] = "2015-11-09 14:21:24"
	smGev[utils.AnswerTime] = "2015-11-09 14:22:02"
	smGev[utils.Usage] = "1m23s"
	smGev[utils.LastUsed] = "21s"
	smGev[utils.OriginHost] = "127.0.0.1"
	smGev["Extra1"] = "Value1"
	smGev["Extra2"] = 5
	if smGev.GetName() != "TEST_EVENT" {
		t.Error("Unexpected: ", smGev.GetName())
	}
	if smGev.GetCGRID(utils.META_DEFAULT) != "cade401f46f046311ed7f62df3dfbb84adb98aad" {
		t.Error("Unexpected: ", smGev.GetCGRID(utils.META_DEFAULT))
	}
	if smGev.GetOriginID(utils.META_DEFAULT) != "12345" {
		t.Error("Unexpected: ", smGev.GetOriginID(utils.META_DEFAULT))
	}
	if !reflect.DeepEqual(smGev.GetSessionIds(), []string{"12345"}) {
		t.Error("Unexpected: ", smGev.GetSessionIds())
	}
	if smGev.GetTOR(utils.META_DEFAULT) != "*voice" {
		t.Error("Unexpected: ", smGev.GetTOR(utils.META_DEFAULT))
	}
	if smGev.GetAccount(utils.META_DEFAULT) != "account1" {
		t.Error("Unexpected: ", smGev.GetAccount(utils.META_DEFAULT))
	}
	if smGev.GetSubject(utils.META_DEFAULT) != "subject1" {
		t.Error("Unexpected: ", smGev.GetSubject(utils.META_DEFAULT))
	}
	if smGev.GetDestination(utils.META_DEFAULT) != "+4986517174963" {
		t.Error("Unexpected: ", smGev.GetDestination(utils.META_DEFAULT))
	}
	if smGev.GetCategory(utils.META_DEFAULT) != "call" {
		t.Error("Unexpected: ", smGev.GetCategory(utils.META_DEFAULT))
	}
	if smGev.GetTenant(utils.META_DEFAULT) != "cgrates.org" {
		t.Error("Unexpected: ", smGev.GetTenant(utils.META_DEFAULT))
	}
	if smGev.GetReqType(utils.META_DEFAULT) != "*prepaid" {
		t.Error("Unexpected: ", smGev.GetReqType(utils.META_DEFAULT))
	}
	if st, err := smGev.GetSetupTime(utils.META_DEFAULT, "UTC"); err != nil {
		t.Error(err)
	} else if !st.Equal(time.Date(2015, 11, 9, 14, 21, 24, 0, time.UTC)) {
		t.Error("Unexpected: ", st)
	}
	if at, err := smGev.GetAnswerTime(utils.META_DEFAULT, "UTC"); err != nil {
		t.Error(err)
	} else if !at.Equal(time.Date(2015, 11, 9, 14, 22, 2, 0, time.UTC)) {
		t.Error("Unexpected: ", at)
	}
	if et, err := smGev.GetEndTime(utils.META_DEFAULT, "UTC"); err != nil {
		t.Error(err)
	} else if !et.Equal(time.Date(2015, 11, 9, 14, 23, 25, 0, time.UTC)) {
		t.Error("Unexpected: ", et)
	}
	if dur, err := smGev.GetUsage(utils.META_DEFAULT); err != nil {
		t.Error(err)
	} else if dur != time.Duration(83)*time.Second {
		t.Error("Unexpected: ", dur)
	}
	if lastUsed, err := smGev.GetLastUsed(utils.META_DEFAULT); err != nil {
		t.Error(err)
	} else if lastUsed != time.Duration(21)*time.Second {
		t.Error("Unexpected: ", lastUsed)
	}
	if smGev.GetOriginatorIP(utils.META_DEFAULT) != "127.0.0.1" {
		t.Error("Unexpected: ", smGev.GetOriginatorIP(utils.META_DEFAULT))
	}
	if extrFlds := smGev.GetExtraFields(); !reflect.DeepEqual(extrFlds,
		map[string]string{"Extra1": "Value1", "Extra2": "5", "LastUsed": "21s"}) {
		t.Error("Unexpected: ", extrFlds)
	}
}

func TestSMGenericEventGetSessionTTL(t *testing.T) {
	smGev := SMGenericEvent{}
	smGev[utils.EVENT_NAME] = "TEST_SESSION_TTL"
	cfgSesTTL := time.Duration(5 * time.Second)
	if sTTL := smGev.GetSessionTTL(time.Duration(5*time.Second), nil); sTTL != cfgSesTTL {
		t.Errorf("Expecting: %v, received: %v", cfgSesTTL, sTTL)
	}
	smGev[utils.SessionTTL] = "6s"
	eSesTTL := time.Duration(6 * time.Second)
	if sTTL := smGev.GetSessionTTL(time.Duration(5*time.Second), nil); sTTL != eSesTTL {
		t.Errorf("Expecting: %v, received: %v", eSesTTL, sTTL)
	}
	sesTTLMaxDelay := time.Duration(10 * time.Second)
	if sTTL := smGev.GetSessionTTL(time.Duration(5*time.Second), &sesTTLMaxDelay); sTTL == eSesTTL || sTTL > eSesTTL+sesTTLMaxDelay {
		t.Errorf("Received: %v", sTTL)
	}
}

func TestSMGenericEventAsCDR(t *testing.T) {
	smGev := SMGenericEvent{}
	smGev[utils.EVENT_NAME] = "TEST_EVENT"
	smGev[utils.TOR] = utils.SMS
	smGev[utils.OriginID] = "12345"
	smGev[utils.Account] = "account1"
	smGev[utils.Subject] = "subject1"
	smGev[utils.Destination] = "+4986517174963"
	smGev[utils.Category] = "call"
	smGev[utils.Tenant] = "cgrates.org"
	smGev[utils.RequestType] = utils.META_PREPAID
	smGev[utils.SetupTime] = "2015-11-09 14:21:24"
	smGev[utils.AnswerTime] = "2015-11-09 14:22:02"
	smGev[utils.Usage] = "1m23s"
	smGev[utils.OriginHost] = "10.0.3.15"
	smGev["Extra1"] = "Value1"
	smGev["Extra2"] = 5
	eStoredCdr := &engine.CDR{CGRID: "70c4d16dce41d1f2777b4e8442cff39cf87f5f19",
		ToR: utils.SMS, OriginID: "12345", OriginHost: "10.0.3.15", Source: "SMG_TEST_EVENT",
		RequestType: utils.META_PREPAID,
		Tenant:      "cgrates.org", Category: "call", Account: "account1", Subject: "subject1",
		Destination: "+4986517174963", SetupTime: time.Date(2015, 11, 9, 14, 21, 24, 0, time.UTC),
		AnswerTime:  time.Date(2015, 11, 9, 14, 22, 2, 0, time.UTC),
		Usage:       time.Duration(83) * time.Second,
		ExtraFields: map[string]string{"Extra1": "Value1", "Extra2": "5"}, Cost: -1}
	if storedCdr := smGev.AsCDR(cfg, "UTC"); !reflect.DeepEqual(eStoredCdr, storedCdr) {
		t.Errorf("Expecting: %+v, received: %+v", eStoredCdr, storedCdr)
	}
}

func TestSMGenericEventAsLcrRequest(t *testing.T) {
	smGev := SMGenericEvent{}
	smGev[utils.EVENT_NAME] = "TEST_EVENT"
	smGev[utils.TOR] = utils.VOICE
	smGev[utils.OriginID] = "12345"
	smGev[utils.Direction] = utils.OUT
	smGev[utils.Account] = "account1"
	smGev[utils.Subject] = "subject1"
	smGev[utils.Destination] = "+4986517174963"
	smGev[utils.Category] = "call"
	smGev[utils.Tenant] = "cgrates.org"
	smGev[utils.RequestType] = utils.META_PREPAID
	smGev[utils.SetupTime] = "2015-11-09 14:21:24"
	smGev[utils.AnswerTime] = "2015-11-09 14:22:02"
	smGev[utils.Usage] = "1m23s"
	smGev[utils.PDD] = "300ms"
	smGev[utils.SUPPLIER] = "supplier1"
	smGev[utils.DISCONNECT_CAUSE] = "NORMAL_DISCONNECT"
	smGev[utils.OriginHost] = "10.0.3.15"
	smGev["Extra1"] = "Value1"
	smGev["Extra2"] = 5
	eLcrReq := &engine.LcrRequest{Direction: utils.OUT, Tenant: "cgrates.org", Category: "call",
		Account: "account1", Subject: "subject1", Destination: "+4986517174963", SetupTime: "2015-11-09 14:21:24", Duration: "1m23s"}
	if lcrReq := smGev.AsLcrRequest(); !reflect.DeepEqual(eLcrReq, lcrReq) {
		t.Errorf("Expecting: %+v, received: %+v", eLcrReq, lcrReq)
	}
}

func TestSMGenericEventGetFieldAsString(t *testing.T) {
	smGev := SMGenericEvent{}
	smGev[utils.EVENT_NAME] = "TEST_EVENT"
	smGev[utils.TOR] = utils.VOICE
	smGev[utils.OriginID] = "12345"
	smGev[utils.Direction] = utils.OUT
	smGev[utils.Account] = "account1"
	smGev[utils.Subject] = "subject1"
	eFldVal := utils.VOICE
	if strVal, err := smGev.GetFieldAsString(utils.TOR); err != nil {
		t.Error(err)
	} else if strVal != eFldVal {
		t.Errorf("Expecting:\n%s\nReceived:\n%s", eFldVal, strVal)
	}
}
