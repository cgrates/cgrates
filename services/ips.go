// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package services

import (
	"sync"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/apis"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/ips"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewIPService returns the IP Service
func NewIPService(cfg *config.CGRConfig) *IPService {
	return &IPService{
		cfg: cfg,
	}
}

// IPService implements Service interface
type IPService struct {
	mu  sync.RWMutex
	cfg *config.CGRConfig
	ips *ips.IPs
}

// Start handles the service start.
func (s *IPService) Start(shutdown *utils.SyncedChan, registry *servmanager.Registry) error {
	srvDeps, err := registry.WaitForServices(shutdown, utils.StateServiceUP,
		[]string{
			utils.CommonListenerS,
			utils.ConnManager,
			utils.CacheS,
			utils.FilterS,
			utils.DB,
		},
		s.cfg.GeneralCfg().ConnectTimeout)
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
	dbs := srvDeps[utils.DB].(*DBService)

	s.mu.Lock()
	defer s.mu.Unlock()
	s.ips = ips.NewIPService(s.cfg, dbs.DataManager(), cacheS.CacheS(), fs.FilterS(), cms.ConnManager())
	s.ips.StartLoop(context.TODO())
	srv, err := newRPCService(apis.NewIPSv1(s.ips), utils.IPsV1)
	if err != nil {
		return err
	}
	cl.RpcRegister(srv)
	cms.AddInternalConn(utils.IPs, srv)
	return nil
}

// Reload handles configuration changes.
func (s *IPService) Reload(_ *utils.SyncedChan, _ *servmanager.Registry) error {
	s.mu.Lock()
	s.ips.Reload(context.TODO())
	s.mu.Unlock()
	return nil
}

// Shutdown stops the service.
func (s *IPService) Shutdown(registry *servmanager.Registry) error {
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
