/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

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
