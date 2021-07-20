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
	apierCfgPath = path.Join(costDataDir, "conf", "samples", APIerSv2ConfigDIR)
	apierCfg, err = config.NewCGRConfigFromPath(apierCfgPath)
	if err != nil {
		t.Error(err)
	}
	apierCfg.DataFolderPath = costDataDir // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(apierCfg)
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
	time.Sleep(500*time.Millisecond)
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
	time.Sleep(500 * time.Millisecond)
}

func testAPIerVerifyAttributesAfterLoad(t *testing.T) {
	ev := &engine.AttrArgsProcessEvent{
		Context: utils.StringPointer("simpleauth"),
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testAPIerAfterDelete",
			Event: map[string]interface{}{
				utils.Account: "1001",
			},
		},
	}

	eAttrPrf := &AttributeWithCache{
		AttributeProfile: &engine.AttributeProfile{
			Tenant:    ev.Tenant,
			ID:        "ATTR_1001_SIMPLEAUTH",
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Contexts:  []string{"simpleauth"},
			Attributes: []*engine.Attribute{
				{
					FilterIDs: []string{},
					Path:      utils.MetaReq + utils.NestingSep + "Password",
					Type:      utils.META_CONSTANT,
					Value:     config.NewRSRParsersMustCompile("CGRateS.org", true, utils.INFIELD_SEP),
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
	time.Sleep(500 * time.Millisecond)
}

func testAPIerAfterDelete(t *testing.T) {
	var reply *engine.AttributeProfile
	if err := apierRPC.Call(utils.APIerSv1GetAttributeProfile,
		utils.TenantIDWithArgDispatcher{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ATTR_1001_SIMPLEAUTH"}}, &reply); err == nil ||
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
				utils.Account: "1001",
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
		SetupTime:     utils.META_NOW,
		Usage:         "1h",
	}
	var reply dispatchers.RatingPlanCost
	if err := apierRPC.Call(utils.RALsV1GetRatingPlansCost, arg, &reply); err != nil {
		t.Error(err)
	} else if reply.RatingPlanID != "RP_1001" {
		t.Error("Unexpected RatingPlanID: ", reply.RatingPlanID)
	} else if *reply.EventCost.Cost != 6.5118 {
		t.Error("Unexpected Cost: ", *reply.EventCost.Cost)
	} else if *reply.EventCost.Usage != time.Duration(time.Hour) {
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
		SetupTime:     utils.META_NOW,
		Usage:         "1h",
	}
	var reply dispatchers.RatingPlanCost
	if err := apierRPC.Call(utils.RALsV1GetRatingPlansCost, arg, &reply); err != nil {
		t.Error(err)
	} else if reply.RatingPlanID != "RP_1001" {
		t.Error("Unexpected RatingPlanID: ", reply.RatingPlanID)
	} else if *reply.EventCost.Cost != 36 {
		t.Error("Unexpected Cost: ", *reply.EventCost.Cost)
	} else if *reply.EventCost.Usage != time.Duration(time.Hour) {
		t.Error("Unexpected Usage: ", *reply.EventCost.Usage)
	}
}

func testAPIerGetRatingPlanCost3(t *testing.T) {
	arg := &utils.RatingPlanCostArg{
		Destination:   "1001",
		RatingPlanIDs: []string{"RP_1001", "RP_1002"},
		SetupTime:     utils.META_NOW,
		Usage:         "1h",
	}
	var reply dispatchers.RatingPlanCost
	if err := apierRPC.Call(utils.RALsV1GetRatingPlansCost, arg, &reply); err != nil {
		t.Error(err)
	} else if reply.RatingPlanID != "RP_1002" {
		t.Error("Unexpected RatingPlanID: ", reply.RatingPlanID)
	} else if *reply.EventCost.Cost != 6.5118 {
		t.Error("Unexpected Cost: ", *reply.EventCost.Cost)
	} else if *reply.EventCost.Usage != time.Duration(time.Hour) {
		t.Error("Unexpected Usage: ", *reply.EventCost.Usage)
	}
}

func testAPIerGetActionPlanIDs(t *testing.T) {
	var reply []string
	if err := apierRPC.Call(utils.APIerSv1GetActionPlanIDs,
		utils.TenantArgWithPaginator{TenantArg: utils.TenantArg{Tenant: "cgrates.org"}},
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
	expected := []string{"RP_1002_LOW", "RP_1003", "RP_1001", "RP_SMS", "RP_1002"}
	if err := apierRPC.Call(utils.APIerSv1GetRatingPlanIDs,
		utils.TenantArgWithPaginator{TenantArg: utils.TenantArg{Tenant: "cgrates.org"}},
		&reply); err != nil {
		t.Error(err)
	} else if len(reply) != 5 {
		t.Errorf("Expected: 5 , received: <%+v>", len(reply))
	} else {
		sort.Strings(reply)
		sort.Strings(expected)
		if !reflect.DeepEqual(reply, expected) {
			t.Errorf("Expected: <%+v> , received: <%+v>", expected, reply)

		}
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
				ActivationTime: time.Date(2014, 1, 14, 0, 0, 0,0, time.UTC),
				RatingPlanId:   "RP_SMS",
			},
		},
	}
	attrGetRatingPlan := &utils.AttrGetRatingProfile{
		Tenant: "cgrates.org", Category: "sms", Subject: utils.ANY}
	var rpl engine.RatingProfile
	if err := apierRPC.Call(utils.APIerSv1GetRatingProfile, attrGetRatingPlan, &rpl); err != nil {
		t.Errorf("Got error on APIerSv1.GetRatingProfile: %+v", err)
	} else if !reflect.DeepEqual(expected, rpl) {
		t.Errorf("Calling APIerSv1.GetRatingProfile expected: %+v, received: %+v", utils.ToJSON(expected), utils.ToJSON(rpl))
	}

	if err := apierRPC.Call(utils.APIerSv1RemoveRatingProfile, &AttrRemoveRatingProfile{
		Tenant:   "cgrates.org",
		Category: "sms",
		Subject:  utils.ANY,
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
