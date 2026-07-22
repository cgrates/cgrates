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
	tpDstRtPathIn     string
	tpDstRtPathOut    string
	tpDstRtCfgIn      *config.CGRConfig
	tpDstRtCfgOut     *config.CGRConfig
	tpDstRtMigrator   *Migrator
	tpDestinationRate []*utils.TPDestinationRate
)

var sTestsTpDstRtIT = []func(t *testing.T){
	testTpDstRtITConnect,
	testTpDstRtITFlush,
	testTpDstRtITPopulate,
	testTpDstRtITMove,
	testTpDstRtITCheckData,
}

func TestTpDstRtMove(t *testing.T) {
	for _, stest := range sTestsTpDstRtIT {
		t.Run("TestTpDstRtMove", stest)
	}
	tpDstRtMigrator.Close()
}

func testTpDstRtITConnect(t *testing.T) {
	var err error
	tpDstRtPathIn = path.Join(*utils.DataDir, "conf", "samples", "tutmongo")
	tpDstRtCfgIn, err = config.NewCGRConfigFromPath(tpDstRtPathIn)
	if err != nil {
		t.Fatal(err)
	}
	tpDstRtPathOut = path.Join(*utils.DataDir, "conf", "samples", "tutmysql")
	tpDstRtCfgOut, err = config.NewCGRConfigFromPath(tpDstRtPathOut)
	if err != nil {
		t.Fatal(err)
	}
	storDBIn, err := NewMigratorStorDB(tpDstRtCfgIn.StorDbCfg().Type,
		tpDstRtCfgIn.StorDbCfg().Host, tpDstRtCfgIn.StorDbCfg().Port,
		tpDstRtCfgIn.StorDbCfg().Name, tpDstRtCfgIn.StorDbCfg().User,
		tpDstRtCfgIn.StorDbCfg().Password, tpDstRtCfgIn.GeneralCfg().DBDataEncoding,
		tpDstRtCfgIn.StorDbCfg().StringIndexedFields, tpDstRtCfgIn.StorDbCfg().PrefixIndexedFields,
		tpDstRtCfgIn.StorDbCfg().Opts, nil)
	if err != nil {
		log.Fatal(err)
	}
	storDBOut, err := NewMigratorStorDB(tpDstRtCfgOut.StorDbCfg().Type,
		tpDstRtCfgOut.StorDbCfg().Host, tpDstRtCfgOut.StorDbCfg().Port,
		tpDstRtCfgOut.StorDbCfg().Name, tpDstRtCfgOut.StorDbCfg().User,
		tpDstRtCfgOut.StorDbCfg().Password, tpDstRtCfgOut.GeneralCfg().DBDataEncoding,
		tpDstRtCfgIn.StorDbCfg().StringIndexedFields, tpDstRtCfgIn.StorDbCfg().PrefixIndexedFields,
		tpDstRtCfgOut.StorDbCfg().Opts, nil)
	if err != nil {
		log.Fatal(err)
	}
	tpDstRtMigrator, err = NewMigrator(nil, nil, storDBIn, storDBOut, false, false, false, false)
	if err != nil {
		log.Fatal(err)
	}
}

func testTpDstRtITFlush(t *testing.T) {
	if err := tpDstRtMigrator.storDBIn.StorDB().Flush(
		path.Join(tpDstRtCfgIn.DataFolderPath, "storage", dbPath(tpDstRtCfgIn.StorDbCfg().Type))); err != nil {
		t.Error(err)
	}

	if err := tpDstRtMigrator.storDBOut.StorDB().Flush(
		path.Join(tpDstRtCfgOut.DataFolderPath, "storage", dbPath(tpDstRtCfgOut.StorDbCfg().Type))); err != nil {
		t.Error(err)
	}
}

func testTpDstRtITPopulate(t *testing.T) {
	tpDestinationRate = []*utils.TPDestinationRate{
		{
			TPid: utils.TestSQL,
			ID:   "DR_FREESWITCH_USERS",
			DestinationRates: []*utils.DestinationRate{
				{
					DestinationId:    "FS_USERS",
					RateId:           "RT_FS_USERS",
					RoundingMethod:   "*up",
					RoundingDecimals: 2},
			},
		},
	}
	if err := tpDstRtMigrator.storDBIn.StorDB().SetTPDestinationRates(tpDestinationRate); err != nil {
		t.Error("Error when setting TpDestinationRate ", err.Error())
	}
	currentVersion := engine.CurrentStorDBVersions()
	err := tpDstRtMigrator.storDBIn.StorDB().SetVersions(currentVersion, false)
	if err != nil {
		t.Error("Error when setting version for TpDestinationRate ", err.Error())
	}
}

func testTpDstRtITMove(t *testing.T) {
	_, err := tpDstRtMigrator.Migrate([]string{utils.MetaTpDestinationRates})
	if err != nil {
		t.Error("Error when migrating TpDestinationRate ", err.Error())
	}
}

func testTpDstRtITCheckData(t *testing.T) {
	result, err := tpDstRtMigrator.storDBOut.StorDB().GetTPDestinationRates(
		tpDestinationRate[0].TPid, tpDestinationRate[0].ID, nil)
	if err != nil {
		t.Error("Error when getting TpDestinationRate ", err.Error())
	}
	if !reflect.DeepEqual(tpDestinationRate[0], result[0]) {
		t.Errorf("Expecting: %+v, received: %+v",
			utils.ToJSON(tpDestinationRate[0]), utils.ToJSON(result[0]))
	}
	result, err = tpDstRtMigrator.storDBIn.StorDB().GetTPDestinationRates(
		tpDestinationRate[0].TPid, tpDestinationRate[0].ID, nil)
	if err != utils.ErrNotFound {
		t.Error(err)
	}
}
