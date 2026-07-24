// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdRelaodConfigSection{
		name:      "reload_config",
		rpcMethod: utils.ConfigSv1ReloadConfigFromPath,
		rpcParams: &config.ConfigReloadWithArgDispatcher{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdRelaodConfigSection struct {
	name      string
	rpcMethod string
	rpcParams *config.ConfigReloadWithArgDispatcher
	*CommandExecuter
}

func (self *CmdRelaodConfigSection) Name() string {
	return self.name
}

func (self *CmdRelaodConfigSection) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdRelaodConfigSection) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &config.ConfigReloadWithArgDispatcher{ArgDispatcher: new(utils.ArgDispatcher)}
	}
	return self.rpcParams
}

func (self *CmdRelaodConfigSection) PostprocessRpcParams() error {
	return nil
}

func (self *CmdRelaodConfigSection) RpcResult() any {
	var s string
	return &s
}
