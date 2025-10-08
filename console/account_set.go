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
	c := &CmdSetAccount{
		name:      "account_set",
		rpcMethod: utils.AdminSv1SetAccount,
		rpcParams: &utils.AccountWithAPIOpts{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

type CmdSetAccount struct {
	name      string
	rpcMethod string
	rpcParams *utils.AccountWithAPIOpts
	*CommandExecuter
}

func (self *CmdSetAccount) Name() string {
	return self.name
}

func (self *CmdSetAccount) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdSetAccount) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &utils.AccountWithAPIOpts{Account: new(utils.Account)}
	}
	return self.rpcParams
}

func (self *CmdSetAccount) PostprocessRpcParams() error {
	return nil
}

func (self *CmdSetAccount) RpcResult() any {
	var s string
	return &s
}
