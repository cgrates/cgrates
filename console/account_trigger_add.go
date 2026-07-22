// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdAccountAddTriggers{
		name:      "account_triggers_add",
		rpcMethod: utils.APIerSv1AddAccountActionTriggers,
		rpcParams: &v1.AttrAddAccountActionTriggers{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdAccountAddTriggers struct {
	name      string
	rpcMethod string
	rpcParams *v1.AttrAddAccountActionTriggers
	*CommandExecuter
}

func (self *CmdAccountAddTriggers) Name() string {
	return self.name
}

func (self *CmdAccountAddTriggers) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdAccountAddTriggers) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &v1.AttrAddAccountActionTriggers{}
	}
	return self.rpcParams
}

func (self *CmdAccountAddTriggers) PostprocessRpcParams() error {
	return nil
}

func (self *CmdAccountAddTriggers) RpcResult() any {
	var s string
	return &s
}
