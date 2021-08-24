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
	"io"
	"net/http"
	"net/http/httptest"
	"path"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
)

var (
	rsSrv       *httptest.Server
	rsBody      []byte
	rsCfgPath   string
	rsCfg       *config.CGRConfig
	rsRPC       *birpc.Client
	rsConfigDIR string //run tests for specific configuration

	sTestsRs = []func(t *testing.T){
		testResourceSInitCfg,
		testResourceSInitDataDB,
		testResourceSResetStorDB,
		testResourceSStartEngine,
		testResourceSRPCConn,
		testResourceSGetResourceBeforeSet,
		testResourceSSetResourceProfiles,
		testResourceSGetResourceAfterSet,
		testResourceSGetResourceWithConfigAfterSet,
		testResourceSGetResourceProfileIDs,
		testResourceSGetResourceProfileCount,
		testResourceSGetResourcesForEvent,
		testResourceSAllocateResources,
		testResourceSAuthorizeResourcesBeforeRelease,
		testResourceSReleaseResources,
		testResourceSAuthorizeResourcesAfterRelease,
		testResourceSRemoveResourceProfiles,
		testResourceSGetResourceProfilesAfterRemove,
		testResourceSPing,
		testResourceSKillEngine,

		// check threshold behaviour after allocation/release of resources
		testResourceSInitCfg,
		testResourceSInitDataDB,
		testResourceSResetStorDB,
		testResourceSStartEngine,
		testResourceSRPCConn,
		testResourceSStartServer,
		testResourceSSetActionProfile,
		testResourceSSetThresholdProfile,
		testResourceSSetResourceProfile,
		testResourceSCheckThresholdAfterResourceAllocate,
		testResourceSCheckThresholdAfterResourceRelease,
		testResourceSStopServer,
		testResourceSKillEngine,
	}
)

func TestResourceSIT(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		// rsConfigDIR = "resources_internal"
		t.SkipNow()
	case utils.MetaMongo:
		rsConfigDIR = "resources_mongo"
	case utils.MetaMySQL:
		rsConfigDIR = "resources_mysql"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsRs {
		t.Run(rsConfigDIR, stest)
	}
}

func testResourceSInitCfg(t *testing.T) {
	var err error
	rsCfgPath = path.Join(*dataDir, "conf", "samples", rsConfigDIR)
	rsCfg, err = config.NewCGRConfigFromPath(rsCfgPath)
	if err != nil {
		t.Error(err)
	}
}

func testResourceSInitDataDB(t *testing.T) {
	if err := engine.InitDataDB(rsCfg); err != nil {
		t.Fatal(err)
	}
}

func testResourceSResetStorDB(t *testing.T) {
	if err := engine.InitStorDB(rsCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testResourceSStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(rsCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testResourceSRPCConn(t *testing.T) {
	var err error
	rsRPC, err = newRPCClient(rsCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

//Kill the engine when it is about to be finished
func testResourceSKillEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}

func testResourceSGetResourceBeforeSet(t *testing.T) { // cache it with not found
	var rplyRes *engine.Resource
	if err := rsRPC.Call(context.Background(), utils.ResourceSv1GetResource,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "ResGroup1",
			}}, &rplyRes); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func testResourceSSetResourceProfiles(t *testing.T) {
	rsPrf1 := &engine.ResourceProfileWithAPIOpts{
		ResourceProfile: &engine.ResourceProfile{
			Tenant:            "cgrates.org",
			ID:                "ResGroup1",
			FilterIDs:         []string{"*string:~*req.Account:1001"},
			Limit:             10,
			AllocationMessage: "Approved",
			Weight:            20,
			ThresholdIDs:      []string{utils.MetaNone},
		},
	}

	var reply string
	if err := rsRPC.Call(context.Background(), utils.AdminSv1SetResourceProfile,
		rsPrf1, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}

	rsPrf2 := &engine.ResourceProfileWithAPIOpts{
		ResourceProfile: &engine.ResourceProfile{
			Tenant:            "cgrates.org",
			ID:                "ResGroup2",
			FilterIDs:         []string{"*string:~*req.Account:1001"},
			Limit:             10,
			AllocationMessage: "Approved",
			Weight:            10,
			ThresholdIDs:      []string{utils.MetaNone},
		},
	}

	if err := rsRPC.Call(context.Background(), utils.AdminSv1SetResourceProfile,
		rsPrf2, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
}

func testResourceSGetResourceAfterSet(t *testing.T) {
	var rplyRes engine.Resource
	var rplyResPrf engine.ResourceProfile
	expRes := engine.Resource{
		Tenant: "cgrates.org",
		ID:     "ResGroup1",
		Usages: make(map[string]*engine.ResourceUsage),
	}
	expResPrf := engine.ResourceProfile{
		Tenant:            "cgrates.org",
		ID:                "ResGroup1",
		FilterIDs:         []string{"*string:~*req.Account:1001"},
		Limit:             10,
		AllocationMessage: "Approved",
		Weight:            20,
		ThresholdIDs:      []string{utils.MetaNone},
	}

	if err := rsRPC.Call(context.Background(), utils.ResourceSv1GetResource,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "ResGroup1"}}, &rplyRes); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rplyRes, expRes) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(expRes), utils.ToJSON(rplyRes))
	}

	if err := rsRPC.Call(context.Background(), utils.AdminSv1GetResourceProfile,
		utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "ResGroup1",
		}, &rplyResPrf); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rplyResPrf, expResPrf) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(expResPrf), utils.ToJSON(rplyResPrf))
	}

	expRes = engine.Resource{
		Tenant: "cgrates.org",
		ID:     "ResGroup2",
		Usages: make(map[string]*engine.ResourceUsage),
	}
	expResPrf = engine.ResourceProfile{
		Tenant:            "cgrates.org",
		ID:                "ResGroup2",
		FilterIDs:         []string{"*string:~*req.Account:1001"},
		Limit:             10,
		AllocationMessage: "Approved",
		Weight:            10,
		ThresholdIDs:      []string{utils.MetaNone},
	}

	if err := rsRPC.Call(context.Background(), utils.ResourceSv1GetResource,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "ResGroup2"}}, &rplyRes); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rplyRes, expRes) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(expRes), utils.ToJSON(rplyRes))
	}

	if err := rsRPC.Call(context.Background(), utils.AdminSv1GetResourceProfile,
		utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "ResGroup2",
		}, &rplyResPrf); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rplyResPrf, expResPrf) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(expResPrf), utils.ToJSON(rplyResPrf))
	}
}

func testResourceSGetResourceWithConfigAfterSet(t *testing.T) {
	var rplyRes engine.ResourceWithConfig
	expRes := engine.ResourceWithConfig{
		Resource: &engine.Resource{
			Tenant: "cgrates.org",
			ID:     "ResGroup2",
			Usages: make(map[string]*engine.ResourceUsage),
		},
		Config: &engine.ResourceProfile{
			Tenant:            "cgrates.org",
			ID:                "ResGroup2",
			FilterIDs:         []string{"*string:~*req.Account:1001"},
			Limit:             10,
			AllocationMessage: "Approved",
			Weight:            10,
			ThresholdIDs:      []string{utils.MetaNone},
		},
	}

	if err := rsRPC.Call(context.Background(), utils.ResourceSv1GetResourceWithConfig,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "ResGroup2"}}, &rplyRes); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rplyRes, expRes) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(expRes), utils.ToJSON(rplyRes))
	}
}
func testResourceSGetResourceProfileIDs(t *testing.T) {
	var reply []string
	exp := []string{"ResGroup1", "ResGroup2"}
	if err := rsRPC.Call(context.Background(), utils.AdminSv1GetResourceProfileIDs,
		&utils.PaginatorWithTenant{
			Tenant:    "cgrates.org",
			Paginator: utils.Paginator{}}, &reply); err != nil {
		t.Error(err)
	} else {
		sort.Strings(reply)
		if !reflect.DeepEqual(reply, exp) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", exp, reply)
		}
	}
}

func testResourceSGetResourceProfileCount(t *testing.T) {
	var reply int
	if err := rsRPC.Call(context.Background(), utils.AdminSv1GetResourceProfileCount,
		&utils.TenantWithAPIOpts{Tenant: "cgrates.org"}, &reply); err != nil {
		t.Error(err)
	} else if reply != 2 {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", 2, reply)
	}
}

func testResourceSGetResourcesForEvent(t *testing.T) {
	args := &utils.ArgRSv1ResourceUsage{
		CGREvent: &utils.CGREvent{
			Event: map[string]interface{}{
				utils.AccountField: "1001",
			},
			ID: "EventTest",
		},
		UsageID: "RU_Test",
		Units:   2,
	}

	exp := engine.Resources{
		{
			Tenant: "cgrates.org",
			ID:     "ResGroup1",
			Usages: make(map[string]*engine.ResourceUsage),
		},
		{
			Tenant: "cgrates.org",
			ID:     "ResGroup2",
			Usages: make(map[string]*engine.ResourceUsage),
		},
	}
	var reply engine.Resources
	if err := rsRPC.Call(context.Background(), utils.ResourceSv1GetResourcesForEvent,
		args, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, exp) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(exp), utils.ToJSON(reply))
	}
}

func testResourceSAllocateResources(t *testing.T) {
	args := &utils.ArgRSv1ResourceUsage{
		CGREvent: &utils.CGREvent{
			Event: map[string]interface{}{
				utils.AccountField: "1001",
			},
			ID: "EventTest",
		},
		UsageTTL: utils.DurationPointer(time.Minute),
		UsageID:  "RU_Test",
		Units:    6,
	}

	var reply string
	if err := rsRPC.Call(context.Background(), utils.ResourceSv1AllocateResources,
		args, &reply); err != nil {
		t.Error(err)
	} else if reply != "Approved" {
		t.Error("Unexpected reply returned:", reply)
	}
}

func testResourceSAuthorizeResourcesBeforeRelease(t *testing.T) {
	args := &utils.ArgRSv1ResourceUsage{
		CGREvent: &utils.CGREvent{
			Event: map[string]interface{}{
				utils.AccountField: "1001",
			},
			ID: "EventTest",
		},
		UsageID:  "RU_Test",
		UsageTTL: utils.DurationPointer(time.Minute),
		Units:    7,
	}

	var reply string
	if err := rsRPC.Call(context.Background(), utils.ResourceSv1AuthorizeResources,
		args, &reply); err == nil || err.Error() != utils.ErrResourceUnauthorized.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrResourceUnauthorized, err)
	}
}

func testResourceSReleaseResources(t *testing.T) {
	args := &utils.ArgRSv1ResourceUsage{
		CGREvent: &utils.CGREvent{
			Event: map[string]interface{}{
				utils.AccountField: "1001",
			},
			ID: "EventTest",
		},
		UsageTTL: utils.DurationPointer(time.Minute),
		UsageID:  "RU_Test",
		Units:    4,
	}
	var reply string

	if err := rsRPC.Call(context.Background(), utils.ResourceSv1ReleaseResources,
		args, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned:", reply)
	}
}

func testResourceSAuthorizeResourcesAfterRelease(t *testing.T) {
	args := &utils.ArgRSv1ResourceUsage{
		CGREvent: &utils.CGREvent{
			Event: map[string]interface{}{
				utils.AccountField: "1001",
			},
			ID: "EventTest",
		},
		UsageID:  "RU_Test",
		UsageTTL: utils.DurationPointer(time.Minute),
		Units:    7,
	}

	var reply string
	if err := rsRPC.Call(context.Background(), utils.ResourceSv1AuthorizeResources,
		args, &reply); err != nil {
		t.Error(err)
	} else if reply != "Approved" {
		t.Error("Unexpected reply returned:", reply)
	}
}

func testResourceSRemoveResourceProfiles(t *testing.T) {
	var reply string

	if err := rsRPC.Call(context.Background(), utils.AdminSv1RemoveResourceProfile,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "ResGroup1",
		}}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned:", reply)
	}

	if err := rsRPC.Call(context.Background(), utils.AdminSv1RemoveResourceProfile,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "ResGroup2",
		}}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned:", reply)
	}
}

func testResourceSGetResourceProfilesAfterRemove(t *testing.T) {
	var rplyResPrf engine.ResourceProfile
	if err := rsRPC.Call(context.Background(), utils.AdminSv1GetResourceProfile,
		utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "ResGroup1",
		}, &rplyResPrf); err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}

	if err := rsRPC.Call(context.Background(), utils.AdminSv1GetResourceProfile,
		utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "ResGroup2",
		}, &rplyResPrf); err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func testResourceSPing(t *testing.T) {
	var reply string
	if err := rsRPC.Call(context.Background(), utils.ResourceSv1Ping,
		new(utils.CGREvent), &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Error("Unexpected reply returned:", reply)
	}
}

func testResourceSStartServer(t *testing.T) {
	rsSrv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error
		rsBody, err = io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
		}

		r.Body.Close()
	}))
}

func testResourceSStopServer(t *testing.T) {
	rsSrv.Close()
}

func testResourceSSetActionProfile(t *testing.T) {
	actPrf := &engine.ActionProfileWithAPIOpts{
		ActionProfile: &engine.ActionProfile{
			Tenant: "cgrates.org",
			ID:     "actPrfID",
			Actions: []*engine.APAction{
				{
					ID:   "actID",
					Type: utils.MetaHTTPPost,
					Diktats: []*engine.APDiktat{
						{
							Path: rsSrv.URL,
						},
					},
					TTL: time.Duration(time.Minute),
				},
			},
		},
	}

	var reply *string
	if err := rsRPC.Call(context.Background(), utils.AdminSv1SetActionProfile,
		actPrf, &reply); err != nil {
		t.Error(err)
	}
}

func testResourceSSetThresholdProfile(t *testing.T) {
	thPrf := &engine.ThresholdProfileWithAPIOpts{
		ThresholdProfile: &engine.ThresholdProfile{
			Tenant:           "cgrates.org",
			FilterIDs:        []string{"*string:~*opts.*eventType:ResourceUpdate"},
			ID:               "THD_1",
			MaxHits:          -1,
			ActionProfileIDs: []string{"actPrfID"},
		},
	}
	var reply string
	if err := rsRPC.Call(context.Background(), utils.AdminSv1SetThresholdProfile,
		thPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}

	args := &utils.TenantID{
		Tenant: "cgrates.org",
		ID:     "THD_1",
	}
	var result *engine.ThresholdProfile
	if err := rsRPC.Call(context.Background(), utils.AdminSv1GetThresholdProfile,
		args, &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(result, thPrf.ThresholdProfile) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(thPrf.ThresholdProfile), utils.ToJSON(result))
	}
}

func testResourceSSetResourceProfile(t *testing.T) {
	rsPrf := &engine.ResourceProfile{
		Tenant:            "cgrates.org",
		ID:                "RES_1",
		FilterIDs:         []string{"*string:~*req.Account:1001"},
		AllocationMessage: "Approved",
		Limit:             10,
		ThresholdIDs:      []string{"THD_1"},
	}
	var reply string
	if err := rsRPC.Call(context.Background(), utils.AdminSv1SetResourceProfile,
		rsPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}

	var result *engine.ResourceProfile
	if err := rsRPC.Call(context.Background(), utils.AdminSv1GetResourceProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: rsPrf.ID}, &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(result, rsPrf) {
		t.Errorf("expected: %+v, received: %+v",
			utils.ToJSON(rsPrf), utils.ToJSON(result))
	}
}

func testResourceSCheckThresholdAfterResourceAllocate(t *testing.T) {
	var reply string
	argsRU := utils.ArgRSv1ResourceUsage{
		UsageID: "RU_1",
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "EV_1",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
			},
		},

		Units: 1,
	}

	expBody := `{"*opts":{"*eventType":"ResourceUpdate"},"*req":{"EventType":"ResourceUpdate","ResourceID":"RES_1","Usage":0}}`
	if err := rsRPC.Call(context.Background(), utils.ResourceSv1AllocateResources,
		argsRU, &reply); err != nil {
		t.Error(err)
	} else if reply != "Approved" {
		t.Error("Unexpected reply returned", reply)
	}

	if expBody != string(rsBody) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", expBody, string(rsBody))
	}

	args := &utils.TenantID{
		Tenant: "cgrates.org",
		ID:     "THD_1",
	}
	var result *engine.Threshold
	if err := rsRPC.Call(context.Background(), utils.ThresholdSv1GetThreshold,
		args, &result); err != nil {
		t.Error(err)
	} else if result.Hits != 1 {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", 1, result.Hits)
	}
}

func testResourceSCheckThresholdAfterResourceRelease(t *testing.T) {
	argsRU := &utils.ArgRSv1ResourceUsage{
		UsageID: "RU_1",
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "EV_1",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
			},
		},
	}
	var reply string
	if err := rsRPC.Call(context.Background(), utils.ResourceSv1ReleaseResources,
		argsRU, &reply); err != nil {
		t.Error(err)
	}

	args := &utils.TenantID{
		Tenant: "cgrates.org",
		ID:     "THD_1",
	}
	var result *engine.Threshold
	if err := rsRPC.Call(context.Background(), utils.ThresholdSv1GetThreshold,
		args, &result); err != nil {
		t.Error(err)
	} else if result.Hits != 2 {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", 2, result.Hits)
	}
}
