//go:build integration
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

package apis

import (
	"os"
	"path"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/analyzers"
	"github.com/cgrates/cgrates/chargers"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/loaders"
	"github.com/cgrates/cgrates/utils"
)

var (
	anzCfgPath string
	anzCfg     *config.CGRConfig
	anzBiRPC   *birpc.Client

	sTestsAnz = []func(t *testing.T){
		testAnalyzerSInitCfg,
		testAnalyzerSInitDataDb,
		testAnalyzerSResetStorDb,
		testAnalyzerSStartEngine,
		testAnzBiSRPCConn,
		testAnalyzerSLoad,
		// testAnalyzerSGetAttributeProfiles,
		testAnalyzerSChargerSv1ProcessEvent,
		testAnalyzerSSearchCall1,
		testAnalyzerSSearchCall2,
		testAnalyzerSGetFilterIDs,
		testAnalyzerSSearchCall3,
		testAnalyzerSKillEngine,
	}
)

func TestAnalyzerSIT(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal, utils.MetaMySQL, utils.MetaPostgres:
		t.SkipNow()
	case utils.MetaMongo:
		anzCfgPath = path.Join(*utils.DataDir, "conf", "samples", "analyzers")
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsAnz {
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

	anzCfg, err = config.NewCGRConfigFromPath(context.Background(), anzCfgPath)
	if err != nil {
		t.Error(err)
	}
}

func testAnalyzerSInitDataDb(t *testing.T) {
	if err := engine.InitDataDB(anzCfg); err != nil {
		t.Fatal(err)
	}
}

func testAnalyzerSResetStorDb(t *testing.T) {
	if err := engine.InitStorDB(anzCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testAnalyzerSStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(anzCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

func testAnzBiSRPCConn(t *testing.T) {
	anzBiRPC = engine.NewRPCClient(t, anzCfg.ListenCfg(), *utils.Encoding)
}

func testAnalyzerSLoad(t *testing.T) {
	var reply string

	if err := anzBiRPC.Call(context.Background(), utils.LoaderSv1Run, &loaders.ArgsProcessFolder{
		APIOpts: map[string]any{utils.MetaCache: utils.MetaReload},
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	time.Sleep(100 * time.Millisecond)
}

func testAnalyzerSKillEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
	if err := os.RemoveAll(anzCfg.AnalyzerSCfg().DBPath); err != nil {
		t.Fatal(err)
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

	if err := anzBiRPC.Call(context.Background(), utils.ChargerSv1ProcessEvent, cgrEv, &result2); err != nil {
		t.Fatal(err)
	}
}

func testAnalyzerSSearchCall1(t *testing.T) {
	time.Sleep(10 * time.Millisecond)
	var result []map[string]any
	queryArgs := &analyzers.QueryArgs{
		HeaderFilters: `+RequestEncoding:\*internal +RequestMethod:AttributeSv1\.ProcessEvent`,
	}
	if err := anzBiRPC.Call(context.Background(), utils.AnalyzerSv1StringQuery, queryArgs, &result); err != nil {
		t.Error(err)
	} else if len(result) != 1 {
		t.Errorf("Unexpected result: %s", utils.ToJSON(result))
	}
}

func testAnalyzerSSearchCall2(t *testing.T) {
	var result []map[string]any
	queryArgs := &analyzers.QueryArgs{
		HeaderFilters: `+RequestEncoding:\*internal +RequestMethod:ChargerSv1\.ProcessEvent`,
	}
	if err := anzBiRPC.Call(context.Background(), utils.AnalyzerSv1StringQuery, queryArgs, &result); err != nil {
		t.Error(err)
	} else if len(result) != 1 {
		t.Errorf("Unexpected result: %s", utils.ToJSON(result))
	}
}

func testAnalyzerSGetFilterIDs(t *testing.T) {
	var filterIDs []string
	if err := anzBiRPC.Call(context.Background(), utils.AdminSv1GetFilterIDs,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &filterIDs); err != nil {
		t.Error(err)
	}
	expIDs := []string{"FLTR_ACNT_1001", "FLTR_ACNT_1001_1002", "FLTR_ACNT_1002", "FLTR_ACNT_1003", "FLTR_ACNT_1003_1001", "FLTR_DST_FS", "FLTR_RES"}
	sort.Slice(filterIDs, func(i, j int) bool {
		return filterIDs[i] < filterIDs[j]
	})
	if !reflect.DeepEqual(filterIDs, expIDs) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expIDs), utils.ToJSON(filterIDs))
	}

	var result engine.Filter
	args := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "cgrates.org",
			ID:     filterIDs[0],
		},
		APIOpts: map[string]any{},
	}
	if err := anzBiRPC.Call(context.Background(), utils.AdminSv1GetFilter, args, &result); err != nil {
		t.Error(err)
	}
	expFilter := engine.Filter{
		Tenant: "cgrates.org",
		ID:     "FLTR_ACNT_1001",
		Rules: []*engine.FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Account",
				Values:  []string{"1001"},
			},
		},
	}
	if !reflect.DeepEqual(result, expFilter) {
		t.Errorf("Expected %v \n but received \n %v", expFilter, result)
	}
	time.Sleep(50 * time.Millisecond)
}

func testAnalyzerSSearchCall3(t *testing.T) {
	var result []map[string]any
	queryArgs := &analyzers.QueryArgs{
		HeaderFilters: `+RequestEncoding:\*json +RequestMethod:AdminSv1\.GetFilter`,
	}
	if err := anzBiRPC.Call(context.Background(), utils.AnalyzerSv1StringQuery, queryArgs, &result); err != nil {
		t.Error(err)
	} else if len(result) != 2 {
		t.Errorf("Unexpected result: %s", utils.ToJSON(result))
	}
}
