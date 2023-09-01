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
	"github.com/cgrates/cgrates/loaders"
	"github.com/cgrates/cgrates/utils"
)

func NewLoaderSv1(ldrS *loaders.LoaderService) *LoaderSv1 {
	return &LoaderSv1{ldrS: ldrS}
}

// Exports RPC from LoaderService
type LoaderSv1 struct {
	ldrS *loaders.LoaderService
}

// Call implements birpc.ClientConnector interface for internal RPC
func (ldrSv1 *LoaderSv1) Call(ctx *context.Context, serviceMethod string,
	args any, reply any) error {
	return utils.APIerRPCCall(ldrSv1, serviceMethod, args, reply)
}

func (ldrSv1 *LoaderSv1) Load(ctx *context.Context, args *loaders.ArgsProcessFolder,
	rply *string) error {
	return ldrSv1.ldrS.V1Load(ctx, args, rply)
}

func (ldrSv1 *LoaderSv1) Remove(ctx *context.Context, args *loaders.ArgsProcessFolder,
	rply *string) error {
	return ldrSv1.ldrS.V1Remove(ctx, args, rply)
}

func (rsv1 *LoaderSv1) Ping(ctx *context.Context, ign *utils.CGREvent, reply *string) error {
	*reply = utils.Pong
	return nil
}
