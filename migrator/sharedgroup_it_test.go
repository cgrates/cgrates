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
	shrGrpPathIn   string
	shrGrpPathOut  string
	shrGrpCfgIn    *config.CGRConfig
	shrGrpCfgOut   *config.CGRConfig
	shrGrpMigrator *Migrator
	shrSharedGroup string
)

var sTestsShrGrpIT = []func(t *testing.T){
	testShrGrpITConnect,
	testShrGrpITFlush,
	testShrGrpITMigrateAndMove,
}

func TestSharedGroupITRedis(t *testing.T) {
	var err error
	shrGrpPathIn = path.Join(*dataDir, "conf", "samples", "tutmysql")
	shrGrpCfgIn, err = config.NewCGRConfigFromFolder(shrGrpPathIn)
	if err != nil {
		t.Fatal(err)
	}
	shrGrpCfgOut, err = config.NewCGRConfigFromFolder(shrGrpPathIn)
	if err != nil {
		t.Fatal(err)
	}
	shrSharedGroup = utils.Migrate
	for _, stest := range sTestsShrGrpIT {
		t.Run("TestSharedGroupITMigrateRedis", stest)
	}
}

func TestSharedGroupITMongo(t *testing.T) {
	var err error
	shrGrpPathIn = path.Join(*dataDir, "conf", "samples", "tutmongo")
	shrGrpCfgIn, err = config.NewCGRConfigFromFolder(shrGrpPathIn)
	if err != nil {
		t.Fatal(err)
	}
	shrGrpCfgOut, err = config.NewCGRConfigFromFolder(shrGrpPathIn)
	if err != nil {
		t.Fatal(err)
	}
	shrSharedGroup = utils.Migrate
	for _, stest := range sTestsShrGrpIT {
		t.Run("TestSharedGroupITMigrateMongo", stest)
	}
}

func TestSharedGroupITMove(t *testing.T) {
	var err error
	shrGrpPathIn = path.Join(*dataDir, "conf", "samples", "tutmongo")
	shrGrpCfgIn, err = config.NewCGRConfigFromFolder(shrGrpPathIn)
	if err != nil {
		t.Fatal(err)
	}
	shrGrpPathOut = path.Join(*dataDir, "conf", "samples", "tutmysql")
	shrGrpCfgOut, err = config.NewCGRConfigFromFolder(shrGrpPathOut)
	if err != nil {
		t.Fatal(err)
	}
	shrSharedGroup = utils.Move
	for _, stest := range sTestsShrGrpIT {
		t.Run("TestSharedGroupITMove", stest)
	}
}

func TestSharedGroupITMoveEncoding(t *testing.T) {
	var err error
	shrGrpPathIn = path.Join(*dataDir, "conf", "samples", "tutmongo")
	shrGrpCfgIn, err = config.NewCGRConfigFromFolder(shrGrpPathIn)
	if err != nil {
		t.Fatal(err)
	}
	shrGrpPathOut = path.Join(*dataDir, "conf", "samples", "tutmongojson")
	shrGrpCfgOut, err = config.NewCGRConfigFromFolder(shrGrpPathOut)
	if err != nil {
		t.Fatal(err)
	}
	shrSharedGroup = utils.Move
	for _, stest := range sTestsShrGrpIT {
		t.Run("TestSharedGroupITMoveEncoding", stest)
	}
}

func TestSharedGroupITMoveEncoding2(t *testing.T) {
	var err error
	shrGrpPathIn = path.Join(*dataDir, "conf", "samples", "tutmysql")
	shrGrpCfgIn, err = config.NewCGRConfigFromFolder(shrGrpPathIn)
	if err != nil {
		t.Fatal(err)
	}
	shrGrpPathOut = path.Join(*dataDir, "conf", "samples", "tutmysqljson")
	shrGrpCfgOut, err = config.NewCGRConfigFromFolder(shrGrpPathOut)
	if err != nil {
		t.Fatal(err)
	}
	shrSharedGroup = utils.Move
	for _, stest := range sTestsShrGrpIT {
		t.Run("TestSharedGroupITMoveEncoding2", stest)
	}
}

func testShrGrpITConnect(t *testing.T) {
	dataDBIn, err := NewMigratorDataDB(shrGrpCfgIn.DataDbCfg().DataDbType,
		shrGrpCfgIn.DataDbCfg().DataDbHost, shrGrpCfgIn.DataDbCfg().DataDbPort,
		shrGrpCfgIn.DataDbCfg().DataDbName, shrGrpCfgIn.DataDbCfg().DataDbUser,
		shrGrpCfgIn.DataDbCfg().DataDbPass, shrGrpCfgIn.GeneralCfg().DBDataEncoding,
		config.CgrConfig().CacheCfg(), "")
	if err != nil {
		log.Fatal(err)
	}
	dataDBOut, err := NewMigratorDataDB(shrGrpCfgOut.DataDbCfg().DataDbType,
		shrGrpCfgOut.DataDbCfg().DataDbHost, shrGrpCfgOut.DataDbCfg().DataDbPort,
		shrGrpCfgOut.DataDbCfg().DataDbName, shrGrpCfgOut.DataDbCfg().DataDbUser,
		shrGrpCfgOut.DataDbCfg().DataDbPass, shrGrpCfgOut.GeneralCfg().DBDataEncoding,
		config.CgrConfig().CacheCfg(), "")
	if err != nil {
		log.Fatal(err)
	}
	shrGrpMigrator, err = NewMigrator(dataDBIn, dataDBOut,
		nil, nil,
		false, false, false)
	if err != nil {
		log.Fatal(err)
	}
}

func testShrGrpITFlush(t *testing.T) {
	shrGrpMigrator.dmOut.DataManager().DataDB().Flush("")
	if err := engine.SetDBVersions(shrGrpMigrator.dmOut.DataManager().DataDB()); err != nil {
		t.Error("Error  ", err.Error())
	}
}

func testShrGrpITMigrateAndMove(t *testing.T) {
	v1shrGrp := &v1SharedGroup{
		Id: "Test",
		AccountParameters: map[string]*engine.SharingParameters{
			"test": {Strategy: "*highest"},
		},
		MemberIds: []string{"1", "2", "3"},
	}
	shrGrp := &engine.SharedGroup{
		Id: "Test",
		AccountParameters: map[string]*engine.SharingParameters{
			"test": {Strategy: "*highest"},
		},
		MemberIds: utils.NewStringMap("1", "2", "3"),
	}

	switch shrSharedGroup {
	case utils.Migrate:
		err := shrGrpMigrator.dmIN.setV1SharedGroup(v1shrGrp)
		if err != nil {
			t.Error("Error when setting v1 SharedGroup ", err.Error())
		}
		currentVersion := engine.Versions{utils.StatS: 2, utils.Thresholds: 2, utils.Accounts: 2, utils.Actions: 2, utils.ActionTriggers: 2, utils.ActionPlans: 2, utils.SharedGroups: 1}
		err = shrGrpMigrator.dmOut.DataManager().DataDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for SharedGroup ", err.Error())
		}
		err, _ = shrGrpMigrator.Migrate([]string{utils.MetaSharedGroups})
		if err != nil {
			t.Error("Error when migrating SharedGroup ", err.Error())
		}
		result, err := shrGrpMigrator.dmOut.DataManager().GetSharedGroup(v1shrGrp.Id, true, utils.NonTransactional)
		if err != nil {
			t.Error("Error when getting SharedGroup ", err.Error())
		}
		if !reflect.DeepEqual(shrGrp, result) {
			t.Errorf("Expecting: %+v, received: %+v", shrGrp, result)
		}
	case utils.Move:
		if err := shrGrpMigrator.dmIN.DataManager().SetSharedGroup(shrGrp, utils.NonTransactional); err != nil {
			t.Error("Error when setting SharedGroup ", err.Error())
		}
		currentVersion := engine.CurrentDataDBVersions()
		err := shrGrpMigrator.dmOut.DataManager().DataDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for SharedGroup ", err.Error())
		}
		err, _ = shrGrpMigrator.Migrate([]string{utils.MetaSharedGroups})
		if err != nil {
			t.Error("Error when migrating SharedGroup ", err.Error())
		}
		result, err := shrGrpMigrator.dmOut.DataManager().GetSharedGroup(v1shrGrp.Id, true, utils.NonTransactional)
		if err != nil {
			t.Error("Error when getting SharedGroup ", err.Error())
		}
		if !reflect.DeepEqual(shrGrp, result) {
			t.Errorf("Expecting: %+v, received: %+v", shrGrp, result)
		}
	}
}
