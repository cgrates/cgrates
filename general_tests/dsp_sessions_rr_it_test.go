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
	"fmt"
	"testing"
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

func TestDspSessionsPrepaidRoundRobin(t *testing.T) {
	cfgMain := `{
"listen": {
	"rpc_json": "127.0.0.1:2012",
	"rpc_gob": "127.0.0.1:2013",
	"http": "127.0.0.1:2080"
},
"data_db": {
	"db_type": "*redis",
	"db_host": "127.0.0.1",
	"db_port": 6379,
	"db_name": "10"
},
"stor_db": {
	"db_type": "*mysql",
	"db_host": "127.0.0.1",
	"db_port": 3306,
	"db_name": "cgrates",
	"db_user": "cgrates",
	"db_password": "CGRateS.org"
},
"rpc_conns": {
	"conn_dsp": {
		"conns": [{"address": "127.0.0.1:3012", "transport":"*json"}]
	}
},
"sessions": {
	"enabled": true,
	"chargers_conns": ["*internal"],
	"rals_conns": ["conn_dsp"],
	"cdrs_conns": ["conn_dsp"]
},
"chargers": {
	"enabled": true
},
"apiers": {
	"enabled": true
}
}`

	cfgDispatcher := `{
"listen": {
	"rpc_json": "127.0.0.1:3012",
	"rpc_gob": "127.0.0.1:3013",
	"http": "127.0.0.1:3080"
},
"data_db": {
	"db_type": "*redis",
	"db_host": "127.0.0.1",
	"db_port": 6379,
	"db_name": "10"
},
"stor_db": {
	"db_type": "*mysql",
	"db_host": "127.0.0.1",
	"db_port": 3306,
	"db_name": "cgrates",
	"db_user": "cgrates",
	"db_password": "CGRateS.org"
},
"dispatchers": {
	"enabled": true
}
}`

	cfgRALs1 := `{
"listen": {
	"rpc_json": "127.0.0.1:4012",
	"rpc_gob": "127.0.0.1:4013",
	"http": "127.0.0.1:4080"
},
"data_db": {
	"db_type": "*redis",
	"db_host": "127.0.0.1",
	"db_port": 6379,
	"db_name": "10"
},
"stor_db": {
	"db_type": "*mysql",
	"db_host": "127.0.0.1",
	"db_port": 3306,
	"db_name": "cgrates",
	"db_user": "cgrates",
	"db_password": "CGRateS.org"
},
"rals": {
	"enabled": true
},
"cdrs": {
	"enabled": true,
	"chargers_conns": ["*internal"],
	"rals_conns": ["*internal"]
},
"chargers": {
	"enabled": true
},
"apiers": {
	"enabled": true
}
}`

	cfgRALs2 := `{
"listen": {
	"rpc_json": "127.0.0.1:5012",
	"rpc_gob": "127.0.0.1:5013",
	"http": "127.0.0.1:5080"
},
"data_db": {
	"db_type": "*redis",
	"db_host": "127.0.0.1",
	"db_port": 6379,
	"db_name": "10"
},
"stor_db": {
	"db_type": "*mysql",
	"db_host": "127.0.0.1",
	"db_port": 3306,
	"db_name": "cgrates",
	"db_user": "cgrates",
	"db_password": "CGRateS.org"
},
"rals": {
	"enabled": true
},
"cdrs": {
	"enabled": true,
	"chargers_conns": ["*internal"],
	"rals_conns": ["*internal"]
},
"chargers": {
	"enabled": true
},
"apiers": {
	"enabled": true
}
}`

	tpFiles := map[string]string{
		utils.DestinationRatesCsv: `#Id,DestinationId,RatesTag,RoundingMethod,RoundingDecimals,MaxCost,MaxCostStrategy
DR_ANY,*any,RT_ANY,*up,0,0,`,
		utils.RatesCsv: `#Id,ConnectFee,Rate,RateUnit,RateIncrement,GroupIntervalStart
RT_ANY,0,1,1s,1s,0s`,
		utils.RatingPlansCsv: `#Id,DestinationRatesId,TimingTag,Weight
RP_ANY,DR_ANY,*any,10`,
		utils.RatingProfilesCsv: `#Tenant,Category,Subject,ActivationTime,RatingPlanId,RatesFallbackSubject
cgrates.org,call,1001,2014-01-14T00:00:00Z,RP_ANY,
cgrates.org,call,1002,2014-01-14T00:00:00Z,RP_ANY,
cgrates.org,call,1003,2014-01-14T00:00:00Z,RP_ANY,`,
		utils.ChargersCsv: `#Tenant,ID,FilterIDs,ActivationInterval,RunID,AttributeIDs,Weight
cgrates.org,DEFAULT,,,*default,*none,0`,
		utils.DispatcherHostsCsv: `#Tenant,ID,Address,Transport,TLS
cgrates.org,RALS1,127.0.0.1:4012,*json,false
cgrates.org,RALS2,127.0.0.1:5012,*json,false`,
		// empty Subsystems field matches any subsystem
		utils.DispatcherProfilesCsv: `#Tenant,ID,Subsystems,FilterIDs,ActivationInterval,Strategy,StrategyParameters,ConnID,ConnFilterIDs,ConnWeight,ConnBlocker,ConnParameters,Weight
cgrates.org,DSP_RALS,,,,*round_robin,,RALS1,,10,false,,10
cgrates.org,DSP_RALS,,,,,,RALS2,,10,,,`,
	}

	ngMain := TestEnvironment{
		ConfigJSON: cfgMain,
		TpFiles:    tpFiles,
	}
	clientMain, _ := ngMain.Setup(t, 0)

	ngRALs1 := TestEnvironment{
		ConfigJSON:     cfgRALs1,
		PreserveDataDB: true,
		PreserveStorDB: true,
	}
	ngRALs1.Setup(t, 0)

	ngRALs2 := TestEnvironment{
		ConfigJSON:     cfgRALs2,
		PreserveDataDB: true,
		PreserveStorDB: true,
	}
	ngRALs2.Setup(t, 0)

	ngDispatcher := TestEnvironment{
		ConfigJSON:     cfgDispatcher,
		PreserveDataDB: true,
		PreserveStorDB: true,
	}
	ngDispatcher.Setup(t, 0)

	var (
		sessionNo = 0
		balanceID = "test"
	)

	runPrepaidSession := func(t *testing.T, account, destination string, initUsage time.Duration, updateUsages ...time.Duration) {
		t.Helper()
		sessionNo++
		originID := fmt.Sprintf("session_%s_%d", account, sessionNo)
		setupTime := time.Date(2024, time.January, 1, 12, 0, 0, 0, time.UTC)
		answerTime := time.Date(2024, time.January, 1, 12, 0, 10, 0, time.UTC)

		// AuthorizeEvent extracts RouteID from the event via *route_id field.
		authArgs := &sessions.V1AuthorizeArgs{
			GetMaxUsage: true,
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     utils.GenUUID(),
				Event: map[string]any{
					utils.OriginID:    originID,
					utils.Tenant:      "cgrates.org",
					utils.Category:    "call",
					utils.ToR:         utils.VOICE,
					utils.RequestType: utils.META_PREPAID,
					utils.Account:     account,
					utils.Subject:     account,
					utils.Destination: destination,
					utils.SetupTime:   setupTime,
					utils.Usage:       initUsage,
					utils.MetaRouteID: account,
				},
			},
		}
		var authRpl sessions.V1AuthorizeReply
		if err := clientMain.Call(utils.SessionSv1AuthorizeEvent, authArgs, &authRpl); err != nil {
			t.Fatalf("AuthorizeEvent failed: %v", err)
		}

		// InitiateSession stores ArgDispatcher on the session for subsequent calls.
		initArgs := &sessions.V1InitSessionArgs{
			InitSession: true,
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     utils.GenUUID(),
				Event: map[string]any{
					utils.OriginID:    originID,
					utils.Tenant:      "cgrates.org",
					utils.Category:    "call",
					utils.ToR:         utils.VOICE,
					utils.RequestType: utils.META_PREPAID,
					utils.Account:     account,
					utils.Subject:     account,
					utils.Destination: destination,
					utils.SetupTime:   setupTime,
					utils.AnswerTime:  answerTime,
					utils.Usage:       initUsage,
				},
			},
			ArgDispatcher: &utils.ArgDispatcher{
				RouteID: utils.StringPointer(account),
			},
		}
		var initRpl sessions.V1InitSessionReply
		if err := clientMain.Call(utils.SessionSv1InitiateSession, initArgs, &initRpl); err != nil {
			t.Fatalf("InitiateSession failed: %v", err)
		}

		for i, updateUsage := range updateUsages {
			updateArgs := &sessions.V1UpdateSessionArgs{
				UpdateSession: true,
				CGREvent: &utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     utils.GenUUID(),
					Event: map[string]any{
						utils.OriginID:    originID,
						utils.Tenant:      "cgrates.org",
						utils.Category:    "call",
						utils.ToR:         utils.VOICE,
						utils.RequestType: utils.META_PREPAID,
						utils.Account:     account,
						utils.Subject:     account,
						utils.Destination: destination,
						utils.SetupTime:   setupTime,
						utils.AnswerTime:  answerTime,
						utils.Usage:       updateUsage,
					},
				},
			}
			var updateRpl sessions.V1UpdateSessionReply
			if err := clientMain.Call(utils.SessionSv1UpdateSession, updateArgs, &updateRpl); err != nil {
				t.Fatalf("UpdateSession %d failed: %v", i+1, err)
			}
		}

		termArgs := &sessions.V1TerminateSessionArgs{
			TerminateSession: true,
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     utils.GenUUID(),
				Event: map[string]any{
					utils.OriginID:    originID,
					utils.Tenant:      "cgrates.org",
					utils.Category:    "call",
					utils.ToR:         utils.VOICE,
					utils.RequestType: utils.META_PREPAID,
					utils.Account:     account,
					utils.Subject:     account,
					utils.Destination: destination,
					utils.SetupTime:   setupTime,
					utils.AnswerTime:  answerTime,
				},
			},
		}
		var termRpl string
		if err := clientMain.Call(utils.SessionSv1TerminateSession, termArgs, &termRpl); err != nil {
			t.Fatalf("TerminateSession failed: %v", err)
		}

		// ProcessCDR requires explicit ArgDispatcher since it doesn't depend on the session.
		cdrArgs := &utils.CGREventWithArgDispatcher{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     utils.GenUUID(),
				Event: map[string]any{
					utils.OriginID:    originID,
					utils.Tenant:      "cgrates.org",
					utils.Category:    "call",
					utils.ToR:         utils.VOICE,
					utils.RequestType: utils.META_PREPAID,
					utils.Account:     account,
					utils.Subject:     account,
					utils.Destination: destination,
					utils.SetupTime:   setupTime,
					utils.AnswerTime:  answerTime,
					utils.Usage:       0,
				},
			},
			ArgDispatcher: &utils.ArgDispatcher{
				RouteID: utils.StringPointer(account),
			},
		}
		var cdrRpl string
		if err := clientMain.Call(utils.SessionSv1ProcessCDR, cdrArgs, &cdrRpl); err != nil {
			t.Fatalf("ProcessCDR failed: %v", err)
		}
	}

	setBalance := func(t *testing.T, acc string, value float64) {
		t.Helper()
		var reply string
		if err := clientMain.Call(utils.APIerSv2SetBalance,
			utils.AttrSetBalance{
				Tenant:      "cgrates.org",
				Account:     acc,
				Value:       value,
				BalanceType: utils.MONETARY,
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
		if err := clientMain.Call(utils.APIerSv2GetAccount,
			&utils.AttrGetAccount{
				Tenant:  "cgrates.org",
				Account: acc,
			}, &acnt); err != nil {
			t.Fatalf("GetAccount failed: %v", err)
		}
		if bal := acnt.BalanceMap[utils.MONETARY][0]; bal == nil {
			t.Errorf("balance not found for account %q", acc)
		} else if bal.Value != want {
			t.Errorf("account %q balance = %v, want %v", acc, bal.Value, want)
		}
	}

	setBalance(t, "1001", 100)
	setBalance(t, "1002", 100)
	setBalance(t, "1003", 100)

	// First cycle establishes route bindings per account.
	runPrepaidSession(t, "1001", "1099", 10*time.Second, 5*time.Second, 3*time.Second)
	runPrepaidSession(t, "1002", "1099", 10*time.Second, 5*time.Second, 3*time.Second)
	runPrepaidSession(t, "1003", "1099", 10*time.Second, 5*time.Second, 3*time.Second)

	// Second cycle verifies same account hits the same engine.
	runPrepaidSession(t, "1001", "1099", 5*time.Second, 3*time.Second, 2*time.Second)
	runPrepaidSession(t, "1002", "1099", 5*time.Second, 3*time.Second, 2*time.Second)
	runPrepaidSession(t, "1003", "1099", 5*time.Second, 3*time.Second, 2*time.Second)

	checkBalance(t, "1001", 72) // 100 - 18 - 10
	checkBalance(t, "1002", 72)
	checkBalance(t, "1003", 72)
}
