// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdGetThresholdProfile{
		name:      "threshold_profile",
		rpcMethod: utils.APIerSv1GetThresholdProfile,
		rpcParams: &utils.TenantIDWithArgDispatcher{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

type CmdGetThresholdProfile struct {
	name      string
	rpcMethod string
	rpcParams *utils.TenantIDWithArgDispatcher
	*CommandExecuter
}

func (self *CmdGetThresholdProfile) Name() string {
	return self.name
}

func (self *CmdGetThresholdProfile) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetThresholdProfile) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &utils.TenantIDWithArgDispatcher{
			TenantID:      new(utils.TenantID),
			ArgDispatcher: new(utils.ArgDispatcher),
		}
	}
	return self.rpcParams
}

func (self *CmdGetThresholdProfile) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetThresholdProfile) RpcResult() any {
	var atr engine.ThresholdProfile
	return &atr
}

func (self *CmdGetThresholdProfile) GetFormatedResult(result any) string {
	return GetFormatedResult(result, map[string]struct{}{
		"MinSleep": {},
	})
}
