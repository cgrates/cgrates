// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdGetCost{
		name:       "cost",
		rpcMethod:  utils.APIerSv1GetCost,
		clientArgs: []string{utils.Tenant, utils.Category, utils.Subject, utils.AnswerTime, utils.Destination, utils.Usage},
		rpcParams:  &v1.AttrGetCost{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdGetCost struct {
	name       string
	rpcMethod  string
	rpcParams  *v1.AttrGetCost
	clientArgs []string
	*CommandExecuter
}

func (self *CmdGetCost) Name() string {
	return self.name
}

func (self *CmdGetCost) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetCost) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &v1.AttrGetCost{APIOpts: make(map[string]any)}
	}
	return self.rpcParams
}

func (self *CmdGetCost) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetCost) RpcResult() any {
	return &engine.EventCost{}
}

func (self *CmdGetCost) ClientArgs() []string {
	return self.clientArgs
}

func (self *CmdGetCost) GetFormatedResult(result any) string {
	return GetFormatedResult(result, utils.StringSet{
		utils.Usage:              {},
		utils.GroupIntervalStart: {},
		utils.RateIncrement:      {},
		utils.RateUnit:           {},
	})
}
