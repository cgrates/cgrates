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
package general_tests

import (
	"path"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/apis"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/loaders"
	"github.com/cgrates/cgrates/utils"
)

var (
	loadersIDBIdxCfgDir                        string
	loadersIDBIdxCfgPath                       string
	loadersIDBIdxCfgPathInternal               = path.Join(*dataDir, "conf", "samples", "loaders_indexes_internal_db")
	loadersIDBIdxCfg, loadersIDBIdxCfgInternal *config.CGRConfig
	loadersIDBIdxRPC, loadersIDBIdxRPCInternal *birpc.Client

	LoadersIDBIdxTests = []func(t *testing.T){
		testLoadersIDBIdxItLoadConfig,
		testLoadersIDBIdxItFlushDBs,
		testLoadersIDBIdxItStartEngines,
		testLoadersIDBIdxItRPCConn,
		testLoadersIDBIdxItLoad,
		testLoadersIDBIdxCheckAttributes,
		testLoadersIDBIdxCheckAttributesIndexes,
		testLoadersIDBIdxItStopCgrEngine,
	}
)

func TestLoadersIDBIdxIt(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		loadersIDBIdxCfgDir = "tutinternal"
	case utils.MetaMySQL:
		loadersIDBIdxCfgDir = "tutmysql"
	case utils.MetaMongo:
		loadersIDBIdxCfgDir = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range LoadersIDBIdxTests {
		t.Run(loadersIDBIdxCfgDir, stest)
	}
}

func testLoadersIDBIdxItLoadConfig(t *testing.T) {
	loadersIDBIdxCfgPath = path.Join(*dataDir, "conf", "samples", loadersIDBIdxCfgDir)
	if loadersIDBIdxCfg, err = config.NewCGRConfigFromPath(context.Background(), loadersIDBIdxCfgPath); err != nil {
		t.Error(err)
	}
	if loadersIDBIdxCfgInternal, err = config.NewCGRConfigFromPath(context.Background(), loadersIDBIdxCfgPathInternal); err != nil {
		t.Error(err)
	}
}

func testLoadersIDBIdxItFlushDBs(t *testing.T) {
	if err := engine.InitDataDB(loadersIDBIdxCfg); err != nil {
		t.Fatal(err)
	}
	if err := engine.InitStorDB(loadersIDBIdxCfg); err != nil {
		t.Fatal(err)
	}
}

func testLoadersIDBIdxItStartEngines(t *testing.T) {
	if _, err := engine.StopStartEngine(loadersIDBIdxCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
	if _, err := engine.StartEngine(loadersIDBIdxCfgPathInternal, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testLoadersIDBIdxItRPCConn(t *testing.T) {
	var err error
	if loadersIDBIdxRPC, err = engine.NewRPCClient(loadersIDBIdxCfg.ListenCfg(), *encoding); err != nil {
		t.Fatal(err)
	}
	if loadersIDBIdxRPCInternal, err = engine.NewRPCClient(loadersIDBIdxCfgInternal.ListenCfg(), *encoding); err != nil {
		t.Fatal(err)
	}
}

func testLoadersIDBIdxItLoad(t *testing.T) {
	var reply string
	if err := loadersIDBIdxRPCInternal.Call(context.Background(), utils.LoaderSv1Run,
		&loaders.ArgsProcessFolder{
			APIOpts: map[string]any{
				utils.MetaStopOnError: false,
				utils.MetaCache:       utils.MetaReload,
			},
		}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned:", reply)
	}
	time.Sleep(100 * time.Millisecond)
}

func testLoadersIDBIdxCheckAttributes(t *testing.T) {
	exp := &engine.APIAttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_1001_SIMPLEAUTH",
		FilterIDs: []string{"*string:~*opts.*context:simpleauth", "*string:~*req.Account:1001"},
		Attributes: []*engine.ExternalAttribute{{
			Path:  utils.MetaReq + utils.NestingSep + "Password",
			Type:  utils.MetaConstant,
			Value: "CGRateS.org",
		}},
		Blockers: utils.DynamicBlockers{
			{
				Blocker: false,
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
	}

	var reply *engine.APIAttributeProfile
	if err := loadersIDBIdxRPC.Call(context.Background(), utils.AdminSv1GetAttributeProfile,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "ATTR_1001_SIMPLEAUTH",
			},
		}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, reply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(exp), utils.ToJSON(reply))
	}
}

func testLoadersIDBIdxCheckAttributesIndexes(t *testing.T) {
	expIdx := []string{
		"*string:*opts.*context:*sessions:ATTR_1001_SESSIONAUTH",
		"*string:*opts.*context:*sessions:ATTR_1002_SESSIONAUTH",
		"*string:*opts.*context:*sessions:ATTR_1003_SESSIONAUTH",
		"*string:*opts.*context:simpleauth:ATTR_1001_SIMPLEAUTH",
		"*string:*opts.*context:simpleauth:ATTR_1002_SIMPLEAUTH",
		"*string:*opts.*context:simpleauth:ATTR_1003_SIMPLEAUTH",
		"*string:*req.Account:1001:ATTR_1001_SESSIONAUTH",
		"*string:*req.Account:1001:ATTR_1001_SIMPLEAUTH",
		"*string:*req.Account:1002:ATTR_1002_SESSIONAUTH",
		"*string:*req.Account:1002:ATTR_1002_SIMPLEAUTH",
		"*string:*req.Account:1003:ATTR_1003_SESSIONAUTH",
		"*string:*req.Account:1003:ATTR_1003_SIMPLEAUTH",
		"*string:*req.SubscriberId:1006:ATTR_ACC_ALIAS",
	}
	var indexes []string
	if err := loadersIDBIdxRPC.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&apis.AttrGetFilterIndexes{
			ItemType:   utils.MetaAttributes,
			Tenant:     "cgrates.org",
			FilterType: utils.MetaString,
			Context:    "simpleauth",
		}, &indexes); err != nil {
		t.Error(err)
	} else if sort.Strings(indexes); !reflect.DeepEqual(indexes, expIdx) {
		t.Errorf("Expecting: %+v, received: %+v",
			utils.ToJSON(expIdx), utils.ToJSON(indexes))
	}
}

func testLoadersIDBIdxItStopCgrEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
