//go:build integration || flaky || offline || call || aws || race || performance

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package utils

import (
	"flag"
)

var (
	DataDir   = flag.String("data_dir", "/usr/share/cgrates", "Path to the CGR data directory.")
	WaitRater = flag.Int("wait_rater", 100, "Time (in ms) to wait for rater initialization.")
	Encoding  = flag.String("rpc", MetaJSON, "Encoding type for RPC communication (e.g., JSON).")
	DBType    = flag.String("dbtype", MetaInternal, "Type of database (Internal/Mongo/MySQL/Postgres).")
)
