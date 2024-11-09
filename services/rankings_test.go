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

// import (
// 	"sync"
// 	"testing"
//
// 	"github.com/cgrates/birpc"
// 	"github.com/cgrates/cgrates/commonlisteners"
// 	"github.com/cgrates/cgrates/config"
// 	"github.com/cgrates/cgrates/engine"
// 	"github.com/cgrates/cgrates/utils"
// )
//
// func TestNewRankingService(t *testing.T) {
// 	cfg := &config.CGRConfig{}
// 	dm := &DataDBService{}
// 	cacheS := &CacheService{}
// 	filterSChan := make(chan *engine.FilterS)
// 	birpc := make(chan birpc.ClientConnector)
// 	server := &commonlisteners.CommonListenerS{}
// 	connMgr := &engine.ConnManager{}
// 	anz := &AnalyzerService{}
// 	srvDep := make(map[string]*sync.WaitGroup)
//
// 	rankingService := NewRankingService(cfg, dm, cacheS, filterSChan, server, birpc, connMgr, anz, srvDep)
//
// 	if rankingService == nil {
// 		t.Error("Expected non-nil RankingService, got nil")
// 	}
//
// 	if _, ok := rankingService.(*RankingService); !ok {
// 		t.Errorf("Expected type *RankingService, got %T", rankingService)
// 	}
//
// }
//
// func TestRankingServiceName(t *testing.T) {
// 	rankingService := &RankingService{}
//
// 	serviceName := rankingService.ServiceName()
//
// 	expectedServiceName := utils.RankingS
// 	if serviceName != expectedServiceName {
// 		t.Errorf("Expected service name '%s', but got '%s'", expectedServiceName, serviceName)
// 	}
// }
