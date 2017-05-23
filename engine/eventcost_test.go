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
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"
)

func TestNewEventCostFromCallCost(t *testing.T) {
	acntSummary := &AccountSummary{
		Tenant: "cgrates.org",
		ID:     "dan",
		BalanceSummaries: []*BalanceSummary{
			&BalanceSummary{
				Type:     "*monetary",
				Value:    50,
				Disabled: false},
			&BalanceSummary{
				ID:       "4b8b53d7-c1a1-4159-b845-4623a00a0165",
				Type:     "*monetary",
				Value:    25,
				Disabled: false},
			&BalanceSummary{
				Type:     "*voice",
				Value:    200,
				Disabled: false,
			},
		},
		AllowNegative: false,
		Disabled:      false,
	}
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
				Cost:      0.6,
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
		AccountSummary: acntSummary,
	}

	eEC := &EventCost{
		CGRID:     "164b0422fdc6a5117031b427439482c6a4f90e41",
		RunID:     utils.META_DEFAULT,
		StartTime: time.Date(2017, 1, 9, 16, 18, 21, 0, time.UTC),
		Cost:      utils.Float64Pointer(0.85),
		Usage:     utils.DurationPointer(time.Duration(2 * time.Minute)),
		Charges: []*ChargingInterval{
			&ChargingInterval{
				RatingUUID: "f2518464-68b8-42f4-acec-aef23d714314",
				Increments: []*ChargingIncrement{
					&ChargingIncrement{
						Usage:             time.Duration(0),
						Cost:              0.1,
						BalanceChargeUUID: "44e97dec-8a7e-43d0-8b0a-736d46b5613e",
						CompressFactor:    1,
					},
					&ChargingIncrement{
						Usage:             time.Duration(1 * time.Second),
						Cost:              0,
						BalanceChargeUUID: "a555cde8-4bd0-408a-afbc-c3ba64888927",
						CompressFactor:    30,
					},
					&ChargingIncrement{
						Usage:             time.Duration(1 * time.Second),
						Cost:              0.005,
						BalanceChargeUUID: "906bfd0f-035c-40a3-93a8-46f71627983e",
						CompressFactor:    30,
					},
				},
				CompressFactor: 1,
				usage:          utils.DurationPointer(time.Duration(60 * time.Second)),
				cost:           utils.Float64Pointer(0.25),
				ecUsageIdx:     utils.DurationPointer(time.Duration(0)),
			},
			&ChargingInterval{
				RatingUUID: "f2518464-68b8-42f4-acec-aef23d714314",
				Increments: []*ChargingIncrement{
					&ChargingIncrement{
						Usage:             time.Duration(1 * time.Second),
						Cost:              0.01,
						BalanceChargeUUID: "c890a899-df43-497a-9979-38492713f57b",
						CompressFactor:    60,
					},
				},
				CompressFactor: 1,
				usage:          utils.DurationPointer(time.Duration(60 * time.Second)),
				cost:           utils.Float64Pointer(0.6),
				ecUsageIdx:     utils.DurationPointer(time.Duration(60 * time.Second)),
			},
		},
		Rating: Rating{
			"4607d907-02c3-4f2b-bc08-95a0dcc7222c": &RatingUnit{
				RoundingMethod:    "*up",
				RoundingDecimals:  5,
				TimingUUID:        "27f1e5f8-05bb-4f1c-a596-bf1010ad296c",
				RatesUUID:         "e5eb0f1c-3612-4e8c-b749-7f8f41dd90d4",
				RatingFiltersUUID: "7e73a00d-be53-4083-a1ee-8ee0b546c62a",
			},
			"f2518464-68b8-42f4-acec-aef23d714314": &RatingUnit{
				ConnectFee:        0.1,
				RoundingMethod:    "*up",
				RoundingDecimals:  5,
				TimingUUID:        "27f1e5f8-05bb-4f1c-a596-bf1010ad296c",
				RatesUUID:         "6504fb84-6b27-47a8-a1c6-c0d843959f89",
				RatingFiltersUUID: "7e73a00d-be53-4083-a1ee-8ee0b546c62a",
			},
		},
		Accounting: Accounting{
			"c890a899-df43-497a-9979-38492713f57b": &BalanceCharge{
				AccountID:   "cgrates.org:dan",
				BalanceUUID: "8c54a9e9-d610-4c82-bcb5-a315b9a65010",
				Units:       0.01,
			},
			"a894f8f1-206a-4457-99ce-df21a0c7fedc": &BalanceCharge{
				AccountID:   "cgrates.org:dan",
				BalanceUUID: "8c54a9e9-d610-4c82-bcb5-a315b9a65010",
				Units:       0.005,
			},
			"44e97dec-8a7e-43d0-8b0a-736d46b5613e": &BalanceCharge{
				AccountID:   "cgrates.org:dan",
				BalanceUUID: "8c54a9e9-d610-4c82-bcb5-a315b9a65010",
				Units:       0.1,
			},
			"906bfd0f-035c-40a3-93a8-46f71627983e": &BalanceCharge{
				AccountID:       "cgrates.org:dan",
				BalanceUUID:     "7a54a9e9-d610-4c82-bcb5-a315b9a65010",
				RatingUUID:      "4607d907-02c3-4f2b-bc08-95a0dcc7222c",
				Units:           1,
				ExtraChargeUUID: "a894f8f1-206a-4457-99ce-df21a0c7fedc",
			},
			"a555cde8-4bd0-408a-afbc-c3ba64888927": &BalanceCharge{
				AccountID:       "cgrates.org:dan",
				BalanceUUID:     "9d54a9e9-d610-4c82-bcb5-a315b9a65089",
				Units:           1,
				ExtraChargeUUID: "*none",
			},
		},
		RatingFilters: RatingFilters{
			"7e73a00d-be53-4083-a1ee-8ee0b546c62a": RatingMatchedFilters{
				"DestinationID":     "GERMANY",
				"DestinationPrefix": "+49",
				"RatingPlanID":      "RPL_RETAIL1",
				"Subject":           "*out:cgrates.org:call:*any",
			},
		},
		Rates: ChargedRates{
			"6504fb84-6b27-47a8-a1c6-c0d843959f89": RateGroups{
				&Rate{
					GroupIntervalStart: time.Duration(0),
					Value:              0.01,
					RateIncrement:      time.Duration(1 * time.Minute),
					RateUnit:           time.Duration(1 * time.Second)},
			},
			"e5eb0f1c-3612-4e8c-b749-7f8f41dd90d4": RateGroups{
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
			"27f1e5f8-05bb-4f1c-a596-bf1010ad296c": &ChargedTiming{
				StartTime: "00:00:00",
			},
		},
		AccountSummary: acntSummary,
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
	ec.ComputeEventCostUsageIndexes()
	for i := range ec.Charges {
		// Make sure main rating is correct
		if cc.Timespans[i].RateInterval.Rating != nil &&
			!reflect.DeepEqual(cc.Timespans[i].RateInterval.Rating.Rates,
				ec.Rates[ec.Rating[ec.Charges[i].RatingUUID].RatesUUID]) {
			t.Errorf("Index: %d, expecting: %s, received: %s",
				i, utils.ToJSON(cc.Timespans[i].RateInterval.Rating.Rates),
				ec.Rates[ec.Rating[ec.Charges[i].RatingUUID].RatesUUID])
		}
		// Make sure it matches also the expected rates
		if !reflect.DeepEqual(eEC.Rates[eEC.Rating[eEC.Charges[i].RatingUUID].RatesUUID],
			ec.Rates[ec.Rating[ec.Charges[i].RatingUUID].RatesUUID]) {
			t.Errorf("Index: %d, expecting: %s, received: %s", i,
				utils.ToJSON(eEC.Rates[eEC.Rating[eEC.Charges[i].RatingUUID].RatesUUID]),
				utils.ToJSON(ec.Rates[ec.Rating[ec.Charges[i].RatingUUID].RatesUUID]))
		}
		if len(eEC.Charges[i].Increments) != len(ec.Charges[i].Increments) {
			t.Errorf("Index %d, expecting: %+v, received: %+v",
				i, eEC.Charges[i].Increments, ec.Charges[i].Increments)
		}
		if !reflect.DeepEqual(eEC.Charges[i].Usage(), ec.Charges[i].Usage()) {
			t.Errorf("Expecting: %v, received: %v",
				eEC.Charges[i].Usage(), ec.Charges[i].Usage())
		}
		if !reflect.DeepEqual(eEC.Charges[i].Cost(), ec.Charges[i].Cost()) {
			t.Errorf("Expecting: %f, received: %f",
				eEC.Charges[i].Cost(), ec.Charges[i].Cost())
		}
		if !reflect.DeepEqual(eEC.Charges[i].ecUsageIdx, ec.Charges[i].ecUsageIdx) {
			t.Errorf("Expecting: %v, received: %v",
				eEC.Charges[i].ecUsageIdx, ec.Charges[i].ecUsageIdx)
		}
		cIlStartTime := ec.Charges[i].StartTime(ec.StartTime)
		if !cc.Timespans[i].TimeStart.Equal(cIlStartTime) {
			t.Errorf("Expecting: %v, received: %v",
				cc.Timespans[i].TimeStart, cIlStartTime)
		}
		if !cc.Timespans[i].TimeEnd.Equal(ec.Charges[i].EndTime(cIlStartTime)) {
			t.Errorf("Expecting: %v, received: %v",
				cc.Timespans[i].TimeStart, ec.Charges[i].EndTime(cIlStartTime))
		}
	}
	if len(ec.Rating) != len(eEC.Rating) {
		t.Errorf("Expecting: %+v, received: %+v", eEC, ec)
	}
	// Compare to original timestamp
	if !reflect.DeepEqual(cc.Timespans[0].Increments[2].BalanceInfo.Unit.RateInterval.Rating.Rates,
		ec.Rates[ec.Rating[ec.Accounting[ec.Charges[0].Increments[2].BalanceChargeUUID].RatingUUID].RatesUUID]) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(cc.Timespans[0].Increments[2].BalanceInfo.Unit.RateInterval.Rating.Rates),
			utils.ToJSON(ec.Rates[ec.Rating[ec.Accounting[ec.Charges[0].Increments[2].BalanceChargeUUID].RatingUUID].RatesUUID]))
	}
	// Compare to expected EC
	if !reflect.DeepEqual(eEC.Rates[eEC.Rating[eEC.Accounting[eEC.Charges[0].Increments[2].BalanceChargeUUID].RatingUUID].RatesUUID],
		ec.Rates[ec.Rating[ec.Accounting[ec.Charges[0].Increments[2].BalanceChargeUUID].RatingUUID].RatesUUID]) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eEC.Rates[eEC.Rating[eEC.Accounting[eEC.Charges[0].Increments[2].BalanceChargeUUID].RatingUUID].RatesUUID]),
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
	if !reflect.DeepEqual(eEC.AccountSummary, ec.AccountSummary) {
		t.Errorf("Expecting: %+v, received: %+v", eEC, ec)
	}
}

func TestEventCostAsCallCost(t *testing.T) {
	acntSummary := &AccountSummary{
		Tenant: "cgrates.org",
		ID:     "dan",
		BalanceSummaries: []*BalanceSummary{
			&BalanceSummary{
				Type:     "*monetary",
				Value:    50,
				Disabled: false},
			&BalanceSummary{
				ID:       "4b8b53d7-c1a1-4159-b845-4623a00a0165",
				Type:     "*monetary",
				Value:    25,
				Disabled: false},
			&BalanceSummary{
				Type:     "*voice",
				Value:    200,
				Disabled: false,
			},
		},
		AllowNegative: false,
		Disabled:      false,
	}
	ec := &EventCost{
		CGRID:     "164b0422fdc6a5117031b427439482c6a4f90e41",
		RunID:     utils.META_DEFAULT,
		StartTime: time.Date(2017, 1, 9, 16, 18, 21, 0, time.UTC),
		Cost:      utils.Float64Pointer(0.85),
		Usage:     utils.DurationPointer(time.Duration(2 * time.Minute)),
		Charges: []*ChargingInterval{
			&ChargingInterval{
				RatingUUID: "f2518464-68b8-42f4-acec-aef23d714314",
				Increments: []*ChargingIncrement{
					&ChargingIncrement{
						Usage:             time.Duration(0),
						Cost:              0.1,
						BalanceChargeUUID: "44e97dec-8a7e-43d0-8b0a-736d46b5613e",
						CompressFactor:    1,
					},
					&ChargingIncrement{
						Usage:             time.Duration(1 * time.Second),
						Cost:              0,
						BalanceChargeUUID: "a555cde8-4bd0-408a-afbc-c3ba64888927",
						CompressFactor:    30,
					},
					&ChargingIncrement{
						Usage:             time.Duration(1 * time.Second),
						Cost:              0.005,
						BalanceChargeUUID: "906bfd0f-035c-40a3-93a8-46f71627983e",
						CompressFactor:    30,
					},
				},
				CompressFactor: 1,
			},
			&ChargingInterval{
				RatingUUID: "f2518464-68b8-42f4-acec-aef23d714314",
				Increments: []*ChargingIncrement{
					&ChargingIncrement{
						Usage:             time.Duration(1 * time.Second),
						Cost:              0.01,
						BalanceChargeUUID: "c890a899-df43-497a-9979-38492713f57b",
						CompressFactor:    60,
					},
				},
				CompressFactor: 1,
			},
		},
		AccountSummary: acntSummary,
		Rating: Rating{
			"4607d907-02c3-4f2b-bc08-95a0dcc7222c": &RatingUnit{
				RoundingMethod:    "*up",
				RoundingDecimals:  5,
				TimingUUID:        "27f1e5f8-05bb-4f1c-a596-bf1010ad296c",
				RatesUUID:         "e5eb0f1c-3612-4e8c-b749-7f8f41dd90d4",
				RatingFiltersUUID: "7e73a00d-be53-4083-a1ee-8ee0b546c62a",
			},
			"f2518464-68b8-42f4-acec-aef23d714314": &RatingUnit{
				ConnectFee:        0.1,
				RoundingMethod:    "*up",
				RoundingDecimals:  5,
				TimingUUID:        "27f1e5f8-05bb-4f1c-a596-bf1010ad296c",
				RatesUUID:         "6504fb84-6b27-47a8-a1c6-c0d843959f89",
				RatingFiltersUUID: "7e73a00d-be53-4083-a1ee-8ee0b546c62a",
			},
		},
		Accounting: Accounting{
			"c890a899-df43-497a-9979-38492713f57b": &BalanceCharge{
				AccountID:   "cgrates.org:dan",
				BalanceUUID: "8c54a9e9-d610-4c82-bcb5-a315b9a65010",
				Units:       0.01,
			},
			"a894f8f1-206a-4457-99ce-df21a0c7fedc": &BalanceCharge{
				AccountID:   "cgrates.org:dan",
				BalanceUUID: "8c54a9e9-d610-4c82-bcb5-a315b9a65010",
				Units:       0.005,
			},
			"44e97dec-8a7e-43d0-8b0a-736d46b5613e": &BalanceCharge{
				AccountID:   "cgrates.org:dan",
				BalanceUUID: "8c54a9e9-d610-4c82-bcb5-a315b9a65010",
				Units:       0.1,
			},
			"906bfd0f-035c-40a3-93a8-46f71627983e": &BalanceCharge{
				AccountID:       "cgrates.org:dan",
				BalanceUUID:     "7a54a9e9-d610-4c82-bcb5-a315b9a65010",
				RatingUUID:      "4607d907-02c3-4f2b-bc08-95a0dcc7222c",
				Units:           1,
				ExtraChargeUUID: "a894f8f1-206a-4457-99ce-df21a0c7fedc",
			},
			"a555cde8-4bd0-408a-afbc-c3ba64888927": &BalanceCharge{
				AccountID:       "cgrates.org:dan",
				BalanceUUID:     "9d54a9e9-d610-4c82-bcb5-a315b9a65089",
				Units:           1,
				ExtraChargeUUID: "*none",
			},
		},
		RatingFilters: RatingFilters{
			"7e73a00d-be53-4083-a1ee-8ee0b546c62a": RatingMatchedFilters{
				"DestinationID":     "GERMANY",
				"DestinationPrefix": "+49",
				"RatingPlanID":      "RPL_RETAIL1",
				"Subject":           "*out:cgrates.org:call:*any",
			},
		},
		Rates: ChargedRates{
			"6504fb84-6b27-47a8-a1c6-c0d843959f89": RateGroups{
				&Rate{
					GroupIntervalStart: time.Duration(0),
					Value:              0.01,
					RateIncrement:      time.Duration(1 * time.Minute),
					RateUnit:           time.Duration(1 * time.Second)},
			},
			"e5eb0f1c-3612-4e8c-b749-7f8f41dd90d4": RateGroups{
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
			"27f1e5f8-05bb-4f1c-a596-bf1010ad296c": &ChargedTiming{
				StartTime: "00:00:00",
			},
		},
	}
	eCC := &CallCost{
		Cost:           0.85,
		RatedUsage:     120.0,
		AccountSummary: acntSummary,
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
							Monetary: &MonetaryInfo{
								UUID: "8c54a9e9-d610-4c82-bcb5-a315b9a65010"},
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
								Consumed: 1.0,
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
								Consumed: 1.0,
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
								UUID: "8c54a9e9-d610-4c82-bcb5-a315b9a65010"},
							AccountID: "cgrates.org:dan",
						},
						CompressFactor: 30,
					},
				},
			},

			&TimeSpan{
				TimeStart: time.Date(2017, 1, 9, 16, 19, 21, 0, time.UTC),
				TimeEnd:   time.Date(2017, 1, 9, 16, 20, 21, 0, time.UTC),
				Cost:      0.6,
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
							Monetary: &MonetaryInfo{
								UUID: "8c54a9e9-d610-4c82-bcb5-a315b9a65010"},
							AccountID: "cgrates.org:dan",
						},
						CompressFactor: 60,
					},
				},
			},
		},
	}
	cc := ec.AsCallCost()
	if !reflect.DeepEqual(eCC, cc) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eCC), utils.ToJSON(cc))
	}
}

func TestEventCostTrim(t *testing.T) {
	acntSummary := &AccountSummary{
		Tenant: "cgrates.org",
		ID:     "dan",
		BalanceSummaries: []*BalanceSummary{
			&BalanceSummary{
				Type:     "*monetary",
				Value:    50,
				Disabled: false},
			&BalanceSummary{
				ID:       "4b8b53d7-c1a1-4159-b845-4623a00a0165",
				Type:     "*monetary",
				Value:    25,
				Disabled: false},
			&BalanceSummary{
				Type:     "*voice",
				Value:    200,
				Disabled: false,
			},
		},
		AllowNegative: false,
		Disabled:      false,
	}
	ec := &EventCost{
		CGRID:     "164b0422fdc6a5117031b427439482c6a4f90e41",
		RunID:     utils.META_DEFAULT,
		StartTime: time.Date(2017, 1, 9, 16, 18, 21, 0, time.UTC),
		Cost:      utils.Float64Pointer(2.05),
		Usage:     utils.DurationPointer(time.Duration(4 * time.Minute)),
		Charges: []*ChargingInterval{
			&ChargingInterval{
				RatingUUID: "f2518464-68b8-42f4-acec-aef23d714314",
				Increments: []*ChargingIncrement{
					&ChargingIncrement{
						Usage:             time.Duration(0),
						Cost:              0.1,
						BalanceChargeUUID: "44e97dec-8a7e-43d0-8b0a-736d46b5613e",
						CompressFactor:    1,
					},
					&ChargingIncrement{
						Usage:             time.Duration(1 * time.Second),
						Cost:              0,
						BalanceChargeUUID: "a555cde8-4bd0-408a-afbc-c3ba64888927",
						CompressFactor:    30,
					},
					&ChargingIncrement{
						Usage:             time.Duration(1 * time.Second),
						Cost:              0.005,
						BalanceChargeUUID: "906bfd0f-035c-40a3-93a8-46f71627983e",
						CompressFactor:    30,
					},
				},
				CompressFactor: 1,
			},
			&ChargingInterval{
				RatingUUID: "f2518464-68b8-42f4-acec-aef23d714314",
				Increments: []*ChargingIncrement{
					&ChargingIncrement{
						Usage:             time.Duration(1 * time.Second),
						Cost:              0.01,
						BalanceChargeUUID: "c890a899-df43-497a-9979-38492713f57b",
						CompressFactor:    60,
					},
				},
				CompressFactor: 3,
			},
		},
		AccountSummary: acntSummary,
		Rating: Rating{
			"4607d907-02c3-4f2b-bc08-95a0dcc7222c": &RatingUnit{
				RoundingMethod:    "*up",
				RoundingDecimals:  5,
				TimingUUID:        "27f1e5f8-05bb-4f1c-a596-bf1010ad296c",
				RatesUUID:         "e5eb0f1c-3612-4e8c-b749-7f8f41dd90d4",
				RatingFiltersUUID: "7e73a00d-be53-4083-a1ee-8ee0b546c62a",
			},
			"f2518464-68b8-42f4-acec-aef23d714314": &RatingUnit{
				ConnectFee:        0.1,
				RoundingMethod:    "*up",
				RoundingDecimals:  5,
				TimingUUID:        "27f1e5f8-05bb-4f1c-a596-bf1010ad296c",
				RatesUUID:         "6504fb84-6b27-47a8-a1c6-c0d843959f89",
				RatingFiltersUUID: "7e73a00d-be53-4083-a1ee-8ee0b546c62a",
			},
		},
		Accounting: Accounting{
			"c890a899-df43-497a-9979-38492713f57b": &BalanceCharge{
				AccountID:   "cgrates.org:dan",
				BalanceUUID: "8c54a9e9-d610-4c82-bcb5-a315b9a65010",
				Units:       0.01,
			},
			"a894f8f1-206a-4457-99ce-df21a0c7fedc": &BalanceCharge{
				AccountID:   "cgrates.org:dan",
				BalanceUUID: "8c54a9e9-d610-4c82-bcb5-a315b9a65010",
				Units:       0.005,
			},
			"44e97dec-8a7e-43d0-8b0a-736d46b5613e": &BalanceCharge{
				AccountID:   "cgrates.org:dan",
				BalanceUUID: "8c54a9e9-d610-4c82-bcb5-a315b9a65010",
				Units:       0.1,
			},
			"906bfd0f-035c-40a3-93a8-46f71627983e": &BalanceCharge{
				AccountID:       "cgrates.org:dan",
				BalanceUUID:     "7a54a9e9-d610-4c82-bcb5-a315b9a65010",
				RatingUUID:      "4607d907-02c3-4f2b-bc08-95a0dcc7222c",
				Units:           1,
				ExtraChargeUUID: "a894f8f1-206a-4457-99ce-df21a0c7fedc",
			},
			"a555cde8-4bd0-408a-afbc-c3ba64888927": &BalanceCharge{
				AccountID:       "cgrates.org:dan",
				BalanceUUID:     "9d54a9e9-d610-4c82-bcb5-a315b9a65089",
				Units:           1,
				ExtraChargeUUID: "*none",
			},
		},
		RatingFilters: RatingFilters{
			"7e73a00d-be53-4083-a1ee-8ee0b546c62a": RatingMatchedFilters{
				"DestinationID":     "GERMANY",
				"DestinationPrefix": "+49",
				"RatingPlanID":      "RPL_RETAIL1",
				"Subject":           "*out:cgrates.org:call:*any",
			},
		},
		Rates: ChargedRates{
			"6504fb84-6b27-47a8-a1c6-c0d843959f89": RateGroups{
				&Rate{
					GroupIntervalStart: time.Duration(0),
					Value:              0.01,
					RateIncrement:      time.Duration(1 * time.Minute),
					RateUnit:           time.Duration(1 * time.Second)},
			},
			"e5eb0f1c-3612-4e8c-b749-7f8f41dd90d4": RateGroups{
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
			"27f1e5f8-05bb-4f1c-a596-bf1010ad296c": &ChargedTiming{
				StartTime: "00:00:00",
			},
		},
	}
	origEC := ec.Clone()
	if srplsEC, err := ec.Trim(time.Duration(4 * time.Minute)); err != nil {
		t.Error(err)
	} else if srplsEC != nil {
		t.Errorf("Expecting nil, got: %+v", srplsEC)
	}
	if srplsEC, err := ec.Trim(time.Duration(0)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(origEC, srplsEC) {
		t.Errorf("Expecting: %s,\n received: %s", utils.ToJSON(origEC), utils.ToJSON(srplsEC))
	}
}
