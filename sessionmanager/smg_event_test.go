/*
Real-time Charging System for Telecom & ISP environments
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

package sessionmanager

import (
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var cfg, _ = config.NewDefaultCGRConfig()

func TestSMGenericEventParseFields(t *testing.T) {
	smGev := SMGenericEvent{}
	smGev[utils.EVENT_NAME] = "TEST_EVENT"
	smGev[utils.TOR] = "*voice"
	smGev[utils.ACCID] = "12345"
	smGev[utils.DIRECTION] = "*out"
	smGev[utils.ACCOUNT] = "account1"
	smGev[utils.SUBJECT] = "subject1"
	smGev[utils.DESTINATION] = "+4986517174963"
	smGev[utils.CATEGORY] = "call"
	smGev[utils.TENANT] = "cgrates.org"
	smGev[utils.REQTYPE] = "*prepaid"
	smGev[utils.SETUP_TIME] = "2015-11-09 14:21:24"
	smGev[utils.ANSWER_TIME] = "2015-11-09 14:22:02"
	smGev[utils.USAGE] = "1m23s"
	smGev[utils.PDD] = "300ms"
	smGev[utils.SUPPLIER] = "supplier1"
	smGev[utils.DISCONNECT_CAUSE] = "NORMAL_DISCONNECT"
	smGev[utils.CDRHOST] = "127.0.0.1"
	smGev["Extra1"] = "Value1"
	smGev["Extra2"] = 5
	if smGev.GetName() != "TEST_EVENT" {
		t.Error("Unexpected: ", smGev.GetName())
	}
	if smGev.GetCgrId("UTC") != "0711eaa78e53937f1593dabc08c83ea04a915f2e" {
		t.Error("Unexpected: ", smGev.GetCgrId("UTC"))
	}
	if smGev.GetUUID() != "12345" {
		t.Error("Unexpected: ", smGev.GetUUID())
	}
	if !reflect.DeepEqual(smGev.GetSessionIds(), []string{"12345"}) {
		t.Error("Unexpected: ", smGev.GetSessionIds())
	}
	if smGev.GetDirection(utils.META_DEFAULT) != "*out" {
		t.Error("Unexpected: ", smGev.GetDirection(utils.META_DEFAULT))
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
	if pdd, err := smGev.GetPdd(utils.META_DEFAULT); err != nil {
		t.Error(err)
	} else if pdd != time.Duration(300)*time.Millisecond {
		t.Error("Unexpected: ", pdd)
	}
	if smGev.GetSupplier(utils.META_DEFAULT) != "supplier1" {
		t.Error("Unexpected: ", smGev.GetSupplier(utils.META_DEFAULT))
	}
	if smGev.GetDisconnectCause(utils.META_DEFAULT) != "NORMAL_DISCONNECT" {
		t.Error("Unexpected: ", smGev.GetDisconnectCause(utils.META_DEFAULT))
	}
	if smGev.GetOriginatorIP(utils.META_DEFAULT) != "127.0.0.1" {
		t.Error("Unexpected: ", smGev.GetOriginatorIP(utils.META_DEFAULT))
	}
	if extrFlds := smGev.GetExtraFields(); !reflect.DeepEqual(extrFlds, map[string]string{"Extra1": "Value1", "Extra2": "5"}) {
		t.Error("Unexpected: ", extrFlds)
	}
}

func TestSMGenericEventAsStoredCdr(t *testing.T) {
	smGev := SMGenericEvent{}
	smGev[utils.EVENT_NAME] = "TEST_EVENT"
	smGev[utils.TOR] = utils.SMS
	smGev[utils.ACCID] = "12345"
	smGev[utils.DIRECTION] = utils.OUT
	smGev[utils.ACCOUNT] = "account1"
	smGev[utils.SUBJECT] = "subject1"
	smGev[utils.DESTINATION] = "+4986517174963"
	smGev[utils.CATEGORY] = "call"
	smGev[utils.TENANT] = "cgrates.org"
	smGev[utils.REQTYPE] = utils.META_PREPAID
	smGev[utils.SETUP_TIME] = "2015-11-09 14:21:24"
	smGev[utils.ANSWER_TIME] = "2015-11-09 14:22:02"
	smGev[utils.USAGE] = "1m23s"
	smGev[utils.PDD] = "300ms"
	smGev[utils.SUPPLIER] = "supplier1"
	smGev[utils.DISCONNECT_CAUSE] = "NORMAL_DISCONNECT"
	smGev[utils.CDRHOST] = "10.0.3.15"
	smGev["Extra1"] = "Value1"
	smGev["Extra2"] = 5
	eStoredCdr := &engine.CDR{CGRID: "0711eaa78e53937f1593dabc08c83ea04a915f2e",
		ToR: utils.SMS, OriginID: "12345", OriginHost: "10.0.3.15", Source: "SMG_TEST_EVENT", RequestType: utils.META_PREPAID,
		Direction: utils.OUT, Tenant: "cgrates.org", Category: "call", Account: "account1", Subject: "subject1",
		Destination: "+4986517174963", SetupTime: time.Date(2015, 11, 9, 14, 21, 24, 0, time.UTC), AnswerTime: time.Date(2015, 11, 9, 14, 22, 2, 0, time.UTC),
		Usage: time.Duration(83) * time.Second, PDD: time.Duration(300) * time.Millisecond, Supplier: "supplier1", DisconnectCause: "NORMAL_DISCONNECT",
		ExtraFields: map[string]string{"Extra1": "Value1", "Extra2": "5"}, Cost: -1}
	if storedCdr := smGev.AsStoredCdr(cfg, "UTC"); !reflect.DeepEqual(eStoredCdr, storedCdr) {
		t.Errorf("Expecting: %+v, received: %+v", eStoredCdr, storedCdr)
	}
}

func TestSMGenericEventAsLcrRequest(t *testing.T) {
	smGev := SMGenericEvent{}
	smGev[utils.EVENT_NAME] = "TEST_EVENT"
	smGev[utils.TOR] = utils.VOICE
	smGev[utils.ACCID] = "12345"
	smGev[utils.DIRECTION] = utils.OUT
	smGev[utils.ACCOUNT] = "account1"
	smGev[utils.SUBJECT] = "subject1"
	smGev[utils.DESTINATION] = "+4986517174963"
	smGev[utils.CATEGORY] = "call"
	smGev[utils.TENANT] = "cgrates.org"
	smGev[utils.REQTYPE] = utils.META_PREPAID
	smGev[utils.SETUP_TIME] = "2015-11-09 14:21:24"
	smGev[utils.ANSWER_TIME] = "2015-11-09 14:22:02"
	smGev[utils.USAGE] = "1m23s"
	smGev[utils.PDD] = "300ms"
	smGev[utils.SUPPLIER] = "supplier1"
	smGev[utils.DISCONNECT_CAUSE] = "NORMAL_DISCONNECT"
	smGev[utils.CDRHOST] = "10.0.3.15"
	smGev["Extra1"] = "Value1"
	smGev["Extra2"] = 5
	eLcrReq := &engine.LcrRequest{Direction: utils.OUT, Tenant: "cgrates.org", Category: "call",
		Account: "account1", Subject: "subject1", Destination: "+4986517174963", SetupTime: "2015-11-09 14:21:24", Duration: "1m23s"}
	if lcrReq := smGev.AsLcrRequest(); !reflect.DeepEqual(eLcrReq, lcrReq) {
		t.Errorf("Expecting: %+v, received: %+v", eLcrReq, lcrReq)
	}
}
