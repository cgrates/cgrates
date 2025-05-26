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

package console

import (
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdResetStatQueue{
		name:      "stats_reset_queue",
		rpcMethod: utils.StatSv1ResetStatQueue,
		rpcParams: &utils.TenantIDWithAPIOpts{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

type CmdResetStatQueue struct {
	name      string
	rpcMethod string
	rpcParams *utils.TenantIDWithAPIOpts
	*CommandExecuter
}

func (self *CmdResetStatQueue) Name() string {
	return self.name
}

func (self *CmdResetStatQueue) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdResetStatQueue) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &utils.TenantIDWithAPIOpts{
			TenantID: new(utils.TenantID),
			APIOpts:  make(map[string]any),
		}
	}
	return self.rpcParams
}

func (self *CmdResetStatQueue) PostprocessRpcParams() error {
	return nil
}

func (self *CmdResetStatQueue) RpcResult() any {
	var s string
	return &s
}
