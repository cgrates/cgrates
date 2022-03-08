/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package console

import (
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdChargersProcessEvent{
		name:      "chargers_process_event",
		rpcMethod: utils.ChargerSv1ProcessEvent,
		rpcParams: &utils.CGREvent{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

type CmdChargersProcessEvent struct {
	name      string
	rpcMethod string
	rpcParams *utils.CGREvent
	*CommandExecuter
}

func (self *CmdChargersProcessEvent) Name() string {
	return self.name
}

func (self *CmdChargersProcessEvent) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdChargersProcessEvent) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = new(utils.CGREvent)
	}
	return self.rpcParams
}

func (self *CmdChargersProcessEvent) PostprocessRpcParams() error {
	return nil
}

func (self *CmdChargersProcessEvent) RpcResult() interface{} {
	var atr []*engine.ChrgSProcessEventReply
	return &atr
}

func (self *CmdChargersProcessEvent) GetFormatedResult(result interface{}) string {
	return GetFormatedResult(result, utils.StringSet{
		utils.Usage: {},
	})
}
