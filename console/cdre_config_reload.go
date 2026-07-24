// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdCdreConfigReload{
		name:      "cdre_config_reload",
		rpcMethod: utils.APIerSv1ReloadCdreConfig,
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdCdreConfigReload struct {
	name      string
	rpcMethod string
	rpcParams *v1.ConfigPathArg
	*CommandExecuter
}

func (self *CmdCdreConfigReload) Name() string {
	return self.name
}

func (self *CmdCdreConfigReload) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdCdreConfigReload) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = new(v1.ConfigPathArg)
	}
	return self.rpcParams
}

func (self *CmdCdreConfigReload) PostprocessRpcParams() error {
	return nil
}

func (self *CmdCdreConfigReload) RpcResult() any {
	var s string
	return &s
}
