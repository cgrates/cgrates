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

// NewResourceService returns the Resource Service
func NewResourceService(cfg *config.CGRConfig, dm *engine.DataManager,
	cacheS *engine.CacheS, filterSChan chan *engine.FilterS,
	server *utils.Server, thrsChan,
	dispatcherChan chan rpcclient.RpcClientConnection) servmanager.Service {
	return &ResourceService{
		connChan:       make(chan rpcclient.RpcClientConnection, 1),
		cfg:            cfg,
		dm:             dm,
		cacheS:         cacheS,
		filterSChan:    filterSChan,
		server:         server,
		thrsChan:       thrsChan,
		dispatcherChan: dispatcherChan,
	}
}

// ResourceService implements Service interface
type ResourceService struct {
	sync.RWMutex
	cfg            *config.CGRConfig
	dm             *engine.DataManager
	cacheS         *engine.CacheS
	filterSChan    chan *engine.FilterS
	server         *utils.Server
	thrsChan       chan rpcclient.RpcClientConnection
	dispatcherChan chan rpcclient.RpcClientConnection

	reS      *engine.ResourceService
	rpc      *v1.ResourceSv1
	connChan chan rpcclient.RpcClientConnection
}

// Start should handle the sercive start
func (reS *ResourceService) Start() (err error) {
	if reS.IsRunning() {
		return fmt.Errorf("service aleady running")
	}

	<-reS.cacheS.GetPrecacheChannel(utils.CacheResourceProfiles)
	<-reS.cacheS.GetPrecacheChannel(utils.CacheResources)
	<-reS.cacheS.GetPrecacheChannel(utils.CacheResourceFilterIndexes)

	filterS := <-reS.filterSChan
	reS.filterSChan <- filterS

	var thdSConn rpcclient.RpcClientConnection
	if thdSConn, err = NewConnection(reS.cfg, reS.thrsChan, reS.dispatcherChan, reS.cfg.ResourceSCfg().ThresholdSConns); err != nil {
		utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to ThresholdS: %s", utils.ResourceS, err.Error()))
		return
	}

	reS.Lock()
	defer reS.Unlock()
	reS.reS, err = engine.NewResourceService(reS.dm, reS.cfg, thdSConn, filterS)
	if err != nil {
		utils.Logger.Crit(fmt.Sprintf("<%s> Could not init, error: %s", utils.ResourceS, err.Error()))
		return
	}
	utils.Logger.Info(fmt.Sprintf("<%s> starting <%s> subsystem", utils.CoreS, utils.ResourceS))
	reS.reS.StartLoop()
	reS.rpc = v1.NewResourceSv1(reS.reS)
	if !reS.cfg.DispatcherSCfg().Enabled {
		reS.server.RpcRegister(reS.rpc)
	}
	reS.connChan <- reS.rpc
	return
}

// GetIntenternalChan returns the internal connection chanel
func (reS *ResourceService) GetIntenternalChan() (conn chan rpcclient.RpcClientConnection) {
	return reS.connChan
}

// Reload handles the change of config
func (reS *ResourceService) Reload() (err error) {
	var thdSConn rpcclient.RpcClientConnection
	if thdSConn, err = NewConnection(reS.cfg, reS.thrsChan, reS.dispatcherChan, reS.cfg.ResourceSCfg().ThresholdSConns); err != nil {
		utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to ThresholdS: %s", utils.ResourceS, err.Error()))
		return
	}
	reS.Lock()
	reS.reS.SetThresholdConnection(thdSConn)
	reS.reS.Reload()
	reS.Unlock()
	return
}

// Shutdown stops the service
func (reS *ResourceService) Shutdown() (err error) {
	reS.Lock()
	defer reS.Unlock()
	if err = reS.reS.Shutdown(); err != nil {
		return
	}
	reS.reS = nil
	reS.rpc = nil
	<-reS.connChan
	return
}

// IsRunning returns if the service is running
func (reS *ResourceService) IsRunning() bool {
	reS.RLock()
	defer reS.RUnlock()
	return reS != nil && reS.reS != nil
}

// ServiceName returns the service name
func (reS *ResourceService) ServiceName() string {
	return utils.ResourceS
}

// ShouldRun returns if the service should be running
func (reS *ResourceService) ShouldRun() bool {
	return reS.cfg.ResourceSCfg().Enabled
}
