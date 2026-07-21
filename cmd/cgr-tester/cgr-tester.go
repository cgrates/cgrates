// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/cgrates/cgrates/utils"
)

var (
	cgrTesterFlags = flag.NewFlagSet("cgr-tester", flag.ContinueOnError)
	version        = cgrTesterFlags.Bool("version", false, "Prints the application version.")
)

func main() {
	if err := cgrTesterFlags.Parse(os.Args[1:]); err != nil {
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
}
