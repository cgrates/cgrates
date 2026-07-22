//go:build kafka

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package general_tests

import (
	"flag"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var parallelism = flag.Int("parallelism", 100, "goroutines per GOMAXPROCS")

// Setup: podman compose up -d kafka
//
//	go test -tags=kafka -bench BenchmarkKafkaExport -benchtime 10s ./general_tests/ -dbtype "*internal"
func BenchmarkKafkaExport(b *testing.B) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaMongo, utils.MetaPostgres:
		b.SkipNow()
	default:
		b.Fatal("unsupported dbtype value")
	}

	brokerURL := "localhost:9092"
	topic := fmt.Sprintf("bench-%d", time.Now().UnixNano())

	createKafkaTopics(b, brokerURL, true, topic)

	content := fmt.Sprintf(`{
"ees": {
    "enabled": true,
    "exporters": [
        {
            "id": "kafka_bench",
            "type": "*kafka_json_map",
            "export_path": "%s",
            "synchronous": true,
            "opts": {
                "kafkaTopic": "%s"
            },
            "failed_posts_dir": "*none"
        }
    ]
}
}`, brokerURL, topic)

	ng := engine.TestEngine{
		ConfigJSON: content,
		DBCfg:      engine.InternalDBCfg,
	}
	client, _ := ng.Run(b)

	var seq, completed atomic.Int64
	b.SetParallelism(*parallelism)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			n := seq.Add(1)
			var reply map[string]map[string]any
			if err := client.Call(context.Background(), utils.EeSv1ProcessEvent,
				&engine.CGREventWithEeIDs{
					EeIDs: []string{"kafka_bench"},
					CGREvent: &utils.CGREvent{
						Tenant: "cgrates.org",
						ID:     fmt.Sprintf("bench_%d", n),
						Event: map[string]any{
							utils.ToR:          utils.MetaVoice,
							utils.OriginID:     fmt.Sprintf("bench-%d", n),
							utils.AccountField: "1001",
							utils.Destination:  "1002",
							utils.Usage:        10 * time.Second,
						},
					},
				}, &reply); err != nil {
				b.Error(err)
			} else {
				completed.Add(1)
			}
		}
	})
	b.StopTimer()

	records := consumeKafkaN(b, brokerURL, topic, int(completed.Load()), 30*time.Second)
	if got, want := int64(len(records)), completed.Load(); got != want {
		b.Errorf("kafka records: got %d, want %d", got, want)
	}
	b.Logf("completed %d/%d exports", completed.Load(), int64(b.N))
}
