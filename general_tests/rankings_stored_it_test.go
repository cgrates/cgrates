//go:build integration
// +build integration

/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/
package general_tests

import (
	"fmt"
	"os"
	"slices"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestRankingStore(t *testing.T) {
	var dbConfig engine.DBCfg
	switch *utils.DBType {
	case utils.MetaInternal:
		dbConfig = engine.DBCfg{
			DB: &engine.DBParams{
				DBConns: map[string]engine.DBConn{
					utils.MetaDefault: {
						Type: utils.StringPointer(utils.MetaInternal),
						Opts: engine.Opts{
							InternalDBDumpPath: utils.StringPointer("/tmp/internal_db"),
						},
					},
				},
			},
		}
		if err := os.MkdirAll("/tmp/internal_db", 0755); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			if err := os.RemoveAll("/tmp/internal_db"); err != nil {
				t.Error(err)
			}
		})
	case utils.MetaRedis:
		dbConfig = engine.RedisDBCfg
	case utils.MetaMySQL:
		dbConfig = engine.DBCfg{
			DB: &engine.DBParams{
				DBConns: map[string]engine.DBConn{
					utils.MetaDefault: {
						Type: utils.StringPointer(utils.MetaRedis),
						Host: utils.StringPointer("127.0.0.1"),
						Port: utils.IntPointer(6379),
						Name: utils.StringPointer("10"),
						User: utils.StringPointer(utils.CGRateSLwr),
					},
					utils.StorDB: {
						Type:     utils.StringPointer(utils.MetaMySQL),
						Host:     utils.StringPointer("127.0.0.1"),
						Port:     utils.IntPointer(3306),
						Name:     utils.StringPointer(utils.CGRateSLwr),
						User:     utils.StringPointer(utils.CGRateSLwr),
						Password: utils.StringPointer("CGRateS.org"),
					},
				},
				Items: map[string]engine.Item{
					utils.MetaCDRs: {
						Limit:  utils.IntPointer(-1),
						DbConn: utils.StringPointer(utils.StorDB),
					},
					utils.MetaStatQueueProfiles: {
						Limit:  utils.IntPointer(-1),
						DbConn: utils.StringPointer(utils.StorDB),
					},
					utils.MetaStatQueues: {
						Limit:  utils.IntPointer(-1),
						DbConn: utils.StringPointer(utils.StorDB),
					},
					utils.MetaRankingProfiles: {
						Limit:  utils.IntPointer(-1),
						DbConn: utils.StringPointer(utils.StorDB),
					},
					utils.MetaRankings: {
						Limit:  utils.IntPointer(-1),
						DbConn: utils.StringPointer(utils.StorDB),
					},
				},
			},
		}
	case utils.MetaMongo:
		dbConfig = engine.MongoDBCfg
	case utils.MetaPostgres:
		dbConfig = engine.DBCfg{
			DB: &engine.DBParams{
				DBConns: map[string]engine.DBConn{
					utils.MetaDefault: {
						Type: utils.StringPointer(utils.MetaRedis),
						Host: utils.StringPointer("127.0.0.1"),
						Port: utils.IntPointer(6379),
						Name: utils.StringPointer("10"),
						User: utils.StringPointer(utils.CGRateSLwr),
					},
					utils.StorDB: {
						Type:     utils.StringPointer(utils.MetaPostgres),
						Host:     utils.StringPointer("127.0.0.1"),
						Port:     utils.IntPointer(5432),
						Name:     utils.StringPointer(utils.CGRateSLwr),
						User:     utils.StringPointer(utils.CGRateSLwr),
						Password: utils.StringPointer("CGRateS.org"),
					},
				},
				Items: map[string]engine.Item{
					utils.MetaCDRs: {
						Limit:  utils.IntPointer(-1),
						DbConn: utils.StringPointer(utils.StorDB),
					},
					utils.MetaStatQueueProfiles: {
						Limit:  utils.IntPointer(-1),
						DbConn: utils.StringPointer(utils.StorDB),
					},
					utils.MetaStatQueues: {
						Limit:  utils.IntPointer(-1),
						DbConn: utils.StringPointer(utils.StorDB),
					},
					utils.MetaRankingProfiles: {
						Limit:  utils.IntPointer(-1),
						DbConn: utils.StringPointer(utils.StorDB),
					},
					utils.MetaRankings: {
						Limit:  utils.IntPointer(-1),
						DbConn: utils.StringPointer(utils.StorDB),
					},
				},
			},
		}
	default:
		t.Fatal("unsupported dbtype value")
	}
	content := `{

"logger": {
	"level": 7
},
"rankings": {
	"enabled": true,
	"store_interval": "1500ms",
	"stats_conns":["*localhost"],
},

"stats": {
	"enabled": true,
	"store_interval": "-1",
},

"admins": {
	"enabled": true,
},

}
`
	tpFiles := map[string]string{
		utils.RankingsCsv: `#Tenant[0],Id[1],Schedule[2],StatIDs[3],MetricIDs[4],Sorting[5],SortingParameters[6],Stored[7],ThresholdIDs[8]
cgrates.org,RANK1,@every 1s,Stats1;Stats2;Stats3;Stats4,,*asc,*acc;*pdd:false;*acd,,`,
		utils.StatsCsv: `#Tenant[0],Id[1],FilterIDs[2],Weights[3],Blockers[4],QueueLength[5],TTL[6],MinItems[7],Stored[8],ThresholdIDs[9],MetricIDs[10],MetricFilterIDs[11],MetricBlockers[12]
cgrates.org,Stats1,*string:~*req.Account:1001,,,,-1,,,,*acc;*acd;*pdd,,
cgrates.org,Stats2,*string:~*req.Account:1002,,,,-1,,,,*acc;*acd;*pdd,,
cgrates.org,Stats3,*string:~*req.Account:1003,,,,-1,,,,*acc;*acd;*pdd,,
cgrates.org,Stats4,*string:~*req.Account:1004,,,,-1,,,,*acc;*acd;*pdd,,`}

	ng := engine.TestEngine{
		ConfigJSON: content,
		TpFiles:    tpFiles,
		DBCfg:      dbConfig,
		Encoding:   *utils.Encoding,
	}
	client, _ := ng.Run(t)
	time.Sleep(200 * time.Millisecond)
	var lastUpdate time.Time
	t.Run("RankingSchedule", func(t *testing.T) {
		var scheduled int
		if err := client.Call(context.Background(), utils.RankingSv1ScheduleQueries,
			&utils.ArgScheduleRankingQueries{RankingIDs: []string{"RANK1"}, TenantIDWithAPIOpts: utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org"}}}, &scheduled); err != nil {
			t.Fatal(err)
		} else if scheduled != 1 {
			t.Errorf("expected 1, got %d", scheduled)
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
				}, APIOpts: map[string]any{
					utils.MetaUsage: time.Duration(1800+60) / time.Duration(i) * time.Second,
					utils.MetaCost:  20.0 + float64((i*7)%10)/2,
					utils.MetaPDD:   time.Duration(10+i*2) * time.Second,
				}}, &reply); err != nil {
				t.Error(err)
			}
		}
	})

	t.Run("GetRankingsAfterStoreInterval", func(t *testing.T) {
		rankingsChan := make(chan *utils.Ranking, 1)
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		go func() {
			ticker := time.NewTicker(600 * time.Millisecond)
			var rnk utils.Ranking
			for {
				select {
				case <-ticker.C:
					err := client.Call(context.Background(), utils.RankingSv1GetRanking, &utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "RANK1"}}, &rnk)
					if err != nil {
						if err.Error() != utils.ErrNotFound.Error() {
							t.Errorf("Ranking retrieval error: %v", err)
						}
						continue
					} else if rnk.LastUpdate.IsZero() {
						continue
					}
					rankingsChan <- &rnk
				case <-ctx.Done():
					return
				}

			}
		}()

		select {
		case rnk := <-rankingsChan:
			lastUpdate = rnk.LastUpdate
			sortedStatIDs := []string{"Stats3", "Stats2", "Stats1", "Stats4"}
			if !slices.Equal(rnk.SortedStatIDs, sortedStatIDs) {
				t.Error("should have sorted statids")
			}
		case <-ctx.Done():
			t.Error("Didn't get any ranking from db")
		}
	})

	t.Run("ProcessStats", func(t *testing.T) {
		var reply []string
		j := 5
		for i := 1; i <= 4; i++ {
			j--
			if err := client.Call(context.Background(), utils.StatSv1ProcessEvent, &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     fmt.Sprintf("event%d", i),
				Event: map[string]any{
					utils.AccountField: fmt.Sprintf("100%d", i),
					utils.AnswerTime:   time.Date(2024, 8, 22, 14, 25, 0, 0, time.UTC),
				}, APIOpts: map[string]any{
					utils.MetaUsage: time.Duration(1500+60) / time.Duration(j) * time.Second,
					utils.MetaCost:  20.0 + float64((j*8)%10)/2,
					utils.MetaPDD:   time.Duration(11+j*2) * time.Second,
				}}, &reply); err != nil {
				t.Error(err)
			}
		}
	})

	t.Run("RankingsSetConfig", func(t *testing.T) {
		var reply string
		// setting store interval to 0
		if err := client.Call(context.Background(), utils.ConfigSv1SetConfig, &config.SetConfigArgs{
			Tenant: "cgrates.org",
			Config: map[string]any{
				"rankings": map[string]any{
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
	t.Run("RankingSchedule", func(t *testing.T) {
		var scheduled int
		if err := client.Call(context.Background(), utils.RankingSv1ScheduleQueries,
			&utils.ArgScheduleRankingQueries{RankingIDs: []string{"RANK1"}, TenantIDWithAPIOpts: utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org"}}}, &scheduled); err != nil {
			t.Fatal(err)
		} else if scheduled != 1 {
			t.Errorf("expected 1, got %d", scheduled)
		}
	})

	t.Run("GetRankingsNotStored", func(t *testing.T) {
		metricsChan := make(chan *utils.Ranking, 1)
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		go func() {
			for {
				ticker := time.NewTicker(700 * time.Millisecond)
				select {
				case <-ticker.C:
					var rnk utils.Ranking
					err := client.Call(context.Background(), utils.RankingSv1GetRanking, &utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "RANK1"}}, &rnk)
					if err != nil {
						if err.Error() != utils.ErrNotFound.Error() {
							t.Errorf("Ranking retrieval error: %v", err)
						}
						continue
					} else if rnk.LastUpdate.Equal(lastUpdate) {
						continue
					}
					metricsChan <- &rnk
				case <-ctx.Done():
					return
				}
			}
		}()

		select {
		case rnk := <-metricsChan:
			lastUpdate = rnk.LastUpdate
		case <-ctx.Done():
			t.Error("Didn't get any Ranking from db")
		}
	})
	t.Run("RankingsSetConfig", func(t *testing.T) {
		var reply string
		// setting store interval to -1
		if err := client.Call(context.Background(), utils.ConfigSv1SetConfig, &config.SetConfigArgs{
			Tenant: "cgrates.org",
			Config: map[string]any{
				"rankings": map[string]any{
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
	t.Run("RankingsSchedule", func(t *testing.T) {
		var scheduled int
		if err := client.Call(context.Background(), utils.RankingSv1ScheduleQueries,
			&utils.ArgScheduleRankingQueries{RankingIDs: []string{"RANK1"}, TenantIDWithAPIOpts: utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org"}}}, &scheduled); err != nil {
			t.Fatal(err)
		} else if scheduled != 1 {
			t.Errorf("expected 1, got %d", scheduled)
		}
	})
	t.Run("GetRankingsStoredUnlimited", func(t *testing.T) {
		rankingChan := make(chan *utils.Ranking, 1)
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		go func() {
			ticker := time.NewTicker(1000 * time.Millisecond)
			for {
				select {
				case <-ticker.C:
					var rnk utils.Ranking
					err := client.Call(context.Background(), utils.RankingSv1GetRanking, &utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "RANK1"}}, &rnk)
					if err != nil {
						if err.Error() != utils.ErrNotFound.Error() {
							t.Errorf("Ranking retrieval error: %v", err)
						}
						continue
					} else if rnk.LastUpdate.Equal(lastUpdate) {
						continue
					}
					rankingChan <- &rnk
				case <-ctx.Done():
					return
				}
			}
		}()
		select {
		case <-rankingChan:
		case <-ctx.Done():
			t.Error("Didn't get any Ranking from db")
		}
	})
}
