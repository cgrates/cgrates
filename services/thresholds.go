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

// NewThresholdService returns the Threshold Service
func NewThresholdService(cfg *config.CGRConfig, dm *DataDBService,
	cacheS *CacheService, filterSChan chan *engine.FilterS,
	connMgr *engine.ConnManager,
	clSChan chan *commonlisteners.CommonListenerS, internalThresholdSChan chan birpc.ClientConnector,
	anzChan chan *AnalyzerService, srvDep map[string]*sync.WaitGroup,
	srvIndexer *servmanager.ServiceIndexer) servmanager.Service {
	return &ThresholdService{
		connChan:    internalThresholdSChan,
		cfg:         cfg,
		dm:          dm,
		cacheS:      cacheS,
		filterSChan: filterSChan,
		clSChan:     clSChan,
		anzChan:     anzChan,
		srvDep:      srvDep,
		connMgr:     connMgr,
		srvIndexer:  srvIndexer,
	}
}

// ThresholdService implements Service interface
type ThresholdService struct {
	sync.RWMutex

	clSChan     chan *commonlisteners.CommonListenerS
	dm          *DataDBService
	anzChan     chan *AnalyzerService
	cacheS      *CacheService
	filterSChan chan *engine.FilterS

	thrs *engine.ThresholdS
	cl   *commonlisteners.CommonListenerS

	connChan chan birpc.ClientConnector
	connMgr  *engine.ConnManager
	cfg      *config.CGRConfig
	srvDep   map[string]*sync.WaitGroup

	intRPCconn birpc.ClientConnector       // expose API methods over internal connection
	srvIndexer *servmanager.ServiceIndexer // access directly services from here
	stateDeps  *StateDependencies          // channel subscriptions for state changes
}

// Start should handle the sercive start
func (thrs *ThresholdService) Start(ctx *context.Context, _ context.CancelFunc) (err error) {
	if thrs.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}

	thrs.srvDep[utils.DataDB].Add(1)
	thrs.cl = <-thrs.clSChan
	thrs.clSChan <- thrs.cl
	if err = thrs.cacheS.WaitToPrecache(ctx,
		utils.CacheThresholdProfiles,
		utils.CacheThresholds,
		utils.CacheThresholdFilterIndexes); err != nil {
		return
	}
	var filterS *engine.FilterS
	if filterS, err = waitForFilterS(ctx, thrs.filterSChan); err != nil {
		return
	}
	var datadb *engine.DataManager
	if datadb, err = thrs.dm.WaitForDM(ctx); err != nil {
		return
	}
	anz := <-thrs.anzChan
	thrs.anzChan <- anz

	thrs.Lock()
	defer thrs.Unlock()
	thrs.thrs = engine.NewThresholdService(datadb, thrs.cfg, filterS, thrs.connMgr)

	utils.Logger.Info(fmt.Sprintf("<%s> starting <%s> subsystem", utils.CoreS, utils.ThresholdS))
	thrs.thrs.StartLoop(ctx)
	srv, _ := engine.NewService(thrs.thrs)
	// srv, _ := birpc.NewService(apis.NewThresholdSv1(thrs.thrs), "", false)
	if !thrs.cfg.DispatcherSCfg().Enabled {
		for _, s := range srv {
			thrs.cl.RpcRegister(s)
		}
	}
	thrs.intRPCconn = anz.GetInternalCodec(srv, utils.ThresholdS)
	thrs.connChan <- thrs.intRPCconn
	return
}

// Reload handles the change of config
func (thrs *ThresholdService) Reload(ctx *context.Context, _ context.CancelFunc) (_ error) {
	thrs.Lock()
	thrs.thrs.Reload(ctx)
	thrs.Unlock()
	return
}

// Shutdown stops the service
func (thrs *ThresholdService) Shutdown() (_ error) {
	defer thrs.srvDep[utils.DataDB].Done()
	thrs.Lock()
	defer thrs.Unlock()
	thrs.thrs.Shutdown(context.TODO())
	thrs.thrs = nil
	<-thrs.connChan
	thrs.cl.RpcUnregisterName(utils.ThresholdSv1)
	return
}

// IsRunning returns if the service is running
func (thrs *ThresholdService) IsRunning() bool {
	thrs.RLock()
	defer thrs.RUnlock()
	return thrs.thrs != nil
}

// ServiceName returns the service name
func (thrs *ThresholdService) ServiceName() string {
	return utils.ThresholdS
}

// ShouldRun returns if the service should be running
func (thrs *ThresholdService) ShouldRun() bool {
	return thrs.cfg.ThresholdSCfg().Enabled
}

// StateChan returns signaling channel of specific state
func (thrs *ThresholdService) StateChan(stateID string) chan struct{} {
	return thrs.stateDeps.StateChan(stateID)
}

// IntRPCConn returns the internal connection used by RPCClient
func (thrs *ThresholdService) IntRPCConn() birpc.ClientConnector {
	return thrs.intRPCconn
}
