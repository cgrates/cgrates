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
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewThresholdService returns the Threshold Service
func NewThresholdService(cfg *config.CGRConfig) *ThresholdService {
	return &ThresholdService{
		cfg:       cfg,
		stateDeps: NewStateDependencies([]string{utils.StateServiceUP, utils.StateServiceDOWN}),
	}
}

// ThresholdService implements Service interface
type ThresholdService struct {
	mu        sync.RWMutex
	cfg       *config.CGRConfig
	thrs      *engine.ThresholdS
	stateDeps *StateDependencies // channel subscriptions for state changes
}

// Start should handle the sercive start
func (thrs *ThresholdService) Start(shutdown *utils.SyncedChan, registry *servmanager.ServiceRegistry) (err error) {
	srvDeps, err := WaitForServicesToReachState(utils.StateServiceUP,
		[]string{
			utils.CommonListenerS,
			utils.ConnManager,
			utils.CacheS,
			utils.FilterS,
			utils.DB,
		},
		registry, thrs.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return err
	}
	cl := srvDeps[utils.CommonListenerS].(*CommonListenerService).CLS()
	cms := srvDeps[utils.ConnManager].(*ConnManagerService)
	cacheS := srvDeps[utils.CacheS].(*CacheService)
	if err = cacheS.WaitToPrecache(shutdown,
		utils.CacheThresholdProfiles,
		utils.CacheThresholds,
		utils.CacheThresholdFilterIndexes); err != nil {
		return
	}
	fs := srvDeps[utils.FilterS].(*FilterService)
	dbs := srvDeps[utils.DB].(*DataDBService)

	thrs.mu.Lock()
	defer thrs.mu.Unlock()
	thrs.thrs = engine.NewThresholdService(dbs.DataManager(), thrs.cfg, fs.FilterS(), cms.ConnManager())
	thrs.thrs.StartLoop(context.TODO())
	srv, _ := engine.NewService(thrs.thrs)
	// srv, _ := birpc.NewService(apis.NewThresholdSv1(thrs.thrs), "", false)
	for _, s := range srv {
		cl.RpcRegister(s)
	}
	cms.AddInternalConn(utils.ThresholdS, srv)
	return
}

// Reload handles the change of config
func (thrs *ThresholdService) Reload(_ *utils.SyncedChan, _ *servmanager.ServiceRegistry) error {
	thrs.mu.Lock()
	thrs.thrs.Reload(context.TODO())
	thrs.mu.Unlock()
	return nil
}

// Shutdown stops the service
func (thrs *ThresholdService) Shutdown(registry *servmanager.ServiceRegistry) error {
	thrs.mu.Lock()
	defer thrs.mu.Unlock()
	thrs.thrs.Shutdown(context.TODO())
	thrs.thrs = nil
	cl := registry.Lookup(utils.CommonListenerS).(*CommonListenerService).CLS()
	cl.RpcUnregisterName(utils.ThresholdSv1)
	return nil
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
