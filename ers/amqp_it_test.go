//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package ers

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	amqp "github.com/rabbitmq/amqp091-go"
)

func TestAMQPER(t *testing.T) {
	cfg, err := config.NewCGRConfigFromJSONStringWithDefaults(`{
"ers": {									// EventReaderService
	"enabled": true,						// starts the EventReader service: <true|false>
	"conns": {
		"*sessions": [{"connIDs": ["*localhost"]}]
	},
	"readers": [
		{
			"id": "amqp",										// identifier of the EventReader profile
			"type": "*amqpJSONMap",							// reader type <*fileCSV>
			"runDelay":  "-1",									// sleep interval in seconds between consecutive runs, -1 to use automation via inotify or 0 to disable running all together
			"concurrentRequests": 1024,						// maximum simultaneous requests/files to process, 0 for unlimited
			"sourcePath": "amqp://guest:guest@localhost:5672/",// read data from this path
			"opts": {
				"amqpQueueID": "cdrs3",
				"amqpConsumerTag": "test-key",
				"amqpExchange": "test-exchange",
				"amqpExchangeType": "direct",
				"amqpRoutingKey": "test-key",
			},
			"processedPath": "",								// move processed data here
			"tenant": "cgrates.org",							// tenant used by import
			"filters": [],										// limit parsing based on the filters
			"flags": [],										// flags to influence the event processing
			"fields":[									// import fields template, tag will match internally CDR field, in case of .csv value will be represented by index of the field value
				{"tag": "OriginID", "type": "*composed", "value": "~*req.OriginID", "path": "*cgreq.OriginID"},
				{"tag": "readerId", "type": "*variable", "value": "~*vars.*readerID", "path": "*cgreq.ReaderID"},
			],
		},
	],
},
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

	if rdr, err = NewAMQPER(cfg, 1, rdrEvents, make(chan *erEvent, 1),
		rdrErr, engine.NewCacheS(cfg, nil, nil, nil, locker), new(engine.FilterS), rdrExit); err != nil {
		t.Fatal(err)
	}
	connection, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()

	channel, err := connection.Channel()
	if err != nil {
		t.Fatal(err)
	}

	rdr.Serve()
	randomOriginID := utils.UUIDSha1Prefix()
	if err = channel.PublishWithContext(
		context.Background(),
		"test-exchange", // publish to an exchange
		"test-key",      // routing to 0 or more queues
		false,           // mandatory
		false,           // immediate
		amqp.Publishing{
			ContentType:  utils.ContentJSON,
			Body:         []byte(fmt.Sprintf(`{"OriginID": "%s"}`, randomOriginID)),
			DeliveryMode: amqp.Persistent, // 1=non-persistent, 2=persistent
		},
	); err != nil {
		t.Fatal(err)
	}
	select {
	case err = <-rdrErr:
		t.Error(err)
	case ev := <-rdrEvents:
		if ev.rdrCfg.ID != "amqp" {
			t.Errorf("Expected 'amqp' received `%s`", ev.rdrCfg.ID)
		}
		expected := &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     ev.cgrEvent.ID,
			Event: map[string]any{
				"OriginID": randomOriginID,
				"ReaderID": "amqp",
			},
			APIOpts: map[string]any{},
		}
		if !reflect.DeepEqual(ev.cgrEvent, expected) {
			t.Errorf("Expected %s ,received %s", utils.ToJSON(expected), utils.ToJSON(ev.cgrEvent))
		}
	case <-time.After(10 * time.Second):
		t.Fatal("Timeout")
	}

	if _, err := channel.QueueDelete("cdrs3", false, false, false); err != nil {
		t.Fatal(err)
	}
	close(rdrExit)
}
