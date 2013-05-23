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
	"code.google.com/p/goconf/conf"
	"errors"
	"fmt"
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
	DataDBType              string
	DataDBHost              string // The host to connect to. Values that start with / are for UNIX domain sockets.
	DataDBPort              string // The port to bind to.
	DataDBName              string // The name of the database to connect to.
	DataDBUser              string // The user to sign in as.
	DataDBPass              string // The user's password.
	LogDBType               string // Should reflect the database type used to store logs
	LogDBHost               string // The host to connect to. Values that start with / are for UNIX domain sockets.
	LogDBPort               string // The port to bind to.
	LogDBName               string // The name of the database to connect to.
	LogDBUser               string // The user to sign in as.
	LogDBPass               string // The user's password.
	RPCEncoding        	string // RPC encoding used on APIs: <gob|json>.
	DefaultReqType        string // Use this request type if not defined on top
	DefaultTOR            string // set default type of record
	DefaultTenant         string // set default tenant
	DefaultSubject        string // set default rating subject, useful in case of fallback
	RaterEnabled            bool   // start standalone server (no balancer)
	RaterBalancer           string // balancer address host:port
	RaterListen             string // listening address host:port
	BalancerEnabled         bool
	BalancerListen          string // Json RPC server address
	SchedulerEnabled        bool
	CDRSListen		string // CDRS's listening interface: <x.y.z.y:1234>.
	CDRSfsJSONEnabled	bool	// Enable the handler for FreeSWITCH JSON CDRs: <enabled|disabled>.
	CDRSMediator           string // Address where to reach the Mediator. Empty for disabling mediation. <""|internal>
	SMEnabled               bool
	SMSwitchType            string
	SMRater                 string // address where to access rater. Can be internal, direct rater address or the address of a balancer
	SMRaterReconnects       int    // Number of reconnect attempts to rater
	SMDebitInterval         int    // the period to be debited in advanced during a call (in seconds)
	MediatorEnabled         bool
	MediatorListen          string // Mediator's listening interface: <internal>.
	MediatorCDRType         string // sets the type of cdrs we are processing.
	MediatorCDRInDir        string // Freeswitch Master CSV CDR path.
	MediatorCDROutDir       string // Freeswitch Master CSV CDR output path.
	MediatorRater           string // address where to access rater. Can be internal, direct rater address or the address of a balancer
	MediatorRaterReconnects int    // Number of reconnect attempts to rater
	MediatorPseudoprepaid   bool
	FreeswitchServer        string // freeswitch address host:port
	FreeswitchPass          string // FS socket password
	FreeswitchDirectionIdx  string
	FreeswitchTORIdx        string
	FreeswitchTenantIdx     string
	FreeswitchSubjectIdx    string
	FreeswitchAccountIdx    string
	FreeswitchDestIdx       string
	FreeswitchTimeStartIdx  string
	FreeswitchDurationIdx   string
	FreeswitchUUIDIdx       string
	FreeswitchReconnects    int // number of times to attempt reconnect after connect fails
}

func ( self *CGRConfig ) setDefaults() error {
	self.DataDBType = REDIS
	self.DataDBHost = "127.0.0.1"
	self.DataDBPort = "6379"
	self.DataDBName = "10"
	self.DataDBUser = ""
	self.DataDBPass = ""
	self.LogDBType = MONGO
	self.LogDBHost = "localhost"
	self.LogDBPort = "27017"
	self.LogDBName = "cgrates"
	self.LogDBUser = ""
	self.LogDBPass = ""
	self.RPCEncoding = GOB
	self.DefaultTOR = "0"
	self.DefaultTenant = "0"
	self.DefaultSubject = "0"
	self.RaterEnabled = false
	self.RaterBalancer = DISABLED
	self.RaterListen = "127.0.0.1:2012"
	self.BalancerEnabled = false
	self.BalancerListen = "127.0.0.1:2013"
	self.SchedulerEnabled = false
	self.CDRSListen = "127.0.0.1:2022"
	self.CDRSfsJSONEnabled = false
	self.CDRSMediator = INTERNAL
	self.MediatorEnabled = false
	self.MediatorListen = "127.0.0.1:2032"
	self.MediatorCDRInDir = "/var/log/freeswitch/cdr-csv"
	self.MediatorCDROutDir = "/var/log/cgrates/cdr_out"
	self.MediatorRater = "127.0.0.1:2012"
	self.MediatorRaterReconnects = 3
	self.MediatorPseudoprepaid = false
	self.MediatorCDRType = "freeswitch_csv"
	self.SMEnabled = false
	self.SMSwitchType = FS
	self.SMRater = "127.0.0.1:2012"
	self.SMRaterReconnects = 3
	self.SMDebitInterval = 10
	self.DefaultReqType = ""
	self.FreeswitchServer = "127.0.0.1:8021"
	self.FreeswitchPass = "ClueCon"
	self.FreeswitchReconnects = 5
	self.FreeswitchUUIDIdx = "10"
	self.FreeswitchTORIdx = "-1"
	self.FreeswitchTenantIdx = "-1"
	self.FreeswitchDirectionIdx = "-1"
	self.FreeswitchSubjectIdx = "-1"
	self.FreeswitchAccountIdx = "-1"
	self.FreeswitchDestIdx = "-1"
	self.FreeswitchTimeStartIdx = "-1"
	self.FreeswitchDurationIdx = "-1"

	return nil
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
	if hasOpt = c.HasOption("global", "datadb_type"); hasOpt {
		cfg.DataDBType, _ = c.GetString("global", "datadb_type")
	}
	if hasOpt = c.HasOption("global", "datadb_host"); hasOpt {
		cfg.DataDBHost, _ = c.GetString("global", "datadb_host")
	}
	if hasOpt = c.HasOption("global", "datadb_port"); hasOpt {
		cfg.DataDBPort, _ = c.GetString("global", "datadb_port")
	}
	if hasOpt = c.HasOption("global", "datadb_name"); hasOpt {
		cfg.DataDBName, _ = c.GetString("global", "datadb_name")
	}
	if hasOpt = c.HasOption("global", "datadb_user"); hasOpt {
		cfg.DataDBUser, _ = c.GetString("global", "datadb_user")
	}
	if hasOpt = c.HasOption("global", "datadb_passwd"); hasOpt {
		cfg.DataDBPass, _ = c.GetString("global", "datadb_passwd")
	}
	if hasOpt = c.HasOption("global", "logdb_type"); hasOpt {
		cfg.LogDBType, _ = c.GetString("global", "logdb_type")
	}
	if hasOpt = c.HasOption("global", "logdb_host"); hasOpt {
		cfg.LogDBHost, _ = c.GetString("global", "logdb_host")
	}
	if hasOpt = c.HasOption("global", "logdb_port"); hasOpt {
		cfg.LogDBPort, _ = c.GetString("global", "logdb_port")
	}
	if hasOpt = c.HasOption("global", "logdb_name"); hasOpt {
		cfg.LogDBName, _ = c.GetString("global", "logdb_name")
	}
	if hasOpt = c.HasOption("global", "logdb_user"); hasOpt {
		cfg.LogDBUser, _ = c.GetString("global", "logdb_user")
	}
	if hasOpt = c.HasOption("global", "logdb_passwd"); hasOpt {
		cfg.LogDBPass, _ = c.GetString("global", "logdb_passwd")
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
	if hasOpt = c.HasOption("cdrs", "listen"); hasOpt {
		cfg.CDRSListen, _ = c.GetString("cdrs", "listen")
	}
	if hasOpt = c.HasOption("cdrs", "freeswitch_json_enabled"); hasOpt {
		cfg.CDRSfsJSONEnabled, _ = c.GetBool("cdrs", "freeswitch_json_enabled")
	}
	if hasOpt = c.HasOption("cdrs", "mediator"); hasOpt {
		cfg.CDRSMediator, _ = c.GetString("cdrs", "mediator")
	}
	if hasOpt = c.HasOption("mediator", "enabled"); hasOpt {
		cfg.MediatorEnabled, _ = c.GetBool("mediator", "enabled")
	}
	if hasOpt = c.HasOption("mediator", "listen"); hasOpt {
		cfg.MediatorListen, _ = c.GetString("mediator", "listen")
	}
	if hasOpt = c.HasOption("mediator", "cdr_in_dir"); hasOpt {
		cfg.MediatorCDRInDir, _ = c.GetString("mediator", "cdr_in_dir")
	}
	if hasOpt = c.HasOption("mediator", "cdr_out_dir"); hasOpt {
		cfg.MediatorCDROutDir, _ = c.GetString("mediator", "cdr_out_dir")
	}
	if hasOpt = c.HasOption("mediator", "rater"); hasOpt {
		cfg.MediatorRater, _ = c.GetString("mediator", "rater")
	}
	if hasOpt = c.HasOption("mediator", "rater_reconnects"); hasOpt {
		cfg.MediatorRaterReconnects, _ = c.GetInt("mediator", "rater_reconnects")
	}
	if hasOpt = c.HasOption("mediator", "pseudoprepaid"); hasOpt {
		cfg.MediatorPseudoprepaid, _ = c.GetBool("mediator", "pseudoprepaid")
	}
	if hasOpt = c.HasOption("mediator", "cdr_type"); hasOpt {
		cfg.MediatorCDRType, _ = c.GetString("mediator", "cdr_type")
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
	if hasOpt = c.HasOption("freeswitch", "uuid_index"); hasOpt {
		cfg.FreeswitchUUIDIdx, _ = c.GetString("freeswitch", "uuid_index")
	}
	if hasOpt = c.HasOption("freeswitch", "tor_index"); hasOpt {
		cfg.FreeswitchTORIdx, _ = c.GetString("freeswitch", "tor_index")
	}
	if hasOpt = c.HasOption("freeswitch", "tenant_index"); hasOpt {
		cfg.FreeswitchTenantIdx, _ = c.GetString("freeswitch", "tenant_index")
	}
	if hasOpt = c.HasOption("freeswitch", "direction_index"); hasOpt {
		cfg.FreeswitchDirectionIdx, _ = c.GetString("freeswitch", "direction_index")
	}
	if hasOpt = c.HasOption("freeswitch", "subject_index"); hasOpt {
		cfg.FreeswitchSubjectIdx, _ = c.GetString("freeswitch", "subject_index")
	}
	if hasOpt = c.HasOption("freeswitch", "account_index"); hasOpt {
		cfg.FreeswitchAccountIdx, _ = c.GetString("freeswitch", "account_index")
	}
	if hasOpt = c.HasOption("freeswitch", "destination_index"); hasOpt {
		cfg.FreeswitchDestIdx, _ = c.GetString("freeswitch", "destination_index")
	}
	if hasOpt = c.HasOption("freeswitch", "time_start_index"); hasOpt {
		cfg.FreeswitchTimeStartIdx, _ = c.GetString("freeswitch", "time_start_index")
	}
	if hasOpt = c.HasOption("freeswitch", "duration_index"); hasOpt {
		cfg.FreeswitchDurationIdx, _ = c.GetString("freeswitch", "duration_index")
	}

	return cfg, nil
}
