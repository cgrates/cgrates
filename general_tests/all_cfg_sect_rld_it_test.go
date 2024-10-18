//go:build flaky

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
	"os"
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
	testSectCfgDir  string
	testSectCfgPath string
	testSectCfg     *config.CGRConfig
	testSectRPC     *birpc.Client

	testSectTests = []func(t *testing.T){
		testSectLoadConfig,
		testSectResetDataDB,
		testSectResetStorDb,
		testSectStartEngine,
		testSectRPCConn,
		//testSectConfigSReloadGeneral,
		testSectConfigSReloadCores,
		testSectConfigSReloadRPCConns,
		//testSectConfigSReloadDataDB,
		//testSectConfigSReloadStorDB,
		testSectConfigSReloadListen,
		testSectConfigSReloadThresholds,
		testSectConfigSReloadRALS,

		testSectConfigSReloadRoutes,
		testSectConfigSReloadAttributes,
		testSectConfigSReloadResources,
		testSectConfigSReloadChargers,
		testSectConfigSReloadSessions,

		testSectConfigSReloadDNSAgent,
		testSectConfigSReloadHTTPAgent,
		testSectConfigSReloadApiers,

		testSectConfigSReloadRegistrarC,

		testSectConfigSReloadTLS,
		testSectConfigSReloadHTTP,
		testSectConfigSReloadSchedulers,
		testSectConfigSReloadCaches,
		testSectConfigSReloadFilters,

		testSectConfigSReloadCDRS,
		testSectConfigSReloadERS,
		testSectConfigSReloadEES,
		testSectConfigSReloadRadiusAgent,

		testSectConfigSReloadDiameterAgent,
		testSectConfigSReloadHTTPAgent,

		testSectConfigSReloadStats,

		testSectConfigSReloadLoaders,
		testSectConfigSReloadMailer,

		testSectConfigSReloadLoader,
		testSectConfigSReloadDispatchers,
		testSectConfigSReloadAnalyzer,

		testSectConfigSReloadConfigs,
		testSectConfigSReloadAPIBan,
		testSectConfigSReloadTemplates,
		testSectConfigSReloadSIPAgent,
		testSectConfigSReloadMigrator,
		testSectConfigSReloadSuretax,
		testSectConfigSReloadAsteriskAgent,
		//testSectConfigSReloadKamailioAgent,
		//testSectConfigSReloadFreeswitchAgent,
		testSectStopCgrEngine,
	}
)

func TestSectChange(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		testSectCfgDir = "reload_sect_internal"
	case utils.MetaMySQL:
		testSectCfgDir = "tutmysql"
	case utils.MetaMongo:
		testSectCfgDir = "reload_sect_mongo"
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
	var err error
	testSectCfgPath = path.Join(*utils.DataDir, "conf", "samples", testSectCfgDir)
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
	if err := os.MkdirAll("/var/spool/cgrates/analyzers", 0755); err != nil {
		t.Error(err)
	}
	if _, err := engine.StopStartEngine(testSectCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

func testSectRPCConn(t *testing.T) {
	testSectRPC = engine.NewRPCClient(t, testSectCfg.ListenCfg())
}

func testSectConfigSReloadCores(t *testing.T) {
	var replyPingBf string
	if err := testSectRPC.Call(context.Background(), utils.CoreSv1Ping, &utils.CGREvent{}, &replyPingBf); err != nil {
		t.Error(err)
	} else if replyPingBf != utils.Pong {
		t.Errorf("Expected OK received: %s", replyPingBf)
	}
	var reply string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1SetConfigFromJSON, &config.SetConfigFromJSONArgs{
		Tenant: "cgrates.org",
		Config: "{\"cores\":{\"caps\":0,\"caps_stats_interval\":\"0\",\"caps_strategy\":\"*busy\",\"shutdown_timeout\":\"1s\"}}",
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := "{\"cores\":{\"caps\":0,\"caps_stats_interval\":\"0\",\"caps_strategy\":\"*busy\",\"shutdown_timeout\":\"1s\"}}"
	var rpl string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:  "cgrates.org",
		Section: config.CoreSCfgJson,
	}, &rpl); err != nil {
		t.Error(err)
	} else if cfgStr != rpl {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl))
	}

	var replyPingAf string
	if err := testSectRPC.Call(context.Background(), utils.CoreSv1Ping, &utils.CGREvent{}, &replyPingAf); err != nil {
		t.Error(err)
	} else if replyPingAf != utils.Pong {
		t.Errorf("Expected OK received: %s", replyPingAf)
	}
}

func testSectConfigSReloadRPCConns(t *testing.T) {

	var reply string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1SetConfigFromJSON, &config.SetConfigFromJSONArgs{
		Tenant: "cgrates.org",
		Config: "{\"rpc_conns\":{\"*bijson_localhost\":{\"conns\":[{\"address\":\"127.0.0.1:2014\",\"transport\":\"*birpc_json\"}],\"poolSize\":0,\"strategy\":\"*first\"},\"*birpc_internal\":{\"conns\":[{\"address\":\"*birpc_internal\",\"transport\":\"\"}],\"poolSize\":0,\"strategy\":\"*first\"},\"*internal\":{\"conns\":[{\"address\":\"*internal\",\"transport\":\"\"}],\"poolSize\":0,\"strategy\":\"*first\"},\"*localhost\":{\"conns\":[{\"address\":\"127.0.0.1:2012\",\"transport\":\"*json\"}],\"poolSize\":0,\"strategy\":\"*first\"}}}",
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := "{\"rpc_conns\":{\"*bijson_localhost\":{\"conns\":[{\"address\":\"127.0.0.1:2014\",\"transport\":\"*birpc_json\"}],\"poolSize\":0,\"strategy\":\"*first\"},\"*birpc_internal\":{\"conns\":[{\"address\":\"*birpc_internal\",\"transport\":\"\"}],\"poolSize\":0,\"strategy\":\"*first\"},\"*internal\":{\"conns\":[{\"address\":\"*internal\",\"transport\":\"\"}],\"poolSize\":0,\"strategy\":\"*first\"},\"*localhost\":{\"conns\":[{\"address\":\"127.0.0.1:2012\",\"transport\":\"*json\"}],\"poolSize\":0,\"strategy\":\"*first\"}}}"
	var rpl string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:  "cgrates.org",
		Section: config.RPCConnsJsonName,
	}, &rpl); err != nil {
		t.Error(err)
	} else if cfgStr != rpl {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl))
	}
}

func testSectConfigSReloadDataDB(t *testing.T) {

	var reply string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1SetConfigFromJSON, &config.SetConfigFromJSONArgs{
		Tenant: "cgrates.org",
		Config: "{\"data_db\":{\"db_host\":\"127.0.0.1\",\"db_name\":\"10\",\"db_password\":\"\",\"db_port\":6379,\"db_type\":\"*internal\",\"db_user\":\"cgrates\",\"items\":{\"*account_action_plans\":{\"remote\":false,\"replicate\":false},\"*accounts\":{\"remote\":false,\"replicate\":false},\"*action_plans\":{\"remote\":false,\"replicate\":false},\"*action_triggers\":{\"remote\":false,\"replicate\":false},\"*actions\":{\"remote\":false,\"replicate\":false},\"*attribute_profiles\":{\"remote\":false,\"replicate\":false},\"*charger_profiles\":{\"remote\":false,\"replicate\":false},\"*destinations\":{\"remote\":false,\"replicate\":false},\"*dispatcher_hosts\":{\"remote\":false,\"replicate\":false},\"*dispatcher_profiles\":{\"remote\":false,\"replicate\":false},\"*filters\":{\"remote\":false,\"replicate\":false},\"*indexes\":{\"remote\":false,\"replicate\":false},\"*load_ids\":{\"remote\":false,\"replicate\":false},\"*rating_plans\":{\"remote\":false,\"replicate\":false},\"*rating_profiles\":{\"remote\":false,\"replicate\":false},\"*resource_profiles\":{\"remote\":false,\"replicate\":false},\"*resources\":{\"remote\":false,\"replicate\":false},\"*reverse_destinations\":{\"remote\":false,\"replicate\":false},\"*route_profiles\":{\"remote\":false,\"replicate\":false},\"*shared_groups\":{\"remote\":false,\"replicate\":false},\"*statqueue_profiles\":{\"remote\":false,\"replicate\":false},\"*statqueues\":{\"remote\":false,\"replicate\":false},\"*threshold_profiles\":{\"remote\":false,\"replicate\":false},\"*thresholds\":{\"remote\":false,\"replicate\":false},\"*timings\":{\"remote\":false,\"replicate\":false}},\"opts\":{\"mongoQueryTimeout\":\"10s\",\"redisCACertificate\":\"\",\"redisClientCertificate\":\"\",\"redisClientKey\":\"\",\"redisCluster\":false,\"redisClusterOndownDelay\":\"0\",\"redisClusterSync\":\"5s\",\"redisSentinel\":\"\",\"redisTLS\":false},\"remote_conn_id\":\"\",\"remote_conns\":[],\"replication_cache\":\"\",\"replication_conns\":[],\"replication_filtered\":false}}",
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := "{\"data_db\":{\"db_host\":\"127.0.0.1\",\"db_name\":\"10\",\"db_password\":\"\",\"db_port\":6379,\"db_type\":\"*internal\",\"db_user\":\"cgrates\",\"items\":{\"*account_action_plans\":{\"remote\":false,\"replicate\":false},\"*accounts\":{\"remote\":false,\"replicate\":false},\"*action_plans\":{\"remote\":false,\"replicate\":false},\"*action_triggers\":{\"remote\":false,\"replicate\":false},\"*actions\":{\"remote\":false,\"replicate\":false},\"*attribute_profiles\":{\"remote\":false,\"replicate\":false},\"*charger_profiles\":{\"remote\":false,\"replicate\":false},\"*destinations\":{\"remote\":false,\"replicate\":false},\"*dispatcher_hosts\":{\"remote\":false,\"replicate\":false},\"*dispatcher_profiles\":{\"remote\":false,\"replicate\":false},\"*filters\":{\"remote\":false,\"replicate\":false},\"*indexes\":{\"remote\":false,\"replicate\":false},\"*load_ids\":{\"remote\":false,\"replicate\":false},\"*rating_plans\":{\"remote\":false,\"replicate\":false},\"*rating_profiles\":{\"remote\":false,\"replicate\":false},\"*resource_profiles\":{\"remote\":false,\"replicate\":false},\"*resources\":{\"remote\":false,\"replicate\":false},\"*reverse_destinations\":{\"remote\":false,\"replicate\":false},\"*route_profiles\":{\"remote\":false,\"replicate\":false},\"*shared_groups\":{\"remote\":false,\"replicate\":false},\"*statqueue_profiles\":{\"remote\":false,\"replicate\":false},\"*statqueues\":{\"remote\":false,\"replicate\":false},\"*threshold_profiles\":{\"remote\":false,\"replicate\":false},\"*thresholds\":{\"remote\":false,\"replicate\":false},\"*timings\":{\"remote\":false,\"replicate\":false}},\"opts\":{\"mongoQueryTimeout\":\"10s\",\"redisCACertificate\":\"\",\"redisClientCertificate\":\"\",\"redisClientKey\":\"\",\"redisCluster\":false,\"redisClusterOndownDelay\":\"0\",\"redisClusterSync\":\"5s\",\"redisSentinel\":\"\",\"redisTLS\":false},\"remote_conn_id\":\"\",\"remote_conns\":[],\"replication_cache\":\"\",\"replication_conns\":[],\"replication_filtered\":false}}"
	var rpl string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:  "cgrates.org",
		Section: config.DATADB_JSN,
	}, &rpl); err != nil {
		t.Error(err)
	} else if cfgStr != rpl {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl))
	}
}

func testSectConfigSReloadStorDB(t *testing.T) {

	var reply string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1SetConfigFromJSON, &config.SetConfigFromJSONArgs{
		Tenant: "cgrates.org",
		Config: "{\"stor_db\":{\"db_host\":\"127.0.0.1\",\"db_name\":\"cgrates\",\"db_password\":\"CGRateS.org\",\"db_port\":3306,\"db_type\":\"*internal\",\"db_user\":\"cgrates\",\"items\":{\"*cdrs\":{\"remote\":false,\"replicate\":false},\"*tp_actions\":{\"remote\":false,\"replicate\":false},\"*tp_attributes\":{\"remote\":false,\"replicate\":false},\"*tp_chargers\":{\"remote\":false,\"replicate\":false},\"*tp_destination_rates\":{\"remote\":false,\"replicate\":false},\"*tp_dispatcher_hosts\":{\"remote\":false,\"replicate\":false},\"*tp_dispatcher_profiles\":{\"remote\":false,\"replicate\":false},\"*tp_filters\":{\"remote\":false,\"replicate\":false},\"*tp_rates\":{\"remote\":false,\"replicate\":false},\"*tp_rating_plans\":{\"remote\":false,\"replicate\":false},\"*tp_shared_groups\":{\"remote\":false,\"replicate\":false},\"*tp_stats\":{\"remote\":false,\"replicate\":false},\"*tp_thresholds\":{\"remote\":false,\"replicate\":false},\"*tp_timings\":{\"remote\":false,\"replicate\":false},\"opts\":{\"mongoQueryTimeout\":\"10s\",\"mysqlLocation\":\"Local\",\"postgresSSLMode\":\"disable\",\"sqlConnMaxLifetime\":0,\"sqlMaxIdleConns\":10,\"sqlMaxOpenConns\":100},\"prefix_indexed_fields\":[],\"remote_conns\":null,\"replication_conns\":null,\"string_indexed_fields\":[]}}",
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := "{\"stor_db\":{\"db_host\":\"127.0.0.1\",\"db_name\":\"cgrates\",\"db_password\":\"CGRateS.org\",\"db_port\":3306,\"db_type\":\"*internal\",\"db_user\":\"cgrates\",\"items\":{\"*cdrs\":{\"remote\":false,\"replicate\":false},\"*tp_action_triggers\":{\"remote\":false,\"replicate\":false},\"*tp_actions\":{\"remote\":false,\"replicate\":false},\"*tp_attributes\":{\"remote\":false,\"replicate\":false},\"*tp_chargers\":{\"remote\":false,\"replicate\":false},\"*tp_destination_rates\":{\"remote\":false,\"replicate\":false},\"*tp_dispatcher_hosts\":{\"remote\":false,\"replicate\":false},\"*tp_dispatcher_profiles\":{\"remote\":false,\"replicate\":false},\"*tp_filters\":{\"remote\":false,\"replicate\":false},\"*tp_rates\":{\"remote\":false,\"replicate\":false},\"*tp_rating_plans\":{\"remote\":false,\"replicate\":false},\"*tp_shared_groups\":{\"remote\":false,\"replicate\":false},\"*tp_stats\":{\"remote\":false,\"replicate\":false},\"*tp_thresholds\":{\"remote\":false,\"replicate\":false},\"*tp_timings\":{\"remote\":false,\"replicate\":false},\"opts\":{\"mongoQueryTimeout\":\"10s\",\"mysqlLocation\":\"Local\",\"postgresSSLMode\":\"disable\",\"sqlConnMaxLifetime\":0,\"sqlMaxIdleConns\":10,\"sqlMaxOpenConns\":100},\"prefix_indexed_fields\":[],\"remote_conns\":null,\"replication_conns\":null,\"string_indexed_fields\":[]}}"
	var rpl string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:  "cgrates.org",
		Section: config.STORDB_JSN,
	}, &rpl); err != nil {
		t.Error(err)
	} else if cfgStr != rpl {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl))
	}
}

func testSectConfigSReloadListen(t *testing.T) {

	var reply string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1SetConfigFromJSON, &config.SetConfigFromJSONArgs{
		Tenant: "cgrates.org",
		Config: "{\"listen\":{\"http\":\":2080\",\"http_tls\":\"127.0.0.1:2280\",\"rpc_gob\":\":2013\",\"rpc_gob_tls\":\"127.0.0.1:2023\",\"rpc_json\":\":2012\",\"rpc_json_tls\":\"127.0.0.1:2022\"}}",
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := "{\"listen\":{\"http\":\":2080\",\"http_tls\":\"127.0.0.1:2280\",\"rpc_gob\":\":2013\",\"rpc_gob_tls\":\"127.0.0.1:2023\",\"rpc_json\":\":2012\",\"rpc_json_tls\":\"127.0.0.1:2022\"}}"
	var rpl string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:  "cgrates.org",
		Section: config.LISTEN_JSN,
	}, &rpl); err != nil {
		t.Error(err)
	} else if cfgStr != rpl {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl))
	}
}

func testSectConfigSReloadTLS(t *testing.T) {

	var reply string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1SetConfigFromJSON, &config.SetConfigFromJSONArgs{
		Tenant: "cgrates.org",
		Config: "{\"tls\":{\"ca_certificate\":\"\",\"client_certificate\":\"\",\"client_key\":\"\",\"server_certificate\":\"\",\"server_key\":\"\",\"server_name\":\"\",\"server_policy\":4}}",
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := "{\"tls\":{\"ca_certificate\":\"\",\"client_certificate\":\"\",\"client_key\":\"\",\"server_certificate\":\"\",\"server_key\":\"\",\"server_name\":\"\",\"server_policy\":4}}"
	var rpl string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:  "cgrates.org",
		Section: config.TlsCfgJson,
	}, &rpl); err != nil {
		t.Error(err)
	} else if cfgStr != rpl {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl))
	}
}

func testSectConfigSReloadHTTP(t *testing.T) {

	var reply string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1SetConfigFromJSON, &config.SetConfigFromJSONArgs{
		Tenant: "cgrates.org",
		Config: `{"http":{"auth_users":{},"client_opts":{"dialFallbackDelay":"300ms","dialKeepAlive":"30s","dialTimeout":"30s","disableCompression":false,"disableKeepAlives":false,"expectContinueTimeout":"0","forceAttemptHttp2":true,"idleConnTimeout":"90s","maxConnsPerHost":0,"maxIdleConns":100,"maxIdleConnsPerHost":2,"responseHeaderTimeout":"0","skipTlsVerify":false,"tlsHandshakeTimeout":"10s"},"freeswitch_cdrs_url":"/freeswitch_json","http_cdrs":"/cdr_http","json_rpc_url":"/jsonrpc","registrars_url":"/registrar","use_basic_auth":false,"ws_url":"/ws"}}`,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := `{"http":{"auth_users":{},"client_opts":{"dialFallbackDelay":"300ms","dialKeepAlive":"30s","dialTimeout":"30s","disableCompression":false,"disableKeepAlives":false,"expectContinueTimeout":"0s","forceAttemptHttp2":true,"idleConnTimeout":"1m30s","maxConnsPerHost":0,"maxIdleConns":100,"maxIdleConnsPerHost":2,"responseHeaderTimeout":"0s","skipTlsVerify":false,"tlsHandshakeTimeout":"10s"},"freeswitch_cdrs_url":"/freeswitch_json","http_cdrs":"/cdr_http","json_rpc_url":"/jsonrpc","prometheus_url":"/prometheus","registrars_url":"/registrar","use_basic_auth":false,"ws_url":"/ws"}}`

	var rpl string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:  "cgrates.org",
		Section: config.HTTP_JSN,
	}, &rpl); err != nil {
		t.Error(err)
	} else if cfgStr != rpl {
		t.Errorf("\nExpected %+v ,\n received: %+v", cfgStr, rpl)
	}
}

func testSectConfigSReloadSchedulers(t *testing.T) {

	var replyPingBf string
	if err := testSectRPC.Call(context.Background(), utils.SchedulerSv1Ping, &utils.CGREvent{}, &replyPingBf); err != nil {
		t.Error(err)
	} else if replyPingBf != utils.Pong {
		t.Errorf("Expected OK received: %s", replyPingBf)
	}

	var reply string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1SetConfigFromJSON, &config.SetConfigFromJSONArgs{
		Tenant: "cgrates.org",
		Config: "{\"schedulers\":{\"cdrs_conns\":[\"*internal\"],\"dynaprepaid_actionplans\":[],\"enabled\":true,\"filters\":[],\"stats_conns\":[\"*localhost\"],\"thresholds_conns\":[]}}",
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := "{\"schedulers\":{\"cdrs_conns\":[\"*internal\"],\"dynaprepaid_actionplans\":[],\"enabled\":true,\"filters\":[],\"stats_conns\":[\"*localhost\"],\"thresholds_conns\":[]}}"
	var rpl string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:  "cgrates.org",
		Section: config.SCHEDULER_JSN,
	}, &rpl); err != nil {
		t.Error(err)
	} else if cfgStr != rpl {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl))
	}

	var replyPingAf string
	if err := testSectRPC.Call(context.Background(), utils.SchedulerSv1Ping, &utils.CGREvent{}, &replyPingAf); err != nil {
		t.Error(err)
	} else if replyPingAf != utils.Pong {
		t.Errorf("Expected OK received: %s", replyPingAf)
	}
}

func testSectConfigSReloadCaches(t *testing.T) {

	var replyPingBf string
	if err := testSectRPC.Call(context.Background(), utils.CacheSv1Ping, &utils.CGREvent{}, &replyPingBf); err != nil {
		t.Error(err)
	} else if replyPingBf != utils.Pong {
		t.Errorf("Expected OK received: %s", replyPingBf)
	}

	var reply string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1SetConfigFromJSON, &config.SetConfigFromJSONArgs{
		Tenant: "cgrates.org",
		Config: `{"caches":{"partitions":{"*account_action_plans":{"limit":0,"precache":false,"replicate":false,"static_ttl":false},"*action_plans":{"limit":0,"precache":false,"replicate":false,"static_ttl":false},"*action_triggers":{"limit":0,"precache":false,"replicate":false,"static_ttl":false},"*actions":{"limit":0,"precache":false,"replicate":false,"static_ttl":false},"*apiban":{"limit":0,"precache":false,"replicate":false,"static_ttl":false,"ttl":"2m0s"},"*attribute_filter_indexes":{"limit":0,"precache":false,"replicate":false,"static_ttl":false},"*attribute_profiles":{"limit":0,"precache":false,"replicate":false,"static_ttl":false},"*caps_events":{"limit":0,"precache":false,"replicate":false,"static_ttl":false},"*cdr_ids":{"limit":0,"precache":false,"replicate":false,"static_ttl":false,"ttl":"10m0s"},"*charger_filter_indexes":{"limit":0,"precache":false,"replicate":false,"static_ttl":false},"*charger_profiles":{"limit":0,"precache":false,"replicate":false,"static_ttl":false},"*closed_sessions":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":"10s"},"*destinations":{"limit":0,"precache":false,"replicate":false,"static_ttl":false},"*diameter_messages":{"limit":0,"precache":false,"replicate":false,"static_ttl":false,"ttl":"3h0m0s"},"*dispatcher_filter_indexes":{"limit":0,"precache":false,"replicate":false,"static_ttl":false},"*dispatcher_hosts":{"limit":0,"precache":false,"replicate":false,"static_ttl":false},"*dispatcher_loads":{"limit":0,"precache":false,"replicate":false,"static_ttl":false},"*dispatcher_profiles":{"limit":0,"precache":false,"replicate":false,"static_ttl":false},"*dispatcher_routes":{"limit":0,"precache":false,"replicate":false,"static_ttl":false},"*dispatchers":{"limit":0,"precache":false,"replicate":false,"static_ttl":false},"*event_charges":{"limit":0,"precache":false,"replicate":false,"static_ttl":false,"ttl":"10s"},"*event_resources":{"limit":0,"precache":false,"replicate":false,"static_ttl":false},"*filters":{"limit":0,"precache":false,"replicate":false,"static_ttl":false},"*load_ids":{"limit":0,"precache":false,"replicate":false,"static_ttl":false},"*rating_plans":{"limit":0,"precache":false,"replicate":false,"static_ttl":false},"*rating_profiles":{"limit":0,"precache":false,"replicate":false,"static_ttl":false},"*replication_hosts":{"limit":0,"precache":false,"replicate":false,"static_ttl":false},"*resource_filter_indexes":{"limit":0,"precache":false,"replicate":false,"static_ttl":false},"*resource_profiles":{"limit":0,"precache":false,"replicate":false,"static_ttl":false},"*resources":{"limit":0,"precache":false,"replicate":false,"static_ttl":false},"*reverse_destinations":{"limit":0,"precache":false,"replicate":false,"static_ttl":false},"*reverse_filter_indexes":{"limit":0,"precache":false,"replicate":false,"static_ttl":false},"*route_filter_indexes":{"limit":0,"precache":false,"replicate":false,"static_ttl":false},"*route_profiles":{"limit":0,"precache":false,"replicate":false,"static_ttl":false},"*rpc_connections":{"limit":0,"precache":false,"replicate":false,"static_ttl":false},"*rpc_responses":{"limit":0,"precache":false,"replicate":false,"static_ttl":false,"ttl":"2s"},"*shared_groups":{"limit":0,"precache":false,"replicate":false,"static_ttl":false},"*stat_filter_indexes":{"limit":0,"precache":false,"replicate":false,"static_ttl":false},"*statqueue_profiles":{"limit":0,"precache":false,"replicate":false,"static_ttl":false},"*statqueues":{"limit":0,"precache":false,"replicate":false,"static_ttl":false},"*stir":{"limit":0,"precache":false,"replicate":false,"static_ttl":false,"ttl":"3h0m0s"},"*threshold_filter_indexes":{"limit":0,"precache":false,"replicate":false,"static_ttl":false},"*threshold_profiles":{"limit":0,"precache":false,"replicate":false,"static_ttl":false},"*thresholds":{"limit":0,"precache":false,"replicate":false,"static_ttl":false},"*timings":{"limit":0,"precache":false,"replicate":false,"static_ttl":false},"*tmp_rating_profiles":{"limit":0,"precache":false,"replicate":false,"static_ttl":false,"ttl":"1m0s"},"*uch":{"limit":0,"precache":false,"replicate":false,"static_ttl":false,"ttl":"3h0m0s"}},"replication_conns":[]}}"`,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := `{"caches":{"partitions":{"*account_action_plans":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*action_plans":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*action_triggers":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*actions":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*apiban":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"2m0s"},"*attribute_filter_indexes":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*attribute_profiles":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*caps_events":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*cdr_ids":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"10m0s"},"*charger_filter_indexes":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*charger_profiles":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*closed_sessions":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"10s"},"*destinations":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*diameter_messages":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"3h0m0s"},"*dispatcher_filter_indexes":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*dispatcher_hosts":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*dispatcher_loads":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*dispatcher_profiles":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*dispatcher_routes":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*dispatchers":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*event_charges":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"10s"},"*event_resources":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*filters":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*load_ids":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*radius_packets":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"3h0m0s"},"*ranking_profiles":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*rating_plans":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*rating_profiles":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*replication_hosts":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*resource_filter_indexes":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*resource_profiles":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*resources":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*reverse_destinations":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*reverse_filter_indexes":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*route_filter_indexes":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*route_profiles":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*rpc_connections":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*rpc_responses":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"2s"},"*sentrypeer":{"limit":-1,"precache":false,"remote":false,"replicate":false,"static_ttl":true,"ttl":"24h0m0s"},"*shared_groups":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*stat_filter_indexes":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*statqueue_profiles":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*statqueues":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*stir":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"3h0m0s"},"*threshold_filter_indexes":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*threshold_profiles":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*thresholds":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*timings":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false},"*tmp_rating_profiles":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"1m0s"},"*uch":{"limit":0,"precache":false,"remote":false,"replicate":false,"static_ttl":false,"ttl":"3h0m0s"}},"remote_conns":[],"replication_conns":[]}}`
	var rpl string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:  "cgrates.org",
		Section: config.CACHE_JSN,
	}, &rpl); err != nil {
		t.Error(err)
	} else if cfgStr != rpl {
		t.Errorf("\nExpected %+v ,\n received: %+v", cfgStr, rpl)
	}

	var replyPingAf string
	if err := testSectRPC.Call(context.Background(), utils.CacheSv1Ping, &utils.CGREvent{}, &replyPingAf); err != nil {
		t.Error(err)
	} else if replyPingAf != utils.Pong {
		t.Errorf("Expected OK received: %s", replyPingAf)
	}
}

func testSectConfigSReloadFilters(t *testing.T) {

	var reply string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1SetConfigFromJSON, &config.SetConfigFromJSONArgs{
		Tenant: "cgrates.org",
		Config: "{\"filters\":{\"apiers_conns\":[\"*internal\"],\"resources_conns\":[\"*internal\"],\"stats_conns\":[\"*localhost\"]}}",
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := "{\"filters\":{\"apiers_conns\":[\"*internal\"],\"resources_conns\":[\"*internal\"],\"stats_conns\":[\"*localhost\"]}}"
	var rpl string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:  "cgrates.org",
		Section: config.FilterSjsn,
	}, &rpl); err != nil {
		t.Error(err)
	} else if cfgStr != rpl {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl))
	}
}

func testSectConfigSReloadRALS(t *testing.T) {

	var reply string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1SetConfigFromJSON, &config.SetConfigFromJSONArgs{
		Tenant: "cgrates.org",
		Config: `{"rals":{"balance_rating_subject":{"*any":"*zero1ns","*voice":"*zero1s"},"enabled":true,"fallback_depth":3,"max_computed_usage":{"*any":"189h0m0s","*data":"107374182400","*mms":"10000","*sms":"10000","*voice":"72h0m0s"},"max_increments":3000000,"remove_expired":true,"rp_subject_prefix_matching":false,"sessions_conns":[],"stats_conns":[],"thresholds_conns":["*internal"]}}`,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := `{"rals":{"balance_rating_subject":{"*any":"*zero1ns","*voice":"*zero1s"},"enabled":true,"fallback_depth":3,"max_computed_usage":{"*any":"189h0m0s","*data":"107374182400","*mms":"10000","*sms":"10000","*voice":"72h0m0s"},"max_increments":3000000,"remove_expired":true,"rp_subject_prefix_matching":false,"sessions_conns":[],"stats_conns":[],"thresholds_conns":["*internal"]}}`
	var rpl string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:  "cgrates.org",
		Section: config.RALS_JSN,
	}, &rpl); err != nil {
		t.Error(err)
	} else if cfgStr != rpl {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl))
	}

	var replyPingAf string
	if err := testSectRPC.Call(context.Background(), utils.RALsV1Ping, &utils.CGREvent{}, &replyPingAf); err != nil {
		t.Error(err)
	} else if replyPingAf != utils.Pong {
		t.Errorf("Expected OK received: %s", replyPingAf)
	}
}

func testSectConfigSReloadCDRS(t *testing.T) {

	var replyPingBf string
	if err := testSectRPC.Call(context.Background(), utils.CDRsV1Ping, &utils.CGREvent{}, &replyPingBf); err != nil {
		t.Error(err)
	} else if replyPingBf != utils.Pong {
		t.Errorf("Expected OK received: %s", replyPingBf)
	}

	var reply string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1SetConfigFromJSON, &config.SetConfigFromJSONArgs{
		Tenant: "cgrates.org",
		Config: "{\"cdrs\":{\"attributes_conns\":[],\"chargers_conns\":[\"*internal\"],\"ees_conns\":[],\"enabled\":true,\"extra_fields\":[],\"online_cdr_exports\":[],\"rals_conns\":[\"*internal\"],\"scheduler_conns\":[],\"session_cost_retries\":5,\"stats_conns\":[],\"store_cdrs\":true,\"thresholds_conns\":[]}}",
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := "{\"cdrs\":{\"attributes_conns\":[],\"chargers_conns\":[\"*internal\"],\"ees_conns\":[],\"enabled\":true,\"extra_fields\":[],\"online_cdr_exports\":[],\"rals_conns\":[\"*internal\"],\"scheduler_conns\":[],\"session_cost_retries\":5,\"stats_conns\":[],\"store_cdrs\":true,\"thresholds_conns\":[]}}"
	var rpl string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:  "cgrates.org",
		Section: config.CDRS_JSN,
	}, &rpl); err != nil {
		t.Error(err)
	} else if cfgStr != rpl {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl))
	}

	var replyPingAf string
	if err := testSectRPC.Call(context.Background(), utils.CDRsV1Ping, &utils.CGREvent{}, &replyPingAf); err != nil {
		t.Error(err)
	} else if replyPingAf != utils.Pong {
		t.Errorf("Expected OK received: %s", replyPingAf)
	}
}

func testSectConfigSReloadERS(t *testing.T) {

	var reply string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1SetConfigFromJSON, &config.SetConfigFromJSONArgs{
		Tenant: "cgrates.org",
		Config: `{"ers":{"concurrent_events":1,"ees_conns":[],"enabled":true,"partial_cache_ttl":"1s","readers":[{"cache_dump_fields":[],"concurrent_requests":1024,"fields":[{"mandatory":true,"path":"*cgreq.ToR","tag":"ToR","type":"*variable","value":"~*req.2"},{"mandatory":true,"path":"*cgreq.OriginID","tag":"OriginID","type":"*variable","value":"~*req.3"},{"mandatory":true,"path":"*cgreq.RequestType","tag":"RequestType","type":"*variable","value":"~*req.4"},{"mandatory":true,"path":"*cgreq.Tenant","tag":"Tenant","type":"*variable","value":"~*req.6"},{"mandatory":true,"path":"*cgreq.Category","tag":"Category","type":"*variable","value":"~*req.7"},{"mandatory":true,"path":"*cgreq.Account","tag":"Account","type":"*variable","value":"~*req.8"},{"mandatory":true,"path":"*cgreq.Subject","tag":"Subject","type":"*variable","value":"~*req.9"},{"mandatory":true,"path":"*cgreq.Destination","tag":"Destination","type":"*variable","value":"~*req.10"},{"mandatory":true,"path":"*cgreq.SetupTime","tag":"SetupTime","type":"*variable","value":"~*req.11"},{"mandatory":true,"path":"*cgreq.AnswerTime","tag":"AnswerTime","type":"*variable","value":"~*req.12"},{"mandatory":true,"path":"*cgreq.Usage","tag":"Usage","type":"*variable","value":"~*req.13"}],"filters":[],"flags":[],"id":"*default","max_reconnect_interval":"5m0s","opts":{"csvFieldSeparator":",","csvHeaderDefineChar":":","csvRowLength":0,"natsSubject":"cgrates_cdrs","partialCacheAction":"*none","partialOrderField":"~*req.AnswerTime","xmlRootPath":""},"partial_commit_fields":[],"processed_path":"/var/spool/cgrates/ers/out","reconnects":-1,"run_delay":"0","source_path":"/var/spool/cgrates/ers/in","tenant":"","timezone":"","type":"*none"}],"sessions_conns":["*internal"]}}`,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := `{"ers":{"concurrent_events":1,"ees_conns":[],"enabled":true,"partial_cache_ttl":"1s","readers":[{"cache_dump_fields":[],"concurrent_requests":1024,"fields":[{"mandatory":true,"path":"*cgreq.ToR","tag":"ToR","type":"*variable","value":"~*req.2"},{"mandatory":true,"path":"*cgreq.OriginID","tag":"OriginID","type":"*variable","value":"~*req.3"},{"mandatory":true,"path":"*cgreq.RequestType","tag":"RequestType","type":"*variable","value":"~*req.4"},{"mandatory":true,"path":"*cgreq.Tenant","tag":"Tenant","type":"*variable","value":"~*req.6"},{"mandatory":true,"path":"*cgreq.Category","tag":"Category","type":"*variable","value":"~*req.7"},{"mandatory":true,"path":"*cgreq.Account","tag":"Account","type":"*variable","value":"~*req.8"},{"mandatory":true,"path":"*cgreq.Subject","tag":"Subject","type":"*variable","value":"~*req.9"},{"mandatory":true,"path":"*cgreq.Destination","tag":"Destination","type":"*variable","value":"~*req.10"},{"mandatory":true,"path":"*cgreq.SetupTime","tag":"SetupTime","type":"*variable","value":"~*req.11"},{"mandatory":true,"path":"*cgreq.AnswerTime","tag":"AnswerTime","type":"*variable","value":"~*req.12"},{"mandatory":true,"path":"*cgreq.Usage","tag":"Usage","type":"*variable","value":"~*req.13"}],"filters":[],"flags":[],"id":"*default","max_reconnect_interval":"5m0s","opts":{"csvFieldSeparator":",","csvHeaderDefineChar":":","csvRowLength":0,"natsSubject":"cgrates_cdrs","partialCacheAction":"*none","partialOrderField":"~*req.AnswerTime","xmlRootPath":""},"partial_commit_fields":[],"processed_path":"/var/spool/cgrates/ers/out","reconnects":-1,"run_delay":"0","source_path":"/var/spool/cgrates/ers/in","tenant":"","timezone":"","type":"*none"}],"sessions_conns":["*internal"]}}`
	var rpl string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:  "cgrates.org",
		Section: "ers",
	}, &rpl); err != nil {
		t.Error(err)
	} else if cfgStr != rpl {
		t.Errorf("\nExpected %+v ,\n received: %+v", cfgStr, rpl)
	}
}

func testSectConfigSReloadEES(t *testing.T) {

	var reply string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1SetConfigFromJSON, &config.SetConfigFromJSONArgs{
		Tenant: "cgrates.org",
		Config: "{\"ees\":{\"attributes_conns\":[],\"cache\":{\"*file_csv\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"5s\"}},\"enabled\":true,\"exporters\":[{\"attempts\":1,\"attribute_context\":\"\",\"attribute_ids\":[],\"concurrent_requests\":0,\"export_path\":\"/var/spool/cgrates/ees\",\"fields\":[],\"filters\":[],\"flags\":[],\"id\":\"*default\",\"opts\":{},\"synchronous\":false,\"timezone\":\"\",\"type\":\"*none\"}]}}",
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := "{\"ees\":{\"attributes_conns\":[],\"cache\":{\"*file_csv\":{\"limit\":-1,\"precache\":false,\"remote\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"5s\"}},\"enabled\":true,\"exporters\":[{\"attempts\":1,\"attribute_context\":\"\",\"attribute_ids\":[],\"concurrent_requests\":0,\"export_path\":\"/var/spool/cgrates/ees\",\"failed_posts_dir\":\"/var/spool/cgrates/failed_posts\",\"fields\":[],\"filters\":[],\"flags\":[],\"id\":\"*default\",\"opts\":{},\"synchronous\":false,\"timezone\":\"\",\"type\":\"*none\"}]}}"
	var rpl string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:  "cgrates.org",
		Section: "ees",
	}, &rpl); err != nil {
		t.Error(err)
	} else if cfgStr != rpl {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl))
	}
}

func testSectConfigSReloadSessions(t *testing.T) {

	var reply string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1SetConfigFromJSON, &config.SetConfigFromJSONArgs{
		Tenant: "cgrates.org",
		Config: "{\"sessions\":{\"alterable_fields\":[],\"attributes_conns\":[\"*internal\"],\"cdrs_conns\":[\"*internal\"],\"channel_sync_interval\":\"0\",\"chargers_conns\":[\"*internal\"],\"client_protocol\":2,\"debit_interval\":\"0\",\"default_usage\":{\"*any\":\"3h0m0s\",\"*data\":\"1048576\",\"*sms\":\"1\",\"*voice\":\"3h0m0s\"},\"enabled\":true,\"listen_bigob\":\"\",\"listen_bijson\":\"127.0.0.1:2014\",\"min_dur_low_balance\":\"0\",\"rals_conns\":[\"*internal\"],\"replication_conns\":[],\"resources_conns\":[\"*internal\"],\"routes_conns\":[\"*internal\"],\"scheduler_conns\":[],\"session_indexes\":[\"OriginID\"],\"session_ttl\":\"0\",\"stats_conns\":[],\"stir\":{\"allowed_attest\":[\"*any\"],\"default_attest\":\"A\",\"payload_maxduration\":\"-1\",\"privatekey_path\":\"\",\"publickey_path\":\"\"},\"store_session_costs\":false,\"terminate_attempts\":5,\"thresholds_conns\":[]}}",
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := "{\"sessions\":{\"alterable_fields\":[],\"attributes_conns\":[\"*internal\"],\"backup_interval\":\"0\",\"cdrs_conns\":[\"*internal\"],\"channel_sync_interval\":\"0\",\"chargers_conns\":[\"*internal\"],\"client_protocol\":2,\"debit_interval\":\"0\",\"default_usage\":{\"*any\":\"3h0m0s\",\"*data\":\"1048576\",\"*sms\":\"1\",\"*voice\":\"3h0m0s\"},\"enabled\":true,\"listen_bigob\":\"\",\"listen_bijson\":\"127.0.0.1:2014\",\"min_dur_low_balance\":\"0\",\"rals_conns\":[\"*internal\"],\"replication_conns\":[],\"resources_conns\":[\"*internal\"],\"routes_conns\":[\"*internal\"],\"scheduler_conns\":[],\"session_indexes\":[\"OriginID\"],\"session_ttl\":\"0\",\"stale_chan_max_extra_usage\":\"0\",\"stats_conns\":[],\"stir\":{\"allowed_attest\":[\"*any\"],\"default_attest\":\"A\",\"payload_maxduration\":\"-1\",\"privatekey_path\":\"\",\"publickey_path\":\"\"},\"store_session_costs\":false,\"terminate_attempts\":5,\"thresholds_conns\":[]}}"
	var rpl string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:  "cgrates.org",
		Section: config.SessionSJson,
	}, &rpl); err != nil {
		t.Error(err)
	} else if cfgStr != rpl {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl))
	}

	var replyPingAf string
	if err := testSectRPC.Call(context.Background(), utils.SessionSv1Ping, &utils.CGREvent{}, &replyPingAf); err != nil {
		t.Error(err)
	} else if replyPingAf != utils.Pong {
		t.Errorf("Expected OK received: %s", replyPingAf)
	}
}

func testSectConfigSReloadAsteriskAgent(t *testing.T) {

	var reply string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1SetConfigFromJSON, &config.SetConfigFromJSONArgs{
		Tenant: "cgrates.org",
		Config: `{"asterisk_agent": {
			"enabled": true,
			"sessions_conns": ["*birpc_internal"],
			"create_cdr": true,
			"asterisk_conns":[
				{"address": "127.0.0.1:8088", "user": "cgrates", "password": "CGRateS.org", "connect_attempts": 3,"reconnects": 5}
			],
		},
		}}`,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := "{\"asterisk_agent\":{\"asterisk_conns\":[{\"address\":\"127.0.0.1:8088\",\"alias\":\"\",\"connect_attempts\":3,\"max_reconnect_interval\":\"0s\",\"password\":\"CGRateS.org\",\"reconnects\":5,\"user\":\"cgrates\"}],\"create_cdr\":true,\"enabled\":true,\"sessions_conns\":[\"*birpc_internal\"]}}"
	var rpl string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:  "cgrates.org",
		Section: config.AsteriskAgentJSN,
	}, &rpl); err != nil {
		t.Error(err)
	} else if cfgStr != rpl {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl))
	}
}

func testSectConfigSReloadFreeswitchAgent(t *testing.T) {

	var reply string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1SetConfigFromJSON, &config.SetConfigFromJSONArgs{
		Tenant: "cgrates.org",
		Config: `{"freeswitch_agent": {
			"enabled": true,
			"sessions_conns": ["*birpc_internal"],
			"subscribe_park": true,
			"create_cdr": true,
			"extra_fields": ["extra_fields"],
			"low_balance_ann_file": "low_balance_ann_file",
			"empty_balance_context": "empty_balance_context",
			"empty_balance_ann_file": "empty_balance_ann_file",
			"max_wait_connection": "2s",
			"event_socket_conns":[
				{"address": "127.0.0.1:8021", "password": "ClueCon", "reconnects": 5,"alias":"alias"}
			],
		},
		}}`,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := "{\"freeswitch_agent\":{\"create_cdr\":true,\"empty_balance_ann_file\":\"empty_balance_ann_file\",\"empty_balance_context\":\"empty_balance_context\",\"enabled\":true,\"event_socket_conns\":[{\"address\":\"127.0.0.1:8021\",\"alias\":\"alias\",\"password\":\"ClueCon\",\"reconnects\":5}],\"extra_fields\":\"extra_fields\",\"low_balance_ann_file\":\"low_balance_ann_file\",\"max_wait_connection\":\"2s\",\"sessions_conns\":[\"*birpc_internal\"],\"subscribe_park\":true}}"
	var rpl string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:  "cgrates.org",
		Section: config.FreeSWITCHAgentJSN,
	}, &rpl); err != nil {
		t.Error(err)
	} else if cfgStr != rpl {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl))
	}

	var reply2 string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1SetConfigFromJSON, &config.SetConfigFromJSONArgs{
		Tenant: "cgrates.org",
		Config: `{"freeswitch_agent": {
			"enabled": false,
			"sessions_conns": [],
			"subscribe_park": false,
			"create_cdr": false,
			"extra_fields": [],
			"low_balance_ann_file": "",
			"empty_balance_context": "",
			"empty_balance_ann_file": "",
			"max_wait_connection": "",
			"event_socket_conns":[
				{"address": "", "password": "", "reconnects": 0,"alias":""}
			],
		},
		}}`,
	}, &reply2); err != nil {
		t.Error(err)
	} else if reply2 != utils.OK {
		t.Errorf("Expected OK received: %+v", reply2)
	}
}

func testSectConfigSReloadKamailioAgent(t *testing.T) {

	var reply string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1SetConfigFromJSON, &config.SetConfigFromJSONArgs{
		Tenant: "cgrates.org",
		Config: `{"kamailio_agent": {
			"enabled": true,
			"sessions_conns": ["*birpc_internal"],
			"create_cdr": true,
			"timezone": "local",
			"evapi_conns":[
				{"address": "127.0.0.1:8448", "reconnects": 5}
			],
		},
		}}`,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := "{\"kamailio_agent\":{\"create_cdr\":true,\"enabled\":true,\"evapi_conns\":[{\"address\":\"127.0.0.1:8448\",\"alias\":\"\",\"reconnects\":5}],\"sessions_conns\":[\"*birpc_internal\"],\"timezone\":\"local\"}}"
	var rpl string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:  "cgrates.org",
		Section: config.KamailioAgentJSN,
	}, &rpl); err != nil {
		t.Error(err)
	} else if cfgStr != rpl {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl))
	}
	var reply2 string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1SetConfigFromJSON, &config.SetConfigFromJSONArgs{
		Tenant: "cgrates.org",
		Config: `{"kamailio_agent": {
			"enabled": false,
			"sessions_conns": [],
			"create_cdr": false,
			"timezone": "",
			"evapi_conns":[
				{"address": "", "reconnects": 0}
			],
		},
		}}`,
	}, &reply2); err != nil {
		t.Error(err)
	} else if reply2 != utils.OK {
		t.Errorf("Expected OK received: %+v", reply2)
	}
}

func testSectConfigSReloadDiameterAgent(t *testing.T) {

	var reply string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1SetConfigFromJSON, &config.SetConfigFromJSONArgs{
		Tenant: "cgrates.org",
		Config: `{"diameter_agent": {
			"enabled": true,
			"listen": "127.0.0.1:3868",
			"listen_net": "tcp",
			"dictionaries_path": "/usr/share/cgrates/diameter/dict/",
			"sessions_conns": ["*birpc_internal"],
			"origin_host": "CGR-DA",
			"origin_realm": "cgrates.org",
			"vendor_id": 1,
			"product_name": "CGRateS",
			"synced_conn_requests": false,
			"asr_template": "asr_template",
			"rar_template": "rar_template",
			"forced_disconnect": "*none",
			"request_processors": [
				{
					"id": "cgrates",
					"tenant": "1",
					"filters": [],
					 "flags": ["1"],
					"request_fields": [
					   {"path": "randomPath"},
					  ],
					"reply_fields": [
						 {"path": "randomPath"},
					 ],
				   }
			],
		},
		}}`,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := `{"diameter_agent":{"asr_template":"asr_template","dictionaries_path":"/usr/share/cgrates/diameter/dict/","enabled":true,"forced_disconnect":"*none","listen":"127.0.0.1:3868","listen_net":"tcp","origin_host":"CGR-DA","origin_realm":"cgrates.org","product_name":"CGRateS","rar_template":"rar_template","request_processors":[{"filters":[],"flags":["1"],"id":"cgrates","reply_fields":[{"path":"randomPath","tag":"randomPath"}],"request_fields":[{"path":"randomPath","tag":"randomPath"}],"tenant":"1","timezone":""}],"sessions_conns":["*birpc_internal"],"synced_conn_requests":false,"vendor_id":1}}`
	var rpl string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:  "cgrates.org",
		Section: "diameter_agent",
	}, &rpl); err != nil {
		t.Error(err)
	} else if cfgStr != rpl {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl))
	}
}

func testSectConfigSReloadRadiusAgent(t *testing.T) {

	var reply string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1SetConfigFromJSON, &config.SetConfigFromJSONArgs{
		Tenant: "cgrates.org",
		Config: `{"radius_agent": {
			"enabled": true,
			"listeners":[
				{
					"network": "udp",
					"auth_address": "127.0.0.1:1812",
					"acct_address": "127.0.0.1:1813"
				}
			],	
			"client_secrets": {
				"*default": "CGRateS.org"
			},
			"client_dictionaries": {
				"*default": [
				"/usr/share/cgrates/radius/dict/",
				],
			},
			"sessions_conns": ["*internal"],
			"request_processors": [
				{
					"id": "cgrates",
					"tenant": "1",
					"filters": [],
					 "flags": ["1"],
					"request_fields": [
					   {"path": "randomPath"},
					  ],
					"reply_fields": [
						 {"path": "randomPath"},
					 ],
				   }
			],
		},
		}}`,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := `{"radius_agent":{"client_dictionaries":{"*default":["/usr/share/cgrates/radius/dict/"]},"client_secrets":{"*default":"CGRateS.org"},"coa_template":"*coa","dmr_template":"*dmr","enabled":true,"listeners":[{"acct_address":"127.0.0.1:1813","auth_address":"127.0.0.1:1812","network":"udp"}],"request_processors":[{"filters":[],"flags":["1"],"id":"cgrates","reply_fields":[{"path":"randomPath","tag":"randomPath"}],"request_fields":[{"path":"randomPath","tag":"randomPath"}],"tenant":"1","timezone":""}],"requests_cache_key":"","sessions_conns":["*internal"]}}`
	var rpl string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:  "cgrates.org",
		Section: "radius_agent",
	}, &rpl); err != nil {
		t.Error(err)
	} else if cfgStr != rpl {
		t.Errorf("expected %s,\nreceived: %s", cfgStr, rpl)
	}
}

func testSectConfigSReloadHTTPAgent(t *testing.T) {

	var reply string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1SetConfigFromJSON, &config.SetConfigFromJSONArgs{
		Tenant: "cgrates.org",
		Config: `{"http_agent": [
			{
				"id": "conecto1",
				"url": "/conecto",
				"sessions_conns": ["*internal"],
				"request_payload":	"*url",
				"reply_payload":	"*xml",
				"request_processors": [],
			}
		],
		}}`,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := "{\"http_agent\":[{\"id\":\"conecto1\",\"reply_payload\":\"*xml\",\"request_payload\":\"*url\",\"request_processors\":[],\"sessions_conns\":[\"*internal\"],\"url\":\"/conecto\"}]}"
	var rpl string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:  "cgrates.org",
		Section: config.HttpAgentJson,
	}, &rpl); err != nil {
		t.Error(err)
	} else if cfgStr != rpl {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl))
	}
}

func testSectConfigSReloadDNSAgent(t *testing.T) {

	var reply string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1SetConfigFromJSON, &config.SetConfigFromJSONArgs{
		Tenant: "cgrates.org",
		Config: `{"dns_agent": {
			"enabled": true,
			"listeners":[
				{
					"address": "127.0.0.1:2053",							
					"network": "udp"									
				}
			],
			"sessions_conns": ["*internal"],
			"timezone": "local",
			"request_processors": [
				{
					"id": "random",
				},
			],
		},
		}}`,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := "{\"dns_agent\":{\"enabled\":true,\"listeners\":[{\"address\":\"127.0.0.1:2053\",\"network\":\"udp\"}],\"request_processors\":[{\"filters\":null,\"flags\":null,\"id\":\"random\",\"timezone\":\"\"}],\"sessions_conns\":[\"*internal\"],\"timezone\":\"local\"}}"
	var rpl string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:  "cgrates.org",
		Section: config.DNSAgentJson,
	}, &rpl); err != nil {
		t.Error(err)
	} else if cfgStr != rpl {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl))
	}
}

func testSectConfigSReloadAttributes(t *testing.T) {

	var reply string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1SetConfigFromJSON, &config.SetConfigFromJSONArgs{
		Tenant: "cgrates.org",
		Config: `{"attributes": {
			"enabled": true,
			"stats_conns": ["*internal"],
			"resources_conns": ["*internal"],
			"apiers_conns": ["*internal"],
			"indexed_selects": true,
			"string_indexed_fields": ["string_indexed_fields"],
			"prefix_indexed_fields": ["prefix_indexed_fields"],
			"suffix_indexed_fields": ["suffix_indexed_fields"],
			"nested_fields": true,
			"process_runs": 1,
			"any_context": true
		},}}`,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := "{\"attributes\":{\"any_context\":true,\"apiers_conns\":[\"*internal\"],\"enabled\":true,\"indexed_selects\":true,\"nested_fields\":true,\"opts\":{\"*processRuns\":1,\"*profileIDs\":[],\"*profileIgnoreFilters\":false,\"*profileRuns\":0},\"prefix_indexed_fields\":[\"prefix_indexed_fields\"],\"resources_conns\":[\"*internal\"],\"stats_conns\":[\"*internal\"],\"string_indexed_fields\":[\"string_indexed_fields\"],\"suffix_indexed_fields\":[\"suffix_indexed_fields\"]}}"

	var rpl string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:  "cgrates.org",
		Section: config.ATTRIBUTE_JSN,
	}, &rpl); err != nil {
		t.Error(err)
	} else if cfgStr != rpl {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl))
	}

	var replyPingAf string
	if err := testSectRPC.Call(context.Background(), utils.AttributeSv1Ping, &utils.CGREvent{}, &replyPingAf); err != nil {
		t.Error(err)
	} else if replyPingAf != utils.Pong {
		t.Errorf("Expected OK received: %s", replyPingAf)
	}

}

func testSectConfigSReloadChargers(t *testing.T) {

	var reply string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1SetConfigFromJSON, &config.SetConfigFromJSONArgs{
		Tenant: "cgrates.org",
		Config: `{"chargers": {
			"enabled": true,
			"attributes_conns": ["*internal"],
			"indexed_selects": true,
			"string_indexed_fields": ["string_indexed_fields"],
			"prefix_indexed_fields": ["prefix_indexed_fields"],
			"suffix_indexed_fields": ["suffix_indexed_fields"],
			"nested_fields": true,
		},}}`,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := "{\"chargers\":{\"attributes_conns\":[\"*internal\"],\"enabled\":true,\"indexed_selects\":true,\"nested_fields\":true,\"prefix_indexed_fields\":[\"prefix_indexed_fields\"],\"string_indexed_fields\":[\"string_indexed_fields\"],\"suffix_indexed_fields\":[\"suffix_indexed_fields\"]}}"
	var rpl string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:  "cgrates.org",
		Section: config.ChargerSCfgJson,
	}, &rpl); err != nil {
		t.Error(err)
	} else if cfgStr != rpl {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl))
	}

	var replyPingAf string
	if err := testSectRPC.Call(context.Background(), utils.ChargerSv1Ping, &utils.CGREvent{}, &replyPingAf); err != nil {
		t.Error(err)
	} else if replyPingAf != utils.Pong {
		t.Errorf("Expected OK received: %s", replyPingAf)
	}
}

func testSectConfigSReloadResources(t *testing.T) {

	var reply string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1SetConfigFromJSON, &config.SetConfigFromJSONArgs{
		Tenant: "cgrates.org",
		Config: `{"resources": {
			"enabled": true,
			"store_interval": "-1",
			"thresholds_conns": ["*internal"],
			"indexed_selects": true,
			"string_indexed_fields": ["string_indexed_fields"],
			"prefix_indexed_fields": ["prefix_indexed_fields"],
			"suffix_indexed_fields": ["suffix_indexed_fields"],
			"nested_fields": true,
		},}}`,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := "{\"resources\":{\"enabled\":true,\"indexed_selects\":true,\"nested_fields\":true,\"opts\":{\"*units\":1,\"*usageID\":\"\"},\"prefix_indexed_fields\":[\"prefix_indexed_fields\"],\"store_interval\":\"-1ns\",\"string_indexed_fields\":[\"string_indexed_fields\"],\"suffix_indexed_fields\":[\"suffix_indexed_fields\"],\"thresholds_conns\":[\"*internal\"]}}"
	var rpl string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:  "cgrates.org",
		Section: config.RESOURCES_JSON,
	}, &rpl); err != nil {
		t.Error(err)
	} else if cfgStr != rpl {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl))
	}

	var replyPingAf string
	if err := testSectRPC.Call(context.Background(), utils.ResourceSv1Ping, &utils.CGREvent{}, &replyPingAf); err != nil {
		t.Error(err)
	} else if replyPingAf != utils.Pong {
		t.Errorf("Expected OK received: %s", replyPingAf)
	}
}

func testSectConfigSReloadStats(t *testing.T) {

	var reply string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1SetConfigFromJSON, &config.SetConfigFromJSONArgs{
		Tenant: "cgrates.org",
		Config: `{"stats": {
			"enabled": true,
			"store_interval": "-1",
			"store_uncompressed_limit": 1,
			"thresholds_conns": ["*internal"],
			"indexed_selects": true,
			"string_indexed_fields": ["string_indexed_fields"],
			"prefix_indexed_fields": ["prefix_indexed_fields"],
			"suffix_indexed_fields": ["suffix_indexed_fields"],
			"nested_fields": true,
		},
		}}`,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := "{\"stats\":{\"enabled\":true,\"indexed_selects\":true,\"nested_fields\":true,\"opts\":{\"*profileIDs\":[],\"*profileIgnoreFilters\":false},\"prefix_indexed_fields\":[\"prefix_indexed_fields\"],\"store_interval\":\"-1ns\",\"store_uncompressed_limit\":1,\"string_indexed_fields\":[\"string_indexed_fields\"],\"suffix_indexed_fields\":[\"suffix_indexed_fields\"],\"thresholds_conns\":[\"*internal\"]}}"
	var rpl string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:  "cgrates.org",
		Section: config.STATS_JSON,
	}, &rpl); err != nil {
		t.Error(err)
	} else if cfgStr != rpl {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl))
	}
	var replyPingAf string
	if err := testSectRPC.Call(context.Background(), utils.StatSv1Ping, &utils.CGREvent{}, &replyPingAf); err != nil {
		t.Error(err)
	} else if replyPingAf != utils.Pong {
		t.Errorf("Expected OK received: %s", replyPingAf)
	}

}

func testSectConfigSReloadThresholds(t *testing.T) {

	var reply string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1SetConfigFromJSON, &config.SetConfigFromJSONArgs{
		Tenant: "cgrates.org",
		Config: "{\"thresholds\":{\"enabled\":true,\"indexed_selects\":true,\"nested_fields\":true,\"prefix_indexed_fields\":[\"prefix_indexed_fields\"],\"store_interval\":\"-1\",\"string_indexed_fields\":[\"string_indexed_fields\"],\"suffix_indexed_fields\":[\"suffix_indexed_fields\"]}}",
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := "{\"thresholds\":{\"enabled\":true,\"indexed_selects\":true,\"nested_fields\":true,\"opts\":{\"*profileIDs\":[],\"*profileIgnoreFilters\":false},\"prefix_indexed_fields\":[\"prefix_indexed_fields\"],\"store_interval\":\"-1ns\",\"string_indexed_fields\":[\"string_indexed_fields\"],\"suffix_indexed_fields\":[\"suffix_indexed_fields\"]}}"
	var rpl string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:  "cgrates.org",
		Section: config.THRESHOLDS_JSON,
	}, &rpl); err != nil {
		t.Error(err)
	} else if cfgStr != rpl {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl))
	}

	var replyPingAf string
	if err := testSectRPC.Call(context.Background(), utils.ThresholdSv1Ping, &utils.CGREvent{}, &replyPingAf); err != nil {
		t.Error(err)
	} else if replyPingAf != utils.Pong {
		t.Errorf("Expected OK received: %s", replyPingAf)
	}
}

func testSectConfigSReloadRoutes(t *testing.T) {

	var reply string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1SetConfigFromJSON, &config.SetConfigFromJSONArgs{
		Tenant: "cgrates.org",
		Config: `{"routes": {
					"enabled": true,
					"indexed_selects": true,
					"string_indexed_fields": ["string_indexed_fields"],
					"prefix_indexed_fields": ["prefix_indexed_fields"],
					"suffix_indexed_fields": ["suffix_indexed_fields"],
					"nested_fields": true,
					"attributes_conns": ["*localhost"],
					"resources_conns": ["*localhost"],
					"stats_conns": ["*localhost"],
					"rals_conns": ["*localhost"],
					"default_ratio":1
		},
		}}`,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := "{\"routes\":{\"attributes_conns\":[\"*localhost\"],\"default_ratio\":1,\"enabled\":true,\"indexed_selects\":true,\"nested_fields\":true,\"opts\":{\"*context\":\"*routes\",\"*ignoreErrors\":false,\"*maxCost\":\"\"},\"prefix_indexed_fields\":[\"prefix_indexed_fields\"],\"rals_conns\":[\"*localhost\"],\"resources_conns\":[\"*localhost\"],\"stats_conns\":[\"*localhost\"],\"string_indexed_fields\":[\"string_indexed_fields\"],\"suffix_indexed_fields\":[\"suffix_indexed_fields\"]}}"
	var rpl string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:  "cgrates.org",
		Section: config.RouteSJson,
	}, &rpl); err != nil {
		t.Error(err)
	} else if cfgStr != rpl {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl))
	}

	var replyPingAf string
	if err := testSectRPC.Call(context.Background(), utils.RouteSv1Ping, &utils.CGREvent{}, &replyPingAf); err != nil {
		t.Error(err)
	} else if replyPingAf != utils.Pong {
		t.Errorf("Expected OK received: %s", replyPingAf)
	}
}

func testSectConfigSReloadLoaders(t *testing.T) {

	var reply string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1ReloadConfig, &config.ReloadArgs{
		Tenant:  "cgrates.org",
		Path:    path.Join(*utils.DataDir, "conf", "samples", "tutinternal"),
		Section: config.LoaderJson,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := "{\"loaders\":[{\"caches_conns\":[\"*internal\"],\"data\":[{\"fields\":[{\"mandatory\":true,\"path\":\"Tenant\",\"tag\":\"TenantID\",\"type\":\"*variable\",\"value\":\"~*req.0\"},{\"mandatory\":true,\"path\":\"ID\",\"tag\":\"ProfileID\",\"type\":\"*variable\",\"value\":\"~*req.1\"},{\"path\":\"Contexts\",\"tag\":\"Contexts\",\"type\":\"*variable\",\"value\":\"~*req.2\"},{\"path\":\"FilterIDs\",\"tag\":\"FilterIDs\",\"type\":\"*variable\",\"value\":\"~*req.3\"},{\"path\":\"ActivationInterval\",\"tag\":\"ActivationInterval\",\"type\":\"*variable\",\"value\":\"~*req.4\"},{\"path\":\"AttributeFilterIDs\",\"tag\":\"AttributeFilterIDs\",\"type\":\"*variable\",\"value\":\"~*req.5\"},{\"path\":\"Path\",\"tag\":\"Path\",\"type\":\"*variable\",\"value\":\"~*req.6\"},{\"path\":\"Type\",\"tag\":\"Type\",\"type\":\"*variable\",\"value\":\"~*req.7\"},{\"path\":\"Value\",\"tag\":\"Value\",\"type\":\"*variable\",\"value\":\"~*req.8\"},{\"path\":\"Blocker\",\"tag\":\"Blocker\",\"type\":\"*variable\",\"value\":\"~*req.9\"},{\"path\":\"Weight\",\"tag\":\"Weight\",\"type\":\"*variable\",\"value\":\"~*req.10\"}],\"file_name\":\"Attributes.csv\",\"flags\":null,\"type\":\"*attributes\"},{\"fields\":[{\"mandatory\":true,\"path\":\"Tenant\",\"tag\":\"Tenant\",\"type\":\"*variable\",\"value\":\"~*req.0\"},{\"mandatory\":true,\"path\":\"ID\",\"tag\":\"ID\",\"type\":\"*variable\",\"value\":\"~*req.1\"},{\"path\":\"Type\",\"tag\":\"Type\",\"type\":\"*variable\",\"value\":\"~*req.2\"},{\"path\":\"Element\",\"tag\":\"Element\",\"type\":\"*variable\",\"value\":\"~*req.3\"},{\"path\":\"Values\",\"tag\":\"Values\",\"type\":\"*variable\",\"value\":\"~*req.4\"},{\"path\":\"ActivationInterval\",\"tag\":\"ActivationInterval\",\"type\":\"*variable\",\"value\":\"~*req.5\"}],\"file_name\":\"Filters.csv\",\"flags\":null,\"type\":\"*filters\"},{\"fields\":[{\"mandatory\":true,\"path\":\"Tenant\",\"tag\":\"Tenant\",\"type\":\"*variable\",\"value\":\"~*req.0\"},{\"mandatory\":true,\"path\":\"ID\",\"tag\":\"ID\",\"type\":\"*variable\",\"value\":\"~*req.1\"},{\"path\":\"FilterIDs\",\"tag\":\"FilterIDs\",\"type\":\"*variable\",\"value\":\"~*req.2\"},{\"path\":\"ActivationInterval\",\"tag\":\"ActivationInterval\",\"type\":\"*variable\",\"value\":\"~*req.3\"},{\"path\":\"UsageTTL\",\"tag\":\"TTL\",\"type\":\"*variable\",\"value\":\"~*req.4\"},{\"path\":\"Limit\",\"tag\":\"Limit\",\"type\":\"*variable\",\"value\":\"~*req.5\"},{\"path\":\"AllocationMessage\",\"tag\":\"AllocationMessage\",\"type\":\"*variable\",\"value\":\"~*req.6\"},{\"path\":\"Blocker\",\"tag\":\"Blocker\",\"type\":\"*variable\",\"value\":\"~*req.7\"},{\"path\":\"Stored\",\"tag\":\"Stored\",\"type\":\"*variable\",\"value\":\"~*req.8\"},{\"path\":\"Weight\",\"tag\":\"Weight\",\"type\":\"*variable\",\"value\":\"~*req.9\"},{\"path\":\"ThresholdIDs\",\"tag\":\"ThresholdIDs\",\"type\":\"*variable\",\"value\":\"~*req.10\"}],\"file_name\":\"Resources.csv\",\"flags\":null,\"type\":\"*resources\"},{\"fields\":[{\"mandatory\":true,\"path\":\"Tenant\",\"tag\":\"Tenant\",\"type\":\"*variable\",\"value\":\"~*req.0\"},{\"mandatory\":true,\"path\":\"ID\",\"tag\":\"ID\",\"type\":\"*variable\",\"value\":\"~*req.1\"},{\"path\":\"FilterIDs\",\"tag\":\"FilterIDs\",\"type\":\"*variable\",\"value\":\"~*req.2\"},{\"path\":\"ActivationInterval\",\"tag\":\"ActivationInterval\",\"type\":\"*variable\",\"value\":\"~*req.3\"},{\"path\":\"QueueLength\",\"tag\":\"QueueLength\",\"type\":\"*variable\",\"value\":\"~*req.4\"},{\"path\":\"TTL\",\"tag\":\"TTL\",\"type\":\"*variable\",\"value\":\"~*req.5\"},{\"path\":\"MinItems\",\"tag\":\"MinItems\",\"type\":\"*variable\",\"value\":\"~*req.6\"},{\"path\":\"MetricIDs\",\"tag\":\"MetricIDs\",\"type\":\"*variable\",\"value\":\"~*req.7\"},{\"path\":\"MetricFilterIDs\",\"tag\":\"MetricFilterIDs\",\"type\":\"*variable\",\"value\":\"~*req.8\"},{\"path\":\"Blocker\",\"tag\":\"Blocker\",\"type\":\"*variable\",\"value\":\"~*req.9\"},{\"path\":\"Stored\",\"tag\":\"Stored\",\"type\":\"*variable\",\"value\":\"~*req.10\"},{\"path\":\"Weight\",\"tag\":\"Weight\",\"type\":\"*variable\",\"value\":\"~*req.11\"},{\"path\":\"ThresholdIDs\",\"tag\":\"ThresholdIDs\",\"type\":\"*variable\",\"value\":\"~*req.12\"}],\"file_name\":\"Stats.csv\",\"flags\":null,\"type\":\"*stats\"},{\"fields\":[{\"mandatory\":true,\"path\":\"Tenant\",\"tag\":\"Tenant\",\"type\":\"*variable\",\"value\":\"~*req.0\"},{\"mandatory\":true,\"path\":\"ID\",\"tag\":\"ID\",\"type\":\"*variable\",\"value\":\"~*req.1\"},{\"path\":\"FilterIDs\",\"tag\":\"FilterIDs\",\"type\":\"*variable\",\"value\":\"~*req.2\"},{\"path\":\"ActivationInterval\",\"tag\":\"ActivationInterval\",\"type\":\"*variable\",\"value\":\"~*req.3\"},{\"path\":\"MaxHits\",\"tag\":\"MaxHits\",\"type\":\"*variable\",\"value\":\"~*req.4\"},{\"path\":\"MinHits\",\"tag\":\"MinHits\",\"type\":\"*variable\",\"value\":\"~*req.5\"},{\"path\":\"MinSleep\",\"tag\":\"MinSleep\",\"type\":\"*variable\",\"value\":\"~*req.6\"},{\"path\":\"Blocker\",\"tag\":\"Blocker\",\"type\":\"*variable\",\"value\":\"~*req.7\"},{\"path\":\"Weight\",\"tag\":\"Weight\",\"type\":\"*variable\",\"value\":\"~*req.8\"},{\"path\":\"ActionIDs\",\"tag\":\"ActionIDs\",\"type\":\"*variable\",\"value\":\"~*req.9\"},{\"path\":\"Async\",\"tag\":\"Async\",\"type\":\"*variable\",\"value\":\"~*req.10\"}],\"file_name\":\"Thresholds.csv\",\"flags\":null,\"type\":\"*thresholds\"},{\"fields\":[{\"mandatory\":true,\"path\":\"Tenant\",\"tag\":\"Tenant\",\"type\":\"*variable\",\"value\":\"~*req.0\"},{\"mandatory\":true,\"path\":\"ID\",\"tag\":\"ID\",\"type\":\"*variable\",\"value\":\"~*req.1\"},{\"path\":\"FilterIDs\",\"tag\":\"FilterIDs\",\"type\":\"*variable\",\"value\":\"~*req.2\"},{\"path\":\"ActivationInterval\",\"tag\":\"ActivationInterval\",\"type\":\"*variable\",\"value\":\"~*req.3\"},{\"path\":\"Sorting\",\"tag\":\"Sorting\",\"type\":\"*variable\",\"value\":\"~*req.4\"},{\"path\":\"SortingParameters\",\"tag\":\"SortingParameters\",\"type\":\"*variable\",\"value\":\"~*req.5\"},{\"path\":\"RouteID\",\"tag\":\"RouteID\",\"type\":\"*variable\",\"value\":\"~*req.6\"},{\"path\":\"RouteFilterIDs\",\"tag\":\"RouteFilterIDs\",\"type\":\"*variable\",\"value\":\"~*req.7\"},{\"path\":\"RouteAccountIDs\",\"tag\":\"RouteAccountIDs\",\"type\":\"*variable\",\"value\":\"~*req.8\"},{\"path\":\"RouteRatingPlanIDs\",\"tag\":\"RouteRatingPlanIDs\",\"type\":\"*variable\",\"value\":\"~*req.9\"},{\"path\":\"RouteResourceIDs\",\"tag\":\"RouteResourceIDs\",\"type\":\"*variable\",\"value\":\"~*req.10\"},{\"path\":\"RouteStatIDs\",\"tag\":\"RouteStatIDs\",\"type\":\"*variable\",\"value\":\"~*req.11\"},{\"path\":\"RouteWeight\",\"tag\":\"RouteWeight\",\"type\":\"*variable\",\"value\":\"~*req.12\"},{\"path\":\"RouteBlocker\",\"tag\":\"RouteBlocker\",\"type\":\"*variable\",\"value\":\"~*req.13\"},{\"path\":\"RouteParameters\",\"tag\":\"RouteParameters\",\"type\":\"*variable\",\"value\":\"~*req.14\"},{\"path\":\"Weight\",\"tag\":\"Weight\",\"type\":\"*variable\",\"value\":\"~*req.15\"}],\"file_name\":\"Routes.csv\",\"flags\":null,\"type\":\"*routes\"},{\"fields\":[{\"mandatory\":true,\"path\":\"Tenant\",\"tag\":\"Tenant\",\"type\":\"*variable\",\"value\":\"~*req.0\"},{\"mandatory\":true,\"path\":\"ID\",\"tag\":\"ID\",\"type\":\"*variable\",\"value\":\"~*req.1\"},{\"path\":\"FilterIDs\",\"tag\":\"FilterIDs\",\"type\":\"*variable\",\"value\":\"~*req.2\"},{\"path\":\"ActivationInterval\",\"tag\":\"ActivationInterval\",\"type\":\"*variable\",\"value\":\"~*req.3\"},{\"path\":\"RunID\",\"tag\":\"RunID\",\"type\":\"*variable\",\"value\":\"~*req.4\"},{\"path\":\"AttributeIDs\",\"tag\":\"AttributeIDs\",\"type\":\"*variable\",\"value\":\"~*req.5\"},{\"path\":\"Weight\",\"tag\":\"Weight\",\"type\":\"*variable\",\"value\":\"~*req.6\"}],\"file_name\":\"Chargers.csv\",\"flags\":null,\"type\":\"*chargers\"},{\"fields\":[{\"mandatory\":true,\"path\":\"Tenant\",\"tag\":\"Tenant\",\"type\":\"*variable\",\"value\":\"~*req.0\"},{\"mandatory\":true,\"path\":\"ID\",\"tag\":\"ID\",\"type\":\"*variable\",\"value\":\"~*req.1\"},{\"path\":\"Contexts\",\"tag\":\"Contexts\",\"type\":\"*variable\",\"value\":\"~*req.2\"},{\"path\":\"FilterIDs\",\"tag\":\"FilterIDs\",\"type\":\"*variable\",\"value\":\"~*req.3\"},{\"path\":\"ActivationInterval\",\"tag\":\"ActivationInterval\",\"type\":\"*variable\",\"value\":\"~*req.4\"},{\"path\":\"Strategy\",\"tag\":\"Strategy\",\"type\":\"*variable\",\"value\":\"~*req.5\"},{\"path\":\"StrategyParameters\",\"tag\":\"StrategyParameters\",\"type\":\"*variable\",\"value\":\"~*req.6\"},{\"path\":\"ConnID\",\"tag\":\"ConnID\",\"type\":\"*variable\",\"value\":\"~*req.7\"},{\"path\":\"ConnFilterIDs\",\"tag\":\"ConnFilterIDs\",\"type\":\"*variable\",\"value\":\"~*req.8\"},{\"path\":\"ConnWeight\",\"tag\":\"ConnWeight\",\"type\":\"*variable\",\"value\":\"~*req.9\"},{\"path\":\"ConnBlocker\",\"tag\":\"ConnBlocker\",\"type\":\"*variable\",\"value\":\"~*req.10\"},{\"path\":\"ConnParameters\",\"tag\":\"ConnParameters\",\"type\":\"*variable\",\"value\":\"~*req.11\"},{\"path\":\"Weight\",\"tag\":\"Weight\",\"type\":\"*variable\",\"value\":\"~*req.12\"}],\"file_name\":\"DispatcherProfiles.csv\",\"flags\":null,\"type\":\"*dispatchers\"},{\"fields\":[{\"mandatory\":true,\"path\":\"Tenant\",\"tag\":\"Tenant\",\"type\":\"*variable\",\"value\":\"~*req.0\"},{\"mandatory\":true,\"path\":\"ID\",\"tag\":\"ID\",\"type\":\"*variable\",\"value\":\"~*req.1\"},{\"path\":\"Address\",\"tag\":\"Address\",\"type\":\"*variable\",\"value\":\"~*req.2\"},{\"path\":\"Transport\",\"tag\":\"Transport\",\"type\":\"*variable\",\"value\":\"~*req.3\"},{\"path\":\"ConnectAttempts\",\"tag\":\"ConnectAttempts\",\"type\":\"*variable\",\"value\":\"~*req.4\"},{\"path\":\"Reconnects\",\"tag\":\"Reconnects\",\"type\":\"*variable\",\"value\":\"~*req.5\"},{\"path\":\"MaxReconnectInterval\",\"tag\":\"MaxReconnectInterval\",\"type\":\"*variable\",\"value\":\"~*req.6\"},{\"path\":\"ConnectTimeout\",\"tag\":\"ConnectTimeout\",\"type\":\"*variable\",\"value\":\"~*req.7\"},{\"path\":\"ReplyTimeout\",\"tag\":\"ReplyTimeout\",\"type\":\"*variable\",\"value\":\"~*req.8\"},{\"path\":\"TLS\",\"tag\":\"TLS\",\"type\":\"*variable\",\"value\":\"~*req.9\"},{\"path\":\"ClientKey\",\"tag\":\"ClientKey\",\"type\":\"*variable\",\"value\":\"~*req.10\"},{\"path\":\"ClientCertificate\",\"tag\":\"ClientCertificate\",\"type\":\"*variable\",\"value\":\"~*req.11\"},{\"path\":\"CaCertificate\",\"tag\":\"CaCertificate\",\"type\":\"*variable\",\"value\":\"~*req.12\"}],\"file_name\":\"DispatcherHosts.csv\",\"flags\":null,\"type\":\"*dispatcher_hosts\"}],\"dry_run\":false,\"enabled\":false,\"field_separator\":\",\",\"id\":\"*default\",\"lockfile_path\":\".cgr.lck\",\"run_delay\":\"0\",\"tenant\":\"\",\"tp_in_dir\":\"/var/spool/cgrates/loader/in\",\"tp_out_dir\":\"/var/spool/cgrates/loader/out\"}]}"
	var rpl string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:  "cgrates.org",
		Section: config.LoaderJson,
	}, &rpl); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rpl, cfgStr) {
		t.Errorf("\nExpected %v ,\n received: %v", utils.ToJSON(cfgStr), utils.ToJSON(rpl))
	}
}

func testSectConfigSReloadMailer(t *testing.T) {

	var reply string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1SetConfigFromJSON, &config.SetConfigFromJSONArgs{
		Tenant: "cgrates.org",
		Config: "{\"mailer\":{\"auth_password\":\"CGRateS.com\",\"auth_user\":\"cgrates2\",\"from_address\":\"cgr-mailer@internal.localdomain\",\"server\":\"internal\"}}",
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := "{\"mailer\":{\"auth_password\":\"CGRateS.com\",\"auth_user\":\"cgrates2\",\"from_address\":\"cgr-mailer@internal.localdomain\",\"server\":\"internal\"}}"
	var rpl string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
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
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1SetConfigFromJSON, &config.SetConfigFromJSONArgs{
		Tenant: "cgrates.org",
		Config: `{"suretax": {
			"url": "google.ro",
			"client_number": "4",
			"validation_key": "5",
			"business_unit": "6",
			"timezone": "UTC",
			"include_local_cost": true,
			"return_file_code": "01",
			"response_group": "04",
			"response_type": "D5",
			"regulatory_code": "04",
			"client_tracking": "~*req.CGRID",
			"customer_number": "~*req.Subject",
			"orig_number":  "~*req.Subject",
			"term_number": "~*req.Destination",
			"bill_to_number": "3",
			"zipcode": "3",
			"plus4": "3",
			"p2pzipcode": "3",
			"p2pplus4": "3",
			"units": "2",
			"unit_type": "01",
			"tax_included": "02",
			"tax_situs_rule": "05",
			"trans_type_code": "101010",
			"sales_type_code": "RC",
			"tax_exemption_code_list": "3",
		},
	}}`,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := "{\"suretax\":{\"bill_to_number\":\"3\",\"business_unit\":\"6\",\"client_number\":\"4\",\"client_tracking\":\"~*req.CGRID\",\"customer_number\":\"~*req.Subject\",\"include_local_cost\":true,\"orig_number\":\"~*req.Subject\",\"p2pplus4\":\"3\",\"p2pzipcode\":\"3\",\"plus4\":\"3\",\"regulatory_code\":\"04\",\"response_group\":\"04\",\"response_type\":\"D5\",\"return_file_code\":\"01\",\"sales_type_code\":\"RC\",\"tax_exemption_code_list\":\"3\",\"tax_included\":\"02\",\"tax_situs_rule\":\"05\",\"term_number\":\"~*req.Destination\",\"timezone\":\"UTC\",\"trans_type_code\":\"101010\",\"unit_type\":\"01\",\"units\":\"2\",\"url\":\"google.ro\",\"validation_key\":\"5\",\"zipcode\":\"3\"}}"
	var rpl string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:  "cgrates.org",
		Section: config.SURETAX_JSON,
	}, &rpl); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rpl, cfgStr) {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl))
	}
}

func testSectConfigSReloadLoader(t *testing.T) {

	var reply string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1SetConfigFromJSON, &config.SetConfigFromJSONArgs{
		Tenant: "cgrates.org",
		Config: "{\"loader\":{\"caches_conns\":[\"*internal\"],\"data_path\":\".path\",\"disable_reverse\":true,\"field_separator\":\";\",\"gapi_credentials\":\".testCredentials\",\"gapi_token\":\".testToken\",\"scheduler_conns\":[\"*internal\"],\"tpid\":\"testID\"}}",
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := "{\"loader\":{\"caches_conns\":[\"*internal\"],\"data_path\":\".path\",\"disable_reverse\":true,\"field_separator\":\";\",\"gapi_credentials\":\".testCredentials\",\"gapi_token\":\".testToken\",\"scheduler_conns\":[\"*internal\"],\"tpid\":\"testID\"}}"
	var rpl string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
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
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1SetConfigFromJSON, &config.SetConfigFromJSONArgs{
		Tenant: "cgrates.org",
		Config: `{"migrator":{"out_datadb_encoding":"testDatadbEncoding","out_datadb_host":"127.1.1.1","out_datadb_name":"10","out_datadb_opts":{"mongoQueryTimeout":"0s","mongoConnScheme":"mongodb","redisCACertificate":"testRedisCACertificate","redisClientCertificate":"testRedisClientCertificate","redisClientKey":"testRedisClientKey","redisCluster":true,"redisClusterOndownDelay":"4ns","redisClusterSync":"10s","redisConnectAttempts":20,"redisConnectTimeout":"0s","redisMaxConns":10,"redisReadTimeout":"0s","redisSentinel":"redisTest","redisTLS":true,"redisWriteTimeout":"0s"},"out_datadb_password":"dataDBPass","out_datadb_port":"5555","out_datadb_type":"*redisTest","out_datadb_user":"cgratesTest","out_stordb_host":"125.1.1.1","out_stordb_name":"cgratesStorDBName","out_stordb_opts":{"mongoQueryTimeout":"0s","mongoConnScheme":"mongodb","mysqlDSNParams":null,"mysqlLocation":"","pgSSLMode":"","sqlConnMaxLifetime":"0s","sqlMaxIdleConns":0,"sqlMaxOpenConns":0},"out_stordb_password":"StorDBPass","out_stordb_port":"3333","out_stordb_type":"*mongo","out_stordb_user":"cgratesTest","users_filters":["Stats"]}}`,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := `{"migrator":{"out_datadb_encoding":"testDatadbEncoding","out_datadb_host":"127.1.1.1","out_datadb_name":"10","out_datadb_opts":{"mongoConnScheme":"mongodb","mongoQueryTimeout":"0s","redisCACertificate":"testRedisCACertificate","redisClientCertificate":"testRedisClientCertificate","redisClientKey":"testRedisClientKey","redisCluster":true,"redisClusterOndownDelay":"4ns","redisClusterSync":"10s","redisConnectAttempts":20,"redisConnectTimeout":"0s","redisMaxConns":10,"redisPoolPipelineLimit":0,"redisPoolPipelineWindow":"150µs","redisReadTimeout":"0s","redisSentinel":"redisTest","redisTLS":true,"redisWriteTimeout":"0s"},"out_datadb_password":"dataDBPass","out_datadb_port":"5555","out_datadb_type":"*redisTest","out_datadb_user":"cgratesTest","out_stordb_host":"125.1.1.1","out_stordb_name":"cgratesStorDBName","out_stordb_opts":{"mongoConnScheme":"mongodb","mongoQueryTimeout":"0s","mysqlDSNParams":null,"mysqlLocation":"","pgSSLMode":"","sqlConnMaxLifetime":"0s","sqlMaxIdleConns":0,"sqlMaxOpenConns":0},"out_stordb_password":"StorDBPass","out_stordb_port":"3333","out_stordb_type":"*mongo","out_stordb_user":"cgratesTest","users_filters":["Stats"]}}`
	var rpl string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:  "cgrates.org",
		Section: config.CgrMigratorCfgJson,
	}, &rpl); err != nil {
		t.Error(err)
	} else if cfgStr != rpl {
		t.Errorf("\nExpected %+v ,\n received: %+v", cfgStr, rpl)
	}
}

func testSectConfigSReloadDispatchers(t *testing.T) {

	var replyPingBf string
	err := testSectRPC.Call(context.Background(), utils.DispatcherSv1Ping, &utils.CGREvent{}, &replyPingBf)
	if err == nil || err.Error() != "rpc: can't find service DispatcherSv1.Ping" {
		t.Error(err)
	}
	var reply string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1SetConfigFromJSON, &config.SetConfigFromJSONArgs{
		Tenant: "cgrates.org",
		Config: "{\"dispatchers\":{\"any_subsystem\":true,\"attributes_conns\":[\"*internal\"],\"enabled\":true,\"indexed_selects\":true,\"nested_fields\":true,\"prefix_indexed_fields\":[\"*internal\"],\"string_indexed_fields\":[\"*internal\"],\"suffix_indexed_fields\":[\"*internal\"]}}",
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := "{\"dispatchers\":{\"any_subsystem\":true,\"attributes_conns\":[\"*internal\"],\"enabled\":true,\"indexed_selects\":true,\"nested_fields\":true,\"prefix_indexed_fields\":[\"*internal\"],\"prevent_loop\":false,\"string_indexed_fields\":[\"*internal\"],\"suffix_indexed_fields\":[\"*internal\"]}}"
	var rpl string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
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
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1SetConfigFromJSON, &config.SetConfigFromJSONArgs{
		Tenant: "cgrates.org",
		Config: "{\"registrarc\":{\"dispatchers\":{\"enabled\":true,\"hosts\":[{\"Tenant\":\"*default\"}],\"refresh_interval\":\"5m0s\",\"registrars_conns\":[]},\"rpc\":{\"enabled\":true,\"hosts\":[{\"Tenant\":\"*default\"}],\"refresh_interval\":\"5m0s\",\"registrars_conns\":[]}}}",
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := "{\"registrarc\":{\"dispatchers\":{\"hosts\":[{\"Tenant\":\"*default\",\"transport\":\"\"}],\"refresh_interval\":\"5m0s\",\"registrars_conns\":[]},\"rpc\":{\"hosts\":[{\"Tenant\":\"*default\",\"transport\":\"\"}],\"refresh_interval\":\"5m0s\",\"registrars_conns\":[]}}}"
	var rpl string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
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
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1SetConfigFromJSON, &config.SetConfigFromJSONArgs{
		Tenant: "cgrates.org",
		Config: "{\"analyzers\":{\"cleanup_interval\":\"1h0m0s\",\"db_path\":\"/var/spool/cgrates/analyzers\",\"enabled\":true,\"index_type\":\"*scorch\",\"ttl\":\"24h0m0s\"}}",
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := "{\"analyzers\":{\"cleanup_interval\":\"1h0m0s\",\"db_path\":\"/var/spool/cgrates/analyzers\",\"enabled\":true,\"index_type\":\"*scorch\",\"ttl\":\"24h0m0s\"}}"
	var rpl string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
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
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1SetConfigFromJSON, &config.SetConfigFromJSONArgs{
		Tenant: "cgrates.org",
		Config: "{\"apiers\":{\"attributes_conns\":[\"*internal\"],\"caches_conns\":[\"*internal\"],\"ees_conns\":[\"*internal\"],\"enabled\":true,\"scheduler_conns\":[\"*internal\"]}}",
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := "{\"apiers\":{\"attributes_conns\":[\"*internal\"],\"caches_conns\":[\"*internal\"],\"ees_conns\":[\"*internal\"],\"enabled\":true,\"scheduler_conns\":[\"*internal\"]}}"
	var rpl string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
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
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1SetConfigFromJSON, &config.SetConfigFromJSONArgs{
		Tenant: "cgrates.org",
		Config: `{"sip_agent": {
			"enabled": true,
			"listen": "127.0.0.1:5060",
			"listen_net": "udp",
			"sessions_conns": ["*internal"],
			"timezone": "local",
			"retransmission_timer": "100s",
			"request_processors": [
				{
					"id": "OutboundAUTHDryRun",
					"filters": ["*string:~*req.request_type:OutboundAUTH","*string:~*req.Msisdn:497700056231"],
					"tenant": "cgrates.org",
					"flags": ["*dryrun"],
					"timezone":       "",
					"request_fields":[
					],
					"reply_fields":[
						{"tag": "Allow", "path": "*rep.response.Allow", "type": "*constant",
							"value": "1", "mandatory": true},
						{"tag": "Concatenated1", "path": "*rep.response.Concatenated", "type": "*composed",
							"value": "~*req.MCC;/", "mandatory": true},
						{"tag": "Concatenated2", "path": "*rep.response.Concatenated", "type": "*composed",
							"value": "Val1"},
						{"tag": "MaxDuration", "path": "*rep.response.MaxDuration", "type": "*constant",
							"value": "1200", "blocker": true},
						{"tag": "Unused", "path": "*rep.response.Unused", "type": "*constant",
							"value": "0"},
						],
					},
			],
		},
		}}`,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := "{\"sip_agent\":{\"enabled\":true,\"listen\":\"127.0.0.1:5060\",\"listen_net\":\"udp\",\"request_processors\":[{\"filters\":[\"*string:~*req.request_type:OutboundAUTH\",\"*string:~*req.Msisdn:497700056231\"],\"flags\":[\"*dryrun\"],\"id\":\"OutboundAUTHDryRun\",\"reply_fields\":[{\"mandatory\":true,\"path\":\"*rep.response.Allow\",\"tag\":\"Allow\",\"type\":\"*constant\",\"value\":\"1\"},{\"mandatory\":true,\"path\":\"*rep.response.Concatenated\",\"tag\":\"Concatenated1\",\"type\":\"*composed\",\"value\":\"~*req.MCC;/\"},{\"path\":\"*rep.response.Concatenated\",\"tag\":\"Concatenated2\",\"type\":\"*composed\",\"value\":\"Val1\"},{\"blocker\":true,\"path\":\"*rep.response.MaxDuration\",\"tag\":\"MaxDuration\",\"type\":\"*constant\",\"value\":\"1200\"},{\"path\":\"*rep.response.Unused\",\"tag\":\"Unused\",\"type\":\"*constant\",\"value\":\"0\"}],\"request_fields\":[],\"tenant\":\"cgrates.org\",\"timezone\":\"\"}],\"retransmission_timer\":100000000000,\"sessions_conns\":[\"*internal\"],\"timezone\":\"local\"}}"
	var rpl string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:  "cgrates.org",
		Section: config.SIPAgentJson,
	}, &rpl); err != nil {
		t.Error(err)
	} else if cfgStr != rpl {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToJSON(cfgStr), utils.ToJSON(rpl))
	}
}

func testSectConfigSReloadTemplates(t *testing.T) {

	var reply string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1SetConfigFromJSON, &config.SetConfigFromJSONArgs{
		Tenant: "cgrates.org",
		Config: `{"templates":{"*asr":[{"mandatory":true,"path":"*diamreq.Session-Id","tag":"SessionId","type":"*variable","value":"~*req.Session-Id"},{"mandatory":true,"path":"*diamreq.Origin-Host","tag":"OriginHost","type":"*variable","value":"~*req.Destination-Host"},{"mandatory":true,"path":"*diamreq.Origin-Realm","tag":"OriginRealm","type":"*variable","value":"~*req.Destination-Realm"},{"mandatory":true,"path":"*diamreq.Destination-Realm","tag":"DestinationRealm","type":"*variable","value":"~*req.Origin-Realm"},{"mandatory":true,"path":"*diamreq.Destination-Host","tag":"DestinationHost","type":"*variable","value":"~*req.Origin-Host"},{"mandatory":true,"path":"*diamreq.Auth-Application-Id","tag":"AuthApplicationId","type":"*variable","value":"~*vars.*appid"}],"*cca":[{"mandatory":true,"path":"*rep.Session-Id","tag":"SessionId","type":"*variable","value":"~*req.Session-Id"},{"path":"*rep.Result-Code","tag":"ResultCode","type":"*constant","value":"2001"},{"mandatory":true,"path":"*rep.Origin-Host","tag":"OriginHost","type":"*variable","value":"~*vars.OriginHost"},{"mandatory":true,"path":"*rep.Origin-Realm","tag":"OriginRealm","type":"*variable","value":"~*vars.OriginRealm"},{"mandatory":true,"path":"*rep.Auth-Application-Id","tag":"AuthApplicationId","type":"*variable","value":"~*vars.*appid"},{"mandatory":true,"path":"*rep.CC-Request-Type","tag":"CCRequestType","type":"*variable","value":"~*req.CC-Request-Type"},{"mandatory":true,"path":"*rep.CC-Request-Number","tag":"CCRequestNumber","type":"*variable","value":"~*req.CC-Request-Number"}],"*cdrLog":[{"mandatory":true,"path":"*cdr.ToR","tag":"ToR","type":"*variable","value":"~*req.BalanceType"},{"mandatory":true,"path":"*cdr.OriginHost","tag":"OriginHost","type":"*constant","value":"127.0.0.1"},{"mandatory":true,"path":"*cdr.RequestType","tag":"RequestType","type":"*constant","value":"*none"},{"mandatory":true,"path":"*cdr.Tenant","tag":"Tenant","type":"*variable","value":"~*req.Tenant"},{"mandatory":true,"path":"*cdr.Account","tag":"Account","type":"*variable","value":"~*req.Account"},{"mandatory":true,"path":"*cdr.Subject","tag":"Subject","type":"*variable","value":"~*req.Account"},{"mandatory":true,"path":"*cdr.Cost","tag":"Cost","type":"*variable","value":"~*req.Cost"},{"mandatory":true,"path":"*cdr.Source","tag":"Source","type":"*constant","value":"*cdrLog"},{"mandatory":true,"path":"*cdr.Usage","tag":"Usage","type":"*constant","value":"1"},{"mandatory":true,"path":"*cdr.RunID","tag":"RunID","type":"*variable","value":"~*req.ActionType"},{"mandatory":true,"path":"*cdr.SetupTime","tag":"SetupTime","type":"*constant","value":"*now"},{"mandatory":true,"path":"*cdr.AnswerTime","tag":"AnswerTime","type":"*constant","value":"*now"},{"mandatory":true,"path":"*cdr.PreRated","tag":"PreRated","type":"*constant","value":"true"}],"*coa":[{"path":"*radDAReq.User-Name","tag":"User-Name","type":"*variable","value":"~*oreq.User-Name"},{"path":"*radDAReq.NAS-IP-Address","tag":"NAS-IP-Address","type":"*variable","value":"~*oreq.NAS-IP-Address"},{"path":"*radDAReq.Acct-Session-Id","tag":"Acct-Session-Id","type":"*variable","value":"~*oreq.Acct-Session-Id"},{"path":"*radDAReq.Filter-Id","tag":"Filter-Id","type":"*variable","value":"~*req.CustomFilter"}],"*dmr":[{"path":"*radDAReq.User-Name","tag":"User-Name","type":"*variable","value":"~*oreq.User-Name"},{"path":"*radDAReq.NAS-IP-Address","tag":"NAS-IP-Address","type":"*variable","value":"~*oreq.NAS-IP-Address"},{"path":"*radDAReq.Acct-Session-Id","tag":"Acct-Session-Id","type":"*variable","value":"~*oreq.Acct-Session-Id"},{"path":"*radDAReq.Reply-Message","tag":"Reply-Message","type":"*variable","value":"~*req.DisconnectCause"}],"*err":[{"mandatory":true,"path":"*rep.Session-Id","tag":"SessionId","type":"*variable","value":"~*req.Session-Id"},{"mandatory":true,"path":"*rep.Origin-Host","tag":"OriginHost","type":"*variable","value":"~*vars.OriginHost"},{"mandatory":true,"path":"*rep.Origin-Realm","tag":"OriginRealm","type":"*variable","value":"~*vars.OriginRealm"}],"*errSip":[{"mandatory":true,"path":"*rep.Request","tag":"Request","type":"*constant","value":"SIP/2.0 500 Internal Server Error"}],"*rar":[{"mandatory":true,"path":"*diamreq.Session-Id","tag":"SessionId","type":"*variable","value":"~*req.Session-Id"},{"mandatory":true,"path":"*diamreq.Origin-Host","tag":"OriginHost","type":"*variable","value":"~*req.Destination-Host"},{"mandatory":true,"path":"*diamreq.Origin-Realm","tag":"OriginRealm","type":"*variable","value":"~*req.Destination-Realm"},{"mandatory":true,"path":"*diamreq.Destination-Realm","tag":"DestinationRealm","type":"*variable","value":"~*req.Origin-Realm"},{"mandatory":true,"path":"*diamreq.Destination-Host","tag":"DestinationHost","type":"*variable","value":"~*req.Origin-Host"},{"mandatory":true,"path":"*diamreq.Auth-Application-Id","tag":"AuthApplicationId","type":"*variable","value":"~*vars.*appid"},{"path":"*diamreq.Re-Auth-Request-Type","tag":"ReAuthRequestType","type":"*constant","value":"0"}]}}`,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := `{"templates":{"*asr":[{"mandatory":true,"path":"*diamreq.Session-Id","tag":"SessionId","type":"*variable","value":"~*req.Session-Id"},{"mandatory":true,"path":"*diamreq.Origin-Host","tag":"OriginHost","type":"*variable","value":"~*req.Destination-Host"},{"mandatory":true,"path":"*diamreq.Origin-Realm","tag":"OriginRealm","type":"*variable","value":"~*req.Destination-Realm"},{"mandatory":true,"path":"*diamreq.Destination-Realm","tag":"DestinationRealm","type":"*variable","value":"~*req.Origin-Realm"},{"mandatory":true,"path":"*diamreq.Destination-Host","tag":"DestinationHost","type":"*variable","value":"~*req.Origin-Host"},{"mandatory":true,"path":"*diamreq.Auth-Application-Id","tag":"AuthApplicationId","type":"*variable","value":"~*vars.*appid"}],"*cca":[{"mandatory":true,"path":"*rep.Session-Id","tag":"SessionId","type":"*variable","value":"~*req.Session-Id"},{"path":"*rep.Result-Code","tag":"ResultCode","type":"*constant","value":"2001"},{"mandatory":true,"path":"*rep.Origin-Host","tag":"OriginHost","type":"*variable","value":"~*vars.OriginHost"},{"mandatory":true,"path":"*rep.Origin-Realm","tag":"OriginRealm","type":"*variable","value":"~*vars.OriginRealm"},{"mandatory":true,"path":"*rep.Auth-Application-Id","tag":"AuthApplicationId","type":"*variable","value":"~*vars.*appid"},{"mandatory":true,"path":"*rep.CC-Request-Type","tag":"CCRequestType","type":"*variable","value":"~*req.CC-Request-Type"},{"mandatory":true,"path":"*rep.CC-Request-Number","tag":"CCRequestNumber","type":"*variable","value":"~*req.CC-Request-Number"}],"*cdrLog":[{"mandatory":true,"path":"*cdr.ToR","tag":"ToR","type":"*variable","value":"~*req.BalanceType"},{"mandatory":true,"path":"*cdr.OriginHost","tag":"OriginHost","type":"*constant","value":"127.0.0.1"},{"mandatory":true,"path":"*cdr.RequestType","tag":"RequestType","type":"*constant","value":"*none"},{"mandatory":true,"path":"*cdr.Tenant","tag":"Tenant","type":"*variable","value":"~*req.Tenant"},{"mandatory":true,"path":"*cdr.Account","tag":"Account","type":"*variable","value":"~*req.Account"},{"mandatory":true,"path":"*cdr.Subject","tag":"Subject","type":"*variable","value":"~*req.Account"},{"mandatory":true,"path":"*cdr.Cost","tag":"Cost","type":"*variable","value":"~*req.Cost"},{"mandatory":true,"path":"*cdr.Source","tag":"Source","type":"*constant","value":"*cdrLog"},{"mandatory":true,"path":"*cdr.Usage","tag":"Usage","type":"*constant","value":"1"},{"mandatory":true,"path":"*cdr.RunID","tag":"RunID","type":"*variable","value":"~*req.ActionType"},{"mandatory":true,"path":"*cdr.SetupTime","tag":"SetupTime","type":"*constant","value":"*now"},{"mandatory":true,"path":"*cdr.AnswerTime","tag":"AnswerTime","type":"*constant","value":"*now"},{"mandatory":true,"path":"*cdr.PreRated","tag":"PreRated","type":"*constant","value":"true"}],"*coa":[{"path":"*radDAReq.User-Name","tag":"User-Name","type":"*variable","value":"~*oreq.User-Name"},{"path":"*radDAReq.NAS-IP-Address","tag":"NAS-IP-Address","type":"*variable","value":"~*oreq.NAS-IP-Address"},{"path":"*radDAReq.Acct-Session-Id","tag":"Acct-Session-Id","type":"*variable","value":"~*oreq.Acct-Session-Id"},{"path":"*radDAReq.Filter-Id","tag":"Filter-Id","type":"*variable","value":"~*req.CustomFilter"}],"*dmr":[{"path":"*radDAReq.User-Name","tag":"User-Name","type":"*variable","value":"~*oreq.User-Name"},{"path":"*radDAReq.NAS-IP-Address","tag":"NAS-IP-Address","type":"*variable","value":"~*oreq.NAS-IP-Address"},{"path":"*radDAReq.Acct-Session-Id","tag":"Acct-Session-Id","type":"*variable","value":"~*oreq.Acct-Session-Id"},{"path":"*radDAReq.Reply-Message","tag":"Reply-Message","type":"*variable","value":"~*req.DisconnectCause"}],"*err":[{"mandatory":true,"path":"*rep.Session-Id","tag":"SessionId","type":"*variable","value":"~*req.Session-Id"},{"mandatory":true,"path":"*rep.Origin-Host","tag":"OriginHost","type":"*variable","value":"~*vars.OriginHost"},{"mandatory":true,"path":"*rep.Origin-Realm","tag":"OriginRealm","type":"*variable","value":"~*vars.OriginRealm"}],"*errSip":[{"mandatory":true,"path":"*rep.Request","tag":"Request","type":"*constant","value":"SIP/2.0 500 Internal Server Error"}],"*rar":[{"mandatory":true,"path":"*diamreq.Session-Id","tag":"SessionId","type":"*variable","value":"~*req.Session-Id"},{"mandatory":true,"path":"*diamreq.Origin-Host","tag":"OriginHost","type":"*variable","value":"~*req.Destination-Host"},{"mandatory":true,"path":"*diamreq.Origin-Realm","tag":"OriginRealm","type":"*variable","value":"~*req.Destination-Realm"},{"mandatory":true,"path":"*diamreq.Destination-Realm","tag":"DestinationRealm","type":"*variable","value":"~*req.Origin-Realm"},{"mandatory":true,"path":"*diamreq.Destination-Host","tag":"DestinationHost","type":"*variable","value":"~*req.Origin-Host"},{"mandatory":true,"path":"*diamreq.Auth-Application-Id","tag":"AuthApplicationId","type":"*variable","value":"~*vars.*appid"},{"path":"*diamreq.Re-Auth-Request-Type","tag":"ReAuthRequestType","type":"*constant","value":"0"}]}}`
	var rpl string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
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
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1SetConfigFromJSON, &config.SetConfigFromJSONArgs{
		Tenant: "cgrates.org",
		Config: `{"configs":{"enabled":true,"root_dir":"root_dir","url":"/configs/"}}`,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := `{"configs":{"enabled":true,"root_dir":"root_dir","url":"/configs/"}}`
	var rpl string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
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
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1SetConfigFromJSON, &config.SetConfigFromJSONArgs{
		Tenant: "cgrates.org",
		Config: `{"apiban":{"enabled":true,"keys":["keys"]}}`,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %s", reply)
	}
	cfgStr := "{\"apiban\":{\"keys\":[\"keys\"]}}"
	var rpl string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:  "cgrates.org",
		Section: config.APIBanCfgJson,
	}, &rpl); err != nil {
		t.Error(err)
	} else if cfgStr != rpl {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl))
	}

}

func testSectStopCgrEngine(t *testing.T) {
	if err := os.RemoveAll("/var/spool/cgrates/analyzers"); err != nil {
		t.Error(err)
	}
	if err := engine.KillEngine(*utils.WaitRater); err != nil {
		t.Error(err)
	}
}
