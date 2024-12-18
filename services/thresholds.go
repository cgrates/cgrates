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

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/commonlisteners"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewThresholdService returns the Threshold Service
func NewThresholdService(cfg *config.CGRConfig,
	srvDep map[string]*sync.WaitGroup) *ThresholdService {
	return &ThresholdService{
		cfg:       cfg,
		srvDep:    srvDep,
		stateDeps: NewStateDependencies([]string{utils.StateServiceUP, utils.StateServiceDOWN}),
	}
}

// ThresholdService implements Service interface
type ThresholdService struct {
	sync.RWMutex
	cfg *config.CGRConfig

	thrs *engine.ThresholdS
	cl   *commonlisteners.CommonListenerS

	srvDep    map[string]*sync.WaitGroup
	stateDeps *StateDependencies // channel subscriptions for state changes
}

// Start should handle the sercive start
func (thrs *ThresholdService) Start(shutdown chan struct{}, registry *servmanager.ServiceRegistry) (err error) {
	thrs.srvDep[utils.DataDB].Add(1)

	srvDeps, err := WaitForServicesToReachState(utils.StateServiceUP,
		[]string{
			utils.CommonListenerS,
			utils.ConnManager,
			utils.CacheS,
			utils.FilterS,
			utils.DataDB,
		},
		registry, thrs.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return err
	}
	thrs.cl = srvDeps[utils.CommonListenerS].(*CommonListenerService).CLS()
	cms := srvDeps[utils.ConnManager].(*ConnManagerService)
	cacheS := srvDeps[utils.CacheS].(*CacheService)
	if err = cacheS.WaitToPrecache(shutdown,
		utils.CacheThresholdProfiles,
		utils.CacheThresholds,
		utils.CacheThresholdFilterIndexes); err != nil {
		return
	}
	fs := srvDeps[utils.FilterS].(*FilterService)
	dbs := srvDeps[utils.DataDB].(*DataDBService)

	thrs.Lock()
	defer thrs.Unlock()
	thrs.thrs = engine.NewThresholdService(dbs.DataManager(), thrs.cfg, fs.FilterS(), cms.ConnManager())
	thrs.thrs.StartLoop(context.TODO())
	srv, _ := engine.NewService(thrs.thrs)
	// srv, _ := birpc.NewService(apis.NewThresholdSv1(thrs.thrs), "", false)
	if !thrs.cfg.DispatcherSCfg().Enabled {
		for _, s := range srv {
			thrs.cl.RpcRegister(s)
		}
	}
	cms.AddInternalConn(utils.ThresholdS, srv)
	close(thrs.stateDeps.StateChan(utils.StateServiceUP))
	return
}

// Reload handles the change of config
func (thrs *ThresholdService) Reload(_ chan struct{}, _ *servmanager.ServiceRegistry) (_ error) {
	thrs.Lock()
	thrs.thrs.Reload(context.TODO())
	thrs.Unlock()
	return
}

// Shutdown stops the service
func (thrs *ThresholdService) Shutdown(_ *servmanager.ServiceRegistry) (_ error) {
	defer thrs.srvDep[utils.DataDB].Done()
	thrs.Lock()
	defer thrs.Unlock()
	thrs.thrs.Shutdown(context.TODO())
	thrs.thrs = nil
	thrs.cl.RpcUnregisterName(utils.ThresholdSv1)
	close(thrs.stateDeps.StateChan(utils.StateServiceDOWN))
	return
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
