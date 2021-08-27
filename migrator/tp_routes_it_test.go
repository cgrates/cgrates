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
	"sort"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	tpSplPathIn   string
	tpSplPathOut  string
	tpSplCfgIn    *config.CGRConfig
	tpSplCfgOut   *config.CGRConfig
	tpSplMigrator *Migrator
	tpSuppliers   []*utils.TPRouteProfile
)

var sTestsTpSplIT = []func(t *testing.T){
	testTpSplITConnect,
	testTpSplITFlush,
	testTpSplITPopulate,
	testTpSplITMove,
	testTpSplITCheckData,
}

func TestTpSplMove(t *testing.T) {
	for _, stest := range sTestsTpSplIT {
		t.Run("TestTpSplMove", stest)
	}
	tpSplMigrator.Close()
}

func testTpSplITConnect(t *testing.T) {
	var err error
	tpSplPathIn = path.Join(*dataDir, "conf", "samples", "tutmongo")
	tpSplCfgIn, err = config.NewCGRConfigFromPath(tpSplPathIn)
	if err != nil {
		t.Fatal(err)
	}
	tpSplPathOut = path.Join(*dataDir, "conf", "samples", "tutmysql")
	tpSplCfgOut, err = config.NewCGRConfigFromPath(tpSplPathOut)
	if err != nil {
		t.Fatal(err)
	}
	storDBIn, err := NewMigratorStorDB(tpSplCfgIn.StorDbCfg().Type,
		tpSplCfgIn.StorDbCfg().Host, tpSplCfgIn.StorDbCfg().Port,
		tpSplCfgIn.StorDbCfg().Name, tpSplCfgIn.StorDbCfg().User,
		tpSplCfgIn.StorDbCfg().Password, tpSplCfgIn.GeneralCfg().DBDataEncoding,
		tpSplCfgIn.StorDbCfg().StringIndexedFields, tpSplCfgIn.StorDbCfg().PrefixIndexedFields,
		tpSplCfgIn.StorDbCfg().Opts)
	if err != nil {
		log.Fatal(err)
	}
	storDBOut, err := NewMigratorStorDB(tpSplCfgOut.StorDbCfg().Type,
		tpSplCfgOut.StorDbCfg().Host, tpSplCfgOut.StorDbCfg().Port,
		tpSplCfgOut.StorDbCfg().Name, tpSplCfgOut.StorDbCfg().User,
		tpSplCfgOut.StorDbCfg().Password, tpSplCfgOut.GeneralCfg().DBDataEncoding,
		tpSplCfgIn.StorDbCfg().StringIndexedFields, tpSplCfgIn.StorDbCfg().PrefixIndexedFields,
		tpSplCfgOut.StorDbCfg().Opts)
	if err != nil {
		log.Fatal(err)
	}
	tpSplMigrator, err = NewMigrator(nil, nil, storDBIn, storDBOut, false, false, false, false)
	if err != nil {
		log.Fatal(err)
	}
}

func testTpSplITFlush(t *testing.T) {
	if err := tpSplMigrator.storDBIn.StorDB().Flush(
		path.Join(tpSplCfgIn.DataFolderPath, "storage", tpSplCfgIn.StorDbCfg().Type)); err != nil {
		t.Error(err)
	}

	if err := tpSplMigrator.storDBOut.StorDB().Flush(
		path.Join(tpSplCfgOut.DataFolderPath, "storage", tpSplCfgOut.StorDbCfg().Type)); err != nil {
		t.Error(err)
	}
}

func testTpSplITPopulate(t *testing.T) {
	tpSuppliers = []*utils.TPRouteProfile{
		{
			TPid:              "TP1",
			Tenant:            "cgrates.org",
			ID:                "SUPL_1",
			FilterIDs:         []string{"FLTR_ACNT_dan", "FLTR_DST_DE"},
			Sorting:           "*lowest_cost",
			SortingParameters: []string{},
			Routes: []*utils.TPRoute{
				{
					ID:              "supplier1",
					FilterIDs:       []string{"FLTR_1"},
					AccountIDs:      []string{"Acc1", "Acc2"},
					RateProfileIDs:  []string{"RPL_1"},
					ResourceIDs:     []string{"ResGroup1"},
					StatIDs:         []string{"Stat1"},
					Weight:          10,
					Blocker:         false,
					RouteParameters: "SortingParam1",
				},
			},
			Weight: 20,
		},
	}
	if err := tpSplMigrator.storDBIn.StorDB().SetTPRoutes(tpSuppliers); err != nil {
		t.Error("Error when setting TpSuppliers ", err.Error())
	}
	currentVersion := engine.CurrentStorDBVersions()
	err := tpSplMigrator.storDBIn.StorDB().SetVersions(currentVersion, false)
	if err != nil {
		t.Error("Error when setting version for TpSuppliers ", err.Error())
	}
}

func testTpSplITMove(t *testing.T) {
	err, _ := tpSplMigrator.Migrate([]string{utils.MetaTpRoutes})
	if err != nil {
		t.Error("Error when migrating TpSuppliers ", err.Error())
	}
}

func testTpSplITCheckData(t *testing.T) {
	result, err := tpSplMigrator.storDBOut.StorDB().GetTPRoutes(
		tpSuppliers[0].TPid, "", tpSuppliers[0].ID)
	if err != nil {
		t.Error("Error when getting TpSuppliers ", err.Error())
	}
	tpSuppliers[0].SortingParameters = nil // because of converting and empty string into a slice
	sort.Strings(result[0].FilterIDs)
	sort.Strings(tpSuppliers[0].FilterIDs)
	if !reflect.DeepEqual(tpSuppliers[0], result[0]) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(tpSuppliers[0]), utils.ToJSON(result[0]))
	}
	result, err = tpSplMigrator.storDBIn.StorDB().GetTPRoutes(
		tpSuppliers[0].TPid, "", tpSuppliers[0].ID)
	if err != utils.ErrNotFound {
		t.Error(err)
	}
}
