//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package cores

import (
	"net/rpc/jsonrpc"
	"path"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestCAPsStatusAllocated(t *testing.T) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cfgPath := path.Join("/usr/share/cgrates", "conf", "samples", "caps_queue")
	cfg, err := config.NewCGRConfigFromPath(ctx, cfgPath)
	if err != nil {
		t.Error(err)
	}

	if _, err := engine.StopStartEngine(cfgPath, 100); err != nil {
		t.Fatal(err)
	}
	defer engine.KillEngine(100)

	client, err := jsonrpc.Dial(utils.TCP, cfg.ListenCfg().RPCJSONListen)
	if err != nil {
		t.Fatalf("could not establish connection to engine: %v", err)
	}

	cfgStr := `{"cores":{"caps":2,"capsStatsInterval":"0","capsStrategy":"*queue","conns":{},"shutdownTimeout":"1s"}}`

	var rpl string
	if err := client.Call(utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.CoreSJSON},
	}, &rpl); err != nil {
		t.Error(err)
	} else if cfgStr != rpl {
		t.Errorf("Expected %q , \nreceived: %q", cfgStr, rpl)
	}

	var reply map[string]any
	if err := client.Call(utils.CoreSv1Status, utils.TenantWithAPIOpts{},
		&reply); err != nil {
		t.Fatal(err)
	} else if reply[utils.NodeID] != "ConcurrentQueueEngine" {
		t.Errorf("Expected %+v , received: %+v ", "ConcurrentQueueEngine", reply)
	} else if _, has := reply[utils.MetricCapsAllocated]; !has {
		t.Errorf("Expected reply to contain CAPSAllocated , received <%+v>", reply)
	}

}
func TestCAPsStatusPeak(t *testing.T) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cfgPath := path.Join("/usr/share/cgrates", "conf", "samples", "caps_peak")
	cfg, err := config.NewCGRConfigFromPath(ctx, cfgPath)
	if err != nil {
		t.Error(err)
	}
	if _, err := engine.StopStartEngine(cfgPath, 100); err != nil {
		t.Fatal(err)
	}
	defer engine.KillEngine(100)

	client, err := jsonrpc.Dial(utils.TCP, cfg.ListenCfg().RPCJSONListen)
	if err != nil {
		t.Fatalf("could not establish connection to engine: %v", err)
	}
	cfgStr := `{"cores":{"caps":2,"capsStatsInterval":"100ms","capsStrategy":"*queue","conns":{},"shutdownTimeout":"1s"}}`

	var rpl string
	if err := client.Call(utils.ConfigSv1GetConfigAsJSON, &config.SectionWithAPIOpts{
		Tenant:   "cgrates.org",
		Sections: []string{config.CoreSJSON},
	}, &rpl); err != nil {
		t.Error(err)
	} else if cfgStr != rpl {
		t.Errorf("Expected %q , \nreceived: %q", cfgStr, rpl)
	}

	var reply map[string]any
	if err := client.Call(utils.CoreSv1Status, utils.TenantWithAPIOpts{},
		&reply); err != nil {
		t.Fatal(err)
	} else if reply[utils.NodeID] != "CAPSPeakEngine" {
		t.Errorf("Expected %+v , received: %+v ", "CAPSPeakEngine", reply)
	} else if _, has := reply[utils.MetricCapsPeak]; !has {
		t.Errorf("Expected reply to contain CAPSPeak , received <%+v>", reply)
	}

}
