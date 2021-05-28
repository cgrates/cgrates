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
	"flag"
	"fmt"
	"reflect"
	"testing"
	"time"

	amqpv1 "github.com/Azure/go-amqp"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	itTestAMQPv1 = flag.Bool("amqpv1", false, "Run the test for AMQPv1Reader")
)

func TestAMQPERv1(t *testing.T) {
	if !*itTestAMQPv1 {
		t.SkipNow()
	}
	cfg, err := config.NewCGRConfigFromJSONStringWithDefaults(`{
"ers": {									// EventReaderService
	"enabled": true,						// starts the EventReader service: <true|false>
	"sessions_conns":["*localhost"],
	"readers": [
		{
			"id": "amqpv1",										// identifier of the EventReader profile
			"type": "*amqpv1_json_map",							// reader type <*file_csv>
			"run_delay":  "-1",									// sleep interval in seconds between consecutive runs, -1 to use automation via inotify or 0 to disable running all together
			"concurrent_requests": 1024,						// maximum simultaneous requests/files to process, 0 for unlimited
			"source_path": "amqps://RootManageSharedAccessKey:Je8l%2Bt9tyOgZbdA%2B5SmGIJEsEzhZ9VdIO7yRke5EYtM%3D@test0123456y.servicebus.windows.net",// read data from this path
			"opts": {
				"amqpQueueID": "cdrs3",
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

	if rdr, err = NewAMQPv1ER(cfg, 1, rdrEvents, make(chan *erEvent, 1),
		rdrErr, new(engine.FilterS), rdrExit); err != nil {
		t.Fatal(err)
	}
	amqpv1Rdr := rdr.(*AMQPv1ER)
	connection, err := amqpv1.Dial("amqps://RootManageSharedAccessKey:Je8l%2Bt9tyOgZbdA%2B5SmGIJEsEzhZ9VdIO7yRke5EYtM%3D@test0123456y.servicebus.windows.net")
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()

	channel, err := connection.NewSession()
	if err != nil {
		t.Fatal(err)
	}
	defer channel.Close(context.Background())

	randomCGRID := utils.UUIDSha1Prefix()
	sndr, err := channel.NewSender(amqpv1.LinkTargetAddress(amqpv1Rdr.queueID))
	if err != nil {
		t.Fatal(err)
	}
	if err = sndr.Send(context.Background(),
		amqpv1.NewMessage([]byte(fmt.Sprintf(`{"CGRID": "%s"}`, randomCGRID)))); err != nil {
		t.Fatal(err)
	}
	if err = rdr.Serve(); err != nil {
		t.Fatal(err)
	}
	select {
	case err = <-rdrErr:
		t.Error(err)
	case ev := <-rdrEvents:
		if ev.rdrCfg.ID != "amqpv1" {
			t.Errorf("Expected 'amqpv1' received `%s`", ev.rdrCfg.ID)
		}
		expected := &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     ev.cgrEvent.ID,
			Event: map[string]interface{}{
				"CGRID": randomCGRID,
			},
		}
		if !reflect.DeepEqual(ev.cgrEvent, expected) {
			t.Errorf("Expected %s ,received %s", utils.ToJSON(expected), utils.ToJSON(ev.cgrEvent))
		}
	case <-time.After(10 * time.Second):
		t.Fatal("Timeout")
	}

	close(rdrExit)
}

func TestAmqpv1NewAMQPv1ER(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfgIdx := 0
	expected := &AMQPv1ER{
		cgrCfg: cfg,
		cfgIdx: cfgIdx,
		poster: nil,
	}
	cfg.ERsCfg().Readers = []*config.EventReaderCfg{
		{
			ID:             utils.MetaDefault,
			Type:           utils.MetaNone,
			RunDelay:       0,
			ConcurrentReqs: -1,
			SourcePath:     "/var/spool/cgrates/ers/in",
			ProcessedPath:  "/var/spool/cgrates/ers/out",
			Filters:        []string{},
			Opts:           make(map[string]interface{}),
		},
	}

	result, err := NewAMQPv1ER(cfg, cfgIdx, nil, nil, nil, nil, nil)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	expected.poster = result.(*AMQPv1ER).poster
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestAmqpv1NewAMQPv1ER2(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfgIdx := 0
	expected := &AMQPv1ER{
		cgrCfg: cfg,
		cfgIdx: cfgIdx,
		poster: nil,
	}
	cfg.ERsCfg().Readers = []*config.EventReaderCfg{
		{
			ID:             utils.MetaDefault,
			Type:           utils.MetaNone,
			RunDelay:       0,
			ConcurrentReqs: 1,
			SourcePath:     "/var/spool/cgrates/ers/in",
			ProcessedPath:  "/var/spool/cgrates/ers/out",
			Filters:        []string{},
			Opts:           make(map[string]interface{}),
		},
	}

	result, err := NewAMQPv1ER(cfg, cfgIdx, nil, nil, nil, nil, nil)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	expected.poster = result.(*AMQPv1ER).poster
	expected.cap = result.(*AMQPv1ER).cap
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}
