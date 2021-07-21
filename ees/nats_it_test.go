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

func TestNatsEEJetStream(t *testing.T) {
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
	var idx int
	for idx = range cfg.EEsCfg().Exporters {
		if cfg.EEsCfg().Exporters[idx].ID == "NatsJsonMapExporter" {
			break
		}
	}
	evExp, err := NewEventExporter(cfg, idx, new(engine.FilterS))
	if err != nil {
		t.Fatal(err)
	}

	nop, err := engine.GetNatsOpts(cfg.EEsCfg().Exporters[idx].Opts, "natsTest", time.Second)
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
		if name == "test2" {
			if err = js.DeleteStream("test2"); err != nil {
				t.Fatal(err)
			}
			break
		}
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
	expected := `{"Account":"1001","Destination":"1002"}`
	// fmt.Println((<-ch).Data)
	select {
	case data := <-ch:
		if expected != string(data.Data) {
			t.Fatalf("Expected %v \n but received \n %v", expected, string(data.Data))
		}
	case <-time.After(50 * time.Millisecond):
		t.Fatal("Time limit exceeded")
	}
}

func TestNatsEE(t *testing.T) {
	testCreateDirectory(t)
	exec.Command("pkill", "nats-server")

	cmd := exec.Command("nats-server")
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	time.Sleep(50 * time.Millisecond)
	defer cmd.Process.Kill()

	cfgPath := path.Join(*dataDir, "conf", "samples", "ees")
	cfg, err := config.NewCGRConfigFromPath(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	var idx int
	for idx = range cfg.EEsCfg().Exporters {
		if cfg.EEsCfg().Exporters[idx].ID == "NatsJsonMapExporter2" {
			break
		}
	}
	evExp, err := NewEventExporter(cfg, idx, new(engine.FilterS))
	if err != nil {
		t.Fatal(err)
	}

	nop, err := engine.GetNatsOpts(cfg.EEsCfg().Exporters[idx].Opts, "natsTest", time.Second)
	if err != nil {
		t.Fatal(err)
	}
	nc, err := nats.Connect("nats://localhost:4222", nop...)
	if err != nil {
		t.Fatal(err)
	}

	ch := make(chan *nats.Msg, 3)
	_, err = nc.ChanQueueSubscribe("processed_cdrs", "test3", ch)
	if err != nil {
		t.Fatal(err)
	}

	defer nc.Drain()

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
	expected := `{"Account":"1001","Destination":"1002"}`
	// fmt.Println((<-ch).Data)
	select {
	case data := <-ch:
		if expected != string(data.Data) {
			t.Fatalf("Expected %v \n but received \n %v", expected, string(data.Data))
		}
	case <-time.After(50 * time.Millisecond):
		t.Fatal("Time limit exceeded")
	}
}
