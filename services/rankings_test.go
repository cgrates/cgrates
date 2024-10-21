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
	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestRankingServiceServiceName(t *testing.T) {
	expectedServiceName := utils.RankingS
	rankingService := &RankingService{}
	result := rankingService.ServiceName()
	if result != expectedServiceName {
		t.Errorf("Expected ServiceName to return %v, got %v", expectedServiceName, result)
	}
}

func TestRankingServiceReload(t *testing.T) {
	rankingService := &RankingService{}
	err := rankingService.Reload()
	if err != nil {
		t.Errorf("Expected Reload to return no error, got %v", err)
	}
}

func TestNewRankingService(t *testing.T) {
	cfg := &config.CGRConfig{}
	dm := &DataDBService{}
	cacheS := &engine.CacheS{}
	filterSChan := make(chan *engine.FilterS)
	server := &cores.Server{}
	internalRankingSChan := make(chan birpc.ClientConnector)
	connMgr := &engine.ConnManager{}
	anz := &AnalyzerService{}
	srvDep := make(map[string]*sync.WaitGroup)
	service := NewRankingService(
		cfg,
		dm,
		cacheS,
		filterSChan,
		server,
		internalRankingSChan,
		connMgr,
		anz,
		srvDep,
	)
	rankingService, ok := service.(*RankingService)

	if !ok {
		t.Fatalf("Expected *RankingService, got %T", service)
	}

	if rankingService.cfg != cfg {
		t.Errorf("Expected cfg field to be %v, got %v", cfg, rankingService.cfg)
	}

	if rankingService.dm != dm {
		t.Errorf("Expected dm field to be %v, got %v", dm, rankingService.dm)
	}

	if rankingService.cacheS != cacheS {
		t.Errorf("Expected cacheS field to be %v, got %v", cacheS, rankingService.cacheS)
	}

	if rankingService.filterSChan != filterSChan {
		t.Errorf("Expected filterSChan field to be %v, got %v", filterSChan, rankingService.filterSChan)
	}

	if rankingService.server != server {
		t.Errorf("Expected server field to be %v, got %v", server, rankingService.server)
	}

	if rankingService.connChan != internalRankingSChan {
		t.Errorf("Expected connChan field to be %v, got %v", internalRankingSChan, rankingService.connChan)
	}

	if rankingService.connMgr != connMgr {
		t.Errorf("Expected connMgr field to be %v, got %v", connMgr, rankingService.connMgr)
	}

	if rankingService.anz != anz {
		t.Errorf("Expected anz field to be %v, got %v", anz, rankingService.anz)
	}

	if len(rankingService.srvDep) != len(srvDep) {
		t.Errorf("Expected srvDep to have %d entries, got %d", len(srvDep), len(rankingService.srvDep))
	}
}

func TestIsRunning(t *testing.T) {
	rg := &RankingService{}
	result := rg.IsRunning()
	if result != false {
		t.Errorf("Expected IsRunning to return false, got %v", result)
	}
}

func TestNewRankingSv1(t *testing.T) {
	rankingS := &engine.RankingS{}
	rankingSv1 := v1.NewRankingSv1(rankingS)
	if rankingSv1 == nil {
		t.Fatal("Expected a non-nil RankingSv1 instance, got nil")
	}
}
