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
	c := &CmdAttributesProcessEvent{
		name:      "get_attribute_for_event",
		rpcMethod: "AttributeSv1.GetAttributeForEvent",
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdAttributesProcessEvent struct {
	name      string
	rpcMethod string
	rpcParams interface{}
	*CommandExecuter
}

func (self *CmdAttributesProcessEvent) Name() string {
	return self.name
}

func (self *CmdAttributesProcessEvent) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdAttributesProcessEvent) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		mp := make(map[string]interface{})
		self.rpcParams = &mp
	}
	return self.rpcParams
}

func (self *CmdAttributesProcessEvent) PostprocessRpcParams() error { //utils.CGREvent
	param := self.rpcParams.(*map[string]interface{})
	cgrev := utils.CGREvent{
		Tenant: config.CgrConfig().DefaultTenant,
		ID:     utils.UUIDSha1Prefix(),
		Event:  *param,
	}
	if (*param)[utils.Tenant] != nil && (*param)[utils.Tenant].(string) != "" {
		cgrev.Tenant = (*param)[utils.Tenant].(string)
	}
	self.rpcParams = cgrev
	return nil
}

func (self *CmdAttributesProcessEvent) RpcResult() interface{} {
	atr := engine.ExternalAttributeProfile{}
	return &atr
}
