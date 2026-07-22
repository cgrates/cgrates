//go:build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package migrator

import (
	"path"
	"reflect"
	"testing"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	vrsPath      string
	vrsSameOutDB bool
	vrsCfg       *config.CGRConfig
	vrsMigrator  *Migrator
)

var sTestsVrsIT = []func(t *testing.T){
	testVrsITConnect,
	testVrsITFlush,
	testVrsITMigrate,
}

func TestVersionITRedis(t *testing.T) {
	var err error
	vrsPath = path.Join(*utils.DataDir, "conf", "samples", "tutredis")
	vrsCfg, err = config.NewCGRConfigFromPath(context.Background(), vrsPath)
	if err != nil {
		t.Fatal(err)
	}
	vrsSameOutDB = false
	for _, stest := range sTestsVrsIT {
		t.Run("TestVrsionITMigrateRedis", stest)
	}
	vrsMigrator.Close()
}

func TestVersionITMongo(t *testing.T) {
	var err error
	vrsPath = path.Join(*utils.DataDir, "conf", "samples", "tutmongo")
	vrsCfg, err = config.NewCGRConfigFromPath(context.Background(), vrsPath)
	if err != nil {
		t.Fatal(err)
	}
	vrsSameOutDB = true
	for _, stest := range sTestsVrsIT {
		t.Run("TestVrsionITMigrateMongo", stest)
	}
	vrsMigrator.Close()
}

func testVrsITConnect(t *testing.T) {
	locker := engine.NewGuardianLocker(config.NewDefaultCGRConfig())
	cacheS := engine.NewCacheS(vrsCfg, nil, nil, nil, locker)
	dataDBOut, err := NewMigratorDataDBs([]string{utils.MetaDefault}, vrsCfg.GeneralCfg().DBDataEncoding, vrsCfg, cacheS, locker)
	if err != nil {
		t.Fatal(err)
	}
	vrsMigrator = NewMigrator(nil, dataDBOut, false, false)
}

func testVrsITFlush(t *testing.T) {
	vrsMigrator.dmTo.DB()[utils.MetaDefault].Flush("")
	if vrs, err := vrsMigrator.dmTo.DB()[utils.MetaDefault].GetVersions(""); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected err=%s received err=%v and rply=%s", utils.ErrNotFound.Error(), err, utils.ToJSON(vrs))
	}
}

func testVrsITMigrate(t *testing.T) {
	//check if version was set correctly
	// var emptyVers engine.Versions

	vrsMigrator.Migrate([]string{utils.MetaSetVersions})
	if vrsSameOutDB {
		expVrs := engine.CurrentAllDBVersions()
		if vrs, err := vrsMigrator.dmTo.DB()[utils.MetaDefault].GetVersions(""); err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual(expVrs, vrs) {
			t.Errorf("Expected %s received %s", utils.ToJSON(expVrs), utils.ToJSON(vrs))
		}
	} else {
		expVrs := engine.CurrentDataDBVersions()
		if vrs, err := vrsMigrator.dmTo.DB()[utils.MetaDefault].GetVersions(""); err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual(expVrs, vrs) {
			t.Errorf("Expected %s received %s", utils.ToJSON(expVrs), utils.ToJSON(vrs))
		}
	}

	currentVersion := engine.Versions{utils.Attributes: 0}
	err := vrsMigrator.dmTo.DB()[utils.MetaDefault].SetVersions(currentVersion, false)
	if err != nil {
		t.Error("Error when setting version ", err.Error())
	}

	vrsMigrator.Migrate([]string{utils.MetaSetVersions})
	if vrsSameOutDB {
		expVrs := engine.CurrentAllDBVersions()
		if vrs, err := vrsMigrator.dmTo.DB()[utils.MetaDefault].GetVersions(""); err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual(expVrs, vrs) {
			t.Errorf("Expected %s received %s", utils.ToJSON(expVrs), utils.ToJSON(vrs))
		}
	} else {
		expVrs := engine.CurrentDataDBVersions()
		if vrs, err := vrsMigrator.dmTo.DB()[utils.MetaDefault].GetVersions(""); err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual(expVrs, vrs) {
			t.Errorf("Expected %s received %s", utils.ToJSON(expVrs), utils.ToJSON(vrs))
		}

	}
}
