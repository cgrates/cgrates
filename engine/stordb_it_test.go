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
	"path"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

var (
	cfg             *config.CGRConfig
	storDB          StorDB
	storDB2ndDBname string
)

// subtests to be executed for each confDIR
var sTestsStorDBit = []func(t *testing.T){
	testStorDBitFlush,
	testStorDBitCRUDVersions,
	testStorDBitCRUDTpTimings,
	testStorDBitCRUDTpDestinations,
	testStorDBitCRUDTpRates,
	testStorDBitCRUDTpDestinationRates,
	testStorDBitCRUDTpRatingPlans,
	testStorDBitCRUDTpRatingProfiles,
	testStorDBitCRUDTpSharedGroups,
	testStorDBitCRUDTpActions,
	testStorDBitCRUDTpActionPlans,
	testStorDBitCRUDTpActionTriggers,
	testStorDBitCRUDTpAccountActions,
	testStorDBitCRUDTpLCRs,
	testStorDBitCRUDTpDerivedChargers,
	testStorDBitCRUDTpCdrStats,
	testStorDBitCRUDTpUsers,
	testStorDBitCRUDTpResourceLimits,
	testStorDBitCRUDCDRs,
	testStorDBitCRUDSMCosts,
}

func TestStorDBitMySQL(t *testing.T) {
	if cfg, err = config.NewCGRConfigFromFolder(path.Join(*dataDir, "conf", "samples", "storage", "mysql")); err != nil {
		t.Fatal(err)
	}
	if storDB, err = NewMySQLStorage(cfg.StorDBHost, cfg.StorDBPort, cfg.StorDBName,
		cfg.StorDBUser, cfg.StorDBPass, cfg.StorDBMaxOpenConns, cfg.StorDBMaxIdleConns); err != nil {
		t.Fatal(err)
	}
	storDB2ndDBname = "mysql"
	for _, stest := range sTestsStorDBit {
		stestFullName := runtime.FuncForPC(reflect.ValueOf(stest).Pointer()).Name()
		split := strings.Split(stestFullName, ".")
		stestName := split[len(split)-1]
		t.Run(stestName, stest)
	}
}

func TestStorDBitPostgresSQL(t *testing.T) {
	if cfg, err = config.NewCGRConfigFromFolder(path.Join(*dataDir, "conf", "samples", "storage", "postgres")); err != nil {
		t.Fatal(err)
	}
	if storDB, err = NewPostgresStorage(cfg.StorDBHost, cfg.StorDBPort, cfg.StorDBName,
		cfg.StorDBUser, cfg.StorDBPass, cfg.StorDBMaxOpenConns, cfg.StorDBMaxIdleConns); err != nil {
		t.Fatal(err)
	}
	storDB2ndDBname = "postgres"
	for _, stest := range sTestsStorDBit {
		stestFullName := runtime.FuncForPC(reflect.ValueOf(stest).Pointer()).Name()
		split := strings.Split(stestFullName, ".")
		stestName := split[len(split)-1]
		t.Run(stestName, stest)
	}
}

func TestStorDBitMongo(t *testing.T) {
	if cfg, err = config.NewCGRConfigFromFolder(path.Join(*dataDir, "conf", "samples", "storage", "mongo")); err != nil {
		t.Fatal(err)
	}
	if storDB, err = NewMongoStorage(cfg.StorDBHost, cfg.StorDBPort, cfg.StorDBName,
		cfg.StorDBUser, cfg.StorDBPass, utils.StorDB, cfg.StorDBCDRSIndexes, nil, cfg.LoadHistorySize); err != nil {
		t.Fatal(err)
	}
	storDB2ndDBname = "todo"
	for _, stest := range sTestsStorDBit {
		stestFullName := runtime.FuncForPC(reflect.ValueOf(stest).Pointer()).Name()
		split := strings.Split(stestFullName, ".")
		stestName := split[len(split)-1]
		// Fixme: Implement mongo needed versions methods
		if stestName != "testStorDBitCRUDVersions" {
			stestName := split[len(split)-1]
			t.Run(stestName, stest)
		}
	}
}

func testStorDBitCRUDTpTimings(t *testing.T) {
	// READ
	if _, err := storDB.GetTpTimings("testTPid", ""); err != utils.ErrNotFound {
		t.Error(err)
	}
	// WRITE
	var snd = []TpTiming{
		TpTiming{
			Tpid:      "testTPid",
			Tag:       "testTag1",
			Years:     "*any",
			Months:    "*any",
			MonthDays: "*any",
			WeekDays:  "1;2;3;4;5",
			Time:      "01:00:00",
		},
		TpTiming{
			Tpid:      "testTPid",
			Tag:       "testTag2",
			Years:     "*any",
			Months:    "*any",
			MonthDays: "*any",
			WeekDays:  "1;2;3;4;5",
			Time:      "01:00:00",
		},
	}
	if err := storDB.SetTpTimings(snd); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTpTimings("testTPid", ""); err != nil {
		t.Error(err)
	} else {
		rcv[0].Id = 0
		rcv[1].Id = 0
		rcv[0].CreatedAt = time.Time{}
		rcv[1].CreatedAt = time.Time{}
		if !(reflect.DeepEqual(snd[0], rcv[0]) || reflect.DeepEqual(snd[0], rcv[1])) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0], rcv[0], rcv[1])
		}
	}
	// UPDATE
	snd[0].Time = "02:00:00"
	snd[1].Time = "02:00:00"
	if err := storDB.SetTpTimings(snd); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTpTimings("testTPid", ""); err != nil {
		t.Error(err)
	} else {
		rcv[0].Id = 0
		rcv[1].Id = 0
		rcv[0].CreatedAt = time.Time{}
		rcv[1].CreatedAt = time.Time{}
		if !(reflect.DeepEqual(snd[0], rcv[0]) || reflect.DeepEqual(snd[0], rcv[1])) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0], rcv[0], rcv[1])
		}
	}
	// REMOVE
	if err := storDB.RemTpData("", "testTPid", nil); err != nil {
		t.Error(err)
	}
	// READ
	if _, err := storDB.GetTpTimings("testTPid", ""); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testStorDBitCRUDTpDestinations(t *testing.T) {
	// READ
	if _, err := storDB.GetTPDestinations("testTPid", ""); err != utils.ErrNotFound {
		t.Error(err)
	}
	// WRITE
	snd := []*utils.TPDestination{
		&utils.TPDestination{
			TPid:     "testTPid",
			Tag:      "testTag1",
			Prefixes: []string{`0256`, `0257`, `0723`, `+49`},
		},
		&utils.TPDestination{
			TPid:     "testTPid",
			Tag:      "testTag2",
			Prefixes: []string{`0256`, `0257`, `0723`, `+49`},
		},
	}
	if err := storDB.SetTPDestinations(snd); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTPDestinations("testTPid", ""); err != nil {
		t.Error(err)
	} else {
		prfs := make(map[string]bool)
		for _, prf := range snd[0].Prefixes {
			prfs[prf] = true
		}
		pfrOk := true
		for i, _ := range rcv[0].Prefixes {
			found1, _ := prfs[rcv[0].Prefixes[i]]
			found2, _ := prfs[rcv[1].Prefixes[i]]
			if !found1 && !found2 {
				pfrOk = false
			}
		}
		if pfrOk {
			rcv[0].Prefixes = snd[0].Prefixes
			rcv[1].Prefixes = snd[0].Prefixes
		}
		if !(reflect.DeepEqual(snd[0], rcv[0]) || reflect.DeepEqual(snd[0], rcv[1])) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0], rcv[0], rcv[1])
		}
	}
	// UPDATE
	snd[0].Prefixes = []string{`9999`, `0257`, `0723`, `+49`}
	snd[1].Prefixes = []string{`9999`, `0257`, `0723`, `+49`}
	if err := storDB.SetTPDestinations(snd); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTPDestinations("testTPid", ""); err != nil {
		t.Error(err)
	} else {
		prfs := make(map[string]bool)
		for _, prf := range snd[0].Prefixes {
			prfs[prf] = true
		}
		pfrOk := true
		for i, _ := range rcv[0].Prefixes {
			found1, _ := prfs[rcv[0].Prefixes[i]]
			found2, _ := prfs[rcv[1].Prefixes[i]]
			if !found1 && !found2 {
				pfrOk = false
			}
		}
		if pfrOk {
			rcv[0].Prefixes = snd[0].Prefixes
			rcv[1].Prefixes = snd[0].Prefixes
		}
		if !(reflect.DeepEqual(snd[0], rcv[0]) || reflect.DeepEqual(snd[0], rcv[1])) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0], rcv[0], rcv[1])
		}
	}
	// REMOVE
	if err := storDB.RemTpData("", "testTPid", nil); err != nil {
		t.Error(err)
	}
	// READ
	if _, err := storDB.GetTPDestinations("testTPid", ""); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testStorDBitCRUDTpRates(t *testing.T) {
	// READ
	if _, err := storDB.GetTpRates("testTPid", ""); err != utils.ErrNotFound {
		t.Error(err)
	}
	// WRITE
	var snd = []TpRate{
		TpRate{
			Tpid:               "testTPid",
			Tag:                "testTag1",
			ConnectFee:         0.0,
			Rate:               0.0,
			RateUnit:           "60s",
			RateIncrement:      "60s",
			GroupIntervalStart: "0s",
		},
		TpRate{
			Tpid:               "testTPid",
			Tag:                "testTag2",
			ConnectFee:         1.0,
			Rate:               1.0,
			RateUnit:           "70s",
			RateIncrement:      "70s",
			GroupIntervalStart: "0s",
		},
	}
	if err := storDB.SetTpRates(snd); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTpRates("testTPid", ""); err != nil {
		t.Error(err)
	} else {
		rcv[0].Id = 0
		rcv[1].Id = 0
		rcv[0].CreatedAt = time.Time{}
		rcv[1].CreatedAt = time.Time{}
		if !(reflect.DeepEqual(snd[0], rcv[0]) || reflect.DeepEqual(snd[0], rcv[1])) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0], rcv[0], rcv[1])
		}
	}
	// UPDATE
	snd[0].GroupIntervalStart = "1s"
	snd[1].GroupIntervalStart = "1s"
	if err := storDB.SetTpRates(snd); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTpRates("testTPid", ""); err != nil {
		t.Error(err)
	} else {
		rcv[0].Id = 0
		rcv[1].Id = 0
		rcv[0].CreatedAt = time.Time{}
		rcv[1].CreatedAt = time.Time{}
		if !(reflect.DeepEqual(snd[0], rcv[0]) || reflect.DeepEqual(snd[0], rcv[1])) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0], rcv[0], rcv[1])
		}
	}
	// REMOVE
	if err := storDB.RemTpData("", "testTPid", nil); err != nil {
		t.Error(err)
	}
	// READ
	if _, err := storDB.GetTpRates("testTPid", ""); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testStorDBitCRUDTpDestinationRates(t *testing.T) {
	// READ
	if _, err := storDB.GetTpDestinationRates("testTPid", "", nil); err != utils.ErrNotFound {
		t.Error(err)
	}
	// WRITE
	var snd = []TpDestinationRate{
		TpDestinationRate{
			Tpid:             "testTPid",
			Tag:              "testTag1",
			DestinationsTag:  "GERMANY",
			RatesTag:         "RT_1CENT",
			RoundingMethod:   "*up",
			RoundingDecimals: 0,
			MaxCost:          0.0,
			MaxCostStrategy:  "",
		},
		TpDestinationRate{
			Tpid:             "testTPid",
			Tag:              "testTag1",
			DestinationsTag:  "FRANCE",
			RatesTag:         "RT_2CENT",
			RoundingMethod:   "*down",
			RoundingDecimals: 0,
			MaxCost:          0.0,
			MaxCostStrategy:  "",
		},
	}
	if err := storDB.SetTpDestinationRates(snd); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTpDestinationRates("testTPid", "", nil); err != nil {
		t.Error(err)
	} else {
		rcv[0].Id = 0
		rcv[1].Id = 0
		rcv[0].CreatedAt = time.Time{}
		rcv[1].CreatedAt = time.Time{}
		if !(reflect.DeepEqual(snd[0], rcv[0]) || reflect.DeepEqual(snd[0], rcv[1])) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0], rcv[0], rcv[1])
		}
	}
	// UPDATE
	snd[0].MaxCostStrategy = "test"
	snd[1].MaxCostStrategy = "test"
	if err := storDB.SetTpDestinationRates(snd); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTpDestinationRates("testTPid", "", nil); err != nil {
		t.Error(err)
	} else {
		rcv[0].Id = 0
		rcv[1].Id = 0
		rcv[0].CreatedAt = time.Time{}
		rcv[1].CreatedAt = time.Time{}
		if !(reflect.DeepEqual(snd[0], rcv[0]) || reflect.DeepEqual(snd[0], rcv[1])) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0], rcv[0], rcv[1])
		}
	}
	// REMOVE
	if err := storDB.RemTpData("", "testTPid", nil); err != nil {
		t.Error(err)
	}
	// READ
	if _, err := storDB.GetTpDestinationRates("testTPid", "", nil); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testStorDBitCRUDTpRatingPlans(t *testing.T) {
	// READ
	if _, err := storDB.GetTpRatingPlans("testTPid", "", nil); err != utils.ErrNotFound {
		t.Error(err)
	}
	// WRITE
	var snd = []TpRatingPlan{
		TpRatingPlan{
			Tpid:         "testTPid",
			Tag:          "testTag1",
			DestratesTag: "1",
			TimingTag:    "ALWAYS",
			Weight:       0.0,
		},
		TpRatingPlan{
			Tpid:         "testTPid",
			Tag:          "testTag1",
			DestratesTag: "2",
			TimingTag:    "ALWAYS",
			Weight:       1.0,
		},
	}
	if err := storDB.SetTpRatingPlans(snd); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTpRatingPlans("testTPid", "", nil); err != nil {
		t.Error(err)
	} else {
		rcv[0].Id = 0
		rcv[1].Id = 0
		rcv[0].CreatedAt = time.Time{}
		rcv[1].CreatedAt = time.Time{}
		if !(reflect.DeepEqual(snd[0], rcv[0]) || reflect.DeepEqual(snd[0], rcv[1])) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0], rcv[0], rcv[1])
		}
	}
	// UPDATE
	snd[0].TimingTag = "test"
	snd[1].TimingTag = "test"
	if err := storDB.SetTpRatingPlans(snd); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTpRatingPlans("testTPid", "", nil); err != nil {
		t.Error(err)
	} else {
		rcv[0].Id = 0
		rcv[1].Id = 0
		rcv[0].CreatedAt = time.Time{}
		rcv[1].CreatedAt = time.Time{}
		if !(reflect.DeepEqual(snd[0], rcv[0]) || reflect.DeepEqual(snd[0], rcv[1])) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0], rcv[0], rcv[1])
		}
	}
	// REMOVE
	if err := storDB.RemTpData("", "testTPid", nil); err != nil {
		t.Error(err)
	}
	// READ
	if _, err := storDB.GetTpRatingPlans("testTPid", "", nil); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testStorDBitCRUDTpRatingProfiles(t *testing.T) {
	// READ
	var filter = TpRatingProfile{
		Tpid: "testTPid",
	}
	if _, err := storDB.GetTpRatingProfiles(&filter); err != utils.ErrNotFound {
		t.Error(err)
	}
	// WRITE
	var snd = []TpRatingProfile{
		TpRatingProfile{
			Tpid:             "testTPid",
			Loadid:           "TEST_LOADID",
			Direction:        "*out",
			Tenant:           "cgrates.org",
			Category:         "call",
			Subject:          "test",
			ActivationTime:   "2014-07-29T15:00:00Z",
			RatingPlanTag:    "test",
			FallbackSubjects: "",
			CdrStatQueueIds:  "",
		},
		TpRatingProfile{
			Tpid:             "testTPid",
			Loadid:           "TEST_LOADID2",
			Direction:        "*out",
			Tenant:           "cgrates.org",
			Category:         "call",
			Subject:          "test",
			ActivationTime:   "2014-07-29T15:00:00Z",
			RatingPlanTag:    "test",
			FallbackSubjects: "",
			CdrStatQueueIds:  "",
		},
	}
	if err := storDB.SetTpRatingProfiles(snd); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTpRatingProfiles(&filter); err != nil {
		t.Error(err)
	} else {
		rcv[0].Id = 0
		rcv[1].Id = 0
		rcv[0].CreatedAt = time.Time{}
		rcv[1].CreatedAt = time.Time{}
		if !(reflect.DeepEqual(snd[0], rcv[0]) || reflect.DeepEqual(snd[0], rcv[1])) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0], rcv[0], rcv[1])
		}
	}
	// UPDATE
	snd[0].CdrStatQueueIds = "test"
	snd[1].CdrStatQueueIds = "test"
	if err := storDB.SetTpRatingProfiles(snd); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTpRatingProfiles(&filter); err != nil {
		t.Error(err)
	} else {
		rcv[0].Id = 0
		rcv[1].Id = 0
		rcv[0].CreatedAt = time.Time{}
		rcv[1].CreatedAt = time.Time{}
		if !(reflect.DeepEqual(snd[0], rcv[0]) || reflect.DeepEqual(snd[0], rcv[1])) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0], rcv[0], rcv[1])
		}
	}
	// REMOVE
	if err := storDB.RemTpData("", "testTPid", nil); err != nil {
		t.Error(err)
	}
	// READ
	if _, err := storDB.GetTpRatingProfiles(&filter); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testStorDBitCRUDTpSharedGroups(t *testing.T) {
	// READ
	if _, err := storDB.GetTpSharedGroups("testTPid", ""); err != utils.ErrNotFound {
		t.Error(err)
	}
	// WRITE
	var snd = []TpSharedGroup{
		TpSharedGroup{
			Tpid:          "testTPid",
			Tag:           "testTag1",
			Account:       "test1",
			Strategy:      "*lowest_cost",
			RatingSubject: "test",
		},
		TpSharedGroup{
			Tpid:          "testTPid",
			Tag:           "testTag1",
			Account:       "test2",
			Strategy:      "*lowest_cost",
			RatingSubject: "test",
		},
	}
	if err := storDB.SetTpSharedGroups(snd); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTpSharedGroups("testTPid", ""); err != nil {
		t.Error(err)
	} else {
		rcv[0].Id = 0
		rcv[1].Id = 0
		rcv[0].CreatedAt = time.Time{}
		rcv[1].CreatedAt = time.Time{}
		if !(reflect.DeepEqual(snd[0], rcv[0]) || reflect.DeepEqual(snd[0], rcv[1])) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0], rcv[0], rcv[1])
		}
	}
	// UPDATE
	snd[0].Strategy = "test"
	snd[1].Strategy = "test"
	if err := storDB.SetTpSharedGroups(snd); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTpSharedGroups("testTPid", ""); err != nil {
		t.Error(err)
	} else {
		rcv[0].Id = 0
		rcv[1].Id = 0
		rcv[0].CreatedAt = time.Time{}
		rcv[1].CreatedAt = time.Time{}
		if !(reflect.DeepEqual(snd[0], rcv[0]) || reflect.DeepEqual(snd[0], rcv[1])) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0], rcv[0], rcv[1])
		}
	}
	// REMOVE
	if err := storDB.RemTpData("", "testTPid", nil); err != nil {
		t.Error(err)
	}
	// READ
	if _, err := storDB.GetTpSharedGroups("testTPid", ""); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testStorDBitCRUDTpActions(t *testing.T) {
	// READ
	if _, err := storDB.GetTpActions("testTPid", ""); err != utils.ErrNotFound {
		t.Error(err)
	}
	// WRITE
	var snd = []TpAction{
		TpAction{
			Tpid:            "testTPid",
			Tag:             "testTag1",
			Action:          "*topup_reset",
			ExtraParameters: "",
			Filter:          "",
			BalanceTag:      "",
			BalanceType:     "*monetary",
			Directions:      "*out",
			Categories:      "",
			DestinationTags: "DST_ON_NET",
			RatingSubject:   "",
			SharedGroups:    "",
			ExpiryTime:      "*unlimited",
			TimingTags:      "",
			Units:           "10",
			BalanceWeight:   "10",
			BalanceBlocker:  "false",
			BalanceDisabled: "false",
			Weight:          11.0,
		},
		TpAction{
			Tpid:            "testTPid",
			Tag:             "testTag1",
			Action:          "*topup_reset2",
			ExtraParameters: "",
			Filter:          "",
			BalanceTag:      "",
			BalanceType:     "*monetary",
			Directions:      "*out",
			Categories:      "",
			DestinationTags: "DST_ON_NET",
			RatingSubject:   "",
			SharedGroups:    "",
			ExpiryTime:      "*unlimited",
			TimingTags:      "",
			Units:           "10",
			BalanceWeight:   "10",
			BalanceBlocker:  "false",
			BalanceDisabled: "false",
			Weight:          11.0,
		},
	}
	if err := storDB.SetTpActions(snd); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTpActions("testTPid", ""); err != nil {
		t.Error(err)
	} else {
		rcv[0].Id = 0
		rcv[1].Id = 0
		rcv[0].CreatedAt = time.Time{}
		rcv[1].CreatedAt = time.Time{}
		if !(reflect.DeepEqual(snd[0], rcv[0]) || reflect.DeepEqual(snd[0], rcv[1])) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0], rcv[0], rcv[1])
		}
	}
	// UPDATE
	snd[0].Weight = 12.1
	snd[1].Weight = 12.1
	if err := storDB.SetTpActions(snd); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTpActions("testTPid", ""); err != nil {
		t.Error(err)
	} else {
		rcv[0].Id = 0
		rcv[1].Id = 0
		rcv[0].CreatedAt = time.Time{}
		rcv[1].CreatedAt = time.Time{}
		if !(reflect.DeepEqual(snd[0], rcv[0]) || reflect.DeepEqual(snd[0], rcv[1])) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0], rcv[0], rcv[1])
		}
	}
	// REMOVE
	if err := storDB.RemTpData("", "testTPid", nil); err != nil {
		t.Error(err)
	}
	// READ
	if _, err := storDB.GetTpActions("testTPid", ""); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testStorDBitCRUDTpActionPlans(t *testing.T) {
	// READ
	if _, err := storDB.GetTpActionPlans("testTPid", ""); err != utils.ErrNotFound {
		t.Error(err)
	}
	// WRITE
	var snd = []TpActionPlan{
		TpActionPlan{
			Tpid:       "testTPid",
			Tag:        "testTag1",
			ActionsTag: "test",
			TimingTag:  "",
			Weight:     0.0,
		},
		TpActionPlan{
			Tpid:       "testTPid",
			Tag:        "testTag2",
			ActionsTag: "test2",
			TimingTag:  "",
			Weight:     0.0,
		},
	}
	if err := storDB.SetTpActionPlans(snd); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTpActionPlans("testTPid", ""); err != nil {
		t.Error(err)
	} else {
		rcv[0].Id = 0
		rcv[1].Id = 0
		rcv[0].CreatedAt = time.Time{}
		rcv[1].CreatedAt = time.Time{}
		if !(reflect.DeepEqual(snd[0], rcv[0]) || reflect.DeepEqual(snd[0], rcv[1])) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0], rcv[0], rcv[1])
		}
	}
	// UPDATE
	snd[0].TimingTag = "test"
	snd[1].TimingTag = "test"
	if err := storDB.SetTpActionPlans(snd); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTpActionPlans("testTPid", ""); err != nil {
		t.Error(err)
	} else {
		rcv[0].Id = 0
		rcv[1].Id = 0
		rcv[0].CreatedAt = time.Time{}
		rcv[1].CreatedAt = time.Time{}
		if !(reflect.DeepEqual(snd[0], rcv[0]) || reflect.DeepEqual(snd[0], rcv[1])) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0], rcv[0], rcv[1])
		}
	}
	// REMOVE
	if err := storDB.RemTpData("", "testTPid", nil); err != nil {
		t.Error(err)
	}
	// READ
	if _, err := storDB.GetTpActionPlans("testTPid", ""); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testStorDBitCRUDTpActionTriggers(t *testing.T) {
	// READ
	if _, err := storDB.GetTpActionTriggers("testTPid", ""); err != utils.ErrNotFound {
		t.Error(err)
	}
	// WRITE
	var snd = []TpActionTrigger{
		TpActionTrigger{
			Tpid:                   "testTPid",
			Tag:                    "testTag1",
			UniqueId:               "",
			ThresholdType:          "",
			ThresholdValue:         0.0,
			Recurrent:              true,
			MinSleep:               "0",
			ExpiryTime:             "2014-07-29T15:00:00Z",
			ActivationTime:         "2014-07-29T15:00:00Z",
			BalanceTag:             "test",
			BalanceType:            "*monetary",
			BalanceDirections:      "*out",
			BalanceCategories:      "call",
			BalanceDestinationTags: "",
			BalanceRatingSubject:   "test",
			BalanceSharedGroups:    "SHARED_1",
			BalanceExpiryTime:      "2014-07-29T15:00:00Z",
			BalanceTimingTags:      "T1",
			BalanceWeight:          "0.0",
			BalanceBlocker:         "false",
			BalanceDisabled:        "false",
			MinQueuedItems:         0,
			ActionsTag:             "test",
			Weight:                 0.0,
		},
		TpActionTrigger{
			Tpid:                   "testTPid",
			Tag:                    "testTag1",
			UniqueId:               "",
			ThresholdType:          "",
			ThresholdValue:         0.0,
			Recurrent:              true,
			MinSleep:               "0",
			ExpiryTime:             "2014-07-29T15:00:00Z",
			ActivationTime:         "2014-07-29T15:00:00Z",
			BalanceTag:             "test2",
			BalanceType:            "*monetary",
			BalanceDirections:      "*out",
			BalanceCategories:      "call",
			BalanceDestinationTags: "",
			BalanceRatingSubject:   "test",
			BalanceSharedGroups:    "SHARED_1",
			BalanceExpiryTime:      "2014-07-29T15:00:00Z",
			BalanceTimingTags:      "T1",
			BalanceWeight:          "0.0",
			BalanceBlocker:         "false",
			BalanceDisabled:        "false",
			MinQueuedItems:         0,
			ActionsTag:             "test",
			Weight:                 0.0,
		},
	}
	if err := storDB.SetTpActionTriggers(snd); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTpActionTriggers("testTPid", ""); err != nil {
		t.Error(err)
	} else {
		rcv[0].Id = 0
		rcv[1].Id = 0
		rcv[0].CreatedAt = time.Time{}
		rcv[1].CreatedAt = time.Time{}
		if !(reflect.DeepEqual(snd[0], rcv[0]) || reflect.DeepEqual(snd[0], rcv[1])) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0], rcv[0], rcv[1])
		}
	}
	// UPDATE
	snd[0].MinQueuedItems = 2
	snd[1].MinQueuedItems = 2
	if err := storDB.SetTpActionTriggers(snd); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTpActionTriggers("testTPid", ""); err != nil {
		t.Error(err)
	} else {
		rcv[0].Id = 0
		rcv[1].Id = 0
		rcv[0].CreatedAt = time.Time{}
		rcv[1].CreatedAt = time.Time{}
		if !(reflect.DeepEqual(snd[0], rcv[0]) || reflect.DeepEqual(snd[0], rcv[1])) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0], rcv[0], rcv[1])
		}
	}
	// REMOVE
	if err := storDB.RemTpData("", "testTPid", nil); err != nil {
		t.Error(err)
	}
	// READ
	if _, err := storDB.GetTpActionTriggers("testTPid", ""); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testStorDBitCRUDTpAccountActions(t *testing.T) {
	// READ
	var filter = TpAccountAction{
		Tpid:              "testTPid",
		Loadid:            "",
		Tenant:            "",
		Account:           "",
		ActionPlanTag:     "",
		ActionTriggersTag: "",
		AllowNegative:     true,
		Disabled:          true,
	}
	if _, err := storDB.GetTpAccountActions(&filter); err != utils.ErrNotFound {
		t.Error(err)
	}
	// WRITE
	var snd = []TpAccountAction{
		TpAccountAction{
			Tpid:              "testTPid",
			Loadid:            "TEST_LOADID",
			Tenant:            "cgrates.org",
			Account:           "1001",
			ActionPlanTag:     "PACKAGE_10_SHARED_A_5",
			ActionTriggersTag: "STANDARD_TRIGGERS",
			AllowNegative:     true,
			Disabled:          true,
		},
		TpAccountAction{
			Tpid:              "testTPid",
			Loadid:            "TEST_LOADID",
			Tenant:            "cgrates.org",
			Account:           "1002",
			ActionPlanTag:     "PACKAGE_10_SHARED_A_5",
			ActionTriggersTag: "STANDARD_TRIGGERS",
			AllowNegative:     true,
			Disabled:          true,
		},
	}
	if err := storDB.SetTpAccountActions(snd); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTpAccountActions(&filter); err != nil {
		t.Error(err)
	} else {
		rcv[0].Id = 0
		rcv[1].Id = 0
		rcv[0].CreatedAt = time.Time{}
		rcv[1].CreatedAt = time.Time{}
		if !(reflect.DeepEqual(snd[0], rcv[0]) || reflect.DeepEqual(snd[0], rcv[1])) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0], rcv[0], rcv[1])
		}
	}
	// UPDATE
	snd[0].Disabled = false
	snd[1].Disabled = false
	if err := storDB.SetTpAccountActions(snd); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTpAccountActions(&filter); err != nil {
		t.Error(err)
	} else {
		rcv[0].Id = 0
		rcv[1].Id = 0
		rcv[0].CreatedAt = time.Time{}
		rcv[1].CreatedAt = time.Time{}
		if !(reflect.DeepEqual(snd[0], rcv[0]) || reflect.DeepEqual(snd[0], rcv[1])) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0], rcv[0], rcv[1])
		}
	}
	// REMOVE
	if err := storDB.RemTpData("", "testTPid", nil); err != nil {
		t.Error(err)
	}
	// READ
	if _, err := storDB.GetTpAccountActions(&filter); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testStorDBitCRUDTpLCRs(t *testing.T) {
	// READ
	var filter = TpLcrRule{
		Tpid:           "testTPid",
		Direction:      "",
		Tenant:         "",
		Category:       "",
		Account:        "",
		Subject:        "",
		DestinationTag: "",
		RpCategory:     "",
		Strategy:       "",
		StrategyParams: "",
		ActivationTime: "",
		Weight:         0.0,
	}
	if _, err := storDB.GetTpLCRs(&filter); err != utils.ErrNotFound {
		t.Error(err)
	}
	// WRITE
	var snd = []TpLcrRule{
		TpLcrRule{
			Tpid:           "testTPid",
			Direction:      "*in",
			Tenant:         "cgrates.org",
			Category:       "LCR_STANDARD",
			Account:        "1000",
			Subject:        "test",
			DestinationTag: "",
			RpCategory:     "LCR_STANDARD",
			Strategy:       "*lowest_cost",
			StrategyParams: "",
			ActivationTime: "2012-01-01T00:00:00Z",
			Weight:         0.0,
		},
		TpLcrRule{
			Tpid:           "testTPid",
			Direction:      "*out",
			Tenant:         "cgrates.org",
			Category:       "LCR_STANDARD",
			Account:        "1000",
			Subject:        "test2",
			DestinationTag: "",
			RpCategory:     "LCR_STANDARD",
			Strategy:       "*lowest_cost",
			StrategyParams: "",
			ActivationTime: "2012-01-01T00:00:00Z",
			Weight:         0.0,
		},
	}
	if err := storDB.SetTpLCRs(snd); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTpLCRs(&filter); err != nil {
		t.Error(err)
	} else {
		rcv[0].Id = 0
		rcv[1].Id = 0
		rcv[0].CreatedAt = time.Time{}
		rcv[1].CreatedAt = time.Time{}
		if !(reflect.DeepEqual(snd[0], rcv[0]) || reflect.DeepEqual(snd[0], rcv[1])) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0], rcv[0], rcv[1])
		}
	}
	// UPDATE
	snd[0].StrategyParams = "test"
	snd[1].StrategyParams = "test"
	if err := storDB.SetTpLCRs(snd); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTpLCRs(&filter); err != nil {
		t.Error(err)
	} else {
		rcv[0].Id = 0
		rcv[1].Id = 0
		rcv[0].CreatedAt = time.Time{}
		rcv[1].CreatedAt = time.Time{}
		if !(reflect.DeepEqual(snd[0], rcv[0]) || reflect.DeepEqual(snd[0], rcv[1])) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0], rcv[0], rcv[1])
		}
	}
	// REMOVE
	if err := storDB.RemTpData("", "testTPid", nil); err != nil {
		t.Error(err)
	}
	// READ
	if _, err := storDB.GetTpLCRs(&filter); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testStorDBitCRUDTpDerivedChargers(t *testing.T) {
	// READ
	var filter = TpDerivedCharger{
		Tpid: "testTPid",
	}
	if _, err := storDB.GetTpDerivedChargers(&filter); err != utils.ErrNotFound {
		t.Error(err)
	}
	// WRITE
	var snd = []TpDerivedCharger{
		TpDerivedCharger{
			Tpid:                 "testTPid",
			Loadid:               "TEST_LOADID",
			Direction:            "*out",
			Tenant:               "cgrates.org",
			Category:             "call",
			Account:              "1000",
			Subject:              "test",
			DestinationIds:       "",
			Runid:                "default",
			RunFilters:           "",
			ReqTypeField:         "test",
			DirectionField:       "test",
			TenantField:          "test",
			CategoryField:        "test",
			AccountField:         "test",
			SubjectField:         "test",
			DestinationField:     "^+49151708707",
			SetupTimeField:       "test",
			PddField:             "~pdd:s/sip:(.+)/$1/",
			AnswerTimeField:      "~answertime2:s/sip:(.+)/$1/",
			UsageField:           "test",
			SupplierField:        "~supplier2:s/(.+)/$1/",
			DisconnectCauseField: "test",
			RatedField:           "test",
			CostField:            "0",
		},
		TpDerivedCharger{
			Tpid:                 "testTPid",
			Loadid:               "TEST_LOADID",
			Direction:            "*out",
			Tenant:               "cgrates.org",
			Category:             "call",
			Account:              "1001",
			Subject:              "test",
			DestinationIds:       "",
			Runid:                "default",
			RunFilters:           "",
			ReqTypeField:         "test",
			DirectionField:       "test",
			TenantField:          "test",
			CategoryField:        "test",
			AccountField:         "test",
			SubjectField:         "test",
			DestinationField:     "^+49151708707",
			SetupTimeField:       "test",
			PddField:             "~pdd:s/sip:(.+)/$1/",
			AnswerTimeField:      "~answertime2:s/sip:(.+)/$1/",
			UsageField:           "test",
			SupplierField:        "~supplier2:s/(.+)/$1/",
			DisconnectCauseField: "test",
			RatedField:           "test",
			CostField:            "0",
		},
	}
	if err := storDB.SetTpDerivedChargers(snd); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTpDerivedChargers(&filter); err != nil {
		t.Error(err)
	} else {
		rcv[0].Id = 0
		rcv[1].Id = 0
		rcv[0].CreatedAt = time.Time{}
		rcv[1].CreatedAt = time.Time{}
		if !(reflect.DeepEqual(snd[0], rcv[0]) || reflect.DeepEqual(snd[0], rcv[1])) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0], rcv[0], rcv[1])
		}
	}
	// UPDATE
	snd[0].CostField = "test"
	snd[1].CostField = "test"
	if err := storDB.SetTpDerivedChargers(snd); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTpDerivedChargers(&filter); err != nil {
		t.Error(err)
	} else {
		rcv[0].Id = 0
		rcv[1].Id = 0
		rcv[0].CreatedAt = time.Time{}
		rcv[1].CreatedAt = time.Time{}
		if !(reflect.DeepEqual(snd[0], rcv[0]) || reflect.DeepEqual(snd[0], rcv[1])) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0], rcv[0], rcv[1])
		}
	}
	// REMOVE
	if err := storDB.RemTpData("", "testTPid", nil); err != nil {
		t.Error(err)
	}
	// READ
	if _, err := storDB.GetTpDerivedChargers(&filter); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testStorDBitCRUDTpCdrStats(t *testing.T) {
	// READ
	if _, err := storDB.GetTpCdrStats("testTPid", ""); err != utils.ErrNotFound {
		t.Error(err)
	}
	// WRITE
	var snd = []TpCdrstat{
		TpCdrstat{
			Tpid:             "testTPid",
			Tag:              "testTag",
			QueueLength:      0,
			TimeWindow:       "10m",
			SaveInterval:     "10s",
			Metrics:          "ACD",
			SetupInterval:    "",
			Tors:             "",
			CdrHosts:         "",
			CdrSources:       "",
			ReqTypes:         "",
			Directions:       "",
			Tenants:          "test",
			Categories:       "",
			Accounts:         "",
			Subjects:         "1001",
			DestinationIds:   "1003",
			PddInterval:      "",
			UsageInterval:    "",
			Suppliers:        "suppl2",
			DisconnectCauses: "",
			MediationRunids:  "*default",
			RatedAccounts:    "",
			RatedSubjects:    "",
			CostInterval:     "",
			ActionTriggers:   "CDRST1001_WARN",
		},
		TpCdrstat{
			Tpid:             "testTPid",
			Tag:              "testTag",
			QueueLength:      0,
			TimeWindow:       "10m",
			SaveInterval:     "10s",
			Metrics:          "ACD",
			SetupInterval:    "",
			Tors:             "",
			CdrHosts:         "",
			CdrSources:       "",
			ReqTypes:         "",
			Directions:       "",
			Tenants:          "test",
			Categories:       "",
			Accounts:         "",
			Subjects:         "1001",
			DestinationIds:   "1003",
			PddInterval:      "",
			UsageInterval:    "",
			Suppliers:        "suppl2",
			DisconnectCauses: "",
			MediationRunids:  "*default",
			RatedAccounts:    "",
			RatedSubjects:    "",
			CostInterval:     "",
			ActionTriggers:   "CDRST1001_WARN",
		},
	}
	if err := storDB.SetTpCdrStats(snd); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTpCdrStats("testTPid", ""); err != nil {
		t.Error(err)
	} else {
		rcv[0].Id = 0
		rcv[1].Id = 0
		rcv[0].CreatedAt = time.Time{}
		rcv[1].CreatedAt = time.Time{}
		if !(reflect.DeepEqual(snd[0], rcv[0]) || reflect.DeepEqual(snd[0], rcv[1])) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0], rcv[0], rcv[1])
		}
	}
	// UPDATE
	snd[0].Categories = "test"
	snd[1].Categories = "test"
	if err := storDB.SetTpCdrStats(snd); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTpCdrStats("testTPid", ""); err != nil {
		t.Error(err)
	} else {
		rcv[0].Id = 0
		rcv[1].Id = 0
		rcv[0].CreatedAt = time.Time{}
		rcv[1].CreatedAt = time.Time{}
		if !(reflect.DeepEqual(snd[0], rcv[0]) || reflect.DeepEqual(snd[0], rcv[1])) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0], rcv[0], rcv[1])
		}
	}
	// REMOVE
	if err := storDB.RemTpData("", "testTPid", nil); err != nil {
		t.Error(err)
	}
	// READ
	if _, err := storDB.GetTpCdrStats("testTPid", ""); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testStorDBitCRUDTpUsers(t *testing.T) {
	// READ
	var filter = TpUser{
		Tpid:           "testTPid",
		Tenant:         "",
		UserName:       "",
		Masked:         true,
		AttributeName:  "",
		AttributeValue: "",
		Weight:         0.0,
	}
	if _, err := storDB.GetTpUsers(&filter); err != utils.ErrNotFound {
		t.Error(err)
	}
	// WRITE
	var snd = []TpUser{
		TpUser{
			Tpid:           "testTPid",
			Tenant:         "cgrates.org",
			UserName:       "1001",
			Masked:         true,
			AttributeName:  "Account",
			AttributeValue: "1001",
			Weight:         0.0,
		},
		TpUser{
			Tpid:           "testTPid",
			Tenant:         "cgrates.org",
			UserName:       "1001",
			Masked:         true,
			AttributeName:  "Account",
			AttributeValue: "1002",
			Weight:         0.0,
		},
	}
	if err := storDB.SetTpUsers(snd); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTpUsers(&filter); err != nil {
		t.Error(err)
	} else {
		rcv[0].Id = 0
		rcv[1].Id = 0
		rcv[0].CreatedAt = time.Time{}
		rcv[1].CreatedAt = time.Time{}
		if !(reflect.DeepEqual(snd[0], rcv[0]) || reflect.DeepEqual(snd[0], rcv[1])) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0], rcv[0], rcv[1])
		}
	}
	// UPDATE
	snd[0].Masked = false
	snd[1].Masked = false
	if err := storDB.SetTpUsers(snd); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTpUsers(&filter); err != nil {
		t.Error(err)
	} else {
		rcv[0].Id = 0
		rcv[1].Id = 0
		rcv[0].CreatedAt = time.Time{}
		rcv[1].CreatedAt = time.Time{}
		if !(reflect.DeepEqual(snd[0], rcv[0]) || reflect.DeepEqual(snd[0], rcv[1])) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0], rcv[0], rcv[1])
		}
	}
	// REMOVE
	if err := storDB.RemTpData("", "testTPid", nil); err != nil {
		t.Error(err)
	}
	// READ
	if _, err := storDB.GetTpUsers(&filter); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testStorDBitCRUDTpAliases(t *testing.T) {
	// READ
	var filter = TpAlias{
		Tpid:          "testTPid",
		Direction:     "",
		Tenant:        "",
		Category:      "",
		Account:       "",
		Subject:       "",
		DestinationId: "",
		Context:       "",
		Target:        "",
		Original:      "",
		Alias:         "",
		Weight:        0.0,
	}
	if _, err := storDB.GetTpAliases(&filter); err != utils.ErrNotFound {
		t.Error(err)
	}
	// WRITE
	var snd = []TpAlias{
		TpAlias{
			Tpid:          "testTPid",
			Direction:     "*out",
			Tenant:        "cgrates.org",
			Category:      "call",
			Account:       "1006",
			Subject:       "1006",
			DestinationId: "*any",
			Context:       "*rating",
			Target:        "Subject",
			Original:      "1006",
			Alias:         "1001",
			Weight:        10.0,
		},
		TpAlias{
			Tpid:          "testTPid",
			Direction:     "*out",
			Tenant:        "cgrates.org",
			Category:      "call",
			Account:       "1006",
			Subject:       "1006",
			DestinationId: "*any",
			Context:       "*rating",
			Target:        "Subject",
			Original:      "1006",
			Alias:         "1001",
			Weight:        10.0,
		},
	}
	if err := storDB.SetTpAliases(snd); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTpAliases(&filter); err != nil {
		t.Error(err)
	} else {
		rcv[0].Id = 0
		rcv[1].Id = 0
		if !(reflect.DeepEqual(snd[0], rcv[0]) || reflect.DeepEqual(snd[0], rcv[1])) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0], rcv[0], rcv[1])
		}
	}
	// UPDATE
	snd[0].Target = "test"
	snd[1].Target = "test"
	if err := storDB.SetTpAliases(snd); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTpAliases(&filter); err != nil {
		t.Error(err)
	} else {
		rcv[0].Id = 0
		rcv[1].Id = 0
		if !(reflect.DeepEqual(snd[0], rcv[0]) || reflect.DeepEqual(snd[0], rcv[1])) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0], rcv[0], rcv[1])
		}
	}
	// REMOVE
	if err := storDB.RemTpData("", "testTPid", nil); err != nil {
		t.Error(err)
	}
	// READ
	if _, err := storDB.GetTpAliases(&filter); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testStorDBitCRUDTpResourceLimits(t *testing.T) {
	// READ
	if _, err := storDB.GetTpResourceLimits("testTPid", ""); err != utils.ErrNotFound {
		t.Error(err)
	}
	//WRITE
	var snd = TpResourceLimits{
		&TpResourceLimit{
			Tpid:              "testTPid",
			Tag:               "testTag1",
			FilterType:        "test",
			FilterFieldName:   "test",
			FilterFieldValues: "test",
			ActivationTime:    "test",
			Weight:            0.0,
			Limit:             "test",
			ActionTriggerIds:  "test",
		},
		&TpResourceLimit{
			Tpid:              "testTPid",
			Tag:               "testTag2",
			FilterType:        "test",
			FilterFieldName:   "test",
			FilterFieldValues: "test",
			ActivationTime:    "test",
			Weight:            0.0,
			Limit:             "test",
			ActionTriggerIds:  "test",
		},
	}
	if err := storDB.SetTpResourceLimits(snd); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTpResourceLimits("testTPid", ""); err != nil {
		t.Error(err)
	} else {
		snd[0].CreatedAt = time.Time{}
		snd[1].CreatedAt = time.Time{}
		rcv[0].CreatedAt = time.Time{}
		rcv[1].CreatedAt = time.Time{}
		if !(reflect.DeepEqual(snd[0], rcv[0]) || reflect.DeepEqual(snd[0], rcv[1])) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0], rcv[0], rcv[1])
		}
	}
	// UPDATE
	snd[0].Weight = 2.1
	snd[1].Weight = 2.1
	if err := storDB.SetTpResourceLimits(snd); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTpResourceLimits("testTPid", ""); err != nil {
		t.Error(err)
	} else {
		snd[0].CreatedAt = time.Time{}
		snd[1].CreatedAt = time.Time{}
		rcv[0].CreatedAt = time.Time{}
		rcv[1].CreatedAt = time.Time{}
		if !(reflect.DeepEqual(snd[0], rcv[0]) || reflect.DeepEqual(snd[0], rcv[1])) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0], rcv[0], rcv[1])
		}
	}
	// REMOVE
	if err := storDB.RemTpData("", "testTPid", nil); err != nil {
		t.Error(err)
	}
	// READ
	if _, err := storDB.GetTpResourceLimits("testTPid", ""); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testStorDBitCRUDCDRs(t *testing.T) {
	// READ
	var filter = utils.CDRsFilter{}
	if _, _, err := storDB.GetCDRs(&filter, false); err != utils.ErrNotFound {
		t.Error(err)
	}
	// WRITE
	var snd = []*CDR{
		&CDR{
			CGRID:       "88ed9c38005f07576a1e1af293063833b60edcc6",
			RunID:       "1",
			OrderID:     1,
			OriginHost:  "host1",
			OriginID:    "1",
			CostDetails: &CallCost{Timespans: TimeSpans{}},
			ExtraFields: map[string]string{"Service-Context-Id": "voice@huawei.com"},
		},
		&CDR{
			CGRID:       "88ed9c38005f07576a1e1af293063833b60edcc2",
			RunID:       "2",
			OrderID:     2,
			OriginHost:  "host2",
			OriginID:    "2",
			CostDetails: &CallCost{Timespans: TimeSpans{}},
			ExtraFields: map[string]string{"Service-Context-Id": "voice@huawei.com"},
		},
	}
	for _, cdr := range snd {
		if err := storDB.SetCDR(cdr, false); err != nil {
			t.Error(err)
		}
	}

	// READ
	if rcv, _, err := storDB.GetCDRs(&filter, false); err != nil {
		t.Error(err)
	} else {
		if !(reflect.DeepEqual(snd[0].CGRID, rcv[0].CGRID) || reflect.DeepEqual(snd[0].CGRID, rcv[1].CGRID)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].CGRID, rcv[0].CGRID, rcv[1].CGRID)
		}
		if !(reflect.DeepEqual(snd[0].RunID, rcv[0].RunID) || reflect.DeepEqual(snd[0].RunID, rcv[1].RunID)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].RunID, rcv[0].RunID, rcv[1].RunID)
		}
		if !(reflect.DeepEqual(snd[0].OrderID, rcv[0].OrderID) || reflect.DeepEqual(snd[0].OrderID, rcv[1].OrderID)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].OrderID, rcv[0].OrderID, rcv[1].OrderID)
		}
		if !(reflect.DeepEqual(snd[0].OriginHost, rcv[0].OriginHost) || reflect.DeepEqual(snd[0].OriginHost, rcv[1].OriginHost)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].OriginHost, rcv[0].OriginHost, rcv[1].OriginHost)
		}
		if !(reflect.DeepEqual(snd[0].Source, rcv[0].Source) || reflect.DeepEqual(snd[0].Source, rcv[1].Source)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].Source, rcv[0].Source, rcv[1].Source)
		}
		if !(reflect.DeepEqual(snd[0].OriginID, rcv[0].OriginID) || reflect.DeepEqual(snd[0].OriginID, rcv[1].OriginID)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].OriginID, rcv[0].OriginID, rcv[1].OriginID)
		}
		if !(reflect.DeepEqual(snd[0].ToR, rcv[0].ToR) || reflect.DeepEqual(snd[0].ToR, rcv[1].ToR)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].ToR, rcv[0].ToR, rcv[1].ToR)
		}
		if !(reflect.DeepEqual(snd[0].RequestType, rcv[0].RequestType) || reflect.DeepEqual(snd[0].RequestType, rcv[1].RequestType)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].RequestType, rcv[0].RequestType, rcv[1].RequestType)
		}
		if !(reflect.DeepEqual(snd[0].Direction, rcv[0].Direction) || reflect.DeepEqual(snd[0].Direction, rcv[1].Direction)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].Direction, rcv[0].Direction, rcv[1].Direction)
		}
		if !(reflect.DeepEqual(snd[0].Tenant, rcv[0].Tenant) || reflect.DeepEqual(snd[0].Tenant, rcv[1].Tenant)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].Tenant, rcv[0].Tenant, rcv[1].Tenant)
		}
		if !(reflect.DeepEqual(snd[0].Category, rcv[0].Category) || reflect.DeepEqual(snd[0].Category, rcv[1].Category)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].Category, rcv[0].Category, rcv[1].Category)
		}
		if !(reflect.DeepEqual(snd[0].Account, rcv[0].Account) || reflect.DeepEqual(snd[0].Account, rcv[1].Account)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].Account, rcv[0].Account, rcv[1].Account)
		}
		if !(reflect.DeepEqual(snd[0].Subject, rcv[0].Subject) || reflect.DeepEqual(snd[0].Subject, rcv[1].Subject)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].Subject, rcv[0].Subject, rcv[1].Subject)
		}
		if !(reflect.DeepEqual(snd[0].Destination, rcv[0].Destination) || reflect.DeepEqual(snd[0].Destination, rcv[1].Destination)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].Destination, rcv[0].Destination, rcv[1].Destination)
		}
		if !(snd[0].SetupTime.Equal(rcv[0].SetupTime) || snd[0].SetupTime.Equal(rcv[1].SetupTime)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].SetupTime, rcv[0].SetupTime, rcv[1].SetupTime)
		}
		if !(reflect.DeepEqual(snd[0].PDD, rcv[0].PDD) || reflect.DeepEqual(snd[0].PDD, rcv[1].PDD)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].PDD, rcv[0].PDD, rcv[1].PDD)
		}
		if !(snd[0].AnswerTime.Equal(rcv[0].AnswerTime) || snd[0].AnswerTime.Equal(rcv[1].AnswerTime)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].AnswerTime, rcv[0].AnswerTime, rcv[1].AnswerTime)
		}
		if !(reflect.DeepEqual(snd[0].Usage, rcv[0].Usage) || reflect.DeepEqual(snd[0].Usage, rcv[1].Usage)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].Usage, rcv[0].Usage, rcv[1].Usage)
		}
		if !(reflect.DeepEqual(snd[0].Supplier, rcv[0].Supplier) || reflect.DeepEqual(snd[0].Supplier, rcv[1].Supplier)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].Supplier, rcv[0].Supplier, rcv[1].Supplier)
		}
		if !(reflect.DeepEqual(snd[0].DisconnectCause, rcv[0].DisconnectCause) || reflect.DeepEqual(snd[0].DisconnectCause, rcv[1].DisconnectCause)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].DisconnectCause, rcv[0].DisconnectCause, rcv[1].DisconnectCause)
		}
		if !(reflect.DeepEqual(snd[0].ExtraFields, rcv[0].ExtraFields) || reflect.DeepEqual(snd[0].ExtraFields, rcv[1].ExtraFields)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].ExtraFields, rcv[0].ExtraFields, rcv[1].ExtraFields)
		}
		if !(reflect.DeepEqual(snd[0].CostSource, rcv[0].CostSource) || reflect.DeepEqual(snd[0].CostSource, rcv[1].CostSource)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].CostSource, rcv[0].CostSource, rcv[1].CostSource)
		}
		if !(reflect.DeepEqual(snd[0].Cost, rcv[0].Cost) || reflect.DeepEqual(snd[0].Cost, rcv[1].Cost)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].Cost, rcv[0].Cost, rcv[1].Cost)
		}
		if !(reflect.DeepEqual(snd[0].CostDetails, rcv[0].CostDetails) || reflect.DeepEqual(snd[0].CostDetails, rcv[1].CostDetails)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].CostDetails, rcv[0].CostDetails, rcv[1].CostDetails)
		}
		if !(reflect.DeepEqual(snd[0].AccountSummary, rcv[0].AccountSummary) || reflect.DeepEqual(snd[0].AccountSummary, rcv[1].AccountSummary)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].AccountSummary, rcv[0].AccountSummary, rcv[1].AccountSummary)
		}
		if !(reflect.DeepEqual(snd[0].ExtraInfo, rcv[0].ExtraInfo) || reflect.DeepEqual(snd[0].ExtraInfo, rcv[1].ExtraInfo)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].ExtraInfo, rcv[0].ExtraInfo, rcv[1].ExtraInfo)
		}
		if !(reflect.DeepEqual(snd[0].Rated, rcv[0].Rated) || reflect.DeepEqual(snd[0].Rated, rcv[1].Rated)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].Rated, rcv[0].Rated, rcv[1].Rated)
		}
		if !(reflect.DeepEqual(snd[0].Partial, rcv[0].Partial) || reflect.DeepEqual(snd[0].Partial, rcv[1].Partial)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].Partial, rcv[0].Partial, rcv[1].Partial)
		}
	}
	// UPDATE
	snd[0].OriginHost = "host3"
	snd[1].OriginHost = "host3"
	for _, cdr := range snd {
		if err := storDB.SetCDR(cdr, true); err != nil {
			t.Error(err)
		}
	}
	// READ
	if rcv, _, err := storDB.GetCDRs(&filter, false); err != nil {
		t.Error(err)
	} else {
		if !(reflect.DeepEqual(snd[0].OriginHost, rcv[0].OriginHost) || reflect.DeepEqual(snd[0].OriginHost, rcv[1].OriginHost)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].OriginHost, rcv[0].OriginHost, rcv[1].OriginHost)
		}
	}
	// REMOVE
	if _, _, err := storDB.GetCDRs(&filter, true); err != nil {
		t.Error(err)
	}
	// READ
	if _, _, err := storDB.GetCDRs(&filter, false); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testStorDBitCRUDSMCosts(t *testing.T) {
	// READ
	if _, err := storDB.GetSMCosts("", "", "", ""); err != utils.ErrNotFound {
		t.Error(err)
	}
	// WRITE
	var snd = []*SMCost{
		&SMCost{
			CGRID:       "88ed9c38005f07576a1e1af293063833b60edcc6",
			RunID:       "1",
			OriginHost:  "host2",
			OriginID:    "2",
			CostDetails: &CallCost{Timespans: TimeSpans{}},
		},
		&SMCost{
			CGRID:       "88ed9c38005f07576a1e1af293063833b60edcc2",
			RunID:       "2",
			OriginHost:  "host2",
			OriginID:    "2",
			CostDetails: &CallCost{Timespans: TimeSpans{}},
		},
	}
	for _, smc := range snd {
		if err := storDB.SetSMCost(smc); err != nil {
			t.Error(err)
		}
	}
	// READ
	if rcv, err := storDB.GetSMCosts("", "", "host2", ""); err != nil {
		t.Error(err)
	} else {
		if !(reflect.DeepEqual(snd[0].CGRID, rcv[0].CGRID) || reflect.DeepEqual(snd[0].CGRID, rcv[1].CGRID)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].CGRID, rcv[0].CGRID, rcv[1].CGRID)
		}
		if !(reflect.DeepEqual(snd[0].RunID, rcv[0].RunID) || reflect.DeepEqual(snd[0].RunID, rcv[1].RunID)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].RunID, rcv[0].RunID, rcv[1].RunID)
		}
		if !(reflect.DeepEqual(snd[0].OriginHost, rcv[0].OriginHost) || reflect.DeepEqual(snd[0].OriginHost, rcv[1].OriginHost)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].OriginHost, rcv[0].OriginHost, rcv[1].OriginHost)
		}
		if !(reflect.DeepEqual(snd[0].OriginID, rcv[0].OriginID) || reflect.DeepEqual(snd[0].OriginID, rcv[1].OriginID)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].OriginID, rcv[0].OriginID, rcv[1].OriginID)
		}
		if !(reflect.DeepEqual(snd[0].CostDetails, rcv[0].CostDetails) || reflect.DeepEqual(snd[0].CostDetails, rcv[1].CostDetails)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].CostDetails, rcv[0].CostDetails, rcv[1].CostDetails)
		}
	}
	// REMOVE
	for _, smc := range snd {
		if err := storDB.RemoveSMCost(smc); err != nil {
			t.Error(err)
		}
	}
	// READ
	if _, err := storDB.GetSMCosts("", "", "", ""); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testStorDBitFlush(t *testing.T) {
	if err := storDB.Flush(path.Join(cfg.DataFolderPath, "storage", cfg.StorDBType)); err != nil {
		t.Error(err)
	}
}

func testStorDBitCRUDVersions(t *testing.T) {
	// CREATE
	vrs := Versions{utils.COST_DETAILS: 1}
	if err := storDB.SetVersions(vrs, true); err != nil {
		t.Error(err)
	}
	if rcv, err := storDB.GetVersions(""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(vrs, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", vrs, rcv)
	}

	// UPDATE
	vrs = Versions{utils.COST_DETAILS: 2, "OTHER_KEY": 1}
	if err := storDB.SetVersions(vrs, false); err != nil {
		t.Error(err)
	}
	if rcv, err := storDB.GetVersions(""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(vrs, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", vrs, rcv)
	}

	// REMOVE
	vrs = Versions{"OTHER_KEY": 1}
	if err := storDB.RemoveVersions(vrs); err != nil {
		t.Error(err)
	}
	if rcv, err := storDB.GetVersions(utils.COST_DETAILS); err != nil {
		t.Error(err)
	} else if len(rcv) != 1 || rcv[utils.COST_DETAILS] != 2 {
		t.Errorf("Received: %+v", rcv)
	}
	if _, err := storDB.GetVersions("UNKNOWN"); err != nil {
		t.Error(err)
	}
	vrs = Versions{"UNKNOWN": 1}
	if err := storDB.RemoveVersions(vrs); err != nil {
		t.Error(err)
	}
	if err := storDB.RemoveVersions(nil); err != nil {
		t.Error(err)
	}
	if rcv, err := storDB.GetVersions(""); err != nil {
		t.Error(err)
	} else if len(rcv) != 0 {
		t.Errorf("Received: %+v", rcv)
	}

}
