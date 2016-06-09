/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can Storagetribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITH*out ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package v2

import (
	"flag"
	"net/rpc"
	"net/rpc/jsonrpc"
	"path"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var testIT = flag.Bool("integration", false, "Perform the tests in integration mode, not by default.")
var storDBType = flag.String("stordb_type", utils.MYSQL, "Perform the tests for MongoDB, not by default.")

var tpCfgPath string
var tpCfg *config.CGRConfig
var tpRPC *rpc.Client

func TestTPitLoadConfig(t *testing.T) {
	if !*testIT {
		return
	}
	var err error
	switch *storDBType {
	case utils.MYSQL:
		tpCfgPath = path.Join(*dataDir, "conf", "samples", "tutmysql")
	case utils.POSTGRES:
		tpCfgPath = path.Join(*dataDir, "conf", "samples", "tutpostgres")
	case utils.MONGO:
		tpCfgPath = path.Join(*dataDir, "conf", "samples", "tutmongo")
	default:
		t.Fatalf("Unsupported stordb_type: %s", *storDBType)
	}
	if tpCfg, err = config.NewCGRConfigFromFolder(tpCfgPath); err != nil {
		t.Error(err)
	}
}

// Remove data in both rating and accounting db
func TestTPitResetDataDb(t *testing.T) {
	if !*testIT {
		return
	}
	if err := engine.InitDataDb(tpCfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func TestTPitResetStorDb(t *testing.T) {
	if !*testIT {
		return
	}
	if err := engine.InitStorDb(tpCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func TestTPitStartEngine(t *testing.T) {
	if !*testIT {
		return
	}
	if _, err := engine.StopStartEngine(tpCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func TestTPitRpcConn(t *testing.T) {
	if !*testIT {
		return
	}
	var err error
	tpRPC, err = jsonrpc.Dial("tcp", tpCfg.RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}
