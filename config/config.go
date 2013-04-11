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
	DataDBType             string
	DataDBHost             string // The host to connect to. Values that start with / are for UNIX domain sockets.
	DataDBPort             string // The port to bind to.
	DataDBName             string // The name of the database to connect to.
	DataDBUser             string // The user to sign in as.
	DataDBPass             string // The user's password.
	LogDBType              string // Should reflect the database type used to store logs
	LogDBHost              string // The host to connect to. Values that start with / are for UNIX domain sockets.
	LogDBPort              string // The port to bind to.
	LogDBName              string // The name of the database to connect to.
	LogDBUser              string // The user to sign in as.
	LogDBPass              string // The user's password.
	RaterEnabled           bool   // start standalone server (no balancer)
	RaterBalancer          string // balancer address host:port
	RaterListen            string // listening address host:port
	RaterRPCEncoding       string // use JSON for RPC encoding
	BalancerEnabled        bool
	BalancerListen         string // Json RPC server address
	BalancerRPCEncoding    string // use JSON for RPC encoding
	SchedulerEnabled       bool
	SMEnabled              bool
	SMSwitchType           string
	SMRater                string // address where to access rater. Can be internal, direct rater address or the address of a balancer
	SMDebitInterval          int    // the period to be debited in advanced during a call (in seconds)
	SMRPCEncoding          string // use JSON for RPC encoding
	SMDefaultReqType       string // Use this request type if not defined on top
	SMDefaultTOR           string // set default type of record
	SMDefaultTenant        string // set default tenant
	SMDefaultSubject       string // set default rating subject, useful in case of fallback
	MediatorEnabled        bool
	MediatorCDRType		string // sets the type of cdrs we are processing.
	MediatorCDRInDir        string // Freeswitch Master CSV CDR path.
	MediatorCDROutDir     string // Freeswitch Master CSV CDR output path.
	MediatorRater          string // address where to access rater. Can be internal, direct rater address or the address of a balancer
	MediatorRPCEncoding    string // use JSON for RPC encoding
	MediatorSkipDB         bool
	MediatorPseudoprepaid  bool
	FreeswitchServer       string // freeswitch address host:port
	FreeswitchPass         string // FS socket password
	FreeswitchDirectionIdx string
	FreeswitchTORIdx       string
	FreeswitchTenantIdx    string
	FreeswitchSubjectIdx   string
	FreeswitchAccountIdx   string
	FreeswitchDestIdx      string
	FreeswitchTimeStartIdx string
	FreeswitchDurationIdx  string
	FreeswitchUUIDIdx      string
	FreeswitchReconnects   int // number of times to attempt reconnect after connect fails
}

// Instantiate a new CGRConfig setting defaults or reading from file
func NewCGRConfig(cfgPath *string) (*CGRConfig, error) {
	c, err := conf.ReadConfigFile(*cfgPath)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Could not open the configuration file: %s", err))
	}
	cfg := &CGRConfig{}
	var hasOpt bool
	cfg.DataDBType = REDIS
	if hasOpt = c.HasOption("global", "datadb_type"); hasOpt {
		cfg.DataDBType, _ = c.GetString("global", "datadb_type")
	}
	cfg.DataDBHost = "127.0.0.1"
	if hasOpt = c.HasOption("global", "datadb_host"); hasOpt {
		cfg.DataDBHost, _ = c.GetString("global", "datadb_host")
	}
	cfg.DataDBPort = "6379"
	if hasOpt = c.HasOption("global", "datadb_port"); hasOpt {
		cfg.DataDBPort, _ = c.GetString("global", "datadb_port")
	}
	cfg.DataDBName = "10"
	if hasOpt = c.HasOption("global", "datadb_name"); hasOpt {
		cfg.DataDBName, _ = c.GetString("global", "datadb_name")
	}
	cfg.DataDBUser = ""
	if hasOpt = c.HasOption("global", "datadb_user"); hasOpt {
		cfg.DataDBUser, _ = c.GetString("global", "datadb_user")
	}
	cfg.DataDBPass = ""
	if hasOpt = c.HasOption("global", "datadb_passwd"); hasOpt {
		cfg.DataDBPass, _ = c.GetString("global", "datadb_passwd")
	}
	cfg.LogDBType = MONGO
	if hasOpt = c.HasOption("global", "logdb_type"); hasOpt {
		cfg.LogDBType, _ = c.GetString("global", "logdb_type")
	}
	cfg.LogDBHost = "localhost"
	if hasOpt = c.HasOption("global", "logdb_host"); hasOpt {
		cfg.LogDBHost, _ = c.GetString("global", "logdb_host")
	}
	cfg.LogDBPort = "27017"
	if hasOpt = c.HasOption("global", "logdb_port"); hasOpt {
		cfg.LogDBPort, _ = c.GetString("global", "logdb_port")
	}
	cfg.LogDBName = "cgrates"
	if hasOpt = c.HasOption("global", "logdb_name"); hasOpt {
		cfg.LogDBName, _ = c.GetString("global", "logdb_name")
	}
	cfg.LogDBUser = "cgrates"
	if hasOpt = c.HasOption("global", "logdb_user"); hasOpt {
		cfg.LogDBUser, _ = c.GetString("global", "logdb_user")
	}
	cfg.LogDBPass = "CGRateS.org"
	if hasOpt = c.HasOption("global", "logdb_passwd"); hasOpt {
		cfg.LogDBPass, _ = c.GetString("global", "logdb_passwd")
	}
	cfg.RaterEnabled = false
	if hasOpt = c.HasOption("rater", "enabled"); hasOpt {
		cfg.RaterEnabled, _ = c.GetBool("rater", "enabled")
	}
	cfg.RaterBalancer = DISABLED
	if hasOpt = c.HasOption("rater", "balancer"); hasOpt {
		cfg.RaterBalancer, _ = c.GetString("rater", "balancer")
	}
	cfg.RaterListen = "127.0.0.1:2012"
	if hasOpt = c.HasOption("rater", "listen"); hasOpt {
		cfg.RaterListen, _ = c.GetString("rater", "listen")
	}
	cfg.RaterRPCEncoding = GOB
	if hasOpt = c.HasOption("rater", "rpc_encoding"); hasOpt {
		cfg.RaterRPCEncoding, _ = c.GetString("rater", "rpc_encoding")
	}
	cfg.BalancerEnabled = false
	if hasOpt = c.HasOption("balancer", "enabled"); hasOpt {
		cfg.BalancerEnabled, _ = c.GetBool("balancer", "enabled")
	}
	cfg.BalancerListen = "127.0.0.1:2013"
	if hasOpt = c.HasOption("balancer", "listen"); hasOpt {
		cfg.BalancerListen, _ = c.GetString("balancer", "listen")
	}
	cfg.BalancerRPCEncoding = GOB
	if hasOpt = c.HasOption("balancer", "rpc_encoding"); hasOpt {
		cfg.BalancerRPCEncoding, _ = c.GetString("balancer", "rpc_encoding")
	}
	cfg.SchedulerEnabled = false
	if hasOpt = c.HasOption("scheduler", "enabled"); hasOpt {
		cfg.SchedulerEnabled, _ = c.GetBool("scheduler", "enabled")
	}
	cfg.MediatorEnabled = false
	if hasOpt = c.HasOption("mediator", "enabled"); hasOpt {
		cfg.MediatorEnabled, _ = c.GetBool("mediator", "enabled")
	}
	cfg.MediatorCDRInDir = "/var/log/freeswitch/cdr-csv/"
	if hasOpt = c.HasOption("mediator", "cdr_in_dir"); hasOpt {
		cfg.MediatorCDRInDir, _ = c.GetString("mediator", "cdr_in_dir")
	}
	cfg.MediatorCDROutDir = "/var/log/cgrates/cdr_out/"
	if hasOpt = c.HasOption("mediator", "cdr_out_dir"); hasOpt {
		cfg.MediatorCDROutDir, _ = c.GetString("mediator", "cdr_out_dir")
	}
	cfg.MediatorRater = "127.0.0.1:2012"
	if hasOpt = c.HasOption("mediator", "rater"); hasOpt {
		cfg.MediatorRater, _ = c.GetString("mediator", "rater")
	}
	cfg.MediatorRPCEncoding = GOB
	if hasOpt = c.HasOption("mediator", "rpc_encoding"); hasOpt {
		cfg.MediatorRPCEncoding, _ = c.GetString("mediator", "rpc_encoding")
	}
	cfg.MediatorSkipDB = false
	if hasOpt = c.HasOption("mediator", "skipdb"); hasOpt {
		cfg.MediatorSkipDB, _ = c.GetBool("mediator", "skipdb")
	}
	cfg.MediatorPseudoprepaid = false
	if hasOpt = c.HasOption("mediator", "pseudoprepaid"); hasOpt {
		cfg.MediatorPseudoprepaid, _ = c.GetBool("mediator", "pseudoprepaid")
	}
	cfg.MediatorCDRType = "freeswitch_csv"
	if hasOpt = c.HasOption("mediator", "cdr_type"); hasOpt {
		cfg.MediatorCDRType, _ = c.GetString("mediator", "cdr_type")
	}
	cfg.SMEnabled = false
	if hasOpt = c.HasOption("session_manager", "enabled"); hasOpt {
		cfg.SMEnabled, _ = c.GetBool("session_manager", "enabled")
	}
	cfg.SMSwitchType = FS
	if hasOpt = c.HasOption("session_manager", "switch_type"); hasOpt {
		cfg.SMSwitchType, _ = c.GetString("session_manager", "switch_type")
	}
	cfg.SMRater = "127.0.0.1:2012"
	if hasOpt = c.HasOption("session_manager", "rater"); hasOpt {
		cfg.SMRater, _ = c.GetString("session_manager", "rater")
	}
	cfg.SMDebitInterval = 10
	if hasOpt = c.HasOption("session_manager", "debit_interval"); hasOpt {
		cfg.SMDebitInterval, _ = c.GetInt("session_manager", "debit_interval")
	}
	cfg.SMRPCEncoding = GOB
	if hasOpt = c.HasOption("session_manager", "rpc_encoding"); hasOpt {
		cfg.SMRPCEncoding, _ = c.GetString("session_manager", "rpc_encoding")
	}
	cfg.SMDefaultReqType = "" // By default CGRateS is inactive, customer should activate when he feels he is ready
	if hasOpt = c.HasOption("session_manager", "default_reqtype"); hasOpt {
		cfg.SMDefaultReqType, _ = c.GetString("session_manager", "default_reqtype")
	}
	cfg.SMDefaultTOR = "0"
	if hasOpt = c.HasOption("session_manager", "default_tor"); hasOpt {
		cfg.SMDefaultTOR, _ = c.GetString("session_manager", "default_tor")
	}
	cfg.SMDefaultTenant = "0"
	if hasOpt = c.HasOption("session_manager", "default_tenant"); hasOpt {
		cfg.SMDefaultTenant, _ = c.GetString("session_manager", "default_tenant")
	}
	cfg.SMDefaultSubject = "0"
	if hasOpt = c.HasOption("session_manager", "default_subject"); hasOpt {
		cfg.SMDefaultSubject, _ = c.GetString("session_manager", "default_subject")
	}
	cfg.FreeswitchServer = "127.0.0.1:8021"
	if hasOpt = c.HasOption("freeswitch", "server"); hasOpt {
		cfg.FreeswitchServer, _ = c.GetString("freeswitch", "server")
	}
	cfg.FreeswitchPass = "ClueCon"
	if hasOpt = c.HasOption("freeswitch", "passwd"); hasOpt {
		cfg.FreeswitchPass, _ = c.GetString("freeswitch", "passwd")
	}
	cfg.FreeswitchReconnects = 5
	if hasOpt = c.HasOption("freeswitch", "reconnects"); hasOpt {
		cfg.FreeswitchReconnects, _ = c.GetInt("freeswitch", "reconnects")
	}
	cfg.FreeswitchTORIdx = ""
	if hasOpt = c.HasOption("freeswitch", "tor_index"); hasOpt {
		cfg.FreeswitchTORIdx, _ = c.GetString("freeswitch", "tor_index")
	}
	cfg.FreeswitchTenantIdx = ""
	if hasOpt = c.HasOption("freeswitch", "tenant_index"); hasOpt {
		cfg.FreeswitchTenantIdx, _ = c.GetString("freeswitch", "tenant_index")
	}
	cfg.FreeswitchDirectionIdx = ""
	if hasOpt = c.HasOption("freeswitch", "direction_index"); hasOpt {
		cfg.FreeswitchDirectionIdx, _ = c.GetString("freeswitch", "direction_index")
	}
	cfg.FreeswitchSubjectIdx = ""
	if hasOpt = c.HasOption("freeswitch", "subject_index"); hasOpt {
		cfg.FreeswitchSubjectIdx, _ = c.GetString("freeswitch", "subject_index")
	}
	cfg.FreeswitchAccountIdx = ""
	if hasOpt = c.HasOption("freeswitch", "account_index"); hasOpt {
		cfg.FreeswitchAccountIdx, _ = c.GetString("freeswitch", "account_index")
	}
	cfg.FreeswitchDestIdx = ""
	if hasOpt = c.HasOption("freeswitch", "destination_index"); hasOpt {
		cfg.FreeswitchDestIdx, _ = c.GetString("freeswitch", "destination_index")
	}
	cfg.FreeswitchTimeStartIdx = ""
	if hasOpt = c.HasOption("freeswitch", "time_start_index"); hasOpt {
		cfg.FreeswitchTimeStartIdx, _ = c.GetString("freeswitch", "time_start_index")
	}
	cfg.FreeswitchDurationIdx = ""
	if hasOpt = c.HasOption("freeswitch", "duration_index"); hasOpt {
		cfg.FreeswitchDurationIdx, _ = c.GetString("freeswitch", "duration_index")
	}
	cfg.FreeswitchUUIDIdx = ""
	if hasOpt = c.HasOption("freeswitch", "uuid_index"); hasOpt {
		cfg.FreeswitchUUIDIdx, _ = c.GetString("freeswitch", "uuid_index")
	}

	return cfg, nil

}
