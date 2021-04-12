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
	"time"

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
	args, reply interface{}) error {
	return utils.APIerRPCCall(cS, serviceMethod, args, reply)
}

func (cS *CoreSv1) Status(arg *utils.TenantWithAPIOpts, reply *map[string]interface{}) error {
	return cS.cS.Status(arg, reply)
}

// Ping used to determinate if component is active
func (cS *CoreSv1) Ping(ign *utils.CGREvent, reply *string) error {
	*reply = utils.Pong
	return nil
}

// Sleep is used to test the concurrent requests mechanism
func (cS *CoreSv1) Sleep(arg *utils.DurationArgs, reply *string) error {
	time.Sleep(arg.Duration)
	*reply = utils.OK
	return nil
}
