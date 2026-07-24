// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdGetDataCost{
		name:       "datacost",
		rpcMethod:  utils.APIerSv1GetDataCost,
		clientArgs: []string{"Category", "Tenant", "Account", "Subject", "StartTime", "Usage"},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdGetDataCost struct {
	name       string
	rpcMethod  string
	rpcParams  *v1.AttrGetDataCost
	clientArgs []string
	*CommandExecuter
}

func (self *CmdGetDataCost) Name() string {
	return self.name
}

func (self *CmdGetDataCost) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetDataCost) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &v1.AttrGetDataCost{ArgDispatcher: new(utils.ArgDispatcher)}
	}
	return self.rpcParams
}

func (self *CmdGetDataCost) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetDataCost) RpcResult() any {
	return &engine.DataCost{}
}

func (self *CmdGetDataCost) ClientArgs() []string {
	return self.clientArgs
}

func (self *CmdGetDataCost) GetFormatedResult(result any) string {
	return GetFormatedResult(result, map[string]struct{}{
		"Usage":              {},
		"GroupIntervalStart": {},
		"RateIncrement":      {},
		"RateUnit":           {},
	})
}
