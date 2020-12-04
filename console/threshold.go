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
	c := &CmdGetThreshold{
		name:      "threshold",
		rpcMethod: utils.ThresholdSv1GetThreshold,
		rpcParams: &utils.TenantIDWithOpts{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

type CmdGetThreshold struct {
	name      string
	rpcMethod string
	rpcParams *utils.TenantIDWithOpts
	*CommandExecuter
}

func (self *CmdGetThreshold) Name() string {
	return self.name
}

func (self *CmdGetThreshold) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetThreshold) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &utils.TenantIDWithOpts{
			TenantID: new(utils.TenantID),
			Opts:     make(map[string]interface{}),
		}
	}
	return self.rpcParams
}

func (self *CmdGetThreshold) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetThreshold) RpcResult() interface{} {
	var atr engine.Threshold
	return &atr
}

func (self *CmdGetThreshold) GetFormatedResult(result interface{}) string {
	return GetFormatedResult(result, utils.StringSet{
		"MinSleep": {},
	})
}
