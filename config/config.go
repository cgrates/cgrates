/*
Real-time Charging System for Telecom & ISP environments
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

const (
	DISABLED = "disabled"
	JSON     = "json"
	GOB      = "gob"
	POSTGRES = "postgres"
	MONGO    = "mongo"
	REDIS    = "redis"
	SAME     = "same"
	FS       = "freeswitch"
)

var (
	cgrCfg            *CGRConfig     // will be shared
	dfltFsConnConfig  *FsConnConfig  // Default FreeSWITCH Connection configuration, built out of json default configuration
	dfltKamConnConfig *KamConnConfig // Default Kamailio Connection configuration
	dfltHaPoolConfig  *HaPoolConfig
)

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
	cfg.InstanceID = utils.GenUUID()
	cfg.DataFolderPath = "/usr/share/cgrates/"
	cfg.SmGenericConfig = new(SmGenericConfig)
	cfg.SmFsConfig = new(SmFsConfig)
	cfg.SmKamConfig = new(SmKamConfig)
	cfg.SmOsipsConfig = new(SmOsipsConfig)
	cfg.diameterAgentCfg = new(DiameterAgentCfg)
	cfg.ConfigReloads = make(map[string]chan struct{})
	cfg.ConfigReloads[utils.CDRC] = make(chan struct{}, 1)
	cfg.ConfigReloads[utils.CDRC] <- struct{}{} // Unlock the channel
	cfg.ConfigReloads[utils.CDRE] = make(chan struct{}, 1)
	cfg.ConfigReloads[utils.CDRE] <- struct{}{} // Unlock the channel
	cfg.ConfigReloads[utils.SURETAX] = make(chan struct{}, 1)
	cfg.ConfigReloads[utils.SURETAX] <- struct{}{} // Unlock the channel
	cfg.ConfigReloads[utils.DIAMETER_AGENT] = make(chan struct{}, 1)
	cfg.ConfigReloads[utils.DIAMETER_AGENT] <- struct{}{} // Unlock the channel
	cgrJsonCfg, err := NewCgrJsonCfgFromReader(strings.NewReader(CGRATES_CFG_JSON))
	if err != nil {
		return nil, err
	}
	cfg.MaxCallDuration = time.Duration(3) * time.Hour // Hardcoded for now
	if err := cfg.loadFromJsonCfg(cgrJsonCfg); err != nil {
		return nil, err
	}
	cfg.dfltCdreProfile = cfg.CdreProfiles[utils.META_DEFAULT].Clone() // So default will stay unique, will have nil pointer in case of no defaults loaded which is an extra check
	cfg.dfltCdrcProfile = cfg.CdrcProfiles["/var/spool/cgrates/cdrc/in"][0].Clone()
	dfltFsConnConfig = cfg.SmFsConfig.EventSocketConns[0] // We leave it crashing here on purpose if no Connection defaults defined
	dfltKamConnConfig = cfg.SmKamConfig.EvapiConns[0]
	if err := cfg.checkConfigSanity(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func NewCGRConfigFromJsonString(cfgJsonStr string) (*CGRConfig, error) {
	cfg := new(CGRConfig)
	if jsnCfg, err := NewCgrJsonCfgFromReader(strings.NewReader(cfgJsonStr)); err != nil {
		return nil, err
	} else if err := cfg.loadFromJsonCfg(jsnCfg); err != nil {
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
					return err
				} else if err := cfg.loadFromJsonCfg(cgrJsonCfg); err != nil {
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
	InstanceID               string // Identifier for this engine instance
	TpDbType                 string
	TpDbHost                 string // The host to connect to. Values that start with / are for UNIX domain sockets.
	TpDbPort                 string // The port to bind to.
	TpDbName                 string // The name of the database to connect to.
	TpDbUser                 string // The user to sign in as.
	TpDbPass                 string // The user's password.
	DataDbType               string
	DataDbHost               string // The host to connect to. Values that start with / are for UNIX domain sockets.
	DataDbPort               string // The port to bind to.
	DataDbName               string // The name of the database to connect to.
	DataDbUser               string // The user to sign in as.
	DataDbPass               string // The user's password.
	LoadHistorySize          int    // Maximum number of records to archive in load history
	StorDBType               string // Should reflect the database type used to store logs
	StorDBHost               string // The host to connect to. Values that start with / are for UNIX domain sockets.
	StorDBPort               string // Th e port to bind to.
	StorDBName               string // The name of the database to connect to.
	StorDBUser               string // The user to sign in as.
	StorDBPass               string // The user's password.
	StorDBMaxOpenConns       int    // Maximum database connections opened
	StorDBMaxIdleConns       int    // Maximum idle connections to keep opened
	StorDBCDRSIndexes        []string
	DBDataEncoding           string        // The encoding used to store object data in strings: <msgpack|json>
	RPCJSONListen            string        // RPC JSON listening address
	RPCGOBListen             string        // RPC GOB listening address
	HTTPListen               string        // HTTP listening address
	DefaultReqType           string        // Use this request type if not defined on top
	DefaultCategory          string        // set default type of record
	DefaultTenant            string        // set default tenant
	DefaultTimezone          string        // default timezone for timestamps where not specified <""|UTC|Local|$IANA_TZ_DB>
	Reconnects               int           // number of recconect attempts in case of connection lost <-1 for infinite | nb>
	ConnectTimeout           time.Duration // timeout for RPC connection attempts
	ReplyTimeout             time.Duration // timeout replies if not reaching back
	ConnectAttempts          int           // number of initial connection attempts before giving up
	ResponseCacheTTL         time.Duration // the life span of a cached response
	InternalTtl              time.Duration // maximum duration to wait for internal connections before giving up
	RoundingDecimals         int           // Number of decimals to round end prices at
	HttpSkipTlsVerify        bool          // If enabled Http Client will accept any TLS certificate
	TpExportPath             string        // Path towards export folder for offline Tariff Plans
	HttpPosterAttempts       int
	HttpFailedDir            string          // Directory path where we store failed http requests
	MaxCallDuration          time.Duration   // The maximum call duration (used by responder when querying DerivedCharging) // ToDo: export it in configuration file
	LockingTimeout           time.Duration   // locking mechanism timeout to avoid deadlocks
	CacheDumpDir             string          // cache dump for faster start (leave empty to disable)b
	RALsEnabled              bool            // start standalone server (no balancer)
	RALsBalancer             string          // balancer address host:port
	RALsCDRStatSConns        []*HaPoolConfig // address where to reach the cdrstats service. Empty to disable stats gathering  <""|internal|x.y.z.y:1234>
	RALsHistorySConns        []*HaPoolConfig
	RALsPubSubSConns         []*HaPoolConfig
	RALsUserSConns           []*HaPoolConfig
	RALsAliasSConns          []*HaPoolConfig
	RpSubjectPrefixMatching  bool // enables prefix matching for the rating profile subject
	LcrSubjectPrefixMatching bool // enables prefix matching for the lcr subject
	BalancerEnabled          bool
	SchedulerEnabled         bool
	CDRSEnabled              bool                 // Enable CDR Server service
	CDRSExtraFields          []*utils.RSRField    // Extra fields to store in CDRs
	CDRSStoreCdrs            bool                 // store cdrs in storDb
	CDRSRaterConns           []*HaPoolConfig      // address where to reach the Rater for cost calculation: <""|internal|x.y.z.y:1234>
	CDRSPubSubSConns         []*HaPoolConfig      // address where to reach the pubsub service: <""|internal|x.y.z.y:1234>
	CDRSUserSConns           []*HaPoolConfig      // address where to reach the users service: <""|internal|x.y.z.y:1234>
	CDRSAliaseSConns         []*HaPoolConfig      // address where to reach the aliases service: <""|internal|x.y.z.y:1234>
	CDRSStatSConns           []*HaPoolConfig      // address where to reach the cdrstats service. Empty to disable stats gathering  <""|internal|x.y.z.y:1234>
	CDRSCdrReplication       []*CdrReplicationCfg // Replicate raw CDRs to a number of servers
	CDRStatsEnabled          bool                 // Enable CDR Stats service
	CDRStatsSaveInterval     time.Duration        // Save interval duration
	CdreProfiles             map[string]*CdreConfig
	CdrcProfiles             map[string][]*CdrcConfig // Number of CDRC instances running imports, format map[dirPath][]{Configs}
	SmGenericConfig          *SmGenericConfig
	SmFsConfig               *SmFsConfig              // SMFreeSWITCH configuration
	SmKamConfig              *SmKamConfig             // SM-Kamailio Configuration
	SmOsipsConfig            *SmOsipsConfig           // SMOpenSIPS Configuration
	diameterAgentCfg         *DiameterAgentCfg        // DiameterAgent configuration
	HistoryServer            string                   // Address where to reach the master history server: <internal|x.y.z.y:1234>
	HistoryServerEnabled     bool                     // Starts History as server: <true|false>.
	HistoryDir               string                   // Location on disk where to store history files.
	HistorySaveInterval      time.Duration            // The timout duration between pubsub writes
	PubSubServerEnabled      bool                     // Starts PubSub as server: <true|false>.
	AliasesServerEnabled     bool                     // Starts PubSub as server: <true|false>.
	UserServerEnabled        bool                     // Starts User as server: <true|false>
	UserServerIndexes        []string                 // List of user profile field indexes
	ResourceLimiterCfg       *ResourceLimiterConfig   // Configuration for resource limiter
	MailerServer             string                   // The server to use when sending emails out
	MailerAuthUser           string                   // Authenticate to email server using this user
	MailerAuthPass           string                   // Authenticate to email server with this password
	MailerFromAddr           string                   // From address used when sending emails out
	DataFolderPath           string                   // Path towards data folder, for tests internal usage, not loading out of .json options
	sureTaxCfg               *SureTaxCfg              // Load here SureTax configuration, as pointer so we can have runtime reloads in the future
	ConfigReloads            map[string]chan struct{} // Signals to specific entities that a config reload should occur
	// Cache defaults loaded from json and needing clones
	dfltCdreProfile *CdreConfig // Default cdreConfig profile
	dfltCdrcProfile *CdrcConfig // Default cdrcConfig profile
}

func (self *CGRConfig) checkConfigSanity() error {
	// Rater checks
	if self.RALsEnabled {
		if self.RALsBalancer == utils.MetaInternal && !self.BalancerEnabled {
			return errors.New("Balancer not enabled but requested by Rater component.")
		}
		for _, connCfg := range self.RALsCDRStatSConns {
			if connCfg.Address == utils.MetaInternal && !self.CDRStatsEnabled {
				return errors.New("CDRStats not enabled but requested by Rater component.")
			}
		}
		for _, connCfg := range self.RALsHistorySConns {
			if connCfg.Address == utils.MetaInternal && !self.HistoryServerEnabled {
				return errors.New("History server not enabled but requested by Rater component.")
			}
		}
		for _, connCfg := range self.RALsPubSubSConns {
			if connCfg.Address == utils.MetaInternal && !self.PubSubServerEnabled {
				return errors.New("PubSub server not enabled but requested by Rater component.")
			}
		}
		for _, connCfg := range self.RALsAliasSConns {
			if connCfg.Address == utils.MetaInternal && !self.AliasesServerEnabled {
				return errors.New("Alias server not enabled but requested by Rater component.")
			}
		}
		for _, connCfg := range self.RALsUserSConns {
			if connCfg.Address == utils.MetaInternal && !self.UserServerEnabled {
				return errors.New("User service not enabled but requested by Rater component.")
			}
		}
	}
	// CDRServer checks
	if self.CDRSEnabled {
		for _, cdrsRaterConn := range self.CDRSRaterConns {
			if cdrsRaterConn.Address == utils.MetaInternal && !self.RALsEnabled {
				return errors.New("RALs not enabled but requested by CDRS component.")
			}
		}
		for _, connCfg := range self.CDRSPubSubSConns {
			if connCfg.Address == utils.MetaInternal && !self.PubSubServerEnabled {
				return errors.New("PubSubS not enabled but requested by CDRS component.")
			}
		}
		for _, connCfg := range self.CDRSUserSConns {
			if connCfg.Address == utils.MetaInternal && !self.UserServerEnabled {
				return errors.New("UserS not enabled but requested by CDRS component.")
			}
		}
		for _, connCfg := range self.CDRSAliaseSConns {
			if connCfg.Address == utils.MetaInternal && !self.AliasesServerEnabled {
				return errors.New("AliaseS not enabled but requested by CDRS component.")
			}
		}
		for _, connCfg := range self.CDRSStatSConns {
			if connCfg.Address == utils.MetaInternal && !self.CDRStatsEnabled {
				return errors.New("CDRStatS not enabled but requested by CDRS component.")
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
			for _, conn := range cdrcInst.CdrsConns {
				if conn.Address == utils.MetaInternal && !self.CDRSEnabled {
					return errors.New("CDRS not enabled but referenced from CDRC")
				}
			}
			if len(cdrcInst.ContentFields) == 0 {
				return errors.New("CdrC enabled but no fields to be processed defined!")
			}
			if cdrcInst.CdrFormat == utils.CSV {
				for _, cdrFld := range cdrcInst.ContentFields {
					for _, rsrFld := range cdrFld.Value {
						if _, errConv := strconv.Atoi(rsrFld.Id); errConv != nil && !rsrFld.IsStatic() {
							return fmt.Errorf("CDR fields must be indices in case of .csv files, have instead: %s", rsrFld.Id)
						}
					}
				}
			}
		}
	}
	// SMGeneric checks
	if self.SmGenericConfig.Enabled {
		if len(self.SmGenericConfig.RALsConns) == 0 {
			return errors.New("<SMGeneric> RALs definition is mandatory!")
		}
		for _, smgRALsConn := range self.SmGenericConfig.RALsConns {
			if smgRALsConn.Address == utils.MetaInternal && !self.RALsEnabled {
				return errors.New("<SMGeneric> RALs not enabled but requested by SMGeneric component.")
			}
		}
		if len(self.SmGenericConfig.CDRsConns) == 0 {
			return errors.New("<SMGeneric> CDRs definition is mandatory!")
		}
		for _, smgCDRSConn := range self.SmGenericConfig.CDRsConns {
			if smgCDRSConn.Address == utils.MetaInternal && !self.CDRSEnabled {
				return errors.New("<SMGeneric> CDRS not enabled but referenced by SMGeneric component")
			}
		}
	}
	// SMFreeSWITCH checks
	if self.SmFsConfig.Enabled {
		if len(self.SmFsConfig.RALsConns) == 0 {
			return errors.New("<SMFreeSWITCH> RALs definition is mandatory!")
		}
		for _, smFSRaterConn := range self.SmFsConfig.RALsConns {
			if smFSRaterConn.Address == utils.MetaInternal && !self.RALsEnabled {
				return errors.New("<SMFreeSWITCH> RALs not enabled but requested by SMFreeSWITCH component.")
			}
		}
		if len(self.SmFsConfig.CDRsConns) == 0 {
			return errors.New("<SMFreeSWITCH> CDRS definition is mandatory!")
		}
		for _, smFSCDRSConn := range self.SmFsConfig.CDRsConns {
			if smFSCDRSConn.Address == utils.MetaInternal && !self.CDRSEnabled {
				return errors.New("CDRS not enabled but referenced by SMFreeSWITCH component")
			}
		}
	}
	// SM-Kamailio checks
	if self.SmKamConfig.Enabled {
		if len(self.SmKamConfig.RALsConns) == 0 {
			return errors.New("Rater definition is mandatory!")
		}
		for _, smKamRaterConn := range self.SmKamConfig.RALsConns {
			if smKamRaterConn.Address == utils.MetaInternal && !self.RALsEnabled {
				return errors.New("Rater not enabled but requested by SM-Kamailio component.")
			}
		}
		if len(self.SmKamConfig.CDRsConns) == 0 {
			return errors.New("Cdrs definition is mandatory!")
		}
		for _, smKamCDRSConn := range self.SmKamConfig.CDRsConns {
			if smKamCDRSConn.Address == utils.MetaInternal && !self.CDRSEnabled {
				return errors.New("CDRS not enabled but referenced by SM-Kamailio component")
			}
		}
	}
	// SMOpenSIPS checks
	if self.SmOsipsConfig.Enabled {
		if len(self.SmOsipsConfig.RALsConns) == 0 {
			return errors.New("<SMOpenSIPS> Rater definition is mandatory!")
		}
		for _, smOsipsRaterConn := range self.SmOsipsConfig.RALsConns {
			if smOsipsRaterConn.Address == utils.MetaInternal && !self.RALsEnabled {
				return errors.New("<SMOpenSIPS> RALs not enabled but requested by SMOpenSIPS component.")
			}
		}
		if len(self.SmOsipsConfig.CDRsConns) == 0 {
			return errors.New("<SMOpenSIPS> CDRs definition is mandatory!")
		}

		for _, smOsipsCDRSConn := range self.SmOsipsConfig.CDRsConns {
			if smOsipsCDRSConn.Address == utils.MetaInternal && !self.CDRSEnabled {
				return errors.New("<SMOpenSIPS> CDRS not enabled but referenced by SMOpenSIPS component")
			}
		}
	}
	// DAgent checks
	if self.diameterAgentCfg.Enabled {
		for _, daSMGConn := range self.diameterAgentCfg.SMGenericConns {
			if daSMGConn.Address == utils.MetaInternal && !self.SmGenericConfig.Enabled {
				return errors.New("SMGeneric not enabled but referenced by DiameterAgent component")
			}
		}
		for _, daPubSubSConn := range self.diameterAgentCfg.PubSubConns {
			if daPubSubSConn.Address == utils.MetaInternal && !self.PubSubServerEnabled {
				return errors.New("PubSubS not enabled but requested by DiameterAgent component.")
			}
		}
	}
	return nil
}

// Loads from json configuration object, will be used for defaults, config from file and reload, might need lock
func (self *CGRConfig) loadFromJsonCfg(jsnCfg *CgrJsonCfg) error {

	// Load sections out of JSON config, stop on error
	jsnGeneralCfg, err := jsnCfg.GeneralJsonCfg()
	if err != nil {
		return err
	}

	jsnListenCfg, err := jsnCfg.ListenJsonCfg()
	if err != nil {
		return err
	}

	jsnTpDbCfg, err := jsnCfg.DbJsonCfg(TPDB_JSN)
	if err != nil {
		return err
	}

	jsnDataDbCfg, err := jsnCfg.DbJsonCfg(DATADB_JSN)
	if err != nil {
		return err
	}

	jsnStorDbCfg, err := jsnCfg.DbJsonCfg(STORDB_JSN)
	if err != nil {
		return err
	}

	jsnBalancerCfg, err := jsnCfg.BalancerJsonCfg()
	if err != nil {
		return err
	}

	jsnRALsCfg, err := jsnCfg.RalsJsonCfg()
	if err != nil {
		return err
	}

	jsnSchedCfg, err := jsnCfg.SchedulerJsonCfg()
	if err != nil {
		return err
	}

	jsnCdrsCfg, err := jsnCfg.CdrsJsonCfg()
	if err != nil {
		return err
	}

	jsnCdrstatsCfg, err := jsnCfg.CdrStatsJsonCfg()
	if err != nil {
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

	jsnSmGenericCfg, err := jsnCfg.SmGenericJsonCfg()
	if err != nil {
		return err
	}

	jsnSmFsCfg, err := jsnCfg.SmFsJsonCfg()
	if err != nil {
		return err
	}

	jsnSmKamCfg, err := jsnCfg.SmKamJsonCfg()
	if err != nil {
		return err
	}

	jsnSmOsipsCfg, err := jsnCfg.SmOsipsJsonCfg()
	if err != nil {
		return err
	}

	jsnDACfg, err := jsnCfg.DiameterAgentJsonCfg()
	if err != nil {
		return err
	}

	jsnHistServCfg, err := jsnCfg.HistServJsonCfg()
	if err != nil {
		return err
	}

	jsnPubSubServCfg, err := jsnCfg.PubSubServJsonCfg()
	if err != nil {
		return err
	}

	jsnAliasesServCfg, err := jsnCfg.AliasesServJsonCfg()
	if err != nil {
		return err
	}

	jsnUserServCfg, err := jsnCfg.UserServJsonCfg()
	if err != nil {
		return err
	}

	jsnRLSCfg, err := jsnCfg.ResourceLimiterJsonCfg()
	if err != nil {
		return err
	}

	jsnMailerCfg, err := jsnCfg.MailerJsonCfg()
	if err != nil {
		return err
	}

	jsnSureTaxCfg, err := jsnCfg.SureTaxJsonCfg()
	if err != nil {
		return err
	}

	// All good, start populating config variables
	if jsnTpDbCfg != nil {
		if jsnTpDbCfg.Db_type != nil {
			self.TpDbType = *jsnTpDbCfg.Db_type
		}
		if jsnTpDbCfg.Db_host != nil {
			self.TpDbHost = *jsnTpDbCfg.Db_host
		}
		if jsnTpDbCfg.Db_port != nil {
			self.TpDbPort = strconv.Itoa(*jsnTpDbCfg.Db_port)
		}
		if jsnTpDbCfg.Db_name != nil {
			self.TpDbName = *jsnTpDbCfg.Db_name
		}
		if jsnTpDbCfg.Db_user != nil {
			self.TpDbUser = *jsnTpDbCfg.Db_user
		}
		if jsnTpDbCfg.Db_password != nil {
			self.TpDbPass = *jsnTpDbCfg.Db_password
		}
	}

	if jsnDataDbCfg != nil {
		if jsnDataDbCfg.Db_type != nil {
			self.DataDbType = *jsnDataDbCfg.Db_type
		}
		if jsnDataDbCfg.Db_host != nil {
			self.DataDbHost = *jsnDataDbCfg.Db_host
		}
		if jsnDataDbCfg.Db_port != nil {
			self.DataDbPort = strconv.Itoa(*jsnDataDbCfg.Db_port)
		}
		if jsnDataDbCfg.Db_name != nil {
			self.DataDbName = *jsnDataDbCfg.Db_name
		}
		if jsnDataDbCfg.Db_user != nil {
			self.DataDbUser = *jsnDataDbCfg.Db_user
		}
		if jsnDataDbCfg.Db_password != nil {
			self.DataDbPass = *jsnDataDbCfg.Db_password
		}
		if jsnDataDbCfg.Load_history_size != nil {
			self.LoadHistorySize = *jsnDataDbCfg.Load_history_size
		}
	}

	if jsnStorDbCfg != nil {
		if jsnStorDbCfg.Db_type != nil {
			self.StorDBType = *jsnStorDbCfg.Db_type
		}
		if jsnStorDbCfg.Db_host != nil {
			self.StorDBHost = *jsnStorDbCfg.Db_host
		}
		if jsnStorDbCfg.Db_port != nil {
			self.StorDBPort = strconv.Itoa(*jsnStorDbCfg.Db_port)
		}
		if jsnStorDbCfg.Db_name != nil {
			self.StorDBName = *jsnStorDbCfg.Db_name
		}
		if jsnStorDbCfg.Db_user != nil {
			self.StorDBUser = *jsnStorDbCfg.Db_user
		}
		if jsnStorDbCfg.Db_password != nil {
			self.StorDBPass = *jsnStorDbCfg.Db_password
		}
		if jsnStorDbCfg.Max_open_conns != nil {
			self.StorDBMaxOpenConns = *jsnStorDbCfg.Max_open_conns
		}
		if jsnStorDbCfg.Max_idle_conns != nil {
			self.StorDBMaxIdleConns = *jsnStorDbCfg.Max_idle_conns
		}
		if jsnStorDbCfg.Cdrs_indexes != nil {
			self.StorDBCDRSIndexes = *jsnStorDbCfg.Cdrs_indexes
		}
	}

	if jsnGeneralCfg != nil {
		if jsnGeneralCfg.Dbdata_encoding != nil {
			self.DBDataEncoding = *jsnGeneralCfg.Dbdata_encoding
		}
		if jsnGeneralCfg.Default_request_type != nil {
			self.DefaultReqType = *jsnGeneralCfg.Default_request_type
		}
		if jsnGeneralCfg.Default_category != nil {
			self.DefaultCategory = *jsnGeneralCfg.Default_category
		}
		if jsnGeneralCfg.Default_tenant != nil {
			self.DefaultTenant = *jsnGeneralCfg.Default_tenant
		}
		if jsnGeneralCfg.Connect_attempts != nil {
			self.ConnectAttempts = *jsnGeneralCfg.Connect_attempts
		}
		if jsnGeneralCfg.Response_cache_ttl != nil {
			if self.ResponseCacheTTL, err = utils.ParseDurationWithSecs(*jsnGeneralCfg.Response_cache_ttl); err != nil {
				return err
			}
		}
		if jsnGeneralCfg.Reconnects != nil {
			self.Reconnects = *jsnGeneralCfg.Reconnects
		}
		if jsnGeneralCfg.Connect_timeout != nil {
			if self.ConnectTimeout, err = utils.ParseDurationWithSecs(*jsnGeneralCfg.Connect_timeout); err != nil {
				return err
			}
		}
		if jsnGeneralCfg.Reply_timeout != nil {
			if self.ReplyTimeout, err = utils.ParseDurationWithSecs(*jsnGeneralCfg.Reply_timeout); err != nil {
				return err
			}
		}
		if jsnGeneralCfg.Rounding_decimals != nil {
			self.RoundingDecimals = *jsnGeneralCfg.Rounding_decimals
		}
		if jsnGeneralCfg.Http_skip_tls_verify != nil {
			self.HttpSkipTlsVerify = *jsnGeneralCfg.Http_skip_tls_verify
		}
		if jsnGeneralCfg.Tpexport_dir != nil {
			self.TpExportPath = *jsnGeneralCfg.Tpexport_dir
		}
		if jsnGeneralCfg.Httpposter_attempts != nil {
			self.HttpPosterAttempts = *jsnGeneralCfg.Httpposter_attempts
		}
		if jsnGeneralCfg.Http_failed_dir != nil {
			self.HttpFailedDir = *jsnGeneralCfg.Http_failed_dir
		}
		if jsnGeneralCfg.Default_timezone != nil {
			self.DefaultTimezone = *jsnGeneralCfg.Default_timezone
		}
		if jsnGeneralCfg.Internal_ttl != nil {
			if self.InternalTtl, err = utils.ParseDurationWithSecs(*jsnGeneralCfg.Internal_ttl); err != nil {
				return err
			}
		}
		if jsnGeneralCfg.Locking_timeout != nil {
			if self.LockingTimeout, err = utils.ParseDurationWithSecs(*jsnGeneralCfg.Locking_timeout); err != nil {
				return err
			}
		}
		if jsnGeneralCfg.Cache_dump_dir != nil {
			self.CacheDumpDir = *jsnGeneralCfg.Cache_dump_dir
		}
	}

	if jsnListenCfg != nil {
		if jsnListenCfg.Rpc_json != nil {
			self.RPCJSONListen = *jsnListenCfg.Rpc_json
		}
		if jsnListenCfg.Rpc_gob != nil {
			self.RPCGOBListen = *jsnListenCfg.Rpc_gob
		}
		if jsnListenCfg.Http != nil {
			self.HTTPListen = *jsnListenCfg.Http
		}
	}

	if jsnRALsCfg != nil {
		if jsnRALsCfg.Enabled != nil {
			self.RALsEnabled = *jsnRALsCfg.Enabled
		}
		if jsnRALsCfg.Balancer != nil {
			self.RALsBalancer = *jsnRALsCfg.Balancer
		}
		if jsnRALsCfg.Cdrstats_conns != nil {
			self.RALsCDRStatSConns = make([]*HaPoolConfig, len(*jsnRALsCfg.Cdrstats_conns))
			for idx, jsnHaCfg := range *jsnRALsCfg.Cdrstats_conns {
				self.RALsCDRStatSConns[idx] = NewDfltHaPoolConfig()
				self.RALsCDRStatSConns[idx].loadFromJsonCfg(jsnHaCfg)
			}
		}
		if jsnRALsCfg.Historys_conns != nil {
			self.RALsHistorySConns = make([]*HaPoolConfig, len(*jsnRALsCfg.Historys_conns))
			for idx, jsnHaCfg := range *jsnRALsCfg.Historys_conns {
				self.RALsHistorySConns[idx] = NewDfltHaPoolConfig()
				self.RALsHistorySConns[idx].loadFromJsonCfg(jsnHaCfg)
			}
		}
		if jsnRALsCfg.Pubsubs_conns != nil {
			self.RALsPubSubSConns = make([]*HaPoolConfig, len(*jsnRALsCfg.Pubsubs_conns))
			for idx, jsnHaCfg := range *jsnRALsCfg.Pubsubs_conns {
				self.RALsPubSubSConns[idx] = NewDfltHaPoolConfig()
				self.RALsPubSubSConns[idx].loadFromJsonCfg(jsnHaCfg)
			}
		}
		if jsnRALsCfg.Aliases_conns != nil {
			self.RALsAliasSConns = make([]*HaPoolConfig, len(*jsnRALsCfg.Aliases_conns))
			for idx, jsnHaCfg := range *jsnRALsCfg.Aliases_conns {
				self.RALsAliasSConns[idx] = NewDfltHaPoolConfig()
				self.RALsAliasSConns[idx].loadFromJsonCfg(jsnHaCfg)
			}
		}
		if jsnRALsCfg.Users_conns != nil {
			self.RALsUserSConns = make([]*HaPoolConfig, len(*jsnRALsCfg.Users_conns))
			for idx, jsnHaCfg := range *jsnRALsCfg.Users_conns {
				self.RALsUserSConns[idx] = NewDfltHaPoolConfig()
				self.RALsUserSConns[idx].loadFromJsonCfg(jsnHaCfg)
			}
		}
		if jsnRALsCfg.Rp_subject_prefix_matching != nil {
			self.RpSubjectPrefixMatching = *jsnRALsCfg.Rp_subject_prefix_matching
		}
		if jsnRALsCfg.Lcr_subject_prefix_matching != nil {
			self.LcrSubjectPrefixMatching = *jsnRALsCfg.Lcr_subject_prefix_matching
		}
	}

	if jsnBalancerCfg != nil && jsnBalancerCfg.Enabled != nil {
		self.BalancerEnabled = *jsnBalancerCfg.Enabled
	}

	if jsnSchedCfg != nil && jsnSchedCfg.Enabled != nil {
		self.SchedulerEnabled = *jsnSchedCfg.Enabled
	}

	if jsnCdrsCfg != nil {
		if jsnCdrsCfg.Enabled != nil {
			self.CDRSEnabled = *jsnCdrsCfg.Enabled
		}
		if jsnCdrsCfg.Extra_fields != nil {
			if self.CDRSExtraFields, err = utils.ParseRSRFieldsFromSlice(*jsnCdrsCfg.Extra_fields); err != nil {
				return err
			}
		}
		if jsnCdrsCfg.Store_cdrs != nil {
			self.CDRSStoreCdrs = *jsnCdrsCfg.Store_cdrs
		}
		if jsnCdrsCfg.Rals_conns != nil {
			self.CDRSRaterConns = make([]*HaPoolConfig, len(*jsnCdrsCfg.Rals_conns))
			for idx, jsnHaCfg := range *jsnCdrsCfg.Rals_conns {
				self.CDRSRaterConns[idx] = NewDfltHaPoolConfig()
				self.CDRSRaterConns[idx].loadFromJsonCfg(jsnHaCfg)
			}
		}
		if jsnCdrsCfg.Pubsubs_conns != nil {
			self.CDRSPubSubSConns = make([]*HaPoolConfig, len(*jsnCdrsCfg.Pubsubs_conns))
			for idx, jsnHaCfg := range *jsnCdrsCfg.Pubsubs_conns {
				self.CDRSPubSubSConns[idx] = NewDfltHaPoolConfig()
				self.CDRSPubSubSConns[idx].loadFromJsonCfg(jsnHaCfg)
			}
		}
		if jsnCdrsCfg.Users_conns != nil {
			self.CDRSUserSConns = make([]*HaPoolConfig, len(*jsnCdrsCfg.Users_conns))
			for idx, jsnHaCfg := range *jsnCdrsCfg.Users_conns {
				self.CDRSUserSConns[idx] = NewDfltHaPoolConfig()
				self.CDRSUserSConns[idx].loadFromJsonCfg(jsnHaCfg)
			}
		}
		if jsnCdrsCfg.Aliases_conns != nil {
			self.CDRSAliaseSConns = make([]*HaPoolConfig, len(*jsnCdrsCfg.Aliases_conns))
			for idx, jsnHaCfg := range *jsnCdrsCfg.Aliases_conns {
				self.CDRSAliaseSConns[idx] = NewDfltHaPoolConfig()
				self.CDRSAliaseSConns[idx].loadFromJsonCfg(jsnHaCfg)
			}
		}
		if jsnCdrsCfg.Cdrstats_conns != nil {
			self.CDRSStatSConns = make([]*HaPoolConfig, len(*jsnCdrsCfg.Cdrstats_conns))
			for idx, jsnHaCfg := range *jsnCdrsCfg.Cdrstats_conns {
				self.CDRSStatSConns[idx] = NewDfltHaPoolConfig()
				self.CDRSStatSConns[idx].loadFromJsonCfg(jsnHaCfg)
			}
		}
		if jsnCdrsCfg.Cdr_replication != nil {
			self.CDRSCdrReplication = make([]*CdrReplicationCfg, len(*jsnCdrsCfg.Cdr_replication))
			for idx, rplJsonCfg := range *jsnCdrsCfg.Cdr_replication {
				self.CDRSCdrReplication[idx] = new(CdrReplicationCfg)
				if rplJsonCfg.Transport != nil {
					self.CDRSCdrReplication[idx].Transport = *rplJsonCfg.Transport
				}
				if rplJsonCfg.Address != nil {
					self.CDRSCdrReplication[idx].Address = *rplJsonCfg.Address
				}
				if rplJsonCfg.Synchronous != nil {
					self.CDRSCdrReplication[idx].Synchronous = *rplJsonCfg.Synchronous
				}
				self.CDRSCdrReplication[idx].Attempts = 1
				if rplJsonCfg.Attempts != nil {
					self.CDRSCdrReplication[idx].Attempts = *rplJsonCfg.Attempts
				}
				if rplJsonCfg.Cdr_filter != nil {
					if self.CDRSCdrReplication[idx].CdrFilter, err = utils.ParseRSRFields(*rplJsonCfg.Cdr_filter, utils.INFIELD_SEP); err != nil {
						return err
					}
				}
			}
		}
	}

	if jsnCdrstatsCfg != nil {
		if jsnCdrstatsCfg.Enabled != nil {
			self.CDRStatsEnabled = *jsnCdrstatsCfg.Enabled
			if jsnCdrstatsCfg.Save_Interval != nil {
				if self.CDRStatsSaveInterval, err = utils.ParseDurationWithSecs(*jsnCdrstatsCfg.Save_Interval); err != nil {
					return err
				}
			}
		}
	}

	if jsnCdreCfg != nil {
		if self.CdreProfiles == nil {
			self.CdreProfiles = make(map[string]*CdreConfig)
		}
		for profileName, jsnCdre1Cfg := range jsnCdreCfg {
			if _, hasProfile := self.CdreProfiles[profileName]; !hasProfile { // New profile, create before loading from json
				self.CdreProfiles[profileName] = new(CdreConfig)
				if profileName != utils.META_DEFAULT {
					self.CdreProfiles[profileName] = self.dfltCdreProfile.Clone() // Clone default so we do not inherit pointers
				}
			}
			if err = self.CdreProfiles[profileName].loadFromJsonCfg(jsnCdre1Cfg); err != nil { // Update the existing profile with content from json config
				return err
			}
		}
	}
	if jsnCdrcCfg != nil {
		if self.CdrcProfiles == nil {
			self.CdrcProfiles = make(map[string][]*CdrcConfig)
		}
		for _, jsnCrc1Cfg := range jsnCdrcCfg {
			if _, hasDir := self.CdrcProfiles[*jsnCrc1Cfg.Cdr_in_dir]; !hasDir {
				self.CdrcProfiles[*jsnCrc1Cfg.Cdr_in_dir] = make([]*CdrcConfig, 0)
			}
			var cdrcInstCfg *CdrcConfig
			if *jsnCrc1Cfg.Id == utils.META_DEFAULT && self.dfltCdrcProfile == nil {
				cdrcInstCfg = new(CdrcConfig)
			} else {
				cdrcInstCfg = self.dfltCdrcProfile.Clone() // Clone default so we do not inherit pointers
			}
			if err := cdrcInstCfg.loadFromJsonCfg(jsnCrc1Cfg); err != nil {
				return err
			}
			self.CdrcProfiles[*jsnCrc1Cfg.Cdr_in_dir] = append(self.CdrcProfiles[*jsnCrc1Cfg.Cdr_in_dir], cdrcInstCfg)
		}
	}
	if jsnSmGenericCfg != nil {
		if err := self.SmGenericConfig.loadFromJsonCfg(jsnSmGenericCfg); err != nil {
			return err
		}
	}
	if jsnSmFsCfg != nil {
		if err := self.SmFsConfig.loadFromJsonCfg(jsnSmFsCfg); err != nil {
			return err
		}
	}

	if jsnSmKamCfg != nil {
		if err := self.SmKamConfig.loadFromJsonCfg(jsnSmKamCfg); err != nil {
			return err
		}
	}

	if jsnSmOsipsCfg != nil {
		if err := self.SmOsipsConfig.loadFromJsonCfg(jsnSmOsipsCfg); err != nil {
			return err
		}
	}

	if jsnDACfg != nil {
		if err := self.diameterAgentCfg.loadFromJsonCfg(jsnDACfg); err != nil {
			return err
		}
	}

	if jsnHistServCfg != nil {
		if jsnHistServCfg.Enabled != nil {
			self.HistoryServerEnabled = *jsnHistServCfg.Enabled
		}
		if jsnHistServCfg.History_dir != nil {
			self.HistoryDir = *jsnHistServCfg.History_dir
		}
		if jsnHistServCfg.Save_interval != nil {
			if self.HistorySaveInterval, err = utils.ParseDurationWithSecs(*jsnHistServCfg.Save_interval); err != nil {
				return err
			}
		}
	}

	if jsnPubSubServCfg != nil {
		if jsnPubSubServCfg.Enabled != nil {
			self.PubSubServerEnabled = *jsnPubSubServCfg.Enabled
		}
	}

	if jsnAliasesServCfg != nil {
		if jsnAliasesServCfg.Enabled != nil {
			self.AliasesServerEnabled = *jsnAliasesServCfg.Enabled
		}
	}

	if jsnRLSCfg != nil {
		if self.ResourceLimiterCfg.loadFromJsonCfg(jsnRLSCfg); err != nil {
			return err
		}
	}

	if jsnUserServCfg != nil {
		if jsnUserServCfg.Enabled != nil {
			self.UserServerEnabled = *jsnUserServCfg.Enabled
		}
		if jsnUserServCfg.Indexes != nil {
			self.UserServerIndexes = *jsnUserServCfg.Indexes
		}
	}

	if jsnMailerCfg != nil {
		if jsnMailerCfg.Server != nil {
			self.MailerServer = *jsnMailerCfg.Server
		}
		if jsnMailerCfg.Auth_user != nil {
			self.MailerAuthUser = *jsnMailerCfg.Auth_user
		}
		if jsnMailerCfg.Auth_password != nil {
			self.MailerAuthPass = *jsnMailerCfg.Auth_password
		}
		if jsnMailerCfg.From_address != nil {
			self.MailerFromAddr = *jsnMailerCfg.From_address
		}
	}

	if jsnSureTaxCfg != nil { // New config for SureTax
		if self.sureTaxCfg, err = NewSureTaxCfgWithDefaults(); err != nil {
			return err
		}
		if err := self.sureTaxCfg.loadFromJsonCfg(jsnSureTaxCfg); err != nil {
			return err
		}
	}

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
