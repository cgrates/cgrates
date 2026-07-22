//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
	"conns": {
		"*sessions": [{"connIDs": ["*localhost"]}]
	},
	"readers": [
		{
			"id": "kafka",										
			"type": "*kafkaJSONMap",							
			"runDelay":  "-1",									
			"concurrentRequests": 1024,						
			"sourcePath": "localhost:9092",
			"tenant": "cgrates.org",
			"filters": [],
			"flags": [],
			"fields":[
				{"tag": "OriginID", "type": "*composed", "value": "~*req.OriginID", "path": "*cgreq.OriginID"},
				{"tag": "readerId", "type": "*variable", "value": "~*vars.*readerID", "path": "*cgreq.ReaderID"},
			]
		}
	]
}
}`)
	locker := engine.NewGuardianLocker(cfg)
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
		rdrErr, engine.NewCacheS(cfg, nil, nil, nil, locker), new(engine.FilterS), rdrExit); err != nil {
		t.Fatal(err)
	}
	rdr.Serve()

	randomOriginID := utils.UUIDSha1Prefix()
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
			Key:   []byte(randomOriginID), // for the moment we do not process the key
			Value: []byte(fmt.Sprintf(`{"OriginID": "%s"}`, randomOriginID)),
		})
		if err := res.FirstErr(); err != nil {
			t.Error("failed to write messages:", err)
		}
	}(randomOriginID)

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
			Event: map[string]any{
				"OriginID": randomOriginID,
				"ReaderID": cfg.ERsCfg().Readers[1].ID,
			},
			APIOpts: map[string]any{},
		}
		if !reflect.DeepEqual(ev.cgrEvent, expected) {
			t.Errorf("Expected %s ,received %s", utils.ToJSON(expected), utils.ToJSON(ev.cgrEvent))
		}
	case <-time.After(10 * time.Second):
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
