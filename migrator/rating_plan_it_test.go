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
	rtplPathIn   string
	rtplPathOut  string
	rtplCfgIn    *config.CGRConfig
	rtplCfgOut   *config.CGRConfig
	rtplMigrator *Migrator
	rtplAction   string
)

var sTestsRtPlIT = []func(t *testing.T){
	testRtPlITConnect,
	testRtPlITFlush,
	testRtPlITMigrateAndMove,
}

func TestRatingPlanITMove1(t *testing.T) {
	var err error
	rtplPathIn = path.Join(*dataDir, "conf", "samples", "tutmongo")
	rtplCfgIn, err = config.NewCGRConfigFromPath(rtplPathIn)
	if err != nil {
		t.Fatal(err)
	}
	rtplPathOut = path.Join(*dataDir, "conf", "samples", "tutmysql")
	rtplCfgOut, err = config.NewCGRConfigFromPath(rtplPathOut)
	if err != nil {
		t.Fatal(err)
	}
	rtplAction = utils.Move
	for _, stest := range sTestsRtPlIT {
		t.Run("TestRatingPlanITMove", stest)
	}
	rtplMigrator.Close()
}

func TestRatingPlanITMove2(t *testing.T) {
	var err error
	rtplPathIn = path.Join(*dataDir, "conf", "samples", "tutmysql")
	rtplCfgIn, err = config.NewCGRConfigFromPath(rtplPathIn)
	if err != nil {
		t.Fatal(err)
	}
	rtplPathOut = path.Join(*dataDir, "conf", "samples", "tutmongo")
	rtplCfgOut, err = config.NewCGRConfigFromPath(rtplPathOut)
	if err != nil {
		t.Fatal(err)
	}
	rtplAction = utils.Move
	for _, stest := range sTestsRtPlIT {
		t.Run("TestRatingPlanITMove", stest)
	}
	rtplMigrator.Close()
}

func TestRatingPlanITMoveEncoding(t *testing.T) {
	var err error
	rtplPathIn = path.Join(*dataDir, "conf", "samples", "tutmongo")
	rtplCfgIn, err = config.NewCGRConfigFromPath(rtplPathIn)
	if err != nil {
		t.Fatal(err)
	}
	rtplPathOut = path.Join(*dataDir, "conf", "samples", "tutmongojson")
	rtplCfgOut, err = config.NewCGRConfigFromPath(rtplPathOut)
	if err != nil {
		t.Fatal(err)
	}
	rtplAction = utils.Move
	for _, stest := range sTestsRtPlIT {
		t.Run("TestRatingPlanITMoveEncoding", stest)
	}
	rtplMigrator.Close()
}

func TestRatingPlanITMoveEncoding2(t *testing.T) {
	var err error
	rtplPathIn = path.Join(*dataDir, "conf", "samples", "tutmysql")
	rtplCfgIn, err = config.NewCGRConfigFromPath(rtplPathIn)
	if err != nil {
		t.Fatal(err)
	}
	rtplPathOut = path.Join(*dataDir, "conf", "samples", "tutmysqljson")
	rtplCfgOut, err = config.NewCGRConfigFromPath(rtplPathOut)
	if err != nil {
		t.Fatal(err)
	}
	rtplAction = utils.Move
	for _, stest := range sTestsRtPlIT {
		t.Run("TestRatingPlanITMoveEncoding2", stest)
	}
	rtplMigrator.Close()
}

func testRtPlITConnect(t *testing.T) {
	dataDBIn, err := NewMigratorDataDB(rtplCfgIn.DataDbCfg().Type,
		rtplCfgIn.DataDbCfg().Host, rtplCfgIn.DataDbCfg().Port,
		rtplCfgIn.DataDbCfg().Name, rtplCfgIn.DataDbCfg().User,
		rtplCfgIn.DataDbCfg().Password, rtplCfgIn.GeneralCfg().DBDataEncoding,
		config.CgrConfig().CacheCfg(), ratePrfCfgIn.DataDbCfg().Opts)
	if err != nil {
		log.Fatal(err)
	}
	dataDBOut, err := NewMigratorDataDB(rtplCfgOut.DataDbCfg().Type,
		rtplCfgOut.DataDbCfg().Host, rtplCfgOut.DataDbCfg().Port,
		rtplCfgOut.DataDbCfg().Name, rtplCfgOut.DataDbCfg().User,
		rtplCfgOut.DataDbCfg().Password, rtplCfgOut.GeneralCfg().DBDataEncoding,
		config.CgrConfig().CacheCfg(), ratePrfCfgOut.DataDbCfg().Opts)
	if err != nil {
		log.Fatal(err)
	}
	if rtplPathIn == rtplPathOut {
		rtplMigrator, err = NewMigrator(dataDBIn, dataDBOut, nil, nil,
			false, true, false, false)
	} else {
		rtplMigrator, err = NewMigrator(dataDBIn, dataDBOut, nil, nil,
			false, false, false, false)
	}
	if err != nil {
		log.Fatal(err)
	}
}

func testRtPlITFlush(t *testing.T) {
	if err := rtplMigrator.dmOut.DataManager().DataDB().Flush(""); err != nil {
		t.Error(err)
	}
	if isEmpty, err := rtplMigrator.dmOut.DataManager().DataDB().IsDBEmpty(); err != nil {
		t.Error(err)
	} else if isEmpty != true {
		t.Errorf("Expecting: true got :%+v", isEmpty)
	}
	if err := engine.SetDBVersions(rtplMigrator.dmOut.DataManager().DataDB()); err != nil {
		t.Error("Error  ", err.Error())
	}
	if err := rtplMigrator.dmIN.DataManager().DataDB().Flush(""); err != nil {
		t.Error(err)
	}
	if isEmpty, err := rtplMigrator.dmIN.DataManager().DataDB().IsDBEmpty(); err != nil {
		t.Error(err)
	} else if isEmpty != true {
		t.Errorf("Expecting: true got :%+v", isEmpty)
	}
	if err := engine.SetDBVersions(rtplMigrator.dmIN.DataManager().DataDB()); err != nil {
		t.Error("Error  ", err.Error())
	}
}

func testRtPlITMigrateAndMove(t *testing.T) {
	rtplPlan := &engine.RatingPlan{
		Id:      "RT_PLAN1",
		Timings: map[string]*engine.RITiming{},
		Ratings: map[string]*engine.RIRate{
			"asjkilj": &engine.RIRate{
				ConnectFee:       10,
				RoundingMethod:   utils.MetaRoundingUp,
				RoundingDecimals: 1,
				MaxCost:          10,
			},
		},
		DestinationRates: map[string]engine.RPRateList{},
	}
	switch rtplAction {
	case utils.Migrate: // for the momment only one version of rating plans exists
	case utils.Move:
		if err := rtplMigrator.dmIN.DataManager().SetRatingPlan(rtplPlan, utils.NonTransactional); err != nil {
			t.Error(err)
		}
		currentVersion := engine.CurrentDataDBVersions()
		err := rtplMigrator.dmIN.DataManager().DataDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for RatingPlan ", err.Error())
		}

		_, err = rtplMigrator.dmOut.DataManager().GetRatingPlan("RT_PLAN1", true, utils.NonTransactional)
		if err != utils.ErrNotFound {
			t.Error(err)
		}

		err, _ = rtplMigrator.Migrate([]string{utils.MetaRatingPlans})
		if err != nil {
			t.Error("Error when migrating RatingPlan ", err.Error())
		}
		result, err := rtplMigrator.dmOut.DataManager().GetRatingPlan("RT_PLAN1", true, utils.NonTransactional)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(result, rtplPlan) {
			t.Errorf("Expecting: %+v, received: %+v", rtplPlan, result)
		}
		result, err = rtplMigrator.dmIN.DataManager().GetRatingPlan("RT_PLAN1", true, utils.NonTransactional)
		if err != utils.ErrNotFound {
			t.Error(err)
		} else if rtplMigrator.stats[utils.RatingPlan] != 1 {
			t.Errorf("Expected 1, received: %v", rtplMigrator.stats[utils.RatingPlan])
		}
	}
}
