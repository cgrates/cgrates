// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdGetThreshold{
		name:      "threshold",
		rpcMethod: utils.ThresholdSv1GetThreshold,
		rpcParams: &utils.TenantIDWithAPIOpts{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

type CmdGetThreshold struct {
	name      string
	rpcMethod string
	rpcParams *utils.TenantIDWithAPIOpts
	*CommandExecuter
}

func (self *CmdGetThreshold) Name() string {
	return self.name
}

func (self *CmdGetThreshold) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetThreshold) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &utils.TenantIDWithAPIOpts{
			TenantID: new(utils.TenantID),
			APIOpts:  make(map[string]any),
		}
	}
	return self.rpcParams
}

func (self *CmdGetThreshold) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetThreshold) RpcResult() any {
	var atr engine.Threshold
	return &atr
}

func (self *CmdGetThreshold) GetFormatedResult(result any) string {
	return GetFormatedResult(result, utils.StringSet{
		utils.MinSleep: {},
	})
}
