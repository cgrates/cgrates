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
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/osipsdagram"
)

var osipsEv = &OsipsEvent{osipsEvent: &osipsdagram.OsipsEvent{Name: "E_ACC_CDR",
	AttrValues: map[string]string{"to_tag": "4ea9687f", "cgr_account": "dan", "setuptime": "7", "created": "1406370492", "method": "INVITE", "callid": "ODVkMDI2Mzc2MDY5N2EzODhjNTAzNTdlODhiZjRlYWQ",
		"sip_reason": "OK", "time": "1406370499", "cgr_reqtype": "prepaid", "cgr_subject": "dan", "cgr_destination": "+4986517174963", "cgr_tenant": "itsyscom.com", "sip_code": "200",
		"duration": "20", "from_tag": "eb082607", "extra1": "val1", "extra2": "val2"}}}

func TestOsipsEventInterface(t *testing.T) {
	var _ Event = Event(osipsEv)
}

func TestOsipsEventParseStatic(t *testing.T) {
	setupTime, _ := osipsEv.GetSetupTime("^2013-12-07 08:42:24")
	answerTime, _ := osipsEv.GetAnswerTime("^2013-12-07 08:42:24")
	dur, _ := osipsEv.GetDuration("^60s")
	if osipsEv.GetReqType("^test") != "test" ||
		osipsEv.GetDirection("^test") != "test" ||
		osipsEv.GetTenant("^test") != "test" ||
		osipsEv.GetCategory("^test") != "test" ||
		osipsEv.GetAccount("^test") != "test" ||
		osipsEv.GetSubject("^test") != "test" ||
		osipsEv.GetDestination("^test") != "test" ||
		setupTime != time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC) ||
		answerTime != time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC) ||
		dur != time.Duration(60)*time.Second {
		t.Error("Values out of static not matching",
			osipsEv.GetReqType("^test") != "test",
			osipsEv.GetDirection("^test") != "test",
			osipsEv.GetTenant("^test") != "test",
			osipsEv.GetCategory("^test") != "test",
			osipsEv.GetAccount("^test") != "test",
			osipsEv.GetSubject("^test") != "test",
			osipsEv.GetDestination("^test") != "test",
			setupTime != time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC),
			answerTime != time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC),
			dur != time.Duration(60)*time.Second)
	}
}

func TestOsipsEventGetValues(t *testing.T) {
	cfg, _ = config.NewDefaultCGRConfig()
	config.SetCgrConfig(cfg)
	setupTime, _ := osipsEv.GetSetupTime(utils.META_DEFAULT)
	answerTime, _ := osipsEv.GetAnswerTime(utils.META_DEFAULT)
	endTime, _ := osipsEv.GetEndTime()
	dur, _ := osipsEv.GetDuration(utils.META_DEFAULT)
	if osipsEv.GetName() != "E_ACC_CDR" ||
		osipsEv.GetCgrId() != utils.Sha1("ODVkMDI2Mzc2MDY5N2EzODhjNTAzNTdlODhiZjRlYWQ"+";"+"eb082607"+";"+"4ea9687f", setupTime.UTC().String()) ||
		osipsEv.GetUUID() != "ODVkMDI2Mzc2MDY5N2EzODhjNTAzNTdlODhiZjRlYWQ;eb082607;4ea9687f" ||
		osipsEv.GetDirection(utils.META_DEFAULT) != utils.OUT ||
		osipsEv.GetSubject(utils.META_DEFAULT) != "dan" ||
		osipsEv.GetAccount(utils.META_DEFAULT) != "dan" ||
		osipsEv.GetDestination(utils.META_DEFAULT) != "+4986517174963" ||
		osipsEv.GetCallDestNr(utils.META_DEFAULT) != "+4986517174963" ||
		osipsEv.GetCategory(utils.META_DEFAULT) != cfg.DefaultCategory ||
		osipsEv.GetTenant(utils.META_DEFAULT) != "itsyscom.com" ||
		osipsEv.GetReqType(utils.META_DEFAULT) != "prepaid" ||
		setupTime != time.Date(2014, 7, 26, 12, 28, 12, 0, time.Local) ||
		answerTime != time.Date(2014, 7, 26, 12, 28, 19, 0, time.Local) ||
		endTime != time.Date(2014, 7, 26, 12, 28, 39, 0, time.Local) ||
		dur != time.Duration(20*time.Second) {
		t.Error("GetValues not matching: ", osipsEv.GetName() != "E_ACC_CDR",
			osipsEv.GetCgrId() != utils.Sha1("ODVkMDI2Mzc2MDY5N2EzODhjNTAzNTdlODhiZjRlYWQ"+";"+"eb082607"+";"+"4ea9687f", setupTime.UTC().String()),
			osipsEv.GetUUID() != "ODVkMDI2Mzc2MDY5N2EzODhjNTAzNTdlODhiZjRlYWQ;eb082607;4ea9687f",
			osipsEv.GetDirection(utils.META_DEFAULT) != utils.OUT,
			osipsEv.GetSubject(utils.META_DEFAULT) != "dan",
			osipsEv.GetAccount(utils.META_DEFAULT) != "dan",
			osipsEv.GetDestination(utils.META_DEFAULT) != "+4986517174963",
			osipsEv.GetCallDestNr(utils.META_DEFAULT) != "+4986517174963",
			osipsEv.GetCategory(utils.META_DEFAULT) != cfg.DefaultCategory,
			osipsEv.GetTenant(utils.META_DEFAULT) != "itsyscom.com",
			osipsEv.GetReqType(utils.META_DEFAULT) != "prepaid",
			setupTime != time.Date(2014, 7, 26, 12, 28, 12, 0, time.Local),
			answerTime != time.Date(2014, 7, 26, 12, 28, 19, 0, time.Local),
			endTime != time.Date(2014, 7, 26, 12, 28, 39, 0, time.Local),
			dur != time.Duration(20*time.Second),
		)
	}
}

func TestOsipsEventMissingParameter(t *testing.T) {
	if osipsEv.MissingParameter() {
		t.Errorf("Wrongly detected missing parameter: %+v", osipsEv)
	}
	osipsEv2 := &OsipsEvent{osipsEvent: &osipsdagram.OsipsEvent{Name: "E_ACC_CDR",
		AttrValues: map[string]string{"to_tag": "4ea9687f", "cgr_account": "dan", "setuptime": "7", "created": "1406370492", "method": "INVITE", "callid": "ODVkMDI2Mzc2MDY5N2EzODhjNTAzNTdlODhiZjRlYWQ",
			"sip_reason": "OK", "time": "1406370499", "cgr_reqtype": "prepaid", "cgr_subject": "dan", "cgr_tenant": "itsyscom.com", "sip_code": "200",
			"duration": "20", "from_tag": "eb082607"}}}
	if !osipsEv2.MissingParameter() {
		t.Error("Failed to detect missing parameter.")
	}
}

func TestOsipsEventAsStoredCdr(t *testing.T) {
	eStoredCdr := &utils.StoredCdr{CgrId: utils.Sha1("ODVkMDI2Mzc2MDY5N2EzODhjNTAzNTdlODhiZjRlYWQ;eb082607;4ea9687f", time.Date(2014, 7, 26, 12, 28, 12, 0, time.Local).UTC().String()),
		TOR: utils.VOICE, AccId: "ODVkMDI2Mzc2MDY5N2EzODhjNTAzNTdlODhiZjRlYWQ;eb082607;4ea9687f", CdrHost: "localhost", CdrSource: "OSIPS_E_ACC_CDR", ReqType: "prepaid",
		Direction: utils.OUT, Tenant: "itsyscom.com", Category: "call", Account: "dan", Subject: "dan",
		Destination: "+4986517174963", SetupTime: time.Date(2014, 7, 26, 12, 28, 12, 0, time.Local), AnswerTime: time.Date(2014, 7, 26, 12, 28, 19, 0, time.Local),
		Usage: time.Duration(20) * time.Second, ExtraFields: map[string]string{"extra1": "val1", "extra2": "val2"}, Cost: -1}
	if storedCdr := osipsEv.AsStoredCdr(); !reflect.DeepEqual(eStoredCdr, storedCdr) {
		t.Errorf("Expecting: %+v, received: %+v", eStoredCdr, storedCdr)
	}
}
