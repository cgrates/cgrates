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
		utils.Internal: map[string]string{
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
	cfg.radiusAgentCfg.ClientDictionaries = make(map[string]string)
	cfg.radiusAgentCfg.ClientSecrets = make(map[string]string)
	cfg.dnsAgentCfg = new(DNSAgentCfg)
	cfg.attributeSCfg = new(AttributeSCfg)
	cfg.chargerSCfg = new(ChargerSCfg)
	cfg.resourceSCfg = new(ResourceSConfig)
	cfg.statsCfg = new(StatSCfg)
	cfg.thresholdSCfg = new(ThresholdSCfg)
	cfg.routeSCfg = new(RouteSCfg)
	cfg.sureTaxCfg = new(SureTaxCfg)
	cfg.dispatcherSCfg = new(DispatcherSCfg)
	cfg.registrarCCfg = new(RegistrarCCfgs)
	cfg.registrarCCfg.RPC = new(RegistrarCCfg)
	cfg.registrarCCfg.Dispatcher = new(RegistrarCCfg)
	cfg.registrarCCfg.RPC.Hosts = make(map[string][]*RemoteHost)
	cfg.registrarCCfg.Dispatcher.Hosts = make(map[string][]*RemoteHost)
	cfg.loaderCgrCfg = new(LoaderCgrCfg)
	cfg.migratorCgrCfg = new(MigratorCgrCfg)
	cfg.migratorCgrCfg.OutDataDBOpts = make(map[string]interface{})
	cfg.migratorCgrCfg.OutStorDBOpts = make(map[string]interface{})
	cfg.loaderCfg = make(LoaderSCfgs, 0)
	cfg.admS = new(AdminSCfg)
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
	cfg.configDBCfg = new(DataDbCfg)
	cfg.configDBCfg.Items = make(map[string]*ItemOpt)
	cfg.configDBCfg.Opts = make(map[string]interface{})

	cfg.cacheDP = make(utils.MapStorage)

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
	dfltRemoteHost = cfg.rpcConns[utils.MetaLocalHost].Conns[0].Clone()
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

	if err = cfg.loadConfigFromPath(path, []func(ConfigDB) error{cfg.loadFromJSONCfg}, false); err != nil {
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

	err = cfg.loadConfigFromPath(path, []func(ConfigDB) error{cfg.loadFromJSONCfg}, true)
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

	templates FCTemplates

	generalCfg       *GeneralCfg       // General config
	dataDbCfg        *DataDbCfg        // Database config
	storDbCfg        *StorDbCfg        // StroreDb config
	tlsCfg           *TLSCfg           // TLS config
	cacheCfg         *CacheCfg         // Cache config
	listenCfg        *ListenCfg        // Listen config
	httpCfg          *HTTPCfg          // HTTP config
	filterSCfg       *FilterSCfg       // FilterS config
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
	registrarCCfg    *RegistrarCCfgs   // RegistrarC config
	loaderCgrCfg     *LoaderCgrCfg     // LoaderCgr config
	migratorCgrCfg   *MigratorCgrCfg   // MigratorCgr config
	analyzerSCfg     *AnalyzerSCfg     // AnalyzerS config
	admS             *AdminSCfg        // APIer config
	ersCfg           *ERsCfg           // EventReader config
	eesCfg           *EEsCfg           // EventExporter config
	rateSCfg         *RateSCfg         // RateS config
	actionSCfg       *ActionSCfg       // ActionS config
	sipAgentCfg      *SIPAgentCfg      // SIPAgent config
	configSCfg       *ConfigSCfg       // ConfigS config
	apiBanCfg        *APIBanCfg        // APIBan config
	coreSCfg         *CoreSCfg         // CoreS config
	accountSCfg      *AccountSCfg      // AccountS config
	configDBCfg      *DataDbCfg        // ConfigDB conifg

	cacheDP    utils.MapStorage
	cacheDPMux sync.RWMutex

	db ConfigDB // to store the last dbConn that executed an config update
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

// loadFromJSONCfg Loads from json configuration object, will be used for defaults, config from file and reload, might need lock
func (cfg *CGRConfig) loadFromJSONCfg(jsnCfg ConfigDB) (err error) {
	// Load sections out of JSON config, stop on error
	for _, loadFunc := range []func(ConfigDB) error{
		cfg.loadRPCConns,
		cfg.loadGeneralCfg, cfg.loadTemplateSCfg, cfg.loadCacheCfg, cfg.loadListenCfg,
		cfg.loadHTTPCfg, cfg.loadDataDBCfg, cfg.loadStorDBCfg,
		cfg.loadFilterSCfg,
		cfg.loadCdrsCfg, cfg.loadSessionSCfg,
		cfg.loadFreeswitchAgentCfg, cfg.loadKamAgentCfg,
		cfg.loadAsteriskAgentCfg, cfg.loadDiameterAgentCfg, cfg.loadRadiusAgentCfg,
		cfg.loadDNSAgentCfg, cfg.loadHTTPAgentCfg, cfg.loadAttributeSCfg,
		cfg.loadChargerSCfg, cfg.loadResourceSCfg, cfg.loadStatSCfg,
		cfg.loadThresholdSCfg, cfg.loadRouteSCfg, cfg.loadLoaderSCfg,
		cfg.loadSureTaxCfg, cfg.loadDispatcherSCfg,
		cfg.loadLoaderCgrCfg, cfg.loadMigratorCgrCfg, cfg.loadTLSCgrCfg,
		cfg.loadAnalyzerCgrCfg, cfg.loadApierCfg, cfg.loadErsCfg, cfg.loadEesCfg,
		cfg.loadRateSCfg, cfg.loadSIPAgentCfg, cfg.loadRegistrarCCfg,
		cfg.loadConfigSCfg, cfg.loadAPIBanCgrCfg, cfg.loadCoreSCfg, cfg.loadActionSCfg,
		cfg.loadAccountSCfg, cfg.loadConfigDBCfg} {
		if err = loadFunc(jsnCfg); err != nil {
			return
		}
	}
	return
}

// loadRPCConns loads the RPCConns section of the configuration
func (cfg *CGRConfig) loadRPCConns(jsnCfg ConfigDB) (err error) {
	var jsnRPCConns RPCConnsJson
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
	cfg.rpcConns[utils.MetaLocalHost] = &RPCConn{
		Strategy: rpcclient.PoolFirst,
		PoolSize: 0,
		Conns: []*RemoteHost{{
			Address:   "127.0.0.1:2012",
			Transport: utils.MetaJSON,
		}},
	}
	cfg.rpcConns[utils.MetaBiJSONLocalHost] = &RPCConn{
		Strategy: rpcclient.PoolFirst,
		PoolSize: 0,
		Conns: []*RemoteHost{{
			Address:   "127.0.0.1:2014",
			Transport: rpcclient.BiRPCJSON,
		}},
	}
	for key, val := range jsnRPCConns {
		cfg.rpcConns[key] = NewDfltRPCConn()
		cfg.rpcConns[key].loadFromJSONCfg(val)
	}
	return
}

// loadGeneralCfg loads the General section of the configuration
func (cfg *CGRConfig) loadGeneralCfg(jsnCfg ConfigDB) (err error) {
	var jsnGeneralCfg *GeneralJsonCfg
	if jsnGeneralCfg, err = jsnCfg.GeneralJsonCfg(); err != nil {
		return
	}
	return cfg.generalCfg.loadFromJSONCfg(jsnGeneralCfg)
}

// loadCacheCfg loads the Cache section of the configuration
func (cfg *CGRConfig) loadCacheCfg(jsnCfg ConfigDB) (err error) {
	var jsnCacheCfg *CacheJsonCfg
	if jsnCacheCfg, err = jsnCfg.CacheJsonCfg(); err != nil {
		return
	}
	return cfg.cacheCfg.loadFromJSONCfg(jsnCacheCfg)
}

// loadListenCfg loads the Listen section of the configuration
func (cfg *CGRConfig) loadListenCfg(jsnCfg ConfigDB) (err error) {
	var jsnListenCfg *ListenJsonCfg
	if jsnListenCfg, err = jsnCfg.ListenJsonCfg(); err != nil {
		return
	}
	return cfg.listenCfg.loadFromJSONCfg(jsnListenCfg)
}

// loadHTTPCfg loads the Http section of the configuration
func (cfg *CGRConfig) loadHTTPCfg(jsnCfg ConfigDB) (err error) {
	var jsnHTTPCfg *HTTPJsonCfg
	if jsnHTTPCfg, err = jsnCfg.HttpJsonCfg(); err != nil {
		return
	}
	return cfg.httpCfg.loadFromJSONCfg(jsnHTTPCfg)
}

// loadDataDBCfg loads the DataDB section of the configuration
func (cfg *CGRConfig) loadDataDBCfg(jsnCfg ConfigDB) (err error) {
	var jsnDataDbCfg *DbJsonCfg
	if jsnDataDbCfg, err = jsnCfg.DbJsonCfg(DataDBJSON); err != nil {
		return
	}
	if err = cfg.dataDbCfg.loadFromJSONCfg(jsnDataDbCfg); err != nil {
		return
	}
	return

}

// loadStorDBCfg loads the StorDB section of the configuration
func (cfg *CGRConfig) loadStorDBCfg(jsnCfg ConfigDB) (err error) {
	var jsnDataDbCfg *DbJsonCfg
	if jsnDataDbCfg, err = jsnCfg.DbJsonCfg(StorDBJSON); err != nil {
		return
	}
	return cfg.storDbCfg.loadFromJSONCfg(jsnDataDbCfg)
}

// loadFilterSCfg loads the FilterS section of the configuration
func (cfg *CGRConfig) loadFilterSCfg(jsnCfg ConfigDB) (err error) {
	var jsnFilterSCfg *FilterSJsonCfg
	if jsnFilterSCfg, err = jsnCfg.FilterSJsonCfg(); err != nil {
		return
	}
	return cfg.filterSCfg.loadFromJSONCfg(jsnFilterSCfg)
}

// loadCdrsCfg loads the Cdrs section of the configuration
func (cfg *CGRConfig) loadCdrsCfg(jsnCfg ConfigDB) (err error) {
	var jsnCdrsCfg *CdrsJsonCfg
	if jsnCdrsCfg, err = jsnCfg.CdrsJsonCfg(); err != nil {
		return
	}
	return cfg.cdrsCfg.loadFromJSONCfg(jsnCdrsCfg)
}

// loadSessionSCfg loads the SessionS section of the configuration
func (cfg *CGRConfig) loadSessionSCfg(jsnCfg ConfigDB) (err error) {
	var jsnSessionSCfg *SessionSJsonCfg
	if jsnSessionSCfg, err = jsnCfg.SessionSJsonCfg(); err != nil {
		return
	}
	return cfg.sessionSCfg.loadFromJSONCfg(jsnSessionSCfg)
}

// loadFreeswitchAgentCfg loads the FreeswitchAgent section of the configuration
func (cfg *CGRConfig) loadFreeswitchAgentCfg(jsnCfg ConfigDB) (err error) {
	var jsnSmFsCfg *FreeswitchAgentJsonCfg
	if jsnSmFsCfg, err = jsnCfg.FreeswitchAgentJsonCfg(); err != nil {
		return
	}
	return cfg.fsAgentCfg.loadFromJSONCfg(jsnSmFsCfg)
}

// loadKamAgentCfg loads the KamAgent section of the configuration
func (cfg *CGRConfig) loadKamAgentCfg(jsnCfg ConfigDB) (err error) {
	var jsnKamAgentCfg *KamAgentJsonCfg
	if jsnKamAgentCfg, err = jsnCfg.KamAgentJsonCfg(); err != nil {
		return
	}
	return cfg.kamAgentCfg.loadFromJSONCfg(jsnKamAgentCfg)
}

// loadAsteriskAgentCfg loads the AsteriskAgent section of the configuration
func (cfg *CGRConfig) loadAsteriskAgentCfg(jsnCfg ConfigDB) (err error) {
	var jsnSMAstCfg *AsteriskAgentJsonCfg
	if jsnSMAstCfg, err = jsnCfg.AsteriskAgentJsonCfg(); err != nil {
		return
	}
	return cfg.asteriskAgentCfg.loadFromJSONCfg(jsnSMAstCfg)
}

// loadDiameterAgentCfg loads the DiameterAgent section of the configuration
func (cfg *CGRConfig) loadDiameterAgentCfg(jsnCfg ConfigDB) (err error) {
	var jsnDACfg *DiameterAgentJsonCfg
	if jsnDACfg, err = jsnCfg.DiameterAgentJsonCfg(); err != nil {
		return
	}
	return cfg.diameterAgentCfg.loadFromJSONCfg(jsnDACfg, cfg.generalCfg.RSRSep)
}

// loadRadiusAgentCfg loads the RadiusAgent section of the configuration
func (cfg *CGRConfig) loadRadiusAgentCfg(jsnCfg ConfigDB) (err error) {
	var jsnRACfg *RadiusAgentJsonCfg
	if jsnRACfg, err = jsnCfg.RadiusAgentJsonCfg(); err != nil {
		return
	}
	return cfg.radiusAgentCfg.loadFromJSONCfg(jsnRACfg, cfg.generalCfg.RSRSep)
}

// loadDNSAgentCfg loads the DNSAgent section of the configuration
func (cfg *CGRConfig) loadDNSAgentCfg(jsnCfg ConfigDB) (err error) {
	var jsnDNSCfg *DNSAgentJsonCfg
	if jsnDNSCfg, err = jsnCfg.DNSAgentJsonCfg(); err != nil {
		return
	}
	return cfg.dnsAgentCfg.loadFromJSONCfg(jsnDNSCfg, cfg.generalCfg.RSRSep)
}

// loadHTTPAgentCfg loads the HttpAgent section of the configuration
func (cfg *CGRConfig) loadHTTPAgentCfg(jsnCfg ConfigDB) (err error) {
	var jsnHTTPAgntCfg *[]*HttpAgentJsonCfg
	if jsnHTTPAgntCfg, err = jsnCfg.HttpAgentJsonCfg(); err != nil {
		return
	}
	return cfg.httpAgentCfg.loadFromJSONCfg(jsnHTTPAgntCfg, cfg.generalCfg.RSRSep)
}

// loadAttributeSCfg loads the AttributeS section of the configuration
func (cfg *CGRConfig) loadAttributeSCfg(jsnCfg ConfigDB) (err error) {
	var jsnAttributeSCfg *AttributeSJsonCfg
	if jsnAttributeSCfg, err = jsnCfg.AttributeServJsonCfg(); err != nil {
		return
	}
	return cfg.attributeSCfg.loadFromJSONCfg(jsnAttributeSCfg)
}

// loadChargerSCfg loads the ChargerS section of the configuration
func (cfg *CGRConfig) loadChargerSCfg(jsnCfg ConfigDB) (err error) {
	var jsnChargerSCfg *ChargerSJsonCfg
	if jsnChargerSCfg, err = jsnCfg.ChargerServJsonCfg(); err != nil {
		return
	}
	return cfg.chargerSCfg.loadFromJSONCfg(jsnChargerSCfg)
}

// loadResourceSCfg loads the ResourceS section of the configuration
func (cfg *CGRConfig) loadResourceSCfg(jsnCfg ConfigDB) (err error) {
	var jsnRLSCfg *ResourceSJsonCfg
	if jsnRLSCfg, err = jsnCfg.ResourceSJsonCfg(); err != nil {
		return
	}
	return cfg.resourceSCfg.loadFromJSONCfg(jsnRLSCfg)
}

// loadStatSCfg loads the StatS section of the configuration
func (cfg *CGRConfig) loadStatSCfg(jsnCfg ConfigDB) (err error) {
	var jsnStatSCfg *StatServJsonCfg
	if jsnStatSCfg, err = jsnCfg.StatSJsonCfg(); err != nil {
		return
	}
	return cfg.statsCfg.loadFromJSONCfg(jsnStatSCfg)
}

// loadThresholdSCfg loads the ThresholdS section of the configuration
func (cfg *CGRConfig) loadThresholdSCfg(jsnCfg ConfigDB) (err error) {
	var jsnThresholdSCfg *ThresholdSJsonCfg
	if jsnThresholdSCfg, err = jsnCfg.ThresholdSJsonCfg(); err != nil {
		return
	}
	return cfg.thresholdSCfg.loadFromJSONCfg(jsnThresholdSCfg)
}

// loadRouteSCfg loads the RouteS section of the configuration
func (cfg *CGRConfig) loadRouteSCfg(jsnCfg ConfigDB) (err error) {
	var jsnRouteSCfg *RouteSJsonCfg
	if jsnRouteSCfg, err = jsnCfg.RouteSJsonCfg(); err != nil {
		return
	}
	return cfg.routeSCfg.loadFromJSONCfg(jsnRouteSCfg)
}

// loadLoaderSCfg loads the LoaderS section of the configuration
func (cfg *CGRConfig) loadLoaderSCfg(jsnCfg ConfigDB) (err error) {
	var jsnLoaderCfg []*LoaderJsonCfg
	if jsnLoaderCfg, err = jsnCfg.LoaderJsonCfg(); err != nil {
		return
	}
	// cfg.loaderCfg = make(LoaderSCfgs, len(jsnLoaderCfg))
	for _, profile := range jsnLoaderCfg {
		loadSCfgp := NewDfltLoaderSCfg()
		if err = loadSCfgp.loadFromJSONCfg(profile, cfg.templates, cfg.generalCfg.RSRSep); err != nil {
			return
		}
		cfg.loaderCfg = append(cfg.loaderCfg, loadSCfgp) // use append so the loaderS profile to be loaded from multiple files
	}
	return
}

// loadSureTaxCfg loads the SureTax section of the configuration
func (cfg *CGRConfig) loadSureTaxCfg(jsnCfg ConfigDB) (err error) {
	var jsnSureTaxCfg *SureTaxJsonCfg
	if jsnSureTaxCfg, err = jsnCfg.SureTaxJsonCfg(); err != nil {
		return
	}
	return cfg.sureTaxCfg.loadFromJSONCfg(jsnSureTaxCfg)
}

// loadDispatcherSCfg loads the DispatcherS section of the configuration
func (cfg *CGRConfig) loadDispatcherSCfg(jsnCfg ConfigDB) (err error) {
	var jsnDispatcherSCfg *DispatcherSJsonCfg
	if jsnDispatcherSCfg, err = jsnCfg.DispatcherSJsonCfg(); err != nil {
		return
	}
	return cfg.dispatcherSCfg.loadFromJSONCfg(jsnDispatcherSCfg)
}

// loadRegistrarCCfg loads the RegistrarC section of the configuration
func (cfg *CGRConfig) loadRegistrarCCfg(jsnCfg ConfigDB) (err error) {
	var jsnRegistrarCCfg *RegistrarCJsonCfgs
	if jsnRegistrarCCfg, err = jsnCfg.RegistrarCJsonCfgs(); err != nil {
		return
	}
	return cfg.registrarCCfg.loadFromJSONCfg(jsnRegistrarCCfg)
}

// loadLoaderCgrCfg loads the Loader section of the configuration
func (cfg *CGRConfig) loadLoaderCgrCfg(jsnCfg ConfigDB) (err error) {
	var jsnLoaderCgrCfg *LoaderCfgJson
	if jsnLoaderCgrCfg, err = jsnCfg.LoaderCfgJson(); err != nil {
		return
	}
	return cfg.loaderCgrCfg.loadFromJSONCfg(jsnLoaderCgrCfg)
}

// loadMigratorCgrCfg loads the Migrator section of the configuration
func (cfg *CGRConfig) loadMigratorCgrCfg(jsnCfg ConfigDB) (err error) {
	var jsnMigratorCgrCfg *MigratorCfgJson
	if jsnMigratorCgrCfg, err = jsnCfg.MigratorCfgJson(); err != nil {
		return
	}
	return cfg.migratorCgrCfg.loadFromJSONCfg(jsnMigratorCgrCfg)
}

// loadTLSCgrCfg loads the Tls section of the configuration
func (cfg *CGRConfig) loadTLSCgrCfg(jsnCfg ConfigDB) (err error) {
	var jsnTLSCgrCfg *TlsJsonCfg
	if jsnTLSCgrCfg, err = jsnCfg.TlsCfgJson(); err != nil {
		return
	}
	return cfg.tlsCfg.loadFromJSONCfg(jsnTLSCgrCfg)
}

// loadAnalyzerCgrCfg loads the Analyzer section of the configuration
func (cfg *CGRConfig) loadAnalyzerCgrCfg(jsnCfg ConfigDB) (err error) {
	var jsnAnalyzerCgrCfg *AnalyzerSJsonCfg
	if jsnAnalyzerCgrCfg, err = jsnCfg.AnalyzerCfgJson(); err != nil {
		return
	}
	return cfg.analyzerSCfg.loadFromJSONCfg(jsnAnalyzerCgrCfg)
}

// loadAPIBanCgrCfg loads the Analyzer section of the configuration
func (cfg *CGRConfig) loadAPIBanCgrCfg(jsnCfg ConfigDB) (err error) {
	var jsnAPIBanCfg *APIBanJsonCfg
	if jsnAPIBanCfg, err = jsnCfg.ApiBanCfgJson(); err != nil {
		return
	}
	return cfg.apiBanCfg.loadFromJSONCfg(jsnAPIBanCfg)
}

// loadApierCfg loads the Apier section of the configuration
func (cfg *CGRConfig) loadApierCfg(jsnCfg ConfigDB) (err error) {
	var jsnApierCfg *AdminSJsonCfg
	if jsnApierCfg, err = jsnCfg.AdminSCfgJson(); err != nil {
		return
	}
	return cfg.admS.loadFromJSONCfg(jsnApierCfg)
}

// loadCoreSCfg loads the CoreS section of the configuration
func (cfg *CGRConfig) loadCoreSCfg(jsnCfg ConfigDB) (err error) {
	var jsnCoreCfg *CoreSJsonCfg
	if jsnCoreCfg, err = jsnCfg.CoreSJSON(); err != nil {
		return
	}
	return cfg.coreSCfg.loadFromJSONCfg(jsnCoreCfg)
}

// loadErsCfg loads the Ers section of the configuration
func (cfg *CGRConfig) loadErsCfg(jsnCfg ConfigDB) (err error) {
	var jsnERsCfg *ERsJsonCfg
	if jsnERsCfg, err = jsnCfg.ERsJsonCfg(); err != nil {
		return
	}
	return cfg.ersCfg.loadFromJSONCfg(jsnERsCfg, cfg.templates, cfg.generalCfg.RSRSep, cfg.dfltEvRdr)
}

// loadEesCfg loads the Ees section of the configuration
func (cfg *CGRConfig) loadEesCfg(jsnCfg ConfigDB) (err error) {
	var jsnEEsCfg *EEsJsonCfg
	if jsnEEsCfg, err = jsnCfg.EEsJsonCfg(); err != nil {
		return
	}
	return cfg.eesCfg.loadFromJSONCfg(jsnEEsCfg, cfg.templates, cfg.generalCfg.RSRSep, cfg.dfltEvExp)
}

// loadRateSCfg loads the rates section of the configuration
func (cfg *CGRConfig) loadRateSCfg(jsnCfg ConfigDB) (err error) {
	var jsnRateCfg *RateSJsonCfg
	if jsnRateCfg, err = jsnCfg.RateCfgJson(); err != nil {
		return
	}
	return cfg.rateSCfg.loadFromJSONCfg(jsnRateCfg)
}

// loadSIPAgentCfg loads the sip_agent section of the configuration
func (cfg *CGRConfig) loadSIPAgentCfg(jsnCfg ConfigDB) (err error) {
	var jsnSIPAgentCfg *SIPAgentJsonCfg
	if jsnSIPAgentCfg, err = jsnCfg.SIPAgentJsonCfg(); err != nil {
		return
	}
	return cfg.sipAgentCfg.loadFromJSONCfg(jsnSIPAgentCfg, cfg.generalCfg.RSRSep)
}

// loadTemplateSCfg loads the Template section of the configuration
func (cfg *CGRConfig) loadTemplateSCfg(jsnCfg ConfigDB) (err error) {
	var jsnTemplateCfg map[string][]*FcTemplateJsonCfg
	if jsnTemplateCfg, err = jsnCfg.TemplateSJsonCfg(); err != nil {
		return
	}
	for k, val := range jsnTemplateCfg {
		if cfg.templates[k], err = FCTemplatesFromFCTemplatesJSONCfg(val, cfg.generalCfg.RSRSep); err != nil {
			return
		}
	}
	return
}

func (cfg *CGRConfig) loadConfigSCfg(jsnCfg ConfigDB) (err error) {
	var jsnConfigSCfg *ConfigSCfgJson
	if jsnConfigSCfg, err = jsnCfg.ConfigSJsonCfg(); err != nil {
		return
	}
	return cfg.configSCfg.loadFromJSONCfg(jsnConfigSCfg)
}

// loadActionSCfg loads the ActionS section of the configuration
func (cfg *CGRConfig) loadActionSCfg(jsnCfg ConfigDB) (err error) {
	var jsnActionCfg *ActionSJsonCfg
	if jsnActionCfg, err = jsnCfg.ActionSCfgJson(); err != nil {
		return
	}
	return cfg.actionSCfg.loadFromJSONCfg(jsnActionCfg)
}

// loadAccountSCfg loads the AccountS section of the configuration
func (cfg *CGRConfig) loadAccountSCfg(jsnCfg ConfigDB) (err error) {
	var jsnActionCfg *AccountSJsonCfg
	if jsnActionCfg, err = jsnCfg.AccountSCfgJson(); err != nil {
		return
	}
	return cfg.accountSCfg.loadFromJSONCfg(jsnActionCfg)
}

// loadConfigDBCfg loads the ConfigDB section of the configuration
func (cfg *CGRConfig) loadConfigDBCfg(jsnCfg ConfigDB) (err error) {
	var jsnDBCfg *DbJsonCfg
	if jsnDBCfg, err = jsnCfg.DbJsonCfg(ConfigDBJSON); err != nil {
		return
	}
	return cfg.configDBCfg.loadFromJSONCfg(jsnDBCfg)
}

// SureTaxCfg use locking to retrieve the configuration, possibility later for runtime reload
func (cfg *CGRConfig) SureTaxCfg() *SureTaxCfg {
	cfg.lks[SureTaxJSON].Lock()
	defer cfg.lks[SureTaxJSON].Unlock()
	return cfg.sureTaxCfg
}

// DiameterAgentCfg returns the config for Diameter Agent
func (cfg *CGRConfig) DiameterAgentCfg() *DiameterAgentCfg {
	cfg.lks[DiameterAgentJSON].Lock()
	defer cfg.lks[DiameterAgentJSON].Unlock()
	return cfg.diameterAgentCfg
}

// RadiusAgentCfg returns the config for Radius Agent
func (cfg *CGRConfig) RadiusAgentCfg() *RadiusAgentCfg {
	cfg.lks[RadiusAgentJSON].Lock()
	defer cfg.lks[RadiusAgentJSON].Unlock()
	return cfg.radiusAgentCfg
}

// DNSAgentCfg returns the config for DNS Agent
func (cfg *CGRConfig) DNSAgentCfg() *DNSAgentCfg {
	cfg.lks[DNSAgentJSON].Lock()
	defer cfg.lks[DNSAgentJSON].Unlock()
	return cfg.dnsAgentCfg
}

// AttributeSCfg returns the config for AttributeS
func (cfg *CGRConfig) AttributeSCfg() *AttributeSCfg {
	cfg.lks[AttributeSJSON].Lock()
	defer cfg.lks[AttributeSJSON].Unlock()
	return cfg.attributeSCfg
}

// ChargerSCfg returns the config for ChargerS
func (cfg *CGRConfig) ChargerSCfg() *ChargerSCfg {
	cfg.lks[ChargerSJSON].Lock()
	defer cfg.lks[ChargerSJSON].Unlock()
	return cfg.chargerSCfg
}

// ResourceSCfg returns the config for ResourceS
func (cfg *CGRConfig) ResourceSCfg() *ResourceSConfig { // not done
	cfg.lks[ResourceSJSON].Lock()
	defer cfg.lks[ResourceSJSON].Unlock()
	return cfg.resourceSCfg
}

// StatSCfg returns the config for StatS
func (cfg *CGRConfig) StatSCfg() *StatSCfg { // not done
	cfg.lks[StatSJSON].Lock()
	defer cfg.lks[StatSJSON].Unlock()
	return cfg.statsCfg
}

// ThresholdSCfg returns the config for ThresholdS
func (cfg *CGRConfig) ThresholdSCfg() *ThresholdSCfg {
	cfg.lks[ThresholdSJSON].Lock()
	defer cfg.lks[ThresholdSJSON].Unlock()
	return cfg.thresholdSCfg
}

// RouteSCfg returns the config for RouteS
func (cfg *CGRConfig) RouteSCfg() *RouteSCfg {
	cfg.lks[RouteSJSON].Lock()
	defer cfg.lks[RouteSJSON].Unlock()
	return cfg.routeSCfg
}

// SessionSCfg returns the config for SessionS
func (cfg *CGRConfig) SessionSCfg() *SessionSCfg {
	cfg.lks[SessionSJSON].Lock()
	defer cfg.lks[SessionSJSON].Unlock()
	return cfg.sessionSCfg
}

// FsAgentCfg returns the config for FsAgent
func (cfg *CGRConfig) FsAgentCfg() *FsAgentCfg {
	cfg.lks[FreeSWITCHAgentJSON].Lock()
	defer cfg.lks[FreeSWITCHAgentJSON].Unlock()
	return cfg.fsAgentCfg
}

// KamAgentCfg returns the config for KamAgent
func (cfg *CGRConfig) KamAgentCfg() *KamAgentCfg {
	cfg.lks[KamailioAgentJSON].Lock()
	defer cfg.lks[KamailioAgentJSON].Unlock()
	return cfg.kamAgentCfg
}

// AsteriskAgentCfg returns the config for AsteriskAgent
func (cfg *CGRConfig) AsteriskAgentCfg() *AsteriskAgentCfg {
	cfg.lks[AsteriskAgentJSON].Lock()
	defer cfg.lks[AsteriskAgentJSON].Unlock()
	return cfg.asteriskAgentCfg
}

// HTTPAgentCfg returns the config for HttpAgent
func (cfg *CGRConfig) HTTPAgentCfg() HTTPAgentCfgs {
	cfg.lks[HTTPAgentJSON].Lock()
	defer cfg.lks[HTTPAgentJSON].Unlock()
	return cfg.httpAgentCfg
}

// FilterSCfg returns the config for FilterS
func (cfg *CGRConfig) FilterSCfg() *FilterSCfg {
	cfg.lks[FilterSJSON].Lock()
	defer cfg.lks[FilterSJSON].Unlock()
	return cfg.filterSCfg
}

// CacheCfg returns the config for Cache
func (cfg *CGRConfig) CacheCfg() *CacheCfg {
	cfg.lks[CacheJSON].Lock()
	defer cfg.lks[CacheJSON].Unlock()
	return cfg.cacheCfg
}

// LoaderCfg returns the Loader Service
func (cfg *CGRConfig) LoaderCfg() LoaderSCfgs {
	cfg.lks[LoaderSJSON].Lock()
	defer cfg.lks[LoaderSJSON].Unlock()
	return cfg.loaderCfg
}

// LoaderCgrCfg returns the config for cgr-loader
func (cfg *CGRConfig) LoaderCgrCfg() *LoaderCgrCfg {
	cfg.lks[LoaderJSON].Lock()
	defer cfg.lks[LoaderJSON].Unlock()
	return cfg.loaderCgrCfg
}

// DispatcherSCfg returns the config for DispatcherS
func (cfg *CGRConfig) DispatcherSCfg() *DispatcherSCfg {
	cfg.lks[DispatcherSJSON].Lock()
	defer cfg.lks[DispatcherSJSON].Unlock()
	return cfg.dispatcherSCfg
}

// RegistrarCCfg returns the config for RegistrarC
func (cfg *CGRConfig) RegistrarCCfg() *RegistrarCCfgs {
	cfg.lks[DispatcherSJSON].Lock()
	defer cfg.lks[DispatcherSJSON].Unlock()
	return cfg.registrarCCfg
}

// MigratorCgrCfg returns the config for Migrator
func (cfg *CGRConfig) MigratorCgrCfg() *MigratorCgrCfg {
	cfg.lks[MigratorJSON].Lock()
	defer cfg.lks[MigratorJSON].Unlock()
	return cfg.migratorCgrCfg
}

// DataDbCfg returns the config for DataDb
func (cfg *CGRConfig) DataDbCfg() *DataDbCfg {
	cfg.lks[DataDBJSON].Lock()
	defer cfg.lks[DataDBJSON].Unlock()
	return cfg.dataDbCfg
}

// StorDbCfg returns the config for StorDb
func (cfg *CGRConfig) StorDbCfg() *StorDbCfg {
	cfg.lks[StorDBJSON].Lock()
	defer cfg.lks[StorDBJSON].Unlock()
	return cfg.storDbCfg
}

// GeneralCfg returns the General config section
func (cfg *CGRConfig) GeneralCfg() *GeneralCfg {
	cfg.lks[GeneralJSON].Lock()
	defer cfg.lks[GeneralJSON].Unlock()
	return cfg.generalCfg
}

// TLSCfg returns the config for Tls
func (cfg *CGRConfig) TLSCfg() *TLSCfg {
	cfg.lks[TlsJSON].Lock()
	defer cfg.lks[TlsJSON].Unlock()
	return cfg.tlsCfg
}

// ListenCfg returns the server Listen config
func (cfg *CGRConfig) ListenCfg() *ListenCfg {
	cfg.lks[ListenJSON].Lock()
	defer cfg.lks[ListenJSON].Unlock()
	return cfg.listenCfg
}

// HTTPCfg returns the config for HTTP
func (cfg *CGRConfig) HTTPCfg() *HTTPCfg {
	cfg.lks[HTTPJSON].Lock()
	defer cfg.lks[HTTPJSON].Unlock()
	return cfg.httpCfg
}

// CdrsCfg returns the config for CDR Server
func (cfg *CGRConfig) CdrsCfg() *CdrsCfg {
	cfg.lks[CDRsJSON].Lock()
	defer cfg.lks[CDRsJSON].Unlock()
	return cfg.cdrsCfg
}

// AnalyzerSCfg returns the config for AnalyzerS
func (cfg *CGRConfig) AnalyzerSCfg() *AnalyzerSCfg {
	cfg.lks[AnalyzerSJSON].Lock()
	defer cfg.lks[AnalyzerSJSON].Unlock()
	return cfg.analyzerSCfg
}

// AdminSCfg reads the Apier configuration
func (cfg *CGRConfig) AdminSCfg() *AdminSCfg {
	cfg.lks[AdminSJSON].Lock()
	defer cfg.lks[AdminSJSON].Unlock()
	return cfg.admS
}

// ERsCfg reads the EventReader configuration
func (cfg *CGRConfig) ERsCfg() *ERsCfg {
	cfg.lks[ERsJSON].RLock()
	defer cfg.lks[ERsJSON].RUnlock()
	return cfg.ersCfg
}

// EEsCfg reads the EventExporter configuration
func (cfg *CGRConfig) EEsCfg() *EEsCfg {
	cfg.lks[EEsJSON].RLock()
	defer cfg.lks[EEsJSON].RUnlock()
	return cfg.eesCfg
}

// EEsNoLksCfg reads the EventExporter configuration without locks
func (cfg *CGRConfig) EEsNoLksCfg() *EEsCfg {
	return cfg.eesCfg
}

// RateSCfg reads the RateS configuration
func (cfg *CGRConfig) RateSCfg() *RateSCfg {
	cfg.lks[RateSJSON].RLock()
	defer cfg.lks[RateSJSON].RUnlock()
	return cfg.rateSCfg
}

// ActionSCfg reads the ActionS configuration
func (cfg *CGRConfig) ActionSCfg() *ActionSCfg {
	cfg.lks[ActionSJSON].RLock()
	defer cfg.lks[ActionSJSON].RUnlock()
	return cfg.actionSCfg
}

// AccountSCfg reads the AccountS configuration
func (cfg *CGRConfig) AccountSCfg() *AccountSCfg {
	cfg.lks[AccountSJSON].RLock()
	defer cfg.lks[AccountSJSON].RUnlock()
	return cfg.accountSCfg
}

// SIPAgentCfg reads the Apier configuration
func (cfg *CGRConfig) SIPAgentCfg() *SIPAgentCfg {
	cfg.lks[SIPAgentJSON].Lock()
	defer cfg.lks[SIPAgentJSON].Unlock()
	return cfg.sipAgentCfg
}

// RPCConns reads the RPCConns configuration
func (cfg *CGRConfig) RPCConns() RPCConns {
	cfg.lks[RPCConnsJSON].RLock()
	defer cfg.lks[RPCConnsJSON].RUnlock()
	return cfg.rpcConns
}

// TemplatesCfg returns the config for templates
func (cfg *CGRConfig) TemplatesCfg() FCTemplates {
	cfg.lks[TemplatesJSON].Lock()
	defer cfg.lks[TemplatesJSON].Unlock()
	return cfg.templates
}

// ConfigSCfg returns the configs configuration
func (cfg *CGRConfig) ConfigSCfg() *ConfigSCfg {
	cfg.lks[ConfigSJSON].RLock()
	defer cfg.lks[ConfigSJSON].RUnlock()
	return cfg.configSCfg
}

// APIBanCfg reads the ApiBan configuration
func (cfg *CGRConfig) APIBanCfg() *APIBanCfg {
	cfg.lks[APIBanJSON].Lock()
	defer cfg.lks[APIBanJSON].Unlock()
	return cfg.apiBanCfg
}

// CoreSCfg reads the CoreS configuration
func (cfg *CGRConfig) CoreSCfg() *CoreSCfg {
	cfg.lks[CoreSJSON].Lock()
	defer cfg.lks[CoreSJSON].Unlock()
	return cfg.coreSCfg
}

// ConfigDBCfg reads the CoreS configuration
func (cfg *CGRConfig) ConfigDBCfg() *DataDbCfg {
	cfg.lks[ConfigDBJSON].Lock()
	defer cfg.lks[ConfigDBJSON].Unlock()
	return cfg.configDBCfg
}

// GetReloadChan returns the reload chanel for the given section
func (cfg *CGRConfig) GetReloadChan(sectID string) chan struct{} {
	return cfg.rldChans[sectID]
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

// LockSections will lock the given sections
// User needs to know what he is doing since this can panic
func (cfg *CGRConfig) LockSections(lkIDs ...string) {
	for _, lkID := range lkIDs {
		cfg.lks[lkID].Lock()
	}
}

// UnlockSections will unlock the given sections
// User needs to know what he is doing since this can panic
func (cfg *CGRConfig) UnlockSections(lkIDs ...string) {
	for _, lkID := range lkIDs {
		cfg.lks[lkID].Unlock()
	}
}

func (cfg *CGRConfig) getLoadFunctions() map[string]func(ConfigDB) error {
	return map[string]func(ConfigDB) error{
		GeneralJSON:         cfg.loadGeneralCfg,
		DataDBJSON:          cfg.loadDataDBCfg,
		StorDBJSON:          cfg.loadStorDBCfg,
		ListenJSON:          cfg.loadListenCfg,
		TlsJSON:             cfg.loadTLSCgrCfg,
		HTTPJSON:            cfg.loadHTTPCfg,
		CacheJSON:           cfg.loadCacheCfg,
		FilterSJSON:         cfg.loadFilterSCfg,
		CDRsJSON:            cfg.loadCdrsCfg,
		ERsJSON:             cfg.loadErsCfg,
		EEsJSON:             cfg.loadEesCfg,
		SessionSJSON:        cfg.loadSessionSCfg,
		AsteriskAgentJSON:   cfg.loadAsteriskAgentCfg,
		FreeSWITCHAgentJSON: cfg.loadFreeswitchAgentCfg,
		KamailioAgentJSON:   cfg.loadKamAgentCfg,
		DiameterAgentJSON:   cfg.loadDiameterAgentCfg,
		RadiusAgentJSON:     cfg.loadRadiusAgentCfg,
		HTTPAgentJSON:       cfg.loadHTTPAgentCfg,
		DNSAgentJSON:        cfg.loadDNSAgentCfg,
		AttributeSJSON:      cfg.loadAttributeSCfg,
		ChargerSJSON:        cfg.loadChargerSCfg,
		ResourceSJSON:       cfg.loadResourceSCfg,
		StatSJSON:           cfg.loadStatSCfg,
		ThresholdSJSON:      cfg.loadThresholdSCfg,
		RouteSJSON:          cfg.loadRouteSCfg,
		LoaderSJSON:         cfg.loadLoaderSCfg,
		SureTaxJSON:         cfg.loadSureTaxCfg,
		LoaderJSON:          cfg.loadLoaderCgrCfg,
		MigratorJSON:        cfg.loadMigratorCgrCfg,
		DispatcherSJSON:     cfg.loadDispatcherSCfg,
		RegistrarCJSON:      cfg.loadRegistrarCCfg,
		AnalyzerSJSON:       cfg.loadAnalyzerCgrCfg,
		AdminSJSON:          cfg.loadApierCfg,
		RPCConnsJSON:        cfg.loadRPCConns,
		RateSJSON:           cfg.loadRateSCfg,
		SIPAgentJSON:        cfg.loadSIPAgentCfg,
		TemplatesJSON:       cfg.loadTemplateSCfg,
		ConfigSJSON:         cfg.loadConfigSCfg,
		APIBanJSON:          cfg.loadAPIBanCgrCfg,
		CoreSJSON:           cfg.loadCoreSCfg,
		ActionSJSON:         cfg.loadActionSCfg,
		AccountSJSON:        cfg.loadAccountSCfg,
		ConfigDBJSON:        cfg.loadConfigDBCfg,
	}
}

func (cfg *CGRConfig) loadCfgWithLocks(path, section string) (err error) {
	var loadFuncs []func(ConfigDB) error
	loadMap := cfg.getLoadFunctions()
	if section == utils.EmptyString || section == utils.MetaAll {
		cfg.lockSections()
		defer cfg.unlockSections()
		for _, sec := range sortedCfgSections {
			loadFuncs = append(loadFuncs, loadMap[sec])
		}
	} else if fnct, has := loadMap[section]; !has {
		return fmt.Errorf("Invalid section: <%s> ", section)
	} else {
		cfg.lks[section].Lock()
		defer cfg.lks[section].Unlock()
		loadFuncs = append(loadFuncs, fnct)
	}
	return cfg.loadConfigFromPath(path, loadFuncs, false)
}

func (*CGRConfig) loadConfigFromReader(rdr io.Reader, loadFuncs []func(jsnCfg ConfigDB) error, envOff bool) (err error) {
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
func (cfg *CGRConfig) loadConfigFromPath(path string, loadFuncs []func(jsnCfg ConfigDB) error, envOff bool) (err error) {
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

func (cfg *CGRConfig) loadConfigFromFolder(cfgDir string, loadFuncs []func(jsnCfg ConfigDB) error, envOff bool) (err error) {
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
		return fmt.Errorf("No config file found on path %s ", cfgDir)
	}
	return
}

// loadConfigFromFile loads the config from a file
// extracted from a loadConfigFromFolder in order to test all cases
func (cfg *CGRConfig) loadConfigFromFile(jsonFilePath string, loadFuncs []func(jsnCfg ConfigDB) error, envOff bool) (err error) {
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

func (cfg *CGRConfig) loadConfigFromHTTP(urlPaths string, loadFuncs []func(jsnCfg ConfigDB) error) (err error) {
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
		cfg.rldChans[section] = make(chan struct{})
	}
}

func (cfg *CGRConfig) loadCfgFromJSONWithLocks(rdr io.Reader, sections []string) (err error) {
	var loadFuncs []func(ConfigDB) error
	loadMap := cfg.getLoadFunctions()
	cfg.LockSections(sections...)
	defer cfg.UnlockSections(sections...)
	for _, section := range sections {
		loadFuncs = append(loadFuncs, loadMap[section])
	}
	return cfg.loadConfigFromReader(rdr, loadFuncs, false)
}

// reloadSections sends a signal to the reload channel for the needed sections
// the list of sections should be always valid because we load the config first with this list
func (cfg *CGRConfig) reloadSections(sections ...string) {
	subsystemsThatNeedDataDB := utils.NewStringSet([]string{DataDBJSON,
		CDRsJSON, SessionSJSON, AttributeSJSON,
		ChargerSJSON, ResourceSJSON, StatSJSON, ThresholdSJSON,
		RouteSJSON, LoaderSJSON, DispatcherSJSON, RateSJSON, AdminSJSON, AccountSJSON,
		ActionSJSON})
	subsystemsThatNeedStorDB := utils.NewStringSet([]string{StorDBJSON, CDRsJSON, AdminSJSON})
	needsDataDB := false
	needsStorDB := false
	for _, section := range sections {
		if !needsDataDB && subsystemsThatNeedDataDB.Has(section) {
			needsDataDB = true
			cfg.rldChans[DataDBJSON] <- struct{}{} // reload datadb before
		}
		if !needsStorDB && subsystemsThatNeedStorDB.Has(section) {
			needsStorDB = true
			cfg.rldChans[StorDBJSON] <- struct{}{} // reload stordb before
		}
		if needsDataDB && needsStorDB {
			break
		}
	}
	runtime.Gosched()
	for _, section := range sections {
		switch section {
		case ConfigSJSON:
		case GeneralJSON: // nothing to reload
		case RPCConnsJSON: // nothing to reload
			cfg.rldChans[RPCConnsJSON] <- struct{}{}
		case DataDBJSON: // reloaded before
		case StorDBJSON: // reloaded before
		case ListenJSON:
		case CacheJSON:
		case FilterSJSON:
		case SureTaxJSON:
		case LoaderJSON:
		case MigratorJSON:
		case TemplatesJSON:
		case TlsJSON: // nothing to reload
		case APIBanJSON: // nothing to reload
		case CoreSJSON: // nothing to reload
		case HTTPJSON:
			cfg.rldChans[HTTPJSON] <- struct{}{}
		case CDRsJSON:
			cfg.rldChans[CDRsJSON] <- struct{}{}
		case ERsJSON:
			cfg.rldChans[ERsJSON] <- struct{}{}
		case SessionSJSON:
			cfg.rldChans[SessionSJSON] <- struct{}{}
		case AsteriskAgentJSON:
			cfg.rldChans[AsteriskAgentJSON] <- struct{}{}
		case FreeSWITCHAgentJSON:
			cfg.rldChans[FreeSWITCHAgentJSON] <- struct{}{}
		case KamailioAgentJSON:
			cfg.rldChans[KamailioAgentJSON] <- struct{}{}
		case DiameterAgentJSON:
			cfg.rldChans[DiameterAgentJSON] <- struct{}{}
		case RadiusAgentJSON:
			cfg.rldChans[RadiusAgentJSON] <- struct{}{}
		case HTTPAgentJSON:
			cfg.rldChans[HTTPAgentJSON] <- struct{}{}
		case DNSAgentJSON:
			cfg.rldChans[DNSAgentJSON] <- struct{}{}
		case AttributeSJSON:
			cfg.rldChans[AttributeSJSON] <- struct{}{}
		case ChargerSJSON:
			cfg.rldChans[ChargerSJSON] <- struct{}{}
		case ResourceSJSON:
			cfg.rldChans[ResourceSJSON] <- struct{}{}
		case StatSJSON:
			cfg.rldChans[StatSJSON] <- struct{}{}
		case ThresholdSJSON:
			cfg.rldChans[ThresholdSJSON] <- struct{}{}
		case RouteSJSON:
			cfg.rldChans[RouteSJSON] <- struct{}{}
		case LoaderSJSON:
			cfg.rldChans[LoaderSJSON] <- struct{}{}
		case DispatcherSJSON:
			cfg.rldChans[DispatcherSJSON] <- struct{}{}
		case AnalyzerSJSON:
			cfg.rldChans[AnalyzerSJSON] <- struct{}{}
		case AdminSJSON:
			cfg.rldChans[AdminSJSON] <- struct{}{}
		case EEsJSON:
			cfg.rldChans[EEsJSON] <- struct{}{}
		case SIPAgentJSON:
			cfg.rldChans[SIPAgentJSON] <- struct{}{}
		case RateSJSON:
			cfg.rldChans[RateSJSON] <- struct{}{}
		case RegistrarCJSON:
			cfg.rldChans[RegistrarCJSON] <- struct{}{}
		case AccountSJSON:
			cfg.rldChans[AccountSJSON] <- struct{}{}
		case ActionSJSON:
			cfg.rldChans[ActionSJSON] <- struct{}{}
		case ConfigDBJSON: // no reload for this
		}
	}
}

// AsMapInterface returns the config as a map[string]interface{}
func (cfg *CGRConfig) AsMapInterface(separator string) (mp map[string]interface{}) {
	return map[string]interface{}{
		LoaderSJSON:         cfg.loaderCfg.AsMapInterface(separator),
		HTTPAgentJSON:       cfg.httpAgentCfg.AsMapInterface(separator),
		RPCConnsJSON:        cfg.rpcConns.AsMapInterface(),
		GeneralJSON:         cfg.generalCfg.AsMapInterface(),
		DataDBJSON:          cfg.dataDbCfg.AsMapInterface(),
		StorDBJSON:          cfg.storDbCfg.AsMapInterface(),
		TlsJSON:             cfg.tlsCfg.AsMapInterface(),
		CacheJSON:           cfg.cacheCfg.AsMapInterface(),
		ListenJSON:          cfg.listenCfg.AsMapInterface(),
		HTTPJSON:            cfg.httpCfg.AsMapInterface(),
		FilterSJSON:         cfg.filterSCfg.AsMapInterface(),
		CDRsJSON:            cfg.cdrsCfg.AsMapInterface(),
		SessionSJSON:        cfg.sessionSCfg.AsMapInterface(),
		FreeSWITCHAgentJSON: cfg.fsAgentCfg.AsMapInterface(separator),
		KamailioAgentJSON:   cfg.kamAgentCfg.AsMapInterface(),
		AsteriskAgentJSON:   cfg.asteriskAgentCfg.AsMapInterface(),
		DiameterAgentJSON:   cfg.diameterAgentCfg.AsMapInterface(separator),
		RadiusAgentJSON:     cfg.radiusAgentCfg.AsMapInterface(separator),
		DNSAgentJSON:        cfg.dnsAgentCfg.AsMapInterface(separator),
		AttributeSJSON:      cfg.attributeSCfg.AsMapInterface(),
		ChargerSJSON:        cfg.chargerSCfg.AsMapInterface(),
		ResourceSJSON:       cfg.resourceSCfg.AsMapInterface(),
		StatSJSON:           cfg.statsCfg.AsMapInterface(),
		ThresholdSJSON:      cfg.thresholdSCfg.AsMapInterface(),
		RouteSJSON:          cfg.routeSCfg.AsMapInterface(),
		SureTaxJSON:         cfg.sureTaxCfg.AsMapInterface(separator),
		DispatcherSJSON:     cfg.dispatcherSCfg.AsMapInterface(),
		RegistrarCJSON:      cfg.registrarCCfg.AsMapInterface(),
		LoaderJSON:          cfg.loaderCgrCfg.AsMapInterface(),
		MigratorJSON:        cfg.migratorCgrCfg.AsMapInterface(),
		AnalyzerSJSON:       cfg.analyzerSCfg.AsMapInterface(),
		AdminSJSON:          cfg.admS.AsMapInterface(),
		ERsJSON:             cfg.ersCfg.AsMapInterface(separator),
		APIBanJSON:          cfg.apiBanCfg.AsMapInterface(),
		EEsJSON:             cfg.eesCfg.AsMapInterface(separator),
		RateSJSON:           cfg.rateSCfg.AsMapInterface(),
		SIPAgentJSON:        cfg.sipAgentCfg.AsMapInterface(separator),
		TemplatesJSON:       cfg.templates.AsMapInterface(separator),
		ConfigSJSON:         cfg.configSCfg.AsMapInterface(),
		CoreSJSON:           cfg.coreSCfg.AsMapInterface(),
		ActionSJSON:         cfg.actionSCfg.AsMapInterface(),
		AccountSJSON:        cfg.accountSCfg.AsMapInterface(),
		ConfigDBJSON:        cfg.configDBCfg.AsMapInterface(),
	}
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
		registrarCCfg:    cfg.registrarCCfg.Clone(),
		loaderCgrCfg:     cfg.loaderCgrCfg.Clone(),
		migratorCgrCfg:   cfg.migratorCgrCfg.Clone(),
		analyzerSCfg:     cfg.analyzerSCfg.Clone(),
		admS:             cfg.admS.Clone(),
		ersCfg:           cfg.ersCfg.Clone(),
		eesCfg:           cfg.eesCfg.Clone(),
		rateSCfg:         cfg.rateSCfg.Clone(),
		sipAgentCfg:      cfg.sipAgentCfg.Clone(),
		configSCfg:       cfg.configSCfg.Clone(),
		apiBanCfg:        cfg.apiBanCfg.Clone(),
		coreSCfg:         cfg.coreSCfg.Clone(),
		actionSCfg:       cfg.actionSCfg.Clone(),
		accountSCfg:      cfg.accountSCfg.Clone(),
		configDBCfg:      cfg.configDBCfg.Clone(),

		cacheDP: make(utils.MapStorage),
	}
	cln.initChanels()
	return
}

// GetDataProvider returns the config as a data provider interface
func (cfg *CGRConfig) GetDataProvider() utils.DataProvider {
	cfg.cacheDPMux.RLock()
	if len(cfg.cacheDP) < len(sortedCfgSections) {
		cfg.cacheDP = cfg.AsMapInterface(cfg.GeneralCfg().RSRSep)
	}
	mp := cfg.cacheDP.Clone()
	cfg.cacheDPMux.RUnlock()
	return mp
}
