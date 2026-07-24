// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdSetRatingProfile{
		name:      "ratingprofile_set",
		rpcMethod: utils.APIerSv1SetRatingProfile,
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdSetRatingProfile struct {
	name      string
	rpcMethod string
	rpcParams *utils.AttrSetRatingProfile
	rpcResult string
	*CommandExecuter
}

func (self *CmdSetRatingProfile) Name() string {
	return self.name
}

func (self *CmdSetRatingProfile) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdSetRatingProfile) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = new(utils.AttrSetRatingProfile)
	}
	return self.rpcParams
}

func (self *CmdSetRatingProfile) PostprocessRpcParams() error {
	return nil
}

func (self *CmdSetRatingProfile) RpcResult() any {
	var s string
	return &s
}
