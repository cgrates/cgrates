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
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdSetAttributeProfile{
		name:      "attribute_profile_set",
		rpcMethod: utils.AdminSv1SetAttributeProfile,
		rpcParams: &utils.APIAttributeProfileWithAPIOpts{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

type CmdSetAttributeProfile struct {
	name      string
	rpcMethod string
	rpcParams *utils.APIAttributeProfileWithAPIOpts
	*CommandExecuter
}

func (self *CmdSetAttributeProfile) Name() string {
	return self.name
}

func (self *CmdSetAttributeProfile) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdSetAttributeProfile) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &utils.APIAttributeProfileWithAPIOpts{APIAttributeProfile: new(utils.APIAttributeProfile)}
	}
	return self.rpcParams
}

func (self *CmdSetAttributeProfile) PostprocessRpcParams() error {
	return nil
}

func (self *CmdSetAttributeProfile) RpcResult() any {
	var s string
	return &s
}
