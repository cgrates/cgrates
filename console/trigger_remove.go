// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdRemoveTriggers{
		name:      "triggers_remove",
		rpcMethod: utils.APIerSv1RemoveActionTrigger,
		rpcParams: &v1.AttrRemoveActionTrigger{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdRemoveTriggers struct {
	name      string
	rpcMethod string
	rpcParams *v1.AttrRemoveActionTrigger
	*CommandExecuter
}

func (self *CmdRemoveTriggers) Name() string {
	return self.name
}

func (self *CmdRemoveTriggers) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdRemoveTriggers) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &v1.AttrRemoveActionTrigger{}
	}
	return self.rpcParams
}

func (self *CmdRemoveTriggers) PostprocessRpcParams() error {
	return nil
}

func (self *CmdRemoveTriggers) RpcResult() any {
	var s string
	return &s
}
