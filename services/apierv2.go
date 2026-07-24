// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package services

import (
	"sync"

	"github.com/cgrates/birpc"
	v2 "github.com/cgrates/cgrates/apier/v2"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

// NewAPIerSv2Service returns the APIerSv2 Service
func NewAPIerSv2Service(apiv1 *APIerSv1Service, cfg *config.CGRConfig,
	server *utils.Server,
	internalAPIerSv2Chan chan birpc.ClientConnector) *APIerSv2Service {
	return &APIerSv2Service{
		apiv1:    apiv1,
		connChan: internalAPIerSv2Chan,
		cfg:      cfg,
		server:   server,
	}
}

// APIerSv2Service implements Service interface
type APIerSv2Service struct {
	sync.RWMutex
	cfg    *config.CGRConfig
	server *utils.Server

	apiv1    *APIerSv1Service
	api      *v2.APIerSv2
	connChan chan birpc.ClientConnector
}

// Start should handle the sercive start
// For this service the start should be called from RAL Service
func (api *APIerSv2Service) Start() (err error) {
	if api.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}

	apiV1Chan := api.apiv1.GetAPIerSv1Chan()
	apiV1 := <-apiV1Chan
	apiV1Chan <- apiV1

	api.Lock()
	defer api.Unlock()

	api.api = &v2.APIerSv2{
		APIerSv1: *apiV1,
	}

	if !api.cfg.DispatcherSCfg().Enabled {
		api.server.RpcRegister(api.api)
		api.server.RpcRegisterName(utils.ApierV2, api.api)
	}

	utils.RegisterRpcParams("", &v2.CDRsV2{})
	utils.RegisterRpcParams("", api.api)
	utils.RegisterRpcParams(utils.ApierV2, api.api)

	api.connChan <- api.api
	return
}

// Reload handles the change of config
func (api *APIerSv2Service) Reload() (err error) {
	return
}

// Shutdown stops the service
func (api *APIerSv2Service) Shutdown() (err error) {
	api.Lock()
	defer api.Unlock()
	api.api = nil
	<-api.connChan
	return
}

// IsRunning returns if the service is running
func (api *APIerSv2Service) IsRunning() bool {
	api.RLock()
	defer api.RUnlock()
	return api != nil && api.api != nil
}

// ServiceName returns the service name
func (api *APIerSv2Service) ServiceName() string {
	return utils.APIerSv2
}

// ShouldRun returns if the service should be running
func (api *APIerSv2Service) ShouldRun() bool {
	return api.cfg.ApierCfg().Enabled
}
