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
		testSectConfigSReloadTLS,
		testSectConfigSReloadHTTP,
		testSectConfigSReloadCaches,
		testSectConfigSReloadFilters,
		testSectConfigSReloadCDRS,
		testSectConfigSReloadERS,
		testSectConfigSReloadEES,
		testSectConfigSReloadSessions,
		testSectConfigSReloadAsteriskAgent,
		testSectConfigSReloadFreeswitchAgent,
		testSectConfigSReloadKamailioAgent,
		testSectConfigSReloadDiameterAgent,
		testSectConfigSReloadHTTPAgent,
		testSectConfigSReloadDNSAgent,
		testSectConfigSReloadAttributes,
		testSectConfigSReloadChargers,
		testSectConfigSReloadResources,
		testSectConfigSReloadStats,
		testSectConfigSReloadThresholds,
		testSectConfigSReloadRoutes,
		testSectConfigSReloadLoaders,
		testSectConfigSReloadSuretax,
		testSectConfigSReloadLoader,
		testSectConfigSReloadMigrator,
		testSectConfigSReloadDispatchers,
		testSectConfigSReloadRegistrarC,
		testSectConfigSReloadAnalyzer,
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
	if testSectCfg, err = config.NewCGRConfigFromPath(context.Background(), testSectCfgPath); err != nil {
		t.Error(err)
	}
}

func testSectResetDataDB(t *testing.T) {
	if err := engine.InitDataDB(testSectCfg); err != nil {
		t.Fatal(err)
	}
}

func testSectResetStorDb(t *testing.T) {
	if err := engine.InitStorDB(testSectCfg); err != nil {
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

func testSectConfigSReloadGeneral(t *testing.T) {

	var reply string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1SetConfigFromJSON, &config.SetConfigFromJSONArgs{
		Tenant: "cgrates.org",
		Config: "{\"general\":{\"connect_attempts\":5,\"connect_timeout\":\"1s\",\"dbdata_encoding\":\"*msgpack\",\"default_caching\":\"*reload\",\"default_category\":\"call\",\"default_request_type\":\"*rated\",\"default_tenant\":\"cgrates.org\",\"default_timezone\":\"Local\",\"digest_equal\":\":\",\"digest_separator\":\",\",\"failed_posts_dir\":\"/var/spool/cgrates/failed_posts\",\"failed_posts_ttl\":\"5s\",\"locking_timeout\":\"0\",\"log_level\":7,\"logger\":\"*syslog\",\"max_parallel_conns\":100,\"node_id\":\"98ead14\",\"poster_attempts\":3,\"reconnects\":-1,\"reply_timeout\":\"50s\",\"rounding_decimals\":5,\"rsr_separator\":\";\",\"tpexport_dir\":\"/var/spool/cgrates/tpe\"}}",
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := "{\"general\":{\"connect_attempts\":5,\"connect_timeout\":\"1s\",\"dbdata_encoding\":\"*msgpack\",\"default_caching\":\"*reload\",\"default_category\":\"call\",\"default_request_type\":\"*rated\",\"default_tenant\":\"cgrates.org\",\"default_timezone\":\"Local\",\"digest_equal\":\":\",\"digest_separator\":\",\",\"failed_posts_dir\":\"/var/spool/cgrates/failed_posts\",\"failed_posts_ttl\":\"5s\",\"locking_timeout\":\"0\",\"log_level\":7,\"logger\":\"*syslog\",\"max_parallel_conns\":100,\"node_id\":\"98ead14\",\"poster_attempts\":3,\"reconnects\":-1,\"reply_timeout\":\"50s\",\"rounding_decimals\":5,\"rsr_separator\":\";\",\"tpexport_dir\":\"/var/spool/cgrates/tpe\"}}"
	var rpl string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.GeneralJSON},
	}, &rpl); err != nil {
		t.Error(err)
	} else if cfgStr != rpl {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl))
	}
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
		Tenant:   "cgrates.org",
		Sections: []string{config.CoreSJSON},
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
		Tenant:   "cgrates.org",
		Sections: []string{config.RPCConnsJSON},
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
		Config: "{\"data_db\":{\"db_host\":\"127.0.0.1\",\"db_name\":\"10\",\"db_password\":\"\",\"db_port\":6379,\"db_type\":\"*internal\",\"db_user\":\"cgrates\",\"items\":{\"*account_action_plans\":{\"remote\":false,\"replicate\":false},\"*accounts\":{\"remote\":false,\"replicate\":false},\"*action_plans\":{\"remote\":false,\"replicate\":false},\"*action_triggers\":{\"remote\":false,\"replicate\":false},\"*actions\":{\"remote\":false,\"replicate\":false},\"*attribute_profiles\":{\"remote\":false,\"replicate\":false},\"*charger_profiles\":{\"remote\":false,\"replicate\":false},\"*destinations\":{\"remote\":false,\"replicate\":false},\"*dispatcher_hosts\":{\"remote\":false,\"replicate\":false},\"*dispatcher_profiles\":{\"remote\":false,\"replicate\":false},\"*filters\":{\"remote\":false,\"replicate\":false},\"*indexes\":{\"remote\":false,\"replicate\":false},\"*load_ids\":{\"remote\":false,\"replicate\":false},\"*rating_plans\":{\"remote\":false,\"replicate\":false},\"*rating_profiles\":{\"remote\":false,\"replicate\":false},\"*resource_profiles\":{\"remote\":false,\"replicate\":false},\"*resources\":{\"remote\":false,\"replicate\":false},\"*reverse_destinations\":{\"remote\":false,\"replicate\":false},\"*route_profiles\":{\"remote\":false,\"replicate\":false},\"*statqueues\":{\"remote\":false,\"replicate\":false},\"*threshold_profiles\":{\"remote\":false,\"replicate\":false},\"*thresholds\":{\"remote\":false,\"replicate\":false},\"opts\":{\"mongoQueryTimeout\":\"10s\",\"redisCACertificate\":\"\",\"redisClientCertificate\":\"\",\"redisClientKey\":\"\",\"redisCluster\":false,\"redisClusterOndownDelay\":\"0\",\"redisClusterSync\":\"5s\",\"redisSentinel\":\"\",\"redisTLS\":false},\"remote_conn_id\":\"\",\"remote_conns\":[],\"replication_cache\":\"\",\"replication_conns\":[],\"replication_filtered\":false}}",
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := "{\"data_db\":{\"db_host\":\"127.0.0.1\",\"db_name\":\"10\",\"db_password\":\"\",\"db_port\":6379,\"db_type\":\"*internal\",\"db_user\":\"cgrates\",\"items\":{\"*account_action_plans\":{\"remote\":false,\"replicate\":false},\"*accounts\":{\"remote\":false,\"replicate\":false},\"*action_plans\":{\"remote\":false,\"replicate\":false},\"*action_triggers\":{\"remote\":false,\"replicate\":false},\"*actions\":{\"remote\":false,\"replicate\":false},\"*attribute_profiles\":{\"remote\":false,\"replicate\":false},\"*charger_profiles\":{\"remote\":false,\"replicate\":false},\"*destinations\":{\"remote\":false,\"replicate\":false},\"*dispatcher_hosts\":{\"remote\":false,\"replicate\":false},\"*dispatcher_profiles\":{\"remote\":false,\"replicate\":false},\"*filters\":{\"remote\":false,\"replicate\":false},\"*indexes\":{\"remote\":false,\"replicate\":false},\"*load_ids\":{\"remote\":false,\"replicate\":false},\"*rating_plans\":{\"remote\":false,\"replicate\":false},\"*rating_profiles\":{\"remote\":false,\"replicate\":false},\"*resource_profiles\":{\"remote\":false,\"replicate\":false},\"*resources\":{\"remote\":false,\"replicate\":false},\"*reverse_destinations\":{\"remote\":false,\"replicate\":false},\"*route_profiles\":{\"remote\":false,\"replicate\":false},\"*statqueue_profiles\":{\"remote\":false,\"replicate\":false},\"*statqueues\":{\"remote\":false,\"replicate\":false},\"*threshold_profiles\":{\"remote\":false,\"replicate\":false},\"*thresholds\":{\"remote\":false,\"replicate\":false},\"opts\":{\"mongoQueryTimeout\":\"10s\",\"redisCACertificate\":\"\",\"redisClientCertificate\":\"\",\"redisClientKey\":\"\",\"redisCluster\":false,\"redisClusterOndownDelay\":\"0\",\"redisClusterSync\":\"5s\",\"redisSentinel\":\"\",\"redisTLS\":false},\"remote_conn_id\":\"\",\"remote_conns\":[],\"replication_cache\":\"\",\"replication_conns\":[],\"replication_filtered\":false}}"
	var rpl string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.DataDBJSON},
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
		Config: "{\"stor_db\":{\"db_host\":\"127.0.0.1\",\"db_name\":\"cgrates\",\"db_password\":\"CGRateS.org\",\"db_port\":3306,\"db_type\":\"*internal\",\"db_user\":\"cgrates\",\"items\":{\"*cdrs\":{\"remote\":false,\"replicate\":false},\"*session_costs\":{\"remote\":false,\"replicate\":false},\"*tp_account_actions\":{\"remote\":false,\"replicate\":false},\"*tp_action_plans\":{\"remote\":false,\"replicate\":false},\"*tp_action_triggers\":{\"remote\":false,\"replicate\":false},\"*tp_actions\":{\"remote\":false,\"replicate\":false},\"*tp_attributes\":{\"remote\":false,\"replicate\":false},\"*tp_chargers\":{\"remote\":false,\"replicate\":false},\"*tp_destination_rates\":{\"remote\":false,\"replicate\":false},\"*tp_destinations\":{\"remote\":false,\"replicate\":false},\"*tp_dispatcher_hosts\":{\"remote\":false,\"replicate\":false},\"*tp_dispatcher_profiles\":{\"remote\":false,\"replicate\":false},\"*tp_filters\":{\"remote\":false,\"replicate\":false},\"*tp_rates\":{\"remote\":false,\"replicate\":false},\"*tp_rating_plans\":{\"remote\":false,\"replicate\":false},\"*tp_resources\":{\"remote\":false,\"replicate\":false},\"*tp_routes\":{\"remote\":false,\"replicate\":false},\"*tp_stats\":{\"remote\":false,\"replicate\":false},\"*tp_thresholds\":{\"remote\":false,\"replicate\":false},\"*tp_timings\":{\"remote\":false,\"replicate\":false},\"*versions\":{\"remote\":false,\"replicate\":false}},\"opts\":{\"mongoQueryTimeout\":\"10s\",\"mysqlLocation\":\"Local\",\"postgresSSLMode\":\"disable\",\"sqlConnMaxLifetime\":0,\"sqlMaxIdleConns\":10,\"sqlMaxOpenConns\":100},\"prefix_indexed_fields\":[],\"remote_conns\":null,\"replication_conns\":null,\"string_indexed_fields\":[]}}",
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := "{\"stor_db\":{\"db_host\":\"127.0.0.1\",\"db_name\":\"cgrates\",\"db_password\":\"CGRateS.org\",\"db_port\":3306,\"db_type\":\"*internal\",\"db_user\":\"cgrates\",\"items\":{\"*cdrs\":{\"remote\":false,\"replicate\":false},\"*session_costs\":{\"remote\":false,\"replicate\":false},\"*tp_account_actions\":{\"remote\":false,\"replicate\":false},\"*tp_action_plans\":{\"remote\":false,\"replicate\":false},\"*tp_action_triggers\":{\"remote\":false,\"replicate\":false},\"*tp_actions\":{\"remote\":false,\"replicate\":false},\"*tp_attributes\":{\"remote\":false,\"replicate\":false},\"*tp_chargers\":{\"remote\":false,\"replicate\":false},\"*tp_destination_rates\":{\"remote\":false,\"replicate\":false},\"*tp_destinations\":{\"remote\":false,\"replicate\":false},\"*tp_dispatcher_hosts\":{\"remote\":false,\"replicate\":false},\"*tp_dispatcher_profiles\":{\"remote\":false,\"replicate\":false},\"*tp_filters\":{\"remote\":false,\"replicate\":false},\"*tp_rates\":{\"remote\":false,\"replicate\":false},\"*tp_rating_plans\":{\"remote\":false,\"replicate\":false},\"*tp_resources\":{\"remote\":false,\"replicate\":false},\"*tp_routes\":{\"remote\":false,\"replicate\":false},\"*tp_stats\":{\"remote\":false,\"replicate\":false},\"*tp_thresholds\":{\"remote\":false,\"replicate\":false},\"*tp_timings\":{\"remote\":false,\"replicate\":false},\"*versions\":{\"remote\":false,\"replicate\":false}},\"opts\":{\"mongoQueryTimeout\":\"10s\",\"mysqlLocation\":\"Local\",\"postgresSSLMode\":\"disable\",\"sqlConnMaxLifetime\":0,\"sqlMaxIdleConns\":10,\"sqlMaxOpenConns\":100},\"prefix_indexed_fields\":[],\"remote_conns\":null,\"replication_conns\":null,\"string_indexed_fields\":[]}}"
	var rpl string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.StorDBJSON},
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
		Tenant:   "cgrates.org",
		Sections: []string{config.ListenJSON},
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
		Tenant:   "cgrates.org",
		Sections: []string{config.TlsJSON},
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
		Config: "{\"http_agent\":[]}",
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := "{\"http_agent\":[]}"
	var rpl string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.HTTPAgentJSON},
	}, &rpl); err != nil {
		t.Error(err)
	} else if cfgStr != rpl {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl))
	}
}

func testSectConfigSReloadCaches(t *testing.T) {
	var replyPingBf string
	if err := testSectRPC.Call(context.Background(), utils.CacheSv1Ping, &utils.CGREvent{}, &replyPingBf); err != nil {
		t.Error(err)
	} else if replyPingBf != utils.Pong {
		t.Errorf("Expected OK received: %s", replyPingBf)
	}

	if testSectCfgDir == "tutmysql" {
		var reply string
		if err := testSectRPC.Call(context.Background(), utils.ConfigSv1SetConfigFromJSON, &config.SetConfigFromJSONArgs{
			Tenant: "cgrates.org",
			Config: "{\"caches\":{\"partitions\":{\"*account_filter_indexes\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*accounts\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*action_profile_filter_indexes\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*action_profiles\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*apiban\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"2m0s\"},\"*attribute_filter_indexes\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*attribute_profiles\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*caps_events\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*cdr_ids\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"10m0s\"},\"*cdrs\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*charger_filter_indexes\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*charger_profiles\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*closed_sessions\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"10s\"},\"*diameter_messages\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"3h0m0s\"},\"*dispatcher_filter_indexes\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*dispatcher_hosts\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*dispatcher_loads\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*dispatcher_profiles\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*dispatcher_routes\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*dispatchers\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*event_charges\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"10s\"},\"*event_resources\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*filters\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*load_ids\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*rate_filter_indexes\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*rate_profile_filter_indexes\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*rate_profiles\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*replication_hosts\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*resource_filter_indexes\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*resource_profiles\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*resources\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*reverse_filter_indexes\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*route_filter_indexes\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*route_profiles\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*rpc_connections\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*rpc_responses\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"2s\"},\"*session_costs\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*stat_filter_indexes\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*statqueue_profiles\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*statqueues\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*stir\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"3h0m0s\"},\"*threshold_filter_indexes\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*threshold_profiles\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*thresholds\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*tp_attributes\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*tp_chargers\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*tp_dispatcher_hosts\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*tp_dispatcher_profiles\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*tp_filters\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*tp_resources\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*tp_routes\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*tp_stats\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*tp_thresholds\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*uch\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"3h0m0s\"},\"*versions\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false}},\"replication_conns\":[]}}",
		}, &reply); err != nil {
			t.Error(err)
		} else if reply != utils.OK {
			t.Errorf("Expected OK received: %+v", reply)
		}
		cfgStr := "{\"caches\":{\"partitions\":{\"*account_filter_indexes\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*accounts\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*action_profile_filter_indexes\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*action_profiles\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*apiban\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"2m0s\"},\"*attribute_filter_indexes\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*attribute_profiles\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*caps_events\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*cdr_ids\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"10m0s\"},\"*cdrs\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*charger_filter_indexes\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*charger_profiles\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*closed_sessions\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"10s\"},\"*diameter_messages\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"3h0m0s\"},\"*dispatcher_filter_indexes\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*dispatcher_hosts\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*dispatcher_loads\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*dispatcher_profiles\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*dispatcher_routes\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*dispatchers\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*event_charges\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"10s\"},\"*event_resources\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*filters\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*load_ids\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*rate_filter_indexes\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*rate_profile_filter_indexes\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*rate_profiles\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*replication_hosts\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*resource_filter_indexes\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*resource_profiles\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*resources\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*reverse_filter_indexes\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*route_filter_indexes\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*route_profiles\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*rpc_connections\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*rpc_responses\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"2s\"},\"*session_costs\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*stat_filter_indexes\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*statqueue_profiles\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*statqueues\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*stir\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"3h0m0s\"},\"*threshold_filter_indexes\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*threshold_profiles\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*thresholds\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*tp_attributes\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*tp_chargers\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*tp_dispatcher_hosts\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*tp_dispatcher_profiles\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*tp_filters\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*tp_resources\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*tp_routes\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*tp_stats\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*tp_thresholds\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*uch\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"3h0m0s\"},\"*versions\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false}},\"replication_conns\":[]}}"
		var rpl string
		if err := testSectRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
			Tenant:   "cgrates.org",
			Sections: []string{config.CacheJSON},
		}, &rpl); err != nil {
			t.Error(err)
		} else if cfgStr != rpl {
			t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl))
		}
	} else if testSectCfgDir == "tutmongo" {
		var reply string
		if err := testSectRPC.Call(context.Background(), utils.ConfigSv1SetConfigFromJSON, &config.SetConfigFromJSONArgs{
			Tenant: "cgrates.org",
			Config: "{\"caches\":{\"partitions\":{\"*account_filter_indexes\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*accounts\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*action_profile_filter_indexes\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*action_profiles\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*apiban\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"2m0s\"},\"*attribute_filter_indexes\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*attribute_profiles\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*caps_events\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*cdr_ids\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"10m0s\"},\"*cdrs\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*charger_filter_indexes\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*charger_profiles\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*closed_sessions\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"10s\"},\"*diameter_messages\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"3h0m0s\"},\"*dispatcher_filter_indexes\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*dispatcher_hosts\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*dispatcher_loads\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*dispatcher_profiles\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*dispatcher_routes\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*dispatchers\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*event_charges\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"10s\"},\"*event_resources\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*filters\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*load_ids\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*rate_filter_indexes\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*rate_profile_filter_indexes\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*rate_profiles\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*replication_hosts\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*resource_filter_indexes\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*resource_profiles\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*resources\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*reverse_filter_indexes\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*route_filter_indexes\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*route_profiles\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*rpc_connections\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*rpc_responses\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"2s\"},\"*session_costs\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*stat_filter_indexes\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*statqueue_profiles\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*statqueues\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*stir\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"3h0m0s\"},\"*threshold_filter_indexes\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*threshold_profiles\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*thresholds\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*tp_attributes\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*tp_chargers\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*tp_dispatcher_hosts\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*tp_dispatcher_profiles\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*tp_filters\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*tp_resources\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*tp_routes\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*tp_stats\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*tp_thresholds\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*uch\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"3h0m0s\"},\"*versions\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false}},\"replication_conns\":[]}}",
		}, &reply); err != nil {
			t.Error(err)
		} else if reply != utils.OK {
			t.Errorf("Expected OK received: %+v", reply)
		}
		cfgStr := "{\"caches\":{\"partitions\":{\"*account_filter_indexes\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*accounts\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*action_profile_filter_indexes\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*action_profiles\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*apiban\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"2m0s\"},\"*attribute_filter_indexes\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*attribute_profiles\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*caps_events\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*cdr_ids\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"10m0s\"},\"*cdrs\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*charger_filter_indexes\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*charger_profiles\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*closed_sessions\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"10s\"},\"*diameter_messages\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"3h0m0s\"},\"*dispatcher_filter_indexes\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*dispatcher_hosts\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*dispatcher_loads\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*dispatcher_profiles\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*dispatcher_routes\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*dispatchers\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*event_charges\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"10s\"},\"*event_resources\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*filters\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*load_ids\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*rate_filter_indexes\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*rate_profile_filter_indexes\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*rate_profiles\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*replication_hosts\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*resource_filter_indexes\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*resource_profiles\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*resources\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*reverse_filter_indexes\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*route_filter_indexes\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*route_profiles\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*rpc_connections\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*rpc_responses\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"2s\"},\"*session_costs\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*stat_filter_indexes\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*statqueue_profiles\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*statqueues\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*stir\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"3h0m0s\"},\"*threshold_filter_indexes\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*threshold_profiles\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*thresholds\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*tp_attributes\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*tp_chargers\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*tp_dispatcher_hosts\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*tp_dispatcher_profiles\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*tp_filters\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*tp_resources\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*tp_routes\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*tp_stats\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*tp_thresholds\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*uch\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"3h0m0s\"},\"*versions\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false}},\"replication_conns\":[]}}"
		var rpl string
		if err := testSectRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
			Tenant:   "cgrates.org",
			Sections: []string{config.CacheJSON},
		}, &rpl); err != nil {
			t.Error(err)
		} else if cfgStr != rpl {
			t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl))
		}
	} else if testSectCfgDir == "tutinternal" {

		var reply string
		if err := testSectRPC.Call(context.Background(), utils.ConfigSv1SetConfigFromJSON, &config.SetConfigFromJSONArgs{
			Tenant: "cgrates.org",
			Config: "{\"caches\":{\"partitions\":{\"*account_filter_indexes\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*accounts\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*action_profile_filter_indexes\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*action_profiles\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*apiban\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"2m0s\"},\"*attribute_filter_indexes\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*attribute_profiles\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*caps_events\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*cdr_ids\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"10m0s\"},\"*cdrs\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*charger_filter_indexes\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*charger_profiles\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*closed_sessions\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"10s\"},\"*diameter_messages\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"3h0m0s\"},\"*dispatcher_filter_indexes\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*dispatcher_hosts\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*dispatcher_loads\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*dispatcher_profiles\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*dispatcher_routes\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*dispatchers\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*event_charges\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"10s\"},\"*event_resources\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*filters\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*load_ids\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*rate_filter_indexes\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*rate_profile_filter_indexes\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*rate_profiles\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*replication_hosts\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*resource_filter_indexes\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*resource_profiles\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*resources\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*reverse_filter_indexes\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*route_filter_indexes\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*route_profiles\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*rpc_connections\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*rpc_responses\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"2s\"},\"*session_costs\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*stat_filter_indexes\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*statqueue_profiles\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*statqueues\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*stir\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"3h0m0s\"},\"*threshold_filter_indexes\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*threshold_profiles\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*thresholds\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*tp_attributes\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*tp_chargers\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*tp_dispatcher_hosts\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*tp_dispatcher_profiles\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*tp_filters\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*tp_resources\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*tp_routes\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*tp_stats\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*tp_thresholds\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*uch\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"3h0m0s\"},\"*versions\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false}},\"replication_conns\":[]}}",
		}, &reply); err != nil {
			t.Error(err)
		} else if reply != utils.OK {
			t.Errorf("Expected OK received: %+v", reply)
		}
		cfgStr := "{\"caches\":{\"partitions\":{\"*account_filter_indexes\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*accounts\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*action_profile_filter_indexes\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*action_profiles\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*apiban\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"2m0s\"},\"*attribute_filter_indexes\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*attribute_profiles\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*caps_events\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*cdr_ids\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"10m0s\"},\"*cdrs\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*charger_filter_indexes\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*charger_profiles\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*closed_sessions\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"10s\"},\"*diameter_messages\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"3h0m0s\"},\"*dispatcher_filter_indexes\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*dispatcher_hosts\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*dispatcher_loads\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*dispatcher_profiles\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*dispatcher_routes\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*dispatchers\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*event_charges\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"10s\"},\"*event_resources\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*filters\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*load_ids\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*rate_filter_indexes\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*rate_profile_filter_indexes\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*rate_profiles\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*replication_hosts\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*resource_filter_indexes\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*resource_profiles\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*resources\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*reverse_filter_indexes\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*route_filter_indexes\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*route_profiles\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*rpc_connections\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*rpc_responses\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"2s\"},\"*session_costs\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*stat_filter_indexes\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*statqueue_profiles\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*statqueues\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*stir\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"3h0m0s\"},\"*threshold_filter_indexes\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*threshold_profiles\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*thresholds\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*tp_attributes\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*tp_chargers\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*tp_dispatcher_hosts\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*tp_dispatcher_profiles\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*tp_filters\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*tp_resources\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*tp_routes\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*tp_stats\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*tp_thresholds\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false},\"*uch\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"3h0m0s\"},\"*versions\":{\"limit\":0,\"precache\":false,\"replicate\":false,\"static_ttl\":false}},\"replication_conns\":[]}}"
		var rpl string
		if err := testSectRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
			Tenant:   "cgrates.org",
			Sections: []string{config.CacheJSON},
		}, &rpl); err != nil {
			t.Error(err)
		} else if cfgStr != rpl {
			t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl))
		}
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
		Config: "{\"filters\":{\"accounts_conns\":[\"*internal\"],\"resources_conns\":[\"*internal\"],\"stats_conns\":[\"*localhost\"]}}",
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := "{\"filters\":{\"accounts_conns\":[\"*internal\"],\"resources_conns\":[\"*internal\"],\"stats_conns\":[\"*localhost\"]}}"
	var rpl string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.FilterSJSON},
	}, &rpl); err != nil {
		t.Error(err)
	} else if cfgStr != rpl {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl))
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
		Config: "{\"cdrs\":{\"accounts_conns\":[],\"actions_conns\":[],\"attributes_conns\":[],\"chargers_conns\":[\"*internal\"],\"ees_conns\":[],\"enabled\":true,\"extra_fields\":[],\"online_cdr_exports\":null,\"opts\":{\"*accountS\":[],\"*attributeS\":[],\"*chargerS\":[],\"*eeS\":[],\"*rateS\":[],\"*statS\":[],\"*thresholdS\":[]},\"rates_conns\":[],\"session_cost_retries\":5,\"stats_conns\":[],\"store_cdrs\":true,\"thresholds_conns\":[]}}",
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := "{\"cdrs\":{\"accounts_conns\":[],\"actions_conns\":[],\"attributes_conns\":[],\"chargers_conns\":[\"*internal\"],\"ees_conns\":[],\"enabled\":true,\"extra_fields\":[],\"online_cdr_exports\":null,\"opts\":{\"*accountS\":[],\"*attributeS\":[],\"*chargerS\":[],\"*eeS\":[],\"*rateS\":[],\"*statS\":[],\"*thresholdS\":[]},\"rates_conns\":[],\"session_cost_retries\":5,\"stats_conns\":[],\"store_cdrs\":true,\"thresholds_conns\":[]}}"
	var rpl string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.CDRsJSON},
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
		Config: "{\"ers\":{\"enabled\":true,\"partial_cache_ttl\":\"1s\",\"readers\":[{\"cache_dump_fields\":[],\"concurrent_requests\":1024,\"fields\":[{\"mandatory\":true,\"path\":\"*cgreq.ToR\",\"tag\":\"ToR\",\"type\":\"*variable\",\"value\":\"~*req.2\"},{\"mandatory\":true,\"path\":\"*cgreq.OriginID\",\"tag\":\"OriginID\",\"type\":\"*variable\",\"value\":\"~*req.3\"},{\"mandatory\":true,\"path\":\"*cgreq.RequestType\",\"tag\":\"RequestType\",\"type\":\"*variable\",\"value\":\"~*req.4\"},{\"mandatory\":true,\"path\":\"*cgreq.Tenant\",\"tag\":\"Tenant\",\"type\":\"*variable\",\"value\":\"~*req.6\"},{\"mandatory\":true,\"path\":\"*cgreq.Category\",\"tag\":\"Category\",\"type\":\"*variable\",\"value\":\"~*req.7\"},{\"mandatory\":true,\"path\":\"*cgreq.Account\",\"tag\":\"Account\",\"type\":\"*variable\",\"value\":\"~*req.8\"},{\"mandatory\":true,\"path\":\"*cgreq.Subject\",\"tag\":\"Subject\",\"type\":\"*variable\",\"value\":\"~*req.9\"},{\"mandatory\":true,\"path\":\"*cgreq.Destination\",\"tag\":\"Destination\",\"type\":\"*variable\",\"value\":\"~*req.10\"},{\"mandatory\":true,\"path\":\"*cgreq.SetupTime\",\"tag\":\"SetupTime\",\"type\":\"*variable\",\"value\":\"~*req.11\"},{\"mandatory\":true,\"path\":\"*cgreq.AnswerTime\",\"tag\":\"AnswerTime\",\"type\":\"*variable\",\"value\":\"~*req.12\"},{\"mandatory\":true,\"path\":\"*cgreq.Usage\",\"tag\":\"Usage\",\"type\":\"*variable\",\"value\":\"~*req.13\"}],\"filters\":[],\"flags\":[],\"id\":\"*default\",\"opts\":{\"csvFieldSeparator\":\",\",\"csvHeaderDefineChar\":\":\",\"csvRowLength\":0,\"natsSubject\":\"cgrates_cdrs\",\"partialCacheAction\":\"*none\",\"partialOrderField\":\"~*req.AnswerTime\",\"xmlRootPath\":\"\"},\"partial_commit_fields\":[],\"processed_path\":\"/var/spool/cgrates/ers/out\",\"run_delay\":\"0\",\"source_path\":\"/var/spool/cgrates/ers/in\",\"tenant\":\"\",\"timezone\":\"\",\"type\":\"*none\"}],\"sessions_conns\":[\"*internal\"]}}",
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := "{\"ers\":{\"enabled\":true,\"partial_cache_ttl\":\"1s\",\"readers\":[{\"cache_dump_fields\":[],\"concurrent_requests\":1024,\"fields\":[{\"mandatory\":true,\"path\":\"*cgreq.ToR\",\"tag\":\"ToR\",\"type\":\"*variable\",\"value\":\"~*req.2\"},{\"mandatory\":true,\"path\":\"*cgreq.OriginID\",\"tag\":\"OriginID\",\"type\":\"*variable\",\"value\":\"~*req.3\"},{\"mandatory\":true,\"path\":\"*cgreq.RequestType\",\"tag\":\"RequestType\",\"type\":\"*variable\",\"value\":\"~*req.4\"},{\"mandatory\":true,\"path\":\"*cgreq.Tenant\",\"tag\":\"Tenant\",\"type\":\"*variable\",\"value\":\"~*req.6\"},{\"mandatory\":true,\"path\":\"*cgreq.Category\",\"tag\":\"Category\",\"type\":\"*variable\",\"value\":\"~*req.7\"},{\"mandatory\":true,\"path\":\"*cgreq.Account\",\"tag\":\"Account\",\"type\":\"*variable\",\"value\":\"~*req.8\"},{\"mandatory\":true,\"path\":\"*cgreq.Subject\",\"tag\":\"Subject\",\"type\":\"*variable\",\"value\":\"~*req.9\"},{\"mandatory\":true,\"path\":\"*cgreq.Destination\",\"tag\":\"Destination\",\"type\":\"*variable\",\"value\":\"~*req.10\"},{\"mandatory\":true,\"path\":\"*cgreq.SetupTime\",\"tag\":\"SetupTime\",\"type\":\"*variable\",\"value\":\"~*req.11\"},{\"mandatory\":true,\"path\":\"*cgreq.AnswerTime\",\"tag\":\"AnswerTime\",\"type\":\"*variable\",\"value\":\"~*req.12\"},{\"mandatory\":true,\"path\":\"*cgreq.Usage\",\"tag\":\"Usage\",\"type\":\"*variable\",\"value\":\"~*req.13\"}],\"filters\":[],\"flags\":[],\"id\":\"*default\",\"opts\":{\"csvFieldSeparator\":\",\",\"csvHeaderDefineChar\":\":\",\"csvRowLength\":0,\"natsSubject\":\"cgrates_cdrs\",\"partialCacheAction\":\"*none\",\"partialOrderField\":\"~*req.AnswerTime\",\"xmlRootPath\":\"\"},\"partial_commit_fields\":[],\"processed_path\":\"/var/spool/cgrates/ers/out\",\"run_delay\":\"0\",\"source_path\":\"/var/spool/cgrates/ers/in\",\"tenant\":\"\",\"timezone\":\"\",\"type\":\"*none\"}],\"sessions_conns\":[\"*internal\"]}}"
	var rpl string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.ERsJSON},
	}, &rpl); err != nil {
		t.Error(err)
	} else if cfgStr != rpl {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl))
	}
}

func testSectConfigSReloadEES(t *testing.T) {

	var reply string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1SetConfigFromJSON, &config.SetConfigFromJSONArgs{
		Tenant: "cgrates.org",
		Config: "{\"ees\":{\"attributes_conns\":[],\"cache\":{\"*fileCSV\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"5s\"},\"*file_csv\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"5s\"}},\"enabled\":true,\"exporters\":[{\"attempts\":1,\"attribute_context\":\"\",\"attribute_ids\":[],\"concurrent_requests\":0,\"export_path\":\"/var/spool/cgrates/ees\",\"failed_posts_dir\":\"/var/spool/cgrates/failed_posts\",\"fields\":[],\"filters\":[],\"flags\":[],\"id\":\"*default\",\"opts\":{},\"synchronous\":false,\"timezone\":\"\",\"type\":\"*none\"}]}}",
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := "{\"ees\":{\"attributes_conns\":[],\"cache\":{\"*fileCSV\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"5s\"},\"*file_csv\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"5s\"}},\"enabled\":true,\"exporters\":[{\"attempts\":1,\"attribute_context\":\"\",\"attribute_ids\":[],\"concurrent_requests\":0,\"export_path\":\"/var/spool/cgrates/ees\",\"failed_posts_dir\":\"/var/spool/cgrates/failed_posts\",\"fields\":[],\"filters\":[],\"flags\":[],\"id\":\"*default\",\"opts\":{},\"synchronous\":false,\"timezone\":\"\",\"type\":\"*none\"}]}}"
	var rpl string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.EEsJSON},
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
		Config: "{\"sessions\":{\"accounts_conns\":[],\"actions_conns\":[],\"alterable_fields\":[],\"attributes_conns\":[\"*internal\"],\"cdrs_conns\":[\"*internal\"],\"channel_sync_interval\":\"0\",\"chargers_conns\":[\"*internal\"],\"client_protocol\":1,\"default_usage\":{\"*any\":\"3h0m0s\",\"*data\":\"1048576\",\"*sms\":\"1\",\"*voice\":\"3h0m0s\"},\"enabled\":true,\"listen_bigob\":\"\",\"listen_bijson\":\"127.0.0.1:2014\",\"min_dur_low_balance\":\"0\",\"opts\":{\"*accountS\":[],\"*attributeS\":[],\"*attributesDerivedReply\":[],\"*blockerError\":[],\"*cdrS\":[],\"*cdrsDerivedReply\":[],\"*chargeable\":[],\"*chargerS\":[],\"*debitInterval\":[],\"*forceDuration\":[],\"*initiate\":[],\"*maxUsage\":[],\"*message\":[],\"*resourceS\":[],\"*resourcesAllocate\":[],\"*resourcesAuthorize\":[],\"*resourcesDerivedReply\":[],\"*resourcesRelease\":[],\"*routeS\":[],\"*routesDerivedReply\":[],\"*statS\":[],\"*statsDerivedReply\":[],\"*terminate\":[],\"*thresholdS\":[],\"*thresholdsDerivedReply\":[],\"*ttl\":[],\"*ttlLastUsage\":[],\"*ttlLastUsed\":[],\"*ttlMaxDelay\":[],\"*ttlUsage\":[],\"*update\":[]},\"rates_conns\":[],\"replication_conns\":[],\"resources_conns\":[\"*internal\"],\"routes_conns\":[\"*internal\"],\"session_indexes\":[\"OriginID\"],\"stats_conns\":[],\"stir\":{\"allowed_attest\":[\"*any\"],\"default_attest\":\"A\",\"payload_maxduration\":\"-1\",\"privatekey_path\":\"\",\"publickey_path\":\"\"},\"store_session_costs\":false,\"terminate_attempts\":5,\"thresholds_conns\":[]}}",
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := "{\"sessions\":{\"accounts_conns\":[],\"actions_conns\":[],\"alterable_fields\":[],\"attributes_conns\":[\"*internal\"],\"cdrs_conns\":[\"*internal\"],\"channel_sync_interval\":\"0\",\"chargers_conns\":[\"*internal\"],\"client_protocol\":1,\"default_usage\":{\"*any\":\"3h0m0s\",\"*data\":\"1048576\",\"*sms\":\"1\",\"*voice\":\"3h0m0s\"},\"enabled\":true,\"listen_bigob\":\"\",\"listen_bijson\":\"127.0.0.1:2014\",\"min_dur_low_balance\":\"0\",\"opts\":{\"*accountS\":[],\"*attributeS\":[],\"*attributesDerivedReply\":[],\"*blockerError\":[],\"*cdrS\":[],\"*cdrsDerivedReply\":[],\"*chargeable\":[],\"*chargerS\":[],\"*debitInterval\":[],\"*forceDuration\":[],\"*initiate\":[],\"*maxUsage\":[],\"*message\":[],\"*resourceS\":[],\"*resourcesAllocate\":[],\"*resourcesAuthorize\":[],\"*resourcesDerivedReply\":[],\"*resourcesRelease\":[],\"*routeS\":[],\"*routesDerivedReply\":[],\"*statS\":[],\"*statsDerivedReply\":[],\"*terminate\":[],\"*thresholdS\":[],\"*thresholdsDerivedReply\":[],\"*ttl\":[],\"*ttlLastUsage\":[],\"*ttlLastUsed\":[],\"*ttlMaxDelay\":[],\"*ttlUsage\":[],\"*update\":[]},\"rates_conns\":[],\"replication_conns\":[],\"resources_conns\":[\"*internal\"],\"routes_conns\":[\"*internal\"],\"session_indexes\":[\"OriginID\"],\"stats_conns\":[],\"stir\":{\"allowed_attest\":[\"*any\"],\"default_attest\":\"A\",\"payload_maxduration\":\"-1\",\"privatekey_path\":\"\",\"publickey_path\":\"\"},\"store_session_costs\":false,\"terminate_attempts\":5,\"thresholds_conns\":[]}}"
	var rpl string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.SessionSJSON},
	}, &rpl); err != nil {
		t.Error(err)
	} else if cfgStr != rpl {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl))
	}

}

func testSectConfigSReloadAsteriskAgent(t *testing.T) {

	var reply string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1SetConfigFromJSON, &config.SetConfigFromJSONArgs{
		Tenant: "cgrates.org",
		Config: "{\"asterisk_agent\":{\"asterisk_conns\":[{\"address\":\"127.0.0.1:8088\",\"alias\":\"\",\"connect_attempts\":3,\"password\":\"CGRateS.org\",\"reconnects\":5,\"user\":\"cgrates\"}],\"create_cdr\":false,\"enabled\":false,\"sessions_conns\":[\"*birpc_internal\"]}}",
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := "{\"asterisk_agent\":{\"asterisk_conns\":[{\"address\":\"127.0.0.1:8088\",\"alias\":\"\",\"connect_attempts\":3,\"password\":\"CGRateS.org\",\"reconnects\":5,\"user\":\"cgrates\"}],\"create_cdr\":false,\"enabled\":false,\"sessions_conns\":[\"*birpc_internal\"]}}"
	var rpl string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.AsteriskAgentJSON},
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
		Config: "{\"freeswitch_agent\":{\"create_cdr\":false,\"empty_balance_ann_file\":\"\",\"empty_balance_context\":\"\",\"enabled\":false,\"event_socket_conns\":[{\"address\":\"127.0.0.1:8021\",\"alias\":\"127.0.0.1:8021\",\"password\":\"ClueCon\",\"reconnects\":5}],\"extra_fields\":[],\"low_balance_ann_file\":\"\",\"max_wait_connection\":\"2s\",\"sessions_conns\":[\"*birpc_internal\"],\"subscribe_park\":true}}",
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := "{\"freeswitch_agent\":{\"create_cdr\":false,\"empty_balance_ann_file\":\"\",\"empty_balance_context\":\"\",\"enabled\":false,\"event_socket_conns\":[{\"address\":\"127.0.0.1:8021\",\"alias\":\"127.0.0.1:8021\",\"password\":\"ClueCon\",\"reconnects\":5}],\"extra_fields\":[],\"low_balance_ann_file\":\"\",\"max_wait_connection\":\"2s\",\"sessions_conns\":[\"*birpc_internal\"],\"subscribe_park\":true}}"
	var rpl string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.FreeSWITCHAgentJSON},
	}, &rpl); err != nil {
		t.Error(err)
	} else if cfgStr != rpl {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl))
	}
}

func testSectConfigSReloadKamailioAgent(t *testing.T) {

	var reply string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1SetConfigFromJSON, &config.SetConfigFromJSONArgs{
		Tenant: "cgrates.org",
		Config: "{\"kamailio_agent\":{\"create_cdr\":false,\"enabled\":false,\"evapi_conns\":[{\"address\":\"127.0.0.1:8448\",\"alias\":\"\",\"reconnects\":5}],\"sessions_conns\":[\"*birpc_internal\"],\"timezone\":\"\"}}",
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := "{\"kamailio_agent\":{\"create_cdr\":false,\"enabled\":false,\"evapi_conns\":[{\"address\":\"127.0.0.1:8448\",\"alias\":\"\",\"reconnects\":5}],\"sessions_conns\":[\"*birpc_internal\"],\"timezone\":\"\"}}"
	var rpl string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.KamailioAgentJSON},
	}, &rpl); err != nil {
		t.Error(err)
	} else if cfgStr != rpl {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl))
	}
}

func testSectConfigSReloadDiameterAgent(t *testing.T) {

	var reply string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1SetConfigFromJSON, &config.SetConfigFromJSONArgs{
		Tenant: "cgrates.org",
		Config: "{\"diameter_agent\":{\"asr_template\":\"\",\"concurrent_requests\":-1,\"dictionaries_path\":\"/usr/share/cgrates/diameter/dict/\",\"enabled\":true,\"forced_disconnect\":\"*none\",\"listen\":\"127.0.0.1:3868\",\"listen_net\":\"tcp\",\"origin_host\":\"CGR-DA\",\"origin_realm\":\"cgrates.org\",\"product_name\":\"CGRateS\",\"rar_template\":\"\",\"request_processors\":[],\"sessions_conns\":[\"*birpc_internal\"],\"synced_conn_requests\":false,\"vendor_id\":0}}",
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := "{\"diameter_agent\":{\"asr_template\":\"\",\"concurrent_requests\":-1,\"dictionaries_path\":\"/usr/share/cgrates/diameter/dict/\",\"enabled\":true,\"forced_disconnect\":\"*none\",\"listen\":\"127.0.0.1:3868\",\"listen_net\":\"tcp\",\"origin_host\":\"CGR-DA\",\"origin_realm\":\"cgrates.org\",\"product_name\":\"CGRateS\",\"rar_template\":\"\",\"request_processors\":[],\"sessions_conns\":[\"*birpc_internal\"],\"synced_conn_requests\":false,\"vendor_id\":0}}"
	var rpl string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.DiameterAgentJSON},
	}, &rpl); err != nil {
		t.Error(err)
	} else if cfgStr != rpl {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl))
	}
}

func testSectConfigSReloadHTTPAgent(t *testing.T) {

	var reply string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1SetConfigFromJSON, &config.SetConfigFromJSONArgs{
		Tenant: "cgrates.org",
		Config: "{\"http_agent\":[]}",
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := "{\"http_agent\":[]}"
	var rpl string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.HTTPAgentJSON},
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
		Config: "{\"dns_agent\":{\"enabled\":true,\"listen\":\"127.0.0.1:2053\",\"listen_net\":\"udp\",\"request_processors\":[],\"sessions_conns\":[\"*internal\"],\"timezone\":\"\"}}",
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := "{\"dns_agent\":{\"enabled\":true,\"listen\":\"127.0.0.1:2053\",\"listen_net\":\"udp\",\"request_processors\":[],\"sessions_conns\":[\"*internal\"],\"timezone\":\"\"}}"
	var rpl string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.DNSAgentJSON},
	}, &rpl); err != nil {
		t.Error(err)
	} else if cfgStr != rpl {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl))
	}
}

func testSectConfigSReloadAttributes(t *testing.T) {

	var replyPingBf string
	if err := testSectRPC.Call(context.Background(), utils.AttributeSv1Ping, &utils.CGREvent{}, &replyPingBf); err != nil {
		t.Error(err)
	} else if replyPingBf != utils.Pong {
		t.Errorf("Expected OK received: %s", replyPingBf)
	}

	var reply string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1SetConfigFromJSON, &config.SetConfigFromJSONArgs{
		Tenant: "cgrates.org",
		Config: "{\"attributes\":{\"accounts_conns\":[\"*localhost\"],\"enabled\":true,\"indexed_selects\":true,\"nested_fields\":false,\"opts\":{\"*processRuns\":[],\"*profileIDs\":[],\"*profileIgnoreFilters\":[],\"*profileRuns\":[]},\"prefix_indexed_fields\":[],\"resources_conns\":[\"*localhost\"],\"stats_conns\":[\"*localhost\"],\"suffix_indexed_fields\":[]}}",
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := "{\"attributes\":{\"accounts_conns\":[\"*localhost\"],\"enabled\":true,\"indexed_selects\":true,\"nested_fields\":false,\"opts\":{\"*processRuns\":[],\"*profileIDs\":[],\"*profileIgnoreFilters\":[],\"*profileRuns\":[]},\"prefix_indexed_fields\":[],\"resources_conns\":[\"*localhost\"],\"stats_conns\":[\"*localhost\"],\"suffix_indexed_fields\":[]}}"

	var rpl string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.AttributeSJSON},
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

	var replyPingBf string
	if err := testSectRPC.Call(context.Background(), utils.ChargerSv1Ping, &utils.CGREvent{}, &replyPingBf); err != nil {
		t.Error(err)
	} else if replyPingBf != utils.Pong {
		t.Errorf("Expected OK received: %s", replyPingBf)
	}

	var reply string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1SetConfigFromJSON, &config.SetConfigFromJSONArgs{
		Tenant: "cgrates.org",
		Config: "{\"chargers\":{\"attributes_conns\":[\"*internal\"],\"enabled\":true,\"indexed_selects\":true,\"nested_fields\":false,\"prefix_indexed_fields\":[],\"suffix_indexed_fields\":[]}}",
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := "{\"chargers\":{\"attributes_conns\":[\"*internal\"],\"enabled\":true,\"indexed_selects\":true,\"nested_fields\":false,\"prefix_indexed_fields\":[],\"suffix_indexed_fields\":[]}}"
	var rpl string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.ChargerSJSON},
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

	var replyPingBf string
	if err := testSectRPC.Call(context.Background(), utils.ResourceSv1Ping, &utils.CGREvent{}, &replyPingBf); err != nil {
		t.Error(err)
	} else if replyPingBf != utils.Pong {
		t.Errorf("Expected OK received: %s", replyPingBf)
	}

	var reply string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1SetConfigFromJSON, &config.SetConfigFromJSONArgs{
		Tenant: "cgrates.org",
		Config: "{\"resources\":{\"enabled\":true,\"indexed_selects\":true,\"nested_fields\":false,\"opts\":{\"*units\":[],\"*usageID\":[],\"*usageTTL\":[]},\"prefix_indexed_fields\":[],\"store_interval\":\"-1ns\",\"suffix_indexed_fields\":[],\"thresholds_conns\":[\"*internal\"]}}",
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := "{\"resources\":{\"enabled\":true,\"indexed_selects\":true,\"nested_fields\":false,\"opts\":{\"*units\":[],\"*usageID\":[],\"*usageTTL\":[]},\"prefix_indexed_fields\":[],\"store_interval\":\"-1ns\",\"suffix_indexed_fields\":[],\"thresholds_conns\":[\"*internal\"]}}"
	var rpl string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.ResourceSJSON},
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

	var replyPingBf string
	if err := testSectRPC.Call(context.Background(), utils.StatSv1Ping, &utils.CGREvent{}, &replyPingBf); err != nil {
		t.Error(err)
	} else if replyPingBf != utils.Pong {
		t.Errorf("Expected OK received: %s", replyPingBf)
	}

	var reply string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1SetConfigFromJSON, &config.SetConfigFromJSONArgs{
		Tenant: "cgrates.org",
		Config: "{\"stats\":{\"enabled\":true,\"indexed_selects\":true,\"nested_fields\":false,\"opts\":{\"*profileIDs\":[],\"*profileIgnoreFilters\":[],\"*roundingDecimals\":[]},\"prefix_indexed_fields\":[],\"store_interval\":\"-1ns\",\"store_uncompressed_limit\":0,\"suffix_indexed_fields\":[],\"thresholds_conns\":[\"*internal\"]}}",
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := "{\"stats\":{\"enabled\":true,\"indexed_selects\":true,\"nested_fields\":false,\"opts\":{\"*profileIDs\":[],\"*profileIgnoreFilters\":[],\"*roundingDecimals\":[]},\"prefix_indexed_fields\":[],\"store_interval\":\"-1ns\",\"store_uncompressed_limit\":0,\"suffix_indexed_fields\":[],\"thresholds_conns\":[\"*internal\"]}}"
	var rpl string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.StatSJSON},
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

	var replyPingBf string
	if err := testSectRPC.Call(context.Background(), utils.ThresholdSv1Ping, &utils.CGREvent{}, &replyPingBf); err != nil {
		t.Error(err)
	} else if replyPingBf != utils.Pong {
		t.Errorf("Expected OK received: %s", replyPingBf)
	}

	var reply string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1SetConfigFromJSON, &config.SetConfigFromJSONArgs{
		Tenant: "cgrates.org",
		Config: "{\"thresholds\":{\"actions_conns\":[],\"enabled\":true,\"indexed_selects\":true,\"nested_fields\":false,\"opts\":{\"*profileIDs\":[],\"*profileIgnoreFilters\":[]},\"prefix_indexed_fields\":[],\"store_interval\":\"-1ns\",\"suffix_indexed_fields\":[]}}",
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := "{\"thresholds\":{\"actions_conns\":[],\"enabled\":true,\"indexed_selects\":true,\"nested_fields\":false,\"opts\":{\"*profileIDs\":[],\"*profileIgnoreFilters\":[]},\"prefix_indexed_fields\":[],\"store_interval\":\"-1ns\",\"suffix_indexed_fields\":[]}}"
	var rpl string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.ThresholdSJSON},
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

	var replyPingBf string
	if err := testSectRPC.Call(context.Background(), utils.RouteSv1Ping, &utils.CGREvent{}, &replyPingBf); err != nil {
		t.Error(err)
	} else if replyPingBf != utils.Pong {
		t.Errorf("Expected OK received: %s", replyPingBf)
	}

	var reply string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1SetConfigFromJSON, &config.SetConfigFromJSONArgs{
		Tenant: "cgrates.org",
		Config: "{\"routes\":{\"accounts_conns\":[],\"attributes_conns\":[],\"default_ratio\":1,\"enabled\":true,\"indexed_selects\":true,\"nested_fields\":false,\"opts\":{\"*context\":[],\"*ignoreErrors\":[],\"*limit\":[],\"*maxCost\":[],\"*offset\":[],\"*profileCount\":[],\"*usage\":[]},\"prefix_indexed_fields\":[\"*req.Destination\"],\"rates_conns\":[],\"resources_conns\":[\"*internal\"],\"stats_conns\":[\"*internal\"],\"suffix_indexed_fields\":[]}}",
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := "{\"routes\":{\"accounts_conns\":[],\"attributes_conns\":[],\"default_ratio\":1,\"enabled\":true,\"indexed_selects\":true,\"nested_fields\":false,\"opts\":{\"*context\":[],\"*ignoreErrors\":[],\"*limit\":[],\"*maxCost\":[],\"*offset\":[],\"*profileCount\":[],\"*usage\":[]},\"prefix_indexed_fields\":[\"*req.Destination\"],\"rates_conns\":[],\"resources_conns\":[\"*internal\"],\"stats_conns\":[\"*internal\"],\"suffix_indexed_fields\":[]}}"
	var rpl string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.RouteSJSON},
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
		Section: config.LoaderSJSON,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := "{\"loaders\":[{\"action\":\"*store\",\"cache\":{\"*accounts\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"5s\"},\"*action_profiles\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"5s\"},\"*attributes\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"5s\"},\"*chargers\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"5s\"},\"*dispatcher_hosts\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"5s\"},\"*dispatchers\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"5s\"},\"*filters\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"5s\"},\"*rate_profiles\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"5s\"},\"*resources\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"5s\"},\"*routes\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"5s\"},\"*stats\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"5s\"},\"*thresholds\":{\"limit\":-1,\"precache\":false,\"replicate\":false,\"static_ttl\":false,\"ttl\":\"5s\"}},\"caches_conns\":[\"*internal\"],\"data\":[{\"fields\":[{\"mandatory\":true,\"path\":\"Tenant\",\"tag\":\"Tenant\",\"type\":\"*variable\",\"value\":\"~*req.0\"},{\"mandatory\":true,\"path\":\"ID\",\"tag\":\"ID\",\"type\":\"*variable\",\"value\":\"~*req.1\"},{\"new_branch\":true,\"path\":\"Rules.Type\",\"tag\":\"Type\",\"type\":\"*variable\",\"value\":\"~*req.2\"},{\"path\":\"Rules.Element\",\"tag\":\"Element\",\"type\":\"*variable\",\"value\":\"~*req.3\"},{\"path\":\"Rules.Values\",\"tag\":\"Values\",\"type\":\"*variable\",\"value\":\"~*req.4\"}],\"file_name\":\"Filters.csv\",\"flags\":null,\"type\":\"*filters\"},{\"fields\":[{\"mandatory\":true,\"path\":\"Tenant\",\"tag\":\"TenantID\",\"type\":\"*variable\",\"value\":\"~*req.0\"},{\"mandatory\":true,\"path\":\"ID\",\"tag\":\"ProfileID\",\"type\":\"*variable\",\"value\":\"~*req.1\"},{\"path\":\"FilterIDs\",\"tag\":\"FilterIDs\",\"type\":\"*variable\",\"value\":\"~*req.2\"},{\"path\":\"Weight\",\"tag\":\"Weight\",\"type\":\"*variable\",\"value\":\"~*req.3\"},{\"new_branch\":true,\"path\":\"Attributes.FilterIDs\",\"tag\":\"AttributeFilterIDs\",\"type\":\"*variable\",\"value\":\"~*req.4\"},{\"path\":\"Attributes.Path\",\"tag\":\"Path\",\"type\":\"*variable\",\"value\":\"~*req.5\"},{\"path\":\"Attributes.Type\",\"tag\":\"Type\",\"type\":\"*variable\",\"value\":\"~*req.6\"},{\"path\":\"Attributes.Value\",\"tag\":\"Value\",\"type\":\"*variable\",\"value\":\"~*req.7\"},{\"path\":\"Blocker\",\"tag\":\"Blocker\",\"type\":\"*variable\",\"value\":\"~*req.8\"}],\"file_name\":\"Attributes.csv\",\"flags\":null,\"type\":\"*attributes\"},{\"fields\":[{\"mandatory\":true,\"path\":\"Tenant\",\"tag\":\"Tenant\",\"type\":\"*variable\",\"value\":\"~*req.0\"},{\"mandatory\":true,\"path\":\"ID\",\"tag\":\"ID\",\"type\":\"*variable\",\"value\":\"~*req.1\"},{\"path\":\"FilterIDs\",\"tag\":\"FilterIDs\",\"type\":\"*variable\",\"value\":\"~*req.2\"},{\"path\":\"Weight\",\"tag\":\"Weight\",\"type\":\"*variable\",\"value\":\"~*req.3\"},{\"path\":\"UsageTTL\",\"tag\":\"TTL\",\"type\":\"*variable\",\"value\":\"~*req.4\"},{\"path\":\"Limit\",\"tag\":\"Limit\",\"type\":\"*variable\",\"value\":\"~*req.5\"},{\"path\":\"AllocationMessage\",\"tag\":\"AllocationMessage\",\"type\":\"*variable\",\"value\":\"~*req.6\"},{\"path\":\"Blocker\",\"tag\":\"Blocker\",\"type\":\"*variable\",\"value\":\"~*req.7\"},{\"path\":\"Stored\",\"tag\":\"Stored\",\"type\":\"*variable\",\"value\":\"~*req.8\"},{\"path\":\"ThresholdIDs\",\"tag\":\"ThresholdIDs\",\"type\":\"*variable\",\"value\":\"~*req.9\"}],\"file_name\":\"Resources.csv\",\"flags\":null,\"type\":\"*resources\"},{\"fields\":[{\"mandatory\":true,\"path\":\"Tenant\",\"tag\":\"Tenant\",\"type\":\"*variable\",\"value\":\"~*req.0\"},{\"mandatory\":true,\"path\":\"ID\",\"tag\":\"ID\",\"type\":\"*variable\",\"value\":\"~*req.1\"},{\"path\":\"FilterIDs\",\"tag\":\"FilterIDs\",\"type\":\"*variable\",\"value\":\"~*req.2\"},{\"path\":\"Weight\",\"tag\":\"Weight\",\"type\":\"*variable\",\"value\":\"~*req.3\"},{\"path\":\"QueueLength\",\"tag\":\"QueueLength\",\"type\":\"*variable\",\"value\":\"~*req.4\"},{\"path\":\"TTL\",\"tag\":\"TTL\",\"type\":\"*variable\",\"value\":\"~*req.5\"},{\"path\":\"MinItems\",\"tag\":\"MinItems\",\"type\":\"*variable\",\"value\":\"~*req.6\"},{\"new_branch\":true,\"path\":\"Metrics.MetricID\",\"tag\":\"MetricIDs\",\"type\":\"*variable\",\"value\":\"~*req.7\"},{\"path\":\"Metrics.FilterIDs\",\"tag\":\"MetricFilterIDs\",\"type\":\"*variable\",\"value\":\"~*req.8\"},{\"path\":\"Blocker\",\"tag\":\"Blocker\",\"type\":\"*variable\",\"value\":\"~*req.9\"},{\"path\":\"Stored\",\"tag\":\"Stored\",\"type\":\"*variable\",\"value\":\"~*req.10\"},{\"path\":\"ThresholdIDs\",\"tag\":\"ThresholdIDs\",\"type\":\"*variable\",\"value\":\"~*req.11\"}],\"file_name\":\"Stats.csv\",\"flags\":null,\"type\":\"*stats\"},{\"fields\":[{\"mandatory\":true,\"path\":\"Tenant\",\"tag\":\"Tenant\",\"type\":\"*variable\",\"value\":\"~*req.0\"},{\"mandatory\":true,\"path\":\"ID\",\"tag\":\"ID\",\"type\":\"*variable\",\"value\":\"~*req.1\"},{\"path\":\"FilterIDs\",\"tag\":\"FilterIDs\",\"type\":\"*variable\",\"value\":\"~*req.2\"},{\"path\":\"Weight\",\"tag\":\"Weight\",\"type\":\"*variable\",\"value\":\"~*req.3\"},{\"path\":\"MaxHits\",\"tag\":\"MaxHits\",\"type\":\"*variable\",\"value\":\"~*req.4\"},{\"path\":\"MinHits\",\"tag\":\"MinHits\",\"type\":\"*variable\",\"value\":\"~*req.5\"},{\"path\":\"MinSleep\",\"tag\":\"MinSleep\",\"type\":\"*variable\",\"value\":\"~*req.6\"},{\"path\":\"Blocker\",\"tag\":\"Blocker\",\"type\":\"*variable\",\"value\":\"~*req.7\"},{\"path\":\"ActionProfileIDs\",\"tag\":\"ActionProfileIDs\",\"type\":\"*variable\",\"value\":\"~*req.8\"},{\"path\":\"Async\",\"tag\":\"Async\",\"type\":\"*variable\",\"value\":\"~*req.9\"}],\"file_name\":\"Thresholds.csv\",\"flags\":null,\"type\":\"*thresholds\"},{\"fields\":[{\"mandatory\":true,\"path\":\"Tenant\",\"tag\":\"Tenant\",\"type\":\"*variable\",\"value\":\"~*req.0\"},{\"mandatory\":true,\"path\":\"ID\",\"tag\":\"ID\",\"type\":\"*variable\",\"value\":\"~*req.1\"},{\"path\":\"FilterIDs\",\"tag\":\"FilterIDs\",\"type\":\"*variable\",\"value\":\"~*req.2\"},{\"path\":\"Weights\",\"tag\":\"Weights\",\"type\":\"*variable\",\"value\":\"~*req.3\"},{\"path\":\"Sorting\",\"tag\":\"Sorting\",\"type\":\"*variable\",\"value\":\"~*req.4\"},{\"path\":\"SortingParameters\",\"tag\":\"SortingParameters\",\"type\":\"*variable\",\"value\":\"~*req.5\"},{\"new_branch\":true,\"path\":\"Routes.ID\",\"tag\":\"RouteID\",\"type\":\"*variable\",\"value\":\"~*req.6\"},{\"path\":\"Routes.FilterIDs\",\"tag\":\"RouteFilterIDs\",\"type\":\"*variable\",\"value\":\"~*req.7\"},{\"path\":\"Routes.AccountIDs\",\"tag\":\"RouteAccountIDs\",\"type\":\"*variable\",\"value\":\"~*req.8\"},{\"path\":\"Routes.RateProfileIDs\",\"tag\":\"RouteRateProfileIDs\",\"type\":\"*variable\",\"value\":\"~*req.9\"},{\"path\":\"Routes.ResourceIDs\",\"tag\":\"RouteResourceIDs\",\"type\":\"*variable\",\"value\":\"~*req.10\"},{\"path\":\"Routes.StatIDs\",\"tag\":\"RouteStatIDs\",\"type\":\"*variable\",\"value\":\"~*req.11\"},{\"path\":\"Routes.Weights\",\"tag\":\"RouteWeights\",\"type\":\"*variable\",\"value\":\"~*req.12\"},{\"path\":\"Routes.Blocker\",\"tag\":\"RouteBlocker\",\"type\":\"*variable\",\"value\":\"~*req.13\"},{\"path\":\"Routes.RouteParameters\",\"tag\":\"RouteParameters\",\"type\":\"*variable\",\"value\":\"~*req.14\"}],\"file_name\":\"Routes.csv\",\"flags\":null,\"type\":\"*routes\"},{\"fields\":[{\"mandatory\":true,\"path\":\"Tenant\",\"tag\":\"Tenant\",\"type\":\"*variable\",\"value\":\"~*req.0\"},{\"mandatory\":true,\"path\":\"ID\",\"tag\":\"ID\",\"type\":\"*variable\",\"value\":\"~*req.1\"},{\"path\":\"FilterIDs\",\"tag\":\"FilterIDs\",\"type\":\"*variable\",\"value\":\"~*req.2\"},{\"path\":\"Weight\",\"tag\":\"Weight\",\"type\":\"*variable\",\"value\":\"~*req.3\"},{\"path\":\"RunID\",\"tag\":\"RunID\",\"type\":\"*variable\",\"value\":\"~*req.4\"},{\"path\":\"AttributeIDs\",\"tag\":\"AttributeIDs\",\"type\":\"*variable\",\"value\":\"~*req.5\"}],\"file_name\":\"Chargers.csv\",\"flags\":null,\"type\":\"*chargers\"},{\"fields\":[{\"mandatory\":true,\"path\":\"Tenant\",\"tag\":\"Tenant\",\"type\":\"*variable\",\"value\":\"~*req.0\"},{\"mandatory\":true,\"path\":\"ID\",\"tag\":\"ID\",\"type\":\"*variable\",\"value\":\"~*req.1\"},{\"path\":\"FilterIDs\",\"tag\":\"FilterIDs\",\"type\":\"*variable\",\"value\":\"~*req.2\"},{\"path\":\"Weight\",\"tag\":\"Weight\",\"type\":\"*variable\",\"value\":\"~*req.3\"},{\"path\":\"Strategy\",\"tag\":\"Strategy\",\"type\":\"*variable\",\"value\":\"~*req.4\"},{\"path\":\"StrategyParams\",\"tag\":\"StrategyParameters\",\"type\":\"*variable\",\"value\":\"~*req.5\"},{\"new_branch\":true,\"path\":\"Hosts.ID\",\"tag\":\"ConnID\",\"type\":\"*variable\",\"value\":\"~*req.6\"},{\"path\":\"Hosts.FilterIDs\",\"tag\":\"ConnFilterIDs\",\"type\":\"*variable\",\"value\":\"~*req.7\"},{\"path\":\"Hosts.Weight\",\"tag\":\"ConnWeight\",\"type\":\"*variable\",\"value\":\"~*req.8\"},{\"path\":\"Hosts.Blocker\",\"tag\":\"ConnBlocker\",\"type\":\"*variable\",\"value\":\"~*req.9\"},{\"path\":\"Hosts.Params\",\"tag\":\"ConnParameters\",\"type\":\"*variable\",\"value\":\"~*req.10\"}],\"file_name\":\"DispatcherProfiles.csv\",\"flags\":null,\"type\":\"*dispatchers\"},{\"fields\":[{\"mandatory\":true,\"path\":\"Tenant\",\"tag\":\"Tenant\",\"type\":\"*variable\",\"value\":\"~*req.0\"},{\"mandatory\":true,\"path\":\"ID\",\"tag\":\"ID\",\"type\":\"*variable\",\"value\":\"~*req.1\"},{\"path\":\"Address\",\"tag\":\"Address\",\"type\":\"*variable\",\"value\":\"~*req.2\"},{\"path\":\"Transport\",\"tag\":\"Transport\",\"type\":\"*variable\",\"value\":\"~*req.3\"},{\"path\":\"ConnectAttempts\",\"tag\":\"ConnectAttempts\",\"type\":\"*variable\",\"value\":\"~*req.4\"},{\"path\":\"Reconnects\",\"tag\":\"Reconnects\",\"type\":\"*variable\",\"value\":\"~*req.5\"},{\"path\":\"ConnectTimeout\",\"tag\":\"ConnectTimeout\",\"type\":\"*variable\",\"value\":\"~*req.6\"},{\"path\":\"ReplyTimeout\",\"tag\":\"ReplyTimeout\",\"type\":\"*variable\",\"value\":\"~*req.7\"},{\"path\":\"TLS\",\"tag\":\"TLS\",\"type\":\"*variable\",\"value\":\"~*req.8\"},{\"path\":\"ClientKey\",\"tag\":\"ClientKey\",\"type\":\"*variable\",\"value\":\"~*req.9\"},{\"path\":\"ClientCertificate\",\"tag\":\"ClientCertificate\",\"type\":\"*variable\",\"value\":\"~*req.10\"},{\"path\":\"CaCertificate\",\"tag\":\"CaCertificate\",\"type\":\"*variable\",\"value\":\"~*req.11\"}],\"file_name\":\"DispatcherHosts.csv\",\"flags\":null,\"type\":\"*dispatcher_hosts\"},{\"fields\":[{\"mandatory\":true,\"path\":\"Tenant\",\"tag\":\"Tenant\",\"type\":\"*variable\",\"value\":\"~*req.0\"},{\"mandatory\":true,\"path\":\"ID\",\"tag\":\"ID\",\"type\":\"*variable\",\"value\":\"~*req.1\"},{\"path\":\"FilterIDs\",\"tag\":\"FilterIDs\",\"type\":\"*variable\",\"value\":\"~*req.2\"},{\"path\":\"Weights\",\"tag\":\"Weights\",\"type\":\"*variable\",\"value\":\"~*req.3\"},{\"path\":\"MinCost\",\"tag\":\"MinCost\",\"type\":\"*variable\",\"value\":\"~*req.4\"},{\"path\":\"MaxCost\",\"tag\":\"MaxCost\",\"type\":\"*variable\",\"value\":\"~*req.5\"},{\"path\":\"MaxCostStrategy\",\"tag\":\"MaxCostStrategy\",\"type\":\"*variable\",\"value\":\"~*req.6\"},{\"filters\":[\"*notempty:~*req.7:\"],\"path\":\"Rates[\\u003c~*req.7\\u003e].FilterIDs\",\"tag\":\"RateFilterIDs\",\"type\":\"*variable\",\"value\":\"~*req.8\"},{\"filters\":[\"*notempty:~*req.7:\"],\"path\":\"Rates[\\u003c~*req.7\\u003e].ActivationTimes\",\"tag\":\"RateActivationTimes\",\"type\":\"*variable\",\"value\":\"~*req.9\"},{\"filters\":[\"*notempty:~*req.7:\"],\"path\":\"Rates[\\u003c~*req.7\\u003e].Weights\",\"tag\":\"RateWeights\",\"type\":\"*variable\",\"value\":\"~*req.10\"},{\"filters\":[\"*notempty:~*req.7:\"],\"path\":\"Rates[\\u003c~*req.7\\u003e].Blocker\",\"tag\":\"RateBlocker\",\"type\":\"*variable\",\"value\":\"~*req.11\"},{\"filters\":[\"*notempty:~*req.7:\"],\"new_branch\":true,\"path\":\"Rates[\\u003c~*req.7\\u003e].IntervalRates.IntervalStart\",\"tag\":\"RateIntervalStart\",\"type\":\"*variable\",\"value\":\"~*req.12\"},{\"filters\":[\"*notempty:~*req.7:\"],\"path\":\"Rates[\\u003c~*req.7\\u003e].IntervalRates.FixedFee\",\"tag\":\"RateFixedFee\",\"type\":\"*variable\",\"value\":\"~*req.13\"},{\"filters\":[\"*notempty:~*req.7:\"],\"path\":\"Rates[\\u003c~*req.7\\u003e].IntervalRates.RecurrentFee\",\"tag\":\"RateRecurrentFee\",\"type\":\"*variable\",\"value\":\"~*req.14\"},{\"filters\":[\"*notempty:~*req.7:\"],\"path\":\"Rates[\\u003c~*req.7\\u003e].IntervalRates.Unit\",\"tag\":\"RateUnit\",\"type\":\"*variable\",\"value\":\"~*req.15\"},{\"filters\":[\"*notempty:~*req.7:\"],\"path\":\"Rates[\\u003c~*req.7\\u003e].IntervalRates.Increment\",\"tag\":\"RateIncrement\",\"type\":\"*variable\",\"value\":\"~*req.16\"}],\"file_name\":\"Rates.csv\",\"flags\":null,\"type\":\"*rate_profiles\"},{\"fields\":[{\"mandatory\":true,\"path\":\"Tenant\",\"tag\":\"Tenant\",\"type\":\"*variable\",\"value\":\"~*req.0\"},{\"mandatory\":true,\"path\":\"ID\",\"tag\":\"ID\",\"type\":\"*variable\",\"value\":\"~*req.1\"},{\"path\":\"FilterIDs\",\"tag\":\"FilterIDs\",\"type\":\"*variable\",\"value\":\"~*req.2\"},{\"path\":\"Weight\",\"tag\":\"Weight\",\"type\":\"*variable\",\"value\":\"~*req.3\"},{\"path\":\"Schedule\",\"tag\":\"Schedule\",\"type\":\"*variable\",\"value\":\"~*req.4\"},{\"path\":\"Targets[\\u003c~*req.5\\u003e]\",\"tag\":\"TargetIDs\",\"type\":\"*variable\",\"value\":\"~*req.6\"},{\"filters\":[\"*notempty:~*req.7:\"],\"path\":\"Actions[\\u003c~*req.7\\u003e].FilterIDs\",\"tag\":\"ActionFilterIDs\",\"type\":\"*variable\",\"value\":\"~*req.8\"},{\"filters\":[\"*notempty:~*req.7:\"],\"path\":\"Actions[\\u003c~*req.7\\u003e].Blocker\",\"tag\":\"ActionBlocker\",\"type\":\"*variable\",\"value\":\"~*req.9\"},{\"filters\":[\"*notempty:~*req.7:\"],\"path\":\"Actions[\\u003c~*req.7\\u003e].TTL\",\"tag\":\"ActionTTL\",\"type\":\"*variable\",\"value\":\"~*req.10\"},{\"filters\":[\"*notempty:~*req.7:\"],\"path\":\"Actions[\\u003c~*req.7\\u003e].Type\",\"tag\":\"ActionType\",\"type\":\"*variable\",\"value\":\"~*req.11\"},{\"filters\":[\"*notempty:~*req.7:\"],\"path\":\"Actions[\\u003c~*req.7\\u003e].Opts\",\"tag\":\"ActionOpts\",\"type\":\"*variable\",\"value\":\"~*req.12\"},{\"filters\":[\"*notempty:~*req.7:\"],\"new_branch\":true,\"path\":\"Actions[\\u003c~*req.7\\u003e].Diktats.Path\",\"tag\":\"ActionPath\",\"type\":\"*variable\",\"value\":\"~*req.13\"},{\"filters\":[\"*notempty:~*req.7:\"],\"path\":\"Actions[\\u003c~*req.7\\u003e].Diktats.Value\",\"tag\":\"ActionValue\",\"type\":\"*variable\",\"value\":\"~*req.14\"}],\"file_name\":\"Actions.csv\",\"flags\":null,\"type\":\"*action_profiles\"},{\"fields\":[{\"mandatory\":true,\"path\":\"Tenant\",\"tag\":\"Tenant\",\"type\":\"*variable\",\"value\":\"~*req.0\"},{\"mandatory\":true,\"path\":\"ID\",\"tag\":\"ID\",\"type\":\"*variable\",\"value\":\"~*req.1\"},{\"path\":\"FilterIDs\",\"tag\":\"FilterIDs\",\"type\":\"*variable\",\"value\":\"~*req.2\"},{\"path\":\"Weights\",\"tag\":\"Weights\",\"type\":\"*variable\",\"value\":\"~*req.3\"},{\"path\":\"Opts\",\"tag\":\"Opts\",\"type\":\"*variable\",\"value\":\"~*req.4\"},{\"filters\":[\"*notempty:~*req.5:\"],\"path\":\"Balances[\\u003c~*req.5\\u003e].FilterIDs\",\"tag\":\"BalanceFilterIDs\",\"type\":\"*variable\",\"value\":\"~*req.6\"},{\"filters\":[\"*notempty:~*req.5:\"],\"path\":\"Balances[\\u003c~*req.5\\u003e].Weights\",\"tag\":\"BalanceWeights\",\"type\":\"*variable\",\"value\":\"~*req.7\"},{\"filters\":[\"*notempty:~*req.5:\"],\"path\":\"Balances[\\u003c~*req.5\\u003e].Type\",\"tag\":\"BalanceType\",\"type\":\"*variable\",\"value\":\"~*req.8\"},{\"filters\":[\"*notempty:~*req.5:\"],\"path\":\"Balances[\\u003c~*req.5\\u003e].Units\",\"tag\":\"BalanceUnits\",\"type\":\"*variable\",\"value\":\"~*req.9\"},{\"filters\":[\"*notempty:~*req.5:\"],\"path\":\"Balances[\\u003c~*req.5\\u003e].UnitFactors\",\"tag\":\"BalanceUnitFactors\",\"type\":\"*variable\",\"value\":\"~*req.10\"},{\"filters\":[\"*notempty:~*req.5:\"],\"path\":\"Balances[\\u003c~*req.5\\u003e].Opts\",\"tag\":\"BalanceOpts\",\"type\":\"*variable\",\"value\":\"~*req.11\"},{\"filters\":[\"*notempty:~*req.5:\"],\"path\":\"Balances[\\u003c~*req.5\\u003e].CostIncrements\",\"tag\":\"BalanceCostIncrements\",\"type\":\"*variable\",\"value\":\"~*req.12\"},{\"filters\":[\"*notempty:~*req.5:\"],\"path\":\"Balances[\\u003c~*req.5\\u003e].AttributeIDs\",\"tag\":\"BalanceAttributeIDs\",\"type\":\"*variable\",\"value\":\"~*req.13\"},{\"filters\":[\"*notempty:~*req.5:\"],\"path\":\"Balances[\\u003c~*req.5\\u003e].RateProfileIDs\",\"tag\":\"BalanceRateProfileIDs\",\"type\":\"*variable\",\"value\":\"~*req.14\"},{\"path\":\"ThresholdIDs\",\"tag\":\"ThresholdIDs\",\"type\":\"*variable\",\"value\":\"~*req.15\"}],\"file_name\":\"Accounts.csv\",\"flags\":null,\"type\":\"*accounts\"}],\"enabled\":false,\"field_separator\":\",\",\"id\":\"*default\",\"lockfile_path\":\".cgr.lck\",\"opts\":{\"*cache\":\"\",\"*forceLock\":false,\"*stopOnError\":false,\"*withIndex\":true},\"run_delay\":\"0\",\"tenant\":\"\",\"tp_in_dir\":\"/var/spool/cgrates/loader/in\",\"tp_out_dir\":\"/var/spool/cgrates/loader/out\"}]}"
	var rpl string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.LoaderSJSON},
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
		Config: "{\"suretax\":{\"bill_to_number\":\"\",\"business_unit\":\"\",\"client_number\":\"\",\"client_tracking\":\"~*req.CGRID\",\"customer_number\":\"~*req.Subject\",\"include_local_cost\":false,\"orig_number\":\"~*req.Subject\",\"p2pplus4\":\"\",\"p2pzipcode\":\"\",\"plus4\":\"\",\"regulatory_code\":\"03\",\"response_group\":\"03\",\"response_type\":\"D4\",\"return_file_code\":\"0\",\"sales_type_code\":\"R\",\"tax_exemption_code_list\":\"\",\"tax_included\":\"0\",\"tax_situs_rule\":\"04\",\"term_number\":\"~*req.Destination\",\"timezone\":\"Local\",\"trans_type_code\":\"010101\",\"unit_type\":\"00\",\"units\":\"1\",\"url\":\"\",\"validation_key\":\"\",\"zipcode\":\"\"}}",
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := "{\"suretax\":{\"bill_to_number\":\"\",\"business_unit\":\"\",\"client_number\":\"\",\"client_tracking\":\"~*req.CGRID\",\"customer_number\":\"~*req.Subject\",\"include_local_cost\":false,\"orig_number\":\"~*req.Subject\",\"p2pplus4\":\"\",\"p2pzipcode\":\"\",\"plus4\":\"\",\"regulatory_code\":\"03\",\"response_group\":\"03\",\"response_type\":\"D4\",\"return_file_code\":\"0\",\"sales_type_code\":\"R\",\"tax_exemption_code_list\":\"\",\"tax_included\":\"0\",\"tax_situs_rule\":\"04\",\"term_number\":\"~*req.Destination\",\"timezone\":\"Local\",\"trans_type_code\":\"010101\",\"unit_type\":\"00\",\"units\":\"1\",\"url\":\"\",\"validation_key\":\"\",\"zipcode\":\"\"}}"
	var rpl string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.SureTaxJSON},
	}, &rpl); err != nil {
		t.Error(err)
	} else if cfgStr != rpl {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl))
	}
}

func testSectConfigSReloadLoader(t *testing.T) {

	var reply string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1SetConfigFromJSON, &config.SetConfigFromJSONArgs{
		Tenant: "cgrates.org",
		Config: "{\"loader\":{\"actions_conns\":[\"*localhost\"],\"caches_conns\":[\"*localhost\"],\"data_path\":\"./\",\"disable_reverse\":false,\"field_separator\":\",\",\"gapi_credentials\":\".gapi/credentials.json\",\"gapi_token\":\".gapi/token.json\",\"tpid\":\"\"}}",
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := "{\"loader\":{\"actions_conns\":[\"*localhost\"],\"caches_conns\":[\"*localhost\"],\"data_path\":\"./\",\"disable_reverse\":false,\"field_separator\":\",\",\"gapi_credentials\":\".gapi/credentials.json\",\"gapi_token\":\".gapi/token.json\",\"tpid\":\"\"}}"
	var rpl string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.LoaderJSON},
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
		Config: "{\"migrator\":{\"out_datadb_encoding\":\"msgpack\",\"out_datadb_host\":\"127.0.0.1\",\"out_datadb_name\":\"10\",\"out_datadb_opts\":{\"mongoQueryTimeout\":\"0s\",\"redisCACertificate\":\"\",\"redisClientCertificate\":\"\",\"redisClientKey\":\"\",\"redisCluster\":false,\"redisClusterOndownDelay\":\"0s\",\"redisClusterSync\":\"5s\",\"redisSentinel\":\"\",\"redisTLS\":false},\"out_datadb_password\":\"\",\"out_datadb_port\":\"6379\",\"out_datadb_type\":\"redis\",\"out_datadb_user\":\"cgrates\",\"out_stordb_host\":\"127.0.0.1\",\"out_stordb_name\":\"cgrates\",\"out_stordb_opts\":{\"mongoQueryTimeout\":\"0s\",\"mysqlLocation\":\"\",\"sqlConnMaxLifetime\":\"0s\",\"sqlMaxIdleConns\":0,\"sqlMaxOpenConns\":0,\"sslMode\":\"\"},\"out_stordb_password\":\"CGRateS.org\",\"out_stordb_port\":\"3306\",\"out_stordb_type\":\"mysql\",\"out_stordb_user\":\"cgrates\",\"users_filters\":[\"Account\"]}}",
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := "{\"migrator\":{\"out_datadb_encoding\":\"msgpack\",\"out_datadb_host\":\"127.0.0.1\",\"out_datadb_name\":\"10\",\"out_datadb_opts\":{\"mongoQueryTimeout\":\"0s\",\"redisCACertificate\":\"\",\"redisClientCertificate\":\"\",\"redisClientKey\":\"\",\"redisCluster\":false,\"redisClusterOndownDelay\":\"0s\",\"redisClusterSync\":\"5s\",\"redisSentinel\":\"\",\"redisTLS\":false},\"out_datadb_password\":\"\",\"out_datadb_port\":\"6379\",\"out_datadb_type\":\"redis\",\"out_datadb_user\":\"cgrates\",\"out_stordb_host\":\"127.0.0.1\",\"out_stordb_name\":\"cgrates\",\"out_stordb_opts\":{\"mongoQueryTimeout\":\"0s\",\"mysqlLocation\":\"\",\"sqlConnMaxLifetime\":\"0s\",\"sqlMaxIdleConns\":0,\"sqlMaxOpenConns\":0,\"sslMode\":\"\"},\"out_stordb_password\":\"CGRateS.org\",\"out_stordb_port\":\"3306\",\"out_stordb_type\":\"mysql\",\"out_stordb_user\":\"cgrates\",\"users_filters\":[\"Account\"]}}"
	var rpl string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.MigratorJSON},
	}, &rpl); err != nil {
		t.Error(err)
	} else if cfgStr != rpl {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl))
	}
}

func testSectConfigSReloadDispatchers(t *testing.T) {

	var reply string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1SetConfigFromJSON, &config.SetConfigFromJSONArgs{
		Tenant: "cgrates.org",
		Config: "{\"dispatchers\":{\"attributes_conns\":[],\"enabled\":true,\"indexed_selects\":true,\"nested_fields\":false,\"prefix_indexed_fields\":[],\"suffix_indexed_fields\":[]}}",
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := "{\"dispatchers\":{\"attributes_conns\":[],\"enabled\":true,\"indexed_selects\":true,\"nested_fields\":false,\"prefix_indexed_fields\":[],\"suffix_indexed_fields\":[]}}"
	var rpl string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.DispatcherSJSON},
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
		Config: "{\"registrarc\":{\"dispatchers\":{\"enabled\":true,\"hosts\":[],\"refresh_interval\":\"5m0s\",\"registrars_conns\":[]},\"rpc\":{\"enabled\":true,\"hosts\":[],\"refresh_interval\":\"5m0s\",\"registrars_conns\":[]}}}",
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := "{\"registrarc\":{\"dispatchers\":{\"hosts\":[],\"refresh_interval\":\"5m0s\",\"registrars_conns\":[]},\"rpc\":{\"hosts\":[],\"refresh_interval\":\"5m0s\",\"registrars_conns\":[]}}}"
	var rpl string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.RegistrarCJSON},
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
		Tenant:   "cgrates.org",
		Sections: []string{config.AnalyzerSJSON},
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
		Config: "{\"sip_agent\":{\"enabled\":true,\"listen\":\"127.0.0.1:5060\",\"listen_net\":\"udp\",\"request_processors\":[],\"retransmission_timer\":\"1s\",\"sessions_conns\":[\"*internal\"],\"timezone\":\"\"}}",
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := "{\"sip_agent\":{\"enabled\":true,\"listen\":\"127.0.0.1:5060\",\"listen_net\":\"udp\",\"request_processors\":[],\"retransmission_timer\":\"1s\",\"sessions_conns\":[\"*internal\"],\"timezone\":\"\"}}"
	var rpl string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.SIPAgentJSON},
	}, &rpl); err != nil {
		t.Error(err)
	} else if cfgStr != rpl {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToJSON(rpl))
	}
}

func testSectConfigSReloadTemplates(t *testing.T) {

	var reply string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1SetConfigFromJSON, &config.SetConfigFromJSONArgs{
		Tenant: "cgrates.org",
		Config: "{\"templates\":{\"*asr\":[{\"mandatory\":true,\"path\":\"*diamreq.Session-Id\",\"tag\":\"SessionId\",\"type\":\"*variable\",\"value\":\"~*req.Session-Id\"},{\"mandatory\":true,\"path\":\"*diamreq.Origin-Host\",\"tag\":\"OriginHost\",\"type\":\"*variable\",\"value\":\"~*req.Destination-Host\"},{\"mandatory\":true,\"path\":\"*diamreq.Origin-Realm\",\"tag\":\"OriginRealm\",\"type\":\"*variable\",\"value\":\"~*req.Destination-Realm\"},{\"mandatory\":true,\"path\":\"*diamreq.Destination-Realm\",\"tag\":\"DestinationRealm\",\"type\":\"*variable\",\"value\":\"~*req.Origin-Realm\"},{\"mandatory\":true,\"path\":\"*diamreq.Destination-Host\",\"tag\":\"DestinationHost\",\"type\":\"*variable\",\"value\":\"~*req.Origin-Host\"},{\"mandatory\":true,\"path\":\"*diamreq.Auth-Application-Id\",\"tag\":\"AuthApplicationId\",\"type\":\"*variable\",\"value\":\"~*vars.*appid\"}],\"*cca\":[{\"mandatory\":true,\"path\":\"*rep.Session-Id\",\"tag\":\"SessionId\",\"type\":\"*variable\",\"value\":\"~*req.Session-Id\"},{\"path\":\"*rep.Result-Code\",\"tag\":\"ResultCode\",\"type\":\"*constant\",\"value\":\"2001\"},{\"mandatory\":true,\"path\":\"*rep.Origin-Host\",\"tag\":\"OriginHost\",\"type\":\"*variable\",\"value\":\"~*vars.OriginHost\"},{\"mandatory\":true,\"path\":\"*rep.Origin-Realm\",\"tag\":\"OriginRealm\",\"type\":\"*variable\",\"value\":\"~*vars.OriginRealm\"},{\"mandatory\":true,\"path\":\"*rep.Auth-Application-Id\",\"tag\":\"AuthApplicationId\",\"type\":\"*variable\",\"value\":\"~*vars.*appid\"},{\"mandatory\":true,\"path\":\"*rep.CC-Request-Type\",\"tag\":\"CCRequestType\",\"type\":\"*variable\",\"value\":\"~*req.CC-Request-Type\"},{\"mandatory\":true,\"path\":\"*rep.CC-Request-Number\",\"tag\":\"CCRequestNumber\",\"type\":\"*variable\",\"value\":\"~*req.CC-Request-Number\"}],\"*cdrLog\":[{\"mandatory\":true,\"path\":\"*cdr.ToR\",\"tag\":\"ToR\",\"type\":\"*variable\",\"value\":\"~*req.BalanceType\"},{\"mandatory\":true,\"path\":\"*cdr.OriginHost\",\"tag\":\"OriginHost\",\"type\":\"*constant\",\"value\":\"127.0.0.1\"},{\"mandatory\":true,\"path\":\"*cdr.RequestType\",\"tag\":\"RequestType\",\"type\":\"*constant\",\"value\":\"*none\"},{\"mandatory\":true,\"path\":\"*cdr.Tenant\",\"tag\":\"Tenant\",\"type\":\"*variable\",\"value\":\"~*req.Tenant\"},{\"mandatory\":true,\"path\":\"*cdr.Account\",\"tag\":\"Account\",\"type\":\"*variable\",\"value\":\"~*req.Account\"},{\"mandatory\":true,\"path\":\"*cdr.Subject\",\"tag\":\"Subject\",\"type\":\"*variable\",\"value\":\"~*req.Account\"},{\"mandatory\":true,\"path\":\"*cdr.Cost\",\"tag\":\"Cost\",\"type\":\"*variable\",\"value\":\"~*req.Cost\"},{\"mandatory\":true,\"path\":\"*cdr.Source\",\"tag\":\"Source\",\"type\":\"*constant\",\"value\":\"*cdrLog\"},{\"mandatory\":true,\"path\":\"*cdr.Usage\",\"tag\":\"Usage\",\"type\":\"*constant\",\"value\":\"1\"},{\"mandatory\":true,\"path\":\"*cdr.RunID\",\"tag\":\"RunID\",\"type\":\"*variable\",\"value\":\"~*req.ActionType\"},{\"mandatory\":true,\"path\":\"*cdr.SetupTime\",\"tag\":\"SetupTime\",\"type\":\"*constant\",\"value\":\"*now\"},{\"mandatory\":true,\"path\":\"*cdr.AnswerTime\",\"tag\":\"AnswerTime\",\"type\":\"*constant\",\"value\":\"*now\"},{\"mandatory\":true,\"path\":\"*cdr.PreRated\",\"tag\":\"PreRated\",\"type\":\"*constant\",\"value\":\"true\"}],\"*err\":[{\"mandatory\":true,\"path\":\"*rep.Session-Id\",\"tag\":\"SessionId\",\"type\":\"*variable\",\"value\":\"~*req.Session-Id\"},{\"mandatory\":true,\"path\":\"*rep.Origin-Host\",\"tag\":\"OriginHost\",\"type\":\"*variable\",\"value\":\"~*vars.OriginHost\"},{\"mandatory\":true,\"path\":\"*rep.Origin-Realm\",\"tag\":\"OriginRealm\",\"type\":\"*variable\",\"value\":\"~*vars.OriginRealm\"}],\"*errSip\":[{\"mandatory\":true,\"path\":\"*rep.Request\",\"tag\":\"Request\",\"type\":\"*constant\",\"value\":\"SIP/2.0 500 Internal Server Error\"}],\"*rar\":[{\"mandatory\":true,\"path\":\"*diamreq.Session-Id\",\"tag\":\"SessionId\",\"type\":\"*variable\",\"value\":\"~*req.Session-Id\"},{\"mandatory\":true,\"path\":\"*diamreq.Origin-Host\",\"tag\":\"OriginHost\",\"type\":\"*variable\",\"value\":\"~*req.Destination-Host\"},{\"mandatory\":true,\"path\":\"*diamreq.Origin-Realm\",\"tag\":\"OriginRealm\",\"type\":\"*variable\",\"value\":\"~*req.Destination-Realm\"},{\"mandatory\":true,\"path\":\"*diamreq.Destination-Realm\",\"tag\":\"DestinationRealm\",\"type\":\"*variable\",\"value\":\"~*req.Origin-Realm\"},{\"mandatory\":true,\"path\":\"*diamreq.Destination-Host\",\"tag\":\"DestinationHost\",\"type\":\"*variable\",\"value\":\"~*req.Origin-Host\"},{\"mandatory\":true,\"path\":\"*diamreq.Auth-Application-Id\",\"tag\":\"AuthApplicationId\",\"type\":\"*variable\",\"value\":\"~*vars.*appid\"},{\"path\":\"*diamreq.Re-Auth-Request-Type\",\"tag\":\"ReAuthRequestType\",\"type\":\"*constant\",\"value\":\"0\"}]}}",
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := "{\"templates\":{\"*asr\":[{\"mandatory\":true,\"path\":\"*diamreq.Session-Id\",\"tag\":\"SessionId\",\"type\":\"*variable\",\"value\":\"~*req.Session-Id\"},{\"mandatory\":true,\"path\":\"*diamreq.Origin-Host\",\"tag\":\"OriginHost\",\"type\":\"*variable\",\"value\":\"~*req.Destination-Host\"},{\"mandatory\":true,\"path\":\"*diamreq.Origin-Realm\",\"tag\":\"OriginRealm\",\"type\":\"*variable\",\"value\":\"~*req.Destination-Realm\"},{\"mandatory\":true,\"path\":\"*diamreq.Destination-Realm\",\"tag\":\"DestinationRealm\",\"type\":\"*variable\",\"value\":\"~*req.Origin-Realm\"},{\"mandatory\":true,\"path\":\"*diamreq.Destination-Host\",\"tag\":\"DestinationHost\",\"type\":\"*variable\",\"value\":\"~*req.Origin-Host\"},{\"mandatory\":true,\"path\":\"*diamreq.Auth-Application-Id\",\"tag\":\"AuthApplicationId\",\"type\":\"*variable\",\"value\":\"~*vars.*appid\"}],\"*cca\":[{\"mandatory\":true,\"path\":\"*rep.Session-Id\",\"tag\":\"SessionId\",\"type\":\"*variable\",\"value\":\"~*req.Session-Id\"},{\"path\":\"*rep.Result-Code\",\"tag\":\"ResultCode\",\"type\":\"*constant\",\"value\":\"2001\"},{\"mandatory\":true,\"path\":\"*rep.Origin-Host\",\"tag\":\"OriginHost\",\"type\":\"*variable\",\"value\":\"~*vars.OriginHost\"},{\"mandatory\":true,\"path\":\"*rep.Origin-Realm\",\"tag\":\"OriginRealm\",\"type\":\"*variable\",\"value\":\"~*vars.OriginRealm\"},{\"mandatory\":true,\"path\":\"*rep.Auth-Application-Id\",\"tag\":\"AuthApplicationId\",\"type\":\"*variable\",\"value\":\"~*vars.*appid\"},{\"mandatory\":true,\"path\":\"*rep.CC-Request-Type\",\"tag\":\"CCRequestType\",\"type\":\"*variable\",\"value\":\"~*req.CC-Request-Type\"},{\"mandatory\":true,\"path\":\"*rep.CC-Request-Number\",\"tag\":\"CCRequestNumber\",\"type\":\"*variable\",\"value\":\"~*req.CC-Request-Number\"}],\"*cdrLog\":[{\"mandatory\":true,\"path\":\"*cdr.ToR\",\"tag\":\"ToR\",\"type\":\"*variable\",\"value\":\"~*req.BalanceType\"},{\"mandatory\":true,\"path\":\"*cdr.OriginHost\",\"tag\":\"OriginHost\",\"type\":\"*constant\",\"value\":\"127.0.0.1\"},{\"mandatory\":true,\"path\":\"*cdr.RequestType\",\"tag\":\"RequestType\",\"type\":\"*constant\",\"value\":\"*none\"},{\"mandatory\":true,\"path\":\"*cdr.Tenant\",\"tag\":\"Tenant\",\"type\":\"*variable\",\"value\":\"~*req.Tenant\"},{\"mandatory\":true,\"path\":\"*cdr.Account\",\"tag\":\"Account\",\"type\":\"*variable\",\"value\":\"~*req.Account\"},{\"mandatory\":true,\"path\":\"*cdr.Subject\",\"tag\":\"Subject\",\"type\":\"*variable\",\"value\":\"~*req.Account\"},{\"mandatory\":true,\"path\":\"*cdr.Cost\",\"tag\":\"Cost\",\"type\":\"*variable\",\"value\":\"~*req.Cost\"},{\"mandatory\":true,\"path\":\"*cdr.Source\",\"tag\":\"Source\",\"type\":\"*constant\",\"value\":\"*cdrLog\"},{\"mandatory\":true,\"path\":\"*cdr.Usage\",\"tag\":\"Usage\",\"type\":\"*constant\",\"value\":\"1\"},{\"mandatory\":true,\"path\":\"*cdr.RunID\",\"tag\":\"RunID\",\"type\":\"*variable\",\"value\":\"~*req.ActionType\"},{\"mandatory\":true,\"path\":\"*cdr.SetupTime\",\"tag\":\"SetupTime\",\"type\":\"*constant\",\"value\":\"*now\"},{\"mandatory\":true,\"path\":\"*cdr.AnswerTime\",\"tag\":\"AnswerTime\",\"type\":\"*constant\",\"value\":\"*now\"},{\"mandatory\":true,\"path\":\"*cdr.PreRated\",\"tag\":\"PreRated\",\"type\":\"*constant\",\"value\":\"true\"}],\"*err\":[{\"mandatory\":true,\"path\":\"*rep.Session-Id\",\"tag\":\"SessionId\",\"type\":\"*variable\",\"value\":\"~*req.Session-Id\"},{\"mandatory\":true,\"path\":\"*rep.Origin-Host\",\"tag\":\"OriginHost\",\"type\":\"*variable\",\"value\":\"~*vars.OriginHost\"},{\"mandatory\":true,\"path\":\"*rep.Origin-Realm\",\"tag\":\"OriginRealm\",\"type\":\"*variable\",\"value\":\"~*vars.OriginRealm\"}],\"*errSip\":[{\"mandatory\":true,\"path\":\"*rep.Request\",\"tag\":\"Request\",\"type\":\"*constant\",\"value\":\"SIP/2.0 500 Internal Server Error\"}],\"*rar\":[{\"mandatory\":true,\"path\":\"*diamreq.Session-Id\",\"tag\":\"SessionId\",\"type\":\"*variable\",\"value\":\"~*req.Session-Id\"},{\"mandatory\":true,\"path\":\"*diamreq.Origin-Host\",\"tag\":\"OriginHost\",\"type\":\"*variable\",\"value\":\"~*req.Destination-Host\"},{\"mandatory\":true,\"path\":\"*diamreq.Origin-Realm\",\"tag\":\"OriginRealm\",\"type\":\"*variable\",\"value\":\"~*req.Destination-Realm\"},{\"mandatory\":true,\"path\":\"*diamreq.Destination-Realm\",\"tag\":\"DestinationRealm\",\"type\":\"*variable\",\"value\":\"~*req.Origin-Realm\"},{\"mandatory\":true,\"path\":\"*diamreq.Destination-Host\",\"tag\":\"DestinationHost\",\"type\":\"*variable\",\"value\":\"~*req.Origin-Host\"},{\"mandatory\":true,\"path\":\"*diamreq.Auth-Application-Id\",\"tag\":\"AuthApplicationId\",\"type\":\"*variable\",\"value\":\"~*vars.*appid\"},{\"path\":\"*diamreq.Re-Auth-Request-Type\",\"tag\":\"ReAuthRequestType\",\"type\":\"*constant\",\"value\":\"0\"}]}}"
	var rpl string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.TemplatesJSON},
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
		Config: `{"configs":{"enabled":true,"root_dir":"/var/spool/cgrates/configs","url":"/configs/"}}`,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %+v", reply)
	}
	cfgStr := `{"configs":{"enabled":true,"root_dir":"/var/spool/cgrates/configs","url":"/configs/"}}`
	var rpl string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.ConfigSJSON},
	}, &rpl); err != nil {
		t.Error(err)
	} else if cfgStr != rpl {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl))
	}
}

func testSectConfigSReloadAPIBan(t *testing.T) {
	var replyRld string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1ReloadConfig, &config.ReloadArgs{
		Tenant:  "cgrates.org",
		Section: config.APIBanJSON,
	}, &replyRld); err != nil {
		t.Error(err)
	} else if replyRld != utils.OK {
		t.Errorf("Expected OK received: %+v", replyRld)
	}
	var reply string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1SetConfigFromJSON, &config.SetConfigFromJSONArgs{
		Tenant: "cgrates.org",
		Config: `{"apiban":{"enabled":true,"keys":["testKey"]}}`,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %s", reply)
	}
	cfgStr := "{\"apiban\":{\"enabled\":true,\"keys\":[\"testKey\"]}}"
	var rpl string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.APIBanJSON},
	}, &rpl); err != nil {
		t.Error(err)
	} else if cfgStr != rpl {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl))
	}
}

func testSectConfigSReloadActions(t *testing.T) {
	var replyRld string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1ReloadConfig, &config.ReloadArgs{
		Tenant:  "cgrates.org",
		Section: config.ActionSJSON,
	}, &replyRld); err != nil {
		t.Error(err)
	} else if replyRld != utils.OK {
		t.Errorf("Expected OK received: %+v", replyRld)
	}
	var reply string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1SetConfigFromJSON, &config.SetConfigFromJSONArgs{
		Tenant: "cgrates.org",
		Config: `"actions": {
			"enabled": false,
			"cdrs_conns": [],
			"ees_conns": [],
			"thresholds_conns": [],
			"stats_conns": [],
			"accounts_conns": [],
			"tenants": [],
			"indexed_selects": true,
			//"string_indexed_fields": [],
			"prefix_indexed_fields": [],
			"suffix_indexed_fields": [],
			"nested_fields": false,
			"dynaprepaid_actionprofile": [],
		},`,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %s", reply)
	}
	cfgStr := `"actions": {
		"enabled": false,
		"cdrs_conns": [],
		"ees_conns": [],
		"thresholds_conns": [],
		"stats_conns": [],
		"accounts_conns": [],
		"tenants": [],
		"indexed_selects": true,
		//"string_indexed_fields": [],
		"prefix_indexed_fields": [],
		"suffix_indexed_fields": [],
		"nested_fields": false,
		"dynaprepaid_actionprofile": [],
	},`
	var rpl string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.APIBanJSON},
	}, &rpl); err != nil {
		t.Error(err)
	} else if cfgStr != rpl {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl))
	}
}

func testSectConfigSReloadAccounts(t *testing.T) {
	var replyRld string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1ReloadConfig, &config.ReloadArgs{
		Tenant:  "cgrates.org",
		Section: config.AccountSJSON,
	}, &replyRld); err != nil {
		t.Error(err)
	} else if replyRld != utils.OK {
		t.Errorf("Expected OK received: %+v", replyRld)
	}
	var reply string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1SetConfigFromJSON, &config.SetConfigFromJSONArgs{
		Tenant: "cgrates.org",
		Config: `"accounts": {
			"enabled": false,
			"indexed_selects": true,
			"attributes_conns": [],
			"rates_conns": [],
			"thresholds_conns": [],
			//"string_indexed_fields": [],
			"prefix_indexed_fields": [],
			"suffix_indexed_fields": [],
			"nested_fields": false,
			"max_iterations": 1000,
			"max_usage": "72h",
		},`,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %s", reply)
	}
	cfgStr := `"accounts": {
		"enabled": false,
		"indexed_selects": true,
		"attributes_conns": [],
		"rates_conns": [],
		"thresholds_conns": [],
		//"string_indexed_fields": [],
		"prefix_indexed_fields": [],
		"suffix_indexed_fields": [],
		"nested_fields": false,
		"max_iterations": 1000,
		"max_usage": "72h",
	},`
	var rpl string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.AccountSJSON},
	}, &rpl); err != nil {
		t.Error(err)
	} else if cfgStr != rpl {
		t.Errorf("\nExpected %+v ,\n received: %+v", utils.ToIJSON(cfgStr), utils.ToIJSON(rpl))
	}
}

func testSectConfigSReloadConfigDB(t *testing.T) {
	var replyRld string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1ReloadConfig, &config.ReloadArgs{
		Tenant:  "cgrates.org",
		Section: config.ConfigDBJSON,
	}, &replyRld); err != nil {
		t.Error(err)
	} else if replyRld != utils.OK {
		t.Errorf("Expected OK received: %+v", replyRld)
	}
	var reply string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1SetConfigFromJSON, &config.SetConfigFromJSONArgs{
		Tenant: "cgrates.org",
		Config: `"config_db": {
			"db_type": "*internal",
			"db_host": "",
			"db_port": 0,
			"db_name": "",
			"db_user": "",
			"db_password": "",
			"opts": {
				"redisSentinel": "",
				"redisCluster": false,
				"redisClusterSync": "5s",
				"redisClusterOndownDelay": "0",
				"mongoQueryTimeout": "10s",
				"redisTLS": false, tion
				"redisClientCertificate": "",
				"redisClientKey": "",
				"redisCACertificate": "",
			}
		},`,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK received: %s", reply)
	}
	cfgStr := `"config_db": {
		"db_type": "*internal",
		"db_host": "",
		"db_port": 0,
		"db_name": "",
		"db_user": "",
		"db_password": "",
		"opts": {
			"redisSentinel": "",
			"redisCluster": false,
			"redisClusterSync": "5s",
			"redisClusterOndownDelay": "0",
			"mongoQueryTimeout": "10s",
			"redisTLS": false, tion
			"redisClientCertificate": "",
			"redisClientKey": "",
			"redisCACertificate": "",
		}
	},`
	var rpl string
	if err := testSectRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.ConfigDBJSON},
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
