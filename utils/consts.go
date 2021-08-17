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
	VERSION                      = "v0.10.3~dev"
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
	MetaSIPCID                  = "*sipcid"
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
	SessionTTLLastUsage         = "SessionTTLLastUsage"
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
	StartTime                   = "StartTime"
	AccountSummary              = "AccountSummary"
	RatingFilters               = "RatingFilters"
	RatingFilter                = "RatingFilter"
	Accounting                  = "Accounting"
	Rating                      = "Rating"
	Charges                     = "Charges"
	CompressFactor              = "CompressFactor"
	Increments                  = "Increments"
	Balance                     = "Balance"
	BalanceSummaries            = "BalanceSummaries"
	Type                        = "Type"
	YearsFieldName              = "Years"
	MonthsFieldName             = "Months"
	MonthDaysFieldName          = "MonthDays"
	WeekDaysFieldName           = "WeekDays"
	GroupIntervalStart          = "GroupIntervalStart"
	RateIncrement               = "RateIncrement"
	RateUnit                    = "RateUnit"
	BalanceUUID                 = "BalanceUUID"
	RatingID                    = "RatingID"
	ExtraChargeID               = "ExtraChargeID"
	ConnectFee                  = "ConnectFee"
	RoundingMethod              = "RoundingMethod"
	RoundingDecimals            = "RoundingDecimals"
	MaxCostStrategy             = "MaxCostStrategy"
	TimingID                    = "TimingID"
	RatesID                     = "RatesID"
	RatingFiltersID             = "RatingFiltersID"
	AccountingID                = "AccountingID"
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
	MetaSuppliersMaxCost        = "*suppliers_maxcost"
	MetaMaxCost                 = "*maxcost"
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
	MetaTmp                     = "*tmp"
	CGROriginHost               = "cgr_originhost"
	MetaInitiate                = "*initiate"
	MetaFD                      = "*fd"
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
	MetaIP2Hex                  = "*ip2hex"
	MetaString2Hex              = "*string2hex"
	MetaReload                  = "*reload"
	MetaLoad                    = "*load"
	MetaRemove                  = "*remove"
	MetaRemoveAll               = "*removeall"
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
	MetaDiamreq                 = "*diamreq"
	MetaGroup                   = "*group"
	InternalRPCSet              = "InternalRPCSet"
	FileName                    = "FileName"
	MetaBusy                    = "*busy"
	MetaQueue                   = "*queue"
	MetaRounding                = "*rounding"
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

	MetaEC = "*ec"
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

// APIerSv1 APIs
const (
	ApierV1                                   = "ApierV1"
	ApierV2                                   = "ApierV2"
	APIerSv1                                  = "APIerSv1"
	APIerSv1ComputeFilterIndexes              = "APIerSv1.ComputeFilterIndexes"
	APIerSv1ComputeFilterIndexIDs             = "APIerSv1.ComputeFilterIndexIDs"
	APIerSv1Ping                              = "APIerSv1.Ping"
	APIerSv1SetDispatcherProfile              = "APIerSv1.SetDispatcherProfile"
	APIerSv1GetDispatcherProfile              = "APIerSv1.GetDispatcherProfile"
	APIerSv1GetDispatcherProfileIDs           = "APIerSv1.GetDispatcherProfileIDs"
	APIerSv1RemoveDispatcherProfile           = "APIerSv1.RemoveDispatcherProfile"
	APIerSv1SetDispatcherHost                 = "APIerSv1.SetDispatcherHost"
	APIerSv1GetDispatcherHost                 = "APIerSv1.GetDispatcherHost"
	APIerSv1GetDispatcherHostIDs              = "APIerSv1.GetDispatcherHostIDs"
	APIerSv1RemoveDispatcherHost              = "APIerSv1.RemoveDispatcherHost"
	APIerSv1GetEventCost                      = "APIerSv1.GetEventCost"
	APIerSv1LoadTariffPlanFromFolder          = "APIerSv1.LoadTariffPlanFromFolder"
	APIerSv1GetCost                           = "APIerSv1.GetCost"
	APIerSv1SetBalance                        = "APIerSv1.SetBalance"
	APIerSv1GetFilter                         = "APIerSv1.GetFilter"
	APIerSv1GetFilterIndexes                  = "APIerSv1.GetFilterIndexes"
	APIerSv1RemoveFilterIndexes               = "APIerSv1.RemoveFilterIndexes"
	APIerSv1RemoveFilter                      = "APIerSv1.RemoveFilter"
	APIerSv1SetFilter                         = "APIerSv1.SetFilter"
	APIerSv1GetFilterIDs                      = "APIerSv1.GetFilterIDs"
	APIerSv1GetAccountActionPlansIndexHealth  = "APIerSv1.GetAccountActionPlansIndexHealth"
	APIerSv1GetReverseDestinationsIndexHealth = "APIerSv1.GetReverseDestinationsIndexHealth"
	APIerSv1GetThresholdsIndexesHealth        = "APIerSv1.GetThresholdsIndexesHealth"
	APIerSv1GetResourcesIndexesHealth         = "APIerSv1.GetResourcesIndexesHealth"
	APIerSv1GetStatsIndexesHealth             = "APIerSv1.GetStatsIndexesHealth"
	APIerSv1GetSuppliersIndexesHealth         = "APIerSv1.GetSuppliersIndexesHealth"
	APIerSv1GetChargersIndexesHealth          = "APIerSv1.GetChargersIndexesHealth"
	APIerSv1GetAttributesIndexesHealth        = "APIerSv1.GetAttributesIndexesHealth"
	APIerSv1GetDispatchersIndexesHealth       = "APIerSv1.GetDispatchersIndexesHealth"
	APIerSv1GetRatingProfile                  = "APIerSv1.GetRatingProfile"
	APIerSv1RemoveRatingProfile               = "APIerSv1.RemoveRatingProfile"
	APIerSv1SetRatingProfile                  = "APIerSv1.SetRatingProfile"
	APIerSv1GetRatingProfileIDs               = "APIerSv1.GetRatingProfileIDs"
	APIerSv1SetDataDBVersions                 = "APIerSv1.SetDataDBVersions"
	APIerSv1SetStorDBVersions                 = "APIerSv1.SetStorDBVersions"
	APIerSv1GetAccountActionPlan              = "APIerSv1.GetAccountActionPlan"
	APIerSv1ComputeActionPlanIndexes          = "APIerSv1.ComputeActionPlanIndexes"
	APIerSv1GetActions                        = "APIerSv1.GetActions"
	APIerSv1GetActionPlan                     = "APIerSv1.GetActionPlan"
	APIerSv1GetActionPlanIDs                  = "APIerSv1.GetActionPlanIDs"
	APIerSv1GetRatingPlanIDs                  = "APIerSv1.GetRatingPlanIDs"
	APIerSv1GetRatingPlan                     = "APIerSv1.GetRatingPlan"
	APIerSv1RemoveRatingPlan                  = "APIerSv1.RemoveRatingPlan"
	APIerSv1GetDestination                    = "APIerSv1.GetDestination"
	APIerSv1GetReverseDestination             = "APIerSv1.GetReverseDestination"
	APIerSv1AddBalance                        = "APIerSv1.AddBalance"
	APIerSv1DebitBalance                      = "APIerSv1.DebitBalance"
	APIerSv1SetAccount                        = "APIerSv1.SetAccount"
	APIerSv1GetAccountsCount                  = "APIerSv1.GetAccountsCount"
	APIerSv1GetDataDBVersions                 = "APIerSv1.GetDataDBVersions"
	APIerSv1GetStorDBVersions                 = "APIerSv1.GetStorDBVersions"
	APIerSv1GetCDRs                           = "APIerSv1.GetCDRs"
	APIerSv1GetTPAccountActions               = "APIerSv1.GetTPAccountActions"
	APIerSv1SetTPAccountActions               = "APIerSv1.SetTPAccountActions"
	APIerSv1GetTPAccountActionsByLoadId       = "APIerSv1.GetTPAccountActionsByLoadId"
	APIerSv1GetTPAccountActionLoadIds         = "APIerSv1.GetTPAccountActionLoadIds"
	APIerSv1GetTPAccountActionIds             = "APIerSv1.GetTPAccountActionIds"
	APIerSv1RemoveTPAccountActions            = "APIerSv1.RemoveTPAccountActions"
	APIerSv1GetTPActionPlan                   = "APIerSv1.GetTPActionPlan"
	APIerSv1SetTPActionPlan                   = "APIerSv1.SetTPActionPlan"
	APIerSv1GetTPActionPlanIds                = "APIerSv1.GetTPActionPlanIds"
	APIerSv1SetTPActionTriggers               = "APIerSv1.SetTPActionTriggers"
	APIerSv1GetTPActionTriggers               = "APIerSv1.GetTPActionTriggers"
	APIerSv1RemoveTPActionTriggers            = "APIerSv1.RemoveTPActionTriggers"
	APIerSv1GetTPActionTriggerIds             = "APIerSv1.GetTPActionTriggerIds"
	APIerSv1GetTPActions                      = "APIerSv1.GetTPActions"
	APIerSv1RemoveTPActionPlan                = "APIerSv1.RemoveTPActionPlan"
	APIerSv1GetTPAttributeProfile             = "APIerSv1.GetTPAttributeProfile"
	APIerSv1SetTPAttributeProfile             = "APIerSv1.SetTPAttributeProfile"
	APIerSv1GetTPAttributeProfileIds          = "APIerSv1.GetTPAttributeProfileIds"
	APIerSv1RemoveTPAttributeProfile          = "APIerSv1.RemoveTPAttributeProfile"
	APIerSv1GetTPCharger                      = "APIerSv1.GetTPCharger"
	APIerSv1SetTPCharger                      = "APIerSv1.SetTPCharger"
	APIerSv1RemoveTPCharger                   = "APIerSv1.RemoveTPCharger"
	APIerSv1GetTPChargerIDs                   = "APIerSv1.GetTPChargerIDs"
	APIerSv1SetTPFilterProfile                = "APIerSv1.SetTPFilterProfile"
	APIerSv1GetTPFilterProfile                = "APIerSv1.GetTPFilterProfile"
	APIerSv1GetTPFilterProfileIds             = "APIerSv1.GetTPFilterProfileIds"
	APIerSv1RemoveTPFilterProfile             = "APIerSv1.RemoveTPFilterProfile"
	APIerSv1GetTPDestination                  = "APIerSv1.GetTPDestination"
	APIerSv1SetTPDestination                  = "APIerSv1.SetTPDestination"
	APIerSv1GetTPDestinationIDs               = "APIerSv1.GetTPDestinationIDs"
	APIerSv1RemoveTPDestination               = "APIerSv1.RemoveTPDestination"
	APIerSv1GetTPResource                     = "APIerSv1.GetTPResource"
	APIerSv1SetTPResource                     = "APIerSv1.SetTPResource"
	APIerSv1RemoveTPResource                  = "APIerSv1.RemoveTPResource"
	APIerSv1SetTPRate                         = "APIerSv1.SetTPRate"
	APIerSv1GetTPRate                         = "APIerSv1.GetTPRate"
	APIerSv1RemoveTPRate                      = "APIerSv1.RemoveTPRate"
	APIerSv1GetTPRateIds                      = "APIerSv1.GetTPRateIds"
	APIerSv1SetTPThreshold                    = "APIerSv1.SetTPThreshold"
	APIerSv1GetTPThreshold                    = "APIerSv1.GetTPThreshold"
	APIerSv1GetTPThresholdIDs                 = "APIerSv1.GetTPThresholdIDs"
	APIerSv1RemoveTPThreshold                 = "APIerSv1.RemoveTPThreshold"
	APIerSv1SetTPStat                         = "APIerSv1.SetTPStat"
	APIerSv1GetTPStat                         = "APIerSv1.GetTPStat"
	APIerSv1RemoveTPStat                      = "APIerSv1.RemoveTPStat"
	APIerSv1GetTPDestinationRate              = "APIerSv1.GetTPDestinationRate"
	APIerSv1SetTPSupplierProfile              = "APIerSv1.SetTPSupplierProfile"
	APIerSv1GetTPSupplierProfile              = "APIerSv1.GetTPSupplierProfile"
	APIerSv1GetTPSupplierProfileIDs           = "APIerSv1.GetTPSupplierProfileIDs"
	APIerSv1RemoveTPSupplierProfile           = "APIerSv1.RemoveTPSupplierProfile"
	APIerSv1GetTPDispatcherProfile            = "APIerSv1.GetTPDispatcherProfile"
	APIerSv1SetTPDispatcherProfile            = "APIerSv1.SetTPDispatcherProfile"
	APIerSv1RemoveTPDispatcherProfile         = "APIerSv1.RemoveTPDispatcherProfile"
	APIerSv1GetTPDispatcherProfileIDs         = "APIerSv1.GetTPDispatcherProfileIDs"
	APIerSv1GetTPSharedGroups                 = "APIerSv1.GetTPSharedGroups"
	APIerSv1SetTPSharedGroups                 = "APIerSv1.SetTPSharedGroups"
	APIerSv1GetTPSharedGroupIds               = "APIerSv1.GetTPSharedGroupIds"
	APIerSv1RemoveTPSharedGroups              = "APIerSv1.RemoveTPSharedGroups"
	APIerSv1ExportCDRs                        = "APIerSv1.ExportCDRs"
	APIerSv1GetTPRatingPlan                   = "APIerSv1.GetTPRatingPlan"
	APIerSv1SetTPRatingPlan                   = "APIerSv1.SetTPRatingPlan"
	APIerSv1GetTPRatingPlanIds                = "APIerSv1.GetTPRatingPlanIds"
	APIerSv1RemoveTPRatingPlan                = "APIerSv1.RemoveTPRatingPlan"
	APIerSv1SetTPActions                      = "APIerSv1.SetTPActions"
	APIerSv1GetTPActionIds                    = "APIerSv1.GetTPActionIds"
	APIerSv1RemoveTPActions                   = "APIerSv1.RemoveTPActions"
	APIerSv1SetActionPlan                     = "APIerSv1.SetActionPlan"
	APIerSv1ExecuteAction                     = "APIerSv1.ExecuteAction"
	APIerSv1SetTPRatingProfile                = "APIerSv1.SetTPRatingProfile"
	APIerSv1GetTPRatingProfile                = "APIerSv1.GetTPRatingProfile"
	APIerSv1RemoveTPRatingProfile             = "APIerSv1.RemoveTPRatingProfile"
	APIerSv1SetTPDestinationRate              = "APIerSv1.SetTPDestinationRate"
	APIerSv1GetTPRatingProfileLoadIds         = "APIerSv1.GetTPRatingProfileLoadIds"
	APIerSv1GetTPRatingProfilesByLoadID       = "APIerSv1.GetTPRatingProfilesByLoadID"
	APIerSv1GetTPRatingProfileIds             = "APIerSv1.GetTPRatingProfileIds"
	APIerSv1GetTPDestinationRateIds           = "APIerSv1.GetTPDestinationRateIds"
	APIerSv1RemoveTPDestinationRate           = "APIerSv1.RemoveTPDestinationRate"
	APIerSv1ImportTariffPlanFromFolder        = "APIerSv1.ImportTariffPlanFromFolder"
	APIerSv1ExportTPToFolder                  = "APIerSv1.ExportTPToFolder"
	APIerSv1LoadRatingPlan                    = "APIerSv1.LoadRatingPlan"
	APIerSv1LoadRatingProfile                 = "APIerSv1.LoadRatingProfile"
	APIerSv1LoadAccountActions                = "APIerSv1.LoadAccountActions"
	APIerSv1SetActions                        = "APIerSv1.SetActions"
	APIerSv1AddTriggeredAction                = "APIerSv1.AddTriggeredAction"
	APIerSv1GetAccountActionTriggers          = "APIerSv1.GetAccountActionTriggers"
	APIerSv1AddAccountActionTriggers          = "APIerSv1.AddAccountActionTriggers"
	APIerSv1ResetAccountActionTriggers        = "APIerSv1.ResetAccountActionTriggers"
	APIerSv1SetAccountActionTriggers          = "APIerSv1.SetAccountActionTriggers"
	APIerSv1RemoveAccountActionTriggers       = "APIerSv1.RemoveAccountActionTriggers"
	APIerSv1GetScheduledActions               = "APIerSv1.GetScheduledActions"
	APIerSv1RemoveActionTiming                = "APIerSv1.RemoveActionTiming"
	APIerSv1ComputeReverseDestinations        = "APIerSv1.ComputeReverseDestinations"
	APIerSv1ComputeAccountActionPlans         = "APIerSv1.ComputeAccountActionPlans"
	APIerSv1SetDestination                    = "APIerSv1.SetDestination"
	APIerSv1GetDataCost                       = "APIerSv1.GetDataCost"
	APIerSv1ReplayFailedPosts                 = "APIerSv1.ReplayFailedPosts"
	APIerSv1RemoveAccount                     = "APIerSv1.RemoveAccount"
	APIerSv1DebitUsage                        = "APIerSv1.DebitUsage"
	APIerSv1GetCacheStats                     = "APIerSv1.GetCacheStats"
	APIerSv1ReloadCache                       = "APIerSv1.ReloadCache"
	APIerSv1GetActionTriggers                 = "APIerSv1.GetActionTriggers"
	APIerSv1SetActionTrigger                  = "APIerSv1.SetActionTrigger"
	APIerSv1RemoveActionPlan                  = "APIerSv1.RemoveActionPlan"
	APIerSv1RemoveActions                     = "APIerSv1.RemoveActions"
	APIerSv1RemoveBalances                    = "APIerSv1.RemoveBalances"
	APIerSv1ReloadCdreConfig                  = "APIerSv1.ReloadCdreConfig"
	APIerSv1GetLoadHistory                    = "APIerSv1.GetLoadHistory"
	APIerSv1GetLoadIDs                        = "APIerSv1.GetLoadIDs"
	APIerSv1ExecuteScheduledActions           = "APIerSv1.ExecuteScheduledActions"
	APIerSv1GetLoadTimes                      = "APIerSv1.GetLoadTimes"
	APIerSv1GetSharedGroup                    = "APIerSv1.GetSharedGroup"
	APIerSv1RemoveActionTrigger               = "APIerSv1.RemoveActionTrigger"
	APIerSv1GetAccount                        = "APIerSv1.GetAccount"
)

// APIerSv1 TP APIs
const (
	APIerSv1SetTPTiming              = "APIerSv1.SetTPTiming"
	APIerSv1GetTPTiming              = "APIerSv1.GetTPTiming"
	APIerSv1RemoveTPTiming           = "APIerSv1.RemoveTPTiming"
	APIerSv1GetTPTimingIds           = "APIerSv1.GetTPTimingIds"
	APIerSv1LoadTariffPlanFromStorDb = "APIerSv1.LoadTariffPlanFromStorDb"
	APIerSv1RemoveTPFromFolder       = "APIerSv1.RemoveTPFromFolder"
)

// APIerSv2 APIs
const (
	APIerSv2                           = "APIerSv2"
	APIerSv2LoadTariffPlanFromFolder   = "APIerSv2.LoadTariffPlanFromFolder"
	APIerSv2GetCDRs                    = "APIerSv2.GetCDRs"
	APIerSv2GetAccount                 = "APIerSv2.GetAccount"
	APIerSv2GetAccounts                = "APIerSv2.GetAccounts"
	APIerSv2SetAccount                 = "APIerSv2.SetAccount"
	APIerSv2CountCDRs                  = "APIerSv2.CountCDRs"
	APIerSv2SetBalance                 = "APIerSv2.SetBalance"
	APIerSv2SetActions                 = "APIerSv2.SetActions"
	APIerSv2RemoveTPTiming             = "APIerSv2.RemoveTPTiming"
	APIerSv2GetTPDestination           = "APIerSv2.GetTPDestination"
	APIerSv2SetTPDestination           = "APIerSv2.SetTPDestination"
	APIerSv2RemoveTPDestination        = "APIerSv2.RemoveTPDestination"
	APIerSv2GetTPDestinationIDs        = "APIerSv2.GetTPDestinationIDs"
	APIerSv2GetTPTiming                = "APIerSv2.GetTPTiming"
	APIerSv2SetTPTiming                = "APIerSv2.SetTPTiming"
	APIerSv2SetAccountActionTriggers   = "APIerSv2.SetAccountActionTriggers"
	APIerSv2GetAccountActionTriggers   = "APIerSv2.GetAccountActionTriggers"
	APIerSv2SetActionPlan              = "APIerSv2.SetActionPlan"
	APIerSv2GetActions                 = "APIerSv2.GetActions"
	APIerSv2GetDestinations            = "APIerSv2.GetDestinations"
	APIerSv2GetCacheStats              = "APIerSv2.GetCacheStats"
	APIerSv2ExecuteAction              = "APIerSv2.ExecuteAction"
	APIerSv2ResetAccountActionTriggers = "APIerSv2.ResetAccountActionTriggers"
	APIerSv2RemoveActions              = "APIerSv2.RemoveActions"
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
	CoreSv1Sleep  = "CoreSv1.Sleep"
)

// SupplierS APIs
const (
	SupplierSv1GetSuppliers                = "SupplierSv1.GetSuppliers"
	SupplierSv1GetSupplierProfilesForEvent = "SupplierSv1.GetSupplierProfilesForEvent"
	SupplierSv1Ping                        = "SupplierSv1.Ping"
	APIerSv1GetSupplierProfile             = "APIerSv1.GetSupplierProfile"
	APIerSv1GetSupplierProfileIDs          = "APIerSv1.GetSupplierProfileIDs"
	APIerSv1RemoveSupplierProfile          = "APIerSv1.RemoveSupplierProfile"
	APIerSv1SetSupplierProfile             = "APIerSv1.SetSupplierProfile"
)

// AttributeS APIs
const (
	APIerSv1SetAttributeProfile      = "APIerSv1.SetAttributeProfile"
	APIerSv1GetAttributeProfile      = "APIerSv1.GetAttributeProfile"
	APIerSv1GetAttributeProfileIDs   = "APIerSv1.GetAttributeProfileIDs"
	APIerSv1RemoveAttributeProfile   = "APIerSv1.RemoveAttributeProfile"
	APIerSv2SetAttributeProfile      = "APIerSv2.SetAttributeProfile"
	AttributeSv1GetAttributeForEvent = "AttributeSv1.GetAttributeForEvent"
	AttributeSv1ProcessEvent         = "AttributeSv1.ProcessEvent"
	AttributeSv1Ping                 = "AttributeSv1.Ping"
)

// ChargerS APIs
const (
	ChargerSv1Ping                = "ChargerSv1.Ping"
	ChargerSv1GetChargersForEvent = "ChargerSv1.GetChargersForEvent"
	ChargerSv1ProcessEvent        = "ChargerSv1.ProcessEvent"
	APIerSv1GetChargerProfile     = "APIerSv1.GetChargerProfile"
	APIerSv1RemoveChargerProfile  = "APIerSv1.RemoveChargerProfile"
	APIerSv1SetChargerProfile     = "APIerSv1.SetChargerProfile"
	APIerSv1GetChargerProfileIDs  = "APIerSv1.GetChargerProfileIDs"
)

// ThresholdS APIs
const (
	ThresholdSv1ProcessEvent          = "ThresholdSv1.ProcessEvent"
	ThresholdSv1GetThreshold          = "ThresholdSv1.GetThreshold"
	ThresholdSv1GetThresholdIDs       = "ThresholdSv1.GetThresholdIDs"
	ThresholdSv1Ping                  = "ThresholdSv1.Ping"
	ThresholdSv1GetThresholdsForEvent = "ThresholdSv1.GetThresholdsForEvent"
	APIerSv1GetThresholdProfileIDs    = "APIerSv1.GetThresholdProfileIDs"
	APIerSv1GetThresholdProfile       = "APIerSv1.GetThresholdProfile"
	APIerSv1RemoveThresholdProfile    = "APIerSv1.RemoveThresholdProfile"
	APIerSv1SetThresholdProfile       = "APIerSv1.SetThresholdProfile"
)

// StatS APIs
const (
	StatSv1ProcessEvent            = "StatSv1.ProcessEvent"
	StatSv1GetQueueIDs             = "StatSv1.GetQueueIDs"
	StatSv1GetQueueStringMetrics   = "StatSv1.GetQueueStringMetrics"
	StatSv1GetQueueFloatMetrics    = "StatSv1.GetQueueFloatMetrics"
	StatSv1Ping                    = "StatSv1.Ping"
	StatSv1GetStatQueuesForEvent   = "StatSv1.GetStatQueuesForEvent"
	StatSv1GetStatQueue            = "StatSv1.GetStatQueue"
	APIerSv1GetStatQueueProfile    = "APIerSv1.GetStatQueueProfile"
	APIerSv1RemoveStatQueueProfile = "APIerSv1.RemoveStatQueueProfile"
	APIerSv1SetStatQueueProfile    = "APIerSv1.SetStatQueueProfile"
	APIerSv1GetStatQueueProfileIDs = "APIerSv1.GetStatQueueProfileIDs"
)

// ResourceS APIs
const (
	ResourceSv1AuthorizeResources   = "ResourceSv1.AuthorizeResources"
	ResourceSv1GetResourcesForEvent = "ResourceSv1.GetResourcesForEvent"
	ResourceSv1AllocateResources    = "ResourceSv1.AllocateResources"
	ResourceSv1ReleaseResources     = "ResourceSv1.ReleaseResources"
	ResourceSv1Ping                 = "ResourceSv1.Ping"
	ResourceSv1GetResource          = "ResourceSv1.GetResource"
	APIerSv1SetResourceProfile      = "APIerSv1.SetResourceProfile"
	APIerSv1RemoveResourceProfile   = "APIerSv1.RemoveResourceProfile"
	APIerSv1GetResourceProfile      = "APIerSv1.GetResourceProfile"
	APIerSv1GetResourceProfileIDs   = "APIerSv1.GetResourceProfileIDs"
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
	SessionSv1Sleep                      = "SessionSv1.Sleep"
)

// Responder APIs
const (
	Responder                            = "Responder"
	ResponderDebit                       = "Responder.Debit"
	ResponderRefundIncrements            = "Responder.RefundIncrements"
	ResponderGetMaxSessionTime           = "Responder.GetMaxSessionTime"
	ResponderMaxDebit                    = "Responder.MaxDebit"
	ResponderRefundRounding              = "Responder.RefundRounding"
	ResponderGetCost                     = "Responder.GetCost"
	ResponderGetCostOnRatingPlans        = "Responder.GetCostOnRatingPlans"
	ResponderGetMaxSessionTimeOnAccounts = "Responder.GetMaxSessionTimeOnAccounts"
	ResponderShutdown                    = "Responder.Shutdown"
	ResponderPing                        = "Responder.Ping"
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
	CacheReplicationHosts        = "*replication_hosts"
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

// GeneralCfg
const (
	NodeIDCfg             = "node_id"
	LoggerCfg             = "logger"
	LogLevelCfg           = "log_level"
	HttpSkipTlsVerifyCfg  = "http_skip_tls_verify"
	RoundingDecimalsCfg   = "rounding_decimals"
	DBDataEncodingCfg     = "dbdata_encoding"
	TpExportPathCfg       = "tpexport_dir"
	PosterAttemptsCfg     = "poster_attempts"
	FailedPostsDirCfg     = "failed_posts_dir"
	FailedPostsTTLCfg     = "failed_posts_ttl"
	DefaultReqTypeCfg     = "default_request_type"
	DefaultCategoryCfg    = "default_category"
	DefaultTenantCfg      = "default_tenant"
	DefaultTimezoneCfg    = "default_timezone"
	DefaultCachingCfg     = "default_caching"
	ConnectAttemptsCfg    = "connect_attempts"
	ReconnectsCfg         = "reconnects"
	ConnectTimeoutCfg     = "connect_timeout"
	ReplyTimeoutCfg       = "reply_timeout"
	LockingTimeoutCfg     = "locking_timeout"
	DigestSeparatorCfg    = "digest_separator"
	DigestEqualCfg        = "digest_equal"
	RSRSepCfg             = "rsr_separator"
	MaxParallelConnsCfg   = "max_parallel_conns"
	ConcurrentRequestsCfg = "concurrent_requests"
	ConcurrentStrategyCfg = "concurrent_strategy"
)

// StorDbCfg
const (
	TypeCfg                = "type"
	MaxOpenConnsCfg        = "max_open_conns"
	MaxIdleConnsCfg        = "max_idle_conns"
	ConnMaxLifetimeCfg     = "conn_max_lifetime"
	StringIndexedFieldsCfg = "string_indexed_fields"
	PrefixIndexedFieldsCfg = "prefix_indexed_fields"
	QueryTimeoutCfg        = "query_timeout"
	SSLModeCfg             = "sslmode"
	ItemsCfg               = "items"
)

// DataDbCfg
const (
	DataDbTypeCfg          = "db_type"
	DataDbHostCfg          = "db_host"
	DataDbPortCfg          = "db_port"
	DataDbNameCfg          = "db_name"
	DataDbUserCfg          = "db_user"
	DataDbPassCfg          = "db_password"
	DataDbSentinelNameCfg  = "redis_sentinel"
	RmtConnsCfg            = "remote_conns"
	RplConnsCfg            = "replication_conns"
	ReplicationFilteredCfg = "replication_filtered"
)

// ItemOpt
const (
	RemoteCfg    = "remote"
	ReplicateCfg = "replicate"
	TTLCfg       = "ttl"
	LimitCfg     = "limit"
	StaticTTLCfg = "static_ttl"
)

// Tls
const (
	ServerCerificateCfg = "server_certificate"
	ServerKeyCfg        = "server_key"
	ServerPolicyCfg     = "server_policy"
	ServerNameCfg       = "server_name"
	ClientCerificateCfg = "client_certificate"
	ClientKeyCfg        = "client_key"
	CaCertificateCfg    = "ca_certificate"
)

// ListenCfg
const (
	RPCJSONListenCfg    = "rpc_json"
	RPCGOBListenCfg     = "rpc_gob"
	HTTPListenCfg       = "http"
	RPCJSONTLSListenCfg = "rpc_json_tls"
	RPCGOBTLSListenCfg  = "rpc_gob_tls"
	HTTPTLSListenCfg    = "http_tls"
)

// HTTPCfg
const (
	HTTPJsonRPCURLCfg        = "json_rpc_url"
	HTTPWSURLCfg             = "ws_url"
	HTTPFreeswitchCDRsURLCfg = "freeswitch_cdrs_url"
	HTTPCDRsURLCfg           = "http_cdrs"
	HTTPUseBasicAuthCfg      = "use_basic_auth"
	HTTPAuthUsersCfg         = "auth_users"
)

// FilterSCfg
const (
	StatSConnsCfg     = "stats_conns"
	ResourceSConnsCfg = "resources_conns"
	ApierSConnsCfg    = "apiers_conns"
)

// RalsCfg
const (
	EnabledCfg                 = "enabled"
	ThresholdSConnsCfg         = "thresholds_conns"
	CacheSConnsCfg             = "caches_conns"
	RpSubjectPrefixMatchingCfg = "rp_subject_prefix_matching"
	RemoveExpiredCfg           = "remove_expired"
	MaxComputedUsageCfg        = "max_computed_usage"
	BalanceRatingSubjectCfg    = "balance_rating_subject"
	MaxIncrementsCfg           = "max_increments"
)

// SchedulerCfg
const (
	CDRsConnsCfg = "cdrs_conns"
	FiltersCfg   = "filters"
)

// CdrsCfg
const (
	ExtraFieldsCfg      = "extra_fields"
	StoreCdrsCfg        = "store_cdrs"
	SMCostRetriesCfg    = "session_cost_retries"
	ChargerSConnsCfg    = "chargers_conns"
	AttributeSConnsCfg  = "attributes_conns"
	OnlineCDRExportsCfg = "online_cdr_exports"
)

// SessionSCfg
const (
	ListenBijsonCfg        = "listen_bijson"
	RALsConnsCfg           = "rals_conns"
	ResSConnsCfg           = "resources_conns"
	ThreshSConnsCfg        = "thresholds_conns"
	SupplSConnsCfg         = "suppliers_conns"
	AttrSConnsCfg          = "attributes_conns"
	ReplicationConnsCfg    = "replication_conns"
	DebitIntervalCfg       = "debit_interval"
	StoreSCostsCfg         = "store_session_costs"
	SessionTTLCfg          = "session_ttl"
	SessionTTLMaxDelayCfg  = "session_ttl_max_delay"
	SessionTTLLastUsedCfg  = "session_ttl_last_used"
	SessionTTLUsageCfg     = "session_ttl_usage"
	SessionTTLLastUsageCfg = "session_ttl_last_usage"
	SessionIndexesCfg      = "session_indexes"
	ClientProtocolCfg      = "client_protocol"
	ChannelSyncIntervalCfg = "channel_sync_interval"
	TerminateAttemptsCfg   = "terminate_attempts"
	AlterableFieldsCfg     = "alterable_fields"
	MinDurLowBalanceCfg    = "min_dur_low_balance"
	DefaultUsageCfg        = "default_usage"
)

// FsAgentCfg
const (
	SessionSConnsCfg       = "sessions_conns"
	SubscribeParkCfg       = "subscribe_park"
	CreateCdrCfg           = "create_cdr"
	LowBalanceAnnFileCfg   = "low_balance_ann_file"
	EmptyBalanceContextCfg = "empty_balance_context"
	EmptyBalanceAnnFileCfg = "empty_balance_ann_file"
	MaxWaitConnectionCfg   = "max_wait_connection"
	EventSocketConnsCfg    = "event_socket_conns"
)

// From Config
const (
	AddressCfg = "address"
	Password   = "password"
	AliasCfg   = "alias"

	// KamAgentCfg
	EvapiConnsCfg = "evapi_conns"
	TimezoneCfg   = "timezone"
	TimezoneCfgC  = "Timezone"

	// AsteriskConnCfg
	UserCf = "user"

	// AsteriskAgentCfg
	CreateCDRCfg     = "create_cdr"
	AsteriskConnsCfg = "asterisk_conns"

	// DiameterAgentCfg
	ListenNetCfg         = "listen_net"
	ListenCfg            = "listen"
	DictionariesPathCfg  = "dictionaries_path"
	OriginHostCfg        = "origin_host"
	OriginRealmCfg       = "origin_realm"
	VendorIdCfg          = "vendor_id"
	ProductNameCfg       = "product_name"
	ConcurrentReqsCfg    = "concurrent_requests"
	SyncedConnReqsCfg    = "synced_conn_requests"
	ASRTemplateCfg       = "asr_template"
	RARTemplateCfg       = "rar_template"
	ForcedDisconnectCfg  = "forced_disconnect"
	TemplatesCfg         = "templates"
	RequestProcessorsCfg = "request_processors"

	// RequestProcessor
	RequestFieldsCfg = "request_fields"
	ReplyFieldsCfg   = "reply_fields"

	// RadiusAgentCfg
	ListenAuthCfg         = "listen_auth"
	ListenAcctCfg         = "listen_acct"
	ClientSecretsCfg      = "client_secrets"
	ClientDictionariesCfg = "client_dictionaries"

	// AttributeSCfg
	IndexedSelectsCfg = "indexed_selects"
	ProcessRunsCfg    = "process_runs"
	NestedFieldsCfg   = "nested_fields"

	// ChargerSCfg
	StoreIntervalCfg = "store_interval"

	// StatSCfg
	StoreUncompressedLimitCfg = "store_uncompressed_limit"

	// Cache
	PartitionsCfg = "partitions"
	PrecacheCfg   = "precache"

	// CdreCfg
	ExportFormatCfg      = "export_format"
	ExportPathCfg        = "export_path"
	AttributeSContextCfg = "attributes_context"
	SynchronousCfg       = "synchronous"
	AttemptsCfg          = "attempts"

	//LoaderSCfg
	IdCfg           = "id"
	DryRunCfg       = "dry_run"
	LockFileNameCfg = "lock_filename"
	TpInDirCfg      = "tp_in_dir"
	TpOutDirCfg     = "tp_out_dir"
	DataCfg         = "data"

	DefaultRatioCfg            = "default_ratio"
	ReadersCfg                 = "readers"
	PoolSize                   = "poolSize"
	Conns                      = "conns"
	FilenameCfg                = "file_name"
	RequestPayloadCfg          = "request_payload"
	ReplyPayloadCfg            = "reply_payload"
	TransportCfg               = "transport"
	StrategyCfg                = "strategy"
	Dynaprepaid_actionplansCfg = "dynaprepaid_actionplans"
)

// FC Template
const (
	TagCfg             = "tag"
	TypeCf             = "type"
	PathCfg            = "path"
	ValueCfg           = "value"
	WidthCfg           = "width"
	StripCfg           = "strip"
	PaddingCfg         = "padding"
	MandatoryCfg       = "mandatory"
	AttributeIDCfg     = "attribute_id"
	NewBranchCfg       = "new_branch"
	BlockerCfg         = "blocker"
	BreakOnSuccessCfg  = "break_on_success"
	Handler_id         = "handler_id"
	LayoutCfg          = "layout"
	CostShiftDigitsCfg = "cost_shift_digits"
	MaskDestIDCfg      = "mask_destinationd_id"
	MaskLenCfg         = "mask_length"
)

// SureTax
const (
	UrlCfg                  = "url"
	ClientNumberCfg         = "client_number"
	ValidationKeyCfg        = "validation_key"
	BusinessUnitCfg         = "business_unit"
	IncludeLocalCostCfg     = "include_local_cost"
	ReturnFileCodeCfg       = "return_file_code"
	ResponseGroupCfg        = "response_group"
	ResponseTypeCfg         = "response_type"
	RegulatoryCodeCfg       = "regulatory_code"
	ClientTrackingCfg       = "client_tracking"
	CustomerNumberCfg       = "customer_number"
	OrigNumberCfg           = "orig_number"
	TermNumberCfg           = "term_number"
	BillToNumberCfg         = "bill_to_number"
	ZipcodeCfg              = "zipcode"
	Plus4Cfg                = "plus4"
	P2PZipcodeCfg           = "p2pzipcode"
	P2PPlus4Cfg             = "p2pplus4"
	UnitsCfg                = "units"
	UnitTypeCfg             = "unit_type"
	TaxIncludedCfg          = "tax_included"
	TaxSitusRuleCfg         = "tax_situs_rule"
	TransTypeCodeCfg        = "trans_type_code"
	SalesTypeCodeCfg        = "sales_type_code"
	TaxExemptionCodeListCfg = "tax_exemption_code_list"
)

// LoaderCgrCfg
const (
	TpIDCfg            = "tpid"
	DataPathCfg        = "data_path"
	DisableReverseCfg  = "disable_reverse"
	FieldSeparatorCfg  = "field_separator"
	CachesConnsCfg     = "caches_conns"
	SchedulerConnsCfg  = "scheduler_conns"
	GapiCredentialsCfg = "gapi_credentials"
	GapiTokenCfg       = "gapi_token"
)

// MigratorCgrCfg
const (
	OutDataDBTypeCfg          = "out_datadb_type"
	OutDataDBHostCfg          = "out_datadb_host"
	OutDataDBPortCfg          = "out_datadb_port"
	OutDataDBNameCfg          = "out_datadb_name"
	OutDataDBUserCfg          = "out_datadb_user"
	OutDataDBPasswordCfg      = "out_datadb_password"
	OutDataDBEncodingCfg      = "out_datadb_encoding"
	OutDataDBRedisSentinelCfg = "out_datadb_redis_sentinel"
	OutStorDBTypeCfg          = "out_stordb_type"
	OutStorDBHostCfg          = "out_stordb_host"
	OutStorDBPortCfg          = "out_stordb_port"
	OutStorDBNameCfg          = "out_stordb_name"
	OutStorDBUserCfg          = "out_stordb_user"
	OutStorDBPasswordCfg      = "out_stordb_password"
	UsersFiltersCfg           = "users_filters"
)

// MailerCfg
const (
	MailerServerCfg   = "server"
	MailerAuthUserCfg = "auth_user"
	MailerAuthPassCfg = "auth_password"
	MailerFromAddrCfg = "from_address"
)

// EventReaderCfg
const (
	IDCfg                       = "id"
	RowLengthCfg                = "row_length"
	FieldSepCfg                 = "field_separator"
	RunDelayCfg                 = "run_delay"
	SourcePathCfg               = "source_path"
	ProcessedPathCfg            = "processed_path"
	XmlRootPathCfg              = "xml_root_path"
	TenantCfg                   = "tenant"
	FlagsCfg                    = "flags"
	FailedCallsPrefixCfg        = "failed_calls_prefix"
	PartialRecordCacheCfg       = "partial_record_cache"
	PartialCacheExpiryActionCfg = "partial_cache_expiry_action"
	FieldsCfg                   = "fields"
	CacheDumpFieldsCfg          = "cache_dump_fields"
)

// CGRConfig
const (
	CdreProfiles     = "cdre"             // from JSON
	LoaderCfg        = "loaders"          // from JSON
	HttpAgentCfg     = "http_agent"       // from JSON
	RpcConns         = "rpc_conns"        // from JSON
	GeneralCfg       = "general"          // from JSON
	DataDbCfg        = "data_db"          // from JSON
	StorDbCfg        = "stor_db"          // from JSON
	TlsCfg           = "tls"              // from JSON
	CacheCfg         = "caches"           // from JSON
	HttpCfg          = "http"             // from JSON
	FilterSCfg       = "filters"          // from JSON
	RalsCfg          = "rals"             // from JSON
	SchedulerCfg     = "schedulers"       // from JSON
	CdrsCfg          = "cdrs"             // from JSON
	SessionSCfg      = "sessions"         // from JSON
	FsAgentCfg       = "freeswitch_agent" // from JSON
	KamAgentCfg      = "kamailio_agent"   // from JSON
	AsteriskAgentCfg = "asterisk_agent"   // from JSON
	DiameterAgentCfg = "diameter_agent"   // from JSON
	RadiusAgentCfg   = "radius_agent"     // from JSON
	DnsAgentCfg      = "dns_agent"        // from JSON
	AttributeSCfg    = "attributes"       // from JSON
	ChargerSCfg      = "chargers"         // from JSON
	ResourceSCfg     = "resources"        // from JSON
	StatsCfg         = "stats"            // from JSON
	ThresholdSCfg    = "thresholds"       // from JSON
	SupplierSCfg     = "suppliers"        // from JSON
	SureTaxCfg       = "suretax"          // from JSON
	DispatcherSCfg   = "dispatchers"      // from JSON
	LoaderCgrCfg     = "loader"           // from JSON
	MigratorCgrCfg   = "migrator"         // from JSON
	MailerCfg        = "mailer"           // from JSON
	AnalyzerSCfg     = "analyzers"        // from JSON
	Apier            = "apiers"           // from JSON
	ErsCfg           = "ers"              // from JSON

)

// Strip/Padding strategy
var (
	// common
	MetaRight = "*right"
	MetaLeft  = "*left"
	// only for strip
	MetaXRight = "*xright"
	MetaXLeft  = "*xleft"
	// only for padding
	MetaZeroLeft = "*zeroleft"
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
