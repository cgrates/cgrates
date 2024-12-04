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
	"github.com/cgrates/cgrates/actions"
	"github.com/cgrates/cgrates/commonlisteners"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewActionService returns the Action Service
func NewActionService(cfg *config.CGRConfig, dm *DataDBService,
	cacheS *CacheService, filterSChan chan *engine.FilterS,
	connMgr *engine.ConnManager,
	clSChan chan *commonlisteners.CommonListenerS,
	anzChan chan *AnalyzerService, srvDep map[string]*sync.WaitGroup,
	srvIndexer *servmanager.ServiceIndexer) servmanager.Service {
	return &ActionService{
		connMgr:     connMgr,
		cfg:         cfg,
		dm:          dm,
		cacheS:      cacheS,
		filterSChan: filterSChan,
		clSChan:     clSChan,
		anzChan:     anzChan,
		srvDep:      srvDep,
		rldChan:     make(chan struct{}, 1),
		srvIndexer:  srvIndexer,
		stateDeps:   NewStateDependencies([]string{utils.StateServiceUP}),
	}
}

// ActionService implements Service interface
type ActionService struct {
	sync.RWMutex

	clSChan     chan *commonlisteners.CommonListenerS
	dm          *DataDBService
	anzChan     chan *AnalyzerService
	cacheS      *CacheService
	filterSChan chan *engine.FilterS

	acts *actions.ActionS
	cl   *commonlisteners.CommonListenerS

	rldChan  chan struct{}
	stopChan chan struct{}

	connMgr *engine.ConnManager
	cfg     *config.CGRConfig
	srvDep  map[string]*sync.WaitGroup

	intRPCconn birpc.ClientConnector       // share the API object implementing API calls for internal
	srvIndexer *servmanager.ServiceIndexer // access directly services from here
	stateDeps  *StateDependencies          // channel subscriptions for state changes
}

// Start should handle the service start
func (acts *ActionService) Start(ctx *context.Context, _ context.CancelFunc) (err error) {
	if acts.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}

	acts.cl = <-acts.clSChan
	acts.clSChan <- acts.cl
	if err = acts.cacheS.WaitToPrecache(ctx,
		utils.CacheActionProfiles,
		utils.CacheActionProfilesFilterIndexes); err != nil {
		return
	}
	var filterS *engine.FilterS
	if filterS, err = waitForFilterS(ctx, acts.filterSChan); err != nil {
		return
	}
	var datadb *engine.DataManager
	if datadb, err = acts.dm.WaitForDM(ctx); err != nil {
		return
	}
	anz := <-acts.anzChan
	acts.anzChan <- anz

	acts.Lock()
	defer acts.Unlock()
	acts.acts = actions.NewActionS(acts.cfg, filterS, datadb, acts.connMgr)
	acts.stopChan = make(chan struct{})
	go acts.acts.ListenAndServe(acts.stopChan, acts.rldChan)

	utils.Logger.Info(fmt.Sprintf("<%s> starting <%s> subsystem", utils.CoreS, utils.ActionS))
	srv, err := engine.NewServiceWithPing(acts.acts, utils.ActionSv1, utils.V1Prfx)
	if err != nil {
		return
	}
	// srv, _ := birpc.NewService(apis.NewActionSv1(acts.acts), "", false)
	if !acts.cfg.DispatcherSCfg().Enabled {
		acts.cl.RpcRegister(srv)
	}

	acts.intRPCconn = anz.GetInternalCodec(srv, utils.ActionS)
	close(acts.stateDeps.StateChan(utils.StateServiceUP))
	return
}

// Reload handles the change of config
func (acts *ActionService) Reload(*context.Context, context.CancelFunc) (err error) {
	acts.rldChan <- struct{}{}
	return // for the moment nothing to reload
}

// Shutdown stops the service
func (acts *ActionService) Shutdown() (err error) {
	acts.Lock()
	defer acts.Unlock()
	close(acts.stopChan)
	acts.acts.Shutdown()
	acts.acts = nil
	acts.cl.RpcUnregisterName(utils.ActionSv1)
	return
}

// IsRunning returns if the service is running
func (acts *ActionService) IsRunning() bool {
	acts.RLock()
	defer acts.RUnlock()
	return acts.acts != nil
}

// ServiceName returns the service name
func (acts *ActionService) ServiceName() string {
	return utils.ActionS
}

// ShouldRun returns if the service should be running
func (acts *ActionService) ShouldRun() bool {
	return acts.cfg.ActionSCfg().Enabled
}

// StateChan returns signaling channel of specific state
func (acts *ActionService) StateChan(stateID string) chan struct{} {
	return acts.stateDeps.StateChan(stateID)
}

// IntRPCConn returns the internal connection used by RPCClient
func (acts *ActionService) IntRPCConn() birpc.ClientConnector {
	return acts.intRPCconn
}
