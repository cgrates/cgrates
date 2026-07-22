//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package general_tests

import (
	"fmt"
	"path"
	"slices"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestRankingSchedule(t *testing.T) {
	var rnkConfigDir string
	switch *utils.DBType {
	case utils.MetaInternal, utils.MetaPostgres:
		t.SkipNow()
	case utils.MetaMySQL:
		rnkConfigDir = "ranking_mysql"
	case utils.MetaMongo:
		rnkConfigDir = "ranking_mongo"
	default:
		t.Fatal("Unkwown database type")
	}

	ng := engine.TestEngine{
		ConfigPath:     path.Join(*utils.DataDir, "conf", "samples", rnkConfigDir),
		PreserveStorDB: true,
		PreserveDataDB: true,
		PreStartHook: func(t testing.TB, c *config.CGRConfig) {
			engine.FlushDBs(t, c, true, true)
			engine.LoadCSVsWithCGRLoader(t, c.ConfigPath, path.Join(*utils.DataDir, "tariffplans", "tutrankings"), nil, nil, "-caches_address=")
		},
	}
	client, _ := ng.Run(t)

	t.Run("ProcessStats", func(t *testing.T) {
		var reply []string
		for i := 1; i <= 4; i++ {
			if err := client.Call(context.Background(), utils.StatSv1ProcessEvent, &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     fmt.Sprintf("event%d", i),
				Event: map[string]any{
					utils.AccountField: fmt.Sprintf("100%d", i),
					utils.AnswerTime:   time.Date(2024, 8, 22, 14, 25, 0, 0, time.UTC),
					utils.Usage:        time.Duration(1800+60) / time.Duration(i) * time.Second,
					utils.Cost:         20.0 + float64((i*7)%10)/2,
					utils.PDD:          time.Duration(10+i*2) * time.Second,
				}}, &reply); err != nil {
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

	t.Run("SetRankingProfile", func(t *testing.T) {
		rankingProfile := &engine.RankingProfileWithAPIOpts{
			RankingProfile: &engine.RankingProfile{
				Tenant:            "cgrates.org",
				ID:                "RANK3",
				Schedule:          "@every 1s",
				StatIDs:           []string{"Stats2", "Stats3", "Stats1"},
				Sorting:           "*desc",
				SortingParameters: []string{"*acd", "*pdd:false", "*acc"},
			},
		}
		var result string
		if err := client.Call(context.Background(), utils.APIerSv1SetRankingProfile, rankingProfile, &result); err != nil {
			t.Error(err)
		} else if result != utils.OK {
			t.Error("Unexpected reply returned", result)
		}
	})

	t.Run("ScheduleManualRankings", func(t *testing.T) {
		var scheduled int
		if err := client.Call(context.Background(), utils.RankingSv1ScheduleQueries, &utils.ArgScheduleRankingQueries{TenantIDWithAPIOpts: utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org"}}, RankingIDs: []string{"RANK3"}}, &scheduled); err != nil {
			t.Error(err)
		} else if scheduled != 1 {
			t.Errorf("Expected 1 scheduled rankings, got %d", scheduled)
		}
		var schedRankings []utils.ScheduledRanking
		if err := client.Call(context.Background(), utils.RankingSv1GetSchedule, &utils.ArgScheduledRankings{RankingIDPrefixes: []string{"RANK3"}, TenantIDWithAPIOpts: utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org"}}}, &schedRankings); err != nil {
			t.Error(err)
		} else if len(schedRankings) == 0 {
			t.Errorf("Expected  scheduled rankings")
		} else {
			for _, schedRanking := range schedRankings {
				if schedRanking.Next.IsZero() {
					t.Errorf("Expected to have a scheduled time got %v", schedRanking.Next)
				}
			}
		}
	})
	time.Sleep(time.Second)
	t.Run("GetRankings", func(t *testing.T) {
		sortedStatIds := []string{"Stats1", "Stats2", "Stats3"}
		var rnk engine.Ranking
		if err := client.Call(context.Background(), utils.RankingSv1GetRanking, &utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "RANK3"}}, &rnk); err != nil {
			t.Error(err)
		} else if len(rnk.SortedStatIDs) == 0 {
			t.Error("no ranked statIDs")
		} else if !slices.Equal(sortedStatIds, rnk.SortedStatIDs) {
			t.Errorf("expected sorted statIDs %v, got %v", sortedStatIds, rnk.SortedStatIDs)
		}
	})

	t.Run("RankingSetConfig", func(t *testing.T) {
		var reply string
		// store interval is set to 0
		if err := client.Call(context.Background(), utils.ConfigSv1SetConfig, &config.SetConfigArgs{
			Tenant: "cgrates.org",
			Config: map[string]any{
				"rankings": map[string]any{
					"enabled":        true,
					"store_interval": "0",
					"stats_conns":    []string{"*localhost"},
				},
			},
		}, &reply); err != nil {
			t.Error(err)
		} else if reply != utils.OK {
			t.Errorf("Expected OK received: %s", reply)
		}
	})
	t.Run("SetRankingProfile", func(t *testing.T) {
		rankingProfile := &engine.RankingProfileWithAPIOpts{
			RankingProfile: &engine.RankingProfile{
				Tenant:            "cgrates.org",
				ID:                "RANK4",
				Schedule:          "@every 1s",
				StatIDs:           []string{"Stats2", "Stats3", "Stats1"},
				Sorting:           "*desc",
				SortingParameters: []string{"*pdd:false", "*acd"},
			},
		}
		var result string
		if err := client.Call(context.Background(), utils.APIerSv1SetRankingProfile, rankingProfile, &result); err != nil {
			t.Error(err)
		} else if result != utils.OK {
			t.Error("Unexpected reply returned", result)
		}
	})
	t.Run("ScheduleManualRankings", func(t *testing.T) {
		var scheduled int
		if err := client.Call(context.Background(), utils.RankingSv1ScheduleQueries, &utils.ArgScheduleRankingQueries{TenantIDWithAPIOpts: utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org"}}, RankingIDs: []string{"RANK4"}}, &scheduled); err != nil {
			t.Error(err)
		} else if scheduled != 1 {
			t.Errorf("Expected 1 scheduled rankings, got %d", scheduled)
		}
		var schedRankings []utils.ScheduledRanking
		if err := client.Call(context.Background(), utils.RankingSv1GetSchedule, &utils.ArgScheduledRankings{RankingIDPrefixes: []string{"RANK4"}, TenantIDWithAPIOpts: utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org"}}}, &schedRankings); err != nil {
			t.Error(err)
		} else if len(schedRankings) == 0 {
			t.Errorf("Expected  scheduled rankings")
		} else {
			for _, schedRanking := range schedRankings {
				if schedRanking.Next.IsZero() {
					t.Errorf("Expected to have a scheduled time got %v", schedRanking.Next)
				}
			}
		}
	})
	time.Sleep(time.Second)
	t.Run("GetRankings", func(t *testing.T) {
		var reply string
		if err := client.Call(context.Background(), utils.CacheSv1Clear, &utils.AttrCacheIDsWithAPIOpts{}, &reply); err != nil {
			t.Error(err)
		} else if reply != utils.OK {
			t.Error("Calling CacheSv1.ReloadCache got reply: ", reply)
		}

		var rnk engine.Ranking
		if err := client.Call(context.Background(), utils.RankingSv1GetRanking, &utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "RANK4"}}, &rnk); err != nil {
			t.Error(err)
		} else if len(rnk.SortedStatIDs) != 0 || !rnk.LastUpdate.IsZero() {
			t.Error("expected updated ranking not set in db")
		}
	})
}
