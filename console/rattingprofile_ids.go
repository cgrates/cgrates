// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdGetRatingProfileIDs{
		name:      "ratingprofil_ids",
		rpcMethod: utils.APIerSv1GetRatingProfileIDs,
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdGetRatingProfileIDs struct {
	name      string
	rpcMethod string
	rpcParams *utils.TenantArgWithPaginator
	*CommandExecuter
}

func (self *CmdGetRatingProfileIDs) Name() string {
	return self.name
}

func (self *CmdGetRatingProfileIDs) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetRatingProfileIDs) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = new(utils.TenantArgWithPaginator)
	}
	return self.rpcParams
}

func (self *CmdGetRatingProfileIDs) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetRatingProfileIDs) RpcResult() any {
	var s []string
	return &s
}
