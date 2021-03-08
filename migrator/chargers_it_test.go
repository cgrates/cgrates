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
	chrgPathIn = path.Join(*dataDir, "conf", "samples", "tutmongo")
	chrgCfgIn, err = config.NewCGRConfigFromPath(chrgPathIn)
	if err != nil {
		t.Fatal(err)
	}
	chrgPathOut = path.Join(*dataDir, "conf", "samples", "tutmysql")
	chrgCfgOut, err = config.NewCGRConfigFromPath(chrgPathOut)
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
	chrgPathIn = path.Join(*dataDir, "conf", "samples", "tutmysql")
	chrgCfgIn, err = config.NewCGRConfigFromPath(chrgPathIn)
	if err != nil {
		t.Fatal(err)
	}
	chrgPathOut = path.Join(*dataDir, "conf", "samples", "tutmongo")
	chrgCfgOut, err = config.NewCGRConfigFromPath(chrgPathOut)
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
	chrgPathIn = path.Join(*dataDir, "conf", "samples", "tutmongo")
	chrgCfgIn, err = config.NewCGRConfigFromPath(chrgPathIn)
	if err != nil {
		t.Fatal(err)
	}
	chrgPathOut = path.Join(*dataDir, "conf", "samples", "tutmongojson")
	chrgCfgOut, err = config.NewCGRConfigFromPath(chrgPathOut)
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
	chrgPathIn = path.Join(*dataDir, "conf", "samples", "tutmysql")
	chrgCfgIn, err = config.NewCGRConfigFromPath(chrgPathIn)
	if err != nil {
		t.Fatal(err)
	}
	chrgPathOut = path.Join(*dataDir, "conf", "samples", "tutmysqljson")
	chrgCfgOut, err = config.NewCGRConfigFromPath(chrgPathOut)
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
	dataDBIn, err := NewMigratorDataDB(chrgCfgIn.DataDbCfg().Type,
		chrgCfgIn.DataDbCfg().Host, chrgCfgIn.DataDbCfg().Port,
		chrgCfgIn.DataDbCfg().Name, chrgCfgIn.DataDbCfg().User,
		chrgCfgIn.DataDbCfg().Password, chrgCfgIn.GeneralCfg().DBDataEncoding,
		config.CgrConfig().CacheCfg(), chrgCfgIn.DataDbCfg().Opts)
	if err != nil {
		log.Fatal(err)
	}
	dataDBOut, err := NewMigratorDataDB(chrgCfgOut.DataDbCfg().Type,
		chrgCfgOut.DataDbCfg().Host, chrgCfgOut.DataDbCfg().Port,
		chrgCfgOut.DataDbCfg().Name, chrgCfgOut.DataDbCfg().User,
		chrgCfgOut.DataDbCfg().Password, chrgCfgOut.GeneralCfg().DBDataEncoding,
		config.CgrConfig().CacheCfg(), chrgCfgOut.DataDbCfg().Opts)
	if err != nil {
		log.Fatal(err)
	}
	if reflect.DeepEqual(chrgPathIn, chrgPathOut) {
		chrgMigrator, err = NewMigrator(dataDBIn, dataDBOut, nil, nil,
			false, true, false, false)
	} else {
		chrgMigrator, err = NewMigrator(dataDBIn, dataDBOut, nil, nil,
			false, false, false, false)
	}
	if err != nil {
		log.Fatal(err)
	}
}

func testChrgITFlush(t *testing.T) {
	if err := chrgMigrator.dmOut.DataManager().DataDB().Flush(""); err != nil {
		t.Error(err)
	}
	if isEmpty, err := chrgMigrator.dmOut.DataManager().DataDB().IsDBEmpty(); err != nil {
		t.Error(err)
	} else if isEmpty != true {
		t.Errorf("Expecting: true got :%+v", isEmpty)
	}
	if err := engine.SetDBVersions(chrgMigrator.dmOut.DataManager().DataDB()); err != nil {
		t.Error("Error  ", err.Error())
	}
	if err := chrgMigrator.dmIN.DataManager().DataDB().Flush(""); err != nil {
		t.Error(err)
	}
	if isEmpty, err := chrgMigrator.dmIN.DataManager().DataDB().IsDBEmpty(); err != nil {
		t.Error(err)
	} else if isEmpty != true {
		t.Errorf("Expecting: true got :%+v", isEmpty)
	}
	if err := engine.SetDBVersions(chrgMigrator.dmIN.DataManager().DataDB()); err != nil {
		t.Error("Error  ", err.Error())
	}
}

func testChrgITMigrateAndMove(t *testing.T) {
	chrgPrf := &engine.ChargerProfile{
		Tenant:    "cgrates.org",
		ID:        "CHRG_1",
		FilterIDs: []string{"*string:Accont:1001"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		AttributeIDs: []string{"ATTR_1"},
		Weight:       20,
	}
	chrgPrf2 := &engine.ChargerProfile{
		Tenant:    "cgrates.com",
		ID:        "CHRG_1",
		FilterIDs: []string{"*string:Accont:1001"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		AttributeIDs: []string{"ATTR_1"},
		Weight:       20,
	}
	switch chrgAction {
	case utils.Migrate: // for the momment only one version of chargers exists
	case utils.Move:
		if err := chrgMigrator.dmIN.DataManager().SetChargerProfile(chrgPrf, false); err != nil {
			t.Error(err)
		}
		if err := chrgMigrator.dmIN.DataManager().SetChargerProfile(chrgPrf2, false); err != nil {
			t.Error(err)
		}
		currentVersion := engine.CurrentDataDBVersions()
		err := chrgMigrator.dmIN.DataManager().DataDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for Chargers ", err.Error())
		}

		_, err = chrgMigrator.dmOut.DataManager().GetChargerProfile("cgrates.org",
			"CHRG_1", false, false, utils.NonTransactional)
		if err != utils.ErrNotFound {
			t.Error(err)
		}

		err, _ = chrgMigrator.Migrate([]string{utils.MetaChargers})
		if err != nil {
			t.Error("Error when migrating Chargers ", err.Error())
		}
		if result, err := chrgMigrator.dmOut.DataManager().GetChargerProfile("cgrates.org",
			"CHRG_1", false, false, utils.NonTransactional); err != nil {
			t.Fatal(err)
		} else if !reflect.DeepEqual(result, chrgPrf) {
			t.Errorf("Expecting: %+v, received: %+v", chrgPrf, result)
		}
		if result, err := chrgMigrator.dmOut.DataManager().GetChargerProfile("cgrates.com",
			"CHRG_1", false, false, utils.NonTransactional); err != nil {
			t.Fatal(err)
		} else if !reflect.DeepEqual(result, chrgPrf2) {
			t.Errorf("Expecting: %+v, received: %+v", chrgPrf2, result)
		}
		if _, err = chrgMigrator.dmIN.DataManager().GetChargerProfile("cgrates.org",
			"CHRG_1", false, false, utils.NonTransactional); err == nil || err != utils.ErrNotFound {
			t.Error(err)
		}
		if _, err = chrgMigrator.dmIN.DataManager().GetChargerProfile("cgrates.com",
			"CHRG_1", false, false, utils.NonTransactional); err == nil || err != utils.ErrNotFound {
			t.Error(err)
		} else if chrgMigrator.stats[utils.Chargers] != 2 {
			t.Errorf("Expected 2, received: %v", chrgMigrator.stats[utils.Chargers])
		}
	}
}
