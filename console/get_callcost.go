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

	"github.com/cgrates/cgrates/apier"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	commands["get_callcost"] = &CmdGetCallCost{}
}

// Commander implementation
type CmdGetCallCost struct {
	rpcMethod string
	rpcParams *apier.AttrGetCallCost
	rpcResult *engine.CallCost
}

// name should be exec's name
func (self *CmdGetCallCost) Usage(name string) string {
	return fmt.Sprintf("\n\tUsage: cgr-console [cfg_opts...{-h}] get_callcost <cgrid> [<runid>]")
}

// set param defaults
func (self *CmdGetCallCost) defaults() error {
	self.rpcMethod = "ApierV1.GetCallCostLog"
	self.rpcParams = &apier.AttrGetCallCost{RunId: utils.DEFAULT_RUNID}
	return nil
}

// Parses command line args and builds CmdBalance value
func (self *CmdGetCallCost) FromArgs(args []string) error {
	if len(args) < 3 {
		return fmt.Errorf(self.Usage(""))
	}
	// Args look OK, set defaults before going further
	self.defaults()
	self.rpcParams.CgrId = args[2]
	if len(args) == 4 {
		self.rpcParams.RunId = args[3]
	}
	return nil
}

func (self *CmdGetCallCost) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetCallCost) RpcParams() interface{} {
	return self.rpcParams
}

func (self *CmdGetCallCost) RpcResult() interface{} {
	return &self.rpcResult
}
