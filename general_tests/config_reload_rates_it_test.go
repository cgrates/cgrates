//go:build integration
// +build integration

/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/
package general_tests

import (
	"path"
	"testing"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	testRateCfgDir  string
	testRateCfgPath string
	testRateCfg     *config.CGRConfig
	testRateRPC     *birpc.Client

	testRateTests = []func(t *testing.T){
		testRateLoadConfig,
		testRateFlushDBs,

		testRateStartEngine,
		testRateRPCConn,
		testRateConfigSReloadRates,
		testRateStopCgrEngine,
	}
)

func TestRateChange(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		testRateCfgDir = "cfg_rld_rates_internal"
	case utils.MetaMySQL:
		testRateCfgDir = "cfg_rld_rates_mysql"
	case utils.MetaMongo:
		testRateCfgDir = "cfg_rld_rates_mongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, testRateest := range testRateTests {
		t.Run(testRateCfgDir, testRateest)
	}
}

func testRateLoadConfig(t *testing.T) {
	var err error
	testRateCfgPath = path.Join(*utils.DataDir, "conf", "samples", testRateCfgDir)
	if testRateCfg, err = config.NewCGRConfigFromPath(context.Background(), testRateCfgPath); err != nil {
		t.Error(err)
	}
}

func testRateFlushDBs(t *testing.T) {
	if err := engine.InitDataDB(testRateCfg); err != nil {
		t.Fatal(err)
	}
	if err := engine.InitStorDB(testRateCfg); err != nil {
		t.Fatal(err)
	}
}

func testRateStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(testRateCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

func testRateConfigSReloadRates(t *testing.T) {
	var replyPingBf string
	if err := testRateRPC.Call(context.Background(), utils.RateSv1CostForEvent, &utils.CGREvent{}, &replyPingBf); err == nil || err.Error() != "rpc: can't find service RateSv1.CostForEvent" {
		t.Error(err)
	}
	var reply string
	if err := testRateRPC.Call(context.Background(), utils.ConfigSv1SetConfigFromJSON, &config.SetConfigFromJSONArgs{
		Tenant: "cgrates.org",
		Config: "{\"rates\":{\"enabled\":true}}",
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := "{\"rates\":{\"enabled\":true,\"exists_indexed_fields\":[],\"indexed_selects\":true,\"nested_fields\":false,\"notexists_indexed_fields\":[],\"opts\":{\"*intervalStart\":[{\"FilterIDs\":null,\"Tenant\":\"\"}],\"*profileIDs\":[],\"*profileIgnoreFilters\":[{\"FilterIDs\":null,\"Tenant\":\"\"}],\"*startTime\":[{\"FilterIDs\":null,\"Tenant\":\"\"}],\"*usage\":[{\"FilterIDs\":null,\"Tenant\":\"\"}]},\"prefix_indexed_fields\":[],\"rate_exists_indexed_fields\":[],\"rate_indexed_selects\":true,\"rate_nested_fields\":false,\"rate_notexists_indexed_fields\":[],\"rate_prefix_indexed_fields\":[],\"rate_suffix_indexed_fields\":[],\"suffix_indexed_fields\":[],\"verbosity\":1000}}"
	var rpl string
	if err := testRateRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.RateSJSON},
	}, &rpl); err != nil {
		t.Error(err)
	} else if cfgStr != rpl {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl))
	}

	var replyPingAf string
	if err := testRateRPC.Call(context.Background(), utils.RateSv1CostForEvent, &utils.CGREvent{}, &replyPingAf); err == nil || err.Error() != "NOT_FOUND" {
		t.Error(err)
	}
}

func testRateRPCConn(t *testing.T) {
	testRateRPC = engine.NewRPCClient(t, testRateCfg.ListenCfg(), *utils.Encoding)
}

func testRateStopCgrEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
