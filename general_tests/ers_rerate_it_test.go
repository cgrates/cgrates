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
	"bytes"
	"encoding/json"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"gorm.io/gorm"
)

func assertCDRRow(t *testing.T, db *gorm.DB, table string, want *utils.CDR) {
	t.Helper()

	var rows []utils.CDRSQLTable
	if err := db.Table(table).Find(&rows).Error; err != nil {
		t.Fatalf("failed to query %q: %v", table, err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected one row in %q, got %d", table, len(rows))
	}
	got := rows[0]
	normalizeJSON := func(value any) any {
		encoded, err := json.Marshal(value)
		if err != nil {
			t.Fatalf("failed to marshal JSON value: %v", err)
		}
		var normalized any
		if err := json.Unmarshal(encoded, &normalized); err != nil {
			t.Fatalf("failed to normalize JSON value: %v", err)
		}
		return normalized
	}
	if got.Tenant != want.Tenant ||
		!reflect.DeepEqual(normalizeJSON(got.Opts), normalizeJSON(want.Opts)) ||
		!reflect.DeepEqual(normalizeJSON(got.Event), normalizeJSON(want.Event)) {
		t.Errorf("unexpected row in %q:\nexpected tenant=%q opts=%s event=%s\nreceived tenant=%q opts=%s event=%s",
			table,
			want.Tenant, utils.ToJSON(want.Opts), utils.ToJSON(want.Event),
			got.Tenant, utils.ToJSON(got.Opts), utils.ToJSON(got.Event))
	}
}

func TestERSReRate(t *testing.T) {
	cdr := &utils.CDR{ // sample with values not realisticy calculated
		Tenant: "cgrates.org",
		Opts: map[string]any{
			utils.MetaRates:      true,
			utils.MetaUR:         true, // gives event in proccessevent reply
			utils.MetaEEs:        true, // to export cdr (either or )
			utils.MetaCGRid:      urID,
			utils.MetaURID:       urID,
			utils.OptsCDRsExport: false,
			utils.MetaChargeID:   urID,
			utils.MetaChargers:   true,
			utils.MetaCost:       1.01,
			utils.MetaExporterID: "cdr_exporter",
			utils.MetaOriginID:   "oid2",
			utils.MetaRatesCost: &utils.RateProfileCost{
				ID:              "RP_1002",
				Cost:            utils.NewDecimalFromFloat64(2.3),
				MinCost:         utils.NewDecimalFromFloat64(0),
				MaxCost:         utils.NewDecimalFromFloat64(0),
				MaxCostStrategy: "",
				CostIntervals: []*utils.RateSIntervalCost{
					{
						Increments: []*utils.RateSIncrementCost{
							{
								Usage:             utils.NewDecimalFromUsageIgnoreErr("2m"),
								RateID:            "RT_WEEK",
								RateIntervalIndex: 0,
								CompressFactor:    1,
							},
							{
								Usage:             utils.NewDecimalFromUsageIgnoreErr("1s"),
								RateID:            "RT_WEEK",
								RateIntervalIndex: 1,
								CompressFactor:    60,
							},
						},
						CompressFactor: 1,
					},
				},
				Rates: map[string]*utils.IntervalRate{
					"RP_1002_LOW": {
						IntervalStart: utils.NewDecimalFromFloat64(0),
						FixedFee:      utils.NewDecimalFromFloat64(0.1),
						RecurrentFee:  utils.NewDecimalFromFloat64(0.01),
						Unit:          utils.NewDecimalFromUsageIgnoreErr("1s"),
						Increment:     utils.NewDecimalFromUsageIgnoreErr("1m"),
					},
				},
				Altered: nil,
			},
			utils.MetaRunID:  utils.MetaDefault,
			utils.MetaSubsys: utils.MetaChargers,
			utils.MetaUsage:  "10000000000",
		},
		Event: map[string]any{
			utils.OrderID:      123,
			utils.ToR:          utils.MetaVoice,
			utils.OriginID:     "oid2",
			utils.OriginHost:   "192.168.1.1",
			utils.Source:       "test",
			utils.RequestType:  utils.MetaPrepaid,
			utils.Category:     utils.Call,
			utils.AccountField: "1001",
			utils.Subject:      "1001",
			utils.Destination:  "1002",
			utils.SetupTime:    timeStart,
			utils.AnswerTime:   timeStart,
			utils.Usage:        10 * time.Second,
			utils.ExtraInfo:    "extraInfo",
			utils.ExtraFields:  map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		},
	}
	db := openTestDB(t, "cgrates2", utils.CDRsTBL, cdr)
	db2 := openTestDB(t, "cgrates2", "cdrs2")

	buf := &bytes.Buffer{}
	ng := engine.TestEngine{
		ConfigJSON: `{
  "logger": {
	"level": 7,
  },
  "admins": {
	"enabled": true,
  },
  "rates": {
	"enabled": true
  },
  "sessions": {
    "enabled": true,
	"opts": {
	  "*ees": [{"value": true}],
	  "*rates": [{"value": true}],
	  "*ur": [{"value": true}],
	},
    "conns": {
      "*rates": [{"connIDs": ["*localhost"]}],
      "*ees": [{"connIDs": ["*localhost"]}],
      "*chargers": [{"connIDs": ["*localhost"]}]
	}
  },
  "ees": {
	"enabled": true,
	"exporters": [{
			"id": "cdr_exporter",
			"type": "*cgrcdr",
			"exportPath": "*mysql://cgrates:CGRateS.org@127.0.0.1:3306",
        	"flags": ["*log"],
			"opts": {
				"sqlDBName": "cgrates2",
				"sqlTableName": "cdrs2"
			},
			"synchronous": true,
			"blocker": false,
			"attempts": 1,
			"failedPostsDir": "*none"
		},
	]
  },
  "ers": {
    "enabled": true,
	"conns": {
		"*ees": [{"connIDs": ["*localhost"]}],
		"*sessions": [{"connIDs": ["*localhost"]}],
	},
    "readers": [
      {
        "id": "cgrcdr",
        "runDelay": "1m",
        "type": "*cgrcdr",
        "sourcePath": "*mysql://cgrates:CGRateS.org@127.0.0.1:3306",
        "startDelay": "100ms",
        "flags": ["*log","*event"],
        "tenant": "cgrates.org",
        "opts": {
          "sqlDBName": "cgrates2",
          "sqlTableName": "cdrs",
          "sqlBatchSize": 1
        },
		// "fields":[
		// 	{"tag": "OptRates", "path": "*opts.*rates", "type": "*constant",
		// 		"value": "true"},
		// 	// {"tag": "OptEEs", "path": "*opts.*ees", "type": "*constant",
		// 	// 	"value": "true"},
		// 	{"tag": "OptUsageRecord", "path": "*opts.*ur", "type": "*constant",
		// 		"value": "true"}
		// ],
      }
    ]
  }
}`,
		DBCfg:            getDBCfg(t),
		Encoding:         *utils.Encoding,
		TpPath:           filepath.Join(*utils.DataDir, "tariffplans", "tutorial"),
		LogBuffer:        buf,
		GracefulShutdown: true,
	}
	client, _ := ng.Run(t)

	ersLogCgrCDR := "<ERs> LOG, reader: <cgrcdr>"
	waitForLog(t, buf, ersLogCgrCDR, 2*time.Second)
	if got := strings.Count(buf.String(), ersLogCgrCDR); got != 1 {
		t.Fatalf("expected 1 LOG record, got %d", got)
	}
	ev := parseCGREvent(t, buf)
	exp := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "da39a3ee5e6b4b0d3255bfef95601890afd80709",
		Event: map[string]interface{}{
			utils.AccountField: "1001",
			utils.AnswerTime:   timeStart.Format(time.RFC3339),
			utils.Category:     utils.Call,
			utils.Destination:  "1002",
			utils.ExtraFields:  map[string]any{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
			utils.ExtraInfo:    "extraInfo",
			utils.OrderID:      float64(123),
			utils.OriginHost:   "192.168.1.1",
			utils.OriginID:     "oid2",
			utils.RequestType:  "*prepaid",
			utils.SetupTime:    timeStart.Format(time.RFC3339),
			utils.Source:       "test",
			utils.Subject:      "1001",
			utils.ToR:          "*voice",
			utils.Usage:        float64(10000000000),
		},
		APIOpts: map[string]interface{}{
			utils.MetaEEs:        true,
			utils.MetaUR:         true,
			utils.MetaURID:       urID,
			utils.OptsCDRsExport: false,
			utils.MetaCGRid:      urID,
			utils.MetaChargeID:   urID,
			utils.MetaChargers:   true,
			utils.MetaCost:       1.01,
			utils.MetaExporterID: "cdr_exporter",
			utils.MetaOriginID:   "oid2",
			utils.MetaRatesCost: map[string]any{
				utils.Altered: nil,
				utils.Cost:    2.3,
				utils.CostIntervals: []any{
					map[string]any{
						utils.CompressFactor: 1.0,
						utils.Increments: []any{
							map[string]any{
								utils.CompressFactor:    1.0,
								utils.RateID:            "RT_WEEK",
								utils.RateIntervalIndex: 0.0,
								utils.Usage:             120000000000.0,
							},
							map[string]any{
								utils.CompressFactor:    60.0,
								utils.RateID:            "RT_WEEK",
								utils.RateIntervalIndex: 1.0,
								utils.Usage:             1000000000.0,
							},
						},
					},
				},
				utils.ID:              "RP_1002",
				utils.MaxCost:         0.0,
				utils.MaxCostStrategy: "",
				utils.MinCost:         0.0,
				utils.Rates: map[string]any{
					"RP_1002_LOW": map[string]any{
						utils.FixedFee:      0.1,
						utils.Increment:     60000000000.0,
						utils.IntervalStart: 0.0,
						utils.RecurrentFee:  0.01,
						utils.Unit:          1000000000.0,
					},
				},
			},
			utils.MetaRates:  true,
			utils.MetaRunID:  "*default",
			utils.MetaSubsys: "*chargers",
			utils.MetaUsage:  "10000000000",
		},
	}
	if !reflect.DeepEqual(ev, exp) {
		t.Errorf("expected \n<%#v>\nreceived\n<%#v>", exp, ev)
	}
	assertCDRRow(t, db, utils.CDRsTBL, cdr)

	waitForLog(t, buf, "<EEs> LOG, exporter <cdr_exporter>, message:", 2*time.Second)
	if got := strings.Count(buf.String(), "<EEs> LOG, exporter <cdr_exporter>, message:"); got != 1 {
		t.Fatalf("expected 1 LOG record, got %d", got)
	}
	logOutput := buf.String()
	_, after, ok := strings.Cut(logOutput, "<EEs> LOG, exporter <cdr_exporter>, message: ")
	if !ok {
		t.Fatalf("CGREvent not found in log output:\n%s", logOutput)
	}
	ev = nil
	if err := json.NewDecoder(strings.NewReader(after)).Decode(&ev); err != nil {
		t.Fatal(err)
	}
	rateID := ev.APIOpts[utils.MetaRatesCost].(map[string]any)[utils.CostIntervals].([]any)[0].(map[string]any)[utils.Increments].([]any)[0].(map[string]any)[utils.RateID].(string)
	exp = &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "da39a3ee5e6b4b0d3255bfef95601890afd80709",
		Event: map[string]interface{}{
			utils.AccountField: "1001",
			utils.AnswerTime:   timeStart.Format(time.RFC3339),
			utils.Category:     utils.Call,
			utils.Destination:  "1002",
			utils.ExtraFields:  map[string]any{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
			utils.ExtraInfo:    "extraInfo",
			utils.OrderID:      float64(123),
			utils.OriginHost:   "192.168.1.1",
			utils.OriginID:     "oid2",
			utils.RequestType:  "*prepaid",
			utils.SetupTime:    timeStart.Format(time.RFC3339),
			utils.Source:       "test",
			utils.Subject:      "1001",
			utils.ToR:          "*voice",
			utils.Usage:        float64(10000000000),
		},
		APIOpts: map[string]interface{}{
			utils.MetaEEs:        true,
			utils.MetaUR:         true,
			utils.MetaURID:       urID,
			utils.OptsCDRsExport: false,
			utils.MetaCGRid:      urID,
			utils.MetaChargeID:   urID,
			utils.MetaChargers:   true,
			utils.MetaCost:       1.01,
			utils.MetaExporterID: "cdr_exporter",
			utils.MetaOriginID:   "oid2",
			utils.MetaRatesCost: map[string]any{
				utils.Altered: nil,
				utils.Cost:    0.008333333333333333,
				utils.CostIntervals: []any{
					map[string]any{
						utils.CompressFactor: 1.0,
						utils.Increments: []any{
							map[string]any{
								utils.CompressFactor:    10.0,
								utils.RateID:            rateID,
								utils.RateIntervalIndex: 0.0,
								utils.Usage:             10000000000.0,
							},
						},
					},
				},
				utils.ID:              "RP_2",
				utils.MaxCost:         0.0,
				utils.MaxCostStrategy: "",
				utils.MinCost:         0.0,
				utils.Rates: map[string]any{
					rateID: map[string]any{
						utils.FixedFee:      0.0,
						utils.Increment:     1000000000.0,
						utils.IntervalStart: 0.0,
						utils.RecurrentFee:  0.05,
						utils.Unit:          60000000000.0,
					},
				},
			},
			utils.MetaRates:  true,
			utils.MetaRunID:  "*default",
			utils.MetaSubsys: "*chargers",
			utils.MetaUsage:  "10000000000",
		},
	}
	if !reflect.DeepEqual(ev, exp) {
		t.Errorf("expected \n%#v\nreceived\n%#v", exp, ev)
	}
	assertCDRRow(t, db2, "cdrs2", &utils.CDR{
		Tenant: exp.Tenant,
		Opts:   exp.APIOpts,
		Event:  exp.Event,
	})

	// check if anything happend to accounts (should be same as coming from tariffplans)
	// 	#Tenant,ID,FilterIDs,Weights,Blockers,Opts,BalanceID,BalanceFilterIDs,BalanceWeights,BalanceBlockers,BalanceType,BalanceUnits,BalanceUnitFactors,BalanceOpts,BalanceCostIncrements,BalanceAttributeIDs,BalanceRateProfileIDs,ThresholdIDs
	// cgrates.org,1001,*string:~*req.Account:1001,;10,,,MonetaryBalance,,;10,,*concrete,10,,,;1s;;,,RP_2,
	// cgrates.org,1002,*string:~*req.Account:1002,;10,,,MonetaryBalance,,;10,,*concrete,10,,,;1s;;,,RP_2,

	var acnts []*utils.Account
	if err := client.Call(context.Background(), utils.AdminSv1GetAccounts,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &acnts); err != nil {
		t.Errorf("AdminSv1GetAccounts failed unexpectedly: %v", err)
	}
	if len(acnts) != 2 {
		t.Fatalf("AdminSv1GetAccounts len(acnts)=%v, want 2", len(acnts))
	}
	sort.Slice(acnts, func(i, j int) bool {
		return acnts[i].ID > acnts[j].ID
	})
	expAccs := []*utils.Account{
		{
			Tenant:    "cgrates.org",
			ID:        "1002",
			FilterIDs: []string{"*string:~*req.Account:1002"},
			Weights: utils.DynamicWeights{
				&utils.DynamicWeight{
					Weight: 10,
				},
			},
			Opts: map[string]any{},
			Balances: map[string]*utils.Balance{
				"MonetaryBalance": {
					ID: "MonetaryBalance",
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{
							Weight: 10,
						},
					},
					Type:  utils.MetaConcrete,
					Units: utils.NewDecimalFromFloat64(10),
					Opts:  map[string]any{},
					CostIncrements: []*utils.CostIncrement{
						{
							Increment: utils.NewDecimalFromFloat64(1000000000),
						},
					},
					RateProfileIDs: []string{"RP_2"},
				},
			},
		},
		{
			Tenant:    "cgrates.org",
			ID:        "1001",
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Weights: utils.DynamicWeights{
				&utils.DynamicWeight{
					Weight: 10,
				},
			},
			Opts: map[string]any{},
			Balances: map[string]*utils.Balance{
				"MonetaryBalance": {
					ID: "MonetaryBalance",
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{
							Weight: 10,
						},
					},
					Type:  utils.MetaConcrete,
					Units: utils.NewDecimalFromFloat64(10),
					Opts:  map[string]any{},
					CostIncrements: []*utils.CostIncrement{
						{
							Increment: utils.NewDecimalFromFloat64(1000000000),
						},
					},
					RateProfileIDs: []string{"RP_2"},
				},
			},
		},
	}
	if !reflect.DeepEqual(expAccs, acnts) {
		t.Errorf("expected \n%v\nreceived\n%v", utils.ToJSON(expAccs), utils.ToIJSON(acnts))
	}
}

func TestERSReRateWithAccount(t *testing.T) {
	cdr := &utils.CDR{ // sample with values not realisticy calculated
		Tenant: "cgrates.org",
		Opts: map[string]any{
			utils.MetaRates:      true,
			utils.MetaUR:         true, // gives event in proccessevent reply
			utils.MetaEEs:        true, // to export cdr (either or )
			utils.MetaCGRid:      urID,
			utils.MetaURID:       urID,
			utils.OptsCDRsExport: false,
			utils.MetaChargeID:   urID,
			utils.MetaChargers:   true,
			utils.MetaCost:       1.01,
			utils.MetaExporterID: "cdr_exporter",
			utils.MetaOriginID:   "oid2",
			utils.MetaRatesCost: &utils.RateProfileCost{
				ID:              "RP_1002",
				Cost:            utils.NewDecimalFromFloat64(2.3),
				MinCost:         utils.NewDecimalFromFloat64(0),
				MaxCost:         utils.NewDecimalFromFloat64(0),
				MaxCostStrategy: "",
				CostIntervals: []*utils.RateSIntervalCost{
					{
						Increments: []*utils.RateSIncrementCost{
							{
								Usage:             utils.NewDecimalFromUsageIgnoreErr("2m"),
								RateID:            "RT_WEEK",
								RateIntervalIndex: 0,
								CompressFactor:    1,
							},
							{
								Usage:             utils.NewDecimalFromUsageIgnoreErr("1s"),
								RateID:            "RT_WEEK",
								RateIntervalIndex: 1,
								CompressFactor:    60,
							},
						},
						CompressFactor: 1,
					},
				},
				Rates: map[string]*utils.IntervalRate{
					"RP_1002_LOW": {
						IntervalStart: utils.NewDecimalFromFloat64(0),
						FixedFee:      utils.NewDecimalFromFloat64(0.1),
						RecurrentFee:  utils.NewDecimalFromFloat64(0.01),
						Unit:          utils.NewDecimalFromUsageIgnoreErr("1s"),
						Increment:     utils.NewDecimalFromUsageIgnoreErr("1m"),
					},
				},
				Altered: nil,
			},
			utils.MetaRunID:  utils.MetaDefault,
			utils.MetaSubsys: utils.MetaChargers,
			utils.MetaUsage:  "10000000000",
		},
		Event: map[string]any{
			utils.OrderID:      123,
			utils.ToR:          utils.MetaVoice,
			utils.OriginID:     "oid2",
			utils.OriginHost:   "192.168.1.1",
			utils.Source:       "test",
			utils.RequestType:  utils.MetaPrepaid,
			utils.Category:     utils.Call,
			utils.AccountField: "1001",
			utils.Subject:      "1001",
			utils.Destination:  "1002",
			utils.SetupTime:    timeStart,
			utils.AnswerTime:   timeStart,
			utils.Usage:        10 * time.Second,
			utils.ExtraInfo:    "extraInfo",
			utils.ExtraFields:  map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		},
	}
	db := openTestDB(t, "cgrates2", utils.CDRsTBL, cdr)
	db2 := openTestDB(t, "cgrates2", "cdrs2")

	buf := &bytes.Buffer{}
	ng := engine.TestEngine{
		ConfigJSON: `{
  "logger": {
	"level": 7,
  },
  "admins": {
	"enabled": true,
  },  
  "rates": {
	"enabled": true
  },
  "accounts": {
	"enabled": true,
	"conns": {
	  "*rates": [{"connIDs": ["*localhost"]}],
	},
  },
  "sessions": {
    "enabled": true,
	"opts": {
	  "*rates": [{"value": true}],
	  "*accounts": [{"value": true}],
	  "*ur": [{"value": true}],
	  "*ees": [{"value": true}],
	},
    "conns": {
      "*accounts": [{"connIDs": ["*localhost"]}],
      "*rates": [{"connIDs": ["*localhost"]}],
      "*ees": [{"connIDs": ["*localhost"]}],
      "*chargers": [{"connIDs": ["*localhost"]}]
	}
  },
  "ees": {
	"enabled": true,
	"exporters": [{
			"id": "cdr_exporter",
			"type": "*cgrcdr",
			"exportPath": "*mysql://cgrates:CGRateS.org@127.0.0.1:3306",
        	"flags": ["*log"],
			"opts": {
				"sqlDBName": "cgrates2",
				"sqlTableName": "cdrs2"
			},
			"synchronous": true,
			"blocker": false,
			"attempts": 1,
			"failedPostsDir": "*none"
		},
	]
  },
  "ers": {
    "enabled": true,
	"conns": {
		"*ees": [{"connIDs": ["*localhost"]}],
		"*sessions": [{"connIDs": ["*localhost"]}],
	},
    "readers": [
      {
        "id": "cgrcdr",
        "runDelay": "1m",
        "type": "*cgrcdr",
        "sourcePath": "*mysql://cgrates:CGRateS.org@127.0.0.1:3306",
        "startDelay": "100ms",
        "flags": ["*log","*event"],
        "tenant": "cgrates.org",
        "opts": {
          "sqlDBName": "cgrates2",
          "sqlTableName": "cdrs",
          "sqlBatchSize": 1
        },
		// "fields":[
		// 	{"tag": "OptAccounts", "path": "*opts.*accounts", "type": "*constant",
		// 		"value": "true"},
		// 	{"tag": "OptRates", "path": "*opts.*rates", "type": "*constant",
		// 		"value": "true"},
		// 	// {"tag": "OptEEs", "path": "*opts.*ees", "type": "*constant",
		// 	// 	"value": "true"},
		// 	{"tag": "OptUsageRecord", "path": "*opts.*ur", "type": "*constant",
		// 		"value": "true"}
		// ],
      }
    ]
  }
}`,
		DBCfg:            getDBCfg(t),
		Encoding:         *utils.Encoding,
		TpPath:           filepath.Join(*utils.DataDir, "tariffplans", "tutorial"),
		LogBuffer:        buf,
		GracefulShutdown: true,
	}
	client, _ := ng.Run(t)

	ersLogCgrCDR := "<ERs> LOG, reader: <cgrcdr>"
	waitForLog(t, buf, ersLogCgrCDR, 2*time.Second)
	if got := strings.Count(buf.String(), ersLogCgrCDR); got != 1 {
		t.Fatalf("expected 1 LOG record, got %d", got)
	}
	ev := parseCGREvent(t, buf)
	exp := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "da39a3ee5e6b4b0d3255bfef95601890afd80709",
		Event: map[string]interface{}{
			utils.AccountField: "1001",
			utils.AnswerTime:   timeStart.Format(time.RFC3339),
			utils.Category:     utils.Call,
			utils.Destination:  "1002",
			utils.ExtraFields:  map[string]any{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
			utils.ExtraInfo:    "extraInfo",
			utils.OrderID:      float64(123),
			utils.OriginHost:   "192.168.1.1",
			utils.OriginID:     "oid2",
			utils.RequestType:  "*prepaid",
			utils.SetupTime:    timeStart.Format(time.RFC3339),
			utils.Source:       "test",
			utils.Subject:      "1001",
			utils.ToR:          "*voice",
			utils.Usage:        float64(10000000000),
		},
		APIOpts: map[string]interface{}{
			utils.MetaEEs:        true,
			utils.MetaUR:         true,
			utils.MetaURID:       urID,
			utils.OptsCDRsExport: false,
			utils.MetaCGRid:      urID,
			utils.MetaChargeID:   urID,
			utils.MetaChargers:   true,
			utils.MetaCost:       1.01,
			utils.MetaExporterID: "cdr_exporter",
			utils.MetaOriginID:   "oid2",
			utils.MetaRatesCost: map[string]any{
				utils.Altered: nil,
				utils.Cost:    2.3,
				utils.CostIntervals: []any{
					map[string]any{
						utils.CompressFactor: 1.0,
						utils.Increments: []any{
							map[string]any{
								utils.CompressFactor:    1.0,
								utils.RateID:            "RT_WEEK",
								utils.RateIntervalIndex: 0.0,
								utils.Usage:             120000000000.0,
							},
							map[string]any{
								utils.CompressFactor:    60.0,
								utils.RateID:            "RT_WEEK",
								utils.RateIntervalIndex: 1.0,
								utils.Usage:             1000000000.0,
							},
						},
					},
				},
				utils.ID:              "RP_1002",
				utils.MaxCost:         0.0,
				utils.MaxCostStrategy: "",
				utils.MinCost:         0.0,
				utils.Rates: map[string]any{
					"RP_1002_LOW": map[string]any{
						utils.FixedFee:      0.1,
						utils.Increment:     60000000000.0,
						utils.IntervalStart: 0.0,
						utils.RecurrentFee:  0.01,
						utils.Unit:          1000000000.0,
					},
				},
			},
			utils.MetaRates:  true,
			utils.MetaRunID:  "*default",
			utils.MetaSubsys: "*chargers",
			utils.MetaUsage:  "10000000000",
		},
	}
	if !reflect.DeepEqual(ev, exp) {
		t.Errorf("expected \n<%#v>\nreceived\n<%#v>", exp, ev)
	}
	assertCDRRow(t, db, utils.CDRsTBL, cdr)

	waitForLog(t, buf, "<EEs> LOG, exporter <cdr_exporter>, message:", 2*time.Second)
	if got := strings.Count(buf.String(), "<EEs> LOG, exporter <cdr_exporter>, message:"); got != 1 {
		t.Fatalf("expected 1 LOG record, got %d", got)
	}
	logOutput := buf.String()
	_, after, ok := strings.Cut(logOutput, "<EEs> LOG, exporter <cdr_exporter>, message: ")
	if !ok {
		t.Fatalf("CGREvent not found in log output:\n%s", logOutput)
	}
	ev = nil
	if err := json.NewDecoder(strings.NewReader(after)).Decode(&ev); err != nil {
		t.Fatal(err)
	}
	rateID := ev.APIOpts[utils.MetaRatesCost].(map[string]any)[utils.CostIntervals].([]any)[0].(map[string]any)[utils.Increments].([]any)[0].(map[string]any)[utils.RateID].(string)
	var accounting1, accounting2, joinedChargeIDs, ratingID, accRateID, chargingID string
	accounting := ev.APIOpts[utils.MetaAccountsCost].(map[string]any)[utils.Accounting].(map[string]any)
	chargingID = ev.APIOpts[utils.MetaAccountsCost].(map[string]any)[utils.Charges].([]any)[0].(map[string]any)[utils.ChargingID].(string)
	for key := range ev.APIOpts[utils.MetaAccountsCost].(map[string]any)[utils.Rates].(map[string]any) {
		accRateID = key
	}
	for key := range accounting {
		if accounting[key].(map[string]any)["BalanceID"] == "*mockabstract" {
			accounting1 = key
			joinedChargeIDs = accounting[key].(map[string]any)["JoinedChargeIDs"].([]interface{})[0].(string)
			ratingID = accounting[key].(map[string]any)["RatingID"].(string)
		} else {
			accounting2 = key
		}
	}
	exp = &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "da39a3ee5e6b4b0d3255bfef95601890afd80709",
		Event: map[string]interface{}{
			utils.AccountField: "1001",
			utils.AnswerTime:   timeStart.Format(time.RFC3339),
			utils.Category:     utils.Call,
			utils.Destination:  "1002",
			utils.ExtraFields:  map[string]any{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
			utils.ExtraInfo:    "extraInfo",
			utils.OrderID:      float64(123),
			utils.OriginHost:   "192.168.1.1",
			utils.OriginID:     "oid2",
			utils.RequestType:  "*prepaid",
			utils.SetupTime:    timeStart.Format(time.RFC3339),
			utils.Source:       "test",
			utils.Subject:      "1001",
			utils.ToR:          "*voice",
			utils.Usage:        float64(10000000000),
		},
		APIOpts: map[string]interface{}{
			utils.MetaEEs:        true,
			utils.MetaUR:         true,
			utils.MetaURID:       urID,
			utils.OptsCDRsExport: false,
			utils.MetaCGRid:      urID,
			utils.MetaChargeID:   urID,
			utils.MetaChargers:   true,
			utils.MetaCost:       1.01,
			utils.MetaExporterID: "cdr_exporter",
			utils.MetaOriginID:   "oid2",
			utils.MetaRatesCost: map[string]any{
				utils.Altered: nil,
				utils.Cost:    0.008333333333333333,
				utils.CostIntervals: []any{
					map[string]any{
						utils.CompressFactor: 1.0,
						utils.Increments: []any{
							map[string]any{
								utils.CompressFactor:    10.0,
								utils.RateID:            rateID,
								utils.RateIntervalIndex: 0.0,
								utils.Usage:             10000000000.0,
							},
						},
					},
				},
				utils.ID:              "RP_2",
				utils.MaxCost:         0.0,
				utils.MaxCostStrategy: "",
				utils.MinCost:         0.0,
				utils.Rates: map[string]any{
					rateID: map[string]any{
						utils.FixedFee:      0.0,
						utils.Increment:     1000000000.0,
						utils.IntervalStart: 0.0,
						utils.RecurrentFee:  0.05,
						utils.Unit:          60000000000.0,
					},
				},
			},
			utils.MetaAccountsCost: map[string]any{
				utils.Abstracts: 10000000000.0,
				utils.Accounting: map[string]any{
					accounting1: map[string]any{
						utils.AccountID:       "1001",
						utils.AttributeIDs:    nil,
						utils.BalanceID:       "*mockabstract",
						utils.BalanceLimit:    nil,
						utils.JoinedChargeIDs: []interface{}{joinedChargeIDs},
						utils.RatingID:        ratingID,
						utils.UnitFactorID:    "",
						utils.Units:           10000000000.0,
					},
					accounting2: map[string]any{
						utils.AccountID:       "1001",
						utils.AttributeIDs:    nil,
						utils.BalanceID:       "MonetaryBalance",
						utils.BalanceLimit:    0.0,
						utils.JoinedChargeIDs: nil,
						utils.RatingID:        "",
						utils.UnitFactorID:    "",
						utils.Units:           0.008333333333333333,
					},
				},
				utils.AccountsStr: map[string]any{
					"1001": map[string]any{
						utils.Balances: map[string]any{
							"MonetaryBalance": map[string]any{
								utils.AttributeIDs: nil,
								utils.Blockers:     nil,
								utils.CostIncrements: []any{
									map[string]any{
										utils.FilterIDs:    nil,
										utils.FixedFee:     nil,
										utils.Increment:    1000000000.0,
										utils.RecurrentFee: nil,
									},
								},
								utils.FilterIDs:      nil,
								utils.ID:             "MonetaryBalance",
								utils.Opts:           map[string]any{},
								utils.RateProfileIDs: []interface{}{"RP_2"},
								utils.Type:           "*concrete",
								utils.UnitFactors:    nil,
								utils.Units:          9.991666666666667,
								utils.Weights: []any{
									map[string]any{
										utils.FilterIDs: nil,
										utils.Weight:    10.0,
									},
								},
							},
						},
						utils.Blockers:     nil,
						utils.FilterIDs:    []interface{}{"*string:~*req.Account:1001"},
						utils.ID:           "1001",
						utils.Opts:         map[string]any{},
						utils.Tenant:       "cgrates.org",
						utils.ThresholdIDs: nil,
						utils.Weights: []any{
							map[string]any{
								utils.FilterIDs: nil,
								utils.Weight:    10.0,
							},
						},
					},
				},
				utils.Charges: []any{
					map[string]any{
						utils.ChargingID:     chargingID,
						utils.CompressFactor: 1.0,
					},
				},
				utils.Concretes: 0.008333333333333333,
				utils.Rates: map[string]any{
					accRateID: map[string]any{
						utils.FixedFee:      nil,
						utils.Increment:     nil,
						utils.IntervalStart: nil,
						utils.RecurrentFee:  nil,
						utils.Unit:          nil,
					},
				},
				utils.Rating: map[string]any{
					ratingID: map[string]any{
						utils.CompressFactor: 1.0,
						utils.Increments: []any{
							map[string]any{
								utils.CompressFactor:    1.0,
								utils.IncrementStart:    nil,
								utils.RateID:            accRateID,
								utils.RateIntervalIndex: 0.0,
								utils.Usage:             nil,
							},
						},
						utils.IntervalStart: nil,
					},
				},
				utils.UnitFactors: map[string]any{},
			},
			utils.MetaRates:  true,
			utils.MetaRunID:  "*default",
			utils.MetaSubsys: "*chargers",
			utils.MetaUsage:  "10000000000",
		},
	}
	if !reflect.DeepEqual(ev, exp) {
		t.Errorf("expected \n%#v\nreceived\n%#v", exp, ev)
	}
	assertCDRRow(t, db2, "cdrs2", &utils.CDR{
		Tenant: exp.Tenant,
		Opts:   exp.APIOpts,
		Event:  exp.Event,
	})

	// check if accounts were debited (should be diffenent from tariffplans)
	var acnts []*utils.Account
	if err := client.Call(context.Background(), utils.AdminSv1GetAccounts,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &acnts); err != nil {
		t.Errorf("AdminSv1GetAccounts failed unexpectedly: %v", err)
	}
	if len(acnts) != 2 {
		t.Fatalf("AdminSv1GetAccounts len(acnts)=%v, want 2", len(acnts))
	}
	sort.Slice(acnts, func(i, j int) bool {
		return acnts[i].ID > acnts[j].ID
	})
	expAccs := []*utils.Account{
		{
			Tenant:    "cgrates.org",
			ID:        "1002",
			FilterIDs: []string{"*string:~*req.Account:1002"},
			Weights: utils.DynamicWeights{
				&utils.DynamicWeight{
					Weight: 10,
				},
			},
			Opts: map[string]any{},
			Balances: map[string]*utils.Balance{
				"MonetaryBalance": {
					ID: "MonetaryBalance",
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{
							Weight: 10,
						},
					},
					Type:  utils.MetaConcrete,
					Units: utils.NewDecimalFromFloat64(10),
					Opts:  map[string]any{},
					CostIncrements: []*utils.CostIncrement{
						{
							Increment: utils.NewDecimalFromFloat64(1000000000),
						},
					},
					RateProfileIDs: []string{"RP_2"},
				},
			},
		},
		{
			Tenant:    "cgrates.org",
			ID:        "1001",
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Weights: utils.DynamicWeights{
				&utils.DynamicWeight{
					Weight: 10,
				},
			},
			Opts: map[string]any{},
			Balances: map[string]*utils.Balance{
				"MonetaryBalance": {
					ID: "MonetaryBalance",
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{
							Weight: 10,
						},
					},
					Type:  utils.MetaConcrete,
					Units: utils.NewDecimalFromFloat64(9.991666666666667), // debited
					Opts:  map[string]any{},
					CostIncrements: []*utils.CostIncrement{
						{
							Increment: utils.NewDecimalFromFloat64(1000000000),
						},
					},
					RateProfileIDs: []string{"RP_2"},
				},
			},
		},
	}
	if !reflect.DeepEqual(expAccs, acnts) {
		t.Errorf("expected \n%v\nreceived\n%v", utils.ToJSON(expAccs), utils.ToIJSON(acnts))
	}
}

func TestERSReRateEventOpts(t *testing.T) {
	cdr := &utils.CDR{ // sample with values not realisticy calculated
		Tenant: "cgrates.org",
		Opts: map[string]any{
			utils.MetaRates:      true,
			utils.MetaUR:         true, // gives event in proccessevent reply
			utils.MetaEEs:        true, // to export cdr (either or )
			utils.MetaCGRid:      urID,
			utils.MetaURID:       urID,
			utils.OptsCDRsExport: false,
			utils.MetaChargeID:   urID,
			utils.MetaChargers:   true,
			utils.MetaCost:       1.01,
			utils.MetaExporterID: "cdr_exporter",
			utils.MetaOriginID:   "oid2",
			utils.MetaRatesCost: &utils.RateProfileCost{
				ID:              "RP_1002",
				Cost:            utils.NewDecimalFromFloat64(2.3),
				MinCost:         utils.NewDecimalFromFloat64(0),
				MaxCost:         utils.NewDecimalFromFloat64(0),
				MaxCostStrategy: "",
				CostIntervals: []*utils.RateSIntervalCost{
					{
						Increments: []*utils.RateSIncrementCost{
							{
								Usage:             utils.NewDecimalFromUsageIgnoreErr("2m"),
								RateID:            "RT_WEEK",
								RateIntervalIndex: 0,
								CompressFactor:    1,
							},
							{
								Usage:             utils.NewDecimalFromUsageIgnoreErr("1s"),
								RateID:            "RT_WEEK",
								RateIntervalIndex: 1,
								CompressFactor:    60,
							},
						},
						CompressFactor: 1,
					},
				},
				Rates: map[string]*utils.IntervalRate{
					"RP_1002_LOW": {
						IntervalStart: utils.NewDecimalFromFloat64(0),
						FixedFee:      utils.NewDecimalFromFloat64(0.1),
						RecurrentFee:  utils.NewDecimalFromFloat64(0.01),
						Unit:          utils.NewDecimalFromUsageIgnoreErr("1s"),
						Increment:     utils.NewDecimalFromUsageIgnoreErr("1m"),
					},
				},
				Altered: nil,
			},
			utils.MetaRunID:  utils.MetaDefault,
			utils.MetaSubsys: utils.MetaChargers,
			utils.MetaUsage:  "10000000000",
		},
		Event: map[string]any{
			utils.OrderID:      123,
			utils.ToR:          utils.MetaVoice,
			utils.OriginID:     "oid2",
			utils.OriginHost:   "192.168.1.1",
			utils.Source:       "test",
			utils.RequestType:  utils.MetaPrepaid,
			utils.Category:     utils.Call,
			utils.AccountField: "1001",
			utils.Subject:      "1001",
			utils.Destination:  "1002",
			utils.SetupTime:    timeStart,
			utils.AnswerTime:   timeStart,
			utils.Usage:        10 * time.Second,
			utils.ExtraInfo:    "extraInfo",
			utils.ExtraFields:  map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		},
	}
	db := openTestDB(t, "cgrates2", utils.CDRsTBL, cdr)
	db2 := openTestDB(t, "cgrates2", "cdrs2")

	buf := &bytes.Buffer{}
	ng := engine.TestEngine{
		ConfigJSON: `{
  "logger": {
	"level": 7,
  },
  "admins": {
	"enabled": true,
  },
  "rates": {
	"enabled": true
  },
  "sessions": {
    "enabled": true,
    "conns": {
      "*rates": [{"connIDs": ["*localhost"]}],
      "*ees": [{"connIDs": ["*localhost"]}],
      "*chargers": [{"connIDs": ["*localhost"]}]
	}
  },
  "ees": {
	"enabled": true,
	"exporters": [{
			"id": "cdr_exporter",
			"type": "*cgrcdr",
			"exportPath": "*mysql://cgrates:CGRateS.org@127.0.0.1:3306",
        	"flags": ["*log"],
			"opts": {
				"sqlDBName": "cgrates2",
				"sqlTableName": "cdrs2"
			},
			"synchronous": true,
			"blocker": false,
			"attempts": 1,
			"failedPostsDir": "*none"
		},
	]
  },
  "ers": {
    "enabled": true,
	"conns": {
		"*ees": [{"connIDs": ["*localhost"]}],
		"*sessions": [{"connIDs": ["*localhost"]}],
	},
    "readers": [
      {
        "id": "cgrcdr",
        "runDelay": "1m",
        "type": "*cgrcdr",
        "sourcePath": "*mysql://cgrates:CGRateS.org@127.0.0.1:3306",
        "startDelay": "100ms",
        "flags": ["*log","*event"],
        "tenant": "cgrates.org",
        "opts": {
          "sqlDBName": "cgrates2",
          "sqlTableName": "cdrs",
          "sqlBatchSize": 1
        },
		// "fields":[
		// 	{"tag": "OptRates", "path": "*opts.*rates", "type": "*constant",
		// 		"value": "true"},
		// 	// {"tag": "OptEEs", "path": "*opts.*ees", "type": "*constant",
		// 	// 	"value": "true"},
		// 	{"tag": "OptUsageRecord", "path": "*opts.*ur", "type": "*constant",
		// 		"value": "true"}
		// ],
      }
    ]
  }
}`,
		DBCfg:            getDBCfg(t),
		Encoding:         *utils.Encoding,
		TpPath:           filepath.Join(*utils.DataDir, "tariffplans", "tutorial"),
		LogBuffer:        buf,
		GracefulShutdown: true,
	}
	client, _ := ng.Run(t)

	ersLogCgrCDR := "<ERs> LOG, reader: <cgrcdr>"
	waitForLog(t, buf, ersLogCgrCDR, 2*time.Second)
	if got := strings.Count(buf.String(), ersLogCgrCDR); got != 1 {
		t.Fatalf("expected 1 LOG record, got %d", got)
	}
	ev := parseCGREvent(t, buf)
	exp := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "da39a3ee5e6b4b0d3255bfef95601890afd80709",
		Event: map[string]interface{}{
			utils.AccountField: "1001",
			utils.AnswerTime:   timeStart.Format(time.RFC3339),
			utils.Category:     utils.Call,
			utils.Destination:  "1002",
			utils.ExtraFields:  map[string]any{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
			utils.ExtraInfo:    "extraInfo",
			utils.OrderID:      float64(123),
			utils.OriginHost:   "192.168.1.1",
			utils.OriginID:     "oid2",
			utils.RequestType:  "*prepaid",
			utils.SetupTime:    timeStart.Format(time.RFC3339),
			utils.Source:       "test",
			utils.Subject:      "1001",
			utils.ToR:          "*voice",
			utils.Usage:        float64(10000000000),
		},
		APIOpts: map[string]interface{}{
			utils.MetaEEs:        true,
			utils.MetaUR:         true,
			utils.MetaURID:       urID,
			utils.OptsCDRsExport: false,
			utils.MetaCGRid:      urID,
			utils.MetaChargeID:   urID,
			utils.MetaChargers:   true,
			utils.MetaCost:       1.01,
			utils.MetaExporterID: "cdr_exporter",
			utils.MetaOriginID:   "oid2",
			utils.MetaRatesCost: map[string]any{
				utils.Altered: nil,
				utils.Cost:    2.3,
				utils.CostIntervals: []any{
					map[string]any{
						utils.CompressFactor: 1.0,
						utils.Increments: []any{
							map[string]any{
								utils.CompressFactor:    1.0,
								utils.RateID:            "RT_WEEK",
								utils.RateIntervalIndex: 0.0,
								utils.Usage:             120000000000.0,
							},
							map[string]any{
								utils.CompressFactor:    60.0,
								utils.RateID:            "RT_WEEK",
								utils.RateIntervalIndex: 1.0,
								utils.Usage:             1000000000.0,
							},
						},
					},
				},
				utils.ID:              "RP_1002",
				utils.MaxCost:         0.0,
				utils.MaxCostStrategy: "",
				utils.MinCost:         0.0,
				utils.Rates: map[string]any{
					"RP_1002_LOW": map[string]any{
						utils.FixedFee:      0.1,
						utils.Increment:     60000000000.0,
						utils.IntervalStart: 0.0,
						utils.RecurrentFee:  0.01,
						utils.Unit:          1000000000.0,
					},
				},
			},
			utils.MetaRates:  true,
			utils.MetaRunID:  "*default",
			utils.MetaSubsys: "*chargers",
			utils.MetaUsage:  "10000000000",
		},
	}
	if !reflect.DeepEqual(ev, exp) {
		t.Errorf("expected \n<%#v>\nreceived\n<%#v>", exp, ev)
	}
	assertCDRRow(t, db, utils.CDRsTBL, cdr)

	waitForLog(t, buf, "<EEs> LOG, exporter <cdr_exporter>, message:", 2*time.Second)
	if got := strings.Count(buf.String(), "<EEs> LOG, exporter <cdr_exporter>, message:"); got != 1 {
		t.Fatalf("expected 1 LOG record, got %d", got)
	}
	logOutput := buf.String()
	_, after, ok := strings.Cut(logOutput, "<EEs> LOG, exporter <cdr_exporter>, message: ")
	if !ok {
		t.Fatalf("CGREvent not found in log output:\n%s", logOutput)
	}
	ev = nil
	if err := json.NewDecoder(strings.NewReader(after)).Decode(&ev); err != nil {
		t.Fatal(err)
	}
	rateID := ev.APIOpts[utils.MetaRatesCost].(map[string]any)[utils.CostIntervals].([]any)[0].(map[string]any)[utils.Increments].([]any)[0].(map[string]any)[utils.RateID].(string)
	exp = &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "da39a3ee5e6b4b0d3255bfef95601890afd80709",
		Event: map[string]interface{}{
			utils.AccountField: "1001",
			utils.AnswerTime:   timeStart.Format(time.RFC3339),
			utils.Category:     utils.Call,
			utils.Destination:  "1002",
			utils.ExtraFields:  map[string]any{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
			utils.ExtraInfo:    "extraInfo",
			utils.OrderID:      float64(123),
			utils.OriginHost:   "192.168.1.1",
			utils.OriginID:     "oid2",
			utils.RequestType:  "*prepaid",
			utils.SetupTime:    timeStart.Format(time.RFC3339),
			utils.Source:       "test",
			utils.Subject:      "1001",
			utils.ToR:          "*voice",
			utils.Usage:        float64(10000000000),
		},
		APIOpts: map[string]interface{}{
			utils.MetaEEs:        true,
			utils.MetaUR:         true,
			utils.MetaURID:       urID,
			utils.OptsCDRsExport: false,
			utils.MetaCGRid:      urID,
			utils.MetaChargeID:   urID,
			utils.MetaChargers:   true,
			utils.MetaCost:       1.01,
			utils.MetaExporterID: "cdr_exporter",
			utils.MetaOriginID:   "oid2",
			utils.MetaRatesCost: map[string]any{
				utils.Altered: nil,
				utils.Cost:    0.008333333333333333,
				utils.CostIntervals: []any{
					map[string]any{
						utils.CompressFactor: 1.0,
						utils.Increments: []any{
							map[string]any{
								utils.CompressFactor:    10.0,
								utils.RateID:            rateID,
								utils.RateIntervalIndex: 0.0,
								utils.Usage:             10000000000.0,
							},
						},
					},
				},
				utils.ID:              "RP_2",
				utils.MaxCost:         0.0,
				utils.MaxCostStrategy: "",
				utils.MinCost:         0.0,
				utils.Rates: map[string]any{
					rateID: map[string]any{
						utils.FixedFee:      0.0,
						utils.Increment:     1000000000.0,
						utils.IntervalStart: 0.0,
						utils.RecurrentFee:  0.05,
						utils.Unit:          60000000000.0,
					},
				},
			},
			utils.MetaRates:  true,
			utils.MetaRunID:  "*default",
			utils.MetaSubsys: "*chargers",
			utils.MetaUsage:  "10000000000",
		},
	}
	if !reflect.DeepEqual(ev, exp) {
		t.Errorf("expected \n%#v\nreceived\n%#v", exp, ev)
	}
	assertCDRRow(t, db2, "cdrs2", &utils.CDR{
		Tenant: exp.Tenant,
		Opts:   exp.APIOpts,
		Event:  exp.Event,
	})

	// check if anything happend to accounts (should be same as coming from tariffplans)
	// 	#Tenant,ID,FilterIDs,Weights,Blockers,Opts,BalanceID,BalanceFilterIDs,BalanceWeights,BalanceBlockers,BalanceType,BalanceUnits,BalanceUnitFactors,BalanceOpts,BalanceCostIncrements,BalanceAttributeIDs,BalanceRateProfileIDs,ThresholdIDs
	// cgrates.org,1001,*string:~*req.Account:1001,;10,,,MonetaryBalance,,;10,,*concrete,10,,,;1s;;,,RP_2,
	// cgrates.org,1002,*string:~*req.Account:1002,;10,,,MonetaryBalance,,;10,,*concrete,10,,,;1s;;,,RP_2,

	var acnts []*utils.Account
	if err := client.Call(context.Background(), utils.AdminSv1GetAccounts,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &acnts); err != nil {
		t.Errorf("AdminSv1GetAccounts failed unexpectedly: %v", err)
	}
	if len(acnts) != 2 {
		t.Fatalf("AdminSv1GetAccounts len(acnts)=%v, want 2", len(acnts))
	}
	sort.Slice(acnts, func(i, j int) bool {
		return acnts[i].ID > acnts[j].ID
	})
	expAccs := []*utils.Account{
		{
			Tenant:    "cgrates.org",
			ID:        "1002",
			FilterIDs: []string{"*string:~*req.Account:1002"},
			Weights: utils.DynamicWeights{
				&utils.DynamicWeight{
					Weight: 10,
				},
			},
			Opts: map[string]any{},
			Balances: map[string]*utils.Balance{
				"MonetaryBalance": {
					ID: "MonetaryBalance",
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{
							Weight: 10,
						},
					},
					Type:  utils.MetaConcrete,
					Units: utils.NewDecimalFromFloat64(10),
					Opts:  map[string]any{},
					CostIncrements: []*utils.CostIncrement{
						{
							Increment: utils.NewDecimalFromFloat64(1000000000),
						},
					},
					RateProfileIDs: []string{"RP_2"},
				},
			},
		},
		{
			Tenant:    "cgrates.org",
			ID:        "1001",
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Weights: utils.DynamicWeights{
				&utils.DynamicWeight{
					Weight: 10,
				},
			},
			Opts: map[string]any{},
			Balances: map[string]*utils.Balance{
				"MonetaryBalance": {
					ID: "MonetaryBalance",
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{
							Weight: 10,
						},
					},
					Type:  utils.MetaConcrete,
					Units: utils.NewDecimalFromFloat64(10),
					Opts:  map[string]any{},
					CostIncrements: []*utils.CostIncrement{
						{
							Increment: utils.NewDecimalFromFloat64(1000000000),
						},
					},
					RateProfileIDs: []string{"RP_2"},
				},
			},
		},
	}
	if !reflect.DeepEqual(expAccs, acnts) {
		t.Errorf("expected \n%v\nreceived\n%v", utils.ToJSON(expAccs), utils.ToIJSON(acnts))
	}
}

func TestERSReRateWithAccountEventOpts(t *testing.T) {
	cdr := &utils.CDR{ // sample with values not realisticy calculated
		Tenant: "cgrates.org",
		Opts: map[string]any{
			utils.MetaAccounts:   true,
			utils.MetaRates:      true,
			utils.MetaUR:         true, // gives event in proccessevent reply
			utils.MetaEEs:        true, // to export cdr (either or )
			utils.MetaCGRid:      urID,
			utils.MetaURID:       urID,
			utils.OptsCDRsExport: false,
			utils.MetaChargeID:   urID,
			utils.MetaChargers:   true,
			utils.MetaCost:       1.01,
			utils.MetaExporterID: "cdr_exporter",
			utils.MetaOriginID:   "oid2",
			utils.MetaRatesCost: &utils.RateProfileCost{
				ID:              "RP_1002",
				Cost:            utils.NewDecimalFromFloat64(2.3),
				MinCost:         utils.NewDecimalFromFloat64(0),
				MaxCost:         utils.NewDecimalFromFloat64(0),
				MaxCostStrategy: "",
				CostIntervals: []*utils.RateSIntervalCost{
					{
						Increments: []*utils.RateSIncrementCost{
							{
								Usage:             utils.NewDecimalFromUsageIgnoreErr("2m"),
								RateID:            "RT_WEEK",
								RateIntervalIndex: 0,
								CompressFactor:    1,
							},
							{
								Usage:             utils.NewDecimalFromUsageIgnoreErr("1s"),
								RateID:            "RT_WEEK",
								RateIntervalIndex: 1,
								CompressFactor:    60,
							},
						},
						CompressFactor: 1,
					},
				},
				Rates: map[string]*utils.IntervalRate{
					"RP_1002_LOW": {
						IntervalStart: utils.NewDecimalFromFloat64(0),
						FixedFee:      utils.NewDecimalFromFloat64(0.1),
						RecurrentFee:  utils.NewDecimalFromFloat64(0.01),
						Unit:          utils.NewDecimalFromUsageIgnoreErr("1s"),
						Increment:     utils.NewDecimalFromUsageIgnoreErr("1m"),
					},
				},
				Altered: nil,
			},
			utils.MetaRunID:  utils.MetaDefault,
			utils.MetaSubsys: utils.MetaChargers,
			utils.MetaUsage:  "10000000000",
		},
		Event: map[string]any{
			utils.OrderID:      123,
			utils.ToR:          utils.MetaVoice,
			utils.OriginID:     "oid2",
			utils.OriginHost:   "192.168.1.1",
			utils.Source:       "test",
			utils.RequestType:  utils.MetaPrepaid,
			utils.Category:     utils.Call,
			utils.AccountField: "1001",
			utils.Subject:      "1001",
			utils.Destination:  "1002",
			utils.SetupTime:    timeStart,
			utils.AnswerTime:   timeStart,
			utils.Usage:        10 * time.Second,
			utils.ExtraInfo:    "extraInfo",
			utils.ExtraFields:  map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		},
	}
	db := openTestDB(t, "cgrates2", utils.CDRsTBL, cdr)
	db2 := openTestDB(t, "cgrates2", "cdrs2")

	buf := &bytes.Buffer{}
	ng := engine.TestEngine{
		ConfigJSON: `{
  "logger": {
	"level": 7,
  },
  "admins": {
	"enabled": true,
  },  
  "rates": {
	"enabled": true
  },
  "accounts": {
	"enabled": true,
	"conns": {
	  "*rates": [{"connIDs": ["*localhost"]}],
	},
  },
  "sessions": {
    "enabled": true,
    "conns": {
      "*accounts": [{"connIDs": ["*localhost"]}],
      "*rates": [{"connIDs": ["*localhost"]}],
      "*ees": [{"connIDs": ["*localhost"]}],
      "*chargers": [{"connIDs": ["*localhost"]}]
	}
  },
  "ees": {
	"enabled": true,
	"exporters": [{
			"id": "cdr_exporter",
			"type": "*cgrcdr",
			"exportPath": "*mysql://cgrates:CGRateS.org@127.0.0.1:3306",
        	"flags": ["*log"],
			"opts": {
				"sqlDBName": "cgrates2",
				"sqlTableName": "cdrs2"
			},
			"synchronous": true,
			"blocker": false,
			"attempts": 1,
			"failedPostsDir": "*none"
		},
	]
  },
  "ers": {
    "enabled": true,
	"conns": {
		"*ees": [{"connIDs": ["*localhost"]}],
		"*sessions": [{"connIDs": ["*localhost"]}],
	},
    "readers": [
      {
        "id": "cgrcdr",
        "runDelay": "1m",
        "type": "*cgrcdr",
        "sourcePath": "*mysql://cgrates:CGRateS.org@127.0.0.1:3306",
        "startDelay": "100ms",
        "flags": ["*log","*event"],
        "tenant": "cgrates.org",
        "opts": {
          "sqlDBName": "cgrates2",
          "sqlTableName": "cdrs",
          "sqlBatchSize": 1
        },
		// "fields":[
		// 	{"tag": "OptAccounts", "path": "*opts.*accounts", "type": "*constant",
		// 		"value": "true"},
		// 	{"tag": "OptRates", "path": "*opts.*rates", "type": "*constant",
		// 		"value": "true"},
		// 	// {"tag": "OptEEs", "path": "*opts.*ees", "type": "*constant",
		// 	// 	"value": "true"},
		// 	{"tag": "OptUsageRecord", "path": "*opts.*ur", "type": "*constant",
		// 		"value": "true"}
		// ],
      }
    ]
  }
}`,
		DBCfg:            getDBCfg(t),
		Encoding:         *utils.Encoding,
		TpPath:           filepath.Join(*utils.DataDir, "tariffplans", "tutorial"),
		LogBuffer:        buf,
		GracefulShutdown: true,
	}
	client, _ := ng.Run(t)

	ersLogCgrCDR := "<ERs> LOG, reader: <cgrcdr>"
	waitForLog(t, buf, ersLogCgrCDR, 2*time.Second)
	if got := strings.Count(buf.String(), ersLogCgrCDR); got != 1 {
		t.Fatalf("expected 1 LOG record, got %d", got)
	}
	ev := parseCGREvent(t, buf)
	exp := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "da39a3ee5e6b4b0d3255bfef95601890afd80709",
		Event: map[string]interface{}{
			utils.AccountField: "1001",
			utils.AnswerTime:   timeStart.Format(time.RFC3339),
			utils.Category:     utils.Call,
			utils.Destination:  "1002",
			utils.ExtraFields:  map[string]any{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
			utils.ExtraInfo:    "extraInfo",
			utils.OrderID:      float64(123),
			utils.OriginHost:   "192.168.1.1",
			utils.OriginID:     "oid2",
			utils.RequestType:  "*prepaid",
			utils.SetupTime:    timeStart.Format(time.RFC3339),
			utils.Source:       "test",
			utils.Subject:      "1001",
			utils.ToR:          "*voice",
			utils.Usage:        float64(10000000000),
		},
		APIOpts: map[string]interface{}{
			utils.MetaAccounts:   true,
			utils.MetaEEs:        true,
			utils.MetaUR:         true,
			utils.MetaURID:       urID,
			utils.OptsCDRsExport: false,
			utils.MetaCGRid:      urID,
			utils.MetaChargeID:   urID,
			utils.MetaChargers:   true,
			utils.MetaCost:       1.01,
			utils.MetaExporterID: "cdr_exporter",
			utils.MetaOriginID:   "oid2",
			utils.MetaRatesCost: map[string]any{
				utils.Altered: nil,
				utils.Cost:    2.3,
				utils.CostIntervals: []any{
					map[string]any{
						utils.CompressFactor: 1.0,
						utils.Increments: []any{
							map[string]any{
								utils.CompressFactor:    1.0,
								utils.RateID:            "RT_WEEK",
								utils.RateIntervalIndex: 0.0,
								utils.Usage:             120000000000.0,
							},
							map[string]any{
								utils.CompressFactor:    60.0,
								utils.RateID:            "RT_WEEK",
								utils.RateIntervalIndex: 1.0,
								utils.Usage:             1000000000.0,
							},
						},
					},
				},
				utils.ID:              "RP_1002",
				utils.MaxCost:         0.0,
				utils.MaxCostStrategy: "",
				utils.MinCost:         0.0,
				utils.Rates: map[string]any{
					"RP_1002_LOW": map[string]any{
						utils.FixedFee:      0.1,
						utils.Increment:     60000000000.0,
						utils.IntervalStart: 0.0,
						utils.RecurrentFee:  0.01,
						utils.Unit:          1000000000.0,
					},
				},
			},
			utils.MetaRates:  true,
			utils.MetaRunID:  "*default",
			utils.MetaSubsys: "*chargers",
			utils.MetaUsage:  "10000000000",
		},
	}
	if !reflect.DeepEqual(ev, exp) {
		t.Errorf("expected \n<%#v>\nreceived\n<%#v>", exp, ev)
	}
	assertCDRRow(t, db, utils.CDRsTBL, cdr)

	waitForLog(t, buf, "<EEs> LOG, exporter <cdr_exporter>, message:", 2*time.Second)
	if got := strings.Count(buf.String(), "<EEs> LOG, exporter <cdr_exporter>, message:"); got != 1 {
		t.Fatalf("expected 1 LOG record, got %d", got)
	}
	logOutput := buf.String()
	_, after, ok := strings.Cut(logOutput, "<EEs> LOG, exporter <cdr_exporter>, message: ")
	if !ok {
		t.Fatalf("CGREvent not found in log output:\n%s", logOutput)
	}
	ev = nil
	if err := json.NewDecoder(strings.NewReader(after)).Decode(&ev); err != nil {
		t.Fatal(err)
	}
	rateID := ev.APIOpts[utils.MetaRatesCost].(map[string]any)[utils.CostIntervals].([]any)[0].(map[string]any)[utils.Increments].([]any)[0].(map[string]any)[utils.RateID].(string)
	var accounting1, accounting2, joinedChargeIDs, ratingID, accRateID, chargingID string
	accounting := ev.APIOpts[utils.MetaAccountsCost].(map[string]any)[utils.Accounting].(map[string]any)
	chargingID = ev.APIOpts[utils.MetaAccountsCost].(map[string]any)[utils.Charges].([]any)[0].(map[string]any)[utils.ChargingID].(string)
	for key := range ev.APIOpts[utils.MetaAccountsCost].(map[string]any)[utils.Rates].(map[string]any) {
		accRateID = key
	}
	for key := range accounting {
		if accounting[key].(map[string]any)["BalanceID"] == "*mockabstract" {
			accounting1 = key
			joinedChargeIDs = accounting[key].(map[string]any)["JoinedChargeIDs"].([]interface{})[0].(string)
			ratingID = accounting[key].(map[string]any)["RatingID"].(string)
		} else {
			accounting2 = key
		}
	}
	exp = &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "da39a3ee5e6b4b0d3255bfef95601890afd80709",
		Event: map[string]interface{}{
			utils.AccountField: "1001",
			utils.AnswerTime:   timeStart.Format(time.RFC3339),
			utils.Category:     utils.Call,
			utils.Destination:  "1002",
			utils.ExtraFields:  map[string]any{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
			utils.ExtraInfo:    "extraInfo",
			utils.OrderID:      float64(123),
			utils.OriginHost:   "192.168.1.1",
			utils.OriginID:     "oid2",
			utils.RequestType:  "*prepaid",
			utils.SetupTime:    timeStart.Format(time.RFC3339),
			utils.Source:       "test",
			utils.Subject:      "1001",
			utils.ToR:          "*voice",
			utils.Usage:        float64(10000000000),
		},
		APIOpts: map[string]interface{}{
			utils.MetaAccounts:   true,
			utils.MetaEEs:        true,
			utils.MetaUR:         true,
			utils.MetaURID:       urID,
			utils.OptsCDRsExport: false,
			utils.MetaCGRid:      urID,
			utils.MetaChargeID:   urID,
			utils.MetaChargers:   true,
			utils.MetaCost:       1.01,
			utils.MetaExporterID: "cdr_exporter",
			utils.MetaOriginID:   "oid2",
			utils.MetaRatesCost: map[string]any{
				utils.Altered: nil,
				utils.Cost:    0.008333333333333333,
				utils.CostIntervals: []any{
					map[string]any{
						utils.CompressFactor: 1.0,
						utils.Increments: []any{
							map[string]any{
								utils.CompressFactor:    10.0,
								utils.RateID:            rateID,
								utils.RateIntervalIndex: 0.0,
								utils.Usage:             10000000000.0,
							},
						},
					},
				},
				utils.ID:              "RP_2",
				utils.MaxCost:         0.0,
				utils.MaxCostStrategy: "",
				utils.MinCost:         0.0,
				utils.Rates: map[string]any{
					rateID: map[string]any{
						utils.FixedFee:      0.0,
						utils.Increment:     1000000000.0,
						utils.IntervalStart: 0.0,
						utils.RecurrentFee:  0.05,
						utils.Unit:          60000000000.0,
					},
				},
			},
			utils.MetaAccountsCost: map[string]any{
				utils.Abstracts: 10000000000.0,
				utils.Accounting: map[string]any{
					accounting1: map[string]any{
						utils.AccountID:       "1001",
						utils.AttributeIDs:    nil,
						utils.BalanceID:       "*mockabstract",
						utils.BalanceLimit:    nil,
						utils.JoinedChargeIDs: []interface{}{joinedChargeIDs},
						utils.RatingID:        ratingID,
						utils.UnitFactorID:    "",
						utils.Units:           10000000000.0,
					},
					accounting2: map[string]any{
						utils.AccountID:       "1001",
						utils.AttributeIDs:    nil,
						utils.BalanceID:       "MonetaryBalance",
						utils.BalanceLimit:    0.0,
						utils.JoinedChargeIDs: nil,
						utils.RatingID:        "",
						utils.UnitFactorID:    "",
						utils.Units:           0.008333333333333333,
					},
				},
				utils.AccountsStr: map[string]any{
					"1001": map[string]any{
						utils.Balances: map[string]any{
							"MonetaryBalance": map[string]any{
								utils.AttributeIDs: nil,
								utils.Blockers:     nil,
								utils.CostIncrements: []any{
									map[string]any{
										utils.FilterIDs:    nil,
										utils.FixedFee:     nil,
										utils.Increment:    1000000000.0,
										utils.RecurrentFee: nil,
									},
								},
								utils.FilterIDs:      nil,
								utils.ID:             "MonetaryBalance",
								utils.Opts:           map[string]any{},
								utils.RateProfileIDs: []interface{}{"RP_2"},
								utils.Type:           "*concrete",
								utils.UnitFactors:    nil,
								utils.Units:          9.991666666666667,
								utils.Weights: []any{
									map[string]any{
										utils.FilterIDs: nil,
										utils.Weight:    10.0,
									},
								},
							},
						},
						utils.Blockers:     nil,
						utils.FilterIDs:    []interface{}{"*string:~*req.Account:1001"},
						utils.ID:           "1001",
						utils.Opts:         map[string]any{},
						utils.Tenant:       "cgrates.org",
						utils.ThresholdIDs: nil,
						utils.Weights: []any{
							map[string]any{
								utils.FilterIDs: nil,
								utils.Weight:    10.0,
							},
						},
					},
				},
				utils.Charges: []any{
					map[string]any{
						utils.ChargingID:     chargingID,
						utils.CompressFactor: 1.0,
					},
				},
				utils.Concretes: 0.008333333333333333,
				utils.Rates: map[string]any{
					accRateID: map[string]any{
						utils.FixedFee:      nil,
						utils.Increment:     nil,
						utils.IntervalStart: nil,
						utils.RecurrentFee:  nil,
						utils.Unit:          nil,
					},
				},
				utils.Rating: map[string]any{
					ratingID: map[string]any{
						utils.CompressFactor: 1.0,
						utils.Increments: []any{
							map[string]any{
								utils.CompressFactor:    1.0,
								utils.IncrementStart:    nil,
								utils.RateID:            accRateID,
								utils.RateIntervalIndex: 0.0,
								utils.Usage:             nil,
							},
						},
						utils.IntervalStart: nil,
					},
				},
				utils.UnitFactors: map[string]any{},
			},
			utils.MetaRates:  true,
			utils.MetaRunID:  "*default",
			utils.MetaSubsys: "*chargers",
			utils.MetaUsage:  "10000000000",
		},
	}
	if !reflect.DeepEqual(ev, exp) {
		t.Errorf("expected \n%#v\nreceived\n%#v", exp, ev)
	}
	assertCDRRow(t, db2, "cdrs2", &utils.CDR{
		Tenant: exp.Tenant,
		Opts:   exp.APIOpts,
		Event:  exp.Event,
	})

	// check if accounts were debited (should be diffenent from tariffplans)
	var acnts []*utils.Account
	if err := client.Call(context.Background(), utils.AdminSv1GetAccounts,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &acnts); err != nil {
		t.Errorf("AdminSv1GetAccounts failed unexpectedly: %v", err)
	}
	if len(acnts) != 2 {
		t.Fatalf("AdminSv1GetAccounts len(acnts)=%v, want 2", len(acnts))
	}
	sort.Slice(acnts, func(i, j int) bool {
		return acnts[i].ID > acnts[j].ID
	})
	expAccs := []*utils.Account{
		{
			Tenant:    "cgrates.org",
			ID:        "1002",
			FilterIDs: []string{"*string:~*req.Account:1002"},
			Weights: utils.DynamicWeights{
				&utils.DynamicWeight{
					Weight: 10,
				},
			},
			Opts: map[string]any{},
			Balances: map[string]*utils.Balance{
				"MonetaryBalance": {
					ID: "MonetaryBalance",
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{
							Weight: 10,
						},
					},
					Type:  utils.MetaConcrete,
					Units: utils.NewDecimalFromFloat64(10),
					Opts:  map[string]any{},
					CostIncrements: []*utils.CostIncrement{
						{
							Increment: utils.NewDecimalFromFloat64(1000000000),
						},
					},
					RateProfileIDs: []string{"RP_2"},
				},
			},
		},
		{
			Tenant:    "cgrates.org",
			ID:        "1001",
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Weights: utils.DynamicWeights{
				&utils.DynamicWeight{
					Weight: 10,
				},
			},
			Opts: map[string]any{},
			Balances: map[string]*utils.Balance{
				"MonetaryBalance": {
					ID: "MonetaryBalance",
					Weights: utils.DynamicWeights{
						&utils.DynamicWeight{
							Weight: 10,
						},
					},
					Type:  utils.MetaConcrete,
					Units: utils.NewDecimalFromFloat64(9.991666666666667), // debited
					Opts:  map[string]any{},
					CostIncrements: []*utils.CostIncrement{
						{
							Increment: utils.NewDecimalFromFloat64(1000000000),
						},
					},
					RateProfileIDs: []string{"RP_2"},
				},
			},
		},
	}
	if !reflect.DeepEqual(expAccs, acnts) {
		t.Errorf("expected \n%v\nreceived\n%v", utils.ToJSON(expAccs), utils.ToIJSON(acnts))
	}
}
