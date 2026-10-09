// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package services

import (
	"github.com/cgrates/cgrates/commonlisteners"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewCommonListenerService instantiates a new CommonListenerService.
func NewCommonListenerService(cfg *config.CGRConfig) *CommonListenerService {
	return &CommonListenerService{
		cfg: cfg,
	}
}

// CommonListenerService implements Service interface.
type CommonListenerService struct {
	cfg *config.CGRConfig
	cls *commonlisteners.CommonListenerS
}

// Start handles the service start.
func (s *CommonListenerService) Start(shutdown *utils.SyncedChan, registry *servmanager.Registry) error {
	cs, err := registry.WaitForService(shutdown, utils.CapS, utils.StateServiceUP,
		s.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return err
	}
	s.cls = commonlisteners.NewCommonListenerS(cs.(*CapService).Caps())
	if s.cfg.ConfigSCfg().Enabled {
		s.cls.RegisterHTTPFunc(s.cfg.ConfigSCfg().URL, config.HandlerConfigS(s.cfg))
	}
	return nil
}

// Reload handles the config changes.
func (s *CommonListenerService) Reload(_ *utils.SyncedChan, _ *servmanager.Registry) error {
	return nil
}

// Shutdown stops the service.
func (s *CommonListenerService) Shutdown(_ *servmanager.Registry) error {
	return nil
}

// ServiceName returns the service name
func (s *CommonListenerService) ServiceName() string {
	return utils.CommonListenerS
}

// ShouldRun returns if the service should be running.
func (s *CommonListenerService) ShouldRun() bool {
	return true
}

// CLS returns the CommonListenerS object.
func (s *CommonListenerService) CLS() *commonlisteners.CommonListenerS {
	return s.cls
}
