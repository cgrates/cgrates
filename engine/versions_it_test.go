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
	"log"
	"path"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

var (
	storageDb Storage
	dm3       *DataManager
	dbtype    string
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
	if dm3, err = ConfigureDataStorage(cfg.DataDbCfg().DataDbType,
		cfg.DataDbCfg().DataDbHost, cfg.DataDbCfg().DataDbPort,
		cfg.DataDbCfg().DataDbName, cfg.DataDbCfg().DataDbUser,
		cfg.DataDbCfg().DataDbPass, cfg.GeneralCfg().DBDataEncoding,
		cfg.CacheCfg(), ""); err != nil {
		log.Fatal(err)
	}
	storageDb, err = ConfigureStorStorage(cfg.StorDbCfg().StorDBType,
		cfg.StorDbCfg().StorDBHost, cfg.StorDbCfg().StorDBPort,
		cfg.StorDbCfg().StorDBName, cfg.StorDbCfg().StorDBUser,
		cfg.StorDbCfg().StorDBPass, cfg.GeneralCfg().DBDataEncoding,
		config.CgrConfig().StorDbCfg().StorDBMaxOpenConns,
		config.CgrConfig().StorDbCfg().StorDBMaxIdleConns,
		config.CgrConfig().StorDbCfg().StorDBConnMaxLifetime,
		config.CgrConfig().StorDbCfg().StorDBCDRSIndexes)
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
	dm3, err = ConfigureDataStorage(cfg.DataDbCfg().DataDbType,
		cfg.DataDbCfg().DataDbHost, cfg.DataDbCfg().DataDbPort,
		cfg.DataDbCfg().DataDbName, cfg.DataDbCfg().DataDbUser,
		cfg.DataDbCfg().DataDbPass, cfg.GeneralCfg().DBDataEncoding, cfg.CacheCfg(), "")
	if err != nil {
		log.Fatal(err)
	}

	storageDb, err = ConfigureStorStorage(cfg.StorDbCfg().StorDBType,
		cfg.StorDbCfg().StorDBHost, cfg.StorDbCfg().StorDBPort,
		cfg.StorDbCfg().StorDBName, cfg.StorDbCfg().StorDBUser,
		cfg.StorDbCfg().StorDBPass, cfg.GeneralCfg().DBDataEncoding,
		config.CgrConfig().StorDbCfg().StorDBMaxOpenConns,
		config.CgrConfig().StorDbCfg().StorDBMaxIdleConns,
		config.CgrConfig().StorDbCfg().StorDBConnMaxLifetime,
		config.CgrConfig().StorDbCfg().StorDBCDRSIndexes)
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
	dm3, err = ConfigureDataStorage(cfg.DataDbCfg().DataDbType,
		cfg.DataDbCfg().DataDbHost, cfg.DataDbCfg().DataDbPort,
		cfg.DataDbCfg().DataDbName, cfg.DataDbCfg().DataDbUser,
		cfg.DataDbCfg().DataDbPass, cfg.GeneralCfg().DBDataEncoding, cfg.CacheCfg(), "")
	if err != nil {
		log.Fatal(err)
	}
	storageDb, err = ConfigureStorStorage(cfg.StorDbCfg().StorDBType,
		cfg.StorDbCfg().StorDBHost, cfg.StorDbCfg().StorDBPort,
		cfg.StorDbCfg().StorDBName, cfg.StorDbCfg().StorDBUser,
		cfg.StorDbCfg().StorDBPass, cfg.GeneralCfg().DBDataEncoding,
		config.CgrConfig().StorDbCfg().StorDBMaxOpenConns,
		config.CgrConfig().StorDbCfg().StorDBMaxIdleConns,
		config.CgrConfig().StorDbCfg().StorDBConnMaxLifetime,
		config.CgrConfig().StorDbCfg().StorDBCDRSIndexes)
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
	if err := storageDb.Flush(path.Join(cfg.DataFolderPath, "storage", cfg.StorDbCfg().StorDBType)); err != nil {
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
	case utils.MAPSTOR:
		currentVersion = allVersions
		testVersion = allVersions
		testVersion[utils.Accounts] = 1
		test = "Migration needed: please backup cgr data and run : <cgr-migrator -migrate=*accounts>"
	case utils.MONGO, utils.REDIS:
		currentVersion = dataDbVersions
		testVersion = dataDbVersions
		testVersion[utils.Accounts] = 1

		test = "Migration needed: please backup cgr data and run : <cgr-migrator -migrate=*accounts>"
	}

	//dataDB
	if _, rcvErr := dm3.DataDB().GetVersions(""); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := CheckVersions(dm3.DataDB()); err != nil {
		t.Error(err)
	}
	if rcv, err := dm3.DataDB().GetVersions(""); err != nil {
		t.Error(err)
	} else if len(currentVersion) != len(rcv) {
		t.Errorf("Expecting: %v, received: %v", currentVersion, rcv)
	}
	if err = dm3.DataDB().RemoveVersions(currentVersion); err != nil {
		t.Error(err)
	}
	if _, rcvErr := dm3.DataDB().GetVersions(""); rcvErr != utils.ErrNotFound {
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
	case utils.MAPSTOR:
		currentVersion = allVersions
		testVersion = allVersions
		testVersion[utils.Accounts] = 1
		test = "Migration needed: please backup cgr data and run : <cgr-migrator -migrate=*accounts>"
	case utils.MONGO, utils.POSTGRES, utils.MYSQL:
		currentVersion = storDbVersions
		testVersion = allVersions
		testVersion[utils.CostDetails] = 1
		test = "Migration needed: please backup cgr data and run : <cgr-migrator -migrate=*cost_details>"
	}
	//storageDb

	if err := CheckVersions(storageDb); err != nil {
		t.Error(err)
	}
	if rcv, err := storageDb.GetVersions(""); err != nil {
		t.Error(err)
	} else if len(currentVersion) != len(rcv) {
		t.Errorf("Expecting: %v, received: %v", currentVersion, rcv)
	}
	if err = storageDb.RemoveVersions(currentVersion); err != nil {
		t.Error(err)
	}
	if _, rcvErr := storageDb.GetVersions(""); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := storageDb.SetVersions(testVersion, false); err != nil {
		t.Error(err)
	}
	if err := CheckVersions(storageDb); err != nil && err.Error() != test {
		t.Error(err)
	}
	if err = storageDb.RemoveVersions(testVersion); err != nil {
		t.Error(err)
	}

}
