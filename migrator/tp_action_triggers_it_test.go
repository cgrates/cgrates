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
	tpActTrgPathIn   string
	tpActTrgPathOut  string
	tpActTrgCfgIn    *config.CGRConfig
	tpActTrgCfgOut   *config.CGRConfig
	tpActTrgMigrator *Migrator
	tpActionTriggers []*utils.TPActionTriggers
)

var sTestsTpActTrgIT = []func(t *testing.T){
	testTpActTrgITConnect,
	testTpActTrgITFlush,
	testTpActTrgITPopulate,
	testTpActTrgITMove,
	testTpActTrgITCheckData,
}

func TestTpActTrgMove(t *testing.T) {
	for _, stest := range sTestsTpActTrgIT {
		t.Run("TestTpActTrgMove", stest)
	}
	tpActTrgMigrator.Close()
}

func testTpActTrgITConnect(t *testing.T) {
	var err error
	tpActTrgPathIn = path.Join(*dataDir, "conf", "samples", "tutmongo")
	tpActTrgCfgIn, err = config.NewCGRConfigFromPath(tpActTrgPathIn)
	if err != nil {
		t.Fatal(err)
	}
	tpActTrgPathOut = path.Join(*dataDir, "conf", "samples", "tutmysql")
	tpActTrgCfgOut, err = config.NewCGRConfigFromPath(tpActTrgPathOut)
	if err != nil {
		t.Fatal(err)
	}
	storDBIn, err := NewMigratorStorDB(tpActTrgCfgIn.StorDbCfg().Type,
		tpActTrgCfgIn.StorDbCfg().Host, tpActTrgCfgIn.StorDbCfg().Port,
		tpActTrgCfgIn.StorDbCfg().Name, tpActTrgCfgIn.StorDbCfg().User,
		tpActTrgCfgIn.StorDbCfg().Password, tpActTrgCfgIn.GeneralCfg().DBDataEncoding,
		tpActTrgCfgIn.StorDbCfg().StringIndexedFields, tpActTrgCfgIn.StorDbCfg().PrefixIndexedFields,
		tpActTrgCfgIn.StorDbCfg().Opts)
	if err != nil {
		log.Fatal(err)
	}
	storDBOut, err := NewMigratorStorDB(tpActTrgCfgOut.StorDbCfg().Type,
		tpActTrgCfgOut.StorDbCfg().Host, tpActTrgCfgOut.StorDbCfg().Port,
		tpActTrgCfgOut.StorDbCfg().Name, tpActTrgCfgOut.StorDbCfg().User,
		tpActTrgCfgOut.StorDbCfg().Password, tpActTrgCfgOut.GeneralCfg().DBDataEncoding,
		tpActTrgCfgIn.StorDbCfg().StringIndexedFields, tpActTrgCfgIn.StorDbCfg().PrefixIndexedFields,
		tpActTrgCfgOut.StorDbCfg().Opts)
	if err != nil {
		log.Fatal(err)
	}
	tpActTrgMigrator, err = NewMigrator(nil, nil, storDBIn, storDBOut, false, false, false, false)
	if err != nil {
		log.Fatal(err)
	}
}

func testTpActTrgITFlush(t *testing.T) {
	if err := tpActTrgMigrator.storDBIn.StorDB().Flush(
		path.Join(tpActTrgCfgIn.DataFolderPath, "storage", tpActTrgCfgIn.StorDbCfg().Type)); err != nil {
		t.Error(err)
	}

	if err := tpActTrgMigrator.storDBOut.StorDB().Flush(
		path.Join(tpActTrgCfgOut.DataFolderPath, "storage", tpActTrgCfgOut.StorDbCfg().Type)); err != nil {
		t.Error(err)
	}
}

func testTpActTrgITPopulate(t *testing.T) {
	tpActionTriggers = []*utils.TPActionTriggers{
		{
			TPid: "TPAct",
			ID:   "ID",
			ActionTriggers: []*utils.TPActionTrigger{
				{
					Id:                    "ID",
					UniqueID:              "",
					ThresholdType:         "*max_event_counter",
					ThresholdValue:        5,
					Recurrent:             false,
					MinSleep:              "0",
					ExpirationDate:        "",
					ActivationDate:        "",
					BalanceId:             "",
					BalanceType:           "*monetary",
					BalanceDestinationIds: "FS_USERS",
					BalanceWeight:         "",
					BalanceExpirationDate: "",
					BalanceTimingTags:     "",
					BalanceRatingSubject:  "",
					BalanceCategories:     "",
					BalanceSharedGroups:   "",
					BalanceBlocker:        "",
					BalanceDisabled:       "",
					ActionsId:             "LOG_WARNING",
					Weight:                10,
				},
				{
					Id:                    "ID",
					UniqueID:              "",
					ThresholdType:         "*min_balance",
					ThresholdValue:        2,
					Recurrent:             false,
					MinSleep:              "0",
					ExpirationDate:        "",
					ActivationDate:        "",
					BalanceId:             "",
					BalanceType:           "*monetary",
					BalanceDestinationIds: "FS_USERS",
					BalanceWeight:         "",
					BalanceExpirationDate: "",
					BalanceTimingTags:     "",
					BalanceRatingSubject:  "",
					BalanceCategories:     "",
					BalanceSharedGroups:   "",
					BalanceBlocker:        "",
					BalanceDisabled:       "",
					ActionsId:             "LOG_WARNING",
					Weight:                10,
				},
			},
		},
	}
	if err := tpActTrgMigrator.storDBIn.StorDB().SetTPActionTriggers(tpActionTriggers); err != nil {
		t.Error("Error when setting TpActionTriggers ", err.Error())
	}
	currentVersion := engine.CurrentStorDBVersions()
	err := tpActTrgMigrator.storDBIn.StorDB().SetVersions(currentVersion, false)
	if err != nil {
		t.Error("Error when setting version for TpActionTriggers ", err.Error())
	}
}

func testTpActTrgITMove(t *testing.T) {
	err, _ := tpActTrgMigrator.Migrate([]string{utils.MetaTpActionTriggers})
	if err != nil {
		t.Error("Error when migrating TpActionTriggers ", err.Error())
	}
}

func testTpActTrgITCheckData(t *testing.T) {
	result, err := tpActTrgMigrator.storDBOut.StorDB().GetTPActionTriggers(
		tpActionTriggers[0].TPid, tpActionTriggers[0].ID)
	if err != nil {
		t.Error("Error when getting TpActionTriggers ", err.Error())
	}
	if !reflect.DeepEqual(tpActionTriggers[0], result[0]) {
		t.Errorf("Expecting: %+v, received: %+v",
			utils.ToJSON(tpActionTriggers[0]), utils.ToJSON(result[0]))
	}
	result, err = tpActTrgMigrator.storDBIn.StorDB().GetTPActionTriggers(
		tpActionTriggers[0].TPid, tpActionTriggers[0].ID)
	if err != utils.ErrNotFound {
		t.Error(err)
	}
}
