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
"resources": {
	"enabled": true
},
"routes": {
	"enabled": true
},
"sessions": {
	"enabled": true,
	"conns": {
		"*accounts": [{"connIDs": ["*localhost"]}],
		"*rates": [{"connIDs": ["*localhost"]}],
		"*resources": [{"connIDs": ["*localhost"]}],
		"*routes": [{"connIDs": ["*localhost"]}],
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
				{"tag": "ResourceAllocation", "path": "*exp.cgrResourceAllocation", "type": "*variable", "value": "~*req.cgrResourceAllocation"},
				{"tag": "Routes", "path": "*exp.cgrRoutes", "type": "*variable", "value": "~*req.cgrRoutes"}
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
				{"tag": "Resources", "path": "*opts.*resources", "type": "*constant", "value": "true"},
				{"tag": "ResourcesUsageID", "path": "*opts.*resourcesUsageID", "type": "*variable", "value": "~*req.Unique-ID"},
				{"tag": "ResourcesUnits", "path": "*opts.*resourcesUnits", "type": "*constant", "value": "1"},
				{"tag": "Routes", "path": "*opts.*routes", "type": "*constant", "value": "true"},
				{"tag": "Authorize", "path": "*opts.*authorize", "type": "*constant", "value": "true"}
			],
			"replyFields": [
				{"tag": "ResourceAllocation", "path": "*rep.cgrResourceAllocation", "type": "*variable", "value": "~*cgrep.ResourceAllocation[*primary]", "mandatory": true},
				{"tag": "Routes", "path": "*rep.cgrRoutes", "type": "*composed", "value": "ARRAY::3|:"},
				{"tag": "Routes", "path": "*rep.cgrRoutes", "type": "*composed", "value": "~*cgrep.RouteProfiles[*primary][0].Routes[0].RouteID", "mandatory": true},
				{"tag": "Routes", "path": "*rep.cgrRoutes", "type": "*composed", "value": "|:"},
				{"tag": "Routes", "path": "*rep.cgrRoutes", "type": "*composed", "value": "~*cgrep.RouteProfiles[*primary][0].Routes[1].RouteID", "mandatory": true},
				{"tag": "Routes", "path": "*rep.cgrRoutes", "type": "*composed", "value": "|:"},
				{"tag": "Routes", "path": "*rep.cgrRoutes", "type": "*composed", "value": "~*cgrep.RouteProfiles[*primary][0].Routes[2].RouteID", "mandatory": true}
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
				{"tag": "ResourceAllocation", "path": "*cgreq.cgrResourceAllocation", "type": "*variable", "value": "~*req.variable_cgrResourceAllocation"},
				{"tag": "Routes", "path": "*cgreq.cgrRoutes", "type": "*variable", "value": "~*req.variable_cgrRoutes"},
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

	call := calltest.CallParams{
		From:     "1002",
		To:       "1001",
		HoldTime: 2 * time.Second,
	}
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

	if err := client.Call(context.Background(), utils.AdminSv1SetResourceProfile,
		&utils.ResourceProfileWithAPIOpts{
			ResourceProfile: &utils.ResourceProfile{
				ID:                "RES_FREESWITCH_CALL",
				FilterIDs:         []string{"*string:~*req.Account:" + call.From},
				UsageTTL:          time.Minute,
				Limit:             10,
				AllocationMessage: "RES_FREESWITCH_CALL",
				Stored:            true,
				Weights:           utils.DynamicWeights{{Weight: 10}},
			},
		}, new(string)); err != nil {
		t.Fatalf("set resource profile: %v", err)
	}
	if err := client.Call(context.Background(), utils.AdminSv1SetRouteProfile,
		&utils.RouteProfileWithAPIOpts{
			RouteProfile: &utils.RouteProfile{
				ID:        "ROUTE_FREESWITCH_CALL",
				FilterIDs: []string{"*string:~*req.Account:" + call.From},
				Weights:   utils.DynamicWeights{{Weight: 10}},
				Sorting:   utils.MetaWeight,
				Routes: []*utils.Route{
					{ID: "vendor1", Weights: utils.DynamicWeights{{Weight: 30}}},
					{ID: "vendor2", Weights: utils.DynamicWeights{{Weight: 20}}},
					{ID: "vendor3", Weights: utils.DynamicWeights{{Weight: 10}}},
				},
			},
		}, new(string)); err != nil {
		t.Fatalf("set route profile: %v", err)
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
	if got := utils.IfaceAsString(export["cgrResourceAllocation"]); got != "RES_FREESWITCH_CALL" {
		t.Errorf("export cgrResourceAllocation = %q, want %q", got, "RES_FREESWITCH_CALL")
	}
	if got := utils.IfaceAsString(export["cgrRoutes"]); got != "ARRAY::3|:vendor1|:vendor2|:vendor3" {
		t.Errorf("export cgrRoutes = %q, want %q", got, "ARRAY::3|:vendor1|:vendor2|:vendor3")
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
}
