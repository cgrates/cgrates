// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdGetRatingProfile{
		name:      "ratingprofile",
		rpcMethod: utils.APIerSv1GetRatingProfile,
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdGetRatingProfile struct {
	name      string
	rpcMethod string
	rpcParams *utils.AttrGetRatingProfile
	rpcResult string
	*CommandExecuter
}

func (self *CmdGetRatingProfile) Name() string {
	return self.name
}

func (self *CmdGetRatingProfile) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetRatingProfile) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = new(utils.AttrGetRatingProfile)
	}
	return self.rpcParams
}

func (self *CmdGetRatingProfile) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetRatingProfile) RpcResult() any {
	var s engine.RatingProfile
	return &s
}
