package console


import (
	"fmt"
	"reflect"
	"path/filepath"
)


// Console Command interface
type Commander interface {
	FromArgs(args []string)	error 		// Load data from os arguments or flag.Args()
	usage(string)		string		// usage message
	defaults()		error		// set default field values
	idxArgsToFields()	map[int]string	// field's index position in command arguments
	RpcMethod()		string		// Method which should be called remotely
	RpcParams()		interface{}	// Parameters to send out on rpc
	RpcResult()		interface{}	// Only requirement is to have a String method to print on console
}


// Set command fields based on indexes defined in default()
func CmdFieldsFromArgs( cmd Commander, args []string ) {
	for idx := range args {
		fldName, hasIdx := cmd.idxArgsToFields()[idx]
		if !hasIdx {
			continue
		}
		// field defined to be set by os.Args index
		if fld := reflect.ValueOf(cmd).Elem().FieldByName(fldName); fld.Kind() == reflect.String {
			fld.SetString(args[idx])
		} else if fld.Kind() == reflect.Int {
			fld.SetInt(1) // Placeholder for future usage of data types other than strings
		}
	}
}

// Process args and return right command Value or error
func GetCommandValue( args []string ) ( Commander, error ) {
	if len(args) < 2 {
                return nil, fmt.Errorf( "usage: %s <command>\n", filepath.Base(args[0]) )
        }
        cmd := args[1]
        var cmdVal Commander
	switch cmd {
	case "status":
		cmdVal = &CmdStatus{}
	case "get_balance":
		cmdVal = &CmdGetBalance{}
	default:
		return nil, fmt.Errorf( "usage: %s <command>\n", filepath.Base(args[0]) )
	}
	if err := cmdVal.FromArgs(args); err != nil {
		return nil, err
        }
	return cmdVal, nil
}


