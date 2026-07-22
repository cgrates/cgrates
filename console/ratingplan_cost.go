// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"github.com/cgrates/cgrates/dispatchers"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdGetRatingPlanCost{
		name:      "ratingplan_cost",
		rpcMethod: utils.RALsV1GetRatingPlansCost,
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdGetRatingPlanCost struct {
	name      string
	rpcMethod string
	rpcParams *utils.RatingPlanCostArg
	rpcResult string
	*CommandExecuter
}

func (self *CmdGetRatingPlanCost) Name() string {
	return self.name
}

func (self *CmdGetRatingPlanCost) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetRatingPlanCost) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = new(utils.RatingPlanCostArg)
	}
	return self.rpcParams
}

func (self *CmdGetRatingPlanCost) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetRatingPlanCost) RpcResult() any {
	var s dispatchers.RatingPlanCost
	return &s
}
