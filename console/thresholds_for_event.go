// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

//

func init() {
	c := &CmdThresholdsForEvent{
		name:      "thresholds_for_event",
		rpcMethod: utils.ThresholdSv1GetThresholdsForEvent,
		rpcParams: &engine.ArgsProcessEvent{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

type CmdThresholdsForEvent struct {
	name      string
	rpcMethod string
	rpcParams *engine.ArgsProcessEvent
	*CommandExecuter
}

func (self *CmdThresholdsForEvent) Name() string {
	return self.name
}

func (self *CmdThresholdsForEvent) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdThresholdsForEvent) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &engine.ArgsProcessEvent{ArgDispatcher: new(utils.ArgDispatcher)}
	}
	return self.rpcParams
}

func (self *CmdThresholdsForEvent) PostprocessRpcParams() error {
	return nil
}

func (self *CmdThresholdsForEvent) RpcResult() any {
	var s *engine.Thresholds
	return &s
}
