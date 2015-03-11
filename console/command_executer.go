/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012-2015 ITsysCOM

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package console

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/cgrates/cgrates/utils"
)

var (
	lineR = regexp.MustCompile(`(\w+)\s*=\s*(\[.+?\]|.+?)(?:\s+|$)`)
	jsonR = regexp.MustCompile(`"(\w+)":(\[.+?\]|.+?)[,|}]`)
)

// Commander implementation
type CommandExecuter struct {
	command Commander
}

func (ce *CommandExecuter) Usage() string {
	jsn, _ := json.Marshal(ce.command.RpcParams())
	return fmt.Sprintf("\n\tUsage: %s %s \n", ce.command.Name(), FromJSON(jsn, ce.command.ClientArgs()))
}

// Parses command line args and builds CmdBalance value
func (ce *CommandExecuter) FromArgs(args string, verbose bool) error {
	if err := json.Unmarshal(ToJSON(args), ce.command.RpcParams()); err != nil {
		return err
	}
	if verbose {
		jsn, _ := json.Marshal(ce.command.RpcParams())
		fmt.Println(ce.command.Name(), FromJSON(jsn, ce.command.ClientArgs()))
	}
	return nil
}

func (ce *CommandExecuter) ClientArgs() (args []string) {
	val := reflect.ValueOf(ce.command.RpcParams()).Elem()
	for i := 0; i < val.NumField(); i++ {
		typeField := val.Type().Field(i)
		args = append(args, typeField.Name)
	}
	return
}

// To be overwritten by commands that do not need a rpc call
func (ce *CommandExecuter) LocalExecute() string {
	return ""
}

func ToJSON(line string) (jsn []byte) {
	if !strings.Contains(line, "=") {
		line = fmt.Sprintf("Item=\"%s\"", line)
	}
	jsn = append(jsn, '{')
	for _, group := range lineR.FindAllStringSubmatch(line, -1) {
		if len(group) == 3 {
			jsn = append(jsn, []byte(fmt.Sprintf("\"%s\":%s,", group[1], group[2]))...)
		}
	}
	jsn = bytes.TrimRight(jsn, ",")
	jsn = append(jsn, '}')
	return
}

func FromJSON(jsn []byte, interestingFields []string) (line string) {
	if !bytes.Contains(jsn, []byte{':'}) {
		return fmt.Sprintf("\"%s\"", string(jsn))
	}
	for _, group := range jsonR.FindAllSubmatch(jsn, -1) {
		if len(group) == 3 {
			if utils.IsSliceMember(interestingFields, string(group[1])) {
				line += fmt.Sprintf("%s=%s ", group[1], group[2])
			}
		}
	}
	return strings.TrimSpace(line)
}
