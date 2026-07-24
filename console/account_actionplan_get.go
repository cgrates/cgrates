// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdGetAccountActionPlan{
		name:      "account_actionplan_get",
		rpcMethod: utils.APIerSv1GetAccountActionPlan,
		rpcParams: &utils.TenantAccount{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdGetAccountActionPlan struct {
	name      string
	rpcMethod string
	rpcParams *utils.TenantAccount
	*CommandExecuter
}

func (self *CmdGetAccountActionPlan) Name() string {
	return self.name
}

func (self *CmdGetAccountActionPlan) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetAccountActionPlan) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &utils.TenantAccount{}
	}
	return self.rpcParams
}

func (self *CmdGetAccountActionPlan) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetAccountActionPlan) RpcResult() any {
	s := make([]*v1.AccountActionTiming, 0)
	return &s
}
