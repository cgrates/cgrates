//go:build integration

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
package general_tests

import (
	"testing"
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestBalanceBlocker(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

	content := `{

"general": {
	"log_level": 7
},
"data_db": {
	"db_type": "*internal"
},
"stor_db": {
	"db_type": "*internal"
},
"rals": {
	"enabled": true
},
"cdrs": {
	"enabled": true,
	"rals_conns": ["*internal"]
},
"schedulers": {
	"enabled": true
},
"apiers": {
	"enabled": true,
	"scheduler_conns": ["*internal"]
}
}`

	tpFiles := map[string]string{
		utils.DestinationRatesCsv: `#Id,DestinationId,RatesTag,RoundingMethod,RoundingDecimals,MaxCost,MaxCostStrategy
DR_ANY,*any,RT_ANY,*up,20,0,`,
		utils.RatesCsv: `#Id,ConnectFee,Rate,RateUnit,RateIncrement,GroupIntervalStart
RT_ANY,0.4,0.1,1s,1s,0`,
		utils.RatingPlansCsv: `#Id,DestinationRatesId,TimingTag,Weight
RP_ANY,DR_ANY,*any,10`,
		utils.RatingProfilesCsv: `#Tenant,Category,Subject,ActivationTime,RatingPlanId,RatesFallbackSubject
cgrates.org,call,1001,,RP_ANY,`,
	}

	testEnv := TestEnvironment{
		ConfigJSON: content,
		TpFiles:    tpFiles,
	}
	client, _ := testEnv.Setup(t, 0)

	var reply string
	if err := client.Call(utils.APIerSv1SetBalance, &utils.AttrSetBalance{
		Tenant:      "cgrates.org",
		Account:     "1001",
		BalanceType: utils.MONETARY,
		Value:       1,
		Balance: map[string]any{
			utils.ID:      "test",
			utils.Blocker: true,
		},
	}, &reply); err != nil {
		t.Fatal(err)
	}

	// Attempt to debit 1.2, but due to the blocker, only 1 unit can be debited.
	t.Run("ProcessCDR1", func(t *testing.T) {
		var reply string
		if err := client.Call(utils.CDRsV1ProcessEvent,
			&engine.ArgV1ProcessEvent{
				Flags: []string{utils.MetaRALs},
				CGREvent: utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     "event1",
					Event: map[string]any{
						utils.RunID:       "*default",
						utils.Tenant:      "cgrates.org",
						utils.Category:    "call",
						utils.ToR:         utils.VOICE,
						utils.OriginID:    "processCDR1",
						utils.OriginHost:  "127.0.0.1",
						utils.RequestType: utils.POSTPAID,
						utils.Account:     "1001",
						utils.Destination: "1002",
						utils.SetupTime:   time.Date(2021, time.February, 2, 16, 14, 50, 0, time.UTC),
						utils.AnswerTime:  time.Date(2021, time.February, 2, 16, 15, 0, 0, time.UTC),
						utils.Usage:       "8s",
					},
				},
			}, &reply); err != nil {
			t.Fatal(err)
		}
	})

	// Attempt to debit 0.7, but due to the blocker and the empty balance, the final cost should be -1.
	t.Run("ProcessCDR2", func(t *testing.T) {
		var reply string
		if err := client.Call(utils.CDRsV1ProcessEvent,
			&engine.ArgV1ProcessEvent{
				Flags: []string{utils.MetaRALs},
				CGREvent: utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     "event1",
					Event: map[string]any{
						utils.RunID:       "*default",
						utils.Tenant:      "cgrates.org",
						utils.Category:    "call",
						utils.ToR:         utils.VOICE,
						utils.OriginID:    "processCDR2",
						utils.OriginHost:  "127.0.0.1",
						utils.RequestType: utils.POSTPAID,
						utils.Account:     "1001",
						utils.Destination: "1002",
						utils.SetupTime:   time.Date(2021, time.February, 2, 16, 14, 50, 0, time.UTC),
						utils.AnswerTime:  time.Date(2021, time.February, 2, 16, 15, 0, 0, time.UTC),
						utils.Usage:       "3s",
					},
				},
			}, &reply); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("check CDRs", func(t *testing.T) {
		var cdrs []*engine.CDR
		if err := client.Call(utils.CDRsV1GetCDRs, &utils.RPCCDRsFilterWithArgDispatcher{
			RPCCDRsFilter: &utils.RPCCDRsFilter{
				OrderBy: utils.Usage + ";desc",
			}}, &cdrs); err != nil {
			t.Fatal(err)
		}
		if len(cdrs) != 2 {
			t.Fatalf("expected to receive 2 CDRs: %v", utils.ToJSON(cdrs))
		}
		if cdrs[0].Cost != -1 {
			t.Fatalf("expected cost to be -1, received <%v>", utils.ToJSON(cdrs[0]))
		}
		cost := 0.0
		if cdrs[1].CostDetails.Cost != nil {
			cost = *cdrs[1].CostDetails.Cost
		}
		if cost != 0.7 {
			t.Errorf("cdrs[1].CostDetails.Cost = %v, want %v", cost, 0.7)
		}
		balanceSummaries := cdrs[1].CostDetails.AccountSummary.BalanceSummaries
		if len(balanceSummaries) != 1 {
			t.Errorf("expected only 1 balance summary inside second CDR, got %s", utils.ToJSON(balanceSummaries))
		}
	})
}
