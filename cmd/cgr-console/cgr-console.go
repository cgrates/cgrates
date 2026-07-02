/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/rpcconsole"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
	"github.com/reeflective/readline"
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
)

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
	a.repl()
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

func printKeys() {
	fmt.Print(`keys:
  Tab            complete the current word, or open a menu of matches (Tab again to cycle)
  Ctrl-F         search inside an open menu by substring
  Ctrl-Up/Down   previous/next service group in the menu
  Ctrl-R         search command history
  Enter          run the line
  Ctrl-C         clear the line. Ctrl-D on an empty line quits

Defaults. Override any of them in ~/.inputrc.
`)
}

func (a *app) repl() {
	fmt.Printf("cgr-console, connected to %s.\n", *server)
	fmt.Printf("%d methods available. Type help to list them, help keys for shortcuts, exit to quit.\n\n",
		len(a.client.Methods()))
	rl := readline.NewShell()
	setCompletionDefaults(rl)
	rl.Prompt.Primary(func() string { return "cgr> " })
	rl.Completer = a.complete
	loadHistory(rl)
	for {
		line, err := rl.Readline()
		if err != nil {
			if errors.Is(err, io.EOF) {
				fmt.Println("bye!")
				return
			}
			if errors.Is(err, readline.ErrInterrupt) {
				continue
			}
			fmt.Fprintf(os.Stderr, "readline: %v\n", err)
			continue
		}
		switch strings.ToLower(strings.TrimSpace(line)) {
		case utils.QuitCgr, utils.ExitCgr, utils.ByeCgr, utils.CloseCgr:
			fmt.Println("bye!")
			return
		default:
			a.execute(line)
		}
	}
}

// setCompletionDefaults seeds display defaults, then reloads ~/.inputrc on top
// so the user's own settings win.
func setCompletionDefaults(rl *readline.Shell) {
	defaults := map[string]any{
		"menu-complete-display-prefix": true,
		"completion-ignore-case":       true,
		"colored-completion-prefix":    true,
		"completion-query-items":       500,
	}
	for name, val := range defaults {
		_ = rl.Config.Set(name, val)
	}
	_ = rl.Keymap.ReloadConfig(rl.Opts...)
}

// loadHistory attaches a history file at ~/.cgr/history, ignoring errors so an
// unwritable home never blocks the REPL.
func loadHistory(rl *readline.Shell) {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}
	dir := filepath.Join(home, ".cgr")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return
	}
	if hist, err := readline.NewHistoryFromFile(filepath.Join(dir, "history")); err == nil {
		rl.History.Add("cgr", hist)
	}
}

// complete is reeflective's Completer entry point. DisplayList shows each match
// on its own line instead of packed columns.
func (a *app) complete(line []rune, cursor int) readline.Completions {
	return a.completions(line, cursor).DisplayList()
}

func (a *app) completions(line []rune, cursor int) readline.Completions {
	done, current := tokens(line, cursor)
	if len(done) == 0 {
		return a.completeMethods()
	}
	if done[0] == "help" || done[0] == "?" {
		if len(done) == 1 {
			return a.completeMethods()
		}
		return readline.CompleteMessage("help takes a single method name")
	}
	md := a.client.Describe(done[0])
	if md == nil {
		return readline.CompleteMessage("unknown method %s", done[0])
	}
	if key, _, isValue := strings.Cut(current, "="); isValue {
		if fd := rpcconsole.ArgField(md, key); fd != nil && len(fd.Fields) != 0 {
			return completeNested(*fd)
		}
		return completeValue(md, key)
	}
	return completeFields(md, typedKeys(done[1:]))
}

// completeMethods offers every method grouped under its service, with the alias
// as the value and the RPC name as its description.
func (a *app) completeMethods() readline.Completions {
	methods := a.client.Methods()
	comps := make([]readline.Completion, 0, len(methods))
	for _, m := range methods {
		alias := rpcconsole.Alias(m)
		svc, _, _ := strings.Cut(alias, ".")
		comps = append(comps, readline.Completion{Value: alias, Description: m, Tag: svc})
	}
	return readline.CompleteRaw(comps)
}

// completeFields offers the argument keys not already on the line, each with its
// type. Which keys are required is the handler's call, not derivable here.
func completeFields(md *utils.MethodDescriptor, typed map[string]struct{}) readline.Completions {
	comps := make([]readline.Completion, 0, len(md.Args))
	for _, f := range md.Args {
		if _, ok := typed[f.Name]; ok {
			continue
		}
		comps = append(comps, readline.Completion{
			Value:       f.Name + "=",
			Description: f.Type,
		})
	}
	if len(comps) == 0 {
		return readline.CompleteMessage("all arguments set for %s", md.Method)
	}
	return readline.CompleteRaw(comps).
		Usage("%s", rpcconsole.ArgSignature(md)).
		NoSpace('=').
		JustifyDescriptions()
}

// completeValue offers values for the field being typed.
func completeValue(md *utils.MethodDescriptor, key string) readline.Completions {
	typ := rpcconsole.FieldType(md, key)
	switch typ {
	case "bool":
		return readline.CompleteValues(key+"=true", key+"=false").Tag(key + " values")
	case "duration":
		return readline.CompleteMessage("%s = <duration>, e.g. 1s, 100ms, 1h30m", key)
	case "":
		return readline.CompleteMessage("%s has no argument %q", md.Method, key)
	default:
		return readline.CompleteMessage("%s = <%s>", key, typ)
	}
}

// completeNested hints a value with inner fields, listing them one level deep.
func completeNested(fd utils.FieldDescriptor) readline.Completions {
	return readline.CompleteMessage("%s = <%s>", fd.Name, fd.Type).
		Usage("%s", rpcconsole.InnerFields(fd))
}

// tokens splits the text before the cursor into finished tokens and the partial
// word under it. It tokenizes like rpcconsole.BuildParams so completion and
// execution agree on token boundaries.
func tokens(line []rune, cursor int) (done []string, current string) {
	prefix := string(line[:cursor])
	fields := rpcconsole.SplitArgs(prefix)
	if len(fields) > 0 && !endsWithSpace(prefix) {
		return fields[:len(fields)-1], fields[len(fields)-1]
	}
	return fields, ""
}

func endsWithSpace(s string) bool {
	return strings.HasSuffix(s, " ") || strings.HasSuffix(s, "\t")
}

// typedKeys returns the argument keys already present on the line.
func typedKeys(args []string) map[string]struct{} {
	keys := make(map[string]struct{}, len(args))
	for _, a := range args {
		if k, _, ok := strings.Cut(a, "="); ok {
			keys[k] = struct{}{}
		}
	}
	return keys
}
