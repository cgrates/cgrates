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
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"net/rpc"
	"net/rpc/jsonrpc"
	"path"
	"testing"
)

var cdrsPsqlCfgPath string
var cdrsPsqlCfg *config.CGRConfig
var cdrsPsqlRpc *rpc.Client

func TestV2CdrsPsqlInitConfig(t *testing.T) {
	if !*testLocal {
		return
	}
	var err error
	cdrsPsqlCfgPath = path.Join(*dataDir, "conf", "samples", "cdrsv2psql_local_test.cfg")
	cdrsPsqlCfg, err = config.NewCGRConfigFromFile(&cdrsPsqlCfgPath)
	if err != nil {
		t.Fatal(err)
	}
}

func TestV2CdrsPsqlInitDataDb(t *testing.T) {
	if !*testLocal {
		return
	}
	if err := engine.InitDataDb(cdrsPsqlCfg); err != nil {
		t.Fatal(err)
	}
}

func TestV2CdrsPsqlStartEngine(t *testing.T) {
	if !*testLocal {
		return
	}
	if err := engine.StartEngine(cdrsPsqlCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func TestV2CdrsPsqlPsqlRpcConn(t *testing.T) {
	if !*testLocal {
		return
	}
	var err error
	cdrsPsqlRpc, err = jsonrpc.Dial("tcp", cdrsPsqlCfg.RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

func TestV2CdrsPsqlGetCdrs(t *testing.T) {
	if !*testLocal {
		return
	}
	var reply []*utils.StoredCdr
	req := utils.AttrGetCdrs{}
	if err := cdrsPsqlRpc.Call("ApierV2.GetCdrs", req, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} /*else if len(reply) != 2 {
		t.Error("Unexpected number of CDRs returned: ", len(reply))
	}*/
}

func TestV2CdrsPsqlCountCdrs(t *testing.T) {
	if !*testLocal {
		return
	}
	var reply int64
	req := utils.AttrGetCdrs{}
	if err := cdrsPsqlRpc.Call("ApierV2.CountCdrs", req, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} /*else if len(reply) != 2 {
		t.Error("Unexpected number of CDRs returned: ", len(reply))
	}*/
}

func TestV2CdrsPsqlStopEngine(t *testing.T) {
	if !*testLocal {
		return
	}
	if err := engine.StopEngine(*waitRater); err != nil {
		t.Error(err)
	}
}
