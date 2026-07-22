// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdSetThreshold{
		name:      "thresholds_profile_set",
		rpcMethod: utils.APIerSv1SetThresholdProfile,
		rpcParams: &engine.ThresholdProfileWithAPIOpts{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdSetThreshold struct {
	name      string
	rpcMethod string
	rpcParams *engine.ThresholdProfileWithAPIOpts
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
		self.rpcParams = &engine.ThresholdProfileWithAPIOpts{
			ThresholdProfile: new(engine.ThresholdProfile),
			APIOpts:          map[string]any{},
		}
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
