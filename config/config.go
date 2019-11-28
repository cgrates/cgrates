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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/cgrates/cgrates/utils"
)

var (
	DBDefaults               DbDefaults
	cgrCfg                   *CGRConfig  // will be shared
	dfltFsConnConfig         *FsConnCfg  // Default FreeSWITCH Connection configuration, built out of json default configuration
	dfltKamConnConfig        *KamConnCfg // Default Kamailio Connection configuration
	dfltRemoteHost           *RemoteHost
	dfltAstConnCfg           *AsteriskConnCfg
	dfltLoaderConfig         *LoaderSCfg
	dfltLoaderDataTypeConfig *LoaderDataType
)

func NewDbDefaults() DbDefaults {
	deflt := DbDefaults{
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

type DbDefaults map[string]map[string]string

func (dbDflt DbDefaults) DBName(dbType string, flagInput string) string {
	if flagInput != utils.MetaDynamic {
		return flagInput
	}
	return dbDflt[dbType]["DbName"]
}

func (DbDefaults) DBUser(dbType string, flagInput string) string {
	if flagInput != utils.MetaDynamic {
		return flagInput
	}
	return utils.CGRATES
}

func (DbDefaults) DBHost(dbType string, flagInput string) string {
	if flagInput != utils.MetaDynamic {
		return flagInput
	}
	return utils.LOCALHOST
}

func (dbcfg DbDefaults) DBPort(dbType string, flagInput string) string {
	if flagInput != utils.MetaDynamic {
		return flagInput
	}
	return dbcfg[dbType]["DbPort"]
}

func (dbcfg DbDefaults) DBPass(dbType string, flagInput string) string {
	if flagInput != utils.MetaDynamic {
		return flagInput
	}
	return dbcfg[dbType]["DbPass"]
}

func init() {
	cgrCfg, _ = NewDefaultCGRConfig()
	DBDefaults = NewDbDefaults()
}

// Used to retrieve system configuration from other packages
func CgrConfig() *CGRConfig {
	return cgrCfg
}

// Used to set system configuration from other places
func SetCgrConfig(cfg *CGRConfig) {
	cgrCfg = cfg
}

func NewDefaultCGRConfig() (cfg *CGRConfig, err error) {
	cfg = new(CGRConfig)
	cfg.initChanels()
	cfg.DataFolderPath = "/usr/share/cgrates/"
	cfg.MaxCallDuration = time.Duration(3) * time.Hour // Hardcoded for now

	cfg.rpcConns = make(map[string]*RpcConn)
	cfg.generalCfg = new(GeneralCfg)
	cfg.generalCfg.NodeID = utils.UUIDSha1Prefix()
	cfg.dataDbCfg = new(DataDbCfg)
	cfg.storDbCfg = new(StorDbCfg)
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
	cfg.CdrcProfiles = make(map[string][]*CdrcCfg)
	cfg.analyzerSCfg = new(AnalyzerSCfg)
	cfg.sessionSCfg = new(SessionSCfg)
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
	cfg.ConfigReloads[utils.CDRC] = make(chan struct{}, 1)
	cfg.ConfigReloads[utils.CDRC] <- struct{}{} // Unlock the channel
	cfg.ConfigReloads[utils.CDRE] = make(chan struct{}, 1)
	cfg.ConfigReloads[utils.CDRE] <- struct{}{} // Unlock the channel

	var cgrJsonCfg *CgrJsonCfg
	if cgrJsonCfg, err = NewCgrJsonCfgFromBytes([]byte(CGRATES_CFG_JSON)); err != nil {
		return
	}
	if err = cfg.loadFromJsonCfg(cgrJsonCfg); err != nil {
		return
	}

	cfg.dfltCdreProfile = cfg.CdreProfiles[utils.META_DEFAULT].Clone() // So default will stay unique, will have nil pointer in case of no defaults loaded which is an extra check
	cfg.dfltCdrcProfile = cfg.CdrcProfiles["/var/spool/cgrates/cdrc/in"][0].Clone()
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
	dfltCdrcProfile *CdrcCfg        // Default cdrcConfig profile
	dfltEvRdr       *EventReaderCfg // default event reader

	CdreProfiles map[string]*CdreCfg   // Cdre config profiles
	CdrcProfiles map[string][]*CdrcCfg // Number of CDRC instances running imports, format map[dirPath][]{Configs}
	loaderCfg    LoaderSCfgs           // LoaderS configs
	httpAgentCfg HttpAgentCfgs         // HttpAgent configs

	ConfigReloads map[string]chan struct{} // Signals to specific entities that a config reload should occur
	rldChans      map[string]chan struct{} // index here the channels used for reloads

	rpcConns map[string]*RpcConn

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

var possibleReaderTypes = utils.NewStringSet([]string{utils.MetaFileCSV, utils.MetaKafkajsonMap})

func (cfg *CGRConfig) checkConfigSanity() error {
	// Rater checks
	if cfg.ralsCfg.Enabled && !cfg.dispatcherSCfg.Enabled {
		if !cfg.statsCfg.Enabled {
			for _, connCfg := range cfg.ralsCfg.StatSConns {
				if connCfg.Address == utils.MetaInternal {
					return fmt.Errorf("<%s> not enabled but requested by <%s> component.",
						utils.StatS, utils.RALService)
				}
			}
		}
		if !cfg.thresholdSCfg.Enabled {
			for _, connCfg := range cfg.ralsCfg.ThresholdSConns {
				if connCfg.Address == utils.MetaInternal {
					return fmt.Errorf("<%s> not enabled but requested by <%s> component.",
						utils.ThresholdS, utils.RALService)
				}
			}
		}
	}
	// CDRServer checks
	if cfg.cdrsCfg.Enabled && !cfg.dispatcherSCfg.Enabled {
		if !cfg.chargerSCfg.Enabled {
			for _, conn := range cfg.cdrsCfg.ChargerSConns {
				if conn.Address == utils.MetaInternal {
					return fmt.Errorf("<%s> not enabled but requested by <%s> component.", utils.Chargers, utils.CDRs)
				}
			}
		}
		if !cfg.ralsCfg.Enabled {
			for _, cdrsRaterConn := range cfg.cdrsCfg.RaterConns {
				if cdrsRaterConn.Address == utils.MetaInternal {
					return fmt.Errorf("<%s> not enabled but requested by <%s> component.", utils.RALService, utils.CDRs)
				}
			}
		}
		if !cfg.attributeSCfg.Enabled {
			for _, connCfg := range cfg.cdrsCfg.AttributeSConns {
				if connCfg.Address == utils.MetaInternal {
					return fmt.Errorf("<%s> not enabled but requested by <%s> component.", utils.AttributeS, utils.CDRs)
				}
			}
		}
		if !cfg.statsCfg.Enabled {
			for _, connCfg := range cfg.cdrsCfg.StatSConns {
				if connCfg.Address == utils.MetaInternal {
					return fmt.Errorf("<%s> not enabled but requested by <%s> component.", utils.StatService, utils.CDRs)
				}
			}
		}
		for _, cdrePrfl := range cfg.cdrsCfg.OnlineCDRExports {
			if _, hasIt := cfg.CdreProfiles[cdrePrfl]; !hasIt {
				return fmt.Errorf("<%s> Cannot find CDR export template with ID: <%s>", utils.CDRs, cdrePrfl)
			}
		}
		if !cfg.thresholdSCfg.Enabled {
			for _, connCfg := range cfg.cdrsCfg.ThresholdSConns {
				if connCfg.Address == utils.MetaInternal {
					return fmt.Errorf("%s not enabled but requested by %s component.", utils.ThresholdS, utils.CDRs)
				}
			}
		}
	}
	// CDRC sanity checks
	for _, cdrcCfgs := range cfg.CdrcProfiles {
		for _, cdrcInst := range cdrcCfgs {
			if !cdrcInst.Enabled {
				continue
			}
			if len(cdrcInst.CdrsConns) == 0 {
				return fmt.Errorf("<%s> Instance: %s, %s enabled but no %s defined!", utils.CDRC, cdrcInst.ID, utils.CDRC, utils.CDRs)
			}
			if !cfg.cdrsCfg.Enabled && !cfg.dispatcherSCfg.Enabled {
				for _, conn := range cdrcInst.CdrsConns {
					if conn.Address == utils.MetaInternal {
						return fmt.Errorf("<%s> not enabled but referenced from <%s>", utils.CDRs, utils.CDRC)
					}
				}
			}
			if len(cdrcInst.ContentFields) == 0 {
				return fmt.Errorf("<%s> enabled but no fields to be processed defined!", utils.CDRC)
			}
		}
	}
	// Loaders sanity checks
	for _, ldrSCfg := range cfg.loaderCfg {
		if !ldrSCfg.Enabled {
			continue
		}
		for _, dir := range []string{ldrSCfg.TpInDir, ldrSCfg.TpOutDir} {
			if _, err := os.Stat(dir); err != nil && os.IsNotExist(err) {
				return fmt.Errorf("<%s> Nonexistent folder: %s", utils.LoaderS, dir)
			}
		}
		for _, data := range ldrSCfg.Data {
			if !posibleLoaderTypes.Has(data.Type) {
				return fmt.Errorf("<%s> unsupported data type %s", utils.LoaderS, data.Type)
			}

			for _, field := range data.Fields {
				if field.Type != utils.META_COMPOSED && field.Type != utils.MetaString && field.Type != utils.MetaVariable {
					return fmt.Errorf("<%s> invalid field type %s for %s at %s", utils.LoaderS, field.Type, data.Type, field.Tag)
				}
			}
		}
	}
	// SessionS checks
	if cfg.sessionSCfg.Enabled && !cfg.dispatcherSCfg.Enabled {
		if cfg.sessionSCfg.TerminateAttempts < 1 {
			return fmt.Errorf("<%s> 'terminate_attempts' should be at least 1", utils.SessionS)
		}
		if !cfg.chargerSCfg.Enabled {
			for _, conn := range cfg.sessionSCfg.ChargerSConns {
				if conn.Address == utils.MetaInternal {
					return fmt.Errorf("<%s> %s not enabled", utils.SessionS, utils.ChargerS)
				}
			}
		}
		if !cfg.ralsCfg.Enabled {
			for _, smgRALsConn := range cfg.sessionSCfg.RALsConns {
				if smgRALsConn.Address == utils.MetaInternal {
					return fmt.Errorf("<%s> %s not enabled but requested by SMGeneric component.", utils.SessionS, utils.RALService)
				}
			}
		}
		if !cfg.resourceSCfg.Enabled {
			for _, conn := range cfg.sessionSCfg.ResSConns {
				if conn.Address == utils.MetaInternal {
					return fmt.Errorf("<%s> %s not enabled but requested by SMGeneric component.", utils.SessionS, utils.ResourceS)
				}
			}
		}
		if !cfg.thresholdSCfg.Enabled {
			for _, conn := range cfg.sessionSCfg.ThreshSConns {
				if conn.Address == utils.MetaInternal {
					return fmt.Errorf("<%s> %s not enabled but requested by SMGeneric component.", utils.SessionS, utils.ThresholdS)
				}
			}
		}
		if !cfg.statsCfg.Enabled {
			for _, conn := range cfg.sessionSCfg.StatSConns {
				if conn.Address == utils.MetaInternal {
					return fmt.Errorf("<%s> %s not enabled but requested by SMGeneric component.", utils.SessionS, utils.StatS)
				}
			}
		}
		if !cfg.supplierSCfg.Enabled {
			for _, conn := range cfg.sessionSCfg.SupplSConns {
				if conn.Address == utils.MetaInternal {
					return fmt.Errorf("<%s> %s not enabled but requested by SMGeneric component.", utils.SessionS, utils.SupplierS)
				}
			}
		}
		if !cfg.attributeSCfg.Enabled {
			for _, conn := range cfg.sessionSCfg.AttrSConns {
				if conn.Address == utils.MetaInternal {
					return fmt.Errorf("<%s> %s not enabled but requested by SMGeneric component.", utils.SessionS, utils.AttributeS)
				}
			}
		}
		if !cfg.cdrsCfg.Enabled {
			for _, smgCDRSConn := range cfg.sessionSCfg.CDRsConns {
				if smgCDRSConn.Address == utils.MetaInternal {
					return fmt.Errorf("<%s> CDRS not enabled but referenced by SMGeneric component", utils.SessionS)
				}
			}
		}
		if cfg.cacheCfg[utils.CacheClosedSessions].Limit == 0 {
			return fmt.Errorf("<%s> %s needs to be != 0, received: %d", utils.CacheS, utils.CacheClosedSessions, cfg.cacheCfg[utils.CacheClosedSessions].Limit)
		}
	}

	// FreeSWITCHAgent checks
	if cfg.fsAgentCfg.Enabled {
		if len(cfg.fsAgentCfg.SessionSConns) == 0 {
			return fmt.Errorf("<%s> no %s connections defined",
				utils.FreeSWITCHAgent, utils.SessionS)
		}
		if !cfg.dispatcherSCfg.Enabled && // if dispatcher is enabled all internal connections are managed by it
			!cfg.sessionSCfg.Enabled {
			for _, connCfg := range cfg.fsAgentCfg.SessionSConns {
				if connCfg.Address == utils.MetaInternal {
					return fmt.Errorf("%s not enabled but referenced by %s",
						utils.SessionS, utils.FreeSWITCHAgent)
				}
			}
		}
	}
	// KamailioAgent checks
	if cfg.kamAgentCfg.Enabled {
		if len(cfg.kamAgentCfg.SessionSConns) == 0 {
			return fmt.Errorf("<%s> no %s connections defined",
				utils.KamailioAgent, utils.SessionS)
		}
		if !cfg.dispatcherSCfg.Enabled && // if dispatcher is enabled all internal connections are managed by it
			!cfg.sessionSCfg.Enabled {
			for _, connCfg := range cfg.kamAgentCfg.SessionSConns {
				if connCfg.Address == utils.MetaInternal {
					return fmt.Errorf("%s not enabled but referenced by %s",
						utils.SessionS, utils.KamailioAgent)
				}
			}
		}
	}
	// AsteriskAgent checks
	if cfg.asteriskAgentCfg.Enabled {
		if len(cfg.asteriskAgentCfg.SessionSConns) == 0 {
			return fmt.Errorf("<%s> no %s connections defined",
				utils.AsteriskAgent, utils.SessionS)
		}
		if !cfg.dispatcherSCfg.Enabled && // if dispatcher is enabled all internal connections are managed by it
			!cfg.sessionSCfg.Enabled {
			for _, smAstSMGConn := range cfg.asteriskAgentCfg.SessionSConns {
				if smAstSMGConn.Address == utils.MetaInternal {
					return fmt.Errorf("%s not enabled but referenced by %s",
						utils.SessionS, utils.AsteriskAgent)
				}
			}
		}
	}
	// DAgent checks
	if cfg.diameterAgentCfg.Enabled {
		if len(cfg.diameterAgentCfg.SessionSConns) == 0 {
			return fmt.Errorf("<%s> no %s connections defined",
				utils.DiameterAgent, utils.SessionS)
		}
		if !cfg.dispatcherSCfg.Enabled && // if dispatcher is enabled all internal connections are managed by it
			!cfg.sessionSCfg.Enabled {
			for _, daSMGConn := range cfg.diameterAgentCfg.SessionSConns {
				if daSMGConn.Address == utils.MetaInternal {
					return fmt.Errorf("%s not enabled but referenced by %s",
						utils.SessionS, utils.DiameterAgent)
				}
			}
		}
	}
	if cfg.radiusAgentCfg.Enabled {
		if len(cfg.radiusAgentCfg.SessionSConns) == 0 {
			return fmt.Errorf("<%s> no %s connections defined",
				utils.RadiusAgent, utils.SessionS)
		}
		if !cfg.dispatcherSCfg.Enabled && // if dispatcher is enabled all internal connections are managed by it
			!cfg.sessionSCfg.Enabled {
			for _, raSMGConn := range cfg.radiusAgentCfg.SessionSConns {
				if raSMGConn.Address == utils.MetaInternal {
					return fmt.Errorf("%s not enabled but referenced by %s",
						utils.SessionS, utils.RadiusAgent)
				}
			}
		}
	}
	if cfg.dnsAgentCfg.Enabled {
		if len(cfg.dnsAgentCfg.SessionSConns) == 0 {
			return fmt.Errorf("<%s> no %s connections defined",
				utils.DNSAgent, utils.SessionS)
		}
		if !cfg.dispatcherSCfg.Enabled && // if dispatcher is enabled all internal connections are managed by it
			!cfg.sessionSCfg.Enabled {
			for _, sSConn := range cfg.dnsAgentCfg.SessionSConns {
				if sSConn.Address == utils.MetaInternal {
					return fmt.Errorf("%s not enabled but referenced by %s", utils.SessionS, utils.DNSAgent)
				}
			}
		}
	}
	// HTTPAgent checks
	for _, httpAgentCfg := range cfg.httpAgentCfg {
		// httpAgent checks
		if !cfg.dispatcherSCfg.Enabled && // if dispatcher is enabled all internal connections are managed by it
			cfg.sessionSCfg.Enabled {
			for _, sSConn := range httpAgentCfg.SessionSConns {
				if sSConn.Address == utils.MetaInternal {
					return errors.New("SessionS not enabled but referenced by HttpAgent component")
				}
			}
		}
		if !utils.SliceHasMember([]string{utils.MetaUrl, utils.MetaXml}, httpAgentCfg.RequestPayload) {
			return fmt.Errorf("<%s> unsupported request payload %s",
				utils.HTTPAgent, httpAgentCfg.RequestPayload)
		}
		if !utils.SliceHasMember([]string{utils.MetaTextPlain, utils.MetaXml}, httpAgentCfg.ReplyPayload) {
			return fmt.Errorf("<%s> unsupported reply payload %s",
				utils.HTTPAgent, httpAgentCfg.ReplyPayload)
		}
	}
	if cfg.attributeSCfg.Enabled {
		if cfg.attributeSCfg.ProcessRuns < 1 {
			return fmt.Errorf("<%s> process_runs needs to be bigger than 0", utils.AttributeS)
		}
	}
	if cfg.chargerSCfg.Enabled && !cfg.dispatcherSCfg.Enabled &&
		(cfg.attributeSCfg == nil || !cfg.attributeSCfg.Enabled) {
		for _, connCfg := range cfg.chargerSCfg.AttributeSConns {
			if connCfg.Address == utils.MetaInternal {
				return errors.New("AttributeS not enabled but requested by ChargerS component.")
			}
		}
	}
	// ResourceLimiter checks
	if cfg.resourceSCfg.Enabled && !cfg.thresholdSCfg.Enabled && !cfg.dispatcherSCfg.Enabled {
		for _, connCfg := range cfg.resourceSCfg.ThresholdSConns {
			if connCfg.Address == utils.MetaInternal {
				return errors.New("ThresholdS not enabled but requested by ResourceS component.")
			}
		}
	}
	// StatS checks
	if cfg.statsCfg.Enabled && !cfg.thresholdSCfg.Enabled && !cfg.dispatcherSCfg.Enabled {
		for _, connCfg := range cfg.statsCfg.ThresholdSConns {
			if connCfg.Address == utils.MetaInternal {
				return errors.New("ThresholdS not enabled but requested by StatS component.")
			}
		}
	}
	// SupplierS checks
	if cfg.supplierSCfg.Enabled && !cfg.dispatcherSCfg.Enabled {
		if !cfg.resourceSCfg.Enabled {
			for _, connCfg := range cfg.supplierSCfg.ResourceSConns {
				if connCfg.Address == utils.MetaInternal {
					return errors.New("ResourceS not enabled but requested by SupplierS component.")
				}
			}
		}
		if !cfg.statsCfg.Enabled {
			for _, connCfg := range cfg.supplierSCfg.StatSConns {
				if connCfg.Address == utils.MetaInternal {
					return errors.New("StatS not enabled but requested by SupplierS component.")
				}
			}
		}
		if !cfg.attributeSCfg.Enabled {
			for _, connCfg := range cfg.supplierSCfg.AttributeSConns {
				if connCfg.Address == utils.MetaInternal {
					return errors.New("AttributeS not enabled but requested by SupplierS component.")
				}
			}
		}
	}
	// Scheduler check connection with CDR Server
	if !cfg.cdrsCfg.Enabled && !cfg.dispatcherSCfg.Enabled {
		for _, connCfg := range cfg.schedulerCfg.CDRsConns {
			if connCfg.Address == utils.MetaInternal {
				return errors.New("CDR Server not enabled but requested by Scheduler")
			}
		}
	}
	// EventReader sanity checks
	if cfg.ersCfg.Enabled {
		if !cfg.sessionSCfg.Enabled {
			for _, connCfg := range cfg.ersCfg.SessionSConns {
				if connCfg.Address == utils.MetaInternal {
					return errors.New("SessionS not enabled but requested by EventReader component.")
				}
			}
		}
		for _, rdr := range cfg.ersCfg.Readers {
			if !possibleReaderTypes.Has(rdr.Type) {
				return fmt.Errorf("<%s> unsupported data type: %s for reader with ID: %s", utils.ERs, rdr.Type, rdr.ID)
			}

			if rdr.Type == utils.MetaFileCSV {
				for _, dir := range []string{rdr.ProcessedPath, rdr.SourcePath} {
					if _, err := os.Stat(dir); err != nil && os.IsNotExist(err) {
						return fmt.Errorf("<%s> Nonexistent folder: %s for reader with ID: %s", utils.ERs, dir, rdr.ID)
					}
				}
				if rdr.FieldSep == utils.EmptyString {
					return fmt.Errorf("<%s> empty FieldSep for reader with ID: %s", utils.ERs, rdr.ID)
				}
			}
			if rdr.Type == utils.MetaKafkajsonMap && rdr.RunDelay > 0 {
				return fmt.Errorf("<%s> RunDelay field can not be bigger than zero for reader with ID: %s", utils.ERs, rdr.ID)
			}
		}
	}
	// StorDB sanity checks
	if cfg.storDbCfg.Type == utils.POSTGRES {
		if !utils.IsSliceMember([]string{utils.PostgressSSLModeDisable, utils.PostgressSSLModeAllow,
			utils.PostgressSSLModePrefer, utils.PostgressSSLModeRequire, utils.PostgressSSLModeVerifyCa,
			utils.PostgressSSLModeVerifyFull}, cfg.storDbCfg.SSLMode) {
			return fmt.Errorf("<%s> Unsuported sslmode for storDB", utils.StorDB)
		}
	}
	// DataDB sanity checks
	if cfg.dataDbCfg.DataDbType == utils.INTERNAL {
		for key, config := range cfg.cacheCfg {
			if key == utils.CacheDiameterMessages || key == utils.CacheClosedSessions {
				if config.Limit == 0 {
					return fmt.Errorf("<%s> %s needs to be != 0 when DataBD is *internal, found 0.", utils.CacheS, key)
				}
				continue
			}
			if config.Limit != 0 {
				return fmt.Errorf("<%s> %s needs to be 0 when DataBD is *internal, received : %d", utils.CacheS, key, config.Limit)
			}
		}
		if cfg.resourceSCfg.Enabled == true && cfg.resourceSCfg.StoreInterval != -1 {
			return fmt.Errorf("<%s> StoreInterval needs to be -1 when DataBD is *internal, received : %d", utils.ResourceS, cfg.resourceSCfg.StoreInterval)
		}
		if cfg.statsCfg.Enabled == true && cfg.statsCfg.StoreInterval != -1 {
			return fmt.Errorf("<%s> StoreInterval needs to be -1 when DataBD is *internal, received : %d", utils.StatS, cfg.statsCfg.StoreInterval)
		}
		if cfg.thresholdSCfg.Enabled == true && cfg.thresholdSCfg.StoreInterval != -1 {
			return fmt.Errorf("<%s> StoreInterval needs to be -1 when DataBD is *internal, received : %d", utils.ThresholdS, cfg.thresholdSCfg.StoreInterval)
		}
	}
	for item, val := range cfg.dataDbCfg.Items {
		if val.Remote == true && len(cfg.dataDbCfg.RmtConns) == 0 {
			return fmt.Errorf("Remote connections required by: <%s>", item)
		}
		if val.Replicate == true && len(cfg.dataDbCfg.RplConns) == 0 {
			return fmt.Errorf("Replicate connections required by: <%s>", item)
		}
	}
	return nil
}

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
		cfg.loadRpcConns,
		cfg.loadGeneralCfg, cfg.loadCacheCfg, cfg.loadListenCfg,
		cfg.loadHttpCfg, cfg.loadDataDBCfg, cfg.loadStorDBCfg,
		cfg.loadFilterSCfg, cfg.loadRalSCfg, cfg.loadSchedulerCfg,
		cfg.loadCdrsCfg, cfg.loadCdreCfg, cfg.loadCdrcCfg,
		cfg.loadSessionSCfg, cfg.loadFreeswitchAgentCfg, cfg.loadKamAgentCfg,
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

// loadRpcConns loads the RPCConns section of the configuration
func (cfg *CGRConfig) loadRpcConns(jsnCfg *CgrJsonCfg) (err error) {
	//var jsnRpcConns map[string]*RpcConnsJson
	//if jsnRpcConns, err = jsnCfg.RpcConnJsonCfg(); err != nil {
	//	return
	//}
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

// loadHttpCfg loads the Http section of the configuration
func (cfg *CGRConfig) loadHttpCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnHttpCfg *HTTPJsonCfg
	if jsnHttpCfg, err = jsnCfg.HttpJsonCfg(); err != nil {
		return
	}
	return cfg.httpCfg.loadFromJsonCfg(jsnHttpCfg)
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
		var customCfg *CgrJsonCfg
		var cacheJsonCfg *CacheJsonCfg
		if customCfg, err = NewCgrJsonCfgFromBytes([]byte(CGRATES_CFG_JSON_DISABLED_CACHE)); err != nil {
			return
		}
		if cacheJsonCfg, err = customCfg.CacheJsonCfg(); err != nil {
			return
		}
		if err = cfg.cacheCfg.loadFromJsonCfg(cacheJsonCfg); err != nil {
			return
		}
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
				if profileName != utils.META_DEFAULT {
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

// loadCdrcCfg loads the Cdrc section of the configuration
func (cfg *CGRConfig) loadCdrcCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnCdrcCfg []*CdrcJsonCfg
	if jsnCdrcCfg, err = jsnCfg.CdrcJsonCfg(); err != nil {
		return
	}
	if jsnCdrcCfg != nil {
		for _, jsnCrc1Cfg := range jsnCdrcCfg {
			if jsnCrc1Cfg.Id == nil || *jsnCrc1Cfg.Id == "" {
				return utils.ErrCDRCNoProfileID
			}
			if *jsnCrc1Cfg.Id == utils.META_DEFAULT {
				if cfg.dfltCdrcProfile == nil {
					cfg.dfltCdrcProfile = new(CdrcCfg)
				}
			}
			indxFound := -1 // Will be different than -1 if an instance with same id will be found
			pathFound := "" // Will be populated with the path where slice of cfgs was found
			var cdrcInstCfg *CdrcCfg
			for path := range cfg.CdrcProfiles {
				for i := range cfg.CdrcProfiles[path] {
					if cfg.CdrcProfiles[path][i].ID == *jsnCrc1Cfg.Id {
						indxFound = i
						pathFound = path
						cdrcInstCfg = cfg.CdrcProfiles[path][i]
						break
					}
				}
			}
			if cdrcInstCfg == nil {
				cdrcInstCfg = cfg.dfltCdrcProfile.Clone()
			}
			if err := cdrcInstCfg.loadFromJsonCfg(jsnCrc1Cfg, cfg.generalCfg.RSRSep); err != nil {
				return err
			}
			if cdrcInstCfg.CDRInPath == "" {
				return utils.ErrCDRCNoInPath
			}
			if _, hasDir := cfg.CdrcProfiles[cdrcInstCfg.CDRInPath]; !hasDir {
				cfg.CdrcProfiles[cdrcInstCfg.CDRInPath] = make([]*CdrcCfg, 0)
			}
			if indxFound != -1 { // Replace previous config so we have inheritance
				cfg.CdrcProfiles[pathFound][indxFound] = cdrcInstCfg
			} else {
				cfg.CdrcProfiles[cdrcInstCfg.CDRInPath] = append(cfg.CdrcProfiles[cdrcInstCfg.CDRInPath], cdrcInstCfg)
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
	return cfg.diameterAgentCfg.loadFromJsonCfg(jsnDACfg, cfg.GeneralCfg().RSRSep)
}

// loadRadiusAgentCfg loads the RadiusAgent section of the configuration
func (cfg *CGRConfig) loadRadiusAgentCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnRACfg *RadiusAgentJsonCfg
	if jsnRACfg, err = jsnCfg.RadiusAgentJsonCfg(); err != nil {
		return
	}
	return cfg.radiusAgentCfg.loadFromJsonCfg(jsnRACfg, cfg.GeneralCfg().RSRSep)
}

// loadDNSAgentCfg loads the DNSAgent section of the configuration
func (cfg *CGRConfig) loadDNSAgentCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnDNSCfg *DNSAgentJsonCfg
	if jsnDNSCfg, err = jsnCfg.DNSAgentJsonCfg(); err != nil {
		return
	}
	return cfg.dnsAgentCfg.loadFromJsonCfg(jsnDNSCfg, cfg.GeneralCfg().RSRSep)
}

// loadHttpAgentCfg loads the HttpAgent section of the configuration
func (cfg *CGRConfig) loadHttpAgentCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnHttpAgntCfg *[]*HttpAgentJsonCfg
	if jsnHttpAgntCfg, err = jsnCfg.HttpAgentJsonCfg(); err != nil {
		return
	}
	return cfg.httpAgentCfg.loadFromJsonCfg(jsnHttpAgntCfg, cfg.GeneralCfg().RSRSep)
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
			loadSCfgp.loadFromJsonCfg(profile, cfg.GeneralCfg().RSRSep)
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
	return cfg.ersCfg.loadFromJsonCfg(jsnERsCfg, cfg.GeneralCfg().RSRSep, cfg.dfltEvRdr)
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
	cfg.lks[Apier].Lock()
	defer cfg.lks[Apier].Unlock()
	return cfg.apier
}

// ERsCfg reads the EventReader configuration
func (cfg *CGRConfig) ERsCfg() *ERsCfg {
	cfg.lks[ERsJson].RLock()
	defer cfg.lks[ERsJson].RUnlock()
	return cfg.ersCfg
}

// GetReloadChan returns the reload chanel for the given section
func (cfg *CGRConfig) GetReloadChan(sectID string) chan struct{} {
	return cfg.rldChans[sectID]
}

// Call implements rpcclient.RpcClientConnection interface for internal RPC
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
	case Apier:
		jsonString = utils.ToJSON(cfg.ApierCfg())
	case CDRC_JSN:
		jsonString = utils.ToJSON(cfg.CdrcProfiles)
	case CDRE_JSN:
		jsonString = utils.ToJSON(cfg.CdreProfiles)
	case ERsJson:
		jsonString = utils.ToJSON(cfg.ERsCfg())
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

// V1ReloadConfig reloads the configuration
func (cfg *CGRConfig) V1ReloadConfig(args *ConfigReloadWithArgDispatcher, reply *string) (err error) {
	if missing := utils.MissingStructFields(args, []string{"Path"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err = cfg.loadConfig(args.Path, args.Section); err != nil {
		return
	}
	//  lock all sections
	cfg.rLockSections()

	err = cfg.checkConfigSanity()

	cfg.rUnlockSections() // unlock before checking the error

	if err != nil {
		return
	}

	if err = cfg.reloadSection(args.Section); err != nil {
		return
	}
	*reply = utils.OK
	return
}

func (cfg *CGRConfig) reloadSection(section string) (err error) {
	var fall bool
	switch section {
	default:
		return fmt.Errorf("Invalid section: <%s>", section)
	case utils.EmptyString, utils.MetaAll:
		fall = true
		fallthrough
	case GENERAL_JSN: // nothing to reload
		if !fall {
			break
		}
		fallthrough
	case DATADB_JSN:
		cfg.rldChans[DATADB_JSN] <- struct{}{}
		time.Sleep(1) // to force the context switch( to be sure we start the DB before a service that needs it)
		if !fall {
			break
		}
		fallthrough
	case STORDB_JSN:
		cfg.rldChans[STORDB_JSN] <- struct{}{}
		time.Sleep(1) // to force the context switch( to be sure we start the DB before a service that needs it)
		if !fall {
			cfg.rldChans[CDRS_JSN] <- struct{}{}
			cfg.rldChans[Apier] <- struct{}{}
			break
		}
		fallthrough
	case LISTEN_JSN:
		if !fall {
			break
		}
		fallthrough
	case TlsCfgJson: // nothing to reload
		if !fall {
			break
		}
		fallthrough
	case HTTP_JSN:
		if !fall {
			break
		}
		fallthrough
	case SCHEDULER_JSN:
		if !fall {
			cfg.rldChans[DATADB_JSN] <- struct{}{} // reload datadb before
			time.Sleep(1)                          // to force the context switch( to be sure we start the DB before a service that needs it)
		}
		cfg.rldChans[SCHEDULER_JSN] <- struct{}{}
		if !fall {
			break
		}
		fallthrough
	case RALS_JSN:
		if !fall {
			cfg.rldChans[DATADB_JSN] <- struct{}{} // reload datadb before
			cfg.rldChans[STORDB_JSN] <- struct{}{}
			time.Sleep(1) // to force the context switch( to be sure we start the DB before a service that needs it)
		}
		cfg.rldChans[RALS_JSN] <- struct{}{}
		if !fall {
			break
		}
		fallthrough
	case CDRS_JSN:
		if !fall {
			cfg.rldChans[DATADB_JSN] <- struct{}{} // reload datadb before
			cfg.rldChans[STORDB_JSN] <- struct{}{}
			time.Sleep(1) // to force the context switch( to be sure we start the DB before a service that needs it)
		}
		cfg.rldChans[CDRS_JSN] <- struct{}{}
		if !fall {
			break
		}
		fallthrough
	case CDRC_JSN:
		if !fall {
			break
		}
		fallthrough
	case ERsJson:
		cfg.rldChans[ERsJson] <- struct{}{}
		if !fall {
			break
		}
		fallthrough
	case SessionSJson:
		if !fall {
			cfg.rldChans[DATADB_JSN] <- struct{}{} // reload datadb before
			time.Sleep(1)                          // to force the context switch( to be sure we start the DB before a service that needs it)
		}
		cfg.rldChans[SessionSJson] <- struct{}{}
		if !fall {
			break
		}
		fallthrough
	case AsteriskAgentJSN:
		cfg.rldChans[AsteriskAgentJSN] <- struct{}{}
		if !fall {
			break
		}
		fallthrough
	case FreeSWITCHAgentJSN:
		cfg.rldChans[FreeSWITCHAgentJSN] <- struct{}{}
		if !fall {
			break
		}
		fallthrough
	case KamailioAgentJSN:
		cfg.rldChans[KamailioAgentJSN] <- struct{}{}
		if !fall {
			break
		}
		fallthrough
	case DA_JSN:
		cfg.rldChans[DA_JSN] <- struct{}{}
		if !fall {
			break
		}
		fallthrough
	case RA_JSN:
		cfg.rldChans[RA_JSN] <- struct{}{}
		if !fall {
			break
		}
		fallthrough
	case HttpAgentJson:
		cfg.rldChans[HttpAgentJson] <- struct{}{}
		if !fall {
			break
		}
		fallthrough
	case DNSAgentJson:
		cfg.rldChans[DNSAgentJson] <- struct{}{}
		if !fall {
			break
		}
		fallthrough
	case ATTRIBUTE_JSN:
		if !fall {
			cfg.rldChans[DATADB_JSN] <- struct{}{} // reload datadb before
			time.Sleep(1)                          // to force the context switch( to be sure we start the DB before a service that needs it)
		}
		cfg.rldChans[ATTRIBUTE_JSN] <- struct{}{}
		if !fall {
			break
		}
		fallthrough
	case ChargerSCfgJson:
		if !fall {
			cfg.rldChans[DATADB_JSN] <- struct{}{} // reload datadb before
			time.Sleep(1)                          // to force the context switch( to be sure we start the DB before a service that needs it)
		}
		cfg.rldChans[ChargerSCfgJson] <- struct{}{}
		if !fall {
			break
		}
		fallthrough
	case RESOURCES_JSON:
		if !fall {
			cfg.rldChans[DATADB_JSN] <- struct{}{} // reload datadb before
			time.Sleep(1)                          // to force the context switch( to be sure we start the DB before a service that needs it)
		}
		cfg.rldChans[RESOURCES_JSON] <- struct{}{}
		if !fall {
			break
		}
		fallthrough
	case STATS_JSON:
		if !fall {
			cfg.rldChans[DATADB_JSN] <- struct{}{} // reload datadb before
			time.Sleep(1)                          // to force the context switch( to be sure we start the DB before a service that needs it)
		}
		cfg.rldChans[STATS_JSON] <- struct{}{}
		if !fall {
			break
		}
		fallthrough
	case THRESHOLDS_JSON:
		if !fall {
			cfg.rldChans[DATADB_JSN] <- struct{}{} // reload datadb before
			time.Sleep(1)                          // to force the context switch( to be sure we start the DB before a service that needs it)
		}
		cfg.rldChans[THRESHOLDS_JSON] <- struct{}{}
		if !fall {
			break
		}
		fallthrough
	case SupplierSJson:
		if !fall {
			cfg.rldChans[DATADB_JSN] <- struct{}{} // reload datadb before
			time.Sleep(1)                          // to force the context switch( to be sure we start the DB before a service that needs it)
		}
		cfg.rldChans[SupplierSJson] <- struct{}{}
		if !fall {
			break
		}
		fallthrough
	case LoaderJson:
		if !fall {
			cfg.rldChans[DATADB_JSN] <- struct{}{} // reload datadb before
			time.Sleep(1)                          // to force the context switch( to be sure we start the DB before a service that needs it)
		}
		cfg.rldChans[LoaderJson] <- struct{}{}
		if !fall {
			break
		}
		fallthrough
	case DispatcherSJson:
		if !fall {
			cfg.rldChans[DATADB_JSN] <- struct{}{} // reload datadb before
			time.Sleep(1)                          // to force the context switch( to be sure we start the DB before a service that needs it)
		}
		cfg.rldChans[DispatcherSJson] <- struct{}{}
		if !fall {
			break
		}
		fallthrough
	case AnalyzerCfgJson:
		if !fall {
			break
		}
		fallthrough
	case Apier:
		if !fall {
			cfg.rldChans[STORDB_JSN] <- struct{}{}
			time.Sleep(1) // to force the context switch( to be sure we start the DB before a service that needs it)
		}
		cfg.rldChans[Apier] <- struct{}{}
		if !fall {
			break
		}
	}
	return
}

func (cfg *CGRConfig) loadConfig(path, section string) (err error) {
	var loadFuncs []func(*CgrJsonCfg) error
	var fall bool
	switch section {
	default:
		return fmt.Errorf("Invalid section: <%s>", section)
	case utils.EmptyString, utils.MetaAll:
		fall = true
		fallthrough
	case GENERAL_JSN:
		cfg.lks[GENERAL_JSN].Lock()
		defer cfg.lks[GENERAL_JSN].Unlock()
		loadFuncs = append(loadFuncs, cfg.loadGeneralCfg)
		if !fall {
			break
		}
		fallthrough
	case DATADB_JSN:
		cfg.lks[DATADB_JSN].Lock()
		defer cfg.lks[DATADB_JSN].Unlock()
		loadFuncs = append(loadFuncs, cfg.loadDataDBCfg)
		if !fall {
			break
		}
		fallthrough
	case STORDB_JSN:
		cfg.lks[STORDB_JSN].Lock()
		defer cfg.lks[STORDB_JSN].Unlock()
		loadFuncs = append(loadFuncs, cfg.loadStorDBCfg)
		if !fall {
			break
		}
		fallthrough
	case LISTEN_JSN:
		cfg.lks[LISTEN_JSN].Lock()
		defer cfg.lks[LISTEN_JSN].Unlock()
		loadFuncs = append(loadFuncs, cfg.loadListenCfg)
		if !fall {
			break
		}
		fallthrough
	case TlsCfgJson:
		cfg.lks[TlsCfgJson].Lock()
		defer cfg.lks[TlsCfgJson].Unlock()
		loadFuncs = append(loadFuncs, cfg.loadTlsCgrCfg)
		if !fall {
			break
		}
		fallthrough
	case HTTP_JSN:
		cfg.lks[HTTP_JSN].Lock()
		defer cfg.lks[HTTP_JSN].Unlock()
		loadFuncs = append(loadFuncs, cfg.loadHttpCfg)
		if !fall {
			break
		}
		fallthrough
	case SCHEDULER_JSN:
		cfg.lks[SCHEDULER_JSN].Lock()
		defer cfg.lks[SCHEDULER_JSN].Unlock()
		loadFuncs = append(loadFuncs, cfg.loadSchedulerCfg)
		if !fall {
			break
		}
		fallthrough
	case CACHE_JSN:
		cfg.lks[CACHE_JSN].Lock()
		defer cfg.lks[CACHE_JSN].Unlock()
		loadFuncs = append(loadFuncs, cfg.loadCacheCfg)
		if !fall {
			break
		}
		fallthrough
	case FilterSjsn:
		cfg.lks[FilterSjsn].Lock()
		defer cfg.lks[FilterSjsn].Unlock()
		loadFuncs = append(loadFuncs, cfg.loadFilterSCfg)
		if !fall {
			break
		}
		fallthrough
	case RALS_JSN:
		cfg.lks[RALS_JSN].Lock()
		defer cfg.lks[RALS_JSN].Unlock()
		loadFuncs = append(loadFuncs, cfg.loadRalSCfg)
		if !fall {
			break
		}
		fallthrough
	case CDRS_JSN:
		cfg.lks[CDRS_JSN].Lock()
		defer cfg.lks[CDRS_JSN].Unlock()
		loadFuncs = append(loadFuncs, cfg.loadCdrsCfg)
		if !fall {
			break
		}
		fallthrough
	case CDRE_JSN:
		cfg.lks[CDRE_JSN].Lock()
		defer cfg.lks[CDRE_JSN].Unlock()
		loadFuncs = append(loadFuncs, cfg.loadCdreCfg)
		if !fall {
			break
		}
		fallthrough
	case CDRC_JSN:
		cfg.lks[CDRC_JSN].Lock()
		defer cfg.lks[CDRC_JSN].Unlock()
		loadFuncs = append(loadFuncs, cfg.loadCdrcCfg)
		if !fall {
			break
		}
		fallthrough
	case ERsJson:
		cfg.lks[ERsJson].Lock()
		defer cfg.lks[ERsJson].Unlock()
		loadFuncs = append(loadFuncs, cfg.loadErsCfg)
		if !fall {
			break
		}
		fallthrough
	case SessionSJson:
		cfg.lks[SessionSJson].Lock()
		defer cfg.lks[SessionSJson].Unlock()
		loadFuncs = append(loadFuncs, cfg.loadSessionSCfg)
		if !fall {
			break
		}
		fallthrough
	case AsteriskAgentJSN:
		cfg.lks[AsteriskAgentJSN].Lock()
		defer cfg.lks[AsteriskAgentJSN].Unlock()
		loadFuncs = append(loadFuncs, cfg.loadAsteriskAgentCfg)
		if !fall {
			break
		}
		fallthrough
	case FreeSWITCHAgentJSN:
		cfg.lks[FreeSWITCHAgentJSN].Lock()
		defer cfg.lks[FreeSWITCHAgentJSN].Unlock()
		loadFuncs = append(loadFuncs, cfg.loadFreeswitchAgentCfg)
		if !fall {
			break
		}
		fallthrough
	case KamailioAgentJSN:
		cfg.lks[KamailioAgentJSN].Lock()
		defer cfg.lks[KamailioAgentJSN].Unlock()
		loadFuncs = append(loadFuncs, cfg.loadKamAgentCfg)
		if !fall {
			break
		}
		fallthrough
	case DA_JSN:
		cfg.lks[DA_JSN].Lock()
		defer cfg.lks[DA_JSN].Unlock()
		loadFuncs = append(loadFuncs, cfg.loadDiameterAgentCfg)
		if !fall {
			break
		}
		fallthrough
	case RA_JSN:
		cfg.lks[RA_JSN].Lock()
		defer cfg.lks[RA_JSN].Unlock()
		loadFuncs = append(loadFuncs, cfg.loadRadiusAgentCfg)
		if !fall {
			break
		}
		fallthrough
	case HttpAgentJson:
		cfg.lks[HttpAgentJson].Lock()
		defer cfg.lks[HttpAgentJson].Unlock()
		loadFuncs = append(loadFuncs, cfg.loadHttpAgentCfg)
		if !fall {
			break
		}
		fallthrough
	case DNSAgentJson:
		cfg.lks[DNSAgentJson].Lock()
		defer cfg.lks[DNSAgentJson].Unlock()
		loadFuncs = append(loadFuncs, cfg.loadDNSAgentCfg)
		if !fall {
			break
		}
		fallthrough
	case ATTRIBUTE_JSN:
		cfg.lks[ATTRIBUTE_JSN].Lock()
		defer cfg.lks[ATTRIBUTE_JSN].Unlock()
		loadFuncs = append(loadFuncs, cfg.loadAttributeSCfg)
		if !fall {
			break
		}
		fallthrough
	case ChargerSCfgJson:
		cfg.lks[ChargerSCfgJson].Lock()
		defer cfg.lks[ChargerSCfgJson].Unlock()
		loadFuncs = append(loadFuncs, cfg.loadChargerSCfg)
		if !fall {
			break
		}
		fallthrough
	case RESOURCES_JSON:
		cfg.lks[RESOURCES_JSON].Lock()
		defer cfg.lks[RESOURCES_JSON].Unlock()
		loadFuncs = append(loadFuncs, cfg.loadResourceSCfg)
		if !fall {
			break
		}
		fallthrough
	case STATS_JSON:
		cfg.lks[STATS_JSON].Lock()
		defer cfg.lks[STATS_JSON].Unlock()
		loadFuncs = append(loadFuncs, cfg.loadStatSCfg)
		if !fall {
			break
		}
		fallthrough
	case THRESHOLDS_JSON:
		cfg.lks[THRESHOLDS_JSON].Lock()
		defer cfg.lks[THRESHOLDS_JSON].Unlock()
		loadFuncs = append(loadFuncs, cfg.loadThresholdSCfg)
		if !fall {
			break
		}
		fallthrough
	case SupplierSJson:
		cfg.lks[SupplierSJson].Lock()
		defer cfg.lks[SupplierSJson].Unlock()
		loadFuncs = append(loadFuncs, cfg.loadSupplierSCfg)
		if !fall {
			break
		}
		fallthrough
	case LoaderJson:
		cfg.lks[LoaderJson].Lock()
		defer cfg.lks[LoaderJson].Unlock()
		loadFuncs = append(loadFuncs, cfg.loadLoaderSCfg)
		if !fall {
			break
		}
		fallthrough
	case MAILER_JSN:
		cfg.lks[MAILER_JSN].Lock()
		defer cfg.lks[MAILER_JSN].Unlock()
		loadFuncs = append(loadFuncs, cfg.loadMailerCfg)
		if !fall {
			break
		}
		fallthrough
	case SURETAX_JSON:
		cfg.lks[SURETAX_JSON].Lock()
		defer cfg.lks[SURETAX_JSON].Unlock()
		loadFuncs = append(loadFuncs, cfg.loadSureTaxCfg)
		if !fall {
			break
		}
		fallthrough
	case CgrLoaderCfgJson:
		cfg.lks[CgrLoaderCfgJson].Lock()
		defer cfg.lks[CgrLoaderCfgJson].Unlock()
		loadFuncs = append(loadFuncs, cfg.loadLoaderCgrCfg)
		if !fall {
			break
		}
		fallthrough
	case CgrMigratorCfgJson:
		cfg.lks[CgrMigratorCfgJson].Lock()
		defer cfg.lks[CgrMigratorCfgJson].Unlock()
		loadFuncs = append(loadFuncs, cfg.loadMigratorCgrCfg)
		if !fall {
			break
		}
		fallthrough
	case DispatcherSJson:
		cfg.lks[DispatcherSJson].Lock()
		defer cfg.lks[DispatcherSJson].Unlock()
		loadFuncs = append(loadFuncs, cfg.loadDispatcherSCfg)
		if !fall {
			break
		}
		fallthrough
	case AnalyzerCfgJson:
		cfg.lks[AnalyzerCfgJson].Lock()
		defer cfg.lks[AnalyzerCfgJson].Unlock()
		loadFuncs = append(loadFuncs, cfg.loadAnalyzerCgrCfg)
		if !fall {
			break
		}
		fallthrough
	case Apier:
		cfg.lks[Apier].Lock()
		defer cfg.lks[Apier].Unlock()
		loadFuncs = append(loadFuncs, cfg.loadApierCfg)
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
		return cfg.loadConfigFromHttp(path, loadFuncs) // prefix protocol
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

func (cfg *CGRConfig) loadConfigFromHttp(urlPaths string, loadFuncs []func(jsnCfg *CgrJsonCfg) error) (err error) {
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
			return
		}
	}
	return
}

// populates the config locks and the reload channels
func (cfg *CGRConfig) initChanels() {
	cfg.lks = make(map[string]*sync.RWMutex)
	cfg.rldChans = make(map[string]chan struct{})
	for _, section := range []string{GENERAL_JSN, DATADB_JSN, STORDB_JSN, LISTEN_JSN, TlsCfgJson, HTTP_JSN, SCHEDULER_JSN, CACHE_JSN, FilterSjsn, RALS_JSN,
		CDRS_JSN, CDRE_JSN, CDRC_JSN, ERsJson, SessionSJson, AsteriskAgentJSN, FreeSWITCHAgentJSN, KamailioAgentJSN,
		DA_JSN, RA_JSN, HttpAgentJson, DNSAgentJson, ATTRIBUTE_JSN, ChargerSCfgJson, RESOURCES_JSON, STATS_JSON, THRESHOLDS_JSON,
		SupplierSJson, LoaderJson, MAILER_JSN, SURETAX_JSON, CgrLoaderCfgJson, CgrMigratorCfgJson, DispatcherSJson, AnalyzerCfgJson, Apier} {
		cfg.lks[section] = new(sync.RWMutex)
		cfg.rldChans[section] = make(chan struct{}, 1)
	}
}
