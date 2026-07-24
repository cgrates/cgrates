// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdRemoveActionPlan{
		name:      "actionplan_remove",
		rpcMethod: utils.APIerSv1RemoveActionPlan,
		rpcParams: &v1.AttrGetActionPlan{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdRemoveActionPlan struct {
	name      string
	rpcMethod string
	rpcParams *v1.AttrGetActionPlan
	*CommandExecuter
}

func (self *CmdRemoveActionPlan) Name() string {
	return self.name
}

func (self *CmdRemoveActionPlan) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdRemoveActionPlan) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &v1.AttrGetActionPlan{}
	}
	return self.rpcParams
}

func (self *CmdRemoveActionPlan) PostprocessRpcParams() error {
	return nil
}

func (self *CmdRemoveActionPlan) RpcResult() any {
	var s string
	return &s
}
