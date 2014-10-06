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
	"strconv"
	"strings"
	"time"

	"code.google.com/p/goconf/conf"
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

var cgrCfg *CGRConfig // will be shared

// Used to retrieve system configuration from other packages
func CgrConfig() *CGRConfig {
	return cgrCfg
}

// Used to set system configuration from other places
func SetCgrConfig(cfg *CGRConfig) {
	cgrCfg = cfg
}

// Holds system configuration, defaults are overwritten with values from config file if found
type CGRConfig struct {
	RatingDBType            string
	RatingDBHost            string // The host to connect to. Values that start with / are for UNIX domain sockets.
	RatingDBPort            string // The port to bind to.
	RatingDBName            string // The name of the database to connect to.
	RatingDBUser            string // The user to sign in as.
	RatingDBPass            string // The user's password.
	AccountDBType           string
	AccountDBHost           string             // The host to connect to. Values that start with / are for UNIX domain sockets.
	AccountDBPort           string             // The port to bind to.
	AccountDBName           string             // The name of the database to connect to.
	AccountDBUser           string             // The user to sign in as.
	AccountDBPass           string             // The user's password.
	StorDBType              string             // Should reflect the database type used to store logs
	StorDBHost              string             // The host to connect to. Values that start with / are for UNIX domain sockets.
	StorDBPort              string             // Th e port to bind to.
	StorDBName              string             // The name of the database to connect to.
	StorDBUser              string             // The user to sign in as.
	StorDBPass              string             // The user's password.
	StorDBMaxOpenConns      int                // Maximum database connections opened
	StorDBMaxIdleConns      int                // Maximum idle connections to keep opened
	DBDataEncoding          string             // The encoding used to store object data in strings: <msgpack|json>
	RPCJSONListen           string             // RPC JSON listening address
	RPCGOBListen            string             // RPC GOB listening address
	HTTPListen              string             // HTTP listening address
	DefaultReqType          string             // Use this request type if not defined on top
	DefaultCategory         string             // set default type of record
	DefaultTenant           string             // set default tenant
	DefaultSubject          string             // set default rating subject, useful in case of fallback
	RoundingDecimals        int                // Number of decimals to round end prices at
	HttpSkipTlsVerify       bool               // If enabled Http Client will accept any TLS certificate
	XmlCfgDocument          *CgrXmlCfgDocument // Load additional configuration inside xml document
	RaterEnabled            bool               // start standalone server (no balancer)
	RaterBalancer           string             // balancer address host:port
	BalancerEnabled         bool
	SchedulerEnabled        bool
	CDRSEnabled             bool              // Enable CDR Server service
	CDRSExtraFields         []*utils.RSRField // Extra fields to store in CDRs
	CDRSMediator            string            // Address where to reach the Mediator. Empty for disabling mediation. <""|internal>
	CDRSStats               string            // Address where to reach the Mediator. <""|intenal>
	CDRSStoreDisable        bool              // When true, CDRs will not longer be saved in stordb, useful for cdrstats only scenario
	CDRStatsEnabled         bool              // Enable CDR Stats service
	CDRStatConfig           *CdrStatsConfig   // Active cdr stats configuration instances
	CdreDefaultInstance     *CdreConfig       // Will be used in the case no specific one selected by API
	CdrcInstances           []*CdrcConfig     // Number of CDRC instances running imports
	SMEnabled               bool
	SMSwitchType            string
	SMRater                 string        // address where to access rater. Can be internal, direct rater address or the address of a balancer
	SMCdrS                  string        //
	SMReconnects            int           // Number of reconnect attempts to rater
	SMDebitInterval         int           // the period to be debited in advanced during a call (in seconds)
	SMMaxCallDuration       time.Duration // The maximum duration of a call
	SMMinCallDuration       time.Duration // Only authorize calls with allowed duration bigger than this
	MediatorEnabled         bool          // Starts Mediator service: <true|false>.
	MediatorReconnects      int           // Number of reconnects to rater before giving up.
	MediatorRater           string
	MediatorStats           string                // Address where to reach the Rater: <internal|x.y.z.y:1234>
	MediatorStoreDisable    bool                  // When true, CDRs will not longer be saved in stordb, useful for cdrstats only scenario
	DerivedChargers         utils.DerivedChargers // System wide derived chargers, added to the account level ones
	CombinedDerivedChargers bool                  // Combine accounts specific derived_chargers with server configured
	FreeswitchServer        string                // freeswitch address host:port
	FreeswitchPass          string                // FS socket password
	FreeswitchReconnects    int                   // number of times to attempt reconnect after connect fails
	FSMinDurLowBalance      time.Duration         // Threshold which will trigger low balance warnings
	FSLowBalanceAnnFile     string                // File to be played when low balance is reached
	FSEmptyBalanceContext   string                // If defined, call will be transfered to this context on empty balance
	FSEmptyBalanceAnnFile   string                // File to be played before disconnecting prepaid calls (applies only if no context defined)
	FSCdrExtraFields        []*utils.RSRField     // Extra fields to store in CDRs in case of processing them
	OsipsListenUdp          string                // Address where to listen for event datagrams coming from OpenSIPS
	OsipsMiAddr             string                // Adress where to reach OpenSIPS mi_datagram module
	OsipsEvSubscInterval    time.Duration         // Refresh event subscription at this interval
	OsipsReconnects         int                   // Number of attempts on connect failure.
	HistoryAgentEnabled     bool                  // Starts History as an agent: <true|false>.
	HistoryServer           string                // Address where to reach the master history server: <internal|x.y.z.y:1234>
	HistoryServerEnabled    bool                  // Starts History as server: <true|false>.
	HistoryDir              string                // Location on disk where to store history files.
	HistorySaveInterval     time.Duration         // The timout duration between history writes
	MailerServer            string                // The server to use when sending emails out
	MailerAuthUser          string                // Authenticate to email server using this user
	MailerAuthPass          string                // Authenticate to email server with this password
	MailerFromAddr          string                // From address used when sending emails out
	DataFolderPath          string                // Path towards data folder, for tests internal usage, not loading out of .cfg options
}

func (self *CGRConfig) setDefaults() error {
	self.RatingDBType = REDIS
	self.RatingDBHost = "127.0.0.1"
	self.RatingDBPort = "6379"
	self.RatingDBName = "10"
	self.RatingDBUser = ""
	self.RatingDBPass = ""
	self.AccountDBType = REDIS
	self.AccountDBHost = "127.0.0.1"
	self.AccountDBPort = "6379"
	self.AccountDBName = "11"
	self.AccountDBUser = ""
	self.AccountDBPass = ""
	self.StorDBType = utils.MYSQL
	self.StorDBHost = "localhost"
	self.StorDBPort = "3306"
	self.StorDBName = "cgrates"
	self.StorDBUser = "cgrates"
	self.StorDBPass = "CGRateS.org"
	self.StorDBMaxOpenConns = 100
	self.StorDBMaxIdleConns = 10
	self.DBDataEncoding = utils.MSGPACK
	self.RPCJSONListen = "127.0.0.1:2012"
	self.RPCGOBListen = "127.0.0.1:2013"
	self.HTTPListen = "127.0.0.1:2080"
	self.DefaultReqType = utils.RATED
	self.DefaultCategory = "call"
	self.DefaultTenant = "cgrates.org"
	self.DefaultSubject = "cgrates"
	self.RoundingDecimals = 10
	self.HttpSkipTlsVerify = false
	self.XmlCfgDocument = nil
	self.RaterEnabled = false
	self.RaterBalancer = ""
	self.BalancerEnabled = false
	self.SchedulerEnabled = false
	self.CDRSEnabled = false
	self.CDRSExtraFields = []*utils.RSRField{}
	self.CDRSMediator = ""
	self.CDRSStats = ""
	self.CDRSStoreDisable = false
	self.CDRStatsEnabled = false
	self.CDRStatConfig = NewCdrStatsConfigWithDefaults()
	self.CdreDefaultInstance = NewDefaultCdreConfig()
	self.CdrcInstances = []*CdrcConfig{NewDefaultCdrcConfig()} // This instance is just for the sake of defaults, it will be replaced when the file is loaded with the one resulted from there
	self.MediatorEnabled = false
	self.MediatorRater = utils.INTERNAL
	self.MediatorReconnects = 3
	self.MediatorStats = ""
	self.MediatorStoreDisable = false
	self.DerivedChargers = make(utils.DerivedChargers, 0)
	self.CombinedDerivedChargers = true
	self.SMEnabled = false
	self.SMSwitchType = FS
	self.SMRater = utils.INTERNAL
	self.SMCdrS = ""
	self.SMReconnects = 3
	self.SMDebitInterval = 10
	self.SMMaxCallDuration = time.Duration(3) * time.Hour
	self.SMMinCallDuration = time.Duration(0)
	self.FreeswitchServer = "127.0.0.1:8021"
	self.FreeswitchPass = "ClueCon"
	self.FreeswitchReconnects = 5
	self.FSMinDurLowBalance = time.Duration(5) * time.Second
	self.FSLowBalanceAnnFile = ""
	self.FSEmptyBalanceContext = ""
	self.FSEmptyBalanceAnnFile = ""
	self.FSCdrExtraFields = []*utils.RSRField{}
	self.OsipsListenUdp = "127.0.0.1:2020"
	self.OsipsMiAddr = "127.0.0.1:8020"
	self.OsipsEvSubscInterval = time.Duration(60) * time.Second
	self.OsipsReconnects = 3
	self.HistoryAgentEnabled = false
	self.HistoryServerEnabled = false
	self.HistoryServer = utils.INTERNAL
	self.HistoryDir = "/var/log/cgrates/history"
	self.HistorySaveInterval = time.Duration(1) * time.Second
	self.MailerServer = "localhost:25"
	self.MailerAuthUser = "cgrates"
	self.MailerAuthPass = "CGRateS.org"
	self.MailerFromAddr = "cgr-mailer@localhost.localdomain"
	self.DataFolderPath = "/usr/share/cgrates/"
	return nil
}

func (self *CGRConfig) checkConfigSanity() error {
	// CDRC sanity checks
	for _, cdrcInst := range self.CdrcInstances {
		if cdrcInst.Enabled == true {
			if len(cdrcInst.CdrFields) == 0 {
				return errors.New("CdrC enabled but no fields to be processed defined!")
			}
			if cdrcInst.CdrFormat == utils.CSV {
				for _, cdrFld := range cdrcInst.CdrFields {
					for _, rsrFld := range cdrFld.Value {
						if _, errConv := strconv.Atoi(rsrFld.Id); errConv != nil && !rsrFld.IsStatic() {
							return fmt.Errorf("CDR fields must be indices in case of .csv files, have instead: %s", rsrFld.Id)
						}
					}
				}
			}
		}
	}
	if self.CDRSStats == utils.INTERNAL && !self.CDRStatsEnabled {
		return errors.New("CDRStats not enabled but requested by CDRS component.")
	}
	if self.MediatorStats == utils.INTERNAL && !self.CDRStatsEnabled {
		return errors.New("CDRStats not enabled but requested by Mediator.")
	}
	if self.SMCdrS == utils.INTERNAL && !self.CDRSEnabled {
		return errors.New("CDRS not enabled but requested by SessionManager")
	}
	return nil
}

func NewDefaultCGRConfig() (*CGRConfig, error) {
	cfg := &CGRConfig{}
	cfg.setDefaults()
	if err := cfg.checkConfigSanity(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// Unifies the config handling for both tests and real path
func NewCGRConfig(c *conf.ConfigFile) (*CGRConfig, error) {
	cfg, err := loadConfig(c)
	if err != nil {
		return nil, err
	}
	if err := cfg.checkConfigSanity(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// Instantiate a new CGRConfig setting defaults or reading from file
func NewCGRConfigFromFile(cfgPath *string) (*CGRConfig, error) {
	c, err := conf.ReadConfigFile(*cfgPath)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Could not open the configuration file: %s", err))
	}
	return NewCGRConfig(c)
}

func NewCGRConfigFromBytes(data []byte) (*CGRConfig, error) {
	c, err := conf.ReadConfigBytes(data)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Could not open the configuration file: %s", err))
	}
	return NewCGRConfig(c)
}

func loadConfig(c *conf.ConfigFile) (*CGRConfig, error) {
	cfg := &CGRConfig{}
	cfg.setDefaults()
	var hasOpt bool
	var err error
	if hasOpt = c.HasOption("global", "ratingdb_type"); hasOpt {
		cfg.RatingDBType, _ = c.GetString("global", "ratingdb_type")
	}
	if hasOpt = c.HasOption("global", "ratingdb_host"); hasOpt {
		cfg.RatingDBHost, _ = c.GetString("global", "ratingdb_host")
	}
	if hasOpt = c.HasOption("global", "ratingdb_port"); hasOpt {
		cfg.RatingDBPort, _ = c.GetString("global", "ratingdb_port")
	}
	if hasOpt = c.HasOption("global", "ratingdb_name"); hasOpt {
		cfg.RatingDBName, _ = c.GetString("global", "ratingdb_name")
	}
	if hasOpt = c.HasOption("global", "ratingdb_user"); hasOpt {
		cfg.RatingDBUser, _ = c.GetString("global", "ratingdb_user")
	}
	if hasOpt = c.HasOption("global", "ratingdb_passwd"); hasOpt {
		cfg.RatingDBPass, _ = c.GetString("global", "ratingdb_passwd")
	}
	if hasOpt = c.HasOption("global", "accountdb_type"); hasOpt {
		cfg.AccountDBType, _ = c.GetString("global", "accountdb_type")
	}
	if hasOpt = c.HasOption("global", "accountdb_host"); hasOpt {
		cfg.AccountDBHost, _ = c.GetString("global", "accountdb_host")
	}
	if hasOpt = c.HasOption("global", "accountdb_port"); hasOpt {
		cfg.AccountDBPort, _ = c.GetString("global", "accountdb_port")
	}
	if hasOpt = c.HasOption("global", "accountdb_name"); hasOpt {
		cfg.AccountDBName, _ = c.GetString("global", "accountdb_name")
	}
	if hasOpt = c.HasOption("global", "accountdb_user"); hasOpt {
		cfg.AccountDBUser, _ = c.GetString("global", "accountdb_user")
	}
	if hasOpt = c.HasOption("global", "accountdb_passwd"); hasOpt {
		cfg.AccountDBPass, _ = c.GetString("global", "accountdb_passwd")
	}
	if hasOpt = c.HasOption("global", "stordb_type"); hasOpt {
		cfg.StorDBType, _ = c.GetString("global", "stordb_type")
	}
	if hasOpt = c.HasOption("global", "stordb_host"); hasOpt {
		cfg.StorDBHost, _ = c.GetString("global", "stordb_host")
	}
	if hasOpt = c.HasOption("global", "stordb_port"); hasOpt {
		cfg.StorDBPort, _ = c.GetString("global", "stordb_port")
	}
	if hasOpt = c.HasOption("global", "stordb_name"); hasOpt {
		cfg.StorDBName, _ = c.GetString("global", "stordb_name")
	}
	if hasOpt = c.HasOption("global", "stordb_user"); hasOpt {
		cfg.StorDBUser, _ = c.GetString("global", "stordb_user")
	}
	if hasOpt = c.HasOption("global", "stordb_passwd"); hasOpt {
		cfg.StorDBPass, _ = c.GetString("global", "stordb_passwd")
	}
	if hasOpt = c.HasOption("global", "stordb_max_open_conns"); hasOpt {
		cfg.StorDBMaxOpenConns, _ = c.GetInt("global", "stordb_max_open_conns")
	}
	if hasOpt = c.HasOption("global", "stordb_max_idle_conns"); hasOpt {
		cfg.StorDBMaxIdleConns, _ = c.GetInt("global", "stordb_max_idle_conns")
	}
	if hasOpt = c.HasOption("global", "dbdata_encoding"); hasOpt {
		cfg.DBDataEncoding, _ = c.GetString("global", "dbdata_encoding")
	}
	if hasOpt = c.HasOption("global", "rpc_json_listen"); hasOpt {
		cfg.RPCJSONListen, _ = c.GetString("global", "rpc_json_listen")
	}
	if hasOpt = c.HasOption("global", "rpc_gob_listen"); hasOpt {
		cfg.RPCGOBListen, _ = c.GetString("global", "rpc_gob_listen")
	}
	if hasOpt = c.HasOption("global", "http_listen"); hasOpt {
		cfg.HTTPListen, _ = c.GetString("global", "http_listen")
	}
	if hasOpt = c.HasOption("global", "default_reqtype"); hasOpt {
		cfg.DefaultReqType, _ = c.GetString("global", "default_reqtype")
	}
	if hasOpt = c.HasOption("global", "default_category"); hasOpt {
		cfg.DefaultCategory, _ = c.GetString("global", "default_category")
	}
	if hasOpt = c.HasOption("global", "default_tenant"); hasOpt {
		cfg.DefaultTenant, _ = c.GetString("global", "default_tenant")
	}
	if hasOpt = c.HasOption("global", "default_subject"); hasOpt {
		cfg.DefaultSubject, _ = c.GetString("global", "default_subject")
	}
	if hasOpt = c.HasOption("global", "rounding_decimals"); hasOpt {
		cfg.RoundingDecimals, _ = c.GetInt("global", "rounding_decimals")
	}
	if hasOpt = c.HasOption("global", "http_skip_tls_veify"); hasOpt {
		cfg.HttpSkipTlsVerify, _ = c.GetBool("global", "http_skip_tls_veify")
	}
	// XML config path defined, try loading the document
	if hasOpt = c.HasOption("global", "xmlcfg_path"); hasOpt {
		xmlCfgPath, _ := c.GetString("global", "xmlcfg_path")
		xmlFile, err := os.Open(xmlCfgPath)
		if err != nil {
			return nil, err
		}
		if cgrXmlCfgDoc, err := ParseCgrXmlConfig(xmlFile); err != nil {
			return nil, err
		} else {
			cfg.XmlCfgDocument = cgrXmlCfgDoc
		}
	}
	if hasOpt = c.HasOption("rater", "enabled"); hasOpt {
		cfg.RaterEnabled, _ = c.GetBool("rater", "enabled")
	}
	if hasOpt = c.HasOption("rater", "balancer"); hasOpt {
		cfg.RaterBalancer, _ = c.GetString("rater", "balancer")
	}
	if hasOpt = c.HasOption("balancer", "enabled"); hasOpt {
		cfg.BalancerEnabled, _ = c.GetBool("balancer", "enabled")
	}
	if hasOpt = c.HasOption("scheduler", "enabled"); hasOpt {
		cfg.SchedulerEnabled, _ = c.GetBool("scheduler", "enabled")
	}
	if hasOpt = c.HasOption("cdrs", "enabled"); hasOpt {
		cfg.CDRSEnabled, _ = c.GetBool("cdrs", "enabled")
	}
	if hasOpt = c.HasOption("cdrs", "extra_fields"); hasOpt {
		extraFieldsStr, _ := c.GetString("cdrs", "extra_fields")
		if extraFields, err := utils.ParseRSRFields(extraFieldsStr, utils.FIELDS_SEP); err != nil {
			return nil, err
		} else {
			cfg.CDRSExtraFields = extraFields
		}
	}
	if hasOpt = c.HasOption("cdrs", "mediator"); hasOpt {
		cfg.CDRSMediator, _ = c.GetString("cdrs", "mediator")
	}
	if hasOpt = c.HasOption("cdrs", "cdrstats"); hasOpt {
		cfg.CDRSStats, _ = c.GetString("cdrs", "cdrstats")
	}
	if hasOpt = c.HasOption("cdrs", "store_disable"); hasOpt {
		cfg.CDRSStoreDisable, _ = c.GetBool("cdrs", "store_disable")
	}
	if hasOpt = c.HasOption("cdrstats", "enabled"); hasOpt {
		cfg.CDRStatsEnabled, _ = c.GetBool("cdrstats", "enabled")
	}
	if cfg.CDRStatConfig, err = ParseCfgDefaultCDRStatsConfig(c); err != nil {
		return nil, err
	}
	if hasOpt = c.HasOption("cdre", "cdr_format"); hasOpt {
		cfg.CdreDefaultInstance.CdrFormat, _ = c.GetString("cdre", "cdr_format")
	}
	if hasOpt = c.HasOption("cdre", "mask_destination_id"); hasOpt {
		cfg.CdreDefaultInstance.MaskDestId, _ = c.GetString("cdre", "mask_destination_id")
	}
	if hasOpt = c.HasOption("cdre", "mask_length"); hasOpt {
		cfg.CdreDefaultInstance.MaskLength, _ = c.GetInt("cdre", "mask_length")
	}
	if hasOpt = c.HasOption("cdre", "data_usage_multiply_factor"); hasOpt {
		cfg.CdreDefaultInstance.DataUsageMultiplyFactor, _ = c.GetFloat64("cdre", "data_usage_multiply_factor")
	}
	if hasOpt = c.HasOption("cdre", "cost_multiply_factor"); hasOpt {
		cfg.CdreDefaultInstance.CostMultiplyFactor, _ = c.GetFloat64("cdre", "cost_multiply_factor")
	}
	if hasOpt = c.HasOption("cdre", "cost_rounding_decimals"); hasOpt {
		cfg.CdreDefaultInstance.CostRoundingDecimals, _ = c.GetInt("cdre", "cost_rounding_decimals")
	}
	if hasOpt = c.HasOption("cdre", "cost_shift_digits"); hasOpt {
		cfg.CdreDefaultInstance.CostShiftDigits, _ = c.GetInt("cdre", "cost_shift_digits")
	}
	if hasOpt = c.HasOption("cdre", "export_template"); hasOpt { // Load configs for csv normally from template, fixed_width from xml file
		exportTemplate, _ := c.GetString("cdre", "export_template")
		if strings.HasPrefix(exportTemplate, utils.XML_PROFILE_PREFIX) {
			if xmlTemplates := cfg.XmlCfgDocument.GetCdreCfgs(exportTemplate[len(utils.XML_PROFILE_PREFIX):]); xmlTemplates != nil {
				if cfg.CdreDefaultInstance, err = NewCdreConfigFromXmlCdreCfg(xmlTemplates[exportTemplate[len(utils.XML_PROFILE_PREFIX):]]); err != nil {
					return nil, err
				}
			}
		} else { // Not loading out of template
			if flds, err := NewCfgCdrFieldsFromIds(cfg.CdreDefaultInstance.CdrFormat == utils.CDRE_FIXED_WIDTH,
				strings.Split(exportTemplate, string(utils.CSV_SEP))...); err != nil {
				return nil, err
			} else {
				cfg.CdreDefaultInstance.ContentFields = flds
			}
		}
	}
	if hasOpt = c.HasOption("cdre", "export_dir"); hasOpt {
		cfg.CdreDefaultInstance.ExportDir, _ = c.GetString("cdre", "export_dir")
	}
	// CDRC Default instance parsing
	if cdrcFileCfgInst, err := NewCdrcConfigFromFileParams(c); err != nil {
		return nil, err
	} else {
		cfg.CdrcInstances = []*CdrcConfig{cdrcFileCfgInst}
	}
	if cfg.XmlCfgDocument != nil { // Add the possible configured instances inside xml doc
		for id, xmlCdrcCfg := range cfg.XmlCfgDocument.GetCdrcCfgs("") {
			if cdrcInst, err := NewCdrcConfigFromCgrXmlCdrcCfg(id, xmlCdrcCfg); err != nil {
				return nil, err
			} else {
				cfg.CdrcInstances = append(cfg.CdrcInstances, cdrcInst)
			}
		}
	}
	if hasOpt = c.HasOption("mediator", "enabled"); hasOpt {
		cfg.MediatorEnabled, _ = c.GetBool("mediator", "enabled")
	}
	if hasOpt = c.HasOption("mediator", "rater"); hasOpt {
		cfg.MediatorRater, _ = c.GetString("mediator", "rater")
	}
	if hasOpt = c.HasOption("mediator", "reconnects"); hasOpt {
		cfg.MediatorReconnects, _ = c.GetInt("mediator", "reconnects")
	}
	if hasOpt = c.HasOption("mediator", "cdrstats"); hasOpt {
		cfg.MediatorStats, _ = c.GetString("mediator", "cdrstats")
	}
	if hasOpt = c.HasOption("mediator", "store_disable"); hasOpt {
		cfg.MediatorStoreDisable, _ = c.GetBool("mediator", "store_disable")
	}
	if hasOpt = c.HasOption("session_manager", "enabled"); hasOpt {
		cfg.SMEnabled, _ = c.GetBool("session_manager", "enabled")
	}
	if hasOpt = c.HasOption("session_manager", "switch_type"); hasOpt {
		cfg.SMSwitchType, _ = c.GetString("session_manager", "switch_type")
	}
	if hasOpt = c.HasOption("session_manager", "rater"); hasOpt {
		cfg.SMRater, _ = c.GetString("session_manager", "rater")
	}
	if hasOpt = c.HasOption("session_manager", "cdrs"); hasOpt {
		cfg.SMCdrS, _ = c.GetString("session_manager", "cdrs")
	}
	if hasOpt = c.HasOption("session_manager", "reconnects"); hasOpt {
		cfg.SMReconnects, _ = c.GetInt("session_manager", "reconnects")
	}
	if hasOpt = c.HasOption("session_manager", "debit_interval"); hasOpt {
		cfg.SMDebitInterval, _ = c.GetInt("session_manager", "debit_interval")
	}
	if hasOpt = c.HasOption("session_manager", "min_call_duration"); hasOpt {
		minCallDurStr, _ := c.GetString("session_manager", "min_call_duration")
		if cfg.SMMinCallDuration, err = utils.ParseDurationWithSecs(minCallDurStr); err != nil {
			return nil, err
		}
	}
	if hasOpt = c.HasOption("session_manager", "max_call_duration"); hasOpt {
		maxCallDurStr, _ := c.GetString("session_manager", "max_call_duration")
		if cfg.SMMaxCallDuration, err = utils.ParseDurationWithSecs(maxCallDurStr); err != nil {
			return nil, err
		}
	}
	if hasOpt = c.HasOption("freeswitch", "server"); hasOpt {
		cfg.FreeswitchServer, _ = c.GetString("freeswitch", "server")
	}
	if hasOpt = c.HasOption("freeswitch", "passwd"); hasOpt {
		cfg.FreeswitchPass, _ = c.GetString("freeswitch", "passwd")
	}
	if hasOpt = c.HasOption("freeswitch", "reconnects"); hasOpt {
		cfg.FreeswitchReconnects, _ = c.GetInt("freeswitch", "reconnects")
	}
	if hasOpt = c.HasOption("freeswitch", "min_dur_low_balance"); hasOpt {
		minDurStr, _ := c.GetString("freeswitch", "min_dur_low_balance")
		if cfg.FSMinDurLowBalance, err = utils.ParseDurationWithSecs(minDurStr); err != nil {
			return nil, err
		}
	}
	if hasOpt = c.HasOption("freeswitch", "low_balance_ann_file"); hasOpt {
		cfg.FSLowBalanceAnnFile, _ = c.GetString("freeswitch", "low_balance_ann_file")
	}
	if hasOpt = c.HasOption("freeswitch", "empty_balance_context"); hasOpt {
		cfg.FSEmptyBalanceContext, _ = c.GetString("freeswitch", "empty_balance_context")
	}
	if hasOpt = c.HasOption("freeswitch", "empty_balance_ann_file"); hasOpt {
		cfg.FSEmptyBalanceAnnFile, _ = c.GetString("freeswitch", "empty_balance_ann_file")
	}
	if hasOpt = c.HasOption("freeswitch", "cdr_extra_fields"); hasOpt {
		extraFieldsStr, _ := c.GetString("freeswitch", "cdr_extra_fields")
		if extraFields, err := utils.ParseRSRFields(extraFieldsStr, utils.FIELDS_SEP); err != nil {
			return nil, err
		} else {
			cfg.FSCdrExtraFields = extraFields
		}
	}
	if hasOpt = c.HasOption("opensips", "listen_udp"); hasOpt {
		cfg.OsipsListenUdp, _ = c.GetString("opensips", "listen_udp")
	}
	if hasOpt = c.HasOption("opensips", "mi_addr"); hasOpt {
		cfg.OsipsMiAddr, _ = c.GetString("opensips", "mi_addr")
	}
	if hasOpt = c.HasOption("opensips", "events_subscribe_interval"); hasOpt {
		evSubscIntervalStr, _ := c.GetString("opensips", "events_subscribe_interval")
		if cfg.OsipsEvSubscInterval, err = utils.ParseDurationWithSecs(evSubscIntervalStr); err != nil {
			return nil, err
		}
	}
	if hasOpt = c.HasOption("opensips", "reconnects"); hasOpt {
		cfg.OsipsReconnects, _ = c.GetInt("opensips", "reconnects")
	}
	if cfg.DerivedChargers, err = ParseCfgDerivedCharging(c); err != nil {
		return nil, err
	}
	if hasOpt = c.HasOption("derived_charging", "combined_chargers"); hasOpt {
		cfg.CombinedDerivedChargers, _ = c.GetBool("derived_charging", "combined_chargers")
	}
	if hasOpt = c.HasOption("history_agent", "enabled"); hasOpt {
		cfg.HistoryAgentEnabled, _ = c.GetBool("history_agent", "enabled")
	}
	if hasOpt = c.HasOption("history_agent", "server"); hasOpt {
		cfg.HistoryServer, _ = c.GetString("history_agent", "server")
	}
	if hasOpt = c.HasOption("history_server", "enabled"); hasOpt {
		cfg.HistoryServerEnabled, _ = c.GetBool("history_server", "enabled")
	}
	if hasOpt = c.HasOption("history_server", "history_dir"); hasOpt {
		cfg.HistoryDir, _ = c.GetString("history_server", "history_dir")
	}
	if hasOpt = c.HasOption("history_server", "save_interval"); hasOpt {
		saveIntvlStr, _ := c.GetString("history_server", "save_interval")
		if cfg.HistorySaveInterval, err = utils.ParseDurationWithSecs(saveIntvlStr); err != nil {
			return nil, err
		}
	}
	if hasOpt = c.HasOption("mailer", "server"); hasOpt {
		cfg.MailerServer, _ = c.GetString("mailer", "server")
	}
	if hasOpt = c.HasOption("mailer", "auth_user"); hasOpt {
		cfg.MailerAuthUser, _ = c.GetString("mailer", "auth_user")
	}
	if hasOpt = c.HasOption("mailer", "auth_passwd"); hasOpt {
		cfg.MailerAuthPass, _ = c.GetString("mailer", "auth_passwd")
	}
	if hasOpt = c.HasOption("mailer", "from_address"); hasOpt {
		cfg.MailerFromAddr, _ = c.GetString("mailer", "from_address")
	}
	return cfg, nil
}
