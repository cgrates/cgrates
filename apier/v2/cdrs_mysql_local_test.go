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
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"net/rpc"
	"net/rpc/jsonrpc"
	"path"
	"testing"
)

var testLocal = flag.Bool("local", false, "Perform the tests only on local test environment, not by default.") // This flag will be passed here via "go test -local" args
var dataDir = flag.String("data_dir", "/usr/share/cgrates", "CGR data dir path here")
var waitRater = flag.Int("wait_rater", 500, "Number of miliseconds to wait for rater to start and cache")

var cdrsCfgPath string
var cdrsCfg *config.CGRConfig
var cdrsRpc *rpc.Client

func TestInitConfig(t *testing.T) {
	if !*testLocal {
		return
	}
	var err error
	cdrsCfgPath = path.Join(*dataDir, "conf", "samples", "cdrsv2mysql_local_test.cfg")
	cdrsCfg, err = config.NewCGRConfigFromFile(&cdrsCfgPath)
	if err != nil {
		t.Fatal(err)
	}
}

func TestV2CdrsInitDataDb(t *testing.T) {
	if !*testLocal {
		return
	}
	if err := engine.InitDataDb(cdrsCfg); err != nil {
		t.Fatal(err)
	}
}

func TestV2CdrsStartEngine(t *testing.T) {
	if !*testLocal {
		return
	}
	if err := engine.StartEngine(cdrsCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func TestV2CdrsRpcConn(t *testing.T) {
	if !*testLocal {
		return
	}
	var err error
	cdrsRpc, err = jsonrpc.Dial("tcp", cdrsCfg.RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

func TestV2CdrsGetCdrs(t *testing.T) {
	if !*testLocal {
		return
	}
	var reply []*utils.StoredCdr
	req := utils.AttrGetCdrs{}
	if err := cdrsRpc.Call("ApierV2.GetCdrs", req, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} /*else if len(reply) != 2 {
		t.Error("Unexpected number of CDRs returned: ", len(reply))
	}*/
}

func TestV2CdrsCountCdrs(t *testing.T) {
	if !*testLocal {
		return
	}
	var reply int64
	req := utils.AttrGetCdrs{}
	if err := cdrsRpc.Call("ApierV2.CountCdrs", req, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} /*else if len(reply) != 2 {
		t.Error("Unexpected number of CDRs returned: ", len(reply))
	}*/
}

func TestStopEngine(t *testing.T) {
	if !*testLocal {
		return
	}
	if err := engine.StopEngine(*waitRater); err != nil {
		t.Error(err)
	}
}
