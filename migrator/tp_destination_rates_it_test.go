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
}

func testTpDstRtITConnect(t *testing.T) {
	var err error
	tpDstRtPathIn = path.Join(*dataDir, "conf", "samples", "tutmongo")
	tpDstRtCfgIn, err = config.NewCGRConfigFromFolder(tpDstRtPathIn)
	if err != nil {
		t.Fatal(err)
	}
	tpDstRtPathOut = path.Join(*dataDir, "conf", "samples", "tutmysql")
	tpDstRtCfgOut, err = config.NewCGRConfigFromFolder(tpDstRtPathOut)
	if err != nil {
		t.Fatal(err)
	}
	storDBIn, err := NewMigratorStorDB(tpDstRtCfgIn.StorDbCfg().StorDBType,
		tpDstRtCfgIn.StorDbCfg().StorDBHost, tpDstRtCfgIn.StorDbCfg().StorDBPort,
		tpDstRtCfgIn.StorDbCfg().StorDBName, tpDstRtCfgIn.StorDbCfg().StorDBUser,
		tpDstRtCfgIn.StorDbCfg().StorDBPass,
		config.CgrConfig().StorDbCfg().StorDBMaxOpenConns,
		config.CgrConfig().StorDbCfg().StorDBMaxIdleConns,
		config.CgrConfig().StorDbCfg().StorDBConnMaxLifetime,
		config.CgrConfig().StorDbCfg().StorDBCDRSIndexes)
	if err != nil {
		log.Fatal(err)
	}
	storDBOut, err := NewMigratorStorDB(tpDstRtCfgOut.StorDbCfg().StorDBType,
		tpDstRtCfgOut.StorDbCfg().StorDBHost, tpDstRtCfgOut.StorDbCfg().StorDBPort,
		tpDstRtCfgOut.StorDbCfg().StorDBName, tpDstRtCfgOut.StorDbCfg().StorDBUser,
		tpDstRtCfgOut.StorDbCfg().StorDBPass,
		config.CgrConfig().StorDbCfg().StorDBMaxOpenConns,
		config.CgrConfig().StorDbCfg().StorDBMaxIdleConns,
		config.CgrConfig().StorDbCfg().StorDBConnMaxLifetime,
		config.CgrConfig().StorDbCfg().StorDBCDRSIndexes)
	if err != nil {
		log.Fatal(err)
	}
	tpDstRtMigrator, err = NewMigrator(nil, nil, storDBIn, storDBOut, false, false, false)
	if err != nil {
		log.Fatal(err)
	}
}

func testTpDstRtITFlush(t *testing.T) {
	if err := tpDstRtMigrator.storDBIn.StorDB().Flush(
		path.Join(tpDstRtCfgIn.DataFolderPath, "storage", tpDstRtCfgIn.StorDbCfg().StorDBType)); err != nil {
		t.Error(err)
	}

	if err := tpDstRtMigrator.storDBOut.StorDB().Flush(
		path.Join(tpDstRtCfgOut.DataFolderPath, "storage", tpDstRtCfgOut.StorDbCfg().StorDBType)); err != nil {
		t.Error(err)
	}
}

func testTpDstRtITPopulate(t *testing.T) {
	tpDestinationRate = []*utils.TPDestinationRate{
		{
			TPid: utils.TEST_SQL,
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
	err := tpDstRtMigrator.storDBOut.StorDB().SetVersions(currentVersion, false)
	if err != nil {
		t.Error("Error when setting version for TpDestinationRate ", err.Error())
	}
}

func testTpDstRtITMove(t *testing.T) {
	err, _ := tpDstRtMigrator.Migrate([]string{utils.MetaTpDestinationRates})
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
