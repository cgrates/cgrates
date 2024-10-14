//go:build flaky

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
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	v1 "github.com/cgrates/cgrates/apier/v1"
	v2 "github.com/cgrates/cgrates/apier/v2"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	fltrRplDB string

	fltrRplInternalCfgPath string
	fltrRplInternalCfg     *config.CGRConfig
	fltrRplInternalRPC     birpc.ClientConnector

	fltrRplEngine1CfgPath string
	fltrRplEngine1Cfg     *config.CGRConfig
	fltrRplEngine1RPC     birpc.ClientConnector

	fltrRplEngine2CfgPath string
	fltrRplEngine2Cfg     *config.CGRConfig
	fltrRplEngine2RPC     birpc.ClientConnector

	sTestsFltrRpl = []func(t *testing.T){
		testFltrRplInitCfg,
		testFltrRplInitDBs,
		testFltrRplStartEngine,
		testFltrRplRPCConn,

		testFltrRplAttributeProfile,
		testFltrRplFilters,
		testFltrRplThresholdProfile,
		testFltrRplStatQueueProfile,
		testFltrRplResourceProfile,
		testFltrRplRouteProfile,
		testFltrRplChargerProfile,
		testFltrRplDispatcherProfile,
		testFltrRplDispatcherHost,
		testFltrRplAccount,
		testFltrRplDestination,

		testFltrRplKillEngine,
	}
)

// Test start here
func TestFilteredReplication(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaMySQL:
		fltrRplDB = "redis"
	case utils.MetaMongo:
		fltrRplDB = "mongo"
	case utils.MetaPostgres, utils.MetaInternal:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsFltrRpl {
		t.Run("TestFilteredReplication_"+fltrRplDB, stest)
	}
}

func testFltrRplInitCfg(t *testing.T) {
	var err error

	fltrRplInternalCfgPath = path.Join(*utils.DataDir, "conf", "samples", "filtered_replication", "internal")
	fltrRplEngine1CfgPath = path.Join(*utils.DataDir, "conf", "samples", "filtered_replication", "engine1_"+fltrRplDB)
	fltrRplEngine2CfgPath = path.Join(*utils.DataDir, "conf", "samples", "filtered_replication", "engine2_"+fltrRplDB)

	if fltrRplInternalCfg, err = config.NewCGRConfigFromPath(fltrRplInternalCfgPath); err != nil {
		t.Fatal(err)
	}
	if fltrRplEngine1Cfg, err = config.NewCGRConfigFromPath(fltrRplEngine1CfgPath); err != nil {
		t.Fatal(err)
	}
	if fltrRplEngine2Cfg, err = config.NewCGRConfigFromPath(fltrRplEngine2CfgPath); err != nil {
		t.Fatal(err)
	}
}

func testFltrRplInitDBs(t *testing.T) {
	if err := engine.InitDataDb(fltrRplEngine1Cfg); err != nil {
		t.Fatal(err)
	}

	if err := engine.InitStorDb(fltrRplEngine1Cfg); err != nil {
		t.Fatal(err)
	}

	if err := engine.InitDataDb(fltrRplEngine2Cfg); err != nil {
		t.Fatal(err)
	}

	if err := engine.InitStorDb(fltrRplEngine2Cfg); err != nil {
		t.Fatal(err)
	}
}

func testFltrRplStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(fltrRplInternalCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
	if _, err := engine.StartEngine(fltrRplEngine1CfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
	if _, err := engine.StartEngine(fltrRplEngine2CfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

func testFltrRplRPCConn(t *testing.T) {
	tmp := *utils.Encoding
	// run only under *gob encoding
	*utils.Encoding = utils.MetaGOB
	fltrRplInternalRPC = engine.NewRPCClient(t, fltrRplInternalCfg.ListenCfg())
	fltrRplEngine1RPC = engine.NewRPCClient(t, fltrRplEngine1Cfg.ListenCfg())
	fltrRplEngine2RPC = engine.NewRPCClient(t, fltrRplEngine2Cfg.ListenCfg())
	*utils.Encoding = tmp
}

func testFltrRplAttributeProfile(t *testing.T) {
	attrID := "ATTR1"
	attrPrf := &engine.AttributeProfileWithAPIOpts{
		AttributeProfile: &engine.AttributeProfile{
			Tenant:    "cgrates.org",
			ID:        attrID,
			Contexts:  []string{utils.MetaAny},
			FilterIDs: []string{"*string:~*req.Account:dan"},
			Attributes: []*engine.Attribute{
				{
					Path:  "*req.Category",
					Value: config.NewRSRParsersMustCompile(utils.MetaVoice, utils.InfieldSep),
				},
			},
			Weight: 10,
		},
	}
	var result string
	var replyPrfl *engine.AttributeProfile
	var rplyIDs []string
	// empty
	if err := fltrRplEngine1RPC.Call(context.Background(), utils.APIerSv1GetAttributeProfileIDs, &utils.PaginatorWithTenant{}, &rplyIDs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}
	if err := fltrRplEngine2RPC.Call(context.Background(), utils.APIerSv1GetAttributeProfileIDs, &utils.PaginatorWithTenant{}, &rplyIDs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}

	if err := fltrRplInternalRPC.Call(context.Background(), utils.APIerSv1SetAttributeProfile, attrPrf, &result); err != nil {
		t.Fatal(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := fltrRplInternalRPC.Call(context.Background(), utils.APIerSv1GetAttributeProfile,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: attrID}}, &replyPrfl); err != nil {
		t.Fatal(err)
	}
	replyPrfl.Compile()
	if !reflect.DeepEqual(attrPrf.AttributeProfile, replyPrfl) {
		t.Errorf("Expecting : %s, received: %s", utils.ToJSON(attrPrf.AttributeProfile), utils.ToJSON(replyPrfl))
	}
	replyPrfl = nil

	if err := fltrRplEngine1RPC.Call(context.Background(), utils.APIerSv1GetAttributeProfileIDs, &utils.PaginatorWithTenant{}, &rplyIDs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}
	if err := fltrRplEngine2RPC.Call(context.Background(), utils.APIerSv1GetAttributeProfileIDs, &utils.PaginatorWithTenant{}, &rplyIDs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}

	if err := fltrRplEngine1RPC.Call(context.Background(), utils.APIerSv1GetAttributeProfile,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: attrID}}, &replyPrfl); err != nil {
		t.Fatal(err)
	}
	replyPrfl.Compile()
	if !reflect.DeepEqual(attrPrf.AttributeProfile, replyPrfl) {
		t.Errorf("Expecting : %s, received: %s", utils.ToJSON(attrPrf.AttributeProfile), utils.ToJSON(replyPrfl))
	}
	replyPrfl = nil
	attrPrf.Weight = 15
	if err := fltrRplInternalRPC.Call(context.Background(), utils.APIerSv1SetAttributeProfile, attrPrf, &result); err != nil {
		t.Fatal(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := fltrRplInternalRPC.Call(context.Background(), utils.APIerSv1GetAttributeProfile,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: attrID}}, &replyPrfl); err != nil {
		t.Fatal(err)
	}
	replyPrfl.Compile()
	if !reflect.DeepEqual(attrPrf.AttributeProfile, replyPrfl) {
		t.Errorf("Expecting : %s, received: %s", utils.ToJSON(attrPrf.AttributeProfile), utils.ToJSON(replyPrfl))
	}
	replyPrfl = nil

	// use replicator to see if the attribute was changed in the DB
	if err := fltrRplEngine1RPC.Call(context.Background(), utils.ReplicatorSv1GetAttributeProfile,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: attrID}}, &replyPrfl); err != nil {
		t.Fatal(err)
	}
	replyPrfl.Compile()
	if !reflect.DeepEqual(attrPrf.AttributeProfile, replyPrfl) {
		t.Errorf("Expecting : %s, received: %s", utils.ToJSON(attrPrf.AttributeProfile), utils.ToJSON(replyPrfl))
	}

	if err := fltrRplEngine2RPC.Call(context.Background(), utils.APIerSv1GetAttributeProfileIDs, &utils.PaginatorWithTenant{}, &rplyIDs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}

	if err := fltrRplInternalRPC.Call(context.Background(), utils.APIerSv1RemoveAttributeProfile,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: attrID}}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := fltrRplEngine1RPC.Call(context.Background(), utils.APIerSv1GetAttributeProfileIDs, &utils.PaginatorWithTenant{}, &rplyIDs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}
	if err := fltrRplEngine2RPC.Call(context.Background(), utils.APIerSv1GetAttributeProfileIDs, &utils.PaginatorWithTenant{}, &rplyIDs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func testFltrRplFilters(t *testing.T) {
	fltrID := "FLTR1"
	fltr := &engine.FilterWithAPIOpts{
		Filter: &engine.Filter{
			Tenant: "cgrates.org",
			ID:     fltrID,
			Rules: []*engine.FilterRule{{
				Element: "~*req.Account",
				Type:    utils.MetaString,
				Values:  []string{"dan"},
			}},
		},
	}
	fltr.Compile()
	var result string
	var replyPrfl *engine.Filter
	var rplyIDs []string
	// empty
	if err := fltrRplEngine1RPC.Call(context.Background(), utils.APIerSv1GetFilterIDs, &utils.PaginatorWithTenant{}, &rplyIDs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}
	if err := fltrRplEngine2RPC.Call(context.Background(), utils.APIerSv1GetFilterIDs, &utils.PaginatorWithTenant{}, &rplyIDs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}

	if err := fltrRplInternalRPC.Call(context.Background(), utils.APIerSv1SetFilter, fltr, &result); err != nil {
		t.Fatal(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := fltrRplInternalRPC.Call(context.Background(), utils.APIerSv1GetFilter,
		&utils.TenantID{Tenant: "cgrates.org", ID: fltrID}, &replyPrfl); err != nil {
		t.Fatal(err)
	}
	replyPrfl.Compile()
	if !reflect.DeepEqual(fltr.Filter, replyPrfl) {
		t.Errorf("Expecting : %s, received: %s", utils.ToJSON(fltr.Filter), utils.ToJSON(replyPrfl))
	}
	replyPrfl = nil

	if err := fltrRplEngine1RPC.Call(context.Background(), utils.APIerSv1GetFilterIDs, &utils.PaginatorWithTenant{}, &rplyIDs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}
	if err := fltrRplEngine2RPC.Call(context.Background(), utils.APIerSv1GetFilterIDs, &utils.PaginatorWithTenant{}, &rplyIDs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}

	if err := fltrRplEngine1RPC.Call(context.Background(), utils.APIerSv1GetFilter,
		&utils.TenantID{Tenant: "cgrates.org", ID: fltrID}, &replyPrfl); err != nil {
		t.Fatal(err)
	}
	replyPrfl.Compile()
	if !reflect.DeepEqual(fltr.Filter, replyPrfl) {
		t.Errorf("Expecting : %s, received: %s", utils.ToJSON(fltr.Filter), utils.ToJSON(replyPrfl))
	}
	replyPrfl = nil
	fltr.Rules[0].Type = utils.MetaPrefix
	if err := fltrRplInternalRPC.Call(context.Background(), utils.APIerSv1SetFilter, fltr, &result); err != nil {
		t.Fatal(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := fltrRplInternalRPC.Call(context.Background(), utils.APIerSv1GetFilter,
		&utils.TenantID{Tenant: "cgrates.org", ID: fltrID}, &replyPrfl); err != nil {
		t.Fatal(err)
	}
	replyPrfl.Compile()
	if !reflect.DeepEqual(fltr.Filter, replyPrfl) {
		t.Errorf("Expecting : %s, received: %s", utils.ToJSON(fltr.Filter), utils.ToJSON(replyPrfl))
	}
	replyPrfl = nil

	// use replicator to see if the attribute was changed in the DB
	if err := fltrRplEngine1RPC.Call(context.Background(), utils.ReplicatorSv1GetFilter,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: fltrID}}, &replyPrfl); err != nil {
		t.Fatal(err)
	}
	replyPrfl.Compile()
	if !reflect.DeepEqual(fltr.Filter, replyPrfl) {
		t.Errorf("Expecting : %s, received: %s", utils.ToJSON(fltr.Filter), utils.ToJSON(replyPrfl))
	}

	if err := fltrRplEngine2RPC.Call(context.Background(), utils.APIerSv1GetFilterIDs, &utils.PaginatorWithTenant{}, &rplyIDs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}

	if err := fltrRplInternalRPC.Call(context.Background(), utils.APIerSv1RemoveFilter,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: fltrID}}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := fltrRplEngine1RPC.Call(context.Background(), utils.APIerSv1GetFilterIDs, &utils.PaginatorWithTenant{}, &rplyIDs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}
	if err := fltrRplEngine2RPC.Call(context.Background(), utils.APIerSv1GetFilterIDs, &utils.PaginatorWithTenant{}, &rplyIDs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func testFltrRplThresholdProfile(t *testing.T) {
	thID := "TH1"
	thPrfl := &engine.ThresholdProfileWithAPIOpts{
		ThresholdProfile: &engine.ThresholdProfile{
			Tenant:    "cgrates.org",
			ID:        thID,
			FilterIDs: []string{"*string:~*req.Account:dan"},
			MaxHits:   -1,
			Weight:    20,
		},
	}
	th := engine.Threshold{
		Tenant: "cgrates.org",
		ID:     thID,
	}
	var result string
	var replyPrfl *engine.ThresholdProfile
	var rplyIDs []string
	var replyTh engine.Threshold

	argsTh := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "cgrates.org",
			ID:     thID,
		},
	}
	// empty
	if err := fltrRplEngine1RPC.Call(context.Background(), utils.APIerSv1GetThresholdProfileIDs, &utils.PaginatorWithTenant{}, &rplyIDs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}
	if err := fltrRplEngine2RPC.Call(context.Background(), utils.APIerSv1GetThresholdProfileIDs, &utils.PaginatorWithTenant{}, &rplyIDs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}

	if err := fltrRplEngine1RPC.Call(context.Background(), utils.ReplicatorSv1GetThreshold, argsTh, &replyTh); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}
	if err := fltrRplEngine2RPC.Call(context.Background(), utils.ReplicatorSv1GetThreshold, argsTh, &replyTh); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}

	if err := fltrRplInternalRPC.Call(context.Background(), utils.APIerSv1SetThresholdProfile, thPrfl, &result); err != nil {
		t.Fatal(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := fltrRplInternalRPC.Call(context.Background(), utils.APIerSv1GetThresholdProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: thID}, &replyPrfl); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(thPrfl.ThresholdProfile, replyPrfl) {
		t.Errorf("Expecting : %s, received: %s", utils.ToJSON(thPrfl.ThresholdProfile), utils.ToJSON(replyPrfl))
	}
	replyPrfl = nil

	if err := fltrRplEngine1RPC.Call(context.Background(), utils.APIerSv1GetThresholdProfileIDs, &utils.PaginatorWithTenant{}, &rplyIDs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}
	if err := fltrRplEngine2RPC.Call(context.Background(), utils.APIerSv1GetThresholdProfileIDs, &utils.PaginatorWithTenant{}, &rplyIDs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}

	if err := fltrRplEngine1RPC.Call(context.Background(), utils.ReplicatorSv1GetThreshold, argsTh, &replyTh); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}
	if err := fltrRplEngine2RPC.Call(context.Background(), utils.ReplicatorSv1GetThreshold, argsTh, &replyTh); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}

	if err := fltrRplEngine1RPC.Call(context.Background(), utils.APIerSv1GetThresholdProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: thID}, &replyPrfl); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(thPrfl.ThresholdProfile, replyPrfl) {
		t.Errorf("Expecting : %s, received: %s", utils.ToJSON(thPrfl.ThresholdProfile), utils.ToJSON(replyPrfl))
	}

	if err := fltrRplEngine1RPC.Call(context.Background(), utils.ThresholdSv1GetThreshold, argsTh, &replyTh); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(th, replyTh) {
		t.Errorf("Expecting : %s, received: %s", utils.ToJSON(th), utils.ToJSON(replyTh))
	}

	replyPrfl = nil
	thPrfl.Weight = 10
	if err := fltrRplInternalRPC.Call(context.Background(), utils.APIerSv1SetThresholdProfile, thPrfl, &result); err != nil {
		t.Fatal(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := fltrRplInternalRPC.Call(context.Background(), utils.APIerSv1GetThresholdProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: thID}, &replyPrfl); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(thPrfl.ThresholdProfile, replyPrfl) {
		t.Errorf("Expecting : %s, received: %s", utils.ToJSON(thPrfl.ThresholdProfile), utils.ToJSON(replyPrfl))
	}
	replyPrfl = nil

	// use replicator to see if the attribute was changed in the DB
	if err := fltrRplEngine1RPC.Call(context.Background(), utils.ReplicatorSv1GetThresholdProfile,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: thID}}, &replyPrfl); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(thPrfl.ThresholdProfile, replyPrfl) {
		t.Errorf("Expecting : %s, received: %s", utils.ToJSON(thPrfl.ThresholdProfile), utils.ToJSON(replyPrfl))
	}

	if err := fltrRplEngine2RPC.Call(context.Background(), utils.APIerSv1GetThresholdProfileIDs, &utils.PaginatorWithTenant{}, &rplyIDs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}

	tEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "event1",
		Event: map[string]any{
			utils.AccountField: "dan",
		},
	}
	var thIDs []string
	//Testing ProcessEvent on set thresholdprofile using apier
	if err := fltrRplInternalRPC.Call(context.Background(), utils.ThresholdSv1ProcessEvent, tEv, &thIDs); err != nil {
		t.Fatal(err)
	} else if expected := []string{thID}; !reflect.DeepEqual(expected, thIDs) {
		t.Errorf("Expecting : %s, received: %s", utils.ToJSON(expected), utils.ToJSON(thIDs))
	}

	if err := fltrRplEngine1RPC.Call(context.Background(), utils.ThresholdSv1GetThreshold, argsTh, &replyTh); err != nil {
		t.Fatal(err)
	}
	th.Hits = 1
	replyTh.Snooze = th.Snooze // ignore the snooze as this is relative to time.Now
	if !reflect.DeepEqual(th, replyTh) {
		t.Errorf("Expecting : %s, received: %s", utils.ToJSON(th), utils.ToJSON(replyTh))
	}

	if err := fltrRplInternalRPC.Call(context.Background(), utils.APIerSv1RemoveThresholdProfile,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: thID}}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := fltrRplEngine1RPC.Call(context.Background(), utils.APIerSv1GetThresholdProfileIDs, &utils.PaginatorWithTenant{}, &rplyIDs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}
	if err := fltrRplEngine2RPC.Call(context.Background(), utils.APIerSv1GetThresholdProfileIDs, &utils.PaginatorWithTenant{}, &rplyIDs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}

	if err := fltrRplEngine1RPC.Call(context.Background(), utils.ReplicatorSv1GetThreshold, argsTh, &replyTh); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}
	if err := fltrRplEngine2RPC.Call(context.Background(), utils.ReplicatorSv1GetThreshold, argsTh, &replyTh); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func testFltrRplStatQueueProfile(t *testing.T) {
	stID := "ST1"
	stPrf := &engine.StatQueueProfileWithAPIOpts{
		StatQueueProfile: &engine.StatQueueProfile{
			Tenant:      "cgrates.org",
			ID:          stID,
			FilterIDs:   []string{"*string:~*req.Account:dan"},
			QueueLength: 100,
			TTL:         time.Second,
			Metrics: []*engine.MetricWithFilters{
				{
					MetricID: utils.MetaACD,
				},
			},
			ThresholdIDs: []string{"*none"},
			Blocker:      true,
			Stored:       true,
			Weight:       20,
			MinItems:     1,
		},
	}
	sq := engine.StatQueue{
		Tenant:    "cgrates.org",
		ID:        stID,
		SQItems:   []engine.SQItem{},
		SQMetrics: map[string]engine.StatMetric{},
	}
	var result string
	var replyPrfl *engine.StatQueueProfile
	var rplyIDs []string
	var replySq engine.StatQueue

	argsSq := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "cgrates.org",
			ID:     stID,
		},
	}
	// empty
	if err := fltrRplEngine1RPC.Call(context.Background(), utils.APIerSv1GetStatQueueProfileIDs, &utils.PaginatorWithTenant{}, &rplyIDs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}
	if err := fltrRplEngine2RPC.Call(context.Background(), utils.APIerSv1GetStatQueueProfileIDs, &utils.PaginatorWithTenant{}, &rplyIDs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}

	if err := fltrRplEngine1RPC.Call(context.Background(), utils.ReplicatorSv1GetStatQueue, argsSq, &replySq); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}
	if err := fltrRplEngine2RPC.Call(context.Background(), utils.ReplicatorSv1GetStatQueue, argsSq, &replySq); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}

	if err := fltrRplInternalRPC.Call(context.Background(), utils.APIerSv1SetStatQueueProfile, stPrf, &result); err != nil {
		t.Fatal(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := fltrRplInternalRPC.Call(context.Background(), utils.APIerSv1GetStatQueueProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: stID}, &replyPrfl); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(stPrf.StatQueueProfile, replyPrfl) {
		t.Errorf("Expecting : %s, received: %s", utils.ToJSON(stPrf.StatQueueProfile), utils.ToJSON(replyPrfl))
	}
	replyPrfl = nil

	if err := fltrRplEngine1RPC.Call(context.Background(), utils.APIerSv1GetStatQueueProfileIDs, &utils.PaginatorWithTenant{}, &rplyIDs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}
	if err := fltrRplEngine2RPC.Call(context.Background(), utils.APIerSv1GetStatQueueProfileIDs, &utils.PaginatorWithTenant{}, &rplyIDs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}

	if err := fltrRplEngine1RPC.Call(context.Background(), utils.ReplicatorSv1GetStatQueue, argsSq, &replySq); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}
	if err := fltrRplEngine2RPC.Call(context.Background(), utils.ReplicatorSv1GetStatQueue, argsSq, &replySq); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}

	if err := fltrRplEngine1RPC.Call(context.Background(), utils.APIerSv1GetStatQueueProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: stID}, &replyPrfl); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(stPrf.StatQueueProfile, replyPrfl) {
		t.Errorf("Expecting : %s, received: %s", utils.ToJSON(stPrf.StatQueueProfile), utils.ToJSON(replyPrfl))
	}
	replySq = engine.StatQueue{}
	sq.SQItems = nil
	s, _ := engine.NewACD(1, "", nil)
	sq.SQMetrics = map[string]engine.StatMetric{
		utils.MetaACD: s,
	}
	if err := fltrRplEngine1RPC.Call(context.Background(), utils.StatSv1GetStatQueue, argsSq, &replySq); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(sq, replySq) {
		t.Errorf("Expecting : %s, received: %s", utils.ToJSON(sq), utils.ToJSON(replySq))
	}

	replyPrfl = nil
	stPrf.Weight = 15
	if err := fltrRplInternalRPC.Call(context.Background(), utils.APIerSv1SetStatQueueProfile, stPrf, &result); err != nil {
		t.Fatal(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := fltrRplInternalRPC.Call(context.Background(), utils.APIerSv1GetStatQueueProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: stID}, &replyPrfl); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(stPrf.StatQueueProfile, replyPrfl) {
		t.Errorf("Expecting : %s, received: %s", utils.ToJSON(stPrf.StatQueueProfile), utils.ToJSON(replyPrfl))
	}
	replyPrfl = nil

	// use replicator to see if the attribute was changed in the DB
	if err := fltrRplEngine1RPC.Call(context.Background(), utils.ReplicatorSv1GetStatQueueProfile,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: stID}}, &replyPrfl); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(stPrf.StatQueueProfile, replyPrfl) {
		t.Errorf("Expecting : %s, received: %s", utils.ToJSON(stPrf.StatQueueProfile), utils.ToJSON(replyPrfl))
	}

	if err := fltrRplEngine2RPC.Call(context.Background(), utils.APIerSv1GetStatQueueProfileIDs, &utils.PaginatorWithTenant{}, &rplyIDs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}

	sEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "event1",
		Event: map[string]any{
			utils.AccountField: "dan",
			utils.Usage:        45 * time.Second,
		},
	}
	var sqIDs []string
	//Testing ProcessEvent on set thresholdprofile using apier
	if err := fltrRplInternalRPC.Call(context.Background(), utils.StatSv1ProcessEvent, sEv, &sqIDs); err != nil {
		t.Fatal(err)
	} else if expected := []string{stID}; !reflect.DeepEqual(expected, sqIDs) {
		t.Errorf("Expecting : %s, received: %s", utils.ToJSON(expected), utils.ToJSON(sqIDs))
	}

	if err := fltrRplEngine1RPC.Call(context.Background(), utils.StatSv1GetStatQueue, argsSq, &replySq); err != nil {
		t.Fatal(err)
	}
	sq.SQItems = []engine.SQItem{{
		EventID: "event1",
	}}
	s.AddEvent("event1", utils.MapStorage{utils.MetaReq: map[string]any{utils.Usage: 45 * time.Second}})
	replySq.SQItems[0].ExpiryTime = sq.SQItems[0].ExpiryTime
	if !reflect.DeepEqual(sq, replySq) {
		t.Errorf("Expecting : %s, received: %s", utils.ToJSON(sq), utils.ToJSON(replySq))
	}

	if err := fltrRplInternalRPC.Call(context.Background(), utils.APIerSv1RemoveStatQueueProfile,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: stID}}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := fltrRplEngine1RPC.Call(context.Background(), utils.APIerSv1GetStatQueueProfileIDs, &utils.PaginatorWithTenant{}, &rplyIDs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}
	if err := fltrRplEngine2RPC.Call(context.Background(), utils.APIerSv1GetStatQueueProfileIDs, &utils.PaginatorWithTenant{}, &rplyIDs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}

	if err := fltrRplEngine1RPC.Call(context.Background(), utils.ReplicatorSv1GetStatQueue, argsSq, &replySq); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}
	if err := fltrRplEngine2RPC.Call(context.Background(), utils.ReplicatorSv1GetStatQueue, argsSq, &replySq); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func testFltrRplResourceProfile(t *testing.T) {
	resID := "RES1"
	resPrf := &engine.ResourceProfileWithAPIOpts{
		ResourceProfile: &engine.ResourceProfile{
			Tenant:            "cgrates.org",
			ID:                resID,
			FilterIDs:         []string{"*string:~*req.Account:dan"},
			UsageTTL:          time.Nanosecond,
			Limit:             10,
			AllocationMessage: "MessageAllocation",
			Stored:            true,
			Weight:            20,
			ThresholdIDs:      []string{utils.MetaNone},
		},
	}
	rs := engine.Resource{
		Tenant: "cgrates.org",
		ID:     resID,
		Usages: make(map[string]*engine.ResourceUsage),
	}
	var result string
	var replyPrfl *engine.ResourceProfile
	var rplyIDs []string
	var replyRs engine.Resource

	argsRs := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "cgrates.org",
			ID:     resID,
		},
	}
	// empty
	if err := fltrRplEngine1RPC.Call(context.Background(), utils.APIerSv1GetResourceProfileIDs, &utils.PaginatorWithTenant{}, &rplyIDs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}
	if err := fltrRplEngine2RPC.Call(context.Background(), utils.APIerSv1GetResourceProfileIDs, &utils.PaginatorWithTenant{}, &rplyIDs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}

	if err := fltrRplEngine1RPC.Call(context.Background(), utils.ReplicatorSv1GetResource, argsRs, &replyRs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}
	if err := fltrRplEngine2RPC.Call(context.Background(), utils.ReplicatorSv1GetResource, argsRs, &replyRs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}
	if err := fltrRplInternalRPC.Call(context.Background(), utils.APIerSv1SetResourceProfile, resPrf, &result); err != nil {
		t.Fatal(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := fltrRplInternalRPC.Call(context.Background(), utils.APIerSv1GetResourceProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: resID}, &replyPrfl); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(resPrf.ResourceProfile, replyPrfl) {
		t.Errorf("Expecting : %s, received: %s", utils.ToJSON(resPrf.ResourceProfile), utils.ToJSON(replyPrfl))
	}
	replyPrfl = nil

	if err := fltrRplEngine1RPC.Call(context.Background(), utils.APIerSv1GetResourceProfileIDs, &utils.PaginatorWithTenant{}, &rplyIDs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}
	if err := fltrRplEngine2RPC.Call(context.Background(), utils.APIerSv1GetResourceProfileIDs, &utils.PaginatorWithTenant{}, &rplyIDs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}

	if err := fltrRplEngine1RPC.Call(context.Background(), utils.ReplicatorSv1GetResource, argsRs, &replyRs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}
	if err := fltrRplEngine2RPC.Call(context.Background(), utils.ReplicatorSv1GetResource, argsRs, &replyRs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}

	if err := fltrRplEngine1RPC.Call(context.Background(), utils.APIerSv1GetResourceProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: resID}, &replyPrfl); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(resPrf.ResourceProfile, replyPrfl) {
		t.Errorf("Expecting : %s, received: %s", utils.ToJSON(resPrf.ResourceProfile), utils.ToJSON(replyPrfl))
	}

	if err := fltrRplEngine1RPC.Call(context.Background(), utils.ResourceSv1GetResource, argsRs, &replyRs); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(rs, replyRs) {
		t.Errorf("Expecting : %s, received: %s", utils.ToJSON(rs), utils.ToJSON(replyRs))
	}

	replyPrfl = nil
	resPrf.Weight = 15
	if err := fltrRplInternalRPC.Call(context.Background(), utils.APIerSv1SetResourceProfile, resPrf, &result); err != nil {
		t.Fatal(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := fltrRplInternalRPC.Call(context.Background(), utils.APIerSv1GetResourceProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: resID}, &replyPrfl); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(resPrf.ResourceProfile, replyPrfl) {
		t.Errorf("Expecting : %s, received: %s", utils.ToJSON(resPrf.ResourceProfile), utils.ToJSON(replyPrfl))
	}
	replyPrfl = nil

	// use replicator to see if the attribute was changed in the DB
	if err := fltrRplEngine1RPC.Call(context.Background(), utils.ReplicatorSv1GetResourceProfile,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: resID}}, &replyPrfl); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(resPrf.ResourceProfile, replyPrfl) {
		t.Errorf("Expecting : %s, received: %s", utils.ToJSON(resPrf.ResourceProfile), utils.ToJSON(replyPrfl))
	}

	if err := fltrRplEngine2RPC.Call(context.Background(), utils.APIerSv1GetResourceProfileIDs, &utils.PaginatorWithTenant{}, &rplyIDs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}

	rEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     utils.UUIDSha1Prefix(),
		Event: map[string]any{
			utils.AccountField: "dan",
		},
		APIOpts: map[string]any{
			utils.OptsResourcesUsageID: "651a8db2-4f67-4cf8-b622-169e8a482e61",
			utils.OptsResourcesUnits:   6,
		},
	}
	var rsIDs string
	//Testing ProcessEvent on set thresholdprofile using apier
	if err := fltrRplInternalRPC.Call(context.Background(), utils.ResourceSv1AllocateResources, rEv, &rsIDs); err != nil {
		t.Fatal(err)
	} else if expected := resPrf.AllocationMessage; !reflect.DeepEqual(expected, rsIDs) {
		t.Errorf("Expecting : %s, received: %s", utils.ToJSON(expected), utils.ToJSON(rsIDs))
	}

	if err := fltrRplEngine1RPC.Call(context.Background(), utils.ResourceSv1GetResource, argsRs, &replyRs); err != nil {
		t.Fatal(err)
	}
	usageID := utils.IfaceAsString(rEv.APIOpts[utils.OptsResourcesUsageID])
	rs.TTLIdx = []string{usageID}
	rs.Usages = map[string]*engine.ResourceUsage{
		usageID: {
			Tenant: "cgrates.org",
			ID:     usageID,
			Units:  6,
		},
	}
	replyRs.Usages[usageID].ExpiryTime = time.Time{}
	if !reflect.DeepEqual(rs, replyRs) {
		t.Errorf("Expecting : %s, received: %s", utils.ToJSON(rs), utils.ToJSON(replyRs))
	}

	if err := fltrRplInternalRPC.Call(context.Background(), utils.APIerSv1RemoveResourceProfile,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: resID}}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := fltrRplEngine1RPC.Call(context.Background(), utils.APIerSv1GetResourceProfileIDs, &utils.PaginatorWithTenant{}, &rplyIDs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}
	if err := fltrRplEngine2RPC.Call(context.Background(), utils.APIerSv1GetResourceProfileIDs, &utils.PaginatorWithTenant{}, &rplyIDs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}

	if err := fltrRplEngine1RPC.Call(context.Background(), utils.ReplicatorSv1GetResource, argsRs, &replyRs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}
	if err := fltrRplEngine2RPC.Call(context.Background(), utils.ReplicatorSv1GetResource, argsRs, &replyRs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func testFltrRplRouteProfile(t *testing.T) {
	rpID := "RT1"
	rpPrf := &v1.RouteWithAPIOpts{
		RouteProfile: &engine.RouteProfile{
			Tenant:    "cgrates.org",
			ID:        rpID,
			FilterIDs: []string{"*string:~*req.Account:dan"},
			Sorting:   utils.MetaWeight,
			Routes: []*engine.Route{
				{
					ID:            "local",
					RatingPlanIDs: []string{"RP_LOCAL"},
					Weight:        10,
				},
				{
					ID:            "mobile",
					RatingPlanIDs: []string{"RP_MOBILE"},
					Weight:        30,
				},
			},
			Weight: 10,
		},
	}
	var result string
	var replyPrfl *engine.RouteProfile
	var rplyIDs []string
	// empty
	if err := fltrRplEngine1RPC.Call(context.Background(), utils.APIerSv1GetRouteProfileIDs, &utils.PaginatorWithTenant{}, &rplyIDs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}
	if err := fltrRplEngine2RPC.Call(context.Background(), utils.APIerSv1GetRouteProfileIDs, &utils.PaginatorWithTenant{}, &rplyIDs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}

	if err := fltrRplInternalRPC.Call(context.Background(), utils.APIerSv1SetRouteProfile, rpPrf, &result); err != nil {
		t.Fatal(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := fltrRplInternalRPC.Call(context.Background(), utils.APIerSv1GetRouteProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: rpID}, &replyPrfl); err != nil {
		t.Fatal(err)
	}
	replyPrfl.Compile()
	if !reflect.DeepEqual(rpPrf.RouteProfile, replyPrfl) {
		t.Errorf("Expecting : %s, received: %s", utils.ToJSON(rpPrf.RouteProfile), utils.ToJSON(replyPrfl))
	}
	replyPrfl = nil

	if err := fltrRplEngine1RPC.Call(context.Background(), utils.APIerSv1GetRouteProfileIDs, &utils.PaginatorWithTenant{}, &rplyIDs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}
	if err := fltrRplEngine2RPC.Call(context.Background(), utils.APIerSv1GetRouteProfileIDs, &utils.PaginatorWithTenant{}, &rplyIDs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}

	if err := fltrRplEngine1RPC.Call(context.Background(), utils.APIerSv1GetRouteProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: rpID}, &replyPrfl); err != nil {
		t.Fatal(err)
	}
	replyPrfl.Compile()
	if !reflect.DeepEqual(rpPrf.RouteProfile, replyPrfl) {
		t.Errorf("Expecting : %s, received: %s", utils.ToJSON(rpPrf.RouteProfile), utils.ToJSON(replyPrfl))
	}
	replyPrfl = nil
	rpPrf.Weight = 15
	if err := fltrRplInternalRPC.Call(context.Background(), utils.APIerSv1SetRouteProfile, rpPrf, &result); err != nil {
		t.Fatal(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := fltrRplInternalRPC.Call(context.Background(), utils.APIerSv1GetRouteProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: rpID}, &replyPrfl); err != nil {
		t.Fatal(err)
	}
	replyPrfl.Compile()
	if !reflect.DeepEqual(rpPrf.RouteProfile, replyPrfl) {
		t.Errorf("Expecting : %s, received: %s", utils.ToJSON(rpPrf.RouteProfile), utils.ToJSON(replyPrfl))
	}
	replyPrfl = nil

	// use replicator to see if the attribute was changed in the DB
	if err := fltrRplEngine1RPC.Call(context.Background(), utils.ReplicatorSv1GetRouteProfile,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: rpID}}, &replyPrfl); err != nil {
		t.Fatal(err)
	}
	replyPrfl.Compile()
	if !reflect.DeepEqual(rpPrf.RouteProfile, replyPrfl) {
		t.Errorf("Expecting : %s, received: %s", utils.ToJSON(rpPrf.RouteProfile), utils.ToJSON(replyPrfl))
	}

	if err := fltrRplEngine2RPC.Call(context.Background(), utils.APIerSv1GetRouteProfileIDs, &utils.PaginatorWithTenant{}, &rplyIDs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}

	if err := fltrRplInternalRPC.Call(context.Background(), utils.APIerSv1RemoveRouteProfile,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: rpID}}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := fltrRplEngine1RPC.Call(context.Background(), utils.APIerSv1GetRouteProfileIDs, &utils.PaginatorWithTenant{}, &rplyIDs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}
	if err := fltrRplEngine2RPC.Call(context.Background(), utils.APIerSv1GetRouteProfileIDs, &utils.PaginatorWithTenant{}, &rplyIDs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func testFltrRplChargerProfile(t *testing.T) {
	chID := "CH1"
	chPrf := &v1.ChargerWithAPIOpts{
		ChargerProfile: &engine.ChargerProfile{
			Tenant:       "cgrates.org",
			ID:           chID,
			FilterIDs:    []string{"*string:~*req.Account:dan"},
			RunID:        utils.MetaDefault,
			AttributeIDs: []string{utils.MetaNone},
			Weight:       20,
		},
	}
	var result string
	var replyPrfl *engine.ChargerProfile
	var rplyIDs []string
	// empty
	if err := fltrRplEngine1RPC.Call(context.Background(), utils.APIerSv1GetChargerProfileIDs, &utils.PaginatorWithTenant{}, &rplyIDs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}
	if err := fltrRplEngine2RPC.Call(context.Background(), utils.APIerSv1GetChargerProfileIDs, &utils.PaginatorWithTenant{}, &rplyIDs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}

	if err := fltrRplInternalRPC.Call(context.Background(), utils.APIerSv1SetChargerProfile, chPrf, &result); err != nil {
		t.Fatal(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := fltrRplInternalRPC.Call(context.Background(), utils.APIerSv1GetChargerProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: chID}, &replyPrfl); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(chPrf.ChargerProfile, replyPrfl) {
		t.Errorf("Expecting : %s, received: %s", utils.ToJSON(chPrf.ChargerProfile), utils.ToJSON(replyPrfl))
	}
	replyPrfl = nil

	if err := fltrRplEngine1RPC.Call(context.Background(), utils.APIerSv1GetChargerProfileIDs, &utils.PaginatorWithTenant{}, &rplyIDs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}
	if err := fltrRplEngine2RPC.Call(context.Background(), utils.APIerSv1GetChargerProfileIDs, &utils.PaginatorWithTenant{}, &rplyIDs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}

	if err := fltrRplEngine1RPC.Call(context.Background(), utils.APIerSv1GetChargerProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: chID}, &replyPrfl); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(chPrf.ChargerProfile, replyPrfl) {
		t.Errorf("Expecting : %s, received: %s", utils.ToJSON(chPrf.ChargerProfile), utils.ToJSON(replyPrfl))
	}
	replyPrfl = nil
	chPrf.Weight = 15
	if err := fltrRplInternalRPC.Call(context.Background(), utils.APIerSv1SetChargerProfile, chPrf, &result); err != nil {
		t.Fatal(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := fltrRplInternalRPC.Call(context.Background(), utils.APIerSv1GetChargerProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: chID}, &replyPrfl); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(chPrf.ChargerProfile, replyPrfl) {
		t.Errorf("Expecting : %s, received: %s", utils.ToJSON(chPrf.ChargerProfile), utils.ToJSON(replyPrfl))
	}
	replyPrfl = nil

	// use replicator to see if the attribute was changed in the DB
	if err := fltrRplEngine1RPC.Call(context.Background(), utils.ReplicatorSv1GetChargerProfile,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: chID}}, &replyPrfl); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(chPrf.ChargerProfile, replyPrfl) {
		t.Errorf("Expecting : %s, received: %s", utils.ToJSON(chPrf.ChargerProfile), utils.ToJSON(replyPrfl))
	}

	if err := fltrRplEngine2RPC.Call(context.Background(), utils.APIerSv1GetChargerProfileIDs, &utils.PaginatorWithTenant{}, &rplyIDs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}

	if err := fltrRplInternalRPC.Call(context.Background(), utils.APIerSv1RemoveChargerProfile,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: chID}}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := fltrRplEngine1RPC.Call(context.Background(), utils.APIerSv1GetChargerProfileIDs, &utils.PaginatorWithTenant{}, &rplyIDs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}
	if err := fltrRplEngine2RPC.Call(context.Background(), utils.APIerSv1GetChargerProfileIDs, &utils.PaginatorWithTenant{}, &rplyIDs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func testFltrRplDispatcherProfile(t *testing.T) {
	dspID := "DSP1"
	dspPrf := &v1.DispatcherWithAPIOpts{
		DispatcherProfile: &engine.DispatcherProfile{
			Tenant:     "cgrates.org",
			ID:         dspID,
			Subsystems: []string{utils.MetaSessionS},
			FilterIDs:  []string{"*string:~*req.Account:dan"},
			Weight:     10,
		},
	}
	var result string
	var replyPrfl *engine.DispatcherProfile
	var rplyIDs []string
	// empty
	if err := fltrRplEngine1RPC.Call(context.Background(), utils.APIerSv1GetDispatcherProfileIDs, &utils.PaginatorWithTenant{}, &rplyIDs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}
	if err := fltrRplEngine2RPC.Call(context.Background(), utils.APIerSv1GetDispatcherProfileIDs, &utils.PaginatorWithTenant{}, &rplyIDs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}

	if err := fltrRplInternalRPC.Call(context.Background(), utils.APIerSv1SetDispatcherProfile, dspPrf, &result); err != nil {
		t.Fatal(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := fltrRplInternalRPC.Call(context.Background(), utils.APIerSv1GetDispatcherProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: dspID}, &replyPrfl); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(dspPrf.DispatcherProfile, replyPrfl) {
		t.Errorf("Expecting : %s, received: %s", utils.ToJSON(dspPrf.DispatcherProfile), utils.ToJSON(replyPrfl))
	}
	replyPrfl = nil

	if err := fltrRplEngine1RPC.Call(context.Background(), utils.APIerSv1GetDispatcherProfileIDs, &utils.PaginatorWithTenant{}, &rplyIDs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}
	if err := fltrRplEngine2RPC.Call(context.Background(), utils.APIerSv1GetDispatcherProfileIDs, &utils.PaginatorWithTenant{}, &rplyIDs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}

	if err := fltrRplEngine1RPC.Call(context.Background(), utils.APIerSv1GetDispatcherProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: dspID}, &replyPrfl); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(dspPrf.DispatcherProfile, replyPrfl) {
		t.Errorf("Expecting : %s, received: %s", utils.ToJSON(dspPrf.DispatcherProfile), utils.ToJSON(replyPrfl))
	}
	replyPrfl = nil
	dspPrf.Weight = 15
	if err := fltrRplInternalRPC.Call(context.Background(), utils.APIerSv1SetDispatcherProfile, dspPrf, &result); err != nil {
		t.Fatal(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := fltrRplInternalRPC.Call(context.Background(), utils.APIerSv1GetDispatcherProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: dspID}, &replyPrfl); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(dspPrf.DispatcherProfile, replyPrfl) {
		t.Errorf("Expecting : %s, received: %s", utils.ToJSON(dspPrf.DispatcherProfile), utils.ToJSON(replyPrfl))
	}
	replyPrfl = nil

	// use replicator to see if the attribute was changed in the DB
	if err := fltrRplEngine1RPC.Call(context.Background(), utils.ReplicatorSv1GetDispatcherProfile,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: dspID}}, &replyPrfl); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(dspPrf.DispatcherProfile, replyPrfl) {
		t.Errorf("Expecting : %s, received: %s", utils.ToJSON(dspPrf.DispatcherProfile), utils.ToJSON(replyPrfl))
	}

	if err := fltrRplEngine2RPC.Call(context.Background(), utils.APIerSv1GetDispatcherProfileIDs, &utils.PaginatorWithTenant{}, &rplyIDs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}

	if err := fltrRplInternalRPC.Call(context.Background(), utils.APIerSv1RemoveDispatcherProfile,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: dspID}}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := fltrRplEngine1RPC.Call(context.Background(), utils.APIerSv1GetDispatcherProfileIDs, &utils.PaginatorWithTenant{}, &rplyIDs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}
	if err := fltrRplEngine2RPC.Call(context.Background(), utils.APIerSv1GetDispatcherProfileIDs, &utils.PaginatorWithTenant{}, &rplyIDs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func testFltrRplDispatcherHost(t *testing.T) {
	dspID := "DSH1"
	dspPrf := &engine.DispatcherHostWithAPIOpts{
		DispatcherHost: &engine.DispatcherHost{
			Tenant: "cgrates.org",
			RemoteHost: &config.RemoteHost{
				ID:      dspID,
				Address: "*internal",
			},
		},
	}
	var result string
	var replyPrfl *engine.DispatcherHost
	var rplyIDs []string
	// empty
	if err := fltrRplEngine1RPC.Call(context.Background(), utils.APIerSv1GetDispatcherHostIDs, &utils.PaginatorWithTenant{}, &rplyIDs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}
	if err := fltrRplEngine2RPC.Call(context.Background(), utils.APIerSv1GetDispatcherHostIDs, &utils.PaginatorWithTenant{}, &rplyIDs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}

	if err := fltrRplInternalRPC.Call(context.Background(), utils.APIerSv1SetDispatcherHost, dspPrf, &result); err != nil {
		t.Fatal(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := fltrRplInternalRPC.Call(context.Background(), utils.APIerSv1GetDispatcherHost,
		&utils.TenantID{Tenant: "cgrates.org", ID: dspID}, &replyPrfl); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(dspPrf.DispatcherHost, replyPrfl) {
		t.Errorf("Expecting : %s, received: %s", utils.ToJSON(dspPrf.DispatcherHost), utils.ToJSON(replyPrfl))
	}
	replyPrfl = nil

	if err := fltrRplEngine1RPC.Call(context.Background(), utils.APIerSv1GetDispatcherHostIDs, &utils.PaginatorWithTenant{}, &rplyIDs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}
	if err := fltrRplEngine2RPC.Call(context.Background(), utils.APIerSv1GetDispatcherHostIDs, &utils.PaginatorWithTenant{}, &rplyIDs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}

	if err := fltrRplEngine1RPC.Call(context.Background(), utils.APIerSv1GetDispatcherHost,
		&utils.TenantID{Tenant: "cgrates.org", ID: dspID}, &replyPrfl); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(dspPrf.DispatcherHost, replyPrfl) {
		t.Errorf("Expecting : %s, received: %s", utils.ToJSON(dspPrf.DispatcherHost), utils.ToJSON(replyPrfl))
	}
	replyPrfl = nil
	dspPrf.Address = "127.0.0.1:2012"
	if err := fltrRplInternalRPC.Call(context.Background(), utils.APIerSv1SetDispatcherHost, dspPrf, &result); err != nil {
		t.Fatal(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := fltrRplInternalRPC.Call(context.Background(), utils.APIerSv1GetDispatcherHost,
		&utils.TenantID{Tenant: "cgrates.org", ID: dspID}, &replyPrfl); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(dspPrf.DispatcherHost, replyPrfl) {
		t.Errorf("Expecting : %s, received: %s", utils.ToJSON(dspPrf.DispatcherHost), utils.ToJSON(replyPrfl))
	}
	replyPrfl = nil

	// use replicator to see if the attribute was changed in the DB
	if err := fltrRplEngine1RPC.Call(context.Background(), utils.ReplicatorSv1GetDispatcherHost,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: dspID}}, &replyPrfl); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(dspPrf.DispatcherHost, replyPrfl) {
		t.Errorf("Expecting : %s, received: %s", utils.ToJSON(dspPrf.DispatcherHost), utils.ToJSON(replyPrfl))
	}

	if err := fltrRplEngine2RPC.Call(context.Background(), utils.APIerSv1GetDispatcherHostIDs, &utils.PaginatorWithTenant{}, &rplyIDs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}

	if err := fltrRplInternalRPC.Call(context.Background(), utils.APIerSv1RemoveDispatcherHost,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: dspID}}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := fltrRplEngine1RPC.Call(context.Background(), utils.APIerSv1GetDispatcherHostIDs, &utils.PaginatorWithTenant{}, &rplyIDs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}
	if err := fltrRplEngine2RPC.Call(context.Background(), utils.APIerSv1GetDispatcherHostIDs, &utils.PaginatorWithTenant{}, &rplyIDs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func testFltrRplAccount(t *testing.T) {
	acID := "ATTR1"
	attrPrf := &v2.AttrSetAccount{Tenant: "cgrates.org", Account: acID, ExtraOptions: map[string]bool{utils.Disabled: true}}
	attrAC := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: acID}
	expPrf := &engine.Account{
		ID:       "cgrates.org:" + acID,
		Disabled: true,
	}
	var result string
	var replyPrfl *engine.Account
	var rplyCount int
	// empty
	if err := fltrRplEngine1RPC.Call(context.Background(), utils.APIerSv1GetAccountsCount, &utils.PaginatorWithTenant{}, &rplyCount); err != nil {
		t.Fatal(err)
	} else if rplyCount != 0 {
		t.Fatal("Expected no accounts")
	}
	if err := fltrRplEngine2RPC.Call(context.Background(), utils.APIerSv1GetAccountsCount, &utils.PaginatorWithTenant{}, &rplyCount); err != nil {
		t.Fatal(err)
	} else if rplyCount != 0 {
		t.Fatal("Expected no accounts")
	}

	if err := fltrRplInternalRPC.Call(context.Background(), utils.APIerSv2SetAccount, attrPrf, &result); err != nil {
		t.Fatal(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := fltrRplInternalRPC.Call(context.Background(), utils.APIerSv2GetAccount, attrAC, &replyPrfl); err != nil {
		t.Fatal(err)
	}
	replyPrfl.BalanceMap = nil
	replyPrfl.UnitCounters = nil
	replyPrfl.ActionTriggers = nil
	replyPrfl.UpdateTime = expPrf.UpdateTime

	if !reflect.DeepEqual(expPrf, replyPrfl) {
		t.Errorf("Expecting : %s, received: %s", utils.ToJSON(expPrf), utils.ToJSON(replyPrfl))
	}
	replyPrfl = nil

	if err := fltrRplEngine1RPC.Call(context.Background(), utils.APIerSv1GetAccountsCount, &utils.PaginatorWithTenant{}, &rplyCount); err != nil {
		t.Fatal(err)
	} else if rplyCount != 0 {
		t.Fatal("Expected no accounts")
	}
	if err := fltrRplEngine2RPC.Call(context.Background(), utils.APIerSv1GetAccountsCount, &utils.PaginatorWithTenant{}, &rplyCount); err != nil {
		t.Fatal(err)
	} else if rplyCount != 0 {
		t.Fatal("Expected no accounts")
	}

	if err := fltrRplEngine1RPC.Call(context.Background(), utils.APIerSv2GetAccount, attrAC, &replyPrfl); err != nil {
		t.Fatal(err)
	}
	replyPrfl.BalanceMap = nil
	replyPrfl.UnitCounters = nil
	replyPrfl.ActionTriggers = nil
	replyPrfl.UpdateTime = expPrf.UpdateTime

	if !reflect.DeepEqual(expPrf, replyPrfl) {
		t.Errorf("Expecting : %s, received: %s", utils.ToJSON(expPrf), utils.ToJSON(replyPrfl))
	}
	replyPrfl = nil
	attrPrf.ExtraOptions[utils.Disabled] = false
	expPrf.Disabled = false
	if err := fltrRplInternalRPC.Call(context.Background(), utils.APIerSv2SetAccount, attrPrf, &result); err != nil {
		t.Fatal(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := fltrRplInternalRPC.Call(context.Background(), utils.APIerSv2GetAccount, attrAC, &replyPrfl); err != nil {
		t.Fatal(err)
	}
	replyPrfl.BalanceMap = nil
	replyPrfl.UnitCounters = nil
	replyPrfl.ActionTriggers = nil
	replyPrfl.UpdateTime = expPrf.UpdateTime

	if !reflect.DeepEqual(expPrf, replyPrfl) {
		t.Errorf("Expecting : %s, received: %s", utils.ToJSON(expPrf), utils.ToJSON(replyPrfl))
	}
	replyPrfl = nil

	// use replicator to see if the attribute was changed in the DB
	if err := fltrRplEngine1RPC.Call(context.Background(), utils.ReplicatorSv1GetAccount, &utils.StringWithAPIOpts{
		Arg: expPrf.ID,
	}, &replyPrfl); err != nil {
		t.Fatal(err)
	}
	replyPrfl.BalanceMap = nil
	replyPrfl.UnitCounters = nil
	replyPrfl.ActionTriggers = nil
	replyPrfl.UpdateTime = expPrf.UpdateTime

	if !reflect.DeepEqual(expPrf, replyPrfl) {
		t.Errorf("Expecting : %s, received: %s", utils.ToJSON(expPrf), utils.ToJSON(replyPrfl))
	}

	if err := fltrRplEngine2RPC.Call(context.Background(), utils.APIerSv1GetAccountsCount, &utils.PaginatorWithTenant{}, &rplyCount); err != nil {
		t.Fatal(err)
	} else if rplyCount != 0 {
		t.Fatal("Expected no accounts")
	}
}

func testFltrRplDestination(t *testing.T) {
	dstID := "DST1"
	dstPrf := utils.AttrSetDestination{Id: dstID, Prefixes: []string{"dan"}}
	expPrf := &engine.Destination{
		Id:       dstID,
		Prefixes: []string{"dan"},
	}
	args := &utils.StringWithAPIOpts{
		Arg:    dstID,
		Tenant: "cgrates.org",
	}
	args2 := &utils.StringWithAPIOpts{
		Arg:    "dan",
		Tenant: "cgrates.org",
	}
	var result string
	var replyPrfl *engine.Destination
	var rplyIDs *engine.Destination
	var rplyIDs2 []string
	// empty
	if err := fltrRplEngine1RPC.Call(context.Background(), utils.ReplicatorSv1GetDestination, args, &rplyIDs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}
	if err := fltrRplEngine1RPC.Call(context.Background(), utils.ReplicatorSv1GetReverseDestination, args2, &rplyIDs2); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}
	if err := fltrRplEngine2RPC.Call(context.Background(), utils.ReplicatorSv1GetDestination, args, &rplyIDs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}
	if err := fltrRplEngine2RPC.Call(context.Background(), utils.ReplicatorSv1GetReverseDestination, args2, &rplyIDs2); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}

	if err := fltrRplInternalRPC.Call(context.Background(), utils.APIerSv1SetDestination, dstPrf, &result); err != nil {
		t.Fatal(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := fltrRplInternalRPC.Call(context.Background(), utils.APIerSv1GetDestination, dstID, &replyPrfl); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(expPrf, replyPrfl) {
		t.Errorf("Expecting : %s, received: %s", utils.ToJSON(expPrf), utils.ToJSON(replyPrfl))
	}
	replyPrfl = nil

	if err := fltrRplEngine1RPC.Call(context.Background(), utils.ReplicatorSv1GetDestination, args, &rplyIDs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}
	if err := fltrRplEngine1RPC.Call(context.Background(), utils.ReplicatorSv1GetReverseDestination, args2, &rplyIDs2); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}
	if err := fltrRplEngine2RPC.Call(context.Background(), utils.ReplicatorSv1GetDestination, args, &rplyIDs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}
	if err := fltrRplEngine2RPC.Call(context.Background(), utils.ReplicatorSv1GetReverseDestination, args2, &rplyIDs2); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}

	if err := fltrRplEngine1RPC.Call(context.Background(), utils.APIerSv1GetDestination, dstID, &replyPrfl); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(expPrf, replyPrfl) {
		t.Errorf("Expecting : %s, received: %s", utils.ToJSON(expPrf), utils.ToJSON(replyPrfl))
	}
	if err := fltrRplEngine1RPC.Call(context.Background(), utils.APIerSv1GetReverseDestination, "dan", &rplyIDs2); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual([]string{dstID}, rplyIDs2) {
		t.Errorf("Expecting : %s, received: %s", utils.ToJSON([]string{dstID}), utils.ToJSON(rplyIDs2))
	}
	replyPrfl = nil
	dstPrf.Overwrite = true
	dstPrf.Prefixes = []string{"dan2"}
	expPrf.Prefixes = []string{"dan2"}
	args2.Arg = "dan2"
	if err := fltrRplInternalRPC.Call(context.Background(), utils.APIerSv1SetDestination, dstPrf, &result); err != nil {
		t.Fatal(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := fltrRplInternalRPC.Call(context.Background(), utils.APIerSv1GetDestination, dstID, &replyPrfl); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(expPrf, replyPrfl) {
		t.Errorf("Expecting : %s, received: %s", utils.ToJSON(expPrf), utils.ToJSON(replyPrfl))
	}
	replyPrfl = nil

	// use replicator to see if the attribute was changed in the DB
	if err := fltrRplEngine1RPC.Call(context.Background(), utils.ReplicatorSv1GetDestination, args, &replyPrfl); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(expPrf, replyPrfl) {
		t.Errorf("Expecting : %s, received: %s", utils.ToJSON(expPrf), utils.ToJSON(replyPrfl))
	}
	// use replicator to see if the attribute was changed in the DB
	if err := fltrRplEngine1RPC.Call(context.Background(), utils.ReplicatorSv1GetReverseDestination, args2, &rplyIDs2); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual([]string{dstID}, rplyIDs2) {
		t.Errorf("Expecting : %s, received: %s", utils.ToJSON([]string{dstID}), utils.ToJSON(rplyIDs2))
	}

	if err := fltrRplEngine2RPC.Call(context.Background(), utils.ReplicatorSv1GetDestination, args, &rplyIDs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Unexpected error: %v", err)
	}
	rplyIDs2 = nil
	if err := fltrRplEngine2RPC.Call(context.Background(), utils.ReplicatorSv1GetReverseDestination, args2, &rplyIDs2); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Log(rplyIDs2)
		t.Fatalf("Unexpected error: %v", err)
	}
}

func testFltrRplKillEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
