//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package general_tests

import (
	"os/exec"
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
	fltrUpdateCfgPath1, fltrUpdateCfgPath2 string
	fltrUpdateCfgDIR1, fltrUpdateCfgDIR2   string
	fltrUpdateCfg1, fltrUpdateCfg2         *config.CGRConfig
	fltrUpdateRPC1, fltrUpdateRPC2         *birpc.Client
	testEng1                               *exec.Cmd
	sTestsFilterUpdate                     = []func(t *testing.T){
		testFilterUpdateInitCfg,
		testFilterUpdateFlushDBs,
		testFilterUpdateStartEngine,
		testFilterUpdateRpcConn,
		testFilterUpdateSetFilterE1,
		testFilterUpdateSetAttrProfileE1,
		testFilterUpdateGetAttrProfileForEventEv1E1,
		testFilterUpdateGetAttrProfileForEventEv1E2,
		testFilterUpdateGetAttrProfileForEventEv2E1NotMatching,
		testFilterUpdateGetAttrProfileForEventEv2E2NotMatching,
		testFilterUpdateSetFilterAfterAttrE1,
		testFilterUpdateGetAttrProfileForEventEv1E1NotMatching,
		testFilterUpdateGetAttrProfileForEventEv1E2NotMatching,
		testFilterUpdateGetAttrProfileForEventEv2E1,
		testFilterUpdateGetAttrProfileForEventEv2E2,
		testFilterUpdateStopEngine,
	}
	ev1 = &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "Event1",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
	}
	ev2 = &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "Event2",
		Event: map[string]any{
			utils.AccountField: "1002",
		},
	}
)

func TestFilterUpdateIT(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		t.SkipNow()
	case utils.MetaRedis:
		fltrUpdateCfgDIR1 = "fltr_update_e1_mysql"
		fltrUpdateCfgDIR2 = "tutredis"
	case utils.MetaMySQL:
		t.SkipNow()
	case utils.MetaMongo:
		fltrUpdateCfgDIR1 = "fltr_update_e1_mongo"
		fltrUpdateCfgDIR2 = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}

	for _, stest1 := range sTestsFilterUpdate {
		t.Run(*utils.DBType, stest1)
	}
}

// Init Config
func testFilterUpdateInitCfg(t *testing.T) {
	var err error
	fltrUpdateCfgPath1 = path.Join(*utils.DataDir, "conf", "samples", "cache_replicate", fltrUpdateCfgDIR1)
	if fltrUpdateCfg1, err = config.NewCGRConfigFromPath(context.Background(), fltrUpdateCfgPath1); err != nil {
		t.Fatal(err)
	}
	fltrUpdateCfgPath2 = path.Join(*utils.DataDir, "conf", "samples", fltrUpdateCfgDIR2)
	if fltrUpdateCfg2, err = config.NewCGRConfigFromPath(context.Background(), fltrUpdateCfgPath2); err != nil {
		t.Fatal(err)
	}
}

// Remove data in both rating and accounting db
func testFilterUpdateFlushDBs(t *testing.T) {
	if err := engine.InitDB(fltrUpdateCfg1); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testFilterUpdateStartEngine(t *testing.T) {
	var err error
	if _, err = engine.StopStartEngine(fltrUpdateCfgPath1, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
	if testEng1, err = engine.StartEngine(fltrUpdateCfgPath2, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}

}

// Connect rpc client to rater
func testFilterUpdateRpcConn(t *testing.T) {
	fltrUpdateRPC1 = engine.NewRPCClient(t, fltrUpdateCfg1.ListenCfg(), *utils.Encoding)
	fltrUpdateRPC2 = engine.NewRPCClient(t, fltrUpdateCfg2.ListenCfg(), *utils.Encoding)
}

func testFilterUpdateStopEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}

func testFilterUpdateSetFilterE1(t *testing.T) {
	fltr := &engine.FilterWithAPIOpts{
		Filter: &engine.Filter{
			ID:     "FLTR_ID",
			Tenant: "cgrates.org",
			Rules: []*engine.FilterRule{
				{
					Type:    utils.MetaString,
					Element: "~*req.Account",
					Values:  []string{"1001"},
				},
			},
		},
		APIOpts: map[string]any{
			utils.MetaCache: utils.MetaLoad,
		},
	}

	var reply string
	if err := fltrUpdateRPC1.Call(context.Background(), utils.AdminSv1SetFilter, fltr, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}

	var result *engine.Filter
	if err := fltrUpdateRPC1.Call(context.Background(), utils.AdminSv1GetFilter,
		&utils.TenantID{Tenant: "cgrates.org", ID: "FLTR_ID"}, &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(fltr.Filter, result) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ToJSON(fltr.Filter), utils.ToJSON(result))
	}
}

func testFilterUpdateSetAttrProfileE1(t *testing.T) {
	attrPrf := &utils.APIAttributeProfileWithAPIOpts{
		APIAttributeProfile: &utils.APIAttributeProfile{
			FilterIDs: []string{"FLTR_ID"},
			ID:        "ATTR_ID",
			Tenant:    "cgrates.org",
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
			Attributes: []*utils.ExternalAttribute{
				{
					Path:  "*req.Account",
					Value: "1003",
					Type:  utils.MetaConstant,
				},
			},
		},
		APIOpts: map[string]any{
			utils.MetaCache: utils.MetaNone,
		},
	}

	var reply string
	if err := fltrUpdateRPC1.Call(context.Background(), utils.AdminSv1SetAttributeProfile, attrPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	var result *utils.APIAttributeProfile
	if err := fltrUpdateRPC1.Call(context.Background(), utils.AdminSv1GetAttributeProfile,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ATTR_ID"}}, &result); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(attrPrf.APIAttributeProfile, result) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ToJSON(attrPrf.APIAttributeProfile), utils.ToJSON(result))
	}
}

func testFilterUpdateGetAttrProfileForEventEv1E1(t *testing.T) {
	eAttrPrf := &utils.APIAttributeProfile{
		Tenant:    "cgrates.org",
		FilterIDs: []string{"FLTR_ID"},
		ID:        "ATTR_ID",
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		Attributes: []*utils.ExternalAttribute{
			{
				Path:  "*req.Account",
				Value: "1003",
				Type:  utils.MetaConstant,
			},
		},
	}

	var attrReply *utils.APIAttributeProfile
	if err := fltrUpdateRPC1.Call(context.Background(), utils.AttributeSv1GetAttributeForEvent,
		ev1, &attrReply); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(eAttrPrf, attrReply) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ToJSON(eAttrPrf), utils.ToJSON(attrReply))
	}
}

func testFilterUpdateGetAttrProfileForEventEv1E2(t *testing.T) {
	eAttrPrf := &utils.APIAttributeProfile{
		Tenant:    "cgrates.org",
		FilterIDs: []string{"FLTR_ID"},
		ID:        "ATTR_ID",
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		Attributes: []*utils.ExternalAttribute{
			{
				Path:  "*req.Account",
				Value: "1003",
				Type:  utils.MetaConstant,
			},
		},
	}

	var attrReply *utils.APIAttributeProfile
	if err := fltrUpdateRPC2.Call(context.Background(), utils.AttributeSv1GetAttributeForEvent,
		ev1, &attrReply); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(eAttrPrf, attrReply) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ToJSON(eAttrPrf), utils.ToJSON(attrReply))
	}
}

func testFilterUpdateGetAttrProfileForEventEv2E1(t *testing.T) {
	eAttrPrf := &utils.APIAttributeProfile{
		Tenant:    "cgrates.org",
		FilterIDs: []string{"FLTR_ID"},
		ID:        "ATTR_ID",
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		Attributes: []*utils.ExternalAttribute{
			{
				Path:  "*req.Account",
				Value: "1003",
				Type:  utils.MetaConstant,
			},
		},
	}

	var attrReply *utils.APIAttributeProfile
	if err := fltrUpdateRPC1.Call(context.Background(), utils.AttributeSv1GetAttributeForEvent,
		ev2, &attrReply); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(eAttrPrf, attrReply) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ToJSON(eAttrPrf), utils.ToJSON(attrReply))
	}
}

func testFilterUpdateGetAttrProfileForEventEv2E2(t *testing.T) {
	eAttrPrf := &utils.APIAttributeProfile{
		Tenant:    "cgrates.org",
		FilterIDs: []string{"FLTR_ID"},
		ID:        "ATTR_ID",
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		Attributes: []*utils.ExternalAttribute{
			{
				Path:  "*req.Account",
				Value: "1003",
				Type:  utils.MetaConstant,
			},
		},
	}

	var attrReply *utils.APIAttributeProfile
	if err := fltrUpdateRPC2.Call(context.Background(), utils.AttributeSv1GetAttributeForEvent,
		ev2, &attrReply); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(eAttrPrf, attrReply) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ToJSON(eAttrPrf), utils.ToJSON(attrReply))
	}
}

func testFilterUpdateSetFilterAfterAttrE1(t *testing.T) {
	fltr := &engine.FilterWithAPIOpts{
		Filter: &engine.Filter{
			ID:     "FLTR_ID",
			Tenant: "cgrates.org",
			Rules: []*engine.FilterRule{
				{
					Type:    utils.MetaString,
					Element: "~*req.Account",
					Values:  []string{"1002"},
				},
			},
		},
		APIOpts: map[string]any{
			utils.MetaCache: utils.MetaLoad,
		},
	}

	var reply string
	if err := fltrUpdateRPC1.Call(context.Background(), utils.AdminSv1SetFilter, fltr, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}

	var result *engine.Filter
	if err := fltrUpdateRPC1.Call(context.Background(), utils.AdminSv1GetFilter,
		&utils.TenantID{Tenant: "cgrates.org", ID: "FLTR_ID"}, &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(fltr.Filter, result) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ToJSON(fltr.Filter), utils.ToJSON(result))
	}
}

func testFilterUpdateGetAttrProfileForEventEv1E1NotMatching(t *testing.T) {
	var attrReply *utils.APIAttributeProfile
	if err := fltrUpdateRPC1.Call(context.Background(), utils.AttributeSv1GetAttributeForEvent,
		ev1, &attrReply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func testFilterUpdateGetAttrProfileForEventEv1E2NotMatching(t *testing.T) {
	var attrReply *utils.APIAttributeProfile
	if err := fltrUpdateRPC2.Call(context.Background(), utils.AttributeSv1GetAttributeForEvent,
		ev1, &attrReply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func testFilterUpdateGetAttrProfileForEventEv2E1NotMatching(t *testing.T) {
	var attrReply *utils.APIAttributeProfile
	if err := fltrUpdateRPC1.Call(context.Background(), utils.AttributeSv1GetAttributeForEvent,
		ev2, &attrReply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func testFilterUpdateGetAttrProfileForEventEv2E2NotMatching(t *testing.T) {
	var attrReply *utils.APIAttributeProfile
	if err := fltrUpdateRPC2.Call(context.Background(), utils.AttributeSv1GetAttributeForEvent,
		ev2, &attrReply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}
