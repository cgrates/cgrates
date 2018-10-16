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
	tpActPathIn   string
	tpActPathOut  string
	tpActCfgIn    *config.CGRConfig
	tpActCfgOut   *config.CGRConfig
	tpActMigrator *Migrator
	tpActions     []*utils.TPActions
)

var sTestsTpActIT = []func(t *testing.T){
	testTpActITConnect,
	testTpActITFlush,
	testTpActITPopulate,
	testTpActITMove,
	testTpActITCheckData,
}

func TestTpActMove(t *testing.T) {
	for _, stest := range sTestsTpActIT {
		t.Run("TestTpActMove", stest)
	}
}

func testTpActITConnect(t *testing.T) {
	var err error
	tpActPathIn = path.Join(*dataDir, "conf", "samples", "tutmongo")
	tpActCfgIn, err = config.NewCGRConfigFromFolder(tpActPathIn)
	if err != nil {
		t.Fatal(err)
	}
	tpActPathOut = path.Join(*dataDir, "conf", "samples", "tutmysql")
	tpActCfgOut, err = config.NewCGRConfigFromFolder(tpActPathOut)
	if err != nil {
		t.Fatal(err)
	}
	storDBIn, err := NewMigratorStorDB(tpActCfgIn.StorDbCfg().StorDBType,
		tpActCfgIn.StorDbCfg().StorDBHost, tpActCfgIn.StorDbCfg().StorDBPort,
		tpActCfgIn.StorDbCfg().StorDBName, tpActCfgIn.StorDbCfg().StorDBUser,
		tpActCfgIn.StorDbCfg().StorDBPass,
		config.CgrConfig().StorDbCfg().StorDBMaxOpenConns,
		config.CgrConfig().StorDbCfg().StorDBMaxIdleConns,
		config.CgrConfig().StorDbCfg().StorDBConnMaxLifetime,
		config.CgrConfig().StorDbCfg().StorDBCDRSIndexes)
	if err != nil {
		log.Fatal(err)
	}
	storDBOut, err := NewMigratorStorDB(tpActCfgOut.StorDbCfg().StorDBType,
		tpActCfgOut.StorDbCfg().StorDBHost, tpActCfgOut.StorDbCfg().StorDBPort,
		tpActCfgOut.StorDbCfg().StorDBName, tpActCfgOut.StorDbCfg().StorDBUser,
		tpActCfgOut.StorDbCfg().StorDBPass,
		config.CgrConfig().StorDbCfg().StorDBMaxOpenConns,
		config.CgrConfig().StorDbCfg().StorDBMaxIdleConns,
		config.CgrConfig().StorDbCfg().StorDBConnMaxLifetime,
		config.CgrConfig().StorDbCfg().StorDBCDRSIndexes)
	if err != nil {
		log.Fatal(err)
	}
	tpActMigrator, err = NewMigrator(nil, nil, storDBIn, storDBOut, false, false, false)
	if err != nil {
		log.Fatal(err)
	}
}

func testTpActITFlush(t *testing.T) {
	if err := tpActMigrator.storDBIn.StorDB().Flush(
		path.Join(tpActCfgIn.DataFolderPath, "storage", tpActCfgIn.StorDbCfg().StorDBType)); err != nil {
		t.Error(err)
	}

	if err := tpActMigrator.storDBOut.StorDB().Flush(
		path.Join(tpActCfgOut.DataFolderPath, "storage", tpActCfgOut.StorDbCfg().StorDBType)); err != nil {
		t.Error(err)
	}
}

func testTpActITPopulate(t *testing.T) {
	tpActions = []*utils.TPActions{
		{
			TPid: "TPAcc",
			ID:   "ID",
			Actions: []*utils.TPAction{
				{
					Identifier:      "*log",
					BalanceId:       "BalID1",
					BalanceUuid:     "",
					BalanceType:     "*monetary",
					Directions:      "*out",
					Units:           "120",
					ExpiryTime:      "*unlimited",
					Filter:          "",
					TimingTags:      "2014-01-14T00:00:00Z",
					DestinationIds:  "*any",
					RatingSubject:   "SPECIAL_1002",
					Categories:      "",
					SharedGroups:    "SHARED_A",
					BalanceWeight:   "11",
					ExtraParameters: "",
					BalanceBlocker:  "false",
					BalanceDisabled: "false",
					Weight:          11,
				},
				{
					Identifier:      "*topup_reset",
					BalanceId:       "BalID2",
					BalanceUuid:     "",
					BalanceType:     "*data",
					Directions:      "*out",
					Units:           "10",
					ExpiryTime:      "*unlimited",
					Filter:          "",
					TimingTags:      "2014-01-14T00:00:00Z",
					DestinationIds:  "DST_1002",
					RatingSubject:   "SPECIAL_1002",
					Categories:      "",
					SharedGroups:    "SHARED_A",
					BalanceWeight:   "10",
					ExtraParameters: "",
					BalanceBlocker:  "false",
					BalanceDisabled: "false",
					Weight:          10,
				},
			},
		},
	}
	if err := tpActMigrator.storDBIn.StorDB().SetTPActions(tpActions); err != nil {
		t.Error("Error when setting TpActions ", err.Error())
	}
	currentVersion := engine.CurrentStorDBVersions()
	err := tpActMigrator.storDBOut.StorDB().SetVersions(currentVersion, false)
	if err != nil {
		t.Error("Error when setting version for TpActions ", err.Error())
	}
}

func testTpActITMove(t *testing.T) {
	err, _ := tpActMigrator.Migrate([]string{utils.MetaTpActions})
	if err != nil {
		t.Error("Error when migrating TpActions ", err.Error())
	}
}

func testTpActITCheckData(t *testing.T) {
	result, err := tpActMigrator.storDBOut.StorDB().GetTPActions(
		tpActions[0].TPid, tpActions[0].ID)
	if err != nil {
		t.Error("Error when getting TpActions ", err.Error())
	}
	if !reflect.DeepEqual(tpActions[0], result[0]) {
		t.Errorf("Expecting: %+v, received: %+v",
			utils.ToJSON(tpActions[0]), utils.ToJSON(result[0]))
	}
	result, err = tpActMigrator.storDBIn.StorDB().GetTPActions(
		tpActions[0].TPid, tpActions[0].ID)
	if err != utils.ErrNotFound {
		t.Error(err)
	}
}
