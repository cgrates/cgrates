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
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	acc2PathIn   string
	acc2PathOut  string
	acc2CfgIn    *config.CGRConfig
	acc2CfgOut   *config.CGRConfig
	acc2Migrator *Migrator
)

var sTestsAcc2IT = []func(t *testing.T){
	testAcc2ITConnect,
	testAcc2ITFlush,
	testAcc2ITMigrate,
}

func TestAccMigrateWithInternal(t *testing.T) {
	var err error
	acc2PathIn = path.Join(*dataDir, "conf", "samples", "migwithinternal")
	acc2CfgIn, err = config.NewCGRConfigFromFolder(acc2PathIn)
	if err != nil {
		t.Fatal(err)
	}
	acc2CfgOut, err = config.NewCGRConfigFromFolder(acc2PathIn)
	if err != nil {
		t.Fatal(err)
	}
	for _, stest := range sTestsAcc2IT {
		t.Run("TestAccMigrateWithInternal", stest)
	}
}

func testAcc2ITConnect(t *testing.T) {
	dataDBIn, err := NewMigratorDataDB(acc2CfgIn.DataDbType,
		acc2CfgIn.DataDbHost, acc2CfgIn.DataDbPort, acc2CfgIn.DataDbName,
		acc2CfgIn.DataDbUser, acc2CfgIn.DataDbPass, acc2CfgIn.DBDataEncoding,
		config.CgrConfig().CacheCfg(), *loadHistorySize)
	if err != nil {
		t.Error(err)
	}
	dataDBOut, err := NewMigratorDataDB(acc2CfgOut.DataDbType,
		acc2CfgOut.DataDbHost, acc2CfgOut.DataDbPort, acc2CfgOut.DataDbName,
		acc2CfgOut.DataDbUser, acc2CfgOut.DataDbPass, acc2CfgOut.DBDataEncoding,
		config.CgrConfig().CacheCfg(), *loadHistorySize)
	if err != nil {
		t.Error(err)
	}

	storDBIn, err := NewMigratorStorDB(acc2CfgIn.StorDBType,
		acc2CfgIn.StorDBHost, acc2CfgIn.StorDBPort, acc2CfgIn.StorDBName,
		acc2CfgIn.StorDBUser, acc2CfgIn.StorDBPass, acc2CfgIn.StorDBMaxOpenConns,
		acc2CfgIn.StorDBMaxIdleConns, acc2CfgIn.StorDBConnMaxLifetime, acc2CfgIn.StorDBCDRSIndexes)
	if err != nil {
		t.Error(err)
	}
	storDBOut, err := NewMigratorStorDB(acc2CfgOut.StorDBType,
		acc2CfgOut.StorDBHost, acc2CfgOut.StorDBPort, acc2CfgOut.StorDBName,
		acc2CfgOut.StorDBUser, acc2CfgOut.StorDBPass, acc2CfgOut.StorDBMaxOpenConns,
		acc2CfgOut.StorDBMaxIdleConns, acc2CfgOut.StorDBConnMaxLifetime, acc2CfgOut.StorDBCDRSIndexes)
	if err != nil {
		t.Error(err)
	}
	acc2Migrator, err = NewMigrator(dataDBIn, dataDBOut,
		storDBIn, storDBOut,
		false, false, false)
	if err != nil {
		log.Fatal(err)
	}
}

func testAcc2ITFlush(t *testing.T) {
	acc2Migrator.dmOut.DataManager().DataDB().Flush("")
	if err := engine.SetDBVersions(acc2Migrator.dmOut.DataManager().DataDB()); err != nil {
		t.Error("Error  ", err.Error())
	}
	acc2Migrator.dmIN.DataManager().DataDB().Flush("")
	if err := engine.SetDBVersions(acc2Migrator.dmIN.DataManager().DataDB()); err != nil {
		t.Error("Error  ", err.Error())
	}
	if acc2Migrator.dmOut.DataManager().DataDB().GetStorageType() != utils.REDIS {
		t.Errorf("Unexpected datadb type : %+v", acc2Migrator.dmOut.DataManager().DataDB().GetStorageType())
	}
	if acc2Migrator.storDBIn.StorDB().GetStorageType() != utils.MAPSTOR {
		t.Errorf("Unexpected datadb type : %+v", acc2Migrator.storDBIn.StorDB().GetStorageType())
	}
	if acc2Migrator.storDBOut.StorDB().GetStorageType() != utils.MAPSTOR {
		t.Errorf("Unexpected datadb type : %+v", acc2Migrator.storDBOut.StorDB().GetStorageType())
	}
}

func testAcc2ITMigrate(t *testing.T) {
	timingSlice := []*engine.RITiming{
		&engine.RITiming{
			Years:     utils.Years{},
			Months:    utils.Months{},
			MonthDays: utils.MonthDays{},
			WeekDays:  utils.WeekDays{},
		},
	}
	v1b := &v1Balance{
		Value:          100000,
		Weight:         10,
		DestinationIds: "NAT",
		ExpirationDate: time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC),
		Timings:        timingSlice,
	}
	v1Acc := &v1Account{
		Id: "*OUT:CUSTOMER_1:rif",
		BalanceMap: map[string]v1BalanceChain{
			utils.DATA:  v1BalanceChain{v1b},
			utils.VOICE: v1BalanceChain{v1b},
			utils.MONETARY: v1BalanceChain{
				&v1Balance{Value: 21,
					ExpirationDate: time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC),
					Timings:        timingSlice}}}}

	v2d := &engine.Balance{
		Uuid: "", ID: "",
		Value:          100000,
		Directions:     utils.StringMap{"*OUT": true},
		ExpirationDate: time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC),
		Weight:         10,
		DestinationIDs: utils.StringMap{"NAT": true},
		RatingSubject:  "",
		Categories:     utils.NewStringMap(),
		SharedGroups:   utils.NewStringMap(),
		Timings:        timingSlice,
		TimingIDs:      utils.NewStringMap(""),
		Factor:         engine.ValueFactor{}}
	v2b := &engine.Balance{
		Uuid: "", ID: "",
		Value:          0.0001,
		Directions:     utils.StringMap{"*OUT": true},
		ExpirationDate: time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC),
		Weight:         10,
		DestinationIDs: utils.StringMap{"NAT": true},
		RatingSubject:  "",
		Categories:     utils.NewStringMap(),
		SharedGroups:   utils.NewStringMap(),
		Timings:        timingSlice,
		TimingIDs:      utils.NewStringMap(""),
		Factor:         engine.ValueFactor{}}
	m2 := &engine.Balance{
		Uuid:           "",
		ID:             "",
		Value:          21,
		Directions:     utils.StringMap{"*OUT": true},
		ExpirationDate: time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC),
		DestinationIDs: utils.NewStringMap(""),
		RatingSubject:  "",
		Categories:     utils.NewStringMap(),
		SharedGroups:   utils.NewStringMap(),
		Timings:        timingSlice,
		TimingIDs:      utils.NewStringMap(""),
		Factor:         engine.ValueFactor{}}
	testAccount := &engine.Account{
		ID: "CUSTOMER_1:rif",
		BalanceMap: map[string]engine.Balances{
			utils.DATA:     engine.Balances{v2d},
			utils.VOICE:    engine.Balances{v2b},
			utils.MONETARY: engine.Balances{m2}},
		UnitCounters:   engine.UnitCounters{},
		ActionTriggers: engine.ActionTriggers{},
	}
	// set v1Account
	err := acc2Migrator.dmIN.setV1Account(v1Acc)
	if err != nil {
		t.Error("Error when setting v1 Accounts ", err.Error())
	}
	//set version for account : 1
	currentVersion := engine.Versions{
		utils.StatS:          2,
		utils.Thresholds:     2,
		utils.Accounts:       1,
		utils.Actions:        2,
		utils.ActionTriggers: 2,
		utils.ActionPlans:    2,
		utils.SharedGroups:   2}
	err = acc2Migrator.dmOut.DataManager().DataDB().SetVersions(currentVersion, false)
	if err != nil {
		t.Error("Error when setting version for Accounts ", err.Error())
	}
	//check if version was set correctly
	if vrs, err := acc2Migrator.dmOut.DataManager().DataDB().GetVersions(""); err != nil {
		t.Error(err)
	} else if vrs[utils.Accounts] != 1 {
		t.Errorf("Unexpected version returned: %d", vrs[utils.Accounts])
	}
	//migrate account
	err, _ = acc2Migrator.Migrate([]string{utils.MetaAccounts})
	if err != nil {
		t.Error("Error when migrating Accounts ", err.Error())
	}
	//check if version was updated
	if vrs, err := acc2Migrator.dmOut.DataManager().DataDB().GetVersions(""); err != nil {
		t.Error(err)
	} else if vrs[utils.Accounts] != 3 {
		t.Errorf("Unexpected version returned: %d", vrs[utils.Accounts])
	}
	//check if account was migrate correctly
	result, err := acc2Migrator.dmOut.DataManager().DataDB().GetAccount(testAccount.ID)
	if err != nil {
		t.Error("Error when getting Accounts ", err.Error())
	}
	if !reflect.DeepEqual(testAccount, result) {
		t.Errorf("Expecting: %+v, received: %+v", testAccount, result)
	}
	//check if old account was deleted
	if _, err = acc2Migrator.dmIN.getv1Account(); err != utils.ErrNoMoreData {
		t.Error("Error should be not found : ", err)
	}
}
