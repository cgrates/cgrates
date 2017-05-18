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
package engine

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"
)

func TestNewEventCostFromCallCost(t *testing.T) {
	cc := &CallCost{
		Direction:   utils.META_OUT,
		Category:    "call",
		Tenant:      "cgrates.org",
		Subject:     "dan",
		Account:     "dan",
		Destination: "+4986517174963",
		TOR:         utils.VOICE,
		Cost:        0.85,
		RatedUsage:  120.0,
		Timespans: TimeSpans{
			&TimeSpan{
				TimeStart: time.Date(2017, 1, 9, 16, 18, 21, 0, time.UTC),
				TimeEnd:   time.Date(2017, 1, 9, 16, 19, 21, 0, time.UTC),
				Cost:      0.25,
				RateInterval: &RateInterval{ // standard rating
					Timing: &RITiming{
						StartTime: "00:00:00",
					},
					Rating: &RIRate{
						ConnectFee:       0.1,
						RoundingMethod:   "*up",
						RoundingDecimals: 5,
						Rates: RateGroups{
							&Rate{
								GroupIntervalStart: time.Duration(0),
								Value:              0.01,
								RateUnit:           time.Duration(1 * time.Second),
								RateIncrement:      time.Duration(1 * time.Minute),
							},
						},
					},
				},
				DurationIndex:  time.Duration(1 * time.Minute),
				MatchedSubject: "*out:cgrates.org:call:*any",
				MatchedPrefix:  "+49",
				MatchedDestId:  "GERMANY",
				RatingPlanId:   "RPL_RETAIL1",
				CompressFactor: 1,
				Increments: Increments{
					&Increment{ // ConnectFee
						Cost: 0.1,
						BalanceInfo: &DebitInfo{
							Monetary: &MonetaryInfo{UUID: "8c54a9e9-d610-4c82-bcb5-a315b9a65010",
								ID:    utils.META_DEFAULT,
								Value: 9.9},
							AccountID: "cgrates.org:dan",
						},
						CompressFactor: 1,
					},
					&Increment{ // First 30 seconds free
						Duration: time.Duration(1 * time.Second),
						Cost:     0,
						BalanceInfo: &DebitInfo{
							Unit: &UnitInfo{
								UUID:     "9d54a9e9-d610-4c82-bcb5-a315b9a65089",
								ID:       "free_mins",
								Value:    0,
								Consumed: 1.0,
								TOR:      utils.VOICE,
							},
							AccountID: "cgrates.org:dan",
						},
						CompressFactor: 30,
					},
					&Increment{ // Minutes with special price
						Duration: time.Duration(1 * time.Second),
						Cost:     0.005,
						BalanceInfo: &DebitInfo{
							Unit: &UnitInfo{ // Minutes with special price
								UUID:     "7a54a9e9-d610-4c82-bcb5-a315b9a65010",
								ID:       "discounted_mins",
								Value:    0,
								Consumed: 1.0,
								TOR:      utils.VOICE,
								RateInterval: &RateInterval{
									Timing: &RITiming{
										StartTime: "00:00:00",
									},
									Rating: &RIRate{
										ConnectFee:       0,
										RoundingMethod:   "*up",
										RoundingDecimals: 5,
										Rates: RateGroups{
											&Rate{
												GroupIntervalStart: time.Duration(0),
												Value:              0.005,
												RateUnit:           time.Duration(1 * time.Second),
												RateIncrement:      time.Duration(1 * time.Second),
											},
											&Rate{
												GroupIntervalStart: time.Duration(60 * time.Second),
												Value:              0.005,
												RateUnit:           time.Duration(1 * time.Second),
												RateIncrement:      time.Duration(1 * time.Second),
											},
										},
									},
								},
							},
							Monetary: &MonetaryInfo{
								UUID:  "8c54a9e9-d610-4c82-bcb5-a315b9a65010",
								ID:    utils.META_DEFAULT,
								Value: 9.75},
							AccountID: "cgrates.org:dan",
						},
						CompressFactor: 30,
					},
				},
			},

			&TimeSpan{
				TimeStart: time.Date(2017, 1, 9, 16, 19, 21, 0, time.UTC),
				TimeEnd:   time.Date(2017, 1, 9, 16, 20, 21, 0, time.UTC),
				Cost:      0.01,
				RateInterval: &RateInterval{ // standard rating
					Timing: &RITiming{
						StartTime: "00:00:00",
					},
					Rating: &RIRate{
						ConnectFee:       0.1,
						RoundingMethod:   "*up",
						RoundingDecimals: 5,
						Rates: RateGroups{
							&Rate{
								GroupIntervalStart: time.Duration(0),
								Value:              0.01,
								RateUnit:           time.Duration(1 * time.Second),
								RateIncrement:      time.Duration(1 * time.Minute),
							},
						},
					},
				},
				DurationIndex:  time.Duration(1 * time.Minute),
				MatchedSubject: "*out:cgrates.org:call:*any",
				MatchedPrefix:  "+49",
				MatchedDestId:  "GERMANY",
				RatingPlanId:   "RPL_RETAIL1",
				CompressFactor: 1,
				Increments: Increments{
					&Increment{
						Cost:     0.01,
						Duration: time.Duration(1 * time.Second),
						BalanceInfo: &DebitInfo{
							Monetary: &MonetaryInfo{UUID: "8c54a9e9-d610-4c82-bcb5-a315b9a65010",
								ID:    utils.META_DEFAULT,
								Value: 9.15},
							AccountID: "cgrates.org:dan",
						},
						CompressFactor: 60,
					},
				},
			},
		},
	}

	eEC := &EventCost{
		CGRID: "164b0422fdc6a5117031b427439482c6a4f90e41",
		RunID: utils.META_DEFAULT,
		Cost:  utils.Float64Pointer(0.85),
		Usage: utils.DurationPointer(time.Duration(2 * time.Minute)),
		Charges: []*ChargingInterval{
			&ChargingInterval{
				StartTime:  time.Date(2017, 1, 9, 16, 18, 21, 0, time.UTC),
				RatingUUID: "bebf80cf-cba5-4e36-89dc-86673cff8cc4",
				Increments: []*ChargingIncrement{
					&ChargingIncrement{
						Usage:             time.Duration(0),
						Cost:              0.1,
						BalanceChargeUUID: "716a278d-9ca5-451a-aa59-b6a43f4fb4ef",
						CompressFactor:    1,
					},
					&ChargingIncrement{
						Usage:             time.Duration(1 * time.Second),
						Cost:              0,
						BalanceChargeUUID: "8ee1f8ee-5783-487b-87e3-cb1bb6fd8f9f",
						CompressFactor:    30,
					},
					&ChargingIncrement{
						Usage:             time.Duration(1 * time.Second),
						Cost:              0.005,
						BalanceChargeUUID: "77c904d4-c579-4687-8c28-a1561e39dae2",
						CompressFactor:    30,
					},
				},
				CompressFactor: 1,
			},
			&ChargingInterval{
				StartTime:  time.Date(2017, 1, 9, 16, 19, 21, 0, time.UTC),
				RatingUUID: "bebf80cf-cba5-4e36-89dc-86673cff8cc4",
				Increments: []*ChargingIncrement{
					&ChargingIncrement{
						Usage:             time.Duration(0),
						Cost:              0.01,
						BalanceChargeUUID: "79463f6e-d70f-41ac-9345-76bd21714759",
						CompressFactor:    60,
					},
				},
				CompressFactor: 1,
			},
		},
		Rating: Rating{
			"bebf80cf-cba5-4e36-89dc-86673cff8cc4": &RatingUnit{
				ConnectFee:        0.1,
				RoundingMethod:    "*up",
				RoundingDecimals:  5,
				TimingUUID:        "3e4c7dd1-10f9-4fdc-b7df-8833724933dd",
				RatesUUID:         "5f04c792-5c79-4873-ba39-413342671595",
				RatingFiltersUUID: "8fa45f23-5bb1-44ee-867c-ad09b2bae981",
			},
			"2b7333c0-479c-4e5d-8d72-d089e93b2b6a": &RatingUnit{
				RoundingMethod:    "*up",
				RoundingDecimals:  5,
				TimingUUID:        "3e4c7dd1-10f9-4fdc-b7df-8833724933dd",
				RatesUUID:         "3246cb23-ef2e-4080-ba5b-45300cbede3f",
				RatingFiltersUUID: "8fa45f23-5bb1-44ee-867c-ad09b2bae981",
			},
		},
		Accounting: Accounting{
			"2afef931-eb94-46df-8fb4-3509954e771c": &BalanceCharge{
				AccountID:   "cgrates.org:dan",
				BalanceUUID: "8c54a9e9-d610-4c82-bcb5-a315b9a65010",
				Units:       0.005,
			},
			"716a278d-9ca5-451a-aa59-b6a43f4fb4ef": &BalanceCharge{
				AccountID:   "cgrates.org:dan",
				BalanceUUID: "8c54a9e9-d610-4c82-bcb5-a315b9a65010",
				Units:       0.1,
			},
			"77c904d4-c579-4687-8c28-a1561e39dae2": &BalanceCharge{
				AccountID:       "cgrates.org:dan",
				BalanceUUID:     "7a54a9e9-d610-4c82-bcb5-a315b9a65010",
				RatingUUID:      "2b7333c0-479c-4e5d-8d72-d089e93b2b6a",
				Units:           1,
				ExtraChargeUUID: "2afef931-eb94-46df-8fb4-3509954e771c",
			},
			"79463f6e-d70f-41ac-9345-76bd21714759": &BalanceCharge{
				AccountID:   "cgrates.org:dan",
				BalanceUUID: "8c54a9e9-d610-4c82-bcb5-a315b9a65010",
				Units:       0.01,
			},
			"8ee1f8ee-5783-487b-87e3-cb1bb6fd8f9f": &BalanceCharge{
				AccountID:       "cgrates.org:dan",
				BalanceUUID:     "9d54a9e9-d610-4c82-bcb5-a315b9a65089",
				Units:           1,
				ExtraChargeUUID: "*none",
			},
		},
		RatingFilters: RatingFilters{
			"8fa45f23-5bb1-44ee-867c-ad09b2bae981": RatingMatchedFilters{
				"DestinationID":     "GERMANY",
				"DestinationPrefix": "+49",
				"RatingPlanID":      "RPL_RETAIL1",
				"Subject":           "*out:cgrates.org:call:*any",
			},
		},
		Rates: ChargedRates{
			"3246cb23-ef2e-4080-ba5b-45300cbede3f": RateGroups{
				&Rate{
					GroupIntervalStart: time.Duration(0),
					Value:              0.01,
					RateIncrement:      time.Duration(1 * time.Minute),
					RateUnit:           time.Duration(1 * time.Second)},
			},
			"5f04c792-5c79-4873-ba39-413342671595": RateGroups{
				&Rate{
					GroupIntervalStart: time.Duration(0),
					Value:              0.005,
					RateIncrement:      time.Duration(1 * time.Second),
					RateUnit:           time.Duration(1 * time.Second)},
				&Rate{
					GroupIntervalStart: time.Duration(60 * time.Second),
					Value:              0.005,
					RateIncrement:      time.Duration(1 * time.Second),
					RateUnit:           time.Duration(1 * time.Second)},
			},
		},
		Timings: ChargedTimings{
			"3e4c7dd1-10f9-4fdc-b7df-8833724933dd": &ChargedTiming{
				StartTime: "00:00:00",
			},
		},
	}
	ec := NewEventCostFromCallCost(cc, "164b0422fdc6a5117031b427439482c6a4f90e41", utils.META_DEFAULT)
	if cost := ec.ComputeCost(); cost != cc.Cost {
		t.Errorf("Expecting: %f, received: %f", cc.Cost, cost)
	}
	eUsage := time.Duration(int64(cc.RatedUsage * 1000000000))
	if usage := ec.ComputeUsage(); usage != eUsage {
		t.Errorf("Expecting: %v, received: %v", eUsage, usage)
	}
	if len(ec.Charges) != len(eEC.Charges) {
		t.Errorf("Expecting: %+v, received: %+v", eEC, ec)
	}
	for i := range eEC.Charges {
		// Make sure main rating is correct
		if cc.Timespans[i].RateInterval.Rating != nil &&
			!reflect.DeepEqual(cc.Timespans[i].RateInterval.Rating.Rates, ec.Rates[ec.Rating[ec.Charges[i].RatingUUID].RatesUUID]) {
			t.Errorf("For index: %d, expecting: %s, received: %s",
				i, utils.ToJSON(cc.Timespans[i].RateInterval.Rating.Rates), ec.Rates[ec.Rating[ec.Charges[i].RatingUUID].RatesUUID])
		}
		if len(eEC.Charges[i].Increments) != len(ec.Charges[i].Increments) {
			t.Errorf("At index %d, expecting: %+v, received: %+v", eEC.Charges[i].Increments, ec.Charges[i].Increments)
		}
	}
	if len(ec.Rating) != len(eEC.Rating) {
		t.Errorf("Expecting: %+v, received: %+v", eEC, ec)
	}
	if !reflect.DeepEqual(cc.Timespans[0].Increments[2].BalanceInfo.Unit.RateInterval.Rating.Rates,
		ec.Rates[ec.Rating[ec.Accounting[ec.Charges[0].Increments[2].BalanceChargeUUID].RatingUUID].RatesUUID]) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(cc.Timespans[0].Increments[2].BalanceInfo.Unit.RateInterval.Rating.Rates),
			utils.ToJSON(ec.Rates[ec.Rating[ec.Accounting[ec.Charges[0].Increments[2].BalanceChargeUUID].RatingUUID].RatesUUID]))
	}
	if len(ec.Accounting) != len(eEC.Accounting) {
		t.Errorf("Expecting: %+v, received: %+v", eEC, ec)
	}
	if len(ec.Rates) != len(eEC.Rates) {
		t.Errorf("Expecting: %+v, received: %+v", eEC, ec)
	}
	if len(ec.Timings) != len(eEC.Timings) {
		t.Errorf("Expecting: %+v, received: %+v", eEC, ec)
	}
}

func TestEventCostAsCallCost(t *testing.T) {
	ec := &EventCost{
		CGRID: "164b0422fdc6a5117031b427439482c6a4f90e41",
		RunID: utils.META_DEFAULT,
		Cost:  utils.Float64Pointer(0.85),
		Usage: utils.DurationPointer(time.Duration(2 * time.Minute)),
		Charges: []*ChargingInterval{
			&ChargingInterval{
				StartTime:  time.Date(2017, 1, 9, 16, 18, 21, 0, time.UTC),
				RatingUUID: "bebf80cf-cba5-4e36-89dc-86673cff8cc4",
				Increments: []*ChargingIncrement{
					&ChargingIncrement{
						Usage:             time.Duration(0),
						Cost:              0.1,
						BalanceChargeUUID: "716a278d-9ca5-451a-aa59-b6a43f4fb4ef",
						CompressFactor:    1,
					},
					&ChargingIncrement{
						Usage:             time.Duration(1 * time.Second),
						Cost:              0,
						BalanceChargeUUID: "8ee1f8ee-5783-487b-87e3-cb1bb6fd8f9f",
						CompressFactor:    30,
					},
					&ChargingIncrement{
						Usage:             time.Duration(1 * time.Second),
						Cost:              0.005,
						BalanceChargeUUID: "77c904d4-c579-4687-8c28-a1561e39dae2",
						CompressFactor:    30,
					},
				},
				CompressFactor: 1,
			},
			&ChargingInterval{
				StartTime:  time.Date(2017, 1, 9, 16, 19, 21, 0, time.UTC),
				RatingUUID: "bebf80cf-cba5-4e36-89dc-86673cff8cc4",
				Increments: []*ChargingIncrement{
					&ChargingIncrement{
						Usage:             time.Duration(0),
						Cost:              0.01,
						BalanceChargeUUID: "79463f6e-d70f-41ac-9345-76bd21714759",
						CompressFactor:    60,
					},
				},
				CompressFactor: 1,
			},
		},
		Rating: Rating{
			"bebf80cf-cba5-4e36-89dc-86673cff8cc4": &RatingUnit{
				ConnectFee:        0.1,
				RoundingMethod:    "*up",
				RoundingDecimals:  5,
				TimingUUID:        "3e4c7dd1-10f9-4fdc-b7df-8833724933dd",
				RatesUUID:         "5f04c792-5c79-4873-ba39-413342671595",
				RatingFiltersUUID: "8fa45f23-5bb1-44ee-867c-ad09b2bae981",
			},
			"2b7333c0-479c-4e5d-8d72-d089e93b2b6a": &RatingUnit{
				RoundingMethod:    "*up",
				RoundingDecimals:  5,
				TimingUUID:        "3e4c7dd1-10f9-4fdc-b7df-8833724933dd",
				RatesUUID:         "3246cb23-ef2e-4080-ba5b-45300cbede3f",
				RatingFiltersUUID: "8fa45f23-5bb1-44ee-867c-ad09b2bae981",
			},
		},
		Accounting: Accounting{
			"2afef931-eb94-46df-8fb4-3509954e771c": &BalanceCharge{
				AccountID:   "cgrates.org:dan",
				BalanceUUID: "8c54a9e9-d610-4c82-bcb5-a315b9a65010",
				Units:       0.005,
			},
			"716a278d-9ca5-451a-aa59-b6a43f4fb4ef": &BalanceCharge{
				AccountID:   "cgrates.org:dan",
				BalanceUUID: "8c54a9e9-d610-4c82-bcb5-a315b9a65010",
				Units:       0.1,
			},
			"77c904d4-c579-4687-8c28-a1561e39dae2": &BalanceCharge{
				AccountID:       "cgrates.org:dan",
				BalanceUUID:     "7a54a9e9-d610-4c82-bcb5-a315b9a65010",
				RatingUUID:      "2b7333c0-479c-4e5d-8d72-d089e93b2b6a",
				Units:           1,
				ExtraChargeUUID: "2afef931-eb94-46df-8fb4-3509954e771c",
			},
			"79463f6e-d70f-41ac-9345-76bd21714759": &BalanceCharge{
				AccountID:   "cgrates.org:dan",
				BalanceUUID: "8c54a9e9-d610-4c82-bcb5-a315b9a65010",
				Units:       0.01,
			},
			"8ee1f8ee-5783-487b-87e3-cb1bb6fd8f9f": &BalanceCharge{
				AccountID:       "cgrates.org:dan",
				BalanceUUID:     "9d54a9e9-d610-4c82-bcb5-a315b9a65089",
				Units:           1,
				ExtraChargeUUID: "*none",
			},
		},
		RatingFilters: RatingFilters{
			"8fa45f23-5bb1-44ee-867c-ad09b2bae981": RatingMatchedFilters{
				"DestinationID":     "GERMANY",
				"DestinationPrefix": "+49",
				"RatingPlanID":      "RPL_RETAIL1",
				"Subject":           "*out:cgrates.org:call:*any",
			},
		},
		Rates: ChargedRates{
			"3246cb23-ef2e-4080-ba5b-45300cbede3f": RateGroups{
				&Rate{
					GroupIntervalStart: time.Duration(0),
					Value:              0.01,
					RateIncrement:      time.Duration(1 * time.Minute),
					RateUnit:           time.Duration(1 * time.Second)},
			},
			"5f04c792-5c79-4873-ba39-413342671595": RateGroups{
				&Rate{
					GroupIntervalStart: time.Duration(0),
					Value:              0.005,
					RateIncrement:      time.Duration(1 * time.Second),
					RateUnit:           time.Duration(1 * time.Second)},
				&Rate{
					GroupIntervalStart: time.Duration(60 * time.Second),
					Value:              0.005,
					RateIncrement:      time.Duration(1 * time.Second),
					RateUnit:           time.Duration(1 * time.Second)},
			},
		},
		Timings: ChargedTimings{
			"3e4c7dd1-10f9-4fdc-b7df-8833724933dd": &ChargedTiming{
				StartTime: "00:00:00",
			},
		},
	}
	eCC := &CallCost{
		Direction:   utils.META_OUT,
		Category:    "call",
		Tenant:      "cgrates.org",
		Subject:     "dan",
		Account:     "dan",
		Destination: "+4986517174963",
		TOR:         utils.VOICE,
		Cost:        0.85,
		RatedUsage:  120.0,
		Timespans: TimeSpans{
			&TimeSpan{
				TimeStart: time.Date(2017, 1, 9, 16, 18, 21, 0, time.UTC),
				TimeEnd:   time.Date(2017, 1, 9, 16, 19, 21, 0, time.UTC),
				Cost:      0.25,
				RateInterval: &RateInterval{ // standard rating
					Timing: &RITiming{
						StartTime: "00:00:00",
					},
					Rating: &RIRate{
						ConnectFee:       0.1,
						RoundingMethod:   "*up",
						RoundingDecimals: 5,
						Rates: RateGroups{
							&Rate{
								GroupIntervalStart: time.Duration(0),
								Value:              0.01,
								RateUnit:           time.Duration(1 * time.Second),
								RateIncrement:      time.Duration(1 * time.Minute),
							},
						},
					},
				},
				DurationIndex:  time.Duration(1 * time.Minute),
				MatchedSubject: "*out:cgrates.org:call:*any",
				MatchedPrefix:  "+49",
				MatchedDestId:  "GERMANY",
				RatingPlanId:   "RPL_RETAIL1",
				CompressFactor: 1,
				Increments: Increments{
					&Increment{ // ConnectFee
						Cost: 0.1,
						BalanceInfo: &DebitInfo{
							Monetary: &MonetaryInfo{UUID: "8c54a9e9-d610-4c82-bcb5-a315b9a65010",
								ID:    utils.META_DEFAULT,
								Value: 9.9},
							AccountID: "cgrates.org:dan",
						},
						CompressFactor: 1,
					},
					&Increment{ // First 30 seconds free
						Duration: time.Duration(1 * time.Second),
						Cost:     0,
						BalanceInfo: &DebitInfo{
							Unit: &UnitInfo{
								UUID:     "9d54a9e9-d610-4c82-bcb5-a315b9a65089",
								ID:       "free_mins",
								Value:    0,
								Consumed: 1.0,
								TOR:      utils.VOICE,
							},
							AccountID: "cgrates.org:dan",
						},
						CompressFactor: 30,
					},
					&Increment{ // Minutes with special price
						Duration: time.Duration(1 * time.Second),
						Cost:     0.005,
						BalanceInfo: &DebitInfo{
							Unit: &UnitInfo{ // Minutes with special price
								UUID:     "7a54a9e9-d610-4c82-bcb5-a315b9a65010",
								ID:       "discounted_mins",
								Value:    0,
								Consumed: 1.0,
								TOR:      utils.VOICE,
								RateInterval: &RateInterval{
									Timing: &RITiming{
										StartTime: "00:00:00",
									},
									Rating: &RIRate{
										ConnectFee:       0,
										RoundingMethod:   "*up",
										RoundingDecimals: 5,
										Rates: RateGroups{
											&Rate{
												GroupIntervalStart: time.Duration(0),
												Value:              0.005,
												RateUnit:           time.Duration(1 * time.Second),
												RateIncrement:      time.Duration(1 * time.Second),
											},
											&Rate{
												GroupIntervalStart: time.Duration(60 * time.Second),
												Value:              0.005,
												RateUnit:           time.Duration(1 * time.Second),
												RateIncrement:      time.Duration(1 * time.Second),
											},
										},
									},
								},
							},
							Monetary: &MonetaryInfo{
								UUID:  "8c54a9e9-d610-4c82-bcb5-a315b9a65010",
								ID:    utils.META_DEFAULT,
								Value: 9.75},
							AccountID: "cgrates.org:dan",
						},
						CompressFactor: 30,
					},
				},
			},

			&TimeSpan{
				TimeStart: time.Date(2017, 1, 9, 16, 19, 21, 0, time.UTC),
				TimeEnd:   time.Date(2017, 1, 9, 16, 20, 21, 0, time.UTC),
				Cost:      0.01,
				RateInterval: &RateInterval{ // standard rating
					Timing: &RITiming{
						StartTime: "00:00:00",
					},
					Rating: &RIRate{
						ConnectFee:       0.1,
						RoundingMethod:   "*up",
						RoundingDecimals: 5,
						Rates: RateGroups{
							&Rate{
								GroupIntervalStart: time.Duration(0),
								Value:              0.01,
								RateUnit:           time.Duration(1 * time.Second),
								RateIncrement:      time.Duration(1 * time.Minute),
							},
						},
					},
				},
				DurationIndex:  time.Duration(1 * time.Minute),
				MatchedSubject: "*out:cgrates.org:call:*any",
				MatchedPrefix:  "+49",
				MatchedDestId:  "GERMANY",
				RatingPlanId:   "RPL_RETAIL1",
				CompressFactor: 1,
				Increments: Increments{
					&Increment{
						Cost:     0.01,
						Duration: time.Duration(1 * time.Second),
						BalanceInfo: &DebitInfo{
							Monetary: &MonetaryInfo{UUID: "8c54a9e9-d610-4c82-bcb5-a315b9a65010",
								ID:    utils.META_DEFAULT,
								Value: 9.15},
							AccountID: "cgrates.org:dan",
						},
						CompressFactor: 60,
					},
				},
			},
		},
	}
	cc := ec.AsCallCost("*voice", "cgrates.org", "*out", "call", "dan", "dan", "+4986517174963")
	if len(eCC.Timespans) != len(cc.Timespans) {
		t.Errorf("Expecting: %+v, received: %+v", eCC, cc)
	}
	fmt.Printf("Expecting: %s, \nreceived : %s\n", utils.ToJSON(eCC), utils.ToJSON(cc))
}
