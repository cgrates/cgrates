//go:build performance
// +build performance

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
	"bytes"
	"flag"
	"fmt"
	"slices"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var count = flag.Int("count", 1000, "Number of events to process in the load test")

func BenchmarkStressIPsAllocateIP(b *testing.B) {
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
		b.SkipNow()
	default:
		b.Fatal("unsupported dbtype value")
	}

	content := `{
  "admins": {
    "enabled": true
  },
  "ips": {
    "enabled": true,
    "indexed_selects": false
  }
}`

	ng := engine.TestEngine{
		ConfigJSON: content,
		DBCfg:      dbConfig,
		LogBuffer:  bytes.NewBuffer(nil),
		Encoding:   utils.MetaJSON,
	}
	client, _ := ng.Run(b)

	var reply string
	for i := 1; i <= 3; i++ {
		ipProfile := &utils.IPProfileWithAPIOpts{
			IPProfile: &utils.IPProfile{
				Tenant:    "cgrates.org",
				ID:        fmt.Sprintf("IP_PROF_%d", i),
				FilterIDs: []string{fmt.Sprintf("*string:~*req.Account:%d", i)},
				TTL:       10 * time.Minute,
				Pools: []*utils.IPPool{
					{
						ID:      "POOL_A",
						Range:   fmt.Sprintf("10.%d.0.0/16", i),
						Message: "Allocated by test",
					},
				},
			},
		}
		if err := client.Call(context.Background(), utils.AdminSv1SetIPProfile, ipProfile, &reply); err != nil {
			b.Fatalf("Failed to set IP profile: %v", err)
		}
	}

	var evIdx atomic.Int32
	b.SetParallelism(0)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			currentIdx := evIdx.Add(1)
			args := &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     fmt.Sprintf("event%d", currentIdx),
				Event: map[string]any{
					utils.AccountField: fmt.Sprintf("%d", currentIdx%3+1),
					utils.AnswerTime:   utils.TimePointer(time.Now()),
					utils.Usage:        10,
				},
				APIOpts: map[string]any{
					utils.OptsIPsAllocationID: fmt.Sprintf("alloc%d", currentIdx),
				},
			}

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			var reply utils.AllocatedIP
			if err := client.Call(ctx, utils.IPsV1AllocateIP, args, &reply); err != nil {
				b.Logf("Error processing event %d: %v", currentIdx, err)
			}
		}
	})
}

func TestStressIPsAuthorize(t *testing.T) {
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
			"log_level": 7
		},
		"stor_db": {
			"db_password": "CGRateS.org"
		},
		
        "admins": {
	       "enabled": true,
        },
		"ips": {
            "enabled": true,	
			"indexed_selects":false,
		},
	}`

	ng := engine.TestEngine{
		ConfigJSON: content,
		DBCfg:      dbConfig,
		LogBuffer:  bytes.NewBuffer(nil),
		Encoding:   utils.MetaGOB,
	}
	client, _ := ng.Run(t)

	t.Run("SetIPProfile", func(t *testing.T) {
		var reply string
		for i := 1; i <= 10; i++ {
			ipProfile := &utils.IPProfileWithAPIOpts{
				IPProfile: &utils.IPProfile{
					Tenant:    "cgrates.org",
					ID:        fmt.Sprintf("IP_PROF_%d", i),
					FilterIDs: []string{fmt.Sprintf("*string:~*req.Account:%d", i)},
					TTL:       10 * time.Minute,
					Pools: []*utils.IPPool{
						{
							ID:      "POOL_A",
							Range:   fmt.Sprintf("10.0.0.%d/32", i),
							Message: "Allocated by test",
						},
					},
				},
			}
			if err := client.Call(context.Background(), utils.AdminSv1SetIPProfile, ipProfile, &reply); err != nil {
				t.Fatalf("Failed to set IP profile: %v", err)
			}
		}
	})

	t.Run("IPsAuthorizeEvent", func(t *testing.T) {
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
					args := &utils.CGREvent{
						Tenant: "cgrates.org",
						ID:     utils.GenUUID(),
						Event: map[string]any{
							utils.AccountField: fmt.Sprintf("%d", ((i-1)%10)+1),
						},
						APIOpts: map[string]any{
							utils.OptsIPsAllocationID: utils.GenUUID(),
						},
					}
					var reply utils.AllocatedIP
					if err := client.Call(context.Background(), utils.IPsV1AuthorizeIP, args, &reply); err != nil {
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

		successfulCalls := len(latencySlice)
		if successfulCalls == 0 {
			t.Fatal("No calls succeeded, cannot calculate performance.")
		}

		actualThroughput := float64(successfulCalls) / totalDuration.Seconds()
		slices.Sort(latencySlice)

		t.Logf("--- IP Allocation Performance Load Test Results ---")
		t.Logf("Target Rate:       %d events/sec", *count)
		t.Logf("Successful Calls:  %d", successfulCalls)
		t.Logf("Actual Throughput: %.2f events/sec", actualThroughput)
		t.Logf("Total Duration:    %v", totalDuration)

		// Calculate percentiles
		p50Index := int(float64(successfulCalls) * 0.50)
		p90Index := int(float64(successfulCalls) * 0.90)
		p99Index := int(float64(successfulCalls) * 0.99)

		t.Logf("p50 Latency: %v", latencySlice[p50Index])
		t.Logf("p90 Latency: %v", latencySlice[p90Index])
		t.Logf("p99 Latency: %v", latencySlice[p99Index])
	})

}
