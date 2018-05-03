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
	accPathIn       string
	accPathOut      string
	accCfgIn        *config.CGRConfig
	accCfgOut       *config.CGRConfig
	accMigrator     *Migrator
	accAction       string
	dataDir         = flag.String("data_dir", "/usr/share/cgrates", "CGR data dir path here")
	loadHistorySize = flag.Int("load_history_size", config.CgrConfig().LoadHistorySize, "Limit the number of records in the load history")
)

var sTestsAccIT = []func(t *testing.T){
	testAccITFlush,
	testAccITMigrateAndMove,
}

func TestAccountITRedisConnection(t *testing.T) {
	var err error
	accPathIn = path.Join(*dataDir, "conf", "samples", "tutmongo")
	accCfgIn, err = config.NewCGRConfigFromFolder(accPathIn)
	if err != nil {
		t.Fatal(err)
	}
	dataDBIn, err := engine.ConfigureDataStorage(accCfgIn.DataDbType,
		accCfgIn.DataDbHost, accCfgIn.DataDbPort, accCfgIn.DataDbName,
		accCfgIn.DataDbUser, accCfgIn.DataDbPass, accCfgIn.DBDataEncoding,
		config.CgrConfig().CacheCfg(), *loadHistorySize)
	if err != nil {
		log.Fatal(err)
	}
	dataDBOut, err := engine.ConfigureDataStorage(accCfgIn.DataDbType,
		accCfgIn.DataDbHost, accCfgIn.DataDbPort, accCfgIn.DataDbName,
		accCfgIn.DataDbUser, accCfgIn.DataDbPass, accCfgIn.DBDataEncoding,
		config.CgrConfig().CacheCfg(), *loadHistorySize)
	if err != nil {
		log.Fatal(err)
	}
	oldDataDB, err := ConfigureV1DataStorage(accCfgIn.DataDbType,
		accCfgIn.DataDbHost, accCfgIn.DataDbPort, accCfgIn.DataDbName,
		accCfgIn.DataDbUser, accCfgIn.DataDbPass, accCfgIn.DBDataEncoding)
	if err != nil {
		log.Fatal(err)
	}
	accMigrator, err = NewMigrator(dataDBIn, dataDBOut, accCfgIn.DataDbType,
		accCfgIn.DBDataEncoding, nil, nil, accCfgIn.StorDBType, oldDataDB,
		accCfgIn.DataDbType, accCfgIn.DBDataEncoding, nil, accCfgIn.StorDBType,
		false, false, false, false, false)
	if err != nil {
		log.Fatal(err)
	}
}

func TestAccountITRedis(t *testing.T) {
	accAction = utils.Migrate
	for _, stest := range sTestsAccIT {
		t.Run("TestAccountITMigrateRedis", stest)
	}
}

func TestAccountITMongoConnection(t *testing.T) {
	var err error
	accPathIn = path.Join(*dataDir, "conf", "samples", "tutmysql")
	accCfgIn, err = config.NewCGRConfigFromFolder(accPathIn)
	if err != nil {
		t.Fatal(err)
	}
	dataDBIn, err := engine.ConfigureDataStorage(accCfgIn.DataDbType,
		accCfgIn.DataDbHost, accCfgIn.DataDbPort, accCfgIn.DataDbName,
		accCfgIn.DataDbUser, accCfgIn.DataDbPass, accCfgIn.DBDataEncoding,
		config.CgrConfig().CacheCfg(), *loadHistorySize)
	if err != nil {
		log.Fatal(err)
	}
	dataDBOut, err := engine.ConfigureDataStorage(accCfgIn.DataDbType,
		accCfgIn.DataDbHost, accCfgIn.DataDbPort, accCfgIn.DataDbName,
		accCfgIn.DataDbUser, accCfgIn.DataDbPass, accCfgIn.DBDataEncoding,
		config.CgrConfig().CacheCfg(), *loadHistorySize)
	if err != nil {
		log.Fatal(err)
	}
	oldDataDB, err := ConfigureV1DataStorage(accCfgIn.DataDbType,
		accCfgIn.DataDbHost, accCfgIn.DataDbPort, accCfgIn.DataDbName,
		accCfgIn.DataDbUser, accCfgIn.DataDbPass, accCfgIn.DBDataEncoding)
	if err != nil {
		log.Fatal(err)
	}
	accMigrator, err = NewMigrator(dataDBIn, dataDBOut, accCfgIn.DataDbType,
		accCfgIn.DBDataEncoding, nil, nil, accCfgIn.StorDBType, oldDataDB,
		accCfgIn.DataDbType, accCfgIn.DBDataEncoding, nil, accCfgIn.StorDBType,
		false, false, false, false, false)
	if err != nil {
		log.Fatal(err)
	}

}

func TestAccountITMongo(t *testing.T) {
	accAction = utils.Migrate
	for _, stest := range sTestsAccIT {
		t.Run("TestAccountITMigrateMongo", stest)
	}
}

func TestAccountITMoveConnection(t *testing.T) {
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
	dataDBIn, err := engine.ConfigureDataStorage(accCfgIn.DataDbType,
		accCfgIn.DataDbHost, accCfgIn.DataDbPort, accCfgIn.DataDbName,
		accCfgIn.DataDbUser, accCfgIn.DataDbPass, accCfgIn.DBDataEncoding,
		config.CgrConfig().CacheCfg(), *loadHistorySize)
	if err != nil {
		log.Fatal(err)
	}
	dataDBOut, err := engine.ConfigureDataStorage(accCfgOut.DataDbType,
		accCfgOut.DataDbHost, accCfgOut.DataDbPort, accCfgOut.DataDbName,
		accCfgOut.DataDbUser, accCfgOut.DataDbPass, accCfgOut.DBDataEncoding,
		config.CgrConfig().CacheCfg(), *loadHistorySize)
	if err != nil {
		log.Fatal(err)
	}
	oldDataDB, err := ConfigureV1DataStorage(accCfgOut.DataDbType,
		accCfgOut.DataDbHost, accCfgOut.DataDbPort, accCfgOut.DataDbName,
		accCfgOut.DataDbUser, accCfgOut.DataDbPass, accCfgOut.DBDataEncoding)
	if err != nil {
		log.Fatal(err)
	}
	accMigrator, err = NewMigrator(dataDBIn, dataDBOut, accCfgIn.DataDbType,
		accCfgIn.DBDataEncoding, nil, nil, accCfgIn.StorDBType, oldDataDB,
		accCfgIn.DataDbType, accCfgIn.DBDataEncoding, nil, accCfgIn.StorDBType,
		false, false, false, false, false)
	if err != nil {
		log.Fatal(err)
	}
}

func TestAccountITMove(t *testing.T) {
	accAction = utils.Move
	for _, stest := range sTestsAccIT {
		t.Run("TestAccountITMove", stest)
	}
}

func testAccITFlush(t *testing.T) {
	accMigrator.dmOut.DataDB().Flush("")
	if err := engine.SetDBVersions(accMigrator.dmOut.DataDB()); err != nil {
		t.Error("Error  ", err.Error())
	}
}

func testAccITMigrateAndMove(t *testing.T) {
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
		ActionTriggers: engine.ActionTriggers{}}
	switch accAction {
	case utils.Migrate:
		err := accMigrator.oldDataDB.setV1Account(v1Acc)
		if err != nil {
			t.Error("Error when setting v1 Accounts ", err.Error())
		}
		currentVersion := engine.Versions{
			utils.StatS:          2,
			utils.Thresholds:     2,
			utils.Accounts:       1,
			utils.Actions:        2,
			utils.ActionTriggers: 2,
			utils.ActionPlans:    2,
			utils.SharedGroups:   2}
		err = accMigrator.dmOut.DataDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for Accounts ", err.Error())
		}
		err, _ = accMigrator.Migrate([]string{utils.MetaAccounts})
		if err != nil {
			t.Error("Error when migrating Accounts ", err.Error())
		}
		result, err := accMigrator.dmOut.DataDB().GetAccount(testAccount.ID)
		if err != nil {
			t.Error("Error when getting Accounts ", err.Error())
		}
		if !reflect.DeepEqual(testAccount.BalanceMap["*voice"][0], result.BalanceMap["*voice"][0]) {
			t.Errorf("Expecting: %+v, received: %+v", testAccount.BalanceMap["*voice"][0], result.BalanceMap["*voice"][0])
		} else if !reflect.DeepEqual(testAccount, result) {
			t.Errorf("Expecting: %+v, received: %+v", testAccount, result)
		}
	case utils.Move:
		if err := accMigrator.dmIN.DataDB().SetAccount(testAccount); err != nil {
			log.Print("GOT ERR DMIN", err)
		}
		currentVersion := engine.CurrentDataDBVersions()
		err := accMigrator.dmOut.DataDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for Accounts ", err.Error())
		}
		err, _ = accMigrator.Migrate([]string{utils.MetaAccounts})
		if err != nil {
			t.Error("Error when accMigratorrating Accounts ", err.Error())
		}
		result, err := accMigrator.dmOut.DataDB().GetAccount(testAccount.ID)
		if err != nil {
			t.Error(err)
		}
		if !reflect.DeepEqual(testAccount, result) {
			t.Errorf("Expecting: %+v, received: %+v", testAccount, result)
		}
		result, err = accMigrator.dmIN.DataDB().GetAccount(testAccount.ID)
		if err != utils.ErrNotFound {
			t.Error(err)
		}
	}

}
