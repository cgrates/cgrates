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
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	actTrgPathIn     string
	actTrgPathOut    string
	actTrgCfgIn      *config.CGRConfig
	actTrgCfgOut     *config.CGRConfig
	actTrgMigrator   *Migrator
	actActionTrigger string
)

var sTestsActTrgIT = []func(t *testing.T){
	testActTrgITConnect,
	testActTrgITFlush,
	testActTrgITMigrateAndMove,
}

func TestActionTriggerITRedis(t *testing.T) {
	var err error
	actTrgPathIn = path.Join(*dataDir, "conf", "samples", "tutmysql")
	actTrgCfgIn, err = config.NewCGRConfigFromPath(actTrgPathIn)
	if err != nil {
		t.Fatal(err)
	}
	actTrgPathOut = path.Join(*dataDir, "conf", "samples", "tutmysql")
	actTrgCfgOut, err = config.NewCGRConfigFromPath(actTrgPathOut)
	if err != nil {
		t.Fatal(err)
	}
	actActionTrigger = utils.Migrate
	for _, stest := range sTestsActTrgIT {
		t.Run("TestActionTriggerITMigrateRedis", stest)
	}
	actTrgMigrator.Close()
}

func TestActionTriggerITMongo(t *testing.T) {
	var err error
	actTrgPathIn = path.Join(*dataDir, "conf", "samples", "tutmongo")
	actTrgCfgIn, err = config.NewCGRConfigFromPath(actTrgPathIn)
	if err != nil {
		t.Fatal(err)
	}
	actTrgPathOut = path.Join(*dataDir, "conf", "samples", "tutmongo")
	actTrgCfgOut, err = config.NewCGRConfigFromPath(actTrgPathOut)
	if err != nil {
		t.Fatal(err)
	}
	actActionTrigger = utils.Migrate
	for _, stest := range sTestsActTrgIT {
		t.Run("TestActionTriggerITMigrateMongo", stest)
	}
	actTrgMigrator.Close()
}

func TestActionTriggerITMove(t *testing.T) {
	var err error
	actTrgPathIn = path.Join(*dataDir, "conf", "samples", "tutmongo")
	actTrgCfgIn, err = config.NewCGRConfigFromPath(actTrgPathIn)
	if err != nil {
		t.Fatal(err)
	}
	actTrgPathOut = path.Join(*dataDir, "conf", "samples", "tutmysql")
	actTrgCfgOut, err = config.NewCGRConfigFromPath(actTrgPathOut)
	if err != nil {
		t.Fatal(err)
	}
	actActionTrigger = utils.Move
	for _, stest := range sTestsActTrgIT {
		t.Run("TestActionTriggerITMove", stest)
	}
	actTrgMigrator.Close()
}

func TestActionTriggerITMoveEncoding(t *testing.T) {
	var err error
	actTrgPathIn = path.Join(*dataDir, "conf", "samples", "tutmongo")
	actTrgCfgIn, err = config.NewCGRConfigFromPath(actTrgPathIn)
	if err != nil {
		t.Fatal(err)
	}
	actTrgPathOut = path.Join(*dataDir, "conf", "samples", "tutmongojson")
	actTrgCfgOut, err = config.NewCGRConfigFromPath(actTrgPathOut)
	if err != nil {
		t.Fatal(err)
	}
	actActionTrigger = utils.Move
	for _, stest := range sTestsActTrgIT {
		t.Run("TestActionTriggerITMoveEncoding", stest)
	}
	actTrgMigrator.Close()
}

func TestActionTriggerITMoveEncoding2(t *testing.T) {
	var err error
	actTrgPathIn = path.Join(*dataDir, "conf", "samples", "tutmysql")
	actTrgCfgIn, err = config.NewCGRConfigFromPath(actTrgPathIn)
	if err != nil {
		t.Fatal(err)
	}
	actTrgPathOut = path.Join(*dataDir, "conf", "samples", "tutmysqljson")
	actTrgCfgOut, err = config.NewCGRConfigFromPath(actTrgPathOut)
	if err != nil {
		t.Fatal(err)
	}
	actActionTrigger = utils.Move
	for _, stest := range sTestsActTrgIT {
		t.Run("TestActionTriggerITMoveEncoding2", stest)
	}
	actTrgMigrator.Close()
}

func testActTrgITConnect(t *testing.T) {
	dataDBIn, err := NewMigratorDataDB(actTrgCfgIn.DataDbCfg().Type,
		actTrgCfgIn.DataDbCfg().Host, actTrgCfgIn.DataDbCfg().Port,
		actTrgCfgIn.DataDbCfg().Name, actTrgCfgIn.DataDbCfg().User,
		actTrgCfgIn.DataDbCfg().Password, actTrgCfgIn.GeneralCfg().DBDataEncoding,
		config.CgrConfig().CacheCfg(), actTrgCfgIn.DataDbCfg().Opts, nil)
	if err != nil {
		log.Fatal(err)
	}
	dataDBOut, err := NewMigratorDataDB(actTrgCfgOut.DataDbCfg().Type,
		actTrgCfgOut.DataDbCfg().Host, actTrgCfgOut.DataDbCfg().Port,
		actTrgCfgOut.DataDbCfg().Name, actTrgCfgOut.DataDbCfg().User,
		actTrgCfgOut.DataDbCfg().Password, actTrgCfgOut.GeneralCfg().DBDataEncoding,
		config.CgrConfig().CacheCfg(), actTrgCfgOut.DataDbCfg().Opts, nil)
	if err != nil {
		log.Fatal(err)
	}
	if reflect.DeepEqual(actTrgPathIn, actTrgPathOut) {
		actTrgMigrator, err = NewMigrator(dataDBIn, dataDBOut, nil, nil,
			false, true, false, false)
	} else {
		actTrgMigrator, err = NewMigrator(dataDBIn, dataDBOut, nil, nil,
			false, false, false, false)
	}
	if err != nil {
		log.Fatal(err)
	}
}

func testActTrgITFlush(t *testing.T) {
	actTrgMigrator.dmOut.DataManager().DataDB().Flush("")
	actTrgMigrator.dmIN.DataManager().DataDB().Flush("")
	if err := engine.SetDBVersions(actTrgMigrator.dmOut.DataManager().DataDB()); err != nil {
		t.Error("Error  ", err.Error())
	}
}

func testActTrgITMigrateAndMove(t *testing.T) {
	tim := time.Date(2012, time.February, 27, 23, 59, 59, 0, time.UTC)
	v1actTrg := &v1ActionTriggers{
		&v1ActionTrigger{
			Id:                    "Test",
			BalanceType:           "*monetary",
			BalanceDirection:      "*out",
			ThresholdType:         "*max_balance",
			ThresholdValue:        2,
			ActionsId:             "TEST_ACTIONS",
			Executed:              true,
			BalanceExpirationDate: tim,
		},
	}
	actTrg := engine.ActionTriggers{
		&engine.ActionTrigger{
			ID: "Test",
			Balance: &engine.BalanceFilter{
				ExpirationDate: utils.TimePointer(tim),
				Type:           utils.StringPointer(utils.MetaMonetary),
			},
			ExpirationDate:    tim,
			LastExecutionTime: tim,
			ActivationDate:    tim,
			ThresholdType:     utils.TriggerMaxBalance,
			ThresholdValue:    2,
			ActionsID:         "TEST_ACTIONS",
			Executed:          true,
		},
	}

	switch actActionTrigger {
	case utils.Migrate:
		err := actTrgMigrator.dmIN.setV1ActionTriggers(v1actTrg)
		if err != nil {
			t.Error("Error when setting v1 ActionTriggers ", err.Error())
		}
		currentVersion := engine.Versions{utils.StatS: 2, utils.Thresholds: 2, utils.Accounts: 2, utils.Actions: 2, utils.ActionTriggers: 1, utils.ActionPlans: 2, utils.SharedGroups: 2}
		err = actTrgMigrator.dmIN.DataManager().DataDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for ActionTriggers ", err.Error())
		}
		err, _ = actTrgMigrator.Migrate([]string{utils.MetaActionTriggers})
		if err != nil {
			t.Error("Error when migrating ActionTriggers ", err.Error())
		}
		result, err := actTrgMigrator.dmOut.DataManager().GetActionTriggers((*v1actTrg)[0].Id, false, utils.NonTransactional)
		if err != nil {
			t.Error("Error when getting ActionTriggers ", err.Error())
		}
		if !reflect.DeepEqual(actTrg, result) {
			t.Errorf("Expecting: %+v,\nReceived: %+v", utils.ToJSON(actTrg), utils.ToJSON(result))
		} else if actTrgMigrator.stats[utils.ActionTriggers] != 1 {
			t.Errorf("Expecting: 1, received: %+v", actTrgMigrator.stats[utils.ActionTriggers])
		}
	case utils.Move:
		if err := actTrgMigrator.dmIN.DataManager().SetActionTriggers((*v1actTrg)[0].Id, actTrg); err != nil {
			t.Error("Error when setting ActionTriggers ", err.Error())
		}
		currentVersion := engine.CurrentDataDBVersions()
		err := actTrgMigrator.dmIN.DataManager().DataDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for ActionTriggers ", err.Error())
		}
		err, _ = actTrgMigrator.Migrate([]string{utils.MetaActionTriggers})
		if err != nil {
			t.Error("Error when migrating ActionTriggers ", err.Error())
		}
		result, err := actTrgMigrator.dmOut.DataManager().GetActionTriggers((*v1actTrg)[0].Id, false, utils.NonTransactional)
		if err != nil {
			t.Error("Error when getting ActionTriggers ", err.Error())
		}
		if !reflect.DeepEqual(actTrg, result) {
			t.Errorf("Expecting: %+v, received: %+v", actTrg, result)
		}
		if actTrgMigrator.stats[utils.ActionTriggers] != 1 {
			t.Errorf("Expecting: 1, received: %+v", actTrgMigrator.stats[utils.ActionTriggers])
		}
	}
}
