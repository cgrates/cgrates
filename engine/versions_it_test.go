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
	TestVersion,
	testVersionsFlush,
}

func TestVersionsITMongoConnect(t *testing.T) {
	cdrsMongoCfgPath := path.Join(*dataDir, "conf", "samples", "tutmongo")
	cfg, err := config.NewCGRConfigFromFolder(cdrsMongoCfgPath)
	if err != nil {
		t.Fatal(err)
	}
	dataDB, err := ConfigureDataStorage(cfg.DataDbType, cfg.DataDbHost, cfg.DataDbPort, cfg.DataDbName, cfg.DataDbUser, cfg.DataDbPass, cfg.DBDataEncoding, cfg.CacheConfig, *loadHistorySize)
	if err != nil {
		log.Fatal(err)
	}
	storDB, err := ConfigureStorStorage(cfg.StorDBType, cfg.StorDBHost, cfg.StorDBPort, cfg.StorDBName, cfg.StorDBUser, cfg.StorDBPass, cfg.DBDataEncoding,
		config.CgrConfig().StorDBMaxOpenConns, config.CgrConfig().StorDBMaxIdleConns, config.CgrConfig().StorDBConnMaxLifetime, config.CgrConfig().StorDBCDRSIndexes)
	if err != nil {
		log.Fatal(err)
	}
	storageDb = storDB
	dataDb = dataDB
}

func TestVersionsITMongo(t *testing.T) {
	dbtype = utils.MONGO
	for _, stest := range sTestsITVersions {
		t.Run("TestVersionsITMongo", stest)
	}
}

func TestVersionsITRedisConnect(t *testing.T) {
	cdrsMysqlCfgPath := path.Join(*dataDir, "conf", "samples", "tutmysql")
	cfg, err := config.NewCGRConfigFromFolder(cdrsMysqlCfgPath)
	if err != nil {
		t.Fatal(err)
	}
	dataDB, err := ConfigureDataStorage(cfg.DataDbType, cfg.DataDbHost, cfg.DataDbPort, cfg.DataDbName, cfg.DataDbUser, cfg.DataDbPass, cfg.DBDataEncoding, cfg.CacheConfig, *loadHistorySize)
	if err != nil {
		log.Fatal(err)
	}
	storDB, err := ConfigureStorStorage(cfg.StorDBType, cfg.StorDBHost, cfg.StorDBPort, cfg.StorDBName, cfg.StorDBUser, cfg.StorDBPass, cfg.DBDataEncoding,
		config.CgrConfig().StorDBMaxOpenConns, config.CgrConfig().StorDBMaxIdleConns, config.CgrConfig().StorDBConnMaxLifetime, config.CgrConfig().StorDBCDRSIndexes)
	if err != nil {
		log.Fatal(err)
	}
	storageDb = storDB
	dataDb = dataDB
}

func TestVersionsITRedis(t *testing.T) {
	dbtype = utils.REDIS
	for _, stest := range sTestsITVersions {
		t.Run("TestVersionsITRedis", stest)
	}
}

func TestVersionsITPostgresConnect(t *testing.T) {
	cdrsPostgresCfgPath := path.Join(*dataDir, "conf", "samples", "tutpostgres")
	cfg, err := config.NewCGRConfigFromFolder(cdrsPostgresCfgPath)
	if err != nil {
		t.Fatal(err)
	}
	dataDB, err := ConfigureDataStorage(cfg.DataDbType, cfg.DataDbHost, cfg.DataDbPort, cfg.DataDbName, cfg.DataDbUser, cfg.DataDbPass, cfg.DBDataEncoding, cfg.CacheConfig, *loadHistorySize)
	if err != nil {
		log.Fatal(err)
	}
	storDB, err := ConfigureStorStorage(cfg.StorDBType, cfg.StorDBHost, cfg.StorDBPort, cfg.StorDBName, cfg.StorDBUser, cfg.StorDBPass, cfg.DBDataEncoding,
		config.CgrConfig().StorDBMaxOpenConns, config.CgrConfig().StorDBMaxIdleConns, config.CgrConfig().StorDBConnMaxLifetime, config.CgrConfig().StorDBCDRSIndexes)
	if err != nil {
		log.Fatal(err)
	}
	storageDb = storDB
	dataDb = dataDB
}

func TestMigratorITPostgres(t *testing.T) {
	dbtype = utils.REDIS
	for _, stest := range sTestsITVersions {
		t.Run("TestMigratorITPostgres", stest)
	}
}

func testVersionsFlush(t *testing.T) {
	switch {
	case dbtype == utils.REDIS:
		dataDB := dataDb.(*RedisStorage)
		err := dataDB.Cmd("FLUSHALL").Err
		if err != nil {
			t.Error("Error when flushing Redis ", err.Error())
		}
		if err := storDB.Flush(path.Join(cfg.DataFolderPath, "storage", cfg.StorDBType)); err != nil {
			t.Error(err)
		}
	case dbtype == utils.MONGO:
		err := dataDb.Flush("")
		if err != nil {
			t.Error("Error when flushing Mongo ", err.Error())
		}
		if err := storDB.Flush(path.Join(cfg.DataFolderPath, "storage", cfg.StorDBType)); err != nil {
			t.Error(err)
		}
	}
}

func TestVersion(t *testing.T) {
	var test string
	var currentVersion Versions
	var testVersion Versions
	storType := dataDb.GetStorageType()
	switch storType {
	case utils.MONGO:
		currentVersion = Versions{utils.Accounts: 2, utils.Actions: 2, utils.ActionTriggers: 2, utils.ActionPlans: 2, utils.SharedGroups: 2, utils.COST_DETAILS: 2}
		testVersion = Versions{utils.Accounts: 1, utils.Actions: 2, utils.ActionTriggers: 2, utils.ActionPlans: 2, utils.SharedGroups: 2, utils.COST_DETAILS: 2}
		test = "Migration needed: please backup cgr data and run : <cgr-migrator -migrate=*accounts>"
	case utils.POSTGRES:
		currentVersion = CurrentStorDBVersions()
		testVersion = Versions{utils.COST_DETAILS: 1}
		test = "Migration needed: please backup cgr data and run : <cgr-migrator -migrate=*cost_details>"
	case utils.MYSQL:
		currentVersion = CurrentStorDBVersions()
		testVersion = Versions{utils.COST_DETAILS: 1}
		test = "Migration needed: please backup cgr data and run : <cgr-migrator -migrate=*cost_details>"
	case utils.REDIS:
		currentVersion = CurrentDataDBVersions()
		testVersion = Versions{utils.Accounts: 1, utils.Actions: 2, utils.ActionTriggers: 2, utils.ActionPlans: 2, utils.SharedGroups: 2}
		test = "Migration needed: please backup cgr data and run : <cgr-migrator -migrate=*accounts>"
	case utils.MAPSTOR:
		currentVersion = Versions{utils.Accounts: 2, utils.Actions: 2, utils.ActionTriggers: 2, utils.ActionPlans: 2, utils.SharedGroups: 2, utils.COST_DETAILS: 2}
		testVersion = Versions{utils.Accounts: 1, utils.Actions: 2, utils.ActionTriggers: 2, utils.ActionPlans: 2, utils.SharedGroups: 2, utils.COST_DETAILS: 2}
		test = "Migration needed: please backup cgr data and run : <cgr-migrator -migrate=*accounts>"
	}

	//dataDB
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

	storType = storDb.GetStorageType()
	switch storType {
	case utils.MONGO:
		currentVersion = Versions{utils.Accounts: 2, utils.Actions: 2, utils.ActionTriggers: 2, utils.ActionPlans: 2, utils.SharedGroups: 2, utils.COST_DETAILS: 2}
		testVersion = Versions{utils.Accounts: 1, utils.Actions: 2, utils.ActionTriggers: 2, utils.ActionPlans: 2, utils.SharedGroups: 2, utils.COST_DETAILS: 2}
		test = "Migration needed: please backup cgr data and run : <cgr-migrator -migrate=*accounts>"
	case utils.POSTGRES:
		currentVersion = CurrentStorDBVersions()
		testVersion = Versions{utils.COST_DETAILS: 1}
		test = "Migration needed: please backup cgr data and run : <cgr-migrator -migrate=*cost_details>"
	case utils.MYSQL:
		currentVersion = CurrentStorDBVersions()
		testVersion = Versions{utils.COST_DETAILS: 1}
		test = "Migration needed: please backup cgr data and run : <cgr-migrator -migrate=*cost_details>"
	case utils.REDIS:
		currentVersion = CurrentDataDBVersions()
		testVersion = Versions{utils.Accounts: 1, utils.Actions: 2, utils.ActionTriggers: 2, utils.ActionPlans: 2, utils.SharedGroups: 2}
		test = "Migration needed: please backup cgr data and run : <cgr-migrator -migrate=*accounts>"
	case utils.MAPSTOR:
		currentVersion = Versions{utils.Accounts: 2, utils.Actions: 2, utils.ActionTriggers: 2, utils.ActionPlans: 2, utils.SharedGroups: 2, utils.COST_DETAILS: 2}
		testVersion = Versions{utils.Accounts: 1, utils.Actions: 2, utils.ActionTriggers: 2, utils.ActionPlans: 2, utils.SharedGroups: 2, utils.COST_DETAILS: 2}
		test = "Migration needed: please backup cgr data and run : <cgr-migrator -migrate=*accounts>"
	}
	//storDB
	if _, rcvErr := storDb.GetVersions(utils.TBLVersions); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := CheckVersions(storDb); err != nil {
		t.Error(err)
	}
	if rcv, err := storDb.GetVersions(utils.TBLVersions); err != nil {
		t.Error(err)
	} else if len(currentVersion) != len(rcv) {
		t.Errorf("Expecting: %v, received: %v", currentVersion, rcv)
	}
	if err = storDb.RemoveVersions(currentVersion); err != nil {
		t.Error(err)
	}
	if _, rcvErr := storDb.GetVersions(utils.TBLVersions); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := storDb.SetVersions(testVersion, false); err != nil {
		t.Error(err)
	}
	if err := CheckVersions(storDb); err.Error() != test {
		t.Error(err)
	}
	if err = storDb.RemoveVersions(testVersion); err != nil {
		t.Error(err)
	}

}
