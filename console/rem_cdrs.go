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
	commands["rem_cdrs"] = &CmdRemCdrs{}
}

// Commander implementation
type CmdRemCdrs struct {
	rpcMethod string
	rpcParams *utils.AttrRemCdrs
	rpcResult string
}

// name should be exec's name
func (self *CmdRemCdrs) Usage(name string) string {
	return fmt.Sprintf("\n\tUsage: cgr-console [cfg_opts...{-h}] rem_cdrs <cgrid> [<cdrid> [<cdrid>...]]")
}

// set param defaults
func (self *CmdRemCdrs) defaults() error {
	self.rpcMethod = "ApierV1.RemCdrs"
	self.rpcParams = &utils.AttrRemCdrs{}
	return nil
}

// Parses command line args and builds CmdBalance value
func (self *CmdRemCdrs) FromArgs(args []string) error {
	if len(args) < 3 {
		return fmt.Errorf(self.Usage(""))
	}
	// Args look OK, set defaults before going further
	self.defaults()
	self.rpcParams.CgrIds = args[2:]
	return nil
}

func (self *CmdRemCdrs) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdRemCdrs) RpcParams() interface{} {
	return self.rpcParams
}

func (self *CmdRemCdrs) RpcResult() interface{} {
	return &self.rpcResult
}
