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
	tpAliPathIn   string
	tpAliPathOut  string
	tpAliCfgIn    *config.CGRConfig
	tpAliCfgOut   *config.CGRConfig
	tpAliMigrator *Migrator
	tpAliases     []*utils.TPAliases
)

var sTestsTpAliIT = []func(t *testing.T){
	testTpAliITConnect,
	testTpAliITFlush,
	testTpAliITPopulate,
	testTpAliITMove,
	testTpAliITCheckData,
}

func TestTpAliMove(t *testing.T) {
	for _, stest := range sTestsTpAliIT {
		t.Run("TestTpAliMove", stest)
	}
}

func testTpAliITConnect(t *testing.T) {
	var err error
	tpAliPathIn = path.Join(*dataDir, "conf", "samples", "tutmongo")
	tpAliCfgIn, err = config.NewCGRConfigFromFolder(tpAliPathIn)
	if err != nil {
		t.Fatal(err)
	}
	tpAliPathOut = path.Join(*dataDir, "conf", "samples", "tutmysql")
	tpAliCfgOut, err = config.NewCGRConfigFromFolder(tpAliPathOut)
	if err != nil {
		t.Fatal(err)
	}
	storDBIn, err := NewMigratorStorDB(tpAliCfgIn.StorDbCfg().StorDBType,
		tpAliCfgIn.StorDbCfg().StorDBHost, tpAliCfgIn.StorDbCfg().StorDBPort,
		tpAliCfgIn.StorDbCfg().StorDBName, tpAliCfgIn.StorDbCfg().StorDBUser,
		tpAliCfgIn.StorDbCfg().StorDBPass,
		config.CgrConfig().StorDbCfg().StorDBMaxOpenConns,
		config.CgrConfig().StorDbCfg().StorDBMaxIdleConns,
		config.CgrConfig().StorDbCfg().StorDBConnMaxLifetime,
		config.CgrConfig().StorDbCfg().StorDBCDRSIndexes)
	if err != nil {
		log.Fatal(err)
	}
	storDBOut, err := NewMigratorStorDB(tpAliCfgOut.StorDbCfg().StorDBType,
		tpAliCfgOut.StorDbCfg().StorDBHost, tpAliCfgOut.StorDbCfg().StorDBPort,
		tpAliCfgOut.StorDbCfg().StorDBName, tpAliCfgOut.StorDbCfg().StorDBUser,
		tpAliCfgOut.StorDbCfg().StorDBPass,
		config.CgrConfig().StorDbCfg().StorDBMaxOpenConns,
		config.CgrConfig().StorDbCfg().StorDBMaxIdleConns,
		config.CgrConfig().StorDbCfg().StorDBConnMaxLifetime,
		config.CgrConfig().StorDbCfg().StorDBCDRSIndexes)
	if err != nil {
		log.Fatal(err)
	}
	tpAliMigrator, err = NewMigrator(nil, nil, storDBIn, storDBOut, false, false, false)
	if err != nil {
		log.Fatal(err)
	}
}

func testTpAliITFlush(t *testing.T) {
	if err := tpAliMigrator.storDBIn.StorDB().Flush(
		path.Join(tpAliCfgIn.DataFolderPath, "storage", tpAliCfgIn.StorDbCfg().StorDBType)); err != nil {
		t.Error(err)
	}

	if err := tpAliMigrator.storDBOut.StorDB().Flush(
		path.Join(tpAliCfgOut.DataFolderPath, "storage", tpAliCfgOut.StorDbCfg().StorDBType)); err != nil {
		t.Error(err)
	}
}

func testTpAliITPopulate(t *testing.T) {
	tpAliases = []*utils.TPAliases{
		{
			TPid:      "testTPid1",
			Direction: "*out",
			Tenant:    "cgrates.org",
			Category:  "call",
			Account:   "1006",
			Subject:   "1006",
			Context:   "*rating",
			Values: []*utils.TPAliasValue{
				{
					DestinationId: "*any",
					Target:        "Subject",
					Original:      "1006",
					Alias:         "1001",
					Weight:        2,
				},
			},
		},
		{
			TPid:      "testTPid2",
			Direction: "*out",
			Tenant:    "cgrates.org",
			Category:  "call",
			Account:   "1001",
			Subject:   "1001",
			Context:   "*rating",
			Values: []*utils.TPAliasValue{
				{
					DestinationId: "*any",
					Target:        "Subject",
					Original:      "1001",
					Alias:         "1002",
					Weight:        2,
				},
			},
		},
	}
	if err := tpAliMigrator.storDBIn.StorDB().SetTPAliases(tpAliases); err != nil {
		t.Error("Error when setting TpAliases ", err.Error())
	}
	currentVersion := engine.CurrentStorDBVersions()
	err := tpAliMigrator.storDBOut.StorDB().SetVersions(currentVersion, false)
	if err != nil {
		t.Error("Error when setting version for TpAliases ", err.Error())
	}
}

func testTpAliITMove(t *testing.T) {
	err, _ := tpAliMigrator.Migrate([]string{utils.MetaTpAliases})
	if err != nil {
		t.Error("Error when migrating TpAliases ", err.Error())
	}
}

func testTpAliITCheckData(t *testing.T) {
	filter := &utils.TPAliases{TPid: tpAliases[0].TPid}
	result, err := tpAliMigrator.storDBOut.StorDB().GetTPAliases(filter)
	if err != nil {
		t.Error("Error when getting TpAliases ", err.Error())
	}
	if !reflect.DeepEqual(tpAliases[0], result[0]) {
		t.Errorf("Expecting: %+v, received: %+v",
			utils.ToJSON(tpAliases[0]), utils.ToJSON(result[0]))
	}
	result, err = tpAliMigrator.storDBIn.StorDB().GetTPAliases(filter)
	if err != utils.ErrNotFound {
		t.Error(err)
	}
}
