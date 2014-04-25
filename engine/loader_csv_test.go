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
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"
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
DST_UK_Mobile_BIG5,447956
URG,112
*any,
`
	timings = `
WORKDAYS_00,*any,*any,*any,1;2;3;4;5,00:00:00
WORKDAYS_18,*any,*any,*any,1;2;3;4;5,18:00:00
WEEKENDS,*any,*any,*any,6;7,00:00:00
ONE_TIME_RUN,2012,,,,*asap
ALWAYS,*any,*any,*any,*any,00:00:00
ASAP,*any,*any,*any,*any,*asap
`
	rates = `
R1,0,0.2,60,1,0,*middle,2
R2,0,0.1,60s,1s,0,*middle,2
R3,0,0.05,60s,1s,0,*middle,2
R4,1,1,1s,1s,0,*up,2
R5,0,0.5,1s,1s,0,*down,2
LANDLINE_OFFPEAK,0,1,1,60,0,*up,4
LANDLINE_OFFPEAK,0,1,1,1,60,*up,4
GBP_71,0.000000,5.55555,1s,1s,0s,*up,4
GBP_72,0.000000,7.77777,1s,1s,0s,*up,4
GBP_70,0.000000,1,1,1,0,*up,4
RT_UK_Mobile_BIG5_PKG,0.01,0,20s,20s,0s,*up,8
RT_UK_Mobile_BIG5,0.01,0.10,1s,1s,0s,*up,8
R_URG,0,0,1,1,0,*down,2
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
DR_UK_Mobile_BIG5_PKG,DST_UK_Mobile_BIG5,RT_UK_Mobile_BIG5_PKG
DR_UK_Mobile_BIG5,DST_UK_Mobile_BIG5,RT_UK_Mobile_BIG5
DATA_RATE,*any,LANDLINE_OFFPEAK
RT_URG,URG,R_URG
`
	ratingPlans = `
STANDARD,RT_STANDARD,WORKDAYS_00,10
STANDARD,RT_STD_WEEKEND,WORKDAYS_18,10
STANDARD,RT_STD_WEEKEND,WEEKENDS,10
STANDARD,RT_URG,ALWAYS,20
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
RP_UK_Mobile_BIG5_PKG,DR_UK_Mobile_BIG5_PKG,ALWAYS,10
RP_UK,DR_UK_Mobile_BIG5,ALWAYS,10
RP_DATA,DATA_RATE,ALWAYS,10
`
	ratingProfiles = `
*out,CUSTOMER_1,0,rif:from:tm,2012-01-01T00:00:00Z,PREMIUM,danb
*out,CUSTOMER_1,0,rif:from:tm,2012-02-28T00:00:00Z,STANDARD,danb
*out,CUSTOMER_2,0,danb:87.139.12.167,2012-01-01T00:00:00Z,STANDARD,danb
*out,CUSTOMER_1,0,danb,2012-01-01T00:00:00Z,PREMIUM,
*out,vdf,0,rif,2012-01-01T00:00:00Z,EVENING,
*out,vdf,0,rif,2012-02-28T00:00:00Z,EVENING,
*out,vdf,0,minu;a1;a2;a3,2012-01-01T00:00:00Z,EVENING,
*out,vdf,0,*any,2012-02-28T00:00:00Z,EVENING,
*out,vdf,0,one,2012-02-28T00:00:00Z,STANDARD,
*out,vdf,0,inf,2012-02-28T00:00:00Z,STANDARD,inf
*out,vdf,0,fall,2012-02-28T00:00:00Z,PREMIUM,rif
*out,test,0,trp,2013-10-01T00:00:00Z,TDRT,rif;danb
*out,vdf,0,fallback1,2013-11-18T13:45:00Z,G,fallback2
*out,vdf,0,fallback1,2013-11-18T13:46:00Z,G,fallback2
*out,vdf,0,fallback1,2013-11-18T13:47:00Z,G,fallback2
*out,vdf,0,fallback2,2013-11-18T13:45:00Z,R,rif
*out,cgrates.org,call,*any,2013-01-06T00:00:00Z,RP_UK,
*out,cgrates.org,call,discounted_minutes,2013-01-06T00:00:00Z,RP_UK_Mobile_BIG5_PKG,
*out,cgrates.org,data,rif,2013-01-06T00:00:00Z,RP_DATA,
`
	sharedGroups = `
SG1,*any,*lowest,
SG2,*any,*lowest,one
SG3,*any,*lowest,
`

	actions = `
MINI,*topup_reset,*monetary,*out,10,*unlimited,,,10,,,10
MINI,*topup,*call_duration,*out,100,*unlimited,NAT,test,10,,,10
SHARED,*topup,*monetary,*out,100,*unlimited,,,10,SG1,,10
TOPUP10_AC,*topup_reset,*monetary,*out,1,*unlimited,*any,,10,,,10
TOPUP10_AC1,*topup_reset,*call_duration,*out,40,*unlimited,DST_UK_Mobile_BIG5,discounted_minutes,10,,,10
SE0,*topup_reset,*monetary,*out,0,*unlimited,,,10,SG2,,10
SE10,*topup_reset,*monetary,*out,10,*unlimited,,,5,SG2,,10
SE10,*topup,*monetary,*out,10,*unlimited,,,10,,,10
EE0,*topup_reset,*monetary,*out,0,*unlimited,,,10,SG3,,10
EE0,*allow_negative,*monetary,*out,0,*unlimited,,,10,,,10
`
	actionTimings = `
MORE_MINUTES,MINI,ONE_TIME_RUN,10
MORE_MINUTES,SHARED,ONE_TIME_RUN,10
TOPUP10_AT,TOPUP10_AC,ASAP,10
TOPUP10_AT,TOPUP10_AC1,ASAP,10
TOPUP_SHARED0_AT,SE0,ASAP,10
TOPUP_SHARED10_AT,SE10,ASAP,10
TOPUP_EMPTY_AT,EE0,ASAP,10
`
	actionTriggers = `
STANDARD_TRIGGER,*call_duration,*out,*min_counter,10,false,GERMANY_O2,SOME_1,10
STANDARD_TRIGGER,*call_duration,*out,*max_balance,200,false,GERMANY,SOME_2,10
STANDARD_TRIGGERS,*monetary,*out,*min_balance,2,false,,LOG_WARNING,10
STANDARD_TRIGGERS,*monetary,*out,*max_balance,20,false,,LOG_WARNING,10
STANDARD_TRIGGERS,*monetary,*out,*max_counter,5,false,FS_USERS,LOG_WARNING,10
`
	accountActions = `
vdf,minitsboy;a1;a2,*out,MORE_MINUTES,STANDARD_TRIGGER
cgrates.org,12345,*out,TOPUP10_AT,STANDARD_TRIGGERS
vdf,empty0,*out,TOPUP_SHARED0_AT,
vdf,empty10,*out,TOPUP_SHARED10_AT,
vdf,emptyX,*out,TOPUP_EMPTY_AT,
vdf,emptyY,*out,TOPUP_EMPTY_AT,
`

	derivedCharges = `
#Tenant,Tor,Direction,Account,Subject,RunId,ReqTypeField,DirectionField,TenantField,TorField,AccountField,SubjectField,DestinationField,SetupTimeField,AnswerTimeField,DurationField
cgrates.org,call,*out,dan,dan,extra1,^prepaid,,,,rif,rif,,,,
cgrates.org,call,*out,dan,dan,extra2,,,,,ivo,ivo,,,,
cgrates.org,call,*out,dan,*any,extra1,,,,,rif2,rif2,,,,
`
)

var csvr *CSVReader

func init() {
	csvr = NewStringCSVReader(dataStorage, accountingStorage, ',', destinations, timings, rates, destinationRates, ratingPlans, ratingProfiles,
		sharedGroups, actions, actionTimings, actionTriggers, accountActions, derivedCharges)
	csvr.LoadDestinations()
	csvr.LoadTimings()
	csvr.LoadRates()
	csvr.LoadDestinationRates()
	csvr.LoadRatingPlans()
	csvr.LoadRatingProfiles()
	csvr.LoadSharedGroups()
	csvr.LoadActions()
	csvr.LoadActionTimings()
	csvr.LoadActionTriggers()
	csvr.LoadAccountActions()
	csvr.LoadDerivedChargers()
	csvr.WriteToDatabase(false, false)
	dataStorage.CacheRating(nil, nil, nil, nil)
	accountingStorage.CacheAccounting(nil, nil, nil, nil)
}

func TestLoadDestinations(t *testing.T) {
	if len(csvr.destinations) != 12 {
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
		}
	}
}

func TestLoadTimimgs(t *testing.T) {
	if len(csvr.timings) != 6 {
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
	if len(csvr.rates) != 12 {
		t.Error("Failed to load rates: ", csvr.rates)
	}
	rate := csvr.rates["R1"].RateSlots[0]
	expctRs, err := utils.NewRateSlot(0, 0.2, "60", "1", "0", utils.ROUNDING_MIDDLE, 2)
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
	if expctRs, err = utils.NewRateSlot(0, 1, "1", "60", "0", utils.ROUNDING_UP, 4); err != nil {
		t.Error("Error loading rate: ", rate, err.Error())
	} else if !reflect.DeepEqual(rate, expctRs) ||
		rate.RateUnitDuration() != expctRs.RateUnitDuration() ||
		rate.RateIncrementDuration() != expctRs.RateIncrementDuration() ||
		rate.GroupIntervalStartDuration() != expctRs.GroupIntervalStartDuration() {
		t.Error("Error loading rate: ", rate)
	}
	rate = csvr.rates["LANDLINE_OFFPEAK"].RateSlots[1]
	if expctRs, err = utils.NewRateSlot(0, 1, "1", "1", "60", utils.ROUNDING_UP, 4); err != nil {
		t.Error("Error loading rate: ", rate, err.Error())
	} else if !reflect.DeepEqual(rate, expctRs) ||
		rate.RateUnitDuration() != expctRs.RateUnitDuration() ||
		rate.RateIncrementDuration() != expctRs.RateIncrementDuration() ||
		rate.GroupIntervalStartDuration() != expctRs.GroupIntervalStartDuration() {
		t.Error("Error loading rate: ", rate)
	}
}

func TestLoadDestinationRates(t *testing.T) {
	if len(csvr.destinationRates) != 11 {
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

func TestLoadRatingPlans(t *testing.T) {
	if len(csvr.ratingPlans) != 10 {
		t.Error("Failed to load rating plans: ", len(csvr.ratingPlans))
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
			"96c78ff5": &RITiming{
				Years:     utils.Years{},
				Months:    utils.Months{},
				MonthDays: utils.MonthDays{},
				WeekDays:  utils.WeekDays{},
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
			"2efe78aa": &RIRate{
				ConnectFee: 0,
				Rates: []*Rate{
					&Rate{
						GroupIntervalStart: 0,
						Value:              0,
						RateIncrement:      time.Second,
						RateUnit:           time.Second,
					},
				},
				RoundingMethod:   utils.ROUNDING_DOWN,
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
			"URG": []*RPRate{
				&RPRate{
					Timing: "96c78ff5",
					Rating: "2efe78aa",
					Weight: 20,
				},
			},
		},
	}
	if !reflect.DeepEqual(rplan, expected) {
		t.Errorf("Error loading destination rate timing: %+v", rplan.DestinationRates["URG"][0])
	}
}

func TestLoadRatingProfiles(t *testing.T) {
	if len(csvr.ratingProfiles) != 15 {
		t.Error("Failed to load rating profiles: ", len(csvr.ratingProfiles), csvr.ratingProfiles)
	}
	rp := csvr.ratingProfiles["*out:test:0:trp"]
	expected := &RatingProfile{
		Id: "*out:test:0:trp",
		RatingPlanActivations: RatingPlanActivations{&RatingPlanActivation{
			ActivationTime: time.Date(2013, 10, 1, 0, 0, 0, 0, time.UTC),
			RatingPlanId:   "TDRT",
			FallbackKeys:   []string{"*out:test:0:danb", "*out:test:0:rif"},
		}},
	}
	if !reflect.DeepEqual(rp, expected) {
		t.Errorf("Error loading rating profile: %+v", rp.RatingPlanActivations[0])
	}
}

func TestLoadActions(t *testing.T) {
	if len(csvr.actions) != 7 {
		t.Error("Failed to load actions: ", csvr.actions)
	}
	as1 := csvr.actions["MINI"]
	expected := []*Action{
		&Action{
			Id:               as1[0].Id,
			ActionType:       TOPUP_RESET,
			BalanceType:      CREDIT,
			Direction:        OUTBOUND,
			ExpirationString: UNLIMITED,
			ExtraParameters:  "",
			Weight:           10,
			Balance: &Balance{
				Uuid:   as1[0].Balance.Uuid,
				Value:  10,
				Weight: 10,
			},
		},
		&Action{
			Id:               as1[1].Id,
			ActionType:       TOPUP,
			BalanceType:      MINUTES,
			Direction:        OUTBOUND,
			ExpirationString: UNLIMITED,
			ExtraParameters:  "",
			Weight:           10,
			Balance: &Balance{
				Uuid:          as1[1].Balance.Uuid,
				Value:         100,
				Weight:        10,
				RatingSubject: "test",
				DestinationId: "NAT",
			},
		},
	}
	if !reflect.DeepEqual(as1, expected) {
		t.Error("Error loading action: ", as1)
	}
	as2 := csvr.actions["SHARED"]
	expected = []*Action{
		&Action{
			Id:               as2[0].Id,
			ActionType:       TOPUP,
			BalanceType:      CREDIT,
			Direction:        OUTBOUND,
			ExpirationString: UNLIMITED,
			Weight:           10,
			Balance: &Balance{
				Uuid:        as2[0].Balance.Uuid,
				Value:       100,
				Weight:      10,
				SharedGroup: "SG1",
			},
		},
	}
	if !reflect.DeepEqual(as2, expected) {
		t.Errorf("Error loading action: %+v", as2[0].Balance)
	}
}

func TestLoadSharedGroups(t *testing.T) {
	if len(csvr.sharedGroups) != 3 {
		t.Error("Failed to load actions: ", csvr.sharedGroups)
	}

	sg1 := csvr.sharedGroups["SG1"]
	expected := &SharedGroup{
		Id: "SG1",
		AccountParameters: map[string]*SharingParameters{
			"*any": &SharingParameters{
				Strategy:      "*lowest",
				RatingSubject: "",
			},
		},
	}
	if !reflect.DeepEqual(sg1, expected) {
		t.Error("Error loading shared group: ", sg1.AccountParameters)
	}
	sg2 := csvr.sharedGroups["SG2"]
	expected = &SharedGroup{
		Id: "SG2",
		AccountParameters: map[string]*SharingParameters{
			"*any": &SharingParameters{
				Strategy:      "*lowest",
				RatingSubject: "one",
			},
		},
	}
	if !reflect.DeepEqual(sg2, expected) {
		t.Error("Error loading shared group: ", sg2.AccountParameters)
	}
	/*sg, _ := accountingStorage.GetSharedGroup("SG1", false)
	if len(sg.Members) != 0 {
		t.Errorf("Memebers should be empty: %+v", sg)
	}

	// execute action timings to fill memebers
	atm := csvr.actionsTimings["MORE_MINUTES"][1]
	atm.Execute()
	atm.actions, atm.stCache = nil, time.Time{}

	sg, _ = accountingStorage.GetSharedGroup("SG1", false)
	if len(sg.Members) != 1 {
		t.Errorf("Memebers should not be empty: %+v", sg)
	}*/
}

func TestLoadActionTimings(t *testing.T) {
	if len(csvr.actionsTimings) != 5 {
		t.Error("Failed to load action timings: ", csvr.actionsTimings)
	}
	atm := csvr.actionsTimings["MORE_MINUTES"][0]
	expected := &ActionTiming{
		Uuid:       atm.Uuid,
		Id:         "MORE_MINUTES",
		AccountIds: []string{"*out:vdf:minitsboy"},
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
		t.Errorf("Error loading action timing:\n%+v\n%+v", atm, expected)
	}
}

func TestLoadActionTriggers(t *testing.T) {
	if len(csvr.actionsTriggers) != 2 {
		t.Error("Failed to load action triggers: ", csvr.actionsTriggers)
	}
	atr := csvr.actionsTriggers["STANDARD_TRIGGER"][0]
	expected := &ActionTrigger{
		Id:             atr.Id,
		BalanceType:    MINUTES,
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
		BalanceType:    MINUTES,
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
	if len(csvr.accountActions) != 6 {
		t.Error("Failed to load account actions: ", csvr.accountActions)
	}
	aa := csvr.accountActions["*out:vdf:minitsboy"]
	expected := &Account{
		Id:             "*out:vdf:minitsboy",
		ActionTriggers: csvr.actionsTriggers["STANDARD_TRIGGER"],
	}
	if !reflect.DeepEqual(aa, expected) {
		t.Error("Error loading account action: ", aa)
	}
	// test that it does not overwrite balances
	existing, err := accountingStorage.GetAccount(aa.Id)
	if err != nil || len(existing.BalanceMap) != 2 {
		t.Errorf("The account was not set before load: %+v", existing)
	}
	accountingStorage.SetAccount(aa)
	existing, err = accountingStorage.GetAccount(aa.Id)
	if err != nil || len(existing.BalanceMap) != 2 {
		t.Errorf("The set account altered the balances: %+v", existing)
	}
}

func TestLoadRpAliases(t *testing.T) {
	if len(csvr.rpAliases) != 3 {
		t.Error("Failed to load rp aliases: ", csvr.rpAliases)
	}
	if csvr.rpAliases["a1"] != "minu" ||
		csvr.rpAliases["a2"] != "minu" ||
		csvr.rpAliases["a3"] != "minu" {
		t.Error("Error loading rp aliases: ", csvr.rpAliases)
	}
}

func TestLoadAccAliases(t *testing.T) {
	if len(csvr.accAliases) != 2 {
		t.Error("Failed to load acc aliases: ", csvr.accAliases)
	}
	if csvr.accAliases["a1"] != "minitsboy" ||
		csvr.accAliases["a2"] != "minitsboy" {
		t.Error("Error loading acc aliases: ", csvr.accAliases)
	}
}

func TestLoadDerivedChargers(t *testing.T) {
	if len(csvr.derivedChargers) != 2 {
		t.Error("Failed to load derivedChargers: ", csvr.derivedChargers)
	}
	expCharger1 := utils.DerivedChargers{
		&utils.DerivedCharger{RunId: "extra1", ReqTypeField: "^prepaid", DirectionField: "*default", TenantField: "*default", TorField: "*default",
			AccountField: "rif", SubjectField: "rif", DestinationField: "*default", SetupTimeField: "*default", AnswerTimeField: "*default", DurationField: "*default"},
		&utils.DerivedCharger{RunId: "extra2", ReqTypeField: "*default", DirectionField: "*default", TenantField: "*default", TorField: "*default",
			AccountField: "ivo", SubjectField: "ivo", DestinationField: "*default", SetupTimeField: "*default", AnswerTimeField: "*default", DurationField: "*default"},
	}
	keyCharger1 := utils.DerivedChargersKey("cgrates.org", "call", "*out", "dan", "dan")
	if !reflect.DeepEqual(csvr.derivedChargers[keyCharger1], expCharger1) {
		t.Error("Unexpected charger", csvr.derivedChargers[keyCharger1])
	}
}
