// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
	rks         *engine.RankingS
}

// Start should handle the sercive start
func (rk *RankingService) Start() error {
	if rk.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}
	rk.srvDep[utils.DataDB].Add(1)
	<-rk.cacheS.GetPrecacheChannel(utils.CacheRankingProfiles)
	<-rk.cacheS.GetPrecacheChannel(utils.CacheRankings)
	filterS := <-rk.filterSChan
	rk.filterSChan <- filterS
	dbchan := rk.dm.GetDMChan()
	datadb := <-dbchan
	dbchan <- datadb
	utils.Logger.Info(fmt.Sprintf("<%s> starting <%s> subsystem",
		utils.CoreS, utils.RankingS))

	rk.rks = engine.NewRankingS(datadb, rk.connMgr, filterS, rk.cfg)
	if err := rk.rks.StartRankingS(); err != nil {
		return err
	}
	srv, err := engine.NewService(v1.NewRankingSv1(rk.rks))
	if err != nil {
		return err
	}
	if !rk.cfg.DispatcherSCfg().Enabled {
		rk.server.RpcRegister(srv)
	}
	rk.connChan <- rk.anz.GetInternalCodec(srv, utils.StatS)
	return nil
}

// Reload handles the change of config
func (rk *RankingService) Reload() (err error) {
	return
}

// Shutdown stops the service
func (rk *RankingService) Shutdown() (err error) {
	defer rk.srvDep[utils.DataDB].Done()
	rk.Lock()
	defer rk.Unlock()
	rk.rks.StopRankingS()
	rk.rks = nil
	<-rk.connChan
	return
}

// IsRunning returns if the service is running
func (rk *RankingService) IsRunning() bool {
	rk.RLock()
	defer rk.RUnlock()
	return rk.rks != nil
}

// ServiceName returns the service name
func (rk *RankingService) ServiceName() string {
	return utils.RankingS
}

// ShouldRun returns if the service should be running
func (rk *RankingService) ShouldRun() bool {
	return rk.cfg.RankingSCfg().Enabled
}
