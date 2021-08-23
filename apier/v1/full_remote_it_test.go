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

package v1

import (
	"net/rpc"
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	fullRemInternalCfgPath    string
	fullRemInternalCfgDirPath string
	fullRemInternalCfg        *config.CGRConfig
	fullRemInternalRPC        *rpc.Client

	fullRemEngineOneCfgPath    string
	fullRemEngineOneCfgDirPath string
	fullRemEngineOneCfg        *config.CGRConfig
	fullRemEngineOneRPC        *rpc.Client

	sTestsFullRemoteIT = []func(t *testing.T){
		testFullRemoteITInitCfg,
		testFullRemoteITDataFlush,
		testFullRemoteITStartEngine,
		testFullRemoteITRPCConn,
		testFullRemoteITAttribute,
		testFullRemoteITStatQueue,
		testFullRemoteITThreshold,
		testFullRemoteITResource,
		testFullRemoteITRoute,
		testFullRemoteITFilter,
		testFullRemoteITCharger,
		testFullRemoteITDispatcher,
		testFullRemoteITKillEngine,
	}
)

func TestFullRemoteIT(t *testing.T) {
	fullRemInternalCfgDirPath = "internal"
	fullRemEngineOneCfgDirPath = "remote"

	for _, stest := range sTestsFullRemoteIT {
		t.Run(*dbType, stest)
	}
}

func testFullRemoteITInitCfg(t *testing.T) {
	var err error
	fullRemInternalCfgPath = path.Join(*dataDir, "conf", "samples", "full_remote", fullRemInternalCfgDirPath)
	fullRemInternalCfg, err = config.NewCGRConfigFromPath(fullRemInternalCfgPath)
	if err != nil {
		t.Error(err)
	}

	// prepare config for engine1
	fullRemEngineOneCfgPath = path.Join(*dataDir, "conf", "samples",
		"full_remote", fullRemEngineOneCfgDirPath)
	fullRemEngineOneCfg, err = config.NewCGRConfigFromPath(fullRemEngineOneCfgPath)
	if err != nil {
		t.Error(err)
	}
	fullRemEngineOneCfg.DataFolderPath = *dataDir // Share DataFolderPath through config towards StoreDb for Flush()

}

func testFullRemoteITDataFlush(t *testing.T) {
	if err := engine.InitDataDb(fullRemEngineOneCfg); err != nil {
		t.Fatal(err)
	}
	time.Sleep(100 * time.Millisecond)
}

func testFullRemoteITStartEngine(t *testing.T) {
	engine.KillEngine(100)
	if _, err := engine.StartEngine(fullRemInternalCfgPath, 500); err != nil {
		t.Fatal(err)
	}
	if _, err := engine.StartEngine(fullRemEngineOneCfgPath, 500); err != nil {
		t.Fatal(err)
	}
	time.Sleep(200 * time.Millisecond)
}

func testFullRemoteITRPCConn(t *testing.T) {
	var err error
	fullRemInternalRPC, err = newRPCClient(fullRemInternalCfg.ListenCfg())
	if err != nil {
		t.Fatal(err)
	}
	fullRemEngineOneRPC, err = newRPCClient(fullRemEngineOneCfg.ListenCfg())
	if err != nil {
		t.Fatal(err)
	}
}

func testFullRemoteITAttribute(t *testing.T) {
	// verify for not found in internal
	var reply *engine.AttributeProfile
	if err := fullRemInternalRPC.Call(utils.APIerSv1GetAttributeProfile,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ATTR_1001_SIMPLEAUTH"}},
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Fatal(err)
	}

	var replySet string
	alsPrf := &engine.AttributeProfileWithAPIOpts{
		AttributeProfile: &engine.AttributeProfile{
			Tenant:    "cgrates.org",
			ID:        "ATTR_1001_SIMPLEAUTH",
			Contexts:  []string{"simpleauth"},
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Attributes: []*engine.Attribute{{
				Path:  utils.MetaReq + utils.NestingSep + "Password",
				Type:  utils.MetaConstant,
				Value: config.NewRSRParsersMustCompile("CGRateS.org", utils.InfieldSep),
			}},
			Weight: 20,
		},
	}
	alsPrf.Compile()
	// add an attribute profile in engine1 and verify it internal
	if err := fullRemEngineOneRPC.Call(utils.APIerSv1SetAttributeProfile, alsPrf, &replySet); err != nil {
		t.Error(err)
	} else if replySet != utils.OK {
		t.Error("Unexpected reply returned", replySet)
	}

	if err := fullRemInternalRPC.Call(utils.APIerSv1GetAttributeProfile,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ATTR_1001_SIMPLEAUTH"}},
		&reply); err != nil {
		t.Fatal(err)
	}
	reply.Compile()
	if !reflect.DeepEqual(alsPrf.AttributeProfile, reply) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(alsPrf.AttributeProfile), utils.ToJSON(reply))
	}
	// update the attribute profile and verify it to be updated
	alsPrf.FilterIDs = []string{"*string:~*req.Account:1001", "*string:~*req.Destination:1002"}
	alsPrf.Compile()
	if err := fullRemEngineOneRPC.Call(utils.APIerSv1SetAttributeProfile, alsPrf, &replySet); err != nil {
		t.Error(err)
	} else if replySet != utils.OK {
		t.Error("Unexpected reply returned", replySet)
	}

	if err := fullRemInternalRPC.Call(utils.APIerSv1GetAttributeProfile,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ATTR_1001_SIMPLEAUTH"}},
		&reply); err != nil {
		t.Fatal(err)
	}
	reply.Compile()
	if !reflect.DeepEqual(alsPrf.AttributeProfile, reply) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(alsPrf.AttributeProfile), utils.ToJSON(reply))
	}
}

func testFullRemoteITStatQueue(t *testing.T) {
	// verify for not found in internal
	var reply *engine.StatQueueProfile
	if err := fullRemInternalRPC.Call(utils.APIerSv1GetStatQueueProfile,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}},
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Fatal(err)
	}

	var replySet string
	stat := &engine.StatQueueProfileWithAPIOpts{
		StatQueueProfile: &engine.StatQueueProfile{
			Tenant:    "cgrates.org",
			ID:        "TEST_PROFILE1",
			FilterIDs: []string{"*string:~*req.Account:1001"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
			QueueLength: 10,
			TTL:         10 * time.Second,
			Metrics: []*engine.MetricWithFilters{
				{
					MetricID: utils.MetaACD,
				},
				{
					MetricID: utils.MetaTCD,
				},
			},
			ThresholdIDs: []string{"Val1", "Val2"},
			Blocker:      true,
			Stored:       true,
			Weight:       20,
			MinItems:     1,
		},
	}
	// add a statQueue profile in engine1 and verify it internal
	if err := fullRemEngineOneRPC.Call(utils.APIerSv1SetStatQueueProfile, stat, &replySet); err != nil {
		t.Error(err)
	} else if replySet != utils.OK {
		t.Error("Unexpected reply returned", replySet)
	}

	if err := fullRemInternalRPC.Call(utils.APIerSv1GetStatQueueProfile,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}},
		&reply); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(stat.StatQueueProfile, reply) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(stat.StatQueueProfile), utils.ToJSON(reply))
	}
	// update the statQueue profile and verify it to be updated
	stat.FilterIDs = []string{"*string:~*req.Account:1001", "*string:~*req.Destination:1002"}
	if err := fullRemEngineOneRPC.Call(utils.APIerSv1SetStatQueueProfile, stat, &replySet); err != nil {
		t.Error(err)
	} else if replySet != utils.OK {
		t.Error("Unexpected reply returned", replySet)
	}

	if err := fullRemInternalRPC.Call(utils.APIerSv1GetStatQueueProfile,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}},
		&reply); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(stat.StatQueueProfile, reply) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(stat.StatQueueProfile), utils.ToJSON(reply))
	}
}

func testFullRemoteITThreshold(t *testing.T) {
	// verify for not found in internal
	var reply *engine.ThresholdProfile
	if err := fullRemInternalRPC.Call(utils.APIerSv1GetThresholdProfile,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "THD_Test"}},
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Fatal(err)
	}

	var replySet string
	tPrfl := &engine.ThresholdProfileWithAPIOpts{
		ThresholdProfile: &engine.ThresholdProfile{
			Tenant:    "cgrates.org",
			ID:        "THD_Test",
			FilterIDs: []string{"*string:~*req.Account:1001"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			MaxHits:   -1,
			MinSleep:  5 * time.Minute,
			Blocker:   false,
			Weight:    20.0,
			ActionIDs: []string{"ACT_1"},
			Async:     true,
		},
	}
	// add a threshold profile in engine1 and verify it internal
	if err := fullRemEngineOneRPC.Call(utils.APIerSv1SetThresholdProfile, tPrfl, &replySet); err != nil {
		t.Error(err)
	} else if replySet != utils.OK {
		t.Error("Unexpected reply returned", replySet)
	}

	if err := fullRemInternalRPC.Call(utils.APIerSv1GetThresholdProfile,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "THD_Test"}},
		&reply); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(tPrfl.ThresholdProfile, reply) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(tPrfl.ThresholdProfile), utils.ToJSON(reply))
	}
	// update the threshold profile and verify it to be updated
	tPrfl.FilterIDs = []string{"*string:~*req.Account:1001", "*string:~*req.Destination:1002"}
	if err := fullRemEngineOneRPC.Call(utils.APIerSv1SetThresholdProfile, tPrfl, &replySet); err != nil {
		t.Error(err)
	} else if replySet != utils.OK {
		t.Error("Unexpected reply returned", replySet)
	}

	if err := fullRemInternalRPC.Call(utils.APIerSv1GetThresholdProfile,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "THD_Test"}},
		&reply); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(tPrfl.ThresholdProfile, reply) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(tPrfl.ThresholdProfile), utils.ToJSON(reply))
	}
}

func testFullRemoteITResource(t *testing.T) {
	// verify for not found in internal
	var reply *engine.ResourceProfile
	if err := fullRemInternalRPC.Call(utils.APIerSv1GetResourceProfile,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ResGroup1"}},
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Fatal(err)
	}

	var replySet string
	rlsPrf := &engine.ResourceProfileWithAPIOpts{
		ResourceProfile: &engine.ResourceProfile{
			Tenant:    "cgrates.org",
			ID:        "ResGroup1",
			FilterIDs: []string{"*string:~*req.Account:1001"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 29, 15, 0, 0, 0, time.UTC),
			},
			UsageTTL:          -1,
			Limit:             7,
			AllocationMessage: "",
			Stored:            true,
			Weight:            10,
			ThresholdIDs:      []string{utils.MetaNone},
		},
	}
	// add a threshold profile in engine1 and verify it internal
	if err := fullRemEngineOneRPC.Call(utils.APIerSv1SetResourceProfile, rlsPrf, &replySet); err != nil {
		t.Error(err)
	} else if replySet != utils.OK {
		t.Error("Unexpected reply returned", replySet)
	}

	if err := fullRemInternalRPC.Call(utils.APIerSv1GetResourceProfile,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ResGroup1"}},
		&reply); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(rlsPrf.ResourceProfile, reply) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(rlsPrf.ResourceProfile), utils.ToJSON(reply))
	}
	// update the threshold profile and verify it to be updated
	rlsPrf.FilterIDs = []string{"*string:~*req.Account:1001", "*string:~*req.Destination:1002"}
	if err := fullRemEngineOneRPC.Call(utils.APIerSv1SetResourceProfile, rlsPrf, &replySet); err != nil {
		t.Error(err)
	} else if replySet != utils.OK {
		t.Error("Unexpected reply returned", replySet)
	}

	if err := fullRemInternalRPC.Call(utils.APIerSv1GetResourceProfile,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ResGroup1"}},
		&reply); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(rlsPrf.ResourceProfile, reply) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(rlsPrf.ResourceProfile), utils.ToJSON(reply))
	}
}

func testFullRemoteITRoute(t *testing.T) {
	// verify for not found in internal
	var reply *engine.RouteProfile
	if err := fullRemInternalRPC.Call(utils.APIerSv1GetRouteProfile,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ROUTE_ACNT_1001"}},
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Fatal(err)
	}

	var replySet string
	routePrf := &engine.RouteProfile{
		Tenant:    "cgrates.org",
		ID:        "ROUTE_ACNT_1001",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2017, 11, 27, 0, 0, 0, 0, time.UTC),
		},
		Sorting: utils.MetaWeight,
		Routes: []*engine.Route{{
			ID:     "route1",
			Weight: 10,
		}, {
			ID:     "route2",
			Weight: 20,
		}},
		Weight: 20,
	}
	// add a threshold profile in engine1 and verify it internal
	if err := fullRemEngineOneRPC.Call(utils.APIerSv1SetRouteProfile, routePrf, &replySet); err != nil {
		t.Error(err)
	} else if replySet != utils.OK {
		t.Error("Unexpected reply returned", replySet)
	}

	if err := fullRemInternalRPC.Call(utils.APIerSv1GetRouteProfile,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ROUTE_ACNT_1001"}},
		&reply); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(routePrf, reply) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(routePrf), utils.ToJSON(reply))
	}
	// update the threshold profile and verify it to be updated
	routePrf.FilterIDs = []string{"*string:~*req.Account:1001", "*string:~*req.Destination:1002"}
	if err := fullRemEngineOneRPC.Call(utils.APIerSv1SetRouteProfile, routePrf, &replySet); err != nil {
		t.Error(err)
	} else if replySet != utils.OK {
		t.Error("Unexpected reply returned", replySet)
	}

	if err := fullRemInternalRPC.Call(utils.APIerSv1GetRouteProfile,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ROUTE_ACNT_1001"}},
		&reply); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(routePrf, reply) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(routePrf), utils.ToJSON(reply))
	}
}

func testFullRemoteITFilter(t *testing.T) {
	// verify for not found in internal
	var reply *engine.Filter
	if err := fullRemInternalRPC.Call(utils.APIerSv1GetFilter,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "FLTR_ACNT_1001"}},
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Fatal(err)
	}

	var replySet string
	fltr := &engine.Filter{
		Tenant: "cgrates.org",
		ID:     "FLTR_ACNT_1001",
		Rules: []*engine.FilterRule{
			{
				Type:    utils.MetaString,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.AccountField,
				Values:  []string{"1001"},
			},
		},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 29, 15, 0, 0, 0, time.UTC),
		},
	}
	// add a threshold profile in engine1 and verify it internal
	if err := fullRemEngineOneRPC.Call(utils.APIerSv1SetFilter, fltr, &replySet); err != nil {
		t.Error(err)
	} else if replySet != utils.OK {
		t.Error("Unexpected reply returned", replySet)
	}

	if err := fullRemInternalRPC.Call(utils.APIerSv1GetFilter,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "FLTR_ACNT_1001"}},
		&reply); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(fltr, reply) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(fltr), utils.ToJSON(reply))
	}
	// update the threshold profile and verify it to be updated
	fltr.Rules = []*engine.FilterRule{
		{
			Element: "~*req.Account",
			Type:    utils.MetaString,
			Values:  []string{"1001", "1002"},
		},
		{
			Element: "~*req.Destination",
			Type:    utils.MetaPrefix,
			Values:  []string{"10", "20"},
		},
	}
	if err := fullRemEngineOneRPC.Call(utils.APIerSv1SetFilter, fltr, &replySet); err != nil {
		t.Error(err)
	} else if replySet != utils.OK {
		t.Error("Unexpected reply returned", replySet)
	}

	if err := fullRemInternalRPC.Call(utils.APIerSv1GetFilter,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "FLTR_ACNT_1001"}},
		&reply); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(fltr, reply) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(fltr), utils.ToJSON(reply))
	}
}

func testFullRemoteITCharger(t *testing.T) {
	// verify for not found in internal
	var reply *engine.ChargerProfile
	if err := fullRemInternalRPC.Call(utils.APIerSv1GetChargerProfile,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "DEFAULT"}},
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Fatal(err)
	}

	var replySet string
	chargerProfile := &engine.ChargerProfile{
		Tenant:       "cgrates.org",
		ID:           "DEFAULT",
		RunID:        utils.MetaDefault,
		AttributeIDs: []string{utils.MetaNone},
		Weight:       0,
	}
	// add a threshold profile in engine1 and verify it internal
	if err := fullRemEngineOneRPC.Call(utils.APIerSv1SetChargerProfile, chargerProfile, &replySet); err != nil {
		t.Error(err)
	} else if replySet != utils.OK {
		t.Error("Unexpected reply returned", replySet)
	}

	if err := fullRemInternalRPC.Call(utils.APIerSv1GetChargerProfile,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "DEFAULT"}},
		&reply); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(chargerProfile, reply) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(chargerProfile), utils.ToJSON(reply))
	}
	// update the threshold profile and verify it to be updated
	chargerProfile.FilterIDs = []string{"*string:~*req.Account:1001", "*string:~*req.Destination:1002"}
	if err := fullRemEngineOneRPC.Call(utils.APIerSv1SetChargerProfile, chargerProfile, &replySet); err != nil {
		t.Error(err)
	} else if replySet != utils.OK {
		t.Error("Unexpected reply returned", replySet)
	}

	if err := fullRemInternalRPC.Call(utils.APIerSv1GetChargerProfile,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "DEFAULT"}},
		&reply); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(chargerProfile, reply) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(chargerProfile), utils.ToJSON(reply))
	}
}

func testFullRemoteITDispatcher(t *testing.T) {
	// verify for not found in internal
	var reply *engine.DispatcherProfile
	if err := fullRemInternalRPC.Call(utils.APIerSv1GetDispatcherProfile,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "Dsp1"}},
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Fatal(err)
	}

	var replySet string
	dispatcherProfile = &DispatcherWithAPIOpts{
		DispatcherProfile: &engine.DispatcherProfile{
			Tenant:     "cgrates.org",
			ID:         "Dsp1",
			FilterIDs:  []string{"*string:~*req.Account:1001"},
			Subsystems: []string{utils.MetaAny},
			Strategy:   utils.MetaFirst,
			Weight:     20,
		},
	}
	// add a threshold profile in engine1 and verify it internal
	if err := fullRemEngineOneRPC.Call(utils.APIerSv1SetDispatcherProfile, dispatcherProfile, &replySet); err != nil {
		t.Error(err)
	} else if replySet != utils.OK {
		t.Error("Unexpected reply returned", replySet)
	}

	if err := fullRemInternalRPC.Call(utils.APIerSv1GetDispatcherProfile,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "Dsp1"}},
		&reply); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(dispatcherProfile.DispatcherProfile, reply) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(dispatcherProfile.DispatcherProfile), utils.ToJSON(reply))
	}
	// update the threshold profile and verify it to be updated
	dispatcherProfile.FilterIDs = []string{"*string:~*req.Account:1001", "*string:~*req.Destination:1002"}
	if err := fullRemEngineOneRPC.Call(utils.APIerSv1SetDispatcherProfile, dispatcherProfile, &replySet); err != nil {
		t.Error(err)
	} else if replySet != utils.OK {
		t.Error("Unexpected reply returned", replySet)
	}

	if err := fullRemInternalRPC.Call(utils.APIerSv1GetDispatcherProfile,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "Dsp1"}},
		&reply); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(dispatcherProfile.DispatcherProfile, reply) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(dispatcherProfile.DispatcherProfile), utils.ToJSON(reply))
	}
}

func testFullRemoteITKillEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
