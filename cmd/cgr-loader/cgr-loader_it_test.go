// +build integration

/*
Real-time Online/Offline Charging System (OerS) for Telecom & ISP environments
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

package main

import (
	"bytes"
	"flag"
	"os/exec"
	"path"
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

var (
	dataDir  = flag.String("data_dir", "/usr/share/cgrates", "CGR data dir path here")
	dbType   = flag.String("dbtype", utils.MetaInternal, "The type of DataBase (Internal/Mongo/mySql)")
	encoding = flag.String("rpc", utils.MetaJSON, "what encoding whould be used for rpc comunication")
)

func TestLoadConfig(t *testing.T) {
	// DataDb
	*cfgPath = path.Join(*dataDir, "conf", "samples", "tutmongo")
	*dataDBType = utils.Meta + utils.Redis
	*dataDBHost = "localhost"
	*dataDBPort = "2012"
	*dataDBName = "100"
	*dataDBUser = "cgrates2"
	*dataDBPasswd = "toor"
	*dbRedisSentinel = "sentinel1"
	expDBcfg := &config.DataDbCfg{
		Type:     utils.Redis,
		Host:     "localhost",
		Port:     "2012",
		Name:     "100",
		User:     "cgrates2",
		Password: "toor",
		Opts: map[string]interface{}{
			utils.RedisSentinelNameCfg:       "sentinel1",
			utils.QueryTimeoutCfg:            "10s",
			utils.RedisClusterSyncCfg:        "5s",
			utils.RedisClusterOnDownDelayCfg: "0",
			utils.RedisClusterCfg:            false,
			utils.RedisTLS:                   false,
			utils.RedisClientCertificate:     "",
			utils.RedisClientKey:             "",
			utils.RedisCACertificate:         "",
		},
		RmtConns: []string{},
		RplConns: []string{},
	}
	// StorDB
	*storDBType = utils.MetaPostgres
	*storDBHost = "localhost"
	*storDBPort = "2012"
	*storDBName = "cgrates2"
	*storDBUser = "10"
	*storDBPasswd = "toor"
	expStorDB := &config.StorDbCfg{
		Type:                utils.Postgres,
		Host:                "localhost",
		Port:                "2012",
		Name:                "cgrates2",
		User:                "10",
		Password:            "toor",
		StringIndexedFields: []string{},
		PrefixIndexedFields: []string{},
		Opts: map[string]interface{}{
			utils.ConnMaxLifetimeCfg: 0.,
			utils.QueryTimeoutCfg:    "10s",
			utils.MaxOpenConnsCfg:    100.,
			utils.MaxIdleConnsCfg:    10.,
			utils.SSLModeCfg:         "disable",
			utils.MysqlLocation:      "Local",
		},
	}
	// Loader
	*tpid = "1"
	*disableReverse = true
	*dataPath = "./path"
	*fieldSep = "$"
	*cacheSAddress = ""
	*schedulerAddress = ""
	// General
	*cachingArg = utils.MetaLoad
	*dbDataEncoding = utils.MetaJSON
	*timezone = utils.Local
	ldrCfg := loadConfig()
	ldrCfg.DataDbCfg().Items = nil
	ldrCfg.StorDbCfg().Items = nil
	if !reflect.DeepEqual(ldrCfg.DataDbCfg(), expDBcfg) {
		t.Errorf("Expected %s received %s", utils.ToJSON(expDBcfg), utils.ToJSON(ldrCfg.DataDbCfg()))
	}
	if ldrCfg.GeneralCfg().DBDataEncoding != utils.MetaJSON {
		t.Errorf("Expected %s received %s", utils.MetaJSON, ldrCfg.GeneralCfg().DBDataEncoding)
	}
	if ldrCfg.GeneralCfg().DefaultTimezone != utils.Local {
		t.Errorf("Expected %s received %s", utils.Local, ldrCfg.GeneralCfg().DefaultTimezone)
	}
	if !reflect.DeepEqual(ldrCfg.StorDbCfg(), expStorDB) {
		t.Errorf("Expected %s received %s", utils.ToJSON(expStorDB), utils.ToJSON(ldrCfg.StorDbCfg()))
	}
	if !ldrCfg.LoaderCgrCfg().DisableReverse {
		t.Errorf("Expected %v received %v", true, ldrCfg.LoaderCgrCfg().DisableReverse)
	}
	if ldrCfg.GeneralCfg().DefaultCaching != utils.MetaLoad {
		t.Errorf("Expected %s received %s", utils.MetaLoad, ldrCfg.GeneralCfg().DefaultCaching)
	}
	if *importID == utils.EmptyString {
		t.Errorf("Expected importID to be populated")
	}
	if ldrCfg.LoaderCgrCfg().TpID != "1" {
		t.Errorf("Expected %s received %s", "1", ldrCfg.LoaderCgrCfg().TpID)
	}
	if ldrCfg.LoaderCgrCfg().DataPath != "./path" {
		t.Errorf("Expected %s received %s", "./path", ldrCfg.LoaderCgrCfg().DataPath)
	}
	if ldrCfg.LoaderCgrCfg().FieldSeparator != '$' {
		t.Errorf("Expected %v received %v", '$', ldrCfg.LoaderCgrCfg().FieldSeparator)
	}
	if !reflect.DeepEqual(ldrCfg.LoaderCgrCfg().CachesConns, []string{}) {
		t.Errorf("Expected %v received %v", []string{}, ldrCfg.LoaderCgrCfg().CachesConns)
	}
	if !reflect.DeepEqual(ldrCfg.LoaderCgrCfg().ActionSConns, []string{}) {
		t.Errorf("Expected %v received %v", []string{}, ldrCfg.LoaderCgrCfg().ActionSConns)
	}
	*cacheSAddress = "127.0.0.1"
	*schedulerAddress = "127.0.0.2"
	*rpcEncoding = utils.MetaJSON
	ldrCfg = loadConfig()
	expAddrs := []string{"127.0.0.1"}
	if !reflect.DeepEqual(ldrCfg.LoaderCgrCfg().CachesConns, expAddrs) {
		t.Errorf("Expected %v received %v", expAddrs, ldrCfg.LoaderCgrCfg().CachesConns)
	}
	expAddrs = []string{"127.0.0.2"}
	if !reflect.DeepEqual(ldrCfg.LoaderCgrCfg().ActionSConns, expAddrs) {
		t.Errorf("Expected %v received %v", expAddrs, ldrCfg.LoaderCgrCfg().ActionSConns)
	}
	expaddr := config.RPCConns{
		utils.MetaBiJSONLocalHost: {
			Strategy: rpcclient.PoolFirst,
			PoolSize: 0,
			Conns: []*config.RemoteHost{{
				Address:   "127.0.0.1:2014",
				Transport: rpcclient.BiRPCJSON,
			}},
		},
		utils.MetaInternal: {
			Strategy: rpcclient.PoolFirst,
			PoolSize: 0,
			Conns: []*config.RemoteHost{{
				Address: utils.MetaInternal,
			}},
		},
		rpcclient.BiRPCInternal: {
			Strategy: rpcclient.PoolFirst,
			PoolSize: 0,
			Conns: []*config.RemoteHost{{
				Address: rpcclient.BiRPCInternal,
			}},
		},
		"*localhost": {
			Strategy: rpcclient.PoolFirst,
			Conns:    []*config.RemoteHost{{Address: "127.0.0.1:2012", Transport: utils.MetaJSON}},
		},
		"127.0.0.1": {
			Strategy: rpcclient.PoolFirst,
			Conns:    []*config.RemoteHost{{Address: "127.0.0.1", Transport: utils.MetaJSON}},
		},
		"127.0.0.2": {
			Strategy: rpcclient.PoolFirst,
			Conns:    []*config.RemoteHost{{Address: "127.0.0.2", Transport: utils.MetaJSON}},
		},
	}
	if !reflect.DeepEqual(ldrCfg.RPCConns(), expaddr) {
		t.Errorf("Expected %v received %v", utils.ToJSON(expaddr), utils.ToJSON(ldrCfg.RPCConns()))
	}
}

var (
	ldrItCfgDir  string
	ldrItCfgPath string
	ldrItCfg     *config.CGRConfig
	db           engine.DataDB

	ldrItTests = []func(t *testing.T){
		testLoadItLoadConfig,
		testLoadItResetDataDB,
		testLoadItResetStorDb,
		testLoadItStartLoader,
		testLoadItConnectToDB,
		testLoadItCheckAttributes,
		testLoadItStartLoaderRemove,
		testLoadItCheckAttributes2,

		testLoadItStartLoaderToStorDB,
		testLoadItStartLoaderFlushStorDB,
		testLoadItStartLoaderFromStorDB,
		testLoadItCheckAttributes2,

		testLoadItStartLoaderToStorDB,
		testLoadItCheckAttributes2,
		testLoadItStartLoaderFromStorDB,
		testLoadItCheckAttributes,
	}
)

func TestLoadIt(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		t.SkipNow()
	case utils.MetaMySQL:
		ldrItCfgDir = "tutmysql"
	case utils.MetaMongo:
		ldrItCfgDir = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range ldrItTests {
		t.Run("TestLoadIt", stest)
	}
}

func testLoadItLoadConfig(t *testing.T) {
	var err error
	ldrItCfgPath = path.Join(*dataDir, "conf", "samples", ldrItCfgDir)
	if ldrItCfg, err = config.NewCGRConfigFromPath(ldrItCfgPath); err != nil {
		t.Error(err)
	}
}

func testLoadItResetDataDB(t *testing.T) {
	if err := engine.InitDataDB(ldrItCfg); err != nil {
		t.Fatal(err)
	}
}

func testLoadItResetStorDb(t *testing.T) {
	if err := engine.InitStorDB(ldrItCfg); err != nil {
		t.Fatal(err)
	}
}

func testLoadItStartLoader(t *testing.T) {
	cmd := exec.Command("cgr-loader", "-config_path="+ldrItCfgPath, "-path="+path.Join(*dataDir, "tariffplans", "tutorial"), "-caches_address=", "-scheduler_address=")
	output := bytes.NewBuffer(nil)
	outerr := bytes.NewBuffer(nil)
	cmd.Stdout = output
	cmd.Stderr = outerr
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Log(outerr.String())
		t.Fatal(err)
	}
}

func testLoadItConnectToDB(t *testing.T) {
	var err error
	if db, err = engine.NewDataDBConn(ldrItCfg.DataDbCfg().Type,
		ldrItCfg.DataDbCfg().Host, ldrItCfg.DataDbCfg().Port,
		ldrItCfg.DataDbCfg().Name, ldrItCfg.DataDbCfg().User,
		ldrItCfg.DataDbCfg().Password, ldrItCfg.GeneralCfg().DBDataEncoding,
		ldrItCfg.DataDbCfg().Opts); err != nil {
		t.Fatal(err)
	}
}

func testLoadItCheckAttributes(t *testing.T) {
	eAttrPrf := &engine.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_1001_SIMPLEAUTH",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Contexts:  []string{"simpleauth"},
		Attributes: []*engine.Attribute{
			{
				FilterIDs: []string{},
				Path:      utils.MetaReq + utils.NestingSep + "Password",
				Type:      utils.MetaConstant,
				Value:     config.NewRSRParsersMustCompile("CGRateS.org", utils.InfieldSep),
			},
		},
		Weight: 20.0,
	}
	if attr, err := db.GetAttributeProfileDrv("cgrates.org", "ATTR_1001_SIMPLEAUTH"); err != nil {
		t.Fatal(err)
	} else if attr.Compile(); !reflect.DeepEqual(eAttrPrf, attr) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eAttrPrf), utils.ToJSON(attr))
	}
}

func testLoadItStartLoaderRemove(t *testing.T) {
	cmd := exec.Command("cgr-loader", "-config_path="+ldrItCfgPath, "-path="+path.Join(*dataDir, "tariffplans", "tutorial"), "-caches_address=", "-scheduler_address=", "-remove")
	output := bytes.NewBuffer(nil)
	outerr := bytes.NewBuffer(nil)
	cmd.Stdout = output
	cmd.Stderr = outerr
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Log(outerr.String())
		t.Fatal(err)
	}
}

func testLoadItCheckAttributes2(t *testing.T) {
	if _, err := db.GetAttributeProfileDrv("cgrates.org", "ATTR_1001_SESSIONAUTH"); err != utils.ErrNotFound {
		t.Fatal(err)
	}
}

func testLoadItStartLoaderToStorDB(t *testing.T) {
	cmd := exec.Command("cgr-loader", "-config_path="+ldrItCfgPath, "-path="+path.Join(*dataDir, "tariffplans", "tutorial"), "-caches_address=", "-scheduler_address=", "-to_stordb", "-tpid=TPID")
	output := bytes.NewBuffer(nil)
	outerr := bytes.NewBuffer(nil)
	cmd.Stdout = output
	cmd.Stderr = outerr
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Log(outerr.String())
		t.Fatal(err)
	}
}

func testLoadItStartLoaderFromStorDB(t *testing.T) {
	cmd := exec.Command("cgr-loader", "-config_path="+ldrItCfgPath, "-caches_address=", "-scheduler_address=", "-from_stordb", "-tpid=TPID")
	output := bytes.NewBuffer(nil)
	outerr := bytes.NewBuffer(nil)
	cmd.Stdout = output
	cmd.Stderr = outerr
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Log(outerr.String())
		t.Fatal(err)
	}
}

func testLoadItStartLoaderFlushStorDB(t *testing.T) {
	cmd := exec.Command("cgr-loader", "-config_path="+ldrItCfgPath, "-path="+path.Join(*dataDir, "tariffplans", "dispatchers"), "-caches_address=", "-scheduler_address=", "-to_stordb", "-flush_stordb", "-tpid=TPID")
	output := bytes.NewBuffer(nil)
	outerr := bytes.NewBuffer(nil)
	cmd.Stdout = output
	cmd.Stderr = outerr
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Log(outerr.String())
		t.Fatal(err)
	}
}
