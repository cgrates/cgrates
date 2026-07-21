// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package services

import (
	"sync"

	"github.com/cgrates/birpc"
	"github.com/cgrates/cgrates/apis"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/tpes"
	"github.com/cgrates/cgrates/utils"
)

// NewTPeService is the constructor for the TpeService
func NewTPeService(cfg *config.CGRConfig) *TPeService {
	return &TPeService{
		cfg: cfg,
	}
}

// TypeService implements Service interface
type TPeService struct {
	mu       sync.RWMutex
	cfg      *config.CGRConfig
	tpes     *tpes.TPeS
	srv      *birpc.Service
	stopChan chan struct{}
}

// Start should handle the service start
func (ts *TPeService) Start(shutdown *utils.SyncedChan, registry *servmanager.Registry) (err error) {
	srvDeps, err := registry.WaitForServices(shutdown, utils.StateServiceUP,
		[]string{
			utils.CommonListenerS,
			utils.ConnManager,
			utils.DB,
		},
		ts.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return err
	}
	cl := srvDeps[utils.CommonListenerS].(*CommonListenerService).CLS()
	cm := srvDeps[utils.ConnManager].(*ConnManagerService).ConnManager()
	dbs := srvDeps[utils.DB].(*DBService).DataManager()

	ts.mu.Lock()
	defer ts.mu.Unlock()

	ts.tpes = tpes.NewTPeS(ts.cfg, dbs, cm)
	ts.stopChan = make(chan struct{})
	ts.srv, err = newRPCService(apis.NewTPeSv1(ts.tpes), utils.TPeSv1)
	if err != nil {
		return err
	}
	cl.RpcRegister(ts.srv)
	return nil
}

// Reload handles the change of config
func (ts *TPeService) Reload(_ *utils.SyncedChan, _ *servmanager.Registry) (err error) {
	return
}

// Shutdown stops the service
func (ts *TPeService) Shutdown(registry *servmanager.Registry) (err error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.srv = nil
	close(ts.stopChan)
	cl := registry.Lookup(utils.CommonListenerS).(*CommonListenerService).CLS()
	cl.RpcUnregisterName(utils.TPeSv1)
	return
}

// ServiceName returns the service name
func (ts *TPeService) ServiceName() string {
	return utils.TPeS
}

// ShouldRun returns if the service should be running
func (ts *TPeService) ShouldRun() bool {
	return ts.cfg.TpeSCfg().Enabled
}
