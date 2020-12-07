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
	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdGetDataCost{
		name:       "datacost",
		rpcMethod:  utils.APIerSv1GetDataCost,
		clientArgs: []string{utils.Category, utils.Tenant, utils.Account, utils.Subject, utils.StartTime, utils.Usage},
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

func (self *CmdGetDataCost) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &v1.AttrGetDataCost{Opts: make(map[string]interface{})}
	}
	return self.rpcParams
}

func (self *CmdGetDataCost) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetDataCost) RpcResult() interface{} {
	return &engine.DataCost{}
}

func (self *CmdGetDataCost) ClientArgs() []string {
	return self.clientArgs
}

func (self *CmdGetDataCost) GetFormatedResult(result interface{}) string {
	return GetFormatedResult(result, utils.StringSet{
		utils.Usage:              {},
		utils.GroupIntervalStart: {},
		utils.RateIncrement:      {},
		utils.RateUnit:           {},
	})
}
