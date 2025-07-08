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
	"bytes"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestThresholdEES(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaMySQL:
	case utils.MetaPostgres, utils.MetaInternal, utils.MetaMongo:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}

	content := `{
"logger": {
	"level": 7
},
"data_db": {								
	"db_type": "redis",						
	"db_port": 6379, 						
	"db_name": "10", 					
},
"stor_db": {
	"db_password": "CGRateS.org"
},
"cdrs": {
	"enabled": true,
	"thresholds_conns":["*localhost"],
},
"chargers": {
	"enabled": true,
	"attributes_conns": ["*localhost"]
},
"attributes": {
	"enabled": false,
},
"thresholds": {
	"enabled": true,
	"store_interval": "-1",
	"scheduled_ids": {},
	"indexed_selects": false,
	"ees_conns": ["*localhost"],

},
"ees":{
	 "enabled": true,
	 "exporters": [
	      {
            "id": "exporter1",
            "type": "*virt",
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
		Encoding:   utils.MetaJSON,
		TpFiles: map[string]string{
			utils.ActionsCsv: `#Tenant,ID,FilterIDs,Weights,Blockers,Schedule,TargetType,TargetIDs,ActionID,ActionFilterIDs,ActionTTL,ActionType,ActionOpts,ActionPath,ActionValue
cgrates.org,*log,,,,,,,,,,,,,`,
			utils.ThresholdsCsv: `#Tenant[0],Id[1],FilterIDs[2],Weight[3],MaxHits[4],MinHits[5],MinSleep[6],Blocker[7],ActionProfileIDs[8],Async[9],EeIDs[10]
cgrates.org,Threshold1,*string:~*opts.*account:1001;*string:~*req.RequestType:*prepaid,;20,1,0,1s,false,*log,true,exporter1`,
		},
		LogBuffer: bytes.NewBuffer(nil),
	}
	client, _ := ng.Run(t)
	time.Sleep(200 * time.Millisecond)
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
			},
		}
		var rply string
		if err := client.Call(context.Background(), utils.CDRsV1ProcessEvent, ev, &rply); err != nil {
			t.Error(err)
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
