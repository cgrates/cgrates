// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package v1

import (
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/dispatchers"
	"github.com/cgrates/cgrates/servmanager"
)

func NewServiceManagerV1(sm *servmanager.ServiceManager) *ServiceManagerV1 {
	return &ServiceManagerV1{sm: sm}
}

type ServiceManagerV1 struct {
	sm *servmanager.ServiceManager // Need to have them capitalize so we can export in V2
}

func (servManager *ServiceManagerV1) StartService(ctx *context.Context, args *dispatchers.ArgStartServiceWithAPIOpts, reply *string) (err error) {
	return servManager.sm.V1StartService(ctx, args.ArgStartService, reply)
}

func (servManager *ServiceManagerV1) StopService(ctx *context.Context, args *dispatchers.ArgStartServiceWithAPIOpts, reply *string) (err error) {
	return servManager.sm.V1StopService(ctx, args.ArgStartService, reply)
}

func (servManager *ServiceManagerV1) ServiceStatus(ctx *context.Context, args *dispatchers.ArgStartServiceWithAPIOpts, reply *string) (err error) {
	return servManager.sm.V1ServiceStatus(ctx, args.ArgStartService, reply)
}
