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
	"fmt"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestTrendSchedule(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}
	content := `{

"general": {
	"log_level": 7,
},

"data_db": {
	"db_type": "*internal"
},

"stor_db": {
	"db_type": "*internal"
},

"trends": {
	"enabled": true,
	"store_interval": "-1",
	"stats_conns":["*localhost"],
	"thresholds_conns": [],	
    "ees_conns": [],
},

"stats": {
	"enabled": true,
	"store_interval": "-1",
},

"apiers": {
	"enabled": true,
},

"thresholds": {
	"enabled": true,
	"store_interval": "-1"
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
			"filters":["*string:~*req.TrendID:TREND_1"],
			"fields":[
				{"tag": "TrendID", "path": "*uch.TrendID", "type": "*variable", "value": "~*req.TrendID"},
				{"tag": "EventType", "path": "*uch.EventType", "type": "*variable", "value": "~*opts.*eventType"},
				{"tag": "SumUsage", "path": "*uch.SumUsage", "type": "*variable", "value": "~*req.Metrics.*sum#~*req.Usage.Value"},
				{"tag": "AverageCallDuration", "path": "*uch.AverageCallDuration", "type": "*variable", "value": "~*req.Metrics.*acd.Value"},
				{"tag": "PDD", "path": "*uch.PDD", "type": "*variable", "value": "~*req.Metrics.*pdd.Value"},
				{"tag": "TotalCallDuration", "path": "*uch.TotalCallDuration", "type": "*variable", "value": "~*req.Metrics.*tcd.Value"},
				{"tag": "TotalCallCost", "path": "*uch.TotalCallCost", "type": "*variable", "value": "~*req.Metrics.*tcc.Value"},
			],
		},
	]
}

}
`
	tpFiles := map[string]string{
		utils.TrendsCsv: `#Tenant[0],Id[1],Schedule[2],StatID[3],Metrics[4],TTL[5],QueueLength[6],MinItems[7],CorrelationType[8],Tolerance[9],Stored[10],ThresholdIDs[11]
cgrates.org,TREND_1,@every 1s,Stats1_1,,-1,-1,1,*last,1,false,Threshold1;Threshold2
cgrates.org,TREND_2,@every 1s,Stats1_2,,-1,-1,1,*last,1,false,*none`,
		utils.StatsCsv: `#Tenant[0],Id[1],FilterIDs[2],ActivationInterval[3],QueueLength[4],TTL[5],MinItems[6],Metrics[7],MetricFilterIDs[8],Stored[9],Blocker[10],Weight[11],ThresholdIDs[12]
cgrates.org,Stats1_1,*string:~*req.Account:1001,,,,,*tcc;*acd;*tcd,,,,,
cgrates.org,Stats1_2,*string:~*req.Account:1002,,,,,*sum#~*req.Usage;*pdd,,,,,`,
		utils.ThresholdsCsv: `#Tenant[0],Id[1],FilterIDs[2],ActivationInterval[3],MaxHits[4],MinHits[5],MinSleep[6],Blocker[7],Weight[8],ActionIDs[9],Async[10]
cgrates.org,Threshold1,*string:~*req.Metrics.*acd.ID:*acd,2024-07-29T15:00:00Z,-1,10,1s,false,10,,true
cgrates.org,Threshold2,*string:~*req.Metrics.*pdd.ID:*pdd,2024-07-29T15:00:00Z,-1,10,1s,false,10,,true
`}

	ng := engine.TestEngine{
		ConfigJSON: content,
		TpFiles:    tpFiles,
	}

	client, _ := ng.Run(t)
	var tr *engine.Trend
	t.Run("TrendSchedule", func(t *testing.T) {
		var scheduled int
		if err := client.Call(context.Background(), utils.TrendSv1ScheduleQueries,
			&utils.ArgScheduleTrendQueries{TrendIDs: []string{"TREND_1", "TREND_2"}, TenantIDWithAPIOpts: utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org"}}}, &scheduled); err != nil {
			t.Fatal(err)
		} else if scheduled != 2 {
			t.Errorf("expected 2, got %d", scheduled)
		}
	})
	t.Run("ProcessStats", func(t *testing.T) {
		var reply []string
		for i := range 2 {
			i = i + 1
			if err := client.Call(context.Background(), utils.StatSv1ProcessEvent, &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     fmt.Sprintf("event%d", i),
				Event: map[string]any{
					utils.AccountField: fmt.Sprintf("100%d", i),
					utils.AnswerTime:   time.Date(2024, 8, 22, 14, 25, 0, 0, time.UTC),
					utils.Usage:        time.Duration(1800+60) * time.Second,
					utils.Cost:         20 + float64(i/10),
					utils.PDD:          time.Duration(10+i) * time.Second,
				}}, &reply); err != nil {
				t.Error(err)
			}
		}
	})
	time.Sleep(1 * time.Second)
	t.Run("GetTrend", func(t *testing.T) {
		// GetTrend without pagination parameters in args
		if err := client.Call(context.Background(), utils.TrendSv1GetTrend, &utils.ArgGetTrend{ID: "TREND_1", TenantWithAPIOpts: utils.TenantWithAPIOpts{Tenant: "cgrates.org"}}, &tr); err != nil {
			t.Error(err)
		} else if len(tr.RunTimes) != 1 || len(tr.Metrics) != 1 {
			t.Error("expected metrics to be calculated")
		}
		// GetTrend with RunIndexStart larger than RunTime length
		if err := client.Call(context.Background(), utils.TrendSv1GetTrend, &utils.ArgGetTrend{ID: "TREND_2", RunIndexStart: 1, TenantWithAPIOpts: utils.TenantWithAPIOpts{Tenant: "cgrates.org"}}, &tr); err == nil || err.Error() != utils.ErrNotFound.Error() {
			t.Error(err)
		}
	})

	t.Run("CheckThresholds", func(t *testing.T) {
		var td *engine.Threshold
		if err := client.Call(context.Background(), utils.ThresholdSv1GetThreshold,
			&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "Threshold1"}}, &td); err != nil {
			t.Error(err)
		} else if td.Hits != 0 {
			t.Errorf("Threshold with id %s expected 0 hits, got %d", td.ID, td.Hits)
		}

		if err := client.Call(context.Background(), utils.ThresholdSv1GetThreshold,
			&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "Threshold2"}}, &td); err != nil {
			t.Error(err)
		} else if td.Hits != 0 {
			t.Errorf("Threshold with id %s expected 0 hits, got %d", td.ID, td.Hits)
		}
	})

	t.Run("ProcessStats", func(t *testing.T) {
		var reply []string
		for i := range 2 {
			i = i + 1
			if err := client.Call(context.Background(), utils.StatSv1ProcessEvent, &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     fmt.Sprintf("event%d", i),
				Event: map[string]any{
					utils.AccountField: fmt.Sprintf("100%d", i),
					utils.AnswerTime:   time.Date(2024, 9, 22, 14, 25, 0, 0, time.UTC),
					utils.Usage:        time.Duration(60) * time.Second / time.Duration(i),
					utils.Cost:         30 * float64(i),
					utils.PDD:          time.Duration(20+i) * time.Second,
				}}, &reply); err != nil {
				t.Error(err)
			}
		}
	})

	t.Run("CheckExportedTrends", func(t *testing.T) {
		var eventType any
		if err := client.Call(context.Background(), utils.CacheSv1GetItem, &utils.ArgsGetCacheItemWithAPIOpts{
			Tenant: "cgrates.org",
			ArgsGetCacheItem: utils.ArgsGetCacheItem{
				CacheID: utils.CacheUCH,
				ItemID:  "TrendID",
			},
		}, &eventType); err == nil || err.Error() != utils.ErrNotFound.Error() {
			t.Error(err)
		}

		var exporterID any
		if err := client.Call(context.Background(), utils.CacheSv1GetItem, &utils.ArgsGetCacheItemWithAPIOpts{
			Tenant: "cgrates.org",
			ArgsGetCacheItem: utils.ArgsGetCacheItem{
				CacheID: utils.CacheUCH,
				ItemID:  "EventType",
			},
		}, &exporterID); err == nil || err.Error() != utils.ErrNotFound.Error() {
			t.Error(err)
		}
	})

	t.Run("TrendsSetConfig", func(t *testing.T) {
		var reply string
		if err := client.Call(context.Background(), utils.ConfigSv1SetConfig, &config.SetConfigArgs{
			Tenant: "cgrates.org",
			Config: map[string]any{
				"trends": map[string]any{
					"enabled":          true,
					"stats_conns":      []string{"*localhost"},
					"store_interval":   "-1",
					"thresholds_conns": []string{"*localhost"},
					"ees_conns":        []string{"*localhost"},
					"ees_exporter_ids": []string{"exporter1"},
				},
			},
		}, &reply); err != nil {
			t.Error(err)
		} else if reply != utils.OK {
			t.Errorf("Expected OK received: %s", reply)
		}
	})

	time.Sleep(1 * time.Second)
	t.Run("GetTrends", func(t *testing.T) {
		// GetTrend with RunTimeEnd earlier than first scheduled time
		timeEnd := time.Now().Add(-3 * time.Second).Format(time.RFC3339)
		if err := client.Call(context.Background(), utils.TrendSv1GetTrend, &utils.ArgGetTrend{ID: "TREND_1", RunTimeEnd: timeEnd, TenantWithAPIOpts: utils.TenantWithAPIOpts{Tenant: "cgrates.org"}}, &tr); err == nil || err.Error() != utils.ErrNotFound.Error() {
			t.Error(err)
		}

		// GetTrend with RunIndexEnd large than Runtimes length
		var tr *engine.Trend
		if err := client.Call(context.Background(), utils.TrendSv1GetTrend, &utils.ArgGetTrend{ID: "TREND_2", RunIndexEnd: 3, TenantWithAPIOpts: utils.TenantWithAPIOpts{Tenant: "cgrates.org"}}, &tr); err != nil {
			t.Error(err)
		} else if len(tr.RunTimes) != 2 || len(tr.Metrics) != 2 {
			t.Errorf("expected to 2 runtimes, got %d", len(tr.RunTimes))
		}

		// GetTrend with RunTimeStart after the scheduled runtimes
		timeStart := time.Now().Add(1 * time.Second).Format(time.RFC3339)
		if err := client.Call(context.Background(), utils.TrendSv1GetTrend, &utils.ArgGetTrend{ID: "TREND_2", RunTimeStart: timeStart, TenantWithAPIOpts: utils.TenantWithAPIOpts{Tenant: "cgrates.org"}}, &tr); err == nil || err.Error() != utils.ErrNotFound.Error() {
			t.Error(err)
		}
	})
	t.Run("CheckThresholds", func(t *testing.T) {
		var td *engine.Threshold
		if err := client.Call(context.Background(), utils.ThresholdSv1GetThreshold,
			&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "Threshold1"}}, &td); err != nil {
			t.Error(err)
		} else if td.Hits != 1 {
			t.Errorf("Threshold with id %s expected 1 hits, got %d", td.ID, td.Hits)
		}

		if err := client.Call(context.Background(), utils.ThresholdSv1GetThreshold,
			&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "Threshold2"}}, &td); err != nil {
			t.Error(err)
		} else if td.Hits != 0 {
			t.Errorf("Threshold with id %s expected 0 hits, got %d", td.ID, td.Hits)
		}
	})

	time.Sleep(2 * time.Second)
	t.Run("GetTrends", func(t *testing.T) {
		// GetTrend with all correctParameters
		var tr *engine.Trend
		timeStart := time.Now().Add(-4 * time.Second).Format(time.RFC3339)
		timeEnd := time.Now().Add(-1 * time.Second).Format(time.RFC3339)
		if err := client.Call(context.Background(), utils.TrendSv1GetTrend, &utils.ArgGetTrend{ID: "TREND_1", RunIndexStart: 1, RunIndexEnd: 3, RunTimeStart: timeStart, RunTimeEnd: timeEnd, TenantWithAPIOpts: utils.TenantWithAPIOpts{Tenant: "cgrates.org"}}, &tr); err != nil {
			t.Error(err)
		} else if len(tr.RunTimes) != 1 || len(tr.Metrics) != 1 {
			t.Errorf("expected to 2 runtimes, got %d", len(tr.RunTimes))
		}
		// GetTrend with incorrect indexes
		if err := client.Call(context.Background(), utils.TrendSv1GetTrend, &utils.ArgGetTrend{ID: "TREND_2", RunIndexStart: 5, RunIndexEnd: 3, TenantWithAPIOpts: utils.TenantWithAPIOpts{Tenant: "cgrates.org"}}, &tr); err == nil || err.Error() != utils.ErrNotFound.Error() {
			t.Error(err)
		}
		//GetTrend with incorrect runtime args
		timeEnd = time.Now().Add(-5 * time.Second).Format(time.RFC3339)
		timeStart = time.Now().Add(5 * time.Second).Format(time.RFC3339)
		if err := client.Call(context.Background(), utils.TrendSv1GetTrend, &utils.ArgGetTrend{ID: "TREND_2", RunTimeStart: timeStart, RunTimeEnd: timeEnd, TenantWithAPIOpts: utils.TenantWithAPIOpts{Tenant: "cgrates.org"}}, &tr); err == nil || err.Error() != utils.ErrNotFound.Error() {
			t.Error(err)
		}
		var tr2 engine.Trend
		// GetTrend with both index args
		if err := client.Call(context.Background(), utils.TrendSv1GetTrend, &utils.ArgGetTrend{ID: "TREND_1", RunIndexStart: 3, RunIndexEnd: 4, TenantWithAPIOpts: utils.TenantWithAPIOpts{Tenant: "cgrates.org"}}, &tr2); err != nil {
			t.Error(err)
		} else if len(tr2.RunTimes) != 1 || len(tr2.Metrics) != 1 {
			t.Errorf("expected to 2 runtimes, got %d", len(tr.RunTimes))
		}

		// GetTrend without any pagination args
		if err := client.Call(context.Background(), utils.TrendSv1GetTrend, &utils.ArgGetTrend{ID: "TREND_1", TenantWithAPIOpts: utils.TenantWithAPIOpts{Tenant: "cgrates.org"}}, &tr); err != nil {
			t.Error(err)
		} else if len(tr.RunTimes) != 4 || len(tr.Metrics) != 4 {
			t.Errorf("expected to 4 runtimes, got %d", len(tr.RunTimes))
		}
	})

	t.Run("CheckExportedTrends", func(t *testing.T) {
		var eventType any
		if err := client.Call(context.Background(), utils.CacheSv1GetItem, &utils.ArgsGetCacheItemWithAPIOpts{
			Tenant: "cgrates.org",
			ArgsGetCacheItem: utils.ArgsGetCacheItem{
				CacheID: utils.CacheUCH,
				ItemID:  "EventType",
			},
		}, &eventType); err != nil {
			t.Error(err)
		} else if eventType != utils.TrendUpdate {
			t.Errorf("expected %v, received %v", utils.StatUpdate, eventType)
		}

		var trendID any
		if err := client.Call(context.Background(), utils.CacheSv1GetItem, &utils.ArgsGetCacheItemWithAPIOpts{
			Tenant: "cgrates.org",
			ArgsGetCacheItem: utils.ArgsGetCacheItem{
				CacheID: utils.CacheUCH,
				ItemID:  utils.TrendID,
			},
		}, &trendID); err != nil {
			t.Error(err)
		} else if trendID != "TREND_1" {
			t.Errorf("expected %v, received %v", "STAT_EES", trendID)
		}

		var averageCallCost any
		if err := client.Call(context.Background(), utils.CacheSv1GetItem, &utils.ArgsGetCacheItemWithAPIOpts{
			Tenant: "cgrates.org",
			ArgsGetCacheItem: utils.ArgsGetCacheItem{
				CacheID: utils.CacheUCH,
				ItemID:  "AverageCallDuration",
			},
		}, &averageCallCost); err != nil {
			t.Error(err)
		} else if averageCallCost != "960000000000" {
			t.Errorf("expected %v, received %v", "960000000000", averageCallCost)
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
		} else if totalCallDuration != "1920000000000" {
			t.Errorf("expected %v, received %v", "1920000000000", totalCallDuration)
		}
		var totalCallCost any
		if err := client.Call(context.Background(), utils.CacheSv1GetItem, &utils.ArgsGetCacheItemWithAPIOpts{
			Tenant: "cgrates.org",
			ArgsGetCacheItem: utils.ArgsGetCacheItem{
				CacheID: utils.CacheUCH,
				ItemID:  "TotalCallCost",
			},
		}, &totalCallCost); err != nil {
			t.Error(err)
		} else if totalCallCost != "50" {
			t.Errorf("expected %v, received %v", "50", totalCallCost)
		}

	})

}
