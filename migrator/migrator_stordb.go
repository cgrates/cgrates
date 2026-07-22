// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package migrator

import (
	"github.com/cgrates/cgrates/engine"
)

type MigratorStorDB interface {
	getV1CDR() (v1Cdr *v1Cdrs, err error)
	setV1CDR(v1Cdr *v1Cdrs) (err error)
	remV1CDRs(v1Cdr *v1Cdrs) (err error)
	createV1SMCosts() (err error)
	renameV1SMCosts() (err error)
	getV2SMCost() (v2Cost *v2SessionsCost, err error)
	setV2SMCost(v2Cost *v2SessionsCost) (err error)
	remV2SMCost(v2Cost *v2SessionsCost) (err error)
	StorDB() engine.StorDB
	close()
}
