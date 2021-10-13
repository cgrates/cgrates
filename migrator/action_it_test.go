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
	actPathIn   string
	actPathOut  string
	actCfgIn    *config.CGRConfig
	actCfgOut   *config.CGRConfig
	actMigrator *Migrator
	actAction   string
)

var sTestsActIT = []func(t *testing.T){
	testActITConnect,
	testActITFlush,
	testActITMigrateAndMove,
}

func TestActionITRedis(t *testing.T) {
	var err error
	actPathIn = path.Join(*dataDir, "conf", "samples", "tutmysql")
	actCfgIn, err = config.NewCGRConfigFromPath(actPathIn)
	if err != nil {
		t.Fatal(err)
	}
	actPathOut = path.Join(*dataDir, "conf", "samples", "tutmysql")
	actCfgOut, err = config.NewCGRConfigFromPath(actPathOut)
	if err != nil {
		t.Fatal(err)
	}
	actAction = utils.Migrate
	for _, stest := range sTestsActIT {
		t.Run("TestActionITMigrateRedis", stest)
	}
	actMigrator.Close()
}

func TestActionITMongo(t *testing.T) {
	var err error
	actPathIn = path.Join(*dataDir, "conf", "samples", "tutmongo")
	actCfgIn, err = config.NewCGRConfigFromPath(actPathIn)
	if err != nil {
		t.Fatal(err)
	}
	actPathOut = path.Join(*dataDir, "conf", "samples", "tutmongo")
	actCfgOut, err = config.NewCGRConfigFromPath(actPathOut)
	if err != nil {
		t.Fatal(err)
	}
	actAction = utils.Migrate
	for _, stest := range sTestsActIT {
		t.Run("TestActionITMigrateMongo", stest)
	}
	actMigrator.Close()
}

func TestActionITMove(t *testing.T) {
	var err error
	actPathIn = path.Join(*dataDir, "conf", "samples", "tutmongo")
	actCfgIn, err = config.NewCGRConfigFromPath(actPathIn)
	if err != nil {
		t.Fatal(err)
	}
	actPathOut = path.Join(*dataDir, "conf", "samples", "tutmysql")
	actCfgOut, err = config.NewCGRConfigFromPath(actPathOut)
	if err != nil {
		t.Fatal(err)
	}
	actAction = utils.Move
	for _, stest := range sTestsActIT {
		t.Run("TestActionITMove", stest)
	}
	actMigrator.Close()
}

func TestActionITMoveEncoding(t *testing.T) {
	var err error
	actPathIn = path.Join(*dataDir, "conf", "samples", "tutmongo")
	actCfgIn, err = config.NewCGRConfigFromPath(actPathIn)
	if err != nil {
		t.Fatal(err)
	}
	actPathOut = path.Join(*dataDir, "conf", "samples", "tutmongojson")
	actCfgOut, err = config.NewCGRConfigFromPath(actPathOut)
	if err != nil {
		t.Fatal(err)
	}
	actAction = utils.Move
	for _, stest := range sTestsActIT {
		t.Run("TestActionITMoveEncoding", stest)
	}
	actMigrator.Close()
}

func TestActionITMigrateMongo2Redis(t *testing.T) {
	var err error
	actPathIn = path.Join(*dataDir, "conf", "samples", "tutmongo")
	actCfgIn, err = config.NewCGRConfigFromPath(actPathIn)
	if err != nil {
		t.Fatal(err)
	}
	actPathOut = path.Join(*dataDir, "conf", "samples", "tutmysql")
	actCfgOut, err = config.NewCGRConfigFromPath(actPathOut)
	if err != nil {
		t.Fatal(err)
	}
	actAction = utils.Migrate
	for _, stest := range sTestsActIT {
		t.Run("TestActionITMigrateMongo2Redis", stest)
	}
	actMigrator.Close()
}

func TestActionITMoveEncoding2(t *testing.T) {
	var err error
	actPathIn = path.Join(*dataDir, "conf", "samples", "tutmysql")
	actCfgIn, err = config.NewCGRConfigFromPath(actPathIn)
	if err != nil {
		t.Fatal(err)
	}
	actPathOut = path.Join(*dataDir, "conf", "samples", "tutmysqljson")
	actCfgOut, err = config.NewCGRConfigFromPath(actPathOut)
	if err != nil {
		t.Fatal(err)
	}
	actAction = utils.Move
	for _, stest := range sTestsActIT {
		t.Run("TestActionITMoveEncoding2", stest)
	}
	actMigrator.Close()
}

func testActITConnect(t *testing.T) {
	dataDBIn, err := NewMigratorDataDB(actCfgIn.DataDbCfg().Type,
		actCfgIn.DataDbCfg().Host, actCfgIn.DataDbCfg().Port,
		actCfgIn.DataDbCfg().Name, actCfgIn.DataDbCfg().User,
		actCfgIn.DataDbCfg().Password, actCfgIn.GeneralCfg().DBDataEncoding,
		config.CgrConfig().CacheCfg(), actCfgIn.DataDbCfg().Opts, nil)
	if err != nil {
		log.Fatal(err)
	}
	dataDBOut, err := NewMigratorDataDB(actCfgOut.DataDbCfg().Type,
		actCfgOut.DataDbCfg().Host, actCfgOut.DataDbCfg().Port,
		actCfgOut.DataDbCfg().Name, actCfgOut.DataDbCfg().User,
		actCfgOut.DataDbCfg().Password, actCfgOut.GeneralCfg().DBDataEncoding,
		config.CgrConfig().CacheCfg(), actCfgOut.DataDbCfg().Opts, nil)
	if err != nil {
		log.Fatal(err)
	}
	if reflect.DeepEqual(actPathIn, actPathOut) {
		actMigrator, err = NewMigrator(dataDBIn, dataDBOut, nil, nil,
			false, true, false, false)
	} else {
		actMigrator, err = NewMigrator(dataDBIn, dataDBOut, nil, nil,
			false, false, false, false)
	}
	if err != nil {
		log.Fatal(err)
	}
}

func testActITFlush(t *testing.T) {
	actMigrator.dmOut.DataManager().DataDB().Flush("")
	actMigrator.dmIN.DataManager().DataDB().Flush("")
	if err := engine.SetDBVersions(actMigrator.dmIN.DataManager().DataDB()); err != nil {
		t.Error("Error  ", err.Error())
	}
	if err := engine.SetDBVersions(actMigrator.dmOut.DataManager().DataDB()); err != nil {
		t.Error("Error  ", err.Error())
	}
}

func testActITMigrateAndMove(t *testing.T) {
	timingSlice := []*engine.RITiming{
		{
			Years:     utils.Years{},
			Months:    utils.Months{},
			MonthDays: utils.MonthDays{},
			WeekDays:  utils.WeekDays{},
		},
	}

	v1act := &v1Action{
		Id:               "test",
		ActionType:       "",
		BalanceType:      "",
		Direction:        "INBOUND",
		ExtraParameters:  "",
		ExpirationString: "",
		Balance: &v1Balance{
			Timings: timingSlice,
		},
	}

	v1acts := &v1Actions{
		v1act,
	}

	act := &engine.Actions{
		&engine.Action{
			Id:               "test",
			ActionType:       "",
			ExtraParameters:  "",
			ExpirationString: "",
			Weight:           0.00,
			Balance: &engine.BalanceFilter{
				Timings: timingSlice,
			},
		},
	}
	switch actAction {
	case utils.Migrate:
		err := actMigrator.dmIN.setV1Actions(v1acts)
		if err != nil {
			t.Error("Error when setting v1 Actions ", err.Error())
		}
		currentVersion := engine.Versions{
			utils.StatS:          2,
			utils.Thresholds:     2,
			utils.Accounts:       2,
			utils.Actions:        1,
			utils.ActionTriggers: 2,
			utils.ActionPlans:    2,
			utils.SharedGroups:   2,
		}
		err = actMigrator.dmIN.DataManager().DataDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for Actions ", err.Error())
		}
		err, _ = actMigrator.Migrate([]string{utils.MetaActions})
		if err != nil {
			t.Error("Error when migrating Actions ", err.Error())
		}
		result, err := actMigrator.dmOut.DataManager().GetActions(v1act.Id, true, utils.NonTransactional)
		if err != nil {
			t.Error("Error when getting Actions ", err.Error())
		}
		if !reflect.DeepEqual(act, &result) {
			t.Errorf("Expecting: %+v, received: %+v", act, &result)
		} else if actMigrator.stats[utils.Actions] != 1 {
			t.Errorf("Expecting: 1, received: %+v", actMigrator.stats[utils.Actions])
		}
	case utils.Move:
		if err := actMigrator.dmIN.DataManager().SetActions(v1act.Id, *act); err != nil {
			t.Error("Error when setting ActionPlan ", err.Error())
		}
		currentVersion := engine.CurrentDataDBVersions()
		err := actMigrator.dmIN.DataManager().DataDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for Actions ", err.Error())
		}
		err, _ = actMigrator.Migrate([]string{utils.MetaActions})
		if err != nil {
			t.Error("Error when migrating Actions ", err.Error())
		}
		result, err := actMigrator.dmOut.DataManager().GetActions(v1act.Id, true, utils.NonTransactional)
		if err != nil {
			t.Error("Error when getting Actions ", err.Error())
		}
		if !reflect.DeepEqual(act, &result) {
			t.Errorf("Expecting: %+v, received: %+v", act, &result)
		} else if actMigrator.stats[utils.Actions] != 1 {
			t.Errorf("Expecting: 1, received: %+v", actMigrator.stats[utils.Actions])
		}
	}
}
