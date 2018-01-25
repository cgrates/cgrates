/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
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

package utils

var (
	CDRExportFormats = []string{DRYRUN, MetaFileCSV, MetaFileFWV, MetaHTTPjsonCDR, MetaHTTPjsonMap, MetaHTTPjson, META_HTTP_POST, MetaAMQPjsonCDR, MetaAMQPjsonMap}
	PrimaryCdrFields = []string{CGRID, Source, OriginHost, OriginID, TOR, RequestType, Direction, Tenant, Category, Account, Subject, Destination, SetupTime, PDD, AnswerTime, Usage,
		SUPPLIER, DISCONNECT_CAUSE, COST, RATED, PartialField, MEDI_RUNID}
	GitLastLog                  string // If set, it will be processed as part of versioning
	PosterTransportContentTypes = map[string]string{
		MetaHTTPjsonCDR: CONTENT_JSON,
		MetaHTTPjsonMap: CONTENT_JSON,
		MetaHTTPjson:    CONTENT_JSON,
		META_HTTP_POST:  CONTENT_FORM,
		MetaAMQPjsonCDR: CONTENT_JSON,
		MetaAMQPjsonMap: CONTENT_JSON,
	}
	CDREFileSuffixes = map[string]string{
		MetaHTTPjsonCDR: JSNSuffix,
		MetaHTTPjsonMap: JSNSuffix,
		MetaAMQPjsonCDR: JSNSuffix,
		MetaAMQPjsonMap: JSNSuffix,
		META_HTTP_POST:  FormSuffix,
		MetaFileCSV:     CSVSuffix,
		MetaFileFWV:     FWVSuffix,
	}
	CacheInstanceToPrefix = map[string]string{
		CacheDestinations:              DESTINATION_PREFIX,
		CacheReverseDestinations:       REVERSE_DESTINATION_PREFIX,
		CacheRatingPlans:               RATING_PLAN_PREFIX,
		CacheRatingProfiles:            RATING_PROFILE_PREFIX,
		CacheLCRRules:                  LCR_PREFIX,
		CacheCDRStatS:                  CDR_STATS_PREFIX,
		CacheActions:                   ACTION_PREFIX,
		CacheActionPlans:               ACTION_PLAN_PREFIX,
		CacheAccountActionPlans:        AccountActionPlansPrefix,
		CacheActionTriggers:            ACTION_TRIGGER_PREFIX,
		CacheSharedGroups:              SHARED_GROUP_PREFIX,
		CacheAliases:                   ALIASES_PREFIX,
		CacheReverseAliases:            REVERSE_ALIASES_PREFIX,
		CacheDerivedChargers:           DERIVEDCHARGERS_PREFIX,
		CacheResourceProfiles:          ResourceProfilesPrefix,
		CacheResources:                 ResourcesPrefix,
		CacheEventResources:            EventResourcesPrefix,
		CacheTimings:                   TimingsPrefix,
		CacheStatQueueProfiles:         StatQueueProfilePrefix,
		CacheStatQueues:                StatQueuePrefix,
		CacheThresholdProfiles:         ThresholdProfilePrefix,
		CacheThresholds:                ThresholdPrefix,
		CacheFilters:                   FilterPrefix,
		CacheSupplierProfiles:          SupplierProfilePrefix,
		CacheAttributeProfiles:         AttributeProfilePrefix,
		CacheResourceFilterIndexes:     ResourceFilterIndexes,
		CacheResourceFilterRevIndexes:  ResourceFilterRevIndexes,
		CacheStatFilterIndexes:         StatFilterIndexes,
		CacheStatFilterRevIndexes:      StatFilterRevIndexes,
		CacheThresholdFilterIndexes:    ThresholdFilterIndexes,
		CacheThresholdFilterRevIndexes: ThresholdFilterRevIndexes,
		CacheSupplierFilterIndexes:     SupplierFilterIndexes,
		CacheSupplierFilterRevIndexes:  SupplierFilterRevIndexes,
		CacheAttributeFilterIndexes:    AttributeFilterIndexes,
		CacheAttributeFilterRevIndexes: AttributeFilterRevIndexes,
	}
	CachePrefixToInstance map[string]string // will be built on init
)

const (
	CGRateS                       = "CGRateS"
	VERSION                       = "0.9.1~rc8"
	GitLastLogFileName            = ".git_lastlog.txt"
	DIAMETER_FIRMWARE_REVISION    = 918
	REDIS_MAX_CONNS               = 10
	CGRATES                       = "cgrates"
	POSTGRES                      = "postgres"
	MYSQL                         = "mysql"
	MONGO                         = "mongo"
	DataManager                   = "DataManager"
	REDIS                         = "redis"
	MAPSTOR                       = "mapstor"
	LOCALHOST                     = "127.0.0.1"
	FSCDR_FILE_CSV                = "freeswitch_file_csv"
	FSCDR_HTTP_JSON               = "freeswitch_http_json"
	NOT_IMPLEMENTED               = "not implemented"
	PREPAID                       = "prepaid"
	META_PREPAID                  = "*prepaid"
	POSTPAID                      = "postpaid"
	META_POSTPAID                 = "*postpaid"
	PSEUDOPREPAID                 = "pseudoprepaid"
	META_PSEUDOPREPAID            = "*pseudoprepaid"
	META_RATED                    = "*rated"
	META_NONE                     = "*none"
	META_NOW                      = "*now"
	ROUNDING_UP                   = "*up"
	ROUNDING_MIDDLE               = "*middle"
	ROUNDING_DOWN                 = "*down"
	ANY                           = "*any"
	UNLIMITED                     = "*unlimited"
	ZERO                          = "*zero"
	ASAP                          = "*asap"
	USERS                         = "*users"
	COMMENT_CHAR                  = '#'
	CSV_SEP                       = ','
	FALLBACK_SEP                  = ';'
	INFIELD_SEP                   = ";"
	MetaPipe                      = "*|"
	FIELDS_SEP                    = ","
	InInFieldSep                  = ":"
	STATIC_HDRVAL_SEP             = "::"
	REGEXP_PREFIX                 = "~"
	FILTER_VAL_START              = "("
	FILTER_VAL_END                = ")"
	JSON                          = "json"
	GOB                           = "gob"
	MSGPACK                       = "msgpack"
	CSV_LOAD                      = "CSVLOAD"
	CGRID                         = "CGRID"
	TOR                           = "ToR"
	ORDERID                       = "OrderID"
	OriginID                      = "OriginID"
	InitialOriginID               = "InitialOriginID"
	OriginIDPrefix                = "OriginIDPrefix"
	Source                        = "Source"
	OriginHost                    = "OriginHost"
	RequestType                   = "RequestType"
	Direction                     = "Direction"
	Tenant                        = "Tenant"
	Category                      = "Category"
	Context                       = "Context"
	Account                       = "Account"
	Subject                       = "Subject"
	Destination                   = "Destination"
	SetupTime                     = "SetupTime"
	AnswerTime                    = "AnswerTime"
	Usage                         = "Usage"
	LastUsed                      = "LastUsed"
	PDD                           = "PDD"
	SUPPLIER                      = "Supplier"
	MEDI_RUNID                    = "RunID"
	COST                          = "Cost"
	COST_DETAILS                  = "CostDetails"
	RATED                         = "rated"
	RATED_FLD                     = "Rated"
	PartialField                  = "Partial"
	DEFAULT_RUNID                 = "*default"
	META_DEFAULT                  = "*default"
	STATIC_VALUE_PREFIX           = "^"
	CSV                           = "csv"
	FWV                           = "fwv"
	PartialCSV                    = "partial_csv"
	DRYRUN                        = "dry_run"
	META_COMBIMED                 = "*combimed"
	MetaInternal                  = "*internal"
	MetaInline                    = "*inline"
	ZERO_RATING_SUBJECT_PREFIX    = "*zero"
	OK                            = "OK"
	CDRE_FIXED_WIDTH              = "fwv"
	XML_PROFILE_PREFIX            = "*xml:"
	CDRE                          = "cdre"
	CDRC                          = "cdrc"
	MASK_CHAR                     = "*"
	CONCATENATED_KEY_SEP          = ":"
	FORKED_CDR                    = "forked_cdr"
	UNIT_TEST                     = "UNIT_TEST"
	HDR_VAL_SEP                   = "/"
	MONETARY                      = "*monetary"
	SMS                           = "*sms"
	MMS                           = "*mms"
	GENERIC                       = "*generic"
	DATA                          = "*data"
	VOICE                         = "*voice"
	MAX_COST_FREE                 = "*free"
	MAX_COST_DISCONNECT           = "*disconnect"
	HOURS                         = "hours"
	MINUTES                       = "minutes"
	NANOSECONDS                   = "nanoseconds"
	SECONDS                       = "seconds"
	OUT                           = "*out"
	IN                            = "*in"
	META_OUT                      = "*out"
	META_ANY                      = "*any"
	MetaExists                    = "*exists"
	CDR_IMPORT                    = "cdr_import"
	CDR_EXPORT                    = "cdr_export"
	ASR                           = "ASR"
	ACD                           = "ACD"
	FILTER_REGEXP_TPL             = "$1$2$3$4$5"
	TASKS_KEY                     = "tasks"
	ACTION_PLAN_PREFIX            = "apl_"
	AccountActionPlansPrefix      = "aap_"
	ACTION_TRIGGER_PREFIX         = "atr_"
	REVERSE_ACTION_TRIGGER_PREFIX = "rtr_"
	RATING_PLAN_PREFIX            = "rpl_"
	RATING_PROFILE_PREFIX         = "rpf_"
	ACTION_PREFIX                 = "act_"
	SHARED_GROUP_PREFIX           = "shg_"
	ACCOUNT_PREFIX                = "acc_"
	DESTINATION_PREFIX            = "dst_"
	REVERSE_DESTINATION_PREFIX    = "rds_"
	LCR_PREFIX                    = "lcr_"
	DERIVEDCHARGERS_PREFIX        = "dcs_"
	CDR_STATS_QUEUE_PREFIX        = "csq_"
	PUBSUB_SUBSCRIBERS_PREFIX     = "pss_"
	USERS_PREFIX                  = "usr_"
	ALIASES_PREFIX                = "als_"
	REVERSE_ALIASES_PREFIX        = "rls_"
	ResourcesPrefix               = "res_"
	ResourceProfilesPrefix        = "rsp_"
	ThresholdPrefix               = "thd_"
	TimingsPrefix                 = "tmg_"
	FilterPrefix                  = "ftr_"
	FilterIndex                   = "fti_"
	CDR_STATS_PREFIX              = "cst_"
	TEMP_DESTINATION_PREFIX       = "tmp_"
	LOG_CALL_COST_PREFIX          = "cco_"
	LOG_ACTION_TIMMING_PREFIX     = "ltm_"
	LOG_ACTION_TRIGGER_PREFIX     = "ltr_"
	VERSION_PREFIX                = "ver_"
	LOG_ERR                       = "ler_"
	LOG_CDR                       = "cdr_"
	LOG_MEDIATED_CDR              = "mcd_"
	StatQueueProfilePrefix        = "sqp_"
	SupplierProfilePrefix         = "spp_"
	AttributeProfilePrefix        = "alp_"
	ThresholdProfilePrefix        = "thp_"
	StatQueuePrefix               = "stq_"
	LOADINST_KEY                  = "load_history"
	SESSION_MANAGER_SOURCE        = "SMR"
	MEDIATOR_SOURCE               = "MED"
	CDRS_SOURCE                   = "CDRS"
	SCHED_SOURCE                  = "SCH"
	RATER_SOURCE                  = "RAT"
	CREATE_CDRS_TABLES_SQL        = "create_cdrs_tables.sql"
	CREATE_TARIFFPLAN_TABLES_SQL  = "create_tariffplan_tables.sql"
	TEST_SQL                      = "TEST_SQL"
	DESTINATIONS_LOAD_THRESHOLD   = 0.1
	META_CONSTANT                 = "*constant"
	META_FILLER                   = "*filler"
	META_HANDLER                  = "*handler"
	META_HTTP_POST                = "*http_post"
	MetaHTTPjson                  = "*http_json"
	MetaHTTPjsonCDR               = "*http_json_cdr"
	META_HTTP_JSONRPC             = "*http_jsonrpc"
	MetaHTTPjsonMap               = "*http_json_map"
	MetaAMQPjsonCDR               = "*amqp_json_cdr"
	MetaAMQPjsonMap               = "*amqp_json_map"
	NANO_MULTIPLIER               = 1000000000
	CGR_AUTHORIZE                 = "CGR_AUTHORIZE"
	CONFIG_DIR                    = "/etc/cgrates/"
	CGR_ACCOUNT                   = "cgr_account"
	CGR_SUPPLIER                  = "cgr_supplier"
	CGR_DESTINATION               = "cgr_destination"
	CGR_SUBJECT                   = "cgr_subject"
	CGR_CATEGORY                  = "cgr_category"
	CGR_REQTYPE                   = "cgr_reqtype"
	CGR_TENANT                    = "cgr_tenant"
	CGR_TOR                       = "cgr_tor"
	CGR_OriginID                  = "cgr_originid"
	CGR_HOST                      = "cgr_host"
	CGR_PDD                       = "cgr_pdd"
	DISCONNECT_CAUSE              = "DisconnectCause"
	CGR_DISCONNECT_CAUSE          = "cgr_disconnectcause"
	CGR_COMPUTELCR                = "cgr_computelcr"
	CGR_SUPPLIERS                 = "cgr_suppliers"
	CGRFlags                      = "cgr_flags"
	KAM_FLATSTORE                 = "kamailio_flatstore"
	OSIPS_FLATSTORE               = "opensips_flatstore"
	MAX_DEBIT_CACHE_PREFIX        = "MAX_DEBIT_"
	REFUND_INCR_CACHE_PREFIX      = "REFUND_INCR_"
	REFUND_ROUND_CACHE_PREFIX     = "REFUND_ROUND_"
	GET_SESS_RUNS_CACHE_PREFIX    = "GET_SESS_RUNS_"
	GET_DERIV_MAX_SESS_TIME       = "GET_DERIV_MAX_SESS_TIME_"
	LOG_CALL_COST_CACHE_PREFIX    = "LOG_CALL_COSTS_"
	LCRCachePrefix                = "LCR_"
	MetaRating                    = "*rating"
	NOT_AVAILABLE                 = "N/A"
	MetaEmpty                     = "*empty"
	CALL                          = "call"
	EXTRA_FIELDS                  = "ExtraFields"
	META_SURETAX                  = "*sure_tax"
	MetaDynamic                   = "*dynamic"
	SURETAX                       = "suretax"
	DIAMETER_AGENT                = "diameter_agent"
	COUNTER_EVENT                 = "*event"
	COUNTER_BALANCE               = "*balance"
	EVENT_NAME                    = "EventName"
	COMPUTE_LCR                   = "ComputeLcr"
	CGR_AUTHORIZATION             = "CgrAuthorization"
	CGR_SESSION_START             = "CgrSessionStart"
	CGR_SESSION_UPDATE            = "CgrSessionUpdate"
	CGR_SESSION_END               = "CgrSessionEnd"
	CGR_LCR_REQUEST               = "CgrLcrRequest"
	// action trigger threshold types
	TRIGGER_MIN_EVENT_COUNTER    = "*min_event_counter"
	TRIGGER_MIN_BALANCE_COUNTER  = "*min_balance_counter"
	TRIGGER_MAX_EVENT_COUNTER    = "*max_event_counter"
	TRIGGER_MAX_BALANCE_COUNTER  = "*max_balance_counter"
	TRIGGER_MIN_BALANCE          = "*min_balance"
	TRIGGER_MAX_BALANCE          = "*max_balance"
	TRIGGER_BALANCE_EXPIRED      = "*balance_expired"
	HIERARCHY_SEP                = ">"
	META_COMPOSED                = "*composed"
	NegativePrefix               = "!"
	MatchStartPrefix             = "^"
	MatchEndPrefix               = "$"
	SMG                          = "SMG"
	MetaGrouped                  = "*grouped"
	MetaRaw                      = "*raw"
	CreatedAt                    = "CreatedAt"
	UpdatedAt                    = "UpdatedAt"
	HandlerArgSep                = "|"
	FlagForceDuration            = "fd"
	NodeID                       = "NodeID"
	ActiveGoroutines             = "ActiveGoroutines"
	SessionTTL                   = "SessionTTL"
	SessionTTLMaxDelay           = "SessionTTLMaxDelay"
	SessionTTLLastUsed           = "SessionTTLLastUsed"
	SessionTTLUsage              = "SessionTTLUsage"
	HandlerSubstractUsage        = "*substract_usage"
	XML                          = "xml"
	MetaGOBrpc                   = "*gob"
	MetaJSONrpc                  = "*json"
	MetaDateTime                 = "*datetime"
	MetaMaskedDestination        = "*masked_destination"
	MetaUnixTimestamp            = "*unix_timestamp"
	MetaPostCDR                  = "*post_cdr"
	MetaDumpToFile               = "*dump_to_file"
	NonTransactional             = "" // used in transactional cache mechanism
	EVT_ACCOUNT_BALANCE_MODIFIED = "ACCOUNT_BALANCE_MODIFIED"
	EVT_ACTION_TRIGGER_FIRED     = "ACTION_TRIGGER_FIRED"
	EVT_ACTION_TIMING_FIRED      = "ACTION_TRIGGER_FIRED"
	SMAsterisk                   = "sm_asterisk"
	DataDB                       = "data_db"
	StorDB                       = "stor_db"
	Cache                        = "cache"
	NotFoundCaps                 = "NOT_FOUND"
	ItemNotFound                 = "item not found"
	ItemNotCloneable             = "item not cloneable"
	NotCloneableCaps             = "NOT_CLONEABLE"
	ServerErrorCaps              = "SERVER_ERROR"
	MandatoryIEMissingCaps       = "MANDATORY_IE_MISSING"
	AttributesNotFoundCaps       = "ATTRIBUTES_NOT_FOUND"
	AttributesNotFound           = "attributes not found"
	UnsupportedCachePrefix       = "unsupported cache prefix"
	CDRSCtx                      = "cdrs"
	MandatoryInfoMissing         = "mandatory information missing"
	UnsupportedServiceIDCaps     = "UNSUPPORTED_SERVICE_ID"
	ServiceManager               = "service_manager"
	ServiceAlreadyRunning        = "service already running"
	ServiceNotRunning            = "service not running"
	RunningCaps                  = "RUNNING"
	StoppedCaps                  = "STOPPED"
	SchedulerNotRunningCaps      = "SCHEDULLER_NOT_RUNNING"
	MetaScheduler                = "*scheduler"
	MetaCostDetails              = "*cost_details"
	MetaAccounts                 = "*accounts"
	MetaActionPlans              = "*action_plans"
	MetaActionTriggers           = "*action_triggers"
	MetaActions                  = "*actions"
	MetaSharedGroups             = "*shared_groups"
	MetaStats                    = "*stats"
	MetaThresholds               = "*thresholds"
	MetaSuppliers                = "*suppliers"
	MetaAttributes               = "*attributes"
	Migrator                     = "migrator"
	UnsupportedMigrationTask     = "unsupported migration task"
	NoStorDBConnection           = "not connected to StorDB"
	UndefinedVersion             = "undefined version"
	UnsupportedDB                = "unsupported database"
	ACCOUNT_SUMMARY              = "AccountSummary"
	TxtSuffix                    = ".txt"
	JSNSuffix                    = ".json"
	FormSuffix                   = ".form"
	CSVSuffix                    = ".csv"
	FWVSuffix                    = ".fwv"
	CONTENT_JSON                 = "json"
	CONTENT_FORM                 = "form"
	CONTENT_TEXT                 = "text"
	FileLockPrefix               = "file_"
	ActionsPoster                = "act"
	CDRPoster                    = "cdr"
	MetaFileCSV                  = "*file_csv"
	MetaFileFWV                  = "*file_fwv"
	Accounts                     = "Accounts"
	AccountService               = "AccountS"
	Actions                      = "Actions"
	ActionPlans                  = "ActionPlans"
	ActionTriggers               = "ActionTriggers"
	SharedGroups                 = "SharedGroups"
	MetaEveryMinute              = "*every_minute"
	MetaHourly                   = "*hourly"
	ID                           = "ID"
	Thresholds                   = "Thresholds"
	Suppliers                    = "Suppliers"
	StatS                        = "stats"
	RALService                   = "RALs"
	CostSource                   = "CostSource"
	ExtraInfo                    = "ExtraInfo"
	MetaPrefix                   = "*"
	EventResourcesPrefix         = "ers_"
	MetaSysLog                   = "*syslog"
	MetaStdLog                   = "*stdout"
	MetaNever                    = "*never"
	EventType                    = "EventType"
	EventSource                  = "EventSource"
	AccountID                    = "AccountID"
	ResourceID                   = "ResourceID"
	TotalUsage                   = "TotalUsage"
	StatID                       = "StatID"
	BalanceType                  = "BalanceType"
	BalanceID                    = "BalanceID"
	Units                        = "Units"
	AccountUpdate                = "AccountUpdate"
	BalanceUpdate                = "BalanceUpdate"
	StatUpdate                   = "StatUpdate"
	ResourceUpdate               = "ResourceUpdate"
	CDR                          = "CDR"
	CDRs                         = "CDRs"
	ExpiryTime                   = "ExpiryTime"
	AllowNegative                = "AllowNegative"
	Disabled                     = "Disabled"
	Action                       = "Action"
	MetaNow                      = "*now"
	TpRatingPlans                = "TpRatingPlans"
	TpFilters                    = "TpFilters"
	TpDestinationRates           = "TpDestinationRates"
	TpActionTriggers             = "TpActionTriggers"
	TpAccountActionsV            = "TpAccountActions"
	TpActionPlans                = "TpActionPlans"
	TpActions                    = "TpActions"
	TpDerivedCharges             = "TpDerivedCharges"
	TpThresholds                 = "TpThresholds"
	TpSuppliers                  = "TpSuppliers"
	TpStats                      = "TpStats"
	TpSharedGroups               = "TpSharedGroups"
	TpRatingProfiles             = "TpRatingProfiles"
	TpResources                  = "TpResources"
	TpRates                      = "TpRates"
	TpTiming                     = "TpTiming"
	TpResource                   = "TpResource"
	TpAliases                    = "TpAliases"
	TpUsers                      = "TpUsers"
	TpDerivedChargersV           = "TpDerivedChargers"
	TpCdrStats                   = "TpCdrStats"
	TpDestinations               = "TpDestinations"
	TpRatingPlan                 = "TpRatingPlan"
	TpRatingProfile              = "TpRatingProfile"
	Timing                       = "Timing"
	RQF                          = "RQF"
	Resource                     = "Resource"
	ReverseAlias                 = "ReverseAlias"
	Alias                        = "Alias"
	User                         = "User"
	Subscribers                  = "Subscribers"
	DerivedChargersV             = "DerivedChargers"
	CdrStats                     = "CdrStats"
	Destinations                 = "Destinations"
	ReverseDestinations          = "ReverseDestinations"
	LCR                          = "LCR"
	RatingPlan                   = "RatingPlan"
	RatingProfile                = "RatingProfile"
	MetaRatingPlans              = "*ratingplans"
	MetaRatingProfile            = "*ratingprofile"
	MetaDestinations             = "*destinations"
	MetaReverseDestinations      = "*reversedestinations"
	MetaLCR                      = "*lcr"
	MetaCdrStats                 = "*cdrstats"
	MetaTiming                   = "*Timing"
	MetaRQF                      = "*RQF"
	MetaResource                 = "*Resource"
	MetaReverseAlias             = "*ReverseAlias"
	MetaAlias                    = "*Alias"
	MetaUser                     = "*User"
	MetaSubscribers              = "*Subscribers"
	MetaDerivedChargersV         = "*DerivedChargers"
	MetaStorDB                   = "*stordb"
	MetaDataDB                   = "*datadb"
	MetaWeight                   = "*weight"
	MetaLeastCost                = "*least_cost"
	Weight                       = "Weight"
	Cost                         = "Cost"
	RatingPlanID                 = "RatingPlanID"
	MetaSessionS                 = "*sessions"
	FreeSWITCHAgent              = "FreeSWITCHAgent"
	MetaDefault                  = "*default"
)

//MetaMetrics
const (
	MetaASR     = "*asr"
	MetaACD     = "*acd"
	MetaTCD     = "*tcd"
	MetaACC     = "*acc"
	MetaTCC     = "*tcc"
	MetaPDD     = "*pdd"
	MetaDDC     = "*ddc"
	MetaSum     = "*sum"
	MetaAverage = "*average"
)

//Services
const (
	SessionS    = "SessionS"
	AttributeS  = "AttributeS"
	SupplierS   = "SupplierS"
	ResourceS   = "ResourceS"
	StatService = "StatS"
	FilterS     = "FilterS"
)

//Migrator Metas
const (
	MetaSetVersions        = "*set_versions"
	MetaTpRatingPlans      = "*tp_rating_plans"
	MetaTpFilters          = "*tp_filters"
	MetaTpDestinationRates = "*tp_destination_rates"
	MetaTpActionTriggers   = "*tp_action_triggers"
	MetaTpAccountActions   = "*tp_account_actions"
	MetaTpActionPlans      = "*tp_action_plans"
	MetaTpActions          = "*tp_actions"
	MetaTpDerivedChargers  = "*tp_derived_charges"
	MetaTpThresholds       = "*tp_thresholds"
	MetaTpSuppliers        = "*tp_suppliers"
	MetaTpStats            = "*tp_stats"
	MetaTpSharedGroups     = "*tp_shared_groups"
	MetaTpRatingProfiles   = "*tp_rating_profiles"
	MetaTpResources        = "*tp_resources"
	MetaTpRates            = "*tp_rates"
	MetaTpTiming           = "*tp_timing"
	MetaTpResource         = "*tp_resource"
	MetaTpAliases          = "*tp_aliases"
	MetaTpUsers            = "*tp_users"
	MetaTpDerivedChargersV = "*tp_derived_chargers"
	MetaTpCdrStats         = "*tp_cdrstats"
	MetaTpDestinations     = "*tp_destinations"
	MetaTpRatingPlan       = "*tp_rating_plan"
	MetaTpRatingProfile    = "*tp_rating_profile"
)

// MetaFilterIndexesAPIs
const (
	ApierV1ComputeFilterIndexes = "ApierV1.ComputeFilterIndexes"
)

const (
	ApierV2LoadTariffPlanFromFolder = "ApierV2.LoadTariffPlanFromFolder"
)

// MetaSupplierAPIs
const (
	SupplierSv1GetSuppliers = "SupplierSv1.GetSuppliers"
)

// AttributeS APIs
const (
	AttributeSv1GetAttributeForEvent = "AttributeSv1.GetAttributeForEvent"
	AttributeSv1ProcessEvent         = "AttributeSv1.ProcessEvent"
)

//ThresholdS APIs
const (
	ThresholdSv1ProcessEvent    = "ThresholdSv1.ProcessEvent"
	ThresholdSv1GetThreshold    = "ThresholdSv1.GetThreshold"
	ThresholdSv1GetThresholdIDs = "ThresholdSv1.GetThresholdIDs"
)

//StatS APIs
const (
	StatSv1ProcessEvent          = "StatSv1.ProcessEvent"
	StatSv1GetQueueIDs           = "StatSv1.GetQueueIDs"
	StatSv1GetQueueStringMetrics = "StatSv1.GetQueueStringMetrics"
)

//ResourceS APIs
const (
	ResourceSv1AuthorizeResources   = "ResourceSv1.AuthorizeResources"
	ResourceSv1GetResourcesForEvent = "ResourceSv1.GetResourcesForEvent"
	ResourceSv1AllocateResources    = "ResourceSv1.AllocateResources"
	ResourceSv1ReleaseResources     = "ResourceSv1.ReleaseResources"
)

//SessionS APIs
const (
	SessionSv1AuthorizeEvent           = "SessionSv1.AuthorizeEvent"
	SessionSv1AuthorizeEventWithDigest = "SessionSv1.AuthorizeEventWithDigest"
	SessionSv1InitiateSession          = "SessionSv1.InitiateSession"
	SessionSv1UpdateSession            = "SessionSv1.UpdateSession"
	SessionSv1TerminateSession         = "SessionSv1.TerminateSession"
	SessionSv1ProcessCDR               = "SessionSv1.ProcessCDR"
	SessionSv1ProcessEvent             = "SessionSv1.ProcessEvent"
	SessionSv1DisconnectSession        = "SessionSv1.DisconnectSession"
	SessionSv1GetActiveSessions        = "SessionSv1.GetActiveSessions"
	SessionSv1GetPassiveSessions       = "SessionSv1.GetPassiveSessions"
	SMGenericV1InitiateSession         = "SMGenericV1.InitiateSession"
	SMGenericV2InitiateSession         = "SMGenericV2.InitiateSession"
	SMGenericV2UpdateSession           = "SMGenericV2.UpdateSession"
)

//CSV file name
const (
	TIMINGS_CSV           = "Timings.csv"
	DESTINATIONS_CSV      = "Destinations.csv"
	RATES_CSV             = "Rates.csv"
	DESTINATION_RATES_CSV = "DestinationRates.csv"
	RATING_PLANS_CSV      = "RatingPlans.csv"
	RATING_PROFILES_CSV   = "RatingProfiles.csv"
	SHARED_GROUPS_CSV     = "SharedGroups.csv"
	LCRS_CSV              = "LcrRules.csv"
	ACTIONS_CSV           = "Actions.csv"
	ACTION_PLANS_CSV      = "ActionPlans.csv"
	ACTION_TRIGGERS_CSV   = "ActionTriggers.csv"
	ACCOUNT_ACTIONS_CSV   = "AccountActions.csv"
	DERIVED_CHARGERS_CSV  = "DerivedChargers.csv"
	CDR_STATS_CSV         = "CdrStats.csv"
	USERS_CSV             = "Users.csv"
	ALIASES_CSV           = "Aliases.csv"
	ResourcesCsv          = "Resources.csv"
	StatsCsv              = "Stats.csv"
	ThresholdsCsv         = "Thresholds.csv"
	FiltersCsv            = "Filters.csv"
	SuppliersCsv          = "Suppliers.csv"
	AttributesCsv         = "Attributes.csv"
)

//Table Name
const (
	TBLTPTimings          = "tp_timings"
	TBLTPDestinations     = "tp_destinations"
	TBLTPRates            = "tp_rates"
	TBLTPDestinationRates = "tp_destination_rates"
	TBLTPRatingPlans      = "tp_rating_plans"
	TBLTPRateProfiles     = "tp_rating_profiles"
	TBLTPSharedGroups     = "tp_shared_groups"
	TBLTPCdrStats         = "tp_cdr_stats"
	TBLTPLcrs             = "tp_lcr_rules"
	TBLTPActions          = "tp_actions"
	TBLTPActionPlans      = "tp_action_plans"
	TBLTPActionTriggers   = "tp_action_triggers"
	TBLTPAccountActions   = "tp_account_actions"
	TBLTPDerivedChargers  = "tp_derived_chargers"
	TBLTPUsers            = "tp_users"
	TBLTPAliases          = "tp_aliases"
	TBLTPResources        = "tp_resources"
	TBLTPStats            = "tp_stats"
	TBLTPThresholds       = "tp_thresholds"
	TBLTPFilters          = "tp_filters"
	SMCostsTBL            = "sm_costs"
	CDRsTBL               = "cdrs"
	TBLTPSuppliers        = "tp_suppliers"
	TBLTPAttributes       = "tp_attributes"
	TBLVersions           = "versions"
)

//Cache Name
const (
	CacheDestinations              = "destinations"
	CacheReverseDestinations       = "reverse_destinations"
	CacheRatingPlans               = "rating_plans"
	CacheRatingProfiles            = "rating_profiles"
	CacheLCRRules                  = "lcr_rules"
	CacheCDRStatS                  = "cdr_stats"
	CacheActions                   = "actions"
	CacheActionPlans               = "action_plans"
	CacheAccountActionPlans        = "account_action_plans"
	CacheActionTriggers            = "action_triggers"
	CacheSharedGroups              = "shared_groups"
	CacheAliases                   = "aliases"
	CacheReverseAliases            = "reverse_aliases"
	CacheDerivedChargers           = "derived_chargers"
	CacheResources                 = "resources"
	CacheResourceProfiles          = "resource_profiles"
	CacheTimings                   = "timings"
	CacheEventResources            = "event_resources"
	CacheStatQueueProfiles         = "statqueue_profiles"
	CacheStatQueues                = "statqueues"
	CacheThresholdProfiles         = "threshold_profiles"
	CacheThresholds                = "thresholds"
	CacheFilters                   = "filters"
	CacheSupplierProfiles          = "supplier_profiles"
	CacheAttributeProfiles         = "attribute_profiles"
	CacheResourceFilterIndexes     = "resource_filter_indexes"
	CacheResourceFilterRevIndexes  = "resource_filter_revindexes"
	CacheStatFilterIndexes         = "stat_filter_indexes"
	CacheStatFilterRevIndexes      = "stat_filter_revindexes"
	CacheThresholdFilterIndexes    = "threshold_filter_indexes"
	CacheThresholdFilterRevIndexes = "threshold_filter_revindexes"
	CacheSupplierFilterIndexes     = "supplier_filter_indexes"
	CacheSupplierFilterRevIndexes  = "supplier_filter_revindexes"
	CacheAttributeFilterIndexes    = "attribute_filter_indexes"
	CacheAttributeFilterRevIndexes = "attribute_filter_revindexes"
)

//Prefix for indexing
const (
	ResourceFilterIndexes     = "rfi_"
	ResourceFilterRevIndexes  = "rfr_"
	StatFilterIndexes         = "sfi_"
	StatFilterRevIndexes      = "sfr_"
	ThresholdFilterIndexes    = "tfi_"
	ThresholdFilterRevIndexes = "tfr_"
	SupplierFilterIndexes     = "spi_"
	SupplierFilterRevIndexes  = "spr_"
	AttributeFilterIndexes    = "afi_"
	AttributeFilterRevIndexes = "afr_"
)

func buildCacheInstRevPrefixes() {
	CachePrefixToInstance = make(map[string]string)
	for k, v := range CacheInstanceToPrefix {
		CachePrefixToInstance[v] = k
	}
}

func init() {
	buildCacheInstRevPrefixes()
}
