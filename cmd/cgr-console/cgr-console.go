// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/rpcconsole"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

var (
	consoleFlags    = flag.NewFlagSet("cgr-console", flag.ContinueOnError)
	version         = consoleFlags.Bool(utils.VersionCgr, false, "Prints the application version.")
	verbose         = consoleFlags.Bool(utils.VerboseCgr, false, "Show extra info about command execution.")
	server          = consoleFlags.String(utils.MailerServerCfg, "127.0.0.1:2012", "server address host:port")
	certificatePath = consoleFlags.String(utils.CertPathCgr, "", "path to certificate for tls connection")
	keyPath         = consoleFlags.String(utils.KeyPathCgr, "", "path to key for tls connection")
	caPath          = consoleFlags.String(utils.CAPathCgr, "", "path to CA for tls connection(only for self sign certificate)")
	tls             = consoleFlags.Bool(utils.TLSNoCaps, false, "TLS connection")
	replyTimeout    = consoleFlags.Int(utils.ReplyTimeoutCfg, 300, "Reply timeout in seconds")
	simpleMode      bool
)

func init() {
	consoleFlags.BoolVar(&simpleMode, "simple", false, "Use the readline interface")
	consoleFlags.BoolVar(&simpleMode, "s", false, "shorthand for -simple")

}

type app struct {
	client  *rpcconsole.Client
	verbose bool
}

func main() {
	if err := consoleFlags.Parse(os.Args[1:]); err != nil {
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

	cl, err := rpcclient.NewRPCClient(context.Background(), utils.TCP, *server, *tls,
		*keyPath, *certificatePath, *caPath, 3, 3, 0, utils.FibDuration, time.Second,
		time.Duration(*replyTimeout)*time.Second, utils.MetaJSON, nil, false, nil)
	if err != nil {
		consoleFlags.PrintDefaults()
		log.Fatal("Could not connect to server " + *server)
	}
	client, err := rpcconsole.NewClient(cl)
	if err != nil {
		log.Fatal("Could not fetch methods: " + err.Error())
	}
	a := &app{client: client, verbose: *verbose}

	if args := consoleFlags.Args(); len(args) != 0 {
		a.execute(strings.Join(args, " "))
		return
	}
	if !simpleMode && isTTY(os.Stdin) && isTTY(os.Stdout) {
		if err := runTUI(client, *server); err != nil {
			log.Fatal(err)
		}
		return
	}
	a.repl()
}

// isTTY reports whether f is an interactive terminal, so pipes and scripts
// fall back to the readline shell instead of the TUI.
func isTTY(f *os.File) bool {
	fi, err := f.Stat()
	return err == nil && fi.Mode()&os.ModeCharDevice != 0
}

func (a *app) execute(input string) {
	input = strings.TrimSpace(input)
	if input == "" {
		return
	}
	fields := strings.Fields(input)
	method := fields[0]
	if method == "help" || method == "?" {
		a.help(fields)
		return
	}
	md := a.client.Describe(method)
	params, err := rpcconsole.BuildParams(strings.TrimSpace(strings.TrimPrefix(input, method)), md)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not parse arguments: %v\n", err)
		return
	}
	if md != nil {
		for k := range params {
			if rpcconsole.FieldType(md, k) == "" {
				fmt.Fprintf(os.Stderr, "warning: %s has no argument %q\n", method, k)
			}
		}
	}
	if a.verbose {
		jsn, _ := json.Marshal(params)
		fmt.Fprintf(os.Stderr, "> %s %s\n", method, jsn)
	}
	reply, err := a.client.Call(method, params)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return
	}
	fmt.Println(rpcconsole.Format(reply, md))
}

func (a *app) help(fields []string) {
	switch {
	case len(fields) == 1:
		for _, m := range a.client.Methods() {
			fmt.Println(rpcconsole.Alias(m))
		}
	case fields[1] == "keys":
		printKeys()
	default:
		md := a.client.Describe(fields[1])
		if md == nil {
			fmt.Fprintf(os.Stderr, "unknown method: %s\n", fields[1])
			return
		}
		fmt.Println(rpcconsole.ArgSignature(md))
		fmt.Print(rpcconsole.ArgTree(md))
	}
}
