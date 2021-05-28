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
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/streadway/amqp"
)

func TestAMQPER(t *testing.T) {
	cfg, err := config.NewCGRConfigFromJSONStringWithDefaults(`{
"ers": {									// EventReaderService
	"enabled": true,						// starts the EventReader service: <true|false>
	"sessions_conns":["*localhost"],
	"readers": [
		{
			"id": "amqp",										// identifier of the EventReader profile
			"type": "*amqp_json_map",							// reader type <*file_csv>
			"run_delay":  "-1",									// sleep interval in seconds between consecutive runs, -1 to use automation via inotify or 0 to disable running all together
			"concurrent_requests": 1024,						// maximum simultaneous requests/files to process, 0 for unlimited
			"source_path": "amqp://guest:guest@localhost:5672/",// read data from this path
			"opts": {
				"amqpQueueID": "cdrs3",
				"amqpConsumerTag": "test-key",
				"amqpExchange": "test-exchange",
				"amqpExchangeType": "direct",
				"amqpRoutingKey": "test-key",
			},
			"processed_path": "",								// move processed data here
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
	if err := cfg.CheckConfigSanity(); err != nil {
		t.Fatal(err)
	}
	utils.Logger, _ = utils.Newlogger(utils.MetaSysLog, cfg.GeneralCfg().NodeID)
	utils.Logger.SetLogLevel(7)

	rdrEvents = make(chan *erEvent, 1)
	rdrErr = make(chan error, 1)
	rdrExit = make(chan struct{}, 1)

	if rdr, err = NewAMQPER(cfg, 1, rdrEvents, make(chan *erEvent, 1),
		rdrErr, new(engine.FilterS), rdrExit); err != nil {
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
	randomCGRID := utils.UUIDSha1Prefix()
	if err = channel.Publish(
		"test-exchange", // publish to an exchange
		"test-key",      // routing to 0 or more queues
		false,           // mandatory
		false,           // immediate
		amqp.Publishing{
			ContentType:  utils.ContentJSON,
			Body:         []byte(fmt.Sprintf(`{"CGRID": "%s"}`, randomCGRID)),
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
			Event: map[string]interface{}{
				"CGRID": randomCGRID,
			},
			APIOpts: map[string]interface{}{},
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

func TestAMQPERServeError(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfgIdx := 0
	expected := "AMQP scheme must be either 'amqp://' or 'amqps://'"
	rdr, err := NewAMQPER(cfg, cfgIdx, nil, nil, nil, nil, nil)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	err2 := rdr.Serve()
	if err2 == nil || err2.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err2)
	}
}
