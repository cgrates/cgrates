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
		AccountField, Subject, Destination, SetupTime, AnswerTime, Usage, Cost, Rated, Partial, RunID,
		PreRated, CostSource, CostDetails, ExtraInfo, OrderID})
	PostPaidRatedSlice = []string{MetaPostpaid, MetaRated}

	GitCommitDate string // If set, it will be processed as part of versioning
	GitCommitHash string // If set, it will be processed as part of versioning

	extraDBPartition = NewStringSet([]string{CacheDispatchers,
		CacheDispatcherRoutes, CacheDispatcherLoads, CacheDiameterMessages, CacheRadiusPackets, CacheRPCResponses, CacheClosedSessions,
		CacheCDRIDs, CacheRPCConnections, CacheUCH, CacheSTIR, CacheEventCharges, MetaAPIBan, MetaSentryPeer,
		CacheRatingProfilesTmp, CacheCapsEvents, CacheReplicationHosts})

	DataDBPartitions = NewStringSet([]string{CacheDestinations, CacheReverseDestinations, CacheRatingPlans,
		CacheRatingProfiles, CacheDispatcherProfiles, CacheDispatcherHosts, CacheChargerProfiles, CacheActions, CacheActionTriggers, CacheSharedGroups, CacheTimings,
		CacheResourceProfiles, CacheResources, CacheEventResources, CacheStatQueueProfiles, CacheRankingProfiles, CacheStatQueues,
		CacheThresholdProfiles, CacheThresholds, CacheFilters, CacheRouteProfiles, CacheAttributeProfiles, CacheTrendProfiles, CacheTrends,
		CacheResourceFilterIndexes, CacheStatFilterIndexes, CacheThresholdFilterIndexes, CacheRouteFilterIndexes,
		CacheAttributeFilterIndexes, CacheChargerFilterIndexes, CacheDispatcherFilterIndexes, CacheLoadIDs,
		CacheReverseFilterIndexes, CacheActionPlans, CacheAccountActionPlans, CacheAccounts, CacheVersions})

	StorDBPartitions = NewStringSet([]string{CacheTBLTPTimings, CacheTBLTPDestinations, CacheTBLTPRates, CacheTBLTPDestinationRates,
		CacheTBLTPRatingPlans, CacheTBLTPRatingProfiles, CacheTBLTPSharedGroups, CacheTBLTPActions,
		CacheTBLTPActionPlans, CacheTBLTPActionTriggers, CacheTBLTPAccountActions, CacheTBLTPResources,
		CacheTBLTPStats, CacheTBLTPThresholds, CacheTBLTPRankings, CacheTBLTPFilters, CacheSessionCostsTBL, CacheCDRsTBL,
		CacheTBLTPRoutes, CacheTBLTPAttributes, CacheTBLTPChargers, CacheTBLTPDispatchers,
		CacheTBLTPDispatcherHosts, CacheVersions})

	// CachePartitions enables creation of cache partitions
	CachePartitions = JoinStringSet(extraDBPartition, DataDBPartitions)

	CacheInstanceToPrefix = map[string]string{
		CacheDestinations:            DestinationPrefix,
		CacheReverseDestinations:     ReverseDestinationPrefix,
		CacheRatingPlans:             RatingPlanPrefix,
		CacheRatingProfiles:          RatingProfilePrefix,
		CacheActions:                 ActionPrefix,
		CacheActionPlans:             ActionPlanPrefix,
		CacheAccountActionPlans:      AccountActionPlansPrefix,
		CacheActionTriggers:          ActionTriggerPrefix,
		CacheSharedGroups:            SharedGroupPrefix,
		CacheResourceProfiles:        ResourceProfilesPrefix,
		CacheResources:               ResourcesPrefix,
		CacheTimings:                 TimingsPrefix,
		CacheStatQueueProfiles:       StatQueueProfilePrefix,
		CacheStatQueues:              StatQueuePrefix,
		CacheRankingProfiles:         RankingsProfilePrefix,
		CacheTrendProfiles:           TrendsProfilePrefix,
		CacheTrends:                  TrendPrefix,
		CacheThresholdProfiles:       ThresholdProfilePrefix,
		CacheThresholds:              ThresholdPrefix,
		CacheFilters:                 FilterPrefix,
		CacheRouteProfiles:           RouteProfilePrefix,
		CacheAttributeProfiles:       AttributeProfilePrefix,
		CacheChargerProfiles:         ChargerProfilePrefix,
		CacheDispatcherProfiles:      DispatcherProfilePrefix,
		CacheDispatcherHosts:         DispatcherHostPrefix,
		CacheResourceFilterIndexes:   ResourceFilterIndexes,
		CacheStatFilterIndexes:       StatFilterIndexes,
		CacheThresholdFilterIndexes:  ThresholdFilterIndexes,
		CacheRouteFilterIndexes:      RouteFilterIndexes,
		CacheAttributeFilterIndexes:  AttributeFilterIndexes,
		CacheChargerFilterIndexes:    ChargerFilterIndexes,
		CacheDispatcherFilterIndexes: DispatcherFilterIndexes,

		CacheLoadIDs:              LoadIDPrefix,
		CacheAccounts:             AccountPrefix,
		CacheReverseFilterIndexes: FilterIndexPrfx,
		MetaSentryPeer:            MetaSentryPeer,
		MetaAPIBan:                MetaAPIBan, // special case as it is not in a DB
		CacheDispatchers:          MetaDispatchers,
	}
	CachePrefixToInstance map[string]string    // will be built on init
	CacheIndexesToPrefix  = map[string]string{ // used by match index to get all the ids when index selects is disabled and for compute indexes
		CacheThresholdFilterIndexes:  ThresholdProfilePrefix,
		CacheResourceFilterIndexes:   ResourceProfilesPrefix,
		CacheStatFilterIndexes:       StatQueueProfilePrefix,
		CacheRouteFilterIndexes:      RouteProfilePrefix,
		CacheAttributeFilterIndexes:  AttributeProfilePrefix,
		CacheChargerFilterIndexes:    ChargerProfilePrefix,
		CacheDispatcherFilterIndexes: DispatcherProfilePrefix,
		CacheReverseFilterIndexes:    FilterPrefix,
	}

	CacheInstanceToCacheIndex = map[string]string{
		CacheThresholdProfiles:  CacheThresholdFilterIndexes,
		CacheResourceProfiles:   CacheResourceFilterIndexes,
		CacheStatQueueProfiles:  CacheStatFilterIndexes,
		CacheRouteProfiles:      CacheRouteFilterIndexes,
		CacheAttributeProfiles:  CacheAttributeFilterIndexes,
		CacheChargerProfiles:    CacheChargerFilterIndexes,
		CacheDispatcherProfiles: CacheDispatcherFilterIndexes,
		CacheFilters:            CacheReverseFilterIndexes,
	}

	// NonMonetaryBalances are types of balances which are not handled as monetary
	NonMonetaryBalances = NewStringSet([]string{MetaVoice, MetaSMS, MetaData, MetaGeneric})

	// AccountableRequestTypes are the ones handled by Accounting subsystem
	AccountableRequestTypes = NewStringSet([]string{MetaPrepaid, MetaPostpaid, MetaPseudoPrepaid})

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
		TBLTPTrends:           CacheTBLTPTrends,
		TBLTPRankings:         CacheTBLTPRankings,
		TBLTPThresholds:       CacheTBLTPThresholds,
		TBLTPFilters:          CacheTBLTPFilters,
		SessionCostsTBL:       CacheSessionCostsTBL,
		CDRsTBL:               CacheCDRsTBL,
		TBLTPRoutes:           CacheTBLTPRoutes,
		TBLTPAttributes:       CacheTBLTPAttributes,
		TBLTPChargers:         CacheTBLTPChargers,
		TBLTPDispatchers:      CacheTBLTPDispatchers,
		TBLTPDispatcherHosts:  CacheTBLTPDispatcherHosts,
	}

	// ProtectedSFlds are the fields that sessions should not alter
	ProtectedSFlds = NewStringSet([]string{CGRID, OriginHost, OriginID, Usage})
)

const (
	CGRateS                  = "CGRateS"
	Version                  = "v0.11.0~dev"
	DiameterFirmwareRevision = 918
	CGRateSLwr               = "cgrates"
	Postgres                 = "postgres"
	MySQL                    = "mysql"
	Mongo                    = "mongo"
	Redis                    = "redis"
	Internal                 = "internal"
	DataManager              = "DataManager"
	Localhost                = "127.0.0.1"
	Prepaid                  = "prepaid"
	MetaPrepaid              = "*prepaid"
	Postpaid                 = "postpaid"
	MetaPostpaid             = "*postpaid"
	PseudoPrepaid            = "pseudoprepaid"
	MetaPseudoPrepaid        = "*pseudoprepaid"
	MetaRated                = "*rated"
	MetaNone                 = "*none"
	MetaNow                  = "*now"
	MetaRoundingUp           = "*up"
	MetaRoundingMiddle       = "*middle"
	MetaRoundingDown         = "*down"
	MetaAny                  = "*any"
	MetaAll                  = "*all"
	MetaSingle               = "*single"
	MetaZero                 = "*zero"
	MetaASAP                 = "*asap"
	MetaNil                  = "*nil"
	MetaSpace                = "*space"
	MetaChar                 = "*char"
	CommentChar              = '#'
	CSVSep                   = ','
	FallbackSep              = ';'
	InfieldSep               = ";"
	MetaPipe                 = "*|"
	FieldsSep                = ","
	InInFieldSep             = ":"
	StaticHDRValSep          = "::"
	FilterValStart           = "("
	FilterValEnd             = ")"
	PlusChar                 = "+"
	JSON                     = "json"
	JSONCaps                 = "JSON"
	GOBCaps                  = "GOB"
	MsgPack                  = "msgpack"
	CSVLoad                  = "CSVLOAD"
	CGRID                    = "CGRID"
	ToR                      = "ToR"
	OrderID                  = "OrderID"
	OriginID                 = "OriginID"
	InitialOriginID          = "InitialOriginID"
	OriginIDPrefix           = "OriginIDPrefix"
	Source                   = "Source"
	OriginHost               = "OriginHost"
	RequestType              = "RequestType"
	Direction                = "Direction"
	Tenant                   = "Tenant"
	Category                 = "Category"
	Contexts                 = "Contexts"
	AccountField             = "Account"
	BalancesFld              = "Balances"
	Subject                  = "Subject"
	Destination              = "Destination"
	SetupTime                = "SetupTime"
	AnswerTime               = "AnswerTime"
	Usage                    = "Usage"
	DurationIndex            = "DurationIndex"
	MaxRateUnit              = "MaxRateUnit"
	DebitInterval            = "DebitInterval"
	TimeStart                = "TimeStart"
	TimeEnd                  = "TimeEnd"
	CallDuration             = "CallDuration"
	FallbackSubject          = "FallbackSubject"
	DryRun                   = "DryRun"

	CustomValue             = "CustomValue"
	Value                   = "Value"
	Filter                  = "Filter"
	LastUsed                = "LastUsed"
	PDD                     = "PDD"
	Route                   = "Route"
	RunID                   = "RunID"
	AttributeIDs            = "AttributeIDs"
	MetaReqRunID            = "*req.RunID"
	Cost                    = "Cost"
	CostDetails             = "CostDetails"
	EventCost               = "EventCost"
	EeIDs                   = "EeIDs"
	Rated                   = "rated"
	Partial                 = "Partial"
	PreRated                = "PreRated"
	StaticValuePrefix       = "^"
	CSV                     = "csv"
	FWV                     = "fwv"
	MetaCombimed            = "*combimed"
	MetaMongo               = "*mongo"
	MetaRedis               = "*redis"
	MetaPostgres            = "*postgres"
	MetaInternal            = "*internal"
	MetaLocalHost           = "*localhost"
	MetaBiJSONLocalHost     = "*bijson_localhost"
	MetaRatingSubjectPrefix = "*zero"
	OK                      = "OK"
	MetaFileXML             = "*file_xml"
	MetaFileJSON            = "*file_json"
	MaskChar                = "*"
	ConcatenatedKeySep      = ":"
	UnitTest                = "UNIT_TEST"
	HDRValSep               = "/"
	MetaMonetary            = "*monetary"
	MetaSMS                 = "*sms"
	MetaMMS                 = "*mms"
	MetaGeneric             = "*generic"
	MetaData                = "*data"
	MetaMaxCostFree         = "*free"
	MetaMaxCostDisconnect   = "*disconnect"
	MetaOut                 = "*out"
	MetaPause               = "*pause"

	MetaVoice                 = "*voice"
	ACD                       = "ACD"
	TasksKey                  = "tasks"
	ActionPlanPrefix          = "apl_"
	AccountActionPlansPrefix  = "aap_"
	ActionTriggerPrefix       = "atr_"
	RatingPlanPrefix          = "rpl_"
	RatingProfilePrefix       = "rpf_"
	ActionPrefix              = "act_"
	SharedGroupPrefix         = "shg_"
	AccountPrefix             = "acc_"
	DestinationPrefix         = "dst_"
	ReverseDestinationPrefix  = "rds_"
	DerivedChargersPrefix     = "dcs_"
	UsersPrefix               = "usr_"
	ResourcesPrefix           = "res_"
	ResourceProfilesPrefix    = "rsp_"
	ThresholdPrefix           = "thd_"
	TrendPrefix               = "trd_"
	RankingPrefix             = "rnk_"
	TimingsPrefix             = "tmg_"
	FilterPrefix              = "ftr_"
	CDRsStatsPrefix           = "cst_"
	VersionPrefix             = "ver_"
	StatQueueProfilePrefix    = "sqp_"
	RouteProfilePrefix        = "rpp_"
	RatePrefix                = "rep_"
	AttributeProfilePrefix    = "alp_"
	ChargerProfilePrefix      = "cpp_"
	DispatcherProfilePrefix   = "dpp_"
	DispatcherHostPrefix      = "dph_"
	ThresholdProfilePrefix    = "thp_"
	StatQueuePrefix           = "stq_"
	RankingsProfilePrefix     = "rgp_"
	TrendsProfilePrefix       = "trp_"
	LoadIDPrefix              = "lid_"
	SessionsBackupPrefix      = "sbk_"
	LoadInstKey               = "load_history"
	CreateCDRsTablesSQL       = "create_cdrs_tables.sql"
	CreateTariffPlanTablesSQL = "create_tariffplan_tables.sql"
	TestSQL                   = "TEST_SQL"
	MetaConstant              = "*constant"
	MetaPositive              = "*positive"
	MetaNegative              = "*negative"
	MetaLast                  = "*last"

	MetaFiller                = "*filler"
	MetaHTTPPost              = "*http_post"
	MetaHTTPjsonMap           = "*http_json_map"
	MetaAMQPjsonMap           = "*amqp_json_map"
	MetaAMQPV1jsonMap         = "*amqpv1_json_map"
	MetaRPC                   = "*rpc"
	MetaSQSjsonMap            = "*sqs_json_map"
	MetaKafkajsonMap          = "*kafka_json_map"
	MetaNatsjsonMap           = "*nats_json_map"
	MetaSQL                   = "*sql"
	MetaMySQL                 = "*mysql"
	MetaS3jsonMap             = "*s3_json_map"
	ConfigPath                = "/etc/cgrates/"
	DisconnectCause           = "DisconnectCause"
	MetaRating                = "*rating"
	NotAvailable              = "N/A"
	Call                      = "call"
	ExtraFields               = "ExtraFields"
	SourceBalanceSummary      = "SourceBalanceSummary"
	DestinationBalanceSummary = "DestinationBalanceSummary"
	MetaSureTax               = "*sure_tax"
	MetaDynamic               = "*dynamic"
	MetaCounterEvent          = "*event"
	MetaBalance               = "*balance"
	MetaAccount               = "*account"
	EventName                 = "EventName"
	// action trigger threshold types
	TriggerMinEventCounter   = "*min_event_counter"
	TriggerMaxEventCounter   = "*max_event_counter"
	TriggerMaxBalanceCounter = "*max_balance_counter"
	TriggerMinBalance        = "*min_balance"
	TriggerMaxBalance        = "*max_balance"
	TriggerBalanceExpired    = "*balance_expired"
	HierarchySep             = ">"
	MetaComposed             = "*composed"
	MetaUsageDifference      = "*usage_difference"
	MetaDifference           = "*difference"
	MetaVariable             = "*variable"
	MetaCCUsage              = "*cc_usage"
	MetaSIPCID               = "*sipcid"
	MetaValueExponent        = "*value_exponent"
	NegativePrefix           = "!"
	MatchStartPrefix         = "^"
	MatchGreaterThanOrEqual  = ">="
	MatchLessThanOrEqual     = "<="
	MatchGreaterThan         = ">"
	MatchLessThan            = "<"
	MatchEndPrefix           = "$"
	MetaRaw                  = "*raw"
	CreatedAt                = "CreatedAt"
	UpdatedAt                = "UpdatedAt"
	HandlerSubstractUsage    = "*substract_usage"
	XML                      = "xml"
	MetaGOB                  = "*gob"
	MetaJSON                 = "*json"
	MetaMSGPACK              = "*msgpack"
	MetaDateTime             = "*datetime"
	MetaMaskedDestination    = "*masked_destination"
	MetaUnixTimestamp        = "*unix_timestamp"
	MetaPostCDR              = "*post_cdr"
	MetaDumpToFile           = "*dump_to_file"
	MetaDumpToJSON           = "*dump_to_json"
	NonTransactional         = ""
	DataDB                   = "data_db"
	StorDB                   = "stor_db"
	NotFoundCaps             = "NOT_FOUND"
	ServerErrorCaps          = "SERVER_ERROR"
	MandatoryIEMissingCaps   = "MANDATORY_IE_MISSING"
	UnsupportedCachePrefix   = "unsupported cache prefix"
	CDRsCtx                  = "cdrs"
	MandatoryInfoMissing     = "mandatory information missing"
	UnsupportedServiceIDCaps = "UNSUPPORTED_SERVICE_ID"
	ServiceManager           = "service_manager"
	ServiceAlreadyRunning    = "service already running"
	RunningCaps              = "RUNNING"
	StoppedCaps              = "STOPPED"
	SchedulerNotRunningCaps  = "SCHEDULER_NOT_RUNNING"
	MetaScheduler            = "*scheduler"
	MetaSessionsCosts        = "*sessions_costs"
	MetaRALs                 = "*rals"
	MetaReplicator           = "*replicator"
	MetaRerate               = "*rerate"
	MetaRefund               = "*refund"
	MetaStats                = "*stats"
	MetaTrends               = "*trends"
	MetaRankings             = "*rankings"
	MetaResponder            = "*responder"
	MetaCore                 = "*core"
	MetaServiceManager       = "*servicemanager"
	MetaChargers             = "*chargers"
	MetaReprocess            = "*reprocess"
	MetaBlockerError         = "*blocker_error"
	MetaConfig               = "*config"
	MetaDispatchers          = "*dispatchers"
	MetaRegistrarC           = "*registrarc"
	MetaDispatcherHosts      = "*dispatcher_hosts"
	MetaFilters              = "*filters"
	MetaCDRs                 = "*cdrs"
	MetaDC                   = "*dc"
	MetaCaches               = "*caches"
	MetaUCH                  = "*uch"
	MetaGuardian             = "*guardians"
	MetaEEs                  = "*ees"
	MetaERs                  = "*ers"
	MetaContinue             = "*continue"
	Migrator                 = "migrator"
	UnsupportedMigrationTask = "unsupported migration task"
	NoStorDBConnection       = "not connected to StorDB"
	UndefinedVersion         = "undefined version"
	TxtSuffix                = ".txt"
	JSNSuffix                = ".json"
	GOBSuffix                = ".gob"
	XMLSuffix                = ".xml"
	CSVSuffix                = ".csv"
	FWVSuffix                = ".fwv"
	ContentJSON              = "json"
	ContentForm              = "form"
	FileLockPrefix           = "file_"
	ActionsPoster            = "act"
	CDRPoster                = "cdr"
	MetaFileCSV              = "*file_csv"
	MetaVirt                 = "*virt"
	MetaElastic              = "*els"
	MetaFileFWV              = "*file_fwv"
	MetaFile                 = "*file"
	Accounts                 = "Accounts"
	AccountService           = "AccountS"
	AccountS                 = "AccountS"
	Actions                  = "Actions"
	ActionPlans              = "ActionPlans"
	ActionTriggers           = "ActionTriggers"
	BalanceMap               = "BalanceMap"
	CounterType              = "CounterType"
	Counters                 = "Counters"
	UnitCounters             = "UnitCounters"
	UpdateTime               = "UpdateTime"
	SharedGroups             = "SharedGroups"
	Timings                  = "Timings"
	Rates                    = "Rates"
	DestinationRates         = "DestinationRates"
	RatingPlans              = "RatingPlans"
	RatingProfiles           = "RatingProfiles"
	AccountActions           = "AccountActions"
	Resources                = "Resources"
	Stats                    = "Stats"
	Trends                   = "Trends"
	Filters                  = "Filters"
	DispatcherProfiles       = "DispatcherProfiles"
	DispatcherHosts          = "DispatcherHosts"
	MetaEveryMinute          = "*every_minute"
	MetaHourly               = "*hourly"
	ID                       = "ID"
	UniqueID                 = "UniqueID"
	Address                  = "Address"
	Addresses                = "Addresses"
	Transport                = "Transport"
	TLS                      = "TLS"
	Subsystems               = "Subsystems"
	Strategy                 = "Strategy"
	StrategyParameters       = "StrategyParameters"
	ConnID                   = "ConnID"
	ConnFilterIDs            = "ConnFilterIDs"
	ConnWeight               = "ConnWeight"
	ConnBlocker              = "ConnBlocker"
	ConnParameters           = "ConnParameters"

	Thresholds            = "Thresholds"
	Routes                = "Routes"
	Attributes            = "Attributes"
	Chargers              = "Chargers"
	Dispatchers           = "Dispatchers"
	StatS                 = "Stats"
	LoadIDsVrs            = "LoadIDs"
	RALService            = "RALs"
	GlobalVarS            = "GlobalVarS"
	CostSource            = "CostSource"
	ExtraInfo             = "ExtraInfo"
	Meta                  = "*"
	MetaSysLog            = "*syslog"
	MetaStdLog            = "*stdout"
	EventSource           = "EventSource"
	AccountID             = "AccountID"
	AccountIDs            = "AccountIDs"
	SourceAccountID       = "SourceAccountID"
	DestinationAccountID  = "DestinationAccountID"
	SourceBalanceID       = "SourceBalanceID"
	DestinationBalanceID  = "DestinationBalanceID"
	ResourceID            = "ResourceID"
	TotalUsage            = "TotalUsage"
	StatID                = "StatID"
	StatIDs               = "StatIDs"
	SortedStatIDs         = "SortedStatIDs"
	LastUpdate            = "LastUpdate"
	TrendID               = "TrendID"
	RankingID             = "RankingID"
	BalanceType           = "BalanceType"
	BalanceID             = "BalanceID"
	BalanceDestinationIds = "BalanceDestinationIds"
	BalanceWeight         = "BalanceWeight"
	BalanceExpirationDate = "BalanceExpirationDate"
	BalanceTimingTags     = "BalanceTimingTags"
	BalanceRatingSubject  = "BalanceRatingSubject"
	BalanceCategories     = "BalanceCategories"
	BalanceSharedGroups   = "BalanceSharedGroups"
	BalanceBlocker        = "BalanceBlocker"
	BalanceDisabled       = "BalanceDisabled"
	BalanceFactorID       = "BalanceFactorID"
	Units                 = "Units"
	AccountUpdate         = "AccountUpdate"
	StatUpdate            = "StatUpdate"
	TrendUpdate           = "TrendUpdate"
	RankingUpdate         = "RankingUpdate"
	ResourceUpdate        = "ResourceUpdate"
	CDR                   = "CDR"
	CDRs                  = "CDRs"
	ExpiryTime            = "ExpiryTime"
	AllowNegative         = "AllowNegative"
	Disabled              = "Disabled"
	Initial               = "Initial"
	Action                = "Action"

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
	Limit                    = "Limit"
	UsageTTL                 = "UsageTTL"
	AllocationMessage        = "AllocationMessage"
	Stored                   = "Stored"
	RatingSubject            = "RatingSubject"
	Categories               = "Categories"
	Blocker                  = "Blocker"
	RatingPlanID             = "RatingPlanID"
	StartTime                = "StartTime"
	EndTime                  = "EndTime"
	AccountSummary           = "AccountSummary"
	RatingFilters            = "RatingFilters"
	RatingFilter             = "RatingFilter"
	Accounting               = "Accounting"
	Rating                   = "Rating"
	Charges                  = "Charges"
	CompressFactor           = "CompressFactor"
	Increments               = "Increments"
	BalanceField             = "Balance"
	BalanceSummaries         = "BalanceSummaries"
	ExtraCharge              = "ExtraCharge"
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
	BalanceFactor            = "BalanceFactor"
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
	CgrAcd                   = "cgr_acd"
	ActivationIntervalString = "ActivationInterval"
	MaxHits                  = "MaxHits"
	MinHits                  = "MinHits"
	Async                    = "Async"
	Sorting                  = "Sorting"
	SortingParameters        = "SortingParameters"
	RouteAccountIDs          = "RouteAccountIDs"
	RouteRatingplanIDs       = "RouteRatingplanIDs"
	RouteStatIDs             = "RouteStatIDs"
	RouteWeight              = "RouteWeight"
	RouteParameters          = "RouteParameters"
	RouteBlocker             = "RouteBlocker"
	RouteResourceIDs         = "RouteResourceIDs"
	RouteFilterIDs           = "RouteFilterIDs"
	AttributeFilterIDs       = "AttributeFilterIDs"
	QueueLength              = "QueueLength"
	CorrelationType          = "CorrelationType"
	Tolerance                = "Tolerance"
	TTL                      = "TTL"
	MinItems                 = "MinItems"
	MetricIDs                = "MetricIDs"
	Metrics                  = "Metrics"
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
	MetaRoutesMaxCost        = "*routes_maxcost"
	MetaMaxCost              = "*maxcost"
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
	MetaOReq                 = "*oreq"
	MetaReq                  = "*req"
	MetaAsm                  = "*asm"
	MetaVars                 = "*vars"
	MetaRep                  = "*rep"
	MetaExp                  = "*exp"
	MetaHdr                  = "*hdr"
	MetaTrl                  = "*trl"
	MetaTmp                  = "*tmp"
	MetaOpts                 = "*opts"
	MetaCfg                  = "*cfg"
	MetaDynReq               = "~*req"
	MetaScPrefix             = "~*sc." // used for SMCostFilter
	MetaEventTimestamp       = "*eventTimestamp"
	CGROriginHost            = "cgr_originhost"
	MetaInitiate             = "*initiate"
	MetaUpdate               = "*update"
	MetaTerminate            = "*terminate"
	MetaEvent                = "*event"
	MetaMessage              = "*message"
	MetaDryRun               = "*dryrun"
	MetaRALsDryRun           = "*ralsDryRun"
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
	MetaCmd                  = "*cmd"
	MetaEnv                  = "*env:" // use in config for describing enviormant variables
	MetaTemplate             = "*template"
	MetaCCA                  = "*cca"
	MetaErr                  = "*err"
	OriginRealm              = "OriginRealm"
	ProductName              = "ProductName"
	IdxStart                 = "["
	IdxEnd                   = "]"
	IdxCombination           = "]["

	// *vars consts in agents
	MetaAppID          = "*appid"
	MetaSessionID      = "*sessionID" // used to retrieve RADIUS Access-Reqest packets of active sessions
	JanusAdminSubProto = "janus-admin-protocol"

	RemoteHost              = "RemoteHost"
	Local                   = "local"
	TCP                     = "tcp"
	UDP                     = "udp"
	VersionName             = "Version"
	MetaTenant              = "*tenant"
	ResourceUsage           = "ResourceUsage"
	MetaStrip               = "*strip"
	MetaDuration            = "*duration"
	MetaDurationFormat      = "*durfmt"
	MetaLibPhoneNumber      = "*libphonenumber"
	MetaTimeString          = "*time_string"
	MetaIP2Hex              = "*ip2hex"
	MetaString2Hex          = "*string2hex"
	MetaUnixTime            = "*unixtime"
	MetaLen                 = "*len"
	MetaSlice               = "*slice"
	MetaSIPURIMethod        = "*sipuri_method"
	MetaSIPURIHost          = "*sipuri_host"
	MetaSIPURIUser          = "*sipuri_user"
	E164DomainConverter     = "*e164Domain"
	E164Converter           = "*e164"
	URLDecConverter         = "*urldecode"
	URLEncConverter         = "*urlencode"
	MetaReload              = "*reload"
	MetaLoad                = "*load"
	MetaFloat64             = "*float64"
	MetaRemove              = "*remove"
	MetaRemoveAll           = "*removeall"
	MetaStore               = "*store"
	MetaClear               = "*clear"
	MetaExport              = "*export"
	MetaExporterID          = "*exporterID"
	MetaTimeNow             = "*time_now"
	MetaFirstEventATime     = "*first_event_atime"
	MetaLastEventATime      = "*last_event_atime"
	MetaEventNumber         = "*event_number"
	LoadIDs                 = "load_ids"
	DNSAgent                = "DNSAgent"
	InitS                   = "InitS"
	TLSNoCaps               = "tls"
	UsageID                 = "UsageID"
	Replacement             = "Replacement"
	Regexp                  = "Regexp"
	Order                   = "Order"
	Preference              = "Preference"
	Flags                   = "Flags"
	Service                 = "Service"
	ApierV                  = "ApierV"
	MetaApier               = "*apier"
	MetaAnalyzer            = "*analyzer"
	CGREventString          = "CGREvent"
	MetaTextPlain           = "*text_plain"
	MetaIgnoreErrors        = "*ignore_errors"
	MetaRelease             = "*release"
	MetaAllocate            = "*allocate"
	MetaAuthorize           = "*authorize"
	MetaSTIRAuthenticate    = "*stir_authenticate"
	MetaSTIRInitiate        = "*stir_initiate"
	MetaInit                = "*init"
	MetaRatingPlanCost      = "*rating_plan_cost"
	ERs                     = "ERs"
	EEs                     = "EEs"
	Ratio                   = "Ratio"
	Load                    = "Load"
	Slash                   = "/"
	UUID                    = "UUID"
	Uuid                    = "Uuid"
	ActionsID               = "ActionsID"
	MetaAct                 = "*act"
	MetaAcnt                = "*acnt"
	DestinationPrefixName   = "DestinationPrefix"
	DestinationID           = "DestinationID"
	ExportTemplate          = "ExportTemplate"
	ExportFormat            = "ExportFormat"
	Synchronous             = "Synchronous"
	Attempts                = "Attempts"
	FieldSeparator          = "FieldSeparator"
	ExportPath              = "ExportPath"
	ExporterIDs             = "ExporterIDs"
	TimeNow                 = "TimeNow"
	ExportFileName          = "ExportFileName"
	GroupID                 = "GroupID"
	ThresholdType           = "ThresholdType"
	ThresholdValue          = "ThresholdValue"
	Recurrent               = "Recurrent"
	Executed                = "Executed"
	LastExecutionTime       = "LastExecutionTime"
	MinSleep                = "MinSleep"
	ActivationDate          = "ActivationDate"
	ExpirationDate          = "ExpirationDate"
	MinQueuedItems          = "MinQueuedItems"
	OrderIDStart            = "OrderIDStart"
	OrderIDEnd              = "OrderIDEnd"
	MinCost                 = "MinCost"
	MaxCost                 = "MaxCost"
	MetaLoaders             = "*loaders"
	TmpSuffix               = ".tmp"
	MetaDiamreq             = "*diamreq"
	MetaRadDAReq            = "*radDAReq"
	MetaRadCoATemplate      = "*radCoATemplate"
	MetaRadDMRTemplate      = "*radDMRTemplate"
	MetaCost                = "*cost"
	MetaGroup               = "*group"
	InternalRPCSet          = "InternalRPCSet"
	MetaFileName            = "*fileName"
	MetaFileLineNumber      = "*fileLineNumber"
	MetaReaderID            = "*readerID"
	MetaRadauth             = "*radauth"
	UserPassword            = "UserPassword"
	RadauthFailed           = "RADAUTH_FAILED"
	MetaPAP                 = "*pap"
	MetaCHAP                = "*chap"
	MetaMSCHAPV2            = "*mschapv2"
	MetaDynaprepaid         = "*dynaprepaid"
	MetaFD                  = "*fd"
	SortingData             = "SortingData"
	Count                   = "Count"
	ProfileID               = "ProfileID"
	SortedRoutes            = "SortedRoutes"
	MetaMonthly             = "*monthly"
	MetaYearly              = "*yearly"
	MetaDaily               = "*daily"
	MetaWeekly              = "*weekly"
	Underline               = "_"
	MetaPartial             = "*partial"
	MetaBusy                = "*busy"
	MetaQueue               = "*queue"
	MetaMonthEnd            = "*month_end"
	APIKey                  = "ApiKey"
	RouteID                 = "RouteID"
	MetaMonthlyEstimated    = "*monthly_estimated"
	MetaProcessedProfileIDs = "*processedProfileIDs"
	MetaAttrPrfTenantID     = "*apTenantID"
	HashtagSep              = "#"
	MetaRounding            = "*rounding"
	StatsNA                 = -1.0
	InvalidUsage            = -1
	InvalidDuration         = time.Duration(-1)
	Schedule                = "Schedule"
	ActionFilterIDs         = "ActionFilterIDs"
	ActionBlocker           = "ActionBlocker"
	ActionTTL               = "ActionTTL"
	ActionOpts              = "ActionOpts"
	ActionPath              = "ActionPath"
	TPid                    = "TPid"
	LoadId                  = "LoadId"
	ActionPlanId            = "ActionPlanId"
	AccountActionsId        = "AccountActionsId"
	Loadid                  = "loadid"
	AccountLowerCase        = "account"
	ActionPlan              = "ActionPlan"
	ActionsId               = "ActionsId"
	TimingId                = "TimingId"
	Prefixes                = "Prefixes"
	RateSlots               = "RateSlots"
	RatingPlanBindings      = "RatingPlanBindings"
	RatingPlanActivations   = "RatingPlanActivations"
	CategoryLowerCase       = "category"
	SubjectLowerCase        = "subject"
	RatingProfileID         = "RatingProfileID"
	Time                    = "Time"
	TargetIDs               = "TargetIDs"
	TargetType              = "TargetType"
	MetaRow                 = "*row"
	BalanceFilterIDs        = "BalanceFilterIDs"
	BalanceOpts             = "BalanceOpts"
	MetaConcrete            = "*concrete"
	MetaAbstract            = "*abstract"
	MetaBalanceLimit        = "*balanceLimit"
	MetaBalanceUnlimited    = "*balanceUnlimited"
	MetaTemplateID          = "*templateID"
	MetaCdrLog              = "*cdrLog"
	MetaCDR                 = "*cdr"
	MetaExporterIDs         = "*exporterIDs"
	MetaAsync               = "*async"
	MetaUsage               = "*usage"
	Weights                 = "Weights"
	UnitFactors             = "UnitFactors"
	CostIncrements          = "CostIncrements"
	Factors                 = "Factors"
	Method                  = "Method"
	Static                  = "Static"
	Params                  = "Params"
	GetTotalValue           = "GetTotalValue"
	Increment               = "Increment"
	FixedFee                = "FixedFee"
	RecurrentFee            = "RecurrentFee"
	Diktats                 = "Diktats"
	BalanceIDs              = "BalanceIDs"
	MetaCostIncrement       = "*costIncrement"
	Length                  = "Length"

	// dns
	DNSQueryType          = "QueryType"
	DNSQueryName          = "QueryName"
	DNSOption             = "Option"
	DNSRcode              = "Rcode"
	DNSId                 = "Id"
	DNSResponse           = "Response"
	DNSOpcode             = "Opcode"
	DNSAuthoritative      = "Authoritative"
	DNSTruncated          = "Truncated"
	DNSRecursionDesired   = "RecursionDesired"
	DNSRecursionAvailable = "RecursionAvailable"
	DNSZero               = "Zero"
	DNSAuthenticatedData  = "AuthenticatedData"
	DNSCheckingDisabled   = "CheckingDisabled"
	DNSQuestion           = "Question"
	DNSAnswer             = "Answer"
	DNSNs                 = "Ns"
	DNSExtra              = "Extra"
	DNSName               = "Name"
	DNSQtype              = "Qtype"
	DNSQclass             = "Qclass"
	DNSFamily             = "Family"
	DNSSourceNetmask      = "SourceNetmask"
	DNSSourceScope        = "SourceScope"
	DNSLease              = "Lease"
	DNSKeyLease           = "KeyLease"
	DNSLeaseLife          = "LeaseLife"
	DNSTimeout            = "Timeout"
	DNSInfoCode           = "InfoCode"
	DNSExtraText          = "ExtraText"
	DNSNsid               = "Nsid"
	DNSCookie             = "Cookie"
	DNSDAU                = "DAU"
	DNSDHU                = "DHU"
	DNSN3U                = "N3U"
	DNSExpire             = "Expire"
	DNSPadding            = "Padding"
	DNSUri                = "Uri"
	DNSHdr                = "Hdr"
	DNSA                  = "A"
	DNSTarget             = "Target"
	DNSPriority           = "Priority"
	DNSPort               = "Port"
	DNSRrtype             = "Rrtype"
	DNSClass              = "Class"
	DNSTtl                = "Ttl"
	DNSRdlength           = "Rdlength"
	DNSData               = "Data"
)

// CoreSv1.Status metrics
const (
	PID            = "pid"
	NodeID         = "node_id"
	GoVersion      = "go_version"
	VersionLower   = "version"
	Goroutines     = "goroutines"
	OSThreadsInUse = "os_threads_in_use"
	CAPSAllocated  = "caps_allocated"
	CAPSPeak       = "caps_peak"
	RunningSince   = "running_since"
	OpenFiles      = "open_files"
	CPUTime        = "cpu_time"
	ActiveMemory   = "active_memory"
	SystemMemory   = "system_memory"
	ResidentMemory = "resident_memory"
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
	MetaRankingProfiles     = "*ranking_profiles"
	MetaTrendProfiles       = "*trend_profiles"
	MetaThresholdProfiles   = "*threshold_profiles"
	MetaRouteProfiles       = "*route_profiles"
	MetaAttributeProfiles   = "*attribute_profiles"
	MetaDispatcherProfiles  = "*dispatcher_profiles"
	MetaChargerProfiles     = "*charger_profiles"
	MetaSharedGroups        = "*shared_groups"
	MetaThresholds          = "*thresholds"
	MetaRoutes              = "*routes"
	MetaAttributes          = "*attributes"
	MetaSessionsBackup      = "*sessions_backup"
	MetaLoadIDs             = "*load_ids"
	MetaNodeID              = "*node_id"
	MetaAscending           = "*ascending"
	MetaDescending          = "*descending"
	MetaDesc                = "*desc"
	MetaAsc                 = "*asc"
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
	MetaDMR      = "*dmr"
	MetaCoA      = "*coa"
)

// Services
const (
	AnalyzerS   = "AnalyzerS"
	ApierS      = "ApierS"
	AttributeS  = "AttributeS"
	CacheS      = "CacheS"
	CDRServer   = "CDRServer"
	ChargerS    = "ChargerS"
	ConfigS     = "ConfigS"
	DispatcherS = "DispatcherS"
	EeS         = "EeS"
	ErS         = "ErS"
	FilterS     = "FilterS"
	GuardianS   = "GuardianS"
	LoaderS     = "LoaderS"
	RALs        = "RALs"
	RegistrarC  = "RegistrarC"
	ReplicatorS = "ReplicatorS"
	ResourceS   = "ResourceS"
	ResponderS  = "ResponderS"
	RouteS      = "RouteS"
	SessionS    = "SessionS"
	StatService = "StatS"
	TrendS      = "TrendS"
	RankingS    = "RankingS"
	ThresholdS  = "ThresholdS"
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
)

// Actions
const (
	MetaLog                     = "*log"
	MetaResetTriggers           = "*reset_triggers"
	MetaSetRecurrent            = "*set_recurrent"
	MetaUnsetRecurrent          = "*unset_recurrent"
	MetaAllowNegative           = "*allow_negative"
	MetaDenyNegative            = "*deny_negative"
	MetaResetAccount            = "*reset_account"
	MetaRemoveAccount           = "*remove_account"
	MetaRemoveBalance           = "*remove_balance"
	MetaTopUpReset              = "*topup_reset"
	MetaTopUp                   = "*topup"
	MetaDebitReset              = "*debit_reset"
	MetaDebit                   = "*debit"
	MetaTransferBalance         = "*transfer_balance"
	MetaResetCounters           = "*reset_counters"
	MetaEnableAccount           = "*enable_account"
	MetaDisableAccount          = "*disable_account"
	HttpPostAsync               = "*http_post_async"
	MetaMailAsync               = "*mail_async"
	MetaUnlimited               = "*unlimited"
	CDRLog                      = "*cdrlog"
	MetaSetDDestinations        = "*set_ddestinations"
	MetaTransferMonetaryDefault = "*transfer_monetary_default"
	MetaCgrRpc                  = "*cgr_rpc"
	MetaAlterSessions           = "*alter_sessions"
	MetaForceDisconnectSessions = "*force_disconnect_sessions"
	TopUpZeroNegative           = "*topup_zero_negative"
	SetExpiry                   = "*set_expiry"
	MetaPublishAccount          = "*publish_account"
	MetaRemoveSessionCosts      = "*remove_session_costs"
	MetaRemoveExpired           = "*remove_expired"
	MetaPostEvent               = "*post_event"
	MetaCDRAccount              = "*reset_account_cdr"
	MetaResetThreshold          = "*reset_threshold"
	MetaResetStatQueue          = "*reset_stat_queue"
	MetaRemoteSetAccount        = "*remote_set_account"
	ActionID                    = "ActionID"
	ActionType                  = "ActionType"
	ActionValue                 = "ActionValue"
	BalanceValue                = "BalanceValue"
	BalanceUnits                = "BalanceUnits"
	ExtraParameters             = "ExtraParameters"

	MetaAddBalance = "*add_balance"
	MetaSetBalance = "*set_balance"
	MetaRemBalance = "*rem_balance"
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
	CapRouteProfiles        = "RouteProfiles"
	CapThresholds           = "Thresholds"
	CapStatQueues           = "StatQueues"
)

// cgr-tester
const (
	CGRTester = "CGRTester"
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
	TpTrends             = "TpTrends"
	TpRankings           = "TpRankings"
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
	TrendSv1           = "TrendSv1"
	RankingSv1         = "RankingSv1"
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

// Filter types
const (
	MetaNot                = "*not"
	MetaString             = "*string"
	MetaPrefix             = "*prefix"
	MetaSuffix             = "*suffix"
	MetaBoth               = "*both"
	MetaEmpty              = "*empty"
	MetaExists             = "*exists"
	MetaTimings            = "*timings"
	MetaRSR                = "*rsr"
	MetaDestinations       = "*destinations"
	MetaLessThan           = "*lt"
	MetaLessOrEqual        = "*lte"
	MetaGreaterThan        = "*gt"
	MetaGreaterOrEqual     = "*gte"
	MetaResources          = "*resources"
	MetaEqual              = "*eq"
	MetaIPNet              = "*ipnet"
	MetaAPIBan             = "*apiban"
	MetaSentryPeer         = "*sentrypeer"
	MetaToken              = "*token"
	MetaIp                 = "*ip"
	MetaNumber             = "*number"
	MetaActivationInterval = "*ai"
	MetaRegex              = "*regex"
	MetaContains           = "*contains"
	MetaHTTP               = "*http"

	MetaNotString             = "*notstring"
	MetaNotPrefix             = "*notprefix"
	MetaNotSuffix             = "*notsuffix"
	MetaNotEmpty              = "*notempty"
	MetaNotExists             = "*notexists"
	MetaNotTimings            = "*nottimings"
	MetaNotRSR                = "*notrsr"
	MetaNotStatS              = "*notstats"
	MetaNotDestinations       = "*notdestinations"
	MetaNotResources          = "*notresources"
	MetaNotEqual              = "*noteq"
	MetaNotIPNet              = "*notipnet"
	MetaNotAPIBan             = "*notapiban"
	MetaNotSentryPeer         = "*notsentrypeer"
	MetaNotActivationInterval = "*notai"
	MetaNotRegex              = "*notregex"
	MetaNotContains           = "*notcontains"

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
	ReplicatorSv1GetRanking              = "ReplicatorSv1.GetRanking"
	ReplicatorSv1GetRankingProfile       = "ReplicatorSv1.GetRankingProfile"
	ReplicatorSv1GetTrend                = "ReplicatorSv1.GetTrend"
	ReplicatorSv1GetTrendProfile         = "ReplicatorSv1.GetTrendProfile"
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
	ReplicatorSv1SetRanking              = "ReplicatorSv1.SetRanking"
	ReplicatorSv1SetRankingProfile       = "ReplicatorSv1.SetRankingProfile"
	ReplicatorSv1SetTrend                = "ReplicatorSv1.SetTrend"
	ReplicatorSv1SetTrendProfile         = "ReplicatorSv1.SetTrendProfile"
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
	ReplicatorSv1SetDispatcherHost       = "ReplicatorSv1.SetDispatcherHost"
	ReplicatorSv1SetLoadIDs              = "ReplicatorSv1.SetLoadIDs"
	ReplicatorSv1SetBackupSessions       = "ReplicatorSv1.SetBackupSessions"
	ReplicatorSv1RemoveSessionBackup     = "ReplicatorSv1.RemoveSessionBackup"
	ReplicatorSv1RemoveThreshold         = "ReplicatorSv1.RemoveThreshold"
	ReplicatorSv1RemoveDestination       = "ReplicatorSv1.RemoveDestination"
	ReplicatorSv1RemoveAccount           = "ReplicatorSv1.RemoveAccount"
	ReplicatorSv1RemoveStatQueue         = "ReplicatorSv1.RemoveStatQueue"
	ReplicatorSv1RemoveFilter            = "ReplicatorSv1.RemoveFilter"
	ReplicatorSv1RemoveThresholdProfile  = "ReplicatorSv1.RemoveThresholdProfile"
	ReplicatorSv1RemoveStatQueueProfile  = "ReplicatorSv1.RemoveStatQueueProfile"
	ReplicatorSv1RemoveRanking           = "ReplicatorSv1.RemoveRanking"
	ReplicatorSv1RemoveRankingProfile    = "ReplicatorSv1.RemoveRankingProfile"
	ReplicatorSv1RemoveTrend             = "ReplicatorSv1.RemoveTrend"
	ReplicatorSv1RemoveTrendProfile      = "ReplicatorSv1.RemoveTrendProfile"
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
	ReplicatorSv1RemoveDispatcherHost    = "ReplicatorSv1.RemoveDispatcherHost"
	ReplicatorSv1GetIndexes              = "ReplicatorSv1.GetIndexes"
	ReplicatorSv1SetIndexes              = "ReplicatorSv1.SetIndexes"
	ReplicatorSv1RemoveIndexes           = "ReplicatorSv1.RemoveIndexes"
)

// APIerSv1 APIs
const (
	ApierV1                                   = "ApierV1"
	ApierV2                                   = "ApierV2"
	APIerSv1                                  = "APIerSv1"
	APIerSv1ComputeFilterIndexes              = "APIerSv1.ComputeFilterIndexes"
	APIerSv1ComputeFilterIndexIDs             = "APIerSv1.ComputeFilterIndexIDs"
	APIerSv1GetAccountActionPlansIndexHealth  = "APIerSv1.GetAccountActionPlansIndexHealth"
	APIerSv1GetReverseDestinationsIndexHealth = "APIerSv1.GetReverseDestinationsIndexHealth"
	APIerSv1GetReverseFilterHealth            = "APIerSv1.GetReverseFilterHealth"
	APIerSv1GetThresholdsIndexesHealth        = "APIerSv1.GetThresholdsIndexesHealth"
	APIerSv1GetResourcesIndexesHealth         = "APIerSv1.GetResourcesIndexesHealth"
	APIerSv1GetStatsIndexesHealth             = "APIerSv1.GetStatsIndexesHealth"
	APIerSv1GetRoutesIndexesHealth            = "APIerSv1.GetRoutesIndexesHealth"
	APIerSv1GetChargersIndexesHealth          = "APIerSv1.GetChargersIndexesHealth"
	APIerSv1GetAttributesIndexesHealth        = "APIerSv1.GetAttributesIndexesHealth"
	APIerSv1GetDispatchersIndexesHealth       = "APIerSv1.GetDispatchersIndexesHealth"
	APIerSv1Ping                              = "APIerSv1.Ping"
	APIerSv1SetDispatcherProfile              = "APIerSv1.SetDispatcherProfile"
	APIerSv1GetDispatcherProfile              = "APIerSv1.GetDispatcherProfile"
	APIerSv1GetDispatcherProfileIDs           = "APIerSv1.GetDispatcherProfileIDs"
	APIerSv1RemoveDispatcherProfile           = "APIerSv1.RemoveDispatcherProfile"
	APIerSv1SetBalances                       = "APIerSv1.SetBalances"
	APIerSv1SetDispatcherHost                 = "APIerSv1.SetDispatcherHost"
	APIerSv1GetDispatcherHost                 = "APIerSv1.GetDispatcherHost"
	APIerSv1GetDispatcherHostIDs              = "APIerSv1.GetDispatcherHostIDs"
	APIerSv1RemoveDispatcherHost              = "APIerSv1.RemoveDispatcherHost"
	APIerSv1GetEventCost                      = "APIerSv1.GetEventCost"
	APIerSv1LoadTariffPlanFromFolder          = "APIerSv1.LoadTariffPlanFromFolder"
	APIerSv1ExportToFolder                    = "APIerSv1.ExportToFolder"
	APIerSv1GetCost                           = "APIerSv1.GetCost"
	APIerSv1SetBalance                        = "APIerSv1.SetBalance"
	APIerSv1TransferBalance                   = "APIerSv1.TransferBalance"
	APIerSv1GetFilter                         = "APIerSv1.GetFilter"
	APIerSv1GetFilterIndexes                  = "APIerSv1.GetFilterIndexes"
	APIerSv1RemoveFilterIndexes               = "APIerSv1.RemoveFilterIndexes"
	APIerSv1RemoveFilter                      = "APIerSv1.RemoveFilter"
	APIerSv1SetFilter                         = "APIerSv1.SetFilter"
	APIerSv1GetFilterIDs                      = "APIerSv1.GetFilterIDs"
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
	APIerSv1RemoveDestination                 = "APIerSv1.RemoveDestination"
	APIerSv1GetReverseDestination             = "APIerSv1.GetReverseDestination"
	APIerSv1AddBalance                        = "APIerSv1.AddBalance"
	APIerSv1DebitBalance                      = "APIerSv1.DebitBalance"
	APIerSv1SetAccount                        = "APIerSv1.SetAccount"
	APIerSv1GetAccountsCount                  = "APIerSv1.GetAccountsCount"
	APIerSv1GetDataDBVersions                 = "APIerSv1.GetDataDBVersions"
	APIerSv1GetStorDBVersions                 = "APIerSv1.GetStorDBVersions"
	APIerSv1GetCDRs                           = "APIerSv1.GetCDRs"
	APIerSv1RemoveCDRs                        = "APIerSv1.RemoveCDRs"
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
	APIerSv1SetTPRanking                      = "APIerSv1.SetTPRanking"
	APIerSv1GetTPRanking                      = "APIerSv1.GetTPRanking"
	APIerSv1RemoveTPRanking                   = "APIerSv1.RemoveTPRanking"
	APIerSv1SetTPTrend                        = "APIerSv1.SetTPTrend"
	APIerSv1GetTPTrend                        = "APIerSv1.GetTPTrend"
	APIerSv1RemoveTPTrend                     = "APIerSv1.RemoveTPTrend"
	APIerSv1GetTPDestinationRate              = "APIerSv1.GetTPDestinationRate"
	APIerSv1SetTPRouteProfile                 = "APIerSv1.SetTPRouteProfile"
	APIerSv1GetTPRouteProfile                 = "APIerSv1.GetTPRouteProfile"
	APIerSv1GetTPRouteProfileIDs              = "APIerSv1.GetTPRouteProfileIDs"
	APIerSv1RemoveTPRouteProfile              = "APIerSv1.RemoveTPRouteProfile"
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
	APIerSv1GetLoadHistory                    = "APIerSv1.GetLoadHistory"
	APIerSv1GetLoadIDs                        = "APIerSv1.GetLoadIDs"
	APIerSv1GetLoadTimes                      = "APIerSv1.GetLoadTimes"
	APIerSv1ExecuteScheduledActions           = "APIerSv1.ExecuteScheduledActions"
	APIerSv1GetSharedGroup                    = "APIerSv1.GetSharedGroup"
	APIerSv1RemoveActionTrigger               = "APIerSv1.RemoveActionTrigger"
	APIerSv1GetAccount                        = "APIerSv1.GetAccount"
	APIerSv1GetAttributeProfileCount          = "APIerSv1.GetAttributeProfileCount"
	APIerSv1GetMaxUsage                       = "APIerSv1.GetMaxUsage"
	APIerSv1GetTiming                         = "APIerSv1.GetTiming"
	APIerSv1SetTiming                         = "APIerSv1.SetTiming"
	APIerSv1RemoveTiming                      = "APIerSv1.RemoveTiming"
	APIerSV1GetAccountCost                    = "APIerSv1.GetAccountCost"
	APIerSV1TimingIsActiveAt                  = "APIerSv1.TimingIsActiveAt"
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

// CoreS APIs
const (
	CoreS                       = "CoreS"
	CoreSv1                     = "CoreSv1"
	CoreSv1Status               = "CoreSv1.Status"
	CoreSv1Ping                 = "CoreSv1.Ping"
	CoreSv1Panic                = "CoreSv1.Panic"
	CoreSv1Sleep                = "CoreSv1.Sleep"
	CoreSv1StartCPUProfiling    = "CoreSv1.StartCPUProfiling"
	CoreSv1StopCPUProfiling     = "CoreSv1.StopCPUProfiling"
	CoreSv1StartMemoryProfiling = "CoreSv1.StartMemoryProfiling"
	CoreSv1StopMemoryProfiling  = "CoreSv1.StopMemoryProfiling"
)

// RouteS APIs
const (
	RouteSv1GetRoutes                = "RouteSv1.GetRoutes"
	RouteSv1GetRoutesList            = "RouteSv1.GetRoutesList"
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
	ThresholdSv1ProcessEvent          = "ThresholdSv1.ProcessEvent"
	ThresholdSv1GetThreshold          = "ThresholdSv1.GetThreshold"
	ThresholdSv1ResetThreshold        = "ThresholdSv1.ResetThreshold"
	ThresholdSv1GetThresholdIDs       = "ThresholdSv1.GetThresholdIDs"
	ThresholdSv1Ping                  = "ThresholdSv1.Ping"
	ThresholdSv1GetThresholdsForEvent = "ThresholdSv1.GetThresholdsForEvent"
	APIerSv1GetThresholdProfileIDs    = "APIerSv1.GetThresholdProfileIDs"
	APIerSv1GetThresholdProfileCount  = "APIerSv1.GetThresholdProfileCount"
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
	StatSv1V1GetQueueIDs           = "StatSv1.GetQueueIDs"
	StatSv1ResetStatQueue          = "StatSv1.ResetStatQueue"
	APIerSv1GetStatQueueProfile    = "APIerSv1.GetStatQueueProfile"
	APIerSv1RemoveStatQueueProfile = "APIerSv1.RemoveStatQueueProfile"
	APIerSv1SetStatQueueProfile    = "APIerSv1.SetStatQueueProfile"
	APIerSv1GetStatQueueProfileIDs = "APIerSv1.GetStatQueueProfileIDs"
)

// TrendS APIs
const (
	APIerSv1SetTrendProfile    = "APIerSv1.SetTrendProfile"
	APIerSv1RemoveTrendProfile = "APIerSv1.RemoveTrendProfile"
	APIerSv1GetTrendProfile    = "APIerSv1.GetTrendProfile"
	APIerSv1GetTrendProfileIDs = "APIerSv1.GetTrendProfileIDs"
	TrendSv1Ping               = "TrendSv1.Ping"
	TrendSv1ScheduleQueries    = "TrendSv1.ScheduleQueries"
	TrendSv1GetTrend           = "TrendSv1.GetTrend"
	TrendSv1GetScheduledTrends = "TrendSv1.GetScheduledTrends"
	TrendSv1GetTrendSummary    = "TrendSv1.GetTrendSummary"
)

// RankingS APIs
const (
	APIerSv1SetRankingProfile    = "APIerSv1.SetRankingProfile"
	APIerSv1RemoveRankingProfile = "APIerSv1.RemoveRankingProfile"
	APIerSv1GetRankingProfile    = "APIerSv1.GetRankingProfile"
	APIerSv1GetRankingProfileIDs = "APIerSv1.GetRankingProfileIDs"
	RankingSv1Ping               = "RankingSv1.Ping"
)

// ResourceS APIs
const (
	ResourceSv1AuthorizeResources    = "ResourceSv1.AuthorizeResources"
	ResourceSv1GetResourcesForEvent  = "ResourceSv1.GetResourcesForEvent"
	ResourceSv1AllocateResources     = "ResourceSv1.AllocateResources"
	ResourceSv1ReleaseResources      = "ResourceSv1.ReleaseResources"
	ResourceSv1Ping                  = "ResourceSv1.Ping"
	ResourceSv1GetResourceWithConfig = "ResourceSv1.GetResourceWithConfig"
	ResourceSv1GetResource           = "ResourceSv1.GetResource"
	APIerSv1SetResourceProfile       = "APIerSv1.SetResourceProfile"
	APIerSv1RemoveResourceProfile    = "APIerSv1.RemoveResourceProfile"
	APIerSv1GetResourceProfile       = "APIerSv1.GetResourceProfile"
	APIerSv1GetResourceProfileIDs    = "APIerSv1.GetResourceProfileIDs"
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
	SessionSv1GetActiveSessions          = "SessionSv1.GetActiveSessions"
	SessionSv1GetActiveSessionsCount     = "SessionSv1.GetActiveSessionsCount"
	SessionSv1ForceDisconnect            = "SessionSv1.ForceDisconnect"
	SessionSv1GetPassiveSessions         = "SessionSv1.GetPassiveSessions"
	SessionSv1GetPassiveSessionsCount    = "SessionSv1.GetPassiveSessionsCount"
	SessionSv1SetPassiveSession          = "SessionSv1.SetPassiveSession"
	SessionSv1Ping                       = "SessionSv1.Ping"
	SessionSv1RegisterInternalBiJSONConn = "SessionSv1.RegisterInternalBiJSONConn"
	SessionSv1ReplicateSessions          = "SessionSv1.ReplicateSessions"
	SessionSv1ActivateSessions           = "SessionSv1.ActivateSessions"
	SessionSv1DeactivateSessions         = "SessionSv1.DeactivateSessions"
	SMGenericV1InitiateSession           = "SMGenericV1.InitiateSession"
	SessionSv1AlterSessions              = "SessionSv1.AlterSessions"
	SessionSv1DisconnectPeer             = "SessionSv1.DisconnectPeer"
	SessionSv1STIRAuthenticate           = "SessionSv1.STIRAuthenticate"
	SessionSv1STIRIdentity               = "SessionSv1.STIRIdentity"
	SessionSv1Sleep                      = "SessionSv1.Sleep"
	SessionSv1CapsError                  = "SessionSv1.CapsError"
	SessionSv1BackupActiveSessions       = "SessionSv1.BackupActiveSessions"
)

// Agent APIs
const (
	AgentV1                    = "AgentV1"
	AgentV1DisconnectSession   = "AgentV1.DisconnectSession"
	AgentV1GetActiveSessionIDs = "AgentV1.GetActiveSessionIDs"
	AgentV1AlterSession        = "AgentV1.AlterSession"
	AgentV1DisconnectPeer      = "AgentV1.DisconnectPeer"
	AgentV1WarnDisconnect      = "AgentV1.WarnDisconnect"
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
	DispatcherSv1                    = "DispatcherSv1"
	DispatcherSv1Ping                = "DispatcherSv1.Ping"
	DispatcherSv1GetProfilesForEvent = "DispatcherSv1.GetProfilesForEvent"
	DispatcherSv1Apier               = "DispatcherSv1.Apier"
	DispatcherServicePing            = "DispatcherService.Ping"
	DispatcherSv1RemoteStatus        = "DispatcherSv1.RemoteStatus"
	DispatcherSv1RemoteSleep         = "DispatcherSv1.RemoteSleep"
	DispatcherSv1RemotePing          = "DispatcherSv1.RemotePing"
)

// RegistrarS APIs
const (
	RegistrarSv1RegisterDispatcherHosts   = "RegistrarSv1.RegisterDispatcherHosts"
	RegistrarSv1UnregisterDispatcherHosts = "RegistrarSv1.UnregisterDispatcherHosts"

	RegistrarSv1RegisterRPCHosts   = "RegistrarSv1.RegisterRPCHosts"
	RegistrarSv1UnregisterRPCHosts = "RegistrarSv1.UnregisterRPCHosts"
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
	CacheSv1GetItem           = "CacheSv1.GetItem"
	CacheSv1GetItemWithRemote = "CacheSv1.GetItemWithRemote"
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
	CDRsV1ReprocessCDRs      = "CDRsV1.ReprocessCDRs"
	CDRsV1GetCDRs            = "CDRsV1.GetCDRs"
	CDRsV1ProcessCDR         = "CDRsV1.ProcessCDR"
	CDRsV1ProcessExternalCDR = "CDRsV1.ProcessExternalCDR"
	CDRsV1StoreSessionCost   = "CDRsV1.StoreSessionCost"
	CDRsV1ProcessEvent       = "CDRsV1.ProcessEvent"
	CDRsV1ProcessEvents      = "CDRsV1.ProcessEvents"
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

// ERs
const (
	ErSv1          = "ErSv1"
	ErSv1Ping      = "ErSv1.Ping"
	ErSv1RunReader = "ErSv1.RunReader"
)

// cgr_ variables
const (
	CGRAccount         = "cgr_account"
	CGRRoute           = "cgr_route"
	CGRDestination     = "cgr_destination"
	CGRSubject         = "cgr_subject"
	CGRCategory        = "cgr_category"
	CGRReqType         = "cgr_reqtype"
	CGRTenant          = "cgr_tenant"
	CGRPdd             = "cgr_pdd"
	CGRDisconnectCause = "cgr_disconnectcause"
	CGRComputeLCR      = "cgr_computelcr"
	CGRRoutes          = "cgr_routes"
	CGRFlags           = "cgr_flags"
	CGROpts            = "cgr_opts"
)

// CSV file name
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
	TrendsCsv             = "Trends.csv"
	RankingsCsv           = "Rankings.csv"
	ThresholdsCsv         = "Thresholds.csv"
	FiltersCsv            = "Filters.csv"
	RoutesCsv             = "Routes.csv"
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
	TBLTPRatingProfiles   = "tp_rating_profiles"
	TBLTPSharedGroups     = "tp_shared_groups"
	TBLTPActions          = "tp_actions"
	TBLTPActionPlans      = "tp_action_plans"
	TBLTPActionTriggers   = "tp_action_triggers"
	TBLTPAccountActions   = "tp_account_actions"
	TBLTPResources        = "tp_resources"
	TBLTPStats            = "tp_stats"
	TBLTPRankings         = "tp_rankings"
	TBLTPTrends           = "tp_trends"
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
	CacheRankingProfiles         = "*ranking_profiles"
	CacheTrendProfiles           = "*trend_profiles"
	CacheTrends                  = "*trends"
	CacheRankings                = "*rankings"
	CacheThresholdProfiles       = "*threshold_profiles"
	CacheThresholds              = "*thresholds"
	CacheFilters                 = "*filters"
	CacheRouteProfiles           = "*route_profiles"
	CacheAttributeProfiles       = "*attribute_profiles"
	CacheChargerProfiles         = "*charger_profiles"
	CacheDispatcherProfiles      = "*dispatcher_profiles"
	CacheDispatcherHosts         = "*dispatcher_hosts"
	CacheDispatchers             = "*dispatchers"
	CacheDispatcherRoutes        = "*dispatcher_routes"
	CacheDispatcherLoads         = "*dispatcher_loads"
	CacheResourceFilterIndexes   = "*resource_filter_indexes"
	CacheStatFilterIndexes       = "*stat_filter_indexes"
	CacheThresholdFilterIndexes  = "*threshold_filter_indexes"
	CacheRankingFilterIndexes    = "ranking_filter_indexes"
	CacheRouteFilterIndexes      = "*route_filter_indexes"
	CacheAttributeFilterIndexes  = "*attribute_filter_indexes"
	CacheChargerFilterIndexes    = "*charger_filter_indexes"
	CacheDispatcherFilterIndexes = "*dispatcher_filter_indexes"
	CacheDiameterMessages        = "*diameter_messages"
	CacheRadiusPackets           = "*radius_packets"
	CacheRPCResponses            = "*rpc_responses"
	CacheClosedSessions          = "*closed_sessions"
	MetaPrecaching               = "*precaching"
	MetaReady                    = "*ready"
	CacheLoadIDs                 = "*load_ids"
	CacheRPCConnections          = "*rpc_connections"
	CacheCDRIDs                  = "*cdr_ids"
	CacheRatingProfilesTmp       = "*tmp_rating_profiles"
	CacheUCH                     = "*uch"
	CacheSTIR                    = "*stir"
	CacheEventCharges            = "*event_charges"
	CacheReverseFilterIndexes    = "*reverse_filter_indexes"
	CacheAccounts                = "*accounts"
	CacheVersions                = "*versions"
	CacheCapsEvents              = "*caps_events"
	CacheSessionsBackup          = "*sessions_backup"
	CacheReplicationHosts        = "*replication_hosts"

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
	CacheTBLTPTrends           = "*tp_trends"
	CacheTBLTPRankings         = "*tp_rankings"
	CacheTBLTPThresholds       = "*tp_thresholds"
	CacheTBLTPFilters          = "*tp_filters"
	CacheSessionCostsTBL       = "*session_costs"
	CacheCDRsTBL               = "*cdrs"
	CacheTBLTPRoutes           = "*tp_routes"
	CacheTBLTPAttributes       = "*tp_attributes"
	CacheTBLTPChargers         = "*tp_chargers"
	CacheTBLTPDispatchers      = "*tp_dispatcher_profiles"
	CacheTBLTPDispatcherHosts  = "*tp_dispatcher_hosts"
)

// Prefix for indexing
const (
	ResourceFilterIndexes   = "rfi_"
	StatFilterIndexes       = "sfi_"
	ThresholdFilterIndexes  = "tfi_"
	AttributeFilterIndexes  = "afi_"
	ChargerFilterIndexes    = "cfi_"
	DispatcherFilterIndexes = "dfi_"
	ActionPlanIndexes       = "api_"
	RouteFilterIndexes      = "rti_"
	FilterIndexPrfx         = "fii_"
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
	JanusAgent      = "JanusAgent"
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
	PgSSLModeDisable    = "disable"
	PgSSLModeAllow      = "allow"
	PgSSLModePrefer     = "prefer"
	PgSSLModeRequire    = "require"
	PgSSLModeVerifyCA   = "verify-ca"
	PgSSLModeVerifyFull = "verify-full"
)

// GeneralCfg
const (
	NodeIDCfg               = "node_id"
	LoggerCfg               = "logger"
	LogLevelCfg             = "log_level"
	RoundingDecimalsCfg     = "rounding_decimals"
	DBDataEncodingCfg       = "dbdata_encoding"
	TpExportPathCfg         = "tpexport_dir"
	PosterAttemptsCfg       = "poster_attempts"
	FailedPostsDirCfg       = "failed_posts_dir"
	FailedPostsTTLCfg       = "failed_posts_ttl"
	DefaultReqTypeCfg       = "default_request_type"
	DefaultCategoryCfg      = "default_category"
	DefaultTenantCfg        = "default_tenant"
	DefaultTimezoneCfg      = "default_timezone"
	DefaultCachingCfg       = "default_caching"
	CachingDlayCfg          = "caching_delay"
	ConnectAttemptsCfg      = "connect_attempts"
	ReconnectsCfg           = "reconnects"
	MaxReconnectIntervalCfg = "max_reconnect_interval"
	ConnectTimeoutCfg       = "connect_timeout"
	ReplyTimeoutCfg         = "reply_timeout"
	LockingTimeoutCfg       = "locking_timeout"
	DigestSeparatorCfg      = "digest_separator"
	DigestEqualCfg          = "digest_equal"
	RSRSepCfg               = "rsr_separator"
	MaxParallelConnsCfg     = "max_parallel_conns"
	EEsConnsCfg             = "ees_conns"
)

// StorDbCfg
const (
	TypeCfg                = "type"
	SQLMaxOpenConnsCfg     = "sqlMaxOpenConns"
	SQLMaxIdleConnsCfg     = "sqlMaxIdleConns"
	SQLConnMaxLifetimeCfg  = "sqlConnMaxLifetime"
	StringIndexedFieldsCfg = "string_indexed_fields"
	PrefixIndexedFieldsCfg = "prefix_indexed_fields"
	SuffixIndexedFieldsCfg = "suffix_indexed_fields"
	MongoQueryTimeoutCfg   = "mongoQueryTimeout"
	MongoConnSchemeCfg     = "mongoConnScheme"
	PgSSLModeCfg           = "pgSSLMode"
	PgSSLCertCfg           = "pgSSLCert"
	PgSSLKeyCfg            = "pgSSLKey"
	PgSSLPasswordCfg       = "pgSSLPassword"
	PgSSLCertModeCfg       = "pgSSLCertMode"
	PgSSLRootCertCfg       = "pgSSLRootCert"
	PgSchema               = "pgSchema"
	ItemsCfg               = "items"
	OptsCfg                = "opts"
	Tenants                = "tenants"
	MysqlLocation          = "mysqlLocation"
)

// DataDbCfg
const (
	DataDbTypeCfg              = "db_type"
	DataDbHostCfg              = "db_host"
	DataDbPortCfg              = "db_port"
	DataDbNameCfg              = "db_name"
	DataDbUserCfg              = "db_user"
	DataDbPassCfg              = "db_password"
	RedisMaxConnsCfg           = "redisMaxConns"
	RedisConnectAttemptsCfg    = "redisConnectAttempts"
	RedisSentinelNameCfg       = "redisSentinel"
	RedisClusterCfg            = "redisCluster"
	RedisClusterSyncCfg        = "redisClusterSync"
	RedisClusterOnDownDelayCfg = "redisClusterOndownDelay"
	RedisPoolPipelineWindowCfg = "redisPoolPipelineWindow"
	RedisPoolPipelineLimitCfg  = "redisPoolPipelineLimit"
	RedisConnectTimeoutCfg     = "redisConnectTimeout"
	RedisReadTimeoutCfg        = "redisReadTimeout"
	RedisWriteTimeoutCfg       = "redisWriteTimeout"
	RedisTLS                   = "redisTLS"
	RedisClientCertificate     = "redisClientCertificate"
	RedisClientKey             = "redisClientKey"
	RedisCACertificate         = "redisCACertificate"
	ReplicationFilteredCfg     = "replication_filtered"
	ReplicationCache           = "replication_cache"
	RemoteConnIDCfg            = "remote_conn_id"
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
	HTTPJsonRPCURLCfg        = "json_rpc_url"
	RegistrarSURLCfg         = "registrars_url"
	PrometheusURLCfg         = "prometheus_url"
	HTTPWSURLCfg             = "ws_url"
	HTTPFreeswitchCDRsURLCfg = "freeswitch_cdrs_url"
	HTTPCDRsURLCfg           = "http_cdrs"
	PprofPathCfg             = "pprof_path"
	HTTPUseBasicAuthCfg      = "use_basic_auth"
	HTTPAuthUsersCfg         = "auth_users"
	HTTPClientOptsCfg        = "client_opts"
	ConfigsURL               = "configs_url"

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
	TrendSConnsCfg    = "trends_conns"
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
	FallbackDepthCfg           = "fallback_depth"
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
	ListenBijsonCfg           = "listen_bijson"
	ListenBigobCfg            = "listen_bigob"
	RALsConnsCfg              = "rals_conns"
	ResSConnsCfg              = "resources_conns"
	ThreshSConnsCfg           = "thresholds_conns"
	RouteSConnsCfg            = "routes_conns"
	AttrSConnsCfg             = "attributes_conns"
	ReplicationConnsCfg       = "replication_conns"
	RemoteConnsCfg            = "remote_conns"
	DebitIntervalCfg          = "debit_interval"
	StoreSCostsCfg            = "store_session_costs"
	SessionTTLCfg             = "session_ttl"
	SessionTTLMaxDelayCfg     = "session_ttl_max_delay"
	SessionTTLLastUsedCfg     = "session_ttl_last_used"
	SessionTTLLastUsageCfg    = "session_ttl_last_usage"
	SessionTTLUsageCfg        = "session_ttl_usage"
	SessionIndexesCfg         = "session_indexes"
	ClientProtocolCfg         = "client_protocol"
	ChannelSyncIntervalCfg    = "channel_sync_interval"
	StaleChanMaxExtraUsageCfg = "stale_chan_max_extra_usage"
	TerminateAttemptsCfg      = "terminate_attempts"
	AlterableFieldsCfg        = "alterable_fields"
	MinDurLowBalanceCfg       = "min_dur_low_balance"
	DefaultUsageCfg           = "default_usage"
	STIRCfg                   = "stir"
	BackupIntervalCfg         = "backup_interval"

	AllowedAtestCfg       = "allowed_attest"
	PayloadMaxdurationCfg = "payload_maxduration"
	DefaultAttestCfg      = "default_attest"
	PublicKeyPathCfg      = "publickey_path"
	PrivateKeyPathCfg     = "privatekey_path"
)

// FsAgentCfg
const (
	SessionSConnsCfg          = "sessions_conns"
	SubscribeParkCfg          = "subscribe_park"
	CreateCdrCfg              = "create_cdr"
	LowBalanceAnnFileCfg      = "low_balance_ann_file"
	EmptyBalanceContextCfg    = "empty_balance_context"
	EmptyBalanceAnnFileCfg    = "empty_balance_ann_file"
	MaxWaitConnectionCfg      = "max_wait_connection"
	EventSocketConnsCfg       = "event_socket_conns"
	EmptyBalanceContext       = "empty_balance_context"
	ActiveSessionDelimiterCfg = "active_session_delimiter"
)

// From Config
const (
	AddressCfg       = "address"
	Password         = "password"
	AliasCfg         = "alias"
	AccountSConnsCfg = "accounts_conns"
	AdminAddressCfg  = "admin_address"
	AdminPasswordCfg = "admin_password"

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
	NetworkCfg           = "network"
	ListenersCfg         = "listeners"
	ListenCfg            = "listen"
	DictionariesPathCfg  = "dictionaries_path"
	OriginHostCfg        = "origin_host"
	OriginRealmCfg       = "origin_realm"
	VendorIDCfg          = "vendor_id"
	ProductNameCfg       = "product_name"
	SyncedConnReqsCfg    = "synced_conn_requests"
	ASRTemplateCfg       = "asr_template"
	RARTemplateCfg       = "rar_template"
	ForcedDisconnectCfg  = "forced_disconnect"
	TemplatesCfg         = "templates"
	RequestProcessorsCfg = "request_processors"

	JanusConnsCfg = "janus_conns"
	// RequestProcessor
	RequestFieldsCfg = "request_fields"
	ReplyFieldsCfg   = "reply_fields"

	// RadiusAgentCfg
	AuthAddrCfg           = "auth_address"
	AcctAddrCfg           = "acct_address"
	ClientSecretsCfg      = "client_secrets"
	ClientDictionariesCfg = "client_dictionaries"
	ClientDaAddressesCfg  = "client_da_addresses"
	RequestsCacheKeyCfg   = "requests_cache_key"
	DMRTemplateCfg        = "dmr_template"
	CoATemplateCfg        = "coa_template"
	HostCfg               = "host"
	PortCfg               = "port"

	// AttributeSCfg
	IndexedSelectsCfg           = "indexed_selects"
	MetaProfileIDs              = "*profileIDs"
	MetaProcessRuns             = "*processRuns"
	MetaProfileRuns             = "*profileRuns"
	MetaProfileIgnoreFiltersCfg = "*profileIgnoreFilters"
	NestedFieldsCfg             = "nested_fields"
	AnyContextCfg               = "any_context"

	// ChargerSCfg
	StoreIntervalCfg = "store_interval"

	// StatSCfg
	StoreUncompressedLimitCfg = "store_uncompressed_limit"
	EEsExporterIDsCfg         = "ees_exporter_ids"

	// Cache
	PartitionsCfg = "partitions"
	PrecacheCfg   = "precache"

	// CdreCfg
	ExportPathCfg         = "export_path"
	AttributeSContextCfg  = "attributes_context"
	SynchronousCfg        = "synchronous"
	AttemptsCfg           = "attempts"
	AttributeContextCfg   = "attribute_context"
	AttributeIDsCfg       = "attribute_ids"
	ConcurrentRequestsCfg = "concurrent_requests"

	//LoaderSCfg
	DryRunCfg       = "dry_run"
	LockFilePathCfg = "lockfile_path"
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

	// ResourceSCfg
	MetaUsageIDCfg  = "*usageID"
	MetaUsageTTLCfg = "*usageTTL"
	MetaUnitsCfg    = "*units"

	// RoutesCfg
	MetaProfileCountCfg = "*profileCount"
	MetaIgnoreErrorsCfg = "*ignoreErrors"
	MetaMaxCostCfg      = "*maxCost"
	MetaLimitCfg        = "*limit"
	MetaOffsetCfg       = "*offset"

	// AnalyzerSCfg
	CleanupIntervalCfg = "cleanup_interval"
	IndexTypeCfg       = "index_type"
	DBPathCfg          = "db_path"

	// CoreSCfg
	CapsCfg              = "caps"
	CapsStrategyCfg      = "caps_strategy"
	CapsStatsIntervalCfg = "caps_stats_interval"
	ShutdownTimeoutCfg   = "shutdown_timeout"

	// AccountSCfg
	MaxIterations = "max_iterations"
	MaxUsage      = "max_usage"

	// DispatcherSCfg
	AnySubsystemCfg = "any_subsystem"
	PreventLoopCfg  = "prevent_loop"
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
	ScheduledIDsCfg    = "scheduled_ids"
)

// MigratorCgrCfg
const (
	OutDataDBTypeCfg       = "out_datadb_type"
	OutDataDBHostCfg       = "out_datadb_host"
	OutDataDBPortCfg       = "out_datadb_port"
	OutDataDBNameCfg       = "out_datadb_name"
	OutDataDBUserCfg       = "out_datadb_user"
	OutDataDBPasswordCfg   = "out_datadb_password"
	OutDataDBEncodingCfg   = "out_datadb_encoding"
	OutDataDBRedisSentinel = "out_redis_sentinel"
	OutStorDBTypeCfg       = "out_stordb_type"
	OutStorDBHostCfg       = "out_stordb_host"
	OutStorDBPortCfg       = "out_stordb_port"
	OutStorDBNameCfg       = "out_stordb_name"
	OutStorDBUserCfg       = "out_stordb_user"
	OutStorDBPasswordCfg   = "out_stordb_password"
	OutStorDBOptsCfg       = "out_stordb_opts"
	OutDataDBOptsCfg       = "out_datadb_opts"
	UsersFiltersCfg        = "users_filters"
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
	IDCfg                  = "id"
	CacheCfg               = "cache"
	ConcurrentEventsCfg    = "concurrent_events"
	FieldSepCfg            = "field_separator"
	RunDelayCfg            = "run_delay"
	SourcePathCfg          = "source_path"
	ProcessedPathCfg       = "processed_path"
	TenantCfg              = "tenant"
	FlagsCfg               = "flags"
	FieldsCfg              = "fields"
	EEsSuccessIDsCfg       = "ees_success_ids"
	EEsFailedIDsCfg        = "ees_failed_ids"
	CacheDumpFieldsCfg     = "cache_dump_fields"
	PartialCommitFieldsCfg = "partial_commit_fields"
	PartialCacheTTLCfg     = "partial_cache_ttl"
)

// RegistrarCCfg
const (
	RPCCfg             = "rpc"
	DispatcherCfg      = "dispatchers"
	RegistrarsConnsCfg = "registrars_conns"
	HostsCfg           = "hosts"
	RefreshIntervalCfg = "refresh_interval"
)

// APIBanCfg
const (
	KeysCfg = "keys"
)

// SentryPeerCfg
const (
	ClientIdCfg      = "client_id"
	ClientSecretCfg  = "client_secret"
	AudienceCfg      = "audience"
	GrantTypeCfg     = "grant_type"
	ContentType      = "Content-Type"
	JsonBody         = "application/json"
	AuthorizationHdr = "Authorization"
	BearerAuth       = "Bearer"
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
var CGROptionsSet = NewStringSet([]string{OptsSessionsTTL,
	OptsSessionsTTLMaxDelay, OptsSessionsTTLLastUsed, OptsSessionsTTLLastUsage, OptsSessionsTTLUsage,
	OptsDebitInterval, OptsStirATest, OptsStirPayloadMaxDuration, OptsStirIdentity,
	OptsStirOriginatorTn, OptsStirOriginatorURI, OptsStirDestinationTn, OptsStirDestinationURI,
	OptsStirPublicKeyPath, OptsStirPrivateKeyPath, OptsAPIKey, OptsRouteID, OptsContext,
	OptsAttributesProcessRuns, OptsAttributesProfileIDs, OptsRoutesLimit, OptsRoutesOffset,
	OptsRoutesIgnoreErrors, OptsRoutesMaxCost, OptsChargeable, RemoteHostOpt, CacheOpt,
	OptsRoutesProfileCount, OptsDispatchersProfilesCount, OptsAttributesProfileRuns,
	OptsAttributesProfileIgnoreFilters, OptsStatsProfileIDs, OptsStatsProfileIgnoreFilters,
	OptsThresholdsProfileIDs, OptsThresholdsProfileIgnoreFilters, OptsResourcesUsageID, OptsResourcesUsageTTL,
	OptsResourcesUnits, OptsAttributeS, OptsThresholdS, OptsChargerS, OptsStatS, OptsRALs, OptsRerate,
	OptsRefund})

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
	OptsSessionsTTL          = "*sessionsTTL"
	OptsSessionsTTLMaxDelay  = "*sessionsTTLMaxDelay"
	OptsSessionsTTLLastUsed  = "*sessionsTTLLastUsed"
	OptsSessionsTTLLastUsage = "*sessionsTTLLastUsage"
	OptsSessionsTTLUsage     = "*sessionsTTLUsage"
	OptsDebitInterval        = "*sessionsDebitInterval"
	OptsChargeable           = "*sessionsChargeable"
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
	OptsAPIKey                   = "*apiKey"
	OptsRouteID                  = "*routeID"
	OptsDispatchersProfilesCount = "*dispatchersProfilesCount"
	// EEs
	OptsEEsVerbose = "*eesVerbose"
	// Resources
	OptsResourcesUsageID  = "*rsUsageID"
	OptsResourcesUsageTTL = "*rsUsageTTL"
	OptsResourcesUnits    = "*rsUnits"
	// Routes
	OptsRoutesProfileCount = "*rouProfileCount"
	OptsRoutesLimit        = "*rouLimit"
	OptsRoutesOffset       = "*rouOffset"
	OptsRoutesIgnoreErrors = "*rouIgnoreErrors"
	OptsRoutesMaxCost      = "*rouMaxCost"
	// Stats
	OptsStatsProfileIDs           = "*stsProfileIDs"
	OptsStatsProfileIgnoreFilters = "*stsProfileIgnoreFilters"
	// Thresholds
	OptsThresholdsProfileIDs           = "*thdProfileIDs"
	OptsThresholdsProfileIgnoreFilters = "*thdProfileIgnoreFilters"
	//CDRs and Sessions
	OptsAttributeS = "*attributeS"
	OptsChargerS   = "*chargerS"
	OptsStatS      = "*statS"
	OptsThresholdS = "*thresholdS"
	OptsRALs       = "*ralS"
	OptsRerate     = "*rerate"
	OptsRefund     = "*refund"
	// Others
	OptsContext                        = "*context"
	MetaSubsys                         = "*subsys"
	MetaMethod                         = "*reqMethod"
	OptsAttributesProfileIDs           = "*attrProfileIDs"
	OptsAttributesProcessRuns          = "*attrProcessRuns"
	OptsAttributesProfileRuns          = "*attrProfileRuns"
	OptsAttributesProfileIgnoreFilters = "*attrProfileIgnoreFilters"
	MetaEventType                      = "*eventType"
	EventType                          = "EventType"
	SchedulerInit                      = "SchedulerInit"

	RemoteHostOpt = "*rmtHost"
	CacheOpt      = "*cache"
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
	DestinationIDs = "DestinationIDs"
	RatingPlanIDs  = "RatingPlanIDs"
	ActionIDs      = "ActionIDs"
	ThresholdIDs   = "ThresholdIDs"
	FilterIDs      = "FilterIDs"
	TimingIDs      = "TimingIDs"
)

// Poster and Event reader constants
const (
	SQSPoster = "SQSPoster"
	S3Poster  = "S3Poster"

	// General constants for posters and readers
	DefaultQueueID = "cgrates_cdrs"

	// sqs and s3
	AWSRegion = "awsRegion"
	AWSKey    = "awsKey"
	AWSSecret = "awsSecret"
	AWSToken  = "awsToken"

	// sqs
	SQSQueueID = "sqsQueueID"

	// s3
	S3Bucket     = "s3BucketID"
	S3FolderPath = "s3FolderPath"

	// sql
	SQLDefaultDBName  = "cgrates"
	SQLDefaultSSLMode = "disable"

	SQLDBNameOpt    = "sqlDBName"
	SQLTableNameOpt = "sqlTableName"

	SQLMaxOpenConns    = "sqlMaxOpenConns"
	SQLConnMaxLifetime = "sqlConnMaxLifetime"
	MYSQLDSNParams     = "mysqlDSNParams"

	// fileCSV
	CSVRowLengthOpt     = "csvRowLength"
	CSVFieldSepOpt      = "csvFieldSeparator"
	CSVLazyQuotes       = "csvLazyQuotes"
	HeaderDefineCharOpt = "csvHeaderDefineChar"

	// fileXML
	XMLRootPathOpt = "xmlRootPath"

	// amqp
	AMQPDefaultConsumerTag = "cgrates"
	DefaultExchangeType    = "direct"

	AMQPQueueID      = "amqpQueueID"
	AMQPConsumerTag  = "amqpConsumerTag"
	AMQPExchange     = "amqpExchange"
	AMQPExchangeType = "amqpExchangeType"
	AMQPRoutingKey   = "amqpRoutingKey"
	AMQPUsername     = "amqpUsername"
	AMQPPassword     = "amqpPassword"

	// kafka
	KafkaDefaultTopic   = "cgrates"
	KafkaDefaultGroupID = "cgrates"
	KafkaDefaultMaxWait = time.Millisecond

	KafkaTopic         = "kafkaTopic"
	KafkaBatchSize     = "kafkaBatchSize"
	KafkaTLS           = "kafkaTLS"
	KafkaCAPath        = "kafkaCAPath"
	KafkaSkipTLSVerify = "kafkaSkipTLSVerify"
	KafkaGroupID       = "kafkaGroupID"
	KafkaMaxWait       = "kafkaMaxWait"

	// partial
	PartialOpt = "*partial"

	PartialOrderFieldOpt       = "partialOrderField"
	PartialCacheActionOpt      = "partialCacheAction"
	PartialPathOpt             = "partialPath"
	PartialCSVFieldSepartorOpt = "partialcsvFieldSeparator"

	// EEs Elasticsearch options
	ElsIndex               = "elsIndex"
	ElsIfPrimaryTerm       = "elsIfPrimaryTerm"
	ElsIfSeqNo             = "elsIfSeqNo"
	ElsOpType              = "elsOpType"
	ElsPipeline            = "elsPipeline"
	ElsRouting             = "elsRouting"
	ElsTimeout             = "elsTimeout"
	ElsVersionLow          = "elsVersion"
	ElsVersionType         = "elsVersionType"
	ElsWaitForActiveShards = "elsWaitForActiveShards"

	//EES ElasticSearch Logger Options
	ElsJson  = "elsJson"
	ElsColor = "elsColor"
	ElsText  = "elsText"

	// nats
	NatsSubject              = "natsSubject"
	NatsQueueID              = "natsQueueID"
	NatsConsumerName         = "natsConsumerName"
	NatsStreamName           = "natsStreamName"
	NatsJWTFile              = "natsJWTFile"
	NatsSeedFile             = "natsSeedFile"
	NatsClientCertificate    = "natsClientCertificate"
	NatsClientKey            = "natsClientKey"
	NatsCertificateAuthority = "natsCertificateAuthority"
	NatsJetStream            = "natsJetStream"
	NatsJetStreamMaxWait     = "natsJetStreamMaxWait"

	// rpc
	RpcCodec        = "rpcCodec"
	ServiceMethod   = "serviceMethod"
	KeyPath         = "keyPath"
	CertPath        = "certPath"
	CaPath          = "caPath"
	Tls             = "tls"
	ConnIDs         = "connIDs"
	RpcConnTimeout  = "rpcConnTimeout"
	RpcReplyTimeout = "rpcReplyTimeout"
	RPCAPIOpts      = "rpcAPIOpts"
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

// CMD constants
const (
	//Common
	VerboseCgr      = "verbose"
	VersionCgr      = "version"
	QuitCgr         = "quit"
	ExitCgr         = "exit"
	ByeCgr          = "bye"
	CloseCgr        = "close"
	CfgPathCgr      = "config_path"
	DataDBTypeCgr   = "datadb_type"
	DataDBHostCgr   = "datadb_host"
	DataDBPortCgr   = "datadb_port"
	DataDBNameCgr   = "datadb_name"
	DataDBUserCgr   = "datadb_user"
	DataDBPasswdCgr = "datadb_passwd"
	//Cgr console
	CgrConsole     = "cgr-console"
	HomeCgr        = "HOME"
	HistoryCgr     = "/.cgr_history"
	RpcEncodingCgr = "rpc_encoding"
	CertPathCgr    = "crt_path"
	KeyPathCgr     = "key_path"
	CAPathCgr      = "ca_path"
	HelpCgr        = "help"
	SepCgr         = " "
	//Cgr engine
	CgrEngine            = "cgr-engine"
	PrintCfgCgr          = "print_config"
	PidCgr               = "pid"
	HttpPprofCgr         = "http_pprof"
	CpuProfDirCgr        = "cpuprof_dir"
	MemProfDirCgr        = "memprof_dir"
	MemProfIntervalCgr   = "memprof_interval"
	MemProfMaxFilesCgr   = "memprof_maxfiles"
	MemProfTimestampCgr  = "memprof_timestamp"
	ScheduledShutdownCgr = "scheduled_shutdown"
	SingleCpuCgr         = "singlecpu"
	PreloadCgr           = "preload"
	SetVersionsCgr       = "set_versions"
	MemProfFinalFile     = "mem_final.prof"
	CpuPathCgr           = "cpu.prof"
	//Cgr loader
	CgrLoader         = "cgr-loader"
	StorDBTypeCgr     = "stordb_type"
	StorDBHostCgr     = "stordb_host"
	StorDBPortCgr     = "stordb_port"
	StorDBNameCgr     = "stordb_name"
	StorDBUserCgr     = "stordb_user"
	StorDBPasswdCgr   = "stordb_passwd"
	CachingArgCgr     = "caching"
	FieldSepCgr       = "field_sep"
	ImportIDCgr       = "import_id"
	DisableReverseCgr = "disable_reverse_mappings"
	FlushStorDB       = "flush_stordb"
	RemoveCgr         = "remove"
	FromStorDBCgr     = "from_stordb"
	ToStorDBcgr       = "to_stordb"
	CacheSAddress     = "caches_address"
	SchedulerAddress  = "scheduler_address"
	//Cgr migrator
	CgrMigrator = "cgr-migrator"
	ExecCgr     = "exec"
)

// SessionS disconnect causes

const (
	ForcedDisconnect = "FORCED_DISCONNECT"
	SessionTimeout   = "SESSION_TIMEOUT"
)

var AnzIndexType = StringSet{ // AnzIndexType are the analyzers possible index types
	MetaScorch:   {},
	MetaBoltdb:   {},
	MetaLeveldb:  {},
	MetaMoss:     {},
	MetaInternal: {},
}

// StringTmplType a string set used, by agentRequest and eventRequest to determine if the returned template type is string
var StringTmplType = StringSet{
	MetaConstant:        struct{}{},
	MetaVariable:        struct{}{},
	MetaComposed:        struct{}{},
	MetaUsageDifference: struct{}{},
	MetaPrefix:          struct{}{},
	MetaSuffix:          struct{}{},
	MetaSIPCID:          struct{}{},
}

// Time duration suffix
const (
	NsSuffix = "ns"
	UsSuffix = "us"
	µSuffix  = "µs"
	MsSuffix = "ms"
	SSuffix  = "s"
	MSuffix  = "m"
	HSuffix  = "h"
)

func buildCacheInstRevPrefixes() {
	CachePrefixToInstance = make(map[string]string)
	for k, v := range CacheInstanceToPrefix {
		CachePrefixToInstance[v] = k
	}
}

func init() {
	buildCacheInstRevPrefixes()
	CachePartitions.Remove(CacheAccounts)
	CachePartitions.Remove(CacheVersions)
}
