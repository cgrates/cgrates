//go:build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package engine

import (
	"log"
	"path"
	"strings"
	"testing"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

var (
	dm3               *DataManager
	versionsConfigDIR string
	vrsCfg            *config.CGRConfig
	sTestsITVersions  = []func(t *testing.T){
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
	case utils.MetaRedis:
		versionsConfigDIR = "tutredis"
	case utils.MetaMySQL:
		versionsConfigDIR = "tutmysql"
	case utils.MetaMongo:
		versionsConfigDIR = "tutmongo"
	case utils.MetaPostgres:
		versionsConfigDIR = "tutpostgres"
	default:
		t.Fatal("Unknown Database type")
	}

	for _, stest := range sTestsITVersions {
		t.Run(versionsConfigDIR, stest)
	}
}

func testInitConfig(t *testing.T) {
	var err error
	if vrsCfg, err = config.NewCGRConfigFromPath(context.Background(), path.Join(*utils.DataDir, "conf", "samples", versionsConfigDIR)); err != nil {
		t.Fatal(err)
	}
}

func testInitDataDB(t *testing.T) {
	dbConn, err := NewDBConn(vrsCfg.DbCfg().DBConns[utils.MetaDefault].Type,
		vrsCfg.DbCfg().DBConns[utils.MetaDefault].Host, vrsCfg.DbCfg().DBConns[utils.MetaDefault].Port,
		vrsCfg.DbCfg().DBConns[utils.MetaDefault].Name, vrsCfg.DbCfg().DBConns[utils.MetaDefault].User,
		vrsCfg.DbCfg().DBConns[utils.MetaDefault].Password, vrsCfg.GeneralCfg().DBDataEncoding, vrsCfg.DbCfg().DBConns[utils.MetaDefault].StringIndexedFields, vrsCfg.DbCfg().DBConns[utils.MetaDefault].PrefixIndexedFields,
		vrsCfg.DbCfg().DBConns[utils.MetaDefault].Opts, vrsCfg.DbCfg().Items)
	if err != nil {
		log.Fatal(err)
	}
	dbCM := NewDBConnManager(map[string]DataDB{utils.MetaDefault: dbConn}, vrsCfg.DbCfg())
	dm3 = NewDataManager(dbCM, vrsCfg, nil)
	cacheS := NewCacheS(vrsCfg, nil, nil, nil)
	dm3.SetCache(cacheS)

	if err != nil {
		log.Fatal(err)
	}
}

func testVersionsFlush(t *testing.T) {
	err := dm3.DB()[utils.MetaDefault].Flush(path.Join(vrsCfg.DataFolderPath, "storage", strings.Trim(vrsCfg.DbCfg().DBConns[utils.MetaDefault].Type, "*")))
	if err != nil {
		t.Error("Error when flushing ", err.Error())
	}

}

func testVersion(t *testing.T) {
	var test string
	var currentVersion Versions
	var testVersion Versions
	dataDbVersions := CurrentDataDBVersions()

	allVersions := make(Versions)
	for k, v := range dataDbVersions {
		allVersions[k] = v
	}

	storType := dm3.DB()[utils.MetaDefault].GetStorageType()
	switch storType {
	case utils.MetaInternal:
		currentVersion = allVersions
		testVersion = allVersions
		testVersion[utils.AccountsStr] = 2
		test = "datadb version mismatch for Accounts (have 2, want 1): back up your data, flush the datadb and reload"
	case utils.MetaMongo, utils.MetaRedis, utils.MetaMySQL, utils.MetaPostgres:
		currentVersion = dataDbVersions
		testVersion = dataDbVersions
		testVersion[utils.AccountsStr] = 2
		test = "datadb version mismatch for Accounts (have 2, want 1): back up your data, flush the datadb and reload"
	}

	//dataDB
	if _, rcvErr := dm3.DB()[utils.MetaDefault].GetVersions(""); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := CheckVersions(dm3.DB()[utils.MetaDefault]); err != nil {
		t.Error(err)
	}
	if rcv, err := dm3.DB()[utils.MetaDefault].GetVersions(""); err != nil {
		t.Error(err)
	} else if len(currentVersion) != len(rcv) {
		t.Errorf("Expecting: %v, received: %v", currentVersion, rcv)
	}
	if err := dm3.DB()[utils.MetaDefault].RemoveVersions(currentVersion); err != nil {
		t.Error(err)
	}
	if _, rcvErr := dm3.DB()[utils.MetaDefault].GetVersions(""); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := dm3.DB()[utils.MetaDefault].SetVersions(testVersion, false); err != nil {
		t.Error(err)
	}
	if err := CheckVersions(dm3.DB()[utils.MetaDefault]); err == nil || err.Error() != test {
		t.Error(err)
	}
	if err := dm3.DB()[utils.MetaDefault].RemoveVersions(testVersion); err != nil {
		t.Error(err)
	}
}
