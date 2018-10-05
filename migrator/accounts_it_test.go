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
	"flag"
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
	accPathIn   string
	accPathOut  string
	accCfgIn    *config.CGRConfig
	accCfgOut   *config.CGRConfig
	accMigrator *Migrator
	accAction   string
	dataDir     = flag.String("data_dir", "/usr/share/cgrates", "CGR data dir path here")
)

var sTestsAccIT = []func(t *testing.T){
	testAccITConnect,
	testAccITFlush,
	testAccITMigrateAndMove,
}

func TestAccountMigrateITRedis(t *testing.T) {
	var err error
	accPathIn = path.Join(*dataDir, "conf", "samples", "tutmysql")
	accCfgIn, err = config.NewCGRConfigFromFolder(accPathIn)
	if err != nil {
		t.Fatal(err)
	}
	accCfgOut, err = config.NewCGRConfigFromFolder(accPathIn)
	if err != nil {
		t.Fatal(err)
	}
	accAction = utils.Migrate
	for _, stest := range sTestsAccIT {
		t.Run("TestAccountITMigrateRedis", stest)
	}
}

func TestAccountMigrateITMongo(t *testing.T) {
	var err error
	accPathIn = path.Join(*dataDir, "conf", "samples", "tutmongo")
	accCfgIn, err = config.NewCGRConfigFromFolder(accPathIn)
	if err != nil {
		t.Fatal(err)
	}
	accCfgOut, err = config.NewCGRConfigFromFolder(accPathIn)
	if err != nil {
		t.Fatal(err)
	}
	accAction = utils.Migrate
	for _, stest := range sTestsAccIT {
		t.Run("TestAccountITMigrateMongo", stest)
	}
}

func TestAccountITMove(t *testing.T) {
	var err error
	accPathIn = path.Join(*dataDir, "conf", "samples", "tutmongo")
	accCfgIn, err = config.NewCGRConfigFromFolder(accPathIn)
	if err != nil {
		t.Fatal(err)
	}
	accPathOut = path.Join(*dataDir, "conf", "samples", "tutmysql")
	accCfgOut, err = config.NewCGRConfigFromFolder(accPathOut)
	if err != nil {
		t.Fatal(err)
	}
	accAction = utils.Move
	for _, stest := range sTestsAccIT {
		t.Run("TestAccountITMove", stest)
	}
}

func TestAccountITMoveEncoding(t *testing.T) {
	var err error
	accPathIn = path.Join(*dataDir, "conf", "samples", "tutmongo")
	accCfgIn, err = config.NewCGRConfigFromFolder(accPathIn)
	if err != nil {
		t.Fatal(err)
	}
	accPathOut = path.Join(*dataDir, "conf", "samples", "tutmongojson")
	accCfgOut, err = config.NewCGRConfigFromFolder(accPathOut)
	if err != nil {
		t.Fatal(err)
	}
	accAction = utils.Move
	for _, stest := range sTestsAccIT {
		t.Run("TestAccountITMove", stest)
	}
}

func TestAccountITMoveEncoding2(t *testing.T) {
	var err error
	accPathIn = path.Join(*dataDir, "conf", "samples", "tutmysql")
	accCfgIn, err = config.NewCGRConfigFromFolder(accPathIn)
	if err != nil {
		t.Fatal(err)
	}
	accPathOut = path.Join(*dataDir, "conf", "samples", "tutmysqljson")
	accCfgOut, err = config.NewCGRConfigFromFolder(accPathOut)
	if err != nil {
		t.Fatal(err)
	}
	accAction = utils.Move
	for _, stest := range sTestsAccIT {
		t.Run("TestAccountITMove", stest)
	}
}

func testAccITConnect(t *testing.T) {
	dataDBIn, err := NewMigratorDataDB(accCfgIn.DataDbCfg().DataDbType,
		accCfgIn.DataDbCfg().DataDbHost, accCfgIn.DataDbCfg().DataDbPort,
		accCfgIn.DataDbCfg().DataDbName, accCfgIn.DataDbCfg().DataDbUser,
		accCfgIn.DataDbCfg().DataDbPass, accCfgIn.GeneralCfg().DBDataEncoding,
		config.CgrConfig().CacheCfg(), "")
	if err != nil {
		log.Fatal(err)
	}
	dataDBOut, err := NewMigratorDataDB(accCfgOut.DataDbCfg().DataDbType,
		accCfgOut.DataDbCfg().DataDbHost, accCfgOut.DataDbCfg().DataDbPort,
		accCfgOut.DataDbCfg().DataDbName, accCfgOut.DataDbCfg().DataDbUser,
		accCfgOut.DataDbCfg().DataDbPass, accCfgOut.GeneralCfg().DBDataEncoding,
		config.CgrConfig().CacheCfg(), "")
	if err != nil {
		log.Fatal(err)
	}
	accMigrator, err = NewMigrator(dataDBIn, dataDBOut,
		nil, nil,
		false, false, false)
	if err != nil {
		log.Fatal(err)
	}
}

func testAccITFlush(t *testing.T) {
	accMigrator.dmOut.DataManager().DataDB().Flush("")
	if err := engine.SetDBVersions(accMigrator.dmOut.DataManager().DataDB()); err != nil {
		t.Error("Error  ", err.Error())
	}
	accMigrator.dmIN.DataManager().DataDB().Flush("")
	if err := engine.SetDBVersions(accMigrator.dmIN.DataManager().DataDB()); err != nil {
		t.Error("Error  ", err.Error())
	}
}

func testAccITMigrateAndMove(t *testing.T) {
	timingSlice := []*engine.RITiming{
		{
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
			utils.DATA:  {v1b},
			utils.VOICE: {v1b},
			utils.MONETARY: {
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
			utils.DATA:     {v2d},
			utils.VOICE:    {v2b},
			utils.MONETARY: {m2}},
		UnitCounters:   engine.UnitCounters{},
		ActionTriggers: engine.ActionTriggers{},
	}
	switch accAction {
	case utils.Migrate:
		// set v1Account
		err := accMigrator.dmIN.setV1Account(v1Acc)
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
		err = accMigrator.dmOut.DataManager().DataDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for Accounts ", err.Error())
		}
		//check if version was set correctly
		if vrs, err := accMigrator.dmOut.DataManager().DataDB().GetVersions(""); err != nil {
			t.Error(err)
		} else if vrs[utils.Accounts] != 1 {
			t.Errorf("Unexpected version returned: %d", vrs[utils.Accounts])
		}
		//migrate account
		err, _ = accMigrator.Migrate([]string{utils.MetaAccounts})
		if err != nil {
			t.Error("Error when migrating Accounts ", err.Error())
		}
		//check if version was updated
		if vrs, err := accMigrator.dmOut.DataManager().DataDB().GetVersions(""); err != nil {
			t.Error(err)
		} else if vrs[utils.Accounts] != 3 {
			t.Errorf("Unexpected version returned: %d", vrs[utils.Accounts])
		}
		//check if account was migrate correctly
		result, err := accMigrator.dmOut.DataManager().DataDB().GetAccount(testAccount.ID)
		if err != nil {
			t.Error("Error when getting Accounts ", err.Error())
		}
		if !reflect.DeepEqual(testAccount, result) {
			t.Errorf("Expecting: %+v, received: %+v", testAccount, result)
		}
		//check if old account was deleted
		if _, err = accMigrator.dmIN.getv1Account(); err != utils.ErrNoMoreData {
			t.Error("Error should be not found : ", err)
		}
	case utils.Move:
		//set an account in dmIN
		if err := accMigrator.dmIN.DataManager().DataDB().SetAccount(testAccount); err != nil {
			t.Error(err)
		}
		//set versions for account
		currentVersion := engine.CurrentDataDBVersions()
		err := accMigrator.dmOut.DataManager().DataDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for Accounts ", err.Error())
		}
		//migrate accounts
		err, _ = accMigrator.Migrate([]string{utils.MetaAccounts})
		if err != nil {
			t.Error("Error when accMigratorrating Accounts ", err.Error())
		}
		//check if account was migrate correctly
		result, err := accMigrator.dmOut.DataManager().DataDB().GetAccount(testAccount.ID)
		if err != nil {
			t.Error(err)
		}
		if !reflect.DeepEqual(testAccount, result) {
			t.Errorf("Expecting: %+v, received: %+v", testAccount, result)
		}
		//check if old account was deleted
		result, err = accMigrator.dmIN.DataManager().DataDB().GetAccount(testAccount.ID)
		if err != utils.ErrNotFound {
			t.Error(err)
		}
	}
}
