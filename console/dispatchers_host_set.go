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
	c := &CmdSetDispatcherHost{
		name:      "dispatchers_host_set",
		rpcMethod: utils.APIerSv1SetDispatcherHost,
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdSetDispatcherHost struct {
	name      string
	rpcMethod string
	rpcParams *engine.DispatcherHostWithOpts
	*CommandExecuter
}

func (self *CmdSetDispatcherHost) Name() string {
	return self.name
}

func (self *CmdSetDispatcherHost) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdSetDispatcherHost) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &engine.DispatcherHostWithOpts{
			DispatcherHost: new(engine.DispatcherHost),
			Opts:           make(map[string]interface{}),
		}
	}
	return self.rpcParams
}

func (self *CmdSetDispatcherHost) PostprocessRpcParams() error {
	return nil
}

func (self *CmdSetDispatcherHost) RpcResult() interface{} {
	var s string
	return &s
}
