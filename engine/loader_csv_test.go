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
	"log"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

const (
	testTPID               = "LoaderCSVTests"
	DestinationsCSVContent = `
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
NAT,+49
RET,0723
RET,0724
SPEC,0723045
PSTN_71,+4971
PSTN_72,+4972
PSTN_70,+4970
DST_UK_Mobile_BIG5,447956
URG,112
EU_LANDLINE,444
EXOTIC,999
`
	TimingsCSVContent = `
WORKDAYS_00,*any,*any,*any,1;2;3;4;5,00:00:00
WORKDAYS_18,*any,*any,*any,1;2;3;4;5,18:00:00
WEEKENDS,*any,*any,*any,6;7,00:00:00
ONE_TIME_RUN,2012,,,,*asap
`
	RatesCSVContent = `
R1,0,0.2,60s,1s,0s
R2,0,0.1,60s,1s,0s
R3,0,0.05,60s,1s,0s
R4,1,1,1s,1s,0s
R5,0,0.5,1s,1s,0s
LANDLINE_OFFPEAK,0,1,1s,60s,0s
LANDLINE_OFFPEAK,0,1,1s,1s,60s
GBP_71,0.000000,5.55555,1s,1s,0s
GBP_72,0.000000,7.77777,1s,1s,0s
GBP_70,0.000000,1,1s,1s,0s
RT_UK_Mobile_BIG5_PKG,0.01,0,20s,20s,0s
RT_UK_Mobile_BIG5,0.01,0.10,1s,1s,0s
R_URG,0,0,1s,1s,0s
MX,0,1,1s,1s,0s
DY,0.15,0.05,60s,1s,0s
CF,1.12,0,1s,1s,0s
`
	DestinationRatesCSVContent = `
RT_STANDARD,GERMANY,R1,*middle,4,0,
RT_STANDARD,GERMANY_O2,R2,*middle,4,0,
RT_STANDARD,GERMANY_PREMIUM,R2,*middle,4,0,
RT_DEFAULT,ALL,R2,*middle,4,0,
RT_STD_WEEKEND,GERMANY,R2,*middle,4,0,
RT_STD_WEEKEND,GERMANY_O2,R3,*middle,4,0,
P1,NAT,R4,*middle,4,0,
P2,NAT,R5,*middle,4,0,
T1,NAT,LANDLINE_OFFPEAK,*middle,4,0,
T2,GERMANY,GBP_72,*middle,4,0,
T2,GERMANY_O2,GBP_70,*middle,4,0,
T2,GERMANY_PREMIUM,GBP_71,*middle,4,0,
GER,GERMANY,R4,*middle,4,0,
DR_UK_Mobile_BIG5_PKG,DST_UK_Mobile_BIG5,RT_UK_Mobile_BIG5_PKG,*middle,4,,
DR_UK_Mobile_BIG5,DST_UK_Mobile_BIG5,RT_UK_Mobile_BIG5,*middle,4,,
DATA_RATE,*any,LANDLINE_OFFPEAK,*middle,4,0,
RT_URG,URG,R_URG,*middle,4,0,
MX_FREE,RET,MX,*middle,4,10,*free
MX_DISC,RET,MX,*middle,4,10,*disconnect
RT_DY,RET,DY,*up,2,0,
RT_DY,EU_LANDLINE,CF,*middle,4,0,
`
	RatingPlansCSVContent = `
STANDARD,RT_STANDARD,WORKDAYS_00,10
STANDARD,RT_STD_WEEKEND,WORKDAYS_18,10
STANDARD,RT_STD_WEEKEND,WEEKENDS,10
STANDARD,RT_URG,*any,20
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
RP_UK_Mobile_BIG5_PKG,DR_UK_Mobile_BIG5_PKG,*any,10
RP_UK,DR_UK_Mobile_BIG5,*any,10
RP_DATA,DATA_RATE,*any,10
RP_MX,MX_DISC,WORKDAYS_00,10
RP_MX,MX_FREE,WORKDAYS_18,10
GER_ONLY,GER,*any,10
ANY_PLAN,DATA_RATE,*any,10
DY_PLAN,RT_DY,*any,10
`
	RatingProfilesCSVContent = `
CUSTOMER_1,0,rif:from:tm,2012-01-01T00:00:00Z,PREMIUM,danb
CUSTOMER_1,0,rif:from:tm,2012-02-28T00:00:00Z,STANDARD,danb
CUSTOMER_2,0,danb:87.139.12.167,2012-01-01T00:00:00Z,STANDARD,danb
CUSTOMER_1,0,danb,2012-01-01T00:00:00Z,PREMIUM,
vdf,0,rif,2012-01-01T00:00:00Z,EVENING,
vdf,call,rif,2012-02-28T00:00:00Z,EVENING,
vdf,call,dan,2012-01-01T00:00:00Z,EVENING,
vdf,0,minu,2012-01-01T00:00:00Z,EVENING,
vdf,0,*any,2012-02-28T00:00:00Z,EVENING,
vdf,0,one,2012-02-28T00:00:00Z,STANDARD,
vdf,0,inf,2012-02-28T00:00:00Z,STANDARD,inf
vdf,0,fall,2012-02-28T00:00:00Z,PREMIUM,rif
test,0,trp,2013-10-01T00:00:00Z,TDRT,rif;danb
vdf,0,fallback1,2013-11-18T13:45:00Z,G,fallback2
vdf,0,fallback1,2013-11-18T13:46:00Z,G,fallback2
vdf,0,fallback1,2013-11-18T13:47:00Z,G,fallback2
vdf,0,fallback2,2013-11-18T13:45:00Z,R,rif
cgrates.org,call,*any,2013-01-06T00:00:00Z,RP_UK,
cgrates.org,call,discounted_minutes,2013-01-06T00:00:00Z,RP_UK_Mobile_BIG5_PKG,
cgrates.org,data,rif,2013-01-06T00:00:00Z,RP_DATA,
cgrates.org,call,max,2013-03-23T00:00:00Z,RP_MX,
cgrates.org,call,nt,2012-02-28T00:00:00Z,GER_ONLY,
cgrates.org,LCR_STANDARD,max,2013-03-23T00:00:00Z,RP_MX,
cgrates.org,call,money,2015-02-28T00:00:00Z,EVENING,
cgrates.org,call,dy,2015-02-28T00:00:00Z,DY_PLAN,
cgrates.org,call,block,2015-02-28T00:00:00Z,DY_PLAN,
cgrates.org,call,round,2016-06-30T00:00:00Z,DEFAULT,
`
	SharedGroupsCSVContent = `
SG1,*any,*lowest,
SG2,*any,*lowest,one
SG3,*any,*lowest,
`
	ActionsCSVContent = `
MINI,*topup_reset,,,,*monetary,,,,,*unlimited,,10,10,false,false,10
MINI,*topup,,,,*voice,,NAT,test,,*unlimited,,100s,10,false,false,10
SHARED,*topup,,,,*monetary,,,,SG1,*unlimited,,100,10,false,false,10
TOPUP10_AC,*topup_reset,,,,*monetary,,*any,,,*unlimited,,1,10,false,false,10
TOPUP10_AC1,*topup_reset,,,,*voice,,DST_UK_Mobile_BIG5,discounted_minutes,,*unlimited,,40s,10,false,false,10
SE0,*topup_reset,,,,*monetary,,,,SG2,*unlimited,,0,10,false,false,10
SE10,*topup_reset,,,,*monetary,,,,SG2,*unlimited,,10,5,false,false,10
SE10,*topup,,,,*monetary,,,,,*unlimited,,10,10,false,false,10
EE0,*topup_reset,,,,*monetary,,,,SG3,*unlimited,,0,10,false,false,10
EE0,*allow_negative,,,,*monetary,,,,,*unlimited,,0,10,false,false,10
DEFEE,*cdrlog,"{""Category"":""^ddi"",""MediationRunId"":""^did_run""}",,,,,,,,,,,,false,false,10
NEG,*allow_negative,,,,*monetary,,,,,*unlimited,,0,10,false,false,10
BLOCK,*topup,,,bblocker,*monetary,,NAT,,,*unlimited,,1,20,true,false,20
BLOCK,*topup,,,bfree,*monetary,,,,,*unlimited,,20,10,false,false,10
BLOCK_EMPTY,*topup,,,bblocker,*monetary,,NAT,,,*unlimited,,0,20,true,false,20
BLOCK_EMPTY,*topup,,,bfree,*monetary,,,,,*unlimited,,20,10,false,false,10
FILTER,*topup,,*string:~*req.BalanceMap.*monetary[0].ID:*default;*lt:~*req.BalanceMap.*monetary[0].Value:0,bfree,*monetary,,,,,*unlimited,,20,10,false,false,10
EXP,*topup,,,,*voice,,,,,*monthly,*any,300s,10,false,false,10
NOEXP,*topup,,,,*voice,,,,,*unlimited,*any,50s,10,false,false,10
VF,*debit,,,,*monetary,,,,,*unlimited,*any,"{""Method"":""*incremental"",""Params"":{""Units"":10, ""Interval"":""month"", ""Increment"":""day""}}",10,false,false,10
TOPUP_RST_GNR_1000,*topup_reset,"{""*voice"": 60.0,""*data"":1024.0,""*sms"":1.0}",,,*generic,,*any,,,*unlimited,,1000,20,false,false,10
`
	ActionPlansCSVContent = `
MORE_MINUTES,MINI,ONE_TIME_RUN,10
MORE_MINUTES,SHARED,ONE_TIME_RUN,10
TOPUP10_AT,TOPUP10_AC,*asap,10
TOPUP10_AT,TOPUP10_AC1,*asap,10
TOPUP_SHARED0_AT,SE0,*asap,10
TOPUP_SHARED10_AT,SE10,*asap,10
TOPUP_EMPTY_AT,EE0,*asap,10
POST_AT,NEG,*asap,10
BLOCK_AT,BLOCK,*asap,10
BLOCK_EMPTY_AT,BLOCK_EMPTY,*asap,10
EXP_AT,EXP,*asap,10
`

	ActionTriggersCSVContent = `
STANDARD_TRIGGER,st0,*min_event_counter,10,false,0,,,,*voice,,GERMANY_O2,,,,,,,,SOME_1,10
STANDARD_TRIGGER,st1,*max_balance,200,false,0,,,,*voice,,GERMANY,,,,,,,,SOME_2,10
STANDARD_TRIGGERS,,*min_balance,2,false,0,,,,*monetary,,,,,,,,,,LOG_WARNING,10
STANDARD_TRIGGERS,,*max_balance,20,false,0,,,,*monetary,,,,,,,,,,LOG_WARNING,10
STANDARD_TRIGGERS,,*max_event_counter,5,false,0,,,,*monetary,,FS_USERS,,,,,,,,LOG_WARNING,10
`
	AccountActionsCSVContent = `
vdf,minitsboy,MORE_MINUTES,STANDARD_TRIGGER,,
cgrates.org,12345,TOPUP10_AT,STANDARD_TRIGGERS,,
cgrates.org,123456,TOPUP10_AT,STANDARD_TRIGGERS,,
cgrates.org,dy,TOPUP10_AT,STANDARD_TRIGGERS,,
cgrates.org,remo,TOPUP10_AT,,,
vdf,empty0,TOPUP_SHARED0_AT,,,
vdf,empty10,TOPUP_SHARED10_AT,,,
vdf,emptyX,TOPUP_EMPTY_AT,,,
vdf,emptyY,TOPUP_EMPTY_AT,,,
vdf,post,POST_AT,,,
cgrates.org,alodis,TOPUP_EMPTY_AT,,true,true
cgrates.org,block,BLOCK_AT,,false,false
cgrates.org,block_empty,BLOCK_EMPTY_AT,,false,false
cgrates.org,expo,EXP_AT,,false,false
cgrates.org,expnoexp,,,false,false
cgrates.org,vf,,,false,false
cgrates.org,round,TOPUP10_AT,,false,false
`
	ResourcesCSVContent = `
#Tenant[0],Id[1],FilterIDs[2],ActivationInterval[3],TTL[4],Limit[5],AllocationMessage[6],Blocker[7],Stored[8],Weight[9],Thresholds[10]
cgrates.org,ResGroup21,*string:~*req.Account:1001,2014-07-29T15:00:00Z,1s,2,call,true,true,10,
cgrates.org,ResGroup22,*string:~*req.Account:dan,2014-07-29T15:00:00Z,3600s,2,premium_call,true,true,10,
`
	StatsCSVContent = `
#Tenant[0],Id[1],FilterIDs[2],ActivationInterval[3],QueueLength[4],TTL[5],MinItems[6],Metrics[7],MetricFilterIDs[8],Stored[9],Blocker[10],Weight[11],ThresholdIDs[12]
cgrates.org,TestStats,*string:~*req.Account:1001,2014-07-29T15:00:00Z,100,1s,2,*sum#~*req.Value;*average#~*req.Value,,true,true,20,Th1;Th2
cgrates.org,TestStats,,,,,2,*sum#~*req.Usage,,true,true,20,
cgrates.org,TestStats2,FLTR_1,2014-07-29T15:00:00Z,100,1s,2,*sum#~*req.Value;*sum#~*req.Usage;*average#~*req.Value;*average#~*req.Usage,,true,true,20,Th
cgrates.org,TestStats2,,,,,2,*sum#~*req.Cost;*average#~*req.Cost,,true,true,20,
`
	RankingsCSVContent = `
#Tenant[0],Id[1],Schedule[2],StatIDs[3],MetricIDs[4],Sorting[5],SortingParameters[6],StoredThresholdIDs[7]
cgrates.org,Ranking1,@every 5m,Stats2;Stats3;Stats4,Metric1;Metric3,*asc,,true,THD1;THD2
`
	TrendsCSVContent = `
#Tenant[0],Id[1],Schedule[2],StatID[3],Metrics[4],TTL[5],QueueLength[6],MinItems[7],CorrelationType[8],Tolerance[9],Stored[10],ThresholdIDs[11]
cgrates.org,TREND1,0 12 * * *,Stats2,*acc;*tcc,-1,-1,1,*average,2.1,true,TD1;TD2
`
	ThresholdsCSVContent = `
#Tenant[0],Id[1],FilterIDs[2],ActivationInterval[3],MaxHits[4],MinHits[5],MinSleep[6],Blocker[7],Weight[8],ActionIDs[9],Async[10]
cgrates.org,Threshold1,*string:~*req.Account:1001;*string:~*req.RunID:*default,2014-07-29T15:00:00Z,12,10,1s,true,10,THRESH1,true
`

	FiltersCSVContent = `
#Tenant[0],ID[1],Type[2],Element[3],Values[4],ActivationInterval[5]
cgrates.org,FLTR_1,*string,~*req.Account,1001;1002,2014-07-29T15:00:00Z
cgrates.org,FLTR_1,*prefix,~*req.Destination,10;20,2014-07-29T15:00:00Z
cgrates.org,FLTR_1,*rsr,~*req.Subject,~^1.*1$,
cgrates.org,FLTR_1,*rsr,~*req.Destination,1002,
cgrates.org,FLTR_ACNT_dan,*string,~*req.Account,dan,2014-07-29T15:00:00Z
cgrates.org,FLTR_DST_DE,*destinations,~*req.Destination,DST_DE,2014-07-29T15:00:00Z
cgrates.org,FLTR_DST_NL,*destinations,~*req.Destination,DST_NL,2014-07-29T15:00:00Z
`
	RoutesCSVContent = `
#Tenant[0],ID[1],FilterIDs[2],ActivationInterval[3],Sorting[4],SortingParameters[5],RouteID[6],RouteFilterIDs[7],RouteAccountIDs[8],RouteRatingPlanIDs[9],RouteResourceIDs[10],RouteStatIDs[11],RouteWeight[12],RouteBlocker[13],RouteParameters[14],Weight[15]
cgrates.org,RoutePrf1,*string:~*req.Account:dan,2014-07-29T15:00:00Z,*lc,,route1,FLTR_ACNT_dan,Account1;Account1_1,RPL_1,ResGroup1,Stat1,10,true,param1,20
cgrates.org,RoutePrf1,,,,,route1,,,RPL_2,ResGroup2,,10,,,
cgrates.org,RoutePrf1,,,,,route1,FLTR_DST_DE,Account2,RPL_3,ResGroup3,Stat2,10,,,
cgrates.org,RoutePrf1,,,,,route1,,,,ResGroup4,Stat3,10,,,
`
	AttributesCSVContent = `
#Tenant,ID,Contexts,FilterIDs,ActivationInterval,AttributeFilterIDs,Path,Type,Value,Blocker,Weight
cgrates.org,ALS1,con1,*string:~*req.Account:1001,2014-07-29T15:00:00Z,*string:~*req.Field1:Initial,*req.Field1,*variable,Sub1,true,20
cgrates.org,ALS1,con2;con3,,,,*req.Field2,*variable,Sub2,true,20
`
	ChargersCSVContent = `
#Tenant,ID,FilterIDs,ActivationInterval,RunID,AttributeIDs,Weight
cgrates.org,Charger1,*string:~*req.Account:1001,2014-07-29T15:00:00Z,*rated,ATTR_1001_SIMPLEAUTH,20
`
	DispatcherCSVContent = `
#Tenant,ID,FilterIDs,ActivationInterval,Strategy,Hosts,Weight
cgrates.org,D1,*any,*string:~*req.Account:1001,2014-07-29T15:00:00Z,*first,,C1,*gt:~*req.Usage:10,10,false,192.168.56.203,20
cgrates.org,D1,,,,*first,,C2,*lt:~*req.Usage:10,10,false,192.168.56.204,
`
	DispatcherHostCSVContent = `
#Tenant[0],ID[1],Address[2],Transport[3],ConnectAttempts[4],Reconnects[5],MaxReconnectInterval[6],ConnectTimeout[7],ReplyTimeout[8],Tls[9],ClientKey[10],ClientCertificate[11],CaCertificate[12]
cgrates.org,ALL,127.0.0.1:6012,*json,1,3,5m,1m,2m,false,,,
`
)

var csvr *TpReader

func init() {
	var err error
	csvr, err = NewTpReader(dm.dataDB, NewStringCSVStorage(utils.CSVSep,
		DestinationsCSVContent, TimingsCSVContent, RatesCSVContent, DestinationRatesCSVContent,
		RatingPlansCSVContent, RatingProfilesCSVContent, SharedGroupsCSVContent,
		ActionsCSVContent, ActionPlansCSVContent, ActionTriggersCSVContent, AccountActionsCSVContent,
		ResourcesCSVContent, StatsCSVContent, TrendsCSVContent, RankingsCSVContent, ThresholdsCSVContent, FiltersCSVContent,
		RoutesCSVContent, AttributesCSVContent, ChargersCSVContent, DispatcherCSVContent,
		DispatcherHostCSVContent), testTPID, "", nil, nil, false)
	if err != nil {
		log.Print("error when creating TpReader:", err)
	}
	if err := csvr.LoadDestinations(); err != nil {
		log.Print("error in LoadDestinations:", err)
	}
	if err := csvr.LoadTimings(); err != nil {
		log.Print("error in LoadTimings:", err)
	}
	if err := csvr.LoadRates(); err != nil {
		log.Print("error in LoadRates:", err)
	}
	if err := csvr.LoadDestinationRates(); err != nil {
		log.Print("error in LoadDestRates:", err)
	}
	if err := csvr.LoadRatingPlans(); err != nil {
		log.Print("error in LoadRatingPlans:", err)
	}
	if err := csvr.LoadRatingProfiles(); err != nil {
		log.Print("error in LoadRatingProfiles:", err)
	}
	if err := csvr.LoadSharedGroups(); err != nil {
		log.Print("error in LoadSharedGroups:", err)
	}
	if err := csvr.LoadActions(); err != nil {
		log.Print("error in LoadActions:", err)
	}
	if err := csvr.LoadActionPlans(); err != nil {
		log.Print("error in LoadActionPlans:", err)
	}
	if err := csvr.LoadActionTriggers(); err != nil {
		log.Print("error in LoadActionTriggers:", err)
	}
	if err := csvr.LoadAccountActions(); err != nil {
		log.Print("error in LoadAccountActions:", err)
	}
	if err := csvr.LoadFilters(); err != nil {
		log.Print("error in LoadFilter:", err)
	}
	if err := csvr.LoadResourceProfiles(); err != nil {
		log.Print("error in LoadResourceProfiles:", err)
	}
	if err := csvr.LoadStats(); err != nil {
		log.Print("error in LoadStats:", err)
	}
	if err := csvr.LoadRankings(); err != nil {
		log.Print("error in LoadRankings:", err)
	}
	if err := csvr.LoadTrends(); err != nil {
		log.Print("error in LoadTrends:", err)
	}
	if err := csvr.LoadThresholds(); err != nil {
		log.Print("error in LoadThresholds:", err)
	}
	if err := csvr.LoadRouteProfiles(); err != nil {
		log.Print("error in LoadRouteProfiles:", err)
	}
	if err := csvr.LoadAttributeProfiles(); err != nil {
		log.Print("error in LoadAttributeProfiles:", err)
	}
	if err := csvr.LoadChargerProfiles(); err != nil {
		log.Print("error in LoadChargerProfiles:", err)
	}
	if err := csvr.LoadDispatcherProfiles(); err != nil {
		log.Print("error in LoadDispatcherProfiles:", err)
	}
	if err := csvr.LoadDispatcherHosts(); err != nil {
		log.Print("error in LoadDispatcherHosts:", err)
	}
	if err := csvr.WriteToDatabase(false, false); err != nil {
		log.Print("error when writing into database ", err)
	}
}

func TestLoadDestinations(t *testing.T) {
	if len(csvr.destinations) != 14 {
		t.Error("Failed to load destinations: ", len(csvr.destinations))
	}
	for _, d := range csvr.destinations {
		switch d.Id {
		case "NAT":
			if !reflect.DeepEqual(d.Prefixes, []string{`0256`, `0257`, `0723`, `+49`}) {
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
	if len(csvr.timings) != 14 {
		t.Error("Failed to load timings: ", csvr.timings)
	}
	timing := csvr.timings["WORKDAYS_00"]
	if !reflect.DeepEqual(timing, &utils.TPTiming{
		ID:        "WORKDAYS_00",
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
		ID:        "WORKDAYS_18",
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
		ID:        "WEEKENDS",
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
		ID:        "ONE_TIME_RUN",
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
	if len(csvr.rates) != 15 {
		t.Error("Failed to load rates: ", len(csvr.rates))
	}
	rate := csvr.rates["R1"].RateSlots[0]
	expctRs, err := utils.NewRateSlot(0, 0.2, "60s", "1s", "0s")
	if err != nil {
		t.Error("Error loading rate: ", rate, err.Error())
	} else if !reflect.DeepEqual(rate, expctRs) ||
		rate.RateUnitDuration() != expctRs.RateUnitDuration() ||
		rate.RateIncrementDuration() != expctRs.RateIncrementDuration() ||
		rate.GroupIntervalStartDuration() != expctRs.GroupIntervalStartDuration() {
		t.Error("Error loading rate: ", rate, expctRs)
	}
	rate = csvr.rates["R2"].RateSlots[0]
	if expctRs, err = utils.NewRateSlot(0, 0.1, "60s", "1s", "0s"); err != nil {
		t.Error("Error loading rate: ", rate, err.Error())
	} else if !reflect.DeepEqual(rate, expctRs) ||
		rate.RateUnitDuration() != expctRs.RateUnitDuration() ||
		rate.RateIncrementDuration() != expctRs.RateIncrementDuration() ||
		rate.GroupIntervalStartDuration() != expctRs.GroupIntervalStartDuration() {
		t.Errorf("Expecting: %+v, received: %+v", expctRs, rate)
	}
	rate = csvr.rates["R3"].RateSlots[0]
	if expctRs, err = utils.NewRateSlot(0, 0.05, "60s", "1s", "0s"); err != nil {
		t.Error("Error loading rate: ", rate, err.Error())
	} else if !reflect.DeepEqual(rate, expctRs) ||
		rate.RateUnitDuration() != expctRs.RateUnitDuration() ||
		rate.RateIncrementDuration() != expctRs.RateIncrementDuration() ||
		rate.GroupIntervalStartDuration() != expctRs.GroupIntervalStartDuration() {
		t.Error("Error loading rate: ", rate)
	}
	rate = csvr.rates["R4"].RateSlots[0]
	if expctRs, err = utils.NewRateSlot(1, 1.0, "1s", "1s", "0s"); err != nil {
		t.Error("Error loading rate: ", rate, err.Error())
	} else if !reflect.DeepEqual(rate, expctRs) ||
		rate.RateUnitDuration() != expctRs.RateUnitDuration() ||
		rate.RateIncrementDuration() != expctRs.RateIncrementDuration() ||
		rate.GroupIntervalStartDuration() != expctRs.GroupIntervalStartDuration() {
		t.Error("Error loading rate: ", rate)
	}
	rate = csvr.rates["R5"].RateSlots[0]
	if expctRs, err = utils.NewRateSlot(0, 0.5, "1s", "1s", "0s"); err != nil {
		t.Error("Error loading rate: ", rate, err.Error())
	} else if !reflect.DeepEqual(rate, expctRs) ||
		rate.RateUnitDuration() != expctRs.RateUnitDuration() ||
		rate.RateIncrementDuration() != expctRs.RateIncrementDuration() ||
		rate.GroupIntervalStartDuration() != expctRs.GroupIntervalStartDuration() {
		t.Error("Error loading rate: ", rate)
	}
	rate = csvr.rates["LANDLINE_OFFPEAK"].RateSlots[0]
	if expctRs, err = utils.NewRateSlot(0, 1, "1s", "60s", "0s"); err != nil {
		t.Error("Error loading rate: ", rate, err.Error())
	} else if !reflect.DeepEqual(rate, expctRs) ||
		rate.RateUnitDuration() != expctRs.RateUnitDuration() ||
		rate.RateIncrementDuration() != expctRs.RateIncrementDuration() ||
		rate.GroupIntervalStartDuration() != expctRs.GroupIntervalStartDuration() {
		t.Error("Error loading rate: ", rate)
	}
	rate = csvr.rates["LANDLINE_OFFPEAK"].RateSlots[1]
	if expctRs, err = utils.NewRateSlot(0, 1, "1s", "1s", "60s"); err != nil {
		t.Error("Error loading rate: ", rate, err.Error())
	} else if !reflect.DeepEqual(rate, expctRs) ||
		rate.RateUnitDuration() != expctRs.RateUnitDuration() ||
		rate.RateIncrementDuration() != expctRs.RateIncrementDuration() ||
		rate.GroupIntervalStartDuration() != expctRs.GroupIntervalStartDuration() {
		t.Error("Error loading rate: ", rate)
	}
}

func TestLoadDestinationRates(t *testing.T) {
	if len(csvr.destinationRates) != 15 {
		t.Error("Failed to load destinationrates: ", len(csvr.destinationRates))
	}
	drs := csvr.destinationRates["RT_STANDARD"]
	dr := &utils.TPDestinationRate{
		TPid: testTPID,
		ID:   "RT_STANDARD",
		DestinationRates: []*utils.DestinationRate{
			{
				DestinationId:    "GERMANY",
				RateId:           "R1",
				Rate:             csvr.rates["R1"],
				RoundingMethod:   utils.MetaRoundingMiddle,
				RoundingDecimals: 4,
			},
			{
				DestinationId:    "GERMANY_O2",
				RateId:           "R2",
				Rate:             csvr.rates["R2"],
				RoundingMethod:   utils.MetaRoundingMiddle,
				RoundingDecimals: 4,
			},
			{
				DestinationId:    "GERMANY_PREMIUM",
				RateId:           "R2",
				Rate:             csvr.rates["R2"],
				RoundingMethod:   utils.MetaRoundingMiddle,
				RoundingDecimals: 4,
			},
		},
	}
	if !reflect.DeepEqual(drs, dr) {
		t.Errorf("Error loading destination rate: \n%+v \n%+v", drs.DestinationRates[0], dr.DestinationRates[0])
	}
	drs = csvr.destinationRates["RT_DEFAULT"]
	if !reflect.DeepEqual(drs, &utils.TPDestinationRate{
		TPid: testTPID,
		ID:   "RT_DEFAULT",
		DestinationRates: []*utils.DestinationRate{
			{
				DestinationId:    "ALL",
				RateId:           "R2",
				Rate:             csvr.rates["R2"],
				RoundingMethod:   utils.MetaRoundingMiddle,
				RoundingDecimals: 4,
			},
		},
	}) {
		t.Errorf("Error loading destination rate: %+v", drs.DestinationRates[0])
	}
	drs = csvr.destinationRates["RT_STD_WEEKEND"]
	if !reflect.DeepEqual(drs, &utils.TPDestinationRate{
		TPid: testTPID,
		ID:   "RT_STD_WEEKEND",
		DestinationRates: []*utils.DestinationRate{
			{
				DestinationId:    "GERMANY",
				RateId:           "R2",
				Rate:             csvr.rates["R2"],
				RoundingMethod:   utils.MetaRoundingMiddle,
				RoundingDecimals: 4,
			},
			{
				DestinationId:    "GERMANY_O2",
				RateId:           "R3",
				Rate:             csvr.rates["R3"],
				RoundingMethod:   utils.MetaRoundingMiddle,
				RoundingDecimals: 4,
			},
		},
	}) {
		t.Error("Error loading destination rate: ", drs)
	}
	drs = csvr.destinationRates["P1"]
	if !reflect.DeepEqual(drs, &utils.TPDestinationRate{
		TPid: testTPID,
		ID:   "P1",
		DestinationRates: []*utils.DestinationRate{
			{
				DestinationId:    "NAT",
				RateId:           "R4",
				Rate:             csvr.rates["R4"],
				RoundingMethod:   utils.MetaRoundingMiddle,
				RoundingDecimals: 4,
			},
		},
	}) {
		t.Error("Error loading destination rate: ", drs)
	}
	drs = csvr.destinationRates["P2"]
	if !reflect.DeepEqual(drs, &utils.TPDestinationRate{
		TPid: testTPID,
		ID:   "P2",
		DestinationRates: []*utils.DestinationRate{
			{
				DestinationId:    "NAT",
				RateId:           "R5",
				Rate:             csvr.rates["R5"],
				RoundingMethod:   utils.MetaRoundingMiddle,
				RoundingDecimals: 4,
			},
		},
	}) {
		t.Error("Error loading destination rate: ", drs)
	}
	drs = csvr.destinationRates["T1"]
	if !reflect.DeepEqual(drs, &utils.TPDestinationRate{
		TPid: testTPID,
		ID:   "T1",
		DestinationRates: []*utils.DestinationRate{
			{
				DestinationId:    "NAT",
				RateId:           "LANDLINE_OFFPEAK",
				Rate:             csvr.rates["LANDLINE_OFFPEAK"],
				RoundingMethod:   utils.MetaRoundingMiddle,
				RoundingDecimals: 4,
			},
		},
	}) {
		t.Error("Error loading destination rate: ", drs)
	}
	drs = csvr.destinationRates["T2"]
	if !reflect.DeepEqual(drs, &utils.TPDestinationRate{
		TPid: testTPID,
		ID:   "T2",
		DestinationRates: []*utils.DestinationRate{
			{
				DestinationId:    "GERMANY",
				RateId:           "GBP_72",
				Rate:             csvr.rates["GBP_72"],
				RoundingMethod:   utils.MetaRoundingMiddle,
				RoundingDecimals: 4,
			},
			{
				DestinationId:    "GERMANY_O2",
				RateId:           "GBP_70",
				Rate:             csvr.rates["GBP_70"],
				RoundingMethod:   utils.MetaRoundingMiddle,
				RoundingDecimals: 4,
			},
			{
				DestinationId:    "GERMANY_PREMIUM",
				RateId:           "GBP_71",
				Rate:             csvr.rates["GBP_71"],
				RoundingMethod:   utils.MetaRoundingMiddle,
				RoundingDecimals: 4,
			},
		},
	}) {
		t.Error("Error loading destination rate: ", drs)
	}
}

func TestLoadRatingPlans(t *testing.T) {
	if len(csvr.ratingPlans) != 14 {
		t.Error("Failed to load rating plans: ", len(csvr.ratingPlans))
	}
	rplan := csvr.ratingPlans["STANDARD"]
	expected := &RatingPlan{
		Id: "STANDARD",
		Timings: map[string]*RITiming{
			"59a981b9": {
				Years:     utils.Years{},
				Months:    utils.Months{},
				MonthDays: utils.MonthDays{},
				WeekDays:  utils.WeekDays{1, 2, 3, 4, 5},
				StartTime: "00:00:00",
				tag:       "WORKDAYS_00",
			},
			"2d9ca6c4": {
				Years:     utils.Years{},
				Months:    utils.Months{},
				MonthDays: utils.MonthDays{},
				WeekDays:  utils.WeekDays{1, 2, 3, 4, 5},
				StartTime: "18:00:00",
				tag:       "WORKDAYS_18",
			},
			"ec8ed374": {
				Years:     utils.Years{},
				Months:    utils.Months{},
				MonthDays: utils.MonthDays{},
				WeekDays:  utils.WeekDays{time.Saturday, time.Sunday},
				StartTime: "00:00:00",
				tag:       "WEEKENDS",
			},
			"83429156": {
				Years:     utils.Years{},
				Months:    utils.Months{},
				MonthDays: utils.MonthDays{},
				WeekDays:  utils.WeekDays{},
				StartTime: "00:00:00",
				tag:       "*any",
			},
		},
		Ratings: map[string]*RIRate{
			"ebefae11": {
				ConnectFee: 0,
				Rates: []*RGRate{
					{
						GroupIntervalStart: 0,
						Value:              0.2,
						RateIncrement:      time.Second,
						RateUnit:           time.Minute,
					},
				},
				RoundingMethod:   utils.MetaRoundingMiddle,
				RoundingDecimals: 4,
				tag:              "R1",
			},
			"fac0138e": {
				ConnectFee: 0,
				Rates: []*RGRate{
					{
						GroupIntervalStart: 0,
						Value:              0.1,
						RateIncrement:      time.Second,
						RateUnit:           time.Minute,
					},
				},
				RoundingMethod:   utils.MetaRoundingMiddle,
				RoundingDecimals: 4,
				tag:              "R2",
			},
			"781bfa03": {
				ConnectFee: 0,
				Rates: []*RGRate{
					{
						GroupIntervalStart: 0,
						Value:              0.05,
						RateIncrement:      time.Second,
						RateUnit:           time.Minute,
					},
				},
				RoundingMethod:   utils.MetaRoundingMiddle,
				RoundingDecimals: 4,
				tag:              "R3",
			},
			"f692daa4": {
				ConnectFee: 0,
				Rates: []*RGRate{
					{
						GroupIntervalStart: 0,
						Value:              0,
						RateIncrement:      time.Second,
						RateUnit:           time.Second,
					},
				},
				RoundingMethod:   utils.MetaRoundingMiddle,
				RoundingDecimals: 4,
				tag:              "R_URG",
			},
		},
		DestinationRates: map[string]RPRateList{
			"GERMANY": []*RPRate{
				{
					Timing: "ec8ed374",
					Rating: "ebefae11",
					Weight: 10,
				},
				{
					Timing: "83429156",
					Rating: "fac0138e",
					Weight: 10,
				},
				{
					Timing: "a60bfb13",
					Rating: "fac0138e",
					Weight: 10,
				},
			},
			"GERMANY_O2": []*RPRate{
				{
					Timing: "ec8ed374",
					Rating: "fac0138e",
					Weight: 10,
				},
				{
					Timing: "83429156",
					Rating: "781bfa03",
					Weight: 10,
				},
				{
					Timing: "a60bfb13",
					Rating: "781bfa03",
					Weight: 10,
				},
			},
			"GERMANY_PREMIUM": []*RPRate{
				{
					Timing: "ec8ed374",
					Rating: "16e9ee19",
					Weight: 10,
				},
			},
			"URG": []*RPRate{
				{
					Timing: "2d9ca64",
					Rating: "f692daa4",
					Weight: 20,
				},
			},
		},
	}
	if !reflect.DeepEqual(rplan.Ratings, expected.Ratings) {
		/*for tag, key := range rplan.Ratings {
			log.Print(tag, key)
		}*/
		t.Errorf("Expecting:\n%s\nReceived:\n%s", utils.ToIJSON(expected.Ratings), utils.ToIJSON(rplan.Ratings))
	}
	anyTiming := &RITiming{
		ID:         utils.MetaAny,
		Years:      utils.Years{},
		Months:     utils.Months{},
		MonthDays:  utils.MonthDays{},
		WeekDays:   utils.WeekDays{},
		StartTime:  "00:00:00",
		EndTime:    "",
		cronString: "",
		tag:        utils.MetaAny,
	}

	if !reflect.DeepEqual(csvr.ratingPlans["ANY_PLAN"].Timings["b9b78731"], anyTiming) {
		t.Errorf("Error using *any timing in rating plans: %+v : %+v", utils.ToJSON(csvr.ratingPlans["ANY_PLAN"].Timings["b9b78731"]), utils.ToJSON(anyTiming))
	}
}

func TestLoadRatingProfiles(t *testing.T) {
	if len(csvr.ratingProfiles) != 24 {
		t.Error("Failed to load rating profiles: ", len(csvr.ratingProfiles), csvr.ratingProfiles)
	}
	rp := csvr.ratingProfiles["*out:test:0:trp"]
	expected := &RatingProfile{
		Id: "*out:test:0:trp",
		RatingPlanActivations: RatingPlanActivations{
			&RatingPlanActivation{
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
	if len(csvr.actions) != 16 {
		t.Error("Failed to load actions: ", len(csvr.actions))
	}
	as1 := csvr.actions["MINI"]
	expected := []*Action{
		{
			Id:               "MINI",
			ActionType:       utils.MetaTopUpReset,
			ExpirationString: utils.MetaUnlimited,
			ExtraParameters:  "",
			Weight:           10,
			Balance: &BalanceFilter{
				Type:           utils.StringPointer(utils.MetaMonetary),
				Uuid:           as1[0].Balance.Uuid,
				Value:          &utils.ValueFormula{Static: 10},
				Weight:         utils.Float64Pointer(10),
				DestinationIDs: nil,
				TimingIDs:      nil,
				SharedGroups:   nil,
				Categories:     nil,
				Disabled:       utils.BoolPointer(false),
				Blocker:        utils.BoolPointer(false),
			},
		},
		{
			Id:               "MINI",
			ActionType:       utils.MetaTopUp,
			ExpirationString: utils.MetaUnlimited,
			ExtraParameters:  "",
			Weight:           10,
			Balance: &BalanceFilter{
				Type:           utils.StringPointer(utils.MetaVoice),
				Uuid:           as1[1].Balance.Uuid,
				Value:          &utils.ValueFormula{Static: 100 * float64(time.Second)},
				Weight:         utils.Float64Pointer(10),
				RatingSubject:  utils.StringPointer("test"),
				DestinationIDs: utils.StringMapPointer(utils.NewStringMap("NAT")),
				TimingIDs:      nil,
				SharedGroups:   nil,
				Categories:     nil,
				Disabled:       utils.BoolPointer(false),
				Blocker:        utils.BoolPointer(false),
			},
		},
	}
	if !reflect.DeepEqual(as1, expected) {
		t.Errorf("expecting: %s received: %s",
			utils.ToIJSON(expected), utils.ToIJSON(as1))
	}
	as2 := csvr.actions["SHARED"]
	expected = []*Action{
		{
			Id:               "SHARED",
			ActionType:       utils.MetaTopUp,
			ExpirationString: utils.MetaUnlimited,
			Weight:           10,
			Balance: &BalanceFilter{
				Type:           utils.StringPointer(utils.MetaMonetary),
				DestinationIDs: nil,
				Uuid:           as2[0].Balance.Uuid,
				Value:          &utils.ValueFormula{Static: 100},
				Weight:         utils.Float64Pointer(10),
				SharedGroups:   utils.StringMapPointer(utils.NewStringMap("SG1")),
				TimingIDs:      nil,
				Categories:     nil,
				Disabled:       utils.BoolPointer(false),
				Blocker:        utils.BoolPointer(false),
			},
		},
	}
	if !reflect.DeepEqual(as2, expected) {
		t.Errorf("Error loading action: %s", utils.ToIJSON(as2))
	}
	as3 := csvr.actions["DEFEE"]
	expected = []*Action{
		{
			Id:              "DEFEE",
			ActionType:      utils.CDRLog,
			ExtraParameters: `{"Category":"^ddi","MediationRunId":"^did_run"}`,
			Weight:          10,
			Balance: &BalanceFilter{
				Uuid:           as3[0].Balance.Uuid,
				DestinationIDs: nil,
				TimingIDs:      nil,
				Categories:     nil,
				SharedGroups:   nil,
				Blocker:        utils.BoolPointer(false),
				Disabled:       utils.BoolPointer(false),
			},
		},
	}
	if !reflect.DeepEqual(as3, expected) {
		t.Errorf("Error loading action: %+v", as3[0].Balance)
	}
	asGnrc := csvr.actions["TOPUP_RST_GNR_1000"]
	//TOPUP_RST_GNR_1000,*topup_reset,"{""*voice"": 60.0,""*data"":1024.0,""*sms"":1.0}",,,*generic,*out,,*any,,,*unlimited,,1000,20,false,false,10
	expected = []*Action{
		{
			Id:               "TOPUP_RST_GNR_1000",
			ActionType:       utils.MetaTopUpReset,
			ExtraParameters:  `{"*voice": 60.0,"*data":1024.0,"*sms":1.0}`,
			Weight:           10,
			ExpirationString: utils.MetaUnlimited,
			Balance: &BalanceFilter{
				Uuid:     asGnrc[0].Balance.Uuid,
				Type:     utils.StringPointer(utils.MetaGeneric),
				Value:    &utils.ValueFormula{Static: 1000},
				Weight:   utils.Float64Pointer(20),
				Disabled: utils.BoolPointer(false),
				Blocker:  utils.BoolPointer(false),
			},
		},
	}
	if !reflect.DeepEqual(asGnrc, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected[0].Balance, asGnrc[0].Balance)
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
			"*any": {
				Strategy:      "*lowest",
				RatingSubject: "",
			},
		},
	}
	if !reflect.DeepEqual(sg1, expected) {
		t.Errorf("Expected: %s, received %s ", utils.ToJSON(expected), utils.ToJSON(sg1))
	}
	sg2 := csvr.sharedGroups["SG2"]
	expected = &SharedGroup{
		Id: "SG2",
		AccountParameters: map[string]*SharingParameters{
			"*any": {
				Strategy:      "*lowest",
				RatingSubject: "one",
			},
		},
	}
	if !reflect.DeepEqual(sg2, expected) {
		t.Errorf("Expected: %s, received %s ", utils.ToJSON(expected), utils.ToJSON(sg2))
	}
	/*sg, _ := dataStorage.GetSharedGroup("SG1", false)
	  if len(sg.Members) != 0 {
	      t.Errorf("Memebers should be empty: %+v", sg)
	  }

	  // execute action timings to fill memebers
	  atm := csvr.actionPlans["MORE_MINUTES"][1]
	  atm.Execute()
	  atm.actions, atm.stCache = nil, time.Time{}

	  sg, _ = dataStorage.GetSharedGroup("SG1", false)
	  if len(sg.Members) != 1 {
	      t.Errorf("Memebers should not be empty: %+v", sg)
	  }*/
}

func TestLoadActionTimings(t *testing.T) {
	if len(csvr.actionPlans) != 9 {
		t.Error("Failed to load action timings: ", len(csvr.actionPlans))
	}
	atm := csvr.actionPlans["MORE_MINUTES"]
	expected := &ActionPlan{
		Id:         "MORE_MINUTES",
		AccountIDs: utils.StringMap{},
		ActionTimings: []*ActionTiming{
			{
				Uuid: atm.ActionTimings[0].Uuid,
				Timing: &RateInterval{
					Timing: &RITiming{
						ID:        "ONE_TIME_RUN",
						Years:     utils.Years{2012},
						Months:    utils.Months{},
						MonthDays: utils.MonthDays{},
						WeekDays:  utils.WeekDays{},
						StartTime: utils.MetaASAP,
					},
				},
				Weight:    10,
				ActionsID: "MINI",
			},
			{
				Uuid: atm.ActionTimings[1].Uuid,
				Timing: &RateInterval{
					Timing: &RITiming{
						ID:        "ONE_TIME_RUN",
						Years:     utils.Years{2012},
						Months:    utils.Months{},
						MonthDays: utils.MonthDays{},
						WeekDays:  utils.WeekDays{},
						StartTime: utils.MetaASAP,
					},
				},
				Weight:    10,
				ActionsID: "SHARED",
			},
		},
	}
	if !reflect.DeepEqual(atm, expected) {
		t.Errorf("Error loading action timing:\n%+v", utils.ToJSON(atm))
	}
}

func TestLoadActionTriggers(t *testing.T) {
	if len(csvr.actionsTriggers) != 2 {
		t.Error("Failed to load action triggers: ", len(csvr.actionsTriggers))
	}
	atr := csvr.actionsTriggers["STANDARD_TRIGGER"][0]
	expected := &ActionTrigger{
		ID:             "STANDARD_TRIGGER",
		UniqueID:       "st0",
		ThresholdType:  utils.TriggerMinEventCounter,
		ThresholdValue: 10,
		Balance: &BalanceFilter{
			ID:             nil,
			Type:           utils.StringPointer(utils.MetaVoice),
			DestinationIDs: utils.StringMapPointer(utils.NewStringMap("GERMANY_O2")),
			Categories:     nil,
			TimingIDs:      nil,
			SharedGroups:   nil,
			Disabled:       nil,
			Blocker:        nil,
		},
		Weight:    10,
		ActionsID: "SOME_1",
		Executed:  false,
	}
	if !reflect.DeepEqual(atr, expected) {
		t.Errorf("Expected: %+v, received %+v", utils.ToJSON(expected), utils.ToJSON(atr))
	}
	atr = csvr.actionsTriggers["STANDARD_TRIGGER"][1]
	expected = &ActionTrigger{
		ID:             "STANDARD_TRIGGER",
		UniqueID:       "st1",
		ThresholdType:  utils.TriggerMaxBalance,
		ThresholdValue: 200,
		Balance: &BalanceFilter{
			Type:           utils.StringPointer(utils.MetaVoice),
			DestinationIDs: utils.StringMapPointer(utils.NewStringMap("GERMANY")),
			Categories:     nil,
			TimingIDs:      nil,
			SharedGroups:   nil,
		},
		Weight:    10,
		ActionsID: "SOME_2",
		Executed:  false,
	}
	if !reflect.DeepEqual(atr, expected) {
		t.Errorf("Error loading action trigger: %+v", atr)
	}
}

func TestLoadAccountActions(t *testing.T) {
	if len(csvr.accountActions) != 17 {
		t.Error("Failed to load account actions: ", len(csvr.accountActions))
	}
	aa := csvr.accountActions["vdf:minitsboy"]
	expected := &Account{
		ID: "vdf:minitsboy",
		UnitCounters: UnitCounters{
			utils.MetaVoice: []*UnitCounter{
				{
					CounterType: "*event",
					Counters: CounterFilters{
						&CounterFilter{
							Value: 0,
							Filter: &BalanceFilter{
								ID:             utils.StringPointer("st0"),
								Type:           utils.StringPointer(utils.MetaVoice),
								DestinationIDs: utils.StringMapPointer(utils.NewStringMap("GERMANY_O2")),
								SharedGroups:   nil,
								Categories:     nil,
								TimingIDs:      nil,
							},
						},
					},
				},
			},
		},
		ActionTriggers: csvr.actionsTriggers["STANDARD_TRIGGER"],
	}
	// set propper uuid
	for i, atr := range aa.ActionTriggers {
		csvr.actionsTriggers["STANDARD_TRIGGER"][i].ID = atr.ID
	}
	for i, b := range aa.UnitCounters[utils.MetaVoice][0].Counters {
		expected.UnitCounters[utils.MetaVoice][0].Counters[i].Filter.ID = b.Filter.ID
	}
	if !reflect.DeepEqual(aa.UnitCounters[utils.MetaVoice][0].Counters[0], expected.UnitCounters[utils.MetaVoice][0].Counters[0]) {
		t.Errorf("Error loading account action: %+v", utils.ToIJSON(aa.UnitCounters[utils.MetaVoice][0].Counters[0].Filter))
	}
	// test that it does not overwrite balances
	existing, err := dm.GetAccount(aa.ID)
	if err != nil || len(existing.BalanceMap) != 2 {
		t.Errorf("The account was not set before load: %+v", existing)
	}
	dm.SetAccount(aa)
	existing, err = dm.GetAccount(aa.ID)
	if err != nil || len(existing.BalanceMap) != 2 {
		t.Errorf("The set account altered the balances: %+v", existing)
	}
}

func TestLoadResourceProfiles(t *testing.T) {
	eResProfiles := map[utils.TenantID]*utils.TPResourceProfile{
		{Tenant: "cgrates.org", ID: "ResGroup21"}: {
			TPid:      testTPID,
			Tenant:    "cgrates.org",
			ID:        "ResGroup21",
			FilterIDs: []string{"*string:~*req.Account:1001"},
			ActivationInterval: &utils.TPActivationInterval{
				ActivationTime: "2014-07-29T15:00:00Z",
			},
			UsageTTL:          "1s",
			AllocationMessage: "call",
			Weight:            10,
			Limit:             "2",
			Blocker:           true,
			Stored:            true,
		},
		{Tenant: "cgrates.org", ID: "ResGroup22"}: {
			TPid:      testTPID,
			Tenant:    "cgrates.org",
			ID:        "ResGroup22",
			FilterIDs: []string{"*string:~*req.Account:dan"},
			ActivationInterval: &utils.TPActivationInterval{
				ActivationTime: "2014-07-29T15:00:00Z",
			},
			UsageTTL:          "3600s",
			AllocationMessage: "premium_call",
			Blocker:           true,
			Stored:            true,
			Weight:            10,
			Limit:             "2",
		},
	}
	resKey := utils.TenantID{Tenant: "cgrates.org", ID: "ResGroup21"}
	if len(csvr.resProfiles) != len(eResProfiles) {
		t.Errorf("Failed to load ResourceProfiles: %s", utils.ToIJSON(csvr.resProfiles))
	} else if !reflect.DeepEqual(eResProfiles[resKey], csvr.resProfiles[resKey]) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eResProfiles[resKey]), utils.ToJSON(csvr.resProfiles[resKey]))
	}
}

func TestLoadStatQueueProfiles(t *testing.T) {
	eStats := map[utils.TenantID]*utils.TPStatProfile{
		{Tenant: "cgrates.org", ID: "TestStats"}: {
			Tenant:    "cgrates.org",
			TPid:      testTPID,
			ID:        "TestStats",
			FilterIDs: []string{"*string:~*req.Account:1001"},
			ActivationInterval: &utils.TPActivationInterval{
				ActivationTime: "2014-07-29T15:00:00Z",
			},
			QueueLength: 100,
			TTL:         "1s",
			Metrics: []*utils.MetricWithFilters{
				{
					MetricID: "*sum#Value",
				},
				{
					MetricID: "*sum#Usage",
				},
				{
					MetricID: "*average#Value",
				},
			},
			ThresholdIDs: []string{"Th1", "Th2"},
			Blocker:      true,
			Stored:       true,
			Weight:       20,
			MinItems:     2,
		},
		{Tenant: "cgrates.org", ID: "TestStats2"}: {
			Tenant:    "cgrates.org",
			TPid:      testTPID,
			ID:        "TestStats2",
			FilterIDs: []string{"FLTR_1"},
			ActivationInterval: &utils.TPActivationInterval{
				ActivationTime: "2014-07-29T15:00:00Z",
			},
			QueueLength: 100,
			TTL:         "1s",
			Metrics: []*utils.MetricWithFilters{
				{
					MetricID: "*sum#Value",
				},
				{
					MetricID: "*sum#Usage",
				},
				{
					FilterIDs: []string{"*string:Account:1001"},
					MetricID:  "*sum#Cost",
				},
				{
					MetricID: "*average#Value",
				},
				{
					MetricID: "*average#Usage",
				},
				{
					FilterIDs: []string{"*string:Account:1001"},
					MetricID:  "*average#Cost",
				},
			},
			ThresholdIDs: []string{"Th"},
			Blocker:      true,
			Stored:       true,
			Weight:       20,
			MinItems:     2,
		},
	}
	stKeys := []utils.TenantID{
		{Tenant: "cgrates.org", ID: "TestStats"},
		{Tenant: "cgrates.org", ID: "TestStats2"},
	}
	for _, stKey := range stKeys {
		if len(csvr.sqProfiles) != len(eStats) {
			t.Errorf("Failed to load StatQueueProfiles: %s",
				utils.ToJSON(csvr.sqProfiles))
		} else if !reflect.DeepEqual(eStats[stKey].Tenant,
			csvr.sqProfiles[stKey].Tenant) {
			t.Errorf("Expecting: %s, received: %s",
				eStats[stKey].Tenant, csvr.sqProfiles[stKey].Tenant)
		} else if !reflect.DeepEqual(eStats[stKey].ID,
			csvr.sqProfiles[stKey].ID) {
			t.Errorf("Expecting: %s, received: %s",
				eStats[stKey].ID, csvr.sqProfiles[stKey].ID)
		} else if !reflect.DeepEqual(len(eStats[stKey].ThresholdIDs),
			len(csvr.sqProfiles[stKey].ThresholdIDs)) {
			t.Errorf("Expecting: %s, received: %s",
				utils.ToJSON(eStats[stKey].ThresholdIDs),
				utils.ToJSON(csvr.sqProfiles[stKey].ThresholdIDs))
		} else if !reflect.DeepEqual(len(eStats[stKey].Metrics),
			len(csvr.sqProfiles[stKey].Metrics)) {
			t.Errorf("Expecting: %s, \n received: %s",
				utils.ToJSON(eStats[stKey].Metrics),
				utils.ToJSON(csvr.sqProfiles[stKey].Metrics))
		}
	}
}

func TestLoadRankingProfiles(t *testing.T) {
	eRankings := map[utils.TenantID]*utils.TPRankingProfile{
		{Tenant: "cgrates.org", ID: "Ranking1"}: {
			TPid:         testTPID,
			Tenant:       "cgrates.org",
			ID:           "Ranking1",
			Schedule:     "15m",
			StatIDs:      []string{"Stats2", "Stats3", "Stats4"},
			MetricIDs:    []string{"Metric1", "Metric3"},
			Sorting:      "*asc",
			ThresholdIDs: []string{"THD1", "THD2"},
		},
	}
	rgkey := utils.TenantID{Tenant: "cgrates.org", ID: "RANKING1"}
	if len(eRankings) != len(csvr.rgProfiles) {
		t.Errorf("Failed to load RankingProfiles: %+v", csvr.rgProfiles)
	} else if diff := cmp.Diff(eRankings[rgkey], csvr.rgProfiles[rgkey], cmpopts.SortSlices(func(a, b string) bool {
		return a < b
	})); diff != "" {
		t.Errorf("Wrong TPRankingProfiles (-expected +got):\n%s", diff)
	}
}

func TestTrendProfiles(t *testing.T) {
	eTrends := map[utils.TenantID]*utils.TPTrendsProfile{
		{Tenant: "cgrates.org", ID: "TREND1"}: {
			TPid:            testTPID,
			Tenant:          "cgrates.org",
			ID:              "TREND1",
			Schedule:        "0 12 * * *",
			StatID:          "Stats2",
			Metrics:         []string{"*acc", "*tcc"},
			QueueLength:     -1,
			TTL:             "-1",
			MinItems:        1,
			CorrelationType: "*average",
			Tolerance:       2.1,
			Stored:          true,
			ThresholdIDs:    []string{"TD1", "TD2"},
		},
	}
	rgkey := utils.TenantID{Tenant: "cgrates.org", ID: "TREND1"}
	if len(eTrends) != len(csvr.trProfiles) {
		t.Errorf("Failed to load TrendProfiles: %+v", csvr.trProfiles)
	} else if diff := cmp.Diff(eTrends[rgkey], csvr.trProfiles[rgkey], cmpopts.SortSlices(func(a, b string) bool {
		return a < b
	})); diff != "" {
		t.Errorf("Wrong TrendProfiles (-expected +got):\n%s", diff)
	}
}

func TestLoadThresholdProfiles(t *testing.T) {
	eThresholds := map[utils.TenantID]*utils.TPThresholdProfile{
		{Tenant: "cgrates.org", ID: "Threshold1"}: {
			TPid:      testTPID,
			Tenant:    "cgrates.org",
			ID:        "Threshold1",
			FilterIDs: []string{"*string:~*req.Account:1001", "*string:~*req.RunID:*default"},
			ActivationInterval: &utils.TPActivationInterval{
				ActivationTime: "2014-07-29T15:00:00Z",
			},
			MaxHits:   12,
			MinHits:   10,
			MinSleep:  "1s",
			Blocker:   true,
			Weight:    10,
			ActionIDs: []string{"THRESH1"},
			Async:     true,
		},
	}
	eThresholdReverse := map[utils.TenantID]*utils.TPThresholdProfile{
		{Tenant: "cgrates.org", ID: "Threshold1"}: {
			TPid:      testTPID,
			Tenant:    "cgrates.org",
			ID:        "Threshold1",
			FilterIDs: []string{"*string:~*req.RunID:*default", "*string:~*req.Account:1001"},
			ActivationInterval: &utils.TPActivationInterval{
				ActivationTime: "2014-07-29T15:00:00Z",
			},
			MaxHits:   12,
			MinHits:   10,
			MinSleep:  "1s",
			Blocker:   true,
			Weight:    10,
			ActionIDs: []string{"THRESH1"},
			Async:     true,
		},
	}
	thkey := utils.TenantID{Tenant: "cgrates.org", ID: "Threshold1"}
	if len(csvr.thProfiles) != len(eThresholds) {
		t.Errorf("Failed to load ThresholdProfiles: %s", utils.ToIJSON(csvr.thProfiles))
	} else if !reflect.DeepEqual(eThresholds[thkey], csvr.thProfiles[thkey]) &&
		!reflect.DeepEqual(eThresholdReverse[thkey], csvr.thProfiles[thkey]) {
		t.Errorf("Expecting: %+v , %+v , received: %+v", eThresholds[thkey],
			eThresholdReverse[thkey], csvr.thProfiles[thkey])
	}
}

func TestLoadFilters(t *testing.T) {
	eFilters := map[utils.TenantID]*utils.TPFilterProfile{
		{Tenant: "cgrates.org", ID: "FLTR_1"}: {
			TPid:   testTPID,
			Tenant: "cgrates.org",
			ID:     "FLTR_1",
			Filters: []*utils.TPFilter{
				{
					Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.AccountField,
					Type:    utils.MetaString,
					Values:  []string{"1001", "1002"},
				},
				{
					Element: "~*req.Destination",
					Type:    utils.MetaPrefix,
					Values:  []string{"10", "20"},
				},
				{
					Element: "~*req.Subject",
					Type:    utils.MetaRSR,
					Values:  []string{"~^1.*1$"},
				},
				{
					Element: "~*req.Destination",
					Type:    utils.MetaRSR,
					Values:  []string{"1002"},
				},
			},
			ActivationInterval: &utils.TPActivationInterval{
				ActivationTime: "2014-07-29T15:00:00Z",
			},
		},
		{Tenant: "cgrates.org", ID: "FLTR_ACNT_dan"}: {
			TPid:   testTPID,
			Tenant: "cgrates.org",
			ID:     "FLTR_ACNT_dan",
			Filters: []*utils.TPFilter{
				{
					Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.AccountField,
					Type:    utils.MetaString,
					Values:  []string{"dan"},
				},
			},
			ActivationInterval: &utils.TPActivationInterval{
				ActivationTime: "2014-07-29T15:00:00Z",
			},
		},
		{Tenant: "cgrates.org", ID: "FLTR_DST_DE"}: {
			TPid:   testTPID,
			Tenant: "cgrates.org",
			ID:     "FLTR_DST_DE",
			Filters: []*utils.TPFilter{
				{
					Element: "~*req.Destination",
					Type:    utils.MetaDestinations,
					Values:  []string{"DST_DE"},
				},
			},
			ActivationInterval: &utils.TPActivationInterval{
				ActivationTime: "2014-07-29T15:00:00Z",
			},
		},
		{Tenant: "cgrates.org", ID: "FLTR_DST_NL"}: {
			TPid:   testTPID,
			Tenant: "cgrates.org",
			ID:     "FLTR_DST_NL",
			Filters: []*utils.TPFilter{
				{
					Element: "~*req.Destination",
					Type:    utils.MetaDestinations,
					Values:  []string{"DST_NL"},
				},
			},
			ActivationInterval: &utils.TPActivationInterval{
				ActivationTime: "2014-07-29T15:00:00Z",
			},
		},
	}
	fltrKey := utils.TenantID{Tenant: "cgrates.org", ID: "FLTR_1"}
	if len(csvr.filters) != len(eFilters) {
		t.Errorf("Failed to load Filters: %s", utils.ToIJSON(csvr.filters))
	} else if !reflect.DeepEqual(eFilters[fltrKey], csvr.filters[fltrKey]) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eFilters[fltrKey]), utils.ToJSON(csvr.filters[fltrKey]))
	}
}

func TestLoadRouteProfiles(t *testing.T) {
	eSppProfile := &utils.TPRouteProfile{
		TPid:      testTPID,
		Tenant:    "cgrates.org",
		ID:        "RoutePrf1",
		FilterIDs: []string{"*string:~*req.Account:dan"},
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-29T15:00:00Z",
		},
		Sorting: utils.MetaLC,
		Routes: []*utils.TPRoute{
			{
				ID:              "route1",
				FilterIDs:       []string{"FLTR_ACNT_dan"},
				AccountIDs:      []string{"Account1", "Account1_1"},
				RatingPlanIDs:   []string{"RPL_1"},
				ResourceIDs:     []string{"ResGroup1"},
				StatIDs:         []string{"Stat1"},
				Weight:          10,
				Blocker:         true,
				RouteParameters: "param1",
			},
			{
				ID:              "route1",
				RatingPlanIDs:   []string{"RPL_2"},
				ResourceIDs:     []string{"ResGroup2", "ResGroup4"},
				StatIDs:         []string{"Stat3"},
				Weight:          10,
				Blocker:         false,
				RouteParameters: utils.EmptyString,
			},
			{
				ID:              "route1",
				FilterIDs:       []string{"FLTR_DST_DE"},
				AccountIDs:      []string{"Account2"},
				RatingPlanIDs:   []string{"RPL_3"},
				ResourceIDs:     []string{"ResGroup3"},
				StatIDs:         []string{"Stat2"},
				Weight:          10,
				Blocker:         false,
				RouteParameters: utils.EmptyString,
			},
		},
		Weight: 20,
	}
	sort.Slice(eSppProfile.Routes, func(i, j int) bool {
		return strings.Compare(eSppProfile.Routes[i].ID+strings.Join(eSppProfile.Routes[i].FilterIDs, utils.ConcatenatedKeySep),
			eSppProfile.Routes[j].ID+strings.Join(eSppProfile.Routes[j].FilterIDs, utils.ConcatenatedKeySep)) < 0
	})
	resKey := utils.TenantID{Tenant: "cgrates.org", ID: "RoutePrf1"}
	if len(csvr.routeProfiles) != 1 {
		t.Errorf("Failed to load RouteProfiles: %s", utils.ToIJSON(csvr.routeProfiles))
	} else {
		rcvRoute := csvr.routeProfiles[resKey]
		if rcvRoute == nil {
			t.Fatal("Missing route")
		}
		sort.Slice(rcvRoute.Routes, func(i, j int) bool {
			return strings.Compare(rcvRoute.Routes[i].ID+strings.Join(rcvRoute.Routes[i].FilterIDs, utils.ConcatenatedKeySep),
				rcvRoute.Routes[j].ID+strings.Join(rcvRoute.Routes[j].FilterIDs, utils.ConcatenatedKeySep)) < 0
		})
		if !reflect.DeepEqual(eSppProfile, rcvRoute) {
			t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eSppProfile), utils.ToJSON(rcvRoute))
		}
	}
}

func TestLoadAttributeProfiles(t *testing.T) {
	eAttrProfiles := map[utils.TenantID]*utils.TPAttributeProfile{
		{Tenant: "cgrates.org", ID: "ALS1"}: {
			TPid:      testTPID,
			Tenant:    "cgrates.org",
			ID:        "ALS1",
			Contexts:  []string{"con1", "con2", "con3"},
			FilterIDs: []string{"*string:~*req.Account:1001"},
			ActivationInterval: &utils.TPActivationInterval{
				ActivationTime: "2014-07-29T15:00:00Z",
			},
			Attributes: []*utils.TPAttribute{
				{
					FilterIDs: []string{"*string:~*req.Field1:Initial"},
					Path:      utils.MetaReq + utils.NestingSep + "Field1",
					Type:      utils.MetaVariable,
					Value:     "Sub1",
				},
				{
					FilterIDs: []string{},
					Path:      utils.MetaReq + utils.NestingSep + "Field2",
					Type:      utils.MetaVariable,
					Value:     "Sub2",
				},
			},
			Blocker: true,
			Weight:  20,
		},
	}
	resKey := utils.TenantID{Tenant: "cgrates.org", ID: "ALS1"}
	if len(csvr.attributeProfiles) != len(eAttrProfiles) {
		t.Errorf("Failed to load attributeProfiles: %s", utils.ToIJSON(csvr.attributeProfiles))
	} else if !reflect.DeepEqual(eAttrProfiles[resKey].Tenant, csvr.attributeProfiles[resKey].Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", eAttrProfiles[resKey].Tenant, csvr.attributeProfiles[resKey].Tenant)
	} else if !reflect.DeepEqual(eAttrProfiles[resKey].ID, csvr.attributeProfiles[resKey].ID) {
		t.Errorf("Expecting: %+v, received: %+v", eAttrProfiles[resKey].ID, csvr.attributeProfiles[resKey].ID)
	} else if len(eAttrProfiles[resKey].Contexts) != len(csvr.attributeProfiles[resKey].Contexts) {
		t.Errorf("Expecting: %+v, received: %+v", eAttrProfiles[resKey].Contexts, csvr.attributeProfiles[resKey].Contexts)
	} else if !reflect.DeepEqual(eAttrProfiles[resKey].FilterIDs, csvr.attributeProfiles[resKey].FilterIDs) {
		t.Errorf("Expecting: %+v, received: %+v", eAttrProfiles[resKey].FilterIDs, csvr.attributeProfiles[resKey].FilterIDs)
	} else if !reflect.DeepEqual(eAttrProfiles[resKey].ActivationInterval.ActivationTime, csvr.attributeProfiles[resKey].ActivationInterval.ActivationTime) {
		t.Errorf("Expecting: %+v, received: %+v", eAttrProfiles[resKey].ActivationInterval, csvr.attributeProfiles[resKey].ActivationInterval)
	} else if !reflect.DeepEqual(eAttrProfiles[resKey].Attributes, csvr.attributeProfiles[resKey].Attributes) {
		t.Errorf("Expecting: %+v, received: %+v", eAttrProfiles[resKey].Attributes, csvr.attributeProfiles[resKey].Attributes)
	} else if !reflect.DeepEqual(eAttrProfiles[resKey].Blocker, csvr.attributeProfiles[resKey].Blocker) {
		t.Errorf("Expecting: %+v, received: %+v", eAttrProfiles[resKey].Blocker, csvr.attributeProfiles[resKey].Blocker)
	}
}

func TestLoadChargerProfiles(t *testing.T) {
	eChargerProfiles := map[utils.TenantID]*utils.TPChargerProfile{
		{Tenant: "cgrates.org", ID: "Charger1"}: {
			TPid:      testTPID,
			Tenant:    "cgrates.org",
			ID:        "Charger1",
			FilterIDs: []string{"*string:~*req.Account:1001"},
			ActivationInterval: &utils.TPActivationInterval{
				ActivationTime: "2014-07-29T15:00:00Z",
			},
			RunID:        "*rated",
			AttributeIDs: []string{"ATTR_1001_SIMPLEAUTH"},
			Weight:       20,
		},
	}
	cppKey := utils.TenantID{Tenant: "cgrates.org", ID: "Charger1"}
	if len(csvr.chargerProfiles) != len(eChargerProfiles) {
		t.Errorf("Failed to load chargerProfiles: %s", utils.ToIJSON(csvr.chargerProfiles))
	} else if !reflect.DeepEqual(eChargerProfiles[cppKey], csvr.chargerProfiles[cppKey]) {
		t.Errorf("Expecting: %+v, received: %+v", eChargerProfiles[cppKey], csvr.chargerProfiles[cppKey])
	}
}

func TestLoadDispatcherProfiles(t *testing.T) {
	eDispatcherProfiles := &utils.TPDispatcherProfile{
		TPid:       testTPID,
		Tenant:     "cgrates.org",
		ID:         "D1",
		Subsystems: []string{"*any"},
		FilterIDs:  []string{"*string:~*req.Account:1001"},
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-29T15:00:00Z",
		},
		Strategy: "*first",
		Weight:   20,
		Hosts: []*utils.TPDispatcherHostProfile{
			{
				ID:        "C1",
				FilterIDs: []string{"*gt:~*req.Usage:10"},
				Weight:    10,
				Params:    []any{"192.168.56.203"},
				Blocker:   false,
			},
			{
				ID:        "C2",
				FilterIDs: []string{"*lt:~*req.Usage:10"},
				Weight:    10,
				Params:    []any{"192.168.56.204"},
				Blocker:   false,
			},
		},
	}
	if len(csvr.dispatcherProfiles) != 1 {
		t.Errorf("Failed to load dispatcherProfiles: %s", utils.ToIJSON(csvr.dispatcherProfiles))
	}
	dppKey := utils.TenantID{Tenant: "cgrates.org", ID: "D1"}
	sort.Slice(eDispatcherProfiles.Hosts, func(i, j int) bool {
		return eDispatcherProfiles.Hosts[i].ID < eDispatcherProfiles.Hosts[j].ID
	})
	sort.Slice(csvr.dispatcherProfiles[dppKey].Hosts, func(i, j int) bool {
		return csvr.dispatcherProfiles[dppKey].Hosts[i].ID < csvr.dispatcherProfiles[dppKey].Hosts[j].ID
	})

	if !reflect.DeepEqual(eDispatcherProfiles, csvr.dispatcherProfiles[dppKey]) {
		t.Errorf("Expecting: %+v, received: %+v",
			utils.ToJSON(eDispatcherProfiles), utils.ToJSON(csvr.dispatcherProfiles[dppKey]))
	}
}

func TestLoadDispatcherHosts(t *testing.T) {
	eDispatcherHosts := &utils.TPDispatcherHost{
		TPid:   testTPID,
		Tenant: "cgrates.org",
		ID:     "ALL",
		Conn: &utils.TPDispatcherHostConn{
			Address:              "127.0.0.1:6012",
			Transport:            utils.MetaJSON,
			ConnectAttempts:      1,
			Reconnects:           3,
			MaxReconnectInterval: 5 * time.Minute,
			ConnectTimeout:       1 * time.Minute,
			ReplyTimeout:         2 * time.Minute,
			TLS:                  false,
		},
	}

	dphKey := utils.TenantID{Tenant: "cgrates.org", ID: "ALL"}
	if len(csvr.dispatcherHosts) != 1 {
		t.Fatalf("Failed to load DispatcherHosts: %v", len(csvr.dispatcherHosts))
	}
	if !reflect.DeepEqual(eDispatcherHosts, csvr.dispatcherHosts[dphKey]) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eDispatcherHosts), utils.ToJSON(csvr.dispatcherHosts[dphKey]))
	}
}
