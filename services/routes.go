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
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

// NewRouteService returns the Route Service
func NewRouteService(cfg *config.CGRConfig, dm *DataDBService,
	cacheS *engine.CacheS, filterSChan chan *engine.FilterS,
	server *cores.Server, internalRouteSChan chan rpcclient.ClientConnector,
	connMgr *engine.ConnManager, anz *AnalyzerService,
	srvDep map[string]*sync.WaitGroup) servmanager.Service {
	return &RouteService{
		connChan:    internalRouteSChan,
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

// RouteService implements Service interface
type RouteService struct {
	sync.RWMutex
	cfg         *config.CGRConfig
	dm          *DataDBService
	cacheS      *engine.CacheS
	filterSChan chan *engine.FilterS
	server      *cores.Server
	connMgr     *engine.ConnManager

	routeS   *engine.RouteService
	rpc      *v1.RouteSv1
	connChan chan rpcclient.ClientConnector
	anz      *AnalyzerService
	srvDep   map[string]*sync.WaitGroup
}

// Start should handle the sercive start
func (routeS *RouteService) Start() (err error) {
	if routeS.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}

	<-routeS.cacheS.GetPrecacheChannel(utils.CacheRouteProfiles)
	<-routeS.cacheS.GetPrecacheChannel(utils.CacheRouteFilterIndexes)

	filterS := <-routeS.filterSChan
	routeS.filterSChan <- filterS
	dbchan := routeS.dm.GetDMChan()
	datadb := <-dbchan
	dbchan <- datadb

	routeS.Lock()
	defer routeS.Unlock()
	routeS.routeS, err = engine.NewRouteService(datadb, filterS, routeS.cfg,
		routeS.connMgr)
	if err != nil {
		utils.Logger.Crit(fmt.Sprintf("<%s> Could not init, error: %s",
			utils.RouteS, err.Error()))
		return
	}

	utils.Logger.Info(fmt.Sprintf("<%s> starting <%s> subsystem", utils.CoreS, utils.RouteS))
	routeS.rpc = v1.NewRouteSv1(routeS.routeS)
	if !routeS.cfg.DispatcherSCfg().Enabled {
		routeS.server.RpcRegister(routeS.rpc)
	}
	routeS.connChan <- routeS.anz.GetInternalCodec(routeS.rpc, utils.RouteS)
	return
}

// Reload handles the change of config
func (routeS *RouteService) Reload() (err error) {
	return
}

// Shutdown stops the service
func (routeS *RouteService) Shutdown() (err error) {
	routeS.Lock()
	defer routeS.Unlock()
	if err = routeS.routeS.Shutdown(); err != nil {
		return
	}
	routeS.routeS = nil
	routeS.rpc = nil
	<-routeS.connChan
	return
}

// IsRunning returns if the service is running
func (routeS *RouteService) IsRunning() bool {
	routeS.RLock()
	defer routeS.RUnlock()
	return routeS != nil && routeS.routeS != nil
}

// ServiceName returns the service name
func (routeS *RouteService) ServiceName() string {
	return utils.RouteS
}

// ShouldRun returns if the service should be running
func (routeS *RouteService) ShouldRun() bool {
	return routeS.cfg.RouteSCfg().Enabled
}
