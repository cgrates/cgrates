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
package v2

import (
	"net/rpc"
	"path"
	"reflect"
	"sync"
	"testing"
	"time"

	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	cdrsCfgPath string
	cdrsCfg     *config.CGRConfig
	cdrsRpc     *rpc.Client
	cdrsConfDIR string // run the tests for specific configuration

	// subtests to be executed for each confDIR
	sTestsCDRsIT = []func(t *testing.T){
		testV2CDRsInitConfig,
		testV2CDRsInitDataDb,
		testV2CDRsInitCdrDb,
		testV2CDRsStartEngine,
		testV2CDRsRpcConn,
		testV2CDRsLoadTariffPlanFromFolder,
		testV2CDRsProcessCDR,
		testV2CDRsGetCdrs,
		testV2CDRsRateCDRs,
		testV2CDRsGetCdrs2,
		testV2CDRsUsageNegative,
		testV2CDRsDifferentTenants,

		testV2CDRsRemoveRatingProfiles,
		testV2CDRsProcessCDRNoRattingPlan,
		testV2CDRsGetCdrsNoRattingPlan,

		testV2CDRsRateCDRsWithRatingPlan,
		testV2CDRsGetCdrsWithRatingPlan,

		testV2CDRsSetThreshold,
		testV2CDRsProcessCDRWithThreshold,
		testV2CDRsGetThreshold,
		testV2CDRsResetThresholdAction,

		testv2CDRsGetCDRsDest,

		testV2CDRsInitDataDb,
		testV2CDRsInitCdrDb,
		testV2CDRsRerate,

		testV2CDRsLoadTariffPlanFromFolder,
		testv2CDRsDynaPrepaid,

		//testV2CDRsDuplicateCDRs,

		testV2CDRsKillEngine,
	}
)

// Tests starting here
func TestCDRsIT(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		cdrsConfDIR = "cdrsv2internal"
	case utils.MetaMySQL:
		cdrsConfDIR = "cdrsv2mysql"
	case utils.MetaMongo:
		cdrsConfDIR = "cdrsv2mongo"
	case utils.MetaPostgres:
		cdrsConfDIR = "cdrsv2psql"
	default:
		t.Fatal("Unknown Database type")
	}

	for _, stest := range sTestsCDRsIT {
		t.Run(cdrsConfDIR, stest)
	}
}

func testV2CDRsInitConfig(t *testing.T) {
	var err error
	cdrsCfgPath = path.Join(*dataDir, "conf", "samples", cdrsConfDIR)
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
	if _, err := engine.StopStartEngine(cdrsCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testV2CDRsRpcConn(t *testing.T) {
	cdrsRpc, err = newRPCClient(cdrsCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

func testV2CDRsLoadTariffPlanFromFolder(t *testing.T) {
	var loadInst utils.LoadInstance
	if err := cdrsRpc.Call(utils.APIerSv2LoadTariffPlanFromFolder,
		&utils.AttrLoadTpFromFolder{FolderPath: path.Join(
			*dataDir, "tariffplans", "testit")}, &loadInst); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond) // Give time for scheduler to execute topups
}

func testV2CDRsProcessCDR(t *testing.T) {
	args := &engine.ArgV1ProcessEvent{
		Flags: []string{utils.MetaRALs},
		CGREventWithOpts: utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				Event: map[string]interface{}{
					utils.OriginID:    "testV2CDRsProcessCDR1",
					utils.OriginHost:  "192.168.1.1",
					utils.Source:      "testV2CDRsProcessCDR",
					utils.RequestType: utils.META_RATED,
					// utils.Category:    "call", //it will be populated as default in MapEvent.AsCDR
					utils.AccountField: "testV2CDRsProcessCDR",
					utils.Subject:      "ANY2CNT",
					utils.Destination:  "+4986517174963",
					utils.AnswerTime:   time.Date(2018, 8, 24, 16, 00, 26, 0, time.UTC),
					utils.Usage:        time.Minute,
					"field_extr1":      "val_extr1",
					"fieldextr2":       "valextr2",
				},
			},
		},
	}

	var reply string
	if err := cdrsRpc.Call(utils.CDRsV1ProcessEvent, args, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
}

func testV2CDRsGetCdrs(t *testing.T) {
	var cdrCnt int64
	req := utils.AttrGetCdrs{}
	if err := cdrsRpc.Call(utils.APIerSv2CountCDRs, &req, &cdrCnt); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if cdrCnt != 3 {
		t.Error("Unexpected number of CDRs returned: ", cdrCnt)
	}
	var cdrs []*engine.ExternalCDR
	args := utils.RPCCDRsFilter{RunIDs: []string{"raw"}}
	if err := cdrsRpc.Call(utils.APIerSv2GetCDRs, &args, &cdrs); err != nil {
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
	if err := cdrsRpc.Call(utils.APIerSv2GetCDRs, &args, &cdrs); err != nil {
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
	args = utils.RPCCDRsFilter{RunIDs: []string{"SupplierCharges"}}
	if err := cdrsRpc.Call(utils.APIerSv2GetCDRs, &args, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Cost != 0.0102 {
			t.Errorf("Unexpected cost for CDR: %f", cdrs[0].Cost)
		}
		if cdrs[0].ExtraFields["PayPalAccount"] != "paypal@cgrates.org" {
			t.Errorf("PayPalAccount should be added by AttributeS, have: %s",
				cdrs[0].ExtraFields["PayPalAccount"])
		}
	}
}

// Should re-rate the supplier1 cost with RP_ANY2CNT
func testV2CDRsRateCDRs(t *testing.T) {
	var rpl engine.RatingProfile
	attrGetRatingPlan := &utils.AttrGetRatingProfile{
		Tenant: "cgrates.org", Category: "call", Subject: "SUPPLIER1"}
	actTime, err := utils.ParseTimeDetectLayout("2018-01-01T00:00:00Z", "")
	if err != nil {
		t.Error(err)
	}
	expected := engine.RatingProfile{
		Id: "*out:cgrates.org:call:SUPPLIER1",
		RatingPlanActivations: engine.RatingPlanActivations{
			{
				ActivationTime: actTime,
				RatingPlanId:   "RP_ANY1CNT",
			},
		},
	}
	if err := cdrsRpc.Call(utils.APIerSv1GetRatingProfile, attrGetRatingPlan, &rpl); err != nil {
		t.Errorf("Got error on APIerSv1.GetRatingProfile: %+v", err)
	} else if !reflect.DeepEqual(expected, rpl) {
		t.Errorf("Calling APIerSv1.GetRatingProfile expected: %+v, received: %+v", utils.ToJSON(expected), utils.ToJSON(rpl))
	}

	rpf := &utils.AttrSetRatingProfile{
		Tenant:   "cgrates.org",
		Category: "call",
		Subject:  "SUPPLIER1",
		RatingPlanActivations: []*utils.TPRatingActivation{
			{
				ActivationTime: "2018-01-01T00:00:00Z",
				RatingPlanId:   "RP_ANY2CNT"}},
		Overwrite: true,
	}
	var reply string
	if err := cdrsRpc.Call(utils.APIerSv1SetRatingProfile, &rpf, &reply); err != nil {
		t.Error("Got error on APIerSv1.SetRatingProfile: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Calling APIerSv1.SetRatingProfile got reply: ", reply)
	}

	expected = engine.RatingProfile{
		Id: "*out:cgrates.org:call:SUPPLIER1",
		RatingPlanActivations: engine.RatingPlanActivations{
			{
				ActivationTime: actTime,
				RatingPlanId:   "RP_ANY2CNT",
			},
		},
	}
	if err := cdrsRpc.Call(utils.APIerSv1GetRatingProfile, attrGetRatingPlan, &rpl); err != nil {
		t.Errorf("Got error on APIerSv1.GetRatingProfile: %+v", err)
	} else if !reflect.DeepEqual(expected, rpl) {
		t.Errorf("Calling APIerSv1.GetRatingProfile expected: %+v, received: %+v", utils.ToJSON(expected), utils.ToJSON(rpl))
	}

	if err := cdrsRpc.Call(utils.CDRsV1RateCDRs, &engine.ArgRateCDRs{
		RPCCDRsFilter: utils.RPCCDRsFilter{NotRunIDs: []string{"raw"}},
		Flags:         []string{"*chargers:false"},
	}, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
}

func testV2CDRsGetCdrs2(t *testing.T) {
	var cdrCnt int64
	req := utils.AttrGetCdrs{}
	if err := cdrsRpc.Call(utils.APIerSv2CountCDRs, &req, &cdrCnt); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if cdrCnt != 3 {
		t.Error("Unexpected number of CDRs returned: ", cdrCnt)
	}
	var cdrs []*engine.ExternalCDR
	args := utils.RPCCDRsFilter{RunIDs: []string{"raw"}}
	if err := cdrsRpc.Call(utils.APIerSv2GetCDRs, &args, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Cost != -1.0 {
			t.Errorf("Unexpected cost for CDR: %f", cdrs[0].Cost)
		}
	}
	args = utils.RPCCDRsFilter{RunIDs: []string{"CustomerCharges"}}
	if err := cdrsRpc.Call(utils.APIerSv2GetCDRs, &args, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Cost != 0.0198 {
			t.Errorf("Unexpected cost for CDR: %f", cdrs[0].Cost)
		}
	}
	args = utils.RPCCDRsFilter{RunIDs: []string{"SupplierCharges"}}
	if err := cdrsRpc.Call(utils.APIerSv2GetCDRs, &args, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Cost != 0.0102 {
			t.Errorf("Unexpected cost for CDR: %f", cdrs[0].Cost)
		}
	}
}

func testV2CDRsUsageNegative(t *testing.T) {
	argsCdr := &engine.ArgV1ProcessEvent{
		Flags: []string{utils.MetaRALs},
		CGREventWithOpts: utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				Event: map[string]interface{}{
					utils.OriginID:     "testV2CDRsUsageNegative",
					utils.OriginHost:   "192.168.1.1",
					utils.Source:       "testV2CDRsUsageNegative",
					utils.RequestType:  utils.META_RATED,
					utils.Category:     "call",
					utils.AccountField: "testV2CDRsUsageNegative",
					utils.Subject:      "ANY2CNT",
					utils.Destination:  "+4986517174963",
					utils.AnswerTime:   time.Date(2018, 8, 24, 16, 00, 26, 0, time.UTC),
					utils.Usage:        -time.Minute,
					"field_extr1":      "val_extr1",
					"fieldextr2":       "valextr2",
				},
			},
		},
	}
	var reply string
	if err := cdrsRpc.Call(utils.CDRsV1ProcessEvent, argsCdr, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}

	var cdrs []*engine.ExternalCDR
	args := utils.RPCCDRsFilter{RunIDs: []string{"raw"}, OriginIDs: []string{"testV2CDRsUsageNegative"}}
	if err := cdrsRpc.Call(utils.APIerSv2GetCDRs, &args, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Cost != -1.0 {
			t.Errorf("Unexpected cost for CDR: %f", cdrs[0].Cost)
		}
		if cdrs[0].Usage != "-1m0s" {
			t.Errorf("Unexpected usage for CDR: %s", cdrs[0].Usage)
		}
	}
	cdrs = nil // gob doesn't modify zero-value fields
	args = utils.RPCCDRsFilter{RunIDs: []string{"CustomerCharges"}, OriginIDs: []string{"testV2CDRsUsageNegative"}}
	if err := cdrsRpc.Call(utils.APIerSv2GetCDRs, &args, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Cost != 0 {
			t.Errorf("Unexpected cost for CDR: %f", cdrs[0].Cost)
		}
		if cdrs[0].Usage != "0s" {
			t.Errorf("Unexpected usage for CDR: %s", cdrs[0].Usage)
		}
	}
	cdrs = nil // gob doesn't modify zero-value fields
	args = utils.RPCCDRsFilter{RunIDs: []string{"SupplierCharges"}, OriginIDs: []string{"testV2CDRsUsageNegative"}}
	if err := cdrsRpc.Call(utils.APIerSv2GetCDRs, &args, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Cost != 0 {
			t.Errorf("Unexpected cost for CDR: %f", cdrs[0].Cost)
		}
		if cdrs[0].Usage != "0s" {
			t.Errorf("Unexpected usage for CDR: %s", cdrs[0].Usage)
		}
	}
}

func testV2CDRsDifferentTenants(t *testing.T) {
	//add an attribute
	alsPrf := &v1.AttributeWithCache{
		AttributeProfile: &engine.AttributeProfile{
			Tenant:    "cgrates.com",
			ID:        "ATTR_Tenant",
			Contexts:  []string{utils.META_ANY},
			FilterIDs: []string{"*string:~*req.Tenant:cgrates.com"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			Attributes: []*engine.Attribute{
				{
					Path: utils.MetaTenant,
					Type: utils.META_CONSTANT,
					Value: config.RSRParsers{
						&config.RSRParser{
							Rules: "CustomTenant",
						},
					},
				},
				{
					Path: utils.MetaReq + utils.NestingSep + utils.Tenant,
					Type: utils.META_CONSTANT,
					Value: config.RSRParsers{
						&config.RSRParser{
							Rules: "CustomTenant",
						},
					},
				},
			},
			Blocker: false,
			Weight:  10,
		},
		Cache: utils.StringPointer(utils.MetaReload),
	}
	alsPrf.Compile()
	var result string
	if err := cdrsRpc.Call(utils.APIerSv1SetAttributeProfile, alsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var reply *engine.AttributeProfile
	if err := cdrsRpc.Call(utils.APIerSv1GetAttributeProfile,
		&utils.TenantIDWithOpts{TenantID: &utils.TenantID{Tenant: "cgrates.com", ID: "ATTR_Tenant"}}, &reply); err != nil {
		t.Fatal(err)
	}
	reply.Compile()
	if !reflect.DeepEqual(alsPrf.AttributeProfile, reply) {
		t.Errorf("Expecting : %+v, received: %+v", alsPrf.AttributeProfile, reply)
	}
	//add a charger
	chargerProfile := &v1.ChargerWithCache{
		ChargerProfile: &engine.ChargerProfile{
			Tenant: "CustomTenant",
			ID:     "CustomCharger",
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			RunID:        "CustomRunID",
			AttributeIDs: []string{"*none"},
			Weight:       20,
		},
		Cache: utils.StringPointer(utils.MetaReload),
	}
	if err := cdrsRpc.Call(utils.APIerSv1SetChargerProfile, chargerProfile, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var reply2 *engine.ChargerProfile
	if err := cdrsRpc.Call(utils.APIerSv1GetChargerProfile,
		&utils.TenantID{Tenant: "CustomTenant", ID: "CustomCharger"}, &reply2); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(chargerProfile.ChargerProfile, reply2) {
		t.Errorf("Expecting : %+v, received: %+v", chargerProfile.ChargerProfile, reply2)
	}

	argsCdr := &engine.ArgV1ProcessEvent{
		Flags: []string{utils.MetaAttributes, utils.MetaChargers, "*stats:false", "*thresholds:false", utils.MetaStore},
		CGREventWithOpts: utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.com",
				Event: map[string]interface{}{
					utils.OriginID:     "testV2CDRsDifferentTenants",
					utils.OriginHost:   "192.168.1.1",
					utils.Source:       "testV2CDRsDifferentTenants",
					utils.RequestType:  utils.META_RATED,
					utils.Category:     "call",
					utils.AccountField: "testV2CDRsDifferentTenants",
					utils.Destination:  "+4986517174963",
					utils.Tenant:       "cgrates.com",
					utils.AnswerTime:   time.Date(2018, 8, 24, 16, 00, 26, 0, time.UTC),
					utils.Usage:        time.Second,
					"field_extr1":      "val_extr1",
					"fieldextr2":       "valextr2",
				},
			},
		},
	}
	var reply3 string
	if err := cdrsRpc.Call(utils.CDRsV1ProcessEvent, argsCdr, &reply3); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if reply3 != utils.OK {
		t.Error("Unexpected reply received: ", reply3)
	}

	var cdrs []*engine.ExternalCDR
	args := utils.RPCCDRsFilter{Tenants: []string{"CustomTenant"}}
	if err := cdrsRpc.Call(utils.APIerSv2GetCDRs, &args, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 { // no raw Charger defined
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	}
}

func testV2CDRsRemoveRatingProfiles(t *testing.T) {
	var reply string
	if err := cdrsRpc.Call(utils.APIerSv1RemoveRatingProfile, &v1.AttrRemoveRatingProfile{
		Tenant:   "cgrates.org",
		Category: utils.CALL,
		Subject:  utils.ANY,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected: %s, received: %s ", utils.OK, reply)
	}
	if err := cdrsRpc.Call(utils.APIerSv1RemoveRatingProfile, &v1.AttrRemoveRatingProfile{
		Tenant:   "cgrates.org",
		Category: utils.CALL,
		Subject:  "SUPPLIER1",
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected: %s, received: %s ", utils.OK, reply)
	}
}

func testV2CDRsProcessCDRNoRattingPlan(t *testing.T) {
	args := &engine.ArgV1ProcessEvent{
		Flags: []string{utils.MetaRALs},
		CGREventWithOpts: utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				Event: map[string]interface{}{
					utils.OriginID:     "testV2CDRsProcessCDR4",
					utils.OriginHost:   "192.168.1.1",
					utils.Source:       "testV2CDRsProcessCDR4",
					utils.RequestType:  utils.META_RATED,
					utils.AccountField: "testV2CDRsProcessCDR4",
					utils.Subject:      "NoSubject",
					utils.Destination:  "+1234567",
					utils.AnswerTime:   time.Date(2018, 8, 24, 16, 00, 26, 0, time.UTC),
					utils.Usage:        time.Minute,
					"field_extr1":      "val_extr1",
					"fieldextr2":       "valextr2",
				},
			},
		},
	}

	var reply string
	if err := cdrsRpc.Call(utils.CDRsV1ProcessEvent, args, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
}

func testV2CDRsGetCdrsNoRattingPlan(t *testing.T) {
	var cdrCnt int64
	req := utils.AttrGetCdrs{}
	if err := cdrsRpc.Call(utils.APIerSv2CountCDRs, &req, &cdrCnt); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if cdrCnt != 10 {
		t.Error("Unexpected number of CDRs returned: ", cdrCnt)
	}
	var cdrs []*engine.ExternalCDR
	args := utils.RPCCDRsFilter{RunIDs: []string{"raw"}, Accounts: []string{"testV2CDRsProcessCDR4"}}
	if err := cdrsRpc.Call(utils.APIerSv2GetCDRs, &args, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Cost != -1.0 {
			t.Errorf("Unexpected cost for CDR: %f", cdrs[0].Cost)
		}
	}
	args = utils.RPCCDRsFilter{RunIDs: []string{"CustomerCharges"}, Accounts: []string{"testV2CDRsProcessCDR4"}}
	if err := cdrsRpc.Call(utils.APIerSv2GetCDRs, &args, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Cost != -1 {
			t.Errorf("Unexpected cost for CDR: %f", cdrs[0].Cost)
		}
		if cdrs[0].ExtraInfo != utils.ErrRatingPlanNotFound.Error() {
			t.Errorf("Expected ExtraInfo : %s received :%s", utils.ErrRatingPlanNotFound.Error(), cdrs[0].ExtraInfo)
		}
	}
	args = utils.RPCCDRsFilter{RunIDs: []string{"SupplierCharges"}, Accounts: []string{"testV2CDRsProcessCDR4"}}
	if err := cdrsRpc.Call(utils.APIerSv2GetCDRs, &args, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Cost != -1 {
			t.Errorf("Unexpected cost for CDR: %f", cdrs[0].Cost)
		}
		if cdrs[0].ExtraInfo != utils.ErrRatingPlanNotFound.Error() {
			t.Errorf("Expected ExtraInfo : %s received :%s", utils.ErrRatingPlanNotFound.Error(), cdrs[0].ExtraInfo)
		}
	}
}

// Should re-rate the supplier1 cost with RP_ANY2CNT
func testV2CDRsRateCDRsWithRatingPlan(t *testing.T) {
	rpf := &utils.AttrSetRatingProfile{
		Tenant:   "cgrates.org",
		Category: "call",
		Subject:  "SUPPLIER1",
		RatingPlanActivations: []*utils.TPRatingActivation{
			{
				ActivationTime: "2018-01-01T00:00:00Z",
				RatingPlanId:   "RP_ANY1CNT"}},
		Overwrite: true,
	}
	var reply string
	if err := cdrsRpc.Call(utils.APIerSv1SetRatingProfile, &rpf, &reply); err != nil {
		t.Error("Got error on APIerSv1.SetRatingProfile: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Calling APIerSv1.SetRatingProfile got reply: ", reply)
	}

	rpf = &utils.AttrSetRatingProfile{
		Tenant:   "cgrates.org",
		Category: "call",
		Subject:  utils.ANY,
		RatingPlanActivations: []*utils.TPRatingActivation{
			{
				ActivationTime: "2018-01-01T00:00:00Z",
				RatingPlanId:   "RP_TESTIT1"}},
		Overwrite: true,
	}
	if err := cdrsRpc.Call(utils.APIerSv1SetRatingProfile, &rpf, &reply); err != nil {
		t.Error("Got error on APIerSv1.SetRatingProfile: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Calling APIerSv1.SetRatingProfile got reply: ", reply)
	}

	if err := cdrsRpc.Call(utils.CDRsV1RateCDRs, &engine.ArgRateCDRs{
		RPCCDRsFilter: utils.RPCCDRsFilter{NotRunIDs: []string{"raw"}, Accounts: []string{"testV2CDRsProcessCDR4"}},
		Flags:         []string{"*chargers:true"},
	}, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
}

func testV2CDRsGetCdrsWithRatingPlan(t *testing.T) {
	var cdrCnt int64
	req := utils.AttrGetCdrs{}
	if err := cdrsRpc.Call(utils.APIerSv2CountCDRs, &req, &cdrCnt); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if cdrCnt != 10 {
		t.Error("Unexpected number of CDRs returned: ", cdrCnt)
	}
	var cdrs []*engine.ExternalCDR
	args := utils.RPCCDRsFilter{RunIDs: []string{"raw"}, Accounts: []string{"testV2CDRsProcessCDR4"}}
	if err := cdrsRpc.Call(utils.APIerSv2GetCDRs, &args, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Cost != -1.0 {
			t.Errorf("Unexpected cost for CDR: %f", cdrs[0].Cost)
		}
	}
	cdrs = []*engine.ExternalCDR{} // gob will not update zero value fields
	args = utils.RPCCDRsFilter{RunIDs: []string{"CustomerCharges"}, Accounts: []string{"testV2CDRsProcessCDR4"}}
	if err := cdrsRpc.Call(utils.APIerSv2GetCDRs, &args, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Cost != 0.0102 {
			t.Errorf("Unexpected cost for CDR: %f", cdrs[0].Cost)
		}
		if cdrs[0].ExtraInfo != "" {
			t.Errorf("Expected ExtraInfo : %s received :%s", "", cdrs[0].ExtraInfo)
		}
	}
	cdrs = []*engine.ExternalCDR{} // gob will not update zero value fields
	args = utils.RPCCDRsFilter{RunIDs: []string{"SupplierCharges"}, Accounts: []string{"testV2CDRsProcessCDR4"}}
	if err := cdrsRpc.Call(utils.APIerSv2GetCDRs, &args, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Cost != 0.0102 {
			t.Errorf("Unexpected cost for CDR: %f", cdrs[0].Cost)
		}
		if cdrs[0].ExtraInfo != "" {
			t.Errorf("Expected ExtraInfo : %s received :%s", "", cdrs[0].ExtraInfo)
		}
	}
}

func testV2CDRsSetThreshold(t *testing.T) {
	var reply string
	if err := cdrsRpc.Call(utils.APIerSv2SetActions, &utils.AttrSetActions{
		ActionsId: "ACT_LOG",
		Actions:   []*utils.TPAction{{Identifier: utils.LOG}},
	}, &reply); err != nil && err.Error() != utils.ErrExists.Error() {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv2.SetActions received: %s", reply)
	}
	tPrfl := engine.ThresholdWithCache{
		ThresholdProfile: &engine.ThresholdProfile{
			Tenant: "cgrates.org",
			ID:     "THD_Test",
			FilterIDs: []string{
				"*lt:~*req.CostDetails.AccountSummary.BalanceSummaries[0].Value:10",
				"*string:~*req.Account:1005", // only for indexes
			},
			MaxHits:   -1,
			Weight:    30,
			ActionIDs: []string{"ACT_LOG"},
		},
	}
	if err := cdrsRpc.Call(utils.APIerSv1SetThresholdProfile, tPrfl, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	attrSetAcnt := AttrSetAccount{
		Tenant:  "cgrates.org",
		Account: "1005",
		ExtraOptions: map[string]bool{
			utils.AllowNegative: true,
		},
	}
	if err := cdrsRpc.Call(utils.APIerSv2SetAccount, &attrSetAcnt, &reply); err != nil {
		t.Fatal(err)
	}
	attrs := &utils.AttrSetBalance{
		Tenant:      "cgrates.org",
		Account:     "1005",
		BalanceType: utils.MONETARY,
		Value:       1,
		Balance: map[string]interface{}{
			utils.ID:     utils.MetaDefault,
			utils.Weight: 10.0,
		},
	}
	if err := cdrsRpc.Call(utils.APIerSv2SetBalance, attrs, &reply); err != nil {
		t.Fatal(err)
	}
}

func testV2CDRsProcessCDRWithThreshold(t *testing.T) {
	args := &engine.ArgV1ProcessEvent{
		Flags: []string{utils.MetaThresholds, utils.MetaRALs, utils.ConcatenatedKey(utils.MetaChargers, "false")},
		CGREventWithOpts: utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				Event: map[string]interface{}{
					utils.OriginID:     "testV2CDRsProcessCDRWithThreshold",
					utils.OriginHost:   "192.168.1.1",
					utils.Source:       "testV2CDRsProcessCDRWithThreshold",
					utils.RequestType:  utils.META_POSTPAID,
					utils.Category:     "call",
					utils.AccountField: "1005",
					utils.Subject:      "ANY2CNT",
					utils.Destination:  "+4986517174963",
					utils.AnswerTime:   time.Date(2018, 8, 24, 16, 00, 26, 0, time.UTC),
					utils.Usage:        100 * time.Minute,
					"field_extr1":      "val_extr1",
					"fieldextr2":       "valextr2",
				},
			},
		},
	}
	var reply string
	if err := cdrsRpc.Call(utils.CDRsV1ProcessEvent, args, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
}

func testV2CDRsGetThreshold(t *testing.T) {
	var td engine.Threshold
	if err := cdrsRpc.Call(utils.ThresholdSv1GetThreshold,
		&utils.TenantIDWithOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "THD_Test"}}, &td); err != nil {
		t.Error(err)
	} else if td.Hits != 1 {
		t.Errorf("Expecting threshold to be hit once received: %v", td.Hits)
	}
}

func testv2CDRsGetCDRsDest(t *testing.T) {
	args := &engine.ArgV1ProcessEvent{
		Flags: []string{utils.MetaStore},
		CGREventWithOpts: utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				Event: map[string]interface{}{
					utils.CGRID:        "9b3cd5e698af94f8916220866c831a982ed16318",
					utils.ToR:          utils.VOICE,
					utils.RunID:        "raw",
					utils.OriginID:     "25160047719:0",
					utils.OriginHost:   "192.168.1.1",
					utils.Source:       "*sessions",
					utils.RequestType:  utils.META_NONE,
					utils.Category:     "call",
					utils.AccountField: "1001",
					utils.Subject:      "1001",
					utils.Destination:  "+4915117174963",
					utils.AnswerTime:   time.Date(2018, 8, 24, 16, 00, 26, 0, time.UTC),
					utils.Usage:        100 * time.Minute,
				},
			},
		},
	}
	var reply string
	if err := cdrsRpc.Call(utils.CDRsV1ProcessEvent, args, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}

	var cdrs []*engine.ExternalCDR
	if err := cdrsRpc.Call(utils.APIerSv2GetCDRs, &utils.RPCCDRsFilter{DestinationPrefixes: []string{"+4915117174963"}},
		&cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 3 {
		t.Error("Unexpected number of CDRs returned: ", len(reply))
	}
}

func testV2CDRsRerate(t *testing.T) {
	var reply string
	if err := cdrsRpc.Call(utils.CacheSv1Clear, &utils.AttrCacheIDsWithOpts{
		CacheIDs: nil,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Reply: ", reply)
	}
	//add a charger
	chargerProfile := &v1.ChargerWithCache{
		ChargerProfile: &engine.ChargerProfile{
			Tenant: "cgrates.org",
			ID:     "Default",
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			RunID:        utils.MetaDefault,
			AttributeIDs: []string{"*none"},
			Weight:       20,
		},
		Cache: utils.StringPointer(utils.MetaReload),
	}
	if err := cdrsRpc.Call(utils.APIerSv1SetChargerProfile, chargerProfile, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}

	attrSetAcnt := AttrSetAccount{
		Tenant:  "cgrates.org",
		Account: "voiceAccount",
	}
	if err := cdrsRpc.Call(utils.APIerSv2SetAccount, &attrSetAcnt, &reply); err != nil {
		t.Fatal(err)
	}
	attrs := &utils.AttrSetBalance{
		Tenant:      "cgrates.org",
		Account:     "voiceAccount",
		BalanceType: utils.VOICE,
		Value:       600000000000,
		Balance: map[string]interface{}{
			utils.ID:        utils.MetaDefault,
			"RatingSubject": "*zero1m",
			utils.Weight:    10.0,
		},
	}
	if err := cdrsRpc.Call(utils.APIerSv2SetBalance, attrs, &reply); err != nil {
		t.Fatal(err)
	}

	var acnt *engine.Account
	if err := cdrsRpc.Call(utils.APIerSv2GetAccount,
		&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "voiceAccount"}, &acnt); err != nil {
		t.Error(err)
	} else if len(acnt.BalanceMap) != 1 || acnt.BalanceMap[utils.VOICE][0].Value != 600000000000 {
		t.Errorf("Unexpected balance received: %+v", acnt.BalanceMap[utils.VOICE][0])
	}

	args := &engine.ArgV1ProcessEvent{
		Flags: []string{utils.MetaRerate},
		CGREventWithOpts: utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				Event: map[string]interface{}{
					utils.OriginID:     "testV2CDRsRerate",
					utils.OriginHost:   "192.168.1.1",
					utils.Source:       "testV2CDRsRerate",
					utils.RequestType:  utils.META_PSEUDOPREPAID,
					utils.AccountField: "voiceAccount",
					utils.Subject:      "ANY2CNT",
					utils.Destination:  "+4986517174963",
					utils.AnswerTime:   time.Date(2018, 8, 24, 16, 00, 26, 0, time.UTC),
					utils.Usage:        2 * time.Minute,
					"field_extr1":      "val_extr1",
					"fieldextr2":       "valextr2",
				},
			},
		},
	}

	var rplProcEv []*utils.EventWithFlags
	if err := cdrsRpc.Call(utils.CDRsV2ProcessEvent, args, &rplProcEv); err != nil {
		t.Error("Unexpected error: ", err.Error())
	}

	if err := cdrsRpc.Call(utils.APIerSv2GetAccount,
		&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "voiceAccount"}, &acnt); err != nil {
		t.Error(err)
	} else if len(acnt.BalanceMap) != 1 || acnt.BalanceMap[utils.VOICE][0].Value != 480000000000 {
		t.Errorf("Unexpected balance received: %+v", acnt.BalanceMap[utils.VOICE][0])
	}

	args2 := &engine.ArgV1ProcessEvent{
		Flags: []string{utils.MetaRerate},
		CGREventWithOpts: utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				Event: map[string]interface{}{
					utils.OriginID:     "testV2CDRsRerate",
					utils.OriginHost:   "192.168.1.1",
					utils.Source:       "testV2CDRsRerate",
					utils.RequestType:  utils.META_PSEUDOPREPAID,
					utils.AccountField: "voiceAccount",
					utils.Subject:      "ANY2CNT",
					utils.Destination:  "+4986517174963",
					utils.AnswerTime:   time.Date(2018, 8, 24, 16, 00, 26, 0, time.UTC),
					utils.Usage:        time.Minute,
					"field_extr1":      "val_extr1",
					"fieldextr2":       "valextr2",
				},
			},
		},
	}

	if err := cdrsRpc.Call(utils.CDRsV2ProcessEvent, args2, &rplProcEv); err != nil {
		t.Error("Unexpected error: ", err.Error())
	}

	if err := cdrsRpc.Call(utils.APIerSv2GetAccount,
		&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "voiceAccount"}, &acnt); err != nil {
		t.Error(err)
	} else if len(acnt.BalanceMap) != 1 || acnt.BalanceMap[utils.VOICE][0].Value != 540000000000 {
		t.Errorf("Unexpected balance received: %+v", acnt.BalanceMap[utils.VOICE][0])
	}
}

func testv2CDRsDynaPrepaid(t *testing.T) {
	var acnt engine.Account
	if err := cdrsRpc.Call(utils.APIerSv2GetAccount,
		&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "CreatedAccount"}, &acnt); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}

	args := &engine.ArgV1ProcessEvent{
		Flags: []string{utils.MetaRALs},
		CGREventWithOpts: utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				Event: map[string]interface{}{
					utils.OriginID:     "testv2CDRsDynaPrepaid",
					utils.OriginHost:   "192.168.1.1",
					utils.Source:       "testv2CDRsDynaPrepaid",
					utils.RequestType:  utils.MetaDynaprepaid,
					utils.AccountField: "CreatedAccount",
					utils.Subject:      "NoSubject",
					utils.Destination:  "+1234567",
					utils.AnswerTime:   time.Date(2018, 8, 24, 16, 00, 26, 0, time.UTC),
					utils.Usage:        time.Minute,
				},
			},
		},
	}

	var reply string
	if err := cdrsRpc.Call(utils.CDRsV1ProcessEvent, args, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}

	if err := cdrsRpc.Call(utils.APIerSv2GetAccount,
		&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "CreatedAccount"}, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MONETARY][0].Value != 9.9694 {
		t.Errorf("Unexpected balance received: %+v", acnt.BalanceMap[utils.MONETARY][0])
	}
}

func testV2CDRsDuplicateCDRs(t *testing.T) {
	var reply string
	if err := cdrsRpc.Call(utils.CacheSv1Clear, &utils.AttrCacheIDsWithOpts{
		CacheIDs: nil,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Reply: ", reply)
	}
	//add a charger
	chargerProfile := &v1.ChargerWithCache{
		ChargerProfile: &engine.ChargerProfile{
			Tenant: "cgrates.org",
			ID:     "Default",
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			RunID:        utils.MetaDefault,
			AttributeIDs: []string{"*none"},
			Weight:       20,
		},
		Cache: utils.StringPointer(utils.MetaReload),
	}
	if err := cdrsRpc.Call(utils.APIerSv1SetChargerProfile, chargerProfile, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}

	attrSetAcnt := AttrSetAccount{
		Tenant:  "cgrates.org",
		Account: "testV2CDRsDuplicateCDRs",
	}
	if err := cdrsRpc.Call(utils.APIerSv2SetAccount, &attrSetAcnt, &reply); err != nil {
		t.Fatal(err)
	}
	attrs := &utils.AttrSetBalance{
		Tenant:      "cgrates.org",
		Account:     "testV2CDRsDuplicateCDRs",
		BalanceType: utils.VOICE,
		Value:       600000000000,
		Balance: map[string]interface{}{
			utils.ID:        utils.MetaDefault,
			"RatingSubject": "*zero1m",
			utils.Weight:    10.0,
		},
	}
	if err := cdrsRpc.Call(utils.APIerSv2SetBalance, attrs, &reply); err != nil {
		t.Fatal(err)
	}

	var acnt *engine.Account
	if err := cdrsRpc.Call(utils.APIerSv2GetAccount,
		&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "testV2CDRsDuplicateCDRs"}, &acnt); err != nil {
		t.Error(err)
	} else if len(acnt.BalanceMap) != 1 || acnt.BalanceMap[utils.VOICE][0].Value != 600000000000 {
		t.Errorf("Unexpected balance received: %+v", acnt.BalanceMap[utils.VOICE][0])
	}

	args := &engine.ArgV1ProcessEvent{
		Flags: []string{utils.MetaRerate},
		CGREventWithOpts: utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				Event: map[string]interface{}{
					utils.OriginID:     "testV2CDRsDuplicateCDRs",
					utils.OriginHost:   "192.168.1.1",
					utils.Source:       "testV2CDRsDuplicateCDRs",
					utils.RequestType:  utils.META_PSEUDOPREPAID,
					utils.AccountField: "testV2CDRsDuplicateCDRs",
					utils.Subject:      "ANY2CNT",
					utils.Destination:  "+4986517174963",
					utils.AnswerTime:   time.Date(2018, 8, 24, 16, 00, 26, 0, time.UTC),
					utils.Usage:        2 * time.Minute,
					"field_extr1":      "val_extr1",
					"fieldextr2":       "valextr2",
				},
			},
		},
	}

	var rplProcEv []*utils.EventWithFlags
	if err := cdrsRpc.Call(utils.CDRsV2ProcessEvent, args, &rplProcEv); err != nil {
		t.Error("Unexpected error: ", err.Error())
	}

	if err := cdrsRpc.Call(utils.APIerSv2GetAccount,
		&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "testV2CDRsDuplicateCDRs"}, &acnt); err != nil {
		t.Error(err)
	} else if len(acnt.BalanceMap) != 1 || acnt.BalanceMap[utils.VOICE][0].Value != 480000000000 {
		t.Errorf("Unexpected balance received: %+v", acnt.BalanceMap[utils.VOICE][0])
	}

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			var rplProcEv []*utils.EventWithFlags
			args2 := &engine.ArgV1ProcessEvent{
				Flags: []string{utils.MetaRerate},
				CGREventWithOpts: utils.CGREventWithOpts{
					CGREvent: &utils.CGREvent{
						Tenant: "cgrates.org",
						Event: map[string]interface{}{
							utils.OriginID:     "testV2CDRsDuplicateCDRs",
							utils.OriginHost:   "192.168.1.1",
							utils.Source:       "testV2CDRsDuplicateCDRs",
							utils.RequestType:  utils.META_PSEUDOPREPAID,
							utils.AccountField: "testV2CDRsDuplicateCDRs",
							utils.Subject:      "ANY2CNT",
							utils.Destination:  "+4986517174963",
							utils.AnswerTime:   time.Date(2018, 8, 24, 16, 00, 26, 0, time.UTC),
							utils.Usage:        time.Minute,
							"field_extr1":      "val_extr1",
							"fieldextr2":       "valextr2",
						},
					},
				},
			}
			if err := cdrsRpc.Call(utils.CDRsV2ProcessEvent, args2, &rplProcEv); err != nil {
				t.Error("Unexpected error: ", err.Error())
			}
			wg.Done()
		}()
	}
	wg.Wait()

	if err := cdrsRpc.Call(utils.APIerSv2GetAccount,
		&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "testV2CDRsDuplicateCDRs"}, &acnt); err != nil {
		t.Error(err)
	} else if len(acnt.BalanceMap) != 1 || acnt.BalanceMap[utils.VOICE][0].Value != 540000000000 {
		t.Errorf("Unexpected balance received: %+v", acnt.BalanceMap[utils.VOICE][0])
	}
}

func testV2CDRsKillEngine(t *testing.T) {
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
}

func testV2CDRsResetThresholdAction(t *testing.T) {
	var reply string
	if err := cdrsRpc.Call(utils.APIerSv2SetActions, &utils.AttrSetActions{
		ActionsId: "ACT_RESET_THD",
		Actions:   []*utils.TPAction{{Identifier: utils.MetaResetThreshold, ExtraParameters: "cgrates.org:THD_Test"}},
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv2.SetActions received: %s", reply)
	}
	attrs := utils.AttrExecuteAction{Tenant: "cgrates.org", ActionsId: "ACT_RESET_THD"}
	if err := cdrsRpc.Call(utils.APIerSv1ExecuteAction, attrs, &reply); err != nil {
		t.Error(err)
	}
	var td engine.Threshold
	if err := cdrsRpc.Call(utils.ThresholdSv1GetThreshold,
		&utils.TenantIDWithOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "THD_Test"}}, &td); err != nil {
		t.Error(err)
	} else if td.Hits != 0 {
		t.Errorf("Expecting threshold to be reset received: %v", td.Hits)
	}
}
