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

package ees

import (
	"os/exec"
	"path"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/nats-io/nats.go"
)

func TestNatsEE(t *testing.T) {
	testCreateDirectory(t)
	var err error
	cmd := exec.Command("nats-server", "-js") // Start the nats-server.
	if err := cmd.Start(); err != nil {
		t.Fatal(err) // Only if nats-server is not installed.
	}
	time.Sleep(50 * time.Millisecond)
	defer cmd.Process.Kill()
	cfgPath := path.Join(*dataDir, "conf", "samples", "ees")
	cfg, err := config.NewCGRConfigFromPath(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	evExp, err := NewEventExporter(cfg, 5, new(engine.FilterS))
	if err != nil {
		t.Fatal(err)
	}

	nop, err := engine.GetNatsOpts(cfg.EEsCfg().Exporters[5].Opts, "natsTest", time.Second)
	if err != nil {
		t.Fatal(err)
	}

	nc, err := nats.Connect("nats://localhost:4222", nop...)
	if err != nil {
		t.Fatal(err)
	}
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
	}

	if _, err = js.AddStream(&nats.StreamConfig{
		Name:     "test",
		Subjects: []string{"processed_cdrs"},
	}); err != nil {
		t.Fatal(err)
	}

	if err = js.PurgeStream("test"); err != nil {
		t.Fatal(err)
	}

	ch := make(chan *nats.Msg, 3)
	_, err = js.QueueSubscribe("processed_cdrs", "test3", func(msg *nats.Msg) {
		ch <- msg
	}, nats.Durable("test4"))
	if err != nil {
		t.Fatal(err)
	}

	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		Event: map[string]interface{}{
			"Account":     "1001",
			"Destination": "1002",
		},
	}
	if err := evExp.ExportEvent(cgrEv); err != nil {
		t.Fatal(err)
	}
	testCleanDirectory(t)
	// fmt.Println(string((<-ch).Data))
}
