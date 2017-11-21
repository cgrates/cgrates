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
package engine

import (
	"bytes"
	"encoding/csv"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func TestCsvCdrWriter(t *testing.T) {
	writer := &bytes.Buffer{}
	cfg, _ := config.NewDefaultCGRConfig()
	storedCdr1 := &CDR{
		CGRID: utils.Sha1("dsafdsaf", time.Unix(1383813745, 0).UTC().String()),
		ToR:   utils.VOICE, OriginID: "dsafdsaf", OriginHost: "192.168.1.1",
		RequestType: utils.META_RATED, Tenant: "cgrates.org", Category: "call",
		Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime:  time.Unix(1383813745, 0).UTC(),
		AnswerTime: time.Unix(1383813746, 0).UTC(),
		Usage:      time.Duration(10) * time.Second, RunID: utils.DEFAULT_RUNID, Cost: 1.01,
		ExtraFields: map[string]string{"extra1": "val_extra1", "extra2": "val_extra2", "extra3": "val_extra3"},
	}
	cdre, err := NewCDRExporter([]*CDR{storedCdr1}, cfg.CdreProfiles["*default"], utils.MetaFileCSV, "", "", "firstexport",
		true, 1, ',', map[string]float64{}, 0.0, cfg.RoundingDecimals, cfg.HttpSkipTlsVerify, nil)
	if err != nil {
		t.Error("Unexpected error received: ", err)
	}
	if err = cdre.processCDRs(); err != nil {
		t.Error(err)
	}
	csvWriter := csv.NewWriter(writer)
	if err := cdre.writeCsv(csvWriter); err != nil {
		t.Error("Unexpected error: ", err)
	}
	expected := `dbafe9c8614c785a65aabd116dd3959c3c56f7f6,*default,*voice,dsafdsaf,*rated,cgrates.org,call,1001,1001,1002,2013-11-07T08:42:25Z,2013-11-07T08:42:26Z,10,1.01000`
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
	storedCdr1 := &CDR{CGRID: utils.Sha1("dsafdsaf", time.Unix(1383813745, 0).UTC().String()),
		ToR: utils.VOICE, OriginID: "dsafdsaf", OriginHost: "192.168.1.1",
		RequestType: utils.META_RATED, Tenant: "cgrates.org", Category: "call",
		Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime:  time.Unix(1383813745, 0).UTC(),
		AnswerTime: time.Unix(1383813746, 0).UTC(),
		Usage:      time.Duration(10) * time.Second,
		RunID:      utils.DEFAULT_RUNID, Cost: 1.01,
		ExtraFields: map[string]string{"extra1": "val_extra1", "extra2": "val_extra2", "extra3": "val_extra3"},
	}
	cdre, err := NewCDRExporter([]*CDR{storedCdr1}, cfg.CdreProfiles["*default"], utils.MetaFileCSV, "", "", "firstexport",
		true, 1, '|', map[string]float64{}, 0.0, cfg.RoundingDecimals, cfg.HttpSkipTlsVerify, nil)
	if err != nil {
		t.Error("Unexpected error received: ", err)
	}
	if err = cdre.processCDRs(); err != nil {
		t.Error(err)
	}
	csvWriter := csv.NewWriter(writer)
	if err := cdre.writeCsv(csvWriter); err != nil {
		t.Error("Unexpected error: ", err)
	}
	expected := `dbafe9c8614c785a65aabd116dd3959c3c56f7f6|*default|*voice|dsafdsaf|*rated|cgrates.org|call|1001|1001|1002|2013-11-07T08:42:25Z|2013-11-07T08:42:26Z|10|1.01000`
	result := strings.TrimSpace(writer.String())
	if result != expected {
		t.Errorf("Expected: \n%s received: \n%s.", expected, result)
	}
	if cdre.TotalCost() != 1.01 {
		t.Error("Unexpected TotalCost: ", cdre.TotalCost())
	}
}
