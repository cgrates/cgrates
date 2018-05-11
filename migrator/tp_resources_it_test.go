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
	tpResPathIn   string
	tpResPathOut  string
	tpResCfgIn    *config.CGRConfig
	tpResCfgOut   *config.CGRConfig
	tpResMigrator *Migrator
	tpResources   []*utils.TPResource
)

var sTestsTpResIT = []func(t *testing.T){
	testTpResITConnect,
	testTpResITFlush,
	testTpResITPopulate,
	testTpResITMove,
	testTpResITCheckData,
}

func TestTpResMove(t *testing.T) {
	for _, stest := range sTestsTpResIT {
		t.Run("TestTpResMove", stest)
	}
}

func testTpResITConnect(t *testing.T) {
	var err error
	tpResPathIn = path.Join(*dataDir, "conf", "samples", "tutmongo")
	tpResCfgIn, err = config.NewCGRConfigFromFolder(tpResPathIn)
	if err != nil {
		t.Fatal(err)
	}
	tpResPathOut = path.Join(*dataDir, "conf", "samples", "tutmysql")
	tpResCfgOut, err = config.NewCGRConfigFromFolder(tpResPathOut)
	if err != nil {
		t.Fatal(err)
	}
	storDBIn, err := NewMigratorStorDB(tpResCfgIn.StorDBType, tpResCfgIn.StorDBHost,
		tpResCfgIn.StorDBPort, tpResCfgIn.StorDBName,
		tpResCfgIn.StorDBUser, tpResCfgIn.StorDBPass,
		config.CgrConfig().StorDBMaxOpenConns,
		config.CgrConfig().StorDBMaxIdleConns,
		config.CgrConfig().StorDBConnMaxLifetime,
		config.CgrConfig().StorDBCDRSIndexes)
	if err != nil {
		log.Fatal(err)
	}
	storDBOut, err := NewMigratorStorDB(tpResCfgOut.StorDBType,
		tpResCfgOut.StorDBHost, tpResCfgOut.StorDBPort, tpResCfgOut.StorDBName,
		tpResCfgOut.StorDBUser, tpResCfgOut.StorDBPass,
		config.CgrConfig().StorDBMaxOpenConns,
		config.CgrConfig().StorDBMaxIdleConns,
		config.CgrConfig().StorDBConnMaxLifetime,
		config.CgrConfig().StorDBCDRSIndexes)
	if err != nil {
		log.Fatal(err)
	}
	tpResMigrator, err = NewMigrator(nil, nil, storDBIn, storDBOut,
		false, false, false)
	if err != nil {
		log.Fatal(err)
	}
}

func testTpResITFlush(t *testing.T) {
	if err := tpResMigrator.storDBIn.StorDB().Flush(
		path.Join(tpResCfgIn.DataFolderPath, "storage", tpResCfgIn.StorDBType)); err != nil {
		t.Error(err)
	}

	if err := tpResMigrator.storDBOut.StorDB().Flush(
		path.Join(tpResCfgOut.DataFolderPath, "storage", tpResCfgOut.StorDBType)); err != nil {
		t.Error(err)
	}
}

func testTpResITPopulate(t *testing.T) {
	tpResources = []*utils.TPResource{
		&utils.TPResource{
			Tenant:    "cgrates.org",
			TPid:      "TPR1",
			ID:        "ResGroup1",
			FilterIDs: []string{"FLTR_1"},
			ActivationInterval: &utils.TPActivationInterval{
				ActivationTime: "2014-07-29T15:00:00Z",
				ExpiryTime:     "",
			},
			UsageTTL:          "1s",
			Limit:             "7",
			AllocationMessage: "",
			Blocker:           true,
			Stored:            true,
			Weight:            20,
			ThresholdIDs:      []string{"ValOne", "ValTwo"},
		},
	}
	if err := tpResMigrator.storDBIn.StorDB().SetTPResources(tpResources); err != nil {
		t.Error("Error when setting TpResources ", err.Error())
	}
	currentVersion := engine.CurrentStorDBVersions()
	err := tpResMigrator.storDBOut.StorDB().SetVersions(currentVersion, false)
	if err != nil {
		t.Error("Error when setting version for TpResources ", err.Error())
	}
}

func testTpResITMove(t *testing.T) {
	err, _ := tpResMigrator.Migrate([]string{utils.MetaTpResources})
	if err != nil {
		t.Error("Error when migrating TpResources ", err.Error())
	}
}

func testTpResITCheckData(t *testing.T) {
	result, err := tpResMigrator.storDBOut.StorDB().GetTPResources(
		tpResources[0].TPid, tpResources[0].ID)
	if err != nil {
		t.Error("Error when getting TpResources ", err.Error())
	}
	if !reflect.DeepEqual(tpResources[0], result[0]) {
		t.Errorf("Expecting: %+v, received: %+v", tpResources[0], result[0])
	}
	result, err = tpResMigrator.storDBIn.StorDB().GetTPResources(
		tpResources[0].TPid, tpResources[0].ID)
	if err != utils.ErrNotFound {
		t.Error(err)
	}
}
