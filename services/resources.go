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

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/commonlisteners"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewResourceService returns the Resource Service
func NewResourceService(cfg *config.CGRConfig, dm *DataDBService,
	filterSChan chan *engine.FilterS,
	clSChan chan *commonlisteners.CommonListenerS,
	connMgr *engine.ConnManager,
	srvDep map[string]*sync.WaitGroup,
	srvIndexer *servmanager.ServiceIndexer) servmanager.Service {
	return &ResourceService{
		cfg:         cfg,
		dm:          dm,
		filterSChan: filterSChan,
		clSChan:     clSChan,
		connMgr:     connMgr,
		srvDep:      srvDep,
		srvIndexer:  srvIndexer,
		stateDeps:   NewStateDependencies([]string{utils.StateServiceUP}),
	}
}

// ResourceService implements Service interface
type ResourceService struct {
	sync.RWMutex

	clSChan     chan *commonlisteners.CommonListenerS
	dm          *DataDBService
	filterSChan chan *engine.FilterS

	reS *engine.ResourceS
	cl  *commonlisteners.CommonListenerS

	connMgr *engine.ConnManager
	cfg     *config.CGRConfig
	srvDep  map[string]*sync.WaitGroup

	intRPCconn birpc.ClientConnector       // expose API methods over internal connection
	srvIndexer *servmanager.ServiceIndexer // access directly services from here
	stateDeps  *StateDependencies          // channel subscriptions for state changes
}

// Start should handle the service start
func (reS *ResourceService) Start(ctx *context.Context, _ context.CancelFunc) (err error) {
	if reS.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}

	reS.srvDep[utils.DataDB].Add(1)
	reS.cl = <-reS.clSChan
	reS.clSChan <- reS.cl
	cacheS := reS.srvIndexer.GetService(utils.CacheS).(*CacheService)
	if utils.StructChanTimeout(cacheS.StateChan(utils.StateServiceUP), reS.cfg.GeneralCfg().ConnectTimeout) {
		return utils.NewServiceStateTimeoutError(utils.ResourceS, utils.CacheS, utils.StateServiceUP)
	}
	if err = cacheS.WaitToPrecache(ctx,
		utils.CacheResourceProfiles,
		utils.CacheResources,
		utils.CacheResourceFilterIndexes); err != nil {
		return
	}
	var filterS *engine.FilterS
	if filterS, err = waitForFilterS(ctx, reS.filterSChan); err != nil {
		return
	}
	var datadb *engine.DataManager
	if datadb, err = reS.dm.WaitForDM(ctx); err != nil {
		return
	}
	anz := reS.srvIndexer.GetService(utils.AnalyzerS).(*AnalyzerService)
	if utils.StructChanTimeout(anz.StateChan(utils.StateServiceUP), reS.cfg.GeneralCfg().ConnectTimeout) {
		return utils.NewServiceStateTimeoutError(utils.ResourceS, utils.AnalyzerS, utils.StateServiceUP)
	}

	reS.Lock()
	defer reS.Unlock()
	reS.reS = engine.NewResourceService(datadb, reS.cfg, filterS, reS.connMgr)
	utils.Logger.Info(fmt.Sprintf("<%s> starting <%s> subsystem", utils.CoreS, utils.ResourceS))
	reS.reS.StartLoop(ctx)
	srv, _ := engine.NewService(reS.reS)
	// srv, _ := birpc.NewService(apis.NewResourceSv1(reS.reS), "", false)
	if !reS.cfg.DispatcherSCfg().Enabled {
		for _, s := range srv {
			reS.cl.RpcRegister(s)
		}
	}

	reS.intRPCconn = anz.GetInternalCodec(srv, utils.ResourceS)
	close(reS.stateDeps.StateChan(utils.StateServiceUP))
	return
}

// Reload handles the change of config
func (reS *ResourceService) Reload(ctx *context.Context, _ context.CancelFunc) (err error) {
	reS.Lock()
	reS.reS.Reload(ctx)
	reS.Unlock()
	return
}

// Shutdown stops the service
func (reS *ResourceService) Shutdown() (err error) {
	defer reS.srvDep[utils.DataDB].Done()
	reS.Lock()
	defer reS.Unlock()
	reS.reS.Shutdown(context.TODO()) //we don't verify the error because shutdown never returns an error
	reS.reS = nil
	reS.cl.RpcUnregisterName(utils.ResourceSv1)
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

// StateChan returns signaling channel of specific state
func (reS *ResourceService) StateChan(stateID string) chan struct{} {
	return reS.stateDeps.StateChan(stateID)
}

// IntRPCConn returns the internal connection used by RPCClient
func (reS *ResourceService) IntRPCConn() birpc.ClientConnector {
	return reS.intRPCconn
}
