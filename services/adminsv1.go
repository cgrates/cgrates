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

	"github.com/cgrates/cgrates/apis"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewAdminSv1Service returns the AdminSv1 Service
func NewAdminSv1Service(cfg *config.CGRConfig) *AdminSv1Service {
	return &AdminSv1Service{
		cfg: cfg,
	}
}

// AdminSv1Service implements Service interface
type AdminSv1Service struct {
	mu       sync.RWMutex
	cfg      *config.CGRConfig
	api      *apis.AdminSv1
	stopChan chan struct{}
}

// Start should handle the sercive start
// For this service the start should be called from RAL Service
func (s *AdminSv1Service) Start(shutdown *utils.SyncedChan, registry *servmanager.Registry) (err error) {
	srvDeps, err := registry.WaitForServices(shutdown, utils.StateServiceUP,
		[]string{
			utils.CommonListenerS,
			utils.ConnManager,
			utils.FilterS,
			utils.DB,
		},
		s.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return err
	}
	cl := srvDeps[utils.CommonListenerS].(*CommonListenerService).CLS()
	cms := srvDeps[utils.ConnManager].(*ConnManagerService)
	fs := srvDeps[utils.FilterS].(*FilterService).FilterS()
	dm := srvDeps[utils.DB].(*DBService).DataManager()

	s.mu.Lock()
	defer s.mu.Unlock()

	s.api = apis.NewAdminSv1(s.cfg, dm, cms.ConnManager(), fs)

	srv, _ := engine.NewService(s.api)
	// srv, _ := birpc.NewService(s.api, "", false)

	for _, s := range srv {
		cl.RpcRegister(s)
	}
	rpl, _ := engine.NewService(apis.NewReplicatorSv1(dm, s.api))
	for _, svc := range rpl {
		cl.RpcRegister(svc)
	}
	cms.AddInternalConn(utils.AdminS, srv)
	return
}

// Reload handles the change of config
func (s *AdminSv1Service) Reload(_ *utils.SyncedChan, _ *servmanager.Registry) (err error) {
	return
}

// Shutdown stops the service
func (s *AdminSv1Service) Shutdown(registry *servmanager.Registry) (err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	// close(s.stopChan)
	s.api = nil
	cl := registry.Lookup(utils.CommonListenerS).(*CommonListenerService).CLS()
	cl.RpcUnregisterName(utils.AdminSv1)
	return
}

// ServiceName returns the service name
func (s *AdminSv1Service) ServiceName() string {
	return utils.AdminS
}

// ShouldRun returns if the service should be running
func (s *AdminSv1Service) ShouldRun() bool {
	return s.cfg.AdminSCfg().Enabled
}
