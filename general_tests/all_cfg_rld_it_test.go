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
	"testing"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	testCfgDir  string
	testCfgPath string
	testCfg     *config.CGRConfig
	testRPC     *birpc.Client

	testTestsR = []func(t *testing.T){
		testCfgLoadConfig,
		testResetDataDB,
		testResetStorDb,
		testStartEngine,
		testRPCConn,
		testConfigSReload,
		testStopCgrEngine,
	}
)

func TestRldCfg(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		testCfgDir = "tutinternal"
	case utils.MetaMySQL:
		testCfgDir = "tutmysql"
	case utils.MetaMongo:
		testCfgDir = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, testRld := range testTestsR {
		t.Run(testCfgDir, testRld)
	}
}

func testCfgLoadConfig(t *testing.T) {
	testCfgPath = path.Join(*dataDir, "conf", "samples", testCfgDir)
	if testCfg, err = config.NewCGRConfigFromPath(testCfgPath); err != nil {
		t.Error(err)
	}
}

func testResetDataDB(t *testing.T) {
	if err := engine.InitDataDB(testCfg); err != nil {
		t.Fatal(err)
	}
}

func testResetStorDb(t *testing.T) {
	if err := engine.InitStorDB(testCfg); err != nil {
		t.Fatal(err)
	}
}

func testStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(testCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testRPCConn(t *testing.T) {
	var err error
	testRPC, err = newRPCClient(testCfg.ListenCfg())
	if err != nil {
		t.Fatal(err)
	}
}

func testConfigSReload(t *testing.T) {

	var reply string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1ReloadConfig, &config.ReloadArgs{
		Tenant:  "cgrates.org",
		Section: utils.MetaAll,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %s", reply)
	}

	cfgStr := "{\"cores\":{\"caps\":0,\"caps_stats_interval\":\"0\",\"caps_strategy\":\"*busy\",\"shutdown_timeout\":\"1s\"}}"
	var rpl2 string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.CoreSJSON},
	}, &rpl2); err != nil {
		t.Error(err)
	} else if cfgStr != rpl2 {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl2))
	}

	cfgStr = "{\"rpc_conns\":{\"*bijson_localhost\":{\"conns\":[{\"address\":\"127.0.0.1:2014\",\"transport\":\"*birpc_json\"}],\"poolSize\":0,\"strategy\":\"*first\"},\"*birpc_internal\":{\"conns\":[{\"address\":\"*birpc_internal\",\"transport\":\"\"}],\"poolSize\":0,\"strategy\":\"*first\"},\"*internal\":{\"conns\":[{\"address\":\"*internal\",\"transport\":\"\"}],\"poolSize\":0,\"strategy\":\"*first\"},\"*localhost\":{\"conns\":[{\"address\":\"127.0.0.1:2012\",\"transport\":\"*json\"}],\"poolSize\":0,\"strategy\":\"*first\"}}}"
	var rpl string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.RPCConnsJSON},
	}, &rpl); err != nil {
		t.Error(err)
	} else if cfgStr != rpl {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl))
	}

	cfgStr = "{\"listen\":{\"http\":\":2080\",\"http_tls\":\"127.0.0.1:2280\",\"rpc_gob\":\":2013\",\"rpc_gob_tls\":\"127.0.0.1:2023\",\"rpc_json\":\":2012\",\"rpc_json_tls\":\"127.0.0.1:2022\"}}"
	var rpl3 string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.ListenJSON},
	}, &rpl3); err != nil {
		t.Error(err)
	} else if cfgStr != rpl3 {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl3))
	}

	cfgStr = "{\"tls\":{\"ca_certificate\":\"\",\"client_certificate\":\"\",\"client_key\":\"\",\"server_certificate\":\"\",\"server_key\":\"\",\"server_name\":\"\",\"server_policy\":4}}"
	var rpl4 string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.TlsJSON},
	}, &rpl4); err != nil {
		t.Error(err)
	} else if cfgStr != rpl4 {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl4))
	}

	cfgStr = "{\"http\":{\"auth_users\":{},\"client_opts\":{\"dialFallbackDelay\":\"300ms\",\"dialKeepAlive\":\"30s\",\"dialTimeout\":\"30s\",\"disableCompression\":false,\"disableKeepAlives\":false,\"expectContinueTimeout\":\"0\",\"forceAttemptHttp2\":true,\"idleConnTimeout\":\"90s\",\"maxConnsPerHost\":0,\"maxIdleConns\":100,\"maxIdleConnsPerHost\":2,\"responseHeaderTimeout\":\"0\",\"skipTlsVerify\":false,\"tlsHandshakeTimeout\":\"10s\"},\"freeswitch_cdrs_url\":\"/freeswitch_json\",\"http_cdrs\":\"/cdr_http\",\"json_rpc_url\":\"/jsonrpc\",\"registrars_url\":\"/registrar\",\"use_basic_auth\":false,\"ws_url\":\"/ws\"}}"

	var rpl5 string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.HTTPJSON},
	}, &rpl5); err != nil {
		t.Error(err)
	} else if cfgStr != rpl5 {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl5))
	}

	cfgStr = "{\"caches\":{\"partitions\":{\"*account_action_plans\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"\"},\"*accounts\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"\"},\"*action_plans\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"\"},\"*action_triggers\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"\"},\"*actions\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"\"},\"*apiban\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"2m0s\"},\"*attribute_filter_indexes\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"\"},\"*attribute_profiles\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"\"},\"*caps_events\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"\"},\"*cdr_ids\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"10m0s\"},\"*cdrs\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"\"},\"*charger_filter_indexes\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"\"},\"*charger_profiles\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"\"},\"*closed_sessions\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"10s\"},\"*destinations\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"\"},\"*diameter_messages\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"3h0m0s\"},\"*dispatcher_filter_indexes\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"\"},\"*dispatcher_hosts\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"\"},\"*dispatcher_loads\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"\"},\"*dispatcher_profiles\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"\"},\"*dispatcher_routes\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"\"},\"*dispatchers\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"\"},\"*event_charges\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"10s\"},\"*event_resources\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"\"},\"*filters\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"\"},\"*load_ids\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"\"},\"*rating_plans\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"\"},\"*rating_profiles\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"\"},\"*replication_hosts\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"\"},\"*resource_filter_indexes\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"\"},\"*resource_profiles\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"\"},\"*resources\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"\"},\"*reverse_destinations\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"\"},\"*reverse_filter_indexes\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"\"},\"*route_filter_indexes\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"\"},\"*route_profiles\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"\"},\"*rpc_connections\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"\"},\"*rpc_responses\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"2s\"},\"*session_costs\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"\"},\"*shared_groups\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"\"},\"*stat_filter_indexes\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"\"},\"*statqueue_profiles\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"\"},\"*statqueues\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"\"},\"*stir\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"3h0m0s\"},\"*threshold_filter_indexes\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"\"},\"*threshold_profiles\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"\"},\"*thresholds\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"\"},\"*timings\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"\"},\"*tmp_rating_profiles\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"1m0s\"},\"*tp_account_actions\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"\"},\"*tp_action_plans\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"\"},\"*tp_action_triggers\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"\"},\"*tp_actions\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"\"},\"*tp_attributes\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"\"},\"*tp_chargers\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"\"},\"*tp_destination_rates\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"\"},\"*tp_destinations\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"\"},\"*tp_dispatcher_hosts\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"\"},\"*tp_dispatcher_profiles\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"\"},\"*tp_filters\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"\"},\"*tp_rates\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"\"},\"*tp_rating_plans\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"\"},\"*tp_rating_profiles\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"\"},\"*tp_resources\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"\"},\"*tp_routes\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"\"},\"*tp_shared_groups\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"\"},\"*tp_stats\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"\"},\"*tp_thresholds\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"\"},\"*tp_timings\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"\"},\"*uch\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"3h0m0s\"},\"*versions\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"\"}},\"replication_conns\":[]}}"
	var rpl7 string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.CacheJSON},
	}, &rpl7); err != nil {
		t.Error(err)
	} else if cfgStr != rpl7 {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl7))
	}

	cfgStr = "{\"filters\":{\"apiers_conns\":[\"*internal\"],\"resources_conns\":[\"*internal\"],\"stats_conns\":[\"*localhost\"]}}"
	var rpl8 string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.FilterSJSON},
	}, &rpl8); err != nil {
		t.Error(err)
	} else if cfgStr != rpl8 {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl8))
	}

	cfgStr = "{\"cdrs\":{\"attributes_conns\":[],\"chargers_conns\":[\"*internal\"],\"ees_conns\":[],\"enabled\":true,\"extra_fields\":[],\"online_cdr_exports\":[],\"rals_conns\":[],\"scheduler_conns\":[],\"session_cost_retries\":5,\"stats_conns\":[],\"store_cdrs\":true,\"thresholds_conns\":[]}}"
	var rpl10 string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.CDRsJSON},
	}, &rpl10); err != nil {
		t.Error(err)
	} else if cfgStr != rpl10 {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl10))
	}
	cfgStr = "{\"ers\":{\"enabled\":false,\"partial_cache_ttl\":\"1s\",\"readers\":[{\"cache_dump_fields\":[],\"concurrent_requests\":1024,\"fields\":[{\"mandatory\":true,\"path\":\"*cgreq.ToR\",\"tag\":\"ToR\",\"type\":\"*variable\",\"value\":\"~*req.2\"},{\"mandatory\":true,\"path\":\"*cgreq.OriginID\",\"tag\":\"OriginID\",\"type\":\"*variable\",\"value\":\"~*req.3\"},{\"mandatory\":true,\"path\":\"*cgreq.RequestType\",\"tag\":\"RequestType\",\"type\":\"*variable\",\"value\":\"~*req.4\"},{\"mandatory\":true,\"path\":\"*cgreq.Tenant\",\"tag\":\"Tenant\",\"type\":\"*variable\",\"value\":\"~*req.6\"},{\"mandatory\":true,\"path\":\"*cgreq.Category\",\"tag\":\"Category\",\"type\":\"*variable\",\"value\":\"~*req.7\"},{\"mandatory\":true,\"path\":\"*cgreq.Account\",\"tag\":\"Account\",\"type\":\"*variable\",\"value\":\"~*req.8\"},{\"mandatory\":true,\"path\":\"*cgreq.Subject\",\"tag\":\"Subject\",\"type\":\"*variable\",\"value\":\"~*req.9\"},{\"mandatory\":true,\"path\":\"*cgreq.Destination\",\"tag\":\"Destination\",\"type\":\"*variable\",\"value\":\"~*req.10\"},{\"mandatory\":true,\"path\":\"*cgreq.SetupTime\",\"tag\":\"SetupTime\",\"type\":\"*variable\",\"value\":\"~*req.11\"},{\"mandatory\":true,\"path\":\"*cgreq.AnswerTime\",\"tag\":\"AnswerTime\",\"type\":\"*variable\",\"value\":\"~*req.12\"},{\"mandatory\":true,\"path\":\"*cgreq.Usage\",\"tag\":\"Usage\",\"type\":\"*variable\",\"value\":\"~*req.13\"}],\"filters\":[],\"flags\":[],\"id\":\"*default\",\"opts\":{\"csvFieldSeparator\":\",\",\"csvHeaderDefineChar\":\":\",\"csvRowLength\":0,\"natsSubject\":\"cgrates_cdrs\",\"partialCacheAction\":\"*none\",\"partialOrderField\":\"~*req.AnswerTime\",\"xmlRootPath\":\"\"},\"partial_commit_fields\":[],\"processed_path\":\"/var/spool/cgrates/ers/out\",\"run_delay\":\"0\",\"source_path\":\"/var/spool/cgrates/ers/in\",\"tenant\":\"\",\"timezone\":\"\",\"type\":\"*none\"}],\"sessions_conns\":[\"*internal\"]}}"
	var rpl11 string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.ERsJSON},
	}, &rpl11); err != nil {
		t.Error(err)
	} else if cfgStr != rpl11 {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl11))
	}
	cfgStr = "{\"ees\":{\"attributes_conns\":[],\"cache\":{\"*file_csv\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"5s\"}},\"enabled\":false,\"exporters\":[{\"attempts\":1,\"attribute_context\":\"\",\"attribute_ids\":[],\"concurrent_requests\":0,\"export_path\":\"/var/spool/cgrates/ees\",\"fields\":[],\"filters\":[],\"flags\":[],\"id\":\"*default\",\"opts\":{},\"synchronous\":false,\"timezone\":\"\",\"type\":\"*none\"}]}}"
	var rpl12 string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.EEsJSON},
	}, &rpl12); err != nil {
		t.Error(err)
	} else if cfgStr != rpl12 {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl12))
	}
	cfgStr = "{\"sessions\":{\"alterable_fields\":[],\"attributes_conns\":[\"*internal\"],\"cdrs_conns\":[\"*localhost\"],\"channel_sync_interval\":\"0\",\"chargers_conns\":[\"*internal\"],\"client_protocol\":1,\"debit_interval\":\"0\",\"default_usage\":{\"*any\":\"3h0m0s\",\"*data\":\"1048576\",\"*sms\":\"1\",\"*voice\":\"3h0m0s\"},\"enabled\":true,\"listen_bigob\":\"\",\"listen_bijson\":\"127.0.0.1:2014\",\"min_dur_low_balance\":\"0\",\"rals_conns\":[\"*internal\"],\"replication_conns\":[],\"resources_conns\":[\"*internal\"],\"routes_conns\":[\"*internal\"],\"scheduler_conns\":[],\"session_indexes\":[\"OriginID\"],\"session_ttl\":\"0\",\"stats_conns\":[],\"stir\":{\"allowed_attest\":[\"*any\"],\"default_attest\":\"A\",\"payload_maxduration\":\"-1\",\"privatekey_path\":\"\",\"publickey_path\":\"\"},\"store_session_costs\":false,\"terminate_attempts\":5,\"thresholds_conns\":[]}}"
	var rpl13 string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.SessionSJSON},
	}, &rpl13); err != nil {
		t.Error(err)
	} else if cfgStr != rpl13 {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl13))
	}
	cfgStr = "{\"asterisk_agent\":{\"asterisk_conns\":[{\"address\":\"127.0.0.1:8088\",\"alias\":\"\",\"connect_attempts\":3,\"password\":\"CGRateS.org\",\"reconnects\":5,\"user\":\"cgrates\"}],\"create_cdr\":false,\"enabled\":false,\"sessions_conns\":[\"*birpc_internal\"]}}"
	var rpl14 string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.AsteriskAgentJSON},
	}, &rpl14); err != nil {
		t.Error(err)
	} else if cfgStr != rpl14 {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl14))
	}
	cfgStr = "{\"freeswitch_agent\":{\"create_cdr\":false,\"empty_balance_ann_file\":\"\",\"empty_balance_context\":\"\",\"enabled\":false,\"event_socket_conns\":[{\"address\":\"127.0.0.1:8021\",\"alias\":\"127.0.0.1:8021\",\"password\":\"ClueCon\",\"reconnects\":5}],\"extra_fields\":\"\",\"low_balance_ann_file\":\"\",\"max_wait_connection\":\"2s\",\"sessions_conns\":[\"*birpc_internal\"],\"subscribe_park\":true}}"
	var rpl15 string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.FreeSWITCHAgentJSON},
	}, &rpl15); err != nil {
		t.Error(err)
	} else if cfgStr != rpl15 {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl15))
	}
	cfgStr = "{\"kamailio_agent\":{\"create_cdr\":false,\"enabled\":false,\"evapi_conns\":[{\"address\":\"127.0.0.1:8448\",\"alias\":\"\",\"reconnects\":5}],\"sessions_conns\":[\"*birpc_internal\"],\"timezone\":\"\"}}"
	var rpl16 string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.KamailioAgentJSON},
	}, &rpl16); err != nil {
		t.Error(err)
	} else if cfgStr != rpl16 {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl16))
	}
	cfgStr = "{\"diameter_agent\":{\"asr_template\":\"\",\"concurrent_requests\":-1,\"dictionaries_path\":\"/usr/share/cgrates/diameter/dict/\",\"enabled\":false,\"forced_disconnect\":\"*none\",\"listen\":\"127.0.0.1:3868\",\"listen_net\":\"tcp\",\"origin_host\":\"CGR-DA\",\"origin_realm\":\"cgrates.org\",\"product_name\":\"CGRateS\",\"rar_template\":\"\",\"request_processors\":[],\"sessions_conns\":[\"*birpc_internal\"],\"synced_conn_requests\":false,\"vendor_id\":0}}"
	var rpl17 string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.DiameterAgentJSON},
	}, &rpl17); err != nil {
		t.Error(err)
	} else if cfgStr != rpl17 {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl17))
	}

	cfgStr = "{\"http_agent\":[]}"
	var rpl18 string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.HTTPAgentJSON},
	}, &rpl18); err != nil {
		t.Error(err)
	} else if cfgStr != rpl18 {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl18))
	}

	cfgStr = "{\"dns_agent\":{\"enabled\":false,\"listen\":\"127.0.0.1:2053\",\"listen_net\":\"udp\",\"request_processors\":[],\"sessions_conns\":[\"*internal\"],\"timezone\":\"\"}}"
	var rpl19 string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.DNSAgentJSON},
	}, &rpl19); err != nil {
		t.Error(err)
	} else if cfgStr != rpl19 {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl19))
	}

	cfgStr = "{\"attributes\":{\"any_context\":true,\"apiers_conns\":[\"*localhost\"],\"enabled\":true,\"indexed_selects\":true,\"nested_fields\":false,\"prefix_indexed_fields\":[],\"process_runs\":1,\"resources_conns\":[\"*localhost\"],\"stats_conns\":[\"*localhost\"],\"suffix_indexed_fields\":[]}}"
	var rpl20 string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.AttributeSJSON},
	}, &rpl20); err != nil {
		t.Error(err)
	} else if cfgStr != rpl20 {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl20))
	}

	cfgStr = "{\"chargers\":{\"attributes_conns\":[\"*internal\"],\"enabled\":true,\"indexed_selects\":true,\"nested_fields\":false,\"prefix_indexed_fields\":[],\"suffix_indexed_fields\":[]}}"
	var rpl21 string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.ChargerSJSON},
	}, &rpl21); err != nil {
		t.Error(err)
	} else if cfgStr != rpl21 {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl21))
	}
	cfgStr = "{\"resources\":{\"enabled\":true,\"indexed_selects\":true,\"nested_fields\":false,\"prefix_indexed_fields\":[],\"store_interval\":\"1s\",\"suffix_indexed_fields\":[],\"thresholds_conns\":[\"*internal\"]}}"
	var rpl22 string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.ResourceSJSON},
	}, &rpl22); err != nil {
		t.Error(err)
	} else if cfgStr != rpl22 {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl22))
	}
	cfgStr = "{\"stats\":{\"enabled\":true,\"indexed_selects\":true,\"nested_fields\":false,\"prefix_indexed_fields\":[],\"store_interval\":\"1s\",\"store_uncompressed_limit\":0,\"suffix_indexed_fields\":[],\"thresholds_conns\":[\"*internal\"]}}"
	var rpl23 string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.StatSJSON},
	}, &rpl23); err != nil {
		t.Error(err)
	} else if cfgStr != rpl23 {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl23))
	}

	cfgStr = "{\"thresholds\":{\"enabled\":true,\"indexed_selects\":true,\"nested_fields\":false,\"prefix_indexed_fields\":[],\"store_interval\":\"1s\",\"suffix_indexed_fields\":[]}}"
	var rpl24 string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.ThresholdSJSON},
	}, &rpl24); err != nil {
		t.Error(err)
	} else if cfgStr != rpl24 {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl24))
	}

	cfgStr = "{\"routes\":{\"attributes_conns\":[],\"default_ratio\":1,\"enabled\":true,\"indexed_selects\":true,\"nested_fields\":false,\"prefix_indexed_fields\":[\"*req.Destination\"],\"rals_conns\":[\"*internal\"],\"resources_conns\":[\"*internal\"],\"stats_conns\":[\"*internal\"],\"suffix_indexed_fields\":[]}}"
	var rpl25 string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.RouteSJSON},
	}, &rpl25); err != nil {
		t.Error(err)
	} else if cfgStr != rpl25 {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl25))
	}

	cfgStr = "{\"loaders\":[{\"caches_conns\":[\"*internal\"],\"data\":[{\"fields\":[{\"mandatory\":true,\"path\":\"Tenant\",\"tag\":\"TenantID\",\"type\":\"*variable\",\"value\":\"~*req.0\"},{\"mandatory\":true,\"path\":\"ID\",\"tag\":\"ProfileID\",\"type\":\"*variable\",\"value\":\"~*req.1\"},{\"path\":\"Contexts\",\"tag\":\"Contexts\",\"type\":\"*variable\",\"value\":\"~*req.2\"},{\"path\":\"FilterIDs\",\"tag\":\"FilterIDs\",\"type\":\"*variable\",\"value\":\"~*req.3\"},{\"path\":\"ActivationInterval\",\"tag\":\"ActivationInterval\",\"type\":\"*variable\",\"value\":\"~*req.4\"},{\"path\":\"AttributeFilterIDs\",\"tag\":\"AttributeFilterIDs\",\"type\":\"*variable\",\"value\":\"~*req.5\"},{\"path\":\"Path\",\"tag\":\"Path\",\"type\":\"*variable\",\"value\":\"~*req.6\"},{\"path\":\"Type\",\"tag\":\"Type\",\"type\":\"*variable\",\"value\":\"~*req.7\"},{\"path\":\"Value\",\"tag\":\"Value\",\"type\":\"*variable\",\"value\":\"~*req.8\"},{\"path\":\"Blocker\",\"tag\":\"Blocker\",\"type\":\"*variable\",\"value\":\"~*req.9\"},{\"path\":\"Weight\",\"tag\":\"Weight\",\"type\":\"*variable\",\"value\":\"~*req.10\"}],\"file_name\":\"Attributes.csv\",\"flags\":null,\"type\":\"*attributes\"},{\"fields\":[{\"mandatory\":true,\"path\":\"Tenant\",\"tag\":\"Tenant\",\"type\":\"*variable\",\"value\":\"~*req.0\"},{\"mandatory\":true,\"path\":\"ID\",\"tag\":\"ID\",\"type\":\"*variable\",\"value\":\"~*req.1\"},{\"path\":\"Type\",\"tag\":\"Type\",\"type\":\"*variable\",\"value\":\"~*req.2\"},{\"path\":\"Element\",\"tag\":\"Element\",\"type\":\"*variable\",\"value\":\"~*req.3\"},{\"path\":\"Values\",\"tag\":\"Values\",\"type\":\"*variable\",\"value\":\"~*req.4\"},{\"path\":\"ActivationInterval\",\"tag\":\"ActivationInterval\",\"type\":\"*variable\",\"value\":\"~*req.5\"}],\"file_name\":\"Filters.csv\",\"flags\":null,\"type\":\"*filters\"},{\"fields\":[{\"mandatory\":true,\"path\":\"Tenant\",\"tag\":\"Tenant\",\"type\":\"*variable\",\"value\":\"~*req.0\"},{\"mandatory\":true,\"path\":\"ID\",\"tag\":\"ID\",\"type\":\"*variable\",\"value\":\"~*req.1\"},{\"path\":\"FilterIDs\",\"tag\":\"FilterIDs\",\"type\":\"*variable\",\"value\":\"~*req.2\"},{\"path\":\"ActivationInterval\",\"tag\":\"ActivationInterval\",\"type\":\"*variable\",\"value\":\"~*req.3\"},{\"path\":\"UsageTTL\",\"tag\":\"TTL\",\"type\":\"*variable\",\"value\":\"~*req.4\"},{\"path\":\"Limit\",\"tag\":\"Limit\",\"type\":\"*variable\",\"value\":\"~*req.5\"},{\"path\":\"AllocationMessage\",\"tag\":\"AllocationMessage\",\"type\":\"*variable\",\"value\":\"~*req.6\"},{\"path\":\"Blocker\",\"tag\":\"Blocker\",\"type\":\"*variable\",\"value\":\"~*req.7\"},{\"path\":\"Stored\",\"tag\":\"Stored\",\"type\":\"*variable\",\"value\":\"~*req.8\"},{\"path\":\"Weight\",\"tag\":\"Weight\",\"type\":\"*variable\",\"value\":\"~*req.9\"},{\"path\":\"ThresholdIDs\",\"tag\":\"ThresholdIDs\",\"type\":\"*variable\",\"value\":\"~*req.10\"}],\"file_name\":\"Resources.csv\",\"flags\":null,\"type\":\"*resources\"},{\"fields\":[{\"mandatory\":true,\"path\":\"Tenant\",\"tag\":\"Tenant\",\"type\":\"*variable\",\"value\":\"~*req.0\"},{\"mandatory\":true,\"path\":\"ID\",\"tag\":\"ID\",\"type\":\"*variable\",\"value\":\"~*req.1\"},{\"path\":\"FilterIDs\",\"tag\":\"FilterIDs\",\"type\":\"*variable\",\"value\":\"~*req.2\"},{\"path\":\"ActivationInterval\",\"tag\":\"ActivationInterval\",\"type\":\"*variable\",\"value\":\"~*req.3\"},{\"path\":\"QueueLength\",\"tag\":\"QueueLength\",\"type\":\"*variable\",\"value\":\"~*req.4\"},{\"path\":\"TTL\",\"tag\":\"TTL\",\"type\":\"*variable\",\"value\":\"~*req.5\"},{\"path\":\"MinItems\",\"tag\":\"MinItems\",\"type\":\"*variable\",\"value\":\"~*req.6\"},{\"path\":\"MetricIDs\",\"tag\":\"MetricIDs\",\"type\":\"*variable\",\"value\":\"~*req.7\"},{\"path\":\"MetricFilterIDs\",\"tag\":\"MetricFilterIDs\",\"type\":\"*variable\",\"value\":\"~*req.8\"},{\"path\":\"Blocker\",\"tag\":\"Blocker\",\"type\":\"*variable\",\"value\":\"~*req.9\"},{\"path\":\"Stored\",\"tag\":\"Stored\",\"type\":\"*variable\",\"value\":\"~*req.10\"},{\"path\":\"Weight\",\"tag\":\"Weight\",\"type\":\"*variable\",\"value\":\"~*req.11\"},{\"path\":\"ThresholdIDs\",\"tag\":\"ThresholdIDs\",\"type\":\"*variable\",\"value\":\"~*req.12\"}],\"file_name\":\"Stats.csv\",\"flags\":null,\"type\":\"*stats\"},{\"fields\":[{\"mandatory\":true,\"path\":\"Tenant\",\"tag\":\"Tenant\",\"type\":\"*variable\",\"value\":\"~*req.0\"},{\"mandatory\":true,\"path\":\"ID\",\"tag\":\"ID\",\"type\":\"*variable\",\"value\":\"~*req.1\"},{\"path\":\"FilterIDs\",\"tag\":\"FilterIDs\",\"type\":\"*variable\",\"value\":\"~*req.2\"},{\"path\":\"ActivationInterval\",\"tag\":\"ActivationInterval\",\"type\":\"*variable\",\"value\":\"~*req.3\"},{\"path\":\"MaxHits\",\"tag\":\"MaxHits\",\"type\":\"*variable\",\"value\":\"~*req.4\"},{\"path\":\"MinHits\",\"tag\":\"MinHits\",\"type\":\"*variable\",\"value\":\"~*req.5\"},{\"path\":\"MinSleep\",\"tag\":\"MinSleep\",\"type\":\"*variable\",\"value\":\"~*req.6\"},{\"path\":\"Blocker\",\"tag\":\"Blocker\",\"type\":\"*variable\",\"value\":\"~*req.7\"},{\"path\":\"Weight\",\"tag\":\"Weight\",\"type\":\"*variable\",\"value\":\"~*req.8\"},{\"path\":\"ActionIDs\",\"tag\":\"ActionIDs\",\"type\":\"*variable\",\"value\":\"~*req.9\"},{\"path\":\"Async\",\"tag\":\"Async\",\"type\":\"*variable\",\"value\":\"~*req.10\"}],\"file_name\":\"Thresholds.csv\",\"flags\":null,\"type\":\"*thresholds\"},{\"fields\":[{\"mandatory\":true,\"path\":\"Tenant\",\"tag\":\"Tenant\",\"type\":\"*variable\",\"value\":\"~*req.0\"},{\"mandatory\":true,\"path\":\"ID\",\"tag\":\"ID\",\"type\":\"*variable\",\"value\":\"~*req.1\"},{\"path\":\"FilterIDs\",\"tag\":\"FilterIDs\",\"type\":\"*variable\",\"value\":\"~*req.2\"},{\"path\":\"ActivationInterval\",\"tag\":\"ActivationInterval\",\"type\":\"*variable\",\"value\":\"~*req.3\"},{\"path\":\"Sorting\",\"tag\":\"Sorting\",\"type\":\"*variable\",\"value\":\"~*req.4\"},{\"path\":\"SortingParameters\",\"tag\":\"SortingParameters\",\"type\":\"*variable\",\"value\":\"~*req.5\"},{\"path\":\"RouteID\",\"tag\":\"RouteID\",\"type\":\"*variable\",\"value\":\"~*req.6\"},{\"path\":\"RouteFilterIDs\",\"tag\":\"RouteFilterIDs\",\"type\":\"*variable\",\"value\":\"~*req.7\"},{\"path\":\"RouteAccountIDs\",\"tag\":\"RouteAccountIDs\",\"type\":\"*variable\",\"value\":\"~*req.8\"},{\"path\":\"RouteRatingPlanIDs\",\"tag\":\"RouteRatingPlanIDs\",\"type\":\"*variable\",\"value\":\"~*req.9\"},{\"path\":\"RouteResourceIDs\",\"tag\":\"RouteResourceIDs\",\"type\":\"*variable\",\"value\":\"~*req.10\"},{\"path\":\"RouteStatIDs\",\"tag\":\"RouteStatIDs\",\"type\":\"*variable\",\"value\":\"~*req.11\"},{\"path\":\"RouteWeight\",\"tag\":\"RouteWeight\",\"type\":\"*variable\",\"value\":\"~*req.12\"},{\"path\":\"RouteBlocker\",\"tag\":\"RouteBlocker\",\"type\":\"*variable\",\"value\":\"~*req.13\"},{\"path\":\"RouteParameters\",\"tag\":\"RouteParameters\",\"type\":\"*variable\",\"value\":\"~*req.14\"},{\"path\":\"Weight\",\"tag\":\"Weight\",\"type\":\"*variable\",\"value\":\"~*req.15\"}],\"file_name\":\"Routes.csv\",\"flags\":null,\"type\":\"*routes\"},{\"fields\":[{\"mandatory\":true,\"path\":\"Tenant\",\"tag\":\"Tenant\",\"type\":\"*variable\",\"value\":\"~*req.0\"},{\"mandatory\":true,\"path\":\"ID\",\"tag\":\"ID\",\"type\":\"*variable\",\"value\":\"~*req.1\"},{\"path\":\"FilterIDs\",\"tag\":\"FilterIDs\",\"type\":\"*variable\",\"value\":\"~*req.2\"},{\"path\":\"ActivationInterval\",\"tag\":\"ActivationInterval\",\"type\":\"*variable\",\"value\":\"~*req.3\"},{\"path\":\"RunID\",\"tag\":\"RunID\",\"type\":\"*variable\",\"value\":\"~*req.4\"},{\"path\":\"AttributeIDs\",\"tag\":\"AttributeIDs\",\"type\":\"*variable\",\"value\":\"~*req.5\"},{\"path\":\"Weight\",\"tag\":\"Weight\",\"type\":\"*variable\",\"value\":\"~*req.6\"}],\"file_name\":\"Chargers.csv\",\"flags\":null,\"type\":\"*chargers\"},{\"fields\":[{\"mandatory\":true,\"path\":\"Tenant\",\"tag\":\"Tenant\",\"type\":\"*variable\",\"value\":\"~*req.0\"},{\"mandatory\":true,\"path\":\"ID\",\"tag\":\"ID\",\"type\":\"*variable\",\"value\":\"~*req.1\"},{\"path\":\"Contexts\",\"tag\":\"Contexts\",\"type\":\"*variable\",\"value\":\"~*req.2\"},{\"path\":\"FilterIDs\",\"tag\":\"FilterIDs\",\"type\":\"*variable\",\"value\":\"~*req.3\"},{\"path\":\"ActivationInterval\",\"tag\":\"ActivationInterval\",\"type\":\"*variable\",\"value\":\"~*req.4\"},{\"path\":\"Strategy\",\"tag\":\"Strategy\",\"type\":\"*variable\",\"value\":\"~*req.5\"},{\"path\":\"StrategyParameters\",\"tag\":\"StrategyParameters\",\"type\":\"*variable\",\"value\":\"~*req.6\"},{\"path\":\"ConnID\",\"tag\":\"ConnID\",\"type\":\"*variable\",\"value\":\"~*req.7\"},{\"path\":\"ConnFilterIDs\",\"tag\":\"ConnFilterIDs\",\"type\":\"*variable\",\"value\":\"~*req.8\"},{\"path\":\"ConnWeight\",\"tag\":\"ConnWeight\",\"type\":\"*variable\",\"value\":\"~*req.9\"},{\"path\":\"ConnBlocker\",\"tag\":\"ConnBlocker\",\"type\":\"*variable\",\"value\":\"~*req.10\"},{\"path\":\"ConnParameters\",\"tag\":\"ConnParameters\",\"type\":\"*variable\",\"value\":\"~*req.11\"},{\"path\":\"Weight\",\"tag\":\"Weight\",\"type\":\"*variable\",\"value\":\"~*req.12\"}],\"file_name\":\"DispatcherProfiles.csv\",\"flags\":null,\"type\":\"*dispatchers\"},{\"fields\":[{\"mandatory\":true,\"path\":\"Tenant\",\"tag\":\"Tenant\",\"type\":\"*variable\",\"value\":\"~*req.0\"},{\"mandatory\":true,\"path\":\"ID\",\"tag\":\"ID\",\"type\":\"*variable\",\"value\":\"~*req.1\"},{\"path\":\"Address\",\"tag\":\"Address\",\"type\":\"*variable\",\"value\":\"~*req.2\"},{\"path\":\"Transport\",\"tag\":\"Transport\",\"type\":\"*variable\",\"value\":\"~*req.3\"},{\"path\":\"ConnectAttempts\",\"tag\":\"ConnectAttempts\",\"type\":\"*variable\",\"value\":\"~*req.4\"},{\"path\":\"Reconnects\",\"tag\":\"Reconnects\",\"type\":\"*variable\",\"value\":\"~*req.5\"},{\"path\":\"ConnectTimeout\",\"tag\":\"ConnectTimeout\",\"type\":\"*variable\",\"value\":\"~*req.6\"},{\"path\":\"ReplyTimeout\",\"tag\":\"ReplyTimeout\",\"type\":\"*variable\",\"value\":\"~*req.7\"},{\"path\":\"TLS\",\"tag\":\"TLS\",\"type\":\"*variable\",\"value\":\"~*req.8\"},{\"path\":\"ClientKey\",\"tag\":\"ClientKey\",\"type\":\"*variable\",\"value\":\"~*req.9\"},{\"path\":\"ClientCertificate\",\"tag\":\"ClientCertificate\",\"type\":\"*variable\",\"value\":\"~*req.10\"},{\"path\":\"CaCertificate\",\"tag\":\"CaCertificate\",\"type\":\"*variable\",\"value\":\"~*req.11\"}],\"file_name\":\"DispatcherHosts.csv\",\"flags\":null,\"type\":\"*dispatcher_hosts\"}],\"dry_run\":false,\"enabled\":false,\"field_separator\":\",\",\"id\":\"*default\",\"lock_filename\":\".cgr.lck\",\"run_delay\":\"0\",\"tenant\":\"\",\"tp_in_dir\":\"/var/spool/cgrates/loader/in\",\"tp_out_dir\":\"/var/spool/cgrates/loader/out\"}]}"
	var rpl26 string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.LoaderSJSON},
	}, &rpl26); err != nil {
		t.Error(err)
	} else if cfgStr != rpl26 {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl26))
	}

	cfgStr = "{\"suretax\":{\"bill_to_number\":\"\",\"business_unit\":\"\",\"client_number\":\"\",\"client_tracking\":\"~*req.CGRID\",\"customer_number\":\"~*req.Subject\",\"include_local_cost\":false,\"orig_number\":\"~*req.Subject\",\"p2pplus4\":\"\",\"p2pzipcode\":\"\",\"plus4\":\"\",\"regulatory_code\":\"03\",\"response_group\":\"03\",\"response_type\":\"D4\",\"return_file_code\":\"0\",\"sales_type_code\":\"R\",\"tax_exemption_code_list\":\"\",\"tax_included\":\"0\",\"tax_situs_rule\":\"04\",\"term_number\":\"~*req.Destination\",\"timezone\":\"Local\",\"trans_type_code\":\"010101\",\"unit_type\":\"00\",\"units\":\"1\",\"url\":\"\",\"validation_key\":\"\",\"zipcode\":\"\"}}"
	var rpl28 string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.SureTaxJSON},
	}, &rpl28); err != nil {
		t.Error(err)
	} else if cfgStr != rpl28 {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl28))
	}

	cfgStr = "{\"loader\":{\"caches_conns\":[\"*localhost\"],\"data_path\":\"./\",\"disable_reverse\":false,\"field_separator\":\",\",\"gapi_credentials\":\".gapi/credentials.json\",\"gapi_token\":\".gapi/token.json\",\"scheduler_conns\":[\"*localhost\"],\"tpid\":\"\"}}"
	var rpl29 string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.LoaderJSON},
	}, &rpl29); err != nil {
		t.Error(err)
	} else if cfgStr != rpl29 {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl29))
	}

	cfgStr = "{\"migrator\":{\"out_datadb_encoding\":\"msgpack\",\"out_datadb_host\":\"127.0.0.1\",\"out_datadb_name\":\"10\",\"out_datadb_opts\":{\"redisCACertificate\":\"\",\"redisClientCertificate\":\"\",\"redisClientKey\":\"\",\"redisCluster\":false,\"redisClusterOndownDelay\":\"0\",\"redisClusterSync\":\"5s\",\"redisSentinel\":\"\",\"redisTLS\":false},\"out_datadb_password\":\"\",\"out_datadb_port\":\"6379\",\"out_datadb_type\":\"redis\",\"out_datadb_user\":\"cgrates\",\"out_stordb_host\":\"127.0.0.1\",\"out_stordb_name\":\"cgrates\",\"out_stordb_opts\":{},\"out_stordb_password\":\"CGRateS.org\",\"out_stordb_port\":\"3306\",\"out_stordb_type\":\"mysql\",\"out_stordb_user\":\"cgrates\",\"users_filters\":[\"Account\"]}}"
	var rpl30 string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.MigratorJSON},
	}, &rpl30); err != nil {
		t.Error(err)
	} else if cfgStr != rpl30 {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl30))
	}
	cfgStr = "{\"dispatchers\":{\"any_subsystem\":true,\"attributes_conns\":[],\"enabled\":false,\"indexed_selects\":true,\"nested_fields\":false,\"prefix_indexed_fields\":[],\"suffix_indexed_fields\":[]}}"
	var rpl31 string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.DispatcherSJSON},
	}, &rpl31); err != nil {
		t.Error(err)
	} else if cfgStr != rpl31 {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl31))
	}

	cfgStr = "{\"registrarc\":{\"dispatchers\":{\"hosts\":[],\"refresh_interval\":\"5m0s\",\"registrars_conns\":[]},\"rpc\":{\"hosts\":[],\"refresh_interval\":\"5m0s\",\"registrars_conns\":[]}}}"
	var rpl32 string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.RegistrarCJSON},
	}, &rpl32); err != nil {
		t.Error(err)
	} else if cfgStr != rpl32 {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl32))
	}
	cfgStr = "{\"analyzers\":{\"cleanup_interval\":\"1h0m0s\",\"db_path\":\"/var/spool/cgrates/analyzers\",\"enabled\":false,\"index_type\":\"*scorch\",\"ttl\":\"24h0m0s\"}}"
	var rpl33 string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.AnalyzerSJSON},
	}, &rpl33); err != nil {
		t.Error(err)
	} else if cfgStr != rpl33 {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl33))
	}

	cfgStr = "{\"sip_agent\":{\"enabled\":false,\"listen\":\"127.0.0.1:5060\",\"listen_net\":\"udp\",\"request_processors\":[],\"retransmission_timer\":1000000000,\"sessions_conns\":[\"*internal\"],\"timezone\":\"\"}}"
	var rpl35 string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.SIPAgentJSON},
	}, &rpl35); err != nil {
		t.Error(err)
	} else if cfgStr != rpl35 {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToJSON(cfgStr), utils.ToJSON(rpl35))
	}

	cfgStr = "{\"templates\":{\"*asr\":[{\"mandatory\":true,\"path\":\"*diamreq.Session-Id\",\"tag\":\"SessionId\",\"type\":\"*variable\",\"value\":\"~*req.Session-Id\"},{\"mandatory\":true,\"path\":\"*diamreq.Origin-Host\",\"tag\":\"OriginHost\",\"type\":\"*variable\",\"value\":\"~*req.Destination-Host\"},{\"mandatory\":true,\"path\":\"*diamreq.Origin-Realm\",\"tag\":\"OriginRealm\",\"type\":\"*variable\",\"value\":\"~*req.Destination-Realm\"},{\"mandatory\":true,\"path\":\"*diamreq.Destination-Realm\",\"tag\":\"DestinationRealm\",\"type\":\"*variable\",\"value\":\"~*req.Origin-Realm\"},{\"mandatory\":true,\"path\":\"*diamreq.Destination-Host\",\"tag\":\"DestinationHost\",\"type\":\"*variable\",\"value\":\"~*req.Origin-Host\"},{\"mandatory\":true,\"path\":\"*diamreq.Auth-Application-Id\",\"tag\":\"AuthApplicationId\",\"type\":\"*variable\",\"value\":\"~*vars.*appid\"}],\"*cca\":[{\"mandatory\":true,\"path\":\"*rep.Session-Id\",\"tag\":\"SessionId\",\"type\":\"*variable\",\"value\":\"~*req.Session-Id\"},{\"path\":\"*rep.Result-Code\",\"tag\":\"ResultCode\",\"type\":\"*constant\",\"value\":\"2001\"},{\"mandatory\":true,\"path\":\"*rep.Origin-Host\",\"tag\":\"OriginHost\",\"type\":\"*variable\",\"value\":\"~*vars.OriginHost\"},{\"mandatory\":true,\"path\":\"*rep.Origin-Realm\",\"tag\":\"OriginRealm\",\"type\":\"*variable\",\"value\":\"~*vars.OriginRealm\"},{\"mandatory\":true,\"path\":\"*rep.Auth-Application-Id\",\"tag\":\"AuthApplicationId\",\"type\":\"*variable\",\"value\":\"~*vars.*appid\"},{\"mandatory\":true,\"path\":\"*rep.CC-Request-Type\",\"tag\":\"CCRequestType\",\"type\":\"*variable\",\"value\":\"~*req.CC-Request-Type\"},{\"mandatory\":true,\"path\":\"*rep.CC-Request-Number\",\"tag\":\"CCRequestNumber\",\"type\":\"*variable\",\"value\":\"~*req.CC-Request-Number\"}],\"*cdrLog\":[{\"mandatory\":true,\"path\":\"*cdr.ToR\",\"tag\":\"ToR\",\"type\":\"*variable\",\"value\":\"~*req.BalanceType\"},{\"mandatory\":true,\"path\":\"*cdr.OriginHost\",\"tag\":\"OriginHost\",\"type\":\"*constant\",\"value\":\"127.0.0.1\"},{\"mandatory\":true,\"path\":\"*cdr.RequestType\",\"tag\":\"RequestType\",\"type\":\"*constant\",\"value\":\"*none\"},{\"mandatory\":true,\"path\":\"*cdr.Tenant\",\"tag\":\"Tenant\",\"type\":\"*variable\",\"value\":\"~*req.Tenant\"},{\"mandatory\":true,\"path\":\"*cdr.Account\",\"tag\":\"Account\",\"type\":\"*variable\",\"value\":\"~*req.Account\"},{\"mandatory\":true,\"path\":\"*cdr.Subject\",\"tag\":\"Subject\",\"type\":\"*variable\",\"value\":\"~*req.Account\"},{\"mandatory\":true,\"path\":\"*cdr.Cost\",\"tag\":\"Cost\",\"type\":\"*variable\",\"value\":\"~*req.Cost\"},{\"mandatory\":true,\"path\":\"*cdr.Source\",\"tag\":\"Source\",\"type\":\"*constant\",\"value\":\"*cdrLog\"},{\"mandatory\":true,\"path\":\"*cdr.Usage\",\"tag\":\"Usage\",\"type\":\"*constant\",\"value\":\"1\"},{\"mandatory\":true,\"path\":\"*cdr.RunID\",\"tag\":\"RunID\",\"type\":\"*variable\",\"value\":\"~*req.ActionType\"},{\"mandatory\":true,\"path\":\"*cdr.SetupTime\",\"tag\":\"SetupTime\",\"type\":\"*constant\",\"value\":\"*now\"},{\"mandatory\":true,\"path\":\"*cdr.AnswerTime\",\"tag\":\"AnswerTime\",\"type\":\"*constant\",\"value\":\"*now\"},{\"mandatory\":true,\"path\":\"*cdr.PreRated\",\"tag\":\"PreRated\",\"type\":\"*constant\",\"value\":\"true\"}],\"*err\":[{\"mandatory\":true,\"path\":\"*rep.Session-Id\",\"tag\":\"SessionId\",\"type\":\"*variable\",\"value\":\"~*req.Session-Id\"},{\"mandatory\":true,\"path\":\"*rep.Origin-Host\",\"tag\":\"OriginHost\",\"type\":\"*variable\",\"value\":\"~*vars.OriginHost\"},{\"mandatory\":true,\"path\":\"*rep.Origin-Realm\",\"tag\":\"OriginRealm\",\"type\":\"*variable\",\"value\":\"~*vars.OriginRealm\"}],\"*errSip\":[{\"mandatory\":true,\"path\":\"*rep.Request\",\"tag\":\"Request\",\"type\":\"*constant\",\"value\":\"SIP/2.0 500 Internal Server Error\"}],\"*rar\":[{\"mandatory\":true,\"path\":\"*diamreq.Session-Id\",\"tag\":\"SessionId\",\"type\":\"*variable\",\"value\":\"~*req.Session-Id\"},{\"mandatory\":true,\"path\":\"*diamreq.Origin-Host\",\"tag\":\"OriginHost\",\"type\":\"*variable\",\"value\":\"~*req.Destination-Host\"},{\"mandatory\":true,\"path\":\"*diamreq.Origin-Realm\",\"tag\":\"OriginRealm\",\"type\":\"*variable\",\"value\":\"~*req.Destination-Realm\"},{\"mandatory\":true,\"path\":\"*diamreq.Destination-Realm\",\"tag\":\"DestinationRealm\",\"type\":\"*variable\",\"value\":\"~*req.Origin-Realm\"},{\"mandatory\":true,\"path\":\"*diamreq.Destination-Host\",\"tag\":\"DestinationHost\",\"type\":\"*variable\",\"value\":\"~*req.Origin-Host\"},{\"mandatory\":true,\"path\":\"*diamreq.Auth-Application-Id\",\"tag\":\"AuthApplicationId\",\"type\":\"*variable\",\"value\":\"~*vars.*appid\"},{\"path\":\"*diamreq.Re-Auth-Request-Type\",\"tag\":\"ReAuthRequestType\",\"type\":\"*constant\",\"value\":\"0\"}]}}"
	var rpl36 string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.TemplatesJSON},
	}, &rpl36); err != nil {
		t.Error(err)
	} else if cfgStr != rpl36 {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl36))
	}
	cfgStr = "{\"configs\":{\"enabled\":false,\"root_dir\":\"/var/spool/cgrates/configs\",\"url\":\"/configs/\"}}"
	var rpl37 string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.ConfigSJSON},
	}, &rpl37); err != nil {
		t.Error(err)
	} else if cfgStr != rpl37 {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl37))
	}

	cfgStr = "{\"apiban\":{\"enabled\":false,\"keys\":[]}}"
	var rpl38 string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.APIBanJSON},
	}, &rpl38); err != nil {
		t.Error(err)
	} else if cfgStr != rpl38 {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl38))
	}
}

func testStopCgrEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
