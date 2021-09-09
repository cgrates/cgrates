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
	"reflect"
	"testing"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	supPathIn   string
	supPathOut  string
	supCfgIn    *config.CGRConfig
	supCfgOut   *config.CGRConfig
	supMigrator *Migrator
	supAction   string
)

var sTestsSupIT = []func(t *testing.T){
	testSupITConnect,
	testSupITFlush,
	testSupITMigrateAndMove,
}

func TestSuppliersITMove1(t *testing.T) {
	var err error
	supPathIn = path.Join(*dataDir, "conf", "samples", "tutmongo")
	supCfgIn, err = config.NewCGRConfigFromPath(supPathIn)
	if err != nil {
		t.Fatal(err)
	}
	supPathOut = path.Join(*dataDir, "conf", "samples", "tutmysql")
	supCfgOut, err = config.NewCGRConfigFromPath(supPathOut)
	if err != nil {
		t.Fatal(err)
	}
	supAction = utils.Move
	for _, stest := range sTestsSupIT {
		t.Run("TestSuppliersITMove", stest)
	}
	supMigrator.Close()
}

func TestSuppliersITMove2(t *testing.T) {
	var err error
	supPathIn = path.Join(*dataDir, "conf", "samples", "tutmysql")
	supCfgIn, err = config.NewCGRConfigFromPath(supPathIn)
	if err != nil {
		t.Fatal(err)
	}
	supPathOut = path.Join(*dataDir, "conf", "samples", "tutmongo")
	supCfgOut, err = config.NewCGRConfigFromPath(supPathOut)
	if err != nil {
		t.Fatal(err)
	}
	supAction = utils.Move
	for _, stest := range sTestsSupIT {
		t.Run("TestSuppliersITMove", stest)
	}
	supMigrator.Close()
}

func TestSuppliersITMoveEncoding(t *testing.T) {
	var err error
	supPathIn = path.Join(*dataDir, "conf", "samples", "tutmongo")
	supCfgIn, err = config.NewCGRConfigFromPath(supPathIn)
	if err != nil {
		t.Fatal(err)
	}
	supPathOut = path.Join(*dataDir, "conf", "samples", "tutmongojson")
	supCfgOut, err = config.NewCGRConfigFromPath(supPathOut)
	if err != nil {
		t.Fatal(err)
	}
	supAction = utils.Move
	for _, stest := range sTestsSupIT {
		t.Run("TestSuppliersITMoveEncoding", stest)
	}
	supMigrator.Close()
}

func TestSuppliersITMoveEncoding2(t *testing.T) {
	var err error
	supPathIn = path.Join(*dataDir, "conf", "samples", "tutmysql")
	supCfgIn, err = config.NewCGRConfigFromPath(supPathIn)
	if err != nil {
		t.Fatal(err)
	}
	supPathOut = path.Join(*dataDir, "conf", "samples", "tutmysqljson")
	supCfgOut, err = config.NewCGRConfigFromPath(supPathOut)
	if err != nil {
		t.Fatal(err)
	}
	supAction = utils.Move
	for _, stest := range sTestsSupIT {
		t.Run("TestSuppliersITMoveEncoding2", stest)
	}
	supMigrator.Close()
}

func testSupITConnect(t *testing.T) {
	dataDBIn, err := NewMigratorDataDB(supCfgIn.DataDbCfg().Type,
		supCfgIn.DataDbCfg().Host, supCfgIn.DataDbCfg().Port,
		supCfgIn.DataDbCfg().Name, supCfgIn.DataDbCfg().User,
		supCfgIn.DataDbCfg().Password, supCfgIn.GeneralCfg().DBDataEncoding,
		config.CgrConfig().CacheCfg(), supCfgIn.DataDbCfg().Opts)
	if err != nil {
		log.Fatal(err)
	}
	dataDBOut, err := NewMigratorDataDB(supCfgOut.DataDbCfg().Type,
		supCfgOut.DataDbCfg().Host, supCfgOut.DataDbCfg().Port,
		supCfgOut.DataDbCfg().Name, supCfgOut.DataDbCfg().User,
		supCfgOut.DataDbCfg().Password, supCfgOut.GeneralCfg().DBDataEncoding,
		config.CgrConfig().CacheCfg(), supCfgOut.DataDbCfg().Opts)
	if err != nil {
		log.Fatal(err)
	}
	if reflect.DeepEqual(supPathIn, supPathOut) {
		supMigrator, err = NewMigrator(dataDBIn, dataDBOut, nil, nil,
			false, true, false, false)
	} else {
		supMigrator, err = NewMigrator(dataDBIn, dataDBOut, nil, nil,
			false, false, false, false)
	}
	if err != nil {
		log.Fatal(err)
	}
}

func testSupITFlush(t *testing.T) {
	if err := supMigrator.dmOut.DataManager().DataDB().Flush(""); err != nil {
		t.Error(err)
	}
	if isEmpty, err := supMigrator.dmOut.DataManager().DataDB().IsDBEmpty(); err != nil {
		t.Error(err)
	} else if isEmpty != true {
		t.Errorf("Expecting: true got :%+v", isEmpty)
	}
	if err := engine.SetDBVersions(supMigrator.dmOut.DataManager().DataDB()); err != nil {
		t.Error("Error  ", err.Error())
	}
	if err := supMigrator.dmIN.DataManager().DataDB().Flush(""); err != nil {
		t.Error(err)
	}
	if isEmpty, err := supMigrator.dmIN.DataManager().DataDB().IsDBEmpty(); err != nil {
		t.Error(err)
	} else if isEmpty != true {
		t.Errorf("Expecting: true got :%+v", isEmpty)
	}
	if err := engine.SetDBVersions(supMigrator.dmIN.DataManager().DataDB()); err != nil {
		t.Error("Error  ", err.Error())
	}
}

func testSupITMigrateAndMove(t *testing.T) {
	supPrfl := &engine.RouteProfile{
		Tenant:    "cgrates.org",
		ID:        "SUP1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		// Weights:           10,
		Sorting:           utils.MetaQOS,
		SortingParameters: []string{},
		Routes: []*engine.Route{{
			ID:             "Sup",
			FilterIDs:      []string{},
			AccountIDs:     []string{"1001"},
			RateProfileIDs: []string{"RT_PLAN1"},
			ResourceIDs:    []string{"RES1"},
			// Weights:        10,
		}},
	}
	switch supAction {
	case utils.Migrate: // for the moment only one version of rating plans exists
	case utils.Move:
		if err := supMigrator.dmIN.DataManager().SetRouteProfile(context.TODO(), supPrfl, true); err != nil {
			t.Error(err)
		}
		if _, err := supMigrator.dmIN.DataManager().GetRouteProfile(context.TODO(), "cgrates.org", "SUP1", false, false, utils.NonTransactional); err != nil {
			t.Error(err)
		}
		currentVersion := engine.CurrentDataDBVersions()
		err := supMigrator.dmIN.DataManager().DataDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for Suppliers ", err.Error())
		}

		_, err = supMigrator.dmOut.DataManager().GetRouteProfile(context.TODO(), "cgrates.org", "SUP1", false, false, utils.NonTransactional)
		if err != utils.ErrNotFound {
			t.Error(err)
		}

		err, _ = supMigrator.Migrate([]string{utils.MetaRoutes})
		if err != nil {
			t.Error("Error when migrating Suppliers ", err.Error())
		}
		result, err := supMigrator.dmOut.DataManager().GetRouteProfile(context.TODO(), "cgrates.org", "SUP1", false, false, utils.NonTransactional)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(result, supPrfl) {
			t.Errorf("Expecting: %+v, received: %+v", supPrfl, result)
		}
		result, err = supMigrator.dmIN.DataManager().GetRouteProfile(context.TODO(), "cgrates.org", "SUP1", false, false, utils.NonTransactional)
		if err != utils.ErrNotFound {
			t.Error(err)
		} else if supMigrator.stats[utils.Routes] != 1 {
			t.Errorf("Expected 1, received: %v", supMigrator.stats[utils.Routes])
		}
	}
}
