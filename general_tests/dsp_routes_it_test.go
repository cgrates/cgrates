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
	host1Cfg = `{
"general": {
	"node_id": "host1",
	"log_level": 7
},
"listen": {
	"rpc_json": ":4012",
	"rpc_gob": ":4013",
	"http": ":4080"
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
	host2Cfg = `{
"general": {
	"node_id": "host2",
	"log_level": 7
},
"listen": {
	"rpc_json": ":6012",
	"rpc_gob": ":6013",
	"http": ":6080"
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
"apiers": {
	"enabled": true
},
"rpc_conns": {
	"gob_cache": {
		"strategy": "*first",
		"conns": [
			{
				"address": "127.0.0.1:6013",
				"transport":"*gob"
			}
		]
	}
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
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

	host1 := TestEngine{ // first engine, port 4012
		ConfigJSON: host1Cfg,
	}
	ng1Client, _ := host1.Run(t)
	host2 := TestEngine{ // second engine, port 6012
		ConfigJSON: host2Cfg,
	}
	ng2Client, _ := host2.Run(t)

	// Send Status requests with *dispatchers on false.
	checkStatus(t, ng1Client, false, "account#dan.bogos", "host1")
	checkStatus(t, ng2Client, false, "account#dan.bogos", "host2")

	// Check that dispatcher routes were not cached due to *dispatchers being false.
	getCacheItem(t, ng1Client, false, utils.CacheDispatcherRoutes, "account#dan.bogos:*core", nil)
	getCacheItem(t, ng1Client, false, utils.CacheDispatcherRoutes, "account#dan.bogos:*core", nil)
}

func TestDispatcherRoutes(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

	setter := TestEngine{ // engine used to set dispatcher hosts/profiles (:2012)
		ConfigJSON: hostSetterCfg,
	}
	setterClient, _ := setter.Run(t)

	// Starting only the second dispatcher engine, for now.
	host2 := TestEngine{
		ConfigJSON:     host2Cfg,
		PreserveDataDB: true,
		PreserveStorDB: true,
	}
	ng2Client, _ := host2.Run(t)

	setDispatcherHost(t, setterClient, "host1", 4012)
	setDispatcherHost(t, setterClient, "host2", 6012)
	setDispatcherProfile(t, setterClient, "dsp_test", "host1;10", "host2;5")

	// Send status request to the second engine. "host2" will match, even though "host1" has the bigger weight.
	// That's because the first engine has not been started yet.
	checkStatus(t, ng2Client, true, "account#dan.bogos", "host2")

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
	host1 := TestEngine{
		ConfigJSON:     host1Cfg,
		PreserveDataDB: true,
		PreserveStorDB: true,
	}
	ng1Client, _ := host1.Run(t)

	// "host2" will match again due to being cached previously.
	checkStatus(t, ng1Client, true, "account#dan.bogos", "host2")

	// Clear cache and try again.
	clearCache(t, ng1Client)
	clearCache(t, ng2Client)

	// This time it will match "host1" which has the bigger weight.
	checkStatus(t, ng1Client, true, "account#dan.bogos", "host1")

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
	setDispatcherProfile(t, setterClient, "dsp_test", "host2;5")
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
	setDispatcherProfile(t, setterClient, "dsp_test2", "host1;50", "host2;150")
	getCacheItem(t, ng1Client, false, utils.CacheDispatcherRoutes, "account#dan.bogos:*core", nil)
	getCacheItem(t, ng1Client, false, utils.CacheDispatcherProfiles, "cgrates.org:dsp_test2", nil)
	getCacheItem(t, ng1Client, false, utils.CacheDispatchers, "cgrates.org:dsp_test2", nil)
}

func checkStatus(t *testing.T, client *birpc.Client, dispatch bool, routeID, expNodeID string) {
	t.Helper()
	args := &utils.TenantWithAPIOpts{
		Tenant: "cgrates.org",
		APIOpts: map[string]any{
			utils.OptsRouteID:     routeID,
			utils.MetaDispatchers: dispatch,
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

func clearCache(t *testing.T, client *birpc.Client) {
	t.Helper()
	var reply string
	if err := client.Call(context.Background(), utils.CacheSv1Clear,
		&utils.AttrCacheIDsWithAPIOpts{
			APIOpts: map[string]any{
				utils.MetaDispatchers: false,
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

func setDispatcherProfile(t *testing.T, client *birpc.Client, id string, hosts ...string) {
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
			Strategy:   "*weight",
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
