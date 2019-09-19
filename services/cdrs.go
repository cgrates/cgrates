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

// NewCDRServer returns the CDR Service
func NewCDRServer(connChan chan rpcclient.RpcClientConnection) servmanager.Service {
	return &CDRServer{
		connChan: connChan,
	}
}

// CDRServer implements Service interface
// ToDo: Add the rest of functionality
// only the chanel without reload functionality
type CDRServer struct {
	// cdrS    *engine.CDRServer
	// rpc      *v1.CDRsV1
	connChan chan rpcclient.RpcClientConnection
}

// Start should handle the sercive start
func (cdrS *CDRServer) Start(sp servmanager.ServiceProvider, waitCache bool) (err error) {
	// if cdrS.IsRunning() {
	// return fmt.Errorf("service aleady running")
	// }
	return utils.ErrNotImplemented
}

// GetIntenternalChan returns the internal connection chanel
func (cdrS *CDRServer) GetIntenternalChan() (conn chan rpcclient.RpcClientConnection) {
	return cdrS.connChan
}

// Reload handles the change of config
func (cdrS *CDRServer) Reload(sp servmanager.ServiceProvider) (err error) {
	return utils.ErrNotImplemented
}

// Shutdown stops the service
func (cdrS *CDRServer) Shutdown() (err error) {
	return utils.ErrNotImplemented
	// if err = cdrS.cdrS.Shutdown(); err != nil {
	// 	return
	// }
	// cdrS.cdrS = nil
	// cdrS.rpc = nil
	// <-cdrS.connChan
	// return
}

// GetRPCInterface returns the interface to register for server
func (cdrS *CDRServer) GetRPCInterface() interface{} {
	return nil //cdrS.rpc
}

// IsRunning returns if the service is running
func (cdrS *CDRServer) IsRunning() bool {
	return cdrS != nil // && cdrS.cdrS != nil
}

// ServiceName returns the service name
func (cdrS *CDRServer) ServiceName() string {
	return utils.CDRServer
}
