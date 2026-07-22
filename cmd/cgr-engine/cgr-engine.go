// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"os"

	"github.com/cgrates/cgrates/services"
)

func main() {
	services.RunCGREngine(os.Args[1:])
}
