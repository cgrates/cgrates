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
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	rtprflPathIn   string
	rtprflPathOut  string
	rtprflCfgIn    *config.CGRConfig
	rtprflCfgOut   *config.CGRConfig
	rtprflMigrator *Migrator
	rtprflAction   string
)

var sTestsRtPrfIT = []func(t *testing.T){
	testRtPrfITConnect,
	testRtPrfITFlush,
	testRtPrfITMigrateAndMove,
}

func TestRatingProfileITMove1(t *testing.T) {
	var err error
	rtprflPathIn = path.Join(*dataDir, "conf", "samples", "tutmongo")
	rtprflCfgIn, err = config.NewCGRConfigFromPath(rtprflPathIn)
	if err != nil {
		t.Fatal(err)
	}
	rtprflPathOut = path.Join(*dataDir, "conf", "samples", "tutmysql")
	rtprflCfgOut, err = config.NewCGRConfigFromPath(rtprflPathOut)
	if err != nil {
		t.Fatal(err)
	}
	rtprflAction = utils.Move
	for _, stest := range sTestsRtPrfIT {
		t.Run("TestRatingProfileITMove", stest)
	}
	rtprflMigrator.Close()
}

func TestRatingProfileITMove2(t *testing.T) {
	var err error
	rtprflPathIn = path.Join(*dataDir, "conf", "samples", "tutmysql")
	rtprflCfgIn, err = config.NewCGRConfigFromPath(rtprflPathIn)
	if err != nil {
		t.Fatal(err)
	}
	rtprflPathOut = path.Join(*dataDir, "conf", "samples", "tutmongo")
	rtprflCfgOut, err = config.NewCGRConfigFromPath(rtprflPathOut)
	if err != nil {
		t.Fatal(err)
	}
	rtprflAction = utils.Move
	for _, stest := range sTestsRtPrfIT {
		t.Run("TestRatingProfileITMove", stest)
	}
	rtprflMigrator.Close()
}

func TestRatingProfileITMoveEncoding(t *testing.T) {
	var err error
	rtprflPathIn = path.Join(*dataDir, "conf", "samples", "tutmongo")
	rtprflCfgIn, err = config.NewCGRConfigFromPath(rtprflPathIn)
	if err != nil {
		t.Fatal(err)
	}
	rtprflPathOut = path.Join(*dataDir, "conf", "samples", "tutmongojson")
	rtprflCfgOut, err = config.NewCGRConfigFromPath(rtprflPathOut)
	if err != nil {
		t.Fatal(err)
	}
	rtprflAction = utils.Move
	for _, stest := range sTestsRtPrfIT {
		t.Run("TestRatingProfileITMoveEncoding", stest)
	}
	rtprflMigrator.Close()
}

func TestRatingProfileITMoveEncoding2(t *testing.T) {
	var err error
	rtprflPathIn = path.Join(*dataDir, "conf", "samples", "tutmysql")
	rtprflCfgIn, err = config.NewCGRConfigFromPath(rtprflPathIn)
	if err != nil {
		t.Fatal(err)
	}
	rtprflPathOut = path.Join(*dataDir, "conf", "samples", "tutmysqljson")
	rtprflCfgOut, err = config.NewCGRConfigFromPath(rtprflPathOut)
	if err != nil {
		t.Fatal(err)
	}
	rtprflAction = utils.Move
	for _, stest := range sTestsRtPrfIT {
		t.Run("TestRatingProfileITMoveEncoding2", stest)
	}
	rtprflMigrator.Close()
}

func testRtPrfITConnect(t *testing.T) {
	dataDBIn, err := NewMigratorDataDB(rtprflCfgIn.DataDbCfg().DataDbType,
		rtprflCfgIn.DataDbCfg().DataDbHost, rtprflCfgIn.DataDbCfg().DataDbPort,
		rtprflCfgIn.DataDbCfg().DataDbName, rtprflCfgIn.DataDbCfg().DataDbUser,
		rtprflCfgIn.DataDbCfg().DataDbPass, rtprflCfgIn.GeneralCfg().DBDataEncoding,
		config.CgrConfig().CacheCfg(), rtprflCfgIn.DataDbCfg().Opts)
	if err != nil {
		log.Fatal(err)
	}
	dataDBOut, err := NewMigratorDataDB(rtprflCfgOut.DataDbCfg().DataDbType,
		rtprflCfgOut.DataDbCfg().DataDbHost, rtprflCfgOut.DataDbCfg().DataDbPort,
		rtprflCfgOut.DataDbCfg().DataDbName, rtprflCfgOut.DataDbCfg().DataDbUser,
		rtprflCfgOut.DataDbCfg().DataDbPass, rtprflCfgOut.GeneralCfg().DBDataEncoding,
		config.CgrConfig().CacheCfg(), rtplCfgOut.DataDbCfg().Opts)
	if err != nil {
		log.Fatal(err)
	}
	if reflect.DeepEqual(rtprflPathIn, rtprflPathOut) {
		rtprflMigrator, err = NewMigrator(dataDBIn, dataDBOut, nil, nil,
			false, true, false, false)
	} else {
		rtprflMigrator, err = NewMigrator(dataDBIn, dataDBOut, nil, nil,
			false, false, false, false)
	}
	if err != nil {
		log.Fatal(err)
	}
}

func testRtPrfITFlush(t *testing.T) {
	if err := rtprflMigrator.dmOut.DataManager().DataDB().Flush(""); err != nil {
		t.Error(err)
	}
	if isEmpty, err := rtprflMigrator.dmOut.DataManager().DataDB().IsDBEmpty(); err != nil {
		t.Error(err)
	} else if isEmpty != true {
		t.Errorf("Expecting: true got :%+v", isEmpty)
	}
	if err := engine.SetDBVersions(rtprflMigrator.dmOut.DataManager().DataDB()); err != nil {
		t.Error("Error  ", err.Error())
	}
	if err := rtprflMigrator.dmIN.DataManager().DataDB().Flush(""); err != nil {
		t.Error(err)
	}
	if isEmpty, err := rtprflMigrator.dmIN.DataManager().DataDB().IsDBEmpty(); err != nil {
		t.Error(err)
	} else if isEmpty != true {
		t.Errorf("Expecting: true got :%+v", isEmpty)
	}
	if err := engine.SetDBVersions(rtprflMigrator.dmIN.DataManager().DataDB()); err != nil {
		t.Error("Error  ", err.Error())
	}
}

func testRtPrfITMigrateAndMove(t *testing.T) {
	rtprfl := &engine.RatingProfile{
		Id: "RT_Profile",
		RatingPlanActivations: engine.RatingPlanActivations{
			&engine.RatingPlanActivation{
				ActivationTime: time.Now().Round(time.Second).UTC(),
				RatingPlanId:   "RP_PLAN1",
				FallbackKeys:   []string{"1001"},
			},
		},
	}
	switch rtprflAction {
	case utils.Migrate: // for the momment only one version of rating plans exists
	case utils.Move:
		if err := rtprflMigrator.dmIN.DataManager().SetRatingProfile(rtprfl, utils.NonTransactional); err != nil {
			t.Error(err)
		}
		currentVersion := engine.CurrentDataDBVersions()
		err := rtprflMigrator.dmIN.DataManager().DataDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for RatingProfile ", err.Error())
		}

		_, err = rtprflMigrator.dmOut.DataManager().GetRatingProfile("RT_Profile", true, utils.NonTransactional)
		if err != utils.ErrNotFound {
			t.Error(err)
		}

		err, _ = rtprflMigrator.Migrate([]string{utils.MetaRatingProfiles})
		if err != nil {
			t.Error("Error when migrating RatingProfile ", err.Error())
		}
		result, err := rtprflMigrator.dmOut.DataManager().GetRatingProfile("RT_Profile", true, utils.NonTransactional)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(result, rtprfl) {
			t.Errorf("Expecting: %+v, received: %+v", rtprfl, result)
		}
		result, err = rtprflMigrator.dmIN.DataManager().GetRatingProfile("RT_Profile", true, utils.NonTransactional)
		if err != utils.ErrNotFound {
			t.Error(err)
		} else if rtprflMigrator.stats[utils.RatingProfile] != 1 {
			t.Errorf("Expected 1, received: %v", rtprflMigrator.stats[utils.RatingProfile])
		}
	}
}
