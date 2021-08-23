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
	inPath = path.Join(*dataDir, "conf", "samples", "tutmysql")
	testLoadIdsStart("TestLoadIDsMigrateITRedis", t)
}

func TestLoadIDsMigrateITMongo(t *testing.T) {
	inPath = path.Join(*dataDir, "conf", "samples", "tutmongo")
	testLoadIdsStart("TestLoadIDsMigrateITMongo", t)
}

func TestLoadIDsITMigrateMongo2Redis(t *testing.T) {
	inPath = path.Join(*dataDir, "conf", "samples", "tutmongo")
	outPath = path.Join(*dataDir, "conf", "samples", "tutmysql")
	testLoadIdsStart("TestLoadIDsITMigrateMongo2Redis", t)
}

func testLoadIdsStart(testName string, t *testing.T) {
	var err error
	if loadCfgIn, err = config.NewCGRConfigFromPath(inPath); err != nil {
		t.Fatal(err)
	}
	config.SetCgrConfig(loadCfgIn)
	if loadCfgOut, err = config.NewCGRConfigFromPath(outPath); err != nil {
		t.Fatal(err)
	}
	for _, stest := range sTestsLoadIdsIT {
		t.Run(testName, stest)
	}
	loadMigrator.Close()
}

func testLoadIdsITConnect(t *testing.T) {
	dataDBIn, err := NewMigratorDataDB(loadCfgIn.DataDbCfg().Type,
		loadCfgIn.DataDbCfg().Host, loadCfgIn.DataDbCfg().Port,
		loadCfgIn.DataDbCfg().Name, loadCfgIn.DataDbCfg().User,
		loadCfgIn.DataDbCfg().Password, loadCfgIn.GeneralCfg().DBDataEncoding,
		config.CgrConfig().CacheCfg(), loadCfgIn.DataDbCfg().Opts)
	if err != nil {
		log.Fatal(err)
	}
	dataDBOut, err := NewMigratorDataDB(loadCfgOut.DataDbCfg().Type,
		loadCfgOut.DataDbCfg().Host, loadCfgOut.DataDbCfg().Port,
		loadCfgOut.DataDbCfg().Name, loadCfgOut.DataDbCfg().User,
		loadCfgOut.DataDbCfg().Password, loadCfgOut.GeneralCfg().DBDataEncoding,
		config.CgrConfig().CacheCfg(), loadCfgOut.DataDbCfg().Opts)
	if err != nil {
		log.Fatal(err)
	}
	if inPath == outPath {
		loadMigrator, err = NewMigrator(dataDBIn, dataDBOut,
			nil, nil, false, true, false, false)
	} else {
		loadMigrator, err = NewMigrator(dataDBIn, dataDBOut,
			nil, nil, false, false, false, false)
	}
	if err != nil {
		log.Fatal(err)
	}
}

func testLoadIdsITFlush(t *testing.T) {
	loadMigrator.dmOut.DataManager().DataDB().Flush("")
	if err := engine.SetDBVersions(loadMigrator.dmOut.DataManager().DataDB()); err != nil {
		t.Error("Error  ", err.Error())
	}
	loadMigrator.dmIN.DataManager().DataDB().Flush("")
	if err := engine.SetDBVersions(loadMigrator.dmIN.DataManager().DataDB()); err != nil {
		t.Error("Error  ", err.Error())
	}
}

func testLoadIdsITMigrateAndMove(t *testing.T) {

	err := loadMigrator.dmIN.DataManager().DataDB().SetLoadIDsDrv(context.TODO(), map[string]int64{"account": 1}) // this will be deleated
	if err != nil {
		t.Error("Error when setting new loadID ", err.Error())
	}
	currentVersion := engine.Versions{utils.LoadIDsVrs: 0}
	err = loadMigrator.dmIN.DataManager().DataDB().SetVersions(currentVersion, false)
	if err != nil {
		t.Error("Error when setting version for LoadIDs ", err.Error())
	}
	//check if version was set correctly
	if vrs, err := loadMigrator.dmIN.DataManager().DataDB().GetVersions(""); err != nil {
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
	if vrs, err := loadMigrator.dmOut.DataManager().DataDB().GetVersions(""); err != nil {
		t.Error(err)
	} else if vrs[utils.LoadIDsVrs] != 1 {
		t.Errorf("Unexpected version returned: %d", vrs[utils.LoadIDsVrs])
	}
	//check if user was migrate correctly
	_, err = loadMigrator.dmOut.DataManager().DataDB().GetItemLoadIDsDrv(context.TODO(), "")
	if err != utils.ErrNotFound {
		t.Error("Error should be not found : ", err)
	}
	// no need to modify the LoadIDs from dmIN
	// if _, err = loadMigrator.dmIN.DataManager().DataDB().GetItemLoadIDsDrv(context.TODO(),""); err != utils.ErrNotFound {
	// 	t.Error("Error should be not found : ", err)
	// }
}
