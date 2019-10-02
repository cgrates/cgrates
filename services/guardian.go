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
	"sync"

	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

// NewGuardianService returns the Guardian Service
func NewGuardianService() servmanager.Service {
	return &GuardianService{
		connChan: make(chan rpcclient.RpcClientConnection, 1),
	}
}

// GuardianService implements Service interface
type GuardianService struct {
	sync.RWMutex
	rpc      *v1.GuardianSv1
	connChan chan rpcclient.RpcClientConnection
}

// Start should handle the sercive start
// populates internal channel for RPC conns
func (grd *GuardianService) Start(sp servmanager.ServiceProvider, waitGuardian bool) (err error) {
	// safe to not check GuardianS should never be stoped and then started again
	// if grd.IsRunning() {
	// return fmt.Errorf("service aleady running")
	// }

	grd.Lock()
	defer grd.Unlock()

	grd.rpc = v1.NewGuardianSv1()
	if !sp.GetConfig().DispatcherSCfg().Enabled {
		sp.GetServer().RpcRegister(grd.rpc)
	}
	grd.connChan <- grd.rpc

	return
}

// GetIntenternalChan returns the internal connection chanel
func (grd *GuardianService) GetIntenternalChan() (conn chan rpcclient.RpcClientConnection) {
	return grd.connChan
}

// Reload handles the change of config
func (grd *GuardianService) Reload(sp servmanager.ServiceProvider) (err error) {
	return
}

// Shutdown stops the service
func (grd *GuardianService) Shutdown() (err error) {
	return
}

// GetRPCInterface returns the interface to register for server
func (grd *GuardianService) GetRPCInterface() interface{} {
	return grd.rpc
}

// IsRunning returns if the service is running
func (grd *GuardianService) IsRunning() bool {
	grd.RLock()
	defer grd.RUnlock()
	return grd != nil && grd.rpc != nil
}

// ServiceName returns the service name
func (grd *GuardianService) ServiceName() string {
	return utils.GuardianS
}
