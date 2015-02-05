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
	"bytes"
	"encoding/csv"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"strings"
	"testing"
	"time"
)

func TestCsvCdrWriter(t *testing.T) {
	writer := &bytes.Buffer{}
	cfg, _ := config.NewDefaultCGRConfig()
	logDb, _ := engine.NewMapStorage()
	storedCdr1 := &utils.StoredCdr{CgrId: utils.Sha1("dsafdsaf", time.Unix(1383813745, 0).UTC().String()), TOR: utils.VOICE, AccId: "dsafdsaf", CdrHost: "192.168.1.1",
		ReqType: "rated", Direction: "*out", Tenant: "cgrates.org",
		Category: "call", Account: "1001", Subject: "1001", Destination: "1002", SetupTime: time.Unix(1383813745, 0).UTC(), AnswerTime: time.Unix(1383813746, 0).UTC(),
		Usage: time.Duration(10) * time.Second, MediationRunId: utils.DEFAULT_RUNID,
		ExtraFields: map[string]string{"extra1": "val_extra1", "extra2": "val_extra2", "extra3": "val_extra3"}, Cost: 1.01,
	}
	cdre, err := NewCdrExporter([]*utils.StoredCdr{storedCdr1}, logDb, cfg.CdreProfiles["*default"], utils.CSV, ',', "firstexport", 0.0, 0.0, 0.0, 0, 4,
		cfg.RoundingDecimals, "", 0, cfg.HttpSkipTlsVerify)
	if err != nil {
		t.Error("Unexpected error received: ", err)
	}
	csvWriter := csv.NewWriter(writer)
	if err := cdre.writeCsv(csvWriter); err != nil {
		t.Error("Unexpected error: ", err)
	}
	expected := `dbafe9c8614c785a65aabd116dd3959c3c56f7f6,*default,*voice,dsafdsaf,rated,*out,cgrates.org,call,1001,1001,1002,2013-11-07T08:42:25Z,2013-11-07T08:42:26Z,10,1.0100`
	result := strings.TrimSpace(writer.String())
	if result != expected {
		t.Errorf("Expected: \n%s received: \n%s.", expected, result)
	}
	if cdre.TotalCost() != 1.01 {
		t.Error("Unexpected TotalCost: ", cdre.TotalCost())
	}
}

func TestAlternativeFieldSeparator(t *testing.T) {
	writer := &bytes.Buffer{}
	cfg, _ := config.NewDefaultCGRConfig()
	logDb, _ := engine.NewMapStorage()
	storedCdr1 := &utils.StoredCdr{CgrId: utils.Sha1("dsafdsaf", time.Unix(1383813745, 0).UTC().String()), TOR: utils.VOICE, AccId: "dsafdsaf", CdrHost: "192.168.1.1",
		ReqType: "rated", Direction: "*out", Tenant: "cgrates.org",
		Category: "call", Account: "1001", Subject: "1001", Destination: "1002", SetupTime: time.Unix(1383813745, 0).UTC(), AnswerTime: time.Unix(1383813746, 0).UTC(),
		Usage: time.Duration(10) * time.Second, MediationRunId: utils.DEFAULT_RUNID,
		ExtraFields: map[string]string{"extra1": "val_extra1", "extra2": "val_extra2", "extra3": "val_extra3"}, Cost: 1.01,
	}
	cdre, err := NewCdrExporter([]*utils.StoredCdr{storedCdr1}, logDb, cfg.CdreProfiles["*default"], utils.CSV, '|', "firstexport", 0.0, 0.0, 0.0, 0, 4, cfg.RoundingDecimals, "", 0, cfg.HttpSkipTlsVerify)
	if err != nil {
		t.Error("Unexpected error received: ", err)
	}
	csvWriter := csv.NewWriter(writer)
	if err := cdre.writeCsv(csvWriter); err != nil {
		t.Error("Unexpected error: ", err)
	}
	expected := `dbafe9c8614c785a65aabd116dd3959c3c56f7f6|*default|*voice|dsafdsaf|rated|*out|cgrates.org|call|1001|1001|1002|2013-11-07T08:42:25Z|2013-11-07T08:42:26Z|10|1.0100`
	result := strings.TrimSpace(writer.String())
	if result != expected {
		t.Errorf("Expected: \n%s received: \n%s.", expected, result)
	}
	if cdre.TotalCost() != 1.01 {
		t.Error("Unexpected TotalCost: ", cdre.TotalCost())
	}
}
