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

// subtests to be executed
var (
	doubleRemovePath   string
	doubleRemoveDIR    string
	doubleRemove       *config.CGRConfig
	doubleRemoveTenant = "cgrates.org"
	doubleRemoveRPC    *birpc.Client

	doubleRemoveIT = []func(t *testing.T){
		testdoubleRemoveLoadConfig,
		testdoubleRemoveInitDataDb,
		testdoubleRemoveStartEngine,
		testdoubleRemoveRpcConn,

		testdoubleRemoveStatQueueProfile,
		testdoubleRemoveThresholdProfile,
		testdoubleRemoveRouteProfile,
		testdoubleRemoveAttributeProfile,
		testdoubleRemoveChargerProfile,
		testdoubleRemoveResourceProfile,
		testdoubleRemoveDispatcherProfile,
		testdoubleRemoveDispatcherHost,
		testdoubleRemoveRateProfile,
		testdoubleRemoveActionProfile,

		testdoubleRemoveKillEngine,
	}
)

//Test starts here
func TestDoubleRemoveIT(t *testing.T) {
	switch *dbType {
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
	doubleRemovePath = path.Join(*dataDir, "conf", "samples", doubleRemoveDIR)
	if doubleRemove, err = config.NewCGRConfigFromPath(context.Background(), doubleRemovePath); err != nil {
		t.Error(err)
	}
}

func testdoubleRemoveInitDataDb(t *testing.T) {
	if err := engine.InitDataDB(doubleRemove); err != nil {
		t.Fatal(err)
	}
}

func testdoubleRemoveStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(doubleRemovePath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testdoubleRemoveRpcConn(t *testing.T) {
	var err error
	doubleRemoveRPC, err = newRPCClient(doubleRemove.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

func testdoubleRemoveStatQueueProfile(t *testing.T) {
	// check
	var reply *engine.StatQueueProfile
	if err := doubleRemoveRPC.Call(context.Background(), utils.AdminSv1GetStatQueueProfile,
		&utils.TenantID{Tenant: doubleRemoveTenant, ID: "TEST_PROFILE1"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	// set
	statConfig := &engine.StatQueueProfileWithAPIOpts{
		StatQueueProfile: &engine.StatQueueProfile{
			Tenant:      doubleRemoveTenant,
			ID:          "TEST_PROFILE1",
			FilterIDs:   []string{"*ai:~*opts.*startTime:2020-04-18T14:25:00Z|2020-04-18T14:25:00Z"},
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
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
			MinItems: 1,
		},
	}
	var result string
	if err := doubleRemoveRPC.Call(context.Background(), utils.AdminSv1SetStatQueueProfile, statConfig, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	//check
	if err := doubleRemoveRPC.Call(context.Background(), utils.AdminSv1GetStatQueueProfile,
		&utils.TenantID{Tenant: doubleRemoveTenant, ID: "TEST_PROFILE1"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(statConfig.StatQueueProfile, reply) {
		t.Errorf("Expecting: %+v, received: %+v", statConfig.StatQueueProfile, reply)
	}

	//remove
	if err := doubleRemoveRPC.Call(context.Background(), utils.AdminSv1RemoveStatQueueProfile,
		&utils.TenantID{Tenant: doubleRemoveTenant, ID: "TEST_PROFILE1"}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := doubleRemoveRPC.Call(context.Background(), utils.AdminSv1RemoveStatQueueProfile,
		&utils.TenantID{Tenant: doubleRemoveTenant, ID: "TEST_PROFILE1"}, &result); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := doubleRemoveRPC.Call(context.Background(), utils.AdminSv1RemoveStatQueueProfile,
		&utils.TenantID{Tenant: doubleRemoveTenant, ID: "TEST_PROFILE1"}, &result); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	// check
	if err := doubleRemoveRPC.Call(context.Background(), utils.AdminSv1GetStatQueueProfile,
		&utils.TenantID{Tenant: doubleRemoveTenant, ID: "TEST_PROFILE1"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testdoubleRemoveThresholdProfile(t *testing.T) {
	// check
	var reply *engine.ThresholdProfile
	if err := doubleRemoveRPC.Call(context.Background(), utils.AdminSv1GetThresholdProfile,
		&utils.TenantID{Tenant: doubleRemoveTenant, ID: "TH_PROFILE"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	// set
	thPrf := &engine.ThresholdProfileWithAPIOpts{
		ThresholdProfile: &engine.ThresholdProfile{
			Tenant:           doubleRemoveTenant,
			ID:               "TH_PROFILE",
			FilterIDs:        []string{"*string:~*req.Account:1001"},
			Blocker:          true,
			MaxHits:          5,
			MinHits:          3,
			ActionProfileIDs: []string{utils.MetaNone},
			Async:            true,
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
		},
	}
	var result string
	if err := doubleRemoveRPC.Call(context.Background(), utils.AdminSv1SetThresholdProfile, thPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	//check
	if err := doubleRemoveRPC.Call(context.Background(), utils.AdminSv1GetThresholdProfile,
		&utils.TenantID{Tenant: doubleRemoveTenant, ID: "TH_PROFILE"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(thPrf.ThresholdProfile, reply) {
		t.Errorf("Expecting: %+v, received: %+v", thPrf.ThresholdProfile, reply)
	}

	//remove
	if err := doubleRemoveRPC.Call(context.Background(), utils.AdminSv1RemoveThresholdProfile,
		&utils.TenantID{Tenant: doubleRemoveTenant, ID: "TH_PROFILE"}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := doubleRemoveRPC.Call(context.Background(), utils.AdminSv1RemoveThresholdProfile,
		&utils.TenantID{Tenant: doubleRemoveTenant, ID: "TH_PROFILE"}, &result); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := doubleRemoveRPC.Call(context.Background(), utils.AdminSv1RemoveThresholdProfile,
		&utils.TenantID{Tenant: doubleRemoveTenant, ID: "TH_PROFILE"}, &result); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	// check
	if err := doubleRemoveRPC.Call(context.Background(), utils.AdminSv1GetThresholdProfile,
		&utils.TenantID{Tenant: doubleRemoveTenant, ID: "TH_PROFILE"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testdoubleRemoveRouteProfile(t *testing.T) {
	// check
	var reply *engine.RouteProfile
	if err := doubleRemoveRPC.Call(context.Background(), utils.AdminSv1GetRouteProfile,
		&utils.TenantID{Tenant: doubleRemoveTenant, ID: "ROUTE_PROFILE"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	// set
	routePrf := &engine.RouteProfileWithAPIOpts{
		RouteProfile: &engine.RouteProfile{
			Tenant:    doubleRemoveTenant,
			ID:        "ROUTE_PROFILE",
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Sorting:   utils.MetaWeight,
			Routes: []*engine.Route{
				{
					ID: "ROUTE",
				},
			},
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
		},
	}
	var result string
	if err := doubleRemoveRPC.Call(context.Background(), utils.AdminSv1SetRouteProfile, routePrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	//check
	if err := doubleRemoveRPC.Call(context.Background(), utils.AdminSv1GetRouteProfile,
		&utils.TenantID{Tenant: doubleRemoveTenant, ID: "ROUTE_PROFILE"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(routePrf.RouteProfile, reply) {
		t.Errorf("Expecting: %+v, received: %+v", routePrf.RouteProfile, reply)
	}

	//remove
	if err := doubleRemoveRPC.Call(context.Background(), utils.AdminSv1RemoveRouteProfile,
		&utils.TenantID{Tenant: doubleRemoveTenant, ID: "ROUTE_PROFILE"}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := doubleRemoveRPC.Call(context.Background(), utils.AdminSv1RemoveRouteProfile,
		&utils.TenantID{Tenant: doubleRemoveTenant, ID: "ROUTE_PROFILE"}, &result); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := doubleRemoveRPC.Call(context.Background(), utils.AdminSv1RemoveRouteProfile,
		&utils.TenantID{Tenant: doubleRemoveTenant, ID: "ROUTE_PROFILE"}, &result); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	// check
	if err := doubleRemoveRPC.Call(context.Background(), utils.AdminSv1GetRouteProfile,
		&utils.TenantID{Tenant: doubleRemoveTenant, ID: "ROUTE_PROFILE"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testdoubleRemoveAttributeProfile(t *testing.T) {
	// check
	var reply *engine.APIAttributeProfile
	if err := doubleRemoveRPC.Call(context.Background(), utils.AdminSv1GetAttributeProfile,
		&utils.TenantID{Tenant: doubleRemoveTenant, ID: "ATTR_PROFILE"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	// set
	attrPrf := &engine.APIAttributeProfileWithAPIOpts{
		APIAttributeProfile: &engine.APIAttributeProfile{
			Tenant:    doubleRemoveTenant,
			ID:        "ATTR_PROFILE",
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Blocker:   false,
			Attributes: []*engine.ExternalAttribute{
				{
					Path:  "*req.Destination",
					Type:  utils.MetaConstant,
					Value: "12018209998",
				},
			},
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
		},
	}
	var result string
	if err := doubleRemoveRPC.Call(context.Background(), utils.AdminSv1SetAttributeProfile, attrPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	//check
	if err := doubleRemoveRPC.Call(context.Background(), utils.AdminSv1GetAttributeProfile,
		&utils.TenantID{Tenant: doubleRemoveTenant, ID: "ATTR_PROFILE"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(attrPrf.APIAttributeProfile, reply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(attrPrf.APIAttributeProfile), utils.ToJSON(reply))
	}

	//remove
	if err := doubleRemoveRPC.Call(context.Background(), utils.AdminSv1RemoveAttributeProfile,
		&utils.TenantID{Tenant: doubleRemoveTenant, ID: "ATTR_PROFILE"}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := doubleRemoveRPC.Call(context.Background(), utils.AdminSv1RemoveAttributeProfile,
		&utils.TenantID{Tenant: doubleRemoveTenant, ID: "ATTR_PROFILE"}, &result); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := doubleRemoveRPC.Call(context.Background(), utils.AdminSv1RemoveAttributeProfile,
		&utils.TenantID{Tenant: doubleRemoveTenant, ID: "ATTR_PROFILE"}, &result); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	// check
	if err := doubleRemoveRPC.Call(context.Background(), utils.AdminSv1GetAttributeProfile,
		&utils.TenantID{Tenant: doubleRemoveTenant, ID: "ATTR_PROFILE"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testdoubleRemoveChargerProfile(t *testing.T) {
	// check
	var reply *engine.ChargerProfile
	if err := doubleRemoveRPC.Call(context.Background(), utils.AdminSv1GetChargerProfile,
		&utils.TenantID{Tenant: doubleRemoveTenant, ID: "CHARGER_PROFILE"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	// set
	chrgPrf := &engine.ChargerProfileWithAPIOpts{
		ChargerProfile: &engine.ChargerProfile{
			Tenant:    doubleRemoveTenant,
			ID:        "CHARGER_PROFILE",
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
		},
	}
	var result string
	if err := doubleRemoveRPC.Call(context.Background(), utils.AdminSv1SetChargerProfile, chrgPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	//check
	if err := doubleRemoveRPC.Call(context.Background(), utils.AdminSv1GetChargerProfile,
		&utils.TenantID{Tenant: doubleRemoveTenant, ID: "CHARGER_PROFILE"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(chrgPrf.ChargerProfile, reply) {
		t.Errorf("Expecting: %+v, received: %+v", chrgPrf.ChargerProfile, reply)
	}

	//remove
	if err := doubleRemoveRPC.Call(context.Background(), utils.AdminSv1RemoveChargerProfile,
		&utils.TenantID{Tenant: doubleRemoveTenant, ID: "CHARGER_PROFILE"}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := doubleRemoveRPC.Call(context.Background(), utils.AdminSv1RemoveChargerProfile,
		&utils.TenantID{Tenant: doubleRemoveTenant, ID: "CHARGER_PROFILE"}, &result); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := doubleRemoveRPC.Call(context.Background(), utils.AdminSv1RemoveChargerProfile,
		&utils.TenantID{Tenant: doubleRemoveTenant, ID: "CHARGER_PROFILE"}, &result); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	// check
	if err := doubleRemoveRPC.Call(context.Background(), utils.AdminSv1GetChargerProfile,
		&utils.TenantID{Tenant: doubleRemoveTenant, ID: "CHARGER_PROFILE"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testdoubleRemoveResourceProfile(t *testing.T) {
	// check
	var reply *engine.ResourceProfile
	if err := doubleRemoveRPC.Call(context.Background(), utils.AdminSv1GetResourceProfile,
		&utils.TenantID{Tenant: doubleRemoveTenant, ID: "RESOURCE_PROFILE"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	// set
	resPrf := &engine.ResourceProfileWithAPIOpts{
		ResourceProfile: &engine.ResourceProfile{
			Tenant:    doubleRemoveTenant,
			ID:        "RESOURCE_PROFILE",
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Blocker:   true,
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
		},
	}
	var result string
	if err := doubleRemoveRPC.Call(context.Background(), utils.AdminSv1SetResourceProfile, resPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	//check
	if err := doubleRemoveRPC.Call(context.Background(), utils.AdminSv1GetResourceProfile,
		&utils.TenantID{Tenant: doubleRemoveTenant, ID: "RESOURCE_PROFILE"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(resPrf.ResourceProfile, reply) {
		t.Errorf("Expecting: %+v, received: %+v", resPrf.ResourceProfile, reply)
	}

	//remove
	if err := doubleRemoveRPC.Call(context.Background(), utils.AdminSv1RemoveResourceProfile,
		&utils.TenantID{Tenant: doubleRemoveTenant, ID: "RESOURCE_PROFILE"}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := doubleRemoveRPC.Call(context.Background(), utils.AdminSv1RemoveResourceProfile,
		&utils.TenantID{Tenant: doubleRemoveTenant, ID: "RESOURCE_PROFILE"}, &result); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := doubleRemoveRPC.Call(context.Background(), utils.AdminSv1RemoveResourceProfile,
		&utils.TenantID{Tenant: doubleRemoveTenant, ID: "RESOURCE_PROFILE"}, &result); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	// check
	if err := doubleRemoveRPC.Call(context.Background(), utils.AdminSv1GetResourceProfile,
		&utils.TenantID{Tenant: doubleRemoveTenant, ID: "RESOURCE_PROFILE"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testdoubleRemoveDispatcherProfile(t *testing.T) {
	// check
	var reply *engine.DispatcherProfile
	if err := doubleRemoveRPC.Call(context.Background(), utils.AdminSv1GetDispatcherProfile,
		&utils.TenantID{Tenant: doubleRemoveTenant, ID: "DSP_PRF"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	// set
	dspPrf := &engine.DispatcherProfileWithAPIOpts{
		DispatcherProfile: &engine.DispatcherProfile{
			Tenant:    doubleRemoveTenant,
			ID:        "DSP_PRF",
			FilterIDs: []string{"*string:~*req.Account:1001"},
		},
	}
	var result string
	if err := doubleRemoveRPC.Call(context.Background(), utils.AdminSv1SetDispatcherProfile, dspPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	//check
	if err := doubleRemoveRPC.Call(context.Background(), utils.AdminSv1GetDispatcherProfile,
		&utils.TenantID{Tenant: doubleRemoveTenant, ID: "DSP_PRF"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(dspPrf.DispatcherProfile, reply) {
		t.Errorf("Expecting: %+v, received: %+v", dspPrf.DispatcherProfile, reply)
	}

	//remove
	if err := doubleRemoveRPC.Call(context.Background(), utils.AdminSv1RemoveDispatcherProfile,
		&utils.TenantID{Tenant: doubleRemoveTenant, ID: "DSP_PRF"}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := doubleRemoveRPC.Call(context.Background(), utils.AdminSv1RemoveDispatcherProfile,
		&utils.TenantID{Tenant: doubleRemoveTenant, ID: "DSP_PRF"}, &result); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := doubleRemoveRPC.Call(context.Background(), utils.AdminSv1RemoveDispatcherProfile,
		&utils.TenantID{Tenant: doubleRemoveTenant, ID: "DSP_PRF"}, &result); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	// check
	if err := doubleRemoveRPC.Call(context.Background(), utils.AdminSv1GetDispatcherProfile,
		&utils.TenantID{Tenant: doubleRemoveTenant, ID: "DSP_PRF"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testdoubleRemoveDispatcherHost(t *testing.T) {
	// check
	var reply *engine.DispatcherHost
	if err := doubleRemoveRPC.Call(context.Background(), utils.AdminSv1GetDispatcherHost,
		&utils.TenantID{Tenant: doubleRemoveTenant, ID: "DSP_HOST"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	// set
	dspHost := &engine.DispatcherHostWithAPIOpts{
		DispatcherHost: &engine.DispatcherHost{
			Tenant: doubleRemoveTenant,
			RemoteHost: &config.RemoteHost{
				ID: "DSP_HOST",
			},
		},
	}
	var result string
	if err := doubleRemoveRPC.Call(context.Background(), utils.AdminSv1SetDispatcherHost, dspHost, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	//check
	if err := doubleRemoveRPC.Call(context.Background(), utils.AdminSv1GetDispatcherHost,
		&utils.TenantID{Tenant: doubleRemoveTenant, ID: "DSP_HOST"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(dspHost.DispatcherHost, reply) {
		t.Errorf("Expecting: %+v, received: %+v", dspHost.DispatcherHost, reply)
	}

	//remove
	if err := doubleRemoveRPC.Call(context.Background(), utils.AdminSv1RemoveDispatcherHost,
		&utils.TenantID{Tenant: doubleRemoveTenant, ID: "DSP_HOST"}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := doubleRemoveRPC.Call(context.Background(), utils.AdminSv1RemoveDispatcherHost,
		&utils.TenantID{Tenant: doubleRemoveTenant, ID: "DSP_HOST"}, &result); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := doubleRemoveRPC.Call(context.Background(), utils.AdminSv1RemoveDispatcherHost,
		&utils.TenantID{Tenant: doubleRemoveTenant, ID: "DSP_HOST"}, &result); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	// check
	if err := doubleRemoveRPC.Call(context.Background(), utils.AdminSv1GetDispatcherHost,
		&utils.TenantID{Tenant: doubleRemoveTenant, ID: "DSP_HOST"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testdoubleRemoveRateProfile(t *testing.T) {
	// check
	var reply *utils.RateProfile
	if err := doubleRemoveRPC.Call(context.Background(), utils.AdminSv1GetRateProfile,
		&utils.TenantID{Tenant: doubleRemoveTenant, ID: "RATE_PROFILE"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	// set
	thPrf := &utils.RateProfileWithAPIOpts{
		RateProfile: &utils.RateProfile{
			Tenant:    doubleRemoveTenant,
			ID:        "RATE_PROFILE",
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
			Rates: map[string]*utils.Rate{
				"RATE1": {
					ID: "RATE1",
					Weights: utils.DynamicWeights{
						{
							Weight: 0,
						},
					},
					ActivationTimes: "* * * * *",
					IntervalRates: []*utils.IntervalRate{
						{
							IntervalStart: utils.NewDecimal(0, 0),
							RecurrentFee:  utils.NewDecimal(2, 1),
						},
					},
				},
			},
		},
	}
	var result string
	if err := doubleRemoveRPC.Call(context.Background(), utils.AdminSv1SetRateProfile, thPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	//check
	if err := doubleRemoveRPC.Call(context.Background(), utils.AdminSv1GetRateProfile,
		&utils.TenantID{Tenant: doubleRemoveTenant, ID: "RATE_PROFILE"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(thPrf.RateProfile, reply) {
		t.Errorf("Expecting: %+v, received: %+v", thPrf.RateProfile, reply)
	}

	//remove
	if err := doubleRemoveRPC.Call(context.Background(), utils.AdminSv1RemoveRateProfile,
		&utils.TenantID{Tenant: doubleRemoveTenant, ID: "RATE_PROFILE"}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := doubleRemoveRPC.Call(context.Background(), utils.AdminSv1RemoveRateProfile,
		&utils.TenantID{Tenant: doubleRemoveTenant, ID: "RATE_PROFILE"}, &result); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := doubleRemoveRPC.Call(context.Background(), utils.AdminSv1RemoveRateProfile,
		&utils.TenantID{Tenant: doubleRemoveTenant, ID: "RATE_PROFILE"}, &result); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	// check
	if err := doubleRemoveRPC.Call(context.Background(), utils.AdminSv1GetRateProfile,
		&utils.TenantID{Tenant: doubleRemoveTenant, ID: "RATE_PROFILE"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testdoubleRemoveActionProfile(t *testing.T) {
	// check
	var reply *engine.ActionProfile
	if err := doubleRemoveRPC.Call(context.Background(), utils.AdminSv1GetActionProfile,
		&utils.TenantID{Tenant: doubleRemoveTenant, ID: "ACTION_PROFILE"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	// set
	thPrf := &engine.ActionProfileWithAPIOpts{
		ActionProfile: &engine.ActionProfile{
			Tenant:    doubleRemoveTenant,
			ID:        "ACTION_PROFILE",
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
			Actions: []*engine.APAction{
				{
					ID: "ACTION",
				},
			},
		},
	}
	var result string
	if err := doubleRemoveRPC.Call(context.Background(), utils.AdminSv1SetActionProfile, thPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	//check
	if err := doubleRemoveRPC.Call(context.Background(), utils.AdminSv1GetActionProfile,
		&utils.TenantID{Tenant: doubleRemoveTenant, ID: "ACTION_PROFILE"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(thPrf.ActionProfile, reply) {
		t.Errorf("Expecting: %+v, received: %+v", thPrf.ActionProfile, reply)
	}

	//remove
	if err := doubleRemoveRPC.Call(context.Background(), utils.AdminSv1RemoveActionProfile,
		&utils.TenantID{Tenant: doubleRemoveTenant, ID: "ACTION_PROFILE"}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := doubleRemoveRPC.Call(context.Background(), utils.AdminSv1RemoveActionProfile,
		&utils.TenantID{Tenant: doubleRemoveTenant, ID: "ACTION_PROFILE"}, &result); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := doubleRemoveRPC.Call(context.Background(), utils.AdminSv1RemoveActionProfile,
		&utils.TenantID{Tenant: doubleRemoveTenant, ID: "ACTION_PROFILE"}, &result); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	// check
	if err := doubleRemoveRPC.Call(context.Background(), utils.AdminSv1GetActionProfile,
		&utils.TenantID{Tenant: doubleRemoveTenant, ID: "ACTION_PROFILE"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testdoubleRemoveKillEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
