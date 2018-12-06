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
	actCfgIn, err = config.NewCGRConfigFromFolder(actPathIn)
	if err != nil {
		t.Fatal(err)
	}
	actCfgOut, err = config.NewCGRConfigFromFolder(actPathIn)
	if err != nil {
		t.Fatal(err)
	}
	actAction = utils.Migrate
	for _, stest := range sTestsActIT {
		t.Run("TestActionITMigrateRedis", stest)
	}
}

func TestActionITMongo(t *testing.T) {
	var err error
	actPathIn = path.Join(*dataDir, "conf", "samples", "tutmongo")
	actCfgIn, err = config.NewCGRConfigFromFolder(actPathIn)
	if err != nil {
		t.Fatal(err)
	}
	actCfgOut, err = config.NewCGRConfigFromFolder(actPathIn)
	if err != nil {
		t.Fatal(err)
	}
	actAction = utils.Migrate
	for _, stest := range sTestsActIT {
		t.Run("TestActionITMigrateMongo", stest)
	}
}

func TestActionITMove(t *testing.T) {
	var err error
	actPathIn = path.Join(*dataDir, "conf", "samples", "tutmongo")
	actCfgIn, err = config.NewCGRConfigFromFolder(actPathIn)
	if err != nil {
		t.Fatal(err)
	}
	actPathOut = path.Join(*dataDir, "conf", "samples", "tutmysql")
	actCfgOut, err = config.NewCGRConfigFromFolder(actPathOut)
	if err != nil {
		t.Fatal(err)
	}
	actAction = utils.Move
	for _, stest := range sTestsActIT {
		t.Run("TestActionITMove", stest)
	}
}

func TestActionITMoveEncoding(t *testing.T) {
	var err error
	actPathIn = path.Join(*dataDir, "conf", "samples", "tutmongo")
	actCfgIn, err = config.NewCGRConfigFromFolder(actPathIn)
	if err != nil {
		t.Fatal(err)
	}
	actPathOut = path.Join(*dataDir, "conf", "samples", "tutmongojson")
	actCfgOut, err = config.NewCGRConfigFromFolder(actPathOut)
	if err != nil {
		t.Fatal(err)
	}
	actAction = utils.Move
	for _, stest := range sTestsActIT {
		t.Run("TestActionITMoveEncoding", stest)
	}
}

/*
func TestActionITMoveEncoding2(t *testing.T) {
	var err error
	actPathIn = path.Join(*dataDir, "conf", "samples", "tutmysql")
	actCfgIn, err = config.NewCGRConfigFromFolder(actPathIn)
	if err != nil {
		t.Fatal(err)
	}
	actPathOut = path.Join(*dataDir, "conf", "samples", "tutmysqljson")
	actCfgOut, err = config.NewCGRConfigFromFolder(actPathOut)
	if err != nil {
		t.Fatal(err)
	}
	actAction = utils.Move
	for _, stest := range sTestsActIT {
		t.Run("TestActionITMoveEncoding2", stest)
	}
}*/

func testActITConnect(t *testing.T) {
	dataDBIn, err := NewMigratorDataDB(actCfgIn.DataDbCfg().DataDbType,
		actCfgIn.DataDbCfg().DataDbHost, actCfgIn.DataDbCfg().DataDbPort,
		actCfgIn.DataDbCfg().DataDbName, actCfgIn.DataDbCfg().DataDbUser,
		actCfgIn.DataDbCfg().DataDbPass, actCfgIn.GeneralCfg().DBDataEncoding,
		config.CgrConfig().CacheCfg(), "")
	if err != nil {
		log.Fatal(err)
	}
	dataDBOut, err := NewMigratorDataDB(actCfgOut.DataDbCfg().DataDbType,
		actCfgOut.DataDbCfg().DataDbHost, actCfgOut.DataDbCfg().DataDbPort,
		actCfgOut.DataDbCfg().DataDbName, actCfgOut.DataDbCfg().DataDbUser,
		actCfgOut.DataDbCfg().DataDbPass, actCfgOut.GeneralCfg().DBDataEncoding,
		config.CgrConfig().CacheCfg(), "")
	if err != nil {
		log.Fatal(err)
	}
	actMigrator, err = NewMigrator(dataDBIn, dataDBOut,
		nil, nil,
		false, false, false)
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
		currentVersion := engine.Versions{utils.StatS: 2, utils.Thresholds: 2, utils.Accounts: 2, utils.Actions: 1, utils.ActionTriggers: 2, utils.ActionPlans: 2, utils.SharedGroups: 2}
		err = actMigrator.dmOut.DataManager().DataDB().SetVersions(currentVersion, false)
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
		}
	case utils.Move:
		if err := actMigrator.dmIN.DataManager().SetActions(v1act.Id, *act, utils.NonTransactional); err != nil {
			t.Error("Error when setting ActionPlan ", err.Error())
		}
		currentVersion := engine.CurrentDataDBVersions()
		err := actMigrator.dmOut.DataManager().DataDB().SetVersions(currentVersion, false)
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
		}
	}
}
