/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) 2012-2015 ITsysCOM GmbH

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

package engine

import (
	"path"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

var cdrsMasterCfgPath, cdrsSlaveCfgPath string
var cdrsMasterCfg, cdrsSlaveCfg *config.CGRConfig

//var cdrsHttpJsonRpc *rpcclient.RpcClient

var waitRater = 500

func TestCdrsInitConfig(t *testing.T) {
	if !*testLocal {
		return
	}
	var err error
	cdrsMasterCfgPath = path.Join(
		*dataDir, "conf", "samples", "cdrsreplicationmaster")
	if cdrsMasterCfg, err = config.NewCGRConfigFromFolder(cdrsMasterCfgPath); err != nil {
		t.Fatal("Got config error: ", err.Error())
	}
	cdrsSlaveCfgPath = path.Join(
		*dataDir, "conf", "samples", "cdrsreplicationslave")
	if cdrsSlaveCfg, err = config.NewCGRConfigFromFolder(cdrsSlaveCfgPath); err != nil {
		t.Fatal("Got config error: ", err.Error())
	}
}

// InitDb so we can rely on count
func TestCdrsInitCdrDb(t *testing.T) {
	if !*testLocal {
		return
	}
	if err := InitStorDb(cdrsMasterCfg); err != nil {
		t.Fatal(err)
	}
	if err := InitStorDb(cdrsSlaveCfg); err != nil {
		t.Fatal(err)
	}
}

func TestCdrsStartMasterEngine(t *testing.T) {
	if !*testLocal {
		return
	}
	if _, err := StopStartEngine(cdrsMasterCfgPath, waitRater); err != nil {
		t.Fatal(err)
	}
}

func TestCdrsStartSlaveEngine(t *testing.T) {
	if !*testLocal {
		return
	}
	if _, err := StartEngine(cdrsSlaveCfgPath, waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func TestCdrsHttpJsonRpcCdrReplication(t *testing.T) {
	if !*testLocal {
		return
	}
	cdrsHttpJsonRpc, err := rpcclient.NewRpcClient(
		"tcp",
		cdrsMasterCfg.CDRSCdrReplication[0].Server,
		3,
		cdrsMasterCfg.CDRSCdrReplication[0].Transport[1:])
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
	testCdr1 := &StoredCdr{
		CgrId: utils.Sha1(
			"httpjsonrpc1",
			time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC).String()),
		TOR:         utils.VOICE,
		AccId:       "httpjsonrpc1",
		CdrHost:     "192.168.1.1",
		CdrSource:   "UNKNOWN",
		ReqType:     utils.META_PSEUDOPREPAID,
		Direction:   "*out",
		Tenant:      "cgrates.org",
		Category:    "call",
		Account:     "1001",
		Subject:     "1001",
		Destination: "1002",
		SetupTime:   time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC),
		AnswerTime:  time.Date(2013, 12, 7, 8, 42, 26, 0, time.UTC),
		Usage:       time.Duration(10) * time.Second,
		ExtraFields: map[string]string{
			"field_extr1": "val_extr1",
			"fieldextr2":  "valextr2"},
		MediationRunId: utils.DEFAULT_RUNID,
		Cost:           1.201,
		Rated:          true}
	var reply string
	if err := cdrsHttpJsonRpc.Call("CdrsV2.ProcessCdr", testCdr1, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
	var rcvedCdrs []*ExternalCdr
	if err := cdrsHttpJsonRpc.Call("ApierV2.GetCdrs", utils.RpcCdrsFilter{CgrIds: []string{testCdr1.CgrId}}, &rcvedCdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(rcvedCdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(rcvedCdrs))
	}
}
