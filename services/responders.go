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
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

// NewResponderService returns the Resonder Service
func NewResponderService(connChan chan rpcclient.RpcClientConnection) servmanager.Service {
	return &ResponderService{
		connChan: connChan,
	}
}

// ResponderService implements Service interface
// ToDo: Add the rest of functionality
// only the chanel without reload functionality
type ResponderService struct {
	// resp    *engine.ResponderService
	// rpc      *v1.respV1
	connChan chan rpcclient.RpcClientConnection
}

// Start should handle the sercive start
func (resp *ResponderService) Start(sp servmanager.ServiceProvider, waitCache bool) (err error) {
	// if resp.IsRunning() {
	// return fmt.Errorf("service aleady running")
	// }
	return utils.ErrNotImplemented
}

// GetIntenternalChan returns the internal connection chanel
func (resp *ResponderService) GetIntenternalChan() (conn chan rpcclient.RpcClientConnection) {
	return resp.connChan
}

// Reload handles the change of config
func (resp *ResponderService) Reload(sp servmanager.ServiceProvider) (err error) {
	return utils.ErrNotImplemented
}

// Shutdown stops the service
func (resp *ResponderService) Shutdown() (err error) {
	return utils.ErrNotImplemented
	// if err = resp.resp.Shutdown(); err != nil {
	// 	return
	// }
	// resp.resp = nil
	// resp.rpc = nil
	// <-resp.connChan
	// return
}

// GetRPCInterface returns the interface to register for server
func (resp *ResponderService) GetRPCInterface() interface{} {
	return nil //resp.rpc
}

// IsRunning returns if the service is running
func (resp *ResponderService) IsRunning() bool {
	return resp != nil // && resp.resp != nil
}

// ServiceName returns the service name
func (resp *ResponderService) ServiceName() string {
	return utils.ResponderS
}
