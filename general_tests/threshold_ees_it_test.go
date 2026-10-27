//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package general_tests

import (
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

func TestThresholdEES(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaRedis, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

	content := `{
"logger": {
	"level": 7
},
"sessions": {
	"enabled": true,
	"conns": {
		"*thresholds": [{"connIDs": ["*localhost"]}]
	}
},
"actions": {
	"enabled": true,
},
"thresholds": {
	"enabled": true,
	"storeInterval": "-1",
	"scheduledIDs": {},
	"indexedSelects": false,
	"conns": {
		"*ees": [{"connIDs": ["*localhost"]}],
		"*actions": [{"connIDs": ["*localhost"]}]
	},

},
"ees":{
	 "enabled": true,
	 "exporters": [
	      {
            "id": "exporter1",
            "type": "*virt",
            "synchronous": true,
			"fields":[
				{"tag": "ID", "path": "*uch.ID", "type": "*variable", "value": "~*req.ID"},
				{"tag": "EventType", "path": "*uch.EventType", "type": "*variable", "value": "~*req.EventType"},
				{"tag": "FilterIDs", "path": "*uch.FilterIDs", "type": "*variable", "value": "~*req.Config.FilterIDs"},
				{"tag": "EeIDs", "path": "*uch.EeIDs", "type": "*variable", "value": "~*req.Config.EeIDs"},
				{"tag": "ActionProfileIDs", "path": "*uch.ActionProfileIDs", "type": "*variable", "value": "~*req.Config.ActionProfileIDs"},
				{"tag": "Hits", "path": "*uch.Hits", "type": "*variable", "value": "~*req.Hits"},
			],
        },
	 ],
},
"admins": {
	"enabled": true
}}`

	ng := engine.TestEngine{
		ConfigJSON: content,
		DBCfg:      engine.InternalDBCfg,
		Encoding:   utils.MetaJSON,
		TpFiles: map[string]string{
			utils.ActionsCsv: `#Tenant,ID,FilterIDs,Weights,Blockers,Schedule,TargetType,TargetIDs,ActionID,ActionFilterIDs,ActionTTL,ActionType,ActionOpts,ActionPath,ActionValue
cgrates.org,*log,,,,,,,,,,,,,`,
			utils.ThresholdsCsv: `#Tenant[0],Id[1],FilterIDs[2],Weight[3],MaxHits[4],MinHits[5],MinSleep[6],Blocker[7],AttributeIDs[8],ActionProfileIDs[9],Async[10],EeIDs[11]
cgrates.org,Threshold1,*string:~*opts.*account:1001;*string:~*req.RequestType:*prepaid,;20,1,0,1s,false,,*log,false,exporter1`,
		},
	}
	client, cfg := ng.Run(t)
	engine.WaitForServiceStart(t, client, utils.SessionS, cfg.GeneralCfg().ConnectTimeout)
	t.Run("ThresholdEES", func(t *testing.T) {
		ev := &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "Event1",
			Event: map[string]any{
				utils.RequestType:  utils.MetaPrepaid,
				utils.AccountField: "1001",
				utils.Destination:  "1002",
				utils.Usage:        time.Minute,
			},
			APIOpts: map[string]any{
				utils.MetaAccount:    "1001",
				utils.MetaThresholds: true,
				utils.MetaOriginID:   "Event1",
			},
		}
		var reply sessions.V1ProcessEventReply
		if err := client.Call(context.Background(), utils.SessionSv1ProcessEvent, ev, &reply); err != nil {
			t.Fatal(err)
		}
		if got := reply.ThresholdIDs[utils.MetaPrimary]; len(got) != 1 || got[0] != "Threshold1" {
			t.Fatalf("expected Threshold1 in SessionS reply, got %v", reply.ThresholdIDs)
		}
	})

	t.Run("ThresholdEESCheck", func(t *testing.T) {
		var thresholdID any
		if err := client.Call(context.Background(), utils.CacheSv1GetItem, &utils.ArgsGetCacheItemWithAPIOpts{
			Tenant: "cgrates.org",
			ArgsGetCacheItem: utils.ArgsGetCacheItem{
				CacheID: utils.CacheUCH,
				ItemID:  "ID",
			},
		}, &thresholdID); err != nil {
			t.Error(err)
		} else if thresholdID != "Threshold1" {
			t.Errorf("Expected threshold ID to be '%v'", thresholdID)
		}

		var eventType any
		if err := client.Call(context.Background(), utils.CacheSv1GetItem, &utils.ArgsGetCacheItemWithAPIOpts{
			Tenant: "cgrates.org",
			ArgsGetCacheItem: utils.ArgsGetCacheItem{
				CacheID: utils.CacheUCH,
				ItemID:  "EventType",
			},
		}, &eventType); err != nil {
			t.Error(err)
		} else if eventType != utils.ThresholdHit {
			t.Errorf("Expected event type to be '%v'", eventType)
		}

		var filterIDs any
		if err := client.Call(context.Background(), utils.CacheSv1GetItem, &utils.ArgsGetCacheItemWithAPIOpts{
			Tenant: "cgrates.org",
			ArgsGetCacheItem: utils.ArgsGetCacheItem{
				CacheID: utils.CacheUCH,
				ItemID:  "FilterIDs",
			},
		}, &filterIDs); err != nil {
			t.Error(err)
		} else if filterIDs != "[\"*string:~*opts.*account:1001\",\"*string:~*req.RequestType:*prepaid\"]" {
			t.Errorf("Expected filter IDs to be '%v'", filterIDs)
		}

		var hits any
		if err := client.Call(context.Background(), utils.CacheSv1GetItem, &utils.ArgsGetCacheItemWithAPIOpts{
			Tenant: "cgrates.org",
			ArgsGetCacheItem: utils.ArgsGetCacheItem{
				CacheID: utils.CacheUCH,
				ItemID:  "Hits",
			},
		}, &hits); err != nil {
			t.Error(err)
		} else if hits != "1" {
			t.Errorf("Expected filter IDs to be '%v'", hits)
		}

		var actionPrfs any
		if err := client.Call(context.Background(), utils.CacheSv1GetItem, &utils.ArgsGetCacheItemWithAPIOpts{
			Tenant: "cgrates.org",
			ArgsGetCacheItem: utils.ArgsGetCacheItem{
				CacheID: utils.CacheUCH,
				ItemID:  "ActionProfileIDs",
			},
		}, &actionPrfs); err != nil {
			t.Error(err)
		} else if actionPrfs != "[\"*log\"]" {
			t.Errorf("Expected filter IDs to be '%v'", hits)
		}
	})
}
