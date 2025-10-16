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
	"os"
	"path"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/attributes"
	"github.com/cgrates/cgrates/chargers"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/loaders"
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
		testAnalyzerSResetDbs,

		testAnalyzerSStartEngine,
		testAnalyzerSRPCConn,
		testAnalyzerSLoadTarrifPlans,
		testAnalyzerSChargerSv1ProcessEvent,
		testAnalyzerSV1Search,
		testAnalyzerSV1Search2,
		testAnalyzerSV1SearchWithContentFilters,
		testAnalyzerSV1BirPCSession,
		testAnalyzerSv1MultipleQuery,
		testAnalyzerSKillEngine,
	}
)

type smock struct{}

func (*smock) DisconnectPeer(ctx *context.Context,
	args *utils.AttrDisconnectSession, reply *string) error {
	return utils.ErrNotFound
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
	anzCfg, err = config.NewCGRConfigFromPath(context.Background(), anzCfgPath)
	if err != nil {
		t.Error(err)
	}
}

func testAnalyzerSResetDbs(t *testing.T) {
	if err := engine.InitDB(anzCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testAnalyzerSStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(anzCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

//make loader in confi

// Connect rpc client to rater
func testAnalyzerSRPCConn(t *testing.T) {
	srv, err := birpc.NewService(new(smock), utils.SessionSv1, true)
	if err != nil {
		t.Fatal(err)
	}
	anzRPC = engine.NewRPCClient(t, anzCfg.ListenCfg(), *utils.Encoding)
	anzBiRPC, err = utils.NewBiJSONrpcClient(anzCfg.SessionSCfg().ListenBiJSON, srv)
	if err != nil {
		t.Fatal(err)
	}
}

func testAnalyzerSLoadTarrifPlans(t *testing.T) {
	var reply string
	if err := anzRPC.Call(context.Background(), utils.LoaderSv1Run, &loaders.ArgsProcessFolder{
		APIOpts: map[string]any{utils.MetaCache: utils.MetaReload},
	}, &reply); err != nil {
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
	var result2 []*chargers.ChrgSProcessEventReply

	processedEv := []*chargers.ChrgSProcessEventReply{
		{
			ChargerSProfile: "DEFAULT",
			AlteredFields: []*attributes.FieldsAltered{
				{
					MatchedProfileID: utils.MetaDefault,
					Fields:           []string{utils.MetaOptsRunID, utils.MetaOpts + utils.NestingSep + utils.MetaChargeID, utils.MetaOpts + utils.NestingSep + utils.MetaSubsys},
				},
			},
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "event1",
				Event: map[string]any{
					"Account":     "1010",
					"Destination": "999",
					"Subject":     "Something_inter",
				},
				APIOpts: map[string]any{
					"*chargeID": "51d52496c3d63ffd60ba91e69aa532d89cc5bd79",
					"*subsys":   "*chargers",
					"*runID":    "*default",
				},
			},
		},
		{
			ChargerSProfile: "Raw",
			AlteredFields: []*attributes.FieldsAltered{
				{
					MatchedProfileID: utils.MetaDefault,
					Fields:           []string{utils.MetaOptsRunID, utils.MetaOpts + utils.NestingSep + utils.MetaChargeID, utils.MetaOpts + utils.NestingSep + utils.MetaSubsys},
				},
				{
					MatchedProfileID: "*constant:*req.RequestType:*none",
					Fields:           []string{"*req.RequestType"},
				},
			},
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "event1",
				Event: map[string]any{
					"Account":     "1010",
					"Destination": "999",
					"RequestType": "*none",
					"Subject":     "Something_inter",
				},
				APIOpts: map[string]any{
					"*chargeID":       "94e6cdc358e52bd7061f224a4bcf5faa57735989",
					"*runID":          "*raw",
					"*attrProfileIDs": []any{"*constant:*req.RequestType:*none"},
					"*context":        "*chargers",
					"*subsys":         "*chargers",
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
	if err := anzRPC.Call(context.Background(), utils.AnalyzerSv1StringQuery, &QueryArgs{HeaderFilters: `+RequestEncoding:\*birpc_json +RequestMethod:"SessionSv1.DisconnectPeer"`}, &result); err != nil {
		t.Error(err)
	} else if len(result) != 1 {
		t.Errorf("Unexpected result: %s", utils.ToJSON(result))
	}
}

func testAnalyzerSv1MultipleQuery(t *testing.T) {
	filterProfiles := []*engine.FilterWithAPIOpts{
		{
			Filter: &engine.Filter{
				ID:     "TestA_FILTER1",
				Tenant: "cgrates.org",
				Rules: []*engine.FilterRule{
					{
						Type:    utils.MetaString,
						Element: "~*req.Account",
						Values:  []string{"1001"},
					},
					{
						Type:    utils.MetaPrefix,
						Element: "~*req.Destination",
						Values:  []string{"10"},
					},
				},
			},
		},
		{
			Filter: &engine.Filter{
				ID:     "TestA_FILTER2",
				Tenant: "cgrates.org",
				Rules: []*engine.FilterRule{
					{
						Type:    utils.MetaString,
						Element: "~*req.Account",
						Values:  []string{"1002"},
					},
					{
						Type:    utils.MetaPrefix,
						Element: "~*req.Destination",
						Values:  []string{"10"},
					},
				},
			},
		},
		{
			Filter: &engine.Filter{
				ID:     "TestA_FILTER3",
				Tenant: "cgrates.org",
				Rules: []*engine.FilterRule{
					{
						Type:    utils.MetaString,
						Element: "~*req.Account",
						Values:  []string{"1003"},
					},
					{
						Type:    utils.MetaPrefix,
						Element: "~*req.Destination",
						Values:  []string{"10"},
					},
				},
			},
		},
		{
			Filter: &engine.Filter{
				ID:     "TestB_FILTER1",
				Tenant: "cgrates.org",
				Rules: []*engine.FilterRule{
					{
						Type:    utils.MetaString,
						Element: "~*req.Account",
						Values:  []string{"2001"},
					},
					{
						Type:    utils.MetaPrefix,
						Element: "~*req.Destination",
						Values:  []string{"20"},
					},
				},
			},
		},
		{
			Filter: &engine.Filter{
				ID:     "TestB_FILTER2",
				Tenant: "cgrates.org",
				Rules: []*engine.FilterRule{
					{
						Type:    utils.MetaString,
						Element: "~*req.Account",
						Values:  []string{"2002"},
					},
					{
						Type:    utils.MetaPrefix,
						Element: "~*req.Destination",
						Values:  []string{"20"},
					},
				},
			},
		},
	}

	var reply string
	for _, filterProfile := range filterProfiles {
		if err := anzRPC.Call(context.Background(), utils.AdminSv1SetFilter,
			filterProfile, &reply); err != nil {
			t.Error(err)
		} else if reply != utils.OK {
			t.Error(err)
		}
	}
	time.Sleep(50 * time.Millisecond)
	var result []map[string]any
	if err := anzRPC.Call(context.Background(), utils.AnalyzerSv1StringQuery, &QueryArgs{
		HeaderFilters:  `+RequestMethod:"AdminSv1.SetFilter"`,
		ContentFilters: []string{"*prefix:~*req.ID:TestA"},
	}, &result); err != nil {
		t.Error(err)
	} else if len(result) != 3 {
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
