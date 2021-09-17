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
	connDb  engine.DataDBDriver
	cfgPath string
	cfgCfg  *config.CGRConfig
	cfgRPC  *birpc.Client
	cfgDIR  string //run tests for specific configuration

	sTestsCfg = []func(t *testing.T){
		testCfgInitCfg,
		testCfgInitDataDb,
		testCfgResetStorDb,
		testCfgResetConfigDBStore,
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
		testCfgInitDataDbStore,
		testCfgResetStorDbStore,
		testCfgResetConfigDBStore,
		testCfgStartEngineStore,
		testCfgRPCConnStore,
		testCfgDataDBConnStore,
		testCfgGetConfigStoreNil,
		testCfgStoreConfigStore,
		testCfgGetConfigStore,
		testCfgSetGetConfigStore,
		testCfgGetConfigStoreAgain,
		testCfgMdfSectConfigStore,
		testCfgReloadConfigStore,
		testCfgGetAfterReloadStore,
		testCfgKillEngineStore,
	}
)

func TestCfgSIT(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		t.SkipNow()
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
	cfgCfg, err = config.NewCGRConfigFromPath(context.Background(), cfgPath)
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
			"accounts_conns":        []string{"*localhost"},
			"enabled":               true,
			"indexed_selects":       true,
			"nested_fields":         false,
			"prefix_indexed_fields": []string{},
			"resources_conns":       []string{"*localhost"},
			"stats_conns":           []string{"*localhost"},
			"suffix_indexed_fields": []string{},
			utils.OptsCfg: map[string]interface{}{
				utils.MetaAttributeIDsCfg: []string(nil),
				utils.MetaProcessRunsCfg:  1,
				utils.MetaProfileRunsCfg:  0,
			},
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
					"accounts_conns":        []string{"*internal"},
					"enabled":               true,
					"indexed_selects":       false,
					"nested_fields":         false,
					"prefix_indexed_fields": []string{},
					"resources_conns":       []string{"*internal"},
					"stats_conns":           []string{"*internal"},
					"profile_runs":          0.,
					"suffix_indexed_fields": []string{},
					utils.OptsCfg: map[string]interface{}{
						utils.MetaProcessRunsCfg: 2,
					},
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
			"accounts_conns":        []string{"*internal"},
			"enabled":               true,
			"indexed_selects":       false,
			"nested_fields":         false,
			"prefix_indexed_fields": []string{},
			"resources_conns":       []string{"*internal"},
			"stats_conns":           []string{"*internal"},
			"suffix_indexed_fields": []string{},
			utils.OptsCfg: map[string]interface{}{
				utils.MetaAttributeIDsCfg: []string(nil),
				utils.MetaProcessRunsCfg:  2,
				utils.MetaProfileRunsCfg:  0,
			},
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
			utils.OptsCfg: map[string]interface{}{
				utils.MetaRateProfileIDsCfg: []string(nil),
			},
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
			Config:  "{\"attributes\":{\"accounts_conns\":[\"*internal\"],\"enabled\":true,\"indexed_selects\":false,\"nested_fields\":false,\"prefix_indexed_fields\":[],\"resources_conns\":[\"*internal\"],\"stats_conns\":[\"*localhost\"],\"suffix_indexed_fields\":[],\"opts\":{\"*processRuns\":2}}}",
			DryRun:  false,
		},
		&reply); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(`"OK"`, utils.ToJSON(reply)) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "OK", utils.ToJSON(reply))
	}
	expectedGet := "{\"attributes\":{\"accounts_conns\":[\"*internal\"],\"enabled\":true,\"indexed_selects\":false,\"nested_fields\":false,\"opts\":{\"*attributeIDs\":null,\"*processRuns\":2,\"*profileRuns\":0},\"prefix_indexed_fields\":[],\"resources_conns\":[\"*internal\"],\"stats_conns\":[\"*localhost\"],\"suffix_indexed_fields\":[]}}"
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
	cfgCfg, err = config.NewCGRConfigFromPath(context.Background(), cfgPath)
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

func testCfgDataDBConnStore(t *testing.T) {
	var err error
	connDb, err = engine.NewDataDBConn(cfgCfg.ConfigDBCfg().Type,
		cfgCfg.ConfigDBCfg().Host, cfgCfg.ConfigDBCfg().Port,
		cfgCfg.ConfigDBCfg().Name, cfgCfg.ConfigDBCfg().User,
		cfgCfg.ConfigDBCfg().Password, cfgCfg.GeneralCfg().DBDataEncoding,
		cfgCfg.ConfigDBCfg().Opts)
	if err != nil {
		t.Fatal(err)
	}
}

func testCfgGetConfigStoreNil(t *testing.T) {
	attr := new(config.AttributeSJsonCfg)
	if err := connDb.GetSection(context.Background(), config.AttributeSJSON, attr); err != nil {
		t.Fatal(err)
	}
	expected := new(config.AttributeSJsonCfg)
	if !reflect.DeepEqual(attr, expected) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, attr)
	}
}

func testCfgStoreConfigStore(t *testing.T) {
	var reply string
	if err := cfgRPC.Call(context.Background(), utils.ConfigSv1StoreCfgInDB,
		&config.SectionWithAPIOpts{
			APIOpts:  nil,
			Tenant:   utils.CGRateSorg,
			Sections: []string{"attributes"},
		},
		&reply); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual("OK", reply) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "OK", utils.ToJSON(reply))
	}

}

func testCfgGetConfigStore(t *testing.T) {
	attr := new(config.AttributeSJsonCfg)
	if err := connDb.GetSection(context.Background(), config.AttributeSJSON, attr); err != nil {
		t.Fatal(err)
	}
	expected := &config.AttributeSJsonCfg{
		Enabled:               utils.BoolPointer(true),
		Stats_conns:           &[]string{"*localhost"},
		Resources_conns:       &[]string{"*localhost"},
		Accounts_conns:        &[]string{"*localhost"},
		Indexed_selects:       nil,
		String_indexed_fields: nil,
		Prefix_indexed_fields: nil,
		Suffix_indexed_fields: nil,
		Nested_fields:         nil,
		Opts:                  &config.AttributesOptsJson{},
	}
	if !reflect.DeepEqual(attr, expected) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ToJSON(expected), utils.ToJSON(attr))
	}
}

func testCfgSetGetConfigStore(t *testing.T) {
	var reply string

	if err := cfgRPC.Call(context.Background(), utils.ConfigSv1SetConfig,
		&config.SetConfigArgs{
			APIOpts: nil,
			Tenant:  "",
			Config: map[string]interface{}{
				"attributes": map[string]interface{}{
					"accounts_conns":        []string{"*internal"},
					"enabled":               true,
					"indexed_selects":       false,
					"nested_fields":         false,
					"prefix_indexed_fields": []string{},
					"resources_conns":       []string{"*internal"},
					"stats_conns":           []string{"*internal"},
					"profile_runs":          0.,
					"suffix_indexed_fields": []string{},
					utils.OptsCfg: map[string]interface{}{
						utils.MetaProcessRunsCfg: 2,
					},
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
			"accounts_conns":        []string{"*internal"},
			"enabled":               true,
			"indexed_selects":       false,
			"nested_fields":         false,
			"prefix_indexed_fields": []string{},
			"resources_conns":       []string{"*internal"},
			"stats_conns":           []string{"*internal"},
			"suffix_indexed_fields": []string{},
			utils.OptsCfg: map[string]interface{}{
				utils.MetaAttributeIDsCfg: []string(nil),
				utils.MetaProcessRunsCfg:  2,
				utils.MetaProfileRunsCfg:  0,
			},
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

func testCfgGetConfigStoreAgain(t *testing.T) {
	attr := new(config.AttributeSJsonCfg)
	if err := connDb.GetSection(context.Background(), config.AttributeSJSON, attr); err != nil {
		t.Fatal(err)
	}
	expected := &config.AttributeSJsonCfg{
		Enabled:               utils.BoolPointer(true),
		Stats_conns:           &[]string{"*internal"},
		Resources_conns:       &[]string{"*internal"},
		Accounts_conns:        &[]string{"*internal"},
		Indexed_selects:       utils.BoolPointer(false),
		String_indexed_fields: nil,
		Prefix_indexed_fields: nil,
		Suffix_indexed_fields: nil,
		Nested_fields:         nil,
		Opts: &config.AttributesOptsJson{
			ProcessRuns: utils.IntPointer(2),
		},
	}
	if !reflect.DeepEqual(attr, expected) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ToJSON(expected), utils.ToJSON(attr))
	}
}

func testCfgMdfSectConfigStore(t *testing.T) {
	attrSect := &config.AttributeSJsonCfg{
		Enabled:               utils.BoolPointer(true),
		Stats_conns:           &[]string{"*internal"},
		Resources_conns:       &[]string{"*internal"},
		Accounts_conns:        &[]string{"*internal"},
		Indexed_selects:       utils.BoolPointer(true),
		String_indexed_fields: nil,
		Prefix_indexed_fields: nil,
		Suffix_indexed_fields: nil,
		Nested_fields:         nil,
		Opts: &config.AttributesOptsJson{
			ProcessRuns: utils.IntPointer(2),
		},
	}
	err := connDb.SetSection(context.Background(), "attributes", attrSect)

	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
}

func testCfgReloadConfigStore(t *testing.T) {
	var rldArgs string
	if err := cfgRPC.Call(context.Background(), utils.ConfigSv1ReloadConfig,
		&config.ReloadArgs{
			APIOpts: nil,
			Tenant:  "",
			Section: "attributes",
			DryRun:  false,
		},
		&rldArgs); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(`"OK"`, utils.ToJSON(rldArgs)) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "OK", utils.ToJSON(rldArgs))
	}
}

func testCfgGetAfterReloadStore(t *testing.T) {
	expectedGet := map[string]interface{}{
		"attributes": map[string]interface{}{
			"accounts_conns":        []string{"*internal"},
			"enabled":               true,
			"indexed_selects":       true,
			"nested_fields":         false,
			"prefix_indexed_fields": []string{},
			"resources_conns":       []string{"*internal"},
			"stats_conns":           []string{"*internal"},
			"suffix_indexed_fields": []string{},
			utils.OptsCfg: map[string]interface{}{
				utils.MetaAttributeIDsCfg: []string(nil),
				utils.MetaProcessRunsCfg:  2,
				utils.MetaProfileRunsCfg:  0,
			},
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

func testCfgKillEngineStore(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
