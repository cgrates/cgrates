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
	v2 "github.com/cgrates/cgrates/apier/v2"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestSetRemoveProfilesWithCachingDelay(t *testing.T) {
	var cfgDir string
	switch *utils.DBType {
	case utils.MetaInternal:
		t.SkipNow()
	case utils.MetaMySQL:
		cfgDir = "apier_mysql"
	case utils.MetaMongo:
		cfgDir = "apier_mongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	ng := engine.TestEngine{
		ConfigPath: path.Join(*utils.DataDir, "conf", "samples", cfgDir),
		TpPath:     path.Join(*utils.DataDir, "tariffplans", "tutorial"),
	}
	client, _ := ng.Run(t)

	t.Run("RemoveTPFromFolder", func(t *testing.T) {
		var reply string
		attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*utils.DataDir, "tariffplans", "tutorial")}
		startTime := time.Now()
		if err := client.Call(context.Background(), utils.APIerSv1RemoveTPFromFolder, attrs, &reply); err != nil {
			t.Error(err)
		} else if reply != utils.OK {
			t.Errorf("expected reply <OK>, received <%v>", reply)
		}
		elapsedTime := time.Since(startTime)
		expectedDuration := 1 * time.Second
		if elapsedTime < expectedDuration || elapsedTime >= 2*time.Second {
			t.Errorf("Expected elapsed time of at least %v, but got %v", expectedDuration, elapsedTime)
		}
		time.Sleep(100 * time.Millisecond)
	})

	t.Run("SetAttributeProfile", func(t *testing.T) {
		ev := &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testAttributeSGetAttributeForEventWihMetaAnyContext",
			Event: map[string]any{
				utils.AccountField: "acc",
				utils.Destination:  "+491511231234",
			},
			APIOpts: map[string]any{
				utils.OptsContext: utils.MetaCDRs,
			},
		}
		eAttrPrf := &engine.AttributeProfileWithAPIOpts{
			AttributeProfile: &engine.AttributeProfile{
				Tenant:    ev.Tenant,
				ID:        "ATTR_1",
				FilterIDs: []string{"*string:~*req.Account:acc"},
				Contexts:  []string{utils.MetaAny},
				ActivationInterval: &utils.ActivationInterval{
					ActivationTime: time.Date(2014, 1, 14, 0, 0, 0, 0, time.UTC)},
				Attributes: []*engine.Attribute{
					{
						Path:  utils.MetaReq + utils.NestingSep + utils.AccountField,
						Value: config.NewRSRParsersMustCompile("1001", utils.InfieldSep),
					},
				},
				Weight: 10.0,
			},
		}
		eAttrPrf.Compile()
		var result string
		startTime := time.Now()
		if err := client.Call(context.Background(), utils.APIerSv1SetAttributeProfile, eAttrPrf, &result); err != nil {
			t.Error(err)
		} else if result != utils.OK {
			t.Error("Unexpected reply returned", result)
		}
		elapsedTime := time.Since(startTime)
		expectedDuration := 1 * time.Second
		if elapsedTime < expectedDuration || elapsedTime >= 2*time.Second {
			t.Errorf("Expected elapsed time of at least %v, but got %v", expectedDuration, elapsedTime)
		}
	})

	t.Run("RemoveAttributeProfile", func(t *testing.T) {

		eAttrPrf := &utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "ATTR_1",
			},
		}

		var result string
		startTime := time.Now()
		if err := client.Call(context.Background(), utils.APIerSv1RemoveAttributeProfile, eAttrPrf, &result); err != nil {
			t.Error(err)
		} else if result != utils.OK {
			t.Error("Unexpected reply returned", result)
		}
		elapsedTime := time.Since(startTime)
		expectedDuration := 1 * time.Second
		if elapsedTime < expectedDuration || elapsedTime >= 2*time.Second {
			t.Errorf("Expected elapsed time of at least %v, but got %v", expectedDuration, elapsedTime)
		}
	})

	t.Run("SetDispatcherProfile", func(t *testing.T) {

		eDspPrf := &engine.DispatcherProfileWithAPIOpts{
			DispatcherProfile: &engine.DispatcherProfile{
				Tenant: "cgrates.org",
				ID:     "DSP_1",
			},
		}

		var result string
		startTime := time.Now()
		if err := client.Call(context.Background(), utils.APIerSv1SetDispatcherProfile, eDspPrf, &result); err != nil {
			t.Error(err)
		} else if result != utils.OK {
			t.Error("Unexpected reply returned", result)
		}
		elapsedTime := time.Since(startTime)
		expectedDuration := 1 * time.Second
		if elapsedTime < expectedDuration || elapsedTime >= 2*time.Second {
			t.Errorf("Expected elapsed time of at least %v, but got %v", expectedDuration, elapsedTime)
		}
	})

	t.Run("RemoveDispatcherProfile", func(t *testing.T) {

		eDspPrf := &utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "DSP_1",
			},
		}

		var result string
		startTime := time.Now()
		if err := client.Call(context.Background(), utils.APIerSv1RemoveDispatcherProfile, eDspPrf, &result); err != nil {
			t.Error(err)
		} else if result != utils.OK {
			t.Error("Unexpected reply returned", result)
		}
		elapsedTime := time.Since(startTime)
		expectedDuration := 1 * time.Second
		if elapsedTime < expectedDuration || elapsedTime >= 2*time.Second {
			t.Errorf("Expected elapsed time of at least %v, but got %v", expectedDuration, elapsedTime)
		}
	})

	t.Run("SetResourceProfile", func(t *testing.T) {

		eRscPrf := &engine.ResourceProfileWithAPIOpts{
			ResourceProfile: &engine.ResourceProfile{
				Tenant: "cgrates.org",
				ID:     "RSC_1",
			},
		}

		var result string
		startTime := time.Now()
		if err := client.Call(context.Background(), utils.APIerSv1SetResourceProfile, eRscPrf, &result); err != nil {
			t.Error(err)
		} else if result != utils.OK {
			t.Error("Unexpected reply returned", result)
		}
		elapsedTime := time.Since(startTime)
		expectedDuration := 1 * time.Second
		if elapsedTime < expectedDuration || elapsedTime >= 2*time.Second {
			t.Errorf("Expected elapsed time of at least %v, but got %v", expectedDuration, elapsedTime)
		}
	})

	t.Run("RemoveResourceProfile", func(t *testing.T) {

		eRscPrf := &utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "RSC_1",
			},
		}

		var result string
		startTime := time.Now()
		if err := client.Call(context.Background(), utils.APIerSv1RemoveResourceProfile, eRscPrf, &result); err != nil {
			t.Error(err)
		} else if result != utils.OK {
			t.Error("Unexpected reply returned", result)
		}
		elapsedTime := time.Since(startTime)
		expectedDuration := 1 * time.Second
		if elapsedTime < expectedDuration || elapsedTime >= 2*time.Second {
			t.Errorf("Expected elapsed time of at least %v, but got %v", expectedDuration, elapsedTime)
		}
	})

	t.Run("SetRouteProfile", func(t *testing.T) {

		eRoutePrf := &engine.RouteProfileWithAPIOpts{
			RouteProfile: &engine.RouteProfile{
				Tenant: "cgrates.org",
				ID:     "ROUTE_1",
			},
		}

		var result string
		startTime := time.Now()
		if err := client.Call(context.Background(), utils.APIerSv1SetRouteProfile, eRoutePrf, &result); err != nil {
			t.Error(err)
		} else if result != utils.OK {
			t.Error("Unexpected reply returned", result)
		}
		elapsedTime := time.Since(startTime)
		expectedDuration := 1 * time.Second
		if elapsedTime < expectedDuration || elapsedTime >= 2*time.Second {
			t.Errorf("Expected elapsed time of at least %v, but got %v", expectedDuration, elapsedTime)
		}
	})

	t.Run("RemoveRouteProfile", func(t *testing.T) {

		eRoutePrf := &utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "ROUTE_1",
			},
		}

		var result string
		startTime := time.Now()
		if err := client.Call(context.Background(), utils.APIerSv1RemoveRouteProfile, eRoutePrf, &result); err != nil {
			t.Error(err)
		} else if result != utils.OK {
			t.Error("Unexpected reply returned", result)
		}
		elapsedTime := time.Since(startTime)
		expectedDuration := 1 * time.Second
		if elapsedTime < expectedDuration || elapsedTime >= 2*time.Second {
			t.Errorf("Expected elapsed time of at least %v, but got %v", expectedDuration, elapsedTime)
		}
	})

	t.Run("SetStatQueueProfile", func(t *testing.T) {

		eSQPrf := &engine.StatQueueProfileWithAPIOpts{
			StatQueueProfile: &engine.StatQueueProfile{
				Tenant: "cgrates.org",
				ID:     "SQ_1",
			},
		}

		var result string
		startTime := time.Now()
		if err := client.Call(context.Background(), utils.APIerSv1SetStatQueueProfile, eSQPrf, &result); err != nil {
			t.Error(err)
		} else if result != utils.OK {
			t.Error("Unexpected reply returned", result)
		}
		elapsedTime := time.Since(startTime)
		expectedDuration := 1 * time.Second
		if elapsedTime < expectedDuration || elapsedTime >= 2*time.Second {
			t.Errorf("Expected elapsed time of at least %v, but got %v", expectedDuration, elapsedTime)
		}
	})

	t.Run("RemoveStatQueueProfile", func(t *testing.T) {

		eSQPrf := &utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "SQ_1",
			},
		}

		var result string
		startTime := time.Now()
		if err := client.Call(context.Background(), utils.APIerSv1RemoveStatQueueProfile, eSQPrf, &result); err != nil {
			t.Error(err)
		} else if result != utils.OK {
			t.Error("Unexpected reply returned", result)
		}
		elapsedTime := time.Since(startTime)
		expectedDuration := 1 * time.Second
		if elapsedTime < expectedDuration || elapsedTime >= 2*time.Second {
			t.Errorf("Expected elapsed time of at least %v, but got %v", expectedDuration, elapsedTime)
		}
	})

	t.Run("SetThresholdProfile", func(t *testing.T) {

		eTHPrf := &engine.ThresholdProfileWithAPIOpts{
			ThresholdProfile: &engine.ThresholdProfile{
				Tenant: "cgrates.org",
				ID:     "THRHLD_1",
			},
		}

		var result string
		startTime := time.Now()
		if err := client.Call(context.Background(), utils.APIerSv1SetThresholdProfile, eTHPrf, &result); err != nil {
			t.Error(err)
		} else if result != utils.OK {
			t.Error("Unexpected reply returned", result)
		}
		elapsedTime := time.Since(startTime)
		expectedDuration := 1 * time.Second
		if elapsedTime < expectedDuration || elapsedTime >= 2*time.Second {
			t.Errorf("Expected elapsed time of at least %v, but got %v", expectedDuration, elapsedTime)
		}
	})

	t.Run("RemoveThresholdProfile", func(t *testing.T) {

		eTHPrf := &utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "THRHLD_1",
			},
		}

		var result string
		startTime := time.Now()
		if err := client.Call(context.Background(), utils.APIerSv1RemoveThresholdProfile, eTHPrf, &result); err != nil {
			t.Error(err)
		} else if result != utils.OK {
			t.Error("Unexpected reply returned", result)
		}
		elapsedTime := time.Since(startTime)
		expectedDuration := 1 * time.Second
		if elapsedTime < expectedDuration || elapsedTime >= 2*time.Second {
			t.Errorf("Expected elapsed time of at least %v, but got %v", expectedDuration, elapsedTime)
		}
	})

	t.Run("V2SetAttributeProfile", func(t *testing.T) {

		extAlsPrf := &v2.AttributeWithAPIOpts{
			APIAttributeProfile: &engine.APIAttributeProfile{
				Tenant:    "cgrates.org",
				ID:        "ExternalAttribute",
				Contexts:  []string{utils.MetaSessionS, utils.MetaCDRs},
				FilterIDs: []string{"*string:~*req.Account:1001"},
				ActivationInterval: &utils.ActivationInterval{
					ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
					ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
				},
				Attributes: []*engine.ExternalAttribute{
					{
						Path:  utils.MetaReq + utils.NestingSep + "Account",
						Value: "1001",
					},
				},
				Weight: 20,
			},
		}
		var result string
		startTime := time.Now()
		if err := client.Call(context.Background(), utils.APIerSv2SetAttributeProfile, extAlsPrf, &result); err != nil {
			t.Error(err)
		} else if result != utils.OK {
			t.Error("Unexpected reply returned", result)
		}
		elapsedTime := time.Since(startTime)
		expectedDuration := 1 * time.Second
		if elapsedTime < expectedDuration || elapsedTime >= 2*time.Second {
			t.Errorf("Expected elapsed time of at least %v, but got %v", expectedDuration, elapsedTime)
		}
	})

	t.Run("V2LoadTariffPlanFromFolder", func(t *testing.T) {
		attrs := &utils.AttrLoadTpFromFolder{
			FolderPath: path.Join(*utils.DataDir, "tariffplans", "tutorial"),
		}

		exp := utils.LoadInstance{}

		var result utils.LoadInstance
		startTime := time.Now()
		if err := client.Call(context.Background(), utils.APIerSv2LoadTariffPlanFromFolder, attrs, &result); err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual(result, exp) {
			t.Errorf("Expected <%+v>, \nreceived \n<%+v>", exp, result)
		}
		elapsedTime := time.Since(startTime)
		expectedDuration := 1 * time.Second
		if elapsedTime < expectedDuration || elapsedTime >= 2*time.Second {
			t.Errorf("Expected elapsed time of at least %v, but got %v", expectedDuration, elapsedTime)
		}
	})

	t.Run("ReplicatorSv1SetThresholdProfile", func(t *testing.T) {

		eTHPrf := &engine.ThresholdProfileWithAPIOpts{
			ThresholdProfile: &engine.ThresholdProfile{
				Tenant: "cgrates.org",
				ID:     "THRHLD_2",
			},
		}

		var result string
		startTime := time.Now()
		if err := client.Call(context.Background(), utils.ReplicatorSv1SetThresholdProfile, eTHPrf, &result); err != nil {
			t.Error(err)
		} else if result != utils.OK {
			t.Error("Unexpected reply returned", result)
		}
		elapsedTime := time.Since(startTime)
		expectedDuration := 1 * time.Second
		if elapsedTime < expectedDuration || elapsedTime >= 2*time.Second {
			t.Errorf("Expected elapsed time of at least %v, but got %v", expectedDuration, elapsedTime)
		}
	})

	t.Run("ReplicatorSv1SetStatQueueProfile", func(t *testing.T) {

		eSQPrf := &engine.StatQueueProfileWithAPIOpts{
			StatQueueProfile: &engine.StatQueueProfile{
				Tenant: "cgrates.org",
				ID:     "SQ_2",
			},
		}

		var result string
		startTime := time.Now()
		if err := client.Call(context.Background(), utils.ReplicatorSv1SetStatQueueProfile, eSQPrf, &result); err != nil {
			t.Error(err)
		} else if result != utils.OK {
			t.Error("Unexpected reply returned", result)
		}
		elapsedTime := time.Since(startTime)
		expectedDuration := 1 * time.Second
		if elapsedTime < expectedDuration || elapsedTime >= 2*time.Second {
			t.Errorf("Expected elapsed time of at least %v, but got %v", expectedDuration, elapsedTime)
		}
	})

	t.Run("ReplicatorSv1SetResourceProfile", func(t *testing.T) {

		eRscPrf := &engine.ResourceProfileWithAPIOpts{
			ResourceProfile: &engine.ResourceProfile{
				Tenant: "cgrates.org",
				ID:     "RSC_2",
			},
		}

		var result string
		startTime := time.Now()
		if err := client.Call(context.Background(), utils.ReplicatorSv1SetResourceProfile, eRscPrf, &result); err != nil {
			t.Error(err)
		} else if result != utils.OK {
			t.Error("Unexpected reply returned", result)
		}
		elapsedTime := time.Since(startTime)
		expectedDuration := 1 * time.Second
		if elapsedTime < expectedDuration || elapsedTime >= 2*time.Second {
			t.Errorf("Expected elapsed time of at least %v, but got %v", expectedDuration, elapsedTime)
		}
	})

	t.Run("ReplicatorSv1SetRatingProfile", func(t *testing.T) {

		eRatingPrf := &engine.RatingProfileWithAPIOpts{
			Tenant: "cgrates.org",
			RatingProfile: &engine.RatingProfile{
				Id: "RATE_1",
			},
		}

		var result string
		startTime := time.Now()
		if err := client.Call(context.Background(), utils.ReplicatorSv1SetRatingProfile, eRatingPrf, &result); err != nil {
			t.Error(err)
		} else if result != utils.OK {
			t.Error("Unexpected reply returned", result)
		}
		elapsedTime := time.Since(startTime)
		expectedDuration := 1 * time.Second
		if elapsedTime < expectedDuration || elapsedTime >= 2*time.Second {
			t.Errorf("Expected elapsed time of at least %v, but got %v", expectedDuration, elapsedTime)
		}
	})

	t.Run("ReplicatorSv1SetRouteProfile", func(t *testing.T) {

		eRoutePrf := &engine.RouteProfileWithAPIOpts{
			RouteProfile: &engine.RouteProfile{
				Tenant: "cgrates.org",
				ID:     "ROUTE_2",
			},
		}

		var result string
		startTime := time.Now()
		if err := client.Call(context.Background(), utils.ReplicatorSv1SetRouteProfile, eRoutePrf, &result); err != nil {
			t.Error(err)
		} else if result != utils.OK {
			t.Error("Unexpected reply returned", result)
		}
		elapsedTime := time.Since(startTime)
		expectedDuration := 1 * time.Second
		if elapsedTime < expectedDuration || elapsedTime >= 2*time.Second {
			t.Errorf("Expected elapsed time of at least %v, but got %v", expectedDuration, elapsedTime)
		}
	})

	t.Run("ReplicatorSv1SetAttributeProfile", func(t *testing.T) {

		ev := &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testAttributeSGetAttributeForEventWihMetaAnyContext",
			Event: map[string]any{
				utils.AccountField: "acc",
				utils.Destination:  "+491511231234",
			},
			APIOpts: map[string]any{
				utils.OptsContext: utils.MetaCDRs,
			},
		}
		eAttrPrf := &engine.AttributeProfileWithAPIOpts{
			AttributeProfile: &engine.AttributeProfile{
				Tenant:    ev.Tenant,
				ID:        "ATTR_3",
				FilterIDs: []string{"*string:~*req.Account:acc"},
				Contexts:  []string{utils.MetaAny},
				ActivationInterval: &utils.ActivationInterval{
					ActivationTime: time.Date(2014, 1, 14, 0, 0, 0, 0, time.UTC)},
				Attributes: []*engine.Attribute{
					{
						Path:  utils.MetaReq + utils.NestingSep + utils.AccountField,
						Value: config.NewRSRParsersMustCompile("1001", utils.InfieldSep),
					},
				},
				Weight: 10.0,
			},
		}
		eAttrPrf.Compile()

		var result string
		startTime := time.Now()
		if err := client.Call(context.Background(), utils.ReplicatorSv1SetAttributeProfile, eAttrPrf, &result); err != nil {
			t.Error(err)
		} else if result != utils.OK {
			t.Error("Unexpected reply returned", result)
		}
		elapsedTime := time.Since(startTime)
		expectedDuration := 1 * time.Second
		if elapsedTime < expectedDuration || elapsedTime >= 2*time.Second {
			t.Errorf("Expected elapsed time of at least %v, but got %v", expectedDuration, elapsedTime)
		}
	})

	t.Run("ReplicatorSv1SetChargerProfile", func(t *testing.T) {

		eChrgPrf := &engine.ChargerProfileWithAPIOpts{
			ChargerProfile: &engine.ChargerProfile{
				Tenant: "cgrates.org",
				ID:     "CHRG_1",
			},
		}

		var result string
		startTime := time.Now()
		if err := client.Call(context.Background(), utils.ReplicatorSv1SetChargerProfile, eChrgPrf, &result); err != nil {
			t.Error(err)
		} else if result != utils.OK {
			t.Error("Unexpected reply returned", result)
		}
		elapsedTime := time.Since(startTime)
		expectedDuration := 1 * time.Second
		if elapsedTime < expectedDuration || elapsedTime >= 2*time.Second {
			t.Errorf("Expected elapsed time of at least %v, but got %v", expectedDuration, elapsedTime)
		}
	})

	t.Run("ReplicatorSv1SetDispatcherProfile", func(t *testing.T) {

		eDspPrf := &engine.DispatcherProfileWithAPIOpts{
			DispatcherProfile: &engine.DispatcherProfile{
				Tenant: "cgrates.org",
				ID:     "DSP_1",
			},
		}

		var result string
		startTime := time.Now()
		if err := client.Call(context.Background(), utils.ReplicatorSv1SetDispatcherProfile, eDspPrf, &result); err != nil {
			t.Error(err)
		} else if result != utils.OK {
			t.Error("Unexpected reply returned", result)
		}
		elapsedTime := time.Since(startTime)
		expectedDuration := 1 * time.Second
		if elapsedTime < expectedDuration || elapsedTime >= 2*time.Second {
			t.Errorf("Expected elapsed time of at least %v, but got %v", expectedDuration, elapsedTime)
		}
	})

	t.Run("ReplicatorSv1RemoveThresholdProfile", func(t *testing.T) {

		eTHPrf := &utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "THRHLD_2",
			},
		}

		var result string
		startTime := time.Now()
		if err := client.Call(context.Background(), utils.ReplicatorSv1RemoveThresholdProfile, eTHPrf, &result); err != nil {
			t.Error(err)
		} else if result != utils.OK {
			t.Error("Unexpected reply returned", result)
		}
		elapsedTime := time.Since(startTime)
		expectedDuration := 1 * time.Second
		if elapsedTime < expectedDuration || elapsedTime >= 2*time.Second {
			t.Errorf("Expected elapsed time of at least %v, but got %v", expectedDuration, elapsedTime)
		}
	})

	t.Run("ReplicatorSv1RemoveStatQueueProfile", func(t *testing.T) {

		eSQPrf := &utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "SQ_2",
			},
		}

		var result string
		startTime := time.Now()
		if err := client.Call(context.Background(), utils.ReplicatorSv1RemoveStatQueueProfile, eSQPrf, &result); err != nil {
			t.Error(err)
		} else if result != utils.OK {
			t.Error("Unexpected reply returned", result)
		}
		elapsedTime := time.Since(startTime)
		expectedDuration := 1 * time.Second
		if elapsedTime < expectedDuration || elapsedTime >= 2*time.Second {
			t.Errorf("Expected elapsed time of at least %v, but got %v", expectedDuration, elapsedTime)
		}
	})

	t.Run("ReplicatorSv1RemoveResourceProfile", func(t *testing.T) {
		eRscPrf := &utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "RSC_2",
			},
		}

		var result string
		startTime := time.Now()
		if err := client.Call(context.Background(), utils.ReplicatorSv1RemoveResourceProfile, eRscPrf, &result); err != nil {
			t.Error(err)
		} else if result != utils.OK {
			t.Error("Unexpected reply returned", result)
		}
		elapsedTime := time.Since(startTime)
		expectedDuration := 1 * time.Second
		if elapsedTime < expectedDuration || elapsedTime >= 2*time.Second {
			t.Errorf("Expected elapsed time of at least %v, but got %v", expectedDuration, elapsedTime)
		}
	})

	t.Run("ReplicatorSv1RemoveRatingProfile", func(t *testing.T) {
		eRtingPrf := &utils.StringWithAPIOpts{
			Tenant: "cgrates.org",
			Arg:    "RATE_1",
		}

		var result string
		startTime := time.Now()
		if err := client.Call(context.Background(), utils.ReplicatorSv1RemoveRatingProfile, eRtingPrf, &result); err != nil {
			t.Error(err)
		} else if result != utils.OK {
			t.Error("Unexpected reply returned", result)
		}
		elapsedTime := time.Since(startTime)
		expectedDuration := 1 * time.Second
		if elapsedTime < expectedDuration || elapsedTime >= 2*time.Second {
			t.Errorf("Expected elapsed time of at least %v, but got %v", expectedDuration, elapsedTime)
		}
	})

	t.Run("ReplicatorSv1RemoveRouteProfile", func(t *testing.T) {
		eRoutePrf := &utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "ROUTE_2",
			},
		}

		var result string
		startTime := time.Now()
		if err := client.Call(context.Background(), utils.ReplicatorSv1RemoveRouteProfile, eRoutePrf, &result); err != nil {
			t.Error(err)
		} else if result != utils.OK {
			t.Error("Unexpected reply returned", result)
		}
		elapsedTime := time.Since(startTime)
		expectedDuration := 1 * time.Second
		if elapsedTime < expectedDuration || elapsedTime >= 2*time.Second {
			t.Errorf("Expected elapsed time of at least %v, but got %v", expectedDuration, elapsedTime)
		}
	})

}
