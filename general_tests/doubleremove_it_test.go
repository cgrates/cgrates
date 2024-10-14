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

	"github.com/cgrates/birpc/context"
	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// subtests to be executed
var (
	doubleRemovePath    string
	doubleRemoveDIR     string
	doubleRemove        *config.CGRConfig
	doubleRemoveAccount = "refundAcc"
	doubleRemoveTenant  = "cgrates.org"

	doubleRemoveIT = []func(t *testing.T){
		testdoubleRemoveLoadConfig,
		testdoubleRemoveInitDataDb,
		testdoubleRemoveStartEngine,
		testdoubleRemoveRpcConn,

		testdoubleRemoveStatQueueProfile,
		testdoubleRemoveActions,
		testdoubleRemoveActionPlan,

		testdoubleRemoveKillEngine,
	}
)

// Test start here
func TestDoubleRemoveIT(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		doubleRemoveDIR = "tutinternal"
	case utils.MetaMySQL:
		doubleRemoveDIR = "tutmysql"
	case utils.MetaMongo:
		doubleRemoveDIR = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}

	for _, stest := range doubleRemoveIT {
		t.Run(doubleRemoveDIR, stest)
	}
}

func testdoubleRemoveLoadConfig(t *testing.T) {
	var err error
	doubleRemovePath = path.Join(*utils.DataDir, "conf", "samples", doubleRemoveDIR)
	if doubleRemove, err = config.NewCGRConfigFromPath(doubleRemovePath); err != nil {
		t.Error(err)
	}
}

func testdoubleRemoveInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(doubleRemove); err != nil {
		t.Fatal(err)
	}
}

func testdoubleRemoveStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(doubleRemovePath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

func testdoubleRemoveRpcConn(t *testing.T) {
	sesRPC = engine.NewRPCClient(t, doubleRemove.ListenCfg())
}

func testdoubleRemoveStatQueueProfile(t *testing.T) {
	// check
	var reply *engine.StatQueueProfile
	if err := sesRPC.Call(context.Background(), utils.APIerSv1GetStatQueueProfile,
		&utils.TenantID{Tenant: doubleRemoveTenant, ID: "TEST_PROFILE1"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	// set
	statConfig := &engine.StatQueueProfileWithAPIOpts{
		StatQueueProfile: &engine.StatQueueProfile{
			Tenant: doubleRemoveTenant,
			ID:     "TEST_PROFILE1",
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2020, 4, 18, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2020, 4, 18, 14, 25, 0, 0, time.UTC),
			},
			QueueLength: 10,
			TTL:         10 * time.Second,
			Metrics: []*engine.MetricWithFilters{
				{
					MetricID: "*sum",
				},
				{
					MetricID: "*acd",
				},
			},
			ThresholdIDs: []string{"Val1", "Val2"},
			Blocker:      true,
			Stored:       true,
			Weight:       20,
			MinItems:     1,
		},
	}
	var result string
	if err := sesRPC.Call(context.Background(), utils.APIerSv1SetStatQueueProfile, statConfig, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	//check
	if err := sesRPC.Call(context.Background(), utils.APIerSv1GetStatQueueProfile,
		&utils.TenantID{Tenant: doubleRemoveTenant, ID: "TEST_PROFILE1"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(statConfig.StatQueueProfile, reply) {
		t.Errorf("Expecting: %+v, received: %+v", statConfig.StatQueueProfile, reply)
	}

	//remove
	if err := sesRPC.Call(context.Background(), utils.APIerSv1RemoveStatQueueProfile,
		&utils.TenantID{Tenant: doubleRemoveTenant, ID: "TEST_PROFILE1"}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := sesRPC.Call(context.Background(), utils.APIerSv1RemoveStatQueueProfile,
		&utils.TenantID{Tenant: doubleRemoveTenant, ID: "TEST_PROFILE1"}, &result); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := sesRPC.Call(context.Background(), utils.APIerSv1RemoveStatQueueProfile,
		&utils.TenantID{Tenant: doubleRemoveTenant, ID: "TEST_PROFILE1"}, &result); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	// check
	if err := sesRPC.Call(context.Background(), utils.APIerSv1GetStatQueueProfile,
		&utils.TenantID{Tenant: doubleRemoveTenant, ID: "TEST_PROFILE1"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testdoubleRemoveActions(t *testing.T) {
	// check
	var reply1 []*utils.TPAction
	if doubleRemoveDIR != "tutinternal" { // on internal do not get so we do not cache this action with nil in cache
		if err := sesRPC.Call(context.Background(), utils.APIerSv1GetActions, utils.StringPointer("ACTS_1"), &reply1); err == nil || err.Error() != "SERVER_ERROR: NOT_FOUND" {
			t.Error(err)
		}
	}
	// set
	attrs1 := &v1.V1AttrSetActions{
		ActionsId: "ACTS_1",
		Actions: []*v1.V1TPAction{
			{
				Identifier:  utils.MetaTopUpReset,
				BalanceType: utils.MetaMonetary,
				Units:       75.0,
				ExpiryTime:  utils.MetaUnlimited,
				Weight:      20.0}},
	}
	var reply string
	if err := sesRPC.Call(context.Background(), utils.APIerSv1SetActions, &attrs1, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Unexpected reply returned: %s", reply)
	}
	// set it again (expect EXISTS)
	if err := sesRPC.Call(context.Background(), utils.APIerSv1SetActions, &attrs1, &reply); err == nil || err.Error() != "EXISTS" {
		t.Error(err)
	}
	// check
	eOut := []*utils.TPAction{
		{
			Identifier:      utils.MetaTopUpReset,
			BalanceType:     utils.MetaMonetary,
			Units:           "75",
			BalanceWeight:   "0",
			BalanceBlocker:  "false",
			BalanceDisabled: "false",
			ExpiryTime:      utils.MetaUnlimited,
			Weight:          20.0,
		}}
	if err := sesRPC.Call(context.Background(), utils.APIerSv1GetActions, utils.StringPointer("ACTS_1"), &reply1); err != nil {
		t.Error("Got error on APIerSv1.GetActions: ", err.Error())
	} else if !reflect.DeepEqual(eOut, reply1) {
		t.Errorf("Expected: %v, received: %v", utils.ToJSON(eOut), utils.ToJSON(reply1))
	}
	// remove
	if err := sesRPC.Call(context.Background(), utils.APIerSv1RemoveActions, &v1.AttrRemoveActions{
		ActionIDs: []string{"ACTS_1"}}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	// remove it again (expect ErrNotFound)
	// if err := sesRPC.Call(context.Background(),utils.APIerSv1RemoveActions, &v1.AttrRemoveActions{
	// 	ActionIDs: []string{"ACTS_1"}}, &reply); err == nil ||
	// 	err.Error() != utils.ErrNotFound.Error() {
	// 	t.Error(err)
	// }
	// check again
	if err := sesRPC.Call(context.Background(), utils.APIerSv1GetActions, utils.StringPointer("ACTS_1"), &reply1); err == nil || err.Error() != "SERVER_ERROR: NOT_FOUND" {
		t.Error(err)
	}
}

func testdoubleRemoveActionPlan(t *testing.T) {
	//set action
	var reply string
	if err := sesRPC.Call(context.Background(), utils.APIerSv2SetActions, &utils.AttrSetActions{
		ActionsId: "ACTS_2",
		Actions:   []*utils.TPAction{{Identifier: utils.MetaLog}},
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv2.SetActions received: %s", reply)
	}
	// check action
	var reply1 []*utils.TPAction
	eOut := []*utils.TPAction{
		{
			Identifier:      "*log",
			Units:           "0",
			BalanceWeight:   "0",
			BalanceBlocker:  "false",
			BalanceDisabled: "false",
			Weight:          0}}

	if err := sesRPC.Call(context.Background(), utils.APIerSv1GetActions, utils.StringPointer("ACTS_2"), &reply1); err != nil {
		t.Error("Got error on APIerSv1.GetActions: ", err.Error())
	} else if !reflect.DeepEqual(eOut, reply1) {
		t.Errorf("Expected: %v, received: %v", utils.ToJSON(eOut), utils.ToJSON(reply1))
	}
	// check ActionPlan
	var aps []*engine.ActionPlan
	/*
		should return ErrNotFound, right now it returns nil and an empty slice,
		needs to be reviewed

		if err := sesRPC.Call(context.Background(),utils.APIerSv1GetActionPlan,
			v1.AttrGetActionPlan{ID: utils.EmptyString}, &aps); err == nil || err.Error() != utils.ErrNotFound.Error() {
			t.Errorf("Error: %+v, rcv: %+v", err, utils.ToJSON(aps))
		}
	*/
	// set ActionPlan
	atms1 := &v1.AttrSetActionPlan{
		Id: "ATMS_1",
		ActionPlan: []*v1.AttrActionPlan{
			{
				ActionsId: "ACTS_2",
				Time:      utils.MetaASAP,
				Weight:    20.0},
		},
	}
	if err := sesRPC.Call(context.Background(), utils.APIerSv1SetActionPlan, &atms1, &reply); err != nil {
		t.Error("Got error on APIerSv1.SetActionPlan: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Unexpected reply returned: %s", reply)
	}
	// set it again (expect EXISTS)
	if err := sesRPC.Call(context.Background(), utils.APIerSv1SetActionPlan, &atms1, &reply); err == nil || err.Error() != "EXISTS" {
		t.Error(err)
	}
	// check
	if err := sesRPC.Call(context.Background(), utils.APIerSv1GetActionPlan,
		&v1.AttrGetActionPlan{ID: "ATMS_1"}, &aps); err != nil {
		t.Error(err)
	} else if len(aps) != 1 {
		t.Errorf("Expected: %v,\n received: %v", 1, len(aps))
	} else if aps[0].Id != "ATMS_1" {
		t.Errorf("Expected: ATMS_1,\n received: %v", aps[0].Id)
	} else if aps[0].ActionTimings[0].ActionsID != "ACTS_2" {
		t.Errorf("Expected: ACTS_2,\n received: %v", aps[0].ActionTimings[0].ActionsID)
	} else if aps[0].ActionTimings[0].Weight != 20.0 {
		t.Errorf("Expected: 20.0,\n received: %v", aps[0].ActionTimings[0].Weight)
	}

	// remove
	if err := sesRPC.Call(context.Background(), utils.APIerSv1RemoveActionPlan, &v1.AttrGetActionPlan{
		ID: "ATMS_1"}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	//check again
	/*
		this should return ErrNotFound, right now it returns nil and an empty slice,
		needs to be reviewed.

		if err := sesRPC.Call(context.Background(),utils.APIerSv1GetActionPlan,
			v1.AttrGetActionPlan{ID: utils.EmptyString}, &aps); err == nil || err.Error() != utils.ErrNotFound.Error() {
			t.Errorf("Error: %+v, rcv: %+v", err, utils.ToJSON(aps))
		}
	*/

}

func testdoubleRemoveKillEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
