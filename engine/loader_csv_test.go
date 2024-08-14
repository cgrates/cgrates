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

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func TestLoaderCSV(t *testing.T) {
	/* NOTE: Destinations have not been ported. Keep this just in case.
	   	DestinationsCSVContent := `
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
	*/

	/* TODO: add a test for loading Actions
	   	ActionsCSVContent := `
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
	   FILTER,*topup,,"{""*and"":[{""Value"":{""*lt"":0}},{""Id"":{""*eq"":""*default""}}]}",bfree,*monetary,,,,,*unlimited,,20,10,false,false,10
	   EXP,*topup,,,,*voice,,,,,*monthly,*any,300s,10,false,false,10
	   NOEXP,*topup,,,,*voice,,,,,*unlimited,*any,50s,10,false,false,10
	   VF,*debit,,,,*monetary,,,,,*unlimited,*any,"{""Method"":""*incremental"",""Params"":{""Units"":10, ""Interval"":""month"", ""Increment"":""day""}}",10,false,false,10
	   TOPUP_RST_GNR_1000,*topup_reset,"{""*voice"": 60.0,""*data"":1024.0,""*sms"":1.0}",,,*generic,,*any,,,*unlimited,,1000,20,false,false,10
	   `
	*/

	ResourcesCSVContent := `
#Tenant[0],Id[1],FilterIDs[2],Weights[3],TTL[4],Limit[5],AllocationMessage[6],Blocker[7],Stored[8],Thresholds[9]
cgrates.org,ResGroup21,*string:~*req.Account:1001,;10,1s,2,call,true,true,
cgrates.org,ResGroup22,*string:~*req.Account:dan,;10,3600s,2,premium_call,true,true,
`
	StatsCSVContent := `
#Tenant[0],Id[1],FilterIDs[2],Weights[3],Blockers[4],QueueLength[5],TTL[6],MinItems[7],Stored[8],ThresholdIDs[9],Metrics[10],MetricFilterIDs[11],MetricBlockers[12]
cgrates.org,TestStats,*string:~*req.Account:1001,;20,;true,100,1s,2,true,Th1;Th2,*sum#~*req.Value;*average#~*req.Value,,
cgrates.org,TestStats,,;20,;true,,,2,true,,*sum#~*req.Usage,,
cgrates.org,TestStats2,FLTR_1,;20,;true,100,1s,2,true,Th,*sum#~*req.Value;*sum#~*req.Usage;*average#~*req.Value;*average#~*req.Usage,,
cgrates.org,TestStats2,,;20,;true,,,2,true,,*sum#~*req.Cost;*average#~*req.Cost,,
`
	RankingsCSVContent := `
#Tenant[0],Id[1],QueryInterval[2],StatIDs[2],MetricIDs[3],Sorting[4],SortingParameters[5],ThresholdIDs[6]
cgrates.org,Ranking1,15m,Stats2;Stats3;Stats4,Metric1;Metric3,*asc,,THD1;THD2
`
	TrendsCSVContent := `
#Tenant[0],Id[1],QueryInterval[2],StatID[3],QueueLength[4],TTL[5],PurgeFilterIDs[6],Trend[7],ThresholdIDs[7]
cgrates.org,TREND1,5m,Stats2,-1,-1,,*average,TD1;THD2
`
	ThresholdsCSVContent := `
#Tenant[0],Id[1],FilterIDs[2],Weights[3],MaxHits[4],MinHits[5],MinSleep[6],Blocker[7],ActionProfileIDs[8],Async[9]
cgrates.org,Threshold1,*string:~*req.Account:1001;*string:~*req.RunID:*default,;10,12,10,1s,true,THRESH1,true
`

	FiltersCSVContent := `
#Tenant[0],ID[1],Type[2],Element[3],Values[4]
cgrates.org,FLTR_1,*string,~*req.Account,1001;1002
cgrates.org,FLTR_1,*prefix,~*req.Destination,10;20
cgrates.org,FLTR_1,*rsr,~*req.Subject,~^1.*1$
cgrates.org,FLTR_1,*rsr,~*req.Destination,1002
cgrates.org,FLTR_ACNT_dan,*string,~*req.Account,dan
cgrates.org,FLTR_DST_DE,*destinations,~*req.Destination,DST_DE
cgrates.org,FLTR_DST_NL,*destinations,~*req.Destination,DST_NL
`
	RoutesCSVContent := `
#Tenant[0],ID[1],FilterIDs[2],Weights[3],Blockers[4],Sorting[5],SortingParameters[6],RouteID[7],RouteFilterIDs[8],RouteAccountIDs[9],RouteRateProfileIDs[10],RouteResourceIDs[11],RouteStatIDs[12],RouteWeights[13],RouteBlockers[14],RouteParameters[15]
cgrates.org,RoutePrf1,*string:~*req.Account:dan,;20,,*lc,,route1,FLTR_ACNT_dan,Account1;Account1_1,RPL_1,ResGroup1,Stat1,;10,;true,param1
cgrates.org,RoutePrf1,,,,,,route1,,,RPL_2,ResGroup2,,;10,,
cgrates.org,RoutePrf1,,,,,,route1,FLTR_DST_DE,Account2,RPL_3,ResGroup3,Stat2,;10,,
cgrates.org,RoutePrf1,,,,,,route1,,,,ResGroup4,Stat3,;10,,
`
	AttributesCSVContent := `
#Tenant,ID,FilterIDs,Weights,Blockers,AttributeFilterIDs,AttributeBlockers,Path,Type,Value
cgrates.org,ALS1,*string:~*req.Account:1001;*string:~*opts.*context:con1,;20,;true,*string:~*req.Field1:Initial,;true,*req.Field1,*variable,Sub1
cgrates.org,ALS1,*string:~*opts.*context:con2|con3,,,,*string:~*req.Account:1002;true,*req.Field2,*variable,Sub2
`
	ChargersCSVContent := `
#Tenant,ID,FilterIDs,Weights,Blockers,RunID,AttributeIDs
cgrates.org,Charger1,*string:~*req.Account:1001,;20,,*rated,ATTR_1001_SIMPLEAUTH
`
	DispatcherCSVContent := `
#Tenant,ID,FilterIDs,Weight,Strategy,StrategyParameters,ConnID,ConnFilterIDs,ConnWeight,ConnBlocker,ConnParameters
cgrates.org,D1,*string:~*req.Account:1001,20,*first,,C1,*gt:~*req.Usage:10,10,false,192.168.56.203
cgrates.org,D1,,,*first,,C2,*lt:~*req.Usage:10,10,false,192.168.56.204
`
	DispatcherHostCSVContent := `
#Tenant[0],ID[1],Address[2],Transport[3],ConnectAttempts[4],Reconnects[5],MaxReconnectInterval[6],ConnectTimeout[7],ReplyTimeout[8],Tls[9],ClientKey[10],ClientCertificate[11],CaCertificate[12]
cgrates.org,ALL,127.0.0.1:6012,*json,1,3,5m,1m,2m,false,,,
`
	RateProfileCSVContent := `
#Tenant,ID,FilterIDs,Weights,MinCost,MaxCost,MaxCostStrategy,RateID,RateFilterIDs,RateActivationStart,RateWeights,RateBlocker,RateIntervalStart,RateFixedFee,RateRecurrentFee,RateUnit,RateIncrement
cgrates.org,RP1,*string:~*req.Subject:1001,;0,0.1,0.6,*free,RT_WEEK,,"* * * * 1-5",;0,false,0s,0,0.12,1m,1m
cgrates.org,RP1,,,,,,RT_WEEK,,,,,1m,1.234,0.06,1m,1s
cgrates.org,RP1,,,,,,RT_WEEKEND,,"* * * * 0,6",;10,false,0s,0.089,0.06,1m,1s
cgrates.org,RP1,,,,,,RT_CHRISTMAS,,* * 24 12 *,;30,false,0s,0.0564,0.06,1m,1s
`
	ActionProfileCSVContent := `
#Tenant,ID,FilterIDs,Weights,Blockers,Schedule,TargetType,TargetIDs,ActionID,ActionFilterIDs,ActionTTL,ActionType,ActionOpts,ActionPath,ActionValue
cgrates.org,ONE_TIME_ACT,,;10,,*asap,*accounts,1001;1002,TOPUP,,0s,*add_balance,,*balance.TestBalance.Value,10
cgrates.org,ONE_TIME_ACT,,,,,,,SET_BALANCE_TEST_DATA,,0s,*set_balance,,*balance.TestDataBalance.Type,*data
cgrates.org,ONE_TIME_ACT,,,,,,,TOPUP_TEST_DATA,,0s,*add_balance,,*balance.TestDataBalance.Value,1024
cgrates.org,ONE_TIME_ACT,,,,,,,SET_BALANCE_TEST_VOICE,,0s,*set_balance,,*balance.TestVoiceBalance.Type,*voice
cgrates.org,ONE_TIME_ACT,,,,,,,TOPUP_TEST_VOICE,,0s,*add_balance,,*balance.TestVoiceBalance.Value,15m15s
cgrates.org,ONE_TIME_ACT,,,,,,,TOPUP_TEST_VOICE,,0s,*add_balance,,*balance.TestVoiceBalance2.Value,15m15s
`

	AccountCSVContent := `
#Tenant,ID,FilterIDs,Weights,Blockers,Opts,BalanceID,BalanceFilterIDs,BalanceWeights,BalanceBlockers,BalanceType,BalanceUnits,BalanceUnitFactors,BalanceOpts,BalanceCostIncrements,BalanceAttributeIDs,BalanceRateProfileIDs,ThresholdIDs
cgrates.org,1001,,;20,,,MonetaryBalance,,;10,,*monetary,14,fltr1&fltr2;100;fltr3;200,,fltr1&fltr2;1.3;2.3;3.3,attr1;attr2,,*none
cgrates.org,1001,,,,,VoiceBalance,,;10,*string:~*req.Destination:1002;true;;false,*voice,1h,,,,,,
`

	testTPID := "LoaderCSVTests"
	csvr, err := NewTpReader(NewInternalDB(nil, nil, config.CgrConfig().DataDbCfg().Items), NewStringCSVStorage(utils.CSVSep,
		ResourcesCSVContent, StatsCSVContent, RankingsCSVContent, TrendsCSVContent, ThresholdsCSVContent, FiltersCSVContent,
		RoutesCSVContent, AttributesCSVContent, ChargersCSVContent, DispatcherCSVContent,
		DispatcherHostCSVContent, RateProfileCSVContent, ActionProfileCSVContent, AccountCSVContent), testTPID, "", nil, nil, false)
	if err != nil {
		log.Print("error when creating TpReader:", err)
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
	if err := csvr.LoadRateProfiles(); err != nil {
		log.Print("error in LoadRateProfiles:", err)
	}
	if err := csvr.LoadActionProfiles(); err != nil {
		log.Print("error in LoadActionProfiles: ", err)
	}
	if err := csvr.LoadAccounts(); err != nil {
		log.Print("error in LoadActionProfiles: ", err)
	}
	if err := csvr.WriteToDatabase(false, false); err != nil {
		log.Print("error when writing into database ", err)
	}

	t.Run("load ResourceProfiles", func(t *testing.T) {
		eResProfiles := map[utils.TenantID]*utils.TPResourceProfile{
			{Tenant: "cgrates.org", ID: "ResGroup21"}: {
				TPid:              testTPID,
				Tenant:            "cgrates.org",
				ID:                "ResGroup21",
				FilterIDs:         []string{"*string:~*req.Account:1001"},
				UsageTTL:          "1s",
				AllocationMessage: "call",
				Weights:           ";10",
				Limit:             "2",
				Blocker:           true,
				Stored:            true,
			},
			{Tenant: "cgrates.org", ID: "ResGroup22"}: {
				TPid:              testTPID,
				Tenant:            "cgrates.org",
				ID:                "ResGroup22",
				FilterIDs:         []string{"*string:~*req.Account:dan"},
				UsageTTL:          "3600s",
				AllocationMessage: "premium_call",
				Blocker:           true,
				Stored:            true,
				Weights:           ";10",
				Limit:             "2",
			},
		}
		resKey := utils.TenantID{Tenant: "cgrates.org", ID: "ResGroup21"}
		if len(csvr.resProfiles) != len(eResProfiles) {
			t.Errorf("Failed to load ResourceProfiles: %s", utils.ToIJSON(csvr.resProfiles))
		} else if !reflect.DeepEqual(eResProfiles[resKey], csvr.resProfiles[resKey]) {
			t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eResProfiles[resKey]), utils.ToJSON(csvr.resProfiles[resKey]))
		}
	})

	t.Run("load StatProfiles", func(t *testing.T) {
		eStats := map[utils.TenantID]*utils.TPStatProfile{
			{Tenant: "cgrates.org", ID: "TestStats"}: {
				Tenant:      "cgrates.org",
				TPid:        testTPID,
				ID:          "TestStats",
				FilterIDs:   []string{"*string:~*req.Account:1001", "*ai:~*req.AnswerTime:2014-07-29T15:00:00Z"},
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
				Blockers:     ";true",
				Stored:       true,
				Weights:      ";20",
				MinItems:     2,
			},
			{Tenant: "cgrates.org", ID: "TestStats2"}: {
				Tenant:      "cgrates.org",
				TPid:        testTPID,
				ID:          "TestStats2",
				FilterIDs:   []string{"FLTR_1", "*ai:~*req.AnswerTime:2014-07-29T15:00:00Z"},
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
				Blockers:     ";true",
				Stored:       true,
				Weights:      ";20",
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
	})

	t.Run("load ThresholdProfiles", func(t *testing.T) {
		eThresholds := map[utils.TenantID]*utils.TPThresholdProfile{
			{Tenant: "cgrates.org", ID: "Threshold1"}: {
				TPid:             testTPID,
				Tenant:           "cgrates.org",
				ID:               "Threshold1",
				FilterIDs:        []string{"*string:~*req.Account:1001", "*string:~*req.RunID:*default"},
				MaxHits:          12,
				MinHits:          10,
				MinSleep:         "1s",
				Blocker:          true,
				Weights:          ";10",
				ActionProfileIDs: []string{"THRESH1"},
				Async:            true,
			},
		}
		eThresholdReverse := map[utils.TenantID]*utils.TPThresholdProfile{
			{Tenant: "cgrates.org", ID: "Threshold1"}: {
				TPid:             testTPID,
				Tenant:           "cgrates.org",
				ID:               "Threshold1",
				FilterIDs:        []string{"*string:~*req.RunID:*default", "*string:~*req.Account:1001"},
				MaxHits:          12,
				MinHits:          10,
				MinSleep:         "1s",
				Blocker:          true,
				Weights:          ";10",
				ActionProfileIDs: []string{"THRESH1"},
				Async:            true,
			},
		}
		thkey := utils.TenantID{Tenant: "cgrates.org", ID: "Threshold1"}
		if len(csvr.thProfiles) != len(eThresholds) {
			t.Errorf("Failed to load ThresholdProfiles: %s", utils.ToIJSON(csvr.thProfiles))
		} else if !reflect.DeepEqual(eThresholds[thkey], csvr.thProfiles[thkey]) &&
			!reflect.DeepEqual(eThresholdReverse[thkey], csvr.thProfiles[thkey]) {
			t.Errorf("Expecting: %+v , %+v , received: %+v", utils.ToJSON(eThresholds[thkey]),
				utils.ToJSON(eThresholdReverse[thkey]), utils.ToJSON(csvr.thProfiles[thkey]))
		}
	})

	t.Run("load FilterProfiles", func(t *testing.T) {
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
			},
			{Tenant: "cgrates.org", ID: "FLTR_DST_DE"}: {
				TPid:   testTPID,
				Tenant: "cgrates.org",
				ID:     "FLTR_DST_DE",
				Filters: []*utils.TPFilter{
					{
						Element: "~*req.Destination",
						Values:  []string{"DST_DE"},
					},
				},
			},
			{Tenant: "cgrates.org", ID: "FLTR_DST_NL"}: {
				TPid:   testTPID,
				Tenant: "cgrates.org",
				ID:     "FLTR_DST_NL",
				Filters: []*utils.TPFilter{
					{
						Element: "~*req.Destination",
						Values:  []string{"DST_NL"},
					},
				},
			},
		}
		fltrKey := utils.TenantID{Tenant: "cgrates.org", ID: "FLTR_1"}
		if len(csvr.filters) != len(eFilters) {
			t.Errorf("Failed to load Filters: %s", utils.ToIJSON(csvr.filters))
		} else if !reflect.DeepEqual(eFilters[fltrKey], csvr.filters[fltrKey]) {
			t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eFilters[fltrKey]), utils.ToJSON(csvr.filters[fltrKey]))
		}
	})

	t.Run("load RouteProfiles", func(t *testing.T) {
		eSppProfile := &utils.TPRouteProfile{
			TPid:      testTPID,
			Tenant:    "cgrates.org",
			ID:        "RoutePrf1",
			FilterIDs: []string{"*string:~*req.Account:dan"},
			Sorting:   utils.MetaLC,
			Routes: []*utils.TPRoute{
				{
					ID:              "route1",
					FilterIDs:       []string{"FLTR_ACNT_dan"},
					AccountIDs:      []string{"Account1", "Account1_1"},
					RateProfileIDs:  []string{"RPL_1"},
					ResourceIDs:     []string{"ResGroup1"},
					StatIDs:         []string{"Stat1"},
					Weights:         ";10",
					Blockers:        ";true",
					RouteParameters: "param1",
				},
				{
					ID:              "route1",
					RateProfileIDs:  []string{"RPL_2"},
					ResourceIDs:     []string{"ResGroup2", "ResGroup4"},
					StatIDs:         []string{"Stat3"},
					Weights:         ";10",
					RouteParameters: utils.EmptyString,
				},
				{
					ID:              "route1",
					FilterIDs:       []string{"FLTR_DST_DE"},
					AccountIDs:      []string{"Account2"},
					RateProfileIDs:  []string{"RPL_3"},
					ResourceIDs:     []string{"ResGroup3"},
					StatIDs:         []string{"Stat2"},
					Weights:         ";10",
					RouteParameters: utils.EmptyString,
				},
			},
			Weights: ";20",
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
	})

	t.Run("load AttributeProfiles", func(t *testing.T) {
		eAttrProfiles := map[utils.TenantID]*utils.TPAttributeProfile{
			{Tenant: "cgrates.org", ID: "ALS1"}: {
				TPid:      testTPID,
				Tenant:    "cgrates.org",
				ID:        "ALS1",
				FilterIDs: []string{"*string:~*opts.*context:con1", "*string:~*opts.*context:con2|con3", "*string:~*req.Account:1001"},
				Attributes: []*utils.TPAttribute{
					{
						FilterIDs: []string{"*string:~*req.Field1:Initial"},
						Blockers:  ";true",
						Path:      utils.MetaReq + utils.NestingSep + "Field1",
						Type:      utils.MetaVariable,
						Value:     "Sub1",
					},
					{
						FilterIDs: []string{},
						Blockers:  "*string:~*req.Account:1002;true",
						Path:      utils.MetaReq + utils.NestingSep + "Field2",
						Type:      utils.MetaVariable,
						Value:     "Sub2",
					},
				},
				Blockers: ";true",
				Weights:  ";20",
			},
		}

		resKey := utils.TenantID{Tenant: "cgrates.org", ID: "ALS1"}
		sort.Strings(csvr.attributeProfiles[resKey].FilterIDs)
		if len(csvr.attributeProfiles) != len(eAttrProfiles) {
			t.Errorf("Failed to load attributeProfiles: %s", utils.ToIJSON(csvr.attributeProfiles))
		} else if !reflect.DeepEqual(eAttrProfiles[resKey].Tenant, csvr.attributeProfiles[resKey].Tenant) {
			t.Errorf("Expecting: %+v, received: %+v", eAttrProfiles[resKey].Tenant, csvr.attributeProfiles[resKey].Tenant)
		} else if !reflect.DeepEqual(eAttrProfiles[resKey].ID, csvr.attributeProfiles[resKey].ID) {
			t.Errorf("Expecting: %+v, received: %+v", eAttrProfiles[resKey].ID, csvr.attributeProfiles[resKey].ID)
		} else if !reflect.DeepEqual(eAttrProfiles[resKey].FilterIDs, csvr.attributeProfiles[resKey].FilterIDs) {
			t.Errorf("Expecting: %+v, received: %+v", eAttrProfiles[resKey].FilterIDs, csvr.attributeProfiles[resKey].FilterIDs)
		} else if !reflect.DeepEqual(eAttrProfiles[resKey].Attributes, csvr.attributeProfiles[resKey].Attributes) {
			t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eAttrProfiles[resKey].Attributes), utils.ToJSON(csvr.attributeProfiles[resKey].Attributes))
		} else if !reflect.DeepEqual(eAttrProfiles[resKey].Blockers, csvr.attributeProfiles[resKey].Blockers) {
			t.Errorf("Expecting: %+v, received: %+v", eAttrProfiles[resKey].Blockers, csvr.attributeProfiles[resKey].Blockers)
		}
	})

	t.Run("load ChargerProfiles", func(t *testing.T) {
		eChargerProfiles := map[utils.TenantID]*utils.TPChargerProfile{
			{Tenant: "cgrates.org", ID: "Charger1"}: {
				TPid:         testTPID,
				Tenant:       "cgrates.org",
				ID:           "Charger1",
				FilterIDs:    []string{"*string:~*req.Account:1001"},
				RunID:        "*rated",
				AttributeIDs: []string{"ATTR_1001_SIMPLEAUTH"},
				Weights:      ";20",
			},
		}
		cppKey := utils.TenantID{Tenant: "cgrates.org", ID: "Charger1"}
		if len(csvr.chargerProfiles) != len(eChargerProfiles) {
			t.Errorf("Failed to load chargerProfiles: %s", utils.ToIJSON(csvr.chargerProfiles))
		} else if !reflect.DeepEqual(eChargerProfiles[cppKey], csvr.chargerProfiles[cppKey]) {
			t.Errorf("Expecting: %+v, received: %+v", eChargerProfiles[cppKey], csvr.chargerProfiles[cppKey])
		}
	})

	t.Run("load DispatcherProfiles", func(t *testing.T) {
		eDispatcherProfiles := &utils.TPDispatcherProfile{
			TPid:      testTPID,
			Tenant:    "cgrates.org",
			ID:        "D1",
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Strategy:  utils.MetaFirst,
			Weight:    20,
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
	})

	t.Run("load RateProfiles", func(t *testing.T) {
		eRatePrf := &utils.TPRateProfile{
			TPid:            testTPID,
			Tenant:          "cgrates.org",
			ID:              "RP1",
			FilterIDs:       []string{"*string:~*req.Subject:1001"},
			Weights:         ";0",
			MinCost:         0.1,
			MaxCost:         0.6,
			MaxCostStrategy: "*free",
			Rates: map[string]*utils.TPRate{
				"RT_WEEK": {
					ID:              "RT_WEEK",
					Weights:         ";0",
					ActivationTimes: "* * * * 1-5",
					IntervalRates: []*utils.TPIntervalRate{
						{
							IntervalStart: "0s",
							RecurrentFee:  0.12,
							Unit:          "1m",
							Increment:     "1m",
						},
						{
							IntervalStart: "1m",
							FixedFee:      1.234,
							RecurrentFee:  0.06,
							Unit:          "1m",
							Increment:     "1s",
						},
					},
				},
				"RT_WEEKEND": {
					ID:              "RT_WEEKEND",
					Weights:         ";10",
					ActivationTimes: "* * * * 0,6",
					IntervalRates: []*utils.TPIntervalRate{
						{
							IntervalStart: "0s",
							FixedFee:      0.089,
							RecurrentFee:  0.06,
							Unit:          "1m",
							Increment:     "1s",
						},
					},
				},
				"RT_CHRISTMAS": {
					ID:              "RT_CHRISTMAS",
					Weights:         ";30",
					ActivationTimes: "* * 24 12 *",
					IntervalRates: []*utils.TPIntervalRate{
						{
							IntervalStart: "0s",
							FixedFee:      0.0564,
							RecurrentFee:  0.06,
							Unit:          "1m",
							Increment:     "1s",
						},
					},
				},
			},
		}
		if len(csvr.rateProfiles) != 1 {
			t.Errorf("Failed to load rateProfiles: %s", utils.ToIJSON(csvr.rateProfiles))
		}
		dppKey := utils.TenantID{Tenant: "cgrates.org", ID: "RP1"}
		sort.Slice(csvr.rateProfiles[dppKey].Rates["RT_WEEK"].IntervalRates, func(i, j int) bool {
			return csvr.rateProfiles[dppKey].Rates["RT_WEEK"].IntervalRates[i].IntervalStart < csvr.rateProfiles[dppKey].Rates["RT_WEEK"].IntervalRates[j].IntervalStart
		})

		if !reflect.DeepEqual(eRatePrf, csvr.rateProfiles[dppKey]) {
			t.Errorf("Expecting: %+v,\n received: %+v",
				utils.ToJSON(eRatePrf), utils.ToJSON(csvr.rateProfiles[dppKey]))
		}
	})

	t.Run("load ActionProfiles", func(t *testing.T) {
		expected := &utils.TPActionProfile{
			TPid:     testTPID,
			Tenant:   "cgrates.org",
			ID:       "ONE_TIME_ACT",
			Weights:  ";10",
			Schedule: utils.MetaASAP,
			Targets: []*utils.TPActionTarget{
				{
					TargetType: utils.MetaAccounts,
					TargetIDs:  []string{"1001", "1002"},
				},
			},
			Actions: []*utils.TPAPAction{
				{
					ID:   "TOPUP",
					TTL:  "0s",
					Type: "*add_balance",
					Diktats: []*utils.TPAPDiktat{{
						Path:  "*balance.TestBalance.Value",
						Value: "10",
					}},
				},
				{
					ID:   "SET_BALANCE_TEST_DATA",
					TTL:  "0s",
					Type: "*set_balance",
					Diktats: []*utils.TPAPDiktat{{
						Path:  "*balance.TestDataBalance.Type",
						Value: "*data",
					}},
				},
				{
					ID:   "TOPUP_TEST_DATA",
					TTL:  "0s",
					Type: "*add_balance",
					Diktats: []*utils.TPAPDiktat{{
						Path:  "*balance.TestDataBalance.Value",
						Value: "1024",
					}},
				},
				{
					ID:   "SET_BALANCE_TEST_VOICE",
					TTL:  "0s",
					Type: "*set_balance",
					Diktats: []*utils.TPAPDiktat{{
						Path:  "*balance.TestVoiceBalance.Type",
						Value: "*voice",
					}},
				},
				{
					ID:   "TOPUP_TEST_VOICE",
					TTL:  "0s",
					Type: "*add_balance",
					Diktats: []*utils.TPAPDiktat{{
						Path:  "*balance.TestVoiceBalance.Value",
						Value: "15m15s",
					}, {
						Path:  "*balance.TestVoiceBalance2.Value",
						Value: "15m15s",
					}},
				},
			},
		}

		if len(csvr.actionProfiles) != 1 {
			t.Fatalf("Failed to load ActionProfiles: %s", utils.ToJSON(csvr.actionProfiles))
		}
		actPrfKey := utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "ONE_TIME_ACT",
		}
		sort.Slice(expected.Actions, func(i, j int) bool {
			return false
		})
		sort.Slice(expected.Targets[0].TargetIDs, func(i, j int) bool {
			return expected.Targets[0].TargetIDs[i] < expected.Targets[0].TargetIDs[j]
		})
		sort.Slice(csvr.actionProfiles[actPrfKey].Targets[0].TargetIDs, func(i, j int) bool {
			return csvr.actionProfiles[actPrfKey].Targets[0].TargetIDs[i] < csvr.actionProfiles[actPrfKey].Targets[0].TargetIDs[j]
		})

		if !reflect.DeepEqual(csvr.actionProfiles[actPrfKey], expected) {
			t.Errorf("Expecting: %+v,\n received: %+v",
				utils.ToJSON(expected), utils.ToJSON(csvr.actionProfiles[actPrfKey]))
		}
	})

	t.Run("load DispatcherHosts", func(t *testing.T) {
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
	})

	t.Run("load Accounts", func(t *testing.T) {
		expected := &utils.TPAccount{
			TPid:    testTPID,
			Tenant:  "cgrates.org",
			ID:      "1001",
			Weights: ";20",
			Balances: map[string]*utils.TPAccountBalance{
				"MonetaryBalance": {
					ID:      "MonetaryBalance",
					Weights: ";10",
					Type:    utils.MetaMonetary,
					CostIncrement: []*utils.TPBalanceCostIncrement{
						{
							FilterIDs:    []string{"fltr1", "fltr2"},
							Increment:    "1.3",
							FixedFee:     utils.Float64Pointer(2.3),
							RecurrentFee: utils.Float64Pointer(3.3),
						},
					},
					AttributeIDs: []string{"attr1", "attr2"},
					UnitFactors: []*utils.TPBalanceUnitFactor{
						{
							FilterIDs: []string{"fltr1", "fltr2"},
							Factor:    100,
						},
						{
							FilterIDs: []string{"fltr3"},
							Factor:    200,
						},
					},
					Units: "14",
				},
				"VoiceBalance": {
					ID:       "VoiceBalance",
					Weights:  ";10",
					Blockers: "*string:~*req.Destination:1002;true;;false",
					Type:     utils.MetaVoice,
					Units:    "1h",
				},
			},
			ThresholdIDs: []string{utils.MetaNone},
		}

		if len(csvr.accounts) != 1 {
			t.Fatalf("Failed to load Accounts: %s", utils.ToJSON(csvr.actionProfiles))
		}
		accPrfKey := utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "1001",
		}
		sort.Strings(csvr.accounts[accPrfKey].Balances["MonetaryBalance"].AttributeIDs)
		if !reflect.DeepEqual(csvr.accounts[accPrfKey], expected) {
			t.Errorf("Expecting: %+v,\n received: %+v",
				utils.ToJSON(expected), utils.ToJSON(csvr.accounts[accPrfKey]))
		}
	})
}
