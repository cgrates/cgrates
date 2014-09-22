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
R1,0,0.2,60,1,0
R2,0,0.1,60s,1s,0
R3,0,0.05,60s,1s,0
R4,1,1,1s,1s,0
R5,0,0.5,1s,1s,0
LANDLINE_OFFPEAK,0,1,1,60,0
LANDLINE_OFFPEAK,0,1,1,1,60
GBP_71,0.000000,5.55555,1s,1s,0s
GBP_72,0.000000,7.77777,1s,1s,0s
GBP_70,0.000000,1,1,1,0
RT_UK_Mobile_BIG5_PKG,0.01,0,20s,20s,0s
RT_UK_Mobile_BIG5,0.01,0.10,1s,1s,0s
R_URG,0,0,1,1,0
`
	destinationRates = `
RT_STANDARD,GERMANY,R1,*middle,4
RT_STANDARD,GERMANY_O2,R2,*middle,4
RT_STANDARD,GERMANY_PREMIUM,R2,*middle,4
RT_DEFAULT,ALL,R2,*middle,4
RT_STD_WEEKEND,GERMANY,R2,*middle,4
RT_STD_WEEKEND,GERMANY_O2,R3,*middle,4
P1,NAT,R4,*middle,4
P2,NAT,R5,*middle,4
T1,NAT,LANDLINE_OFFPEAK,*middle,4
T2,GERMANY,GBP_72,*middle,4
T2,GERMANY_O2,GBP_70,*middle,4
T2,GERMANY_PREMIUM,GBP_71,*middle,4
DR_UK_Mobile_BIG5_PKG,DST_UK_Mobile_BIG5,RT_UK_Mobile_BIG5_PKG,*middle,4
DR_UK_Mobile_BIG5,DST_UK_Mobile_BIG5,RT_UK_Mobile_BIG5,*middle,4
DATA_RATE,*any,LANDLINE_OFFPEAK,*middle,4
RT_URG,URG,R_URG,*middle,4
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

	lcrs = `
*in,cgrates.org,*any,EU_LANDLINE,LCR_STANDARD,*static,ivo;dan;rif,2012-01-01T00:00:00Z,10
*in,cgrates.org,*any,*any,LCR_STANDARD,*lowest_cost,,2012-01-01T00:00:00Z,20
`

	actions = `
MINI,*topup_reset,,*monetary,*out,,,,,*unlimited,10,10,10
MINI,*topup,,*voice,*out,,NAT,test,,*unlimited,100,10,10
SHARED,*topup,,*monetary,*out,,,,SG1,*unlimited,100,10,10
TOPUP10_AC,*topup_reset,,*monetary,*out,,*any,,,*unlimited,1,10,10
TOPUP10_AC1,*topup_reset,,*voice,*out,,DST_UK_Mobile_BIG5,discounted_minutes,,*unlimited,40,10,10
SE0,*topup_reset,,*monetary,*out,,,,SG2,*unlimited,0,10,10
SE10,*topup_reset,,*monetary,*out,,,,SG2,*unlimited,10,5,10
SE10,*topup,,*monetary,*out,,,,,*unlimited,10,10,10
EE0,*topup_reset,,*monetary,*out,,,,SG3,*unlimited,0,10,10
EE0,*allow_negative,,*monetary,*out,,,,,*unlimited,0,10,10
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
STANDARD_TRIGGER,*min_counter,10,false,0,*voice,*out,,GERMANY_O2,,,,,,SOME_1,10
STANDARD_TRIGGER,*max_balance,200,false,0,*voice,*out,,GERMANY,,,,,,SOME_2,10
STANDARD_TRIGGERS,*min_balance,2,false,0,*monetary,*out,,,,,,,,LOG_WARNING,10
STANDARD_TRIGGERS,*max_balance,20,false,0,*monetary,*out,,,,,,,,LOG_WARNING,10
STANDARD_TRIGGERS,*max_counter,5,false,0,*monetary,*out,,FS_USERS,,,,,,LOG_WARNING,10
CDRST1_WARN_ASR,*min_asr,45,true,1h,,,,,,,,,3,CDRST_WARN_HTTP,10
CDRST1_WARN_ACD,*min_acd,10,true,1h,,,,,,,,,5,CDRST_WARN_HTTP,10
CDRST1_WARN_ACC,*max_acc,10,true,10m,,,,,,,,,5,CDRST_WARN_HTTP,10
CDRST2_WARN_ASR,*min_asr,30,true,0,,,,,,,,,5,CDRST_WARN_HTTP,10
CDRST2_WARN_ACD,*min_acd,3,true,0,,,,,,,,,5,CDRST_WARN_HTTP,10
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
#Direction,Tenant,Category,Account,Subject,RunId,RunFilter,ReqTypeField,DirectionField,TenantField,TorField,AccountField,SubjectField,DestinationField,SetupTimeField,AnswerTimeField,UsageField
*out,cgrates.org,call,dan,dan,extra1,^filteredHeader1/filterValue1/,^prepaid,,,,rif,rif,,,,
*out,cgrates.org,call,dan,dan,extra2,,,,,,ivo,ivo,,,,
*out,cgrates.org,call,dan,*any,extra1,,,,,,rif2,rif2,,,,
`
	cdrStats = `
#Id,QueueLength,TimeWindow,Metrics,SetupInterval,TOR,CdrHost,CdrSource,ReqType,Direction,Tenant,Category,Account,Subject,DestinationPrefix,UsageInterval,MediationRunIds,RatedAccount,RatedSubject,CostInterval,Triggers
CDRST1,5,60m,ASR,2014-07-29T15:00:00Z;2014-07-29T16:00:00Z,*voice,87.139.12.167,FS_JSON,rated,*out,cgrates.org,call,dan,dan,49,5m;10m,default,rif,rif,0;2,STANDARD_TRIGGERS
CDRST1,,,ACD,,,,,,,,,,,,,,,,,STANDARD_TRIGGER
CDRST1,,,ACC,,,,,,,,,,,,,,,,,
CDRST2,10,10m,ASR,,,,,,,cgrates.org,call,,,,,,,,,
CDRST2,,,ACD,,,,,,,,,,,,,,,,,
`
)

var csvr *CSVReader

func init() {
	csvr = NewStringCSVReader(dataStorage, accountingStorage, ',', destinations, timings, rates, destinationRates, ratingPlans, ratingProfiles,
		sharedGroups, lcrs, actions, actionTimings, actionTriggers, accountActions, derivedCharges, cdrStats)
	csvr.LoadDestinations()
	csvr.LoadTimings()
	csvr.LoadRates()
	csvr.LoadDestinationRates()
	csvr.LoadRatingPlans()
	csvr.LoadRatingProfiles()
	csvr.LoadSharedGroups()
	csvr.LoadLCRs()
	csvr.LoadActions()
	csvr.LoadActionTimings()
	csvr.LoadActionTriggers()
	csvr.LoadAccountActions()
	csvr.LoadDerivedChargers()
	csvr.LoadCdrStats()
	csvr.WriteToDatabase(false, false)
	dataStorage.CacheRating(nil, nil, nil, nil, nil)
	accountingStorage.CacheAccounting(nil, nil, nil, nil)
}

func TestLoadDestinations(t *testing.T) {
	if len(csvr.destinations) != 11 {
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
	expctRs, err := utils.NewRateSlot(0, 0.2, "60", "1", "0")
	if err != nil {
		t.Error("Error loading rate: ", rate, err.Error())
	} else if !reflect.DeepEqual(rate, expctRs) ||
		rate.RateUnitDuration() != expctRs.RateUnitDuration() ||
		rate.RateIncrementDuration() != expctRs.RateIncrementDuration() ||
		rate.GroupIntervalStartDuration() != expctRs.GroupIntervalStartDuration() {
		t.Error("Error loading rate: ", rate, expctRs)
	}
	rate = csvr.rates["R2"].RateSlots[0]
	if expctRs, err = utils.NewRateSlot(0, 0.1, "60s", "1s", "0"); err != nil {
		t.Error("Error loading rate: ", rate, err.Error())
	} else if !reflect.DeepEqual(rate, expctRs) ||
		rate.RateUnitDuration() != expctRs.RateUnitDuration() ||
		rate.RateIncrementDuration() != expctRs.RateIncrementDuration() ||
		rate.GroupIntervalStartDuration() != expctRs.GroupIntervalStartDuration() {
		t.Error("Error loading rate: ", rate)
	}
	rate = csvr.rates["R3"].RateSlots[0]
	if expctRs, err = utils.NewRateSlot(0, 0.05, "60s", "1s", "0"); err != nil {
		t.Error("Error loading rate: ", rate, err.Error())
	} else if !reflect.DeepEqual(rate, expctRs) ||
		rate.RateUnitDuration() != expctRs.RateUnitDuration() ||
		rate.RateIncrementDuration() != expctRs.RateIncrementDuration() ||
		rate.GroupIntervalStartDuration() != expctRs.GroupIntervalStartDuration() {
		t.Error("Error loading rate: ", rate)
	}
	rate = csvr.rates["R4"].RateSlots[0]
	if expctRs, err = utils.NewRateSlot(1, 1.0, "1s", "1s", "0"); err != nil {
		t.Error("Error loading rate: ", rate, err.Error())
	} else if !reflect.DeepEqual(rate, expctRs) ||
		rate.RateUnitDuration() != expctRs.RateUnitDuration() ||
		rate.RateIncrementDuration() != expctRs.RateIncrementDuration() ||
		rate.GroupIntervalStartDuration() != expctRs.GroupIntervalStartDuration() {
		t.Error("Error loading rate: ", rate)
	}
	rate = csvr.rates["R5"].RateSlots[0]
	if expctRs, err = utils.NewRateSlot(0, 0.5, "1s", "1s", "0"); err != nil {
		t.Error("Error loading rate: ", rate, err.Error())
	} else if !reflect.DeepEqual(rate, expctRs) ||
		rate.RateUnitDuration() != expctRs.RateUnitDuration() ||
		rate.RateIncrementDuration() != expctRs.RateIncrementDuration() ||
		rate.GroupIntervalStartDuration() != expctRs.GroupIntervalStartDuration() {
		t.Error("Error loading rate: ", rate)
	}
	rate = csvr.rates["LANDLINE_OFFPEAK"].RateSlots[0]
	if expctRs, err = utils.NewRateSlot(0, 1, "1", "60", "0"); err != nil {
		t.Error("Error loading rate: ", rate, err.Error())
	} else if !reflect.DeepEqual(rate, expctRs) ||
		rate.RateUnitDuration() != expctRs.RateUnitDuration() ||
		rate.RateIncrementDuration() != expctRs.RateIncrementDuration() ||
		rate.GroupIntervalStartDuration() != expctRs.GroupIntervalStartDuration() {
		t.Error("Error loading rate: ", rate)
	}
	rate = csvr.rates["LANDLINE_OFFPEAK"].RateSlots[1]
	if expctRs, err = utils.NewRateSlot(0, 1, "1", "1", "60"); err != nil {
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
				DestinationId:    "GERMANY",
				Rate:             csvr.rates["R1"],
				RoundingMethod:   utils.ROUNDING_MIDDLE,
				RoundingDecimals: 4,
			},
			&utils.DestinationRate{
				DestinationId:    "GERMANY_O2",
				Rate:             csvr.rates["R2"],
				RoundingMethod:   utils.ROUNDING_MIDDLE,
				RoundingDecimals: 4,
			},
			&utils.DestinationRate{
				DestinationId:    "GERMANY_PREMIUM",
				Rate:             csvr.rates["R2"],
				RoundingMethod:   utils.ROUNDING_MIDDLE,
				RoundingDecimals: 4,
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
				DestinationId:    "ALL",
				Rate:             csvr.rates["R2"],
				RoundingMethod:   utils.ROUNDING_MIDDLE,
				RoundingDecimals: 4,
			},
		},
	}) {
		t.Errorf("Error loading destination rate: %+v", drs.DestinationRates[0])
	}
	drs = csvr.destinationRates["RT_STD_WEEKEND"]
	if !reflect.DeepEqual(drs, &utils.TPDestinationRate{
		DestinationRateId: "RT_STD_WEEKEND",
		DestinationRates: []*utils.DestinationRate{
			&utils.DestinationRate{
				DestinationId:    "GERMANY",
				Rate:             csvr.rates["R2"],
				RoundingMethod:   utils.ROUNDING_MIDDLE,
				RoundingDecimals: 4,
			},
			&utils.DestinationRate{
				DestinationId:    "GERMANY_O2",
				Rate:             csvr.rates["R3"],
				RoundingMethod:   utils.ROUNDING_MIDDLE,
				RoundingDecimals: 4,
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
				DestinationId:    "NAT",
				Rate:             csvr.rates["R4"],
				RoundingMethod:   utils.ROUNDING_MIDDLE,
				RoundingDecimals: 4,
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
				DestinationId:    "NAT",
				Rate:             csvr.rates["R5"],
				RoundingMethod:   utils.ROUNDING_MIDDLE,
				RoundingDecimals: 4,
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
				DestinationId:    "NAT",
				Rate:             csvr.rates["LANDLINE_OFFPEAK"],
				RoundingMethod:   utils.ROUNDING_MIDDLE,
				RoundingDecimals: 4,
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
				DestinationId:    "GERMANY",
				Rate:             csvr.rates["GBP_72"],
				RoundingMethod:   utils.ROUNDING_MIDDLE,
				RoundingDecimals: 4,
			},
			&utils.DestinationRate{
				DestinationId:    "GERMANY_O2",
				Rate:             csvr.rates["GBP_70"],
				RoundingMethod:   utils.ROUNDING_MIDDLE,
				RoundingDecimals: 4,
			},
			&utils.DestinationRate{
				DestinationId:    "GERMANY_PREMIUM",
				Rate:             csvr.rates["GBP_71"],
				RoundingMethod:   utils.ROUNDING_MIDDLE,
				RoundingDecimals: 4,
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
			"822a5aef": &RIRate{
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
				RoundingDecimals: 4,
			},
			"4a25c533": &RIRate{
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
				RoundingDecimals: 4,
			},
			"b05c5f6b": &RIRate{
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
				RoundingDecimals: 4,
			},
			"9f49ef8e": &RIRate{
				ConnectFee: 0,
				Rates: []*Rate{
					&Rate{
						GroupIntervalStart: 0,
						Value:              0,
						RateIncrement:      time.Second,
						RateUnit:           time.Second,
					},
				},
				RoundingMethod:   utils.ROUNDING_MIDDLE,
				RoundingDecimals: 4,
			},
		},
		DestinationRates: map[string]RPRateList{
			"GERMANY": []*RPRate{
				&RPRate{
					Timing: "14ae6e41",
					Rating: "822a5aef",
					Weight: 10,
				},
				&RPRate{
					Timing: "9a6f8e32",
					Rating: "4a25c533",
					Weight: 10,
				},
				&RPRate{
					Timing: "7181e535",
					Rating: "4a25c533",
					Weight: 10,
				},
			},
			"GERMANY_O2": []*RPRate{
				&RPRate{
					Timing: "14ae6e41",
					Rating: "4a25c533",
					Weight: 10,
				},
				&RPRate{
					Timing: "9a6f8e32",
					Rating: "b05c5f6b",
					Weight: 10,
				},
				&RPRate{
					Timing: "7181e535",
					Rating: "b05c5f6b",
					Weight: 10,
				},
			},
			"GERMANY_PREMIUM": []*RPRate{
				&RPRate{
					Timing: "14ae6e41",
					Rating: "4a25c533",
					Weight: 10,
				},
			},
			"URG": []*RPRate{
				&RPRate{
					Timing: "96c78ff5",
					Rating: "9f49ef8e",
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
	if !reflect.DeepEqual(as1[1], expected[1]) {
		t.Errorf("Error loading action1: %+v", as1[0].Balance)
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
		t.Error("Failed to shared groups: ", csvr.sharedGroups)
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

func TestLoadLCRs(t *testing.T) {
	if len(csvr.lcrs) != 1 {
		t.Error("Failed to load LCRs: ", csvr.lcrs)
	}

	lcr := csvr.lcrs["*in:cgrates.org:*any"]
	expected := &LCR{
		Direction: "*in",
		Tenant:    "cgrates.org",
		Customer:  "*any",
		Activations: []*LCRActivation{
			&LCRActivation{
				ActivationTime: time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC),
				Entries: []*LCREntry{
					&LCREntry{
						DestinationId: "EU_LANDLINE",
						Category:      "LCR_STANDARD",
						Strategy:      "*static",
						Suppliers:     "ivo;dan;rif",
						Weight:        10,
					},
					&LCREntry{
						DestinationId: "*any",
						Category:      "LCR_STANDARD",
						Strategy:      "*lowest_cost",
						Suppliers:     "",
						Weight:        20,
					},
				},
			},
		},
	}
	if !reflect.DeepEqual(lcr, expected) {
		t.Errorf("Error loading lcr %+v: ", lcr.Activations)
	}
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
	if len(csvr.actionsTriggers) != 7 {
		t.Error("Failed to load action triggers: ", len(csvr.actionsTriggers))
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
	if csvr.rpAliases[utils.RatingSubjectAliasKey("vdf", "a1")] != "minu" ||
		csvr.rpAliases[utils.RatingSubjectAliasKey("vdf", "a2")] != "minu" ||
		csvr.rpAliases[utils.RatingSubjectAliasKey("vdf", "a3")] != "minu" {
		t.Error("Error loading rp aliases: ", csvr.rpAliases)
	}
}

func TestLoadAccAliases(t *testing.T) {
	if len(csvr.accAliases) != 2 {
		t.Error("Failed to load acc aliases: ", csvr.accAliases)
	}
	if csvr.accAliases[utils.AccountAliasKey("vdf", "a1")] != "minitsboy" ||
		csvr.accAliases[utils.AccountAliasKey("vdf", "a2")] != "minitsboy" {
		t.Error("Error loading acc aliases: ", csvr.accAliases)
	}
}

func TestLoadDerivedChargers(t *testing.T) {
	if len(csvr.derivedChargers) != 2 {
		t.Error("Failed to load derivedChargers: ", csvr.derivedChargers)
	}
	expCharger1 := utils.DerivedChargers{
		&utils.DerivedCharger{RunId: "extra1", RunFilters: "^filteredHeader1/filterValue1/", ReqTypeField: "^prepaid", DirectionField: "*default",
			TenantField: "*default", CategoryField: "*default", AccountField: "rif", SubjectField: "rif", DestinationField: "*default",
			SetupTimeField: "*default", AnswerTimeField: "*default", UsageField: "*default"},
		&utils.DerivedCharger{RunId: "extra2", ReqTypeField: "*default", DirectionField: "*default", TenantField: "*default", CategoryField: "*default",
			AccountField: "ivo", SubjectField: "ivo", DestinationField: "*default", SetupTimeField: "*default", AnswerTimeField: "*default", UsageField: "*default"},
	}
	keyCharger1 := utils.DerivedChargersKey("*out", "cgrates.org", "call", "dan", "dan")
	if !reflect.DeepEqual(csvr.derivedChargers[keyCharger1], expCharger1) {
		t.Error("Unexpected charger", csvr.derivedChargers[keyCharger1])
	}
}
func TestLoadCdrStats(t *testing.T) {
	if len(csvr.cdrStats) != 2 {
		t.Error("Failed to load cdr stats: ", csvr.cdrStats)
	}
	cdrStats1 := &CdrStats{
		Id:          "CDRST1",
		QueueLength: 5,
		TimeWindow:  60 * time.Minute,
		Metrics:     []string{"ASR", "ACD", "ACC"},
		SetupInterval: []time.Time{
			time.Date(2014, 7, 29, 15, 0, 0, 0, time.UTC),
			time.Date(2014, 7, 29, 16, 0, 0, 0, time.UTC),
		},
		TOR:               []string{utils.VOICE},
		CdrHost:           []string{"87.139.12.167"},
		CdrSource:         []string{"FS_JSON"},
		ReqType:           []string{"rated"},
		Direction:         []string{utils.OUT},
		Tenant:            []string{"cgrates.org"},
		Category:          []string{"call"},
		Account:           []string{"dan"},
		Subject:           []string{"dan"},
		DestinationPrefix: []string{"49"},
		UsageInterval:     []time.Duration{5 * time.Minute, 10 * time.Minute},
		MediationRunIds:   []string{"default"},
		RatedAccount:      []string{"rif"},
		RatedSubject:      []string{"rif"},
		CostInterval:      []float64{0, 2},
	}
	cdrStats1.Triggers = append(cdrStats1.Triggers, csvr.actionsTriggers["STANDARD_TRIGGERS"]...)
	cdrStats1.Triggers = append(cdrStats1.Triggers, csvr.actionsTriggers["STANDARD_TRIGGER"]...)
	if !reflect.DeepEqual(csvr.cdrStats[cdrStats1.Id], cdrStats1) {
		t.Error("Unexpected stats", csvr.cdrStats[cdrStats1.Id])
	}
}
