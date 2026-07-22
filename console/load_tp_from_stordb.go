// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &LoadTpFromStorDb{
		name:      "load_tp_from_stordb",
		rpcMethod: utils.APIerSv1LoadTariffPlanFromStorDb,
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type LoadTpFromStorDb struct {
	name      string
	rpcMethod string
	rpcParams *v1.AttrLoadTpFromStorDb
	rpcResult string
	*CommandExecuter
}

func (self *LoadTpFromStorDb) Name() string {
	return self.name
}

func (self *LoadTpFromStorDb) RpcMethod() string {
	return self.rpcMethod
}

func (self *LoadTpFromStorDb) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &v1.AttrLoadTpFromStorDb{}
	}
	return self.rpcParams
}

func (self *LoadTpFromStorDb) PostprocessRpcParams() error {
	return nil
}

func (self *LoadTpFromStorDb) RpcResult() any {
	var s string
	return &s
}
