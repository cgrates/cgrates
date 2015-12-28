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
	"net"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/osipsdagram"
)

var addr, _ = net.ResolveUDPAddr("udp", "172.16.254.77:42574")
var osipsEv = &OsipsEvent{osipsEvent: &osipsdagram.OsipsEvent{Name: "E_ACC_CDR",
	AttrValues: map[string]string{"to_tag": "4ea9687f", "cgr_account": "dan", "setuptime": "7", "created": "1406370492", "method": "INVITE", "callid": "ODVkMDI2Mzc2MDY5N2EzODhjNTAzNTdlODhiZjRlYWQ",
		"sip_reason": "OK", "time": "1406370499", "cgr_reqtype": utils.META_PREPAID, "cgr_subject": "dan", "cgr_destination": "+4986517174963", "cgr_tenant": "itsyscom.com", "sip_code": "200",
		"duration": "20", CGR_PDD: "3s", "from_tag": "eb082607", "extra1": "val1", "extra2": "val2", "cgr_supplier": "supplier3"}, OriginatorAddress: addr}}

func TestOsipsEventInterface(t *testing.T) {
	var _ engine.Event = engine.Event(osipsEv)
}

func TestOsipsEventParseStatic(t *testing.T) {
	setupTime, _ := osipsEv.GetSetupTime("^2013-12-07 08:42:24", "")
	answerTime, _ := osipsEv.GetAnswerTime("^2013-12-07 08:42:24", "")
	dur, _ := osipsEv.GetDuration("^60s")
	PDD, _ := osipsEv.GetPdd("^10s")
	if osipsEv.GetReqType("^test") != "test" ||
		osipsEv.GetDirection("^test") != "test" ||
		osipsEv.GetTenant("^test") != "test" ||
		osipsEv.GetCategory("^test") != "test" ||
		osipsEv.GetAccount("^test") != "test" ||
		osipsEv.GetSubject("^test") != "test" ||
		osipsEv.GetDestination("^test") != "test" ||
		setupTime != time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC) ||
		answerTime != time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC) ||
		dur != time.Duration(60)*time.Second ||
		PDD != time.Duration(10)*time.Second ||
		osipsEv.GetSupplier("^test") != "test" ||
		osipsEv.GetDisconnectCause("^test") != "test" {
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
			dur != time.Duration(60)*time.Second,
			PDD != time.Duration(10)*time.Second,
			osipsEv.GetSupplier("^test") != "test",
			osipsEv.GetDisconnectCause("^test") != "test")
	}
}

func TestOsipsEventGetValues(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	config.SetCgrConfig(cfg)
	setupTime, _ := osipsEv.GetSetupTime(utils.META_DEFAULT, "")
	eSetupTime, _ := utils.ParseTimeDetectLayout("1406370492", "")
	answerTime, _ := osipsEv.GetAnswerTime(utils.META_DEFAULT, "")
	eAnswerTime, _ := utils.ParseTimeDetectLayout("1406370499", "")
	dur, _ := osipsEv.GetDuration(utils.META_DEFAULT)
	PDD, _ := osipsEv.GetPdd(utils.META_DEFAULT)
	endTime, _ := osipsEv.GetEndTime(utils.META_DEFAULT, "")
	if osipsEv.GetName() != "E_ACC_CDR" ||
		osipsEv.GetCgrId("") != utils.Sha1("ODVkMDI2Mzc2MDY5N2EzODhjNTAzNTdlODhiZjRlYWQ", setupTime.UTC().String()) ||
		osipsEv.GetUUID() != "ODVkMDI2Mzc2MDY5N2EzODhjNTAzNTdlODhiZjRlYWQ" ||
		osipsEv.GetDirection(utils.META_DEFAULT) != utils.OUT ||
		osipsEv.GetSubject(utils.META_DEFAULT) != "dan" ||
		osipsEv.GetAccount(utils.META_DEFAULT) != "dan" ||
		osipsEv.GetDestination(utils.META_DEFAULT) != "+4986517174963" ||
		osipsEv.GetCallDestNr(utils.META_DEFAULT) != "+4986517174963" ||
		osipsEv.GetCategory(utils.META_DEFAULT) != cfg.DefaultCategory ||
		osipsEv.GetTenant(utils.META_DEFAULT) != "itsyscom.com" ||
		osipsEv.GetReqType(utils.META_DEFAULT) != utils.META_PREPAID ||
		!setupTime.Equal(eSetupTime) ||
		!answerTime.Equal(eAnswerTime) ||
		!endTime.Equal(eAnswerTime.Add(dur)) ||
		dur != time.Duration(20*time.Second) ||
		PDD != time.Duration(3)*time.Second ||
		osipsEv.GetSupplier(utils.META_DEFAULT) != "supplier3" ||
		osipsEv.GetDisconnectCause(utils.META_DEFAULT) != "200" ||
		osipsEv.GetOriginatorIP(utils.META_DEFAULT) != "172.16.254.77" {
		t.Error("GetValues not matching: ", osipsEv.GetName() != "E_ACC_CDR",
			osipsEv.GetCgrId("") != utils.Sha1("ODVkMDI2Mzc2MDY5N2EzODhjNTAzNTdlODhiZjRlYWQ", setupTime.UTC().String()),
			osipsEv.GetUUID() != "ODVkMDI2Mzc2MDY5N2EzODhjNTAzNTdlODhiZjRlYWQ",
			osipsEv.GetDirection(utils.META_DEFAULT) != utils.OUT,
			osipsEv.GetSubject(utils.META_DEFAULT) != "dan",
			osipsEv.GetAccount(utils.META_DEFAULT) != "dan",
			osipsEv.GetDestination(utils.META_DEFAULT) != "+4986517174963",
			osipsEv.GetCallDestNr(utils.META_DEFAULT) != "+4986517174963",
			osipsEv.GetCategory(utils.META_DEFAULT) != cfg.DefaultCategory,
			osipsEv.GetTenant(utils.META_DEFAULT) != "itsyscom.com",
			osipsEv.GetReqType(utils.META_DEFAULT) != utils.META_PREPAID,
			!setupTime.Equal(time.Date(2014, 7, 26, 12, 28, 12, 0, time.UTC)),
			!answerTime.Equal(time.Date(2014, 7, 26, 12, 28, 19, 0, time.Local)),
			!endTime.Equal(time.Date(2014, 7, 26, 12, 28, 39, 0, time.Local)),
			dur != time.Duration(20*time.Second),
			PDD != time.Duration(3)*time.Second,
			osipsEv.GetSupplier(utils.META_DEFAULT) != "supplier3",
			osipsEv.GetDisconnectCause(utils.META_DEFAULT) != "200",
			osipsEv.GetOriginatorIP(utils.META_DEFAULT) != "172.16.254.77",
		)
	}
}

func TestOsipsEventMissingParameter(t *testing.T) {
	if !osipsEv.MissingParameter("") {
		t.Errorf("Wrongly detected missing parameter: %+v", osipsEv)
	}
	osipsEv2 := &OsipsEvent{osipsEvent: &osipsdagram.OsipsEvent{Name: "E_ACC_CDR",
		AttrValues: map[string]string{"to_tag": "4ea9687f", "cgr_account": "dan", "setuptime": "7", "created": "1406370492", "method": "INVITE", "callid": "ODVkMDI2Mzc2MDY5N2EzODhjNTAzNTdlODhiZjRlYWQ",
			"sip_reason": "OK", "time": "1406370499", "cgr_reqtype": utils.META_PREPAID, "cgr_subject": "dan", "cgr_tenant": "itsyscom.com", "sip_code": "200",
			"duration": "20", "from_tag": "eb082607"}}}
	if !osipsEv2.MissingParameter("") {
		t.Error("Failed to detect missing parameter.")
	}
}

func TestOsipsEventAsStoredCdr(t *testing.T) {
	setupTime, _ := utils.ParseTimeDetectLayout("1406370492", "")
	answerTime, _ := utils.ParseTimeDetectLayout("1406370499", "")
	eStoredCdr := &engine.CDR{CGRID: utils.Sha1("ODVkMDI2Mzc2MDY5N2EzODhjNTAzNTdlODhiZjRlYWQ", setupTime.UTC().String()),
		ToR: utils.VOICE, OriginID: "ODVkMDI2Mzc2MDY5N2EzODhjNTAzNTdlODhiZjRlYWQ", OriginHost: "172.16.254.77", Source: "OSIPS_E_ACC_CDR",
		RequestType: utils.META_PREPAID,
		Direction:   utils.OUT, Tenant: "itsyscom.com", Category: "call", Account: "dan", Subject: "dan",
		Destination: "+4986517174963", SetupTime: setupTime, AnswerTime: answerTime,
		Usage: time.Duration(20) * time.Second, PDD: time.Duration(3) * time.Second, Supplier: "supplier3", DisconnectCause: "200", ExtraFields: map[string]string{"extra1": "val1", "extra2": "val2"}, Cost: -1}
	if storedCdr := osipsEv.AsStoredCdr(""); !reflect.DeepEqual(eStoredCdr, storedCdr) {
		t.Errorf("Expecting: %+v, received: %+v", eStoredCdr, storedCdr)
	}
}

func TestOsipsAccMissedToStoredCdr(t *testing.T) {
	setupTime, _ := utils.ParseTimeDetectLayout("1431182699", "")
	osipsEv := &OsipsEvent{osipsEvent: &osipsdagram.OsipsEvent{Name: "E_ACC_MISSED_EVENT",
		AttrValues: map[string]string{"method": "INVITE", "from_tag": "5cb81eaa", "to_tag": "", "callid": "27b1e6679ad0109b5d756e42bb4c9c28@0:0:0:0:0:0:0:0",
			"sip_code": "404", "sip_reason": "Not Found", "time": "1431182699", "cgr_reqtype": utils.META_PSEUDOPREPAID,
			"cgr_account": "1001", "cgr_destination": "1002", utils.CGR_SUPPLIER: "supplier1",
			"duration": "", "dialog_id": "3547:277000822", "extra1": "val1", "extra2": "val2"}, OriginatorAddress: addr,
	}}
	eStoredCdr := &engine.CDR{CGRID: utils.Sha1("27b1e6679ad0109b5d756e42bb4c9c28@0:0:0:0:0:0:0:0", setupTime.UTC().String()),
		ToR: utils.VOICE, OriginID: "27b1e6679ad0109b5d756e42bb4c9c28@0:0:0:0:0:0:0:0", OriginHost: "172.16.254.77", Source: "OSIPS_E_ACC_MISSED_EVENT",
		RequestType: utils.META_PSEUDOPREPAID, Direction: utils.OUT, Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Supplier: "supplier1",
		DisconnectCause: "404", Destination: "1002", SetupTime: setupTime, AnswerTime: setupTime,
		Usage: time.Duration(0), ExtraFields: map[string]string{"extra1": "val1", "extra2": "val2"}, Cost: -1}
	if storedCdr := osipsEv.AsStoredCdr(""); !reflect.DeepEqual(eStoredCdr, storedCdr) {
		t.Errorf("Expecting: %+v, received: %+v", eStoredCdr, storedCdr)
	}

}

func TestOsipsUpdateDurationFromEvent(t *testing.T) {
	osipsEv := &OsipsEvent{osipsEvent: &osipsdagram.OsipsEvent{Name: "E_ACC_EVENT",
		AttrValues: map[string]string{"method": "INVITE", "from_tag": "87d02470", "to_tag": "a671a98", "callid": "05dac0aaa716c9814f855f0e8fee6936@0:0:0:0:0:0:0:0",
			"sip_code": "200", "sip_reason": "OK", "time": "1430579770", "cgr_reqtype": utils.META_PREPAID,
			"cgr_account": "1001", "cgr_destination": "1002", utils.CGR_SUPPLIER: "supplier1",
			"duration": "", "dialog_id": "3547:277000822", "extra1": "val1", "extra2": "val2"}, OriginatorAddress: addr,
	}}
	updatedEv := &OsipsEvent{osipsEvent: &osipsdagram.OsipsEvent{Name: "E_ACC_EVENT",
		AttrValues: map[string]string{"method": "BYE", "from_tag": "a671a98", "to_tag": "87d02470", "callid": "05dac0aaa716c9814f855f0e8fee6936@0:0:0:0:0:0:0:0",
			"sip_code": "200", "sip_reason": "OK", "time": "1430579797", "cgr_reqtype": "",
			"cgr_account": "", "cgr_destination": "", utils.CGR_SUPPLIER: "",
			"duration": "", "dialog_id": "3547:277000822"}, OriginatorAddress: addr,
	}}
	eOsipsEv := &OsipsEvent{osipsEvent: &osipsdagram.OsipsEvent{Name: "E_ACC_EVENT",
		AttrValues: map[string]string{"method": "UPDATE", "from_tag": "87d02470", "to_tag": "a671a98", "callid": "05dac0aaa716c9814f855f0e8fee6936@0:0:0:0:0:0:0:0",
			"sip_code": "200", "sip_reason": "OK", "time": "1430579770", "cgr_reqtype": utils.META_PREPAID,
			"cgr_account": "1001", "cgr_destination": "1002", utils.CGR_SUPPLIER: "supplier1",
			"duration": "27s", "dialog_id": "3547:277000822", "extra1": "val1", "extra2": "val2"}, OriginatorAddress: addr,
	}}
	if err := osipsEv.updateDurationFromEvent(updatedEv); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eOsipsEv, osipsEv) {
		t.Errorf("Expecting: %+v, received: %+v", eOsipsEv.osipsEvent, osipsEv.osipsEvent)
	}
}
