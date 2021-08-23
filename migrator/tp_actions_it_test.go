//go:build integration
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
	tpActMigrator.Close()
}

func testTpActITConnect(t *testing.T) {
	var err error
	tpActPathIn = path.Join(*dataDir, "conf", "samples", "tutmongo")
	tpActCfgIn, err = config.NewCGRConfigFromPath(tpActPathIn)
	if err != nil {
		t.Fatal(err)
	}
	tpActPathOut = path.Join(*dataDir, "conf", "samples", "tutmysql")
	tpActCfgOut, err = config.NewCGRConfigFromPath(tpActPathOut)
	if err != nil {
		t.Fatal(err)
	}
	storDBIn, err := NewMigratorStorDB(tpActCfgIn.StorDbCfg().Type,
		tpActCfgIn.StorDbCfg().Host, tpActCfgIn.StorDbCfg().Port,
		tpActCfgIn.StorDbCfg().Name, tpActCfgIn.StorDbCfg().User,
		tpActCfgIn.StorDbCfg().Password, tpActCfgIn.GeneralCfg().DBDataEncoding, tpActCfgIn.StorDbCfg().SSLMode,
		tpActCfgIn.StorDbCfg().MaxOpenConns, tpActCfgIn.StorDbCfg().MaxIdleConns,
		tpActCfgIn.StorDbCfg().ConnMaxLifetime, tpActCfgIn.StorDbCfg().StringIndexedFields,
		tpActCfgIn.StorDbCfg().PrefixIndexedFields, tpActCfgIn.StorDbCfg().Items)
	if err != nil {
		log.Fatal(err)
	}
	storDBOut, err := NewMigratorStorDB(tpActCfgOut.StorDbCfg().Type,
		tpActCfgOut.StorDbCfg().Host, tpActCfgOut.StorDbCfg().Port,
		tpActCfgOut.StorDbCfg().Name, tpActCfgOut.StorDbCfg().User,
		tpActCfgOut.StorDbCfg().Password, tpActCfgOut.GeneralCfg().DBDataEncoding, tpActCfgIn.StorDbCfg().SSLMode,
		tpActCfgIn.StorDbCfg().MaxOpenConns, tpActCfgIn.StorDbCfg().MaxIdleConns,
		tpActCfgIn.StorDbCfg().ConnMaxLifetime, tpActCfgIn.StorDbCfg().StringIndexedFields,
		tpActCfgIn.StorDbCfg().PrefixIndexedFields, tpActCfgOut.StorDbCfg().Items)
	if err != nil {
		log.Fatal(err)
	}
	tpActMigrator, err = NewMigrator(nil, nil, storDBIn, storDBOut, false, false, false, false)
	if err != nil {
		log.Fatal(err)
	}
}

func testTpActITFlush(t *testing.T) {
	if err := tpActMigrator.storDBIn.StorDB().Flush(
		path.Join(tpActCfgIn.DataFolderPath, "storage", tpActCfgIn.StorDbCfg().Type)); err != nil {
		t.Error(err)
	}

	if err := tpActMigrator.storDBOut.StorDB().Flush(
		path.Join(tpActCfgOut.DataFolderPath, "storage", tpActCfgOut.StorDbCfg().Type)); err != nil {
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
