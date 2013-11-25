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
	"github.com/cgrates/cgrates/utils"
)

func init() {
	commands["set_accountactions"] = &CmdSetAccountActions{}
}

// Commander implementation
type CmdSetAccountActions struct {
	rpcMethod string
	rpcParams *utils.TPAccountActions
	rpcResult string
}

// name should be exec's name
func (self *CmdSetAccountActions) Usage(name string) string {
	return fmt.Sprintf("\n\tUsage: cgr-console [cfg_opts...{-h}] set_accountactions <tpid> <loadid> <tenant> <account>")
}

// set param defaults
func (self *CmdSetAccountActions) defaults() error {
	self.rpcMethod = "ApierV1.SetAccountActions"
	return nil
}

// Parses command line args and builds CmdBalance value
func (self *CmdSetAccountActions) FromArgs(args []string) error {
	if len(args) < 3 {
		return fmt.Errorf(self.Usage(""))
	}
	// Args look OK, set defaults before going further
	self.defaults()
	self.rpcParams = &utils.TPAccountActions{TPid: args[2], LoadId: args[3], Tenant: args[4], Account: args[5], Direction: "*out"}
	return nil
}

func (self *CmdSetAccountActions) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdSetAccountActions) RpcParams() interface{} {
	return self.rpcParams
}

func (self *CmdSetAccountActions) RpcResult() interface{} {
	return &self.rpcResult
}
