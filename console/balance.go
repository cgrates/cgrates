/* Implementing balance related console commands.
 */
package console

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
)

// Console Command interface
type Command interface {
	FromOsArgs(args []string) error // Load data from os arguments or flag.Args()
	usage(string) string            // usage message
	defaults() error                // set default field values
}

type CmdGetBalance struct {
	User        string
	BalanceType string
	Direction   string
}

// name should be exec's name
func (self *CmdGetBalance) usage(name string) string {
	return fmt.Sprintf("usage: %s get_balance <user> <baltype> [<direction>]", name)
}

// set param defaults
func (self *CmdGetBalance) defaults() error {
	self.BalanceType = "MONETARY"
	self.Direction = "OUT"
	return nil
}

// Parses command line args and builds CmdBalance value
func (self *CmdGetBalance) FromOsArgs(args []string) error {
	// Map arg indexes to "self.Value"s
	idxArgsToFields := map[int]string{2: "User", 3: "BalanceType", 4: "Direction"}
	if len(os.Args) < 3 {
		return fmt.Errorf(self.usage(filepath.Base(args[0])))
	}
	// Args look OK, set defaults before going further
	self.defaults()
	// Dynamically set field values
	for idx := range args {
		fldName, hasIdx := idxArgsToFields[idx]
		if !hasIdx {
			continue
		}
		// field defined to be set by os.Args index
		if fld := reflect.ValueOf(self).Elem().FieldByName(fldName); fld.Kind() == reflect.String {
			fld.SetString(args[idx])
		} else if fld.Kind() == reflect.Int {
			fld.SetInt(1) // Placeholder for future usage of data types other than strings
		}
	}
	return nil
}
