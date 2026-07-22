// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import "github.com/cgrates/cgrates/utils"

func init() {
	c := &CmdRemoveAccount{
		name:      "account_remove",
		rpcMethod: utils.APIerSv1RemoveAccount,
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdRemoveAccount struct {
	name      string
	rpcMethod string
	rpcParams *utils.AttrRemoveAccount
	*CommandExecuter
}

func (self *CmdRemoveAccount) Name() string {
	return self.name
}

func (self *CmdRemoveAccount) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdRemoveAccount) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &utils.AttrRemoveAccount{}
	}
	return self.rpcParams
}

func (self *CmdRemoveAccount) PostprocessRpcParams() error {
	return nil
}

func (self *CmdRemoveAccount) RpcResult() any {
	var s string
	return &s
}
