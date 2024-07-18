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
	"github.com/cgrates/birpc/context"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewRankingService returns the RankingS Service
func NewRankingService(cfg *config.CGRConfig, dm *DataDBService,
	cacheS *CacheService, filterSChan chan *engine.FilterS,
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
	cacheS      *CacheService
	filterSChan chan *engine.FilterS
	server      *cores.Server
	connMgr     *engine.ConnManager
	connChan    chan birpc.ClientConnector
	anz         *AnalyzerService
	ran         *engine.RankingS
	srvDep      map[string]*sync.WaitGroup
}

// Start should handle the sercive start
func (ran *RankingService) Start(ctx *context.Context, _ context.CancelFunc) (err error) {
	if ran.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}
	ran.srvDep[utils.DataDB].Add(1)
	ran.srvDep[utils.DataDB].Add(1)
	if err = ran.cacheS.WaitToPrecache(ctx, utils.CacheTrendProfiles); err != nil {
		return err
	}
	var datadb *engine.DataManager
	if datadb, err = ran.dm.WaitForDM(ctx); err != nil {
		return
	}

	var filterS *engine.FilterS
	if filterS, err = waitForFilterS(ctx, ran.filterSChan); err != nil {
		return
	}
	ran.ran = engine.NewRankingService(datadb, ran.cfg, filterS, ran.connMgr)
	utils.Logger.Info(fmt.Sprintf("<%s> starting <%s> subsystem",
		utils.CoreS, utils.RankingS))
	srv, _ := engine.NewService(ran.ran)
	if !ran.cfg.DispatcherSCfg().Enabled {
		for _, s := range srv {
			ran.server.RpcRegister(s)
		}
	}
	ran.connChan <- ran.anz.GetInternalCodec(srv, utils.RankingS)
	return nil
}

// Reload handles the change of config
func (ran *RankingService) Reload(ctx *context.Context, _ context.CancelFunc) (err error) {
	return
}

// Shutdown stops the service
func (ran *RankingService) Shutdown() (err error) {
	defer ran.srvDep[utils.DataDB].Done()
	ran.Lock()
	defer ran.Unlock()
	<-ran.connChan
	ran.server.RpcUnregisterName(utils.RankingSv1)
	return
}

// IsRunning returns if the service is running
func (ran *RankingService) IsRunning() bool {
	ran.RLock()
	defer ran.RUnlock()
	return false
}

// ServiceName returns the service name
func (ran *RankingService) ServiceName() string {
	return utils.RankingS
}

// ShouldRun returns if the service should be running
func (ran *RankingService) ShouldRun() bool {
	return ran.cfg.RankingSCfg().Enabled
}
