/* Implementing balance related console commands.
 */
package console

import (
	"fmt"
	"path/filepath"
)

type PrmsGetBalance struct {
	Tenant		string
	User		string
	Direction	string
	BalanceTag	string
}

type CmdGetBalance struct {
	rpcMethod        string
	rpcParams        PrmsGetBalance
	rpcResult        string
	idxArgsToRpcPrms map[int]string
}

// name should be exec's name
func (self *CmdGetBalance) Usage(name string) string {
	return fmt.Sprintf("\n\tUsage: %s get_balance <tenant> <user> [<balance_tag> [<direction>]]", name)
}

// set param defaults
func (self *CmdGetBalance) defaults() error {
	self.idxArgsToRpcPrms = map[int]string{2: "Tenant", 3: "User", 4: "BalanceTag", 5:"Direction" }
	self.rpcMethod = "Responder.GetBalance"
	self.rpcParams.BalanceTag = "MONETARY"
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
