//go:build integration
// +build integration

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

package v1

import (
	"net/rpc"
	"os"
	"path"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	tFIdxHRpc *rpc.Client

	sTestsFilterIndexesSHealth = []func(t *testing.T){
		testV1FIdxHLoadConfig,
		testV1FIdxHdxInitDataDb,
		testV1FIdxHResetStorDb,
		testV1FIdxHStartEngine,
		testV1FIdxHRpcConn,

		testV1FIdxHLoadTPs,
		testV1FIdxHAccountActionPlansHealth,
		testV1FIdxHReverseDestinationHealth,
		testV1FIdxHdxInitDataDb,
		testV1FIdxHResetStorDb,

		testV1FIdxHLoadFromFolderTutorial,
		testV1FIdxGetThresholdsIndexesHealth,
		testV1FIdxGetResourcesIndexesHealth,
		testV1FIdxGetStatsIndexesHealth,
		testV1FIdxGetSupplierIndexesHealth,
		testV1FIdxGetChargersIndexesHealth,
		testV1FIdxGetAttributesIndexesHealth,
		testV1FIdxCacheClear,

		testV1FIdxHdxInitDataDb,
		testV1FIdxHResetStorDb,
		testV1FIdxHLoadFromFolderDispatchers,
		testV1FIdxHGetDispatchersIndexesHealth,

		testV1FIdxHStopEngine,
	}
)

// Test start here
func TestFIdxHealthIT(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		tSv1ConfDIR = "tutinternal"
	case utils.MetaMySQL:
		tSv1ConfDIR = "tutmysql"
	case utils.MetaMongo:
		tSv1ConfDIR = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsFilterIndexesSHealth {
		t.Run(tSv1ConfDIR, stest)
	}
}

func testV1FIdxHLoadConfig(t *testing.T) {
	tSv1CfgPath = path.Join(*utils.DataDir, "conf", "samples", tSv1ConfDIR)
	var err error
	if tSv1Cfg, err = config.NewCGRConfigFromPath(tSv1CfgPath); err != nil {
		t.Error(err)
	}
}

func testV1FIdxHdxInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(tSv1Cfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func testV1FIdxHResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(tSv1Cfg); err != nil {
		t.Fatal(err)
	}
}

func testV1FIdxHStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(tSv1CfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

func testV1FIdxHRpcConn(t *testing.T) {
	var err error
	tFIdxHRpc, err = newRPCClient(tSv1Cfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

func testV1FIdxHLoadTPs(t *testing.T) {
	if err := os.RemoveAll("/tmp/TestFIdxHealthIT"); err != nil {
		t.Error(err)
	}
	if err := os.MkdirAll("/tmp/TestFIdxHealthIT", 0755); err != nil {
		t.Error(err)
	}
	defer os.RemoveAll("/tmp/TestFIdxHealthIT")
	writeFile := func(fileName, data string) error {
		csvFile, err := os.Create(path.Join("/tmp/TestFIdxHealthIT", fileName))
		if err != nil {
			return err
		}
		defer csvFile.Close()
		_, err = csvFile.WriteString(data)
		if err != nil {
			return err

		}
		return csvFile.Sync()
	}

	// Create and populate AccountActions.csv
	if err := writeFile(utils.AccountActionsCsv, `
#Tenant,Account,ActionPlanID,ActionTriggersID,AllowNegative,Disabled
cgrates.org,1001,STANDARD_PLAN,STANDARD_TRIGGERS,,
cgrates.org,1002,STANDARD_PLAN,STANDARD_TRIGGERS,,
cgrates.org,1003,STANDARD_PLAN,STANDARD_TRIGGERS,,
`); err != nil {
		t.Fatal(err)
	}

	// Create and populate ActionPlans.csv
	if err := writeFile(utils.ActionPlansCsv, `
#Id,ActionsId,TimingId,Weight
STANDARD_PLAN,TOPUP_RST_MONETARY_10,*asap,10
STANDARD_PLAN,TOPUP_RST_5M_VOICE,*asap,10
STANDARD_PLAN,TOPUP_RST_10M_VOICE,*asap,10
STANDARD_PLAN,TOPUP_RST_100_SMS,*asap,10
STANDARD_PLAN,TOPUP_RST_1024_DATA,*asap,10
STANDARD_PLAN,TOPUP_RST_1024_DATA,TM_NOON,10
`); err != nil {
		t.Fatal(err)
	}

	// Create and populate ActionTriggers.csv
	if err := writeFile(utils.ActionTriggersCsv, `
#ID[0],UniqueID[1],ThresholdType[2],ThresholdValue[3],Recurrent[4],MinSleep[5],ExpiryTime[6],ActivationTime[7],BalanceTag[8],BalanceType[9],BalanceCategories[10],BalanceDestinationIDs[11],BalanceRatingSubject[12],BalanceSharedGroup[13],BalanceExpiryTime[14],BalanceTimingIDs[15],BalanceWeight[16],BalanceBlocker[17],BalanceDisabled[18],ActionsID[19],Weight[20]
STANDARD_TRIGGERS,,*min_balance,2,,,,,,*monetary,,,,,,,,,,TOPUP_BONUS_10SMS,10
STANDARD_TRIGGERS,,*max_balance,100,,,,,,*monetary,,,,,,,,,,DISABLE_ACCOUNT,10
`); err != nil {
		t.Fatal(err)
	}

	// Create and populate Actions.csv
	if err := writeFile(utils.ActionsCsv, `
#ActionsId[0],Action[1],ExtraParameters[2],Filter[3],BalanceId[4],BalanceType[5],Categories[6],DestinationIds[7],RatingSubject[8],SharedGroup[9],ExpiryTime[10],TimingIds[11],Units[12],BalanceWeight[13],BalanceBlocker[14],BalanceDisabled[15],Weight[16]
TOPUP_RST_MONETARY_10,*topup_reset,,,*default,*monetary,,,,,,,10,10,,,10
TOPUP_RST_10M_VOICE,*topup_reset,,,PER_CALL,*voice,,DST_10,RPF_SPECIAL_BLC,,,,10m,10,,,10
TOPUP_RST_5M_VOICE,*topup_reset,,,FREE_MINS,*voice,,,*zero1m,,,,5m,20,,,10
TOPUP_RST_100_SMS,*topup_reset,,,FREE_SMSes,*sms,,,,,,,100,20,,,10
TOPUP_BONUS_10SMS,*topup,,,BONUS_SMSes,*sms,,DST_50,,,,,10,30,,,10
TOPUP_RST_1024_DATA,*topup_reset,,,FREE_DATA,*data,,,,,,,1024,20,,,10
LOG_WARNING,*log,,,,,,,,,,,,,,,10
DISABLE_ACCOUNT,*disable_account,,,,,,,,,,,,,,,10
WARN_HTTP_ASYNC,*http_post_async,http://$path/$to/$warn,,,,,,,,,,,,false,false,10
`); err != nil {
		t.Fatal(err)
	}

	// Create and populate Attributes.csv
	if err := writeFile(utils.AttributesCsv, `
#Tenant,ID,Contexts,FilterIDs,ActivationInterval,AttributeFilterIDs,Path,Type,Value,Blocker,Weight
cgrates.org,ATTR_CRG_SUPPLIER1,*sessions;*cdrs,,,,*req.Category,*constant,reseller1,false,0
cgrates.org,ATTR_CRG_SUPPLIER1,,,,,*req.RequestType,*constant,*rated,,
cgrates.org,ATTR_1001_AUTH,auth,*string:~*req.Account:1001,,,*req.Password,*constant,CGRateS.org,false,20
cgrates.org,ATTR_1002_AUTH,auth,*string:~*req.Account:1002,,,*req.Password,*constant,CGRateS.org,false,20
cgrates.org,ATTR_1003_AUTH,auth,*string:~*req.Account:1003,,,*req.Password,*constant,CGRateS.org,false,20
`); err != nil {
		t.Fatal(err)
	}

	// Create and populate Chargers.csv
	if err := writeFile(utils.ChargersCsv, `
#Tenant,ID,FilterIDs,ActivationInterval,RunID,AttributeIDs,Weight
cgrates.org,CGR_DEFAULT,,,*default,*none,0
cgrates.org,CRG_RESELLER1,,,reseller1,ATTR_CRG_RESELLER1,1
`); err != nil {
		t.Fatal(err)
	}

	// Create and populate DestinationRates.csv
	if err := writeFile(utils.DestinationRatesCsv, `
#ID,DestinationsID,RatesID,RoundingMethod,RoundingDecimals,MaxCost,MaxCostStrategy
DR_10_120C,DST_10,RT_120C,*up,4,,
DR_10_60C,DST_10,RT_60C,*up,4,,
DR_2030_120C,DST_2030,RT_120C,*up,4,,
DR_20_60C,DST_20,RT_60C,*up,4,,
DR_VOICEMAIL_FREE,DST_VOICEMAIL,RT_0,*up,4,,
DR_1002_60C,DST_1002,RT_60C,*up,4,,
DR_ANY_10C_CN,*any,RT_10C_CN,*up,4,,
DR_ANY_1024_1,*any,RT_1024_1,*up,4,,
DR_1002_10C1,DST_1002,RT_10C1,*up,4,,
DR_10_20C1,DST_10,RT_20C1,*up,4,,
DR_1CNT,*any,RT_1CNT,*up,4,,
DR_10CNT,*any,RT_10CNT,*up,4,,
`); err != nil {
		t.Fatal(err)
	}

	// Create and populate Destinations.csv
	if err := writeFile(utils.DestinationsCsv, `
#ID,Prefix
DST_10,10
DST_20,20
DST_2030,20
DST_2030,30
DST_VOICEMAIL,voicemail
DST_1002,1002
DST_50,50
`); err != nil {
		t.Fatal(err)
	}

	// Create and populate Rates.csv
	if err := writeFile(utils.RatesCsv, `
#ID,ConnectFee,Rate,RateUnit,RateIncrement,GroupIntervalStart
RT_120C,0.2,1.2,1m,1m,0
RT_120C,,1.2,1m,1s,1m
RT_60C,0.1,0.01,1s,1s,0
RT_0,0,0,1m,1m,0
RT_10C_CN,0.1,0,0,1s,0
RT_1024_1,0,1,1024,1024,0
RT_10C1,0,0.1,1,1,0
RT_20C1,0,0.2,1,1,0
RT_1CNT,0,0.01,1s,1s,0
RT_10CNT,0.1,0,1s,1s,0
`); err != nil {
		t.Fatal(err)
	}

	// Create and populate RatingPlans.csv
	if err := writeFile(utils.RatingPlansCsv, `
#ID,DestinationRatesID,TimingID,Weight
RP_STANDARD,DR_10_120C,PEAK,10
RP_STANDARD,DR_10_60C,OFFPEAK_MORNING,10
RP_STANDARD,DR_10_60C,OFFPEAK_EVENING,10
RP_STANDARD,DR_10_60C,OFFPEAK_WEEKEND,10
RP_STANDARD,DR_2030_120C,*any,10
RP_STANDARD,DR_20_60C,NEW_YEAR,20
RP_STANDARD,DR_VOICEMAIL_FREE,*any,10
RP_1001,DR_1002_60C,*any,10
RP_SPECIAL_BLC,DR_ANY_10C_CN,*any,10
RP_DATA,DR_ANY_1024_1,*any,10
RP_SMS,DR_1002_10C1,*any,10
RP_SMS,DR_10_20C1,*any,10
RP_1CNT,DR_1CNT,*any,0
RP_10CNT,DR_10CNT,*any,0
`); err != nil {
		t.Fatal(err)
	}

	// Create and populate RatingProfiles.csv
	if err := writeFile(utils.RatingProfilesCsv, `
#Tenant,Category,Subject,ActivationTime,RatingPlanID,FallbackSubject
cgrates.org,call,*any,2019-03-01T00:00:00Z,RP_STANDARD,
cgrates.org,call,1001,2019-03-01T00:00:00Z,RP_1001,*any
cgrates.org,call,RPF_SPECIAL_BLC,2019-03-01T00:00:00Z,RP_SPECIAL_BLC,
cgrates.org,data,*any,2019-03-01T00:00:00Z,RP_DATA,
cgrates.org,sms,*any,2019-03-01T00:00:00Z,RP_SMS,
cgrates.org,reseller1,*any,2019-03-01T00:00:00Z,RP_10CNT,
`); err != nil {
		t.Fatal(err)
	}

	// Create and populate Timings.csv
	if err := writeFile(utils.TimingsCsv, `
#ID,Years,Months,MonthDays,WeekDays,Time
PEAK,*any,*any,*any,1;2;3;4;5,08:00:00
OFFPEAK_MORNING,*any,*any,*any,1;2;3;4;5,00:00:00
OFFPEAK_EVENING,*any,*any,*any,1;2;3;4;5,19:00:00
OFFPEAK_WEEKEND,*any,*any,*any,6;7,00:00:00
NEW_YEAR,*any,1,1,*any,00:00:00
TM_NOON,*any,*any,*any,*any,12:00:00
`); err != nil {
		t.Fatal(err)
	}

	var loadInst string
	if err := tFIdxHRpc.Call(utils.APIerSv1LoadTariffPlanFromFolder,
		&utils.AttrLoadTpFromFolder{FolderPath: "/tmp/TestFIdxHealthIT"}, &loadInst); err != nil {
		t.Error(err)
	}
}

func testV1FIdxHAccountActionPlansHealth(t *testing.T) {
	var reply engine.AccountActionPlanIHReply
	if err := tFIdxHRpc.Call(utils.APIerSv1GetAccountActionPlansIndexHealth, engine.IndexHealthArgsWith2Ch{
		IndexCacheLimit:  -1,
		ObjectCacheLimit: -1,
	}, &reply); err != nil {
		t.Error(err)
	}
	exp := engine.AccountActionPlanIHReply{
		MissingAccountActionPlans: map[string][]string{},
		BrokenReferences:          map[string][]string{},
	}
	if !reflect.DeepEqual(exp, reply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(exp), utils.ToJSON(reply))
	}
}

func testV1FIdxHReverseDestinationHealth(t *testing.T) {
	var reply engine.ReverseDestinationsIHReply
	if err := tFIdxHRpc.Call(utils.APIerSv1GetReverseDestinationsIndexHealth, engine.IndexHealthArgsWith2Ch{
		IndexCacheLimit:  -1,
		ObjectCacheLimit: -1,
	}, &reply); err != nil {
		t.Error(err)
	}
	exp := engine.ReverseDestinationsIHReply{
		MissingReverseDestinations: map[string][]string{},
		BrokenReferences:           map[string][]string{},
	}
	if !reflect.DeepEqual(exp, reply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(exp), utils.ToJSON(reply))
	}
}

func testV1FIdxCacheClear(t *testing.T) {
	var reply string
	if err := tFIdxHRpc.Call(utils.CacheSv1Clear,
		&utils.AttrCacheIDsWithArgDispatcher{}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Calling CacheSv1.ReloadCache got reply: ", reply)
	}
}

func testV1FIdxHLoadFromFolderTutorial(t *testing.T) {
	var reply string
	if err := tFIdxHRpc.Call(utils.CacheSv1Clear, &utils.AttrCacheIDsWithArgDispatcher{
		CacheIDs: nil,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Reply: ", reply)
	}
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*utils.DataDir, "tariffplans", "tutorial")}
	if err := tFIdxHRpc.Call(utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(100 * time.Millisecond)
}

func testV1FIdxGetThresholdsIndexesHealth(t *testing.T) {
	// set another threshold profile different than the one from tariffplan
	tPrfl = &engine.ThresholdWithCache{
		ThresholdProfile: &engine.ThresholdProfile{
			Tenant: tenant,
			ID:     "TEST_PROFILE1",
			FilterIDs: []string{"*string:~*req.Account:1004",
				"*prefix:~*opts.Destination:+442",
				"*prefix:~*opts.Destination:+554"},
			MaxHits:   1,
			MinSleep:  5 * time.Minute,
			Blocker:   false,
			Weight:    20.0,
			ActionIDs: []string{"ACT_1", "ACT_2"},
			Async:     true,
		},
	}

	var rplyok string
	if err := tFIdxHRpc.Call(utils.APIerSv1SetThresholdProfile, tPrfl, &rplyok); err != nil {
		t.Error(err)
	} else if rplyok != utils.OK {
		t.Error("Unexpected reply returned", rplyok)
	}

	// check all the indexes for thresholds
	expiIdx := []string{
		"*string:~*req.Account:1002:THD_ACNT_1002",
		"*string:~*req.Account:1001:THD_ACNT_1001",
		"*string:~*req.Account:1004:TEST_PROFILE1",
		"*prefix:~*opts.Destination:+442:TEST_PROFILE1",
		"*prefix:~*opts.Destination:+554:TEST_PROFILE1",
	}
	var result []string
	if err := tFIdxHRpc.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		ItemType: utils.MetaThresholds,
		Tenant:   "cgrates.org",
	}, &result); err != nil {
		t.Error(err)
	} else {
		sort.Strings(result)
		sort.Strings(expiIdx)
		if !reflect.DeepEqual(expiIdx, result) {
			t.Errorf("Expecting: %+v, received: %+v", expiIdx, result)
		}
	}

	// all indexes are set and points to their objects correctly
	args := &engine.IndexHealthArgsWith3Ch{}
	expRPly := &engine.FilterIHReply{
		MissingObjects: []string{},
		MissingIndexes: map[string][]string{},
		BrokenIndexes:  map[string][]string{},
		MissingFilters: map[string][]string{},
	}
	rply := &engine.FilterIHReply{
		MissingObjects: []string{},
	}
	if err := tFIdxHRpc.Call(utils.APIerSv1GetThresholdsIndexesHealth,
		args, &rply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rply, expRPly) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expRPly), utils.ToJSON(rply))
	}

	// removing a profile + their indexes
	if err := tFIdxHRpc.Call(utils.APIerSv1RemoveThresholdProfile,
		&utils.TenantIDWithCache{
			Tenant: "cgrates.org",
			ID:     "THD_ACNT_1002",
		}, &rplyok); err != nil {
		t.Error(err)
	} else if rplyok != utils.OK {
		t.Errorf("UNexpected reply returned")
	}

	// check all the indexes for thresholds
	expiIdx = []string{
		"*string:~*req.Account:1001:THD_ACNT_1001",
		"*string:~*req.Account:1004:TEST_PROFILE1",
		"*prefix:~*opts.Destination:+442:TEST_PROFILE1",
		"*prefix:~*opts.Destination:+554:TEST_PROFILE1",
	}
	if err := tFIdxHRpc.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		ItemType: utils.MetaThresholds,
		Tenant:   "cgrates.org",
	}, &result); err != nil {
		t.Error(err)
	} else {
		sort.Strings(result)
		sort.Strings(expiIdx)
		if !reflect.DeepEqual(expiIdx, result) {
			t.Errorf("Expecting: %+v, received: %+v", expiIdx, result)
		}
	}
	//as we removed the object, the index specified is removed too, so the health of the indexes is fine
	expRPly = &engine.FilterIHReply{
		MissingObjects: []string{},
		MissingIndexes: map[string][]string{},
		BrokenIndexes:  map[string][]string{},
		MissingFilters: map[string][]string{},
	}
	if err := tFIdxHRpc.Call(utils.APIerSv1GetThresholdsIndexesHealth,
		args, &rply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rply, expRPly) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expRPly), utils.ToJSON(rply))
	}
}

func testV1FIdxGetResourcesIndexesHealth(t *testing.T) {
	// set another resource profile different than the one from tariffplan
	var reply string
	rlsPrf := &ResourceWithCache{
		ResourceProfile: &engine.ResourceProfile{
			Tenant: "cgrates.org",
			ID:     "ResGroup2",
			FilterIDs: []string{"*string:~*req.Account:1001",
				"*prefix:~*opts.Destination:+334;+122"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 29, 15, 0, 0, 0, time.UTC),
			},
			UsageTTL:          -1,
			Limit:             7,
			AllocationMessage: "",
			Stored:            true,
			Weight:            10,
			ThresholdIDs:      []string{utils.META_NONE},
		},
	}
	if err := tFIdxHRpc.Call(utils.APIerSv1SetResourceProfile, rlsPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}

	// check all the indexes for resources
	expIdx := []string{
		"*string:~*req.Account:1001:ResGroup2",
		"*prefix:~*opts.Destination:+334:ResGroup2",
		"*prefix:~*opts.Destination:+122:ResGroup2",
		"*string:~*req.Account:1001:ResGroup1",
		"*string:~*req.Account:1002:ResGroup1",
		"*string:~*req.Account:1003:ResGroup1",
	}
	var result []string
	if err := tFIdxHRpc.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		ItemType: utils.MetaResources,
		Tenant:   "cgrates.org",
	}, &result); err != nil {
		t.Error(err)
	} else {
		sort.Strings(result)
		sort.Strings(expIdx)
		if !reflect.DeepEqual(expIdx, result) {
			t.Errorf("Expecting: %+v, received: %+v", expIdx, result)
		}
	}

	// all indexes are set and points to their objects correctly
	expRPly := &engine.FilterIHReply{
		MissingObjects: []string{},
		MissingIndexes: map[string][]string{},
		BrokenIndexes:  map[string][]string{},
		MissingFilters: map[string][]string{},
	}
	args := &engine.IndexHealthArgsWith3Ch{}
	rply := &engine.FilterIHReply{
		MissingObjects: []string{},
	}
	if err := tFIdxHRpc.Call(utils.APIerSv1GetResourcesIndexesHealth,
		args, &rply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rply, expRPly) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expRPly), utils.ToJSON(rply))
	}

	// removing a profile + their indexes
	if err := tFIdxHRpc.Call(utils.APIerSv1RemoveResourceProfile,
		utils.TenantIDWithCache{
			Tenant: "cgrates.org",
			ID:     "ResGroup2",
		}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("UNexpected reply returned")
	}

	//as we removed the object, the index specified is removed too, so the health of the indexes is fine
	if err := tFIdxHRpc.Call(utils.APIerSv1GetResourcesIndexesHealth,
		args, &rply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rply, expRPly) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expRPly), utils.ToJSON(rply))
	}
}

func testV1FIdxGetStatsIndexesHealth(t *testing.T) {
	// set another stats profile different than the one from tariffplan
	statConfig = &engine.StatQueueWithCache{
		StatQueueProfile: &engine.StatQueueProfile{
			Tenant: "cgrates.org",
			ID:     "TEST_STATPROFILE_1",
			FilterIDs: []string{"*string:~*req.OriginID:RandomID",
				"*prefix:~*opts.Destination:+332;+234"},
			QueueLength: 10,
			TTL:         10 * time.Second,
			Metrics: []*engine.MetricWithFilters{
				{
					MetricID: utils.MetaTCD,
				},
			},
			ThresholdIDs: []string{"*none"},
			Blocker:      true,
			Stored:       true,
			Weight:       20,
			MinItems:     1,
		},
	}
	var rply string
	if err := tFIdxHRpc.Call(utils.APIerSv1SetStatQueueProfile, statConfig, &rply); err != nil {
		t.Error(err)
	} else if rply != utils.OK {
		t.Error("Unexpected reply returned", rply)
	}

	// check all the indexes for statsQueue
	expIdx := []string{
		"*string:~*req.OriginID:RandomID:TEST_STATPROFILE_1",
		"*prefix:~*opts.Destination:+332:TEST_STATPROFILE_1",
		"*prefix:~*opts.Destination:+234:TEST_STATPROFILE_1",
		"*string:~*req.Account:1001:Stats2",
		"*string:~*req.Account:1002:Stats2",
		"*string:~*req.RunID:*default:Stats2",
		"*string:~*req.Destination:1001:Stats2",
		"*string:~*req.Destination:1002:Stats2",
		"*string:~*req.Destination:1003:Stats2",
		"*string:~*req.Account:1003:Stats2_1",
		"*string:~*req.RunID:*default:Stats2_1",
		"*string:~*req.Destination:1001:Stats2_1",
	}
	var result []string
	if err := tFIdxHRpc.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		ItemType: utils.MetaStats,
		Tenant:   "cgrates.org",
	}, &result); err != nil {
		t.Error(err)
	} else {
		sort.Strings(result)
		sort.Strings(expIdx)
		if !reflect.DeepEqual(expIdx, result) {
			t.Errorf("Expecting: %+v, received: %+v", expIdx, result)
		}
	}

	// all indexes are set and points to their objects correctly
	expRPly := &engine.FilterIHReply{
		MissingObjects: []string{},
		MissingIndexes: map[string][]string{},
		BrokenIndexes:  map[string][]string{},
		MissingFilters: map[string][]string{},
	}
	args := &engine.IndexHealthArgsWith3Ch{}
	rplyFl := &engine.FilterIHReply{
		MissingObjects: []string{},
	}
	if err := tFIdxHRpc.Call(utils.APIerSv1GetStatsIndexesHealth,
		args, &rplyFl); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rplyFl, expRPly) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expRPly), utils.ToJSON(rplyFl))
	}

	// removing a profile + their indexes
	if err := tFIdxHRpc.Call(utils.APIerSv1RemoveStatQueueProfile,
		&utils.TenantIDWithCache{
			Tenant: "cgrates.org",
			ID:     "Stats2",
		}, &rply); err != nil {
		t.Error(err)
	} else if rply != utils.OK {
		t.Errorf("UNexpected reply returned")
	}

	//as we removed the object, the index specified is removed too, so the health of the indexes is fine
	if err := tFIdxHRpc.Call(utils.APIerSv1GetStatsIndexesHealth,
		args, &rplyFl); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rplyFl, expRPly) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expRPly), utils.ToJSON(rplyFl))
	}
}

func testV1FIdxGetSupplierIndexesHealth(t *testing.T) {
	// set another routes profile different than the one from tariffplan
	rPrf := &SupplierWithCache{
		SupplierProfile: &engine.SupplierProfile{
			Tenant:            tenant,
			ID:                "TEST_PROFILE1",
			FilterIDs:         []string{"*prefix:~*req.Destination:+23331576354"},
			Sorting:           "Sort1",
			SortingParameters: []string{"Param1", "Param2"},
			Suppliers: []*engine.Supplier{{
				ID:            "SPL1",
				RatingPlanIDs: []string{"RP1"},
				FilterIDs:     []string{"FLTR_1"},
				Weight:        20,
				Blocker:       false,
			}},
			Weight: 10,
		},
	}
	var reply string
	if err := tFIdxHRpc.Call(utils.APIerSv1SetSupplierProfile, rPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}

	// check all the indexes for routes
	expIdx := []string{
		"*prefix:~*req.Destination:+23331576354:TEST_PROFILE1",
		"*string:~*req.Account:1001:SPL_ACNT_1001",
		"*string:~*req.Account:1002:SPL_ACNT_1002",
		"*string:~*req.Account:1003:SPL_ACNT_1003",
	}
	var result []string
	if err := tFIdxHRpc.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		ItemType: utils.MetaSuppliers,
		Tenant:   "cgrates.org",
	}, &result); err != nil {
		t.Error(err)
	} else {
		sort.Strings(result)
		sort.Strings(expIdx)
		if !reflect.DeepEqual(expIdx, result) {
			t.Errorf("Expecting: %+v, received: %+v", expIdx, result)
		}
	}

	// all indexes are set and points to their objects correctly
	expRPly := &engine.FilterIHReply{
		MissingObjects: []string{},
		MissingIndexes: map[string][]string{},
		BrokenIndexes:  map[string][]string{},
		MissingFilters: map[string][]string{},
	}
	args := &engine.IndexHealthArgsWith3Ch{}
	rplyFl := &engine.FilterIHReply{
		MissingObjects: []string{},
	}
	if err := tFIdxHRpc.Call(utils.APIerSv1GetSuppliersIndexesHealth,
		args, &rplyFl); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rplyFl, expRPly) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expRPly), utils.ToJSON(rplyFl))
	}

	// removing a profile + their indexes
	if err := tFIdxHRpc.Call(utils.APIerSv1RemoveSupplierProfile,
		&utils.TenantIDWithCache{
			Tenant: "cgrates.org",
			ID:     "TEST_PROFILE1",
		}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("UNexpected reply returned")
	}

	//as we removed the object, the index specified is removed too, so the health of the indexes is fine
	if err := tFIdxHRpc.Call(utils.APIerSv1GetSuppliersIndexesHealth,
		args, &rplyFl); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rplyFl, expRPly) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expRPly), utils.ToJSON(rplyFl))
	}
}

func testV1FIdxGetChargersIndexesHealth(t *testing.T) {
	// set another charger profile different than the one from tariffplan
	chargerProfile := &ChargerWithCache{
		ChargerProfile: &engine.ChargerProfile{
			Tenant: "cgrates.org",
			ID:     "Default",
			FilterIDs: []string{"*string:~*req.Destination:+1442",
				"*prefix:~*opts.Accounts:1002;1004"},
			RunID:        utils.MetaDefault,
			AttributeIDs: []string{"*none"},
			Weight:       20,
		},
	}
	var reply string
	if err := tFIdxHRpc.Call(utils.APIerSv1SetChargerProfile, chargerProfile, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}

	// those 2 charger object (*none:*any:*any index) are from tutorial2 tariffplan, so on imternal we must delete them by api
	if tSv1Cfg.DataDbCfg().DataDbType == utils.MetaInternal {
		var result string
		if err := tFIdxHRpc.Call(utils.APIerSv1RemoveChargerProfile,
			&utils.TenantIDWithCache{
				Tenant: "cgrates.org",
				ID:     "CRG_RESELLER1",
			}, &result); err != nil {
			t.Error(err)
		} else if result != utils.OK {
			t.Errorf("Unexpected reply returned")
		}
		if err := tFIdxHRpc.Call(utils.APIerSv1RemoveChargerProfile,
			&utils.TenantIDWithCache{
				Tenant: "cgrates.org",
				ID:     "CGR_DEFAULT",
			}, &result); err != nil {
			t.Error(err)
		} else if result != utils.OK {
			t.Errorf("Unexpected reply returned")
		}
	}

	// check all the indexes for chargers
	expIdx := []string{
		"*string:~*req.Destination:+1442:Default",
		"*prefix:~*opts.Accounts:1002:Default",
		"*prefix:~*opts.Accounts:1004:Default",
		"*none:*any:*any:DEFAULT",
		"*none:*any:*any:Raw",
	}
	var result []string
	if err := tFIdxHRpc.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		ItemType: utils.MetaChargers,
		Tenant:   "cgrates.org",
	}, &result); err != nil {
		t.Error(err)
	} else {
		sort.Strings(result)
		sort.Strings(expIdx)
		if !reflect.DeepEqual(expIdx, result) {
			t.Errorf("Expecting: %+v, received: %+v", expIdx, result)
		}
	}

	// all indexes are set and points to their objects correctly
	expRPly := &engine.FilterIHReply{
		MissingObjects: []string{},
		MissingIndexes: map[string][]string{},
		BrokenIndexes:  map[string][]string{},
		MissingFilters: map[string][]string{},
	}
	args := &engine.IndexHealthArgsWith3Ch{}
	rplyFl := &engine.FilterIHReply{
		MissingObjects: []string{},
	}
	if err := tFIdxHRpc.Call(utils.APIerSv1GetChargersIndexesHealth,
		args, &rplyFl); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rplyFl, expRPly) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expRPly), utils.ToJSON(rplyFl))
	}

	// removing a profile + their indexes
	if err := tFIdxHRpc.Call(utils.APIerSv1RemoveChargerProfile,
		&utils.TenantIDWithCache{
			Tenant: "cgrates.org",
			ID:     "Raw",
		}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Unexpected reply returned")
	}

	//as we removed the object, the index specified is removed too, so the health of the indexes is fine
	if err := tFIdxHRpc.Call(utils.APIerSv1GetChargersIndexesHealth,
		args, &rplyFl); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rplyFl, expRPly) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expRPly), utils.ToJSON(rplyFl))
	}
}

func testV1FIdxGetAttributesIndexesHealth(t *testing.T) {
	// Attributes.csv from tutorial tariffplan got lots of profiles, so we will not set another attribute for this test
	// check all the indexes for attributes
	// simpleauth context
	expIdx := []string{
		"*string:~*req.Account:1001:ATTR_1001_SIMPLEAUTH",
		"*string:~*req.Account:1002:ATTR_1002_SIMPLEAUTH",
		"*string:~*req.Account:1003:ATTR_1003_SIMPLEAUTH",
	}
	var result []string
	if err := tFIdxHRpc.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		ItemType: utils.MetaAttributes,
		Context:  "simpleauth",
		Tenant:   "cgrates.org",
	}, &result); err != nil {
		t.Error(err)
	} else {
		sort.Strings(result)
		sort.Strings(expIdx)
		if !reflect.DeepEqual(expIdx, result) {
			t.Errorf("Expecting: %+v, received: %+v", expIdx, result)
		}
	}

	// this attr object (*none:*any:*any index) must be deleted with api
	if tSv1Cfg.DataDbCfg().DataDbType == utils.MetaInternal {
		var result string
		if err := tFIdxHRpc.Call(utils.APIerSv1RemoveAttributeProfile,
			&utils.TenantIDWithCache{
				Tenant: "cgrates.org",
				ID:     "ATTR_CRG_SUPPLIER1",
			}, &result); err != nil {
			t.Error(err)
		} else if result != utils.OK {
			t.Errorf("Unexpected reply returned")
		}
	}

	// *sessions context
	expIdx = []string{
		"*string:~*req.Account:1001:ATTR_1001_SESSIONAUTH",
		"*string:~*req.Account:1002:ATTR_1002_SESSIONAUTH",
		"*string:~*req.Account:1003:ATTR_1003_SESSIONAUTH",
	}
	if err := tFIdxHRpc.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		ItemType: utils.MetaAttributes,
		Context:  utils.MetaSessionS,
		Tenant:   "cgrates.org",
	}, &result); err != nil {
		t.Error(err)
	} else {
		sort.Strings(result)
		sort.Strings(expIdx)
		if !reflect.DeepEqual(expIdx, result) {
			t.Errorf("Expecting: %+v, received: %+v", expIdx, result)
		}
	}

	// *any context tenant: cgrates.org
	expIdx = []string{
		"*string:~*req.SubscriberId:1006:ATTR_ACC_ALIAS",
	}
	if err := tFIdxHRpc.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		ItemType: utils.MetaAttributes,
		Context:  utils.META_ANY,
		Tenant:   "cgrates.org",
	}, &result); err != nil {
		t.Error(err)
	} else {
		sort.Strings(result)
		sort.Strings(expIdx)
		if !reflect.DeepEqual(expIdx, result) {
			t.Errorf("Expecting: %+v, received: %+v", expIdx, result)
		}
	}

	// *any context tenant: cgrates.com
	expIdx = []string{
		"*string:~*req.SubscriberId:1006:ATTR_TNT_ALIAS",
		"*string:~*req.Account:1001:ATTR_TNT_1001",
		"*string:~*req.Account:testDiamInitWithSessionDisconnect:ATTR_TNT_DISC",
	}
	if err := tFIdxHRpc.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		Tenant:   "cgrates.com",
		ItemType: utils.MetaAttributes,
		Context:  utils.META_ANY,
	}, &result); err != nil {
		t.Error(err)
	} else {
		sort.Strings(result)
		sort.Strings(expIdx)
		if !reflect.DeepEqual(expIdx, result) {
			t.Errorf("Expecting: %+v, received: %+v", expIdx, result)
		}
	}

	//as we removed the object, the index specified is removed too, so the health of the indexes is fine
	expRPly := &engine.FilterIHReply{
		MissingObjects: []string{},
		MissingIndexes: map[string][]string{},
		BrokenIndexes:  map[string][]string{},
		MissingFilters: map[string][]string{},
	}
	args := &engine.IndexHealthArgsWith3Ch{}
	rplyFl := &engine.FilterIHReply{
		MissingObjects: []string{},
	}
	if err := tFIdxHRpc.Call(utils.APIerSv1GetAttributesIndexesHealth,
		args, &rplyFl); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rplyFl, expRPly) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expRPly), utils.ToJSON(rplyFl))
	}
}

func testV1FIdxHLoadFromFolderDispatchers(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*utils.DataDir, "tariffplans", "dispatchers")}
	if err := tFIdxHRpc.Call(utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(100 * time.Millisecond)
}

func testV1FIdxHGetDispatchersIndexesHealth(t *testing.T) {
	// *any context
	expIdx := []string{
		"*none:*any:*any:PING1",
		"*string:~*req.EventName:UnexistedHost:PING2",
		"*string:~*req.EventName:Event1:EVENT1",
		"*string:~*req.EventName:RoundRobin:EVENT2",
		"*string:~*req.EventName:Random:EVENT3",
		"*string:~*req.EventName:Broadcast:EVENT4",
		"*string:~*req.EventName:Internal:EVENT5",
		// "*string:~*opts.*method:DispatcherSv1.GetProfilesForEvent:EVENT6",
		// "*string:~*opts.EventType:LoadDispatcher:EVENT7",
	}
	var result []string
	if err := tFIdxHRpc.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		ItemType: utils.MetaDispatchers,
		Context:  utils.META_ANY,
		Tenant:   "cgrates.org",
	}, &result); err != nil {
		t.Error(err)
	} else {
		sort.Strings(result)
		sort.Strings(expIdx)
		if !reflect.DeepEqual(expIdx, result) {
			t.Errorf("Expecting: %+v, received: %+v", expIdx, result)
		}
	}

	// all indexes are set and points to their objects correctly
	expRPly := &engine.FilterIHReply{
		MissingObjects: []string{},
		MissingIndexes: map[string][]string{},
		BrokenIndexes:  map[string][]string{},
		MissingFilters: map[string][]string{},
	}
	args := &engine.IndexHealthArgsWith3Ch{}
	rplyFl := &engine.FilterIHReply{
		MissingObjects: []string{},
	}
	if err := tFIdxHRpc.Call(utils.APIerSv1GetDispatchersIndexesHealth,
		args, &rplyFl); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rplyFl, expRPly) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expRPly), utils.ToJSON(rplyFl))
	}

	var reply string
	// removing a profile + their indexes
	if err := tFIdxHRpc.Call(utils.APIerSv1RemoveDispatcherProfile,
		&utils.TenantIDWithCache{
			Tenant: "cgrates.org",
			ID:     "PING2",
		}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Unexpected reply returned")
	}

	//as we removed the object, the index specified is removed too, so the health of the indexes is fine
	args = &engine.IndexHealthArgsWith3Ch{}
	if err := tFIdxHRpc.Call(utils.APIerSv1GetDispatchersIndexesHealth,
		args, &rplyFl); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rplyFl, expRPly) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expRPly), utils.ToJSON(rplyFl))
	}
}

func testV1FIdxHStopEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
