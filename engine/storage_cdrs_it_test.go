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
	if err := testGetCDRs(cfg); err != nil {
		t.Error(err)
	}
	if err := testSetCDR(cfg); err != nil {
		t.Error(err)
	}
	if err := testSMCosts(cfg); err != nil {
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
	if err := testGetCDRs(cfg); err != nil {
		t.Error(err)
	}
	if err := testSetCDR(cfg); err != nil {
		t.Error(err)
	}
	if err := testSMCosts(cfg); err != nil {
		t.Error(err)
	}
}

func TestITCDRsMongo(t *testing.T) {
	if !*testIntegration {
		return
	}
	cfg, err := config.NewCGRConfigFromFolder(path.Join(*dataDir, "conf", "samples", "storage", "mongo"))
	if err != nil {
		t.Error(err)
	}
	if err := testGetCDRs(cfg); err != nil {
		t.Error(err)
	}
	if err := testSetCDR(cfg); err != nil {
		t.Error(err)
	}
	if err := testSMCosts(cfg); err != nil {
		t.Error(err)
	}
}

// helper function to populate CDRs and check if they were stored in storDb
func testSetCDR(cfg *config.CGRConfig) error {
	if err := InitStorDb(cfg); err != nil {
		return err
	}
	cdrStorage, err := ConfigureCdrStorage(cfg.StorDBType, cfg.StorDBHost, cfg.StorDBPort, cfg.StorDBName, cfg.StorDBUser, cfg.StorDBPass,
		cfg.StorDBMaxOpenConns, cfg.StorDBMaxIdleConns)
	if err != nil {
		return err
	}
	rawCDR := &CDR{
		CGRID:           utils.Sha1("testevent1", time.Date(2015, 12, 12, 14, 52, 0, 0, time.UTC).String()),
		RunID:           utils.MetaRaw,
		OrderID:         time.Now().UnixNano(),
		OriginHost:      "127.0.0.1",
		Source:          "testSetCDRs",
		OriginID:        "testevent1",
		ToR:             utils.VOICE,
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
		ToR:             utils.VOICE,
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

func testSMCosts(cfg *config.CGRConfig) error {
	if err := InitStorDb(cfg); err != nil {
		return err
	}
	cdrStorage, err := ConfigureCdrStorage(cfg.StorDBType, cfg.StorDBHost, cfg.StorDBPort, cfg.StorDBName, cfg.StorDBUser, cfg.StorDBPass,
		cfg.StorDBMaxOpenConns, cfg.StorDBMaxIdleConns)
	if err != nil {
		return err
	}
	cc := &CallCost{
		Direction:   utils.OUT,
		Destination: "+4986517174963",
		Timespans: []*TimeSpan{
			&TimeSpan{
				TimeStart:     time.Date(2015, 12, 28, 8, 53, 0, 0, time.UTC).Local(), // MongoDB saves timestamps in local timezone
				TimeEnd:       time.Date(2015, 12, 28, 8, 54, 40, 0, time.UTC).Local(),
				DurationIndex: 0,
				RateInterval:  &RateInterval{Rating: &RIRate{Rates: RateGroups{&Rate{GroupIntervalStart: 0, Value: 100, RateIncrement: 10 * time.Second, RateUnit: time.Second}}}},
			},
		},
		TOR: utils.VOICE,
	}
	if err := cdrStorage.LogCallCost("164b0422fdc6a5117031b427439482c6a4f90e41", utils.META_DEFAULT, utils.UNIT_TEST, cc); err != nil {
		return err
	}
	if rcvCC, err := cdrStorage.GetCallCostLog("164b0422fdc6a5117031b427439482c6a4f90e41", utils.META_DEFAULT); err != nil {
		return err
	} else if len(cc.Timespans) != len(rcvCC.Timespans) { // cc.Timespans[0].RateInterval.Rating.Rates[0], rcvCC.Timespans[0].RateInterval.Rating.Rates[0])
		return fmt.Errorf("Expecting: %+v, received: %+v", cc, rcvCC)
	}
	return nil
}

func testGetCDRs(cfg *config.CGRConfig) error {
	if err := InitStorDb(cfg); err != nil {
		return err
	}
	cdrStorage, err := ConfigureCdrStorage(cfg.StorDBType, cfg.StorDBHost, cfg.StorDBPort, cfg.StorDBName, cfg.StorDBUser, cfg.StorDBPass,
		cfg.StorDBMaxOpenConns, cfg.StorDBMaxIdleConns)
	if err != nil {
		return err
	}
	// All CDRs, no filter
	if CDRs, _, err := cdrStorage.GetCDRs(new(utils.CDRsFilter)); err != nil {
		return err
	} else if len(CDRs) != 0 {
		return fmt.Errorf("Unexpected number of CDRs returned: ", CDRs)
	}
	cdrs := []*CDR{
		&CDR{
			CGRID:           utils.Sha1("testevent1", time.Date(2015, 12, 12, 14, 52, 0, 0, time.UTC).String()),
			RunID:           utils.MetaRaw,
			OriginHost:      "127.0.0.1",
			Source:          "testGetCDRs",
			OriginID:        "testevent1",
			ToR:             utils.VOICE,
			RequestType:     utils.META_PREPAID,
			Direction:       utils.OUT,
			Tenant:          "cgrates.org",
			Category:        "call",
			Account:         "1001",
			Subject:         "1001",
			Destination:     "1002",
			SetupTime:       time.Date(2015, 12, 12, 14, 52, 0, 0, time.UTC),
			PDD:             time.Duration(20) * time.Millisecond,
			AnswerTime:      time.Date(2015, 12, 12, 14, 52, 20, 0, time.UTC),
			Usage:           time.Duration(35) * time.Second,
			Supplier:        "SUPPLIER1",
			DisconnectCause: "NORMAL_DISCONNECT",
			ExtraFields:     map[string]string{"ExtraHeader1": "ExtraVal1", "ExtraHeader2": "ExtraVal2"},
			CostSource:      "",
			Cost:            -1,
		},
		&CDR{
			CGRID:           utils.Sha1("testevent1", time.Date(2015, 12, 12, 14, 52, 0, 0, time.UTC).String()),
			RunID:           utils.META_DEFAULT,
			OriginHost:      "127.0.0.1",
			Source:          "testGetCDRs",
			OriginID:        "testevent1",
			ToR:             utils.VOICE,
			RequestType:     utils.META_PREPAID,
			Direction:       utils.OUT,
			Tenant:          "cgrates.org",
			Category:        "call",
			Account:         "1001",
			Subject:         "1001",
			Destination:     "1002",
			SetupTime:       time.Date(2015, 12, 12, 14, 52, 0, 0, time.UTC),
			PDD:             time.Duration(20) * time.Millisecond,
			AnswerTime:      time.Date(2015, 12, 12, 14, 52, 20, 0, time.UTC),
			Usage:           time.Duration(35) * time.Second,
			Supplier:        "SUPPLIER1",
			DisconnectCause: "NORMAL_DISCONNECT",
			ExtraFields:     map[string]string{"ExtraHeader1": "ExtraVal1", "ExtraHeader2": "ExtraVal2"},
			CostSource:      "testGetCDRs",
			Cost:            0.17,
		},
		&CDR{
			CGRID:           utils.Sha1("testevent1", time.Date(2015, 12, 12, 14, 52, 0, 0, time.UTC).String()),
			RunID:           "run2",
			OriginHost:      "127.0.0.1",
			Source:          "testGetCDRs",
			OriginID:        "testevent1",
			ToR:             utils.VOICE,
			RequestType:     utils.META_RATED,
			Direction:       utils.OUT,
			Tenant:          "cgrates.org",
			Category:        "call_derived",
			Account:         "1001",
			Subject:         "1002",
			Destination:     "1002",
			SetupTime:       time.Date(2015, 12, 12, 14, 52, 0, 0, time.UTC),
			PDD:             time.Duration(20) * time.Millisecond,
			AnswerTime:      time.Date(2015, 12, 12, 14, 52, 20, 0, time.UTC),
			Usage:           time.Duration(35) * time.Second,
			Supplier:        "SUPPLIER1",
			DisconnectCause: "NORMAL_DISCONNECT",
			ExtraFields:     map[string]string{"ExtraHeader1": "ExtraVal1", "ExtraHeader2": "ExtraVal2"},
			CostSource:      "testGetCDRs",
			Cost:            0.17,
		},
		&CDR{
			CGRID:           utils.Sha1("testevent2", time.Date(2015, 12, 29, 12, 58, 0, 0, time.UTC).String()),
			RunID:           utils.META_DEFAULT,
			OriginHost:      "192.168.1.12",
			Source:          "testGetCDRs",
			OriginID:        "testevent2",
			ToR:             utils.VOICE,
			RequestType:     utils.META_POSTPAID,
			Direction:       utils.OUT,
			Tenant:          "itsyscom.com",
			Category:        "call",
			Account:         "1004",
			Subject:         "1004",
			Destination:     "1007",
			SetupTime:       time.Date(2015, 12, 29, 12, 58, 0, 0, time.UTC),
			PDD:             time.Duration(10) * time.Millisecond,
			AnswerTime:      time.Date(2015, 12, 29, 12, 59, 0, 0, time.UTC),
			Usage:           time.Duration(0) * time.Second,
			Supplier:        "SUPPLIER1",
			DisconnectCause: "NO_ANSWER",
			ExtraFields:     map[string]string{"ExtraHeader1": "ExtraVal1", "ExtraHeader2": "ExtraVal2"},
			CostSource:      "rater1",
			Cost:            0,
		},
		&CDR{
			CGRID:           utils.Sha1("testevent3", time.Date(2015, 12, 28, 12, 58, 0, 0, time.UTC).String()),
			RunID:           utils.MetaRaw,
			OriginHost:      "192.168.1.13",
			Source:          "testGetCDRs3",
			OriginID:        "testevent3",
			ToR:             utils.VOICE,
			RequestType:     utils.META_PSEUDOPREPAID,
			Direction:       utils.OUT,
			Tenant:          "cgrates.org",
			Category:        "call",
			Account:         "1002",
			Subject:         "1002",
			Destination:     "1003",
			SetupTime:       time.Date(2015, 12, 28, 12, 58, 0, 0, time.UTC),
			PDD:             time.Duration(20) * time.Millisecond,
			AnswerTime:      time.Date(2015, 12, 28, 12, 58, 30, 0, time.UTC),
			Usage:           time.Duration(125) * time.Second,
			Supplier:        "SUPPLIER2",
			DisconnectCause: "NORMAL_DISCONNECT",
			ExtraFields:     map[string]string{},
			CostSource:      "",
			Cost:            -1,
		},
		&CDR{
			CGRID:           utils.Sha1("testevent3", time.Date(2015, 12, 28, 12, 58, 0, 0, time.UTC).String()),
			RunID:           utils.META_DEFAULT,
			OriginHost:      "192.168.1.13",
			Source:          "testGetCDRs3",
			OriginID:        "testevent3",
			ToR:             utils.VOICE,
			RequestType:     utils.META_RATED,
			Direction:       utils.OUT,
			Tenant:          "cgrates.org",
			Category:        "call",
			Account:         "1002",
			Subject:         "1002",
			Destination:     "1003",
			SetupTime:       time.Date(2015, 12, 28, 12, 58, 0, 0, time.UTC),
			PDD:             time.Duration(20) * time.Millisecond,
			AnswerTime:      time.Date(2015, 12, 28, 12, 58, 30, 0, time.UTC),
			Usage:           time.Duration(125) * time.Second,
			Supplier:        "SUPPLIER2",
			DisconnectCause: "NORMAL_DISCONNECT",
			ExtraFields:     map[string]string{},
			CostSource:      "testSetCDRs",
			Cost:            -1,
			ExtraInfo:       "AccountNotFound",
		},
		&CDR{
			CGRID:           utils.Sha1("testevent4", time.Date(2015, 12, 14, 14, 52, 0, 0, time.UTC).String()),
			RunID:           utils.MetaRaw,
			OriginHost:      "192.168.1.14",
			Source:          "testGetCDRs",
			OriginID:        "testevent4",
			ToR:             utils.VOICE,
			RequestType:     utils.META_PSEUDOPREPAID,
			Direction:       utils.OUT,
			Tenant:          "itsyscom.com",
			Category:        "call",
			Account:         "1003",
			Subject:         "1003",
			Destination:     "1007",
			SetupTime:       time.Date(2015, 12, 14, 14, 52, 0, 0, time.UTC),
			PDD:             time.Duration(2) * time.Second,
			AnswerTime:      time.Date(2015, 12, 12, 14, 52, 20, 0, time.UTC),
			Usage:           time.Duration(64) * time.Second,
			Supplier:        "SUPPLIER1",
			DisconnectCause: "NORMAL_DISCONNECT",
			ExtraFields:     map[string]string{"ExtraHeader3": "ExtraVal3"},
			CostSource:      "",
			Cost:            -1,
		},
		&CDR{
			CGRID:           utils.Sha1("testevent4", time.Date(2015, 12, 14, 14, 52, 0, 0, time.UTC).String()),
			RunID:           utils.META_DEFAULT,
			OriginHost:      "192.168.1.14",
			Source:          "testGetCDRs",
			OriginID:        "testevent4",
			ToR:             utils.VOICE,
			RequestType:     utils.META_RATED,
			Direction:       utils.OUT,
			Tenant:          "itsyscom.com",
			Category:        "call",
			Account:         "1003",
			Subject:         "1003",
			Destination:     "1007",
			SetupTime:       time.Date(2015, 12, 14, 14, 52, 0, 0, time.UTC),
			PDD:             time.Duration(2) * time.Second,
			AnswerTime:      time.Date(2015, 12, 12, 14, 52, 20, 0, time.UTC),
			Usage:           time.Duration(64) * time.Second,
			Supplier:        "SUPPLIER1",
			DisconnectCause: "NORMAL_DISCONNECT",
			ExtraFields:     map[string]string{"ExtraHeader3": "ExtraVal3"},
			CostSource:      "testSetCDRs",
			Cost:            1.205,
		},
		&CDR{
			CGRID:           utils.Sha1("testevent5", time.Date(2015, 12, 15, 18, 22, 0, 0, time.UTC).String()),
			RunID:           utils.MetaRaw,
			OriginHost:      "127.0.0.1",
			Source:          "testGetCDRs5",
			OriginID:        "testevent5",
			ToR:             utils.SMS,
			RequestType:     utils.META_PREPAID,
			Direction:       utils.OUT,
			Tenant:          "cgrates.org",
			Category:        "sms",
			Account:         "1001",
			Subject:         "1001",
			Destination:     "1002",
			SetupTime:       time.Date(2015, 12, 15, 18, 22, 0, 0, time.UTC),
			PDD:             time.Duration(0),
			AnswerTime:      time.Date(2015, 12, 15, 18, 22, 0, 0, time.UTC),
			Usage:           time.Duration(1) * time.Second,
			Supplier:        "SUPPLIER3",
			DisconnectCause: "SENT_OK",
			ExtraFields:     map[string]string{"Hdr4": "HdrVal4"},
			CostSource:      "",
			Cost:            -1,
		},
		&CDR{
			CGRID:           utils.Sha1("testevent5", time.Date(2015, 12, 15, 18, 22, 0, 0, time.UTC).String()),
			RunID:           utils.META_DEFAULT,
			OriginHost:      "127.0.0.1",
			Source:          "testGetCDRs5",
			OriginID:        "testevent5",
			ToR:             utils.SMS,
			RequestType:     utils.META_PREPAID,
			Direction:       utils.OUT,
			Tenant:          "cgrates.org",
			Category:        "sms",
			Account:         "1001",
			Subject:         "1001",
			Destination:     "1002",
			SetupTime:       time.Date(2015, 12, 15, 18, 22, 0, 0, time.UTC),
			PDD:             time.Duration(0),
			AnswerTime:      time.Date(2015, 12, 15, 18, 22, 0, 0, time.UTC),
			Usage:           time.Duration(1) * time.Second,
			Supplier:        "SUPPLIER3",
			DisconnectCause: "SENT_OK",
			ExtraFields:     map[string]string{"Hdr4": "HdrVal4"},
			CostSource:      "rater",
			Cost:            0.15,
		},
	}
	// Store all CDRs
	for _, cdr := range cdrs {
		if err := cdrStorage.SetCDR(cdr, false); err != nil {
			return fmt.Errorf("CDR: %+v, SetCDR err: %s", cdr, err.Error())
		}
	}
	// All CDRs, no filter
	if CDRs, _, err := cdrStorage.GetCDRs(new(utils.CDRsFilter)); err != nil {
		return err
	} else if len(CDRs) != 10 {
		return fmt.Errorf("GetCDRs, unexpected number of CDRs returned: %d", len(CDRs))
	}
	// Count ALL
	if CDRs, count, err := cdrStorage.GetCDRs(&utils.CDRsFilter{Count: true}); err != nil {
		return err
	} else if len(CDRs) != 0 {
		return fmt.Errorf("CountCDRs, unexpected number of CDRs returned: %+v", CDRs)
	} else if count != 10 {
		return fmt.Errorf("CountCDRs, unexpected count of CDRs returned: %+v", count)
	}
	// Limit 5
	if CDRs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{Paginator: utils.Paginator{Limit: utils.IntPointer(5), Offset: utils.IntPointer(0)}}); err != nil {
		return err
	} else if len(CDRs) != 5 {
		return fmt.Errorf("Limit 5, unexpected number of CDRs returned: %+v", CDRs)
	}
	// Offset 5
	if CDRs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{Paginator: utils.Paginator{Limit: utils.IntPointer(5), Offset: utils.IntPointer(0)}}); err != nil {
		return err
	} else if len(CDRs) != 5 {
		return fmt.Errorf("Offset 5, unexpected number of CDRs returned: %+v", CDRs)
	}
	// Offset with limit 2
	if CDRs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{Paginator: utils.Paginator{Limit: utils.IntPointer(2), Offset: utils.IntPointer(5)}}); err != nil {
		return err
	} else if len(CDRs) != 2 {
		return fmt.Errorf("Offset with limit 2, unexpected number of CDRs returned: %+v", CDRs)
	}
	// Filter on cgrids
	if CDRs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{CGRIDs: []string{
		utils.Sha1("testevent1", time.Date(2015, 12, 12, 14, 52, 0, 0, time.UTC).String()),
		utils.Sha1("testevent3", time.Date(2015, 12, 28, 12, 58, 0, 0, time.UTC).String()),
	}}); err != nil {
		return err
	} else if len(CDRs) != 5 {
		return fmt.Errorf("Filter on CGRIDs, unexpected number of CDRs returned: %+v", CDRs)
	}
	// Count on CGRIDS
	if _, count, err := cdrStorage.GetCDRs(&utils.CDRsFilter{CGRIDs: []string{
		utils.Sha1("testevent1", time.Date(2015, 12, 12, 14, 52, 0, 0, time.UTC).String()),
		utils.Sha1("testevent3", time.Date(2015, 12, 28, 12, 58, 0, 0, time.UTC).String()),
	}, Count: true}); err != nil {
		return err
	} else if count != 5 {
		return fmt.Errorf("Count on CGRIDs, unexpected count of CDRs returned: %d", count)
	}
	// Filter on cgrids plus reqType
	if CDRs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{CGRIDs: []string{
		utils.Sha1("testevent1", time.Date(2015, 12, 12, 14, 52, 0, 0, time.UTC).String()),
		utils.Sha1("testevent3", time.Date(2015, 12, 28, 12, 58, 0, 0, time.UTC).String()),
	}, RequestTypes: []string{utils.META_PREPAID}}); err != nil {
		return err
	} else if len(CDRs) != 2 {
		return fmt.Errorf("Filter on cgrids plus reqType, unexpected number of CDRs returned: %+v", CDRs)
	}
	// Count on multiple filter
	if _, count, err := cdrStorage.GetCDRs(&utils.CDRsFilter{CGRIDs: []string{
		utils.Sha1("testevent1", time.Date(2015, 12, 12, 14, 52, 0, 0, time.UTC).String()),
		utils.Sha1("testevent3", time.Date(2015, 12, 28, 12, 58, 0, 0, time.UTC).String()),
	}, RequestTypes: []string{utils.META_PREPAID}, Count: true}); err != nil {
		return err
	} else if count != 2 {
		return fmt.Errorf("Count on multiple filter, unexpected count of CDRs returned: %d", count)
	}
	// Filter on RunID
	if CDRs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{RunIDs: []string{utils.DEFAULT_RUNID}}); err != nil {
		return err
	} else if len(CDRs) != 5 {
		return fmt.Errorf("Filter on RunID, unexpected number of CDRs returned: %+v", CDRs)
	}
	// Filter on TOR
	if CDRs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{ToRs: []string{utils.SMS}}); err != nil {
		return err
	} else if len(CDRs) != 2 {
		return fmt.Errorf("Filter on TOR, unexpected number of CDRs returned: %+v", CDRs)
	}
	// Filter on multiple TOR
	if CDRs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{ToRs: []string{utils.SMS, utils.VOICE}}); err != nil {
		return err
	} else if len(CDRs) != 10 {
		return fmt.Errorf("Filter on multiple TOR, unexpected number of CDRs returned: %+v", CDRs)
	}
	// Filter on OriginHost
	if CDRs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{OriginHosts: []string{"127.0.0.1"}}); err != nil {
		return err
	} else if len(CDRs) != 5 {
		return fmt.Errorf("Filter on OriginHost, unexpected number of CDRs returned: %+v", CDRs)
	}
	// Filter on multiple OriginHost
	if CDRs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{OriginHosts: []string{"127.0.0.1", "192.168.1.12"}}); err != nil {
		return err
	} else if len(CDRs) != 6 {
		return fmt.Errorf("Filter on OriginHosts, unexpected number of CDRs returned: %+v", CDRs)
	}
	// Filter on Source
	if CDRs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{Sources: []string{"testGetCDRs"}}); err != nil {
		return err
	} else if len(CDRs) != 6 {
		return fmt.Errorf("Filter on Source, unexpected number of CDRs returned: %+v", CDRs)
	}
	// Filter on multiple Sources
	if CDRs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{Sources: []string{"testGetCDRs", "testGetCDRs5"}}); err != nil {
		return err
	} else if len(CDRs) != 8 {
		return fmt.Errorf("Filter on Sources, unexpected number of CDRs returned: %+v", CDRs)
	}
	// Filter on reqType
	if CDRs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{RequestTypes: []string{utils.META_PREPAID}}); err != nil {
		return err
	} else if len(CDRs) != 4 {
		return fmt.Errorf("Filter on RequestType, unexpected number of CDRs returned: %+v", len(CDRs))
	}
	// Filter on multiple reqType
	if CDRs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{RequestTypes: []string{utils.META_PREPAID, utils.META_PSEUDOPREPAID}}); err != nil {
		return err
	} else if len(CDRs) != 6 {
		return fmt.Errorf("Filter on RequestTypes, unexpected number of CDRs returned: %+v", CDRs)
	}

	// Filter on direction
	if CDRs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{Directions: []string{utils.OUT}}); err != nil {
		return err
	} else if len(CDRs) != 10 {
		return fmt.Errorf("Filter on Direction, unexpected number of CDRs returned: %+v", CDRs)
	}
	// Filter on Tenant
	if CDRs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{Tenants: []string{"itsyscom.com"}}); err != nil {
		return err
	} else if len(CDRs) != 3 {
		return fmt.Errorf("Filter on Tenant, unexpected number of CDRs returned: %+v", CDRs)
	}
	// Filter on multiple tenants
	if CDRs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{Tenants: []string{"itsyscom.com", "cgrates.org"}}); err != nil {
		return err
	} else if len(CDRs) != 10 {
		return fmt.Errorf("Filter on Tenants, Unexpected number of CDRs returned: %+v", CDRs)
	}
	// Filter on Category
	if CDRs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{Categories: []string{"call"}}); err != nil {
		return err
	} else if len(CDRs) != 7 {
		return fmt.Errorf("Filter on Category, unexpected number of CDRs returned: %+v", CDRs)
	}
	// Filter on multiple categories
	if CDRs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{Categories: []string{"sms", "call_derived"}}); err != nil {
		return err
	} else if len(CDRs) != 3 {
		return fmt.Errorf("Filter on Categories, unexpected number of CDRs returned:  %+v", CDRs)
	}
	// Filter on account
	if CDRs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{Accounts: []string{"1002"}}); err != nil {
		return err
	} else if len(CDRs) != 2 {
		return fmt.Errorf("Filter on Account, unexpected number of CDRs returned:  %+v", CDRs)
	}
	// Filter on multiple account
	if CDRs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{Accounts: []string{"1001", "1002"}}); err != nil {
		return err
	} else if len(CDRs) != 7 {
		return fmt.Errorf("Filter on Accounts, unexpected number of CDRs returned:  %+v", CDRs)
	}
	// Filter on subject
	if CDRs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{Subjects: []string{"1004"}}); err != nil {
		return err
	} else if len(CDRs) != 1 {
		return fmt.Errorf("Filter on Subject, unexpected number of CDRs returned:  %+v", CDRs)
	}
	// Filter on multiple subject
	if CDRs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{Subjects: []string{"1002", "1003"}}); err != nil {
		return err
	} else if len(CDRs) != 5 {
		return fmt.Errorf("Filter on Subjects, unexpected number of CDRs returned:  %+v", CDRs)
	}
	// Filter on destPrefix
	if CDRs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{DestinationPrefixes: []string{"10"}}); err != nil {
		return err
	} else if len(CDRs) != 10 {
		return fmt.Errorf("Filter on DestinationPrefix, unexpected number of CDRs returned: %+v", CDRs)
	}
	// Filter on multiple destPrefixes
	if CDRs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{DestinationPrefixes: []string{"1002", "1003"}}); err != nil {
		return err
	} else if len(CDRs) != 7 {
		return fmt.Errorf("Filter on DestinationPrefix, unexpected number of CDRs returned: %+v", CDRs)
	}
	// Filter on not destPrefix
	if CDRs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{NotDestinationPrefixes: []string{"10"}}); err != nil {
		return err
	} else if len(CDRs) != 0 {
		return fmt.Errorf("Filter on NotDestinationPrefix, unexpected number of CDRs returned: %+v", CDRs)
	}
	// Filter on not destPrefixes
	if CDRs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{NotDestinationPrefixes: []string{"1001", "1002"}}); err != nil {
		return err
	} else if len(CDRs) != 5 {
		return fmt.Errorf("Filter on NotDestinationPrefix, unexpected number of CDRs returned: %+v", CDRs)
	}
	// Filter on hasPrefix and not HasPrefix
	if CDRs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{DestinationPrefixes: []string{"1002", "1003"},
		NotDestinationPrefixes: []string{"1002"}}); err != nil {
		return err
	} else if len(CDRs) != 2 {
		return fmt.Errorf("Filter on DestinationPrefix and NotDestinationPrefix, unexpected number of CDRs returned: %+v", CDRs)
	}

	// Filter on MaxCost
	var orderIdStart, orderIdEnd int64 // Capture also orderIds for the next test
	if CDRs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{MaxCost: utils.Float64Pointer(0.0)}); err != nil {
		return err
	} else if len(CDRs) != 5 {
		return fmt.Errorf("Filter on MaxCost, unexpected number of CDRs returned: ", CDRs)
	} else {
		for i, cdr := range CDRs {
			if i == 0 {
				orderIdStart = cdr.OrderID
			}
			if cdr.OrderID < orderIdStart {
				orderIdStart = cdr.OrderID
			}
			if cdr.OrderID > orderIdEnd {
				orderIdEnd = cdr.OrderID
			}
		}
	}
	// Filter on orderIdStart
	if CDRs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{OrderIDStart: &orderIdStart}); err != nil {
		return err
	} else if len(CDRs) != 10 {
		return fmt.Errorf("Filter on OrderIDStart, unexpected number of CDRs returned: %d", len(CDRs))
	}
	// Filter on orderIdStart and orderIdEnd
	if CDRs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{OrderIDStart: &orderIdStart, OrderIDEnd: &orderIdEnd}); err != nil {
		return err
	} else if len(CDRs) != 8 {
		return fmt.Errorf("Filter on OrderIDStart OrderIDEnd, unexpected number of CDRs returned: %d", len(CDRs))
	}
	var timeStart, timeEnd time.Time
	// Filter on timeStart
	timeStart = time.Date(2015, 12, 28, 0, 0, 0, 0, time.UTC)
	if CDRs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{AnswerTimeStart: &timeStart}); err != nil {
		return err
	} else if len(CDRs) != 3 {
		return fmt.Errorf("Filter on AnswerTimeStart, unexpected number of CDRs returned: %d", len(CDRs))
	}
	// Filter on timeStart and timeEnd
	timeEnd = time.Date(2015, 12, 29, 0, 0, 0, 0, time.UTC)
	if CDRs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{AnswerTimeStart: &timeStart, AnswerTimeEnd: &timeEnd}); err != nil {
		return err
	} else if len(CDRs) != 2 {
		return fmt.Errorf("Filter on AnswerTimeStart AnswerTimeEnd, unexpected number of CDRs returned: %+v", CDRs)
	}
	// Filter on MinPDD
	if CDRs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{MinPDD: "20ms"}); err != nil {
		return err
	} else if len(CDRs) != 7 {
		return fmt.Errorf("Filter on MinPDD, unexpected number of CDRs returned: %+v", CDRs)
	}
	// Filter on maxPdd
	if CDRs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{MaxPDD: "1s"}); err != nil {
		return err
	} else if len(CDRs) != 8 {
		return fmt.Errorf("Filter on MaxPDD, unexpected number of CDRs returned: %+v", CDRs)
	}
	// Filter on minPdd, maxPdd
	if CDRs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{MinPDD: "10ms", MaxPDD: "1s"}); err != nil {
		return err
	} else if len(CDRs) != 6 {
		return fmt.Errorf("Filter on MinPDD MaxPDD, unexpected number of CDRs returned: %+v", CDRs)
	}
	// Combined filter
	if CDRs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{RequestTypes: []string{utils.META_RATED}, AnswerTimeStart: &timeStart, AnswerTimeEnd: &timeEnd}); err != nil {
		return err
	} else if len(CDRs) != 1 {
		return fmt.Errorf("Filter on RequestTypes AnswerTimeStart AnswerTimeEnd, unexpected number of CDRs returned: %+v", CDRs)
	}
	return nil
}
