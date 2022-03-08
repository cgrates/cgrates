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
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdRelaodConfigSection{
		name:      "reload_config",
		rpcMethod: utils.ConfigSv1ReloadConfig,
		rpcParams: &config.ReloadArgs{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdRelaodConfigSection struct {
	name      string
	rpcMethod string
	rpcParams *config.ReloadArgs
	*CommandExecuter
}

func (self *CmdRelaodConfigSection) Name() string {
	return self.name
}

func (self *CmdRelaodConfigSection) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdRelaodConfigSection) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &config.ReloadArgs{APIOpts: make(map[string]interface{})}
	}
	return self.rpcParams
}

func (self *CmdRelaodConfigSection) PostprocessRpcParams() error {
	return nil
}

func (self *CmdRelaodConfigSection) RpcResult() interface{} {
	var s string
	return &s
}
