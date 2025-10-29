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
		{
			RatingID: "c1a5ab9",
			Increments: []*ChargingIncrement{
				{
					Usage:          time.Duration(0),
					Cost:           0.1,
					AccountingID:   "9bdad10",
					CompressFactor: 1,
				},
				{
					Usage:          time.Duration(1 * time.Second),
					Cost:           0,
					AccountingID:   "3455b83",
					CompressFactor: 10,
				},
				{
					Usage:          time.Duration(10 * time.Second),
					Cost:           0.01,
					AccountingID:   "a012888",
					CompressFactor: 2,
				},
				{
					Usage:          time.Duration(1 * time.Second),
					Cost:           0.005,
					AccountingID:   "44d6c02",
					CompressFactor: 30,
				},
			},
			CompressFactor: 1,
		},
		{
			RatingID: "c1a5ab9",
			Increments: []*ChargingIncrement{
				{
					Usage:          time.Duration(1 * time.Second),
					Cost:           0.01,
					AccountingID:   "a012888",
					CompressFactor: 60,
				},
			},
			CompressFactor: 4,
		},
		{
			RatingID: "c1a5ab9",
			Increments: []*ChargingIncrement{
				{
					Usage:          time.Duration(1 * time.Second),
					Cost:           0,
					AccountingID:   "3455b83",
					CompressFactor: 10,
				},
				{
					Usage:          time.Duration(10 * time.Second),
					Cost:           0.01,
					AccountingID:   "a012888",
					CompressFactor: 2,
				},
				{
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
			{
				UUID:  "8c54a9e9-d610-4c82-bcb5-a315b9a65010",
				ID:    "BALANCE_1",
				Type:  utils.MONETARY,
				Value: 50,
			},
			{
				UUID:  "7a54a9e9-d610-4c82-bcb5-a315b9a65010",
				ID:    "BALANCE_2",
				Type:  utils.MONETARY,
				Value: 25,
			},
			{
				UUID:  "4b8b53d7-c1a1-4159-b845-4623a00a0165",
				ID:    "BALANCE_3",
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
			MaxCostStrategy:  utils.MAX_COST_DISCONNECT,
		},
		"c1a5ab9": &RatingUnit{
			ConnectFee:       0.1,
			RoundingMethod:   "*up",
			RoundingDecimals: 5,
			TimingID:         "7f324ab",
			RatesID:          "ec1a177",
			RatingFiltersID:  "43e77dc",
			MaxCostStrategy:  utils.MAX_COST_DISCONNECT,
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
			{
				Type:     "*monetary",
				Value:    50,
				Disabled: false},
			{
				ID:       "4b8b53d7-c1a1-4159-b845-4623a00a0165",
				Type:     "*monetary",
				Value:    25,
				Disabled: false},
			{
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
		Cost:        0.75,
		RatedUsage:  120.0,
		Timespans: TimeSpans{
			&TimeSpan{
				TimeStart: time.Date(2017, 1, 9, 16, 18, 21, 0, time.UTC),
				TimeEnd:   time.Date(2017, 1, 9, 16, 19, 21, 0, time.UTC),
				Cost:      0.15,
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
				RoundIncrement: &Increment{
					Cost: -0.1,
					BalanceInfo: &DebitInfo{
						Monetary: &MonetaryInfo{UUID: "8c54a9e9-d610-4c82-bcb5-a315b9a65010",
							ID:    utils.MetaDefault,
							Value: 9.9},
						AccountID: "cgrates.org:dan",
					},
					CompressFactor: 1,
				},
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
			{
				RatingID: "f2518464-68b8-42f4-acec-aef23d714314",
				Increments: []*ChargingIncrement{
					{
						Usage:          time.Duration(0),
						Cost:           0.1,
						AccountingID:   "44e97dec-8a7e-43d0-8b0a-736d46b5613e",
						CompressFactor: 1,
					},
					{
						Usage:          time.Duration(1 * time.Second),
						Cost:           0,
						AccountingID:   "a555cde8-4bd0-408a-afbc-c3ba64888927",
						CompressFactor: 30,
					},
					{
						Usage:          time.Duration(1 * time.Second),
						Cost:           0.005,
						AccountingID:   "906bfd0f-035c-40a3-93a8-46f71627983e",
						CompressFactor: 30,
					},
					{
						Cost:           -0.1,
						AccountingID:   "44e97dec-8a7e-43d0-8b0a-e34a152",
						CompressFactor: 1,
					},
				},
				CompressFactor: 1,
				usage:          utils.DurationPointer(time.Duration(60 * time.Second)),
				cost:           utils.Float64Pointer(0.15),
				ecUsageIdx:     utils.DurationPointer(time.Duration(0)),
			},
			{
				RatingID: "f2518464-68b8-42f4-acec-aef23d714314",
				Increments: []*ChargingIncrement{
					{
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
			"44e97dec-8a7e-43d0-8b0a-e34a152": &BalanceCharge{
				AccountID:     "cgrates.org:dan",
				BalanceUUID:   "8c54a9e9-d610-4c82-bcb5-a315b9a65010",
				RatingID:      "*rounding",
				Units:         -0.1,
				ExtraChargeID: "",
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

	// Compare to expected EC
	if !reflect.DeepEqual(eEC.Accounting[eEC.Charges[0].Increments[3].AccountingID],
		ec.Accounting[ec.Charges[0].Increments[3].AccountingID]) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eEC.Accounting[eEC.Charges[0].Increments[3].AccountingID]),
			utils.ToJSON(ec.Accounting[ec.Charges[0].Increments[3].AccountingID]))
	}
	ec.Charges[0].Increments[3].AccountingID = eEC.Charges[0].Increments[3].AccountingID
	if !reflect.DeepEqual(eEC.Charges[0].Increments[3],
		ec.Charges[0].Increments[3]) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eEC.Charges[0].Increments[3]),
			utils.ToJSON(ec.Charges[0].Increments[3]))
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
			{
				Type:     "*monetary",
				Value:    50,
				Disabled: false},
			{
				ID:       "4b8b53d7-c1a1-4159-b845-4623a00a0165",
				Type:     "*monetary",
				Value:    25,
				Disabled: false},
			{
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
			{
				RatingID: "f2518464-68b8-42f4-acec-aef23d714314",
				Increments: []*ChargingIncrement{
					{
						Usage:          time.Duration(0),
						Cost:           0.1,
						AccountingID:   "44e97dec-8a7e-43d0-8b0a-736d46b5613e",
						CompressFactor: 1,
					},
					{
						Usage:          time.Duration(1 * time.Second),
						Cost:           0,
						AccountingID:   "a555cde8-4bd0-408a-afbc-c3ba64888927",
						CompressFactor: 30,
					},
					{
						Usage:          time.Duration(1 * time.Second),
						Cost:           0.005,
						AccountingID:   "906bfd0f-035c-40a3-93a8-46f71627983e",
						CompressFactor: 30,
					},
				},
				CompressFactor: 1,
			},
			{
				RatingID: "f2518464-68b8-42f4-acec-aef23d714314",
				Increments: []*ChargingIncrement{
					{
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

func TestECAsCallCost2(t *testing.T) {
	eCC := &CallCost{
		ToR:        utils.VOICE,
		Cost:       0,
		RatedUsage: 60000000000,
		Timespans: TimeSpans{
			&TimeSpan{
				TimeStart: time.Date(2017, 1, 9, 16, 18, 21, 0, time.UTC),
				TimeEnd:   time.Date(2017, 1, 9, 16, 19, 21, 0, time.UTC),
				Cost:      0,
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
				DurationIndex:  time.Minute,
				MatchedSubject: "*out:cgrates.org:call:*any",
				MatchedPrefix:  "+49",
				MatchedDestId:  "GERMANY",
				RatingPlanId:   "RPL_RETAIL1",
				CompressFactor: 1,
				Increments: Increments{
					&Increment{ // ConnectFee
						Cost:           0,
						Duration:       time.Minute,
						BalanceInfo:    &DebitInfo{},
						CompressFactor: 1,
					},
				},
			},
		},
	}
	ec := NewEventCostFromCallCost(eCC, "cgrID", utils.MetaDefault)

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
		{
			RatingID: "c1a5ab9",
			Increments: []*ChargingIncrement{
				{
					Usage:          time.Duration(0),
					Cost:           0.1,
					AccountingID:   "9bdad10",
					CompressFactor: 1,
				},
				{
					Usage:          time.Duration(1 * time.Second),
					Cost:           0,
					AccountingID:   "3455b83",
					CompressFactor: 10,
				},
				{
					Usage:          time.Duration(10 * time.Second),
					Cost:           0.01,
					AccountingID:   "a012888",
					CompressFactor: 2,
				},
				{
					Usage:          time.Duration(1 * time.Second),
					Cost:           0.005,
					AccountingID:   "44d6c02",
					CompressFactor: 30,
				},
			},
			CompressFactor: 1,
		},
		{
			RatingID: "c1a5ab9",
			Increments: []*ChargingIncrement{
				{
					Usage:          time.Duration(1 * time.Second),
					Cost:           0.01,
					AccountingID:   "a012888",
					CompressFactor: 60,
				},
			},
			CompressFactor: 2,
		},
		{
			RatingID: "c1a5ab9",
			Increments: []*ChargingIncrement{
				{
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
		{
			RatingID: "c1a5ab9",
			Increments: []*ChargingIncrement{
				{
					Usage:          time.Duration(1 * time.Second),
					Cost:           0.01,
					AccountingID:   "a012888",
					CompressFactor: 50,
				},
			},
			CompressFactor: 1,
		},
		{
			RatingID: "c1a5ab9",
			Increments: []*ChargingIncrement{
				{
					Usage:          time.Duration(1 * time.Second),
					Cost:           0.01,
					AccountingID:   "a012888",
					CompressFactor: 60,
				},
			},
			CompressFactor: 1,
		},
		{
			RatingID: "c1a5ab9",
			Increments: []*ChargingIncrement{
				{
					Usage:          time.Duration(1 * time.Second),
					Cost:           0,
					AccountingID:   "3455b83",
					CompressFactor: 10,
				},
				{
					Usage:          time.Duration(10 * time.Second),
					Cost:           0.01,
					AccountingID:   "a012888",
					CompressFactor: 2,
				},
				{
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
			{
				RatingID: "cc68da4",
				Increments: []*ChargingIncrement{
					{
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
				{
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
			{
				RatingID: "6a83227",
				Increments: []*ChargingIncrement{
					{
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
				{
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
			{
				RatingID: "cc68da4",
				Increments: []*ChargingIncrement{
					{
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
				{
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
			{
				RatingID: "cc68da4",
				Increments: []*ChargingIncrement{
					{
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
			{
				RatingID: "6a83227",
				Increments: []*ChargingIncrement{
					{
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
			{
				RatingID: "cc68da4",
				Increments: []*ChargingIncrement{
					{
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
			{
				RatingID: "cc68da4",
				Increments: []*ChargingIncrement{
					{
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
			{
				RatingID: "6a83227",
				Increments: []*ChargingIncrement{
					{
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
			{
				RatingID: "cc68da4",
				Increments: []*ChargingIncrement{
					{
						Usage:          time.Duration(102400),
						AccountingID:   "0d87a64",
						CompressFactor: 103,
					},
				},
				CompressFactor: 2,
			},
			{
				RatingID: "cc68da4",
				Increments: []*ChargingIncrement{
					{
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
			{
				RatingID: "cc68da4",
				Increments: []*ChargingIncrement{
					{
						Usage:          time.Duration(100),
						AccountingID:   "0d87a64",
						CompressFactor: 1,
					},
				},
				CompressFactor: 1,
			},
			{
				RatingID: "cc68da4",
				Increments: []*ChargingIncrement{
					{
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
			{
				RatingID: "6a83227",
				Increments: []*ChargingIncrement{
					{
						Usage:          time.Duration(102400),
						AccountingID:   "9288f93",
						CompressFactor: 42,
					},
					{
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
			{
				RatingID: "cc68da4",
				Increments: []*ChargingIncrement{
					{
						Usage:          time.Duration(100),
						AccountingID:   "0d87a64",
						CompressFactor: 1,
					},
				},
				CompressFactor: 1,
			},
			{
				RatingID: "cc68da4",
				Increments: []*ChargingIncrement{
					{
						Usage:          time.Duration(102400),
						AccountingID:   "0d87a64",
						CompressFactor: 145,
					},
				},
				CompressFactor: 1,
			},
			{
				RatingID: "cc68da4",
				Increments: []*ChargingIncrement{
					{
						Usage:          time.Duration(10240),
						AccountingID:   "0d87a64",
						CompressFactor: 20,
					},
				},
				CompressFactor: 1,
			},
			{
				RatingID: "cc68da4",
				Increments: []*ChargingIncrement{
					{
						Usage:          time.Duration(102400),
						AccountingID:   "0d87a64",
						CompressFactor: 42,
					},
					{
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
			{
				RatingID: "cc68da4",
				Increments: []*ChargingIncrement{
					{
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
			{
				RatingID: "6a83227",
				Increments: []*ChargingIncrement{
					{
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
			{
				RatingID: "cc68da4",
				Increments: []*ChargingIncrement{
					{
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
				MaxCostStrategy:  utils.MAX_COST_DISCONNECT,
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
			{
				RatingID: "21a5ab9",
				Increments: []*ChargingIncrement{
					{
						Usage:          time.Duration(0),
						Cost:           0.1,
						AccountingID:   "9bdad10",
						CompressFactor: 1,
					},
					{
						Usage:          time.Duration(1 * time.Second),
						Cost:           0,
						AccountingID:   "3455b83",
						CompressFactor: 10,
					},
					{
						Usage:          time.Duration(10 * time.Second),
						Cost:           0.01,
						AccountingID:   "2012888",
						CompressFactor: 2,
					},
					{
						Usage:          time.Duration(1 * time.Second),
						Cost:           0.005,
						AccountingID:   "44d6c02",
						CompressFactor: 30,
					},
				},
				CompressFactor: 1,
			},
			{
				RatingID: "21a5ab9",
				Increments: []*ChargingIncrement{
					{
						Usage:          time.Duration(1 * time.Second),
						Cost:           0.01,
						AccountingID:   "2012888",
						CompressFactor: 60,
					},
				},
				CompressFactor: 4,
			},
			{
				RatingID: "21a5ab9",
				Increments: []*ChargingIncrement{
					{
						Usage:          time.Duration(1 * time.Second),
						Cost:           0,
						AccountingID:   "3455b83",
						CompressFactor: 10,
					},
					{
						Usage:          time.Duration(10 * time.Second),
						Cost:           0.01,
						AccountingID:   "2012888",
						CompressFactor: 2,
					},
					{
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
				{
					UUID:     "8c54a9e9-d610-4c82-bcb5-a315b9a65010",
					ID:       "BALANCE_1",
					Type:     utils.MONETARY,
					Value:    50,
					Disabled: false,
				},
				{
					UUID:     "7a54a9e9-d610-4c82-bcb5-a315b9a65010",
					ID:       "BALANCE_2",
					Type:     utils.MONETARY,
					Value:    25,
					Disabled: false,
				},
				{
					UUID:     "4b8b53d7-c1a1-4159-b845-4623a00a0165",
					ID:       "BALANCE_3",
					Type:     utils.VOICE,
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
				MaxCostStrategy:  utils.MAX_COST_DISCONNECT,
			},
			"21a5ab9": &RatingUnit{
				ConnectFee:       0.1,
				RoundingMethod:   "*up",
				RoundingDecimals: 5,
				TimingID:         "2f324ab",
				RatesID:          "2c1a177",
				RatingFiltersID:  "23e77dc",
				MaxCostStrategy:  utils.MAX_COST_DISCONNECT,
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
	ecDP := config.NewObjectDP(testEC, nil)
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

func TestECAsCallCost3(t *testing.T) {
	eCC := &CallCost{
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
		AccountSummary: &AccountSummary{
			Tenant: "cgrates.org",
			ID:     "dan",
			BalanceSummaries: []*BalanceSummary{
				{
					Type:     utils.MONETARY,
					Value:    50,
					Disabled: false},
				{
					ID:       "4b8b53d7-c1a1-4159-b845-4623a00a0165",
					Type:     utils.MONETARY,
					Value:    25,
					Disabled: false},
				{
					Type:     utils.VOICE,
					Value:    200,
					Disabled: false,
				},
			},
			AllowNegative: false,
			Disabled:      false,
		},
	}

	eEC := &EventCost{
		CGRID:     "164b0422fdc6a5117031b427439482c6a4f90e41",
		RunID:     utils.MetaDefault,
		StartTime: time.Date(2017, 1, 9, 16, 18, 21, 0, time.UTC),
		Charges: []*ChargingInterval{
			{
				RatingID: "21a5ab9",
				Increments: []*ChargingIncrement{
					{
						Usage:          time.Duration(0),
						Cost:           0.1,
						AccountingID:   "9bdad10",
						CompressFactor: 1,
					},
					{
						Usage:          time.Duration(1 * time.Second),
						Cost:           0,
						AccountingID:   "3455b83",
						CompressFactor: 10,
					},
					{
						Usage:          time.Duration(10 * time.Second),
						Cost:           0.01,
						AccountingID:   "2012888",
						CompressFactor: 2,
					},
					{
						Usage:          time.Duration(1 * time.Second),
						Cost:           0.005,
						AccountingID:   "44d6c02",
						CompressFactor: 30,
					},
				},
				CompressFactor: 1,
			},
			{
				RatingID: "21a5ab9",
				Increments: []*ChargingIncrement{
					{
						Usage:          time.Duration(1 * time.Second),
						Cost:           0.01,
						AccountingID:   "2012888",
						CompressFactor: 60,
					},
				},
				CompressFactor: 4,
			},
			{
				RatingID: "21a5ab9",
				Increments: []*ChargingIncrement{
					{
						Usage:          time.Duration(1 * time.Second),
						Cost:           0,
						AccountingID:   "3455b83",
						CompressFactor: 10,
					},
					{
						Usage:          time.Duration(10 * time.Second),
						Cost:           0.01,
						AccountingID:   "2012888",
						CompressFactor: 2,
					},
					{
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
				{
					UUID:     "8c54a9e9-d610-4c82-bcb5-a315b9a65010",
					Type:     utils.MONETARY,
					Value:    50,
					Disabled: false},
				{
					UUID:     "7a54a9e9-d610-4c82-bcb5-a315b9a65010",
					Type:     utils.MONETARY,
					Value:    25,
					Disabled: false},
				{
					UUID:     "4b8b53d7-c1a1-4159-b845-4623a00a0165",
					Type:     utils.VOICE,
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
	ec := NewEventCostFromCallCost(eCC, "cgrID", utils.MetaDefault)
	eEC.SyncKeys(ec)
	ec.Merge(eEC)

	if _, err := ec.Trim(time.Second); err != nil {
		t.Fatal(err)
	}

}

func TestECAsDataProvider2(t *testing.T) {
	ecDP := testEC
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
	ecDP = &EventCost{AccountSummary: &AccountSummary{}}
	ecDP.initCache()
	if _, err := ecDP.FieldAsInterface([]string{"AccountSummary", "BalanceSummaries[0]", "Type"}); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Unexpected error:%v", err)
	}
	if _, err := ecDP.FieldAsInterface([]string{"Charges[0]", "Increments"}); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Unexpected error:%v", err)
	}
	ecDP.Charges = []*ChargingInterval{{}}
	if _, err := ecDP.FieldAsInterface([]string{"Charges[0]", "Increments[0]"}); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Unexpected error:%v", err)
	}
	ecDP.Rating = Rating{"": {}}
	ecDP.Rates = ChargedRates{"": {}, "b": {}}
	if _, err := ecDP.FieldAsInterface([]string{"Charges[0]", "Rating", "Rates[0]"}); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Unexpected error:%v", err)
	}

	if _, err := ecDP.FieldAsInterface([]string{"Rates", "b[0]"}); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Unexpected error:%v", err)
	}
}

func TestECAsDataProviderTT(t *testing.T) {
	ecDP := testEC
	testDPs := []struct {
		name   string
		fields []string
		exp    any
	}{
		{
			name:   "StartTime",
			fields: []string{"Charges[0]", "Rating", "Timing", "StartTime"},
			exp:    "00:00:00",
		},
		{
			name:   "RateIncrement",
			fields: []string{"Charges[0]", "Rating", "Rates[0]", "RateIncrement"},
			exp:    "1m0s",
		},
		{
			name:   "DestinationID",
			fields: []string{"Charges[0]", "Rating", "RatingFilter", "DestinationID"},
			exp:    "GERMANY",
		},
		{
			name:   "Units",
			fields: []string{"Charges[0]", "Increments[0]", "Accounting", "Units"},
			exp:    "0.1",
		},
		{
			name:   "Subject",
			fields: []string{"Charges[0]", "Rating", "RatingFilter", "Subject"},
			exp:    "*out:cgrates.org:call:*any",
		},
		{
			name:   "Value",
			fields: []string{"Charges[2]", "Increments[2]", "Accounting", "Balance", "Value"},
			exp:    "200",
		},
		{
			name:   "Type",
			fields: []string{"Charges[1]", "Increments[0]", "Accounting", "Balance", "Type"},
			exp:    "*monetary",
		},
		{
			name:   "UUID",
			fields: []string{"Charges[0]", "Increments[3]", "Accounting", "Balance", "UUID"},
			exp:    "4b8b53d7-c1a1-4159-b845-4623a00a0165",
		},
		{
			name:   "ID",
			fields: []string{"Charges[0]", "Increments[3]", "Accounting", "Balance", "ID"},
			exp:    "BALANCE_3",
		},
		{
			name:   "Disabled",
			fields: []string{"Charges[0]", "Increments[3]", "Accounting", "Balance", "Disabled"},
			exp:    "false",
		},
		{
			name:   "IncrementCost",
			fields: []string{"Charges[0]", "Increments[2]", "Cost"},
			exp:    "0.01",
		},
		{
			name:   "DestinationPrefix",
			fields: []string{"Charges[2]", "Rating", "RatingFilter", "DestinationPrefix"},
			exp:    "+49",
		},
		{
			name:   "RatingPlanID",
			fields: []string{"Charges[2]", "Rating", "RatingFilter", "RatingPlanID"},
			exp:    "RPL_RETAIL1",
		},
		{
			name:   "AccountID",
			fields: []string{"Charges[0]", "Increments[3]", "Accounting", "AccountID"},
			exp:    "cgrates.org:dan",
		},
		{
			name:   "RateValue",
			fields: []string{"Charges[1]", "Rating", "Rates[0]", "Value"},
			exp:    "0.01",
		},
		{
			name:   "RateUnit",
			fields: []string{"Charges[2]", "Rating", "Rates[0]", "RateUnit"},
			exp:    "1s",
		},
		{
			name:   "ExtraCharge",
			fields: []string{"Charges[0]", "Increments[1]", "Accounting", "ExtraChargeID"},
			exp:    "*none",
		},
		{
			name:   "AccountSummary",
			fields: []string{"AccountSummary", "BalanceSummaries[1]", "Value"},
			exp:    "25",
		},
		{
			name:   "RoundingMethod",
			fields: []string{"AccountSummary", "AllowNegative"},
			exp:    "false",
		},
		{
			name:   "CompressFactor",
			fields: []string{"Charges[0]", "CompressFactor"},
			exp:    "1",
		},
		{
			name:   "ConnectFee through Accounting",
			fields: []string{"Charges[0]", "Increments[3]", "Accounting", "Rating", "ConnectFee"},
			exp:    "0",
		},
		{
			name:   "RoundingMethod through Accounting",
			fields: []string{"Charges[0]", "Increments[3]", "Accounting", "Rating", "RoundingMethod"},
			exp:    utils.ROUNDING_UP,
		},
		{
			name:   "RoundingDecimals through Accounting",
			fields: []string{"Charges[0]", "Increments[3]", "Accounting", "Rating", "RoundingDecimals"},
			exp:    "5",
		},
		{
			name:   "MaxCost through Accounting",
			fields: []string{"Charges[0]", "Increments[3]", "Accounting", "Rating", "MaxCost"},
			exp:    "0",
		},
		{
			name:   "MaxCostStrategy through Accounting",
			fields: []string{"Charges[0]", "Increments[3]", "Accounting", "Rating", "MaxCostStrategy"},
			exp:    utils.MAX_COST_DISCONNECT,
		},
		{
			name:   "TimingID through Accounting",
			fields: []string{"Charges[0]", "Increments[3]", "Accounting", "Rating", "TimingID"},
			exp:    "7f324ab",
		},
		{
			name:   "RatesID through Accounting",
			fields: []string{"Charges[0]", "Increments[3]", "Accounting", "Rating", "RatesID"},
			exp:    "4910ecf",
		},
		{
			name:   "RatingFiltersID through Accounting",
			fields: []string{"Charges[0]", "Increments[3]", "Accounting", "Rating", "RatingFiltersID"},
			exp:    "43e77dc",
		},
		{
			name:   "ConnectFee",
			fields: []string{"Charges[0]", "Rating", "ConnectFee"},
			exp:    "0.1",
		},
		{
			name:   "RoundingMethod",
			fields: []string{"Charges[0]", "Rating", "RoundingMethod"},
			exp:    utils.ROUNDING_UP,
		},
		{
			name:   "RoundingDecimals",
			fields: []string{"Charges[0]", "Rating", "RoundingDecimals"},
			exp:    "5",
		},
		{
			name:   "MaxCost",
			fields: []string{"Charges[0]", "Rating", "MaxCost"},
			exp:    "0",
		},
		{
			name:   "MaxCostStrategy",
			fields: []string{"Charges[0]", "Rating", "MaxCostStrategy"},
			exp:    utils.MAX_COST_DISCONNECT,
		},
		{
			name:   "TimingID",
			fields: []string{"Charges[0]", "Rating", "TimingID"},
			exp:    "7f324ab",
		},
		{
			name:   "RatesID",
			fields: []string{"Charges[0]", "Rating", "RatesID"},
			exp:    "ec1a177",
		},
		{
			name:   "RatingFiltersID",
			fields: []string{"Charges[0]", "Rating", "RatingFiltersID"},
			exp:    "43e77dc",
		},
		{
			name:   "DestinationID through Accounting",
			fields: []string{"Charges[0]", "Increments[3]", "Accounting", "Rating", "RatingFilter", "DestinationID"},
			exp:    "GERMANY",
		},
		{
			name:   "Value through ExtraCharge",
			fields: []string{"Charges[0]", "Increments[3]", "Accounting", "ExtraCharge", "Balance", "Value"},
			exp:    "50",
		},
		{
			name:   "Type through ExtraCharge",
			fields: []string{"Charges[0]", "Increments[3]", "Accounting", "ExtraCharge", "Balance", "Type"},
			exp:    utils.MONETARY,
		},
		{
			name:   "UUID through ExtraCharge",
			fields: []string{"Charges[0]", "Increments[3]", "Accounting", "ExtraCharge", "Balance", "UUID"},
			exp:    "8c54a9e9-d610-4c82-bcb5-a315b9a65010",
		},
		{
			name:   "ID through ExtraCharge",
			fields: []string{"Charges[0]", "Increments[3]", "Accounting", "ExtraCharge", "Balance", "ID"},
			exp:    "BALANCE_1",
		},
		{
			name:   "Disabled through ExtraCharge",
			fields: []string{"Charges[0]", "Increments[3]", "Accounting", "ExtraCharge", "Balance", "Disabled"},
			exp:    "false",
		},
		{
			name:   "AccountID through ExtraCharge",
			fields: []string{"Charges[0]", "Increments[3]", "Accounting", "ExtraCharge", "AccountID"},
			exp:    "cgrates.org:dan",
		},
		{
			name:   "Units through ExtraCharge",
			fields: []string{"Charges[0]", "Increments[3]", "Accounting", "ExtraCharge", "Units"},
			exp:    "0.005",
		},
	}

	for _, testDp := range testDPs {

		t.Run(testDp.name, func(t *testing.T) {
			if val, err := ecDP.FieldAsString(testDp.fields); err != nil {
				t.Error(err)
			} else if testDp.exp != val {
				t.Errorf("Expecting: <%s> \nreceived: <%s>", testDp.exp, val)
			}
		})
	}
}

func TestECFieldAsInterfaceNilEventCost(t *testing.T) {
	dft, _ := config.NewDefaultCGRConfig()
	cdr, err := NewMapEvent(map[string]any{}).AsCDR(dft, "cgrates.org", "UTC")
	if err != nil {
		t.Fatal(err)
	}
	nM := cdr.AsMapStorage()
	if _, err := nM.FieldAsInterface([]string{"*ec", "Charges[0]", "Increments[0]", "Accounting", "Balance", "ID"}); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Expected error:%s, received: %v", utils.ErrNotFound.Error(), err)
	}

	if _, err := nM.FieldAsInterface([]string{"*req", "CostDetails", "Charges[0]", "Increments[0]", "Accounting", "Balance", "ID"}); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Expected error:%s, received: %v", utils.ErrNotFound.Error(), err)
	}
}

func TestEventCostFieldAsInterface(t *testing.T) {
	ec := EventCost{
		cache: utils.MapStorage{"test": "val1"},
	}

	_, err := ec.FieldAsInterface([]string{"test", "val1"})

	if err != nil {
		if err.Error() != "WRONG_PATH" {
			t.Error(err)
		}
	}

	ec = EventCost{
		cache: utils.MapStorage{"test[": nil},
	}

	_, err = ec.FieldAsInterface([]string{"test["})

	if err != nil {
		if err.Error() != "NOT_FOUND" {
			t.Error(err)
		}
	}
}

func TestEventCostfieldAsInterface(t *testing.T) {
	td := 1 * time.Second
	fl := 1.2
	as := AccountSummary{ID: "test"}

	ec := EventCost{
		CGRID:          "test",
		RunID:          "test",
		StartTime:      time.Date(1999, 8, 28, 1, 12, 34, 1334, time.Local),
		Usage:          &td,
		Cost:           &fl,
		Charges:        []*ChargingInterval{{RatingID: "test"}},
		AccountSummary: &as,
		Rating:         Rating{},
		Accounting:     Accounting{},
		RatingFilters:  RatingFilters{},
		Rates:          ChargedRates{},
		Timings:        ChargedTimings{},
	}

	tests := []struct {
		name string
		arg  []string
		val  any
		err  string
	}{
		{
			name: "get path index error check",
			arg:  []string{"test"},
			val:  nil,
			err:  "unsupported field prefix: <test>",
		},
		{
			name: "case charges",
			arg:  []string{"Charges"},
			val:  []*ChargingInterval{{RatingID: "test"}},
			err:  "",
		},
		{
			name: "case StartTime",
			arg:  []string{"StartTime"},
			val:  time.Date(1999, 8, 28, 1, 12, 34, 1334, time.Local),
			err:  "",
		},
		{
			name: "case Usage",
			arg:  []string{"Usage"},
			val:  &td,
			err:  "",
		},
		{
			name: "case Cost",
			arg:  []string{"Cost"},
			val:  &fl,
			err:  "",
		},
		{
			name: "case charges error",
			arg:  []string{"Charges", "test"},
			val:  nil,
			err:  "NOT_FOUND",
		},
		{
			name: "case CGRID error",
			arg:  []string{"CGRID", "test"},
			val:  nil,
			err:  "NOT_FOUND",
		},
		{
			name: "case RunID error",
			arg:  []string{"RunID", "test"},
			val:  nil,
			err:  "NOT_FOUND",
		},
		{
			name: "case StartTime error",
			arg:  []string{"StartTime", "test"},
			val:  nil,
			err:  "NOT_FOUND",
		},
		{
			name: "case Usage error",
			arg:  []string{"Usage", "test"},
			val:  nil,
			err:  "NOT_FOUND",
		},
		{
			name: "case Cost error",
			arg:  []string{"Cost", "test"},
			val:  nil,
			err:  "NOT_FOUND",
		},
		{
			name: "case AccountSummary error",
			arg:  []string{"AccountSummary"},
			val:  &as,
			err:  "",
		},
		{
			name: "case Timings error",
			arg:  []string{"Timings"},
			val:  ChargedTimings{},
			err:  "",
		},
		{
			name: "case Timings",
			arg:  []string{"Timings", "test"},
			val:  nil,
			err:  "NOT_FOUND",
		},
		{
			name: "case Rates error",
			arg:  []string{"Rates"},
			val:  ChargedRates{},
			err:  "",
		},
		{
			name: "case RatingFilters error",
			arg:  []string{"RatingFilters"},
			val:  RatingFilters{},
			err:  "",
		},
		{
			name: "case RatingFilters error",
			arg:  []string{"RatingFilters", "test"},
			val:  nil,
			err:  "NOT_FOUND",
		},
		{
			name: "case Accounting error",
			arg:  []string{"Accounting"},
			val:  Accounting{},
			err:  "",
		},
		{
			name: "case Accounting error",
			arg:  []string{"Accounting", "test"},
			val:  nil,
			err:  "NOT_FOUND",
		},
		{
			name: "case Rating error",
			arg:  []string{"Rating"},
			val:  Rating{},
			err:  "",
		},
		{
			name: "case Rating error",
			arg:  []string{"Rating", "test"},
			val:  nil,
			err:  "NOT_FOUND",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rcv, err := ec.fieldAsInterface(tt.arg)

			if err != nil {
				if err.Error() != tt.err {
					t.Fatal(err)
				}
			}

			if !reflect.DeepEqual(rcv, tt.val) {
				t.Errorf("expected %v, received %v", tt.val, rcv)
			}
		})
	}
}

func TestEventCostgetChargesForPath(t *testing.T) {
	ec := EventCost{}
	chr := ChargingInterval{
		CompressFactor: 1,
		Increments: []*ChargingIncrement{
			{CompressFactor: 1},
		},
	}

	tests := []struct {
		name string
		fl   []string
		chr  *ChargingInterval
		exp  any
		err  string
	}{
		{
			name: "nil charging interval",
			fl:   []string{},
			chr:  nil,
			exp:  nil,
			err:  "NOT_FOUND",
		},
		{
			name: "empty field path",
			fl:   []string{},
			chr:  &chr,
			exp:  &chr,
			err:  "",
		},
		{
			name: "compress factor with second non existing field",
			fl:   []string{"CompressFactor", "test"},
			chr:  &chr,
			exp:  nil,
			err:  "NOT_FOUND",
		},
		{
			name: "compress factor case",
			fl:   []string{"CompressFactor"},
			chr:  &chr,
			exp:  1,
			err:  "",
		},
		{
			name: "index on non existing fiels",
			fl:   []string{"test[0]"},
			chr:  &chr,
			exp:  nil,
			err:  "unsupported field prefix: <test>",
		},
		{
			name: "increments case with index",
			fl:   []string{"Increments[0]"},
			chr:  &chr,
			exp:  &ChargingIncrement{CompressFactor: 1},
			err:  "",
		},
		{
			name: "increments case no index",
			fl:   []string{"Increments"},
			chr:  &chr,
			exp: []*ChargingIncrement{
				{CompressFactor: 1},
			},
			err: "",
		},
		{
			name: "index is nill",
			fl:   []string{"Increments", "test"},
			chr:  &chr,
			exp:  nil,
			err:  "NOT_FOUND",
		},
		{
			name: "extrafield in increments",
			fl:   []string{"Increments[0]", "CompressFactor"},
			chr:  &chr,
			exp:  1,
			err:  "NOT_FOUND",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rcv, err := ec.getChargesForPath(tt.fl, tt.chr)

			if err != nil {
				if err.Error() != tt.err {
					t.Errorf("error: want %v, got %v", tt.err, err)
				}
			}

			if !reflect.DeepEqual(rcv, tt.exp) {
				t.Errorf("expected %v, received %v", tt.exp, rcv)
			}
		})
	}
}

func TestEventCostgetRatingForPath(t *testing.T) {
	tm := 1 * time.Second
	r := RatingUnit{
		RatesID: "1",
	}
	r2 := RatingUnit{
		RatesID:         "2",
		TimingID:        "timingID",
		RatingFiltersID: "rf",
	}
	ec := EventCost{
		Rates: ChargedRates{
			"2": {{
				GroupIntervalStart: tm,
				Value:              1.2,
				RateIncrement:      tm,
				RateUnit:           tm,
			}},
		},
		Timings: ChargedTimings{
			"timingID": {StartTime: "00:00:00"},
		},
		RatingFilters: RatingFilters{
			"rf": {"test": 1},
		},
	}

	tests := []struct {
		name    string
		fldPath []string
		rating  *RatingUnit
		exp     any
		err     string
	}{
		{
			name:    "nil rating",
			fldPath: []string{},
			rating:  nil,
			exp:     nil,
			err:     "NOT_FOUND",
		},
		{
			name:    "empty fldPath",
			fldPath: []string{},
			rating:  &r,
			exp:     &r,
			err:     "",
		},
		{
			name:    "non exisitng field with index",
			fldPath: []string{"test[0]"},
			rating:  &r,
			exp:     nil,
			err:     "unsupported field prefix: <test[0]>",
		},
		{
			name:    "rating not found in event cost",
			fldPath: []string{"Rates[0]"},
			rating:  &r,
			exp:     nil,
			err:     "NOT_FOUND",
		},
		{
			name:    "rating found",
			fldPath: []string{"Rates[0]"},
			rating:  &r2,
			exp: &Rate{
				GroupIntervalStart: tm,
				Value:              1.2,
				RateIncrement:      tm,
				RateUnit:           tm,
			},
			err: "",
		},
		{
			name:    "rates case not found in event cost",
			fldPath: []string{"Rates"},
			rating:  &r,
			exp:     nil,
			err:     "NOT_FOUND",
		},
		{
			name:    "rates case field path length diff from 1",
			fldPath: []string{"Rates", "test"},
			rating:  &r2,
			exp:     nil,
			err:     "NOT_FOUND",
		},
		{
			name:    "rates case found",
			fldPath: []string{"Rates"},
			rating:  &r2,
			exp: RateGroups{{
				GroupIntervalStart: tm,
				Value:              1.2,
				RateIncrement:      tm,
				RateUnit:           tm,
			}},
			err: "",
		},
		{
			name:    "Timing case not found in event cost",
			fldPath: []string{"Timing"},
			rating:  &r,
			exp:     nil,
			err:     "NOT_FOUND",
		},
		{
			name:    "Timing case field path length is 1",
			fldPath: []string{"Timing"},
			rating:  &r2,
			exp:     &ChargedTiming{StartTime: "00:00:00"},
			err:     "",
		},
		{
			name:    "Timing case extrafield found",
			fldPath: []string{"Timing", "StartTime"},
			rating:  &r2,
			exp:     "00:00:00",
			err:     "",
		},
		{
			name:    "RatingFilter case not found in event cost",
			fldPath: []string{"RatingFilter"},
			rating:  &r,
			exp:     nil,
			err:     "NOT_FOUND",
		},
		{
			name:    "RatingFilter case field path length is 1",
			fldPath: []string{"RatingFilter"},
			rating:  &r2,
			exp:     RatingMatchedFilters{"test": 1},
			err:     "",
		},
		{
			name:    "RatingFilter case extrafield found",
			fldPath: []string{"RatingFilter", "test"},
			rating:  &r2,
			exp:     1,
			err:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rcv, err := ec.getRatingForPath(tt.fldPath, tt.rating)

			if err != nil {
				if err.Error() != tt.err {
					t.Errorf("error: want %v, got %v", tt.err, err)
				}
			}

			if !reflect.DeepEqual(rcv, tt.exp) {
				t.Errorf("expected %v, received %v", tt.exp, rcv)
			}
		})
	}
}

func TestEventCostgetAcountingForPath(t *testing.T) {
	ec := EventCost{
		AccountSummary: &AccountSummary{
			BalanceSummaries: BalanceSummaries{{
				UUID: "bl2",
			}},
		},
	}
	bc := BalanceCharge{
		BalanceUUID: "bl",
	}
	bc2 := BalanceCharge{
		BalanceUUID: "bl2",
	}

	tests := []struct {
		name    string
		fldPath []string
		bc      *BalanceCharge
		exp     any
		err     string
	}{
		{
			name:    "nil balance charge",
			fldPath: []string{},
			bc:      nil,
			exp:     nil,
			err:     "NOT_FOUND",
		},
		{
			name:    "empty fldPath",
			fldPath: []string{},
			bc:      &bc,
			exp:     &bc,
			err:     "NOT_FOUND",
		},
		{
			name:    "balance not found in event cost",
			fldPath: []string{"Balance"},
			bc:      &bc,
			exp:     nil,
			err:     "NOT_FOUND",
		},
		{
			name:    "balance not found in event cost",
			fldPath: []string{"Balance"},
			bc:      &bc2,
			exp: &BalanceSummary{
				UUID: "bl2",
			},
			err: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rcv, err := ec.getAcountingForPath(tt.fldPath, tt.bc)

			if err != nil {
				if err.Error() != tt.err {
					t.Error(err)
				}
			}

			if !reflect.DeepEqual(rcv, tt.exp) {
				t.Errorf("expected %v, received %v", tt.exp, rcv)
			}
		})
	}
}

func TestEventCostString(t *testing.T) {
	tm := 1 * time.Second
	fl := 1.2
	e := EventCost{
		CGRID:          "test",
		RunID:          "test",
		StartTime:      time.Date(2021, 8, 15, 14, 30, 45, 100, time.UTC),
		Usage:          &tm,
		Cost:           &fl,
		Charges:        []*ChargingInterval{},
		AccountSummary: &AccountSummary{},
		Rating:         Rating{},
		Accounting:     Accounting{},
		RatingFilters:  RatingFilters{},
		Rates:          ChargedRates{},
		Timings:        ChargedTimings{},
	}

	rcv := e.String()
	exp := `{"CGRID":"test","RunID":"test","StartTime":"2021-08-15T14:30:45.0000001Z","Usage":1000000000,"Cost":1.2,"Charges":[],"AccountSummary":{"Tenant":"","ID":"","BalanceSummaries":null,"AllowNegative":false,"Disabled":false},"Rating":{},"Accounting":{},"RatingFilters":{},"Rates":{},"Timings":{}}`

	if rcv != exp {
		t.Errorf("expected %s, received %s", exp, rcv)
	}
}

func TestEventCostFieldAsString(t *testing.T) {
	e := EventCost{
		RunID: "test",
		cache: utils.MapStorage{},
	}

	rcv, err := e.FieldAsString([]string{"RunID"})
	if err != nil {
		if err.Error() != "err" {
			t.Error(err)
		}
	}

	if rcv != "test" {
		t.Error(rcv)
	}

	rcv, err = e.FieldAsString([]string{"test"})
	if err != nil {
		if err.Error() != "unsupported field prefix: <test>" {
			t.Error(err)
		}
	}

	if rcv != "" {
		t.Error(rcv)
	}
}

func TestEventCostRemoteHost(t *testing.T) {
	ec := EventCost{}

	rcv := ec.RemoteHost()

	if rcv.String() != "local" {
		t.Error(rcv.String())
	}
}

func TestEventCostnewChargingIncrement(t *testing.T) {
	e := EventCost{}

	rcv := e.newChargingIncrement(&Increment{
		Duration:       1 * time.Second,
		Cost:           1,
		CompressFactor: 1,
	}, RatingMatchedFilters{}, false)
	exp := &ChargingIncrement{
		Usage:          1 * time.Second,
		Cost:           1,
		CompressFactor: 1,
	}

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected %v, received %v", exp, rcv)
	}
}

func TestEventCostAsRefundIncrements(t *testing.T) {
	str := "test"
	td := 1 * time.Second
	ec := &EventCost{
		CGRID:     str,
		RunID:     str,
		StartTime: time.Date(2021, 8, 15, 14, 30, 45, 100, time.Local),
		Usage:     &td,
	}

	rcv := ec.AsRefundIncrements(str)
	exp := &CallDescriptor{
		CgrID:         ec.CGRID,
		RunID:         ec.RunID,
		ToR:           str,
		TimeStart:     ec.StartTime,
		TimeEnd:       ec.StartTime.Add(ec.GetUsage()),
		DurationIndex: ec.GetUsage(),
	}

	if !reflect.DeepEqual(exp, rcv) {
		t.Errorf("\nexpected: %s\nreceived: %s\n", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}

func TestEventCostappendChargingIntervalFromEventCost(t *testing.T) {
	str := "test"
	td := 1 * time.Second
	fl := 1.2
	str2 := "test2"
	td2 := 2 * time.Millisecond
	ec := &EventCost{
		CGRID:     str,
		RunID:     str,
		StartTime: time.Date(2021, 8, 15, 14, 30, 45, 100, time.Local),
		Usage:     &td,
		Charges: []*ChargingInterval{{
			RatingID: str,
			Increments: []*ChargingIncrement{{
				Usage:          td,
				Cost:           fl,
				AccountingID:   str,
				CompressFactor: 1,
			}},
			CompressFactor: 1,
			usage:          &td,
			ecUsageIdx:     &td,
			cost:           &fl,
		},
		}}
	oEC := &EventCost{
		CGRID:     str2,
		RunID:     str2,
		StartTime: time.Date(2022, 8, 15, 14, 30, 45, 100, time.Local),
		Usage:     &td2,
		Charges: []*ChargingInterval{{
			RatingID: str,
			Increments: []*ChargingIncrement{{
				Usage:          td,
				Cost:           fl,
				AccountingID:   str,
				CompressFactor: 1,
			}},
			CompressFactor: 1,
			usage:          &td,
			ecUsageIdx:     &td,
			cost:           &fl,
		}},
	}

	ec.appendChargingIntervalFromEventCost(oEC, 0)

	if ec.Charges[0].CompressFactor != 2 {
		t.Error("didn't append charging interval from event cost")
	}
}
