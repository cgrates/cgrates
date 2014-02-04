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
	"github.com/cgrates/cgrates/engine"
)

func init() {
	commands["get_balance"] = &CmdGetBalance{}
}

// Commander implementation
type CmdGetBalance struct {
	rpcMethod string
	rpcParams *apier.AttrGetBalance
	rpcResult float64
}

// name should be exec's name
func (self *CmdGetBalance) Usage(name string) string {
	return fmt.Sprintf("\n\tUsage: cgr-console [cfg_opts...{-h}] get_balance <tenant> <account> [<*monetary|*sms|*internet|*internet_time|*minutes> [<direction>]]")
}

// set param defaults
func (self *CmdGetBalance) defaults() error {
	self.rpcMethod = "ApierV1.GetBalance"
	self.rpcParams = &apier.AttrGetBalance{BalanceId: engine.CREDIT}
	self.rpcParams.Direction = "*out"
	return nil
}

// Parses command line args and builds CmdBalance value
func (self *CmdGetBalance) FromArgs(args []string) error {
	if len(args) < 4 {
		return fmt.Errorf(self.Usage(""))
	}
	// Args look OK, set defaults before going further
	self.defaults()
	self.rpcParams.Tenant = args[2]
	self.rpcParams.Account = args[3]
	if len(args) > 4 {
		self.rpcParams.BalanceId = args[4]
	}
	if len(args) > 5 {

		self.rpcParams.Direction = args[5]
	}
	return nil
}

func (self *CmdGetBalance) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetBalance) RpcParams() interface{} {
	return self.rpcParams
}

func (self *CmdGetBalance) RpcResult() interface{} {
	return &self.rpcResult
}
