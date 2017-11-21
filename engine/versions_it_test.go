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
	dm3             *DataManager
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
	if cfg, err = config.NewCGRConfigFromFolder(path.Join(*dataDir, "conf", "samples", "tutmongo")); err != nil {
		t.Fatal(err)
	}
	if dm3, err = ConfigureDataStorage(cfg.DataDbType, cfg.DataDbHost,
		cfg.DataDbPort, cfg.DataDbName, cfg.DataDbUser, cfg.DataDbPass,
		cfg.DBDataEncoding, cfg.CacheCfg(), *loadHistorySize); err != nil {
		log.Fatal(err)
	}
	storageDb, err = ConfigureStorStorage(cfg.StorDBType, cfg.StorDBHost,
		cfg.StorDBPort, cfg.StorDBName, cfg.StorDBUser, cfg.StorDBPass, cfg.DBDataEncoding,
		config.CgrConfig().StorDBMaxOpenConns, config.CgrConfig().StorDBMaxIdleConns,
		config.CgrConfig().StorDBConnMaxLifetime, config.CgrConfig().StorDBCDRSIndexes)
	if err != nil {
		log.Fatal(err)
	}
	dbtype = utils.MONGO
	for _, stest := range sTestsITVersions {
		t.Run("TestVersionsITMongo", stest)
	}
}

func TestVersionsITRedisMYSQL(t *testing.T) {
	var err error
	if cfg, err = config.NewCGRConfigFromFolder(path.Join(*dataDir, "conf", "samples", "tutmysql")); err != nil {
		t.Fatal(err)
	}
	dm3, err = ConfigureDataStorage(cfg.DataDbType, cfg.DataDbHost, cfg.DataDbPort,
		cfg.DataDbName, cfg.DataDbUser, cfg.DataDbPass, cfg.DBDataEncoding, cfg.CacheCfg(), *loadHistorySize)
	if err != nil {
		log.Fatal(err)
	}

	storageDb, err = ConfigureStorStorage(cfg.StorDBType, cfg.StorDBHost, cfg.StorDBPort,
		cfg.StorDBName, cfg.StorDBUser, cfg.StorDBPass, cfg.DBDataEncoding,
		config.CgrConfig().StorDBMaxOpenConns, config.CgrConfig().StorDBMaxIdleConns,
		config.CgrConfig().StorDBConnMaxLifetime, config.CgrConfig().StorDBCDRSIndexes)
	if err != nil {
		log.Fatal(err)
	}
	dbtype = utils.REDIS
	for _, stest := range sTestsITVersions {
		t.Run("TestVersionsITRedis", stest)
	}
}

func TestVersionsITRedisPostgres(t *testing.T) {
	var err error
	if cfg, err = config.NewCGRConfigFromFolder(path.Join(*dataDir, "conf", "samples", "storage", "postgres")); err != nil {
		t.Fatal(err)
	}
	dm3, err = ConfigureDataStorage(cfg.DataDbType, cfg.DataDbHost, cfg.DataDbPort,
		cfg.DataDbName, cfg.DataDbUser, cfg.DataDbPass, cfg.DBDataEncoding, cfg.CacheCfg(), *loadHistorySize)
	if err != nil {
		log.Fatal(err)
	}
	storageDb, err = ConfigureStorStorage(cfg.StorDBType, cfg.StorDBHost,
		cfg.StorDBPort, cfg.StorDBName, cfg.StorDBUser, cfg.StorDBPass, cfg.DBDataEncoding,
		config.CgrConfig().StorDBMaxOpenConns, config.CgrConfig().StorDBMaxIdleConns,
		config.CgrConfig().StorDBConnMaxLifetime, config.CgrConfig().StorDBCDRSIndexes)
	if err != nil {
		log.Fatal(err)
	}

	dbtype = utils.REDIS
	for _, stest := range sTestsITVersions {
		t.Run("TestMigratorITPostgres", stest)
	}
}

func testVersionsFlush(t *testing.T) {
	err := dm3.DataDB().Flush("")
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
	dataDbVersions := CurrentDataDBVersions()
	storDbVersions := CurrentStorDBVersions()

	allVersions := make(Versions)
	for k, v := range dataDbVersions {
		allVersions[k] = v
	}
	for k, v := range storDbVersions {
		allVersions[k] = v
	}

	storType := dm3.DataDB().GetStorageType()
	switch storType {
	case utils.MONGO, utils.MAPSTOR:
		currentVersion = allVersions
		testVersion = allVersions
		testVersion[utils.Accounts] = 1
		test = "Migration needed: please backup cgr data and run : <cgr-migrator -migrate=*accounts>"
	case utils.REDIS:
		currentVersion = dataDbVersions
		testVersion = dataDbVersions
		testVersion[utils.Accounts] = 1

		test = "Migration needed: please backup cgr data and run : <cgr-migrator -migrate=*accounts>"
	}

	//dataDB
	if _, rcvErr := dm3.DataDB().GetVersions(utils.TBLVersions); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := CheckVersions(dm3.DataDB()); err != nil {
		t.Error(err)
	}
	if rcv, err := dm3.DataDB().GetVersions(utils.TBLVersions); err != nil {
		t.Error(err)
	} else if len(currentVersion) != len(rcv) {
		t.Errorf("Expecting: %v, received: %v", currentVersion, rcv)
	}
	if err = dm3.DataDB().RemoveVersions(currentVersion); err != nil {
		t.Error(err)
	}
	if _, rcvErr := dm3.DataDB().GetVersions(utils.TBLVersions); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := dm3.DataDB().SetVersions(testVersion, false); err != nil {
		t.Error(err)
	}
	if err := CheckVersions(dm3.DataDB()); err.Error() != test {
		t.Error(err)
	}
	if err = dm3.DataDB().RemoveVersions(testVersion); err != nil {
		t.Error(err)
	}
	storType = storageDb.GetStorageType()
	switch storType {
	case utils.MONGO, utils.MAPSTOR:
		currentVersion = allVersions
		testVersion = allVersions
		testVersion[utils.Accounts] = 1
		test = "Migration needed: please backup cgr data and run : <cgr-migrator -migrate=*accounts>"
	case utils.POSTGRES, utils.MYSQL:
		currentVersion = storDbVersions
		testVersion = allVersions
		testVersion[utils.COST_DETAILS] = 1
		test = "Migration needed: please backup cgr data and run : <cgr-migrator -migrate=*cost_details>"
	}
	//storageDb

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
