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
	"time"
	"strconv"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	commands["get_cost"] = &CmdGetCost{}
}

// Commander implementation
type CmdGetCost struct {
	rpcMethod string
	rpcParams *engine.CallDescriptor
	rpcResult engine.CallCost
}

// name should be exec's name
func (self *CmdGetCost) Usage(name string) string {
	return fmt.Sprintf("\n\tUsage: cgr-console [cfg_opts...{-h}] get_cost <tor> <tenant> <subject> <destination> <start_time|*now> <duration>")
}

// set param defaults
func (self *CmdGetCost) defaults() error {
	self.rpcMethod = "Responder.GetCost"
	self.rpcParams = &engine.CallDescriptor{ Direction: "*out" }
	return nil
}

// Parses command line args and builds CmdBalance value
func (self *CmdGetCost) FromArgs(args []string) error {
	if len(args) != 8 {
		return fmt.Errorf(self.Usage(""))
	}
	// Args look OK, set defaults before going further
	self.defaults()
	var tStart time.Time
	var err error
	if args[6] == "*now" {
		tStart = time.Now()
	} else {
		tStart,err = utils.ParseDate(args[6])
		if err != nil {
			fmt.Println("Cannot parse start time")
			return fmt.Errorf(self.Usage(""))
		}
	}
	if _, err := strconv.Atoi(args[7]); err == nil { // No suffix, default to seconds
		args[7] += "s"
	}
	callDur, err := time.ParseDuration(args[7])
	if err != nil {
		fmt.Println("Cannot parse duration")
		return fmt.Errorf(self.Usage(""))
	}
	self.rpcParams.TOR = args[2]
	self.rpcParams.Tenant = args[3]
	self.rpcParams.Subject = args[4]
	self.rpcParams.Destination = args[5]
	self.rpcParams.TimeStart = tStart
	self.rpcParams.CallDuration = callDur
	self.rpcParams.TimeEnd = tStart.Add(callDur)
	return nil
}

func (self *CmdGetCost) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetCost) RpcParams() interface{} {
	return self.rpcParams
}

func (self *CmdGetCost) RpcResult() interface{} {
	return &self.rpcResult
}
