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
	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdSetRateProfile{
		name:      "rateprofile_set",
		rpcMethod: utils.APIerSv1SetRateProfile,
		rpcParams: &v1.RateProfileWithCache{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

type CmdSetRateProfile struct {
	name      string
	rpcMethod string
	rpcParams *v1.RateProfileWithCache
	*CommandExecuter
}

func (self *CmdSetRateProfile) Name() string {
	return self.name
}

func (self *CmdSetRateProfile) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdSetRateProfile) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &v1.RateProfileWithCache{RateProfileWithOpts: new(engine.RateProfileWithOpts)}
	}
	return self.rpcParams
}

func (self *CmdSetRateProfile) PostprocessRpcParams() error {
	return nil
}

func (self *CmdSetRateProfile) RpcResult() interface{} {
	var s string
	return &s
}
