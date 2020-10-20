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
	"fmt"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os/exec"
	"path"
	"reflect"
	"testing"
	"time"

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
		testConfigSReloadConfigFromJSONEEs,
		testConfigSv1GetJSONSectionWithoutTenant,
		testConfigSKillEngine,
		testConfigStartEngineWithConfigs,
		testConfigStartEngineFromHTTP,
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
	if err := configRPC.Call(utils.ConfigSv1ReloadConfigFromJSON, &config.JSONReloadWithOpts{
		Tenant: "cgrates.org",
		JSON: map[string]interface{}{
			"sessions": map[string]interface{}{
				"enabled":          true,
				"resources_conns":  []string{"*localhost"},
				"routes_conns":     []string{"*localhost"},
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
		"Enabled":             true,
		"ListenBijson":        "127.0.0.1:2014",
		"ChargerSConns":       []interface{}{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers)},
		"RALsConns":           []interface{}{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResponder)},
		"ResSConns":           []interface{}{utils.MetaLocalHost},
		"ThreshSConns":        []interface{}{},
		"StatSConns":          []interface{}{},
		"RouteSConns":         []interface{}{utils.MetaLocalHost},
		"AttrSConns":          []interface{}{utils.MetaLocalHost},
		"CDRsConns":           []interface{}{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCDRs)},
		"ReplicationConns":    []interface{}{},
		"MinDurLowBalance":    0.,
		"SessionIndexes":      map[string]interface{}{"OriginID": true},
		"ClientProtocol":      1.,
		"TerminateAttempts":   5.,
		"ChannelSyncInterval": 0.,
		"DebitInterval":       0.,
		"SessionTTL":          0.,
		"SessionTTLLastUsed":  nil,
		"SessionTTLMaxDelay":  nil,
		"SessionTTLUsage":     nil,
		"SessionTTLLastUsage": nil,
		"StoreSCosts":         false,
		"AlterableFields":     map[string]interface{}{},
		"STIRCfg": map[string]interface{}{
			"AllowedAttest": map[string]interface{}{
				utils.META_ANY: map[string]interface{}{},
			},
			"DefaultAttest":      "A",
			"PayloadMaxduration": -1.,
			"PrivateKeyPath":     "",
			"PublicKeyPath":      "",
		},
		"SchedulerConns": []interface{}{},
	}
	if *encoding == utils.MetaGOB {
		var empty []interface{}
		exp["ReplicationConns"] = empty
		exp["SchedulerConns"] = empty
		exp["StatSConns"] = empty
		exp["ThreshSConns"] = empty
	}
	var rpl map[string]interface{}
	if err := configRPC.Call(utils.ConfigSv1GetJSONSection, &config.SectionWithOpts{
		Tenant:  "cgrates.org",
		Section: config.SessionSJson,
	}, &rpl); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, rpl) {
		t.Errorf("Expected %+v , received: %+v ", utils.ToJSON(exp), utils.ToJSON(rpl))
	}
}

func testConfigSv1GetJSONSectionWithoutTenant(t *testing.T) {
	exp := map[string]interface{}{
		"Enabled":             true,
		"ListenBijson":        "127.0.0.1:2014",
		"ChargerSConns":       []interface{}{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers)},
		"RALsConns":           []interface{}{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResponder)},
		"ResSConns":           []interface{}{utils.MetaLocalHost},
		"ThreshSConns":        []interface{}{},
		"StatSConns":          []interface{}{},
		"RouteSConns":         []interface{}{utils.MetaLocalHost},
		"AttrSConns":          []interface{}{utils.MetaLocalHost},
		"CDRsConns":           []interface{}{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCDRs)},
		"ReplicationConns":    []interface{}{},
		"MinDurLowBalance":    0.,
		"SessionIndexes":      map[string]interface{}{"OriginID": true},
		"ClientProtocol":      1.,
		"TerminateAttempts":   5.,
		"ChannelSyncInterval": 0.,
		"DebitInterval":       0.,
		"SessionTTL":          0.,
		"SessionTTLLastUsed":  nil,
		"SessionTTLMaxDelay":  nil,
		"SessionTTLUsage":     nil,
		"SessionTTLLastUsage": nil,
		"StoreSCosts":         false,
		"AlterableFields":     map[string]interface{}{},
		"STIRCfg": map[string]interface{}{
			"AllowedAttest": map[string]interface{}{
				utils.META_ANY: map[string]interface{}{},
			},
			"DefaultAttest":      "A",
			"PayloadMaxduration": -1.,
			"PrivateKeyPath":     "",
			"PublicKeyPath":      "",
		},
		"SchedulerConns": []interface{}{},
	}
	var rpl map[string]interface{}
	if err := configRPC.Call(utils.ConfigSv1GetJSONSection, &config.SectionWithOpts{
		Section: config.SessionSJson,
	}, &rpl); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, rpl) {
		t.Errorf("Expected %+v , received: %+v ", utils.ToJSON(exp), utils.ToJSON(rpl))
	}
}

func testConfigSReloadConfigFromJSONEEs(t *testing.T) {
	if *encoding == utils.MetaGOB {
		t.SkipNow()
	}
	var reply string
	if err := configRPC.Call(utils.ConfigSv1ReloadConfigFromJSON, &config.JSONReloadWithOpts{
		JSON: map[string]interface{}{
			"ees": map[string]interface{}{
				"enabled":          true,
				"attributes_conns": []string{},
				"cache":            map[string]interface{}{},
				"exporters": []interface{}{map[string]interface{}{
					"id":     utils.MetaDefault,
					"fields": []interface{}{},
				}},
			},
		},
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %s", reply)
	}
	eporter := map[string]interface{}{
		"Attempts":      1.,
		"AttributeSCtx": "",
		"AttributeSIDs": []interface{}{},
		"ExportPath":    "/var/spool/cgrates/ees",
		"FieldSep":      ",",
		"Fields":        []interface{}{},
		"Filters":       []interface{}{},
		"Flags":         map[string]interface{}{},
		"ID":            "*default",
		"Synchronous":   false,
		"Tenant":        nil,
		"Timezone":      "",
		"Type":          "*none",
		"Opts":          map[string]interface{}{},
	}
	exp := map[string]interface{}{
		"Enabled":         true,
		"AttributeSConns": []interface{}{},
		"Cache":           map[string]interface{}{"*file_csv": map[string]interface{}{"Limit": -1., "Precache": false, "Replicate": false, "StaticTTL": false, "TTL": 5000000000.}},
		"Exporters":       []interface{}{eporter},
	}
	var rpl map[string]interface{}
	if err := configRPC.Call(utils.ConfigSv1GetJSONSection, &config.SectionWithOpts{
		Section: config.EEsJson,
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

func testConfigStartEngineWithConfigs(t *testing.T) {
	var err error
	configCfgPath = path.Join(*dataDir, "conf", "samples", "configs_active")
	configCfg, err = config.NewCGRConfigFromPath(configCfgPath)
	if err != nil {
		t.Error(err)
	}
	if _, err := engine.StopStartEngine(configCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
	configRPC, err = newRPCClient(configCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
	var rply map[string]interface{}
	if err := configRPC.Call(utils.CoreSv1Status, &utils.TenantWithOpts{}, &rply); err != nil {
		t.Error(err)
	} else if rply[utils.NodeID] != "EngineWithConfigSActive" {
		t.Errorf("Expected %+v , received: %+v ", "EngineWithConfigSActive", rply)
	}
}

func testConfigStartEngineFromHTTP(t *testing.T) {
	enginePath, err := exec.LookPath("cgr-engine")
	if err != nil {
		t.Error(err)
	}
	engine := exec.Command(enginePath, "-config_path", "http://127.0.0.1:3080/configs/tutmysql/cgrates.json")
	if err := engine.Start(); err != nil {
		t.Error(err)
	}
	fib := utils.Fib()
	var jsonClnt *rpc.Client
	var connected bool
	for i := 0; i < 200; i++ {
		time.Sleep(time.Duration(fib()) * time.Millisecond)
		if jsonClnt, err = jsonrpc.Dial(utils.TCP, "localhost:2012"); err != nil {
			utils.Logger.Warning(fmt.Sprintf("Error <%s> when opening test connection to: <%s>",
				err.Error(), "localhost:2012"))
		} else {
			connected = true
			break
		}
	}
	if !connected {
		t.Errorf("engine did not open port <%s>", "localhost:2012")
	}
	time.Sleep(time.Duration(500) * time.Millisecond)
	var rply map[string]interface{}
	if err := jsonClnt.Call(utils.CoreSv1Status, &utils.TenantWithOpts{}, &rply); err != nil {
		t.Error(err)
	}
}
