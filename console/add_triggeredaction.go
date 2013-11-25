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
	"strconv"
)

func init() {
	commands["add_triggeredaction"] = &CmdAddTriggeredAction{}
}

// Commander implementation
type CmdAddTriggeredAction struct {
	rpcMethod string
	rpcParams *apier.AttrAddActionTrigger
	rpcResult string
}

// name should be exec's name
func (self *CmdAddTriggeredAction) Usage(name string) string {
	return fmt.Sprintf("\n\tUsage: cgr-console [cfg_opts...{-h}] add_triggeredaction <tenant> <account> <balanceid> <thresholdvalue> <destinationid> <weight> <actionsid> [<direction>]")
}

// set param defaults
func (self *CmdAddTriggeredAction) defaults() error {
	self.rpcMethod = "ApierV1.AddTriggeredAction"
	self.rpcParams = &apier.AttrAddActionTrigger{Direction: "*out"}
	return nil
}

// Parses command line args and builds CmdBalance value
func (self *CmdAddTriggeredAction) FromArgs(args []string) error {
	if len(args) < 9 {
		return fmt.Errorf(self.Usage(""))
	}
	// Args look OK, set defaults before going further
	self.defaults()
	self.rpcParams.Tenant = args[2]
	self.rpcParams.Account = args[3]
	self.rpcParams.BalanceId = args[4]
	thresholdvalue, err := strconv.ParseFloat(args[5], 64)
	if err != nil {
		return err
	}
	self.rpcParams.ThresholdValue = thresholdvalue
	self.rpcParams.DestinationId = args[6]
	weight, err := strconv.ParseFloat(args[7], 64)
	if err != nil {
		return err
	}
	self.rpcParams.Weight = weight
	self.rpcParams.ActionsId = args[8]

	if len(args) > 9 {
		self.rpcParams.Direction = args[9]
	}
	return nil
}

func (self *CmdAddTriggeredAction) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdAddTriggeredAction) RpcParams() interface{} {
	return self.rpcParams
}

func (self *CmdAddTriggeredAction) RpcResult() interface{} {
	return &self.rpcResult
}
