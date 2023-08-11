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
	rdrEvents chan *erEvent
	rdrErr    chan error
	rdrExit   chan struct{}
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
			Topic:             "cgrates_cdrs",
			NumPartitions:     1,
			ReplicationFactor: 1,
		},
	}

	err = controllerConn.CreateTopics(topicConfigs...)
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := config.NewCGRConfigFromJsonStringWithDefaults(`{
"ers": {									// EventReaderService
	"enabled": true,						// starts the EventReader service: <true|false>
	"readers": [
		{
			"id": "kafka",										// identifier of the EventReader profile
			"type": "*kafka_json_map",							// reader type <*file_csv>
			"run_delay":  "-1",									// sleep interval in seconds between consecutive runs, -1 to use automation via inotify or 0 to disable running all together
			"concurrent_requests": 1024,						// maximum simultaneous requests/files to process, 0 for unlimited
			"source_path": "localhost:9092",					// read data from this path
			// "processed_path": "/var/spool/cgrates/ers/out",	// move processed data here
			"tenant": "cgrates.org",							// tenant used by import
			"filters": [],										// limit parsing based on the filters
			"flags": [],										// flags to influence the event processing
			"fields":[									// import fields template, tag will match internally CDR field, in case of .csv value will be represented by index of the field value
				{"tag": "CGRID", "type": "*composed", "value": "~*req.CGRID", "path": "*cgreq.CGRID"},
			],
		},
	],
},
}`)
	if err != nil {
		t.Fatal(err)
	}

	rdrEvents = make(chan *erEvent, 1)
	rdrErr = make(chan error, 1)
	rdrExit = make(chan struct{}, 1)

	if kfk, err = NewKafkaER(cfg, 1, rdrEvents,
		rdrErr, new(engine.FilterS), rdrExit); err != nil {
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
	case err = <-rdrErr:
		t.Error(err)
	case ev := <-rdrEvents:
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
	rdrExit <- struct{}{}

	// Delete kafka topic

	partitions, err := conn.ReadPartitions("cgrates_cdrs")
	if err != nil {
		t.Fatal(err)
	}

	if len(partitions) != 1 || partitions[0].Topic != "cgrates_cdrs" {
		t.Fatal("expected topic named cgrates_cdrs to exist")
	}

	if err := conn.DeleteTopics("cgrates_cdrs"); err != nil {
		t.Fatal(err)
	}

	experr := `[5] Leader Not Available: the cluster is in the middle of a leadership election and there is currently no leader for this partition and hence it is unavailable for writes`
	_, err = conn.ReadPartitions("cgrates_cdrs")
	if err == nil || err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}
