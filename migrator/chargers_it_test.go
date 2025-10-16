//go:build integration
// +build integration

/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package migrator

import (
	"log"
	"path"
	"reflect"
	"testing"

	"github.com/cgrates/birpc/context"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	chrgPathIn   string
	chrgPathOut  string
	chrgCfgIn    *config.CGRConfig
	chrgCfgOut   *config.CGRConfig
	chrgMigrator *Migrator
	chrgAction   string
)

var sTestsChrgIT = []func(t *testing.T){
	testChrgITConnect,
	testChrgITFlush,
	testChrgITMigrateAndMove,
}

func TestChargersITMove1(t *testing.T) {
	var err error
	chrgPathIn = path.Join(*utils.DataDir, "conf", "samples", "tutmongo")
	chrgCfgIn, err = config.NewCGRConfigFromPath(context.Background(), chrgPathIn)
	if err != nil {
		t.Fatal(err)
	}
	chrgPathOut = path.Join(*utils.DataDir, "conf", "samples", "tutmysql")
	chrgCfgOut, err = config.NewCGRConfigFromPath(context.Background(), chrgPathOut)
	if err != nil {
		t.Fatal(err)
	}
	chrgAction = utils.Move
	for _, stest := range sTestsChrgIT {
		t.Run("TestChargersITMove", stest)
	}
	chrgMigrator.Close()
}

func TestChargersITMove2(t *testing.T) {
	var err error
	chrgPathIn = path.Join(*utils.DataDir, "conf", "samples", "tutmysql")
	chrgCfgIn, err = config.NewCGRConfigFromPath(context.Background(), chrgPathIn)
	if err != nil {
		t.Fatal(err)
	}
	chrgPathOut = path.Join(*utils.DataDir, "conf", "samples", "tutmongo")
	chrgCfgOut, err = config.NewCGRConfigFromPath(context.Background(), chrgPathOut)
	if err != nil {
		t.Fatal(err)
	}
	chrgAction = utils.Move
	for _, stest := range sTestsChrgIT {
		t.Run("TestChargersITMove", stest)
	}
	chrgMigrator.Close()
}

func TestChargersITMoveEncoding(t *testing.T) {
	var err error
	chrgPathIn = path.Join(*utils.DataDir, "conf", "samples", "tutmongo")
	chrgCfgIn, err = config.NewCGRConfigFromPath(context.Background(), chrgPathIn)
	if err != nil {
		t.Fatal(err)
	}
	chrgPathOut = path.Join(*utils.DataDir, "conf", "samples", "tutmongojson")
	chrgCfgOut, err = config.NewCGRConfigFromPath(context.Background(), chrgPathOut)
	if err != nil {
		t.Fatal(err)
	}
	chrgAction = utils.Move
	for _, stest := range sTestsChrgIT {
		t.Run("TestChargersITMoveEncoding", stest)
	}
	chrgMigrator.Close()
}

func TestChargersITMoveEncoding2(t *testing.T) {
	var err error
	chrgPathIn = path.Join(*utils.DataDir, "conf", "samples", "tutmysql")
	chrgCfgIn, err = config.NewCGRConfigFromPath(context.Background(), chrgPathIn)
	if err != nil {
		t.Fatal(err)
	}
	chrgPathOut = path.Join(*utils.DataDir, "conf", "samples", "tutmysqljson")
	chrgCfgOut, err = config.NewCGRConfigFromPath(context.Background(), chrgPathOut)
	if err != nil {
		t.Fatal(err)
	}
	chrgAction = utils.Move
	for _, stest := range sTestsChrgIT {
		t.Run("TestChargersITMoveEncoding2", stest)
	}
	chrgMigrator.Close()
}

func testChrgITConnect(t *testing.T) {
	dataDBIn, err := NewMigratorDataDBs([]string{utils.MetaDefault}, chrgCfgIn.GeneralCfg().DBDataEncoding, chrgCfgIn)
	if err != nil {
		log.Fatal(err)
	}
	dataDBOut, err := NewMigratorDataDBs([]string{utils.MetaDefault}, chrgCfgOut.GeneralCfg().DBDataEncoding, chrgCfgOut)
	if err != nil {
		log.Fatal(err)
	}
	if reflect.DeepEqual(chrgPathIn, chrgPathOut) {
		chrgMigrator, err = NewMigrator(chrgCfgOut.DbCfg(), dataDBIn, dataDBOut,
			false, true)
	} else {
		chrgMigrator, err = NewMigrator(chrgCfgOut.DbCfg(), dataDBIn, dataDBOut,
			false, false)
	}
	if err != nil {
		log.Fatal(err)
	}
}

func testChrgITFlush(t *testing.T) {
	if err := chrgMigrator.dmTo[utils.MetaDefault].DataManager().DataDB()[utils.MetaDefault].Flush(""); err != nil {
		t.Error(err)
	}
	if isEmpty, err := chrgMigrator.dmTo[utils.MetaDefault].DataManager().DataDB()[utils.MetaDefault].IsDBEmpty(); err != nil {
		t.Error(err)
	} else if isEmpty != true {
		t.Errorf("Expecting: true got :%+v", isEmpty)
	}
	if err := engine.SetDBVersions(chrgMigrator.dmTo[utils.MetaDefault].DataManager().DataDB()[utils.MetaDefault]); err != nil {
		t.Error("Error  ", err.Error())
	}
	if err := chrgMigrator.dmFrom[utils.MetaDefault].DataManager().DataDB()[utils.MetaDefault].Flush(""); err != nil {
		t.Error(err)
	}
	if isEmpty, err := chrgMigrator.dmFrom[utils.MetaDefault].DataManager().DataDB()[utils.MetaDefault].IsDBEmpty(); err != nil {
		t.Error(err)
	} else if isEmpty != true {
		t.Errorf("Expecting: true got :%+v", isEmpty)
	}
	if err := engine.SetDBVersions(chrgMigrator.dmFrom[utils.MetaDefault].DataManager().DataDB()[utils.MetaDefault]); err != nil {
		t.Error("Error  ", err.Error())
	}
}

func testChrgITMigrateAndMove(t *testing.T) {
	chrgPrf := &utils.ChargerProfile{
		Tenant:       "cgrates.org",
		ID:           "CHRG_1",
		FilterIDs:    []string{"*string:Accont:1001", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z|2014-07-14T14:25:00Z"},
		AttributeIDs: []string{"ATTR_1"},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
	}
	chrgPrf2 := &utils.ChargerProfile{
		Tenant:       "cgrates.com",
		ID:           "CHRG_1",
		FilterIDs:    []string{"*string:Accont:1001", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z|2014-07-14T14:25:00Z"},
		AttributeIDs: []string{"ATTR_1"},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
	}
	switch chrgAction {
	case utils.Migrate: // for the momment only one version of chargers exists
	case utils.Move:
		if err := chrgMigrator.dmFrom[utils.MetaDefault].DataManager().SetChargerProfile(context.Background(), chrgPrf, false); err != nil {
			t.Error(err)
		}
		if err := chrgMigrator.dmFrom[utils.MetaDefault].DataManager().SetChargerProfile(context.Background(), chrgPrf2, false); err != nil {
			t.Error(err)
		}
		currentVersion := engine.CurrentDataDBVersions()
		err := chrgMigrator.dmFrom[utils.MetaDefault].DataManager().DataDB()[utils.MetaDefault].SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for Chargers ", err.Error())
		}

		_, err = chrgMigrator.dmTo[utils.MetaDefault].DataManager().GetChargerProfile(context.Background(), "cgrates.org",
			"CHRG_1", false, false, utils.NonTransactional)
		if err != utils.ErrNotFound {
			t.Error(err)
		}

		err, _ = chrgMigrator.Migrate([]string{utils.MetaChargers})
		if err != nil {
			t.Error("Error when migrating Chargers ", err.Error())
		}
		if result, err := chrgMigrator.dmTo[utils.MetaDefault].DataManager().GetChargerProfile(context.Background(), "cgrates.org",
			"CHRG_1", false, false, utils.NonTransactional); err != nil {
			t.Fatal(err)
		} else if !reflect.DeepEqual(result, chrgPrf) {
			t.Errorf("Expecting: %+v, received: %+v", chrgPrf, result)
		}
		if result, err := chrgMigrator.dmTo[utils.MetaDefault].DataManager().GetChargerProfile(context.Background(), "cgrates.com",
			"CHRG_1", false, false, utils.NonTransactional); err != nil {
			t.Fatal(err)
		} else if !reflect.DeepEqual(result, chrgPrf2) {
			t.Errorf("Expecting: %+v, received: %+v", chrgPrf2, result)
		}
		if _, err = chrgMigrator.dmFrom[utils.MetaDefault].DataManager().GetChargerProfile(context.Background(), "cgrates.org",
			"CHRG_1", false, false, utils.NonTransactional); err == nil || err != utils.ErrNotFound {
			t.Error(err)
		}
		if _, err = chrgMigrator.dmFrom[utils.MetaDefault].DataManager().GetChargerProfile(context.Background(), "cgrates.com",
			"CHRG_1", false, false, utils.NonTransactional); err == nil || err != utils.ErrNotFound {
			t.Error(err)
		} else if chrgMigrator.stats[utils.Chargers] != 2 {
			t.Errorf("Expected 2, received: %v", chrgMigrator.stats[utils.Chargers])
		}
	}
}
