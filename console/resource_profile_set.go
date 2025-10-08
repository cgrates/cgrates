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
	c := &CmdSetResourceProfile{
		name:      "resource_profile_set",
		rpcMethod: utils.AdminSv1SetResourceProfile,
		rpcParams: &utils.ResourceProfileWithAPIOpts{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdSetResourceProfile struct {
	name      string
	rpcMethod string
	rpcParams *utils.ResourceProfileWithAPIOpts
	*CommandExecuter
}

func (self *CmdSetResourceProfile) Name() string {
	return self.name
}

func (self *CmdSetResourceProfile) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdSetResourceProfile) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &utils.ResourceProfileWithAPIOpts{
			ResourceProfile: new(utils.ResourceProfile),
			APIOpts:         make(map[string]any),
		}
	}
	return self.rpcParams
}

func (self *CmdSetResourceProfile) PostprocessRpcParams() error {
	return nil
}

func (self *CmdSetResourceProfile) RpcResult() any {
	var s string
	return &s
}
