/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012-2015 ITsysCOM

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
	"flag"
	"fmt"
	"path"
	//"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

var testIntegration = flag.Bool("integration", false, "Perform the tests in integration mode, not by default.") // This flag will be passed here via "go test -local" args

func TestITCDRsMySQL(t *testing.T) {
	if !*testIntegration {
		return
	}
	cfg, err := config.NewCGRConfigFromFolder(path.Join(*dataDir, "conf", "samples", "storage", "mysql"))
	if err != nil {
		t.Error(err)
	}
	if err := InitStorDb(cfg); err != nil {
		t.Error(err)
	}
	mysqlDb, err := NewMySQLStorage(cfg.StorDBHost, cfg.StorDBPort, cfg.StorDBName, cfg.StorDBUser, cfg.StorDBPass,
		cfg.StorDBMaxOpenConns, cfg.StorDBMaxIdleConns)
	if err != nil {
		t.Error("Error on opening database connection: ", err)
	}
	if err := testSetCDR(mysqlDb); err != nil {
		t.Error(err)
	}
}

func TestITCDRsPSQL(t *testing.T) {
	if !*testIntegration {
		return
	}
	cfg, err := config.NewCGRConfigFromFolder(path.Join(*dataDir, "conf", "samples", "storage", "postgres"))
	if err != nil {
		t.Error(err)
	}
	if err := InitStorDb(cfg); err != nil {
		t.Error(err)
	}
	psqlDb, err := NewPostgresStorage(cfg.StorDBHost, cfg.StorDBPort, cfg.StorDBName, cfg.StorDBUser, cfg.StorDBPass,
		cfg.StorDBMaxOpenConns, cfg.StorDBMaxIdleConns)
	if err != nil {
		t.Error("Error on opening database connection: ", err)
	}
	if err := testSetCDR(psqlDb); err != nil {
		t.Error(err)
	}
}

// helper function to populate CDRs and check if they were stored in storDb
func testSetCDR(cdrStorage CdrStorage) error {
	rawCDR := &CDR{
		CGRID:           utils.Sha1("testevent1", time.Date(2015, 12, 12, 14, 52, 0, 0, time.UTC).String()),
		RunID:           utils.MetaRaw,
		OriginHost:      "127.0.0.1",
		Source:          "testSetCDRs",
		OriginID:        "testevent1",
		TOR:             utils.VOICE,
		RequestType:     utils.META_PREPAID,
		Direction:       utils.OUT,
		Tenant:          "cgrates.org",
		Category:        "call",
		Account:         "1004",
		Subject:         "1004",
		Destination:     "1007",
		SetupTime:       time.Date(2015, 12, 12, 14, 52, 0, 0, time.UTC),
		PDD:             time.Duration(20) * time.Millisecond,
		AnswerTime:      time.Date(2015, 12, 12, 14, 52, 20, 0, time.UTC),
		Usage:           time.Duration(35) * time.Second,
		Supplier:        "SUPPLIER1",
		DisconnectCause: "NORMAL_DISCONNECT",
		ExtraFields:     map[string]string{"ExtraHeader1": "ExtraVal1", "ExtraHeader2": "ExtraVal2"},
		Cost:            -1,
	}
	if err := cdrStorage.SetCDR(rawCDR, false); err != nil {
		return fmt.Errorf("rawCDR: %+v, SetCDR err: %s", rawCDR, err.Error())
	}
	if cdrs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{CGRIDs: []string{rawCDR.CGRID}, RunIDs: []string{utils.MetaRaw}}); err != nil {
		return fmt.Errorf("rawCDR: %+v, GetCDRs err: %s", rawCDR, err.Error())
	} else if len(cdrs) != 1 {
		return fmt.Errorf("rawCDR %+v, Unexpected number of CDRs returned: %d", rawCDR, len(cdrs))
	}
	ratedCDR := &CDR{
		CGRID:           utils.Sha1("testevent1", time.Date(2015, 12, 12, 14, 52, 0, 0, time.UTC).String()),
		RunID:           utils.META_DEFAULT,
		OriginHost:      "127.0.0.1",
		Source:          "testSetCDRs",
		OriginID:        "testevent1",
		TOR:             utils.VOICE,
		RequestType:     utils.META_PREPAID,
		Direction:       utils.OUT,
		Tenant:          "cgrates.org",
		Category:        "call",
		Account:         "1004",
		Subject:         "1004",
		Destination:     "1007",
		SetupTime:       time.Date(2015, 12, 12, 14, 52, 0, 0, time.UTC),
		PDD:             time.Duration(20) * time.Millisecond,
		AnswerTime:      time.Date(2015, 12, 12, 14, 52, 20, 0, time.UTC),
		Usage:           time.Duration(35) * time.Second,
		Supplier:        "SUPPLIER1",
		DisconnectCause: "NORMAL_DISCONNECT",
		ExtraFields:     map[string]string{"ExtraHeader1": "ExtraVal1", "ExtraHeader2": "ExtraVal2"},
		CostSource:      "testSetCDRs",
		Cost:            0.17,
	}
	if err := cdrStorage.SetCDR(ratedCDR, false); err != nil {
		return fmt.Errorf("ratedCDR: %+v, SetCDR err: %s", ratedCDR, err.Error())
	}
	if cdrs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{CGRIDs: []string{ratedCDR.CGRID}, RunIDs: []string{ratedCDR.RunID}}); err != nil {
		return fmt.Errorf("ratedCDR: %+v, GetCDRs err: %s", ratedCDR, err.Error())
	} else if len(cdrs) != 1 {
		return fmt.Errorf("ratedCDR %+v, Unexpected number of CDRs returned: %d", ratedCDR, len(cdrs))
	} else {
		if cdrs[0].RunID != ratedCDR.RunID {
			return fmt.Errorf("Unexpected ratedCDR received: %+v", cdrs[0])
		}
		if cdrs[0].RequestType != ratedCDR.RequestType {
			return fmt.Errorf("Unexpected ratedCDR received: %+v", cdrs[0])
		}
		if cdrs[0].Cost != ratedCDR.Cost {
			return fmt.Errorf("Unexpected ratedCDR received: %+v", cdrs[0])
		}
	}
	// Make sure duplicating does not work
	if err := cdrStorage.SetCDR(ratedCDR, false); err == nil {
		return fmt.Errorf("Duplicating ratedCDR: %+v works", ratedCDR)
	}
	ratedCDR.RequestType = utils.META_RATED
	ratedCDR.Cost = 0.34
	if err := cdrStorage.SetCDR(ratedCDR, true); err != nil {
		return fmt.Errorf("Rerating ratedCDR: %+v, SetCDR err: %s", ratedCDR, err.Error())
	}
	if cdrs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{CGRIDs: []string{ratedCDR.CGRID}, RunIDs: []string{ratedCDR.RunID}}); err != nil {
		return fmt.Errorf("Rerating ratedCDR: %+v, GetCDRs err: %s", ratedCDR, err.Error())
	} else if len(cdrs) != 1 {
		return fmt.Errorf("Rerating ratedCDR %+v, Unexpected number of CDRs returned: %d", ratedCDR, len(cdrs))
	} else {
		if cdrs[0].RunID != ratedCDR.RunID {
			return fmt.Errorf("Unexpected ratedCDR received after rerating: %+v", cdrs[0])
		}
		if cdrs[0].RequestType != ratedCDR.RequestType {
			return fmt.Errorf("Unexpected ratedCDR received after rerating: %+v", cdrs[0])
		}
		if cdrs[0].Cost != ratedCDR.Cost {
			return fmt.Errorf("Unexpected ratedCDR received after rerating: %+v", cdrs[0])
		}
	}
	return nil
}
