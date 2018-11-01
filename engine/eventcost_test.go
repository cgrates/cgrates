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

// testEC is used as sample through various tests
var testEC = &EventCost{
	CGRID:     "164b0422fdc6a5117031b427439482c6a4f90e41",
	RunID:     utils.META_DEFAULT,
	StartTime: time.Date(2017, 1, 9, 16, 18, 21, 0, time.UTC),
	Charges: []*ChargingInterval{
		&ChargingInterval{
			RatingID: "c1a5ab9",
			Increments: []*ChargingIncrement{
				&ChargingIncrement{
					Usage:          time.Duration(0),
					Cost:           0.1,
					AccountingID:   "9bdad10",
					CompressFactor: 1,
				},
				&ChargingIncrement{
					Usage:          time.Duration(1 * time.Second),
					Cost:           0,
					AccountingID:   "3455b83",
					CompressFactor: 10,
				},
				&ChargingIncrement{
					Usage:          time.Duration(10 * time.Second),
					Cost:           0.01,
					AccountingID:   "a012888",
					CompressFactor: 2,
				},
				&ChargingIncrement{
					Usage:          time.Duration(1 * time.Second),
					Cost:           0.005,
					AccountingID:   "44d6c02",
					CompressFactor: 30,
				},
			},
			CompressFactor: 1,
		},
		&ChargingInterval{
			RatingID: "c1a5ab9",
			Increments: []*ChargingIncrement{
				&ChargingIncrement{
					Usage:          time.Duration(1 * time.Second),
					Cost:           0.01,
					AccountingID:   "a012888",
					CompressFactor: 60,
				},
			},
			CompressFactor: 4,
		},
		&ChargingInterval{
			RatingID: "c1a5ab9",
			Increments: []*ChargingIncrement{
				&ChargingIncrement{
					Usage:          time.Duration(1 * time.Second),
					Cost:           0,
					AccountingID:   "3455b83",
					CompressFactor: 10,
				},
				&ChargingIncrement{
					Usage:          time.Duration(10 * time.Second),
					Cost:           0.01,
					AccountingID:   "a012888",
					CompressFactor: 2,
				},
				&ChargingIncrement{
					Usage:          time.Duration(1 * time.Second),
					Cost:           0.005,
					AccountingID:   "44d6c02",
					CompressFactor: 30,
				},
			},
			CompressFactor: 5,
		},
	},
	AccountSummary: &AccountSummary{
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
	},
	Rating: Rating{
		"3cd6425": &RatingUnit{
			RoundingMethod:   "*up",
			RoundingDecimals: 5,
			TimingID:         "7f324ab",
			RatesID:          "4910ecf",
			RatingFiltersID:  "43e77dc",
		},
		"c1a5ab9": &RatingUnit{
			ConnectFee:       0.1,
			RoundingMethod:   "*up",
			RoundingDecimals: 5,
			TimingID:         "7f324ab",
			RatesID:          "ec1a177",
			RatingFiltersID:  "43e77dc",
		},
	},
	Accounting: Accounting{
		"a012888": &BalanceCharge{
			AccountID:   "cgrates.org:dan",
			BalanceUUID: "8c54a9e9-d610-4c82-bcb5-a315b9a65010",
			Units:       0.01,
		},
		"188bfa6": &BalanceCharge{
			AccountID:   "cgrates.org:dan",
			BalanceUUID: "8c54a9e9-d610-4c82-bcb5-a315b9a65010",
			Units:       0.005,
		},
		"9bdad10": &BalanceCharge{
			AccountID:   "cgrates.org:dan",
			BalanceUUID: "8c54a9e9-d610-4c82-bcb5-a315b9a65010",
			Units:       0.1,
		},
		"44d6c02": &BalanceCharge{
			AccountID:     "cgrates.org:dan",
			BalanceUUID:   "7a54a9e9-d610-4c82-bcb5-a315b9a65010",
			RatingID:      "3cd6425",
			Units:         1,
			ExtraChargeID: "188bfa6",
		},
		"3455b83": &BalanceCharge{
			AccountID:     "cgrates.org:dan",
			BalanceUUID:   "9d54a9e9-d610-4c82-bcb5-a315b9a65089",
			Units:         1,
			ExtraChargeID: "*none",
		},
	},
	RatingFilters: RatingFilters{
		"43e77dc": RatingMatchedFilters{
			"DestinationID":     "GERMANY",
			"DestinationPrefix": "+49",
			"RatingPlanID":      "RPL_RETAIL1",
			"Subject":           "*out:cgrates.org:call:*any",
		},
	},
	Rates: ChargedRates{
		"ec1a177": RateGroups{
			&Rate{
				GroupIntervalStart: time.Duration(0),
				Value:              0.01,
				RateIncrement:      time.Duration(1 * time.Minute),
				RateUnit:           time.Duration(1 * time.Second)},
		},
		"4910ecf": RateGroups{
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
		"7f324ab": &ChargedTiming{
			StartTime: "00:00:00",
		},
	},
}

func TestECClone(t *testing.T) {
	ec := testEC.Clone()
	if !reflect.DeepEqual(testEC, ec) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(testEC), utils.ToJSON(ec))
	}
	// making sure we don't influence the original values
	ec.Usage = utils.DurationPointer(time.Duration(1 * time.Second))
	if testEC.Usage != nil {
		t.Error("Usage is not nil")
	}
	ec.Cost = utils.Float64Pointer(1.0)
	if testEC.Cost != nil {
		t.Error("Cost is not nil")
	}
	ec.Charges[0].Increments[0].Cost = 1.0
	if testEC.Charges[0].Increments[0].Cost == 1.0 {
		t.Error("Cost is 1.0")
	}
	ec.AccountSummary.Disabled = true
	if testEC.AccountSummary.Disabled {
		t.Error("Account is disabled")
	}
	ec.AccountSummary.BalanceSummaries[0].Value = 5.0
	if testEC.AccountSummary.BalanceSummaries[0].Value == 5.0 {
		t.Error("Wrong balance summary")
	}
	ec.Rates["ec1a177"][0].Value = 5.0
	if testEC.Rates["ec1a177"][0].Value == 5.0 {
		t.Error("Wrong  Value")
	}
	delete(ec.Rates, "ec1a177")
	if _, has := testEC.Rates["ec1a177"]; !has {
		t.Error("Key removed from testEC")
	}
	ec.Timings["7f324ab"].StartTime = "10:00:00"
	if testEC.Timings["7f324ab"].StartTime == "10:00:00" {
		t.Error("Wrong StartTime")
	}
	delete(ec.Timings, "7f324ab")
	if _, has := testEC.Timings["7f324ab"]; !has {
		t.Error("Key removed from testEC")
	}
	ec.RatingFilters["43e77dc"]["DestinationID"] = "GERMANY_MOBILE"
	if testEC.RatingFilters["43e77dc"]["DestinationID"] == "GERMANY_MOBILE" {
		t.Error("Wrong DestinationID")
	}
	delete(ec.RatingFilters, "43e77dc")
	if _, has := testEC.RatingFilters["43e77dc"]; !has {
		t.Error("Key removed from testEC")
	}
	ec.Accounting["a012888"].Units = 5.0
	if testEC.Accounting["a012888"].Units == 5.0 {
		t.Error("Wrong Units")
	}
	delete(ec.Accounting, "a012888")
	if _, has := testEC.Accounting["a012888"]; !has {
		t.Error("Key removed from testEC")
	}

}

func TestECComputeAndReset(t *testing.T) {
	ec := testEC.Clone()
	eEc := testEC.Clone()
	eEc.Usage = utils.DurationPointer(time.Duration(10 * time.Minute))
	eEc.Cost = utils.Float64Pointer(3.52)
	eEc.Charges[0].ecUsageIdx = utils.DurationPointer(time.Duration(0))
	eEc.Charges[0].usage = utils.DurationPointer(time.Duration(1 * time.Minute))
	eEc.Charges[0].cost = utils.Float64Pointer(0.27)
	eEc.Charges[1].ecUsageIdx = utils.DurationPointer(time.Duration(1 * time.Minute))
	eEc.Charges[1].usage = utils.DurationPointer(time.Duration(1 * time.Minute))
	eEc.Charges[1].cost = utils.Float64Pointer(0.6)
	eEc.Charges[2].ecUsageIdx = utils.DurationPointer(time.Duration(5 * time.Minute))
	eEc.Charges[2].usage = utils.DurationPointer(time.Duration(1 * time.Minute))
	eEc.Charges[2].cost = utils.Float64Pointer(0.17)
	ec.Compute()
	if !reflect.DeepEqual(eEc, ec) {
		t.Errorf("Expecting: %s\n, received: %s", utils.ToJSON(eEc), utils.ToJSON(ec))
	}
	ec.ResetCounters()
	if !reflect.DeepEqual(testEC, ec) {
		t.Errorf("Expecting: %+v, received: %+v", testEC, ec)
	}
}

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
				RatingID: "f2518464-68b8-42f4-acec-aef23d714314",
				Increments: []*ChargingIncrement{
					&ChargingIncrement{
						Usage:          time.Duration(0),
						Cost:           0.1,
						AccountingID:   "44e97dec-8a7e-43d0-8b0a-736d46b5613e",
						CompressFactor: 1,
					},
					&ChargingIncrement{
						Usage:          time.Duration(1 * time.Second),
						Cost:           0,
						AccountingID:   "a555cde8-4bd0-408a-afbc-c3ba64888927",
						CompressFactor: 30,
					},
					&ChargingIncrement{
						Usage:          time.Duration(1 * time.Second),
						Cost:           0.005,
						AccountingID:   "906bfd0f-035c-40a3-93a8-46f71627983e",
						CompressFactor: 30,
					},
				},
				CompressFactor: 1,
				usage:          utils.DurationPointer(time.Duration(60 * time.Second)),
				cost:           utils.Float64Pointer(0.25),
				ecUsageIdx:     utils.DurationPointer(time.Duration(0)),
			},
			&ChargingInterval{
				RatingID: "f2518464-68b8-42f4-acec-aef23d714314",
				Increments: []*ChargingIncrement{
					&ChargingIncrement{
						Usage:          time.Duration(1 * time.Second),
						Cost:           0.01,
						AccountingID:   "c890a899-df43-497a-9979-38492713f57b",
						CompressFactor: 60,
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
				RoundingMethod:   "*up",
				RoundingDecimals: 5,
				TimingID:         "27f1e5f8-05bb-4f1c-a596-bf1010ad296c",
				RatesID:          "e5eb0f1c-3612-4e8c-b749-7f8f41dd90d4",
				RatingFiltersID:  "7e73a00d-be53-4083-a1ee-8ee0b546c62a",
			},
			"f2518464-68b8-42f4-acec-aef23d714314": &RatingUnit{
				ConnectFee:       0.1,
				RoundingMethod:   "*up",
				RoundingDecimals: 5,
				TimingID:         "27f1e5f8-05bb-4f1c-a596-bf1010ad296c",
				RatesID:          "6504fb84-6b27-47a8-a1c6-c0d843959f89",
				RatingFiltersID:  "7e73a00d-be53-4083-a1ee-8ee0b546c62a",
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
				AccountID:     "cgrates.org:dan",
				BalanceUUID:   "7a54a9e9-d610-4c82-bcb5-a315b9a65010",
				RatingID:      "4607d907-02c3-4f2b-bc08-95a0dcc7222c",
				Units:         1,
				ExtraChargeID: "a894f8f1-206a-4457-99ce-df21a0c7fedc",
			},
			"a555cde8-4bd0-408a-afbc-c3ba64888927": &BalanceCharge{
				AccountID:     "cgrates.org:dan",
				BalanceUUID:   "9d54a9e9-d610-4c82-bcb5-a315b9a65089",
				Units:         1,
				ExtraChargeID: "*none",
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
	if cost := ec.GetCost(); cost != cc.Cost {
		t.Errorf("Expecting: %f, received: %f", cc.Cost, cost)
	}
	eUsage := time.Duration(int64(cc.RatedUsage * 1000000000))
	if usage := ec.GetUsage(); usage != eUsage {
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
				ec.Rates[ec.Rating[ec.Charges[i].RatingID].RatesID]) {
			t.Errorf("Index: %d, expecting: %+v, received: %+v",
				i, utils.ToJSON(cc.Timespans[i].RateInterval.Rating.Rates),
				ec.Rates[ec.Rating[ec.Charges[i].RatingID].RatesID])
		}
		// Make sure it matches also the expected rates
		if !reflect.DeepEqual(eEC.Rates[eEC.Rating[eEC.Charges[i].RatingID].RatesID],
			ec.Rates[ec.Rating[ec.Charges[i].RatingID].RatesID]) {
			t.Errorf("Index: %d, expecting: %s, received: %s", i,
				utils.ToJSON(eEC.Rates[eEC.Rating[eEC.Charges[i].RatingID].RatesID]),
				utils.ToJSON(ec.Rates[ec.Rating[ec.Charges[i].RatingID].RatesID]))
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
		ec.Rates[ec.Rating[ec.Accounting[ec.Charges[0].Increments[2].AccountingID].RatingID].RatesID]) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(cc.Timespans[0].Increments[2].BalanceInfo.Unit.RateInterval.Rating.Rates),
			utils.ToJSON(ec.Rates[ec.Rating[ec.Accounting[ec.Charges[0].Increments[2].AccountingID].RatingID].RatesID]))
	}
	// Compare to expected EC
	if !reflect.DeepEqual(eEC.Rates[eEC.Rating[eEC.Accounting[eEC.Charges[0].Increments[2].AccountingID].RatingID].RatesID],
		ec.Rates[ec.Rating[ec.Accounting[ec.Charges[0].Increments[2].AccountingID].RatingID].RatesID]) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eEC.Rates[eEC.Rating[eEC.Accounting[eEC.Charges[0].Increments[2].AccountingID].RatingID].RatesID]),
			utils.ToJSON(ec.Rates[ec.Rating[ec.Accounting[ec.Charges[0].Increments[2].AccountingID].RatingID].RatesID]))
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

func TestECAsCallCost(t *testing.T) {
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
				RatingID: "f2518464-68b8-42f4-acec-aef23d714314",
				Increments: []*ChargingIncrement{
					&ChargingIncrement{
						Usage:          time.Duration(0),
						Cost:           0.1,
						AccountingID:   "44e97dec-8a7e-43d0-8b0a-736d46b5613e",
						CompressFactor: 1,
					},
					&ChargingIncrement{
						Usage:          time.Duration(1 * time.Second),
						Cost:           0,
						AccountingID:   "a555cde8-4bd0-408a-afbc-c3ba64888927",
						CompressFactor: 30,
					},
					&ChargingIncrement{
						Usage:          time.Duration(1 * time.Second),
						Cost:           0.005,
						AccountingID:   "906bfd0f-035c-40a3-93a8-46f71627983e",
						CompressFactor: 30,
					},
				},
				CompressFactor: 1,
			},
			&ChargingInterval{
				RatingID: "f2518464-68b8-42f4-acec-aef23d714314",
				Increments: []*ChargingIncrement{
					&ChargingIncrement{
						Usage:          time.Duration(1 * time.Second),
						Cost:           0.01,
						AccountingID:   "c890a899-df43-497a-9979-38492713f57b",
						CompressFactor: 60,
					},
				},
				CompressFactor: 1,
			},
		},
		AccountSummary: acntSummary,
		Rating: Rating{
			"4607d907-02c3-4f2b-bc08-95a0dcc7222c": &RatingUnit{
				RoundingMethod:   "*up",
				RoundingDecimals: 5,
				TimingID:         "27f1e5f8-05bb-4f1c-a596-bf1010ad296c",
				RatesID:          "e5eb0f1c-3612-4e8c-b749-7f8f41dd90d4",
				RatingFiltersID:  "7e73a00d-be53-4083-a1ee-8ee0b546c62a",
			},
			"f2518464-68b8-42f4-acec-aef23d714314": &RatingUnit{
				ConnectFee:       0.1,
				RoundingMethod:   "*up",
				RoundingDecimals: 5,
				TimingID:         "27f1e5f8-05bb-4f1c-a596-bf1010ad296c",
				RatesID:          "6504fb84-6b27-47a8-a1c6-c0d843959f89",
				RatingFiltersID:  "7e73a00d-be53-4083-a1ee-8ee0b546c62a",
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
				AccountID:     "cgrates.org:dan",
				BalanceUUID:   "7a54a9e9-d610-4c82-bcb5-a315b9a65010",
				RatingID:      "4607d907-02c3-4f2b-bc08-95a0dcc7222c",
				Units:         1,
				ExtraChargeID: "a894f8f1-206a-4457-99ce-df21a0c7fedc",
			},
			"a555cde8-4bd0-408a-afbc-c3ba64888927": &BalanceCharge{
				AccountID:     "cgrates.org:dan",
				BalanceUUID:   "9d54a9e9-d610-4c82-bcb5-a315b9a65089",
				Units:         1,
				ExtraChargeID: "*none",
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
		RatedUsage:     120000000000,
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

func TestECTrimZeroAndFull(t *testing.T) {
	ec := testEC.Clone()
	if srplsEC, err := ec.Trim(time.Duration(10 * time.Minute)); err != nil {
		t.Error(err)
	} else if srplsEC != nil {
		t.Errorf("Expecting nil, got: %+v", srplsEC)
	}
	eFullSrpls := testEC.Clone()
	eFullSrpls.Usage = utils.DurationPointer(time.Duration(10 * time.Minute))
	eFullSrpls.Charges[0].usage = utils.DurationPointer(time.Duration(1 * time.Minute))
	eFullSrpls.Charges[1].usage = utils.DurationPointer(time.Duration(1 * time.Minute))
	eFullSrpls.Charges[2].usage = utils.DurationPointer(time.Duration(1 * time.Minute))
	if srplsEC, err := ec.Trim(time.Duration(0)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eFullSrpls, srplsEC) {
		t.Errorf("\tExpecting: %s,\n\treceived: %s",
			utils.ToJSON(eFullSrpls), utils.ToJSON(srplsEC))
	}
}

func TestECTrimMiddle1(t *testing.T) {
	// trim in the middle of increments
	ec := testEC.Clone()
	eEC := testEC.Clone()
	eEC.Charges = []*ChargingInterval{
		&ChargingInterval{
			RatingID: "c1a5ab9",
			Increments: []*ChargingIncrement{
				&ChargingIncrement{
					Usage:          time.Duration(0),
					Cost:           0.1,
					AccountingID:   "9bdad10",
					CompressFactor: 1,
				},
				&ChargingIncrement{
					Usage:          time.Duration(1 * time.Second),
					Cost:           0,
					AccountingID:   "3455b83",
					CompressFactor: 10,
				},
				&ChargingIncrement{
					Usage:          time.Duration(10 * time.Second),
					Cost:           0.01,
					AccountingID:   "a012888",
					CompressFactor: 2,
				},
				&ChargingIncrement{
					Usage:          time.Duration(1 * time.Second),
					Cost:           0.005,
					AccountingID:   "44d6c02",
					CompressFactor: 30,
				},
			},
			CompressFactor: 1,
		},
		&ChargingInterval{
			RatingID: "c1a5ab9",
			Increments: []*ChargingIncrement{
				&ChargingIncrement{
					Usage:          time.Duration(1 * time.Second),
					Cost:           0.01,
					AccountingID:   "a012888",
					CompressFactor: 60,
				},
			},
			CompressFactor: 2,
		},
		&ChargingInterval{
			RatingID: "c1a5ab9",
			Increments: []*ChargingIncrement{
				&ChargingIncrement{
					Usage:          time.Duration(1 * time.Second),
					Cost:           0.01,
					AccountingID:   "a012888",
					CompressFactor: 10,
				},
			},
			CompressFactor: 1,
		},
	}
	eSrplsEC := testEC.Clone()
	eSrplsEC.StartTime = time.Date(2017, 1, 9, 16, 21, 31, 0, time.UTC)
	eSrplsEC.Charges = []*ChargingInterval{
		&ChargingInterval{
			RatingID: "c1a5ab9",
			Increments: []*ChargingIncrement{
				&ChargingIncrement{
					Usage:          time.Duration(1 * time.Second),
					Cost:           0.01,
					AccountingID:   "a012888",
					CompressFactor: 50,
				},
			},
			CompressFactor: 1,
		},
		&ChargingInterval{
			RatingID: "c1a5ab9",
			Increments: []*ChargingIncrement{
				&ChargingIncrement{
					Usage:          time.Duration(1 * time.Second),
					Cost:           0.01,
					AccountingID:   "a012888",
					CompressFactor: 60,
				},
			},
			CompressFactor: 1,
		},
		&ChargingInterval{
			RatingID: "c1a5ab9",
			Increments: []*ChargingIncrement{
				&ChargingIncrement{
					Usage:          time.Duration(1 * time.Second),
					Cost:           0,
					AccountingID:   "3455b83",
					CompressFactor: 10,
				},
				&ChargingIncrement{
					Usage:          time.Duration(10 * time.Second),
					Cost:           0.01,
					AccountingID:   "a012888",
					CompressFactor: 2,
				},
				&ChargingIncrement{
					Usage:          time.Duration(1 * time.Second),
					Cost:           0.005,
					AccountingID:   "44d6c02",
					CompressFactor: 30,
				},
			},
			CompressFactor: 5,
		},
	}

	reqDuration := time.Duration(190 * time.Second)
	srplsEC, err := ec.Trim(reqDuration)
	if err != nil {
		t.Error(err)
	}
	if reqDuration != *ec.Usage {
		t.Errorf("Expecting request duration: %v, received: %v", reqDuration, *ec.Usage)
	}
	eSrplsDur := time.Duration(410 * time.Second)
	if srplsUsage := srplsEC.GetUsage(); srplsUsage != eSrplsDur {
		t.Errorf("Expecting surplus duration: %v, received: %v", eSrplsDur, srplsUsage)
	}
	ec.ResetCounters()
	srplsEC.ResetCounters()
	if !reflect.DeepEqual(eEC, ec) {
		t.Errorf("Expecting: %s\n, received: %s", utils.ToIJSON(eEC), utils.ToIJSON(ec))
	}
	// test surplus, which is not easy to estimate due it's different item ids
	if !eSrplsEC.StartTime.Equal(srplsEC.StartTime) ||
		len(eSrplsEC.Charges) != len(srplsEC.Charges) {
		t.Errorf("Expecting: \n%s, received: \n%s", utils.ToIJSON(eSrplsEC), utils.ToIJSON(srplsEC))
	}
}

// TestECTrimMUsage is targeting simpler testing of the durations trimmed/remainders
// using subtests so we can cover the tests with less code
func TestECTrimMUsage(t *testing.T) {
	// each subtest will trim at some usage duration
	testCases := []struct {
		atUsage    time.Duration
		ecUsage    time.Duration
		ecCost     float64
		srplsUsage time.Duration
		srplsCost  float64
	}{
		{time.Duration(5 * time.Second), time.Duration(5 * time.Second), 0.1,
			time.Duration(595 * time.Second), 3.42},
		{time.Duration(10 * time.Second), time.Duration(10 * time.Second), 0.1,
			time.Duration(590 * time.Second), 3.42},
		{time.Duration(15 * time.Second), time.Duration(20 * time.Second), 0.11,
			time.Duration(580 * time.Second), 3.41},
		{time.Duration(20 * time.Second), time.Duration(20 * time.Second), 0.11,
			time.Duration(580 * time.Second), 3.41},
		{time.Duration(25 * time.Second), time.Duration(30 * time.Second), 0.12,
			time.Duration(570 * time.Second), 3.40},
		{time.Duration(38 * time.Second), time.Duration(38 * time.Second), 0.16,
			time.Duration(562 * time.Second), 3.36},
		{time.Duration(60 * time.Second), time.Duration(60 * time.Second), 0.27,
			time.Duration(540 * time.Second), 3.25},
		{time.Duration(62 * time.Second), time.Duration(62 * time.Second), 0.29,
			time.Duration(538 * time.Second), 3.23},
		{time.Duration(120 * time.Second), time.Duration(120 * time.Second), 0.87,
			time.Duration(480 * time.Second), 2.65},
		{time.Duration(121 * time.Second), time.Duration(121 * time.Second), 0.88,
			time.Duration(479 * time.Second), 2.64},
		{time.Duration(180 * time.Second), time.Duration(180 * time.Second), 1.47,
			time.Duration(420 * time.Second), 2.05},
		{time.Duration(250 * time.Second), time.Duration(250 * time.Second), 2.17,
			time.Duration(350 * time.Second), 1.35},
		{time.Duration(299 * time.Second), time.Duration(299 * time.Second), 2.66,
			time.Duration(301 * time.Second), 0.86},
		{time.Duration(300 * time.Second), time.Duration(300 * time.Second), 2.67,
			time.Duration(300 * time.Second), 0.85},
		{time.Duration(302 * time.Second), time.Duration(302 * time.Second), 2.67,
			time.Duration(298 * time.Second), 0.85},
		{time.Duration(310 * time.Second), time.Duration(310 * time.Second), 2.67,
			time.Duration(290 * time.Second), 0.85},
		{time.Duration(316 * time.Second), time.Duration(320 * time.Second), 2.68,
			time.Duration(280 * time.Second), 0.84},
		{time.Duration(320 * time.Second), time.Duration(320 * time.Second), 2.68,
			time.Duration(280 * time.Second), 0.84},
		{time.Duration(321 * time.Second), time.Duration(330 * time.Second), 2.69,
			time.Duration(270 * time.Second), 0.83},
		{time.Duration(330 * time.Second), time.Duration(330 * time.Second), 2.69,
			time.Duration(270 * time.Second), 0.83},
		{time.Duration(331 * time.Second), time.Duration(331 * time.Second), 2.695,
			time.Duration(269 * time.Second), 0.825},
		{time.Duration(359 * time.Second), time.Duration(359 * time.Second), 2.835,
			time.Duration(241 * time.Second), 0.685},
		{time.Duration(360 * time.Second), time.Duration(360 * time.Second), 2.84,
			time.Duration(240 * time.Second), 0.68},
		{time.Duration(376 * time.Second), time.Duration(380 * time.Second), 2.85,
			time.Duration(220 * time.Second), 0.67},
		{time.Duration(391 * time.Second), time.Duration(391 * time.Second), 2.865,
			time.Duration(209 * time.Second), 0.655},
		{time.Duration(479 * time.Second), time.Duration(479 * time.Second), 3.175,
			time.Duration(121 * time.Second), 0.345},
		{time.Duration(599 * time.Second), time.Duration(599 * time.Second), 3.515,
			time.Duration(1 * time.Second), 0.005},
	}
	for _, tC := range testCases {
		t.Run(fmt.Sprintf("AtUsage:%s", tC.atUsage), func(t *testing.T) {
			var ec, srplsEC *EventCost
			ec = testEC.Clone()
			if srplsEC, err = ec.Trim(tC.atUsage); err != nil {
				t.Fatal(err)
			}
			if ec.GetUsage() != tC.ecUsage {
				t.Errorf("Wrongly trimmed EC: %s", utils.ToIJSON(ec))
			} else if ec.GetCost() != tC.ecCost {
				t.Errorf("Wrong cost for event: %s", utils.ToIJSON(ec))
			}
			if srplsEC.GetUsage() != tC.srplsUsage {
				t.Errorf("Wrong usage: %v for surplusEC: %s", srplsEC.GetUsage(), utils.ToIJSON(srplsEC))
			} else if srplsEC.GetCost() != tC.srplsCost {
				t.Errorf("Wrong cost: %f in surplus: %s", srplsEC.GetCost(), utils.ToIJSON(srplsEC))
			}
		})
	}
}

func TestECMergeGT(t *testing.T) {
	// InitialEventCost
	ecGT := &EventCost{
		CGRID:     "7636f3f1a06dffa038ba7900fb57f52d28830a24",
		RunID:     utils.META_DEFAULT,
		StartTime: time.Date(2018, 7, 27, 0, 59, 21, 0, time.UTC),
		Charges: []*ChargingInterval{
			&ChargingInterval{
				RatingID: "cc68da4",
				Increments: []*ChargingIncrement{
					&ChargingIncrement{
						Usage:          time.Duration(102400),
						AccountingID:   "0d87a64",
						CompressFactor: 103,
					},
				},
				CompressFactor: 1,
			},
		},
		AccountSummary: &AccountSummary{
			Tenant: "cgrates.org",
			ID:     "dan",
			BalanceSummaries: []*BalanceSummary{
				&BalanceSummary{
					UUID:  "9a767726-fe69-4940-b7bd-f43de9f0f8a5",
					ID:    "addon_data",
					Type:  utils.DATA,
					Value: 10726871040},
			},
		},
		Rating: Rating{
			"cc68da4": &RatingUnit{
				RatesID:         "06dee2e",
				RatingFiltersID: "216b0a5",
			},
		},
		Accounting: Accounting{
			"0d87a64": &BalanceCharge{
				AccountID:     "cgrates.org:dan",
				BalanceUUID:   "9a767726-fe69-4940-b7bd-f43de9f0f8a5",
				Units:         102400,
				ExtraChargeID: utils.META_NONE,
			},
		},
		RatingFilters: RatingFilters{
			"216b0a5": RatingMatchedFilters{
				"DestinationID":     utils.META_ANY,
				"DestinationPrefix": "42502",
				"RatingPlanID":      utils.META_NONE,
				"Subject":           "9a767726-fe69-4940-b7bd-f43de9f0f8a5",
			},
		},
		Rates: ChargedRates{
			"06dee2e": RateGroups{
				&Rate{
					RateIncrement: time.Duration(102400),
					RateUnit:      time.Duration(102400)},
			},
		},
	}
	ecGTUpdt := &EventCost{
		CGRID:     "7636f3f1a06dffa038ba7900fb57f52d28830a24",
		RunID:     utils.META_DEFAULT,
		StartTime: time.Date(2018, 7, 27, 0, 59, 38, 0105472, time.UTC),
		Charges: []*ChargingInterval{
			&ChargingInterval{
				RatingID: "6a83227",
				Increments: []*ChargingIncrement{
					&ChargingIncrement{
						Usage:          time.Duration(102400),
						AccountingID:   "9288f93",
						CompressFactor: 84,
					},
				},
				CompressFactor: 1,
			},
		},
		AccountSummary: &AccountSummary{
			Tenant: "cgrates.org",
			ID:     "dan",
			BalanceSummaries: []*BalanceSummary{
				&BalanceSummary{
					UUID:  "9a767726-fe69-4940-b7bd-f43de9f0f8a5",
					ID:    "addon_data",
					Type:  utils.DATA,
					Value: 10718269440},
			},
		},
		Rating: Rating{
			"6a83227": &RatingUnit{
				RatesID:         "52f8b0f",
				RatingFiltersID: "17f7216",
			},
		},
		Accounting: Accounting{
			"9288f93": &BalanceCharge{
				AccountID:     "cgrates.org:dan",
				BalanceUUID:   "9a767726-fe69-4940-b7bd-f43de9f0f8a5",
				Units:         102400,
				ExtraChargeID: utils.META_NONE,
			},
		},
		RatingFilters: RatingFilters{
			"17f7216": RatingMatchedFilters{
				"DestinationID":     utils.META_ANY,
				"DestinationPrefix": "42502",
				"RatingPlanID":      utils.META_NONE,
				"Subject":           "9a767726-fe69-4940-b7bd-f43de9f0f8a5",
			},
		},
		Rates: ChargedRates{
			"52f8b0f": RateGroups{
				&Rate{
					RateIncrement: time.Duration(102400),
					RateUnit:      time.Duration(102400)},
			},
		},
	}
	ecGT.Merge(ecGTUpdt)
	ecExpct := &EventCost{
		CGRID:     "7636f3f1a06dffa038ba7900fb57f52d28830a24",
		RunID:     utils.META_DEFAULT,
		StartTime: time.Date(2018, 7, 27, 0, 59, 21, 0, time.UTC),
		Charges: []*ChargingInterval{
			&ChargingInterval{
				RatingID: "cc68da4",
				Increments: []*ChargingIncrement{
					&ChargingIncrement{
						Usage:          time.Duration(102400),
						AccountingID:   "0d87a64",
						CompressFactor: 187,
					},
				},
				CompressFactor: 1,
			},
		},
		AccountSummary: &AccountSummary{
			Tenant: "cgrates.org",
			ID:     "dan",
			BalanceSummaries: []*BalanceSummary{
				&BalanceSummary{
					UUID:  "9a767726-fe69-4940-b7bd-f43de9f0f8a5",
					ID:    "addon_data",
					Type:  utils.DATA,
					Value: 10718269440},
			},
		},
		Rating: Rating{
			"cc68da4": &RatingUnit{
				RatesID:         "06dee2e",
				RatingFiltersID: "216b0a5",
			},
		},
		Accounting: Accounting{
			"0d87a64": &BalanceCharge{
				AccountID:     "cgrates.org:dan",
				BalanceUUID:   "9a767726-fe69-4940-b7bd-f43de9f0f8a5",
				Units:         102400,
				ExtraChargeID: utils.META_NONE,
			},
		},
		RatingFilters: RatingFilters{
			"216b0a5": RatingMatchedFilters{
				"DestinationID":     utils.META_ANY,
				"DestinationPrefix": "42502",
				"RatingPlanID":      utils.META_NONE,
				"Subject":           "9a767726-fe69-4940-b7bd-f43de9f0f8a5",
			},
		},
		Rates: ChargedRates{
			"06dee2e": RateGroups{
				&Rate{
					RateIncrement: time.Duration(102400),
					RateUnit:      time.Duration(102400)},
			},
		},
	}
	if len(ecGT.Charges) != len(ecExpct.Charges) ||
		!reflect.DeepEqual(ecGT.Charges[0].TotalUsage(), ecExpct.Charges[0].TotalUsage()) ||
		!reflect.DeepEqual(ecGT.Charges[0].TotalCost(), ecExpct.Charges[0].TotalCost()) {
		t.Errorf("expecting: %s\n\n, received: %s",
			utils.ToJSON(ecExpct), utils.ToJSON(ecGT))
	}

}

func TestECAppendCIlFromEC(t *testing.T) {
	// Standard compressing 1-1
	ec := &EventCost{
		Charges: []*ChargingInterval{
			&ChargingInterval{
				RatingID: "cc68da4",
				Increments: []*ChargingIncrement{
					&ChargingIncrement{
						Usage:          time.Duration(102400),
						AccountingID:   "0d87a64",
						CompressFactor: 103,
					},
				},
				CompressFactor: 1,
			},
		},
		Rating: Rating{
			"cc68da4": &RatingUnit{
				RatesID:         "06dee2e",
				RatingFiltersID: "216b0a5",
			},
		},
		Accounting: Accounting{
			"0d87a64": &BalanceCharge{
				AccountID:     "cgrates.org:dan",
				BalanceUUID:   "9a767726-fe69-4940-b7bd-f43de9f0f8a5",
				Units:         102400,
				ExtraChargeID: utils.META_NONE,
			},
		},
		RatingFilters: RatingFilters{
			"216b0a5": RatingMatchedFilters{
				"DestinationID":     utils.META_ANY,
				"DestinationPrefix": "42502",
				"RatingPlanID":      utils.META_NONE,
				"Subject":           "9a767726-fe69-4940-b7bd-f43de9f0f8a5",
			},
		},
		Rates: ChargedRates{
			"06dee2e": RateGroups{
				&Rate{
					RateIncrement: time.Duration(102400),
					RateUnit:      time.Duration(102400)},
			},
		},
	}
	oEC := &EventCost{
		Charges: []*ChargingInterval{
			&ChargingInterval{
				RatingID: "6a83227",
				Increments: []*ChargingIncrement{
					&ChargingIncrement{
						Usage:          time.Duration(102400),
						AccountingID:   "9288f93",
						CompressFactor: 84,
					},
				},
				CompressFactor: 1,
			},
		},
		Rating: Rating{
			"6a83227": &RatingUnit{
				RatesID:         "52f8b0f",
				RatingFiltersID: "17f7216",
			},
		},
		Accounting: Accounting{
			"9288f93": &BalanceCharge{
				AccountID:     "cgrates.org:dan",
				BalanceUUID:   "9a767726-fe69-4940-b7bd-f43de9f0f8a5",
				Units:         102400,
				ExtraChargeID: utils.META_NONE,
			},
		},
		RatingFilters: RatingFilters{
			"17f7216": RatingMatchedFilters{
				"DestinationID":     utils.META_ANY,
				"DestinationPrefix": "42502",
				"RatingPlanID":      utils.META_NONE,
				"Subject":           "9a767726-fe69-4940-b7bd-f43de9f0f8a5",
			},
		},
		Rates: ChargedRates{
			"52f8b0f": RateGroups{
				&Rate{
					RateIncrement: time.Duration(102400),
					RateUnit:      time.Duration(102400)},
			},
		},
	}
	ec.appendCIlFromEC(oEC, 0)
	eEC := &EventCost{
		Charges: []*ChargingInterval{
			&ChargingInterval{
				RatingID: "cc68da4",
				Increments: []*ChargingIncrement{
					&ChargingIncrement{
						Usage:          time.Duration(102400),
						AccountingID:   "0d87a64",
						CompressFactor: 187,
					},
				},
				CompressFactor: 1,
			},
		},
		Rating: Rating{
			"cc68da4": &RatingUnit{
				RatesID:         "06dee2e",
				RatingFiltersID: "216b0a5",
			},
		},
		Accounting: Accounting{
			"0d87a64": &BalanceCharge{
				AccountID:     "cgrates.org:dan",
				BalanceUUID:   "9a767726-fe69-4940-b7bd-f43de9f0f8a5",
				Units:         102400,
				ExtraChargeID: utils.META_NONE,
			},
		},
		RatingFilters: RatingFilters{
			"216b0a5": RatingMatchedFilters{
				"DestinationID":     utils.META_ANY,
				"DestinationPrefix": "42502",
				"RatingPlanID":      utils.META_NONE,
				"Subject":           "9a767726-fe69-4940-b7bd-f43de9f0f8a5",
			},
		},
		Rates: ChargedRates{
			"06dee2e": RateGroups{
				&Rate{
					RateIncrement: time.Duration(102400),
					RateUnit:      time.Duration(102400)},
			},
		},
	}
	if !reflect.DeepEqual(eEC, ec) {
		t.Errorf("expecting: %s, received: %s", utils.ToJSON(eEC), utils.ToJSON(ec))
	}

	// Second case, do not compress if first interval's compress factor is different than 1
	ec = &EventCost{
		Charges: []*ChargingInterval{
			&ChargingInterval{
				RatingID: "cc68da4",
				Increments: []*ChargingIncrement{
					&ChargingIncrement{
						Usage:          time.Duration(102400),
						AccountingID:   "0d87a64",
						CompressFactor: 103,
					},
				},
				CompressFactor: 2,
			},
		},
		Rating: Rating{
			"cc68da4": &RatingUnit{
				RatesID:         "06dee2e",
				RatingFiltersID: "216b0a5",
			},
		},
		Accounting: Accounting{
			"0d87a64": &BalanceCharge{
				AccountID:     "cgrates.org:dan",
				BalanceUUID:   "9a767726-fe69-4940-b7bd-f43de9f0f8a5",
				Units:         102400,
				ExtraChargeID: utils.META_NONE,
			},
		},
		RatingFilters: RatingFilters{
			"216b0a5": RatingMatchedFilters{
				"DestinationID":     utils.META_ANY,
				"DestinationPrefix": "42502",
				"RatingPlanID":      utils.META_NONE,
				"Subject":           "9a767726-fe69-4940-b7bd-f43de9f0f8a5",
			},
		},
		Rates: ChargedRates{
			"06dee2e": RateGroups{
				&Rate{
					RateIncrement: time.Duration(102400),
					RateUnit:      time.Duration(102400)},
			},
		},
	}
	oEC = &EventCost{
		Charges: []*ChargingInterval{
			&ChargingInterval{
				RatingID: "6a83227",
				Increments: []*ChargingIncrement{
					&ChargingIncrement{
						Usage:          time.Duration(102400),
						AccountingID:   "9288f93",
						CompressFactor: 84,
					},
				},
				CompressFactor: 1,
			},
		},
		Rating: Rating{
			"6a83227": &RatingUnit{
				RatesID:         "52f8b0f",
				RatingFiltersID: "17f7216",
			},
		},
		Accounting: Accounting{
			"9288f93": &BalanceCharge{
				AccountID:     "cgrates.org:dan",
				BalanceUUID:   "9a767726-fe69-4940-b7bd-f43de9f0f8a5",
				Units:         102400,
				ExtraChargeID: utils.META_NONE,
			},
		},
		RatingFilters: RatingFilters{
			"17f7216": RatingMatchedFilters{
				"DestinationID":     utils.META_ANY,
				"DestinationPrefix": "42502",
				"RatingPlanID":      utils.META_NONE,
				"Subject":           "9a767726-fe69-4940-b7bd-f43de9f0f8a5",
			},
		},
		Rates: ChargedRates{
			"52f8b0f": RateGroups{
				&Rate{
					RateIncrement: time.Duration(102400),
					RateUnit:      time.Duration(102400)},
			},
		},
	}
	ec.appendCIlFromEC(oEC, 0)
	eEC = &EventCost{
		Charges: []*ChargingInterval{
			&ChargingInterval{
				RatingID: "cc68da4",
				Increments: []*ChargingIncrement{
					&ChargingIncrement{
						Usage:          time.Duration(102400),
						AccountingID:   "0d87a64",
						CompressFactor: 103,
					},
				},
				CompressFactor: 2,
			},
			&ChargingInterval{
				RatingID: "cc68da4",
				Increments: []*ChargingIncrement{
					&ChargingIncrement{
						Usage:          time.Duration(102400),
						AccountingID:   "0d87a64",
						CompressFactor: 84,
					},
				},
				CompressFactor: 1,
			},
		},
		Rating: Rating{
			"cc68da4": &RatingUnit{
				RatesID:         "06dee2e",
				RatingFiltersID: "216b0a5",
			},
		},
		Accounting: Accounting{
			"0d87a64": &BalanceCharge{
				AccountID:     "cgrates.org:dan",
				BalanceUUID:   "9a767726-fe69-4940-b7bd-f43de9f0f8a5",
				Units:         102400,
				ExtraChargeID: utils.META_NONE,
			},
		},
		RatingFilters: RatingFilters{
			"216b0a5": RatingMatchedFilters{
				"DestinationID":     utils.META_ANY,
				"DestinationPrefix": "42502",
				"RatingPlanID":      utils.META_NONE,
				"Subject":           "9a767726-fe69-4940-b7bd-f43de9f0f8a5",
			},
		},
		Rates: ChargedRates{
			"06dee2e": RateGroups{
				&Rate{
					RateIncrement: time.Duration(102400),
					RateUnit:      time.Duration(102400)},
			},
		},
	}
	if !reflect.DeepEqual(eEC, ec) {
		t.Errorf("expecting: %s, received: %s", utils.ToJSON(eEC), utils.ToJSON(ec))
	}

	// Third case, split oEC
	ec = &EventCost{
		Charges: []*ChargingInterval{
			&ChargingInterval{
				RatingID: "cc68da4",
				Increments: []*ChargingIncrement{
					&ChargingIncrement{
						Usage:          time.Duration(100),
						AccountingID:   "0d87a64",
						CompressFactor: 1,
					},
				},
				CompressFactor: 1,
			},
			&ChargingInterval{
				RatingID: "cc68da4",
				Increments: []*ChargingIncrement{
					&ChargingIncrement{
						Usage:          time.Duration(102400),
						AccountingID:   "0d87a64",
						CompressFactor: 103,
					},
				},
				CompressFactor: 1,
			},
		},
		Rating: Rating{
			"cc68da4": &RatingUnit{
				RatesID:         "06dee2e",
				RatingFiltersID: "216b0a5",
			},
		},
		Accounting: Accounting{
			"0d87a64": &BalanceCharge{
				AccountID:     "cgrates.org:dan",
				BalanceUUID:   "9a767726-fe69-4940-b7bd-f43de9f0f8a5",
				Units:         102400,
				ExtraChargeID: utils.META_NONE,
			},
		},
		RatingFilters: RatingFilters{
			"216b0a5": RatingMatchedFilters{
				"DestinationID":     utils.META_ANY,
				"DestinationPrefix": "42502",
				"RatingPlanID":      utils.META_NONE,
				"Subject":           "9a767726-fe69-4940-b7bd-f43de9f0f8a5",
			},
		},
		Rates: ChargedRates{
			"06dee2e": RateGroups{
				&Rate{
					RateIncrement: time.Duration(102400),
					RateUnit:      time.Duration(102400)},
			},
		},
	}
	oEC = &EventCost{
		Charges: []*ChargingInterval{
			&ChargingInterval{
				RatingID: "6a83227",
				Increments: []*ChargingIncrement{
					&ChargingIncrement{
						Usage:          time.Duration(102400),
						AccountingID:   "9288f93",
						CompressFactor: 42,
					},
					&ChargingIncrement{
						Usage:          time.Duration(10240),
						AccountingID:   "9288f93",
						CompressFactor: 20,
					},
				},
				CompressFactor: 3,
			},
		},
		Rating: Rating{
			"6a83227": &RatingUnit{
				RatesID:         "52f8b0f",
				RatingFiltersID: "17f7216",
			},
		},
		Accounting: Accounting{
			"9288f93": &BalanceCharge{
				AccountID:     "cgrates.org:dan",
				BalanceUUID:   "9a767726-fe69-4940-b7bd-f43de9f0f8a5",
				Units:         102400,
				ExtraChargeID: utils.META_NONE,
			},
		},
		RatingFilters: RatingFilters{
			"17f7216": RatingMatchedFilters{
				"DestinationID":     utils.META_ANY,
				"DestinationPrefix": "42502",
				"RatingPlanID":      utils.META_NONE,
				"Subject":           "9a767726-fe69-4940-b7bd-f43de9f0f8a5",
			},
		},
		Rates: ChargedRates{
			"52f8b0f": RateGroups{
				&Rate{
					RateIncrement: time.Duration(102400),
					RateUnit:      time.Duration(102400)},
			},
		},
	}
	ec.appendCIlFromEC(oEC, 0)
	eEC = &EventCost{
		Charges: []*ChargingInterval{
			&ChargingInterval{
				RatingID: "cc68da4",
				Increments: []*ChargingIncrement{
					&ChargingIncrement{
						Usage:          time.Duration(100),
						AccountingID:   "0d87a64",
						CompressFactor: 1,
					},
				},
				CompressFactor: 1,
			},
			&ChargingInterval{
				RatingID: "cc68da4",
				Increments: []*ChargingIncrement{
					&ChargingIncrement{
						Usage:          time.Duration(102400),
						AccountingID:   "0d87a64",
						CompressFactor: 145,
					},
				},
				CompressFactor: 1,
			},
			&ChargingInterval{
				RatingID: "cc68da4",
				Increments: []*ChargingIncrement{
					&ChargingIncrement{
						Usage:          time.Duration(10240),
						AccountingID:   "0d87a64",
						CompressFactor: 20,
					},
				},
				CompressFactor: 1,
			},
			&ChargingInterval{
				RatingID: "cc68da4",
				Increments: []*ChargingIncrement{
					&ChargingIncrement{
						Usage:          time.Duration(102400),
						AccountingID:   "0d87a64",
						CompressFactor: 42,
					},
					&ChargingIncrement{
						Usage:          time.Duration(10240),
						AccountingID:   "0d87a64",
						CompressFactor: 20,
					},
				},
				CompressFactor: 2,
			},
		},
		Rating: Rating{
			"cc68da4": &RatingUnit{
				RatesID:         "06dee2e",
				RatingFiltersID: "216b0a5",
			},
		},
		Accounting: Accounting{
			"0d87a64": &BalanceCharge{
				AccountID:     "cgrates.org:dan",
				BalanceUUID:   "9a767726-fe69-4940-b7bd-f43de9f0f8a5",
				Units:         102400,
				ExtraChargeID: utils.META_NONE,
			},
		},
		RatingFilters: RatingFilters{
			"216b0a5": RatingMatchedFilters{
				"DestinationID":     utils.META_ANY,
				"DestinationPrefix": "42502",
				"RatingPlanID":      utils.META_NONE,
				"Subject":           "9a767726-fe69-4940-b7bd-f43de9f0f8a5",
			},
		},
		Rates: ChargedRates{
			"06dee2e": RateGroups{
				&Rate{
					RateIncrement: time.Duration(102400),
					RateUnit:      time.Duration(102400)},
			},
		},
	}
	if !reflect.DeepEqual(eEC, ec) {
		t.Errorf("expecting: %s, received: %s", utils.ToJSON(eEC), utils.ToJSON(ec))
	}

	// Fourth case, increase ChargingInterval.CompressFactor
	ec = &EventCost{
		Charges: []*ChargingInterval{
			&ChargingInterval{
				RatingID: "cc68da4",
				Increments: []*ChargingIncrement{
					&ChargingIncrement{
						Usage:          time.Duration(102400),
						AccountingID:   "0d87a64",
						CompressFactor: 103,
					},
				},
				CompressFactor: 2,
			},
		},
		Rating: Rating{
			"cc68da4": &RatingUnit{
				RatesID:         "06dee2e",
				RatingFiltersID: "216b0a5",
			},
		},
		Accounting: Accounting{
			"0d87a64": &BalanceCharge{
				AccountID:     "cgrates.org:dan",
				BalanceUUID:   "9a767726-fe69-4940-b7bd-f43de9f0f8a5",
				Units:         102400,
				ExtraChargeID: utils.META_NONE,
			},
		},
		RatingFilters: RatingFilters{
			"216b0a5": RatingMatchedFilters{
				"DestinationID":     utils.META_ANY,
				"DestinationPrefix": "42502",
				"RatingPlanID":      utils.META_NONE,
				"Subject":           "9a767726-fe69-4940-b7bd-f43de9f0f8a5",
			},
		},
		Rates: ChargedRates{
			"06dee2e": RateGroups{
				&Rate{
					RateIncrement: time.Duration(102400),
					RateUnit:      time.Duration(102400)},
			},
		},
	}
	oEC = &EventCost{
		Charges: []*ChargingInterval{
			&ChargingInterval{
				RatingID: "6a83227",
				Increments: []*ChargingIncrement{
					&ChargingIncrement{
						Usage:          time.Duration(102400),
						AccountingID:   "9288f93",
						CompressFactor: 103,
					},
				},
				CompressFactor: 3,
			},
		},
		Rating: Rating{
			"6a83227": &RatingUnit{
				RatesID:         "52f8b0f",
				RatingFiltersID: "17f7216",
			},
		},
		Accounting: Accounting{
			"9288f93": &BalanceCharge{
				AccountID:     "cgrates.org:dan",
				BalanceUUID:   "9a767726-fe69-4940-b7bd-f43de9f0f8a5",
				Units:         102400,
				ExtraChargeID: utils.META_NONE,
			},
		},
		RatingFilters: RatingFilters{
			"17f7216": RatingMatchedFilters{
				"DestinationID":     utils.META_ANY,
				"DestinationPrefix": "42502",
				"RatingPlanID":      utils.META_NONE,
				"Subject":           "9a767726-fe69-4940-b7bd-f43de9f0f8a5",
			},
		},
		Rates: ChargedRates{
			"52f8b0f": RateGroups{
				&Rate{
					RateIncrement: time.Duration(102400),
					RateUnit:      time.Duration(102400)},
			},
		},
	}
	ec.appendCIlFromEC(oEC, 0)
	eEC = &EventCost{
		Charges: []*ChargingInterval{
			&ChargingInterval{
				RatingID: "cc68da4",
				Increments: []*ChargingIncrement{
					&ChargingIncrement{
						Usage:          time.Duration(102400),
						AccountingID:   "0d87a64",
						CompressFactor: 103,
					},
				},
				CompressFactor: 5,
			},
		},
		Rating: Rating{
			"cc68da4": &RatingUnit{
				RatesID:         "06dee2e",
				RatingFiltersID: "216b0a5",
			},
		},
		Accounting: Accounting{
			"0d87a64": &BalanceCharge{
				AccountID:     "cgrates.org:dan",
				BalanceUUID:   "9a767726-fe69-4940-b7bd-f43de9f0f8a5",
				Units:         102400,
				ExtraChargeID: utils.META_NONE,
			},
		},
		RatingFilters: RatingFilters{
			"216b0a5": RatingMatchedFilters{
				"DestinationID":     utils.META_ANY,
				"DestinationPrefix": "42502",
				"RatingPlanID":      utils.META_NONE,
				"Subject":           "9a767726-fe69-4940-b7bd-f43de9f0f8a5",
			},
		},
		Rates: ChargedRates{
			"06dee2e": RateGroups{
				&Rate{
					RateIncrement: time.Duration(102400),
					RateUnit:      time.Duration(102400)},
			},
		},
	}
	if !reflect.DeepEqual(eEC, ec) {
		t.Errorf("expecting: %s, received: %s", utils.ToJSON(eEC), utils.ToJSON(ec))
	}
}
