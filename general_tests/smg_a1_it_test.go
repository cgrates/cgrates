// +build integration

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
package general_tests

import (
	"net/rpc"
	"net/rpc/jsonrpc"
	"path"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	smgA1CfgPath string
	smgA1Cfg     *config.CGRConfig
	smgA1rpc     *rpc.Client
)

func TestSMGa1ITLoadConfig(t *testing.T) {
	smgA1CfgPath = path.Join(*dataDir, "conf", "samples", "tutmongo")
	if smgA1Cfg, err = config.NewCGRConfigFromFolder(tpCfgPath); err != nil {
		t.Error(err)
	}
}

func TestSMGa1ITResetDataDB(t *testing.T) {
	if err := engine.InitDataDb(smgA1Cfg); err != nil {
		t.Fatal(err)
	}
}

func TestSMGa1ITResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(smgA1Cfg); err != nil {
		t.Fatal(err)
	}
}

func TestSMGa1ITStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(smgA1CfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func TestSMGa1ITRPCConn(t *testing.T) {
	var err error
	smgA1rpc, err = jsonrpc.Dial("tcp", smgA1Cfg.RPCJSONListen)
	if err != nil {
		t.Fatal(err)
	}
}

func TestSMGa1ITLoadTPFromFolder(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "test", "smg_a1")}
	if err := smgA1rpc.Call("ApierV1.LoadTariffPlanFromFolder", attrs, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(reply)
	}
	time.Sleep(time.Duration(100 * time.Millisecond))
	tStart, _ := utils.ParseDate("2017-03-03T10:39:33Z")
	tEnd, _ := utils.ParseDate("2017-03-03T12:30:13Z") // Equivalent of 10240 which is a chunk of data charged
	cd := engine.CallDescriptor{
		Direction:   "*out",
		Category:    "data1",
		Tenant:      "cgrates.org",
		Subject:     "subj_rpdata1",
		Destination: "data",
		TimeStart:   tStart,
		TimeEnd:     tEnd,
	}
	var cc engine.CallCost
	if err := smgA1rpc.Call("Responder.GetCost", cd, &cc); err != nil {
		t.Error("Got error on Responder.GetCost: ", err.Error())
	} else if cc.Cost != 0.0 || cc.RatedUsage != 10240 {
		t.Errorf("Calling Responder.GetCost got callcost: %v", cc)
	}
}
