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
	tpResPathIn   string
	tpResPathOut  string
	tpResCfgIn    *config.CGRConfig
	tpResCfgOut   *config.CGRConfig
	tpResMigrator *Migrator
	tpResources   []*utils.TPResourceProfile
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
	tpResMigrator.Close()
}

func testTpResITConnect(t *testing.T) {
	var err error
	tpResPathIn = path.Join(*utils.DataDir, "conf", "samples", "tutmongo")
	tpResCfgIn, err = config.NewCGRConfigFromPath(tpResPathIn)
	if err != nil {
		t.Fatal(err)
	}
	tpResPathOut = path.Join(*utils.DataDir, "conf", "samples", "tutmysql")
	tpResCfgOut, err = config.NewCGRConfigFromPath(tpResPathOut)
	if err != nil {
		t.Fatal(err)
	}
	storDBIn, err := NewMigratorStorDB(tpResCfgIn.StorDbCfg().Type,
		tpResCfgIn.StorDbCfg().Host, tpResCfgIn.StorDbCfg().Port,
		tpResCfgIn.StorDbCfg().Name, tpResCfgIn.StorDbCfg().User,
		tpResCfgIn.StorDbCfg().Password, tpResCfgIn.GeneralCfg().DBDataEncoding,
		tpResCfgIn.StorDbCfg().StringIndexedFields, tpResCfgIn.StorDbCfg().PrefixIndexedFields,
		tpResCfgIn.StorDbCfg().Opts, nil)
	if err != nil {
		log.Fatal(err)
	}
	storDBOut, err := NewMigratorStorDB(tpResCfgOut.StorDbCfg().Type,
		tpResCfgOut.StorDbCfg().Host, tpResCfgOut.StorDbCfg().Port,
		tpResCfgOut.StorDbCfg().Name, tpResCfgOut.StorDbCfg().User,
		tpResCfgOut.StorDbCfg().Password, tpResCfgOut.GeneralCfg().DBDataEncoding,
		tpResCfgIn.StorDbCfg().StringIndexedFields, tpResCfgIn.StorDbCfg().PrefixIndexedFields,
		tpResCfgOut.StorDbCfg().Opts, nil)
	if err != nil {
		log.Fatal(err)
	}
	tpResMigrator, err = NewMigrator(nil, nil, storDBIn, storDBOut,
		false, false, false, false)
	if err != nil {
		log.Fatal(err)
	}
}

func testTpResITFlush(t *testing.T) {
	if err := tpResMigrator.storDBIn.StorDB().Flush(
		path.Join(tpResCfgIn.DataFolderPath, "storage", dbPath(tpResCfgIn.StorDbCfg().Type))); err != nil {
		t.Error(err)
	}

	if err := tpResMigrator.storDBOut.StorDB().Flush(
		path.Join(tpResCfgOut.DataFolderPath, "storage", dbPath(tpResCfgOut.StorDbCfg().Type))); err != nil {
		t.Error(err)
	}
}

func testTpResITPopulate(t *testing.T) {
	tpResources = []*utils.TPResourceProfile{
		{
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
	err := tpResMigrator.storDBIn.StorDB().SetVersions(currentVersion, false)
	if err != nil {
		t.Error("Error when setting version for TpResources ", err.Error())
	}
}

func testTpResITMove(t *testing.T) {
	_, err := tpResMigrator.Migrate([]string{utils.MetaTpResources})
	if err != nil {
		t.Error("Error when migrating TpResources ", err.Error())
	}
}

func testTpResITCheckData(t *testing.T) {
	result, err := tpResMigrator.storDBOut.StorDB().GetTPResources(
		tpResources[0].TPid, "", tpResources[0].ID)
	if err != nil {
		t.Error("Error when getting TpResources ", err.Error())
	}
	sort.Strings(result[0].ThresholdIDs)
	if !reflect.DeepEqual(tpResources[0], result[0]) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(tpResources[0]), utils.ToJSON(result[0]))
	}
	result, err = tpResMigrator.storDBIn.StorDB().GetTPResources(
		tpResources[0].TPid, "", tpResources[0].ID)
	if err != utils.ErrNotFound {
		t.Error(err)
	}
}
