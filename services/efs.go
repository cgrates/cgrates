// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package services

import (
	"sync"

	"github.com/cgrates/birpc"
	"github.com/cgrates/cgrates/apis"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/efs"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// ExportFailoverService is the service structure for ExportFailover
type ExportFailoverService struct {
	mu       sync.Mutex
	efS      *efs.EfS
	srv      *birpc.Service
	stopChan chan struct{}
	cfg      *config.CGRConfig
}

// NewExportFailoverService is the constructor for the TpeService
func NewExportFailoverService(cfg *config.CGRConfig) *ExportFailoverService {
	return &ExportFailoverService{
		cfg: cfg,
	}
}

// Start should handle the service start
func (s *ExportFailoverService) Start(shutdown *utils.SyncedChan, registry *servmanager.Registry) (err error) {
	srvDeps, err := registry.WaitForServices(shutdown, utils.StateServiceUP,
		[]string{
			utils.CommonListenerS,
			utils.ConnManager,
		},
		s.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return
	}
	cl := srvDeps[utils.CommonListenerS].(*CommonListenerService).CLS()
	cms := srvDeps[utils.ConnManager].(*ConnManagerService)

	s.mu.Lock()
	defer s.mu.Unlock()

	s.efS = efs.NewEfs(s.cfg, cms.ConnManager())
	s.stopChan = make(chan struct{})
	s.srv, err = newRPCService(apis.NewEfSv1(s.efS), utils.EfSv1)
	if err != nil {
		return err
	}
	cl.RpcRegister(s.srv)
	cms.AddInternalConn(utils.EFs, s.srv)
	efs.InitFailedPostCache(s.cfg.EFsCfg().FailedPostsTTL, s.cfg.EFsCfg().FailedPostsStaticTTL)
	return
}

// Reload handles the change of config
func (s *ExportFailoverService) Reload(_ *utils.SyncedChan, _ *servmanager.Registry) (err error) {
	return
}

// Shutdown stops the service
func (s *ExportFailoverService) Shutdown(registry *servmanager.Registry) (err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.srv = nil
	close(s.stopChan)

	cl := registry.Lookup(utils.CommonListenerS).(*CommonListenerService).CLS()
	cl.RpcUnregisterName(utils.EfSv1)
	// NEXT SHOULD EXPORT ALL THE SHUTDOWN LOGGERS TO WRITE
	return
}

// ShouldRun returns if the service should be running
func (s *ExportFailoverService) ShouldRun() bool {
	return s.cfg.EFsCfg().Enabled
}

// ServiceName returns the service name
func (s *ExportFailoverService) ServiceName() string {
	return utils.EFs
}
