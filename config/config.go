/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2013 ITsysCOM

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
	RatingDBType             string
	RatingDBHost             string // The host to connect to. Values that start with / are for UNIX domain sockets.
	RatingDBPort             string // The port to bind to.
	RatingDBName             string // The name of the database to connect to.
	RatingDBUser             string // The user to sign in as.
	RatingDBPass             string // The user's password.
	AccountDBType            string
	AccountDBHost            string // The host to connect to. Values that start with / are for UNIX domain sockets.
	AccountDBPort            string // The port to bind to.
	AccountDBName            string // The name of the database to connect to.
	AccountDBUser            string // The user to sign in as.
	AccountDBPass            string // The user's password.
	StorDBType               string // Should reflect the database type used to store logs
	StorDBHost               string // The host to connect to. Values that start with / are for UNIX domain sockets.
	StorDBPort               string // Th e port to bind to.
	StorDBName               string // The name of the database to connect to.
	StorDBUser               string // The user to sign in as.
	StorDBPass               string // The user's password.
	DBDataEncoding           string // The encoding used to store object data in strings: <msgpack|json>
	RPCJSONListen            string // RPC JSON listening address
	RPCGOBListen             string // RPC GOB listening address
	HTTPListen               string // HTTP listening address
	DefaultReqType           string // Use this request type if not defined on top
	DefaultTOR               string // set default type of record
	DefaultTenant            string // set default tenant
	DefaultSubject           string // set default rating subject, useful in case of fallback
	RoundingMethod           string // Rounding method for the end price: <*up|*middle|*down>
	RoundingDecimals         int    // Number of decimals to round end prices at
	RaterEnabled             bool   // start standalone server (no balancer)
	RaterBalancer            string // balancer address host:port
	BalancerEnabled          bool
	SchedulerEnabled         bool
	CDRSEnabled              bool                              // Enable CDR Server service
	CDRSExtraFields          []string                          // Extra fields to store in CDRs
	CDRSSearchReplaceRules   map[string]*utils.ReSearchReplace // Key will be the field name, values their compiled ReSearchReplace rules
	CDRSMediator             string                            // Address where to reach the Mediator. Empty for disabling mediation. <""|internal>
	CdreCdrFormat            string                            // Format of the exported CDRs. <csv>
	CdreExtraFields          []string                          // Extra fields list to add in exported CDRs
	CdreDir                  string                            // Path towards exported cdrs directory
	CdrcEnabled              bool                              // Enable CDR client functionality
	CdrcCdrs                 string                            // Address where to reach CDR server
	CdrcCdrsMethod           string                            // Mechanism to use when posting CDRs on server  <http_cgr>
	CdrcRunDelay             time.Duration                     // Sleep interval between consecutive runs, 0 to use automation via inotify
	CdrcCdrType              string                            // CDR file format <csv>.
	CdrcCdrInDir             string                            // Absolute path towards the directory where the CDRs are stored.
	CdrcCdrOutDir            string                            // Absolute path towards the directory where processed CDRs will be moved.
	CdrcSourceId             string                            // Tag identifying the source of the CDRs within CGRS database.
	CdrcAccIdField           string                            // Accounting id field identifier. Use index number in case of .csv cdrs.
	CdrcReqTypeField         string                            // Request type field identifier. Use index number in case of .csv cdrs.
	CdrcDirectionField       string                            // Direction field identifier. Use index numbers in case of .csv cdrs.
	CdrcTenantField          string                            // Tenant field identifier. Use index numbers in case of .csv cdrs.
	CdrcTorField             string                            // Type of Record field identifier. Use index numbers in case of .csv cdrs.
	CdrcAccountField         string                            // Account field identifier. Use index numbers in case of .csv cdrs.
	CdrcSubjectField         string                            // Subject field identifier. Use index numbers in case of .csv CDRs.
	CdrcDestinationField     string                            // Destination field identifier. Use index numbers in case of .csv cdrs.
	CdrcSetupTimeField       string                            // Setup time field identifier. Use index numbers in case of .csv cdrs.
	CdrcAnswerTimeField      string                            // Answer time field identifier. Use index numbers in case of .csv cdrs.
	CdrcDurationField        string                            // Duration field identifier. Use index numbers in case of .csv cdrs.
	CdrcExtraFields          []string                          // Extra fields to extract, special format in case of .csv "field1:index1,field2:index2"
	SMEnabled                bool
	SMSwitchType             string
	SMRater                  string        // address where to access rater. Can be internal, direct rater address or the address of a balancer
	SMRaterReconnects        int           // Number of reconnect attempts to rater
	SMDebitInterval          int           // the period to be debited in advanced during a call (in seconds)
	SMMaxCallDuration        time.Duration // The maximum duration of a call
	SMRunIds                 []string      // Identifiers of additional sessions control.
	SMReqTypeFields          []string      // Name of request type fields to be used during additional sessions control <""|*default|field_name>.
	SMDirectionFields        []string      // Name of direction fields to be used during additional sessions control <""|*default|field_name>.
	SMTenantFields           []string      // Name of tenant fields to be used during additional sessions control <""|*default|field_name>.
	SMTORFields              []string      // Name of tor fields to be used during additional sessions control <""|*default|field_name>.
	SMAccountFields          []string      // Name of account fields to be used during additional sessions control <""|*default|field_name>.
	SMSubjectFields          []string      // Name of fields to be used during additional sessions control <""|*default|field_name>.
	SMDestFields             []string      // Name of destination fields to be used during additional sessions control <""|*default|field_name>.
	SMSetupTimeFields        []string      // Name of setup_time fields to be used during additional sessions control <""|*default|field_name>.
	SMAnswerTimeFields       []string      // Name of answer_time fields to be used during additional sessions control <""|*default|field_name>.
	SMDurationFields         []string      // Name of duration fields to be used during additional sessions control <""|*default|field_name>.
	MediatorEnabled          bool          // Starts Mediator service: <true|false>.
	MediatorRater            string        // Address where to reach the Rater: <internal|x.y.z.y:1234>
	MediatorRaterReconnects  int           // Number of reconnects to rater before giving up.
	MediatorRunIds           []string      // Identifiers for each mediation run on CDRs
	MediatorReqTypeFields    []string      // Name of request type fields to be used during mediation. Use index number in case of .csv cdrs.
	MediatorDirectionFields  []string      // Name of direction fields to be used during mediation. Use index numbers in case of .csv cdrs.
	MediatorTenantFields     []string      // Name of tenant fields to be used during mediation. Use index numbers in case of .csv cdrs.
	MediatorTORFields        []string      // Name of tor fields to be used during mediation. Use index numbers in case of .csv cdrs.
	MediatorAccountFields    []string      // Name of account fields to be used during mediation. Use index numbers in case of .csv cdrs.
	MediatorSubjectFields    []string      // Name of subject fields to be used during mediation. Use index numbers in case of .csv cdrs.
	MediatorDestFields       []string      // Name of destination fields to be used during mediation. Use index numbers in case of .csv cdrs.
	MediatorSetupTimeFields  []string      // Name of setup_time fields to be used during mediation. Use index numbers in case of .csv cdrs.
	MediatorAnswerTimeFields []string      // Name of answer_time fields to be used during mediation. Use index numbers in case of .csv cdrs.
	MediatorDurationFields   []string      // Name of duration fields to be used during mediation. Use index numbers in case of .csv cdrs.
	FreeswitchServer         string        // freeswitch address host:port
	FreeswitchPass           string        // FS socket password
	FreeswitchReconnects     int           // number of times to attempt reconnect after connect fails
	HistoryAgentEnabled      bool          // Starts History as an agent: <true|false>.
	HistoryServer            string        // Address where to reach the master history server: <internal|x.y.z.y:1234>
	HistoryServerEnabled     bool          // Starts History as server: <true|false>.
	HistoryDir               string        // Location on disk where to store history files.
	HistorySaveInterval      time.Duration // The timout duration between history writes
	MailerServer             string        // The server to use when sending emails out
	MailerAuthUser           string        // Authenticate to email server using this user
	MailerAuthPass           string        // Authenticate to email server with this password
	MailerFromAddr           string        // From address used when sending emails out
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
	self.DefaultTOR = "call"
	self.DefaultTenant = "cgrates.org"
	self.DefaultSubject = "cgrates"
	self.RoundingMethod = utils.ROUNDING_MIDDLE
	self.RoundingDecimals = 4
	self.RaterEnabled = false
	self.RaterBalancer = ""
	self.BalancerEnabled = false
	self.SchedulerEnabled = false
	self.CDRSEnabled = false
	self.CDRSExtraFields = []string{}
	self.CDRSMediator = ""
	self.CdreCdrFormat = "csv"
	self.CdreExtraFields = []string{}
	self.CdreDir = "/var/log/cgrates/cdr/cdrexport/csv"
	self.CdrcEnabled = false
	self.CdrcCdrs = utils.INTERNAL
	self.CdrcCdrsMethod = "http_cgr"
	self.CdrcRunDelay = time.Duration(0)
	self.CdrcCdrType = "csv"
	self.CdrcCdrInDir = "/var/log/cgrates/cdr/cdrc/in"
	self.CdrcCdrOutDir = "/var/log/cgrates/cdr/cdrc/out"
	self.CdrcSourceId = "freeswitch_csv"
	self.CdrcAccIdField = "0"
	self.CdrcReqTypeField = "1"
	self.CdrcDirectionField = "2"
	self.CdrcTenantField = "3"
	self.CdrcTorField = "4"
	self.CdrcAccountField = "5"
	self.CdrcSubjectField = "6"
	self.CdrcDestinationField = "7"
	self.CdrcSetupTimeField = "8"
	self.CdrcAnswerTimeField = "9"
	self.CdrcDurationField = "10"
	self.CdrcExtraFields = []string{}
	self.MediatorEnabled = false
	self.MediatorRater = "internal"
	self.MediatorRaterReconnects = 3
	self.MediatorRunIds = []string{}
	self.MediatorSubjectFields = []string{}
	self.MediatorReqTypeFields = []string{}
	self.MediatorDirectionFields = []string{}
	self.MediatorTenantFields = []string{}
	self.MediatorTORFields = []string{}
	self.MediatorAccountFields = []string{}
	self.MediatorDestFields = []string{}
	self.MediatorSetupTimeFields = []string{}
	self.MediatorAnswerTimeFields = []string{}
	self.MediatorDurationFields = []string{}
	self.SMEnabled = false
	self.SMSwitchType = FS
	self.SMRater = "internal"
	self.SMRaterReconnects = 3
	self.SMDebitInterval = 10
	self.SMMaxCallDuration = time.Duration(3) * time.Hour
	self.SMRunIds = []string{}
	self.SMReqTypeFields = []string{}
	self.SMDirectionFields = []string{}
	self.SMTenantFields = []string{}
	self.SMTORFields = []string{}
	self.SMAccountFields = []string{}
	self.SMSubjectFields = []string{}
	self.SMDestFields = []string{}
	self.SMSetupTimeFields = []string{}
	self.SMAnswerTimeFields = []string{}
	self.SMDurationFields = []string{}
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
	return nil
}

func (self *CGRConfig) checkConfigSanity() error {
	// SessionManager should have same fields config length for session emulation
	if len(self.SMReqTypeFields) != len(self.SMRunIds) ||
		len(self.SMDirectionFields) != len(self.SMRunIds) ||
		len(self.SMTenantFields) != len(self.SMRunIds) ||
		len(self.SMTORFields) != len(self.SMRunIds) ||
		len(self.SMAccountFields) != len(self.SMRunIds) ||
		len(self.SMSubjectFields) != len(self.SMRunIds) ||
		len(self.SMDestFields) != len(self.SMRunIds) ||
		len(self.SMSetupTimeFields) != len(self.SMRunIds) ||
		len(self.SMAnswerTimeFields) != len(self.SMRunIds) ||
		len(self.SMDurationFields) != len(self.SMRunIds) {
		return errors.New("<ConfigSanity> Inconsistent fields length for SessionManager session emulation")
	}
	// Mediator needs to have consistent extra fields definition
	if len(self.MediatorReqTypeFields) != len(self.MediatorRunIds) ||
		len(self.MediatorDirectionFields) != len(self.MediatorRunIds) ||
		len(self.MediatorTenantFields) != len(self.MediatorRunIds) ||
		len(self.MediatorTORFields) != len(self.MediatorRunIds) ||
		len(self.MediatorAccountFields) != len(self.MediatorRunIds) ||
		len(self.MediatorSubjectFields) != len(self.MediatorRunIds) ||
		len(self.MediatorDestFields) != len(self.MediatorRunIds) ||
		len(self.MediatorSetupTimeFields) != len(self.MediatorRunIds) ||
		len(self.MediatorAnswerTimeFields) != len(self.MediatorRunIds) ||
		len(self.MediatorDurationFields) != len(self.MediatorRunIds) {
		return errors.New("<ConfigSanity> Inconsistent fields length for Mediator extra fields")
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

// Instantiate a new CGRConfig setting defaults or reading from file
func NewCGRConfig(cfgPath *string) (*CGRConfig, error) {
	c, err := conf.ReadConfigFile(*cfgPath)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Could not open the configuration file: %s", err))
	}
	cfg, err := loadConfig(c)
	if err != nil {
		return nil, err
	}
	if err := cfg.checkConfigSanity(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func NewCGRConfigBytes(data []byte) (*CGRConfig, error) {
	c, err := conf.ReadConfigBytes(data)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Could not open the configuration file: %s", err))
	}
	cfg, err := loadConfig(c)
	if err != nil {
		return nil, err
	}
	if err := cfg.checkConfigSanity(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func loadConfig(c *conf.ConfigFile) (*CGRConfig, error) {
	cfg := &CGRConfig{}
	cfg.setDefaults()
	var hasOpt bool
	var errParse error
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
	if hasOpt = c.HasOption("global", "default_tor"); hasOpt {
		cfg.DefaultTOR, _ = c.GetString("global", "default_tor")
	}
	if hasOpt = c.HasOption("global", "default_tenant"); hasOpt {
		cfg.DefaultTenant, _ = c.GetString("global", "default_tenant")
	}
	if hasOpt = c.HasOption("global", "default_subject"); hasOpt {
		cfg.DefaultSubject, _ = c.GetString("global", "default_subject")
	}
	if hasOpt = c.HasOption("global", "rounding_method"); hasOpt {
		cfg.RoundingMethod, _ = c.GetString("global", "rounding_method")
	}
	if hasOpt = c.HasOption("global", "rounding_decimals"); hasOpt {
		cfg.RoundingDecimals, _ = c.GetInt("global", "rounding_decimals")
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
		if cfg.CDRSExtraFields, errParse = ConfigSlice(c, "cdrs", "extra_fields"); errParse != nil {
			return nil, errParse
		}
	}
	if hasOpt = c.HasOption("cdrs", "mediator"); hasOpt {
		cfg.CDRSMediator, _ = c.GetString("cdrs", "mediator")
	}
	if hasOpt = c.HasOption("cdre", "cdr_format"); hasOpt {
		cfg.CdreCdrFormat, _ = c.GetString("cdre", "cdr_format")
	}
	if hasOpt = c.HasOption("cdre", "extra_fields"); hasOpt {
		if cfg.CdreExtraFields, errParse = ConfigSlice(c, "cdre", "extra_fields"); errParse != nil {
			return nil, errParse
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
	if hasOpt = c.HasOption("cdrc", "cdrs_method"); hasOpt {
		cfg.CdrcCdrsMethod, _ = c.GetString("cdrc", "cdrs_method")
	}
	if hasOpt = c.HasOption("cdrc", "run_delay"); hasOpt {
		durStr, _ := c.GetString("cdrc", "run_delay")
		if cfg.CdrcRunDelay, errParse = utils.ParseDurationWithSecs(durStr); errParse != nil {
			return nil, errParse
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
	if hasOpt = c.HasOption("cdrc", "accid_field"); hasOpt {
		cfg.CdrcAccIdField, _ = c.GetString("cdrc", "accid_field")
	}
	if hasOpt = c.HasOption("cdrc", "reqtype_field"); hasOpt {
		cfg.CdrcReqTypeField, _ = c.GetString("cdrc", "reqtype_field")
	}
	if hasOpt = c.HasOption("cdrc", "direction_field"); hasOpt {
		cfg.CdrcDirectionField, _ = c.GetString("cdrc", "direction_field")
	}
	if hasOpt = c.HasOption("cdrc", "tenant_field"); hasOpt {
		cfg.CdrcTenantField, _ = c.GetString("cdrc", "tenant_field")
	}
	if hasOpt = c.HasOption("cdrc", "tor_field"); hasOpt {
		cfg.CdrcTorField, _ = c.GetString("cdrc", "tor_field")
	}
	if hasOpt = c.HasOption("cdrc", "account_field"); hasOpt {
		cfg.CdrcAccountField, _ = c.GetString("cdrc", "account_field")
	}
	if hasOpt = c.HasOption("cdrc", "subject_field"); hasOpt {
		cfg.CdrcSubjectField, _ = c.GetString("cdrc", "subject_field")
	}
	if hasOpt = c.HasOption("cdrc", "destination_field"); hasOpt {
		cfg.CdrcDestinationField, _ = c.GetString("cdrc", "destination_field")
	}
	if hasOpt = c.HasOption("cdrc", "setup_time_field"); hasOpt {
		cfg.CdrcSetupTimeField, _ = c.GetString("cdrc", "setup_time_field")
	}
	if hasOpt = c.HasOption("cdrc", "answer_time_field"); hasOpt {
		cfg.CdrcAnswerTimeField, _ = c.GetString("cdrc", "answer_time_field")
	}
	if hasOpt = c.HasOption("cdrc", "duration_field"); hasOpt {
		cfg.CdrcDurationField, _ = c.GetString("cdrc", "duration_field")
	}
	if hasOpt = c.HasOption("cdrc", "extra_fields"); hasOpt {
		if cfg.CdrcExtraFields, errParse = ConfigSlice(c, "cdrc", "extra_fields"); errParse != nil {
			return nil, errParse
		}
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
	if hasOpt = c.HasOption("mediator", "run_ids"); hasOpt {
		if cfg.MediatorRunIds, errParse = ConfigSlice(c, "mediator", "run_ids"); errParse != nil {
			return nil, errParse
		}
	}
	if hasOpt = c.HasOption("mediator", "subject_fields"); hasOpt {
		if cfg.MediatorSubjectFields, errParse = ConfigSlice(c, "mediator", "subject_fields"); errParse != nil {
			return nil, errParse
		}
	}
	if hasOpt = c.HasOption("mediator", "reqtype_fields"); hasOpt {
		if cfg.MediatorReqTypeFields, errParse = ConfigSlice(c, "mediator", "reqtype_fields"); errParse != nil {
			return nil, errParse
		}
	}
	if hasOpt = c.HasOption("mediator", "direction_fields"); hasOpt {
		if cfg.MediatorDirectionFields, errParse = ConfigSlice(c, "mediator", "direction_fields"); errParse != nil {
			return nil, errParse
		}
	}
	if hasOpt = c.HasOption("mediator", "tenant_fields"); hasOpt {
		if cfg.MediatorTenantFields, errParse = ConfigSlice(c, "mediator", "tenant_fields"); errParse != nil {
			return nil, errParse
		}
	}
	if hasOpt = c.HasOption("mediator", "tor_fields"); hasOpt {
		if cfg.MediatorTORFields, errParse = ConfigSlice(c, "mediator", "tor_fields"); errParse != nil {
			return nil, errParse
		}
	}
	if hasOpt = c.HasOption("mediator", "account_fields"); hasOpt {
		if cfg.MediatorAccountFields, errParse = ConfigSlice(c, "mediator", "account_fields"); errParse != nil {
			return nil, errParse
		}
	}
	if hasOpt = c.HasOption("mediator", "destination_fields"); hasOpt {
		if cfg.MediatorDestFields, errParse = ConfigSlice(c, "mediator", "destination_fields"); errParse != nil {
			return nil, errParse
		}
	}
	if hasOpt = c.HasOption("mediator", "setup_time_fields"); hasOpt {
		if cfg.MediatorSetupTimeFields, errParse = ConfigSlice(c, "mediator", "setup_time_fields"); errParse != nil {
			return nil, errParse
		}
	}
	if hasOpt = c.HasOption("mediator", "answer_time_fields"); hasOpt {
		if cfg.MediatorAnswerTimeFields, errParse = ConfigSlice(c, "mediator", "answer_time_fields"); errParse != nil {
			return nil, errParse
		}
	}
	if hasOpt = c.HasOption("mediator", "duration_fields"); hasOpt {
		if cfg.MediatorDurationFields, errParse = ConfigSlice(c, "mediator", "duration_fields"); errParse != nil {
			return nil, errParse
		}
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
		if cfg.SMMaxCallDuration, errParse = utils.ParseDurationWithSecs(maxCallDurStr); errParse != nil {
			return nil, errParse
		}
	}
	if hasOpt = c.HasOption("session_manager", "run_ids"); hasOpt {
		if cfg.SMRunIds, errParse = ConfigSlice(c, "session_manager", "run_ids"); errParse != nil {
			return nil, errParse
		}
	}
	if hasOpt = c.HasOption("session_manager", "reqtype_fields"); hasOpt {
		if cfg.SMReqTypeFields, errParse = ConfigSlice(c, "session_manager", "reqtype_fields"); errParse != nil {
			return nil, errParse
		}
	}
	if hasOpt = c.HasOption("session_manager", "direction_fields"); hasOpt {
		if cfg.SMDirectionFields, errParse = ConfigSlice(c, "session_manager", "direction_fields"); errParse != nil {
			return nil, errParse
		}
	}
	if hasOpt = c.HasOption("session_manager", "tenant_fields"); hasOpt {
		if cfg.SMTenantFields, errParse = ConfigSlice(c, "session_manager", "tenant_fields"); errParse != nil {
			return nil, errParse
		}
	}
	if hasOpt = c.HasOption("session_manager", "tor_fields"); hasOpt {
		if cfg.SMTORFields, errParse = ConfigSlice(c, "session_manager", "tor_fields"); errParse != nil {
			return nil, errParse
		}
	}
	if hasOpt = c.HasOption("session_manager", "account_fields"); hasOpt {
		if cfg.SMAccountFields, errParse = ConfigSlice(c, "session_manager", "account_fields"); errParse != nil {
			return nil, errParse
		}
	}
	if hasOpt = c.HasOption("session_manager", "subject_fields"); hasOpt {
		if cfg.SMSubjectFields, errParse = ConfigSlice(c, "session_manager", "subject_fields"); errParse != nil {
			return nil, errParse
		}
	}
	if hasOpt = c.HasOption("session_manager", "destination_fields"); hasOpt {
		if cfg.SMDestFields, errParse = ConfigSlice(c, "session_manager", "destination_fields"); errParse != nil {
			return nil, errParse
		}
	}
	if hasOpt = c.HasOption("session_manager", "setup_time_fields"); hasOpt {
		if cfg.SMSetupTimeFields, errParse = ConfigSlice(c, "session_manager", "setup_time_fields"); errParse != nil {
			return nil, errParse
		}
	}
	if hasOpt = c.HasOption("session_manager", "answer_time_fields"); hasOpt {
		if cfg.SMAnswerTimeFields, errParse = ConfigSlice(c, "session_manager", "answer_time_fields"); errParse != nil {
			return nil, errParse
		}
	}
	if hasOpt = c.HasOption("session_manager", "duration_fields"); hasOpt {
		if cfg.SMDurationFields, errParse = ConfigSlice(c, "session_manager", "duration_fields"); errParse != nil {
			return nil, errParse
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
		if cfg.HistorySaveInterval, errParse = utils.ParseDurationWithSecs(saveIntvlStr); errParse != nil {
			return nil, errParse
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
