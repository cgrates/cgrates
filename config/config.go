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

	"code.google.com/p/goconf/conf"
	"github.com/cgrates/cgrates/utils"
)

const (
	DISABLED = "disabled"
	INTERNAL = "internal"
	JSON     = "json"
	GOB      = "gob"
	POSTGRES = "postgres"
	MONGO    = "mongo"
	REDIS    = "redis"
	SAME     = "same"
	FS       = "freeswitch"
)

// Holds system configuration, defaults are overwritten with values from config file if found
type CGRConfig struct {
	RatingDBType               string
	RatingDBHost               string // The host to connect to. Values that start with / are for UNIX domain sockets.
	RatingDBPort               string // The port to bind to.
	RatingDBName               string // The name of the database to connect to.
	RatingDBUser               string // The user to sign in as.
	RatingDBPass               string // The user's password.
	AccountDBType               string
	AccountDBHost               string // The host to connect to. Values that start with / are for UNIX domain sockets.
	AccountDBPort               string // The port to bind to.
	AccountDBName               string // The name of the database to connect to.
	AccountDBUser               string // The user to sign in as.
	AccountDBPass               string // The user's password.
	StorDBType               string // Should reflect the database type used to store logs
	StorDBHost               string // The host to connect to. Values that start with / are for UNIX domain sockets.
	StorDBPort               string // The port to bind to.
	StorDBName               string // The name of the database to connect to.
	StorDBUser               string // The user to sign in as.
	StorDBPass               string // The user's password.
	DBDataEncoding           string // The encoding used to store object data in strings: <msgpack|json>
	RPCEncoding              string // RPC encoding used on APIs: <gob|json>.
	DefaultReqType           string // Use this request type if not defined on top
	DefaultTOR               string // set default type of record
	DefaultTenant            string // set default tenant
	DefaultSubject           string // set default rating subject, useful in case of fallback
	RoundingMethod           string // Rounding method for the end price: <*up|*middle|*down>
	RoundingDecimals         int    // Number of decimals to round end prices at
	RaterEnabled             bool   // start standalone server (no balancer)
	RaterBalancer            string // balancer address host:port
	RaterListen              string // listening address host:port
	BalancerEnabled          bool
	BalancerListen           string // Json RPC server address
	SchedulerEnabled         bool
	CDRSEnabled              bool     // Enable CDR Server service
	CDRSListen               string   // CDRS's listening interface: <x.y.z.y:1234>.
	CDRSExtraFields          []string //Extra fields to store in CDRs
	CDRSMediator             string   // Address where to reach the Mediator. Empty for disabling mediation. <""|internal>
	CDRSExportPath           string   // Path towards exported cdrs
	CDRSExportExtraFields    []string // Extra fields list to add in exported CDRs
	SMEnabled                bool
	SMSwitchType             string
	SMRater                  string   // address where to access rater. Can be internal, direct rater address or the address of a balancer
	SMRaterReconnects        int      // Number of reconnect attempts to rater
	SMDebitInterval          int      // the period to be debited in advanced during a call (in seconds)
	MediatorEnabled          bool     // Starts Mediator service: <true|false>.
	MediatorListen           string   // Mediator's listening interface: <internal>.
	MediatorRater            string   // Address where to reach the Rater: <internal|x.y.z.y:1234>
	MediatorRaterReconnects  int      // Number of reconnects to rater before giving up.
	MediatorCDRType          string   // CDR type <freeswitch_http_json|freeswitch_file_csv>.
	MediatorAccIdField       string   // Name of field identifying accounting id used during mediation. Use index number in case of .csv cdrs.
	MediatorRunIds           []string // Identifiers for each mediation run on CDRs
	MediatorSubjectFields    []string // Name of subject fields to be used during mediation. Use index numbers in case of .csv cdrs.
	MediatorReqTypeFields    []string // Name of request type fields to be used during mediation. Use index number in case of .csv cdrs.
	MediatorDirectionFields  []string // Name of direction fields to be used during mediation. Use index numbers in case of .csv cdrs.
	MediatorTenantFields     []string // Name of tenant fields to be used during mediation. Use index numbers in case of .csv cdrs.
	MediatorTORFields        []string // Name of tor fields to be used during mediation. Use index numbers in case of .csv cdrs.
	MediatorAccountFields    []string // Name of account fields to be used during mediation. Use index numbers in case of .csv cdrs.
	MediatorDestFields       []string // Name of destination fields to be used during mediation. Use index numbers in case of .csv cdrs.
	MediatorTimeAnswerFields []string // Name of time_start fields to be used during mediation. Use index numbers in case of .csv cdrs.
	MediatorDurationFields   []string // Name of duration fields to be used during mediation. Use index numbers in case of .csv cdrs.
	MediatorCDRInDir         string   // Absolute path towards the directory where the CDRs are kept (file stored CDRs).
	MediatorCDROutDir        string   // Absolute path towards the directory where processed CDRs will be exported (file stored CDRs).
	FreeswitchServer         string   // freeswitch address host:port
	FreeswitchPass           string   // FS socket password
	FreeswitchReconnects     int      // number of times to attempt reconnect after connect fails
	HistoryAgentEnabled      bool     // Starts History as an agent: <true|false>.
	HistoryServerEnabled     bool     // Starts History as server: <true|false>.
	HistoryServer            string   // Address where to reach the master history server: <internal|x.y.z.y:1234>
	HistoryListen            string   // History server listening interface: <internal|x.y.z.y:1234>
	HistoryPath              string   // Location on disk where to store history files.
	HistorySavePeriod        string   // The timout duration between history writes
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
	self.RPCEncoding = JSON
	self.DefaultReqType = utils.RATED
	self.DefaultTOR = "0"
	self.DefaultTenant = "0"
	self.DefaultSubject = "0"
	self.RoundingMethod = utils.ROUNDING_MIDDLE
	self.RoundingDecimals = 4
	self.RaterEnabled = false
	self.RaterBalancer = DISABLED
	self.RaterListen = "127.0.0.1:2012"
	self.BalancerEnabled = false
	self.BalancerListen = "127.0.0.1:2013"
	self.SchedulerEnabled = false
	self.CDRSEnabled = false
	self.CDRSListen = "127.0.0.1:2022"
	self.CDRSExtraFields = []string{}
	self.CDRSMediator = ""
	self.CDRSExportPath = "/var/log/cgrates/cdr/out"
	self.CDRSExportExtraFields = []string{}
	self.MediatorEnabled = false
	self.MediatorListen = "127.0.0.1:2032"
	self.MediatorRater = "127.0.0.1:2012"
	self.MediatorRaterReconnects = 3
	self.MediatorCDRType = utils.FSCDR_HTTP_JSON
	self.MediatorAccIdField = "accid"
	self.MediatorRunIds = []string{"default"}
	self.MediatorSubjectFields = []string{"subject"}
	self.MediatorReqTypeFields = []string{"reqtype"}
	self.MediatorDirectionFields = []string{"direction"}
	self.MediatorTenantFields = []string{"tenant"}
	self.MediatorTORFields = []string{"tor"}
	self.MediatorAccountFields = []string{"account"}
	self.MediatorDestFields = []string{"destination"}
	self.MediatorTimeAnswerFields = []string{"time_answer"}
	self.MediatorDurationFields = []string{"duration"}
	self.MediatorCDRInDir = "/var/log/freeswitch/cdr-csv"
	self.MediatorCDROutDir = "/var/log/cgrates/cdr/out/freeswitch/csv"
	self.SMEnabled = false
	self.SMSwitchType = FS
	self.SMRater = "127.0.0.1:2012"
	self.SMRaterReconnects = 3
	self.SMDebitInterval = 10
	self.FreeswitchServer = "127.0.0.1:8021"
	self.FreeswitchPass = "ClueCon"
	self.FreeswitchReconnects = 5
	self.HistoryAgentEnabled = false
	self.HistoryServerEnabled = false
	self.HistoryServer = "127.0.0.1:2013"
	self.HistoryListen = "127.0.0.1:2013"
	self.HistoryPath = "/var/log/cgrates/history"
	self.HistorySavePeriod = "1s"
	return nil
}

func NewDefaultCGRConfig() (*CGRConfig, error) {
	cfg := &CGRConfig{}
	cfg.setDefaults()
	return cfg, nil
}

// Instantiate a new CGRConfig setting defaults or reading from file
func NewCGRConfig(cfgPath *string) (*CGRConfig, error) {
	c, err := conf.ReadConfigFile(*cfgPath)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Could not open the configuration file: %s", err))
	}
	return loadConfig(c)
}

func NewCGRConfigBytes(data []byte) (*CGRConfig, error) {
	c, err := conf.ReadConfigBytes(data)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Could not open the configuration file: %s", err))
	}
	return loadConfig(c)
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
	if hasOpt = c.HasOption("global", "rpc_encoding"); hasOpt {
		cfg.RPCEncoding, _ = c.GetString("global", "rpc_encoding")
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
	if hasOpt = c.HasOption("rater", "listen"); hasOpt {
		cfg.RaterListen, _ = c.GetString("rater", "listen")
	}
	if hasOpt = c.HasOption("balancer", "enabled"); hasOpt {
		cfg.BalancerEnabled, _ = c.GetBool("balancer", "enabled")
	}
	if hasOpt = c.HasOption("balancer", "listen"); hasOpt {
		cfg.BalancerListen, _ = c.GetString("balancer", "listen")
	}
	if hasOpt = c.HasOption("scheduler", "enabled"); hasOpt {
		cfg.SchedulerEnabled, _ = c.GetBool("scheduler", "enabled")
	}
	if hasOpt = c.HasOption("cdrs", "enabled"); hasOpt {
		cfg.CDRSEnabled, _ = c.GetBool("cdrs", "enabled")
	}
	if hasOpt = c.HasOption("cdrs", "listen"); hasOpt {
		cfg.CDRSListen, _ = c.GetString("cdrs", "listen")
	}
	if hasOpt = c.HasOption("cdrs", "extra_fields"); hasOpt {
		if cfg.CDRSExtraFields, errParse = ConfigSlice(c, "cdrs", "extra_fields"); errParse != nil {
			return nil, errParse
		}
	}
	if hasOpt = c.HasOption("cdrs", "mediator"); hasOpt {
		cfg.CDRSMediator, _ = c.GetString("cdrs", "mediator")
	}
	if hasOpt = c.HasOption("cdrs", "export_path"); hasOpt {
		cfg.CDRSExportPath, _ = c.GetString("cdrs", "export_path")
	}
	if hasOpt = c.HasOption("cdrs", "export_extra_fields"); hasOpt {
		if cfg.CDRSExportExtraFields, errParse = ConfigSlice(c, "cdrs", "export_extra_fields"); errParse != nil {
			return nil, errParse
		}
	}
	if hasOpt = c.HasOption("mediator", "enabled"); hasOpt {
		cfg.MediatorEnabled, _ = c.GetBool("mediator", "enabled")
	}
	if hasOpt = c.HasOption("mediator", "listen"); hasOpt {
		cfg.MediatorListen, _ = c.GetString("mediator", "listen")
	}
	if hasOpt = c.HasOption("mediator", "rater"); hasOpt {
		cfg.MediatorRater, _ = c.GetString("mediator", "rater")
	}
	if hasOpt = c.HasOption("mediator", "rater_reconnects"); hasOpt {
		cfg.MediatorRaterReconnects, _ = c.GetInt("mediator", "rater_reconnects")
	}
	if hasOpt = c.HasOption("mediator", "cdr_type"); hasOpt {
		cfg.MediatorCDRType, _ = c.GetString("mediator", "cdr_type")
	}
	if hasOpt = c.HasOption("mediator", "accid_field"); hasOpt {
		cfg.MediatorAccIdField, _ = c.GetString("mediator", "accid_field")
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
	if hasOpt = c.HasOption("mediator", "time_answer_fields"); hasOpt {
		if cfg.MediatorTimeAnswerFields, errParse = ConfigSlice(c, "mediator", "time_answer_fields"); errParse != nil {
			return nil, errParse
		}
	}
	if hasOpt = c.HasOption("mediator", "duration_fields"); hasOpt {
		if cfg.MediatorDurationFields, errParse = ConfigSlice(c, "mediator", "duration_fields"); errParse != nil {
			return nil, errParse
		}
	}
	if hasOpt = c.HasOption("mediator", "cdr_in_dir"); hasOpt {
		cfg.MediatorCDRInDir, _ = c.GetString("mediator", "cdr_in_dir")
	}
	if hasOpt = c.HasOption("mediator", "cdr_out_dir"); hasOpt {
		cfg.MediatorCDROutDir, _ = c.GetString("mediator", "cdr_out_dir")
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
	if hasOpt = c.HasOption("history_server", "listen"); hasOpt {
		cfg.HistoryListen, _ = c.GetString("history_server", "listen")
	}
	if hasOpt = c.HasOption("history_server", "path"); hasOpt {
		cfg.HistoryPath, _ = c.GetString("history_server", "path")
	}
	if hasOpt = c.HasOption("history_server", "save_period"); hasOpt {
		cfg.HistorySavePeriod, _ = c.GetString("history_server", "save_period")
	}
	return cfg, nil
}
