// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdGetJSONConfig{
		name:      "get_json_section",
		rpcMethod: utils.ConfigSv1GetJSONSection,
		rpcParams: &config.StringWithArgDispatcher{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdGetJSONConfig struct {
	name      string
	rpcMethod string
	rpcParams *config.StringWithArgDispatcher
	*CommandExecuter
}

func (self *CmdGetJSONConfig) Name() string {
	return self.name
}

func (self *CmdGetJSONConfig) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetJSONConfig) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &config.StringWithArgDispatcher{ArgDispatcher: new(utils.ArgDispatcher)}
	}
	return self.rpcParams
}

func (self *CmdGetJSONConfig) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetJSONConfig) RpcResult() any {
	var s map[string]any
	return &s
}
