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
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewLoggerService instantiates a new LoggerService.
func NewLoggerService(cfg *config.CGRConfig, loggerType string) *LoggerService {
	return &LoggerService{
		cfg:        cfg,
		loggerType: loggerType,
	}
}

// LoggerService implements Service interface.
type LoggerService struct {
	cfg        *config.CGRConfig
	loggerType string
}

// Start handles the service start.
func (s *LoggerService) Start(shutdown *utils.SyncedChan, registry *servmanager.Registry) error {
	if s.loggerType != utils.MetaKafkaLog {
		return nil
	}
	deps := []string{utils.ConnManager}
	if s.cfg.EFsCfg().Enabled {
		deps = append(deps, utils.EFs)
	}
	srvDeps, err := registry.WaitForServices(shutdown, utils.StateServiceUP,
		deps, s.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return err
	}
	cm := srvDeps[utils.ConnManager].(*ConnManagerService).ConnManager()
	utils.Logger = engine.NewExportLogger(context.TODO(), s.cfg.GeneralCfg().DefaultTenant, cm, s.cfg)
	return nil
}

// Reload handles the config changes.
func (s *LoggerService) Reload(_ *utils.SyncedChan, _ *servmanager.Registry) error {
	return nil
}

// Shutdown stops the service.
func (s *LoggerService) Shutdown(_ *servmanager.Registry) error {
	return nil
}

// ServiceName returns the service name
func (s *LoggerService) ServiceName() string {
	return utils.LoggerS
}

// ShouldRun returns if the service should be running.
func (s *LoggerService) ShouldRun() bool {
	return true
}
