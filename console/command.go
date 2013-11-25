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
	"errors"
	"fmt"
	"strings"
)

var (
	commands = make(map[string]Commander)
)

// Console Command interface
type Commander interface {
	FromArgs(args []string) error // Load data from os arguments or flag.Args()
	Usage(string) string          // usage message
	RpcMethod() string            // Method which should be called remotely
	RpcParams() interface{}       // Parameters to send out on rpc
	RpcResult() interface{}       // Only requirement is to have a String method to print on console
	defaults() error              // set defaults wherever necessary
}

// Process args and return right command Value or error
func GetCommandValue(args []string) (Commander, error) {
	if len(args) < 2 {
		return nil, errors.New("\n\tUsage: cgr-console [cfg_opts...{-h}] <command>\n")
	}
	cmdVal, exists := commands[args[1]]
	if !exists {
		var keys []string
		for key, _ := range commands {
			keys = append(keys, key)
		}
		return nil, fmt.Errorf("\n\tUsage: cgr-console [cfg_opts...{-h}] <%s>\n", strings.Join(keys, "|"))
	}
	if err := cmdVal.FromArgs(args); err != nil {
		return nil, err
	}
	return cmdVal, nil
}
