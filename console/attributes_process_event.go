// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdAttributesProcessEvent{
		name:      "attributes_process_event",
		rpcMethod: utils.AttributeSv1ProcessEvent,
		rpcParams: &utils.CGREvent{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

type CmdAttributesProcessEvent struct {
	name      string
	rpcMethod string
	rpcParams *utils.CGREvent
	*CommandExecuter
}

func (self *CmdAttributesProcessEvent) Name() string {
	return self.name
}

func (self *CmdAttributesProcessEvent) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdAttributesProcessEvent) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = new(utils.CGREvent)
	}
	return self.rpcParams
}

func (self *CmdAttributesProcessEvent) PostprocessRpcParams() error {
	if self.rpcParams != nil && self.rpcParams.Time == nil {
		self.rpcParams.Time = utils.TimePointer(time.Now())
	}
	return nil
}

func (self *CmdAttributesProcessEvent) RpcResult() any {
	var atr engine.AttrSProcessEventReply
	return &atr
}

func (self *CmdAttributesProcessEvent) GetFormatedResult(result any) string {
	return GetFormatedResult(result, utils.StringSet{
		utils.Usage: {},
	})
}
