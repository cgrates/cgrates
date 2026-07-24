//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package ers

import (
	"context"
	"fmt"
	"net"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	kafka "github.com/segmentio/kafka-go"
)

var (
	kfkEvents chan *erEvent
	kfkErr    chan error
	kfkExit   chan struct{}
	kfk       EventReader
)

func TestKafkaER(t *testing.T) {
	// Create kafka topic

	conn, err := kafka.Dial("tcp", "localhost:9092")
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		t.Fatal(err)
	}
	controllerConn, err := kafka.Dial("tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	if err != nil {
		t.Fatal(err)
	}
	defer controllerConn.Close()

	topicConfigs := []kafka.TopicConfig{
		{
			Topic:             "cgrates",
			NumPartitions:     1,
			ReplicationFactor: 1,
		},
	}

	err = controllerConn.CreateTopics(topicConfigs...)
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := config.NewCGRConfigFromJsonStringWithDefaults(`{
"ers": {									
	"enabled": true,						
	"readers": [
		{
			"id": "kafka",										
			"type": "*kafka_json_map",							
			"run_delay":  "-1",									
			"concurrent_requests": 1024,						
			"source_path": "localhost:9092",	
			"tenant": "cgrates.org",							
			"filters": [],										
			"flags": [],										
			"fields":[									
				{"tag": "CGRID", "type": "*composed", "value": "~*req.CGRID", "path": "*cgreq.CGRID"}
			]
		}
	]
}
}`)
	if err != nil {
		t.Fatal(err)
	}

	kfkEvents = make(chan *erEvent, 1)
	kfkErr = make(chan error, 1)
	kfkExit = make(chan struct{}, 1)

	if kfk, err = NewKafkaER(cfg, 1, kfkEvents,
		kfkErr, new(engine.FilterS), kfkExit); err != nil {
		t.Fatal(err)
	}
	kfk.Serve()

	randomCGRID := utils.UUIDSha1Prefix()
	go func(key string) {
		w := kafka.Writer{
			Addr:  kafka.TCP("localhost:9092"),
			Topic: defaultTopic,
		}
		err := w.WriteMessages(context.Background(),
			kafka.Message{
				Key:   []byte(randomCGRID), // for the moment we do not process the key
				Value: []byte(fmt.Sprintf(`{"CGRID": "%s"}`, randomCGRID)),
			},
		)
		if err != nil {
			t.Error("failed to write messages:", err)
		}
		err = w.Close()
		if err != nil {
			t.Error("failed to close writer:", err)
		}
	}(randomCGRID)

	select {
	case err = <-kfkErr:
		t.Error(err)
	case ev := <-kfkEvents:
		if ev.rdrCfg.ID != "kafka" {
			t.Errorf("Expected 'kakfa' received `%s`", ev.rdrCfg.ID)
		}
		expected := &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     ev.cgrEvent.ID,
			Time:   ev.cgrEvent.Time,
			Event: map[string]any{
				"CGRID": randomCGRID,
			},
		}
		if !reflect.DeepEqual(ev.cgrEvent, expected) {
			t.Errorf("Expected %s ,received %s", utils.ToJSON(expected), utils.ToJSON(ev.cgrEvent))
		}
	case <-time.After(10 * time.Second):
		t.Fatal("Timeout")
	}
	kfkExit <- struct{}{}

	// Delete kafka topic

	partitions, err := conn.ReadPartitions("cgrates")
	if err != nil {
		t.Fatal(err)
	}

	if len(partitions) != 1 || partitions[0].Topic != "cgrates" {
		t.Fatal("expected topic named cgrates to exist")
	}

	if err := conn.DeleteTopics("cgrates"); err != nil {
		t.Fatal(err)
	}

	experr := `[3] Unknown Topic Or Partition: the request is for a topic or partition that does not exist on this broker`
	_, err = conn.ReadPartitions("cgrates")
	if err == nil || err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}
