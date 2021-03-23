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
	v2 "github.com/cgrates/cgrates/apier/v2"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdSetAttributes{
		name:      "attributes_profile_set",
		rpcMethod: utils.APIerSv2SetAttributeProfile,
		rpcParams: &v2.AttributeWithAPIOpts{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

type CmdSetAttributes struct {
	name      string
	rpcMethod string
	rpcParams *v2.AttributeWithAPIOpts
	*CommandExecuter
}

func (self *CmdSetAttributes) Name() string {
	return self.name
}

func (self *CmdSetAttributes) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdSetAttributes) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &v2.AttributeWithAPIOpts{APIAttributeProfile: new(engine.APIAttributeProfile)}
	}
	return self.rpcParams
}

func (self *CmdSetAttributes) PostprocessRpcParams() error {
	return nil
}

func (self *CmdSetAttributes) RpcResult() interface{} {
	var s string
	return &s
}
