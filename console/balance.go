/* Implementing balance related console commands.
 */
package console

import (
	"fmt"
)

func init() {
	commands["get_balance"] = &CmdGetBalance{}
}

// Data being sent to the rpc responder
type ArgsGetBalance struct {
	Tenant    string
	User      string
	Direction string
	BalanceId string
}

// Received back from query
type ReplyGetBalance struct {
	Tenant    string
	User      string
	Direction string
	BalanceId string
	Balance   float64
}

// Commander implementation
type CmdGetBalance struct {
	rpcMethod string
	rpcParams *ArgsGetBalance
	rpcResult *ReplyGetBalance
}

// name should be exec's name
func (self *CmdGetBalance) Usage(name string) string {
	return fmt.Sprintf("\n\tUsage: cgr-console [cfg_opts...{-h}] get_balance <tenant> <user> [<balanceid> [<direction>]]")
}

// set param defaults
func (self *CmdGetBalance) defaults() error {
	self.rpcMethod = "Responder.GetBalance"
	self.rpcParams = &ArgsGetBalance{BalanceId: "MONETARYOUT", Direction: "OUT"}
	return nil
}

// Parses command line args and builds CmdBalance value
func (self *CmdGetBalance) FromArgs(args []string) error {
	if len(args) < 4 {
		return fmt.Errorf(self.Usage(""))
	}
	// Args look OK, set defaults before going further	
	self.defaults()
	self.rpcParams.Tenant = args[2]
	self.rpcParams.User = args[3]
	if len(args) > 4 {
		self.rpcParams.BalanceId = args[4]
	}
	if len(args) > 5 {
		self.rpcParams.Direction = args[5]
	}
	return nil
}

func (self *CmdGetBalance) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetBalance) RpcParams() interface{} {
	return self.rpcParams
}

func (self *CmdGetBalance) RpcResult() interface{} {
	self.rpcResult = &ReplyGetBalance{}
	return self.rpcResult
}
