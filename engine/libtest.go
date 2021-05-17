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
	"bytes"
	"fmt"
	"io"
	"net/rpc/jsonrpc"
	"os"
	"os/exec"
	"path"
	"time"

	"github.com/cgrates/birpc/context"

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

	ResourcesCSVContent = `
#Tenant[0],Id[1],FilterIDs[2],Weight[3],TTL[4],Limit[5],AllocationMessage[6],Blocker[7],Stored[8],Thresholds[9]
cgrates.org,ResGroup21,*string:~*req.Account:1001,10,1s,2,call,true,true,
cgrates.org,ResGroup22,*string:~*req.Account:dan,10,3600s,2,premium_call,true,true,
`
	StatsCSVContent = `
#Tenant[0],Id[1],FilterIDs[2],Weight[3],QueueLength[4],TTL[5],MinItems[6],Metrics[7],MetricFilterIDs[8],Stored[9],Blocker[10],ThresholdIDs[11]
cgrates.org,TestStats,*string:~*req.Account:1001,20,100,1s,2,*sum#~*req.Value;*average#~*req.Value,,true,true,Th1;Th2
cgrates.org,TestStats,,20,,,2,*sum#~*req.Usage,,true,true,
cgrates.org,TestStats2,FLTR_1,20,100,1s,2,*sum#~*req.Value;*sum#~*req.Usage;*average#~*req.Value;*average#~*req.Usage,,true,true,Th
cgrates.org,TestStats2,,20,,,2,*sum#~*req.Cost;*average#~*req.Cost,,true,true,
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
#Tenant[0],ID[1],FilterIDs[2],Weight[3],Sorting[4],SortingParameters[5],RouteID[6],RouteFilterIDs[7],RouteAccountIDs[8],RouteRatingPlanIDs[9],RouteResourceIDs[10],RouteStatIDs[11],RouteWeight[12],RouteBlocker[13],RouteParameters[14]
cgrates.org,RoutePrf1,*string:~*req.Account:dan,20,*lc,,route1,FLTR_ACNT_dan,Account1;Account1_1,RPL_1,ResGroup1,Stat1,10,true,param1
cgrates.org,RoutePrf1,,,,,route1,,,RPL_2,ResGroup2,,10,,
cgrates.org,RoutePrf1,,,,,route1,FLTR_DST_DE,Account2,RPL_3,ResGroup3,Stat2,10,,
cgrates.org,RoutePrf1,,,,,route1,,,,ResGroup4,Stat3,10,,
`
	AttributesCSVContent = `
#Tenant,ID,Contexts,FilterIDs,Weight,AttributeFilterIDs,Path,Type,Value,Blocker
cgrates.org,ALS1,con1,*string:~*req.Account:1001,20,*string:~*req.Field1:Initial,*req.Field1,*variable,Sub1,true
cgrates.org,ALS1,con2;con3,,20,,*req.Field2,*variable,Sub2,true
`
	ChargersCSVContent = `
#Tenant,ID,FilterIDs,Weight,RunID,AttributeIDs
cgrates.org,Charger1,*string:~*req.Account:1001,20,*rated,ATTR_1001_SIMPLEAUTH
`
	DispatcherCSVContent = `
#Tenant,ID,Subsystems,FilterIDs,Weight,Strategy,StrategyParameters,ConnID,ConnFilterIDs,ConnWeight,ConnBlocker,ConnParameters
cgrates.org,D1,*any,*string:~*req.Account:1001,20,*first,,C1,*gt:~*req.Usage:10,10,false,192.168.56.203
cgrates.org,D1,,,,*first,,C2,*lt:~*req.Usage:10,10,false,192.168.56.204
`
	DispatcherHostCSVContent = `
#Tenant[0],ID[1],Address[2],Transport[3],TLS[4]
cgrates.org,ALL1,127.0.0.1:2012,*json,true
`
	RateProfileCSVContent = `
#Tenant,ID,FilterIDs,Weights,MinCost,MaxCost,MaxCostStrategy,RateID,RateFilterIDs,RateActivationStart,RateWeights,RateBlocker,RateIntervalStart,RateFixedFee,RateRecurrentFee,RateUnit,RateIncrement
cgrates.org,RP1,*string:~*req.Subject:1001,;0,0.1,0.6,*free,RT_WEEK,,"* * * * 1-5",;0,false,0s,0,0.12,1m,1m
cgrates.org,RP1,,,,,,RT_WEEK,,,,,1m,1.234,0.06,1m,1s
cgrates.org,RP1,,,,,,RT_WEEKEND,,"* * * * 0,6",;10,false,0s,0.089,0.06,1m,1s
cgrates.org,RP1,,,,,,RT_CHRISTMAS,,* * 24 12 *,;30,false,0s,0.0564,0.06,1m,1s
`
	ActionProfileCSVContent = `
#Tenant,ID,FilterIDs,Weight,Schedule,TargetType,TargetIDs,ActionID,ActionFilterIDs,ActionBlocker,ActionTTL,ActionType,ActionOpts,ActionPath,ActionValue
cgrates.org,ONE_TIME_ACT,,10,*asap,*accounts,1001;1002,TOPUP,,false,0s,*add_balance,,*balance.TestBalance.Value,10
cgrates.org,ONE_TIME_ACT,,,,,,SET_BALANCE_TEST_DATA,,false,0s,*set_balance,,*balance.TestDataBalance.Type,*data
cgrates.org,ONE_TIME_ACT,,,,,,TOPUP_TEST_DATA,,false,0s,*add_balance,,*balance.TestDataBalance.Value,1024
cgrates.org,ONE_TIME_ACT,,,,,,SET_BALANCE_TEST_VOICE,,false,0s,*set_balance,,*balance.TestVoiceBalance.Type,*voice
cgrates.org,ONE_TIME_ACT,,,,,,TOPUP_TEST_VOICE,,false,0s,*add_balance,,*balance.TestVoiceBalance.Value,15m15s
cgrates.org,ONE_TIME_ACT,,,,,,TOPUP_TEST_VOICE,,false,0s,*add_balance,,*balance.TestVoiceBalance2.Value,15m15s
`

	AccountCSVContent = `
#Tenant,ID,FilterIDs,ActivationInterval,Weights,Opts,BalanceID,BalanceFilterIDs,BalanceWeights,BalanceType,BalanceUnits,BalanceUnitFactors,BalanceOpts,BalanceCostIncrements,BalanceAttributeIDs,BalanceRateProfileIDs,ThresholdIDs
cgrates.org,1001,,,;20,,MonetaryBalance,,;10,*monetary,14,fltr1&fltr2;100;fltr3;200,,fltr1&fltr2;1.3;2.3;3.3,attr1;attr2,,*none
cgrates.org,1001,,,,,VoiceBalance,,;10,*voice,3600000000000,,,,,,
`
)

func InitDataDB(cfg *config.CGRConfig) error {
	d, err := NewDataDBConn(cfg.DataDbCfg().Type,
		cfg.DataDbCfg().Host, cfg.DataDbCfg().Port,
		cfg.DataDbCfg().Name, cfg.DataDbCfg().User,
		cfg.DataDbCfg().Password, cfg.GeneralCfg().DBDataEncoding,
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

func InitStorDB(cfg *config.CGRConfig) error {
	storDB, err := NewStorDBConn(cfg.StorDbCfg().Type,
		cfg.StorDbCfg().Host, cfg.StorDbCfg().Port,
		cfg.StorDbCfg().Name, cfg.StorDbCfg().User,
		cfg.StorDbCfg().Password, cfg.GeneralCfg().DBDataEncoding,
		cfg.StorDbCfg().StringIndexedFields, cfg.StorDbCfg().PrefixIndexedFields,
		cfg.StorDbCfg().Opts)
	if err != nil {
		return err
	}
	if err := storDB.Flush(path.Join(cfg.DataFolderPath, "storage",
		cfg.StorDbCfg().Type)); err != nil {
		return err
	}
	if utils.IsSliceMember([]string{utils.Mongo, utils.MySQL, utils.Postgres},
		cfg.StorDbCfg().Type) {
		if err := SetDBVersions(storDB); err != nil {
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

func LoadTariffPlanFromFolder(tpPath, timezone string, dm *DataManager, disableReverse bool,
	cacheConns, schedConns []string) error {
	loader, err := NewTpReader(dm.dataDB, NewFileCSVStorage(utils.CSVSep, tpPath), "",
		timezone, cacheConns, schedConns, false)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	if err := loader.LoadAll(); err != nil {
		return utils.NewErrServerError(err)
	}
	if err := loader.WriteToDatabase(false, disableReverse); err != nil {
		return utils.NewErrServerError(err)
	}
	return nil
}

type PjsuaAccount struct {
	ID, Username, Password, Realm, Registrar string
}

// Returns file reference where we can write to control pjsua in terminal
func StartPjsuaListener(acnts []*PjsuaAccount, localPort, waitDur time.Duration) (*os.File, error) {
	cmdArgs := []string{fmt.Sprintf("--local-port=%d", localPort), "--null-audio", "--auto-answer=200", "--max-calls=32", "--app-log-level=0"}
	for idx, acnt := range acnts {
		if idx != 0 {
			cmdArgs = append(cmdArgs, "--next-account")
		}
		cmdArgs = append(cmdArgs, "--id="+acnt.ID, "--registrar="+acnt.Registrar, "--username="+acnt.Username, "--password="+acnt.Password, "--realm="+acnt.Realm)
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

func PjsuaCallURI(acnt *PjsuaAccount, dstURI, outboundURI string, callDur time.Duration, localPort int) error {
	cmdArgs := []string{"--null-audio", "--app-log-level=0", fmt.Sprintf("--local-port=%d", localPort), fmt.Sprintf("--duration=%d", int(callDur.Seconds())),
		"--outbound=" + outboundURI, "--id=" + acnt.ID, "--username=" + acnt.Username, "--password=" + acnt.Password, "--realm=" + acnt.Realm, dstURI}

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
		utils.MetaDefault:                      {},
		utils.CacheAttributeFilterIndexes:      {},
		utils.CacheAttributeProfiles:           {},
		utils.CacheChargerFilterIndexes:        {},
		utils.CacheChargerProfiles:             {},
		utils.CacheDispatcherFilterIndexes:     {},
		utils.CacheDispatcherProfiles:          {},
		utils.CacheDispatcherHosts:             {},
		utils.CacheDispatcherRoutes:            {},
		utils.CacheDispatcherLoads:             {},
		utils.CacheDispatchers:                 {},
		utils.CacheEventResources:              {},
		utils.CacheFilters:                     {},
		utils.CacheResourceFilterIndexes:       {},
		utils.CacheResourceProfiles:            {},
		utils.CacheResources:                   {},
		utils.CacheRPCResponses:                {},
		utils.CacheStatFilterIndexes:           {},
		utils.CacheStatQueueProfiles:           {},
		utils.CacheStatQueues:                  {},
		utils.CacheSTIR:                        {},
		utils.CacheRouteFilterIndexes:          {},
		utils.CacheRouteProfiles:               {},
		utils.CacheThresholdFilterIndexes:      {},
		utils.CacheThresholdProfiles:           {},
		utils.CacheThresholds:                  {},
		utils.CacheRateProfiles:                {},
		utils.CacheRateProfilesFilterIndexes:   {},
		utils.CacheRateFilterIndexes:           {},
		utils.CacheTimings:                     {},
		utils.CacheDiameterMessages:            {},
		utils.CacheClosedSessions:              {},
		utils.CacheLoadIDs:                     {},
		utils.CacheRPCConnections:              {},
		utils.CacheCDRIDs:                      {},
		utils.CacheUCH:                         {},
		utils.CacheEventCharges:                {},
		utils.CacheReverseFilterIndexes:        {},
		utils.MetaAPIBan:                       {},
		utils.CacheCapsEvents:                  {},
		utils.CacheActionProfiles:              {},
		utils.CacheActionProfilesFilterIndexes: {},
		utils.CacheAccounts:                    {},
		utils.CacheAccountsFilterIndexes:       {},
		utils.CacheReplicationHosts:            {},

		utils.CacheVersions:             {},
		utils.CacheTBLTPTimings:         {},
		utils.CacheTBLTPResources:       {},
		utils.CacheTBLTPStats:           {},
		utils.CacheTBLTPThresholds:      {},
		utils.CacheTBLTPFilters:         {},
		utils.CacheSessionCostsTBL:      {},
		utils.CacheCDRsTBL:              {},
		utils.CacheTBLTPRoutes:          {},
		utils.CacheTBLTPAttributes:      {},
		utils.CacheTBLTPChargers:        {},
		utils.CacheTBLTPDispatchers:     {},
		utils.CacheTBLTPDispatcherHosts: {},
		utils.CacheTBLTPRateProfiles:    {},
		utils.CacheTBLTPActionProfiles:  {},
		utils.CacheTBLTPAccounts:        {},
	}
}
