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
	tpActTrgPathIn = path.Join(*utils.DataDir, "conf", "samples", "tutmongo")
	tpActTrgCfgIn, err = config.NewCGRConfigFromPath(tpActTrgPathIn)
	if err != nil {
		t.Fatal(err)
	}
	tpActTrgPathOut = path.Join(*utils.DataDir, "conf", "samples", "tutmysql")
	tpActTrgCfgOut, err = config.NewCGRConfigFromPath(tpActTrgPathOut)
	if err != nil {
		t.Fatal(err)
	}
	storDBIn, err := NewMigratorStorDB(tpActTrgCfgIn.StorDbCfg().Type,
		tpActTrgCfgIn.StorDbCfg().Host, tpActTrgCfgIn.StorDbCfg().Port,
		tpActTrgCfgIn.StorDbCfg().Name, tpActTrgCfgIn.StorDbCfg().User,
		tpActTrgCfgIn.StorDbCfg().Password, tpActTrgCfgIn.GeneralCfg().DBDataEncoding, tpActTrgCfgIn.StorDbCfg().SSLMode,
		tpActTrgCfgIn.StorDbCfg().MaxOpenConns, tpActTrgCfgIn.StorDbCfg().MaxIdleConns,
		tpActTrgCfgIn.StorDbCfg().ConnMaxLifetime, tpActTrgCfgIn.StorDbCfg().StringIndexedFields,
		tpActTrgCfgIn.StorDbCfg().PrefixIndexedFields, tpActTrgCfgIn.StorDbCfg().Items)
	if err != nil {
		log.Fatal(err)
	}
	storDBOut, err := NewMigratorStorDB(tpActTrgCfgOut.StorDbCfg().Type,
		tpActTrgCfgOut.StorDbCfg().Host, tpActTrgCfgOut.StorDbCfg().Port,
		tpActTrgCfgOut.StorDbCfg().Name, tpActTrgCfgOut.StorDbCfg().User,
		tpActTrgCfgOut.StorDbCfg().Password, tpActTrgCfgOut.GeneralCfg().DBDataEncoding, tpActTrgCfgIn.StorDbCfg().SSLMode,
		tpActTrgCfgIn.StorDbCfg().MaxOpenConns, tpActTrgCfgIn.StorDbCfg().MaxIdleConns,
		tpActTrgCfgIn.StorDbCfg().ConnMaxLifetime, tpActTrgCfgIn.StorDbCfg().StringIndexedFields,
		tpActTrgCfgIn.StorDbCfg().PrefixIndexedFields, tpActTrgCfgOut.StorDbCfg().Items)
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
		path.Join(tpActTrgCfgIn.DataFolderPath, "storage", dbPath(tpActTrgCfgIn.StorDbCfg().Type))); err != nil {
		t.Error(err)
	}

	if err := tpActTrgMigrator.storDBOut.StorDB().Flush(
		path.Join(tpActTrgCfgOut.DataFolderPath, "storage", dbPath(tpActTrgCfgOut.StorDbCfg().Type))); err != nil {
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
	err := tpActTrgMigrator.storDBOut.StorDB().SetVersions(currentVersion, false)
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
