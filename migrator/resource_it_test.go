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
	resPathIn   string
	resPathOut  string
	resCfgIn    *config.CGRConfig
	resCfgOut   *config.CGRConfig
	resMigrator *Migrator
	resAction   string
)

var sTestsResIT = []func(t *testing.T){
	testResITConnect,
	testResITFlush,
	testResITMigrateAndMove,
}

func TestResourceITMove1(t *testing.T) {
	var err error
	resPathIn = path.Join(*dataDir, "conf", "samples", "tutmongo")
	resCfgIn, err = config.NewCGRConfigFromPath(resPathIn)
	if err != nil {
		t.Fatal(err)
	}
	resPathOut = path.Join(*dataDir, "conf", "samples", "tutmysql")
	resCfgOut, err = config.NewCGRConfigFromPath(resPathOut)
	if err != nil {
		t.Fatal(err)
	}
	resAction = utils.Move
	for _, stest := range sTestsResIT {
		t.Run("TestResourceITMove", stest)
	}
	resMigrator.Close()
}

func TestResourceITMove2(t *testing.T) {
	var err error
	resPathIn = path.Join(*dataDir, "conf", "samples", "tutmysql")
	resCfgIn, err = config.NewCGRConfigFromPath(resPathIn)
	if err != nil {
		t.Fatal(err)
	}
	resPathOut = path.Join(*dataDir, "conf", "samples", "tutmongo")
	resCfgOut, err = config.NewCGRConfigFromPath(resPathOut)
	if err != nil {
		t.Fatal(err)
	}
	resAction = utils.Move
	for _, stest := range sTestsResIT {
		t.Run("TestResourceITMove", stest)
	}
	resMigrator.Close()
}

func TestResourceITMoveEncoding(t *testing.T) {
	var err error
	resPathIn = path.Join(*dataDir, "conf", "samples", "tutmongo")
	resCfgIn, err = config.NewCGRConfigFromPath(resPathIn)
	if err != nil {
		t.Fatal(err)
	}
	resPathOut = path.Join(*dataDir, "conf", "samples", "tutmongojson")
	resCfgOut, err = config.NewCGRConfigFromPath(resPathOut)
	if err != nil {
		t.Fatal(err)
	}
	resAction = utils.Move
	for _, stest := range sTestsResIT {
		t.Run("TestResourceITMoveEncoding", stest)
	}
	resMigrator.Close()
}

func TestResourceITMoveEncoding2(t *testing.T) {
	var err error
	resPathIn = path.Join(*dataDir, "conf", "samples", "tutmysql")
	resCfgIn, err = config.NewCGRConfigFromPath(resPathIn)
	if err != nil {
		t.Fatal(err)
	}
	resPathOut = path.Join(*dataDir, "conf", "samples", "tutmysqljson")
	resCfgOut, err = config.NewCGRConfigFromPath(resPathOut)
	if err != nil {
		t.Fatal(err)
	}
	resAction = utils.Move
	for _, stest := range sTestsResIT {
		t.Run("TestResourceITMoveEncoding2", stest)
	}
	resMigrator.Close()
}

func testResITConnect(t *testing.T) {
	dataDBIn, err := NewMigratorDataDB(resCfgIn.DataDbCfg().Type,
		resCfgIn.DataDbCfg().Host, resCfgIn.DataDbCfg().Port,
		resCfgIn.DataDbCfg().Name, resCfgIn.DataDbCfg().User,
		resCfgIn.DataDbCfg().Password, resCfgIn.GeneralCfg().DBDataEncoding,
		config.CgrConfig().CacheCfg(), resCfgIn.DataDbCfg().Opts)
	if err != nil {
		log.Fatal(err)
	}
	dataDBOut, err := NewMigratorDataDB(resCfgOut.DataDbCfg().Type,
		resCfgOut.DataDbCfg().Host, resCfgOut.DataDbCfg().Port,
		resCfgOut.DataDbCfg().Name, resCfgOut.DataDbCfg().User,
		resCfgOut.DataDbCfg().Password, resCfgOut.GeneralCfg().DBDataEncoding,
		config.CgrConfig().CacheCfg(), resCfgOut.DataDbCfg().Opts)
	if err != nil {
		log.Fatal(err)
	}
	if reflect.DeepEqual(resPathIn, resPathOut) {
		resMigrator, err = NewMigrator(dataDBIn, dataDBOut, nil, nil,
			false, true, false, false)
	} else {
		resMigrator, err = NewMigrator(dataDBIn, dataDBOut, nil, nil,
			false, false, false, false)
	}
	if err != nil {
		log.Fatal(err)
	}
}

func testResITFlush(t *testing.T) {
	if err := resMigrator.dmOut.DataManager().DataDB().Flush(""); err != nil {
		t.Error(err)
	}
	if isEmpty, err := resMigrator.dmOut.DataManager().DataDB().IsDBEmpty(); err != nil {
		t.Error(err)
	} else if isEmpty != true {
		t.Errorf("Expecting: true got :%+v", isEmpty)
	}
	if err := engine.SetDBVersions(resMigrator.dmOut.DataManager().DataDB()); err != nil {
		t.Error("Error  ", err.Error())
	}
	if err := resMigrator.dmIN.DataManager().DataDB().Flush(""); err != nil {
		t.Error(err)
	}
	if isEmpty, err := resMigrator.dmIN.DataManager().DataDB().IsDBEmpty(); err != nil {
		t.Error(err)
	} else if isEmpty != true {
		t.Errorf("Expecting: true got :%+v", isEmpty)
	}
	if err := engine.SetDBVersions(resMigrator.dmIN.DataManager().DataDB()); err != nil {
		t.Error("Error  ", err.Error())
	}
}

func testResITMigrateAndMove(t *testing.T) {
	resPrfl := &engine.ResourceProfile{
		Tenant:       "cgrates.org",
		ID:           "RES1",
		FilterIDs:    []string{"*string:~Account:1001"},
		UsageTTL:     time.Second,
		Limit:        1,
		Weight:       10,
		ThresholdIDs: []string{"TH1"},
	}
	switch resAction {
	case utils.Migrate: // for the momment only one version of rating plans exists
	case utils.Move:
		if err := resMigrator.dmIN.DataManager().SetResourceProfile(resPrfl, true); err != nil {
			t.Error(err)
		}
		currentVersion := engine.CurrentDataDBVersions()
		err := resMigrator.dmIN.DataManager().DataDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for Resource ", err.Error())
		}

		_, err = resMigrator.dmOut.DataManager().GetResourceProfile("cgrates.org", "RES1", false, false, utils.NonTransactional)
		if err != utils.ErrNotFound {
			t.Error(err)
		}

		err, _ = resMigrator.Migrate([]string{utils.MetaResources})
		if err != nil {
			t.Error("Error when migrating Resource ", err.Error())
		}
		result, err := resMigrator.dmOut.DataManager().GetResourceProfile("cgrates.org", "RES1", false, false, utils.NonTransactional)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(result, resPrfl) {
			t.Errorf("Expecting: %+v, received: %+v", resPrfl, result)
		}
		result, err = resMigrator.dmIN.DataManager().GetResourceProfile("cgrates.org", "RES1", false, false, utils.NonTransactional)
		if err != utils.ErrNotFound {
			t.Error(err)
		} else if resMigrator.stats[utils.Resource] != 1 {
			t.Errorf("Expected 1, received: %v", resMigrator.stats[utils.Resource])
		}
	}
}
