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

	"github.com/cgrates/cgrates/attributes"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewAttributeService returns the Attribute Service
func NewAttributeService(cfg *config.CGRConfig) *AttributeService {
	return &AttributeService{
		cfg:       cfg,
		stateDeps: NewStateDependencies([]string{utils.StateServiceUP, utils.StateServiceDOWN}),
	}
}

// AttributeService implements Service interface
type AttributeService struct {
	mu        sync.Mutex
	cfg       *config.CGRConfig
	stateDeps *StateDependencies
}

// Start should handle the service start
func (attrS *AttributeService) Start(shutdown *utils.SyncedChan, registry *servmanager.ServiceRegistry) (err error) {
	srvDeps, err := WaitForServicesToReachState(utils.StateServiceUP,
		[]string{
			utils.CommonListenerS,
			utils.ConnManager,
			utils.CacheS,
			utils.FilterS,
			utils.DataDB,
		},
		registry, attrS.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return
	}
	cl := srvDeps[utils.CommonListenerS].(*CommonListenerService).CLS()
	cms := srvDeps[utils.ConnManager].(*ConnManagerService)
	cacheS := srvDeps[utils.CacheS].(*CacheService)
	if err = cacheS.WaitToPrecache(shutdown,
		utils.CacheAttributeProfiles,
		utils.CacheAttributeFilterIndexes); err != nil {
		return
	}
	fs := srvDeps[utils.FilterS].(*FilterService).FilterS()
	dm := srvDeps[utils.DataDB].(*DataDBService).DataManager()

	attrS.mu.Lock()
	defer attrS.mu.Unlock()
	attrService := attributes.NewAttributeService(dm, fs, attrS.cfg)
	srv, _ := engine.NewService(attrService)
	// srv, _ := birpc.NewService(attrS.rpc, "", false)
	for _, s := range srv {
		cl.RpcRegister(s)
	}
	cms.AddInternalConn(utils.AttributeS, srv)
	return
}

// Reload handles the change of config
func (attrS *AttributeService) Reload(_ *utils.SyncedChan, _ *servmanager.ServiceRegistry) (err error) {
	return // for the moment nothing to reload
}

// Shutdown stops the service
func (attrS *AttributeService) Shutdown(registry *servmanager.ServiceRegistry) (err error) {
	attrS.mu.Lock()
	defer attrS.mu.Unlock()
	cl := registry.Lookup(utils.CommonListenerS).(*CommonListenerService).CLS()
	cl.RpcUnregisterName(utils.AttributeSv1)
	return
}

// ServiceName returns the service name
func (attrS *AttributeService) ServiceName() string {
	return utils.AttributeS
}

// ShouldRun returns if the service should be running
func (attrS *AttributeService) ShouldRun() bool {
	return attrS.cfg.AttributeSCfg().Enabled
}

// StateChan returns signaling channel of specific state
func (attrS *AttributeService) StateChan(stateID string) chan struct{} {
	return attrS.stateDeps.StateChan(stateID)
}
