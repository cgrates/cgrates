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
	dataDBIn, err := NewMigratorDataDB(supCfgIn.DataDbCfg().DataDbType,
		supCfgIn.DataDbCfg().DataDbHost, supCfgIn.DataDbCfg().DataDbPort,
		supCfgIn.DataDbCfg().DataDbName, supCfgIn.DataDbCfg().DataDbUser,
		supCfgIn.DataDbCfg().DataDbPass, supCfgIn.GeneralCfg().DBDataEncoding,
		config.CgrConfig().CacheCfg(), "", supCfgIn.DataDbCfg().Items)
	if err != nil {
		log.Fatal(err)
	}
	dataDBOut, err := NewMigratorDataDB(supCfgOut.DataDbCfg().DataDbType,
		supCfgOut.DataDbCfg().DataDbHost, supCfgOut.DataDbCfg().DataDbPort,
		supCfgOut.DataDbCfg().DataDbName, supCfgOut.DataDbCfg().DataDbUser,
		supCfgOut.DataDbCfg().DataDbPass, supCfgOut.GeneralCfg().DBDataEncoding,
		config.CgrConfig().CacheCfg(), "", supCfgOut.DataDbCfg().Items)
	if err != nil {
		log.Fatal(err)
	}
	supMigrator, err = NewMigrator(dataDBIn, dataDBOut,
		nil, nil,
		false, false, false, false)
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
		t.Errorf("\nExpecting: true got :%+v", isEmpty)
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
		t.Errorf("\nExpecting: true got :%+v", isEmpty)
	}
	if err := engine.SetDBVersions(supMigrator.dmIN.DataManager().DataDB()); err != nil {
		t.Error("Error  ", err.Error())
	}
}

func testSupITMigrateAndMove(t *testing.T) {
	supPrfl := &engine.SupplierProfile{
		Tenant:            "cgrates.org",
		ID:                "SUP1",
		FilterIDs:         []string{"*string:~Account:1001"},
		Weight:            10,
		Sorting:           utils.MetaQOS,
		SortingParameters: []string{},
		Suppliers: []*engine.Supplier{
			&engine.Supplier{
				ID:            "Sup",
				FilterIDs:     []string{},
				AccountIDs:    []string{"1001"},
				RatingPlanIDs: []string{"RT_PLAN1"},
				ResourceIDs:   []string{"RES1"},
				Weight:        10,
			},
		},
	}
	switch supAction {
	case utils.Migrate: // for the momment only one version of rating plans exists
	case utils.Move:
		if err := supMigrator.dmIN.DataManager().SetSupplierProfile(supPrfl, true); err != nil {
			t.Error(err)
		}
		if _, err := supMigrator.dmIN.DataManager().GetSupplierProfile("cgrates.org", "SUP1", false, false, utils.NonTransactional); err != nil {
			t.Error(err)
		}
		currentVersion := engine.CurrentDataDBVersions()
		err := supMigrator.dmOut.DataManager().DataDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for Suppliers ", err.Error())
		}

		_, err = supMigrator.dmOut.DataManager().GetSupplierProfile("cgrates.org", "SUP1", false, false, utils.NonTransactional)
		if err != utils.ErrNotFound {
			t.Error(err)
		}

		err, _ = supMigrator.Migrate([]string{utils.MetaSuppliers})
		if err != nil {
			t.Error("Error when migrating Suppliers ", err.Error())
		}
		result, err := supMigrator.dmOut.DataManager().GetSupplierProfile("cgrates.org", "SUP1", false, false, utils.NonTransactional)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(result, supPrfl) {
			t.Errorf("Expecting: %+v, received: %+v", supPrfl, result)
		}
		result, err = supMigrator.dmIN.DataManager().GetSupplierProfile("cgrates.org", "SUP1", false, false, utils.NonTransactional)
		if err != utils.ErrNotFound {
			t.Error(err)
		}
	}
}
