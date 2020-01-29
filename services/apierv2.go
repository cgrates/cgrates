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
	"fmt"
	"sync"

	v2 "github.com/cgrates/cgrates/apier/v2"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

// NewApierV2Service returns the ApierV2 Service
func NewApierV2Service(apiv1 *ApierV1Service, cfg *config.CGRConfig,
	server *utils.Server,
	internalAPIerV2Chan chan rpcclient.ClientConnector) *ApierV2Service {
	return &ApierV2Service{
		apiv1:    apiv1,
		connChan: internalAPIerV2Chan,
		cfg:      cfg,
		server:   server,
	}
}

// ApierV2Service implements Service interface
type ApierV2Service struct {
	sync.RWMutex
	cfg    *config.CGRConfig
	server *utils.Server

	apiv1    *ApierV1Service
	api      *v2.ApierV2
	connChan chan rpcclient.ClientConnector
}

// Start should handle the sercive start
// For this service the start should be called from RAL Service
func (api *ApierV2Service) Start() (err error) {
	if api.IsRunning() {
		return fmt.Errorf("service aleady running")
	}

	apiV1Chan := api.apiv1.GetApierV1Chan()
	apiV1 := <-apiV1Chan
	apiV1Chan <- apiV1

	api.Lock()
	defer api.Unlock()

	api.api = &v2.ApierV2{
		ApierV1: *apiV1,
	}

	if !api.cfg.DispatcherSCfg().Enabled {
		api.server.RpcRegister(api.api)
	}

	utils.RegisterRpcParams("", &v2.CDRsV2{})
	utils.RegisterRpcParams("", api.api)

	api.connChan <- api.api
	return
}

// GetIntenternalChan returns the internal connection chanel
func (api *ApierV2Service) GetIntenternalChan() (conn chan rpcclient.ClientConnector) {
	return api.connChan
}

// Reload handles the change of config
func (api *ApierV2Service) Reload() (err error) {
	return
}

// Shutdown stops the service
func (api *ApierV2Service) Shutdown() (err error) {
	api.Lock()
	defer api.Unlock()
	api.api = nil
	<-api.connChan
	return
}

// IsRunning returns if the service is running
func (api *ApierV2Service) IsRunning() bool {
	api.RLock()
	defer api.RUnlock()
	return api != nil && api.api != nil
}

// ServiceName returns the service name
func (api *ApierV2Service) ServiceName() string {
	return utils.ApierV2
}

// ShouldRun returns if the service should be running
func (api *ApierV2Service) ShouldRun() bool {
	return api.cfg.RalsCfg().Enabled
}
