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
)

// this function will overwrite default values with the ones present in the config file
func readConfig(c *conf.ConfigFile) {
	var hasOpt bool
	if hasOpt = c.HasOption("global", "datadb_type"); hasOpt {
		data_db_type, _ = c.GetString("global", "datadb_type")
	}
	if hasOpt = c.HasOption("global", "datadb_host"); hasOpt {
		data_db_host, _ = c.GetString("global", "datadb_host")
	}
	if hasOpt = c.HasOption("global", "datadb_port"); hasOpt {
		data_db_port, _ = c.GetString("global", "datadb_port")
	}
	if hasOpt = c.HasOption("global", "datadb_name"); hasOpt {
		data_db_name, _ = c.GetString("global", "datadb_name")
	}
	if hasOpt = c.HasOption("global", "datadb_user"); hasOpt {
		data_db_user, _ = c.GetString("global", "datadb_user")
	}
	if hasOpt = c.HasOption("global", "datadb_passwd"); hasOpt {
		data_db_pass, _ = c.GetString("global", "datadb_passwd")
	}
	if hasOpt = c.HasOption("global", "logdb_type"); hasOpt {
		log_db_type, _ = c.GetString("global", "logdb_type")
	}
	if hasOpt = c.HasOption("global", "logdb_host"); hasOpt {
		log_db_host, _ = c.GetString("global", "logdb_host")
	}
	if hasOpt = c.HasOption("global", "logdb_port"); hasOpt {
		log_db_port, _ = c.GetString("global", "logdb_port")
	}
	if hasOpt = c.HasOption("global", "logdb_name"); hasOpt {
		log_db_name, _ = c.GetString("global", "logdb_name")
	}
	if hasOpt = c.HasOption("global", "logdb_user"); hasOpt {
		log_db_user, _ = c.GetString("global", "logdb_user")
	}
	if hasOpt = c.HasOption("global", "logdb_passwd"); hasOpt {
		log_db_pass, _ = c.GetString("global", "logdb_passwd")
	}
	if hasOpt = c.HasOption("rater", "enabled"); hasOpt {
		rater_enabled, _ = c.GetBool("rater", "enabled")
	}
	if hasOpt = c.HasOption("rater", "balancer"); hasOpt {
		rater_balancer, _ = c.GetString("rater", "balancer")
	}
	if hasOpt = c.HasOption("rater", "listen"); hasOpt {
		rater_listen, _ = c.GetString("rater", "listen")
	}
	if hasOpt = c.HasOption("rater", "rpc_encoding"); hasOpt {
		rater_rpc_encoding, _ = c.GetString("rater", "rpc_encoding")
	}
	if hasOpt = c.HasOption("balancer", "enabled"); hasOpt {
		balancer_enabled, _ = c.GetBool("balancer", "enabled")
	}
	if hasOpt = c.HasOption("balancer", "listen"); hasOpt {
		balancer_listen, _ = c.GetString("balancer", "listen")
	}
	if hasOpt = c.HasOption("balancer", "rpc_encoding"); hasOpt {
		balancer_rpc_encoding, _ = c.GetString("balancer", "rpc_encoding")
	}
	if hasOpt = c.HasOption("scheduler", "enabled"); hasOpt {
		scheduler_enabled, _ = c.GetBool("scheduler", "enabled")
	}
	if hasOpt = c.HasOption("session_manager", "enabled"); hasOpt {
		sm_enabled, _ = c.GetBool("session_manager", "enabled")
	}
	if hasOpt = c.HasOption("session_manager", "switch_type"); hasOpt {
		sm_switch_type, _ = c.GetString("session_manager", "switch_type")
	}
	if hasOpt = c.HasOption("session_manager", "rater"); hasOpt {
		sm_rater, _ = c.GetString("session_manager", "rater")
	}
	if hasOpt = c.HasOption("session_manager", "debit_period"); hasOpt {
		sm_debit_period, _ = c.GetInt("session_manager", "debit_period")
	}
	if hasOpt = c.HasOption("session_manager", "rpc_encoding"); hasOpt {
		sm_rpc_encoding, _ = c.GetString("session_manager", "rpc_encoding")
	}
	if hasOpt = c.HasOption("mediator", "enabled"); hasOpt {
		mediator_enabled, _ = c.GetBool("mediator", "enabled")
	}
	if hasOpt = c.HasOption("mediator", "cdr_path"); hasOpt {
		mediator_cdr_path, _ = c.GetString("mediator", "cdr_path")
	}
	if hasOpt = c.HasOption("mediator", "cdr_out_path"); hasOpt {
		mediator_cdr_out_path, _ = c.GetString("mediator", "cdr_out_path")
	}
	if hasOpt = c.HasOption("mediator", "rater"); hasOpt {
		mediator_rater, _ = c.GetString("mediator", "rater")
	}
	if hasOpt = c.HasOption("mediator", "rpc_encoding"); hasOpt {
		mediator_rpc_encoding, _ = c.GetString("mediator", "rpc_encoding")
	}
	if hasOpt = c.HasOption("mediator", "skipdb"); hasOpt {
		mediator_skipdb, _ = c.GetBool("mediator", "skipdb")
	}
	if hasOpt = c.HasOption("mediator", "pseudo_prepaid"); hasOpt {
		mediator_pseudo_prepaid, _ = c.GetBool("mediator", "pseudo_prepaid")
	}
	if hasOpt = c.HasOption("freeswitch", "server"); hasOpt {
		freeswitch_server, _ = c.GetString("freeswitch", "server")
	}
	if hasOpt = c.HasOption("freeswitch", "pass"); hasOpt {
		freeswitch_pass, _ = c.GetString("freeswitch", "pass")
	}
	if hasOpt = c.HasOption("freeswitch", "tor_index"); hasOpt {
		freeswitch_tor, _ = c.GetString("freeswitch", "tor_index")
	}
	if hasOpt = c.HasOption("freeswitch", "tenant_index"); hasOpt {
		freeswitch_tenant, _ = c.GetString("freeswitch", "tenant_index")
	}
	if hasOpt = c.HasOption("freeswitch", "direction_index"); hasOpt {
		freeswitch_direction, _ = c.GetString("freeswitch", "direction_index")
	}
	if hasOpt = c.HasOption("freeswitch", "subject_index"); hasOpt {
		freeswitch_subject, _ = c.GetString("freeswitch", "subject_index")
	}
	if hasOpt = c.HasOption("freeswitch", "account_index"); hasOpt {
		freeswitch_account, _ = c.GetString("freeswitch", "account_index")
	}
	if hasOpt = c.HasOption("freeswitch", "destination_index"); hasOpt {
		freeswitch_destination, _ = c.GetString("freeswitch", "destination_index")
	}
	if hasOpt = c.HasOption("freeswitch", "time_start_index"); hasOpt {
		freeswitch_time_start, _ = c.GetString("freeswitch", "time_start_index")
	}
	if hasOpt = c.HasOption("freeswitch", "duration_index"); hasOpt {
		freeswitch_duration, _ = c.GetString("freeswitch", "duration_index")
	}
	if hasOpt = c.HasOption("freeswitch", "uuid_index"); hasOpt {
		freeswitch_uuid, _ = c.GetString("freeswitch", "uuid_index")
	}
	if hasOpt = c.HasOption("freeswitch", "reconnects"); hasOpt {
		freeswitch_reconnects, _ = c.GetInt("freeswitch", "reconnects")
	}
}
