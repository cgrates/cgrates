// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdRemoveBalance{
		name:      "balance_remove",
		rpcMethod: utils.APIerSv1RemoveBalances,
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdRemoveBalance struct {
	name      string
	rpcMethod string
	rpcParams *utils.AttrSetBalance
	*CommandExecuter
}

func (self *CmdRemoveBalance) Name() string {
	return self.name
}

func (self *CmdRemoveBalance) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdRemoveBalance) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &utils.AttrSetBalance{BalanceType: utils.MetaMonetary}
	}
	return self.rpcParams
}

func (self *CmdRemoveBalance) PostprocessRpcParams() error {
	return nil
}

func (self *CmdRemoveBalance) RpcResult() any {
	var s string
	return &s
}
