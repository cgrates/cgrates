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

package cdre

import (
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"testing"
	"time"
)

func TestCdreGetCombimedCdrFieldVal(t *testing.T) {
	logDb, _ := engine.NewMapStorage()
	cfg, _ := config.NewDefaultCGRConfig()
	cdrs := []*utils.StoredCdr{
		&utils.StoredCdr{CgrId: utils.Sha1("dsafdsaf", time.Unix(1383813745, 0).UTC().String()), TOR: utils.VOICE, AccId: "dsafdsaf", CdrHost: "192.168.1.1",
			ReqType: "rated", Direction: "*out", Tenant: "cgrates.org",
			Category: "call", Account: "1001", Subject: "1001", Destination: "1002", SetupTime: time.Unix(1383813745, 0).UTC(), AnswerTime: time.Unix(1383813746, 0).UTC(),
			Usage: time.Duration(10) * time.Second, MediationRunId: "RUN_RTL", Cost: 1.01},
		&utils.StoredCdr{CgrId: utils.Sha1("dsafdsaf2", time.Unix(1383813745, 0).UTC().String()), TOR: utils.VOICE, AccId: "dsafdsaf", CdrHost: "192.168.1.1",
			ReqType: "rated", Direction: "*out", Tenant: "cgrates.org",
			Category: "call", Account: "1001", Subject: "1001", Destination: "1002", SetupTime: time.Unix(1383813745, 0).UTC(), AnswerTime: time.Unix(1383813746, 0).UTC(),
			Usage: time.Duration(10) * time.Second, MediationRunId: "CUSTOMER1", Cost: 2.01},
		&utils.StoredCdr{CgrId: utils.Sha1("dsafdsaf", time.Unix(1383813745, 0).UTC().String()), TOR: utils.VOICE, AccId: "dsafdsaf", CdrHost: "192.168.1.1",
			ReqType: "rated", Direction: "*out", Tenant: "cgrates.org",
			Category: "call", Account: "1001", Subject: "1001", Destination: "1002", SetupTime: time.Unix(1383813745, 0).UTC(), AnswerTime: time.Unix(1383813746, 0).UTC(),
			Usage: time.Duration(10) * time.Second, MediationRunId: "CUSTOMER1", Cost: 3.01},
		&utils.StoredCdr{CgrId: utils.Sha1("dsafdsaf", time.Unix(1383813745, 0).UTC().String()), TOR: utils.VOICE, AccId: "dsafdsaf", CdrHost: "192.168.1.1",
			ReqType: "rated", Direction: "*out", Tenant: "cgrates.org",
			Category: "call", Account: "1001", Subject: "1001", Destination: "1002", SetupTime: time.Unix(1383813745, 0).UTC(), AnswerTime: time.Unix(1383813746, 0).UTC(),
			Usage: time.Duration(10) * time.Second, MediationRunId: utils.DEFAULT_RUNID, Cost: 4.01},
		&utils.StoredCdr{CgrId: utils.Sha1("dsafdsaf", time.Unix(1383813745, 0).UTC().String()), TOR: utils.VOICE, AccId: "dsafdsaf", CdrHost: "192.168.1.1",
			ReqType: "rated", Direction: "*out", Tenant: "cgrates.org",
			Category: "call", Account: "1000", Subject: "1001", Destination: "1002", SetupTime: time.Unix(1383813745, 0).UTC(), AnswerTime: time.Unix(1383813746, 0).UTC(),
			Usage: time.Duration(10) * time.Second, MediationRunId: "RETAIL1", Cost: 5.01},
	}

	cdre, err := NewCdrExporter(cdrs, logDb, cfg.CdreDefaultInstance, cfg.CdreDefaultInstance.CdrFormat, "firstexport", 0.0, 0.0, 0, 4,
		cfg.RoundingDecimals, "", 0, cfg.HttpSkipTlsVerify)
	if err != nil {
		t.Error("Unexpected error received: ", err)
	}
	fltrRule, _ := utils.NewRSRField("~mediation_runid:s/default/RUN_RTL/")
	if costVal, err := cdre.getCombimedCdrFieldVal(cdrs[3], fltrRule, &utils.RSRField{Id: "cost"}); err != nil {
		t.Error(err)
	} else if costVal != "1.01" {
		t.Error("Expecting: 1.01, received: ", costVal)
	}
	fltrRule, _ = utils.NewRSRField("~mediation_runid:s/default/RETAIL1/")
	if acntVal, err := cdre.getCombimedCdrFieldVal(cdrs[3], fltrRule, &utils.RSRField{Id: "account"}); err != nil {
		t.Error(err)
	} else if acntVal != "1000" {
		t.Error("Expecting: 1000, received: ", acntVal)
	}
}

func TestGetDateTimeFieldVal(t *testing.T) {
	cdreTst := new(CdrExporter)
	cdrTst := &utils.StoredCdr{CgrId: utils.Sha1("dsafdsaf", time.Unix(1383813745, 0).UTC().String()), TOR: utils.VOICE, AccId: "dsafdsaf", CdrHost: "192.168.1.1",
		ReqType: "rated", Direction: "*out", Tenant: "cgrates.org",
		Category: "call", Account: "1001", Subject: "1001", Destination: "1002", SetupTime: time.Unix(1383813745, 0).UTC(), AnswerTime: time.Unix(1383813746, 0).UTC(),
		Usage: time.Duration(10) * time.Second, MediationRunId: utils.DEFAULT_RUNID, Cost: 1.01,
		ExtraFields: map[string]string{"stop_time": "2014-06-11 19:19:00 +0000 UTC", "fieldextr2": "valextr2"}}
	if cdrVal, err := cdreTst.getDateTimeFieldVal(cdrTst, nil, &utils.RSRField{Id: "stop_time"}, "2006-01-02 15:04:05"); err != nil {
		t.Error(err)
	} else if cdrVal != "2014-06-11 19:19:00" {
		t.Error("Expecting: 2014-06-11 19:19:00, got: ", cdrVal)
	}
	// Test filter
	fltrRule, _ := utils.NewRSRField("~tenant:s/(.+)/itsyscom.com/")
	if _, err := cdreTst.getDateTimeFieldVal(cdrTst, fltrRule, &utils.RSRField{Id: "stop_time"}, "2006-01-02 15:04:05"); err == nil {
		t.Error(err)
	}
	// Test time parse error
	if _, err := cdreTst.getDateTimeFieldVal(cdrTst, nil, &utils.RSRField{Id: "fieldextr2"}, "2006-01-02 15:04:05"); err == nil {
		t.Error("Should give error here, got none.")
	}
}

func TestCdreCdrFieldValue(t *testing.T) {
	cdre := new(CdrExporter)
	cdr := &utils.StoredCdr{CgrId: utils.Sha1("dsafdsaf", time.Unix(1383813745, 0).UTC().String()), TOR: utils.VOICE, AccId: "dsafdsaf", CdrHost: "192.168.1.1",
		ReqType: "rated", Direction: "*out", Tenant: "cgrates.org",
		Category: "call", Account: "1001", Subject: "1001", Destination: "1002", SetupTime: time.Unix(1383813745, 0).UTC(), AnswerTime: time.Unix(1383813746, 0).UTC(),
		Usage: time.Duration(10) * time.Second, MediationRunId: utils.DEFAULT_RUNID, Cost: 1.01}
	fltrRule, _ := utils.NewRSRField("~tenant:s/(.+)/cgrates.org/")
	if val, err := cdre.cdrFieldValue(cdr, fltrRule, &utils.RSRField{Id: "destination"}, ""); err != nil {
		t.Error(err)
	} else if val != cdr.Destination {
		t.Errorf("Expecting: %s, received: %s", cdr.Destination, val)
	}
	fltrRule, _ = utils.NewRSRField("~tenant:s/(.+)/itsyscom.com/")
	if _, err := cdre.cdrFieldValue(cdr, fltrRule, &utils.RSRField{Id: "destination"}, ""); err == nil {
		t.Error("Failed to use filter")
	}
}
