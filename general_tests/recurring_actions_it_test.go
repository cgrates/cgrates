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
package general_tests

import (
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestRecurringActionPlans(t *testing.T) {
	content := `{

"general": {
	"log_level": 7,
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
		utils.AccountActionsCsv: `#Tenant,Account,ActionPlanId,ActionTriggersId,AllowNegative,Disabled
cgrates.org,1001,TOPUP_RECURRING,,,`,
		utils.ActionPlansCsv: `#Id,ActionsId,TimingId,Weight
TOPUP_RECURRING,ACT_TOPUP,RECURRING3S,10`,
		utils.ActionsCsv: `#ActionsId[0],Action[1],ExtraParameters[2],Filter[3],BalanceId[4],BalanceType[5],Categories[6],DestinationIds[7],RatingSubject[8],SharedGroup[9],ExpiryTime[10],TimingIds[11],Units[12],BalanceWeight[13],BalanceBlocker[14],BalanceDisabled[15],Weight[16]
ACT_TOPUP,*topup,,,balance_voice,*voice,,*any,,,*unlimited,RECURRING3S,10s,10,false,false,20
#change the timing
ACT_TOPUP150,*set_balance,,,balance_voice,,,,,,*unlimited,RECURRING10S,,,,,30
ACT_TOPUP150,*topup,,,balance_voice,*voice,,*any,,,*unlimited,RECURRING10S,150s,10,false,false,20`,
		utils.TimingsCsv: `#Id,StartTime,EndTime,WeekDays,MonthDays,Month,Year
RECURRING10S,,,,,*recurring+10s
RECURRING3S,,,,,*recurring+3s`,
	}

	testNg := engine.TestEngine{
		ConfigJSON: content,
		TpFiles:    tpFiles,
	}
	switch *utils.DBType {
	case utils.MetaInternal:
		testNg.DBCfg = engine.InternalDBCfg
	case utils.MetaMongo:
		testNg.DBCfg = engine.MongoDBCfg
	case utils.MetaPostgres:
		testNg.DBCfg = engine.PostgresDBCfg
	}
	client, _ := testNg.Run(t)
	time.Sleep(10 * time.Millisecond) // wait for tps to be loaded

	t.Run("CheckAccountTopup", func(t *testing.T) {
		var acnts []*engine.Account
		if err := client.Call(context.Background(), utils.APIerSv2GetAccounts,
			&utils.AttrGetAccounts{}, &acnts); err != nil {
			t.Error(err)
		} else if len(acnts) != 1 {
			t.Fatalf("Accounts received: %+v", acnts)
		}
		expAcc := &engine.Account{ID: "cgrates.org:1001", UpdateTime: acnts[0].UpdateTime}
		if !reflect.DeepEqual(expAcc, acnts[0]) {
			t.Errorf("Expecting : <%+v>, received: \n<%+v>", expAcc, acnts[0])
		}
	})
	t.Run("CheckAccountTopupAfter2Seconds", func(t *testing.T) {
		time.Sleep(2 * time.Second)
		// action shouldnt have executed yet
		var acnts []*engine.Account
		if err := client.Call(context.Background(), utils.APIerSv2GetAccounts,
			&utils.AttrGetAccounts{}, &acnts); err != nil {
			t.Error(err)
		} else if len(acnts) != 1 {
			t.Fatalf("Accounts received: %+v", acnts)
		}
		expAcc := &engine.Account{ID: "cgrates.org:1001", UpdateTime: acnts[0].UpdateTime}
		if !reflect.DeepEqual(expAcc, acnts[0]) {
			t.Errorf("Expecting : <%+v>, received: \n<%+v>", expAcc, acnts[0])
		}
	})
	t.Run("CheckAccountTopupAfter2MoreSeconds", func(t *testing.T) {
		time.Sleep(2 * time.Second)
		// action should have executed once
		var acnts []*engine.Account
		if err := client.Call(context.Background(), utils.APIerSv2GetAccounts,
			&utils.AttrGetAccounts{}, &acnts); err != nil {
			t.Error(err)
		} else if len(acnts) != 1 {
			t.Fatalf("Accounts received: %+v", acnts)
		}

		expAcc := &engine.Account{
			ID: "cgrates.org:1001",
			BalanceMap: map[string]engine.Balances{
				utils.MetaVoice: {
					&engine.Balance{
						Uuid:           acnts[0].BalanceMap[utils.MetaVoice][0].Uuid,
						ID:             "balance_voice",
						Value:          float64(10 * time.Second),
						ExpirationDate: time.Time{},
						Weight:         10,
						DestinationIDs: utils.StringMap{},
						RatingSubject:  "",
						Categories:     utils.StringMap{},
						SharedGroups:   utils.StringMap{},
						Timings: []*engine.RITiming{
							{
								ID:        "RECURRING3S",
								Years:     []int{},
								Months:    []time.Month{},
								MonthDays: []int{},
								WeekDays:  []time.Weekday{},
								StartTime: "*recurring+3s",
								EndTime:   "",
							},
						},
						TimingIDs: utils.StringMap{"RECURRING3S": true},
						Disabled:  false,
						Factors:   nil,
						Blocker:   false,
					},
				},
			},
			UnitCounters:   nil,
			ActionTriggers: nil,
			AllowNegative:  false,
			Disabled:       false,
			UpdateTime:     acnts[0].UpdateTime,
		}
		if !reflect.DeepEqual(expAcc, acnts[0]) {
			t.Errorf("Expecting : <%+v>, received: \n<%+v>", expAcc, acnts[0])
		}
	})

	t.Run("CheckAccountTopupAfter3MoreSeconds", func(t *testing.T) {
		time.Sleep(3 * time.Second)
		// action should have executed once more
		var acnts []*engine.Account
		if err := client.Call(context.Background(), utils.APIerSv2GetAccounts,
			&utils.AttrGetAccounts{}, &acnts); err != nil {
			t.Error(err)
		} else if len(acnts) != 1 {
			t.Fatalf("Accounts received: %+v", acnts)
		}
		expAcc := &engine.Account{
			ID: "cgrates.org:1001",
			BalanceMap: map[string]engine.Balances{
				utils.MetaVoice: {
					&engine.Balance{
						Uuid:           acnts[0].BalanceMap[utils.MetaVoice][0].Uuid,
						ID:             "balance_voice",
						Value:          float64(20 * time.Second),
						ExpirationDate: time.Time{},
						Weight:         10,
						DestinationIDs: utils.StringMap{},
						RatingSubject:  "",
						Categories:     utils.StringMap{},
						SharedGroups:   utils.StringMap{},
						Timings: []*engine.RITiming{
							{
								ID:        "RECURRING3S",
								Years:     []int{},
								Months:    []time.Month{},
								MonthDays: []int{},
								WeekDays:  []time.Weekday{},
								StartTime: "*recurring+3s",
								EndTime:   "",
							},
						},
						TimingIDs: utils.StringMap{"RECURRING3S": true},
						Disabled:  false,
						Factors:   nil,
						Blocker:   false,
					},
				},
			},
			UnitCounters:   nil,
			ActionTriggers: nil,
			AllowNegative:  false,
			Disabled:       false,
			UpdateTime:     acnts[0].UpdateTime,
		}
		if !reflect.DeepEqual(expAcc, acnts[0]) {
			t.Errorf("Expecting : <%+v>, received: \n<%+v>", expAcc, acnts[0])
		}
	})

	t.Run("ReSetActionPlan", func(t *testing.T) {
		atms1 := &engine.AttrSetActionPlan{
			Overwrite:       true,
			Id:              "TOPUP_RECURRING",
			ReloadScheduler: true,
			ActionPlan: []*engine.AttrActionPlan{{
				ActionsId: "ACT_TOPUP150",
				Time:      "*recurring+10s",
				TimingID:  "RECURRING10S", // takes it as a default timingID without storing it in DB
			}},
		}
		var reply string
		if err := client.Call(context.Background(), utils.APIerSv1SetActionPlan, &atms1, &reply); err != nil {
			t.Error("Got error on APIerSv1.SetActionPlan: ", err.Error())
		} else if reply != utils.OK {
			t.Errorf("Calling APIerSv1.SetActionPlan received: %s", reply)
		}
	})

	t.Run("ReSetAccountActionPlan", func(t *testing.T) {
		argSetAcnt2 := engine.AttrSetAccount{
			Tenant:               "cgrates.org",
			Account:              "1001",
			ActionPlanIDs:        []string{"TOPUP_RECURRING"},
			ActionPlansOverwrite: true,
			ReloadScheduler:      true,
		}
		var reply string
		if err := client.Call(context.Background(), utils.APIerSv2SetAccount, &argSetAcnt2, &reply); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("CheckAccountTopupAfter3MoreSeconds", func(t *testing.T) {
		time.Sleep(3 * time.Second)
		// action should have NOT executed since scheduler was reloaded with new action plan
		var acnts []*engine.Account
		if err := client.Call(context.Background(), utils.APIerSv2GetAccounts,
			&utils.AttrGetAccounts{}, &acnts); err != nil {
			t.Error(err)
		} else if len(acnts) != 1 {
			t.Fatalf("Accounts received: %+v", acnts)
		}
		expAcc := &engine.Account{
			ID: "cgrates.org:1001",
			BalanceMap: map[string]engine.Balances{
				utils.MetaVoice: {
					&engine.Balance{
						Uuid:           acnts[0].BalanceMap[utils.MetaVoice][0].Uuid,
						ID:             "balance_voice",
						Value:          float64(20 * time.Second),
						ExpirationDate: time.Time{},
						Weight:         10,
						DestinationIDs: utils.StringMap{},
						RatingSubject:  "",
						Categories:     utils.StringMap{},
						SharedGroups:   utils.StringMap{},
						Timings: []*engine.RITiming{
							{
								ID:        "RECURRING3S",
								Years:     []int{},
								Months:    []time.Month{},
								MonthDays: []int{},
								WeekDays:  []time.Weekday{},
								StartTime: "*recurring+3s",
								EndTime:   "",
							},
						},
						TimingIDs: utils.StringMap{"RECURRING3S": true},
						Disabled:  false,
						Factors:   nil,
						Blocker:   false,
					},
				},
			},
			UnitCounters:   nil,
			ActionTriggers: nil,
			AllowNegative:  false,
			Disabled:       false,
			UpdateTime:     acnts[0].UpdateTime,
		}
		if !reflect.DeepEqual(expAcc, acnts[0]) {
			t.Errorf("Expecting : <%+v>, received: \n<%+v>", expAcc, acnts[0])
		}
	})

	t.Run("CheckAccountTopupAfter8Seconds", func(t *testing.T) {
		time.Sleep(8 * time.Second)
		// action should have executed for new action plan
		var acnts []*engine.Account
		if err := client.Call(context.Background(), utils.APIerSv2GetAccounts,
			&utils.AttrGetAccounts{}, &acnts); err != nil {
			t.Error(err)
		} else if len(acnts) != 1 {
			t.Fatalf("Accounts received: %+v", acnts)
		}
		expAcc := &engine.Account{
			ID: "cgrates.org:1001",
			BalanceMap: map[string]engine.Balances{
				utils.MetaVoice: {
					&engine.Balance{
						Uuid:           acnts[0].BalanceMap[utils.MetaVoice][0].Uuid,
						ID:             "balance_voice",
						Value:          float64(170 * time.Second),
						ExpirationDate: time.Time{},
						Weight:         10,
						DestinationIDs: utils.StringMap{},
						RatingSubject:  "",
						Categories:     utils.StringMap{},
						SharedGroups:   utils.StringMap{},
						Timings: []*engine.RITiming{
							{
								ID:        "RECURRING10S",
								Years:     []int{},
								Months:    []time.Month{},
								MonthDays: []int{},
								WeekDays:  []time.Weekday{},
								StartTime: "*recurring+10s",
								EndTime:   "",
							},
						},
						TimingIDs: utils.StringMap{"RECURRING10S": true},
						Disabled:  false,
						Factors:   nil,
						Blocker:   false,
					},
				},
			},
			UnitCounters:   nil,
			ActionTriggers: nil,
			AllowNegative:  false,
			Disabled:       false,
			UpdateTime:     acnts[0].UpdateTime,
		}
		if !reflect.DeepEqual(expAcc, acnts[0]) {
			t.Errorf("Expecting : <%+v>, received: \n<%+v>", expAcc, acnts[0])
		}
	})

}
