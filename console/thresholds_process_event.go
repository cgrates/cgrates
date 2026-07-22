// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"time"

	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdThresholdProcessEvent{
		name:      "thresholds_process_event",
		rpcMethod: utils.ThresholdSv1ProcessEvent,
		rpcParams: &utils.CGREvent{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

type CmdThresholdProcessEvent struct {
	name      string
	rpcMethod string
	rpcParams *utils.CGREvent
	*CommandExecuter
}

func (self *CmdThresholdProcessEvent) Name() string {
	return self.name
}

func (self *CmdThresholdProcessEvent) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdThresholdProcessEvent) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = new(utils.CGREvent)
	}
	return self.rpcParams
}

func (self *CmdThresholdProcessEvent) PostprocessRpcParams() error {
	if self.rpcParams != nil && self.rpcParams.Time == nil {
		self.rpcParams.Time = utils.TimePointer(time.Now())
	}
	return nil
}

func (self *CmdThresholdProcessEvent) RpcResult() any {
	var ids []string
	return &ids
}
