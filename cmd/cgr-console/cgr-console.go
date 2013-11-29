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

	"github.com/cgrates/cgrates/console"
	"github.com/cgrates/cgrates/utils"

	"log"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os"
)

var (
	version      = flag.Bool("version", false, "Prints the application version.")
	server       = flag.String("server", "127.0.0.1:2012", "server address host:port")
	rpc_encoding = flag.String("rpc_encoding", "json", "RPC encoding used <gob|json>")
)

func main() {
	flag.Parse()
	if *version {
		fmt.Println("CGRateS " + utils.VERSION)
		return
	}
	var client *rpc.Client
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
	// Strict command parsing starts here
	args := append([]string{os.Args[0]}, flag.Args()...) // Emulate os.Args by prepending the cmd to list of args coming from flag
	cmd, cmdErr := console.GetCommandValue(args)
	if cmdErr != nil {
		log.Fatal(cmdErr)
	}
	res := cmd.RpcResult()
	if rpcErr := client.Call(cmd.RpcMethod(), cmd.RpcParams(), res); rpcErr != nil {
		fmt.Println("Error executing command: " + rpcErr.Error())
	}
	result, _ := json.MarshalIndent(res, "", "  ")
	fmt.Println(string(result))

}
