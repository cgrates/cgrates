// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package services

import (
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/guardian"
)

// NewGuardianService instantiates a new GuardianService.
func NewGuardianService(cfg *config.CGRConfig) *GuardianService {
	return &GuardianService{
		cfg: cfg,
	}
}

// GuardianService implements Service interface.
type GuardianService struct {
	cfg    *config.CGRConfig
	locker *guardian.GuardianLocker
}

// Start handles the service start.
func (s *GuardianService) Start(shutdown *utils.SyncedChan, registry *servmanager.Registry) error {
	_, err := registry.WaitForServices(shutdown, utils.StateServiceUP,
		[]string{
			utils.LoggerS,
		},
		s.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return err
	}

	s.locker = engine.NewGuardianLocker(s.cfg)
	guardian.Guardian = s.locker

	return nil
}

// Reload handles the config changes.
func (s *GuardianService) Reload(_ *utils.SyncedChan, _ *servmanager.Registry) error {
	return nil
}

// Shutdown stops the service.
func (s *GuardianService) Shutdown(registry *servmanager.Registry) error {
	return nil
}

// ServiceName returns the service name
func (s *GuardianService) ServiceName() string {
	return utils.GuardianS
}

// ShouldRun returns if the service should be running.
func (s *GuardianService) ShouldRun() bool {
	return true
}

// Locker returns the process Guardian locker.
func (s *GuardianService) Locker() *guardian.GuardianLocker {
	return s.locker
}
