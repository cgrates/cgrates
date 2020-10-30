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
		"enabled":               true,
		"channel_sync_interval": "0",
		"alterable_fields":      []interface{}{},
		"client_protocol":       1.,
		"debit_interval":        "0",
		"listen_bijson":         "127.0.0.1:2014",
		"session_ttl":           "0",
		"session_indexes":       []interface{}{utils.OriginID},
		"attributes_conns":      []interface{}{utils.MetaLocalHost},
		"cdrs_conns":            []interface{}{utils.MetaInternal},
		"chargers_conns":        []interface{}{utils.MetaInternal},
		"rals_conns":            []interface{}{utils.MetaInternal},
		"replication_conns":     []interface{}{},
		"resources_conns":       []interface{}{utils.MetaLocalHost},
		"routes_conns":          []interface{}{utils.MetaLocalHost},
		"scheduler_conns":       []interface{}{},
		"thresholds_conns":      []interface{}{},
		"stats_conns":           []interface{}{},
		"min_dur_low_balance":   "0",
		"stir": map[string]interface{}{
			"allowed_attest":      []interface{}{utils.META_ANY},
			"default_attest":      "A",
			"payload_maxduration": "-1",
			"privatekey_path":     "",
			"publickey_path":      "",
		},
		"store_session_costs": false,
		"terminate_attempts":  5.,
	}
	if *encoding == utils.MetaGOB {
		var empty []interface{}
		exp["replication_conns"] = empty
		exp["scheduler_conns"] = empty
		exp["stats_conns"] = empty
		exp["thresholds_conns"] = empty
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
		"enabled":               true,
		"listen_bijson":         "127.0.0.1:2014",
		"chargers_conns":        []interface{}{utils.MetaInternal},
		"rals_conns":            []interface{}{utils.MetaInternal},
		"resources_conns":       []interface{}{utils.MetaLocalHost},
		"thresholds_conns":      []interface{}{},
		"stats_conns":           []interface{}{},
		"routes_conns":          []interface{}{utils.MetaLocalHost},
		"attributes_conns":      []interface{}{utils.MetaLocalHost},
		"cdrs_conns":            []interface{}{utils.MetaInternal},
		"replication_conns":     []interface{}{},
		"scheduler_conns":       []interface{}{},
		"session_indexes":       []interface{}{"OriginID"},
		"client_protocol":       1.,
		"terminate_attempts":    5.,
		"channel_sync_interval": "0",
		"debit_interval":        "0",
		"session_ttl":           "0",
		"store_session_costs":   false,
		"min_dur_low_balance":   "0",
		"alterable_fields":      []interface{}{},
		"stir": map[string]interface{}{
			"allowed_attest":      []interface{}{utils.META_ANY},
			"default_attest":      "A",
			"payload_maxduration": "-1",
			"privatekey_path":     "",
			"publickey_path":      "",
		},
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
		"attempts":          1.,
		"attribute_context": "",
		"attribute_ids":     []interface{}{},
		"export_path":       "/var/spool/cgrates/ees",
		"field_separator":   ",",
		"fields":            []interface{}{},
		"filters":           []interface{}{},
		"flags":             []interface{}{},
		"id":                "*default",
		"synchronous":       false,
		"tenant":            "",
		"timezone":          "",
		"type":              "*none",
		"opts":              map[string]interface{}{},
	}
	exp := map[string]interface{}{
		"enabled":          true,
		"attributes_conns": []interface{}{},
		"cache":            map[string]interface{}{"*file_csv": map[string]interface{}{"limit": -1., "precache": false, "replicate": false, "static_ttl": false, "ttl": "5s"}},
		"exporters":        []interface{}{eporter},
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
	time.Sleep(500 * time.Millisecond)
	var rply map[string]interface{}
	if err := jsonClnt.Call(utils.CoreSv1Status, &utils.TenantWithOpts{}, &rply); err != nil {
		t.Error(err)
	}
}
