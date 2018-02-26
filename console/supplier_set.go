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

import "github.com/cgrates/cgrates/engine"

func init() {
	c := &CmdSetSupplier{
		name:      "supplier_set",
		rpcMethod: "ApierV1.SetSupplierProfile",
		rpcParams: &engine.SupplierProfile{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

type CmdSetSupplier struct {
	name      string
	rpcMethod string
	rpcParams *engine.SupplierProfile
	*CommandExecuter
}

func (self *CmdSetSupplier) Name() string {
	return self.name
}

func (self *CmdSetSupplier) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdSetSupplier) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &engine.SupplierProfile{}
	}
	return self.rpcParams
}

func (self *CmdSetSupplier) PostprocessRpcParams() error {
	return nil
}

func (self *CmdSetSupplier) RpcResult() interface{} {
	var s string
	return &s
}
