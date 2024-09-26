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
	"math/rand"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
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
cgrates.org,TREND_1,@every 1s,Stats1_1,,-1,-1,1,*last,1,false,
cgrates.org,TREND_2,@every 1s,Stats1_2,,-1,-1,1,*last,1,false,`,
		utils.StatsCsv: `#Tenant[0],Id[1],FilterIDs[2],ActivationInterval[3],QueueLength[4],TTL[5],MinItems[6],Metrics[7],MetricFilterIDs[8],Stored[9],Blocker[10],Weight[11],ThresholdIDs[12]
cgrates.org,Stats1_1,*string:~*req.Account:1001,,,,,*tcc;*acd;*tcd,,,,,
cgrates.org,Stats1_2,*string:~*req.Account:1002,,,,,*sum#~*req.Usage;*pdd,,,,,`}

	testEnv := TestEnvironment{
		Name:       "TestTrendSchedule",
		ConfigJSON: content,
		TpFiles:    tpFiles,
	}

	client, _ := testEnv.Setup(t, *utils.WaitRater)
	t.Run("CheckTrendSchedule", func(t *testing.T) {
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
		if err := client.Call(context.Background(), utils.StatSv1ProcessEvent, &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     fmt.Sprintf("event%d", 1),
			Event: map[string]any{
				utils.AccountField: "1001",
				utils.AnswerTime:   time.Date(2024, 8, 22, 14, 25, 0, 0, time.UTC),
				utils.Usage:        time.Duration(rand.Intn(3600)+60) * time.Second,
				utils.Cost:         rand.Float64()*20 + 0.1,
				utils.PDD:          time.Duration(rand.Intn(20)+1) * time.Second,
			}}, &reply); err != nil {
			t.Error(err)
		}

	})
	time.Sleep(1 * time.Second)
	t.Run("TestGetTrend", func(t *testing.T) {
		var tr engine.Trend
		if err := client.Call(context.Background(), utils.TrendSv1GetTrend, &utils.ArgGetTrend{ID: "TREND_1", TenantWithAPIOpts: utils.TenantWithAPIOpts{Tenant: "cgrates.org"}}, &tr); err != nil {
			t.Error(err)
		} else if len(tr.RunTimes) != 1 && len(tr.Metrics) != 1 {
			t.Error("expected metrics to be calculated")
		}
	})

	t.Run("ProcessStats", func(t *testing.T) {
		var reply []string
		if err := client.Call(context.Background(), utils.StatSv1ProcessEvent, &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     fmt.Sprintf("event%d", 2),
			Event: map[string]any{
				utils.AccountField: "1001",
				utils.AnswerTime:   time.Date(2024, 9, 22, 14, 25, 0, 0, time.UTC),
				utils.Usage:        time.Duration(rand.Intn(3600)+60) * time.Second / 2,
				utils.Cost:         rand.Float64() * 30,
				utils.PDD:          time.Duration(rand.Intn(20)+4) * time.Second,
			}}, &reply); err != nil {
			t.Error(err)
		}
	})

	time.Sleep(1 * time.Second)
	t.Run("TestGetTrend", func(t *testing.T) {
		var tr engine.Trend
		if err := client.Call(context.Background(), utils.TrendSv1GetTrend, &utils.ArgGetTrend{ID: "TREND_1", TenantWithAPIOpts: utils.TenantWithAPIOpts{Tenant: "cgrates.org"}}, &tr); err != nil {
			t.Error(err)
		} else if len(tr.RunTimes) != 2 && len(tr.Metrics) != 2 {
			t.Error("expected metrics to be calculated")
		} else if tr.Metrics[tr.RunTimes[1]]["*acd"].TrendLabel != utils.MetaNegative {
			t.Error("expected TrendLabel to be negative")
		} else if tr.Metrics[tr.RunTimes[1]]["*tcc"].TrendLabel != utils.MetaPositive {
			t.Error("expected TrendLabel to be positive")
		} else if tr.Metrics[tr.RunTimes[1]]["*tcd"].TrendLabel != utils.MetaPositive {
			t.Error("expected TrendLabel to be positive")
		}
	})
}
