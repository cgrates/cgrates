// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package services

import (
	"sync"

	"github.com/cgrates/cgrates/apis"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/ees"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewEventExporterService constructs EventExporterService
func NewEventExporterService(cfg *config.CGRConfig) *EventExporterService {
	return &EventExporterService{
		cfg: cfg,
	}
}

// EventExporterService is the service structure for EventExporterS
type EventExporterService struct {
	mu  sync.RWMutex
	cfg *config.CGRConfig
	eeS *ees.EeS
}

// Start should handle the service start
func (es *EventExporterService) Start(shutdown *utils.SyncedChan, registry *servmanager.Registry) error {
	srvDeps, err := registry.WaitForServices(shutdown, utils.StateServiceUP,
		[]string{
			utils.CommonListenerS,
			utils.ConnManager,
			utils.CacheS,
			utils.FilterS,
			utils.DB,
		},
		es.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return err
	}
	cl := srvDeps[utils.CommonListenerS].(*CommonListenerService).CLS()
	cms := srvDeps[utils.ConnManager].(*ConnManagerService)
	fs := srvDeps[utils.FilterS].(*FilterService).FilterS()
	dbs := srvDeps[utils.DB].(*DBService)
	cacheS := srvDeps[utils.CacheS].(*CacheService)

	es.mu.Lock()
	defer es.mu.Unlock()

	es.eeS, err = ees.NewEventExporterS(es.cfg, cacheS.CacheS(), fs, cms.ConnManager(), dbs.DataManager())
	if err != nil {
		return err
	}

	srv, err := newRPCService(apis.NewEeSv1(es.eeS), utils.EeSv1)
	if err != nil {
		return err
	}
	cl.RpcRegister(srv)
	cms.AddInternalConn(utils.EEs, srv)
	return nil
}

// Reload handles the change of config
func (es *EventExporterService) Reload(_ *utils.SyncedChan, _ *servmanager.Registry) error {
	es.mu.Lock()
	defer es.mu.Unlock()
	es.eeS.ClearExporterCache()
	return es.eeS.SetupExporterCache()
}

// Shutdown stops the service
func (es *EventExporterService) Shutdown(registry *servmanager.Registry) error {
	es.mu.Lock()
	defer es.mu.Unlock()
	es.eeS.ClearExporterCache()
	es.eeS = nil
	cl := registry.Lookup(utils.CommonListenerS).(*CommonListenerService).CLS()
	cl.RpcUnregisterName(utils.EeSv1)
	return nil
}

// ServiceName returns the service name
func (es *EventExporterService) ServiceName() string {
	return utils.EEs
}

// ShouldRun returns if the service should be running
func (es *EventExporterService) ShouldRun() (should bool) {
	return es.cfg.EEsCfg().Enabled
}
