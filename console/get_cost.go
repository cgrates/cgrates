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
	"encoding/json"
	"fmt"

	"github.com/cgrates/cgrates/engine"
)

func init() {
	commands["get_cost"] = &CmdGetCost{
		rpcMethod:  "Responder.GetCost",
		clientArgs: []string{"Direction", "TOR", "Tenant", "Subject", "Account", "Destination", "TimeStart", "TimeEnd", "CallDuration", "FallbackSubject"},
	}
}

// Commander implementation
type CmdGetCost struct {
	rpcMethod  string
	rpcParams  *engine.CallDescriptor
	rpcResult  engine.CallCost
	clientArgs []string
}

func (self *CmdGetCost) Usage() string {
	jsn, _ := json.Marshal(engine.CallDescriptor{Direction: "*out"})
	return "\n\tUsage: get_cost " + FromJSON(jsn, self.clientArgs) + "\n"
}

// Parses command line args and builds CmdBalance value
func (self *CmdGetCost) FromArgs(args string, verbose bool) error {
	if len(args) == 0 {
		return fmt.Errorf(self.Usage())
	}
	// defaults
	self.rpcParams = &engine.CallDescriptor{Direction: "*out"}

	if err := json.Unmarshal(ToJSON(args), &self.rpcParams); err != nil {
		return err
	}
	if verbose {
		jsn, _ := json.Marshal(self.rpcParams)
		fmt.Println("get_cost ", FromJSON(jsn, self.clientArgs))
	}
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

func (self *CmdGetCost) ClientArgs() []string {
	return self.clientArgs
}
