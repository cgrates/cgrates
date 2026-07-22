// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdGetActionPlan{
		name:      "actionplan_get",
		rpcMethod: utils.APIerSv1GetActionPlan,
		rpcParams: &v1.AttrGetActionPlan{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdGetActionPlan struct {
	name      string
	rpcMethod string
	rpcParams *v1.AttrGetActionPlan
	*CommandExecuter
}

func (self *CmdGetActionPlan) Name() string {
	return self.name
}

func (self *CmdGetActionPlan) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetActionPlan) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &v1.AttrGetActionPlan{}
	}
	return self.rpcParams
}

func (self *CmdGetActionPlan) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetActionPlan) RpcResult() any {
	s := make([]*engine.ActionPlan, 0)
	return &s
}
