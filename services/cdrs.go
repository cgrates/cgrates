// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package services

import (
	"runtime"
	"sync"

	"github.com/cgrates/cgrates/apis"
	"github.com/cgrates/cgrates/cdrs"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewCDRServer returns the CDR Server
func NewCDRServer(cfg *config.CGRConfig) *CDRService {
	return &CDRService{
		cfg: cfg,
	}
}

// CDRService implements Service interface
type CDRService struct {
	mu   sync.RWMutex
	cfg  *config.CGRConfig
	cdrS *cdrs.CDRServer
}

// Start should handle the sercive start
func (cs *CDRService) Start(shutdown *utils.SyncedChan, registry *servmanager.Registry) (err error) {
	srvDeps, err := registry.WaitForServices(shutdown, utils.StateServiceUP,
		[]string{
			utils.CommonListenerS,
			utils.ConnManager,
			utils.CacheS,
			utils.FilterS,
			utils.DB,
		},
		cs.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return err
	}
	cl := srvDeps[utils.CommonListenerS].(*CommonListenerService).CLS()
	cms := srvDeps[utils.ConnManager].(*ConnManagerService)
	fs := srvDeps[utils.FilterS].(*FilterService).FilterS()
	dbs := srvDeps[utils.DB].(*DBService)
	cacheS := srvDeps[utils.CacheS].(*CacheService)

	cs.mu.Lock()
	defer cs.mu.Unlock()

	cs.cdrS = cdrs.NewCDRServer(cs.cfg, dbs.DataManager(), cacheS.CacheS(), fs, cms.ConnManager())
	runtime.Gosched()
	srv, err := newRPCService(apis.NewCdrSv1(cs.cdrS), utils.CDRsV1)
	if err != nil {
		return err
	}
	cl.RpcRegister(srv)
	cms.AddInternalConn(utils.CDRServer, srv)
	return
}

// Reload handles the change of config
func (cs *CDRService) Reload(_ *utils.SyncedChan, _ *servmanager.Registry) (err error) {
	return
}

// Shutdown stops the service
func (cs *CDRService) Shutdown(registry *servmanager.Registry) (err error) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.cdrS = nil
	cl := registry.Lookup(utils.CommonListenerS).(*CommonListenerService).CLS()
	cl.RpcUnregisterName(utils.CDRsV1)
	return
}

// ServiceName returns the service name
func (cs *CDRService) ServiceName() string {
	return utils.CDRServer
}

// ShouldRun returns if the service should be running
func (cs *CDRService) ShouldRun() bool {
	return cs.cfg.CdrsCfg().Enabled
}
