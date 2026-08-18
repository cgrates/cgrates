//go:build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package general_tests

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/ers"
	"github.com/cgrates/cgrates/utils"
)

func TestUsageRecordsSessionDebit(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaRedis, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

	exports := make(chan map[string]any, 1)
	exportServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var event map[string]any
		if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
			t.Errorf("decoding export body: %v", err)
			w.WriteHeader(http.StatusBadRequest)
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
			"id": "ur_export",
			"type": "*httpJSONMap",
			"exportPath": "%s",
			"synchronous": true,
			"fields": [
				{"tag": "AccountsCost", "path": "*exp.*accountsCost", "type": "*variable", "value": "~*opts.*accountsCost"},
				{"tag": "URID", "path": "*exp.*urID", "type": "*variable", "value": "~*opts.*urID"},
				{"tag": "UR", "path": "*exp.*ur", "type": "*variable", "value": "~*opts.*ur"}
			]
		}
	]
},
"httpAgent": [
	{
		"id": "usage_records",
		"url": "/usage-records",
		"conns": {
			"*sessions": [{"connIDs": ["*localhost"]}]
		},
		"requestPayload": "*url",
		"replyPayload": "*text",
		"requestProcessors": [
			{
				"id": "usage_record",
				"flags": ["*event"],
				"requestFields": [
					{"tag": "Account", "path": "*cgreq.Account", "type": "*variable", "value": "~*req.account", "mandatory": true},
					{"tag": "CGRID", "path": "*opts.*cgrID", "type": "*variable", "value": "~*req.sessionID"},
					{"tag": "URID", "path": "*opts.*urID", "type": "*variable", "value": "~*req.sessionID"},
					{"tag": "TotalUsage", "path": "*opts.*totalUsage", "type": "*variable", "value": "~*req.usage"},
					{"tag": "Session", "path": "*opts.*session", "type": "*constant", "value": "true"},
					{"tag": "Terminate", "path": "*opts.*terminate", "type": "*constant", "value": "true"},
					{"tag": "Rates", "path": "*opts.*rates", "type": "*constant", "value": "true"},
					{"tag": "Accounts", "path": "*opts.*accounts", "type": "*constant", "value": "true"},
					{"tag": "Debit", "path": "*opts.*debit", "type": "*constant", "value": "true"},
					{"tag": "UR", "path": "*opts.*ur", "type": "*constant", "value": "true"},
					{"tag": "EEs", "path": "*opts.*ees", "type": "*constant", "value": "true"}
				],
				"replyFields": [
					{
						"tag": "Error",
						"path": "*rep.Error",
						"type": "*variable",
						"value": "~*cgrep.Error",
						"filters": ["*notempty:~*cgrep.Error:"],
						"blocker": true
					},
					{"tag": "OK", "path": "*rep.OK", "type": "*constant", "value": "1"}
				]
			}
		]
	}
]
}`, exportServer.URL)

	testEngine := engine.TestEngine{
		ConfigJSON: cfgJSON,
		DBCfg:      engine.InternalDBCfg,
		Encoding:   *utils.Encoding,
	}
	client, cfg := testEngine.Run(t)
	usageRecordsURL := "http://" + cfg.ListenCfg().HTTPListen + "/usage-records"

	if err := client.Call(context.Background(), utils.AdminSv1SetRateProfile,
		&utils.APIRateProfile{
			RateProfile: &utils.RateProfile{
				ID: "RP_USAGE_RECORD",
				Rates: map[string]*utils.Rate{
					"RT_USAGE_RECORD": {
						ID: "RT_USAGE_RECORD",
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
		t.Fatalf("AdminSv1SetRateProfile unexpected err: %v", err)
	}
	if err := client.Call(context.Background(), utils.AdminSv1SetAccount,
		&utils.AccountWithAPIOpts{
			Account: &utils.Account{
				ID:        "1001",
				FilterIDs: []string{"*string:~*req.Account:1001"},
				Balances: map[string]*utils.Balance{
					"MONETARY": {
						ID:    "MONETARY",
						Type:  utils.MetaConcrete,
						Units: utils.NewDecimal(10, 0),
					},
				},
			},
		}, new(string)); err != nil {
		t.Fatalf("AdminSv1SetAccount unexpected err: %v", err)
	}
	usageRecordID := utils.Sha1("usage-record-debit")
	resp, err := (&http.Client{Timeout: 5 * time.Second}).PostForm(usageRecordsURL, url.Values{
		"account":   {"1001"},
		"sessionID": {usageRecordID},
		"usage":     {"1m"},
	})
	if err != nil {
		t.Fatalf("PostForm unexpected err: %v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected HTTP status %s: %s", resp.Status, body)
	}
	if !strings.Contains(string(body), "OK=1") {
		t.Fatalf("unexpected reply body: %s", body)
	}

	var account utils.Account
	if err := client.Call(context.Background(), utils.AdminSv1GetAccount,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{ID: "1001"},
		}, &account); err != nil {
		t.Fatalf("AdminSv1GetAccount unexpected err: %v", err)
	}
	units := account.Balances["MONETARY"].Units
	if units == nil || units.Compare(utils.NewDecimal(9, 0)) != 0 {
		t.Errorf("unexpected monetary balance: %v, want 9", units)
	}

	var event map[string]any
	select {
	case event = <-exports:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for EEs export")
	}
	if !utils.OptAsBool(event, utils.MetaUR) {
		t.Errorf("unexpected *ur in export: %v", event[utils.MetaUR])
	}
	if got := utils.IfaceAsString(event[utils.MetaURID]); got != usageRecordID {
		t.Errorf("unexpected %s in export: %q, want %q", utils.MetaURID, got, usageRecordID)
	}
	var accountsCost utils.EventCharges
	if err := json.Unmarshal([]byte(utils.IfaceAsString(event[utils.MetaAccountsCost])), &accountsCost); err != nil {
		t.Fatalf("decoding *accountsCost: %v", err)
	}
	if accountsCost.Concretes == nil || accountsCost.Concretes.Compare(utils.NewDecimal(1, 0)) != 0 {
		t.Errorf("unexpected *accountsCost.Concretes: %v, want 1", accountsCost.Concretes)
	}
}

func TestUsageRecordsHTTPAttribute(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaRedis, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

	attributeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_, _ = io.WriteString(w, `{"ID":"discount"}`)
	}))
	defer attributeServer.Close()

	cfgJSON := `{
"admins": {
	"enabled": true
},
"attributes": {
	"enabled": true,
	"opts": {
		"*profileIDs": [{"values": ["ATTR_PLAN", "ATTR_RATE"]}],
		"*processRuns": [{"value": "2"}]
	}
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
		"*attributes": [{"connIDs": ["*localhost"]}],
		"*accounts": [{"connIDs": ["*localhost"]}],
		"*rates": [{"connIDs": ["*localhost"]}]
	}
},
"httpAgent": [
	{
		"id": "usage_records",
		"url": "/usage-records",
		"conns": {
			"*sessions": [{"connIDs": ["*localhost"]}]
		},
		"requestPayload": "*url",
		"replyPayload": "*text",
		"requestProcessors": [
			{
				"id": "usage_record",
				"flags": ["*event"],
				"requestFields": [
					{"tag": "Account", "path": "*cgreq.Account", "type": "*variable", "value": "~*req.account", "mandatory": true},
					{"tag": "AnswerTime", "path": "*cgreq.AnswerTime", "type": "*variable", "value": "~*req.answerTime", "mandatory": true},
					{"tag": "OriginID", "path": "*opts.*originID", "type": "*variable", "value": "~*req.recordID", "mandatory": true},
					{"tag": "Usage", "path": "*opts.*usage", "type": "*variable", "value": "~*req.usage", "mandatory": true},
					{"tag": "Attributes", "path": "*opts.*attributes", "type": "*constant", "value": "true"},
					{"tag": "Rates", "path": "*opts.*rates", "type": "*constant", "value": "true"},
					{"tag": "Accounts", "path": "*opts.*accounts", "type": "*constant", "value": "true"},
					{"tag": "Debit", "path": "*opts.*debit", "type": "*constant", "value": "true"},
					{"tag": "UR", "path": "*opts.*ur", "type": "*constant", "value": "true"},
					{"tag": "BlockerError", "path": "*opts.*blockerError", "type": "*constant", "value": "true"}
				],
				"replyFields": [
					{
						"tag": "Error",
						"path": "*rep.Error",
						"type": "*variable",
						"value": "~*cgrep.Error",
						"filters": ["*notempty:~*cgrep.Error:"],
						"blocker": true
					},
					{"tag": "OK", "path": "*rep.OK", "type": "*constant", "value": "1"}
				]
			}
		]
	}
]
}`

	testEngine := engine.TestEngine{
		ConfigJSON: cfgJSON,
		DBCfg:      engine.InternalDBCfg,
		Encoding:   *utils.Encoding,
	}
	client, cfg := testEngine.Run(t)
	usageRecordsURL := "http://" + cfg.ListenCfg().HTTPListen + "/usage-records"

	httpAttributeType := utils.MetaHTTP + utils.HashtagSep + utils.IdxStart + attributeServer.URL + utils.IdxEnd
	attributeProfiles := []*utils.APIAttributeProfile{
		{
			ID: "ATTR_PLAN",
			Attributes: []*utils.ExternalAttribute{
				{
					Path:  "*req.Subscriber.Plan",
					Type:  httpAttributeType,
					Value: "~*req.Account",
				},
			},
		},
		{
			ID:        "ATTR_RATE",
			FilterIDs: []string{"*string:~*req.Subscriber.Plan.ID:discount"},
			Attributes: []*utils.ExternalAttribute{
				{
					Path:  "*req.RateGroup",
					Type:  utils.MetaVariable,
					Value: "~*req.Subscriber.Plan.ID",
				},
			},
		},
	}
	for _, profile := range attributeProfiles {
		if err := client.Call(context.Background(), utils.AdminSv1SetAttributeProfile,
			&utils.APIAttributeProfileWithAPIOpts{APIAttributeProfile: profile}, new(string)); err != nil {
			t.Fatalf("AdminSv1SetAttributeProfile %s: %v", profile.ID, err)
		}
	}

	if err := client.Call(context.Background(), utils.AdminSv1SetRateProfile,
		&utils.APIRateProfile{
			RateProfile: &utils.RateProfile{
				ID: "RP_TEST",
				Rates: map[string]*utils.Rate{
					"RT_DEFAULT": {
						ID:      "RT_DEFAULT",
						Weights: utils.DynamicWeights{{Weight: 0}},
						IntervalRates: []*utils.IntervalRate{
							{
								IntervalStart: utils.NewDecimal(0, 0),
								RecurrentFee:  utils.NewDecimal(1, 0),
								Unit:          utils.NewDecimal(int64(time.Minute), 0),
								Increment:     utils.NewDecimal(int64(time.Minute), 0),
							},
						},
					},
					"RT_DISCOUNT": {
						ID:        "RT_DISCOUNT",
						FilterIDs: []string{"*string:~*req.RateGroup:discount"},
						Weights:   utils.DynamicWeights{{Weight: 10}},
						IntervalRates: []*utils.IntervalRate{
							{
								IntervalStart: utils.NewDecimal(0, 0),
								RecurrentFee:  utils.NewDecimal(5, 1),
								Unit:          utils.NewDecimal(int64(time.Minute), 0),
								Increment:     utils.NewDecimal(int64(time.Minute), 0),
							},
						},
					},
				},
			},
		}, new(string)); err != nil {
		t.Fatalf("AdminSv1SetRateProfile: %v", err)
	}
	if err := client.Call(context.Background(), utils.AdminSv1SetAccount,
		&utils.AccountWithAPIOpts{
			Account: &utils.Account{
				ID:        "test",
				FilterIDs: []string{"*string:~*req.Account:test"},
				Balances: map[string]*utils.Balance{
					"JAN": {
						ID:    "JAN",
						Type:  utils.MetaConcrete,
						Units: utils.NewDecimal(10, 0),
						FilterIDs: []string{
							"*string:~*req.Subscriber.Plan.ID:discount",
							"*ai:~*req.AnswerTime:2026-01-01T00:00:00Z|2026-02-01T00:00:00Z",
						},
					},
					"FEB": {
						ID:    "FEB",
						Type:  utils.MetaConcrete,
						Units: utils.NewDecimal(10, 0),
						FilterIDs: []string{
							"*string:~*req.Subscriber.Plan.ID:discount",
							"*ai:~*req.AnswerTime:2026-02-01T00:00:00Z|2026-03-01T00:00:00Z",
						},
					},
				},
			},
		}, new(string)); err != nil {
		t.Fatalf("AdminSv1SetAccount: %v", err)
	}

	records := []struct {
		id         string
		answerTime string
		usage      string
	}{
		{
			id:         "february",
			answerTime: "2026-02-15T10:00:00Z",
			usage:      "2m",
		},
		{
			id:         "january",
			answerTime: "2026-01-15T10:00:00Z",
			usage:      "1m",
		},
	}
	httpClient := &http.Client{Timeout: 5 * time.Second}
	for _, record := range records {
		resp, err := httpClient.PostForm(usageRecordsURL, url.Values{
			"account":    {"test"},
			"recordID":   {record.id},
			"answerTime": {record.answerTime},
			"usage":      {record.usage},
		})
		if err != nil {
			t.Fatalf("PostForm %s: %v", record.id, err)
		}
		body, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("PostForm %s returned %s: %s", record.id, resp.Status, body)
		}
		if !strings.Contains(string(body), "OK=1") {
			t.Fatalf("PostForm %s returned %q", record.id, body)
		}
	}

	var account utils.Account
	if err := client.Call(context.Background(), utils.AdminSv1GetAccount,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{ID: "test"},
		}, &account); err != nil {
		t.Fatalf("AdminSv1GetAccount: %v", err)
	}
	wantBalances := []struct {
		id    string
		units *utils.Decimal
	}{
		{id: "JAN", units: utils.NewDecimal(95, 1)},
		{id: "FEB", units: utils.NewDecimal(9, 0)},
	}
	for _, want := range wantBalances {
		balance := account.Balances[want.id]
		if balance == nil || balance.Units == nil || balance.Units.Compare(want.units) != 0 {
			t.Errorf("%s balance = %v, want %v", want.id, balance, want.units)
		}
	}
}

func TestUsageRecordsRerating(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaMySQL:
	case utils.MetaInternal, utils.MetaRedis, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

	testEngine := engine.TestEngine{
		ConfigJSON: `{
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
			"id": "usage_records",
			"type": "*cgrcdr",
			"exportPath": "*mysql://cgrates:CGRateS.org@127.0.0.1:3306",
			"opts": {
				"sqlDBName": "cgrates",
				"sqlTableName": "cdrs"
			},
			"synchronous": true,
			"blocker": true,
			"attempts": 1,
			"failedPostsDir": "*none"
		}
	]
},
"ers": {
	"enabled": true,
	"conns": {
		"*sessions": [{"connIDs": ["*localhost"]}]
	},
	"readers": [
		{
			"id": "cgrcdr",
			"runDelay": "0",
			"type": "*cgrcdr",
			"sourcePath": "*mysql://cgrates:CGRateS.org@127.0.0.1:3306",
			"processedPath": "",
			"flags": ["*event"],
			"opts": {
				"sqlDBName": "cgrates",
				"sqlTableName": "cdrs"
			},
			"fields": [
				{"tag": "Refund", "path": "*opts.*refund", "type": "*constant", "value": "true"},
				{"tag": "BlockerError", "path": "*opts.*blockerError", "type": "*constant", "value": "true"}
			]
		}
	]
},
"httpAgent": [
	{
		"id": "usage_records",
		"url": "/usage-records",
		"conns": {
			"*sessions": [{"connIDs": ["*localhost"]}]
		},
		"requestPayload": "*url",
		"replyPayload": "*text",
		"requestProcessors": [
			{
				"id": "usage_record",
				"flags": ["*event"],
				"requestFields": [
					{"tag": "Account", "path": "*cgreq.Account", "type": "*variable", "value": "~*req.account", "mandatory": true},
					{"tag": "OriginID", "path": "*opts.*originID", "type": "*variable", "value": "~*req.urID", "mandatory": true},
					{"tag": "URID", "path": "*opts.*urID", "type": "*variable", "value": "~*req.urID", "mandatory": true},
					{"tag": "Usage", "path": "*opts.*usage", "type": "*variable", "value": "~*req.usage", "mandatory": true},
					{"tag": "Rates", "path": "*opts.*rates", "type": "*constant", "value": "true"},
					{"tag": "Accounts", "path": "*opts.*accounts", "type": "*constant", "value": "true"},
					{"tag": "Debit", "path": "*opts.*debit", "type": "*constant", "value": "true"},
					{"tag": "UR", "path": "*opts.*ur", "type": "*constant", "value": "true"},
					{"tag": "EEs", "path": "*opts.*ees", "type": "*constant", "value": "true"}
				],
				"replyFields": [
					{
						"tag": "Error",
						"path": "*rep.Error",
						"type": "*variable",
						"value": "~*cgrep.Error",
						"filters": ["*notempty:~*cgrep.Error:"],
						"blocker": true
					},
					{"tag": "OK", "path": "*rep.OK", "type": "*constant", "value": "1"}
				]
			}
		]
	}
]
}`,
		DBCfg:            engine.MySQLDBCfg,
		Encoding:         *utils.Encoding,
		GracefulShutdown: true,
	}
	client, cfg := testEngine.Run(t)
	usageRecordsURL := "http://" + cfg.ListenCfg().HTTPListen + "/usage-records"

	setRate := func(cost int64) {
		t.Helper()
		if err := client.Call(context.Background(), utils.AdminSv1SetRateProfile,
			&utils.APIRateProfile{
				RateProfile: &utils.RateProfile{
					ID: "RP_USAGE_RECORD",
					Rates: map[string]*utils.Rate{
						"RT_USAGE_RECORD": {
							ID: "RT_USAGE_RECORD",
							IntervalRates: []*utils.IntervalRate{
								{
									IntervalStart: utils.NewDecimal(0, 0),
									RecurrentFee:  utils.NewDecimal(cost, 0),
									Unit:          utils.NewDecimal(int64(time.Minute), 0),
									Increment:     utils.NewDecimal(int64(time.Minute), 0),
								},
							},
						},
					},
				},
				APIOpts: map[string]any{utils.MetaRateSOverwrite: true},
			}, new(string)); err != nil {
			t.Fatal(err)
		}
	}
	setRate(1)

	for _, accountID := range []string{"1001", "1002"} {
		if err := client.Call(context.Background(), utils.AdminSv1SetAccount,
			&utils.AccountWithAPIOpts{
				Account: &utils.Account{
					ID:        accountID,
					FilterIDs: []string{fmt.Sprintf("*string:~*req.Account:%s", accountID)},
					Balances: map[string]*utils.Balance{
						"MONETARY": {
							ID:    "MONETARY",
							Type:  utils.MetaConcrete,
							Units: utils.NewDecimal(10, 0),
						},
					},
				},
			}, new(string)); err != nil {
			t.Fatal(err)
		}
	}

	records := []struct {
		account string
		urID    string
	}{
		{account: "1001", urID: "usage-record-1001-1"},
		{account: "1001", urID: "usage-record-1001-2"},
		{account: "1002", urID: "usage-record-1002-1"},
		{account: "1002", urID: "usage-record-1002-2"},
	}
	httpClient := &http.Client{Timeout: 5 * time.Second}
	for _, record := range records {
		resp, err := httpClient.PostForm(usageRecordsURL, url.Values{
			"account": {record.account},
			"urID":    {record.urID},
			"usage":   {"1m"},
		})
		if err != nil {
			t.Fatal(err)
		}
		body, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("PostForm returned %s: %s", resp.Status, body)
		}
		if !strings.Contains(string(body), "OK=1") {
			t.Fatalf("PostForm returned %q", body)
		}
	}

	type usageRecord struct {
		account string
		cost    *utils.Decimal
	}
	getCDRs := func() map[string]usageRecord {
		t.Helper()
		var rows []*utils.CDR
		if err := client.Call(context.Background(), utils.AdminSv1GetCDRs,
			&utils.CDRFilters{Tenant: "cgrates.org"}, &rows); err != nil {
			t.Fatal(err)
		}
		if len(rows) != len(records) {
			t.Fatalf("got %d CDRs, want %d", len(rows), len(records))
		}
		cdrs := make(map[string]usageRecord, len(rows))
		for _, cdr := range rows {
			urID := utils.IfaceAsString(cdr.Opts[utils.MetaURID])
			if urID == "" {
				t.Fatal("CDR has no *urID")
			}
			if _, has := cdrs[urID]; has {
				t.Fatalf("duplicate *urID %q", urID)
			}
			value, has := cdr.Opts[utils.MetaAccountsCost]
			if !has {
				t.Fatalf("CDR %q has no *accountsCost", urID)
			}
			b, err := json.Marshal(value)
			if err != nil {
				t.Fatal(err)
			}
			var accountsCost utils.EventCharges
			if err := json.Unmarshal(b, &accountsCost); err != nil {
				t.Fatal(err)
			}
			cdrs[urID] = usageRecord{
				account: utils.IfaceAsString(cdr.Event[utils.AccountField]),
				cost:    accountsCost.Concretes,
			}
		}
		return cdrs
	}
	checkCDRs := func(want map[string]int64) {
		t.Helper()
		cdrs := getCDRs()
		for _, record := range records {
			cdr, has := cdrs[record.urID]
			if !has {
				t.Errorf("CDR %q not found", record.urID)
				continue
			}
			if cdr.account != record.account {
				t.Errorf("%s Account = %q, want %q", record.urID, cdr.account, record.account)
			}
			cost := want[record.account]
			if cdr.cost == nil || cdr.cost.Compare(utils.NewDecimal(cost, 0)) != 0 {
				t.Errorf("%s *accountsCost.Concretes = %v, want %d", record.urID, cdr.cost, cost)
			}
		}
	}
	checkBalance := func(accountID string, want int64) {
		t.Helper()
		var account utils.Account
		if err := client.Call(context.Background(), utils.AdminSv1GetAccount,
			&utils.TenantIDWithAPIOpts{
				TenantID: &utils.TenantID{ID: accountID},
			}, &account); err != nil {
			t.Fatal(err)
		}
		balance := account.Balances["MONETARY"]
		if balance == nil || balance.Units == nil || balance.Units.Compare(utils.NewDecimal(want, 0)) != 0 {
			t.Errorf("%s balance = %v, want %d", accountID, balance, want)
		}
	}

	checkCDRs(map[string]int64{"1001": 1, "1002": 1})
	checkBalance("1001", 8)
	checkBalance("1002", 8)

	setRate(2)
	if err := client.Call(context.Background(), utils.ErSv1RunReader,
		&ers.V1RunReaderParams{
			ReaderID: "cgrcdr",
			Filters:  []string{"*string:~*req.event.Account:1001"},
		}, new(string)); err != nil {
		t.Fatal(err)
	}
	waitFor(t, func() bool {
		cdrs := getCDRs()
		for _, record := range records {
			if record.account != "1001" {
				continue
			}
			cdr := cdrs[record.urID]
			if cdr.cost == nil || cdr.cost.Compare(utils.NewDecimal(2, 0)) != 0 {
				return false
			}
		}
		return true
	}, "not all CDRs for 1001 cost 2", 5*time.Second)

	checkCDRs(map[string]int64{"1001": 2, "1002": 1})
	checkBalance("1001", 6)
	checkBalance("1002", 8)
}
