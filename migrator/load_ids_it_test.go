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
	"testing"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	loadCfgIn    *config.CGRConfig
	loadCfgOut   *config.CGRConfig
	loadMigrator *Migrator
)

var sTestsLoadIdsIT = []func(t *testing.T){
	testLoadIdsITConnect,
	testLoadIdsITFlush,
	testLoadIdsITMigrateAndMove,
}

func TestLoadIDsMigrateITRedis(t *testing.T) {
	inPath = path.Join(*utils.DataDir, "conf", "samples", "tutredis")
	testLoadIdsStart("TestLoadIDsMigrateITRedis", t)
}

func TestLoadIDsMigrateITMongo(t *testing.T) {
	inPath = path.Join(*utils.DataDir, "conf", "samples", "tutmongo")
	testLoadIdsStart("TestLoadIDsMigrateITMongo", t)
}

func TestLoadIDsITMigrateMongo2Redis(t *testing.T) {
	inPath = path.Join(*utils.DataDir, "conf", "samples", "tutmongo")
	outPath = path.Join(*utils.DataDir, "conf", "samples", "tutredis")
	testLoadIdsStart("TestLoadIDsITMigrateMongo2Redis", t)
}

func testLoadIdsStart(testName string, t *testing.T) {
	var err error
	if loadCfgIn, err = config.NewCGRConfigFromPath(context.Background(), inPath); err != nil {
		t.Fatal(err)
	}
	config.SetCgrConfig(loadCfgIn)
	if loadCfgOut, err = config.NewCGRConfigFromPath(context.Background(), outPath); err != nil {
		t.Fatal(err)
	}
	for _, stest := range sTestsLoadIdsIT {
		t.Run(testName, stest)
	}
	loadMigrator.Close()
}

func testLoadIdsITConnect(t *testing.T) {
	dataDBIn, err := NewMigratorDataDBs([]string{utils.MetaDefault}, loadCfgIn.GeneralCfg().DBDataEncoding, loadCfgIn)
	if err != nil {
		log.Fatal(err)
	}
	dataDBOut, err := NewMigratorDataDBs([]string{utils.MetaDefault}, loadCfgOut.GeneralCfg().DBDataEncoding, loadCfgOut)
	if err != nil {
		log.Fatal(err)
	}
	if inPath == outPath {
		loadMigrator, err = NewMigrator(loadCfgOut.DbCfg(), dataDBIn, dataDBOut,
			false, true)
	} else {
		loadMigrator, err = NewMigrator(loadCfgOut.DbCfg(), dataDBIn, dataDBOut,
			false, false)
	}
	if err != nil {
		log.Fatal(err)
	}
}

func testLoadIdsITFlush(t *testing.T) {
	loadMigrator.dmTo[utils.MetaDefault].DataManager().DataDB()[utils.MetaDefault].Flush("")
	if err := engine.SetDBVersions(loadMigrator.dmTo[utils.MetaDefault].DataManager().DataDB()[utils.MetaDefault]); err != nil {
		t.Error("Error  ", err.Error())
	}
	loadMigrator.dmFrom[utils.MetaDefault].DataManager().DataDB()[utils.MetaDefault].Flush("")
	if err := engine.SetDBVersions(loadMigrator.dmFrom[utils.MetaDefault].DataManager().DataDB()[utils.MetaDefault]); err != nil {
		t.Error("Error  ", err.Error())
	}
}

func testLoadIdsITMigrateAndMove(t *testing.T) {

	err := loadMigrator.dmFrom[utils.MetaDefault].DataManager().DataDB()[utils.MetaDefault].SetLoadIDsDrv(context.TODO(), map[string]int64{"account": 1}) // this will be deleated
	if err != nil {
		t.Error("Error when setting new loadID ", err.Error())
	}
	currentVersion := engine.Versions{utils.LoadIDsVrs: 0}
	err = loadMigrator.dmFrom[utils.MetaDefault].DataManager().DataDB()[utils.MetaDefault].SetVersions(currentVersion, false)
	if err != nil {
		t.Error("Error when setting version for LoadIDs ", err.Error())
	}
	//check if version was set correctly
	if vrs, err := loadMigrator.dmFrom[utils.MetaDefault].DataManager().DataDB()[utils.MetaDefault].GetVersions(""); err != nil {
		t.Error(err)
	} else if vrs[utils.LoadIDsVrs] != 0 {
		t.Errorf("Unexpected version returned: %d", vrs[utils.LoadIDsVrs])
	}
	//migrate user
	err, _ = loadMigrator.Migrate([]string{utils.MetaLoadIDs})
	if err != nil {
		t.Error("Error when migrating LoadIDs ", err.Error())
	}
	//check if version was updated
	if vrs, err := loadMigrator.dmTo[utils.MetaDefault].DataManager().DataDB()[utils.MetaDefault].GetVersions(""); err != nil {
		t.Error(err)
	} else if vrs[utils.LoadIDsVrs] != 1 {
		t.Errorf("Unexpected version returned: %d", vrs[utils.LoadIDsVrs])
	}
	//check if user was migrate correctly
	_, err = loadMigrator.dmTo[utils.MetaDefault].DataManager().DataDB()[utils.MetaDefault].GetItemLoadIDsDrv(context.TODO(), "")
	if err != utils.ErrNotFound {
		t.Error("Error should be not found : ", err)
	}
	// no need to modify the LoadIDs from dmFrom
	// if _, err = loadMigrator.dmFrom[utils.MetaDefault].DataManager().DataDB()[utils.MetaDefault].GetItemLoadIDsDrv(context.TODO(),""); err != utils.ErrNotFound {
	// 	t.Error("Error should be not found : ", err)
	// }
}
