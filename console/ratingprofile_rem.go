// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"reflect"

	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdRemRatingProfile{
		name:      "ratingprofile_rem",
		rpcMethod: utils.APIerSv1RemoveRatingProfile,
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

func (self *CmdRemRatingProfile) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &v1.AttrRemoveRatingProfile{}
	}
	return self.rpcParams
}

func (self *CmdRemRatingProfile) PostprocessRpcParams() error {
	if reflect.DeepEqual(self.rpcParams, &v1.AttrRemoveRatingProfile{}) {
		return utils.ErrMandatoryIeMissing
	}
	return nil
}

func (self *CmdRemRatingProfile) RpcResult() any {
	var s string
	return &s
}
