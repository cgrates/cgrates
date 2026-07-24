// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
	RpcParams(bool) any                       // Parameters to send out on rpc
	PostprocessRpcParams() error              // Corrects rpc parameters when needed
	RpcResult() any                           // Only requirement is to have a String method to print on console
	ClientArgs() []string                     // for autocompletion
	Name() string
	LocalExecute() string
	GetFormatedResult(result any) string
}

func GetCommands() map[string]Commander {
	return commands
}

func getAvailabelCommandsErr() error {
	var keys []string
	for key := range commands {
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
		cmdName = command[:]
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
	if err := cmdVal.PostprocessRpcParams(); err != nil {
		return nil, err
	}
	return cmdVal, nil
}

type StringWrapper struct {
	Item string
}

type StringSliceWrapper struct {
	Items []string
}

type StringMapWrapper struct {
	Items map[string]string
}

type EmptyWrapper struct{}
