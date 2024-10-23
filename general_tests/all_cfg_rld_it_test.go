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
		testResetDBs,

		testStartEngine,
		testRPCConn,
		testConfigSReload,
		testStopCgrEngine,
	}
)

func TestRldCfg(t *testing.T) {
	switch *utils.DBType {
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
	testCfgPath = path.Join(*utils.DataDir, "conf", "samples", testCfgDir)
	var err error
	if testCfg, err = config.NewCGRConfigFromPath(context.Background(), testCfgPath); err != nil {
		t.Error(err)
	}
}

func testResetDBs(t *testing.T) {
	if err := engine.InitDataDB(testCfg); err != nil {
		t.Fatal(err)
	}
	if err := engine.InitStorDB(testCfg); err != nil {
		t.Fatal(err)
	}
}

func testStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(testCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

func testRPCConn(t *testing.T) {
	testRPC = engine.NewRPCClient(t, testCfg.ListenCfg(), *utils.Encoding)
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

	cfgStr := `{"cores":{"caps":0,"caps_stats_interval":"0","caps_strategy":"*busy","ees_conns":[],"shutdown_timeout":"1s"}}`
	var rpl2 string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.CoreSJSON},
	}, &rpl2); err != nil {
		t.Error(err)
	} else if cfgStr != rpl2 {
		t.Errorf("\nExpected %+v ,\n received: %+v", cfgStr, rpl2)
	}

	cfgStr = `{"rpc_conns":{"*bijson_localhost":{"conns":[{"address":"127.0.0.1:2014","transport":"*birpc_json"}],"poolSize":0,"strategy":"*first"},"*birpc_internal":{"conns":[{"address":"*birpc_internal","transport":""}],"poolSize":0,"strategy":"*first"},"*internal":{"conns":[{"address":"*internal","transport":""}],"poolSize":0,"strategy":"*first"},"*localhost":{"conns":[{"address":"127.0.0.1:2012","transport":"*json"}],"poolSize":0,"strategy":"*first"}}}`
	var rpl string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.RPCConnsJSON},
	}, &rpl); err != nil {
		t.Error(err)
	} else if cfgStr != rpl {
		t.Errorf("\nExpected %+v ,\n received: %+v", cfgStr, rpl)
	}

	cfgStr = `{"listen":{"http":":2080","http_tls":"127.0.0.1:2280","rpc_gob":":2013","rpc_gob_tls":"127.0.0.1:2023","rpc_json":":2012","rpc_json_tls":"127.0.0.1:2022"}}`
	var rpl3 string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.ListenJSON},
	}, &rpl3); err != nil {
		t.Error(err)
	} else if cfgStr != rpl3 {
		t.Errorf("\nExpected %+v ,\n received: %+v", cfgStr, rpl3)
	}

	cfgStr = `{"tls":{"ca_certificate":"","client_certificate":"","client_key":"","server_certificate":"","server_key":"","server_name":"","server_policy":4}}`
	var rpl4 string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.TlsJSON},
	}, &rpl4); err != nil {
		t.Error(err)
	} else if cfgStr != rpl4 {
		t.Errorf("\nExpected %+v ,\n received: %+v", cfgStr, rpl4)
	}

	cfgStr = `{"http":{"auth_users":{},"client_opts":{"dialFallbackDelay":"300ms","dialKeepAlive":"30s","dialTimeout":"30s","disableCompression":false,"disableKeepAlives":false,"expectContinueTimeout":"0s","forceAttemptHttp2":true,"idleConnTimeout":"1m30s","maxConnsPerHost":0,"maxIdleConns":100,"maxIdleConnsPerHost":2,"responseHeaderTimeout":"0s","skipTLSVerification":false,"tlsHandshakeTimeout":"10s"},"freeswitch_cdrs_url":"/freeswitch_json","http_cdrs":"/cdr_http","json_rpc_url":"/jsonrpc","prometheus_url":"/prometheus","registrars_url":"/registrar","use_basic_auth":false,"ws_url":"/ws"}}`

	var rpl5 string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.HTTPJSON},
	}, &rpl5); err != nil {
		t.Error(err)
	} else if cfgStr != rpl5 {
		t.Errorf("\nExpected %+v ,\n received: %+v", cfgStr, rpl5)
	}
	if testCfgDir == "tutmysql" || testCfgDir == "tutmongo" {
		cfgStr = `{"caches":{"partitions":{"*account_filter_indexes":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*accounts":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*action_profile_filter_indexes":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*action_profiles":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*apiban":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"2m0s"},"*attribute_filter_indexes":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*attribute_profiles":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*caps_events":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*cdr_ids":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"10m0s"},"*charger_filter_indexes":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*charger_profiles":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*closed_sessions":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"10s"},"*diameter_messages":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"3h0m0s"},"*dispatcher_filter_indexes":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*dispatcher_hosts":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*dispatcher_loads":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*dispatcher_profiles":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*dispatcher_routes":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*dispatchers":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*event_charges":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"10s"},"*event_resources":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*filters":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*load_ids":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*ranking_profiles":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*rate_filter_indexes":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*rate_profile_filter_indexes":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*rate_profiles":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*replication_hosts":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*resource_filter_indexes":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*resource_profiles":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*resources":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*reverse_filter_indexes":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*route_filter_indexes":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*route_profiles":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*rpc_connections":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*rpc_responses":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"2s"},"*sentrypeer":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":true,"ttl":"24h0m0s"},"*stat_filter_indexes":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*statqueue_profiles":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*statqueues":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*stir":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"3h0m0s"},"*threshold_filter_indexes":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*threshold_profiles":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*thresholds":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*trend_profiles":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*trends":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*uch":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"3h0m0s"}},"remote_conns":[],"replication_conns":[]}}`
		var rpl7 string
		if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
			Tenant:   "cgrates.org",
			Sections: []string{config.CacheJSON},
		}, &rpl7); err != nil {
			t.Error(err)
		} else if cfgStr != rpl7 {
			t.Errorf("\nExpected %+v ,\n received: %+v", cfgStr, rpl7)
		}
	} else if testCfgDir == "tutinternal" {
		cfgStr := `{"caches":{"partitions":{"*account_filter_indexes":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*accounts":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*action_profile_filter_indexes":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*action_profiles":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*apiban":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"2m0s"},"*attribute_filter_indexes":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*attribute_profiles":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*caps_events":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*cdr_ids":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"10m0s"},"*charger_filter_indexes":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*charger_profiles":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*closed_sessions":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"10s"},"*diameter_messages":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"3h0m0s"},"*dispatcher_filter_indexes":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*dispatcher_hosts":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*dispatcher_loads":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*dispatcher_profiles":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*dispatcher_routes":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*dispatchers":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*event_charges":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"10s"},"*event_resources":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*filters":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*load_ids":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*ranking_profiles":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*rate_filter_indexes":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*rate_profile_filter_indexes":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*rate_profiles":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*replication_hosts":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*resource_filter_indexes":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*resource_profiles":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*resources":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*reverse_filter_indexes":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*route_filter_indexes":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*route_profiles":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*rpc_connections":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*rpc_responses":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"2s"},"*sentrypeer":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":true,"ttl":"24h0m0s"},"*stat_filter_indexes":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*statqueue_profiles":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*statqueues":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*stir":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"3h0m0s"},"*threshold_filter_indexes":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*threshold_profiles":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*thresholds":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*trend_profiles":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*trends":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*uch":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"3h0m0s"}},"remote_conns":[],"replication_conns":[]}}`
		var rpl7 string
		if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
			Tenant:   "cgrates.org",
			Sections: []string{config.CacheJSON},
		}, &rpl7); err != nil {
			t.Error(err)
		} else if cfgStr != rpl7 {
			t.Errorf("\nExpected %+v ,\n received: %+v", cfgStr, rpl7)
		}

	}

	cfgStr = `{"filters":{"accounts_conns":["*internal"],"resources_conns":["*internal"],"stats_conns":["*internal"]}}`
	var rpl8 string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.FilterSJSON},
	}, &rpl8); err != nil {
		t.Error(err)
	} else if cfgStr != rpl8 {
		t.Errorf("\nExpected %+v ,\n received: %+v", cfgStr, rpl8)
	}

	cfgStr = `{"cdrs":{"accounts_conns":[],"actions_conns":[],"attributes_conns":[],"chargers_conns":["*internal"],"ees_conns":[],"enabled":true,"extra_fields":[],"online_cdr_exports":null,"opts":{"*accounts":[],"*attributes":[],"*chargers":[],"*ees":[],"*rates":[],"*refund":[],"*rerate":[],"*stats":[],"*store":[],"*thresholds":[]},"rates_conns":[],"session_cost_retries":5,"stats_conns":[],"thresholds_conns":[]}}`
	var rpl10 string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.CDRsJSON},
	}, &rpl10); err != nil {
		t.Error(err)
	} else if cfgStr != rpl10 {
		t.Errorf("\nExpected %+v ,\n received: %+v", cfgStr, rpl10)
	}
	cfgStr = `{"ers":{"enabled":false,"partial_cache_ttl":"1s","readers":[{"cache_dump_fields":[],"concurrent_requests":1024,"fields":[{"mandatory":true,"path":"*cgreq.ToR","tag":"ToR","type":"*variable","value":"~*req.2"},{"mandatory":true,"path":"*cgreq.OriginID","tag":"OriginID","type":"*variable","value":"~*req.3"},{"mandatory":true,"path":"*cgreq.RequestType","tag":"RequestType","type":"*variable","value":"~*req.4"},{"mandatory":true,"path":"*cgreq.Tenant","tag":"Tenant","type":"*variable","value":"~*req.6"},{"mandatory":true,"path":"*cgreq.Category","tag":"Category","type":"*variable","value":"~*req.7"},{"mandatory":true,"path":"*cgreq.Account","tag":"Account","type":"*variable","value":"~*req.8"},{"mandatory":true,"path":"*cgreq.Subject","tag":"Subject","type":"*variable","value":"~*req.9"},{"mandatory":true,"path":"*cgreq.Destination","tag":"Destination","type":"*variable","value":"~*req.10"},{"mandatory":true,"path":"*cgreq.SetupTime","tag":"SetupTime","type":"*variable","value":"~*req.11"},{"mandatory":true,"path":"*cgreq.AnswerTime","tag":"AnswerTime","type":"*variable","value":"~*req.12"},{"mandatory":true,"path":"*cgreq.Usage","tag":"Usage","type":"*variable","value":"~*req.13"}],"filters":[],"flags":[],"id":"*default","max_reconnect_interval":"5m0s","opts":{"csvFieldSeparator":",","csvHeaderDefineChar":":","csvRowLength":0,"natsSubject":"cgrates_cdrs","partialCacheAction":"*none","partialOrderField":"~*req.AnswerTime"},"partial_commit_fields":[],"processed_path":"/var/spool/cgrates/ers/out","reconnects":-1,"run_delay":"0","source_path":"/var/spool/cgrates/ers/in","tenant":"","timezone":"","type":"*none"}],"sessions_conns":["*internal"]}}`
	var rpl11 string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.ERsJSON},
	}, &rpl11); err != nil {
		t.Error(err)
	} else if cfgStr != rpl11 {
		t.Errorf("\nExpected %+v ,\n received: %+v", cfgStr, rpl11)
	}
	cfgStr = `{"ees":{"attributes_conns":[],"cache":{"*fileCSV":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"5s"}},"enabled":false,"exporters":[{"attempts":1,"attribute_context":"","attribute_ids":[],"blocker":false,"concurrent_requests":0,"efs_conns":["*internal"],"export_path":"/var/spool/cgrates/ees","failed_posts_dir":"/var/spool/cgrates/failed_posts","fields":[],"filters":[],"flags":[],"id":"*default","opts":{},"synchronous":false,"timezone":"","type":"*none"}]}}`
	var rpl12 string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.EEsJSON},
	}, &rpl12); err != nil {
		t.Error(err)
	} else if cfgStr != rpl12 {
		t.Errorf("\nExpected %+v ,\n received: %+v", cfgStr, rpl12)
	}

	cfgStr = `{"sessions":{"accounts_conns":[],"actions_conns":[],"alterable_fields":[],"attributes_conns":["*internal"],"cdrs_conns":["*internal"],"channel_sync_interval":"0","chargers_conns":["*internal"],"client_protocol":1,"default_usage":{"*any":"3h0m0s","*data":"1048576","*sms":"1","*voice":"3h0m0s"},"enabled":true,"listen_bigob":"","listen_bijson":"127.0.0.1:2014","min_dur_low_balance":"0","opts":{"*accounts":[],"*attributes":[],"*attributesDerivedReply":[],"*blockerError":[],"*cdrs":[],"*cdrsDerivedReply":[],"*chargeable":[],"*chargers":[],"*debitInterval":[],"*forceDuration":[],"*initiate":[],"*maxUsage":[],"*message":[],"*resources":[],"*resourcesAllocate":[],"*resourcesAuthorize":[],"*resourcesDerivedReply":[],"*resourcesRelease":[],"*routes":[],"*routesDerivedReply":[],"*stats":[],"*statsDerivedReply":[],"*terminate":[],"*thresholds":[],"*thresholdsDerivedReply":[],"*ttl":[],"*ttlLastUsage":[],"*ttlLastUsed":[],"*ttlMaxDelay":[],"*ttlUsage":[],"*update":[]},"rates_conns":["*internal"],"replication_conns":[],"resources_conns":["*internal"],"routes_conns":["*internal"],"session_indexes":["OriginID"],"stats_conns":[],"stir":{"allowed_attest":["*any"],"default_attest":"A","payload_maxduration":"-1","privatekey_path":"","publickey_path":""},"store_session_costs":false,"terminate_attempts":5,"thresholds_conns":[]}}`
	var rpl13 string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.SessionSJSON},
	}, &rpl13); err != nil {
		t.Error(err)
	} else if cfgStr != rpl13 {
		t.Errorf("\nExpected %+v ,\n received: %+v", cfgStr, rpl13)
	}
	cfgStr = `{"asterisk_agent":{"asterisk_conns":[{"address":"127.0.0.1:8088","alias":"","connect_attempts":3,"max_reconnect_interval":"0s","password":"CGRateS.org","reconnects":5,"user":"cgrates"}],"create_cdr":false,"enabled":false,"sessions_conns":["*birpc_internal"]}}`
	var rpl14 string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.AsteriskAgentJSON},
	}, &rpl14); err != nil {
		t.Error(err)
	} else if cfgStr != rpl14 {
		t.Errorf("\nExpected %+v ,\n received: %+v", cfgStr, rpl14)
	}
	cfgStr = `{"freeswitch_agent":{"active_session_delimiter":",","create_cdr":false,"empty_balance_ann_file":"","empty_balance_context":"","enabled":false,"event_socket_conns":[{"address":"127.0.0.1:8021","alias":"127.0.0.1:8021","max_reconnect_interval":"0s","password":"ClueCon","reconnects":5,"reply_timeout":"1m0s"}],"extra_fields":[],"low_balance_ann_file":"","max_wait_connection":"2s","sessions_conns":["*birpc_internal"],"subscribe_park":true}}`
	var rpl15 string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.FreeSWITCHAgentJSON},
	}, &rpl15); err != nil {
		t.Error(err)
	} else if cfgStr != rpl15 {
		t.Errorf("\nExpected %+v ,\n received: %+v", cfgStr, rpl15)
	}
	cfgStr = `{"kamailio_agent":{"create_cdr":false,"enabled":false,"evapi_conns":[{"address":"127.0.0.1:8448","alias":"","max_reconnect_interval":"0s","reconnects":5}],"sessions_conns":["*birpc_internal"],"timezone":""}}`
	var rpl16 string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.KamailioAgentJSON},
	}, &rpl16); err != nil {
		t.Error(err)
	} else if cfgStr != rpl16 {
		t.Errorf("\nExpected %+v ,\n received: %+v", cfgStr, rpl16)
	}
	cfgStr = `{"diameter_agent":{"asr_template":"","concurrent_requests":-1,"dictionaries_path":"/usr/share/cgrates/diameter/dict/","enabled":false,"forced_disconnect":"*none","listen":"127.0.0.1:3868","listen_net":"tcp","origin_host":"CGR-DA","origin_realm":"cgrates.org","product_name":"CGRateS","rar_template":"","request_processors":[],"sessions_conns":["*birpc_internal"],"synced_conn_requests":false,"vendor_id":0}}`
	var rpl17 string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.DiameterAgentJSON},
	}, &rpl17); err != nil {
		t.Error(err)
	} else if cfgStr != rpl17 {
		t.Errorf("\nExpected %+v ,\n received: %+v", cfgStr, rpl17)
	}

	cfgStr = `{"http_agent":[]}`
	var rpl18 string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.HTTPAgentJSON},
	}, &rpl18); err != nil {
		t.Error(err)
	} else if cfgStr != rpl18 {
		t.Errorf("\nExpected %+v ,\n received: %+v", cfgStr, rpl18)
	}

	cfgStr = `{"dns_agent":{"enabled":false,"listeners":[{"address":"127.0.0.1:53","network":"udp"}],"request_processors":[],"sessions_conns":["*internal"],"timezone":""}}`
	var rpl19 string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.DNSAgentJSON},
	}, &rpl19); err != nil {
		t.Error(err)
	} else if cfgStr != rpl19 {
		t.Errorf("\nExpected %+v ,\n received: %+v", cfgStr, rpl19)
	}

	cfgStr = `{"attributes":{"accounts_conns":["*localhost"],"enabled":true,"exists_indexed_fields":[],"indexed_selects":true,"nested_fields":false,"notexists_indexed_fields":[],"opts":{"*processRuns":[],"*profileIDs":[],"*profileIgnoreFilters":[],"*profileRuns":[]},"prefix_indexed_fields":[],"resources_conns":["*localhost"],"stats_conns":["*localhost"],"suffix_indexed_fields":[]}}`
	var rpl20 string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.AttributeSJSON},
	}, &rpl20); err != nil {
		t.Error(err)
	} else if cfgStr != rpl20 {
		t.Errorf("\nExpected %+v ,\n received: %+v", cfgStr, rpl20)
	}

	cfgStr = `{"chargers":{"attributes_conns":["*internal"],"enabled":true,"exists_indexed_fields":[],"indexed_selects":true,"nested_fields":false,"notexists_indexed_fields":[],"prefix_indexed_fields":[],"suffix_indexed_fields":[]}}`
	var rpl21 string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.ChargerSJSON},
	}, &rpl21); err != nil {
		t.Error(err)
	} else if cfgStr != rpl21 {
		t.Errorf("\nExpected %+v ,\n received: %+v", cfgStr, rpl21)
	}
	if testCfgDir == "tutmysql" || testCfgDir == "tutmongo" {
		cfgStr = `{"resources":{"enabled":true,"exists_indexed_fields":[],"indexed_selects":true,"nested_fields":false,"notexists_indexed_fields":[],"opts":{"*units":[],"*usageID":[],"*usageTTL":[]},"prefix_indexed_fields":[],"store_interval":"1s","suffix_indexed_fields":[],"thresholds_conns":["*internal"]}}`
		var rpl22 string
		if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
			Tenant:   "cgrates.org",
			Sections: []string{config.ResourceSJSON},
		}, &rpl22); err != nil {
			t.Error(err)
		} else if cfgStr != rpl22 {
			t.Errorf("\nExpected %+v ,\n received: %+v", cfgStr, rpl22)
		}
	} else if testCfgDir == "tutinternal" {
		cfgStr = `{"resources":{"enabled":true,"exists_indexed_fields":[],"indexed_selects":true,"nested_fields":false,"notexists_indexed_fields":[],"opts":{"*units":[],"*usageID":[],"*usageTTL":[]},"prefix_indexed_fields":[],"store_interval":"-1ns","suffix_indexed_fields":[],"thresholds_conns":["*internal"]}}`
		var rpl22 string
		if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
			Tenant:   "cgrates.org",
			Sections: []string{config.ResourceSJSON},
		}, &rpl22); err != nil {
			t.Error(err)
		} else if cfgStr != rpl22 {
			t.Errorf("\nExpected %+v ,\n received: %+v", cfgStr, rpl22)
		}
	}
	if testCfgDir == "tutmysql" || testCfgDir == "tutmongo" {
		cfgStr = `{"stats":{"ees_conns":[],"ees_exporter_ids":null,"enabled":true,"exists_indexed_fields":[],"indexed_selects":true,"nested_fields":false,"notexists_indexed_fields":[],"opts":{"*profileIDs":[],"*profileIgnoreFilters":[],"*prometheusStatIDs":[],"*roundingDecimals":[]},"prefix_indexed_fields":[],"store_interval":"1s","store_uncompressed_limit":0,"suffix_indexed_fields":[],"thresholds_conns":["*internal"]}}`
		var rpl23 string
		if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
			Tenant:   "cgrates.org",
			Sections: []string{config.StatSJSON},
		}, &rpl23); err != nil {
			t.Error(err)
		} else if cfgStr != rpl23 {
			t.Errorf("\nExpected %+v ,\n received: %+v", cfgStr, rpl23)
		}
	} else if testCfgDir == "tutinternal" {
		cfgStr = `{"stats":{"ees_conns":[],"ees_exporter_ids":null,"enabled":true,"exists_indexed_fields":[],"indexed_selects":true,"nested_fields":false,"notexists_indexed_fields":[],"opts":{"*profileIDs":[],"*profileIgnoreFilters":[],"*prometheusStatIDs":[],"*roundingDecimals":[]},"prefix_indexed_fields":[],"store_interval":"-1ns","store_uncompressed_limit":0,"suffix_indexed_fields":[],"thresholds_conns":["*internal"]}}`
		var rpl23 string
		if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
			Tenant:   "cgrates.org",
			Sections: []string{config.StatSJSON},
		}, &rpl23); err != nil {
			t.Error(err)
		} else if cfgStr != rpl23 {
			t.Errorf("\nExpected %+v ,\n received: %+v", cfgStr, rpl23)
		}
	}
	if testCfgDir == "tutmysql" || testCfgDir == "tutmongo" {
		cfgStr = `{"thresholds":{"actions_conns":[],"enabled":true,"exists_indexed_fields":[],"indexed_selects":true,"nested_fields":false,"notexists_indexed_fields":[],"opts":{"*profileIDs":[],"*profileIgnoreFilters":[]},"prefix_indexed_fields":[],"store_interval":"1s","suffix_indexed_fields":[]}}`
		var rpl24 string
		if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
			Tenant:   "cgrates.org",
			Sections: []string{config.ThresholdSJSON},
		}, &rpl24); err != nil {
			t.Error(err)
		} else if cfgStr != rpl24 {
			t.Errorf("\nExpected %+v ,\n received: %+v", cfgStr, rpl24)
		}
	} else if testCfgDir == "tutinternal" {
		cfgStr = `{"thresholds":{"actions_conns":[],"enabled":true,"exists_indexed_fields":[],"indexed_selects":true,"nested_fields":false,"notexists_indexed_fields":[],"opts":{"*profileIDs":[],"*profileIgnoreFilters":[]},"prefix_indexed_fields":[],"store_interval":"-1ns","suffix_indexed_fields":[]}}`
		var rpl24 string
		if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
			Tenant:   "cgrates.org",
			Sections: []string{config.ThresholdSJSON},
		}, &rpl24); err != nil {
			t.Error(err)
		} else if cfgStr != rpl24 {
			t.Errorf("\nExpected %+v ,\n received: %+v", cfgStr, rpl24)
		}
	}
	cfgStr = `{"routes":{"accounts_conns":[],"attributes_conns":[],"default_ratio":1,"enabled":true,"exists_indexed_fields":[],"indexed_selects":true,"nested_fields":false,"notexists_indexed_fields":[],"opts":{"*context":[],"*ignoreErrors":[],"*limit":[],"*maxCost":[],"*maxItems":[],"*offset":[],"*profileCount":[],"*usage":[]},"prefix_indexed_fields":["*req.Destination"],"rates_conns":["*internal"],"resources_conns":["*internal"],"stats_conns":["*internal"],"suffix_indexed_fields":[]}}`
	var rpl25 string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.RouteSJSON},
	}, &rpl25); err != nil {
		t.Error(err)
	} else if cfgStr != rpl25 {
		t.Errorf("\nExpected %+v ,\n received: %+v", cfgStr, rpl25)
	}

	if testCfgDir == "tutinternal" {
		cfgStr := `{"loaders":[{"action":"*store","cache":{"*accounts":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"5s"},"*action_profiles":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"5s"},"*attributes":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"5s"},"*chargers":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"5s"},"*dispatcher_hosts":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"5s"},"*dispatchers":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"5s"},"*filters":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"5s"},"*rate_profiles":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"5s"},"*resources":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"5s"},"*routes":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"5s"},"*stats":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"5s"},"*thresholds":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"5s"}},"caches_conns":["*internal"],"data":[{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"new_branch":true,"path":"Rules.Type","tag":"Type","type":"*variable","value":"~*req.2"},{"path":"Rules.Element","tag":"Element","type":"*variable","value":"~*req.3"},{"path":"Rules.Values","tag":"Values","type":"*variable","value":"~*req.4"}],"file_name":"Filters.csv","flags":null,"type":"*filters"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"TenantID","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ProfileID","type":"*variable","value":"~*req.1"},{"path":"FilterIDs","tag":"FilterIDs","type":"*variable","value":"~*req.2"},{"path":"Weights","tag":"Weights","type":"*variable","value":"~*req.3"},{"path":"Blockers","tag":"Blockers","type":"*variable","value":"~*req.4"},{"new_branch":true,"path":"Attributes.FilterIDs","tag":"AttributeFilterIDs","type":"*variable","value":"~*req.5"},{"path":"Attributes.Blockers","tag":"AttributeBlockers","type":"*variable","value":"~*req.6"},{"path":"Attributes.Path","tag":"Path","type":"*variable","value":"~*req.7"},{"path":"Attributes.Type","tag":"Type","type":"*variable","value":"~*req.8"},{"path":"Attributes.Value","tag":"Value","type":"*variable","value":"~*req.9"}],"file_name":"Attributes.csv","flags":null,"type":"*attributes"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"path":"FilterIDs","tag":"FilterIDs","type":"*variable","value":"~*req.2"},{"path":"Weights","tag":"Weights","type":"*variable","value":"~*req.3"},{"path":"UsageTTL","tag":"TTL","type":"*variable","value":"~*req.4"},{"path":"Limit","tag":"Limit","type":"*variable","value":"~*req.5"},{"path":"AllocationMessage","tag":"AllocationMessage","type":"*variable","value":"~*req.6"},{"path":"Blocker","tag":"Blocker","type":"*variable","value":"~*req.7"},{"path":"Stored","tag":"Stored","type":"*variable","value":"~*req.8"},{"path":"ThresholdIDs","tag":"ThresholdIDs","type":"*variable","value":"~*req.9"}],"file_name":"Resources.csv","flags":null,"type":"*resources"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"path":"FilterIDs","tag":"FilterIDs","type":"*variable","value":"~*req.2"},{"path":"Weights","tag":"Weights","type":"*variable","value":"~*req.3"},{"path":"Blockers","tag":"Blockers","type":"*variable","value":"~*req.4"},{"path":"QueueLength","tag":"QueueLength","type":"*variable","value":"~*req.5"},{"path":"TTL","tag":"TTL","type":"*variable","value":"~*req.6"},{"path":"MinItems","tag":"MinItems","type":"*variable","value":"~*req.7"},{"path":"Stored","tag":"Stored","type":"*variable","value":"~*req.8"},{"path":"ThresholdIDs","tag":"ThresholdIDs","type":"*variable","value":"~*req.9"},{"new_branch":true,"path":"Metrics.MetricID","tag":"MetricIDs","type":"*variable","value":"~*req.10"},{"path":"Metrics.FilterIDs","tag":"MetricFilterIDs","type":"*variable","value":"~*req.11"},{"path":"Metrics.Blockers","tag":"MetricBlockers","type":"*variable","value":"~*req.12"}],"file_name":"Stats.csv","flags":null,"type":"*stats"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"path":"FilterIDs","tag":"FilterIDs","type":"*variable","value":"~*req.2"},{"path":"Weights","tag":"Weights","type":"*variable","value":"~*req.3"},{"path":"MaxHits","tag":"MaxHits","type":"*variable","value":"~*req.4"},{"path":"MinHits","tag":"MinHits","type":"*variable","value":"~*req.5"},{"path":"MinSleep","tag":"MinSleep","type":"*variable","value":"~*req.6"},{"path":"Blocker","tag":"Blocker","type":"*variable","value":"~*req.7"},{"path":"ActionProfileIDs","tag":"ActionProfileIDs","type":"*variable","value":"~*req.8"},{"path":"Async","tag":"Async","type":"*variable","value":"~*req.9"}],"file_name":"Thresholds.csv","flags":null,"type":"*thresholds"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"path":"FilterIDs","tag":"FilterIDs","type":"*variable","value":"~*req.2"},{"path":"Weights","tag":"Weights","type":"*variable","value":"~*req.3"},{"path":"Blockers","tag":"Blockers","type":"*variable","value":"~*req.4"},{"path":"Sorting","tag":"Sorting","type":"*variable","value":"~*req.5"},{"path":"SortingParameters","tag":"SortingParameters","type":"*variable","value":"~*req.6"},{"new_branch":true,"path":"Routes.ID","tag":"RouteID","type":"*variable","value":"~*req.7"},{"path":"Routes.FilterIDs","tag":"RouteFilterIDs","type":"*variable","value":"~*req.8"},{"path":"Routes.AccountIDs","tag":"RouteAccountIDs","type":"*variable","value":"~*req.9"},{"path":"Routes.RateProfileIDs","tag":"RouteRateProfileIDs","type":"*variable","value":"~*req.10"},{"path":"Routes.ResourceIDs","tag":"RouteResourceIDs","type":"*variable","value":"~*req.11"},{"path":"Routes.StatIDs","tag":"RouteStatIDs","type":"*variable","value":"~*req.12"},{"path":"Routes.Weights","tag":"RouteWeights","type":"*variable","value":"~*req.13"},{"path":"Routes.Blockers","tag":"RouteBlockers","type":"*variable","value":"~*req.14"},{"path":"Routes.RouteParameters","tag":"RouteParameters","type":"*variable","value":"~*req.15"}],"file_name":"Routes.csv","flags":null,"type":"*routes"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"path":"FilterIDs","tag":"FilterIDs","type":"*variable","value":"~*req.2"},{"path":"Weights","tag":"Weights","type":"*variable","value":"~*req.3"},{"path":"Blockers","tag":"Blockers","type":"*variable","value":"~*req.4"},{"path":"RunID","tag":"RunID","type":"*variable","value":"~*req.5"},{"path":"AttributeIDs","tag":"AttributeIDs","type":"*variable","value":"~*req.6"}],"file_name":"Chargers.csv","flags":null,"type":"*chargers"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"path":"FilterIDs","tag":"FilterIDs","type":"*variable","value":"~*req.2"},{"path":"Weight","tag":"Weight","type":"*variable","value":"~*req.3"},{"path":"Strategy","tag":"Strategy","type":"*variable","value":"~*req.4"},{"path":"StrategyParams","tag":"StrategyParameters","type":"*variable","value":"~*req.5"},{"new_branch":true,"path":"Hosts.ID","tag":"ConnID","type":"*variable","value":"~*req.6"},{"path":"Hosts.FilterIDs","tag":"ConnFilterIDs","type":"*variable","value":"~*req.7"},{"path":"Hosts.Weight","tag":"ConnWeight","type":"*variable","value":"~*req.8"},{"path":"Hosts.Blocker","tag":"ConnBlocker","type":"*variable","value":"~*req.9"},{"path":"Hosts.Params","tag":"ConnParameters","type":"*variable","value":"~*req.10"}],"file_name":"DispatcherProfiles.csv","flags":null,"type":"*dispatchers"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"path":"Address","tag":"Address","type":"*variable","value":"~*req.2"},{"path":"Transport","tag":"Transport","type":"*variable","value":"~*req.3"},{"path":"ConnectAttempts","tag":"ConnectAttempts","type":"*variable","value":"~*req.4"},{"path":"Reconnects","tag":"Reconnects","type":"*variable","value":"~*req.5"},{"path":"MaxReconnectInterval","tag":"MaxReconnectInterval","type":"*variable","value":"~*req.6"},{"path":"ConnectTimeout","tag":"ConnectTimeout","type":"*variable","value":"~*req.7"},{"path":"ReplyTimeout","tag":"ReplyTimeout","type":"*variable","value":"~*req.8"},{"path":"TLS","tag":"TLS","type":"*variable","value":"~*req.9"},{"path":"ClientKey","tag":"ClientKey","type":"*variable","value":"~*req.10"},{"path":"ClientCertificate","tag":"ClientCertificate","type":"*variable","value":"~*req.11"},{"path":"CaCertificate","tag":"CaCertificate","type":"*variable","value":"~*req.12"}],"file_name":"DispatcherHosts.csv","flags":null,"type":"*dispatcher_hosts"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"path":"FilterIDs","tag":"FilterIDs","type":"*variable","value":"~*req.2"},{"path":"Weights","tag":"Weights","type":"*variable","value":"~*req.3"},{"path":"MinCost","tag":"MinCost","type":"*variable","value":"~*req.4"},{"path":"MaxCost","tag":"MaxCost","type":"*variable","value":"~*req.5"},{"path":"MaxCostStrategy","tag":"MaxCostStrategy","type":"*variable","value":"~*req.6"},{"filters":["*notempty:~*req.7:"],"path":"Rates[\u003c~*req.7\u003e].FilterIDs","tag":"RateFilterIDs","type":"*variable","value":"~*req.8"},{"filters":["*notempty:~*req.7:"],"path":"Rates[\u003c~*req.7\u003e].ActivationTimes","tag":"RateActivationTimes","type":"*variable","value":"~*req.9"},{"filters":["*notempty:~*req.7:"],"path":"Rates[\u003c~*req.7\u003e].Weights","tag":"RateWeights","type":"*variable","value":"~*req.10"},{"filters":["*notempty:~*req.7:"],"path":"Rates[\u003c~*req.7\u003e].Blocker","tag":"RateBlocker","type":"*variable","value":"~*req.11"},{"filters":["*notempty:~*req.7:"],"new_branch":true,"path":"Rates[\u003c~*req.7\u003e].IntervalRates.IntervalStart","tag":"RateIntervalStart","type":"*variable","value":"~*req.12"},{"filters":["*notempty:~*req.7:"],"path":"Rates[\u003c~*req.7\u003e].IntervalRates.FixedFee","tag":"RateFixedFee","type":"*variable","value":"~*req.13"},{"filters":["*notempty:~*req.7:"],"path":"Rates[\u003c~*req.7\u003e].IntervalRates.RecurrentFee","tag":"RateRecurrentFee","type":"*variable","value":"~*req.14"},{"filters":["*notempty:~*req.7:"],"path":"Rates[\u003c~*req.7\u003e].IntervalRates.Unit","tag":"RateUnit","type":"*variable","value":"~*req.15"},{"filters":["*notempty:~*req.7:"],"path":"Rates[\u003c~*req.7\u003e].IntervalRates.Increment","tag":"RateIncrement","type":"*variable","value":"~*req.16"}],"file_name":"Rates.csv","flags":null,"type":"*rate_profiles"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"path":"FilterIDs","tag":"FilterIDs","type":"*variable","value":"~*req.2"},{"path":"Weights","tag":"Weights","type":"*variable","value":"~*req.3"},{"path":"Blockers","tag":"Blockers","type":"*variable","value":"~*req.4"},{"path":"Schedule","tag":"Schedule","type":"*variable","value":"~*req.5"},{"path":"Targets[\u003c~*req.6\u003e]","tag":"TargetIDs","type":"*variable","value":"~*req.7"},{"filters":["*notempty:~*req.8:"],"path":"Actions[\u003c~*req.8\u003e].FilterIDs","tag":"ActionFilterIDs","type":"*variable","value":"~*req.9"},{"filters":["*notempty:~*req.8:"],"path":"Actions[\u003c~*req.8\u003e].TTL","tag":"ActionTTL","type":"*variable","value":"~*req.10"},{"filters":["*notempty:~*req.8:"],"path":"Actions[\u003c~*req.8\u003e].Type","tag":"ActionType","type":"*variable","value":"~*req.11"},{"filters":["*notempty:~*req.8:"],"path":"Actions[\u003c~*req.8\u003e].Opts","tag":"ActionOpts","type":"*variable","value":"~*req.12"},{"filters":["*notempty:~*req.8:"],"new_branch":true,"path":"Actions[\u003c~*req.8\u003e].Diktats.Path","tag":"ActionPath","type":"*variable","value":"~*req.13"},{"filters":["*notempty:~*req.8:"],"path":"Actions[\u003c~*req.8\u003e].Diktats.Value","tag":"ActionValue","type":"*variable","value":"~*req.14"}],"file_name":"Actions.csv","flags":null,"type":"*action_profiles"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"path":"FilterIDs","tag":"FilterIDs","type":"*variable","value":"~*req.2"},{"path":"Weights","tag":"Weights","type":"*variable","value":"~*req.3"},{"path":"Blockers","tag":"Blockers","type":"*variable","value":"~*req.4"},{"path":"Opts","tag":"Opts","type":"*variable","value":"~*req.5"},{"filters":["*notempty:~*req.6:"],"path":"Balances[\u003c~*req.6\u003e].FilterIDs","tag":"BalanceFilterIDs","type":"*variable","value":"~*req.7"},{"filters":["*notempty:~*req.6:"],"path":"Balances[\u003c~*req.6\u003e].Weights","tag":"BalanceWeights","type":"*variable","value":"~*req.8"},{"filters":["*notempty:~*req.6:"],"path":"Balances[\u003c~*req.6\u003e].Blockers","tag":"BalanceBlockers","type":"*variable","value":"~*req.9"},{"filters":["*notempty:~*req.6:"],"path":"Balances[\u003c~*req.6\u003e].Type","tag":"BalanceType","type":"*variable","value":"~*req.10"},{"filters":["*notempty:~*req.6:"],"path":"Balances[\u003c~*req.6\u003e].Units","tag":"BalanceUnits","type":"*variable","value":"~*req.11"},{"filters":["*notempty:~*req.6:"],"path":"Balances[\u003c~*req.6\u003e].UnitFactors","tag":"BalanceUnitFactors","type":"*variable","value":"~*req.12"},{"filters":["*notempty:~*req.6:"],"path":"Balances[\u003c~*req.6\u003e].Opts","tag":"BalanceOpts","type":"*variable","value":"~*req.13"},{"filters":["*notempty:~*req.6:"],"path":"Balances[\u003c~*req.6\u003e].CostIncrements","tag":"BalanceCostIncrements","type":"*variable","value":"~*req.14"},{"filters":["*notempty:~*req.6:"],"path":"Balances[\u003c~*req.6\u003e].AttributeIDs","tag":"BalanceAttributeIDs","type":"*variable","value":"~*req.15"},{"filters":["*notempty:~*req.6:"],"path":"Balances[\u003c~*req.6\u003e].RateProfileIDs","tag":"BalanceRateProfileIDs","type":"*variable","value":"~*req.16"},{"path":"ThresholdIDs","tag":"ThresholdIDs","type":"*variable","value":"~*req.17"}],"file_name":"Accounts.csv","flags":null,"type":"*accounts"}],"enabled":false,"field_separator":",","id":"*default","lockfile_path":".cgr.lck","opts":{"*cache":"","*forceLock":false,"*stopOnError":false,"*withIndex":true},"run_delay":"0","tenant":"","tp_in_dir":"/var/spool/cgrates/loader/in","tp_out_dir":"/var/spool/cgrates/loader/out"}]}`
		var rpl26 string
		if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
			Tenant:   "cgrates.org",
			Sections: []string{config.LoaderSJSON},
		}, &rpl26); err != nil {
			t.Error(err)
		} else if cfgStr != rpl26 {
			t.Errorf("\nExpected %+v ,\n received: %+v", cfgStr, rpl26)
		}

	} else if testCfgDir == "tutmysql" || testCfgDir == "tutmongo" {
		cfgStr = `{"loaders":[{"action":"*store","cache":{"*accounts":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"5s"},"*action_profiles":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"5s"},"*attributes":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"5s"},"*chargers":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"5s"},"*dispatcher_hosts":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"5s"},"*dispatchers":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"5s"},"*filters":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"5s"},"*rate_profiles":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"5s"},"*resources":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"5s"},"*routes":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"5s"},"*stats":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"5s"},"*thresholds":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"5s"}},"caches_conns":["*internal"],"data":[{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"new_branch":true,"path":"Rules.Type","tag":"Type","type":"*variable","value":"~*req.2"},{"path":"Rules.Element","tag":"Element","type":"*variable","value":"~*req.3"},{"path":"Rules.Values","tag":"Values","type":"*variable","value":"~*req.4"}],"file_name":"Filters.csv","flags":null,"type":"*filters"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"TenantID","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ProfileID","type":"*variable","value":"~*req.1"},{"path":"FilterIDs","tag":"FilterIDs","type":"*variable","value":"~*req.2"},{"path":"Weights","tag":"Weights","type":"*variable","value":"~*req.3"},{"path":"Blockers","tag":"Blockers","type":"*variable","value":"~*req.4"},{"new_branch":true,"path":"Attributes.FilterIDs","tag":"AttributeFilterIDs","type":"*variable","value":"~*req.5"},{"path":"Attributes.Blockers","tag":"AttributeBlockers","type":"*variable","value":"~*req.6"},{"path":"Attributes.Path","tag":"Path","type":"*variable","value":"~*req.7"},{"path":"Attributes.Type","tag":"Type","type":"*variable","value":"~*req.8"},{"path":"Attributes.Value","tag":"Value","type":"*variable","value":"~*req.9"}],"file_name":"Attributes.csv","flags":null,"type":"*attributes"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"path":"FilterIDs","tag":"FilterIDs","type":"*variable","value":"~*req.2"},{"path":"Weights","tag":"Weights","type":"*variable","value":"~*req.3"},{"path":"UsageTTL","tag":"TTL","type":"*variable","value":"~*req.4"},{"path":"Limit","tag":"Limit","type":"*variable","value":"~*req.5"},{"path":"AllocationMessage","tag":"AllocationMessage","type":"*variable","value":"~*req.6"},{"path":"Blocker","tag":"Blocker","type":"*variable","value":"~*req.7"},{"path":"Stored","tag":"Stored","type":"*variable","value":"~*req.8"},{"path":"ThresholdIDs","tag":"ThresholdIDs","type":"*variable","value":"~*req.9"}],"file_name":"Resources.csv","flags":null,"type":"*resources"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"path":"FilterIDs","tag":"FilterIDs","type":"*variable","value":"~*req.2"},{"path":"Weights","tag":"Weights","type":"*variable","value":"~*req.3"},{"path":"Blockers","tag":"Blockers","type":"*variable","value":"~*req.4"},{"path":"QueueLength","tag":"QueueLength","type":"*variable","value":"~*req.5"},{"path":"TTL","tag":"TTL","type":"*variable","value":"~*req.6"},{"path":"MinItems","tag":"MinItems","type":"*variable","value":"~*req.7"},{"path":"Stored","tag":"Stored","type":"*variable","value":"~*req.8"},{"path":"ThresholdIDs","tag":"ThresholdIDs","type":"*variable","value":"~*req.9"},{"new_branch":true,"path":"Metrics.MetricID","tag":"MetricIDs","type":"*variable","value":"~*req.10"},{"path":"Metrics.FilterIDs","tag":"MetricFilterIDs","type":"*variable","value":"~*req.11"},{"path":"Metrics.Blockers","tag":"MetricBlockers","type":"*variable","value":"~*req.12"}],"file_name":"Stats.csv","flags":null,"type":"*stats"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"path":"FilterIDs","tag":"FilterIDs","type":"*variable","value":"~*req.2"},{"path":"Weights","tag":"Weights","type":"*variable","value":"~*req.3"},{"path":"MaxHits","tag":"MaxHits","type":"*variable","value":"~*req.4"},{"path":"MinHits","tag":"MinHits","type":"*variable","value":"~*req.5"},{"path":"MinSleep","tag":"MinSleep","type":"*variable","value":"~*req.6"},{"path":"Blocker","tag":"Blocker","type":"*variable","value":"~*req.7"},{"path":"ActionProfileIDs","tag":"ActionProfileIDs","type":"*variable","value":"~*req.8"},{"path":"Async","tag":"Async","type":"*variable","value":"~*req.9"}],"file_name":"Thresholds.csv","flags":null,"type":"*thresholds"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"path":"FilterIDs","tag":"FilterIDs","type":"*variable","value":"~*req.2"},{"path":"Weights","tag":"Weights","type":"*variable","value":"~*req.3"},{"path":"Blockers","tag":"Blockers","type":"*variable","value":"~*req.4"},{"path":"Sorting","tag":"Sorting","type":"*variable","value":"~*req.5"},{"path":"SortingParameters","tag":"SortingParameters","type":"*variable","value":"~*req.6"},{"new_branch":true,"path":"Routes.ID","tag":"RouteID","type":"*variable","value":"~*req.7"},{"path":"Routes.FilterIDs","tag":"RouteFilterIDs","type":"*variable","value":"~*req.8"},{"path":"Routes.AccountIDs","tag":"RouteAccountIDs","type":"*variable","value":"~*req.9"},{"path":"Routes.RateProfileIDs","tag":"RouteRateProfileIDs","type":"*variable","value":"~*req.10"},{"path":"Routes.ResourceIDs","tag":"RouteResourceIDs","type":"*variable","value":"~*req.11"},{"path":"Routes.StatIDs","tag":"RouteStatIDs","type":"*variable","value":"~*req.12"},{"path":"Routes.Weights","tag":"RouteWeights","type":"*variable","value":"~*req.13"},{"path":"Routes.Blockers","tag":"RouteBlockers","type":"*variable","value":"~*req.14"},{"path":"Routes.RouteParameters","tag":"RouteParameters","type":"*variable","value":"~*req.15"}],"file_name":"Routes.csv","flags":null,"type":"*routes"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"path":"FilterIDs","tag":"FilterIDs","type":"*variable","value":"~*req.2"},{"path":"Weights","tag":"Weights","type":"*variable","value":"~*req.3"},{"path":"Blockers","tag":"Blockers","type":"*variable","value":"~*req.4"},{"path":"RunID","tag":"RunID","type":"*variable","value":"~*req.5"},{"path":"AttributeIDs","tag":"AttributeIDs","type":"*variable","value":"~*req.6"}],"file_name":"Chargers.csv","flags":null,"type":"*chargers"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"path":"FilterIDs","tag":"FilterIDs","type":"*variable","value":"~*req.2"},{"path":"Weight","tag":"Weight","type":"*variable","value":"~*req.3"},{"path":"Strategy","tag":"Strategy","type":"*variable","value":"~*req.4"},{"path":"StrategyParams","tag":"StrategyParameters","type":"*variable","value":"~*req.5"},{"new_branch":true,"path":"Hosts.ID","tag":"ConnID","type":"*variable","value":"~*req.6"},{"path":"Hosts.FilterIDs","tag":"ConnFilterIDs","type":"*variable","value":"~*req.7"},{"path":"Hosts.Weight","tag":"ConnWeight","type":"*variable","value":"~*req.8"},{"path":"Hosts.Blocker","tag":"ConnBlocker","type":"*variable","value":"~*req.9"},{"path":"Hosts.Params","tag":"ConnParameters","type":"*variable","value":"~*req.10"}],"file_name":"DispatcherProfiles.csv","flags":null,"type":"*dispatchers"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"path":"Address","tag":"Address","type":"*variable","value":"~*req.2"},{"path":"Transport","tag":"Transport","type":"*variable","value":"~*req.3"},{"path":"ConnectAttempts","tag":"ConnectAttempts","type":"*variable","value":"~*req.4"},{"path":"Reconnects","tag":"Reconnects","type":"*variable","value":"~*req.5"},{"path":"MaxReconnectInterval","tag":"MaxReconnectInterval","type":"*variable","value":"~*req.6"},{"path":"ConnectTimeout","tag":"ConnectTimeout","type":"*variable","value":"~*req.7"},{"path":"ReplyTimeout","tag":"ReplyTimeout","type":"*variable","value":"~*req.8"},{"path":"TLS","tag":"TLS","type":"*variable","value":"~*req.9"},{"path":"ClientKey","tag":"ClientKey","type":"*variable","value":"~*req.10"},{"path":"ClientCertificate","tag":"ClientCertificate","type":"*variable","value":"~*req.11"},{"path":"CaCertificate","tag":"CaCertificate","type":"*variable","value":"~*req.12"}],"file_name":"DispatcherHosts.csv","flags":null,"type":"*dispatcher_hosts"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"path":"FilterIDs","tag":"FilterIDs","type":"*variable","value":"~*req.2"},{"path":"Weights","tag":"Weights","type":"*variable","value":"~*req.3"},{"path":"MinCost","tag":"MinCost","type":"*variable","value":"~*req.4"},{"path":"MaxCost","tag":"MaxCost","type":"*variable","value":"~*req.5"},{"path":"MaxCostStrategy","tag":"MaxCostStrategy","type":"*variable","value":"~*req.6"},{"filters":["*notempty:~*req.7:"],"path":"Rates[\u003c~*req.7\u003e].FilterIDs","tag":"RateFilterIDs","type":"*variable","value":"~*req.8"},{"filters":["*notempty:~*req.7:"],"path":"Rates[\u003c~*req.7\u003e].ActivationTimes","tag":"RateActivationTimes","type":"*variable","value":"~*req.9"},{"filters":["*notempty:~*req.7:"],"path":"Rates[\u003c~*req.7\u003e].Weights","tag":"RateWeights","type":"*variable","value":"~*req.10"},{"filters":["*notempty:~*req.7:"],"path":"Rates[\u003c~*req.7\u003e].Blocker","tag":"RateBlocker","type":"*variable","value":"~*req.11"},{"filters":["*notempty:~*req.7:"],"new_branch":true,"path":"Rates[\u003c~*req.7\u003e].IntervalRates.IntervalStart","tag":"RateIntervalStart","type":"*variable","value":"~*req.12"},{"filters":["*notempty:~*req.7:"],"path":"Rates[\u003c~*req.7\u003e].IntervalRates.FixedFee","tag":"RateFixedFee","type":"*variable","value":"~*req.13"},{"filters":["*notempty:~*req.7:"],"path":"Rates[\u003c~*req.7\u003e].IntervalRates.RecurrentFee","tag":"RateRecurrentFee","type":"*variable","value":"~*req.14"},{"filters":["*notempty:~*req.7:"],"path":"Rates[\u003c~*req.7\u003e].IntervalRates.Unit","tag":"RateUnit","type":"*variable","value":"~*req.15"},{"filters":["*notempty:~*req.7:"],"path":"Rates[\u003c~*req.7\u003e].IntervalRates.Increment","tag":"RateIncrement","type":"*variable","value":"~*req.16"}],"file_name":"Rates.csv","flags":null,"type":"*rate_profiles"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"path":"FilterIDs","tag":"FilterIDs","type":"*variable","value":"~*req.2"},{"path":"Weights","tag":"Weights","type":"*variable","value":"~*req.3"},{"path":"Blockers","tag":"Blockers","type":"*variable","value":"~*req.4"},{"path":"Schedule","tag":"Schedule","type":"*variable","value":"~*req.5"},{"path":"Targets[\u003c~*req.6\u003e]","tag":"TargetIDs","type":"*variable","value":"~*req.7"},{"filters":["*notempty:~*req.8:"],"path":"Actions[\u003c~*req.8\u003e].FilterIDs","tag":"ActionFilterIDs","type":"*variable","value":"~*req.9"},{"filters":["*notempty:~*req.8:"],"path":"Actions[\u003c~*req.8\u003e].TTL","tag":"ActionTTL","type":"*variable","value":"~*req.10"},{"filters":["*notempty:~*req.8:"],"path":"Actions[\u003c~*req.8\u003e].Type","tag":"ActionType","type":"*variable","value":"~*req.11"},{"filters":["*notempty:~*req.8:"],"path":"Actions[\u003c~*req.8\u003e].Opts","tag":"ActionOpts","type":"*variable","value":"~*req.12"},{"filters":["*notempty:~*req.8:"],"new_branch":true,"path":"Actions[\u003c~*req.8\u003e].Diktats.Path","tag":"ActionPath","type":"*variable","value":"~*req.13"},{"filters":["*notempty:~*req.8:"],"path":"Actions[\u003c~*req.8\u003e].Diktats.Value","tag":"ActionValue","type":"*variable","value":"~*req.14"}],"file_name":"Actions.csv","flags":null,"type":"*action_profiles"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"path":"FilterIDs","tag":"FilterIDs","type":"*variable","value":"~*req.2"},{"path":"Weights","tag":"Weights","type":"*variable","value":"~*req.3"},{"path":"Blockers","tag":"Blockers","type":"*variable","value":"~*req.4"},{"path":"Opts","tag":"Opts","type":"*variable","value":"~*req.5"},{"filters":["*notempty:~*req.6:"],"path":"Balances[\u003c~*req.6\u003e].FilterIDs","tag":"BalanceFilterIDs","type":"*variable","value":"~*req.7"},{"filters":["*notempty:~*req.6:"],"path":"Balances[\u003c~*req.6\u003e].Weights","tag":"BalanceWeights","type":"*variable","value":"~*req.8"},{"filters":["*notempty:~*req.6:"],"path":"Balances[\u003c~*req.6\u003e].Blockers","tag":"BalanceBlockers","type":"*variable","value":"~*req.9"},{"filters":["*notempty:~*req.6:"],"path":"Balances[\u003c~*req.6\u003e].Type","tag":"BalanceType","type":"*variable","value":"~*req.10"},{"filters":["*notempty:~*req.6:"],"path":"Balances[\u003c~*req.6\u003e].Units","tag":"BalanceUnits","type":"*variable","value":"~*req.11"},{"filters":["*notempty:~*req.6:"],"path":"Balances[\u003c~*req.6\u003e].UnitFactors","tag":"BalanceUnitFactors","type":"*variable","value":"~*req.12"},{"filters":["*notempty:~*req.6:"],"path":"Balances[\u003c~*req.6\u003e].Opts","tag":"BalanceOpts","type":"*variable","value":"~*req.13"},{"filters":["*notempty:~*req.6:"],"path":"Balances[\u003c~*req.6\u003e].CostIncrements","tag":"BalanceCostIncrements","type":"*variable","value":"~*req.14"},{"filters":["*notempty:~*req.6:"],"path":"Balances[\u003c~*req.6\u003e].AttributeIDs","tag":"BalanceAttributeIDs","type":"*variable","value":"~*req.15"},{"filters":["*notempty:~*req.6:"],"path":"Balances[\u003c~*req.6\u003e].RateProfileIDs","tag":"BalanceRateProfileIDs","type":"*variable","value":"~*req.16"},{"path":"ThresholdIDs","tag":"ThresholdIDs","type":"*variable","value":"~*req.17"}],"file_name":"Accounts.csv","flags":null,"type":"*accounts"}],"enabled":false,"field_separator":",","id":"*default","lockfile_path":".cgr.lck","opts":{"*cache":"","*forceLock":false,"*stopOnError":false,"*withIndex":true},"run_delay":"0","tenant":"","tp_in_dir":"/var/spool/cgrates/loader/in","tp_out_dir":"/var/spool/cgrates/loader/out"}]}`
		var rpl26 string
		if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
			Tenant:   "cgrates.org",
			Sections: []string{config.LoaderSJSON},
		}, &rpl26); err != nil {
			t.Error(err)
		} else if cfgStr != rpl26 {
			t.Errorf("\nExpected %+v ,\n received: %+v", cfgStr, rpl26)
		}
	}

	cfgStr = `{"suretax":{"bill_to_number":"","business_unit":"","client_number":"","client_tracking":"~*opts.*originID","customer_number":"~*req.Subject","include_local_cost":false,"orig_number":"~*req.Subject","p2pplus4":"","p2pzipcode":"","plus4":"","regulatory_code":"03","response_group":"03","response_type":"D4","return_file_code":"0","sales_type_code":"R","tax_exemption_code_list":"","tax_included":"0","tax_situs_rule":"04","term_number":"~*req.Destination","timezone":"Local","trans_type_code":"010101","unit_type":"00","units":"1","url":"","validation_key":"","zipcode":""}}`
	var rpl28 string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.SureTaxJSON},
	}, &rpl28); err != nil {
		t.Error(err)
	} else if cfgStr != rpl28 {
		t.Errorf("\nExpected %+v ,\n received: %+v", cfgStr, rpl28)
	}

	cfgStr = `{"loader":{"actions_conns":["*localhost"],"caches_conns":["*localhost"],"data_path":"./","disable_reverse":false,"field_separator":",","gapi_credentials":".gapi/credentials.json","gapi_token":".gapi/token.json","tpid":""}}`
	var rpl29 string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.LoaderJSON},
	}, &rpl29); err != nil {
		t.Error(err)
	} else if cfgStr != rpl29 {
		t.Errorf("\nExpected %+v ,\n received: %+v", cfgStr, rpl29)
	}
	if testCfgDir == "tutmysql" {
		cfgStr = `{"migrator":{"out_datadb_encoding":"msgpack","out_datadb_host":"127.0.0.1","out_datadb_name":"10","out_datadb_opts":{"mongoConnScheme":"mongodb","mongoQueryTimeout":"10s","redisCACertificate":"","redisClientCertificate":"","redisClientKey":"","redisCluster":false,"redisClusterOndownDelay":"0s","redisClusterSync":"5s","redisConnectAttempts":20,"redisConnectTimeout":"0s","redisMaxConns":10,"redisReadTimeout":"0s","redisSentinel":"","redisTLS":false,"redisWriteTimeout":"0s"},"out_datadb_password":"","out_datadb_port":"6379","out_datadb_type":"*redis","out_datadb_user":"cgrates","users_filters":["Account"]}}`
		var rpl30 string
		if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
			Tenant:   "cgrates.org",
			Sections: []string{config.MigratorJSON},
		}, &rpl30); err != nil {
			t.Error(err)
		} else if cfgStr != rpl30 {
			t.Errorf("\nExpected %+v ,\n received: %+v", cfgStr, rpl30)
		}
	} else if testCfgDir == "tutmongo" {
		cfgStr = `{"migrator":{"out_datadb_encoding":"msgpack","out_datadb_host":"127.0.0.1","out_datadb_name":"10","out_datadb_opts":{"mongoConnScheme":"mongodb","mongoQueryTimeout":"10s","redisCACertificate":"","redisClientCertificate":"","redisClientKey":"","redisCluster":false,"redisClusterOndownDelay":"0s","redisClusterSync":"5s","redisConnectAttempts":20,"redisConnectTimeout":"0s","redisMaxConns":10,"redisReadTimeout":"0s","redisSentinel":"","redisTLS":false,"redisWriteTimeout":"0s"},"out_datadb_password":"","out_datadb_port":"27017","out_datadb_type":"*mongo","out_datadb_user":"cgrates","users_filters":["Account"]}}`
		var rpl30 string
		if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
			Tenant:   "cgrates.org",
			Sections: []string{config.MigratorJSON},
		}, &rpl30); err != nil {
			t.Error(err)
		} else if cfgStr != rpl30 {
			t.Errorf("\nExpected %+v ,\n received: %+v", cfgStr, rpl30)
		}
	}
	cfgStr = `{"dispatchers":{"attributes_conns":[],"enabled":false,"exists_indexed_fields":[],"indexed_selects":true,"nested_fields":false,"notexists_indexed_fields":[],"opts":{"*dispatchers":[]},"prefix_indexed_fields":[],"suffix_indexed_fields":[]}}`
	var rpl31 string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.DispatcherSJSON},
	}, &rpl31); err != nil {
		t.Error(err)
	} else if cfgStr != rpl31 {
		t.Errorf("\nExpected %+v ,\n received: %+v", cfgStr, rpl31)
	}

	cfgStr = `{"registrarc":{"dispatchers":{"hosts":[],"refresh_interval":"5m0s","registrars_conns":[]},"rpc":{"hosts":[],"refresh_interval":"5m0s","registrars_conns":[]}}}`
	var rpl32 string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.RegistrarCJSON},
	}, &rpl32); err != nil {
		t.Error(err)
	} else if cfgStr != rpl32 {
		t.Errorf("\nExpected %+v ,\n received: %+v", cfgStr, rpl32)
	}
	cfgStr = `{"analyzers":{"cleanup_interval":"1h0m0s","db_path":"/var/spool/cgrates/analyzers","ees_conns":[],"enabled":false,"index_type":"*scorch","opts":{"*exporterIDs":[]},"ttl":"24h0m0s"}}`
	var rpl33 string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.AnalyzerSJSON},
	}, &rpl33); err != nil {
		t.Error(err)
	} else if cfgStr != rpl33 {
		t.Errorf("\nExpected %+v ,\n received: %+v", cfgStr, rpl33)
	}

	cfgStr = `{"sip_agent":{"enabled":false,"listen":"127.0.0.1:5060","listen_net":"udp","request_processors":[],"retransmission_timer":"1s","sessions_conns":["*internal"],"timezone":""}}`
	var rpl35 string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.SIPAgentJSON},
	}, &rpl35); err != nil {
		t.Error(err)
	} else if cfgStr != rpl35 {
		t.Errorf("\nExpected %+v ,\n received: %+v", cfgStr, rpl35)
	}

	cfgStr = `{"templates":{"*asr":[{"mandatory":true,"path":"*diamreq.Session-Id","tag":"SessionId","type":"*variable","value":"~*req.Session-Id"},{"mandatory":true,"path":"*diamreq.Origin-Host","tag":"OriginHost","type":"*variable","value":"~*req.Destination-Host"},{"mandatory":true,"path":"*diamreq.Origin-Realm","tag":"OriginRealm","type":"*variable","value":"~*req.Destination-Realm"},{"mandatory":true,"path":"*diamreq.Destination-Realm","tag":"DestinationRealm","type":"*variable","value":"~*req.Origin-Realm"},{"mandatory":true,"path":"*diamreq.Destination-Host","tag":"DestinationHost","type":"*variable","value":"~*req.Origin-Host"},{"mandatory":true,"path":"*diamreq.Auth-Application-Id","tag":"AuthApplicationId","type":"*variable","value":"~*vars.*appid"}],"*cca":[{"mandatory":true,"path":"*rep.Session-Id","tag":"SessionId","type":"*variable","value":"~*req.Session-Id"},{"path":"*rep.Result-Code","tag":"ResultCode","type":"*constant","value":"2001"},{"mandatory":true,"path":"*rep.Origin-Host","tag":"OriginHost","type":"*variable","value":"~*vars.OriginHost"},{"mandatory":true,"path":"*rep.Origin-Realm","tag":"OriginRealm","type":"*variable","value":"~*vars.OriginRealm"},{"mandatory":true,"path":"*rep.Auth-Application-Id","tag":"AuthApplicationId","type":"*variable","value":"~*vars.*appid"},{"mandatory":true,"path":"*rep.CC-Request-Type","tag":"CCRequestType","type":"*variable","value":"~*req.CC-Request-Type"},{"mandatory":true,"path":"*rep.CC-Request-Number","tag":"CCRequestNumber","type":"*variable","value":"~*req.CC-Request-Number"}],"*cdrLog":[{"mandatory":true,"path":"*cdr.ToR","tag":"ToR","type":"*variable","value":"~*req.BalanceType"},{"mandatory":true,"path":"*cdr.OriginHost","tag":"OriginHost","type":"*constant","value":"127.0.0.1"},{"mandatory":true,"path":"*cdr.RequestType","tag":"RequestType","type":"*constant","value":"*none"},{"mandatory":true,"path":"*cdr.Tenant","tag":"Tenant","type":"*variable","value":"~*req.Tenant"},{"mandatory":true,"path":"*cdr.Account","tag":"Account","type":"*variable","value":"~*req.Account"},{"mandatory":true,"path":"*cdr.Subject","tag":"Subject","type":"*variable","value":"~*req.Account"},{"mandatory":true,"path":"*cdr.Cost","tag":"Cost","type":"*variable","value":"~*req.Cost"},{"mandatory":true,"path":"*cdr.Source","tag":"Source","type":"*constant","value":"*cdrLog"},{"mandatory":true,"path":"*cdr.Usage","tag":"Usage","type":"*constant","value":"1"},{"mandatory":true,"path":"*cdr.RunID","tag":"RunID","type":"*variable","value":"~*req.ActionType"},{"mandatory":true,"path":"*cdr.SetupTime","tag":"SetupTime","type":"*constant","value":"*now"},{"mandatory":true,"path":"*cdr.AnswerTime","tag":"AnswerTime","type":"*constant","value":"*now"},{"mandatory":true,"path":"*cdr.PreRated","tag":"PreRated","type":"*constant","value":"true"}],"*err":[{"mandatory":true,"path":"*rep.Session-Id","tag":"SessionId","type":"*variable","value":"~*req.Session-Id"},{"mandatory":true,"path":"*rep.Origin-Host","tag":"OriginHost","type":"*variable","value":"~*vars.OriginHost"},{"mandatory":true,"path":"*rep.Origin-Realm","tag":"OriginRealm","type":"*variable","value":"~*vars.OriginRealm"}],"*errSip":[{"mandatory":true,"path":"*rep.Request","tag":"Request","type":"*constant","value":"SIP/2.0 500 Internal Server Error"}],"*rar":[{"mandatory":true,"path":"*diamreq.Session-Id","tag":"SessionId","type":"*variable","value":"~*req.Session-Id"},{"mandatory":true,"path":"*diamreq.Origin-Host","tag":"OriginHost","type":"*variable","value":"~*req.Destination-Host"},{"mandatory":true,"path":"*diamreq.Origin-Realm","tag":"OriginRealm","type":"*variable","value":"~*req.Destination-Realm"},{"mandatory":true,"path":"*diamreq.Destination-Realm","tag":"DestinationRealm","type":"*variable","value":"~*req.Origin-Realm"},{"mandatory":true,"path":"*diamreq.Destination-Host","tag":"DestinationHost","type":"*variable","value":"~*req.Origin-Host"},{"mandatory":true,"path":"*diamreq.Auth-Application-Id","tag":"AuthApplicationId","type":"*variable","value":"~*vars.*appid"},{"path":"*diamreq.Re-Auth-Request-Type","tag":"ReAuthRequestType","type":"*constant","value":"0"}]}}`
	var rpl36 string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.TemplatesJSON},
	}, &rpl36); err != nil {
		t.Error(err)
	} else if cfgStr != rpl36 {
		t.Errorf("\nExpected %+v ,\n received: %+v", cfgStr, rpl36)
	}
	cfgStr = `{"configs":{"enabled":false,"root_dir":"/var/spool/cgrates/configs","url":"/configs/"}}`
	var rpl37 string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.ConfigSJSON},
	}, &rpl37); err != nil {
		t.Error(err)
	} else if cfgStr != rpl37 {
		t.Errorf("\nExpected %+v ,\n received: %+v", cfgStr, rpl37)
	}

	cfgStr = `{"apiban":{"enabled":false,"keys":[]}}`
	var rpl38 string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.APIBanJSON},
	}, &rpl38); err != nil {
		t.Error(err)
	} else if cfgStr != rpl38 {
		t.Errorf("\nExpected %+v ,\n received: %+v", cfgStr, rpl38)
	}
}

func testStopCgrEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
