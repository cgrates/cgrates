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
	"slices"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestRankingSchedule(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unkwown database type")
	}

	content := `{
    "general": {
        "log_level": 7,
    },
    "data_db": {
        "db_type": "redis",
        "db_port": 6379,
        "db_name": "10",
    },
    "stor_db": {
        "db_password": "CGRateS.org",
    },
    "rankings": {
        "enabled": true,
        "store_interval": "-1",
        "scheduled_ids": {},
        "stats_conns": [
            "*localhost"
        ]
    },
    "stats": {
        "enabled": true,
        "store_interval": "-1",
    },
    "admins": {
        "enabled": true,
    }
}
`
	tpFiles := map[string]string{
		utils.RankingsCsv: `#Tenant[0],Id[1],Schedule[2],StatIDs[3],MetricIDs[4],Sorting[5],SortingParameters[6],Stored[7],ThresholdIDs[8]
cgrates.org,RANK1,@every 1s,Stats1;Stats2;Stats3;Stats4,,*asc,*acc;*pdd:false;*acd,,
cgrates.org,RANK2,@every 1s,Stats3;Stats4;Stats1;Stats2,,*desc,*acc;*pdd:false;*acd,,`,
		utils.StatsCsv: `#Tenant[0],Id[1],FilterIDs[2],Weights[3],Blockers[4],QueueLength[5],TTL[6],MinItems[7],Stored[8],ThresholdIDs[9],MetricIDs[10],MetricFilterIDs[11],MetricBlockers[12]
cgrates.org,Stats1,*string:~*req.Account:1001,;30,,,-1,0,,,*acc;*acd;*pdd,,
cgrates.org,Stats2,*string:~*req.Account:1002,;30,,,-1,0,,,*acc;*acd;*pdd,,
cgrates.org,Stats3,*string:~*req.Account:1003,;30,,,-1,0,,,*acc;*acd;*pdd,,
cgrates.org,Stats4,*string:~*req.Account:1004,;30,,,-1,0,,,*acc;*acd;*pdd,,`,
	}
	ng := engine.TestEngine{
		ConfigJSON: content,
		TpFiles:    tpFiles,
		Encoding:   *utils.Encoding,
	}
	client, _ := ng.Run(t)
	time.Sleep(100 * time.Millisecond)
	t.Run("ScheduleManualRankings", func(t *testing.T) {
		var scheduled int
		if err := client.Call(context.Background(), utils.RankingSv1ScheduleQueries, &utils.ArgScheduleRankingQueries{TenantIDWithAPIOpts: utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org"}}, RankingIDs: []string{"RANK1", "RANK2"}}, &scheduled); err != nil {
			t.Error(err)
		} else if scheduled != 2 {
			t.Errorf("Expected 1 scheduled rankings, got %d", scheduled)
		}
		var schedRankings []utils.ScheduledRanking
		if err := client.Call(context.Background(), utils.RankingSv1GetSchedule, &utils.ArgScheduledRankings{RankingIDPrefixes: []string{"RANK"}, TenantIDWithAPIOpts: utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org"}}}, &schedRankings); err != nil {
			t.Error(err)
		} else if len(schedRankings) != 2 {
			t.Errorf("Expected  scheduled rankings")
		} else {
			for _, schedRanking := range schedRankings {
				if schedRanking.Next.IsZero() {
					t.Errorf("Expected to have a scheduled time got %v", schedRanking.Next)
				}
			}
		}
	})

	t.Run("ProcessStats", func(t *testing.T) {
		var reply []string
		for i := 1; i <= 4; i++ {
			if err := client.Call(context.Background(), utils.StatSv1ProcessEvent, &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     fmt.Sprintf("event%d", i),
				Event: map[string]any{
					utils.AccountField: fmt.Sprintf("100%d", i),
					utils.AnswerTime:   time.Date(2024, 8, 22, 14, 25, 0, 0, time.UTC),
				},
				APIOpts: map[string]any{
					utils.MetaUsage: time.Duration(1800+60) / time.Duration(i) * time.Second,
					utils.MetaCost:  20.0 + float64((i*7)%10)/2,
					utils.MetaPDD:   time.Duration(10+i*2) * time.Second,
				},
			}, &reply); err != nil {
				t.Error(err)
			}
		}
	})
	time.Sleep(1 * time.Second)

	t.Run("GetRankings", func(t *testing.T) {
		var rnk engine.Ranking
		sortedStatIds := []string{"Stats3", "Stats2", "Stats1", "Stats4"}
		if err := client.Call(context.Background(), utils.RankingSv1GetRanking, &utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "RANK1"}}, &rnk); err != nil {
			t.Error(err)
		} else if len(rnk.SortedStatIDs) == 0 {
			t.Error("no ranked statIDs")
		} else if !slices.Equal(sortedStatIds, rnk.SortedStatIDs) {
			t.Errorf("expected sorted statIDs %v, got %v", sortedStatIds, rnk.SortedStatIDs)
		}

		sortedStatIds = []string{"Stats4", "Stats1", "Stats2", "Stats3"}
		if err := client.Call(context.Background(), utils.RankingSv1GetRanking, &utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "RANK2"}}, &rnk); err != nil {
			t.Error(err)
		} else if len(rnk.SortedStatIDs) == 0 {
			t.Error("no ranked statIDs")
		} else if !slices.Equal(sortedStatIds, rnk.SortedStatIDs) {
			t.Errorf("expected sorted statIDs %v, got %v", sortedStatIds, rnk.SortedStatIDs)
		}
	})

}
