// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdAccountSetTriggers{
		name:      "account_triggers_set",
		rpcMethod: utils.APIerSv1SetAccountActionTriggers,
		rpcParams: &v1.AttrSetAccountActionTriggers{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdAccountSetTriggers struct {
	name      string
	rpcMethod string
	rpcParams *v1.AttrSetAccountActionTriggers
	*CommandExecuter
}

func (self *CmdAccountSetTriggers) Name() string {
	return self.name
}

func (self *CmdAccountSetTriggers) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdAccountSetTriggers) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &v1.AttrSetAccountActionTriggers{}
	}
	return self.rpcParams
}

func (self *CmdAccountSetTriggers) PostprocessRpcParams() error {
	return nil
}

func (self *CmdAccountSetTriggers) RpcResult() any {
	var s string
	return &s
}
