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

// package migrator

// import (
// 	"log"
// 	"path"
// 	"reflect"
// 	"testing"

// 	"github.com/cgrates/cgrates/config"
// 	"github.com/cgrates/cgrates/engine"
// 	"github.com/cgrates/cgrates/utils"
// )

// var (
// 	tpSplPathIn   string
// 	tpSplPathOut  string
// 	tpSplCfgIn    *config.CGRConfig
// 	tpSplCfgOut   *config.CGRConfig
// 	tpSplMigrator *Migrator
// 	tpSuppliers   []*utils.TPSupplierProfile
// )

// var sTestsTpSplIT = []func(t *testing.T){
// 	testTpSplITConnect,
// 	testTpSplITFlush,
// 	testTpSplITPopulate,
// 	testTpSplITMove,
// 	testTpSplITCheckData,
// }

// func TestTpSplMove(t *testing.T) {
// 	for _, stest := range sTestsTpSplIT {
// 		t.Run("TestTpSplMove", stest)
// 	}
// }

// func testTpSplITConnect(t *testing.T) {
// 	var err error
// 	tpSplPathIn = path.Join(*dataDir, "conf", "samples", "tutmongo")
// 	tpSplCfgIn, err = config.NewCGRConfigFromFolder(tpSplPathIn)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	tpSplPathOut = path.Join(*dataDir, "conf", "samples", "tutmysql")
// 	tpSplCfgOut, err = config.NewCGRConfigFromFolder(tpSplPathOut)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	storDBIn, err := engine.ConfigureStorDB(tpSplCfgIn.StorDBType, tpSplCfgIn.StorDBHost,
// 		tpSplCfgIn.StorDBPort, tpSplCfgIn.StorDBName,
// 		tpSplCfgIn.StorDBUser, tpSplCfgIn.StorDBPass,
// 		config.CgrConfig().StorDBMaxOpenConns,
// 		config.CgrConfig().StorDBMaxIdleConns,
// 		config.CgrConfig().StorDBConnMaxLifetime,
// 		config.CgrConfig().StorDBCDRSIndexes)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	storDBOut, err := engine.ConfigureStorDB(tpSplCfgOut.StorDBType,
// 		tpSplCfgOut.StorDBHost, tpSplCfgOut.StorDBPort, tpSplCfgOut.StorDBName,
// 		tpSplCfgOut.StorDBUser, tpSplCfgOut.StorDBPass,
// 		config.CgrConfig().StorDBMaxOpenConns,
// 		config.CgrConfig().StorDBMaxIdleConns,
// 		config.CgrConfig().StorDBConnMaxLifetime,
// 		config.CgrConfig().StorDBCDRSIndexes)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	tpSplMigrator, err = NewMigrator(nil, nil, tpSplCfgIn.DataDbType,
// 		tpSplCfgIn.DBDataEncoding, storDBIn, storDBOut, tpSplCfgIn.StorDBType, nil,
// 		tpSplCfgIn.DataDbType, tpSplCfgIn.DBDataEncoding, nil,
// 		tpSplCfgIn.StorDBType, false, false, false, false, false)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// }

// func testTpSplITFlush(t *testing.T) {
// 	if err := tpSplMigrator.storDBIn.Flush(
// 		path.Join(tpSplCfgIn.DataFolderPath, "storage", tpSplCfgIn.StorDBType)); err != nil {
// 		t.Error(err)
// 	}

// 	if err := tpSplMigrator.storDBOut.Flush(
// 		path.Join(tpSplCfgOut.DataFolderPath, "storage", tpSplCfgOut.StorDBType)); err != nil {
// 		t.Error(err)
// 	}
// }

// func testTpSplITPopulate(t *testing.T) {
// 	tpSuppliers = []*utils.TPSupplierProfile{
// 		&utils.TPSupplierProfile{
// 			TPid:      "TP1",
// 			Tenant:    "cgrates.org",
// 			ID:        "SUPL_1",
// 			FilterIDs: []string{"FLTR_ACNT_dan", "FLTR_DST_DE"},
// 			ActivationInterval: &utils.TPActivationInterval{
// 				ActivationTime: "2014-07-29T15:00:00Z",
// 				ExpiryTime:     "",
// 			},
// 			Sorting:           "*lowest_cost",
// 			SortingParameters: []string{},
// 			Suppliers: []*utils.TPSupplier{
// 				&utils.TPSupplier{
// 					ID:                 "supplier1",
// 					FilterIDs:          []string{"FLTR_1"},
// 					AccountIDs:         []string{"Acc1", "Acc2"},
// 					RatingPlanIDs:      []string{"RPL_1"},
// 					ResourceIDs:        []string{"ResGroup1"},
// 					StatIDs:            []string{"Stat1"},
// 					Weight:             10,
// 					Blocker:            false,
// 					SupplierParameters: "SortingParam1",
// 				},
// 			},
// 			Weight: 20,
// 		},
// 	}
// 	if err := tpSplMigrator.storDBIn.SetTPSuppliers(tpSuppliers); err != nil {
// 		t.Error("Error when setting TpFilter ", err.Error())
// 	}
// 	currentVersion := engine.CurrentStorDBVersions()
// 	err := tpSplMigrator.storDBOut.SetVersions(currentVersion, false)
// 	if err != nil {
// 		t.Error("Error when setting version for TpFilter ", err.Error())
// 	}
// }

// func testTpSplITMove(t *testing.T) {
// 	err, _ := tpSplMigrator.Migrate([]string{utils.MetaTpSuppliers})
// 	if err != nil {
// 		t.Error("Error when migrating TpFilter ", err.Error())
// 	}
// }

// func testTpSplITCheckData(t *testing.T) {
// 	result, err := tpSplMigrator.storDBOut.GetTPSuppliers(
// 		tpSuppliers[0].TPid, tpSuppliers[0].ID)
// 	if err != nil {
// 		t.Error("Error when getting TpFilter ", err.Error())
// 	}
// 	if !reflect.DeepEqual(tpSuppliers[0], result[0]) {
// 		t.Errorf("Expecting: %+v, received: %+v", tpSuppliers[0], result[0])
// 	}
// 	result, err = tpSplMigrator.storDBIn.GetTPSuppliers(
// 		tpSuppliers[0].TPid, tpSuppliers[0].ID)
// 	if err != utils.ErrNotFound {
// 		t.Error(err)
// 	}
// }
