/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package services

import (
	"sync"

	"github.com/cgrates/cgrates/engine"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewGlobalVarS .
func NewGlobalVarS(cfg *config.CGRConfig) *GlobalVarS {
	return &GlobalVarS{
		cfg:       cfg,
		stateDeps: NewStateDependencies([]string{utils.StateServiceUP, utils.StateServiceDOWN}),
	}
}

// GlobalVarS implements Agent interface
type GlobalVarS struct {
	mu        sync.Mutex
	cfg       *config.CGRConfig
	stateDeps *StateDependencies // channel subscriptions for state changes
}

// Start should handle the sercive start
func (gv *GlobalVarS) Start(_ chan struct{}, _ *servmanager.ServiceRegistry) error {
	engine.SetHTTPPstrTransport(gv.cfg.HTTPCfg().ClientOpts)
	utils.DecimalContext.MaxScale = gv.cfg.GeneralCfg().DecimalMaxScale
	utils.DecimalContext.MinScale = gv.cfg.GeneralCfg().DecimalMinScale
	utils.DecimalContext.Precision = gv.cfg.GeneralCfg().DecimalPrecision
	utils.DecimalContext.RoundingMode = gv.cfg.GeneralCfg().DecimalRoundingMode
	close(gv.stateDeps.StateChan(utils.StateServiceUP))
	return nil
}

// Reload handles the change of config
func (gv *GlobalVarS) Reload(_ chan struct{}, _ *servmanager.ServiceRegistry) error {
	engine.SetHTTPPstrTransport(gv.cfg.HTTPCfg().ClientOpts)
	utils.DecimalContext.MaxScale = gv.cfg.GeneralCfg().DecimalMaxScale
	utils.DecimalContext.MinScale = gv.cfg.GeneralCfg().DecimalMinScale
	utils.DecimalContext.Precision = gv.cfg.GeneralCfg().DecimalPrecision
	utils.DecimalContext.RoundingMode = gv.cfg.GeneralCfg().DecimalRoundingMode
	return nil
}

// Shutdown stops the service
func (gv *GlobalVarS) Shutdown(_ *servmanager.ServiceRegistry) error {
	close(gv.StateChan(utils.StateServiceDOWN))
	return nil
}

// ServiceName returns the service name
func (gv *GlobalVarS) ServiceName() string {
	return utils.GlobalVarS
}

// ShouldRun returns if the service should be running
func (gv *GlobalVarS) ShouldRun() bool {
	return true
}

// StateChan returns signaling channel of specific state
func (gv *GlobalVarS) StateChan(stateID string) chan struct{} {
	return gv.stateDeps.StateChan(stateID)
}

// Lock implements the sync.Locker interface
func (s *GlobalVarS) Lock() {
	s.mu.Lock()
}

// Unlock implements the sync.Locker interface
func (s *GlobalVarS) Unlock() {
	s.mu.Unlock()
}
