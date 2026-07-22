// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdAccountRemoveTriggers{
		name:      "account_triggers_remove",
		rpcMethod: utils.APIerSv1RemoveAccountActionTriggers,
		rpcParams: &v1.AttrRemoveAccountActionTriggers{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdAccountRemoveTriggers struct {
	name      string
	rpcMethod string
	rpcParams *v1.AttrRemoveAccountActionTriggers
	*CommandExecuter
}

func (self *CmdAccountRemoveTriggers) Name() string {
	return self.name
}

func (self *CmdAccountRemoveTriggers) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdAccountRemoveTriggers) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &v1.AttrRemoveAccountActionTriggers{}
	}
	return self.rpcParams
}

func (self *CmdAccountRemoveTriggers) PostprocessRpcParams() error {
	return nil
}

func (self *CmdAccountRemoveTriggers) RpcResult() any {
	var s string
	return &s
}
