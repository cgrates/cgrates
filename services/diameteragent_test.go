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
