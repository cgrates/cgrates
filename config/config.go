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
	"net/url"
	"os"
	"path/filepath"
	"strconv"
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

func (self DbDefaults) DBPort(dbType string, flagInput string) string {
	if flagInput != utils.MetaDynamic {
		return flagInput
	}
	return self[dbType]["DbPort"]
}

func (self DbDefaults) DBPass(dbType string, flagInput string) string {
	if flagInput != utils.MetaDynamic {
		return flagInput
	}
	return self[dbType]["DbPass"]
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

func NewDefaultCGRConfig() (*CGRConfig, error) {
	cfg := new(CGRConfig)
	cfg.lks = make(map[string]*sync.RWMutex)
	cfg.lks[ERsJson] = new(sync.RWMutex)
	cfg.DataFolderPath = "/usr/share/cgrates/"
	cfg.MaxCallDuration = time.Duration(3) * time.Hour // Hardcoded for now

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
	cfg.ralsCfg.RALsMaxComputedUsage = make(map[string]time.Duration)
	cfg.ralsCfg.RALsBalanceRatingSubject = make(map[string]string)
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
	cfg.loaderCfg = make([]*LoaderSCfg, 0)
	cfg.apier = new(ApierCfg)
	cfg.ersCfg = new(ERsCfg)

	cfg.ConfigReloads = make(map[string]chan struct{})
	cfg.ConfigReloads[utils.CDRC] = make(chan struct{}, 1)
	cfg.ConfigReloads[utils.CDRC] <- struct{}{} // Unlock the channel
	cfg.ConfigReloads[utils.CDRE] = make(chan struct{}, 1)
	cfg.ConfigReloads[utils.CDRE] <- struct{}{} // Unlock the channel
	cfg.ConfigReloads[utils.SURETAX] = make(chan struct{}, 1)
	cfg.ConfigReloads[utils.SURETAX] <- struct{}{} // Unlock the channel
	cfg.ConfigReloads[utils.DIAMETER_AGENT] = make(chan struct{}, 1)
	cfg.ConfigReloads[utils.DIAMETER_AGENT] <- struct{}{} // Unlock the channel
	cfg.ConfigReloads[utils.SMAsterisk] = make(chan struct{}, 1)
	cfg.ConfigReloads[utils.SMAsterisk] <- struct{}{} // Unlock the channel

	cfg.rldChans = make(map[string]chan struct{})
	cfg.rldChans[ERsJson] = make(chan struct{}, 1)

	cgrJsonCfg, err := NewCgrJsonCfgFromReader(strings.NewReader(CGRATES_CFG_JSON))
	if err != nil {
		return nil, err
	}
	if err := cfg.loadFromJsonCfg(cgrJsonCfg); err != nil {
		return nil, err
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
	if err := cfg.checkConfigSanity(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func NewCGRConfigFromJsonStringWithDefaults(cfgJsonStr string) (*CGRConfig, error) {
	cfg, _ := NewDefaultCGRConfig()
	if jsnCfg, err := NewCgrJsonCfgFromReader(strings.NewReader(cfgJsonStr)); err != nil {
		return nil, err
	} else if err := cfg.loadFromJsonCfg(jsnCfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

// Reads all .json files out of a folder/subfolders and loads them up in lexical order
func NewCGRConfigFromPath(path string) (*CGRConfig, error) {
	cfg, err := NewDefaultCGRConfig()
	if err != nil {
		return nil, err
	}
	cfg.ConfigPath = path

	if err := updateConfigFromPath(path, func(jsnCfg *CgrJsonCfg) error {
		return cfg.loadFromJsonCfg(jsnCfg)
	}); err != nil {
		return nil, err
	}
	if err = cfg.checkConfigSanity(); err != nil { // should we check only the updated sections?
		return nil, err
	}
	return cfg, nil
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
	loaderCfg    []*LoaderSCfg         // LoaderS configs
	httpAgentCfg HttpAgentCfgs         // HttpAgent configs

	ConfigReloads map[string]chan struct{} // Signals to specific entities that a config reload should occur
	rldChans      map[string]chan struct{} // index here the channels used for reloads

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

func (self *CGRConfig) checkConfigSanity() error {
	// Rater checks
	if self.ralsCfg.RALsEnabled && !self.dispatcherSCfg.Enabled {
		if !self.statsCfg.Enabled {
			for _, connCfg := range self.ralsCfg.RALsStatSConns {
				if connCfg.Address == utils.MetaInternal {
					return fmt.Errorf("%s not enabled but requested by %s component.",
						utils.StatS, utils.RALService)
				}
			}
		}
		if !self.thresholdSCfg.Enabled {
			for _, connCfg := range self.ralsCfg.RALsThresholdSConns {
				if connCfg.Address == utils.MetaInternal {
					return fmt.Errorf("%s not enabled but requested by %s component.",
						utils.ThresholdS, utils.RALService)
				}
			}
		}
	}
	// CDRServer checks
	if self.cdrsCfg.CDRSEnabled && !self.dispatcherSCfg.Enabled {
		if !self.chargerSCfg.Enabled {
			for _, conn := range self.cdrsCfg.CDRSChargerSConns {
				if conn.Address == utils.MetaInternal {
					return errors.New("ChargerS not enabled but requested by CDRS component.")
				}
			}
		}
		if !self.ralsCfg.RALsEnabled {
			for _, cdrsRaterConn := range self.cdrsCfg.CDRSRaterConns {
				if cdrsRaterConn.Address == utils.MetaInternal {
					return errors.New("RALs not enabled but requested by CDRS component.")
				}
			}
		}
		if !self.attributeSCfg.Enabled {
			for _, connCfg := range self.cdrsCfg.CDRSAttributeSConns {
				if connCfg.Address == utils.MetaInternal {
					return errors.New("AttributeS not enabled but requested by CDRS component.")
				}
			}
		}
		if !self.statsCfg.Enabled {
			for _, connCfg := range self.cdrsCfg.CDRSStatSConns {
				if connCfg.Address == utils.MetaInternal {
					return errors.New("StatS not enabled but requested by CDRS component.")
				}
			}
		}
		for _, cdrePrfl := range self.cdrsCfg.CDRSOnlineCDRExports {
			if _, hasIt := self.CdreProfiles[cdrePrfl]; !hasIt {
				return fmt.Errorf("<CDRS> Cannot find CDR export template with ID: <%s>", cdrePrfl)
			}
		}
		if !self.thresholdSCfg.Enabled {
			for _, connCfg := range self.cdrsCfg.CDRSThresholdSConns {
				if connCfg.Address == utils.MetaInternal {
					return errors.New("ThresholdS not enabled but requested by CDRS component.")
				}
			}
		}
	}
	// CDRC sanity checks
	for _, cdrcCfgs := range self.CdrcProfiles {
		for _, cdrcInst := range cdrcCfgs {
			if !cdrcInst.Enabled {
				continue
			}
			if len(cdrcInst.CdrsConns) == 0 {
				return fmt.Errorf("<CDRC> Instance: %s, CdrC enabled but no CDRS defined!", cdrcInst.ID)
			}
			if !self.cdrsCfg.CDRSEnabled && !self.dispatcherSCfg.Enabled {
				for _, conn := range cdrcInst.CdrsConns {
					if conn.Address == utils.MetaInternal {
						return errors.New("CDRS not enabled but referenced from CDRC")
					}
				}
			}
			if len(cdrcInst.ContentFields) == 0 {
				return errors.New("CdrC enabled but no fields to be processed defined!")
			}
			if cdrcInst.CdrFormat == utils.MetaFileCSV {
				for _, cdrFld := range cdrcInst.ContentFields {
					for _, rsrFld := range cdrFld.Value {
						if rsrFld.attrName != "" {
							if _, errConv := strconv.Atoi(rsrFld.attrName); errConv != nil {
								return fmt.Errorf("CDR fields must be indices in case of .csv files, have instead: %s", rsrFld.attrName)
							}
						}
					}
				}
			}
		}
	}
	// Loaders sanity checks
	for _, ldrSCfg := range self.loaderCfg {
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
	if self.sessionSCfg.Enabled && !self.dispatcherSCfg.Enabled {
		if !self.chargerSCfg.Enabled {
			for _, conn := range self.sessionSCfg.ChargerSConns {
				if conn.Address == utils.MetaInternal {
					return fmt.Errorf("<%s> %s not enabled", utils.SessionS, utils.ChargerS)
				}
			}
		}
		if !self.ralsCfg.RALsEnabled {
			for _, smgRALsConn := range self.sessionSCfg.RALsConns {
				if smgRALsConn.Address == utils.MetaInternal {
					return fmt.Errorf("<%s> %s not enabled but requested by SMGeneric component.", utils.SessionS, utils.RALService)
				}
			}
		}
		if !self.resourceSCfg.Enabled {
			for _, conn := range self.sessionSCfg.ResSConns {
				if conn.Address == utils.MetaInternal {
					return fmt.Errorf("<%s> %s not enabled but requested by SMGeneric component.", utils.SessionS, utils.ResourceS)
				}
			}
		}
		if !self.thresholdSCfg.Enabled {
			for _, conn := range self.sessionSCfg.ThreshSConns {
				if conn.Address == utils.MetaInternal {
					return fmt.Errorf("<%s> %s not enabled but requested by SMGeneric component.", utils.SessionS, utils.ThresholdS)
				}
			}
		}
		if !self.statsCfg.Enabled {
			for _, conn := range self.sessionSCfg.StatSConns {
				if conn.Address == utils.MetaInternal {
					return fmt.Errorf("<%s> %s not enabled but requested by SMGeneric component.", utils.SessionS, utils.StatS)
				}
			}
		}
		if !self.supplierSCfg.Enabled {
			for _, conn := range self.sessionSCfg.SupplSConns {
				if conn.Address == utils.MetaInternal {
					return fmt.Errorf("<%s> %s not enabled but requested by SMGeneric component.", utils.SessionS, utils.SupplierS)
				}
			}
		}
		if !self.attributeSCfg.Enabled {
			for _, conn := range self.sessionSCfg.AttrSConns {
				if conn.Address == utils.MetaInternal {
					return fmt.Errorf("<%s> %s not enabled but requested by SMGeneric component.", utils.SessionS, utils.AttributeS)
				}
			}
		}
		if !self.cdrsCfg.CDRSEnabled {
			for _, smgCDRSConn := range self.sessionSCfg.CDRsConns {
				if smgCDRSConn.Address == utils.MetaInternal {
					return fmt.Errorf("<%s> CDRS not enabled but referenced by SMGeneric component", utils.SessionS)
				}
			}
		}
	}
	// FreeSWITCHAgent checks
	if self.fsAgentCfg.Enabled {
		if len(self.fsAgentCfg.SessionSConns) == 0 {
			return fmt.Errorf("<%s> no %s connections defined",
				utils.FreeSWITCHAgent, utils.SessionS)
		}
		if !self.dispatcherSCfg.Enabled && // if dispatcher is enabled all internal connections are managed by it
			!self.sessionSCfg.Enabled {
			for _, connCfg := range self.fsAgentCfg.SessionSConns {
				if connCfg.Address == utils.MetaInternal {
					return fmt.Errorf("%s not enabled but referenced by %s",
						utils.SessionS, utils.FreeSWITCHAgent)
				}
			}
		}
	}
	// KamailioAgent checks
	if self.kamAgentCfg.Enabled {
		if len(self.kamAgentCfg.SessionSConns) == 0 {
			return fmt.Errorf("<%s> no %s connections defined",
				utils.KamailioAgent, utils.SessionS)
		}
		if !self.dispatcherSCfg.Enabled && // if dispatcher is enabled all internal connections are managed by it
			!self.sessionSCfg.Enabled {
			for _, connCfg := range self.kamAgentCfg.SessionSConns {
				if connCfg.Address == utils.MetaInternal {
					return fmt.Errorf("%s not enabled but referenced by %s",
						utils.SessionS, utils.KamailioAgent)
				}
			}
		}
	}
	// AsteriskAgent checks
	if self.asteriskAgentCfg.Enabled {
		if len(self.asteriskAgentCfg.SessionSConns) == 0 {
			return fmt.Errorf("<%s> no %s connections defined",
				utils.AsteriskAgent, utils.SessionS)
		}
		if !self.dispatcherSCfg.Enabled && // if dispatcher is enabled all internal connections are managed by it
			!self.sessionSCfg.Enabled {
			for _, smAstSMGConn := range self.asteriskAgentCfg.SessionSConns {
				if smAstSMGConn.Address == utils.MetaInternal {
					return fmt.Errorf("%s not enabled but referenced by %s",
						utils.SessionS, utils.AsteriskAgent)
				}
			}
		}
	}
	// DAgent checks
	if self.diameterAgentCfg.Enabled {
		if len(self.diameterAgentCfg.SessionSConns) == 0 {
			return fmt.Errorf("<%s> no %s connections defined",
				utils.DiameterAgent, utils.SessionS)
		}
		if !self.dispatcherSCfg.Enabled && // if dispatcher is enabled all internal connections are managed by it
			!self.sessionSCfg.Enabled {
			for _, daSMGConn := range self.diameterAgentCfg.SessionSConns {
				if daSMGConn.Address == utils.MetaInternal {
					return fmt.Errorf("%s not enabled but referenced by %s",
						utils.SessionS, utils.DiameterAgent)
				}
			}
		}
	}
	if self.radiusAgentCfg.Enabled {
		if len(self.radiusAgentCfg.SessionSConns) == 0 {
			return fmt.Errorf("<%s> no %s connections defined",
				utils.RadiusAgent, utils.SessionS)
		}
		if !self.dispatcherSCfg.Enabled && // if dispatcher is enabled all internal connections are managed by it
			!self.sessionSCfg.Enabled {
			for _, raSMGConn := range self.radiusAgentCfg.SessionSConns {
				if raSMGConn.Address == utils.MetaInternal {
					return fmt.Errorf("%s not enabled but referenced by %s",
						utils.SessionS, utils.RadiusAgent)
				}
			}
		}
	}
	if self.dnsAgentCfg.Enabled {
		if len(self.dnsAgentCfg.SessionSConns) == 0 {
			return fmt.Errorf("<%s> no %s connections defined",
				utils.DNSAgent, utils.SessionS)
		}
		if !self.dispatcherSCfg.Enabled && // if dispatcher is enabled all internal connections are managed by it
			!self.sessionSCfg.Enabled {
			for _, sSConn := range self.dnsAgentCfg.SessionSConns {
				if sSConn.Address == utils.MetaInternal {
					return fmt.Errorf("%s not enabled but referenced by %s", utils.SessionS, utils.DNSAgent)
				}
			}
		}
	}
	// HTTPAgent checks
	for _, httpAgentCfg := range self.httpAgentCfg {
		// httpAgent checks
		if !self.dispatcherSCfg.Enabled && // if dispatcher is enabled all internal connections are managed by it
			self.sessionSCfg.Enabled {
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
	if self.attributeSCfg.Enabled {
		if self.attributeSCfg.ProcessRuns < 1 {
			return fmt.Errorf("<%s> process_runs needs to be bigger than 0", utils.AttributeS)
		}
	}
	if self.chargerSCfg.Enabled && !self.dispatcherSCfg.Enabled &&
		(self.attributeSCfg == nil || !self.attributeSCfg.Enabled) {
		for _, connCfg := range self.chargerSCfg.AttributeSConns {
			if connCfg.Address == utils.MetaInternal {
				return errors.New("AttributeS not enabled but requested by ChargerS component.")
			}
		}
	}
	// ResourceLimiter checks
	if self.resourceSCfg.Enabled && !self.thresholdSCfg.Enabled && !self.dispatcherSCfg.Enabled {
		for _, connCfg := range self.resourceSCfg.ThresholdSConns {
			if connCfg.Address == utils.MetaInternal {
				return errors.New("ThresholdS not enabled but requested by ResourceS component.")
			}
		}
	}
	// StatS checks
	if self.statsCfg.Enabled && !self.thresholdSCfg.Enabled && !self.dispatcherSCfg.Enabled {
		for _, connCfg := range self.statsCfg.ThresholdSConns {
			if connCfg.Address == utils.MetaInternal {
				return errors.New("ThresholdS not enabled but requested by StatS component.")
			}
		}
	}
	// SupplierS checks
	if self.supplierSCfg.Enabled && !self.dispatcherSCfg.Enabled {
		for _, connCfg := range self.supplierSCfg.RALsConns {
			if connCfg.Address != utils.MetaInternal {
				return errors.New("Only <*internal> RALs connectivity allowed in SupplierS for now")
			}
			if !self.ralsCfg.RALsEnabled {
				return errors.New("RALs not enabled but requested by SupplierS component.")
			}
		}
		if !self.resourceSCfg.Enabled {
			for _, connCfg := range self.supplierSCfg.ResourceSConns {
				if connCfg.Address == utils.MetaInternal {
					return errors.New("ResourceS not enabled but requested by SupplierS component.")
				}
			}
		}
		if !self.resourceSCfg.Enabled {
			for _, connCfg := range self.supplierSCfg.StatSConns {
				if connCfg.Address == utils.MetaInternal {
					return errors.New("StatS not enabled but requested by SupplierS component.")
				}
			}
		}
		if !self.attributeSCfg.Enabled {
			for _, connCfg := range self.supplierSCfg.AttributeSConns {
				if connCfg.Address == utils.MetaInternal {
					return errors.New("AttributeS not enabled but requested by SupplierS component.")
				}
			}
		}
	}
	// Scheduler check connection with CDR Server
	if !self.cdrsCfg.CDRSEnabled && !self.dispatcherSCfg.Enabled {
		for _, connCfg := range self.schedulerCfg.CDRsConns {
			if connCfg.Address == utils.MetaInternal {
				return errors.New("CDR Server not enabled but requested by Scheduler")
			}
		}
	}
	return nil
}

func (self *CGRConfig) LazySanityCheck() {
	for _, cdrePrfl := range self.cdrsCfg.CDRSOnlineCDRExports {
		if cdreProfile, hasIt := self.CdreProfiles[cdrePrfl]; hasIt && (cdreProfile.ExportFormat == utils.MetaS3jsonMap || cdreProfile.ExportFormat == utils.MetaSQSjsonMap) {
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
	if err = cfg.loadGeneralCfg(jsnCfg); err != nil {
		return err
	}

	if err = cfg.loadCacheCfg(jsnCfg); err != nil {
		return err
	}

	if err = cfg.loadListenCfg(jsnCfg); err != nil {
		return err
	}

	if err = cfg.loadHttpCfg(jsnCfg); err != nil {
		return err
	}

	if err = cfg.loadDataDBCfg(jsnCfg); err != nil {
		return err
	}

	if err = cfg.loadStorDBCfg(jsnCfg); err != nil {
		return err
	}

	if err = cfg.loadFilterSCfg(jsnCfg); err != nil {
		return
	}

	if err = cfg.loadRalSCfg(jsnCfg); err != nil {
		return
	}

	if err = cfg.loadSchedulerCfg(jsnCfg); err != nil {
		return err
	}

	if err = cfg.loadCdrsCfg(jsnCfg); err != nil {
		return err
	}

	if err = cfg.loadCdreCfg(jsnCfg); err != nil {
		return err
	}

	if err = cfg.loadCdrcCfg(jsnCfg); err != nil {
		return err
	}

	if err = cfg.loadSessionSCfg(jsnCfg); err != nil {
		return err
	}

	if err = cfg.loadFreeswitchAgentCfg(jsnCfg); err != nil {
		return err
	}

	if err = cfg.loadKamAgentCfg(jsnCfg); err != nil {
		return err
	}

	if err = cfg.loadAsteriskAgentCfg(jsnCfg); err != nil {
		return err
	}

	if err = cfg.loadDiameterAgentCfg(jsnCfg); err != nil {
		return err
	}

	if err = cfg.loadRadiusAgentCfg(jsnCfg); err != nil {
		return err
	}

	if err = cfg.loadDNSAgentCfg(jsnCfg); err != nil {
		return err
	}

	if err = cfg.loadHttpAgentCfg(jsnCfg); err != nil {
		return err
	}

	if err = cfg.loadAttributeServCfg(jsnCfg); err != nil {
		return err
	}

	if err = cfg.loadChargerServCfg(jsnCfg); err != nil {
		return err
	}

	if err = cfg.loadResourceSCfg(jsnCfg); err != nil {
		return err
	}

	if err = cfg.loadStatSCfg(jsnCfg); err != nil {
		return err
	}

	if err = cfg.loadThresholdSCfg(jsnCfg); err != nil {
		return err
	}

	if err = cfg.loadSupplierSCfg(jsnCfg); err != nil {
		return err
	}

	if err = cfg.loadLoaderCfg(jsnCfg); err != nil {
		return err
	}

	if err = cfg.loadMailerCfg(jsnCfg); err != nil {
		return err
	}

	if err = cfg.loadSureTaxCfg(jsnCfg); err != nil {
		return err
	}

	if err = cfg.loadDispatcherSCfg(jsnCfg); err != nil {
		return err
	}

	if err = cfg.loadLoaderCgrCfg(jsnCfg); err != nil {
		return err
	}

	if err = cfg.loadMigratorCgrCfg(jsnCfg); err != nil {
		return err
	}

	if err = cfg.loadTlsCgrCfg(jsnCfg); err != nil {
		return err
	}

	if err = cfg.loadAnalyzerCgrCfg(jsnCfg); err != nil {
		return err
	}

	if err = cfg.loadApierCfg(jsnCfg); err != nil {
		return err
	}

	if err = cfg.loadErsCfg(jsnCfg); err != nil {
		return err
	}

	return nil
}

func (cfg *CGRConfig) loadGeneralCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnGeneralCfg *GeneralJsonCfg
	if jsnGeneralCfg, err = jsnCfg.GeneralJsonCfg(); err != nil {
		return err
	}
	return cfg.generalCfg.loadFromJsonCfg(jsnGeneralCfg)
}

func (cfg *CGRConfig) loadCacheCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnCacheCfg *CacheJsonCfg
	if jsnCacheCfg, err = jsnCfg.CacheJsonCfg(); err != nil {
		return err
	}
	return cfg.cacheCfg.loadFromJsonCfg(jsnCacheCfg)
}

func (cfg *CGRConfig) loadListenCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnListenCfg *ListenJsonCfg
	if jsnListenCfg, err = jsnCfg.ListenJsonCfg(); err != nil {
		return err
	}
	return cfg.listenCfg.loadFromJsonCfg(jsnListenCfg)
}

func (cfg *CGRConfig) loadHttpCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnHttpCfg *HTTPJsonCfg
	if jsnHttpCfg, err = jsnCfg.HttpJsonCfg(); err != nil {
		return err
	}
	return cfg.httpCfg.loadFromJsonCfg(jsnHttpCfg)
}

func (cfg *CGRConfig) loadDataDBCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnDataDbCfg *DbJsonCfg
	if jsnDataDbCfg, err = jsnCfg.DbJsonCfg(DATADB_JSN); err != nil {
		return err
	}
	return cfg.dataDbCfg.loadFromJsonCfg(jsnDataDbCfg)
}

func (cfg *CGRConfig) loadStorDBCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnDataDbCfg *DbJsonCfg
	if jsnDataDbCfg, err = jsnCfg.DbJsonCfg(STORDB_JSN); err != nil {
		return err
	}
	return cfg.storDbCfg.loadFromJsonCfg(jsnDataDbCfg)
}

func (cfg *CGRConfig) loadFilterSCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnFilterSCfg *FilterSJsonCfg
	if jsnFilterSCfg, err = jsnCfg.FilterSJsonCfg(); err != nil {
		return err
	}
	return cfg.filterSCfg.loadFromJsonCfg(jsnFilterSCfg)
}

func (cfg *CGRConfig) loadRalSCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnRALsCfg *RalsJsonCfg
	if jsnRALsCfg, err = jsnCfg.RalsJsonCfg(); err != nil {
		return err
	}
	return cfg.ralsCfg.loadFromJsonCfg(jsnRALsCfg)
}

func (cfg *CGRConfig) loadSchedulerCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnSchedCfg *SchedulerJsonCfg
	if jsnSchedCfg, err = jsnCfg.SchedulerJsonCfg(); err != nil {
		return err
	}
	return cfg.schedulerCfg.loadFromJsonCfg(jsnSchedCfg)
}

func (cfg *CGRConfig) loadCdrsCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnCdrsCfg *CdrsJsonCfg
	if jsnCdrsCfg, err = jsnCfg.CdrsJsonCfg(); err != nil {
		return err
	}
	return cfg.cdrsCfg.loadFromJsonCfg(jsnCdrsCfg)
}

func (cfg *CGRConfig) loadCdreCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnCdreCfg map[string]*CdreJsonCfg
	if jsnCdreCfg, err = jsnCfg.CdreJsonCfgs(); err != nil {
		return err
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
				return err
			}
		}
	}
	return
}

func (cfg *CGRConfig) loadCdrcCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnCdrcCfg []*CdrcJsonCfg
	if jsnCdrcCfg, err = jsnCfg.CdrcJsonCfg(); err != nil {
		return err
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

func (cfg *CGRConfig) loadSessionSCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnSessionSCfg *SessionSJsonCfg
	if jsnSessionSCfg, err = jsnCfg.SessionSJsonCfg(); err != nil {
		return err
	}
	return cfg.sessionSCfg.loadFromJsonCfg(jsnSessionSCfg)
}

func (cfg *CGRConfig) loadFreeswitchAgentCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnSmFsCfg *FreeswitchAgentJsonCfg
	if jsnSmFsCfg, err = jsnCfg.FreeswitchAgentJsonCfg(); err != nil {
		return err
	}
	return cfg.fsAgentCfg.loadFromJsonCfg(jsnSmFsCfg)
}

func (cfg *CGRConfig) loadKamAgentCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnKamAgentCfg *KamAgentJsonCfg
	if jsnKamAgentCfg, err = jsnCfg.KamAgentJsonCfg(); err != nil {
		return err
	}
	return cfg.kamAgentCfg.loadFromJsonCfg(jsnKamAgentCfg)
}

func (cfg *CGRConfig) loadAsteriskAgentCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnSMAstCfg *AsteriskAgentJsonCfg
	if jsnSMAstCfg, err = jsnCfg.AsteriskAgentJsonCfg(); err != nil {
		return err
	}
	return cfg.asteriskAgentCfg.loadFromJsonCfg(jsnSMAstCfg)
}

func (cfg *CGRConfig) loadDiameterAgentCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnDACfg *DiameterAgentJsonCfg
	if jsnDACfg, err = jsnCfg.DiameterAgentJsonCfg(); err != nil {
		return err
	}
	return cfg.diameterAgentCfg.loadFromJsonCfg(jsnDACfg, cfg.GeneralCfg().RsrSepatarot)
}

func (cfg *CGRConfig) loadRadiusAgentCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnRACfg *RadiusAgentJsonCfg
	if jsnRACfg, err = jsnCfg.RadiusAgentJsonCfg(); err != nil {
		return err
	}
	return cfg.radiusAgentCfg.loadFromJsonCfg(jsnRACfg, cfg.GeneralCfg().RsrSepatarot)
}

func (cfg *CGRConfig) loadDNSAgentCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnDNSCfg *DNSAgentJsonCfg
	if jsnDNSCfg, err = jsnCfg.DNSAgentJsonCfg(); err != nil {
		return err
	}
	return cfg.dnsAgentCfg.loadFromJsonCfg(jsnDNSCfg, cfg.GeneralCfg().RsrSepatarot)
}

func (cfg *CGRConfig) loadHttpAgentCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnHttpAgntCfg *[]*HttpAgentJsonCfg
	if jsnHttpAgntCfg, err = jsnCfg.HttpAgentJsonCfg(); err != nil {
		return err
	}
	return cfg.httpAgentCfg.loadFromJsonCfg(jsnHttpAgntCfg, cfg.GeneralCfg().RsrSepatarot)
}

func (cfg *CGRConfig) loadAttributeServCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnAttributeSCfg *AttributeSJsonCfg
	if jsnAttributeSCfg, err = jsnCfg.AttributeServJsonCfg(); err != nil {
		return err
	}
	return cfg.attributeSCfg.loadFromJsonCfg(jsnAttributeSCfg)
}

func (cfg *CGRConfig) loadChargerServCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnChargerSCfg *ChargerSJsonCfg
	if jsnChargerSCfg, err = jsnCfg.ChargerServJsonCfg(); err != nil {
		return err
	}
	return cfg.chargerSCfg.loadFromJsonCfg(jsnChargerSCfg)
}

func (cfg *CGRConfig) loadResourceSCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnRLSCfg *ResourceSJsonCfg
	if jsnRLSCfg, err = jsnCfg.ResourceSJsonCfg(); err != nil {
		return err
	}
	return cfg.resourceSCfg.loadFromJsonCfg(jsnRLSCfg)
}

func (cfg *CGRConfig) loadStatSCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnStatSCfg *StatServJsonCfg
	if jsnStatSCfg, err = jsnCfg.StatSJsonCfg(); err != nil {
		return err
	}
	return cfg.statsCfg.loadFromJsonCfg(jsnStatSCfg)
}

func (cfg *CGRConfig) loadThresholdSCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnThresholdSCfg *ThresholdSJsonCfg
	if jsnThresholdSCfg, err = jsnCfg.ThresholdSJsonCfg(); err != nil {
		return err
	}
	return cfg.thresholdSCfg.loadFromJsonCfg(jsnThresholdSCfg)
}

func (cfg *CGRConfig) loadSupplierSCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnSupplierSCfg *SupplierSJsonCfg
	if jsnSupplierSCfg, err = jsnCfg.SupplierSJsonCfg(); err != nil {
		return err
	}
	return cfg.supplierSCfg.loadFromJsonCfg(jsnSupplierSCfg)
}

func (cfg *CGRConfig) loadLoaderCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnLoaderCfg []*LoaderJsonCfg
	if jsnLoaderCfg, err = jsnCfg.LoaderJsonCfg(); err != nil {
		return err
	}
	if jsnLoaderCfg != nil {
		// cfg.loaderCfg = make([]*LoaderSCfg, len(jsnLoaderCfg))
		for _, profile := range jsnLoaderCfg {
			loadSCfgp := NewDfltLoaderSCfg()
			loadSCfgp.loadFromJsonCfg(profile, cfg.GeneralCfg().RsrSepatarot)
			cfg.loaderCfg = append(cfg.loaderCfg, loadSCfgp) // use apend so the loaderS profile to be loaded from multiple files
		}
	}
	return nil
}

func (cfg *CGRConfig) loadMailerCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnMailerCfg *MailerJsonCfg
	if jsnMailerCfg, err = jsnCfg.MailerJsonCfg(); err != nil {
		return err
	}
	return cfg.mailerCfg.loadFromJsonCfg(jsnMailerCfg)
}

func (cfg *CGRConfig) loadSureTaxCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnSureTaxCfg *SureTaxJsonCfg
	if jsnSureTaxCfg, err = jsnCfg.SureTaxJsonCfg(); err != nil {
		return err
	}
	return cfg.sureTaxCfg.loadFromJsonCfg(jsnSureTaxCfg)
}

func (cfg *CGRConfig) loadDispatcherSCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnDispatcherSCfg *DispatcherSJsonCfg
	if jsnDispatcherSCfg, err = jsnCfg.DispatcherSJsonCfg(); err != nil {
		return err
	}
	return cfg.dispatcherSCfg.loadFromJsonCfg(jsnDispatcherSCfg)
}

func (cfg *CGRConfig) loadLoaderCgrCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnLoaderCgrCfg *LoaderCfgJson
	if jsnLoaderCgrCfg, err = jsnCfg.LoaderCfgJson(); err != nil {
		return err
	}
	return cfg.loaderCgrCfg.loadFromJsonCfg(jsnLoaderCgrCfg)
}

func (cfg *CGRConfig) loadMigratorCgrCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnMigratorCgrCfg *MigratorCfgJson
	if jsnMigratorCgrCfg, err = jsnCfg.MigratorCfgJson(); err != nil {
		return err
	}
	return cfg.migratorCgrCfg.loadFromJsonCfg(jsnMigratorCgrCfg)
}

func (cfg *CGRConfig) loadTlsCgrCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnTlsCgrCfg *TlsJsonCfg
	if jsnTlsCgrCfg, err = jsnCfg.TlsCfgJson(); err != nil {
		return err
	}
	return cfg.tlsCfg.loadFromJsonCfg(jsnTlsCgrCfg)
}

func (cfg *CGRConfig) loadAnalyzerCgrCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnAnalyzerCgrCfg *AnalyzerSJsonCfg
	if jsnAnalyzerCgrCfg, err = jsnCfg.AnalyzerCfgJson(); err != nil {
		return err
	}
	return cfg.analyzerSCfg.loadFromJsonCfg(jsnAnalyzerCgrCfg)
}

func (cfg *CGRConfig) loadApierCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnApierCfg *ApierJsonCfg
	if jsnApierCfg, err = jsnCfg.ApierCfgJson(); err != nil {
		return
	}
	return cfg.apier.loadFromJsonCfg(jsnApierCfg)
}

func (cfg *CGRConfig) loadErsCfg(jsnCfg *CgrJsonCfg) (err error) {
	var jsnERsCfg *ERsJsonCfg
	if jsnERsCfg, err = jsnCfg.ERsJsonCfg(); err != nil {
		return
	}
	return cfg.ersCfg.loadFromJsonCfg(jsnERsCfg, cfg.GeneralCfg().RSRSep, self.dfltEvRdr)
}

// Use locking to retrieve the configuration, possibility later for runtime reload
func (self *CGRConfig) SureTaxCfg() *SureTaxCfg {
	cfgChan := <-self.ConfigReloads[utils.SURETAX] // Lock config for read or reloads
	defer func() { self.ConfigReloads[utils.SURETAX] <- cfgChan }()
	return self.sureTaxCfg
}

func (self *CGRConfig) DiameterAgentCfg() *DiameterAgentCfg {
	cfgChan := <-self.ConfigReloads[utils.DIAMETER_AGENT] // Lock config for read or reloads
	defer func() { self.ConfigReloads[utils.DIAMETER_AGENT] <- cfgChan }()
	return self.diameterAgentCfg
}

func (self *CGRConfig) RadiusAgentCfg() *RadiusAgentCfg {
	return self.radiusAgentCfg
}

func (self *CGRConfig) DNSAgentCfg() *DNSAgentCfg {
	return self.dnsAgentCfg
}

func (cfg *CGRConfig) AttributeSCfg() *AttributeSCfg {
	return cfg.attributeSCfg
}

func (cfg *CGRConfig) ChargerSCfg() *ChargerSCfg {
	return cfg.chargerSCfg
}

// ToDo: fix locking here
func (self *CGRConfig) ResourceSCfg() *ResourceSConfig {
	return self.resourceSCfg
}

// ToDo: fix locking
func (cfg *CGRConfig) StatSCfg() *StatSCfg {
	return cfg.statsCfg
}

func (cfg *CGRConfig) ThresholdSCfg() *ThresholdSCfg {
	return cfg.thresholdSCfg
}

func (cfg *CGRConfig) SupplierSCfg() *SupplierSCfg {
	return cfg.supplierSCfg
}

func (cfg *CGRConfig) SessionSCfg() *SessionSCfg {
	return cfg.sessionSCfg
}

func (self *CGRConfig) FsAgentCfg() *FsAgentCfg {
	return self.fsAgentCfg
}

func (self *CGRConfig) KamAgentCfg() *KamAgentCfg {
	return self.kamAgentCfg
}

// ToDo: fix locking here
func (self *CGRConfig) AsteriskAgentCfg() *AsteriskAgentCfg {
	return self.asteriskAgentCfg
}

func (self *CGRConfig) HttpAgentCfg() []*HttpAgentCfg {
	return self.httpAgentCfg
}

func (cfg *CGRConfig) FilterSCfg() *FilterSCfg {
	return cfg.filterSCfg
}

func (cfg *CGRConfig) CacheCfg() CacheCfg {
	return cfg.cacheCfg
}

func (cfg *CGRConfig) LoaderCfg() []*LoaderSCfg {
	return cfg.loaderCfg
}

func (cfg *CGRConfig) DispatcherSCfg() *DispatcherSCfg {
	return cfg.dispatcherSCfg
}

func (cfg *CGRConfig) LoaderCgrCfg() *LoaderCgrCfg {
	return cfg.loaderCgrCfg
}

func (cfg *CGRConfig) MigratorCgrCfg() *MigratorCgrCfg {
	return cfg.migratorCgrCfg
}

func (cfg *CGRConfig) SchedulerCfg() *SchedulerCfg {
	return cfg.schedulerCfg
}

func (cfg *CGRConfig) DataDbCfg() *DataDbCfg {
	return cfg.dataDbCfg
}

func (cfg *CGRConfig) StorDbCfg() *StorDbCfg {
	return cfg.storDbCfg
}

func (cfg *CGRConfig) GeneralCfg() *GeneralCfg {
	return cfg.generalCfg
}

func (cfg *CGRConfig) TlsCfg() *TlsCfg {
	return cfg.tlsCfg
}

func (cfg *CGRConfig) ListenCfg() *ListenCfg {
	return cfg.listenCfg
}

func (cfg *CGRConfig) HTTPCfg() *HTTPCfg {
	return cfg.httpCfg
}

func (cfg *CGRConfig) RalsCfg() *RalsCfg {
	return cfg.ralsCfg
}

func (cfg *CGRConfig) CdrsCfg() *CdrsCfg {
	return cfg.cdrsCfg
}

func (cfg *CGRConfig) MailerCfg() *MailerCfg {
	return cfg.mailerCfg
}

func (cfg *CGRConfig) AnalyzerSCfg() *AnalyzerSCfg {
	return cfg.analyzerSCfg
}

func (cfg *CGRConfig) ApierCfg() *ApierCfg {
	return cfg.apier
}

// ERsCfg reads the EventReader configuration
func (cfg *CGRConfig) ERsCfg() *ERsCfg {
	cfg.lks[ERsJson].RLock()
	defer cfg.lks[ERsJson].RUnlock()
	return cfg.ersCfg
}

func (cfg *CGRConfig) GetReloadChan(sectID string) chan struct{} {
	return cfg.rldChans[sectID]
}

// Call implements rpcclient.RpcClientConnection interface for internal RPC
func (cSv1 *CGRConfig) Call(serviceMethod string,
	args interface{}, reply interface{}) error {
	return utils.APIerRPCCall(cSv1, serviceMethod, args, reply)
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
	case FILTERS_JSON:
		jsonString = utils.ToJSON(cfg.FilterSCfg())
	case RALS_JSN:
		jsonString = utils.ToJSON(cfg.RalsCfg())
	case SCHEDULER_JSN:
		jsonString = utils.ToJSON(cfg.SchedulerCfg())
	case CDRS_JSN:
		jsonString = utils.ToJSON(cfg.CdrsCfg())
	case SessionSJson:
		jsonString = utils.ToJSON(cfg.SessionSCfg())
	case FS_JSN:
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
	case DispatcherJson:
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
	Section string
	Path    string
}

func (cfg *CGRConfig) RLockSections() {
	for _, lk := range cfg.lks {
		lk.RLock()
	}
}

func (cfg *CGRConfig) RUnlockSections() {
	for _, lk := range cfg.lks {
		lk.RUnlock()
	}
}

func (cfg *CGRConfig) LockSections() {
	for _, lk := range cfg.lks {
		lk.Lock()
	}
}

func (cfg *CGRConfig) UnlockSections() {
	for _, lk := range cfg.lks {
		lk.Unlock()
	}
}

func (cfg *CGRConfig) V1ReloadConfig(args *ConfigReloadWithArgDispatcher, reply *string) (err error) {
	var reloadFunction func()
	if reloadFunction, err = cfg.loadConfig(args.Path, args.Section); err != nil {
		return err
	}
	//  lock all sections
	cfg.RLockSections()

	err = cfg.checkConfigSanity()

	cfg.RUnlockSections() // unlock before checking the error

	if err != nil {
		return err
	}

	reloadFunction()
	*reply = utils.OK
	return nil
}

func (cfg *CGRConfig) loadConfig(path, section string) (reload func(), err error) {
	var parseFunction func(jsnCfg *CgrJsonCfg) error
	switch section {
	case utils.EmptyString:
		cfg.LockSections()
		defer cfg.UnlockSections()
		parseFunction = cfg.loadFromJsonCfg
		reload = func() {}
	case ERsJson:
		cfg.lks[ERsJson].Lock()
		defer cfg.lks[ERsJson].Unlock()
		parseFunction = cfg.loadErsCfg
		reload = func() { cfg.rldChans[ERsJson] <- struct{}{} }
	default:
		return nil, fmt.Errorf("Invalid section: <%s>", section)
	}
	err = updateConfigFromPath(path, parseFunction)
	return
}

// Reads all .json files out of a folder/subfolders and loads them up in lexical order
func updateConfigFromPath(path string, parseFunction func(jsnCfg *CgrJsonCfg) error) error {
	if isUrl(path) {
		return updateConfigFromHttp(path, parseFunction) // prefix protocol
	}
	fi, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return utils.ErrPathNotReachable(path)
		}
		return err
	} else if !fi.IsDir() && path != utils.CONFIG_PATH { // If config dir defined, needs to exist, not checking for default
		return fmt.Errorf("Path: %s not a directory.", path)
	}
	if fi.IsDir() {
		return updateConfigFromFolder(path, parseFunction)
	}
	return nil
}

func updateConfigFromFolder(cfgDir string, parseFunction func(jsnCfg *CgrJsonCfg) error) error {
	jsonFilesFound := false
	err := filepath.Walk(cfgDir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() || isHidden(info.Name()) { // also ignore hidden files and folders
			return nil
		}
		cfgFiles, err := filepath.Glob(filepath.Join(path, "*.json"))
		if err != nil {
			return err
		}
		if cfgFiles == nil { // No need of processing further since there are no config files in the folder
			return nil
		}
		if !jsonFilesFound {
			jsonFilesFound = true
		}
		for _, jsonFilePath := range cfgFiles {
			if cgrJsonCfg, err := NewCgrJsonCfgFromFile(jsonFilePath); err != nil {
				utils.Logger.Err(fmt.Sprintf("<CGR-CFG> Error <%s> reading config from path: <%s>", err.Error(), jsonFilePath))
				return err
			} else if err := parseFunction(cgrJsonCfg); err != nil {
				utils.Logger.Err(fmt.Sprintf("<CGR-CFG> Error <%s> loading config from path: <%s>", err.Error(), jsonFilePath))
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	if !jsonFilesFound {
		return fmt.Errorf("No config file found on path %s", cfgDir)
	}
	return nil
}

func updateConfigFromHttp(urlPaths string, parseFunction func(jsnCfg *CgrJsonCfg) error) error {
	for _, urlPath := range strings.Split(urlPaths, utils.INFIELD_SEP) {
		if _, err := url.ParseRequestURI(urlPath); err != nil {
			return err
		}
		if cgrJsonCfg, err := NewCgrJsonCfgFromHttp(urlPath); err != nil {
			utils.Logger.Err(fmt.Sprintf("<CGR-CFG> Error <%s> reading config from path: <%s>", err.Error(), urlPath))
			return err
		} else if err := parseFunction(cgrJsonCfg); err != nil {
			utils.Logger.Err(fmt.Sprintf("<CGR-CFG> Error <%s> loading config from path: <%s>", err.Error(), urlPath))
			return err
		}
	}
	return nil
}
