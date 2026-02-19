//go:build integration
// +build integration

/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package sessions

import (
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestSessionSv1ProcessEventDynamicRates(t *testing.T) {
	var dbcfg engine.DBCfg
	switch *utils.DBType {
	case utils.MetaInternal:
		dbcfg = engine.InternalDBCfg
	case utils.MetaRedis:
		dbcfg = engine.RedisDBCfg
	case utils.MetaMySQL:
		dbcfg = engine.MySQLDBCfg
	case utils.MetaMongo:
		dbcfg = engine.MongoDBCfg
	case utils.MetaPostgres:
		dbcfg = engine.PostgresDBCfg
	default:
		t.Fatal("unsupported dbtype value")
	}

	ng := engine.TestEngine{
		ConfigJSON: `{
"sessions": {
    "enabled": true,
    "rates_conns": ["*localhost"],
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
		DBCfg:    dbcfg,
		Encoding: *utils.Encoding,
	}

	client, _ := ng.Run(t)
	time.Sleep(100 * time.Millisecond)

	t.Run("dynamicMatch", func(t *testing.T) {
		var rply V1ProcessEventReply
		err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "dynMatch",
				APIOpts: map[string]any{
					utils.MetaUsage: 1 * time.Minute,
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

		cost, exists := rply.RateSCost[utils.MetaDefault]
		if !exists {
			t.Fatalf("no RateSCost entry for *default runID, got: %v", rply.RateSCost)
		}

		if cost != 1.0 {
			t.Errorf("RateSCost[*default] = %g, want 1.0", cost)
		}
	})

	t.Run("dynamicNoMatch", func(t *testing.T) {
		var rply V1ProcessEventReply
		err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "dynNoMatch",
				APIOpts: map[string]any{
					utils.MetaUsage: 1 * time.Minute,
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
