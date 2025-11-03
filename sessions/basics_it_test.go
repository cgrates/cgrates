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
package sessions

import (
	"fmt"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestSessionBasics(t *testing.T) {
	var dbcfg engine.DBCfg
	switch *utils.DBType {
	case utils.MetaInternal:
		dbcfg = engine.InternalDBCfg
	case utils.MetaRedis:
		t.SkipNow()
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
"logger": {
	"type": "*stdout"
},
"db": {
	"db_conns": {
		"*default": {
			"db_type": "*internal"
    	}
	},
	"opts":{
		"internalDBRewriteInterval": "0s",
		"internalDBDumpInterval": "0s"
	}
},
"sessions": {
    "enabled": true,
    "accounts_conns": ["*internal"],
	"rates_conns": ["*internal"],
    "cdrs_conns": ["*internal"]
},
"cdrs": {
    "enabled": true,
    "accounts_conns": ["*internal"],
	"rates_conns": ["*internal"]
},
"accounts": {
    "enabled": true,
	"rates_conns": ["*internal"]
},
"admins": {
    "enabled": true
},
"rates": {
	"enabled": true
}
}`,
		TpFiles: map[string]string{
			utils.RatesCsv: `
#Tenant,ID,FilterIDs,Weights,MinCost,MaxCost,MaxCostStrategy,RateID,RateFilterIDs,RateActivationStart,RateWeights,RateBlocker,RateIntervalStart,RateFixedFee,RateRecurrentFee,RateUnit,RateIncrement
cgrates.org,RP_STANDARD,,;10,,,,RT_STANDARD,*string:~*req.Destination:1002,"* * * * *",;10,false,0s,1,1,1m,1m
cgrates.org,RP_STANDARD,,,,,,RT_STANDARD,,,,,1m,0,0.6,1m,1s
cgrates.org,RP_FALLBACK,,;0,,,,RT_FALLBACK,*string:~*req.Destination:1002,"* * * * *",;0,false,0s,0,0.01,1s,1s`,
		},
		DBCfg:    dbcfg,
		Encoding: *utils.Encoding,
		// LogBuffer: new(bytes.Buffer),
	}
	// t.Cleanup(func() { fmt.Println(ng.LogBuffer) })
	client, _ := ng.Run(t)

	// account helpers
	setAccount := func(t *testing.T, id string, balances []*utils.Balance) {
		t.Helper()
		acnt := &utils.AccountWithAPIOpts{
			Account: &utils.Account{
				Tenant: "cgrates.org",
				ID:     id,
				FilterIDs: []string{
					fmt.Sprintf("*string:~*req.Account:%s", id),
				},
			},
		}
		acnt.Balances = make(map[string]*utils.Balance)
		for _, bal := range balances {
			acnt.Balances[bal.ID] = bal
		}

		var replySet string
		if err := client.Call(context.Background(), utils.AdminSv1SetAccount,
			acnt, &replySet); err != nil {
			t.Error(err)
		}
	}

	checkAccountBalances := func(t *testing.T, acntID string, wantBalances map[string]float64) {
		t.Helper()
		var acnt utils.Account
		if err := client.Call(context.Background(), utils.AdminSv1GetAccount,
			&utils.TenantIDWithAPIOpts{
				TenantID: &utils.TenantID{
					Tenant: "cgrates.org",
					ID:     "1001",
				},
			}, &acnt); err != nil {
			t.Fatal(err)
		}
		for blncID, val := range wantBalances {
			gotUnits := acnt.Balances[blncID].Units
			wantUnits := utils.NewDecimalFromFloat64(val)
			if gotUnits.Compare(wantUnits) != 0 {
				t.Errorf("acnt %q balance %q units=%s, want %s",
					acntID, blncID, gotUnits.String(), wantUnits.String())
			}
		}
	}

	// cdr helpers
	cdrNo := 0
	processCDR := func(t *testing.T, acnt, dest, usage string, flags ...string) *utils.CDR {
		t.Helper()
		cdrNo++
		originID := fmt.Sprintf("processCDR%d", cdrNo)
		cgrEv := &utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]any{
				utils.AccountField: acnt,
				utils.Destination:  dest,
				utils.AnswerTime:   "2018-01-07T17:00:00Z",
			},
			APIOpts: map[string]any{
				utils.MetaOriginID: originID,
				utils.MetaUsage:    usage,
			},
		}
		for _, flag := range flags {
			cgrEv.APIOpts[flag] = true
		}
		var rplyProcCDR string
		if err := client.Call(context.Background(), utils.SessionSv1ProcessCDR,
			cgrEv, &rplyProcCDR); err != nil {
			t.Error(err)
		}

		var cdrs []*utils.CDR
		if err := client.Call(context.Background(), utils.AdminSv1GetCDRs,
			&utils.CDRFilters{
				FilterIDs: []string{
					fmt.Sprintf("*string:~*opts.*originID:%s", originID),
					"*exists:~*opts.*originID:",
					"*notexists:~*req.NonExistentField:",
					"*notempty:~*opts.*originID:",
				},
			}, &cdrs); err != nil {
			t.Fatal(err)
		}
		if len(cdrs) != 1 {
			t.Fatalf("%s received %d cdrs, want exactly one", utils.AdminSv1GetCDRs, len(cdrs))
		}
		return cdrs[0]
	}

	getCostDetails := func(t *testing.T, cdr *utils.CDR, field string) map[string]any {
		t.Helper()
		v, has := cdr.Opts[field]
		if !has {
			t.Fatalf("missing %q field in CDR opts", field)
		}
		costDetails, ok := v.(map[string]any)
		if !ok {
			t.Fatalf("cdr field %q of wrong type %T, want map[string]any", field, v)
		}
		return costDetails
	}

	checkCDR := func(t *testing.T, cdr *utils.CDR, wantCosts map[string]float64) {
		t.Helper()
		var got float64
		for costKey, want := range wantCosts {
			switch costKey {
			case utils.Abstracts, utils.Concretes:
				cd := getCostDetails(t, cdr, utils.MetaAccountSCost)
				if cd == nil {
					t.Fatalf("Nil costDetails")
				}
				var canCast bool
				got, canCast = cd[costKey].(float64)
				if !canCast {
					t.Fatalf("Could not cast cdr.Opts[utils.MetaCost] to float64")
				}
			case utils.Cost:
				cd := getCostDetails(t, cdr, utils.MetaRateSCost)
				if cd == nil {
					t.Fatalf("Nil costDetails")
				}
				var canCast bool
				got, canCast = cd[costKey].(float64)
				if !canCast {
					t.Fatalf("Could not cast cdr.Opts[utils.MetaCost] to float64")
				}
			case utils.MetaCost:
				var canCast bool
				got, canCast = cdr.Opts[utils.MetaCost].(float64)
				if !canCast {
					t.Fatalf("Could not cast cdr.Opts[utils.MetaCost] to float64")
				}
			default:
				t.Fatalf("invalid cdr cost key: %q", costKey)
			}
			if got != want {
				t.Errorf("cdr %s = %g, want %g", costKey, got, want)
			}
		}
	}

	checkCDRMongo := func(t *testing.T, cdr *utils.CDR, wantCosts map[string]float64) {
		t.Helper()
		var got float64
		for costKey, want := range wantCosts {
			switch costKey {
			case "abstracts", "concretes":
				cd := getCostDetails(t, cdr, utils.MetaAccountSCost)
				if cd == nil {
					t.Fatalf("Nil costDetails")
				}
				got = cd[costKey].(float64)
			case "cost":
				cd := getCostDetails(t, cdr, utils.MetaRateSCost)
				if cd == nil {
					t.Fatalf("Nil costDetails")
				}
				got = cd[costKey].(float64)
			case utils.MetaCost:
				got = cdr.Opts[utils.MetaCost].(float64)
			default:
				t.Fatalf("invalid cdr cost key: %q", costKey)
			}
			if got != want {
				t.Errorf("cdr %s = %g, want %g", costKey, got, want)
			}
		}
	}

	// session helpers
	authEvent := func(t *testing.T, wantUsage, wantErr string) {
		t.Helper()
		var reply V1AuthorizeReply
		err := client.Call(context.Background(), utils.SessionSv1AuthorizeEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				APIOpts: map[string]any{
					utils.MetaAccounts: true,
				},
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  "1002",
					utils.SetupTime:    "2018-01-07T17:00:00Z",
				},
			}, &reply)
		assertError(t, utils.SessionSv1AuthorizeEvent, err, wantErr)
		if err == nil {
			wantDecimal := utils.NewDecimalFromUsageIgnoreErr(wantUsage)
			if reply.MaxUsage.Compare(wantDecimal) != 0 {
				t.Errorf("%s reply.MaxUsage=%s, want %s",
					utils.SessionSv1AuthorizeEvent, reply.MaxUsage, wantDecimal)
				t.Logf("%s reply: %s", utils.SessionSv1AuthorizeEvent, utils.ToJSON(reply))
			}
		}
	}

	authEventWithDigest := func(t *testing.T, wantUsage time.Duration, wantErr string) {
		t.Helper()
		var reply V1AuthorizeReplyWithDigest
		err := client.Call(context.Background(), utils.SessionSv1AuthorizeEventWithDigest,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				APIOpts: map[string]any{
					utils.MetaAccounts: true,
				},
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  "1002",
					utils.SetupTime:    "2018-01-07T17:00:00Z",
				},
			}, &reply)
		assertError(t, utils.SessionSv1AuthorizeEventWithDigest, err, wantErr)
		if err == nil {
			if got := time.Duration(wantUsage).Nanoseconds(); got != reply.MaxUsage {
				t.Errorf("%s reply.MaxUsage=%d, want %d", utils.SessionSv1AuthorizeEventWithDigest, reply.MaxUsage, got)
				t.Logf("%s reply: %s", utils.SessionSv1AuthorizeEvent, utils.ToJSON(reply))
			}
		}
	}

	t.Run("auth and cdr", func(t *testing.T) {
		// Account requested not found, should fail here with error
		authEvent(t, "", "ACCOUNTS_ERROR:NOT_FOUND")

		// Available less than requested(1m)
		setAccount(t, "1001", []*utils.Balance{
			{
				ID:      "ABSTRACT1",
				Type:    utils.MetaAbstract,
				Weights: utils.DynamicWeights{&utils.DynamicWeight{Weight: 20.0}},
				CostIncrements: []*utils.CostIncrement{
					{
						Increment:    utils.NewDecimalFromUsageIgnoreErr("1s"),
						RecurrentFee: utils.NewDecimalFromFloat64(0.01),
					},
				},
				Units: utils.NewDecimalFromUsageIgnoreErr("1m"),
			},
			{
				ID:      "CONCRETE1",
				Type:    utils.MetaConcrete,
				Weights: utils.DynamicWeights{&utils.DynamicWeight{Weight: 10.0}},
				CostIncrements: []*utils.CostIncrement{
					{
						Increment:    utils.NewDecimalFromUsageIgnoreErr("1s"),
						RecurrentFee: utils.NewDecimalFromFloat64(0.01),
					},
				},
				Units: utils.NewDecimalFromFloat64(0.5),
			},
		})

		authEvent(t, "50s", "")
		authEventWithDigest(t, 50*time.Second, "")

		setAccount(t, "1001", []*utils.Balance{
			{
				ID:      "CONCRETE1",
				Type:    utils.MetaConcrete,
				Weights: utils.DynamicWeights{&utils.DynamicWeight{Weight: 10.0}},
				CostIncrements: []*utils.CostIncrement{
					{
						Increment:    utils.NewDecimalFromUsageIgnoreErr("1s"),
						RecurrentFee: utils.NewDecimalFromFloat64(0.01),
					},
				},
				Units: utils.NewDecimalFromFloat64(10),
			},
		})
		authEvent(t, "1m", "")
		authEventWithDigest(t, time.Minute, "")

		// accounting via CostIncrements
		cdr := processCDR(t, "1001", "1002", "1m30s", utils.MetaAccounts)
		if *utils.DBType == utils.MetaMongo { // field names are lowercase on mongo
			checkCDRMongo(t, cdr,
				map[string]float64{
					"abstracts":    90000000000.0,
					"concretes":    0.9,
					utils.MetaCost: 0.9,
				})
		} else {
			checkCDR(t, cdr,
				map[string]float64{
					utils.Abstracts: 90000000000.0,
					utils.Concretes: 0.9,
					utils.MetaCost:  0.9,
				})
		}

		checkAccountBalances(t, "1001", map[string]float64{
			"CONCRETE1": 9.1,
		})
	})

	t.Run("rates accounting", func(t *testing.T) {
		setAccount(t, "1001", []*utils.Balance{
			{
				ID:      "CONCRETE1",
				Type:    utils.MetaConcrete,
				Weights: utils.DynamicWeights{&utils.DynamicWeight{Weight: 10.0}},
				Units:   utils.NewDecimalFromFloat64(10),
			},
		})
		cdr := processCDR(t, "1001", "1002", "2m30s", utils.MetaAccounts)
		if *utils.DBType == utils.MetaMongo { // field names are lowercase on mongo
			checkCDRMongo(t, cdr,
				map[string]float64{
					"abstracts":    float64(150 * time.Second),
					"concretes":    2.9,
					utils.MetaCost: 2.9,
				})
		} else {
			checkCDR(t, cdr,
				map[string]float64{
					utils.Abstracts: float64(150 * time.Second),
					utils.Concretes: 2.9,
					utils.MetaCost:  2.9,
				})
		}
		checkAccountBalances(t, "1001", map[string]float64{
			"CONCRETE1": 7.1,
		})
	})

	t.Run("rating", func(t *testing.T) {
		setAccount(t, "1001", []*utils.Balance{
			{
				ID:      "CONCRETE1",
				Type:    utils.MetaConcrete,
				Weights: utils.DynamicWeights{&utils.DynamicWeight{Weight: 10.0}},
				Units:   utils.NewDecimalFromFloat64(10),
			},
		})
		cdr := processCDR(t, "1001", "1002", "2m30s", utils.MetaRates)
		if *utils.DBType == utils.MetaMongo { // field names are lowercase on mongo
			checkCDRMongo(t, cdr,
				map[string]float64{
					"cost":         2.9,
					utils.MetaCost: 2.9,
				})
		} else {
			checkCDR(t, cdr,
				map[string]float64{
					utils.Cost:     2.9,
					utils.MetaCost: 2.9,
				})
		}
	})

	t.Run("rates accounting with fallback", func(t *testing.T) {
		t.Skip("looping through all max_increments inside maxDebitAbstractsFromConcretes")
		setAccount(t, "1001", []*utils.Balance{
			{
				ID:      "CONCRETE1",
				Type:    utils.MetaConcrete,
				Weights: utils.DynamicWeights{&utils.DynamicWeight{Weight: 10.0}},
				Units:   utils.NewDecimalFromFloat64(2.9), // balance only enough for 2m30s usage
			},
		})
		cdr := processCDR(t, "1001", "1002", "3m15s", utils.MetaAccounts)
		if *utils.DBType == utils.MetaMongo { // field names are lowercase on mongo
			checkCDRMongo(t, cdr,
				map[string]float64{
					"abstracts":    float64(150 * time.Second),
					"concretes":    2.9,
					utils.MetaCost: 2.9,
				})
		} else {
			checkCDR(t, cdr,
				map[string]float64{
					utils.Abstracts: float64(150 * time.Second),
					utils.Concretes: 2.9,
					utils.MetaCost:  2.9,
				})
		}
		checkAccountBalances(t, "1001", map[string]float64{
			"CONCRETE1": 7.1,
		})
	})

	t.Run("rating and accounting", func(t *testing.T) {
		setAccount(t, "1001", []*utils.Balance{
			{
				ID:   "CONCRETE1",
				Type: utils.MetaConcrete,
				CostIncrements: []*utils.CostIncrement{
					{
						Increment:    utils.NewDecimalFromUsageIgnoreErr("1s"),
						RecurrentFee: utils.NewDecimalFromFloat64(0.01),
					},
				},
				Units: utils.NewDecimalFromFloat64(10),
			},
		})
		cdr := processCDR(t, "1001", "1002", "2m30s", utils.MetaAccounts, utils.MetaRates)
		if *utils.DBType == utils.MetaMongo { // field names are lowercase on mongo
			checkCDRMongo(t, cdr,
				map[string]float64{
					"abstracts":    float64(150 * time.Second),
					"concretes":    1.5,
					utils.MetaCost: 1.5,
					"cost":         2.9,
				})
		} else {
			checkCDR(t, cdr,
				map[string]float64{
					utils.Abstracts: float64(150 * time.Second),
					utils.Concretes: 1.5,
					utils.MetaCost:  1.5,
					utils.Cost:      2.9,
				})
		}
		checkAccountBalances(t, "1001", map[string]float64{
			"CONCRETE1": 8.5,
		})
	})
}

func assertError(t *testing.T, method string, err error, wantErr string) {
	t.Helper()
	if wantErr == "" {
		if err != nil {
			t.Fatalf("%s: unexpected error: got %v, want none", method, err)
		}
	} else {
		if err == nil {
			t.Fatalf("%s: expected error %q, got none", method, wantErr)
		}
		if err.Error() != wantErr {
			t.Fatalf("%s: error mismatch: got %q, want %q", method, err.Error(), wantErr)
		}
	}
}

func TestSessionLifecycle(t *testing.T) {
	var dbcfg engine.DBCfg
	switch *utils.DBType {
	case utils.MetaInternal:
		dbcfg = engine.InternalDBCfg
	case utils.MetaRedis:
		t.SkipNow()
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
"db": {
	"db_conns": {
		"*default": {
			"db_type": "*internal"
    	}
	},
	"opts":{
		"internalDBRewriteInterval": "0s",
		"internalDBDumpInterval": "0s"
	}
},
"sessions": {
	"enabled": true,
	"chargers_conns": ["*localhost"],
	"alterable_fields": ["AlterableField"],
	"terminate_attempts": 1
},
"chargers": {
	"enabled": true
},
"admins": {
	"enabled": true
}
}`,
		DBCfg:    dbcfg,
		Encoding: *utils.Encoding,
		// LogBuffer: new(bytes.Buffer),
	}
	// t.Cleanup(func() { fmt.Println(ng.LogBuffer) })
	client, _ := ng.Run(t)

	initSession := func(t *testing.T, originID string) {
		t.Helper()
		var reply V1InitSessionReply
		err := client.Call(context.Background(), utils.SessionSv1InitiateSession,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  "1002",
					utils.SetupTime:    "2018-01-07T17:00:00Z",
					"AlterableField":   "test_val",
				},
				APIOpts: map[string]any{
					utils.MetaOriginID: originID,
					utils.MetaInitiate: true,
					utils.MetaChargers: true,
				},
			}, &reply)
		if err != nil {
			t.Fatalf("failed to initiate session %s: %v", originID, err)
		}
	}

	updateSession := func(t *testing.T, originID string) {
		t.Helper()
		var reply V1UpdateSessionReply
		err := client.Call(context.Background(), utils.SessionSv1UpdateSession,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  "1002",
					utils.SetupTime:    "2018-01-07T17:00:00Z",
					"AlterableField":   "new_val",
				},
				APIOpts: map[string]any{
					utils.MetaOriginID: originID,
					utils.MetaUpdate:   true,
				},
			}, &reply)
		if err != nil {
			t.Fatalf("failed to update session %s: %v", originID, err)
		}
	}

	termSession := func(t *testing.T, originID string) {
		t.Helper()
		var reply string
		err := client.Call(context.Background(), utils.SessionSv1TerminateSession,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  "1002",
					utils.SetupTime:    "2018-01-07T17:00:00Z",
				},
				APIOpts: map[string]any{
					utils.MetaOriginID:  originID,
					utils.MetaTerminate: true,
				},
			}, &reply)
		if err != nil {
			t.Fatalf("failed to terminate session %s: %v", originID, err)
		}
	}

	checkActiveSessions := func(t *testing.T, wantCount int, filters ...string) []*ExternalSession {
		t.Helper()
		var sessions []*ExternalSession
		if err := client.Call(context.Background(), utils.SessionSv1GetActiveSessions,
			&utils.SessionFilter{
				Filters: filters,
			}, &sessions); err != nil {
			if wantCount == 0 && err.Error() == utils.ErrNotFound.Error() {
				t.Logf("no active sessions found (expected)")
				return nil
			}
			t.Fatalf("failed to get active sessions: %v", err)
		}
		if len(sessions) != wantCount {
			t.Fatalf("%s received %d sessions, want exactly %d",
				utils.SessionSv1GetActiveSessions, len(sessions), wantCount)
		}
		t.Logf("%s reply: %s", utils.SessionSv1GetActiveSessions, utils.ToIJSON(sessions))
		return sessions
	}

	var replySetCharger string
	if err := client.Call(context.Background(), utils.AdminSv1SetChargerProfile,
		&utils.ChargerProfileWithAPIOpts{
			ChargerProfile: &utils.ChargerProfile{
				Tenant:       "cgrates.org",
				ID:           "DEFAULT",
				RunID:        utils.MetaDefault,
				AttributeIDs: []string{utils.MetaNone},
			},
		}, &replySetCharger); err != nil {
		t.Fatal(err)
	}

	checkActiveSessions(t, 0)

	sessionID := "test-session-123"
	initSession(t, sessionID)

	sessions := checkActiveSessions(t, 1)
	want := "test_val"
	if sessions[0].CGREvent.Event["AlterableField"] != want {
		t.Errorf("after init, AlterableField = %v, want %s", sessions[0].CGREvent.Event["AlterableField"], want)
	}

	updateSession(t, sessionID)

	sessions = checkActiveSessions(t, 1)
	want = "new_val"
	if sessions[0].CGREvent.Event["AlterableField"] != want {
		t.Errorf("after update, AlterableField = %v, want %s", sessions[0].CGREvent.Event["AlterableField"], want)
	}

	termSession(t, sessionID)
	checkActiveSessions(t, 0)
}
