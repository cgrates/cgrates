//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
	tpAccActPathIn   string
	tpAccActPathOut  string
	tpAccActCfgIn    *config.CGRConfig
	tpAccActCfgOut   *config.CGRConfig
	tpAccActMigrator *Migrator
	tpAccountActions []*utils.TPAccountActions
)

var sTestsTpAccActIT = []func(t *testing.T){
	testTpAccActITConnect,
	testTpAccActITFlush,
	testTpAccActITPopulate,
	testTpAccActITMove,
	testTpAccActITCheckData,
}

func TestTpAccActMove(t *testing.T) {
	for _, stest := range sTestsTpAccActIT {
		t.Run("TestTpAccActMove", stest)
	}
	tpAccActMigrator.Close()
}

func testTpAccActITConnect(t *testing.T) {
	var err error
	tpAccActPathIn = path.Join(*utils.DataDir, "conf", "samples", "tutmongo")
	tpAccActCfgIn, err = config.NewCGRConfigFromPath(tpAccActPathIn)
	if err != nil {
		t.Fatal(err)
	}
	tpAccActPathOut = path.Join(*utils.DataDir, "conf", "samples", "tutmysql")
	tpAccActCfgOut, err = config.NewCGRConfigFromPath(tpAccActPathOut)
	if err != nil {
		t.Fatal(err)
	}
	storDBIn, err := NewMigratorStorDB(tpAccActCfgIn.StorDbCfg().Type,
		tpAccActCfgIn.StorDbCfg().Host, tpAccActCfgIn.StorDbCfg().Port,
		tpAccActCfgIn.StorDbCfg().Name, tpAccActCfgIn.StorDbCfg().User,
		tpAccActCfgIn.StorDbCfg().Password, tpAccActCfgIn.GeneralCfg().DBDataEncoding, tpAccActCfgIn.StorDbCfg().SSLMode,
		tpAccActCfgIn.StorDbCfg().MaxOpenConns, tpAccActCfgIn.StorDbCfg().MaxIdleConns,
		tpAccActCfgIn.StorDbCfg().ConnMaxLifetime, tpAccActCfgIn.StorDbCfg().StringIndexedFields,
		tpAccActCfgIn.StorDbCfg().PrefixIndexedFields, tpAccActCfgIn.StorDbCfg().Items)
	if err != nil {
		log.Fatal(err)
	}
	storDBOut, err := NewMigratorStorDB(tpAccActCfgOut.StorDbCfg().Type,
		tpAccActCfgOut.StorDbCfg().Host, tpAccActCfgOut.StorDbCfg().Port,
		tpAccActCfgOut.StorDbCfg().Name, tpAccActCfgOut.StorDbCfg().User,
		tpAccActCfgOut.StorDbCfg().Password, tpAccActCfgOut.GeneralCfg().DBDataEncoding, tpAccActCfgIn.StorDbCfg().SSLMode,
		tpAccActCfgIn.StorDbCfg().MaxOpenConns, tpAccActCfgIn.StorDbCfg().MaxIdleConns,
		tpAccActCfgIn.StorDbCfg().ConnMaxLifetime, tpAccActCfgIn.StorDbCfg().StringIndexedFields,
		tpAccActCfgIn.StorDbCfg().PrefixIndexedFields, tpAccActCfgOut.StorDbCfg().Items)
	if err != nil {
		log.Fatal(err)
	}
	tpAccActMigrator, err = NewMigrator(nil, nil, storDBIn, storDBOut, false, false, false, false)
	if err != nil {
		log.Fatal(err)
	}
}

func testTpAccActITFlush(t *testing.T) {

	if err := tpAccActMigrator.storDBIn.StorDB().Flush(
		path.Join(tpAccActCfgIn.DataFolderPath, "storage", dbPath(tpAccActCfgIn.StorDbCfg().Type))); err != nil {
		t.Error(err)
	}

	if err := tpAccActMigrator.storDBOut.StorDB().Flush(
		path.Join(tpAccActCfgOut.DataFolderPath, "storage", dbPath(tpAccActCfgOut.StorDbCfg().Type))); err != nil {
		t.Error(err)
	}
}

func testTpAccActITPopulate(t *testing.T) {
	tpAccountActions = []*utils.TPAccountActions{
		{
			TPid:          "TPAcc",
			LoadId:        "ID",
			Tenant:        "cgrates.org",
			Account:       "1001",
			ActionPlanId:  "PREPAID_10",
			AllowNegative: true,
			Disabled:      false,
		},
	}
	if err := tpAccActMigrator.storDBIn.StorDB().SetTPAccountActions(tpAccountActions); err != nil {
		t.Error("Error when setting TpAccountActions ", err.Error())
	}
	currentVersion := engine.CurrentStorDBVersions()
	err := tpAccActMigrator.storDBOut.StorDB().SetVersions(currentVersion, false)
	if err != nil {
		t.Error("Error when setting version for TpAccountActions ", err.Error())
	}
}

func testTpAccActITMove(t *testing.T) {
	err, _ := tpAccActMigrator.Migrate([]string{utils.MetaTpAccountActions})
	if err != nil {
		t.Error("Error when migrating TpAccountActions ", err.Error())
	}
}

func testTpAccActITCheckData(t *testing.T) {
	filter := &utils.TPAccountActions{TPid: tpAccountActions[0].TPid}
	result, err := tpAccActMigrator.storDBOut.StorDB().GetTPAccountActions(filter)
	if err != nil {
		t.Error("Error when getting TpAccountActions ", err.Error())
	}
	if !reflect.DeepEqual(tpAccountActions[0], result[0]) {
		t.Errorf("Expecting: %+v, received: %+v",
			utils.ToJSON(tpAccountActions[0]), utils.ToJSON(result[0]))
	}
	result, err = tpAccActMigrator.storDBIn.StorDB().GetTPAccountActions(filter)
	if err != utils.ErrNotFound {
		t.Error(err)
	}
}
