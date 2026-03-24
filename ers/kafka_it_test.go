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
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kgo"
)

var (
	rdrEvents chan *erEvent
	rdrErr    chan error
	rdrExit   chan struct{}
	rdr       EventReader
)

func TestKafkaER(t *testing.T) {

	// Create kafka topic
	cl, err := kgo.NewClient(kgo.SeedBrokers("localhost:9092"))
	if err != nil {
		t.Fatal(err)
	}
	adm := kadm.NewClient(cl)
	_, err = adm.CreateTopics(context.Background(), 1, 1, nil, utils.KafkaDefaultTopic)
	if err != nil {
		t.Fatal(err)
	}
	cl.Close()

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
				{"tag": "readerId", "type": "*variable", "value": "~*vars.*readerID", "path": "*cgreq.ReaderID"}
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
		produceCl, err := kgo.NewClient(
			kgo.SeedBrokers("localhost:9092"),
			kgo.DefaultProduceTopic(utils.KafkaDefaultTopic),
		)
		if err != nil {
			t.Error("failed to create producer:", err)
			return
		}
		defer produceCl.Close()
		res := produceCl.ProduceSync(context.Background(), &kgo.Record{
			Key:   []byte(randomCGRID),
			Value: []byte(fmt.Sprintf(`{"CGRID": "%s"}`, randomCGRID)),
		})
		if err := res.FirstErr(); err != nil {
			t.Error("failed to write messages:", err)
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
	cl2, err := kgo.NewClient(kgo.SeedBrokers("localhost:9092"))
	if err != nil {
		t.Fatal(err)
	}
	defer cl2.Close()
	adm2 := kadm.NewClient(cl2)

	topics, err := adm2.ListTopics(context.Background(), utils.KafkaDefaultTopic)
	if err != nil {
		t.Fatal(err)
	}
	if !topics.Has(utils.KafkaDefaultTopic) {
		t.Fatal("expected topic named cgrates to exist")
	}

	if _, err := adm2.DeleteTopics(context.Background(), utils.KafkaDefaultTopic); err != nil {
		t.Fatal(err)
	}

	topics, err = adm2.ListTopics(context.Background(), utils.KafkaDefaultTopic)
	if err != nil {
		t.Fatal(err)
	}
	if topics.Has(utils.KafkaDefaultTopic) {
		t.Error("expected topic to be deleted")
	}
}
