// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdAddBalance{
		name:      "balance_add",
		rpcMethod: utils.APIerSv1AddBalance,
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdAddBalance struct {
	name      string
	rpcMethod string
	rpcParams *v1.AttrAddBalance
	*CommandExecuter
}

func (self *CmdAddBalance) Name() string {
	return self.name
}

func (self *CmdAddBalance) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdAddBalance) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &v1.AttrAddBalance{BalanceType: utils.MetaMonetary, Overwrite: false}
	}
	return self.rpcParams
}

func (self *CmdAddBalance) PostprocessRpcParams() error {
	return nil
}

func (self *CmdAddBalance) RpcResult() any {
	var s string
	return &s
}
