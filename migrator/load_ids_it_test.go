//go:build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package migrator

import (
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
	testLoadIdsITMove,
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
	if loadCfgOut, err = config.NewCGRConfigFromPath(context.Background(), outPath); err != nil {
		t.Fatal(err)
	}
	for _, stest := range sTestsLoadIdsIT {
		t.Run(testName, stest)
	}
	loadMigrator.Close()
}

func testLoadIdsITConnect(t *testing.T) {
	locker := engine.NewLocker(config.NewDefaultCGRConfig())
	cacheIn := engine.NewCacheS(loadCfgIn, nil, nil, nil, locker)
	dataDBIn, err := NewMigratorDataDBs([]string{utils.MetaDefault}, loadCfgIn.GeneralCfg().DBDataEncoding, loadCfgIn, cacheIn, locker)
	if err != nil {
		t.Fatal(err)
	}
	cacheOut := engine.NewCacheS(loadCfgOut, nil, nil, nil, locker)
	dataDBOut, err := NewMigratorDataDBs([]string{utils.MetaDefault}, loadCfgOut.GeneralCfg().DBDataEncoding, loadCfgOut, cacheOut, locker)
	if err != nil {
		t.Fatal(err)
	}
	loadMigrator = NewMigrator(dataDBIn, dataDBOut, false, inPath == outPath)
}

func testLoadIdsITFlush(t *testing.T) {
	loadMigrator.dmTo.DB()[utils.MetaDefault].Flush("")
	if err := engine.SetDBVersions(loadMigrator.dmTo.DB()[utils.MetaDefault]); err != nil {
		t.Error("Error  ", err.Error())
	}
	loadMigrator.dmFrom.DB()[utils.MetaDefault].Flush("")
	if err := engine.SetDBVersions(loadMigrator.dmFrom.DB()[utils.MetaDefault]); err != nil {
		t.Error("Error  ", err.Error())
	}
}

func testLoadIdsITMove(t *testing.T) {

	err := loadMigrator.dmFrom.DB()[utils.MetaDefault].SetLoadIDsDrv(context.TODO(), map[string]int64{"account": 1}) // this will be deleated
	if err != nil {
		t.Error("Error when setting new loadID ", err.Error())
	}
	currentVersion := engine.Versions{utils.LoadIDsVrs: 0}
	err = loadMigrator.dmFrom.DB()[utils.MetaDefault].SetVersions(currentVersion, false)
	if err != nil {
		t.Error("Error when setting version for LoadIDs ", err.Error())
	}
	//check if version was set correctly
	if vrs, err := loadMigrator.dmFrom.DB()[utils.MetaDefault].GetVersions(""); err != nil {
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
	if vrs, err := loadMigrator.dmTo.DB()[utils.MetaDefault].GetVersions(""); err != nil {
		t.Error(err)
	} else if vrs[utils.LoadIDsVrs] != 1 {
		t.Errorf("Unexpected version returned: %d", vrs[utils.LoadIDsVrs])
	}
	//check if user was migrate correctly
	_, err = loadMigrator.dmTo.DB()[utils.MetaDefault].GetItemLoadIDsDrv(context.TODO(), "")
	if err != utils.ErrNotFound {
		t.Error("Error should be not found : ", err)
	}
	// no need to modify the LoadIDs from dmFrom
	// if _, err = loadMigrator.dmFrom.DataManager().DataDB()[utils.MetaDefault].GetItemLoadIDsDrv(context.TODO(),""); err != utils.ErrNotFound {
	// 	t.Error("Error should be not found : ", err)
	// }
}
