//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package sessions

import (
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestSessionSv1ProcessEventDynamicRates(t *testing.T) {

	ng := engine.TestEngine{
		ConfigJSON: `{
"sessions": {
    "enabled": true,
    "conns": {
    	"*rates": [{"connIDs": ["*localhost"]}]
    },
    "opts": {
        "*rates": [
            {
                "filterIDs": ["*string:~*req.Destination:1002"],
                "value": true
            }
        ]
    }
},
"rates": {
    "enabled": true
},
"admins": {
    "enabled": true
},
"configs": {
    "enabled": true
}
}`,
		TpFiles: map[string]string{
			utils.RatesCsv: `#Tenant,ID,FilterIDs,Weights,MinCost,MaxCost,MaxCostStrategy,RateID,RateFilterIDs,RateActivationStart,RateWeights,RateBlocker,RateIntervalStart,RateFixedFee,RateRecurrentFee,RateUnit,RateIncrement
cgrates.org,RP_SIMPLE,,;10,,,,RT_SIMPLE,*string:~*req.Destination:1002,"* * * * *",;10,false,0s,0,1,1m,1m`,
		},
		DBCfg:    engine.InternalDBCfg,
		Encoding: *utils.Encoding,
		// LogBuffer: new(bytes.Buffer),

	}

	// t.Cleanup(func() {
	// 	if ng.LogBuffer != nil {
	// 		fmt.Println(ng.LogBuffer)
	// 	}
	// })

	client, _ := ng.Run(t)
	time.Sleep(500 * time.Millisecond)

	t.Run("dynamicMatch", func(t *testing.T) {
		var rply V1ProcessEventReply
		err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "dynMatch",
				APIOpts: map[string]any{
					utils.MetaUsage:    1 * time.Minute,
					utils.MetaOriginID: "OriginID",
				},
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  "1002",
					utils.AnswerTime:   "2018-01-07T17:00:00Z",
				},
			}, &rply)
		if err != nil {
			t.Fatalf("ProcessEvent failed: %v", err)
		}

		if rply.RateSCost == nil {
			t.Fatal("RateSCost should not be nil when dynamic filter matches")
		}

		cost, exists := rply.RateSCost[utils.MetaPrimary]
		if !exists {
			t.Fatalf("no RateSCost entry for *primary runID, got: %v", rply.RateSCost)
		}

		if cost != 1.0 {
			t.Errorf("RateSCost[*primary] = %g, want 1.0", cost)
		}
	})

	t.Run("dynamicNoMatch", func(t *testing.T) {
		var rply V1ProcessEventReply
		err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "dynNoMatch",
				APIOpts: map[string]any{
					utils.MetaUsage:    1 * time.Minute,
					utils.MetaOriginID: "OriginID",
				},
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  "9999",
					utils.AnswerTime:   "2018-01-07T17:00:00Z",
				},
			}, &rply)
		if err != nil {
			t.Fatalf("ProcessEvent failed: %v", err)
		}

		if len(rply.RateSCost) > 0 {
			t.Errorf("RateSCost should be empty when filter does not match, got: %v", rply.RateSCost)
		}
	})
}
