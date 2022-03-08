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
	"sort"
	"strings"
	"time"

	"github.com/cgrates/cgrates/utils"
)

var (
	lineR = regexp.MustCompile(`(\w+)\s*=\s*(\[.*?\]|\".*?\"|\{.*?\}|.+?)(?:\s+|$)`)
	jsonR = regexp.MustCompile(`"(\w+)":(\[.+?\]|.+?)(?:,|}$)`)
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
			// log.Printf("%v (%v : %v)", typeField.Name, valField.Kind(), typeField.PkgPath)
			if len(typeField.PkgPath) > 0 { //unexported field
				continue
			}
			switch valField.Kind() {
			case reflect.Ptr, reflect.Struct:
				if valField.Kind() == reflect.Ptr {
					valField = reflect.New(valField.Type().Elem()).Elem()
					if valField.Kind() != reflect.Struct {
						// log.Printf("Here: %v (%v)", typeField.Name, valField.Kind())
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
	return utils.EmptyString
}

func ToJSON(line string) (jsn []byte) {
	if !strings.Contains(line, utils.AttrValueSep) && line != utils.EmptyString {
		line = fmt.Sprintf("Item=\"%s\"", line)
	}
	jsn = append(jsn, '{')
	for _, group := range lineR.FindAllStringSubmatch(line, -1) {
		if len(group) == 3 {
			jsn = append(jsn, []byte(fmt.Sprintf("\"%s\":%s,", group[1], group[2]))...)
		}
	}
	jsn = bytes.TrimRight(jsn, utils.FieldsSep)
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

func getStringValue(v interface{}, defaultDurationFields utils.StringSet) string {
	switch o := v.(type) {
	case nil:
		return "null"
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64:
		return fmt.Sprintf(`%v`, o)
	case string:
		return fmt.Sprintf(`"%s"`, o)
	case map[string]interface{}:
		return getMapAsString(o, defaultDurationFields)
	case []interface{}:
		return getSliceAsString(o, defaultDurationFields)
	}
	return utils.ToJSON(v)
}

func getSliceAsString(mp []interface{}, defaultDurationFields utils.StringSet) (out string) {
	out = utils.IdxStart
	for _, v := range mp {
		out += fmt.Sprintf(`%s,`, getStringValue(v, defaultDurationFields))
	}
	return strings.TrimSuffix(out, utils.FieldsSep) + utils.IdxEnd
}

func getMapAsString(mp map[string]interface{}, defaultDurationFields utils.StringSet) (out string) {
	// in order to find the data faster
	keylist := []string{} // add key value pairs to list so at the end we can sort them
	for k, v := range mp {
		if defaultDurationFields.Has(k) {
			if t, err := utils.IfaceAsDuration(v); err == nil {
				keylist = append(keylist, fmt.Sprintf(`"%s":"%s"`, k, t.String()))
				continue
			}
		}
		keylist = append(keylist, fmt.Sprintf(`"%s":%s`, k, getStringValue(v, defaultDurationFields)))
	}
	sort.Strings(keylist)
	return fmt.Sprintf(`{%s}`, strings.Join(keylist, utils.FieldsSep))
}

func GetFormatedResult(result interface{}, defaultDurationFields utils.StringSet) string {
	jsonResult, _ := json.Marshal(result)
	var mp map[string]interface{}
	if err := json.Unmarshal(jsonResult, &mp); err != nil {
		out, _ := json.MarshalIndent(result, utils.EmptyString, " ")
		return string(out)
	}
	mpstr := getMapAsString(mp, defaultDurationFields)
	var out bytes.Buffer
	json.Indent(&out, []byte(mpstr), utils.EmptyString, " ")
	return out.String()
}

func GetFormatedSliceResult(result interface{}, defaultDurationFields utils.StringSet) string {
	jsonResult, _ := json.Marshal(result)
	var mp []interface{}
	if err := json.Unmarshal(jsonResult, &mp); err != nil {
		out, _ := json.MarshalIndent(result, utils.EmptyString, " ")
		return string(out)
	}
	mpstr := getSliceAsString(mp, defaultDurationFields)
	var out bytes.Buffer
	json.Indent(&out, []byte(mpstr), utils.EmptyString, " ")
	return out.String()
}

func (ce *CommandExecuter) GetFormatedResult(result interface{}) string {
	out, _ := json.MarshalIndent(result, utils.EmptyString, " ")
	return string(out)
}
