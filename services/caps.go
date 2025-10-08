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
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewCapService instantiates a new CapService.
func NewCapService(cfg *config.CGRConfig) *CapService {
	return &CapService{
		cfg:       cfg,
		stateDeps: NewStateDependencies([]string{utils.StateServiceUP, utils.StateServiceDOWN}),
	}
}

// CapService implements Service interface.
type CapService struct {
	cfg       *config.CGRConfig
	caps      *engine.Caps
	stateDeps *StateDependencies // channel subscriptions for state changes
}

// Start handles the service start.
func (s *CapService) Start(_ *utils.SyncedChan, registry *servmanager.ServiceRegistry) error {
	s.caps = engine.NewCaps(s.cfg.CoreSCfg().Caps, s.cfg.CoreSCfg().CapsStrategy)
	return nil
}

// Reload handles the config changes.
func (s *CapService) Reload(_ *utils.SyncedChan, _ *servmanager.ServiceRegistry) error {
	return nil
}

// Shutdown stops the service.
func (s *CapService) Shutdown(_ *servmanager.ServiceRegistry) error {
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

// StateChan returns signaling channel of specific state
func (s *CapService) StateChan(stateID string) chan struct{} {
	return s.stateDeps.StateChan(stateID)
}

// Caps returns the Caps object.
func (s *CapService) Caps() *engine.Caps {
	return s.caps
}
