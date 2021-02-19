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

package v1

import (
	"net/rpc"
	"path"
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	configCfgPath   string
	configCfg       *config.CGRConfig
	configRPC       *rpc.Client
	configConfigDIR string //run tests for specific configuration

	sTestsConfig = []func(t *testing.T){
		testConfigSInitCfg,
		testConfigSInitDataDb,
		testConfigSResetStorDb,
		testConfigSStartEngine,
		testConfigSRPCConn,
		testConfigSReloadConfigFromJSONSessionS,
		testConfigSKillEngine,
	}
)

//Test start here
func TestConfigSIT(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		configConfigDIR = "tutinternal"
	case utils.MetaMySQL:
		configConfigDIR = "tutmysql"
	case utils.MetaMongo:
		configConfigDIR = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsConfig {
		t.Run(configConfigDIR, stest)
	}
}

func testConfigSInitCfg(t *testing.T) {
	var err error
	configCfgPath = path.Join(*dataDir, "conf", "samples", configConfigDIR)
	configCfg, err = config.NewCGRConfigFromPath(configCfgPath)
	if err != nil {
		t.Error(err)
	}
	configCfg.DataFolderPath = *dataDir
	config.SetCgrConfig(configCfg)
}

func testConfigSInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(configCfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func testConfigSResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(configCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testConfigSStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(configCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testConfigSRPCConn(t *testing.T) {
	var err error
	configRPC, err = newRPCClient(configCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}
func testConfigSReloadConfigFromJSONSessionS(t *testing.T) {

	var reply string
	if err := configRPC.Call(utils.ConfigSv1ReloadConfigFromJSON, &config.JSONReloadWithArgDispatcher{
		JSON: map[string]interface{}{
			"sessions": map[string]interface{}{
				"enabled":          true,
				"resources_conns":  []string{"*localhost"},
				"suppliers_conns":  []string{"*localhost"},
				"attributes_conns": []string{"*localhost"},
				"rals_conns":       []string{"*internal"},
				"cdrs_conns":       []string{"*internal"},
				"configs_conns":    []string{"*internal"},
			},
		},
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %s", reply)
	}
	exp := map[string]interface{}{
		"Enabled":       true,
		"ListenBijson":  "127.0.0.1:2014",
		"ChargerSConns": []interface{}{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers)},
		"RALsConns":     []interface{}{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResponder)},
		"ResSConns":     []interface{}{utils.MetaLocalHost},
		"ThreshSConns":  []interface{}{},
		"StatSConns":    []interface{}{},
		"SupplSConns":   []interface{}{utils.MetaLocalHost},
		"AttrSConns":    []interface{}{utils.MetaLocalHost},
		"CDRsConns":     []interface{}{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCDRs)},

		"ReplicationConns":    []interface{}{},
		"SessionIndexes":      map[string]interface{}{"OriginID": true},
		"ClientProtocol":      1.,
		"TerminateAttempts":   5.,
		"ChannelSyncInterval": 0.,
		"DebitInterval":       0.,
		"SessionTTL":          0.,
		"SessionTTLLastUsed":  nil,
		"SessionTTLLastUsage": nil,
		"SessionTTLMaxDelay":  nil,
		"SessionTTLUsage":     nil,
		"StoreSCosts":         false,
		"AlterableFields":     map[string]interface{}{},
		"DefaultUsage": map[string]interface{}{
			utils.META_ANY: "3h0m0s",
			utils.VOICE:    "3h0m0s",
			utils.DATA:     "1048576",
			utils.SMS:      "1",
		},
	}
	if *encoding == utils.MetaGOB {
		var empty []interface{}
		exp["ThreshSConns"] = empty
		exp["StatSConns"] = empty
		exp["ReplicationConns"] = empty
	}
	var rpl map[string]interface{}
	if err := configRPC.Call(utils.ConfigSv1GetJSONSection, &config.StringWithArgDispatcher{
		Section: config.SessionSJson,
	}, &rpl); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, rpl) {
		t.Errorf("Expected %+v , received: %+v ", utils.ToJSON(exp), utils.ToJSON(rpl))
	}
}

func testConfigSKillEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
