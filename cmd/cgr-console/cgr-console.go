// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"log"
	"os"

	"github.com/cgrates/cgrates/console"
)

func main() {
	if err := console.Run(os.Args, "cgr"); err != nil {
		log.Fatal(err)
	}
}
