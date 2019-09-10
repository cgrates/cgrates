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
	"reflect"
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
	cfg, err := config.NewCGRConfigFromJsonStringWithDefaults(`{
"ers": {									// EventReaderService
	"enabled": true,						// starts the EventReader service: <true|false>
	"readers": [
		{
			"id": "kafka",										// identifier of the EventReader profile
			"type": "*kafka_json_map",							// reader type <*file_csv>
			"run_delay": -1,									// sleep interval in seconds between consecutive runs, -1 to use automation via inotify or 0 to disable running all together
			"concurrent_requests": 1024,						// maximum simultaneous requests/files to process, 0 for unlimited
			"source_path": "localhost:9092",					// read data from this path
			// "processed_path": "/var/spool/cgrates/cdrc/out",	// move processed data here
			"tenant": "cgrates.org",							// tenant used by import
			"filters": [],										// limit parsing based on the filters
			"flags": [],										// flags to influence the event processing
			// "header_fields": [],								// template of the import header fields
			"content_fields":[									// import content_fields template, tag will match internally CDR field, in case of .csv value will be represented by index of the field value
				{"tag": "CGRID", "type": "*composed", "value": "~*req.CGRID", "field_id": "CGRID"},
			],
			// "trailer_fields": [],								// template of the import trailer fields
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
	w := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   defaultTopic,
	})

	w.WriteMessages(context.Background(),
		kafka.Message{
			Key:   []byte("TestKey"), // for the momment we do not proccess the key
			Value: []byte(`{"CGRID": "RandomCGRID"}`),
		},
	)

	w.Close()
	// tStart := time.Now()
	select {
	case err = <-rdrErr:
		t.Error(err)
	case ev := <-rdrEvents:
		// fmt.Printf("It took %s to proccess the message.\n", time.Now().Sub(tStart))
		if ev.rdrCfg.ID != "kafka" {
			t.Errorf("Expected 'kakfa' received `%s`", ev.rdrCfg.ID)
		}
		expected := &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     ev.cgrEvent.ID,
			Time:   ev.cgrEvent.Time,
			Event: map[string]interface{}{
				"CGRID": "RandomCGRID",
			},
		}
		if !reflect.DeepEqual(ev.cgrEvent, expected) {
			t.Errorf("Expected %s ,received %s", utils.ToJSON(expected), utils.ToJSON(ev.cgrEvent))
		}
	case <-time.After(10 * time.Second):
		t.Errorf("Timeout")
	}
	rdrExit <- struct{}{}
}
