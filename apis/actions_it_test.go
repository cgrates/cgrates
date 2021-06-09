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
	"path"
	"reflect"
	"testing"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	// actSrv       *httptest.Server
	// actBody      []byte
	actCfgPath   string
	actCfg       *config.CGRConfig
	actRPC       *birpc.Client
	actConfigDIR string //run tests for specific configuration

	sTestsAct = []func(t *testing.T){
		testActionsInitCfg,
		testActionsInitDataDB,
		testActionsResetStorDB,
		testActionsStartEngine,
		testActionsRPCConn,
		testActionsGetActionProfileBeforeSet,
		testActionsGetActionProfileIDsBeforeSet,
		testActionsGetActionProfileCountBeforeSet,
		testActionsSetActionProfile,
		testActionsGetActionProfileAfterSet,
		testActionsGetActionProfileIDsAfterSet,
		testActionsGetActionProfileCountAfterSet,
		testActionsRemoveActionProfile,
		testActionsGetActionProfileAfterRemove,
		testActionsPing,
		// testActionsStartServer,
		// testActionsStopServer,
		testActionsKillEngine,
	}
)

func TestActionsIT(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		actConfigDIR = "apis_actions_internal"
	case utils.MetaMongo:
		actConfigDIR = "apis_actions_mongo"
	case utils.MetaMySQL:
		actConfigDIR = "apis_actions_mysql"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsAct {
		t.Run(actConfigDIR, stest)
	}
}

func testActionsInitCfg(t *testing.T) {
	var err error
	actCfgPath = path.Join(*dataDir, "conf", "samples", actConfigDIR)
	actCfg, err = config.NewCGRConfigFromPath(actCfgPath)
	if err != nil {
		t.Error(err)
	}
}

func testActionsInitDataDB(t *testing.T) {
	if err := engine.InitDataDB(actCfg); err != nil {
		t.Fatal(err)
	}
}

func testActionsResetStorDB(t *testing.T) {
	if err := engine.InitStorDB(actCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testActionsStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(actCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testActionsRPCConn(t *testing.T) {
	var err error
	actRPC, err = newRPCClient(actCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

//Kill the engine when it is about to be finished
func testActionsKillEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}

func testActionsGetActionProfileBeforeSet(t *testing.T) {
	var rplyAct engine.ActionProfile
	if err := actRPC.Call(context.Background(), utils.AdminSv1GetActionProfile,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "actPrfID",
			}}, &rplyAct); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func testActionsGetActionProfileIDsBeforeSet(t *testing.T) {
	var rplyActIDs []string
	if err := actRPC.Call(context.Background(), utils.AdminSv1GetActionProfileIDs,
		&utils.PaginatorWithTenant{
			Tenant:    "cgrates.org",
			Paginator: utils.Paginator{},
		}, &rplyActIDs); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func testActionsGetActionProfileCountBeforeSet(t *testing.T) {
	var rplyCount int
	if err := actRPC.Call(context.Background(), utils.AdminSv1GetActionProfileCount,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "actPrfID",
			}}, &rplyCount); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	} else if rplyCount != 0 {
		t.Errorf("expected <%+v>, \nreceived: <%+v>", 0, rplyCount)
	}
}

func testActionsSetActionProfile(t *testing.T) {
	actPrf := &engine.ActionProfileWithAPIOpts{
		ActionProfile: &engine.ActionProfile{
			Tenant: "cgrates.org",
			ID:     "actPrfID",
			Actions: []*engine.APAction{
				{
					ID: "actID",
				},
			},
		},
	}

	var reply string
	if err := actRPC.Call(context.Background(), utils.AdminSv1SetActionProfile,
		actPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned:", reply)
	}

	var rplyActPrf engine.ActionProfile
	if err := actRPC.Call(context.Background(), utils.AdminSv1GetActionProfile,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "actPrfID",
			}}, &rplyActPrf); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rplyActPrf, *actPrf.ActionProfile) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", actPrf.ActionProfile, rplyActPrf)
	}
}

func testActionsGetActionProfileAfterSet(t *testing.T) {
	expAct := engine.ActionProfile{
		Tenant: "cgrates.org",
		ID:     "actPrfID",
		Actions: []*engine.APAction{
			{
				ID: "actID",
			},
		},
	}

	var rplyAct engine.ActionProfile
	if err := actRPC.Call(context.Background(), utils.AdminSv1GetActionProfile,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "actPrfID",
			}}, &rplyAct); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rplyAct, expAct) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(expAct), utils.ToJSON(rplyAct))
	}
}

func testActionsGetActionProfileIDsAfterSet(t *testing.T) {
	expActIDs := []string{"actPrfID"}

	var rplyActIDs []string
	if err := actRPC.Call(context.Background(), utils.AdminSv1GetActionProfileIDs,
		&utils.PaginatorWithTenant{
			Tenant:    "cgrates.org",
			Paginator: utils.Paginator{},
		}, &rplyActIDs); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rplyActIDs, expActIDs) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", expActIDs, rplyActIDs)
	}
}

func testActionsGetActionProfileCountAfterSet(t *testing.T) {
	var rplyCount int
	if err := actRPC.Call(context.Background(), utils.AdminSv1GetActionProfileCount,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "actPrfID",
			}}, &rplyCount); err != nil {
		t.Error(err)
	} else if rplyCount != 1 {
		t.Errorf("expected <%+v>, \nreceived: <%+v>", 1, rplyCount)
	}
}

func testActionsRemoveActionProfile(t *testing.T) {
	var reply string
	if err := actRPC.Call(context.Background(), utils.AdminSv1RemoveActionProfile,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "actPrfID",
			}}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned:", reply)
	}
}

func testActionsGetActionProfileAfterRemove(t *testing.T) {
	var rplyAct engine.ActionProfile
	if err := actRPC.Call(context.Background(), utils.AdminSv1GetActionProfile,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "actPrfID",
			}}, &rplyAct); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func testActionsPing(t *testing.T) {
	var reply string
	if err := actRPC.Call(context.Background(), utils.StatSv1Ping,
		new(utils.CGREvent), &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Error("Unexpected reply returned:", reply)
	}
}

// func testActionsStartServer(t *testing.T) {
// 	actSrv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		var err error
// 		actBody, err = io.ReadAll(r.Body)
// 		if err != nil {
// 			w.WriteHeader(http.StatusNotFound)
// 		}

// 		r.Body.Close()
// 	}))
// }

// func testActionsStopServer(t *testing.T) {
// 	actSrv.Close()
// }
