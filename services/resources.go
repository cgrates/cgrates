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
func NewResourceService(cfg *config.CGRConfig,
	connMgr *engine.ConnManager,
	srvDep map[string]*sync.WaitGroup) *ResourceService {
	return &ResourceService{
		cfg:       cfg,
		connMgr:   connMgr,
		srvDep:    srvDep,
		stateDeps: NewStateDependencies([]string{utils.StateServiceUP, utils.StateServiceDOWN}),
	}
}

// ResourceService implements Service interface
type ResourceService struct {
	sync.RWMutex

	reS *engine.ResourceS
	cl  *commonlisteners.CommonListenerS

	connMgr *engine.ConnManager
	cfg     *config.CGRConfig
	srvDep  map[string]*sync.WaitGroup

	intRPCconn birpc.ClientConnector // expose API methods over internal connection
	stateDeps  *StateDependencies    // channel subscriptions for state changes
}

// Start should handle the service start
func (reS *ResourceService) Start(shutdown *utils.SyncedChan, registry *servmanager.ServiceRegistry) (err error) {
	reS.srvDep[utils.DataDB].Add(1)

	srvDeps, err := waitForServicesToReachState(utils.StateServiceUP,
		[]string{
			utils.CommonListenerS,
			utils.CacheS,
			utils.FilterS,
			utils.DataDB,
			utils.AnalyzerS,
		},
		registry, reS.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return err
	}
	reS.cl = srvDeps[utils.CommonListenerS].(*CommonListenerService).CLS()
	cacheS := srvDeps[utils.CacheS].(*CacheService)
	if err = cacheS.WaitToPrecache(shutdown,
		utils.CacheResourceProfiles,
		utils.CacheResources,
		utils.CacheResourceFilterIndexes); err != nil {
		return
	}
	fs := srvDeps[utils.FilterS].(*FilterService)
	dbs := srvDeps[utils.DataDB].(*DataDBService)
	anz := srvDeps[utils.AnalyzerS].(*AnalyzerService)

	reS.Lock()
	defer reS.Unlock()
	reS.reS = engine.NewResourceService(dbs.DataManager(), reS.cfg, fs.FilterS(), reS.connMgr)
	reS.reS.StartLoop(context.TODO())
	srv, _ := engine.NewService(reS.reS)
	// srv, _ := birpc.NewService(apis.NewResourceSv1(reS.reS), "", false)
	if !reS.cfg.DispatcherSCfg().Enabled {
		for _, s := range srv {
			reS.cl.RpcRegister(s)
		}
	}

	reS.intRPCconn = anz.GetInternalCodec(srv, utils.ResourceS)
	return
}

// Reload handles the change of config
func (reS *ResourceService) Reload(_ *utils.SyncedChan, _ *servmanager.ServiceRegistry) (err error) {
	reS.Lock()
	reS.reS.Reload(context.TODO())
	reS.Unlock()
	return
}

// Shutdown stops the service
func (reS *ResourceService) Shutdown(_ *servmanager.ServiceRegistry) (err error) {
	defer reS.srvDep[utils.DataDB].Done()
	reS.Lock()
	defer reS.Unlock()
	reS.reS.Shutdown(context.TODO()) //we don't verify the error because shutdown never returns an error
	reS.reS = nil
	reS.cl.RpcUnregisterName(utils.ResourceSv1)
	return
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
