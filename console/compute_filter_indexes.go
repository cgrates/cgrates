// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdComputeFilterIndexes{
		name:      "compute_filter_indexes",
		rpcMethod: "APIerSv1.ComputeFilterIndexes",
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

type CmdComputeFilterIndexes struct {
	name      string
	rpcMethod string
	rpcParams *utils.ArgsComputeFilterIndexes
	*CommandExecuter
}

func (self *CmdComputeFilterIndexes) Name() string {
	return self.name
}

func (self *CmdComputeFilterIndexes) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdComputeFilterIndexes) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &utils.ArgsComputeFilterIndexes{}
	}
	return self.rpcParams
}

func (self *CmdComputeFilterIndexes) PostprocessRpcParams() error {
	return nil
}

func (self *CmdComputeFilterIndexes) RpcResult() any {
	var reply string
	return &reply
}
