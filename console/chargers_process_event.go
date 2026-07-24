// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdChargersProcessEvent{
		name:      "chargers_process_event",
		rpcMethod: utils.ChargerSv1ProcessEvent,
		rpcParams: &utils.CGREventWithArgDispatcher{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

type CmdChargersProcessEvent struct {
	name      string
	rpcMethod string
	rpcParams *utils.CGREventWithArgDispatcher
	*CommandExecuter
}

func (self *CmdChargersProcessEvent) Name() string {
	return self.name
}

func (self *CmdChargersProcessEvent) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdChargersProcessEvent) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &utils.CGREventWithArgDispatcher{
			CGREvent:      new(utils.CGREvent),
			ArgDispatcher: new(utils.ArgDispatcher),
		}
	}
	return self.rpcParams
}

func (self *CmdChargersProcessEvent) PostprocessRpcParams() error {
	if self.rpcParams != nil && self.rpcParams.CGREvent != nil &&
		self.rpcParams.Time == nil {
		self.rpcParams.Time = utils.TimePointer(time.Now())
	}
	return nil
}

func (self *CmdChargersProcessEvent) RpcResult() any {
	var atr []*engine.ChrgSProcessEventReply
	return &atr
}

func (self *CmdChargersProcessEvent) GetFormatedResult(result any) string {
	return GetFormatedResult(result, map[string]struct{}{
		"Usage": {},
	})
}
