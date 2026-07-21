// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package services

import (
	"fmt"
	"sync"

	"github.com/cgrates/cgrates/apis"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/ers"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewEventReaderService returns the EventReader Service
func NewEventReaderService(cfg *config.CGRConfig) *EventReaderService {
	return &EventReaderService{
		rldChan: make(chan struct{}, 1),
		cfg:     cfg,
	}
}

// EventReaderService implements Service interface
type EventReaderService struct {
	mu       sync.RWMutex
	cfg      *config.CGRConfig
	ers      *ers.ERService
	rldChan  chan struct{}
	stopChan chan struct{}
}

// Start should handle the sercive start
func (erS *EventReaderService) Start(shutdown *utils.SyncedChan, registry *servmanager.Registry) (err error) {
	srvDeps, err := registry.WaitForServices(shutdown, utils.StateServiceUP,
		[]string{
			utils.CommonListenerS,
			utils.ConnManager,
			utils.CacheS,
			utils.FilterS,
			utils.DB,
		},
		erS.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return err
	}
	cl := srvDeps[utils.CommonListenerS].(*CommonListenerService).CLS()
	cms := srvDeps[utils.ConnManager].(*ConnManagerService)
	fs := srvDeps[utils.FilterS].(*FilterService)
	dbs := srvDeps[utils.DB].(*DBService)
	cacheS := srvDeps[utils.CacheS].(*CacheService)

	erS.mu.Lock()
	defer erS.mu.Unlock()

	// remake the stop chan
	erS.stopChan = make(chan struct{})

	// build the service
	erS.ers = ers.NewERService(dbs.DataManager(), erS.cfg, cacheS.CacheS(), fs.FilterS(), cms.ConnManager())
	go erS.listenAndServe(erS.ers, erS.stopChan, erS.rldChan, shutdown)

	srv, err := newRPCService(apis.NewErSv1(erS.ers), utils.ErSv1)
	if err != nil {
		return err
	}
	cl.RpcRegister(srv)
	cms.AddInternalConn(utils.ERs, srv)
	return
}

func (erS *EventReaderService) listenAndServe(ers *ers.ERService, stopChan, rldChan chan struct{}, shutdown *utils.SyncedChan) (err error) {
	if err = ers.ListenAndServe(stopChan, rldChan); err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> error: <%v>", utils.ERs, err))
		shutdown.CloseOnce()
	}
	return
}

// Reload handles the change of config
func (erS *EventReaderService) Reload(_ *utils.SyncedChan, _ *servmanager.Registry) (err error) {
	erS.mu.RLock()
	erS.rldChan <- struct{}{}
	erS.mu.RUnlock()
	return
}

// Shutdown stops the service
func (erS *EventReaderService) Shutdown(registry *servmanager.Registry) (err error) {
	erS.mu.Lock()
	defer erS.mu.Unlock()
	close(erS.stopChan)
	erS.ers = nil
	cl := registry.Lookup(utils.CommonListenerS).(*CommonListenerService).CLS()
	cl.RpcUnregisterName(utils.ErSv1)
	return
}

// ServiceName returns the service name
func (erS *EventReaderService) ServiceName() string {
	return utils.ERs
}

// ShouldRun returns if the service should be running
func (erS *EventReaderService) ShouldRun() bool {
	return erS.cfg.ERsCfg().Enabled
}
