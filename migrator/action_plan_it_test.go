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
	actPlnPathIn   string
	actPlnPathOut  string
	actPlnCfgIn    *config.CGRConfig
	actPlnCfgOut   *config.CGRConfig
	actPlnMigrator *Migrator
	actActionPlan  string
)

var sTestsActPlnIT = []func(t *testing.T){
	testActPlnITConnect,
	testActPlnITFlush,
	testActPlnITMigrateAndMove,
}

func TestActionPlanITRedis(t *testing.T) {
	var err error
	actPlnPathIn = path.Join(*dataDir, "conf", "samples", "tutmysql")
	actPlnCfgIn, err = config.NewCGRConfigFromFolder(actPlnPathIn)
	if err != nil {
		t.Fatal(err)
	}
	actPlnCfgOut, err = config.NewCGRConfigFromFolder(actPlnPathIn)
	if err != nil {
		t.Fatal(err)
	}
	actActionPlan = utils.Migrate
	for _, stest := range sTestsActPlnIT {
		t.Run("TestActionPlanITMigrateRedis", stest)
	}
}

func TestActionPlanITMongo(t *testing.T) {
	var err error
	actPlnPathIn = path.Join(*dataDir, "conf", "samples", "tutmongo")
	actPlnCfgIn, err = config.NewCGRConfigFromFolder(actPlnPathIn)
	if err != nil {
		t.Fatal(err)
	}
	actPlnCfgOut, err = config.NewCGRConfigFromFolder(actPlnPathIn)
	if err != nil {
		t.Fatal(err)
	}
	actActionPlan = utils.Migrate
	for _, stest := range sTestsActPlnIT {
		t.Run("TestActionPlanITMigrateMongo", stest)
	}
}

func TestActionPlanITMove(t *testing.T) {
	var err error
	actPlnPathIn = path.Join(*dataDir, "conf", "samples", "tutmongo")
	actPlnCfgIn, err = config.NewCGRConfigFromFolder(actPlnPathIn)
	if err != nil {
		t.Fatal(err)
	}
	actPlnPathOut = path.Join(*dataDir, "conf", "samples", "tutmysql")
	actPlnCfgOut, err = config.NewCGRConfigFromFolder(actPlnPathOut)
	if err != nil {
		t.Fatal(err)
	}
	actActionPlan = utils.Move
	for _, stest := range sTestsActPlnIT {
		t.Run("TestActionPlanITMove", stest)
	}
}

func TestActionPlanITMoveEncoding(t *testing.T) {
	var err error
	actPlnPathIn = path.Join(*dataDir, "conf", "samples", "tutmongo")
	actPlnCfgIn, err = config.NewCGRConfigFromFolder(actPlnPathIn)
	if err != nil {
		t.Fatal(err)
	}
	actPlnPathOut = path.Join(*dataDir, "conf", "samples", "tutmongojson")
	actPlnCfgOut, err = config.NewCGRConfigFromFolder(actPlnPathOut)
	if err != nil {
		t.Fatal(err)
	}
	actActionPlan = utils.Move
	for _, stest := range sTestsActPlnIT {
		t.Run("TestActionPlanITMoveEncoding", stest)
	}
}

func TestActionPlanITMoveEncoding2(t *testing.T) {
	var err error
	actPlnPathIn = path.Join(*dataDir, "conf", "samples", "tutmysql")
	actPlnCfgIn, err = config.NewCGRConfigFromFolder(actPlnPathIn)
	if err != nil {
		t.Fatal(err)
	}
	actPlnPathOut = path.Join(*dataDir, "conf", "samples", "tutmysqljson")
	actPlnCfgOut, err = config.NewCGRConfigFromFolder(actPlnPathOut)
	if err != nil {
		t.Fatal(err)
	}
	actActionPlan = utils.Move
	for _, stest := range sTestsActPlnIT {
		t.Run("TestActionPlanITMoveEncoding2", stest)
	}
}

func testActPlnITConnect(t *testing.T) {
	dataDBIn, err := NewMigratorDataDB(actPlnCfgIn.DataDbCfg().DataDbType,
		actPlnCfgIn.DataDbCfg().DataDbHost, actPlnCfgIn.DataDbCfg().DataDbPort,
		actPlnCfgIn.DataDbCfg().DataDbName, actPlnCfgIn.DataDbCfg().DataDbUser,
		actPlnCfgIn.DataDbCfg().DataDbPass, actPlnCfgIn.GeneralCfg().DBDataEncoding,
		config.CgrConfig().CacheCfg(), "")
	if err != nil {
		log.Fatal(err)
	}
	dataDBOut, err := NewMigratorDataDB(actPlnCfgOut.DataDbCfg().DataDbType,
		actPlnCfgOut.DataDbCfg().DataDbHost, actPlnCfgOut.DataDbCfg().DataDbPort,
		actPlnCfgOut.DataDbCfg().DataDbName, actPlnCfgOut.DataDbCfg().DataDbUser,
		actPlnCfgOut.DataDbCfg().DataDbPass, actPlnCfgOut.GeneralCfg().DBDataEncoding,
		config.CgrConfig().CacheCfg(), "")
	if err != nil {
		log.Fatal(err)
	}
	actPlnMigrator, err = NewMigrator(dataDBIn, dataDBOut,
		nil, nil,
		false, false, false)
	if err != nil {
		log.Fatal(err)
	}
}

func testActPlnITFlush(t *testing.T) {
	actPlnMigrator.dmOut.DataManager().DataDB().Flush("")
	if err := engine.SetDBVersions(actPlnMigrator.dmOut.DataManager().DataDB()); err != nil {
		t.Error("Error  ", err.Error())
	}
}

func testActPlnITMigrateAndMove(t *testing.T) {
	timingSlice := &engine.RITiming{
		Years:     utils.Years{},
		Months:    utils.Months{},
		MonthDays: utils.MonthDays{},
		WeekDays:  utils.WeekDays{},
	}

	v1actPln := &v1ActionPlans{
		&v1ActionPlan{
			Id:         "test",
			AccountIds: []string{"one"},
			Timing: &engine.RateInterval{
				Timing: timingSlice,
			},
		},
	}

	actPln := &engine.ActionPlan{
		Id:         "test",
		AccountIDs: utils.StringMap{"one": true},
		ActionTimings: []*engine.ActionTiming{
			{
				Timing: &engine.RateInterval{
					Timing: timingSlice,
				},
			},
		},
	}

	switch actActionPlan {
	case utils.Migrate:
		err := actPlnMigrator.dmIN.setV1ActionPlans(v1actPln)
		if err != nil {
			t.Error("Error when setting v1 ActionPlan ", err.Error())
		}
		currentVersion := engine.Versions{utils.StatS: 2, utils.Thresholds: 2, utils.Accounts: 2, utils.Actions: 2, utils.ActionTriggers: 2, utils.ActionPlans: 1, utils.SharedGroups: 2}
		err = actPlnMigrator.dmOut.DataManager().DataDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for ActionPlan ", err.Error())
		}
		err, _ = actPlnMigrator.Migrate([]string{utils.MetaActionPlans})
		if err != nil {
			t.Error("Error when migrating ActionPlan ", err.Error())
		}
		result, err := actPlnMigrator.dmOut.DataManager().DataDB().GetActionPlan((*v1actPln)[0].Id, true, utils.NonTransactional)
		if err != nil {
			t.Error("Error when getting ActionPlan ", err.Error())
		}
		// compared fields, uuid is generated in ActionTiming
		if !reflect.DeepEqual(actPln.Id, result.Id) {
			t.Errorf("Expecting: %+v, received: %+v", actPln.Id, result.Id)
		} else if !reflect.DeepEqual(actPln.AccountIDs, result.AccountIDs) {
			t.Errorf("Expecting: %+v, received: %+v", actPln.AccountIDs, result.AccountIDs)
		} else if !reflect.DeepEqual(actPln.ActionTimings[0].Timing, result.ActionTimings[0].Timing) {
			t.Errorf("Expecting: %+v, received: %+v", actPln.ActionTimings[0].Timing, result.ActionTimings[0].Timing)
		}
	case utils.Move:
		if err := actPlnMigrator.dmIN.DataManager().DataDB().SetActionPlan((*v1actPln)[0].Id, actPln, true, utils.NonTransactional); err != nil {
			t.Error("Error when setting ActionPlan ", err.Error())
		}
		currentVersion := engine.CurrentDataDBVersions()
		err := actPlnMigrator.dmOut.DataManager().DataDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for ActionPlan ", err.Error())
		}
		err, _ = actPlnMigrator.Migrate([]string{utils.MetaActionPlans})
		if err != nil {
			t.Error("Error when migrating ActionPlan ", err.Error())
		}
		result, err := actPlnMigrator.dmOut.DataManager().DataDB().GetActionPlan((*v1actPln)[0].Id, true, utils.NonTransactional)
		if err != nil {
			t.Error("Error when getting ActionPlan ", err.Error())
		}
		// compared fields, uuid is generated in ActionTiming
		if !reflect.DeepEqual(actPln.Id, result.Id) {
			t.Errorf("Expecting: %+v, received: %+v", actPln.Id, result.Id)
		} else if !reflect.DeepEqual(actPln.AccountIDs, result.AccountIDs) {
			t.Errorf("Expecting: %+v, received: %+v", actPln.AccountIDs, result.AccountIDs)
		} else if !reflect.DeepEqual(actPln.ActionTimings[0].Timing, result.ActionTimings[0].Timing) {
			t.Errorf("Expecting: %+v, received: %+v", actPln.ActionTimings[0].Timing, result.ActionTimings[0].Timing)
		}
		result, err = actPlnMigrator.dmIN.DataManager().DataDB().GetActionPlan((*v1actPln)[0].Id, true, utils.NonTransactional)
		if err != utils.ErrNotFound {
			t.Error(err)
		}
	}
}
