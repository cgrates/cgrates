//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package engine

import (
	"log"
	"path"
	"strings"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

var (
	storageDb         Storage
	dm3               *DataManager
	versionsConfigDIR string

	sTestsITVersions = []func(t *testing.T){
		testInitConfig,
		testInitDataDB,
		testVersionsFlush,
		testVersion,
		testVersionsFlush,
	}
)

func TestVersionsIT(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		t.SkipNow()
	case utils.MetaMySQL:
		versionsConfigDIR = "tutmysql"
	case utils.MetaMongo:
		versionsConfigDIR = "tutmongo"
	case utils.MetaPostgres:
		versionsConfigDIR = "storage/postgres"
	default:
		t.Fatal("Unknown Database type")
	}

	for _, stest := range sTestsITVersions {
		t.Run(versionsConfigDIR, stest)
	}
}

func testInitConfig(t *testing.T) {
	if cfg, err = config.NewCGRConfigFromPath(path.Join(*utils.DataDir, "conf", "samples", versionsConfigDIR)); err != nil {
		t.Fatal(err)
	}
}

func testInitDataDB(t *testing.T) {
	dbConn, err := NewDataDBConn(cfg.DataDbCfg().DataDbType,
		cfg.DataDbCfg().DataDbHost, cfg.DataDbCfg().DataDbPort,
		cfg.DataDbCfg().DataDbName, cfg.DataDbCfg().DataDbUser,
		cfg.DataDbCfg().DataDbPass, cfg.GeneralCfg().DBDataEncoding,
		"", cfg.DataDbCfg().Items)
	if err != nil {
		log.Fatal(err)
	}
	dm3 = NewDataManager(dbConn, cfg.CacheCfg(), nil)

	storageDb, err = NewStorDBConn(cfg.StorDbCfg().Type,
		cfg.StorDbCfg().Host, cfg.StorDbCfg().Port,
		cfg.StorDbCfg().Name, cfg.StorDbCfg().User,
		cfg.StorDbCfg().Password, cfg.GeneralCfg().DBDataEncoding, cfg.StorDbCfg().SSLMode,
		cfg.StorDbCfg().MaxOpenConns, cfg.StorDbCfg().MaxIdleConns,
		cfg.StorDbCfg().ConnMaxLifetime, cfg.StorDbCfg().StringIndexedFields,
		cfg.StorDbCfg().PrefixIndexedFields, cfg.StorDbCfg().Items)
	if err != nil {
		log.Fatal(err)
	}
}

func testVersionsFlush(t *testing.T) {
	err := dm3.DataDB().Flush("")
	if err != nil {
		t.Error("Error when flushing Mongo ", err.Error())
	}
	db_Path := strings.Trim(cfg.StorDbCfg().Type, "*")
	if err := storageDb.Flush(path.Join(cfg.DataFolderPath, "storage", db_Path)); err != nil {
		t.Error(err)
	}
	SetDBVersions(storageDb)
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
	case utils.MetaInternal:
		currentVersion = allVersions
		testVersion = allVersions
		testVersion[utils.Accounts] = 1
		test = "Migration needed: please backup cgr data and run : <cgr-migrator -exec=*accounts>"
	case utils.MetaMongo, utils.MetaRedis:
		currentVersion = dataDbVersions
		testVersion = dataDbVersions
		testVersion[utils.Accounts] = 1

		test = "Migration needed: please backup cgr data and run : <cgr-migrator -exec=*accounts>"
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
	case utils.MetaInternal:
		currentVersion = allVersions
		testVersion = allVersions
		testVersion[utils.Accounts] = 1
		test = "Migration needed: please backup cgr data and run : <cgr-migrator -exec=*accounts>"
	case utils.MetaMongo, utils.MetaPostgres, utils.MetaMySQL:
		currentVersion = storDbVersions
		testVersion = allVersions
		testVersion[utils.CostDetails] = 1
		test = "Migration needed: please backup cgr data and run : <cgr-migrator -exec=*cost_details>"
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
