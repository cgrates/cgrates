// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdGetStatQueueProfile{
		name:      "stats_profile",
		rpcMethod: utils.APIerSv1GetStatQueueProfile,
		rpcParams: &utils.TenantID{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdGetStatQueueProfile struct {
	name      string
	rpcMethod string
	rpcParams *utils.TenantID
	*CommandExecuter
}

func (self *CmdGetStatQueueProfile) Name() string {
	return self.name
}

func (self *CmdGetStatQueueProfile) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetStatQueueProfile) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &utils.TenantID{}
	}
	return self.rpcParams
}

func (self *CmdGetStatQueueProfile) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetStatQueueProfile) RpcResult() any {
	var atr engine.StatQueueProfile
	return &atr
}

func (self *CmdGetStatQueueProfile) GetFormatedResult(result any) string {
	return GetFormatedResult(result, utils.StringSet{
		utils.TTL: {},
	})
}
