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

func TestThresholdEventEEs(t *testing.T) {
	var dbConfig engine.DBCfg
	switch *utils.DBType {
	case utils.MetaMySQL:
	case utils.MetaMongo:
		dbConfig = engine.MongoDBCfg
	case utils.MetaPostgres, utils.MetaInternal:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}
	content := `{
	"general": {
		"log_level": 7,
	},
	"apiers": {
		"enabled": true
	},
	"cdrs":{
	  "enabled": true,
	   "stats_conns": ["*localhost"],
	},
    "stats": {
      "enabled": true,
      "indexed_selects":false,
      "thresholds_conns":  ["*localhost"],
   },
	"thresholds": {
		"enabled": true,
		"indexed_selects":false,
		"ees_conns": ["*localhost"]
	},
	"ees": {
	"enabled": true,
	"exporters": [
		{
			"id": "exporter1",
			"type": "*virt",
			"attempts": 1,
			"synchronous": true,
			"fields":[
				{"tag": "Filter1", "path": "*uch.Filter1", "type": "*variable", "value": "~*req.Config.FilterIDs[0]"},
				{"tag": "Filter2", "path": "*uch.Filter2", "type": "*variable", "value": "~*req.Config.FilterIDs[1]"},
			],
		},
	]
}
	}`

	csvFiles := map[string]string{
		utils.StatsCsv: `#Tenant[0],Id[1],FilterIDs[2],ActivationInterval[3],QueueLength[4],TTL[5],MinItems[6],Metrics[7],MetricFilterIDs[8],Stored[9],Blocker[10],Weight[11],ThresholdIDs[12]
cgrates.org,SQ_1,*string:~*req.Account:1001,,,-1,,*sum#1,,false,,,TH1`,
		utils.ThresholdsCsv: `#Tenant[0],Id[1],FilterIDs[2],ActivationInterval[3],MaxHits[4],MinHits[5],MinSleep[6],Blocker[7],Weight[8],ActionIDs[9],Async[10],EeIDs[11]
cgrates.org,TH1,*string:~*req.StatID:SQ_1;*eq:~*req.*sum#1:2,,-1,1,0,false,,ACT_LOG,false,exporter1`,
		utils.ActionsCsv: `#ActionsId[0],Action[1],ExtraParameters[2],Filter[3],BalanceId[4],BalanceType[5],Categories[6],DestinationIds[7],RatingSubject[8],SharedGroup[9],ExpiryTime[10],TimingIds[11],Units[12],BalanceWeight[13],BalanceBlocker[14],BalanceDisabled[15],Weight[16]
ACT_LOG,*log,,,,,,,,,,,,,,,0`,
	}

	ng := engine.TestEngine{
		ConfigJSON: content,
		DBCfg:      dbConfig,
		LogBuffer:  bytes.NewBuffer(nil),
		TpFiles:    csvFiles,
	}
	client, _ := ng.Run(t)

	t.Run("CDREventStatsToThreshold", func(t *testing.T) {
		// event from StatS to ThresholdS returns NOT_FOUND but should be ignored
		var reply string
		if err := client.Call(context.Background(),
			utils.CDRsV1ProcessEvent,
			&engine.ArgV1ProcessEvent{
				CGREvent: utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     "TestEv1",
					Event: map[string]any{
						utils.ToR:          utils.MetaVoice,
						utils.OriginID:     "Origin2",
						utils.RequestType:  utils.MetaPrepaid,
						utils.AccountField: "1001",
						utils.Subject:      "1001",
						utils.Destination:  "1002",
						utils.Usage:        time.Minute,
					},
				},
			}, &reply); err != nil {
			t.Error("Unexpected error: ", err.Error())
		} else if reply != utils.OK {
			t.Error("Unexpected reply received: ", reply)
		}

	})

	t.Run("ThresholdsToEEsEvent", func(t *testing.T) {
		// it matches the threshold and passes the event to EEs without any errors
		var reply string
		if err := client.Call(context.Background(),
			utils.CDRsV1ProcessEvent,
			&engine.ArgV1ProcessEvent{
				CGREvent: utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     "TestEv1",
					Event: map[string]any{
						utils.ToR:          utils.MetaVoice,
						utils.OriginID:     "Origin1",
						utils.RequestType:  utils.MetaPrepaid,
						utils.AccountField: "1001",
						utils.Subject:      "1001",
						utils.Destination:  "1002",
						utils.Usage:        time.Minute,
					},
				},
			}, &reply); err != nil {
			t.Error("Unexpected error: ", err.Error())
		} else if reply != utils.OK {
			t.Error("Unexpected reply received: ", reply)
		}
	})
	t.Run("CheckExporterIDs", func(t *testing.T) {
		// filters in event always should be in ascending order
		var filter1 any
		if err := client.Call(context.Background(), utils.CacheSv1GetItem, &utils.ArgsGetCacheItemWithAPIOpts{
			Tenant: "cgrates.org",
			ArgsGetCacheItem: utils.ArgsGetCacheItem{
				CacheID: utils.CacheUCH,
				ItemID:  "Filter1",
			},
		}, &filter1); err != nil {
			t.Error(err)
		} else if filter1 != "*eq:~*req.*sum#1:2" {
			t.Errorf("expected %v, received %v", "*eq:~*req.*sum#1:2", filter1)
		}
		var filter2 any
		if err := client.Call(context.Background(), utils.CacheSv1GetItem, &utils.ArgsGetCacheItemWithAPIOpts{
			Tenant: "cgrates.org",
			ArgsGetCacheItem: utils.ArgsGetCacheItem{
				CacheID: utils.CacheUCH,
				ItemID:  "Filter2",
			},
		}, &filter2); err != nil {
			t.Error(err)
		} else if filter2 != "*string:~*req.StatID:SQ_1" {
			t.Errorf("expected %v, received %v", "*string:~*req.StatID:SQ_1", filter2)
		}
	})
}
