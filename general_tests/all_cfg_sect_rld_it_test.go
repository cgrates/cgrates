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
	"net/rpc"
	"path"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	testSectCfgDir  string
	testSectCfgPath string
	testSectCfg     *config.CGRConfig
	testSectRPC     *rpc.Client

	testSectTests = []func(t *testing.T){
		testSectLoadConfig,
		testSectResetDataDB,
		testSectResetStorDb,
		testSectStartEngine,
		testSectRPCConn,
		testSectConfigSReloadResources,
		testSectConfigSReloadStats,
		testSectConfigSReloadThresholds,
		testSectConfigSReloadRoutes,
		testSectConfigSReloadLoaders,
		testSectConfigSReloadMailer,
		testSectConfigSReloadSuretax,
		testSectConfigSReloadLoader,
		testSectConfigSReloadMigrator,
		testSectConfigSReloadDispatchers,
		testSectConfigSReloadRegistrarC,
		testSectConfigSReloadAnalyzer,
		testSectConfigSReloadApiers,
		testSectConfigSReloadSIPAgent,
		testSectConfigSReloadTemplates,
		testSectConfigSReloadConfigs,
		testSectConfigSReloadAPIBan,
		testSectStopCgrEngine,
	}
)

func TestSectChange(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		testSectCfgDir = "tutinternal"
	case utils.MetaMySQL:
		testSectCfgDir = "tutmysql"
	case utils.MetaMongo:
		testSectCfgDir = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, testSectest := range testSectTests {
		t.Run(testSectCfgDir, testSectest)
	}
}

func testSectLoadConfig(t *testing.T) {
	testSectCfgPath = path.Join(*dataDir, "conf", "samples", testSectCfgDir)
	if testSectCfg, err = config.NewCGRConfigFromPath(testSectCfgPath); err != nil {
		t.Error(err)
	}
}

func testSectResetDataDB(t *testing.T) {
	if err := engine.InitDataDb(testSectCfg); err != nil {
		t.Fatal(err)
	}
}

func testSectResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(testSectCfg); err != nil {
		t.Fatal(err)
	}
}

func testSectStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(testSectCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testSectRPCConn(t *testing.T) {
	var err error
	testSectRPC, err = newRPCClient(testSectCfg.ListenCfg())
	if err != nil {
		t.Fatal(err)
	}
}

func testSectConfigSReloadResources(t *testing.T) {

	var reply string
	if err := testSectRPC.Call(utils.ConfigSv1ReloadConfig, &config.ReloadArgs{
		Tenant:  "cgrates.org",
		Path:    path.Join(*dataDir, "conf", "samples", "tutinternal"),
		Section: config.RESOURCES_JSON,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := "{\"resources\":{\"enabled\":true,\"indexed_selects\":true,\"nested_fields\":false,\"prefix_indexed_fields\":[],\"store_interval\":\"-1ns\",\"suffix_indexed_fields\":[],\"thresholds_conns\":[\"*internal\"]}}"
	var rpl string
	if err := testSectRPC.Call(utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:  "cgrates.org",
		Section: config.RESOURCES_JSON,
	}, &rpl); err != nil {
		t.Error(err)
	} else if cfgStr != rpl {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl))
	}
}

func testSectConfigSReloadStats(t *testing.T) {

	var reply string
	if err := testSectRPC.Call(utils.ConfigSv1ReloadConfig, &config.ReloadArgs{
		Tenant:  "cgrates.org",
		Path:    path.Join(*dataDir, "conf", "samples", "tutinternal"),
		Section: config.STATS_JSON,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := "{\"stats\":{\"enabled\":true,\"indexed_selects\":true,\"nested_fields\":false,\"prefix_indexed_fields\":[],\"store_interval\":\"-1ns\",\"store_uncompressed_limit\":0,\"suffix_indexed_fields\":[],\"thresholds_conns\":[\"*internal\"]}}"
	var rpl string
	if err := testSectRPC.Call(utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:  "cgrates.org",
		Section: config.STATS_JSON,
	}, &rpl); err != nil {
		t.Error(err)
	} else if cfgStr != rpl {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl))
	}
}

func testSectConfigSReloadThresholds(t *testing.T) {

	var reply string
	if err := testSectRPC.Call(utils.ConfigSv1ReloadConfig, &config.ReloadArgs{
		Tenant:  "cgrates.org",
		Path:    path.Join(*dataDir, "conf", "samples", "tutinternal"),
		Section: config.THRESHOLDS_JSON,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := "{\"thresholds\":{\"enabled\":true,\"indexed_selects\":true,\"nested_fields\":false,\"prefix_indexed_fields\":[],\"store_interval\":\"-1ns\",\"suffix_indexed_fields\":[]}}"
	var rpl string
	if err := testSectRPC.Call(utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:  "cgrates.org",
		Section: config.THRESHOLDS_JSON,
	}, &rpl); err != nil {
		t.Error(err)
	} else if cfgStr != rpl {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl))
	}
}

func testSectConfigSReloadRoutes(t *testing.T) {

	var reply string
	if err := testSectRPC.Call(utils.ConfigSv1ReloadConfig, &config.ReloadArgs{
		Tenant:  "cgrates.org",
		Path:    path.Join(*dataDir, "conf", "samples", "tutinternal"),
		Section: config.RouteSJson,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := "{\"routes\":{\"attributes_conns\":[],\"default_ratio\":1,\"enabled\":true,\"indexed_selects\":true,\"nested_fields\":false,\"prefix_indexed_fields\":[\"*req.Destination\"],\"rals_conns\":[\"*internal\"],\"resources_conns\":[\"*internal\"],\"stats_conns\":[\"*internal\"],\"suffix_indexed_fields\":[]}}"
	var rpl string
	if err := testSectRPC.Call(utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:  "cgrates.org",
		Section: config.RouteSJson,
	}, &rpl); err != nil {
		t.Error(err)
	} else if cfgStr != rpl {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl))
	}
}

func testSectConfigSReloadLoaders(t *testing.T) {

	var reply string
	if err := testSectRPC.Call(utils.ConfigSv1ReloadConfig, &config.ReloadArgs{
		Tenant:  "cgrates.org",
		Path:    path.Join(*dataDir, "conf", "samples", "tutinternal"),
		Section: config.LoaderJson,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := "{\"loaders\":[{\"caches_conns\":[\"*internal\"],\"data\":[{\"fields\":[{\"mandatory\":true,\"path\":\"Tenant\",\"tag\":\"TenantID\",\"type\":\"*variable\",\"value\":\"~*req.0\"},{\"mandatory\":true,\"path\":\"ID\",\"tag\":\"ProfileID\",\"type\":\"*variable\",\"value\":\"~*req.1\"},{\"path\":\"Contexts\",\"tag\":\"Contexts\",\"type\":\"*variable\",\"value\":\"~*req.2\"},{\"path\":\"FilterIDs\",\"tag\":\"FilterIDs\",\"type\":\"*variable\",\"value\":\"~*req.3\"},{\"path\":\"ActivationInterval\",\"tag\":\"ActivationInterval\",\"type\":\"*variable\",\"value\":\"~*req.4\"},{\"path\":\"AttributeFilterIDs\",\"tag\":\"AttributeFilterIDs\",\"type\":\"*variable\",\"value\":\"~*req.5\"},{\"path\":\"Path\",\"tag\":\"Path\",\"type\":\"*variable\",\"value\":\"~*req.6\"},{\"path\":\"Type\",\"tag\":\"Type\",\"type\":\"*variable\",\"value\":\"~*req.7\"},{\"path\":\"Value\",\"tag\":\"Value\",\"type\":\"*variable\",\"value\":\"~*req.8\"},{\"path\":\"Blocker\",\"tag\":\"Blocker\",\"type\":\"*variable\",\"value\":\"~*req.9\"},{\"path\":\"Weight\",\"tag\":\"Weight\",\"type\":\"*variable\",\"value\":\"~*req.10\"}],\"file_name\":\"Attributes.csv\",\"flags\":null,\"type\":\"*attributes\"},{\"fields\":[{\"mandatory\":true,\"path\":\"Tenant\",\"tag\":\"Tenant\",\"type\":\"*variable\",\"value\":\"~*req.0\"},{\"mandatory\":true,\"path\":\"ID\",\"tag\":\"ID\",\"type\":\"*variable\",\"value\":\"~*req.1\"},{\"path\":\"Type\",\"tag\":\"Type\",\"type\":\"*variable\",\"value\":\"~*req.2\"},{\"path\":\"Element\",\"tag\":\"Element\",\"type\":\"*variable\",\"value\":\"~*req.3\"},{\"path\":\"Values\",\"tag\":\"Values\",\"type\":\"*variable\",\"value\":\"~*req.4\"},{\"path\":\"ActivationInterval\",\"tag\":\"ActivationInterval\",\"type\":\"*variable\",\"value\":\"~*req.5\"}],\"file_name\":\"Filters.csv\",\"flags\":null,\"type\":\"*filters\"},{\"fields\":[{\"mandatory\":true,\"path\":\"Tenant\",\"tag\":\"Tenant\",\"type\":\"*variable\",\"value\":\"~*req.0\"},{\"mandatory\":true,\"path\":\"ID\",\"tag\":\"ID\",\"type\":\"*variable\",\"value\":\"~*req.1\"},{\"path\":\"FilterIDs\",\"tag\":\"FilterIDs\",\"type\":\"*variable\",\"value\":\"~*req.2\"},{\"path\":\"ActivationInterval\",\"tag\":\"ActivationInterval\",\"type\":\"*variable\",\"value\":\"~*req.3\"},{\"path\":\"UsageTTL\",\"tag\":\"TTL\",\"type\":\"*variable\",\"value\":\"~*req.4\"},{\"path\":\"Limit\",\"tag\":\"Limit\",\"type\":\"*variable\",\"value\":\"~*req.5\"},{\"path\":\"AllocationMessage\",\"tag\":\"AllocationMessage\",\"type\":\"*variable\",\"value\":\"~*req.6\"},{\"path\":\"Blocker\",\"tag\":\"Blocker\",\"type\":\"*variable\",\"value\":\"~*req.7\"},{\"path\":\"Stored\",\"tag\":\"Stored\",\"type\":\"*variable\",\"value\":\"~*req.8\"},{\"path\":\"Weight\",\"tag\":\"Weight\",\"type\":\"*variable\",\"value\":\"~*req.9\"},{\"path\":\"ThresholdIDs\",\"tag\":\"ThresholdIDs\",\"type\":\"*variable\",\"value\":\"~*req.10\"}],\"file_name\":\"Resources.csv\",\"flags\":null,\"type\":\"*resources\"},{\"fields\":[{\"mandatory\":true,\"path\":\"Tenant\",\"tag\":\"Tenant\",\"type\":\"*variable\",\"value\":\"~*req.0\"},{\"mandatory\":true,\"path\":\"ID\",\"tag\":\"ID\",\"type\":\"*variable\",\"value\":\"~*req.1\"},{\"path\":\"FilterIDs\",\"tag\":\"FilterIDs\",\"type\":\"*variable\",\"value\":\"~*req.2\"},{\"path\":\"ActivationInterval\",\"tag\":\"ActivationInterval\",\"type\":\"*variable\",\"value\":\"~*req.3\"},{\"path\":\"QueueLength\",\"tag\":\"QueueLength\",\"type\":\"*variable\",\"value\":\"~*req.4\"},{\"path\":\"TTL\",\"tag\":\"TTL\",\"type\":\"*variable\",\"value\":\"~*req.5\"},{\"path\":\"MinItems\",\"tag\":\"MinItems\",\"type\":\"*variable\",\"value\":\"~*req.6\"},{\"path\":\"MetricIDs\",\"tag\":\"MetricIDs\",\"type\":\"*variable\",\"value\":\"~*req.7\"},{\"path\":\"MetricFilterIDs\",\"tag\":\"MetricFilterIDs\",\"type\":\"*variable\",\"value\":\"~*req.8\"},{\"path\":\"Blocker\",\"tag\":\"Blocker\",\"type\":\"*variable\",\"value\":\"~*req.9\"},{\"path\":\"Stored\",\"tag\":\"Stored\",\"type\":\"*variable\",\"value\":\"~*req.10\"},{\"path\":\"Weight\",\"tag\":\"Weight\",\"type\":\"*variable\",\"value\":\"~*req.11\"},{\"path\":\"ThresholdIDs\",\"tag\":\"ThresholdIDs\",\"type\":\"*variable\",\"value\":\"~*req.12\"}],\"file_name\":\"Stats.csv\",\"flags\":null,\"type\":\"*stats\"},{\"fields\":[{\"mandatory\":true,\"path\":\"Tenant\",\"tag\":\"Tenant\",\"type\":\"*variable\",\"value\":\"~*req.0\"},{\"mandatory\":true,\"path\":\"ID\",\"tag\":\"ID\",\"type\":\"*variable\",\"value\":\"~*req.1\"},{\"path\":\"FilterIDs\",\"tag\":\"FilterIDs\",\"type\":\"*variable\",\"value\":\"~*req.2\"},{\"path\":\"ActivationInterval\",\"tag\":\"ActivationInterval\",\"type\":\"*variable\",\"value\":\"~*req.3\"},{\"path\":\"MaxHits\",\"tag\":\"MaxHits\",\"type\":\"*variable\",\"value\":\"~*req.4\"},{\"path\":\"MinHits\",\"tag\":\"MinHits\",\"type\":\"*variable\",\"value\":\"~*req.5\"},{\"path\":\"MinSleep\",\"tag\":\"MinSleep\",\"type\":\"*variable\",\"value\":\"~*req.6\"},{\"path\":\"Blocker\",\"tag\":\"Blocker\",\"type\":\"*variable\",\"value\":\"~*req.7\"},{\"path\":\"Weight\",\"tag\":\"Weight\",\"type\":\"*variable\",\"value\":\"~*req.8\"},{\"path\":\"ActionIDs\",\"tag\":\"ActionIDs\",\"type\":\"*variable\",\"value\":\"~*req.9\"},{\"path\":\"Async\",\"tag\":\"Async\",\"type\":\"*variable\",\"value\":\"~*req.10\"}],\"file_name\":\"Thresholds.csv\",\"flags\":null,\"type\":\"*thresholds\"},{\"fields\":[{\"mandatory\":true,\"path\":\"Tenant\",\"tag\":\"Tenant\",\"type\":\"*variable\",\"value\":\"~*req.0\"},{\"mandatory\":true,\"path\":\"ID\",\"tag\":\"ID\",\"type\":\"*variable\",\"value\":\"~*req.1\"},{\"path\":\"FilterIDs\",\"tag\":\"FilterIDs\",\"type\":\"*variable\",\"value\":\"~*req.2\"},{\"path\":\"ActivationInterval\",\"tag\":\"ActivationInterval\",\"type\":\"*variable\",\"value\":\"~*req.3\"},{\"path\":\"Sorting\",\"tag\":\"Sorting\",\"type\":\"*variable\",\"value\":\"~*req.4\"},{\"path\":\"SortingParameters\",\"tag\":\"SortingParameters\",\"type\":\"*variable\",\"value\":\"~*req.5\"},{\"path\":\"RouteID\",\"tag\":\"RouteID\",\"type\":\"*variable\",\"value\":\"~*req.6\"},{\"path\":\"RouteFilterIDs\",\"tag\":\"RouteFilterIDs\",\"type\":\"*variable\",\"value\":\"~*req.7\"},{\"path\":\"RouteAccountIDs\",\"tag\":\"RouteAccountIDs\",\"type\":\"*variable\",\"value\":\"~*req.8\"},{\"path\":\"RouteRatingPlanIDs\",\"tag\":\"RouteRatingPlanIDs\",\"type\":\"*variable\",\"value\":\"~*req.9\"},{\"path\":\"RouteResourceIDs\",\"tag\":\"RouteResourceIDs\",\"type\":\"*variable\",\"value\":\"~*req.10\"},{\"path\":\"RouteStatIDs\",\"tag\":\"RouteStatIDs\",\"type\":\"*variable\",\"value\":\"~*req.11\"},{\"path\":\"RouteWeight\",\"tag\":\"RouteWeight\",\"type\":\"*variable\",\"value\":\"~*req.12\"},{\"path\":\"RouteBlocker\",\"tag\":\"RouteBlocker\",\"type\":\"*variable\",\"value\":\"~*req.13\"},{\"path\":\"RouteParameters\",\"tag\":\"RouteParameters\",\"type\":\"*variable\",\"value\":\"~*req.14\"},{\"path\":\"Weight\",\"tag\":\"Weight\",\"type\":\"*variable\",\"value\":\"~*req.15\"}],\"file_name\":\"Routes.csv\",\"flags\":null,\"type\":\"*routes\"},{\"fields\":[{\"mandatory\":true,\"path\":\"Tenant\",\"tag\":\"Tenant\",\"type\":\"*variable\",\"value\":\"~*req.0\"},{\"mandatory\":true,\"path\":\"ID\",\"tag\":\"ID\",\"type\":\"*variable\",\"value\":\"~*req.1\"},{\"path\":\"FilterIDs\",\"tag\":\"FilterIDs\",\"type\":\"*variable\",\"value\":\"~*req.2\"},{\"path\":\"ActivationInterval\",\"tag\":\"ActivationInterval\",\"type\":\"*variable\",\"value\":\"~*req.3\"},{\"path\":\"RunID\",\"tag\":\"RunID\",\"type\":\"*variable\",\"value\":\"~*req.4\"},{\"path\":\"AttributeIDs\",\"tag\":\"AttributeIDs\",\"type\":\"*variable\",\"value\":\"~*req.5\"},{\"path\":\"Weight\",\"tag\":\"Weight\",\"type\":\"*variable\",\"value\":\"~*req.6\"}],\"file_name\":\"Chargers.csv\",\"flags\":null,\"type\":\"*chargers\"},{\"fields\":[{\"mandatory\":true,\"path\":\"Tenant\",\"tag\":\"Tenant\",\"type\":\"*variable\",\"value\":\"~*req.0\"},{\"mandatory\":true,\"path\":\"ID\",\"tag\":\"ID\",\"type\":\"*variable\",\"value\":\"~*req.1\"},{\"path\":\"Contexts\",\"tag\":\"Contexts\",\"type\":\"*variable\",\"value\":\"~*req.2\"},{\"path\":\"FilterIDs\",\"tag\":\"FilterIDs\",\"type\":\"*variable\",\"value\":\"~*req.3\"},{\"path\":\"ActivationInterval\",\"tag\":\"ActivationInterval\",\"type\":\"*variable\",\"value\":\"~*req.4\"},{\"path\":\"Strategy\",\"tag\":\"Strategy\",\"type\":\"*variable\",\"value\":\"~*req.5\"},{\"path\":\"StrategyParameters\",\"tag\":\"StrategyParameters\",\"type\":\"*variable\",\"value\":\"~*req.6\"},{\"path\":\"ConnID\",\"tag\":\"ConnID\",\"type\":\"*variable\",\"value\":\"~*req.7\"},{\"path\":\"ConnFilterIDs\",\"tag\":\"ConnFilterIDs\",\"type\":\"*variable\",\"value\":\"~*req.8\"},{\"path\":\"ConnWeight\",\"tag\":\"ConnWeight\",\"type\":\"*variable\",\"value\":\"~*req.9\"},{\"path\":\"ConnBlocker\",\"tag\":\"ConnBlocker\",\"type\":\"*variable\",\"value\":\"~*req.10\"},{\"path\":\"ConnParameters\",\"tag\":\"ConnParameters\",\"type\":\"*variable\",\"value\":\"~*req.11\"},{\"path\":\"Weight\",\"tag\":\"Weight\",\"type\":\"*variable\",\"value\":\"~*req.12\"}],\"file_name\":\"DispatcherProfiles.csv\",\"flags\":null,\"type\":\"*dispatchers\"},{\"fields\":[{\"mandatory\":true,\"path\":\"Tenant\",\"tag\":\"Tenant\",\"type\":\"*variable\",\"value\":\"~*req.0\"},{\"mandatory\":true,\"path\":\"ID\",\"tag\":\"ID\",\"type\":\"*variable\",\"value\":\"~*req.1\"},{\"path\":\"Address\",\"tag\":\"Address\",\"type\":\"*variable\",\"value\":\"~*req.2\"},{\"path\":\"Transport\",\"tag\":\"Transport\",\"type\":\"*variable\",\"value\":\"~*req.3\"},{\"path\":\"Synchronous\",\"tag\":\"Synchronous\",\"type\":\"*variable\",\"value\":\"~*req.4\"},{\"path\":\"ConnectAttempts\",\"tag\":\"ConnectAttempts\",\"type\":\"*variable\",\"value\":\"~*req.5\"},{\"path\":\"Reconnects\",\"tag\":\"Reconnects\",\"type\":\"*variable\",\"value\":\"~*req.6\"},{\"path\":\"ConnectTimeout\",\"tag\":\"ConnectTimeout\",\"type\":\"*variable\",\"value\":\"~*req.7\"},{\"path\":\"ReplyTimeout\",\"tag\":\"ReplyTimeout\",\"type\":\"*variable\",\"value\":\"~*req.8\"},{\"path\":\"TLS\",\"tag\":\"TLS\",\"type\":\"*variable\",\"value\":\"~*req.9\"},{\"path\":\"ClientKey\",\"tag\":\"ClientKey\",\"type\":\"*variable\",\"value\":\"~*req.10\"},{\"path\":\"ClientCertificate\",\"tag\":\"ClientCertificate\",\"type\":\"*variable\",\"value\":\"~*req.11\"},{\"path\":\"CaCertificate\",\"tag\":\"CaCertificate\",\"type\":\"*variable\",\"value\":\"~*req.12\"}],\"file_name\":\"DispatcherHosts.csv\",\"flags\":null,\"type\":\"*dispatcher_hosts\"}],\"dry_run\":false,\"enabled\":false,\"field_separator\":\",\",\"id\":\"*default\",\"lock_filename\":\".cgr.lck\",\"run_delay\":\"0\",\"tenant\":\"\",\"tp_in_dir\":\"/var/spool/cgrates/loader/in\",\"tp_out_dir\":\"/var/spool/cgrates/loader/out\"}]}"
	var rpl string
	if err := testSectRPC.Call(utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:  "cgrates.org",
		Section: config.LoaderJson,
	}, &rpl); err != nil {
		t.Error(err)
	} else if cfgStr != rpl {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl))
	}
}

func testSectConfigSReloadMailer(t *testing.T) {

	var reply string
	if err := testSectRPC.Call(utils.ConfigSv1ReloadConfig, &config.ReloadArgs{
		Tenant:  "cgrates.org",
		Path:    path.Join(*dataDir, "conf", "samples", "tutinternal"),
		Section: config.MAILER_JSN,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := "{\"mailer\":{\"auth_password\":\"CGRateS.org\",\"auth_user\":\"cgrates\",\"from_address\":\"cgr-mailer@localhost.localdomain\",\"server\":\"localhost\"}}"
	var rpl string
	if err := testSectRPC.Call(utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:  "cgrates.org",
		Section: config.MAILER_JSN,
	}, &rpl); err != nil {
		t.Error(err)
	} else if cfgStr != rpl {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl))
	}
}

func testSectConfigSReloadSuretax(t *testing.T) {

	var reply string
	if err := testSectRPC.Call(utils.ConfigSv1ReloadConfig, &config.ReloadArgs{
		Tenant:  "cgrates.org",
		Path:    path.Join(*dataDir, "conf", "samples", "tutinternal"),
		Section: config.SURETAX_JSON,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := "{\"suretax\":{\"bill_to_number\":\"\",\"business_unit\":\"\",\"client_number\":\"\",\"client_tracking\":\"~*req.CGRID\",\"customer_number\":\"~*req.Subject\",\"include_local_cost\":false,\"orig_number\":\"~*req.Subject\",\"p2pplus4\":\"\",\"p2pzipcode\":\"\",\"plus4\":\"\",\"regulatory_code\":\"03\",\"response_group\":\"03\",\"response_type\":\"D4\",\"return_file_code\":\"0\",\"sales_type_code\":\"R\",\"tax_exemption_code_list\":\"\",\"tax_included\":\"0\",\"tax_situs_rule\":\"04\",\"term_number\":\"~*req.Destination\",\"timezone\":\"Local\",\"trans_type_code\":\"010101\",\"unit_type\":\"00\",\"units\":\"1\",\"url\":\"\",\"validation_key\":\"\",\"zipcode\":\"\"}}"
	var rpl string
	if err := testSectRPC.Call(utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:  "cgrates.org",
		Section: config.SURETAX_JSON,
	}, &rpl); err != nil {
		t.Error(err)
	} else if cfgStr != rpl {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl))
	}
}

func testSectConfigSReloadLoader(t *testing.T) {

	var reply string
	if err := testSectRPC.Call(utils.ConfigSv1ReloadConfig, &config.ReloadArgs{
		Tenant:  "cgrates.org",
		Path:    path.Join(*dataDir, "conf", "samples", "tutinternal"),
		Section: config.CgrLoaderCfgJson,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := "{\"loader\":{\"caches_conns\":[\"*localhost\"],\"data_path\":\"./\",\"disable_reverse\":false,\"field_separator\":\",\",\"gapi_credentials\":\".gapi/credentials.json\",\"gapi_token\":\".gapi/token.json\",\"scheduler_conns\":[\"*localhost\"],\"tpid\":\"\"}}"
	var rpl string
	if err := testSectRPC.Call(utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:  "cgrates.org",
		Section: config.CgrLoaderCfgJson,
	}, &rpl); err != nil {
		t.Error(err)
	} else if cfgStr != rpl {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl))
	}
}

func testSectConfigSReloadMigrator(t *testing.T) {

	var reply string
	if err := testSectRPC.Call(utils.ConfigSv1ReloadConfig, &config.ReloadArgs{
		Tenant:  "cgrates.org",
		Path:    path.Join(*dataDir, "conf", "samples", "tutinternal"),
		Section: config.CgrMigratorCfgJson,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := "{\"migrator\":{\"out_datadb_encoding\":\"msgpack\",\"out_datadb_host\":\"127.0.0.1\",\"out_datadb_name\":\"10\",\"out_datadb_opts\":{\"redisCACertificate\":\"\",\"redisClientCertificate\":\"\",\"redisClientKey\":\"\",\"redisCluster\":false,\"redisClusterOndownDelay\":\"0\",\"redisClusterSync\":\"5s\",\"redisSentinel\":\"\",\"redisTLS\":false},\"out_datadb_password\":\"\",\"out_datadb_port\":\"6379\",\"out_datadb_type\":\"redis\",\"out_datadb_user\":\"cgrates\",\"out_stordb_host\":\"127.0.0.1\",\"out_stordb_name\":\"cgrates\",\"out_stordb_opts\":{},\"out_stordb_password\":\"\",\"out_stordb_port\":\"3306\",\"out_stordb_type\":\"mysql\",\"out_stordb_user\":\"cgrates\",\"users_filters\":[]}}"
	var rpl string
	if err := testSectRPC.Call(utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:  "cgrates.org",
		Section: config.CgrMigratorCfgJson,
	}, &rpl); err != nil {
		t.Error(err)
	} else if cfgStr != rpl {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl))
	}
}

func testSectConfigSReloadDispatchers(t *testing.T) {

	var reply string
	if err := testSectRPC.Call(utils.ConfigSv1ReloadConfig, &config.ReloadArgs{
		Tenant:  "cgrates.org",
		Path:    path.Join(*dataDir, "conf", "samples", "tutinternal"),
		Section: config.DispatcherSJson,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := "{\"dispatchers\":{\"any_subsystem\":true,\"attributes_conns\":[],\"enabled\":false,\"indexed_selects\":true,\"nested_fields\":false,\"prefix_indexed_fields\":[],\"suffix_indexed_fields\":[]}}"
	var rpl string
	if err := testSectRPC.Call(utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:  "cgrates.org",
		Section: config.DispatcherSJson,
	}, &rpl); err != nil {
		t.Error(err)
	} else if cfgStr != rpl {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl))
	}
}

func testSectConfigSReloadRegistrarC(t *testing.T) {

	var reply string
	if err := testSectRPC.Call(utils.ConfigSv1ReloadConfig, &config.ReloadArgs{
		Tenant:  "cgrates.org",
		Path:    path.Join(*dataDir, "conf", "samples", "tutinternal"),
		Section: config.RegistrarCJson,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := "{\"registrarc\":{\"dispatchers\":{\"hosts\":[],\"refresh_interval\":\"5m0s\",\"registrars_conns\":[]},\"rpc\":{\"hosts\":[],\"refresh_interval\":\"5m0s\",\"registrars_conns\":[]}}}"
	var rpl string
	if err := testSectRPC.Call(utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:  "cgrates.org",
		Section: config.RegistrarCJson,
	}, &rpl); err != nil {
		t.Error(err)
	} else if cfgStr != rpl {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl))
	}
}

func testSectConfigSReloadAnalyzer(t *testing.T) {

	var reply string
	if err := testSectRPC.Call(utils.ConfigSv1ReloadConfig, &config.ReloadArgs{
		Tenant:  "cgrates.org",
		Path:    path.Join(*dataDir, "conf", "samples", "tutinternal"),
		Section: config.AnalyzerCfgJson,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := "{\"analyzers\":{\"cleanup_interval\":\"1h0m0s\",\"db_path\":\"/var/spool/cgrates/analyzers\",\"enabled\":false,\"index_type\":\"*scorch\",\"ttl\":\"24h0m0s\"}}"
	var rpl string
	if err := testSectRPC.Call(utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:  "cgrates.org",
		Section: config.AnalyzerCfgJson,
	}, &rpl); err != nil {
		t.Error(err)
	} else if cfgStr != rpl {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl))
	}
}

func testSectConfigSReloadApiers(t *testing.T) {

	var reply string
	if err := testSectRPC.Call(utils.ConfigSv1ReloadConfig, &config.ReloadArgs{
		Tenant:  "cgrates.org",
		Path:    path.Join(*dataDir, "conf", "samples", "tutinternal"),
		Section: config.ApierS,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := "{\"apiers\":{\"attributes_conns\":[],\"caches_conns\":[\"*internal\"],\"ees_conns\":[],\"enabled\":true,\"scheduler_conns\":[\"*internal\"]}}"
	var rpl string
	if err := testSectRPC.Call(utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:  "cgrates.org",
		Section: config.ApierS,
	}, &rpl); err != nil {
		t.Error(err)
	} else if cfgStr != rpl {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl))
	}
}

func testSectConfigSReloadSIPAgent(t *testing.T) {

	var reply string
	if err := testSectRPC.Call(utils.ConfigSv1ReloadConfig, &config.ReloadArgs{
		Tenant:  "cgrates.org",
		Path:    path.Join(*dataDir, "conf", "samples", "tutinternal"),
		Section: config.SIPAgentJson,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := "{\"sip_agent\":{\"enabled\":false,\"listen\":\"127.0.0.1:5060\",\"listen_net\":\"udp\",\"request_processors\":[],\"retransmission_timer\":1000000000,\"sessions_conns\":[\"*internal\"],\"timezone\":\"\"}}"
	var rpl string
	if err := testSectRPC.Call(utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:  "cgrates.org",
		Section: config.SIPAgentJson,
	}, &rpl); err != nil {
		t.Error(err)
	} else if cfgStr != rpl {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl))
	}
}

func testSectConfigSReloadTemplates(t *testing.T) {

	var reply string
	if err := testSectRPC.Call(utils.ConfigSv1ReloadConfig, &config.ReloadArgs{
		Tenant:  "cgrates.org",
		Path:    path.Join(*dataDir, "conf", "samples", "tutinternal"),
		Section: config.TemplatesJson,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := "{\"templates\":{\"*asr\":[{\"mandatory\":true,\"path\":\"*diamreq.Session-Id\",\"tag\":\"SessionId\",\"type\":\"*variable\",\"value\":\"~*req.Session-Id\"},{\"mandatory\":true,\"path\":\"*diamreq.Origin-Host\",\"tag\":\"OriginHost\",\"type\":\"*variable\",\"value\":\"~*req.Destination-Host\"},{\"mandatory\":true,\"path\":\"*diamreq.Origin-Realm\",\"tag\":\"OriginRealm\",\"type\":\"*variable\",\"value\":\"~*req.Destination-Realm\"},{\"mandatory\":true,\"path\":\"*diamreq.Destination-Realm\",\"tag\":\"DestinationRealm\",\"type\":\"*variable\",\"value\":\"~*req.Origin-Realm\"},{\"mandatory\":true,\"path\":\"*diamreq.Destination-Host\",\"tag\":\"DestinationHost\",\"type\":\"*variable\",\"value\":\"~*req.Origin-Host\"},{\"mandatory\":true,\"path\":\"*diamreq.Auth-Application-Id\",\"tag\":\"AuthApplicationId\",\"type\":\"*variable\",\"value\":\"~*vars.*appid\"}],\"*cca\":[{\"mandatory\":true,\"path\":\"*rep.Session-Id\",\"tag\":\"SessionId\",\"type\":\"*variable\",\"value\":\"~*req.Session-Id\"},{\"path\":\"*rep.Result-Code\",\"tag\":\"ResultCode\",\"type\":\"*constant\",\"value\":\"2001\"},{\"mandatory\":true,\"path\":\"*rep.Origin-Host\",\"tag\":\"OriginHost\",\"type\":\"*variable\",\"value\":\"~*vars.OriginHost\"},{\"mandatory\":true,\"path\":\"*rep.Origin-Realm\",\"tag\":\"OriginRealm\",\"type\":\"*variable\",\"value\":\"~*vars.OriginRealm\"},{\"mandatory\":true,\"path\":\"*rep.Auth-Application-Id\",\"tag\":\"AuthApplicationId\",\"type\":\"*variable\",\"value\":\"~*vars.*appid\"},{\"mandatory\":true,\"path\":\"*rep.CC-Request-Type\",\"tag\":\"CCRequestType\",\"type\":\"*variable\",\"value\":\"~*req.CC-Request-Type\"},{\"mandatory\":true,\"path\":\"*rep.CC-Request-Number\",\"tag\":\"CCRequestNumber\",\"type\":\"*variable\",\"value\":\"~*req.CC-Request-Number\"}],\"*cdrLog\":[{\"mandatory\":true,\"path\":\"*cdr.ToR\",\"tag\":\"ToR\",\"type\":\"*variable\",\"value\":\"~*req.BalanceType\"},{\"mandatory\":true,\"path\":\"*cdr.OriginHost\",\"tag\":\"OriginHost\",\"type\":\"*constant\",\"value\":\"127.0.0.1\"},{\"mandatory\":true,\"path\":\"*cdr.RequestType\",\"tag\":\"RequestType\",\"type\":\"*constant\",\"value\":\"*none\"},{\"mandatory\":true,\"path\":\"*cdr.Tenant\",\"tag\":\"Tenant\",\"type\":\"*variable\",\"value\":\"~*req.Tenant\"},{\"mandatory\":true,\"path\":\"*cdr.Account\",\"tag\":\"Account\",\"type\":\"*variable\",\"value\":\"~*req.Account\"},{\"mandatory\":true,\"path\":\"*cdr.Subject\",\"tag\":\"Subject\",\"type\":\"*variable\",\"value\":\"~*req.Account\"},{\"mandatory\":true,\"path\":\"*cdr.Cost\",\"tag\":\"Cost\",\"type\":\"*variable\",\"value\":\"~*req.Cost\"},{\"mandatory\":true,\"path\":\"*cdr.Source\",\"tag\":\"Source\",\"type\":\"*constant\",\"value\":\"*cdrLog\"},{\"mandatory\":true,\"path\":\"*cdr.Usage\",\"tag\":\"Usage\",\"type\":\"*constant\",\"value\":\"1\"},{\"mandatory\":true,\"path\":\"*cdr.RunID\",\"tag\":\"RunID\",\"type\":\"*variable\",\"value\":\"~*req.ActionType\"},{\"mandatory\":true,\"path\":\"*cdr.SetupTime\",\"tag\":\"SetupTime\",\"type\":\"*constant\",\"value\":\"*now\"},{\"mandatory\":true,\"path\":\"*cdr.AnswerTime\",\"tag\":\"AnswerTime\",\"type\":\"*constant\",\"value\":\"*now\"},{\"mandatory\":true,\"path\":\"*cdr.PreRated\",\"tag\":\"PreRated\",\"type\":\"*constant\",\"value\":\"true\"}],\"*err\":[{\"mandatory\":true,\"path\":\"*rep.Session-Id\",\"tag\":\"SessionId\",\"type\":\"*variable\",\"value\":\"~*req.Session-Id\"},{\"mandatory\":true,\"path\":\"*rep.Origin-Host\",\"tag\":\"OriginHost\",\"type\":\"*variable\",\"value\":\"~*vars.OriginHost\"},{\"mandatory\":true,\"path\":\"*rep.Origin-Realm\",\"tag\":\"OriginRealm\",\"type\":\"*variable\",\"value\":\"~*vars.OriginRealm\"}],\"*errSip\":[{\"mandatory\":true,\"path\":\"*rep.Request\",\"tag\":\"Request\",\"type\":\"*constant\",\"value\":\"SIP/2.0 500 Internal Server Error\"}],\"*rar\":[{\"mandatory\":true,\"path\":\"*diamreq.Session-Id\",\"tag\":\"SessionId\",\"type\":\"*variable\",\"value\":\"~*req.Session-Id\"},{\"mandatory\":true,\"path\":\"*diamreq.Origin-Host\",\"tag\":\"OriginHost\",\"type\":\"*variable\",\"value\":\"~*req.Destination-Host\"},{\"mandatory\":true,\"path\":\"*diamreq.Origin-Realm\",\"tag\":\"OriginRealm\",\"type\":\"*variable\",\"value\":\"~*req.Destination-Realm\"},{\"mandatory\":true,\"path\":\"*diamreq.Destination-Realm\",\"tag\":\"DestinationRealm\",\"type\":\"*variable\",\"value\":\"~*req.Origin-Realm\"},{\"mandatory\":true,\"path\":\"*diamreq.Destination-Host\",\"tag\":\"DestinationHost\",\"type\":\"*variable\",\"value\":\"~*req.Origin-Host\"},{\"mandatory\":true,\"path\":\"*diamreq.Auth-Application-Id\",\"tag\":\"AuthApplicationId\",\"type\":\"*variable\",\"value\":\"~*vars.*appid\"},{\"path\":\"*diamreq.Re-Auth-Request-Type\",\"tag\":\"ReAuthRequestType\",\"type\":\"*constant\",\"value\":\"0\"}]}}"
	var rpl string
	if err := testSectRPC.Call(utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:  "cgrates.org",
		Section: config.TemplatesJson,
	}, &rpl); err != nil {
		t.Error(err)
	} else if cfgStr != rpl {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl))
	}
}

func testSectConfigSReloadConfigs(t *testing.T) {

	var reply string
	if err := testSectRPC.Call(utils.ConfigSv1ReloadConfig, &config.ReloadArgs{
		Tenant:  "cgrates.org",
		Path:    path.Join(*dataDir, "conf", "samples", "tutinternal"),
		Section: config.ConfigSJson,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := `{"configs":{"enabled":false,"root_dir":"/var/spool/cgrates/configs","url":"/configs/"}}`
	var rpl string
	if err := testSectRPC.Call(utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:  "cgrates.org",
		Section: config.ConfigSJson,
	}, &rpl); err != nil {
		t.Error(err)
	} else if cfgStr != rpl {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl))
	}
}

func testSectConfigSReloadAPIBan(t *testing.T) {

	var reply string
	if err := testSectRPC.Call(utils.ConfigSv1ReloadConfig, &config.ReloadArgs{
		Tenant:  "cgrates.org",
		Path:    path.Join(*dataDir, "conf", "samples", "tutinternal"),
		Section: config.APIBanCfgJson,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %s", reply)
	}
	cfgStr := `{"apiban":{"enabled":false,"keys":[]}}`
	var rpl string
	if err := testSectRPC.Call(utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:  "cgrates.org",
		Section: config.APIBanCfgJson,
	}, &rpl); err != nil {
		t.Error(err)
	} else if cfgStr != rpl {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl))
	}
}

func testSectStopCgrEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
