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
	"runtime"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/nats-io/nats.go"
)

func TestNatsER(t *testing.T) {
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
			// "processed_path": "/var/spool/cgrates/ers/out",	
			"tenant": "cgrates.org",							
			"filters": [],										
			"flags": [],										
			"fields":[									
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
	// js, err := nc.JetStream()
	// if err != nil {
	// t.Fatal(err)
	// }
	go rdr.Serve()
	runtime.Gosched()
	time.Sleep(time.Second)
	randomCGRID := utils.UUIDSha1Prefix()
	if err = nc.Publish(utils.DefaultQueueID, []byte(fmt.Sprintf(`{"CGRID": "%s"}`, randomCGRID))); err != nil {
		t.Fatal(err)
	}
	// if _, err = js.Publish(utils.DefaultQueueID, []byte(fmt.Sprintf(`{"CGRID": "%s"}`, randomCGRID))); err != nil {
	// t.Fatal(err)
	// }

	nc.FlushTimeout(time.Second)
	nc.Flush()
	nc.Drain()

	select {
	case err = <-rdrErr:
		t.Error(err)
	case ev := <-rdrEvents:
		if ev.rdrCfg.ID != "nats" {
			t.Errorf("Expected 'kakfa' received `%s`", ev.rdrCfg.ID)
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
			t.Errorf("Expected %s ,received %s", utils.ToJSON(expected), utils.ToJSON(ev.cgrEvent))
		}
	case <-time.After(10 * time.Second):
		t.Fatal("Timeout")
	}
	close(rdrExit)
}
