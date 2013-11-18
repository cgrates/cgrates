/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2013 ITsysCOM

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
	"github.com/cgrates/cgrates/utils"
	"reflect"
	"testing"
	"time"
)

var (
	destinations = `
#Tag,Prefix
GERMANY,49
GERMANY_O2,41
GERMANY_PREMIUM,43
ALL,49
ALL,41
ALL,43
NAT,0256
NAT,0257
NAT,0723
RET,0723
RET,0724
PSTN_71,+4971
PSTN_72,+4972
PSTN_70,+4970
`
	timings = `
WORKDAYS_00,*any,*any,*any,1;2;3;4;5,00:00:00
WORKDAYS_18,*any,*any,*any,1;2;3;4;5,18:00:00
WEEKENDS,*any,*any,*any,6;7,00:00:00
ONE_TIME_RUN,2012,,,,*asap
`
	rates = `
R1,0,0.2,60s,1s,0,*middle,2
R2,0,0.1,60s,1s,0,*middle,2
R3,0,0.05,60s,1s,0,*middle,2
R4,1,1,1s,1s,0,*up,2
R5,0,0.5,1s,1s,0,*down,2
LANDLINE_OFFPEAK,0,1,1s,60s,0s,*up,4
LANDLINE_OFFPEAK,0,1,1s,1s,60s,*up,4
GBP_71,0.000000,5.55555,1s,1s,0s,*up,4
GBP_72,0.000000,7.77777,1s,1s,0s,*up,4
GBP_70,0.000000,1,1s,1s,0s,*up,4
`
	destinationRates = `
RT_STANDARD,GERMANY,R1
RT_STANDARD,GERMANY_O2,R2
RT_STANDARD,GERMANY_PREMIUM,R2
RT_DEFAULT,ALL,R2
RT_STD_WEEKEND,GERMANY,R2
RT_STD_WEEKEND,GERMANY_O2,R3
P1,NAT,R4
P2,NAT,R5
T1,NAT,LANDLINE_OFFPEAK
T2,GERMANY,GBP_72
T2,GERMANY_O2,GBP_70
T2,GERMANY_PREMIUM,GBP_71
`
	destinationRateTimings = `
STANDARD,RT_STANDARD,WORKDAYS_00,10
STANDARD,RT_STD_WEEKEND,WORKDAYS_18,10
STANDARD,RT_STD_WEEKEND,WEEKENDS,10
PREMIUM,RT_STANDARD,WORKDAYS_00,10
PREMIUM,RT_STD_WEEKEND,WORKDAYS_18,10
PREMIUM,RT_STD_WEEKEND,WEEKENDS,10
DEFAULT,RT_DEFAULT,WORKDAYS_00,10
EVENING,P1,WORKDAYS_00,10
EVENING,P2,WORKDAYS_18,10
EVENING,P2,WEEKENDS,10
TDRT,T1,WORKDAYS_00,10
TDRT,T2,WORKDAYS_00,10
G,RT_STANDARD,WORKDAYS_00,10
R,P1,WORKDAYS_00,10
`
	ratingProfiles = `
CUSTOMER_1,0,*out,rif:from:tm,2012-01-01T00:00:00Z,PREMIUM,danb
CUSTOMER_1,0,*out,rif:from:tm,2012-02-28T00:00:00Z,STANDARD,danb
CUSTOMER_2,0,*out,danb:87.139.12.167,2012-01-01T00:00:00Z,STANDARD,danb
CUSTOMER_1,0,*out,danb,2012-01-01T00:00:00Z,PREMIUM,
vdf,0,*out,rif,2012-01-01T00:00:00Z,EVENING,
vdf,0,*out,rif,2012-02-28T00:00:00Z,EVENING,
vdf,0,*out,minu,2012-01-01T00:00:00Z,EVENING,
vdf,0,*out,*any,2012-02-28T00:00:00Z,EVENING,
vdf,0,*out,one,2012-02-28T00:00:00Z,STANDARD,
vdf,0,*out,inf,2012-02-28T00:00:00Z,STANDARD,inf
vdf,0,*out,fall,2012-02-28T00:00:00Z,PREMIUM,rif
test,0,*out,trp,2013-10-01T00:00:00Z,TDRT,rif;danb
vdf,0,*out,fallback1,2013-11-18T13:45:00Z,G,fallback2
vdf,0,*out,fallback1,2013-11-18T13:46:00Z,G,fallback2
vdf,0,*out,fallback1,2013-11-18T13:47:00Z,G,fallback2
vdf,0,*out,fallback2,2013-11-18T13:45:00Z,R,rif
`
	actions = `
MINI,*topup_reset,*monetary,*out,10,*unlimited,,,10,,10
MINI,*topup,*minutes,*out,100,*unlimited,NAT,test,10,,10
`
	actionTimings = `
MORE_MINUTES,MINI,ONE_TIME_RUN,10
`
	actionTriggers = `
STANDARD_TRIGGER,*minutes,*out,*min_counter,10,GERMANY_O2,SOME_1,10
STANDARD_TRIGGER,*minutes,*out,*max_balance,200,GERMANY,SOME_2,10
`
	accountActions = `
vdf,minitsboy,*out,MORE_MINUTES,STANDARD_TRIGGER
`
)

var csvr *CSVReader

func init() {
	csvr = NewStringCSVReader(storageGetter, ',', destinations, timings, rates, destinationRates, destinationRateTimings, ratingProfiles, actions, actionTimings, actionTriggers, accountActions)
	csvr.LoadDestinations()
	csvr.LoadTimings()
	csvr.LoadRates()
	csvr.LoadDestinationRates()
	csvr.LoadRatingPlans()
	csvr.LoadRatingProfiles()
	csvr.LoadActions()
	csvr.LoadActionTimings()
	csvr.LoadActionTriggers()
	csvr.LoadAccountActions()
	csvr.WriteToDatabase(false, false)
	storageGetter.PreCache(nil, nil)
}

func TestLoadDestinations(t *testing.T) {
	if len(csvr.destinations) != 9 {
		t.Error("Failed to load destinations: ", len(csvr.destinations))
	}
	for _, d := range csvr.destinations {
		switch d.Id {
		case "NAT":
			if !reflect.DeepEqual(d.Prefixes, []string{`0256`, `0257`, `0723`}) {
				t.Error("Faild to load destinations", d)
			}
		case "ALL":
			if !reflect.DeepEqual(d.Prefixes, []string{`49`, `41`, `43`}) {
				t.Error("Faild to load destinations", d)
			}
		case "RET":
			if !reflect.DeepEqual(d.Prefixes, []string{`0723`, `0724`}) {
				t.Error("Faild to load destinations", d)
			}
		case "GERMANY":
			if !reflect.DeepEqual(d.Prefixes, []string{`49`}) {
				t.Error("Faild to load destinations", d)
			}
		case "GERMANY_O2":
			if !reflect.DeepEqual(d.Prefixes, []string{`41`}) {
				t.Error("Faild to load destinations", d)
			}
		case "GERMANY_PREMIUM":
			if !reflect.DeepEqual(d.Prefixes, []string{`43`}) {
				t.Error("Faild to load destinations", d)
			}
		case "PSTN_71":
			if !reflect.DeepEqual(d.Prefixes, []string{`+4971`}) {
				t.Error("Faild to load destinations", d)
			}
		case "PSTN_72":
			if !reflect.DeepEqual(d.Prefixes, []string{`+4972`}) {
				t.Error("Faild to load destinations", d)
			}
		case "PSTN_70":
			if !reflect.DeepEqual(d.Prefixes, []string{`+4970`}) {
				t.Error("Faild to load destinations", d)
			}
		default:
			t.Error("Unknown destination tag!")
		}
	}
}

func TestLoadTimimgs(t *testing.T) {
	if len(csvr.timings) != 4 {
		t.Error("Failed to load timings: ", csvr.timings)
	}
	timing := csvr.timings["WORKDAYS_00"]
	if !reflect.DeepEqual(timing, &utils.TPTiming{
		Id:        "WORKDAYS_00",
		Years:     utils.Years{},
		Months:    utils.Months{},
		MonthDays: utils.MonthDays{},
		WeekDays:  utils.WeekDays{1, 2, 3, 4, 5},
		StartTime: "00:00:00",
	}) {
		t.Error("Error loading timing: ", timing)
	}
	timing = csvr.timings["WORKDAYS_18"]
	if !reflect.DeepEqual(timing, &utils.TPTiming{
		Id:        "WORKDAYS_18",
		Years:     utils.Years{},
		Months:    utils.Months{},
		MonthDays: utils.MonthDays{},
		WeekDays:  utils.WeekDays{1, 2, 3, 4, 5},
		StartTime: "18:00:00",
	}) {
		t.Error("Error loading timing: ", timing)
	}
	timing = csvr.timings["WEEKENDS"]
	if !reflect.DeepEqual(timing, &utils.TPTiming{
		Id:        "WEEKENDS",
		Years:     utils.Years{},
		Months:    utils.Months{},
		MonthDays: utils.MonthDays{},
		WeekDays:  utils.WeekDays{time.Saturday, time.Sunday},
		StartTime: "00:00:00",
	}) {
		t.Error("Error loading timing: ", timing)
	}
	timing = csvr.timings["ONE_TIME_RUN"]
	if !reflect.DeepEqual(timing, &utils.TPTiming{
		Id:        "ONE_TIME_RUN",
		Years:     utils.Years{2012},
		Months:    utils.Months{},
		MonthDays: utils.MonthDays{},
		WeekDays:  utils.WeekDays{},
		StartTime: "*asap",
	}) {
		t.Error("Error loading timing: ", timing)
	}
}

func TestLoadRates(t *testing.T) {
	if len(csvr.rates) != 9 {
		t.Error("Failed to load rates: ", csvr.rates)
	}
	rate := csvr.rates["R1"].RateSlots[0]
	expctRs, err := utils.NewRateSlot(0, 0.2, "60s", "1s", "0", utils.ROUNDING_MIDDLE, 2)
	if err != nil {
		t.Error("Error loading rate: ", rate, err.Error())
	} else if !reflect.DeepEqual(rate, expctRs) ||
		rate.RateUnitDuration() != expctRs.RateUnitDuration() ||
		rate.RateIncrementDuration() != expctRs.RateIncrementDuration() ||
		rate.GroupIntervalStartDuration() != expctRs.GroupIntervalStartDuration() {
		t.Error("Error loading rate: ", rate, expctRs)
	}
	rate = csvr.rates["R2"].RateSlots[0]
	if expctRs, err = utils.NewRateSlot(0, 0.1, "60s", "1s", "0", utils.ROUNDING_MIDDLE, 2); err != nil {
		t.Error("Error loading rate: ", rate, err.Error())
	} else if !reflect.DeepEqual(rate, expctRs) ||
		rate.RateUnitDuration() != expctRs.RateUnitDuration() ||
		rate.RateIncrementDuration() != expctRs.RateIncrementDuration() ||
		rate.GroupIntervalStartDuration() != expctRs.GroupIntervalStartDuration() {
		t.Error("Error loading rate: ", rate)
	}
	rate = csvr.rates["R3"].RateSlots[0]
	if expctRs, err = utils.NewRateSlot(0, 0.05, "60s", "1s", "0", utils.ROUNDING_MIDDLE, 2); err != nil {
		t.Error("Error loading rate: ", rate, err.Error())
	} else if !reflect.DeepEqual(rate, expctRs) ||
		rate.RateUnitDuration() != expctRs.RateUnitDuration() ||
		rate.RateIncrementDuration() != expctRs.RateIncrementDuration() ||
		rate.GroupIntervalStartDuration() != expctRs.GroupIntervalStartDuration() {
		t.Error("Error loading rate: ", rate)
	}
	rate = csvr.rates["R4"].RateSlots[0]
	if expctRs, err = utils.NewRateSlot(1, 1.0, "1s", "1s", "0", utils.ROUNDING_UP, 2); err != nil {
		t.Error("Error loading rate: ", rate, err.Error())
	} else if !reflect.DeepEqual(rate, expctRs) ||
		rate.RateUnitDuration() != expctRs.RateUnitDuration() ||
		rate.RateIncrementDuration() != expctRs.RateIncrementDuration() ||
		rate.GroupIntervalStartDuration() != expctRs.GroupIntervalStartDuration() {
		t.Error("Error loading rate: ", rate)
	}
	rate = csvr.rates["R5"].RateSlots[0]
	if expctRs, err = utils.NewRateSlot(0, 0.5, "1s", "1s", "0", utils.ROUNDING_DOWN, 2); err != nil {
		t.Error("Error loading rate: ", rate, err.Error())
	} else if !reflect.DeepEqual(rate, expctRs) ||
		rate.RateUnitDuration() != expctRs.RateUnitDuration() ||
		rate.RateIncrementDuration() != expctRs.RateIncrementDuration() ||
		rate.GroupIntervalStartDuration() != expctRs.GroupIntervalStartDuration() {
		t.Error("Error loading rate: ", rate)
	}
	rate = csvr.rates["LANDLINE_OFFPEAK"].RateSlots[0]
	if expctRs, err = utils.NewRateSlot(0, 1, "1s", "60s", "0s", utils.ROUNDING_UP, 4); err != nil {
		t.Error("Error loading rate: ", rate, err.Error())
	} else if !reflect.DeepEqual(rate, expctRs) ||
		rate.RateUnitDuration() != expctRs.RateUnitDuration() ||
		rate.RateIncrementDuration() != expctRs.RateIncrementDuration() ||
		rate.GroupIntervalStartDuration() != expctRs.GroupIntervalStartDuration() {
		t.Error("Error loading rate: ", rate)
	}
	rate = csvr.rates["LANDLINE_OFFPEAK"].RateSlots[1]
	if expctRs, err = utils.NewRateSlot(0, 1, "1s", "1s", "60s", utils.ROUNDING_UP, 4); err != nil {
		t.Error("Error loading rate: ", rate, err.Error())
	} else if !reflect.DeepEqual(rate, expctRs) ||
		rate.RateUnitDuration() != expctRs.RateUnitDuration() ||
		rate.RateIncrementDuration() != expctRs.RateIncrementDuration() ||
		rate.GroupIntervalStartDuration() != expctRs.GroupIntervalStartDuration() {
		t.Error("Error loading rate: ", rate)
	}
}

func TestLoadDestinationRates(t *testing.T) {
	if len(csvr.destinationRates) != 7 {
		t.Error("Failed to load destinationrates: ", csvr.destinationRates)
	}
	drs := csvr.destinationRates["RT_STANDARD"]
	dr := &utils.TPDestinationRate{
		TPid:              "",
		DestinationRateId: "RT_STANDARD",
		DestinationRates: []*utils.DestinationRate{
			&utils.DestinationRate{
				DestinationId: "GERMANY",
				Rate:          csvr.rates["R1"],
			},
			&utils.DestinationRate{
				DestinationId: "GERMANY_O2",
				Rate:          csvr.rates["R2"],
			},
			&utils.DestinationRate{
				DestinationId: "GERMANY_PREMIUM",
				Rate:          csvr.rates["R2"],
			},
		},
	}
	if !reflect.DeepEqual(drs, dr) {
		t.Errorf("Error loading destination rate: \n%+v \n%+v", drs, dr)
	}
	drs = csvr.destinationRates["RT_DEFAULT"]
	if !reflect.DeepEqual(drs, &utils.TPDestinationRate{
		DestinationRateId: "RT_DEFAULT",
		DestinationRates: []*utils.DestinationRate{
			&utils.DestinationRate{
				DestinationId: "ALL",
				Rate:          csvr.rates["R2"],
			},
		},
	}) {
		t.Errorf("Error loading destination rate: %+v", drs)
	}
	drs = csvr.destinationRates["RT_STD_WEEKEND"]
	if !reflect.DeepEqual(drs, &utils.TPDestinationRate{
		DestinationRateId: "RT_STD_WEEKEND",
		DestinationRates: []*utils.DestinationRate{
			&utils.DestinationRate{
				DestinationId: "GERMANY",
				Rate:          csvr.rates["R2"],
			},
			&utils.DestinationRate{
				DestinationId: "GERMANY_O2",
				Rate:          csvr.rates["R3"],
			},
		},
	}) {
		t.Error("Error loading destination rate: ", drs)
	}
	drs = csvr.destinationRates["P1"]
	if !reflect.DeepEqual(drs, &utils.TPDestinationRate{
		DestinationRateId: "P1",
		DestinationRates: []*utils.DestinationRate{
			&utils.DestinationRate{
				DestinationId: "NAT",
				Rate:          csvr.rates["R4"],
			},
		},
	}) {
		t.Error("Error loading destination rate: ", drs)
	}
	drs = csvr.destinationRates["P2"]
	if !reflect.DeepEqual(drs, &utils.TPDestinationRate{
		DestinationRateId: "P2",
		DestinationRates: []*utils.DestinationRate{
			&utils.DestinationRate{
				DestinationId: "NAT",
				Rate:          csvr.rates["R5"],
			},
		},
	}) {
		t.Error("Error loading destination rate: ", drs)
	}
	drs = csvr.destinationRates["T1"]
	if !reflect.DeepEqual(drs, &utils.TPDestinationRate{
		DestinationRateId: "T1",
		DestinationRates: []*utils.DestinationRate{
			&utils.DestinationRate{
				DestinationId: "NAT",
				Rate:          csvr.rates["LANDLINE_OFFPEAK"],
			},
		},
	}) {
		t.Error("Error loading destination rate: ", drs)
	}
	drs = csvr.destinationRates["T2"]
	if !reflect.DeepEqual(drs, &utils.TPDestinationRate{
		DestinationRateId: "T2",
		DestinationRates: []*utils.DestinationRate{
			&utils.DestinationRate{
				DestinationId: "GERMANY",
				Rate:          csvr.rates["GBP_72"],
			},
			&utils.DestinationRate{
				DestinationId: "GERMANY_O2",
				Rate:          csvr.rates["GBP_70"],
			},
			&utils.DestinationRate{
				DestinationId: "GERMANY_PREMIUM",
				Rate:          csvr.rates["GBP_71"],
			},
		},
	}) {
		t.Error("Error loading destination rate: ", drs)
	}
}

func TestLoadDestinationRateTimings(t *testing.T) {
	if len(csvr.ratingPlans) != 7 {
		t.Error("Failed to load rate timings: ", csvr.ratingPlans)
	}
	rplan := csvr.ratingPlans["STANDARD"]
	expected := &RatingPlan{
		Id: "STANDARD",
		Timings: map[string]*RITiming{
			"14ae6e41": &RITiming{
				Years:     utils.Years{},
				Months:    utils.Months{},
				MonthDays: utils.MonthDays{},
				WeekDays:  utils.WeekDays{1, 2, 3, 4, 5},
				StartTime: "00:00:00",
			},
			"9a6f8e32": &RITiming{
				Years:     utils.Years{},
				Months:    utils.Months{},
				MonthDays: utils.MonthDays{},
				WeekDays:  utils.WeekDays{1, 2, 3, 4, 5},
				StartTime: "18:00:00",
			},
			"7181e535": &RITiming{
				Years:     utils.Years{},
				Months:    utils.Months{},
				MonthDays: utils.MonthDays{},
				WeekDays:  utils.WeekDays{time.Saturday, time.Sunday},
				StartTime: "00:00:00",
			},
		},
		Ratings: map[string]*RIRate{
			"d54545c1": &RIRate{
				ConnectFee: 0,
				Rates: []*Rate{
					&Rate{
						GroupIntervalStart: 0,
						Value:              0.2,
						RateIncrement:      time.Second,
						RateUnit:           time.Minute,
					},
				},
				RoundingMethod:   utils.ROUNDING_MIDDLE,
				RoundingDecimals: 2,
			},
			"4bb00b9c": &RIRate{
				ConnectFee: 0,
				Rates: []*Rate{
					&Rate{
						GroupIntervalStart: 0,
						Value:              0.1,
						RateIncrement:      time.Second,
						RateUnit:           time.Minute,
					},
				},
				RoundingMethod:   utils.ROUNDING_MIDDLE,
				RoundingDecimals: 2,
			},
			"e06c337f": &RIRate{
				ConnectFee: 0,
				Rates: []*Rate{
					&Rate{
						GroupIntervalStart: 0,
						Value:              0.05,
						RateIncrement:      time.Second,
						RateUnit:           time.Minute,
					},
				},
				RoundingMethod:   utils.ROUNDING_MIDDLE,
				RoundingDecimals: 2,
			},
		},
		DestinationRates: map[string]RPRateList{
			"GERMANY": []*RPRate{
				&RPRate{
					Timing: "14ae6e41",
					Rating: "d54545c1",
					Weight: 10,
				},
				&RPRate{
					Timing: "9a6f8e32",
					Rating: "4bb00b9c",
					Weight: 10,
				},
				&RPRate{
					Timing: "7181e535",
					Rating: "4bb00b9c",
					Weight: 10,
				},
			},
			"GERMANY_O2": []*RPRate{
				&RPRate{
					Timing: "14ae6e41",
					Rating: "4bb00b9c",
					Weight: 10,
				},
				&RPRate{
					Timing: "9a6f8e32",
					Rating: "e06c337f",
					Weight: 10,
				},
				&RPRate{
					Timing: "7181e535",
					Rating: "e06c337f",
					Weight: 10,
				},
			},
			"GERMANY_PREMIUM": []*RPRate{
				&RPRate{
					Timing: "14ae6e41",
					Rating: "4bb00b9c",
					Weight: 10,
				},
			},
		},
	}
	if !reflect.DeepEqual(rplan, expected) {
		t.Errorf("Error loading destination rate timing: %+v", rplan)
	}
}

func TestLoadRatingProfiles(t *testing.T) {
	if len(csvr.ratingProfiles) != 12 {
		t.Error("Failed to load rating profiles: ", len(csvr.ratingProfiles), csvr.ratingProfiles)
	}
	rp := csvr.ratingProfiles["*out:test:0:trp"]
	expected := &RatingProfile{
		Id: "*out:test:0:trp",
		RatingPlanActivations: RatingPlanActivations{&RatingPlanActivation{
			ActivationTime: time.Date(2013, 10, 1, 0, 0, 0, 0, time.UTC),
			RatingPlanId:   "TDRT",
			FallbackKeys:   []string{"*out:test:0:rif", "*out:test:0:danb"},
		}},
	}
	if !reflect.DeepEqual(rp, expected) {
		t.Errorf("Error loading rating profile: %+v", rp.RatingPlanActivations[0])
	}
}

func TestLoadActions(t *testing.T) {
	if len(csvr.actions) != 1 {
		t.Error("Failed to load actions: ", csvr.actions)
	}
	as := csvr.actions["MINI"]
	expected := []*Action{
		&Action{
			Id:               as[0].Id,
			ActionType:       TOPUP_RESET,
			BalanceId:        CREDIT,
			Direction:        OUTBOUND,
			ExpirationString: UNLIMITED,
			ExtraParameters:  "",
			Weight:           10,
			Balance: &Balance{
				Uuid:   as[0].Balance.Uuid,
				Value:  10,
				Weight: 10,
			},
		},
		&Action{
			Id:               as[1].Id,
			ActionType:       TOPUP,
			BalanceId:        MINUTES,
			Direction:        OUTBOUND,
			ExpirationString: UNLIMITED,
			ExtraParameters:  "",
			Weight:           10,
			Balance: &Balance{
				Uuid:          as[1].Balance.Uuid,
				Value:         100,
				Weight:        10,
				RateSubject:   "test",
				DestinationId: "NAT",
			},
		},
	}
	if !reflect.DeepEqual(as, expected) {
		t.Error("Error loading action: ", as)
	}
}

func TestLoadActionTimings(t *testing.T) {
	if len(csvr.actionsTimings) != 1 {
		t.Error("Failed to load action timings: ", csvr.actionsTimings)
	}
	atm := csvr.actionsTimings["MORE_MINUTES"][0]
	expected := &ActionTiming{
		Id:             atm.Id,
		Tag:            "MORE_MINUTES",
		UserBalanceIds: []string{"*out:vdf:minitsboy"},
		Timing: &RateInterval{
			Timing: &RITiming{
				Years:     utils.Years{2012},
				Months:    utils.Months{},
				MonthDays: utils.MonthDays{},
				WeekDays:  utils.WeekDays{},
				StartTime: ASAP,
			},
		},
		Weight:    10,
		ActionsId: "MINI",
	}
	if !reflect.DeepEqual(atm, expected) {
		t.Error("Error loading action timing: ", atm, expected)
	}
}

func TestLoadActionTriggers(t *testing.T) {
	if len(csvr.actionsTriggers) != 1 {
		t.Error("Failed to load action triggers: ", csvr.actionsTriggers)
	}
	atr := csvr.actionsTriggers["STANDARD_TRIGGER"][0]
	expected := &ActionTrigger{
		Id:             atr.Id,
		BalanceId:      MINUTES,
		Direction:      OUTBOUND,
		ThresholdType:  TRIGGER_MIN_COUNTER,
		ThresholdValue: 10,
		DestinationId:  "GERMANY_O2",
		Weight:         10,
		ActionsId:      "SOME_1",
		Executed:       false,
	}
	if !reflect.DeepEqual(atr, expected) {
		t.Error("Error loading action trigger: ", atr)
	}
	atr = csvr.actionsTriggers["STANDARD_TRIGGER"][1]
	expected = &ActionTrigger{
		Id:             atr.Id,
		BalanceId:      MINUTES,
		Direction:      OUTBOUND,
		ThresholdType:  TRIGGER_MAX_BALANCE,
		ThresholdValue: 200,
		DestinationId:  "GERMANY",
		Weight:         10,
		ActionsId:      "SOME_2",
		Executed:       false,
	}
	if !reflect.DeepEqual(atr, expected) {
		t.Error("Error loading action trigger: ", atr)
	}
}

func TestLoadAccountActions(t *testing.T) {
	if len(csvr.accountActions) != 1 {
		t.Error("Failed to load account actions: ", csvr.accountActions)
	}
	aa := csvr.accountActions[0]
	expected := &UserBalance{
		Id:             "*out:vdf:minitsboy",
		Type:           UB_TYPE_PREPAID,
		ActionTriggers: csvr.actionsTriggers["STANDARD_TRIGGER"],
	}
	if !reflect.DeepEqual(aa, expected) {
		t.Error("Error loading account action: ", aa)
	}
}

/*
vdf,minitsboy,*out,MORE_MINUTES,STANDARD_TRIGGER
*/
