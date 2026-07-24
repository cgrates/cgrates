// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdExecuteAction{
		name:      "action_execute",
		rpcMethod: utils.APIerSv1ExecuteAction,
		rpcParams: &utils.AttrExecuteAction{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdExecuteAction struct {
	name      string
	rpcMethod string
	rpcParams *utils.AttrExecuteAction
	*CommandExecuter
}

func (self *CmdExecuteAction) Name() string {
	return self.name
}

func (self *CmdExecuteAction) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdExecuteAction) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &utils.AttrExecuteAction{}
	}
	return self.rpcParams
}

func (self *CmdExecuteAction) PostprocessRpcParams() error {
	return nil
}

func (self *CmdExecuteAction) RpcResult() any {
	var s string
	return &s
}
