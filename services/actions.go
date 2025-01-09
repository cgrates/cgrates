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
	"github.com/cgrates/cgrates/actions"
	"github.com/cgrates/cgrates/commonlisteners"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewActionService returns the Action Service
func NewActionService(cfg *config.CGRConfig,
	connMgr *engine.ConnManager) *ActionService {
	return &ActionService{
		connMgr:   connMgr,
		cfg:       cfg,
		rldChan:   make(chan struct{}, 1),
		stateDeps: NewStateDependencies([]string{utils.StateServiceUP, utils.StateServiceDOWN}),
	}
}

// ActionService implements Service interface
type ActionService struct {
	sync.RWMutex

	acts *actions.ActionS
	cl   *commonlisteners.CommonListenerS

	rldChan  chan struct{}
	stopChan chan struct{}

	connMgr *engine.ConnManager
	cfg     *config.CGRConfig

	intRPCconn birpc.ClientConnector // share the API object implementing API calls for internal
	stateDeps  *StateDependencies    // channel subscriptions for state changes
}

// Start should handle the service start
func (acts *ActionService) Start(shutdown *utils.SyncedChan, registry *servmanager.ServiceRegistry) (err error) {
	srvDeps, err := waitForServicesToReachState(utils.StateServiceUP,
		[]string{
			utils.CommonListenerS,
			utils.CacheS,
			utils.FilterS,
			utils.DataDB,
			utils.AnalyzerS,
		},
		registry, acts.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return err
	}
	acts.cl = srvDeps[utils.CommonListenerS].(*CommonListenerService).CLS()
	cacheS := srvDeps[utils.CacheS].(*CacheService)
	if err = cacheS.WaitToPrecache(shutdown,
		utils.CacheActionProfiles,
		utils.CacheActionProfilesFilterIndexes); err != nil {
		return err
	}
	fs := srvDeps[utils.FilterS].(*FilterService)
	dbs := srvDeps[utils.DataDB].(*DataDBService)
	anz := srvDeps[utils.AnalyzerS].(*AnalyzerService)

	acts.Lock()
	defer acts.Unlock()
	acts.acts = actions.NewActionS(acts.cfg, fs.FilterS(), dbs.DataManager(), acts.connMgr)
	acts.stopChan = make(chan struct{})
	go acts.acts.ListenAndServe(acts.stopChan, acts.rldChan)
	srv, err := engine.NewServiceWithPing(acts.acts, utils.ActionSv1, utils.V1Prfx)
	if err != nil {
		return
	}
	// srv, _ := birpc.NewService(apis.NewActionSv1(acts.acts), "", false)
	if !acts.cfg.DispatcherSCfg().Enabled {
		acts.cl.RpcRegister(srv)
	}

	acts.intRPCconn = anz.GetInternalCodec(srv, utils.ActionS)
	return
}

// Reload handles the change of config
func (acts *ActionService) Reload(_ *utils.SyncedChan, _ *servmanager.ServiceRegistry) (err error) {
	acts.rldChan <- struct{}{}
	return // for the moment nothing to reload
}

// Shutdown stops the service
func (acts *ActionService) Shutdown(_ *servmanager.ServiceRegistry) (err error) {
	acts.Lock()
	defer acts.Unlock()
	close(acts.stopChan)
	acts.acts.Shutdown()
	acts.acts = nil
	acts.cl.RpcUnregisterName(utils.ActionSv1)
	return
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
