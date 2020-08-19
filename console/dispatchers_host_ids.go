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
	c := &CmdGetDispatcherHostIDs{
		name:      "dispatchers_host_ids",
		rpcMethod: utils.APIerSv1GetDispatcherHostIDs,
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdGetDispatcherHostIDs struct {
	name      string
	rpcMethod string
	rpcParams *utils.TenantArgWithPaginator
	*CommandExecuter
}

func (self *CmdGetDispatcherHostIDs) Name() string {
	return self.name
}

func (self *CmdGetDispatcherHostIDs) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetDispatcherHostIDs) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = new(utils.TenantArgWithPaginator)
	}
	return self.rpcParams
}

func (self *CmdGetDispatcherHostIDs) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetDispatcherHostIDs) RpcResult() interface{} {
	var s []string
	return &s
}
