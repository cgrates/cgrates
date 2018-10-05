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
	tpTimPathIn   string
	tpTimPathOut  string
	tpTimCfgIn    *config.CGRConfig
	tpTimCfgOut   *config.CGRConfig
	tpTimMigrator *Migrator
	tpTimings     []*utils.ApierTPTiming
)

var sTestsTpTimIT = []func(t *testing.T){
	testTpTimITConnect,
	testTpTimITFlush,
	testTpTimITPopulate,
	testTpTimITMove,
	testTpTimITCheckData,
}

func TestTpTimMove(t *testing.T) {
	for _, stest := range sTestsTpTimIT {
		t.Run("TestTpTimMove", stest)
	}
}

func testTpTimITConnect(t *testing.T) {
	var err error
	tpTimPathIn = path.Join(*dataDir, "conf", "samples", "tutmongo")
	tpTimCfgIn, err = config.NewCGRConfigFromFolder(tpTimPathIn)
	if err != nil {
		t.Fatal(err)
	}
	tpTimPathOut = path.Join(*dataDir, "conf", "samples", "tutmysql")
	tpTimCfgOut, err = config.NewCGRConfigFromFolder(tpTimPathOut)
	if err != nil {
		t.Fatal(err)
	}
	storDBIn, err := NewMigratorStorDB(tpTimCfgIn.StorDbCfg().StorDBType,
		tpTimCfgIn.StorDbCfg().StorDBHost, tpTimCfgIn.StorDbCfg().StorDBPort,
		tpTimCfgIn.StorDbCfg().StorDBName, tpTimCfgIn.StorDbCfg().StorDBUser,
		tpTimCfgIn.StorDbCfg().StorDBPass,
		config.CgrConfig().StorDbCfg().StorDBMaxOpenConns,
		config.CgrConfig().StorDbCfg().StorDBMaxIdleConns,
		config.CgrConfig().StorDbCfg().StorDBConnMaxLifetime,
		config.CgrConfig().StorDbCfg().StorDBCDRSIndexes)
	if err != nil {
		log.Fatal(err)
	}
	storDBOut, err := NewMigratorStorDB(tpTimCfgOut.StorDbCfg().StorDBType,
		tpTimCfgOut.StorDbCfg().StorDBHost, tpTimCfgOut.StorDbCfg().StorDBPort,
		tpTimCfgOut.StorDbCfg().StorDBName, tpTimCfgOut.StorDbCfg().StorDBUser,
		tpTimCfgOut.StorDbCfg().StorDBPass,
		config.CgrConfig().StorDbCfg().StorDBMaxOpenConns,
		config.CgrConfig().StorDbCfg().StorDBMaxIdleConns,
		config.CgrConfig().StorDbCfg().StorDBConnMaxLifetime,
		config.CgrConfig().StorDbCfg().StorDBCDRSIndexes)
	if err != nil {
		log.Fatal(err)
	}
	tpTimMigrator, err = NewMigrator(nil, nil, storDBIn, storDBOut, false, false, false)
	if err != nil {
		log.Fatal(err)
	}
}

func testTpTimITFlush(t *testing.T) {
	if err := tpTimMigrator.storDBIn.StorDB().Flush(
		path.Join(tpTimCfgIn.DataFolderPath, "storage", tpTimCfgIn.StorDbCfg().StorDBType)); err != nil {
		t.Error(err)
	}

	if err := tpTimMigrator.storDBOut.StorDB().Flush(
		path.Join(tpTimCfgOut.DataFolderPath, "storage", tpTimCfgOut.StorDbCfg().StorDBType)); err != nil {
		t.Error(err)
	}
}

func testTpTimITPopulate(t *testing.T) {
	tpTimings = []*utils.ApierTPTiming{
		{
			TPid:      "TPT1",
			ID:        "Timing",
			Years:     "2017",
			Months:    "05",
			MonthDays: "01",
			WeekDays:  "1",
			Time:      "15:00:00Z",
		},
	}
	if err := tpTimMigrator.storDBIn.StorDB().SetTPTimings(tpTimings); err != nil {
		t.Error("Error when setting TpTimings ", err.Error())
	}
	currentVersion := engine.CurrentStorDBVersions()
	err := tpTimMigrator.storDBOut.StorDB().SetVersions(currentVersion, false)
	if err != nil {
		t.Error("Error when setting version for TpTimings ", err.Error())
	}
}

func testTpTimITMove(t *testing.T) {
	err, _ := tpTimMigrator.Migrate([]string{utils.MetaTpTiming})
	if err != nil {
		t.Error("Error when migrating TpTimings ", err.Error())
	}
}

func testTpTimITCheckData(t *testing.T) {
	result, err := tpTimMigrator.storDBOut.StorDB().GetTPTimings(
		tpTimings[0].TPid, tpTimings[0].ID)
	if err != nil {
		t.Error("Error when getting TpTimings ", err.Error())
	}
	if !reflect.DeepEqual(tpTimings[0], result[0]) {
		t.Errorf("Expecting: %+v, received: %+v", tpTimings[0], result[0])
	}
	result, err = tpTimMigrator.storDBIn.StorDB().GetTPTimings(
		tpTimings[0].TPid, tpTimings[0].ID)
	if err != utils.ErrNotFound {
		t.Error(err)
	}
}
