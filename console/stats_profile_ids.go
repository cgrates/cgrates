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
	c := &CmdGetStatQueueIDs{
		name:      "stats_profile_ids",
		rpcMethod: utils.AdminSv1GetStatQueueProfileIDs,
		rpcParams: &utils.PaginatorWithTenant{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdGetStatQueueIDs struct {
	name      string
	rpcMethod string
	rpcParams *utils.PaginatorWithTenant
	*CommandExecuter
}

func (self *CmdGetStatQueueIDs) Name() string {
	return self.name
}

func (self *CmdGetStatQueueIDs) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetStatQueueIDs) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &utils.PaginatorWithTenant{}
	}
	return self.rpcParams
}

func (self *CmdGetStatQueueIDs) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetStatQueueIDs) RpcResult() interface{} {
	var atr []string
	return &atr
}
