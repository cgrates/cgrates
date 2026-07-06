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

package tpes

import (
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"
)

func TestActionProfileRecordsSeparateTargetsAndActions(t *testing.T) {
	records := actionProfileRecords(&utils.ActionProfile{
		Tenant:    utils.CGRateSorg,
		ID:        "ap1",
		FilterIDs: []string{"fltr1"},
		Schedule:  utils.MetaASAP,
		Targets: map[string]utils.StringSet{
			utils.MetaResources: {"res1": {}},
			utils.MetaAccounts:  {"acc1": {}, "acc2": {}},
		},
		Actions: []*utils.APAction{{
			ID:   "act1",
			TTL:  time.Second,
			Type: utils.MetaSetBalance,
			Diktats: []*utils.APDiktat{
				{ID: "d1"},
				{ID: "d2"},
			},
		}},
	})
	if len(records) != 4 {
		t.Fatalf("expected 4 records, received %d", len(records))
	}
	for _, record := range records {
		if len(record) != len(actionHeader) {
			t.Fatalf("expected record length %d, received %d", len(actionHeader), len(record))
		}
	}
	if records[0][2] != "fltr1" || records[0][5] != utils.MetaASAP {
		t.Fatalf("first record should carry profile fields: %+v", records[0])
	}
	if records[1][2] != "" || records[2][2] != "" {
		t.Fatalf("only first record should carry profile fields: %+v", records)
	}
	if records[0][6] != utils.MetaAccounts || records[1][6] != utils.MetaResources {
		t.Fatalf("expected target rows before action rows: %+v", records)
	}
	if records[2][8] != "act1" || records[2][15] != "d1" || records[3][8] != "act1" || records[3][15] != "d2" {
		t.Fatalf("expected one action row per diktat: %+v", records)
	}
}

func TestAccountRecordsKeepsProfileOpts(t *testing.T) {
	records := accountRecords(&utils.Account{
		Tenant:       utils.CGRateSorg,
		ID:           "acc1",
		FilterIDs:    []string{"fltr1"},
		Opts:         map[string]any{"k1": "v1"},
		ThresholdIDs: []string{"th1"},
		Balances: map[string]*utils.Balance{
			"bal2": {ID: "bal2", Type: utils.MetaMonetary, Units: utils.NewDecimal(2, 0)},
			"bal1": {ID: "bal1", Type: utils.MetaVoice, Units: utils.NewDecimal(1, 0)},
		},
	})
	if len(records) != 2 {
		t.Fatalf("expected 2 records, received %d", len(records))
	}
	for _, record := range records {
		if len(record) != len(accountHeader) {
			t.Fatalf("expected record length %d, received %d", len(accountHeader), len(record))
		}
	}
	if records[0][5] != "k1:v1" || records[0][17] != "th1" {
		t.Fatalf("first record should carry account fields: %+v", records[0])
	}
	if records[1][5] != "" || records[1][17] != "" {
		t.Fatalf("second record should not repeat account fields: %+v", records[1])
	}
	if got := []string{records[0][6], records[1][6]}; !reflect.DeepEqual(got, []string{"bal1", "bal2"}) {
		t.Fatalf("expected sorted balances, received %+v", got)
	}
}

func TestRateProfileRecordsExpandsRateIntervals(t *testing.T) {
	records := rateProfileRecords(&utils.RateProfile{
		Tenant:          utils.CGRateSorg,
		ID:              "rp1",
		FilterIDs:       []string{"fltr1"},
		MinCost:         utils.NewDecimal(1, 0),
		MaxCost:         utils.NewDecimal(10, 0),
		MaxCostStrategy: utils.MetaMaxCostFree,
		Rates: map[string]*utils.Rate{
			"rt2": {
				ID: "rt2",
				IntervalRates: []*utils.IntervalRate{{
					IntervalStart: utils.NewDecimal(0, 0),
					FixedFee:      utils.NewDecimal(3, 0),
				}},
			},
			"rt1": {
				ID:        "rt1",
				FilterIDs: []string{"rfltr1"},
				IntervalRates: []*utils.IntervalRate{{
					IntervalStart: utils.NewDecimal(0, 0),
				}, {
					IntervalStart: utils.NewDecimal(60, 0),
				}},
			},
		},
	})
	if len(records) != 3 {
		t.Fatalf("expected 3 records, received %d", len(records))
	}
	for _, record := range records {
		if len(record) != len(rateHeader) {
			t.Fatalf("expected record length %d, received %d", len(rateHeader), len(record))
		}
	}
	if records[0][2] != "fltr1" || records[1][2] != "" {
		t.Fatalf("profile fields should only be on first rate row: %+v", records)
	}
	if got := []string{records[0][7], records[1][7], records[2][7]}; !reflect.DeepEqual(got, []string{"rt1", "rt1", "rt2"}) {
		t.Fatalf("expected sorted rate rows, received %+v", got)
	}
	if records[0][8] != "rfltr1" || records[1][8] != "" {
		t.Fatalf("rate fields should only be on first interval row: %+v", records)
	}
}
