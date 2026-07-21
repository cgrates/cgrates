// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package services

import (
	"sync"

	"github.com/cgrates/cgrates/apis"
	"github.com/cgrates/cgrates/config"
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

	srv, err := newRPCService(s.api, utils.AdminSv1)
	if err != nil {
		return err
	}
	cl.RpcRegister(srv)
	rpl, err := newRPCService(apis.NewReplicatorSv1(dm, s.api), utils.ReplicatorSv1)
	if err != nil {
		return err
	}
	cl.RpcRegister(rpl)
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
