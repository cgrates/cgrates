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
	dataDBOut, err := NewMigratorDataDBs([]string{utils.MetaDefault}, vrsCfg.GeneralCfg().DBDataEncoding, vrsCfg)
	if err != nil {
		t.Fatal(err)
	}
	vrsMigrator, err = NewMigrator(vrsCfg.DbCfg(), nil, dataDBOut,
		false, false)
	if err != nil {
		t.Fatal(err)
	}
}

func testVrsITFlush(t *testing.T) {
	vrsMigrator.dmTo[utils.MetaDefault].DataManager().DataDB()[utils.MetaDefault].Flush("")
	if vrs, err := vrsMigrator.dmTo[utils.MetaDefault].DataManager().DataDB()[utils.MetaDefault].GetVersions(""); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected err=%s received err=%v and rply=%s", utils.ErrNotFound.Error(), err, utils.ToJSON(vrs))
	}
}

func testVrsITMigrate(t *testing.T) {
	//check if version was set correctly
	// var emptyVers engine.Versions

	vrsMigrator.Migrate([]string{utils.MetaSetVersions})
	if vrsSameOutDB {
		expVrs := engine.CurrentAllDBVersions()
		if vrs, err := vrsMigrator.dmTo[utils.MetaDefault].DataManager().DataDB()[utils.MetaDefault].GetVersions(""); err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual(expVrs, vrs) {
			t.Errorf("Expected %s received %s", utils.ToJSON(expVrs), utils.ToJSON(vrs))
		}
	} else {
		expVrs := engine.CurrentDataDBVersions()
		if vrs, err := vrsMigrator.dmTo[utils.MetaDefault].DataManager().DataDB()[utils.MetaDefault].GetVersions(""); err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual(expVrs, vrs) {
			t.Errorf("Expected %s received %s", utils.ToJSON(expVrs), utils.ToJSON(vrs))
		}
	}

	currentVersion := engine.Versions{utils.Attributes: 0}
	err := vrsMigrator.dmTo[utils.MetaDefault].DataManager().DataDB()[utils.MetaDefault].SetVersions(currentVersion, false)
	if err != nil {
		t.Error("Error when setting version ", err.Error())
	}

	vrsMigrator.Migrate([]string{utils.MetaSetVersions})
	if vrsSameOutDB {
		expVrs := engine.CurrentAllDBVersions()
		if vrs, err := vrsMigrator.dmTo[utils.MetaDefault].DataManager().DataDB()[utils.MetaDefault].GetVersions(""); err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual(expVrs, vrs) {
			t.Errorf("Expected %s received %s", utils.ToJSON(expVrs), utils.ToJSON(vrs))
		}
	} else {
		expVrs := engine.CurrentDataDBVersions()
		if vrs, err := vrsMigrator.dmTo[utils.MetaDefault].DataManager().DataDB()[utils.MetaDefault].GetVersions(""); err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual(expVrs, vrs) {
			t.Errorf("Expected %s received %s", utils.ToJSON(expVrs), utils.ToJSON(vrs))
		}

	}
}
