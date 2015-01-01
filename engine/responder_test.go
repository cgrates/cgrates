/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) 2012-2014 ITsysCOM GmbH

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

package engine

import (
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

var rsponder *Responder

// Test internal abilites of GetDerivedChargers
func TestResponderGetDerivedChargers(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	cfgedDC := utils.DerivedChargers{&utils.DerivedCharger{RunId: "responder1", ReqTypeField: "test", DirectionField: "test", TenantField: "test",
		CategoryField: "test", AccountField: "test", SubjectField: "test", DestinationField: "test", SetupTimeField: "test", AnswerTimeField: "test", UsageField: "test"}}
	cfg.DerivedChargers = cfgedDC
	config.SetCgrConfig(cfg)
	rsponder = &Responder{}
	attrs := utils.AttrDerivedChargers{Tenant: "cgrates.org", Category: "call", Direction: "*out", Account: "responder_test", Subject: "responder_test"}
	var dcs utils.DerivedChargers
	if err := rsponder.GetDerivedChargers(attrs, &dcs); err != nil {
		t.Error("Unexpected error", err.Error())
	} else if !reflect.DeepEqual(dcs, cfgedDC) {
		t.Errorf("Expecting: %v, received: %v ", cfgedDC, dcs)
	}
}

func TestGetDerivedMaxSessionTime(t *testing.T) {
	config.CgrConfig().CombinedDerivedChargers = false
	testTenant := "vdf"
	cdr := &utils.StoredCdr{CgrId: utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()), OrderId: 123, TOR: utils.VOICE, AccId: "dsafdsaf",
		CdrHost: "192.168.1.1", CdrSource: "test", ReqType: "rated", Direction: "*out", Tenant: testTenant, Category: "call", Account: "dan", Subject: "dan",
		Destination: "1002", SetupTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		MediationRunId: utils.DEFAULT_RUNID, Usage: time.Duration(10) * time.Second, ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		Cost: 1.01, RatedAccount: "dan", RatedSubject: "dan"}
	var maxSessionTime float64
	if err := rsponder.GetDerivedMaxSessionTime(cdr.AsEvent(""), &maxSessionTime); err != nil {
		t.Error(err)
	} else if maxSessionTime != -1 {
		t.Error("Unexpected maxSessionTime received: ", maxSessionTime)
	}
	keyCharger1 := utils.ConcatenatedKey("*out", testTenant, "call", "dan", "dan")
	charger1 := utils.DerivedChargers{
		&utils.DerivedCharger{RunId: "extra1", ReqTypeField: "^prepaid", DirectionField: "*default", TenantField: "*default", CategoryField: "^0",
			AccountField: "^minitsboy", SubjectField: "^rif", DestinationField: "^0256", SetupTimeField: "*default", AnswerTimeField: "*default", UsageField: "*default"},
		&utils.DerivedCharger{RunId: "extra2", ReqTypeField: "*default", DirectionField: "*default", TenantField: "*default", CategoryField: "*default",
			AccountField: "ivo", SubjectField: "ivo", DestinationField: "*default", SetupTimeField: "*default", AnswerTimeField: "*default", UsageField: "*default"},
		&utils.DerivedCharger{RunId: "extra3", ReqTypeField: "^pseudoprepaid", DirectionField: "*default", TenantField: "*default", CategoryField: "^0",
			AccountField: "^minu", SubjectField: "^rif", DestinationField: "^0256", SetupTimeField: "*default", AnswerTimeField: "*default", UsageField: "*default"},
	}
	if err := accountingStorage.SetDerivedChargers(keyCharger1, charger1); err != nil {
		t.Error("Error on setting DerivedChargers", err.Error())
	}
	accountingStorage.CacheAccounting(nil, nil, nil, nil)
	var dcs utils.DerivedChargers
	attrs := utils.AttrDerivedChargers{Tenant: testTenant, Category: "call", Direction: "*out", Account: "dan", Subject: "dan"}
	if err := rsponder.GetDerivedChargers(attrs, &dcs); err != nil {
		t.Error("Unexpected error", err.Error())
	} else if !reflect.DeepEqual(dcs, charger1) {
		t.Errorf("Expecting: %+v, received: %+v ", charger1, dcs)
	}
	if err := rsponder.GetDerivedMaxSessionTime(cdr.AsEvent(""), &maxSessionTime); err != nil {
		t.Error(err)
	} else if maxSessionTime != 9.9e+10 { // Smallest one
		t.Error("Unexpected maxSessionTime received: ", maxSessionTime)
	}
}

func TestGetSessionRuns(t *testing.T) {
	config.CgrConfig().CombinedDerivedChargers = false
	testTenant := "vdf"
	cdr := &utils.StoredCdr{CgrId: utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()), OrderId: 123, TOR: utils.VOICE, AccId: "dsafdsaf",
		CdrHost: "192.168.1.1", CdrSource: "test", ReqType: "prepaid", Direction: "*out", Tenant: testTenant, Category: "call", Account: "dan2", Subject: "dan2",
		Destination: "1002", SetupTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		MediationRunId: utils.DEFAULT_RUNID, Usage: time.Duration(10) * time.Second, ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		Cost: 1.01, RatedAccount: "dan", RatedSubject: "dan"}
	keyCharger1 := utils.ConcatenatedKey("*out", testTenant, "call", "dan2", "dan2")
	dfDC := &utils.DerivedCharger{RunId: utils.DEFAULT_RUNID, ReqTypeField: "*default", DirectionField: "*default", TenantField: "*default", CategoryField: "*default",
		AccountField: "*default", SubjectField: "*default", DestinationField: "*default", SetupTimeField: "*default", AnswerTimeField: "*default", UsageField: "*default"}
	extra1DC := &utils.DerivedCharger{RunId: "extra1", ReqTypeField: "^prepaid", DirectionField: "*default", TenantField: "*default", CategoryField: "^0",
		AccountField: "^minitsboy", SubjectField: "^rif", DestinationField: "^0256", SetupTimeField: "*default", AnswerTimeField: "*default", UsageField: "*default"}
	extra2DC := &utils.DerivedCharger{RunId: "extra2", ReqTypeField: "*default", DirectionField: "*default", TenantField: "*default", CategoryField: "*default",
		AccountField: "^ivo", SubjectField: "^ivo", DestinationField: "*default", SetupTimeField: "*default", AnswerTimeField: "*default", UsageField: "*default"}
	extra3DC := &utils.DerivedCharger{RunId: "extra3", ReqTypeField: "^pseudoprepaid", DirectionField: "*default", TenantField: "*default", CategoryField: "^0",
		AccountField: "^minu", SubjectField: "^rif", DestinationField: "^0256", SetupTimeField: "*default", AnswerTimeField: "*default", UsageField: "*default"}
	charger1 := utils.DerivedChargers{extra1DC, extra2DC, extra3DC}
	if err := accountingStorage.SetDerivedChargers(keyCharger1, charger1); err != nil {
		t.Error("Error on setting DerivedChargers", err.Error())
	}
	accountingStorage.CacheAccounting(nil, nil, nil, nil)
	sesRuns := make([]*SessionRun, 0)
	eSRuns := []*SessionRun{
		&SessionRun{DerivedCharger: extra1DC,
			CallDescriptor: &CallDescriptor{Direction: "*out", Category: "0", Tenant: "vdf", Subject: "rif", Account: "minitsboy", Destination: "0256", TimeStart: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC)}},
		&SessionRun{DerivedCharger: extra2DC,
			CallDescriptor: &CallDescriptor{Direction: "*out", Category: "call", Tenant: "vdf", Subject: "ivo", Account: "ivo", Destination: "1002", TimeStart: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC)}},
		&SessionRun{DerivedCharger: dfDC,
			CallDescriptor: &CallDescriptor{Direction: "*out", Category: "call", Tenant: "vdf", Subject: "dan2", Account: "dan2", Destination: "1002", TimeStart: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC)}}}
	if err := rsponder.GetSessionRuns(cdr.AsEvent(""), &sesRuns); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eSRuns, sesRuns) {
		t.Errorf("Received: %+v", sesRuns)
	}
}
