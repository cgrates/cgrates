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
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/console"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
	"github.com/peterh/liner"
)

var (
	cgrConsoleFlags = flag.NewFlagSet(utils.CgrConsole, flag.ContinueOnError)
	historyFN       = os.Getenv(utils.HomeCgr) + utils.HistoryCgr
	version         = cgrConsoleFlags.Bool(utils.ElsVersionLow, false, "Prints the application version.")
	verbose         = cgrConsoleFlags.Bool(utils.VerboseCgr, false, "Show extra info about command execution.")
	server          = cgrConsoleFlags.String(utils.MailerServerCfg, "127.0.0.1:2012", "server address host:port")
	rpcEncoding     = cgrConsoleFlags.String(utils.RpcEncodingCgr, utils.MetaJSON, "RPC encoding used <*gob|*json>")
	certificatePath = cgrConsoleFlags.String(utils.CertPathCgr, utils.EmptyString, "path to certificate for tls connection")
	keyPath         = cgrConsoleFlags.String(utils.KeyPathCgr, utils.EmptyString, "path to key for tls connection")
	caPath          = cgrConsoleFlags.String(utils.CAPathCgr, utils.EmptyString, "path to CA for tls connection(only for self sign certificate)")
	tls             = cgrConsoleFlags.Bool(utils.TLSNoCaps, false, "TLS connection")
	replyTimeOut    = cgrConsoleFlags.Int(utils.ReplyTimeoutCfg, 300, "Reply timeout in seconds ")
	client          *rpcclient.RPCClient
)

func executeCommand(command string) {
	if strings.TrimSpace(command) == utils.EmptyString {
		return
	}
	if strings.TrimSpace(command) == utils.HelpCgr {
		commands := console.GetCommands()
		orderedKeys := make([]string, len(commands))
		fmt.Println("Commands:")
		for name := range commands {
			if name != utils.EmptyString {
				orderedKeys = append(orderedKeys, name)
			}
		}
		sort.Strings(orderedKeys)
		for _, name := range orderedKeys {
			if commands[name] == nil {
				continue
			}
			fmt.Println(name, commands[name].Usage())
		}
		return
	}
	if strings.HasPrefix(command, utils.HelpCgr) {
		words := strings.Split(command, utils.SepCgr)
		if len(words) > 1 {
			commands := console.GetCommands()
			if cmd, ok := commands[words[1]]; ok {
				fmt.Print(cmd.Usage())
			} else {
				fmt.Print("Available commands: ")
				for name := range commands {
					fmt.Print(name + utils.SepCgr)
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
	if cmd.RpcMethod() != utils.EmptyString {
		res := cmd.RpcResult()
		param := cmd.RpcParams(false)
		switch param.(type) {
		case *console.EmptyWrapper:
			param = utils.EmptyString
		case *console.StringWrapper:
			param = param.(*console.StringWrapper).Item
		case *console.StringSliceWrapper:
			param = param.(*console.StringSliceWrapper).Items
		case *console.StringMapWrapper:
			param = param.(*console.StringMapWrapper).Items
		}

		if rpcErr := client.Call(context.TODO(), cmd.RpcMethod(), param, res); rpcErr != nil {
			fmt.Println("Error executing command: " + rpcErr.Error())
		} else {
			fmt.Println(cmd.GetFormatedResult(res))
		}
	} else {
		fmt.Println(cmd.LocalExecute())
	}
}

func main() {
	if err := cgrConsoleFlags.Parse(os.Args[1:]); err != nil {
		return
	}
	if *version {
		if rcv, err := utils.GetCGRVersion(); err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(rcv)
		}
		return
	}
	var err error

	client, err = rpcclient.NewRPCClient(context.Background(), utils.TCP, *server, *tls, *keyPath, *certificatePath, *caPath, 3, 3,
		time.Second, time.Duration(*replyTimeOut)*time.Second, *rpcEncoding, nil, false, nil)
	if err != nil {
		cgrConsoleFlags.PrintDefaults()
		log.Fatal("Could not connect to server " + *server)
	}

	if len(cgrConsoleFlags.Args()) != 0 {
		executeCommand(strings.Join(cgrConsoleFlags.Args(), utils.SepCgr))
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
				lastSpace := strings.LastIndex(line, utils.SepCgr)
				lastSpace++
				for _, arg := range cmd.ClientArgs() {
					if strings.HasPrefix(arg, line[lastSpace:]) {
						comp = append(comp, line[:lastSpace]+arg)
					}
				}
			}
		}
		return
	})

	if f, err := os.Open(historyFN); err == nil {
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
			case utils.QuitCgr, utils.ExitCgr, utils.ByeCgr, utils.CloseCgr:
				fmt.Println("\nbye!")
				stop = true
			default:
				executeCommand(command)
			}
		}
	}

	if f, err := os.Create(historyFN); err != nil {
		log.Print("Error writing history file: ", err)
	} else {
		line.WriteHistory(f)
		f.Close()
	}
}
