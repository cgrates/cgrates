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

package v1

import (
	"net/rpc"
	"os/exec"
	"path"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/cgrates/cgrates/scheduler"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/dispatchers"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	apierCfgPath      string
	apierCfg          *config.CGRConfig
	apierRPC          *rpc.Client
	APIerSv2ConfigDIR string //run tests for specific configuration

	sTestsAPIer = []func(t *testing.T){
		testAPIerInitCfg,
		testAPIerInitDataDb,
		testAPIerResetStorDb,
		testAPIerStartEngine,
		testAPIerRPCConn,
		testAPIerLoadFromFolder,
		testAPIerVerifyAttributesAfterLoad,
		testAPIerRemoveTPFromFolder,
		testAPIerAfterDelete,
		testAPIerVerifyAttributesAfterDelete,
		testAPIerLoadFromFolder,
		testAPIerGetRatingPlanCost,
		testAPIerGetRatingPlanCost2,
		testAPIerGetRatingPlanCost3,
		testAPIerGetActionPlanIDs,
		testAPIerGetRatingPlanIDs,
		testAPIerSetActionPlanDfltTime,
		testAPIerLoadRatingPlan,
		testAPIerLoadRatingPlan2,
		testAPIerLoadRatingProfile,
		testAPIerLoadFromFolderAccountAction,
		testAPIerKillEngine,

		testAPIerInitDataDb,
		testAPIerResetStorDb,
		testAPIerStartEngineSleep,
		testAPIerRPCConn,
		testApierSetAndRemoveRatingProfileAnySubject,
		testAPIerKillEngine,
	}
)

//Test start here
func TestApierIT2(t *testing.T) {
	// no need for a new config with *gob transport in this case
	switch *dbType {
	case utils.MetaInternal:
		APIerSv2ConfigDIR = "tutinternal"
		sTestsAPIer = sTestsAPIer[:len(sTestsAPIer)-6]
	case utils.MetaMySQL:
		APIerSv2ConfigDIR = "tutmysql"
	case utils.MetaMongo:
		APIerSv2ConfigDIR = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsAPIer {
		t.Run(APIerSv2ConfigDIR, stest)
	}
}

func testAPIerInitCfg(t *testing.T) {
	var err error
	apierCfgPath = path.Join(*dataDir, "conf", "samples", APIerSv2ConfigDIR)
	apierCfg, err = config.NewCGRConfigFromPath(apierCfgPath)
	if err != nil {
		t.Error(err)
	}
}

func testAPIerInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(apierCfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func testAPIerResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(apierCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testAPIerStartEngineSleep(t *testing.T) {
	time.Sleep(500 * time.Millisecond)
	if _, err := engine.StopStartEngine(apierCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testAPIerStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(apierCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testAPIerRPCConn(t *testing.T) {
	var err error
	apierRPC, err = newRPCClient(apierCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testAPIerLoadFromFolder(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "tutorial")}
	if err := apierRPC.Call(utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(100 * time.Millisecond)
}

func testAPIerVerifyAttributesAfterLoad(t *testing.T) {
	ev := &engine.AttrArgsProcessEvent{
		Context: utils.StringPointer("simpleauth"),
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testAPIerAfterDelete",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
			},
		},
	}

	eAttrPrf := &engine.AttributeProfileWithAPIOpts{
		AttributeProfile: &engine.AttributeProfile{
			Tenant:    ev.Tenant,
			ID:        "ATTR_1001_SIMPLEAUTH",
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Contexts:  []string{"simpleauth"},
			Attributes: []*engine.Attribute{
				{
					FilterIDs: []string{},
					Path:      utils.MetaReq + utils.NestingSep + "Password",
					Type:      utils.MetaConstant,
					Value:     config.NewRSRParsersMustCompile("CGRateS.org", utils.InfieldSep),
				},
			},
			Weight: 20.0,
		},
	}
	if *encoding == utils.MetaGOB {
		eAttrPrf.Attributes[0].FilterIDs = nil // in gob emtpty slice is encoded as nil
	}
	eAttrPrf.Compile()
	var attrReply *engine.AttributeProfile
	if err := apierRPC.Call(utils.AttributeSv1GetAttributeForEvent,
		ev, &attrReply); err != nil {
		t.Error(err)
	}
	if attrReply == nil {
		t.Errorf("Expecting attrReply to not be nil")
		// attrReply shoud not be nil so exit function
		// to avoid nil segmentation fault;
		// if this happens try to run this test manualy
		return
	}
	attrReply.Compile() // Populate private variables in RSRParsers
	if !reflect.DeepEqual(eAttrPrf.AttributeProfile, attrReply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eAttrPrf.AttributeProfile), utils.ToJSON(attrReply))
	}
}

func testAPIerRemoveTPFromFolder(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "tutorial")}
	if err := apierRPC.Call(utils.APIerSv1RemoveTPFromFolder, attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(100 * time.Millisecond)
}

func testAPIerAfterDelete(t *testing.T) {
	var reply *engine.AttributeProfile
	if err := apierRPC.Call(utils.APIerSv1GetAttributeProfile,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ATTR_1001_SIMPLEAUTH"}}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatal(err)
	}
	var replyTh *engine.ThresholdProfile
	if err := apierRPC.Call(utils.APIerSv1GetThresholdProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "THD_ACNT_1001"}, &replyTh); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}

}

func testAPIerVerifyAttributesAfterDelete(t *testing.T) {
	ev := &engine.AttrArgsProcessEvent{
		Context: utils.StringPointer("simpleauth"),
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testAPIerAfterDelete",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
			},
		},
	}
	var attrReply *engine.AttributeProfile
	if err := apierRPC.Call(utils.AttributeSv1GetAttributeForEvent,
		ev, &attrReply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testAPIerGetRatingPlanCost(t *testing.T) {
	arg := &utils.RatingPlanCostArg{
		Destination:   "1002",
		RatingPlanIDs: []string{"RP_1001", "RP_1002"},
		SetupTime:     utils.MetaNow,
		Usage:         "1h",
	}
	var reply dispatchers.RatingPlanCost
	if err := apierRPC.Call(utils.RALsV1GetRatingPlansCost, arg, &reply); err != nil {
		t.Error(err)
	} else if reply.RatingPlanID != "RP_1001" {
		t.Error("Unexpected RatingPlanID: ", reply.RatingPlanID)
	} else if *reply.EventCost.Cost != 6.5118 {
		t.Error("Unexpected Cost: ", *reply.EventCost.Cost)
	} else if *reply.EventCost.Usage != time.Hour {
		t.Error("Unexpected Usage: ", *reply.EventCost.Usage)
	}
}

// we need to discuss about this case
// because 1003 have the following DestinationRate
// DR_1003_MAXCOST_DISC,DST_1003,RT_1CNT_PER_SEC,*up,4,0.12,*disconnect
func testAPIerGetRatingPlanCost2(t *testing.T) {
	arg := &utils.RatingPlanCostArg{
		Destination:   "1003",
		RatingPlanIDs: []string{"RP_1001", "RP_1002"},
		SetupTime:     utils.MetaNow,
		Usage:         "1h",
	}
	var reply dispatchers.RatingPlanCost
	if err := apierRPC.Call(utils.RALsV1GetRatingPlansCost, arg, &reply); err != nil {
		t.Error(err)
	} else if reply.RatingPlanID != "RP_1001" {
		t.Error("Unexpected RatingPlanID: ", reply.RatingPlanID)
	} else if *reply.EventCost.Cost != 36 {
		t.Error("Unexpected Cost: ", *reply.EventCost.Cost)
	} else if *reply.EventCost.Usage != time.Hour {
		t.Error("Unexpected Usage: ", *reply.EventCost.Usage)
	}
}

func testAPIerGetRatingPlanCost3(t *testing.T) {
	arg := &utils.RatingPlanCostArg{
		Destination:   "1001",
		RatingPlanIDs: []string{"RP_1001", "RP_1002"},
		SetupTime:     utils.MetaNow,
		Usage:         "1h",
	}
	var reply dispatchers.RatingPlanCost
	if err := apierRPC.Call(utils.RALsV1GetRatingPlansCost, arg, &reply); err != nil {
		t.Error(err)
	} else if reply.RatingPlanID != "RP_1002" {
		t.Error("Unexpected RatingPlanID: ", reply.RatingPlanID)
	} else if *reply.EventCost.Cost != 6.5118 {
		t.Error("Unexpected Cost: ", *reply.EventCost.Cost)
	} else if *reply.EventCost.Usage != time.Hour {
		t.Error("Unexpected Usage: ", *reply.EventCost.Usage)
	}
}

func testAPIerGetActionPlanIDs(t *testing.T) {
	var reply []string
	if err := apierRPC.Call(utils.APIerSv1GetActionPlanIDs,
		&utils.PaginatorWithTenant{Tenant: "cgrates.org"},
		&reply); err != nil {
		t.Error(err)
	} else if len(reply) != 1 {
		t.Errorf("Expected: 1 , received: <%+v>", len(reply))
	} else if reply[0] != "AP_PACKAGE_10" {
		t.Errorf("Expected: AP_PACKAGE_10 , received: <%+v>", reply[0])
	}
}

func testAPIerGetRatingPlanIDs(t *testing.T) {
	var reply []string
	expected := []string{"RP_1002_LOW", "RP_1003", "RP_1001", "RP_MMS", "RP_SMS", "RP_1002"}
	if err := apierRPC.Call(utils.APIerSv1GetRatingPlanIDs,
		&utils.PaginatorWithTenant{Tenant: "cgrates.org"},
		&reply); err != nil {
		t.Error(err)
	}
	sort.Strings(reply)
	sort.Strings(expected)
	if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expected: <%+v> , received: <%+v>", utils.ToJSON(expected), utils.ToJSON(reply))

	}
}

func testAPIerSetActionPlanDfltTime(t *testing.T) {
	var reply1 string
	hourlyAP := &AttrSetActionPlan{
		Id: "AP_HOURLY",
		ActionPlan: []*AttrActionPlan{
			{
				ActionsId: "ACT_TOPUP_RST_10",
				Time:      utils.MetaHourly,
				Weight:    20.0,
			},
		},
		ReloadScheduler: true,
	}
	if err := apierRPC.Call(utils.APIerSv1SetActionPlan, &hourlyAP, &reply1); err != nil {
		t.Error("Got error on APIerSv1.SetActionPlan: ", err.Error())
	} else if reply1 != utils.OK {
		t.Errorf("Calling APIerSv1.SetActionPlan received: %s", reply1)
	}
	dailyAP := &AttrSetActionPlan{
		Id: "AP_DAILY",
		ActionPlan: []*AttrActionPlan{
			{
				ActionsId: "ACT_TOPUP_RST_10",
				Time:      utils.MetaDaily,
				Weight:    20.0,
			},
		},
		ReloadScheduler: true,
	}
	if err := apierRPC.Call(utils.APIerSv1SetActionPlan, &dailyAP, &reply1); err != nil {
		t.Error("Got error on APIerSv1.SetActionPlan: ", err.Error())
	} else if reply1 != utils.OK {
		t.Errorf("Calling APIerSv1.SetActionPlan received: %s", reply1)
	}
	weeklyAP := &AttrSetActionPlan{
		Id: "AP_WEEKLY",
		ActionPlan: []*AttrActionPlan{
			{
				ActionsId: "ACT_TOPUP_RST_10",
				Time:      utils.MetaWeekly,
				Weight:    20.0,
			},
		},
		ReloadScheduler: true,
	}
	if err := apierRPC.Call(utils.APIerSv1SetActionPlan, &weeklyAP, &reply1); err != nil {
		t.Error("Got error on APIerSv1.SetActionPlan: ", err.Error())
	} else if reply1 != utils.OK {
		t.Errorf("Calling APIerSv1.SetActionPlan received: %s", reply1)
	}
	monthlyAP := &AttrSetActionPlan{
		Id: "AP_MONTHLY",
		ActionPlan: []*AttrActionPlan{
			{
				ActionsId: "ACT_TOPUP_RST_10",
				Time:      utils.MetaMonthly,
				Weight:    20.0,
			},
		},
		ReloadScheduler: true,
	}
	if err := apierRPC.Call(utils.APIerSv1SetActionPlan, &monthlyAP, &reply1); err != nil {
		t.Error("Got error on APIerSv1.SetActionPlan: ", err.Error())
	} else if reply1 != utils.OK {
		t.Errorf("Calling APIerSv1.SetActionPlan received: %s", reply1)
	}
	var rply []*scheduler.ScheduledAction
	if err := apierRPC.Call(utils.APIerSv1GetScheduledActions,
		scheduler.ArgsGetScheduledActions{}, &rply); err != nil {
		t.Error("Unexpected error: ", err)
	} else {
		for _, schedAct := range rply {
			switch schedAct.ActionPlanID {
			case "AP_WEEKLY":
				t1 := time.Now().AddDate(0, 0, 7)
				if schedAct.NextRunTime.Before(t1.Add(-2*time.Second)) ||
					schedAct.NextRunTime.After(t1.Add(time.Second)) {
					t.Errorf("Expected the nextRuntime to be after 1 week,but received: <%+v>", utils.ToJSON(schedAct))
				}
			case "AP_DAILY":
				t1 := time.Now().AddDate(0, 0, 1)
				if schedAct.NextRunTime.Before(t1.Add(-2*time.Second)) ||
					schedAct.NextRunTime.After(t1.Add(time.Second)) {
					t.Errorf("Expected the nextRuntime to be after 1 day,but received: <%+v>", utils.ToJSON(schedAct))
				}
			case "AP_HOURLY":
				if schedAct.NextRunTime.Before(time.Now().Add(59*time.Minute+58*time.Second)) ||
					schedAct.NextRunTime.After(time.Now().Add(time.Hour+time.Second)) {
					t.Errorf("Expected the nextRuntime to be after 1 hour,but received: <%+v>", utils.ToJSON(schedAct))
				}
			case "AP_MONTHLY":
				// *monthly needs to mach exactly the day
				tnow := time.Now()
				expected := tnow.AddDate(0, 1, 0)
				expected = time.Date(expected.Year(), expected.Month(), tnow.Day(), tnow.Hour(),
					tnow.Minute(), tnow.Second(), 0, schedAct.NextRunTime.Location())
				if schedAct.NextRunTime.Before(expected.Add(-time.Second)) ||
					schedAct.NextRunTime.After(expected.Add(time.Second)) {
					t.Errorf("Expected the nextRuntime to be after 1 month,but received: <%+v>", utils.ToJSON(schedAct))
				}
			}
		}
	}
}

func testAPIerLoadRatingPlan(t *testing.T) {
	attrs := utils.AttrSetDestination{Id: "DEST_CUSTOM", Prefixes: []string{"+4986517174963", "+4986517174960"}}
	var reply string
	if err := apierRPC.Call(utils.APIerSv1SetDestination, &attrs, &reply); err != nil {
		t.Error("Unexpected error", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}

	rt := &utils.TPRateRALs{TPid: "TP_SAMPLE", ID: "SAMPLE_RATE_ID", RateSlots: []*utils.RateSlot{
		{ConnectFee: 0, Rate: 0, RateUnit: "1s", RateIncrement: "1s", GroupIntervalStart: "0s"},
	}}
	if err := apierRPC.Call(utils.APIerSv1SetTPRate, rt, &reply); err != nil {
		t.Error("Got error on APIerSv1.SetTPRate: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received when calling APIerSv1.SetTPRate: ", reply)
	}

	dr := &utils.TPDestinationRate{TPid: "TP_SAMPLE", ID: "DR_SAMPLE_DESTINATION_RATE", DestinationRates: []*utils.DestinationRate{
		{DestinationId: "DEST_CUSTOM", RateId: "SAMPLE_RATE_ID",
			RoundingMethod: "*up", RoundingDecimals: 4},
	}}
	if err := apierRPC.Call(utils.APIerSv1SetTPDestinationRate, dr, &reply); err != nil {
		t.Error("Got error on APIerSv1.SetTPDestinationRate: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received when calling APIerSv1.SetTPDestinationRate: ", reply)
	}

	rp := &utils.TPRatingPlan{TPid: "TP_SAMPLE", ID: "RPl_SAMPLE_RATING_PLAN",
		RatingPlanBindings: []*utils.TPRatingPlanBinding{
			{DestinationRatesId: "DR_SAMPLE_DESTINATION_RATE", TimingId: utils.MetaAny,
				Weight: 10},
		}}

	if err := apierRPC.Call(utils.APIerSv1SetTPRatingPlan, rp, &reply); err != nil {
		t.Error("Got error on APIerSv1.SetTPRatingPlan: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received when calling APIerSv1.SetTPRatingPlan: ", reply)
	}

	if err := apierRPC.Call(utils.APIerSv1LoadRatingPlan, &AttrLoadRatingPlan{TPid: "TP_SAMPLE", RatingPlanId: "RPl_SAMPLE_RATING_PLAN"}, &reply); err != nil {
		t.Error("Got error on APIerSv1.LoadRatingPlan: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Calling APIerSv1.LoadRatingPlan got reply: ", reply)
	}

	rpRply := new(engine.RatingPlan)
	rplnId := "RPl_SAMPLE_RATING_PLAN"
	if err := apierRPC.Call(utils.APIerSv1GetRatingPlan, &rplnId, rpRply); err != nil {
		t.Error("Got error on APIerSv1.GetRatingPlan: ", err.Error())
	}

}

func testAPIerLoadRatingPlan2(t *testing.T) {
	var reply string

	dr := &utils.TPDestinationRate{TPid: "TP_SAMPLE", ID: "DR_WITH_ERROR", DestinationRates: []*utils.DestinationRate{
		{DestinationId: "DST_NOT_FOUND", RateId: "SAMPLE_RATE_ID",
			RoundingMethod: "*up", RoundingDecimals: 4},
	}}
	if err := apierRPC.Call(utils.APIerSv1SetTPDestinationRate, dr, &reply); err != nil {
		t.Error("Got error on APIerSv1.SetTPDestinationRate: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received when calling APIerSv1.SetTPDestinationRate: ", reply)
	}

	rp := &utils.TPRatingPlan{TPid: "TP_SAMPLE", ID: "RPL_WITH_ERROR",
		RatingPlanBindings: []*utils.TPRatingPlanBinding{
			{DestinationRatesId: "DR_WITH_ERROR", TimingId: utils.MetaAny,
				Weight: 10},
		}}

	if err := apierRPC.Call(utils.APIerSv1SetTPRatingPlan, rp, &reply); err != nil {
		t.Error("Got error on APIerSv1.SetTPRatingPlan: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received when calling APIerSv1.SetTPRatingPlan: ", reply)
	}

	if err := apierRPC.Call(utils.APIerSv1LoadRatingPlan,
		&AttrLoadRatingPlan{TPid: "TP_SAMPLE", RatingPlanId: "RPL_WITH_ERROR"}, &reply); err == nil {
		t.Error("Expected to get error: ", err)
	}

}

func testAPIerLoadRatingProfile(t *testing.T) {
	var reply string
	rpf := &utils.TPRatingProfile{
		TPid:     "TP_SAMPLE",
		LoadId:   "TP_SAMPLE",
		Tenant:   "cgrates.org",
		Category: "call",
		Subject:  utils.MetaAny,
		RatingPlanActivations: []*utils.TPRatingActivation{{
			ActivationTime:   "2012-01-01T00:00:00Z",
			RatingPlanId:     "RPl_SAMPLE_RATING_PLAN",
			FallbackSubjects: utils.EmptyString,
		}},
	}
	// add a TPRatingProfile
	if err := apierRPC.Call(utils.APIerSv1SetTPRatingProfile, rpf, &reply); err != nil {
		t.Error(err)
	}
	// load the TPRatingProfile into dataDB
	argsRPrf := &utils.TPRatingProfile{
		TPid: "TP_SAMPLE", LoadId: "TP_SAMPLE",
		Tenant: "cgrates.org", Category: "call", Subject: "*any"}
	if err := apierRPC.Call(utils.APIerSv1LoadRatingProfile, argsRPrf, &reply); err != nil {
		t.Error(err)
	}

	// verify if was added correctly
	var rpl engine.RatingProfile
	attrGetRatingPlan := &utils.AttrGetRatingProfile{
		Tenant: "cgrates.org", Category: "call", Subject: utils.MetaAny}
	actTime, err := utils.ParseTimeDetectLayout("2012-01-01T00:00:00Z", utils.EmptyString)
	if err != nil {
		t.Error(err)
	}
	expected := engine.RatingProfile{
		Id: "*out:cgrates.org:call:*any",
		RatingPlanActivations: engine.RatingPlanActivations{
			{
				ActivationTime: actTime,
				RatingPlanId:   "RPl_SAMPLE_RATING_PLAN",
			},
		},
	}
	if err := apierRPC.Call(utils.APIerSv1GetRatingProfile, attrGetRatingPlan, &rpl); err != nil {
		t.Errorf("Got error on APIerSv1.GetRatingProfile: %+v", err)
	} else if !reflect.DeepEqual(expected, rpl) {
		t.Errorf("Calling APIerSv1.GetRatingProfile expected: %+v, received: %+v", utils.ToJSON(expected), utils.ToJSON(rpl))
	}

	// add new RatingPlan
	rp := &utils.TPRatingPlan{TPid: "TP_SAMPLE", ID: "RPl_SAMPLE_RATING_PLAN2",
		RatingPlanBindings: []*utils.TPRatingPlanBinding{
			{DestinationRatesId: "DR_SAMPLE_DESTINATION_RATE", TimingId: utils.MetaAny,
				Weight: 10},
		}}

	if err := apierRPC.Call(utils.APIerSv1SetTPRatingPlan, rp, &reply); err != nil {
		t.Error("Got error on APIerSv1.SetTPRatingPlan: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received when calling APIerSv1.SetTPRatingPlan: ", reply)
	}

	if err := apierRPC.Call(utils.APIerSv1LoadRatingPlan, &AttrLoadRatingPlan{TPid: "TP_SAMPLE", RatingPlanId: "RPl_SAMPLE_RATING_PLAN2"}, &reply); err != nil {
		t.Error("Got error on APIerSv1.LoadRatingPlan: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Calling APIerSv1.LoadRatingPlan got reply: ", reply)
	}

	// overwrite the existing TPRatingProfile with a new RatingPlanActivations
	rpf = &utils.TPRatingProfile{
		TPid:     "TP_SAMPLE",
		LoadId:   "TP_SAMPLE",
		Tenant:   "cgrates.org",
		Category: "call",
		Subject:  utils.MetaAny,
		RatingPlanActivations: []*utils.TPRatingActivation{
			{
				ActivationTime:   "2012-01-01T00:00:00Z",
				RatingPlanId:     "RPl_SAMPLE_RATING_PLAN",
				FallbackSubjects: utils.EmptyString,
			},
			{
				ActivationTime:   "2012-02-02T00:00:00Z",
				RatingPlanId:     "RPl_SAMPLE_RATING_PLAN2",
				FallbackSubjects: utils.EmptyString,
			},
		},
	}

	if err := apierRPC.Call(utils.APIerSv1SetTPRatingProfile, rpf, &reply); err != nil {
		t.Error(err)
	}

	// load the TPRatingProfile into dataDB
	// because the RatingProfile exists the RatingPlanActivations will be merged
	if err := apierRPC.Call(utils.APIerSv1LoadRatingProfile, argsRPrf, &reply); err != nil {
		t.Error(err)
	}
	actTime2, err := utils.ParseTimeDetectLayout("2012-02-02T00:00:00Z", utils.EmptyString)
	if err != nil {
		t.Error(err)
	}
	expected = engine.RatingProfile{
		Id: "*out:cgrates.org:call:*any",
		RatingPlanActivations: engine.RatingPlanActivations{
			{
				ActivationTime: actTime,
				RatingPlanId:   "RPl_SAMPLE_RATING_PLAN",
			},
			{
				ActivationTime: actTime2,
				RatingPlanId:   "RPl_SAMPLE_RATING_PLAN2",
			},
		},
	}
	if err := apierRPC.Call(utils.APIerSv1GetRatingProfile, attrGetRatingPlan, &rpl); err != nil {
		t.Errorf("Got error on APIerSv1.GetRatingProfile: %+v", err)
	} else if !reflect.DeepEqual(expected, rpl) {
		t.Errorf("Calling APIerSv1.GetRatingProfile expected: %+v, received: %+v", utils.ToJSON(expected), utils.ToJSON(rpl))
	}

}

func testAPIerLoadFromFolderAccountAction(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "tutorial")}
	if err := apierRPC.Call(utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(100 * time.Millisecond)
	attrs2 := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "account_action_from_tutorial")}
	if err := apierRPC.Call(utils.APIerSv1LoadTariffPlanFromFolder, attrs2, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(100 * time.Millisecond)
	var acnt *engine.Account
	attrAcnt := &utils.AttrGetAccount{
		Tenant:  "cgrates.org",
		Account: "AccountWithAPFromTutorial",
	}
	if err := apierRPC.Call(utils.APIerSv2GetAccount, attrAcnt, &acnt); err != nil {
		t.Error(err)
	} else if rply := acnt.BalanceMap[utils.MetaMonetary].GetTotalValue(); rply != 10.0 {
		t.Errorf("Expecting: %v, received: %v",
			10.0, rply)
	}
}

func testApierSetAndRemoveRatingProfileAnySubject(t *testing.T) {
	loader := exec.Command("cgr-loader", "-config_path", apierCfgPath, "-path", path.Join(*dataDir, "tariffplans", "tutorial"))
	if err := loader.Run(); err != nil {
		t.Error(err)
	}

	rpf := &utils.AttrSetRatingProfile{
		Tenant:   "cgrates.org",
		Category: "call",
		Subject:  "SUPPLIER1",
		RatingPlanActivations: []*utils.TPRatingActivation{
			{
				ActivationTime: "2018-01-01T00:00:00Z",
				RatingPlanId:   "RP_SMS",
			},
		},
		Overwrite: true,
	}
	var reply string
	if err := apierRPC.Call(utils.APIerSv1SetRatingProfile, rpf, &reply); err != nil {
		t.Error("Got error on APIerSv1.SetRatingProfile: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Calling APIerSv1.SetRatingProfile got reply: ", reply)
	}

	expected := engine.RatingProfile{
		Id: "*out:cgrates.org:sms:*any",
		RatingPlanActivations: engine.RatingPlanActivations{
			{
				ActivationTime: time.Date(2014, 1, 14, 0, 0, 0, 0, time.UTC),
				RatingPlanId:   "RP_SMS",
			},
		},
	}
	attrGetRatingPlan := &utils.AttrGetRatingProfile{
		Tenant: "cgrates.org", Category: "sms", Subject: utils.MetaAny}
	var rpl engine.RatingProfile
	if err := apierRPC.Call(utils.APIerSv1GetRatingProfile, attrGetRatingPlan, &rpl); err != nil {
		t.Errorf("Got error on APIerSv1.GetRatingProfile: %+v", err)
	} else if !reflect.DeepEqual(expected, rpl) {
		t.Errorf("Calling APIerSv1.GetRatingProfile expected: %+v, received: %+v", utils.ToJSON(expected), utils.ToJSON(rpl))
	}

	if err := apierRPC.Call(utils.APIerSv1RemoveRatingProfile, &AttrRemoveRatingProfile{
		Tenant:   "cgrates.org",
		Category: "sms",
		Subject:  utils.MetaAny,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected: %s, received: %s ", utils.OK, reply)
	}

	if err := apierRPC.Call(utils.APIerSv1GetRatingProfile,
		attrGetRatingPlan, &rpl); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected %v, \n but received %v", utils.ErrNotFound, err)
	}
}

func testAPIerKillEngine(t *testing.T) {
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
}
