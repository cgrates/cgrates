/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT MetaAny WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package engine

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/rpc/jsonrpc"
	"os"
	"os/exec"
	"path"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
	"github.com/creack/pty"
)

var (
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
FILTER,*topup,,"{""*and"":[{""Value"":{""*lt"":0}},{""Id"":{""*eq"":""*default""}}]}",bfree,*monetary,,,,,*unlimited,,20,10,false,false,10
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
#Tenant[0],ID[1],Address[2],Transport[3],TLS[4]
cgrates.org,ALL1,127.0.0.1:2012,*json,true
`
	RateProfileCSVContent = `
#Tenant,ID,FilterIDs,ActivationInterval,Weights,MinCost,MaxCost,MaxCostStrategy,RateID,RateFilterIDs,RateActivationStart,RateWeights,RateBlocker,RateIntervalStart,RateFixedFee,RateRecurrentFee,RateUnit,RateIncrement
cgrates.org,RP1,*string:~*req.Subject:1001,,;0,0.1,0.6,*free,RT_WEEK,,"* * * * 1-5",;0,false,0s,0,0.12,1m,1m
cgrates.org,RP1,,,,,,,RT_WEEK,,,,,1m,1.234,0.06,1m,1s
cgrates.org,RP1,,,,,,,RT_WEEKEND,,"* * * * 0,6",;10,false,0s,0.089,0.06,1m,1s
cgrates.org,RP1,,,,,,,RT_CHRISTMAS,,* * 24 12 *,;30,false,0s,0.0564,0.06,1m,1s
`
	ActionProfileCSVContent = `
#Tenant,ID,FilterIDs,ActivationInterval,Weight,Schedule,TargetType,TargetIDs,ActionID,ActionFilterIDs,ActionBlocker,ActionTTL,ActionType,ActionOpts,ActionPath,ActionValue
cgrates.org,ONE_TIME_ACT,,,10,*asap,*accounts,1001;1002,TOPUP,,false,0s,*add_balance,,*balance.TestBalance.Value,10
cgrates.org,ONE_TIME_ACT,,,,,,,SET_BALANCE_TEST_DATA,,false,0s,*set_balance,,*balance.TestDataBalance.Type,*data
cgrates.org,ONE_TIME_ACT,,,,,,,TOPUP_TEST_DATA,,false,0s,*add_balance,,*balance.TestDataBalance.Value,1024
cgrates.org,ONE_TIME_ACT,,,,,,,SET_BALANCE_TEST_VOICE,,false,0s,*set_balance,,*balance.TestVoiceBalance.Type,*voice
cgrates.org,ONE_TIME_ACT,,,,,,,TOPUP_TEST_VOICE,,false,0s,*add_balance,,*balance.TestVoiceBalance.Value,15m15s
cgrates.org,ONE_TIME_ACT,,,,,,,TOPUP_TEST_VOICE,,false,0s,*add_balance,,*balance.TestVoiceBalance2.Value,15m15s
`

	AccountProfileCSVContent = `
#Tenant,ID,FilterIDs,ActivationInterval,Weights,Opts,BalanceID,BalanceFilterIDs,BalanceWeights,BalanceType,BalanceUnits,BalanceUnitFactors,BalanceOpts,BalanceCostIncrements,BalanceAttributeIDs,BalanceRateProfileIDs,ThresholdIDs
cgrates.org,1001,,,;20,,MonetaryBalance,,;10,*monetary,14,fltr1&fltr2;100;fltr3;200,,fltr1&fltr2;1.3;2.3;3.3,attr1;attr2,,*none
cgrates.org,1001,,,,,VoiceBalance,,;10,*voice,3600000000000,,,,,,
`
)

func InitDataDb(cfg *config.CGRConfig) error {
	d, err := NewDataDBConn(cfg.DataDbCfg().DataDbType,
		cfg.DataDbCfg().DataDbHost, cfg.DataDbCfg().DataDbPort,
		cfg.DataDbCfg().DataDbName, cfg.DataDbCfg().DataDbUser,
		cfg.DataDbCfg().DataDbPass, cfg.GeneralCfg().DBDataEncoding,
		cfg.DataDbCfg().Opts)
	if err != nil {
		return err
	}
	dm := NewDataManager(d, cfg.CacheCfg(), connMgr)

	if err := dm.DataDB().Flush(""); err != nil {
		return err
	}
	//	Write version before starting
	if err := OverwriteDBVersions(dm.dataDB); err != nil {
		return err
	}
	return nil
}

func InitStorDb(cfg *config.CGRConfig) error {
	storDb, err := NewStorDBConn(cfg.StorDbCfg().Type,
		cfg.StorDbCfg().Host, cfg.StorDbCfg().Port,
		cfg.StorDbCfg().Name, cfg.StorDbCfg().User,
		cfg.StorDbCfg().Password, cfg.GeneralCfg().DBDataEncoding,
		cfg.StorDbCfg().StringIndexedFields, cfg.StorDbCfg().PrefixIndexedFields,
		cfg.StorDbCfg().Opts)
	if err != nil {
		return err
	}
	if err := storDb.Flush(path.Join(cfg.DataFolderPath, "storage",
		cfg.StorDbCfg().Type)); err != nil {
		return err
	}
	if utils.IsSliceMember([]string{utils.Mongo, utils.MySQL, utils.Postgres},
		cfg.StorDbCfg().Type) {
		if err := SetDBVersions(storDb); err != nil {
			return err
		}
	}
	return nil
}

// Return reference towards the command started so we can stop it if necessary
func StartEngine(cfgPath string, waitEngine int) (*exec.Cmd, error) {
	enginePath, err := exec.LookPath("cgr-engine")
	if err != nil {
		return nil, err
	}
	engine := exec.Command(enginePath, "-config_path", cfgPath)
	if err := engine.Start(); err != nil {
		return nil, err
	}
	cfg, err := config.NewCGRConfigFromPath(cfgPath)
	if err != nil {
		return nil, err
	}
	fib := utils.Fib()
	var connected bool
	for i := 0; i < 200; i++ {
		time.Sleep(time.Duration(fib()) * time.Millisecond)
		if _, err := jsonrpc.Dial(utils.TCP, cfg.ListenCfg().RPCJSONListen); err != nil {
			utils.Logger.Warning(fmt.Sprintf("Error <%s> when opening test connection to: <%s>",
				err.Error(), cfg.ListenCfg().RPCJSONListen))
		} else {
			connected = true
			break
		}
	}
	if !connected {
		return nil, fmt.Errorf("engine did not open port <%s>", cfg.ListenCfg().RPCJSONListen)
	}
	time.Sleep(time.Duration(waitEngine) * time.Millisecond) // wait for rater to register all subsystems
	return engine, nil
}

// StartEngineWithContext return reference towards the command started so we can stop it if necessary
func StartEngineWithContext(ctx context.Context, cfgPath string, waitEngine int) (engine *exec.Cmd, err error) {
	engine = exec.CommandContext(ctx, "cgr-engine", "-config_path", cfgPath)
	if err = engine.Start(); err != nil {
		return nil, err
	}
	var cfg *config.CGRConfig
	if cfg, err = config.NewCGRConfigFromPath(cfgPath); err != nil {
		return
	}
	fib := utils.Fib()
	for i := 0; i < 200; i++ {
		time.Sleep(time.Duration(fib()) * time.Millisecond)
		if _, err = jsonrpc.Dial(utils.TCP, cfg.ListenCfg().RPCJSONListen); err != nil {
			continue
		}
		time.Sleep(time.Duration(waitEngine) * time.Millisecond) // wait for rater to register all subsystems
		return
	}
	utils.Logger.Warning(fmt.Sprintf("Error <%s> when opening test connection to: <%s>",
		err.Error(), cfg.ListenCfg().RPCJSONListen))
	err = fmt.Errorf("engine did not open port <%s>", cfg.ListenCfg().RPCJSONListen)
	return
}

func KillEngine(waitEngine int) error {
	return KillProcName("cgr-engine", waitEngine)
}

func StopStartEngine(cfgPath string, waitEngine int) (*exec.Cmd, error) {
	KillEngine(waitEngine)
	return StartEngine(cfgPath, waitEngine)
}

func LoadTariffPlanFromFolder(tpPath, timezone string, dm *DataManager, disable_reverse bool,
	cacheConns, schedConns []string) error {
	loader, err := NewTpReader(dm.dataDB, NewFileCSVStorage(utils.CSVSep, tpPath), "",
		timezone, cacheConns, schedConns, false)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	if err := loader.LoadAll(); err != nil {
		return utils.NewErrServerError(err)
	}
	if err := loader.WriteToDatabase(false, disable_reverse); err != nil {
		return utils.NewErrServerError(err)
	}
	return nil
}

type PjsuaAccount struct {
	Id, Username, Password, Realm, Registrar string
}

// Returns file reference where we can write to control pjsua in terminal
func StartPjsuaListener(acnts []*PjsuaAccount, localPort, waitDur time.Duration) (*os.File, error) {
	cmdArgs := []string{fmt.Sprintf("--local-port=%d", localPort), "--null-audio", "--auto-answer=200", "--max-calls=32", "--app-log-level=0"}
	for idx, acnt := range acnts {
		if idx != 0 {
			cmdArgs = append(cmdArgs, "--next-account")
		}
		cmdArgs = append(cmdArgs, "--id="+acnt.Id, "--registrar="+acnt.Registrar, "--username="+acnt.Username, "--password="+acnt.Password, "--realm="+acnt.Realm)
	}
	pjsuaPath, err := exec.LookPath("pjsua")
	if err != nil {
		return nil, err
	}
	pjsua := exec.Command(pjsuaPath, cmdArgs...)
	fPty, err := pty.Start(pjsua)
	if err != nil {
		return nil, err
	}
	buf := new(bytes.Buffer)
	io.Copy(os.Stdout, buf) // Free the content since otherwise pjsua will not start
	time.Sleep(waitDur)     // Give time to rater to fire up
	return fPty, nil
}

func PjsuaCallUri(acnt *PjsuaAccount, dstUri, outboundUri string, callDur time.Duration, localPort int) error {
	cmdArgs := []string{"--null-audio", "--app-log-level=0", fmt.Sprintf("--local-port=%d", localPort), fmt.Sprintf("--duration=%d", int(callDur.Seconds())),
		"--outbound=" + outboundUri, "--id=" + acnt.Id, "--username=" + acnt.Username, "--password=" + acnt.Password, "--realm=" + acnt.Realm, dstUri}

	pjsuaPath, err := exec.LookPath("pjsua")
	if err != nil {
		return err
	}
	pjsua := exec.Command(pjsuaPath, cmdArgs...)
	fPty, err := pty.Start(pjsua)
	if err != nil {
		return err
	}
	buf := new(bytes.Buffer)
	io.Copy(os.Stdout, buf)
	go func() {
		time.Sleep(callDur + 2*time.Second)
		fPty.Write([]byte("q\n")) // Destroy the listener
	}()
	return nil
}

func KillProcName(procName string, waitMs int) (err error) {
	if err = exec.Command("pkill", procName).Run(); err != nil {
		return
	}
	time.Sleep(time.Duration(waitMs) * time.Millisecond)
	return
}

func ForceKillProcName(procName string, waitMs int) error {
	if err := exec.Command("pkill", "-9", procName).Run(); err != nil {
		return err
	}
	time.Sleep(time.Duration(waitMs) * time.Millisecond)
	return nil
}

func CallScript(scriptPath string, subcommand string, waitMs int) error {
	if err := exec.Command(scriptPath, subcommand).Run(); err != nil {
		return err
	}
	time.Sleep(time.Duration(waitMs) * time.Millisecond) // Give time to rater to fire up
	return nil
}

func GetDefaultEmptyCacheStats() map[string]*ltcache.CacheStats {
	return map[string]*ltcache.CacheStats{
		utils.MetaDefault:                       {},
		utils.CacheAccountActionPlans:           {},
		utils.CacheActionPlans:                  {},
		utils.CacheActionTriggers:               {},
		utils.CacheActions:                      {},
		utils.CacheAttributeFilterIndexes:       {},
		utils.CacheAttributeProfiles:            {},
		utils.CacheChargerFilterIndexes:         {},
		utils.CacheChargerProfiles:              {},
		utils.CacheDispatcherFilterIndexes:      {},
		utils.CacheDispatcherProfiles:           {},
		utils.CacheDispatcherHosts:              {},
		utils.CacheDispatcherRoutes:             {},
		utils.CacheDispatcherLoads:              {},
		utils.CacheDispatchers:                  {},
		utils.CacheDestinations:                 {},
		utils.CacheEventResources:               {},
		utils.CacheFilters:                      {},
		utils.CacheRatingPlans:                  {},
		utils.CacheRatingProfiles:               {},
		utils.CacheResourceFilterIndexes:        {},
		utils.CacheResourceProfiles:             {},
		utils.CacheResources:                    {},
		utils.CacheReverseDestinations:          {},
		utils.CacheRPCResponses:                 {},
		utils.CacheSharedGroups:                 {},
		utils.CacheStatFilterIndexes:            {},
		utils.CacheStatQueueProfiles:            {},
		utils.CacheStatQueues:                   {},
		utils.CacheSTIR:                         {},
		utils.CacheRouteFilterIndexes:           {},
		utils.CacheRouteProfiles:                {},
		utils.CacheThresholdFilterIndexes:       {},
		utils.CacheThresholdProfiles:            {},
		utils.CacheThresholds:                   {},
		utils.CacheRateProfiles:                 {},
		utils.CacheRateProfilesFilterIndexes:    {},
		utils.CacheRateFilterIndexes:            {},
		utils.CacheTimings:                      {},
		utils.CacheDiameterMessages:             {},
		utils.CacheClosedSessions:               {},
		utils.CacheLoadIDs:                      {},
		utils.CacheRPCConnections:               {},
		utils.CacheCDRIDs:                       {},
		utils.CacheRatingProfilesTmp:            {},
		utils.CacheUCH:                          {},
		utils.CacheEventCharges:                 {},
		utils.CacheReverseFilterIndexes:         {},
		utils.MetaAPIBan:                        {},
		utils.CacheCapsEvents:                   {},
		utils.CacheActionProfiles:               {},
		utils.CacheActionProfilesFilterIndexes:  {},
		utils.CacheAccountProfiles:              {},
		utils.CacheAccountProfilesFilterIndexes: {},
		utils.CacheReplicationHosts:             {},

		utils.CacheAccounts:              {},
		utils.CacheVersions:              {},
		utils.CacheTBLTPTimings:          {},
		utils.CacheTBLTPDestinations:     {},
		utils.CacheTBLTPRates:            {},
		utils.CacheTBLTPDestinationRates: {},
		utils.CacheTBLTPRatingPlans:      {},
		utils.CacheTBLTPRatingProfiles:   {},
		utils.CacheTBLTPSharedGroups:     {},
		utils.CacheTBLTPActions:          {},
		utils.CacheTBLTPActionPlans:      {},
		utils.CacheTBLTPActionTriggers:   {},
		utils.CacheTBLTPAccountActions:   {},
		utils.CacheTBLTPResources:        {},
		utils.CacheTBLTPStats:            {},
		utils.CacheTBLTPThresholds:       {},
		utils.CacheTBLTPFilters:          {},
		utils.CacheSessionCostsTBL:       {},
		utils.CacheCDRsTBL:               {},
		utils.CacheTBLTPRoutes:           {},
		utils.CacheTBLTPAttributes:       {},
		utils.CacheTBLTPChargers:         {},
		utils.CacheTBLTPDispatchers:      {},
		utils.CacheTBLTPDispatcherHosts:  {},
		utils.CacheTBLTPRateProfiles:     {},
		utils.CacheTBLTPActionProfiles:   {},
		utils.CacheTBLTPAccountProfiles:  {},
	}
}

func GetDefaultEmptyArgCachePrefix() map[string][]string {
	return map[string][]string{
		utils.DestinationPrefix:             nil,
		utils.ReverseDestinationPrefix:      nil,
		utils.RatingPlanPrefix:              nil,
		utils.RatingProfilePrefix:           nil,
		utils.ActionPrefix:                  nil,
		utils.ActionPlanPrefix:              nil,
		utils.AccountActionPlansPrefix:      nil,
		utils.ActionTriggerPrefix:           nil,
		utils.SharedGroupPrefix:             nil,
		utils.ResourceProfilesPrefix:        nil,
		utils.ResourcesPrefix:               nil,
		utils.StatQueuePrefix:               nil,
		utils.StatQueueProfilePrefix:        nil,
		utils.ThresholdPrefix:               nil,
		utils.ThresholdProfilePrefix:        nil,
		utils.FilterPrefix:                  nil,
		utils.RouteProfilePrefix:            nil,
		utils.AttributeProfilePrefix:        nil,
		utils.ChargerProfilePrefix:          nil,
		utils.DispatcherProfilePrefix:       nil,
		utils.DispatcherHostPrefix:          nil,
		utils.RateProfilePrefix:             nil,
		utils.ActionProfilePrefix:           nil,
		utils.TimingsPrefix:                 nil,
		utils.AttributeFilterIndexes:        nil,
		utils.ResourceFilterIndexes:         nil,
		utils.StatFilterIndexes:             nil,
		utils.ThresholdFilterIndexes:        nil,
		utils.RouteFilterIndexes:            nil,
		utils.ChargerFilterIndexes:          nil,
		utils.DispatcherFilterIndexes:       nil,
		utils.RateProfilesFilterIndexPrfx:   nil,
		utils.RateFilterIndexPrfx:           nil,
		utils.ActionProfilesFilterIndexPrfx: nil,
		utils.FilterIndexPrfx:               nil,
	}
}
