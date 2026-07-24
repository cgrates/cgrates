// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdAccountResetTriggers{
		name:      "account_triggers_reset",
		rpcMethod: utils.APIerSv1ResetAccountActionTriggers,
		rpcParams: &v1.AttrRemoveAccountActionTriggers{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdAccountResetTriggers struct {
	name      string
	rpcMethod string
	rpcParams *v1.AttrRemoveAccountActionTriggers
	*CommandExecuter
}

func (self *CmdAccountResetTriggers) Name() string {
	return self.name
}

func (self *CmdAccountResetTriggers) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdAccountResetTriggers) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &v1.AttrRemoveAccountActionTriggers{}
	}
	return self.rpcParams
}

func (self *CmdAccountResetTriggers) PostprocessRpcParams() error {
	return nil
}

func (self *CmdAccountResetTriggers) RpcResult() any {
	var s string
	return &s
}
