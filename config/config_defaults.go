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
	"http_skip_tls_verify": false,						// if enabled Http Client will accept any TLS certificate
	"rounding_decimals": 10,							// system level precision for floats
	"dbdata_encoding": "msgpack",						// encoding used to store object data in strings: <msgpack|json>
	"tpexport_dir": "/var/log/cgrates/tpe",				// path towards export folder for offline Tariff Plans
	"http_failed_dir": "/var/log/cgrates/http_failed",	// directory path where we store failed http requests
	"default_reqtype": "*rated",						// default request type to consider when missing from requests: <""|*prepaid|*postpaid|*pseudoprepaid|*rated>
	"default_category": "call",							// default Type of Record to consider when missing from requests
	"default_tenant": "cgrates.org",					// default Tenant to consider when missing from requests
	"default_subject": "cgrates",						// default rating Subject to consider when missing from requests
	"default_timezone": "Local",						// default timezone for timestamps where not specified <""|UTC|Local|$IANA_TZ_DB>
	"connect_attempts": 3,								// initial server connect attempts
	"reconnects": -1,									// number of retries in case of connection lost
	"response_cache_ttl": "3s",							// the life span of a cached response
	"internal_ttl": "2m",								// maximum duration to wait for internal connections before giving up
},


"listen": {
	"rpc_json": "127.0.0.1:2012",			// RPC JSON listening address
	"rpc_gob": "127.0.0.1:2013",			// RPC GOB listening address
	"http": "127.0.0.1:2080",				// HTTP listening address
},


"tariffplan_db": {							// database used to store active tariff plan configuration
	"db_type": "redis",						// tariffplan_db type: <redis>
	"db_host": "127.0.0.1",					// tariffplan_db host address
	"db_port": 6379, 						// port to reach the tariffplan_db
	"db_name": "10", 						// tariffplan_db name to connect to
	"db_user": "", 							// sername to use when connecting to tariffplan_db
	"db_passwd": "", 						// password to use when connecting to tariffplan_db
},


"data_db": {								// database used to store runtime data (eg: accounts, cdr stats)
	"db_type": "redis",						// data_db type: <redis>
	"db_host": "127.0.0.1",					// data_db host address
	"db_port": 6379, 						// data_db port to reach the database
	"db_name": "11", 						// data_db database name to connect to
	"db_user": "", 							// username to use when connecting to data_db
	"db_passwd": "", 						// password to use when connecting to data_db
	"load_history_size": 10,				// Number of records in the load history
},


"stor_db": {								// database used to store offline tariff plans and CDRs
	"db_type": "mysql",						// stor database type to use: <mysql|postgres>
	"db_host": "127.0.0.1",					// the host to connect to
	"db_port": 3306,						// the port to reach the stordb
	"db_name": "cgrates",					// stor database name
	"db_user": "cgrates",					// username to use when connecting to stordb
	"db_passwd": "CGRateS.org",				// password to use when connecting to stordb
	"max_open_conns": 100,					// maximum database connections opened
	"max_idle_conns": 10,					// maximum database connections idle
},


"balancer": {
	"enabled": false,						// start Balancer service: <true|false>
},


"rater": {
	"enabled": false,						// enable Rater service: <true|false>
	"balancer": "",							// register to balancer as worker: <""|internal|x.y.z.y:1234>
	"cdrstats": "",							// address where to reach the cdrstats service, empty to disable stats functionality: <""|internal|x.y.z.y:1234>
	"historys": "",							// address where to reach the history service, empty to disable history functionality: <""|internal|x.y.z.y:1234>
	"pubsubs": "",							// address where to reach the pubusb service, empty to disable pubsub functionality: <""|internal|x.y.z.y:1234>
	"users": "",							// address where to reach the user service, empty to disable user profile functionality: <""|internal|x.y.z.y:1234>
	"aliases": "",							// address where to reach the aliases service, empty to disable aliases functionality: <""|internal|x.y.z.y:1234>
},


"scheduler": {
	"enabled": false,						// start Scheduler service: <true|false>
},


"cdrs": {
	"enabled": false,						// start the CDR Server service:  <true|false>
	"extra_fields": [],						// extra fields to store in CDRs for non-generic CDRs
	"store_cdrs": true,						// store cdrs in storDb
	"rater": "internal",					// address where to reach the Rater for cost calculation, empty to disable functionality: <""|internal|x.y.z.y:1234>
	"pubsubs": "",							// address where to reach the pubusb service, empty to disable pubsub functionality: <""|internal|x.y.z.y:1234>
	"users": "",							// address where to reach the user service, empty to disable user profile functionality: <""|internal|x.y.z.y:1234>
	"aliases": "",							// address where to reach the aliases service, empty to disable aliases functionality: <""|internal|x.y.z.y:1234>
	"cdrstats": "",							// address where to reach the cdrstats service, empty to disable stats functionality<""|internal|x.y.z.y:1234>
	"cdr_replication":[],					// replicate the raw CDR to a number of servers
},


"cdrstats": {
	"enabled": false,						// starts the cdrstats service: <true|false>
	"save_interval": "1m",					// interval to save changed stats into dataDb storage
},


"cdre": {
	"*default": {
		"cdr_format": "csv",							// exported CDRs format <csv>
		"field_separator": ",",
		"data_usage_multiply_factor": 1,				// multiply data usage before export (eg: convert from KBytes to Bytes)
		"sms_usage_multiply_factor": 1,					// multiply data usage before export (eg: convert from SMS unit to call duration in some billing systems)
		"generic_usage_multiply_factor": 1,					// multiply data usage before export (eg: convert from GENERIC unit to call duration in some billing systems)
		"cost_multiply_factor": 1,						// multiply cost before export, eg: add VAT
		"cost_rounding_decimals": -1,					// rounding decimals for Cost values. -1 to disable rounding
		"cost_shift_digits": 0,							// shift digits in the cost on export (eg: convert from EUR to cents)
		"mask_destination_id": "MASKED_DESTINATIONS",	// destination id containing called addresses to be masked on export
		"mask_length": 0,								// length of the destination suffix to be masked
		"export_dir": "/var/log/cgrates/cdre",			// path where the exported CDRs will be placed
		"header_fields": [],							// template of the exported header fields
		"content_fields": [								// template of the exported content fields
			{"tag": "CgrId", "cdr_field_id": "CgrId", "type": "cdrfield", "value": "CgrId"},
			{"tag":"RunId", "cdr_field_id": "MediationRunId", "type": "cdrfield", "value": "MediationRunId"},
			{"tag":"Tor", "cdr_field_id": "TOR", "type": "cdrfield", "value": "TOR"},
			{"tag":"AccId", "cdr_field_id": "AccId", "type": "cdrfield", "value": "AccId"},
			{"tag":"ReqType", "cdr_field_id": "ReqType", "type": "cdrfield", "value": "ReqType"},
			{"tag":"Direction", "cdr_field_id": "Direction", "type": "cdrfield", "value": "Direction"},
			{"tag":"Tenant", "cdr_field_id": "Tenant", "type": "cdrfield", "value": "Tenant"},
			{"tag":"Category", "cdr_field_id": "Category", "type": "cdrfield", "value": "Category"},
			{"tag":"Account", "cdr_field_id": "Account", "type": "cdrfield", "value": "Account"},
			{"tag":"Subject", "cdr_field_id": "Subject", "type": "cdrfield", "value": "Subject"},
			{"tag":"Destination", "cdr_field_id": "Destination", "type": "cdrfield", "value": "Destination"},
			{"tag":"SetupTime", "cdr_field_id": "SetupTime", "type": "cdrfield", "value": "SetupTime", "layout": "2006-01-02T15:04:05Z07:00"},
			{"tag":"AnswerTime", "cdr_field_id": "AnswerTime", "type": "cdrfield", "value": "AnswerTime", "layout": "2006-01-02T15:04:05Z07:00"},
			{"tag":"Usage", "cdr_field_id": "Usage", "type": "cdrfield", "value": "Usage"},
			{"tag":"Cost", "cdr_field_id": "Cost", "type": "cdrfield", "value": "Cost"},
		],
		"trailer_fields": [],							// template of the exported trailer fields
	}
},


"cdrc": {
	"*default": {
		"enabled": false,							// enable CDR client functionality
		"dry_run": false,							// do not send the CDRs to CDRS, just parse them
		"cdrs": "internal",							// address where to reach CDR server. <internal|x.y.z.y:1234>
		"cdr_format": "csv",						// CDR file format <csv|freeswitch_csv|fwv|opensips_flatstore>
		"field_separator": ",",						// separator used in case of csv files
		"timezone": "",								// timezone for timestamps where not specified <""|UTC|Local|$IANA_TZ_DB>
		"run_delay": 0,								// sleep interval in seconds between consecutive runs, 0 to use automation via inotify
		"max_open_files": 1024,						// maximum simultaneous files to process, 0 for unlimited
		"data_usage_multiply_factor": 1024,			// conversion factor for data usage
		"cdr_in_dir": "/var/log/cgrates/cdrc/in",	// absolute path towards the directory where the CDRs are stored
		"cdr_out_dir": "/var/log/cgrates/cdrc/out",	// absolute path towards the directory where processed CDRs will be moved
		"failed_calls_prefix": "missed_calls",		// used in case of flatstore CDRs to avoid searching for BYE records
		"cdr_source_id": "freeswitch_csv",			// free form field, tag identifying the source of the CDRs within CDRS database
		"cdr_filter": "",							// filter CDR records to import
		"partial_record_cache": "10s",				// duration to cache partial records when not pairing
		"header_fields": [],						// template of the import header fields
		"content_fields":[							// import content_fields template, tag will match internally CDR field, in case of .csv value will be represented by index of the field value
			{"tag": "tor", "cdr_field_id": "TOR", "type": "cdrfield", "value": "2", "mandatory": true},
			{"tag": "accid", "cdr_field_id": "AccId", "type": "cdrfield", "value": "3", "mandatory": true},
			{"tag": "reqtype", "cdr_field_id": "ReqType", "type": "cdrfield", "value": "4", "mandatory": true},
			{"tag": "direction", "cdr_field_id": "Direction", "type": "cdrfield", "value": "5", "mandatory": true},
			{"tag": "tenant", "cdr_field_id": "Tenant", "type": "cdrfield", "value": "6", "mandatory": true},
			{"tag": "category", "cdr_field_id": "Category", "type": "cdrfield", "value": "7", "mandatory": true},
			{"tag": "account", "cdr_field_id": "Account", "type": "cdrfield", "value": "8", "mandatory": true},
			{"tag": "subject", "cdr_field_id": "Subject", "type": "cdrfield", "value": "9", "mandatory": true},
			{"tag": "destination", "cdr_field_id": "Destination", "type": "cdrfield", "value": "10", "mandatory": true},
			{"tag": "setup_time", "cdr_field_id": "SetupTime", "type": "cdrfield", "value": "11", "mandatory": true},
			{"tag": "answer_time", "cdr_field_id": "AnswerTime", "type": "cdrfield", "value": "12", "mandatory": true},
			{"tag": "usage", "cdr_field_id": "Usage", "type": "cdrfield", "value": "13", "mandatory": true},
		],
		"trailer_fields": [],							// template of the import trailer fields
	}
},

"sm_generic": {
	"enabled": false,						// starts SessionManager service: <true|false>
	"rater": "internal",					// address where to reach the Rater <""|internal|127.0.0.1:2013>
	"cdrs": "internal",						// address where to reach CDR Server <""|internal|x.y.z.y:1234>
	"debit_interval": "10s",				// interval to perform debits on.
	"min_call_duration": "0s",				// only authorize calls with allowed duration higher than this
	"max_call_duration": "3h",				// maximum call duration a prepaid call can last
},


"sm_freeswitch": {
	"enabled": false,				// starts SessionManager service: <true|false>
	"rater": "internal",			// address where to reach the Rater <""|internal|127.0.0.1:2013>
	"cdrs": "internal",				// address where to reach CDR Server, empty to disable CDR capturing <""|internal|x.y.z.y:1234>
	"create_cdr": false,			// create CDR out of events and sends them to CDRS component
	"extra_fields": [],				// extra fields to store in auth/CDRs when creating them
	"debit_interval": "10s",		// interval to perform debits on.
	"min_call_duration": "0s",		// only authorize calls with allowed duration higher than this
	"max_call_duration": "3h",		// maximum call duration a prepaid call can last
	"min_dur_low_balance": "5s",	// threshold which will trigger low balance warnings for prepaid calls (needs to be lower than debit_interval)
	"low_balance_ann_file": "",		// file to be played when low balance is reached for prepaid calls
	"empty_balance_context": "",	// if defined, prepaid calls will be transfered to this context on empty balance
	"empty_balance_ann_file": "",	// file to be played before disconnecting prepaid calls on empty balance (applies only if no context defined)
	"subscribe_park": true,			// subscribe via fsock to receive park events
	"channel_sync_interval": "5m",	// sync channels with freeswitch regularly
	"connections":[					// instantiate connections to multiple FreeSWITCH servers
		{"server": "127.0.0.1:8021", "password": "ClueCon", "reconnects": 5}
	],
},


"sm_kamailio": {
	"enabled": false,				// starts SessionManager service: <true|false>
	"rater": "internal",			// address where to reach the Rater <""|internal|127.0.0.1:2013>
	"cdrs": "internal",				// address where to reach CDR Server, empty to disable CDR capturing <""|internal|x.y.z.y:1234>
	"create_cdr": false,			// create CDR out of events and sends them to CDRS component
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
	"cdrs": "internal",					// address where to reach CDR Server, empty to disable CDR capturing <""|internal|x.y.z.y:1234>
	"reconnects": 5,					// number of reconnects if connection is lost
	"create_cdr": false,				// create CDR out of events and sends them to CDRS component
	"debit_interval": "10s",			// interval to perform debits on.
	"min_call_duration": "0s",			// only authorize calls with allowed duration higher than this
	"max_call_duration": "3h",			// maximum call duration a prepaid call can last
	"events_subscribe_interval": "60s",	// automatic events subscription to OpenSIPS, 0 to disable it
	"mi_addr": "127.0.0.1:8020",		// address where to reach OpenSIPS MI to send session disconnects
},


"historys": {
	"enabled": false,							// starts History service: <true|false>.
	"history_dir": "/var/log/cgrates/history",	// location on disk where to store history files.
	"save_interval": "1s",						// interval to save changed cache into .git archive
},


"pubsubs": {
	"enabled": false,							// starts PubSub service: <true|false>.
},


"aliases": {
	"enabled": false,							// starts Aliases service: <true|false>.
},


"users": {
	"enabled": false,							// starts User service: <true|false>.
	"indexes": [],								// user profile field indexes
},


"mailer": {
	"server": "localhost",								// the server to use when sending emails out
	"auth_user": "cgrates",								// authenticate to email server using this user
	"auth_passwd": "CGRateS.org",						// authenticate to email server with this password
	"from_address": "cgr-mailer@localhost.localdomain"	// from address used when sending emails out
},


"suretax": {
	"url": "",								// API url
	"client_number": "",					// client number, provided by SureTax
	"validation_key": "",					// validation key provided by SureTax
	"timezone": "Local",					// convert the time of the events to this timezone before sending request out <UTC|Local|$IANA_TZ_DB>
	"include_local_cost": false,			// sum local calculated cost with tax one in final cost
	"return_file_code": "0",				// default or Quote purposes <0|Q>
	"response_group": "03",					// determines how taxes are grouped for the response <03|13>
	"response_type": "D4",					// determines the granularity of taxes and (optionally) the decimal precision for the tax calculations and amounts in the response
	"regulatory_code": "03",				// provider type
	"client_tracking": "CgrId",				// template extracting client information out of StoredCdr; <$RSRFields>
	"customer_number": "Subject",			// template extracting customer number out of StoredCdr; <$RSRFields>
	"orig_number":  "Subject", 				// template extracting origination number out of StoredCdr; <$RSRFields>
	"term_number": "Destination",			// template extracting termination number out of StoredCdr; <$RSRFields>
	"bill_to_number": "",					// template extracting billed to number out of StoredCdr; <$RSRFields>
	"zipcode": "",							// template extracting billing zip code out of StoredCdr; <$RSRFields>
	"plus4": "",							// template extracting billing zip code extension out of StoredCdr; <$RSRFields>
	"p2pzipcode": "",						// template extracting secondary zip code out of StoredCdr; <$RSRFields>
	"p2pplus4": "",							// template extracting secondary zip code extension out of StoredCdr; <$RSRFields>
	"units": "^1",							// template extracting number of “lines” or unique charges contained within the revenue out of StoredCdr; <$RSRFields>
	"unit_type": "^00",						// template extracting number of unique access lines out of StoredCdr; <$RSRFields>
	"tax_included": "^0",					// template extracting tax included in revenue out of StoredCdr; <$RSRFields>
	"tax_situs_rule": "^04",				// template extracting tax situs rule out of StoredCdr; <$RSRFields>
	"trans_type_code": "^010101",			// template extracting transaction type indicator out of StoredCdr; <$RSRFields>
	"sales_type_code": "^R",				// template extracting sales type code out of StoredCdr; <$RSRFields>
	"tax_exemption_code_list": "",			// template extracting tax exemption code list out of StoredCdr; <$RSRFields>
},

}`
