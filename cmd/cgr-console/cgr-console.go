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

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/cgrates/cgrates/console"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
	"github.com/peterh/liner"
)

var (
	history_fn   = os.Getenv("HOME") + "/.cgr_history"
	version      = flag.Bool("version", false, "Prints the application version.")
	verbose      = flag.Bool("verbose", false, "Show extra info about command execution.")
	server       = flag.String("server", "127.0.0.1:2012", "server address host:port")
	rpc_encoding = flag.String("rpc_encoding", "json", "RPC encoding used <gob|json>")
	client       *rpcclient.RpcClient
)

func executeCommand(command string) {
	if strings.TrimSpace(command) == "" {
		return
	}
	if strings.TrimSpace(command) == "help" {
		commands := console.GetCommands()
		fmt.Println("Commands:")
		for name, cmd := range commands {
			fmt.Print(name, cmd.Usage())
		}
		return
	}
	if strings.HasPrefix(command, "help") {
		words := strings.Split(command, " ")
		if len(words) > 1 {
			commands := console.GetCommands()
			if cmd, ok := commands[words[1]]; ok {
				fmt.Print(cmd.Usage())
			} else {
				fmt.Print("Available commands: ")
				for name, _ := range commands {
					fmt.Print(name + " ")
				}
				fmt.Println()
			}
			return
		}
	}
	cmd, cmdErr := console.GetCommandValue(command, *verbose)
	if cmdErr != nil {
		fmt.Println(cmdErr)
		return
	}
	if cmd.RpcMethod() != "" {
		res := cmd.RpcResult()
		param := cmd.RpcParams(false)
		//log.Print(reflect.TypeOf(param))
		switch param.(type) {
		case *console.EmptyWrapper:
			param = ""
		case *console.StringWrapper:
			param = param.(*console.StringWrapper).Item
		case *console.StringSliceWrapper:
			param = param.(*console.StringSliceWrapper).Items
		case *console.StringMapWrapper:
			param = param.(*console.StringMapWrapper).Items
		}
		//log.Printf("Param: %+v", param)

		if rpcErr := client.Call(cmd.RpcMethod(), param, res); rpcErr != nil {
			fmt.Println("Error executing command: " + rpcErr.Error())
		} else {
			result, _ := json.MarshalIndent(res, "", " ")
			fmt.Println(string(result))
		}
	} else {
		fmt.Println(cmd.LocalExecute())
	}
}

func main() {
	flag.Parse()
	if *version {
		fmt.Println(utils.GetCGRVersion())
		return
	}
	var err error
	client, err = rpcclient.NewRpcClient("tcp", *server, 3, 3,
		time.Duration(1*time.Second), time.Duration(5*time.Minute), *rpc_encoding, nil, false)
	if err != nil {
		flag.PrintDefaults()
		log.Fatal("Could not connect to server " + *server)
	}

	if len(flag.Args()) != 0 {
		executeCommand(strings.Join(flag.Args(), " "))
		return
	}

	fmt.Println("Welcome to CGRateS console!")
	fmt.Print("Type `help` for a list of commands\n\n")

	line := liner.NewLiner()
	defer line.Close()

	line.SetCompleter(func(line string) (comp []string) {
		commands := console.GetCommands()
		for name, cmd := range commands {
			if strings.HasPrefix(name, strings.ToLower(line)) {
				comp = append(comp, name)
			}
			// try arguments
			if strings.HasPrefix(line, name) {
				// get last word
				lastSpace := strings.LastIndex(line, " ")
				lastSpace += 1
				for _, arg := range cmd.ClientArgs() {
					if strings.HasPrefix(arg, line[lastSpace:]) {
						comp = append(comp, line[:lastSpace]+arg)
					}
				}
			}
		}
		return
	})

	if f, err := os.Open(history_fn); err == nil {
		line.ReadHistory(f)
		f.Close()
	}

	stop := false
	for !stop {
		if command, err := line.Prompt("cgr> "); err != nil {
			if err == io.EOF {
				fmt.Println("\nbye!")
				stop = true
			} else {
				fmt.Print("Error reading line: ", err)
			}
		} else {
			line.AppendHistory(command)
			switch strings.ToLower(strings.TrimSpace(command)) {
			case "quit", "exit", "bye", "close":
				fmt.Println("\nbye!")
				stop = true
			default:
				executeCommand(command)
			}
		}
	}

	if f, err := os.Create(history_fn); err != nil {
		log.Print("Error writing history file: ", err)
	} else {
		line.WriteHistory(f)
		f.Close()
	}
}
