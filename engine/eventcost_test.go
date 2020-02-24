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

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

// testEC is used as sample through various tests
var testEC = &EventCost{
	CGRID:     "164b0422fdc6a5117031b427439482c6a4f90e41",
	RunID:     utils.MetaDefault,
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
				UUID:  "8c54a9e9-d610-4c82-bcb5-a315b9a65010",
				Type:  utils.MONETARY,
				Value: 50,
			},
			&BalanceSummary{
				UUID:  "7a54a9e9-d610-4c82-bcb5-a315b9a65010",
				Type:  utils.MONETARY,
				Value: 25,
			},
			&BalanceSummary{
				UUID:  "4b8b53d7-c1a1-4159-b845-4623a00a0165",
				Type:  utils.VOICE,
				Value: 200,
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
			BalanceUUID:   "4b8b53d7-c1a1-4159-b845-4623a00a0165",
			RatingID:      "3cd6425",
			Units:         1,
			ExtraChargeID: "188bfa6",
		},
		"3455b83": &BalanceCharge{
			AccountID:     "cgrates.org:dan",
			BalanceUUID:   "4b8b53d7-c1a1-4159-b845-4623a00a0165",
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
	cache: utils.MapStorage{},
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
		Category:    "call",
		Tenant:      "cgrates.org",
		Subject:     "dan",
		Account:     "dan",
		Destination: "+4986517174963",
		ToR:         utils.VOICE,
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
								ID:    utils.MetaDefault,
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
								ToR:      utils.VOICE,
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
								ToR:      utils.VOICE,
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
								ID:    utils.MetaDefault,
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
								ID:    utils.MetaDefault,
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
		RunID:     utils.MetaDefault,
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
	ec := NewEventCostFromCallCost(cc, "164b0422fdc6a5117031b427439482c6a4f90e41", utils.MetaDefault)
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

func TestECAsRefundIncrements(t *testing.T) {
	eCD := &CallDescriptor{
		CgrID:         "164b0422fdc6a5117031b427439482c6a4f90e41",
		RunID:         utils.MetaDefault,
		ToR:           utils.VOICE,
		TimeStart:     time.Date(2017, 1, 9, 16, 18, 21, 0, time.UTC),
		TimeEnd:       time.Date(2017, 1, 9, 16, 28, 21, 0, time.UTC),
		DurationIndex: time.Duration(10 * time.Minute),
	}
	eCD.Increments = Increments{
		&Increment{
			Duration:       time.Duration(0),
			Cost:           0.1,
			CompressFactor: 1,
			BalanceInfo: &DebitInfo{
				AccountID: "cgrates.org:dan",
				Monetary: &MonetaryInfo{
					UUID: "8c54a9e9-d610-4c82-bcb5-a315b9a65010"}},
		},
		&Increment{
			Duration:       time.Duration(1 * time.Second),
			Cost:           0,
			CompressFactor: 10,
			BalanceInfo: &DebitInfo{
				AccountID: "cgrates.org:dan",
				Unit: &UnitInfo{
					UUID: "4b8b53d7-c1a1-4159-b845-4623a00a0165"}},
		},
		&Increment{
			Duration:       time.Duration(10 * time.Second),
			Cost:           0.01,
			CompressFactor: 2,
			BalanceInfo: &DebitInfo{
				AccountID: "cgrates.org:dan",
				Monetary: &MonetaryInfo{
					UUID: "8c54a9e9-d610-4c82-bcb5-a315b9a65010"}},
		},
		&Increment{
			Duration:       time.Duration(1 * time.Second),
			Cost:           0.005,
			CompressFactor: 30,
			BalanceInfo: &DebitInfo{
				AccountID: "cgrates.org:dan",
				Unit: &UnitInfo{
					UUID: "4b8b53d7-c1a1-4159-b845-4623a00a0165"},
				Monetary: &MonetaryInfo{
					UUID: "8c54a9e9-d610-4c82-bcb5-a315b9a65010"}},
		},
		&Increment{
			Duration:       time.Duration(1 * time.Second),
			Cost:           0.01,
			CompressFactor: 60,
			BalanceInfo: &DebitInfo{
				AccountID: "cgrates.org:dan",
				Monetary: &MonetaryInfo{
					UUID: "8c54a9e9-d610-4c82-bcb5-a315b9a65010"}},
		},
		&Increment{
			Duration:       time.Duration(1 * time.Second),
			Cost:           0.01,
			CompressFactor: 60,
			BalanceInfo: &DebitInfo{
				AccountID: "cgrates.org:dan",
				Monetary: &MonetaryInfo{
					UUID: "8c54a9e9-d610-4c82-bcb5-a315b9a65010"}},
		},
		&Increment{
			Duration:       time.Duration(1 * time.Second),
			Cost:           0.01,
			CompressFactor: 60,
			BalanceInfo: &DebitInfo{
				AccountID: "cgrates.org:dan",
				Monetary: &MonetaryInfo{
					UUID: "8c54a9e9-d610-4c82-bcb5-a315b9a65010"}},
		},
		&Increment{
			Duration:       time.Duration(1 * time.Second),
			Cost:           0.01,
			CompressFactor: 60,
			BalanceInfo: &DebitInfo{
				AccountID: "cgrates.org:dan",
				Monetary: &MonetaryInfo{
					UUID: "8c54a9e9-d610-4c82-bcb5-a315b9a65010"}},
		},
		&Increment{
			Duration:       time.Duration(1 * time.Second),
			Cost:           0,
			CompressFactor: 10,
			BalanceInfo: &DebitInfo{
				AccountID: "cgrates.org:dan",
				Unit: &UnitInfo{
					UUID: "4b8b53d7-c1a1-4159-b845-4623a00a0165"}},
		},
		&Increment{
			Duration:       time.Duration(10 * time.Second),
			Cost:           0.01,
			CompressFactor: 2,
			BalanceInfo: &DebitInfo{
				AccountID: "cgrates.org:dan",
				Monetary: &MonetaryInfo{
					UUID: "8c54a9e9-d610-4c82-bcb5-a315b9a65010"}},
		},
		&Increment{
			Duration:       time.Duration(1 * time.Second),
			Cost:           0.005,
			CompressFactor: 30,
			BalanceInfo: &DebitInfo{
				AccountID: "cgrates.org:dan",
				Unit: &UnitInfo{
					UUID: "4b8b53d7-c1a1-4159-b845-4623a00a0165"},
				Monetary: &MonetaryInfo{
					UUID: "8c54a9e9-d610-4c82-bcb5-a315b9a65010"}},
		},
		&Increment{
			Duration:       time.Duration(1 * time.Second),
			Cost:           0,
			CompressFactor: 10,
			BalanceInfo: &DebitInfo{
				AccountID: "cgrates.org:dan",
				Unit: &UnitInfo{
					UUID: "4b8b53d7-c1a1-4159-b845-4623a00a0165"}},
		},
		&Increment{
			Duration:       time.Duration(10 * time.Second),
			Cost:           0.01,
			CompressFactor: 2,
			BalanceInfo: &DebitInfo{
				AccountID: "cgrates.org:dan",
				Monetary: &MonetaryInfo{
					UUID: "8c54a9e9-d610-4c82-bcb5-a315b9a65010"}},
		},
		&Increment{
			Duration:       time.Duration(1 * time.Second),
			Cost:           0.005,
			CompressFactor: 30,
			BalanceInfo: &DebitInfo{
				AccountID: "cgrates.org:dan",
				Unit: &UnitInfo{
					UUID: "4b8b53d7-c1a1-4159-b845-4623a00a0165"},
				Monetary: &MonetaryInfo{
					UUID: "8c54a9e9-d610-4c82-bcb5-a315b9a65010"}},
		},
		&Increment{
			Duration:       time.Duration(1 * time.Second),
			Cost:           0,
			CompressFactor: 10,
			BalanceInfo: &DebitInfo{
				AccountID: "cgrates.org:dan",
				Unit: &UnitInfo{
					UUID: "4b8b53d7-c1a1-4159-b845-4623a00a0165"}},
		},
		&Increment{
			Duration:       time.Duration(10 * time.Second),
			Cost:           0.01,
			CompressFactor: 2,
			BalanceInfo: &DebitInfo{
				AccountID: "cgrates.org:dan",
				Monetary: &MonetaryInfo{
					UUID: "8c54a9e9-d610-4c82-bcb5-a315b9a65010"}},
		},
		&Increment{
			Duration:       time.Duration(1 * time.Second),
			Cost:           0.005,
			CompressFactor: 30,
			BalanceInfo: &DebitInfo{
				AccountID: "cgrates.org:dan",
				Unit: &UnitInfo{
					UUID: "4b8b53d7-c1a1-4159-b845-4623a00a0165"},
				Monetary: &MonetaryInfo{
					UUID: "8c54a9e9-d610-4c82-bcb5-a315b9a65010"}},
		},
		&Increment{
			Duration:       time.Duration(1 * time.Second),
			Cost:           0,
			CompressFactor: 10,
			BalanceInfo: &DebitInfo{
				AccountID: "cgrates.org:dan",
				Unit: &UnitInfo{
					UUID: "4b8b53d7-c1a1-4159-b845-4623a00a0165"}},
		},
		&Increment{
			Duration:       time.Duration(10 * time.Second),
			Cost:           0.01,
			CompressFactor: 2,
			BalanceInfo: &DebitInfo{
				AccountID: "cgrates.org:dan",
				Monetary: &MonetaryInfo{
					UUID: "8c54a9e9-d610-4c82-bcb5-a315b9a65010"}},
		},
		&Increment{
			Duration:       time.Duration(1 * time.Second),
			Cost:           0.005,
			CompressFactor: 30,
			BalanceInfo: &DebitInfo{
				AccountID: "cgrates.org:dan",
				Unit: &UnitInfo{
					UUID: "4b8b53d7-c1a1-4159-b845-4623a00a0165"},
				Monetary: &MonetaryInfo{
					UUID: "8c54a9e9-d610-4c82-bcb5-a315b9a65010"}},
		},
		&Increment{
			Duration:       time.Duration(1 * time.Second),
			Cost:           0,
			CompressFactor: 10,
			BalanceInfo: &DebitInfo{
				AccountID: "cgrates.org:dan",
				Unit: &UnitInfo{
					UUID: "4b8b53d7-c1a1-4159-b845-4623a00a0165"}},
		},
		&Increment{
			Duration:       time.Duration(10 * time.Second),
			Cost:           0.01,
			CompressFactor: 2,
			BalanceInfo: &DebitInfo{
				AccountID: "cgrates.org:dan",
				Monetary: &MonetaryInfo{
					UUID: "8c54a9e9-d610-4c82-bcb5-a315b9a65010"}},
		},
		&Increment{
			Duration:       time.Duration(1 * time.Second),
			Cost:           0.005,
			CompressFactor: 30,
			BalanceInfo: &DebitInfo{
				AccountID: "cgrates.org:dan",
				Unit: &UnitInfo{
					UUID: "4b8b53d7-c1a1-4159-b845-4623a00a0165"},
				Monetary: &MonetaryInfo{
					UUID: "8c54a9e9-d610-4c82-bcb5-a315b9a65010"}},
		},
	}

	if cd := testEC.Clone().AsRefundIncrements(utils.VOICE); !reflect.DeepEqual(eCD, cd) {
		t.Errorf("expecting: %s\n\n, received: %s", utils.ToIJSON(eCD), utils.ToIJSON(cd))
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
		RunID:     utils.MetaDefault,
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
		ToR:            utils.VOICE,
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
	cc := ec.AsCallCost(utils.EmptyString)
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
	if srplsEC, err := ec.Trim(time.Duration(0)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eFullSrpls, srplsEC) {
		t.Errorf("\tExpecting: %s,\n\treceived: %s",
			utils.ToJSON(eFullSrpls), utils.ToJSON(srplsEC))
	}
	//verify the event cost
	newEc := NewBareEventCost()
	newEc.CGRID = eFullSrpls.CGRID
	newEc.RunID = eFullSrpls.RunID
	newEc.StartTime = eFullSrpls.StartTime
	newEc.AccountSummary = eFullSrpls.AccountSummary.Clone()
	if !reflect.DeepEqual(newEc, ec) {
		t.Errorf("\tExpecting: %s,\n\treceived: %s",
			utils.ToJSON(newEc), utils.ToJSON(ec))
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
		RunID:     utils.MetaDefault,
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
		RunID:     utils.MetaDefault,
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
		RunID:     utils.MetaDefault,
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

func TestECSyncKeys(t *testing.T) {
	ec := testEC.Clone()
	ec.Accounting["a012888"].RatingID = "c1a5ab9"

	refEC := &EventCost{
		Rating: Rating{
			"21a5ab9": &RatingUnit{
				ConnectFee:       0.1,
				RoundingMethod:   "*up",
				RoundingDecimals: 5,
				TimingID:         "2f324ab",
				RatesID:          "2c1a177",
				RatingFiltersID:  "23e77dc",
			},
		},
		Accounting: Accounting{
			"2012888": &BalanceCharge{
				AccountID:   "cgrates.org:dan",
				BalanceUUID: "8c54a9e9-d610-4c82-bcb5-a315b9a65010",
				Units:       0.01,
				RatingID:    "21a5ab9",
			},
			"288bfa6": &BalanceCharge{
				AccountID:   "cgrates.org:dan",
				BalanceUUID: "8c54a9e9-d610-4c82-bcb5-a315b9a65010",
				Units:       0.005,
			},
			"24d6c02": &BalanceCharge{
				AccountID:     "cgrates.org:dan",
				BalanceUUID:   "7a54a9e9-d610-4c82-bcb5-a315b9a65010",
				RatingID:      "3cd6425",
				Units:         1,
				ExtraChargeID: "288bfa6",
			},
		},
		RatingFilters: RatingFilters{
			"23e77dc": RatingMatchedFilters{
				"DestinationID":     "GERMANY",
				"DestinationPrefix": "+49",
				"RatingPlanID":      "RPL_RETAIL1",
				"Subject":           "*out:cgrates.org:call:*any",
			},
		},
		Timings: ChargedTimings{
			"2f324ab": &ChargedTiming{
				StartTime: "00:00:00",
			},
		},
		Rates: ChargedRates{
			"2c1a177": RateGroups{
				&Rate{
					GroupIntervalStart: time.Duration(0),
					Value:              0.01,
					RateIncrement:      time.Duration(1 * time.Minute),
					RateUnit:           time.Duration(1 * time.Second)},
			},
		},
	}

	eEC := &EventCost{
		CGRID:     "164b0422fdc6a5117031b427439482c6a4f90e41",
		RunID:     utils.MetaDefault,
		StartTime: time.Date(2017, 1, 9, 16, 18, 21, 0, time.UTC),
		Charges: []*ChargingInterval{
			&ChargingInterval{
				RatingID: "21a5ab9",
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
						AccountingID:   "2012888",
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
				RatingID: "21a5ab9",
				Increments: []*ChargingIncrement{
					&ChargingIncrement{
						Usage:          time.Duration(1 * time.Second),
						Cost:           0.01,
						AccountingID:   "2012888",
						CompressFactor: 60,
					},
				},
				CompressFactor: 4,
			},
			&ChargingInterval{
				RatingID: "21a5ab9",
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
						AccountingID:   "2012888",
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
					UUID:     "8c54a9e9-d610-4c82-bcb5-a315b9a65010",
					Type:     utils.MONETARY,
					Value:    50,
					Disabled: false},
				&BalanceSummary{
					UUID:     "7a54a9e9-d610-4c82-bcb5-a315b9a65010",
					Type:     utils.MONETARY,
					Value:    25,
					Disabled: false},
				&BalanceSummary{
					UUID:     "4b8b53d7-c1a1-4159-b845-4623a00a0165",
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
				TimingID:         "2f324ab",
				RatesID:          "4910ecf",
				RatingFiltersID:  "23e77dc",
			},
			"21a5ab9": &RatingUnit{
				ConnectFee:       0.1,
				RoundingMethod:   "*up",
				RoundingDecimals: 5,
				TimingID:         "2f324ab",
				RatesID:          "2c1a177",
				RatingFiltersID:  "23e77dc",
			},
		},
		Accounting: Accounting{
			"2012888": &BalanceCharge{
				AccountID:   "cgrates.org:dan",
				BalanceUUID: "8c54a9e9-d610-4c82-bcb5-a315b9a65010",
				Units:       0.01,
				RatingID:    "21a5ab9",
			},
			"288bfa6": &BalanceCharge{
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
				BalanceUUID:   "4b8b53d7-c1a1-4159-b845-4623a00a0165",
				RatingID:      "3cd6425",
				Units:         1,
				ExtraChargeID: "288bfa6",
			},
			"3455b83": &BalanceCharge{
				AccountID:     "cgrates.org:dan",
				BalanceUUID:   "4b8b53d7-c1a1-4159-b845-4623a00a0165",
				Units:         1,
				ExtraChargeID: "*none",
			},
		},
		RatingFilters: RatingFilters{
			"23e77dc": RatingMatchedFilters{
				"DestinationID":     "GERMANY",
				"DestinationPrefix": "+49",
				"RatingPlanID":      "RPL_RETAIL1",
				"Subject":           "*out:cgrates.org:call:*any",
			},
		},
		Rates: ChargedRates{
			"2c1a177": RateGroups{
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
			"2f324ab": &ChargedTiming{
				StartTime: "00:00:00",
			},
		},
		cache: utils.MapStorage{},
	}

	ec.SyncKeys(refEC)
	if !reflect.DeepEqual(eEC, ec) {
		t.Errorf("expecting: %s \nreceived: %s",
			utils.ToIJSON(eEC), utils.ToIJSON(ec))
	}
}

func TestECAsDataProvider(t *testing.T) {
	ecDP := config.NewObjectDP(testEC)
	if data, err := ecDP.FieldAsInterface([]string{"RunID"}); err != nil {
		t.Error(err)
	} else if data != utils.MetaDefault {
		t.Errorf("Expecting: <%s> \nreceived: <%s>", utils.MetaDefault, data)
	}
	if data, err := ecDP.FieldAsInterface([]string{"AccountSummary", "ID"}); err != nil {
		t.Error(err)
	} else if data != "dan" {
		t.Errorf("Expecting: <%s> \nreceived: <%s>", "data", data)
	}
	if data, err := ecDP.FieldAsInterface([]string{"AccountSummary", "BalanceSummaries[1]", "UUID"}); err != nil {
		t.Error(err)
	} else if data != "7a54a9e9-d610-4c82-bcb5-a315b9a65010" {
		t.Errorf("Expecting: <%s> \nreceived: <%s>", "4b8b53d7-c1a1-4159-b845-4623a00a0165", data)
	}
	if data, err := ecDP.FieldAsInterface([]string{"AccountSummary", "BalanceSummaries[2]", "Type"}); err != nil {
		t.Error(err)
	} else if data != "*voice" {
		t.Errorf("Expecting: <%s> \nreceived: <%s>", "*voice", data)
	}
}

func TestInitCache(t *testing.T) {
	eventCost := &EventCost{}
	eventCost.initCache()
	eOut := utils.MapStorage{}
	if !reflect.DeepEqual(eOut, eventCost.cache) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eOut), utils.ToJSON(eventCost.cache))
	}
}

func TestEventCostFieldAsInterface(t *testing.T) {
	eventCost := &EventCost{}
	eventCost.initCache()
	// empty check
	if rcv, err := eventCost.FieldAsInterface([]string{}); err != utils.ErrNotFound {
		t.Errorf("Expecting: %+v, received: %+v", utils.ErrNotFound, err)
	} else if rcv != nil {
		t.Errorf("Expecting: nil, received: %+v", rcv)
	}
	// item found in cache
	eventCost.cache = utils.MapStorage{"test": nil}
	if rcv, err := eventCost.FieldAsInterface([]string{"test"}); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expecting: nil, received: %+v", err)
	} else if rcv != nil {
		t.Errorf("Expecting: nil, received: %+v", rcv)
	}
	// data found in cache
	eventCost.cache = utils.MapStorage{"test": "test"}
	if rcv, err := eventCost.FieldAsInterface([]string{"test"}); err != nil {
		t.Errorf("Expecting: nil, received: %+v", err)
	} else if rcv != "test" {
		t.Errorf("Expecting: nil, received: %+v", rcv)
	}
}

func TestEventCostfieldAsInterface(t *testing.T) {
	eventCost := &EventCost{}
	// default case
	if rcv, err := eventCost.fieldAsInterface([]string{utils.EmptyString}); err == nil || err.Error() != "unsupported field prefix: <>" {
		t.Error(err)
	} else if rcv != nil {
		t.Errorf("Expecting: nil, received: %+v", rcv)
	}
	// case utils.Charges:
	if rcv, err := eventCost.fieldAsInterface([]string{utils.Charges, utils.Charges}); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expecting: %+v, received: %+v", utils.ErrNotFound, err)
	} else if rcv != nil {
		t.Errorf("Expecting: nil, received: %+v", rcv)
	}
	eventCost = &EventCost{
		Charges: []*ChargingInterval{
			&ChargingInterval{
				RatingID: "RatingID1",
			},
			&ChargingInterval{
				RatingID: "RatingID2",
			},
		},
	}
	expectedCharges := []*ChargingInterval{
		&ChargingInterval{
			RatingID: "RatingID1",
		},
		&ChargingInterval{
			RatingID: "RatingID2",
		},
	}
	if rcv, err := eventCost.fieldAsInterface([]string{utils.Charges}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedCharges, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", expectedCharges, rcv)
	}
	// case utils.CGRID:
	if rcv, err := eventCost.fieldAsInterface([]string{utils.CGRID, utils.CGRID}); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expecting: %+v, received: %+v", utils.ErrNotFound, err)
	} else if rcv != nil {
		t.Errorf("Expecting: nil, received: %+v", rcv)
	}
	eventCost = &EventCost{
		CGRID: "CGRID",
	}
	if rcv, err := eventCost.fieldAsInterface([]string{utils.CGRID}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual("CGRID", rcv) {
		t.Errorf("Expecting: \"CGRID\", received: %+v", rcv)
	}
	// case utils.RunID:
	if rcv, err := eventCost.fieldAsInterface([]string{utils.RunID, utils.RunID}); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expecting: %+v, received: %+v", utils.ErrNotFound, err)
	} else if rcv != nil {
		t.Errorf("Expecting: nil, received: %+v", rcv)
	}
	eventCost = &EventCost{
		RunID: "RunID",
	}
	if rcv, err := eventCost.fieldAsInterface([]string{utils.RunID}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual("RunID", rcv) {
		t.Errorf("Expecting: \"RunID\", received: %+v", rcv)
	}
	// case utils.StartTime:
	if rcv, err := eventCost.fieldAsInterface([]string{utils.StartTime, utils.StartTime}); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expecting: %+v, received: %+v", utils.ErrNotFound, err)
	} else if rcv != nil {
		t.Errorf("Expecting: nil, received: %+v", rcv)
	}
	eventCost = &EventCost{
		StartTime: time.Date(2020, time.April, 18, 23, 0, 4, 0, time.UTC),
	}
	expectedStartTime := time.Date(2020, time.April, 18, 23, 0, 4, 0, time.UTC)
	if rcv, err := eventCost.fieldAsInterface([]string{utils.StartTime}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedStartTime, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", expectedStartTime, rcv)
	}
	// case utils.Usage:
	if rcv, err := eventCost.fieldAsInterface([]string{utils.Usage, utils.Usage}); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expecting: %+v, received: %+v", utils.ErrNotFound, err)
	} else if rcv != nil {
		t.Errorf("Expecting: nil, received: %+v", rcv)
	}
	eventCost = &EventCost{
		Usage: utils.DurationPointer(time.Duration(5 * time.Minute)),
	}
	if rcv, err := eventCost.fieldAsInterface([]string{utils.Usage}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(utils.DurationPointer(5*time.Minute), rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.DurationPointer(5*time.Minute), rcv)
	}
	// case utils.Cost:
	if rcv, err := eventCost.fieldAsInterface([]string{utils.Cost, utils.Cost}); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expecting: %+v, received: %+v", utils.ErrNotFound, err)
	} else if rcv != nil {
		t.Errorf("Expecting: nil, received: %+v", rcv)
	}
	eventCost = &EventCost{
		Cost: utils.Float64Pointer(0.7),
	}
	if rcv, err := eventCost.fieldAsInterface([]string{utils.Cost}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(utils.Float64Pointer(0.7), rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.Float64Pointer(0.7), rcv)
	}
	// case utils.AccountSummary:
	eventCost = &EventCost{
		AccountSummary: &AccountSummary{
			ID: "IDtest",
		},
	}
	expectedAccountSummary := &AccountSummary{ID: "IDtest"}
	if rcv, err := eventCost.fieldAsInterface([]string{utils.AccountSummary}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedAccountSummary, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", expectedAccountSummary, rcv)
	}
	eventCost = &EventCost{
		AccountSummary: &AccountSummary{
			ID:     "IDtest1",
			Tenant: "Tenant",
		},
		CGRID: "test",
	}
	if rcv, err := eventCost.fieldAsInterface([]string{utils.AccountSummary, utils.Tenant}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual("Tenant", rcv) {
		t.Errorf("Expecting: Tenant, received: %+v", utils.ToJSON(rcv))
	}
	// case utils.Timings:
	eventCost = &EventCost{
		Timings: ChargedTimings{
			"test1": &ChargedTiming{
				StartTime: "StartTime",
			},
		},
	}
	eTimings := ChargedTimings{"test1": &ChargedTiming{StartTime: "StartTime"}}
	if rcv, err := eventCost.fieldAsInterface([]string{utils.Timings}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eTimings, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", eTimings, rcv)
	}
	eChargedTiming := &ChargedTiming{
		StartTime: "StartTime",
	}
	if rcv, err := eventCost.fieldAsInterface([]string{utils.Timings, "test1"}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eChargedTiming, rcv) {
		fmt.Printf("%T", rcv)
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eChargedTiming), utils.ToJSON(rcv))
	}
	// case utils.Rates:
	eventCost = &EventCost{
		Rates: ChargedRates{
			"test1": RateGroups{
				&Rate{Value: 0.7},
			},
		},
	}
	eChargedRates := ChargedRates{
		"test1": RateGroups{
			&Rate{Value: 0.7},
		},
	}
	if rcv, err := eventCost.fieldAsInterface([]string{utils.Rates}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eChargedRates, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eChargedRates), utils.ToJSON(rcv))
	}
	eRateGroup := RateGroups{
		&Rate{Value: 0.7},
	}
	if rcv, err := eventCost.fieldAsInterface([]string{utils.Rates, "test1"}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eRateGroup, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eRateGroup), utils.ToJSON(rcv))
	}
	// case utils.RatingFilters:
	eventCost = &EventCost{
		RatingFilters: RatingFilters{
			"test1": RatingMatchedFilters{
				AccountActionsCSVContent: "AccountActionsCSVContent",
			},
		},
	}
	eRatingFilters := RatingFilters{
		"test1": RatingMatchedFilters{
			AccountActionsCSVContent: "AccountActionsCSVContent",
		},
	}
	if rcv, err := eventCost.fieldAsInterface([]string{utils.RatingFilters}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eRatingFilters, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eRatingFilters), utils.ToJSON(rcv))
	}
	eRatingMatchedFilters := RatingMatchedFilters{
		AccountActionsCSVContent: "AccountActionsCSVContent",
	}
	if rcv, err := eventCost.fieldAsInterface([]string{utils.RatingFilters, "test1"}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eRatingMatchedFilters, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eRatingMatchedFilters), utils.ToJSON(rcv))
	}
	// case utils.Accounting:
	eventCost = &EventCost{
		Accounting: Accounting{
			"test1": &BalanceCharge{
				AccountID: "AccountID",
			},
		},
	}
	eAccounting := Accounting{
		"test1": &BalanceCharge{
			AccountID: "AccountID",
		},
	}
	if rcv, err := eventCost.fieldAsInterface([]string{utils.Accounting}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eAccounting, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eAccounting), utils.ToJSON(rcv))
	}
	eBalanceCharge := &BalanceCharge{
		AccountID: "AccountID",
	}
	if rcv, err := eventCost.fieldAsInterface([]string{utils.Accounting, "test1"}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eBalanceCharge, rcv) {
		t.Errorf("\nExpecting: %+v, \nreceived: %+v", utils.ToJSON(eBalanceCharge), utils.ToJSON(rcv))
	}
	// case utils.Rating:
	eventCost = &EventCost{
		Rating: Rating{
			"test1": &RatingUnit{
				ConnectFee: 0.7,
			},
		},
	}
	eRating := Rating{
		"test1": &RatingUnit{
			ConnectFee: 0.7,
		},
	}
	if rcv, err := eventCost.fieldAsInterface([]string{utils.Rating}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eRating, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eRating), utils.ToJSON(rcv))
	}
	eRateUnit := &RatingUnit{
		ConnectFee: 0.7,
	}
	if rcv, err := eventCost.fieldAsInterface([]string{utils.Rating, "test1"}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eRateUnit, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eRateUnit), utils.ToJSON(rcv))
	}
	//default case, utils.Charges
	eventCost = &EventCost{
		Charges: []*ChargingInterval{
			&ChargingInterval{
				RatingID: "RatingID",
			},
		},
	}
	eCharges := []*ChargingInterval{&ChargingInterval{RatingID: "RatingID"}}
	if rcv, err := eventCost.fieldAsInterface([]string{utils.Charges}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCharges, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eCharges), utils.ToJSON(rcv))
	}
	if rcv, err := eventCost.fieldAsInterface([]string{utils.Charges + "[0]"}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCharges[0], rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eCharges[0]), utils.ToJSON(rcv))
	}
	if rcv, err := eventCost.fieldAsInterface([]string{utils.Charges + "[1]"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if rcv != nil {
		t.Errorf("Expecting: nil, received: %+v", rcv)
	}
}

func TestEventCostgetChargesForPath(t *testing.T) {
	eventCost := &EventCost{}
	chargingInterval := &ChargingInterval{
		RatingID: "RatingID",
	}
	// chr == nil
	if rcv, err := eventCost.getChargesForPath(nil, nil); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expecting: %+v, received: %+v", utils.ErrNotFound, err)
	} else if rcv != nil {
		t.Errorf("Expecting: nil, received: %+v", rcv)
	}
	// len(fldPath) == 0
	eChargingInterval := &ChargingInterval{
		RatingID: "RatingID",
	}
	if rcv, err := eventCost.getChargesForPath([]string{}, chargingInterval); err != nil {
		t.Errorf("Expecting: nil, received: %+v", err)
	} else if !reflect.DeepEqual(eChargingInterval, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", eChargingInterval, rcv)
	}
	// fldPath[0] == utils.CompressFactor
	chargingInterval = &ChargingInterval{
		RatingID:       "RatingID",
		CompressFactor: 7,
	}
	if rcv, err := eventCost.getChargesForPath([]string{utils.CompressFactor}, chargingInterval); err != nil {
		t.Errorf("Expecting: nil, received: %+v", err)
	} else if rcv != 7 {
		t.Errorf("Expecting: 7, received: %+v", rcv)
	}
	if rcv, err := eventCost.getChargesForPath([]string{utils.CompressFactor, "mustFail"}, chargingInterval); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expecting: %+v, received: %+v", utils.ErrNotFound, err)
	} else if rcv != nil {
		t.Errorf("Expecting: nil, received: %+v", rcv)
	}
	// fldPath[0] == utils.Rating
	eventCost = &EventCost{
		Rates: ChargedRates{
			"RatingID": RateGroups{&Rate{Value: 0.8}}},
	}
	if rcv, err := eventCost.getChargesForPath([]string{utils.Rating}, chargingInterval); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expecting: %+v, received: %+v", utils.ErrNotFound, err)
	} else if rcv != nil {
		t.Errorf("Expecting: nil, received: %+v", rcv)
	}
	eventCost = &EventCost{
		Rating: Rating{"RatingID": &RatingUnit{}},
		Rates: ChargedRates{
			"RatingID": RateGroups{&Rate{Value: 0.8}}},
	}
	if rcv, err := eventCost.getChargesForPath([]string{utils.Rating, "unsupportedfield"}, chargingInterval); err == nil || err.Error() != "unsupported field prefix: <unsupportedfield>" {
		t.Errorf("Expecting: unsupported field prefix: <unsupportedfield>, received: %+v", err)
	} else if rcv != nil {
		t.Errorf("Expecting: nil, received: %+v", rcv)
	}

	chargingInterval = &ChargingInterval{
		RatingID:       "RatingID",
		CompressFactor: 7,
	}
	eventCost = &EventCost{
		Rating: Rating{"RatingID": &RatingUnit{
			RatesID: "RatesID",
		}},
		Rates: ChargedRates{"RatesID": RateGroups{&Rate{Value: 0.8}}},
	}
	RateGroups := RateGroups{&Rate{Value: 0.8}}
	if rcv, err := eventCost.getChargesForPath([]string{utils.Rating, utils.Rates}, chargingInterval); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(RateGroups, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(RateGroups), utils.ToJSON(rcv))
		t.Errorf("Expecting: %+v, received: %+v", RateGroups, rcv)
	}
	// opath != utils.Increments
	if rcv, err := eventCost.getChargesForPath([]string{"unsupportedfield"}, chargingInterval); err == nil || err.Error() != "unsupported field prefix: <unsupportedfield>" {
		t.Errorf("Expecting: unsupported field prefix: <unsupportedfield>, received: %+v", err)
	} else if rcv != nil {
		t.Errorf("Expecting: nil, received: %+v", rcv)
	}
	// utils.Increments
	eventCost = &EventCost{}
	chargingInterval = &ChargingInterval{
		Increments: []*ChargingIncrement{
			&ChargingIncrement{
				AccountingID: "AccountingID",
			},
		},
	}
	eChargingIncrement := &ChargingIncrement{
		AccountingID: "AccountingID",
	}
	if rcv, err := eventCost.getChargesForPath([]string{"Increments[0]"}, chargingInterval); err != nil {
		t.Errorf("Expecting: nil, received: %+v", err)
	} else if !reflect.DeepEqual(eChargingIncrement, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", eChargingIncrement, rcv)
	}
	eIncrements := []*ChargingIncrement{
		&ChargingIncrement{
			AccountingID: "AccountingID",
		},
	}
	// indx == nil
	if rcv, err := eventCost.getChargesForPath([]string{"Increments"}, chargingInterval); err != nil {
		t.Errorf("Expecting: nil, received: %+v", err)
	} else if !reflect.DeepEqual(eIncrements, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", eIncrements, utils.ToJSON(rcv))
	}
	// len(fldPath) != 1
	if rcv, err := eventCost.getChargesForPath([]string{"Increments", utils.Accounting}, chargingInterval); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expecting: %+v, received: %+v", utils.ErrNotFound, err)
	} else if rcv != nil {
		t.Errorf("Expecting: nil, received: %+v", rcv)
	}
	// fldPath[1] == utils.Accounting
	eventCost = &EventCost{
		Accounting: Accounting{
			"AccountingID": &BalanceCharge{
				AccountID: "AccountID",
			},
		},
	}
	eBalanceCharge := &BalanceCharge{AccountID: "AccountID"}
	if rcv, err := eventCost.getChargesForPath([]string{"Increments[0]", utils.Accounting}, chargingInterval); err != nil {
		t.Errorf("Expecting: nil, received: %+v", err)
	} else if !reflect.DeepEqual(eBalanceCharge, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", eBalanceCharge, rcv)
	}

}

func TestEventCostgetRatingForPath(t *testing.T) {
	eventCost := &EventCost{}
	ratingUnit := &RatingUnit{}
	// rating == nil
	if rcv, err := eventCost.getRatingForPath(nil, nil); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expecting: %+v, received: %+v", utils.ErrNotFound, err)
	} else if rcv != nil {
		t.Errorf("Expecting: nil, received: %+v", rcv)
	}
	// len(fldPath) == 0
	eratingUnit := &RatingUnit{}
	if rcv, err := eventCost.getRatingForPath([]string{}, ratingUnit); err != nil {
		t.Errorf("Expecting: nil, received: %+v", err)
	} else if !reflect.DeepEqual(eratingUnit, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", eratingUnit, rcv)
	}
	// case utils.Rates:
	eventCost = &EventCost{
		Rates: ChargedRates{
			"RatesID": RateGroups{
				&Rate{Value: 0.7},
			},
		},
	}
	eChargedRates := RateGroups{
		&Rate{Value: 0.7},
	}

	// !has || rts == nil
	ratingUnit = &RatingUnit{
		RatesID: "notfound",
	}
	if rcv, err := eventCost.getRatingForPath([]string{utils.Rates}, ratingUnit); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expecting: %+v, received: %+v", utils.ErrNotFound, err)
	} else if rcv != nil {
		t.Errorf("Expecting: nil, received: %+v", rcv)
	}
	// len(fldPath) != 1
	ratingUnit = &RatingUnit{
		RatesID: "RatesID",
	}
	if rcv, err := eventCost.getRatingForPath([]string{utils.Rates, utils.Rates}, ratingUnit); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expecting: %+v, received: %+v", utils.ErrNotFound, err)
	} else if rcv != nil {
		t.Errorf("Expecting: nil, received: %+v", rcv)
	}
	//normal case
	if rcv, err := eventCost.getRatingForPath([]string{utils.Rates}, ratingUnit); err != nil {
		t.Errorf("Expecting: nil, received: %+v", err)
	} else if !reflect.DeepEqual(eChargedRates, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eChargedRates), utils.ToJSON(rcv))
	}
	// case utils.Timing:
	eventCost = &EventCost{
		Timings: ChargedTimings{
			"test1": &ChargedTiming{
				StartTime: "StartTime",
			},
		},
	}
	eTimings := &ChargedTiming{
		StartTime: "StartTime",
	}

	// !has || tmg == nil
	ratingUnit = &RatingUnit{
		TimingID: "notfound",
	}
	if rcv, err := eventCost.getRatingForPath([]string{utils.Timing}, ratingUnit); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expecting: %+v, received: %+v", utils.ErrNotFound, err)
	} else if rcv != nil {
		t.Errorf("Expecting: nil, received: %+v", rcv)
	}
	// len(fldPath) == 1
	ratingUnit = &RatingUnit{
		TimingID: "test1",
	}
	if rcv, err := eventCost.getRatingForPath([]string{utils.Timing}, ratingUnit); err != nil {
		t.Errorf("Expecting: nil, received: %+v", err)
	} else if !reflect.DeepEqual(eTimings, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eTimings), utils.ToJSON(rcv))
	}
	//normal case
	eventCost = &EventCost{
		Timings: ChargedTimings{
			"test1": &ChargedTiming{
				Months:    utils.Months{time.April},
				StartTime: "StartTime",
			},
		},
	}
	eMonths := utils.Months{time.April}
	if rcv, err := eventCost.getRatingForPath([]string{utils.Timing, utils.MonthsFieldName}, ratingUnit); err != nil {
		t.Errorf("Expecting: nil, received: %+v", err)
	} else if !reflect.DeepEqual(eMonths, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eMonths), utils.ToJSON(rcv))
	}
	// case utils.RatingFilter:
	eventCost = &EventCost{
		RatingFilters: RatingFilters{
			"RatingFilters1": RatingMatchedFilters{
				"test1": "test1",
			},
		},
	}
	eRatingMatchedFilters := RatingMatchedFilters{
		"test1": "test1",
	}

	// !has || tmg == nil
	ratingUnit = &RatingUnit{
		TimingID: "notfound",
	}
	if rcv, err := eventCost.getRatingForPath([]string{utils.RatingFilter}, ratingUnit); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expecting: %+v, received: %+v", utils.ErrNotFound, err)
	} else if rcv != nil {
		t.Errorf("Expecting: nil, received: %+v", rcv)
	}
	// len(fldPath) == 1
	ratingUnit = &RatingUnit{
		RatingFiltersID: "RatingFilters1",
	}
	if rcv, err := eventCost.getRatingForPath([]string{utils.RatingFilter}, ratingUnit); err != nil {
		t.Errorf("Expecting: nil, received: %+v", err)
	} else if !reflect.DeepEqual(eRatingMatchedFilters, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", eRatingMatchedFilters, rcv)
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eRatingMatchedFilters), utils.ToJSON(rcv))
	}
	//normal case
	eventCost = &EventCost{
		RatingFilters: RatingFilters{
			"RatingFilters1": RatingMatchedFilters{
				"test1": "test-1",
			},
		},
	}
	if rcv, err := eventCost.getRatingForPath([]string{utils.RatingFilter, "test1"}, ratingUnit); err != nil {
		t.Errorf("Expecting: nil, received: %+v", err)
	} else if !reflect.DeepEqual("test-1", rcv) {
		t.Errorf("Expecting: test-1, received: %+v", utils.ToJSON(rcv))
	}
	//default case
	eventCost = &EventCost{
		Rates: ChargedRates{
			"RatesID": RateGroups{
				&Rate{Value: 0.7},
			},
			"RatesID2": RateGroups{
				&Rate{Value: 0.7},
			},
		},
	}
	if rcv, err := eventCost.getRatingForPath([]string{
		"unsupportedprefix"}, ratingUnit); err == nil || err.Error() != "unsupported field prefix: <unsupportedprefix>" {
		t.Errorf("Expecting: 'unsupported field prefix:  <unsupportedprefix>', received: %+v", err)
	} else if rcv != nil {
		t.Errorf("Expecting: nil, received: %+v", utils.ToJSON(rcv))
	}
	ratingUnit = &RatingUnit{
		RatesID: "RatesID",
	}
	eRate := &Rate{Value: 0.7}
	if rcv, err := eventCost.getRatingForPath([]string{"Rates[0]"}, ratingUnit); err != nil {
		t.Errorf("Expecting: nil, received: %+v", err)
	} else if !reflect.DeepEqual(eRate, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eRate), utils.ToJSON(rcv))
	}
	ratingUnit = &RatingUnit{}
	if rcv, err := eventCost.getRatingForPath([]string{"Rates[1]"}, ratingUnit); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expecting: %+v, received: %+v", utils.ErrNotFound, err)
	} else if rcv != nil {
		t.Errorf("Expecting: nil, received: %+v", rcv)
	}
}

func TestEventCostgetAcountingForPath(t *testing.T) {
	eventCost := &EventCost{}
	balanceCharge := &BalanceCharge{}
	// bc == nil
	if rcv, err := eventCost.getAcountingForPath(nil, nil); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expecting: %+v, received: %+v", utils.ErrNotFound, err)
	} else if rcv != nil {
		t.Errorf("Expecting: nil, received: %+v", rcv)
	}
	// len(fldPath) == 0
	eBalanceCharge := &BalanceCharge{}
	if rcv, err := eventCost.getAcountingForPath([]string{}, balanceCharge); err != nil {
		t.Errorf("Expecting: nil, received: %+v", err)
	} else if !reflect.DeepEqual(eBalanceCharge, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", eBalanceCharge, rcv)
	}
	// fldPath[0] == utils.Balance
	eventCost = &EventCost{
		AccountSummary: &AccountSummary{
			BalanceSummaries: BalanceSummaries{
				&BalanceSummary{
					ID: "ID",
				},
			},
		},
	}
	eBalanceSummaries := &BalanceSummary{
		ID: "ID",
	}
	//len(fldPath) == 1
	if rcv, err := eventCost.getAcountingForPath([]string{utils.Balance}, balanceCharge); err != nil {
		t.Errorf("Expecting: nil, received: %+v", err)
	} else if !reflect.DeepEqual(eBalanceSummaries, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", eBalanceSummaries, rcv)
	}
	// bl == nil
	eventCost = &EventCost{AccountSummary: &AccountSummary{}}
	if rcv, err := eventCost.getAcountingForPath([]string{utils.Balance}, balanceCharge); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expecting: %+v, received: %+v", utils.ErrNotFound, err)
	} else if rcv != nil {
		t.Errorf("Expecting: nil, received: %+v", rcv)
	}
	// len(fldPath) != 1
	eventCost = &EventCost{
		AccountSummary: &AccountSummary{
			BalanceSummaries: BalanceSummaries{
				&BalanceSummary{
					ID: "ID",
				},
			},
		},
	}
	if rcv, err := eventCost.getAcountingForPath([]string{utils.Balance, "ID"}, balanceCharge); err != nil {
		t.Errorf("Expecting: nil, received: %+v", err)
	} else if !reflect.DeepEqual("ID", rcv) {
		t.Errorf("Expecting: \"ID\", received: %+v", rcv)
	}
	//  fldPath[0] != utils.Balance
	balanceCharge = &BalanceCharge{
		AccountID: "AccountID",
	}
	if rcv, err := eventCost.getAcountingForPath([]string{utils.AccountID}, balanceCharge); err != nil {
		t.Errorf("Expecting: nil, received: %+v", err)
	} else if !reflect.DeepEqual("AccountID", rcv) {
		t.Errorf("Expecting: \"AccountID\", received: %+v", rcv)
	}
}

func TestEventCostString(t *testing.T) {
	eventCost := &EventCost{}
	eOut := `{"CGRID":"","RunID":"","StartTime":"0001-01-01T00:00:00Z","Usage":null,"Cost":null,"Charges":null,"AccountSummary":null,"Rating":null,"Accounting":null,"RatingFilters":null,"Rates":null,"Timings":null}`
	if rcv := eventCost.String(); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", eOut, rcv)
	}
	eventCost = &EventCost{
		AccountSummary: &AccountSummary{
			BalanceSummaries: BalanceSummaries{
				&BalanceSummary{
					ID: "ID",
				},
			},
		},
	}
	eOut = `{"CGRID":"","RunID":"","StartTime":"0001-01-01T00:00:00Z","Usage":null,"Cost":null,"Charges":null,"AccountSummary":{"Tenant":"","ID":"","BalanceSummaries":[{"UUID":"","ID":"ID","Type":"","Value":0,"Disabled":false}],"AllowNegative":false,"Disabled":false},"Rating":null,"Accounting":null,"RatingFilters":null,"Rates":null,"Timings":null}`
	if rcv := eventCost.String(); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", eOut, rcv)
	}
}

func TestEventCostFieldAsString(t *testing.T) {
	eventCost := &EventCost{
		CGRID: "CGRID",
	}
	eventCost.initCache()
	if rcv, err := eventCost.FieldAsString([]string{utils.CGRID}); err != nil {
		t.Error(err)
	} else if rcv != "CGRID" {
		t.Errorf("Expecting: CGRID, received: %+v", rcv)
	}
	if rcv, err := eventCost.FieldAsString([]string{"err"}); err == nil || err.Error() != "unsupported field prefix: <err>" {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, utils.EmptyString) {
		t.Errorf("Expecting: EmptyString, received: %+v", rcv)
	}
}

func TestEventCostRemoteHost(t *testing.T) {
	eventCost := &EventCost{}
	eOut := utils.LocalAddr()
	if rcv := eventCost.RemoteHost(); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", eOut, rcv)
	}

}
