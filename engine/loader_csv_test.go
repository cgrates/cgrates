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
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"
)

var (
	testTPID     = "LoaderCSVTests"
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
	timings = `
WORKDAYS_00,*any,*any,*any,1;2;3;4;5,00:00:00
WORKDAYS_18,*any,*any,*any,1;2;3;4;5,18:00:00
WEEKENDS,*any,*any,*any,6;7,00:00:00
ONE_TIME_RUN,2012,,,,*asap
`
	rates = `
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
	destinationRates = `
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
	ratingPlans = `
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
	ratingProfiles = `
*out,CUSTOMER_1,0,rif:from:tm,2012-01-01T00:00:00Z,PREMIUM,danb,
*out,CUSTOMER_1,0,rif:from:tm,2012-02-28T00:00:00Z,STANDARD,danb,
*out,CUSTOMER_2,0,danb:87.139.12.167,2012-01-01T00:00:00Z,STANDARD,danb,
*out,CUSTOMER_1,0,danb,2012-01-01T00:00:00Z,PREMIUM,,
*out,vdf,0,rif,2012-01-01T00:00:00Z,EVENING,,
*out,vdf,call,rif,2012-02-28T00:00:00Z,EVENING,,
*out,vdf,call,dan,2012-01-01T00:00:00Z,EVENING,,
*out,vdf,0,minu,2012-01-01T00:00:00Z,EVENING,,
*out,vdf,0,*any,2012-02-28T00:00:00Z,EVENING,,
*out,vdf,0,one,2012-02-28T00:00:00Z,STANDARD,,
*out,vdf,0,inf,2012-02-28T00:00:00Z,STANDARD,inf,
*out,vdf,0,fall,2012-02-28T00:00:00Z,PREMIUM,rif,
*out,test,0,trp,2013-10-01T00:00:00Z,TDRT,rif;danb,
*out,vdf,0,fallback1,2013-11-18T13:45:00Z,G,fallback2,
*out,vdf,0,fallback1,2013-11-18T13:46:00Z,G,fallback2,
*out,vdf,0,fallback1,2013-11-18T13:47:00Z,G,fallback2,
*out,vdf,0,fallback2,2013-11-18T13:45:00Z,R,rif,
*out,cgrates.org,call,*any,2013-01-06T00:00:00Z,RP_UK,,
*out,cgrates.org,call,discounted_minutes,2013-01-06T00:00:00Z,RP_UK_Mobile_BIG5_PKG,,
*out,cgrates.org,data,rif,2013-01-06T00:00:00Z,RP_DATA,,
*out,cgrates.org,call,max,2013-03-23T00:00:00Z,RP_MX,,
*out,cgrates.org,call,nt,2012-02-28T00:00:00Z,GER_ONLY,,
*in,cgrates.org,LCR_STANDARD,max,2013-03-23T00:00:00Z,RP_MX,,
*out,cgrates.org,call,money,2015-02-28T00:00:00Z,EVENING,,
*out,cgrates.org,call,dy,2015-02-28T00:00:00Z,DY_PLAN,,
*out,cgrates.org,call,block,2015-02-28T00:00:00Z,DY_PLAN,,
*out,cgrates.org,call,round,2016-06-30T00:00:00Z,DEFAULT,,
`
	sharedGroups = `
SG1,*any,*lowest,
SG2,*any,*lowest,one
SG3,*any,*lowest,
`
	actions = `
MINI,*topup_reset,,,,*monetary,*out,,,,,*unlimited,,10,10,false,false,10
MINI,*topup,,,,*voice,*out,,NAT,test,,*unlimited,,100s,10,false,false,10
SHARED,*topup,,,,*monetary,*out,,,,SG1,*unlimited,,100,10,false,false,10
TOPUP10_AC,*topup_reset,,,,*monetary,*out,,*any,,,*unlimited,,1,10,false,false,10
TOPUP10_AC1,*topup_reset,,,,*voice,*out,,DST_UK_Mobile_BIG5,discounted_minutes,,*unlimited,,40s,10,false,false,10
SE0,*topup_reset,,,,*monetary,*out,,,,SG2,*unlimited,,0,10,false,false,10
SE10,*topup_reset,,,,*monetary,*out,,,,SG2,*unlimited,,10,5,false,false,10
SE10,*topup,,,,*monetary,*out,,,,,*unlimited,,10,10,false,false,10
EE0,*topup_reset,,,,*monetary,*out,,,,SG3,*unlimited,,0,10,false,false,10
EE0,*allow_negative,,,,*monetary,*out,,,,,*unlimited,,0,10,false,false,10
DEFEE,*cdrlog,"{""Category"":""^ddi"",""MediationRunId"":""^did_run""}",,,,,,,,,,,,,false,false,10
NEG,*allow_negative,,,,*monetary,*out,,,,,*unlimited,,0,10,false,false,10
BLOCK,*topup,,,bblocker,*monetary,*out,,NAT,,,*unlimited,,1,20,true,false,20
BLOCK,*topup,,,bfree,*monetary,*out,,,,,*unlimited,,20,10,false,false,10
BLOCK_EMPTY,*topup,,,bblocker,*monetary,*out,,NAT,,,*unlimited,,0,20,true,false,20
BLOCK_EMPTY,*topup,,,bfree,*monetary,*out,,,,,*unlimited,,20,10,false,false,10
FILTER,*topup,,"{""*and"":[{""Value"":{""*lt"":0}},{""Id"":{""*eq"":""*default""}}]}",bfree,*monetary,*out,,,,,*unlimited,,20,10,false,false,10
EXP,*topup,,,,*voice,*out,,,,,*monthly,*any,300s,10,false,false,10
NOEXP,*topup,,,,*voice,*out,,,,,*unlimited,*any,50s,10,false,false,10
VF,*debit,,,,*monetary,*out,,,,,*unlimited,*any,"{""Method"":""*incremental"",""Params"":{""Units"":10, ""Interval"":""month"", ""Increment"":""day""}}",10,false,false,10
TOPUP_RST_GNR_1000,*topup_reset,"{""*voice"": 60.0,""*data"":1024.0,""*sms"":1.0}",,,*generic,*out,,*any,,,*unlimited,,1000,20,false,false,10
`
	actionPlans = `
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

	actionTriggers = `
STANDARD_TRIGGER,st0,*min_event_counter,10,false,0,,,,*voice,*out,,GERMANY_O2,,,,,,,,,SOME_1,10
STANDARD_TRIGGER,st1,*max_balance,200,false,0,,,,*voice,*out,,GERMANY,,,,,,,,,SOME_2,10
STANDARD_TRIGGERS,,*min_balance,2,false,0,,,,*monetary,*out,,,,,,,,,,,LOG_WARNING,10
STANDARD_TRIGGERS,,*max_balance,20,false,0,,,,*monetary,*out,,,,,,,,,,,LOG_WARNING,10
STANDARD_TRIGGERS,,*max_event_counter,5,false,0,,,,*monetary,*out,,FS_USERS,,,,,,,,,LOG_WARNING,10
CDRST1_WARN_ASR,,*min_asr,45,true,1h,,,,,,,,,,,,,,,3,CDRST_WARN_HTTP,10
CDRST1_WARN_ACD,,*min_acd,10,true,1h,,,,,,,,,,,,,,,5,CDRST_WARN_HTTP,10
CDRST1_WARN_ACC,,*max_acc,10,true,10m,,,,,,,,,,,,,,,5,CDRST_WARN_HTTP,10
CDRST2_WARN_ASR,,*min_asr,30,true,0,,,,,,,,,,,,,,,5,CDRST_WARN_HTTP,10
CDRST2_WARN_ACD,,*min_acd,3,true,0,,,,,,,,,,,,,,,5,CDRST_WARN_HTTP,10
`
	accountActions = `
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

	derivedCharges = `
#Direction,Tenant,Category,Account,Subject,DestinationIds,RunId,RunFilter,RequestTypeField,DirectionField,TenantField,TorField,AccountField,SubjectField,DestinationField,SetupTimeField,PddField,AnswerTimeField,UsageField,SupplierField,DisconnectCauseField,CostField,RatedField
*out,cgrates.org,call,dan,dan,,extra1,^filteredHeader1/filterValue1/,^prepaid,,,,rif,rif,,,,,,,,,
*out,cgrates.org,call,dan,dan,,extra2,,,,,,ivo,ivo,,,,,,,,,
*out,cgrates.org,call,dan,*any,,extra1,,,,,,rif2,rif2,,,,,,,,,
`
	users = `
#Tenant[0],UserName[1],AttributeName[2],AttributeValue[3],Weight[4]
cgrates.org,rif,false,test0,val0,10
cgrates.org,rif,,test1,val1,10
cgrates.org,dan,,another,value,10
cgrates.org,mas,true,another,value,10
`
	aliases = `
#Direction[0],Tenant[1],Category[2],Account[3],Subject[4],DestinationId[5],Context[6],Target[7],Original[8],Alias[9],Weight[10]
*out,cgrates.org,call,dan,dan,EU_LANDLINE,*rating,Subject,dan,dan1,10
*out,cgrates.org,call,dan,dan,EU_LANDLINE,*rating,Subject,rif,rif1,10
*out,cgrates.org,call,dan,dan,EU_LANDLINE,*rating,Cli,0723,0724,10
*out,cgrates.org,call,dan,dan,GLOBAL1,*rating,Subject,dan,dan2,20
*any,*any,*any,*any,*any,*any,*rating,Subject,*any,rif1,20
*any,*any,*any,*any,*any,*any,*rating,Account,*any,dan1,10
*out,vdf,0,a1,a1,*any,*rating,Subject,a1,minu,10
*out,vdf,0,a1,a1,*any,*rating,Account,a1,minu,10
*out,cgrates.org,call,remo,remo,*any,*rating,Subject,remo,minu,10
*out,cgrates.org,call,remo,remo,*any,*rating,Account,remo,minu,10
`
	resProfiles = `
#Tenant[0],Id[1],FilterIDs[2],ActivationInterval[3],TTL[4],Limit[5],AllocationMessage[6],Blocker[7],Stored[8],Weight[9],Thresholds[10]
cgrates.org,ResGroup21,FLTR_1,2014-07-29T15:00:00Z,1s,2,call,true,true,10,
cgrates.org,ResGroup22,FLTR_ACNT_dan,2014-07-29T15:00:00Z,3600s,2,premium_call,true,true,10,
`
	stats = `
#Tenant[0],Id[1],FilterIDs[2],ActivationInterval[3],QueueLength[4],TTL[5],Metrics[6],Blocker[7],Stored[8],Weight[9],MinItems[10],Thresholds[11]
cgrates.org,TestStats,FLTR_1,2014-07-29T15:00:00Z,100,1s,*sum;*average,Value,true,true,20,2,Th1;Th2
cgrates.org,TestStats,,,,,*sum,Usage,true,true,20,2,
cgrates.org,TestStats2,FLTR_1,2014-07-29T15:00:00Z,100,1s,*sum;*average,Value;Usage,true,true,20,2,Th
cgrates.org,TestStats2,,,,,*sum;*average,Cost,true,true,20,2,
`

	thresholds = `
#Tenant[0],Id[1],FilterIDs[2],ActivationInterval[3],MaxHits[4],MinHits[5],MinSleep[6],Blocker[7],Weight[8],ActionIDs[9],Async[10]
cgrates.org,Threshold1,FLTR_1;FLTR_ACNT_dan,2014-07-29T15:00:00Z,12,10,1s,true,10,THRESH1,true
`

	filters = `
#Tenant[0],ID[1],FilterType[2],FilterFieldName[3],FilterFieldValues[4],ActivationInterval[5]
cgrates.org,FLTR_1,*string,Account,1001;1002,2014-07-29T15:00:00Z
cgrates.org,FLTR_1,*prefix,Destination,10;20,2014-07-29T15:00:00Z
cgrates.org,FLTR_1,*rsr,,Subject(~^1.*1$);Destination(1002),
cgrates.org,FLTR_ACNT_dan,*string,Account,dan,2014-07-29T15:00:00Z
cgrates.org,FLTR_DST_DE,*destinations,Destination,DST_DE,2014-07-29T15:00:00Z
cgrates.org,FLTR_DST_NL,*destinations,Destination,DST_NL,2014-07-29T15:00:00Z
`
	sppProfiles = `
#Tenant,ID,FilterIDs,ActivationInterval,Sorting,SortingParameters,SupplierID,SupplierFilterIDs,SupplierAccountIDs,SupplierRatingPlanIDs,SupplierResourceIDs,SupplierStatIDs,SupplierWeight,SupplierBlocker,SupplierParameters,Weight
cgrates.org,SPP_1,FLTR_ACNT_dan,2014-07-29T15:00:00Z,*lowest_cost,,supplier1,FLTR_ACNT_dan,Account1;Account1_1,RPL_1,ResGroup1,Stat1,10,true,param1,20
cgrates.org,SPP_1,,,,,supplier1,,,RPL_2,ResGroup2,,10,,,
cgrates.org,SPP_1,,,,,supplier1,FLTR_DST_DE,Account2,RPL_3,ResGroup3,Stat2,10,,,
cgrates.org,SPP_1,,,,,supplier1,,,,ResGroup4,Stat3,10,,,
`
	attributeProfiles = `
#Tenant,ID,Contexts,FilterIDs,ActivationInterval,FieldName,Initial,Substitute,Append,Blocker,Weight
cgrates.org,ALS1,con1,FLTR_1,2014-07-29T15:00:00Z,Field1,Initial1,Sub1,true,true,20
cgrates.org,ALS1,con2;con3,,,Field2,Initial2,Sub2,false,,
`
	chargerProfiles = `
#Tenant,ID,FilterIDs,ActivationInterval,RunID,AttributeIDs,Weight
cgrates.org,Charger1,*string:Account:1001,2014-07-29T15:00:00Z,*rated,ATTR_1001_SIMPLEAUTH,20
`
)

var csvr *TpReader

func init() {
	csvr = NewTpReader(dm.dataDB, NewStringCSVStorage(',', destinations, timings, rates, destinationRates, ratingPlans, ratingProfiles,
		sharedGroups, actions, actionPlans, actionTriggers, accountActions, derivedCharges,
		users, aliases, resProfiles, stats, thresholds, filters, sppProfiles, attributeProfiles, chargerProfiles), testTPID, "")

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
	if err := csvr.LoadDerivedChargers(); err != nil {
		log.Print("error in LoadDerivedChargers:", err)
	}
	if err := csvr.LoadUsers(); err != nil {
		log.Print("error in LoadUsers:", err)
	}
	if err := csvr.LoadAliases(); err != nil {
		log.Print("error in LoadAliases:", err)
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
	if err := csvr.LoadThresholds(); err != nil {
		log.Print("error in LoadThresholds:", err)
	}
	if err := csvr.LoadSupplierProfiles(); err != nil {
		log.Print("error in LoadSupplierProfiles:", err)
	}
	if err := csvr.LoadAttributeProfiles(); err != nil {
		log.Print("error in LoadAttributeProfiles:", err)
	}
	if err := csvr.LoadChargerProfiles(); err != nil {
		log.Print("error in LoadChargerProfiles:", err)
	}
	csvr.WriteToDatabase(false, false, false)
	Cache.Clear(nil)
	//dm.LoadDataDBCache(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
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

func TestLoadReverseDestinations(t *testing.T) {
	eRevDsts := map[string][]string{
		"444":     []string{"EU_LANDLINE"},
		"0257":    []string{"NAT"},
		"112":     []string{"URG"},
		"49":      []string{"ALL GERMANY"},
		"+4972":   []string{"PSTN_72"},
		"999":     []string{"EXOTIC"},
		"+4970":   []string{"PSTN_70"},
		"41":      []string{"ALL GERMANY_O2"},
		"0724":    []string{"RET"},
		"0723045": []string{"SPEC"},
		"43":      []string{"GERMANY_PREMIUM ALL"},
		"0256":    []string{"NAT"},
		"+49":     []string{"NAT"},
		"+4971":   []string{"PSTN_71"},
		"447956":  []string{"DST_UK_Mobile_BIG5"},
		"0723":    []string{"RET NAT"},
	}
	if len(eRevDsts) != len(csvr.revDests) {
		t.Errorf("Expecting: %+v, received: %+v", eRevDsts, csvr.revDests)
	}
}

func TestLoadTimimgs(t *testing.T) {
	if len(csvr.timings) != 6 {
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
			&utils.DestinationRate{
				DestinationId:    "GERMANY",
				RateId:           "R1",
				Rate:             csvr.rates["R1"],
				RoundingMethod:   utils.ROUNDING_MIDDLE,
				RoundingDecimals: 4,
			},
			&utils.DestinationRate{
				DestinationId:    "GERMANY_O2",
				RateId:           "R2",
				Rate:             csvr.rates["R2"],
				RoundingMethod:   utils.ROUNDING_MIDDLE,
				RoundingDecimals: 4,
			},
			&utils.DestinationRate{
				DestinationId:    "GERMANY_PREMIUM",
				RateId:           "R2",
				Rate:             csvr.rates["R2"],
				RoundingMethod:   utils.ROUNDING_MIDDLE,
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
			&utils.DestinationRate{
				DestinationId:    "ALL",
				RateId:           "R2",
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
		TPid: testTPID,
		ID:   "RT_STD_WEEKEND",
		DestinationRates: []*utils.DestinationRate{
			&utils.DestinationRate{
				DestinationId:    "GERMANY",
				RateId:           "R2",
				Rate:             csvr.rates["R2"],
				RoundingMethod:   utils.ROUNDING_MIDDLE,
				RoundingDecimals: 4,
			},
			&utils.DestinationRate{
				DestinationId:    "GERMANY_O2",
				RateId:           "R3",
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
		TPid: testTPID,
		ID:   "P1",
		DestinationRates: []*utils.DestinationRate{
			&utils.DestinationRate{
				DestinationId:    "NAT",
				RateId:           "R4",
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
		TPid: testTPID,
		ID:   "P2",
		DestinationRates: []*utils.DestinationRate{
			&utils.DestinationRate{
				DestinationId:    "NAT",
				RateId:           "R5",
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
		TPid: testTPID,
		ID:   "T1",
		DestinationRates: []*utils.DestinationRate{
			&utils.DestinationRate{
				DestinationId:    "NAT",
				RateId:           "LANDLINE_OFFPEAK",
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
		TPid: testTPID,
		ID:   "T2",
		DestinationRates: []*utils.DestinationRate{
			&utils.DestinationRate{
				DestinationId:    "GERMANY",
				RateId:           "GBP_72",
				Rate:             csvr.rates["GBP_72"],
				RoundingMethod:   utils.ROUNDING_MIDDLE,
				RoundingDecimals: 4,
			},
			&utils.DestinationRate{
				DestinationId:    "GERMANY_O2",
				RateId:           "GBP_70",
				Rate:             csvr.rates["GBP_70"],
				RoundingMethod:   utils.ROUNDING_MIDDLE,
				RoundingDecimals: 4,
			},
			&utils.DestinationRate{
				DestinationId:    "GERMANY_PREMIUM",
				RateId:           "GBP_71",
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
	if len(csvr.ratingPlans) != 14 {
		t.Error("Failed to load rating plans: ", len(csvr.ratingPlans))
	}
	rplan := csvr.ratingPlans["STANDARD"]
	expected := &RatingPlan{
		Id: "STANDARD",
		Timings: map[string]*RITiming{
			"59a981b9": &RITiming{
				Years:     utils.Years{},
				Months:    utils.Months{},
				MonthDays: utils.MonthDays{},
				WeekDays:  utils.WeekDays{1, 2, 3, 4, 5},
				StartTime: "00:00:00",
				tag:       "WORKDAYS_00",
			},
			"2d9ca6c4": &RITiming{
				Years:     utils.Years{},
				Months:    utils.Months{},
				MonthDays: utils.MonthDays{},
				WeekDays:  utils.WeekDays{1, 2, 3, 4, 5},
				StartTime: "18:00:00",
				tag:       "WORKDAYS_18",
			},
			"ec8ed374": &RITiming{
				Years:     utils.Years{},
				Months:    utils.Months{},
				MonthDays: utils.MonthDays{},
				WeekDays:  utils.WeekDays{time.Saturday, time.Sunday},
				StartTime: "00:00:00",
				tag:       "WEEKENDS",
			},
			"83429156": &RITiming{
				Years:     utils.Years{},
				Months:    utils.Months{},
				MonthDays: utils.MonthDays{},
				WeekDays:  utils.WeekDays{},
				StartTime: "00:00:00",
				tag:       "*any",
			},
		},
		Ratings: map[string]*RIRate{
			"ebefae11": &RIRate{
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
				tag:              "R1",
			},
			"fac0138e": &RIRate{
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
				tag:              "R2",
			},
			"781bfa03": &RIRate{
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
				tag:              "R3",
			},
			"f692daa4": &RIRate{
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
				tag:              "R_URG",
			},
		},
		DestinationRates: map[string]RPRateList{
			"GERMANY": []*RPRate{
				&RPRate{
					Timing: "ec8ed374",
					Rating: "ebefae11",
					Weight: 10,
				},
				&RPRate{
					Timing: "83429156",
					Rating: "fac0138e",
					Weight: 10,
				},
				&RPRate{
					Timing: "a60bfb13",
					Rating: "fac0138e",
					Weight: 10,
				},
			},
			"GERMANY_O2": []*RPRate{
				&RPRate{
					Timing: "ec8ed374",
					Rating: "fac0138e",
					Weight: 10,
				},
				&RPRate{
					Timing: "83429156",
					Rating: "781bfa03",
					Weight: 10,
				},
				&RPRate{
					Timing: "a60bfb13",
					Rating: "781bfa03",
					Weight: 10,
				},
			},
			"GERMANY_PREMIUM": []*RPRate{
				&RPRate{
					Timing: "ec8ed374",
					Rating: "16e9ee19",
					Weight: 10,
				},
			},
			"URG": []*RPRate{
				&RPRate{
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
		Years:      utils.Years{},
		Months:     utils.Months{},
		MonthDays:  utils.MonthDays{},
		WeekDays:   utils.WeekDays{},
		StartTime:  "00:00:00",
		EndTime:    "",
		cronString: "",
		tag:        utils.ANY,
	}

	if !reflect.DeepEqual(csvr.ratingPlans["ANY_PLAN"].Timings["1323e132"], anyTiming) {
		t.Errorf("Error using *any timing in rating plans: %+v : %+v", csvr.ratingPlans["ANY_PLAN"].Timings["1323e132"], anyTiming)
	}
}

func TestLoadRatingProfiles(t *testing.T) {
	if len(csvr.ratingProfiles) != 24 {
		t.Error("Failed to load rating profiles: ", len(csvr.ratingProfiles), csvr.ratingProfiles)
	}
	rp := csvr.ratingProfiles["*out:test:0:trp"]
	expected := &RatingProfile{
		Id: "*out:test:0:trp",
		RatingPlanActivations: RatingPlanActivations{&RatingPlanActivation{
			ActivationTime:  time.Date(2013, 10, 1, 0, 0, 0, 0, time.UTC),
			RatingPlanId:    "TDRT",
			FallbackKeys:    []string{"*out:test:0:danb", "*out:test:0:rif"},
			CdrStatQueueIds: []string{""},
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
		&Action{
			Id:               "MINI",
			ActionType:       TOPUP_RESET,
			ExpirationString: UNLIMITED,
			ExtraParameters:  "",
			Weight:           10,
			Balance: &BalanceFilter{
				Type:           utils.StringPointer(utils.MONETARY),
				Uuid:           as1[0].Balance.Uuid,
				Directions:     utils.StringMapPointer(utils.NewStringMap(utils.OUT)),
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
		&Action{
			Id:               "MINI",
			ActionType:       TOPUP,
			ExpirationString: UNLIMITED,
			ExtraParameters:  "",
			Weight:           10,
			Balance: &BalanceFilter{
				Type:           utils.StringPointer(utils.VOICE),
				Uuid:           as1[1].Balance.Uuid,
				Directions:     utils.StringMapPointer(utils.NewStringMap(utils.OUT)),
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
		t.Errorf("Error loading action1: %s", utils.ToIJSON(as1))
	}
	as2 := csvr.actions["SHARED"]
	expected = []*Action{
		&Action{
			Id:               "SHARED",
			ActionType:       TOPUP,
			ExpirationString: UNLIMITED,
			Weight:           10,
			Balance: &BalanceFilter{
				Type:           utils.StringPointer(utils.MONETARY),
				Directions:     utils.StringMapPointer(utils.NewStringMap(utils.OUT)),
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
		&Action{
			Id:              "DEFEE",
			ActionType:      CDRLOG,
			ExtraParameters: `{"Category":"^ddi","MediationRunId":"^did_run"}`,
			Weight:          10,
			Balance: &BalanceFilter{
				Uuid:           as3[0].Balance.Uuid,
				Directions:     nil,
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
		&Action{
			Id:               "TOPUP_RST_GNR_1000",
			ActionType:       TOPUP_RESET,
			ExtraParameters:  `{"*voice": 60.0,"*data":1024.0,"*sms":1.0}`,
			Weight:           10,
			ExpirationString: utils.UNLIMITED,
			Balance: &BalanceFilter{
				Uuid:       asGnrc[0].Balance.Uuid,
				Type:       utils.StringPointer(utils.GENERIC),
				Directions: utils.StringMapPointer(utils.NewStringMap(utils.OUT)),
				//DestinationIDs: utils.StringMapPointer(utils.NewStringMap("*any")),
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
			"*any": &SharingParameters{
				Strategy:      "*lowest",
				RatingSubject: "",
			},
		},
	}
	if !reflect.DeepEqual(sg1, expected) {
		t.Error("Error loading shared group: ", sg1.AccountParameters["SG1"])
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
		AccountIDs: utils.StringMap{"vdf:minitsboy": true},
		ActionTimings: []*ActionTiming{
			&ActionTiming{
				Uuid: atm.ActionTimings[0].Uuid,
				Timing: &RateInterval{
					Timing: &RITiming{
						Years:     utils.Years{2012},
						Months:    utils.Months{},
						MonthDays: utils.MonthDays{},
						WeekDays:  utils.WeekDays{},
						StartTime: utils.ASAP,
					},
				},
				Weight:    10,
				ActionsID: "MINI",
			},
			&ActionTiming{
				Uuid: atm.ActionTimings[1].Uuid,
				Timing: &RateInterval{
					Timing: &RITiming{
						Years:     utils.Years{2012},
						Months:    utils.Months{},
						MonthDays: utils.MonthDays{},
						WeekDays:  utils.WeekDays{},
						StartTime: utils.ASAP,
					},
				},
				Weight:    10,
				ActionsID: "SHARED",
			},
		},
	}
	if !reflect.DeepEqual(atm, expected) {
		t.Errorf("Error loading action timing:\n%+v", atm.ActionTimings[1])
	}
}

func TestLoadActionTriggers(t *testing.T) {
	if len(csvr.actionsTriggers) != 7 {
		t.Error("Failed to load action triggers: ", len(csvr.actionsTriggers))
	}
	atr := csvr.actionsTriggers["STANDARD_TRIGGER"][0]
	expected := &ActionTrigger{
		ID:             "STANDARD_TRIGGER",
		UniqueID:       "st0",
		ThresholdType:  utils.TRIGGER_MIN_EVENT_COUNTER,
		ThresholdValue: 10,
		Balance: &BalanceFilter{
			ID:             nil,
			Type:           utils.StringPointer(utils.VOICE),
			Directions:     utils.StringMapPointer(utils.NewStringMap(utils.OUT)),
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
		t.Errorf("Error loading action trigger: %+v", utils.ToIJSON(atr.Balance))
	}
	atr = csvr.actionsTriggers["STANDARD_TRIGGER"][1]
	expected = &ActionTrigger{
		ID:             "STANDARD_TRIGGER",
		UniqueID:       "st1",
		ThresholdType:  utils.TRIGGER_MAX_BALANCE,
		ThresholdValue: 200,
		Balance: &BalanceFilter{
			Type:           utils.StringPointer(utils.VOICE),
			Directions:     utils.StringMapPointer(utils.NewStringMap(utils.OUT)),
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
			utils.VOICE: []*UnitCounter{
				&UnitCounter{
					CounterType: "*event",
					Counters: CounterFilters{
						&CounterFilter{
							Value: 0,
							Filter: &BalanceFilter{
								ID:             utils.StringPointer("st0"),
								Type:           utils.StringPointer(utils.VOICE),
								Directions:     utils.StringMapPointer(utils.NewStringMap("*out")),
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
	for i, b := range aa.UnitCounters[utils.VOICE][0].Counters {
		expected.UnitCounters[utils.VOICE][0].Counters[i].Filter.ID = b.Filter.ID
	}
	if !reflect.DeepEqual(aa.UnitCounters[utils.VOICE][0].Counters[0], expected.UnitCounters[utils.VOICE][0].Counters[0]) {
		t.Errorf("Error loading account action: %+v", utils.ToIJSON(aa.UnitCounters[utils.VOICE][0].Counters[0].Filter))
	}
	// test that it does not overwrite balances
	existing, err := dm.DataDB().GetAccount(aa.ID)
	if err != nil || len(existing.BalanceMap) != 2 {
		t.Errorf("The account was not set before load: %+v", existing)
	}
	dm.DataDB().SetAccount(aa)
	existing, err = dm.DataDB().GetAccount(aa.ID)
	if err != nil || len(existing.BalanceMap) != 2 {
		t.Errorf("The set account altered the balances: %+v", existing)
	}
}

func TestLoadDerivedChargers(t *testing.T) {
	if len(csvr.derivedChargers) != 2 {
		t.Error("Failed to load derivedChargers: ", csvr.derivedChargers)
	}
	expCharger1 := &utils.DerivedChargers{
		DestinationIDs: nil,
		Chargers: []*utils.DerivedCharger{
			&utils.DerivedCharger{RunID: "extra1", RunFilters: "^filteredHeader1/filterValue1/",
				RequestTypeField: "^prepaid", DirectionField: utils.META_DEFAULT,
				TenantField: utils.META_DEFAULT, CategoryField: utils.META_DEFAULT,
				AccountField: "rif", SubjectField: "rif", DestinationField: utils.META_DEFAULT,
				SetupTimeField: utils.META_DEFAULT, PDDField: utils.META_DEFAULT,
				AnswerTimeField: utils.META_DEFAULT, UsageField: utils.META_DEFAULT,
				SupplierField: utils.META_DEFAULT, DisconnectCauseField: utils.META_DEFAULT,
				CostField: utils.META_DEFAULT, PreRatedField: utils.META_DEFAULT},
			&utils.DerivedCharger{RunID: "extra2", RequestTypeField: utils.META_DEFAULT,
				DirectionField: utils.META_DEFAULT, TenantField: utils.META_DEFAULT,
				CategoryField: utils.META_DEFAULT, AccountField: "ivo",
				SubjectField: "ivo", DestinationField: utils.META_DEFAULT,
				SetupTimeField: utils.META_DEFAULT, PDDField: utils.META_DEFAULT,
				AnswerTimeField: utils.META_DEFAULT, UsageField: utils.META_DEFAULT,
				SupplierField: utils.META_DEFAULT, DisconnectCauseField: utils.META_DEFAULT,
				CostField: utils.META_DEFAULT, PreRatedField: utils.META_DEFAULT},
		}}
	keyCharger1 := utils.DerivedChargersKey("*out", "cgrates.org", "call", "dan", "dan")

	if !csvr.derivedChargers[keyCharger1].Equal(expCharger1) {
		t.Errorf("Expecting: %+v, received: %+v",
			expCharger1.Chargers[0], csvr.derivedChargers[keyCharger1].Chargers[0])
	}
}

func TestLoadUsers(t *testing.T) {
	if len(csvr.users) != 3 {
		t.Error("Failed to load users: ", csvr.users)
	}
	user1 := &UserProfile{
		Tenant:   "cgrates.org",
		UserName: "rif",
		Profile: map[string]string{
			"test0": "val0",
			"test1": "val1",
		},
	}

	if !reflect.DeepEqual(csvr.users[user1.GetId()], user1) {
		t.Errorf("Unexpected user %+v", csvr.users[user1.GetId()])
	}
}

func TestLoadAliases(t *testing.T) {
	if len(csvr.aliases) != 4 {
		t.Error("Failed to load aliases: ", len(csvr.aliases))
	}
	alias1 := &Alias{
		Direction: "*out",
		Tenant:    "cgrates.org",
		Category:  "call",
		Account:   "dan",
		Subject:   "dan",
		Context:   "*rating",
		Values: AliasValues{
			&AliasValue{
				DestinationId: "EU_LANDLINE",
				Pairs: AliasPairs{
					"Subject": map[string]string{
						"dan": "dan1",
						"rif": "rif1",
					},
					"Cli": map[string]string{
						"0723": "0724",
					},
				},
				Weight: 10,
			},

			&AliasValue{
				DestinationId: "GLOBAL1",
				Pairs:         AliasPairs{"Subject": map[string]string{"dan": "dan2"}},
				Weight:        20,
			},
		},
	}

	if !reflect.DeepEqual(csvr.aliases[alias1.GetId()], alias1) {
		for _, value := range csvr.aliases[alias1.GetId()].Values {
			t.Logf("Value: %+v", value)
		}
		t.Errorf("Unexpected alias %+v", csvr.aliases[alias1.GetId()])
	}
}

func TestLoadReverseAliases(t *testing.T) {
	eRevAliases := map[string][]string{
		"minuAccount*rating": []string{"*out:cgrates.org:call:remo:remo:*rating:*any", "*out:vdf:0:a1:a1:*rating:*any"},
		"dan1Subject*rating": []string{"*out:cgrates.org:call:dan:dan:*rating:EU_LANDLINE"},
		"rif1Subject*rating": []string{"*out:cgrates.org:call:dan:dan:*rating:EU_LANDLINE", "*any:*any:*any:*any:*any:*rating:*any"},
		"0724Cli*rating":     []string{"*out:cgrates.org:call:dan:dan:*rating:EU_LANDLINE"},
		"dan2Subject*rating": []string{"*out:cgrates.org:call:dan:dan:*rating:GLOBAL1"},
		"dan1Account*rating": []string{"*any:*any:*any:*any:*any:*rating:*any"},
		"minuSubject*rating": []string{"*out:cgrates.org:call:remo:remo:*rating:*any", "*out:vdf:0:a1:a1:*rating:*any"},
	}
	if len(eRevAliases) != len(csvr.revAliases) {
		t.Errorf("Expecting: %+v, received: %+v", eRevAliases, csvr.revAliases)
	}
}

func TestLoadResourceProfiles(t *testing.T) {
	eResProfiles := map[utils.TenantID]*utils.TPResource{
		utils.TenantID{Tenant: "cgrates.org", ID: "ResGroup21"}: &utils.TPResource{
			TPid:      testTPID,
			Tenant:    "cgrates.org",
			ID:        "ResGroup21",
			FilterIDs: []string{"FLTR_1"},
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
		utils.TenantID{Tenant: "cgrates.org", ID: "ResGroup22"}: &utils.TPResource{
			TPid:      testTPID,
			Tenant:    "cgrates.org",
			ID:        "ResGroup22",
			FilterIDs: []string{"FLTR_ACNT_dan"},
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
		t.Errorf("Expecting: %+v, received: %+v", eResProfiles[resKey], csvr.resProfiles[resKey])
	}
}

func TestLoadStatQueueProfiles(t *testing.T) {
	eStats := map[utils.TenantID]*utils.TPStats{
		utils.TenantID{Tenant: "cgrates.org", ID: "TestStats"}: &utils.TPStats{
			Tenant:    "cgrates.org",
			TPid:      testTPID,
			ID:        "TestStats",
			FilterIDs: []string{"FLTR_1"},
			ActivationInterval: &utils.TPActivationInterval{
				ActivationTime: "2014-07-29T15:00:00Z",
			},
			QueueLength: 100,
			TTL:         "1s",
			Metrics: []*utils.MetricWithParams{
				&utils.MetricWithParams{
					MetricID:   "*sum:Value",
					Parameters: "Value",
				},
				&utils.MetricWithParams{
					MetricID:   "*average:Value",
					Parameters: "Value",
				},
				&utils.MetricWithParams{
					MetricID:   "*sum:Usage",
					Parameters: "Usage",
				},
			},
			ThresholdIDs: []string{"Th1", "Th2"},
			Blocker:      true,
			Stored:       true,
			Weight:       20,
			MinItems:     2,
		},
		utils.TenantID{Tenant: "cgrates.org", ID: "TestStats2"}: &utils.TPStats{
			Tenant:    "cgrates.org",
			TPid:      testTPID,
			ID:        "TestStats2",
			FilterIDs: []string{"FLTR_1"},
			ActivationInterval: &utils.TPActivationInterval{
				ActivationTime: "2014-07-29T15:00:00Z",
			},
			QueueLength: 100,
			TTL:         "1s",
			Metrics: []*utils.MetricWithParams{
				&utils.MetricWithParams{
					MetricID:   "*sum:Value",
					Parameters: "Value",
				},
				&utils.MetricWithParams{
					MetricID:   "*average:Value",
					Parameters: "Value",
				},
				&utils.MetricWithParams{
					MetricID:   "*sum:Usage",
					Parameters: "Usage",
				},
				&utils.MetricWithParams{
					MetricID:   "*average:Usage",
					Parameters: "Usage",
				},
				&utils.MetricWithParams{
					MetricID:   "*sum:Cost",
					Parameters: "Cost",
				},
				&utils.MetricWithParams{
					MetricID:   "*average:Cost",
					Parameters: "Cost",
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
		utils.TenantID{Tenant: "cgrates.org", ID: "TestStats"},
		utils.TenantID{Tenant: "cgrates.org", ID: "TestStats2"},
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
			t.Errorf("Expecting: %s, received: %s",
				utils.ToJSON(eStats[stKey].Metrics),
				utils.ToJSON(csvr.sqProfiles[stKey].Metrics))
		}
	}
}

func TestLoadThresholdProfiles(t *testing.T) {
	eThresholds := map[utils.TenantID]*utils.TPThreshold{
		utils.TenantID{Tenant: "cgrates.org", ID: "Threshold1"}: &utils.TPThreshold{
			TPid:      testTPID,
			Tenant:    "cgrates.org",
			ID:        "Threshold1",
			FilterIDs: []string{"FLTR_1", "FLTR_ACNT_dan"},
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
	eThresholdReverse := map[utils.TenantID]*utils.TPThreshold{
		utils.TenantID{Tenant: "cgrates.org", ID: "Threshold1"}: &utils.TPThreshold{
			TPid:      testTPID,
			Tenant:    "cgrates.org",
			ID:        "Threshold1",
			FilterIDs: []string{"FLTR_ACNT_dan", "FLTR_1"},
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
		utils.TenantID{Tenant: "cgrates.org", ID: "FLTR_1"}: &utils.TPFilterProfile{
			TPid:   testTPID,
			Tenant: "cgrates.org",
			ID:     "FLTR_1",
			Filters: []*utils.TPFilter{
				&utils.TPFilter{
					FieldName: "Account",
					Type:      "*string",
					Values:    []string{"1001", "1002"},
				},
				&utils.TPFilter{
					FieldName: "Destination",
					Type:      MetaPrefix,
					Values:    []string{"10", "20"},
				},
				&utils.TPFilter{
					FieldName: "",
					Type:      "*rsr",
					Values:    []string{"Subject(~^1.*1$)", "Destination(1002)"},
				},
			},
			ActivationInterval: &utils.TPActivationInterval{
				ActivationTime: "2014-07-29T15:00:00Z",
			},
		},
		utils.TenantID{Tenant: "cgrates.org", ID: "FLTR_ACNT_dan"}: &utils.TPFilterProfile{
			TPid:   testTPID,
			Tenant: "cgrates.org",
			ID:     "FLTR_ACNT_dan",
			Filters: []*utils.TPFilter{
				&utils.TPFilter{
					FieldName: "Account",
					Type:      "*string",
					Values:    []string{"dan"},
				},
			},
			ActivationInterval: &utils.TPActivationInterval{
				ActivationTime: "2014-07-29T15:00:00Z",
			},
		},
		utils.TenantID{Tenant: "cgrates.org", ID: "FLTR_DST_DE"}: &utils.TPFilterProfile{
			TPid:   testTPID,
			Tenant: "cgrates.org",
			ID:     "FLTR_DST_DE",
			Filters: []*utils.TPFilter{
				&utils.TPFilter{
					FieldName: "Destination",
					Type:      "*destinations",
					Values:    []string{"DST_DE"},
				},
			},
			ActivationInterval: &utils.TPActivationInterval{
				ActivationTime: "2014-07-29T15:00:00Z",
			},
		},
		utils.TenantID{Tenant: "cgrates.org", ID: "FLTR_DST_NL"}: &utils.TPFilterProfile{
			TPid:   testTPID,
			Tenant: "cgrates.org",
			ID:     "FLTR_DST_NL",
			Filters: []*utils.TPFilter{
				&utils.TPFilter{
					FieldName: "Destination",
					Type:      "*destinations",
					Values:    []string{"DST_NL"},
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
		t.Errorf("Expecting: %+v, received: %+v", eFilters[fltrKey], csvr.filters[fltrKey])
	}
}

func TestLoadSupplierProfiles(t *testing.T) {
	eSppProfiles := map[utils.TenantID]*utils.TPSupplierProfile{
		utils.TenantID{Tenant: "cgrates.org", ID: "SPP_1"}: &utils.TPSupplierProfile{
			TPid:      testTPID,
			Tenant:    "cgrates.org",
			ID:        "SPP_1",
			FilterIDs: []string{"FLTR_ACNT_dan"},
			ActivationInterval: &utils.TPActivationInterval{
				ActivationTime: "2014-07-29T15:00:00Z",
			},
			Sorting:           "*lowest_cost",
			SortingParameters: []string{},
			Suppliers: []*utils.TPSupplier{
				&utils.TPSupplier{
					ID:                 "supplier1",
					FilterIDs:          []string{"FLTR_ACNT_dan", "FLTR_DST_DE"},
					AccountIDs:         []string{"Account1", "Account1_1", "Account2"},
					RatingPlanIDs:      []string{"RPL_1", "RPL_2", "RPL_3"},
					ResourceIDs:        []string{"ResGroup1", "ResGroup2", "ResGroup3", "ResGroup4"},
					StatIDs:            []string{"Stat1", "Stat2", "Stat3"},
					Weight:             10,
					Blocker:            true,
					SupplierParameters: "param1",
				},
			},
			Weight: 20,
		},
	}
	resKey := utils.TenantID{Tenant: "cgrates.org", ID: "SPP_1"}
	if len(csvr.sppProfiles) != len(eSppProfiles) {
		t.Errorf("Failed to load SupplierProfiles: %s", utils.ToIJSON(csvr.sppProfiles))
	} else if !reflect.DeepEqual(eSppProfiles[resKey], csvr.sppProfiles[resKey]) {
		t.Errorf("Expecting: %+v, received: %+v", eSppProfiles[resKey], csvr.sppProfiles[resKey])
	}
}

func TestLoadAttributeProfiles(t *testing.T) {
	eAttrProfiles := map[utils.TenantID]*utils.TPAttributeProfile{
		utils.TenantID{Tenant: "cgrates.org", ID: "ALS1"}: &utils.TPAttributeProfile{
			TPid:      testTPID,
			Tenant:    "cgrates.org",
			ID:        "ALS1",
			Contexts:  []string{"con1", "con2", "con3"},
			FilterIDs: []string{"FLTR_1"},
			ActivationInterval: &utils.TPActivationInterval{
				ActivationTime: "2014-07-29T15:00:00Z",
			},
			Attributes: []*utils.TPAttribute{
				&utils.TPAttribute{
					FieldName:  "Field1",
					Initial:    "Initial1",
					Substitute: "Sub1",
					Append:     true,
				},
				&utils.TPAttribute{
					FieldName:  "Field2",
					Initial:    "Initial2",
					Substitute: "Sub2",
					Append:     false,
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
		utils.TenantID{Tenant: "cgrates.org", ID: "Charger1"}: &utils.TPChargerProfile{
			TPid:      testTPID,
			Tenant:    "cgrates.org",
			ID:        "Charger1",
			FilterIDs: []string{"*string:Account:1001"},
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

func TestLoadResource(t *testing.T) {
	eResources := []*utils.TenantID{
		&utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "ResGroup21",
		},
		&utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "ResGroup22",
		},
	}
	if len(csvr.resources) != len(eResources) {
		t.Errorf("Failed to load resources expecting 2 but received : %+v", len(csvr.resources))
	}
}

func TestLoadstatQueues(t *testing.T) {
	eStatQueues := []*utils.TenantID{
		&utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "TestStats",
		},
		&utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "TestStats2",
		},
	}
	if len(csvr.statQueues) != len(eStatQueues) {
		t.Errorf("Failed to load statQueues: %s", utils.ToIJSON(csvr.statQueues))
	}
}

func TestLoadThresholds(t *testing.T) {
	eThresholds := []*utils.TenantID{
		&utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "Threshold1",
		},
	}
	if len(csvr.thresholds) != len(eThresholds) {
		t.Errorf("Failed to load thresholds: %s", utils.ToIJSON(csvr.thresholds))
	}
}
