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

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

var (
	cgrCfg *CGRConfig // will be shared

	getDftFsConnCfg  = func() *FsConnCfg { return new(FsConnCfg) }             // returns default FreeSWITCH Connection configuration, built out of json default configuration
	getDftKamConnCfg = func() *KamConnCfg { return new(KamConnCfg) }           // returns default Kamailio Connection configuration
	getDftAstConnCfg = func() *AsteriskConnCfg { return new(AsteriskConnCfg) } // returns default Asterisk Connection configuration

	getDftLoaderCfg = func() *LoaderSCfg { return new(LoaderSCfg) }
	getDftRemHstCfg = func() *RemoteHost { return new(RemoteHost) }

	getDftEvExpCfg = func() *EventExporterCfg { return &EventExporterCfg{Opts: make(map[string]interface{})} }
	getDftEvRdrCfg = func() *EventReaderCfg { return &EventReaderCfg{Opts: make(map[string]interface{})} }
)

func init() {
	cgrCfg = NewDefaultCGRConfig()
	// populate default ERs reader
	for _, rdr := range cgrCfg.ersCfg.Readers {
		if rdr.ID == utils.MetaDefault {
			getDftEvRdrCfg = rdr.Clone
			break
		}
	}

	// populate default EEs exporter
	for _, exp := range cgrCfg.eesCfg.Exporters {
		if exp.ID == utils.MetaDefault {
			getDftEvExpCfg = exp.Clone
			break
		}
	}

	getDftFsConnCfg = cgrCfg.fsAgentCfg.EventSocketConns[0].Clone // We leave it crashing here on purpose if no Connection defaults defined
	getDftKamConnCfg = cgrCfg.kamAgentCfg.EvapiConns[0].Clone
	getDftAstConnCfg = cgrCfg.asteriskAgentCfg.AsteriskConns[0].Clone
	getDftLoaderCfg = cgrCfg.loaderCfg[0].Clone
	getDftRemHstCfg = cgrCfg.rpcConns[utils.MetaLocalHost].Conns[0].Clone
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
	cfg = &CGRConfig{
		DataFolderPath: "/usr/share/cgrates/",

		rpcConns:   make(RPCConns),
		templates:  make(FCTemplates),
		generalCfg: &GeneralCfg{NodeID: utils.UUIDSha1Prefix()},
		dataDbCfg: &DataDbCfg{
			Items: make(map[string]*ItemOpt),
			Opts:  make(map[string]interface{}),
		},
		storDbCfg: &StorDbCfg{
			Items: make(map[string]*ItemOpt),
			Opts:  make(map[string]interface{}),
		},
		tlsCfg:     new(TLSCfg),
		cacheCfg:   &CacheCfg{Partitions: make(map[string]*CacheParamCfg)},
		listenCfg:  new(ListenCfg),
		httpCfg:    &HTTPCfg{ClientOpts: make(map[string]interface{})},
		filterSCfg: new(FilterSCfg),
		cdrsCfg: &CdrsCfg{Opts: &CdrsOpts{
			Accounts:   []*utils.DynamicBoolOpt{},
			Attributes: []*utils.DynamicBoolOpt{},
			Chargers:   []*utils.DynamicBoolOpt{},
			Export:     []*utils.DynamicBoolOpt{},
			Rates:      []*utils.DynamicBoolOpt{},
			Stats:      []*utils.DynamicBoolOpt{},
			Thresholds: []*utils.DynamicBoolOpt{},
		}},
		analyzerSCfg: new(AnalyzerSCfg),
		sessionSCfg: &SessionSCfg{
			STIRCfg:      new(STIRcfg),
			DefaultUsage: make(map[string]time.Duration),
		},
		fsAgentCfg:       new(FsAgentCfg),
		kamAgentCfg:      new(KamAgentCfg),
		asteriskAgentCfg: new(AsteriskAgentCfg),
		diameterAgentCfg: new(DiameterAgentCfg),
		radiusAgentCfg: &RadiusAgentCfg{
			ClientDictionaries: make(map[string]string),
			ClientSecrets:      make(map[string]string),
		},
		dnsAgentCfg: new(DNSAgentCfg),
		attributeSCfg: &AttributeSCfg{Opts: &AttributesOpts{
			AttributeIDs: []*utils.DynamicStringSliceOpt{},
			ProcessRuns:  []*utils.DynamicIntOpt{},
			ProfileRuns:  []*utils.DynamicIntOpt{},
		}},
		chargerSCfg: new(ChargerSCfg),
		resourceSCfg: &ResourceSConfig{Opts: &ResourcesOpts{
			UsageID:  []*utils.DynamicStringOpt{},
			UsageTTL: []*utils.DynamicDurationOpt{},
			Units:    []*utils.DynamicFloat64Opt{},
		}},
		statsCfg: &StatSCfg{Opts: &StatsOpts{
			StatIDs: []*utils.DynamicStringSliceOpt{},
		}},
		thresholdSCfg: &ThresholdSCfg{Opts: &ThresholdsOpts{
			ThresholdIDs: []*utils.DynamicStringSliceOpt{},
		}},
		routeSCfg: &RouteSCfg{Opts: &RoutesOpts{
			Context:      []*utils.DynamicStringOpt{},
			IgnoreErrors: []*utils.DynamicBoolOpt{},
			MaxCost:      []*utils.DynamicInterfaceOpt{},
			ProfileCount: []*utils.DynamicIntOpt{},
			Limit:        []*utils.DynamicIntOpt{},
			Offset:       []*utils.DynamicIntOpt{},
		}},
		sureTaxCfg:     new(SureTaxCfg),
		dispatcherSCfg: new(DispatcherSCfg),
		registrarCCfg: &RegistrarCCfgs{
			RPC:         &RegistrarCCfg{Hosts: make(map[string][]*RemoteHost)},
			Dispatchers: &RegistrarCCfg{Hosts: make(map[string][]*RemoteHost)},
		},
		loaderCgrCfg: new(LoaderCgrCfg),
		migratorCgrCfg: &MigratorCgrCfg{
			OutDataDBOpts: make(map[string]interface{}),
			OutStorDBOpts: make(map[string]interface{}),
		},
		loaderCfg:    make(LoaderSCfgs, 0),
		httpAgentCfg: make(HTTPAgentCfgs, 0),
		admS:         new(AdminSCfg),
		ersCfg:       new(ERsCfg),
		eesCfg:       &EEsCfg{Cache: make(map[string]*CacheParamCfg)},
		rateSCfg: &RateSCfg{Opts: &RatesOpts{
			RateProfileIDs: []*utils.DynamicStringSliceOpt{},
			StartTime:      []*utils.DynamicStringOpt{},
			Usage:          []*utils.DynamicDecimalBigOpt{},
			IntervalStart:  []*utils.DynamicDecimalBigOpt{},
		}},
		actionSCfg: &ActionSCfg{Opts: &ActionsOpts{
			ActionProfileIDs: []*utils.DynamicStringSliceOpt{},
		}},
		sipAgentCfg: new(SIPAgentCfg),
		configSCfg:  new(ConfigSCfg),
		apiBanCfg:   new(APIBanCfg),
		coreSCfg:    new(CoreSCfg),
		accountSCfg: &AccountSCfg{Opts: &AccountsOpts{
			AccountIDs: []*utils.DynamicStringSliceOpt{},
			Usage:      []*utils.DynamicDecimalBigOpt{},
		}},
		configDBCfg: &ConfigDBCfg{
			Opts: make(map[string]interface{}),
		},

		rldCh:   make(chan string, 1),
		cacheDP: make(utils.MapStorage),
	}
	cfg.sections = newSections(cfg)
	cfg.initChanels()

	var cgrJSONCfg *CgrJsonCfg
	if cgrJSONCfg, err = NewCgrJsonCfgFromBytes(config); err != nil {
		return
	}
	if err = cfg.sections.Load(context.Background(), cgrJSONCfg, cfg); err != nil {
		return
	}
	err = cfg.checkConfigSanity()
	return
}

// NewCGRConfigFromJSONStringWithDefaults returns the given config with the default option loaded
func NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr string) (cfg *CGRConfig, err error) {
	cfg = NewDefaultCGRConfig()
	jsnCfg := new(CgrJsonCfg)
	if err = NewRjReaderFromBytes([]byte(cfgJSONStr)).Decode(jsnCfg); err != nil {
		return
	}
	err = cfg.sections.Load(context.Background(), jsnCfg, cfg)
	return
}

// NewCGRConfigFromPath reads all json files out of a folder/subfolders and loads them up in lexical order
func NewCGRConfigFromPath(ctx *context.Context, path string) (cfg *CGRConfig, err error) {
	cfg = NewDefaultCGRConfig()
	cfg.ConfigPath = path

	if err = loadConfigFromPath(ctx, path, cfg.sections, false, cfg); err != nil {
		return
	}
	err = cfg.checkConfigSanity()
	return
}

// newCGRConfigFromPathWithoutEnv reads all json files out of a folder/subfolders and loads them up in lexical order
// it will not read *env variables and will not checkConfigSanity as it is not needed for configs
func newCGRConfigFromPathWithoutEnv(ctx *context.Context, path string) (cfg *CGRConfig, err error) {
	cfg = NewDefaultCGRConfig()
	cfg.ConfigPath = path

	err = loadConfigFromPath(ctx, path, cfg.sections, true, cfg)
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
	sections       Sections
	rldCh          chan string // index here the channels used for reloads
	lks            map[string]*sync.RWMutex
	db             ConfigDB // to store the last dbConn that executed an config update
	DataFolderPath string   // Path towards data folder, for tests internal usage, not loading out of .json options
	ConfigPath     string   // Path towards config

	loaderCfg    LoaderSCfgs   // LoaderS configs
	httpAgentCfg HTTPAgentCfgs // HttpAgent configs
	rpcConns     RPCConns
	templates    FCTemplates

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
	configDBCfg      *ConfigDBCfg      // ConfigDB conifg

	cacheDP    utils.MapStorage
	cacheDPMux sync.RWMutex
}

var posibleLoaderTypes = utils.NewStringSet([]string{utils.MetaAttributes,
	utils.MetaResources, utils.MetaFilters, utils.MetaStats,
	utils.MetaRoutes, utils.MetaThresholds, utils.MetaChargers,
	utils.MetaDispatchers, utils.MetaDispatcherHosts, utils.MetaRateProfiles,
	utils.MetaAccounts, utils.MetaActionProfiles})

var possibleReaderTypes = utils.NewStringSet([]string{utils.MetaFileCSV,
	utils.MetaKafkajsonMap, utils.MetaFileXML, utils.MetaSQL, utils.MetaFileFWV,
	utils.MetaFileJSON, utils.MetaNone, utils.MetaAMQPjsonMap, utils.MetaS3jsonMap,
	utils.MetaSQSjsonMap, utils.MetaAMQPV1jsonMap, utils.MetaNatsjsonMap})

var possibleExporterTypes = utils.NewStringSet([]string{utils.MetaFileCSV, utils.MetaNone, utils.MetaFileFWV,
	utils.MetaHTTPPost, utils.MetaHTTPjsonMap, utils.MetaAMQPjsonMap, utils.MetaAMQPV1jsonMap, utils.MetaSQSjsonMap,
	utils.MetaKafkajsonMap, utils.MetaS3jsonMap, utils.MetaElastic, utils.MetaVirt, utils.MetaSQL, utils.MetaNatsjsonMap,
	utils.MetaLog})

func (cfg *CGRConfig) AddSection(sec Section) {
	cfg.sections = append(cfg.sections, sec)
	cfg.lks[sec.SName()] = new(sync.RWMutex)
}

func (cfg *CGRConfig) GetAllSectionIDs() (s []string) {
	s = make([]string, 0, len(cfg.sections))
	for _, f := range cfg.sections {
		s = append(s, f.SName())
	}
	return
}

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

// loadConfigDBCfg loads the ConfigDB section of the configuration
func (cfg *CGRConfig) loadConfigDBCfg(ctx *context.Context, jsnCfg ConfigDB) (err error) {
	jsnDBCfg := new(DbJsonCfg)
	if err = jsnCfg.GetSection(ctx, ConfigDBJSON, jsnDBCfg); err != nil {
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
func (cfg *CGRConfig) ConfigDBCfg() *ConfigDBCfg {
	cfg.lks[ConfigDBJSON].Lock()
	defer cfg.lks[ConfigDBJSON].Unlock()
	return cfg.configDBCfg
}

// GetReloadChan returns the reload chanel for the given section
func (cfg *CGRConfig) GetReloadChan() chan string {
	return cfg.rldCh
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

// RLock will read-lock locks with ID.
// User needs to know what he is doing since this can panic
func (cfg *CGRConfig) RLock(sID string) {
	cfg.lks[sID].RLock()
}

// RUnlock will read-unlock locks with ID.
// User needs to know what he is doing since this can panic
func (cfg *CGRConfig) RUnlock(sID string) {
	cfg.lks[sID].RUnlock()
}

// Lock will lock the given section
// User needs to know what he is doing since this can panic
func (cfg *CGRConfig) Lock(sID string) {
	cfg.lks[sID].Lock()
}

// Unlock will unlock the given section
// User needs to know what he is doing since this can panic
func (cfg *CGRConfig) Unlock(sID string) {
	cfg.lks[sID].Unlock()
}

// RLocks will read-lock locks with IDs.
// User needs to know what he is doing since this can panic
func (cfg *CGRConfig) RLocks(lkIDs ...string) {
	for _, lkID := range lkIDs {
		cfg.RLock(lkID)
	}
}

// RUnlocks will read-unlock locks with IDs.
// User needs to know what he is doing since this can panic
func (cfg *CGRConfig) RUnlocks(lkIDs ...string) {
	for _, lkID := range lkIDs {
		cfg.RUnlock(lkID)
	}
}

// LockSections will lock the given sections
// User needs to know what he is doing since this can panic
func (cfg *CGRConfig) LockSections(lkIDs ...string) {
	for _, lkID := range lkIDs {
		cfg.Lock(lkID)
	}
}

// UnlockSections will unlock the given sections
// User needs to know what he is doing since this can panic
func (cfg *CGRConfig) UnlockSections(lkIDs ...string) {
	for _, lkID := range lkIDs {
		cfg.Unlock(lkID)
	}
}

func (cfg *CGRConfig) loadCfgWithLocks(ctx *context.Context, path, section string) (err error) {
	sections := cfg.sections
	if section == utils.EmptyString || section == utils.MetaAll {
		cfg.lockSections()
		defer cfg.unlockSections()
	} else if sec, has := sections.Get(section); !has {
		return fmt.Errorf("Invalid section: <%s> ", section)
	} else {
		cfg.lks[section].Lock()
		defer cfg.lks[section].Unlock()
		sections = Sections{sec}
	}
	return loadConfigFromPath(ctx, path, sections, false, cfg)
}

func loadConfigFromReader(ctx *context.Context, rdr io.Reader, loadFuncs Sections, envOff bool, cfg *CGRConfig) (err error) {
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
	return loadFuncs.Load(ctx, jsnCfg, cfg)
}

// Reads all .json files out of a folder/subfolders and loads them up in lexical order
func loadConfigFromPath(ctx *context.Context, path string, loadFuncs Sections, envOff bool, cfg *CGRConfig) (err error) {
	if utils.IsURL(path) {
		return loadConfigFromHTTP(ctx, path, loadFuncs, cfg) // prefix protocol
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
	return loadConfigFromFolder(ctx, path, loadFuncs, envOff, cfg)
}

func loadConfigFromFolder(ctx *context.Context, cfgDir string, loadFuncs Sections, envOff bool, cfg *CGRConfig) (err error) {
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
			if werr = loadConfigFromFile(ctx, jsonFilePath, loadFuncs, envOff, cfg); werr != nil {
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
func loadConfigFromFile(ctx *context.Context, jsonFilePath string, loadFuncs Sections, envOff bool, cfg *CGRConfig) (err error) {
	var cfgFile *os.File
	cfgFile, err = os.Open(jsonFilePath)
	if err != nil {
		return
	}
	err = loadConfigFromReader(ctx, cfgFile, loadFuncs, envOff, cfg)
	cfgFile.Close()
	if err != nil {
		err = fmt.Errorf("file <%s>:%s", jsonFilePath, err.Error())
	}
	return
}

func loadConfigFromHTTP(ctx *context.Context, urlPaths string, loadFuncs Sections, cfg *CGRConfig) (err error) {
	for _, urlPath := range strings.Split(urlPaths, utils.InfieldSep) {
		if _, err = url.ParseRequestURI(urlPath); err != nil {
			return
		}
		var myClient = &http.Client{
			Timeout: CgrConfig().GeneralCfg().ReplyTimeout,
		}
		var req *http.Request
		if req, err = http.NewRequestWithContext(ctx, "GET", urlPath, nil); err != nil {
			return
		}
		var cfgReq *http.Response
		if cfgReq, err = myClient.Do(req); err != nil {
			return utils.ErrPathNotReachable(urlPath)
		}
		err = loadConfigFromReader(ctx, cfgReq.Body, loadFuncs, false, cfg)
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
	for _, section := range cfg.sections {
		cfg.lks[section.SName()] = new(sync.RWMutex)
	}
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
			cfg.rldCh <- SectionToService[DataDBJSON] // reload datadb before
		}
		if !needsStorDB && subsystemsThatNeedStorDB.Has(section) {
			needsStorDB = true
			cfg.rldCh <- SectionToService[StorDBJSON] // reload stordb before
		}
		if needsDataDB && needsStorDB {
			break
		}
	}
	runtime.Gosched()
	for _, section := range sections {
		if srv := SectionToService[section]; srv != utils.EmptyString &&
			section != DataDBJSON &&
			section != StorDBJSON {
			cfg.rldCh <- srv
		}
	}
}

// AsMapInterface returns the config as a map[string]interface{}
func (cfg *CGRConfig) AsMapInterface(separator string) (mp map[string]interface{}) {
	return cfg.sections.AsMapInterface(separator)
}

// Clone returns a deep copy of CGRConfig
func (cfg *CGRConfig) Clone() (cln *CGRConfig) {
	cln = &CGRConfig{
		DataFolderPath: cfg.DataFolderPath,
		ConfigPath:     cfg.ConfigPath,

		loaderCfg:        *cfg.loaderCfg.Clone(),
		httpAgentCfg:     *cfg.httpAgentCfg.Clone(),
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
		rldCh:            make(chan string),
		cacheDP:          make(utils.MapStorage),
	}
	cln.sections = newSections(cln)
	for _, sec := range cfg.sections[len(cln.sections):] {
		cln.sections = append(cln.sections, sec.CloneSection())
	}
	cln.initChanels()
	return
}

// GetDataProvider returns the config as a data provider interface
func (cfg *CGRConfig) GetDataProvider() utils.MapStorage {
	cfg.cacheDPMux.RLock()
	if len(cfg.cacheDP) < len(cfg.sections) {
		cfg.cacheDP = cfg.AsMapInterface(cfg.GeneralCfg().RSRSep)
	}
	mp := cfg.cacheDP.Clone()
	cfg.cacheDPMux.RUnlock()
	return mp
}
