//go:build integration
// +build integration

/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package analyzers

import (
	"errors"
	"os"
	"path"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/birpc/jsonrpc"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

var (
	anzCfgPath string
	anzCfg     *config.CGRConfig
	anzRPC     *birpc.Client
	anzBiRPC   *birpc.BirpcClient

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
		testAnalyzerSV1BirPCSession,
		testAnalyzerSKillEngine,
	}
)

func newRPCClient(cfg *config.ListenCfg) (c *birpc.Client, err error) {
	switch *utils.Encoding {
	case utils.MetaJSON:
		return jsonrpc.Dial(utils.TCP, cfg.RPCJSONListen)
	case utils.MetaGOB:
		return birpc.Dial(utils.TCP, cfg.RPCGOBListen)
	default:
		return nil, errors.New("UNSUPPORTED_RPC")
	}
}

// Test start here
func TestAnalyzerSIT(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaMongo:
	case utils.MetaInternal, utils.MetaMySQL, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

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
	anzCfgPath = path.Join(*utils.DataDir, "conf", "samples", "analyzers")
	anzCfg, err = config.NewCGRConfigFromPath(anzCfgPath)
	if err != nil {
		t.Error(err)
	}
}

func testAnalyzerSInitDataDb(t *testing.T) {
	if err := engine.InitDataDB(anzCfg); err != nil {
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
	if _, err := engine.StopStartEngine(anzCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testAnalyzerSRPCConn(t *testing.T) {
	srv, err := birpc.NewService(new(smock), utils.AgentV1, true)
	if err != nil {
		t.Fatal(err)
	}
	anzRPC, err = newRPCClient(anzCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
	anzBiRPC, err = utils.NewBiJSONrpcClient(anzCfg.ListenCfg().BiJSONListen, srv)
	if err != nil {
		t.Fatal(err)
	}
}

func testAnalyzerSLoadTarrifPlans(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*utils.DataDir, "tariffplans", "tutorial")}
	if err := anzRPC.Call(context.Background(), utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
}

func testAnalyzerSChargerSv1ProcessEvent(t *testing.T) {
	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "event1",
		Event: map[string]any{
			utils.AccountField: "1010",
			utils.Subject:      "Something_inter",
			utils.Destination:  "999",
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
				Event: map[string]any{
					"Account":     "1010",
					"Destination": "999",
					"RunID":       "*default",
					"Subject":     "Something_inter",
				},
				APIOpts: map[string]any{"*subsys": "*chargers"},
			},
		},
		{
			ChargerSProfile:    "Raw",
			AttributeSProfiles: []string{"*constant:*req.RequestType:*none"},
			AlteredFields:      []string{"*req.RunID", "*req.RequestType"},
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "event1",
				Event: map[string]any{
					"Account":     "1010",
					"Destination": "999",
					"RequestType": "*none",
					"RunID":       "*raw",
					"Subject":     "Something_inter",
				},
				APIOpts: map[string]any{
					"*subsys":                      "*chargers",
					utils.OptsAttributesProfileIDs: []any{"*constant:*req.RequestType:*none"},
				},
			},
		},
	}

	if err := anzRPC.Call(context.Background(), utils.ChargerSv1ProcessEvent, cgrEv, &result2); err != nil {
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
	// need to wait in order for the log gorutine to execute
	time.Sleep(10 * time.Millisecond)
	var result []map[string]any
	if err := anzRPC.Call(context.Background(), utils.AnalyzerSv1StringQuery, &QueryArgs{HeaderFilters: `+RequestEncoding:\*internal +RequestMethod:AttributeSv1\.ProcessEvent`}, &result); err != nil {
		t.Error(err)
	} else if len(result) != 1 {
		t.Errorf("Unexpected result: %s", utils.ToJSON(result))
	}
}

func testAnalyzerSV1Search2(t *testing.T) {
	var result []map[string]any
	if err := anzRPC.Call(context.Background(), utils.AnalyzerSv1StringQuery, &QueryArgs{HeaderFilters: `+RequestEncoding:\*json +RequestMethod:ChargerSv1\.ProcessEvent`}, &result); err != nil {
		t.Error(err)
	} else if len(result) != 1 {
		t.Errorf("Unexpected result: %s", utils.ToJSON(result))
	}
}

func testAnalyzerSV1SearchWithContentFilters(t *testing.T) {
	var result []map[string]any
	if err := anzRPC.Call(context.Background(), utils.AnalyzerSv1StringQuery, &QueryArgs{
		HeaderFilters:  `+RequestEncoding:\*json`,
		ContentFilters: []string{"*string:~*req.Event.Account:1010"},
	}, &result); err != nil {
		t.Error(err)
	} else if len(result) != 1 {
		t.Errorf("Unexpected result: %s", utils.ToJSON(result))
	}
}

func testAnalyzerSV1BirPCSession(t *testing.T) {
	var rply string
	anzBiRPC.Call(context.Background(), utils.SessionSv1STIRIdentity,
		&sessions.V1STIRIdentityArgs{}, &rply) // only call to register the birpc
	if err := anzRPC.Call(context.Background(), utils.SessionSv1DisconnectPeer, &utils.DPRArgs{}, &rply); err == nil ||
		err.Error() != utils.ErrPartiallyExecuted.Error() {
		t.Fatal(err)
	}
	time.Sleep(10 * time.Millisecond)
	var result []map[string]any
	if err := anzRPC.Call(context.Background(), utils.AnalyzerSv1StringQuery, &QueryArgs{HeaderFilters: `+RequestEncoding:\*birpc_json +RequestMethod:"AgentV1.DisconnectPeer"`}, &result); err != nil {
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

type smock struct{}

func (*smock) DisconnectPeer(ctx *context.Context,
	args *utils.DPRArgs, reply *string) error {
	return utils.ErrNotFound
}
