//go:build integration
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

package main

import (
	"bytes"
	"flag"
	"fmt"
	"net"
	"os/exec"
	"path"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/birpc/jsonrpc"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

var (
	dataDir   = flag.String("data_dir", "/usr/share/cgrates", "CGR data dir path here")
	dbType    = flag.String("dbtype", utils.MetaInternal, "The type of DataBase (Internal/Mongo/mySql)")
	waitRater = flag.Int("wait_rater", 100, "Number of milliseconds to wait for rater to start and cache")
	encoding  = flag.String("rpc", utils.MetaJSON, "what encoding whould be used for rpc comunication")
)

func TestLoadConfig(t *testing.T) {
	// DataDb
	*cfgPath = path.Join(*dataDir, "conf", "samples", "tutmysql")
	*dataDBType = utils.MetaRedis
	*dataDBHost = "localhost"
	*dataDBPort = "2012"
	*dataDBName = "100"
	*dataDBUser = "cgrates2"
	*dataDBPasswd = "toor"
	*dbRedisSentinel = "sentinel1"
	expDBcfg := &config.DataDbCfg{
		Type:     utils.MetaRedis,
		Host:     "localhost",
		Port:     "2012",
		Name:     "100",
		User:     "cgrates2",
		Password: "toor",
		Opts: &config.DataDBOpts{
			RedisMaxConns:           10,
			RedisConnectAttempts:    20,
			RedisSentinel:           "sentinel1",
			MongoQueryTimeout:       10 * time.Second,
			RedisClusterSync:        5 * time.Second,
			RedisClusterOndownDelay: 0,
			RedisCluster:            false,
			RedisTLS:                false,
		},
		RmtConns: []string{},
		RplConns: []string{},
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
	*cachingDlay = 5 * time.Second
	*dbDataEncoding = utils.MetaJSON
	*timezone = utils.Local
	ldrCfg := loadConfig()
	ldrCfg.DataDbCfg().Items = nil
	if !reflect.DeepEqual(ldrCfg.DataDbCfg(), expDBcfg) {
		t.Errorf("Expected %s received %s", utils.ToJSON(expDBcfg), utils.ToJSON(ldrCfg.DataDbCfg()))
	}
	if ldrCfg.GeneralCfg().DBDataEncoding != utils.MetaJSON {
		t.Errorf("Expected %s received %s", utils.MetaJSON, ldrCfg.GeneralCfg().DBDataEncoding)
	}
	if ldrCfg.GeneralCfg().DefaultTimezone != utils.Local {
		t.Errorf("Expected %s received %s", utils.Local, ldrCfg.GeneralCfg().DefaultTimezone)
	}
	if !ldrCfg.LoaderCgrCfg().DisableReverse {
		t.Errorf("Expected %v received %v", true, ldrCfg.LoaderCgrCfg().DisableReverse)
	}
	if ldrCfg.GeneralCfg().DefaultCaching != utils.MetaLoad {
		t.Errorf("Expected %s received %s", utils.MetaLoad, ldrCfg.GeneralCfg().DefaultCaching)
	}
	if ldrCfg.GeneralCfg().CachingDelay != 5*time.Second {
		t.Errorf("Expected %s received %s", 5*time.Second, ldrCfg.GeneralCfg().CachingDelay)
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

		testLoadItStartLoader,
		testLoadItCheckTenantFlag,
		testLoadItStartLoaderWithTenant,

		testLoadItConnectToDB,
		testLoadItCheckAttributes,
		testLoadItStartLoaderRemove,
		testLoadItCheckAttributes2,

		testLoadItCheckAttributes2,

		testLoadItCheckAttributes2,
		testLoadItCheckAttributes,

		testLoadItStartLoaderWithDelayWConf,
		testLoadItStartLoaderWithDelayWFlag,
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
	if ldrItCfg, err = config.NewCGRConfigFromPath(context.Background(), ldrItCfgPath); err != nil {
		t.Error(err)
	}
}

func testLoadItResetDataDB(t *testing.T) {
	if err := engine.InitDataDB(ldrItCfg); err != nil {
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
		ldrItCfg.DataDbCfg().Opts, ldrItCfg.DataDbCfg().Items); err != nil {
		t.Fatal(err)
	}
}

func testLoadItCheckAttributes(t *testing.T) {
	eAttrPrf := &engine.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_1001_SIMPLEAUTH",
		FilterIDs: []string{"*string:~*req.Account:1001", "*string:~*opts.*context:simpleauth"},
		Attributes: []*engine.Attribute{
			{
				FilterIDs: []string{},
				Path:      utils.MetaReq + utils.NestingSep + "Password",
				Type:      utils.MetaConstant,
				Value:     config.NewRSRParsersMustCompile("CGRateS.org", utils.InfieldSep),
			},
		},
		Blockers: utils.DynamicBlockers{
			{
				Blocker: false,
			},
		},
	}
	eAttrPrf.Weights = make(utils.DynamicWeights, 1)
	eAttrPrf.Weights[0] = &utils.DynamicWeight{
		Weight: 20.0,
	}

	if err := db.SetAttributeProfileDrv(context.Background(), eAttrPrf); err != nil {
		t.Fatal(err)
	}

	if attr, err := db.GetAttributeProfileDrv(context.Background(), "cgrates.org", "ATTR_1001_SIMPLEAUTH"); err != nil {
		t.Fatal(err)
	} else {
		attr.Compile()
		sort.Strings(eAttrPrf.FilterIDs)
		sort.Strings(attr.FilterIDs)
		if !reflect.DeepEqual(eAttrPrf, attr) {
			t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eAttrPrf), utils.ToJSON(attr))
		}
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
	if _, err := db.GetAttributeProfileDrv(context.Background(), "cgrates.org", "ATTR_1001_SESSIONAUTH"); err != utils.ErrNotFound {
		t.Fatal(err)
	}
}

func testLoadItStartLoaderWithTenant(t *testing.T) {
	cmd := exec.Command("cgr-loader", "-config_path="+ldrItCfgPath, "-path="+path.Join(*dataDir, "tariffplans", "tutorial"), fmt.Sprintf("-caches_address=%s", address), "-scheduler_address=", `-tenant="tenant.com"`, "-verbose")
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	listener.Close()
	if resp != "\"tenant.com\"" {
		t.Errorf("Expected \"tenant.com\" \n but received \n %q", resp)
	}
}

type mockCache int

func (c *mockCache) ReloadCache(ctx *context.Context, args *utils.AttrReloadCacheWithAPIOpts, reply *string) (err error) {
	resp = args.Tenant
	*reply = "OK"
	return nil
}

func (c *mockCache) Clear(ctx *context.Context, args *utils.AttrCacheIDsWithAPIOpts,
	reply *string) error {
	*reply = args.Tenant
	return nil
}

var address string
var listener net.Listener
var resp string

func testLoadItCheckTenantFlag(t *testing.T) {
	err := birpc.RegisterName("CacheSv1", new(mockCache))
	if err != nil {
		t.Error(err)
	}

	listener, err = net.Listen("tcp", ":0")
	if err != nil {
		t.Error(err)
	}
	address = listener.Addr().String()

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go birpc.ServeCodec(jsonrpc.NewServerCodec(conn))
		}
	}()
}

func testLoadItStartLoaderWithDelayWConf(t *testing.T) {
	cmd := exec.Command("cgr-loader", "-config_path="+path.Join(*dataDir, "conf", "samples", "apis_config_mysql"), "-path="+path.Join(*dataDir, "tariffplans", "tutorial"), "-caches_address=", "-scheduler_address=", "-tpid=TPID")
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

func testLoadItStartLoaderWithDelayWFlag(t *testing.T) {
	cmd := exec.Command("cgr-loader", "-path="+path.Join(*dataDir, "tariffplans", "tutorial"), "-caches_address=", "-scheduler_address=", "-tpid=TPID", "-caching_delay=5s")
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
