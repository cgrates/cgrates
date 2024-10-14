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
