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
	"strconv"
)

func init() {
	commands["add_balance"] = &CmdAddBalance{}
}

// Commander implementation
type CmdAddBalance struct {
	rpcMethod string
	rpcParams *apier.AttrAddBalance
	rpcResult string
}

// name should be exec's name
func (self *CmdAddBalance) Usage(name string) string {
	return fmt.Sprintf("\n\tUsage: cgr-console [cfg_opts...{-h}] add_balance <tenant> <account> <value> [<balanceid=monetary|sms|internet|internet_time|minutes> [<weight> [overwrite]]]")
}

// set param defaults
func (self *CmdAddBalance) defaults() error {
	self.rpcMethod = "ApierV1.AddBalance"
	self.rpcParams = &apier.AttrAddBalance{BalanceId: engine.CREDIT}
	self.rpcParams.Direction = "*out"
	return nil
}

// Parses command line args and builds CmdBalance value
func (self *CmdAddBalance) FromArgs(args []string) error {
	var err error
	if len(args) < 5 {
		return fmt.Errorf(self.Usage(""))
	}
	// Args look OK, set defaults before going further
	self.defaults()
	self.rpcParams.Tenant = args[2]
	self.rpcParams.Account = args[3]
	value, err := strconv.ParseFloat(args[4], 64)
	if err != nil {
		return err
	}
	self.rpcParams.Value = value
	if len(args) > 5 {
		self.rpcParams.BalanceId = args[5]
	}
	if len(args) > 6 {
		if self.rpcParams.Weight, err = strconv.ParseFloat(args[6], 64); err != nil {
			return fmt.Errorf("Cannot parse weight parameter")
		}
	}
	if len(args) > 7 {
		if args[7] == "overwrite" {
			self.rpcParams.Overwrite = true
		}
	}
	return nil
}

func (self *CmdAddBalance) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdAddBalance) RpcParams() interface{} {
	return self.rpcParams
}

func (self *CmdAddBalance) RpcResult() interface{} {
	return &self.rpcResult
}
