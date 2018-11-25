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
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/cgrates/cgrates/utils"
)

var (
	DBDefaults               DbDefaults
	cgrCfg                   *CGRConfig  // will be shared
	dfltFsConnConfig         *FsConnCfg  // Default FreeSWITCH Connection configuration, built out of json default configuration
	dfltKamConnConfig        *KamConnCfg // Default Kamailio Connection configuration
	dfltHaPoolConfig         *HaPoolConfig
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

	//Depricated
	cfg.SmOsipsConfig = new(SmOsipsConfig)

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

	cgrJsonCfg, err := NewCgrJsonCfgFromReader(strings.NewReader(CGRATES_CFG_JSON))
	if err != nil {
		return nil, err
	}
	if err := cfg.loadFromJsonCfg(cgrJsonCfg); err != nil {
		return nil, err
	}

	cfg.dfltCdreProfile = cfg.CdreProfiles[utils.META_DEFAULT].Clone() // So default will stay unique, will have nil pointer in case of no defaults loaded which is an extra check
	cfg.dfltCdrcProfile = cfg.CdrcProfiles["/var/spool/cgrates/cdrc/in"][0].Clone()
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
func NewCGRConfigFromFolder(cfgDir string) (*CGRConfig, error) {
	cfg, err := NewDefaultCGRConfig()
	if err != nil {

		return nil, err
	}
	fi, err := os.Stat(cfgDir)
	if err != nil {
		if strings.HasSuffix(err.Error(), "no such file or directory") {
			return cfg, nil
		}
		return nil, err
	} else if !fi.IsDir() && cfgDir != utils.CONFIG_DIR { // If config dir defined, needs to exist, not checking for default
		return nil, fmt.Errorf("Path: %s not a directory.", cfgDir)
	}
	if fi.IsDir() {
		jsonFilesFound := false
		err = filepath.Walk(cfgDir, func(path string, info os.FileInfo, err error) error {
			if !info.IsDir() {
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
				} else if err := cfg.loadFromJsonCfg(cgrJsonCfg); err != nil {
					utils.Logger.Err(fmt.Sprintf("<CGR-CFG> Error <%s> loading config from path: <%s>", err.Error(), jsonFilePath))
					return err
				}
			}
			return nil
		})
		if err != nil {
			return nil, err
		} else if !jsonFilesFound {
			return nil, fmt.Errorf("No config file found on path %s", cfgDir)
		}
	}
	if err := cfg.checkConfigSanity(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// Holds system configuration, defaults are overwritten with values from config file if found
type CGRConfig struct {
	MaxCallDuration time.Duration // The maximum call duration (used by responder when querying DerivedCharging) // ToDo: export it in configuration file
	DataFolderPath  string        // Path towards data folder, for tests internal usage, not loading out of .json options

	// Cache defaults loaded from json and needing clones
	dfltCdreProfile *CdreCfg // Default cdreConfig profile
	dfltCdrcProfile *CdrcCfg // Default cdrcConfig profile

	CdreProfiles map[string]*CdreCfg   // Cdre config profiles
	CdrcProfiles map[string][]*CdrcCfg // Number of CDRC instances running imports, format map[dirPath][]{Configs}
	loaderCfg    []*LoaderSCfg         // LoaderS configs
	httpAgentCfg HttpAgentCfgs         // HttpAgent configs

	ConfigReloads map[string]chan struct{} // Signals to specific entities that a config reload should occur

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

	// Deprecated
	SmOsipsConfig        *SmOsipsConfig // SMOpenSIPS Configuration
	PubSubServerEnabled  bool           // Starts PubSub as server: <true|false>.
	AliasesServerEnabled bool           // Starts PubSub as server: <true|false>.
	UserServerEnabled    bool           // Starts User as server: <true|false>
	UserServerIndexes    []string       // List of user profile field indexes
}

func (self *CGRConfig) checkConfigSanity() error {
	// Rater checks
	if self.ralsCfg.RALsEnabled {
		if !self.statsCfg.Enabled {
			for _, connCfg := range self.ralsCfg.RALsStatSConns {
				if connCfg.Address == utils.MetaInternal {
					return errors.New("StatS not enabled but requested by RALs component.")
				}
			}
		}
		if !self.PubSubServerEnabled {
			for _, connCfg := range self.ralsCfg.RALsPubSubSConns {
				if connCfg.Address == utils.MetaInternal {
					return errors.New("PubSub server not enabled but requested by RALs component.")
				}
			}
		}
		if !self.AliasesServerEnabled {
			for _, connCfg := range self.ralsCfg.RALsAliasSConns {
				if connCfg.Address == utils.MetaInternal {
					return errors.New("Alias server not enabled but requested by RALs component.")
				}
			}
		}
		if !self.UserServerEnabled {
			for _, connCfg := range self.ralsCfg.RALsUserSConns {
				if connCfg.Address == utils.MetaInternal {
					return errors.New("User service not enabled but requested by RALs component.")
				}
			}
		}
		if !self.thresholdSCfg.Enabled {
			for _, connCfg := range self.ralsCfg.RALsThresholdSConns {
				if connCfg.Address == utils.MetaInternal {
					return errors.New("ThresholdS not enabled but requested by RALs component.")
				}
			}
		}
	}
	// CDRServer checks
	if self.cdrsCfg.CDRSEnabled {
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
		if !self.PubSubServerEnabled {
			for _, connCfg := range self.cdrsCfg.CDRSPubSubSConns {
				if connCfg.Address == utils.MetaInternal {
					return errors.New("PubSubS not enabled but requested by CDRS component.")
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
		if !self.UserServerEnabled {
			for _, connCfg := range self.cdrsCfg.CDRSUserSConns {
				if connCfg.Address == utils.MetaInternal {
					return errors.New("UserS not enabled but requested by CDRS component.")
				}
			}
		}
		if !self.AliasesServerEnabled {
			for _, connCfg := range self.cdrsCfg.CDRSAliaseSConns {
				if connCfg.Address == utils.MetaInternal {
					return errors.New("AliaseS not enabled but requested by CDRS component.")
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
			if !self.cdrsCfg.CDRSEnabled {
				for _, conn := range cdrcInst.CdrsConns {
					if conn.Address == utils.MetaInternal {
						return errors.New("CDRS not enabled but referenced from CDRC")
					}
				}
			}
			if len(cdrcInst.ContentFields) == 0 {
				return errors.New("CdrC enabled but no fields to be processed defined!")
			}
			if cdrcInst.CdrFormat == utils.CSV {
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
			if !utils.IsSliceMember([]string{utils.MetaAttributes,
				utils.MetaResources, utils.MetaFilters, utils.MetaStats,
				utils.MetaSuppliers, utils.MetaThresholds}, data.Type) {
				return fmt.Errorf("<%s> unsupported data type %s", utils.LoaderS, data.Type)
			}

			for _, field := range data.Fields {
				if field.Type != utils.META_COMPOSED && field.Type != utils.MetaString {
					return fmt.Errorf("<%s> invalid field type %s for %s at %s", utils.LoaderS, field.Type, data.Type, field.Tag)
				}
			}
		}
	}
	// SessionS checks
	if self.sessionSCfg.Enabled {
		if len(self.sessionSCfg.RALsConns) == 0 {
			return errors.New("<SessionS> RALs definition is mandatory!")
		}
		if !self.chargerSCfg.Enabled {
			for _, conn := range self.sessionSCfg.ChargerSConns {
				if conn.Address == utils.MetaInternal {
					return errors.New("<SessionS> ChargerS not enabled but requested")
				}
			}
		}
		if !self.ralsCfg.RALsEnabled {
			for _, smgRALsConn := range self.sessionSCfg.RALsConns {
				if smgRALsConn.Address == utils.MetaInternal {
					return errors.New("<SessionS> RALs not enabled but requested by SMGeneric component.")
				}
			}
		}
		if !self.resourceSCfg.Enabled {
			for _, conn := range self.sessionSCfg.ResSConns {
				if conn.Address == utils.MetaInternal {
					return errors.New("<SessionS> ResourceS not enabled but requested by SMGeneric component.")
				}
			}
		}
		if !self.thresholdSCfg.Enabled {
			for _, conn := range self.sessionSCfg.ThreshSConns {
				if conn.Address == utils.MetaInternal {
					return errors.New("<SessionS> ThresholdS not enabled but requested by SMGeneric component.")
				}
			}
		}
		if !self.statsCfg.Enabled {
			for _, conn := range self.sessionSCfg.StatSConns {
				if conn.Address == utils.MetaInternal {
					return errors.New("<SessionS> StatS not enabled but requested by SMGeneric component.")
				}
			}
		}
		if !self.supplierSCfg.Enabled {
			for _, conn := range self.sessionSCfg.SupplSConns {
				if conn.Address == utils.MetaInternal {
					return errors.New("<SessionS> SupplierS not enabled but requested by SMGeneric component.")
				}
			}
		}
		if !self.attributeSCfg.Enabled {
			for _, conn := range self.sessionSCfg.AttrSConns {
				if conn.Address == utils.MetaInternal {
					return errors.New("<SessionS> AttributeS not enabled but requested by SMGeneric component.")
				}
			}
		}
		if len(self.sessionSCfg.CDRsConns) == 0 {
			return errors.New("<SessionS> CDRs definition is mandatory!")
		}
		if !self.cdrsCfg.CDRSEnabled {
			for _, smgCDRSConn := range self.sessionSCfg.CDRsConns {
				if smgCDRSConn.Address == utils.MetaInternal {
					return errors.New("<SessionS> CDRS not enabled but referenced by SMGeneric component")
				}
			}
		}
	}
	// FreeSWITCHAgent checks
	if self.fsAgentCfg.Enabled {
		for _, connCfg := range self.fsAgentCfg.SessionSConns {
			if connCfg.Address != utils.MetaInternal {
				return errors.New("only <*internal> connectivity allowed in in <freeswitch_agent> towards <sessions> for now")
			}
			if connCfg.Address == utils.MetaInternal &&
				!self.sessionSCfg.Enabled {
				return errors.New("<sessions> not enabled but referenced by <freeswitch_agent>")
			}
		}
	}
	// KamailioAgent checks
	if self.kamAgentCfg.Enabled {
		for _, connCfg := range self.kamAgentCfg.SessionSConns {
			if connCfg.Address != utils.MetaInternal {
				return errors.New("only <*internal> connectivity allowed in in <kamailio_agent> towards <sessions> for now")
			}
			if connCfg.Address == utils.MetaInternal &&
				!self.sessionSCfg.Enabled {
				return errors.New("<sessions> not enabled but referenced by <kamailio_agent>")
			}
		}
	}
	// SMOpenSIPS checks
	if self.SmOsipsConfig.Enabled {
		if len(self.SmOsipsConfig.RALsConns) == 0 {
			return errors.New("<SMOpenSIPS> Rater definition is mandatory!")
		}
		if !self.ralsCfg.RALsEnabled {
			for _, smOsipsRaterConn := range self.SmOsipsConfig.RALsConns {
				if smOsipsRaterConn.Address == utils.MetaInternal {
					return errors.New("<SMOpenSIPS> RALs not enabled.")
				}
			}
		}
		if len(self.SmOsipsConfig.CDRsConns) == 0 {
			return errors.New("<SMOpenSIPS> CDRs definition is mandatory!")
		}
		if !self.cdrsCfg.CDRSEnabled {
			for _, smOsipsCDRSConn := range self.SmOsipsConfig.CDRsConns {
				if smOsipsCDRSConn.Address == utils.MetaInternal {
					return errors.New("<SMOpenSIPS> CDRS not enabled.")
				}
			}
		}
	}
	// AsteriskAgent checks
	if self.asteriskAgentCfg.Enabled {
		/*if len(self.asteriskAgentCfg.SessionSConns) == 0 {
			return errors.New("<SMAsterisk> SMG definition is mandatory!")
		}
		for _, smAstSMGConn := range self.asteriskAgentCfg.SessionSConns {
			if smAstSMGConn.Address == utils.MetaInternal && !self.sessionSCfg.Enabled {
				return errors.New("<SMAsterisk> SMG not enabled.")
			}
		}
		*/
		if !self.sessionSCfg.Enabled {
			return errors.New("<SMAsterisk> SMG not enabled.")
		}
	}
	// DAgent checks
	if self.diameterAgentCfg.Enabled && !self.sessionSCfg.Enabled {
		for _, daSMGConn := range self.diameterAgentCfg.SessionSConns {
			if daSMGConn.Address == utils.MetaInternal {
				return fmt.Errorf("%s not enabled but referenced by %s component",
					utils.SessionS, utils.DiameterAgent)
			}
		}
	}
	if self.radiusAgentCfg.Enabled && !self.sessionSCfg.Enabled {
		for _, raSMGConn := range self.radiusAgentCfg.SessionSConns {
			if raSMGConn.Address == utils.MetaInternal {
				return errors.New("SMGeneric not enabled but referenced by RadiusAgent component")
			}
		}
	}
	// HTTPAgent checks
	for _, httpAgentCfg := range self.httpAgentCfg {
		// httpAgent checks
		for _, sSConn := range httpAgentCfg.SessionSConns {
			if sSConn.Address == utils.MetaInternal && self.sessionSCfg.Enabled {
				return errors.New("SessionS not enabled but referenced by HttpAgent component")
			}
		}
		if !utils.IsSliceMember([]string{utils.MetaUrl, utils.MetaXml}, httpAgentCfg.RequestPayload) {
			return fmt.Errorf("<%s> unsupported request payload %s",
				utils.HTTPAgent, httpAgentCfg.RequestPayload)
		}
		if !utils.IsSliceMember([]string{utils.MetaXml}, httpAgentCfg.ReplyPayload) {
			return fmt.Errorf("<%s> unsupported reply payload %s",
				utils.HTTPAgent, httpAgentCfg.ReplyPayload)
		}
	}
	if self.attributeSCfg.Enabled {
		if self.attributeSCfg.ProcessRuns < 1 {
			return fmt.Errorf("<%s> process_runs needs to be bigger than 0", utils.AttributeS)
		}
	}
	if self.chargerSCfg.Enabled {
		for _, connCfg := range self.chargerSCfg.AttributeSConns {
			if connCfg.Address == utils.MetaInternal &&
				(self.attributeSCfg == nil || !self.attributeSCfg.Enabled) {
				return errors.New("AttributeS not enabled but requested by ChargerS component.")
			}
		}
	}
	// ResourceLimiter checks
	if self.resourceSCfg.Enabled && !self.thresholdSCfg.Enabled {
		for _, connCfg := range self.resourceSCfg.ThresholdSConns {
			if connCfg.Address == utils.MetaInternal {
				return errors.New("ThresholdS not enabled but requested by ResourceS component.")
			}
		}
	}
	// StatS checks
	if self.statsCfg.Enabled && !self.thresholdSCfg.Enabled {
		for _, connCfg := range self.statsCfg.ThresholdSConns {
			if connCfg.Address == utils.MetaInternal {
				return errors.New("ThresholdS not enabled but requested by StatS component.")
			}
		}
	}
	// SupplierS checks
	if self.supplierSCfg.Enabled {
		for _, connCfg := range self.supplierSCfg.RALsConns {
			if connCfg.Address != utils.MetaInternal {
				return errors.New("Only <*internal> RALs connectivity allowed in SupplierS for now")
			}
			if connCfg.Address == utils.MetaInternal && !self.ralsCfg.RALsEnabled {
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
	// DispaterS checks
	if self.dispatcherSCfg.Enabled {
		if !utils.IsSliceMember([]string{utils.MetaFirst, utils.MetaRandom, utils.MetaNext,
			utils.MetaBroadcast}, self.dispatcherSCfg.DispatchingStrategy) {
			return fmt.Errorf("<%s> unsupported dispatching strategy %s",
				utils.DispatcherS, self.dispatcherSCfg.DispatchingStrategy)
		}
	}
	// Scheduler check connection with CDR Server
	if !self.cdrsCfg.CDRSEnabled {
		for _, connCfg := range self.schedulerCfg.CDRsConns {
			if connCfg.Address == utils.MetaInternal {
				return errors.New("CDR Server not enabled but requested by Scheduler")
			}
		}
	}
	return nil
}

// Loads from json configuration object, will be used for defaults, config from file and reload, might need lock
func (self *CGRConfig) loadFromJsonCfg(jsnCfg *CgrJsonCfg) (err error) {
	// Load sections out of JSON config, stop on error
	jsnGeneralCfg, err := jsnCfg.GeneralJsonCfg()
	if err != nil {
		return err
	}
	if err := self.generalCfg.loadFromJsonCfg(jsnGeneralCfg); err != nil {
		return err
	}

	jsnCacheCfg, err := jsnCfg.CacheJsonCfg()
	if err != nil {
		return err
	}
	if err := self.cacheCfg.loadFromJsonCfg(jsnCacheCfg); err != nil {
		return err
	}

	jsnListenCfg, err := jsnCfg.ListenJsonCfg()
	if err != nil {
		return err
	}
	if err := self.listenCfg.loadFromJsonCfg(jsnListenCfg); err != nil {
		return err
	}

	jsnHttpCfg, err := jsnCfg.HttpJsonCfg()
	if err != nil {
		return err
	}
	if err := self.httpCfg.loadFromJsonCfg(jsnHttpCfg); err != nil {
		return err
	}

	jsnDataDbCfg, err := jsnCfg.DbJsonCfg(DATADB_JSN)
	if err != nil {
		return err
	}
	if err := self.dataDbCfg.loadFromJsonCfg(jsnDataDbCfg); err != nil {
		return err
	}

	jsnStorDbCfg, err := jsnCfg.DbJsonCfg(STORDB_JSN)
	if err != nil {
		return err
	}
	if err := self.storDbCfg.loadFromJsonCfg(jsnStorDbCfg); err != nil {
		return err
	}

	jsnFilterSCfg, err := jsnCfg.FilterSJsonCfg()
	if err != nil {
		return err
	}
	if err = self.filterSCfg.loadFromJsonCfg(jsnFilterSCfg); err != nil {
		return
	}

	jsnRALsCfg, err := jsnCfg.RalsJsonCfg()
	if err != nil {
		return err
	}
	if err = self.ralsCfg.loadFromJsonCfg(jsnRALsCfg); err != nil {
		return
	}

	jsnSchedCfg, err := jsnCfg.SchedulerJsonCfg()
	if err != nil {
		return err
	}
	if err := self.schedulerCfg.loadFromJsonCfg(jsnSchedCfg); err != nil {
		return err
	}

	jsnCdrsCfg, err := jsnCfg.CdrsJsonCfg()
	if err != nil {
		return err
	}
	if err := self.cdrsCfg.loadFromJsonCfg(jsnCdrsCfg); err != nil {
		return err
	}

	jsnCdreCfg, err := jsnCfg.CdreJsonCfgs()
	if err != nil {
		return err
	}

	jsnCdrcCfg, err := jsnCfg.CdrcJsonCfg()
	if err != nil {
		return err
	}

	jsnsessionSCfg, err := jsnCfg.SessionSJsonCfg()
	if err != nil {
		return err
	}
	if err := self.sessionSCfg.loadFromJsonCfg(jsnsessionSCfg); err != nil {
		return err
	}

	jsnSmFsCfg, err := jsnCfg.FreeswitchAgentJsonCfg()
	if err != nil {
		return err
	}
	if err := self.fsAgentCfg.loadFromJsonCfg(jsnSmFsCfg); err != nil {
		return err
	}

	jsnKamAgentCfg, err := jsnCfg.KamAgentJsonCfg()
	if err != nil {
		return err
	}
	if err := self.kamAgentCfg.loadFromJsonCfg(jsnKamAgentCfg); err != nil {
		return err
	}

	jsnSMAstCfg, err := jsnCfg.AsteriskAgentJsonCfg()
	if err != nil {
		return err
	}
	if err := self.asteriskAgentCfg.loadFromJsonCfg(jsnSMAstCfg); err != nil {
		return err
	}

	jsnDACfg, err := jsnCfg.DiameterAgentJsonCfg()
	if err != nil {
		return err
	}
	if err := self.diameterAgentCfg.loadFromJsonCfg(jsnDACfg, self.generalCfg.RsrSepatarot); err != nil {
		return err
	}

	jsnRACfg, err := jsnCfg.RadiusAgentJsonCfg()
	if err != nil {
		return err
	}
	if err := self.radiusAgentCfg.loadFromJsonCfg(jsnRACfg, self.generalCfg.RsrSepatarot); err != nil {
		return err
	}

	jsnHttpAgntCfg, err := jsnCfg.HttpAgentJsonCfg()
	if err != nil {
		return err
	}
	if err := self.httpAgentCfg.loadFromJsonCfg(jsnHttpAgntCfg, self.generalCfg.RsrSepatarot); err != nil {
		return err
	}

	jsnAttributeSCfg, err := jsnCfg.AttributeServJsonCfg()
	if err != nil {
		return err
	}
	if self.attributeSCfg.loadFromJsonCfg(jsnAttributeSCfg); err != nil {
		return err
	}

	jsnChargerSCfg, err := jsnCfg.ChargerServJsonCfg()
	if err != nil {
		return err
	}
	if self.chargerSCfg.loadFromJsonCfg(jsnChargerSCfg); err != nil {
		return err
	}

	jsnRLSCfg, err := jsnCfg.ResourceSJsonCfg()
	if err != nil {
		return err
	}
	if self.resourceSCfg.loadFromJsonCfg(jsnRLSCfg); err != nil {
		return err
	}

	jsnStatSCfg, err := jsnCfg.StatSJsonCfg()
	if err != nil {
		return err
	}
	if self.statsCfg.loadFromJsonCfg(jsnStatSCfg); err != nil {
		return err
	}

	jsnThresholdSCfg, err := jsnCfg.ThresholdSJsonCfg()
	if err != nil {
		return err
	}
	if self.thresholdSCfg.loadFromJsonCfg(jsnThresholdSCfg); err != nil {
		return err
	}

	jsnSupplierSCfg, err := jsnCfg.SupplierSJsonCfg()
	if err != nil {
		return err
	}
	if self.supplierSCfg.loadFromJsonCfg(jsnSupplierSCfg); err != nil {
		return err
	}

	jsnLoaderCfg, err := jsnCfg.LoaderJsonCfg()
	if err != nil {
		return err
	}

	jsnMailerCfg, err := jsnCfg.MailerJsonCfg()
	if err != nil {
		return err
	}
	if self.mailerCfg.loadFromJsonCfg(jsnMailerCfg); err != nil {
		return err
	}

	jsnSureTaxCfg, err := jsnCfg.SureTaxJsonCfg()
	if err != nil {
		return err
	}
	if err := self.sureTaxCfg.loadFromJsonCfg(jsnSureTaxCfg); err != nil {
		return err
	}

	jsnDispatcherCfg, err := jsnCfg.DispatcherSJsonCfg()
	if err != nil {
		return err
	}
	if self.dispatcherSCfg.loadFromJsonCfg(jsnDispatcherCfg); err != nil {
		return err
	}

	jsnLoaderCgrCfg, err := jsnCfg.LoaderCfgJson()
	if err != nil {
		return nil
	}
	if self.loaderCgrCfg.loadFromJsonCfg(jsnLoaderCgrCfg); err != nil {
		return err
	}

	jsnMigratorCgrCfg, err := jsnCfg.MigratorCfgJson()
	if err != nil {
		return nil
	}
	if self.migratorCgrCfg.loadFromJsonCfg(jsnMigratorCgrCfg); err != nil {
		return err
	}

	jsnTlsCgrCfg, err := jsnCfg.TlsCfgJson()
	if err != nil {
		return nil
	}
	if err := self.tlsCfg.loadFromJsonCfg(jsnTlsCgrCfg); err != nil {
		return err
	}

	jsnAnalyzerCgrCfg, err := jsnCfg.AnalyzerCfgJson()
	if err != nil {
		return nil
	}
	if err := self.analyzerSCfg.loadFromJsonCfg(jsnAnalyzerCgrCfg); err != nil {
		return err
	}

	if jsnCdreCfg != nil {
		for profileName, jsnCdre1Cfg := range jsnCdreCfg {
			if _, hasProfile := self.CdreProfiles[profileName]; !hasProfile { // New profile, create before loading from json
				self.CdreProfiles[profileName] = new(CdreCfg)
				if profileName != utils.META_DEFAULT {
					self.CdreProfiles[profileName] = self.dfltCdreProfile.Clone() // Clone default so we do not inherit pointers
				}
			}
			if err = self.CdreProfiles[profileName].loadFromJsonCfg(jsnCdre1Cfg, self.generalCfg.RsrSepatarot); err != nil { // Update the existing profile with content from json config
				return err
			}
		}
	}

	if jsnLoaderCfg != nil {
		// self.loaderCfg = make([]*LoaderSCfg, len(jsnLoaderCfg))
		for _, profile := range jsnLoaderCfg {
			loadSCfgp := NewDfltLoaderSCfg()
			loadSCfgp.loadFromJsonCfg(profile, self.generalCfg.RsrSepatarot)
			self.loaderCfg = append(self.loaderCfg, loadSCfgp) // use apend so the loaderS profile to be loaded from multiple files
		}
	}

	if jsnCdrcCfg != nil {
		for _, jsnCrc1Cfg := range jsnCdrcCfg {
			if jsnCrc1Cfg.Id == nil || *jsnCrc1Cfg.Id == "" {
				return utils.ErrCDRCNoProfileID
			}
			if *jsnCrc1Cfg.Id == utils.META_DEFAULT {
				if self.dfltCdrcProfile == nil {
					self.dfltCdrcProfile = new(CdrcCfg)
				}
			}
			indxFound := -1 // Will be different than -1 if an instance with same id will be found
			pathFound := "" // Will be populated with the path where slice of cfgs was found
			var cdrcInstCfg *CdrcCfg
			for path := range self.CdrcProfiles {
				for i := range self.CdrcProfiles[path] {
					if self.CdrcProfiles[path][i].ID == *jsnCrc1Cfg.Id {
						indxFound = i
						pathFound = path
						cdrcInstCfg = self.CdrcProfiles[path][i]
						break
					}
				}
			}
			if cdrcInstCfg == nil {
				cdrcInstCfg = self.dfltCdrcProfile.Clone()
			}
			if err := cdrcInstCfg.loadFromJsonCfg(jsnCrc1Cfg, self.generalCfg.RsrSepatarot); err != nil {
				return err
			}
			if cdrcInstCfg.CdrInDir == "" {
				return utils.ErrCDRCNoInDir
			}
			if _, hasDir := self.CdrcProfiles[cdrcInstCfg.CdrInDir]; !hasDir {
				self.CdrcProfiles[cdrcInstCfg.CdrInDir] = make([]*CdrcCfg, 0)
			}
			if indxFound != -1 { // Replace previous config so we have inheritance
				self.CdrcProfiles[pathFound][indxFound] = cdrcInstCfg
			} else {
				self.CdrcProfiles[cdrcInstCfg.CdrInDir] = append(self.CdrcProfiles[cdrcInstCfg.CdrInDir], cdrcInstCfg)
			}
		}
	}

	//Depricated

	jsnPubSubServCfg, err := jsnCfg.PubSubServJsonCfg()
	if err != nil {
		return err
	}
	if jsnPubSubServCfg != nil {
		if jsnPubSubServCfg.Enabled != nil {
			self.PubSubServerEnabled = *jsnPubSubServCfg.Enabled
		}
	}

	jsnAliasesServCfg, err := jsnCfg.AliasesServJsonCfg()
	if err != nil {
		return err
	}
	if jsnAliasesServCfg != nil {
		if jsnAliasesServCfg.Enabled != nil {
			self.AliasesServerEnabled = *jsnAliasesServCfg.Enabled
		}
	}

	jsnUserServCfg, err := jsnCfg.UserServJsonCfg()
	if err != nil {
		return err
	}
	if jsnUserServCfg != nil {
		if jsnUserServCfg.Enabled != nil {
			self.UserServerEnabled = *jsnUserServCfg.Enabled
		}
		if jsnUserServCfg.Indexes != nil {
			self.UserServerIndexes = *jsnUserServCfg.Indexes
		}
	}
	///depricated^^^
	return nil
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
