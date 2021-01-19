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

package migrator

import (
	"log"
	"path"
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	ratePrfPathIn   string
	ratePrfPathOut  string
	ratePrfCfgIn    *config.CGRConfig
	ratePrfCfgOut   *config.CGRConfig
	ratePrfMigrator *Migrator
	ratePrfAction   string
)

var sTestsRatePrfIT = []func(t *testing.T){
	testRatePrfITConnect,
	testRatePrfITFlush,
	testRatePrfITMigrateAndMove,
}

func TestRatePrfITMove1(t *testing.T) {
	var err error
	ratePrfPathIn = path.Join(*dataDir, "conf", "samples", "tutmongo")
	ratePrfCfgIn, err = config.NewCGRConfigFromPath(ratePrfPathIn)
	if err != nil {
		t.Fatal(err)
	}
	ratePrfPathOut = path.Join(*dataDir, "conf", "samples", "tutmysql")
	ratePrfCfgOut, err = config.NewCGRConfigFromPath(ratePrfPathOut)
	if err != nil {
		t.Fatal(err)
	}
	ratePrfAction = utils.Move
	for _, stest := range sTestsRatePrfIT {
		t.Run("TestRatePrfITMove", stest)
	}
	ratePrfMigrator.Close()
}

func TestRatePrfITMove2(t *testing.T) {
	var err error
	ratePrfPathIn = path.Join(*dataDir, "conf", "samples", "tutmysql")
	ratePrfCfgIn, err = config.NewCGRConfigFromPath(ratePrfPathIn)
	if err != nil {
		t.Fatal(err)
	}
	ratePrfPathOut = path.Join(*dataDir, "conf", "samples", "tutmongo")
	ratePrfCfgOut, err = config.NewCGRConfigFromPath(ratePrfPathOut)
	if err != nil {
		t.Fatal(err)
	}
	ratePrfAction = utils.Move
	for _, stest := range sTestsRatePrfIT {
		t.Run("TestRatePrfITMove", stest)
	}
	ratePrfMigrator.Close()
}

func TestRatePrfITMoveEncoding(t *testing.T) {
	var err error
	ratePrfPathIn = path.Join(*dataDir, "conf", "samples", "tutmongo")
	ratePrfCfgIn, err = config.NewCGRConfigFromPath(ratePrfPathIn)
	if err != nil {
		t.Fatal(err)
	}
	ratePrfPathOut = path.Join(*dataDir, "conf", "samples", "tutmongojson")
	ratePrfCfgOut, err = config.NewCGRConfigFromPath(ratePrfPathOut)
	if err != nil {
		t.Fatal(err)
	}
	ratePrfAction = utils.Move
	for _, stest := range sTestsRatePrfIT {
		t.Run("TestRatePrfITMoveEncoding", stest)
	}
	ratePrfMigrator.Close()
}

func TestRatePrfITMoveEncoding2(t *testing.T) {
	var err error
	ratePrfPathIn = path.Join(*dataDir, "conf", "samples", "tutmysql")
	ratePrfCfgIn, err = config.NewCGRConfigFromPath(ratePrfPathIn)
	if err != nil {
		t.Fatal(err)
	}
	ratePrfPathOut = path.Join(*dataDir, "conf", "samples", "tutmysqljson")
	ratePrfCfgOut, err = config.NewCGRConfigFromPath(ratePrfPathOut)
	if err != nil {
		t.Fatal(err)
	}
	ratePrfAction = utils.Move
	for _, stest := range sTestsRatePrfIT {
		t.Run("TestRatePrfITMoveEncoding2", stest)
	}
	ratePrfMigrator.Close()
}

func testRatePrfITConnect(t *testing.T) {
	dataDBIn, err := NewMigratorDataDB(ratePrfCfgIn.DataDbCfg().DataDbType,
		ratePrfCfgIn.DataDbCfg().DataDbHost, ratePrfCfgIn.DataDbCfg().DataDbPort,
		ratePrfCfgIn.DataDbCfg().DataDbName, ratePrfCfgIn.DataDbCfg().DataDbUser,
		ratePrfCfgIn.DataDbCfg().DataDbPass, ratePrfCfgIn.GeneralCfg().DBDataEncoding,
		config.CgrConfig().CacheCfg(), ratePrfCfgIn.DataDbCfg().Opts)
	if err != nil {
		log.Fatal(err)
	}
	dataDBOut, err := NewMigratorDataDB(ratePrfCfgOut.DataDbCfg().DataDbType,
		ratePrfCfgOut.DataDbCfg().DataDbHost, ratePrfCfgOut.DataDbCfg().DataDbPort,
		ratePrfCfgOut.DataDbCfg().DataDbName, ratePrfCfgOut.DataDbCfg().DataDbUser,
		ratePrfCfgOut.DataDbCfg().DataDbPass, ratePrfCfgOut.GeneralCfg().DBDataEncoding,
		config.CgrConfig().CacheCfg(), ratePrfCfgOut.DataDbCfg().Opts)
	if err != nil {
		log.Fatal(err)
	}
	if reflect.DeepEqual(ratePrfPathIn, ratePrfPathOut) {
		ratePrfMigrator, err = NewMigrator(dataDBIn, dataDBOut, nil, nil,
			false, true, false, false)
	} else {
		ratePrfMigrator, err = NewMigrator(dataDBIn, dataDBOut, nil, nil,
			false, false, false, false)
	}
	if err != nil {
		log.Fatal(err)
	}
}

func testRatePrfITFlush(t *testing.T) {
	if err := ratePrfMigrator.dmOut.DataManager().DataDB().Flush(""); err != nil {
		t.Error(err)
	}
	if isEmpty, err := ratePrfMigrator.dmOut.DataManager().DataDB().IsDBEmpty(); err != nil {
		t.Error(err)
	} else if isEmpty != true {
		t.Errorf("Expecting: true got :%+v", isEmpty)
	}
	if err := engine.SetDBVersions(ratePrfMigrator.dmOut.DataManager().DataDB()); err != nil {
		t.Error("Error  ", err.Error())
	}
	if err := ratePrfMigrator.dmIN.DataManager().DataDB().Flush(""); err != nil {
		t.Error(err)
	}
	if isEmpty, err := ratePrfMigrator.dmIN.DataManager().DataDB().IsDBEmpty(); err != nil {
		t.Error(err)
	} else if isEmpty != true {
		t.Errorf("Expecting: true got :%+v", isEmpty)
	}
	if err := engine.SetDBVersions(ratePrfMigrator.dmIN.DataManager().DataDB()); err != nil {
		t.Error("Error  ", err.Error())
	}
}

func testRatePrfITMigrateAndMove(t *testing.T) {
	minDec, err := utils.NewDecimalFromUsage("1m")
	if err != nil {
		t.Error(err)
	}
	secDec, err := utils.NewDecimalFromUsage("1s")
	if err != nil {
		t.Error(err)
	}
	rPrf := &engine.RateProfile{
		Tenant:          "cgrates.org",
		ID:              "RP1",
		FilterIDs:       []string{"*string:~*req.Subject:1001"},
		Weight:          0,
		MinCost:         utils.NewDecimal(1, 1),
		MaxCost:         utils.NewDecimal(6, 1),
		MaxCostStrategy: "*free",
		Rates: map[string]*engine.Rate{
			"FIRST_GI": {
				ID:        "FIRST_GI",
				FilterIDs: []string{"*gi:~*req.Usage:0"},
				Weight:    0,

				IntervalRates: []*engine.IntervalRate{
					{
						RecurrentFee: utils.NewDecimal(12, 2),
						Unit:         minDec,
						Increment:    minDec,
					},
				},
				Blocker: false,
			},
			"SECOND_GI": {
				ID:        "SECOND_GI",
				FilterIDs: []string{"*gi:~*req.Usage:1m"},
				Weight:    10,
				IntervalRates: []*engine.IntervalRate{
					{
						RecurrentFee: utils.NewDecimal(6, 2),
						Unit:         minDec,
						Increment:    secDec,
					},
				},
				Blocker: false,
			},
		},
	}
	if err := rPrf.Compile(); err != nil {
		t.Fatal(err)
	}
	switch ratePrfAction {
	case utils.Migrate: //QQ for the moment only one version of rate profiles exists
	case utils.Move:
		if err := ratePrfMigrator.dmIN.DataManager().SetRateProfile(rPrf, true); err != nil {
			t.Error(err)
		}
		currentVersion := engine.CurrentDataDBVersions()
		err := ratePrfMigrator.dmIN.DataManager().DataDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for RatePrf ", err.Error())
		}

		_, err = ratePrfMigrator.dmOut.DataManager().GetRateProfile("cgrates.org", "RP1", false, false, utils.NonTransactional)
		if err != utils.ErrNotFound {
			t.Error(err)
		}

		err, _ = ratePrfMigrator.Migrate([]string{utils.MetaRateProfiles})
		if err != nil {
			t.Error("Error when migrating RatePrf ", err.Error())
		}
		ratePrfult, err := ratePrfMigrator.dmOut.DataManager().GetRateProfile("cgrates.org", "RP1", false, false, utils.NonTransactional)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(ratePrfult, rPrf) {
			t.Errorf("Expecting: %+v, received: %+v", rPrf, ratePrfult)
		}
		ratePrfult, err = ratePrfMigrator.dmIN.DataManager().GetRateProfile("cgrates.org", "RP1", false, false, utils.NonTransactional)
		if err != utils.ErrNotFound {
			t.Error(err)
		} else if ratePrfMigrator.stats[utils.RateProfiles] != 1 {
			t.Errorf("Expected 1, received: %v", ratePrfMigrator.stats[utils.RateProfiles])
		}
	}
}
