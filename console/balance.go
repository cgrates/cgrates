/* Implementing balance related console commands.
 */
package console

import (
	"fmt"
	"path/filepath"
)

// Data being sent to the rpc responder
type ArgsGetBalance struct {
	Tenant		string
	User		string
	Direction	string
	BalanceId	string
}


// Received back from query
type ReplyGetBalance struct {
	Tenant		string
	User		string
	Direction	string
	BalanceId	string
	Balance		float64
}

// Commander implementation
type CmdGetBalance struct {
	rpcMethod        string
	rpcParams        ArgsGetBalance
	rpcResult        ReplyGetBalance
	idxArgsToRpcPrms map[int]string
}

// name should be exec's name
func (self *CmdGetBalance) Usage(name string) string {
	return fmt.Sprintf("\n\tUsage: %s [cfg_opts...{-h}] get_balance <tenant> <user> [<balanceid> [<direction>]]", name)
}

// set param defaults
func (self *CmdGetBalance) defaults() error {
	self.idxArgsToRpcPrms = map[int]string{2: "Tenant", 3: "User", 4: "BalanceId", 5:"Direction" }
	self.rpcMethod = "Responder.GetBalance"
	self.rpcParams.BalanceId = "MONETARY"
	self.rpcParams.Direction = "OUT"
	return nil
}

// Parses command line args and builds CmdBalance value
func (self *CmdGetBalance) FromArgs(args []string) error {
	if len(args) < 4 {
		return fmt.Errorf(self.Usage(filepath.Base(args[0])))
	}
	// Args look OK, set defaults before going further
	self.defaults()
	// Dynamically set rpc params
	CmdRpcPrmsFromArgs(&self.rpcParams, args, self.idxArgsToRpcPrms)
	return nil
}

func (self *CmdGetBalance) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetBalance) RpcParams() interface{} {
	return &self.rpcParams
}

func (self *CmdGetBalance) RpcResult() interface{} {
	return &self.rpcResult
}
