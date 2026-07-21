// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package services

import (
	"sync"

	"github.com/cgrates/cgrates/apis"
	"github.com/cgrates/cgrates/attributes"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewAttributeService returns the Attribute Service
func NewAttributeService(cfg *config.CGRConfig) *AttributeService {
	return &AttributeService{
		cfg: cfg,
	}
}

// AttributeService implements Service interface
type AttributeService struct {
	mu  sync.Mutex
	cfg *config.CGRConfig
}

// Start should handle the service start
func (attrS *AttributeService) Start(shutdown *utils.SyncedChan, registry *servmanager.Registry) (err error) {
	srvDeps, err := registry.WaitForServices(shutdown, utils.StateServiceUP,
		[]string{
			utils.CommonListenerS,
			utils.ConnManager,
			utils.CacheS,
			utils.FilterS,
			utils.DB,
		},
		attrS.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return
	}
	cl := srvDeps[utils.CommonListenerS].(*CommonListenerService).CLS()
	cms := srvDeps[utils.ConnManager].(*ConnManagerService)
	cacheS := srvDeps[utils.CacheS].(*CacheService)
	if err = cacheS.WaitToPrecache(shutdown,
		utils.CacheAttributeProfiles,
		utils.CacheAttributeFilterIndexes); err != nil {
		return
	}
	fs := srvDeps[utils.FilterS].(*FilterService).FilterS()
	dm := srvDeps[utils.DB].(*DBService).DataManager()

	attrS.mu.Lock()
	defer attrS.mu.Unlock()
	attrService := attributes.NewAttributeService(dm, fs, cms.ConnManager(), attrS.cfg)
	srv, err := newRPCService(apis.NewAttributeSv1(attrService), utils.AttributeSv1)
	if err != nil {
		return err
	}
	cl.RpcRegister(srv)
	cms.AddInternalConn(utils.AttributeS, srv)
	return
}

// Reload handles the change of config
func (attrS *AttributeService) Reload(_ *utils.SyncedChan, _ *servmanager.Registry) (err error) {
	return // for the moment nothing to reload
}

// Shutdown stops the service
func (attrS *AttributeService) Shutdown(registry *servmanager.Registry) (err error) {
	attrS.mu.Lock()
	defer attrS.mu.Unlock()
	cl := registry.Lookup(utils.CommonListenerS).(*CommonListenerService).CLS()
	cl.RpcUnregisterName(utils.AttributeSv1)
	return
}

// ServiceName returns the service name
func (attrS *AttributeService) ServiceName() string {
	return utils.AttributeS
}

// ShouldRun returns if the service should be running
func (attrS *AttributeService) ShouldRun() bool {
	return attrS.cfg.AttributeSCfg().Enabled
}
