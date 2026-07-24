// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	v2 "github.com/cgrates/cgrates/apier/v2"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdAddAccount{
		name:      "account_set",
		rpcMethod: utils.APIerSv2SetAccount,
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdAddAccount struct {
	name      string
	rpcMethod string
	rpcParams *v2.AttrSetAccount
	*CommandExecuter
}

func (self *CmdAddAccount) Name() string {
	return self.name
}

func (self *CmdAddAccount) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdAddAccount) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &v2.AttrSetAccount{}
	}
	return self.rpcParams
}

func (self *CmdAddAccount) PostprocessRpcParams() error {
	return nil
}

func (self *CmdAddAccount) RpcResult() any {
	var s string
	return &s
}
