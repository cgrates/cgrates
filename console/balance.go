/* Implementing balance related console commands.
 */
package console

import (
	"fmt"
	"os"
	"path/filepath"
)


type PrmsGetBalance	struct {
	User		string
	BalanceType	string
	Direction	string
}

type CmdGetBalance struct {
	rpcMethod		string
	rpcParams		PrmsGetBalance
	rpcResult		string
	idxArgsToRpcPrms	map[int]string
}

// name should be exec's name
func (self *CmdGetBalance) Usage(name string) string {
	return fmt.Sprintf("usage: %s get_balance <user> <baltype> [<direction>]", name)
}

// set param defaults
func (self *CmdGetBalance) defaults() error {
	self.idxArgsToRpcPrms = map[int]string{2: "User", 3: "BalanceType", 4: "Direction"}
	self.rpcMethod = "Responder.GetBalance"
	self.rpcParams.BalanceType = "MONETARY"
	self.rpcParams.Direction = "OUT"
	return nil
}

// Parses command line args and builds CmdBalance value
func (self *CmdGetBalance) FromArgs(args []string) error {
	if len(os.Args) < 3 {
		return fmt.Errorf(self.Usage(filepath.Base(args[0])))
	}
	// Args look OK, set defaults before going further
	self.defaults()
	// Dynamically set rpc params
	CmdRpcPrmsFromArgs( self.rpcParams, args, self.idxArgsToRpcPrms )
	return nil
}

func (self *CmdGetBalance) RpcMethod () string {
	return self.rpcMethod
}

func (self *CmdGetBalance) RpcParams() interface{} {
	return self.rpcParams
}

func (self *CmdGetBalance) RpcResult() interface{} {
	return &self.rpcResult
}
