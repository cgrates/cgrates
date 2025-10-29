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

package ees

import (
	"os"
	"os/exec"
	"path"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

func TestNatsEEJetStream(t *testing.T) {
	exec.Command("pkill", "nats-server")
	cmd := exec.Command("nats-server", "-js")
	if err := cmd.Start(); err != nil {
		t.Fatal(err) // most probably not installed
	}
	time.Sleep(50 * time.Millisecond)
	defer cmd.Process.Kill()

	testCreateDirectory(t)
	cgrCfg, err := config.NewCGRConfigFromPath(path.Join(*utils.DataDir, "conf", "samples", "ees"))
	if err != nil {
		t.Fatal(err)
	}
	var cfg *config.EventExporterCfg
	for _, cfg = range cgrCfg.EEsCfg().Exporters {
		if cfg.ID == "NatsJsonMapExporter" {
			break
		}
	}
	evExp, err := NewEventExporter(cfg, cgrCfg, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	nop, err := GetNatsOpts(cfg.Opts.NATS, "natsTest", time.Second)
	if err != nil {
		t.Fatal(err)
	}

	nc, err := nats.Connect(nats.DefaultURL, nop...)
	if err != nil {
		t.Fatal(err)
	}
	defer nc.Drain()

	js, err := jetstream.New(nc)
	if err != nil {
		t.Fatal(err)
	}

	_, err = js.CreateStream(context.Background(), jetstream.StreamConfig{
		Name:     "test2",
		Subjects: []string{"processed_cdrs"},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer js.DeleteStream(context.Background(), "test2")

	ch := make(chan string, 3)
	var cons jetstream.Consumer
	cons, err = js.CreateOrUpdateConsumer(context.Background(), "test2", jetstream.ConsumerConfig{
		Durable:       "test4",
		FilterSubject: "processed_cdrs",
		AckPolicy:     jetstream.AckAllPolicy,
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = cons.Consume(func(msg jetstream.Msg) {
		ch <- string(msg.Data())
	})
	if err != nil {
		t.Fatal(err)
	}

	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		Event: map[string]any{
			"Account":     "1001",
			"Destination": "1002",
		},
	}
	if err := exportEventWithExporter(evExp, cgrEv, true, cgrCfg, new(engine.FilterS)); err != nil {
		t.Fatal(err)
	}
	testCleanDirectory(t)
	expected := `{"Account":"1001","Destination":"1002"}`
	// fmt.Println((<-ch).Data)
	select {
	case data := <-ch:
		if expected != data {
			t.Fatalf("Expected %v \n but received \n %v", expected, data)
		}
	case <-time.After(50 * time.Millisecond):
		t.Fatal("Time limit exceeded")
	}
}

func TestNatsEE(t *testing.T) {
	exec.Command("pkill", "nats-server")
	cmd := exec.Command("nats-server")
	if err := cmd.Start(); err != nil {
		t.Fatal(err) // most probably not installed
	}
	time.Sleep(50 * time.Millisecond)
	defer cmd.Process.Kill()

	testCreateDirectory(t)
	cgrCfg, err := config.NewCGRConfigFromPath(path.Join(*utils.DataDir, "conf", "samples", "ees"))
	if err != nil {
		t.Fatal(err)
	}
	var cfg *config.EventExporterCfg
	for _, cfg = range cgrCfg.EEsCfg().Exporters {
		if cfg.ID == "NatsJsonMapExporter2" {
			break
		}
	}
	evExp, err := NewEventExporter(cfg, cgrCfg, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	nop, err := GetNatsOpts(cfg.Opts.NATS, "natsTest", time.Second)
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
		Event: map[string]any{
			"Account":     "1001",
			"Destination": "1002",
		},
	}
	if err := exportEventWithExporter(evExp, cgrEv, true, cgrCfg, new(engine.FilterS)); err != nil {
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

func TestGetNatsOptsSeedFile(t *testing.T) {
	if _, err := os.Create("/tmp/nkey.txt"); err != nil {
		t.Error(err)
	}
	defer os.Remove("/tmp/nkey.txt")
	nkey := "SUACSSL3UAHUDXKFSNVUZRF5UHPMWZ6BFDTJ7M6USDXIEDNPPQYYYCU3VY"
	os.WriteFile("/tmp/nkey.txt", []byte(nkey), 0777)

	opts := &config.NATSOpts{
		SeedFile: utils.StringPointer("/tmp/nkey.txt"),
	}

	nodeID := "node_id1"
	connTimeout := 2 * time.Second

	_, err := GetNatsOpts(opts, nodeID, connTimeout)
	if err != nil {
		t.Error(err)
	}

	//test error
	os.WriteFile("/tmp/nkey.txt", []byte(""), 0777)
	_, err = GetNatsOpts(opts, nodeID, connTimeout)
	if err == nil || err.Error() != "nkeys: no nkey seed found" {
		t.Errorf("expected \"%s\" but received \"%s\"", "nkeys: no nkey seed found", err.Error())
	}
}
