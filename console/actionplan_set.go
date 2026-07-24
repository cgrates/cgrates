// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdSetActionPlan{
		name:      "actionplan_set",
		rpcMethod: utils.APIerSv1SetActionPlan,
		rpcParams: &v1.AttrSetActionPlan{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdSetActionPlan struct {
	name      string
	rpcMethod string
	rpcParams *v1.AttrSetActionPlan
	*CommandExecuter
}

func (self *CmdSetActionPlan) Name() string {
	return self.name
}

func (self *CmdSetActionPlan) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdSetActionPlan) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &v1.AttrSetActionPlan{}
	}
	return self.rpcParams
}

func (self *CmdSetActionPlan) PostprocessRpcParams() error {
	return nil
}

func (self *CmdSetActionPlan) RpcResult() any {
	var s string
	return &s
}
