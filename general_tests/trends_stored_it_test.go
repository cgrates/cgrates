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
	var dbConfig string
	switch *utils.DBType {
	case utils.MetaMySQL:
		dbConfig = `
"data_db": {								
	"db_type": "redis",						
	"db_port": 6379, 						
	"db_name": "10", 						
},

"stor_db": {
	"db_password": "CGRateS.org",
},`
	case utils.MetaMongo:
		dbConfig = `
"data_db": {
	"db_type": "mongo",
	"db_name": "10",
	"db_port": 27017,
},

"stor_db": {
	"db_type": "mongo",
	"db_name": "cgrates",
	"db_port": 27017,
	"db_password": "",
},`
	case utils.MetaInternal, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}
	content := `{

"general": {
	"log_level": 7,
},
` + dbConfig + `
"trends": {
	"enabled": true,
	"store_interval": "2100ms",
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
	time.Sleep(1 * time.Second)
	t.Run("GetTrendBeforeStoreInterval", func(t *testing.T) {
		var reply string
		if err := client.Call(context.Background(), utils.CacheSv1Clear, &utils.AttrCacheIDsWithAPIOpts{}, &reply); err != nil {
			t.Error(err)
		} else if reply != utils.OK {
			t.Error("Calling CacheSv1.ReloadCache got reply: ", reply)
		}
		var tr *engine.Trend
		if err := client.Call(context.Background(), utils.TrendSv1GetTrend, &utils.ArgGetTrend{ID: "TREND_1", TenantWithAPIOpts: utils.TenantWithAPIOpts{Tenant: "cgrates.org"}}, &tr); err == nil || err.Error() != utils.ErrNotFound.Error() {
			t.Error(err)
		}
	})

	time.Sleep(1200 * time.Millisecond)
	t.Run("GetTrendsAfterStoreInterval", func(t *testing.T) {
		var reply string
		if err := client.Call(context.Background(), utils.CacheSv1Clear, &utils.AttrCacheIDsWithAPIOpts{}, &reply); err != nil {
			t.Error(err)
		} else if reply != utils.OK {
			t.Error("Calling CacheSv1.ReloadCache got reply: ", reply)
		}
		var trnd *engine.Trend
		if err := client.Call(context.Background(), utils.TrendSv1GetTrend, &utils.ArgGetTrend{ID: "TREND_1", TenantWithAPIOpts: utils.TenantWithAPIOpts{Tenant: "cgrates.org"}}, &trnd); err != nil {
			t.Error(err)
		} else if len(trnd.RunTimes) != 1 || len(trnd.Metrics) != 1 {
			t.Errorf("expected to 1 runtimes, got %d", len(trnd.RunTimes))
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

	time.Sleep(1 * time.Second)
	t.Run("GetTrendsNotStored", func(t *testing.T) {
		var reply string
		if err := client.Call(context.Background(), utils.CacheSv1Clear, &utils.AttrCacheIDsWithAPIOpts{}, &reply); err != nil {
			t.Error(err)
		} else if reply != utils.OK {
			t.Error("Calling CacheSv1.ReloadCache got reply: ", reply)
		}

		var tr *engine.Trend
		//	the trend will be not updated since storeinterval is set to 0
		if err := client.Call(context.Background(), utils.TrendSv1GetTrend, &utils.ArgGetTrend{ID: "TREND_1", TenantWithAPIOpts: utils.TenantWithAPIOpts{Tenant: "cgrates.org"}}, &tr); err != nil {
			t.Error(err)
		} else if len(tr.RunTimes) != 1 || len(tr.Metrics) != 1 {
			t.Errorf("expected to 1 runtimes, got %d", len(tr.RunTimes))
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
	time.Sleep(1 * time.Second)
	t.Run("GetTrendsStoredUnlimited", func(t *testing.T) {
		var reply string
		if err := client.Call(context.Background(), utils.CacheSv1Clear, &utils.AttrCacheIDsWithAPIOpts{}, &reply); err != nil {
			t.Error(err)
		} else if reply != utils.OK {
			t.Error("Calling CacheSv1.ReloadCache got reply: ", reply)
		}
		var tr *engine.Trend
		//	the trend will be  updated since storeinterval is set to -1
		if err := client.Call(context.Background(), utils.TrendSv1GetTrend, &utils.ArgGetTrend{ID: "TREND_1", TenantWithAPIOpts: utils.TenantWithAPIOpts{Tenant: "cgrates.org"}}, &tr); err != nil {
			t.Error(err)
		} else if len(tr.RunTimes) != 2 || len(tr.Metrics) != 2 {
			t.Errorf("expected to 2 runtimes, got %d", len(tr.RunTimes))
		}
	})
}
