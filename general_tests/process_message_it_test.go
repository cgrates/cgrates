//go:build integration

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
	"fmt"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

// TestProcessMessage tests the functionality of the SessionSv1.ProcessMessage API
func TestProcessMessage(t *testing.T) {
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
"apiers": {
	"enabled": true
},
"rals": {
	"enabled": true
},
"chargers": {
	"enabled": true
},
"sessions": {
	"enabled": true,
	"chargers_conns": ["*internal"],
	"rals_conns": ["*internal"]
}
}`

	tpFiles := map[string]string{
		utils.ChargersCsv: `#Id,ActionsId,TimingId,Weight
#Tenant,ID,FilterIDs,ActivationInterval,RunID,AttributeIDs,Weight
cgrates.org,DEFAULT,,,DEFAULT,*none,20`,
		utils.DestinationRatesCsv: `#Id,DestinationId,RatesTag,RoundingMethod,RoundingDecimals,MaxCost,MaxCostStrategy
DR_ANY,*any,RT_ANY,*up,20,0,`,
		utils.RatesCsv: `#Id,ConnectFee,Rate,RateUnit,RateIncrement,GroupIntervalStart
RT_ANY,0,1,1,1,0`,
		utils.RatingPlansCsv: `#Id,DestinationRatesId,TimingTag,Weight
RP_ANY,DR_ANY,*any,10`,
		utils.RatingProfilesCsv: `#Tenant,Category,Subject,ActivationTime,RatingPlanId,RatesFallbackSubject
cgrates.org,sms,subj_test,,RP_ANY,`,
	}

	ng := engine.TestEngine{
		ConfigJSON: content,
		TpFiles:    tpFiles,
	}
	client, _ := ng.Run(t)

	setAccount := func(t *testing.T, acnt string, balance float64) {
		t.Helper()
		var reply string
		if err := client.Call(context.Background(), utils.APIerSv1SetBalance,
			&utils.AttrSetBalance{
				Tenant:      "cgrates.org",
				Account:     acnt,
				BalanceType: utils.MetaSMS,
				Value:       balance,
				Balance: map[string]any{
					utils.ID: "test",
				},
			}, &reply); err != nil {
			t.Error(err)
		}
	}

	getAccountBalance := func(t *testing.T, acntID string, expBalance float64) {
		t.Helper()
		var acnt engine.Account
		attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: acntID}
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
			t.Fatal(err)
		}
		if len(acnt.BalanceMap) != 1 ||
			len(acnt.BalanceMap[utils.MetaSMS]) != 1 {
			t.Fatalf("expected account to have one balance of type *sms, received %v", acnt)
		}
		smsBalance := acnt.BalanceMap[utils.MetaSMS][0]
		if smsBalance.Value != expBalance {
			t.Errorf("received account %q with *sms balance value = %v, want %v", acntID, smsBalance.Value, expBalance)
		}
	}

	processMessage := func(t *testing.T, acnt, reqType, expErr string, wantMaxUsage time.Duration) {
		t.Helper()
		args := &sessions.V1ProcessMessageArgs{
			Debit: true,
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				Event: map[string]any{
					utils.Tenant:       "cgrates.org",
					utils.Category:     "sms",
					utils.ToR:          utils.MetaSMS,
					utils.OriginID:     "test123",
					utils.OriginHost:   "127.0.0.1",
					utils.RequestType:  reqType,
					utils.AccountField: acnt,
					utils.Subject:      "subj_test",
					utils.Destination:  "+40123456789",
					utils.SetupTime:    time.Date(2024, time.September, 16, 15, 0, 0, 0, time.UTC),
					utils.AnswerTime:   time.Date(2024, time.September, 16, 15, 0, 5, 0, time.UTC),
					utils.Usage:        5,
				},
			},
		}
		wantErr := expErr != ""
		if !wantErr {
			// Set Subject only for successful cases to avoid false negatives on non-existent accounts.
			args.CGREvent.Event[utils.Subject] = "subj_test"
		}
		var reply sessions.V1ProcessMessageReply
		err := client.Call(context.Background(), utils.SessionSv1ProcessMessage, args, &reply)
		checkPrefix := fmt.Sprintf("ProcessMessage(acnt:%q,reqType:%q)", acnt, reqType)
		if wantErr {
			if err == nil || err.Error() != expErr {
				t.Fatalf("%s err=%q, want %q", checkPrefix, err, expErr)
			}
			return
		}
		if !wantErr && err != nil {
			t.Fatalf("%s unexpected err: %v", checkPrefix, err)
		}
		if reply.MaxUsage == nil {
			t.Fatalf("%s: reply.MaxUsage cannot be nil", checkPrefix)
		}
		if *reply.MaxUsage != wantMaxUsage {
			t.Errorf("%s reply.MaxUsage = %d, want %d", checkPrefix, *reply.MaxUsage, wantMaxUsage)
		}
	}

	setAccount(t, "1001", 0)
	setAccount(t, "1002", 3)
	setAccount(t, "1003", 8)

	processMessage(t, "1001", utils.MetaNone, "", 5)
	processMessage(t, "1002", utils.MetaNone, "", 5)
	processMessage(t, "1003", utils.MetaNone, "", 5)
	processMessage(t, "1004", utils.MetaNone, "", 5)
	processMessage(t, "1001", utils.MetaPostpaid, "", 5)
	processMessage(t, "1002", utils.MetaPostpaid, "", 5)
	processMessage(t, "1003", utils.MetaPostpaid, "", 5)
	processMessage(t, "1004", utils.MetaPostpaid, "", 5)
	processMessage(t, "1001", utils.MetaRated, "", 5)
	processMessage(t, "1002", utils.MetaRated, "", 5)
	processMessage(t, "1003", utils.MetaRated, "", 5)
	processMessage(t, "1004", utils.MetaRated, "", 5)
	processMessage(t, "1001", utils.MetaPseudoPrepaid, "", 0)
	processMessage(t, "1002", utils.MetaPseudoPrepaid, "", 3)
	processMessage(t, "1003", utils.MetaPseudoPrepaid, "", 5)
	processMessage(t, "1004", utils.MetaPseudoPrepaid, "RALS_ERROR:ACCOUNT_NOT_FOUND", 0)

	// The requests above should not affect the balances.
	getAccountBalance(t, "1001", 0)
	getAccountBalance(t, "1002", 3)
	getAccountBalance(t, "1003", 8)

	processMessage(t, "1001", utils.MetaPrepaid, "", 0)
	processMessage(t, "1002", utils.MetaPrepaid, "", 3)
	processMessage(t, "1003", utils.MetaPrepaid, "", 5)
	processMessage(t, "1004", utils.MetaPrepaid, "RALS_ERROR:ACCOUNT_NOT_FOUND", 0)
	getAccountBalance(t, "1001", 0) // nothing to debit
	getAccountBalance(t, "1002", 0) // 3-3
	getAccountBalance(t, "1003", 3) // 8-5
}
