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

// NewRouteService returns the Route Service
func NewRouteService(cfg *config.CGRConfig, dm *DataDBService,
	cacheS *engine.CacheS, filterSChan chan *engine.FilterS,
	server *cores.Server, internalRouteSChan chan birpc.ClientConnector,
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
	connChan chan birpc.ClientConnector
	anz      *AnalyzerService
	srvDep   map[string]*sync.WaitGroup
}

// Start should handle the sercive start
func (routeS *RouteService) Start() error {
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
	routeS.routeS = engine.NewRouteService(datadb, filterS, routeS.cfg, routeS.connMgr)

	utils.Logger.Info(fmt.Sprintf("<%s> starting <%s> subsystem", utils.CoreS, utils.RouteS))
	srv, err := engine.NewService(v1.NewRouteSv1(routeS.routeS))
	if err != nil {
		return err
	}
	if !routeS.cfg.DispatcherSCfg().Enabled {
		routeS.server.RpcRegister(srv)
	}
	routeS.connChan <- routeS.anz.GetInternalCodec(srv, utils.RouteS)
	return nil
}

// Reload handles the change of config
func (routeS *RouteService) Reload() (err error) {
	return
}

// Shutdown stops the service
func (routeS *RouteService) Shutdown() (err error) {
	routeS.Lock()
	defer routeS.Unlock()
	routeS.routeS.Shutdown() //we don't verify the error because shutdown never returns an error
	routeS.routeS = nil
	<-routeS.connChan
	return
}

// IsRunning returns if the service is running
func (routeS *RouteService) IsRunning() bool {
	routeS.RLock()
	defer routeS.RUnlock()
	return routeS.routeS != nil
}

// ServiceName returns the service name
func (routeS *RouteService) ServiceName() string {
	return utils.RouteS
}

// ShouldRun returns if the service should be running
func (routeS *RouteService) ShouldRun() bool {
	return routeS.cfg.RouteSCfg().Enabled
}
