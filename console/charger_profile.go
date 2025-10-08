/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package console

import (
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdGetChargerProfile{
		name:      "charger_profile",
		rpcMethod: utils.AdminSv1GetChargerProfile,
		rpcParams: &utils.TenantIDWithAPIOpts{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdGetChargerProfile struct {
	name      string
	rpcMethod string
	rpcParams *utils.TenantIDWithAPIOpts
	*CommandExecuter
}

func (self *CmdGetChargerProfile) Name() string {
	return self.name
}

func (self *CmdGetChargerProfile) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetChargerProfile) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &utils.TenantIDWithAPIOpts{}
	}
	return self.rpcParams
}

func (self *CmdGetChargerProfile) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetChargerProfile) RpcResult() any {
	var atr utils.ChargerProfile
	return &atr
}
