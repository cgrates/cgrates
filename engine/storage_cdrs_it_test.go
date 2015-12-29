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
	if storedCdrs, _, err := cdrStorage.GetCDRs(new(utils.CDRsFilter)); err != nil {
		return err
	} else if len(storedCdrs) != 0 {
		return fmt.Errorf("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	/*
		// Count ALL
		if storedCdrs, count, err := cdrStorage.GetCDRs(&utils.CDRsFilter{Count: true}); err != nil {
			t.Error(err.Error())
		} else if len(storedCdrs) != 0 {
			t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
		} else if count != 8 {
			t.Error("Unexpected count of StoredCdrs returned: ", count)
		}
		// Limit 5
		if storedCdrs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{Paginator: utils.Paginator{Limit: utils.IntPointer(5), Offset: utils.IntPointer(0)}}); err != nil {
			t.Error(err.Error())
		} else if len(storedCdrs) != 5 {
			t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
		}
		// Offset 5
		if storedCdrs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{Paginator: utils.Paginator{Limit: utils.IntPointer(5), Offset: utils.IntPointer(0)}}); err != nil {
			t.Error(err.Error())
		} else if len(storedCdrs) != 5 {
			t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
		}
		// Offset with limit 2
		if storedCdrs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{Paginator: utils.Paginator{Limit: utils.IntPointer(2), Offset: utils.IntPointer(5)}}); err != nil {
			t.Error(err.Error())
		} else if len(storedCdrs) != 2 {
			t.Error("Unexpected number of StoredCdrs returned: ", len(storedCdrs))
		}
		// Filter on cgrids
		if storedCdrs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{CgrIds: []string{utils.Sha1("bbb1", time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC).String()),
			utils.Sha1("bbb2", time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC).String())}}); err != nil {
			t.Error(err.Error())
		} else if len(storedCdrs) != 2 {
			t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
		}
		// Count on CGRIDS
		if _, count, err := cdrStorage.GetCDRs(&utils.CDRsFilter{CgrIds: []string{utils.Sha1("bbb1", time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC).String()),
			utils.Sha1("bbb2", time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC).String())}, Count: true}); err != nil {
			t.Error(err.Error())
		} else if count != 2 {
			t.Error("Unexpected count of StoredCdrs returned: ", count)
		}
		// Filter on cgrids plus reqType
		if storedCdrs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{CgrIds: []string{utils.Sha1("bbb1", time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC).String()),
			utils.Sha1("bbb2", time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC).String())}, ReqTypes: []string{utils.META_PREPAID}}); err != nil {
			t.Error(err.Error())
		} else if len(storedCdrs) != 1 {
			t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
		}
		// Count on multiple filter
		if _, count, err := cdrStorage.GetCDRs(&utils.CDRsFilter{CgrIds: []string{utils.Sha1("bbb1", time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC).String()),
			utils.Sha1("bbb2", time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC).String())}, ReqTypes: []string{utils.META_PREPAID}, Count: true}); err != nil {
			t.Error(err.Error())
		} else if count != 1 {
			t.Error("Unexpected count of StoredCdrs returned: ", count)
		}
		// Filter on runId
		if storedCdrs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{RunIds: []string{utils.DEFAULT_RUNID}}); err != nil {
			t.Error(err.Error())
		} else if len(storedCdrs) != 2 {
			t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
		}
		// Filter on TOR
		if storedCdrs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{Tors: []string{utils.SMS}}); err != nil {
			t.Error(err.Error())
		} else if len(storedCdrs) != 0 {
			t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
		}
		// Filter on multiple TOR
		if storedCdrs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{Tors: []string{utils.SMS, utils.VOICE}}); err != nil {
			t.Error(err.Error())
		} else if len(storedCdrs) != 8 {
			t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
		}
		// Filter on cdrHost
		if storedCdrs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{CdrHosts: []string{"192.168.1.2"}}); err != nil {
			t.Error(err.Error())
		} else if len(storedCdrs) != 3 {
			t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
		}
		// Filter on multiple cdrHost
		if storedCdrs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{CdrHosts: []string{"192.168.1.1", "192.168.1.2"}}); err != nil {
			t.Error(err.Error())
		} else if len(storedCdrs) != 8 {
			t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
		}
		// Filter on cdrSource
		if storedCdrs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{CdrSources: []string{"UNKNOWN"}}); err != nil {
			t.Error(err.Error())
		} else if len(storedCdrs) != 1 {
			t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
		}
		// Filter on multiple cdrSource
		if storedCdrs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{CdrSources: []string{"UNKNOWN", "UNKNOWN2"}}); err != nil {
			t.Error(err.Error())
		} else if len(storedCdrs) != 2 {
			t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
		}
		// Filter on reqType
		if storedCdrs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{ReqTypes: []string{utils.META_PREPAID}}); err != nil {
			t.Error(err.Error())
		} else if len(storedCdrs) != 2 {
			t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
		}
		// Filter on multiple reqType
		if storedCdrs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{ReqTypes: []string{utils.META_PREPAID, utils.META_PSEUDOPREPAID}}); err != nil {
			t.Error(err.Error())
		} else if len(storedCdrs) != 3 {
			t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
		}
		// Filter on direction
		if storedCdrs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{Directions: []string{"*out"}}); err != nil {
			t.Error(err.Error())
		} else if len(storedCdrs) != 8 {
			t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
		}
		// Filter on tenant
		if storedCdrs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{Tenants: []string{"itsyscom.com"}}); err != nil {
			t.Error(err.Error())
		} else if len(storedCdrs) != 3 {
			t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
		}
		// Filter on multiple tenants
		if storedCdrs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{Tenants: []string{"itsyscom.com", "cgrates.org"}}); err != nil {
			t.Error(err.Error())
		} else if len(storedCdrs) != 8 {
			t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
		}
		// Filter on category
		if storedCdrs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{Categories: []string{"premium_call"}}); err != nil {
			t.Error(err.Error())
		} else if len(storedCdrs) != 1 {
			t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
		}
		// Filter on multiple categories
		if storedCdrs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{Categories: []string{"premium_call", "call"}}); err != nil {
			t.Error(err.Error())
		} else if len(storedCdrs) != 8 {
			t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
		}
		// Filter on account
		if storedCdrs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{Accounts: []string{"1002"}}); err != nil {
			t.Error(err.Error())
		} else if len(storedCdrs) != 3 {
			t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
		}
		// Filter on multiple account
		if storedCdrs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{Accounts: []string{"1001", "1002"}}); err != nil {
			t.Error(err.Error())
		} else if len(storedCdrs) != 8 {
			t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
		}
		// Filter on subject
		if storedCdrs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{Subjects: []string{"1000"}}); err != nil {
			t.Error(err.Error())
		} else if len(storedCdrs) != 1 {
			t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
		}
		// Filter on multiple subject
		if storedCdrs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{Subjects: []string{"1000", "1002"}}); err != nil {
			t.Error(err.Error())
		} else if len(storedCdrs) != 3 {
			t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
		}
		// Filter on destPrefix
		if storedCdrs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{DestPrefixes: []string{"+498651"}}); err != nil {
			t.Error(err.Error())
		} else if len(storedCdrs) != 3 {
			t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
		}
		// Filter on multiple destPrefixes
		if storedCdrs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{DestPrefixes: []string{"1001", "+498651"}}); err != nil {
			t.Error(err.Error())
		} else if len(storedCdrs) != 4 {
			t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
		}
		// Filter on ratedAccount
		if storedCdrs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{RatedAccounts: []string{"8001"}}); err != nil {
			t.Error(err.Error())
		} else if len(storedCdrs) != 1 {
			t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
		}
		// Filter on ratedSubject
		if storedCdrs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{RatedSubjects: []string{"91001"}}); err != nil {
			t.Error(err.Error())
		} else if len(storedCdrs) != 1 {
			t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
		}
		// Filter on ignoreRated
		var orderIdStart, orderIdEnd int64 // Capture also orderIds for the next test
		if storedCdrs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{MaxCost: utils.Float64Pointer(0.0)}); err != nil {
			t.Error(err.Error())
		} else if len(storedCdrs) != 5 {
			t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
		} else {
			for _, cdr := range storedCdrs {
				if cdr.OrderId < orderIdStart {
					orderIdStart = cdr.OrderId
				}
				if cdr.OrderId > orderIdEnd {
					orderIdEnd = cdr.OrderId
				}
			}
		}
		// Filter on orderIdStart
		if storedCdrs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{OrderIdStart: orderIdStart}); err != nil {
			t.Error(err.Error())
		} else if len(storedCdrs) != 8 {
			t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
		}
		// Filter on orderIdStart and orderIdEnd
		if storedCdrs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{OrderIdStart: orderIdStart, OrderIdEnd: orderIdEnd}); err != nil {
			t.Error(err.Error())
		} else if len(storedCdrs) != 4 {
			t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
		}
		var timeStart, timeEnd time.Time
		// Filter on timeStart
		timeStart = time.Date(2013, 11, 8, 8, 0, 0, 0, time.UTC)
		if storedCdrs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{AnswerTimeStart: &timeStart}); err != nil {
			t.Error(err.Error())
		} else if len(storedCdrs) != 5 {
			t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
		}
		// Filter on timeStart and timeEnd
		timeEnd = time.Date(2013, 12, 1, 8, 0, 0, 0, time.UTC)
		if storedCdrs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{AnswerTimeStart: &timeStart, AnswerTimeEnd: &timeEnd}); err != nil {
			t.Error(err.Error())
		} else if len(storedCdrs) != 2 {
			t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
		}
		// Filter on minPdd
		if storedCdrs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{MinPdd: utils.Float64Pointer(3)}); err != nil {
			t.Error(err.Error())
		} else if len(storedCdrs) != 7 {
			t.Error("Unexpected number of StoredCdrs returned: ", len(storedCdrs))
		}
		// Filter on maxPdd
		if storedCdrs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{MaxPdd: utils.Float64Pointer(3)}); err != nil {
			t.Error(err.Error())
		} else if len(storedCdrs) != 1 {
			t.Error("Unexpected number of StoredCdrs returned: ", len(storedCdrs))
		}
		// Filter on minPdd, maxPdd
		if storedCdrs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{MinPdd: utils.Float64Pointer(3), MaxPdd: utils.Float64Pointer(5)}); err != nil {
			t.Error(err.Error())
		} else if len(storedCdrs) != 5 {
			t.Error("Unexpected number of StoredCdrs returned: ", len(storedCdrs))
		}
		// Combined filter
		if storedCdrs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{ReqTypes: []string{utils.META_RATED}, AnswerTimeStart: &timeStart, AnswerTimeEnd: &timeEnd}); err != nil {
			t.Error(err.Error())
		} else if len(storedCdrs) != 1 {
			t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
		}
		// Filter on ignoreDerived
		if storedCdrs, _, err := cdrStorage.GetCDRs(&utils.CDRsFilter{AnswerTimeStart: &timeStart, AnswerTimeEnd: &timeEnd, FilterOnRated: true}); err != nil {
			t.Error(err.Error())
		} else if len(storedCdrs) != 0 { // ToDo: Recheck this value
			t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
		}
	*/
	return nil
}
