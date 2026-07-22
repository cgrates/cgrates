//go:build performance

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package general_tests

import (
	"bytes"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestStressTrendsProcessEvent(t *testing.T) {
	var dbConfig engine.DBCfg
	switch *utils.DBType {
	case utils.MetaMySQL:
	case utils.MetaMongo, utils.MetaPostgres, utils.MetaInternal:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
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
	"apiers": {
		"enabled": true
	},
	"stats": {
		"enabled": true,
		"store_interval": "-1",
	},
    "trends": {
    	"enabled": true,
    	"store_interval": "-1",
    	"stats_conns":["*localhost"],
    	"thresholds_conns": [],	
        "ees_conns": [],
    },
	}`

	ng := engine.TestEngine{
		ConfigJSON: content,
		DBCfg:      dbConfig,
		LogBuffer:  bytes.NewBuffer(nil),
	}
	//defer fmt.Println(ng.LogBuffer)
	client, _ := ng.Run(t)

	numProfiles := 200
	t.Run("SetStatAndTrendProfiles", func(t *testing.T) {
		var reply string
		for i := 1; i <= numProfiles; i++ {
			statConfig := &engine.StatQueueProfileWithAPIOpts{
				StatQueueProfile: &engine.StatQueueProfile{
					Tenant:      "cgrates.org",
					ID:          fmt.Sprintf("STAT_%v", i),
					FilterIDs:   []string{fmt.Sprintf("*string:~*req.Account:100%d", i)},
					QueueLength: -1,
					TTL:         1 * time.Hour,
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
			trendProfile := &engine.TrendProfileWithAPIOpts{
				TrendProfile: &engine.TrendProfile{
					Tenant:          "cgrates.org",
					ID:              fmt.Sprintf("Trend%d", i),
					ThresholdIDs:    []string{""},
					CorrelationType: utils.MetaAverage,
					TTL:             -1,
					QueueLength:     -1,
					StatID:          fmt.Sprintf("STAT_%d", i),
					Schedule:        "@every 1s",
				},
			}
			if err := client.Call(context.Background(), utils.APIerSv1SetTrendProfile, trendProfile, &reply); err != nil {
				t.Error(err)
			} else if reply != utils.OK {
				t.Errorf("Expected: %v,Received: %v", utils.OK, reply)
			}

			var scheduled int
			if err := client.Call(context.Background(), utils.TrendSv1ScheduleQueries,
				&utils.ArgScheduleTrendQueries{TrendIDs: []string{trendProfile.ID}, TenantIDWithAPIOpts: utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org"}}}, &scheduled); err != nil {
				t.Fatal(err)
			} else if scheduled != 1 {
				t.Errorf("expected 1, got %d", scheduled)
			}

		}
	})
	t.Run("ContinuousEventLoad", func(t *testing.T) {
		var wg sync.WaitGroup
		errCh := make(chan error, numProfiles)
		stopCh := make(chan struct{}) // To stop event generation after duration

		// Run event generation for 30 seconds
		go func() {
			time.Sleep(30 * time.Second)
			close(stopCh)
		}()

		start := time.Now()
		var evcount atomic.Int64
		for i := range 50 { // 50 concurrent workers
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()
				for {
					select {
					case <-stopCh:
						return
					default:
						// Generate events for all accounts cyclically
						accID := rand.Intn(numProfiles) + 1 // Random account matching a profile
						args := &engine.CGREventWithEeIDs{
							CGREvent: &utils.CGREvent{
								Tenant: "cgrates.org",
								ID:     utils.GenUUID(),
								Event: map[string]any{
									utils.AccountField: fmt.Sprintf("100%d", accID),
									utils.Usage:        rand.Intn(60) + 1,
									utils.Cost:         rand.Float64() * 10,
								},
							},
						}
						var reply []string
						if err := client.Call(context.Background(), utils.StatSv1ProcessEvent, args, &reply); err != nil {
							errCh <- fmt.Errorf("worker %d: %v", workerID, err)
						}
						evcount.Add(1)
					}
				}
			}(i)
		}
		// Wait for all workers to finish
		go func() {
			wg.Wait()
			close(errCh)
		}()

		// Collect errors
		for err := range errCh {
			t.Error(err)
		}
		t.Logf("Generated %v events for %v", evcount.Load(), time.Since(start))
	})
}
