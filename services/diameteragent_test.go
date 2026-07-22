// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package services

import (
	"sync"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestDiameterAgentServiceName(t *testing.T) {
	da := &DiameterAgent{}
	want := utils.DiameterAgent
	got := da.ServiceName()

	if got != want {
		t.Errorf("ServiceName() = %v; want %v", got, want)
	}
}

func TestNewDiameterAgent(t *testing.T) {
	cfg := &config.CGRConfig{}
	filterSChan := make(chan *engine.FilterS)
	shdChan := &utils.SyncedChan{}
	connMgr := &engine.ConnManager{}
	srvDep := make(map[string]*sync.WaitGroup)

	service := NewDiameterAgent(cfg, filterSChan, shdChan, connMgr, nil, srvDep)

	da, ok := service.(*DiameterAgent)
	if !ok {
		t.Fatalf("NewDiameterAgent() returned wrong type: got %T, want *DiameterAgent", service)
	}

	if da.cfg != cfg {
		t.Errorf("cfg = %v; want %v", da.cfg, cfg)
	}
	if da.filterSChan != filterSChan {
		t.Errorf("filterSChan = %v; want %v", da.filterSChan, filterSChan)
	}
	if da.shdChan != shdChan {
		t.Errorf("shdChan = %v; want %v", da.shdChan, shdChan)
	}
	if da.connMgr != connMgr {
		t.Errorf("connMgr = %v; want %v", da.connMgr, connMgr)
	}

}
