// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdGetChargersForEvent{
		name:      "chargers_for_event",
		rpcMethod: utils.ChargerSv1GetChargersForEvent,
		rpcParams: &utils.CGREvent{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

type CmdGetChargersForEvent struct {
	name      string
	rpcMethod string
	rpcParams *utils.CGREvent
	*CommandExecuter
}

func (self *CmdGetChargersForEvent) Name() string {
	return self.name
}

func (self *CmdGetChargersForEvent) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetChargersForEvent) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = new(utils.CGREvent)
	}
	return self.rpcParams
}

func (self *CmdGetChargersForEvent) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetChargersForEvent) RpcResult() any {
	var atr engine.ChargerProfiles
	return &atr
}
