// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package services

import (
	"github.com/cgrates/cgrates/apis"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewConfigService instantiates a new ConfigService.
func NewConfigService(cfg *config.CGRConfig) *ConfigService {
	return &ConfigService{
		cfg: cfg,
	}
}

// ConfigService implements Service interface.
type ConfigService struct {
	cfg *config.CGRConfig
}

// Start handles the service start.
func (s *ConfigService) Start(shutdown *utils.SyncedChan, registry *servmanager.Registry) error {
	srvDeps, err := registry.WaitForServices(shutdown, utils.StateServiceUP,
		[]string{
			utils.CommonListenerS,
			utils.ConnManager,
		},
		s.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return err
	}
	cl := srvDeps[utils.CommonListenerS].(*CommonListenerService).CLS()
	cms := srvDeps[utils.ConnManager].(*ConnManagerService)

	svc, err := newRPCService(apis.NewConfigSv1(s.cfg), utils.ConfigSv1)
	if err != nil {
		return err
	}
	cl.RpcRegister(svc)
	cms.AddInternalConn(utils.ConfigS, svc)
	return nil
}

// Reload handles the config changes.
func (s *ConfigService) Reload(_ *utils.SyncedChan, _ *servmanager.Registry) error {
	return nil
}

// Shutdown stops the service.
func (s *ConfigService) Shutdown(registry *servmanager.Registry) error {
	cl := registry.Lookup(utils.CommonListenerS).(*CommonListenerService).CLS()
	cl.RpcUnregisterName(utils.ConfigSv1)
	return nil
}

func (s *ConfigService) ServiceName() string {
	return utils.ConfigS
}

// ShouldRun returns if the service should be running.
func (s *ConfigService) ShouldRun() bool {
	return true
}
