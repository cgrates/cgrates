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
	CDRExportFormats = NewStringSet([]string{DRYRUN, MetaFileCSV, MetaFileFWV, MetaHTTPjsonCDR, MetaHTTPjsonMap,
		MetaHTTPjson, MetaHTTPPost, MetaAMQPjsonCDR, MetaAMQPjsonMap, MetaAMQPV1jsonMap, MetaSQSjsonMap,
		MetaKafkajsonMap, MetaS3jsonMap})
	MainCDRFields = NewStringSet([]string{CGRID, Source, OriginHost, OriginID, ToR, RequestType, Tenant, Category,
		Account, Subject, Destination, SetupTime, AnswerTime, Usage, COST, RATED, Partial, RunID,
		PreRated, CostSource, CostDetails, ExtraInfo, OrderID})
	PostPaidRatedSlice = []string{META_POSTPAID, META_RATED}
	ItemList           = NewStringSet([]string{MetaAccounts, MetaAttributes, MetaChargers, MetaDispatchers, MetaDispatcherHosts,
		MetaFilters, MetaResources, MetaStats, MetaThresholds, MetaSuppliers,
	})
	AttrInlineTypes = NewStringSet([]string{META_CONSTANT, MetaVariable, META_COMPOSED, META_USAGE_DIFFERENCE,
		MetaSum, MetaValueExponent})

	GitLastLog                  string // If set, it will be processed as part of versioning
	PosterTransportContentTypes = map[string]string{
		MetaHTTPjsonCDR:   CONTENT_JSON,
		MetaHTTPjsonMap:   CONTENT_JSON,
		MetaHTTPjson:      CONTENT_JSON,
		MetaHTTPPost:      CONTENT_FORM,
		MetaAMQPjsonCDR:   CONTENT_JSON,
		MetaAMQPjsonMap:   CONTENT_JSON,
		MetaAMQPV1jsonMap: CONTENT_JSON,
		MetaSQSjsonMap:    CONTENT_JSON,
		MetaKafkajsonMap:  CONTENT_JSON,
		MetaS3jsonMap:     CONTENT_JSON,
	}
	CDREFileSuffixes = map[string]string{
		MetaHTTPjsonCDR:   JSNSuffix,
		MetaHTTPjsonMap:   JSNSuffix,
		MetaAMQPjsonCDR:   JSNSuffix,
		MetaAMQPjsonMap:   JSNSuffix,
		MetaAMQPV1jsonMap: JSNSuffix,
		MetaSQSjsonMap:    JSNSuffix,
		MetaKafkajsonMap:  JSNSuffix,
		MetaS3jsonMap:     JSNSuffix,
		MetaHTTPPost:      FormSuffix,
		MetaFileCSV:       CSVSuffix,
		MetaFileFWV:       FWVSuffix,
	}
	// CachePartitions enables creation of cache partitions
	CachePartitions = NewStringSet([]string{CacheDestinations, CacheReverseDestinations,
		CacheRatingPlans, CacheRatingProfiles, CacheActions, CacheActionPlans,
		CacheAccountActionPlans, CacheActionTriggers, CacheSharedGroups, CacheTimings,
		CacheResourceProfiles, CacheResources, CacheEventResources, CacheStatQueueProfiles,
		CacheStatQueues, CacheThresholdProfiles, CacheThresholds, CacheFilters,
		CacheSupplierProfiles, CacheAttributeProfiles, CacheChargerProfiles,
		CacheDispatcherProfiles, CacheDispatcherHosts, CacheResourceFilterIndexes,
		CacheStatFilterIndexes, CacheThresholdFilterIndexes, CacheSupplierFilterIndexes,
		CacheAttributeFilterIndexes, CacheChargerFilterIndexes, CacheDispatcherFilterIndexes,
		CacheDispatcherRoutes, CacheDiameterMessages, CacheRPCResponses, CacheClosedSessions,
		CacheCDRIDs, CacheLoadIDs, CacheRPCConnections, CacheRatingProfilesTmp})
	CacheInstanceToPrefix = map[string]string{
		CacheDestinations:            DESTINATION_PREFIX,
		CacheReverseDestinations:     REVERSE_DESTINATION_PREFIX,
		CacheRatingPlans:             RATING_PLAN_PREFIX,
		CacheRatingProfiles:          RATING_PROFILE_PREFIX,
		CacheActions:                 ACTION_PREFIX,
		CacheActionPlans:             ACTION_PLAN_PREFIX,
		CacheAccountActionPlans:      AccountActionPlansPrefix,
		CacheActionTriggers:          ACTION_TRIGGER_PREFIX,
		CacheSharedGroups:            SHARED_GROUP_PREFIX,
		CacheResourceProfiles:        ResourceProfilesPrefix,
		CacheResources:               ResourcesPrefix,
		CacheTimings:                 TimingsPrefix,
		CacheStatQueueProfiles:       StatQueueProfilePrefix,
		CacheStatQueues:              StatQueuePrefix,
		CacheThresholdProfiles:       ThresholdProfilePrefix,
		CacheThresholds:              ThresholdPrefix,
		CacheFilters:                 FilterPrefix,
		CacheSupplierProfiles:        SupplierProfilePrefix,
		CacheAttributeProfiles:       AttributeProfilePrefix,
		CacheChargerProfiles:         ChargerProfilePrefix,
		CacheDispatcherProfiles:      DispatcherProfilePrefix,
		CacheDispatcherHosts:         DispatcherHostPrefix,
		CacheResourceFilterIndexes:   ResourceFilterIndexes,
		CacheStatFilterIndexes:       StatFilterIndexes,
		CacheThresholdFilterIndexes:  ThresholdFilterIndexes,
		CacheSupplierFilterIndexes:   SupplierFilterIndexes,
		CacheAttributeFilterIndexes:  AttributeFilterIndexes,
		CacheChargerFilterIndexes:    ChargerFilterIndexes,
		CacheDispatcherFilterIndexes: DispatcherFilterIndexes,
		CacheLoadIDs:                 LoadIDPrefix,
		CacheAccounts:                ACCOUNT_PREFIX,
	}
	CachePrefixToInstance map[string]string // will be built on init
	PrefixToIndexCache    = map[string]string{
		ThresholdProfilePrefix:  CacheThresholdFilterIndexes,
		ResourceProfilesPrefix:  CacheResourceFilterIndexes,
		StatQueueProfilePrefix:  CacheStatFilterIndexes,
		SupplierProfilePrefix:   CacheSupplierFilterIndexes,
		AttributeProfilePrefix:  CacheAttributeFilterIndexes,
		ChargerProfilePrefix:    CacheChargerFilterIndexes,
		DispatcherProfilePrefix: CacheDispatcherFilterIndexes,
	}
	CacheIndexesToPrefix map[string]string // will be built on init

	// NonMonetaryBalances are types of balances which are not handled as monetary
	NonMonetaryBalances = NewStringSet([]string{VOICE, SMS, DATA, GENERIC})

	// AccountableRequestTypes are the ones handled by Accounting subsystem
	AccountableRequestTypes = NewStringSet([]string{META_PREPAID, META_POSTPAID, META_PSEUDOPREPAID})

	CacheDataDBPartitions = NewStringSet([]string{CacheDestinations, CacheReverseDestinations,
		CacheRatingPlans, CacheRatingProfiles, CacheActions,
		CacheActionPlans, CacheAccountActionPlans, CacheActionTriggers, CacheSharedGroups, CacheResourceProfiles, CacheResources,
		CacheTimings, CacheStatQueueProfiles, CacheStatQueues, CacheThresholdProfiles, CacheThresholds,
		CacheFilters, CacheSupplierProfiles, CacheAttributeProfiles, CacheChargerProfiles,
		CacheDispatcherProfiles, CacheDispatcherHosts, CacheResourceFilterIndexes, CacheStatFilterIndexes,
		CacheThresholdFilterIndexes, CacheSupplierFilterIndexes, CacheAttributeFilterIndexes,
		CacheChargerFilterIndexes, CacheDispatcherFilterIndexes, CacheLoadIDs, CacheAccounts})

	// ProtectedSFlds are the fields that sessions should not alter
	ProtectedSFlds = NewStringSet([]string{CGRID, OriginHost, OriginID, Usage})
)

const (
	CGRateS                      = "CGRateS"
	VERSION                      = "v0.9.1~rc8"
	DIAMETER_FIRMWARE_REVISION   = 918
	REDIS_MAX_CONNS              = 10
	CGRATES                      = "cgrates"
	POSTGRES                     = "postgres"
	MYSQL                        = "mysql"
	MONGO                        = "mongo"
	REDIS                        = "redis"
	INTERNAL                     = "internal"
	DataManager                  = "DataManager"
	LOCALHOST                    = "127.0.0.1"
	PREPAID                      = "prepaid"
	META_PREPAID                 = "*prepaid"
	POSTPAID                     = "postpaid"
	META_POSTPAID                = "*postpaid"
	PSEUDOPREPAID                = "pseudoprepaid"
	META_PSEUDOPREPAID           = "*pseudoprepaid"
	META_RATED                   = "*rated"
	META_NONE                    = "*none"
	META_NOW                     = "*now"
	ROUNDING_UP                  = "*up"
	ROUNDING_MIDDLE              = "*middle"
	ROUNDING_DOWN                = "*down"
	ANY                          = "*any"
	MetaAll                      = "*all"
	ZERO                         = "*zero"
	ASAP                         = "*asap"
	COMMENT_CHAR                 = '#'
	CSV_SEP                      = ','
	FALLBACK_SEP                 = ';'
	INFIELD_SEP                  = ";"
	MetaPipe                     = "*|"
	FIELDS_SEP                   = ","
	InInFieldSep                 = ":"
	STATIC_HDRVAL_SEP            = "::"
	REGEXP_PREFIX                = "~"
	FILTER_VAL_START             = "("
	FILTER_VAL_END               = ")"
	JSON                         = "json"
	MSGPACK                      = "msgpack"
	CSV_LOAD                     = "CSVLOAD"
	CGRID                        = "CGRID"
	ToR                          = "ToR"
	OrderID                      = "OrderID"
	OriginID                     = "OriginID"
	InitialOriginID              = "InitialOriginID"
	OriginIDPrefix               = "OriginIDPrefix"
	Source                       = "Source"
	OriginHost                   = "OriginHost"
	RequestType                  = "RequestType"
	Direction                    = "Direction"
	Tenant                       = "Tenant"
	Category                     = "Category"
	Context                      = "Context"
	Contexts                     = "Contexts"
	Account                      = "Account"
	Subject                      = "Subject"
	Destination                  = "Destination"
	SetupTime                    = "SetupTime"
	AnswerTime                   = "AnswerTime"
	Usage                        = "Usage"
	Value                        = "Value"
	LastUsed                     = "LastUsed"
	PDD                          = "PDD"
	SUPPLIER                     = "Supplier"
	RunID                        = "RunID"
	MetaReqRunID                 = "*req.RunID"
	COST                         = "Cost"
	CostDetails                  = "CostDetails"
	RATED                        = "rated"
	Partial                      = "Partial"
	PreRated                     = "PreRated"
	STATIC_VALUE_PREFIX          = "^"
	CSV                          = "csv"
	FWV                          = "fwv"
	MetaPartialCSV               = "*partial_csv"
	DRYRUN                       = "dry_run"
	META_COMBIMED                = "*combimed"
	MetaMongo                    = "*mongo"
	MetaPostgres                 = "*postgres"
	MetaInternal                 = "*internal"
	MetaLocalHost                = "*localhost"
	ZERO_RATING_SUBJECT_PREFIX   = "*zero"
	OK                           = "OK"
	MetaFileXML                  = "*file_xml"
	CDRE                         = "cdre"
	MASK_CHAR                    = "*"
	CONCATENATED_KEY_SEP         = ":"
	UNIT_TEST                    = "UNIT_TEST"
	HDR_VAL_SEP                  = "/"
	MONETARY                     = "*monetary"
	SMS                          = "*sms"
	MMS                          = "*mms"
	GENERIC                      = "*generic"
	DATA                         = "*data"
	VOICE                        = "*voice"
	MAX_COST_FREE                = "*free"
	MAX_COST_DISCONNECT          = "*disconnect"
	SECONDS                      = "seconds"
	META_OUT                     = "*out"
	META_ANY                     = "*any"
	ASR                          = "ASR"
	ACD                          = "ACD"
	TASKS_KEY                    = "tasks"
	ACTION_PLAN_PREFIX           = "apl_"
	AccountActionPlansPrefix     = "aap_"
	ACTION_TRIGGER_PREFIX        = "atr_"
	RATING_PLAN_PREFIX           = "rpl_"
	RATING_PROFILE_PREFIX        = "rpf_"
	ACTION_PREFIX                = "act_"
	SHARED_GROUP_PREFIX          = "shg_"
	ACCOUNT_PREFIX               = "acc_"
	DESTINATION_PREFIX           = "dst_"
	REVERSE_DESTINATION_PREFIX   = "rds_"
	DERIVEDCHARGERS_PREFIX       = "dcs_"
	USERS_PREFIX                 = "usr_"
	ResourcesPrefix              = "res_"
	ResourceProfilesPrefix       = "rsp_"
	ThresholdPrefix              = "thd_"
	TimingsPrefix                = "tmg_"
	FilterPrefix                 = "ftr_"
	CDR_STATS_PREFIX             = "cst_"
	VERSION_PREFIX               = "ver_"
	StatQueueProfilePrefix       = "sqp_"
	SupplierProfilePrefix        = "spp_"
	AttributeProfilePrefix       = "alp_"
	ChargerProfilePrefix         = "cpp_"
	DispatcherProfilePrefix      = "dpp_"
	DispatcherHostPrefix         = "dph_"
	ThresholdProfilePrefix       = "thp_"
	StatQueuePrefix              = "stq_"
	LoadIDPrefix                 = "lid_"
	LOADINST_KEY                 = "load_history"
	CREATE_CDRS_TABLES_SQL       = "create_cdrs_tables.sql"
	CREATE_TARIFFPLAN_TABLES_SQL = "create_tariffplan_tables.sql"
	TEST_SQL                     = "TEST_SQL"
	META_CONSTANT                = "*constant"
	META_FILLER                  = "*filler"
	META_HANDLER                 = "*handler"
	MetaHTTPPost                 = "*http_post"
	MetaHTTPjson                 = "*http_json"
	MetaHTTPjsonCDR              = "*http_json_cdr"
	MetaHTTPjsonMap              = "*http_json_map"
	MetaAMQPjsonCDR              = "*amqp_json_cdr"
	MetaAMQPjsonMap              = "*amqp_json_map"
	MetaAMQPV1jsonMap            = "*amqpv1_json_map"
	MetaSQSjsonMap               = "*sqs_json_map"
	MetaKafkajsonMap             = "*kafka_json_map"
	MetaSQL                      = "*sql"
	MetaMySQL                    = "*mysql"
	MetaS3jsonMap                = "*s3_json_map"
	CONFIG_PATH                  = "/etc/cgrates/"
	DISCONNECT_CAUSE             = "DisconnectCause"
	MetaFlatstore                = "*flatstore"
	MetaRating                   = "*rating"
	NOT_AVAILABLE                = "N/A"
	CALL                         = "call"
	EXTRA_FIELDS                 = "ExtraFields"
	META_SURETAX                 = "*sure_tax"
	MetaDynamic                  = "*dynamic"
	COUNTER_EVENT                = "*event"
	COUNTER_BALANCE              = "*balance"
	EVENT_NAME                   = "EventName"
	// action trigger threshold types
	TRIGGER_MIN_EVENT_COUNTER   = "*min_event_counter"
	TRIGGER_MAX_EVENT_COUNTER   = "*max_event_counter"
	TRIGGER_MAX_BALANCE_COUNTER = "*max_balance_counter"
	TRIGGER_MIN_BALANCE         = "*min_balance"
	TRIGGER_MAX_BALANCE         = "*max_balance"
	TRIGGER_BALANCE_EXPIRED     = "*balance_expired"
	HIERARCHY_SEP               = ">"
	META_COMPOSED               = "*composed"
	META_USAGE_DIFFERENCE       = "*usage_difference"
	MetaDifference              = "*difference"
	MetaVariable                = "*variable"
	MetaCCUsage                 = "*cc_usage"
	MetaValueExponent           = "*value_exponent"
	NegativePrefix              = "!"
	MatchStartPrefix            = "^"
	MatchGreaterThanOrEqual     = ">="
	MatchLessThanOrEqual        = "<="
	MatchGreaterThan            = ">"
	MatchLessThan               = "<"
	MatchEndPrefix              = "$"
	MetaRaw                     = "*raw"
	CreatedAt                   = "CreatedAt"
	UpdatedAt                   = "UpdatedAt"
	HandlerArgSep               = "|"
	NodeID                      = "NodeID"
	ActiveGoroutines            = "ActiveGoroutines"
	MemoryUsage                 = "MemoryUsage"
	RunningSince                = "RunningSince"
	GoVersion                   = "GoVersion"
	SessionTTL                  = "SessionTTL"
	SessionTTLMaxDelay          = "SessionTTLMaxDelay"
	SessionTTLLastUsed          = "SessionTTLLastUsed"
	SessionTTLUsage             = "SessionTTLUsage"
	HandlerSubstractUsage       = "*substract_usage"
	XML                         = "xml"
	MetaGOB                     = "*gob"
	MetaJSON                    = "*json"
	MetaMSGPACK                 = "*msgpack"
	MetaDateTime                = "*datetime"
	MetaMaskedDestination       = "*masked_destination"
	MetaUnixTimestamp           = "*unix_timestamp"
	MetaPostCDR                 = "*post_cdr"
	MetaDumpToFile              = "*dump_to_file"
	NonTransactional            = ""
	DataDB                      = "data_db"
	StorDB                      = "stor_db"
	NotFoundCaps                = "NOT_FOUND"
	ServerErrorCaps             = "SERVER_ERROR"
	MandatoryIEMissingCaps      = "MANDATORY_IE_MISSING"
	UnsupportedCachePrefix      = "unsupported cache prefix"
	CDRSCtx                     = "cdrs"
	MandatoryInfoMissing        = "mandatory information missing"
	UnsupportedServiceIDCaps    = "UNSUPPORTED_SERVICE_ID"
	ServiceManager              = "service_manager"
	ServiceAlreadyRunning       = "service already running"
	RunningCaps                 = "RUNNING"
	StoppedCaps                 = "STOPPED"
	SchedulerNotRunningCaps     = "SCHEDULLER_NOT_RUNNING"
	MetaScheduler               = "*scheduler"
	MetaSessionsCosts           = "*sessions_costs"
	MetaRALs                    = "*rals"
	MetaRerate                  = "*rerate"
	MetaRefund                  = "*refund"
	MetaStats                   = "*stats"
	MetaResponder               = "*responder"
	MetaCore                    = "*core"
	MetaServiceManager          = "*servicemanager"
	MetaChargers                = "*chargers"
	MetaConfig                  = "*config"
	MetaDispatchers             = "*dispatchers"
	MetaDispatcherHosts         = "*dispatcher_hosts"
	MetaFilters                 = "*filters"
	MetaCDRs                    = "*cdrs"
	MetaCaches                  = "*caches"
	MetaGuardian                = "*guardians"
	MetaContinue                = "*continue"
	Migrator                    = "migrator"
	UnsupportedMigrationTask    = "unsupported migration task"
	NoStorDBConnection          = "not connected to StorDB"
	UndefinedVersion            = "undefined version"
	TxtSuffix                   = ".txt"
	JSNSuffix                   = ".json"
	GOBSuffix                   = ".gob"
	FormSuffix                  = ".form"
	XMLSuffix                   = ".xml"
	CSVSuffix                   = ".csv"
	FWVSuffix                   = ".fwv"
	CONTENT_JSON                = "json"
	CONTENT_FORM                = "form"
	CONTENT_TEXT                = "text"
	FileLockPrefix              = "file_"
	ActionsPoster               = "act"
	CDRPoster                   = "cdr"
	MetaFileCSV                 = "*file_csv"
	MetaFileFWV                 = "*file_fwv"
	MetaFScsv                   = "*freeswitch_csv"
	Accounts                    = "Accounts"
	AccountService              = "AccountS"
	Actions                     = "Actions"
	ActionPlans                 = "ActionPlans"
	ActionTriggers              = "ActionTriggers"
	SharedGroups                = "SharedGroups"
	TimingIDs                   = "TimingIDs"
	Timings                     = "Timings"
	Rates                       = "Rates"
	DestinationRates            = "DestinationRates"
	RatingPlans                 = "RatingPlans"
	RatingProfiles              = "RatingProfiles"
	AccountActions              = "AccountActions"
	Resources                   = "Resources"
	Stats                       = "Stats"
	Filters                     = "Filters"
	DispatcherProfiles          = "DispatcherProfiles"
	DispatcherHosts             = "DispatcherHosts"
	MetaEveryMinute             = "*every_minute"
	MetaHourly                  = "*hourly"
	ID                          = "ID"
	Thresholds                  = "Thresholds"
	Suppliers                   = "Suppliers"
	Attributes                  = "Attributes"
	Chargers                    = "Chargers"
	Dispatchers                 = "Dispatchers"
	StatS                       = "Stats"
	LoadIDsVrs                  = "LoadIDs"
	RALService                  = "RALs"
	CostSource                  = "CostSource"
	ExtraInfo                   = "ExtraInfo"
	Meta                        = "*"
	MetaSysLog                  = "*syslog"
	MetaStdLog                  = "*stdout"
	EventType                   = "EventType"
	EventSource                 = "EventSource"
	AccountID                   = "AccountID"
	ResourceID                  = "ResourceID"
	TotalUsage                  = "TotalUsage"
	StatID                      = "StatID"
	BalanceType                 = "BalanceType"
	BalanceID                   = "BalanceID"
	BalanceDestinationIds       = "BalanceDestinationIds"
	BalanceWeight               = "BalanceWeight"
	BalanceExpirationDate       = "BalanceExpirationDate"
	BalanceTimingTags           = "BalanceTimingTags"
	BalanceRatingSubject        = "BalanceRatingSubject"
	BalanceCategories           = "BalanceCategories"
	BalanceSharedGroups         = "BalanceSharedGroups"
	BalanceBlocker              = "BalanceBlocker"
	BalanceDisabled             = "BalanceDisabled"
	Units                       = "Units"
	AccountUpdate               = "AccountUpdate"
	BalanceUpdate               = "BalanceUpdate"
	StatUpdate                  = "StatUpdate"
	ResourceUpdate              = "ResourceUpdate"
	CDR                         = "CDR"
	CDRs                        = "CDRs"
	ExpiryTime                  = "ExpiryTime"
	AllowNegative               = "AllowNegative"
	Disabled                    = "Disabled"
	Action                      = "Action"
	MetaNow                     = "*now"
	SessionSCosts               = "SessionSCosts"
	Timing                      = "Timing"
	RQF                         = "RQF"
	Resource                    = "Resource"
	User                        = "User"
	Subscribers                 = "Subscribers"
	DerivedChargersV            = "DerivedChargers"
	Destinations                = "Destinations"
	ReverseDestinations         = "ReverseDestinations"
	RatingPlan                  = "RatingPlan"
	RatingProfile               = "RatingProfile"
	MetaRatingPlans             = "*rating_plans"
	MetaRatingProfiles          = "*rating_profiles"
	MetaUsers                   = "*users"
	MetaSubscribers             = "*subscribers"
	MetaDerivedChargersV        = "*derivedchargers"
	MetaStorDB                  = "*stordb"
	MetaDataDB                  = "*datadb"
	MetaWeight                  = "*weight"
	MetaLC                      = "*lc"
	MetaHC                      = "*hc"
	MetaQOS                     = "*qos"
	MetaReas                    = "*reas"
	MetaReds                    = "*reds"
	Weight                      = "Weight"
	Cost                        = "Cost"
	DestinationIDs              = "DestinationIDs"
	RatingSubject               = "RatingSubject"
	Categories                  = "Categories"
	Blocker                     = "Blocker"
	RatingPlanID                = "RatingPlanID"
	MetaSessionS                = "*sessions"
	MetaDefault                 = "*default"
	Error                       = "Error"
	MetaCgreq                   = "*cgreq"
	MetaCgrep                   = "*cgrep"
	MetaCGRAReq                 = "*cgrareq"
	CGR_ACD                     = "cgr_acd"
	FilterIDs                   = "FilterIDs"
	FieldName                   = "FieldName"
	Path                        = "Path"
	MetaRound                   = "*round"
	Pong                        = "Pong"
	MetaEventCost               = "*event_cost"
	MetaSuppliersEventCost      = "*suppliers_event_cost"
	MetaSuppliersIgnoreErrors   = "*suppliers_ignore_errors"
	Freeswitch                  = "freeswitch"
	Kamailio                    = "kamailio"
	Opensips                    = "opensips"
	Asterisk                    = "asterisk"
	SchedulerS                  = "SchedulerS"
	MetaMultiply                = "*multiply"
	MetaDivide                  = "*divide"
	MetaUrl                     = "*url"
	MetaXml                     = "*xml"
	ApiKey                      = "apikey"
	MetaReq                     = "*req"
	MetaVars                    = "*vars"
	MetaRep                     = "*rep"
	MetaExp                     = "*exp"
	MetaHdr                     = "*hdr"
	MetaTrl                     = "*trl"
	CGROriginHost               = "cgr_originhost"
	MetaInitiate                = "*initiate"
	MetaUpdate                  = "*update"
	MetaTerminate               = "*terminate"
	MetaEvent                   = "*event"
	MetaMessage                 = "*message"
	MetaDryRun                  = "*dryrun"
	Event                       = "Event"
	EmptyString                 = ""
	DynamicDataPrefix           = "~"
	AttrValueSep                = "="
	ANDSep                      = "&"
	PipeSep                     = "|"
	MetaApp                     = "*app"
	MetaAppID                   = "*appid"
	MetaCmd                     = "*cmd"
	MetaEnv                     = "*env:" // use in config for describing enviormant variables
	MetaTemplate                = "*template"
	MetaCCA                     = "*cca"
	MetaErr                     = "*err"
	OriginRealm                 = "OriginRealm"
	ProductName                 = "ProductName"
	IdxStart                    = "["
	IdxEnd                      = "]"
	MetaLog                     = "*log"
	MetaRemoteHost              = "*remote_host"
	RemoteHost                  = "RemoteHost"
	Local                       = "local"
	TCP                         = "tcp"
	CGRDebitInterval            = "CGRDebitInterval"
	Version                     = "Version"
	MetaTenant                  = "*tenant"
	ResourceUsage               = "ResourceUsage"
	MetaDuration                = "*duration"
	MetaLibPhoneNumber          = "*libphonenumber"
	MetaReload                  = "*reload"
	MetaLoad                    = "*load"
	MetaRemove                  = "*remove"
	MetaStore                   = "*store"
	MetaClear                   = "*clear"
	MetaExport                  = "*export"
	LoadIDs                     = "load_ids"
	DNSAgent                    = "DNSAgent"
	TLSNoCaps                   = "tls"
	MetaRouteID                 = "*route_id"
	MetaApiKey                  = "*api_key"
	UsageID                     = "UsageID"
	Rcode                       = "Rcode"
	Replacement                 = "Replacement"
	Regexp                      = "Regexp"
	Order                       = "Order"
	Preference                  = "Preference"
	Flags                       = "Flags"
	Service                     = "Service"
	MetaSuppliersLimit          = "*suppliers_limit"
	MetaSuppliersOffset         = "*suppliers_offset"
	ApierV                      = "ApierV"
	MetaApier                   = "*apier"
	MetaAnalyzer                = "*analyzer"
	CGREventString              = "CGREvent"
	MetaTextPlain               = "*text_plain"
	MetaIgnoreErrors            = "*ignore_errors"
	MetaRelease                 = "*release"
	MetaAllocate                = "*allocate"
	MetaAuthorize               = "*authorize"
	MetaInit                    = "*init"
	MetaRatingPlanCost          = "*rating_plan_cost"
	RatingPlanIDs               = "RatingPlanIDs"
	ERs                         = "ERs"
	Ratio                       = "Ratio"
	Load                        = "Load"
	Slash                       = "/"
	UUID                        = "UUID"
	ActionsID                   = "ActionsID"
	MetaAct                     = "*act"
	DestinationPrefix           = "DestinationPrefix"
	DestinationID               = "DestinationID"
	ExportTemplate              = "ExportTemplate"
	ExportFormat                = "ExportFormat"
	Synchronous                 = "Synchronous"
	Attempts                    = "Attempts"
	FieldSeparator              = "FieldSeparator"
	ExportPath                  = "ExportPath"
	ExportID                    = "ExportID"
	ExportFileName              = "ExportFileName"
	GroupID                     = "GroupID"
	ThresholdType               = "ThresholdType"
	ThresholdValue              = "ThresholdValue"
	Recurrent                   = "Recurrent"
	Executed                    = "Executed"
	MinSleep                    = "MinSleep"
	ActivationDate              = "ActivationDate"
	ExpirationDate              = "ExpirationDate"
	MinQueuedItems              = "MinQueuedItems"
	OrderIDStart                = "OrderIDStart"
	OrderIDEnd                  = "OrderIDEnd"
	MinCost                     = "MinCost"
	MaxCost                     = "MaxCost"
	MetaLoaders                 = "*loaders"
	TmpSuffix                   = ".tmp"
)

// Migrator Action
const (
	Move    = "move"
	Migrate = "migrate"
)

// Meta Items
const (
	MetaAccounts            = "*accounts"
	MetaAccountActionPlans  = "*account_action_plans"
	MetaReverseDestinations = "*reverse_destinations"
	MetaActionPlans         = "*action_plans"
	MetaActionTriggers      = "*action_triggers"
	MetaActions             = "*actions"
	MetaResourceProfile     = "*resource_profiles"
	MetaStatQueueProfiles   = "*statqueue_profiles"
	MetaStatQueues          = "*statqueues"
	MetaThresholdProfiles   = "*threshold_profiles"
	MetaSupplierProfiles    = "*supplier_profiles"
	MetaAttributeProfiles   = "*attribute_profiles"
	MetaFilterIndexes       = "*filter_indexes"
	MetaDispatcherProfiles  = "*dispatcher_profiles"
	MetaChargerProfiles     = "*charger_profiles"
	MetaSharedGroups        = "*shared_groups"
	MetaThresholds          = "*thresholds"
	MetaSuppliers           = "*suppliers"
	MetaAttributes          = "*attributes"
	MetaLoadIDs             = "*load_ids"
)

// MetaMetrics
const (
	MetaASR      = "*asr"
	MetaACD      = "*acd"
	MetaTCD      = "*tcd"
	MetaACC      = "*acc"
	MetaTCC      = "*tcc"
	MetaPDD      = "*pdd"
	MetaDDC      = "*ddc"
	MetaSum      = "*sum"
	MetaAverage  = "*average"
	MetaDistinct = "*distinct"
)

// Services
const (
	SessionS    = "SessionS"
	AttributeS  = "AttributeS"
	SupplierS   = "SupplierS"
	ResourceS   = "ResourceS"
	StatService = "StatS"
	FilterS     = "FilterS"
	ThresholdS  = "ThresholdS"
	DispatcherS = "DispatcherS"
	LoaderS     = "LoaderS"
	ChargerS    = "ChargerS"
	CacheS      = "CacheS"
	AnalyzerS   = "AnalyzerS"
	CDRServer   = "CDRServer"
	ResponderS  = "ResponderS"
	GuardianS   = "GuardianS"
	ApierS      = "ApierS"
)

// Lower service names
const (
	SessionsLow    = "sessions"
	AttributesLow  = "attributes"
	ChargerSLow    = "chargers"
	SuppliersLow   = "suppliers"
	ResourcesLow   = "resources"
	StatServiceLow = "stats"
	ThresholdsLow  = "thresholds"
	DispatcherSLow = "dispatchers"
	AnalyzerSLow   = "analyzers"
	SchedulerSLow  = "schedulers"
	LoaderSLow     = "loaders"
	RALsLow        = "rals"
	ReplicatorLow  = "replicator"
	ApierSLow      = "apiers"
)

// Actions
const (
	LOG                       = "*log"
	RESET_TRIGGERS            = "*reset_triggers"
	SET_RECURRENT             = "*set_recurrent"
	UNSET_RECURRENT           = "*unset_recurrent"
	ALLOW_NEGATIVE            = "*allow_negative"
	DENY_NEGATIVE             = "*deny_negative"
	RESET_ACCOUNT             = "*reset_account"
	REMOVE_ACCOUNT            = "*remove_account"
	SET_BALANCE               = "*set_balance"
	REMOVE_BALANCE            = "*remove_balance"
	TOPUP_RESET               = "*topup_reset"
	TOPUP                     = "*topup"
	DEBIT_RESET               = "*debit_reset"
	DEBIT                     = "*debit"
	RESET_COUNTERS            = "*reset_counters"
	ENABLE_ACCOUNT            = "*enable_account"
	DISABLE_ACCOUNT           = "*disable_account"
	HttpPost                  = "*http_post"
	HttpPostAsync             = "*http_post_async"
	MAIL_ASYNC                = "*mail_async"
	UNLIMITED                 = "*unlimited"
	CDRLOG                    = "*cdrlog"
	SET_DDESTINATIONS         = "*set_ddestinations"
	TRANSFER_MONETARY_DEFAULT = "*transfer_monetary_default"
	CGR_RPC                   = "*cgr_rpc"
	TopUpZeroNegative         = "*topup_zero_negative"
	SetExpiry                 = "*set_expiry"
	MetaPublishAccount        = "*publish_account"
	MetaPublishBalance        = "*publish_balance"
	MetaRemoveSessionCosts    = "*remove_session_costs"
	MetaRemoveExpired         = "*remove_expired"
	MetaPostEvent             = "*post_event"
	MetaCDRAccount            = "*cdr_account"
)

// Migrator Metas
const (
	MetaSetVersions         = "*set_versions"
	MetaEnsureIndexes       = "*ensure_indexes"
	MetaTpRatingPlans       = "*tp_rating_plans"
	MetaTpFilters           = "*tp_filters"
	MetaTpDestinationRates  = "*tp_destination_rates"
	MetaTpActionTriggers    = "*tp_action_triggers"
	MetaTpAccountActions    = "*tp_account_actions"
	MetaTpActionPlans       = "*tp_action_plans"
	MetaTpActions           = "*tp_actions"
	MetaTpThresholds        = "*tp_thresholds"
	MetaTpSuppliers         = "*tp_suppliers"
	MetaTpStats             = "*tp_stats"
	MetaTpSharedGroups      = "*tp_shared_groups"
	MetaTpRatingProfiles    = "*tp_rating_profiles"
	MetaTpResources         = "*tp_resources"
	MetaTpRates             = "*tp_rates"
	MetaTpTimings           = "*tp_timings"
	MetaTpDestinations      = "*tp_destinations"
	MetaTpChargers          = "*tp_chargers"
	MetaTpDispatchers       = "*tp_dispatchers"
	MetaDurationSeconds     = "*duration_seconds"
	MetaDurationNanoseconds = "*duration_nanoseconds"
	CapAttributes           = "Attributes"
	CapResourceMessage      = "ResourceMessage"
	CapResourceAllocation   = "ResourceAllocation"
	CapMaxUsage             = "MaxUsage"
	CapSuppliers            = "Suppliers"
	CapThresholds           = "Thresholds"
	CapStatQueues           = "StatQueues"
)

const (
	TpRatingPlans      = "TpRatingPlans"
	TpFilters          = "TpFilters"
	TpDestinationRates = "TpDestinationRates"
	TpActionTriggers   = "TpActionTriggers"
	TpAccountActionsV  = "TpAccountActions"
	TpActionPlans      = "TpActionPlans"
	TpActions          = "TpActions"
	TpThresholds       = "TpThresholds"
	TpSuppliers        = "TpSuppliers"
	TpStats            = "TpStats"
	TpSharedGroups     = "TpSharedGroups"
	TpRatingProfiles   = "TpRatingProfiles"
	TpResources        = "TpResources"
	TpRates            = "TpRates"
	TpTiming           = "TpTiming"
	TpResource         = "TpResource"
	TpDestinations     = "TpDestinations"
	TpRatingPlan       = "TpRatingPlan"
	TpRatingProfile    = "TpRatingProfile"
	TpChargers         = "TpChargers"
	TpDispatchers      = "TpDispatchers"
)

// Dispatcher Const
const (
	MetaFirst          = "*first"
	MetaRandom         = "*random"
	MetaBroadcast      = "*broadcast"
	MetaRoundRobin     = "*round_robin"
	MetaRatio          = "*ratio"
	ThresholdSv1       = "ThresholdSv1"
	StatSv1            = "StatSv1"
	ResourceSv1        = "ResourceSv1"
	SupplierSv1        = "SupplierSv1"
	AttributeSv1       = "AttributeSv1"
	SessionSv1         = "SessionSv1"
	ChargerSv1         = "ChargerSv1"
	MetaAuth           = "*auth"
	APIKey             = "APIKey"
	RouteID            = "RouteID"
	APIMethods         = "APIMethods"
	NestingSep         = "."
	ArgDispatcherField = "ArgDispatcher"
)

//Filter types
const (
	MetaNot            = "*not"
	MetaString         = "*string"
	MetaPrefix         = "*prefix"
	MetaSuffix         = "*suffix"
	MetaEmpty          = "*empty"
	MetaExists         = "*exists"
	MetaTimings        = "*timings"
	MetaRSR            = "*rsr"
	MetaStatS          = "*stats"
	MetaDestinations   = "*destinations"
	MetaLessThan       = "*lt"
	MetaLessOrEqual    = "*lte"
	MetaGreaterThan    = "*gt"
	MetaGreaterOrEqual = "*gte"
	MetaResources      = "*resources"
	MetaEqual          = "*eq"

	MetaNotString       = "*notstring"
	MetaNotPrefix       = "*notprefix"
	MetaNotSuffix       = "*notsuffix"
	MetaNotEmpty        = "*notempty"
	MetaNotExists       = "*notexists"
	MetaNotTimings      = "*nottimings"
	MetaNotRSR          = "*notrsr"
	MetaNotStatS        = "*notstats"
	MetaNotDestinations = "*notdestinations"
	MetaNotResources    = "*notresources"
	MetaNotEqual        = "*noteq"
)

// ReplicatorSv1 APIs
const (
	ReplicatorSv1                        = "ReplicatorSv1"
	ReplicatorSv1Ping                    = "ReplicatorSv1.Ping"
	ReplicatorSv1GetAccount              = "ReplicatorSv1.GetAccount"
	ReplicatorSv1GetDestination          = "ReplicatorSv1.GetDestination"
	ReplicatorSv1GetReverseDestination   = "ReplicatorSv1.GetReverseDestination"
	ReplicatorSv1GetStatQueue            = "ReplicatorSv1.GetStatQueue"
	ReplicatorSv1GetFilter               = "ReplicatorSv1.GetFilter"
	ReplicatorSv1GetThreshold            = "ReplicatorSv1.GetThreshold"
	ReplicatorSv1GetThresholdProfile     = "ReplicatorSv1.GetThresholdProfile"
	ReplicatorSv1GetStatQueueProfile     = "ReplicatorSv1.GetStatQueueProfile"
	ReplicatorSv1GetTiming               = "ReplicatorSv1.GetTiming"
	ReplicatorSv1GetResource             = "ReplicatorSv1.GetResource"
	ReplicatorSv1GetResourceProfile      = "ReplicatorSv1.GetResourceProfile"
	ReplicatorSv1GetActionTriggers       = "ReplicatorSv1.GetActionTriggers"
	ReplicatorSv1GetShareGroup           = "ReplicatorSv1.GetShareGroup"
	ReplicatorSv1GetActions              = "ReplicatorSv1.GetActions"
	ReplicatorSv1GetActionPlan           = "ReplicatorSv1.GetActionPlan"
	ReplicatorSv1GetAllActionPlans       = "ReplicatorSv1.GetAllActionPlans"
	ReplicatorSv1GetAccountActionPlans   = "ReplicatorSv1.GetAccountActionPlans"
	ReplicatorSv1GetRatingPlan           = "ReplicatorSv1.GetRatingPlan"
	ReplicatorSv1GetRatingProfile        = "ReplicatorSv1.GetRatingProfile"
	ReplicatorSv1GetSupplierProfile      = "ReplicatorSv1.GetSupplierProfile"
	ReplicatorSv1GetAttributeProfile     = "ReplicatorSv1.GetAttributeProfile"
	ReplicatorSv1GetChargerProfile       = "ReplicatorSv1.GetChargerProfile"
	ReplicatorSv1GetDispatcherProfile    = "ReplicatorSv1.GetDispatcherProfile"
	ReplicatorSv1GetDispatcherHost       = "ReplicatorSv1.GetDispatcheHost"
	ReplicatorSv1GetItemLoadIDs          = "ReplicatorSv1.GetItemLoadIDs"
	ReplicatorSv1GetFilterIndexes        = "ReplicatorSv1.GetFilterIndexes"
	ReplicatorSv1MatchFilterIndex        = "ReplicatorSv1.MatchFilterIndex"
	ReplicatorSv1SetThresholdProfile     = "ReplicatorSv1.SetThresholdProfile"
	ReplicatorSv1SetThreshold            = "ReplicatorSv1.SetThreshold"
	ReplicatorSv1SetFilterIndexes        = "ReplicatorSv1.SetFilterIndexes"
	ReplicatorSv1Account                 = "ReplicatorSv1.SetAccount"
	ReplicatorSv1SetDestination          = "ReplicatorSv1.SetDestination"
	ReplicatorSv1SetReverseDestination   = "ReplicatorSv1.SetReverseDestination"
	ReplicatorSv1SetStatQueue            = "ReplicatorSv1.SetStatQueue"
	ReplicatorSv1SetFilter               = "ReplicatorSv1.SetFilter"
	ReplicatorSv1SetStatQueueProfile     = "ReplicatorSv1.SetStatQueueProfile"
	ReplicatorSv1SetTiming               = "ReplicatorSv1.SetTiming"
	ReplicatorSv1SetResource             = "ReplicatorSv1.SetResource"
	ReplicatorSv1SetResourceProfile      = "ReplicatorSv1.SetResourceProfile"
	ReplicatorSv1SetActionTriggers       = "ReplicatorSv1.SetActionTriggers"
	ReplicatorSv1SetSharedGroup          = "ReplicatorSv1.SetSharedGroup"
	ReplicatorSv1SetActions              = "ReplicatorSv1.SetActions"
	ReplicatorSv1SetActionPlan           = "ReplicatorSv1.SetActionPlan"
	ReplicatorSv1SetAccountActionPlans   = "ReplicatorSv1.SetAccountActionPlans"
	ReplicatorSv1SetRatingPlan           = "ReplicatorSv1.SetRatingPlan"
	ReplicatorSv1SetRatingProfile        = "ReplicatorSv1.SetRatingProfile"
	ReplicatorSv1SetSupplierProfile      = "ReplicatorSv1.SetSupplierProfile"
	ReplicatorSv1SetAttributeProfile     = "ReplicatorSv1.SetAttributeProfile"
	ReplicatorSv1SetChargerProfile       = "ReplicatorSv1.SetChargerProfile"
	ReplicatorSv1SetDispatcherProfile    = "ReplicatorSv1.SetDispatcherProfile"
	ReplicatorSv1SetDispatcherHost       = "ReplicatorSv1.SetDispatcherHost"
	ReplicatorSv1SetLoadIDs              = "ReplicatorSv1.SetLoadIDs"
	ReplicatorSv1RemoveThreshold         = "ReplicatorSv1.RemoveThreshold"
	ReplicatorSv1RemoveDestination       = "ReplicatorSv1.RemoveDestination"
	ReplicatorSv1RemoveAccount           = "ReplicatorSv1.RemoveAccount"
	ReplicatorSv1RemoveStatQueue         = "ReplicatorSv1.RemoveStatQueue"
	ReplicatorSv1RemoveFilter            = "ReplicatorSv1.RemoveFilter"
	ReplicatorSv1RemoveThresholdProfile  = "ReplicatorSv1.RemoveThresholdProfile"
	ReplicatorSv1RemoveStatQueueProfile  = "ReplicatorSv1.RemoveStatQueueProfile"
	ReplicatorSv1RemoveTiming            = "ReplicatorSv1.RemoveTiming"
	ReplicatorSv1RemoveResource          = "ReplicatorSv1.RemoveResource"
	ReplicatorSv1RemoveResourceProfile   = "ReplicatorSv1.RemoveResourceProfile"
	ReplicatorSv1RemoveActionTriggers    = "ReplicatorSv1.RemoveActionTriggers"
	ReplicatorSv1RemoveSharedGroup       = "ReplicatorSv1.RemoveSharedGroup"
	ReplicatorSv1RemoveActions           = "ReplicatorSv1.RemoveActions"
	ReplicatorSv1RemoveActionPlan        = "ReplicatorSv1.RemoveActionPlan"
	ReplicatorSv1RemAccountActionPlans   = "ReplicatorSv1.RemAccountActionPlans"
	ReplicatorSv1RemoveRatingPlan        = "ReplicatorSv1.RemoveRatingPlan"
	ReplicatorSv1RemoveRatingProfile     = "ReplicatorSv1.RemoveRatingProfile"
	ReplicatorSv1RemoveSupplierProfile   = "ReplicatorSv1.RemoveSupplierProfile"
	ReplicatorSv1RemoveAttributeProfile  = "ReplicatorSv1.RemoveAttributeProfile"
	ReplicatorSv1RemoveChargerProfile    = "ReplicatorSv1.RemoveChargerProfile"
	ReplicatorSv1RemoveDispatcherProfile = "ReplicatorSv1.RemoveDispatcherProfile"
	ReplicatorSv1RemoveDispatcherHost    = "ReplicatorSv1.RemoveDispatcherHost"
)

// ApierV1 APIs
const (
	ApierV1                            = "ApierV1"
	ApierV1ComputeFilterIndexes        = "ApierV1.ComputeFilterIndexes"
	ApierV1ComputeFilterIndexIDs       = "ApierV1.ComputeFilterIndexIDs"
	ApierV1Ping                        = "ApierV1.Ping"
	ApierV1SetDispatcherProfile        = "ApierV1.SetDispatcherProfile"
	ApierV1GetDispatcherProfile        = "ApierV1.GetDispatcherProfile"
	ApierV1GetDispatcherProfileIDs     = "ApierV1.GetDispatcherProfileIDs"
	ApierV1RemoveDispatcherProfile     = "ApierV1.RemoveDispatcherProfile"
	ApierV1SetDispatcherHost           = "ApierV1.SetDispatcherHost"
	ApierV1GetDispatcherHost           = "ApierV1.GetDispatcherHost"
	ApierV1GetDispatcherHostIDs        = "ApierV1.GetDispatcherHostIDs"
	ApierV1RemoveDispatcherHost        = "ApierV1.RemoveDispatcherHost"
	ApierV1GetEventCost                = "ApierV1.GetEventCost"
	ApierV1LoadTariffPlanFromFolder    = "ApierV1.LoadTariffPlanFromFolder"
	ApierV1GetCost                     = "ApierV1.GetCost"
	ApierV1SetBalance                  = "ApierV1.SetBalance"
	ApierV1GetFilter                   = "ApierV1.GetFilter"
	ApierV1GetFilterIndexes            = "ApierV1.GetFilterIndexes"
	ApierV1RemoveFilterIndexes         = "ApierV1.RemoveFilterIndexes"
	ApierV1RemoveFilter                = "ApierV1.RemoveFilter"
	ApierV1SetFilter                   = "ApierV1.SetFilter"
	ApierV1GetFilterIDs                = "ApierV1.GetFilterIDs"
	ApierV1GetRatingProfile            = "ApierV1.GetRatingProfile"
	ApierV1RemoveRatingProfile         = "ApierV1.RemoveRatingProfile"
	ApierV1SetRatingProfile            = "ApierV1.SetRatingProfile"
	ApierV1GetRatingProfileIDs         = "ApierV1.GetRatingProfileIDs"
	ApierV1SetDataDBVersions           = "ApierV1.SetDataDBVersions"
	ApierV1SetStorDBVersions           = "ApierV1.SetStorDBVersions"
	ApierV1GetAccountActionPlan        = "ApierV1.GetAccountActionPlan"
	ApierV1ComputeActionPlanIndexes    = "ApierV1.ComputeActionPlanIndexes"
	ApierV1GetActions                  = "ApierV1.GetActions"
	ApierV1GetActionPlan               = "ApierV1.GetActionPlan"
	ApierV1GetActionPlanIDs            = "ApierV1.GetActionPlanIDs"
	ApierV1GetRatingPlanIDs            = "ApierV1.GetRatingPlanIDs"
	ApierV1GetRatingPlan               = "ApierV1.GetRatingPlan"
	ApierV1RemoveRatingPlan            = "ApierV1.RemoveRatingPlan"
	ApierV1GetDestination              = "ApierV1.GetDestination"
	ApierV1GetReverseDestination       = "ApierV1.GetReverseDestination"
	ApierV1AddBalance                  = "ApierV1.AddBalance"
	ApierV1DebitBalance                = "ApierV1.DebitBalance"
	ApierV1SetAccount                  = "ApierV1.SetAccount"
	ApierV1GetAccountsCount            = "ApierV1.GetAccountsCount"
	ApierV1GetDataDBVersions           = "ApierV1.GetDataDBVersions"
	ApierV1GetStorDBVersions           = "ApierV1.GetStorDBVersions"
	ApierV1GetCDRs                     = "ApierV1.GetCDRs"
	ApierV1GetTPAccountActions         = "ApierV1.GetTPAccountActions"
	ApierV1SetTPAccountActions         = "ApierV1.SetTPAccountActions"
	ApierV1GetTPAccountActionsByLoadId = "ApierV1.GetTPAccountActionsByLoadId"
	ApierV1GetTPAccountActionLoadIds   = "ApierV1.GetTPAccountActionLoadIds"
	ApierV1GetTPAccountActionIds       = "ApierV1.GetTPAccountActionIds"
	ApierV1RemoveTPAccountActions      = "ApierV1.RemoveTPAccountActions"
	ApierV1GetTPActionPlan             = "ApierV1.GetTPActionPlan"
	ApierV1SetTPActionPlan             = "ApierV1.SetTPActionPlan"
	ApierV1GetTPActionPlanIds          = "ApierV1.GetTPActionPlanIds"
	ApierV1SetTPActionTriggers         = "ApierV1.SetTPActionTriggers"
	ApierV1GetTPActionTriggers         = "ApierV1.GetTPActionTriggers"
	ApierV1RemoveTPActionTriggers      = "ApierV1.RemoveTPActionTriggers"
	ApierV1GetTPActionTriggerIds       = "ApierV1.GetTPActionTriggerIds"
	ApierV1GetTPActions                = "ApierV1.GetTPActions"
	ApierV1RemoveTPActionPlan          = "ApierV1.RemoveTPActionPlan"
	ApierV1GetTPAttributeProfile       = "ApierV1.GetTPAttributeProfile"
	ApierV1SetTPAttributeProfile       = "ApierV1.SetTPAttributeProfile"
	ApierV1GetTPAttributeProfileIds    = "ApierV1.GetTPAttributeProfileIds"
	ApierV1RemoveTPAttributeProfile    = "ApierV1.RemoveTPAttributeProfile"
	ApierV1GetTPCharger                = "ApierV1.GetTPCharger"
	ApierV1SetTPCharger                = "ApierV1.SetTPCharger"
	ApierV1RemoveTPCharger             = "ApierV1.RemoveTPCharger"
	ApierV1GetTPChargerIDs             = "ApierV1.GetTPChargerIDs"
	ApierV1SetTPFilterProfile          = "ApierV1.SetTPFilterProfile"
	ApierV1GetTPFilterProfile          = "ApierV1.GetTPFilterProfile"
	ApierV1GetTPFilterProfileIds       = "ApierV1.GetTPFilterProfileIds"
	ApierV1RemoveTPFilterProfile       = "ApierV1.RemoveTPFilterProfile"
	ApierV1GetTPDestination            = "ApierV1.GetTPDestination"
	ApierV1SetTPDestination            = "ApierV1.SetTPDestination"
	ApierV1GetTPDestinationIDs         = "ApierV1.GetTPDestinationIDs"
	ApierV1RemoveTPDestination         = "ApierV1.RemoveTPDestination"
	ApierV1GetTPResource               = "ApierV1.GetTPResource"
	ApierV1SetTPResource               = "ApierV1.SetTPResource"
	ApierV1RemoveTPResource            = "ApierV1.RemoveTPResource"
	ApierV1SetTPRate                   = "ApierV1.SetTPRate"
	ApierV1GetTPRate                   = "ApierV1.GetTPRate"
	ApierV1RemoveTPRate                = "ApierV1.RemoveTPRate"
	ApierV1GetTPRateIds                = "ApierV1.GetTPRateIds"
	ApierV1SetTPThreshold              = "ApierV1.SetTPThreshold"
	ApierV1GetTPThreshold              = "ApierV1.GetTPThreshold"
	ApierV1GetTPThresholdIDs           = "ApierV1.GetTPThresholdIDs"
	ApierV1RemoveTPThreshold           = "ApierV1.RemoveTPThreshold"
	ApierV1SetTPStat                   = "ApierV1.SetTPStat"
	ApierV1GetTPStat                   = "ApierV1.GetTPStat"
	ApierV1RemoveTPStat                = "ApierV1.RemoveTPStat"
	ApierV1GetTPDestinationRate        = "ApierV1.GetTPDestinationRate"
	ApierV1SetTPSupplierProfile        = "ApierV1.SetTPSupplierProfile"
	ApierV1GetTPSupplierProfile        = "ApierV1.GetTPSupplierProfile"
	ApierV1GetTPSupplierProfileIDs     = "ApierV1.GetTPSupplierProfileIDs"
	ApierV1RemoveTPSupplierProfile     = "ApierV1.RemoveTPSupplierProfile"
	ApierV1GetTPDispatcherProfile      = "ApierV1.GetTPDispatcherProfile"
	ApierV1SetTPDispatcherProfile      = "ApierV1.SetTPDispatcherProfile"
	ApierV1RemoveTPDispatcherProfile   = "ApierV1.RemoveTPDispatcherProfile"
	ApierV1GetTPDispatcherProfileIDs   = "ApierV1.GetTPDispatcherProfileIDs"
	ApierV1GetTPSharedGroups           = "ApierV1.GetTPSharedGroups"
	ApierV1SetTPSharedGroups           = "ApierV1.SetTPSharedGroups"
	ApierV1GetTPSharedGroupIds         = "ApierV1.GetTPSharedGroupIds"
	ApierV1RemoveTPSharedGroups        = "ApierV1.RemoveTPSharedGroups"
	ApierV1ExportCDRs                  = "ApierV1.ExportCDRs"
	ApierV1GetTPRatingPlan             = "ApierV1.GetTPRatingPlan"
	ApierV1SetTPRatingPlan             = "ApierV1.SetTPRatingPlan"
	ApierV1GetTPRatingPlanIds          = "ApierV1.GetTPRatingPlanIds"
	ApierV1RemoveTPRatingPlan          = "ApierV1.RemoveTPRatingPlan"
	ApierV1SetTPActions                = "ApierV1.SetTPActions"
	ApierV1GetTPActionIds              = "ApierV1.GetTPActionIds"
	ApierV1RemoveTPActions             = "ApierV1.RemoveTPActions"
	ApierV1SetActionPlan               = "ApierV1.SetActionPlan"
	ApierV1ExecuteAction               = "ApierV1.ExecuteAction"
	ApierV1SetTPRatingProfile          = "ApierV1.SetTPRatingProfile"
	ApierV1GetTPRatingProfile          = "ApierV1.GetTPRatingProfile"
	ApierV1RemoveTPRatingProfile       = "ApierV1.RemoveTPRatingProfile"
	ApierV1SetTPDestinationRate        = "ApierV1.SetTPDestinationRate"
	ApierV1GetTPRatingProfileLoadIds   = "ApierV1.GetTPRatingProfileLoadIds"
	ApierV1GetTPRatingProfilesByLoadID = "ApierV1.GetTPRatingProfilesByLoadID"
	ApierV1GetTPRatingProfileIds       = "ApierV1.GetTPRatingProfileIds"
	ApierV1GetTPDestinationRateIds     = "ApierV1.GetTPDestinationRateIds"
	ApierV1RemoveTPDestinationRate     = "ApierV1.RemoveTPDestinationRate"
	ApierV1ImportTariffPlanFromFolder  = "ApierV1.ImportTariffPlanFromFolder"
	ApierV1ExportTPToFolder            = "ApierV1.ExportTPToFolder"
	ApierV1LoadRatingPlan              = "ApierV1.LoadRatingPlan"
	ApierV1LoadRatingProfile           = "ApierV1.LoadRatingProfile"
	ApierV1LoadAccountActions          = "ApierV1.LoadAccountActions"
	ApierV1SetActions                  = "ApierV1.SetActions"
	ApierV1AddTriggeredAction          = "ApierV1.AddTriggeredAction"
	ApierV1GetAccountActionTriggers    = "ApierV1.GetAccountActionTriggers"
	ApierV1AddAccountActionTriggers    = "ApierV1.AddAccountActionTriggers"
	ApierV1ResetAccountActionTriggers  = "ApierV1.ResetAccountActionTriggers"
	ApierV1SetAccountActionTriggers    = "ApierV1.SetAccountActionTriggers"
	ApierV1RemoveAccountActionTriggers = "ApierV1.RemoveAccountActionTriggers"
	ApierV1GetScheduledActions         = "ApierV1.GetScheduledActions"
	ApierV1RemoveActionTiming          = "ApierV1.RemoveActionTiming"
	ApierV1ComputeReverseDestinations  = "ApierV1.ComputeReverseDestinations"
	ApierV1ComputeAccountActionPlans   = "ApierV1.ComputeAccountActionPlans"
	ApierV1SetDestination              = "ApierV1.SetDestination"
	ApierV1GetDataCost                 = "ApierV1.GetDataCost"
	ApierV1ReplayFailedPosts           = "ApierV1.ReplayFailedPosts"
	ApierV1RemoveAccount               = "ApierV1.RemoveAccount"
	ApierV1DebitUsage                  = "ApierV1.DebitUsage"
	ApierV1GetCacheStats               = "ApierV1.GetCacheStats"
	ApierV1ReloadCache                 = "ApierV1.ReloadCache"
	ApierV1GetActionTriggers           = "ApierV1.GetActionTriggers"
	ApierV1SetActionTrigger            = "ApierV1.SetActionTrigger"
	ApierV1RemoveActionPlan            = "ApierV1.RemoveActionPlan"
	ApierV1RemoveActions               = "ApierV1.RemoveActions"
	ApierV1RemoveBalances              = "ApierV1.RemoveBalances"
	ApierV1ReloadCdrcConfig            = "ApierV1.ReloadCdrcConfig"
	ApierV1ReloadCdreConfig            = "ApierV1.ReloadCdreConfig"
	ApierV1GetLoadHistory              = "ApierV1.GetLoadHistory"
	ApierV1GetLoadIDs                  = "ApierV1.GetLoadIDs"
	ApierV1GetLoadTimes                = "ApierV1.GetLoadTimes"
	ApierV1ExecuteScheduledActions     = "ApierV1.ExecuteScheduledActions"
	ApierV1GetSharedGroup              = "ApierV1.GetSharedGroup"
	ApierV1RemoveActionTrigger         = "ApierV1.RemoveActionTrigger"
	ApierV1GetAccount                  = "ApierV1.GetAccount"
)

// ApierV1 TP APIs
const (
	ApierV1SetTPTiming              = "ApierV1.SetTPTiming"
	ApierV1GetTPTiming              = "ApierV1.GetTPTiming"
	ApierV1RemoveTPTiming           = "ApierV1.RemoveTPTiming"
	ApierV1GetTPTimingIds           = "ApierV1.GetTPTimingIds"
	ApierV1LoadTariffPlanFromStorDb = "ApierV1.LoadTariffPlanFromStorDb"
	ApierV1RemoveTPFromFolder       = "ApierV1.RemoveTPFromFolder"
)

// ApierV2 APIs
const (
	ApierV2                           = "ApierV2"
	ApierV2LoadTariffPlanFromFolder   = "ApierV2.LoadTariffPlanFromFolder"
	ApierV2GetCDRs                    = "ApierV2.GetCDRs"
	ApierV2GetAccount                 = "ApierV2.GetAccount"
	ApierV2GetAccounts                = "ApierV2.GetAccounts"
	ApierV2SetAccount                 = "ApierV2.SetAccount"
	ApierV2CountCDRs                  = "ApierV2.CountCDRs"
	ApierV2SetBalance                 = "ApierV2.SetBalance"
	ApierV2SetActions                 = "ApierV2.SetActions"
	ApierV2RemoveTPTiming             = "ApierV2.RemoveTPTiming"
	ApierV2GetTPDestination           = "ApierV2.GetTPDestination"
	ApierV2SetTPDestination           = "ApierV2.SetTPDestination"
	ApierV2RemoveTPDestination        = "ApierV2.RemoveTPDestination"
	ApierV2GetTPDestinationIDs        = "ApierV2.GetTPDestinationIDs"
	ApierV2GetTPTiming                = "ApierV2.GetTPTiming"
	ApierV2SetTPTiming                = "ApierV2.SetTPTiming"
	ApierV2SetAccountActionTriggers   = "ApierV2.SetAccountActionTriggers"
	ApierV2GetAccountActionTriggers   = "ApierV2.GetAccountActionTriggers"
	ApierV2SetActionPlan              = "ApierV2.SetActionPlan"
	ApierV2GetActions                 = "ApierV2.GetActions"
	ApierV2GetDestinations            = "ApierV2.GetDestinations"
	ApierV2GetCacheStats              = "ApierV2.GetCacheStats"
	ApierV2ExecuteAction              = "ApierV2.ExecuteAction"
	ApierV2ResetAccountActionTriggers = "ApierV2.ResetAccountActionTriggers"
	ApierV2RemoveActions              = "ApierV2.RemoveActions"
)

const (
	ServiceManagerV1              = "ServiceManagerV1"
	ServiceManagerV1StartService  = "ServiceManagerV1.StartService"
	ServiceManagerV1StopService   = "ServiceManagerV1.StopService"
	ServiceManagerV1ServiceStatus = "ServiceManagerV1.ServiceStatus"
	ServiceManagerV1Ping          = "ServiceManagerV1.Ping"
)

const (
	ConfigSv1                     = "ConfigSv1"
	ConfigSv1GetJSONSection       = "ConfigSv1.GetJSONSection"
	ConfigSv1ReloadConfigFromPath = "ConfigSv1.ReloadConfigFromPath"
	ConfigSv1ReloadConfigFromJSON = "ConfigSv1.ReloadConfigFromJSON"
)

const (
	RALsV1                   = "RALsV1"
	RALsV1GetRatingPlansCost = "RALsV1.GetRatingPlansCost"
	RALsV1Ping               = "RALsV1.Ping"
)

const (
	CoreS         = "CoreS"
	CoreSv1       = "CoreSv1"
	CoreSv1Status = "CoreSv1.Status"
	CoreSv1Ping   = "CoreSv1.Ping"
)

// SupplierS APIs
const (
	SupplierSv1GetSuppliers                = "SupplierSv1.GetSuppliers"
	SupplierSv1GetSupplierProfilesForEvent = "SupplierSv1.GetSupplierProfilesForEvent"
	SupplierSv1Ping                        = "SupplierSv1.Ping"
	ApierV1GetSupplierProfile              = "ApierV1.GetSupplierProfile"
	ApierV1GetSupplierProfileIDs           = "ApierV1.GetSupplierProfileIDs"
	ApierV1RemoveSupplierProfile           = "ApierV1.RemoveSupplierProfile"
	ApierV1SetSupplierProfile              = "ApierV1.SetSupplierProfile"
)

// AttributeS APIs
const (
	ApierV1SetAttributeProfile       = "ApierV1.SetAttributeProfile"
	ApierV1GetAttributeProfile       = "ApierV1.GetAttributeProfile"
	ApierV1GetAttributeProfileIDs    = "ApierV1.GetAttributeProfileIDs"
	ApierV1RemoveAttributeProfile    = "ApierV1.RemoveAttributeProfile"
	ApierV2SetAttributeProfile       = "ApierV2.SetAttributeProfile"
	AttributeSv1GetAttributeForEvent = "AttributeSv1.GetAttributeForEvent"
	AttributeSv1ProcessEvent         = "AttributeSv1.ProcessEvent"
	AttributeSv1Ping                 = "AttributeSv1.Ping"
)

// ChargerS APIs
const (
	ChargerSv1Ping                = "ChargerSv1.Ping"
	ChargerSv1GetChargersForEvent = "ChargerSv1.GetChargersForEvent"
	ChargerSv1ProcessEvent        = "ChargerSv1.ProcessEvent"
	ApierV1GetChargerProfile      = "ApierV1.GetChargerProfile"
	ApierV1RemoveChargerProfile   = "ApierV1.RemoveChargerProfile"
	ApierV1SetChargerProfile      = "ApierV1.SetChargerProfile"
	ApierV1GetChargerProfileIDs   = "ApierV1.GetChargerProfileIDs"
)

// ThresholdS APIs
const (
	ThresholdSv1ProcessEvent          = "ThresholdSv1.ProcessEvent"
	ThresholdSv1GetThreshold          = "ThresholdSv1.GetThreshold"
	ThresholdSv1GetThresholdIDs       = "ThresholdSv1.GetThresholdIDs"
	ThresholdSv1Ping                  = "ThresholdSv1.Ping"
	ThresholdSv1GetThresholdsForEvent = "ThresholdSv1.GetThresholdsForEvent"
	ApierV1GetThresholdProfileIDs     = "ApierV1.GetThresholdProfileIDs"
	ApierV1GetThresholdProfile        = "ApierV1.GetThresholdProfile"
	ApierV1RemoveThresholdProfile     = "ApierV1.RemoveThresholdProfile"
	ApierV1SetThresholdProfile        = "ApierV1.SetThresholdProfile"
)

// StatS APIs
const (
	StatSv1ProcessEvent           = "StatSv1.ProcessEvent"
	StatSv1GetQueueIDs            = "StatSv1.GetQueueIDs"
	StatSv1GetQueueStringMetrics  = "StatSv1.GetQueueStringMetrics"
	StatSv1GetQueueFloatMetrics   = "StatSv1.GetQueueFloatMetrics"
	StatSv1Ping                   = "StatSv1.Ping"
	StatSv1GetStatQueuesForEvent  = "StatSv1.GetStatQueuesForEvent"
	ApierV1GetStatQueueProfile    = "ApierV1.GetStatQueueProfile"
	ApierV1RemoveStatQueueProfile = "ApierV1.RemoveStatQueueProfile"
	ApierV1SetStatQueueProfile    = "ApierV1.SetStatQueueProfile"
	ApierV1GetStatQueueProfileIDs = "ApierV1.GetStatQueueProfileIDs"
)

// ResourceS APIs
const (
	ResourceSv1AuthorizeResources   = "ResourceSv1.AuthorizeResources"
	ResourceSv1GetResourcesForEvent = "ResourceSv1.GetResourcesForEvent"
	ResourceSv1AllocateResources    = "ResourceSv1.AllocateResources"
	ResourceSv1ReleaseResources     = "ResourceSv1.ReleaseResources"
	ResourceSv1Ping                 = "ResourceSv1.Ping"
	ResourceSv1GetResource          = "ResourceSv1.GetResource"
	ApierV1SetResourceProfile       = "ApierV1.SetResourceProfile"
	ApierV1RemoveResourceProfile    = "ApierV1.RemoveResourceProfile"
	ApierV1GetResourceProfile       = "ApierV1.GetResourceProfile"
	ApierV1GetResourceProfileIDs    = "ApierV1.GetResourceProfileIDs"
)

// SessionS APIs
const (
	SessionSv1AuthorizeEvent             = "SessionSv1.AuthorizeEvent"
	SessionSv1AuthorizeEventWithDigest   = "SessionSv1.AuthorizeEventWithDigest"
	SessionSv1InitiateSession            = "SessionSv1.InitiateSession"
	SessionSv1InitiateSessionWithDigest  = "SessionSv1.InitiateSessionWithDigest"
	SessionSv1UpdateSession              = "SessionSv1.UpdateSession"
	SessionSv1SyncSessions               = "SessionSv1.SyncSessions"
	SessionSv1TerminateSession           = "SessionSv1.TerminateSession"
	SessionSv1ProcessCDR                 = "SessionSv1.ProcessCDR"
	SessionSv1ProcessMessage             = "SessionSv1.ProcessMessage"
	SessionSv1ProcessEvent               = "SessionSv1.ProcessEvent"
	SessionSv1DisconnectSession          = "SessionSv1.DisconnectSession"
	SessionSv1GetActiveSessions          = "SessionSv1.GetActiveSessions"
	SessionSv1GetActiveSessionsCount     = "SessionSv1.GetActiveSessionsCount"
	SessionSv1ForceDisconnect            = "SessionSv1.ForceDisconnect"
	SessionSv1GetPassiveSessions         = "SessionSv1.GetPassiveSessions"
	SessionSv1GetPassiveSessionsCount    = "SessionSv1.GetPassiveSessionsCount"
	SessionSv1SetPassiveSession          = "SessionSv1.SetPassiveSession"
	SessionSv1Ping                       = "SessionSv1.Ping"
	SessionSv1GetActiveSessionIDs        = "SessionSv1.GetActiveSessionIDs"
	SessionSv1RegisterInternalBiJSONConn = "SessionSv1.RegisterInternalBiJSONConn"
	SessionSv1ReplicateSessions          = "SessionSv1.ReplicateSessions"
	SessionSv1ActivateSessions           = "SessionSv1.ActivateSessions"
	SessionSv1DeactivateSessions         = "SessionSv1.DeactivateSessions"
	SMGenericV1InitiateSession           = "SMGenericV1.InitiateSession"
)

// Responder APIs
const (
	Responder                  = "Responder"
	ResponderDebit             = "Responder.Debit"
	ResponderRefundIncrements  = "Responder.RefundIncrements"
	ResponderGetMaxSessionTime = "Responder.GetMaxSessionTime"
	ResponderMaxDebit          = "Responder.MaxDebit"
	ResponderRefundRounding    = "Responder.RefundRounding"
	ResponderGetCost           = "Responder.GetCost"
	ResponderShutdown          = "Responder.Shutdown"
	ResponderPing              = "Responder.Ping"
)

// DispatcherS APIs
const (
	DispatcherSv1Ping               = "DispatcherSv1.Ping"
	DispatcherSv1GetProfileForEvent = "DispatcherSv1.GetProfileForEvent"
	DispatcherSv1Apier              = "DispatcherSv1.Apier"
	DispatcherServicePing           = "DispatcherService.Ping"
)

// AnalyzerS APIs
const (
	AnalyzerSv1     = "AnalyzerSv1"
	AnalyzerSv1Ping = "AnalyzerSv1.Ping"
)

// LoaderS APIs
const (
	LoaderSv1       = "LoaderSv1"
	LoaderSv1Load   = "LoaderSv1.Load"
	LoaderSv1Remove = "LoaderSv1.Remove"
	LoaderSv1Ping   = "LoaderSv1.Ping"
)

// CacheS APIs
const (
	CacheSv1                  = "CacheSv1"
	CacheSv1GetCacheStats     = "CacheSv1.GetCacheStats"
	CacheSv1GetItemIDs        = "CacheSv1.GetItemIDs"
	CacheSv1HasItem           = "CacheSv1.HasItem"
	CacheSv1GetItemExpiryTime = "CacheSv1.GetItemExpiryTime"
	CacheSv1RemoveItem        = "CacheSv1.RemoveItem"
	CacheSv1PrecacheStatus    = "CacheSv1.PrecacheStatus"
	CacheSv1HasGroup          = "CacheSv1.HasGroup"
	CacheSv1GetGroupItemIDs   = "CacheSv1.GetGroupItemIDs"
	CacheSv1RemoveGroup       = "CacheSv1.RemoveGroup"
	CacheSv1Clear             = "CacheSv1.Clear"
	CacheSv1ReloadCache       = "CacheSv1.ReloadCache"
	CacheSv1LoadCache         = "CacheSv1.LoadCache"
	CacheSv1FlushCache        = "CacheSv1.FlushCache"
	CacheSv1Ping              = "CacheSv1.Ping"
)

// GuardianS APIs
const (
	GuardianSv1             = "GuardianSv1"
	GuardianSv1RemoteLock   = "GuardianSv1.RemoteLock"
	GuardianSv1RemoteUnlock = "GuardianSv1.RemoteUnlock"
	GuardianSv1Ping         = "GuardianSv1.Ping"
)

// Cdrs APIs
const (
	CDRsV1                   = "CDRsV1"
	CDRsV1GetCDRsCount       = "CDRsV1.GetCDRsCount"
	CDRsV1RateCDRs           = "CDRsV1.RateCDRs"
	CDRsV1GetCDRs            = "CDRsV1.GetCDRs"
	CDRsV1ProcessCDR         = "CDRsV1.ProcessCDR"
	CDRsV1ProcessExternalCDR = "CDRsV1.ProcessExternalCDR"
	CDRsV1StoreSessionCost   = "CDRsV1.StoreSessionCost"
	CDRsV1ProcessEvent       = "CDRsV1.ProcessEvent"
	CDRsV1Ping               = "CDRsV1.Ping"
	CDRsV2                   = "CDRsV2"
	CDRsV2StoreSessionCost   = "CDRsV2.StoreSessionCost"
	CdrsV2ProcessExternalCdr = "CdrsV2.ProcessExternalCdr"
	CdrsV2ProcessCdr         = "CdrsV2.ProcessCdr"
)

// Scheduler
const (
	SchedulerSv1       = "SchedulerSv1"
	SchedulerSv1Ping   = "SchedulerSv1.Ping"
	SchedulerSv1Reload = "SchedulerSv1.Reload"
)

//cgr_ variables
const (
	CGR_ACCOUNT          = "cgr_account"
	CGR_SUPPLIER         = "cgr_supplier"
	CGR_DESTINATION      = "cgr_destination"
	CGR_SUBJECT          = "cgr_subject"
	CGR_CATEGORY         = "cgr_category"
	CGR_REQTYPE          = "cgr_reqtype"
	CGR_TENANT           = "cgr_tenant"
	CGR_PDD              = "cgr_pdd"
	CGR_DISCONNECT_CAUSE = "cgr_disconnectcause"
	CGR_COMPUTELCR       = "cgr_computelcr"
	CGR_SUPPLIERS        = "cgr_suppliers"
	CGRFlags             = "cgr_flags"
)

//CSV file name
const (
	TimingsCsv            = "Timings.csv"
	DestinationsCsv       = "Destinations.csv"
	RatesCsv              = "Rates.csv"
	DestinationRatesCsv   = "DestinationRates.csv"
	RatingPlansCsv        = "RatingPlans.csv"
	RatingProfilesCsv     = "RatingProfiles.csv"
	SharedGroupsCsv       = "SharedGroups.csv"
	ActionsCsv            = "Actions.csv"
	ActionPlansCsv        = "ActionPlans.csv"
	ActionTriggersCsv     = "ActionTriggers.csv"
	AccountActionsCsv     = "AccountActions.csv"
	ResourcesCsv          = "Resources.csv"
	StatsCsv              = "Stats.csv"
	ThresholdsCsv         = "Thresholds.csv"
	FiltersCsv            = "Filters.csv"
	SuppliersCsv          = "Suppliers.csv"
	AttributesCsv         = "Attributes.csv"
	ChargersCsv           = "Chargers.csv"
	DispatcherProfilesCsv = "DispatcherProfiles.csv"
	DispatcherHostsCsv    = "DispatcherHosts.csv"
)

// Table Name
const (
	TBLTPTimings          = "tp_timings"
	TBLTPDestinations     = "tp_destinations"
	TBLTPRates            = "tp_rates"
	TBLTPDestinationRates = "tp_destination_rates"
	TBLTPRatingPlans      = "tp_rating_plans"
	TBLTPRateProfiles     = "tp_rating_profiles"
	TBLTPSharedGroups     = "tp_shared_groups"
	TBLTPActions          = "tp_actions"
	TBLTPActionPlans      = "tp_action_plans"
	TBLTPActionTriggers   = "tp_action_triggers"
	TBLTPAccountActions   = "tp_account_actions"
	TBLTPResources        = "tp_resources"
	TBLTPStats            = "tp_stats"
	TBLTPThresholds       = "tp_thresholds"
	TBLTPFilters          = "tp_filters"
	SessionCostsTBL       = "session_costs"
	CDRsTBL               = "cdrs"
	TBLTPSuppliers        = "tp_suppliers"
	TBLTPAttributes       = "tp_attributes"
	TBLTPChargers         = "tp_chargers"
	TBLVersions           = "versions"
	OldSMCosts            = "sm_costs"
	TBLTPDispatchers      = "tp_dispatcher_profiles"
	TBLTPDispatcherHosts  = "tp_dispatcher_hosts"
)

// Cache Name
const (
	CacheDestinations            = "*destinations"
	CacheReverseDestinations     = "*reverse_destinations"
	CacheRatingPlans             = "*rating_plans"
	CacheRatingProfiles          = "*rating_profiles"
	CacheActions                 = "*actions"
	CacheActionPlans             = "*action_plans"
	CacheAccountActionPlans      = "*account_action_plans"
	CacheActionTriggers          = "*action_triggers"
	CacheSharedGroups            = "*shared_groups"
	CacheResources               = "*resources"
	CacheResourceProfiles        = "*resource_profiles"
	CacheTimings                 = "*timings"
	CacheEventResources          = "*event_resources"
	CacheStatQueueProfiles       = "*statqueue_profiles"
	CacheStatQueues              = "*statqueues"
	CacheThresholdProfiles       = "*threshold_profiles"
	CacheThresholds              = "*thresholds"
	CacheFilters                 = "*filters"
	CacheSupplierProfiles        = "*supplier_profiles"
	CacheAttributeProfiles       = "*attribute_profiles"
	CacheChargerProfiles         = "*charger_profiles"
	CacheDispatcherProfiles      = "*dispatcher_profiles"
	CacheDispatcherHosts         = "*dispatcher_hosts"
	CacheDispatchers             = "*dispatchers"
	CacheDispatcherRoutes        = "*dispatcher_routes"
	CacheResourceFilterIndexes   = "*resource_filter_indexes"
	CacheStatFilterIndexes       = "*stat_filter_indexes"
	CacheThresholdFilterIndexes  = "*threshold_filter_indexes"
	CacheSupplierFilterIndexes   = "*supplier_filter_indexes"
	CacheAttributeFilterIndexes  = "*attribute_filter_indexes"
	CacheChargerFilterIndexes    = "*charger_filter_indexes"
	CacheDispatcherFilterIndexes = "*dispatcher_filter_indexes"
	CacheDiameterMessages        = "*diameter_messages"
	CacheRPCResponses            = "*rpc_responses"
	CacheClosedSessions          = "*closed_sessions"
	MetaPrecaching               = "*precaching"
	MetaReady                    = "*ready"
	CacheLoadIDs                 = "*load_ids"
	CacheAccounts                = "*accounts"
	CacheRPCConnections          = "*rpc_connections"
	CacheCDRIDs                  = "*cdr_ids"
	CacheRatingProfilesTmp       = "*tmp_rating_profiles"
)

// Prefix for indexing
const (
	ResourceFilterIndexes   = "rfi_"
	StatFilterIndexes       = "sfi_"
	ThresholdFilterIndexes  = "tfi_"
	SupplierFilterIndexes   = "spi_"
	AttributeFilterIndexes  = "afi_"
	ChargerFilterIndexes    = "cfi_"
	DispatcherFilterIndexes = "dfi_"
	ActionPlanIndexes       = "api_"
)

// Agents
const (
	KamailioAgent   = "KamailioAgent"
	RadiusAgent     = "RadiusAgent"
	DiameterAgent   = "DiameterAgent"
	FreeSWITCHAgent = "FreeSWITCHAgent"
	AsteriskAgent   = "AsteriskAgent"
	HTTPAgent       = "HTTPAgent"
)

// Poster
const (
	SQSPoster    = "SQSPoster"
	S3Poster     = "S3Poster"
	AWSRegion    = "aws_region"
	AWSKey       = "aws_key"
	AWSSecret    = "aws_secret"
	KafkaTopic   = "topic"
	KafkaGroupID = "group_id"
	KafkaMaxWait = "max_wait"
)

// Google_API
const (
	MetaGoogleAPI             = "*gapi"
	GoogleConfigDirName       = ".gapi"
	GoogleCredentialsFileName = "credentials.json"
	GoogleTokenFileName       = "token.json"
)

// StorDB
var (
	PostgressSSLModeDisable    = "disable"
	PostgressSSLModeAllow      = "allow"
	PostgressSSLModePrefer     = "prefer"
	PostgressSSLModeRequire    = "require"
	PostgressSSLModeVerifyCa   = "verify-ca"
	PostgressSSLModeVerifyFull = "verify-full"
)

func buildCacheInstRevPrefixes() {
	CachePrefixToInstance = make(map[string]string)
	for k, v := range CacheInstanceToPrefix {
		CachePrefixToInstance[v] = k
	}
}

func buildCacheIndexesToPrefix() {
	CacheIndexesToPrefix = make(map[string]string)
	for k, v := range PrefixToIndexCache {
		CacheIndexesToPrefix[v] = k
	}
}

func init() {
	buildCacheInstRevPrefixes()
	buildCacheIndexesToPrefix()
}
