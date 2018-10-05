// +build integration

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
	"errors"
	"fmt"
	"path"
	"strconv"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func TestITCDRsMySQL(t *testing.T) {
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
	cdrStorage, err := ConfigureCdrStorage(cfg.StorDbCfg().StorDBType,
		cfg.StorDbCfg().StorDBHost, cfg.StorDbCfg().StorDBPort,
		cfg.StorDbCfg().StorDBName, cfg.StorDbCfg().StorDBUser,
		cfg.StorDbCfg().StorDBPass, cfg.StorDbCfg().StorDBMaxOpenConns,
		cfg.StorDbCfg().StorDBMaxIdleConns, cfg.StorDbCfg().StorDBConnMaxLifetime,
		cfg.StorDbCfg().StorDBCDRSIndexes)
	if err != nil {
		return err
	}
	rawCDR := &CDR{
		CGRID:       utils.Sha1("testevent1", time.Date(2015, 12, 12, 14, 52, 0, 0, time.UTC).String()),
		RunID:       utils.MetaRaw,
		OrderID:     time.Now().UnixNano(),
		OriginHost:  "127.0.0.1",
		Source:      "testSetCDRs",
		OriginID:    "testevent1",
		ToR:         utils.VOICE,
		RequestType: utils.META_PREPAID,
		Tenant:      "cgrates.org",
		Category:    "call",
		Account:     "1004",
		Subject:     "1004",
		Destination: "1007",
		SetupTime:   time.Date(2015, 12, 12, 14, 52, 0, 0, time.UTC),
		AnswerTime:  time.Date(2015, 12, 12, 14, 52, 20, 0, time.UTC),
		Usage:       time.Duration(35) * time.Second,
		ExtraFields: map[string]string{"ExtraHeader1": "ExtraVal1", "ExtraHeader2": "ExtraVal2"},
		Cost:        -1,
	}
	if err := cdrStorage.SetCDR(rawCDR, false); err != nil {
		return fmt.Errorf("rawCDR: %+v, SetCDR err: %s", rawCDR, err.Error())
	}
	if cdrs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{CGRIDs: []string{rawCDR.CGRID}, RunIDs: []string{utils.MetaRaw}}, false); err != nil {
		return fmt.Errorf("rawCDR: %+v, GetCDRs err: %s", rawCDR, err.Error())
	} else if len(cdrs) != 1 {
		return fmt.Errorf("rawCDR %+v, Unexpected number of CDRs returned: %d", rawCDR, len(cdrs))
	}
	ratedCDR := &CDR{
		CGRID:       utils.Sha1("testevent1", time.Date(2015, 12, 12, 14, 52, 0, 0, time.UTC).String()),
		RunID:       utils.META_DEFAULT,
		OriginHost:  "127.0.0.1",
		Source:      "testSetCDRs",
		OriginID:    "testevent1",
		ToR:         utils.VOICE,
		RequestType: utils.META_PREPAID,
		Tenant:      "cgrates.org",
		Category:    "call",
		Account:     "1004",
		Subject:     "1004",
		Destination: "1007",
		SetupTime:   time.Date(2015, 12, 12, 14, 52, 0, 0, time.UTC),
		AnswerTime:  time.Date(2015, 12, 12, 14, 52, 20, 0, time.UTC),
		Usage:       time.Duration(35) * time.Second,
		ExtraFields: map[string]string{"ExtraHeader1": "ExtraVal1", "ExtraHeader2": "ExtraVal2"},
		CostSource:  "testSetCDRs",
		Cost:        0.17,
	}
	if err := cdrStorage.SetCDR(ratedCDR, false); err != nil {
		return fmt.Errorf("ratedCDR: %+v, SetCDR err: %s", ratedCDR, err.Error())
	}
	if cdrs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{CGRIDs: []string{ratedCDR.CGRID}, RunIDs: []string{ratedCDR.RunID}}, false); err != nil {
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
	if cdrs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{CGRIDs: []string{ratedCDR.CGRID}, RunIDs: []string{ratedCDR.RunID}}, false); err != nil {
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
		return fmt.Errorf("testSMCosts #1 err: %v", err)
	}
	cdrStorage, err := ConfigureCdrStorage(cfg.StorDbCfg().StorDBType,
		cfg.StorDbCfg().StorDBHost, cfg.StorDbCfg().StorDBPort,
		cfg.StorDbCfg().StorDBName, cfg.StorDbCfg().StorDBUser,
		cfg.StorDbCfg().StorDBPass, cfg.StorDbCfg().StorDBMaxOpenConns,
		cfg.StorDbCfg().StorDBMaxIdleConns, cfg.StorDbCfg().StorDBConnMaxLifetime,
		cfg.StorDbCfg().StorDBCDRSIndexes)
	if err != nil {
		return fmt.Errorf("testSMCosts #2 err: %v", err)
	}
	cc := &CallCost{
		Direction:   utils.OUT,
		Destination: "+4986517174963",
		Timespans: []*TimeSpan{
			{
				TimeStart:     time.Date(2015, 12, 28, 8, 53, 0, 0, time.UTC),
				TimeEnd:       time.Date(2015, 12, 28, 8, 54, 40, 0, time.UTC),
				DurationIndex: 0,
				RateInterval:  &RateInterval{Rating: &RIRate{Rates: RateGroups{&Rate{GroupIntervalStart: 0, Value: 100, RateIncrement: 10 * time.Second, RateUnit: time.Second}}}},
			},
		},
		TOR: utils.VOICE,
	}
	if err := cdrStorage.SetSMCost(&SMCost{CGRID: "164b0422fdc6a5117031b427439482c6a4f90e41",
		RunID: utils.META_DEFAULT, OriginHost: "localhost", OriginID: "12345", CostSource: utils.UNIT_TEST,
		CostDetails: NewEventCostFromCallCost(cc, "164b0422fdc6a5117031b427439482c6a4f90e41", utils.META_DEFAULT)}); err != nil {
		return fmt.Errorf("testSMCosts #3 err: %v", err)
	}
	if rcvSMC, err := cdrStorage.GetSMCosts("164b0422fdc6a5117031b427439482c6a4f90e41", utils.META_DEFAULT, "", ""); err != nil {
		return fmt.Errorf("testSMCosts #4 err: %v", err)
	} else if len(rcvSMC) == 0 {
		return errors.New("testSMCosts #5, no SMCosts received")
	}
	// Test query per prefix
	for i := 0; i < 3; i++ {
		if err := cdrStorage.SetSMCost(&SMCost{CGRID: "164b0422fdc6a5117031b427439482c6a4f90e5" + strconv.Itoa(i),
			RunID: utils.META_DEFAULT, OriginHost: "localhost", OriginID: "abc" + strconv.Itoa(i),
			CostSource:  utils.UNIT_TEST,
			CostDetails: NewEventCostFromCallCost(cc, "164b0422fdc6a5117031b427439482c6a4f90e5"+strconv.Itoa(i), utils.META_DEFAULT)}); err != nil {
			return fmt.Errorf("testSMCosts #7 err: %v", err)
		}
	}
	if rcvSMC, err := cdrStorage.GetSMCosts("", utils.META_DEFAULT, "localhost", "abc"); err != nil {
		return fmt.Errorf("testSMCosts #8 err: %v", err)
	} else if len(rcvSMC) != 3 {
		return fmt.Errorf("testSMCosts #9 expecting 3, received: %d", len(rcvSMC))
	}
	return nil
}

func testGetCDRs(cfg *config.CGRConfig) error {
	if err := InitStorDb(cfg); err != nil {
		return fmt.Errorf("testGetCDRs #1: %v", err)
	}
	cdrStorage, err := ConfigureCdrStorage(cfg.StorDbCfg().StorDBType,
		cfg.StorDbCfg().StorDBHost, cfg.StorDbCfg().StorDBPort,
		cfg.StorDbCfg().StorDBName, cfg.StorDbCfg().StorDBUser,
		cfg.StorDbCfg().StorDBPass, cfg.StorDbCfg().StorDBMaxOpenConns,
		cfg.StorDbCfg().StorDBMaxIdleConns, cfg.StorDbCfg().StorDBConnMaxLifetime,
		cfg.StorDbCfg().StorDBCDRSIndexes)
	if err != nil {
		return fmt.Errorf("testGetCDRs #2: %v", err)
	}
	// All CDRs, no filter
	if _, _, err := cdrStorage.GetCDRs(new(utils.CDRsFilter), false); err == nil || err.Error() != utils.NotFoundCaps {
		return fmt.Errorf("testGetCDRs #3: %v", err)
	}
	cdrs := []*CDR{
		{
			CGRID:       utils.Sha1("testevent1", time.Date(2015, 12, 12, 14, 52, 0, 0, time.UTC).String()),
			RunID:       utils.MetaRaw,
			OriginHost:  "127.0.0.1",
			Source:      "testGetCDRs",
			OriginID:    "testevent1",
			ToR:         utils.VOICE,
			RequestType: utils.META_PREPAID,
			Tenant:      "cgrates.org",
			Category:    "call",
			Account:     "1001",
			Subject:     "1001",
			Destination: "1002",
			SetupTime:   time.Date(2015, 12, 12, 14, 52, 0, 0, time.UTC),
			AnswerTime:  time.Date(2015, 12, 12, 14, 52, 20, 0, time.UTC),
			Usage:       time.Duration(35) * time.Second,
			ExtraFields: map[string]string{"ExtraHeader1": "ExtraVal1", "ExtraHeader2": "ExtraVal2"},
			CostSource:  "",
			Cost:        -1,
		},
		{
			CGRID:       utils.Sha1("testevent1", time.Date(2015, 12, 12, 14, 52, 0, 0, time.UTC).String()),
			RunID:       utils.META_DEFAULT,
			OriginHost:  "127.0.0.1",
			Source:      "testGetCDRs",
			OriginID:    "testevent1",
			ToR:         utils.VOICE,
			RequestType: utils.META_PREPAID,
			Tenant:      "cgrates.org",
			Category:    "call",
			Account:     "1001",
			Subject:     "1001",
			Destination: "1002",
			SetupTime:   time.Date(2015, 12, 12, 14, 52, 0, 0, time.UTC),
			AnswerTime:  time.Date(2015, 12, 12, 14, 52, 20, 0, time.UTC),
			Usage:       time.Duration(35) * time.Second,
			ExtraFields: map[string]string{"ExtraHeader1": "ExtraVal1", "ExtraHeader2": "ExtraVal2"},
			CostSource:  "testGetCDRs",
			Cost:        0.17,
		},
		{
			CGRID:       utils.Sha1("testevent1", time.Date(2015, 12, 12, 14, 52, 0, 0, time.UTC).String()),
			RunID:       "run2",
			OriginHost:  "127.0.0.1",
			Source:      "testGetCDRs",
			OriginID:    "testevent1",
			ToR:         utils.VOICE,
			RequestType: utils.META_RATED,
			Tenant:      "cgrates.org",
			Category:    "call_derived",
			Account:     "1001",
			Subject:     "1002",
			Destination: "1002",
			SetupTime:   time.Date(2015, 12, 12, 14, 52, 0, 0, time.UTC),
			AnswerTime:  time.Date(2015, 12, 12, 14, 52, 20, 0, time.UTC),
			Usage:       time.Duration(35) * time.Second,
			ExtraFields: map[string]string{"ExtraHeader1": "ExtraVal1", "ExtraHeader2": "ExtraVal2"},
			CostSource:  "testGetCDRs",
			Cost:        0.17,
		},
		{
			CGRID:       utils.Sha1("testevent2", time.Date(2015, 12, 29, 12, 58, 0, 0, time.UTC).String()),
			RunID:       utils.META_DEFAULT,
			OriginHost:  "192.168.1.12",
			Source:      "testGetCDRs",
			OriginID:    "testevent2",
			ToR:         utils.VOICE,
			RequestType: utils.META_POSTPAID,
			Tenant:      "itsyscom.com",
			Category:    "call",
			Account:     "1004",
			Subject:     "1004",
			Destination: "1007",
			SetupTime:   time.Date(2015, 12, 29, 12, 58, 0, 0, time.UTC),
			AnswerTime:  time.Date(2015, 12, 29, 12, 59, 0, 0, time.UTC),
			Usage:       time.Duration(0) * time.Second,
			ExtraFields: map[string]string{"ExtraHeader1": "ExtraVal1", "ExtraHeader2": "ExtraVal2"},
			CostSource:  "rater1",
			Cost:        0,
		},
		{
			CGRID:       utils.Sha1("testevent3", time.Date(2015, 12, 28, 12, 58, 0, 0, time.UTC).String()),
			RunID:       utils.MetaRaw,
			OriginHost:  "192.168.1.13",
			Source:      "testGetCDRs3",
			OriginID:    "testevent3",
			ToR:         utils.VOICE,
			RequestType: utils.META_PSEUDOPREPAID,
			Tenant:      "cgrates.org",
			Category:    "call",
			Account:     "1002",
			Subject:     "1002",
			Destination: "1003",
			SetupTime:   time.Date(2015, 12, 28, 12, 58, 0, 0, time.UTC),
			AnswerTime:  time.Date(2015, 12, 28, 12, 58, 30, 0, time.UTC),
			Usage:       time.Duration(125) * time.Second,
			ExtraFields: map[string]string{},
			CostSource:  "",
			Cost:        -1,
		},
		{
			CGRID:       utils.Sha1("testevent3", time.Date(2015, 12, 28, 12, 58, 0, 0, time.UTC).String()),
			RunID:       utils.META_DEFAULT,
			OriginHost:  "192.168.1.13",
			Source:      "testGetCDRs3",
			OriginID:    "testevent3",
			ToR:         utils.VOICE,
			RequestType: utils.META_RATED,
			Tenant:      "cgrates.org",
			Category:    "call",
			Account:     "1002",
			Subject:     "1002",
			Destination: "1003",
			SetupTime:   time.Date(2015, 12, 28, 12, 58, 0, 0, time.UTC),
			AnswerTime:  time.Date(2015, 12, 28, 12, 58, 30, 0, time.UTC),
			Usage:       time.Duration(125) * time.Second,
			ExtraFields: map[string]string{},
			CostSource:  "testSetCDRs",
			Cost:        -1,
			ExtraInfo:   "AccountNotFound",
		},
		{
			CGRID:       utils.Sha1("testevent4", time.Date(2015, 12, 14, 14, 52, 0, 0, time.UTC).String()),
			RunID:       utils.MetaRaw,
			OriginHost:  "192.168.1.14",
			Source:      "testGetCDRs",
			OriginID:    "testevent4",
			ToR:         utils.VOICE,
			RequestType: utils.META_PSEUDOPREPAID,
			Tenant:      "itsyscom.com",
			Category:    "call",
			Account:     "1003",
			Subject:     "1003",
			Destination: "1007",
			SetupTime:   time.Date(2015, 12, 14, 14, 52, 0, 0, time.UTC),
			AnswerTime:  time.Date(2015, 12, 12, 14, 52, 20, 0, time.UTC),
			Usage:       time.Duration(64) * time.Second,
			ExtraFields: map[string]string{"ExtraHeader3": "ExtraVal3"},
			CostSource:  "",
			Cost:        -1,
		},
		{
			CGRID:       utils.Sha1("testevent4", time.Date(2015, 12, 14, 14, 52, 0, 0, time.UTC).String()),
			RunID:       utils.META_DEFAULT,
			OriginHost:  "192.168.1.14",
			Source:      "testGetCDRs",
			OriginID:    "testevent4",
			ToR:         utils.VOICE,
			RequestType: utils.META_RATED,
			Tenant:      "itsyscom.com",
			Category:    "call",
			Account:     "1003",
			Subject:     "1003",
			Destination: "1007",
			SetupTime:   time.Date(2015, 12, 14, 14, 52, 0, 0, time.UTC),
			AnswerTime:  time.Date(2015, 12, 12, 14, 52, 20, 0, time.UTC),
			Usage:       time.Duration(64) * time.Second,
			ExtraFields: map[string]string{"ExtraHeader3": "ExtraVal3"},
			CostSource:  "testSetCDRs",
			Cost:        1.205,
		},
		{
			CGRID:       utils.Sha1("testevent5", time.Date(2015, 12, 15, 18, 22, 0, 0, time.UTC).String()),
			RunID:       utils.MetaRaw,
			OriginHost:  "127.0.0.1",
			Source:      "testGetCDRs5",
			OriginID:    "testevent5",
			ToR:         utils.SMS,
			RequestType: utils.META_PREPAID,
			Tenant:      "cgrates.org",
			Category:    "sms",
			Account:     "1001",
			Subject:     "1001",
			Destination: "1002",
			SetupTime:   time.Date(2015, 12, 15, 18, 22, 0, 0, time.UTC),
			AnswerTime:  time.Date(2015, 12, 15, 18, 22, 0, 0, time.UTC),
			Usage:       time.Duration(1) * time.Second,
			ExtraFields: map[string]string{"Service-Context-Id": "voice@huawei.com"},
			CostSource:  "",
			Cost:        -1,
		},
		{
			CGRID:       utils.Sha1("testevent5", time.Date(2015, 12, 15, 18, 22, 0, 0, time.UTC).String()),
			RunID:       utils.META_DEFAULT,
			OriginHost:  "127.0.0.1",
			Source:      "testGetCDRs5",
			OriginID:    "testevent5",
			ToR:         utils.SMS,
			RequestType: utils.META_PREPAID,
			Tenant:      "cgrates.org",
			Category:    "sms",
			Account:     "1001",
			Subject:     "1001",
			Destination: "1002",
			SetupTime:   time.Date(2015, 12, 15, 18, 22, 0, 0, time.UTC),
			AnswerTime:  time.Date(2015, 12, 15, 18, 22, 0, 0, time.UTC),
			Usage:       time.Duration(1) * time.Second,
			ExtraFields: map[string]string{"Service-Context-Id": "voice2@huawei.com"},
			CostSource:  "rater",
			Cost:        0.15,
		},
	}
	// Store all CDRs
	for _, cdr := range cdrs {
		if err := cdrStorage.SetCDR(cdr, false); err != nil {
			return fmt.Errorf("testGetCDRs #4 CDR: %+v, err: %v", cdr, err)
		}
	}
	// All CDRs, no filter
	if CDRs, _, err := cdrStorage.GetCDRs(new(utils.CDRsFilter), false); err != nil {
		return err
	} else if len(CDRs) != 10 {
		return fmt.Errorf("testGetCDRs #5, unexpected number of CDRs returned: %d", len(CDRs))
	}
	// Count ALL
	if CDRs, count, err := cdrStorage.GetCDRs(&utils.CDRsFilter{Count: true}, false); err != nil {
		return fmt.Errorf("testGetCDRs #6 err: %v", err)
	} else if len(CDRs) != 0 {
		return fmt.Errorf("testGetCDRs #7, unexpected number of CDRs returned: %+v", CDRs)
	} else if count != 10 {
		return fmt.Errorf("testGetCDRs #8, unexpected count of CDRs returned: %+v", count)
	}
	// Limit 5
	if CDRs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{Paginator: utils.Paginator{Limit: utils.IntPointer(5), Offset: utils.IntPointer(0)}}, false); err != nil {
		return fmt.Errorf("testGetCDRs #9 err: %v", err)
	} else if len(CDRs) != 5 {
		return fmt.Errorf("testGetCDRs #10, unexpected number of CDRs returned: %+v", CDRs)
	}
	// Offset 5
	if CDRs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{Paginator: utils.Paginator{Limit: utils.IntPointer(5), Offset: utils.IntPointer(0)}}, false); err != nil {
		return err
	} else if len(CDRs) != 5 {
		return fmt.Errorf("testGetCDRs #11, unexpected number of CDRs returned: %+v", CDRs)
	}
	// Offset with limit 2
	if CDRs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{Paginator: utils.Paginator{Limit: utils.IntPointer(2), Offset: utils.IntPointer(5)}}, false); err != nil {
		return fmt.Errorf("testGetCDRs #12 err: %v", err)
	} else if len(CDRs) != 2 {
		return fmt.Errorf("testGetCDRs #13, unexpected number of CDRs returned: %+v", CDRs)
	}
	// Filter on cgrids
	if CDRs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{CGRIDs: []string{
		utils.Sha1("testevent1", time.Date(2015, 12, 12, 14, 52, 0, 0, time.UTC).String()),
		utils.Sha1("testevent3", time.Date(2015, 12, 28, 12, 58, 0, 0, time.UTC).String()),
	}}, false); err != nil {
		return fmt.Errorf("testGetCDRs #14 err: %v", err)
	} else if len(CDRs) != 5 {
		return fmt.Errorf("testGetCDRs #15, unexpected number of CDRs returned: %+v", CDRs)
	}
	// Count on CGRIDS
	if _, count, err := cdrStorage.GetCDRs(&utils.CDRsFilter{CGRIDs: []string{
		utils.Sha1("testevent1", time.Date(2015, 12, 12, 14, 52, 0, 0, time.UTC).String()),
		utils.Sha1("testevent3", time.Date(2015, 12, 28, 12, 58, 0, 0, time.UTC).String()),
	}, Count: true}, false); err != nil {
		return fmt.Errorf("testGetCDRs #16 err: %v", err)
	} else if count != 5 {
		return fmt.Errorf("testGetCDRs #17, unexpected count of CDRs returned: %d", count)
	}
	// Filter on cgrids plus reqType
	if CDRs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{CGRIDs: []string{
		utils.Sha1("testevent1", time.Date(2015, 12, 12, 14, 52, 0, 0, time.UTC).String()),
		utils.Sha1("testevent3", time.Date(2015, 12, 28, 12, 58, 0, 0, time.UTC).String()),
	}, RequestTypes: []string{utils.META_PREPAID}}, false); err != nil {
		return fmt.Errorf("testGetCDRs #18 err: %v", err)
	} else if len(CDRs) != 2 {
		return fmt.Errorf("testGetCDRs #19, unexpected number of CDRs returned: %+v", CDRs)
	}
	// Count on multiple filter
	if _, count, err := cdrStorage.GetCDRs(&utils.CDRsFilter{CGRIDs: []string{
		utils.Sha1("testevent1", time.Date(2015, 12, 12, 14, 52, 0, 0, time.UTC).String()),
		utils.Sha1("testevent3", time.Date(2015, 12, 28, 12, 58, 0, 0, time.UTC).String()),
	}, RequestTypes: []string{utils.META_PREPAID}, Count: true}, false); err != nil {
		return fmt.Errorf("testGetCDRs #20 err: %v", err)
	} else if count != 2 {
		return fmt.Errorf("testGetCDRs #21, unexpected count of CDRs returned: %d", count)
	}
	// Filter on RunID
	if CDRs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{RunIDs: []string{utils.DEFAULT_RUNID}}, false); err != nil {
		return fmt.Errorf("testGetCDRs #22 err: %v", err)
	} else if len(CDRs) != 5 {
		return fmt.Errorf("testGetCDRs #23, unexpected number of CDRs returned: %+v", CDRs)
	}
	// Filter on OriginID
	if CDRs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{OriginIDs: []string{
		"testevent1", "testevent3"}}, false); err != nil {
		return fmt.Errorf("testGetCDRs #22 err: %v", err)
	} else if len(CDRs) != 5 {
		return fmt.Errorf("testGetCDRs #23, unexpected number of CDRs returned: %+v", CDRs)
	}
	// Filter on TOR
	if CDRs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{ToRs: []string{utils.SMS}}, false); err != nil {
		return fmt.Errorf("testGetCDRs #23 err: %v", err)
	} else if len(CDRs) != 2 {
		return fmt.Errorf("testGetCDRs #24, unexpected number of CDRs returned: %+v", CDRs)
	}
	// Filter on multiple TOR
	if CDRs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{ToRs: []string{utils.SMS, utils.VOICE}}, false); err != nil {
		return fmt.Errorf("testGetCDRs #25 err: %v", err)
	} else if len(CDRs) != 10 {
		return fmt.Errorf("testGetCDRs #26, unexpected number of CDRs returned: %+v", CDRs)
	}
	// Filter on OriginHost
	if CDRs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{OriginHosts: []string{"127.0.0.1"}}, false); err != nil {
		return fmt.Errorf("testGetCDRs #27 err: %v", err)
	} else if len(CDRs) != 5 {
		return fmt.Errorf("testGetCDRs #28, unexpected number of CDRs returned: %+v", CDRs)
	}
	// Filter on multiple OriginHost
	if CDRs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{OriginHosts: []string{"127.0.0.1", "192.168.1.12"}}, false); err != nil {
		return fmt.Errorf("testGetCDRs #29 err: %v", err)
	} else if len(CDRs) != 6 {
		return fmt.Errorf("Filter on OriginHosts, unexpected number of CDRs returned: %+v", CDRs)
	}
	// Filter on Source
	if CDRs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{Sources: []string{"testGetCDRs"}}, false); err != nil {
		return fmt.Errorf("testGetCDRs #30 err: %v", err)
	} else if len(CDRs) != 6 {
		return fmt.Errorf("testGetCDRs #31, unexpected number of CDRs returned: %+v", CDRs)
	}
	// Filter on multiple Sources
	if CDRs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{Sources: []string{"testGetCDRs", "testGetCDRs5"}}, false); err != nil {
		return fmt.Errorf("testGetCDRs #32 err: %v", err)
	} else if len(CDRs) != 8 {
		return fmt.Errorf("testGetCDRs #33, unexpected number of CDRs returned: %+v", CDRs)
	}
	// Filter on reqType
	if CDRs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{RequestTypes: []string{utils.META_PREPAID}}, false); err != nil {
		return fmt.Errorf("testGetCDRs #32 err: %v", err)
	} else if len(CDRs) != 4 {
		return fmt.Errorf("testGetCDRs #33, unexpected number of CDRs returned: %+v", len(CDRs))
	}
	// Filter on multiple reqType
	if CDRs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{RequestTypes: []string{utils.META_PREPAID, utils.META_PSEUDOPREPAID}}, false); err != nil {
		return fmt.Errorf("testGetCDRs #34 err: %v", err)
	} else if len(CDRs) != 6 {
		return fmt.Errorf("testGetCDRs #35, unexpected number of CDRs returned: %+v", CDRs)
	}
	// Filter on Tenant
	if CDRs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{Tenants: []string{"itsyscom.com"}}, false); err != nil {
		return fmt.Errorf("testGetCDRs #38 err: %v", err)
	} else if len(CDRs) != 3 {
		return fmt.Errorf("testGetCDRs #39, unexpected number of CDRs returned: %+v", CDRs)
	}
	// Filter on multiple tenants
	if CDRs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{Tenants: []string{"itsyscom.com", "cgrates.org"}}, false); err != nil {
		return fmt.Errorf("testGetCDRs #40 err: %v", err)
	} else if len(CDRs) != 10 {
		return fmt.Errorf("testGetCDRs #41, Unexpected number of CDRs returned: %+v", CDRs)
	}
	// Filter on Category
	if CDRs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{Categories: []string{"call"}}, false); err != nil {
		return fmt.Errorf("testGetCDRs #42 err: %v", err)
	} else if len(CDRs) != 7 {
		return fmt.Errorf("testGetCDRs #43 err, unexpected number of CDRs returned: %+v", CDRs)
	}
	// Filter on multiple categories
	if CDRs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{Categories: []string{"sms", "call_derived"}}, false); err != nil {
		return fmt.Errorf("testGetCDRs #44 err: %v", err)
	} else if len(CDRs) != 3 {
		return fmt.Errorf("testGetCDRs #45 err, unexpected number of CDRs returned:  %+v", CDRs)
	}
	// Filter on account
	if CDRs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{Accounts: []string{"1002"}}, false); err != nil {
		return fmt.Errorf("testGetCDRs #46 err: %v", err)
	} else if len(CDRs) != 2 {
		return fmt.Errorf("testGetCDRs #47, unexpected number of CDRs returned:  %+v", CDRs)
	}
	// Filter on multiple account
	if CDRs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{Accounts: []string{"1001", "1002"}}, false); err != nil {
		return fmt.Errorf("testGetCDRs #48 err: %v", err)
	} else if len(CDRs) != 7 {
		return fmt.Errorf("testGetCDRs #49, unexpected number of CDRs returned:  %+v", CDRs)
	}
	// Filter on subject
	if CDRs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{Subjects: []string{"1004"}}, false); err != nil {
		return fmt.Errorf("testGetCDRs #50, err: %v", err)
	} else if len(CDRs) != 1 {
		return fmt.Errorf("testGetCDRs #51, unexpected number of CDRs returned:  %+v", CDRs)
	}
	// Filter on multiple subject
	if CDRs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{Subjects: []string{"1002", "1003"}}, false); err != nil {
		return fmt.Errorf("testGetCDRs #52, err: %v", err)
	} else if len(CDRs) != 5 {
		return fmt.Errorf("testGetCDRs #53, unexpected number of CDRs returned:  %+v", CDRs)
	}
	// Filter on destPrefix
	if CDRs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{DestinationPrefixes: []string{"10"}}, false); err != nil {
		return fmt.Errorf("testGetCDRs #54, err: %v", err)
	} else if len(CDRs) != 10 {
		return fmt.Errorf("testGetCDRs #55, unexpected number of CDRs returned: %+v", CDRs)
	}
	// Filter on multiple destPrefixes
	if CDRs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{DestinationPrefixes: []string{"1002", "1003"}}, false); err != nil {
		return fmt.Errorf("testGetCDRs #56, err: %v", err)
	} else if len(CDRs) != 7 {
		return fmt.Errorf("testGetCDRs #57, unexpected number of CDRs returned: %+v", CDRs)
	}
	// Filter on not destPrefix
	if CDRs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{NotDestinationPrefixes: []string{"10"}}, false); err == nil || err.Error() != utils.NotFoundCaps {
		return fmt.Errorf("testGetCDRs #58, err: %v", err)
	} else if len(CDRs) != 0 {
		return fmt.Errorf("testGetCDRs #59, unexpected number of CDRs returned: %+v", CDRs)
	}
	// Filter on not destPrefixes
	if CDRs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{NotDestinationPrefixes: []string{"1001", "1002"}}, false); err != nil {
		return fmt.Errorf("testGetCDRs #60, err: %v", err)
	} else if len(CDRs) != 5 {
		return fmt.Errorf("testGetCDRs #61, unexpected number of CDRs returned: %+v", CDRs)
	}
	// Filter on hasPrefix and not HasPrefix
	if CDRs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{DestinationPrefixes: []string{"1002", "1003"},
		NotDestinationPrefixes: []string{"1002"}}, false); err != nil {
		return fmt.Errorf("testGetCDRs #62, err: %v", err)
	} else if len(CDRs) != 2 {
		return fmt.Errorf("testGetCDRs #63, unexpected number of CDRs returned: %+v", CDRs)
	}
	// Filter on MinUsage
	if CDRs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{MinUsage: "125s"}, false); err != nil {
		return fmt.Errorf("testGetCDRs #64, err: %v", err)
	} else if len(CDRs) != 2 {
		return fmt.Errorf("testGetCDRs #65, unexpected number of CDRs returned: %d", len(CDRs))
	}
	// Filter on MaxUsage
	if CDRs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{MaxUsage: "1ms"}, false); err != nil {
		return fmt.Errorf("testGetCDRs #66, err: %v", err)
	} else if len(CDRs) != 1 {
		return fmt.Errorf("testGetCDRs #67, unexpected number of CDRs returned: %d", len(CDRs))
	}
	// Filter on MaxCost
	var orderIdStart, orderIdEnd int64 // Capture also orderIds for the next test
	if CDRs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{MaxCost: utils.Float64Pointer(0.0)}, false); err != nil {
		return fmt.Errorf("testGetCDRs #68, err: %v", err)
	} else if len(CDRs) != 5 {
		return fmt.Errorf("testGetCDRs #69, unexpected number of CDRs returned: %+v", CDRs)
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
	if CDRs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{OrderIDStart: &orderIdStart}, false); err != nil {
		return fmt.Errorf("testGetCDRs #70, err: %v", err)
	} else if len(CDRs) != 10 {
		return fmt.Errorf("testGetCDRs #71, unexpected number of CDRs returned: %d", len(CDRs))
	}
	// Filter on orderIdStart and orderIdEnd
	if CDRs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{OrderIDStart: &orderIdStart, OrderIDEnd: &orderIdEnd}, false); err != nil {
		return fmt.Errorf("testGetCDRs #72, err: %v", err)
	} else if len(CDRs) != 8 {
		return fmt.Errorf("testGetCDRs #73, unexpected number of CDRs returned: %d", len(CDRs))
	}
	var timeStart, timeEnd time.Time
	// Filter on timeStart
	timeStart = time.Date(2015, 12, 28, 0, 0, 0, 0, time.UTC)
	if CDRs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{AnswerTimeStart: &timeStart}, false); err != nil {
		return fmt.Errorf("testGetCDRs #74, err: %v", err)
	} else if len(CDRs) != 3 {
		return fmt.Errorf("testGetCDRs #75, unexpected number of CDRs returned: %d", len(CDRs))
	}
	// Filter on timeStart and timeEnd
	timeEnd = time.Date(2015, 12, 29, 0, 0, 0, 0, time.UTC)
	if CDRs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{AnswerTimeStart: &timeStart, AnswerTimeEnd: &timeEnd}, false); err != nil {
		return fmt.Errorf("testGetCDRs #76, err: %v", err)
	} else if len(CDRs) != 2 {
		return fmt.Errorf("testGetCDRs #77, unexpected number of CDRs returned: %+v", CDRs)
	}
	// Combined filter
	if CDRs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{RequestTypes: []string{utils.META_RATED}, AnswerTimeStart: &timeStart, AnswerTimeEnd: &timeEnd}, false); err != nil {
		return fmt.Errorf("testGetCDRs #84, err: %v", err)
	} else if len(CDRs) != 1 {
		return fmt.Errorf("testGetCDRs #85, unexpected number of CDRs returned: %+v", CDRs)
	}
	// Remove CDRs
	if _, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{RequestTypes: []string{utils.META_RATED}, AnswerTimeStart: &timeStart, AnswerTimeEnd: &timeEnd}, true); err != nil {
		return fmt.Errorf("testGetCDRs #86, err: %v", err)
	}
	// All CDRs, no filter
	if cdrs, _, err := cdrStorage.GetCDRs(new(utils.CDRsFilter), false); err != nil {
		return fmt.Errorf("testGetCDRs #87, err: %v", err)
	} else if len(cdrs) != 9 {
		return fmt.Errorf("testGetCDRs #88, unexpected number of CDRs returned after remove: %d", len(cdrs))
	}
	// Filter on ExtraFields
	if CDRs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{ExtraFields: map[string]string{"Service-Context-Id": "voice@huawei.com"}}, false); err != nil {
		return fmt.Errorf("testGetCDRs #89, err: %v", err)
	} else if len(CDRs) != 1 {
		return fmt.Errorf("testGetCDRs #90, unexpected number of CDRs returned:  %+v", CDRs)
	}
	// Filter *exists on ExtraFields
	if CDRs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{ExtraFields: map[string]string{"Service-Context-Id": "*exists"}}, false); err != nil {
		return fmt.Errorf("testGetCDRs #91, err: %v", err)
	} else if len(CDRs) != 2 {
		return fmt.Errorf("testGetCDRs #92, unexpected number of CDRs returned:  %+v", CDRs)
	}
	// Filter *exists on not ExtraFields
	if CDRs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{NotExtraFields: map[string]string{"Service-Context-Id": "*exists"}}, false); err != nil {
		return fmt.Errorf("testGetCDRs #93, err: %v", err)
	} else if len(CDRs) != 7 {
		return fmt.Errorf("testGetCDRs #94, unexpected number of CDRs returned:  %+v", len(CDRs))
	}
	//Filter by OrderID descendent
	if CDRs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{OrderBy: "OrderID;desc"}, false); err != nil {
		return fmt.Errorf("testGetCDRs #95, err: %v", err)
	} else {
		for i := range CDRs {
			if i+1 > len(CDRs)-1 {
				break
			}
			if CDRs[i].OrderID < CDRs[i+1].OrderID {
				return fmt.Errorf("%+v should be greater than %+v \n", CDRs[i].OrderID, CDRs[i+1].OrderID)
			}
		}
	}
	//Filter by OrderID ascendent
	if CDRs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{OrderBy: "OrderID"}, false); err != nil {
		return fmt.Errorf("testGetCDRs #95, err: %v", err)
	} else {
		for i := range CDRs {
			if i+1 > len(CDRs)-1 {
				break
			}
			if CDRs[i].OrderID > CDRs[i+1].OrderID {
				return fmt.Errorf("%+v sould be smaller than %+v \n", CDRs[i].OrderID, CDRs[i+1].OrderID)
			}
		}
	}
	//Filter by Cost descendent
	if CDRs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{OrderBy: "Cost;desc"}, false); err != nil {
		return fmt.Errorf("testGetCDRs #95, err: %v", err)
	} else {
		for i := range CDRs {
			if i+1 > len(CDRs)-1 {
				break
			}
			if CDRs[i].Cost < CDRs[i+1].Cost {
				return fmt.Errorf("%+v should be greater than %+v \n", CDRs[i].Cost, CDRs[i+1].Cost)
			}
		}
	}
	//Filter by Cost ascendent
	if CDRs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{OrderBy: "Cost"}, false); err != nil {
		return fmt.Errorf("testGetCDRs #95, err: %v", err)
	} else {
		for i := range CDRs {
			if i+1 > len(CDRs)-1 {
				break
			}
			if CDRs[i].Cost > CDRs[i+1].Cost {
				return fmt.Errorf("%+v sould be smaller than %+v \n", CDRs[i].Cost, CDRs[i+1].Cost)
			}
		}
	}
	return nil
}
