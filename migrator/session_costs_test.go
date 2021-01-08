/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

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
				&engine.BalanceSummary{
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
