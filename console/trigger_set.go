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

package console

import (
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdSetTriggers{
		name:      "triggers_set",
		rpcMethod: utils.APIerSv1SetActionTrigger,
		rpcParams: &engine.AttrSetActionTrigger{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdSetTriggers struct {
	name      string
	rpcMethod string
	rpcParams *engine.AttrSetActionTrigger
	*CommandExecuter
}

func (self *CmdSetTriggers) Name() string {
	return self.name
}

func (self *CmdSetTriggers) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdSetTriggers) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &engine.AttrSetActionTrigger{}
	}
	return self.rpcParams
}

func (self *CmdSetTriggers) PostprocessRpcParams() error {
	return nil
}

func (self *CmdSetTriggers) RpcResult() any {
	var s string
	return &s
}
