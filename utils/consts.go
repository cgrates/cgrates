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

import (
	"time"
)

var (
	MainCDRFields = NewStringSet([]string{CGRID, Source, OriginHost, OriginID, ToR, RequestType, Tenant, Category,
		Account, Subject, Destination, SetupTime, AnswerTime, Usage, COST, RATED, Partial, RunID,
		PreRated, CostSource, CostDetails, ExtraInfo, OrderID})
	PostPaidRatedSlice = []string{META_POSTPAID, META_RATED}

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

	// CachePartitions enables creation of cache partitions
	CachePartitions = NewStringSet([]string{CacheDestinations, CacheReverseDestinations,
		CacheRatingPlans, CacheRatingProfiles, CacheActions, CacheActionPlans,
		CacheAccountActionPlans, CacheActionTriggers, CacheSharedGroups, CacheTimings,
		CacheResourceProfiles, CacheResources, CacheEventResources, CacheStatQueueProfiles,
		CacheStatQueues, CacheThresholdProfiles, CacheThresholds, CacheFilters,
		CacheRouteProfiles, CacheAttributeProfiles, CacheChargerProfiles, CacheActionProfiles,
		CacheDispatcherProfiles, CacheDispatcherHosts, CacheDispatchers, CacheResourceFilterIndexes,
		CacheStatFilterIndexes, CacheThresholdFilterIndexes, CacheRouteFilterIndexes,
		CacheAttributeFilterIndexes, CacheChargerFilterIndexes, CacheDispatcherFilterIndexes,
		CacheDispatcherRoutes, CacheDispatcherLoads, CacheDiameterMessages, CacheRPCResponses,
		CacheClosedSessions, CacheCDRIDs, CacheLoadIDs, CacheRPCConnections, CacheRatingProfilesTmp,
		CacheUCH, CacheSTIR, CacheEventCharges, CacheRateProfiles, CacheRateProfilesFilterIndexes,
		CacheRateFilterIndexes, CacheActionProfilesFilterIndexes, CacheReverseFilterIndexes, MetaAPIBan, CacheCapsEvents,
		// only internalDB
		CacheVersions, CacheAccounts,
		CacheTBLTPTimings, CacheTBLTPDestinations, CacheTBLTPRates, CacheTBLTPDestinationRates,
		CacheTBLTPRatingPlans, CacheTBLTPRatingProfiles, CacheTBLTPSharedGroups, CacheTBLTPActions,
		CacheTBLTPActionPlans, CacheTBLTPActionTriggers, CacheTBLTPAccountActions, CacheTBLTPResources,
		CacheTBLTPStats, CacheTBLTPThresholds, CacheTBLTPFilters, CacheSessionCostsTBL, CacheCDRsTBL,
		CacheTBLTPRoutes, CacheTBLTPAttributes, CacheTBLTPChargers, CacheTBLTPDispatchers,
		CacheTBLTPDispatcherHosts, CacheTBLTPRateProfiles, CacheTBLTPActionProfiles})
	CacheInstanceToPrefix = map[string]string{
		CacheDestinations:                DESTINATION_PREFIX,
		CacheReverseDestinations:         REVERSE_DESTINATION_PREFIX,
		CacheRatingPlans:                 RATING_PLAN_PREFIX,
		CacheRatingProfiles:              RATING_PROFILE_PREFIX,
		CacheActions:                     ACTION_PREFIX,
		CacheActionPlans:                 ACTION_PLAN_PREFIX,
		CacheAccountActionPlans:          AccountActionPlansPrefix,
		CacheActionTriggers:              ACTION_TRIGGER_PREFIX,
		CacheSharedGroups:                SHARED_GROUP_PREFIX,
		CacheResourceProfiles:            ResourceProfilesPrefix,
		CacheResources:                   ResourcesPrefix,
		CacheTimings:                     TimingsPrefix,
		CacheStatQueueProfiles:           StatQueueProfilePrefix,
		CacheStatQueues:                  StatQueuePrefix,
		CacheThresholdProfiles:           ThresholdProfilePrefix,
		CacheThresholds:                  ThresholdPrefix,
		CacheFilters:                     FilterPrefix,
		CacheRouteProfiles:               RouteProfilePrefix,
		CacheAttributeProfiles:           AttributeProfilePrefix,
		CacheChargerProfiles:             ChargerProfilePrefix,
		CacheDispatcherProfiles:          DispatcherProfilePrefix,
		CacheDispatcherHosts:             DispatcherHostPrefix,
		CacheRateProfiles:                RateProfilePrefix,
		CacheActionProfiles:              ActionProfilePrefix,
		CacheResourceFilterIndexes:       ResourceFilterIndexes,
		CacheStatFilterIndexes:           StatFilterIndexes,
		CacheThresholdFilterIndexes:      ThresholdFilterIndexes,
		CacheRouteFilterIndexes:          RouteFilterIndexes,
		CacheAttributeFilterIndexes:      AttributeFilterIndexes,
		CacheChargerFilterIndexes:        ChargerFilterIndexes,
		CacheDispatcherFilterIndexes:     DispatcherFilterIndexes,
		CacheRateProfilesFilterIndexes:   RateProfilesFilterIndexPrfx,
		CacheActionProfilesFilterIndexes: ActionProfilesFilterIndexPrfx,
		CacheLoadIDs:                     LoadIDPrefix,
		CacheAccounts:                    ACCOUNT_PREFIX,
		CacheRateFilterIndexes:           RateFilterIndexPrfx,
		CacheReverseFilterIndexes:        FilterIndexPrfx,
		MetaAPIBan:                       MetaAPIBan, // special case as it is not in a DB
	}
	CachePrefixToInstance map[string]string    // will be built on init
	CacheIndexesToPrefix  = map[string]string{ // used by match index to get all the ids when index selects is disabled and for compute indexes
		CacheThresholdFilterIndexes:      ThresholdProfilePrefix,
		CacheResourceFilterIndexes:       ResourceProfilesPrefix,
		CacheStatFilterIndexes:           StatQueueProfilePrefix,
		CacheRouteFilterIndexes:          RouteProfilePrefix,
		CacheAttributeFilterIndexes:      AttributeProfilePrefix,
		CacheChargerFilterIndexes:        ChargerProfilePrefix,
		CacheDispatcherFilterIndexes:     DispatcherProfilePrefix,
		CacheRateProfilesFilterIndexes:   RateProfilePrefix,
		CacheRateFilterIndexes:           RatePrefix,
		CacheActionProfilesFilterIndexes: ActionProfilePrefix,
		CacheReverseFilterIndexes:        FilterPrefix,
	}

	CacheInstanceToCacheIndex = map[string]string{
		CacheThresholdProfiles:  CacheThresholdFilterIndexes,
		CacheResourceProfiles:   CacheResourceFilterIndexes,
		CacheStatQueueProfiles:  CacheStatFilterIndexes,
		CacheRouteProfiles:      CacheRouteFilterIndexes,
		CacheAttributeProfiles:  CacheAttributeFilterIndexes,
		CacheChargerProfiles:    CacheChargerFilterIndexes,
		CacheDispatcherProfiles: CacheDispatcherFilterIndexes,
		CacheRateProfiles:       CacheRateProfilesFilterIndexes,
		CacheActionProfiles:     CacheActionProfilesFilterIndexes,
		CacheFilters:            CacheReverseFilterIndexes,
		// CacheRates:              CacheRateFilterIndexes,
	}

	// NonMonetaryBalances are types of balances which are not handled as monetary
	NonMonetaryBalances = NewStringSet([]string{VOICE, SMS, DATA, GENERIC})

	// AccountableRequestTypes are the ones handled by Accounting subsystem
	AccountableRequestTypes = NewStringSet([]string{META_PREPAID, META_POSTPAID, META_PSEUDOPREPAID})

	CacheStorDBPartitions = map[string]string{
		TBLTPTimings:          CacheTBLTPTimings,
		TBLTPDestinations:     CacheTBLTPDestinations,
		TBLTPRates:            CacheTBLTPRates,
		TBLTPDestinationRates: CacheTBLTPDestinationRates,
		TBLTPRatingPlans:      CacheTBLTPRatingPlans,
		TBLTPRatingProfiles:   CacheTBLTPRatingProfiles,
		TBLTPSharedGroups:     CacheTBLTPSharedGroups,
		TBLTPActions:          CacheTBLTPActions,
		TBLTPActionPlans:      CacheTBLTPActionPlans,
		TBLTPActionTriggers:   CacheTBLTPActionTriggers,
		TBLTPAccountActions:   CacheTBLTPAccountActions,
		TBLTPResources:        CacheTBLTPResources,
		TBLTPStats:            CacheTBLTPStats,
		TBLTPThresholds:       CacheTBLTPThresholds,
		TBLTPFilters:          CacheTBLTPFilters,
		SessionCostsTBL:       CacheSessionCostsTBL,
		CDRsTBL:               CacheCDRsTBL,
		TBLTPRoutes:           CacheTBLTPRoutes,
		TBLTPAttributes:       CacheTBLTPAttributes,
		TBLTPChargers:         CacheTBLTPChargers,
		TBLTPDispatchers:      CacheTBLTPDispatchers,
		TBLTPDispatcherHosts:  CacheTBLTPDispatcherHosts,
		TBLTPRateProfiles:     CacheTBLTPRateProfiles,
		TBLTPActionProfiles:   CacheTBLTPActionProfiles,
	}
	// ProtectedSFlds are the fields that sessions should not alter
	ProtectedSFlds   = NewStringSet([]string{CGRID, OriginHost, OriginID, Usage})
	ArgCacheToPrefix = map[string]string{
		DestinationIDs:        DESTINATION_PREFIX,
		ReverseDestinationIDs: REVERSE_DESTINATION_PREFIX,
		RatingPlanIDs:         RATING_PLAN_PREFIX,
		RatingProfileIDs:      RATING_PROFILE_PREFIX,
		ActionIDs:             ACTION_PREFIX,
		ActionPlanIDs:         ACTION_PLAN_PREFIX,
		AccountActionPlanIDs:  AccountActionPlansPrefix,
		ActionTriggerIDs:      ACTION_TRIGGER_PREFIX,
		SharedGroupIDs:        SHARED_GROUP_PREFIX,
		ResourceProfileIDs:    ResourceProfilesPrefix,
		ResourceIDs:           ResourcesPrefix,
		StatsQueueIDs:         StatQueuePrefix,
		StatsQueueProfileIDs:  StatQueueProfilePrefix,
		ThresholdIDs:          ThresholdPrefix,
		ThresholdProfileIDs:   ThresholdProfilePrefix,
		FilterIDs:             FilterPrefix,
		RouteProfileIDs:       RouteProfilePrefix,
		AttributeProfileIDs:   AttributeProfilePrefix,
		ChargerProfileIDs:     ChargerProfilePrefix,
		DispatcherProfileIDs:  DispatcherProfilePrefix,
		DispatcherHostIDs:     DispatcherHostPrefix,
		RateProfileIDs:        RateProfilePrefix,
		ActionProfileIDs:      ActionProfilePrefix,

		TimingIDs:                    TimingsPrefix,
		AttributeFilterIndexIDs:      AttributeFilterIndexes,
		ResourceFilterIndexIDs:       ResourceFilterIndexes,
		StatFilterIndexIDs:           StatFilterIndexes,
		ThresholdFilterIndexIDs:      ThresholdFilterIndexes,
		RouteFilterIndexIDs:          RouteFilterIndexes,
		ChargerFilterIndexIDs:        ChargerFilterIndexes,
		DispatcherFilterIndexIDs:     DispatcherFilterIndexes,
		RateProfilesFilterIndexIDs:   RateProfilesFilterIndexPrfx,
		RateFilterIndexIDs:           RateFilterIndexPrfx,
		ActionProfilesFilterIndexIDs: ActionProfilesFilterIndexPrfx,
		FilterIndexIDs:               FilterIndexPrfx,
	}
	CacheInstanceToArg map[string]string
	ArgCacheToInstance = map[string]string{
		DestinationIDs:        CacheDestinations,
		ReverseDestinationIDs: CacheReverseDestinations,
		RatingPlanIDs:         CacheRatingPlans,
		RatingProfileIDs:      CacheRatingProfiles,
		ActionIDs:             CacheActions,
		ActionPlanIDs:         CacheActionPlans,
		AccountActionPlanIDs:  CacheAccountActionPlans,
		ActionTriggerIDs:      CacheActionTriggers,
		SharedGroupIDs:        CacheSharedGroups,
		ResourceProfileIDs:    CacheResourceProfiles,
		ResourceIDs:           CacheResources,
		StatsQueueIDs:         CacheStatQueues,
		StatsQueueProfileIDs:  CacheStatQueueProfiles,
		ThresholdIDs:          CacheThresholds,
		ThresholdProfileIDs:   CacheThresholdProfiles,
		FilterIDs:             CacheFilters,
		RouteProfileIDs:       CacheRouteProfiles,
		AttributeProfileIDs:   CacheAttributeProfiles,
		ChargerProfileIDs:     CacheChargerProfiles,
		DispatcherProfileIDs:  CacheDispatcherProfiles,
		DispatcherHostIDs:     CacheDispatcherHosts,
		RateProfileIDs:        CacheRateProfiles,
		ActionProfileIDs:      CacheActionProfiles,

		TimingIDs:                    CacheTimings,
		AttributeFilterIndexIDs:      CacheAttributeFilterIndexes,
		ResourceFilterIndexIDs:       CacheResourceFilterIndexes,
		StatFilterIndexIDs:           CacheStatFilterIndexes,
		ThresholdFilterIndexIDs:      CacheThresholdFilterIndexes,
		RouteFilterIndexIDs:          CacheRouteFilterIndexes,
		ChargerFilterIndexIDs:        CacheChargerFilterIndexes,
		DispatcherFilterIndexIDs:     CacheDispatcherFilterIndexes,
		RateProfilesFilterIndexIDs:   CacheRateProfilesFilterIndexes,
		RateFilterIndexIDs:           CacheRateFilterIndexes,
		FilterIndexIDs:               CacheReverseFilterIndexes,
		ActionProfilesFilterIndexIDs: CacheActionProfilesFilterIndexes,
	}
	ConcurrentReqsLimit    int
	ConcurrentReqsStrategy string
)

const (
	CGRateS                      = "CGRateS"
	VERSION                      = "v0.11.0~dev"
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
	MetaSingle                   = "*single"
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
	FilterValStart               = "("
	FilterValEnd                 = ")"
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
	Contexts                     = "Contexts"
	Account                      = "Account"
	Balances                     = "Balances"
	Subject                      = "Subject"
	Destination                  = "Destination"
	SetupTime                    = "SetupTime"
	AnswerTime                   = "AnswerTime"
	Usage                        = "Usage"
	DurationIndex                = "DurationIndex"
	MaxRateUnit                  = "MaxRateUnit"
	DebitInterval                = "DebitInterval"
	TimeStart                    = "TimeStart"
	TimeEnd                      = "TimeEnd"
	CallDuration                 = "CallDuration"
	FallbackSubject              = "FallbackSubject"
	DryRun                       = "DryRun"
	ExtraFields                  = "ExtraFields"
	CustomValue                  = "CustomValue"
	Value                        = "Value"
	LastUsed                     = "LastUsed"
	PDD                          = "PDD"
	ROUTE                        = "Route"
	RunID                        = "RunID"
	AttributeIDs                 = "AttributeIDs"
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
	META_OUT                     = "*out"
	META_ANY                     = "*any"
	META_VOICE                   = "*voice"
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
	RouteProfilePrefix           = "rpp_"
	RatePrefix                   = "rep_"
	AttributeProfilePrefix       = "alp_"
	ChargerProfilePrefix         = "cpp_"
	DispatcherProfilePrefix      = "dpp_"
	RateProfilePrefix            = "rtp_"
	ActionProfilePrefix          = "acp_"
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
	MetaReplicator              = "*replicator"
	MetaRerate                  = "*rerate"
	MetaRefund                  = "*refund"
	MetaStats                   = "*stats"
	MetaResponder               = "*responder"
	MetaCore                    = "*core"
	MetaServiceManager          = "*servicemanager"
	MetaChargers                = "*chargers"
	MetaBlockerError            = "*blocker_error"
	MetaConfig                  = "*config"
	MetaDispatchers             = "*dispatchers"
	MetaDispatcherh             = "*dispatcherh"
	MetaDispatcherHosts         = "*dispatcher_hosts"
	MetaFilters                 = "*filters"
	MetaCDRs                    = "*cdrs"
	MetaDC                      = "*dc"
	MetaCaches                  = "*caches"
	MetaUCH                     = "*uch"
	MetaGuardian                = "*guardians"
	MetaEEs                     = "*ees"
	MetaRateS                   = "*rates"
	MetaContinue                = "*continue"
	Migrator                    = "migrator"
	UnsupportedMigrationTask    = "unsupported migration task"
	NoStorDBConnection          = "not connected to StorDB"
	UndefinedVersion            = "undefined version"
	TxtSuffix                   = ".txt"
	JSNSuffix                   = ".json"
	GOBSuffix                   = ".gob"
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
	MetaVirt                    = "*virt"
	MetaElastic                 = "*elastic"
	MetaFileFWV                 = "*file_fwv"
	MetaFile                    = "*file"
	Accounts                    = "Accounts"
	AccountService              = "AccountS"
	Actions                     = "Actions"
	ActionPlans                 = "ActionPlans"
	ActionTriggers              = "ActionTriggers"
	BalanceMap                  = "BalanceMap"
	UnitCounters                = "UnitCounters"
	UpdateTime                  = "UpdateTime"
	SharedGroups                = "SharedGroups"
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
	RateProfiles                = "RateProfiles"
	ActionProfiles              = "ActionProfiles"
	MetaEveryMinute             = "*every_minute"
	MetaHourly                  = "*hourly"
	ID                          = "ID"
	Address                     = "Address"
	Addresses                   = "Addresses"
	Transport                   = "Transport"
	TLS                         = "TLS"
	Subsystems                  = "Subsystems"
	Strategy                    = "Strategy"
	StrategyParameters          = "StrategyParameters"
	ConnID                      = "ConnID"
	ConnFilterIDs               = "ConnFilterIDs"
	ConnWeight                  = "ConnWeight"
	ConnBlocker                 = "ConnBlocker"
	ConnParameters              = "ConnParameters"

	Thresholds               = "Thresholds"
	Routes                   = "Routes"
	Attributes               = "Attributes"
	Chargers                 = "Chargers"
	Dispatchers              = "Dispatchers"
	StatS                    = "Stats"
	LoadIDsVrs               = "LoadIDs"
	RALService               = "RALs"
	GlobalVarS               = "GlobalVarS"
	CostSource               = "CostSource"
	ExtraInfo                = "ExtraInfo"
	Meta                     = "*"
	MetaSysLog               = "*syslog"
	MetaStdLog               = "*stdout"
	EventSource              = "EventSource"
	AccountID                = "AccountID"
	AccountIDs               = "AccountIDs"
	ResourceID               = "ResourceID"
	TotalUsage               = "TotalUsage"
	StatID                   = "StatID"
	BalanceType              = "BalanceType"
	BalanceID                = "BalanceID"
	BalanceDestinationIds    = "BalanceDestinationIds"
	BalanceWeight            = "BalanceWeight"
	BalanceExpirationDate    = "BalanceExpirationDate"
	BalanceTimingTags        = "BalanceTimingTags"
	BalanceRatingSubject     = "BalanceRatingSubject"
	BalanceCategories        = "BalanceCategories"
	BalanceSharedGroups      = "BalanceSharedGroups"
	BalanceBlocker           = "BalanceBlocker"
	BalanceDisabled          = "BalanceDisabled"
	Units                    = "Units"
	AccountUpdate            = "AccountUpdate"
	BalanceUpdate            = "BalanceUpdate"
	StatUpdate               = "StatUpdate"
	ResourceUpdate           = "ResourceUpdate"
	CDR                      = "CDR"
	CDRs                     = "CDRs"
	ExpiryTime               = "ExpiryTime"
	AllowNegative            = "AllowNegative"
	Disabled                 = "Disabled"
	Action                   = "Action"
	MetaNow                  = "*now"
	SessionSCosts            = "SessionSCosts"
	Timing                   = "Timing"
	RQF                      = "RQF"
	Resource                 = "Resource"
	User                     = "User"
	Subscribers              = "Subscribers"
	DerivedChargersV         = "DerivedChargers"
	Destinations             = "Destinations"
	ReverseDestinations      = "ReverseDestinations"
	RatingPlan               = "RatingPlan"
	RatingProfile            = "RatingProfile"
	MetaRatingPlans          = "*rating_plans"
	MetaRatingProfiles       = "*rating_profiles"
	MetaUsers                = "*users"
	MetaSubscribers          = "*subscribers"
	MetaDerivedChargersV     = "*derivedchargers"
	MetaStorDB               = "*stordb"
	MetaDataDB               = "*datadb"
	MetaWeight               = "*weight"
	MetaLC                   = "*lc"
	MetaHC                   = "*hc"
	MetaQOS                  = "*qos"
	MetaReas                 = "*reas"
	MetaReds                 = "*reds"
	Weight                   = "Weight"
	Cost                     = "Cost"
	Limit                    = "Limit"
	UsageTTL                 = "UsageTTL"
	AllocationMessage        = "AllocationMessage"
	Stored                   = "Stored"
	RatingSubject            = "RatingSubject"
	Categories               = "Categories"
	Blocker                  = "Blocker"
	RatingPlanID             = "RatingPlanID"
	StartTime                = "StartTime"
	AccountSummary           = "AccountSummary"
	RatingFilters            = "RatingFilters"
	RatingFilter             = "RatingFilter"
	Accounting               = "Accounting"
	Rating                   = "Rating"
	Charges                  = "Charges"
	CompressFactor           = "CompressFactor"
	Increments               = "Increments"
	Balance                  = "Balance"
	BalanceSummaries         = "BalanceSummaries"
	Type                     = "Type"
	Element                  = "Element"
	Values                   = "Values"
	YearsFieldName           = "Years"
	MonthsFieldName          = "Months"
	MonthDaysFieldName       = "MonthDays"
	WeekDaysFieldName        = "WeekDays"
	GroupIntervalStart       = "GroupIntervalStart"
	RateIncrement            = "RateIncrement"
	RateUnit                 = "RateUnit"
	BalanceUUID              = "BalanceUUID"
	RatingID                 = "RatingID"
	ExtraChargeID            = "ExtraChargeID"
	ConnectFee               = "ConnectFee"
	RoundingMethod           = "RoundingMethod"
	RoundingDecimals         = "RoundingDecimals"
	MaxCostStrategy          = "MaxCostStrategy"
	RateID                   = "RateID"
	RateIDs                  = "RateIDs"
	RateFilterIDs            = "RateFilterIDs"
	RateActivationStart      = "RateActivationStart"
	RateWeight               = "RateWeight"
	RateIntervalStart        = "RateIntervalStart"
	RateFixedFee             = "RateFixedFee"
	RateRecurrentFee         = "RateRecurrentFee"
	RateBlocker              = "RateBlocker"
	TimingID                 = "TimingID"
	RatesID                  = "RatesID"
	RatingFiltersID          = "RatingFiltersID"
	AccountingID             = "AccountingID"
	MetaSessionS             = "*sessions"
	MetaDefault              = "*default"
	Error                    = "Error"
	MetaCgreq                = "*cgreq"
	MetaCgrep                = "*cgrep"
	MetaCGRAReq              = "*cgrareq"
	CGR_ACD                  = "cgr_acd"
	ActivationIntervalString = "ActivationInterval"
	MaxHits                  = "MaxHits"
	MinHits                  = "MinHits"
	Async                    = "Async"
	Sorting                  = "Sorting"
	SortingParameters        = "SortingParameters"
	RouteAccountIDs          = "RouteAccountIDs"
	RouteRatingplanIDs       = "RouteRatingplanIDs"
	RouteRateProfileIDs      = "RouteRateProfileIDs"
	RouteStatIDs             = "RouteStatIDs"
	RouteWeight              = "RouteWeight"
	RouteParameters          = "RouteParameters"
	RouteBlocker             = "RouteBlocker"
	RouteResourceIDs         = "RouteResourceIDs"
	RouteFilterIDs           = "RouteFilterIDs"
	AttributeFilterIDs       = "AttributeFilterIDs"
	QueueLength              = "QueueLength"
	TTL                      = "TTL"
	MinItems                 = "MinItems"
	MetricIDs                = "MetricIDs"
	MetricFilterIDs          = "MetricFilterIDs"
	FieldName                = "FieldName"
	Path                     = "Path"
	MetaRound                = "*round"
	Pong                     = "Pong"
	MetaEventCost            = "*event_cost"
	MetaEventUsage           = "*event_usage"
	MetaVoiceUsage           = "*voice_usage"
	MetaSMSUsage             = "*sms_usage"
	MetaDataUsage            = "*data_usage"
	MetaMMSUsage             = "*mms_usage"
	MetaGenericUsage         = "*generic_usage"
	MetaPositiveExports      = "*positive_exports"
	MetaNegativeExports      = "*negative_exports"
	MetaRoutesEventCost      = "*routes_event_cost"
	MetaRoutesIgnoreErrors   = "*routes_ignore_errors"
	Freeswitch               = "freeswitch"
	Kamailio                 = "kamailio"
	Opensips                 = "opensips"
	Asterisk                 = "asterisk"
	SchedulerS               = "SchedulerS"
	MetaMultiply             = "*multiply"
	MetaDivide               = "*divide"
	MetaUrl                  = "*url"
	MetaXml                  = "*xml"
	MetaReq                  = "*req"
	MetaAsm                  = "*asm"
	MetaVars                 = "*vars"
	MetaRep                  = "*rep"
	MetaExp                  = "*exp"
	MetaHdr                  = "*hdr"
	MetaTrl                  = "*trl"
	MetaTmp                  = "*tmp"
	MetaOpts                 = "*opts"
	MetaDynReq               = "~*req"
	MetaScPrefix             = "~*sc." // used for SMCostFilter
	CGROriginHost            = "cgr_originhost"
	MetaInitiate             = "*initiate"
	MetaUpdate               = "*update"
	MetaTerminate            = "*terminate"
	MetaEvent                = "*event"
	MetaMessage              = "*message"
	MetaDryRun               = "*dryrun"
	Event                    = "Event"
	EmptyString              = ""
	DynamicDataPrefix        = "~"
	AttrValueSep             = "="
	ANDSep                   = "&"
	PipeSep                  = "|"
	RSRConstSep              = "`"
	RSRConstChar             = '`'
	RSRDataConverterPrefix   = "{*"
	RSRDataConverterSufix    = "}"
	RSRDynStartChar          = '<'
	RSRDynEndChar            = '>'
	MetaApp                  = "*app"
	MetaAppID                = "*appid"
	MetaCmd                  = "*cmd"
	MetaEnv                  = "*env:" // use in config for describing enviormant variables
	MetaTemplate             = "*template"
	MetaCCA                  = "*cca"
	MetaErr                  = "*err"
	OriginRealm              = "OriginRealm"
	ProductName              = "ProductName"
	IdxStart                 = "["
	IdxEnd                   = "]"
	MetaLog                  = "*log"
	MetaRemoteHost           = "*remote_host"
	RemoteHost               = "RemoteHost"
	Local                    = "local"
	TCP                      = "tcp"
	UDP                      = "udp"
	CGRDebitInterval         = "CGRDebitInterval"
	Version                  = "Version"
	MetaTenant               = "*tenant"
	ResourceUsage            = "ResourceUsage"
	MetaDuration             = "*duration"
	MetaLibPhoneNumber       = "*libphonenumber"
	MetaTimeString           = "*time_string"
	MetaIP2Hex               = "*ip2hex"
	MetaString2Hex           = "*string2hex"
	MetaUnixTime             = "*unixtime"
	MetaSIPURIMethod         = "*sipuri_method"
	MetaSIPURIHost           = "*sipuri_host"
	MetaSIPURIUser           = "*sipuri_user"
	MetaReload               = "*reload"
	MetaLoad                 = "*load"
	MetaRemove               = "*remove"
	MetaRemoveAll            = "*removeall"
	MetaStore                = "*store"
	MetaClear                = "*clear"
	MetaExport               = "*export"
	MetaExportID             = "*export_id"
	MetaTimeNow              = "*time_now"
	MetaFirstEventATime      = "*first_event_atime"
	MetaLastEventATime       = "*last_event_atime"
	MetaEventNumber          = "*event_number"
	LoadIDs                  = "load_ids"
	DNSAgent                 = "DNSAgent"
	TLSNoCaps                = "tls"
	UsageID                  = "UsageID"
	Rcode                    = "Rcode"
	Replacement              = "Replacement"
	Regexp                   = "Regexp"
	Order                    = "Order"
	Preference               = "Preference"
	Flags                    = "Flags"
	Service                  = "Service"
	ApierV                   = "ApierV"
	MetaApier                = "*apier"
	MetaAnalyzer             = "*analyzer"
	CGREventString           = "CGREvent"
	MetaTextPlain            = "*text_plain"
	MetaIgnoreErrors         = "*ignore_errors"
	MetaRelease              = "*release"
	MetaAllocate             = "*allocate"
	MetaAuthorize            = "*authorize"
	MetaSTIRAuthenticate     = "*stir_authenticate"
	MetaSTIRInitiate         = "*stir_initiate"
	MetaInit                 = "*init"
	MetaRatingPlanCost       = "*rating_plan_cost"
	ERs                      = "ERs"
	EEs                      = "EEs"
	Ratio                    = "Ratio"
	Load                     = "Load"
	Slash                    = "/"
	UUID                     = "UUID"
	ActionsID                = "ActionsID"
	MetaAct                  = "*act"
	MetaAcnt                 = "*acnt"
	DestinationPrefix        = "DestinationPrefix"
	DestinationID            = "DestinationID"
	ExportTemplate           = "ExportTemplate"
	ExportFormat             = "ExportFormat"
	Synchronous              = "Synchronous"
	Attempts                 = "Attempts"
	FieldSeparator           = "FieldSeparator"
	ExportPath               = "ExportPath"
	ExporterIDs              = "ExporterIDs"
	TimeNow                  = "TimeNow"
	ExportFileName           = "ExportFileName"
	GroupID                  = "GroupID"
	ThresholdType            = "ThresholdType"
	ThresholdValue           = "ThresholdValue"
	Recurrent                = "Recurrent"
	Executed                 = "Executed"
	MinSleep                 = "MinSleep"
	ActivationDate           = "ActivationDate"
	ExpirationDate           = "ExpirationDate"
	MinQueuedItems           = "MinQueuedItems"
	OrderIDStart             = "OrderIDStart"
	OrderIDEnd               = "OrderIDEnd"
	MinCost                  = "MinCost"
	MaxCost                  = "MaxCost"
	MetaLoaders              = "*loaders"
	TmpSuffix                = ".tmp"
	MetaDiamreq              = "*diamreq"
	MetaCost                 = "*cost"
	MetaGroup                = "*group"
	InternalRPCSet           = "InternalRPCSet"
	FileName                 = "FileName"
	MetaRadauth              = "*radauth"
	UserPassword             = "UserPassword"
	RadauthFailed            = "RADAUTH_FAILED"
	MetaPAP                  = "*pap"
	MetaCHAP                 = "*chap"
	MetaMSCHAPV2             = "*mschapv2"
	MetaDynaprepaid          = "*dynaprepaid"
	MetaFD                   = "*fd"
	SortingData              = "SortingData"
	Count                    = "Count"
	ProfileID                = "ProfileID"
	SortedRoutes             = "SortedRoutes"
	EventExporterS           = "EventExporterS"
	MetaMonthly              = "*monthly"
	MetaYearly               = "*yearly"
	MetaDaily                = "*daily"
	MetaWeekly               = "*weekly"
	RateS                    = "RateS"
	Underline                = "_"
	MetaPartial              = "*partial"
	MetaBusy                 = "*busy"
	MetaQueue                = "*queue"
	MetaMonthEnd             = "*month_end"
	APIKey                   = "ApiKey"
	RouteID                  = "RouteID"
	MetaMonthlyEstimated     = "*monthly_estimated"
	ProcessRuns              = "ProcessRuns"
	HashtagSep               = "#"
	MetaRounding             = "*rounding"
	StatsNA                  = -1.0
	RateProfileMatched       = "RateProfileMatched"
	InvalidDuration          = time.Duration(-1)
	ActionS                  = "ActionS"
	Schedule                 = "Schedule"
	ActionFilterIDs          = "ActionFilterIDs"
	ActionBlocker            = "ActionBlocker"
	ActionTTL                = "ActionTTL"
	ActionOpts               = "ActionOpts"
	ActionPath               = "ActionPath"
	TPid                     = "TPid"
	LoadId                   = "LoadId"
	ActionPlanId             = "ActionPlanId"
	AccountActionsId         = "AccountActionsId"
	Loadid                   = "loadid"
	AccountLowerCase         = "account"
	ActionPlan               = "ActionPlan"
	ActionsId                = "ActionsId"
	TimingId                 = "TimingId"
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
	MetaRouteProfiles       = "*route_profiles"
	MetaAttributeProfiles   = "*attribute_profiles"
	MetaIndexes             = "*indexes"
	MetaDispatcherProfiles  = "*dispatcher_profiles"
	MetaRateProfiles        = "*rate_profiles"
	MetaRateProfileRates    = "*rate_profile_rates"
	MetaChargerProfiles     = "*charger_profiles"
	MetaSharedGroups        = "*shared_groups"
	MetaThresholds          = "*thresholds"
	MetaRoutes              = "*routes"
	MetaAttributes          = "*attributes"
	MetaActionProfiles      = "*action_profiles"
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
	MetaRAR      = "*rar"
)

// Services
const (
	SessionS    = "SessionS"
	AttributeS  = "AttributeS"
	RouteS      = "RouteS"
	ResourceS   = "ResourceS"
	StatService = "StatS"
	FilterS     = "FilterS"
	ThresholdS  = "ThresholdS"
	DispatcherS = "DispatcherS"
	DispatcherH = "DispatcherH"
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
	RoutesLow      = "routes"
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
	EEsLow         = "ees"
	RateSLow       = "rates"
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
	MetaRemoveSessionCosts    = "*remove_session_costs"
	MetaRemoveExpired         = "*remove_expired"
	MetaPostEvent             = "*post_event"
	MetaCDRAccount            = "*reset_account_cdr"
	MetaResetThreshold        = "*reset_threshold"
	MetaResetStatQueue        = "*reset_stat_queue"
	MetaRemoteSetAccount      = "*remote_set_account"
	ActionID                  = "ActionID"
	ActionType                = "ActionType"
	ActionValue               = "ActionValue"
	BalanceValue              = "BalanceValue"
	ExtraParameters           = "ExtraParameters"
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
	MetaTpRoutes            = "*tp_Routes"
	MetaTpStats             = "*tp_stats"
	MetaTpSharedGroups      = "*tp_shared_groups"
	MetaTpRatingProfiles    = "*tp_rating_profiles"
	MetaTpActionProfiles    = "*tp_action_profiles"
	MetaTpRateProfiles      = "*tp_rate_profiles"
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
	CapRoutes               = "Routes"
	CapThresholds           = "Thresholds"
	CapStatQueues           = "StatQueues"
)

const (
	TpRatingPlans        = "TpRatingPlans"
	TpFilters            = "TpFilters"
	TpDestinationRates   = "TpDestinationRates"
	TpActionTriggers     = "TpActionTriggers"
	TpAccountActionsV    = "TpAccountActions"
	TpActionPlans        = "TpActionPlans"
	TpActions            = "TpActions"
	TpThresholds         = "TpThresholds"
	TpRoutes             = "TpRoutes"
	TpAttributes         = "TpAttributes"
	TpStats              = "TpStats"
	TpSharedGroups       = "TpSharedGroups"
	TpRatingProfiles     = "TpRatingProfiles"
	TpResources          = "TpResources"
	TpRates              = "TpRates"
	TpTiming             = "TpTiming"
	TpResource           = "TpResource"
	TpDestinations       = "TpDestinations"
	TpRatingPlan         = "TpRatingPlan"
	TpRatingProfile      = "TpRatingProfile"
	TpChargers           = "TpChargers"
	TpDispatchers        = "TpDispatchers"
	TpDispatcherProfiles = "TpDispatcherProfiles"
	TpDispatcherHosts    = "TpDispatcherHosts"
	TpRateProfiles       = "TpRateProfiles"
	TpActionProfiles     = "TpActionProfiles"
)

// Dispatcher Const
const (
	MetaFirst          = "*first"
	MetaRandom         = "*random"
	MetaRoundRobin     = "*round_robin"
	MetaRatio          = "*ratio"
	MetaDefaultRatio   = "*default_ratio"
	ThresholdSv1       = "ThresholdSv1"
	StatSv1            = "StatSv1"
	ResourceSv1        = "ResourceSv1"
	RouteSv1           = "RouteSv1"
	AttributeSv1       = "AttributeSv1"
	SessionSv1         = "SessionSv1"
	ChargerSv1         = "ChargerSv1"
	MetaAuth           = "*auth"
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
	MetaIPNet          = "*ipnet"
	MetaAPIBan         = "*apiban"

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
	MetaNotIPNet        = "*notipnet"
	MetaNotAPIBan       = "*notapiban"

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
	ReplicatorSv1GetSharedGroup          = "ReplicatorSv1.GetSharedGroup"
	ReplicatorSv1GetActions              = "ReplicatorSv1.GetActions"
	ReplicatorSv1GetActionPlan           = "ReplicatorSv1.GetActionPlan"
	ReplicatorSv1GetAllActionPlans       = "ReplicatorSv1.GetAllActionPlans"
	ReplicatorSv1GetAccountActionPlans   = "ReplicatorSv1.GetAccountActionPlans"
	ReplicatorSv1GetRatingPlan           = "ReplicatorSv1.GetRatingPlan"
	ReplicatorSv1GetRatingProfile        = "ReplicatorSv1.GetRatingProfile"
	ReplicatorSv1GetRouteProfile         = "ReplicatorSv1.GetRouteProfile"
	ReplicatorSv1GetAttributeProfile     = "ReplicatorSv1.GetAttributeProfile"
	ReplicatorSv1GetChargerProfile       = "ReplicatorSv1.GetChargerProfile"
	ReplicatorSv1GetDispatcherProfile    = "ReplicatorSv1.GetDispatcherProfile"
	ReplicatorSv1GetRateProfile          = "ReplicatorSv1.GetRateProfile"
	ReplicatorSv1GetActionProfile        = "ReplicatorSv1.GetActionProfile"
	ReplicatorSv1GetDispatcherHost       = "ReplicatorSv1.GetDispatcherHost"
	ReplicatorSv1GetItemLoadIDs          = "ReplicatorSv1.GetItemLoadIDs"
	ReplicatorSv1SetThresholdProfile     = "ReplicatorSv1.SetThresholdProfile"
	ReplicatorSv1SetThreshold            = "ReplicatorSv1.SetThreshold"
	ReplicatorSv1SetAccount              = "ReplicatorSv1.SetAccount"
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
	ReplicatorSv1SetRouteProfile         = "ReplicatorSv1.SetRouteProfile"
	ReplicatorSv1SetAttributeProfile     = "ReplicatorSv1.SetAttributeProfile"
	ReplicatorSv1SetChargerProfile       = "ReplicatorSv1.SetChargerProfile"
	ReplicatorSv1SetDispatcherProfile    = "ReplicatorSv1.SetDispatcherProfile"
	ReplicatorSv1SetRateProfile          = "ReplicatorSv1.SetRateProfile"
	ReplicatorSv1SetActionProfile        = "ReplicatorSv1.SetActionProfile"
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
	ReplicatorSv1RemoveRouteProfile      = "ReplicatorSv1.RemoveRouteProfile"
	ReplicatorSv1RemoveAttributeProfile  = "ReplicatorSv1.RemoveAttributeProfile"
	ReplicatorSv1RemoveChargerProfile    = "ReplicatorSv1.RemoveChargerProfile"
	ReplicatorSv1RemoveDispatcherProfile = "ReplicatorSv1.RemoveDispatcherProfile"
	ReplicatorSv1RemoveRateProfile       = "ReplicatorSv1.RemoveRateProfile"
	ReplicatorSv1RemoveActionProfile     = "ReplicatorSv1.RemoveActionProfile"
	ReplicatorSv1RemoveDispatcherHost    = "ReplicatorSv1.RemoveDispatcherHost"
	ReplicatorSv1GetIndexes              = "ReplicatorSv1.GetIndexes"
	ReplicatorSv1SetIndexes              = "ReplicatorSv1.SetIndexes"
	ReplicatorSv1RemoveIndexes           = "ReplicatorSv1.RemoveIndexes"
)

// APIerSv1 APIs
const (
	ApierV1                             = "ApierV1"
	ApierV2                             = "ApierV2"
	APIerSv1                            = "APIerSv1"
	APIerSv1ComputeFilterIndexes        = "APIerSv1.ComputeFilterIndexes"
	APIerSv1ComputeFilterIndexIDs       = "APIerSv1.ComputeFilterIndexIDs"
	APIerSv1Ping                        = "APIerSv1.Ping"
	APIerSv1SetDispatcherProfile        = "APIerSv1.SetDispatcherProfile"
	APIerSv1GetDispatcherProfile        = "APIerSv1.GetDispatcherProfile"
	APIerSv1GetDispatcherProfileIDs     = "APIerSv1.GetDispatcherProfileIDs"
	APIerSv1RemoveDispatcherProfile     = "APIerSv1.RemoveDispatcherProfile"
	APIerSv1SetBalances                 = "APIerSv1.SetBalances"
	APIerSv1SetDispatcherHost           = "APIerSv1.SetDispatcherHost"
	APIerSv1GetDispatcherHost           = "APIerSv1.GetDispatcherHost"
	APIerSv1GetDispatcherHostIDs        = "APIerSv1.GetDispatcherHostIDs"
	APIerSv1RemoveDispatcherHost        = "APIerSv1.RemoveDispatcherHost"
	APIerSv1GetEventCost                = "APIerSv1.GetEventCost"
	APIerSv1LoadTariffPlanFromFolder    = "APIerSv1.LoadTariffPlanFromFolder"
	APIerSv1ExportToFolder              = "APIerSv1.ExportToFolder"
	APIerSv1GetCost                     = "APIerSv1.GetCost"
	APIerSv1SetBalance                  = "APIerSv1.SetBalance"
	APIerSv1GetFilter                   = "APIerSv1.GetFilter"
	APIerSv1GetFilterIndexes            = "APIerSv1.GetFilterIndexes"
	APIerSv1RemoveFilterIndexes         = "APIerSv1.RemoveFilterIndexes"
	APIerSv1RemoveFilter                = "APIerSv1.RemoveFilter"
	APIerSv1SetFilter                   = "APIerSv1.SetFilter"
	APIerSv1GetFilterIDs                = "APIerSv1.GetFilterIDs"
	APIerSv1GetRatingProfile            = "APIerSv1.GetRatingProfile"
	APIerSv1RemoveRatingProfile         = "APIerSv1.RemoveRatingProfile"
	APIerSv1SetRatingProfile            = "APIerSv1.SetRatingProfile"
	APIerSv1GetRatingProfileIDs         = "APIerSv1.GetRatingProfileIDs"
	APIerSv1SetDataDBVersions           = "APIerSv1.SetDataDBVersions"
	APIerSv1SetStorDBVersions           = "APIerSv1.SetStorDBVersions"
	APIerSv1GetAccountActionPlan        = "APIerSv1.GetAccountActionPlan"
	APIerSv1ComputeActionPlanIndexes    = "APIerSv1.ComputeActionPlanIndexes"
	APIerSv1GetActions                  = "APIerSv1.GetActions"
	APIerSv1GetActionPlan               = "APIerSv1.GetActionPlan"
	APIerSv1GetActionPlanIDs            = "APIerSv1.GetActionPlanIDs"
	APIerSv1GetRatingPlanIDs            = "APIerSv1.GetRatingPlanIDs"
	APIerSv1GetRatingPlan               = "APIerSv1.GetRatingPlan"
	APIerSv1RemoveRatingPlan            = "APIerSv1.RemoveRatingPlan"
	APIerSv1GetDestination              = "APIerSv1.GetDestination"
	APIerSv1RemoveDestination           = "APIerSv1.RemoveDestination"
	APIerSv1GetReverseDestination       = "APIerSv1.GetReverseDestination"
	APIerSv1AddBalance                  = "APIerSv1.AddBalance"
	APIerSv1DebitBalance                = "APIerSv1.DebitBalance"
	APIerSv1SetAccount                  = "APIerSv1.SetAccount"
	APIerSv1GetAccountsCount            = "APIerSv1.GetAccountsCount"
	APIerSv1GetDataDBVersions           = "APIerSv1.GetDataDBVersions"
	APIerSv1GetStorDBVersions           = "APIerSv1.GetStorDBVersions"
	APIerSv1GetCDRs                     = "APIerSv1.GetCDRs"
	APIerSv1GetTPAccountActions         = "APIerSv1.GetTPAccountActions"
	APIerSv1SetTPAccountActions         = "APIerSv1.SetTPAccountActions"
	APIerSv1GetTPAccountActionsByLoadId = "APIerSv1.GetTPAccountActionsByLoadId"
	APIerSv1GetTPAccountActionLoadIds   = "APIerSv1.GetTPAccountActionLoadIds"
	APIerSv1GetTPAccountActionIds       = "APIerSv1.GetTPAccountActionIds"
	APIerSv1RemoveTPAccountActions      = "APIerSv1.RemoveTPAccountActions"
	APIerSv1GetTPActionPlan             = "APIerSv1.GetTPActionPlan"
	APIerSv1SetTPActionPlan             = "APIerSv1.SetTPActionPlan"
	APIerSv1GetTPActionPlanIds          = "APIerSv1.GetTPActionPlanIds"
	APIerSv1SetTPActionTriggers         = "APIerSv1.SetTPActionTriggers"
	APIerSv1GetTPActionTriggers         = "APIerSv1.GetTPActionTriggers"
	APIerSv1RemoveTPActionTriggers      = "APIerSv1.RemoveTPActionTriggers"
	APIerSv1GetTPActionTriggerIds       = "APIerSv1.GetTPActionTriggerIds"
	APIerSv1GetTPActions                = "APIerSv1.GetTPActions"
	APIerSv1RemoveTPActionPlan          = "APIerSv1.RemoveTPActionPlan"
	APIerSv1GetTPAttributeProfile       = "APIerSv1.GetTPAttributeProfile"
	APIerSv1SetTPAttributeProfile       = "APIerSv1.SetTPAttributeProfile"
	APIerSv1GetTPAttributeProfileIds    = "APIerSv1.GetTPAttributeProfileIds"
	APIerSv1RemoveTPAttributeProfile    = "APIerSv1.RemoveTPAttributeProfile"
	APIerSv1GetTPCharger                = "APIerSv1.GetTPCharger"
	APIerSv1SetTPCharger                = "APIerSv1.SetTPCharger"
	APIerSv1RemoveTPCharger             = "APIerSv1.RemoveTPCharger"
	APIerSv1GetTPChargerIDs             = "APIerSv1.GetTPChargerIDs"
	APIerSv1SetTPFilterProfile          = "APIerSv1.SetTPFilterProfile"
	APIerSv1GetTPFilterProfile          = "APIerSv1.GetTPFilterProfile"
	APIerSv1GetTPFilterProfileIds       = "APIerSv1.GetTPFilterProfileIds"
	APIerSv1RemoveTPFilterProfile       = "APIerSv1.RemoveTPFilterProfile"
	APIerSv1GetTPDestination            = "APIerSv1.GetTPDestination"
	APIerSv1SetTPDestination            = "APIerSv1.SetTPDestination"
	APIerSv1GetTPDestinationIDs         = "APIerSv1.GetTPDestinationIDs"
	APIerSv1RemoveTPDestination         = "APIerSv1.RemoveTPDestination"
	APIerSv1GetTPResource               = "APIerSv1.GetTPResource"
	APIerSv1SetTPResource               = "APIerSv1.SetTPResource"
	APIerSv1RemoveTPResource            = "APIerSv1.RemoveTPResource"
	APIerSv1SetTPRate                   = "APIerSv1.SetTPRate"
	APIerSv1GetTPRate                   = "APIerSv1.GetTPRate"
	APIerSv1RemoveTPRate                = "APIerSv1.RemoveTPRate"
	APIerSv1GetTPRateIds                = "APIerSv1.GetTPRateIds"
	APIerSv1SetTPThreshold              = "APIerSv1.SetTPThreshold"
	APIerSv1GetTPThreshold              = "APIerSv1.GetTPThreshold"
	APIerSv1GetTPThresholdIDs           = "APIerSv1.GetTPThresholdIDs"
	APIerSv1RemoveTPThreshold           = "APIerSv1.RemoveTPThreshold"
	APIerSv1SetTPStat                   = "APIerSv1.SetTPStat"
	APIerSv1GetTPStat                   = "APIerSv1.GetTPStat"
	APIerSv1RemoveTPStat                = "APIerSv1.RemoveTPStat"
	APIerSv1GetTPDestinationRate        = "APIerSv1.GetTPDestinationRate"
	APIerSv1SetTPRouteProfile           = "APIerSv1.SetTPRouteProfile"
	APIerSv1GetTPRouteProfile           = "APIerSv1.GetTPRouteProfile"
	APIerSv1GetTPRouteProfileIDs        = "APIerSv1.GetTPRouteProfileIDs"
	APIerSv1RemoveTPRouteProfile        = "APIerSv1.RemoveTPRouteProfile"
	APIerSv1GetTPDispatcherProfile      = "APIerSv1.GetTPDispatcherProfile"
	APIerSv1SetTPDispatcherProfile      = "APIerSv1.SetTPDispatcherProfile"
	APIerSv1RemoveTPDispatcherProfile   = "APIerSv1.RemoveTPDispatcherProfile"
	APIerSv1GetTPDispatcherProfileIDs   = "APIerSv1.GetTPDispatcherProfileIDs"
	APIerSv1GetTPSharedGroups           = "APIerSv1.GetTPSharedGroups"
	APIerSv1SetTPSharedGroups           = "APIerSv1.SetTPSharedGroups"
	APIerSv1GetTPSharedGroupIds         = "APIerSv1.GetTPSharedGroupIds"
	APIerSv1RemoveTPSharedGroups        = "APIerSv1.RemoveTPSharedGroups"
	APIerSv1ExportCDRs                  = "APIerSv1.ExportCDRs"
	APIerSv1GetTPRatingPlan             = "APIerSv1.GetTPRatingPlan"
	APIerSv1SetTPRatingPlan             = "APIerSv1.SetTPRatingPlan"
	APIerSv1GetTPRatingPlanIds          = "APIerSv1.GetTPRatingPlanIds"
	APIerSv1RemoveTPRatingPlan          = "APIerSv1.RemoveTPRatingPlan"
	APIerSv1SetTPActions                = "APIerSv1.SetTPActions"
	APIerSv1GetTPActionIds              = "APIerSv1.GetTPActionIds"
	APIerSv1RemoveTPActions             = "APIerSv1.RemoveTPActions"
	APIerSv1SetActionPlan               = "APIerSv1.SetActionPlan"
	APIerSv1ExecuteAction               = "APIerSv1.ExecuteAction"
	APIerSv1SetTPRatingProfile          = "APIerSv1.SetTPRatingProfile"
	APIerSv1GetTPRatingProfile          = "APIerSv1.GetTPRatingProfile"
	APIerSv1RemoveTPRatingProfile       = "APIerSv1.RemoveTPRatingProfile"
	APIerSv1SetTPDestinationRate        = "APIerSv1.SetTPDestinationRate"
	APIerSv1GetTPRatingProfileLoadIds   = "APIerSv1.GetTPRatingProfileLoadIds"
	APIerSv1GetTPRatingProfilesByLoadID = "APIerSv1.GetTPRatingProfilesByLoadID"
	APIerSv1GetTPRatingProfileIds       = "APIerSv1.GetTPRatingProfileIds"
	APIerSv1GetTPDestinationRateIds     = "APIerSv1.GetTPDestinationRateIds"
	APIerSv1RemoveTPDestinationRate     = "APIerSv1.RemoveTPDestinationRate"
	APIerSv1ImportTariffPlanFromFolder  = "APIerSv1.ImportTariffPlanFromFolder"
	APIerSv1ExportTPToFolder            = "APIerSv1.ExportTPToFolder"
	APIerSv1LoadRatingPlan              = "APIerSv1.LoadRatingPlan"
	APIerSv1LoadRatingProfile           = "APIerSv1.LoadRatingProfile"
	APIerSv1LoadAccountActions          = "APIerSv1.LoadAccountActions"
	APIerSv1SetActions                  = "APIerSv1.SetActions"
	APIerSv1AddTriggeredAction          = "APIerSv1.AddTriggeredAction"
	APIerSv1GetAccountActionTriggers    = "APIerSv1.GetAccountActionTriggers"
	APIerSv1AddAccountActionTriggers    = "APIerSv1.AddAccountActionTriggers"
	APIerSv1ResetAccountActionTriggers  = "APIerSv1.ResetAccountActionTriggers"
	APIerSv1SetAccountActionTriggers    = "APIerSv1.SetAccountActionTriggers"
	APIerSv1RemoveAccountActionTriggers = "APIerSv1.RemoveAccountActionTriggers"
	APIerSv1GetScheduledActions         = "APIerSv1.GetScheduledActions"
	APIerSv1RemoveActionTiming          = "APIerSv1.RemoveActionTiming"
	APIerSv1ComputeReverseDestinations  = "APIerSv1.ComputeReverseDestinations"
	APIerSv1ComputeAccountActionPlans   = "APIerSv1.ComputeAccountActionPlans"
	APIerSv1SetDestination              = "APIerSv1.SetDestination"
	APIerSv1GetDataCost                 = "APIerSv1.GetDataCost"
	APIerSv1ReplayFailedPosts           = "APIerSv1.ReplayFailedPosts"
	APIerSv1RemoveAccount               = "APIerSv1.RemoveAccount"
	APIerSv1DebitUsage                  = "APIerSv1.DebitUsage"
	APIerSv1GetCacheStats               = "APIerSv1.GetCacheStats"
	APIerSv1ReloadCache                 = "APIerSv1.ReloadCache"
	APIerSv1GetActionTriggers           = "APIerSv1.GetActionTriggers"
	APIerSv1SetActionTrigger            = "APIerSv1.SetActionTrigger"
	APIerSv1RemoveActionPlan            = "APIerSv1.RemoveActionPlan"
	APIerSv1RemoveActions               = "APIerSv1.RemoveActions"
	APIerSv1RemoveBalances              = "APIerSv1.RemoveBalances"
	APIerSv1GetLoadHistory              = "APIerSv1.GetLoadHistory"
	APIerSv1GetLoadIDs                  = "APIerSv1.GetLoadIDs"
	APIerSv1GetLoadTimes                = "APIerSv1.GetLoadTimes"
	APIerSv1ExecuteScheduledActions     = "APIerSv1.ExecuteScheduledActions"
	APIerSv1GetSharedGroup              = "APIerSv1.GetSharedGroup"
	APIerSv1RemoveActionTrigger         = "APIerSv1.RemoveActionTrigger"
	APIerSv1GetAccount                  = "APIerSv1.GetAccount"
	APIerSv1GetAttributeProfileIDsCount = "APIerSv1.GetAttributeProfileIDsCount"
	APIerSv1GetMaxUsage                 = "APIerSv1.GetMaxUsage"
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
	APIerSv2ExportCdrsToFile           = "APIerSv2.ExportCdrsToFile"
	APIerSv2GetAccountsCount           = "APIerSv2.GetAccountsCount"
	APIerSv2GetActionsCount            = "APIerSv2.GetActionsCount"
)

const (
	ServiceManagerV1              = "ServiceManagerV1"
	ServiceManagerV1StartService  = "ServiceManagerV1.StartService"
	ServiceManagerV1StopService   = "ServiceManagerV1.StopService"
	ServiceManagerV1ServiceStatus = "ServiceManagerV1.ServiceStatus"
	ServiceManagerV1Ping          = "ServiceManagerV1.Ping"
)

// ConfigSv1 APIs
const (
	ConfigSv1                  = "ConfigSv1"
	ConfigSv1ReloadConfig      = "ConfigSv1.ReloadConfig"
	ConfigSv1GetConfig         = "ConfigSv1.GetConfig"
	ConfigSv1SetConfig         = "ConfigSv1.SetConfig"
	ConfigSv1GetConfigAsJSON   = "ConfigSv1.GetConfigAsJSON"
	ConfigSv1SetConfigFromJSON = "ConfigSv1.SetConfigFromJSON"
)

const (
	RALsV1                   = "RALsV1"
	RALsV1GetRatingPlansCost = "RALsV1.GetRatingPlansCost"
	RALsV1Ping               = "RALsV1.Ping"
)

const (
	RateSv1             = "RateSv1"
	RateSv1CostForEvent = "RateSv1.CostForEvent"
	RateSv1Ping         = "RateSv1.Ping"
)

const (
	ActionSv1     = "ActionSv1"
	ActionSv1Ping = "ActionSv1.Ping"
)

const (
	CoreS         = "CoreS"
	CoreSv1       = "CoreSv1"
	CoreSv1Status = "CoreSv1.Status"
	CoreSv1Ping   = "CoreSv1.Ping"
	CoreSv1Sleep  = "CoreSv1.Sleep"
)

// RouteS APIs
const (
	RouteSv1GetRoutes                = "RouteSv1.GetRoutes"
	RouteSv1GetRouteProfilesForEvent = "RouteSv1.GetRouteProfilesForEvent"
	RouteSv1Ping                     = "RouteSv1.Ping"
	APIerSv1GetRouteProfile          = "APIerSv1.GetRouteProfile"
	APIerSv1GetRouteProfileIDs       = "APIerSv1.GetRouteProfileIDs"
	APIerSv1RemoveRouteProfile       = "APIerSv1.RemoveRouteProfile"
	APIerSv1SetRouteProfile          = "APIerSv1.SetRouteProfile"
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
	ThresholdSv1ProcessEvent            = "ThresholdSv1.ProcessEvent"
	ThresholdSv1GetThreshold            = "ThresholdSv1.GetThreshold"
	ThresholdSv1ResetThreshold          = "ThresholdSv1.ResetThreshold"
	ThresholdSv1GetThresholdIDs         = "ThresholdSv1.GetThresholdIDs"
	ThresholdSv1Ping                    = "ThresholdSv1.Ping"
	ThresholdSv1GetThresholdsForEvent   = "ThresholdSv1.GetThresholdsForEvent"
	APIerSv1GetThresholdProfileIDs      = "APIerSv1.GetThresholdProfileIDs"
	APIerSv1GetThresholdProfileIDsCount = "APIerSv1.GetThresholdProfileIDsCount"
	APIerSv1GetThresholdProfile         = "APIerSv1.GetThresholdProfile"
	APIerSv1RemoveThresholdProfile      = "APIerSv1.RemoveThresholdProfile"
	APIerSv1SetThresholdProfile         = "APIerSv1.SetThresholdProfile"
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
	StatSv1V1GetQueueIDs           = "StatSv1.GetQueueIDs"
	StatSv1ResetStatQueue          = "StatSv1.ResetStatQueue"
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
	SessionSv1GetCost                    = "SessionSv1.GetCost"
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
	SessionSv1ReAuthorize                = "SessionSv1.ReAuthorize"
	SessionSv1DisconnectPeer             = "SessionSv1.DisconnectPeer"
	SessionSv1WarnDisconnect             = "SessionSv1.WarnDisconnect"
	SessionSv1STIRAuthenticate           = "SessionSv1.STIRAuthenticate"
	SessionSv1STIRIdentity               = "SessionSv1.STIRIdentity"
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
	DispatcherSv1                   = "DispatcherSv1"
	DispatcherSv1Ping               = "DispatcherSv1.Ping"
	DispatcherSv1GetProfileForEvent = "DispatcherSv1.GetProfileForEvent"
	DispatcherSv1Apier              = "DispatcherSv1.Apier"
	DispatcherServicePing           = "DispatcherService.Ping"
)

// DispatcherH APIs
const (
	DispatcherHv1RegisterHosts   = "DispatcherHv1.RegisterHosts"
	DispatcherHv1UnregisterHosts = "DispatcherHv1.UnregisterHosts"
)

// RateProfile APIs
const (
	APIerSv1SetRateProfile         = "APIerSv1.SetRateProfile"
	APIerSv1GetRateProfile         = "APIerSv1.GetRateProfile"
	APIerSv1GetRateProfileIDs      = "APIerSv1.GetRateProfileIDs"
	APIerSv1GetRateProfileIDsCount = "APIerSv1.GetRateProfileIDsCount"
	APIerSv1RemoveRateProfile      = "APIerSv1.RemoveRateProfile"
	APIerSv1SetRateProfileRates    = "APIerSv1.SetRateProfileRates"
	APIerSv1RemoveRateProfileRates = "APIerSv1.RemoveRateProfileRates"
)

// AnalyzerS APIs
const (
	AnalyzerSv1            = "AnalyzerSv1"
	AnalyzerSv1Ping        = "AnalyzerSv1.Ping"
	AnalyzerSv1StringQuery = "AnalyzerSv1.StringQuery"
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
	CacheSv1RemoveItems       = "CacheSv1.RemoveItems"
	CacheSv1PrecacheStatus    = "CacheSv1.PrecacheStatus"
	CacheSv1HasGroup          = "CacheSv1.HasGroup"
	CacheSv1GetGroupItemIDs   = "CacheSv1.GetGroupItemIDs"
	CacheSv1RemoveGroup       = "CacheSv1.RemoveGroup"
	CacheSv1Clear             = "CacheSv1.Clear"
	CacheSv1ReloadCache       = "CacheSv1.ReloadCache"
	CacheSv1LoadCache         = "CacheSv1.LoadCache"
	CacheSv1Ping              = "CacheSv1.Ping"
	CacheSv1ReplicateSet      = "CacheSv1.ReplicateSet"
	CacheSv1ReplicateRemove   = "CacheSv1.ReplicateRemove"
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
	CDRsV2ProcessEvent       = "CDRsV2.ProcessEvent"
)

// Scheduler
const (
	SchedulerSv1                   = "SchedulerSv1"
	SchedulerSv1Ping               = "SchedulerSv1.Ping"
	SchedulerSv1Reload             = "SchedulerSv1.Reload"
	SchedulerSv1ExecuteActions     = "SchedulerSv1.ExecuteActions"
	SchedulerSv1ExecuteActionPlans = "SchedulerSv1.ExecuteActionPlans"
)

// EEs
const (
	EeSv1             = "EeSv1"
	EeSv1Ping         = "EeSv1.Ping"
	EeSv1ProcessEvent = "EeSv1.ProcessEvent"
)

// ActionProfile APIs
const (
	APIerSv1SetActionProfile         = "APIerSv1.SetActionProfile"
	APIerSv1GetActionProfile         = "APIerSv1.GetActionProfile"
	APIerSv1GetActionProfileIDs      = "APIerSv1.GetActionProfileIDs"
	APIerSv1GetActionProfileIDsCount = "APIerSv1.GetActionProfileIDsCount"
	APIerSv1RemoveActionProfile      = "APIerSv1.RemoveActionProfile"
)

//cgr_ variables
const (
	CGR_ACCOUNT          = "cgr_account"
	CGR_ROUTE            = "cgr_route"
	CGR_DESTINATION      = "cgr_destination"
	CGR_SUBJECT          = "cgr_subject"
	CGR_CATEGORY         = "cgr_category"
	CGR_REQTYPE          = "cgr_reqtype"
	CGR_TENANT           = "cgr_tenant"
	CGR_PDD              = "cgr_pdd"
	CGR_DISCONNECT_CAUSE = "cgr_disconnectcause"
	CGR_COMPUTELCR       = "cgr_computelcr"
	CGR_ROUTES           = "cgr_routes"
	CGRFlags             = "cgr_flags"
	CGROpts              = "cgr_opts"
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
	RoutesCsv             = "Routes.csv"
	AttributesCsv         = "Attributes.csv"
	ChargersCsv           = "Chargers.csv"
	DispatcherProfilesCsv = "DispatcherProfiles.csv"
	DispatcherHostsCsv    = "DispatcherHosts.csv"
	RateProfilesCsv       = "RateProfiles.csv"
	ActionProfilesCsv     = "ActionProfiles.csv"
)

// Table Name
const (
	TBLTPTimings          = "tp_timings"
	TBLTPDestinations     = "tp_destinations"
	TBLTPRates            = "tp_rates"
	TBLTPDestinationRates = "tp_destination_rates"
	TBLTPRatingPlans      = "tp_rating_plans"
	TBLTPRatingProfiles   = "tp_rating_profiles"
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
	TBLTPRoutes           = "tp_routes"
	TBLTPAttributes       = "tp_attributes"
	TBLTPChargers         = "tp_chargers"
	TBLVersions           = "versions"
	OldSMCosts            = "sm_costs"
	TBLTPDispatchers      = "tp_dispatcher_profiles"
	TBLTPDispatcherHosts  = "tp_dispatcher_hosts"
	TBLTPRateProfiles     = "tp_rate_profiles"
	TBLTPActionProfiles   = "tp_action_profiles"
)

// Cache Name
const (
	CacheDestinations                = "*destinations"
	CacheReverseDestinations         = "*reverse_destinations"
	CacheRatingPlans                 = "*rating_plans"
	CacheRatingProfiles              = "*rating_profiles"
	CacheActions                     = "*actions"
	CacheActionPlans                 = "*action_plans"
	CacheAccountActionPlans          = "*account_action_plans"
	CacheActionTriggers              = "*action_triggers"
	CacheSharedGroups                = "*shared_groups"
	CacheResources                   = "*resources"
	CacheResourceProfiles            = "*resource_profiles"
	CacheTimings                     = "*timings"
	CacheEventResources              = "*event_resources"
	CacheStatQueueProfiles           = "*statqueue_profiles"
	CacheStatQueues                  = "*statqueues"
	CacheThresholdProfiles           = "*threshold_profiles"
	CacheThresholds                  = "*thresholds"
	CacheFilters                     = "*filters"
	CacheRouteProfiles               = "*route_profiles"
	CacheAttributeProfiles           = "*attribute_profiles"
	CacheChargerProfiles             = "*charger_profiles"
	CacheDispatcherProfiles          = "*dispatcher_profiles"
	CacheDispatcherHosts             = "*dispatcher_hosts"
	CacheDispatchers                 = "*dispatchers"
	CacheDispatcherRoutes            = "*dispatcher_routes"
	CacheDispatcherLoads             = "*dispatcher_loads"
	CacheRateProfiles                = "*rate_profiles"
	CacheActionProfiles              = "*action_profiles"
	CacheResourceFilterIndexes       = "*resource_filter_indexes"
	CacheStatFilterIndexes           = "*stat_filter_indexes"
	CacheThresholdFilterIndexes      = "*threshold_filter_indexes"
	CacheRouteFilterIndexes          = "*route_filter_indexes"
	CacheAttributeFilterIndexes      = "*attribute_filter_indexes"
	CacheChargerFilterIndexes        = "*charger_filter_indexes"
	CacheDispatcherFilterIndexes     = "*dispatcher_filter_indexes"
	CacheDiameterMessages            = "*diameter_messages"
	CacheRPCResponses                = "*rpc_responses"
	CacheClosedSessions              = "*closed_sessions"
	CacheRateProfilesFilterIndexes   = "*rate_profile_filter_indexes"
	CacheActionProfilesFilterIndexes = "*action_profile_filter_indexes"
	CacheRateFilterIndexes           = "*rate_filter_indexes"
	MetaPrecaching                   = "*precaching"
	MetaReady                        = "*ready"
	CacheLoadIDs                     = "*load_ids"
	CacheRPCConnections              = "*rpc_connections"
	CacheCDRIDs                      = "*cdr_ids"
	CacheRatingProfilesTmp           = "*tmp_rating_profiles"
	CacheUCH                         = "*uch"
	CacheSTIR                        = "*stir"
	CacheEventCharges                = "*event_charges"
	CacheReverseFilterIndexes        = "*reverse_filter_indexes"
	CacheAccounts                    = "*accounts"
	CacheVersions                    = "*versions"
	CacheCapsEvents                  = "*caps_events"

	// storDB
	CacheTBLTPTimings          = "*tp_timings"
	CacheTBLTPDestinations     = "*tp_destinations"
	CacheTBLTPRates            = "*tp_rates"
	CacheTBLTPDestinationRates = "*tp_destination_rates"
	CacheTBLTPRatingPlans      = "*tp_rating_plans"
	CacheTBLTPRatingProfiles   = "*tp_rating_profiles"
	CacheTBLTPSharedGroups     = "*tp_shared_groups"
	CacheTBLTPActions          = "*tp_actions"
	CacheTBLTPActionPlans      = "*tp_action_plans"
	CacheTBLTPActionTriggers   = "*tp_action_triggers"
	CacheTBLTPAccountActions   = "*tp_account_actions"
	CacheTBLTPResources        = "*tp_resources"
	CacheTBLTPStats            = "*tp_stats"
	CacheTBLTPThresholds       = "*tp_thresholds"
	CacheTBLTPFilters          = "*tp_filters"
	CacheSessionCostsTBL       = "*session_costs"
	CacheCDRsTBL               = "*cdrs"
	CacheTBLTPRoutes           = "*tp_routes"
	CacheTBLTPAttributes       = "*tp_attributes"
	CacheTBLTPChargers         = "*tp_chargers"
	CacheTBLTPDispatchers      = "*tp_dispatcher_profiles"
	CacheTBLTPDispatcherHosts  = "*tp_dispatcher_hosts"
	CacheTBLTPRateProfiles     = "*tp_rate_profiles"
	CacheTBLTPActionProfiles   = "*tp_action_profiles"
)

// Prefix for indexing
const (
	ResourceFilterIndexes         = "rfi_"
	StatFilterIndexes             = "sfi_"
	ThresholdFilterIndexes        = "tfi_"
	AttributeFilterIndexes        = "afi_"
	ChargerFilterIndexes          = "cfi_"
	DispatcherFilterIndexes       = "dfi_"
	ActionPlanIndexes             = "api_"
	RouteFilterIndexes            = "rti_"
	RateProfilesFilterIndexPrfx   = "rpi_"
	RateFilterIndexPrfx           = "rri_"
	ActionProfilesFilterIndexPrfx = "aci_"
	FilterIndexPrfx               = "fii_"
)

// Agents
const (
	KamailioAgent   = "KamailioAgent"
	RadiusAgent     = "RadiusAgent"
	DiameterAgent   = "DiameterAgent"
	FreeSWITCHAgent = "FreeSWITCHAgent"
	AsteriskAgent   = "AsteriskAgent"
	HTTPAgent       = "HTTPAgent"
	SIPAgent        = "SIPAgent"
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
	NodeIDCfg           = "node_id"
	LoggerCfg           = "logger"
	LogLevelCfg         = "log_level"
	RoundingDecimalsCfg = "rounding_decimals"
	DBDataEncodingCfg   = "dbdata_encoding"
	TpExportPathCfg     = "tpexport_dir"
	PosterAttemptsCfg   = "poster_attempts"
	FailedPostsDirCfg   = "failed_posts_dir"
	FailedPostsTTLCfg   = "failed_posts_ttl"
	DefaultReqTypeCfg   = "default_request_type"
	DefaultCategoryCfg  = "default_category"
	DefaultTenantCfg    = "default_tenant"
	DefaultTimezoneCfg  = "default_timezone"
	DefaultCachingCfg   = "default_caching"
	ConnectAttemptsCfg  = "connect_attempts"
	ReconnectsCfg       = "reconnects"
	ConnectTimeoutCfg   = "connect_timeout"
	ReplyTimeoutCfg     = "reply_timeout"
	LockingTimeoutCfg   = "locking_timeout"
	DigestSeparatorCfg  = "digest_separator"
	DigestEqualCfg      = "digest_equal"
	RSRSepCfg           = "rsr_separator"
	MaxParallelConnsCfg = "max_parallel_conns"
	EEsConnsCfg         = "ees_conns"
)

// StorDbCfg
const (
	TypeCfg                = "type"
	MaxOpenConnsCfg        = "max_open_conns"
	MaxIdleConnsCfg        = "max_idle_conns"
	ConnMaxLifetimeCfg     = "conn_max_lifetime"
	StringIndexedFieldsCfg = "string_indexed_fields"
	PrefixIndexedFieldsCfg = "prefix_indexed_fields"
	SuffixIndexedFieldsCfg = "suffix_indexed_fields"
	QueryTimeoutCfg        = "query_timeout"
	SSLModeCfg             = "sslmode"
	ItemsCfg               = "items"
	OptsCfg                = "opts"
)

// DataDbCfg
const (
	DataDbTypeCfg              = "db_type"
	DataDbHostCfg              = "db_host"
	DataDbPortCfg              = "db_port"
	DataDbNameCfg              = "db_name"
	DataDbUserCfg              = "db_user"
	DataDbPassCfg              = "db_password"
	RedisSentinelNameCfg       = "redis_sentinel"
	RmtConnsCfg                = "remote_conns"
	RplConnsCfg                = "replication_conns"
	RedisClusterCfg            = "redis_cluster"
	RedisClusterSyncCfg        = "redis_cluster_sync"
	RedisClusterOnDownDelayCfg = "redis_cluster_ondown_delay"
	RedisTLS                   = "redis_tls"
	RedisClientCertificate     = "redis_client_certificate"
	RedisClientKey             = "redis_client_key"
	RedisCACertificate         = "redis_ca_certificate"
)

// ItemOpt
const (
	APIKeyCfg    = "api_key"
	RouteIDCfg   = "route_id"
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
	HTTPJsonRPCURLCfg          = "json_rpc_url"
	DispatchersRegistrarURLCfg = "dispatchers_registrar_url"
	HTTPWSURLCfg               = "ws_url"
	HTTPFreeswitchCDRsURLCfg   = "freeswitch_cdrs_url"
	HTTPCDRsURLCfg             = "http_cdrs"
	HTTPUseBasicAuthCfg        = "use_basic_auth"
	HTTPAuthUsersCfg           = "auth_users"
	HTTPClientOptsCfg          = "client_opts"
	ConfigsURL                 = "configs_url"

	HTTPClientTLSClientConfigCfg       = "skipTlsVerify"
	HTTPClientTLSHandshakeTimeoutCfg   = "tlsHandshakeTimeout"
	HTTPClientDisableKeepAlivesCfg     = "disableKeepAlives"
	HTTPClientDisableCompressionCfg    = "disableCompression"
	HTTPClientMaxIdleConnsCfg          = "maxIdleConns"
	HTTPClientMaxIdleConnsPerHostCfg   = "maxIdleConnsPerHost"
	HTTPClientMaxConnsPerHostCfg       = "maxConnsPerHost"
	HTTPClientIdleConnTimeoutCfg       = "idleConnTimeout"
	HTTPClientResponseHeaderTimeoutCfg = "responseHeaderTimeout"
	HTTPClientExpectContinueTimeoutCfg = "expectContinueTimeout"
	HTTPClientForceAttemptHTTP2Cfg     = "forceAttemptHttp2"
	HTTPClientDialTimeoutCfg           = "dialTimeout"
	HTTPClientDialFallbackDelayCfg     = "dialFallbackDelay"
	HTTPClientDialKeepAliveCfg         = "dialKeepAlive"
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
	ExtraFieldsCfg         = "extra_fields"
	StoreCdrsCfg           = "store_cdrs"
	SMCostRetriesCfg       = "session_cost_retries"
	ChargerSConnsCfg       = "chargers_conns"
	AttributeSConnsCfg     = "attributes_conns"
	RetransmissionTimerCfg = "retransmission_timer"
	OnlineCDRExportsCfg    = "online_cdr_exports"
	SessionCostRetires     = "session_cost_retries"
	RateSConnsCfg          = "rates_conns"
)

// SessionSCfg
const (
	ListenBijsonCfg        = "listen_bijson"
	RALsConnsCfg           = "rals_conns"
	ResSConnsCfg           = "resources_conns"
	ThreshSConnsCfg        = "thresholds_conns"
	RouteSConnsCfg         = "routes_conns"
	AttrSConnsCfg          = "attributes_conns"
	ReplicationConnsCfg    = "replication_conns"
	RemoteConnsCfg         = "remote_conns"
	DebitIntervalCfg       = "debit_interval"
	StoreSCostsCfg         = "store_session_costs"
	MinCallDurationCfg     = "min_call_duration"
	MaxCallDurationCfg     = "max_call_duration"
	SessionTTLCfg          = "session_ttl"
	SessionTTLMaxDelayCfg  = "session_ttl_max_delay"
	SessionTTLLastUsedCfg  = "session_ttl_last_used"
	SessionTTLLastUsageCfg = "session_ttl_last_usage"
	SessionTTLUsageCfg     = "session_ttl_usage"
	SessionIndexesCfg      = "session_indexes"
	ClientProtocolCfg      = "client_protocol"
	ChannelSyncIntervalCfg = "channel_sync_interval"
	TerminateAttemptsCfg   = "terminate_attempts"
	AlterableFieldsCfg     = "alterable_fields"
	MinDurLowBalanceCfg    = "min_dur_low_balance"
	STIRCfg                = "stir"

	AllowedAtestCfg       = "allowed_attest"
	PayloadMaxdurationCfg = "payload_maxduration"
	DefaultAttestCfg      = "default_attest"
	PublicKeyPathCfg      = "publickey_path"
	PrivateKeyPathCfg     = "privatekey_path"
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
	EmptyBalanceContext    = "empty_balance_context"
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
	ListenNetCfg          = "listen_net"
	ConcurrentRequestsCfg = "concurrent_requests"
	ListenCfg             = "listen"
	DictionariesPathCfg   = "dictionaries_path"
	OriginHostCfg         = "origin_host"
	OriginRealmCfg        = "origin_realm"
	VendorIDCfg           = "vendor_id"
	ProductNameCfg        = "product_name"
	SyncedConnReqsCfg     = "synced_conn_requests"
	ASRTemplateCfg        = "asr_template"
	RARTemplateCfg        = "rar_template"
	ForcedDisconnectCfg   = "forced_disconnect"
	TemplatesCfg          = "templates"
	RequestProcessorsCfg  = "request_processors"

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
	ExportPathCfg        = "export_path"
	AttributeSContextCfg = "attributes_context"
	SynchronousCfg       = "synchronous"
	AttemptsCfg          = "attempts"
	AttributeContextCfg  = "attribute_context"
	AttributeIDsCfg      = "attribute_ids"

	//LoaderSCfg
	DryRunCfg       = "dry_run"
	LockFileNameCfg = "lock_filename"
	TpInDirCfg      = "tp_in_dir"
	TpOutDirCfg     = "tp_out_dir"
	DataCfg         = "data"

	DefaultRatioCfg           = "default_ratio"
	ReadersCfg                = "readers"
	ExportersCfg              = "exporters"
	PoolSize                  = "poolSize"
	Conns                     = "conns"
	FilenameCfg               = "file_name"
	RequestPayloadCfg         = "request_payload"
	ReplyPayloadCfg           = "reply_payload"
	TransportCfg              = "transport"
	StrategyCfg               = "strategy"
	DynaprepaidActionplansCfg = "dynaprepaid_actionplans"

	//RateSCfg
	RateIndexedSelectsCfg      = "rate_indexed_selects"
	RateNestedFieldsCfg        = "rate_nested_fields"
	RateStringIndexedFieldsCfg = "rate_string_indexed_fields"
	RatePrefixIndexedFieldsCfg = "rate_prefix_indexed_fields"
	RateSuffixIndexedFieldsCfg = "rate_suffix_indexed_fields"
	Verbosity                  = "verbosity"

	// AnalyzerSCfg
	CleanupIntervalCfg = "cleanup_interval"
	IndexTypeCfg       = "index_type"
	DBPathCfg          = "db_path"

	// CoreSCfg
	CapsCfg              = "caps"
	CapsStrategyCfg      = "caps_strategy"
	CapsStatsIntervalCfg = "caps_stats_interval"
	ShutdownTimeoutCfg   = "shutdown_timeout"
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
	HandlerIDCfg       = "handler_id"
	LayoutCfg          = "layout"
	CostShiftDigitsCfg = "cost_shift_digits"
	MaskDestIDCfg      = "mask_destinationd_id"
	MaskLenCfg         = "mask_length"
)

// SureTax
const (
	RootDirCfg              = "root_dir"
	URLCfg                  = "url"
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
	CachesConnsCfg     = "caches_conns"
	SchedulerConnsCfg  = "scheduler_conns"
	GapiCredentialsCfg = "gapi_credentials"
	GapiTokenCfg       = "gapi_token"
)

// MigratorCgrCfg
const (
	OutDataDBTypeCfg     = "out_datadb_type"
	OutDataDBHostCfg     = "out_datadb_host"
	OutDataDBPortCfg     = "out_datadb_port"
	OutDataDBNameCfg     = "out_datadb_name"
	OutDataDBUserCfg     = "out_datadb_user"
	OutDataDBPasswordCfg = "out_datadb_password"
	OutDataDBEncodingCfg = "out_datadb_encoding"
	OutStorDBTypeCfg     = "out_stordb_type"
	OutStorDBHostCfg     = "out_stordb_host"
	OutStorDBPortCfg     = "out_stordb_port"
	OutStorDBNameCfg     = "out_stordb_name"
	OutStorDBUserCfg     = "out_stordb_user"
	OutStorDBPasswordCfg = "out_stordb_password"
	OutStorDBOptsCfg     = "out_stordb_opts"
	OutDataDBOptsCfg     = "out_datadb_opts"
	UsersFiltersCfg      = "users_filters"
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
	CacheCfg                    = "cache"
	RowLengthCfg                = "row_length"
	FieldSepCfg                 = "field_separator"
	HeaderDefCharCfg            = "header_define_character"
	RunDelayCfg                 = "run_delay"
	SourcePathCfg               = "source_path"
	ProcessedPathCfg            = "processed_path"
	XMLRootPathCfg              = "xml_root_path"
	TenantCfg                   = "tenant"
	FlagsCfg                    = "flags"
	FailedCallsPrefixCfg        = "failed_calls_prefix"
	PartialRecordCacheCfg       = "partial_record_cache"
	PartialCacheExpiryActionCfg = "partial_cache_expiry_action"
	FieldsCfg                   = "fields"
	CacheDumpFieldsCfg          = "cache_dump_fields"
)

// DispatcherHCfg
const (
	DispatchersConnsCfg  = "dispatchers_conns"
	HostsCfg             = "hosts"
	RegisterIntervalCfg  = "register_interval"
	RegisterTransportCfg = "register_transport"
	RegisterTLSCfg       = "register_tls"
)

// APIBanCfg
const (
	KeysCfg = "keys"
)

// STIR/SHAKEN
const (
	STIRAlg = "ES256"
	STIRPpt = "shaken"
	STIRTyp = "passport"

	STIRAlgField  = "alg"
	STIRPptField  = "ppt"
	STIRInfoField = "info"

	STIRExtraInfoPrefix = ";info=<"
	STIRExtraInfoSuffix = ">;alg=ES256;ppt=shaken"
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

// CGROptionsSet the possible cgr options
var CGROptionsSet = NewStringSet([]string{OptsRatesStartTime, OptsRatesUsage, OptsSessionTTL, OptsSessionTTLMaxDelay,
	OptsSessionTTLLastUsed, OptsSessionTTLLastUsage, OptsSessionTTLUsage, OptsDebitInterval, OptsStirATest,
	OptsStirPayloadMaxDuration, OptsStirIdentity, OptsStirOriginatorTn, OptsStirOriginatorURI,
	OptsStirDestinationTn, OptsStirDestinationURI, OptsStirPublicKeyPath, OptsStirPrivateKeyPath,
	OptsAPIKey, OptsRouteID, OptsContext, OptsAttributesProcessRuns, OptsRoutesLimit, OptsRoutesOffset})

// EventExporter metrics
const (
	NumberOfEvents    = "NumberOfEvents"
	TotalCost         = "TotalCost"
	PositiveExports   = "PositiveExports"
	NegativeExports   = "NegativeExports"
	FirstExpOrderID   = "FirstExpOrderID"
	LastExpOrderID    = "LastExpOrderID"
	FirstEventATime   = "FirstEventATime"
	LastEventATime    = "LastEventATime"
	TotalDuration     = "TotalDuration"
	TotalDataUsage    = "TotalDataUsage"
	TotalSMSUsage     = "TotalSMSUsage"
	TotalMMSUsage     = "TotalMMSUsage"
	TotalGenericUsage = "TotalGenericUsage"
	FilePath          = "FilePath"
)

// Event Opts
const (
	OptsRoutesLimit         = "*routes_limit"
	OptsRoutesOffset        = "*routes_offset"
	OptsRatesStartTime      = "*ratesStartTime"
	OptsRatesUsage          = "*ratesUsage"
	OptsSessionTTL          = "*sessionTTL"
	OptsSessionTTLMaxDelay  = "*sessionTTLMaxDelay"
	OptsSessionTTLLastUsed  = "*sessionTTLLastUsed"
	OptsSessionTTLLastUsage = "*sessionTTLLastUsage"
	OptsSessionTTLUsage     = "*sessionTTLUsage"
	OptsDebitInterval       = "*sessionDebitInterval"
	// STIR
	OptsStirATest              = "*stirATest"
	OptsStirPayloadMaxDuration = "*stirPayloadMaxDuration"
	OptsStirIdentity           = "*stirIdentity"
	OptsStirOriginatorTn       = "*stirOriginatorTn"
	OptsStirOriginatorURI      = "*stirOriginatorURI"
	OptsStirDestinationTn      = "*stirDestinationTn"
	OptsStirDestinationURI     = "*stirDestinationURI"
	OptsStirPublicKeyPath      = "*stirPublicKeyPath"
	OptsStirPrivateKeyPath     = "*stirPrivateKeyPath"
	// DispatcherS
	OptsAPIKey  = "*apiKey"
	OptsRouteID = "*routeID"
	// EEs
	OptsEEsVerbose = "*eesVerbose"
	// EEs Elasticsearch options
	ElsIndex               = "index"
	ElsIfPrimaryTerm       = "if_primary_term"
	ElsIfSeqNo             = "if_seq_no"
	ElsOpType              = "op_type"
	ElsPipeline            = "pipeline"
	ElsRequireAlias        = "require_alias"
	ElsRouting             = "routing"
	ElsTimeout             = "timeout"
	ElsVersionLow          = "version"
	ElsVersionType         = "version_type"
	ElsWaitForActiveShards = "wait_for_active_shards"
	// Others
	OptsContext               = "*context"
	Subsys                    = "*subsys"
	OptsAttributesProcessRuns = "*processRuns"
	OptsDispatcherMethod      = "*method"
	MetaEventType             = "*eventType"
	EventType                 = "EventType"
)

// Event Flags
const (
	MetaDerivedReply = "*derived_reply"

	MetaIDs = "*ids"

	TrueStr  = "true"
	FalseStr = "false"
)

// ArgCache constats
const (
	DestinationIDs               = "DestinationIDs"
	ReverseDestinationIDs        = "ReverseDestinationIDs"
	RatingPlanIDs                = "RatingPlanIDs"
	RatingProfileIDs             = "RatingProfileIDs"
	ActionIDs                    = "ActionIDs"
	ActionPlanIDs                = "ActionPlanIDs"
	AccountActionPlanIDs         = "AccountActionPlanIDs"
	ActionTriggerIDs             = "ActionTriggerIDs"
	SharedGroupIDs               = "SharedGroupIDs"
	ResourceProfileIDs           = "ResourceProfileIDs"
	ResourceIDs                  = "ResourceIDs"
	StatsQueueIDs                = "StatsQueueIDs"
	StatsQueueProfileIDs         = "StatsQueueProfileIDs"
	ThresholdIDs                 = "ThresholdIDs"
	ThresholdProfileIDs          = "ThresholdProfileIDs"
	FilterIDs                    = "FilterIDs"
	RouteProfileIDs              = "RouteProfileIDs"
	AttributeProfileIDs          = "AttributeProfileIDs"
	ChargerProfileIDs            = "ChargerProfileIDs"
	DispatcherProfileIDs         = "DispatcherProfileIDs"
	DispatcherHostIDs            = "DispatcherHostIDs"
	DispatcherRoutesIDs          = "DispatcherRoutesIDs"
	RateProfileIDs               = "RateProfileIDs"
	ActionProfileIDs             = "ActionProfileIDs"
	TimingIDs                    = "TimingIDs"
	AttributeFilterIndexIDs      = "AttributeFilterIndexIDs"
	ResourceFilterIndexIDs       = "ResourceFilterIndexIDs"
	StatFilterIndexIDs           = "StatFilterIndexIDs"
	ThresholdFilterIndexIDs      = "ThresholdFilterIndexIDs"
	RouteFilterIndexIDs          = "RouteFilterIndexIDs"
	ChargerFilterIndexIDs        = "ChargerFilterIndexIDs"
	DispatcherFilterIndexIDs     = "DispatcherFilterIndexIDs"
	RateProfilesFilterIndexIDs   = "RateProfilesFilterIndexIDs"
	RateFilterIndexIDs           = "RateFilterIndexIDs"
	ActionProfilesFilterIndexIDs = "ActionProfilesFilterIndexIDs"
	FilterIndexIDs               = "FilterIndexIDs"
)

// Poster and Event reader constants
const (
	SQSPoster     = "SQSPoster"
	S3Poster      = "S3Poster"
	AWSRegion     = "awsRegion"
	AWSKey        = "awsKey"
	AWSSecret     = "awsSecret"
	AWSToken      = "awsToken"
	AWSFolderPath = "folderPath"
	KafkaTopic    = "topic"
	KafkaGroupID  = "groupID"
	KafkaMaxWait  = "maxWait"

	// General constants for posters
	DefaultQueueID      = "cgrates_cdrs"
	QueueID             = "queueID"
	DefaultExchangeType = "direct"
	Exchange            = "exchange"
	ExchangeType        = "exchangeType"
	RoutingKey          = "routingKey"

	// for ers:
	AMQPDefaultConsumerTag = "cgrates"
	AMQPConsumerTag        = "consumerTag"

	KafkaDefaultTopic   = "cgrates"
	KafkaDefaultGroupID = "cgrates"
	KafkaDefaultMaxWait = time.Millisecond

	SQLDBName         = "dbName"
	SQLTableName      = "tableName"
	SQLSSLMode        = "sslmode"
	SQLDefaultSSLMode = "disable"
	SQLDefaultDBName  = "cgrates"

	ProcessedOpt = "Processed"
)

// Analyzers constants
const (
	MetaScorch  = "*scorch"
	MetaBoltdb  = "*boltdb"
	MetaLeveldb = "*leveldb"
	MetaMoss    = "*mossdb"

	RequestStartTime = "RequestStartTime"
	RequestDuration  = "RequestDuration"
	RequestParams    = "RequestParams"
	Reply            = "Reply"
	ReplyError       = "ReplyError"
	AnzDBDir         = "db"
	Opts             = "Opts"
)

var (
	// AnzIndexType are the analyzers possible index types
	AnzIndexType = StringSet{
		MetaScorch:  {},
		MetaBoltdb:  {},
		MetaLeveldb: {},
		MetaMoss:    {},
	}
)

func buildCacheInstRevPrefixes() {
	CachePrefixToInstance = make(map[string]string)
	for k, v := range CacheInstanceToPrefix {
		CachePrefixToInstance[v] = k
	}
	CacheInstanceToArg = make(map[string]string)
	for k, v := range ArgCacheToInstance {
		CacheInstanceToArg[v] = k
	}
}

func init() {
	buildCacheInstRevPrefixes()
}
