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

// NewTrendsService returns the TrendS Service
func NewTrendService(cfg *config.CGRConfig, dm *DataDBService,
	cacheS *engine.CacheS, filterSChan chan *engine.FilterS,
	server *cores.Server, internalStatSChan chan birpc.ClientConnector,
	connMgr *engine.ConnManager, anz *AnalyzerService,
	srvDep map[string]*sync.WaitGroup) servmanager.Service {
	return &TrendService{
		connChan:    internalStatSChan,
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

type TrendService struct {
	sync.RWMutex
	cfg         *config.CGRConfig
	dm          *DataDBService
	cacheS      *engine.CacheS
	filterSChan chan *engine.FilterS
	server      *cores.Server
	connMgr     *engine.ConnManager

	trs      *engine.TrendS
	connChan chan birpc.ClientConnector
	anz      *AnalyzerService
	srvDep   map[string]*sync.WaitGroup
}

// Start should handle the sercive start
func (trs *TrendService) Start() error {
	if trs.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}
	trs.srvDep[utils.DataDB].Add(1)
	<-trs.cacheS.GetPrecacheChannel(utils.CacheTrendProfiles)

	filterS := <-trs.filterSChan
	trs.filterSChan <- filterS
	dbchan := trs.dm.GetDMChan()
	datadb := <-dbchan
	dbchan <- datadb

	utils.Logger.Info(fmt.Sprintf("<%s> starting <%s> subsystem",
		utils.CoreS, utils.TrendS))
	trs.Lock()
	defer trs.Unlock()
	trs.trs = engine.NewTrendS(datadb, trs.connMgr, filterS, trs.cfg)
	if err := trs.trs.StartTrendS(); err != nil {
		return err
	}
	srv, err := engine.NewService(v1.NewTrendSv1(trs.trs))
	if err != nil {
		return err
	}
	if !trs.cfg.DispatcherSCfg().Enabled {
		trs.server.RpcRegister(srv)
	}
	trs.connChan <- trs.anz.GetInternalCodec(srv, utils.TrendS)
	return nil
}

// Reload handles the change of config
func (tr *TrendService) Reload() (err error) {
	tr.Lock()
	tr.trs.Reload()
	tr.Unlock()
	return
}

// Shutdown stops the service
func (tr *TrendService) Shutdown() (err error) {
	defer tr.srvDep[utils.DataDB].Done()
	tr.Lock()
	defer tr.Unlock()
	tr.trs.StopTrendS()
	tr.trs = nil
	<-tr.connChan
	return
}

// IsRunning returns if the service is running
func (tr *TrendService) IsRunning() bool {
	tr.RLock()
	defer tr.RUnlock()
	return tr.trs != nil
}

// ServiceName returns the service name
func (tr *TrendService) ServiceName() string {
	return utils.TrendS
}

// ShouldRun returns if the service should be running
func (tr *TrendService) ShouldRun() bool {
	return tr.cfg.TrendSCfg().Enabled
}
