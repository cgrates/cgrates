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

// NewResourceService returns the Resource Service
func NewResourceService(cfg *config.CGRConfig, dm *DataDBService,
	cacheS *engine.CacheS, filterSChan chan *engine.FilterS,
	server *cores.Server, internalResourceSChan chan birpc.ClientConnector,
	connMgr *engine.ConnManager, anz *AnalyzerService,
	srvDep map[string]*sync.WaitGroup) servmanager.Service {
	return &ResourceService{
		connChan:    internalResourceSChan,
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

// ResourceService implements Service interface
type ResourceService struct {
	sync.RWMutex
	cfg         *config.CGRConfig
	dm          *DataDBService
	cacheS      *engine.CacheS
	filterSChan chan *engine.FilterS
	server      *cores.Server

	reS      *engine.ResourceService
	connChan chan birpc.ClientConnector
	connMgr  *engine.ConnManager
	anz      *AnalyzerService
	srvDep   map[string]*sync.WaitGroup
}

// Start should handle the service start
func (reS *ResourceService) Start() error {
	if reS.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}
	reS.srvDep[utils.DataDB].Add(1)
	<-reS.cacheS.GetPrecacheChannel(utils.CacheResourceProfiles)
	<-reS.cacheS.GetPrecacheChannel(utils.CacheResources)
	<-reS.cacheS.GetPrecacheChannel(utils.CacheResourceFilterIndexes)

	filterS := <-reS.filterSChan
	reS.filterSChan <- filterS
	dbchan := reS.dm.GetDMChan()
	datadb := <-dbchan
	dbchan <- datadb

	reS.Lock()
	defer reS.Unlock()
	reS.reS = engine.NewResourceService(datadb, reS.cfg, filterS, reS.connMgr)
	utils.Logger.Info(fmt.Sprintf("<%s> starting <%s> subsystem", utils.CoreS, utils.ResourceS))
	reS.reS.StartLoop()
	srv, err := engine.NewService(v1.NewResourceSv1(reS.reS))
	if err != nil {
		return err
	}
	if !reS.cfg.DispatcherSCfg().Enabled {
		reS.server.RpcRegister(srv)
	}
	reS.connChan <- reS.anz.GetInternalCodec(srv, utils.ResourceS)
	return nil
}

// Reload handles the change of config
func (reS *ResourceService) Reload() (err error) {
	reS.Lock()
	reS.reS.Reload()
	reS.Unlock()
	return
}

// Shutdown stops the service
func (reS *ResourceService) Shutdown() (err error) {
	defer reS.srvDep[utils.DataDB].Done()
	reS.Lock()
	defer reS.Unlock()
	reS.reS.Shutdown() //we don't verify the error because shutdown never returns an error
	reS.reS = nil
	<-reS.connChan
	return
}

// IsRunning returns if the service is running
func (reS *ResourceService) IsRunning() bool {
	reS.RLock()
	defer reS.RUnlock()
	return reS.reS != nil
}

// ServiceName returns the service name
func (reS *ResourceService) ServiceName() string {
	return utils.ResourceS
}

// ShouldRun returns if the service should be running
func (reS *ResourceService) ShouldRun() bool {
	return reS.cfg.ResourceSCfg().Enabled
}
