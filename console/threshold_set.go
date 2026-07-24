// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdSetThreshold{
		name:      "threshold_set",
		rpcMethod: utils.APIerSv1SetThresholdProfile,
		rpcParams: &engine.ThresholdWithCache{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdSetThreshold struct {
	name      string
	rpcMethod string
	rpcParams *engine.ThresholdWithCache
	*CommandExecuter
}

func (self *CmdSetThreshold) Name() string {
	return self.name
}

func (self *CmdSetThreshold) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdSetThreshold) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &engine.ThresholdWithCache{ThresholdProfile: new(engine.ThresholdProfile)}
	}
	return self.rpcParams
}

func (self *CmdSetThreshold) PostprocessRpcParams() error {
	return nil
}

func (self *CmdSetThreshold) RpcResult() any {
	var s string
	return &s
}
