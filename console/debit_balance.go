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
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"time"
)

func init() {
	commands["debit_balance"] = &CmdDebitBalance{}
}

// Commander implementation
type CmdDebitBalance struct {
	rpcMethod string
	rpcParams *engine.CallDescriptor
	rpcResult engine.CallCost
}

// name should be exec's name
func (self *CmdDebitBalance) Usage(name string) string {
	return fmt.Sprintf("\n\tUsage: cgr-console [cfg_opts...{-h}] debit_balance <tor> <tenant> <account> <subject> <destination> <start_time|*now> <duration>")
}

// set param defaults
func (self *CmdDebitBalance) defaults() error {
	self.rpcMethod = "Responder.Debit"
	self.rpcParams = &engine.CallDescriptor{Direction: "*out"}
	return nil
}

// Parses command line args and builds CmdBalance value
func (self *CmdDebitBalance) FromArgs(args []string) error {
	if len(args) != 9 {
		return fmt.Errorf(self.Usage(""))
	}
	// Args look OK, set defaults before going further
	self.defaults()
	var tStart time.Time
	var err error
	if args[7] == "*now" {
		tStart = time.Now()
	} else {
		tStart, err = utils.ParseDate(args[7])
		if err != nil {
			fmt.Println("\n*start_time* should have one of the formats:")
			fmt.Println("\ttime.RFC3339\teg:2013-08-07T17:30:00Z in UTC")
			fmt.Println("\tunix time\teg: 1383823746")
			fmt.Println("\t*now\t\tmetafunction transformed into localtime at query time")
			fmt.Println("\t+dur\t\tduration to be added to localtime (valid suffixes: ns, us/µs, ms, s, m, h)\n")
			return fmt.Errorf(self.Usage(""))
		}
	}
	callDur, err := utils.ParseDurationWithSecs(args[8])
	if err != nil {
		fmt.Println("\n\tExample durations: 60s for 60 seconds, 25m for 25minutes, 1m25s for one minute and 25 seconds\n")
	}
	self.rpcParams.TOR = args[2]
	self.rpcParams.Tenant = args[3]
	self.rpcParams.Account = args[4]
	self.rpcParams.Subject = args[5]
	self.rpcParams.Destination = args[6]
	self.rpcParams.TimeStart = tStart
	self.rpcParams.CallDuration = callDur
	self.rpcParams.TimeEnd = tStart.Add(callDur)
	return nil
}

func (self *CmdDebitBalance) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdDebitBalance) RpcParams() interface{} {
	return self.rpcParams
}

func (self *CmdDebitBalance) RpcResult() interface{} {
	return &self.rpcResult
}
