//go:build integration
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
	"sort"
	"strings"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

var (
	cfg    *config.CGRConfig
	storDB StorDB
)

// subtests to be executed for each confDIR
var sTestsStorDBit = []func(t *testing.T){
	testStorDBitFlush,
	testStorDBitIsDBEmpty,
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
	testStorDBitCRUDTpResources,
	testStorDBitCRUDTpStats,
	testStorDBitCRUDCDRs,
	testStorDBitCRUDSMCosts,
	testStorDBitCRUDSMCosts2,
}

func TestStorDBit(t *testing.T) {
	//var stestName string
	switch *dbType {
	case utils.MetaInternal:
		if cfg, err = config.NewDefaultCGRConfig(); err != nil {
			t.Error(err)
		}
		config.SetCgrConfig(cfg)
		storDB = NewInternalDB(nil, nil, false, cfg.StorDbCfg().Items)
	case utils.MetaMySQL:
		if cfg, err = config.NewCGRConfigFromPath(path.Join(*dataDir, "conf", "samples", "storage", "mysql")); err != nil {
			t.Fatal(err)
		}
		if storDB, err = NewMySQLStorage(cfg.StorDbCfg().Host,
			cfg.StorDbCfg().Port, cfg.StorDbCfg().Name,
			cfg.StorDbCfg().User, cfg.StorDbCfg().Password,
			cfg.StorDbCfg().MaxOpenConns, cfg.StorDbCfg().MaxIdleConns,
			cfg.StorDbCfg().ConnMaxLifetime); err != nil {
			t.Fatal(err)
		}
	case utils.MetaMongo:
		if cfg, err = config.NewCGRConfigFromPath(path.Join(*dataDir, "conf", "samples", "storage", "mongo")); err != nil {
			t.Fatal(err)
		}
		if storDB, err = NewMongoStorage(cfg.StorDbCfg().Host,
			cfg.StorDbCfg().Port, cfg.StorDbCfg().Name,
			cfg.StorDbCfg().User, cfg.StorDbCfg().Password,
			cfg.GeneralCfg().DBDataEncoding,
			utils.StorDB, cfg.StorDbCfg().StringIndexedFields, false); err != nil {
			t.Fatal(err)
		}
	case utils.MetaPostgres:
		if cfg, err = config.NewCGRConfigFromPath(path.Join(*dataDir, "conf", "samples", "storage", "postgres")); err != nil {
			t.Fatal(err)
		}
		if storDB, err = NewPostgresStorage(cfg.StorDbCfg().Host,
			cfg.StorDbCfg().Port, cfg.StorDbCfg().Name,
			cfg.StorDbCfg().User, cfg.StorDbCfg().Password,
			cfg.StorDbCfg().SSLMode, cfg.StorDbCfg().MaxOpenConns,
			cfg.StorDbCfg().MaxIdleConns, cfg.StorDbCfg().ConnMaxLifetime); err != nil {
			t.Fatal(err)
		}
	default:
		t.Fatal("Unknown Database type")
	}

	for _, stest := range sTestsStorDBit {
		stestFullName := runtime.FuncForPC(reflect.ValueOf(stest).Pointer()).Name()
		split := strings.Split(stestFullName, ".")
		stestName := split[len(split)-1]
		// Fixme: Implement mongo needed versions methods
		if (*dbType == utils.MetaMongo || *dbType == utils.MetaInternal) && stestName != "testStorDBitCRUDVersions" {
			stestName := split[len(split)-1]
			t.Run(stestName, stest)
		}
	}
}

func testStorDBitIsDBEmpty(t *testing.T) {
	x := storDB.GetStorageType()
	switch x {
	case utils.MONGO:
		test, err := storDB.IsDBEmpty()
		if err != nil {
			t.Error(err)
		} else if test != true {
			t.Errorf("\nExpecting: true got :%+v", test)
		}
	case utils.POSTGRES, utils.MYSQL:
		test, err := storDB.IsDBEmpty()
		if err != nil {
			t.Error(err)
		} else if test != false {
			t.Errorf("\nExpecting: false got :%+v", test)
		}
	}
}

func testStorDBitCRUDTpTimings(t *testing.T) {
	// READ
	if _, err := storDB.GetTPTimings("testTPid", ""); err != utils.ErrNotFound {
		t.Error(err)
	}
	// WRITE
	var snd = []*utils.ApierTPTiming{
		{
			TPid:      "testTPid",
			ID:        "testTag1",
			Years:     "*any",
			Months:    "*any",
			MonthDays: "*any",
			WeekDays:  "1;2;3;4;5",
			Time:      "01:00:00",
		},
		{
			TPid:      "testTPid",
			ID:        "testTag2",
			Years:     "*any",
			Months:    "*any",
			MonthDays: "*any",
			WeekDays:  "1;2;3;4;5",
			Time:      "01:00:00",
		},
	}
	if err := storDB.SetTPTimings(snd); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTPTimings("testTPid", ""); err != nil {
		t.Error(err)
	} else {
		if !(reflect.DeepEqual(snd[0], rcv[0]) || reflect.DeepEqual(snd[0], rcv[1])) {
			t.Errorf("\nExpecting:\n%+v\nReceived:\n%+v\n||\n%+v", utils.ToIJSON(snd[0]), utils.ToIJSON(rcv[0]), utils.ToIJSON(rcv[1]))
		}
	}
	// UPDATE
	snd[0].Time = "02:00:00"
	snd[1].Time = "02:00:00"
	if err := storDB.SetTPTimings(snd); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTPTimings("testTPid", ""); err != nil {
		t.Error(err)
	} else {
		if !(reflect.DeepEqual(snd[0], rcv[0]) || reflect.DeepEqual(snd[0], rcv[1])) {
			t.Errorf("\nExpecting:\n%+v\nReceived:\n%+v\n||\n%+v", utils.ToIJSON(snd[0]), utils.ToIJSON(rcv[0]), utils.ToIJSON(rcv[1]))
		}
	}
	// REMOVE
	if err := storDB.RemTpData("", "testTPid", nil); err != nil {
		t.Error(err)
	}
	// READ
	if _, err := storDB.GetTPTimings("testTPid", ""); err != utils.ErrNotFound {
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
		{
			TPid:     "testTPid",
			ID:       "testTag1",
			Prefixes: []string{`0256`, `0257`, `0723`, `+49`},
		},
		{
			TPid:     "testTPid",
			ID:       "testTag2",
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
		for i := range rcv[0].Prefixes {
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
			t.Errorf("\nExpecting:\n%+v\nReceived:\n%+v\n||\n%+v", utils.ToIJSON(snd[0]), utils.ToIJSON(rcv[0]), utils.ToIJSON(rcv[1]))
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
		for i := range rcv[0].Prefixes {
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
			t.Errorf("\nExpecting:\n%+v\nReceived:\n%+v\n||\n%+v", utils.ToIJSON(snd[0]), utils.ToIJSON(rcv[0]), utils.ToIJSON(rcv[1]))
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
	if _, err := storDB.GetTPRates("testTPid", ""); err != utils.ErrNotFound {
		t.Error(err)
	}
	// WRITE
	var snd = []*utils.TPRate{
		{
			TPid: "testTPid",
			ID:   "1",
			RateSlots: []*utils.RateSlot{
				{
					ConnectFee:         0.0,
					Rate:               0.0,
					RateUnit:           "60s",
					RateIncrement:      "60s",
					GroupIntervalStart: "0s",
				},
				{
					ConnectFee:         0.0,
					Rate:               0.0,
					RateUnit:           "60s",
					RateIncrement:      "60s",
					GroupIntervalStart: "1s",
				},
			},
		},
		{
			TPid: "testTPid",
			ID:   "2",
			RateSlots: []*utils.RateSlot{
				{
					ConnectFee:         0.0,
					Rate:               0.0,
					RateUnit:           "60s",
					RateIncrement:      "60s",
					GroupIntervalStart: "0s",
				},
			},
		},
	}
	snd[0].RateSlots[0].SetDurations()
	snd[0].RateSlots[1].SetDurations()
	snd[1].RateSlots[0].SetDurations()
	if err := storDB.SetTPRates(snd); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTPRates("testTPid", ""); err != nil {
		t.Error(err)
	} else {
		if !(reflect.DeepEqual(snd[0], rcv[0]) || reflect.DeepEqual(snd[0], rcv[1])) {
			t.Errorf("\nExpecting:\n%+v\nReceived:\n%+v\n||\n%+v", utils.ToIJSON(snd[0]), utils.ToIJSON(rcv[0]), utils.ToIJSON(rcv[1]))
		}
	}
	// UPDATE
	snd[0].RateSlots[0].GroupIntervalStart = "3s"
	snd[1].RateSlots[0].GroupIntervalStart = "3s"
	snd[0].RateSlots[0].SetDurations()
	snd[1].RateSlots[0].SetDurations()
	if err := storDB.SetTPRates(snd); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTPRates("testTPid", ""); err != nil {
		t.Error(err)
	} else {
		if !(reflect.DeepEqual(snd[0], rcv[0]) || reflect.DeepEqual(snd[0], rcv[1])) {
			t.Errorf("\nExpecting:\n%+v\nReceived:\n%+v\n||\n%+v", utils.ToIJSON(snd[0]), utils.ToIJSON(rcv[0]), utils.ToIJSON(rcv[1]))
		}
	}
	// REMOVE
	if err := storDB.RemTpData("", "testTPid", nil); err != nil {
		t.Error(err)
	}
	// READ
	if _, err := storDB.GetTPRates("testTPid", ""); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testStorDBitCRUDTpDestinationRates(t *testing.T) {
	// READ
	if _, err := storDB.GetTPDestinationRates("testTPid", "", nil); err != utils.ErrNotFound {
		t.Error(err)
	}
	// WRITE
	var snd = []*utils.TPDestinationRate{
		{
			TPid: "testTPid",
			ID:   "1",
			DestinationRates: []*utils.DestinationRate{
				{
					DestinationId:    "GERMANY",
					RateId:           "RT_1CENT",
					RoundingMethod:   "*up",
					RoundingDecimals: 0,
					MaxCost:          0.0,
					MaxCostStrategy:  "",
				},
			},
		},
		{
			TPid: "testTPid",
			ID:   "2",
			DestinationRates: []*utils.DestinationRate{
				{
					DestinationId:    "GERMANY",
					RateId:           "RT_1CENT",
					RoundingMethod:   "*up",
					RoundingDecimals: 0,
					MaxCost:          0.0,
					MaxCostStrategy:  "",
				},
			},
		},
	}

	if err := storDB.SetTPDestinationRates(snd); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTPDestinationRates("testTPid", "", nil); err != nil {
		t.Error(err)
	} else {
		if !(reflect.DeepEqual(snd[0], rcv[0]) || reflect.DeepEqual(snd[0], rcv[1])) {
			t.Errorf("\nExpecting:\n%+v\nReceived:\n%+v\n||\n%+v", utils.ToIJSON(snd[0]), utils.ToIJSON(rcv[0]), utils.ToIJSON(rcv[1]))
		}
	}
	// UPDATE
	snd[0].DestinationRates[0].MaxCostStrategy = "test"
	snd[1].DestinationRates[0].MaxCostStrategy = "test"
	if err := storDB.SetTPDestinationRates(snd); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTPDestinationRates("testTPid", "", nil); err != nil {
		t.Error(err)
	} else {
		if !(reflect.DeepEqual(snd[0], rcv[0]) || reflect.DeepEqual(snd[0], rcv[1])) {
			t.Errorf("\nExpecting:\n%+v\nReceived:\n%+v\n||\n%+v", utils.ToIJSON(snd[0]), utils.ToIJSON(rcv[0]), utils.ToIJSON(rcv[1]))
		}
	}
	// REMOVE
	if err := storDB.RemTpData("", "testTPid", nil); err != nil {
		t.Error(err)
	}
	// READ
	if _, err := storDB.GetTPDestinationRates("testTPid", "", nil); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testStorDBitCRUDTpRatingPlans(t *testing.T) {
	// READ
	if _, err := storDB.GetTPRatingPlans("testTPid", "", nil); err != utils.ErrNotFound {
		t.Error(err)
	}
	// WRITE
	var snd = []*utils.TPRatingPlan{
		{
			TPid: "testTPid",
			ID:   "1",
			RatingPlanBindings: []*utils.TPRatingPlanBinding{
				{
					DestinationRatesId: "1",
					TimingId:           "ALWAYS",
					Weight:             0.0,
				},
			},
		},
		{
			TPid: "testTPid",
			ID:   "2",
			RatingPlanBindings: []*utils.TPRatingPlanBinding{
				{
					DestinationRatesId: "2",
					TimingId:           "ALWAYS",
					Weight:             2,
				},
			},
		},
	}
	if err := storDB.SetTPRatingPlans(snd); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTPRatingPlans("testTPid", "", nil); err != nil {
		t.Error(err)
	} else {
		if !(reflect.DeepEqual(snd[0], rcv[0]) || reflect.DeepEqual(snd[0], rcv[1])) {
			t.Errorf("\nExpecting:\n%+v\nReceived:\n%+v\n||\n%+v", utils.ToIJSON(snd[0]), utils.ToIJSON(rcv[0]), utils.ToIJSON(rcv[1]))
		}
	}
	// UPDATE
	snd[0].RatingPlanBindings[0].TimingId = "test"
	snd[1].RatingPlanBindings[0].TimingId = "test"
	if err := storDB.SetTPRatingPlans(snd); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTPRatingPlans("testTPid", "", nil); err != nil {
		t.Error(err)
	} else {
		if !(reflect.DeepEqual(snd[0], rcv[0]) || reflect.DeepEqual(snd[0], rcv[1])) {
			t.Errorf("\nExpecting:\n%+v\nReceived:\n%+v\n||\n%+v", utils.ToIJSON(snd[0]), utils.ToIJSON(rcv[0]), utils.ToIJSON(rcv[1]))
		}
	}
	// REMOVE
	if err := storDB.RemTpData("", "testTPid", nil); err != nil {
		t.Error(err)
	}
	// READ
	if _, err := storDB.GetTPRatingPlans("testTPid", "", nil); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testStorDBitCRUDTpRatingProfiles(t *testing.T) {
	// READ
	var filter = utils.TPRatingProfile{
		TPid: "testTPid",
	}
	if _, err := storDB.GetTPRatingProfiles(&filter); err != utils.ErrNotFound {
		t.Error(err)
	}
	// WRITE
	var snd = []*utils.TPRatingProfile{
		{
			TPid:     "testTPid",
			LoadId:   "TEST_LOADID",
			Tenant:   "cgrates.org",
			Category: "call",
			Subject:  "test",
			RatingPlanActivations: []*utils.TPRatingActivation{
				{
					ActivationTime:   "2014-07-29T15:00:00Z",
					RatingPlanId:     "test",
					FallbackSubjects: "",
				},
			},
		},
		{
			TPid:     "testTPid",
			LoadId:   "TEST_LOADID2",
			Tenant:   "cgrates.org",
			Category: "call",
			Subject:  "test",
			RatingPlanActivations: []*utils.TPRatingActivation{
				{
					ActivationTime:   "2014-07-29T15:00:00Z",
					RatingPlanId:     "test",
					FallbackSubjects: "",
				},
			},
		},
	}
	if err := storDB.SetTPRatingProfiles(snd); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTPRatingProfiles(&filter); err != nil {
		t.Error(err)
	} else {
		if !(reflect.DeepEqual(snd[0], rcv[0]) ||
			reflect.DeepEqual(snd[0], rcv[1])) {
			t.Errorf("\nExpecting:\n%+v\nReceived:\n%+v\n||\n%+v",
				utils.ToIJSON(snd[0]), utils.ToIJSON(rcv[0]), utils.ToIJSON(rcv[1]))
		}
	}
	// UPDATE
	snd[0].RatingPlanActivations = append(snd[0].RatingPlanActivations,
		&utils.TPRatingActivation{
			ActivationTime:   "2019-02-11T15:00:00Z",
			RatingPlanId:     "test",
			FallbackSubjects: "",
		})
	snd[1].RatingPlanActivations = append(snd[1].RatingPlanActivations,
		&utils.TPRatingActivation{
			ActivationTime:   "2019-02-11T15:00:00Z",
			RatingPlanId:     "test",
			FallbackSubjects: "",
		})
	if err := storDB.SetTPRatingProfiles(snd); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTPRatingProfiles(&filter); err != nil {
		t.Error(err)
	} else {
		if len(snd) != len(rcv) ||
			len(snd[0].RatingPlanActivations) != len(rcv[0].RatingPlanActivations) {
			t.Errorf("\nExpecting:\n%+v\nReceived:\n%+v\n||\n%+v",
				utils.ToIJSON(snd[0]), utils.ToIJSON(rcv[0]), utils.ToIJSON(rcv[1]))
		}
	}
	// REMOVE
	if err := storDB.RemTpData("", "testTPid", nil); err != nil {
		t.Error(err)
	}
	// READ
	if _, err := storDB.GetTPRatingProfiles(&filter); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testStorDBitCRUDTpSharedGroups(t *testing.T) {
	// READ
	if _, err := storDB.GetTPSharedGroups("testTPid", ""); err != utils.ErrNotFound {
		t.Error(err)
	}
	// WRITE
	var snd = []*utils.TPSharedGroups{
		{
			TPid: "testTPid",
			ID:   "1",
			SharedGroups: []*utils.TPSharedGroup{
				{
					Account:       "test",
					Strategy:      "*lowest_cost",
					RatingSubject: "test",
				},
			},
		},
		{
			TPid: "testTPid",
			ID:   "2",
			SharedGroups: []*utils.TPSharedGroup{
				{
					Account:       "test",
					Strategy:      "*lowest_cost",
					RatingSubject: "test",
				},
			},
		},
	}
	if err := storDB.SetTPSharedGroups(snd); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTPSharedGroups("testTPid", ""); err != nil {
		t.Error(err)
	} else {
		if !(reflect.DeepEqual(snd[0], rcv[0]) || reflect.DeepEqual(snd[0], rcv[1])) {
			t.Errorf("\nExpecting:\n%+v\nReceived:\n%+v\n||\n%+v", utils.ToIJSON(snd[0]), utils.ToIJSON(rcv[0]), utils.ToIJSON(rcv[1]))
		}
	}
	// UPDATE
	snd[0].SharedGroups[0].Strategy = "test"
	snd[1].SharedGroups[0].Strategy = "test"
	if err := storDB.SetTPSharedGroups(snd); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTPSharedGroups("testTPid", ""); err != nil {
		t.Error(err)
	} else {
		if !(reflect.DeepEqual(snd[0], rcv[0]) || reflect.DeepEqual(snd[0], rcv[1])) {
			t.Errorf("\nExpecting:\n%+v\nReceived:\n%+v\n||\n%+v", utils.ToIJSON(snd[0]), utils.ToIJSON(rcv[0]), utils.ToIJSON(rcv[1]))
		}
	}
	// REMOVE
	if err := storDB.RemTpData("", "testTPid", nil); err != nil {
		t.Error(err)
	}
	// READ
	if _, err := storDB.GetTPSharedGroups("testTPid", ""); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testStorDBitCRUDTpActions(t *testing.T) {
	// READ
	if _, err := storDB.GetTPActions("testTPid", ""); err != utils.ErrNotFound {
		t.Error(err)
	}
	// WRITE
	var snd = []*utils.TPActions{
		{
			TPid: "testTPid",
			ID:   "1",
			Actions: []*utils.TPAction{
				{
					Identifier:      "",
					BalanceId:       "",
					BalanceUuid:     "",
					BalanceType:     "*monetary",
					Units:           "10",
					ExpiryTime:      "*unlimited",
					Filter:          "",
					TimingTags:      "",
					DestinationIds:  "DST_ON_NET",
					RatingSubject:   "",
					Categories:      "",
					SharedGroups:    "",
					BalanceWeight:   "",
					ExtraParameters: "",
					BalanceBlocker:  "false",
					BalanceDisabled: "false",
					Weight:          11.0,
				},
			},
		},
		{
			TPid: "testTPid",
			ID:   "2",
			Actions: []*utils.TPAction{
				{
					Identifier:      "",
					BalanceId:       "",
					BalanceUuid:     "",
					BalanceType:     "*monetary",
					Units:           "10",
					ExpiryTime:      "*unlimited",
					Filter:          "",
					TimingTags:      "",
					DestinationIds:  "DST_ON_NET",
					RatingSubject:   "",
					Categories:      "",
					SharedGroups:    "",
					BalanceWeight:   "",
					ExtraParameters: "",
					BalanceBlocker:  "false",
					BalanceDisabled: "false",
					Weight:          11.0,
				},
			},
		},
	}
	if err := storDB.SetTPActions(snd); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTPActions("testTPid", ""); err != nil {
		t.Error(err)
	} else {
		if !(reflect.DeepEqual(snd[0], rcv[0]) || reflect.DeepEqual(snd[0], rcv[1])) {
			t.Errorf("\nExpecting:\n%+v\nReceived:\n%+v\n||\n%+v", utils.ToIJSON(snd[0]), utils.ToIJSON(rcv[0]), utils.ToIJSON(rcv[1]))
		}
	}
	// UPDATE
	snd[0].Actions[0].Weight = 12.1
	snd[1].Actions[0].Weight = 12.1
	if err := storDB.SetTPActions(snd); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTPActions("testTPid", ""); err != nil {
		t.Error(err)
	} else {
		if !(reflect.DeepEqual(snd[0], rcv[0]) || reflect.DeepEqual(snd[0], rcv[1])) {
			t.Errorf("\nExpecting:\n%+v\nReceived:\n%+v\n||\n%+v", utils.ToIJSON(snd[0]), utils.ToIJSON(rcv[0]), utils.ToIJSON(rcv[1]))
		}
	}
	// REMOVE
	if err := storDB.RemTpData("", "testTPid", nil); err != nil {
		t.Error(err)
	}
	// READ
	if _, err := storDB.GetTPActions("testTPid", ""); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testStorDBitCRUDTpActionPlans(t *testing.T) {
	// READ
	if _, err := storDB.GetTPActionPlans("testTPid", ""); err != utils.ErrNotFound {
		t.Error(err)
	}
	// WRITE
	var snd = []*utils.TPActionPlan{
		{
			TPid: "testTPid",
			ID:   "1",
			ActionPlan: []*utils.TPActionTiming{
				{
					ActionsId: "1",
					TimingId:  "1",
					Weight:    1,
				},
			},
		},
		{
			TPid: "testTPid",
			ID:   "2",
			ActionPlan: []*utils.TPActionTiming{
				{
					ActionsId: "1",
					TimingId:  "1",
					Weight:    1,
				},
			},
		},
	}
	if err := storDB.SetTPActionPlans(snd); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTPActionPlans("testTPid", ""); err != nil {
		t.Error(err)
	} else {
		if !(reflect.DeepEqual(snd[0], rcv[0]) || reflect.DeepEqual(snd[0], rcv[1])) {
			t.Errorf("\nExpecting:\n%+v\nReceived:\n%+v\n||\n%+v", utils.ToIJSON(snd[0]), utils.ToIJSON(rcv[0]), utils.ToIJSON(rcv[1]))
		}
	}
	// UPDATE
	snd[0].ActionPlan[0].TimingId = "test"
	snd[1].ActionPlan[0].TimingId = "test"
	if err := storDB.SetTPActionPlans(snd); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTPActionPlans("testTPid", ""); err != nil {
		t.Error(err)
	} else {
		if !(reflect.DeepEqual(snd[0], rcv[0]) || reflect.DeepEqual(snd[0], rcv[1])) {
			t.Errorf("\nExpecting:\n%+v\nReceived:\n%+v\n||\n%+v", utils.ToIJSON(snd[0]), utils.ToIJSON(rcv[0]), utils.ToIJSON(rcv[1]))
		}
	}
	// REMOVE
	if err := storDB.RemTpData("", "testTPid", nil); err != nil {
		t.Error(err)
	}
	// READ
	if _, err := storDB.GetTPActionPlans("testTPid", ""); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testStorDBitCRUDTpActionTriggers(t *testing.T) {
	// READ
	if _, err := storDB.GetTPActionTriggers("testTPid", ""); err != utils.ErrNotFound {
		t.Error(err)
	}
	// WRITE
	var snd = []*utils.TPActionTriggers{
		{
			TPid: "testTPid",
			ID:   "1",
			ActionTriggers: []*utils.TPActionTrigger{
				{
					Id:                    "1",
					UniqueID:              "",
					ThresholdType:         "1",
					ThresholdValue:        0,
					Recurrent:             true,
					MinSleep:              "",
					ExpirationDate:        "2014-07-29T15:00:00Z",
					ActivationDate:        "2014-07-29T15:00:00Z",
					BalanceId:             "test",
					BalanceType:           "*monetary",
					BalanceDestinationIds: "call",
					BalanceWeight:         "0.0",
					BalanceExpirationDate: "2014-07-29T15:00:00Z",
					BalanceTimingTags:     "T1",
					BalanceRatingSubject:  "test",
					BalanceCategories:     "",
					BalanceSharedGroups:   "SHARED_1",
					BalanceBlocker:        "false",
					BalanceDisabled:       "false",
					ActionsId:             "test",
					Weight:                1.0,
				},
			},
		},
		{
			TPid: "testTPid",
			ID:   "2",
			ActionTriggers: []*utils.TPActionTrigger{
				{
					Id:                    "2",
					UniqueID:              "",
					ThresholdType:         "1",
					ThresholdValue:        0,
					Recurrent:             true,
					MinSleep:              "",
					ExpirationDate:        "2014-07-29T15:00:00Z",
					ActivationDate:        "2014-07-29T15:00:00Z",
					BalanceId:             "test",
					BalanceType:           "*monetary",
					BalanceDestinationIds: "call",
					BalanceWeight:         "0.0",
					BalanceExpirationDate: "2014-07-29T15:00:00Z",
					BalanceTimingTags:     "T1",
					BalanceRatingSubject:  "test",
					BalanceCategories:     "",
					BalanceSharedGroups:   "SHARED_1",
					BalanceBlocker:        "false",
					BalanceDisabled:       "false",
					ActionsId:             "test",
					Weight:                1.0,
				},
			},
		},
	}
	if err := storDB.SetTPActionTriggers(snd); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTPActionTriggers("testTPid", ""); err != nil {
		t.Error(err)
	} else {
		if !(reflect.DeepEqual(snd[0], rcv[0]) || reflect.DeepEqual(snd[0], rcv[1])) {
			t.Errorf("\nExpecting:\n%+v\nReceived:\n%+v\n||\n%+v", utils.ToIJSON(snd[0]), utils.ToIJSON(rcv[0]), utils.ToIJSON(rcv[1]))
		}
	}
	// UPDATE
	snd[0].ActionTriggers[0].ActionsId = "test2"
	snd[1].ActionTriggers[0].ActionsId = "test2"
	if err := storDB.SetTPActionTriggers(snd); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTPActionTriggers("testTPid", ""); err != nil {
		t.Error(err)
	} else {
		if !(reflect.DeepEqual(snd[0], rcv[0]) || reflect.DeepEqual(snd[0], rcv[1])) {
			t.Errorf("\nExpecting:\n%+v\nReceived:\n%+v\n||\n%+v", utils.ToIJSON(snd[0]), utils.ToIJSON(rcv[0]), utils.ToIJSON(rcv[1]))
		}
	}
	// REMOVE
	if err := storDB.RemTpData("", "testTPid", nil); err != nil {
		t.Error(err)
	}
	// READ
	if _, err := storDB.GetTPActionTriggers("testTPid", ""); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testStorDBitCRUDTpAccountActions(t *testing.T) {
	// READ
	var filter = utils.TPAccountActions{
		TPid: "testTPid",
	}
	if _, err := storDB.GetTPAccountActions(&filter); err != utils.ErrNotFound {
		t.Error(err)
	}
	// WRITE
	var snd = []*utils.TPAccountActions{
		{
			TPid:             "testTPid",
			LoadId:           "TEST_LOADID",
			Tenant:           "cgrates.org",
			Account:          "1001",
			ActionPlanId:     "PACKAGE_10_SHARED_A_5",
			ActionTriggersId: "STANDARD_TRIGGERS",
			AllowNegative:    true,
			Disabled:         true,
		},
		{
			TPid:             "testTPid",
			LoadId:           "TEST_LOADID",
			Tenant:           "cgrates.org",
			Account:          "1002",
			ActionPlanId:     "PACKAGE_10_SHARED_A_5",
			ActionTriggersId: "STANDARD_TRIGGERS",
			AllowNegative:    true,
			Disabled:         true,
		},
	}
	if err := storDB.SetTPAccountActions(snd); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTPAccountActions(&filter); err != nil {
		t.Error(err)
	} else {
		if !(reflect.DeepEqual(snd[0], rcv[0]) || reflect.DeepEqual(snd[0], rcv[1])) {
			t.Errorf("\nExpecting:\n%+v\nReceived:\n%+v\n||\n%+v", utils.ToIJSON(snd[0]), utils.ToIJSON(rcv[0]), utils.ToIJSON(rcv[1]))
		}
	}
	// UPDATE
	snd[0].Disabled = false
	snd[1].Disabled = false
	if err := storDB.SetTPAccountActions(snd); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTPAccountActions(&filter); err != nil {
		t.Error(err)
	} else {
		if !(reflect.DeepEqual(snd[0], rcv[0]) || reflect.DeepEqual(snd[0], rcv[1])) {
			t.Errorf("\nExpecting:\n%+v\nReceived:\n%+v\n||\n%+v", utils.ToIJSON(snd[0]), utils.ToIJSON(rcv[0]), utils.ToIJSON(rcv[1]))
		}
	}
	// REMOVE
	if err := storDB.RemTpData("", "testTPid", nil); err != nil {
		t.Error(err)
	}
	// READ
	if _, err := storDB.GetTPAccountActions(&filter); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testStorDBitCRUDTpResources(t *testing.T) {
	// READ
	if _, err := storDB.GetTPResources("testTPid", "", ""); err != utils.ErrNotFound {
		t.Error(err)
	}
	//WRITE
	var snd = []*utils.TPResourceProfile{
		{
			TPid:         "testTPid",
			ID:           "testTag1",
			Weight:       0.0,
			Limit:        "test",
			ThresholdIDs: []string{"1x", "2x"},
			FilterIDs:    []string{"FILTR_RES_1"},
			Blocker:      true,
			Stored:       true,
		},
		{
			TPid:               "testTPid",
			ID:                 "testTag2",
			ActivationInterval: &utils.TPActivationInterval{ActivationTime: "test"},
			Weight:             0.0,
			Limit:              "test",
			ThresholdIDs:       []string{"1x", "2x"},
			FilterIDs:          []string{"FLTR_RES_2"},
			Blocker:            true,
			Stored:             false,
		},
	}
	if err := storDB.SetTPResources(snd); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTPResources("testTPid", "", ""); err != nil {
		t.Error(err)
	} else {
		if !(reflect.DeepEqual(snd[0].TPid, rcv[0].TPid) || reflect.DeepEqual(snd[0].TPid, rcv[1].TPid)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].TPid, rcv[0].TPid, rcv[1].TPid)
		}
		if !(reflect.DeepEqual(snd[0].ID, rcv[0].ID) || reflect.DeepEqual(snd[0].ID, rcv[1].ID)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].ID, rcv[0].ID, rcv[1].ID)
		}
		if !(reflect.DeepEqual(snd[0].ActivationInterval, rcv[0].ActivationInterval) || reflect.DeepEqual(snd[0].ActivationInterval, rcv[1].ActivationInterval)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].TPid, rcv[0].TPid, rcv[1].TPid)
		}
		if !(reflect.DeepEqual(snd[0].Weight, rcv[0].Weight) || reflect.DeepEqual(snd[0].Weight, rcv[1].Weight)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].Weight, rcv[0].Weight, rcv[1].Weight)
		}
		if !(reflect.DeepEqual(snd[0].Limit, rcv[0].Limit) || reflect.DeepEqual(snd[0].Limit, rcv[1].Limit)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].Limit, rcv[0].Limit, rcv[1].Limit)
		}
		sort.Strings(rcv[0].ThresholdIDs)
		sort.Strings(rcv[1].ThresholdIDs)
		if !(reflect.DeepEqual(snd[0].ThresholdIDs, rcv[0].ThresholdIDs) || reflect.DeepEqual(snd[0].ThresholdIDs, rcv[1].ThresholdIDs)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].ThresholdIDs, rcv[0].ThresholdIDs, rcv[1].ThresholdIDs)
		}
	}
	// UPDATE
	snd[0].Weight = 2.1
	snd[1].Weight = 2.1
	if err := storDB.SetTPResources(snd); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTPResources("testTPid", "", ""); err != nil {
		t.Error(err)
	} else {
		if !(reflect.DeepEqual(snd[0].TPid, rcv[0].TPid) || reflect.DeepEqual(snd[0].TPid, rcv[1].TPid)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].TPid, rcv[0].TPid, rcv[1].TPid)
		}
		if !(reflect.DeepEqual(snd[0].ID, rcv[0].ID) || reflect.DeepEqual(snd[0].ID, rcv[1].ID)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].ID, rcv[0].ID, rcv[1].ID)
		}
		if !(reflect.DeepEqual(snd[0].ActivationInterval, rcv[0].ActivationInterval) ||
			reflect.DeepEqual(snd[0].ActivationInterval, rcv[1].ActivationInterval)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].TPid, rcv[0].TPid, rcv[1].TPid)
		}
		if !(reflect.DeepEqual(snd[0].Weight, rcv[0].Weight) || reflect.DeepEqual(snd[0].Weight, rcv[1].Weight)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].Weight, rcv[0].Weight, rcv[1].Weight)
		}
		if !(reflect.DeepEqual(snd[0].Limit, rcv[0].Limit) || reflect.DeepEqual(snd[0].Limit, rcv[1].Limit)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].Limit, rcv[0].Limit, rcv[1].Limit)
		}
		sort.Strings(rcv[0].ThresholdIDs)
		sort.Strings(rcv[1].ThresholdIDs)
		if !(reflect.DeepEqual(snd[0].ThresholdIDs, rcv[0].ThresholdIDs) || reflect.DeepEqual(snd[0].ThresholdIDs, rcv[1].ThresholdIDs)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].ThresholdIDs, rcv[0].ThresholdIDs, rcv[1].ThresholdIDs)
		}
	}
	// REMOVE
	if err := storDB.RemTpData("", "testTPid", nil); err != nil {
		t.Error(err)
	}
	// READ
	if _, err := storDB.GetTPResources("testTPid", "", ""); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testStorDBitCRUDTpStats(t *testing.T) {
	// READ
	if _, err := storDB.GetTPStats("TEST_TPID", "", ""); err != utils.ErrNotFound {
		t.Error(err)
	}
	//WRITE
	eTPs := []*utils.TPStatProfile{
		{
			TPid:      "TEST_TPID",
			Tenant:    "Test",
			ID:        "Stats1",
			FilterIDs: []string{"FLTR_1"},
			ActivationInterval: &utils.TPActivationInterval{
				ActivationTime: "2014-07-29T15:00:00Z",
			},
			QueueLength: 100,
			TTL:         "1s",
			Metrics: []*utils.MetricWithFilters{
				&utils.MetricWithFilters{
					MetricID: "*asr",
				},
			},
			ThresholdIDs: []string{"*none"},
			Weight:       20.0,
			Stored:       true,
			MinItems:     1,
		},
	}

	if err := storDB.SetTPStats(eTPs); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTPStats("TEST_TPID", "", ""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eTPs[0], rcv[0]) {
		t.Errorf("Expecting: %+v,\n received: %+v", utils.ToJSON(eTPs[0]), utils.ToJSON(rcv[0]))
	}

	// UPDATE
	eTPs[0].Metrics = []*utils.MetricWithFilters{
		&utils.MetricWithFilters{
			MetricID: "*asr",
		},
		&utils.MetricWithFilters{
			MetricID: utils.MetaACD,
		},
	}
	if err := storDB.SetTPStats(eTPs); err != nil {
		t.Error(err)
	}
	eTPsReverse := []*utils.TPStatProfile{
		{
			TPid:      "TEST_TPID",
			Tenant:    "Test",
			ID:        "Stats1",
			FilterIDs: []string{"FLTR_1"},
			ActivationInterval: &utils.TPActivationInterval{
				ActivationTime: "2014-07-29T15:00:00Z",
			},
			QueueLength: 100,
			TTL:         "1s",
			Metrics: []*utils.MetricWithFilters{
				&utils.MetricWithFilters{
					MetricID: utils.MetaACD,
				},
				&utils.MetricWithFilters{
					MetricID: "*asr",
				},
			},
			ThresholdIDs: []string{"*none"},
			Weight:       20.0,
			Stored:       true,
			MinItems:     1,
		},
	}
	// READ
	if rcv, err := storDB.GetTPStats("TEST_TPID", "", ""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eTPs[0], rcv[0]) && !reflect.DeepEqual(eTPsReverse[0], rcv[0]) {
		t.Errorf("Expecting: %+v,\n received: %+v || reveived : %+v", utils.ToJSON(eTPs[0]), utils.ToJSON(rcv[0]), utils.ToJSON(eTPsReverse[0]))
	}

	// REMOVE
	if err := storDB.RemTpData(utils.TBLTPStats, "TEST_TPID", nil); err != nil {
		t.Error(err)
	}
	// READ
	if _, err := storDB.GetTPStats("TEST_TPID", "", ""); err != utils.ErrNotFound {
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
		{
			CGRID:       "88ed9c38005f07576a1e1af293063833b60edcc6",
			RunID:       "1",
			OrderID:     0,
			OriginHost:  "host1",
			OriginID:    "1",
			Usage:       1000000000,
			CostDetails: NewBareEventCost(),
			ExtraFields: map[string]string{"Service-Context-Id": "voice@huawei.com"},
		},
		{
			CGRID:       "88ed9c38005f07576a1e1af293063833b60edcc2",
			RunID:       "2",
			OrderID:     0,
			OriginHost:  "host2",
			OriginID:    "2",
			Usage:       1000000000,
			CostDetails: NewBareEventCost(),
			ExtraFields: map[string]string{"Service-Context-Id": "voice@huawei.com"},
		},
	}
	for _, cdr := range snd {
		if err := storDB.SetCDR(cdr, false); err != nil {
			t.Error(err)
		}
	}
	for _, cdr := range snd {
		if err := storDB.SetCDR(cdr, false); err == nil || err != utils.ErrExists {
			t.Error(err) // for mongo will fail because of indexes
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
		// if !(reflect.DeepEqual(snd[0].OrderID, rcv[0].OrderID) || reflect.DeepEqual(snd[0].OrderID, rcv[1].OrderID)) {
		// 	t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].OrderID, rcv[0].OrderID, rcv[1].OrderID)
		// }
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
		if !(snd[0].AnswerTime.Equal(rcv[0].AnswerTime) || snd[0].AnswerTime.Equal(rcv[1].AnswerTime)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].AnswerTime, rcv[0].AnswerTime, rcv[1].AnswerTime)
		}
		if !(reflect.DeepEqual(snd[0].Usage, rcv[0].Usage) || reflect.DeepEqual(snd[0].Usage, rcv[1].Usage)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].Usage, rcv[0].Usage, rcv[1].Usage)
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
		if !(reflect.DeepEqual(snd[0].ExtraInfo, rcv[0].ExtraInfo) || reflect.DeepEqual(snd[0].ExtraInfo, rcv[1].ExtraInfo)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].ExtraInfo, rcv[0].ExtraInfo, rcv[1].ExtraInfo)
		}
		if !(reflect.DeepEqual(snd[0].PreRated, rcv[0].PreRated) || reflect.DeepEqual(snd[0].PreRated, rcv[1].PreRated)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].PreRated, rcv[0].PreRated, rcv[1].PreRated)
		}
		if !(reflect.DeepEqual(snd[0].Partial, rcv[0].Partial) || reflect.DeepEqual(snd[0].Partial, rcv[1].Partial)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].Partial, rcv[0].Partial, rcv[1].Partial)
		}
		if !reflect.DeepEqual(snd[0].CostDetails, rcv[0].CostDetails) {
			t.Errorf("Expecting: %+v, received: %+v", snd[0].CostDetails, rcv[0].CostDetails)
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
		{
			CGRID:       "88ed9c38005f07576a1e1af293063833b60edcc6",
			RunID:       "1",
			OriginHost:  "host2",
			OriginID:    "2",
			CostDetails: NewBareEventCost(),
		},
		{
			CGRID:       "88ed9c38005f07576a1e1af293063833b60edcc2",
			RunID:       "2",
			OriginHost:  "host2",
			OriginID:    "2",
			CostDetails: NewBareEventCost(),
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
		if !reflect.DeepEqual(snd[0].CostDetails, rcv[0].CostDetails) {
			t.Errorf("Expecting: %+v, received: %+v ", utils.ToJSON(snd[0].CostDetails), utils.ToJSON(rcv[0].CostDetails))
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

func testStorDBitCRUDSMCosts2(t *testing.T) {
	// READ
	if _, err := storDB.GetSMCosts("", "", "", ""); err != utils.ErrNotFound {
		t.Error(err)
	}
	// WRITE
	var snd = []*SMCost{
		{
			CGRID:       "CGRID1",
			RunID:       "11",
			OriginHost:  "host22",
			OriginID:    "O1",
			CostDetails: NewBareEventCost(),
		},
		{
			CGRID:       "CGRID2",
			RunID:       "12",
			OriginHost:  "host22",
			OriginID:    "O2",
			CostDetails: NewBareEventCost(),
		},
		{
			CGRID:       "CGRID3",
			RunID:       "13",
			OriginHost:  "host23",
			OriginID:    "O3",
			CostDetails: NewBareEventCost(),
		},
	}
	for _, smc := range snd {
		if err := storDB.SetSMCost(smc); err != nil {
			t.Error(err)
		}
	}
	// READ
	if rcv, err := storDB.GetSMCosts("", "", "host22", ""); err != nil {
		t.Fatal(err)
	} else if len(rcv) != 2 {
		t.Errorf("Expected 2 results received %v ", len(rcv))
	}
	// REMOVE
	if err := storDB.RemoveSMCosts(&utils.SMCostFilter{
		RunIDs:         []string{"12", "13"},
		NotRunIDs:      []string{"11"},
		OriginHosts:    []string{"host22", "host23"},
		NotOriginHosts: []string{"host21"},
	}); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetSMCosts("", "", "", ""); err != nil {
		t.Error(err)
	} else if len(rcv) != 1 {
		t.Errorf("Expected 1 result received %v ", len(rcv))
	}
	// REMOVE
	if err := storDB.RemoveSMCosts(&utils.SMCostFilter{}); err != nil {
		t.Error(err)
	}
	// READ
	if _, err := storDB.GetSMCosts("", "", "", ""); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testStorDBitFlush(t *testing.T) {
	if err := storDB.Flush(path.Join(cfg.DataFolderPath, "storage", cfg.StorDbCfg().Type)); err != nil {
		t.Error(err)
	}
}

func testStorDBitCRUDVersions(t *testing.T) {
	// CREATE
	vrs := Versions{utils.CostDetails: 1}
	if err := storDB.SetVersions(vrs, true); err != nil {
		t.Error(err)
	}
	if rcv, err := storDB.GetVersions(""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(vrs, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", vrs, rcv)
	}

	// UPDATE
	vrs = Versions{utils.CostDetails: 2, "OTHER_KEY": 1}
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
	if rcv, err := storDB.GetVersions(utils.CostDetails); err != nil {
		t.Error(err)
	} else if len(rcv) != 1 || rcv[utils.CostDetails] != 2 {
		t.Errorf("Received: %+v", rcv)
	}

	if _, err := storDB.GetVersions("UNKNOWN"); err != utils.ErrNotFound {
		t.Error(err)
	}

	vrs = Versions{"UNKNOWN": 1}
	if err := storDB.RemoveVersions(vrs); err != nil {
		t.Error(err)
	}

	if err := storDB.RemoveVersions(nil); err != nil {
		t.Error(err)
	}

	if rcv, err := storDB.GetVersions(""); err != utils.ErrNotFound {
		t.Error(err)
	} else if rcv != nil {
		t.Errorf("Received: %+v", rcv)
	}

}
