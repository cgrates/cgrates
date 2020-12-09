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

package analyzers

import (
	"errors"
	"flag"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os"
	"path"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	anzCfgPath string
	anzCfg     *config.CGRConfig
	anzRPC     *rpc.Client

	sTestsAlsPrf = []func(t *testing.T){
		testAnalyzerSInitCfg,
		testAnalyzerSInitDataDb,
		testAnalyzerSResetStorDb,
		testAnalyzerSStartEngine,
		testAnalyzerSRPCConn,
		testAnalyzerSLoadTarrifPlans,
		testAnalyzerSChargerSv1ProcessEvent,
		testAnalyzerSV1Search,
		testAnalyzerSV1Search2,
		testAnalyzerSV1SearchWithContentFilters,
		testAnalyzerSKillEngine,
	}
)

var (
	dataDir   = flag.String("data_dir", "/usr/share/cgrates", "CGR data dir path here")
	waitRater = flag.Int("wait_rater", 100, "Number of miliseconds to wait for rater to start and cache")
	encoding  = flag.String("rpc", utils.MetaJSON, "what encoding whould be uused for rpc comunication")
)

func newRPCClient(cfg *config.ListenCfg) (c *rpc.Client, err error) {
	switch *encoding {
	case utils.MetaJSON:
		return jsonrpc.Dial(utils.TCP, cfg.RPCJSONListen)
	case utils.MetaGOB:
		return rpc.Dial(utils.TCP, cfg.RPCGOBListen)
	default:
		return nil, errors.New("UNSUPPORTED_RPC")
	}
}

//Test start here
func TestAnalyzerSIT(t *testing.T) {
	for _, stest := range sTestsAlsPrf {
		t.Run("TestAnalyzerSIT", stest)
	}
}

func testAnalyzerSInitCfg(t *testing.T) {
	var err error
	if err := os.RemoveAll("/tmp/analyzers/"); err != nil {
		t.Fatal(err)
	}
	if err = os.MkdirAll("/tmp/analyzers/", 0700); err != nil {
		t.Fatal(err)
	}
	anzCfgPath = path.Join(*dataDir, "conf", "samples", "analyzers")
	anzCfg, err = config.NewCGRConfigFromPath(anzCfgPath)
	if err != nil {
		t.Error(err)
	}
}

func testAnalyzerSInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(anzCfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func testAnalyzerSResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(anzCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testAnalyzerSStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(anzCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testAnalyzerSRPCConn(t *testing.T) {
	var err error
	anzRPC, err = newRPCClient(anzCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testAnalyzerSLoadTarrifPlans(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "tutorial")}
	if err := anzRPC.Call(utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	time.Sleep(100 * time.Millisecond)
}

func testAnalyzerSChargerSv1ProcessEvent(t *testing.T) {
	cgrEv := &utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "event1",
			Event: map[string]interface{}{
				utils.Account:     "1010",
				utils.Subject:     "Something_inter",
				utils.Destination: "999",
			},
		},
	}
	var result2 []*engine.ChrgSProcessEventReply

	processedEv := []*engine.ChrgSProcessEventReply{
		{
			ChargerSProfile: "DEFAULT",
			AlteredFields:   []string{"*req.RunID"},
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "event1",
				Event: map[string]interface{}{
					"Account":     "1010",
					"Destination": "999",
					"RunID":       "*default",
					"Subject":     "Something_inter",
				},
			},
			Opts: map[string]interface{}{"*subsys": "*chargers"},
		},
		{
			ChargerSProfile:    "Raw",
			AttributeSProfiles: []string{"*constant:*req.RequestType:*none"},
			AlteredFields:      []string{"*req.RunID", "*req.RequestType"},
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "event1",
				Event: map[string]interface{}{
					"Account":     "1010",
					"Destination": "999",
					"RequestType": "*none",
					"RunID":       "*raw",
					"Subject":     "Something_inter",
				},
			},
			Opts: map[string]interface{}{"*subsys": "*chargers"},
		},
	}

	if err := anzRPC.Call(utils.ChargerSv1ProcessEvent, cgrEv, &result2); err != nil {
		t.Fatal(err)
	}
	sort.Slice(result2, func(i, j int) bool {
		return result2[i].ChargerSProfile < result2[j].ChargerSProfile
	})
	if !reflect.DeepEqual(result2, processedEv) {
		t.Errorf("Expecting : %s, \n received: %s", utils.ToJSON(processedEv), utils.ToJSON(result2))
	}

}

func testAnalyzerSV1Search(t *testing.T) {
	var result []map[string]interface{}
	if err := anzRPC.Call(utils.AnalyzerSv1StringQuery, &QueryArgs{HeaderFilters: `+RequestEncoding:\*internal +RequestMethod:AttributeSv1\.ProcessEvent`}, &result); err != nil {
		t.Error(err)
	} else if len(result) != 1 {
		t.Errorf("Unexpected result: %s", utils.ToJSON(result))
	}
}

func testAnalyzerSV1Search2(t *testing.T) {
	var result []map[string]interface{}
	if err := anzRPC.Call(utils.AnalyzerSv1StringQuery, &QueryArgs{HeaderFilters: `+RequestEncoding:\*json +RequestMethod:ChargerSv1\.ProcessEvent`}, &result); err != nil {
		t.Error(err)
	} else if len(result) != 1 {
		t.Errorf("Unexpected result: %s", utils.ToJSON(result))
	}
}

func testAnalyzerSV1SearchWithContentFilters(t *testing.T) {
	var result []map[string]interface{}
	if err := anzRPC.Call(utils.AnalyzerSv1StringQuery, &QueryArgs{
		HeaderFilters:  `+RequestEncoding:\*json`,
		ContentFilters: []string{"*string:~*req.Event.Account:1010"},
	}, &result); err != nil {
		t.Error(err)
	} else if len(result) != 1 {
		t.Errorf("Unexpected result: %s", utils.ToJSON(result))
	}
}

func testAnalyzerSKillEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
	if err := os.RemoveAll(anzCfg.AnalyzerSCfg().DBPath); err != nil {
		t.Fatal(err)
	}
}
