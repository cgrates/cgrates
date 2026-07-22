// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdSetBalance{
		name:      "balance_set",
		rpcMethod: utils.APIerSv1SetBalance,
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdSetBalance struct {
	name      string
	rpcMethod string
	rpcParams *utils.AttrSetBalance
	*CommandExecuter
}

func (self *CmdSetBalance) Name() string {
	return self.name
}

func (self *CmdSetBalance) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdSetBalance) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &utils.AttrSetBalance{BalanceType: utils.MetaMonetary}
	}
	return self.rpcParams
}

func (self *CmdSetBalance) PostprocessRpcParams() error {
	return nil
}

func (self *CmdSetBalance) RpcResult() any {
	var s string
	return &s
}
