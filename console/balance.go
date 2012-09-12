/* Implementing balance related console commands.
 */
package console

import (
	"fmt"
	"os"
	"path/filepath"
)

type CmdGetBalance struct {
	User		string
	BalanceType	string
	Direction	string
	rpcMethod	string
	rpcParams	interface{}
	rpcResult	string
}

// name should be exec's name
func (self *CmdGetBalance) usage(name string) string {
	return fmt.Sprintf("usage: %s get_balance <user> <baltype> [<direction>]", name)
}

// set param defaults
func (self *CmdGetBalance) defaults() error {
	self.rpcMethod = "Responder.GetBalance"
	self.rpcParams = self
	self.BalanceType = "MONETARY"
	self.Direction = "OUT"
	return nil
}

func( self *CmdGetBalance) idxArgsToFields() map[int]string {
	return map[int]string{2: "User", 3: "BalanceType", 4: "Direction"}
}

// Parses command line args and builds CmdBalance value
func (self *CmdGetBalance) FromArgs(args []string) error {
	if len(os.Args) < 3 {
		return fmt.Errorf(self.usage(filepath.Base(args[0])))
	}
	// Args look OK, set defaults before going further
	self.defaults()
	// Dynamically set field values
	CmdFieldsFromArgs( self, args )
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
