/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package services

import (
	"sync"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/resources"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewResourceService returns the Resource Service
func NewResourceService(cfg *config.CGRConfig) *ResourceService {
	return &ResourceService{
		cfg:       cfg,
		stateDeps: NewStateDependencies([]string{utils.StateServiceUP, utils.StateServiceDOWN}),
	}
}

// ResourceService implements Service interface
type ResourceService struct {
	mu        sync.RWMutex
	cfg       *config.CGRConfig
	reS       *resources.ResourceS
	stateDeps *StateDependencies // channel subscriptions for state changes
}

// Start should handle the service start
func (reS *ResourceService) Start(shutdown *utils.SyncedChan, registry *servmanager.ServiceRegistry) (err error) {
	srvDeps, err := WaitForServicesToReachState(utils.StateServiceUP,
		[]string{
			utils.CommonListenerS,
			utils.ConnManager,
			utils.CacheS,
			utils.FilterS,
			utils.DataDB,
		},
		registry, reS.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return err
	}
	cl := srvDeps[utils.CommonListenerS].(*CommonListenerService).CLS()
	cms := srvDeps[utils.ConnManager].(*ConnManagerService)
	cacheS := srvDeps[utils.CacheS].(*CacheService)
	if err = cacheS.WaitToPrecache(shutdown,
		utils.CacheResourceProfiles,
		utils.CacheResources,
		utils.CacheResourceFilterIndexes); err != nil {
		return
	}
	fs := srvDeps[utils.FilterS].(*FilterService)
	dbs := srvDeps[utils.DataDB].(*DataDBService)

	reS.mu.Lock()
	defer reS.mu.Unlock()
	reS.reS = resources.NewResourceService(dbs.DataManager(), reS.cfg, fs.FilterS(), cms.ConnManager())
	reS.reS.StartLoop(context.TODO())
	srv, _ := engine.NewService(reS.reS)
	// srv, _ := birpc.NewService(apis.NewResourceSv1(reS.reS), "", false)
	for _, s := range srv {
		cl.RpcRegister(s)
	}
	cms.AddInternalConn(utils.ResourceS, srv)
	return
}

// Reload handles the change of config
func (reS *ResourceService) Reload(_ *utils.SyncedChan, _ *servmanager.ServiceRegistry) (err error) {
	reS.mu.Lock()
	reS.reS.Reload(context.TODO())
	reS.mu.Unlock()
	return
}

// Shutdown stops the service
func (reS *ResourceService) Shutdown(registry *servmanager.ServiceRegistry) (err error) {
	reS.mu.Lock()
	defer reS.mu.Unlock()
	reS.reS.Shutdown(context.TODO()) //we don't verify the error because shutdown never returns an error
	reS.reS = nil
	cl := registry.Lookup(utils.CommonListenerS).(*CommonListenerService).CLS()
	cl.RpcUnregisterName(utils.ResourceSv1)
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
