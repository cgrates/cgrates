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
	tpDstPathIn   string
	tpDstPathOut  string
	tpDstCfgIn    *config.CGRConfig
	tpDstCfgOut   *config.CGRConfig
	tpDstMigrator *Migrator
	tpDestination []*utils.TPDestination
)

var sTestsTpDstIT = []func(t *testing.T){
	testTpDstITConnect,
	testTpDstITFlush,
	testTpDstITPopulate,
	testTpDstITMove,
	testTpDstITCheckData,
}

func TestTpDstMove(t *testing.T) {
	for _, stest := range sTestsTpDstIT {
		t.Run("TestTpDstMove", stest)
	}
	tpDstMigrator.Close()
}

func testTpDstITConnect(t *testing.T) {
	var err error
	tpDstPathIn = path.Join(*utils.DataDir, "conf", "samples", "tutmongo")
	tpDstCfgIn, err = config.NewCGRConfigFromPath(tpDstPathIn)
	if err != nil {
		t.Fatal(err)
	}
	tpDstPathOut = path.Join(*utils.DataDir, "conf", "samples", "tutmysql")
	tpDstCfgOut, err = config.NewCGRConfigFromPath(tpDstPathOut)
	if err != nil {
		t.Fatal(err)
	}
	storDBIn, err := NewMigratorStorDB(tpDstCfgIn.StorDbCfg().Type,
		tpDstCfgIn.StorDbCfg().Host, tpDstCfgIn.StorDbCfg().Port,
		tpDstCfgIn.StorDbCfg().Name, tpDstCfgIn.StorDbCfg().User,
		tpDstCfgIn.StorDbCfg().Password, tpDstCfgIn.GeneralCfg().DBDataEncoding, tpDstCfgIn.StorDbCfg().SSLMode,
		tpDstCfgIn.StorDbCfg().MaxOpenConns, tpDstCfgIn.StorDbCfg().MaxIdleConns,
		tpDstCfgIn.StorDbCfg().ConnMaxLifetime, tpDstCfgIn.StorDbCfg().StringIndexedFields,
		tpDstCfgIn.StorDbCfg().PrefixIndexedFields, tpDstCfgIn.StorDbCfg().Items)
	if err != nil {
		log.Fatal(err)
	}
	storDBOut, err := NewMigratorStorDB(tpDstCfgOut.StorDbCfg().Type,
		tpDstCfgOut.StorDbCfg().Host, tpDstCfgOut.StorDbCfg().Port,
		tpDstCfgOut.StorDbCfg().Name, tpDstCfgOut.StorDbCfg().User,
		tpDstCfgOut.StorDbCfg().Password, tpDstCfgOut.GeneralCfg().DBDataEncoding, tpDstCfgIn.StorDbCfg().SSLMode,
		tpDstCfgIn.StorDbCfg().MaxOpenConns, tpDstCfgIn.StorDbCfg().MaxIdleConns,
		tpDstCfgIn.StorDbCfg().ConnMaxLifetime, tpDstCfgIn.StorDbCfg().StringIndexedFields,
		tpDstCfgIn.StorDbCfg().PrefixIndexedFields, tpDstCfgOut.StorDbCfg().Items)
	if err != nil {
		log.Fatal(err)
	}
	tpDstMigrator, err = NewMigrator(nil, nil, storDBIn, storDBOut, false, false, false, false)
	if err != nil {
		log.Fatal(err)
	}
}

func testTpDstITFlush(t *testing.T) {
	if err := tpDstMigrator.storDBIn.StorDB().Flush(
		path.Join(tpDstCfgIn.DataFolderPath, "storage", dbPath(tpDstCfgIn.StorDbCfg().Type))); err != nil {
		t.Error(err)
	}

	if err := tpDstMigrator.storDBOut.StorDB().Flush(
		path.Join(tpDstCfgOut.DataFolderPath, "storage", dbPath(tpDstCfgOut.StorDbCfg().Type))); err != nil {
		t.Error(err)
	}
}

func testTpDstITPopulate(t *testing.T) {
	tpDestination = []*utils.TPDestination{
		{
			TPid:     "TPD",
			ID:       "GERMANY",
			Prefixes: []string{"+49", "+4915"},
		},
	}
	if err := tpDstMigrator.storDBIn.StorDB().SetTPDestinations(tpDestination); err != nil {
		t.Error("Error when setting TpDestination ", err.Error())
	}
	currentVersion := engine.CurrentStorDBVersions()
	err := tpDstMigrator.storDBOut.StorDB().SetVersions(currentVersion, false)
	if err != nil {
		t.Error("Error when setting version for TpDestination ", err.Error())
	}
}

func testTpDstITMove(t *testing.T) {
	err, _ := tpDstMigrator.Migrate([]string{utils.MetaTpDestinations})
	if err != nil {
		t.Error("Error when migrating TpDestination ", err.Error())
	}
}

func testTpDstITCheckData(t *testing.T) {
	result, err := tpDstMigrator.storDBOut.StorDB().GetTPDestinations(
		tpDestination[0].TPid, tpDestination[0].ID)
	if err != nil {
		t.Error("Error when getting TpDestination ", err.Error())
	}
	if !reflect.DeepEqual(tpDestination[0], result[0]) {
		t.Errorf("Expecting: %+v, received: %+v",
			utils.ToJSON(tpDestination[0]), utils.ToJSON(result[0]))
	}
	result, err = tpDstMigrator.storDBIn.StorDB().GetTPDestinations(
		tpDestination[0].TPid, tpDestination[0].ID)
	if err != utils.ErrNotFound {
		t.Error(err)
	}
}
