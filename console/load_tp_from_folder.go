// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import "github.com/cgrates/cgrates/utils"

func init() {
	c := &LoadTpFromFolder{
		name:      "load_tp_from_folder",
		rpcMethod: utils.APIerSv1LoadTariffPlanFromFolder,
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type LoadTpFromFolder struct {
	name      string
	rpcMethod string
	rpcParams *utils.AttrLoadTpFromFolder
	rpcResult string
	*CommandExecuter
}

func (self *LoadTpFromFolder) Name() string {
	return self.name
}

func (self *LoadTpFromFolder) RpcMethod() string {
	return self.rpcMethod
}

func (self *LoadTpFromFolder) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &utils.AttrLoadTpFromFolder{}
	}
	return self.rpcParams
}

func (self *LoadTpFromFolder) PostprocessRpcParams() error {
	return nil
}

func (self *LoadTpFromFolder) RpcResult() any {
	var s string
	return &s
}
