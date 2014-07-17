/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2013 ITsysCOM

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
	"fmt"
	"strings"
)

var (
	commands = make(map[string]Commander)
)

// Console Command interface
type Commander interface {
	FromArgs(args string, verbose bool) error // Load data from os arguments or flag.Args()
	Usage() string                            // usage message
	RpcMethod() string                        // Method which should be called remotely
	RpcParams() interface{}                   // Parameters to send out on rpc
	RpcResult() interface{}                   // Only requirement is to have a String method to print on console
	ClientArgs() []string                     // for autocompletion
	Name() string
	LocalExecute() string
}

func GetCommands() map[string]Commander {
	return commands
}

func getAvailabelCommandsErr() error {
	var keys []string
	for key, _ := range commands {
		keys = append(keys, key)
	}
	return fmt.Errorf("\n\tAvailable commands <%s>\n", strings.Join(keys, "|"))
}

// Process args and return right command Value or error
func GetCommandValue(command string, verbose bool) (Commander, error) {
	if len(command) == 0 {
		return nil, getAvailabelCommandsErr()
	}
	firstSpace := strings.Index(command, " ")
	var cmdName string
	var cmdArgs string
	if firstSpace <= 0 {
		cmdName = command[:len(command)]
		cmdArgs = ""
	} else {
		cmdName = command[:firstSpace]
		cmdArgs = command[firstSpace+1:]
	}
	cmdVal, exists := commands[cmdName]
	if !exists {
		return nil, getAvailabelCommandsErr()
	}
	if err := cmdVal.FromArgs(cmdArgs, verbose); err != nil {
		return nil, err
	}
	return cmdVal, nil
}

type StringWrapper struct {
	Item string
}
