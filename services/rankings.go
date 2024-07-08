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
func (sag *RankingService) Start() error {
	if sag.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}
	sag.srvDep[utils.DataDB].Add(1)
	<-sag.cacheS.GetPrecacheChannel(utils.CacheStatQueueProfiles)
	<-sag.cacheS.GetPrecacheChannel(utils.CacheStatQueues)
	<-sag.cacheS.GetPrecacheChannel(utils.CacheStatFilterIndexes)

	filterS := <-sag.filterSChan
	sag.filterSChan <- filterS
	dbchan := sag.dm.GetDMChan()
	datadb := <-dbchan
	dbchan <- datadb

	utils.Logger.Info(fmt.Sprintf("<%s> starting <%s> subsystem",
		utils.CoreS, utils.RankingS))
	srv, err := engine.NewService(v1.NewRankingSv1())
	if err != nil {
		return err
	}
	if !sag.cfg.DispatcherSCfg().Enabled {
		for _, s := range srv {
			sag.server.RpcRegister(s)
		}
	}
	sag.connChan <- sag.anz.GetInternalCodec(srv, utils.StatS)
	return nil
}

// Reload handles the change of config
func (sag *RankingService) Reload() (err error) {
	return
}

// Shutdown stops the service
func (sag *RankingService) Shutdown() (err error) {
	defer sag.srvDep[utils.DataDB].Done()
	sag.Lock()
	defer sag.Unlock()
	<-sag.connChan
	return
}

// IsRunning returns if the service is running
func (sag *RankingService) IsRunning() bool {
	sag.RLock()
	defer sag.RUnlock()
	return false
}

// ServiceName returns the service name
func (sag *RankingService) ServiceName() string {
	return utils.RankingS
}

// ShouldRun returns if the service should be running
func (sag *RankingService) ShouldRun() bool {
	return sag.cfg.RankingSCfg().Enabled
}
