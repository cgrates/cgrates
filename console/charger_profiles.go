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
	c := &CmdGetChargerProfiles{
		name:      "charger_profiles",
		rpcMethod: utils.AdminSv1GetChargerProfiles,
		rpcParams: &utils.ArgsItemIDs{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdGetChargerProfiles struct {
	name      string
	rpcMethod string
	rpcParams *utils.ArgsItemIDs
	*CommandExecuter
}

func (self *CmdGetChargerProfiles) Name() string {
	return self.name
}

func (self *CmdGetChargerProfiles) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetChargerProfiles) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &utils.ArgsItemIDs{}
	}
	return self.rpcParams
}

func (self *CmdGetChargerProfiles) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetChargerProfiles) RpcResult() interface{} {
	var atr []*engine.ChargerProfile
	return &atr
}
