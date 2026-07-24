// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package services

import (
	"fmt"
	"sync"

	"github.com/cgrates/birpc"
	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewSupplierService returns the Supplier Service
func NewSupplierService(cfg *config.CGRConfig, dm *DataDBService,
	cacheS *engine.CacheS, filterSChan chan *engine.FilterS,
	server *utils.Server, internalSupplierSChan chan birpc.ClientConnector,
	connMgr *engine.ConnManager) servmanager.Service {
	return &SupplierService{
		connChan:    internalSupplierSChan,
		cfg:         cfg,
		dm:          dm,
		cacheS:      cacheS,
		filterSChan: filterSChan,
		server:      server,
		connMgr:     connMgr,
	}
}

// SupplierService implements Service interface
type SupplierService struct {
	sync.RWMutex
	cfg         *config.CGRConfig
	dm          *DataDBService
	cacheS      *engine.CacheS
	filterSChan chan *engine.FilterS
	server      *utils.Server
	connMgr     *engine.ConnManager

	splS     *engine.SupplierService
	rpc      *v1.SupplierSv1
	connChan chan birpc.ClientConnector
}

// Start should handle the sercive start
func (splS *SupplierService) Start() (err error) {
	if splS.IsRunning() {
		return fmt.Errorf("service aleady running")
	}

	<-splS.cacheS.GetPrecacheChannel(utils.CacheSupplierProfiles)
	<-splS.cacheS.GetPrecacheChannel(utils.CacheSupplierFilterIndexes)

	filterS := <-splS.filterSChan
	splS.filterSChan <- filterS
	dbchan := splS.dm.GetDMChan()
	datadb := <-dbchan
	dbchan <- datadb

	splS.Lock()
	defer splS.Unlock()
	splS.splS, err = engine.NewSupplierService(datadb, filterS, splS.cfg,
		splS.connMgr)
	if err != nil {
		utils.Logger.Crit(fmt.Sprintf("<%s> Could not init, error: %s",
			utils.SupplierS, err.Error()))
		return
	}

	utils.Logger.Info(fmt.Sprintf("<%s> starting <%s> subsystem", utils.CoreS, utils.SupplierS))
	splS.rpc = v1.NewSupplierSv1(splS.splS)
	if !splS.cfg.DispatcherSCfg().Enabled {
		splS.server.RpcRegister(splS.rpc)
	}
	splS.connChan <- splS.rpc
	return
}

// Reload handles the change of config
func (splS *SupplierService) Reload() (err error) {
	return
}

// Shutdown stops the service
func (splS *SupplierService) Shutdown() (err error) {
	splS.Lock()
	defer splS.Unlock()
	if err = splS.splS.Shutdown(); err != nil {
		return
	}
	splS.splS = nil
	splS.rpc = nil
	<-splS.connChan
	return
}

// IsRunning returns if the service is running
func (splS *SupplierService) IsRunning() bool {
	splS.RLock()
	defer splS.RUnlock()
	return splS != nil && splS.splS != nil
}

// ServiceName returns the service name
func (splS *SupplierService) ServiceName() string {
	return utils.SupplierS
}

// ShouldRun returns if the service should be running
func (splS *SupplierService) ShouldRun() bool {
	return splS.cfg.SupplierSCfg().Enabled
}
