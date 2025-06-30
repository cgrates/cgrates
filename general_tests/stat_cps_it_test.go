//go:build performance

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
	"flag"
	"fmt"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	count = flag.Int("event_rate", 2000, "event_rate")
)

func TestLoadStatsProcessEvent(t *testing.T) {
	var dbConfig engine.DBCfg
	switch *utils.DBType {
	case utils.MetaInternal:
		dbConfig = engine.DBCfg{
			DataDB: &engine.DBParams{
				Type: utils.StringPointer(utils.MetaInternal),
			},
			StorDB: &engine.DBParams{
				Type: utils.StringPointer(utils.MetaInternal),
			},
		}
	case utils.MetaMySQL:
	case utils.MetaMongo, utils.MetaPostgres:
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
	
	}`

	ng := engine.TestEngine{
		ConfigJSON: content,
		DBCfg:      dbConfig,
		LogBuffer:  bytes.NewBuffer(nil),
	}
	client, _ := ng.Run(t)
	t.Run("SetStatProfile", func(t *testing.T) {
		var reply string
		for i := 1; i <= 10; i++ {
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
		}
	})

	t.Run("StatExportEvent", func(t *testing.T) {
		ticker := time.NewTicker(time.Second / time.Duration(*count))
		defer ticker.Stop()
		jobs := make(chan int, *count)
		for i := 1; i <= *count; i++ {
			jobs <- i
		}
		close(jobs)
		numWrk := 50
		var wg sync.WaitGroup
		latencies := make(chan time.Duration, *count)
		totalCall := time.Now()
		for range numWrk {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := range jobs {
					<-ticker.C
					callStart := time.Now()
					args := &engine.CGREventWithEeIDs{
						CGREvent: &utils.CGREvent{
							Tenant: "cgrates.org",
							ID:     "voiceEvent",
							Time:   utils.TimePointer(time.Now()),
							Event: map[string]any{
								utils.AccountField: fmt.Sprintf("100%d", ((i-1)%10)+1),
								utils.AnswerTime:   utils.TimePointer(time.Now()),
								utils.Usage:        45,
								utils.Cost:         12.1,
							},
						},
					}
					var reply []string
					if err := client.Call(context.Background(), utils.StatSv1ProcessEvent, args, &reply); err != nil {
						t.Errorf("Error processing event %d: %v", i, err)
						continue
					}
					latencies <- time.Since(callStart)
				}
			}()
		}
		wg.Wait()
		totalDuration := time.Since(totalCall)
		close(latencies)
		latencySlice := make([]time.Duration, 0, *count)
		for latency := range latencies {
			latencySlice = append(latencySlice, latency)
		}
		actualThroughput := float64(len(latencySlice)) / totalDuration.Seconds()
		slices.Sort(latencySlice)
		t.Logf("Performance Load Test Results")
		t.Logf("Target Rate: %d events/sec", *count)
		t.Logf("Actual Throughput:  %.2f events/sec", actualThroughput)
		t.Logf("Total Duration:     %v", totalDuration)

		// Calculate percentiles
		p50Index := int(float64(len(latencySlice)) * 0.50)
		p90Index := int(float64(len(latencySlice)) * 0.90)
		p99Index := int(float64(len(latencySlice)) * 0.99)

		t.Logf("p50 Latency: %v", latencySlice[p50Index])
		t.Logf("p90 Latency: %v", latencySlice[p90Index])
		t.Logf("p99 Latency: %v", latencySlice[p99Index])

	})
}

func BenchmarkStatQueueCompress(b *testing.B) {
	sizes := []int{1000, 10000, 50000}
	origQueues := make(map[int][]engine.SQItem, len(sizes))
	now := time.Now()
	for _, size := range sizes {
		items := make([]engine.SQItem, size)
		for i := range size {
			t := now.Add(time.Duration(i) * time.Second)
			items[i] = engine.SQItem{EventID: fmt.Sprintf("e%d", i), ExpiryTime: &t}
		}
		origQueues[size] = items
	}

	for _, size := range sizes {
		size := size
		orig := origQueues[size]
		b.Run(fmt.Sprintf("Compress_N=%d", size), func(b *testing.B) {
			threshold := int64(size / 2)
			round := 2
			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				items := make([]engine.SQItem, len(orig))
				copy(items, orig)
				sq := &engine.StatQueue{
					Tenant:    "cgrates.org",
					ID:        fmt.Sprintf("Q-%d", size),
					SQItems:   items,
					SQMetrics: make(map[string]engine.StatMetric),
				}
				sq.Compress(threshold, round)
			}
		})
	}
}

func BenchmarkStatQueueExpand(b *testing.B) {
	sizes := []int{1000, 10000, 50000}
	origQueues := make(map[int][]engine.SQItem, len(sizes))
	now := time.Now()
	for _, size := range sizes {
		items := make([]engine.SQItem, size/10)
		for i := range items {
			t := now
			items[i] = engine.SQItem{EventID: fmt.Sprintf("e%d", i), ExpiryTime: &t}
		}
		origQueues[size] = items
	}
	for _, size := range sizes {
		size := size
		orig := origQueues[size]
		b.Run(fmt.Sprintf("Expand_N=%d", size), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				items := make([]engine.SQItem, len(orig))
				copy(items, orig)
				sq := &engine.StatQueue{
					Tenant:    "cgrates.org",
					ID:        fmt.Sprintf("Q-%d", size),
					SQItems:   items,
					SQMetrics: make(map[string]engine.StatMetric),
				}
				sq.Expand()
			}
		})
	}
}
