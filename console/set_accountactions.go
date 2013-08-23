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
	commands["set_accountactions"] = &CmdSetAccountActions{}
}

// Commander implementation
type CmdSetAccountActions struct {
	rpcMethod string
	rpcParams *apier.AttrSetAccountActions
	rpcResult string
}

// name should be exec's name
func (self *CmdSetAccountActions) Usage(name string) string {
	return fmt.Sprintf("\n\tUsage: cgr-console [cfg_opts...{-h}] set_accountactions <tpid> <accountactionsid>")
}

// set param defaults
func (self *CmdSetAccountActions) defaults() error {
	self.rpcMethod = "ApierV1.SetAccountActions"
	self.rpcParams = &apier.AttrSetAccountActions{}
	return nil
}

// Parses command line args and builds CmdBalance value
func (self *CmdSetAccountActions) FromArgs(args []string) error {
	if len(args) < 3 {
		return fmt.Errorf(self.Usage(""))
	}
	// Args look OK, set defaults before going further
	self.defaults()
	self.rpcParams.TPid = args[2]
	self.rpcParams.AccountActionsId = args[3]
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
