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
}

func testTpDstITConnect(t *testing.T) {
	var err error
	tpDstPathIn = path.Join(*dataDir, "conf", "samples", "tutmongo")
	tpDstCfgIn, err = config.NewCGRConfigFromFolder(tpDstPathIn)
	if err != nil {
		t.Fatal(err)
	}
	tpDstPathOut = path.Join(*dataDir, "conf", "samples", "tutmysql")
	tpDstCfgOut, err = config.NewCGRConfigFromFolder(tpDstPathOut)
	if err != nil {
		t.Fatal(err)
	}
	storDBIn, err := NewMigratorStorDB(tpDstCfgIn.StorDbCfg().StorDBType,
		tpDstCfgIn.StorDbCfg().StorDBHost, tpDstCfgIn.StorDbCfg().StorDBPort,
		tpDstCfgIn.StorDbCfg().StorDBName, tpDstCfgIn.StorDbCfg().StorDBUser,
		tpDstCfgIn.StorDbCfg().StorDBPass,
		config.CgrConfig().StorDbCfg().StorDBMaxOpenConns,
		config.CgrConfig().StorDbCfg().StorDBMaxIdleConns,
		config.CgrConfig().StorDbCfg().StorDBConnMaxLifetime,
		config.CgrConfig().StorDbCfg().StorDBCDRSIndexes)
	if err != nil {
		log.Fatal(err)
	}
	storDBOut, err := NewMigratorStorDB(tpDstCfgOut.StorDbCfg().StorDBType,
		tpDstCfgOut.StorDbCfg().StorDBHost, tpDstCfgOut.StorDbCfg().StorDBPort,
		tpDstCfgOut.StorDbCfg().StorDBName, tpDstCfgOut.StorDbCfg().StorDBUser,
		tpDstCfgOut.StorDbCfg().StorDBPass,
		config.CgrConfig().StorDbCfg().StorDBMaxOpenConns,
		config.CgrConfig().StorDbCfg().StorDBMaxIdleConns,
		config.CgrConfig().StorDbCfg().StorDBConnMaxLifetime,
		config.CgrConfig().StorDbCfg().StorDBCDRSIndexes)
	if err != nil {
		log.Fatal(err)
	}
	tpDstMigrator, err = NewMigrator(nil, nil, storDBIn, storDBOut, false, false, false)
	if err != nil {
		log.Fatal(err)
	}
}

func testTpDstITFlush(t *testing.T) {
	if err := tpDstMigrator.storDBIn.StorDB().Flush(
		path.Join(tpDstCfgIn.DataFolderPath, "storage", tpDstCfgIn.StorDbCfg().StorDBType)); err != nil {
		t.Error(err)
	}

	if err := tpDstMigrator.storDBOut.StorDB().Flush(
		path.Join(tpDstCfgOut.DataFolderPath, "storage", tpDstCfgOut.StorDbCfg().StorDBType)); err != nil {
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
