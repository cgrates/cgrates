// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package services

import (
	"github.com/cgrates/cgrates/config"
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
	cfg *config.CGRConfig
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

	// TODO: Replace global guardian.Guardian with local instance that should
	// be passed around where needed.
	// Currently only logger option is used, but other options (e.g. for
	// timeout) could be added later.
	opts := make([]guardian.Option, 0, 1)
	if s.cfg.LoggerCfg().Level >= 0 {
		opts = append(opts, guardian.WithLogger(utils.Logger))
	}
	guardian.Guardian = guardian.New(opts...)

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
