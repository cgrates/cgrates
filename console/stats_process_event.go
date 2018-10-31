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
	"time"

	"github.com/cgrates/cgrates/dispatchers"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdStatQueueProcessEvent{
		name:      "stats_process_event",
		rpcMethod: utils.StatSv1ProcessEvent,
		rpcParams: &dispatchers.ArgsStatProcessEventWithApiKey{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdStatQueueProcessEvent struct {
	name      string
	rpcMethod string
	rpcParams *dispatchers.ArgsStatProcessEventWithApiKey
	*CommandExecuter
}

func (self *CmdStatQueueProcessEvent) Name() string {
	return self.name
}

func (self *CmdStatQueueProcessEvent) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdStatQueueProcessEvent) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &dispatchers.ArgsStatProcessEventWithApiKey{}
	}
	return self.rpcParams
}

func (self *CmdStatQueueProcessEvent) PostprocessRpcParams() error {
	if self.rpcParams.Time == nil {
		self.rpcParams.Time = utils.TimePointer(time.Now())
	}
	return nil
}

func (self *CmdStatQueueProcessEvent) RpcResult() interface{} {
	var atr []string
	return &atr
}
