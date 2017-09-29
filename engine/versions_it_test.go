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
package engine

import (
	"flag"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"log"
	"path"
	"testing"
)

var (
	storageDb       Storage
	dataDb          DataDB
	dbtype          string
	loadHistorySize = flag.Int("load_history_size", config.CgrConfig().LoadHistorySize, "Limit the number of records in the load history")
)

var sTestsITVersions = []func(t *testing.T){
	testVersionsFlush,
	testVersion,
	testVersionsFlush,
}

func TestVersionsITMongo(t *testing.T) {
	var err error
	if cfg, err = config.NewCGRConfigFromFolder(path.Join(*dataDir, "conf", "samples", "storage", "mongo")); err != nil {
		t.Fatal(err)
	}
	dataDb, err = ConfigureDataStorage(cfg.DataDbType, cfg.DataDbHost, cfg.DataDbPort, cfg.DataDbName, cfg.DataDbUser, cfg.DataDbPass, cfg.DBDataEncoding, cfg.CacheConfig, *loadHistorySize)
	if err != nil {
		log.Fatal(err)
	}

	if storageDb, err = NewMongoStorage(cfg.StorDBHost, cfg.StorDBPort, cfg.StorDBName,
		cfg.StorDBUser, cfg.StorDBPass, utils.StorDB, cfg.StorDBCDRSIndexes, nil, cfg.LoadHistorySize); err != nil {
		t.Fatal(err)
	}
	dbtype = utils.MONGO
	for _, stest := range sTestsITVersions {
		t.Run("TestVersionsITMongo", stest)
	}
}

func TestVersionsITRedis_MYSQL(t *testing.T) {
	var err error
	if cfg, err = config.NewCGRConfigFromFolder(path.Join(*dataDir, "conf", "samples", "storage", "mysql")); err != nil {
		t.Fatal(err)
	}
	dataDb, err = ConfigureDataStorage(cfg.DataDbType, cfg.DataDbHost, cfg.DataDbPort, cfg.DataDbName, cfg.DataDbUser, cfg.DataDbPass, cfg.DBDataEncoding, cfg.CacheConfig, *loadHistorySize)
	if err != nil {
		log.Fatal(err)
	}

	if storageDb, err = NewMySQLStorage(cfg.StorDBHost, cfg.StorDBPort, cfg.StorDBName,
		cfg.StorDBUser, cfg.StorDBPass, cfg.StorDBMaxOpenConns, cfg.StorDBMaxIdleConns, cfg.StorDBConnMaxLifetime); err != nil {
		t.Fatal(err)
	}
	dbtype = utils.REDIS
	for _, stest := range sTestsITVersions {
		t.Run("TestVersionsITRedis", stest)
	}
}

func TestVersionsITPostgres(t *testing.T) {
	var err error
	if cfg, err = config.NewCGRConfigFromFolder(path.Join(*dataDir, "conf", "samples", "storage", "postgres")); err != nil {
		t.Fatal(err)
	}
	dataDb, err = ConfigureDataStorage(cfg.DataDbType, cfg.DataDbHost, cfg.DataDbPort, cfg.DataDbName, cfg.DataDbUser, cfg.DataDbPass, cfg.DBDataEncoding, cfg.CacheConfig, *loadHistorySize)
	if err != nil {
		log.Fatal(err)
	}
	if storageDb, err = NewPostgresStorage(cfg.StorDBHost, cfg.StorDBPort, cfg.StorDBName,
		cfg.StorDBUser, cfg.StorDBPass, cfg.StorDBMaxOpenConns, cfg.StorDBMaxIdleConns, cfg.StorDBConnMaxLifetime); err != nil {
		t.Fatal(err)
	}
	dbtype = utils.REDIS
	for _, stest := range sTestsITVersions {
		t.Run("TestMigratorITPostgres", stest)
	}
}

func testVersionsFlush(t *testing.T) {
	err := dataDb.Flush("")
	if err != nil {
		t.Error("Error when flushing Mongo ", err.Error())
	}
	if err := storageDb.Flush(path.Join(cfg.DataFolderPath, "storage", cfg.StorDBType)); err != nil {
		t.Error(err)
	}
}

func testVersion(t *testing.T) {
	var test string
	var currentVersion Versions
	var testVersion Versions
	storType := dataDb.GetStorageType()
	switch storType {
	case utils.MONGO, utils.MAPSTOR:
		currentVersion = Versions{utils.Accounts: 2, utils.Actions: 2, utils.ActionTriggers: 2, utils.ActionPlans: 2, utils.SharedGroups: 2, utils.COST_DETAILS: 2}
		testVersion = Versions{utils.Accounts: 1, utils.Actions: 2, utils.ActionTriggers: 2, utils.ActionPlans: 2, utils.SharedGroups: 2, utils.COST_DETAILS: 2}
		test = "Migration needed: please backup cgr data and run : <cgr-migrator -migrate=*accounts>"
	case utils.REDIS:
		currentVersion = CurrentDataDBVersions()
		testVersion = Versions{utils.Accounts: 1, utils.Actions: 2, utils.ActionTriggers: 2, utils.ActionPlans: 2, utils.SharedGroups: 2}
		test = "Migration needed: please backup cgr data and run : <cgr-migrator -migrate=*accounts>"
	}

	//dataDB
	storType = dataDb.GetStorageType()

	log.Print("storType:", storType)

	if _, rcvErr := dataDb.GetVersions(utils.TBLVersions); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := CheckVersions(dataDb); err != nil {
		t.Error(err)
	}
	if rcv, err := dataDb.GetVersions(utils.TBLVersions); err != nil {
		t.Error(err)
	} else if len(currentVersion) != len(rcv) {
		t.Errorf("Expecting: %v, received: %v", currentVersion, rcv)
	}
	if err = dataDb.RemoveVersions(currentVersion); err != nil {
		t.Error(err)
	}
	if _, rcvErr := dataDb.GetVersions(utils.TBLVersions); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := dataDb.SetVersions(testVersion, false); err != nil {
		t.Error(err)
	}
	if err := CheckVersions(dataDb); err.Error() != test {
		t.Error(err)
	}
	if err = dataDb.RemoveVersions(testVersion); err != nil {
		t.Error(err)
	}

	storType = storageDb.GetStorageType()
	switch storType {
	case utils.MONGO, utils.MAPSTOR:
		currentVersion = Versions{utils.Accounts: 2, utils.Actions: 2, utils.ActionTriggers: 2, utils.ActionPlans: 2, utils.SharedGroups: 2, utils.COST_DETAILS: 2}
		testVersion = Versions{utils.Accounts: 1, utils.Actions: 2, utils.ActionTriggers: 2, utils.ActionPlans: 2, utils.SharedGroups: 2, utils.COST_DETAILS: 2}
		test = "Migration needed: please backup cgr data and run : <cgr-migrator -migrate=*accounts>"
	case utils.POSTGRES, utils.MYSQL:
		currentVersion = CurrentStorDBVersions()
		testVersion = Versions{utils.COST_DETAILS: 1}
		test = "Migration needed: please backup cgr data and run : <cgr-migrator -migrate=*cost_details>"
	}
	//storageDb
	storType = storageDb.GetStorageType()

	log.Print("storType:", storType)

	// if _, rcvErr := storageDb.GetVersions(utils.TBLVersions); rcvErr != utils.ErrNotFound {
	// 	t.Error(rcvErr)
	// }
	if err := CheckVersions(storageDb); err != nil {
		t.Error(err)
	}
	if rcv, err := storageDb.GetVersions(utils.TBLVersions); err != nil {
		t.Error(err)
	} else if len(currentVersion) != len(rcv) {
		t.Errorf("Expecting: %v, received: %v", currentVersion, rcv)
	}
	if err = storageDb.RemoveVersions(currentVersion); err != nil {
		t.Error(err)
	}
	if _, rcvErr := storageDb.GetVersions(utils.TBLVersions); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := storageDb.SetVersions(testVersion, false); err != nil {
		t.Error(err)
	}
	if err := CheckVersions(storageDb); err.Error() != test {
		t.Error(err)
	}
	if err = storageDb.RemoveVersions(testVersion); err != nil {
		t.Error(err)
	}

}
