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
	"github.com/cgrates/cgrates/efs"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewLoggerService instantiates a new LoggerService.
func NewLoggerService(cfg *config.CGRConfig, loggerType string) *LoggerService {
	return &LoggerService{
		cfg:        cfg,
		loggerType: loggerType,
		stateDeps:  NewStateDependencies([]string{utils.StateServiceUP, utils.StateServiceDOWN}),
	}
}

// LoggerService implements Service interface.
type LoggerService struct {
	cfg        *config.CGRConfig
	stateDeps  *StateDependencies // channel subscriptions for state changes
	loggerType string
}

// Start handles the service start.
func (s *LoggerService) Start(_ *utils.SyncedChan, registry *servmanager.ServiceRegistry) error {
	if s.loggerType != utils.MetaKafkaLog {
		return nil
	}
	// TODO: check if we should also wait for EFs. Currently, in case of *kafka
	// logger, we log to *stdout until initiated. We should also consider
	// removing ErrLoggerChanged error cases if they turn out to be redundant
	// (see engine/kafka_logger.go).
	cms, err := WaitForServiceState(utils.StateServiceUP, utils.ConnManager, registry,
		s.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return err
	}
	cm := cms.(*ConnManagerService).ConnManager()
	utils.Logger = engine.NewExportLogger(context.TODO(), s.cfg.GeneralCfg().DefaultTenant, cm, s.cfg)
	efs.InitFailedPostCache(s.cfg.EFsCfg().FailedPostsTTL, s.cfg.EFsCfg().FailedPostsStaticTTL)
	return nil
}

// Reload handles the config changes.
func (s *LoggerService) Reload(_ *utils.SyncedChan, _ *servmanager.ServiceRegistry) error {
	return nil
}

// Shutdown stops the service.
func (s *LoggerService) Shutdown(_ *servmanager.ServiceRegistry) error {
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

// StateChan returns signaling channel of specific state
func (s *LoggerService) StateChan(stateID string) chan struct{} {
	return s.stateDeps.StateChan(stateID)
}
