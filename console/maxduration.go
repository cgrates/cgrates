// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdGetMaxDuration{
		name:       "maxduration",
		rpcMethod:  utils.ResponderGetMaxSessionTime,
		clientArgs: []string{utils.Category, utils.ToR, utils.Tenant, utils.Subject, utils.AccountField, utils.Destination, utils.TimeStart, utils.TimeEnd, utils.CallDuration, utils.FallbackSubject},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdGetMaxDuration struct {
	name       string
	rpcMethod  string
	rpcParams  *engine.CallDescriptorWithAPIOpts
	clientArgs []string
	*CommandExecuter
}

func (self *CmdGetMaxDuration) Name() string {
	return self.name
}

func (self *CmdGetMaxDuration) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetMaxDuration) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &engine.CallDescriptorWithAPIOpts{
			CallDescriptor: new(engine.CallDescriptor),
			APIOpts:        make(map[string]any),
		}
	}
	return self.rpcParams
}

func (self *CmdGetMaxDuration) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetMaxDuration) RpcResult() any {
	var d time.Duration
	return &d
}

func (self *CmdGetMaxDuration) ClientArgs() []string {
	return self.clientArgs
}

func (self *CmdGetMaxDuration) GetFormatedResult(result any) string {
	if tv, canCast := result.(*time.Duration); canCast {
		return fmt.Sprintf(`"%s"`, tv.String())
	}
	out, _ := json.MarshalIndent(result, utils.EmptyString, " ")
	return string(out)
}
