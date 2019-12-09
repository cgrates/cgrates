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

	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

// NewSupplierService returns the Supplier Service
func NewSupplierService(cfg *config.CGRConfig, dm *DataDBService,
	cacheS *engine.CacheS, filterSChan chan *engine.FilterS,
	server *utils.Server, internalSupplierSChan chan rpcclient.RpcClientConnection,
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
	connChan chan rpcclient.ClientConnector
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

	splS.Lock()
	defer splS.Unlock()
	splS.splS, err = engine.NewSupplierService(splS.dm.GetDM(), filterS, splS.cfg,
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

// GetIntenternalChan returns the internal connection chanel
func (splS *SupplierService) GetIntenternalChan() (conn chan rpcclient.ClientConnector) {
	return splS.connChan
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
