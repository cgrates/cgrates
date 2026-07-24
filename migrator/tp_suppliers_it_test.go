//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
	tpSuppliers   []*utils.TPSupplierProfile
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
	tpSplPathIn = path.Join(*utils.DataDir, "conf", "samples", "tutmongo")
	tpSplCfgIn, err = config.NewCGRConfigFromPath(tpSplPathIn)
	if err != nil {
		t.Fatal(err)
	}
	tpSplPathOut = path.Join(*utils.DataDir, "conf", "samples", "tutmysql")
	tpSplCfgOut, err = config.NewCGRConfigFromPath(tpSplPathOut)
	if err != nil {
		t.Fatal(err)
	}
	storDBIn, err := NewMigratorStorDB(tpSplCfgIn.StorDbCfg().Type,
		tpSplCfgIn.StorDbCfg().Host, tpSplCfgIn.StorDbCfg().Port,
		tpSplCfgIn.StorDbCfg().Name, tpSplCfgIn.StorDbCfg().User,
		tpSplCfgIn.StorDbCfg().Password, tpSplCfgIn.GeneralCfg().DBDataEncoding, tpSplCfgIn.StorDbCfg().SSLMode,
		tpSplCfgIn.StorDbCfg().MaxOpenConns, tpSplCfgIn.StorDbCfg().MaxIdleConns,
		tpSplCfgIn.StorDbCfg().ConnMaxLifetime, tpSplCfgIn.StorDbCfg().StringIndexedFields,
		tpSplCfgIn.StorDbCfg().PrefixIndexedFields, tpSplCfgIn.StorDbCfg().Items)
	if err != nil {
		log.Fatal(err)
	}
	storDBOut, err := NewMigratorStorDB(tpSplCfgOut.StorDbCfg().Type,
		tpSplCfgOut.StorDbCfg().Host, tpSplCfgOut.StorDbCfg().Port,
		tpSplCfgOut.StorDbCfg().Name, tpSplCfgOut.StorDbCfg().User,
		tpSplCfgOut.StorDbCfg().Password, tpSplCfgOut.GeneralCfg().DBDataEncoding, tpSplCfgIn.StorDbCfg().SSLMode,
		tpSplCfgIn.StorDbCfg().MaxOpenConns, tpSplCfgIn.StorDbCfg().MaxIdleConns,
		tpSplCfgIn.StorDbCfg().ConnMaxLifetime, tpSplCfgIn.StorDbCfg().StringIndexedFields,
		tpSplCfgIn.StorDbCfg().PrefixIndexedFields, tpSplCfgOut.StorDbCfg().Items)
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
		path.Join(tpSplCfgIn.DataFolderPath, "storage", dbPath(tpSplCfgIn.StorDbCfg().Type))); err != nil {
		t.Error(err)
	}

	if err := tpSplMigrator.storDBOut.StorDB().Flush(
		path.Join(tpSplCfgOut.DataFolderPath, "storage", dbPath(tpSplCfgOut.StorDbCfg().Type))); err != nil {
		t.Error(err)
	}
}

func testTpSplITPopulate(t *testing.T) {
	tpSuppliers = []*utils.TPSupplierProfile{
		{
			TPid:      "TP1",
			Tenant:    "cgrates.org",
			ID:        "SUPL_1",
			FilterIDs: []string{"FLTR_ACNT_dan", "FLTR_DST_DE"},
			ActivationInterval: &utils.TPActivationInterval{
				ActivationTime: "2014-07-29T15:00:00Z",
				ExpiryTime:     "",
			},
			Sorting:           "*lowest_cost",
			SortingParameters: []string{},
			Suppliers: []*utils.TPSupplier{
				{
					ID:                 "supplier1",
					FilterIDs:          []string{"FLTR_1"},
					AccountIDs:         []string{"Acc1", "Acc2"},
					RatingPlanIDs:      []string{"RPL_1"},
					ResourceIDs:        []string{"ResGroup1"},
					StatIDs:            []string{"Stat1"},
					Weight:             10,
					Blocker:            false,
					SupplierParameters: "SortingParam1",
				},
			},
			Weight: 20,
		},
	}
	if err := tpSplMigrator.storDBIn.StorDB().SetTPSuppliers(tpSuppliers); err != nil {
		t.Error("Error when setting TpSuppliers ", err.Error())
	}
	currentVersion := engine.CurrentStorDBVersions()
	err := tpSplMigrator.storDBOut.StorDB().SetVersions(currentVersion, false)
	if err != nil {
		t.Error("Error when setting version for TpSuppliers ", err.Error())
	}
}

func testTpSplITMove(t *testing.T) {
	err, _ := tpSplMigrator.Migrate([]string{utils.MetaTpSuppliers})
	if err != nil {
		t.Error("Error when migrating TpSuppliers ", err.Error())
	}
}

func testTpSplITCheckData(t *testing.T) {
	result, err := tpSplMigrator.storDBOut.StorDB().GetTPSuppliers(
		tpSuppliers[0].TPid, "", tpSuppliers[0].ID)
	if err != nil {
		t.Error("Error when getting TpSuppliers ", err.Error())
	}
	sort.Strings(result[0].FilterIDs)
	if !reflect.DeepEqual(tpSuppliers[0], result[0]) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(tpSuppliers[0]), utils.ToJSON(result[0]))
	}
	result, err = tpSplMigrator.storDBIn.StorDB().GetTPSuppliers(
		tpSuppliers[0].TPid, "", tpSuppliers[0].ID)
	if err != utils.ErrNotFound {
		t.Error(err)
	}
}
