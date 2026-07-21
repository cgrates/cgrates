// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
	exportLogger, err := engine.NewExportLogger(context.TODO(), s.cfg.GeneralCfg().DefaultTenant, cm, s.cfg)
	if err != nil {
		return err
	}
	utils.Logger = exportLogger
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
