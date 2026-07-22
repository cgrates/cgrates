// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdGetChargers{
		name:      "chargers_profile",
		rpcMethod: utils.APIerSv1GetChargerProfile,
		rpcParams: &utils.TenantID{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdGetChargers struct {
	name      string
	rpcMethod string
	rpcParams *utils.TenantID
	*CommandExecuter
}

func (self *CmdGetChargers) Name() string {
	return self.name
}

func (self *CmdGetChargers) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetChargers) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &utils.TenantID{}
	}
	return self.rpcParams
}

func (self *CmdGetChargers) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetChargers) RpcResult() any {
	var atr engine.ChargerProfile
	return &atr
}
