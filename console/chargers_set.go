// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdSetChargers{
		name:      "chargers_set",
		rpcMethod: utils.APIerSv1SetChargerProfile,
		rpcParams: &v1.ChargerWithCache{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

type CmdSetChargers struct {
	name      string
	rpcMethod string
	rpcParams *v1.ChargerWithCache
	*CommandExecuter
}

func (self *CmdSetChargers) Name() string {
	return self.name
}

func (self *CmdSetChargers) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdSetChargers) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &v1.ChargerWithCache{ChargerProfile: new(engine.ChargerProfile)}
	}
	return self.rpcParams
}

func (self *CmdSetChargers) PostprocessRpcParams() error {
	return nil
}

func (self *CmdSetChargers) RpcResult() any {
	var s string
	return &s
}
