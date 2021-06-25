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
	"os/exec"
	"reflect"
	"runtime"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/nats-io/nats.go"
)

func TestNatsERJetStream(t *testing.T) {
	// start the nats-server
	exec.Command("pkill", "nats-server")

	cmd := exec.Command("nats-server", "-js")
	if err := cmd.Start(); err != nil {
		t.Fatal(err) // most probably not installed
	}
	time.Sleep(50 * time.Millisecond)
	defer cmd.Process.Kill()
	//

	cfg, err := config.NewCGRConfigFromJSONStringWithDefaults(`{
"ers": {									// EventReaderService
	"enabled": true,						// starts the EventReader service: <true|false>
	"sessions_conns":["*localhost"],
	"readers": [
		{
			"id": "nats",										
			"type": "*nats_json_map",							
			"run_delay":  "-1",									
			"concurrent_requests": 1024,						
			"source_path": "nats://localhost:4222",				
			"processed_path": "",	
			"tenant": "cgrates.org",							
			"filters": [],										
			"flags": [],										
			"fields":[									
				{"tag": "CGRID", "type": "*composed", "value": "~*req.CGRID", "path": "*cgreq.CGRID"},
			],
			"opts": {
				"natsJetStream": true,
				"natsSubjectProcessed": "processed_cdrs",
			}
		},
	],
},
}`)
	utils.Logger.SetLogLevel(7)
	if err != nil {
		t.Fatal(err)
	}
	if err := cfg.CheckConfigSanity(); err != nil {
		t.Fatal(err)
	}
	rdrEvents = make(chan *erEvent, 1)
	rdrErr = make(chan error, 1)
	rdrExit = make(chan struct{}, 1)

	if rdr, err = NewNatsER(cfg, 1, rdrEvents, make(chan *erEvent, 1),
		rdrErr, new(engine.FilterS), rdrExit); err != nil {
		t.Fatal(err)
	}
	nc, err := nats.Connect(rdr.Config().SourcePath, nats.Timeout(time.Second),
		nats.DrainTimeout(time.Second))
	if err != nil {
		t.Fatal(err)
	}
	defer nc.Drain()

	js, err := nc.JetStream()
	if err != nil {
		t.Fatal(err)
	}
	for name := range js.StreamNames() {
		if name == "test" {
			if err = js.DeleteStream("test"); err != nil {
				t.Fatal(err)
			}
			break
		}
		if name == "test2" {
			if err = js.DeleteStream("test2"); err != nil {
				t.Fatal(err)
			}
			break
		}
	}
	if _, err = js.AddStream(&nats.StreamConfig{
		Name:     "test",
		Subjects: []string{utils.DefaultQueueID},
	}); err != nil {
		t.Fatal(err)
	}

	if err = js.PurgeStream("test"); err != nil {
		t.Fatal(err)
	}

	if _, err = js.AddStream(&nats.StreamConfig{
		Name:     "test2",
		Subjects: []string{"processed_cdrs"},
	}); err != nil {
		t.Fatal(err)
	}

	if err = js.PurgeStream("test2"); err != nil {
		t.Fatal(err)
	}
	ch := make(chan *nats.Msg, 3)
	_, err = js.QueueSubscribe("processed_cdrs", "test3", func(msg *nats.Msg) {
		ch <- msg
	}, nats.Durable("test4"))
	if err != nil {
		t.Fatal(err)
	}

	go rdr.Serve()
	runtime.Gosched()
	time.Sleep(10 * time.Nanosecond)

	for i := 0; i < 3; i++ {
		randomCGRID := utils.UUIDSha1Prefix()
		expData := fmt.Sprintf(`{"CGRID": "%s"}`, randomCGRID)
		if _, err = js.Publish(utils.DefaultQueueID, []byte(expData)); err != nil {
			t.Fatal(err)
		}

		nc.FlushTimeout(time.Second)
		nc.Flush()

		select {
		case err = <-rdrErr:
			t.Fatal(err)
		case ev := <-rdrEvents:
			if ev.rdrCfg.ID != "nats" {
				t.Fatalf("Expected 'nats' received `%s`", ev.rdrCfg.ID)
			}
			expected := &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     ev.cgrEvent.ID,
				Time:   ev.cgrEvent.Time,
				Event: map[string]interface{}{
					"CGRID": randomCGRID,
				},
				APIOpts: map[string]interface{}{},
			}
			if !reflect.DeepEqual(ev.cgrEvent, expected) {
				t.Fatalf("Expected %s ,received %s", utils.ToJSON(expected), utils.ToJSON(ev.cgrEvent))
			}
			select {
			case msg := <-ch:
				if expData != string(msg.Data) {
					t.Errorf("Expected %q ,received %q", expData, string(msg.Data))
				}
			case <-time.After(10 * time.Second):
				t.Fatal("Timeout")
			}
		case <-time.After(10 * time.Second):
			t.Fatal("Timeout")
		}
	}
	close(rdrExit)
}

func TestNatsER(t *testing.T) {
	// start the nats-server
	exec.Command("pkill", "nats-server")

	cmd := exec.Command("nats-server")
	if err := cmd.Start(); err != nil {
		t.Fatal(err) // most probably not installed
	}
	time.Sleep(10 * time.Millisecond)
	defer cmd.Process.Kill()
	//

	cfg, err := config.NewCGRConfigFromJSONStringWithDefaults(`{
"ers": {									// EventReaderService
	"enabled": true,						// starts the EventReader service: <true|false>
	"sessions_conns":["*localhost"],
	"readers": [
		{
			"id": "nats",										
			"type": "*nats_json_map",							
			"run_delay":  "-1",									
			"concurrent_requests": 1024,						
			"source_path": "nats://localhost:4222",				
			"processed_path": "",	
			"tenant": "cgrates.org",							
			"filters": [],										
			"flags": [],										
			"fields":[									
				{"tag": "CGRID", "type": "*composed", "value": "~*req.CGRID", "path": "*cgreq.CGRID"},
			],
			"opts": {
				"natsSubjectProcessed": "processed_cdrs",
			}
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
	rdrEvents = make(chan *erEvent, 1)
	rdrErr = make(chan error, 1)
	rdrExit = make(chan struct{}, 1)

	if rdr, err = NewNatsER(cfg, 1, rdrEvents, make(chan *erEvent, 1),
		rdrErr, new(engine.FilterS), rdrExit); err != nil {
		t.Fatal(err)
	}
	nc, err := nats.Connect(rdr.Config().SourcePath, nats.Timeout(time.Second),
		nats.DrainTimeout(time.Second))
	if err != nil {
		t.Fatal(err)
	}
	ch := make(chan *nats.Msg, 3)
	_, err = nc.ChanQueueSubscribe("processed_cdrs", "test3", ch)
	if err != nil {
		t.Fatal(err)
	}

	defer nc.Drain()
	go rdr.Serve()
	runtime.Gosched()
	time.Sleep(100 * time.Millisecond)
	for i := 0; i < 3; i++ {
		randomCGRID := utils.UUIDSha1Prefix()
		expData := fmt.Sprintf(`{"CGRID": "%s"}`, randomCGRID)
		if err = nc.Publish(utils.DefaultQueueID, []byte(expData)); err != nil {
			t.Fatal(err)
		}

		nc.FlushTimeout(time.Second)
		nc.Flush()

		select {
		case err = <-rdrErr:
			t.Fatal(err)
		case ev := <-rdrEvents:
			if ev.rdrCfg.ID != "nats" {
				t.Fatalf("Expected 'nats' received `%s`", ev.rdrCfg.ID)
			}
			expected := &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     ev.cgrEvent.ID,
				Time:   ev.cgrEvent.Time,
				Event: map[string]interface{}{
					"CGRID": randomCGRID,
				},
				APIOpts: map[string]interface{}{},
			}
			if !reflect.DeepEqual(ev.cgrEvent, expected) {
				t.Fatalf("Expected %s ,received %s", utils.ToJSON(expected), utils.ToJSON(ev.cgrEvent))
			}
			select {
			case msg := <-ch:
				if expData != string(msg.Data) {
					t.Errorf("Expected %q ,received %q", expData, string(msg.Data))
				}
			case <-time.After(10 * time.Second):
				t.Fatal("Timeout")
			}
		case <-time.After(10 * time.Second):
			t.Fatal("Timeout")
		}
	}
	close(rdrExit)
}
