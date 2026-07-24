// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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

func (servManager *ServiceManagerV1) StartService(args dispatchers.ArgStartServiceWithApiKey, reply *string) (err error) {
	return servManager.sm.V1StartService(args.ArgStartService, reply)
}

func (servManager *ServiceManagerV1) StopService(args dispatchers.ArgStartServiceWithApiKey, reply *string) (err error) {
	return servManager.sm.V1StopService(args.ArgStartService, reply)
}

func (servManager *ServiceManagerV1) ServiceStatus(args dispatchers.ArgStartServiceWithApiKey, reply *string) (err error) {
	return servManager.sm.V1ServiceStatus(args.ArgStartService, reply)
}

// Ping return pong if the service is active
func (servManager *ServiceManagerV1) Ping(ign *utils.CGREventWithArgDispatcher, reply *string) error {
	*reply = utils.Pong
	return nil
}

// Call implements birpc.ClientConnector interface for internal RPC
func (servManager *ServiceManagerV1) Call(serviceMethod string,
	args any, reply any) error {
	return utils.APIerRPCCall(servManager, serviceMethod, args, reply)
}
