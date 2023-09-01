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
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/utils"
)

func NewCoreSv1(cS *cores.CoreService) *CoreSv1 {
	return &CoreSv1{cS: cS}
}

// CoreSv1 exports RPC from RLs
type CoreSv1 struct {
	cS *cores.CoreService
}

// Call implements birpc.ClientConnector interface for internal RPC
func (cS *CoreSv1) Call(ctx *context.Context, serviceMethod string,
	args any, reply any) error {
	return utils.APIerRPCCall(cS, serviceMethod, args, reply)
}

func (cS *CoreSv1) Status(ctx *context.Context, arg *utils.TenantWithAPIOpts, reply *map[string]any) error {
	return cS.cS.V1Status(ctx, arg, reply)
}

// Ping used to determinate if component is active
func (cS *CoreSv1) Ping(ctx *context.Context, ign *utils.CGREvent, reply *string) error {
	*reply = utils.Pong
	return nil
}

// Sleep is used to test the concurrent requests mechanism
func (cS *CoreSv1) Sleep(ctx *context.Context, args *utils.DurationArgs, reply *string) error {
	return cS.cS.V1Sleep(ctx, args, reply)
}

// StartCPUProfiling is used to start CPUProfiling in the given path
func (cS *CoreSv1) StartCPUProfiling(ctx *context.Context, args *utils.DirectoryArgs, reply *string) error {
	return cS.cS.V1StartCPUProfiling(ctx, args, reply)
}

// StopCPUProfiling is used to stop CPUProfiling. The file should be written on the path
// where the CPUProfiling already started
func (cS *CoreSv1) StopCPUProfiling(ctx *context.Context, args *utils.TenantWithAPIOpts, reply *string) error {
	return cS.cS.V1StopCPUProfiling(ctx, args, reply)
}

// StartMemoryProfiling is used to start MemoryProfiling in the given path
func (cS *CoreSv1) StartMemoryProfiling(ctx *context.Context, args *utils.MemoryPrf, reply *string) error {
	return cS.cS.V1StartMemoryProfiling(ctx, args, reply)
}

// StopMemoryProfiling is used to stop MemoryProfiling. The file should be written on the path
// where the MemoryProfiling already started
func (cS *CoreSv1) StopMemoryProfiling(ctx *context.Context, args *utils.TenantWithAPIOpts, reply *string) error {
	return cS.cS.V1StopMemoryProfiling(ctx, args, reply)

}

func (cS *CoreSv1) Panic(ctx *context.Context, args *utils.PanicMessageArgs, reply *string) error {
	return cS.cS.V1Panic(ctx, args, reply)
}
