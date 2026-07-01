//go:build flaky

/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
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
	case utils.MetaMySQL, utils.MetaMongo, utils.MetaRedis, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, testRld := range testTestsR {
		t.Run(testCfgDir, testRld)
		if t.Failed() {
			break
		}
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
	if err := engine.InitDB(testCfg); err != nil {
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

	cfgStr := `{"cores":{"caps":0,"capsStatsInterval":"0","capsStrategy":"*busy","conns":{},"shutdownTimeout":"1s"}}`
	var rpl2 string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.CoreSJSON},
	}, &rpl2); err != nil {
		t.Error(err)
	} else if cfgStr != rpl2 {
		t.Errorf("\nExpected %+v ,\n received: %+v", cfgStr, rpl2)
	}

	cfgStr = `{"rpcConns":{"*bijsonLocalhost":{"conns":[{"address":"127.0.0.1:2014","transport":"*birpc_json"}],"poolSize":0,"strategy":"*first"},"*birpc_internal":{"conns":[{"address":"*birpc_internal","transport":""}],"poolSize":0,"strategy":"*first"},"*internal":{"conns":[{"address":"*internal","transport":""}],"poolSize":0,"strategy":"*first"},"*localhost":{"conns":[{"address":"127.0.0.1:2012","transport":"*json"}],"poolSize":0,"strategy":"*first"}}}`
	var rpl string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.RPCConnsJSON},
	}, &rpl); err != nil {
		t.Error(err)
	} else if cfgStr != rpl {
		t.Errorf("\nExpected %+v ,\n received: %+v", cfgStr, rpl)
	}

	cfgStr = `{"listen":{"http":":2080","httpTLS":"127.0.0.1:2280","rpcGOB":":2013","rpcGOBtls":"127.0.0.1:2023","rpcJSON":":2012","rpcJSONtls":"127.0.0.1:2022"}}`
	var rpl3 string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.ListenJSON},
	}, &rpl3); err != nil {
		t.Error(err)
	} else if cfgStr != rpl3 {
		t.Errorf("\nExpected %+v ,\n received: %+v", cfgStr, rpl3)
	}

	cfgStr = `{"tls":{"caCertificate":"","clientCertificate":"","clientKey":"","serverCertificate":"","serverKey":"","serverName":"","serverPolicy":4}}`
	var rpl4 string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.TlsJSON},
	}, &rpl4); err != nil {
		t.Error(err)
	} else if cfgStr != rpl4 {
		t.Errorf("\nExpected %+v ,\n received: %+v", cfgStr, rpl4)
	}

	cfgStr = `{"http":{"authUsers":{},"clientOpts":{"dialFallbackDelay":"300ms","dialKeepAlive":"30s","dialTimeout":"30s","disableCompression":false,"disableKeepAlives":false,"expectContinueTimeout":"0s","forceAttemptHttp2":true,"idleConnTimeout":"1m30s","maxConnsPerHost":0,"maxIdleConns":100,"maxIdleConnsPerHost":2,"responseHeaderTimeout":"0s","skipTLSVerification":false,"tlsHandshakeTimeout":"10s"},"freeswitchCDRsURL":"/freeswitch_json","httpCDRs":"/cdr_http","jsonRPCurl":"/jsonrpc","pprofPath":"/debug/pprof/","registrarsURL":"/registrar","useBasicAuth":false,"wsURL":"/ws"}}`

	var rpl5 string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.HTTPJSON},
	}, &rpl5); err != nil {
		t.Error(err)
	} else if cfgStr != rpl5 {
		t.Errorf("\nExpected %+v ,\n received: %+v", cfgStr, rpl5)
	}
	if testCfgDir == "tutredis" || testCfgDir == "tutmysql" || testCfgDir == "tutmongo" {
		cfgStr = `{"caches":{"partitions":{"*accountFilterIndexes":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false},"*accounts":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false},"*actionProfileFilterIndexes":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false},"*actionProfiles":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false},"*apiban":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false,"ttl":"2m0s"},"*attributeFilterIndexes":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false},"*attributeProfiles":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false},"*capsEvents":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false},"*cdrIDs":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false,"ttl":"10m0s"},"*chargerFilterIndexes":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false},"*chargerProfiles":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false},"*closedSessions":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false,"ttl":"10s"},"*diameterMessages":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false,"ttl":"3h0m0s"},"*eventCharges":{"limit":0,"precache":false,"remote":false,"replicate":false,"staticTTL":false,"ttl":"10s"},"*eventResources":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false},"*filters":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false},"*loadIDs":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false},"*rankingProfiles":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false},"*rankings":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false},"*rateFilterIndexes":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false},"*rateProfileFilterIndexes":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false},"*rateProfiles":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false},"*replicationHosts":{"limit":0,"precache":false,"remote":false,"replicate":false,"staticTTL":false},"*resourceFilterIndexes":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false},"*resourceProfiles":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false},"*resources":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false},"*reverseFilterIndexes":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false},"*routeFilterIndexes":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false},"*routeProfiles":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false},"*rpcConnections":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false},"*rpcResponses":{"limit":0,"precache":false,"remote":false,"replicate":false,"staticTTL":false,"ttl":"2s"},"*sentrypeer":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":true,"ttl":"24h0m0s"},"*statFilterIndexes":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false},"*statQueueProfiles":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false},"*statQueues":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false},"*stir":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false,"ttl":"3h0m0s"},"*thresholdFilterIndexes":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false},"*thresholdProfiles":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false},"*thresholds":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false},"*trendProfiles":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false},"*trends":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false},"*uch":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false,"ttl":"3h0m0s"}},"remoteConns":[],"replicationConns":[]}}`
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
		cfgStr := `{"caches":{"partitions":{"*accountFilterIndexes":{"limit":0,"precache":false,"remote":false,"replicate":false,"staticTTL":false},"*accounts":{"limit":0,"precache":false,"remote":false,"replicate":false,"staticTTL":false},"*actionProfileFilterIndexes":{"limit":0,"precache":false,"remote":false,"replicate":false,"staticTTL":false},"*actionProfiles":{"limit":0,"precache":false,"remote":false,"replicate":false,"staticTTL":false},"*apiban":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false,"ttl":"2m0s"},"*attributeFilterIndexes":{"limit":0,"precache":false,"remote":false,"replicate":false,"staticTTL":false},"*attributeProfiles":{"limit":0,"precache":false,"remote":false,"replicate":false,"staticTTL":false},"*capsEvents":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false},"*cdrIDs":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false,"ttl":"10m0s"},"*chargerFilterIndexes":{"limit":0,"precache":false,"remote":false,"replicate":false,"staticTTL":false},"*chargerProfiles":{"limit":0,"precache":false,"remote":false,"replicate":false,"staticTTL":false},"*closedSessions":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false,"ttl":"10s"},"*diameterMessages":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false,"ttl":"3h0m0s"},"*eventCharges":{"limit":0,"precache":false,"remote":false,"replicate":false,"staticTTL":false,"ttl":"10s"},"*eventIPs":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false},"*eventResources":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false},"*filters":{"limit":0,"precache":false,"remote":false,"replicate":false,"staticTTL":false},"*ipAllocations":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false},"*ipFilterIndexes":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false},"*ipProfiles":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false},"*loadIDs":{"limit":0,"precache":false,"remote":false,"replicate":false,"staticTTL":false},"*radiusPackets":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false,"ttl":"3h0m0s"},"*rankingProfiles":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false},"*rankings":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false},"*rateFilterIndexes":{"limit":0,"precache":false,"remote":false,"replicate":false,"staticTTL":false},"*rateProfileFilterIndexes":{"limit":0,"precache":false,"remote":false,"replicate":false,"staticTTL":false},"*rateProfiles":{"limit":0,"precache":false,"remote":false,"replicate":false,"staticTTL":false},"*replicationHosts":{"limit":0,"precache":false,"remote":false,"replicate":false,"staticTTL":false},"*resourceFilterIndexes":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false},"*resourceProfiles":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false},"*resources":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false},"*reverseFilterIndexes":{"limit":0,"precache":false,"remote":false,"replicate":false,"staticTTL":false},"*routeFilterIndexes":{"limit":0,"precache":false,"remote":false,"replicate":false,"staticTTL":false},"*routeProfiles":{"limit":0,"precache":false,"remote":false,"replicate":false,"staticTTL":false},"*rpcConnections":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false},"*rpcResponses":{"limit":0,"precache":false,"remote":false,"replicate":false,"staticTTL":false,"ttl":"2s"},"*sentrypeer":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":true,"ttl":"24h0m0s"},"*statFilterIndexes":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false},"*statQueueProfiles":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false},"*statQueues":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false},"*stir":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false,"ttl":"3h0m0s"},"*thresholdFilterIndexes":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false},"*thresholdProfiles":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false},"*thresholds":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false},"*trendProfiles":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false},"*trends":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false},"*uch":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false,"ttl":"3h0m0s"}},"remoteConns":[],"replicationConns":[]}}`
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

	cfgStr = `{"filters":{"conns":{"*accounts":[{"filterIDs":null,"tenant":"","connIDs":["*internal"]}],"*resources":[{"filterIDs":null,"tenant":"","connIDs":["*internal"]}],"*stats":[{"filterIDs":null,"tenant":"","connIDs":["*internal"]}]}}}`
	var rpl8 string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.FilterSJSON},
	}, &rpl8); err != nil {
		t.Error(err)
	} else if cfgStr != rpl8 {
		t.Errorf("\nExpected %+v ,\n received: %+v", cfgStr, rpl8)
	}

	cfgStr = `{"cdrs":{"conns":{"*chargers":[{"filterIDs":null,"tenant":"","connIDs":["*internal"]}]},"enabled":true,"extraFields":[],"onlineCDRExports":null,"opts":{"*accounts":[{"filterIDs":null,"tenant":""}],"*attributes":[{"filterIDs":null,"tenant":""}],"*chargers":[{"filterIDs":null,"tenant":""}],"*ees":[{"filterIDs":null,"tenant":""}],"*rates":[{"filterIDs":null,"tenant":""}],"*refund":[{"filterIDs":null,"tenant":""}],"*rerate":[{"filterIDs":null,"tenant":""}],"*stats":[{"filterIDs":null,"tenant":""}],"*store":[{"filterIDs":null,"tenant":""}],"*thresholds":[{"filterIDs":null,"tenant":""}]},"sessionCostRetries":5}}`
	var rpl10 string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.CDRsJSON},
	}, &rpl10); err != nil {
		t.Error(err)
	} else if cfgStr != rpl10 {
		t.Errorf("\nExpected %+v ,\n received: %+v", cfgStr, rpl10)
	}
	cfgStr = `{"ers":{"conns":{},"enabled":false,"partialCacheTTL":"1s","readers":[{"cacheDumpFields":[],"concurrentRequests":1024,"eesFailedIDs":[],"eesSuccessIDs":[],"fields":[{"mandatory":true,"path":"*cgreq.ToR","tag":"ToR","type":"*variable","value":"~*req.2"},{"mandatory":true,"path":"*cgreq.OriginID","tag":"OriginID","type":"*variable","value":"~*req.3"},{"mandatory":true,"path":"*cgreq.RequestType","tag":"RequestType","type":"*variable","value":"~*req.4"},{"mandatory":true,"path":"*cgreq.Tenant","tag":"Tenant","type":"*variable","value":"~*req.6"},{"mandatory":true,"path":"*cgreq.Category","tag":"Category","type":"*variable","value":"~*req.7"},{"mandatory":true,"path":"*cgreq.Account","tag":"Account","type":"*variable","value":"~*req.8"},{"mandatory":true,"path":"*cgreq.Subject","tag":"Subject","type":"*variable","value":"~*req.9"},{"mandatory":true,"path":"*cgreq.Destination","tag":"Destination","type":"*variable","value":"~*req.10"},{"mandatory":true,"path":"*cgreq.SetupTime","tag":"SetupTime","type":"*variable","value":"~*req.11"},{"mandatory":true,"path":"*cgreq.AnswerTime","tag":"AnswerTime","type":"*variable","value":"~*req.12"},{"mandatory":true,"path":"*cgreq.Usage","tag":"Usage","type":"*variable","value":"~*req.13"}],"filters":[],"flags":[],"id":"*default","maxReconnectInterval":"5m0s","opts":{"csvFieldSeparator":",","csvHeaderDefineChar":":","csvRowLength":0,"natsSubject":"cgratesCDRs","partialCacheAction":"*none","partialOrderField":"~*req.AnswerTime"},"partialCommitFields":[],"processedPath":"/var/spool/cgrates/ers/out","reconnects":-1,"runDelay":"0","sourcePath":"/var/spool/cgrates/ers/in","startDelay":"0","tenant":"","timezone":"","type":"*none"}]}}`
	var rpl11 string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.ERsJSON},
	}, &rpl11); err != nil {
		t.Error(err)
	} else if cfgStr != rpl11 {
		t.Errorf("\nExpected %+v ,\n received: %+v", cfgStr, rpl11)
	}
	cfgStr = `{"ees":{"cache":{"*fileCSV":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false,"ttl":"5s"}},"conns":{},"enabled":false,"exporters":[{"attempts":1,"attributeContext":"","attributeIDs":[],"blocker":false,"concurrentRequests":0,"conns":{"*efs":[{"filterIDs":null,"tenant":"","connIDs":["*internal"]}]},"exportPath":"/var/spool/cgrates/ees","failedPostsDir":"/var/spool/cgrates/failed_posts","fields":[],"filters":[],"flags":[],"id":"*default","metricsResetSchedule":"","opts":{},"synchronous":false,"timezone":"","type":"*none"}]}}`
	var rpl12 string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.EEsJSON},
	}, &rpl12); err != nil {
		t.Error(err)
	} else if cfgStr != rpl12 {
		t.Errorf("\nExpected %+v ,\n received: %+v", cfgStr, rpl12)
	}

	cfgStr = `{"sessions":{"alterableFields":[],"channelSyncInterval":"0","clientProtocol":1,"conns":{"*attributes":[{"filterIDs":[],"tenant":"","connIDs":["*internal"]}],"*cdrs":[{"filterIDs":[],"tenant":"","connIDs":["*internal"]}],"*chargers":[{"filterIDs":[],"tenant":"","connIDs":["*internal"]}],"*rates":[{"filterIDs":[],"tenant":"","connIDs":["*internal"]}],"*resources":[{"filterIDs":[],"tenant":"","connIDs":["*internal"]}],"*routes":[{"filterIDs":[],"tenant":"","connIDs":["*internal"]}]},"defaultUsage":{"*any":"3h0m0s","*data":"1048576","*sms":"1","*voice":"3h0m0s"},"enabled":true,"listenBiGob":"","listenBiJSON":"127.0.0.1:2014","minDurLowBalance":"0","opts":{"*accounts":[{"filterIDs":null,"tenant":""}],"*accountsForceUsage":[],"*attributes":[{"filterIDs":null,"tenant":""}],"*attributesDerivedReply":[{"filterIDs":null,"tenant":""}],"*blockerError":[{"filterIDs":null,"tenant":""}],"*cdrs":[{"filterIDs":null,"tenant":""}],"*cdrsDerivedReply":[{"filterIDs":null,"tenant":""}],"*chargeable":[{"filterIDs":null,"tenant":""}],"*chargers":[{"filterIDs":null,"tenant":""}],"*debitInterval":[{"filterIDs":null,"tenant":""}],"*ees":[],"*eesIDs":[],"*forceUsage":[],"*initiate":[{"filterIDs":null,"tenant":""}],"*ips":[{"filterIDs":null,"tenant":""}],"*ipsAllocate":[{"filterIDs":null,"tenant":""}],"*ipsAuthorize":[{"filterIDs":null,"tenant":""}],"*ipsRelease":[{"filterIDs":null,"tenant":""}],"*maxUsage":[{"filterIDs":null,"tenant":""}],"*message":[{"filterIDs":null,"tenant":""}],"*originID":[],"*rates":[{"filterIDs":null,"tenant":""}],"*resources":[{"filterIDs":null,"tenant":""}],"*resourcesAllocate":[{"filterIDs":null,"tenant":""}],"*resourcesAuthorize":[{"filterIDs":null,"tenant":""}],"*resourcesDerivedReply":[{"filterIDs":null,"tenant":""}],"*resourcesRelease":[{"filterIDs":null,"tenant":""}],"*routes":[{"filterIDs":null,"tenant":""}],"*routesDerivedReply":[{"filterIDs":null,"tenant":""}],"*stats":[{"filterIDs":null,"tenant":""}],"*statsDerivedReply":[{"filterIDs":null,"tenant":""}],"*terminate":[{"filterIDs":null,"tenant":""}],"*thresholds":[{"filterIDs":null,"tenant":""}],"*thresholdsDerivedReply":[{"filterIDs":null,"tenant":""}],"*ttl":[{"filterIDs":null,"tenant":""}],"*ttlLastUsage":[],"*ttlLastUsed":[],"*ttlMaxDelay":[{"filterIDs":null,"tenant":""}],"*ttlUsage":[],"*update":[{"filterIDs":null,"tenant":""}]},"sessionIndexes":["OriginID"],"stir":{"allowedAttest":["*any"],"defaultAttest":"A","payloadMaxduration":"-1","privateKeyPath":"","publicKeyPath":""},"storeSessionCosts":false,"terminateAttempts":5}}`
	var rpl13 string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.SessionSJSON},
	}, &rpl13); err != nil {
		t.Error(err)
	} else if cfgStr != rpl13 {
		t.Errorf("\nExpected %+v ,\n received: %+v", cfgStr, rpl13)
	}
	cfgStr = `{"asteriskAgent":{"asteriskConns":[{"address":"127.0.0.1:8088","alias":"","ariWebsocket":false,"connectAttempts":3,"maxReconnectInterval":"0s","password":"CGRateS.org","reconnects":5,"user":"cgrates"}],"conns":{"*sessions":[{"filterIDs":null,"tenant":"","connIDs":["*birpc_internal"]}]},"enabled":false}}`
	var rpl14 string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.AsteriskAgentJSON},
	}, &rpl14); err != nil {
		t.Error(err)
	} else if cfgStr != rpl14 {
		t.Errorf("\nExpected %+v ,\n received: %+v", cfgStr, rpl14)
	}
	cfgStr = `{"freeswitchAgent":{"activeSessionDelimiter":",","conns":{"*sessions":[{"filterIDs":null,"tenant":"","connIDs":["*birpc_internal"]}]},"emptyBalanceAnnFile":"","emptyBalanceContext":"","enabled":false,"eventSocketConns":[{"address":"127.0.0.1:8021","alias":"127.0.0.1:8021","maxReconnectInterval":"0s","password":"ClueCon","reconnects":5,"replyTimeout":"1m0s"}],"extraFields":[],"lowBalanceAnnFile":"","maxWaitConnection":"2s","requestProcessors":[],"subscribePark":true}}`
	var rpl15 string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.FreeSWITCHAgentJSON},
	}, &rpl15); err != nil {
		t.Error(err)
	} else if cfgStr != rpl15 {
		t.Errorf("\nExpected %+v ,\n received: %+v", cfgStr, rpl15)
	}
	cfgStr = `{"kamailioAgent":{"conns":{"*sessions":[{"filterIDs":null,"tenant":"","connIDs":["*birpc_internal"]}]},"createCDR":false,"enabled":false,"evapiConns":[{"address":"127.0.0.1:8448","alias":"","maxReconnectInterval":"0s","reconnects":5}],"timezone":""}}`
	var rpl16 string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.KamailioAgentJSON},
	}, &rpl16); err != nil {
		t.Error(err)
	} else if cfgStr != rpl16 {
		t.Errorf("\nExpected %+v ,\n received: %+v", cfgStr, rpl16)
	}
	cfgStr = `{"diameterAgent":{"asrTemplate":"","connHealthCheckInterval":"0s","connStatusStatQueueIDs":[],"connStatusThresholdIDs":[],"conns":{"*sessions":[{"filterIDs":null,"tenant":"","connIDs":["*birpc_internal"]}]},"dictionariesAppendDefaults":true,"dictionariesPath":"/usr/share/cgrates/diameter/dict/","enabled":false,"forcedDisconnect":"*none","listeners":[{"address":"127.0.0.1:3868","network":"tcp"}],"originHost":"CGR-DA","originRealm":"cgrates.org","productName":"CGRateS","rarTemplate":"","requestProcessors":[],"syncedConnRequests":false,"vendorID":0}}`
	var rpl17 string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.DiameterAgentJSON},
	}, &rpl17); err != nil {
		t.Error(err)
	} else if cfgStr != rpl17 {
		t.Errorf("\nExpected %+v ,\n received: %+v", cfgStr, rpl17)
	}

	cfgStr = `{"httpAgent":[]}`
	var rpl18 string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.HTTPAgentJSON},
	}, &rpl18); err != nil {
		t.Error(err)
	} else if cfgStr != rpl18 {
		t.Errorf("\nExpected %+v ,\n received: %+v", cfgStr, rpl18)
	}

	cfgStr = `{"dnsAgent":{"conns":{"*sessions":[{"filterIDs":null,"tenant":"","connIDs":["*birpc_internal"]}]},"enabled":false,"listeners":[{"address":"127.0.0.1:53","network":"udp"}],"requestProcessors":[],"timezone":""}}`
	var rpl19 string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.DNSAgentJSON},
	}, &rpl19); err != nil {
		t.Error(err)
	} else if cfgStr != rpl19 {
		t.Errorf("\nExpected %+v ,\n received: %+v", cfgStr, rpl19)
	}

	cfgStr = `{"attributes":{"conns":{"*accounts":[{"filterIDs":null,"tenant":"","connIDs":["*localhost"]}],"*resources":[{"filterIDs":null,"tenant":"","connIDs":["*localhost"]}],"*stats":[{"filterIDs":null,"tenant":"","connIDs":["*localhost"]}]},"enabled":true,"existsIndexedFields":[],"indexedSelects":true,"nestedFields":false,"notExistsIndexedFields":[],"opts":{"*processRuns":[{"filterIDs":null,"tenant":""}],"*profileIDs":[],"*profileIgnoreFilters":[{"filterIDs":null,"tenant":""}],"*profileRuns":[{"filterIDs":null,"tenant":""}]},"prefixIndexedFields":[],"suffixIndexedFields":[]}}`
	var rpl20 string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.AttributeSJSON},
	}, &rpl20); err != nil {
		t.Error(err)
	} else if cfgStr != rpl20 {
		t.Errorf("\nExpected %+v ,\n received: %+v", cfgStr, rpl20)
	}

	cfgStr = `{"chargers":{"conns":{"*attributes":[{"filterIDs":null,"tenant":"","connIDs":["*internal"]}]},"enabled":true,"existsIndexedFields":[],"indexedSelects":true,"nestedFields":false,"notExistsIndexedFields":[],"prefixIndexedFields":[],"suffixIndexedFields":[]}}`
	var rpl21 string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.ChargerSJSON},
	}, &rpl21); err != nil {
		t.Error(err)
	} else if cfgStr != rpl21 {
		t.Errorf("\nExpected %+v ,\n received: %+v", cfgStr, rpl21)
	}
	if testCfgDir == "tutredis" || testCfgDir == "tutmysql" || testCfgDir == "tutmongo" {
		cfgStr = `{"resources":{"enabled":true,"existsIndexedFields":[],"indexedSelects":true,"nestedFields":false,"notExistsIndexedFields":[],"opts":{"*units":[],"*usageID":[],"*usageTTL":[]},"prefixIndexedFields":[],"storeInterval":"1s","suffixIndexedFields":[],"thresholds_conns":["*internal"]}}`
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
		cfgStr = `{"resources":{"conns":{"*thresholds":[{"filterIDs":null,"tenant":"","connIDs":["*internal"]}]},"enabled":true,"existsIndexedFields":[],"indexedSelects":true,"nestedFields":false,"notExistsIndexedFields":[],"opts":{"*units":[{"filterIDs":null,"tenant":""}],"*usageID":[{"filterIDs":null,"tenant":""}],"*usageTTL":[{"filterIDs":null,"tenant":""}]},"prefixIndexedFields":[],"storeInterval":"-1ns","suffixIndexedFields":[]}}`
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
	if testCfgDir == "tutredis" || testCfgDir == "tutmysql" || testCfgDir == "tutmongo" {
		cfgStr = `{"stats":{"eesConns":[],"eesExporterIDs":null,"enabled":true,"existsIndexedFields":[],"indexedSelects":true,"nestedFields":false,"notExistsIndexedFields":[],"opts":{"*profileIDs":[],"*profileIgnoreFilters":[],"*roundingDecimals":[]},"prefixIndexedFields":[],"storeInterval":"1s","storeUncompressedLimit":0,"suffixIndexedFields":[],"thresholds_conns":["*internal"]}}`
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
		cfgStr = `{"stats":{"conns":{"*thresholds":[{"filterIDs":null,"tenant":"","connIDs":["*internal"]}]},"eesExporterIDs":null,"enabled":true,"existsIndexedFields":[],"indexedSelects":true,"nestedFields":false,"notExistsIndexedFields":[],"opts":{"*profileIDs":[],"*profileIgnoreFilters":[{"filterIDs":null,"tenant":""}],"*roundingDecimals":[]},"prefixIndexedFields":[],"storeInterval":"-1ns","storeUncompressedLimit":0,"suffixIndexedFields":[]}}`
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
	if testCfgDir == "tutredis" || testCfgDir == "tutmysql" || testCfgDir == "tutmongo" {
		cfgStr = `{"thresholds":{"actionsConns":[],"enabled":true,"existsIndexedFields":[],"indexedSelects":true,"nestedFields":false,"notExistsIndexedFields":[],"opts":{"*profileIDs":[],"*profileIgnoreFilters":[]},"prefixIndexedFields":[],"storeInterval":"1s","suffixIndexedFields":[]}}`
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
		cfgStr = `{"thresholds":{"conns":{"*actions":[{"filterIDs":null,"tenant":"","connIDs":["*internal"]}]},"eesExporterIDs":[],"enabled":true,"existsIndexedFields":[],"indexedSelects":true,"nestedFields":false,"notExistsIndexedFields":[],"opts":{"*profileIDs":[],"*profileIgnoreFilters":[{"filterIDs":null,"tenant":""}]},"prefixIndexedFields":[],"storeInterval":"-1ns","suffixIndexedFields":[]}}`
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
	cfgStr = `{"routes":{"conns":{"*rates":[{"filterIDs":null,"tenant":"","connIDs":["*internal"]}],"*resources":[{"filterIDs":null,"tenant":"","connIDs":["*internal"]}],"*stats":[{"filterIDs":null,"tenant":"","connIDs":["*internal"]}]},"defaultRatio":1,"enabled":true,"existsIndexedFields":[],"indexedSelects":true,"nestedFields":false,"notExistsIndexedFields":[],"opts":{"*context":[{"filterIDs":null,"tenant":""}],"*ignoreErrors":[{"filterIDs":null,"tenant":""}],"*limit":[],"*maxCost":[{"tenant":"","Value":""}],"*maxItems":[],"*offset":[],"*profileCount":[{"filterIDs":null,"tenant":""}],"*usage":[{"filterIDs":null,"tenant":""}]},"prefixIndexedFields":["*req.Destination"],"suffixIndexedFields":[]}}`
	var rpl25 string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.RouteSJSON},
	}, &rpl25); err != nil {
		t.Error(err)
	} else if cfgStr != rpl25 {
		t.Errorf("\nExpected %+v ,\n received: %+v", cfgStr, rpl25)
	}

	switch testCfgDir {
	case "tutinternal":
		cfgStr := `{"loaders":[{"action":"*store","cache":{"*accounts":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false,"ttl":"5s"},"*actionProfiles":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false,"ttl":"5s"},"*attributes":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false,"ttl":"5s"},"*chargers":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false,"ttl":"5s"},"*filters":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false,"ttl":"5s"},"*ips":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false,"ttl":"5s"},"*rankings":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false,"ttl":"5s"},"*rateProfiles":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false,"ttl":"5s"},"*resources":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false,"ttl":"5s"},"*routes":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false,"ttl":"5s"},"*stats":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false,"ttl":"5s"},"*thresholds":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false,"ttl":"5s"},"*trends":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false,"ttl":"5s"}},"conns":{"*caches":[{"filterIDs":null,"tenant":"","connIDs":["*internal"]}]},"data":[{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"newBranch":true,"path":"Rules.Type","tag":"Type","type":"*variable","value":"~*req.2"},{"path":"Rules.Element","tag":"Element","type":"*variable","value":"~*req.3"},{"path":"Rules.Values","tag":"Values","type":"*variable","value":"~*req.4"}],"fileName":"Filters.csv","flags":null,"type":"*filters"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"TenantID","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ProfileID","type":"*variable","value":"~*req.1"},{"path":"FilterIDs","tag":"FilterIDs","type":"*variable","value":"~*req.2"},{"path":"Weights","tag":"Weights","type":"*variable","value":"~*req.3"},{"path":"Blockers","tag":"Blockers","type":"*variable","value":"~*req.4"},{"newBranch":true,"path":"Attributes.FilterIDs","tag":"AttributeFilterIDs","type":"*variable","value":"~*req.5"},{"path":"Attributes.Blockers","tag":"AttributeBlockers","type":"*variable","value":"~*req.6"},{"path":"Attributes.Path","tag":"Path","type":"*variable","value":"~*req.7"},{"path":"Attributes.Type","tag":"Type","type":"*variable","value":"~*req.8"},{"path":"Attributes.Value","tag":"Value","type":"*variable","value":"~*req.9"}],"fileName":"Attributes.csv","flags":null,"type":"*attributes"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"path":"FilterIDs","tag":"FilterIDs","type":"*variable","value":"~*req.2"},{"path":"Weights","tag":"Weights","type":"*variable","value":"~*req.3"},{"path":"UsageTTL","tag":"TTL","type":"*variable","value":"~*req.4"},{"path":"Limit","tag":"Limit","type":"*variable","value":"~*req.5"},{"path":"AllocationMessage","tag":"AllocationMessage","type":"*variable","value":"~*req.6"},{"path":"Blocker","tag":"Blocker","type":"*variable","value":"~*req.7"},{"path":"Stored","tag":"Stored","type":"*variable","value":"~*req.8"},{"path":"ThresholdIDs","tag":"ThresholdIDs","type":"*variable","value":"~*req.9"}],"fileName":"Resources.csv","flags":null,"type":"*resources"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"path":"FilterIDs","tag":"FilterIDs","type":"*variable","value":"~*req.2"},{"path":"Weights","tag":"Weights","type":"*variable","value":"~*req.3"},{"path":"TTL","tag":"TTL","type":"*variable","value":"~*req.4"},{"path":"Stored","tag":"Stored","type":"*variable","value":"~*req.5"},{"newBranch":true,"path":"Pools.ID","tag":"PoolID","type":"*variable","value":"~*req.6"},{"path":"Pools.FilterIDs","tag":"PoolFilterIDs","type":"*variable","value":"~*req.7"},{"path":"Pools.Type","tag":"PoolType","type":"*variable","value":"~*req.8"},{"path":"Pools.Range","tag":"PoolRange","type":"*variable","value":"~*req.9"},{"path":"Pools.Strategy","tag":"PoolStrategy","type":"*variable","value":"~*req.10"},{"path":"Pools.Message","tag":"PoolMessage","type":"*variable","value":"~*req.11"},{"path":"Pools.Weights","tag":"PoolWeights","type":"*variable","value":"~*req.12"},{"path":"Pools.Blockers","tag":"PoolBlockers","type":"*variable","value":"~*req.13"}],"fileName":"IPs.csv","flags":null,"type":"*ips"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"path":"FilterIDs","tag":"FilterIDs","type":"*variable","value":"~*req.2"},{"path":"Weights","tag":"Weights","type":"*variable","value":"~*req.3"},{"path":"Blockers","tag":"Blockers","type":"*variable","value":"~*req.4"},{"path":"QueueLength","tag":"QueueLength","type":"*variable","value":"~*req.5"},{"path":"TTL","tag":"TTL","type":"*variable","value":"~*req.6"},{"path":"MinItems","tag":"MinItems","type":"*variable","value":"~*req.7"},{"path":"Stored","tag":"Stored","type":"*variable","value":"~*req.8"},{"path":"ThresholdIDs","tag":"ThresholdIDs","type":"*variable","value":"~*req.9"},{"newBranch":true,"path":"Metrics.MetricID","tag":"MetricIDs","type":"*variable","value":"~*req.10"},{"path":"Metrics.FilterIDs","tag":"MetricFilterIDs","type":"*variable","value":"~*req.11"},{"path":"Metrics.Blockers","tag":"MetricBlockers","type":"*variable","value":"~*req.12"}],"fileName":"Stats.csv","flags":null,"type":"*stats"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"path":"FilterIDs","tag":"FilterIDs","type":"*variable","value":"~*req.2"},{"path":"Weights","tag":"Weights","type":"*variable","value":"~*req.3"},{"path":"MaxHits","tag":"MaxHits","type":"*variable","value":"~*req.4"},{"path":"MinHits","tag":"MinHits","type":"*variable","value":"~*req.5"},{"path":"MinSleep","tag":"MinSleep","type":"*variable","value":"~*req.6"},{"path":"Blocker","tag":"Blocker","type":"*variable","value":"~*req.7"},{"path":"AttributeIDs","tag":"AttributeIDs","type":"*variable","value":"~*req.8"},{"path":"ActionProfileIDs","tag":"ActionProfileIDs","type":"*variable","value":"~*req.9"},{"path":"Async","tag":"Async","type":"*variable","value":"~*req.10"},{"path":"EeIDs","tag":"EeIDs","type":"*variable","value":"~*req.11"}],"fileName":"Thresholds.csv","flags":null,"type":"*thresholds"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"path":"Schedule","tag":"Schedule","type":"*variable","value":"~*req.2"},{"path":"StatID","tag":"StatID","type":"*variable","value":"~*req.3"},{"path":"Metrics","tag":"Metrics","type":"*variable","value":"~*req.4"},{"path":"TTL","tag":"TTL","type":"*variable","value":"~*req.5"},{"path":"QueueLength","tag":"QueueLength","type":"*variable","value":"~*req.6"},{"path":"MinItems","tag":"MinItems","type":"*variable","value":"~*req.7"},{"path":"CorrelationType","tag":"CorrelationType","type":"*variable","value":"~*req.8"},{"path":"Tolerance","tag":"Tolerance","type":"*variable","value":"~*req.9"},{"path":"Stored","tag":"Stored","type":"*variable","value":"~*req.10"},{"path":"ThresholdIDs","tag":"ThresholdIDs","type":"*variable","value":"~*req.11"}],"fileName":"Trends.csv","flags":null,"type":"*trends"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"path":"Schedule","tag":"Schedule","type":"*variable","value":"~*req.2"},{"path":"StatIDs","tag":"StatIDs","type":"*variable","value":"~*req.3"},{"path":"MetricIDs","tag":"MetricIDs","type":"*variable","value":"~*req.4"},{"path":"Sorting","tag":"Sorting","type":"*variable","value":"~*req.5"},{"path":"SortingParameters","tag":"SortingParameters","type":"*variable","value":"~*req.6"},{"path":"Stored","tag":"Stored","type":"*variable","value":"~*req.7"},{"path":"ThresholdIDs","tag":"ThresholdIDs","type":"*variable","value":"~*req.8"}],"fileName":"Rankings.csv","flags":null,"type":"*rankings"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"path":"FilterIDs","tag":"FilterIDs","type":"*variable","value":"~*req.2"},{"path":"Weights","tag":"Weights","type":"*variable","value":"~*req.3"},{"path":"Blockers","tag":"Blockers","type":"*variable","value":"~*req.4"},{"path":"Sorting","tag":"Sorting","type":"*variable","value":"~*req.5"},{"path":"SortingParameters","tag":"SortingParameters","type":"*variable","value":"~*req.6"},{"newBranch":true,"path":"Routes.ID","tag":"RouteID","type":"*variable","value":"~*req.7"},{"path":"Routes.FilterIDs","tag":"RouteFilterIDs","type":"*variable","value":"~*req.8"},{"path":"Routes.AccountIDs","tag":"RouteAccountIDs","type":"*variable","value":"~*req.9"},{"path":"Routes.RateProfileIDs","tag":"RouteRateProfileIDs","type":"*variable","value":"~*req.10"},{"path":"Routes.ResourceIDs","tag":"RouteResourceIDs","type":"*variable","value":"~*req.11"},{"path":"Routes.StatIDs","tag":"RouteStatIDs","type":"*variable","value":"~*req.12"},{"path":"Routes.Weights","tag":"RouteWeights","type":"*variable","value":"~*req.13"},{"path":"Routes.Blockers","tag":"RouteBlockers","type":"*variable","value":"~*req.14"},{"path":"Routes.RouteParameters","tag":"RouteParameters","type":"*variable","value":"~*req.15"}],"fileName":"Routes.csv","flags":null,"type":"*routes"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"path":"FilterIDs","tag":"FilterIDs","type":"*variable","value":"~*req.2"},{"path":"Weights","tag":"Weights","type":"*variable","value":"~*req.3"},{"path":"Blockers","tag":"Blockers","type":"*variable","value":"~*req.4"},{"path":"RunID","tag":"RunID","type":"*variable","value":"~*req.5"},{"path":"AttributeIDs","tag":"AttributeIDs","type":"*variable","value":"~*req.6"}],"fileName":"Chargers.csv","flags":null,"type":"*chargers"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"path":"FilterIDs","tag":"FilterIDs","type":"*variable","value":"~*req.2"},{"path":"Weights","tag":"Weights","type":"*variable","value":"~*req.3"},{"path":"MinCost","tag":"MinCost","type":"*variable","value":"~*req.4"},{"path":"MaxCost","tag":"MaxCost","type":"*variable","value":"~*req.5"},{"path":"MaxCostStrategy","tag":"MaxCostStrategy","type":"*variable","value":"~*req.6"},{"filters":["*notempty:~*req.7:"],"path":"Rates[\u003c~*req.7\u003e].FilterIDs","tag":"RateFilterIDs","type":"*variable","value":"~*req.8"},{"filters":["*notempty:~*req.7:"],"path":"Rates[\u003c~*req.7\u003e].ActivationTimes","tag":"RateActivationTimes","type":"*variable","value":"~*req.9"},{"filters":["*notempty:~*req.7:"],"path":"Rates[\u003c~*req.7\u003e].Weights","tag":"RateWeights","type":"*variable","value":"~*req.10"},{"filters":["*notempty:~*req.7:"],"path":"Rates[\u003c~*req.7\u003e].Blocker","tag":"RateBlocker","type":"*variable","value":"~*req.11"},{"filters":["*notempty:~*req.7:"],"newBranch":true,"path":"Rates[\u003c~*req.7\u003e].IntervalRates.IntervalStart","tag":"RateIntervalStart","type":"*variable","value":"~*req.12"},{"filters":["*notempty:~*req.7:"],"path":"Rates[\u003c~*req.7\u003e].IntervalRates.FixedFee","tag":"RateFixedFee","type":"*variable","value":"~*req.13"},{"filters":["*notempty:~*req.7:"],"path":"Rates[\u003c~*req.7\u003e].IntervalRates.RecurrentFee","tag":"RateRecurrentFee","type":"*variable","value":"~*req.14"},{"filters":["*notempty:~*req.7:"],"path":"Rates[\u003c~*req.7\u003e].IntervalRates.Unit","tag":"RateUnit","type":"*variable","value":"~*req.15"},{"filters":["*notempty:~*req.7:"],"path":"Rates[\u003c~*req.7\u003e].IntervalRates.Increment","tag":"RateIncrement","type":"*variable","value":"~*req.16"}],"fileName":"Rates.csv","flags":null,"type":"*rateProfiles"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"path":"FilterIDs","tag":"FilterIDs","type":"*variable","value":"~*req.2"},{"path":"Weights","tag":"Weights","type":"*variable","value":"~*req.3"},{"path":"Blockers","tag":"Blockers","type":"*variable","value":"~*req.4"},{"path":"Schedule","tag":"Schedule","type":"*variable","value":"~*req.5"},{"path":"Targets[\u003c~*req.6\u003e]","tag":"TargetIDs","type":"*variable","value":"~*req.7"},{"filters":["*notempty:~*req.8:"],"path":"Actions[\u003c~*req.8\u003e].FilterIDs","tag":"ActionFilterIDs","type":"*variable","value":"~*req.9"},{"filters":["*notempty:~*req.8:"],"path":"Actions[\u003c~*req.8\u003e].TTL","tag":"ActionTTL","type":"*variable","value":"~*req.10"},{"filters":["*notempty:~*req.8:"],"path":"Actions[\u003c~*req.8\u003e].Type","tag":"ActionType","type":"*variable","value":"~*req.11"},{"filters":["*notempty:~*req.8:"],"path":"Actions[\u003c~*req.8\u003e].Opts","tag":"ActionOpts","type":"*variable","value":"~*req.12"},{"filters":["*notempty:~*req.8:"],"path":"Actions[\u003c~*req.8\u003e].Weights","tag":"ActionWeights","type":"*variable","value":"~*req.13"},{"filters":["*notempty:~*req.8:"],"path":"Actions[\u003c~*req.8\u003e].Blockers","tag":"ActionBlockers","type":"*variable","value":"~*req.14"},{"filters":["*notempty:~*req.8:"],"newBranch":true,"path":"Actions[\u003c~*req.8\u003e].Diktats.ID","tag":"ActionDiktatsID","type":"*variable","value":"~*req.15"},{"filters":["*notempty:~*req.8:"],"path":"Actions[\u003c~*req.8\u003e].Diktats.FilterIDs","tag":"ActionDiktatsFilterIDs","type":"*variable","value":"~*req.16"},{"filters":["*notempty:~*req.8:"],"path":"Actions[\u003c~*req.8\u003e].Diktats.Opts","tag":"ActionDiktatsOpts","type":"*variable","value":"~*req.17"},{"filters":["*notempty:~*req.8:"],"path":"Actions[\u003c~*req.8\u003e].Diktats.Weights","tag":"ActionDiktatsWeights","type":"*variable","value":"~*req.18"},{"filters":["*notempty:~*req.8:"],"path":"Actions[\u003c~*req.8\u003e].Diktats.Blockers","tag":"ActionDiktatsBlockers","type":"*variable","value":"~*req.19"}],"fileName":"Actions.csv","flags":null,"type":"*actionProfiles"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"path":"FilterIDs","tag":"FilterIDs","type":"*variable","value":"~*req.2"},{"path":"Weights","tag":"Weights","type":"*variable","value":"~*req.3"},{"path":"Blockers","tag":"Blockers","type":"*variable","value":"~*req.4"},{"path":"Opts","tag":"Opts","type":"*variable","value":"~*req.5"},{"filters":["*notempty:~*req.6:"],"path":"Balances[\u003c~*req.6\u003e].FilterIDs","tag":"BalanceFilterIDs","type":"*variable","value":"~*req.7"},{"filters":["*notempty:~*req.6:"],"path":"Balances[\u003c~*req.6\u003e].Weights","tag":"BalanceWeights","type":"*variable","value":"~*req.8"},{"filters":["*notempty:~*req.6:"],"path":"Balances[\u003c~*req.6\u003e].Blockers","tag":"BalanceBlockers","type":"*variable","value":"~*req.9"},{"filters":["*notempty:~*req.6:"],"path":"Balances[\u003c~*req.6\u003e].Type","tag":"BalanceType","type":"*variable","value":"~*req.10"},{"filters":["*notempty:~*req.6:"],"path":"Balances[\u003c~*req.6\u003e].Units","tag":"BalanceUnits","type":"*variable","value":"~*req.11"},{"filters":["*notempty:~*req.6:"],"path":"Balances[\u003c~*req.6\u003e].UnitFactors","tag":"BalanceUnitFactors","type":"*variable","value":"~*req.12"},{"filters":["*notempty:~*req.6:"],"path":"Balances[\u003c~*req.6\u003e].Opts","tag":"BalanceOpts","type":"*variable","value":"~*req.13"},{"filters":["*notempty:~*req.6:"],"path":"Balances[\u003c~*req.6\u003e].CostIncrements","tag":"BalanceCostIncrements","type":"*variable","value":"~*req.14"},{"filters":["*notempty:~*req.6:"],"path":"Balances[\u003c~*req.6\u003e].AttributeIDs","tag":"BalanceAttributeIDs","type":"*variable","value":"~*req.15"},{"filters":["*notempty:~*req.6:"],"path":"Balances[\u003c~*req.6\u003e].RateProfileIDs","tag":"BalanceRateProfileIDs","type":"*variable","value":"~*req.16"},{"path":"ThresholdIDs","tag":"ThresholdIDs","type":"*variable","value":"~*req.17"}],"fileName":"Accounts.csv","flags":null,"type":"*accounts"}],"enabled":false,"fieldSeparator":",","id":"*default","lockfilePath":".cgr.lck","opts":{"*cache":"","*forceLock":false,"*stopOnError":false,"*withIndex":true},"runDelay":"0","tenant":"","tpInPath":"/var/spool/cgrates/loader/in","tpOutPath":""}]}`
		var rpl26 string
		if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
			Tenant:   "cgrates.org",
			Sections: []string{config.LoaderSJSON},
		}, &rpl26); err != nil {
			t.Error(err)
		} else if cfgStr != rpl26 {
			t.Errorf("\nExpected %+v ,\n received: %+v", cfgStr, rpl26)
		}

	case "tutredis", "tutmysql", "tutmongo":
		cfgStr = `{"loaders":[{"action":"*store","cache":{"*accounts":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false,"ttl":"5s"},"*actionProfiles":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false,"ttl":"5s"},"*attributes":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false,"ttl":"5s"},"*chargers":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false,"ttl":"5s"},"*filters":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false,"ttl":"5s"},"*rankings":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false,"ttl":"5s"},"*rateProfiles":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false,"ttl":"5s"},"*resources":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false,"ttl":"5s"},"*routes":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false,"ttl":"5s"},"*stats":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false,"ttl":"5s"},"*thresholds":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false,"ttl":"5s"},"*trends":{"limit":-1,"precache":false,"remote":false,"replicate":false,"staticTTL":false,"ttl":"5s"}},"cachesConns":["*internal"],"data":[{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"newBranch":true,"path":"Rules.Type","tag":"Type","type":"*variable","value":"~*req.2"},{"path":"Rules.Element","tag":"Element","type":"*variable","value":"~*req.3"},{"path":"Rules.Values","tag":"Values","type":"*variable","value":"~*req.4"}],"fileName":"Filters.csv","flags":null,"type":"*filters"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"TenantID","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ProfileID","type":"*variable","value":"~*req.1"},{"path":"FilterIDs","tag":"FilterIDs","type":"*variable","value":"~*req.2"},{"path":"Weights","tag":"Weights","type":"*variable","value":"~*req.3"},{"path":"Blockers","tag":"Blockers","type":"*variable","value":"~*req.4"},{"newBranch":true,"path":"Attributes.FilterIDs","tag":"AttributeFilterIDs","type":"*variable","value":"~*req.5"},{"path":"Attributes.Blockers","tag":"AttributeBlockers","type":"*variable","value":"~*req.6"},{"path":"Attributes.Path","tag":"Path","type":"*variable","value":"~*req.7"},{"path":"Attributes.Type","tag":"Type","type":"*variable","value":"~*req.8"},{"path":"Attributes.Value","tag":"Value","type":"*variable","value":"~*req.9"}],"fileName":"Attributes.csv","flags":null,"type":"*attributes"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"path":"FilterIDs","tag":"FilterIDs","type":"*variable","value":"~*req.2"},{"path":"Weights","tag":"Weights","type":"*variable","value":"~*req.3"},{"path":"UsageTTL","tag":"TTL","type":"*variable","value":"~*req.4"},{"path":"Limit","tag":"Limit","type":"*variable","value":"~*req.5"},{"path":"AllocationMessage","tag":"AllocationMessage","type":"*variable","value":"~*req.6"},{"path":"Blocker","tag":"Blocker","type":"*variable","value":"~*req.7"},{"path":"Stored","tag":"Stored","type":"*variable","value":"~*req.8"},{"path":"ThresholdIDs","tag":"ThresholdIDs","type":"*variable","value":"~*req.9"}],"fileName":"Resources.csv","flags":null,"type":"*resources"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"path":"FilterIDs","tag":"FilterIDs","type":"*variable","value":"~*req.2"},{"path":"Weights","tag":"Weights","type":"*variable","value":"~*req.3"},{"path":"Blockers","tag":"Blockers","type":"*variable","value":"~*req.4"},{"path":"QueueLength","tag":"QueueLength","type":"*variable","value":"~*req.5"},{"path":"TTL","tag":"TTL","type":"*variable","value":"~*req.6"},{"path":"MinItems","tag":"MinItems","type":"*variable","value":"~*req.7"},{"path":"Stored","tag":"Stored","type":"*variable","value":"~*req.8"},{"path":"ThresholdIDs","tag":"ThresholdIDs","type":"*variable","value":"~*req.9"},{"newBranch":true,"path":"Metrics.MetricID","tag":"MetricIDs","type":"*variable","value":"~*req.10"},{"path":"Metrics.FilterIDs","tag":"MetricFilterIDs","type":"*variable","value":"~*req.11"},{"path":"Metrics.Blockers","tag":"MetricBlockers","type":"*variable","value":"~*req.12"}],"fileName":"Stats.csv","flags":null,"type":"*stats"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"path":"FilterIDs","tag":"FilterIDs","type":"*variable","value":"~*req.2"},{"path":"Weights","tag":"Weights","type":"*variable","value":"~*req.3"},{"path":"MaxHits","tag":"MaxHits","type":"*variable","value":"~*req.4"},{"path":"MinHits","tag":"MinHits","type":"*variable","value":"~*req.5"},{"path":"MinSleep","tag":"MinSleep","type":"*variable","value":"~*req.6"},{"path":"Blocker","tag":"Blocker","type":"*variable","value":"~*req.7"},{"path":"ActionProfileIDs","tag":"ActionProfileIDs","type":"*variable","value":"~*req.8"},{"path":"Async","tag":"Async","type":"*variable","value":"~*req.9"}],"fileName":"Thresholds.csv","flags":null,"type":"*thresholds"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"path":"Schedule","tag":"Schedule","type":"*variable","value":"~*req.2"},{"path":"StatID","tag":"StatID","type":"*variable","value":"~*req.3"},{"path":"Metrics","tag":"Metrics","type":"*variable","value":"~*req.4"},{"path":"TTL","tag":"TTL","type":"*variable","value":"~*req.5"},{"path":"QueueLength","tag":"QueueLength","type":"*variable","value":"~*req.6"},{"path":"MinItems","tag":"MinItems","type":"*variable","value":"~*req.7"},{"path":"CorrelationType","tag":"CorrelationType","type":"*variable","value":"~*req.8"},{"path":"Tolerance","tag":"Tolerance","type":"*variable","value":"~*req.9"},{"path":"Stored","tag":"Stored","type":"*variable","value":"~*req.10"},{"path":"ThresholdIDs","tag":"ThresholdIDs","type":"*variable","value":"~*req.11"}],"fileName":"Trends.csv","flags":null,"type":"*trends"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"path":"Schedule","tag":"Schedule","type":"*variable","value":"~*req.2"},{"path":"StatIDs","tag":"StatIDs","type":"*variable","value":"~*req.3"},{"path":"MetricIDs","tag":"MetricIDs","type":"*variable","value":"~*req.4"},{"path":"Sorting","tag":"Sorting","type":"*variable","value":"~*req.5"},{"path":"SortingParameters","tag":"SortingParameters","type":"*variable","value":"~*req.6"},{"path":"Stored","tag":"Stored","type":"*variable","value":"~*req.7"},{"path":"ThresholdIDs","tag":"ThresholdIDs","type":"*variable","value":"~*req.8"}],"fileName":"Rankings.csv","flags":null,"type":"*rankings"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"path":"FilterIDs","tag":"FilterIDs","type":"*variable","value":"~*req.2"},{"path":"Weights","tag":"Weights","type":"*variable","value":"~*req.3"},{"path":"Blockers","tag":"Blockers","type":"*variable","value":"~*req.4"},{"path":"Sorting","tag":"Sorting","type":"*variable","value":"~*req.5"},{"path":"SortingParameters","tag":"SortingParameters","type":"*variable","value":"~*req.6"},{"newBranch":true,"path":"Routes.ID","tag":"RouteID","type":"*variable","value":"~*req.7"},{"path":"Routes.FilterIDs","tag":"RouteFilterIDs","type":"*variable","value":"~*req.8"},{"path":"Routes.AccountIDs","tag":"RouteAccountIDs","type":"*variable","value":"~*req.9"},{"path":"Routes.RateProfileIDs","tag":"RouteRateProfileIDs","type":"*variable","value":"~*req.10"},{"path":"Routes.ResourceIDs","tag":"RouteResourceIDs","type":"*variable","value":"~*req.11"},{"path":"Routes.StatIDs","tag":"RouteStatIDs","type":"*variable","value":"~*req.12"},{"path":"Routes.Weights","tag":"RouteWeights","type":"*variable","value":"~*req.13"},{"path":"Routes.Blockers","tag":"RouteBlockers","type":"*variable","value":"~*req.14"},{"path":"Routes.RouteParameters","tag":"RouteParameters","type":"*variable","value":"~*req.15"}],"fileName":"Routes.csv","flags":null,"type":"*routes"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"path":"FilterIDs","tag":"FilterIDs","type":"*variable","value":"~*req.2"},{"path":"Weights","tag":"Weights","type":"*variable","value":"~*req.3"},{"path":"Blockers","tag":"Blockers","type":"*variable","value":"~*req.4"},{"path":"RunID","tag":"RunID","type":"*variable","value":"~*req.5"},{"path":"AttributeIDs","tag":"AttributeIDs","type":"*variable","value":"~*req.6"}],"fileName":"Chargers.csv","flags":null,"type":"*chargers"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"path":"FilterIDs","tag":"FilterIDs","type":"*variable","value":"~*req.2"},{"path":"Weights","tag":"Weights","type":"*variable","value":"~*req.3"},{"path":"MinCost","tag":"MinCost","type":"*variable","value":"~*req.4"},{"path":"MaxCost","tag":"MaxCost","type":"*variable","value":"~*req.5"},{"path":"MaxCostStrategy","tag":"MaxCostStrategy","type":"*variable","value":"~*req.6"},{"filters":["*notempty:~*req.7:"],"path":"Rates[\u003c~*req.7\u003e].FilterIDs","tag":"RateFilterIDs","type":"*variable","value":"~*req.8"},{"filters":["*notempty:~*req.7:"],"path":"Rates[\u003c~*req.7\u003e].ActivationTimes","tag":"RateActivationTimes","type":"*variable","value":"~*req.9"},{"filters":["*notempty:~*req.7:"],"path":"Rates[\u003c~*req.7\u003e].Weights","tag":"RateWeights","type":"*variable","value":"~*req.10"},{"filters":["*notempty:~*req.7:"],"path":"Rates[\u003c~*req.7\u003e].Blocker","tag":"RateBlocker","type":"*variable","value":"~*req.11"},{"filters":["*notempty:~*req.7:"],"newBranch":true,"path":"Rates[\u003c~*req.7\u003e].IntervalRates.IntervalStart","tag":"RateIntervalStart","type":"*variable","value":"~*req.12"},{"filters":["*notempty:~*req.7:"],"path":"Rates[\u003c~*req.7\u003e].IntervalRates.FixedFee","tag":"RateFixedFee","type":"*variable","value":"~*req.13"},{"filters":["*notempty:~*req.7:"],"path":"Rates[\u003c~*req.7\u003e].IntervalRates.RecurrentFee","tag":"RateRecurrentFee","type":"*variable","value":"~*req.14"},{"filters":["*notempty:~*req.7:"],"path":"Rates[\u003c~*req.7\u003e].IntervalRates.Unit","tag":"RateUnit","type":"*variable","value":"~*req.15"},{"filters":["*notempty:~*req.7:"],"path":"Rates[\u003c~*req.7\u003e].IntervalRates.Increment","tag":"RateIncrement","type":"*variable","value":"~*req.16"}],"fileName":"Rates.csv","flags":null,"type":"*rateProfiles"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"path":"FilterIDs","tag":"FilterIDs","type":"*variable","value":"~*req.2"},{"path":"Weights","tag":"Weights","type":"*variable","value":"~*req.3"},{"path":"Blockers","tag":"Blockers","type":"*variable","value":"~*req.4"},{"path":"Schedule","tag":"Schedule","type":"*variable","value":"~*req.5"},{"path":"Targets[\u003c~*req.6\u003e]","tag":"TargetIDs","type":"*variable","value":"~*req.7"},{"filters":["*notempty:~*req.8:"],"path":"Actions[\u003c~*req.8\u003e].FilterIDs","tag":"ActionFilterIDs","type":"*variable","value":"~*req.9"},{"filters":["*notempty:~*req.8:"],"path":"Actions[\u003c~*req.8\u003e].TTL","tag":"ActionTTL","type":"*variable","value":"~*req.10"},{"filters":["*notempty:~*req.8:"],"path":"Actions[\u003c~*req.8\u003e].Type","tag":"ActionType","type":"*variable","value":"~*req.11"},{"filters":["*notempty:~*req.8:"],"path":"Actions[\u003c~*req.8\u003e].Opts","tag":"ActionOpts","type":"*variable","value":"~*req.12"},{"filters":["*notempty:~*req.8:"],"path":"Actions[\u003c~*req.8\u003e].Weights","tag":"ActionWeights","type":"*variable","value":"~*req.13"},{"filters":["*notempty:~*req.8:"],"path":"Actions[\u003c~*req.8\u003e].Blockers","tag":"ActionBlockers","type":"*variable","value":"~*req.14"},{"filters":["*notempty:~*req.8:"],"newBranch":true,"path":"Actions[\u003c~*req.8\u003e].Diktats.ID","tag":"ActionDiktatsID","type":"*variable","value":"~*req.15"},{"filters":["*notempty:~*req.8:"],"path":"Actions[\u003c~*req.8\u003e].Diktats.FilterIDs","tag":"ActionDiktatsFilterIDs","type":"*variable","value":"~*req.16"},{"filters":["*notempty:~*req.8:"],"path":"Actions[\u003c~*req.8\u003e].Diktats.Opts","tag":"ActionDiktatsOpts","type":"*variable","value":"~*req.17"},{"filters":["*notempty:~*req.8:"],"path":"Actions[\u003c~*req.8\u003e].Diktats.Weights","tag":"ActionDiktatsWeights","type":"*variable","value":"~*req.18"},{"filters":["*notempty:~*req.8:"],"path":"Actions[\u003c~*req.8\u003e].Diktats.Blockers","tag":"ActionDiktatsBlockers","type":"*variable","value":"~*req.19"}],"fileName":"Actions.csv","flags":null,"type":"*actionProfiles"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"path":"FilterIDs","tag":"FilterIDs","type":"*variable","value":"~*req.2"},{"path":"Weights","tag":"Weights","type":"*variable","value":"~*req.3"},{"path":"Blockers","tag":"Blockers","type":"*variable","value":"~*req.4"},{"path":"Opts","tag":"Opts","type":"*variable","value":"~*req.5"},{"filters":["*notempty:~*req.6:"],"path":"Balances[\u003c~*req.6\u003e].FilterIDs","tag":"BalanceFilterIDs","type":"*variable","value":"~*req.7"},{"filters":["*notempty:~*req.6:"],"path":"Balances[\u003c~*req.6\u003e].Weights","tag":"BalanceWeights","type":"*variable","value":"~*req.8"},{"filters":["*notempty:~*req.6:"],"path":"Balances[\u003c~*req.6\u003e].Blockers","tag":"BalanceBlockers","type":"*variable","value":"~*req.9"},{"filters":["*notempty:~*req.6:"],"path":"Balances[\u003c~*req.6\u003e].Type","tag":"BalanceType","type":"*variable","value":"~*req.10"},{"filters":["*notempty:~*req.6:"],"path":"Balances[\u003c~*req.6\u003e].Units","tag":"BalanceUnits","type":"*variable","value":"~*req.11"},{"filters":["*notempty:~*req.6:"],"path":"Balances[\u003c~*req.6\u003e].UnitFactors","tag":"BalanceUnitFactors","type":"*variable","value":"~*req.12"},{"filters":["*notempty:~*req.6:"],"path":"Balances[\u003c~*req.6\u003e].Opts","tag":"BalanceOpts","type":"*variable","value":"~*req.13"},{"filters":["*notempty:~*req.6:"],"path":"Balances[\u003c~*req.6\u003e].CostIncrements","tag":"BalanceCostIncrements","type":"*variable","value":"~*req.14"},{"filters":["*notempty:~*req.6:"],"path":"Balances[\u003c~*req.6\u003e].AttributeIDs","tag":"BalanceAttributeIDs","type":"*variable","value":"~*req.15"},{"filters":["*notempty:~*req.6:"],"path":"Balances[\u003c~*req.6\u003e].RateProfileIDs","tag":"BalanceRateProfileIDs","type":"*variable","value":"~*req.16"},{"path":"ThresholdIDs","tag":"ThresholdIDs","type":"*variable","value":"~*req.17"}],"fileName":"Accounts.csv","flags":null,"type":"*accounts"}],"enabled":false,"fieldSeparator":",","id":"*default","lockfilePath":".cgr.lck","opts":{"*cache":"","*forceLock":false,"*stopOnError":false,"*withIndex":true},"runDelay":"0","tenant":"","tpInPath":"/var/spool/cgrates/loader/in","tpOutPath":"/var/spool/cgrates/loader/out"}]}`
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

	cfgStr = `{"suretax":{"billToNumber":"","businessUnit":"","clientNumber":"","clientTracking":"~*opts.*originID","customerNumber":"~*req.Subject","includeLocalCost":false,"origNumber":"~*req.Subject","p2pplus4":"","p2pzipcode":"","plus4":"","regulatoryCode":"03","responseGroup":"03","responseType":"D4","returnFileCode":"0","salesTypeCode":"R","taxExemptionCodeList":"","taxIncluded":"0","taxSitusRule":"04","termNumber":"~*req.Destination","timezone":"Local","transTypeCode":"010101","unitType":"00","units":"1","url":"","validationKey":"","zipcode":""}}`
	var rpl28 string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.SureTaxJSON},
	}, &rpl28); err != nil {
		t.Error(err)
	} else if cfgStr != rpl28 {
		t.Errorf("\nExpected %+v ,\n received: %+v", cfgStr, rpl28)
	}

	cfgStr = `{"loader":{"actionsConns":["*localhost"],"cachesConns":["*localhost"],"dataPath":"./","disableReverse":false,"fieldSeparator":",","gapiCredentials":".gapi/credentials.json","gapiToken":".gapi/token.json","tpid":""}}`
	var rpl29 string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.LoaderJSON},
	}, &rpl29); err != nil {
		t.Error(err)
	} else if cfgStr != rpl29 {
		t.Errorf("\nExpected %+v ,\n received: %+v", cfgStr, rpl29)
	}
	switch testCfgDir {
	case "tutredis", "tutmysql":
		cfgStr = `{"migrator":{"out_datadb_encoding":"msgpack","out_datadb_host":"127.0.0.1","out_datadb_name":"10","outDBOpts":{"mongoConnScheme":"mongodb","mongoQueryTimeout":"10s","redisCACertificate":"","redisClientCertificate":"","redisClientKey":"","redisCluster":false,"redisClusterOndownDelay":"0s","redisClusterSync":"5s","redisConnectAttempts":20,"redisConnectTimeout":"0s","redisMaxConns":10,"redisPoolPipelineLimit":0,"redisPoolPipelineWindow":"150µs","redisReadTimeout":"0s","redisSentinel":"","redisTLS":false,"redisWriteTimeout":"0s"},"out_datadb_password":"","out_datadb_port":"6379","out_datadb_type":"*redis","out_datadb_user":"cgrates","usersFilters":["Account"]}}`
		var rpl30 string
		if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
			Tenant:   "cgrates.org",
			Sections: []string{config.MigratorJSON},
		}, &rpl30); err != nil {
			t.Error(err)
		} else if cfgStr != rpl30 {
			t.Errorf("\nExpected %+v ,\n received: %+v", cfgStr, rpl30)
		}
	case "tutmongo":
		cfgStr = `{"migrator":{"out_datadb_encoding":"msgpack","out_datadb_host":"127.0.0.1","out_datadb_name":"10","outDBOpts":{"mongoConnScheme":"mongodb","mongoQueryTimeout":"10s","redisCACertificate":"","redisClientCertificate":"","redisClientKey":"","redisCluster":false,"redisClusterOndownDelay":"0s","redisClusterSync":"5s","redisConnectAttempts":20,"redisConnectTimeout":"0s","redisMaxConns":10,"redisPoolPipelineLimit":0,"redisPoolPipelineWindow":"150µs","redisReadTimeout":"0s","redisSentinel":"","redisTLS":false,"redisWriteTimeout":"0s"},"out_datadb_password":"","out_datadb_port":"27017","out_datadb_type":"*mongo","out_datadb_user":"cgrates","usersFilters":["Account"]}}`
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

	cfgStr = `{"registrarc":{"rpc":{"hosts":[],"refreshInterval":"5m0s","registrarsConns":[]}}}`
	var rpl32 string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.RegistrarCJSON},
	}, &rpl32); err != nil {
		t.Error(err)
	} else if cfgStr != rpl32 {
		t.Errorf("\nExpected %+v ,\n received: %+v", cfgStr, rpl32)
	}
	cfgStr = `{"analyzers":{"cleanupInterval":"1h0m0s","conns":{},"dbPath":"/var/spool/cgrates/analyzers","enabled":false,"indexType":"*scorch","opts":{"*exporterIDs":[]},"ttl":"24h0m0s"}}`
	var rpl33 string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.AnalyzerSJSON},
	}, &rpl33); err != nil {
		t.Error(err)
	} else if cfgStr != rpl33 {
		t.Errorf("\nExpected %+v ,\n received: %+v", cfgStr, rpl33)
	}

	cfgStr = `{"sipAgent":{"conns":{"*sessions":[{"filterIDs":null,"tenant":"","connIDs":["*internal"]}]},"enabled":false,"listen":"127.0.0.1:5060","listenNet":"udp","requestProcessors":[],"retransmissionTimer":"1s","timezone":""}}`
	var rpl35 string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.SIPAgentJSON},
	}, &rpl35); err != nil {
		t.Error(err)
	} else if cfgStr != rpl35 {
		t.Errorf("\nExpected %+v ,\n received: %+v", cfgStr, rpl35)
	}

	cfgStr = `{"templates":{"*asr":[{"mandatory":true,"path":"*diamreq.Session-Id","tag":"SessionId","type":"*variable","value":"~*req.Session-Id"},{"mandatory":true,"path":"*diamreq.Origin-Host","tag":"OriginHost","type":"*variable","value":"~*req.Destination-Host"},{"mandatory":true,"path":"*diamreq.Origin-Realm","tag":"OriginRealm","type":"*variable","value":"~*req.Destination-Realm"},{"mandatory":true,"path":"*diamreq.Destination-Realm","tag":"DestinationRealm","type":"*variable","value":"~*req.Origin-Realm"},{"mandatory":true,"path":"*diamreq.Destination-Host","tag":"DestinationHost","type":"*variable","value":"~*req.Origin-Host"},{"mandatory":true,"path":"*diamreq.Auth-Application-Id","tag":"AuthApplicationId","type":"*variable","value":"~*vars.*appid"}],"*cca":[{"mandatory":true,"path":"*rep.Session-Id","tag":"SessionId","type":"*variable","value":"~*req.Session-Id"},{"path":"*rep.Result-Code","tag":"ResultCode","type":"*constant","value":"2001"},{"mandatory":true,"path":"*rep.Origin-Host","tag":"OriginHost","type":"*variable","value":"~*vars.OriginHost"},{"mandatory":true,"path":"*rep.Origin-Realm","tag":"OriginRealm","type":"*variable","value":"~*vars.OriginRealm"},{"mandatory":true,"path":"*rep.Auth-Application-Id","tag":"AuthApplicationId","type":"*variable","value":"~*vars.*appid"},{"mandatory":true,"path":"*rep.CC-Request-Type","tag":"CCRequestType","type":"*variable","value":"~*req.CC-Request-Type"},{"mandatory":true,"path":"*rep.CC-Request-Number","tag":"CCRequestNumber","type":"*variable","value":"~*req.CC-Request-Number"}],"*cdrLog":[{"mandatory":true,"path":"*cdr.ToR","tag":"ToR","type":"*variable","value":"~*req.BalanceType"},{"mandatory":true,"path":"*cdr.OriginHost","tag":"OriginHost","type":"*constant","value":"127.0.0.1"},{"mandatory":true,"path":"*cdr.RequestType","tag":"RequestType","type":"*constant","value":"*none"},{"mandatory":true,"path":"*cdr.Tenant","tag":"Tenant","type":"*variable","value":"~*req.Tenant"},{"mandatory":true,"path":"*cdr.Account","tag":"Account","type":"*variable","value":"~*req.Account"},{"mandatory":true,"path":"*cdr.Subject","tag":"Subject","type":"*variable","value":"~*req.Account"},{"mandatory":true,"path":"*cdr.Cost","tag":"Cost","type":"*variable","value":"~*req.Cost"},{"mandatory":true,"path":"*cdr.Source","tag":"Source","type":"*constant","value":"*cdrLog"},{"mandatory":true,"path":"*cdr.Usage","tag":"Usage","type":"*constant","value":"1"},{"mandatory":true,"path":"*cdr.RunID","tag":"RunID","type":"*variable","value":"~*req.ActionType"},{"mandatory":true,"path":"*cdr.SetupTime","tag":"SetupTime","type":"*constant","value":"*now"},{"mandatory":true,"path":"*cdr.AnswerTime","tag":"AnswerTime","type":"*constant","value":"*now"},{"mandatory":true,"path":"*cdr.PreRated","tag":"PreRated","type":"*constant","value":"true"}],"*coa":[{"path":"*radDAReq.User-Name","tag":"User-Name","type":"*variable","value":"~*oreq.User-Name"},{"path":"*radDAReq.NAS-IP-Address","tag":"NAS-IP-Address","type":"*variable","value":"~*oreq.NAS-IP-Address"},{"path":"*radDAReq.Acct-Session-Id","tag":"Acct-Session-Id","type":"*variable","value":"~*oreq.Acct-Session-Id"},{"path":"*radDAReq.Filter-Id","tag":"Filter-Id","type":"*variable","value":"~*req.CustomFilter"}],"*dmr":[{"path":"*radDAReq.User-Name","tag":"User-Name","type":"*variable","value":"~*oreq.User-Name"},{"path":"*radDAReq.NAS-IP-Address","tag":"NAS-IP-Address","type":"*variable","value":"~*oreq.NAS-IP-Address"},{"path":"*radDAReq.Acct-Session-Id","tag":"Acct-Session-Id","type":"*variable","value":"~*oreq.Acct-Session-Id"},{"path":"*radDAReq.Reply-Message","tag":"Reply-Message","type":"*variable","value":"~*req.DisconnectCause"}],"*err":[{"mandatory":true,"path":"*rep.Session-Id","tag":"SessionId","type":"*variable","value":"~*req.Session-Id"},{"mandatory":true,"path":"*rep.Origin-Host","tag":"OriginHost","type":"*variable","value":"~*vars.OriginHost"},{"mandatory":true,"path":"*rep.Origin-Realm","tag":"OriginRealm","type":"*variable","value":"~*vars.OriginRealm"}],"*errSip":[{"mandatory":true,"path":"*rep.Request","tag":"Request","type":"*constant","value":"SIP/2.0 500 Internal Server Error"}],"*fsr":[{"path":"*cgreq.ToR","tag":"ToR","type":"*constant","value":"*voice"},{"path":"*cgreq.PDD","tag":"PDD","type":"*composed","value":"~*req.variable_progress_mediamsec;ms"},{"path":"*cgreq.ACD","tag":"ACD","type":"*composed","value":"~*req.variable_cdrAcd;s"},{"path":"*cgreq.OriginID","tag":"OriginID","type":"*variable","value":"~*req.Unique-ID"},{"path":"*opts.*originID","tag":"*originID","type":"*variable","value":"~*req.Unique-ID"},{"path":"*cgreq.OriginHost","tag":"OriginHost","type":"*variable","value":"~*vars.OriginHost"},{"path":"*cgreq.Account","tag":"Account","type":"*variable","value":"~*req.Caller-Username"},{"path":"*cgreq.Source","tag":"Source","type":"*composed","value":"FS_;~*req.Event-Name"},{"filters":["*string:*req.variable_process_cdr:false"],"path":"*cgreq.RequestType","tag":"RequestType","type":"*constant","value":"*none"},{"filters":["*string:*req.Caller-Dialplan:inline"],"path":"*cgreq.RequestType","tag":"RequestType","type":"*constant","value":"*none"},{"filters":["*exists:*cgreq.RequestType:"],"path":"*cgreq.RequestType","tag":"RequestType","type":"*constant","value":"*prepaid"},{"path":"*cgreq.Tenant","tag":"Tenant","type":"*constant","value":"cgrates.org"},{"path":"*cgreq.Category","tag":"Category","type":"*constant","value":"call"},{"path":"*cgreq.Subject","tag":"Subject","type":"*variable","value":"~*req.Caller-Username"},{"path":"*cgreq.Destination","tag":"Destination","type":"*variable","value":"~*req.Caller-Destination-Number"},{"path":"*cgreq.SetupTime","tag":"SetupTime","type":"*variable","value":"~*req.Caller-Channel-Created-Time"},{"path":"*cgreq.AnswerTime","tag":"AnswerTime","type":"*variable","value":"~*req.Caller-Channel-Answered-Time"},{"path":"*cgreq.Usage","tag":"Usage","type":"*composed","value":"~*req.variable_billsec;s"},{"path":"*cgreq.Route","tag":"Route","type":"*variable","value":"~*req.variable_cgrRoute"},{"path":"*cgreq.Cost","tag":"Cost","type":"*constant","value":"-1.0"},{"filters":["*notempty:*req.Hangup-Cause:"],"path":"*cgreq.DisconnectCause","tag":"DisconnectCause","type":"*variable","value":"~*req.Hangup-Cause"}],"*rar":[{"mandatory":true,"path":"*diamreq.Session-Id","tag":"SessionId","type":"*variable","value":"~*req.Session-Id"},{"mandatory":true,"path":"*diamreq.Origin-Host","tag":"OriginHost","type":"*variable","value":"~*req.Destination-Host"},{"mandatory":true,"path":"*diamreq.Origin-Realm","tag":"OriginRealm","type":"*variable","value":"~*req.Destination-Realm"},{"mandatory":true,"path":"*diamreq.Destination-Realm","tag":"DestinationRealm","type":"*variable","value":"~*req.Origin-Realm"},{"mandatory":true,"path":"*diamreq.Destination-Host","tag":"DestinationHost","type":"*variable","value":"~*req.Origin-Host"},{"mandatory":true,"path":"*diamreq.Auth-Application-Id","tag":"AuthApplicationId","type":"*variable","value":"~*vars.*appid"},{"path":"*diamreq.Re-Auth-Request-Type","tag":"ReAuthRequestType","type":"*constant","value":"0"}]}}`
	var rpl36 string
	if err := testRPC.Call(context.Background(), utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.TemplatesJSON},
	}, &rpl36); err != nil {
		t.Error(err)
	} else if cfgStr != rpl36 {
		t.Errorf("\nExpected %+v ,\n received: %+v", cfgStr, rpl36)
	}
	cfgStr = `{"configs":{"enabled":false,"rootDir":"/var/spool/cgrates/configs","url":"/configs/"}}`
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
