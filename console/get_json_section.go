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
		rpcMethod: utils.ConfigSv1GetConfig,
		rpcParams: &config.SectionWithAPIOpts{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdGetJSONConfig struct {
	name      string
	rpcMethod string
	rpcParams *config.SectionWithAPIOpts
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
		self.rpcParams = &config.SectionWithAPIOpts{APIOpts: make(map[string]any)}
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
