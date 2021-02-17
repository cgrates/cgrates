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

package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/cgrates/rpcclient"

	"github.com/cgrates/cgrates/utils"
)

var (
	dbDefaultsCfg     dbDefaults
	cgrCfg            *CGRConfig  // will be shared
	dfltFsConnConfig  *FsConnCfg  // Default FreeSWITCH Connection configuration, built out of json default configuration
	dfltKamConnConfig *KamConnCfg // Default Kamailio Connection configuration
	dfltRemoteHost    *RemoteHost
	dfltAstConnCfg    *AsteriskConnCfg
	dfltLoaderConfig  *LoaderSCfg
)

func newDbDefaults() dbDefaults {
	deflt := dbDefaults{
		utils.MySQL: map[string]string{
			"DbName": "cgrates",
			"DbPort": "3306",
			"DbPass": "CGRateS.org",
		},
		utils.Postgres: map[string]string{
			"DbName": "cgrates",
			"DbPort": "5432",
			"DbPass": "CGRateS.org",
		},
		utils.Mongo: map[string]string{
			"DbName": "cgrates",
			"DbPort": "27017",
			"DbPass": "",
		},
		utils.Redis: map[string]string{
			"DbName": "10",
			"DbPort": "6379",
			"DbPass": "",
		},
		utils.INTERNAL: map[string]string{
			"DbName": "internal",
			"DbPort": "internal",
			"DbPass": "internal",
		},
	}
	return deflt
}

type dbDefaults map[string]map[string]string

func (dbDflt dbDefaults) dbName(dbType string, flagInput string) string {
	if flagInput != utils.MetaDynamic {
		return flagInput
	}
	return dbDflt[dbType]["DbName"]
}

func (dbDefaults) dbUser(dbType string, flagInput string) string {
	if flagInput != utils.MetaDynamic {
		return flagInput
	}
	return utils.CGRateSLwr
}

func (dbDefaults) dbHost(dbType string, flagInput string) string {
	if flagInput != utils.MetaDynamic {
		return flagInput
	}
	return utils.Localhost
}

func (dbDflt dbDefaults) dbPort(dbType string, flagInput string) string {
	if flagInput != utils.MetaDynamic {
		return flagInput
	}
	return dbDflt[dbType]["DbPort"]
}

func (dbDflt dbDefaults) dbPass(dbType string, flagInput string) string {
	if flagInput != utils.MetaDynamic {
		return flagInput
	}
	return dbDflt[dbType]["DbPass"]
}

func init() {
	cgrCfg = NewDefaultCGRConfig()
	dbDefaultsCfg = newDbDefaults()
}

// CgrConfig is used to retrieve system configuration from other packages
func CgrConfig() *CGRConfig {
	return cgrCfg
}

// SetCgrConfig is used to set system configuration from other places
func SetCgrConfig(cfg *CGRConfig) {
	cgrCfg = cfg
}

// NewDefaultCGRConfig returns the default configuration
func NewDefaultCGRConfig() (cfg *CGRConfig) {
	cfg, _ = newCGRConfig([]byte(CGRATES_CFG_JSON))
	return
}

func newCGRConfig(config []byte) (cfg *CGRConfig, err error) {
	cfg = new(CGRConfig)
	cfg.initChanels()
	cfg.DataFolderPath = "/usr/share/cgrates/"

	cfg.rpcConns = make(map[string]*RPCConn)
	cfg.templates = make(map[string][]*FCTemplate)
	cfg.generalCfg = new(GeneralCfg)
	cfg.generalCfg.NodeID = utils.UUIDSha1Prefix()
	cfg.dataDbCfg = new(DataDbCfg)
	cfg.dataDbCfg.Items = make(map[string]*ItemOpt)
	cfg.dataDbCfg.Opts = make(map[string]interface{})
	cfg.storDbCfg = new(StorDbCfg)
	cfg.storDbCfg.Items = make(map[string]*ItemOpt)
	cfg.storDbCfg.Opts = make(map[string]interface{})
	cfg.tlsCfg = new(TLSCfg)
	cfg.cacheCfg = new(CacheCfg)
	cfg.cacheCfg.Partitions = make(map[string]*CacheParamCfg)
	cfg.listenCfg = new(ListenCfg)
	cfg.httpCfg = new(HTTPCfg)
	cfg.httpCfg.ClientOpts = make(map[string]interface{})
	cfg.filterSCfg = new(FilterSCfg)
	cfg.ralsCfg = new(RalsCfg)
	cfg.ralsCfg.MaxComputedUsage = make(map[string]time.Duration)
	cfg.ralsCfg.BalanceRatingSubject = make(map[string]string)
	cfg.schedulerCfg = new(SchedulerCfg)
	cfg.cdrsCfg = new(CdrsCfg)
	cfg.analyzerSCfg = new(AnalyzerSCfg)
	cfg.sessionSCfg = new(SessionSCfg)
	cfg.sessionSCfg.STIRCfg = new(STIRcfg)
	cfg.sessionSCfg.DefaultUsage = make(map[string]time.Duration)
	cfg.fsAgentCfg = new(FsAgentCfg)
	cfg.kamAgentCfg = new(KamAgentCfg)
	cfg.asteriskAgentCfg = new(AsteriskAgentCfg)
	cfg.diameterAgentCfg = new(DiameterAgentCfg)
	cfg.radiusAgentCfg = new(RadiusAgentCfg)
	cfg.dnsAgentCfg = new(DNSAgentCfg)
	cfg.attributeSCfg = new(AttributeSCfg)
	cfg.chargerSCfg = new(ChargerSCfg)
	cfg.resourceSCfg = new(ResourceSConfig)
	cfg.statsCfg = new(StatSCfg)
	cfg.thresholdSCfg = new(ThresholdSCfg)
	cfg.routeSCfg = new(RouteSCfg)
	cfg.sureTaxCfg = new(SureTaxCfg)
	cfg.dispatcherSCfg = new(DispatcherSCfg)
	cfg.dispatcherHCfg = new(DispatcherHCfg)
	cfg.dispatcherHCfg.Hosts = make(map[string][]*DispatcherHRegistarCfg)
	cfg.loaderCgrCfg = new(LoaderCgrCfg)
	cfg.migratorCgrCfg = new(MigratorCgrCfg)
	cfg.migratorCgrCfg.OutDataDBOpts = make(map[string]interface{})
	cfg.migratorCgrCfg.OutStorDBOpts = make(map[string]interface{})
	cfg.mailerCfg = new(MailerCfg)
	cfg.loaderCfg = make(LoaderSCfgs, 0)
	cfg.apier = new(ApierCfg)
	cfg.ersCfg = new(ERsCfg)
	cfg.eesCfg = new(EEsCfg)
	cfg.eesCfg.Cache = make(map[string]*CacheParamCfg)
	cfg.rateSCfg = new(RateSCfg)
	cfg.actionSCfg = new(ActionSCfg)
	cfg.sipAgentCfg = new(SIPAgentCfg)
	cfg.configSCfg = new(ConfigSCfg)
	cfg.apiBanCfg = new(APIBanCfg)
	cfg.coreSCfg = new(CoreSCfg)
	cfg.accountSCfg = new(AccountSCfg)

	cfg.cacheDP = make(map[string]utils.MapStorage)

	var cgrJSONCfg *CgrJsonCfg
	if cgrJSONCfg, err = NewCgrJsonCfgFromBytes(config); err != nil {
		return
	}
	if err = cfg.loadFromJSONCfg(cgrJSONCfg); err != nil {
		return
	}

	// populate default ERs reader
	for _, ersRdr := range cfg.ersCfg.Readers {
		if ersRdr.ID == utils.MetaDefault {
			cfg.dfltEvRdr = ersRdr.Clone()
			break
		}
	}
	// populate default EEs exporter
	for _, ersExp := range cfg.eesCfg.Exporters {
		if ersExp.ID == utils.MetaDefault {
			cfg.dfltEvExp = ersExp.Clone()
			break
		}
	}
	dfltFsConnConfig = cfg.fsAgentCfg.EventSocketConns[0] // We leave it crashing here on purpose if no Connection defaults defined
	dfltKamConnConfig = cfg.kamAgentCfg.EvapiConns[0]
	dfltAstConnCfg = cfg.asteriskAgentCfg.AsteriskConns[0]
	dfltLoaderConfig = cfg.loaderCfg[0].Clone()
	dfltRemoteHost = new(RemoteHost)
	*dfltRemoteHost = *cfg.rpcConns[utils.MetaLocalHost].Conns[0]
	err = cfg.checkConfigSanity()
	return
}

// NewCGRConfigFromJSONStringWithDefaults returns the given config with the default option loaded
func NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr string) (cfg *CGRConfig, err error) {
	cfg = NewDefaultCGRConfig()
	jsnCfg := new(CgrJsonCfg)
	if err = NewRjReaderFromBytes([]byte(cfgJSONStr)).Decode(jsnCfg); err != nil {
		return
	} else if err = cfg.loadFromJSONCfg(jsnCfg); err != nil {
		return
	}
	return
}

// NewCGRConfigFromPath reads all json files out of a folder/subfolders and loads them up in lexical order
func NewCGRConfigFromPath(path string) (cfg *CGRConfig, err error) {
	cfg = NewDefaultCGRConfig()
	cfg.ConfigPath = path

	if err = cfg.loadConfigFromPath(path, []func(*CgrJsonCfg) error{cfg.loadFromJSONCfg}, false); err != nil {
		return
	}
	err = cfg.checkConfigSanity()
	return
}

// newCGRConfigFromPathWithoutEnv reads all json files out of a folder/subfolders and loads them up in lexical order
// it will not read *env variables and will not checkConfigSanity as it is not needed for configs
func newCGRConfigFromPathWithoutEnv(path string) (cfg *CGRConfig, err error) {
	cfg = NewDefaultCGRConfig()
	cfg.ConfigPath = path

	err = cfg.loadConfigFromPath(path, []func(*CgrJsonCfg) error{cfg.loadFromJSONCfg}, true)
	return
}

func isHidden(fileName string) bool {
	if fileName == "." || fileName == ".." {
		return false
	}
	return strings.HasPrefix(fileName, ".")
}

// CGRConfig holds system configuration, defaults are overwritten with values from config file if found
type CGRConfig struct {
	lks            map[string]*sync.RWMutex
	DataFolderPath string // Path towards data folder, for tests internal usage, not loading out of .json options
	ConfigPath     string // Path towards config

	// Cache defaults loaded from json and needing clones
	dfltEvRdr *EventReaderCfg   // default event reader
	dfltEvExp *EventExporterCfg // default event exporter

	loaderCfg    LoaderSCfgs   // LoaderS configs
	httpAgentCfg HTTPAgentCfgs // HttpAgent configs

	rldChans map[string]chan struct{} // index here the channels used for reloads

	rpcConns RPCConns

	templates FcTemplates

	generalCfg       *GeneralCfg       // General config
	dataDbCfg        *DataDbCfg        // Database config
	storDbCfg        *StorDbCfg        // StroreDb config
	tlsCfg           *TLSCfg           // TLS config
	cacheCfg         *CacheCfg         // Cache config
	listenCfg        *ListenCfg        // Listen config
	httpCfg          *HTTPCfg          // HTTP config
	filterSCfg       *FilterSCfg       // FilterS config
	ralsCfg          *RalsCfg          // Rals config
	schedulerCfg     *SchedulerCfg     // Scheduler config
	cdrsCfg          *CdrsCfg          // Cdrs config
	sessionSCfg      *SessionSCfg      // SessionS config
	fsAgentCfg       *FsAgentCfg       // FreeSWITCHAgent config
	kamAgentCfg      *KamAgentCfg      // KamailioAgent config
	asteriskAgentCfg *AsteriskAgentCfg // AsteriskAgent config
	diameterAgentCfg *DiameterAgentCfg // DiameterAgent config
	radiusAgentCfg   *RadiusAgentCfg   // RadiusAgent config
	dnsAgentCfg      *DNSAgentCfg      // DNSAgent config
	attributeSCfg    *AttributeSCfg    // AttributeS config
	chargerSCfg      *ChargerSCfg      // ChargerS config
	resourceSCfg     *ResourceSConfig  // ResourceS config
	statsCfg         *StatSCfg         // StatS config
	thresholdSCfg    *ThresholdSCfg    // ThresholdS config
	routeSCfg        *RouteSCfg        // RouteS config
	sureTaxCfg       *SureTaxCfg       // SureTax config
	dispatcherSCfg   *DispatcherSCfg   // DispatcherS config
	dispatcherHCfg   *DispatcherHCfg   // DispatcherH config
	loaderCgrCfg     *LoaderCgrCfg     // LoaderCgr config
	migratorCgrCfg   *MigratorCgrCfg   // MigratorCgr config
	mailerCfg        *MailerCfg        // Mailer config
	analyzerSCfg     *AnalyzerSCfg     // AnalyzerS config
	apier            *ApierCfg         // APIer config
	ersCfg           *ERsCfg           // EventReader config
	eesCfg           *EEsCfg           // EventExporter config
	rateSCfg         *RateSCfg         // RateS config
	actionSCfg       *ActionSCfg       // ActionS config
	sipAgentCfg      *SIPAgentCfg      // SIPAgent config
	configSCfg       *ConfigSCfg       // ConfigS config
	apiBanCfg        *APIBanCfg        // APIBan config
	coreSCfg         *CoreSCfg         // CoreS config
	accountSCfg      *AccountSCfg      // AccountS config

	cacheDP    map[string]utils.MapStorage
	cacheDPMux sync.RWMutex
}

var posibleLoaderTypes = utils.NewStringSet([]string{utils.MetaAttributes,
	utils.MetaResources, utils.MetaFilters, utils.MetaStats,
	utils.MetaRoutes, utils.MetaThresholds, utils.MetaChargers,
	utils.MetaDispatchers, utils.MetaDispatcherHosts, utils.MetaRateProfiles})

var possibleReaderTypes = utils.NewStringSet([]string{utils.MetaFileCSV,
	utils.MetaKafkajsonMap, utils.MetaFileXML, utils.MetaSQL, utils.MetaFileFWV,
	utils.MetaPartialCSV, utils.MetaFlatstore, utils.MetaFileJSON, utils.MetaNone})

var possibleExporterTypes = utils.NewStringSet([]string{utils.MetaFileCSV, utils.MetaNone, utils.MetaFileFWV,
	utils.MetaHTTPPost, utils.MetaHTTPjsonMap, utils.MetaAMQPjsonMap, utils.MetaAMQPV1jsonMap, utils.MetaSQSjsonMap,
	utils.MetaKafkajsonMap, utils.MetaS3jsonMap, utils.MetaElastic, utils.MetaVirt, utils.MetaSQL})

// LazySanityCheck used after check config sanity to display warnings related to the config
func (cfg *CGRConfig) LazySanityCheck() {
	for _, expID := range cfg.cdrsCfg.OnlineCDRExports {
		for _, ee := range cfg.eesCfg.Exporters {
			if ee.ID == expID && ee.Type == utils.MetaS3jsonMap || ee.Type == utils.MetaSQSjsonMap {
				poster := utils.SQSPoster
				if ee.Type == utils.MetaS3jsonMap {
					poster = utils.S3Poster
				}
				argsMap := utils.GetUrlRawArguments(ee.ExportPath)
				for _, arg := range []string{utils.AWSRegion, utils.AWSKey, utils.AWSSecret} {
					if _, has := argsMap[arg]; !has {
						utils.Logger.Warning(fmt.Sprintf("<%s> No %s present for AWS for exporter with ID : <%s>.", poster, arg, ee.ID))
					}
				}
			}
		}
	}
	for _, exporter := range cfg.eesCfg.Exporters {
		if exporter.Type == utils.MetaS3jsonMap || exporter.Type == utils.MetaSQSjsonMap {
			poster := utils.SQSPoster
			if exporter.Type == utils.MetaS3jsonMap {
				poster = utils.S3Poster
			}
			argsMap := utils.GetUrlRawArguments(exporter.ExportPath)
			for _, arg := range []string{utils.AWSRegion, utils.AWSKey, utils.AWSSecret} {
				if _, has := argsMap[arg]; !has {
					utils.Logger.Warning(fmt.Sprintf("<%s> No %s present for AWS for exporter with ID: <%s>.", poster, arg, exporter.ID))
				}
			}
		}
	}
}

// Loads from json configuration object, will be used for defaults, config from file and reload, might need lock
func (cfg *CGRConfig) loadFromJSONCfg(jsnCfg *CgrJsonCfg) (err error) {
	// Load sections out of JSON config, stop on error
	for _, loadFunc := range []func(*CgrJsonCfg) error{
		cfg.loadRPCConns,
		cfg.loadGeneralCfg, cfg.loadTemplateSCfg, cfg.loadCacheCfg, cfg.loadListenCfg,
		cfg.loadHTTPCfg, cfg.loadDataDBCfg, cfg.loadStorDBCfg,
		cfg.loadFilterSCfg, cfg.loadRalSCfg, cfg.loadSchedulerCfg,
		cfg.loadCdrsCfg, cfg.loadSessionSCfg,
		cfg.loadFreeswitchAgentCfg, cfg.loadKamAgentCfg,
		cfg.loadAsteriskAgentCfg, cfg.loadDiameterAgentCfg, cfg.loadRadiusAgentCfg,
		cfg.loadDNSAgentCfg, cfg.loadHTTPAgentCfg, cfg.loadAttributeSCfg,
		cfg.loadChargerSCfg, cfg.loadResourceSCfg, cfg.loadStatSCfg,
		cfg.loadThresholdSCfg, cfg.loadRouteSCfg, cfg.loadLoaderSCfg,
		cfg.loadMailerCfg, cfg.loadSureTaxCfg, cfg.loadDispatcherSCfg,
		cfg.loadLoaderCgrCfg, cfg.loadMigratorCgrCfg, cfg.loadTLSCgrCfg,
		cfg.loadAnalyzerCgrCfg, cfg.loadApierCfg, cfg.loadErsCfg, cfg.loadEesCfg,
		cfg.loadRateSCfg, cfg.loadSIPAgentCfg, cfg.loadDispatcherHCfg,
		cfg.loadConfigSCfg, cfg.loadAPIBanCgrCfg, cfg.loadCoreSCfg, cfg.loadActionSCfg,
		cfg.loadAccountSCfg} {
		if err = loadFunc(jsnCfg); err != nil {
			return
		}
	}
	return
}

// loadRPCConns loads the RPCConns section of the configuration
func (cfg *CGRConfig) loadRPCConns(jsnCfg *CgrJsonCfg) (err error) {
	var jsnRPCConns map[string]*RPCConnsJson
	if jsnRPCConns, err = jsnCfg.RPCConnJsonCfg(); err != nil {
		return
	}
	// hardoded the *internal connection
	cfg.rpcConns[utils.MetaInternal] = &RPCConn{
		Strategy: rpcclient.PoolFirst,
		PoolSize: 0,
		Conns: []*RemoteHost{{
			Address: utils.MetaInternal,
		}},
	}
	cfg.rpcConns[rpcclient.BiRPCInternal] = &RPCConn{
		Strategy: rpcclient.PoolFirst,
		PoolSize: 0,
		Conns: []*RemoteHost{{
			Address: rpcclient.BiRPCInternal,
		}},
	}
	for key, val := range jsnRPCConns {
		cfg.rpcConns[key] = NewDfltRPCConn()
		cfg.rpcConns[key].loadFromJSONCfg(val)
	}
	return
}

// loadGeneralCfg loads the General section of the configuration
func (cfg *CGRConfig) loadGeneralCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnGeneralCfg *GeneralJsonCfg
	if jsnGeneralCfg, err = jsnCfg.GeneralJsonCfg(); err != nil {
		return
	}
	return cfg.generalCfg.loadFromJSONCfg(jsnGeneralCfg)
}

// loadCacheCfg loads the Cache section of the configuration
func (cfg *CGRConfig) loadCacheCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnCacheCfg *CacheJsonCfg
	if jsnCacheCfg, err = jsnCfg.CacheJsonCfg(); err != nil {
		return
	}
	return cfg.cacheCfg.loadFromJSONCfg(jsnCacheCfg)
}

// loadListenCfg loads the Listen section of the configuration
func (cfg *CGRConfig) loadListenCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnListenCfg *ListenJsonCfg
	if jsnListenCfg, err = jsnCfg.ListenJsonCfg(); err != nil {
		return
	}
	return cfg.listenCfg.loadFromJSONCfg(jsnListenCfg)
}

// loadHTTPCfg loads the Http section of the configuration
func (cfg *CGRConfig) loadHTTPCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnHTTPCfg *HTTPJsonCfg
	if jsnHTTPCfg, err = jsnCfg.HttpJsonCfg(); err != nil {
		return
	}
	return cfg.httpCfg.loadFromJSONCfg(jsnHTTPCfg)
}

// loadDataDBCfg loads the DataDB section of the configuration
func (cfg *CGRConfig) loadDataDBCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnDataDbCfg *DbJsonCfg
	if jsnDataDbCfg, err = jsnCfg.DbJsonCfg(DATADB_JSN); err != nil {
		return
	}
	if err = cfg.dataDbCfg.loadFromJSONCfg(jsnDataDbCfg); err != nil {
		return
	}
	return

}

// loadStorDBCfg loads the StorDB section of the configuration
func (cfg *CGRConfig) loadStorDBCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnDataDbCfg *DbJsonCfg
	if jsnDataDbCfg, err = jsnCfg.DbJsonCfg(STORDB_JSN); err != nil {
		return
	}
	return cfg.storDbCfg.loadFromJSONCfg(jsnDataDbCfg)
}

// loadFilterSCfg loads the FilterS section of the configuration
func (cfg *CGRConfig) loadFilterSCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnFilterSCfg *FilterSJsonCfg
	if jsnFilterSCfg, err = jsnCfg.FilterSJsonCfg(); err != nil {
		return
	}
	return cfg.filterSCfg.loadFromJSONCfg(jsnFilterSCfg)
}

// loadRalSCfg loads the RalS section of the configuration
func (cfg *CGRConfig) loadRalSCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnRALsCfg *RalsJsonCfg
	if jsnRALsCfg, err = jsnCfg.RalsJsonCfg(); err != nil {
		return
	}
	return cfg.ralsCfg.loadFromJSONCfg(jsnRALsCfg)
}

// loadSchedulerCfg loads the Scheduler section of the configuration
func (cfg *CGRConfig) loadSchedulerCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnSchedCfg *SchedulerJsonCfg
	if jsnSchedCfg, err = jsnCfg.SchedulerJsonCfg(); err != nil {
		return
	}
	return cfg.schedulerCfg.loadFromJSONCfg(jsnSchedCfg)
}

// loadCdrsCfg loads the Cdrs section of the configuration
func (cfg *CGRConfig) loadCdrsCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnCdrsCfg *CdrsJsonCfg
	if jsnCdrsCfg, err = jsnCfg.CdrsJsonCfg(); err != nil {
		return
	}
	return cfg.cdrsCfg.loadFromJSONCfg(jsnCdrsCfg)
}

// loadSessionSCfg loads the SessionS section of the configuration
func (cfg *CGRConfig) loadSessionSCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnSessionSCfg *SessionSJsonCfg
	if jsnSessionSCfg, err = jsnCfg.SessionSJsonCfg(); err != nil {
		return
	}
	return cfg.sessionSCfg.loadFromJSONCfg(jsnSessionSCfg)
}

// loadFreeswitchAgentCfg loads the FreeswitchAgent section of the configuration
func (cfg *CGRConfig) loadFreeswitchAgentCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnSmFsCfg *FreeswitchAgentJsonCfg
	if jsnSmFsCfg, err = jsnCfg.FreeswitchAgentJsonCfg(); err != nil {
		return
	}
	return cfg.fsAgentCfg.loadFromJSONCfg(jsnSmFsCfg)
}

// loadKamAgentCfg loads the KamAgent section of the configuration
func (cfg *CGRConfig) loadKamAgentCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnKamAgentCfg *KamAgentJsonCfg
	if jsnKamAgentCfg, err = jsnCfg.KamAgentJsonCfg(); err != nil {
		return
	}
	return cfg.kamAgentCfg.loadFromJSONCfg(jsnKamAgentCfg)
}

// loadAsteriskAgentCfg loads the AsteriskAgent section of the configuration
func (cfg *CGRConfig) loadAsteriskAgentCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnSMAstCfg *AsteriskAgentJsonCfg
	if jsnSMAstCfg, err = jsnCfg.AsteriskAgentJsonCfg(); err != nil {
		return
	}
	return cfg.asteriskAgentCfg.loadFromJSONCfg(jsnSMAstCfg)
}

// loadDiameterAgentCfg loads the DiameterAgent section of the configuration
func (cfg *CGRConfig) loadDiameterAgentCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnDACfg *DiameterAgentJsonCfg
	if jsnDACfg, err = jsnCfg.DiameterAgentJsonCfg(); err != nil {
		return
	}
	return cfg.diameterAgentCfg.loadFromJSONCfg(jsnDACfg, cfg.generalCfg.RSRSep)
}

// loadRadiusAgentCfg loads the RadiusAgent section of the configuration
func (cfg *CGRConfig) loadRadiusAgentCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnRACfg *RadiusAgentJsonCfg
	if jsnRACfg, err = jsnCfg.RadiusAgentJsonCfg(); err != nil {
		return
	}
	return cfg.radiusAgentCfg.loadFromJSONCfg(jsnRACfg, cfg.generalCfg.RSRSep)
}

// loadDNSAgentCfg loads the DNSAgent section of the configuration
func (cfg *CGRConfig) loadDNSAgentCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnDNSCfg *DNSAgentJsonCfg
	if jsnDNSCfg, err = jsnCfg.DNSAgentJsonCfg(); err != nil {
		return
	}
	return cfg.dnsAgentCfg.loadFromJSONCfg(jsnDNSCfg, cfg.generalCfg.RSRSep)
}

// loadHTTPAgentCfg loads the HttpAgent section of the configuration
func (cfg *CGRConfig) loadHTTPAgentCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnHTTPAgntCfg *[]*HttpAgentJsonCfg
	if jsnHTTPAgntCfg, err = jsnCfg.HttpAgentJsonCfg(); err != nil {
		return
	}
	return cfg.httpAgentCfg.loadFromJSONCfg(jsnHTTPAgntCfg, cfg.generalCfg.RSRSep)
}

// loadAttributeSCfg loads the AttributeS section of the configuration
func (cfg *CGRConfig) loadAttributeSCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnAttributeSCfg *AttributeSJsonCfg
	if jsnAttributeSCfg, err = jsnCfg.AttributeServJsonCfg(); err != nil {
		return
	}
	return cfg.attributeSCfg.loadFromJSONCfg(jsnAttributeSCfg)
}

// loadChargerSCfg loads the ChargerS section of the configuration
func (cfg *CGRConfig) loadChargerSCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnChargerSCfg *ChargerSJsonCfg
	if jsnChargerSCfg, err = jsnCfg.ChargerServJsonCfg(); err != nil {
		return
	}
	return cfg.chargerSCfg.loadFromJSONCfg(jsnChargerSCfg)
}

// loadResourceSCfg loads the ResourceS section of the configuration
func (cfg *CGRConfig) loadResourceSCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnRLSCfg *ResourceSJsonCfg
	if jsnRLSCfg, err = jsnCfg.ResourceSJsonCfg(); err != nil {
		return
	}
	return cfg.resourceSCfg.loadFromJSONCfg(jsnRLSCfg)
}

// loadStatSCfg loads the StatS section of the configuration
func (cfg *CGRConfig) loadStatSCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnStatSCfg *StatServJsonCfg
	if jsnStatSCfg, err = jsnCfg.StatSJsonCfg(); err != nil {
		return
	}
	return cfg.statsCfg.loadFromJSONCfg(jsnStatSCfg)
}

// loadThresholdSCfg loads the ThresholdS section of the configuration
func (cfg *CGRConfig) loadThresholdSCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnThresholdSCfg *ThresholdSJsonCfg
	if jsnThresholdSCfg, err = jsnCfg.ThresholdSJsonCfg(); err != nil {
		return
	}
	return cfg.thresholdSCfg.loadFromJSONCfg(jsnThresholdSCfg)
}

// loadRouteSCfg loads the RouteS section of the configuration
func (cfg *CGRConfig) loadRouteSCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnRouteSCfg *RouteSJsonCfg
	if jsnRouteSCfg, err = jsnCfg.RouteSJsonCfg(); err != nil {
		return
	}
	return cfg.routeSCfg.loadFromJSONCfg(jsnRouteSCfg)
}

// loadLoaderSCfg loads the LoaderS section of the configuration
func (cfg *CGRConfig) loadLoaderSCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnLoaderCfg []*LoaderJsonCfg
	if jsnLoaderCfg, err = jsnCfg.LoaderJsonCfg(); err != nil {
		return
	}
	if jsnLoaderCfg != nil {
		// cfg.loaderCfg = make(LoaderSCfgs, len(jsnLoaderCfg))
		for _, profile := range jsnLoaderCfg {
			loadSCfgp := NewDfltLoaderSCfg()
			if err = loadSCfgp.loadFromJSONCfg(profile, cfg.templates, cfg.generalCfg.RSRSep); err != nil {
				return
			}
			cfg.loaderCfg = append(cfg.loaderCfg, loadSCfgp) // use apend so the loaderS profile to be loaded from multiple files
		}
	}
	return
}

// loadMailerCfg loads the Mailer section of the configuration
func (cfg *CGRConfig) loadMailerCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnMailerCfg *MailerJsonCfg
	if jsnMailerCfg, err = jsnCfg.MailerJsonCfg(); err != nil {
		return
	}
	return cfg.mailerCfg.loadFromJSONCfg(jsnMailerCfg)
}

// loadSureTaxCfg loads the SureTax section of the configuration
func (cfg *CGRConfig) loadSureTaxCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnSureTaxCfg *SureTaxJsonCfg
	if jsnSureTaxCfg, err = jsnCfg.SureTaxJsonCfg(); err != nil {
		return
	}
	return cfg.sureTaxCfg.loadFromJSONCfg(jsnSureTaxCfg)
}

// loadDispatcherSCfg loads the DispatcherS section of the configuration
func (cfg *CGRConfig) loadDispatcherSCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnDispatcherSCfg *DispatcherSJsonCfg
	if jsnDispatcherSCfg, err = jsnCfg.DispatcherSJsonCfg(); err != nil {
		return
	}
	return cfg.dispatcherSCfg.loadFromJSONCfg(jsnDispatcherSCfg)
}

// loadDispatcherHCfg loads the DispatcherH section of the configuration
func (cfg *CGRConfig) loadDispatcherHCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnDispatcherHCfg *DispatcherHJsonCfg
	if jsnDispatcherHCfg, err = jsnCfg.DispatcherHJsonCfg(); err != nil {
		return
	}
	return cfg.dispatcherHCfg.loadFromJSONCfg(jsnDispatcherHCfg)
}

// loadLoaderCgrCfg loads the Loader section of the configuration
func (cfg *CGRConfig) loadLoaderCgrCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnLoaderCgrCfg *LoaderCfgJson
	if jsnLoaderCgrCfg, err = jsnCfg.LoaderCfgJson(); err != nil {
		return
	}
	return cfg.loaderCgrCfg.loadFromJSONCfg(jsnLoaderCgrCfg)
}

// loadMigratorCgrCfg loads the Migrator section of the configuration
func (cfg *CGRConfig) loadMigratorCgrCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnMigratorCgrCfg *MigratorCfgJson
	if jsnMigratorCgrCfg, err = jsnCfg.MigratorCfgJson(); err != nil {
		return
	}
	return cfg.migratorCgrCfg.loadFromJSONCfg(jsnMigratorCgrCfg)
}

// loadTLSCgrCfg loads the Tls section of the configuration
func (cfg *CGRConfig) loadTLSCgrCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnTLSCgrCfg *TlsJsonCfg
	if jsnTLSCgrCfg, err = jsnCfg.TlsCfgJson(); err != nil {
		return
	}
	return cfg.tlsCfg.loadFromJSONCfg(jsnTLSCgrCfg)
}

// loadAnalyzerCgrCfg loads the Analyzer section of the configuration
func (cfg *CGRConfig) loadAnalyzerCgrCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnAnalyzerCgrCfg *AnalyzerSJsonCfg
	if jsnAnalyzerCgrCfg, err = jsnCfg.AnalyzerCfgJson(); err != nil {
		return
	}
	return cfg.analyzerSCfg.loadFromJSONCfg(jsnAnalyzerCgrCfg)
}

// loadAPIBanCgrCfg loads the Analyzer section of the configuration
func (cfg *CGRConfig) loadAPIBanCgrCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnAPIBanCfg *APIBanJsonCfg
	if jsnAPIBanCfg, err = jsnCfg.ApiBanCfgJson(); err != nil {
		return
	}
	return cfg.apiBanCfg.loadFromJSONCfg(jsnAPIBanCfg)
}

// loadApierCfg loads the Apier section of the configuration
func (cfg *CGRConfig) loadApierCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnApierCfg *ApierJsonCfg
	if jsnApierCfg, err = jsnCfg.ApierCfgJson(); err != nil {
		return
	}
	return cfg.apier.loadFromJSONCfg(jsnApierCfg)
}

// loadCoreSCfg loads the CoreS section of the configuration
func (cfg *CGRConfig) loadCoreSCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnCoreCfg *CoreSJsonCfg
	if jsnCoreCfg, err = jsnCfg.CoreSCfgJson(); err != nil {
		return
	}
	return cfg.coreSCfg.loadFromJSONCfg(jsnCoreCfg)
}

// loadErsCfg loads the Ers section of the configuration
func (cfg *CGRConfig) loadErsCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnERsCfg *ERsJsonCfg
	if jsnERsCfg, err = jsnCfg.ERsJsonCfg(); err != nil {
		return
	}
	return cfg.ersCfg.loadFromJSONCfg(jsnERsCfg, cfg.templates, cfg.generalCfg.RSRSep, cfg.dfltEvRdr, cfg.generalCfg.RSRSep)
}

// loadEesCfg loads the Ees section of the configuration
func (cfg *CGRConfig) loadEesCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnEEsCfg *EEsJsonCfg
	if jsnEEsCfg, err = jsnCfg.EEsJsonCfg(); err != nil {
		return
	}
	return cfg.eesCfg.loadFromJSONCfg(jsnEEsCfg, cfg.templates, cfg.generalCfg.RSRSep, cfg.dfltEvExp)
}

// loadRateSCfg loads the rates section of the configuration
func (cfg *CGRConfig) loadRateSCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnRateCfg *RateSJsonCfg
	if jsnRateCfg, err = jsnCfg.RateCfgJson(); err != nil {
		return
	}
	return cfg.rateSCfg.loadFromJSONCfg(jsnRateCfg)
}

// loadSIPAgentCfg loads the sip_agent section of the configuration
func (cfg *CGRConfig) loadSIPAgentCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnSIPAgentCfg *SIPAgentJsonCfg
	if jsnSIPAgentCfg, err = jsnCfg.SIPAgentJsonCfg(); err != nil {
		return
	}
	return cfg.sipAgentCfg.loadFromJSONCfg(jsnSIPAgentCfg, cfg.generalCfg.RSRSep)
}

// loadTemplateSCfg loads the Template section of the configuration
func (cfg *CGRConfig) loadTemplateSCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnTemplateCfg map[string][]*FcTemplateJsonCfg
	if jsnTemplateCfg, err = jsnCfg.TemplateSJsonCfg(); err != nil {
		return
	}
	if jsnTemplateCfg != nil {
		for k, val := range jsnTemplateCfg {
			if cfg.templates[k], err = FCTemplatesFromFCTemplatesJSONCfg(val, cfg.generalCfg.RSRSep); err != nil {
				return
			}
		}
	}
	return
}

func (cfg *CGRConfig) loadConfigSCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnConfigSCfg *ConfigSCfgJson
	if jsnConfigSCfg, err = jsnCfg.ConfigSJsonCfg(); err != nil {
		return
	}
	return cfg.configSCfg.loadFromJSONCfg(jsnConfigSCfg)
}

// loadActionSCfg loads the ActionS section of the configuration
func (cfg *CGRConfig) loadActionSCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnActionCfg *ActionSJsonCfg
	if jsnActionCfg, err = jsnCfg.ActionSCfgJson(); err != nil {
		return
	}
	return cfg.actionSCfg.loadFromJSONCfg(jsnActionCfg)
}

// loadAccountSCfg loads the AccountS section of the configuration
func (cfg *CGRConfig) loadAccountSCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnActionCfg *AccountSJsonCfg
	if jsnActionCfg, err = jsnCfg.AccountSCfgJson(); err != nil {
		return
	}
	return cfg.accountSCfg.loadFromJSONCfg(jsnActionCfg)
}

// SureTaxCfg use locking to retrieve the configuration, possibility later for runtime reload
func (cfg *CGRConfig) SureTaxCfg() *SureTaxCfg {
	cfg.lks[SURETAX_JSON].Lock()
	defer cfg.lks[SURETAX_JSON].Unlock()
	return cfg.sureTaxCfg
}

// DiameterAgentCfg returns the config for Diameter Agent
func (cfg *CGRConfig) DiameterAgentCfg() *DiameterAgentCfg {
	cfg.lks[DA_JSN].Lock()
	defer cfg.lks[DA_JSN].Unlock()
	return cfg.diameterAgentCfg
}

// RadiusAgentCfg returns the config for Radius Agent
func (cfg *CGRConfig) RadiusAgentCfg() *RadiusAgentCfg {
	cfg.lks[RA_JSN].Lock()
	defer cfg.lks[RA_JSN].Unlock()
	return cfg.radiusAgentCfg
}

// DNSAgentCfg returns the config for DNS Agent
func (cfg *CGRConfig) DNSAgentCfg() *DNSAgentCfg {
	cfg.lks[DNSAgentJson].Lock()
	defer cfg.lks[DNSAgentJson].Unlock()
	return cfg.dnsAgentCfg
}

// AttributeSCfg returns the config for AttributeS
func (cfg *CGRConfig) AttributeSCfg() *AttributeSCfg {
	cfg.lks[ATTRIBUTE_JSN].Lock()
	defer cfg.lks[ATTRIBUTE_JSN].Unlock()
	return cfg.attributeSCfg
}

// ChargerSCfg returns the config for ChargerS
func (cfg *CGRConfig) ChargerSCfg() *ChargerSCfg {
	cfg.lks[ChargerSCfgJson].Lock()
	defer cfg.lks[ChargerSCfgJson].Unlock()
	return cfg.chargerSCfg
}

// ResourceSCfg returns the config for ResourceS
func (cfg *CGRConfig) ResourceSCfg() *ResourceSConfig { // not done
	cfg.lks[RESOURCES_JSON].Lock()
	defer cfg.lks[RESOURCES_JSON].Unlock()
	return cfg.resourceSCfg
}

// StatSCfg returns the config for StatS
func (cfg *CGRConfig) StatSCfg() *StatSCfg { // not done
	cfg.lks[STATS_JSON].Lock()
	defer cfg.lks[STATS_JSON].Unlock()
	return cfg.statsCfg
}

// ThresholdSCfg returns the config for ThresholdS
func (cfg *CGRConfig) ThresholdSCfg() *ThresholdSCfg {
	cfg.lks[THRESHOLDS_JSON].Lock()
	defer cfg.lks[THRESHOLDS_JSON].Unlock()
	return cfg.thresholdSCfg
}

// RouteSCfg returns the config for RouteS
func (cfg *CGRConfig) RouteSCfg() *RouteSCfg {
	cfg.lks[RouteSJson].Lock()
	defer cfg.lks[RouteSJson].Unlock()
	return cfg.routeSCfg
}

// SessionSCfg returns the config for SessionS
func (cfg *CGRConfig) SessionSCfg() *SessionSCfg {
	cfg.lks[SessionSJson].Lock()
	defer cfg.lks[SessionSJson].Unlock()
	return cfg.sessionSCfg
}

// FsAgentCfg returns the config for FsAgent
func (cfg *CGRConfig) FsAgentCfg() *FsAgentCfg {
	cfg.lks[FreeSWITCHAgentJSN].Lock()
	defer cfg.lks[FreeSWITCHAgentJSN].Unlock()
	return cfg.fsAgentCfg
}

// KamAgentCfg returns the config for KamAgent
func (cfg *CGRConfig) KamAgentCfg() *KamAgentCfg {
	cfg.lks[KamailioAgentJSN].Lock()
	defer cfg.lks[KamailioAgentJSN].Unlock()
	return cfg.kamAgentCfg
}

// AsteriskAgentCfg returns the config for AsteriskAgent
func (cfg *CGRConfig) AsteriskAgentCfg() *AsteriskAgentCfg {
	cfg.lks[AsteriskAgentJSN].Lock()
	defer cfg.lks[AsteriskAgentJSN].Unlock()
	return cfg.asteriskAgentCfg
}

// HTTPAgentCfg returns the config for HttpAgent
func (cfg *CGRConfig) HTTPAgentCfg() HTTPAgentCfgs {
	cfg.lks[HttpAgentJson].Lock()
	defer cfg.lks[HttpAgentJson].Unlock()
	return cfg.httpAgentCfg
}

// FilterSCfg returns the config for FilterS
func (cfg *CGRConfig) FilterSCfg() *FilterSCfg {
	cfg.lks[FilterSjsn].Lock()
	defer cfg.lks[FilterSjsn].Unlock()
	return cfg.filterSCfg
}

// CacheCfg returns the config for Cache
func (cfg *CGRConfig) CacheCfg() *CacheCfg {
	cfg.lks[CACHE_JSN].Lock()
	defer cfg.lks[CACHE_JSN].Unlock()
	return cfg.cacheCfg
}

// LoaderCfg returns the Loader Service
func (cfg *CGRConfig) LoaderCfg() LoaderSCfgs {
	cfg.lks[LoaderJson].Lock()
	defer cfg.lks[LoaderJson].Unlock()
	return cfg.loaderCfg
}

// LoaderCgrCfg returns the config for cgr-loader
func (cfg *CGRConfig) LoaderCgrCfg() *LoaderCgrCfg {
	cfg.lks[CgrLoaderCfgJson].Lock()
	defer cfg.lks[CgrLoaderCfgJson].Unlock()
	return cfg.loaderCgrCfg
}

// DispatcherSCfg returns the config for DispatcherS
func (cfg *CGRConfig) DispatcherSCfg() *DispatcherSCfg {
	cfg.lks[DispatcherSJson].Lock()
	defer cfg.lks[DispatcherSJson].Unlock()
	return cfg.dispatcherSCfg
}

// DispatcherHCfg returns the config for DispatcherH
func (cfg *CGRConfig) DispatcherHCfg() *DispatcherHCfg {
	cfg.lks[DispatcherSJson].Lock()
	defer cfg.lks[DispatcherSJson].Unlock()
	return cfg.dispatcherHCfg
}

// MigratorCgrCfg returns the config for Migrator
func (cfg *CGRConfig) MigratorCgrCfg() *MigratorCgrCfg {
	cfg.lks[CgrMigratorCfgJson].Lock()
	defer cfg.lks[CgrMigratorCfgJson].Unlock()
	return cfg.migratorCgrCfg
}

// SchedulerCfg returns the config for Scheduler
func (cfg *CGRConfig) SchedulerCfg() *SchedulerCfg {
	cfg.lks[SCHEDULER_JSN].Lock()
	defer cfg.lks[SCHEDULER_JSN].Unlock()
	return cfg.schedulerCfg
}

// DataDbCfg returns the config for DataDb
func (cfg *CGRConfig) DataDbCfg() *DataDbCfg {
	cfg.lks[DATADB_JSN].Lock()
	defer cfg.lks[DATADB_JSN].Unlock()
	return cfg.dataDbCfg
}

// StorDbCfg returns the config for StorDb
func (cfg *CGRConfig) StorDbCfg() *StorDbCfg {
	cfg.lks[STORDB_JSN].Lock()
	defer cfg.lks[STORDB_JSN].Unlock()
	return cfg.storDbCfg
}

// GeneralCfg returns the General config section
func (cfg *CGRConfig) GeneralCfg() *GeneralCfg {
	cfg.lks[GENERAL_JSN].Lock()
	defer cfg.lks[GENERAL_JSN].Unlock()
	return cfg.generalCfg
}

// TLSCfg returns the config for Tls
func (cfg *CGRConfig) TLSCfg() *TLSCfg {
	cfg.lks[TlsCfgJson].Lock()
	defer cfg.lks[TlsCfgJson].Unlock()
	return cfg.tlsCfg
}

// ListenCfg returns the server Listen config
func (cfg *CGRConfig) ListenCfg() *ListenCfg {
	cfg.lks[LISTEN_JSN].Lock()
	defer cfg.lks[LISTEN_JSN].Unlock()
	return cfg.listenCfg
}

// HTTPCfg returns the config for HTTP
func (cfg *CGRConfig) HTTPCfg() *HTTPCfg {
	cfg.lks[HTTP_JSN].Lock()
	defer cfg.lks[HTTP_JSN].Unlock()
	return cfg.httpCfg
}

// RalsCfg returns the config for Ral Service
func (cfg *CGRConfig) RalsCfg() *RalsCfg {
	cfg.lks[RALS_JSN].Lock()
	defer cfg.lks[RALS_JSN].Unlock()
	return cfg.ralsCfg
}

// CdrsCfg returns the config for CDR Server
func (cfg *CGRConfig) CdrsCfg() *CdrsCfg {
	cfg.lks[CDRS_JSN].Lock()
	defer cfg.lks[CDRS_JSN].Unlock()
	return cfg.cdrsCfg
}

// MailerCfg returns the config for Mailer
func (cfg *CGRConfig) MailerCfg() *MailerCfg {
	cfg.lks[MAILER_JSN].Lock()
	defer cfg.lks[MAILER_JSN].Unlock()
	return cfg.mailerCfg
}

// AnalyzerSCfg returns the config for AnalyzerS
func (cfg *CGRConfig) AnalyzerSCfg() *AnalyzerSCfg {
	cfg.lks[AnalyzerCfgJson].Lock()
	defer cfg.lks[AnalyzerCfgJson].Unlock()
	return cfg.analyzerSCfg
}

// ApierCfg reads the Apier configuration
func (cfg *CGRConfig) ApierCfg() *ApierCfg {
	cfg.lks[ApierS].Lock()
	defer cfg.lks[ApierS].Unlock()
	return cfg.apier
}

// ERsCfg reads the EventReader configuration
func (cfg *CGRConfig) ERsCfg() *ERsCfg {
	cfg.lks[ERsJson].RLock()
	defer cfg.lks[ERsJson].RUnlock()
	return cfg.ersCfg
}

// EEsCfg reads the EventExporter configuration
func (cfg *CGRConfig) EEsCfg() *EEsCfg {
	cfg.lks[EEsJson].RLock()
	defer cfg.lks[EEsJson].RUnlock()
	return cfg.eesCfg
}

// EEsNoLksCfg reads the EventExporter configuration without locks
func (cfg *CGRConfig) EEsNoLksCfg() *EEsCfg {
	return cfg.eesCfg
}

// RateSCfg reads the RateS configuration
func (cfg *CGRConfig) RateSCfg() *RateSCfg {
	cfg.lks[RateSJson].RLock()
	defer cfg.lks[RateSJson].RUnlock()
	return cfg.rateSCfg
}

// ActionSCfg reads the ActionS configuration
func (cfg *CGRConfig) ActionSCfg() *ActionSCfg {
	cfg.lks[ActionSJson].RLock()
	defer cfg.lks[ActionSJson].RUnlock()
	return cfg.actionSCfg
}

// AccountSCfg reads the AccountS configuration
func (cfg *CGRConfig) AccountSCfg() *AccountSCfg {
	cfg.lks[AccountSCfgJson].RLock()
	defer cfg.lks[AccountSCfgJson].RUnlock()
	return cfg.accountSCfg
}

// SIPAgentCfg reads the Apier configuration
func (cfg *CGRConfig) SIPAgentCfg() *SIPAgentCfg {
	cfg.lks[SIPAgentJson].Lock()
	defer cfg.lks[SIPAgentJson].Unlock()
	return cfg.sipAgentCfg
}

// RPCConns reads the RPCConns configuration
func (cfg *CGRConfig) RPCConns() RPCConns {
	cfg.lks[RPCConnsJsonName].RLock()
	defer cfg.lks[RPCConnsJsonName].RUnlock()
	return cfg.rpcConns
}

// TemplatesCfg returns the config for templates
func (cfg *CGRConfig) TemplatesCfg() FcTemplates {
	cfg.lks[TemplatesJson].Lock()
	defer cfg.lks[TemplatesJson].Unlock()
	return cfg.templates
}

// ConfigSCfg returns the configs configuration
func (cfg *CGRConfig) ConfigSCfg() *ConfigSCfg {
	cfg.lks[ConfigSJson].RLock()
	defer cfg.lks[ConfigSJson].RUnlock()
	return cfg.configSCfg
}

// APIBanCfg reads the ApiBan configuration
func (cfg *CGRConfig) APIBanCfg() *APIBanCfg {
	cfg.lks[APIBanCfgJson].Lock()
	defer cfg.lks[APIBanCfgJson].Unlock()
	return cfg.apiBanCfg
}

// CoreSCfg reads the CoreS configuration
func (cfg *CGRConfig) CoreSCfg() *CoreSCfg {
	cfg.lks[CoreSCfgJson].Lock()
	defer cfg.lks[CoreSCfgJson].Unlock()
	return cfg.coreSCfg
}

// GetReloadChan returns the reload chanel for the given section
func (cfg *CGRConfig) GetReloadChan(sectID string) chan struct{} {
	return cfg.rldChans[sectID]
}

// Call implements rpcclient.ClientConnector interface for internal RPC
func (cfg *CGRConfig) Call(serviceMethod string,
	args interface{}, reply interface{}) error {
	return utils.APIerRPCCall(cfg, serviceMethod, args, reply)
}

func (cfg *CGRConfig) rLockSections() {
	for _, lk := range cfg.lks {
		lk.RLock()
	}
}

func (cfg *CGRConfig) rUnlockSections() {
	for _, lk := range cfg.lks {
		lk.RUnlock()
	}
}

func (cfg *CGRConfig) lockSections() {
	for _, lk := range cfg.lks {
		lk.Lock()
	}
}

func (cfg *CGRConfig) unlockSections() {
	for _, lk := range cfg.lks {
		lk.Unlock()
	}
}

// RLocks will read-lock locks with IDs.
// User needs to know what he is doing since this can panic
func (cfg *CGRConfig) RLocks(lkIDs ...string) {
	for _, lkID := range lkIDs {
		cfg.lks[lkID].RLock()
	}
}

// RUnlocks will read-unlock locks with IDs.
// User needs to know what he is doing since this can panic
func (cfg *CGRConfig) RUnlocks(lkIDs ...string) {
	for _, lkID := range lkIDs {
		cfg.lks[lkID].RUnlock()
	}
}

func (cfg *CGRConfig) getLoadFunctions() map[string]func(*CgrJsonCfg) error {
	return map[string]func(*CgrJsonCfg) error{
		GENERAL_JSN:        cfg.loadGeneralCfg,
		DATADB_JSN:         cfg.loadDataDBCfg,
		STORDB_JSN:         cfg.loadStorDBCfg,
		LISTEN_JSN:         cfg.loadListenCfg,
		TlsCfgJson:         cfg.loadTLSCgrCfg,
		HTTP_JSN:           cfg.loadHTTPCfg,
		SCHEDULER_JSN:      cfg.loadSchedulerCfg,
		CACHE_JSN:          cfg.loadCacheCfg,
		FilterSjsn:         cfg.loadFilterSCfg,
		RALS_JSN:           cfg.loadRalSCfg,
		CDRS_JSN:           cfg.loadCdrsCfg,
		ERsJson:            cfg.loadErsCfg,
		EEsJson:            cfg.loadEesCfg,
		SessionSJson:       cfg.loadSessionSCfg,
		AsteriskAgentJSN:   cfg.loadAsteriskAgentCfg,
		FreeSWITCHAgentJSN: cfg.loadFreeswitchAgentCfg,
		KamailioAgentJSN:   cfg.loadKamAgentCfg,
		DA_JSN:             cfg.loadDiameterAgentCfg,
		RA_JSN:             cfg.loadRadiusAgentCfg,
		HttpAgentJson:      cfg.loadHTTPAgentCfg,
		DNSAgentJson:       cfg.loadDNSAgentCfg,
		ATTRIBUTE_JSN:      cfg.loadAttributeSCfg,
		ChargerSCfgJson:    cfg.loadChargerSCfg,
		RESOURCES_JSON:     cfg.loadResourceSCfg,
		STATS_JSON:         cfg.loadStatSCfg,
		THRESHOLDS_JSON:    cfg.loadThresholdSCfg,
		RouteSJson:         cfg.loadRouteSCfg,
		LoaderJson:         cfg.loadLoaderSCfg,
		MAILER_JSN:         cfg.loadMailerCfg,
		SURETAX_JSON:       cfg.loadSureTaxCfg,
		CgrLoaderCfgJson:   cfg.loadLoaderCgrCfg,
		CgrMigratorCfgJson: cfg.loadMigratorCgrCfg,
		DispatcherSJson:    cfg.loadDispatcherSCfg,
		DispatcherHJson:    cfg.loadDispatcherHCfg,
		AnalyzerCfgJson:    cfg.loadAnalyzerCgrCfg,
		ApierS:             cfg.loadApierCfg,
		RPCConnsJsonName:   cfg.loadRPCConns,
		RateSJson:          cfg.loadRateSCfg,
		SIPAgentJson:       cfg.loadSIPAgentCfg,
		TemplatesJson:      cfg.loadTemplateSCfg,
		ConfigSJson:        cfg.loadConfigSCfg,
		APIBanCfgJson:      cfg.loadAPIBanCgrCfg,
		CoreSCfgJson:       cfg.loadCoreSCfg,
		ActionSJson:        cfg.loadActionSCfg,
		AccountSCfgJson:    cfg.loadAccountSCfg,
	}
}

func (cfg *CGRConfig) loadCfgWithLocks(path, section string) (err error) {
	var loadFuncs []func(*CgrJsonCfg) error
	loadMap := cfg.getLoadFunctions()
	if section == utils.EmptyString || section == utils.MetaAll {
		for _, sec := range sortedCfgSections {
			cfg.lks[sec].Lock()
			defer cfg.lks[sec].Unlock()
			loadFuncs = append(loadFuncs, loadMap[sec])
		}
	} else if fnct, has := loadMap[section]; !has {
		return fmt.Errorf("Invalid section: <%s>", section)
	} else {
		cfg.lks[section].Lock()
		defer cfg.lks[section].Unlock()
		loadFuncs = append(loadFuncs, fnct)
	}
	return cfg.loadConfigFromPath(path, loadFuncs, false)
}

func (*CGRConfig) loadConfigFromReader(rdr io.Reader, loadFuncs []func(jsnCfg *CgrJsonCfg) error, envOff bool) (err error) {
	jsnCfg := new(CgrJsonCfg)
	var rjr *RjReader
	if rjr, err = NewRjReader(rdr); err != nil {
		return
	}
	rjr.envOff = envOff
	defer rjr.Close() // make sure we make the buffer nil
	if err = rjr.Decode(jsnCfg); err != nil {
		return
	}
	for _, loadFunc := range loadFuncs {
		if err = loadFunc(jsnCfg); err != nil {
			return
		}
	}
	return
}

// Reads all .json files out of a folder/subfolders and loads them up in lexical order
func (cfg *CGRConfig) loadConfigFromPath(path string, loadFuncs []func(jsnCfg *CgrJsonCfg) error, envOff bool) (err error) {
	if utils.IsURL(path) {
		return cfg.loadConfigFromHTTP(path, loadFuncs) // prefix protocol
	}
	var fi os.FileInfo
	if fi, err = os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return utils.ErrPathNotReachable(path)
		}
		return
	} else if !fi.IsDir() && path != utils.ConfigPath { // If config dir defined, needs to exist, not checking for default
		return fmt.Errorf("path: %s not a directory", path)
	}

	// safe to assume that path is a directory
	return cfg.loadConfigFromFolder(path, loadFuncs, envOff)
}

func (cfg *CGRConfig) loadConfigFromFolder(cfgDir string, loadFuncs []func(jsnCfg *CgrJsonCfg) error, envOff bool) (err error) {
	jsonFilesFound := false
	if err = filepath.Walk(cfgDir, func(path string, info os.FileInfo, err error) (werr error) {
		if !info.IsDir() || isHidden(info.Name()) { // also ignore hidden files and folders
			return
		}
		var cfgFiles []string
		if cfgFiles, werr = filepath.Glob(filepath.Join(path, "*.json")); werr != nil {
			return
		}
		if cfgFiles == nil { // No need of processing further since there are no config files in the folder
			return
		}
		if !jsonFilesFound {
			jsonFilesFound = true
		}
		for _, jsonFilePath := range cfgFiles {
			if werr = cfg.loadConfigFromFile(jsonFilePath, loadFuncs, envOff); werr != nil {
				return
			}
		}
		return
	}); err != nil {
		return
	}
	if !jsonFilesFound {
		return fmt.Errorf("No config file found on path %s", cfgDir)
	}
	return
}

// loadConfigFromFile loads the config from a file
// extracted from a loadConfigFromFolder in order to test all cases
func (cfg *CGRConfig) loadConfigFromFile(jsonFilePath string, loadFuncs []func(jsnCfg *CgrJsonCfg) error, envOff bool) (err error) {
	var cfgFile *os.File
	cfgFile, err = os.Open(jsonFilePath)
	if err != nil {
		return
	}
	err = cfg.loadConfigFromReader(cfgFile, loadFuncs, envOff)
	cfgFile.Close()
	if err != nil {
		err = fmt.Errorf("file <%s>:%s", jsonFilePath, err.Error())
	}
	return
}

func (cfg *CGRConfig) loadConfigFromHTTP(urlPaths string, loadFuncs []func(jsnCfg *CgrJsonCfg) error) (err error) {
	for _, urlPath := range strings.Split(urlPaths, utils.InfieldSep) {
		if _, err = url.ParseRequestURI(urlPath); err != nil {
			return
		}
		var myClient = &http.Client{
			Timeout: CgrConfig().GeneralCfg().ReplyTimeout,
		}
		var cfgReq *http.Response
		cfgReq, err = myClient.Get(urlPath)
		if err != nil {
			return utils.ErrPathNotReachable(urlPath)
		}
		err = cfg.loadConfigFromReader(cfgReq.Body, loadFuncs, false)
		cfgReq.Body.Close()
		if err != nil {
			err = fmt.Errorf("url <%s>:%s", urlPath, err.Error())
			return
		}
	}
	return
}

// populates the config locks and the reload channels
func (cfg *CGRConfig) initChanels() {
	cfg.lks = make(map[string]*sync.RWMutex)
	cfg.rldChans = make(map[string]chan struct{})
	for _, section := range sortedCfgSections {
		cfg.lks[section] = new(sync.RWMutex)
		cfg.rldChans[section] = make(chan struct{}, 1)
	}
}

func (cfg *CGRConfig) loadCfgFromJSONWithLocks(rdr io.Reader, sections []string) (err error) {
	var loadFuncs []func(*CgrJsonCfg) error
	loadMap := cfg.getLoadFunctions()
	for _, section := range sections {
		fnct, has := loadMap[section]
		if !has {
			return fmt.Errorf("Invalid section: <%s>", section)
		}
		cfg.lks[section].Lock()
		defer cfg.lks[section].Unlock()
		loadFuncs = append(loadFuncs, fnct)
	}
	return cfg.loadConfigFromReader(rdr, loadFuncs, false)
}

// reloadSections sends a signal to the reload channel for the needed sections
// the list of sections should be always valid because we load the config first with this list
func (cfg *CGRConfig) reloadSections(sections ...string) {
	subsystemsThatNeedDataDB := utils.NewStringSet([]string{DATADB_JSN, SCHEDULER_JSN,
		RALS_JSN, CDRS_JSN, SessionSJson, ATTRIBUTE_JSN,
		ChargerSCfgJson, RESOURCES_JSON, STATS_JSON, THRESHOLDS_JSON,
		RouteSJson, LoaderJson, DispatcherSJson, RateSJson, ApierS, AccountSCfgJson,
		ActionSJson})
	subsystemsThatNeedStorDB := utils.NewStringSet([]string{STORDB_JSN, RALS_JSN, CDRS_JSN, ApierS})
	needsDataDB := false
	needsStorDB := false
	for _, section := range sections {
		if !needsDataDB && subsystemsThatNeedDataDB.Has(section) {
			needsDataDB = true
			cfg.rldChans[DATADB_JSN] <- struct{}{} // reload datadb before
		}
		if !needsStorDB && subsystemsThatNeedStorDB.Has(section) {
			needsStorDB = true
			cfg.rldChans[STORDB_JSN] <- struct{}{} // reload stordb before
		}
		if needsDataDB && needsStorDB {
			break
		}
	}
	runtime.Gosched()
	for _, section := range sections {
		switch section {
		case ConfigSJson:
		case GENERAL_JSN: // nothing to reload
		case RPCConnsJsonName: // nothing to reload
			cfg.rldChans[RPCConnsJsonName] <- struct{}{}
		case DATADB_JSN: // reloaded before
		case STORDB_JSN: // reloaded before
		case LISTEN_JSN:
		case CACHE_JSN:
		case FilterSjsn:
		case MAILER_JSN:
		case SURETAX_JSON:
		case CgrLoaderCfgJson:
		case CgrMigratorCfgJson:
		case TemplatesJson:
		case TlsCfgJson: // nothing to reload
		case APIBanCfgJson: // nothing to reload
		case CoreSCfgJson: // nothing to reload
		case HTTP_JSN:
			cfg.rldChans[HTTP_JSN] <- struct{}{}
		case SCHEDULER_JSN:
			cfg.rldChans[SCHEDULER_JSN] <- struct{}{}
		case RALS_JSN:
			cfg.rldChans[RALS_JSN] <- struct{}{}
		case CDRS_JSN:
			cfg.rldChans[CDRS_JSN] <- struct{}{}
		case ERsJson:
			cfg.rldChans[ERsJson] <- struct{}{}
		case SessionSJson:
			cfg.rldChans[SessionSJson] <- struct{}{}
		case AsteriskAgentJSN:
			cfg.rldChans[AsteriskAgentJSN] <- struct{}{}
		case FreeSWITCHAgentJSN:
			cfg.rldChans[FreeSWITCHAgentJSN] <- struct{}{}
		case KamailioAgentJSN:
			cfg.rldChans[KamailioAgentJSN] <- struct{}{}
		case DA_JSN:
			cfg.rldChans[DA_JSN] <- struct{}{}
		case RA_JSN:
			cfg.rldChans[RA_JSN] <- struct{}{}
		case HttpAgentJson:
			cfg.rldChans[HttpAgentJson] <- struct{}{}
		case DNSAgentJson:
			cfg.rldChans[DNSAgentJson] <- struct{}{}
		case ATTRIBUTE_JSN:
			cfg.rldChans[ATTRIBUTE_JSN] <- struct{}{}
		case ChargerSCfgJson:
			cfg.rldChans[ChargerSCfgJson] <- struct{}{}
		case RESOURCES_JSON:
			cfg.rldChans[RESOURCES_JSON] <- struct{}{}
		case STATS_JSON:
			cfg.rldChans[STATS_JSON] <- struct{}{}
		case THRESHOLDS_JSON:
			cfg.rldChans[THRESHOLDS_JSON] <- struct{}{}
		case RouteSJson:
			cfg.rldChans[RouteSJson] <- struct{}{}
		case LoaderJson:
			cfg.rldChans[LoaderJson] <- struct{}{}
		case DispatcherSJson:
			cfg.rldChans[DispatcherSJson] <- struct{}{}
		case AnalyzerCfgJson:
			cfg.rldChans[AnalyzerCfgJson] <- struct{}{}
		case ApierS:
			cfg.rldChans[ApierS] <- struct{}{}
		case EEsJson:
			cfg.rldChans[EEsJson] <- struct{}{}
		case SIPAgentJson:
			cfg.rldChans[SIPAgentJson] <- struct{}{}
		case RateSJson:
			cfg.rldChans[RateSJson] <- struct{}{}
		case DispatcherHJson:
			cfg.rldChans[DispatcherHJson] <- struct{}{}
		case AccountSCfgJson:
			cfg.rldChans[AccountSCfgJson] <- struct{}{}
		case ActionSJson:
			cfg.rldChans[ActionSJson] <- struct{}{}
		}
	}
	return
}

// AsMapInterface returns the config as a map[string]interface{}
func (cfg *CGRConfig) AsMapInterface(separator string) (mp map[string]interface{}) {
	return map[string]interface{}{
		LoaderJson:         cfg.loaderCfg.AsMapInterface(separator),
		HttpAgentJson:      cfg.httpAgentCfg.AsMapInterface(separator),
		RPCConnsJsonName:   cfg.rpcConns.AsMapInterface(),
		GENERAL_JSN:        cfg.generalCfg.AsMapInterface(),
		DATADB_JSN:         cfg.dataDbCfg.AsMapInterface(),
		STORDB_JSN:         cfg.storDbCfg.AsMapInterface(),
		TlsCfgJson:         cfg.tlsCfg.AsMapInterface(),
		CACHE_JSN:          cfg.cacheCfg.AsMapInterface(),
		LISTEN_JSN:         cfg.listenCfg.AsMapInterface(),
		HTTP_JSN:           cfg.httpCfg.AsMapInterface(),
		FilterSjsn:         cfg.filterSCfg.AsMapInterface(),
		RALS_JSN:           cfg.ralsCfg.AsMapInterface(),
		SCHEDULER_JSN:      cfg.schedulerCfg.AsMapInterface(),
		CDRS_JSN:           cfg.cdrsCfg.AsMapInterface(),
		SessionSJson:       cfg.sessionSCfg.AsMapInterface(),
		FreeSWITCHAgentJSN: cfg.fsAgentCfg.AsMapInterface(separator),
		KamailioAgentJSN:   cfg.kamAgentCfg.AsMapInterface(),
		AsteriskAgentJSN:   cfg.asteriskAgentCfg.AsMapInterface(),
		DA_JSN:             cfg.diameterAgentCfg.AsMapInterface(separator),
		RA_JSN:             cfg.radiusAgentCfg.AsMapInterface(separator),
		DNSAgentJson:       cfg.dnsAgentCfg.AsMapInterface(separator),
		ATTRIBUTE_JSN:      cfg.attributeSCfg.AsMapInterface(),
		ChargerSCfgJson:    cfg.chargerSCfg.AsMapInterface(),
		RESOURCES_JSON:     cfg.resourceSCfg.AsMapInterface(),
		STATS_JSON:         cfg.statsCfg.AsMapInterface(),
		THRESHOLDS_JSON:    cfg.thresholdSCfg.AsMapInterface(),
		RouteSJson:         cfg.routeSCfg.AsMapInterface(),
		SURETAX_JSON:       cfg.sureTaxCfg.AsMapInterface(separator),
		DispatcherSJson:    cfg.dispatcherSCfg.AsMapInterface(),
		DispatcherHJson:    cfg.dispatcherHCfg.AsMapInterface(),
		CgrLoaderCfgJson:   cfg.loaderCgrCfg.AsMapInterface(),
		CgrMigratorCfgJson: cfg.migratorCgrCfg.AsMapInterface(),
		MAILER_JSN:         cfg.mailerCfg.AsMapInterface(),
		AnalyzerCfgJson:    cfg.analyzerSCfg.AsMapInterface(),
		ApierS:             cfg.apier.AsMapInterface(),
		ERsJson:            cfg.ersCfg.AsMapInterface(separator),
		APIBanCfgJson:      cfg.apiBanCfg.AsMapInterface(),
		EEsJson:            cfg.eesCfg.AsMapInterface(separator),
		RateSJson:          cfg.rateSCfg.AsMapInterface(),
		SIPAgentJson:       cfg.sipAgentCfg.AsMapInterface(separator),
		TemplatesJson:      cfg.templates.AsMapInterface(separator),
		ConfigSJson:        cfg.configSCfg.AsMapInterface(),
		CoreSCfgJson:       cfg.coreSCfg.AsMapInterface(),
		ActionSJson:        cfg.actionSCfg.AsMapInterface(),
		AccountSCfgJson:    cfg.accountSCfg.AsMapInterface(),
	}
}

// ReloadArgs the API params for V1ReloadConfig
type ReloadArgs struct {
	Opts    map[string]interface{}
	Tenant  string
	Path    string
	Section string
	DryRun  bool
}

// V1ReloadConfig reloads the configuration
func (cfg *CGRConfig) V1ReloadConfig(args *ReloadArgs, reply *string) (err error) {
	if missing := utils.MissingStructFields(args, []string{"Path"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	cfgV := cfg
	if args.DryRun {
		cfgV = cfg.Clone()
	}
	cfgV.reloadDPCache(args.Section)
	if err = cfgV.loadCfgWithLocks(args.Path, args.Section); err != nil {
		return
	}
	//  lock all sections
	cfgV.rLockSections()

	err = cfgV.checkConfigSanity()

	cfgV.rUnlockSections() // unlock before checking the error

	if err != nil {
		return
	}
	if !args.DryRun {
		if args.Section == utils.EmptyString || args.Section == utils.MetaAll {
			cfgV.reloadSections(sortedCfgSections...)
		} else {
			cfgV.reloadSections(args.Section)
		}
	}
	*reply = utils.OK
	return
}

// SectionWithOpts the API params for GetConfig
type SectionWithOpts struct {
	Opts    map[string]interface{}
	Tenant  string
	Section string
}

// V1GetConfig will retrieve from CGRConfig a section
func (cfg *CGRConfig) V1GetConfig(args *SectionWithOpts, reply *map[string]interface{}) (err error) {
	args.Section = utils.FirstNonEmpty(args.Section, utils.MetaAll)
	cfg.cacheDPMux.RLock()
	if mp, has := cfg.cacheDP[args.Section]; has && mp != nil {
		*reply = mp
		cfg.cacheDPMux.RUnlock()
		return
	}
	cfg.cacheDPMux.RUnlock()
	defer func() {
		if err != nil {
			return
		}
		cfg.cacheDPMux.Lock()
		cfg.cacheDP[args.Section] = *reply
		cfg.cacheDPMux.Unlock()
	}()
	var mp interface{}
	switch args.Section {
	case utils.MetaAll:
		*reply = cfg.AsMapInterface(cfg.GeneralCfg().RSRSep)
		return
	case GENERAL_JSN:
		mp = cfg.GeneralCfg().AsMapInterface()
	case DATADB_JSN:
		mp = cfg.DataDbCfg().AsMapInterface()
	case STORDB_JSN:
		mp = cfg.StorDbCfg().AsMapInterface()
	case TlsCfgJson:
		mp = cfg.TLSCfg().AsMapInterface()
	case CACHE_JSN:
		mp = cfg.CacheCfg().AsMapInterface()
	case LISTEN_JSN:
		mp = cfg.ListenCfg().AsMapInterface()
	case HTTP_JSN:
		mp = cfg.HTTPCfg().AsMapInterface()
	case FilterSjsn:
		mp = cfg.FilterSCfg().AsMapInterface()
	case RALS_JSN:
		mp = cfg.RalsCfg().AsMapInterface()
	case SCHEDULER_JSN:
		mp = cfg.SchedulerCfg().AsMapInterface()
	case CDRS_JSN:
		mp = cfg.CdrsCfg().AsMapInterface()
	case SessionSJson:
		mp = cfg.SessionSCfg().AsMapInterface()
	case FreeSWITCHAgentJSN:
		mp = cfg.FsAgentCfg().AsMapInterface(cfg.GeneralCfg().RSRSep)
	case KamailioAgentJSN:
		mp = cfg.KamAgentCfg().AsMapInterface()
	case AsteriskAgentJSN:
		mp = cfg.AsteriskAgentCfg().AsMapInterface()
	case DA_JSN:
		mp = cfg.DiameterAgentCfg().AsMapInterface(cfg.GeneralCfg().RSRSep)
	case RA_JSN:
		mp = cfg.RadiusAgentCfg().AsMapInterface(cfg.GeneralCfg().RSRSep)
	case DNSAgentJson:
		mp = cfg.DNSAgentCfg().AsMapInterface(cfg.GeneralCfg().RSRSep)
	case ATTRIBUTE_JSN:
		mp = cfg.AttributeSCfg().AsMapInterface()
	case ChargerSCfgJson:
		mp = cfg.ChargerSCfg().AsMapInterface()
	case RESOURCES_JSON:
		mp = cfg.ResourceSCfg().AsMapInterface()
	case STATS_JSON:
		mp = cfg.StatSCfg().AsMapInterface()
	case THRESHOLDS_JSON:
		mp = cfg.ThresholdSCfg().AsMapInterface()
	case RouteSJson:
		mp = cfg.RouteSCfg().AsMapInterface()
	case SURETAX_JSON:
		mp = cfg.SureTaxCfg().AsMapInterface(cfg.GeneralCfg().RSRSep)
	case DispatcherSJson:
		mp = cfg.DispatcherSCfg().AsMapInterface()
	case DispatcherHJson:
		mp = cfg.DispatcherHCfg().AsMapInterface()
	case LoaderJson:
		mp = cfg.LoaderCfg().AsMapInterface(cfg.GeneralCfg().RSRSep)
	case CgrLoaderCfgJson:
		mp = cfg.LoaderCgrCfg().AsMapInterface()
	case CgrMigratorCfgJson:
		mp = cfg.MigratorCgrCfg().AsMapInterface()
	case ApierS:
		mp = cfg.ApierCfg().AsMapInterface()
	case EEsJson:
		mp = cfg.EEsCfg().AsMapInterface(cfg.GeneralCfg().RSRSep)
	case ERsJson:
		mp = cfg.ERsCfg().AsMapInterface(cfg.GeneralCfg().RSRSep)
	case RPCConnsJsonName:
		mp = cfg.RPCConns().AsMapInterface()
	case SIPAgentJson:
		mp = cfg.SIPAgentCfg().AsMapInterface(cfg.GeneralCfg().RSRSep)
	case TemplatesJson:
		mp = cfg.TemplatesCfg().AsMapInterface(cfg.GeneralCfg().RSRSep)
	case ConfigSJson:
		mp = cfg.ConfigSCfg().AsMapInterface()
	case APIBanCfgJson:
		mp = cfg.APIBanCfg().AsMapInterface()
	case HttpAgentJson:
		mp = cfg.HTTPAgentCfg().AsMapInterface(cfg.GeneralCfg().RSRSep)
	case MAILER_JSN:
		mp = cfg.MailerCfg().AsMapInterface()
	case AnalyzerCfgJson:
		mp = cfg.AnalyzerSCfg().AsMapInterface()
	case RateSJson:
		mp = cfg.RateSCfg().AsMapInterface()
	case CoreSCfgJson:
		mp = cfg.CoreSCfg().AsMapInterface()
	case ActionSJson:
		mp = cfg.ActionSCfg().AsMapInterface()
	case AccountSCfgJson:
		mp = cfg.AccountSCfg().AsMapInterface()
	default:
		return errors.New("Invalid section")
	}
	*reply = map[string]interface{}{args.Section: mp}
	return
}

// SetConfigArgs the API params for V1SetConfig
type SetConfigArgs struct {
	Opts   map[string]interface{}
	Tenant string
	Config map[string]interface{}
	DryRun bool
}

// V1SetConfig reloads the sections of config
func (cfg *CGRConfig) V1SetConfig(args *SetConfigArgs, reply *string) (err error) {
	if len(args.Config) == 0 {
		*reply = utils.OK
		return
	}
	sections := make([]string, 0, len(args.Config))
	for section := range args.Config {
		sections = append(sections, section)
	}
	var b []byte
	if b, err = json.Marshal(args.Config); err != nil {
		return
	}
	cfgV := cfg
	if args.DryRun {
		cfgV = cfg.Clone()
	}

	cfgV.reloadDPCache(sections...)
	if err = cfgV.loadCfgFromJSONWithLocks(bytes.NewBuffer(b), sections); err != nil {
		return
	}

	//  lock all sections
	cfgV.rLockSections()

	err = cfgV.checkConfigSanity()

	cfgV.rUnlockSections() // unlock before checking the error
	if err != nil {
		return
	}
	if !args.DryRun {
		cfgV.reloadSections(sections...)
	}
	*reply = utils.OK
	return
}

//V1GetConfigAsJSON will retrieve from CGRConfig a section as a string
func (cfg *CGRConfig) V1GetConfigAsJSON(args *SectionWithOpts, reply *string) (err error) {
	args.Section = utils.FirstNonEmpty(args.Section, utils.MetaAll)
	cfg.cacheDPMux.RLock()
	if mp, has := cfg.cacheDP[args.Section]; has && mp != nil {
		*reply = utils.ToJSON(mp)
		cfg.cacheDPMux.RUnlock()
		return
	}
	cfg.cacheDPMux.RUnlock()
	var rplyMap utils.MapStorage
	defer func() {
		if err != nil {
			return
		}
		cfg.cacheDPMux.Lock()
		cfg.cacheDP[args.Section] = rplyMap
		cfg.cacheDPMux.Unlock()
	}()
	var mp interface{}
	switch args.Section {
	case utils.MetaAll:
		rplyMap = cfg.AsMapInterface(cfg.GeneralCfg().RSRSep)
		*reply = utils.ToJSON(rplyMap)
		return
	case GENERAL_JSN:
		mp = cfg.GeneralCfg().AsMapInterface()
	case DATADB_JSN:
		mp = cfg.DataDbCfg().AsMapInterface()
	case STORDB_JSN:
		mp = cfg.StorDbCfg().AsMapInterface()
	case TlsCfgJson:
		mp = cfg.TLSCfg().AsMapInterface()
	case CACHE_JSN:
		mp = cfg.CacheCfg().AsMapInterface()
	case LISTEN_JSN:
		mp = cfg.ListenCfg().AsMapInterface()
	case HTTP_JSN:
		mp = cfg.HTTPCfg().AsMapInterface()
	case FilterSjsn:
		mp = cfg.FilterSCfg().AsMapInterface()
	case RALS_JSN:
		mp = cfg.RalsCfg().AsMapInterface()
	case SCHEDULER_JSN:
		mp = cfg.SchedulerCfg().AsMapInterface()
	case CDRS_JSN:
		mp = cfg.CdrsCfg().AsMapInterface()
	case SessionSJson:
		mp = cfg.SessionSCfg().AsMapInterface()
	case FreeSWITCHAgentJSN:
		mp = cfg.FsAgentCfg().AsMapInterface(cfg.GeneralCfg().RSRSep)
	case KamailioAgentJSN:
		mp = cfg.KamAgentCfg().AsMapInterface()
	case AsteriskAgentJSN:
		mp = cfg.AsteriskAgentCfg().AsMapInterface()
	case DA_JSN:
		mp = cfg.DiameterAgentCfg().AsMapInterface(cfg.GeneralCfg().RSRSep)
	case RA_JSN:
		mp = cfg.RadiusAgentCfg().AsMapInterface(cfg.GeneralCfg().RSRSep)
	case DNSAgentJson:
		mp = cfg.DNSAgentCfg().AsMapInterface(cfg.GeneralCfg().RSRSep)
	case ATTRIBUTE_JSN:
		mp = cfg.AttributeSCfg().AsMapInterface()
	case ChargerSCfgJson:
		mp = cfg.ChargerSCfg().AsMapInterface()
	case RESOURCES_JSON:
		mp = cfg.ResourceSCfg().AsMapInterface()
	case STATS_JSON:
		mp = cfg.StatSCfg().AsMapInterface()
	case THRESHOLDS_JSON:
		mp = cfg.ThresholdSCfg().AsMapInterface()
	case RouteSJson:
		mp = cfg.RouteSCfg().AsMapInterface()
	case SURETAX_JSON:
		mp = cfg.SureTaxCfg().AsMapInterface(cfg.GeneralCfg().RSRSep)
	case DispatcherSJson:
		mp = cfg.DispatcherSCfg().AsMapInterface()
	case DispatcherHJson:
		mp = cfg.DispatcherHCfg().AsMapInterface()
	case LoaderJson:
		mp = cfg.LoaderCfg().AsMapInterface(cfg.GeneralCfg().RSRSep)
	case CgrLoaderCfgJson:
		mp = cfg.LoaderCgrCfg().AsMapInterface()
	case CgrMigratorCfgJson:
		mp = cfg.MigratorCgrCfg().AsMapInterface()
	case ApierS:
		mp = cfg.ApierCfg().AsMapInterface()
	case EEsJson:
		mp = cfg.EEsCfg().AsMapInterface(cfg.GeneralCfg().RSRSep)
	case ERsJson:
		mp = cfg.ERsCfg().AsMapInterface(cfg.GeneralCfg().RSRSep)
	case SIPAgentJson:
		mp = cfg.SIPAgentCfg().AsMapInterface(cfg.GeneralCfg().RSRSep)
	case ConfigSJson:
		mp = cfg.ConfigSCfg().AsMapInterface()
	case APIBanCfgJson:
		mp = cfg.APIBanCfg().AsMapInterface()
	case RPCConnsJsonName:
		mp = cfg.RPCConns().AsMapInterface()
	case TemplatesJson:
		mp = cfg.TemplatesCfg().AsMapInterface(cfg.GeneralCfg().RSRSep)
	case HttpAgentJson:
		mp = cfg.HTTPAgentCfg().AsMapInterface(cfg.GeneralCfg().RSRSep)
	case MAILER_JSN:
		mp = cfg.MailerCfg().AsMapInterface()
	case AnalyzerCfgJson:
		mp = cfg.AnalyzerSCfg().AsMapInterface()
	case RateSJson:
		mp = cfg.RateSCfg().AsMapInterface()
	case CoreSCfgJson:
		mp = cfg.CoreSCfg().AsMapInterface()
	case AccountSCfgJson:
		mp = cfg.AccountSCfg().AsMapInterface()
	default:
		return errors.New("Invalid section")
	}
	rplyMap = map[string]interface{}{args.Section: mp}
	*reply = utils.ToJSON(rplyMap)
	return
}

// SetConfigFromJSONArgs the API params for V1SetConfigFromJSON
type SetConfigFromJSONArgs struct {
	Opts   map[string]interface{}
	Tenant string
	Config string
	DryRun bool
}

// V1SetConfigFromJSON reloads the sections of config
func (cfg *CGRConfig) V1SetConfigFromJSON(args *SetConfigFromJSONArgs, reply *string) (err error) {
	if len(args.Config) == 0 {
		*reply = utils.OK
		return
	}
	cfgV := cfg
	if args.DryRun {
		cfgV = cfg.Clone()
	}

	cfgV.reloadDPCache(sortedCfgSections...)
	if err = cfgV.loadCfgFromJSONWithLocks(bytes.NewBufferString(args.Config), sortedCfgSections); err != nil {
		return
	}

	//  lock all sections
	cfgV.rLockSections()
	err = cfgV.checkConfigSanity()
	cfgV.rUnlockSections() // unlock before checking the error
	if err != nil {
		return
	}
	if !args.DryRun {
		cfgV.reloadSections(sortedCfgSections...)
	}
	*reply = utils.OK
	return
}

// Clone returns a deep copy of CGRConfig
func (cfg *CGRConfig) Clone() (cln *CGRConfig) {
	cln = &CGRConfig{
		DataFolderPath: cfg.DataFolderPath,
		ConfigPath:     cfg.ConfigPath,

		dfltEvRdr:        cfg.dfltEvRdr.Clone(),
		dfltEvExp:        cfg.dfltEvExp.Clone(),
		loaderCfg:        cfg.loaderCfg.Clone(),
		httpAgentCfg:     cfg.httpAgentCfg.Clone(),
		rpcConns:         cfg.rpcConns.Clone(),
		templates:        cfg.templates.Clone(),
		generalCfg:       cfg.generalCfg.Clone(),
		dataDbCfg:        cfg.dataDbCfg.Clone(),
		storDbCfg:        cfg.storDbCfg.Clone(),
		tlsCfg:           cfg.tlsCfg.Clone(),
		cacheCfg:         cfg.cacheCfg.Clone(),
		listenCfg:        cfg.listenCfg.Clone(),
		httpCfg:          cfg.httpCfg.Clone(),
		filterSCfg:       cfg.filterSCfg.Clone(),
		ralsCfg:          cfg.ralsCfg.Clone(),
		schedulerCfg:     cfg.schedulerCfg.Clone(),
		cdrsCfg:          cfg.cdrsCfg.Clone(),
		sessionSCfg:      cfg.sessionSCfg.Clone(),
		fsAgentCfg:       cfg.fsAgentCfg.Clone(),
		kamAgentCfg:      cfg.kamAgentCfg.Clone(),
		asteriskAgentCfg: cfg.asteriskAgentCfg.Clone(),
		diameterAgentCfg: cfg.diameterAgentCfg.Clone(),
		radiusAgentCfg:   cfg.radiusAgentCfg.Clone(),
		dnsAgentCfg:      cfg.dnsAgentCfg.Clone(),
		attributeSCfg:    cfg.attributeSCfg.Clone(),
		chargerSCfg:      cfg.chargerSCfg.Clone(),
		resourceSCfg:     cfg.resourceSCfg.Clone(),
		statsCfg:         cfg.statsCfg.Clone(),
		thresholdSCfg:    cfg.thresholdSCfg.Clone(),
		routeSCfg:        cfg.routeSCfg.Clone(),
		sureTaxCfg:       cfg.sureTaxCfg.Clone(),
		dispatcherSCfg:   cfg.dispatcherSCfg.Clone(),
		dispatcherHCfg:   cfg.dispatcherHCfg.Clone(),
		loaderCgrCfg:     cfg.loaderCgrCfg.Clone(),
		migratorCgrCfg:   cfg.migratorCgrCfg.Clone(),
		mailerCfg:        cfg.mailerCfg.Clone(),
		analyzerSCfg:     cfg.analyzerSCfg.Clone(),
		apier:            cfg.apier.Clone(),
		ersCfg:           cfg.ersCfg.Clone(),
		eesCfg:           cfg.eesCfg.Clone(),
		rateSCfg:         cfg.rateSCfg.Clone(),
		sipAgentCfg:      cfg.sipAgentCfg.Clone(),
		configSCfg:       cfg.configSCfg.Clone(),
		apiBanCfg:        cfg.apiBanCfg.Clone(),
		coreSCfg:         cfg.coreSCfg.Clone(),
		actionSCfg:       cfg.actionSCfg.Clone(),
		accountSCfg:      cfg.accountSCfg.Clone(),

		cacheDP: make(map[string]utils.MapStorage),
	}
	cln.initChanels()
	return
}

func (cfg *CGRConfig) reloadDPCache(sections ...string) {
	cfg.cacheDPMux.Lock()
	delete(cfg.cacheDP, utils.MetaAll)
	for _, sec := range sections {
		delete(cfg.cacheDP, sec)
	}
	cfg.cacheDPMux.Unlock()
}

// GetDataProvider returns the config as a data provider interface
func (cfg *CGRConfig) GetDataProvider() utils.DataProvider {
	cfg.cacheDPMux.RLock()
	val, has := cfg.cacheDP[utils.MetaAll]
	cfg.cacheDPMux.RUnlock()
	if !has || val == nil {
		cfg.cacheDPMux.Lock()
		val = utils.MapStorage(cfg.AsMapInterface(cfg.GeneralCfg().RSRSep))
		cfg.cacheDP[utils.MetaAll] = val
		cfg.cacheDPMux.Unlock()
	}
	return val
}
