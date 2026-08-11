//go:build call

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package general_tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/general_tests/calltest"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

func TestFreeSWITCHCall(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaRedis, utils.MetaMongo, utils.MetaPostgres, utils.MetaMySQL:
		t.Skip("freeswitch call uses internal db")
	default:
		t.Fatalf("unsupported dbtype value %q", *utils.DBType)
	}

	exports := make(chan map[string]any, 1)
	exportServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var event map[string]any
		if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
			t.Errorf("decode EEs export: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		exports <- event
	}))
	defer exportServer.Close()

	cfgJSON := fmt.Sprintf(`{
"admins": {
	"enabled": true
},
"rates": {
	"enabled": true
},
"accounts": {
	"enabled": true,
	"conns": {
		"*rates": [{"connIDs": ["*localhost"]}]
	}
},
"sessions": {
	"enabled": true,
	"conns": {
		"*accounts": [{"connIDs": ["*localhost"]}],
		"*rates": [{"connIDs": ["*localhost"]}],
		"*ees": [{"connIDs": ["*localhost"]}]
	}
},
"ees": {
	"enabled": true,
	"exporters": [
		{
			"id": "freeswitch_usage_record",
			"type": "*httpJSONMap",
			"exportPath": "%s",
			"synchronous": true,
			"fields": [
				{"tag": "AccountsCost", "path": "*exp.*accountsCost", "type": "*variable", "value": "~*opts.*accountsCost"},
				{"tag": "UR", "path": "*exp.*ur", "type": "*variable", "value": "~*opts.*ur"}
			]
		}
	]
},
"freeswitchAgent": {
	"enabled": true,
	"subscribePark": true,
	"eventSocketConns": [
		{"address": "127.0.0.1:8021", "password": "ClueCon", "reconnects": 5}
	],
	"conns": {
		"*sessions": [{"connIDs": ["*localhost"]}]
	},
	"requestProcessors": [
		{
			"id": "Park",
			"filters": ["*string:~*req.Event-Name:CHANNEL_PARK"],
			"flags": ["*event"],
			"requestFields": [
				{"tag": "BaseTmpl", "type": "*template", "value": "*fsr"},
				{"tag": "RequestType", "path": "*cgreq.RequestType", "type": "*constant", "value": "*postpaid"},
				{"tag": "Usage", "path": "*opts.*usage", "type": "*constant", "value": "1m"},
				{"tag": "Accounts", "path": "*opts.*accounts", "type": "*constant", "value": "true"},
				{"tag": "Authorize", "path": "*opts.*authorize", "type": "*constant", "value": "true"}
			]
		},
		{
			"id": "Answer",
			"filters": ["*string:~*req.Event-Name:CHANNEL_ANSWER"],
			"flags": ["*none"]
		},
		{
			"id": "Hangup",
			"filters": ["*string:~*req.Event-Name:CHANNEL_HANGUP_COMPLETE"],
			"flags": ["*event"],
			"requestFields": [
				{"tag": "BaseTmpl", "type": "*template", "value": "*fsr"},
				{"tag": "RequestType", "path": "*cgreq.RequestType", "type": "*constant", "value": "*postpaid"},
				{"tag": "Usage", "path": "*opts.*usage", "type": "*composed", "value": "~*req.variable_billsec;s"},
				{"tag": "Rates", "path": "*opts.*rates", "type": "*constant", "value": "true"},
				{"tag": "Accounts", "path": "*opts.*accounts", "type": "*constant", "value": "true"},
				{"tag": "Debit", "path": "*opts.*debit", "type": "*constant", "value": "true"},
				{"tag": "UR", "path": "*opts.*ur", "type": "*constant", "value": "true"},
				{"tag": "EEs", "path": "*opts.*ees", "type": "*constant", "value": "true"}
			]
		}
	]
}
}`, exportServer.URL)

	tutorialDir := filepath.Join(*utils.DataDir, "tutorials", "fs_evsock")
	calltest.FreeSWITCH{
		ConfigDir: filepath.Join(tutorialDir, "freeswitch/etc/freeswitch"),
		ReadyAddr: "127.0.0.1:8021",
	}.Start(t)

	ng := engine.TestEngine{
		ConfigJSON: cfgJSON,
		DBCfg:      engine.InternalDBCfg,
		Encoding:   *utils.Encoding,
	}
	client, _ := ng.Run(t)

	if err := client.Call(context.Background(), utils.AdminSv1SetRateProfile,
		&utils.APIRateProfile{
			RateProfile: &utils.RateProfile{
				ID: "RP_FREESWITCH_CALL",
				Rates: map[string]*utils.Rate{
					"RT_FREESWITCH_CALL": {
						ID: "RT_FREESWITCH_CALL",
						IntervalRates: []*utils.IntervalRate{
							{
								IntervalStart: utils.NewDecimal(0, 0),
								RecurrentFee:  utils.NewDecimal(1, 0),
								Unit:          utils.NewDecimal(int64(time.Minute), 0),
								Increment:     utils.NewDecimal(int64(time.Minute), 0),
							},
						},
					},
				},
			},
		}, new(string)); err != nil {
		t.Fatalf("set rate profile: %v", err)
	}

	call := calltest.CallParams{
		From:     "1002",
		To:       "1001",
		HoldTime: 2 * time.Second,
	}
	if err := client.Call(context.Background(), utils.AdminSv1SetAccount,
		&utils.AccountWithAPIOpts{
			Account: &utils.Account{
				ID:        call.From,
				FilterIDs: []string{"*string:~*req.Account:" + call.From},
				Balances: map[string]*utils.Balance{
					"MONETARY": {
						ID:    "MONETARY",
						Type:  utils.MetaConcrete,
						Units: utils.NewDecimal(10, 0),
					},
				},
			},
		}, new(string)); err != nil {
		t.Fatalf("set account: %v", err)
	}

	calltest.SipgoUAC{
		Addr:     "127.0.0.1:5090",
		OfferSDP: true,
	}.Call(t, call)

	var export map[string]any
	select {
	case export = <-exports:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for EEs export")
	}
	if !utils.OptAsBool(export, utils.MetaUR) {
		t.Errorf("export %s = %v, want true: %v", utils.MetaUR, export[utils.MetaUR], export)
	}
	var accountsCost utils.EventCharges
	if err := json.Unmarshal([]byte(utils.IfaceAsString(export[utils.MetaAccountsCost])), &accountsCost); err != nil {
		t.Fatalf("decode %s: %v", utils.MetaAccountsCost, err)
	}
	if accountsCost.Concretes == nil || accountsCost.Concretes.Compare(utils.NewDecimal(1, 0)) != 0 {
		t.Errorf("export %s.Concretes = %v, want 1", utils.MetaAccountsCost, accountsCost.Concretes)
	}

	var account utils.Account
	if err := client.Call(context.Background(), utils.AdminSv1GetAccount,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{ID: call.From}}, &account); err != nil {
		t.Fatalf("get account: %v", err)
	}
	balance := account.Balances["MONETARY"]
	if balance == nil || balance.Units == nil || balance.Units.Compare(utils.NewDecimal(9, 0)) != 0 {
		t.Errorf("monetary balance = %v, want 9", balance)
	}

	var activeSessions []*sessions.ExternalSession
	if err := client.Call(context.Background(), utils.SessionSv1GetActiveSessions,
		new(utils.SessionFilter), &activeSessions); err != nil && err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("get active sessions: %v", err)
	}
	if len(activeSessions) != 0 {
		t.Errorf("got %d active sessions, want none: %s", len(activeSessions), utils.ToIJSON(activeSessions))
	}
}
