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

package migrator

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestV2toV3Cost(t *testing.T) {
	cc := &engine.CallCost{
		Category:    utils.Call,
		Tenant:      "cgrates.org",
		Subject:     "1001",
		Account:     "1001",
		Destination: "1002",
		ToR:         "ToR",
		Cost:        10,
		Timespans: engine.TimeSpans{
			&engine.TimeSpan{
				TimeStart: time.Now(),
				TimeEnd:   time.Now().Add(time.Minute),
				Cost:      10,
				RateInterval: &engine.RateInterval{
					Rating: &engine.RIRate{
						Rates: engine.RateGroups{
							&engine.RGRate{
								GroupIntervalStart: 0,
								Value:              100,
								RateIncrement:      10 * time.Second,
								RateUnit:           time.Second,
							},
						},
					},
				},
			},
		},
		RatedUsage: 10,
		AccountSummary: &engine.AccountSummary{
			Tenant: "cgrates.org",
			ID:     "1001",
			BalanceSummaries: []*engine.BalanceSummary{
				{
					UUID:  "UUID",
					ID:    "First",
					Type:  utils.MetaMonetary,
					Value: 10,
				},
			},
		},
	}
	sv2 := v2SessionsCost{
		CGRID:       "CGRID",
		RunID:       utils.MetaDefault,
		OriginHost:  utils.FreeSWITCHAgent,
		OriginID:    "Origin1",
		CostSource:  utils.MetaSessionS,
		Usage:       time.Second,
		CostDetails: cc,
	}
	sv3 := &engine.SMCost{
		CGRID:       "CGRID",
		RunID:       utils.MetaDefault,
		OriginHost:  utils.FreeSWITCHAgent,
		OriginID:    "Origin1",
		Usage:       time.Second,
		CostSource:  utils.MetaSessionS,
		CostDetails: engine.NewEventCostFromCallCost(cc, "CGRID", utils.MetaDefault),
	}
	rply := sv2.V2toV3Cost()
	rply.CostDetails = sv3.CostDetails
	if !reflect.DeepEqual(sv3, rply) {
		t.Errorf("Expected: %s ,received: %s", utils.ToJSON(sv3), utils.ToJSON(rply))
	}
}

func TestAsSessionsCostSql(t *testing.T) {

	costDetails := &engine.CallCost{Cost: 100.0}
	v2Cost := &v2SessionsCost{
		CGRID:       "test-cgrid",
		RunID:       "test-runid",
		OriginHost:  "test-originhost",
		OriginID:    "test-originid",
		CostSource:  "test-costsource",
		CostDetails: costDetails,
		Usage:       5 * time.Second,
	}

	result := v2Cost.AsSessionsCostSql()

	if result.Cgrid != v2Cost.CGRID {
		t.Errorf("expected Cgrid %v, got %v", v2Cost.CGRID, result.Cgrid)
	}
	if result.RunID != v2Cost.RunID {
		t.Errorf("expected RunID %v, got %v", v2Cost.RunID, result.RunID)
	}
	if result.OriginHost != v2Cost.OriginHost {
		t.Errorf("expected OriginHost %v, got %v", v2Cost.OriginHost, result.OriginHost)
	}
	if result.OriginID != v2Cost.OriginID {
		t.Errorf("expected OriginID %v, got %v", v2Cost.OriginID, result.OriginID)
	}
	if result.CostSource != v2Cost.CostSource {
		t.Errorf("expected CostSource %v, got %v", v2Cost.CostSource, result.CostSource)
	}
	expectedCostDetails := utils.ToJSON(v2Cost.CostDetails)
	if result.CostDetails != expectedCostDetails {
		t.Errorf("expected CostDetails %v, got %v", expectedCostDetails, result.CostDetails)
	}
	if result.Usage != v2Cost.Usage.Nanoseconds() {
		t.Errorf("expected Usage %v, got %v", v2Cost.Usage.Nanoseconds(), result.Usage)
	}

}

func TestNewV2SessionsCostFromSessionsCostSql(t *testing.T) {

	costDetails := &engine.CallCost{Cost: 100.0}
	costDetailsJson, _ := json.Marshal(costDetails)
	smSql := &engine.SessionCostsSQL{
		Cgrid:       "test-cgrid",
		RunID:       "test-runid",
		OriginHost:  "test-originhost",
		OriginID:    "test-originid",
		CostSource:  "test-costsource",
		CostDetails: string(costDetailsJson),
		Usage:       int64(5 * time.Second),
		CreatedAt:   time.Now(),
	}

	result, err := NewV2SessionsCostFromSessionsCostSql(smSql)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.CGRID != smSql.Cgrid {
		t.Errorf("expected CGRID %v, got %v", smSql.Cgrid, result.CGRID)
	}
	if result.RunID != smSql.RunID {
		t.Errorf("expected RunID %v, got %v", smSql.RunID, result.RunID)
	}
	if result.OriginHost != smSql.OriginHost {
		t.Errorf("expected OriginHost %v, got %v", smSql.OriginHost, result.OriginHost)
	}
	if result.OriginID != smSql.OriginID {
		t.Errorf("expected OriginID %v, got %v", smSql.OriginID, result.OriginID)
	}
	if result.CostSource != smSql.CostSource {
		t.Errorf("expected CostSource %v, got %v", smSql.CostSource, result.CostSource)
	}
	if result.Usage != time.Duration(smSql.Usage) {
		t.Errorf("expected Usage %v, got %v", time.Duration(smSql.Usage), result.Usage)
	}
	if result.CostDetails == nil {
		t.Fatalf("expected CostDetails not nil, got nil")
	} else if result.CostDetails.Cost != costDetails.Cost {
		t.Errorf("expected Cost %v, got %v", costDetails.Cost, result.CostDetails.Cost)
	}
}
