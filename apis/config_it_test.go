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

package apis

import (
	"path"
	"reflect"
	"testing"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	cfgPath string
	cfgCfg  *config.CGRConfig
	cfgRPC  *birpc.Client
	cfgDIR  string //run tests for specific configuration

	sTestsCfg = []func(t *testing.T){
		testCfgInitCfg,
		testCfgInitDataDb,
		testCfgResetStorDb,
		testCfgStartEngine,
		testCfgRPCConn,
		testCfgGetConfigInvalidSection,
		testCfgGetConfig,
		testCfgSetGetConfig,
		testCfgSetEmptyReload,
		testCfgSetJSONGetJSONConfig,
		testCfgKillEngine,
		//Store Cfg in Database Test
		testCfgInitCfgStore,
		testCfgInitCfgStore,
		testCfgInitDataDbStore,
		testCfgResetStorDbStore,
		testCfgStartEngineStore,
		testCfgRPCConnStore,
		testCfgKillEngineStore,
	}
)

func TestCfgSIT(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		cfgDIR = "apis_config_internal"
	case utils.MetaMongo:
		cfgDIR = "apis_config_mongo"
	case utils.MetaMySQL:
		cfgDIR = "apis_config_mysql"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsCfg {
		t.Run(cfgDIR, stest)
	}
}

func testCfgInitCfg(t *testing.T) {
	var err error
	cfgPath = path.Join(*dataDir, "conf", "samples", cfgDIR)
	cfgCfg, err = config.NewCGRConfigFromPath(cfgPath)
	if err != nil {
		t.Error(err)
	}
}

func testCfgInitDataDb(t *testing.T) {
	if err := engine.InitDataDB(cfgCfg); err != nil {
		t.Fatal(err)
	}
}

func testCfgResetStorDb(t *testing.T) {
	if err := engine.InitStorDB(cfgCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testCfgStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(cfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testCfgRPCConn(t *testing.T) {
	var err error
	cfgRPC, err = newRPCClient(cfgCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testCfgGetConfigInvalidSection(t *testing.T) {
	var reply map[string]interface{}
	expected := "Invalid section "
	if err := cfgRPC.Call(context.Background(), utils.ConfigSv1GetConfig,
		&config.SectionWithAPIOpts{
			APIOpts:  nil,
			Tenant:   utils.CGRateSorg,
			Sections: []string{"fakeSection"},
		},
		&reply); err == nil || err.Error() != expected {
		t.Error(err)
	}
}

func testCfgGetConfig(t *testing.T) {
	var reply map[string]interface{}
	expected := map[string]interface{}{
		"attributes": map[string]interface{}{
			"admins_conns":          []string{"*localhost"},
			"enabled":               true,
			"indexed_selects":       true,
			"nested_fields":         false,
			"prefix_indexed_fields": []string{},
			"process_runs":          1,
			"resources_conns":       []string{"*localhost"},
			"stats_conns":           []string{"*localhost"},
			"suffix_indexed_fields": []string{},
		},
	}
	if err := cfgRPC.Call(context.Background(), utils.ConfigSv1GetConfig,
		&config.SectionWithAPIOpts{
			APIOpts:  nil,
			Tenant:   utils.CGRateSorg,
			Sections: []string{"attributes"},
		},
		&reply); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(utils.ToJSON(expected), utils.ToJSON(reply)) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

func testCfgSetGetConfig(t *testing.T) {
	var reply string

	if err := cfgRPC.Call(context.Background(), utils.ConfigSv1SetConfig,
		&config.SetConfigArgs{
			APIOpts: nil,
			Tenant:  "",
			Config: map[string]interface{}{
				"attributes": map[string]interface{}{
					"admins_conns":          []string{"*internal"},
					"enabled":               true,
					"indexed_selects":       false,
					"nested_fields":         false,
					"prefix_indexed_fields": []string{},
					"process_runs":          2,
					"resources_conns":       []string{"*internal"},
					"stats_conns":           []string{"*internal"},
					"suffix_indexed_fields": []string{},
				},
			},
			DryRun: false,
		},
		&reply); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(`"OK"`, utils.ToJSON(reply)) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "OK", utils.ToJSON(reply))
	}
	expectedGet := map[string]interface{}{
		"attributes": map[string]interface{}{
			"admins_conns":          []string{"*internal"},
			"enabled":               true,
			"indexed_selects":       false,
			"nested_fields":         false,
			"prefix_indexed_fields": []string{},
			"process_runs":          2,
			"resources_conns":       []string{"*internal"},
			"stats_conns":           []string{"*internal"},
			"suffix_indexed_fields": []string{},
		},
	}
	var replyGet map[string]interface{}
	if err := cfgRPC.Call(context.Background(), utils.ConfigSv1GetConfig,
		&config.SectionWithAPIOpts{
			APIOpts:  nil,
			Tenant:   utils.CGRateSorg,
			Sections: []string{"attributes"},
		},
		&replyGet); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(utils.ToJSON(expectedGet), utils.ToJSON(replyGet)) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ToJSON(expectedGet), utils.ToJSON(replyGet))
	}
}

func testCfgSetEmptyReload(t *testing.T) {
	var reply string
	if err := cfgRPC.Call(context.Background(), utils.ConfigSv1SetConfig,
		&config.SetConfigArgs{
			APIOpts: nil,
			Tenant:  "",
			Config: map[string]interface{}{
				"rates": map[string]interface{}{
					"enabled":         true,
					"indexed_selects": false,
				},
			},
			DryRun: false,
		},
		&reply); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(`"OK"`, utils.ToJSON(reply)) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "OK", utils.ToJSON(reply))
	}
	var rldArgs string
	if err := cfgRPC.Call(context.Background(), utils.ConfigSv1ReloadConfig,
		&config.ReloadArgs{
			APIOpts: nil,
			Tenant:  "",
			Section: "rates",
			DryRun:  false,
		},
		&rldArgs); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(`"OK"`, utils.ToJSON(rldArgs)) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "OK", utils.ToJSON(rldArgs))
	}
	expectedGet := map[string]interface{}{
		"rates": map[string]interface{}{
			"enabled":                    true,
			"indexed_selects":            false,
			"nested_fields":              false,
			"prefix_indexed_fields":      []string{},
			"rate_indexed_selects":       true,
			"rate_nested_fields":         false,
			"rate_prefix_indexed_fields": []string{},
			"rate_suffix_indexed_fields": []string{},
			"suffix_indexed_fields":      []string{},
			"verbosity":                  1000,
		},
	}
	var replyGet map[string]interface{}
	if err := cfgRPC.Call(context.Background(), utils.ConfigSv1GetConfig,
		&config.SectionWithAPIOpts{
			APIOpts:  nil,
			Tenant:   utils.CGRateSorg,
			Sections: []string{"rates"},
		},
		&replyGet); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(utils.ToJSON(expectedGet), utils.ToJSON(replyGet)) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ToJSON(expectedGet), utils.ToJSON(replyGet))
	}
}

func testCfgSetJSONGetJSONConfig(t *testing.T) {
	var reply string

	if err := cfgRPC.Call(context.Background(), utils.ConfigSv1SetConfigFromJSON,
		&config.SetConfigFromJSONArgs{
			APIOpts: nil,
			Tenant:  "",
			Config:  "{\"attributes\":{\"admins_conns\":[\"*internal\"],\"enabled\":true,\"indexed_selects\":false,\"nested_fields\":false,\"prefix_indexed_fields\":[],\"process_runs\":2,\"resources_conns\":[\"*internal\"],\"stats_conns\":[\"*localhost\"],\"suffix_indexed_fields\":[]}}",
			DryRun:  false,
		},
		&reply); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(`"OK"`, utils.ToJSON(reply)) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "OK", utils.ToJSON(reply))
	}
	expectedGet := "{\"attributes\":{\"admins_conns\":[\"*internal\"],\"enabled\":true,\"indexed_selects\":false,\"nested_fields\":false,\"prefix_indexed_fields\":[],\"process_runs\":2,\"resources_conns\":[\"*internal\"],\"stats_conns\":[\"*localhost\"],\"suffix_indexed_fields\":[]}}"
	var replyGet string
	if err := cfgRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON,
		&config.SectionWithAPIOpts{
			APIOpts:  nil,
			Tenant:   utils.CGRateSorg,
			Sections: []string{"attributes"},
		},
		&replyGet); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(utils.ToJSON(expectedGet), utils.ToJSON(replyGet)) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ToJSON(expectedGet), utils.ToJSON(replyGet))
	}
}

func testCfgInitCfgStore(t *testing.T) {
	var err error
	cfgPath = path.Join(*dataDir, "conf", "samples", cfgDIR)
	cfgCfg, err = config.NewCGRConfigFromPath(cfgPath)
	if err != nil {
		t.Error(err)
	}
}

func testCfgKillEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}

func testCfgInitDataDbStore(t *testing.T) {
	if err := engine.InitDataDB(cfgCfg); err != nil {
		t.Fatal(err)
	}
}

func testCfgResetStorDbStore(t *testing.T) {
	if err := engine.InitStorDB(cfgCfg); err != nil {
		t.Fatal(err)
	}
}

func testCfgResetConfigDBStore(t *testing.T) {
	if err := engine.InitConfigDB(cfgCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testCfgStartEngineStore(t *testing.T) {
	if _, err := engine.StopStartEngine(cfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testCfgRPCConnStore(t *testing.T) {
	var err error
	cfgRPC, err = newRPCClient(cfgCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testCfgKillEngineStore(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
