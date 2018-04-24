/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

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

import "github.com/cgrates/cgrates/apier/v1"

func init() {
	c := &CmdSetRatingProfile{
		name:      "ratingprofile_set",
		rpcMethod: "ApierV1.SetRatingProfile",
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdSetRatingProfile struct {
	name      string
	rpcMethod string
	rpcParams *v1.AttrSetRatingProfile
	rpcResult string
	*CommandExecuter
}

func (self *CmdSetRatingProfile) Name() string {
	return self.name
}

func (self *CmdSetRatingProfile) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdSetRatingProfile) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &v1.AttrSetRatingProfile{}
	}
	return self.rpcParams
}

func (self *CmdSetRatingProfile) PostprocessRpcParams() error {
	return nil
}

func (self *CmdSetRatingProfile) RpcResult() interface{} {
	var s string
	return &s
}
