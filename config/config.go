/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) 2012-2014 ITsysCOM GmbH

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
	DBDataEncoding          string             // The encoding used to store object data in strings: <msgpack|json>
	RPCJSONListen           string             // RPC JSON listening address
	RPCGOBListen            string             // RPC GOB listening address
	HTTPListen              string             // HTTP listening address
	DefaultReqType          string             // Use this request type if not defined on top
	DefaultCategory         string             // set default type of record
	DefaultTenant           string             // set default tenant
	DefaultSubject          string             // set default rating subject, useful in case of fallback
	RoundingDecimals        int                // Number of decimals to round end prices at
	XmlCfgDocument          *CgrXmlCfgDocument // Load additional configuration inside xml document
	RaterEnabled            bool               // start standalone server (no balancer)
	RaterBalancer           string             // balancer address host:port
	BalancerEnabled         bool
	SchedulerEnabled        bool
	CDRSEnabled             bool                       // Enable CDR Server service
	CDRSExtraFields         []*utils.RSRField          // Extra fields to store in CDRs
	CDRSMediator            string                     // Address where to reach the Mediator. Empty for disabling mediation. <""|internal>
	CdreCdrFormat           string                     // Format of the exported CDRs. <csv>
	CdreMaskDestId          string                     // Id of the destination list to be masked in CDRs
	CdreMaskLength          int                        // Number of digits to mask in the destination suffix if destination is in the MaskDestinationdsId
	CdreCostShiftDigits     int                        // Shift digits in the cost on export (eg: convert from EUR to cents)
	CdreDir                 string                     // Path towards exported cdrs directory
	CdreExportedFields      []*utils.RSRField          // List of fields in the exported CDRs
	CdreFWXmlTemplate       *CgrXmlCdreFwCfg           // Use this configuration as export template in case of fixed fields length
	CdrcEnabled             bool                       // Enable CDR client functionality
	CdrcCdrs                string                     // Address where to reach CDR server
	CdrcRunDelay            time.Duration              // Sleep interval between consecutive runs, 0 to use automation via inotify
	CdrcCdrType             string                     // CDR file format <csv>.
	CdrcCdrInDir            string                     // Absolute path towards the directory where the CDRs are stored.
	CdrcCdrOutDir           string                     // Absolute path towards the directory where processed CDRs will be moved.
	CdrcSourceId            string                     // Tag identifying the source of the CDRs within CGRS database.
	CdrcCdrFields           map[string]*utils.RSRField // FieldName/RSRField format. Index number in case of .csv cdrs.
	SMEnabled               bool
	SMSwitchType            string
	SMRater                 string                // address where to access rater. Can be internal, direct rater address or the address of a balancer
	SMRaterReconnects       int                   // Number of reconnect attempts to rater
	SMDebitInterval         int                   // the period to be debited in advanced during a call (in seconds)
	SMMaxCallDuration       time.Duration         // The maximum duration of a call
	MediatorEnabled         bool                  // Starts Mediator service: <true|false>.
	MediatorRater           string                // Address where to reach the Rater: <internal|x.y.z.y:1234>
	MediatorRaterReconnects int                   // Number of reconnects to rater before giving up.
	DerivedChargers         utils.DerivedChargers // System wide derived chargers, added to the account level ones
	CombinedDerivedChargers bool                  // Combine accounts specific derived_chargers with server configured
	FreeswitchServer        string                // freeswitch address host:port
	FreeswitchPass          string                // FS socket password
	FreeswitchReconnects    int                   // number of times to attempt reconnect after connect fails
	HistoryAgentEnabled     bool                  // Starts History as an agent: <true|false>.
	HistoryServer           string                // Address where to reach the master history server: <internal|x.y.z.y:1234>
	HistoryServerEnabled    bool                  // Starts History as server: <true|false>.
	HistoryDir              string                // Location on disk where to store history files.
	HistorySaveInterval     time.Duration         // The timout duration between history writes
	MailerServer            string                // The server to use when sending emails out
	MailerAuthUser          string                // Authenticate to email server using this user
	MailerAuthPass          string                // Authenticate to email server with this password
	MailerFromAddr          string                // From address used when sending emails out
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
	self.DBDataEncoding = utils.MSGPACK
	self.RPCJSONListen = "127.0.0.1:2012"
	self.RPCGOBListen = "127.0.0.1:2013"
	self.HTTPListen = "127.0.0.1:2080"
	self.DefaultReqType = utils.RATED
	self.DefaultCategory = "call"
	self.DefaultTenant = "cgrates.org"
	self.DefaultSubject = "cgrates"
	self.RoundingDecimals = 10
	self.XmlCfgDocument = nil
	self.RaterEnabled = false
	self.RaterBalancer = ""
	self.BalancerEnabled = false
	self.SchedulerEnabled = false
	self.CDRSEnabled = false
	self.CDRSExtraFields = []*utils.RSRField{}
	self.CDRSMediator = ""
	self.CdreCdrFormat = "csv"
	self.CdreMaskDestId = ""
	self.CdreMaskLength = 0
	self.CdreCostShiftDigits = 0
	self.CdreDir = "/var/log/cgrates/cdre"
	self.CdrcEnabled = false
	self.CdrcCdrs = utils.INTERNAL
	self.CdrcRunDelay = time.Duration(0)
	self.CdrcCdrType = utils.CSV
	self.CdrcCdrInDir = "/var/log/cgrates/cdrc/in"
	self.CdrcCdrOutDir = "/var/log/cgrates/cdrc/out"
	self.CdrcSourceId = "freeswitch_csv"
	self.CdrcCdrFields = map[string]*utils.RSRField{
		utils.ACCID:       &utils.RSRField{Id: "0"},
		utils.REQTYPE:     &utils.RSRField{Id: "1"},
		utils.DIRECTION:   &utils.RSRField{Id: "2"},
		utils.TENANT:      &utils.RSRField{Id: "3"},
		utils.CATEGORY:    &utils.RSRField{Id: "4"},
		utils.ACCOUNT:     &utils.RSRField{Id: "5"},
		utils.SUBJECT:     &utils.RSRField{Id: "6"},
		utils.DESTINATION: &utils.RSRField{Id: "7"},
		utils.SETUP_TIME:  &utils.RSRField{Id: "8"},
		utils.ANSWER_TIME: &utils.RSRField{Id: "9"},
		utils.USAGE:       &utils.RSRField{Id: "10"},
	}
	self.MediatorEnabled = false
	self.MediatorRater = "internal"
	self.MediatorRaterReconnects = 3
	self.DerivedChargers = make(utils.DerivedChargers, 0)
	self.CombinedDerivedChargers = true
	self.SMEnabled = false
	self.SMSwitchType = FS
	self.SMRater = "internal"
	self.SMRaterReconnects = 3
	self.SMDebitInterval = 10
	self.SMMaxCallDuration = time.Duration(3) * time.Hour
	self.FreeswitchServer = "127.0.0.1:8021"
	self.FreeswitchPass = "ClueCon"
	self.FreeswitchReconnects = 5
	self.HistoryAgentEnabled = false
	self.HistoryServerEnabled = false
	self.HistoryServer = "internal"
	self.HistoryDir = "/var/log/cgrates/history"
	self.HistorySaveInterval = time.Duration(1) * time.Second
	self.MailerServer = "localhost:25"
	self.MailerAuthUser = "cgrates"
	self.MailerAuthPass = "CGRateS.org"
	self.MailerFromAddr = "cgr-mailer@localhost.localdomain"
	self.CdreExportedFields = []*utils.RSRField{
		&utils.RSRField{Id: utils.CGRID},
		&utils.RSRField{Id: utils.MEDI_RUNID},
		&utils.RSRField{Id: utils.TOR},
		&utils.RSRField{Id: utils.ACCID},
		&utils.RSRField{Id: utils.CDRHOST},
		&utils.RSRField{Id: utils.REQTYPE},
		&utils.RSRField{Id: utils.DIRECTION},
		&utils.RSRField{Id: utils.TENANT},
		&utils.RSRField{Id: utils.CATEGORY},
		&utils.RSRField{Id: utils.ACCOUNT},
		&utils.RSRField{Id: utils.SUBJECT},
		&utils.RSRField{Id: utils.DESTINATION},
		&utils.RSRField{Id: utils.SETUP_TIME},
		&utils.RSRField{Id: utils.ANSWER_TIME},
		&utils.RSRField{Id: utils.USAGE},
		&utils.RSRField{Id: utils.COST},
	}
	return nil
}

func (self *CGRConfig) checkConfigSanity() error {
	// Cdre sanity check for fixed_width
	if self.CdreCdrFormat == utils.CDRE_FIXED_WIDTH {
		if self.XmlCfgDocument == nil {
			return errors.New("Need XmlConfigurationDocument for fixed_width cdr export")
		} else if self.CdreFWXmlTemplate == nil {
			return errors.New("Need XmlTemplate for fixed_width cdr export")
		}
	}
	if self.CdrcCdrType == utils.CSV {
		for _, rsrFld := range self.CdrcCdrFields {
			if _, errConv := strconv.Atoi(rsrFld.Id); errConv != nil {
				return fmt.Errorf("CDR fields must be indices in case of .csv files, have instead: %s", rsrFld.Id)
			}
		}
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
		if extraFields, err := ParseRSRFields(extraFieldsStr); err != nil {
			return nil, err
		} else {
			cfg.CDRSExtraFields = extraFields
		}
	}
	if hasOpt = c.HasOption("cdrs", "mediator"); hasOpt {
		cfg.CDRSMediator, _ = c.GetString("cdrs", "mediator")
	}
	if hasOpt = c.HasOption("cdre", "cdr_format"); hasOpt {
		cfg.CdreCdrFormat, _ = c.GetString("cdre", "cdr_format")
	}
	if hasOpt = c.HasOption("cdre", "mask_destination_id"); hasOpt {
		cfg.CdreMaskDestId, _ = c.GetString("cdre", "mask_destination_id")
	}
	if hasOpt = c.HasOption("cdre", "mask_length"); hasOpt {
		cfg.CdreMaskLength, _ = c.GetInt("cdre", "mask_length")
	}
	if hasOpt = c.HasOption("cdre", "cost_shift_digits"); hasOpt {
		cfg.CdreCostShiftDigits, _ = c.GetInt("cdre", "cost_shift_digits")
	}
	if hasOpt = c.HasOption("cdre", "export_template"); hasOpt { // Load configs for csv normally from template, fixed_width from xml file
		exportTemplate, _ := c.GetString("cdre", "export_template")
		if cfg.CdreCdrFormat != utils.CDRE_FIXED_WIDTH { // Csv most likely
			if extraFields, err := ParseRSRFields(exportTemplate); err != nil {
				return nil, err
			} else {
				cfg.CdreExportedFields = extraFields
			}
		} else if strings.HasPrefix(exportTemplate, utils.XML_PROFILE_PREFIX) {
			if xmlTemplate, err := cfg.XmlCfgDocument.GetCdreFWCfg(exportTemplate[len(utils.XML_PROFILE_PREFIX):]); err != nil {
				return nil, err
			} else {
				cfg.CdreFWXmlTemplate = xmlTemplate
			}
		}
	}
	if hasOpt = c.HasOption("cdre", "export_dir"); hasOpt {
		cfg.CdreDir, _ = c.GetString("cdre", "export_dir")
	}
	if hasOpt = c.HasOption("cdrc", "enabled"); hasOpt {
		cfg.CdrcEnabled, _ = c.GetBool("cdrc", "enabled")
	}
	if hasOpt = c.HasOption("cdrc", "cdrs"); hasOpt {
		cfg.CdrcCdrs, _ = c.GetString("cdrc", "cdrs")
	}
	if hasOpt = c.HasOption("cdrc", "run_delay"); hasOpt {
		durStr, _ := c.GetString("cdrc", "run_delay")
		if cfg.CdrcRunDelay, err = utils.ParseDurationWithSecs(durStr); err != nil {
			return nil, err
		}
	}
	if hasOpt = c.HasOption("cdrc", "cdr_type"); hasOpt {
		cfg.CdrcCdrType, _ = c.GetString("cdrc", "cdr_type")
	}
	if hasOpt = c.HasOption("cdrc", "cdr_in_dir"); hasOpt {
		cfg.CdrcCdrInDir, _ = c.GetString("cdrc", "cdr_in_dir")
	}
	if hasOpt = c.HasOption("cdrc", "cdr_out_dir"); hasOpt {
		cfg.CdrcCdrOutDir, _ = c.GetString("cdrc", "cdr_out_dir")
	}
	if hasOpt = c.HasOption("cdrc", "cdr_source_id"); hasOpt {
		cfg.CdrcSourceId, _ = c.GetString("cdrc", "cdr_source_id")
	}
	// ParseCdrcCdrFields
	accIdFld, _ := c.GetString("cdrc", "accid_field")
	reqtypeFld, _ := c.GetString("cdrc", "reqtype_field")
	directionFld, _ := c.GetString("cdrc", "direction_field")
	tenantFld, _ := c.GetString("cdrc", "tenant_field")
	categoryFld, _ := c.GetString("cdrc", "category_field")
	acntFld, _ := c.GetString("cdrc", "account_field")
	subjectFld, _ := c.GetString("cdrc", "subject_field")
	destFld, _ := c.GetString("cdrc", "destination_field")
	setupTimeFld, _ := c.GetString("cdrc", "setup_time_field")
	answerTimeFld, _ := c.GetString("cdrc", "answer_time_field")
	durFld, _ := c.GetString("cdrc", "usage_field")
	extraFlds, _ := c.GetString("cdrc", "extra_fields")
	if cfg.CdrcCdrFields, err = ParseCdrcCdrFields(accIdFld, reqtypeFld, directionFld, tenantFld, categoryFld, acntFld, subjectFld, destFld,
		setupTimeFld, answerTimeFld, durFld, extraFlds); err != nil {
		return nil, err
	}
	if hasOpt = c.HasOption("mediator", "enabled"); hasOpt {
		cfg.MediatorEnabled, _ = c.GetBool("mediator", "enabled")
	}
	if hasOpt = c.HasOption("mediator", "rater"); hasOpt {
		cfg.MediatorRater, _ = c.GetString("mediator", "rater")
	}
	if hasOpt = c.HasOption("mediator", "rater_reconnects"); hasOpt {
		cfg.MediatorRaterReconnects, _ = c.GetInt("mediator", "rater_reconnects")
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
	if hasOpt = c.HasOption("session_manager", "rater_reconnects"); hasOpt {
		cfg.SMRaterReconnects, _ = c.GetInt("session_manager", "rater_reconnects")
	}
	if hasOpt = c.HasOption("session_manager", "debit_interval"); hasOpt {
		cfg.SMDebitInterval, _ = c.GetInt("session_manager", "debit_interval")
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
