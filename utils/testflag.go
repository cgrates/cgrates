//go:build integration || flaky || call || performance

/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

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
