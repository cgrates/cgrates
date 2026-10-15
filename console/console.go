// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/rpcconsole"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

type app struct {
	client  *rpcconsole.Client
	verbose bool
}

// Run runs a command or starts the interactive console.
// args starts with the program name, whose base name is used in help and titles.
// prefix is used for the prompt and ~/.<prefix>/history.
//
// Flag and command errors are printed. Errors while connecting to the server,
// getting its methods or running the menu are returned.
func Run(args []string, prefix string) error {
	name := filepath.Base(args[0])
	flags := flag.NewFlagSet(name, flag.ContinueOnError)
	version := flags.Bool(utils.VersionCgr, false, "Prints the application version.")
	verbose := flags.Bool(utils.VerboseCgr, false, "Show extra info about command execution.")
	server := flags.String(utils.MailerServerCfg, "127.0.0.1:2012", "server address host:port")
	certificatePath := flags.String(utils.CertPathCgr, "", "path to certificate for tls connection")
	keyPath := flags.String(utils.KeyPathCgr, "", "path to key for tls connection")
	caPath := flags.String(utils.CAPathCgr, "", "path to CA for tls connection(only for self sign certificate)")
	tls := flags.Bool(utils.TLSNoCaps, false, "TLS connection")
	replyTimeout := flags.Int(utils.ReplyTimeoutCfg, 300, "Reply timeout in seconds")
	var menuMode bool
	flags.BoolVar(&menuMode, "menu", false, "Use the tui interface")
	flags.BoolVar(&menuMode, "m", false, "shorthand for -menu")
	if err := flags.Parse(args[1:]); err != nil {
		// Keep the zero exit status after help or a flag error.
		return nil
	}
	if *version {
		if version, err := utils.GetCGRVersion(); err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(version)
		}
		return nil
	}

	cl, err := rpcclient.NewRPCClient(context.Background(), utils.TCP, *server, *tls,
		*keyPath, *certificatePath, *caPath, 3, 3, 0, utils.FibDuration, time.Second,
		time.Duration(*replyTimeout)*time.Second, utils.MetaJSON, nil, false, nil)
	if err != nil {
		flags.PrintDefaults()
		return fmt.Errorf("Could not connect to server %s", *server)
	}
	client, err := rpcconsole.NewClient(cl)
	if err != nil {
		return fmt.Errorf("Could not fetch methods: %w", err)
	}
	a := &app{client: client, verbose: *verbose}

	if args := flags.Args(); len(args) != 0 {
		a.execute(strings.Join(args, " "))
		return nil
	}
	if menuMode {
		return runTUI(client, name, *server)
	}
	a.repl(name, prefix, *server)
	return nil
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
