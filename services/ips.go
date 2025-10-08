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
	"github.com/cgrates/cgrates/ips"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewIPService returns the IP Service
func NewIPService(cfg *config.CGRConfig) *IPService {
	return &IPService{
		cfg:       cfg,
		stateDeps: NewStateDependencies([]string{utils.StateServiceUP, utils.StateServiceDOWN}),
	}
}

// IPService implements Service interface
type IPService struct {
	mu        sync.RWMutex
	cfg       *config.CGRConfig
	ips       *ips.IPService
	stateDeps *StateDependencies // channel subscriptions for state changes
}

// Start handles the service start.
func (s *IPService) Start(shutdown *utils.SyncedChan, registry *servmanager.ServiceRegistry) error {
	srvDeps, err := WaitForServicesToReachState(utils.StateServiceUP,
		[]string{
			utils.CommonListenerS,
			utils.ConnManager,
			utils.CacheS,
			utils.FilterS,
			utils.DataDB,
		},
		registry, s.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return err
	}
	cl := srvDeps[utils.CommonListenerS].(*CommonListenerService).CLS()
	cms := srvDeps[utils.ConnManager].(*ConnManagerService)
	cacheS := srvDeps[utils.CacheS].(*CacheService)
	if err := cacheS.WaitToPrecache(shutdown,
		utils.CacheIPProfiles,
		utils.CacheIPAllocations,
		utils.CacheIPFilterIndexes); err != nil {
		return err
	}
	fs := srvDeps[utils.FilterS].(*FilterService)
	dbs := srvDeps[utils.DataDB].(*DataDBService)

	s.mu.Lock()
	defer s.mu.Unlock()
	s.ips = ips.NewIPService(dbs.DataManager(), s.cfg, fs.FilterS(), cms.ConnManager())
	s.ips.StartLoop(context.TODO())
	srv, err := engine.NewServiceWithPing(s.ips, utils.IPsV1, utils.V1Prfx)
	if err != nil {
		return err
	}
	cl.RpcRegister(srv)
	cms.AddInternalConn(utils.IPs, srv)
	return nil
}

// Reload handles configuration changes.
func (s *IPService) Reload(_ *utils.SyncedChan, _ *servmanager.ServiceRegistry) error {
	s.mu.Lock()
	s.ips.Reload(context.TODO())
	s.mu.Unlock()
	return nil
}

// Shutdown stops the service.
func (s *IPService) Shutdown(registry *servmanager.ServiceRegistry) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ips.Shutdown(context.TODO()) //we don't verify the error because shutdown never returns an error
	s.ips = nil
	cl := registry.Lookup(utils.CommonListenerS).(*CommonListenerService).CLS()
	cl.RpcUnregisterName(utils.IPsV1)
	return nil
}

// ServiceName returns the service name.
func (s *IPService) ServiceName() string {
	return utils.IPs
}

// ShouldRun returns if the service should be running.
func (s *IPService) ShouldRun() bool {
	return s.cfg.IPsCfg().Enabled
}

// StateChan returns signaling channel of specific state
func (s *IPService) StateChan(stateID string) chan struct{} {
	return s.stateDeps.StateChan(stateID)
}
