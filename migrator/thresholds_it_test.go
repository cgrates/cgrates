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
	trsPathIn     string
	trsPathOut    string
	trsCfgIn      *config.CGRConfig
	trsCfgOut     *config.CGRConfig
	trsMigrator   *Migrator
	trsThresholds string
)

var sTestsTrsIT = []func(t *testing.T){
	testTrsITConnect,
	testTrsITFlush,
	testTrsITMigrateAndMove,
}

func TestThresholdsITRedis(t *testing.T) {
	var err error
	trsPathIn = path.Join(*dataDir, "conf", "samples", "tutmysql")
	trsCfgIn, err = config.NewCGRConfigFromFolder(trsPathIn)
	if err != nil {
		t.Fatal(err)
	}
	trsCfgOut, err = config.NewCGRConfigFromFolder(trsPathIn)
	if err != nil {
		t.Fatal(err)
	}
	trsThresholds = utils.Migrate
	for _, stest := range sTestsTrsIT {
		t.Run("TestThresholdsITMigrateRedis", stest)
	}
}

func TestThresholdsITMongo(t *testing.T) {
	var err error
	trsPathIn = path.Join(*dataDir, "conf", "samples", "tutmongo")
	trsCfgIn, err = config.NewCGRConfigFromFolder(trsPathIn)
	if err != nil {
		t.Fatal(err)
	}
	trsCfgOut, err = config.NewCGRConfigFromFolder(trsPathIn)
	if err != nil {
		t.Fatal(err)
	}
	trsThresholds = utils.Migrate
	for _, stest := range sTestsTrsIT {
		t.Run("TestThresholdsITMigrateMongo", stest)
	}
}

func TestThresholdsITMove(t *testing.T) {
	var err error
	trsPathIn = path.Join(*dataDir, "conf", "samples", "tutmongo")
	trsCfgIn, err = config.NewCGRConfigFromFolder(trsPathIn)
	if err != nil {
		t.Fatal(err)
	}
	trsPathOut = path.Join(*dataDir, "conf", "samples", "tutmysql")
	trsCfgOut, err = config.NewCGRConfigFromFolder(trsPathOut)
	if err != nil {
		t.Fatal(err)
	}
	trsThresholds = utils.Move
	for _, stest := range sTestsTrsIT {
		t.Run("TestThresholdsITMove", stest)
	}
}

func TestThresholdsITMoveEncoding(t *testing.T) {
	var err error
	trsPathIn = path.Join(*dataDir, "conf", "samples", "tutmongo")
	trsCfgIn, err = config.NewCGRConfigFromFolder(trsPathIn)
	if err != nil {
		t.Fatal(err)
	}
	trsPathOut = path.Join(*dataDir, "conf", "samples", "tutmongojson")
	trsCfgOut, err = config.NewCGRConfigFromFolder(trsPathOut)
	if err != nil {
		t.Fatal(err)
	}
	trsThresholds = utils.Move
	for _, stest := range sTestsTrsIT {
		t.Run("TestThresholdsITMoveEncoding", stest)
	}
}

func TestThresholdsITMoveEncoding2(t *testing.T) {
	var err error
	trsPathIn = path.Join(*dataDir, "conf", "samples", "tutmysql")
	trsCfgIn, err = config.NewCGRConfigFromFolder(trsPathIn)
	if err != nil {
		t.Fatal(err)
	}
	trsPathOut = path.Join(*dataDir, "conf", "samples", "tutmysqljson")
	trsCfgOut, err = config.NewCGRConfigFromFolder(trsPathOut)
	if err != nil {
		t.Fatal(err)
	}
	trsThresholds = utils.Move
	for _, stest := range sTestsTrsIT {
		t.Run("TestThresholdsITMoveEncoding2", stest)
	}
}

func testTrsITConnect(t *testing.T) {
	dataDBIn, err := NewMigratorDataDB(trsCfgIn.DataDbCfg().DataDbType,
		trsCfgIn.DataDbCfg().DataDbHost, trsCfgIn.DataDbCfg().DataDbPort,
		trsCfgIn.DataDbCfg().DataDbName, trsCfgIn.DataDbCfg().DataDbUser,
		trsCfgIn.DataDbCfg().DataDbPass, trsCfgIn.GeneralCfg().DBDataEncoding,
		config.CgrConfig().CacheCfg(), "")
	if err != nil {
		log.Fatal(err)
	}
	dataDBOut, err := NewMigratorDataDB(trsCfgOut.DataDbCfg().DataDbType,
		trsCfgOut.DataDbCfg().DataDbHost, trsCfgOut.DataDbCfg().DataDbPort,
		trsCfgOut.DataDbCfg().DataDbName, trsCfgOut.DataDbCfg().DataDbUser,
		trsCfgOut.DataDbCfg().DataDbPass, trsCfgOut.GeneralCfg().DBDataEncoding,
		config.CgrConfig().CacheCfg(), "")
	if err != nil {
		log.Fatal(err)
	}
	trsMigrator, err = NewMigrator(dataDBIn, dataDBOut,
		nil, nil,
		false, false, false)
	if err != nil {
		log.Fatal(err)
	}
}

func testTrsITFlush(t *testing.T) {
	trsMigrator.dmOut.DataManager().DataDB().Flush("")
	if err := engine.SetDBVersions(trsMigrator.dmOut.DataManager().DataDB()); err != nil {
		t.Error("Error  ", err.Error())
	}
}

func testTrsITMigrateAndMove(t *testing.T) {
	tim := time.Date(2012, time.February, 27, 23, 59, 59, 0, time.UTC)
	var filters []*engine.FilterRule
	v1trs := &v2ActionTrigger{
		ID:             "test2",              // original csv tag
		UniqueID:       "testUUID",           // individual id
		ThresholdType:  "*min_event_counter", //*min_event_counter, *max_event_counter, *min_balance_counter, *max_balance_counter, *min_balance, *max_balance, *balance_expired
		ThresholdValue: 5.32,
		Recurrent:      false,                          // reset excuted flag each run
		MinSleep:       time.Duration(5) * time.Second, // Minimum duration between two executions in case of recurrent triggers
		ExpirationDate: tim,
		ActivationDate: tim,
		Balance: &engine.BalanceFilter{
			ID:             utils.StringPointer("TESTZ"),
			Timings:        []*engine.RITiming{},
			ExpirationDate: utils.TimePointer(tim),
			Type:           utils.StringPointer(utils.MONETARY),
			Directions:     utils.StringMapPointer(utils.NewStringMap(utils.OUT)),
		},
		Weight:            0,
		ActionsID:         "Action1",
		MinQueuedItems:    10, // Trigger actions only if this number is hit (stats only)
		Executed:          false,
		LastExecutionTime: time.Now(),
	}
	x, err := engine.NewFilterRule(engine.MetaRSR, "Directions", v1trs.Balance.Directions.Slice())
	if err != nil {
		t.Error("Error when creating new NewFilterRule", err.Error())
	}
	filters = append(filters, x)

	tresProf := &engine.ThresholdProfile{
		ID:                 v1trs.ID,
		Tenant:             config.CgrConfig().GeneralCfg().DefaultTenant,
		Weight:             v1trs.Weight,
		ActivationInterval: &utils.ActivationInterval{v1trs.ExpirationDate, v1trs.ActivationDate},
		MinSleep:           v1trs.MinSleep,
	}

	v2trs := &v2Threshold{
		Tenant:    "cgrates.org",
		ID:        "th_rec",
		FilterIDs: []string{},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
		},
		Recurrent: true,
		MinHits:   0,
		MinSleep:  time.Duration(5 * time.Minute),
		Blocker:   false,
		Weight:    20.0,
		ActionIDs: []string{},
		Async:     false,
	}

	tresProf2 := &engine.ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "th_rec",
		FilterIDs: []string{},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
		},
		MaxHits:   -1,
		MinHits:   0,
		MinSleep:  time.Duration(5 * time.Minute),
		Blocker:   false,
		Weight:    20.0,
		ActionIDs: []string{},
		Async:     false,
	}

	v2trs_nonrec := &v2Threshold{
		Tenant:    "cgrates.org",
		ID:        "th_nonrec",
		FilterIDs: []string{},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
		},
		Recurrent: false,
		MinHits:   0,
		MinSleep:  time.Duration(5 * time.Minute),
		Blocker:   false,
		Weight:    20.0,
		ActionIDs: []string{},
		Async:     false,
	}

	tresProf3 := &engine.ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "th_nonrec",
		FilterIDs: []string{},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
		},
		MaxHits:   1,
		MinHits:   0,
		MinSleep:  time.Duration(5 * time.Minute),
		Blocker:   false,
		Weight:    20.0,
		ActionIDs: []string{},
		Async:     false,
	}

	switch trsThresholds {
	case utils.Migrate:
		err := trsMigrator.dmIN.setV2ActionTrigger(v1trs)
		if err != nil {
			t.Error("Error when setting v1 Thresholds ", err.Error())
		}
		currentVersion := engine.Versions{utils.StatS: 2, utils.Thresholds: 1, utils.Accounts: 2, utils.Actions: 2, utils.ActionTriggers: 2, utils.ActionPlans: 2, utils.SharedGroups: 2}
		err = trsMigrator.dmOut.DataManager().DataDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for Thresholds ", err.Error())
		}
		err, _ = trsMigrator.Migrate([]string{utils.MetaThresholds})
		if err != nil {
			t.Error("Error when migrating Thresholds ", err.Error())
		}
		result, err := trsMigrator.dmOut.DataManager().GetThresholdProfile(tresProf.Tenant, tresProf.ID, false, false, utils.NonTransactional)
		if err != nil {
			t.Error("Error when getting Thresholds ", err.Error())
		}
		if !reflect.DeepEqual(tresProf.ID, result.ID) {
			t.Errorf("Expecting: %+v, received: %+v", tresProf.ID, result.ID)
		} else if !reflect.DeepEqual(tresProf.Tenant, result.Tenant) {
			t.Errorf("Expecting: %+v, received: %+v", tresProf.Tenant, result.Tenant)
		} else if !reflect.DeepEqual(tresProf.Weight, result.Weight) {
			t.Errorf("Expecting: %+v, received: %+v", tresProf.Weight, result.Weight)
		} else if !reflect.DeepEqual(tresProf.ActivationInterval, result.ActivationInterval) {
			t.Errorf("Expecting: %+v, received: %+v", tresProf.ActivationInterval, result.ActivationInterval)
		} else if !reflect.DeepEqual(tresProf.MinSleep, result.MinSleep) {
			t.Errorf("Expecting: %+v, received: %+v", tresProf.MinSleep, result.MinSleep)
		}
		//Migrate V2Threshold to NewThreshold
		err = trsMigrator.dmIN.setV2ThresholdProfile(v2trs)
		if err != nil {
			t.Error("Error when setting v1 Thresholds ", err.Error())
		}
		err = trsMigrator.dmIN.setV2ThresholdProfile(v2trs_nonrec)
		if err != nil {
			t.Error("Error when setting v1 Thresholds ", err.Error())
		}

		currentVersion = engine.Versions{utils.StatS: 2, utils.Thresholds: 2, utils.Accounts: 2, utils.Actions: 2, utils.ActionTriggers: 2, utils.ActionPlans: 2, utils.SharedGroups: 2}
		err = trsMigrator.dmOut.DataManager().DataDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for Thresholds ", err.Error())
		}
		err, _ = trsMigrator.Migrate([]string{utils.MetaThresholds})
		if err != nil {
			t.Error("Error when migrating Thresholds ", err.Error())
		}

		result, err = trsMigrator.dmOut.DataManager().GetThresholdProfile(tresProf2.Tenant, tresProf2.ID, false, false, utils.NonTransactional)
		if err != nil {
			t.Error("Error when getting Thresholds ", err.Error())
		}
		if !reflect.DeepEqual(tresProf2, result) {
			t.Errorf("Expectong: %+v, received: %+v", utils.ToJSON(tresProf2), utils.ToJSON(result))
		}

		result, err = trsMigrator.dmOut.DataManager().GetThresholdProfile(tresProf3.Tenant, tresProf3.ID, false, false, utils.NonTransactional)
		if err != nil {
			t.Error("Error when getting Thresholds ", err.Error())
		}
		if !reflect.DeepEqual(tresProf3, result) {
			t.Errorf("Expectong: %+v, received: %+v", utils.ToJSON(tresProf3), utils.ToJSON(result))
		}

	case utils.Move:
		if err := trsMigrator.dmIN.DataManager().SetThresholdProfile(tresProf, false); err != nil {
			t.Error("Error when setting Thresholds ", err.Error())
		}
		currentVersion := engine.CurrentDataDBVersions()
		err := trsMigrator.dmOut.DataManager().DataDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for Thresholds ", err.Error())
		}
		err, _ = trsMigrator.Migrate([]string{utils.MetaThresholds})
		if err != nil {
			t.Error("Error when migrating Thresholds ", err.Error())
		}
		result, err := trsMigrator.dmOut.DataManager().GetThresholdProfile(tresProf.Tenant, tresProf.ID, false, false, utils.NonTransactional)
		if err != nil {
			t.Error("Error when getting Thresholds ", err.Error())
		}
		if !reflect.DeepEqual(tresProf.ID, result.ID) {
			t.Errorf("Expecting: %+v, received: %+v", tresProf.ID, result.ID)
		} else if !reflect.DeepEqual(tresProf.Tenant, result.Tenant) {
			t.Errorf("Expecting: %+v, received: %+v", tresProf.Tenant, result.Tenant)
		} else if !reflect.DeepEqual(tresProf.Weight, result.Weight) {
			t.Errorf("Expecting: %+v, received: %+v", tresProf.Weight, result.Weight)
		} else if !reflect.DeepEqual(tresProf.ActivationInterval, result.ActivationInterval) {
			t.Errorf("Expecting: %+v, received: %+v", tresProf.ActivationInterval, result.ActivationInterval)
		} else if !reflect.DeepEqual(tresProf.MinSleep, result.MinSleep) {
			t.Errorf("Expecting: %+v, received: %+v", tresProf.MinSleep, result.MinSleep)
		}
	}
}
