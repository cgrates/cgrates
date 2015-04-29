/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) 2012-2015 ITsysCOM GmbH

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

package cdre

import (
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestCdreGetCombimedCdrFieldVal(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	cdrs := []*engine.StoredCdr{
		&engine.StoredCdr{CgrId: utils.Sha1("dsafdsaf", time.Unix(1383813745, 0).UTC().String()), TOR: utils.VOICE, AccId: "dsafdsaf", CdrHost: "192.168.1.1",
			ReqType: utils.META_RATED, Direction: "*out", Tenant: "cgrates.org",
			Category: "call", Account: "1001", Subject: "1001", Destination: "1002", SetupTime: time.Unix(1383813745, 0).UTC(), AnswerTime: time.Unix(1383813746, 0).UTC(),
			Usage: time.Duration(10) * time.Second, MediationRunId: "RUN_RTL", Cost: 1.01},
		&engine.StoredCdr{CgrId: utils.Sha1("dsafdsaf2", time.Unix(1383813745, 0).UTC().String()), TOR: utils.VOICE, AccId: "dsafdsaf", CdrHost: "192.168.1.1",
			ReqType: utils.META_RATED, Direction: "*out", Tenant: "cgrates.org",
			Category: "call", Account: "1001", Subject: "1001", Destination: "1002", SetupTime: time.Unix(1383813745, 0).UTC(), AnswerTime: time.Unix(1383813746, 0).UTC(),
			Usage: time.Duration(10) * time.Second, MediationRunId: "CUSTOMER1", Cost: 2.01},
		&engine.StoredCdr{CgrId: utils.Sha1("dsafdsaf", time.Unix(1383813745, 0).UTC().String()), TOR: utils.VOICE, AccId: "dsafdsaf", CdrHost: "192.168.1.1",
			ReqType: utils.META_RATED, Direction: "*out", Tenant: "cgrates.org",
			Category: "call", Account: "1001", Subject: "1001", Destination: "1002", SetupTime: time.Unix(1383813745, 0).UTC(), AnswerTime: time.Unix(1383813746, 0).UTC(),
			Usage: time.Duration(10) * time.Second, MediationRunId: "CUSTOMER1", Cost: 3.01},
		&engine.StoredCdr{CgrId: utils.Sha1("dsafdsaf", time.Unix(1383813745, 0).UTC().String()), TOR: utils.VOICE, AccId: "dsafdsaf", CdrHost: "192.168.1.1",
			ReqType: utils.META_RATED, Direction: "*out", Tenant: "cgrates.org",
			Category: "call", Account: "1001", Subject: "1001", Destination: "1002", SetupTime: time.Unix(1383813745, 0).UTC(), AnswerTime: time.Unix(1383813746, 0).UTC(),
			Usage: time.Duration(10) * time.Second, MediationRunId: utils.DEFAULT_RUNID, Cost: 4.01},
		&engine.StoredCdr{CgrId: utils.Sha1("dsafdsaf", time.Unix(1383813745, 0).UTC().String()), TOR: utils.VOICE, AccId: "dsafdsaf", CdrHost: "192.168.1.1",
			ReqType: utils.META_RATED, Direction: "*out", Tenant: "cgrates.org",
			Category: "call", Account: "1000", Subject: "1001", Destination: "1002", SetupTime: time.Unix(1383813745, 0).UTC(), AnswerTime: time.Unix(1383813746, 0).UTC(),
			Usage: time.Duration(10) * time.Second, MediationRunId: "RETAIL1", Cost: 5.01},
	}
	cdre, err := NewCdrExporter(cdrs, nil, cfg.CdreProfiles["*default"], cfg.CdreProfiles["*default"].CdrFormat, cfg.CdreProfiles["*default"].FieldSeparator,
		"firstexport", 0.0, 0.0, 0.0, 0, 4, cfg.RoundingDecimals, "", 0, cfg.HttpSkipTlsVerify)
	if err != nil {
		t.Error("Unexpected error received: ", err)
	}
	fltrRule, _ := utils.ParseRSRFields("~mediation_runid:s/default/RUN_RTL/", utils.INFIELD_SEP)
	val, _ := utils.ParseRSRFields("cost", utils.INFIELD_SEP)
	cfgCdrFld := &config.CfgCdrField{Tag: "cost", Type: "cdrfield", CdrFieldId: "cost", Value: val, FieldFilter: fltrRule}
	if costVal, err := cdre.getCombimedCdrFieldVal(cdrs[3], cfgCdrFld); err != nil {
		t.Error(err)
	} else if costVal != "1.01" {
		t.Error("Expecting: 1.01, received: ", costVal)
	}
	fltrRule, _ = utils.ParseRSRFields("~mediation_runid:s/default/RETAIL1/", utils.INFIELD_SEP)
	val, _ = utils.ParseRSRFields("account", utils.INFIELD_SEP)
	cfgCdrFld = &config.CfgCdrField{Tag: "account", Type: "cdrfield", CdrFieldId: "account", Value: val, FieldFilter: fltrRule}
	if acntVal, err := cdre.getCombimedCdrFieldVal(cdrs[3], cfgCdrFld); err != nil {
		t.Error(err)
	} else if acntVal != "1000" {
		t.Error("Expecting: 1000, received: ", acntVal)
	}
}

func TestGetDateTimeFieldVal(t *testing.T) {
	cdreTst := new(CdrExporter)
	cdrTst := &engine.StoredCdr{CgrId: utils.Sha1("dsafdsaf", time.Unix(1383813745, 0).UTC().String()), TOR: utils.VOICE, AccId: "dsafdsaf", CdrHost: "192.168.1.1",
		ReqType: utils.META_RATED, Direction: "*out", Tenant: "cgrates.org",
		Category: "call", Account: "1001", Subject: "1001", Destination: "1002", SetupTime: time.Unix(1383813745, 0).UTC(), AnswerTime: time.Unix(1383813746, 0).UTC(),
		Usage: time.Duration(10) * time.Second, MediationRunId: utils.DEFAULT_RUNID, Cost: 1.01,
		ExtraFields: map[string]string{"stop_time": "2014-06-11 19:19:00 +0000 UTC", "fieldextr2": "valextr2"}}
	val, _ := utils.ParseRSRFields("stop_time", utils.INFIELD_SEP)
	layout := "2006-01-02 15:04:05"
	cfgCdrFld := &config.CfgCdrField{Tag: "stop_time", Type: "cdrfield", CdrFieldId: "stop_time", Value: val, Layout: layout}
	if cdrVal, err := cdreTst.getDateTimeFieldVal(cdrTst, cfgCdrFld); err != nil {
		t.Error(err)
	} else if cdrVal != "2014-06-11 19:19:00" {
		t.Error("Expecting: 2014-06-11 19:19:00, got: ", cdrVal)
	}
	// Test filter
	fltr, _ := utils.ParseRSRFields("~tenant:s/(.+)/itsyscom.com/", utils.INFIELD_SEP)
	cfgCdrFld = &config.CfgCdrField{Tag: "stop_time", Type: "cdrfield", CdrFieldId: "stop_time", Value: val, FieldFilter: fltr, Layout: layout}
	if _, err := cdreTst.getDateTimeFieldVal(cdrTst, cfgCdrFld); err == nil {
		t.Error(err)
	}
	val, _ = utils.ParseRSRFields("fieldextr2", utils.INFIELD_SEP)
	cfgCdrFld = &config.CfgCdrField{Tag: "stop_time", Type: "cdrfield", CdrFieldId: "stop_time", Value: val, Layout: layout}
	// Test time parse error
	if _, err := cdreTst.getDateTimeFieldVal(cdrTst, cfgCdrFld); err == nil {
		t.Error("Should give error here, got none.")
	}
}

func TestCdreCdrFieldValue(t *testing.T) {
	cdre := new(CdrExporter)
	cdr := &engine.StoredCdr{CgrId: utils.Sha1("dsafdsaf", time.Unix(1383813745, 0).UTC().String()), TOR: utils.VOICE, AccId: "dsafdsaf", CdrHost: "192.168.1.1",
		ReqType: utils.META_RATED, Direction: "*out", Tenant: "cgrates.org",
		Category: "call", Account: "1001", Subject: "1001", Destination: "1002", SetupTime: time.Unix(1383813745, 0).UTC(), AnswerTime: time.Unix(1383813746, 0).UTC(),
		Usage: time.Duration(10) * time.Second, MediationRunId: utils.DEFAULT_RUNID, Cost: 1.01}
	val, _ := utils.ParseRSRFields("destination", utils.INFIELD_SEP)
	cfgCdrFld := &config.CfgCdrField{Tag: "destination", Type: "cdrfield", CdrFieldId: "destination", Value: val}
	if val, err := cdre.cdrFieldValue(cdr, cfgCdrFld); err != nil {
		t.Error(err)
	} else if val != cdr.Destination {
		t.Errorf("Expecting: %s, received: %s", cdr.Destination, val)
	}
	fltr, _ := utils.ParseRSRFields("~tenant:s/(.+)/itsyscom.com/", utils.INFIELD_SEP)
	cfgCdrFld = &config.CfgCdrField{Tag: "destination", Type: "cdrfield", CdrFieldId: "destination", Value: val, FieldFilter: fltr}
	if _, err := cdre.cdrFieldValue(cdr, cfgCdrFld); err == nil {
		t.Error("Failed to use filter")
	}
}
