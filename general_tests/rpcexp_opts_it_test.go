//go:build integration
// +build integration

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
	"slices"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestRPCExpIT(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaRedis, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

	ng1 := engine.TestEngine{
		ConfigJSON: `{
"db": {
  "db_conns": {
    "*default": {
      "db_type": "*internal"
      }
  },
  "opts":{
    "internalDBRewriteInterval": "0s",
    "internalDBDumpInterval": "0s"
  }
},
"ees": {
	"enabled": true,
	"exporters": [
		{
			"id": "thProcessEv1",
			"type": "*rpc",
			"opts": {
				"rpcCodec": "*json",
				"connIDs": ["thEngine"],
				"serviceMethod": "ThresholdSv1.ProcessEvent",
				"keyPath": "",
				"certPath": "",
				"caPath": "",
				"tls": false,
				"rpcConnTimeout": "1s",
				"rpcReplyTimeout": "5s"
			}
		},
		{
			"id": "thProcessEv2",
			"type": "*rpc",
			"opts": {
				"rpcCodec": "*json",
				"connIDs": ["thEngine"],
				"serviceMethod": "ThresholdSv1.ProcessEvent",
				"keyPath": "",
				"certPath": "",
				"caPath": "",
				"tls": false,
				"rpcConnTimeout": "1s",
				"rpcReplyTimeout": "5s",
				"rpcAPIOpts": {
					"*thdProfileIDs": ["THD_3"]
				}
			}
		}
	]
},
"rpc_conns": {
	"thEngine": {
		"strategy": "*first",
		"conns": [
			{
				"address": "127.0.0.1:22012",
				"transport": "*json"
			}
		]
	}
},
"efs": {
	"enabled": true
}
}`,
		DBCfg:    engine.InternalDBCfg,
		Encoding: *utils.Encoding,
	}
	clientNg1, _ := ng1.Run(t)

	ng2 := engine.TestEngine{
		ConfigJSON: `{
"db": {
	"db_conns": {
		"*default": {
			"db_type": "*internal"
    	}
	},
	"opts":{
		"internalDBRewriteInterval": "0s",
		"internalDBDumpInterval": "0s"
	}
},
"listen": {
	"rpc_json": ":22012",
	"rpc_gob": ":22013",
	"http": ":22080"
},
"thresholds": {
	"enabled": true,
	"store_interval": "-1",
	"opts": {
		"*profileIDs": [
			{
				"Values": ["THD_1"]
			}
		]
	}
}
}`,
		TpFiles: map[string]string{
			utils.ThresholdsCsv: `#Tenant[0],Id[1],FilterIDs[2],Weight[3],MaxHits[4],MinHits[5],MinSleep[6],Blocker[7],ActionProfileIDs[8],Async[9],EeIDs[10]
cgrates.org,THD_1,*string:~*req.Account:1001,;10,5,3,,,*none,true,
cgrates.org,THD_2,*string:~*req.Account:1001,;20,3,2,,,*none,true,
cgrates.org,THD_3,*string:~*req.Account:1001,;15,4,1,,,*none,true,`,
		},
		DBCfg:    engine.InternalDBCfg,
		Encoding: *utils.Encoding,
	}
	clientNg2, _ := ng2.Run(t)

	time.Sleep(10 * time.Millisecond) // wait for ThresholdProfiles to load

	// Send the request to thProcessEv1 exporter without specifying any
	// *thdProfileIDs. On the receiving engine, thresholds is configured to
	// hit only "THD_1" (through *profileIDs config opt).
	exportRPCEvent(t, clientNg1, "thProcessEv1")
	checkThresholdsHits(t, clientNg2, map[string]int{
		"cgrates.org:THD_1": 1,
		"cgrates.org:THD_2": 0,
		"cgrates.org:THD_3": 0,
	})

	// Same as above, but this time specify "THD_2" under *thdProfileIDs,
	// which ignores the configured *profileOpts, and will only hit "THD_2".
	exportRPCEvent(t, clientNg1, "thProcessEv1", "THD_2")
	checkThresholdsHits(t, clientNg2, map[string]int{
		"cgrates.org:THD_1": 1,
		"cgrates.org:THD_2": 1,
		"cgrates.org:THD_3": 0,
	})

	// Same as above, but this time overwrite *thdProfileIDs through the
	// *rpc exporter configuration ("THD_2" -> "THD_3"). This will only hit
	// "THD_3".
	exportRPCEvent(t, clientNg1, "thProcessEv2", "THD_2")
	checkThresholdsHits(t, clientNg2, map[string]int{
		"cgrates.org:THD_1": 1,
		"cgrates.org:THD_2": 1,
		"cgrates.org:THD_3": 1,
	})
}

func exportRPCEvent(t *testing.T, client *birpc.Client, exporterID string, thIDs ...string) {
	t.Helper()
	args := utils.CGREventWithEeIDs{
		EeIDs: []string{exporterID},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     utils.GenUUID(),
			Event: map[string]any{
				utils.AccountField: "1001",
			},
			APIOpts: make(map[string]any),
		},
	}
	if len(thIDs) != 0 {
		args.APIOpts[utils.OptsThresholdsProfileIDs] = thIDs
	}
	method := utils.EeSv1ProcessEvent
	var reply map[string]map[string]any
	if err := client.Call(context.Background(), method, args, &reply); err != nil {
		t.Errorf("%s unexpected err received: %v", method, err)
	}
	time.Sleep(10 * time.Millisecond) // wait for ThresholdSv1.ProcessEvent to finish
}

func checkThresholdsHits(t *testing.T, client *birpc.Client, expectedThHits map[string]int) {
	t.Helper()
	thIDs := make([]string, 0, len(expectedThHits))
	for tntID := range expectedThHits {
		thID := utils.NewTenantID(tntID).ID
		thIDs = append(thIDs, thID)
	}

	method := utils.ThresholdSv1GetThresholdsForEvent
	var reply []*engine.Threshold
	if err := client.Call(context.Background(), utils.ThresholdSv1GetThresholdsForEvent,
		&utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     utils.GenUUID(),
			Event: map[string]any{
				utils.AccountField: "1001",
			},
			APIOpts: map[string]any{
				utils.OptsThresholdsProfileIDs: thIDs,
			},
		}, &reply); err != nil {
		t.Errorf("%s unexpected err received: %v", method, err)
	}

	for tntID, want := range expectedThHits {
		var got int
		if !slices.ContainsFunc(reply, func(th *engine.Threshold) bool {
			has := th.TenantID() == tntID
			if has {
				got = th.Hits
			}
			return has
		}) {
			t.Errorf("%s reply=%s, expected %q to be part of the reply", method, utils.ToJSON(reply), tntID)
		}
		if got != want {
			t.Errorf("%s hits=%d, want %d (threshold %q)", method, got, want, tntID)
		}
	}
}
