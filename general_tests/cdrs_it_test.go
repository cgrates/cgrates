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

package general_tests

import (
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	cdrsCfgPath string
	cdrsCfg     *config.CGRConfig
	cdrsRpc     *birpc.Client
	cdrsConfDIR string // run the tests for specific configuration

	// subtests to be executed for each confDIR
	sTestsCDRsIT = []func(t *testing.T){
		testV2CDRsInitConfig,
		testV2CDRsInitDataDb,
		testV2CDRsInitCdrDb,
		testV2CDRsStartEngine,
		testV2CDRsRpcConn,
		testV2CDRsLoadTariffPlanFromFolder,
		//default process
		testV2CDRsProcessCDR,
		testV2CDRsGetCdrs,
		//custom process
		testV2CDRsProcessCDR2,
		testV2CDRsGetCdrs2,
		testV2CDRsProcessCDR3,
		testV2CDRsGetCdrs3,

		testV2CDRsProcessCDR4,
		testV2CDRsGetCdrs4,

		testV2CDRsSetStats,
		testV2CDRsSetThresholdProfile,

		testV2CDRsProcessCDR5,
		testV2CDRsGetCdrs5,
		testV2CDRsGetStats1,
		testV2CDRsGetThreshold1,
		testV2CDRsProcessCDR6,
		testV2CDRsGetCdrs5,
		testV2CDRsGetStats2,
		testV2CDRsGetThreshold2,
		testV2CDRsProcessCDR7,
		testV2CDRsGetCdrs7,

		testV2CDRsKillEngine,
	}
)

// Tests starting here
func TestCDRsIT(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		cdrsConfDIR = "cdrsv2internal"
	case utils.MetaMySQL:
		cdrsConfDIR = "cdrsv2mysql"
	case utils.MetaMongo:
		cdrsConfDIR = "cdrsv2mongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}

	for _, stest := range sTestsCDRsIT {
		t.Run(cdrsConfDIR, stest)
	}
}

func testV2CDRsInitConfig(t *testing.T) {
	var err error
	cdrsCfgPath = path.Join(*utils.DataDir, "conf", "samples", cdrsConfDIR)
	if *utils.Encoding == utils.MetaGOB {
		cdrsCfgPath = path.Join(*utils.DataDir, "conf", "samples", cdrsConfDIR+"_gob")
	}
	if cdrsCfg, err = config.NewCGRConfigFromPath(cdrsCfgPath); err != nil {
		t.Fatal("Got config error: ", err.Error())
	}
}

func testV2CDRsInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(cdrsCfg); err != nil {
		t.Fatal(err)
	}
}

// InitDb so we can rely on count
func testV2CDRsInitCdrDb(t *testing.T) {
	if err := engine.InitStorDb(cdrsCfg); err != nil {
		t.Fatal(err)
	}
}

func testV2CDRsStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(cdrsCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testV2CDRsRpcConn(t *testing.T) {
	cdrsRpc = engine.NewRPCClient(t, cdrsCfg.ListenCfg())
}

func testV2CDRsLoadTariffPlanFromFolder(t *testing.T) {
	var loadInst utils.LoadInstance
	if err := cdrsRpc.Call(context.Background(), utils.APIerSv2LoadTariffPlanFromFolder,
		&utils.AttrLoadTpFromFolder{FolderPath: path.Join(
			*utils.DataDir, "tariffplans", "testit")}, &loadInst); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(*utils.WaitRater) * time.Millisecond) // Give time for scheduler to execute topups
	var resp string
	if err := cdrsRpc.Call(context.Background(), utils.APIerSv1RemoveChargerProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "SupplierCharges"}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
	var reply *engine.ChargerProfile
	if err := cdrsRpc.Call(context.Background(), utils.APIerSv1GetChargerProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "SupplierCharges"},
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testV2CDRsProcessCDR(t *testing.T) {
	args := &engine.ArgV1ProcessEvent{
		Flags: []string{utils.MetaRALs},

		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]any{
				utils.OriginID:     "testV2CDRsProcessCDR1",
				utils.OriginHost:   "192.168.1.1",
				utils.Source:       "testV2CDRsProcessCDR",
				utils.RequestType:  utils.MetaRated,
				utils.Category:     "call",
				utils.AccountField: "testV2CDRsProcessCDR",
				utils.Subject:      "ANY2CNT",
				utils.Destination:  "+4986517174963",
				utils.AnswerTime:   time.Date(2018, 8, 24, 16, 00, 26, 0, time.UTC),
				utils.Usage:        time.Minute,
				"field_extr1":      "val_extr1",
				"fieldextr2":       "valextr2",
			},
		},
	}

	var reply string
	if err := cdrsRpc.Call(context.Background(), utils.CDRsV1ProcessEvent, args, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
}

func testV2CDRsGetCdrs(t *testing.T) {
	var cdrCnt int64
	req := utils.AttrGetCdrs{}
	if err := cdrsRpc.Call(context.Background(), utils.APIerSv2CountCDRs, &req, &cdrCnt); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if cdrCnt != 2 {
		t.Error("Unexpected number of CDRs returned: ", cdrCnt)
	}
	var cdrs []*engine.ExternalCDR
	args := utils.RPCCDRsFilter{RunIDs: []string{"raw"}}
	if err := cdrsRpc.Call(context.Background(), utils.APIerSv2GetCDRs, &args, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Cost != -1.0 {
			t.Errorf("Unexpected cost for CDR: %f", cdrs[0].Cost)
		}
		if cdrs[0].ExtraFields["PayPalAccount"] != "paypal@cgrates.org" {
			t.Errorf("PayPalAccount should be added by AttributeS, have: %s",
				cdrs[0].ExtraFields["PayPalAccount"])
		}
	}
	args = utils.RPCCDRsFilter{RunIDs: []string{"CustomerCharges"}}
	if err := cdrsRpc.Call(context.Background(), utils.APIerSv2GetCDRs, &args, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Cost != 0.0198 {
			t.Errorf("Unexpected cost for CDR: %f", cdrs[0].Cost)
		}
		if cdrs[0].ExtraFields["PayPalAccount"] != "paypal@cgrates.org" {
			t.Errorf("PayPalAccount should be added by AttributeS, have: %s",
				cdrs[0].ExtraFields["PayPalAccount"])
		}
	}
}

// Disable Attributes process
func testV2CDRsProcessCDR2(t *testing.T) {
	args := &engine.ArgV1ProcessEvent{
		Flags: []string{"*attributes:false", utils.MetaRALs},

		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]any{
				utils.OriginID:     "testV2CDRsProcessCDR2",
				utils.OriginHost:   "192.168.1.1",
				utils.Source:       "testV2CDRsProcessCDR2",
				utils.RequestType:  utils.MetaRated,
				utils.Category:     "call",
				utils.AccountField: "testV2CDRsProcessCDR2",
				utils.Subject:      "ANY2CNT",
				utils.Destination:  "+4986517174963",
				utils.AnswerTime:   time.Date(2018, 8, 24, 16, 00, 26, 0, time.UTC),
				utils.Usage:        time.Minute,
				"field_extr1":      "val_extr1",
				"fieldextr2":       "valextr2",
			},
		},
	}

	var reply string
	if err := cdrsRpc.Call(context.Background(), utils.CDRsV1ProcessEvent, args, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
}

func testV2CDRsGetCdrs2(t *testing.T) {
	var cdrCnt int64
	req := utils.AttrGetCdrs{Accounts: []string{"testV2CDRsProcessCDR2"}}
	if err := cdrsRpc.Call(context.Background(), utils.APIerSv2CountCDRs, &req, &cdrCnt); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if cdrCnt != 2 {
		t.Error("Unexpected number of CDRs returned: ", cdrCnt)
	}
	var cdrs []*engine.ExternalCDR
	args := utils.RPCCDRsFilter{RunIDs: []string{"raw"}, OriginIDs: []string{"testV2CDRsProcessCDR2"}}
	if err := cdrsRpc.Call(context.Background(), utils.APIerSv2GetCDRs, &args, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Cost != -1.0 {
			t.Errorf("Unexpected cost for CDR: %f", cdrs[0].Cost)
		}
		//we disable the connection to AttributeS and PayPalAccount shouldn't be present
		if _, has := cdrs[0].ExtraFields["PayPalAccount"]; has {
			t.Errorf("PayPalAccount should NOT be added by AttributeS, have: %s",
				cdrs[0].ExtraFields["PayPalAccount"])
		}
	}
	args = utils.RPCCDRsFilter{RunIDs: []string{"CustomerCharges"}, OriginIDs: []string{"testV2CDRsProcessCDR2"}}
	if err := cdrsRpc.Call(context.Background(), utils.APIerSv2GetCDRs, &args, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Cost != 0.0198 {
			t.Errorf("Unexpected cost for CDR: %f", cdrs[0].Cost)
		}
		//we disable the connection to AttributeS and PayPalAccount shouldn't be present
		if _, has := cdrs[0].ExtraFields["PayPalAccount"]; has {
			t.Errorf("PayPalAccount should NOT be added by AttributeS, have: %s",
				cdrs[0].ExtraFields["PayPalAccount"])
		}
	}
}

// Disable Attributes and Charger process
func testV2CDRsProcessCDR3(t *testing.T) {
	args := &engine.ArgV1ProcessEvent{
		Flags: []string{"*attributes:false", "*chargers:false", "*rals:false"},
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]any{
				utils.OriginID:     "testV2CDRsProcessCDR3",
				utils.OriginHost:   "192.168.1.1",
				utils.Source:       "testV2CDRsProcessCDR3",
				utils.RequestType:  utils.MetaRated,
				utils.Category:     "call",
				utils.AccountField: "testV2CDRsProcessCDR3",
				utils.Subject:      "ANY2CNT",
				utils.Destination:  "+4986517174963",
				utils.AnswerTime:   time.Date(2018, 8, 24, 16, 00, 26, 0, time.UTC),
				utils.Usage:        time.Minute,
				"field_extr1":      "val_extr1",
				"fieldextr2":       "valextr2",
			},
		},
	}

	var reply string
	if err := cdrsRpc.Call(context.Background(), utils.CDRsV1ProcessEvent, args, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
}

func testV2CDRsGetCdrs3(t *testing.T) {
	var cdrCnt int64
	req := utils.AttrGetCdrs{Accounts: []string{"testV2CDRsProcessCDR3"}}
	if err := cdrsRpc.Call(context.Background(), utils.APIerSv2CountCDRs, &req, &cdrCnt); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if cdrCnt != 1 {
		t.Error("Unexpected number of CDRs returned: ", cdrCnt)
	}
	var cdrs []*engine.ExternalCDR
	args := utils.RPCCDRsFilter{RunIDs: []string{utils.MetaDefault}, OriginIDs: []string{"testV2CDRsProcessCDR3"}}
	if err := cdrsRpc.Call(context.Background(), utils.APIerSv2GetCDRs, &args, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Cost != -1.0 {
			t.Errorf("Unexpected cost for CDR: %f", cdrs[0].Cost)
		}
		//we disable the connection to AttributeS and PayPalAccount shouldn't be present
		if _, has := cdrs[0].ExtraFields["PayPalAccount"]; has {
			t.Errorf("PayPalAccount should NOT be added by AttributeS, have: %s",
				cdrs[0].ExtraFields["PayPalAccount"])
		}
	}
	args = utils.RPCCDRsFilter{RunIDs: []string{"CustomerCharges"}, OriginIDs: []string{"testV2CDRsProcessCDR3"}}
	if err := cdrsRpc.Call(context.Background(), utils.APIerSv2GetCDRs, &args, &cdrs); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error("Unexpected error: ", err)
	}
}

// Enable Attributes process
func testV2CDRsProcessCDR4(t *testing.T) {
	args := &engine.ArgV1ProcessEvent{
		Flags: []string{utils.MetaAttributes, utils.MetaRALs},
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]any{
				utils.OriginID:     "testV2CDRsProcessCDR4",
				utils.OriginHost:   "192.168.1.1",
				utils.Source:       "testV2CDRsProcessCDR4",
				utils.RequestType:  utils.MetaRated,
				utils.Category:     "call",
				utils.AccountField: "testV2CDRsProcessCDR4",
				utils.Subject:      "ANY2CNT",
				utils.Destination:  "+4986517174963",
				utils.AnswerTime:   time.Date(2018, 8, 24, 16, 00, 26, 0, time.UTC),
				utils.Usage:        time.Minute,
				"field_extr1":      "val_extr1",
				"fieldextr2":       "valextr2",
			},
		},
	}

	var reply string
	if err := cdrsRpc.Call(context.Background(), utils.CDRsV1ProcessEvent, args, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
}

func testV2CDRsGetCdrs4(t *testing.T) {
	var cdrCnt int64
	req := utils.AttrGetCdrs{Accounts: []string{"testV2CDRsProcessCDR4"}}
	if err := cdrsRpc.Call(context.Background(), utils.APIerSv2CountCDRs, &req, &cdrCnt); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if cdrCnt != 2 {
		t.Error("Unexpected number of CDRs returned: ", cdrCnt)
	}
	var cdrs []*engine.ExternalCDR
	args := utils.RPCCDRsFilter{
		RunIDs:    []string{"raw"},
		OriginIDs: []string{"testV2CDRsProcessCDR4"},
	}
	if err := cdrsRpc.Call(context.Background(), utils.APIerSv2GetCDRs, &args, &cdrs); err != nil {
		t.Fatal("Unexpected error: ", err.Error())
	}
	if len(cdrs) != 1 {
		t.Fatal("Unexpected number of CDRs returned: ", len(cdrs))
	}
	if cdrs[0].Cost != -1.0 {
		t.Errorf("Unexpected cost for CDR: %f", cdrs[0].Cost)
	}
	if rply, has := cdrs[0].ExtraFields["PayPalAccount"]; !has || rply != "paypal@cgrates.org" {
		t.Errorf("PayPalAccount should be added by AttributeS as: paypal@cgrates.org, have: %s",
			cdrs[0].ExtraFields["PayPalAccount"])
	}
	args = utils.RPCCDRsFilter{
		RunIDs:    []string{"CustomerCharges"},
		OriginIDs: []string{"testV2CDRsProcessCDR4"},
	}
	if err := cdrsRpc.Call(context.Background(), utils.APIerSv2GetCDRs, &args, &cdrs); err != nil {
		t.Fatal("Unexpected error: ", err.Error())
	}
	if len(cdrs) != 1 {
		t.Fatal("Unexpected number of CDRs returned: ", len(cdrs))
	}
	if cdrs[0].Cost != 0.0198 {
		t.Errorf("Unexpected cost for CDR: %f", cdrs[0].Cost)
	}
	if rply, has := cdrs[0].ExtraFields["PayPalAccount"]; !has || rply != "paypal@cgrates.org" {
		t.Errorf("PayPalAccount should be added by AttributeS as: paypal@cgrates.org, have: %s",
			cdrs[0].ExtraFields["PayPalAccount"])
	}
}

func testV2CDRsGetCdrs5(t *testing.T) {
	var cdrCnt int64
	req := utils.RPCCDRsFilter{Accounts: []string{"testV2CDRsProcessCDR5"}}
	if err := cdrsRpc.Call(context.Background(), utils.APIerSv2CountCDRs, &req, &cdrCnt); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if cdrCnt != 0 {
		t.Error("Unexpected number of CDRs returned: ", cdrCnt)
	}
	var cdrs []*engine.ExternalCDR
	args := utils.RPCCDRsFilter{
		RunIDs:    []string{"raw"},
		OriginIDs: []string{"testV2CDRsProcessCDR5"},
	}
	if err := cdrsRpc.Call(context.Background(), utils.APIerSv2GetCDRs, &args, &cdrs); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Fatal("Unexpected error: ", err)
	}
	args = utils.RPCCDRsFilter{
		RunIDs:    []string{"CustomerCharges"},
		OriginIDs: []string{"testV2CDRsProcessCDR5"},
	}
	if err := cdrsRpc.Call(context.Background(), utils.APIerSv2GetCDRs, &args, &cdrs); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Fatal("Unexpected error: ", err.Error())
	}
}

func testV2CDRsSetStats(t *testing.T) {
	var reply *engine.StatQueueProfile
	if err := cdrsRpc.Call(context.Background(), utils.APIerSv1GetStatQueueProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "STS_PoccessCDR"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	statConfig := &engine.StatQueueProfileWithAPIOpts{
		StatQueueProfile: &engine.StatQueueProfile{
			Tenant:    "cgrates.org",
			ID:        "STS_PoccessCDR",
			FilterIDs: []string{"*string:~*req.OriginID:testV2CDRsProcessCDR5"},
			// QueueLength: 10,
			Metrics: []*engine.MetricWithFilters{{
				MetricID: "*sum#~*req.Usage",
			}},
			ThresholdIDs: []string{utils.MetaNone},
			Blocker:      true,
			Stored:       true,
			Weight:       20,
			MinItems:     1,
		},
	}
	var result string
	if err := cdrsRpc.Call(context.Background(), utils.APIerSv1SetStatQueueProfile, statConfig, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := cdrsRpc.Call(context.Background(), utils.APIerSv1GetStatQueueProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "STS_PoccessCDR"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(statConfig.StatQueueProfile, reply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(statConfig.StatQueueProfile), utils.ToJSON(reply))
	}
}

func testV2CDRsSetThresholdProfile(t *testing.T) {
	var actreply string

	// Set Action
	attrsAA := &utils.AttrSetActions{ActionsId: "ACT_THD_PoccessCDR", Actions: []*utils.TPAction{{Identifier: utils.MetaLog}}}
	if err := cdrsRpc.Call(context.Background(), utils.APIerSv2SetActions, attrsAA, &actreply); err != nil && err.Error() != utils.ErrExists.Error() {
		t.Error("Got error on APIerSv2.SetActions: ", err.Error())
	} else if actreply != utils.OK {
		t.Errorf("Calling APIerSv2.SetActions received: %s", actreply)
	}

	// Set Account
	attrsSetAccount := &utils.AttrSetAccount{Tenant: "cgrates.org", Account: "testV2CDRsProcessCDR5"}
	if err := cdrsRpc.Call(context.Background(), utils.APIerSv1SetAccount, attrsSetAccount, &actreply); err != nil {
		t.Error("Got error on APIerSv1.SetAccount: ", err.Error())
	} else if actreply != utils.OK {
		t.Errorf("Calling APIerSv1.SetAccount received: %s", actreply)
	}

	// Set Threshold
	var reply *engine.ThresholdProfile
	var result string
	if err := cdrsRpc.Call(context.Background(), utils.APIerSv1GetThresholdProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "THD_PoccessCDR"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	tPrfl := &engine.ThresholdProfileWithAPIOpts{
		ThresholdProfile: &engine.ThresholdProfile{
			Tenant:    "cgrates.org",
			ID:        "THD_PoccessCDR",
			FilterIDs: []string{"*string:~*req.OriginID:testV2CDRsProcessCDR5"},
			MaxHits:   -1,
			Blocker:   false,
			Weight:    20.0,
			ActionIDs: []string{"ACT_THD_PoccessCDR"},
			Async:     false,
		},
	}
	if err := cdrsRpc.Call(context.Background(), utils.APIerSv1SetThresholdProfile, tPrfl, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := cdrsRpc.Call(context.Background(), utils.APIerSv1GetThresholdProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "THD_PoccessCDR"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tPrfl.ThresholdProfile, reply) {
		t.Errorf("Expecting: %+v, received: %+v", tPrfl.ThresholdProfile, reply)
	}
}

func testV2CDRsProcessCDR5(t *testing.T) {
	args := &engine.ArgV1ProcessEvent{
		Flags: []string{"*store:false", "*stats:false", "*thresholds:false"},
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]any{
				utils.OriginID:     "testV2CDRsProcessCDR5",
				utils.OriginHost:   "192.168.1.1",
				utils.Source:       "testV2CDRsProcessCDR5",
				utils.RequestType:  utils.MetaRated,
				utils.Category:     "call",
				utils.AccountField: "testV2CDRsProcessCDR5",
				utils.Subject:      "ANY2CNT2",
				utils.Destination:  "+4986517174963",
				utils.AnswerTime:   time.Date(2018, 8, 24, 16, 00, 26, 0, time.UTC),
				utils.Usage:        time.Minute,
				"field_extr1":      "val_extr1",
				"fieldextr2":       "valextr2",
			},
		},
	}

	var reply string
	if err := cdrsRpc.Call(context.Background(), utils.CDRsV1ProcessEvent, args, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
}

func testV2CDRsGetStats1(t *testing.T) {
	expectedIDs := []string{"STS_PoccessCDR"}
	var metrics map[string]string
	expectedMetrics := map[string]string{
		utils.MetaSum + utils.HashtagSep + utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Usage: utils.NotAvailable,
	}
	if err := cdrsRpc.Call(context.Background(), utils.StatSv1GetQueueStringMetrics,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: expectedIDs[0]},
		}, &metrics); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedMetrics, metrics) {
		t.Errorf("expecting: %+v, received reply: %s", expectedMetrics, metrics)
	}
}

func testV2CDRsGetThreshold1(t *testing.T) {
	expected := []string{"THD_ACNT_1001", "THD_PoccessCDR"}
	var result []string
	if err := cdrsRpc.Call(context.Background(), utils.APIerSv1GetThresholdProfileIDs,
		&utils.PaginatorWithTenant{Tenant: "cgrates.org"}, &result); err != nil {
		t.Error(err)
	} else if len(expected) != len(result) {
		t.Errorf("Expecting : %+v, received: %+v", expected, result)
	}
	var td engine.Threshold
	if err := cdrsRpc.Call(context.Background(), utils.ThresholdSv1GetThreshold,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "THD_PoccessCDR"},
		}, &td); err != nil {
		t.Error(err)
	} else if td.Hits != 0 {
		t.Errorf("received: %+v", td)
	}
}

func testV2CDRsProcessCDR6(t *testing.T) {
	args := &engine.ArgV1ProcessEvent{
		Flags: []string{"*store:false", "*stats:true", "*thresholds:true"},
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]any{
				utils.OriginID:     "testV2CDRsProcessCDR5",
				utils.OriginHost:   "192.168.1.2",
				utils.Source:       "testV2CDRsProcessCDR6",
				utils.RequestType:  utils.MetaRated,
				utils.Category:     "call",
				utils.AccountField: "testV2CDRsProcessCDR6",
				utils.Subject:      "ANY2CNT",
				utils.Destination:  "+4986517174963",
				utils.AnswerTime:   time.Date(2018, 8, 24, 16, 00, 26, 0, time.UTC),
				utils.Usage:        time.Minute,
				"field_extr1":      "val_extr1",
				"fieldextr2":       "valextr2",
			},
		},
	}

	var reply string
	if err := cdrsRpc.Call(context.Background(), utils.CDRsV1ProcessEvent, args, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
}

func testV2CDRsGetStats2(t *testing.T) {
	expectedIDs := []string{"STS_PoccessCDR"}
	var metrics map[string]string
	expectedMetrics := map[string]string{
		utils.MetaSum + utils.HashtagSep + utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Usage: "120000000000",
	}
	if err := cdrsRpc.Call(context.Background(), utils.StatSv1GetQueueStringMetrics,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: expectedIDs[0]},
		}, &metrics); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedMetrics, metrics) {
		t.Errorf("expecting: %+v, received reply: %s", expectedMetrics, metrics)
	}
}

func testV2CDRsGetThreshold2(t *testing.T) {
	var td engine.Threshold
	if err := cdrsRpc.Call(context.Background(), utils.ThresholdSv1GetThreshold,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "THD_PoccessCDR"},
		}, &td); err != nil {
		t.Error(err)
	} else if td.Hits != 2 { // 2 Chargers
		t.Errorf("received: %+v", td)
	}
}

func testV2CDRsProcessCDR7(t *testing.T) {
	args := &engine.ArgV1ProcessEvent{
		Flags: []string{utils.MetaStore, utils.MetaRALs},

		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]any{
				utils.OriginID:     "testV2CDRsProcessCDR7",
				utils.OriginHost:   "192.168.1.1",
				utils.Source:       "testV2CDRsProcessCDR7",
				utils.RequestType:  utils.MetaRated,
				utils.Category:     "call",
				utils.AccountField: "testV2CDRsProcessCDR7",
				utils.Subject:      "ANY2CNT",
				utils.Destination:  "+4986517174963",
				utils.AnswerTime:   time.Date(2018, 8, 24, 16, 00, 26, 0, time.UTC),
				utils.Usage:        time.Minute,
				"field_extr1":      "val_extr1",
				"fieldextr2":       "valextr2",
			},
		},
	}

	var reply string
	if err := cdrsRpc.Call(context.Background(), utils.CDRsV1ProcessEvent, args, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
}

func testV2CDRsGetCdrs7(t *testing.T) {
	var cdrCnt int64
	req := utils.AttrGetCdrs{Accounts: []string{"testV2CDRsProcessCDR7"}}
	if err := cdrsRpc.Call(context.Background(), utils.APIerSv2CountCDRs, &req, &cdrCnt); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if cdrCnt != 2 {
		t.Error("Unexpected number of CDRs returned: ", cdrCnt)
	}
	var cdrs []*engine.ExternalCDR
	args := utils.RPCCDRsFilter{
		RunIDs:    []string{"raw"},
		OriginIDs: []string{"testV2CDRsProcessCDR7"},
	}
	if err := cdrsRpc.Call(context.Background(), utils.APIerSv2GetCDRs, &args, &cdrs); err != nil {
		t.Fatal("Unexpected error: ", err.Error())
	}
	if len(cdrs) != 1 {
		t.Fatal("Unexpected number of CDRs returned: ", len(cdrs))
	}
	if cdrs[0].Cost != -1 {
		t.Errorf("Unexpected cost for CDR: %f", cdrs[0].Cost)
	}
	if rply, has := cdrs[0].ExtraFields["PayPalAccount"]; !has || rply != "paypal@cgrates.org" {
		t.Errorf("PayPalAccount should be added by AttributeS as: paypal@cgrates.org, have: %s",
			cdrs[0].ExtraFields["PayPalAccount"])
	}
	args = utils.RPCCDRsFilter{
		RunIDs:    []string{"CustomerCharges"},
		OriginIDs: []string{"testV2CDRsProcessCDR7"},
	}
	if err := cdrsRpc.Call(context.Background(), utils.APIerSv2GetCDRs, &args, &cdrs); err != nil {
		t.Fatal("Unexpected error: ", err.Error())
	}
	if len(cdrs) != 1 {
		t.Fatal("Unexpected number of CDRs returned: ", len(cdrs))
	}
	if cdrs[0].Cost != 0.0198 {
		t.Errorf("Unexpected cost for CDR: %f", cdrs[0].Cost)
	}
	if rply, has := cdrs[0].ExtraFields["PayPalAccount"]; !has || rply != "paypal@cgrates.org" {
		t.Errorf("PayPalAccount should be added by AttributeS as: paypal@cgrates.org, have: %s",
			cdrs[0].ExtraFields["PayPalAccount"])
	}
}

func testV2CDRsKillEngine(t *testing.T) {
	if err := engine.KillEngine(*utils.WaitRater); err != nil {
		t.Error(err)
	}
}
