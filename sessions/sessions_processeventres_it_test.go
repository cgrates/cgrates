//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package sessions

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestSessionSv1ProcessEventResourcesAuthorize(t *testing.T) {

	ng := engine.TestEngine{
		ConfigJSON: `{
"logger": { "level": 7 },
"sessions": {
	"enabled": true,
	"conns": {
		"*resources": [{"connIDs": ["*localhost"]}]
	}
},
"resources": {
	"enabled": true,
	"storeInterval": "-1"
},
"admins": { "enabled": true }
}`,
		DBCfg:    engine.InternalDBCfg,
		Encoding: *utils.Encoding,
		// LogBuffer: new(bytes.Buffer),
	}

	client, _ := ng.Run(t)

	// t.Cleanup(func() {
	// 	if ng.LogBuffer != nil {
	// 		fmt.Println(ng.LogBuffer)
	// 	}
	// })

	var reply string

	if err := client.Call(context.Background(), utils.AdminSv1SetResourceProfile,
		&utils.ResourceProfileWithAPIOpts{
			ResourceProfile: &utils.ResourceProfile{
				Tenant:            "cgrates.org",
				ID:                "RES1",
				FilterIDs:         []string{"*string:~*req.Account:1001"},
				Weights:           utils.DynamicWeights{{Weight: 10}},
				UsageTTL:          time.Hour,
				Limit:             3,
				AllocationMessage: "ResourceAllocationSuccess",
				Blocker:           false,
				Stored:            true,
			},
		}, &reply); err != nil {
		t.Fatalf("AdminSv1SetResourceProfile failed: %v", err)
	}

	t.Run("noFlags", func(t *testing.T) {
		var rply V1ProcessEventReply
		if err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "noFlags",
				Event: map[string]any{
					utils.AccountField: "1001",
				},
				APIOpts: map[string]any{
					utils.MetaOriginID: "OriginID",
				},
			}, &rply); err != nil {
			t.Fatal(err)
		}
		if len(rply.ResourceAllocation) != 0 {
			t.Fatalf("expected no allocation, got %v", rply.ResourceAllocation)
		}
	})

	t.Run("resourcesOnly", func(t *testing.T) {
		var rply V1ProcessEventReply
		if err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "resourcesOnly",
				APIOpts: map[string]any{
					utils.MetaResources: true,
					utils.MetaOriginID:  "OriginID",
				},
				Event: map[string]any{
					utils.AccountField: "1001",
				},
			}, &rply); err != nil {
			t.Fatal(err)
		}
		if len(rply.ResourceAllocation) != 0 {
			t.Fatalf("expected no allocation, got %v", rply.ResourceAllocation)
		}
	})

	t.Run("authorizeAndResources", func(t *testing.T) {
		var rply V1ProcessEventReply
		if err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "authorizeAndResources",
				APIOpts: map[string]any{
					utils.MetaAuthorize:        true,
					utils.MetaResources:        true,
					utils.OptsResourcesUsageID: "usage1",
					utils.MetaOriginID:         "OriginID",
				},
				Event: map[string]any{
					utils.AccountField: "1001",
				},
			}, &rply); err != nil {
			t.Fatal(err)
		}
		if msg := rply.ResourceAllocation[utils.MetaPrimary]; msg != "ResourceAllocationSuccess" {
			t.Fatalf("unexpected allocation msg: %q", msg)
		}
	})

	t.Run("resourcesAuthorizeFlag", func(t *testing.T) {
		var rply V1ProcessEventReply
		if err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "resourcesAuthorizeFlag",
				APIOpts: map[string]any{
					utils.MetaResourcesAuthorizeCfg: true,
					utils.OptsResourcesUsageID:      "2",
					utils.OptsResourcesUnits:        1,
					utils.MetaOriginID:              "OriginID",
				},
				Event: map[string]any{
					utils.AccountField: "1001",
				},
			}, &rply); err != nil {
			t.Fatal(err)
		}
		if msg := rply.ResourceAllocation[utils.MetaPrimary]; msg != "ResourceAllocationSuccess" {
			t.Fatalf("unexpected allocation msg: %q", msg)
		}
	})

	t.Run("allocateAndRelease", func(t *testing.T) {
		const usageID = "resource-lifecycle"
		var resource utils.Resource
		if err := client.Call(context.Background(), utils.ResourceSv1GetResource,
			&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "RES1"}},
			&resource); err != nil {
			t.Fatal(err)
		}
		if usage := resource.TotalUsage(); usage != 0 {
			t.Fatalf("resource usage before allocation is %v, want 0", usage)
		}

		var rply V1ProcessEventReply
		if err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "allocateResources",
				APIOpts: map[string]any{
					utils.MetaResourcesAllocateCfg: true,
					utils.OptsResourcesUsageID:     usageID,
					utils.OptsResourcesUnits:       1,
					utils.MetaOriginID:             "OriginID",
				},
				Event: map[string]any{
					utils.AccountField: "1001",
				},
			}, &rply); err != nil {
			t.Fatal(err)
		}
		if msg := rply.ResourceAllocation[utils.MetaPrimary]; msg != "ResourceAllocationSuccess" {
			t.Fatalf("unexpected allocation msg: %q", msg)
		}

		resource = utils.Resource{}
		if err := client.Call(context.Background(), utils.ResourceSv1GetResource,
			&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "RES1"}},
			&resource); err != nil {
			t.Fatal(err)
		}
		if usage := resource.TotalUsage(); usage != 1 {
			t.Fatalf("resource usage after allocation is %v, want 1", usage)
		}

		if err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "releaseResources",
				APIOpts: map[string]any{
					utils.MetaResourcesReleaseCfg: true,
					utils.OptsResourcesUsageID:    usageID,
					utils.MetaOriginID:            "OriginID",
				},
				Event: map[string]any{
					utils.AccountField: "1001",
				},
			}, &rply); err != nil {
			t.Fatal(err)
		}
		resource = utils.Resource{}
		if err := client.Call(context.Background(), utils.ResourceSv1GetResource,
			&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "RES1"}},
			&resource); err != nil {
			t.Fatal(err)
		}
		if usage := resource.TotalUsage(); usage != 0 {
			t.Fatalf("resource usage after release is %v, want 0", usage)
		}
	})

	t.Run("missingUsageID", func(t *testing.T) {
		var rply V1ProcessEventReply
		if err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "missingUsageID",
				APIOpts: map[string]any{
					utils.MetaResourcesAuthorizeCfg: true,
					utils.MetaOriginID:              "OriginID",
				},
				Event: map[string]any{
					utils.AccountField: "1001",
				},
			}, &rply); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if _, exists := rply.ResourceAllocation[utils.MetaPrimary]; !exists {
			t.Fatal("expected allocation entry")
		}
	})

	t.Run("noMatchingProfile", func(t *testing.T) {
		var rply V1ProcessEventReply
		if err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "noMatchingProfile",
				APIOpts: map[string]any{
					utils.MetaResourcesAuthorizeCfg: true,
					utils.OptsResourcesUsageID:      "usage-nomatch",
					utils.MetaOriginID:              "OriginID",
				},
				Event: map[string]any{
					utils.AccountField: "9999",
				},
			}, &rply); err != nil {
			t.Fatal(err)
		}
		if msg := rply.ResourceAllocation[utils.MetaPrimary]; msg != "" {
			t.Fatalf("expected empty allocation msg, got %q", msg)
		}
	})

	t.Run("authorizeWithoutResourcesFlag", func(t *testing.T) {
		var rply V1ProcessEventReply
		if err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "authorizeWithoutResourcesFlag",
				APIOpts: map[string]any{
					utils.MetaAuthorize:        true,
					utils.OptsResourcesUsageID: "usage3",
					utils.MetaOriginID:         "OriginID",
				},
				Event: map[string]any{
					utils.AccountField: "1001",
				},
			}, &rply); err != nil {
			t.Fatal(err)
		}
		if len(rply.ResourceAllocation) != 0 {
			t.Fatalf("expected no allocation without resources flag, got %v", rply.ResourceAllocation)
		}
	})

	t.Run("multipleResourceAllocations", func(t *testing.T) {
		for i := 1; i <= 3; i++ {
			var rply V1ProcessEventReply
			if err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
				&utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     "multipleResourceAllocations",
					APIOpts: map[string]any{
						utils.MetaResourcesAuthorizeCfg: true,
						utils.OptsResourcesUsageID:      utils.ConcatenatedKey("usage-multi", strconv.Itoa(i)),
						utils.OptsResourcesUnits:        1,
						utils.MetaOriginID:              "OriginID",
					},
					Event: map[string]any{
						utils.AccountField: "1001",
					},
				}, &rply); err != nil {
				t.Fatalf("allocation %d failed: %v", i, err)
			}
			if msg := rply.ResourceAllocation[utils.MetaPrimary]; msg != "ResourceAllocationSuccess" {
				t.Fatalf("allocation %d: unexpected msg: %q", i, msg)
			}
		}
	})

	t.Run("resourcesWithZeroUnits", func(t *testing.T) {
		var rply V1ProcessEventReply
		if err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "resourcesWithZeroUnits",
				APIOpts: map[string]any{
					utils.MetaResourcesAuthorizeCfg: true,
					utils.OptsResourcesUsageID:      "usage-zero",
					utils.OptsResourcesUnits:        0,
					utils.MetaOriginID:              "OriginID",
				},
				Event: map[string]any{
					utils.AccountField: "1001",
				},
			}, &rply); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if _, exists := rply.ResourceAllocation[utils.MetaPrimary]; !exists {
			t.Fatal("expected allocation entry")
		}
	})

	t.Run("resourcesAuthorizeCfgWithAuthorize", func(t *testing.T) {
		var rply V1ProcessEventReply
		if err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "resourcesAuthorizeCfgWithAuthorize",
				APIOpts: map[string]any{
					utils.MetaAuthorize:             true,
					utils.MetaResourcesAuthorizeCfg: true,
					utils.OptsResourcesUsageID:      "usage-both-flags",
					utils.MetaOriginID:              "OriginID",
				},
				Event: map[string]any{
					utils.AccountField: "1001",
				},
			}, &rply); err != nil {
			t.Fatal(err)
		}
		if msg := rply.ResourceAllocation[utils.MetaPrimary]; msg != "ResourceAllocationSuccess" {
			t.Fatalf("unexpected allocation msg: %q", msg)
		}
	})
}

func TestSessionSv1ProcessEventAttributesResources(t *testing.T) {
	ng := engine.TestEngine{
		ConfigJSON: `{
"logger": { "level": 7 },
"sessions": {
	"enabled": true,
	"conns": {
		"*attributes": [{"connIDs": ["*localhost"]}],
		"*resources": [{"connIDs": ["*localhost"]}]
	}
},
"attributes": { "enabled": true },
"resources": {
	"enabled": true,
	"storeInterval": "-1"
},
"admins": { "enabled": true }
}`,
		DBCfg:    engine.InternalDBCfg,
		Encoding: *utils.Encoding,
	}
	client, _ := ng.Run(t)

	var reply string
	if err := client.Call(context.Background(), utils.AdminSv1SetAttributeProfile,
		&utils.APIAttributeProfileWithAPIOpts{
			APIAttributeProfile: &utils.APIAttributeProfile{
				Tenant:    "cgrates.org",
				ID:        "ATTR_PREMIUM",
				FilterIDs: []string{"*string:~*req.Account:2001"},
				Weights:   utils.DynamicWeights{{Weight: 10}},
				Attributes: []*utils.ExternalAttribute{
					{
						Path:  "*req.Category",
						Type:  utils.MetaConstant,
						Value: "premium",
					},
				},
			},
		}, &reply); err != nil {
		t.Fatal(err)
	}

	if err := client.Call(context.Background(), utils.AdminSv1SetResourceProfile,
		&utils.ResourceProfileWithAPIOpts{
			ResourceProfile: &utils.ResourceProfile{
				Tenant:            "cgrates.org",
				ID:                "RES_PREMIUM",
				FilterIDs:         []string{"*string:~*req.Category:premium"},
				Weights:           utils.DynamicWeights{{Weight: 10}},
				UsageTTL:          time.Hour,
				Limit:             1,
				AllocationMessage: "AllocatedPremium",
				Stored:            true,
			},
		}, &reply); err != nil {
		t.Fatalf("AdminSv1SetResourceProfile failed: %v", err)
	}

	getUsage := func(t *testing.T) float64 {
		t.Helper()
		var rs utils.Resource
		if err := client.Call(context.Background(), utils.ResourceSv1GetResource,
			&utils.TenantIDWithAPIOpts{
				TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "RES_PREMIUM"},
			}, &rs); err != nil {
			t.Fatal(err)
		}
		return rs.TotalUsage()
	}

	t.Run("resourceMatchesAttributeModifiedEvent", func(t *testing.T) {
		var rply V1ProcessEventReply
		if err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "attrRes-alloc1",
				APIOpts: map[string]any{
					utils.MetaAttributes:           true,
					utils.MetaResourcesAllocateCfg: true,
					utils.OptsResourcesUsageID:     "alloc1",
					utils.OptsResourcesUnits:       1,
					utils.MetaOriginID:             "OriginID",
				},
				Event: map[string]any{
					utils.AccountField: "2001",
				},
			}, &rply); err != nil {
			t.Fatal(err)
		}
		if msg := rply.ResourceAllocation[utils.MetaPrimary]; msg != "AllocatedPremium" {
			t.Errorf("unexpected allocation message: %q", msg)
		}
		if usage := getUsage(t); usage != 1 {
			t.Errorf("RES_PREMIUM usage = %v, want 1", usage)
		}
	})

	t.Run("attributesDisabledNoResourceMatch", func(t *testing.T) {
		var rply V1ProcessEventReply
		if err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "attrRes-noAttrs",
				APIOpts: map[string]any{
					utils.MetaResourcesAllocateCfg: true,
					utils.OptsResourcesUsageID:     "alloc2",
					utils.OptsResourcesUnits:       1,
					utils.MetaOriginID:             "OriginID",
				},
				Event: map[string]any{
					utils.AccountField: "2001",
				},
			}, &rply); err == nil {
			t.Fatal("expected an error when no resource profile matches")
		}
		if usage := getUsage(t); usage != 1 {
			t.Errorf("RES_PREMIUM usage = %v, want 1", usage)
		}
	})

	t.Run("attributeNotMatchedNoResourceMatch", func(t *testing.T) {
		var rply V1ProcessEventReply
		if err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "attrRes-noAccMatch",
				APIOpts: map[string]any{
					utils.MetaAttributes:           true,
					utils.MetaResourcesAllocateCfg: true,
					utils.OptsResourcesUsageID:     "alloc3",
					utils.OptsResourcesUnits:       1,
					utils.MetaOriginID:             "OriginID",
				},
				Event: map[string]any{
					utils.AccountField: "2002",
				},
			}, &rply); err == nil {
			t.Fatal("expected an error when attributes do not match and no resource matches")
		}
		if usage := getUsage(t); usage != 1 {
			t.Errorf("RES_PREMIUM usage = %v, want 1", usage)
		}
	})

	t.Run("allocateOverLimitFails", func(t *testing.T) {
		var rply V1ProcessEventReply
		if err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "attrRes-overLimit",
				APIOpts: map[string]any{
					utils.MetaAttributes:           true,
					utils.MetaResourcesAllocateCfg: true,
					utils.OptsResourcesUsageID:     "alloc4",
					utils.OptsResourcesUnits:       1,
					utils.MetaOriginID:             "OriginID",
				},
				Event: map[string]any{
					utils.AccountField: "2001",
				},
			}, &rply); err == nil {
			t.Fatal("expected an error when allocating over the limit")
		}
		if usage := getUsage(t); usage != 1 {
			t.Errorf("RES_PREMIUM usage = %v, want 1", usage)
		}
	})

	t.Run("cleanup", func(t *testing.T) {
		var rply V1ProcessEventReply
		if err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "attrRes-release",
				APIOpts: map[string]any{
					utils.MetaAttributes:          true,
					utils.MetaResourcesReleaseCfg: true,
					utils.OptsResourcesUsageID:    "alloc1",
					utils.OptsResourcesUnits:      1,
					utils.MetaOriginID:            "OriginID",
				},
				Event: map[string]any{
					utils.AccountField: "2001",
				},
			}, &rply); err != nil {
			t.Fatal(err)
		}
		if usage := getUsage(t); usage != 0 {
			t.Errorf("RES_PREMIUM usage after release = %v, want 0", usage)
		}
	})
}
