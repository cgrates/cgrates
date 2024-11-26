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
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestTrendStore(t *testing.T) {
	var dbConfig engine.DBCfg
	switch *utils.DBType {
	case utils.MetaMySQL:
	case utils.MetaMongo:
		dbConfig = engine.DBCfg{
			DataDB: &engine.DBParams{
				Type: utils.StringPointer("mongo"),
				Port: utils.IntPointer(27017),
				Name: utils.StringPointer("10"),
			},
			StorDB: &engine.DBParams{
				Type:     utils.StringPointer("mongo"),
				Name:     utils.StringPointer("cgrates"),
				Port:     utils.IntPointer(27017),
				Password: utils.StringPointer(""),
			},
		}
	case utils.MetaInternal, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}
	content := `{

"general": {
	"log_level": 7,
},
"trends": {
	"enabled": true,
	"store_interval": "1500ms",
	"stats_conns":["*localhost"],
},

"stats": {
	"enabled": true,
	"store_interval": "-1",
},

"apiers": {
	"enabled": true,
},

}
`
	tpFiles := map[string]string{
		utils.TrendsCsv: `#Tenant[0],Id[1],Schedule[2],StatID[3],Metrics[4],TTL[5],QueueLength[6],MinItems[7],CorrelationType[8],Tolerance[9],Stored[10],ThresholdIDs[11]
cgrates.org,TREND_1,@every 1s,Stats1_1,,-1,-1,1,*last,1,false,`,
		utils.StatsCsv: `#Tenant[0],Id[1],FilterIDs[2],ActivationInterval[3],QueueLength[4],TTL[5],MinItems[6],Metrics[7],MetricFilterIDs[8],Stored[9],Blocker[10],Weight[11],ThresholdIDs[12]
cgrates.org,Stats1_1,*string:~*req.Account:1001,,,,,*tcc;*acd;*tcd,,,,,`}

	ng := engine.TestEngine{
		ConfigJSON: content,
		TpFiles:    tpFiles,
		DBCfg:      dbConfig,
	}

	client, _ := ng.Run(t)

	t.Run("TrendSchedule", func(t *testing.T) {
		var scheduled int
		if err := client.Call(context.Background(), utils.TrendSv1ScheduleQueries,
			&utils.ArgScheduleTrendQueries{TrendIDs: []string{"TREND_1"}, TenantIDWithAPIOpts: utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org"}}}, &scheduled); err != nil {
			t.Fatal(err)
		} else if scheduled != 1 {
			t.Errorf("expected 1, got %d", scheduled)
		}
	})
	t.Run("ProcessStats", func(t *testing.T) {
		var reply []string
		if err := client.Call(context.Background(), utils.StatSv1ProcessEvent, &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "event1",
			Event: map[string]any{
				utils.AccountField: "1001",
				utils.AnswerTime:   time.Date(2024, 8, 22, 14, 25, 0, 0, time.UTC),
				utils.Usage:        time.Duration(1800+60) * time.Second,
				utils.Cost:         20 + float64(10),
				utils.PDD:          time.Duration(10 * time.Second),
			}}, &reply); err != nil {
			t.Error(err)
		}

	})

	t.Run("GetTrendsAfterStoreInterval", func(t *testing.T) {
		metricsChan := make(chan *engine.Trend, 1)
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		go func() {
			ticker := time.NewTicker(600 * time.Millisecond)
			var trnd engine.Trend
			for {
				select {
				case <-ticker.C:
					err := client.Call(context.Background(), utils.TrendSv1GetTrend, &utils.ArgGetTrend{ID: "TREND_1", TenantWithAPIOpts: utils.TenantWithAPIOpts{Tenant: "cgrates.org"}}, &trnd)
					if err != nil {
						if err.Error() != utils.ErrNotFound.Error() {
							t.Errorf("Trend retrieval error: %v", err)
						}
						continue
					}
					metricsChan <- &trnd
				case <-ctx.Done():
					return
				}

			}
		}()

		select {
		case trnd := <-metricsChan:
			if len(trnd.RunTimes) < 1 || len(trnd.Metrics) < 1 {
				t.Errorf("expected at least 1 runtime, got %d", len(trnd.RunTimes))
			}
		case <-ctx.Done():
			t.Error("Didn't get any trend from db")
		}
	})

	t.Run("TrendsSetConfig", func(t *testing.T) {
		var reply string
		// setting store interval to 0
		if err := client.Call(context.Background(), utils.ConfigSv1SetConfig, &config.SetConfigArgs{
			Tenant: "cgrates.org",
			Config: map[string]any{
				"trends": map[string]any{
					"enabled":        true,
					"stats_conns":    []string{"*localhost"},
					"store_interval": "0",
				},
			},
		}, &reply); err != nil {
			t.Error(err)
		} else if reply != utils.OK {
			t.Errorf("Expected OK received: %s", reply)
		}
	})
	t.Run("TrendSchedule", func(t *testing.T) {
		var scheduled int
		if err := client.Call(context.Background(), utils.TrendSv1ScheduleQueries,
			&utils.ArgScheduleTrendQueries{TrendIDs: []string{"TREND_1"}, TenantIDWithAPIOpts: utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org"}}}, &scheduled); err != nil {
			t.Fatal(err)
		} else if scheduled != 1 {
			t.Errorf("expected 1, got %d", scheduled)
		}
	})

	t.Run("GetTrendsNotStored", func(t *testing.T) {
		metricsChan := make(chan *engine.Trend, 1)
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		go func() {
			for {
				ticker := time.NewTicker(700 * time.Millisecond)
				select {
				case <-ticker.C:
					var trnd engine.Trend
					//	the trend will be not updated since storeinterval is set to 0
					err := client.Call(context.Background(), utils.TrendSv1GetTrend, &utils.ArgGetTrend{ID: "TREND_1", TenantWithAPIOpts: utils.TenantWithAPIOpts{Tenant: "cgrates.org"}}, &trnd)
					if err != nil {
						if err.Error() != utils.ErrNotFound.Error() {
							t.Errorf("Trend retrieval error: %v", err)
						}
						continue
					}
					metricsChan <- &trnd
				case <-ctx.Done():
					return
				}
			}
		}()

		select {
		case trnd := <-metricsChan:
			if len(trnd.RunTimes) < 1 && len(trnd.Metrics) < 1 {
				t.Errorf("expected 1 runtime, got %d", len(trnd.RunTimes))
			}
		case <-ctx.Done():
			t.Error("Didn't get any trend from db")
		}
	})
	t.Run("TrendsSetConfig", func(t *testing.T) {
		var reply string
		// setting store interval to -1
		if err := client.Call(context.Background(), utils.ConfigSv1SetConfig, &config.SetConfigArgs{
			Tenant: "cgrates.org",
			Config: map[string]any{
				"trends": map[string]any{
					"enabled":        true,
					"stats_conns":    []string{"*localhost"},
					"store_interval": "-1",
				},
			},
		}, &reply); err != nil {
			t.Error(err)
		} else if reply != utils.OK {
			t.Errorf("Expected OK received: %s", reply)
		}
	})
	t.Run("TrendSchedule", func(t *testing.T) {
		var scheduled int
		if err := client.Call(context.Background(), utils.TrendSv1ScheduleQueries,
			&utils.ArgScheduleTrendQueries{TrendIDs: []string{"TREND_1"}, TenantIDWithAPIOpts: utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org"}}}, &scheduled); err != nil {
			t.Fatal(err)
		} else if scheduled != 1 {
			t.Errorf("expected 1, got %d", scheduled)
		}
	})
	t.Run("GetTrendsStoredUnlimited", func(t *testing.T) {
		metricsChan := make(chan *engine.Trend, 1)
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		go func() {
			ticker := time.NewTicker(1000 * time.Millisecond)
			for {
				select {
				case <-ticker.C:
					var trnd engine.Trend
					err := client.Call(context.Background(), utils.TrendSv1GetTrend, &utils.ArgGetTrend{ID: "TREND_1", TenantWithAPIOpts: utils.TenantWithAPIOpts{Tenant: "cgrates.org"}}, &trnd)
					if err != nil {
						if err.Error() != utils.ErrNotFound.Error() {
							t.Errorf("Trend retrieval error: %v", err)
						}
						continue
					}
					metricsChan <- &trnd
				case <-ctx.Done():
					return
				}
			}
		}()
		select {
		case trnd := <-metricsChan:
			if len(trnd.RunTimes) < 2 || len(trnd.Metrics) < 2 {
				t.Errorf("expected at least 2 runtimes, got %d", len(trnd.RunTimes))
			}
		case <-ctx.Done():
			t.Error("Didn't get any trend from db")
		}
	})
}
