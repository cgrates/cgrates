//go:build kafka

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
	"net"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/twmb/franz-go/pkg/kgo"
)

func TestKafkaExportFailover(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

	brokerURL := "localhost:9092"
	topic := fmt.Sprintf("failover-%d", time.Now().UnixNano())
	failedDir := t.TempDir()

	createKafkaTopics(t, brokerURL, true, topic)

	content := fmt.Sprintf(`{
"apiers": {
    "enabled": true
},
"ees": {
    "enabled": true,
    "failed_posts": {
        "ttl": "1s"
    },
    "exporters": [
        {
            "id": "kafka_failover",
            "type": "*kafka_json_map",
            "export_path": "%s",
            "synchronous": true,
            "attempts": 1,
            "failed_posts_dir": "%s",
            "opts": {
                "kafkaTopic": "%s",
                "kafkaDeliveryTimeout": "2s"
            }
        }
    ]
}
}`, brokerURL, failedDir, topic)

	ng := engine.TestEngine{
		ConfigJSON: content,
		DBCfg:      engine.InternalDBCfg,
	}
	client, _ := ng.Run(t)

	t.Run("export before outage", func(t *testing.T) {
		for i := 1; i <= 3; i++ {
			if err := exportKafkaEvent(client, fmt.Sprintf("initial-%d", i)); err != nil {
				t.Fatalf("initial-%d: %v", i, err)
			}
		}
		records := consumeKafkaN(t, brokerURL, topic, 3, 5*time.Second)
		if len(records) != 3 {
			t.Fatalf("got %d records, want 3", len(records))
		}
	})

	t.Run("exports during outage fail", func(t *testing.T) {
		stopKafka(t)
		time.Sleep(2 * time.Second) // let franz-go detect the lost connection
		for i := 1; i <= 3; i++ {
			start := time.Now()
			err := exportKafkaEvent(client, fmt.Sprintf("blocked-%d", i))
			elapsed := time.Since(start)
			if err == nil {
				t.Errorf("blocked-%d: expected error, got nil", i)
			}
			if elapsed > 10*time.Second {
				t.Errorf("blocked-%d took %v, want <10s", i, elapsed)
			}
		}
	})

	t.Run("no unexpected records after recovery", func(t *testing.T) {
		startKafka(t, brokerURL)
		time.Sleep(5 * time.Second) // let any stale records arrive

		records := consumeKafkaAll(t, brokerURL, topic, 3*time.Second)
		var unexpected int
		for _, r := range records {
			if strings.Contains(string(r.Value), `"blocked-`) {
				unexpected++
			}
		}
		if unexpected > 0 {
			t.Errorf("got %d unexpected records after recovery, want 0", unexpected)
		}
	})

	t.Run("replay failed posts", func(t *testing.T) {
		time.Sleep(2 * time.Second) // wait for failed_posts cache to flush (1s TTL)

		entries, err := os.ReadDir(failedDir)
		if err != nil {
			t.Fatal(err)
		}
		if len(entries) == 0 {
			t.Fatal("no failed post files found")
		}

		var reply string
		if err := client.Call(context.Background(), "APIerSv1.ReplayFailedPosts",
			map[string]any{"SourcePath": failedDir, "FailedPath": "*none"}, &reply); err != nil {
			t.Fatal(err)
		}

		records := consumeKafkaAll(t, brokerURL, topic, 3*time.Second)
		if len(records) != 6 {
			t.Errorf("got %d records after replay, want 6 (3 initial + 3 replayed)", len(records))
		}
	})

	t.Run("new exports after recovery arrive", func(t *testing.T) {
		for i := 1; i <= 3; i++ {
			var err error
			for attempt := 0; attempt < 3; attempt++ {
				if err = exportKafkaEvent(client, fmt.Sprintf("recovery-%d", i)); err == nil {
					break
				}
				time.Sleep(2 * time.Second)
			}
			if err != nil {
				t.Fatalf("recovery-%d: %v", i, err)
			}
		}
		records := consumeKafkaAll(t, brokerURL, topic, 3*time.Second)
		if len(records) != 9 {
			t.Errorf("got %d total records, want 9 (3 initial + 3 replayed + 3 recovery)", len(records))
		}
	})
}

func exportKafkaEvent(client *birpc.Client, originID string) error {
	var reply map[string]map[string]any
	return client.Call(context.Background(), utils.EeSv1ProcessEvent,
		&engine.CGREventWithEeIDs{
			EeIDs: []string{"kafka_failover"},
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "FailoverTest_" + originID,
				Event: map[string]any{
					utils.ToR:          utils.MetaVoice,
					utils.OriginID:     originID,
					utils.AccountField: "1001",
					utils.Destination:  "1002",
					utils.Usage:        10 * time.Second,
				},
			},
		}, &reply)
}

func consumeKafkaN(tb testing.TB, brokerURL, topic string, n int, timeout time.Duration) []*kgo.Record {
	tb.Helper()
	cl, err := kgo.NewClient(
		kgo.SeedBrokers(brokerURL),
		kgo.ConsumeTopics(topic),
		kgo.FetchMaxWait(10*time.Millisecond),
	)
	if err != nil {
		tb.Fatal(err)
	}
	defer cl.Close()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	var records []*kgo.Record
	for len(records) < n {
		fetches := cl.PollFetches(ctx)
		fetches.EachRecord(func(r *kgo.Record) {
			records = append(records, r)
		})
		if ctx.Err() != nil {
			tb.Fatalf("timed out after consuming %d/%d records", len(records), n)
		}
	}
	return records
}

func consumeKafkaAll(tb testing.TB, brokerURL, topic string, timeout time.Duration) []*kgo.Record {
	tb.Helper()
	cl, err := kgo.NewClient(
		kgo.SeedBrokers(brokerURL),
		kgo.ConsumeTopics(topic),
		kgo.FetchMaxWait(10*time.Millisecond),
	)
	if err != nil {
		tb.Fatal(err)
	}
	defer cl.Close()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	var records []*kgo.Record
	for ctx.Err() == nil {
		fetches := cl.PollFetches(ctx)
		fetches.EachRecord(func(r *kgo.Record) {
			records = append(records, r)
		})
	}
	return records
}

func stopKafka(t *testing.T) {
	t.Helper()
	out, err := exec.Command("podman", "stop", "-t", "0", "cgrates_kafka_1").CombinedOutput()
	if err != nil {
		t.Fatalf("stop kafka: %s: %v", out, err)
	}
	t.Cleanup(func() {
		_ = exec.Command("podman", "start", "cgrates_kafka_1").Run()
	})
}

func startKafka(t *testing.T, brokerURL string) {
	t.Helper()
	out, err := exec.Command("podman", "start", "cgrates_kafka_1").CombinedOutput()
	if err != nil {
		t.Fatalf("start kafka: %s: %v", out, err)
	}
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		conn, dialErr := net.DialTimeout("tcp", brokerURL, time.Second)
		if dialErr == nil {
			conn.Close()
			time.Sleep(2 * time.Second)
			return
		}
		time.Sleep(500 * time.Millisecond)
	}
	t.Fatal("kafka did not become ready within 30s")
}

// Setup: podman compose up -d kafka
//
//	go test -tags=kafka ./general_tests/ -run TestKafkaExportFailover -dbtype "*internal" -v

/*
Compose service (kafka):

kafka:
  image: docker.io/apache/kafka:4.2.0
  ports: ["9092:9092"]
  environment:
    KAFKA_NODE_ID: 1
    KAFKA_PROCESS_ROLES: broker,controller
    KAFKA_LISTENERS: PLAINTEXT://0.0.0.0:9092,CONTROLLER://0.0.0.0:9093
    KAFKA_ADVERTISED_LISTENERS: PLAINTEXT://localhost:9092
    KAFKA_CONTROLLER_QUORUM_VOTERS: 1@localhost:9093
    KAFKA_CONTROLLER_LISTENER_NAMES: CONTROLLER
    KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR: 1
    CLUSTER_ID: cgrates-dev-cluster-01
*/
