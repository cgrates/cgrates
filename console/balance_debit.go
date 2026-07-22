// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdBalanceDebit{
		name:      "balance_debit",
		rpcMethod: utils.APIerSv1DebitBalance,
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdBalanceDebit struct {
	name       string
	rpcMethod  string
	rpcParams  *v1.AttrAddBalance
	clientArgs []string
	*CommandExecuter
}

func (self *CmdBalanceDebit) Name() string {
	return self.name
}

func (self *CmdBalanceDebit) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdBalanceDebit) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &v1.AttrAddBalance{}
	}
	return self.rpcParams
}

func (self *CmdBalanceDebit) PostprocessRpcParams() error {
	return nil
}

func (self *CmdBalanceDebit) RpcResult() any {
	var s string
	return &s
}
