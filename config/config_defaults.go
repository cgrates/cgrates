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

const CGRATES_CFG_JSON = `
{

// Real-time Charging System for Telecom & ISP environments
// Copyright (C) ITsysCOM GmbH
//
// This file contains the default configuration hardcoded into CGRateS.
// This is what you get when you load CGRateS with an empty configuration file.

"general": {
	"http_skip_tls_veify": false,			// if enabled Http Client will accept any TLS certificate
	"rounding_decimals": 10,				// system level precision for floats
	"dbdata_encoding": "msgpack",			// encoding used to store object data in strings: <msgpack|json>
	"tpexport_dir": "/var/log/cgrates/tpe",	// path towards export folder for offline Tariff Plans
	"default_reqtype": "rated",				// default request type to consider when missing from requests: <""|prepaid|postpaid|pseudoprepaid|rated>
	"default_category": "call",				// default Type of Record to consider when missing from requests
	"default_tenant": "cgrates.org",		// default Tenant to consider when missing from requests
	"default_subject": "cgrates",			// default rating Subject to consider when missing from requests
},


"listen": {
	"rpc_json": "127.0.0.1:2012",			// RPC JSON listening address
	"rpc_gob": "127.0.0.1:2013",			// RPC GOB listening address
	"http": "127.0.0.1:2080",				// HTTP listening address
},


"rating_db": {
	"db_type": "redis",						// rating subsystem database type: <redis>
	"db_host": "127.0.0.1",					// rating subsystem database host address
	"db_port": 6379, 						// rating subsystem port to reach the database
	"db_name": "10", 						// rating subsystem database name to connect to
	"db_user": "", 							// rating subsystem username to use when connecting to database
	"db_passwd": "", 						// rating subsystem password to use when connecting to database
},


"accounting_db": {
	"db_type": "redis",						// accounting subsystem database: <redis>
	"db_host": "127.0.0.1",					// accounting subsystem database host address
	"db_port": 6379, 						// accounting subsystem port to reach the database
	"db_name": "11", 						// accounting subsystem database name to connect to
	"db_user": "", 							// accounting subsystem username to use when connecting to database
	"db_passwd": "", 						// accounting subsystem password to use when connecting to database
},


"stor_db": {
	"db_type": "mysql",						// stor database type to use: <mysql|postgres>
	"db_host": "127.0.0.1",					// the host to connect to
	"db_port": 3306, 						// the port to reach the stordb
	"db_name": "cgrates", 					// stor database name
	"db_user": "cgrates", 					// username to use when connecting to stordb
	"db_passwd": "CGRateS.org", 			// password to use when connecting to stordb
	"max_open_conns": 0,					// maximum database connections opened
	"max_idle_conns": -1,					// maximum database connections idle
},


"balancer": {
	"enabled": false, 						// start Balancer service: <true|false>
},


"rater": {
	"enabled": false,						// enable Rater service: <true|false>
	"balancer": "",							// register to Balancer as worker: <""|internal|x.y.z.y:1234>
},


"scheduler": {
	"enabled": false,						// start Scheduler service: <true|false>
},


"cdrs": {
	"enabled": false,						// start the CDR Server service:  <true|false>
	"extra_fields": [],						// extra fields to store in CDRs for non-generic CDRs
	"mediator": "",							// address where to reach the Mediator. Empty for disabling mediation. <""|internal>
	"cdrstats": "",							// address where to reach the cdrstats service. Empty to disable stats gathering from raw CDRs <""|internal|x.y.z.y:1234>
	"store_disable": false,					// when true, CDRs will not longer be saved in stordb, useful for cdrstats only scenario
},


"mediator": {
	"enabled": false,						// starts Mediator service: <true|false>.
	"reconnects": 3,						// number of reconnects to rater/cdrs before giving up.
	"rater": "internal",					// address where to reach the Rater: <internal|x.y.z.y:1234>
	"cdrstats": "",							// address where to reach the cdrstats service. Empty to disable stats gathering out of mediated CDRs <""|internal|x.y.z.y:1234>
	"store_disable": false,					// when true, CDRs will not longer be saved in stordb, useful for cdrstats only scenario
},


"cdrstats": {
	"enabled": false,						// starts the cdrstats service: <true|false>
	"queue_length": 50,						// number of items in the stats buffer
	"time_window": "1h",					// will only keep the CDRs who's call setup time is not older than time.Now()-TimeWindow
	"metrics": ["ASR", "ACD", "ACC"],		// stat metric ids to build
	"setup_interval": [],					// filter on CDR SetupTime
	"tors":	[],								// filter on CDR TOR fields
	"cdr_hosts": [],						// filter on CDR CdrHost fields
	"cdr_sources": [], 						// filter on CDR CdrSource fields
	"req_types": [],						// filter on CDR ReqType fields
	"directions": [],						// filter on CDR Direction fields
	"tenants": [],							// filter on CDR Tenant fields
	"categories": [],						// filter on CDR Category fields
	"accounts": [],							// filter on CDR Account fields
	"subjects": [],							// filter on CDR Subject fields
	"destination_prefixes": [],				// filter on CDR Destination prefixes
	"usage_interval": [],					// filter on CDR Usage 
	"mediation_run_ids": [],				// filter on CDR MediationRunId fields
	"rated_accounts": [],					// filter on CDR RatedAccount fields
	"rated_subjects": [],					// filter on CDR RatedSubject fields
	"cost_interval": [],					// filter on CDR Cost
},


"cdre": {
	"*default": {
		"cdr_format": "csv",							// exported CDRs format <csv>
		"field_separator": ",",
		"data_usage_multiply_factor": 1,				// multiply data usage before export (eg: convert from KBytes to Bytes)
		"sms_usage_multiply_factor": 1,					// multiply data usage before export (eg: convert from SMS unit to call duration in some billing systems)
		"cost_multiply_factor": 1,						// multiply cost before export, eg: add VAT
		"cost_rounding_decimals": -1,					// rounding decimals for Cost values. -1 to disable rounding
		"cost_shift_digits": 0,							// shift digits in the cost on export (eg: convert from EUR to cents)
		"mask_destination_id": "MASKED_DESTINATIONS",	// destination id containing called addresses to be masked on export
		"mask_length": 0,								// length of the destination suffix to be masked
		"export_dir": "/var/log/cgrates/cdre",			// path where the exported CDRs will be placed
		"header_fields": [],							// template of the exported header fields
		"content_fields": [								// template of the exported content fields
			{"tag": "CgrId", "cdr_field_id": "cgrid", "type": "cdrfield", "value": "cgrid"},
			{"tag":"RunId", "cdr_field_id": "mediation_runid", "type": "cdrfield", "value": "mediation_runid"},
			{"tag":"Tor", "cdr_field_id": "tor", "type": "cdrfield", "value": "tor"},
			{"tag":"AccId", "cdr_field_id": "accid", "type": "cdrfield", "value": "accid"},
			{"tag":"ReqType", "cdr_field_id": "reqtype", "type": "cdrfield", "value": "reqtype"},
			{"tag":"Direction", "cdr_field_id": "direction", "type": "cdrfield", "value": "direction"},
			{"tag":"Tenant", "cdr_field_id": "tenant", "type": "cdrfield", "value": "tenant"},
			{"tag":"Category", "cdr_field_id": "category", "type": "cdrfield", "value": "category"},
			{"tag":"Account", "cdr_field_id": "account", "type": "cdrfield", "value": "account"},
			{"tag":"Subject", "cdr_field_id": "subject", "type": "cdrfield", "value": "subject"},
			{"tag":"Destination", "cdr_field_id": "destination", "type": "cdrfield", "value": "destination"},
			{"tag":"SetupTime", "cdr_field_id": "setup_time", "type": "cdrfield", "value": "setup_time", "layout": "2006-01-02T15:04:05Z07:00"},
			{"tag":"AnswerTime", "cdr_field_id": "answer_time", "type": "cdrfield", "value": "answer_time", "layout": "2006-01-02T15:04:05Z07:00"},
			{"tag":"Usage", "cdr_field_id": "usage", "type": "cdrfield", "value": "usage"},
			{"tag":"Cost", "cdr_field_id": "cost", "type": "cdrfield", "value": "cost"},			
		],
		"trailer_fields": [],							// template of the exported trailer fields
	}
},


"cdrc": {
	"*default": {
		"enabled": false,							// enable CDR client functionality
		"cdrs_address": "internal",					// address where to reach CDR server. <internal|x.y.z.y:1234>
		"cdr_format": "csv",						// CDR file format <csv|freeswitch_csv|fwv>
		"field_separator": ",",						// separator used in case of csv files
		"run_delay": 0,								// sleep interval in seconds between consecutive runs, 0 to use automation via inotify
		"data_usage_multiply_factor": 1024,			// conversion factor for data usage
		"cdr_in_dir": "/var/log/cgrates/cdrc/in",	// absolute path towards the directory where the CDRs are stored
		"cdr_out_dir": "/var/log/cgrates/cdrc/out",	// absolute path towards the directory where processed CDRs will be moved
		"cdr_source_id": "freeswitch_csv",			// free form field, tag identifying the source of the CDRs within CDRS database
		"cdr_filter": "",							// Filter CDR records to import
		"cdr_fields":[								// import template, tag will match internally CDR field, in case of .csv value will be represented by index of the field value
			{"tag": "tor", "cdr_field_id": "tor", "type": "cdrfield", "value": "2", "mandatory": true},
			{"tag": "accid", "cdr_field_id": "accid", "type": "cdrfield", "value": "3", "mandatory": true},
			{"tag": "reqtype", "cdr_field_id": "reqtype", "type": "cdrfield", "value": "4", "mandatory": true},
			{"tag": "direction", "cdr_field_id": "direction", "type": "cdrfield", "value": "5", "mandatory": true},
			{"tag": "tenant", "cdr_field_id": "tenant", "type": "cdrfield", "value": "6", "mandatory": true},
			{"tag": "category", "cdr_field_id": "category", "type": "cdrfield", "value": "7", "mandatory": true},
			{"tag": "account", "cdr_field_id": "account", "type": "cdrfield", "value": "8", "mandatory": true},
			{"tag": "subject", "cdr_field_id": "subject", "type": "cdrfield", "value": "9", "mandatory": true},
			{"tag": "destination", "cdr_field_id": "destination", "type": "cdrfield", "value": "10", "mandatory": true},
			{"tag": "setup_time", "cdr_field_id": "setup_time", "type": "cdrfield", "value": "11", "mandatory": true},
			{"tag": "answer_time", "cdr_field_id": "answer_time", "type": "cdrfield", "value": "12", "mandatory": true},
			{"tag": "usage", "cdr_field_id": "usage", "type": "cdrfield", "value": "13", "mandatory": true},
		],
	}
},

"sm_freeswitch": {
	"enabled": false,				// starts SessionManager service: <true|false>
	"rater": "internal",			// address where to reach the Rater <""|internal|127.0.0.1:2013>
	"cdrs": "",						// address where to reach CDR Server, empty to disable CDR capturing <""|internal|x.y.z.y:1234>
	"reconnects": 5,				// number of reconnect attempts to rater or cdrs
	"cdr_extra_fields": [],			// extra fields to store in CDRs in case of processing them
	"debit_interval": "10s",		// interval to perform debits on.
	"min_call_duration": "0s",		// only authorize calls with allowed duration higher than this
	"max_call_duration": "3h",		// maximum call duration a prepaid call can last
	"min_dur_low_balance": "5s",	// threshold which will trigger low balance warnings for prepaid calls (needs to be lower than debit_interval)
	"low_balance_ann_file": "",		// file to be played when low balance is reached for prepaid calls
	"empty_balance_context": "",	// if defined, prepaid calls will be transfered to this context on empty balance
	"empty_balance_ann_file": "",	// file to be played before disconnecting prepaid calls on empty balance (applies only if no context defined)
	"connections":[					// instantiate connections to multiple FreeSWITCH servers
		{"server": "127.0.0.1:8021", "password": "ClueCon", "reconnects": 5}
	],
},


"sm_kamailio": {
	"enabled": false,				// starts SessionManager service: <true|false>
	"rater": "internal",			// address where to reach the Rater <""|internal|127.0.0.1:2013>
	"cdrs": "",						// address where to reach CDR Server, empty to disable CDR capturing <""|internal|x.y.z.y:1234>
	"reconnects": 5,				// number of reconnect attempts to rater or cdrs
	"debit_interval": "10s",		// interval to perform debits on.
	"min_call_duration": "0s",		// only authorize calls with allowed duration higher than this
	"max_call_duration": "3h",		// maximum call duration a prepaid call can last
	"connections":[					// instantiate connections to multiple Kamailio servers
		{"evapi_addr": "127.0.0.1:8448", "reconnects": 5}
	],
},


"sm_opensips": {
	"enabled": false,					// starts SessionManager service: <true|false>
	"listen_udp": "127.0.0.1:2020",		// address where to listen for datagram events coming from OpenSIPS
	"rater": "internal",				// address where to reach the Rater <""|internal|127.0.0.1:2013>
	"cdrs": "",							// address where to reach CDR Server, empty to disable CDR capturing <""|internal|x.y.z.y:1234>
	"debit_interval": "10s",			// interval to perform debits on.
	"min_call_duration": "0s",			// only authorize calls with allowed duration higher than this
	"max_call_duration": "3h",			// maximum call duration a prepaid call can last
	"events_subscribe_interval": "60s",	// automatic events subscription to OpenSIPS, 0 to disable it
	"mi_addr": "127.0.0.1:8020",		// address where to reach OpenSIPS MI to send session disconnects
	"reconnects": 5,					// number of reconnects if connection is lost
},


"history_server": {
	"enabled": false,							// starts History service: <true|false>.
	"history_dir": "/var/log/cgrates/history",	// location on disk where to store history files.
	"save_interval": "1s",						// interval to save changed cache into .git archive
},


"history_agent": {
	"enabled": false,			// starts History as a client: <true|false>.
	"server": "internal",		// address where to reach the master history server: <internal|x.y.z.y:1234>
},


"mailer": {
	"server": "localhost",								// the server to use when sending emails out
	"auth_user": "cgrates",								// authenticate to email server using this user
	"auth_passwd": "CGRateS.org",						// authenticate to email server with this password
	"from_address": "cgr-mailer@localhost.localdomain"	// from address used when sending emails out
},


}`
