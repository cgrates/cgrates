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

/*
import (
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)


func TestV1CDRsAsCDR(t *testing.T) {
	cc := &engine.CallCost{
		Category:    "call",
		Account:     "1001",
		Subject:     "1001",
		Tenant:      "cgrates.org",
		Direction:   utils.OUT,
		Destination: "1003",
		Timespans: []*engine.TimeSpan{
			&engine.TimeSpan{
				TimeStart:     time.Date(2016, 4, 6, 13, 30, 0, 0, time.UTC),
				TimeEnd:       time.Date(2016, 4, 6, 13, 31, 30, 0, time.UTC),
				DurationIndex: 0,
				RateInterval: &engine.RateInterval{
					Rating: &engine.RIRate{
						Rates: engine.RateGroups{
							&engine.Rate{
								GroupIntervalStart: 0,
								Value:              0.01,
								RateIncrement:      10 * time.Second,
								RateUnit:           time.Second}}}},
			},
		},
		TOR: utils.VOICE}

	v1Cdr := &v1Cdrs{CGRID: utils.Sha1("testprepaid1", time.Date(2016, 4, 6, 13, 29, 24, 0, time.UTC).String()),
		ToR: utils.VOICE, OriginID: "testprepaid1", OriginHost: "192.168.1.1",
		Source: "TEST_PREPAID_CDR_SMCOST1", RequestType: utils.META_PREPAID, Tenant: "cgrates.org",
		RunID:    utils.META_DEFAULT,
		Category: "call", Account: "1001", Subject: "1001", Destination: "1003",
		SetupTime:   time.Date(2016, 4, 6, 13, 29, 24, 0, time.UTC),
		AnswerTime:  time.Date(2016, 4, 6, 13, 30, 0, 0, time.UTC),
		Usage:       time.Duration(90) * time.Second,
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		CostDetails: cc}

	cdr := v1Cdr.V1toV2Cdr()
	// Create manually EventCost here
		ev := &enbine.EventCost{
			CGRID:         v1Cdr.CGRID,
			RunID:         v1Cdr.RunID,
			StartTime:     v1Cdr.AnswerTime,
			Charges:       []*engine.ChargingInterval{},
			Rating:        map[string]*RatingUnit{},
			Accounting:    map[string]*BalanceCharge{},
			RatingFilters: map[string]RatingMatchedFilters{},
			Rates:         :map[string]RateGroups{},
			Timings:       :map[string]*ChargedTiming{},
		}

	eCDR := &engine.CDR{CGRID: utils.Sha1("testprepaid1", time.Date(2016, 4, 6, 13, 29, 24, 0, time.UTC).String()),
		ToR: utils.VOICE, OriginID: "testprepaid1", OriginHost: "192.168.1.1",
		Source: "TEST_PREPAID_CDR_SMCOST1", RequestType: utils.META_PREPAID, Tenant: "cgrates.org",
		RunID:    utils.META_DEFAULT,
		Category: "call", Account: "1001", Subject: "1001", Destination: "1003",
		SetupTime:   time.Date(2016, 4, 6, 13, 29, 24, 0, time.UTC),
		AnswerTime:  time.Date(2016, 4, 6, 13, 30, 0, 0, time.UTC),
		Usage:       time.Duration(90) * time.Second,
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		CostDetails: engine.NewEventCostFromCallCost(v1Cdr.CostDetails, v1Cdr.CGRID, v1Cdr.RunID)}

	if !reflect.DeepEqual(cdr, eCDR) {
		t.Errorf("Expecting: %+v, received: %+v", cdr, eCDR)
	}

}
*/
