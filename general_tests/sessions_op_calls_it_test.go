//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package general_tests

import (
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

func TestSessionSOpCalls(t *testing.T) {
	ng := engine.TestEngine{
		ConfigJSON: `{
"admins": {
	"enabled": true,
},
"db": {
	"dbConns": {
		"*default": {
			"dbType": "*internal",
			"opts":{
			"internalDBDumpInterval": "0",	
			"internalDBRewriteInterval": "0"
	}
    	}
	},
},
"attributes": {
	"enabled": true,
},
"resources": {
	"enabled": true,
},
"rates": {
	"enabled": true,
	 "ratePrefixIndexedFields": ["*req.Destination"]
},
"accounts": {
	"enabled": true,
	"conns": {
		"*rates": [{"connIDs": ["*localhost"]}]
	}
},
"sessions": {
	"enabled": true,
	"conns":  {
		"*attributes": [{"connIDs": ["*localhost"]}],
		"*resources": [{"connIDs": ["*localhost"]}],
		"*accounts": [{"connIDs": ["*localhost"]}]
	}
}

}`,
		TpFiles: map[string]string{
			utils.AttributesCsv: `#Tenant,ID,FilterIDs,Weights,Blockers,AttributeFilterIDs,AttributeBlockers,Path,Type,Value
cgrates.org,ATTR_1001,*string:~*req.Account:1001,;10,,,,*req.Category,*constant,Customer`,
			utils.ResourcesCsv: `#Tenant[0],Id[1],FilterIDs[2],Weights[3],TTL[4],Limit[5],AllocationMessage[6],Blocker[7],Stored[8],ThresholdIDs[9]
cgrates.org,RES_1001,*string:~*req.Account:1001,;10,-1,2,,false,true,*none`,
			utils.AccountsCsv: `#Tenant,ID,FilterIDs,Weights,Blockers,Opts,BalanceID,BalanceFilterIDs,BalanceWeights,BalanceBlockers,BalanceType,BalanceUnits,BalanceUnitFactors,BalanceOpts,BalanceCostIncrements,BalanceAttributeIDs,BalanceRateProfileIDs,ThresholdIDs
cgrates.org,1001,*string:~*req.Account:1001,;10,,,MONETARY,,;10,,*concrete,10,,,;1s;;,,RP_CUSTOMER,`,
			utils.RatesCsv: `#Tenant,ID,FilterIDs,Weights,MinCost,MaxCost,MaxCostStrategy,RateID,RateFilterIDs,RateActivationStart,RateWeights,RateBlocker,RateIntervalStart,RateFixedFee,RateRecurrentFee,RateUnit,RateIncrement
cgrates.org,RP_CUSTOMER,*string:~*req.Category:Customer,;10,,,,RT_DE,*prefix:~*req.Destination:+49,,,,0s,,0.6,1m,1s
cgrates.org,RP_CUSTOMER,,,,,,RT_RO,*prefix:~*req.Destination:+40,,,,0s,,0.3,1m,1s`,
		},
		DBCfg:    engine.InternalDBCfg,
		Encoding: *utils.Encoding,
	}
	client, _ := ng.Run(t)

	t.Run("Authorize", func(t *testing.T) {
		var rply sessions.V1ProcessEventReply
		if err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     utils.GenUUID(),
				APIOpts: map[string]any{
					utils.MetaOriginID:         "originID",
					utils.MetaAuthorize:        true,
					utils.MetaAttributes:       true,
					utils.MetaAccounts:         true,
					utils.MetaResources:        true,
					utils.OptsResourcesUsageID: "usageID",
					utils.OptsResourcesUnits:   1,
					utils.MetaUsage:            2 * time.Minute,
				},
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Subject:      "1001",
					utils.Destination:  "+4986517174963",
				},
			}, &rply); err != nil {
			t.Fatal(err)
		}
		if len(rply.Attributes[utils.MetaPrimary].AlteredFields) != 1 ||
			rply.Attributes[utils.MetaPrimary].AlteredFields[0].MatchedProfileID != "cgrates.org:ATTR_1001" {
			t.Errorf("received attributes reply: %s", utils.ToJSON(rply.Attributes))
		}
		if rply.ResourceAllocation[utils.MetaPrimary] != "RES_1001" {
			t.Errorf("received resource allocation: %v", rply.ResourceAllocation)
		}
		if rply.AccountsUsage[utils.MetaPrimary] != 2*time.Minute {
			t.Errorf("received accounts usage: %v", rply.AccountsUsage)
		}
	})

	t.Run("Start", func(t *testing.T) {
		var rply sessions.V1ProcessEventReply
		if err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     utils.GenUUID(),
				APIOpts: map[string]any{
					utils.MetaOriginID:             "originID",
					utils.MetaSession:              true,
					utils.MetaAttributes:           true,
					utils.MetaAccounts:             true,
					utils.MetaDebit:                true,
					utils.MetaResourcesAllocateCfg: true,
					utils.OptsResourcesUsageID:     "usageID",
					utils.OptsResourcesUnits:       1,
					utils.MetaUsage:                time.Minute,
				},
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Subject:      "1001",
					utils.Destination:  "+4986517174963",
				},
			}, &rply); err != nil {
			t.Fatal(err)
		}
		if rply.ResourceAllocation[utils.MetaPrimary] != "RES_1001" {
			t.Errorf("received resource allocation: %v", rply.ResourceAllocation)
		}
		if rply.AccountsUsage[utils.MetaPrimary] != time.Minute {
			t.Errorf("received accounts usage: %v", rply.AccountsUsage)
		}

		var aSessions []*sessions.ExternalSession
		if err := client.Call(context.Background(), utils.SessionSv1GetActiveSessions,
			new(utils.SessionFilter), &aSessions); err != nil {
			t.Fatal(err)
		}
		if len(aSessions) != 1 {
			t.Errorf("received active sessions: %s", utils.ToJSON(aSessions))
		}

		var resource utils.Resource
		if err := client.Call(context.Background(), utils.ResourceSv1GetResource,
			&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "RES_1001"}},
			&resource); err != nil {
			t.Fatal(err)
		}
		if resusage := resource.TotalUsage(); resusage != 1 {
			t.Errorf("expected resource usage 1, received %v", resusage)
		}

		var acnt utils.Account
		if err := client.Call(context.Background(), utils.AdminSv1GetAccount,
			&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "1001"}},
			&acnt); err != nil {
			t.Fatal(err)
		}
		if balUnits := acnt.Balances["MONETARY"].Units; balUnits.Compare(utils.NewDecimalFromFloat64(9.4)) != 0 {
			t.Errorf("expected balance 9.4, received %v", balUnits)
		}
	})

	t.Run("Terminate", func(t *testing.T) {
		var rply sessions.V1ProcessEventReply
		if err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     utils.GenUUID(),
				APIOpts: map[string]any{
					utils.MetaOriginID:            "originID",
					utils.MetaSession:             true,
					utils.MetaTerminate:           true,
					utils.MetaAttributes:          true,
					utils.MetaAccounts:            true,
					utils.MetaDebit:               true,
					utils.MetaResourcesReleaseCfg: true,
					utils.OptsResourcesUsageID:    "usageID",
					utils.MetaUsage:               30 * time.Second,
					utils.MetaUR:                  true,
				},
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Subject:      "1001",
					utils.Destination:  "+4986517174963",
				},
			}, &rply); err != nil {
			t.Fatal(err)
		}
		if rply.AccountsUsage[utils.MetaPrimary] != 90*time.Second {
			t.Errorf("unexpected accounts usage: %v", rply.AccountsUsage)
		}
		ur, has := rply.UsageRecords[utils.MetaPrimary]
		if !has {
			t.Fatalf("no ur: %s", utils.ToJSON(rply))
		}
		if ur.Event[utils.Category] != "Customer" {
			t.Errorf("unexpected Category: %v", ur.Event[utils.Category])
		}
		accountsCost, err := utils.IfaceAsEventCharges(ur.APIOpts[utils.MetaAccountsCost])
		if err != nil {
			t.Fatal(err)
		}
		if accountsCost.Concretes.Compare(utils.NewDecimalFromFloat64(0.9)) != 0 {
			t.Errorf("expected cost 0.9, received %v", accountsCost.Concretes)
		}

		var resource utils.Resource
		if err := client.Call(context.Background(), utils.ResourceSv1GetResource,
			&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "RES_1001"}},
			&resource); err != nil {
			t.Fatal(err)
		}

		if resusage := resource.TotalUsage(); resusage != 0 {
			t.Errorf("expected 0, received %v", resusage)
		}

		var acnt utils.Account
		if err := client.Call(context.Background(), utils.AdminSv1GetAccount,
			&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "1001"}},
			&acnt); err != nil {
			t.Fatal(err)
		}
		if units := acnt.Balances["MONETARY"].Units; units.Compare(utils.NewDecimalFromFloat64(9.1)) != 0 {
			t.Errorf("expected balance 9.1, got %v", units)
		}
	})
}
