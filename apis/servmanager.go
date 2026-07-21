// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package apis

import (
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/servmanager"
)

func NewServiceManagerV1(sm *servmanager.ServiceManager) *ServiceManagerV1 {
	return &ServiceManagerV1{sm: sm}
}

type ServiceManagerV1 struct {
	sm *servmanager.ServiceManager
}

func (servManager *ServiceManagerV1) StartService(ctx *context.Context, args *servmanager.ArgsServiceID, reply *string) error {
	return servManager.sm.V1StartService(ctx, args, reply)
}

func (servManager *ServiceManagerV1) StopService(ctx *context.Context, args *servmanager.ArgsServiceID, reply *string) error {
	return servManager.sm.V1StopService(ctx, args, reply)
}

func (servManager *ServiceManagerV1) ServiceStatus(ctx *context.Context, args *servmanager.ArgsServiceID, reply *map[string]string) error {
	return servManager.sm.V1ServiceStatus(ctx, args, reply)
}
