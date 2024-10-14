//go:build integration

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
	"bytes"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

const (
	hostCfg = `{
"general": {
	"node_id": "%s",
	"log_level": 7
},
"listen": {
	"rpc_json": ":%[2]d12",
	"rpc_gob": ":%[2]d13",
	"http": ":%[2]d80"
},
"dispatchers":{
	"enabled": true,
	"prevent_loop": true
},
"caches":{
	"partitions": {
		"*dispatcher_profiles": {
			"limit": -1,
			"remote":true
		},
		"*dispatcher_routes": {
			"limit": -1,
			"remote":true
		},
		"*dispatchers": {
			"limit": -1,
			"remote":true
		}
	},
	"remote_conns": ["gob_cache"]
},
"rpc_conns": {
	"gob_cache": {
		"strategy": "*first",
		"conns": [
			{
				"address": "127.0.0.1:6013",
				"transport": "*gob"
			}
		]
	}
},
"apiers": {
	"enabled": true
}
}`

	hostSetterCfg = `{
"general": {
	"node_id": "setter",
	"log_level": 7
},
"apiers": {
	"enabled": true,
	"caches_conns": ["broadcast_cache"]
},
"rpc_conns": {
	"broadcast_cache": {
		"strategy": "*broadcast",
		"conns": [
			{
				"address": "127.0.0.1:2012",
				"transport": "*json"
			},
			{
				"address": "127.0.0.1:4012",
				"transport": "*json"
			},
			{
				"address": "127.0.0.1:6012",
				"transport": "*json"
			}
		]
	}
}
}`
)

func TestDispatcherRoutesNotFound(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaMySQL:
	case utils.MetaInternal, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

	host1 := engine.TestEngine{ // first engine, port 4012
		ConfigJSON: fmt.Sprintf(hostCfg, "host1", 40),
	}
	ng1Client, _ := host1.Run(t)
	host2 := engine.TestEngine{ // second engine, port 6012
		ConfigJSON: fmt.Sprintf(hostCfg, "host2", 60),
	}
	ng2Client, _ := host2.Run(t)

	// Send Status requests with *dispatchers on false.
	checkStatus(t, ng1Client, false, "account#dan.bogos", "", "host1")
	checkStatus(t, ng2Client, false, "account#dan.bogos", "", "host2")

	// Check that dispatcher routes were not cached due to *dispatchers being false.
	getCacheItem(t, ng1Client, false, utils.CacheDispatcherRoutes, "account#dan.bogos:*core", nil)
	getCacheItem(t, ng1Client, false, utils.CacheDispatcherRoutes, "account#dan.bogos:*core", nil)
}

func TestDispatcherRoutes(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaMySQL:
	case utils.MetaInternal, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

	setter := engine.TestEngine{ // engine used to set dispatcher hosts/profiles (:2012)
		ConfigJSON: hostSetterCfg,
	}
	setterClient, _ := setter.Run(t)

	// Starting only the second dispatcher engine, for now.
	host2 := engine.TestEngine{
		ConfigJSON:     fmt.Sprintf(hostCfg, "host2", 60),
		PreserveDataDB: true,
		PreserveStorDB: true,
	}
	ng2Client, _ := host2.Run(t)

	setDispatcherHost(t, setterClient, "host1", 4012)
	setDispatcherHost(t, setterClient, "host2", 6012)
	setDispatcherProfile(t, setterClient, "dsp_test", utils.MetaWeight, "host1;10", "host2;5")

	// Send status request to the second engine. "host2" will match, even though "host1" has the bigger weight.
	// That's because the first engine has not been started yet.
	checkStatus(t, ng2Client, true, "account#dan.bogos", "", "host2")

	// Check that the dispatcher route has been cached (same for the profile and the dispatcher itself).
	getCacheItem(t, ng2Client, false, utils.CacheDispatcherRoutes, "account#dan.bogos:*core", map[string]any{
		utils.Tenant:    "cgrates.org",
		utils.ProfileID: "dsp_test",
		"HostID":        "host2",
	})
	getCacheItem(t, ng2Client, false, utils.CacheDispatcherProfiles, "cgrates.org:dsp_test", map[string]any{
		utils.FilterIDs: nil,
		"Hosts": []any{
			map[string]any{
				utils.Blocker:   false,
				utils.FilterIDs: nil,
				utils.ID:        "host1",
				utils.Params:    nil,
				utils.Weight:    10.,
			},
			map[string]any{
				utils.Blocker:   false,
				utils.FilterIDs: nil,
				utils.ID:        "host2",
				utils.Params:    nil,
				utils.Weight:    5.,
			},
		},
		utils.ActivationIntervalString: nil,
		utils.ID:                       "dsp_test",
		utils.Strategy:                 "*weight",
		utils.Subsystems:               []any{"*any"},
		"StrategyParams":               nil,
		utils.Tenant:                   "cgrates.org",
		utils.Weight:                   0.,
	})

	// Reply represents a singleResultDispatcher. Unexported, so it's enough to check if it exists.
	getCacheItem(t, ng2Client, false, utils.CacheDispatchers, "cgrates.org:dsp_test", map[string]any{})

	// Start the first engine.
	host1 := engine.TestEngine{
		ConfigJSON:     fmt.Sprintf(hostCfg, "host1", 40),
		PreserveDataDB: true,
		PreserveStorDB: true,
	}
	ng1Client, _ := host1.Run(t)

	// "host2" will match again due to being cached previously.
	checkStatus(t, ng1Client, true, "account#dan.bogos", "", "host2")

	// Clear cache and try again.
	clearCache(t, ng1Client, "")
	clearCache(t, ng2Client, "")

	// This time it will match "host1" which has the bigger weight.
	checkStatus(t, ng1Client, true, "account#dan.bogos", "", "host1")

	// Check the relevant cache items. Should be the same as before, the difference being the HostID
	// from *dispatcher_routes ("host1" instead of "host2").
	getCacheItem(t, ng1Client, false, utils.CacheDispatcherRoutes, "account#dan.bogos:*core",
		map[string]any{
			utils.Tenant:    "cgrates.org",
			utils.ProfileID: "dsp_test",
			"HostID":        "host1",
		})
	getCacheItem(t, ng1Client, false, utils.CacheDispatcherProfiles, "cgrates.org:dsp_test",
		map[string]any{
			utils.ActivationIntervalString: nil,
			utils.FilterIDs:                nil,
			"Hosts": []any{
				map[string]any{
					utils.Blocker:   false,
					utils.FilterIDs: nil,
					utils.ID:        "host1",
					utils.Params:    nil,
					utils.Weight:    10.,
				},
				map[string]any{
					utils.Blocker:   false,
					utils.FilterIDs: nil,
					utils.ID:        "host2",
					utils.Params:    nil,
					utils.Weight:    5.,
				},
			},
			utils.ID:         "dsp_test",
			utils.Strategy:   "*weight",
			utils.Subsystems: []any{"*any"},
			"StrategyParams": nil,
			utils.Tenant:     "cgrates.org",
			utils.Weight:     0.,
		})
	getCacheItem(t, ng1Client, false, utils.CacheDispatchers, "cgrates.org:dsp_test", map[string]any{})

	// Overwrite the DispatcherProfile (removed host1).
	setDispatcherProfile(t, setterClient, "dsp_test", utils.MetaWeight, "host2;5")
	time.Sleep(5 * time.Millisecond) // wait for cache updates to reach all external engines

	// Check that related cache items have been updated automatically.

	// Check that cache dispatcher route/ dispatcher instance was cleared,
	// as previously "host1" matched (which is now removed).
	getCacheItem(t, ng1Client, false, utils.CacheDispatcherRoutes, "account#dan.bogos:*core", nil)
	getCacheItem(t, ng1Client, false, utils.CacheDispatcherProfiles, "cgrates.org:dsp_test",
		map[string]any{
			utils.ActivationIntervalString: nil,
			utils.FilterIDs:                nil,
			"Hosts": []any{
				map[string]any{
					utils.Blocker:   false,
					utils.FilterIDs: nil,
					utils.ID:        "host2",
					utils.Params:    nil,
					utils.Weight:    5.,
				},
			},
			utils.ID:         "dsp_test",
			utils.Strategy:   "*weight",
			utils.Subsystems: []any{"*any"},
			"StrategyParams": nil,
			utils.Tenant:     "cgrates.org",
			utils.Weight:     0.,
		})
	getCacheItem(t, ng1Client, false, utils.CacheDispatchers, "cgrates.org:dsp_test", nil)

	// Nothing happens when setting a different dispatcher profile that's using the same hosts as before.
	setDispatcherProfile(t, setterClient, "dsp_test2", utils.MetaWeight, "host1;50", "host2;150")
	getCacheItem(t, ng1Client, false, utils.CacheDispatcherRoutes, "account#dan.bogos:*core", nil)
	getCacheItem(t, ng1Client, false, utils.CacheDispatcherProfiles, "cgrates.org:dsp_test2", nil)
	getCacheItem(t, ng1Client, false, utils.CacheDispatchers, "cgrates.org:dsp_test2", nil)
}

func TestDispatchersLoadBalanceWithAuth(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaMySQL:
	case utils.MetaInternal, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

	const (
		dspCfg = `{
"general": {
	"node_id": "dispatcher",
	"log_level": 7
},
"apiers": {
	"enabled": true
},
"attributes": {
	"enabled": true
},
"dispatchers": {
	"enabled": true,
	"attributes_conns": ["*internal"]
}
}`
		hostCfg = `{
"general": {
	"node_id": "host%s",
	"log_level": 7
},
"listen": {
	"rpc_json": ":%[2]d12",
	"rpc_gob": ":%[2]d13",
	"http": ":%[2]d80"
},
"apiers": {
	"enabled": true
}
}`
	)

	dsp := engine.TestEngine{ // dispatcher engine
		ConfigJSON: dspCfg,
	}
	clientDsp, _ := dsp.Run(t)
	hostA := engine.TestEngine{ // first worker engine (additionally loads the tps), ports 210xx
		ConfigJSON:     fmt.Sprintf(hostCfg, "A", 210),
		PreserveDataDB: true,
		PreserveStorDB: true,
		TpFiles: map[string]string{
			utils.DispatcherProfilesCsv: `#Tenant,ID,Subsystems,FilterIDs,ActivationInterval,Strategy,StrategyParameters,ConnID,ConnFilterIDs,ConnWeight,ConnBlocker,ConnParameters,Weight
cgrates.org,dsp_test,,,,*round_robin,,,,,,,
cgrates.org,dsp_test,,,,,,hostA,,30,,,
cgrates.org,dsp_test,,,,,,hostB,,20,,,
cgrates.org,dsp_test,,,,,,hostC,,10,,,`,
			utils.DispatcherHostsCsv: `#Tenant[0],ID[1],Address[2],Transport[3],ConnectAttempts[4],Reconnects[5],MaxReconnectInterval[6],ConnectTimeout[7],ReplyTimeout[8],Tls[9],ClientKey[10],ClientCertificate[11],CaCertificate[12]
cgrates.org,hostA,127.0.0.1:21012,*json,1,1,,2s,2s,,,,
cgrates.org,hostB,127.0.0.1:22012,*json,1,1,,2s,2s,,,,
cgrates.org,hostC,127.0.0.1:23012,*json,1,1,,2s,2s,,,,`,
			utils.AttributesCsv: `#Tenant,ID,Contexts,FilterIDs,ActivationInterval,AttributeFilterIDs,Path,Type,Value,Blocker,Weight
cgrates.org,attr_auth,*auth,*string:~*req.ApiKey:12345,,,*req.APIMethods,*constant,CacheSv1.Clear&CoreSv1.Status,false,20`,
		},
	}
	_, _ = hostA.Run(t)
	hostB := engine.TestEngine{ // second worker engine, ports 220xx
		ConfigJSON:     fmt.Sprintf(hostCfg, "B", 220),
		PreserveDataDB: true,
		PreserveStorDB: true,
	}
	_, _ = hostB.Run(t)
	hostC := engine.TestEngine{ // third worker engine, ports 230xx
		PreserveDataDB: true,
		PreserveStorDB: true,
		ConfigJSON:     fmt.Sprintf(hostCfg, "C", 230),
	}
	_, _ = hostC.Run(t)

	// Initial check for dispatcher status.
	checkStatus(t, clientDsp, false, "account#1001", "12345", "dispatcher")

	// Test setup:
	// - 3 CGR engine workers (hostA, hostB, hostC)
	// - 4 accounts (1001, 1002, 1003, 1004)
	// - using round-robin load strategy

	// First round (dispatcher routes not yet cached)
	// Each account is assigned to a host in order, wrapping around to hostA for the 4th account.
	checkStatus(t, clientDsp, true, "account#1001", "12345", "hostA")
	checkStatus(t, clientDsp, true, "account#1002", "12345", "hostB")
	checkStatus(t, clientDsp, true, "account#1003", "12345", "hostC")
	checkStatus(t, clientDsp, true, "account#1004", "12345", "hostA")

	// Second round (previous dispatcher routes are cached)
	// Each account maintains its previously assigned host, regardless of the round-robin order.
	checkStatus(t, clientDsp, true, "account#1001", "12345", "hostA") // without routeID: hostB
	checkStatus(t, clientDsp, true, "account#1002", "12345", "hostB") // without routeID: hostC
	checkStatus(t, clientDsp, true, "account#1003", "12345", "hostC") // without routeID: hostA
	checkStatus(t, clientDsp, true, "account#1004", "12345", "hostA") // without routeID: hostB

	// Third round (clearing cache inbetween status requests)
	checkStatus(t, clientDsp, true, "account#1001", "12345", "hostA") // Without routeID: hostC
	checkStatus(t, clientDsp, true, "account#1002", "12345", "hostB") // Without routeID: hostA

	// Clearing cache resets both the cached dispatcher routes and the
	// round-robin load dispatcher. The assignment will now start over from
	// the beginning.
	clearCache(t, clientDsp, "12345")
	checkStatus(t, clientDsp, true, "account#1003", "12345", "hostA")
	checkStatus(t, clientDsp, true, "account#1004", "12345", "hostB")
}

func TestDispatchersRoutingOnAcc(t *testing.T) {
	t.Skip("skip until we find a way to mention nodeID of the worker processing the request inside the CDR")
	switch *utils.DBType {
	case utils.MetaMySQL:
	case utils.MetaInternal, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

	const (
		dspCfg = `{
"general": {
	"node_id": "dispatcher",
	"log_level": 7
},
"apiers": {
	"enabled": true
},
"dispatchers": {
	"enabled": true
}
}`
		hostCfg = `{
"general": {
	"node_id": "host%s",
	"log_level": 7
},
"listen": {
	"rpc_json": ":%[2]d12",
	"rpc_gob": ":%[2]d13",
	"http": ":%[2]d80"
},
"rals": {
	"enabled": true
},
"cdrs": {
	"enabled": true,
	"rals_conns": ["*internal"]
},
"schedulers": {
	"enabled": true,
	"cdrs_conns": ["*internal"]
},
"apiers": {
	"enabled": true,
	"scheduler_conns": ["*internal"]
},
"sessions": {
	"enabled": true,
	"listen_bijson": "127.0.0.1:%[2]d14",
	"cdrs_conns": ["*internal"],
	"chargers_conns": ["*internal"],
	"rals_conns": ["*internal"]
},
"chargers": {
	"enabled": true
}
}`
	)

	buf := &bytes.Buffer{}
	dsp := engine.TestEngine{ // dispatcher engine
		LogBuffer:  buf,
		ConfigJSON: dspCfg,
	}
	clientDsp, _ := dsp.Run(t)
	hostA := engine.TestEngine{ // first worker engine (additionally loads the tps), ports 210xx
		ConfigJSON:     fmt.Sprintf(hostCfg, "A", 210),
		PreserveDataDB: true,
		PreserveStorDB: true,
		TpFiles: map[string]string{
			utils.DispatcherProfilesCsv: `#Tenant,ID,Subsystems,FilterIDs,ActivationInterval,Strategy,StrategyParameters,ConnID,ConnFilterIDs,ConnWeight,ConnBlocker,ConnParameters,Weight
cgrates.org,dsp_test,,,,*round_robin,,,,,,,
cgrates.org,dsp_test,,,,,,hostA,,30,,,
cgrates.org,dsp_test,,,,,,hostB,,20,,,
cgrates.org,dsp_test,,,,,,hostC,,10,,,`,
			utils.DispatcherHostsCsv: `#Tenant[0],ID[1],Address[2],Transport[3],ConnectAttempts[4],Reconnects[5],MaxReconnectInterval[6],ConnectTimeout[7],ReplyTimeout[8],Tls[9],ClientKey[10],ClientCertificate[11],CaCertificate[12]
cgrates.org,hostA,127.0.0.1:21012,*json,1,1,,2s,2s,,,,
cgrates.org,hostB,127.0.0.1:22012,*json,1,1,,2s,2s,,,,
cgrates.org,hostC,127.0.0.1:23012,*json,1,1,,2s,2s,,,,`,
			utils.AccountActionsCsv: `#Tenant,Account,ActionPlanId,ActionTriggersId,AllowNegative,Disabled
cgrates.org,1001,package_topup,,,
cgrates.org,1002,package_topup,,,
cgrates.org,1003,package_topup,,,
cgrates.org,1004,package_topup,,,`,
			utils.ActionPlansCsv: `#Id,ActionsId,TimingId,Weight
package_topup,act_topup,*asap,10`,
			utils.ActionsCsv: `#ActionsId[0],Action[1],ExtraParameters[2],Filter[3],BalanceId[4],BalanceType[5],Categories[6],DestinationIds[7],RatingSubject[8],SharedGroup[9],ExpiryTime[10],TimingIds[11],Units[12],BalanceWeight[13],BalanceBlocker[14],BalanceDisabled[15],Weight[16]
act_topup,*topup_reset,,,main_balance,*sms,,,,,*unlimited,,10,,,,`,
		},
	}
	_, _ = hostA.Run(t)
	hostB := engine.TestEngine{ // second worker engine, ports 220xx
		ConfigJSON:     fmt.Sprintf(hostCfg, "B", 220),
		PreserveDataDB: true,
		PreserveStorDB: true,
	}
	_, _ = hostB.Run(t)
	hostC := engine.TestEngine{ // third worker engine, ports 230xx
		PreserveDataDB: true,
		PreserveStorDB: true,
		ConfigJSON:     fmt.Sprintf(hostCfg, "C", 230),
	}
	_, _ = hostC.Run(t)

	idx := 0
	processCDR := func(t *testing.T, client *birpc.Client, acc string) {
		idx++
		routeID := "account#:" + acc
		var reply string
		if err := client.Call(context.Background(), utils.SessionSv1ProcessCDR,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				Event: map[string]any{
					utils.RunID:        utils.MetaDefault,
					utils.Tenant:       "cgrates.org",
					utils.Category:     "sms",
					utils.ToR:          utils.MetaSMS,
					utils.OriginID:     fmt.Sprintf("processCDR%d", idx),
					utils.OriginHost:   "127.0.0.1",
					utils.RequestType:  utils.MetaPostpaid,
					utils.AccountField: acc,
					utils.Destination:  "9000",
					utils.SetupTime:    time.Date(2024, time.October, 9, 16, 14, 50, 0, time.UTC),
					utils.AnswerTime:   time.Date(2024, time.October, 9, 16, 15, 0, 0, time.UTC),
					utils.Usage:        1,
				},
				APIOpts: map[string]any{
					utils.OptsRouteID: routeID,
				},
			}, &reply); err != nil {
			t.Errorf("SessionSv1.ProcessCDR(acc: %s, idx: %d) unexpected err: %v", acc, idx, err)
		}
	}

	for range 3 {
		processCDR(t, clientDsp, "1001")
		processCDR(t, clientDsp, "1002")
		processCDR(t, clientDsp, "1003")
		processCDR(t, clientDsp, "1004")
	}

	var cdrs []*engine.CDR
	if err := clientDsp.Call(context.Background(), utils.CDRsV1GetCDRs, &utils.RPCCDRsFilterWithAPIOpts{
		RPCCDRsFilter: &utils.RPCCDRsFilter{}}, &cdrs); err != nil {
		t.Fatal(err)
	}
	// fmt.Println(utils.ToJSON(cdrs))
}

func checkStatus(t *testing.T, client *birpc.Client, dispatch bool, routeID, apiKey, expNodeID string) {
	t.Helper()
	args := &utils.TenantWithAPIOpts{
		Tenant: "cgrates.org",
		APIOpts: map[string]any{
			utils.OptsRouteID:     routeID,
			utils.MetaDispatchers: dispatch,
			utils.OptsAPIKey:      apiKey,
		},
	}
	var reply map[string]any
	if err := client.Call(context.Background(), utils.CoreSv1Status, args, &reply); err != nil {
		t.Errorf("CoreSv1.Status unexpected err: %v", err)
	} else if nodeID := reply[utils.NodeID]; nodeID != expNodeID {
		t.Errorf("CoreSv1.Status NodeID=%q, want %q", nodeID, expNodeID)
	}
}

func getCacheItem(t *testing.T, client *birpc.Client, dispatch bool, cacheID, itemID string, expItem any) {
	t.Helper()
	args := &utils.ArgsGetCacheItemWithAPIOpts{
		Tenant: "cgrates.org",
		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: cacheID,
			ItemID:  itemID,
		},
	}
	if !dispatch {
		args.APIOpts = map[string]any{
			utils.MetaDispatchers: dispatch,
		}
	}
	var reply any
	err := client.Call(context.Background(), utils.CacheSv1GetItem, args, &reply)
	if expItem != nil {
		if err != nil {
			t.Fatalf("CacheSv1.GetItem unexpected err: %v", err)
		}
		if !reflect.DeepEqual(reply, expItem) {
			t.Errorf("CacheSv1.GetItem = %s, want %s", utils.ToJSON(reply), utils.ToJSON(expItem))
		}
		return
	}
	if err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("CacheSv1.GetItem err=%v, want %v", err, utils.ErrNotFound)
	}
}

func clearCache(t *testing.T, client *birpc.Client, apiKey string) {
	t.Helper()
	var reply string
	if err := client.Call(context.Background(), utils.CacheSv1Clear,
		&utils.AttrCacheIDsWithAPIOpts{
			APIOpts: map[string]any{
				utils.MetaDispatchers: false,
				utils.OptsAPIKey:      apiKey,
			},
		}, &reply); err != nil {
		t.Fatalf("CacheSv1.Clear unexpected err: %v", err)
	}
}

func setDispatcherHost(t *testing.T, client *birpc.Client, id string, port int) {
	t.Helper()
	var reply string
	if err := client.Call(context.Background(), utils.APIerSv1SetDispatcherHost,
		&engine.DispatcherHostWithAPIOpts{
			DispatcherHost: &engine.DispatcherHost{
				Tenant: "cgrates.org",
				RemoteHost: &config.RemoteHost{
					ID:              id,
					Address:         fmt.Sprintf("127.0.0.1:%d", port),
					Transport:       "*json",
					ConnectAttempts: 1,
					Reconnects:      3,
					ConnectTimeout:  time.Second,
					ReplyTimeout:    2 * time.Second,
				},
			},
			APIOpts: map[string]any{
				utils.MetaDispatchers: false,
			},
		}, &reply); err != nil {
		t.Errorf("APIerSv1.SetDispatcherHost unexpected err: %v", err)
	}
}

func setDispatcherProfile(t *testing.T, client *birpc.Client, id, strategy string, hosts ...string) {
	t.Helper()
	hostPrfs := make(engine.DispatcherHostProfiles, 0, len(hosts))
	for _, host := range hosts {
		host, weightStr, found := strings.Cut(host, ";")
		if !found {
			t.Fatal("hosts don't respect the 'host;weight' format")
		}
		weight, err := strconv.ParseFloat(weightStr, 64)
		if err != nil {
			t.Fatal(err)
		}
		hostPrfs = append(hostPrfs, &engine.DispatcherHostProfile{
			ID:     host,
			Weight: weight,
		})
	}

	var reply string
	if err := client.Call(context.Background(), utils.APIerSv1SetDispatcherProfile, &engine.DispatcherProfileWithAPIOpts{
		DispatcherProfile: &engine.DispatcherProfile{
			Tenant:     "cgrates.org",
			ID:         id,
			Strategy:   strategy,
			Subsystems: []string{utils.MetaAny},
			Hosts:      hostPrfs,
		},
		APIOpts: map[string]any{
			utils.MetaDispatchers: false,
		},
	}, &reply); err != nil {
		t.Errorf("APIerSv1.SetDispatcherProfile unexpected err: %v", err)
	}
}
