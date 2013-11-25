/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2013 ITsysCOM

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
	"fmt"
	"github.com/cgrates/cgrates/apier/v1"
)

func init() {
	commands["execute_action"] = &CmdExecuteAction{}
}

// Commander implementation
type CmdExecuteAction struct {
	rpcMethod string
	rpcParams *apier.AttrExecuteAction
	rpcResult string
}

// name should be exec's name
func (self *CmdExecuteAction) Usage(name string) string {
	return fmt.Sprintf("\n\tUsage: cgr-console [cfg_opts...{-h}] execute_action <tenant> <account> <actionsid> [<direction>]")
}

// set param defaults
func (self *CmdExecuteAction) defaults() error {
	self.rpcMethod = "ApierV1.ExecuteAction"
	self.rpcParams = &apier.AttrExecuteAction{Direction: "*out"}
	return nil
}

// Parses command line args and builds CmdBalance value
func (self *CmdExecuteAction) FromArgs(args []string) error {
	if len(args) < 5 {
		return fmt.Errorf(self.Usage(""))
	}
	// Args look OK, set defaults before going further
	self.defaults()
	self.rpcParams.Tenant = args[2]
	self.rpcParams.Account = args[3]
	self.rpcParams.ActionsId = args[4]
	if len(args) > 5 {
		self.rpcParams.Direction = args[5]
	}
	return nil
}

func (self *CmdExecuteAction) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdExecuteAction) RpcParams() interface{} {
	return self.rpcParams
}

func (self *CmdExecuteAction) RpcResult() interface{} {
	return &self.rpcResult
}
