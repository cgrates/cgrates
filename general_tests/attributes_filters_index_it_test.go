//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package general_tests

import (
	"path"
	"reflect"
	"testing"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/apis"
	"github.com/cgrates/cgrates/attributes"
	"github.com/cgrates/cgrates/engine"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

var (
	attrFltrCfgPath     string
	attrFltrCfg         *config.CGRConfig
	attrFltrRPC         *birpc.Client
	alsPrfFltrConfigDIR string
	sTestsAlsFltrPrf    = []func(t *testing.T){
		testAttributeFltrSInitCfg,
		testAttributeFltrSFlushDBs,

		testAttributeFltrSStartEngine,
		testAttributeFltrSRPCConn,

		testAttributeSetFltr1,
		testAttributeSetProfile,
		testAttributeSetFltr2,
		testAttributeRemoveFltr,

		testAttributeFltrSStopEngine,
	}
)

func TestAttributeFilterSIT(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaRedis:
		t.SkipNow()
	case utils.MetaMySQL:
		alsPrfFltrConfigDIR = "attributesindexes_mysql"
	case utils.MetaMongo:
		alsPrfFltrConfigDIR = "attributesindexes_mongo"
	case utils.MetaPostgres, utils.MetaInternal:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsAlsFltrPrf {
		t.Run(alsPrfFltrConfigDIR, stest)
	}
}

func testAttributeFltrSInitCfg(t *testing.T) {
	var err error
	attrFltrCfgPath = path.Join(*utils.DataDir, "conf", "samples", alsPrfFltrConfigDIR)
	attrFltrCfg, err = config.NewCGRConfigFromPath(context.Background(), attrFltrCfgPath)
	if err != nil {
		t.Error(err)
	}
}

func testAttributeFltrSFlushDBs(t *testing.T) {
	if err := engine.InitDB(attrFltrCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testAttributeFltrSStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(attrFltrCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testAttributeFltrSRPCConn(t *testing.T) {
	attrFltrRPC = engine.NewRPCClient(t, attrFltrCfg.ListenCfg(), *utils.Encoding)
}

func testAttributeSetFltr1(t *testing.T) {
	filter := &engine.FilterWithAPIOpts{
		Filter: &engine.Filter{
			Tenant: "cgrates.org",
			ID:     "FLTR_1",
			Rules: []*engine.FilterRule{{
				Element: "~*req.Subject",
				Type:    "*prefix",
				Values:  []string{"48"},
			}},
		},
	}
	var result string
	if err := attrFltrRPC.Call(context.Background(), utils.AdminSv1SetFilter, filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	var indexes []string
	if err := attrFltrRPC.Call(context.Background(), utils.AdminSv1GetFilterIndexes, &apis.AttrGetFilterIndexes{
		ItemType: utils.MetaAttributes, Tenant: "cgrates.org", FilterType: utils.MetaPrefix},
		&indexes); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testAttributeSetProfile(t *testing.T) {
	var result string
	alsPrf := &utils.APIAttributeProfileWithAPIOpts{
		APIAttributeProfile: &utils.APIAttributeProfile{
			Tenant:    "cgrates.org",
			ID:        "ApierTest",
			FilterIDs: []string{"FLTR_1"},
			Attributes: []*utils.ExternalAttribute{{
				Path:  "*req.FL1",
				Value: "Al1",
			}},
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
		},
	}
	if err := attrFltrRPC.Call(context.Background(), utils.AdminSv1SetAttributeProfile, alsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		Event: map[string]any{
			"Subject": "44",
		},
		APIOpts: map[string]any{},
	}
	var rplyEv attributes.ProcessEventReply
	if err := attrFltrRPC.Call(context.Background(), utils.AttributeSv1ProcessEvent,
		ev, &rplyEv); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected %+v, received %+v", utils.ErrNotFound, err)
	}

	var indexes []string
	expIdx := []string{
		"*prefix:*req.Subject:48:ApierTest",
	}
	if err := attrFltrRPC.Call(context.Background(), utils.AdminSv1GetFilterIndexes, &apis.AttrGetFilterIndexes{
		ItemType: utils.MetaAttributes, Tenant: "cgrates.org", FilterType: utils.MetaPrefix},
		&indexes); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(indexes, expIdx) {
		t.Errorf("Expecting: %+v, received: %+v",
			utils.ToJSON(expIdx), utils.ToJSON(indexes))
	}
}

func testAttributeSetFltr2(t *testing.T) {
	var result string
	filter := &engine.FilterWithAPIOpts{
		Filter: &engine.Filter{
			Tenant: "cgrates.org",
			ID:     "FLTR_1",
			Rules: []*engine.FilterRule{{
				Element: "~*req.Subject",
				Type:    "*prefix",
				Values:  []string{"44"},
			}},
		},
	}
	if err := attrFltrRPC.Call(context.Background(), utils.AdminSv1SetFilter, filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	//same event for process
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		Event: map[string]any{
			"Subject": "4444",
		},
		APIOpts: map[string]any{},
	}
	exp := attributes.ProcessEventReply{
		AlteredFields: []*attributes.FieldsAltered{
			{
				MatchedProfileID: "cgrates.org:ApierTest",
				Fields:           []string{"*req.FL1"},
			},
		},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]any{
				"Subject": "4444",
				"FL1":     "Al1",
			},
			APIOpts: map[string]any{},
		},
	}
	var rplyEv attributes.ProcessEventReply
	if err := attrFltrRPC.Call(context.Background(), utils.AttributeSv1ProcessEvent,
		ev, &rplyEv); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(exp, rplyEv) {
		t.Errorf("Expected %s, received %s", utils.ToJSON(exp), utils.ToJSON(rplyEv))
	}

	var indexes []string
	expIdx := []string{
		"*prefix:*req.Subject:44:ApierTest",
	}
	if err := attrFltrRPC.Call(context.Background(), utils.AdminSv1GetFilterIndexes, &apis.AttrGetFilterIndexes{
		ItemType: utils.MetaAttributes, Tenant: "cgrates.org", FilterType: utils.MetaPrefix},
		&indexes); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(indexes, expIdx) {
		t.Errorf("Expecting: %+v, received: %+v",
			utils.ToJSON(expIdx), utils.ToJSON(indexes))
	}
}

func testAttributeRemoveFltr(t *testing.T) {
	var result string
	if err := attrFltrRPC.Call(context.Background(), utils.AdminSv1RemoveAttributeProfile, &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ApierTest"}}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	if err := attrFltrRPC.Call(context.Background(), utils.AdminSv1RemoveFilter, &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "FLTR_1"}}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	var indexes []string
	if err := attrFltrRPC.Call(context.Background(), utils.AdminSv1GetFilterIndexes, &apis.AttrGetFilterIndexes{
		ItemType: utils.MetaAttributes, Tenant: "cgrates.org", FilterType: utils.MetaPrefix},
		&indexes); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testAttributeFltrSStopEngine(t *testing.T) {
	if err := engine.KillEngine(*utils.WaitRater); err != nil {
		t.Error(err)
	}
}
