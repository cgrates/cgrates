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

// NewStatService returns the Stat Service
func NewStatService(cfg *config.CGRConfig, connMgr *engine.ConnManager, srvDep map[string]*sync.WaitGroup) *StatService {
	return &StatService{
		cfg:       cfg,
		connMgr:   connMgr,
		srvDep:    srvDep,
		stateDeps: NewStateDependencies([]string{utils.StateServiceUP, utils.StateServiceDOWN}),
	}
}

// StatService implements Service interface
type StatService struct {
	sync.RWMutex

	sts *engine.StatS
	cl  *commonlisteners.CommonListenerS

	connMgr *engine.ConnManager
	cfg     *config.CGRConfig
	srvDep  map[string]*sync.WaitGroup

	intRPCconn birpc.ClientConnector // expose API methods over internal connection
	stateDeps  *StateDependencies    // channel subscriptions for state changes
}

// Start should handle the sercive start
func (sts *StatService) Start(shutdown chan struct{}, registry *servmanager.ServiceRegistry) (err error) {
	sts.srvDep[utils.DataDB].Add(1)

	srvDeps, err := waitForServicesToReachState(utils.StateServiceUP,
		[]string{
			utils.CommonListenerS,
			utils.CacheS,
			utils.FilterS,
			utils.DataDB,
			utils.AnalyzerS,
		},
		registry, sts.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return err
	}
	sts.cl = srvDeps[utils.CommonListenerS].(*CommonListenerService).CLS()
	cacheS := srvDeps[utils.CacheS].(*CacheService)
	if err = cacheS.WaitToPrecache(shutdown,
		utils.CacheStatQueueProfiles,
		utils.CacheStatQueues,
		utils.CacheStatFilterIndexes); err != nil {
		return
	}
	fs := srvDeps[utils.FilterS].(*FilterService)
	dbs := srvDeps[utils.DataDB].(*DataDBService)
	anz := srvDeps[utils.AnalyzerS].(*AnalyzerService)

	sts.Lock()
	defer sts.Unlock()
	sts.sts = engine.NewStatService(dbs.DataManager(), sts.cfg, fs.FilterS(), sts.connMgr)
	sts.sts.StartLoop(context.TODO())
	srv, _ := engine.NewService(sts.sts)
	// srv, _ := birpc.NewService(apis.NewStatSv1(sts.sts), "", false)
	if !sts.cfg.DispatcherSCfg().Enabled {
		for _, s := range srv {
			sts.cl.RpcRegister(s)
		}
	}
	sts.intRPCconn = anz.GetInternalCodec(srv, utils.StatS)
	return
}

// Reload handles the change of config
func (sts *StatService) Reload(_ chan struct{}, _ *servmanager.ServiceRegistry) (err error) {
	sts.Lock()
	sts.sts.Reload(context.TODO())
	sts.Unlock()
	return
}

// Shutdown stops the service
func (sts *StatService) Shutdown(_ *servmanager.ServiceRegistry) (err error) {
	defer sts.srvDep[utils.DataDB].Done()
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

// IntRPCConn returns the internal connection used by RPCClient
func (sts *StatService) IntRPCConn() birpc.ClientConnector {
	return sts.intRPCconn
}
