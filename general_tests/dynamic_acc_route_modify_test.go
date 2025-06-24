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
	"path/filepath"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

func TestDynamicAccountRouteModify(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

	ng := engine.TestEngine{
		ConfigPath: filepath.Join(*utils.DataDir, "conf", "samples", "dynamic_account_threshold"),
		TpPath:     path.Join(*utils.DataDir, "tariffplans", "testit"),
		// LogBuffer:  &bytes.Buffer{},
	}
	// t.Cleanup(func() { fmt.Println(ng.LogBuffer) })
	client, _ := ng.Run(t)

	t.Run("SetInitiativeThresholdProfile", func(t *testing.T) {
		time.Sleep(10 * time.Millisecond) // wait for tps to be loaded
		thdPrf := &engine.ThresholdProfileWithAPIOpts{
			ThresholdProfile: &engine.ThresholdProfile{
				Tenant: "cgrates.org",
				ID:     "THD_DYNAMIC_STATS_AND_THRESHOLD_INIT",
				FilterIDs: []string{"*string:~*req.Source:Terminate",
					"*exists:~*req.Destination:",
					"*lte:~*req.Usage:10s"},
				ActionIDs: []string{"ACT_DYN_THRESHOLD_AND_STATS_CREATION"},
				MinHits:   1,
				MaxHits:   -1,
				Weight:    1, // keep in mind weight should be lower than the dynamicaly created thresholds so that we dont retrigger this threshold for the same Destination
				Async:     true,
			},
		}
		var reply string
		if err := client.Call(context.Background(), utils.APIerSv1SetThresholdProfile, thdPrf, &reply); err != nil {
			t.Error(err)
		} else if reply != utils.OK {
			t.Error("Unexpected reply returned", reply)
		}
	})

	rPrf := &engine.RouteWithAPIOpts{
		RouteProfile: &engine.RouteProfile{
			Tenant:            "cgrates.org",
			ID:                "ROUTE_PROFILE_1",
			Sorting:           "*weight",
			SortingParameters: []string{utils.MetaACD},
			Routes: []*engine.Route{
				{
					ID:            "Route1",
					RatingPlanIDs: []string{"RPL1"},
					FilterIDs:     []string{"*string:~*req.Destination:1002"},
					Weight:        10,
					Blocker:       false,
				},
				{
					ID:            "Route2",
					RatingPlanIDs: []string{"RPL2"},
					FilterIDs:     []string{"*string:~*req.Destination:1003"},
					Weight:        20,
					Blocker:       false,
				},
			},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 29, 15, 0, 0, 0, time.UTC),
			},
			Weight: 10,
		},
	}
	t.Run("SetStartingRouteProfile", func(t *testing.T) {
		var reply string
		if err := client.Call(context.Background(), utils.APIerSv1SetRouteProfile, rPrf, &reply); err != nil {
			t.Error(err)
		} else if reply != utils.OK {
			t.Error("Unexpected reply returned", reply)
		}
	})

	t.Run("SetAfter3sTiming", func(t *testing.T) {
		timing := &utils.TPTimingWithAPIOpts{
			TPTiming: &utils.TPTiming{
				ID:        "TM_AFTER_3S",
				StartTime: "+3s", // timing which will start the moment the action plan is executed. After the duration in StartTime, the Action from the actionPlan will be executed. Action plans executed this way will be triggered only once right when timer finishes
			},
		}
		var reply string
		if err := client.Call(context.Background(), utils.APIerSv1SetTiming, timing, &reply); err != nil {
			t.Error(err)
		} else if reply != utils.OK {
			t.Error("Unexpected reply returned", reply)
		}
	})

	t.Run("DynamicallyModifyRouteUsingActions", func(t *testing.T) {
		attrs1 := &utils.AttrSetActions{
			ActionsId: "ACT_DYN_MODIFY_ROUTE",
			Actions: []*utils.TPAction{
				{
					Identifier:      utils.MetaDynamicAction,
					ExtraParameters: "ResetDynamicThreshold;*reset_threshold;\fcgrates.org:THD_DST_<~*req.Destination>_ROUTE_MODIFY\f;;;;;;;;;;;;;;10;true",
					Weight:          15, // weight is important
				},
				{
					Identifier:      utils.MetaDynamicActionPlan,
					ExtraParameters: "ExecuteResetThreshold;ResetDynamicThreshold;TM_AFTER_3S;10;true", // reset threshold 3 seconds after THD_DST_<~*req.Destination>_ROUTE_MODIFY triggered so that we keep the threshold disabled for the 3 seconds
					Weight:          10,                                                                // weight is important
				},
				{
					Identifier:      utils.MetaDynamicRoute,
					ExtraParameters: ";~*req.Route;;+3s;;;;;;;;;;;;;",
					Weight:          5, // weight is important
				},
			},
		}
		var reply string
		if err := client.Call(context.Background(), utils.APIerSv2SetActions, &attrs1, &reply); err != nil {
			t.Error(err)
		} else if reply != utils.OK {
			t.Errorf("Unexpected reply returned: %s", reply)
		}
	})

	t.Run("SetDynamicThresholdAndStatsAction", func(t *testing.T) {
		attrs1 := &utils.AttrSetActions{
			ActionsId: "ACT_DYN_THRESHOLD_AND_STATS_CREATION",
			Actions: []*utils.TPAction{
				{
					Identifier:      utils.MetaDynamicThreshold,
					ExtraParameters: "*tenant;THD_BLOCKER_DST_<~*req.Destination>;*string:~*req.Destination:<~*req.Destination>;*now;-1;1;;true;3;;true;;",
				},
				{
					Identifier:      utils.MetaDynamicThreshold,
					ExtraParameters: "*tenant;THD_DST_<~*req.Destination>_ROUTE_MODIFY;*string:~*req.Source:Terminate&*string:~*req.Destination:<~*req.Destination>&*lte:~*req.Usage:10s;*now;-1;5;3s;true;4;ACT_DYN_MODIFY_ROUTE;true;;",
				},
			},
		}
		var reply string
		if err := client.Call(context.Background(), utils.APIerSv2SetActions, &attrs1, &reply); err != nil {
			t.Error(err)
		} else if reply != utils.OK {
			t.Errorf("Unexpected reply returned: %s", reply)
		}
	})

	t.Run("SetActionPlanOfDynaPrepaidAccounts", func(t *testing.T) {
		var reply string
		atms1 := &engine.AttrSetActionPlan{
			Id: "DYNA_ACC",
			ActionPlan: []*engine.AttrActionPlan{
				{
					ActionsId: "TOPUP_RST_MONETARY_10",
					Time:      utils.MetaMonthlyEstimated,
					TimingID:  utils.MetaMonthlyEstimated,
					Weight:    20,
				},
			},
			Overwrite: false,
		}
		if err := client.Call(context.Background(), utils.APIerSv1SetActionPlan, &atms1, &reply); err != nil {
			t.Error("Got error on APIerSv1.SetActionPlan: ", err.Error())
		} else if reply != utils.OK {
			t.Errorf("Unexpected reply returned: %s", reply)
		}
	})

	t.Run("Make5TerminateCalls", func(t *testing.T) {
		args1 := &sessions.V1TerminateSessionArgs{
			TerminateSession:  true,
			ProcessThresholds: true,
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				Event: map[string]any{
					utils.ACD:          7 * time.Second,
					utils.AccountField: "CreatedAccount",
					utils.AnswerTime:   time.Date(2018, 8, 24, 16, 00, 26, 0, time.UTC),
					utils.Category:     "call",
					utils.Cost:         -1,
					utils.Destination:  "1002",
					utils.OriginID:     "sessDynaprepaid",
					utils.OriginHost:   "192.168.1.1",
					utils.RequestType:  utils.MetaDynaprepaid,
					utils.Route:        "Route1@ROUTE_PROFILE_1",
					utils.Source:       "Terminate",
					utils.Tenant:       "cgrates.org",
					utils.ToR:          utils.MetaVoice,
					utils.Usage:        7 * time.Second,
				},
			},
		}
		var rply1 string
		if err := client.Call(context.Background(), utils.SessionSv1TerminateSession,
			args1, &rply1); err != nil {
			t.Error(err)
			return
		} else if rply1 != utils.OK {
			t.Errorf("Unexpected reply: %s", rply1)
		}
		for i := range 5 {
			strI := strconv.Itoa(i)
			args1 := &sessions.V1TerminateSessionArgs{
				TerminateSession:  true,
				ProcessThresholds: true,
				CGREvent: &utils.CGREvent{
					Tenant: "cgrates.org",
					Event: map[string]any{
						utils.ACD:          7 * time.Second,
						utils.AccountField: "CreatedAccount",
						utils.AnswerTime:   time.Date(2018, 8, 24, 16, 00, 26, 0, time.UTC),
						utils.Category:     "call",
						utils.Cost:         -1,
						utils.Destination:  "1002",
						utils.OriginID:     "sessPrepaid" + strI,
						utils.OriginHost:   "192.168.1.1",
						utils.RequestType:  utils.MetaDynaprepaid,
						utils.Route:        "Route1@ROUTE_PROFILE_1",
						utils.Source:       "Terminate",
						utils.Tenant:       "cgrates.org",
						utils.ToR:          utils.MetaVoice,
						utils.Usage:        7 * time.Second,
					},
				},
			}
			var rply1 string
			if err := client.Call(context.Background(), utils.SessionSv1TerminateSession,
				args1, &rply1); err != nil {
				t.Error(err)
				return
			} else if rply1 != utils.OK {
				t.Errorf("Unexpected reply: %s", rply1)
			}
		}
	})

	t.Run("CheckModifiedRoute", func(t *testing.T) {
		// wait for route to be modified async
		time.Sleep(10 * time.Millisecond)
		routeProfileFound := new(engine.RouteProfile)
		if err := client.Call(context.Background(), utils.APIerSv1GetRouteProfile,
			&utils.TenantID{Tenant: "cgrates.org", ID: "ROUTE_PROFILE_1"}, &routeProfileFound); err != nil {
			t.Error(err)
		}
		if routeProfileFound.ActivationInterval.ActivationTime.Equal(time.Date(2014, 7, 29, 15, 0, 0, 0, time.UTC)) {
			t.Fatalf("Activation time didnt change, received <%v>", routeProfileFound.ActivationInterval.ActivationTime)
		}
		rPrf.ActivationInterval = routeProfileFound.ActivationInterval
		if !reflect.DeepEqual(utils.ToJSON(rPrf.RouteProfile), utils.ToJSON(routeProfileFound)) {
			t.Errorf("Expected <%v>, \nreceived <%v>", utils.ToJSON(rPrf.RouteProfile), utils.ToJSON(routeProfileFound))
		}
	})
	t.Run("Make5TerminateCalls2", func(t *testing.T) {
		time.Sleep(4 * time.Second) // wait for threshold to be reset
		args1 := &sessions.V1TerminateSessionArgs{
			TerminateSession:  true,
			ProcessThresholds: true,
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				Event: map[string]any{
					utils.ACD:          7 * time.Second,
					utils.AccountField: "CreatedAccount",
					utils.AnswerTime:   time.Date(2018, 8, 24, 16, 00, 26, 0, time.UTC),
					utils.Category:     "call",
					utils.Cost:         -1,
					utils.Destination:  "1002",
					utils.OriginID:     "sessDynaprepaid",
					utils.OriginHost:   "192.168.1.1",
					utils.RequestType:  utils.MetaDynaprepaid,
					utils.Route:        "Route1@ROUTE_PROFILE_1",
					utils.Source:       "Terminate",
					utils.Tenant:       "cgrates.org",
					utils.ToR:          utils.MetaVoice,
					utils.Usage:        7 * time.Second,
				},
			},
		}
		var rply1 string
		if err := client.Call(context.Background(), utils.SessionSv1TerminateSession,
			args1, &rply1); err != nil {
			t.Error(err)
			return
		} else if rply1 != utils.OK {
			t.Errorf("Unexpected reply: %s", rply1)
		}
		for i := range 5 {
			strI := strconv.Itoa(i + 5)
			args1 := &sessions.V1TerminateSessionArgs{
				TerminateSession:  true,
				ProcessThresholds: true,
				CGREvent: &utils.CGREvent{
					Tenant: "cgrates.org",
					Event: map[string]any{
						utils.ACD:          7 * time.Second,
						utils.AccountField: "CreatedAccount",
						utils.AnswerTime:   time.Date(2018, 8, 24, 16, 00, 26, 0, time.UTC),
						utils.Category:     "call",
						utils.Cost:         -1,
						utils.Destination:  "1002",
						utils.OriginID:     "sessPrepaid" + strI,
						utils.OriginHost:   "192.168.1.1",
						utils.RequestType:  utils.MetaDynaprepaid,
						utils.Route:        "Route1@ROUTE_PROFILE_1",
						utils.Source:       "Terminate",
						utils.Tenant:       "cgrates.org",
						utils.ToR:          utils.MetaVoice,
						utils.Usage:        7 * time.Second,
					},
				},
			}
			var rply1 string
			if err := client.Call(context.Background(), utils.SessionSv1TerminateSession,
				args1, &rply1); err != nil {
				t.Error(err)
				return
			} else if rply1 != utils.OK {
				t.Errorf("Unexpected reply: %s", rply1)
			}
		}
	})

	t.Run("CheckModifiedRoute2", func(t *testing.T) {
		// wait for route to be modified async
		time.Sleep(10 * time.Millisecond)
		routeProfileFound := new(engine.RouteProfile)
		if err := client.Call(context.Background(), utils.APIerSv1GetRouteProfile,
			&utils.TenantID{Tenant: "cgrates.org", ID: "ROUTE_PROFILE_1"}, &routeProfileFound); err != nil {
			t.Error(err)
		}
		if routeProfileFound.ActivationInterval.ActivationTime.Equal(rPrf.ActivationInterval.ActivationTime) {
			t.Fatalf("Activation time didnt change, received <%v>", routeProfileFound.ActivationInterval.ActivationTime)
		}
		rPrf.ActivationInterval = routeProfileFound.ActivationInterval
		if !reflect.DeepEqual(utils.ToJSON(rPrf.RouteProfile), utils.ToJSON(routeProfileFound)) {
			t.Errorf("Expected <%v>, \nreceived <%v>", utils.ToJSON(rPrf.RouteProfile), utils.ToJSON(routeProfileFound))
		}
	})
}
