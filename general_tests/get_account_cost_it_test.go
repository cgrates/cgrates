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

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var aSummaryBefore *engine.AccountSummary

func TestGetAccountCost(t *testing.T) {
	ng := engine.TestEngine{
		ConfigPath: path.Join(*utils.DataDir, "conf", "samples", "rerate_cdrs_mysql"),
		TpPath:     path.Join(*utils.DataDir, "tariffplans", "reratecdrs"),
	}
	client, _ := ng.Run(t)

	CGRID := utils.GenUUID()

	t.Run("SetBalance", func(t *testing.T) {
		var reply string
		if err := client.Call(context.Background(), utils.APIerSv1SetBalances,
			utils.AttrSetBalances{
				Tenant:  "cgrates.org",
				Account: "1001",
				Balances: []*utils.AttrBalance{
					{
						BalanceType: utils.MetaVoice,
						Value:       float64(3 * time.Minute),
						Balance: map[string]any{
							utils.ID:            "voiceBalance1",
							utils.RatingSubject: "rs1",
						},
					},
					{
						BalanceType: utils.MetaVoice,
						Value:       float64(4 * time.Minute),
						Balance: map[string]any{
							utils.ID:            "voiceBalance2",
							utils.RatingSubject: "rs2",
						},
					},
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
						ID:            "voiceBalance1",
						Value:         float64(3 * time.Minute),
						RatingSubject: "rs1",
					},
					{
						ID:            "voiceBalance2",
						Value:         float64(4 * time.Minute),
						RatingSubject: "rs2",
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
			expAcnt.BalanceMap[utils.MetaVoice][1].Uuid = acnt.BalanceMap[utils.MetaVoice][1].Uuid
			if !reflect.DeepEqual(acnt, expAcnt) {
				t.Fatalf("expected: <%+v>,\nreceived: <%+v>", utils.ToJSON(expAcnt), utils.ToJSON(acnt))
			}
		}

		expAcntS := &engine.AccountSummary{
			Tenant: "cgrates.org",
			ID:     "1001",
			BalanceSummaries: engine.BalanceSummaries{
				{
					UUID:  acnt.BalanceMap[utils.MetaVoice][0].Uuid,
					ID:    "voiceBalance1",
					Type:  utils.MetaVoice,
					Value: float64(3 * time.Minute),
				},
				{
					UUID:  acnt.BalanceMap[utils.MetaVoice][1].Uuid,
					ID:    "voiceBalance2",
					Type:  utils.MetaVoice,
					Value: float64(4 * time.Minute),
				},
			},
		}
		aSummaryBefore = acnt.AsAccountSummary()
		if !reflect.DeepEqual(expAcntS, aSummaryBefore) {
			t.Fatalf("expected: <%+v>,\nreceived: <%+v>", utils.ToJSON(expAcntS), utils.ToJSON(aSummaryBefore))
		}

	})

	t.Run("GetAccountCost", func(t *testing.T) {
		var reply engine.EventCost
		err := client.Call(context.Background(), utils.APIerSV1GetAccountCost,
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
						utils.RequestType:  utils.MetaPostpaid,
						utils.AccountField: "1001",
						utils.Destination:  "1002",
						utils.SetupTime:    time.Date(2021, time.February, 2, 16, 14, 50, 0, time.UTC),
						utils.AnswerTime:   time.Date(2021, time.February, 2, 16, 15, 0, 0, time.UTC),
						utils.Usage:        4 * time.Minute,
					},
					APIOpts: map[string]any{
						utils.MetaRALsDryRun: true,
					},
				},
			}, &reply)
		if err != nil {
			t.Fatal(err)
		}
		if *reply.Cost != 2.4 {
			t.Errorf("expected cost to be <%+v>, received <%+v>", 2.4, *reply.Cost)
		}
		if *reply.Usage != 4*time.Minute {
			t.Errorf("expected cost to be <%+v>, received <%+v>", 4*time.Minute, *reply.Usage)
		}

	})

	t.Run("CheckAccountBalancesAfterGetAccountCost", checkAccountBalances(client))

	t.Run("ProcessFirstCDR", func(t *testing.T) {
		var reply []*utils.EventWithFlags
		err := client.Call(context.Background(), utils.CDRsV2ProcessEvent,
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
						utils.RequestType:  utils.MetaPostpaid,
						utils.AccountField: "1001",
						utils.Destination:  "1002",
						utils.SetupTime:    time.Date(2021, time.February, 2, 16, 14, 50, 0, time.UTC),
						utils.AnswerTime:   time.Date(2021, time.February, 2, 16, 15, 0, 0, time.UTC),
						utils.Usage:        4 * time.Minute,
					},
					APIOpts: map[string]any{
						utils.MetaRALsDryRun: true,
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
		if cdrs[0].Usage != 4*time.Minute {
			t.Errorf("expected usage to be <%+v>, received <%+v>", 2*time.Minute, cdrs[0].Usage)
		} else if cdrs[0].Cost != 2.4 {
			t.Errorf("expected cost to be <%+v>, received <%+v>", 0.6, cdrs[0].Cost)
		}
	})

	t.Run("CheckAccountBalancesAfterProcessCDR", checkAccountBalances(client))

}

func checkAccountBalances(client *birpc.Client) func(t *testing.T) {
	return func(t *testing.T) {

		expAcnt := engine.Account{
			ID: "cgrates.org:1001",
			BalanceMap: map[string]engine.Balances{
				utils.MetaVoice: {
					{
						ID:            "voiceBalance1",
						Value:         float64(3 * time.Minute),
						RatingSubject: "rs1",
					},
					{
						ID:            "voiceBalance2",
						Value:         float64(4 * time.Minute),
						RatingSubject: "rs2",
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
			expAcnt.BalanceMap[utils.MetaVoice][1].Uuid = acnt.BalanceMap[utils.MetaVoice][1].Uuid
			if !reflect.DeepEqual(acnt, expAcnt) {
				t.Fatalf("expected: <%+v>,\nreceived: <%+v>", utils.ToJSON(expAcnt), utils.ToJSON(acnt))
			}
		}
		aSummaryAfter := acnt.AsAccountSummary()
		if !reflect.DeepEqual(aSummaryBefore, aSummaryAfter) {
			t.Errorf("expected <%+v>, \nreceived <%+v>", utils.ToJSON(aSummaryBefore), utils.ToJSON(aSummaryAfter))
		}
	}
}
