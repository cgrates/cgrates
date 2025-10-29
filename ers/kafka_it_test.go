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

package ers

import (
	"fmt"
	"net"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	kafka "github.com/segmentio/kafka-go"
)

var (
	rdrEvents chan *erEvent
	rdrErr    chan error
	rdrExit   chan struct{}
	rdr       EventReader
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
			Topic:             utils.KafkaDefaultTopic,
			NumPartitions:     1,
			ReplicationFactor: 1,
		},
	}

	err = controllerConn.CreateTopics(topicConfigs...)
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := config.NewCGRConfigFromJSONStringWithDefaults(`{
"ers": {									
	"enabled": true,						
	"sessions_conns":["*localhost"],
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
				{"tag": "CGRID", "type": "*composed", "value": "~*req.CGRID", "path": "*cgreq.CGRID"},
				{"tag": "readerId", "type": "*variable", "value": "~*vars.*readerID", "path": "*cgreq.ReaderID"},
			]
		}
	]
}
}`)
	if err != nil {
		t.Fatal(err)
	}
	if err := cfg.CheckConfigSanity(); err != nil {
		t.Fatal(err)
	}
	rdrEvents = make(chan *erEvent, 1)
	rdrErr = make(chan error, 1)
	rdrExit = make(chan struct{}, 1)

	if rdr, err = NewKafkaER(cfg, 1, rdrEvents, make(chan *erEvent, 1),
		rdrErr, new(engine.FilterS), rdrExit); err != nil {
		t.Fatal(err)
	}
	rdr.Serve()

	randomCGRID := utils.UUIDSha1Prefix()
	go func(key string) {
		w := kafka.Writer{
			Addr:  kafka.TCP("localhost:9092"),
			Topic: utils.KafkaDefaultTopic,
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
	case err = <-rdrErr:
		t.Error(err)
	case ev := <-rdrEvents:
		if ev.rdrCfg.ID != "kafka" {
			t.Errorf("expected %s, received %s", "kafka", ev.rdrCfg.ID)
		}
		expected := &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     ev.cgrEvent.ID,
			Time:   ev.cgrEvent.Time,
			Event: map[string]any{
				"CGRID":    randomCGRID,
				"ReaderID": cfg.ERsCfg().Readers[1].ID,
			},
			APIOpts: map[string]any{},
		}
		if !reflect.DeepEqual(ev.cgrEvent, expected) {
			t.Errorf("Expected %s ,received %s", utils.ToJSON(expected), utils.ToJSON(ev.cgrEvent))
		}
	case <-time.After(30 * time.Second):
		t.Fatal("Timeout")
	}
	close(rdrExit)

	// Delete kafka topic
	partitions, err := conn.ReadPartitions(utils.KafkaDefaultTopic)
	if err != nil {
		t.Fatal(err)
	}

	if len(partitions) != 1 || partitions[0].Topic != utils.KafkaDefaultTopic {
		t.Fatal("expected topic named cgrates to exist")
	}

	if err := conn.DeleteTopics(utils.KafkaDefaultTopic); err != nil {
		t.Fatal(err)
	}

	experr := `[3] Unknown Topic Or Partition: the request is for a topic or partition that does not exist on this broker`
	_, err = conn.ReadPartitions(utils.KafkaDefaultTopic)
	if err == nil || err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}
