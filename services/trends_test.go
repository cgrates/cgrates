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

	"github.com/cgrates/birpc"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestNewTrendService(t *testing.T) {

	cfg := &config.CGRConfig{}
	dm := &DataDBService{}
	cacheS := &engine.CacheS{}
	filterSChan := make(chan *engine.FilterS)
	server := &cores.Server{}
	internalStatSChan := make(chan birpc.ClientConnector)
	connMgr := &engine.ConnManager{}
	anz := &AnalyzerService{}
	srvDep := make(map[string]*sync.WaitGroup)
	service := NewTrendService(cfg, dm, cacheS, filterSChan, server, internalStatSChan, connMgr, anz, srvDep)
	trendService, ok := service.(*TrendService)

	if !ok {
		t.Fatalf("Expected *TrendService, got %T", service)
	}

	if trendService.cfg != cfg {
		t.Errorf("Expected cfg field to be %v, got %v", cfg, trendService.cfg)
	}

	if trendService.dm != dm {
		t.Errorf("Expected dm field to be %v, got %v", dm, trendService.dm)
	}

	if trendService.cacheS != cacheS {
		t.Errorf("Expected cacheS field to be %v, got %v", cacheS, trendService.cacheS)
	}

	if trendService.filterSChan != filterSChan {
		t.Errorf("Expected filterSChan field to be %v, got %v", filterSChan, trendService.filterSChan)
	}

	if trendService.server != server {
		t.Errorf("Expected server field to be %v, got %v", server, trendService.server)
	}

	if trendService.connChan != internalStatSChan {
		t.Errorf("Expected internalStatSChan field to be %v, got %v", internalStatSChan, trendService.connChan)
	}

	if trendService.connMgr != connMgr {
		t.Errorf("Expected connMgr field to be %v, got %v", connMgr, trendService.connMgr)
	}

	if trendService.anz != anz {
		t.Errorf("Expected anz field to be %v, got %v", anz, trendService.anz)
	}

}

func TestTrendServiceServiceName(t *testing.T) {
	trendService := &TrendService{}
	name := trendService.ServiceName()
	if name != utils.TrendS {
		t.Errorf("Expected ServiceName to return %s, got %s", utils.TrendS, name)
	}
}

func TestTrendServiceIsRunning(t *testing.T) {
	trendService := &TrendService{}
	result := trendService.IsRunning()
	if result != false {
		t.Errorf("Expected IsRunning to return false, got %v", result)
	}
}

func TestTrendServiceStartAlreadyRunning(t *testing.T) {

	trendService := &TrendService{}

	trendService.trs = &engine.TrendS{}

	err := trendService.Start()

	if err != utils.ErrServiceAlreadyRunning {
		t.Errorf("Expected error '%v', got '%v'", utils.ErrServiceAlreadyRunning, err)
	}
}
