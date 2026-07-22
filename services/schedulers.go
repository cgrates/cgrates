// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package services

import (
	"sync"

	"github.com/cgrates/birpc"
	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/scheduler"
	"github.com/cgrates/cgrates/utils"
)

// NewSchedulerService returns the Scheduler Service
func NewSchedulerService(cfg *config.CGRConfig, dm *DataDBService,
	cacheS *engine.CacheS, fltrSChan chan *engine.FilterS,
	server *cores.Server, internalSchedulerrSChan chan birpc.ClientConnector,
	connMgr *engine.ConnManager, anz *AnalyzerService,
	srvDep map[string]*sync.WaitGroup) *SchedulerService {
	return &SchedulerService{
		connChan:  internalSchedulerrSChan,
		cfg:       cfg,
		dm:        dm,
		cacheS:    cacheS,
		fltrSChan: fltrSChan,
		server:    server,
		connMgr:   connMgr,
		anz:       anz,
		srvDep:    srvDep,
	}
}

// SchedulerService implements Service interface
type SchedulerService struct {
	sync.RWMutex
	cfg       *config.CGRConfig
	dm        *DataDBService
	cacheS    *engine.CacheS
	fltrSChan chan *engine.FilterS
	server    *cores.Server

	schS     *scheduler.Scheduler
	connChan chan birpc.ClientConnector
	connMgr  *engine.ConnManager
	anz      *AnalyzerService
	srvDep   map[string]*sync.WaitGroup
}

// Start should handle the sercive start
func (schS *SchedulerService) Start() error {
	if schS.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}

	<-schS.cacheS.GetPrecacheChannel(utils.CacheActionPlans) // wait for ActionPlans to be cached

	fltrS := <-schS.fltrSChan
	schS.fltrSChan <- fltrS
	dbchan := schS.dm.GetDMChan()
	datadb := <-dbchan
	dbchan <- datadb

	schS.Lock()
	defer schS.Unlock()
	utils.Logger.Info("<ServiceManager> Starting CGRateS Scheduler.")
	schS.schS = scheduler.NewScheduler(datadb, schS.cfg, fltrS)
	go schS.schS.Loop()

	srv, err := engine.NewService(v1.NewSchedulerSv1(schS.cfg, datadb, fltrS))
	if err != nil {
		return err
	}
	if !schS.cfg.DispatcherSCfg().Enabled {
		schS.server.RpcRegister(srv)
	}
	schS.connChan <- schS.anz.GetInternalCodec(srv, utils.SchedulerS)
	return nil
}

// Reload handles the change of config
func (schS *SchedulerService) Reload() (err error) {
	schS.Lock()
	schS.schS.Reload()
	schS.Unlock()
	return
}

// Shutdown stops the service
func (schS *SchedulerService) Shutdown() (err error) {
	schS.Lock()
	schS.schS.Shutdown()
	schS.schS = nil
	<-schS.connChan
	schS.Unlock()
	return
}

// IsRunning returns if the service is running
func (schS *SchedulerService) IsRunning() bool {
	schS.RLock()
	defer schS.RUnlock()
	return schS.schS != nil
}

// ServiceName returns the service name
func (schS *SchedulerService) ServiceName() string {
	return utils.SchedulerS
}

// GetScheduler returns the Scheduler
func (schS *SchedulerService) GetScheduler() *scheduler.Scheduler {
	schS.RLock()
	defer schS.RUnlock()
	return schS.schS
}

// ShouldRun returns if the service should be running
func (schS *SchedulerService) ShouldRun() bool {
	return schS.cfg.SchedulerCfg().Enabled
}
