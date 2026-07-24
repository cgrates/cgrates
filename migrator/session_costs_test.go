// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package migrator

import (
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestV2toV3Cost(t *testing.T) {
	cc := &engine.CallCost{
		Category:    utils.CALL,
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
							&engine.Rate{
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
					Type:  utils.MONETARY,
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
