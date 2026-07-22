// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdComputeActionPlanIndexes{
		name:      "compute_actionplan_indexes",
		rpcMethod: utils.APIerSv1ComputeActionPlanIndexes,
		rpcParams: new(EmptyWrapper),
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdComputeActionPlanIndexes struct {
	name      string
	rpcMethod string
	rpcParams *EmptyWrapper
	*CommandExecuter
}

func (self *CmdComputeActionPlanIndexes) Name() string {
	return self.name
}

func (self *CmdComputeActionPlanIndexes) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdComputeActionPlanIndexes) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = new(EmptyWrapper)
	}
	return self.rpcParams
}

func (self *CmdComputeActionPlanIndexes) PostprocessRpcParams() error {
	return nil
}

func (self *CmdComputeActionPlanIndexes) RpcResult() any {
	s := ""
	return &s
}
