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
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdStatQueueProcessEvent{
		name:      "stats_process_event",
		rpcMethod: "StatSv1.ProcessEvent",
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdStatQueueProcessEvent struct {
	name      string
	rpcMethod string
	rpcParams interface{}
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
		mp := make(map[string]interface{})
		self.rpcParams = &mp
	}
	return self.rpcParams
}

func (self *CmdStatQueueProcessEvent) PostprocessRpcParams() error { //utils.CGREvent
	var tenant string
	param := self.rpcParams.(*map[string]interface{})
	if (*param)[utils.Tenant] != nil && (*param)[utils.Tenant].(string) != "" {
		tenant = (*param)[utils.Tenant].(string)
		delete((*param), utils.Tenant)
	} else {
		tenant = config.CgrConfig().DefaultTenant
	}
	cgrev := utils.CGREvent{
		Tenant: tenant,
		ID:     utils.UUIDSha1Prefix(),
		Event:  *param,
	}
	self.rpcParams = cgrev
	return nil
}

func (self *CmdStatQueueProcessEvent) RpcResult() interface{} {
	var s string
	return &s
}
