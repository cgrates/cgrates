/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

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
	"time"

	"github.com/cgrates/cgrates/utils"
)

var (
	lineR = regexp.MustCompile(`(\w+)\s*=\s*(\[.*?\]|\".*?\"|\{.*?\}|.+?)(?:\s+|$)`)
	jsonR = regexp.MustCompile(`"(\w+)":(\[.+?\]|.+?)[,|}]`)
)

// Commander implementation
type CommandExecuter struct {
	command Commander
}

func (ce *CommandExecuter) Usage() string {
	jsn, _ := json.Marshal(ce.command.RpcParams(true))
	return fmt.Sprintf("\n\tUsage: %s %s \n", ce.command.Name(), FromJSON(jsn, ce.command.ClientArgs()))
}

// Parses command line args and builds CmdBalance value
func (ce *CommandExecuter) FromArgs(args string, verbose bool) error {
	params := ce.command.RpcParams(true)
	if err := json.Unmarshal(ToJSON(args), params); err != nil {
		return err
	}
	if verbose {
		jsn, _ := json.Marshal(params)
		fmt.Println(ce.command.Name(), FromJSON(jsn, ce.command.ClientArgs()))
	}
	return nil
}

func (ce *CommandExecuter) clientArgs(iface interface{}) (args []string) {
	val := reflect.ValueOf(iface)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
		iface = val.Interface()
	}
	typ := reflect.TypeOf(iface)
	if val.Kind() == reflect.Struct {
		for i := 0; i < typ.NumField(); i++ {
			valField := val.Field(i)
			typeField := typ.Field(i)
			//log.Printf("%v (%v : %v)", typeField.Name, valField.Kind(), typeField.PkgPath)
			if len(typeField.PkgPath) > 0 { //unexported field
				continue
			}
			switch valField.Kind() {
			case reflect.Ptr, reflect.Struct:
				if valField.Kind() == reflect.Ptr {
					valField = reflect.New(valField.Type().Elem()).Elem()
					if valField.Kind() != reflect.Struct {
						//log.Printf("Here: %v (%v)", typeField.Name, valField.Kind())
						args = append(args, typeField.Name)
						continue
					}
				}
				valInterf := valField.Interface()
				if _, canCast := valInterf.(time.Time); canCast {
					args = append(args, typeField.Name)
					continue
				}
				args = append(args, ce.clientArgs(valInterf)...)
			default:
				args = append(args, typeField.Name)
			}
		}
	}
	return
}

func (ce *CommandExecuter) ClientArgs() (args []string) {
	return ce.clientArgs(ce.command.RpcParams(true))
}

// To be overwritten by commands that do not need a rpc call
func (ce *CommandExecuter) LocalExecute() string {
	return ""
}

func ToJSON(line string) (jsn []byte) {
	if !strings.Contains(line, "=") && line != "" {
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
