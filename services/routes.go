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

	"github.com/cgrates/birpc/context"

	"github.com/cgrates/birpc"
	"github.com/cgrates/cgrates/commonlisteners"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewRouteService returns the Route Service
func NewRouteService(cfg *config.CGRConfig, dm *DataDBService,
	cacheS *CacheService, filterSChan chan *engine.FilterS,
	clSChan chan *commonlisteners.CommonListenerS,
	connMgr *engine.ConnManager, anzChan chan *AnalyzerService,
	srvIndexer *servmanager.ServiceIndexer) servmanager.Service {
	return &RouteService{
		cfg:         cfg,
		dm:          dm,
		cacheS:      cacheS,
		filterSChan: filterSChan,
		clSChan:     clSChan,
		connMgr:     connMgr,
		anzChan:     anzChan,
		srvIndexer:  srvIndexer,
		stateDeps:   NewStateDependencies([]string{utils.StateServiceUP}),
	}
}

// RouteService implements Service interface
type RouteService struct {
	sync.RWMutex

	clSChan     chan *commonlisteners.CommonListenerS
	dm          *DataDBService
	anzChan     chan *AnalyzerService
	cacheS      *CacheService
	filterSChan chan *engine.FilterS

	routeS *engine.RouteS
	cl     *commonlisteners.CommonListenerS

	connMgr *engine.ConnManager
	cfg     *config.CGRConfig

	intRPCconn birpc.ClientConnector       // expose API methods over internal connection
	srvIndexer *servmanager.ServiceIndexer // access directly services from here
	stateDeps  *StateDependencies          // channel subscriptions for state changes
}

// Start should handle the sercive start
func (routeS *RouteService) Start(ctx *context.Context, _ context.CancelFunc) (err error) {
	if routeS.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}

	routeS.cl = <-routeS.clSChan
	routeS.clSChan <- routeS.cl
	if err = routeS.cacheS.WaitToPrecache(ctx,
		utils.CacheRouteProfiles,
		utils.CacheRouteFilterIndexes); err != nil {
		return
	}
	var filterS *engine.FilterS
	if filterS, err = waitForFilterS(ctx, routeS.filterSChan); err != nil {
		return
	}
	var datadb *engine.DataManager
	if datadb, err = routeS.dm.WaitForDM(ctx); err != nil {
		return
	}
	anz := <-routeS.anzChan
	routeS.anzChan <- anz

	routeS.Lock()
	defer routeS.Unlock()
	routeS.routeS = engine.NewRouteService(datadb, filterS, routeS.cfg, routeS.connMgr)

	utils.Logger.Info(fmt.Sprintf("<%s> starting <%s> subsystem", utils.CoreS, utils.RouteS))
	srv, _ := engine.NewService(routeS.routeS)
	// srv, _ := birpc.NewService(apis.NewRouteSv1(routeS.routeS), "", false)
	if !routeS.cfg.DispatcherSCfg().Enabled {
		for _, s := range srv {
			routeS.cl.RpcRegister(s)
		}
	}
	routeS.intRPCconn = anz.GetInternalCodec(srv, utils.RouteS)
	close(routeS.stateDeps.StateChan(utils.StateServiceUP))
	return
}

// Reload handles the change of config
func (routeS *RouteService) Reload(*context.Context, context.CancelFunc) (err error) {
	return
}

// Shutdown stops the service
func (routeS *RouteService) Shutdown() (err error) {
	routeS.Lock()
	defer routeS.Unlock()
	routeS.routeS.Shutdown() //we don't verify the error because shutdown never returns an error
	routeS.routeS = nil
	routeS.cl.RpcUnregisterName(utils.RouteSv1)
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

// StateChan returns signaling channel of specific state
func (routeS *RouteService) StateChan(stateID string) chan struct{} {
	return routeS.stateDeps.StateChan(stateID)
}

// IntRPCConn returns the internal connection used by RPCClient
func (routeS *RouteService) IntRPCConn() birpc.ClientConnector {
	return routeS.intRPCconn
}
