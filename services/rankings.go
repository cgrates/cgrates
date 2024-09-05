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
	"fmt"
	"sync"

	"github.com/cgrates/birpc"
	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewRankingService returns the RankingS Service
func NewRankingService(cfg *config.CGRConfig, dm *DataDBService,
	cacheS *engine.CacheS, filterSChan chan *engine.FilterS,
	server *cores.Server, internalRankingSChan chan birpc.ClientConnector,
	connMgr *engine.ConnManager, anz *AnalyzerService,
	srvDep map[string]*sync.WaitGroup) servmanager.Service {
	return &RankingService{
		connChan:    internalRankingSChan,
		cfg:         cfg,
		dm:          dm,
		cacheS:      cacheS,
		filterSChan: filterSChan,
		server:      server,
		connMgr:     connMgr,
		anz:         anz,
		srvDep:      srvDep,
	}
}

type RankingService struct {
	sync.RWMutex
	cfg         *config.CGRConfig
	dm          *DataDBService
	cacheS      *engine.CacheS
	filterSChan chan *engine.FilterS
	server      *cores.Server
	connMgr     *engine.ConnManager
	connChan    chan birpc.ClientConnector
	anz         *AnalyzerService
	srvDep      map[string]*sync.WaitGroup
}

// Start should handle the sercive start
func (rg *RankingService) Start() error {
	if rg.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}
	rg.srvDep[utils.DataDB].Add(1)
	<-rg.cacheS.GetPrecacheChannel(utils.CacheRankingProfiles)
	<-rg.cacheS.GetPrecacheChannel(utils.CacheRankingFilterIndexes)

	filterS := <-rg.filterSChan
	rg.filterSChan <- filterS
	dbchan := rg.dm.GetDMChan()
	datadb := <-dbchan
	dbchan <- datadb

	utils.Logger.Info(fmt.Sprintf("<%s> starting <%s> subsystem",
		utils.CoreS, utils.RankingS))
	srv, err := engine.NewService(v1.NewRankingSv1())
	if err != nil {
		return err
	}
	if !rg.cfg.DispatcherSCfg().Enabled {
		rg.server.RpcRegister(srv)
	}
	rg.connChan <- rg.anz.GetInternalCodec(srv, utils.StatS)
	return nil
}

// Reload handles the change of config
func (rg *RankingService) Reload() (err error) {
	return
}

// Shutdown stops the service
func (rg *RankingService) Shutdown() (err error) {
	defer rg.srvDep[utils.DataDB].Done()
	rg.Lock()
	defer rg.Unlock()
	<-rg.connChan
	return
}

// IsRunning returns if the service is running
func (rg *RankingService) IsRunning() bool {
	rg.RLock()
	defer rg.RUnlock()
	return false
}

// ServiceName returns the service name
func (rg *RankingService) ServiceName() string {
	return utils.RankingS
}

// ShouldRun returns if the service should be running
func (rg *RankingService) ShouldRun() bool {
	return rg.cfg.RankingSCfg().Enabled
}
