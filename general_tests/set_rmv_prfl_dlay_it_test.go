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
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/birpc/jsonrpc"
	"github.com/cgrates/cgrates/apis"
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
		cfgDir = "apis_config_mysql"
	case utils.MetaMongo:
		cfgDir = "apis_config_mongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cfgPath := path.Join("/usr/share/cgrates", "conf", "samples", cfgDir)
	cfg, err := config.NewCGRConfigFromPath(ctx, cfgPath)
	// Flush DBs.
	if err := engine.InitDataDB(cfg); err != nil {
		t.Fatal(err)
	}
	if err := engine.InitStorDB(cfg); err != nil {
		t.Fatal(err)
	}
	if _, err := engine.StopStartEngine(cfgPath, 100); err != nil {
		t.Fatal(err)
	}
	defer engine.KillEngine(100)
	client, err := jsonrpc.Dial(utils.TCP, cfg.ListenCfg().RPCJSONListen)
	if err != nil {
		t.Fatalf("could not establish connection to engine: %v", err)
	}

	t.Run("SetAttributeProfile", func(t *testing.T) {
		eAttrPrf := &engine.APIAttributeProfileWithAPIOpts{
			APIAttributeProfile: &engine.APIAttributeProfile{
				Tenant:    "cgrates.org",
				ID:        "ATTR_1",
				FilterIDs: []string{"*string:~*req.Account:acc"},
				Attributes: []*engine.ExternalAttribute{
					{
						Path:  utils.MetaReq + utils.NestingSep + utils.AccountField,
						Type:  utils.MetaConstant,
						Value: "1001",
					},
				},
				Weights: utils.DynamicWeights{
					{
						Weight: 10,
					},
				},
			},
		}
		var result string
		startTime := time.Now()
		if err := client.Call(context.Background(), utils.AdminSv1SetAttributeProfile, eAttrPrf, &result); err != nil {
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
		if err := client.Call(context.Background(), utils.AdminSv1RemoveAttributeProfile, eAttrPrf, &result); err != nil {
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

		eDspPrf := &apis.DispatcherWithAPIOpts{
			DispatcherProfile: &engine.DispatcherProfile{
				Tenant: "cgrates.org",
				ID:     "DSP_1",
			},
		}

		var result string
		startTime := time.Now()
		if err := client.Call(context.Background(), utils.AdminSv1SetDispatcherProfile, eDspPrf, &result); err != nil {
			t.Error(err)
		} else if result != utils.OK {
			t.Error("Unexpected reply returned", result)
		}
		elapsedTime := time.Since(startTime)
		expectedDuration := 3 * time.Second
		if elapsedTime < expectedDuration || elapsedTime >= 4*time.Second {
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
		if err := client.Call(context.Background(), utils.AdminSv1RemoveDispatcherProfile, eDspPrf, &result); err != nil {
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
		if err := client.Call(context.Background(), utils.AdminSv1SetResourceProfile, eRscPrf, &result); err != nil {
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
		if err := client.Call(context.Background(), utils.AdminSv1RemoveResourceProfile, eRscPrf, &result); err != nil {
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
		if err := client.Call(context.Background(), utils.AdminSv1SetStatQueueProfile, eSQPrf, &result); err != nil {
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
		if err := client.Call(context.Background(), utils.AdminSv1RemoveStatQueueProfile, eSQPrf, &result); err != nil {
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
		if err := client.Call(context.Background(), utils.AdminSv1SetThresholdProfile, eTHPrf, &result); err != nil {
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
		if err := client.Call(context.Background(), utils.AdminSv1RemoveThresholdProfile, eTHPrf, &result); err != nil {
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

		eRatingPrf := &utils.RateProfileWithAPIOpts{
			RateProfile: &utils.RateProfile{
				ID:     "RATE_1",
				Tenant: "cgrates.org",
			},
		}

		var result string
		startTime := time.Now()
		if err := client.Call(context.Background(), utils.ReplicatorSv1SetRateProfile, eRatingPrf, &result); err != nil {
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
				Attributes: []*engine.Attribute{
					{
						Path:  utils.MetaReq + utils.NestingSep + utils.AccountField,
						Value: config.NewRSRParsersMustCompile("1001", utils.InfieldSep),
					},
				},
				Weights: utils.DynamicWeights{
					{
						Weight: 10,
					},
				},
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
		eRtingPrf := &utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "RATE_1",
			},
		}

		var result string
		startTime := time.Now()
		if err := client.Call(context.Background(), utils.ReplicatorSv1RemoveRateProfile, eRtingPrf, &result); err != nil {
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
