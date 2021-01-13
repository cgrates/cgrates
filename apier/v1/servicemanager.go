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

package v1

import (
	"github.com/cgrates/cgrates/dispatchers"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

func NewServiceManagerV1(sm *servmanager.ServiceManager) *ServiceManagerV1 {
	return &ServiceManagerV1{sm: sm}
}

type ServiceManagerV1 struct {
	sm *servmanager.ServiceManager // Need to have them capitalize so we can export in V2
}

func (servManager *ServiceManagerV1) StartService(args *dispatchers.ArgStartServiceWithOpts, reply *string) (err error) {
	return servManager.sm.V1StartService(args.ArgStartService, reply)
}

func (servManager *ServiceManagerV1) StopService(args *dispatchers.ArgStartServiceWithOpts, reply *string) (err error) {
	return servManager.sm.V1StopService(args.ArgStartService, reply)
}

func (servManager *ServiceManagerV1) ServiceStatus(args *dispatchers.ArgStartServiceWithOpts, reply *string) (err error) {
	return servManager.sm.V1ServiceStatus(args.ArgStartService, reply)
}

// Ping return pong if the service is active
func (servManager *ServiceManagerV1) Ping(ign *utils.CGREvent, reply *string) error {
	*reply = utils.Pong
	return nil
}

// Call implements rpcclient.ClientConnector interface for internal RPC
func (servManager *ServiceManagerV1) Call(serviceMethod string,
	args interface{}, reply interface{}) error {
	return utils.APIerRPCCall(servManager, serviceMethod, args, reply)
}
