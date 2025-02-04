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
	ping
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
