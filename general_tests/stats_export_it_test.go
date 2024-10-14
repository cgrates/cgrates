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
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestStatsEEsExport(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

	content := `{

"general": {
	"log_level": 7
},

"data_db": {								
	"db_type": "*internal"
},

"stor_db": {
	"db_type": "*internal"
},

"apiers": {
	"enabled": true
},

"stats": {
	"enabled": true,
	"store_interval": "-1",
	"ees_conns": ["*localhost"],
	"ees_exporter_ids": ["exporter1"],
},

"ees": {
	"enabled": true,
	"attributes_conns":["*localhost"],
	"exporters": [
		{
			"id": "exporter1",
			"type": "*virt",
			"flags": [""],
			"attempts": 1,
			"synchronous": true,
			"fields":[
				{"tag": "EventType", "path": "*uch.EventType", "type": "*variable", "value": "~*req.EventType"},
				{"tag": "StatID", "path": "*uch.StatID", "type": "*variable", "value": "~*req.StatID"},
				{"tag": "ExporterID", "path": "*uch.ExporterID", "type": "*variable", "value": "~*opts.*exporterID"},
				{"tag": "SumUsage", "path": "*uch.SumUsage", "type": "*variable", "value": "~*req.Metrics.*sum#~*req.Usage"},
				{"tag": "AverageCallCost", "path": "*uch.AverageCallCost", "type": "*variable", "value": "~*req.Metrics.*acc"},
				{"tag": "AnswerSeizureRatio", "path": "*uch.AnswerSeizureRatio", "type": "*variable", "value": "~*req.Metrics.*asr"},
				{"tag": "TotalCallDuration", "path": "*uch.TotalCallDuration", "type": "*variable", "value": "~*req.Metrics.*tcd"},
			],
		},
	]
}

}`

	ng := engine.TestEngine{
		ConfigJSON: content,
	}
	client, _ := ng.Run(t)

	t.Run("SetStatProfile", func(t *testing.T) {
		var reply string
		statConfig := &engine.StatQueueProfileWithAPIOpts{
			StatQueueProfile: &engine.StatQueueProfile{
				Tenant:      "cgrates.org",
				ID:          "STAT_EES",
				FilterIDs:   []string{"*string:~*req.Account:1001"},
				QueueLength: 10,
				TTL:         10 * time.Second,
				Metrics: []*engine.MetricWithFilters{
					{
						MetricID: utils.MetaTCD,
					},
					{
						MetricID: utils.MetaASR,
					},
					{
						MetricID: utils.MetaACC,
					},
					{
						MetricID: "*sum#~*req.Usage",
					},
				},
				Blocker: true,
				Stored:  true,
				Weight:  20,
			},
		}
		if err := client.Call(context.Background(), utils.APIerSv1SetStatQueueProfile, statConfig, &reply); err != nil {
			t.Error(err)
		}
	})

	t.Run("StatExportEvent", func(t *testing.T) {
		args := &engine.CGREventWithEeIDs{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "voiceEvent",
				Time:   utils.TimePointer(time.Now()),
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.AnswerTime:   time.Date(2024, 7, 14, 14, 25, 0, 0, time.UTC),
					utils.Usage:        45 * time.Second,
					utils.Cost:         12.1,
				},
			},
		}
		var reply []string
		if err := client.Call(context.Background(), utils.StatSv1ProcessEvent, args, &reply); err != nil {
			t.Error(err)
		}
	})

	t.Run("CheckExportedStats", func(t *testing.T) {
		var eventType any
		if err := client.Call(context.Background(), utils.CacheSv1GetItem, &utils.ArgsGetCacheItemWithAPIOpts{
			Tenant: "cgrates.org",
			ArgsGetCacheItem: utils.ArgsGetCacheItem{
				CacheID: utils.CacheUCH,
				ItemID:  "EventType",
			},
		}, &eventType); err != nil {
			t.Error(err)
		} else if eventType != utils.StatUpdate {
			t.Errorf("expected %v, received %v", utils.StatUpdate, eventType)
		}

		var exporterID any
		if err := client.Call(context.Background(), utils.CacheSv1GetItem, &utils.ArgsGetCacheItemWithAPIOpts{
			Tenant: "cgrates.org",
			ArgsGetCacheItem: utils.ArgsGetCacheItem{
				CacheID: utils.CacheUCH,
				ItemID:  "ExporterID",
			},
		}, &exporterID); err != nil {
			t.Error(err)
		} else if exporterID != "exporter1" {
			t.Errorf("expected %v, received %v", "exporter1", exporterID)
		}

		var statID any
		if err := client.Call(context.Background(), utils.CacheSv1GetItem, &utils.ArgsGetCacheItemWithAPIOpts{
			Tenant: "cgrates.org",
			ArgsGetCacheItem: utils.ArgsGetCacheItem{
				CacheID: utils.CacheUCH,
				ItemID:  utils.StatID,
			},
		}, &statID); err != nil {
			t.Error(err)
		} else if statID != "STAT_EES" {
			t.Errorf("expected %v, received %v", "STAT_EES", statID)
		}
	})

	var averageCallCost any
	if err := client.Call(context.Background(), utils.CacheSv1GetItem, &utils.ArgsGetCacheItemWithAPIOpts{
		Tenant: "cgrates.org",
		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: utils.CacheUCH,
			ItemID:  "AverageCallCost",
		},
	}, &averageCallCost); err != nil {
		t.Error(err)
	} else if averageCallCost != "12.1" {
		t.Errorf("expected %v, received %v", "12.1", averageCallCost)
	}

	var totalCallDuration any
	if err := client.Call(context.Background(), utils.CacheSv1GetItem, &utils.ArgsGetCacheItemWithAPIOpts{
		Tenant: "cgrates.org",
		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: utils.CacheUCH,
			ItemID:  "TotalCallDuration",
		},
	}, &totalCallDuration); err != nil {
		t.Error(err)
	} else if totalCallDuration != "45000000000" {
		t.Errorf("expected %v, received %v", "45000000000", totalCallDuration)
	}

	var sumUsage any
	if err := client.Call(context.Background(), utils.CacheSv1GetItem, &utils.ArgsGetCacheItemWithAPIOpts{
		Tenant: "cgrates.org",
		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: utils.CacheUCH,
			ItemID:  "SumUsage",
		},
	}, &sumUsage); err != nil {
		t.Error(err)
	} else if sumUsage != "45000000000" {
		t.Errorf("expected %v, received %v", "45000000000", sumUsage)
	}

	var answerSeizureRatio any
	if err := client.Call(context.Background(), utils.CacheSv1GetItem, &utils.ArgsGetCacheItemWithAPIOpts{
		Tenant: "cgrates.org",
		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: utils.CacheUCH,
			ItemID:  "AnswerSeizureRatio",
		},
	}, &answerSeizureRatio); err != nil {
		t.Error(err)
	} else if answerSeizureRatio != "100" {
		t.Errorf("expected %v, received %v", "45000000000", answerSeizureRatio)
	}

}
