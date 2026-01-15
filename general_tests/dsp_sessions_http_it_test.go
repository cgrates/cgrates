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
	"io"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// TestDispatcherSessionsHTTP tests HTTPAgent + Sessions + Dispatcher in the same engine,
// with round-robin routing to multiple RALs engines via HTTP requests.
func TestDispatcherSessionsHTTP(t *testing.T) {
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
"http_agent": [
	{
		"id": "HTTPAgent",
		"url": "/sessions",
		"sessions_conns": ["*internal"],
		"request_payload": "*url",
		"reply_payload": "*xml",
		"request_processors": [
			{
				"id": "auth",
				"filters": ["*string:~*req.request_type:auth"],
				"flags": ["*authorize", "*accounts"],
				"request_fields": [
					{"tag": "ToR", "path": "*cgreq.ToR", "type": "*constant", "value": "*voice"},
					{"tag": "OriginID", "path": "*cgreq.OriginID", "type": "*variable", "value": "~*req.session_id"},
					{"tag": "RequestType", "path": "*cgreq.RequestType", "type": "*constant", "value": "*prepaid"},
					{"tag": "Tenant", "path": "*cgreq.Tenant", "type": "*constant", "value": "cgrates.org"},
					{"tag": "Category", "path": "*cgreq.Category", "type": "*constant", "value": "call"},
					{"tag": "Account", "path": "*cgreq.Account", "type": "*variable", "value": "~*req.account"},
					{"tag": "Subject", "path": "*cgreq.Subject", "type": "*variable", "value": "~*req.account"},
					{"tag": "Destination", "path": "*cgreq.Destination", "type": "*variable", "value": "~*req.destination"},
					{"tag": "SetupTime", "path": "*cgreq.SetupTime", "type": "*constant", "value": "*now"},
					{"tag": "Usage", "path": "*cgreq.Usage", "type": "*variable", "value": "~*req.usage"},
					{"tag": "RouteID", "path": "*cgreq.*route_id", "type": "*variable", "value": "~*req.account"}
				],
				"reply_fields": [
					{"tag": "MaxUsage", "path": "*rep.MaxUsage", "type": "*variable", "value": "~*cgrep.MaxUsage{*duration_seconds}"}
				]
			},
			{
				"id": "init",
				"filters": ["*string:~*req.request_type:init"],
				"flags": ["*initiate", "*accounts"],
				"request_fields": [
					{"tag": "ToR", "path": "*cgreq.ToR", "type": "*constant", "value": "*voice"},
					{"tag": "OriginID", "path": "*cgreq.OriginID", "type": "*variable", "value": "~*req.session_id"},
					{"tag": "RequestType", "path": "*cgreq.RequestType", "type": "*constant", "value": "*prepaid"},
					{"tag": "Tenant", "path": "*cgreq.Tenant", "type": "*constant", "value": "cgrates.org"},
					{"tag": "Category", "path": "*cgreq.Category", "type": "*constant", "value": "call"},
					{"tag": "Account", "path": "*cgreq.Account", "type": "*variable", "value": "~*req.account"},
					{"tag": "Subject", "path": "*cgreq.Subject", "type": "*variable", "value": "~*req.account"},
					{"tag": "Destination", "path": "*cgreq.Destination", "type": "*variable", "value": "~*req.destination"},
					{"tag": "SetupTime", "path": "*cgreq.SetupTime", "type": "*constant", "value": "*now"},
					{"tag": "AnswerTime", "path": "*cgreq.AnswerTime", "type": "*constant", "value": "*now"},
					{"tag": "Usage", "path": "*cgreq.Usage", "type": "*variable", "value": "~*req.usage"},
					{"tag": "RouteID", "path": "*cgreq.*route_id", "type": "*variable", "value": "~*req.account"}
				],
				"reply_fields": [
					{"tag": "MaxUsage", "path": "*rep.MaxUsage", "type": "*variable", "value": "~*cgrep.MaxUsage{*duration_seconds}"}
				]
			},
			{
				"id": "update",
				"filters": ["*string:~*req.request_type:update"],
				"flags": ["*update", "*accounts"],
				"request_fields": [
					{"tag": "ToR", "path": "*cgreq.ToR", "type": "*constant", "value": "*voice"},
					{"tag": "OriginID", "path": "*cgreq.OriginID", "type": "*variable", "value": "~*req.session_id"},
					{"tag": "RequestType", "path": "*cgreq.RequestType", "type": "*constant", "value": "*prepaid"},
					{"tag": "Tenant", "path": "*cgreq.Tenant", "type": "*constant", "value": "cgrates.org"},
					{"tag": "Category", "path": "*cgreq.Category", "type": "*constant", "value": "call"},
					{"tag": "Account", "path": "*cgreq.Account", "type": "*variable", "value": "~*req.account"},
					{"tag": "Subject", "path": "*cgreq.Subject", "type": "*variable", "value": "~*req.account"},
					{"tag": "Destination", "path": "*cgreq.Destination", "type": "*variable", "value": "~*req.destination"},
					{"tag": "SetupTime", "path": "*cgreq.SetupTime", "type": "*constant", "value": "*now"},
					{"tag": "AnswerTime", "path": "*cgreq.AnswerTime", "type": "*constant", "value": "*now"},
					{"tag": "Usage", "path": "*cgreq.Usage", "type": "*variable", "value": "~*req.usage"},
					{"tag": "RouteID", "path": "*cgreq.*route_id", "type": "*variable", "value": "~*req.account"}
				],
				"reply_fields": [
					{"tag": "MaxUsage", "path": "*rep.MaxUsage", "type": "*variable", "value": "~*cgrep.MaxUsage{*duration_seconds}"}
				]
			},
			{
				"id": "terminate",
				"filters": ["*string:~*req.request_type:terminate"],
				"flags": ["*terminate", "*accounts", "*cdrs"],
				"request_fields": [
					{"tag": "ToR", "path": "*cgreq.ToR", "type": "*constant", "value": "*voice"},
					{"tag": "OriginID", "path": "*cgreq.OriginID", "type": "*variable", "value": "~*req.session_id"},
					{"tag": "RequestType", "path": "*cgreq.RequestType", "type": "*constant", "value": "*prepaid"},
					{"tag": "Tenant", "path": "*cgreq.Tenant", "type": "*constant", "value": "cgrates.org"},
					{"tag": "Category", "path": "*cgreq.Category", "type": "*constant", "value": "call"},
					{"tag": "Account", "path": "*cgreq.Account", "type": "*variable", "value": "~*req.account"},
					{"tag": "Subject", "path": "*cgreq.Subject", "type": "*variable", "value": "~*req.account"},
					{"tag": "Destination", "path": "*cgreq.Destination", "type": "*variable", "value": "~*req.destination"},
					{"tag": "SetupTime", "path": "*cgreq.SetupTime", "type": "*constant", "value": "*now"},
					{"tag": "AnswerTime", "path": "*cgreq.AnswerTime", "type": "*constant", "value": "*now"},
					{"tag": "RouteID", "path": "*cgreq.*route_id", "type": "*variable", "value": "~*req.account"}
				],
				"reply_fields": [
					{"tag": "Status", "path": "*rep.Status", "type": "*constant", "value": "OK"}
				]
			}
		]
	}
],
"sessions": {
	"enabled": true,
	"chargers_conns": ["*internal"],
	"rals_conns": ["*localhost"],
	"cdrs_conns": ["*localhost"]
},
"dispatchers": {
	"enabled": true
},
"chargers": {
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
		utils.DispatcherProfilesCsv: `#Tenant,ID,Subsystems,FilterIDs,ActivationInterval,Strategy,StrategyParameters,ConnID,ConnFilterIDs,ConnWeight,ConnBlocker,ConnParameters,Weight
cgrates.org,DSP_RALS,,,,*round_robin,,RALS1,,10,false,,10
cgrates.org,DSP_RALS,,,,,,RALS2,,10,,,`,
	}

	// Start RALs1 first and load tariff plan data through it.
	// We cannot load via Main engine because Dispatchers is enabled there and
	// intercepts all API calls (including APIerSv1.LoadTariffPlanFromFolder),
	// causing them to fail.
	// All engines share the same Redis/MySQL, so the data will be available
	// everywhere anyway.
	ngRALs1 := TestEnvironment{
		ConfigJSON: cfgRALs1,
		TpFiles:    tpFiles,
	}
	clientRALs1, _ := ngRALs1.Setup(t, 0)

	ngRALs2 := TestEnvironment{
		ConfigJSON:     cfgRALs2,
		PreserveDataDB: true,
		PreserveStorDB: true,
	}
	ngRALs2.Setup(t, 0)

	ngMain := TestEnvironment{
		ConfigJSON:     cfgMain,
		PreserveDataDB: true,
		PreserveStorDB: true,
	}
	ngMain.Setup(t, 0)

	httpC := &http.Client{}
	var sessionNo int

	sendRequest := func(t *testing.T, reqType, sessionID, account, destination, usage string) string {
		t.Helper()
		// GET also works since HTTPAgent's *url payload uses FormValue()
		//
		// reqUrl := fmt.Sprintf(
		// 	"http://127.0.0.1:2080/sessions?request_type=%s&session_id=%s&account=%s&destination=%s&usage=%s",
		// 	reqType, sessionID, account, destination, usage)
		// resp, err := httpC.Get(reqUrl)
		resp, err := httpC.PostForm("http://127.0.0.1:2080/sessions",
			url.Values{
				"request_type": {reqType},
				"session_id":   {sessionID},
				"account":      {account},
				"destination":  {destination},
				"usage":        {usage},
			})
		if err != nil {
			t.Fatalf("HTTP request failed: %v", err)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("HTTP request returned %d: %s", resp.StatusCode, string(body))
		}
		return string(body)
	}

	runHTTPSession := func(t *testing.T, account, destination string, initUsage time.Duration, updateUsages ...time.Duration) {
		t.Helper()
		sessionNo++
		sessionID := fmt.Sprintf("session_%s_%d", account, sessionNo)

		sendRequest(t, "auth", sessionID, account, destination, initUsage.String())
		sendRequest(t, "init", sessionID, account, destination, initUsage.String())

		for _, u := range updateUsages {
			sendRequest(t, "update", sessionID, account, destination, u.String())
		}

		sendRequest(t, "terminate", sessionID, account, destination, "")
	}

	setBalance := func(t *testing.T, acc string, value float64) {
		t.Helper()
		// Use clientRALs1 for API calls since Main engine's Dispatcher would intercept them.
		var reply string
		if err := clientRALs1.Call(utils.APIerSv2SetBalance,
			utils.AttrSetBalance{
				Tenant:      "cgrates.org",
				Account:     acc,
				Value:       value,
				BalanceType: utils.MONETARY,
				Balance:     map[string]any{utils.ID: "test"},
			}, &reply); err != nil {
			t.Fatal(err)
		}
	}

	checkBalance := func(t *testing.T, acc string, want float64) {
		t.Helper()
		var acnt engine.Account
		if err := clientRALs1.Call(utils.APIerSv2GetAccount,
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

	// First cycle: 10s + 5s + 3s = 18s per account
	runHTTPSession(t, "1001", "1099", 10*time.Second, 5*time.Second, 3*time.Second)
	runHTTPSession(t, "1002", "1099", 10*time.Second, 5*time.Second, 3*time.Second)
	runHTTPSession(t, "1003", "1099", 10*time.Second, 5*time.Second, 3*time.Second)

	// Second cycle: 5s + 3s + 2s = 10s per account
	runHTTPSession(t, "1001", "1099", 5*time.Second, 3*time.Second, 2*time.Second)
	runHTTPSession(t, "1002", "1099", 5*time.Second, 3*time.Second, 2*time.Second)
	runHTTPSession(t, "1003", "1099", 5*time.Second, 3*time.Second, 2*time.Second)

	checkBalance(t, "1001", 72) // 100 - 18 - 10
	checkBalance(t, "1002", 72)
	checkBalance(t, "1003", 72)
}
