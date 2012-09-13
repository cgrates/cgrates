package console

import (
	"fmt"
	"path/filepath"
	"reflect"
)

// Console Command interface
type Commander interface {
	FromArgs(args []string) error // Load data from os arguments or flag.Args()
	Usage(string) string          // usage message
	RpcMethod() string            // Method which should be called remotely
	RpcParams() interface{}       // Parameters to send out on rpc
	RpcResult() interface{}       // Only requirement is to have a String method to print on console
	defaults() error              // set default field values
}

// Set command fields based on indexes defined in default()
func CmdRpcPrmsFromArgs(rpcPrms interface{}, args []string, idxArgsToRpcPrms map[int]string) {
	for idx := range args {
		fldName, hasIdx := idxArgsToRpcPrms[idx]
		if !hasIdx {
			continue
		}
		// field defined to be set by os.Args index
		if fld := reflect.ValueOf(rpcPrms).Elem().FieldByName(fldName); fld.Kind() == reflect.String {
			fld.SetString(args[idx])
		} else if fld.Kind() == reflect.Int {
			fld.SetInt(1) // Placeholder for future usage of data types other than strings
		}
	}
}

// Process args and return right command Value or error
func GetCommandValue(args []string) (Commander, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("usage: %s <command>\n", filepath.Base(args[0]))
	}
	cmd := args[1]
	var cmdVal Commander
	switch cmd {
	case "status":
		cmdVal = &CmdStatus{}
	case "get_balance":
		cmdVal = &CmdGetBalance{}
	default:
		return nil, fmt.Errorf("usage: %s <command>\n", filepath.Base(args[0]))
	}
	if err := cmdVal.FromArgs(args); err != nil {
		return nil, err
	}
	return cmdVal, nil
}
