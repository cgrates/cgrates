// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdSetTriggers{
		name:      "triggers_set",
		rpcMethod: utils.APIerSv1SetActionTrigger,
		rpcParams: &v1.AttrSetActionTrigger{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdSetTriggers struct {
	name      string
	rpcMethod string
	rpcParams *v1.AttrSetActionTrigger
	*CommandExecuter
}

func (self *CmdSetTriggers) Name() string {
	return self.name
}

func (self *CmdSetTriggers) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdSetTriggers) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &v1.AttrSetActionTrigger{}
	}
	return self.rpcParams
}

func (self *CmdSetTriggers) PostprocessRpcParams() error {
	return nil
}

func (self *CmdSetTriggers) RpcResult() any {
	var s string
	return &s
}
