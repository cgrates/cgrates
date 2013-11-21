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
	commands["set_ratingprofile"] = &CmdSetrRatingProfile{}
}

// Commander implementation
type CmdSetrRatingProfile struct {
	rpcMethod string
	rpcParams *utils.TPRatingProfile
	rpcResult string
}

// name should be exec's name
func (self *CmdSetrRatingProfile) Usage(name string) string {
	return fmt.Sprintf("\n\tUsage: cgr-console [cfg_opts...{-h}] set_ratingprofile <tpid> <loadid> <tenant> <tor> <subject>")
}

// set param defaults
func (self *CmdSetrRatingProfile) defaults() error {
	self.rpcMethod = "ApierV1.SetRatingProfile"
	self.rpcParams = &utils.TPRatingProfile{Direction:"*out"}
	return nil
}

// Parses command line args and builds CmdBalance value
func (self *CmdSetrRatingProfile) FromArgs(args []string) error {
	if len(args) < 3 {
		return fmt.Errorf(self.Usage(""))
	}
	// Args look OK, set defaults before going further
	self.defaults()
	self.rpcParams.TPid = args[2]
	self.rpcParams.LoadId = args[3]
	self.rpcParams.Tenant = args[4]
	self.rpcParams.TOR = args[5]
	self.rpcParams.Direction = args[6]
	return nil
}

func (self *CmdSetrRatingProfile) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdSetrRatingProfile) RpcParams() interface{} {
	return self.rpcParams
}

func (self *CmdSetrRatingProfile) RpcResult() interface{} {
	return &self.rpcResult
}
