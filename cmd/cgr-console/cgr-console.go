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

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os"
	"strings"

	"github.com/cgrates/cgrates/console"
	"github.com/cgrates/cgrates/utils"
	"github.com/peterh/liner"
)

var (
	history_fn   = os.Getenv("HOME") + "/.cgr_history"
	version      = flag.Bool("version", false, "Prints the application version.")
	verbose      = flag.Bool("verbose", false, "Show extra info about command execution.")
	server       = flag.String("server", "127.0.0.1:2012", "server address host:port")
	rpc_encoding = flag.String("rpc_encoding", "json", "RPC encoding used <gob|json>")
	client       *rpc.Client
)

func executeCommand(command string) {
	cmd, cmdErr := console.GetCommandValue(command, *verbose)
	if cmdErr != nil {
		fmt.Println(cmdErr)
		return
	}
	res := cmd.RpcResult()
	if rpcErr := client.Call(cmd.RpcMethod(), cmd.RpcParams(), res); rpcErr != nil {
		fmt.Println("Error executing command: " + rpcErr.Error())
	}
	result, _ := json.MarshalIndent(res, "", " ")
	fmt.Println(string(result))
}

func main() {
	flag.Parse()
	if *version {
		fmt.Println("CGRateS " + utils.VERSION)
		return
	}

	var err error
	if *rpc_encoding == "json" {
		client, err = jsonrpc.Dial("tcp", *server)
	} else {
		client, err = rpc.Dial("tcp", *server)
	}
	if err != nil {
		flag.PrintDefaults()
		log.Fatal("Could not connect to server " + *server)
	}
	defer client.Close()

	if len(flag.Args()) != 0 {
		executeCommand(strings.Join(flag.Args(), " "))
		return
	}

	fmt.Println("Welcome to CGRateS console!")
	fmt.Println("Type `help` for a list of commands\n")

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
		fmt.Println("Error writing history file: ", err)
	} else {
		log.Print("Writing history: ", history_fn)
		line.WriteHistory(f)
		f.Close()
	}
}
