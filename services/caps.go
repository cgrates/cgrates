// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package services

import (
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewCapService instantiates a new CapService.
func NewCapService(cfg *config.CGRConfig) *CapService {
	return &CapService{
		cfg: cfg,
	}
}

// CapService implements Service interface.
type CapService struct {
	cfg  *config.CGRConfig
	caps *engine.Caps
}

// Start handles the service start.
func (s *CapService) Start(_ *utils.SyncedChan, registry *servmanager.Registry) error {
	s.caps = engine.NewCaps(s.cfg.CoreSCfg().Caps, s.cfg.CoreSCfg().CapsStrategy)
	return nil
}

// Reload handles the config changes.
func (s *CapService) Reload(_ *utils.SyncedChan, _ *servmanager.Registry) error {
	return nil
}

// Shutdown stops the service.
func (s *CapService) Shutdown(_ *servmanager.Registry) error {
	return nil
}

// ServiceName returns the service name
func (s *CapService) ServiceName() string {
	return utils.CapS
}

// ShouldRun returns if the service should be running.
func (s *CapService) ShouldRun() bool {
	return true
}

// Caps returns the Caps object.
func (s *CapService) Caps() *engine.Caps {
	return s.caps
}
