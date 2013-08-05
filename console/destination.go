package console

import (
	"fmt"
	"github.com/cgrates/cgrates/apier/v1"
)

func init() {
	commands["get_destination"] = &CmdGetDestination{}
}

// Commander implementation
type CmdGetDestination struct {
	rpcMethod string
	rpcParams *apier.AttrDestination
	rpcResult *apier.AttrDestination
}

// name should be exec's name
func (self *CmdGetDestination) Usage(name string) string {
	return fmt.Sprintf("\n\tUsage: cgr-console [cfg_opts...{-h}] get_destination <id>")
}

// set param defaults
func (self *CmdGetDestination) defaults() error {
	self.rpcMethod = "Apier.GetDestination"
	self.rpcParams = &apier.AttrDestination{}
	return nil
}

// Parses command line args and builds CmdBalance value
func (self *CmdGetDestination) FromArgs(args []string) error {
	if len(args) < 3 {
		return fmt.Errorf(self.Usage(""))
	}
	// Args look OK, set defaults before going further
	self.defaults()
	self.rpcParams.Id = args[2]
	return nil
}

func (self *CmdGetDestination) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetDestination) RpcParams() interface{} {
	return self.rpcParams
}

func (self *CmdGetDestination) RpcResult() interface{} {
	self.rpcResult = &apier.AttrDestination{}
	return self.rpcResult
}
