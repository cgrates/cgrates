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

// NewStatService returns the Stat Service
func NewStatService(cfg *config.CGRConfig) *StatService {
	return &StatService{
		cfg:       cfg,
		stateDeps: NewStateDependencies([]string{utils.StateServiceUP, utils.StateServiceDOWN}),
	}
}

// StatService implements Service interface
type StatService struct {
	sync.RWMutex
	cfg *config.CGRConfig

	sts *engine.StatS
	cl  *commonlisteners.CommonListenerS

	stateDeps *StateDependencies // channel subscriptions for state changes
}

// Start should handle the sercive start
func (sts *StatService) Start(shutdown *utils.SyncedChan, registry *servmanager.ServiceRegistry) (err error) {
	srvDeps, err := WaitForServicesToReachState(utils.StateServiceUP,
		[]string{
			utils.CommonListenerS,
			utils.ConnManager,
			utils.CacheS,
			utils.FilterS,
			utils.DataDB,
		},
		registry, sts.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return err
	}
	sts.cl = srvDeps[utils.CommonListenerS].(*CommonListenerService).CLS()
	cms := srvDeps[utils.ConnManager].(*ConnManagerService)
	cacheS := srvDeps[utils.CacheS].(*CacheService)
	if err = cacheS.WaitToPrecache(shutdown,
		utils.CacheStatQueueProfiles,
		utils.CacheStatQueues,
		utils.CacheStatFilterIndexes); err != nil {
		return
	}
	fs := srvDeps[utils.FilterS].(*FilterService)
	dbs := srvDeps[utils.DataDB].(*DataDBService)

	sts.Lock()
	defer sts.Unlock()
	sts.sts = engine.NewStatService(dbs.DataManager(), sts.cfg, fs.FilterS(), cms.ConnManager())
	sts.sts.StartLoop(context.TODO())
	srv, _ := engine.NewService(sts.sts)
	// srv, _ := birpc.NewService(apis.NewStatSv1(sts.sts), "", false)
	if !sts.cfg.DispatcherSCfg().Enabled {
		for _, s := range srv {
			sts.cl.RpcRegister(s)
		}
	}
	cms.AddInternalConn(utils.StatS, srv)
	return
}

// Reload handles the change of config
func (sts *StatService) Reload(_ *utils.SyncedChan, _ *servmanager.ServiceRegistry) (err error) {
	sts.Lock()
	sts.sts.Reload(context.TODO())
	sts.Unlock()
	return
}

// Shutdown stops the service
func (sts *StatService) Shutdown(_ *servmanager.ServiceRegistry) (err error) {
	sts.Lock()
	defer sts.Unlock()
	sts.sts.Shutdown(context.TODO())
	sts.sts = nil
	sts.cl.RpcUnregisterName(utils.StatSv1)
	return
}

// ServiceName returns the service name
func (sts *StatService) ServiceName() string {
	return utils.StatS
}

// ShouldRun returns if the service should be running
func (sts *StatService) ShouldRun() bool {
	return sts.cfg.StatSCfg().Enabled
}

// StateChan returns signaling channel of specific state
func (sts *StatService) StateChan(stateID string) chan struct{} {
	return sts.stateDeps.StateChan(stateID)
}
