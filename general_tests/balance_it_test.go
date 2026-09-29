//go:build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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

func TestVoiceBalanceMonetary(t *testing.T) {
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
"schedulers": {
    "enabled": true
},
"apiers": {
    "enabled": true,
    "scheduler_conns": ["*internal"]
}
}`

	tpFiles := map[string]string{
		utils.DestinationsCsv: `#Id,Prefix
DST1,1002`,
		utils.RatesCsv: `#Id,ConnectFee,Rate,RateUnit,RateIncrement,GroupIntervalStart
RT_10PM,0,10,60s,1s,0`,
		utils.DestinationRatesCsv: `#Id,DestinationId,RatesTag,RoundingMethod,RoundingDecimals,MaxCost,MaxCostStrategy
DR_10PM,DST1,RT_10PM,*up,4,0,`,
		utils.RatingPlansCsv: `#Id,DestinationRatesId,TimingTag,Weight
RP_10PM,DR_10PM,*any,10`,
		utils.RatingProfilesCsv: `#Tenant,Category,Subject,ActivationTime,RatingPlanId,RatesFallbackSubject
cgrates.org,call,Tier1,,RP_10PM,
cgrates.org,call,1001,,RP_10PM,`,
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
		Value:       10,
		Balance: map[string]any{
			utils.ID: "MAIN",
		},
	}, &reply); err != nil {
		t.Fatal(err)
	}

	if err := client.Call(utils.APIerSv1SetBalance, &utils.AttrSetBalance{
		Tenant:      "cgrates.org",
		Account:     "1001",
		BalanceType: utils.VOICE,
		Value:       float64(5 * time.Minute),
		Balance: map[string]any{
			utils.ID:            "VOICE",
			utils.RatingSubject: "Tier1",
		},
	}, &reply); err != nil {
		t.Fatal(err)
	}

	t.Run("MaxDebit", func(t *testing.T) {
		tStart := time.Date(2026, time.September, 28, 8, 0, 0, 0, time.UTC)
		cd := &engine.CallDescriptor{
			Category:      "call",
			Tenant:        "cgrates.org",
			Subject:       "1001",
			Account:       "1001",
			Destination:   "1002",
			ToR:           utils.VOICE,
			TimeStart:     tStart,
			TimeEnd:       tStart.Add(30 * time.Second),
			DurationIndex: 30 * time.Second,
		}
		var cc engine.CallCost
		if err := client.Call(utils.ResponderMaxDebit, cd, &cc); err != nil {
			t.Fatal(err)
		}
		if got := cc.GetDuration(); got != 30*time.Second {
			t.Errorf("granted usage = %v, want 30s", got)
		}
		if cost := utils.Round(cc.Cost, 2, utils.ROUNDING_MIDDLE); cost != 5 {
			t.Errorf("cost = %v, want 5", cost)
		}
	})

	t.Run("CheckBalances", func(t *testing.T) {
		var acnt engine.Account
		if err := client.Call(utils.APIerSv2GetAccount,
			&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}, &acnt); err != nil {
			t.Fatal(err)
		}
		voice := time.Duration(acnt.BalanceMap[utils.VOICE][0].Value)
		if want := 4*time.Minute + 30*time.Second; voice != want {
			t.Errorf("remaining voice = %v, want %v", voice, want)
		}
		if money := utils.Round(acnt.BalanceMap[utils.MONETARY][0].Value, 2, utils.ROUNDING_MIDDLE); money != 5 {
			t.Errorf("remaining money = %v, want 5", money)
		}
	})
}

func TestVoiceBalanceWithoutMonetary(t *testing.T) {
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
"schedulers": {
	"enabled": true
},
"apiers": {
	"enabled": true,
	"scheduler_conns": ["*internal"]
}
}`

	tpFiles := map[string]string{
		utils.DestinationsCsv: `#Id,Prefix
DST1,1002`,
		utils.RatesCsv: `#Id,ConnectFee,Rate,RateUnit,RateIncrement,GroupIntervalStart
RT_10PM,0,10,60s,1s,0`,
		utils.DestinationRatesCsv: `#Id,DestinationId,RatesTag,RoundingMethod,RoundingDecimals,MaxCost,MaxCostStrategy
DR_10PM,DST1,RT_10PM,*up,4,0,`,
		utils.RatingPlansCsv: `#Id,DestinationRatesId,TimingTag,Weight
RP_10PM,DR_10PM,*any,10`,
		utils.RatingProfilesCsv: `#Tenant,Category,Subject,ActivationTime,RatingPlanId,RatesFallbackSubject
cgrates.org,call,Tier1,,RP_10PM,
cgrates.org,call,1001,,RP_10PM,`,
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
		Value:       0,
		Balance: map[string]any{
			utils.ID: "MAIN",
		},
	}, &reply); err != nil {
		t.Fatal(err)
	}

	if err := client.Call(utils.APIerSv1SetBalance, &utils.AttrSetBalance{
		Tenant:      "cgrates.org",
		Account:     "1001",
		BalanceType: utils.VOICE,
		Value:       float64(5 * time.Minute),
		Balance: map[string]any{
			utils.ID:            "VOICE",
			utils.RatingSubject: "Tier1",
		},
	}, &reply); err != nil {
		t.Fatal(err)
	}

	t.Run("MaxDebit", func(t *testing.T) {
		tStart := time.Date(2026, time.September, 28, 8, 0, 0, 0, time.UTC)
		cd := &engine.CallDescriptor{
			Category:      "call",
			Tenant:        "cgrates.org",
			Subject:       "1001",
			Account:       "1001",
			Destination:   "1002",
			ToR:           utils.VOICE,
			TimeStart:     tStart,
			TimeEnd:       tStart.Add(30 * time.Second),
			DurationIndex: 30 * time.Second,
		}
		var cc engine.CallCost
		if err := client.Call(utils.ResponderMaxDebit, cd, &cc); err != nil {
			t.Fatal(err)
		}
		if got := cc.GetDuration(); got != 0 {
			t.Errorf("usage = %v, want 0", got)
		}
		if cc.Cost != 0 {
			t.Errorf("cost = %v, want 0", cc.Cost)
		}
	})

	t.Run("CheckBalances", func(t *testing.T) {
		var acnt engine.Account
		if err := client.Call(utils.APIerSv2GetAccount,
			&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}, &acnt); err != nil {
			t.Fatal(err)
		}
		voice := time.Duration(acnt.BalanceMap[utils.VOICE][0].Value)
		if want := 5 * time.Minute; voice != want {
			t.Errorf("remaining voice = %v, want %v", voice, want)
		}
		if monetary := acnt.BalanceMap[utils.MONETARY][0].Value; monetary != 0 {
			t.Errorf("remaining monetary = %v, want 0", monetary)
		}
	})
}
