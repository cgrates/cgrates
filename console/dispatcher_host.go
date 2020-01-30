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
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdGetDispatcherHost{
		name:      "dispatcher_host",
		rpcMethod: utils.APIerSv1GetDispatcherHost,
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdGetDispatcherHost struct {
	name      string
	rpcMethod string
	rpcParams *utils.TenantID
	*CommandExecuter
}

func (self *CmdGetDispatcherHost) Name() string {
	return self.name
}

func (self *CmdGetDispatcherHost) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetDispatcherHost) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = new(utils.TenantID)
	}
	return self.rpcParams
}

func (self *CmdGetDispatcherHost) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetDispatcherHost) RpcResult() interface{} {
	var s engine.DispatcherHost
	return &s
}
