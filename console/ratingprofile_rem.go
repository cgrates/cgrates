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

import (
	"reflect"

	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdRemRatingProfile{
		name:      "ratingprofile_rem",
		rpcMethod: utils.ApierV1RemoveRatingProfile,
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdRemRatingProfile struct {
	name      string
	rpcMethod string
	rpcParams *v1.AttrRemoveRatingProfile
	rpcResult string
	*CommandExecuter
}

func (self *CmdRemRatingProfile) Name() string {
	return self.name
}

func (self *CmdRemRatingProfile) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdRemRatingProfile) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &v1.AttrRemoveRatingProfile{Direction: utils.OUT}
	}
	return self.rpcParams
}

func (self *CmdRemRatingProfile) PostprocessRpcParams() error {
	if reflect.DeepEqual(self.rpcParams, &v1.AttrRemoveRatingProfile{Direction: utils.OUT}) {
		return utils.ErrMandatoryIeMissing
	}
	return nil
}

func (self *CmdRemRatingProfile) RpcResult() interface{} {
	var s string
	return &s
}
