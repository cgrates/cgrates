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
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestSetBalancesIDs(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
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

	ng := engine.TestEngine{
		ConfigJSON: content,
	}
	client, _ := ng.Run(t)
	time.Sleep(100 * time.Millisecond)
	t.Run("SetBalancesWithID", func(t *testing.T) {

		expiryTime := time.Date(2027, 12, 31, 23, 59, 59, 0, time.UTC)

		var reply string
		if err := client.Call(context.Background(), utils.APIerSv1SetBalances,
			&utils.AttrSetBalances{
				Tenant:  "cgrates.org",
				Account: "1001",
				Balances: []*utils.AttrBalance{
					{
						BalanceType: utils.MetaData,
						Value:       1073741824,
						Balance: map[string]any{
							utils.ID:         "Balance1",
							utils.ExpiryTime: expiryTime,
						},
					},
					{
						BalanceType: utils.MetaMonetary,
						Value:       50,
						Balance: map[string]any{
							utils.ID: "MON_1",
						},
					},
					{
						BalanceType: utils.MetaSMS,
						Value:       500,
						Balance: map[string]any{
							utils.ID:         "Balance1",
							utils.ExpiryTime: expiryTime,
						},
					},
					{
						BalanceType: utils.MetaVoice,
						Value:       3600000000000,
						Balance: map[string]any{
							utils.ID:         "Balance1",
							utils.ExpiryTime: expiryTime,
						},
					},
				},
			}, &reply); err != nil {
			t.Fatal(err)
		}
		if reply != utils.OK {
			t.Errorf("expected OK, got %s", reply)
		}
	})

	t.Run("CheckBalances", func(t *testing.T) {
		var acnt engine.Account
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount,
			&utils.AttrGetAccount{
				Tenant:  "cgrates.org",
				Account: "1001",
			}, &acnt); err != nil {
			t.Fatal(err)
		}

		if len(acnt.BalanceMap) != 4 {
			t.Fatalf("expected 4 balance types, got %d", len(acnt.BalanceMap))
		}
		if len(acnt.BalanceMap[utils.MetaData]) != 1 {
			t.Fatalf("expected 1 *data balance, got %d", len(acnt.BalanceMap[utils.MetaData]))
		}
		dataBalance := acnt.BalanceMap[utils.MetaData][0]
		if dataBalance.ID != "Balance1" {
			t.Errorf("got wrong balance ID %s", dataBalance.ID)
		}
		if dataBalance.Value != 1073741824 {
			t.Errorf("*data balance value: want 1073741824, got %f", dataBalance.Value)
		}

		if len(acnt.BalanceMap[utils.MetaMonetary]) != 1 {
			t.Fatalf("expected 1 *monetary balance, got %d", len(acnt.BalanceMap[utils.MetaMonetary]))
		}
		monetaryBalance := acnt.BalanceMap[utils.MetaMonetary][0]
		if monetaryBalance.ID != "MON_1" {
			t.Errorf("got wrong balance ID %s", monetaryBalance.ID)
		}
		if monetaryBalance.Value != 50 {
			t.Errorf("*monetary balance value: want 50 , got %f", monetaryBalance.Value)
		}

		if len(acnt.BalanceMap[utils.MetaSMS]) != 1 {
			t.Fatalf("expected 1 *sms balance, got %d", len(acnt.BalanceMap[utils.MetaSMS]))
		}
		smsBalance := acnt.BalanceMap[utils.MetaSMS][0]
		if smsBalance.ID != "Balance1" {
			t.Errorf("got wrong balance ID %s", smsBalance.ID)
		}
		if smsBalance.Value != 500 {
			t.Errorf("*sms balance value: want 500, got %f", smsBalance.Value)
		}

		if len(acnt.BalanceMap[utils.MetaVoice]) != 1 {
			t.Fatalf("expected 1 *voice balance, got %d", len(acnt.BalanceMap[utils.MetaVoice]))
		}
		voiceBalance := acnt.BalanceMap[utils.MetaVoice][0]
		if voiceBalance.ID != "Balance1" {
			t.Errorf("got wrong balance id %s", voiceBalance.ID)
		}
		if voiceBalance.Value != 3600000000000 {
			t.Errorf("*voice balance value: want 3600000000000, got %f", voiceBalance.Value)
		}
	})

	t.Run("UpdateBalance", func(t *testing.T) {
		var reply string
		if err := client.Call(context.Background(), utils.APIerSv1SetBalances,
			&utils.AttrSetBalances{
				Tenant:  "cgrates.org",
				Account: "1001",
				Balances: []*utils.AttrBalance{
					{
						BalanceType: utils.MetaData,
						Value:       2147483648,
						Balance: map[string]any{
							utils.ID: "Balance1",
						},
					},
				},
			}, &reply); err != nil {
			t.Fatal(err)
		}

		var acnt engine.Account
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount,
			&utils.AttrGetAccount{
				Tenant:  "cgrates.org",
				Account: "1001",
			}, &acnt); err != nil {
			t.Fatal(err)
		}

		if len(acnt.BalanceMap[utils.MetaData]) != 1 {
			t.Fatalf("expected 1 *data balance, got %d", len(acnt.BalanceMap[utils.MetaData]))
		}
		dataBalance := acnt.BalanceMap[utils.MetaData][0]
		if dataBalance.Value != 2147483648 {
			t.Errorf("*data balance : want 2147483648, got %f", dataBalance.Value)
		}

		smsBalance := acnt.BalanceMap[utils.MetaSMS][0]
		if smsBalance.Value != 500 {
			t.Errorf("*sms balance: want 500, got %f", smsBalance.Value)
		}

		voiceBalance := acnt.BalanceMap[utils.MetaVoice][0]
		if voiceBalance.Value != 3600000000000 {
			t.Errorf("*voice balance : want 3600000000000, got %f", voiceBalance.Value)
		}
	})
}
