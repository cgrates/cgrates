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
		{
			RatingID: "c1a5ab9",
			Increments: []*ChargingIncrement{
				{
					Usage:          0,
					Cost:           0.1,
					AccountingID:   "9bdad10",
					CompressFactor: 1,
				},
				{
					Usage:          time.Second,
					Cost:           0,
					AccountingID:   "3455b83",
					CompressFactor: 10,
				},
				{
					Usage:          10 * time.Second,
					Cost:           0.01,
					AccountingID:   "a012888",
					CompressFactor: 2,
				},
				{
					Usage:          time.Second,
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
					Usage:          time.Second,
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
					Usage:          time.Second,
					Cost:           0,
					AccountingID:   "3455b83",
					CompressFactor: 10,
				},
				{
					Usage:          10 * time.Second,
					Cost:           0.01,
					AccountingID:   "a012888",
					CompressFactor: 2,
				},
				{
					Usage:          time.Second,
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
				Type:  utils.MetaMonetary,
				Value: 50,
			},
			{
				UUID:  "7a54a9e9-d610-4c82-bcb5-a315b9a65010",
				Type:  utils.MetaMonetary,
				Value: 25,
			},
			{
				UUID:  "4b8b53d7-c1a1-4159-b845-4623a00a0165",
				Type:  utils.MetaVoice,
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
			&RGRate{
				GroupIntervalStart: 0,
				Value:              0.01,
				RateIncrement:      time.Minute,
				RateUnit:           time.Second},
		},
		"4910ecf": RateGroups{
			&RGRate{
				GroupIntervalStart: 0,
				Value:              0.005,
				RateIncrement:      time.Second,
				RateUnit:           time.Second},
			&RGRate{
				GroupIntervalStart: 60 * time.Second,
				Value:              0.005,
				RateIncrement:      time.Second,
				RateUnit:           time.Second},
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
	ec.Usage = utils.DurationPointer(time.Second)
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
	eEc.Usage = utils.DurationPointer(10 * time.Minute)
	eEc.Cost = utils.Float64Pointer(3.52)
	eEc.Charges[0].ecUsageIdx = utils.DurationPointer(0)
	eEc.Charges[0].usage = utils.DurationPointer(time.Minute)
	eEc.Charges[0].cost = utils.Float64Pointer(0.27)
	eEc.Charges[1].ecUsageIdx = utils.DurationPointer(time.Minute)
	eEc.Charges[1].usage = utils.DurationPointer(time.Minute)
	eEc.Charges[1].cost = utils.Float64Pointer(0.6)
	eEc.Charges[2].ecUsageIdx = utils.DurationPointer(5 * time.Minute)
	eEc.Charges[2].usage = utils.DurationPointer(time.Minute)
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
		ToR:         utils.MetaVoice,
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
							&RGRate{
								GroupIntervalStart: 0,
								Value:              0.01,
								RateUnit:           time.Second,
								RateIncrement:      time.Minute,
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
				RoundIncrement: &Increment{
					Cost: 0.1,
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
						Duration: time.Second,
						Cost:     0,
						BalanceInfo: &DebitInfo{
							Unit: &UnitInfo{
								UUID:     "9d54a9e9-d610-4c82-bcb5-a315b9a65089",
								ID:       "free_mins",
								Value:    0,
								Consumed: 1.0,
								ToR:      utils.MetaVoice,
							},
							AccountID: "cgrates.org:dan",
						},
						CompressFactor: 30,
					},
					&Increment{ // Minutes with special price
						Duration: time.Second,
						Cost:     0.005,
						BalanceInfo: &DebitInfo{
							Unit: &UnitInfo{ // Minutes with special price
								UUID:     "7a54a9e9-d610-4c82-bcb5-a315b9a65010",
								ID:       "discounted_mins",
								Value:    0,
								Consumed: 1.0,
								ToR:      utils.MetaVoice,
								RateInterval: &RateInterval{
									Timing: &RITiming{
										StartTime: "00:00:00",
									},
									Rating: &RIRate{
										ConnectFee:       0,
										RoundingMethod:   "*up",
										RoundingDecimals: 5,
										Rates: RateGroups{
											&RGRate{
												GroupIntervalStart: 0,
												Value:              0.005,
												RateUnit:           time.Second,
												RateIncrement:      time.Second,
											},
											&RGRate{
												GroupIntervalStart: 60 * time.Second,
												Value:              0.005,
												RateUnit:           time.Second,
												RateIncrement:      time.Second,
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
							&RGRate{
								GroupIntervalStart: 0,
								Value:              0.01,
								RateUnit:           time.Second,
								RateIncrement:      time.Minute,
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
					&Increment{
						Cost:     0.01,
						Duration: time.Second,
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
		Usage:     utils.DurationPointer(2 * time.Minute),
		Charges: []*ChargingInterval{
			{
				RatingID: "f2518464-68b8-42f4-acec-aef23d714314",
				Increments: []*ChargingIncrement{
					{
						Usage:          0,
						Cost:           0.1,
						AccountingID:   "44e97dec-8a7e-43d0-8b0a-736d46b5613e",
						CompressFactor: 1,
					},
					{
						Usage:          time.Second,
						Cost:           0,
						AccountingID:   "a555cde8-4bd0-408a-afbc-c3ba64888927",
						CompressFactor: 30,
					},
					{
						Usage:          time.Second,
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
				usage:          utils.DurationPointer(60 * time.Second),
				cost:           utils.Float64Pointer(0.15),
				ecUsageIdx:     utils.DurationPointer(0),
			},
			{
				RatingID: "f2518464-68b8-42f4-acec-aef23d714314",
				Increments: []*ChargingIncrement{
					{
						Usage:          time.Second,
						Cost:           0.01,
						AccountingID:   "c890a899-df43-497a-9979-38492713f57b",
						CompressFactor: 60,
					},
				},
				CompressFactor: 1,
				usage:          utils.DurationPointer(60 * time.Second),
				cost:           utils.Float64Pointer(0.6),
				ecUsageIdx:     utils.DurationPointer(60 * time.Second),
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
				Units:         0.1,
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
				&RGRate{
					GroupIntervalStart: 0,
					Value:              0.01,
					RateIncrement:      time.Minute,
					RateUnit:           time.Second},
			},
			"e5eb0f1c-3612-4e8c-b749-7f8f41dd90d4": RateGroups{
				&RGRate{
					GroupIntervalStart: 0,
					Value:              0.005,
					RateIncrement:      time.Second,
					RateUnit:           time.Second},
				&RGRate{
					GroupIntervalStart: 60 * time.Second,
					Value:              0.005,
					RateIncrement:      time.Second,
					RateUnit:           time.Second},
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
		ToR:           utils.MetaVoice,
		TimeStart:     time.Date(2017, 1, 9, 16, 18, 21, 0, time.UTC),
		TimeEnd:       time.Date(2017, 1, 9, 16, 28, 21, 0, time.UTC),
		DurationIndex: 10 * time.Minute,
	}
	eCD.Increments = Increments{
		&Increment{
			Duration:       0,
			Cost:           0.1,
			CompressFactor: 1,
			BalanceInfo: &DebitInfo{
				AccountID: "cgrates.org:dan",
				Monetary: &MonetaryInfo{
					UUID: "8c54a9e9-d610-4c82-bcb5-a315b9a65010"}},
		},
		&Increment{
			Duration:       time.Second,
			Cost:           0,
			CompressFactor: 10,
			BalanceInfo: &DebitInfo{
				AccountID: "cgrates.org:dan",
				Unit: &UnitInfo{
					UUID: "4b8b53d7-c1a1-4159-b845-4623a00a0165"}},
		},
		&Increment{
			Duration:       10 * time.Second,
			Cost:           0.01,
			CompressFactor: 2,
			BalanceInfo: &DebitInfo{
				AccountID: "cgrates.org:dan",
				Monetary: &MonetaryInfo{
					UUID: "8c54a9e9-d610-4c82-bcb5-a315b9a65010"}},
		},
		&Increment{
			Duration:       time.Second,
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
			Duration:       time.Second,
			Cost:           0.01,
			CompressFactor: 60,
			BalanceInfo: &DebitInfo{
				AccountID: "cgrates.org:dan",
				Monetary: &MonetaryInfo{
					UUID: "8c54a9e9-d610-4c82-bcb5-a315b9a65010"}},
		},
		&Increment{
			Duration:       time.Second,
			Cost:           0.01,
			CompressFactor: 60,
			BalanceInfo: &DebitInfo{
				AccountID: "cgrates.org:dan",
				Monetary: &MonetaryInfo{
					UUID: "8c54a9e9-d610-4c82-bcb5-a315b9a65010"}},
		},
		&Increment{
			Duration:       time.Second,
			Cost:           0.01,
			CompressFactor: 60,
			BalanceInfo: &DebitInfo{
				AccountID: "cgrates.org:dan",
				Monetary: &MonetaryInfo{
					UUID: "8c54a9e9-d610-4c82-bcb5-a315b9a65010"}},
		},
		&Increment{
			Duration:       time.Second,
			Cost:           0.01,
			CompressFactor: 60,
			BalanceInfo: &DebitInfo{
				AccountID: "cgrates.org:dan",
				Monetary: &MonetaryInfo{
					UUID: "8c54a9e9-d610-4c82-bcb5-a315b9a65010"}},
		},
		&Increment{
			Duration:       time.Second,
			Cost:           0,
			CompressFactor: 10,
			BalanceInfo: &DebitInfo{
				AccountID: "cgrates.org:dan",
				Unit: &UnitInfo{
					UUID: "4b8b53d7-c1a1-4159-b845-4623a00a0165"}},
		},
		&Increment{
			Duration:       10 * time.Second,
			Cost:           0.01,
			CompressFactor: 2,
			BalanceInfo: &DebitInfo{
				AccountID: "cgrates.org:dan",
				Monetary: &MonetaryInfo{
					UUID: "8c54a9e9-d610-4c82-bcb5-a315b9a65010"}},
		},
		&Increment{
			Duration:       time.Second,
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
			Duration:       time.Second,
			Cost:           0,
			CompressFactor: 10,
			BalanceInfo: &DebitInfo{
				AccountID: "cgrates.org:dan",
				Unit: &UnitInfo{
					UUID: "4b8b53d7-c1a1-4159-b845-4623a00a0165"}},
		},
		&Increment{
			Duration:       10 * time.Second,
			Cost:           0.01,
			CompressFactor: 2,
			BalanceInfo: &DebitInfo{
				AccountID: "cgrates.org:dan",
				Monetary: &MonetaryInfo{
					UUID: "8c54a9e9-d610-4c82-bcb5-a315b9a65010"}},
		},
		&Increment{
			Duration:       time.Second,
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
			Duration:       time.Second,
			Cost:           0,
			CompressFactor: 10,
			BalanceInfo: &DebitInfo{
				AccountID: "cgrates.org:dan",
				Unit: &UnitInfo{
					UUID: "4b8b53d7-c1a1-4159-b845-4623a00a0165"}},
		},
		&Increment{
			Duration:       10 * time.Second,
			Cost:           0.01,
			CompressFactor: 2,
			BalanceInfo: &DebitInfo{
				AccountID: "cgrates.org:dan",
				Monetary: &MonetaryInfo{
					UUID: "8c54a9e9-d610-4c82-bcb5-a315b9a65010"}},
		},
		&Increment{
			Duration:       time.Second,
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
			Duration:       time.Second,
			Cost:           0,
			CompressFactor: 10,
			BalanceInfo: &DebitInfo{
				AccountID: "cgrates.org:dan",
				Unit: &UnitInfo{
					UUID: "4b8b53d7-c1a1-4159-b845-4623a00a0165"}},
		},
		&Increment{
			Duration:       10 * time.Second,
			Cost:           0.01,
			CompressFactor: 2,
			BalanceInfo: &DebitInfo{
				AccountID: "cgrates.org:dan",
				Monetary: &MonetaryInfo{
					UUID: "8c54a9e9-d610-4c82-bcb5-a315b9a65010"}},
		},
		&Increment{
			Duration:       time.Second,
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
			Duration:       time.Second,
			Cost:           0,
			CompressFactor: 10,
			BalanceInfo: &DebitInfo{
				AccountID: "cgrates.org:dan",
				Unit: &UnitInfo{
					UUID: "4b8b53d7-c1a1-4159-b845-4623a00a0165"}},
		},
		&Increment{
			Duration:       10 * time.Second,
			Cost:           0.01,
			CompressFactor: 2,
			BalanceInfo: &DebitInfo{
				AccountID: "cgrates.org:dan",
				Monetary: &MonetaryInfo{
					UUID: "8c54a9e9-d610-4c82-bcb5-a315b9a65010"}},
		},
		&Increment{
			Duration:       time.Second,
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

	if cd := testEC.Clone().AsRefundIncrements(utils.MetaVoice); !reflect.DeepEqual(eCD, cd) {
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
		Usage:     utils.DurationPointer(2 * time.Minute),
		Charges: []*ChargingInterval{
			{
				RatingID: "f2518464-68b8-42f4-acec-aef23d714314",
				Increments: []*ChargingIncrement{
					{
						Usage:          0,
						Cost:           0.1,
						AccountingID:   "44e97dec-8a7e-43d0-8b0a-736d46b5613e",
						CompressFactor: 1,
					},
					{
						Usage:          time.Second,
						Cost:           0,
						AccountingID:   "a555cde8-4bd0-408a-afbc-c3ba64888927",
						CompressFactor: 30,
					},
					{
						Usage:          time.Second,
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
						Usage:          time.Second,
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
				&RGRate{
					GroupIntervalStart: 0,
					Value:              0.01,
					RateIncrement:      time.Minute,
					RateUnit:           time.Second},
			},
			"e5eb0f1c-3612-4e8c-b749-7f8f41dd90d4": RateGroups{
				&RGRate{
					GroupIntervalStart: 0,
					Value:              0.005,
					RateIncrement:      time.Second,
					RateUnit:           time.Second},
				&RGRate{
					GroupIntervalStart: 60 * time.Second,
					Value:              0.005,
					RateIncrement:      time.Second,
					RateUnit:           time.Second},
			},
		},
		Timings: ChargedTimings{
			"27f1e5f8-05bb-4f1c-a596-bf1010ad296c": &ChargedTiming{
				StartTime: "00:00:00",
			},
		},
	}
	eCC := &CallCost{
		ToR:            utils.MetaVoice,
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
						ID:        "27f1e5f8-05bb-4f1c-a596-bf1010ad296c",
						StartTime: "00:00:00",
					},
					Rating: &RIRate{
						ConnectFee:       0.1,
						RoundingMethod:   "*up",
						RoundingDecimals: 5,
						Rates: RateGroups{
							&RGRate{
								GroupIntervalStart: 0,
								Value:              0.01,
								RateUnit:           time.Second,
								RateIncrement:      time.Minute,
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
						Cost: 0.1,
						BalanceInfo: &DebitInfo{
							Monetary: &MonetaryInfo{
								UUID: "8c54a9e9-d610-4c82-bcb5-a315b9a65010"},
							AccountID: "cgrates.org:dan",
						},
						CompressFactor: 1,
					},
					&Increment{ // First 30 seconds free
						Duration: time.Second,
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
						Duration: time.Second,
						Cost:     0.005,
						BalanceInfo: &DebitInfo{
							Unit: &UnitInfo{ // Minutes with special price
								UUID:     "7a54a9e9-d610-4c82-bcb5-a315b9a65010",
								Consumed: 1.0,
								RateInterval: &RateInterval{
									Timing: &RITiming{
										ID:        "27f1e5f8-05bb-4f1c-a596-bf1010ad296c",
										StartTime: "00:00:00",
									},
									Rating: &RIRate{
										ConnectFee:       0,
										RoundingMethod:   "*up",
										RoundingDecimals: 5,
										Rates: RateGroups{
											&RGRate{
												GroupIntervalStart: 0,
												Value:              0.005,
												RateUnit:           time.Second,
												RateIncrement:      time.Second,
											},
											&RGRate{
												GroupIntervalStart: 60 * time.Second,
												Value:              0.005,
												RateUnit:           time.Second,
												RateIncrement:      time.Second,
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
						ID:        "27f1e5f8-05bb-4f1c-a596-bf1010ad296c",
						StartTime: "00:00:00",
					},
					Rating: &RIRate{
						ConnectFee:       0.1,
						RoundingMethod:   "*up",
						RoundingDecimals: 5,
						Rates: RateGroups{
							&RGRate{
								GroupIntervalStart: 0,
								Value:              0.01,
								RateUnit:           time.Second,
								RateIncrement:      time.Minute,
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
					&Increment{
						Cost:     0.01,
						Duration: time.Second,
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
		t.Errorf("Expecting: %+v,\n received: %+v", utils.ToJSON(eCC), utils.ToJSON(cc))
	}
}

func TestECAsCallCost2(t *testing.T) {
	eCC := &CallCost{
		ToR:        utils.MetaVoice,
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
							&RGRate{
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
	for k := range ec.Timings {
		eCC.Timespans[0].RateInterval.Timing.ID = k
	}
	cc := ec.AsCallCost(utils.EmptyString)
	if !reflect.DeepEqual(eCC, cc) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eCC), utils.ToJSON(cc))
	}
}

func TestECTrimZeroAndFull(t *testing.T) {
	ec := testEC.Clone()
	if srplsEC, err := ec.Trim(10 * time.Minute); err != nil {
		t.Error(err)
	} else if srplsEC != nil {
		t.Errorf("Expecting nil, got: %+v", srplsEC)
	}
	eFullSrpls := testEC.Clone()
	eFullSrpls.Usage = utils.DurationPointer(10 * time.Minute)
	if srplsEC, err := ec.Trim(0); err != nil {
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
					Usage:          0,
					Cost:           0.1,
					AccountingID:   "9bdad10",
					CompressFactor: 1,
				},
				{
					Usage:          time.Second,
					Cost:           0,
					AccountingID:   "3455b83",
					CompressFactor: 10,
				},
				{
					Usage:          10 * time.Second,
					Cost:           0.01,
					AccountingID:   "a012888",
					CompressFactor: 2,
				},
				{
					Usage:          time.Second,
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
					Usage:          time.Second,
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
					Usage:          time.Second,
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
					Usage:          time.Second,
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
					Usage:          time.Second,
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
					Usage:          time.Second,
					Cost:           0,
					AccountingID:   "3455b83",
					CompressFactor: 10,
				},
				{
					Usage:          10 * time.Second,
					Cost:           0.01,
					AccountingID:   "a012888",
					CompressFactor: 2,
				},
				{
					Usage:          time.Second,
					Cost:           0.005,
					AccountingID:   "44d6c02",
					CompressFactor: 30,
				},
			},
			CompressFactor: 5,
		},
	}

	reqDuration := 190 * time.Second
	srplsEC, err := ec.Trim(reqDuration)
	if err != nil {
		t.Error(err)
	}
	if reqDuration != *ec.Usage {
		t.Errorf("Expecting request duration: %v, received: %v", reqDuration, *ec.Usage)
	}
	eSrplsDur := 410 * time.Second
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
		{5 * time.Second, 5 * time.Second, 0.1,
			595 * time.Second, 3.42},
		{10 * time.Second, 10 * time.Second, 0.1,
			590 * time.Second, 3.42},
		{15 * time.Second, 20 * time.Second, 0.11,
			580 * time.Second, 3.41},
		{20 * time.Second, 20 * time.Second, 0.11,
			580 * time.Second, 3.41},
		{25 * time.Second, 30 * time.Second, 0.12,
			570 * time.Second, 3.40},
		{38 * time.Second, 38 * time.Second, 0.16,
			562 * time.Second, 3.36},
		{60 * time.Second, 60 * time.Second, 0.27,
			540 * time.Second, 3.25},
		{62 * time.Second, 62 * time.Second, 0.29,
			538 * time.Second, 3.23},
		{120 * time.Second, 120 * time.Second, 0.87,
			480 * time.Second, 2.65},
		{121 * time.Second, 121 * time.Second, 0.88,
			479 * time.Second, 2.64},
		{180 * time.Second, 180 * time.Second, 1.47,
			420 * time.Second, 2.05},
		{250 * time.Second, 250 * time.Second, 2.17,
			350 * time.Second, 1.35},
		{299 * time.Second, 299 * time.Second, 2.66,
			301 * time.Second, 0.86},
		{300 * time.Second, 300 * time.Second, 2.67,
			300 * time.Second, 0.85},
		{302 * time.Second, 302 * time.Second, 2.67,
			298 * time.Second, 0.85},
		{310 * time.Second, 310 * time.Second, 2.67,
			290 * time.Second, 0.85},
		{316 * time.Second, 320 * time.Second, 2.68,
			280 * time.Second, 0.84},
		{320 * time.Second, 320 * time.Second, 2.68,
			280 * time.Second, 0.84},
		{321 * time.Second, 330 * time.Second, 2.69,
			270 * time.Second, 0.83},
		{330 * time.Second, 330 * time.Second, 2.69,
			270 * time.Second, 0.83},
		{331 * time.Second, 331 * time.Second, 2.695,
			269 * time.Second, 0.825},
		{359 * time.Second, 359 * time.Second, 2.835,
			241 * time.Second, 0.685},
		{360 * time.Second, 360 * time.Second, 2.84,
			240 * time.Second, 0.68},
		{376 * time.Second, 380 * time.Second, 2.85,
			220 * time.Second, 0.67},
		{391 * time.Second, 391 * time.Second, 2.865,
			209 * time.Second, 0.655},
		{479 * time.Second, 479 * time.Second, 3.175,
			121 * time.Second, 0.345},
		{599 * time.Second, 599 * time.Second, 3.515,
			time.Second, 0.005},
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
						Usage:          102400,
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
					Type:  utils.MetaData,
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
				ExtraChargeID: utils.MetaNone,
			},
		},
		RatingFilters: RatingFilters{
			"216b0a5": RatingMatchedFilters{
				"DestinationID":     utils.MetaAny,
				"DestinationPrefix": "42502",
				"RatingPlanID":      utils.MetaNone,
				"Subject":           "9a767726-fe69-4940-b7bd-f43de9f0f8a5",
			},
		},
		Rates: ChargedRates{
			"06dee2e": RateGroups{
				&RGRate{
					RateIncrement: 102400,
					RateUnit:      102400},
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
						Usage:          102400,
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
					Type:  utils.MetaData,
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
				ExtraChargeID: utils.MetaNone,
			},
		},
		RatingFilters: RatingFilters{
			"17f7216": RatingMatchedFilters{
				"DestinationID":     utils.MetaAny,
				"DestinationPrefix": "42502",
				"RatingPlanID":      utils.MetaNone,
				"Subject":           "9a767726-fe69-4940-b7bd-f43de9f0f8a5",
			},
		},
		Rates: ChargedRates{
			"52f8b0f": RateGroups{
				&RGRate{
					RateIncrement: 102400,
					RateUnit:      102400},
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
						Usage:          102400,
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
					Type:  utils.MetaData,
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
				ExtraChargeID: utils.MetaNone,
			},
		},
		RatingFilters: RatingFilters{
			"216b0a5": RatingMatchedFilters{
				"DestinationID":     utils.MetaAny,
				"DestinationPrefix": "42502",
				"RatingPlanID":      utils.MetaNone,
				"Subject":           "9a767726-fe69-4940-b7bd-f43de9f0f8a5",
			},
		},
		Rates: ChargedRates{
			"06dee2e": RateGroups{
				&RGRate{
					RateIncrement: 102400,
					RateUnit:      102400},
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
						Usage:          102400,
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
				ExtraChargeID: utils.MetaNone,
			},
		},
		RatingFilters: RatingFilters{
			"216b0a5": RatingMatchedFilters{
				"DestinationID":     utils.MetaAny,
				"DestinationPrefix": "42502",
				"RatingPlanID":      utils.MetaNone,
				"Subject":           "9a767726-fe69-4940-b7bd-f43de9f0f8a5",
			},
		},
		Rates: ChargedRates{
			"06dee2e": RateGroups{
				&RGRate{
					RateIncrement: 102400,
					RateUnit:      102400},
			},
		},
	}
	oEC := &EventCost{
		Charges: []*ChargingInterval{
			{
				RatingID: "6a83227",
				Increments: []*ChargingIncrement{
					{
						Usage:          102400,
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
				ExtraChargeID: utils.MetaNone,
			},
		},
		RatingFilters: RatingFilters{
			"17f7216": RatingMatchedFilters{
				"DestinationID":     utils.MetaAny,
				"DestinationPrefix": "42502",
				"RatingPlanID":      utils.MetaNone,
				"Subject":           "9a767726-fe69-4940-b7bd-f43de9f0f8a5",
			},
		},
		Rates: ChargedRates{
			"52f8b0f": RateGroups{
				&RGRate{
					RateIncrement: 102400,
					RateUnit:      102400},
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
						Usage:          102400,
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
				ExtraChargeID: utils.MetaNone,
			},
		},
		RatingFilters: RatingFilters{
			"216b0a5": RatingMatchedFilters{
				"DestinationID":     utils.MetaAny,
				"DestinationPrefix": "42502",
				"RatingPlanID":      utils.MetaNone,
				"Subject":           "9a767726-fe69-4940-b7bd-f43de9f0f8a5",
			},
		},
		Rates: ChargedRates{
			"06dee2e": RateGroups{
				&RGRate{
					RateIncrement: 102400,
					RateUnit:      102400},
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
						Usage:          102400,
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
				ExtraChargeID: utils.MetaNone,
			},
		},
		RatingFilters: RatingFilters{
			"216b0a5": RatingMatchedFilters{
				"DestinationID":     utils.MetaAny,
				"DestinationPrefix": "42502",
				"RatingPlanID":      utils.MetaNone,
				"Subject":           "9a767726-fe69-4940-b7bd-f43de9f0f8a5",
			},
		},
		Rates: ChargedRates{
			"06dee2e": RateGroups{
				&RGRate{
					RateIncrement: 102400,
					RateUnit:      102400},
			},
		},
	}
	oEC = &EventCost{
		Charges: []*ChargingInterval{
			{
				RatingID: "6a83227",
				Increments: []*ChargingIncrement{
					{
						Usage:          102400,
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
				ExtraChargeID: utils.MetaNone,
			},
		},
		RatingFilters: RatingFilters{
			"17f7216": RatingMatchedFilters{
				"DestinationID":     utils.MetaAny,
				"DestinationPrefix": "42502",
				"RatingPlanID":      utils.MetaNone,
				"Subject":           "9a767726-fe69-4940-b7bd-f43de9f0f8a5",
			},
		},
		Rates: ChargedRates{
			"52f8b0f": RateGroups{
				&RGRate{
					RateIncrement: 102400,
					RateUnit:      102400},
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
						Usage:          102400,
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
						Usage:          102400,
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
				ExtraChargeID: utils.MetaNone,
			},
		},
		RatingFilters: RatingFilters{
			"216b0a5": RatingMatchedFilters{
				"DestinationID":     utils.MetaAny,
				"DestinationPrefix": "42502",
				"RatingPlanID":      utils.MetaNone,
				"Subject":           "9a767726-fe69-4940-b7bd-f43de9f0f8a5",
			},
		},
		Rates: ChargedRates{
			"06dee2e": RateGroups{
				&RGRate{
					RateIncrement: 102400,
					RateUnit:      102400},
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
						Usage:          100,
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
						Usage:          102400,
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
				ExtraChargeID: utils.MetaNone,
			},
		},
		RatingFilters: RatingFilters{
			"216b0a5": RatingMatchedFilters{
				"DestinationID":     utils.MetaAny,
				"DestinationPrefix": "42502",
				"RatingPlanID":      utils.MetaNone,
				"Subject":           "9a767726-fe69-4940-b7bd-f43de9f0f8a5",
			},
		},
		Rates: ChargedRates{
			"06dee2e": RateGroups{
				&RGRate{
					RateIncrement: 102400,
					RateUnit:      102400},
			},
		},
	}
	oEC = &EventCost{
		Charges: []*ChargingInterval{
			{
				RatingID: "6a83227",
				Increments: []*ChargingIncrement{
					{
						Usage:          102400,
						AccountingID:   "9288f93",
						CompressFactor: 42,
					},
					{
						Usage:          10240,
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
				ExtraChargeID: utils.MetaNone,
			},
		},
		RatingFilters: RatingFilters{
			"17f7216": RatingMatchedFilters{
				"DestinationID":     utils.MetaAny,
				"DestinationPrefix": "42502",
				"RatingPlanID":      utils.MetaNone,
				"Subject":           "9a767726-fe69-4940-b7bd-f43de9f0f8a5",
			},
		},
		Rates: ChargedRates{
			"52f8b0f": RateGroups{
				&RGRate{
					RateIncrement: 102400,
					RateUnit:      102400},
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
						Usage:          100,
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
						Usage:          102400,
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
						Usage:          10240,
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
						Usage:          102400,
						AccountingID:   "0d87a64",
						CompressFactor: 42,
					},
					{
						Usage:          10240,
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
				ExtraChargeID: utils.MetaNone,
			},
		},
		RatingFilters: RatingFilters{
			"216b0a5": RatingMatchedFilters{
				"DestinationID":     utils.MetaAny,
				"DestinationPrefix": "42502",
				"RatingPlanID":      utils.MetaNone,
				"Subject":           "9a767726-fe69-4940-b7bd-f43de9f0f8a5",
			},
		},
		Rates: ChargedRates{
			"06dee2e": RateGroups{
				&RGRate{
					RateIncrement: 102400,
					RateUnit:      102400},
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
						Usage:          102400,
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
				ExtraChargeID: utils.MetaNone,
			},
		},
		RatingFilters: RatingFilters{
			"216b0a5": RatingMatchedFilters{
				"DestinationID":     utils.MetaAny,
				"DestinationPrefix": "42502",
				"RatingPlanID":      utils.MetaNone,
				"Subject":           "9a767726-fe69-4940-b7bd-f43de9f0f8a5",
			},
		},
		Rates: ChargedRates{
			"06dee2e": RateGroups{
				&RGRate{
					RateIncrement: 102400,
					RateUnit:      102400},
			},
		},
	}
	oEC = &EventCost{
		Charges: []*ChargingInterval{
			{
				RatingID: "6a83227",
				Increments: []*ChargingIncrement{
					{
						Usage:          102400,
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
				ExtraChargeID: utils.MetaNone,
			},
		},
		RatingFilters: RatingFilters{
			"17f7216": RatingMatchedFilters{
				"DestinationID":     utils.MetaAny,
				"DestinationPrefix": "42502",
				"RatingPlanID":      utils.MetaNone,
				"Subject":           "9a767726-fe69-4940-b7bd-f43de9f0f8a5",
			},
		},
		Rates: ChargedRates{
			"52f8b0f": RateGroups{
				&RGRate{
					RateIncrement: 102400,
					RateUnit:      102400},
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
						Usage:          102400,
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
				ExtraChargeID: utils.MetaNone,
			},
		},
		RatingFilters: RatingFilters{
			"216b0a5": RatingMatchedFilters{
				"DestinationID":     utils.MetaAny,
				"DestinationPrefix": "42502",
				"RatingPlanID":      utils.MetaNone,
				"Subject":           "9a767726-fe69-4940-b7bd-f43de9f0f8a5",
			},
		},
		Rates: ChargedRates{
			"06dee2e": RateGroups{
				&RGRate{
					RateIncrement: 102400,
					RateUnit:      102400},
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
				&RGRate{
					GroupIntervalStart: 0,
					Value:              0.01,
					RateIncrement:      time.Minute,
					RateUnit:           time.Second},
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
						Usage:          0,
						Cost:           0.1,
						AccountingID:   "9bdad10",
						CompressFactor: 1,
					},
					{
						Usage:          time.Second,
						Cost:           0,
						AccountingID:   "3455b83",
						CompressFactor: 10,
					},
					{
						Usage:          10 * time.Second,
						Cost:           0.01,
						AccountingID:   "2012888",
						CompressFactor: 2,
					},
					{
						Usage:          time.Second,
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
						Usage:          time.Second,
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
						Usage:          time.Second,
						Cost:           0,
						AccountingID:   "3455b83",
						CompressFactor: 10,
					},
					{
						Usage:          10 * time.Second,
						Cost:           0.01,
						AccountingID:   "2012888",
						CompressFactor: 2,
					},
					{
						Usage:          time.Second,
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
					Type:     utils.MetaMonetary,
					Value:    50,
					Disabled: false},
				{
					UUID:     "7a54a9e9-d610-4c82-bcb5-a315b9a65010",
					Type:     utils.MetaMonetary,
					Value:    25,
					Disabled: false},
				{
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
				&RGRate{
					GroupIntervalStart: 0,
					Value:              0.01,
					RateIncrement:      time.Minute,
					RateUnit:           time.Second},
			},
			"4910ecf": RateGroups{
				&RGRate{
					GroupIntervalStart: 0,
					Value:              0.005,
					RateIncrement:      time.Second,
					RateUnit:           time.Second},
				&RGRate{
					GroupIntervalStart: 60 * time.Second,
					Value:              0.005,
					RateIncrement:      time.Second,
					RateUnit:           time.Second},
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
			{
				RatingID: "RatingID1",
			},
			{
				RatingID: "RatingID2",
			},
		},
	}
	expectedCharges := []*ChargingInterval{
		{
			RatingID: "RatingID1",
		},
		{
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
		Usage: utils.DurationPointer(5 * time.Minute),
	}
	if rcv, err := eventCost.fieldAsInterface([]string{utils.Usage}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(5*time.Minute, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", 5*time.Minute, rcv)
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
	} else if !reflect.DeepEqual(0.7, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", 0.7, rcv)
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
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eChargedTiming), utils.ToJSON(rcv))
	}
	// case utils.Rates:
	eventCost = &EventCost{
		Rates: ChargedRates{
			"test1": RateGroups{
				&RGRate{Value: 0.7},
			},
		},
	}
	eChargedRates := ChargedRates{
		"test1": RateGroups{
			&RGRate{Value: 0.7},
		},
	}
	if rcv, err := eventCost.fieldAsInterface([]string{utils.Rates}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eChargedRates, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eChargedRates), utils.ToJSON(rcv))
	}
	eRateGroup := RateGroups{
		&RGRate{Value: 0.7},
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
		t.Errorf("Expecting: %+v, \nreceived: %+v", utils.ToJSON(eBalanceCharge), utils.ToJSON(rcv))
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
			{
				RatingID: "RatingID",
			},
		},
	}
	eCharges := []*ChargingInterval{{RatingID: "RatingID"}}
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
			"RatingID": RateGroups{&RGRate{Value: 0.8}}},
	}
	if rcv, err := eventCost.getChargesForPath([]string{utils.Rating}, chargingInterval); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expecting: %+v, received: %+v", utils.ErrNotFound, err)
	} else if rcv != nil {
		t.Errorf("Expecting: nil, received: %+v", rcv)
	}
	eventCost = &EventCost{
		Rating: Rating{"RatingID": &RatingUnit{}},
		Rates: ChargedRates{
			"RatingID": RateGroups{&RGRate{Value: 0.8}}},
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
		Rates: ChargedRates{"RatesID": RateGroups{&RGRate{Value: 0.8}}},
	}
	RateGroups := RateGroups{&RGRate{Value: 0.8}}
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
			{
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
		{
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
				&RGRate{Value: 0.7},
			},
		},
	}
	eChargedRates := RateGroups{
		&RGRate{Value: 0.7},
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
				&RGRate{Value: 0.7},
			},
			"RatesID2": RateGroups{
				&RGRate{Value: 0.7},
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
	eRate := &RGRate{Value: 0.7}
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
				{
					ID: "ID",
				},
			},
		},
	}
	eBalanceSummaries := &BalanceSummary{
		ID: "ID",
	}
	//len(fldPath) == 1
	if rcv, err := eventCost.getAcountingForPath([]string{utils.BalanceField}, balanceCharge); err != nil {
		t.Errorf("Expecting: nil, received: %+v", err)
	} else if !reflect.DeepEqual(eBalanceSummaries, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", eBalanceSummaries, rcv)
	}
	// bl == nil
	eventCost = &EventCost{AccountSummary: &AccountSummary{}}
	if rcv, err := eventCost.getAcountingForPath([]string{utils.BalanceField}, balanceCharge); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expecting: %+v, received: %+v", utils.ErrNotFound, err)
	} else if rcv != nil {
		t.Errorf("Expecting: nil, received: %+v", rcv)
	}
	// len(fldPath) != 1
	eventCost = &EventCost{
		AccountSummary: &AccountSummary{
			BalanceSummaries: BalanceSummaries{
				{
					ID: "ID",
				},
			},
		},
	}
	if rcv, err := eventCost.getAcountingForPath([]string{utils.BalanceField, "ID"}, balanceCharge); err != nil {
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
	eOut = `{"CGRID":"","RunID":"","StartTime":"0001-01-01T00:00:00Z","Usage":null,"Cost":null,"Charges":null,"AccountSummary":{"Tenant":"","ID":"","BalanceSummaries":[{"UUID":"","ID":"ID","Type":"","Initial":0,"Value":0,"Disabled":false}],"AllowNegative":false,"Disabled":false},"Rating":null,"Accounting":null,"RatingFilters":null,"Rates":null,"Timings":null}`
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

func TestECAsCallCost3(t *testing.T) {
	eCC := &CallCost{
		Category:    "call",
		Tenant:      "cgrates.org",
		Subject:     "dan",
		Account:     "dan",
		Destination: "+4986517174963",
		ToR:         utils.MetaVoice,
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
							&RGRate{
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
								ToR:      utils.MetaVoice,
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
								ToR:      utils.MetaVoice,
								RateInterval: &RateInterval{
									Timing: &RITiming{
										StartTime: "00:00:00",
									},
									Rating: &RIRate{
										ConnectFee:       0,
										RoundingMethod:   "*up",
										RoundingDecimals: 5,
										Rates: RateGroups{
											&RGRate{
												GroupIntervalStart: time.Duration(0),
												Value:              0.005,
												RateUnit:           time.Duration(1 * time.Second),
												RateIncrement:      time.Duration(1 * time.Second),
											},
											&RGRate{
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
							&RGRate{
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
					Type:     utils.MetaMonetary,
					Value:    50,
					Disabled: false},
				{
					ID:       "4b8b53d7-c1a1-4159-b845-4623a00a0165",
					Type:     utils.MetaMonetary,
					Value:    25,
					Disabled: false},
				{
					Type:     utils.MetaVoice,
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
					Type:     utils.MetaMonetary,
					Value:    50,
					Disabled: false},
				{
					UUID:     "7a54a9e9-d610-4c82-bcb5-a315b9a65010",
					Type:     utils.MetaMonetary,
					Value:    25,
					Disabled: false},
				{
					UUID:     "4b8b53d7-c1a1-4159-b845-4623a00a0165",
					Type:     utils.MetaVoice,
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
				&RGRate{
					GroupIntervalStart: time.Duration(0),
					Value:              0.01,
					RateIncrement:      time.Duration(1 * time.Minute),
					RateUnit:           time.Duration(1 * time.Second)},
			},
			"4910ecf": RateGroups{
				&RGRate{
					GroupIntervalStart: time.Duration(0),
					Value:              0.005,
					RateIncrement:      time.Duration(1 * time.Second),
					RateUnit:           time.Duration(1 * time.Second)},
				&RGRate{
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

func TestECFieldAsInterfaceNilEventCost(t *testing.T) {
	dft := config.NewDefaultCGRConfig()
	cdr, err := NewMapEvent(map[string]interface{}{}).AsCDR(dft, "cgrates.org", "UTC")
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

func TestECnewChargingIncrementMissingBalanceInfo(t *testing.T) {
	ec := &EventCost{}
	incr := &Increment{}
	rf := make(RatingMatchedFilters)

	exp := &ChargingIncrement{
		Usage:          incr.Duration,
		Cost:           incr.Cost,
		CompressFactor: incr.CompressFactor,
	}
	rcv := ec.newChargingIncrement(incr, rf, false, false)

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", exp, rcv)
	}
}

func TestECnewChargingIncrementWithUnitInfo(t *testing.T) {
	ec := &EventCost{
		Accounting: Accounting{},
	}
	incr := &Increment{
		BalanceInfo: &DebitInfo{
			Unit:     &UnitInfo{},
			Monetary: &MonetaryInfo{},
		},
	}
	rf := make(RatingMatchedFilters)

	exp := &ChargingIncrement{
		Usage:          incr.Duration,
		Cost:           incr.Cost,
		CompressFactor: incr.CompressFactor,
		AccountingID:   utils.MetaPause,
	}
	rcv := ec.newChargingIncrement(incr, rf, false, true)

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", exp, rcv)
	}
}

func TestECnewChargingIncrementNoUnitInfo(t *testing.T) {
	ec := &EventCost{
		Accounting: Accounting{},
	}
	incr := &Increment{
		BalanceInfo: &DebitInfo{
			Monetary: &MonetaryInfo{},
		},
	}
	rf := make(RatingMatchedFilters)

	exp := &ChargingIncrement{
		Usage:          incr.Duration,
		Cost:           incr.Cost,
		CompressFactor: incr.CompressFactor,
		AccountingID:   utils.MetaPause,
	}
	expEC := &EventCost{
		Accounting: Accounting{
			utils.MetaPause: &BalanceCharge{},
		},
	}
	rcv := ec.newChargingIncrement(incr, rf, false, true)

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", exp, rcv)
	}

	if !reflect.DeepEqual(ec, expEC) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", expEC, ec)
	}
}

func TestECAsRefundIncrementsNoCharges(t *testing.T) {
	ec := &EventCost{
		Charges:   []*ChargingInterval{},
		CGRID:     "asdfgh",
		RunID:     "runID",
		StartTime: time.Date(2021, 4, 13, 17, 0, 0, 0, time.UTC),
		Usage:     utils.DurationPointer(time.Hour),
		Cost:      utils.Float64Pointer(10),
	}

	exp := &CallDescriptor{
		CgrID:         "asdfgh",
		RunID:         "runID",
		ToR:           utils.MetaVoice,
		TimeStart:     time.Date(2021, 4, 13, 17, 0, 0, 0, time.UTC),
		TimeEnd:       time.Date(2021, 4, 13, 18, 0, 0, 0, time.UTC),
		DurationIndex: time.Hour,
	}

	rcv := ec.AsRefundIncrements(utils.MetaVoice)

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", exp, rcv)
	}

}

func TestECAsRefundIncrements2(t *testing.T) {
	ec := &EventCost{
		Charges: []*ChargingInterval{
			{
				cost:           utils.Float64Pointer(10),
				CompressFactor: 1,
				Increments: []*ChargingIncrement{
					{
						Cost:           5,
						Usage:          2 * time.Second,
						AccountingID:   "accID",
						CompressFactor: 1,
					},
				},
			},
		},
		CGRID:     "asdfgh",
		RunID:     "runID",
		StartTime: time.Date(2021, 4, 13, 17, 0, 0, 0, time.UTC),
		Usage:     utils.DurationPointer(time.Hour),
		Cost:      utils.Float64Pointer(10),
		Accounting: Accounting{
			"accID": &BalanceCharge{
				AccountID:     "bcAccID",
				BalanceUUID:   "bcUUID",
				ExtraChargeID: "extrachargeID",
			},
			"extrachargeID": &BalanceCharge{
				AccountID:   "extraAccID",
				BalanceUUID: "extraBcUUID",
			},
		},
		AccountSummary: &AccountSummary{
			BalanceSummaries: BalanceSummaries{
				{
					UUID: "bcUUID",
					Type: utils.MetaSMS,
				},
				{
					UUID: "extraBcUUID",
					Type: utils.MetaSMS,
				},
			},
		},
	}

	exp := &CallDescriptor{
		CgrID:         "asdfgh",
		RunID:         "runID",
		ToR:           utils.MetaVoice,
		TimeStart:     time.Date(2021, 4, 13, 17, 0, 0, 0, time.UTC),
		TimeEnd:       time.Date(2021, 4, 13, 18, 0, 0, 0, time.UTC),
		DurationIndex: time.Hour,
		Increments: Increments{
			{
				Duration: 2 * time.Second,
				Cost:     5,
				BalanceInfo: &DebitInfo{
					Unit: &UnitInfo{
						UUID: "extraBcUUID",
					},
					AccountID: "bcAccID",
				},
				CompressFactor: 1,
			},
		},
	}
	rcv := ec.AsRefundIncrements(utils.MetaVoice)

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", exp, rcv)
	}
}

func TestECratingGetIDFromEventCostPause(t *testing.T) {
	ec := &EventCost{
		Timings: ChargedTimings{
			utils.MetaPause: &ChargedTiming{},
		},
		RatingFilters: RatingFilters{
			utils.MetaPause: RatingMatchedFilters{},
		},
		Rates: ChargedRates{
			utils.MetaPause: RateGroups{},
		},
		Rating: Rating{
			utils.MetaPause: &RatingUnit{},
		},
	}
	oEC := &EventCost{
		Rating: Rating{
			utils.MetaPause: &RatingUnit{
				ConnectFee:       0.4,
				RoundingMethod:   "*up",
				RoundingDecimals: 4,
				MaxCost:          100,
				MaxCostStrategy:  "*disconnect",
				TimingID:         "TM_NOON",
			},
		},
		Timings: ChargedTimings{
			utils.MetaPause: &ChargedTiming{
				Years:     utils.Years{2010, 2011},
				Months:    utils.Months{1, 2},
				MonthDays: utils.MonthDays{24, 25},
				WeekDays:  utils.WeekDays{2},
				StartTime: "00:00:00",
			},
		},
	}
	oRatingID := utils.MetaPause

	exp := utils.MetaPause
	rcv := ec.ratingGetIDFromEventCost(oEC, oRatingID)

	if rcv != exp {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", exp, rcv)
	}
}

func TestECaccountingGetIDFromEventCostPause(t *testing.T) {
	ec := &EventCost{
		Timings: ChargedTimings{
			utils.MetaPause: &ChargedTiming{},
		},
		RatingFilters: RatingFilters{
			utils.MetaPause: RatingMatchedFilters{},
		},
		Rates: ChargedRates{
			utils.MetaPause: RateGroups{},
		},
		Rating: Rating{
			utils.MetaPause: &RatingUnit{},
		},
		Accounting: Accounting{
			utils.MetaPause: &BalanceCharge{},
		},
	}
	oEC := &EventCost{
		Accounting: Accounting{
			utils.MetaPause: &BalanceCharge{
				AccountID:     "1001",
				BalanceUUID:   "asdfg",
				RatingID:      utils.MetaPause,
				Units:         10,
				ExtraChargeID: "extra",
			},
		},
		Rating: Rating{
			utils.MetaPause: &RatingUnit{
				ConnectFee:       0.4,
				RoundingMethod:   "*up",
				RoundingDecimals: 4,
				MaxCost:          100,
				MaxCostStrategy:  "*disconnect",
				TimingID:         "TM_NOON",
			},
		},
	}
	oAccountingID := utils.MetaPause

	exp := utils.MetaPause
	rcv := ec.accountingGetIDFromEventCost(oEC, oAccountingID)

	if rcv != exp {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", exp, rcv)
	}
}

func TestECappendChargingIntervalFromEventCost(t *testing.T) {
	ec := &EventCost{
		Charges: []*ChargingInterval{
			{
				RatingID:       "RT_ID",
				CompressFactor: 1,
			},
		},
	}
	oEC := &EventCost{
		Charges: []*ChargingInterval{
			{},
			{
				RatingID: "RT_ID",
			},
		},
	}

	cIlIdx := 1

	exp := &EventCost{
		Charges: []*ChargingInterval{
			{
				RatingID:       "RT_ID",
				CompressFactor: 2,
			},
		},
	}
	ec.appendChargingIntervalFromEventCost(oEC, cIlIdx)

	if !reflect.DeepEqual(ec, exp) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", exp, ec)
	}
}

func TestECratingIDForRateIntervalPause(t *testing.T) {
	ec := &EventCost{
		RatingFilters: RatingFilters{},
		Rating:        Rating{},
		Rates:         ChargedRates{},
	}
	ri := &RateInterval{
		Rating: &RIRate{
			ConnectFee:       0.4,
			RoundingMethod:   "*up",
			MaxCost:          100,
			MaxCostStrategy:  "*disconnect",
			RoundingDecimals: 4,
			Rates: RateGroups{
				{
					RateIncrement:      60,
					RateUnit:           60,
					Value:              10,
					GroupIntervalStart: 1,
				},
			},
		},
	}
	rf := RatingMatchedFilters{
		"key": "filter",
	}

	exp := utils.MetaPause
	expEC := &EventCost{
		Rating: Rating{
			utils.MetaPause: &RatingUnit{
				ConnectFee:       0.4,
				RoundingMethod:   "*up",
				MaxCost:          100,
				MaxCostStrategy:  "*disconnect",
				RoundingDecimals: 4,
				RatesID:          utils.MetaPause,
				RatingFiltersID:  utils.MetaPause,
			},
		},
		RatingFilters: RatingFilters{
			utils.MetaPause: RatingMatchedFilters{
				"key": "filter",
			},
		},
		Rates: ChargedRates{
			utils.MetaPause: RateGroups{
				{
					RateIncrement:      60,
					RateUnit:           60,
					Value:              10,
					GroupIntervalStart: 1,
				},
			},
		},
	}
	rcv := ec.ratingIDForRateInterval(ri, rf, true)

	if rcv != exp {
		t.Fatalf("\nexpected: <%+v>, \nreceived: <%+v>", exp, rcv)
	}

	if !reflect.DeepEqual(ec, expEC) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", expEC, ec)
	}

}

func TestECAsCallCost4(t *testing.T) {
	ec := &EventCost{
		Charges: []*ChargingInterval{
			{
				Increments: []*ChargingIncrement{
					{
						Usage:          100,
						Cost:           10,
						AccountingID:   "accID1",
						CompressFactor: 1,
					},
					{
						Usage:          150,
						Cost:           15,
						AccountingID:   "accID2",
						CompressFactor: 1,
					},
				},
			},
		},
		Accounting: Accounting{
			"accID1": &BalanceCharge{
				RatingID:    utils.MetaRounding,
				BalanceUUID: "asdfgh1",
			},
			"accID2": &BalanceCharge{
				RatingID:    utils.MetaRounding,
				BalanceUUID: "asdfgh2",
			},
		},
		Rating: Rating{
			utils.MetaRounding: &RatingUnit{
				ConnectFee:       0.4,
				RoundingMethod:   "*up",
				RoundingDecimals: 4,
				MaxCost:          100,
				MaxCostStrategy:  "*disconnect",
				RatesID:          "RT_ID",
				TimingID:         "TM_NOON",
			},
		},
		Timings: ChargedTimings{
			"TM_NOON": &ChargedTiming{
				StartTime: "00:00:00",
			},
		},
	}
	tor := utils.MetaVoice

	exp := &CallCost{
		ToR: utils.MetaVoice,
		Timespans: TimeSpans{
			{
				Cost:          25,
				DurationIndex: 250,
				Increments: Increments{
					{
						Cost:     10,
						Duration: 100,
						BalanceInfo: &DebitInfo{
							Monetary: &MonetaryInfo{
								UUID: "asdfgh1",
								RateInterval: &RateInterval{
									Timing: &RITiming{
										ID:        "TM_NOON",
										StartTime: "00:00:00",
									},
									Rating: &RIRate{
										ConnectFee:       0.4,
										RoundingMethod:   "*up",
										RoundingDecimals: 4,
										MaxCost:          100,
										MaxCostStrategy:  "*disconnect",
									},
								},
							},
						},
						CompressFactor: 1,
					},
				},
				RoundIncrement: &Increment{
					BalanceInfo: &DebitInfo{
						Monetary: &MonetaryInfo{
							UUID: "asdfgh1",
							RateInterval: &RateInterval{
								Timing: &RITiming{
									ID:        "TM_NOON",
									StartTime: "00:00:00",
								},
								Rating: &RIRate{
									ConnectFee:       0.4,
									RoundingMethod:   "*up",
									RoundingDecimals: 4,
									MaxCost:          100,
									MaxCostStrategy:  "*disconnect",
								},
							},
						},
					},
					Cost:           -10,
					Duration:       100,
					CompressFactor: 1,
				},
			},
		},
	}
	rcv := ec.AsCallCost(tor)

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}

func TestECnewIntervalFromCharge(t *testing.T) {
	ec := &EventCost{
		Charges: []*ChargingInterval{
			{
				Increments: []*ChargingIncrement{
					{
						Usage:          100,
						Cost:           10,
						AccountingID:   "accID1",
						CompressFactor: 1,
					},
					{
						Usage:          150,
						Cost:           15,
						AccountingID:   "accID2",
						CompressFactor: 1,
					},
				},
			},
		},
		Accounting: Accounting{
			"accID1": &BalanceCharge{
				RatingID:    utils.MetaRounding,
				BalanceUUID: "asdfgh1",
				AccountID:   "1001",
			},
			"accID2": &BalanceCharge{
				RatingID:    utils.MetaRounding,
				BalanceUUID: "asdfgh2",
				AccountID:   "1001",
			},
		},
		Rating: Rating{
			utils.MetaRounding: &RatingUnit{
				ConnectFee:       0.4,
				RoundingMethod:   "*up",
				RoundingDecimals: 4,
				MaxCost:          100,
				MaxCostStrategy:  "*disconnect",
				RatesID:          "RT_ID",
				TimingID:         "TM_NOON",
			},
		},
		Timings: ChargedTimings{
			"TM_NOON": &ChargedTiming{
				StartTime: "00:00:00",
			},
		},
		AccountSummary: &AccountSummary{
			Tenant:        "cgrates.org",
			ID:            "1001",
			AllowNegative: true,
			BalanceSummaries: BalanceSummaries{
				&BalanceSummary{
					UUID:     "asdfgh1",
					ID:       "bsID",
					Type:     utils.MetaData,
					Initial:  0,
					Value:    10,
					Disabled: true,
				},
			},
		},
	}
	cInc := &ChargingIncrement{
		AccountingID: "accID1",
	}

	exp := &Increment{
		BalanceInfo: &DebitInfo{
			AccountID: "1001",
			Unit: &UnitInfo{
				UUID: "asdfgh1",
				RateInterval: &RateInterval{
					Rating: &RIRate{
						RoundingDecimals: 4,
						RoundingMethod:   "*up",
						ConnectFee:       0.4,
						MaxCost:          100,
						MaxCostStrategy:  "*disconnect",
					},
					Timing: &RITiming{
						ID:        "TM_NOON",
						StartTime: "00:00:00",
					},
				},
			},
		},
	}
	rcv := ec.newIntervalFromCharge(cInc)

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}

func TestECRemoveStaleReferences(t *testing.T) {
	ec := &EventCost{
		Rating: Rating{
			"unusedKey1": &RatingUnit{
				ConnectFee:       0.4,
				RoundingMethod:   "*up",
				RoundingDecimals: 4,
				MaxCost:          100,
				MaxCostStrategy:  "*disconnect",
				TimingID:         "TM_MORNING",
				RatesID:          "RT_ID",
				RatingFiltersID:  "RF_ID",
			},
		},
		RatingFilters: RatingFilters{
			"unusedKey2": RatingMatchedFilters{},
		},
		Rates: ChargedRates{
			"unusedKey3": RateGroups{
				{
					RateIncrement: 60,
					RateUnit:      60,
					Value:         10,
				},
			},
		},
		Timings: ChargedTimings{
			"unusedKey4": &ChargedTiming{
				StartTime: "00:00:00",
			},
		},
	}

	exp := &EventCost{
		Rating:        Rating{},
		RatingFilters: RatingFilters{},
		Rates:         ChargedRates{},
		Timings: ChargedTimings{
			"unusedKey4": &ChargedTiming{
				StartTime: "00:00:00",
			},
		},
	}
	ec.RemoveStaleReferences()

	if !reflect.DeepEqual(ec, exp) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ToJSON(exp), utils.ToJSON(ec))
	}
}

func TestECTrimUnreachableLastChargingInterval(t *testing.T) {
	ec := &EventCost{
		Usage:          utils.DurationPointer(30 * time.Second),
		AccountSummary: &AccountSummary{},
		Charges: []*ChargingInterval{
			{
				RatingID:   "RT_ID",
				cost:       utils.Float64Pointer(10),
				ecUsageIdx: utils.DurationPointer(1),
			},
		},
	}
	atUsage := 15 * time.Second

	experr := "cannot find last active ChargingInterval"
	rcv, err := ec.Trim(atUsage)

	if err == nil || err.Error() != experr {
		t.Fatalf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	if rcv != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, rcv)
	}
}

func TestECTrimNoActiveIncrement(t *testing.T) {
	ec := &EventCost{
		Usage:          utils.DurationPointer(30 * time.Second),
		AccountSummary: &AccountSummary{},
		Charges: []*ChargingInterval{
			{
				RatingID:       "RT_ID",
				cost:           utils.Float64Pointer(10),
				ecUsageIdx:     utils.DurationPointer(1 * time.Second),
				CompressFactor: 4,
				usage:          utils.DurationPointer(45 * time.Second),
			},
		},
	}
	atUsage := 15 * time.Second

	experr := "no active increment found"
	rcv, err := ec.Trim(atUsage)

	if err == nil || err.Error() != experr {
		t.Fatalf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	if rcv != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, rcv)
	}
}

func TestECTrimUnableToDetectLastActiveChargingInterval(t *testing.T) {
	ec := &EventCost{
		Usage:          utils.DurationPointer(30 * time.Second),
		AccountSummary: &AccountSummary{},
		Charges: []*ChargingInterval{
			{
				RatingID:       "RT_ID1",
				cost:           utils.Float64Pointer(10),
				ecUsageIdx:     utils.DurationPointer(30 * time.Second),
				CompressFactor: 0,
				usage:          utils.DurationPointer(45 * time.Second),
			},
			{
				RatingID:       "RT_ID2",
				cost:           utils.Float64Pointer(10),
				ecUsageIdx:     utils.DurationPointer(25 * time.Second),
				CompressFactor: 4,
				usage:          utils.DurationPointer(45 * time.Second),
			},
		},
	}
	atUsage := 15 * time.Second

	experr := "failed detecting last active ChargingInterval"
	rcv, err := ec.Trim(atUsage)

	if err == nil || err.Error() != experr {
		t.Fatalf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	if rcv != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, rcv)
	}
}

func TestECFieldAsInterfaceEmptyFieldPath(t *testing.T) {
	ec := &EventCost{}
	fldPath := []string{}

	experr := "empty field path"
	rcv, err := ec.FieldAsInterface(fldPath)

	if err == nil || err.Error() != experr {
		t.Fatalf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	if rcv != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, rcv)
	}
}

func TestECfieldAsInterfaceNilECUsage(t *testing.T) {
	ec := &EventCost{}
	fldPath := []string{utils.Usage}

	rcv, err := ec.fieldAsInterface(fldPath)

	if err != nil {
		t.Fatalf("\nexpected: <%+v>, \nreceived: <%+v>", nil, err)
	}

	if rcv != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, rcv)
	}
}

func TestECfieldAsInterfaceNilECCost(t *testing.T) {
	ec := &EventCost{}
	fldPath := []string{utils.Cost}

	rcv, err := ec.fieldAsInterface(fldPath)

	if err != nil {
		t.Fatalf("\nexpected: <%+v>, \nreceived: <%+v>", nil, err)
	}

	if rcv != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, rcv)
	}
}
