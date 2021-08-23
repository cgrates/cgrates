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
	"path"
	"reflect"
	"testing"

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
	vrsPath = path.Join(*dataDir, "conf", "samples", "tutmysql")
	vrsCfg, err = config.NewCGRConfigFromPath(vrsPath)
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
	vrsPath = path.Join(*dataDir, "conf", "samples", "tutmongo")
	vrsCfg, err = config.NewCGRConfigFromPath(vrsPath)
	if err != nil {
		t.Fatal(err)
	}
	vrsCfg.StorDbCfg().Name = vrsCfg.DataDbCfg().DataDbName
	vrsSameOutDB = true
	for _, stest := range sTestsVrsIT {
		t.Run("TestVrsionITMigrateMongo", stest)
	}
	vrsMigrator.Close()
}

func testVrsITConnect(t *testing.T) {
	dataDBOut, err := NewMigratorDataDB(vrsCfg.DataDbCfg().DataDbType,
		vrsCfg.DataDbCfg().DataDbHost, vrsCfg.DataDbCfg().DataDbPort,
		vrsCfg.DataDbCfg().DataDbName, vrsCfg.DataDbCfg().DataDbUser,
		vrsCfg.DataDbCfg().DataDbPass, vrsCfg.GeneralCfg().DBDataEncoding,
		config.CgrConfig().CacheCfg(), "", vrsCfg.DataDbCfg().Items)
	if err != nil {
		t.Fatal(err)
	}

	storDBOut, err := NewMigratorStorDB(vrsCfg.StorDbCfg().Type,
		vrsCfg.StorDbCfg().Host, vrsCfg.StorDbCfg().Port,
		vrsCfg.StorDbCfg().Name, vrsCfg.StorDbCfg().User,
		vrsCfg.StorDbCfg().Password, vrsCfg.GeneralCfg().DBDataEncoding, vrsCfg.StorDbCfg().SSLMode,
		vrsCfg.StorDbCfg().MaxOpenConns, vrsCfg.StorDbCfg().MaxIdleConns,
		vrsCfg.StorDbCfg().ConnMaxLifetime, vrsCfg.StorDbCfg().StringIndexedFields,
		vrsCfg.StorDbCfg().PrefixIndexedFields, vrsCfg.StorDbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	vrsMigrator, err = NewMigrator(nil, dataDBOut, nil, storDBOut,
		false, false, false, vrsSameOutDB)
	if err != nil {
		t.Fatal(err)
	}
}

func testVrsITFlush(t *testing.T) {
	vrsMigrator.dmOut.DataManager().DataDB().Flush("")
	vrsMigrator.storDBOut.StorDB().Flush((path.Join(vrsCfg.DataFolderPath, "storage",
		vrsCfg.StorDbCfg().Type)))
	if vrs, err := vrsMigrator.dmOut.DataManager().DataDB().GetVersions(""); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected err=%s recived err=%v and rply=%s", utils.ErrNotFound.Error(), err, utils.ToJSON(vrs))
	}
	if vrs, err := vrsMigrator.storDBOut.StorDB().GetVersions(""); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected err=%s recived err=%v and rply=%s", utils.ErrNotFound.Error(), err, utils.ToJSON(vrs))
	}
}

func testVrsITMigrate(t *testing.T) {
	//check if version was set correctly
	// var emptyVers engine.Versions

	vrsMigrator.Migrate([]string{utils.MetaSetVersions})
	if vrsSameOutDB {
		expVrs := engine.CurrentAllDBVersions()
		if vrs, err := vrsMigrator.dmOut.DataManager().DataDB().GetVersions(""); err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual(expVrs, vrs) {
			t.Errorf("Expected %s recived %s", utils.ToJSON(expVrs), utils.ToJSON(vrs))
		}
	} else {
		expVrs := engine.CurrentDataDBVersions()
		if vrs, err := vrsMigrator.dmOut.DataManager().DataDB().GetVersions(""); err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual(expVrs, vrs) {
			t.Errorf("Expected %s recived %s", utils.ToJSON(expVrs), utils.ToJSON(vrs))
		}

		expVrs = engine.CurrentStorDBVersions()
		if vrs, err := vrsMigrator.storDBOut.StorDB().GetVersions(""); err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual(expVrs, vrs) {
			t.Errorf("Expected %s recived %s", utils.ToJSON(expVrs), utils.ToJSON(vrs))
		}
	}

	currentVersion := engine.Versions{Alias: 0}
	err := vrsMigrator.dmOut.DataManager().DataDB().SetVersions(currentVersion, false)
	if err != nil {
		t.Error("Error when setting version ", err.Error())
	}
	err = vrsMigrator.storDBOut.StorDB().SetVersions(currentVersion, false)
	if err != nil {
		t.Error("Error when setting version ", err.Error())
	}

	vrsMigrator.Migrate([]string{utils.MetaSetVersions})
	if vrsSameOutDB {
		expVrs := engine.CurrentAllDBVersions()
		if vrs, err := vrsMigrator.dmOut.DataManager().DataDB().GetVersions(""); err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual(expVrs, vrs) {
			t.Errorf("Expected %s recived %s", utils.ToJSON(expVrs), utils.ToJSON(vrs))
		}
	} else {
		expVrs := engine.CurrentDataDBVersions()
		if vrs, err := vrsMigrator.dmOut.DataManager().DataDB().GetVersions(""); err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual(expVrs, vrs) {
			t.Errorf("Expected %s recived %s", utils.ToJSON(expVrs), utils.ToJSON(vrs))
		}

		expVrs = engine.CurrentStorDBVersions()
		if vrs, err := vrsMigrator.storDBOut.StorDB().GetVersions(""); err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual(expVrs, vrs) {
			t.Errorf("Expected %s recived %s", utils.ToJSON(expVrs), utils.ToJSON(vrs))
		}
	}
}
