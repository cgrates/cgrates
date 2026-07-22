// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdGetCostDetails{
		name:      "cost_details",
		rpcMethod: utils.APIerSv1GetEventCost,
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdGetCostDetails struct {
	name      string
	rpcMethod string
	rpcParams *utils.AttrGetCallCost
	rpcResult string
	*CommandExecuter
}

func (self *CmdGetCostDetails) Name() string {
	return self.name
}

func (self *CmdGetCostDetails) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetCostDetails) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &utils.AttrGetCallCost{RunId: utils.MetaDefault}
	}
	return self.rpcParams
}

func (self *CmdGetCostDetails) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetCostDetails) RpcResult() any {
	return &engine.EventCost{}
}

func (self *CmdGetCostDetails) GetFormatedResult(result any) string {
	return GetFormatedResult(result, utils.StringSet{
		utils.Usage:              {},
		utils.GroupIntervalStart: {},
		utils.RateIncrement:      {},
		utils.RateUnit:           {},
	})
}
