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
	"errors"
	"math"
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestRerateCDRs(t *testing.T) {
	var cfgDir string
	switch *utils.DBType {
	case utils.MetaInternal:
		cfgDir = "rerate_cdrs_internal"
	case utils.MetaMySQL:
		cfgDir = "rerate_cdrs_mysql"
	case utils.MetaMongo:
		cfgDir = "rerate_cdrs_mongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	ng := engine.TestEngine{
		ConfigPath: path.Join(*utils.DataDir, "conf", "samples", cfgDir),
		TpPath:     path.Join(*utils.DataDir, "tariffplans", "reratecdrs"),
	}
	client, _ := ng.Run(t)

	CGRID := utils.GenUUID()

	t.Run("SetBalance", func(t *testing.T) {
		var reply string
		if err := client.Call(context.Background(), utils.APIerSv2SetBalance,
			utils.AttrSetBalance{
				Tenant:      "cgrates.org",
				Account:     "1001",
				Value:       float64(time.Minute),
				BalanceType: utils.MetaVoice,
				Balance: map[string]any{
					utils.ID: "voiceBalance1",
				},
			}, &reply); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("CheckInitialBalance", func(t *testing.T) {
		expAcnt := engine.Account{
			ID: "cgrates.org:1001",
			BalanceMap: map[string]engine.Balances{
				utils.MetaVoice: {
					{
						ID:    "voiceBalance1",
						Value: float64(time.Minute),
					},
				},
			},
		}
		var acnt engine.Account
		attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
			t.Fatal(err)
		} else {
			expAcnt.UpdateTime = acnt.UpdateTime
			expAcnt.BalanceMap[utils.MetaVoice][0].Uuid = acnt.BalanceMap[utils.MetaVoice][0].Uuid
			if !reflect.DeepEqual(acnt, expAcnt) {
				t.Fatalf("expected: <%+v>,\nreceived: <%+v>", utils.ToJSON(expAcnt), utils.ToJSON(acnt))
			}
		}
	})

	t.Run("ProcessFirstCDR", func(t *testing.T) {
		var reply string
		err := client.Call(context.Background(), utils.CDRsV1ProcessEvent,
			&engine.ArgV1ProcessEvent{
				Flags: []string{utils.MetaRALs},
				CGREvent: utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     "event1",
					Event: map[string]any{
						utils.RunID:        "run_1",
						utils.CGRID:        CGRID,
						utils.Tenant:       "cgrates.org",
						utils.Category:     "call",
						utils.ToR:          utils.MetaVoice,
						utils.OriginID:     "processCDR1",
						utils.OriginHost:   "OriginHost1",
						utils.RequestType:  utils.MetaPseudoPrepaid,
						utils.AccountField: "1001",
						utils.Destination:  "1002",
						utils.SetupTime:    time.Date(2021, time.February, 2, 16, 14, 50, 0, time.UTC),
						utils.AnswerTime:   time.Date(2021, time.February, 2, 16, 15, 0, 0, time.UTC),
						utils.Usage:        2 * time.Minute,
					},
				},
			}, &reply)
		if err != nil {
			t.Fatal(err)
		}

		var cdrs []*engine.CDR
		err = client.Call(context.Background(), utils.CDRsV1GetCDRs, &utils.RPCCDRsFilterWithAPIOpts{
			RPCCDRsFilter: &utils.RPCCDRsFilter{
				RunIDs: []string{"run_1"},
			}}, &cdrs)
		if err != nil {
			t.Fatal(err)
		}
		if cdrs[0].Usage != 2*time.Minute {
			t.Errorf("expected usage to be <%+v>, received <%+v>", 2*time.Minute, cdrs[0].Usage)
		} else if cdrs[0].Cost != 0.6 {
			t.Errorf("expected cost to be <%+v>, received <%+v>", 0.6, cdrs[0].Cost)
		}
	})

	t.Run("CheckAccountBalancesAfterFirstProcessCDR", func(t *testing.T) {
		expAcnt := engine.Account{
			ID: "cgrates.org:1001",
			BalanceMap: map[string]engine.Balances{
				utils.MetaVoice: {
					{
						ID:    "voiceBalance1",
						Value: 0,
					},
				},
				utils.MetaMonetary: {
					{
						ID:    utils.MetaDefault,
						Value: -0.6,
					},
				},
			},
		}
		var acnt engine.Account
		attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
			t.Fatal(err)
		}

		expAcnt.UpdateTime = acnt.UpdateTime
		expAcnt.BalanceMap[utils.MetaVoice][0].Uuid = acnt.BalanceMap[utils.MetaVoice][0].Uuid
		expAcnt.BalanceMap[utils.MetaMonetary][0].Uuid = acnt.BalanceMap[utils.MetaMonetary][0].Uuid
		acnt.BalanceMap[utils.MetaMonetary][0].Value = math.Round(acnt.BalanceMap[utils.MetaMonetary][0].Value*10) / 10
		if !reflect.DeepEqual(acnt, expAcnt) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ToJSON(expAcnt), utils.ToJSON(acnt))
		}
	})

	t.Run("ProcessSecondCDR", func(t *testing.T) {
		var reply string
		err := client.Call(context.Background(), utils.CDRsV1ProcessEvent,
			&engine.ArgV1ProcessEvent{
				Flags: []string{utils.MetaRALs},
				CGREvent: utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     "event2",
					Event: map[string]any{
						utils.RunID:        "run_2",
						utils.CGRID:        CGRID,
						utils.Tenant:       "cgrates.org",
						utils.Category:     "call",
						utils.ToR:          utils.MetaVoice,
						utils.OriginID:     "processCDR2",
						utils.OriginHost:   "OriginHost2",
						utils.RequestType:  utils.MetaPseudoPrepaid,
						utils.AccountField: "1001",
						utils.Destination:  "1002",
						utils.SetupTime:    time.Date(2021, time.February, 2, 15, 14, 50, 0, time.UTC),
						utils.AnswerTime:   time.Date(2021, time.February, 2, 15, 15, 0, 0, time.UTC),
						utils.Usage:        2 * time.Minute,
					},
				},
			}, &reply)
		if err != nil {
			t.Fatal(err)
		}
		var cdrs []*engine.CDR
		err = client.Call(context.Background(), utils.CDRsV1GetCDRs, &utils.RPCCDRsFilterWithAPIOpts{
			RPCCDRsFilter: &utils.RPCCDRsFilter{
				RunIDs: []string{"run_2"},
			}}, &cdrs)
		if err != nil {
			t.Fatal(err)
		}

		if cdrs[0].Usage != 2*time.Minute {
			t.Errorf("expected usage to be <%+v>, received <%+v>", 2*time.Minute, cdrs[0].Usage)
		} else if cdrs[0].Cost != 1.2 {
			t.Errorf("expected cost to be <%+v>, received <%+v>", 1.2, cdrs[0].Cost)
		}
	})

	t.Run("CheckAccountBalancesAfterSecondProcessCDR", func(t *testing.T) {
		expAcnt := engine.Account{
			ID: "cgrates.org:1001",
			BalanceMap: map[string]engine.Balances{
				utils.MetaVoice: {
					{
						ID:    "voiceBalance1",
						Value: 0,
					},
				},
				utils.MetaMonetary: {
					{
						ID:    utils.MetaDefault,
						Value: -1.8,
					},
				},
			},
		}
		var acnt engine.Account
		attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
		err := client.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt)
		if err != nil {
			t.Fatal(err)
		}

		expAcnt.UpdateTime = acnt.UpdateTime
		expAcnt.BalanceMap[utils.MetaVoice][0].Uuid = acnt.BalanceMap[utils.MetaVoice][0].Uuid
		expAcnt.BalanceMap[utils.MetaMonetary][0].Uuid = acnt.BalanceMap[utils.MetaMonetary][0].Uuid
		acnt.BalanceMap[utils.MetaMonetary][0].Value = math.Round(acnt.BalanceMap[utils.MetaMonetary][0].Value*10) / 10
		if !reflect.DeepEqual(acnt, expAcnt) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ToJSON(expAcnt), utils.ToJSON(acnt))
		}
	})

	t.Run("RerateCDRs", func(t *testing.T) {
		var reply string
		if err := client.Call(context.Background(), utils.CDRsV1RateCDRs, &engine.ArgRateCDRs{
			Flags: []string{utils.MetaRerate},
			RPCCDRsFilter: utils.RPCCDRsFilter{
				OrderBy: utils.AnswerTime,
				CGRIDs:  []string{CGRID},
			}}, &reply); err != nil {
			t.Fatal(err)
		}

		var cdrs []*engine.CDR
		err := client.Call(context.Background(), utils.CDRsV1GetCDRs, &utils.RPCCDRsFilterWithAPIOpts{
			RPCCDRsFilter: &utils.RPCCDRsFilter{
				CGRIDs:  []string{CGRID},
				OrderBy: utils.AnswerTime,
			}}, &cdrs)
		if err != nil {
			t.Fatal(err)
		}

		if cdrs[0].Cost != 0.6 {
			t.Errorf("expected cost to be <%+v>, received <%+v>", 0.6, cdrs[0].Cost)
		} else if cdrs[1].Cost != 1.2 {
			t.Errorf("expected cost to be <%+v>, received <%+v>", 1.2, cdrs[1].Cost)
		}
	})

	t.Run("CheckAccountBalancesAfterRerate", func(t *testing.T) {
		expAcnt := engine.Account{
			ID: "cgrates.org:1001",
			BalanceMap: map[string]engine.Balances{
				utils.MetaVoice: {
					{
						ID:    "voiceBalance1",
						Value: 0,
					},
				},
				utils.MetaMonetary: {
					{
						ID:    utils.MetaDefault,
						Value: -1.8,
					},
				},
			},
		}
		var acnt engine.Account
		attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
		err := client.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt)
		if err != nil {
			t.Fatal(err)
		}

		expAcnt.UpdateTime = acnt.UpdateTime
		expAcnt.BalanceMap[utils.MetaVoice][0].Uuid = acnt.BalanceMap[utils.MetaVoice][0].Uuid
		expAcnt.BalanceMap[utils.MetaMonetary][0].Uuid = acnt.BalanceMap[utils.MetaMonetary][0].Uuid
		acnt.BalanceMap[utils.MetaMonetary][0].Value = math.Round(acnt.BalanceMap[utils.MetaMonetary][0].Value*10) / 10
		if !reflect.DeepEqual(acnt, expAcnt) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ToJSON(expAcnt), utils.ToJSON(acnt))
		}
	})
}

func TestRerateCDRsNoRefund(t *testing.T) {
	var cfgDir string
	switch *utils.DBType {
	case utils.MetaInternal:
		cfgDir = "rerate_cdrs_internal"
	case utils.MetaMySQL:
		cfgDir = "rerate_cdrs_mysql"
	case utils.MetaMongo:
		cfgDir = "rerate_cdrs_mongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	ng := engine.TestEngine{
		ConfigPath: path.Join(*utils.DataDir, "conf", "samples", cfgDir),
		TpPath:     path.Join(*utils.DataDir, "tariffplans", "reratecdrs"),
	}
	client, _ := ng.Run(t)
	CGRID := utils.GenUUID()
	t.Run("ProcessFirstCDR", func(t *testing.T) {
		var reply string
		err := client.Call(context.Background(), utils.CDRsV1ProcessEvent,
			&engine.ArgV1ProcessEvent{
				Flags: []string{utils.MetaRALs},
				CGREvent: utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     "event1",
					Event: map[string]any{
						utils.RunID:        "run_1",
						utils.CGRID:        CGRID,
						utils.Tenant:       "cgrates.org",
						utils.Category:     "call",
						utils.ToR:          utils.MetaVoice,
						utils.OriginID:     "processCDR1",
						utils.OriginHost:   "OriginHost1",
						utils.RequestType:  utils.MetaRated,
						utils.AccountField: "1001",
						utils.Destination:  "1002",
						utils.SetupTime:    time.Date(2021, time.February, 2, 16, 14, 50, 0, time.UTC),
						utils.AnswerTime:   time.Date(2021, time.February, 2, 16, 15, 0, 0, time.UTC),
						utils.Usage:        2 * time.Minute,
					},
				},
			}, &reply)
		if err != nil {
			t.Fatal(err)
		}

		var cdrs []*engine.CDR
		err = client.Call(context.Background(), utils.CDRsV1GetCDRs, &utils.RPCCDRsFilterWithAPIOpts{
			RPCCDRsFilter: &utils.RPCCDRsFilter{
				RunIDs: []string{"run_1"},
			}}, &cdrs)
		if err != nil {
			t.Fatal(err)
		}
		if cdrs[0].Usage != 2*time.Minute {
			t.Errorf("expected usage to be <%+v>, received <%+v>", 2*time.Minute, cdrs[0].Usage)
		} else if cdrs[0].Cost != 1.2 {
			t.Errorf("expected cost to be <%+v>, received <%+v>", 0.6, cdrs[0].Cost)
		}
	})
	t.Run("UpdateTariffplans", func(t *testing.T) {
		newtpFiles := map[string]string{
			utils.RatesCsv: `#Id,ConnectFee,Rate,RateUnit,RateIncrement,GroupIntervalStart
RT_ANY,0,1.7,60s,1s,0s`,
			utils.DestinationRatesCsv: `#Id,DestinationId,RatesTag,RoundingMethod,RoundingDecimals,MaxCost,MaxCostStrategy
DR_ANY,*any,RT_ANY,*up,2,0,`,
			utils.RatingPlansCsv: `#Id,DestinationRatesId,TimingTag,Weight
RP_ANY,DR_ANY,*any,10`,
		}
		engine.LoadCSVs(t, client, "", newtpFiles)

		var reply string
		if err := client.Call(context.Background(), utils.CacheSv1Clear, &utils.AttrCacheIDsWithAPIOpts{}, &reply); err != nil {
			t.Error(err)
		} else if reply != utils.OK {
			t.Error("Calling CacheSv1.ReloadCache got reply: ", reply)
		}
	})

	t.Run("RerateCDRs", func(t *testing.T) {
		var reply string
		if err := client.Call(context.Background(), utils.CDRsV1RateCDRs, &engine.ArgRateCDRs{
			Flags: []string{utils.MetaRerate},
			RPCCDRsFilter: utils.RPCCDRsFilter{
				OrderBy: utils.AnswerTime,
				CGRIDs:  []string{CGRID},
			}}, &reply); err != nil {
			t.Fatal(err)
		}

		var cdrs []*engine.CDR
		err := client.Call(context.Background(), utils.CDRsV1GetCDRs, &utils.RPCCDRsFilterWithAPIOpts{
			RPCCDRsFilter: &utils.RPCCDRsFilter{
				CGRIDs:  []string{CGRID},
				OrderBy: utils.AnswerTime,
			}}, &cdrs)
		if err != nil {
			t.Fatal(err)
		}

		if cdrs[0].Cost != 3.4 {
			t.Errorf("expected cost to be <%+v>, received <%+v>", 3.4, cdrs[0].Cost)
		}
	})

}

func TestRerateFailedCDR(t *testing.T) {
	jsonCfg := `{
"rals": {
	"enabled": true
},
"cdrs": {
	"enabled": true,
	"rals_conns": ["*internal"]
},
"apiers": {
	"enabled": true
}
}`

	tpFiles := map[string]string{
		utils.DestinationRatesCsv: `#Id,DestinationId,RatesTag,RoundingMethod,RoundingDecimals,MaxCost,MaxCostStrategy
DR_ANY,*any,RT_ANY,*up,0,0,`,
		utils.RatesCsv: `#Id,ConnectFee,Rate,RateUnit,RateIncrement,GroupIntervalStart
RT_ANY,0,2,1s,1s,0s`,
		utils.RatingPlansCsv: `#Id,DestinationRatesId,TimingTag,Weight
RP_ANY,DR_ANY,*any,10`,
		utils.RatingProfilesCsv: `#Tenant,Category,Subject,ActivationTime,RatingPlanId,RatesFallbackSubject
cgrates.org,call,1001,,RP_ANY,`,
	}

	ng := engine.TestEngine{
		ConfigJSON: jsonCfg,
		TpFiles:    tpFiles,
		DBCfg:      engine.InternalDBCfg,
	}
	client, _ := ng.Run(t)

	balanceID := "test"
	processCDR := func(t *testing.T, from, to string, usage time.Duration, wantCost float64) {
		t.Helper()
		var reply string
		err := client.Call(context.Background(), utils.CDRsV1ProcessEvent,
			&engine.ArgV1ProcessEvent{
				Flags: []string{utils.MetaRALs},
				CGREvent: utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     "event1",
					Event: map[string]any{
						utils.Tenant:       "cgrates.org",
						utils.Category:     "call",
						utils.ToR:          utils.MetaVoice,
						utils.OriginID:     "processCDR",
						utils.RequestType:  utils.MetaPostpaid,
						utils.AccountField: from,
						utils.Destination:  to,
						utils.SetupTime:    time.Date(2021, time.February, 2, 16, 14, 50, 0, time.UTC),
						utils.AnswerTime:   time.Date(2021, time.February, 2, 16, 15, 0, 0, time.UTC),
						utils.Usage:        usage,
					},
				},
			}, &reply)
		if err != nil {
			t.Fatal(err)
		}
		var cdrs []*engine.CDR
		if err := client.Call(context.Background(), utils.CDRsV1GetCDRs, &utils.RPCCDRsFilterWithAPIOpts{
			RPCCDRsFilter: &utils.RPCCDRsFilter{
				OriginIDs: []string{"processCDR"},
			}}, &cdrs); err != nil {
			t.Errorf("CDRsV1GetCDRs failed unexpectedly: %v", err)
		}
		if cdrs[0].Cost != wantCost {
			t.Errorf("Cost=%v, want %v", cdrs[0].Cost, wantCost)
		}
		errMsg := "ACCOUNT_NOT_FOUND"
		if cdrs[0].Cost == -1 && cdrs[0].ExtraInfo != errMsg {
			t.Errorf("ExtraInfo err msg=%v, want %v", cdrs[0].ExtraInfo, errMsg)
		}
	}

	setAccount := func(t *testing.T, acc string, value float64) {
		t.Helper()
		var reply string
		if err := client.Call(context.Background(), utils.APIerSv2SetBalance,
			utils.AttrSetBalance{
				Tenant:      "cgrates.org",
				Account:     acc,
				Value:       value,
				BalanceType: utils.MetaMonetary,
				Balance: map[string]any{
					utils.ID: balanceID,
				},
			}, &reply); err != nil {
			t.Fatal(err)
		}

	}

	checkBalance := func(t *testing.T, acc string, want float64) {
		t.Helper()
		var acnt engine.Account
		err := client.Call(context.Background(), utils.APIerSv2GetAccount,
			&utils.AttrGetAccount{
				Tenant:  "cgrates.org",
				Account: acc,
			}, &acnt)
		if want == -1 {
			if err == nil || errors.Is(err, utils.ErrNotFound) {
				t.Fatalf("APIerSv2.GetAccount err=%v, want %v", err, utils.ErrNotFound)
			}
			return
		}
		if err != nil {
			t.Fatalf("APIerSv2GetAccount failed unexpectedly: %v", err)
		}
		got, _ := acnt.FindBalanceByID(balanceID)
		if got == nil {
			t.Errorf("acnt.FindBalanceByID(%q) could not find balance", balanceID)
		} else if got.Value != want {
			t.Errorf("acnt.FindBalanceByID(%q) balance value=%v, want %v", balanceID, got.Value, want)
		}
	}

	rerateCDR := func(t *testing.T, cost float64) {
		t.Helper()
		var reply string
		if err := client.Call(context.Background(), utils.CDRsV1RateCDRs, &engine.ArgRateCDRs{
			Flags: []string{utils.MetaRerate},
			RPCCDRsFilter: utils.RPCCDRsFilter{
				OriginIDs: []string{"processCDR"},
			}}, &reply); err != nil {
			t.Fatal(err)
		}
		var cdrs []*engine.CDR
		if err := client.Call(context.Background(), utils.CDRsV1GetCDRs, &utils.RPCCDRsFilterWithAPIOpts{
			RPCCDRsFilter: &utils.RPCCDRsFilter{
				OriginIDs: []string{"processCDR"},
			}}, &cdrs); err != nil {
			t.Errorf("CDRsV1GetCDRs failed unexpectedly: %v", err)
		}
		if cdrs[0].Cost != cost {
			t.Errorf("Cost=%v, want %v", cdrs[0].Cost, cost)
		}
	}

	checkBalance(t, "1001", -1)                      // account does not exist
	processCDR(t, "1001", "1002", 3*time.Second, -1) // processCDR fails due to non-existent acc
	setAccount(t, "1001", 10)                        // create acc with balance 10
	rerateCDR(t, 6)                                  // rerate the failed CDR
	checkBalance(t, "1001", 4)                       // check balance after rerate; expecting 4 (10-6)
}
