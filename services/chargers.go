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

	"github.com/cgrates/cgrates/commonlisteners"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewChargerService returns the Charger Service
func NewChargerService(cfg *config.CGRConfig) *ChargerService {
	return &ChargerService{
		cfg:       cfg,
		stateDeps: NewStateDependencies([]string{utils.StateServiceUP, utils.StateServiceDOWN}),
	}
}

// ChargerService implements Service interface
type ChargerService struct {
	sync.RWMutex
	cfg *config.CGRConfig

	chrS *engine.ChargerS
	cl   *commonlisteners.CommonListenerS

	stateDeps *StateDependencies // channel subscriptions for state changes
}

// Start should handle the service start
func (chrS *ChargerService) Start(shutdown *utils.SyncedChan, registry *servmanager.ServiceRegistry) error {
	srvDeps, err := WaitForServicesToReachState(utils.StateServiceUP,
		[]string{
			utils.CommonListenerS,
			utils.ConnManager,
			utils.CacheS,
			utils.FilterS,
			utils.DataDB,
		},
		registry, chrS.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return err
	}
	chrS.cl = srvDeps[utils.CommonListenerS].(*CommonListenerService).CLS()
	cms := srvDeps[utils.ConnManager].(*ConnManagerService)
	cacheS := srvDeps[utils.CacheS].(*CacheService)
	if err = cacheS.WaitToPrecache(shutdown,
		utils.CacheChargerProfiles,
		utils.CacheChargerFilterIndexes); err != nil {
		return err
	}
	fs := srvDeps[utils.FilterS].(*FilterService)
	dbs := srvDeps[utils.DataDB].(*DataDBService)

	chrS.Lock()
	defer chrS.Unlock()
	chrS.chrS = engine.NewChargerService(dbs.DataManager(), fs.FilterS(), chrS.cfg, cms.ConnManager())
	srv, _ := engine.NewService(chrS.chrS)
	// srv, _ := birpc.NewService(apis.NewChargerSv1(chrS.chrS), "", false)
	if !chrS.cfg.DispatcherSCfg().Enabled {
		for _, s := range srv {
			chrS.cl.RpcRegister(s)
		}
	}
	cms.AddInternalConn(utils.ChargerS, srv)
	return nil
}

// Reload handles the change of config
func (chrS *ChargerService) Reload(_ *utils.SyncedChan, _ *servmanager.ServiceRegistry) (err error) {
	return
}

// Shutdown stops the service
func (chrS *ChargerService) Shutdown(_ *servmanager.ServiceRegistry) (err error) {
	chrS.Lock()
	defer chrS.Unlock()
	chrS.chrS = nil
	chrS.cl.RpcUnregisterName(utils.ChargerSv1)
	return
}

// ServiceName returns the service name
func (chrS *ChargerService) ServiceName() string {
	return utils.ChargerS
}

// ShouldRun returns if the service should be running
func (chrS *ChargerService) ShouldRun() bool {
	return chrS.cfg.ChargerSCfg().Enabled
}

// StateChan returns signaling channel of specific state
func (chrS *ChargerService) StateChan(stateID string) chan struct{} {
	return chrS.stateDeps.StateChan(stateID)
}
