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
	c := &CmdGetMaxUsage{
		name:      "maxusage",
		rpcMethod: utils.APIerSv1GetMaxUsage,
		clientArgs: []string{utils.ToR, utils.RequestType, utils.Tenant,
			utils.Category, utils.AccountField, utils.Subject, utils.Destination,
			utils.SetupTime, utils.AnswerTime, utils.Usage, utils.ExtraFields},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdGetMaxUsage struct {
	name       string
	rpcMethod  string
	rpcParams  *engine.UsageRecordWithAPIOpts
	clientArgs []string
	*CommandExecuter
}

func (self *CmdGetMaxUsage) Name() string {
	return self.name
}

func (self *CmdGetMaxUsage) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetMaxUsage) RpcParams(reset bool) any {
	if reset || self.rpcParams == nil {
		self.rpcParams = &engine.UsageRecordWithAPIOpts{
			UsageRecord: new(engine.UsageRecord),
			APIOpts:     make(map[string]any),
		}
	}
	return self.rpcParams
}

func (self *CmdGetMaxUsage) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetMaxUsage) RpcResult() any {
	var d time.Duration
	return &d
}

func (self *CmdGetMaxUsage) ClientArgs() []string {
	return self.clientArgs
}

func (self *CmdGetMaxUsage) GetFormatedResult(result any) string {
	if tv, canCast := result.(*time.Duration); canCast {
		return fmt.Sprintf(`"%s"`, tv.String())
	}
	out, _ := json.MarshalIndent(result, utils.EmptyString, " ")
	return string(out)
}
