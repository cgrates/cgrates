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
	dbDefaultsCfg            dbDefaults
	cgrCfg                   *CGRConfig  // will be shared
	dfltFsConnConfig         *FsConnCfg  // Default FreeSWITCH Connection configuration, built out of json default configuration
	dfltKamConnConfig        *KamConnCfg // Default Kamailio Connection configuration
	dfltRemoteHost           *RemoteHost
	dfltAstConnCfg           *AsteriskConnCfg
	dfltLoaderConfig         *LoaderSCfg
	dfltLoaderDataTypeConfig *LoaderDataType
)

func newDbDefaults() dbDefaults {
	deflt := dbDefaults{
		utils.MYSQL: map[string]string{
			"DbName": "cgrates",
			"DbPort": "3306",
			"DbPass": "CGRateS.org",
		},
		utils.POSTGRES: map[string]string{
			"DbName": "cgrates",
			"DbPort": "5432",
			"DbPass": "CGRateS.org",
		},
		utils.MONGO: map[string]string{
			"DbName": "cgrates",
			"DbPort": "27017",
			"DbPass": "",
		},
		utils.REDIS: map[string]string{
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
	return utils.CGRATES
}

func (dbDefaults) dbHost(dbType string, flagInput string) string {
	if flagInput != utils.MetaDynamic {
		return flagInput
	}
	return utils.LOCALHOST
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
	cgrCfg, _ = NewDefaultCGRConfig()
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

func NewDefaultCGRConfig() (cfg *CGRConfig, err error) {
	cfg = new(CGRConfig)
	cfg.initChanels()
	cfg.DataFolderPath = "/usr/share/cgrates/"
	cfg.MaxCallDuration = time.Duration(3) * time.Hour // Hardcoded for now

	cfg.rpcConns = make(map[string]*RPCConn)
	cfg.generalCfg = new(GeneralCfg)
	cfg.generalCfg.NodeID = utils.UUIDSha1Prefix()
	cfg.dataDbCfg = new(DataDbCfg)
	cfg.dataDbCfg.Items = make(map[string]*ItemOpt)
	cfg.storDbCfg = new(StorDbCfg)
	cfg.storDbCfg.Items = make(map[string]*ItemOpt)
	cfg.tlsCfg = new(TlsCfg)
	cfg.cacheCfg = make(CacheCfg)
	cfg.listenCfg = new(ListenCfg)
	cfg.httpCfg = new(HTTPCfg)
	cfg.filterSCfg = new(FilterSCfg)
	cfg.ralsCfg = new(RalsCfg)
	cfg.ralsCfg.MaxComputedUsage = make(map[string]time.Duration)
	cfg.ralsCfg.BalanceRatingSubject = make(map[string]string)
	cfg.schedulerCfg = new(SchedulerCfg)
	cfg.cdrsCfg = new(CdrsCfg)
	cfg.CdreProfiles = make(map[string]*CdreCfg)
	cfg.analyzerSCfg = new(AnalyzerSCfg)
	cfg.sessionSCfg = new(SessionSCfg)
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
	cfg.supplierSCfg = new(SupplierSCfg)
	cfg.sureTaxCfg = new(SureTaxCfg)
	cfg.dispatcherSCfg = new(DispatcherSCfg)
	cfg.loaderCgrCfg = new(LoaderCgrCfg)
	cfg.migratorCgrCfg = new(MigratorCgrCfg)
	cfg.mailerCfg = new(MailerCfg)
	cfg.loaderCfg = make(LoaderSCfgs, 0)
	cfg.apier = new(ApierCfg)
	cfg.ersCfg = new(ERsCfg)

	cfg.ConfigReloads = make(map[string]chan struct{})
	cfg.ConfigReloads[utils.CDRE] = make(chan struct{}, 1)
	cfg.ConfigReloads[utils.CDRE] <- struct{}{} // Unlock the channel

	var cgrJsonCfg *CgrJsonCfg
	if cgrJsonCfg, err = NewCgrJsonCfgFromBytes([]byte(CGRATES_CFG_JSON)); err != nil {
		return
	}
	if err = cfg.loadFromJsonCfg(cgrJsonCfg); err != nil {
		return
	}

	cfg.dfltCdreProfile = cfg.CdreProfiles[utils.MetaDefault].Clone() // So default will stay unique, will have nil pointer in case of no defaults loaded which is an extra check
	// populate default ERs reader
	for _, ersRdr := range cfg.ersCfg.Readers {
		if ersRdr.ID == utils.MetaDefault {
			cfg.dfltEvRdr = ersRdr.Clone()
			break
		}
	}
	dfltFsConnConfig = cfg.fsAgentCfg.EventSocketConns[0] // We leave it crashing here on purpose if no Connection defaults defined
	dfltKamConnConfig = cfg.kamAgentCfg.EvapiConns[0]
	dfltAstConnCfg = cfg.asteriskAgentCfg.AsteriskConns[0]
	dfltLoaderConfig = cfg.loaderCfg[0].Clone()
	err = cfg.checkConfigSanity()
	return
}

func NewCGRConfigFromJsonStringWithDefaults(cfgJsonStr string) (cfg *CGRConfig, err error) {
	cfg, _ = NewDefaultCGRConfig()
	jsnCfg := new(CgrJsonCfg)
	if err = NewRjReaderFromBytes([]byte(cfgJsonStr)).Decode(jsnCfg); err != nil {
		return
	} else if err = cfg.loadFromJsonCfg(jsnCfg); err != nil {
		return
	}
	return
}

// Reads all .json files out of a folder/subfolders and loads them up in lexical order
func NewCGRConfigFromPath(path string) (cfg *CGRConfig, err error) {
	if cfg, err = NewDefaultCGRConfig(); err != nil {
		return
	}
	cfg.ConfigPath = path

	if err = cfg.loadConfigFromPath(path, []func(*CgrJsonCfg) error{cfg.loadFromJsonCfg}); err != nil {
		return
	}
	err = cfg.checkConfigSanity()
	return
}

func isHidden(fileName string) bool {
	if fileName == "." || fileName == ".." {
		return false
	}
	return strings.HasPrefix(fileName, ".")
}

// Holds system configuration, defaults are overwritten with values from config file if found
type CGRConfig struct {
	lks             map[string]*sync.RWMutex
	MaxCallDuration time.Duration // The maximum call duration (used by responder when querying DerivedCharging) // ToDo: export it in configuration file
	DataFolderPath  string        // Path towards data folder, for tests internal usage, not loading out of .json options
	ConfigPath      string        // Path towards config

	// Cache defaults loaded from json and needing clones
	dfltCdreProfile *CdreCfg        // Default cdreConfig profile
	dfltEvRdr       *EventReaderCfg // default event reader

	CdreProfiles map[string]*CdreCfg // Cdre config profiles
	loaderCfg    LoaderSCfgs         // LoaderS configs
	httpAgentCfg HttpAgentCfgs       // HttpAgent configs

	ConfigReloads map[string]chan struct{} // Signals to specific entities that a config reload should occur
	rldChans      map[string]chan struct{} // index here the channels used for reloads

	rpcConns map[string]*RPCConn

	generalCfg       *GeneralCfg       // General config
	dataDbCfg        *DataDbCfg        // Database config
	storDbCfg        *StorDbCfg        // StroreDb config
	tlsCfg           *TlsCfg           // TLS config
	cacheCfg         CacheCfg          // Cache config
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
	supplierSCfg     *SupplierSCfg     // SupplierS config
	sureTaxCfg       *SureTaxCfg       // SureTax config
	dispatcherSCfg   *DispatcherSCfg   // DispatcherS config
	loaderCgrCfg     *LoaderCgrCfg     // LoaderCgr config
	migratorCgrCfg   *MigratorCgrCfg   // MigratorCgr config
	mailerCfg        *MailerCfg        // Mailer config
	analyzerSCfg     *AnalyzerSCfg     // AnalyzerS config
	apier            *ApierCfg
	ersCfg           *ERsCfg
}

var posibleLoaderTypes = utils.NewStringSet([]string{utils.MetaAttributes,
	utils.MetaResources, utils.MetaFilters, utils.MetaStats,
	utils.MetaSuppliers, utils.MetaThresholds, utils.MetaChargers,
	utils.MetaDispatchers, utils.MetaDispatcherHosts})

var possibleReaderTypes = utils.NewStringSet([]string{utils.MetaFileCSV,
	utils.MetaKafkajsonMap, utils.MetaFileXML, utils.MetaSQL, utils.MetaFileFWV,
	utils.MetaPartialCSV, utils.MetaFlatstore, utils.META_NONE})

func (cfg *CGRConfig) LazySanityCheck() {
	for _, cdrePrfl := range cfg.cdrsCfg.OnlineCDRExports {
		if cdreProfile, hasIt := cfg.CdreProfiles[cdrePrfl]; hasIt && (cdreProfile.ExportFormat == utils.MetaS3jsonMap || cdreProfile.ExportFormat == utils.MetaSQSjsonMap) {
			poster := utils.SQSPoster
			if cdreProfile.ExportFormat == utils.MetaS3jsonMap {
				poster = utils.S3Poster
			}
			argsMap := utils.GetUrlRawArguments(cdreProfile.ExportPath)
			for _, arg := range []string{utils.AWSRegion, utils.AWSKey, utils.AWSSecret} {
				if _, has := argsMap[arg]; !has {
					utils.Logger.Warning(fmt.Sprintf("<%s> No %s present for AWS for cdre: <%s>.", poster, arg, cdrePrfl))
				}
			}
		}
	}
}

// Loads from json configuration object, will be used for defaults, config from file and reload, might need lock
func (cfg *CGRConfig) loadFromJsonCfg(jsnCfg *CgrJsonCfg) (err error) {
	// Load sections out of JSON config, stop on error
	for _, loadFunc := range []func(*CgrJsonCfg) error{
		cfg.loadRPCConns,
		cfg.loadGeneralCfg, cfg.loadCacheCfg, cfg.loadListenCfg,
		cfg.loadHTTPCfg, cfg.loadDataDBCfg, cfg.loadStorDBCfg,
		cfg.loadFilterSCfg, cfg.loadRalSCfg, cfg.loadSchedulerCfg,
		cfg.loadCdrsCfg, cfg.loadCdreCfg, cfg.loadSessionSCfg,
		cfg.loadFreeswitchAgentCfg, cfg.loadKamAgentCfg,
		cfg.loadAsteriskAgentCfg, cfg.loadDiameterAgentCfg, cfg.loadRadiusAgentCfg,
		cfg.loadDNSAgentCfg, cfg.loadHttpAgentCfg, cfg.loadAttributeSCfg,
		cfg.loadChargerSCfg, cfg.loadResourceSCfg, cfg.loadStatSCfg,
		cfg.loadThresholdSCfg, cfg.loadSupplierSCfg, cfg.loadLoaderSCfg,
		cfg.loadMailerCfg, cfg.loadSureTaxCfg, cfg.loadDispatcherSCfg,
		cfg.loadLoaderCgrCfg, cfg.loadMigratorCgrCfg, cfg.loadTlsCgrCfg,
		cfg.loadAnalyzerCgrCfg, cfg.loadApierCfg, cfg.loadErsCfg} {
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
		Conns: []*RemoteHost{
			&RemoteHost{
				Address: utils.MetaInternal,
			},
		},
	}
	for key, val := range jsnRPCConns {
		cfg.rpcConns[key] = NewDfltRPCConn()
		if err = cfg.rpcConns[key].loadFromJsonCfg(val); err != nil {
			return
		}
	}
	return
}

// loadGeneralCfg loads the General section of the configuration
func (cfg *CGRConfig) loadGeneralCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnGeneralCfg *GeneralJsonCfg
	if jsnGeneralCfg, err = jsnCfg.GeneralJsonCfg(); err != nil {
		return
	}
	return cfg.generalCfg.loadFromJsonCfg(jsnGeneralCfg)
}

// loadCacheCfg loads the Cache section of the configuration
func (cfg *CGRConfig) loadCacheCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnCacheCfg *CacheJsonCfg
	if jsnCacheCfg, err = jsnCfg.CacheJsonCfg(); err != nil {
		return
	}
	return cfg.cacheCfg.loadFromJsonCfg(jsnCacheCfg)
}

// loadListenCfg loads the Listen section of the configuration
func (cfg *CGRConfig) loadListenCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnListenCfg *ListenJsonCfg
	if jsnListenCfg, err = jsnCfg.ListenJsonCfg(); err != nil {
		return
	}
	return cfg.listenCfg.loadFromJsonCfg(jsnListenCfg)
}

// loadHTTPCfg loads the Http section of the configuration
func (cfg *CGRConfig) loadHTTPCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnHTTPCfg *HTTPJsonCfg
	if jsnHTTPCfg, err = jsnCfg.HttpJsonCfg(); err != nil {
		return
	}
	return cfg.httpCfg.loadFromJsonCfg(jsnHTTPCfg)
}

// loadDataDBCfg loads the DataDB section of the configuration
func (cfg *CGRConfig) loadDataDBCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnDataDbCfg *DbJsonCfg
	if jsnDataDbCfg, err = jsnCfg.DbJsonCfg(DATADB_JSN); err != nil {
		return
	}
	if err = cfg.dataDbCfg.loadFromJsonCfg(jsnDataDbCfg); err != nil {
		return
	}
	// in case of internalDB we need to disable the cache
	// so we enforce it here
	if cfg.dataDbCfg.DataDbType == utils.INTERNAL {
		zeroLimit := &CacheParamCfg{Limit: 0,
			TTL: time.Duration(0), StaticTTL: false, Precache: false}
		disabledCache := CacheCfg{
			utils.CacheDestinations:            zeroLimit,
			utils.CacheReverseDestinations:     zeroLimit,
			utils.CacheRatingPlans:             zeroLimit,
			utils.CacheRatingProfiles:          zeroLimit,
			utils.CacheActions:                 zeroLimit,
			utils.CacheActionPlans:             zeroLimit,
			utils.CacheAccountActionPlans:      zeroLimit,
			utils.CacheActionTriggers:          zeroLimit,
			utils.CacheSharedGroups:            zeroLimit,
			utils.CacheTimings:                 zeroLimit,
			utils.CacheResourceProfiles:        zeroLimit,
			utils.CacheResources:               zeroLimit,
			utils.CacheEventResources:          zeroLimit,
			utils.CacheStatQueueProfiles:       zeroLimit,
			utils.CacheStatQueues:              zeroLimit,
			utils.CacheThresholdProfiles:       zeroLimit,
			utils.CacheThresholds:              zeroLimit,
			utils.CacheFilters:                 zeroLimit,
			utils.CacheSupplierProfiles:        zeroLimit,
			utils.CacheAttributeProfiles:       zeroLimit,
			utils.CacheChargerProfiles:         zeroLimit,
			utils.CacheDispatcherProfiles:      zeroLimit,
			utils.CacheDispatcherHosts:         zeroLimit,
			utils.CacheResourceFilterIndexes:   zeroLimit,
			utils.CacheStatFilterIndexes:       zeroLimit,
			utils.CacheThresholdFilterIndexes:  zeroLimit,
			utils.CacheSupplierFilterIndexes:   zeroLimit,
			utils.CacheAttributeFilterIndexes:  zeroLimit,
			utils.CacheChargerFilterIndexes:    zeroLimit,
			utils.CacheDispatcherFilterIndexes: zeroLimit,
			utils.CacheDispatcherRoutes:        zeroLimit,
			utils.CacheRPCResponses:            zeroLimit,
			utils.CacheLoadIDs:                 zeroLimit,
			utils.CacheDiameterMessages: &CacheParamCfg{Limit: -1,
				TTL: time.Duration(3 * time.Hour), StaticTTL: false},
			utils.CacheClosedSessions: &CacheParamCfg{Limit: -1,
				TTL: time.Duration(10 * time.Second), StaticTTL: false},
			utils.CacheRPCConnections: &CacheParamCfg{Limit: -1,
				TTL: time.Duration(0), StaticTTL: false},
		}
		cfg.cacheCfg = disabledCache
	}
	return

}

// loadStorDBCfg loads the StorDB section of the configuration
func (cfg *CGRConfig) loadStorDBCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnDataDbCfg *DbJsonCfg
	if jsnDataDbCfg, err = jsnCfg.DbJsonCfg(STORDB_JSN); err != nil {
		return
	}
	return cfg.storDbCfg.loadFromJsonCfg(jsnDataDbCfg)
}

// loadFilterSCfg loads the FilterS section of the configuration
func (cfg *CGRConfig) loadFilterSCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnFilterSCfg *FilterSJsonCfg
	if jsnFilterSCfg, err = jsnCfg.FilterSJsonCfg(); err != nil {
		return
	}
	return cfg.filterSCfg.loadFromJsonCfg(jsnFilterSCfg)
}

// loadRalSCfg loads the RalS section of the configuration
func (cfg *CGRConfig) loadRalSCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnRALsCfg *RalsJsonCfg
	if jsnRALsCfg, err = jsnCfg.RalsJsonCfg(); err != nil {
		return
	}
	return cfg.ralsCfg.loadFromJsonCfg(jsnRALsCfg)
}

// loadSchedulerCfg loads the Scheduler section of the configuration
func (cfg *CGRConfig) loadSchedulerCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnSchedCfg *SchedulerJsonCfg
	if jsnSchedCfg, err = jsnCfg.SchedulerJsonCfg(); err != nil {
		return
	}
	return cfg.schedulerCfg.loadFromJsonCfg(jsnSchedCfg)
}

// loadCdrsCfg loads the Cdrs section of the configuration
func (cfg *CGRConfig) loadCdrsCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnCdrsCfg *CdrsJsonCfg
	if jsnCdrsCfg, err = jsnCfg.CdrsJsonCfg(); err != nil {
		return
	}
	return cfg.cdrsCfg.loadFromJsonCfg(jsnCdrsCfg)
}

// loadCdreCfg loads the Cdre section of the configuration
func (cfg *CGRConfig) loadCdreCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnCdreCfg map[string]*CdreJsonCfg
	if jsnCdreCfg, err = jsnCfg.CdreJsonCfgs(); err != nil {
		return
	}
	if jsnCdreCfg != nil {
		for profileName, jsnCdre1Cfg := range jsnCdreCfg {
			if _, hasProfile := cfg.CdreProfiles[profileName]; !hasProfile { // New profile, create before loading from json
				cfg.CdreProfiles[profileName] = new(CdreCfg)
				if profileName != utils.MetaDefault {
					cfg.CdreProfiles[profileName] = cfg.dfltCdreProfile.Clone() // Clone default so we do not inherit pointers
				}
			}
			if err = cfg.CdreProfiles[profileName].loadFromJsonCfg(jsnCdre1Cfg, cfg.generalCfg.RSRSep); err != nil { // Update the existing profile with content from json config
				return
			}
		}
	}
	return
}

// loadSessionSCfg loads the SessionS section of the configuration
func (cfg *CGRConfig) loadSessionSCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnSessionSCfg *SessionSJsonCfg
	if jsnSessionSCfg, err = jsnCfg.SessionSJsonCfg(); err != nil {
		return
	}
	return cfg.sessionSCfg.loadFromJsonCfg(jsnSessionSCfg)
}

// loadFreeswitchAgentCfg loads the FreeswitchAgent section of the configuration
func (cfg *CGRConfig) loadFreeswitchAgentCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnSmFsCfg *FreeswitchAgentJsonCfg
	if jsnSmFsCfg, err = jsnCfg.FreeswitchAgentJsonCfg(); err != nil {
		return
	}
	return cfg.fsAgentCfg.loadFromJsonCfg(jsnSmFsCfg)
}

// loadKamAgentCfg loads the KamAgent section of the configuration
func (cfg *CGRConfig) loadKamAgentCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnKamAgentCfg *KamAgentJsonCfg
	if jsnKamAgentCfg, err = jsnCfg.KamAgentJsonCfg(); err != nil {
		return
	}
	return cfg.kamAgentCfg.loadFromJsonCfg(jsnKamAgentCfg)
}

// loadAsteriskAgentCfg loads the AsteriskAgent section of the configuration
func (cfg *CGRConfig) loadAsteriskAgentCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnSMAstCfg *AsteriskAgentJsonCfg
	if jsnSMAstCfg, err = jsnCfg.AsteriskAgentJsonCfg(); err != nil {
		return
	}
	return cfg.asteriskAgentCfg.loadFromJsonCfg(jsnSMAstCfg)
}

// loadDiameterAgentCfg loads the DiameterAgent section of the configuration
func (cfg *CGRConfig) loadDiameterAgentCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnDACfg *DiameterAgentJsonCfg
	if jsnDACfg, err = jsnCfg.DiameterAgentJsonCfg(); err != nil {
		return
	}
	return cfg.diameterAgentCfg.loadFromJsonCfg(jsnDACfg, cfg.generalCfg.RSRSep)
}

// loadRadiusAgentCfg loads the RadiusAgent section of the configuration
func (cfg *CGRConfig) loadRadiusAgentCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnRACfg *RadiusAgentJsonCfg
	if jsnRACfg, err = jsnCfg.RadiusAgentJsonCfg(); err != nil {
		return
	}
	return cfg.radiusAgentCfg.loadFromJsonCfg(jsnRACfg, cfg.generalCfg.RSRSep)
}

// loadDNSAgentCfg loads the DNSAgent section of the configuration
func (cfg *CGRConfig) loadDNSAgentCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnDNSCfg *DNSAgentJsonCfg
	if jsnDNSCfg, err = jsnCfg.DNSAgentJsonCfg(); err != nil {
		return
	}
	return cfg.dnsAgentCfg.loadFromJsonCfg(jsnDNSCfg, cfg.generalCfg.RSRSep)
}

// loadHttpAgentCfg loads the HttpAgent section of the configuration
func (cfg *CGRConfig) loadHttpAgentCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnHttpAgntCfg *[]*HttpAgentJsonCfg
	if jsnHttpAgntCfg, err = jsnCfg.HttpAgentJsonCfg(); err != nil {
		return
	}
	return cfg.httpAgentCfg.loadFromJsonCfg(jsnHttpAgntCfg, cfg.generalCfg.RSRSep)
}

// loadAttributeSCfg loads the AttributeS section of the configuration
func (cfg *CGRConfig) loadAttributeSCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnAttributeSCfg *AttributeSJsonCfg
	if jsnAttributeSCfg, err = jsnCfg.AttributeServJsonCfg(); err != nil {
		return
	}
	return cfg.attributeSCfg.loadFromJsonCfg(jsnAttributeSCfg)
}

// loadChargerSCfg loads the ChargerS section of the configuration
func (cfg *CGRConfig) loadChargerSCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnChargerSCfg *ChargerSJsonCfg
	if jsnChargerSCfg, err = jsnCfg.ChargerServJsonCfg(); err != nil {
		return
	}
	return cfg.chargerSCfg.loadFromJsonCfg(jsnChargerSCfg)
}

// loadResourceSCfg loads the ResourceS section of the configuration
func (cfg *CGRConfig) loadResourceSCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnRLSCfg *ResourceSJsonCfg
	if jsnRLSCfg, err = jsnCfg.ResourceSJsonCfg(); err != nil {
		return
	}
	return cfg.resourceSCfg.loadFromJsonCfg(jsnRLSCfg)
}

// loadStatSCfg loads the StatS section of the configuration
func (cfg *CGRConfig) loadStatSCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnStatSCfg *StatServJsonCfg
	if jsnStatSCfg, err = jsnCfg.StatSJsonCfg(); err != nil {
		return
	}
	return cfg.statsCfg.loadFromJsonCfg(jsnStatSCfg)
}

// loadThresholdSCfg loads the ThresholdS section of the configuration
func (cfg *CGRConfig) loadThresholdSCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnThresholdSCfg *ThresholdSJsonCfg
	if jsnThresholdSCfg, err = jsnCfg.ThresholdSJsonCfg(); err != nil {
		return
	}
	return cfg.thresholdSCfg.loadFromJsonCfg(jsnThresholdSCfg)
}

// loadSupplierSCfg loads the SupplierS section of the configuration
func (cfg *CGRConfig) loadSupplierSCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnSupplierSCfg *SupplierSJsonCfg
	if jsnSupplierSCfg, err = jsnCfg.SupplierSJsonCfg(); err != nil {
		return
	}
	return cfg.supplierSCfg.loadFromJsonCfg(jsnSupplierSCfg)
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
			loadSCfgp.loadFromJsonCfg(profile, cfg.generalCfg.RSRSep)
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
	return cfg.mailerCfg.loadFromJsonCfg(jsnMailerCfg)
}

// loadSureTaxCfg loads the SureTax section of the configuration
func (cfg *CGRConfig) loadSureTaxCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnSureTaxCfg *SureTaxJsonCfg
	if jsnSureTaxCfg, err = jsnCfg.SureTaxJsonCfg(); err != nil {
		return
	}
	return cfg.sureTaxCfg.loadFromJsonCfg(jsnSureTaxCfg)
}

// loadDispatcherSCfg loads the DispatcherS section of the configuration
func (cfg *CGRConfig) loadDispatcherSCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnDispatcherSCfg *DispatcherSJsonCfg
	if jsnDispatcherSCfg, err = jsnCfg.DispatcherSJsonCfg(); err != nil {
		return
	}
	return cfg.dispatcherSCfg.loadFromJsonCfg(jsnDispatcherSCfg)
}

// loadLoaderCgrCfg loads the Loader section of the configuration
func (cfg *CGRConfig) loadLoaderCgrCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnLoaderCgrCfg *LoaderCfgJson
	if jsnLoaderCgrCfg, err = jsnCfg.LoaderCfgJson(); err != nil {
		return
	}
	return cfg.loaderCgrCfg.loadFromJsonCfg(jsnLoaderCgrCfg)
}

// loadMigratorCgrCfg loads the Migrator section of the configuration
func (cfg *CGRConfig) loadMigratorCgrCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnMigratorCgrCfg *MigratorCfgJson
	if jsnMigratorCgrCfg, err = jsnCfg.MigratorCfgJson(); err != nil {
		return
	}
	return cfg.migratorCgrCfg.loadFromJsonCfg(jsnMigratorCgrCfg)
}

// loadTlsCgrCfg loads the Tls section of the configuration
func (cfg *CGRConfig) loadTlsCgrCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnTlsCgrCfg *TlsJsonCfg
	if jsnTlsCgrCfg, err = jsnCfg.TlsCfgJson(); err != nil {
		return
	}
	return cfg.tlsCfg.loadFromJsonCfg(jsnTlsCgrCfg)
}

// loadAnalyzerCgrCfg loads the Analyzer section of the configuration
func (cfg *CGRConfig) loadAnalyzerCgrCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnAnalyzerCgrCfg *AnalyzerSJsonCfg
	if jsnAnalyzerCgrCfg, err = jsnCfg.AnalyzerCfgJson(); err != nil {
		return
	}
	return cfg.analyzerSCfg.loadFromJsonCfg(jsnAnalyzerCgrCfg)
}

// loadApierCfg loads the Apier section of the configuration
func (cfg *CGRConfig) loadApierCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnApierCfg *ApierJsonCfg
	if jsnApierCfg, err = jsnCfg.ApierCfgJson(); err != nil {
		return
	}
	return cfg.apier.loadFromJsonCfg(jsnApierCfg)
}

// loadErsCfg loads the Ers section of the configuration
func (cfg *CGRConfig) loadErsCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnERsCfg *ERsJsonCfg
	if jsnERsCfg, err = jsnCfg.ERsJsonCfg(); err != nil {
		return
	}
	return cfg.ersCfg.loadFromJsonCfg(jsnERsCfg, cfg.generalCfg.RSRSep, cfg.dfltEvRdr)
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

// SupplierSCfg returns the config for SupplierS
func (cfg *CGRConfig) SupplierSCfg() *SupplierSCfg {
	cfg.lks[SupplierSJson].Lock()
	defer cfg.lks[SupplierSJson].Unlock()
	return cfg.supplierSCfg
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

// HttpAgentCfg returns the config for HttpAgent
func (cfg *CGRConfig) HttpAgentCfg() []*HttpAgentCfg {
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
func (cfg *CGRConfig) CacheCfg() CacheCfg {
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

// TlsCfg returns the config for Tls
func (cfg *CGRConfig) TlsCfg() *TlsCfg {
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

// RPCConns reads the RPCConns configuration
func (cfg *CGRConfig) RPCConns() map[string]*RPCConn {
	cfg.lks[RPCConnsJsonName].RLock()
	defer cfg.lks[RPCConnsJsonName].RUnlock()
	return cfg.rpcConns
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

// ToDo: move this structure in utils as is used in other packages
type StringWithArgDispatcher struct {
	*utils.ArgDispatcher
	utils.TenantArg
	Section string
}

//V1GetConfigSection will retrieve from CGRConfig a section
func (cfg *CGRConfig) V1GetConfigSection(args *StringWithArgDispatcher, reply *map[string]interface{}) (err error) {
	var jsonString string
	switch args.Section {
	case GENERAL_JSN:
		jsonString = utils.ToJSON(cfg.GeneralCfg())
	case DATADB_JSN:
		jsonString = utils.ToJSON(cfg.DataDbCfg())
	case STORDB_JSN:
		jsonString = utils.ToJSON(cfg.StorDbCfg())
	case TlsCfgJson:
		jsonString = utils.ToJSON(cfg.TlsCfg())
	case CACHE_JSN:
		jsonString = utils.ToJSON(cfg.CacheCfg())
	case LISTEN_JSN:
		jsonString = utils.ToJSON(cfg.ListenCfg())
	case HTTP_JSN:
		jsonString = utils.ToJSON(cfg.HTTPCfg())
	case FilterSjsn:
		jsonString = utils.ToJSON(cfg.FilterSCfg())
	case RALS_JSN:
		jsonString = utils.ToJSON(cfg.RalsCfg())
	case SCHEDULER_JSN:
		jsonString = utils.ToJSON(cfg.SchedulerCfg())
	case CDRS_JSN:
		jsonString = utils.ToJSON(cfg.CdrsCfg())
	case SessionSJson:
		jsonString = utils.ToJSON(cfg.SessionSCfg())
	case FreeSWITCHAgentJSN:
		jsonString = utils.ToJSON(cfg.FsAgentCfg())
	case KamailioAgentJSN:
		jsonString = utils.ToJSON(cfg.KamAgentCfg())
	case AsteriskAgentJSN:
		jsonString = utils.ToJSON(cfg.AsteriskAgentCfg())
	case DA_JSN:
		jsonString = utils.ToJSON(cfg.DiameterAgentCfg())
	case RA_JSN:
		jsonString = utils.ToJSON(cfg.RadiusAgentCfg())
	case DNSAgentJson:
		jsonString = utils.ToJSON(cfg.DNSAgentCfg())
	case ATTRIBUTE_JSN:
		jsonString = utils.ToJSON(cfg.AttributeSCfg())
	case ChargerSCfgJson:
		jsonString = utils.ToJSON(cfg.ChargerSCfg())
	case RESOURCES_JSON:
		jsonString = utils.ToJSON(cfg.ResourceSCfg())
	case STATS_JSON:
		jsonString = utils.ToJSON(cfg.StatSCfg())
	case THRESHOLDS_JSON:
		jsonString = utils.ToJSON(cfg.ThresholdSCfg())
	case SupplierSJson:
		jsonString = utils.ToJSON(cfg.SupplierSCfg())
	case SURETAX_JSON:
		jsonString = utils.ToJSON(cfg.SureTaxCfg())
	case DispatcherSJson:
		jsonString = utils.ToJSON(cfg.DispatcherSCfg())
	case LoaderJson:
		jsonString = utils.ToJSON(cfg.LoaderCfg())
	case CgrLoaderCfgJson:
		jsonString = utils.ToJSON(cfg.LoaderCgrCfg())
	case CgrMigratorCfgJson:
		jsonString = utils.ToJSON(cfg.MigratorCgrCfg())
	case ApierS:
		jsonString = utils.ToJSON(cfg.ApierCfg())
	case CDRE_JSN:
		jsonString = utils.ToJSON(cfg.CdreProfiles)
	case ERsJson:
		jsonString = utils.ToJSON(cfg.ERsCfg())
	case RPCConnsJsonName:
		jsonString = utils.ToJSON(cfg.RPCConns())
	default:
		return errors.New("Invalid section")
	}
	json.Unmarshal([]byte(jsonString), reply)
	return
}

type ConfigReloadWithArgDispatcher struct {
	*utils.ArgDispatcher
	utils.TenantArg
	Path    string
	Section string
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

// V1ReloadConfigFromPath reloads the configuration
func (cfg *CGRConfig) V1ReloadConfigFromPath(args *ConfigReloadWithArgDispatcher, reply *string) (err error) {
	if missing := utils.MissingStructFields(args, []string{"Path"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err = cfg.loadCfgWithLocks(args.Path, args.Section); err != nil {
		return
	}
	//  lock all sections
	cfg.rLockSections()

	err = cfg.checkConfigSanity()

	cfg.rUnlockSections() // unlock before checking the error

	if err != nil {
		return
	}

	if args.Section == utils.EmptyString || args.Section == utils.MetaAll {
		err = cfg.reloadSections(sortedCfgSections...)
	} else {
		err = cfg.reloadSections(args.Section)
	}
	if err != nil {
		return
	}
	*reply = utils.OK
	return
}

func (cfg *CGRConfig) getLoadFunctions() map[string]func(*CgrJsonCfg) error {
	return map[string]func(*CgrJsonCfg) error{
		GENERAL_JSN:        cfg.loadGeneralCfg,
		DATADB_JSN:         cfg.loadDataDBCfg,
		STORDB_JSN:         cfg.loadStorDBCfg,
		LISTEN_JSN:         cfg.loadListenCfg,
		TlsCfgJson:         cfg.loadTlsCgrCfg,
		HTTP_JSN:           cfg.loadHTTPCfg,
		SCHEDULER_JSN:      cfg.loadSchedulerCfg,
		CACHE_JSN:          cfg.loadCacheCfg,
		FilterSjsn:         cfg.loadFilterSCfg,
		RALS_JSN:           cfg.loadRalSCfg,
		CDRS_JSN:           cfg.loadCdrsCfg,
		CDRE_JSN:           cfg.loadCdreCfg,
		ERsJson:            cfg.loadErsCfg,
		SessionSJson:       cfg.loadSessionSCfg,
		AsteriskAgentJSN:   cfg.loadAsteriskAgentCfg,
		FreeSWITCHAgentJSN: cfg.loadFreeswitchAgentCfg,
		KamailioAgentJSN:   cfg.loadKamAgentCfg,
		DA_JSN:             cfg.loadDiameterAgentCfg,
		RA_JSN:             cfg.loadRadiusAgentCfg,
		HttpAgentJson:      cfg.loadHttpAgentCfg,
		DNSAgentJson:       cfg.loadDNSAgentCfg,
		ATTRIBUTE_JSN:      cfg.loadAttributeSCfg,
		ChargerSCfgJson:    cfg.loadChargerSCfg,
		RESOURCES_JSON:     cfg.loadResourceSCfg,
		STATS_JSON:         cfg.loadStatSCfg,
		THRESHOLDS_JSON:    cfg.loadThresholdSCfg,
		SupplierSJson:      cfg.loadSupplierSCfg,
		LoaderJson:         cfg.loadLoaderSCfg,
		MAILER_JSN:         cfg.loadMailerCfg,
		SURETAX_JSON:       cfg.loadSureTaxCfg,
		CgrLoaderCfgJson:   cfg.loadLoaderCgrCfg,
		CgrMigratorCfgJson: cfg.loadMigratorCgrCfg,
		DispatcherSJson:    cfg.loadDispatcherSCfg,
		AnalyzerCfgJson:    cfg.loadAnalyzerCgrCfg,
		ApierS:             cfg.loadApierCfg,
		RPCConnsJsonName:   cfg.loadRPCConns,
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
	return cfg.loadConfigFromPath(path, loadFuncs)
}

func (*CGRConfig) loadConfigFromReader(rdr io.Reader, loadFuncs []func(jsnCfg *CgrJsonCfg) error) (err error) {
	jsnCfg := new(CgrJsonCfg)
	var rjr *rjReader
	if rjr, err = NewRjReader(rdr); err != nil {
		return
	}
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
func (cfg *CGRConfig) loadConfigFromPath(path string, loadFuncs []func(jsnCfg *CgrJsonCfg) error) (err error) {
	if isUrl(path) {
		return cfg.loadConfigFromHTTP(path, loadFuncs) // prefix protocol
	}
	var fi os.FileInfo
	if fi, err = os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return utils.ErrPathNotReachable(path)
		}
		return
	} else if !fi.IsDir() && path != utils.CONFIG_PATH { // If config dir defined, needs to exist, not checking for default
		return fmt.Errorf("path: %s not a directory", path)
	}
	if fi.IsDir() {
		return cfg.loadConfigFromFolder(path, loadFuncs)
	}
	return
}

func (cfg *CGRConfig) loadConfigFromFolder(cfgDir string, loadFuncs []func(jsnCfg *CgrJsonCfg) error) (err error) {
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
			var cfgFile *os.File
			cfgFile, werr = os.Open(jsonFilePath)
			if werr != nil {
				return
			}

			werr = cfg.loadConfigFromReader(cfgFile, loadFuncs)
			cfgFile.Close()
			if werr != nil {
				werr = fmt.Errorf("file <%s>:%s", jsonFilePath, werr.Error())
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

func (cfg *CGRConfig) loadConfigFromHTTP(urlPaths string, loadFuncs []func(jsnCfg *CgrJsonCfg) error) (err error) {
	for _, urlPath := range strings.Split(urlPaths, utils.INFIELD_SEP) {
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
		err = cfg.loadConfigFromReader(cfgReq.Body, loadFuncs)
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

// JSONReloadWithArgDispatcher the API params for V1ReloadConfigFromJSON
type JSONReloadWithArgDispatcher struct {
	*utils.ArgDispatcher
	utils.TenantArg
	JSON map[string]interface{}
}

// V1ReloadConfigFromJSON reloads the sections of configz
func (cfg *CGRConfig) V1ReloadConfigFromJSON(args *JSONReloadWithArgDispatcher, reply *string) (err error) {
	if len(args.JSON) == 0 {
		*reply = utils.OK
		return
	}
	sections := make([]string, 0, len(args.JSON))
	for section := range args.JSON {
		sections = append(sections, section)
	}

	var b []byte
	if b, err = json.Marshal(args.JSON); err != nil {
		return
	}

	if err = cfg.loadCfgFromJSONWithLocks(bytes.NewBuffer(b), sections); err != nil {
		return
	}

	//  lock all sections
	cfg.rLockSections()

	err = cfg.checkConfigSanity()

	cfg.rUnlockSections() // unlock before checking the error

	err = cfg.reloadSections(sections...)
	if err != nil {
		return
	}
	*reply = utils.OK
	return
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
	return cfg.loadConfigFromReader(rdr, loadFuncs)
}

func (cfg *CGRConfig) reloadSections(sections ...string) (err error) {
	subsystemsThatNeedDataDB := utils.NewStringSet([]string{DATADB_JSN, SCHEDULER_JSN,
		RALS_JSN, CDRS_JSN, SessionSJson, ATTRIBUTE_JSN,
		ChargerSCfgJson, RESOURCES_JSON, STATS_JSON, THRESHOLDS_JSON,
		SupplierSJson, LoaderJson, DispatcherSJson})
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
		default:
			return fmt.Errorf("Invalid section: <%s>", section)
		case GENERAL_JSN: // nothing to reload
		case RPCConnsJsonName: // nothing to reload
			cfg.rldChans[RPCConnsJsonName] <- struct{}{}
		case DATADB_JSN: // reloaded before
		case STORDB_JSN: // reloaded before
		case LISTEN_JSN:
		case TlsCfgJson: // nothing to reload
		case HTTP_JSN:
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
		case SupplierSJson:
			cfg.rldChans[SupplierSJson] <- struct{}{}
		case LoaderJson:
			cfg.rldChans[LoaderJson] <- struct{}{}
		case DispatcherSJson:
			cfg.rldChans[DispatcherSJson] <- struct{}{}
		case AnalyzerCfgJson:
		case ApierS:
			cfg.rldChans[ApierS] <- struct{}{}
		}
		return
	}
	return
}

func (cfg *CGRConfig) AsMapInterface(separator string) map[string]interface{} {
	rpcConns := make(map[string]map[string]interface{}, len(cfg.rpcConns))
	for key, val := range cfg.rpcConns {
		rpcConns[key] = val.AsMapInterface()
	}

	cdreProfiles := make(map[string]map[string]interface{})
	for key, val := range cfg.CdreProfiles {
		cdreProfiles[key] = val.AsMapInterface(separator)
	}

	loaderCfg := make([]map[string]interface{}, len(cfg.loaderCfg))
	for i, item := range cfg.loaderCfg {
		loaderCfg[i] = item.AsMapInterface(separator)
	}

	httpAgentCfg := make([]map[string]interface{}, len(cfg.httpAgentCfg))
	for i, item := range cfg.httpAgentCfg {
		httpAgentCfg[i] = item.AsMapInterface(separator)
	}

	return map[string]interface{}{

		utils.CdreProfiles:     cdreProfiles,
		utils.LoaderCfg:        loaderCfg,
		utils.HttpAgentCfg:     httpAgentCfg,
		utils.RpcConns:         rpcConns,
		utils.GeneralCfg:       cfg.generalCfg.AsMapInterface(),
		utils.DataDbCfg:        cfg.dataDbCfg.AsMapInterface(),
		utils.StorDbCfg:        cfg.storDbCfg.AsMapInterface(),
		utils.TlsCfg:           cfg.tlsCfg.AsMapInterface(),
		utils.CacheCfg:         cfg.cacheCfg.AsMapInterface(),
		utils.ListenCfg:        cfg.listenCfg.AsMapInterface(),
		utils.HttpCfg:          cfg.httpCfg.AsMapInterface(),
		utils.FilterSCfg:       cfg.filterSCfg.AsMapInterface(),
		utils.RalsCfg:          cfg.ralsCfg.AsMapInterface(),
		utils.SchedulerCfg:     cfg.schedulerCfg.AsMapInterface(),
		utils.CdrsCfg:          cfg.cdrsCfg.AsMapInterface(),
		utils.SessionSCfg:      cfg.sessionSCfg.AsMapInterface(),
		utils.FsAgentCfg:       cfg.fsAgentCfg.AsMapInterface(separator),
		utils.KamAgentCfg:      cfg.kamAgentCfg.AsMapInterface(),
		utils.AsteriskAgentCfg: cfg.asteriskAgentCfg.AsMapInterface(),
		utils.DiameterAgentCfg: cfg.diameterAgentCfg.AsMapInterface(separator),
		utils.RadiusAgentCfg:   cfg.radiusAgentCfg.AsMapInterface(separator),
		utils.DnsAgentCfg:      cfg.dnsAgentCfg.AsMapInterface(separator),
		utils.AttributeSCfg:    cfg.attributeSCfg.AsMapInterface(),
		utils.ChargerSCfg:      cfg.chargerSCfg.AsMapInterface(),
		utils.ResourceSCfg:     cfg.resourceSCfg.AsMapInterface(),
		utils.StatsCfg:         cfg.statsCfg.AsMapInterface(),
		utils.ThresholdSCfg:    cfg.thresholdSCfg.AsMapInterface(),
		utils.SupplierSCfg:     cfg.supplierSCfg.AsMapInterface(),
		utils.SureTaxCfg:       cfg.sureTaxCfg.AsMapInterface(separator),
		utils.DispatcherSCfg:   cfg.dispatcherSCfg.AsMapInterface(),
		utils.LoaderCgrCfg:     cfg.loaderCgrCfg.AsMapInterface(),
		utils.MigratorCgrCfg:   cfg.migratorCgrCfg.AsMapInterface(),
		utils.MailerCfg:        cfg.mailerCfg.AsMapInterface(),
		utils.AnalyzerSCfg:     cfg.analyzerSCfg.AsMapInterface(),
		utils.Apier:            cfg.apier.AsMapInterface(),
		utils.ErsCfg:           cfg.ersCfg.AsMapInterface(separator),
	}
}
