// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/cgrates/cgrates/rpcconsole"
	"github.com/cgrates/cgrates/utils"
	"github.com/reeflective/readline"
)

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
