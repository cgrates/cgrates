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
	"net/rpc/jsonrpc"
	"path"
	"reflect"
	"testing"
	"time"

	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var cdrsCfgPath string
var cdrsCfg *config.CGRConfig
var cdrsRpc *rpc.Client
var cdrsConfDIR string // run the tests for specific configuration

// subtests to be executed for each confDIR
var sTestsCDRsIT = []func(t *testing.T){
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
	testV2CDRsGetCdrsWithRattingPlan,

	testV2CDRsKillEngine,
}

// Tests starting here
func TestCDRsITMySQL(t *testing.T) {
	cdrsConfDIR = "cdrsv2mysql"
	for _, stest := range sTestsCDRsIT {
		t.Run(cdrsConfDIR, stest)
	}
}

func TestCDRsITpg(t *testing.T) {
	cdrsConfDIR = "cdrsv2psql"
	for _, stest := range sTestsCDRsIT {
		t.Run(cdrsConfDIR, stest)
	}
}

func TestCDRsITMongo(t *testing.T) {
	cdrsConfDIR = "cdrsv2mongo"
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
	cdrsRpc, err = jsonrpc.Dial("tcp", cdrsCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

func testV2CDRsLoadTariffPlanFromFolder(t *testing.T) {
	var loadInst utils.LoadInstance
	if err := cdrsRpc.Call("ApierV2.LoadTariffPlanFromFolder",
		&utils.AttrLoadTpFromFolder{FolderPath: path.Join(
			*dataDir, "tariffplans", "testit")}, &loadInst); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond) // Give time for scheduler to execute topups
}

func testV2CDRsProcessCDR(t *testing.T) {
	args := &engine.ArgV1ProcessEvent{
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]interface{}{
				utils.OriginID:    "testV2CDRsProcessCDR1",
				utils.OriginHost:  "192.168.1.1",
				utils.Source:      "testV2CDRsProcessCDR",
				utils.RequestType: utils.META_RATED,
				// utils.Category:    "call", //it will be populated as default in MapEvent.AsCDR
				utils.Account:     "testV2CDRsProcessCDR",
				utils.Subject:     "ANY2CNT",
				utils.Destination: "+4986517174963",
				utils.AnswerTime:  time.Date(2018, 8, 24, 16, 00, 26, 0, time.UTC),
				utils.Usage:       time.Duration(1) * time.Minute,
				"field_extr1":     "val_extr1",
				"fieldextr2":      "valextr2",
			},
		},
	}

	var reply string
	if err := cdrsRpc.Call(utils.CDRsV1ProcessEvent, args, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
	time.Sleep(time.Duration(150) * time.Millisecond) // Give time for CDR to be rated
}

func testV2CDRsGetCdrs(t *testing.T) {
	var cdrCnt int64
	req := utils.AttrGetCdrs{}
	if err := cdrsRpc.Call("ApierV2.CountCDRs", req, &cdrCnt); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if cdrCnt != 3 {
		t.Error("Unexpected number of CDRs returned: ", cdrCnt)
	}
	var cdrs []*engine.ExternalCDR
	args := utils.RPCCDRsFilter{RunIDs: []string{utils.MetaRaw}}
	if err := cdrsRpc.Call(utils.ApierV2GetCDRs, args, &cdrs); err != nil {
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
	if err := cdrsRpc.Call(utils.ApierV2GetCDRs, args, &cdrs); err != nil {
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
	if err := cdrsRpc.Call(utils.ApierV2GetCDRs, args, &cdrs); err != nil {
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
	if err := cdrsRpc.Call("ApierV1.GetRatingProfile", attrGetRatingPlan, &rpl); err != nil {
		t.Errorf("Got error on ApierV1.GetRatingProfile: %+v", err)
	} else if !reflect.DeepEqual(expected, rpl) {
		t.Errorf("Calling ApierV1.GetRatingProfile expected: %+v, received: %+v", utils.ToJSON(expected), utils.ToJSON(rpl))
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
	if err := cdrsRpc.Call("ApierV1.SetRatingProfile", rpf, &reply); err != nil {
		t.Error("Got error on ApierV1.SetRatingProfile: ", err.Error())
	} else if reply != "OK" {
		t.Error("Calling ApierV1.SetRatingProfile got reply: ", reply)
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
	if err := cdrsRpc.Call("ApierV1.GetRatingProfile", attrGetRatingPlan, &rpl); err != nil {
		t.Errorf("Got error on ApierV1.GetRatingProfile: %+v", err)
	} else if !reflect.DeepEqual(expected, rpl) {
		t.Errorf("Calling ApierV1.GetRatingProfile expected: %+v, received: %+v", utils.ToJSON(expected), utils.ToJSON(rpl))
	}

	if err := cdrsRpc.Call(utils.CDRsV1RateCDRs, &engine.ArgRateCDRs{
		RPCCDRsFilter: utils.RPCCDRsFilter{NotRunIDs: []string{utils.MetaRaw}},
		ChargerS:      utils.BoolPointer(true),
	}, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
	time.Sleep(time.Duration(150) * time.Millisecond) // Give time for CDR to be rated
}

func testV2CDRsGetCdrs2(t *testing.T) {
	var cdrCnt int64
	req := utils.AttrGetCdrs{}
	if err := cdrsRpc.Call("ApierV2.CountCDRs", req, &cdrCnt); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if cdrCnt != 3 {
		t.Error("Unexpected number of CDRs returned: ", cdrCnt)
	}
	var cdrs []*engine.ExternalCDR
	args := utils.RPCCDRsFilter{RunIDs: []string{utils.MetaRaw}}
	if err := cdrsRpc.Call(utils.ApierV2GetCDRs, args, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Cost != -1.0 {
			t.Errorf("Unexpected cost for CDR: %f", cdrs[0].Cost)
		}
	}
	args = utils.RPCCDRsFilter{RunIDs: []string{"CustomerCharges"}}
	if err := cdrsRpc.Call(utils.ApierV2GetCDRs, args, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Cost != 0.0198 {
			t.Errorf("Unexpected cost for CDR: %f", cdrs[0].Cost)
		}
	}
	args = utils.RPCCDRsFilter{RunIDs: []string{"SupplierCharges"}}
	if err := cdrsRpc.Call(utils.ApierV2GetCDRs, args, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Cost != 0.0198 {
			t.Errorf("Unexpected cost for CDR: %f", cdrs[0].Cost)
		}
	}
}

func testV2CDRsUsageNegative(t *testing.T) {
	argsCdr := &engine.ArgV1ProcessEvent{
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]interface{}{
				utils.OriginID:    "testV2CDRsUsageNegative",
				utils.OriginHost:  "192.168.1.1",
				utils.Source:      "testV2CDRsUsageNegative",
				utils.RequestType: utils.META_RATED,
				utils.Category:    "call",
				utils.Account:     "testV2CDRsUsageNegative",
				utils.Subject:     "ANY2CNT",
				utils.Destination: "+4986517174963",
				utils.AnswerTime:  time.Date(2018, 8, 24, 16, 00, 26, 0, time.UTC),
				utils.Usage:       -time.Duration(1) * time.Minute,
				"field_extr1":     "val_extr1",
				"fieldextr2":      "valextr2",
			},
		},
	}
	var reply string
	if err := cdrsRpc.Call(utils.CDRsV1ProcessEvent, argsCdr, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
	time.Sleep(time.Duration(150) * time.Millisecond) // Give time for CDR to be rated

	var cdrs []*engine.ExternalCDR
	args := utils.RPCCDRsFilter{RunIDs: []string{utils.MetaRaw}, OriginIDs: []string{"testV2CDRsUsageNegative"}}
	if err := cdrsRpc.Call(utils.ApierV2GetCDRs, args, &cdrs); err != nil {
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
	args = utils.RPCCDRsFilter{RunIDs: []string{"CustomerCharges"}, OriginIDs: []string{"testV2CDRsUsageNegative"}}
	if err := cdrsRpc.Call(utils.ApierV2GetCDRs, args, &cdrs); err != nil {
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
	args = utils.RPCCDRsFilter{RunIDs: []string{"SupplierCharges"}, OriginIDs: []string{"testV2CDRsUsageNegative"}}
	if err := cdrsRpc.Call(utils.ApierV2GetCDRs, args, &cdrs); err != nil {
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
			FilterIDs: []string{"*string:~Tenant:cgrates.com"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			Attributes: []*engine.Attribute{
				{
					FieldName: utils.MetaTenant,
					Type:      utils.META_CONSTANT,
					Value: config.RSRParsers{
						&config.RSRParser{
							Rules:           "CustomTenant",
							AllFiltersMatch: true,
						},
					},
				},
				{
					FieldName: utils.Tenant,
					Type:      utils.META_CONSTANT,
					Value: config.RSRParsers{
						&config.RSRParser{
							Rules:           "CustomTenant",
							AllFiltersMatch: true,
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
	if err := cdrsRpc.Call("ApierV1.SetAttributeProfile", alsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var reply *engine.AttributeProfile
	if err := cdrsRpc.Call("ApierV1.GetAttributeProfile",
		&utils.TenantID{Tenant: "cgrates.com", ID: "ATTR_Tenant"}, &reply); err != nil {
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
	if err := cdrsRpc.Call(utils.ApierV1SetChargerProfile, chargerProfile, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var reply2 *engine.ChargerProfile
	if err := cdrsRpc.Call("ApierV1.GetChargerProfile",
		&utils.TenantID{Tenant: "CustomTenant", ID: "CustomCharger"}, &reply2); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(chargerProfile.ChargerProfile, reply2) {
		t.Errorf("Expecting : %+v, received: %+v", chargerProfile.ChargerProfile, reply2)
	}

	argsCdr := &engine.ArgV1ProcessEvent{
		AttributeS: utils.BoolPointer(true),
		ChargerS:   utils.BoolPointer(true),
		StatS:      utils.BoolPointer(false),
		ThresholdS: utils.BoolPointer(false),
		Store:      utils.BoolPointer(true),
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.com",
			Event: map[string]interface{}{
				utils.OriginID:    "testV2CDRsDifferentTenants",
				utils.OriginHost:  "192.168.1.1",
				utils.Source:      "testV2CDRsDifferentTenants",
				utils.RequestType: utils.META_RATED,
				utils.Category:    "call",
				utils.Account:     "testV2CDRsDifferentTenants",
				utils.Destination: "+4986517174963",
				utils.Tenant:      "cgrates.com",
				utils.AnswerTime:  time.Date(2018, 8, 24, 16, 00, 26, 0, time.UTC),
				utils.Usage:       time.Duration(1) * time.Second,
				"field_extr1":     "val_extr1",
				"fieldextr2":      "valextr2",
			},
		},
	}
	var reply3 string
	if err := cdrsRpc.Call(utils.CDRsV1ProcessEvent, argsCdr, &reply3); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if reply3 != utils.OK {
		t.Error("Unexpected reply received: ", reply3)
	}
	time.Sleep(time.Duration(150) * time.Millisecond) // Give time for CDR to be rated

	var cdrs []*engine.ExternalCDR
	args := utils.RPCCDRsFilter{Tenants: []string{"CustomTenant"}}
	if err := cdrsRpc.Call(utils.ApierV2GetCDRs, args, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 2 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	}
}

func testV2CDRsRemoveRatingProfiles(t *testing.T) {
	var reply string
	if err := cdrsRpc.Call(utils.ApierV1RemoveRatingProfile, &v1.AttrRemoveRatingProfile{
		Tenant:   "cgrates.org",
		Category: utils.CALL,
		Subject:  utils.ANY,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected: %s, received: %s ", utils.OK, reply)
	}
	if err := cdrsRpc.Call(utils.ApierV1RemoveRatingProfile, &v1.AttrRemoveRatingProfile{
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
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]interface{}{
				utils.OriginID:    "testV2CDRsProcessCDR4",
				utils.OriginHost:  "192.168.1.1",
				utils.Source:      "testV2CDRsProcessCDR4",
				utils.RequestType: utils.META_RATED,
				utils.Account:     "testV2CDRsProcessCDR4",
				utils.Subject:     "NoSubject",
				utils.Destination: "+1234567",
				utils.AnswerTime:  time.Date(2018, 8, 24, 16, 00, 26, 0, time.UTC),
				utils.Usage:       time.Duration(1) * time.Minute,
				"field_extr1":     "val_extr1",
				"fieldextr2":      "valextr2",
			},
		},
	}

	var reply string
	if err := cdrsRpc.Call(utils.CDRsV1ProcessEvent, args, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
	time.Sleep(time.Duration(150) * time.Millisecond) // Give time for CDR to be rated
}

func testV2CDRsGetCdrsNoRattingPlan(t *testing.T) {
	var cdrCnt int64
	req := utils.AttrGetCdrs{}
	if err := cdrsRpc.Call("ApierV2.CountCDRs", req, &cdrCnt); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if cdrCnt != 11 {
		t.Error("Unexpected number of CDRs returned: ", cdrCnt)
	}
	var cdrs []*engine.ExternalCDR
	args := utils.RPCCDRsFilter{RunIDs: []string{utils.MetaRaw}, Accounts: []string{"testV2CDRsProcessCDR4"}}
	if err := cdrsRpc.Call(utils.ApierV2GetCDRs, args, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Cost != -1.0 {
			t.Errorf("Unexpected cost for CDR: %f", cdrs[0].Cost)
		}
	}
	args = utils.RPCCDRsFilter{RunIDs: []string{"CustomerCharges"}, Accounts: []string{"testV2CDRsProcessCDR4"}}
	if err := cdrsRpc.Call(utils.ApierV2GetCDRs, args, &cdrs); err != nil {
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
	if err := cdrsRpc.Call(utils.ApierV2GetCDRs, args, &cdrs); err != nil {
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
	if err := cdrsRpc.Call("ApierV1.SetRatingProfile", rpf, &reply); err != nil {
		t.Error("Got error on ApierV1.SetRatingProfile: ", err.Error())
	} else if reply != "OK" {
		t.Error("Calling ApierV1.SetRatingProfile got reply: ", reply)
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
	if err := cdrsRpc.Call("ApierV1.SetRatingProfile", rpf, &reply); err != nil {
		t.Error("Got error on ApierV1.SetRatingProfile: ", err.Error())
	} else if reply != "OK" {
		t.Error("Calling ApierV1.SetRatingProfile got reply: ", reply)
	}

	if err := cdrsRpc.Call(utils.CDRsV1RateCDRs, &engine.ArgRateCDRs{
		RPCCDRsFilter: utils.RPCCDRsFilter{NotRunIDs: []string{utils.MetaRaw}, Accounts: []string{"testV2CDRsProcessCDR4"}},
		ChargerS:      utils.BoolPointer(true),
	}, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
	time.Sleep(time.Duration(150) * time.Millisecond) // Give time for CDR to be rated
}

func testV2CDRsGetCdrsWithRattingPlan(t *testing.T) {
	var cdrCnt int64
	req := utils.AttrGetCdrs{}
	if err := cdrsRpc.Call("ApierV2.CountCDRs", req, &cdrCnt); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if cdrCnt != 11 {
		t.Error("Unexpected number of CDRs returned: ", cdrCnt)
	}
	var cdrs []*engine.ExternalCDR
	args := utils.RPCCDRsFilter{RunIDs: []string{utils.MetaRaw}, Accounts: []string{"testV2CDRsProcessCDR4"}}
	if err := cdrsRpc.Call(utils.ApierV2GetCDRs, args, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Cost != -1.0 {
			t.Errorf("Unexpected cost for CDR: %f", cdrs[0].Cost)
		}
	}
	args = utils.RPCCDRsFilter{RunIDs: []string{"CustomerCharges"}, Accounts: []string{"testV2CDRsProcessCDR4"}}
	if err := cdrsRpc.Call(utils.ApierV2GetCDRs, args, &cdrs); err != nil {
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
	args = utils.RPCCDRsFilter{RunIDs: []string{"SupplierCharges"}, Accounts: []string{"testV2CDRsProcessCDR4"}}
	if err := cdrsRpc.Call(utils.ApierV2GetCDRs, args, &cdrs); err != nil {
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

func testV2CDRsKillEngine(t *testing.T) {
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
}
