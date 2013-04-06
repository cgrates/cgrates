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

package main

import (
	"code.google.com/p/goconf/conf"
	"errors"
	"fmt"
)

// Holds system configuration, defaults are overwritten with values from config file if found
type CGRConfig struct {
	data_db_type            string
	data_db_host            string // The host to connect to. Values that start with / are for UNIX domain sockets.
	data_db_port            string // The port to bind to.
	data_db_name            string // The name of the database to connect to.
	data_db_user            string // The user to sign in as.
	data_db_pass            string // The user's password.
	log_db_type             string // Should reflect the database type used to store logs
	log_db_host             string // The host to connect to. Values that start with / are for UNIX domain sockets.
	log_db_port             string // The port to bind to.
	log_db_name             string // The name of the database to connect to.
	log_db_user             string // The user to sign in as.
	log_db_pass             string // The user's password.
	rater_enabled           bool   // start standalone server (no balancer)
	rater_balancer          string // balancer address host:port
	rater_listen            string // listening address host:port
	rater_rpc_encoding      string // use JSON for RPC encoding
	balancer_enabled        bool
	balancer_listen         string // Json RPC server address
	balancer_rpc_encoding   string // use JSON for RPC encoding
	scheduler_enabled       bool
	sm_enabled              bool
	sm_switch_type          string
	sm_rater                string // address where to access rater. Can be internal, direct rater address or the address of a balancer
	sm_debit_period         int    // the period to be debited in advanced during a call (in seconds)
	sm_rpc_encoding         string // use JSON for RPC encoding
	sm_default_tor          string // set default type of record label to 0
	sm_default_tenant       string // set default tenant to 0
	sm_default_subject      string // set default rating subject to 0
	mediator_enabled        bool
	mediator_cdr_path       string // Freeswitch Master CSV CDR path.
	mediator_cdr_out_path   string // Freeswitch Master CSV CDR output path.
	mediator_rater          string // address where to access rater. Can be internal, direct rater address or the address of a balancer
	mediator_rpc_encoding   string // use JSON for RPC encoding
	mediator_skipdb         bool
	mediator_pseudo_prepaid bool
	freeswitch_server       string // freeswitch address host:port
	freeswitch_pass         string // reeswitch address host:port
	freeswitch_direction    string
	freeswitch_tor          string
	freeswitch_tenant       string
	freeswitch_subject      string
	freeswitch_account      string
	freeswitch_destination  string
	freeswitch_time_start   string
	freeswitch_duration     string
	freeswitch_uuid         string
	freeswitch_reconnects   int
}

// Instantiate a new CGRConfig setting defaults or reading from file
func NewCGRConfig(cfgPath *string) (*CGRConfig, error) {
	c, err := conf.ReadConfigFile(*cfgPath)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Could not open the configuration file: %s", err))
	}
	cfg := &CGRConfig{}
	var hasOpt bool
	cfg.data_db_type = REDIS
	if hasOpt = c.HasOption("global", "datadb_type"); hasOpt {
		cfg.data_db_type, _ = c.GetString("global", "datadb_type")
	}
	cfg.data_db_host = "localhost"
	if hasOpt = c.HasOption("global", "datadb_host"); hasOpt {
		cfg.data_db_host, _ = c.GetString("global", "datadb_host")
	}
	cfg.data_db_port = ""
	if hasOpt = c.HasOption("global", "datadb_port"); hasOpt {
		cfg.data_db_port, _ = c.GetString("global", "datadb_port")
	}
	cfg.data_db_name = "10"
	if hasOpt = c.HasOption("global", "datadb_name"); hasOpt {
		cfg.data_db_name, _ = c.GetString("global", "datadb_name")
	}
	cfg.data_db_user = ""
	if hasOpt = c.HasOption("global", "datadb_user"); hasOpt {
		cfg.data_db_user, _ = c.GetString("global", "datadb_user")
	}
	cfg.data_db_pass = ""
	if hasOpt = c.HasOption("global", "datadb_passwd"); hasOpt {
		cfg.data_db_pass, _ = c.GetString("global", "datadb_passwd")
	}
	cfg.log_db_type = MONGO
	if hasOpt = c.HasOption("global", "logdb_type"); hasOpt {
		cfg.log_db_type, _ = c.GetString("global", "logdb_type")
	}
	cfg.log_db_host = "localhost"
	if hasOpt = c.HasOption("global", "logdb_host"); hasOpt {
		cfg.log_db_host, _ = c.GetString("global", "logdb_host")
	}
	cfg.log_db_port = ""
	if hasOpt = c.HasOption("global", "logdb_port"); hasOpt {
		cfg.log_db_port, _ = c.GetString("global", "logdb_port")
	}
	cfg.log_db_name = "cgrates"
	if hasOpt = c.HasOption("global", "logdb_name"); hasOpt {
		cfg.log_db_name, _ = c.GetString("global", "logdb_name")
	}
	cfg.log_db_user = ""
	if hasOpt = c.HasOption("global", "logdb_user"); hasOpt {
		cfg.log_db_user, _ = c.GetString("global", "logdb_user")
	}
	cfg.log_db_pass = ""
	if hasOpt = c.HasOption("global", "logdb_passwd"); hasOpt {
		cfg.log_db_pass, _ = c.GetString("global", "logdb_passwd")
	}
	cfg.rater_enabled = false
	if hasOpt = c.HasOption("rater", "enabled"); hasOpt {
		cfg.rater_enabled, _ = c.GetBool("rater", "enabled")
	}
	cfg.rater_balancer = DISABLED
	if hasOpt = c.HasOption("rater", "balancer"); hasOpt {
		cfg.rater_balancer, _ = c.GetString("rater", "balancer")
	}
	cfg.rater_listen = "127.0.0.1:1234"
	if hasOpt = c.HasOption("rater", "listen"); hasOpt {
		cfg.rater_listen, _ = c.GetString("rater", "listen")
	}
	cfg.rater_rpc_encoding = GOB
	if hasOpt = c.HasOption("rater", "rpc_encoding"); hasOpt {
		cfg.rater_rpc_encoding, _ = c.GetString("rater", "rpc_encoding")
	}
	cfg.balancer_enabled = false
	if hasOpt = c.HasOption("balancer", "enabled"); hasOpt {
		cfg.balancer_enabled, _ = c.GetBool("balancer", "enabled")
	}
	cfg.balancer_listen = "127.0.0.1:2001"
	if hasOpt = c.HasOption("balancer", "listen"); hasOpt {
		cfg.balancer_listen, _ = c.GetString("balancer", "listen")
	}
	cfg.balancer_rpc_encoding = GOB
	if hasOpt = c.HasOption("balancer", "rpc_encoding"); hasOpt {
		cfg.balancer_rpc_encoding, _ = c.GetString("balancer", "rpc_encoding")
	}
	cfg.scheduler_enabled = false
	if hasOpt = c.HasOption("scheduler", "enabled"); hasOpt {
		cfg.scheduler_enabled, _ = c.GetBool("scheduler", "enabled")
	}
	cfg.mediator_enabled = false
	if hasOpt = c.HasOption("mediator", "enabled"); hasOpt {
		cfg.mediator_enabled, _ = c.GetBool("mediator", "enabled")
	}
	cfg.mediator_cdr_path = ""
	if hasOpt = c.HasOption("mediator", "cdr_path"); hasOpt {
		cfg.mediator_cdr_path, _ = c.GetString("mediator", "cdr_path")
	}
	cfg.mediator_cdr_out_path = ""
	if hasOpt = c.HasOption("mediator", "cdr_out_path"); hasOpt {
		cfg.mediator_cdr_out_path, _ = c.GetString("mediator", "cdr_out_path")
	}
	cfg.mediator_rater = INTERNAL
	if hasOpt = c.HasOption("mediator", "rater"); hasOpt {
		cfg.mediator_rater, _ = c.GetString("mediator", "rater")
	}
	cfg.mediator_rpc_encoding = GOB
	if hasOpt = c.HasOption("mediator", "rpc_encoding"); hasOpt {
		cfg.mediator_rpc_encoding, _ = c.GetString("mediator", "rpc_encoding")
	}
	cfg.mediator_skipdb = false
	if hasOpt = c.HasOption("mediator", "skipdb"); hasOpt {
		cfg.mediator_skipdb, _ = c.GetBool("mediator", "skipdb")
	}
	cfg.mediator_pseudo_prepaid = false
	if hasOpt = c.HasOption("mediator", "pseudo_prepaid"); hasOpt {
		cfg.mediator_pseudo_prepaid, _ = c.GetBool("mediator", "pseudo_prepaid")
	}
	cfg.sm_enabled = false
	if hasOpt = c.HasOption("session_manager", "enabled"); hasOpt {
		cfg.sm_enabled, _ = c.GetBool("session_manager", "enabled")
	}
	cfg.sm_switch_type = FS
	if hasOpt = c.HasOption("session_manager", "switch_type"); hasOpt {
		cfg.sm_switch_type, _ = c.GetString("session_manager", "switch_type")
	}
	cfg.sm_rater = INTERNAL
	if hasOpt = c.HasOption("session_manager", "rater"); hasOpt {
		cfg.sm_rater, _ = c.GetString("session_manager", "rater")
	}
	cfg.sm_debit_period = 10
	if hasOpt = c.HasOption("session_manager", "debit_period"); hasOpt {
		cfg.sm_debit_period, _ = c.GetInt("session_manager", "debit_period")
	}
	cfg.sm_rpc_encoding = GOB
	if hasOpt = c.HasOption("session_manager", "rpc_encoding"); hasOpt {
		cfg.sm_rpc_encoding, _ = c.GetString("session_manager", "rpc_encoding")
	}
	cfg.sm_default_tor = "0"
	if hasOpt = c.HasOption("session_manager", "default_tor"); hasOpt {
		cfg.sm_default_tor, _ = c.GetString("session_manager", "default_tor")
	}
	cfg.sm_default_tenant = "0"
	if hasOpt = c.HasOption("session_manager", "default_tenant"); hasOpt {
		cfg.sm_default_tenant, _ = c.GetString("session_manager", "default_tenant")
	}
	cfg.sm_default_subject = "0"
	if hasOpt = c.HasOption("session_manager", "default_subject"); hasOpt {
		cfg.sm_default_subject, _ = c.GetString("session_manager", "default_subject")
	}
	cfg.freeswitch_server = "localhost:8021"
	if hasOpt = c.HasOption("freeswitch", "server"); hasOpt {
		cfg.freeswitch_server, _ = c.GetString("freeswitch", "server")
	}
	cfg.freeswitch_pass = "ClueCon"
	if hasOpt = c.HasOption("freeswitch", "pass"); hasOpt {
		cfg.freeswitch_pass, _ = c.GetString("freeswitch", "pass")
	}
	cfg.freeswitch_reconnects = 5
	if hasOpt = c.HasOption("freeswitch", "reconnects"); hasOpt {
		cfg.freeswitch_reconnects, _ = c.GetInt("freeswitch", "reconnects")
	}
	cfg.freeswitch_tor = ""
	if hasOpt = c.HasOption("freeswitch", "tor_index"); hasOpt {
		cfg.freeswitch_tor, _ = c.GetString("freeswitch", "tor_index")
	}
	cfg.freeswitch_tenant = ""
	if hasOpt = c.HasOption("freeswitch", "tenant_index"); hasOpt {
		cfg.freeswitch_tenant, _ = c.GetString("freeswitch", "tenant_index")
	}
	cfg.freeswitch_direction = ""
	if hasOpt = c.HasOption("freeswitch", "direction_index"); hasOpt {
		cfg.freeswitch_direction, _ = c.GetString("freeswitch", "direction_index")
	}
	cfg.freeswitch_subject = ""
	if hasOpt = c.HasOption("freeswitch", "subject_index"); hasOpt {
		cfg.freeswitch_subject, _ = c.GetString("freeswitch", "subject_index")
	}
	cfg.freeswitch_account = ""
	if hasOpt = c.HasOption("freeswitch", "account_index"); hasOpt {
		cfg.freeswitch_account, _ = c.GetString("freeswitch", "account_index")
	}
	cfg.freeswitch_destination = ""
	if hasOpt = c.HasOption("freeswitch", "destination_index"); hasOpt {
		cfg.freeswitch_destination, _ = c.GetString("freeswitch", "destination_index")
	}
	cfg.freeswitch_time_start = ""
	if hasOpt = c.HasOption("freeswitch", "time_start_index"); hasOpt {
		cfg.freeswitch_time_start, _ = c.GetString("freeswitch", "time_start_index")
	}
	cfg.freeswitch_duration = ""
	if hasOpt = c.HasOption("freeswitch", "duration_index"); hasOpt {
		cfg.freeswitch_duration, _ = c.GetString("freeswitch", "duration_index")
	}
	cfg.freeswitch_uuid = ""
	if hasOpt = c.HasOption("freeswitch", "uuid_index"); hasOpt {
		cfg.freeswitch_uuid, _ = c.GetString("freeswitch", "uuid_index")
	}

	return cfg, nil

}
