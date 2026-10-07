// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"log"
	"os"

	"github.com/cgrates/cgrates/services"
)

func main() {
	if err := services.RunCGREngine(os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}
