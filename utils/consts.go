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

	GitLastLog                  string // If set, it will be processed as part of versioning
	PosterTransportContentTypes = map[string]string{
		MetaHTTPjsonCDR:   ContentJSON,
		MetaHTTPjsonMap:   ContentJSON,
		MetaHTTPjson:      ContentJSON,
		MetaHTTPPost:      ContentForm,
		MetaAMQPjsonCDR:   ContentJSON,
		MetaAMQPjsonMap:   ContentJSON,
		MetaAMQPV1jsonMap: ContentJSON,
		MetaSQSjsonMap:    ContentJSON,
		MetaKafkajsonMap:  ContentJSON,
		MetaS3jsonMap:     ContentJSON,
	}

	extraDBPartition = NewStringSet([]string{CacheDispatchers,
		CacheDispatcherRoutes, CacheDispatcherLoads, CacheDiameterMessages, CacheRPCResponses, CacheClosedSessions,
		CacheCDRIDs, CacheRPCConnections, CacheUCH, CacheSTIR, CacheEventCharges, MetaAPIBan,
		CacheCapsEvents, CacheVersions, CacheReplicationHosts})

	dataDBPartition = NewStringSet([]string{
		CacheResourceProfiles, CacheResources, CacheEventResources, CacheStatQueueProfiles, CacheStatQueues,
		CacheThresholdProfiles, CacheThresholds, CacheFilters, CacheRouteProfiles, CacheAttributeProfiles,
		CacheChargerProfiles, CacheActionProfiles, CacheDispatcherProfiles, CacheDispatcherHosts,
		CacheResourceFilterIndexes, CacheStatFilterIndexes, CacheThresholdFilterIndexes, CacheRouteFilterIndexes,
		CacheAttributeFilterIndexes, CacheChargerFilterIndexes, CacheDispatcherFilterIndexes, CacheLoadIDs,
		CacheRateProfiles, CacheRateProfilesFilterIndexes, CacheRateFilterIndexes,
		CacheActionProfilesFilterIndexes, CacheAccountsFilterIndexes, CacheReverseFilterIndexes,
		CacheAccounts})

	storDBPartition = NewStringSet([]string{
		CacheTBLTPResources, CacheTBLTPStats, CacheTBLTPThresholds, CacheTBLTPFilters, CacheSessionCostsTBL, CacheCDRsTBL,
		CacheTBLTPRoutes, CacheTBLTPAttributes, CacheTBLTPChargers, CacheTBLTPDispatchers,
		CacheTBLTPDispatcherHosts, CacheTBLTPRateProfiles, CacheTBLTPActionProfiles, CacheTBLTPAccounts})

	// CachePartitions enables creation of cache partitions
	CachePartitions = JoinStringSet(extraDBPartition, dataDBPartition, storDBPartition)

	CacheInstanceToPrefix = map[string]string{
		CacheResourceProfiles:            ResourceProfilesPrefix,
		CacheResources:                   ResourcesPrefix,
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
		CacheAccounts:                    AccountPrefix,
		CacheResourceFilterIndexes:       ResourceFilterIndexes,
		CacheStatFilterIndexes:           StatFilterIndexes,
		CacheThresholdFilterIndexes:      ThresholdFilterIndexes,
		CacheRouteFilterIndexes:          RouteFilterIndexes,
		CacheAttributeFilterIndexes:      AttributeFilterIndexes,
		CacheChargerFilterIndexes:        ChargerFilterIndexes,
		CacheDispatcherFilterIndexes:     DispatcherFilterIndexes,
		CacheRateProfilesFilterIndexes:   RateProfilesFilterIndexPrfx,
		CacheActionProfilesFilterIndexes: ActionProfilesFilterIndexPrfx,
		CacheAccountsFilterIndexes:       AccountFilterIndexPrfx,

		CacheLoadIDs:              LoadIDPrefix,
		CacheRateFilterIndexes:    RateFilterIndexPrfx,
		CacheReverseFilterIndexes: FilterIndexPrfx,
		MetaAPIBan:                MetaAPIBan, // special case as it is not in a DB
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
		CacheAccountsFilterIndexes:       AccountPrefix,
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

	CacheStorDBPartitions = map[string]string{
		TBLTPResources:       CacheTBLTPResources,
		TBLTPStats:           CacheTBLTPStats,
		TBLTPThresholds:      CacheTBLTPThresholds,
		TBLTPFilters:         CacheTBLTPFilters,
		SessionCostsTBL:      CacheSessionCostsTBL,
		CDRsTBL:              CacheCDRsTBL,
		TBLTPRoutes:          CacheTBLTPRoutes,
		TBLTPAttributes:      CacheTBLTPAttributes,
		TBLTPChargers:        CacheTBLTPChargers,
		TBLTPDispatchers:     CacheTBLTPDispatchers,
		TBLTPDispatcherHosts: CacheTBLTPDispatcherHosts,
		TBLTPRateProfiles:    CacheTBLTPRateProfiles,
		TBLTPActionProfiles:  CacheTBLTPActionProfiles,
		TBLTPAccounts:        CacheTBLTPAccounts,
	}

	// ProtectedSFlds are the fields that sessions should not alter
	ProtectedSFlds = NewStringSet([]string{CGRID, OriginHost, OriginID, Usage})

	ArgCacheToPrefix = map[string]string{
		ResourceProfileIDs:   ResourceProfilesPrefix,
		ResourceIDs:          ResourcesPrefix,
		StatsQueueIDs:        StatQueuePrefix,
		StatsQueueProfileIDs: StatQueueProfilePrefix,
		ThresholdIDs:         ThresholdPrefix,
		ThresholdProfileIDs:  ThresholdProfilePrefix,
		FilterIDs:            FilterPrefix,
		RouteProfileIDs:      RouteProfilePrefix,
		AttributeProfileIDs:  AttributeProfilePrefix,
		ChargerProfileIDs:    ChargerProfilePrefix,
		DispatcherProfileIDs: DispatcherProfilePrefix,
		DispatcherHostIDs:    DispatcherHostPrefix,
		RateProfileIDs:       RateProfilePrefix,
		ActionProfileIDs:     ActionProfilePrefix,

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
		AccountsFilterIndexIDs:       AccountFilterIndexPrfx,
		FilterIndexIDs:               FilterIndexPrfx,
	}
	CacheInstanceToArg map[string]string
	ArgCacheToInstance = map[string]string{

		ResourceProfileIDs:   CacheResourceProfiles,
		ResourceIDs:          CacheResources,
		StatsQueueIDs:        CacheStatQueues,
		StatsQueueProfileIDs: CacheStatQueueProfiles,
		ThresholdIDs:         CacheThresholds,
		ThresholdProfileIDs:  CacheThresholdProfiles,
		FilterIDs:            CacheFilters,
		RouteProfileIDs:      CacheRouteProfiles,
		AttributeProfileIDs:  CacheAttributeProfiles,
		ChargerProfileIDs:    CacheChargerProfiles,
		DispatcherProfileIDs: CacheDispatcherProfiles,
		DispatcherHostIDs:    CacheDispatcherHosts,
		RateProfileIDs:       CacheRateProfiles,
		ActionProfileIDs:     CacheActionProfiles,

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
		AccountsFilterIndexIDs:       CacheAccountsFilterIndexes,
	}
	ConcurrentReqsLimit    int
	ConcurrentReqsStrategy string
)

const (
	CGRateS                  = "CGRateS"
	CGRateSorg               = "cgrates.org"
	Version                  = "v0.11.0~dev"
	DiameterFirmwareRevision = 918
	RedisMaxConns            = 10
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
	LastUsed                = "LastUsed"
	PDD                     = "PDD"
	Route                   = "Route"
	RunID                   = "RunID"
	AttributeIDs            = "AttributeIDs"
	MetaReqRunID            = "*req.RunID"
	Cost                    = "Cost"
	CostDetails             = "CostDetails"
	Rated                   = "rated"
	Partial                 = "Partial"
	PreRated                = "PreRated"
	StaticValuePrefix       = "^"
	CSV                     = "csv"
	FWV                     = "fwv"
	MetaPartialCSV          = "*partial_csv"
	MetaCombimed            = "*combimed"
	MetaMongo               = "*mongo"
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
	ResourcesPrefix           = "res_"
	ResourceProfilesPrefix    = "rsp_"
	ThresholdPrefix           = "thd_"
	FilterPrefix              = "ftr_"
	CDRsStatsPrefix           = "cst_"
	VersionPrefix             = "ver_"
	StatQueueProfilePrefix    = "sqp_"
	RouteProfilePrefix        = "rpp_"
	RatePrefix                = "rep_"
	AttributeProfilePrefix    = "alp_"
	ChargerProfilePrefix      = "cpp_"
	DispatcherProfilePrefix   = "dpp_"
	RateProfilePrefix         = "rtp_"
	ActionProfilePrefix       = "acp_"
	AccountPrefix             = "acn_"
	DispatcherHostPrefix      = "dph_"
	ThresholdProfilePrefix    = "thp_"
	StatQueuePrefix           = "stq_"
	LoadIDPrefix              = "lid_"
	LoadInstKey               = "load_history"
	CreateCDRsTablesSQL       = "create_cdrs_tables.sql"
	CreateTariffPlanTablesSQL = "create_tariffplan_tables.sql"
	TestSQL                   = "TEST_SQL"
	MetaConstant              = "*constant"
	MetaFiller                = "*filler"
	MetaHTTPPost              = "*http_post"
	MetaHTTPjson              = "*http_json"
	MetaHTTPjsonCDR           = "*http_json_cdr"
	MetaHTTPjsonMap           = "*http_json_map"
	MetaAMQPjsonCDR           = "*amqp_json_cdr"
	MetaAMQPjsonMap           = "*amqp_json_map"
	MetaAMQPV1jsonMap         = "*amqpv1_json_map"
	MetaSQSjsonMap            = "*sqs_json_map"
	MetaKafkajsonMap          = "*kafka_json_map"
	MetaSQL                   = "*sql"
	MetaMySQL                 = "*mysql"
	MetaS3jsonMap             = "*s3_json_map"
	ConfigPath                = "/etc/cgrates/"
	DisconnectCause           = "DisconnectCause"
	MetaFlatstore             = "*flatstore"
	MetaRating                = "*rating"
	NotAvailable              = "N/A"
	Call                      = "call"
	ExtraFields               = "ExtraFields"
	MetaSureTax               = "*sure_tax"
	MetaDynamic               = "*dynamic"
	MetaCounterEvent          = "*event"
	MetaBalance               = "*balance"
	MetaAccount               = "*account"
	EventName                 = "EventName"
	// action trigger threshold types
	TriggerMaxEventCounter = "*max_event_counter"
	TriggerMaxBalance      = "*max_balance"
	HierarchySep           = ">"
	MetaComposed           = "*composed"
	MetaUsageDifference    = "*usage_difference"
	MetaDifference         = "*difference"
	MetaVariable           = "*variable"
	MetaCCUsage            = "*cc_usage"
	MetaValueExponent      = "*value_exponent"
	//rsrparser consts
	NegativePrefix          = "!"
	MatchStartPrefix        = "^"
	MatchGreaterThanOrEqual = ">="
	MatchLessThanOrEqual    = "<="
	MatchGreaterThan        = ">"
	MatchLessThan           = "<"
	MatchEndPrefix          = "$"
	//
	MetaRaw   = "*raw"
	CreatedAt = "CreatedAt"
	UpdatedAt = "UpdatedAt"
	NodeID    = "NodeID"
	//cores consts
	ActiveGoroutines = "ActiveGoroutines"
	MemoryUsage      = "MemoryUsage"
	RunningSince     = "RunningSince"
	GoVersion        = "GoVersion"
	//
	XML                      = "xml"
	MetaGOB                  = "*gob"
	MetaJSON                 = "*json"
	MetaDateTime             = "*datetime"
	MetaMaskedDestination    = "*masked_destination"
	MetaUnixTimestamp        = "*unix_timestamp"
	MetaPostCDR              = "*post_cdr"
	MetaDumpToFile           = "*dump_to_file"
	NonTransactional         = ""
	DataDB                   = "data_db"
	StorDB                   = "stor_db"
	NotFoundCaps             = "NOT_FOUND"
	ServerErrorCaps          = "SERVER_ERROR"
	MandatoryIEMissingCaps   = "MANDATORY_IE_MISSING"
	UnsupportedCachePrefix   = "unsupported cache prefix"
	UnsupportedServiceIDCaps = "UNSUPPORTED_SERVICE_ID"
	ServiceManager           = "service_manager"
	ServiceAlreadyRunning    = "service already running"
	RunningCaps              = "RUNNING"
	StoppedCaps              = "STOPPED"
	MetaAdminS               = "*admins"
	MetaReplicator           = "*replicator"
	MetaRerate               = "*rerate"
	MetaRefund               = "*refund"
	MetaStats                = "*stats"
	MetaCore                 = "*core"
	MetaServiceManager       = "*servicemanager"
	MetaChargers             = "*chargers"
	MetaBlockerError         = "*blocker_error"
	MetaConfig               = "*config"
	MetaDispatchers          = "*dispatchers"
	MetaDispatcherHosts      = "*dispatcher_hosts"
	MetaFilters              = "*filters"
	MetaCDRs                 = "*cdrs"
	MetaDC                   = "*dc"
	MetaCaches               = "*caches"
	MetaUCH                  = "*uch"
	MetaGuardian             = "*guardians"
	MetaEEs                  = "*ees"
	MetaRateS                = "*rates"
	MetaContinue             = "*continue"
	MetaUp                   = "*up"
	Migrator                 = "migrator"
	UnsupportedMigrationTask = "unsupported migration task"
	UndefinedVersion         = "undefined version"
	JSONSuffix               = ".json"
	GOBSuffix                = ".gob"
	XMLSuffix                = ".xml"
	CSVSuffix                = ".csv"
	FWVSuffix                = ".fwv"
	ContentJSON              = "json"
	ContentForm              = "form"
	ContentText              = "text"
	FileLockPrefix           = "file_"
	ActionsPoster            = "act"
	MetaFileCSV              = "*file_csv"
	MetaVirt                 = "*virt"
	MetaElastic              = "*elastic"
	MetaFileFWV              = "*file_fwv"
	MetaFile                 = "*file"
	Accounts                 = "Accounts"
	AccountS                 = "AccountS"
	Actions                  = "Actions"
	BalanceMap               = "BalanceMap"
	UnitCounters             = "UnitCounters"
	UpdateTime               = "UpdateTime"
	Rates                    = "Rates"
	//DestinationRates         = "DestinationRates"
	RatingPlans        = "RatingPlans"
	RatingProfiles     = "RatingProfiles"
	AccountActions     = "AccountActions"
	Resources          = "Resources"
	Stats              = "Stats"
	Filters            = "Filters"
	DispatcherProfiles = "DispatcherProfiles"
	DispatcherHosts    = "DispatcherHosts"
	RateProfiles       = "RateProfiles"
	ActionProfiles     = "ActionProfiles"
	AccountsString     = "Accounts"
	MetaEveryMinute    = "*every_minute"
	MetaHourly         = "*hourly"
	ID                 = "ID"
	Address            = "Address"
	Addresses          = "Addresses"
	Transport          = "Transport"
	TLS                = "TLS"
	Subsystems         = "Subsystems"
	Strategy           = "Strategy"
	StrategyParameters = "StrategyParameters"
	ConnID             = "ConnID"
	ConnFilterIDs      = "ConnFilterIDs"
	ConnWeight         = "ConnWeight"
	ConnBlocker        = "ConnBlocker"
	ConnParameters     = "ConnParameters"

	Thresholds  = "Thresholds"
	Routes      = "Routes"
	Attributes  = "Attributes"
	Chargers    = "Chargers"
	Dispatchers = "Dispatchers"
	StatS       = "Stats"
	LoadIDsVrs  = "LoadIDs"
	GlobalVarS  = "GlobalVarS"
	CostSource  = "CostSource"
	ExtraInfo   = "ExtraInfo"
	Meta        = "*"
	MetaSysLog  = "*syslog"
	MetaStdLog  = "*stdout"
	EventSource = "EventSource"
	AccountID   = "AccountID"
	AccountIDs  = "AccountIDs"
	ResourceID  = "ResourceID"
	TotalUsage  = "TotalUsage"
	StatID      = "StatID"
	BalanceType = "BalanceType"
	BalanceID   = "BalanceID"
	//BalanceDestinationIds = "BalanceDestinationIds"
	BalanceWeight        = "BalanceWeight"
	BalanceRatingSubject = "BalanceRatingSubject"
	BalanceCategories    = "BalanceCategories"
	BalanceBlocker       = "BalanceBlocker"
	BalanceDisabled      = "BalanceDisabled"
	Units                = "Units"
	AccountUpdate        = "AccountUpdate"
	BalanceUpdate        = "BalanceUpdate"
	StatUpdate           = "StatUpdate"
	ResourceUpdate       = "ResourceUpdate"
	CDR                  = "CDR"
	CDRs                 = "CDRs"
	ExpiryTime           = "ExpiryTime"
	AllowNegative        = "AllowNegative"
	Disabled             = "Disabled"
	Action               = "Action"

	SessionSCosts = "SessionSCosts"
	RQF           = "RQF"
	Resource      = "Resource"
	User          = "User"
	Subscribers   = "Subscribers"
	//Destinations             = "Destinations"
	MetaSubscribers          = "*subscribers"
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
	MetaReq                  = "*req"
	MetaVars                 = "*vars"
	MetaRep                  = "*rep"
	MetaExp                  = "*exp"
	MetaHdr                  = "*hdr"
	MetaTrl                  = "*trl"
	MetaTmp                  = "*tmp"
	MetaOpts                 = "*opts"
	MetaCfg                  = "*cfg"
	MetaDynReq               = "~*req"
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
	IdxCombination           = "]["

	MetaRemoteHost        = "*remote_host"
	RemoteHost            = "RemoteHost"
	Local                 = "local"
	TCP                   = "tcp"
	UDP                   = "udp"
	CGRDebitInterval      = "CGRDebitInterval"
	VersionName           = "Version"
	MetaTenant            = "*tenant"
	ResourceUsage         = "ResourceUsage"
	MetaDuration          = "*duration"
	MetaLibPhoneNumber    = "*libphonenumber"
	MetaTimeString        = "*time_string"
	MetaIP2Hex            = "*ip2hex"
	MetaString2Hex        = "*string2hex"
	MetaUnixTime          = "*unixtime"
	MetaSIPURIMethod      = "*sipuri_method"
	MetaSIPURIHost        = "*sipuri_host"
	MetaSIPURIUser        = "*sipuri_user"
	MetaReload            = "*reload"
	MetaLoad              = "*load"
	MetaRemove            = "*remove"
	MetaRemoveAll         = "*removeall"
	MetaStore             = "*store"
	MetaClear             = "*clear"
	MetaExport            = "*export"
	MetaExportID          = "*export_id"
	LoadIDs               = "load_ids"
	DNSAgent              = "DNSAgent"
	TLSNoCaps             = "tls"
	UsageID               = "UsageID"
	Rcode                 = "Rcode"
	Replacement           = "Replacement"
	Regexp                = "Regexp"
	Order                 = "Order"
	Preference            = "Preference"
	Flags                 = "Flags"
	Service               = "Service"
	ApierV                = "ApierV"
	MetaAnalyzer          = "*analyzer"
	CGREventString        = "CGREvent"
	MetaTextPlain         = "*text_plain"
	MetaIgnoreErrors      = "*ignore_errors"
	MetaRelease           = "*release"
	MetaAllocate          = "*allocate"
	MetaAuthorize         = "*authorize"
	MetaSTIRAuthenticate  = "*stir_authenticate"
	MetaSTIRInitiate      = "*stir_initiate"
	MetaInit              = "*init"
	ERs                   = "ERs"
	EEs                   = "EEs"
	Ratio                 = "Ratio"
	Load                  = "Load"
	Slash                 = "/"
	UUID                  = "UUID"
	ActionsID             = "ActionsID"
	MetaAct               = "*act"
	ExportTemplate        = "ExportTemplate"
	Synchronous           = "Synchronous"
	Attempts              = "Attempts"
	FieldSeparator        = "FieldSeparator"
	ExportPath            = "ExportPath"
	ExporterIDs           = "ExporterIDs"
	TimeNow               = "TimeNow"
	ExportFileName        = "ExportFileName"
	GroupID               = "GroupID"
	ThresholdType         = "ThresholdType"
	ThresholdValue        = "ThresholdValue"
	Recurrent             = "Recurrent"
	Executed              = "Executed"
	MinSleep              = "MinSleep"
	ActivationDate        = "ActivationDate"
	ExpirationDate        = "ExpirationDate"
	MinQueuedItems        = "MinQueuedItems"
	OrderIDStart          = "OrderIDStart"
	OrderIDEnd            = "OrderIDEnd"
	MinCost               = "MinCost"
	MaxCost               = "MaxCost"
	MetaLoaders           = "*loaders"
	TmpSuffix             = ".tmp"
	MetaDiamreq           = "*diamreq"
	MetaCost              = "*cost"
	MetaGroup             = "*group"
	InternalRPCSet        = "InternalRPCSet"
	FileName              = "FileName"
	MetaRadauth           = "*radauth"
	UserPassword          = "UserPassword"
	RadauthFailed         = "RADAUTH_FAILED"
	MetaPAP               = "*pap"
	MetaCHAP              = "*chap"
	MetaMSCHAPV2          = "*mschapv2"
	MetaDynaprepaid       = "*dynaprepaid"
	MetaFD                = "*fd"
	SortingData           = "SortingData"
	Count                 = "Count"
	ProfileID             = "ProfileID"
	SortedRoutes          = "SortedRoutes"
	EventExporterS        = "EventExporterS"
	MetaMonthly           = "*monthly"
	MetaYearly            = "*yearly"
	MetaDaily             = "*daily"
	MetaWeekly            = "*weekly"
	RateS                 = "RateS"
	Underline             = "_"
	MetaPartial           = "*partial"
	MetaBusy              = "*busy"
	MetaQueue             = "*queue"
	MetaMonthEnd          = "*month_end"
	APIKey                = "ApiKey"
	RouteID               = "RouteID"
	MetaMonthlyEstimated  = "*monthly_estimated"
	ProcessRuns           = "ProcessRuns"
	HashtagSep            = "#"
	MetaRounding          = "*rounding"
	StatsNA               = -1.0
	InvalidUsage          = -1
	ActionS               = "ActionS"
	Schedule              = "Schedule"
	ActionFilterIDs       = "ActionFilterIDs"
	ActionBlocker         = "ActionBlocker"
	ActionTTL             = "ActionTTL"
	ActionOpts            = "ActionOpts"
	ActionPath            = "ActionPath"
	TPid                  = "TPid"
	LoadId                = "LoadId"
	ActionPlanId          = "ActionPlanId"
	AccountActionsId      = "AccountActionsId"
	Loadid                = "loadid"
	ActionPlan            = "ActionPlan"
	ActionsId             = "ActionsId"
	Prefixes              = "Prefixes"
	RateSlots             = "RateSlots"
	RatingPlanBindings    = "RatingPlanBindings"
	RatingPlanActivations = "RatingPlanActivations"
	RatingProfileID       = "RatingProfileID"
	Time                  = "Time"
	TargetIDs             = "TargetIDs"
	TargetType            = "TargetType"
	MetaRow               = "*row"
	BalanceFilterIDs      = "BalanceFilterIDs"
	BalanceOpts           = "BalanceOpts"
	MetaConcrete          = "*concrete"
	MetaAbstract          = "*abstract"
	MetaTransAbstract     = "*transabstract"
	MetaBalanceLimit      = "*balanceLimit"
	MetaBalanceUnlimited  = "*balanceUnlimited"
	MetaTemplateID        = "*templateID"
	MetaCdrLog            = "*cdrLog"
	MetaCDR               = "*cdr"
	MetaExporterIDs       = "*exporterIDs"
	MetaAsync             = "*async"
	MetaUsage             = "*usage"
	Weights               = "Weights"
	UnitFactors           = "UnitFactors"
	CostIncrements        = "CostIncrements"
	Factor                = "Factor"
	Increment             = "Increment"
	FixedFee              = "FixedFee"
	RecurrentFee          = "RecurrentFee"
	Diktats               = "Diktats"
	BalanceIDs            = "BalanceIDs"
	MetaCostIncrement     = "*costIncrement"
	Length                = "Length"
	MIN_PREFIX_MATCH      = 1
	V1Prfx                = "V1"
)

// Migrator Action
const (
	Move    = "move"
	Migrate = "migrate"
)

// Meta Items
const (
	MetaAccounts           = "*accounts"
	MetaActions            = "*actions"
	MetaResourceProfile    = "*resource_profiles"
	MetaStatQueueProfiles  = "*statqueue_profiles"
	MetaStatQueues         = "*statqueues"
	MetaThresholdProfiles  = "*threshold_profiles"
	MetaRouteProfiles      = "*route_profiles"
	MetaAttributeProfiles  = "*attribute_profiles"
	MetaIndexes            = "*indexes"
	MetaDispatcherProfiles = "*dispatcher_profiles"
	MetaRateProfiles       = "*rate_profiles"
	MetaRateProfileRates   = "*rate_profile_rates"
	MetaChargerProfiles    = "*charger_profiles"
	MetaThresholds         = "*thresholds"
	MetaRoutes             = "*routes"
	MetaAttributes         = "*attributes"
	MetaActionProfiles     = "*action_profiles"
	MetaLoadIDs            = "*load_ids"
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
	RegistrarC  = "RegistrarC"
	LoaderS     = "LoaderS"
	ChargerS    = "ChargerS"
	CacheS      = "CacheS"
	AnalyzerS   = "AnalyzerS"
	CDRServer   = "CDRServer"
	GuardianS   = "GuardianS"
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
	ReplicatorLow  = "replicator"
	EEsLow         = "ees"
	RateSLow       = "rates"
	AccountSLow    = "accounts"
	ActionSLow     = "actions"
)

// Actions
const (
	MetaLog                = "*log"
	MetaTopUpReset         = "*topup_reset"
	MetaTopUp              = "*topup"
	MetaDebitReset         = "*debit_reset"
	MetaDebit              = "*debit"
	MetaEnableAccount      = "*enable_account"
	MetaDisableAccount     = "*disable_account"
	MetaUnlimited          = "*unlimited"
	CDRLog                 = "*cdrlog"
	MetaCgrRpc             = "*cgr_rpc"
	MetaRemoveSessionCosts = "*remove_session_costs"
	MetaPostEvent          = "*post_event"
	MetaCDRAccount         = "*reset_account_cdr"
	MetaResetThreshold     = "*reset_threshold"
	MetaResetStatQueue     = "*reset_stat_queue"
	MetaRemoteSetAccount   = "*remote_set_account"
	ActionID               = "ActionID"
	ActionType             = "ActionType"
	ActionValue            = "ActionValue"
	BalanceValue           = "BalanceValue"
	BalanceUnits           = "BalanceUnits"
	ExtraParameters        = "ExtraParameters"

	MetaAddBalance = "*add_balance"
	MetaSetBalance = "*set_balance"
	MetaRemBalance = "*rem_balance"
)

// Migrator Metas
const (
	MetaSetVersions         = "*set_versions"
	MetaEnsureIndexes       = "*ensure_indexes"
	MetaTpFilters           = "*tp_filters"
	MetaTpThresholds        = "*tp_thresholds"
	MetaTpRoutes            = "*tp_Routes"
	MetaTpStats             = "*tp_stats"
	MetaTpActionProfiles    = "*tp_action_profiles"
	MetaTpRateProfiles      = "*tp_rate_profiles"
	MetaTpResources         = "*tp_resources"
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

const (
	TpFilters            = "TpFilters"
	TpThresholds         = "TpThresholds"
	TpRoutes             = "TpRoutes"
	TpAttributes         = "TpAttributes"
	TpStats              = "TpStats"
	TpResources          = "TpResources"
	TpResource           = "TpResource"
	TpChargers           = "TpChargers"
	TpDispatchers        = "TpDispatchers"
	TpDispatcherProfiles = "TpDispatcherProfiles"
	TpDispatcherHosts    = "TpDispatcherHosts"
	TpRateProfiles       = "TpRateProfiles"
	TpActionProfiles     = "TpActionProfiles"
	TpAccounts           = "TpAccounts"
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
	MetaNot                = "*not"
	MetaString             = "*string"
	MetaPrefix             = "*prefix"
	MetaSuffix             = "*suffix"
	MetaEmpty              = "*empty"
	MetaExists             = "*exists"
	MetaCronExp            = "*cronexp"
	MetaRSR                = "*rsr"
	MetaLessThan           = "*lt"
	MetaLessOrEqual        = "*lte"
	MetaGreaterThan        = "*gt"
	MetaGreaterOrEqual     = "*gte"
	MetaResources          = "*resources"
	MetaEqual              = "*eq"
	MetaIPNet              = "*ipnet"
	MetaAPIBan             = "*apiban"
	MetaActivationInterval = "*ai"

	MetaNotString             = "*notstring"
	MetaNotPrefix             = "*notprefix"
	MetaNotSuffix             = "*notsuffix"
	MetaNotEmpty              = "*notempty"
	MetaNotExists             = "*notexists"
	MetaNotCronExp            = "*notcronexp"
	MetaNotRSR                = "*notrsr"
	MetaNotStatS              = "*notstats"
	MetaNotResources          = "*notresources"
	MetaNotEqual              = "*noteq"
	MetaNotIPNet              = "*notipnet"
	MetaNotAPIBan             = "*notapiban"
	MetaNotActivationInterval = "*notai"

	MetaEC = "*ec"
)

// ReplicatorSv1 APIs
const (
	ReplicatorSv1                     = "ReplicatorSv1"
	ReplicatorSv1Ping                 = "ReplicatorSv1.Ping"
	ReplicatorSv1GetStatQueue         = "ReplicatorSv1.GetStatQueue"
	ReplicatorSv1GetFilter            = "ReplicatorSv1.GetFilter"
	ReplicatorSv1GetThreshold         = "ReplicatorSv1.GetThreshold"
	ReplicatorSv1GetThresholdProfile  = "ReplicatorSv1.GetThresholdProfile"
	ReplicatorSv1GetStatQueueProfile  = "ReplicatorSv1.GetStatQueueProfile"
	ReplicatorSv1GetResource          = "ReplicatorSv1.GetResource"
	ReplicatorSv1GetResourceProfile   = "ReplicatorSv1.GetResourceProfile"
	ReplicatorSv1GetActions           = "ReplicatorSv1.GetActions"
	ReplicatorSv1GetRouteProfile      = "ReplicatorSv1.GetRouteProfile"
	ReplicatorSv1GetAttributeProfile  = "ReplicatorSv1.GetAttributeProfile"
	ReplicatorSv1GetChargerProfile    = "ReplicatorSv1.GetChargerProfile"
	ReplicatorSv1GetDispatcherProfile = "ReplicatorSv1.GetDispatcherProfile"
	ReplicatorSv1GetRateProfile       = "ReplicatorSv1.GetRateProfile"
	ReplicatorSv1GetActionProfile     = "ReplicatorSv1.GetActionProfile"
	ReplicatorSv1GetDispatcherHost    = "ReplicatorSv1.GetDispatcherHost"
	ReplicatorSv1GetAccount           = "ReplicatorSv1.GetAccount"
	ReplicatorSv1GetItemLoadIDs       = "ReplicatorSv1.GetItemLoadIDs"
	ReplicatorSv1SetThresholdProfile  = "ReplicatorSv1.SetThresholdProfile"
	ReplicatorSv1SetThreshold         = "ReplicatorSv1.SetThreshold"
	ReplicatorSv1SetStatQueue         = "ReplicatorSv1.SetStatQueue"
	ReplicatorSv1SetFilter            = "ReplicatorSv1.SetFilter"
	ReplicatorSv1SetStatQueueProfile  = "ReplicatorSv1.SetStatQueueProfile"
	ReplicatorSv1SetResource          = "ReplicatorSv1.SetResource"
	ReplicatorSv1SetResourceProfile   = "ReplicatorSv1.SetResourceProfile"
	ReplicatorSv1SetActions           = "ReplicatorSv1.SetActions"
	ReplicatorSv1SetRouteProfile      = "ReplicatorSv1.SetRouteProfile"
	ReplicatorSv1SetAttributeProfile  = "ReplicatorSv1.SetAttributeProfile"
	ReplicatorSv1SetChargerProfile    = "ReplicatorSv1.SetChargerProfile"
	ReplicatorSv1SetDispatcherProfile = "ReplicatorSv1.SetDispatcherProfile"
	ReplicatorSv1SetRateProfile       = "ReplicatorSv1.SetRateProfile"
	ReplicatorSv1SetActionProfile     = "ReplicatorSv1.SetActionProfile"
	ReplicatorSv1SetAccount           = "ReplicatorSv1.SetAccount"
	ReplicatorSv1SetDispatcherHost    = "ReplicatorSv1.SetDispatcherHost"
	ReplicatorSv1SetLoadIDs           = "ReplicatorSv1.SetLoadIDs"
	ReplicatorSv1RemoveThreshold      = "ReplicatorSv1.RemoveThreshold"

	ReplicatorSv1RemoveStatQueue         = "ReplicatorSv1.RemoveStatQueue"
	ReplicatorSv1RemoveFilter            = "ReplicatorSv1.RemoveFilter"
	ReplicatorSv1RemoveThresholdProfile  = "ReplicatorSv1.RemoveThresholdProfile"
	ReplicatorSv1RemoveStatQueueProfile  = "ReplicatorSv1.RemoveStatQueueProfile"
	ReplicatorSv1RemoveResource          = "ReplicatorSv1.RemoveResource"
	ReplicatorSv1RemoveResourceProfile   = "ReplicatorSv1.RemoveResourceProfile"
	ReplicatorSv1RemoveActions           = "ReplicatorSv1.RemoveActions"
	ReplicatorSv1RemoveRouteProfile      = "ReplicatorSv1.RemoveRouteProfile"
	ReplicatorSv1RemoveAttributeProfile  = "ReplicatorSv1.RemoveAttributeProfile"
	ReplicatorSv1RemoveChargerProfile    = "ReplicatorSv1.RemoveChargerProfile"
	ReplicatorSv1RemoveDispatcherProfile = "ReplicatorSv1.RemoveDispatcherProfile"
	ReplicatorSv1RemoveRateProfile       = "ReplicatorSv1.RemoveRateProfile"
	ReplicatorSv1RemoveActionProfile     = "ReplicatorSv1.RemoveActionProfile"
	ReplicatorSv1RemoveDispatcherHost    = "ReplicatorSv1.RemoveDispatcherHost"
	ReplicatorSv1RemoveAccount           = "ReplicatorSv1.RemoveAccount"
	ReplicatorSv1GetIndexes              = "ReplicatorSv1.GetIndexes"
	ReplicatorSv1SetIndexes              = "ReplicatorSv1.SetIndexes"
	ReplicatorSv1RemoveIndexes           = "ReplicatorSv1.RemoveIndexes"
)

// APIerSv1 APIs
const (
	AdminSv1ComputeFilterIndexes     = "AdminSv1.ComputeFilterIndexes"
	AdminSv1ComputeFilterIndexIDs    = "AdminSv1.ComputeFilterIndexIDs"
	APIerSv1Ping                     = "APIerSv1.Ping"
	APIerSv1SetDispatcherProfile     = "APIerSv1.SetDispatcherProfile"
	APIerSv1GetDispatcherProfile     = "APIerSv1.GetDispatcherProfile"
	APIerSv1GetDispatcherProfileIDs  = "APIerSv1.GetDispatcherProfileIDs"
	APIerSv1RemoveDispatcherProfile  = "APIerSv1.RemoveDispatcherProfile"
	APIerSv1SetBalances              = "APIerSv1.SetBalances"
	APIerSv1SetDispatcherHost        = "APIerSv1.SetDispatcherHost"
	APIerSv1GetDispatcherHost        = "APIerSv1.GetDispatcherHost"
	APIerSv1GetDispatcherHostIDs     = "APIerSv1.GetDispatcherHostIDs"
	APIerSv1RemoveDispatcherHost     = "APIerSv1.RemoveDispatcherHost"
	APIerSv1GetEventCost             = "APIerSv1.GetEventCost"
	APIerSv1LoadTariffPlanFromFolder = "APIerSv1.LoadTariffPlanFromFolder"
	APIerSv1ExportToFolder           = "APIerSv1.ExportToFolder"
	APIerSv1GetCost                  = "APIerSv1.GetCost"
	AdminSv1GetFilter                = "AdminSv1.GetFilter"
	AdminSv1GetFilterIndexes         = "AdminSv1.GetFilterIndexes"
	AdminSv1RemoveFilterIndexes      = "AdminSv1.RemoveFilterIndexes"
	AdminSv1RemoveFilter             = "AdminSv1.RemoveFilter"
	AdminSv1SetFilter                = "AdminSv1.SetFilter"
	AdminSv1GetFilterIDs             = "AdminSv1.GetFilterIDs"
	APIerSv1SetDataDBVersions        = "APIerSv1.SetDataDBVersions"
	APIerSv1SetStorDBVersions        = "APIerSv1.SetStorDBVersions"
	APIerSv1GetActions               = "APIerSv1.GetActions"

	APIerSv1GetDataDBVersions        = "APIerSv1.GetDataDBVersions"
	APIerSv1GetStorDBVersions        = "APIerSv1.GetStorDBVersions"
	APIerSv1GetCDRs                  = "APIerSv1.GetCDRs"
	APIerSv1GetTPActions             = "APIerSv1.GetTPActions"
	APIerSv1GetTPAttributeProfile    = "APIerSv1.GetTPAttributeProfile"
	APIerSv1SetTPAttributeProfile    = "APIerSv1.SetTPAttributeProfile"
	APIerSv1GetTPAttributeProfileIds = "APIerSv1.GetTPAttributeProfileIds"
	APIerSv1RemoveTPAttributeProfile = "APIerSv1.RemoveTPAttributeProfile"
	APIerSv1GetTPCharger             = "APIerSv1.GetTPCharger"
	APIerSv1SetTPCharger             = "APIerSv1.SetTPCharger"
	APIerSv1RemoveTPCharger          = "APIerSv1.RemoveTPCharger"
	APIerSv1GetTPChargerIDs          = "APIerSv1.GetTPChargerIDs"
	APIerSv1SetTPFilterProfile       = "APIerSv1.SetTPFilterProfile"
	APIerSv1GetTPFilterProfile       = "APIerSv1.GetTPFilterProfile"
	APIerSv1GetTPFilterProfileIds    = "APIerSv1.GetTPFilterProfileIds"
	APIerSv1RemoveTPFilterProfile    = "APIerSv1.RemoveTPFilterProfile"

	APIerSv1GetTPResource             = "APIerSv1.GetTPResource"
	APIerSv1SetTPResource             = "APIerSv1.SetTPResource"
	APIerSv1RemoveTPResource          = "APIerSv1.RemoveTPResource"
	APIerSv1SetTPRate                 = "APIerSv1.SetTPRate"
	APIerSv1GetTPRate                 = "APIerSv1.GetTPRate"
	APIerSv1RemoveTPRate              = "APIerSv1.RemoveTPRate"
	APIerSv1GetTPRateIds              = "APIerSv1.GetTPRateIds"
	APIerSv1SetTPThreshold            = "APIerSv1.SetTPThreshold"
	APIerSv1GetTPThreshold            = "APIerSv1.GetTPThreshold"
	APIerSv1GetTPThresholdIDs         = "APIerSv1.GetTPThresholdIDs"
	APIerSv1RemoveTPThreshold         = "APIerSv1.RemoveTPThreshold"
	APIerSv1SetTPStat                 = "APIerSv1.SetTPStat"
	APIerSv1GetTPStat                 = "APIerSv1.GetTPStat"
	APIerSv1RemoveTPStat              = "APIerSv1.RemoveTPStat"
	APIerSv1SetTPRouteProfile         = "APIerSv1.SetTPRouteProfile"
	APIerSv1GetTPRouteProfile         = "APIerSv1.GetTPRouteProfile"
	APIerSv1GetTPRouteProfileIDs      = "APIerSv1.GetTPRouteProfileIDs"
	APIerSv1RemoveTPRouteProfile      = "APIerSv1.RemoveTPRouteProfile"
	APIerSv1GetTPDispatcherProfile    = "APIerSv1.GetTPDispatcherProfile"
	APIerSv1SetTPDispatcherProfile    = "APIerSv1.SetTPDispatcherProfile"
	APIerSv1RemoveTPDispatcherProfile = "APIerSv1.RemoveTPDispatcherProfile"
	APIerSv1GetTPDispatcherProfileIDs = "APIerSv1.GetTPDispatcherProfileIDs"
	APIerSv1ExportCDRs                = "APIerSv1.ExportCDRs"
	APIerSv1SetTPRatingPlan           = "APIerSv1.SetTPRatingPlan"
	APIerSv1SetTPActions              = "APIerSv1.SetTPActions"
	APIerSv1GetTPActionIds            = "APIerSv1.GetTPActionIds"
	APIerSv1RemoveTPActions           = "APIerSv1.RemoveTPActions"
	APIerSv1SetActionPlan             = "APIerSv1.SetActionPlan"
	APIerSv1ExecuteAction             = "APIerSv1.ExecuteAction"
	APIerSv1SetTPRatingProfile        = "APIerSv1.SetTPRatingProfile"
	APIerSv1GetTPRatingProfile        = "APIerSv1.GetTPRatingProfile"

	APIerSv1ImportTariffPlanFromFolder = "APIerSv1.ImportTariffPlanFromFolder"
	APIerSv1ExportTPToFolder           = "APIerSv1.ExportTPToFolder"
	APIerSv1SetActions                 = "APIerSv1.SetActions"

	APIerSv1GetDataCost                 = "APIerSv1.GetDataCost"
	APIerSv1ReplayFailedPosts           = "APIerSv1.ReplayFailedPosts"
	APIerSv1GetCacheStats               = "APIerSv1.GetCacheStats"
	APIerSv1ReloadCache                 = "APIerSv1.ReloadCache"
	APIerSv1RemoveActions               = "APIerSv1.RemoveActions"
	APIerSv1GetLoadHistory              = "APIerSv1.GetLoadHistory"
	APIerSv1GetLoadIDs                  = "APIerSv1.GetLoadIDs"
	APIerSv1GetLoadTimes                = "APIerSv1.GetLoadTimes"
	AdminSv1GetAttributeProfileIDsCount = "AdminSv1.GetAttributeProfileIDsCount"
	APIerSv1GetTPActionProfile          = "APIerSv1.GetTPActionProfile"
	APIerSv1SetTPActionProfile          = "APIerSv1.SetTPActionProfile"
	APIerSv1GetTPActionProfileIDs       = "APIerSv1.GetTPActionProfileIDs"
	APIerSv1RemoveTPActionProfile       = "APIerSv1.RemoveTPActionProfile"
	APIerSv1GetTPRateProfile            = "APIerSv1.GetTPRateProfile"
	APIerSv1SetTPRateProfile            = "APIerSv1.SetTPRateProfile"
	APIerSv1GetTPRateProfileIds         = "APIerSv1.GetTPRateProfileIds"
	APIerSv1RemoveTPRateProfile         = "APIerSv1.RemoveTPRateProfile"
	APIerSv1SetAccount                  = "APIerSv1.SetAccount"
	APIerSv1GetAccount                  = "APIerSv1.GetAccount"
	APIerSv1GetAccountIDs               = "APIerSv1.GetAccountIDs"
	APIerSv1RemoveAccount               = "APIerSv1.RemoveAccount"
	APIerSv1GetAccountIDsCount          = "APIerSv1.GetAccountIDsCount"
	APIerSv1GetTPAccountIDs             = "APIerSv1.GetTPAccountIDs"
	APIerSv1GetTPAccount                = "APIerSv1.GetTPAccount"
	APIerSv1SetTPAccount                = "APIerSv1.SetTPAccount"
	APIerSv1RemoveTPAccount             = "APIerSv1.RemoveTPAccount"
)

// APIerSv1 TP APIs
const (
	APIerSv1LoadTariffPlanFromStorDb = "APIerSv1.LoadTariffPlanFromStorDb"
	APIerSv1RemoveTPFromFolder       = "APIerSv1.RemoveTPFromFolder"
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
	RateSv1             = "RateSv1"
	RateSv1CostForEvent = "RateSv1.CostForEvent"
	RateSv1Ping         = "RateSv1.Ping"
)

const (
	AccountSv1                    = "AccountSv1"
	AccountSv1Ping                = "AccountSv1.Ping"
	AccountSv1AccountsForEvent    = "AccountSv1.AccountsForEvent"
	AccountSv1MaxAbstracts        = "AccountSv1.MaxAbstracts"
	AccountSv1DebitAbstracts      = "AccountSv1.DebitAbstracts"
	AccountSv1MaxConcretes        = "AccountSv1.MaxConcretes"
	AccountSv1DebitConcretes      = "AccountSv1.DebitConcretes"
	AccountSv1ActionSetBalance    = "AccountSv1.ActionSetBalance"
	AccountSv1ActionRemoveBalance = "AccountSv1.ActionRemoveBalance"
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
	AdminSv1SetAttributeProfile      = "AdminSv1.SetAttributeProfile"
	AdminSv1GetAttributeProfile      = "AdminSv1.GetAttributeProfile"
	AdminSv1GetAttributeProfileIDs   = "AdminSv1.GetAttributeProfileIDs"
	AdminSv1RemoveAttributeProfile   = "AdminSv1.RemoveAttributeProfile"
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
	StatSv1ResetStatQueue          = "StatSv1.ResetStatQueue"
	APIerSv1GetStatQueueProfile    = "APIerSv1.GetStatQueueProfile"
	APIerSv1RemoveStatQueueProfile = "APIerSv1.RemoveStatQueueProfile"
	APIerSv1SetStatQueueProfile    = "APIerSv1.SetStatQueueProfile"
	APIerSv1GetStatQueueProfileIDs = "APIerSv1.GetStatQueueProfileIDs"
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
	SessionSv1ReAuthorize                = "SessionSv1.ReAuthorize"
	SessionSv1DisconnectPeer             = "SessionSv1.DisconnectPeer"
	SessionSv1WarnDisconnect             = "SessionSv1.WarnDisconnect"
	SessionSv1STIRAuthenticate           = "SessionSv1.STIRAuthenticate"
	SessionSv1STIRIdentity               = "SessionSv1.STIRIdentity"
	SessionSv1Sleep                      = "SessionSv1.Sleep"
)

// DispatcherS APIs
const (
	DispatcherSv1                   = "DispatcherSv1"
	DispatcherSv1Ping               = "DispatcherSv1.Ping"
	DispatcherSv1GetProfileForEvent = "DispatcherSv1.GetProfileForEvent"
	DispatcherServicePing           = "DispatcherService.Ping"
)

// RegistrarS APIs
const (
	RegistrarSv1RegisterDispatcherHosts   = "RegistrarSv1.RegisterDispatcherHosts"
	RegistrarSv1UnregisterDispatcherHosts = "RegistrarSv1.UnregisterDispatcherHosts"

	RegistrarSv1RegisterRPCHosts   = "RegistrarSv1.RegisterRPCHosts"
	RegistrarSv1UnregisterRPCHosts = "RegistrarSv1.UnregisterRPCHosts"
)

// RateProfile APIs
const (
	AdminSv1SetRateProfile         = "AdminSv1.SetRateProfile"
	AdminSv1GetRateProfile         = "AdminSv1.GetRateProfile"
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

// AdminSv1
const (
	AdminS   = "AdminS"
	AdminSv1 = "AdminSv1"
)

//cgr_ variables
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

//CSV file name
const (
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
	AccountsCsv           = "Accounts.csv"
)

// Table Name
const (
	TBLTPResources       = "tp_resources"
	TBLTPStats           = "tp_stats"
	TBLTPThresholds      = "tp_thresholds"
	TBLTPFilters         = "tp_filters"
	SessionCostsTBL      = "session_costs"
	CDRsTBL              = "cdrs"
	TBLTPRoutes          = "tp_routes"
	TBLTPAttributes      = "tp_attributes"
	TBLTPChargers        = "tp_chargers"
	TBLVersions          = "versions"
	OldSMCosts           = "sm_costs"
	TBLTPDispatchers     = "tp_dispatcher_profiles"
	TBLTPDispatcherHosts = "tp_dispatcher_hosts"
	TBLTPRateProfiles    = "tp_rate_profiles"
	TBLTPActionProfiles  = "tp_action_profiles"
	TBLTPAccounts        = "tp_accounts"
)

// Cache Name
const (
	CacheResources                   = "*resources"
	CacheResourceProfiles            = "*resource_profiles"
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
	CacheAccounts                    = "*accounts"
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
	CacheAccountsFilterIndexes       = "*account_filter_indexes"
	CacheRateFilterIndexes           = "*rate_filter_indexes"
	MetaPrecaching                   = "*precaching"
	MetaReady                        = "*ready"
	CacheLoadIDs                     = "*load_ids"
	CacheRPCConnections              = "*rpc_connections"
	CacheCDRIDs                      = "*cdr_ids"
	CacheUCH                         = "*uch"
	CacheSTIR                        = "*stir"
	CacheEventCharges                = "*event_charges"
	CacheReverseFilterIndexes        = "*reverse_filter_indexes"
	CacheVersions                    = "*versions"
	CacheCapsEvents                  = "*caps_events"
	CacheReplicationHosts            = "*replication_hosts"

	// storDB

	CacheTBLTPResources       = "*tp_resources"
	CacheTBLTPStats           = "*tp_stats"
	CacheTBLTPThresholds      = "*tp_thresholds"
	CacheTBLTPFilters         = "*tp_filters"
	CacheSessionCostsTBL      = "*session_costs"
	CacheCDRsTBL              = "*cdrs"
	CacheTBLTPRoutes          = "*tp_routes"
	CacheTBLTPAttributes      = "*tp_attributes"
	CacheTBLTPChargers        = "*tp_chargers"
	CacheTBLTPDispatchers     = "*tp_dispatcher_profiles"
	CacheTBLTPDispatcherHosts = "*tp_dispatcher_hosts"
	CacheTBLTPRateProfiles    = "*tp_rate_profiles"
	CacheTBLTPActionProfiles  = "*tp_action_profiles"
	CacheTBLTPAccounts        = "*tp_accounts"
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
	AccountFilterIndexPrfx        = "ani_"
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
	GoogleCredentialsFileName = "credentials.json"
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
	SQLMaxOpenConnsCfg     = "sqlMaxOpenConns"
	SQLMaxIdleConnsCfg     = "sqlMaxIdleConns"
	SQLConnMaxLifetimeCfg  = "sqlConnMaxLifetime"
	StringIndexedFieldsCfg = "string_indexed_fields"
	PrefixIndexedFieldsCfg = "prefix_indexed_fields"
	SuffixIndexedFieldsCfg = "suffix_indexed_fields"
	MongoQueryTimeoutCfg   = "mongoQueryTimeout"
	SSLModeCfg             = "sslMode"
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
	RedisSentinelNameCfg       = "redisSentinel"
	RedisClusterCfg            = "redisCluster"
	RedisClusterSyncCfg        = "redisClusterSync"
	RedisClusterOnDownDelayCfg = "redisClusterOndownDelay"
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
	HTTPWSURLCfg             = "ws_url"
	HTTPFreeswitchCDRsURLCfg = "freeswitch_cdrs_url"
	HTTPCDRsURLCfg           = "http_cdrs"
	HTTPUseBasicAuthCfg      = "use_basic_auth"
	HTTPAuthUsersCfg         = "auth_users"
	HTTPClientOptsCfg        = "client_opts"

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
	AdminSConnsCfg    = "admins_conns"
)

const (
	EnabledCfg         = "enabled"
	ThresholdSConnsCfg = "thresholds_conns"
	CacheSConnsCfg     = "caches_conns"
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
	ListenBigobCfg         = "listen_bigob"
	RouteSConnsCfg         = "routes_conns"
	ReplicationConnsCfg    = "replication_conns"
	RemoteConnsCfg         = "remote_conns"
	DebitIntervalCfg       = "debit_interval"
	StoreSCostsCfg         = "store_session_costs"
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
	DefaultUsageCfg        = "default_usage"
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
	AddressCfg       = "address"
	Password         = "password"
	AliasCfg         = "alias"
	AccountSConnsCfg = "accounts_conns"

	// KamAgentCfg
	EvapiConnsCfg = "evapi_conns"
	TimezoneCfg   = "timezone"

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
	ExportPathCfg       = "export_path"
	SynchronousCfg      = "synchronous"
	AttemptsCfg         = "attempts"
	AttributeContextCfg = "attribute_context"
	AttributeIDsCfg     = "attribute_ids"

	//LoaderSCfg
	DryRunCfg       = "dry_run"
	LockFileNameCfg = "lock_filename"
	TpInDirCfg      = "tp_in_dir"
	TpOutDirCfg     = "tp_out_dir"
	DataCfg         = "data"

	DefaultRatioCfg   = "default_ratio"
	ReadersCfg        = "readers"
	ExportersCfg      = "exporters"
	PoolSize          = "poolSize"
	Conns             = "conns"
	FilenameCfg       = "file_name"
	RequestPayloadCfg = "request_payload"
	ReplyPayloadCfg   = "reply_payload"
	TransportCfg      = "transport"
	StrategyCfg       = "strategy"

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

	// AccountSCfg
	MaxIterations = "max_iterations"
	MaxUsage      = "max_usage"
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
	ActionSConnsCfg    = "actions_conns"
	GapiCredentialsCfg = "gapi_credentials"
	GapiTokenCfg       = "gapi_token"
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
	IDCfg              = "id"
	CacheCfg           = "cache"
	FieldSepCfg        = "field_separator"
	RunDelayCfg        = "run_delay"
	SourcePathCfg      = "source_path"
	ProcessedPathCfg   = "processed_path"
	TenantCfg          = "tenant"
	FlagsCfg           = "flags"
	FieldsCfg          = "fields"
	CacheDumpFieldsCfg = "cache_dump_fields"
)

// RegistrarCCfg
const (
	RPCCfg             = "rpc"
	DispatcherCfg      = "dispatcher"
	RegistrarsConnsCfg = "registrars_conns"
	HostsCfg           = "hosts"
	RefreshIntervalCfg = "refresh_interval"
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
var CGROptionsSet = NewStringSet([]string{OptsRatesStartTime, OptsRatesUsage, OptsSessionsTTL,
	OptsSessionsTTLMaxDelay, OptsSessionsTTLLastUsed, OptsSessionsTTLLastUsage, OptsSessionsTTLUsage,
	OptsDebitInterval, OptsStirATest, OptsStirPayloadMaxDuration, OptsStirIdentity,
	OptsStirOriginatorTn, OptsStirOriginatorURI, OptsStirDestinationTn, OptsStirDestinationURI,
	OptsStirPublicKeyPath, OptsStirPrivateKeyPath, OptsAPIKey, OptsRouteID, OptsContext,
	OptsAttributesProcessRuns, OptsRoutesLimit, OptsRoutesOffset, OptsChargeable,
	RemoteHostOpt, CacheOpt})

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
	OptsRouteProfilesCount   = "*routeProfilesCount"
	OptsRoutesLimit          = "*routes_limit"
	OptsRoutesOffset         = "*routes_offset"
	OptsRatesStartTime       = "*ratesStartTime"
	OptsRatesUsage           = "*ratesUsage"
	OptsRatesIntervalStart   = "*ratesIntervalStart"
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
	OptsAPIKey  = "*apiKey"
	OptsRouteID = "*routeID"
	// EEs
	OptsEEsVerbose = "*eesVerbose"
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
	// SQLEe options
	SQLMaxOpenConns    = "sqlMaxOpenConns"
	SQLMaxConnLifetime = "sqlMaxConnLifetime"

	// Others
	OptsContext               = "*context"
	Subsys                    = "*subsys"
	OptsAttributesProcessRuns = "*processRuns"
	MetaEventType             = "*eventType"
	EventType                 = "EventType"
	SchedulerInit             = "SchedulerInit"

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
	AccountsFilterIndexIDs       = "AccountsFilterIndexIDs"
	FilterIndexIDs               = "FilterIndexIDs"
)

// Poster and Event reader constants
const (
	SQSPoster = "SQSPoster"
	S3Poster  = "S3Poster"

	// General constants for posters and readers
	DefaultQueueID = "cgrates_cdrs"
	ProcessedOpt   = "Processed"

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

	// fileCSV
	RowLengthOpt        = "RowLength"
	FieldSepOpt         = "FieldSeparator"
	LazyQuotes          = "LazyQuotes"
	HeaderDefineCharOpt = "csvHeaderDefineChar"

	// partialCSV
	PartialCSVCacheExpiryActionOpt = "csvCacheExpiryAction"
	PartialCSVRecordCacheOpt       = "csvRecordCacheTTL"

	// flatStore
	FlatstorePrfx = "fst"
	OptsMethod    = "*method"
	FstInvite     = "INVITE"
	FstBye        = "BYE"
	FstAck        = "ACK"

	MetaInvite = "*invite"
	MetaBye    = "*bye"
	MetaAck    = "*ack"

	FstFailedCallsPrefixOpt  = "fstFailedCallsPrefix"
	FstPartialRecordCacheOpt = "fstRecordCacheTTL"
	FstMethodOpt             = "fstMethod"
	FstOriginIDOpt           = "fstOriginID"
	FstMadatoryACKOpt        = "fstMadatoryACK"

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

	// kafka
	KafkaDefaultTopic   = "cgrates"
	KafkaDefaultGroupID = "cgrates"
	KafkaDefaultMaxWait = time.Millisecond

	KafkaTopic   = "kafkaTopic"
	KafkaGroupID = "kafkaGroupID"
	KafkaMaxWait = "kafkaMaxWait"
)

var (
	// FstMethodToPrfx used for flatstore to convert the method in DP prefix
	FstMethodToPrfx = map[string]string{
		FstInvite: MetaInvite,
		FstBye:    MetaBye,
		FstAck:    MetaAck,
	}
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

//CMD constants
const (
	//Common
	VerboseCgr      = "verbose"
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
	CheckCfgCgr          = "check_config"
	PidCgr               = "pid"
	HttpPrfPthCgr        = "httprof_path"
	CpuProfDirCgr        = "cpuprof_dir"
	MemProfDirCgr        = "memprof_dir"
	MemProfIntervalCgr   = "memprof_interval"
	MemProfNrFilesCgr    = "memprof_nrfiles"
	ScheduledShutdownCgr = "scheduled_shutdown"
	SingleCpuCgr         = "singlecpu"
	PreloadCgr           = "preload"
	MemProfFileCgr       = "mem_final.prof"
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

var (
	// AnzIndexType are the analyzers possible index types
	AnzIndexType = StringSet{
		MetaScorch:  {},
		MetaBoltdb:  {},
		MetaLeveldb: {},
		MetaMoss:    {},
	}
)

// ActionSv1
const (
	ActionSv1                = "ActionSv1"
	ActionSv1Ping            = "ActionSv1.Ping"
	ActionSv1ScheduleActions = "ActionSv1.ScheduleActions"
	ActionSv1ExecuteActions  = "ActionSv1.ExecuteActions"
)

// StringTmplType a string set used, by agentRequest and eventRequest to determine if the returned template type is string
var StringTmplType = StringSet{
	MetaConstant:        struct{}{},
	MetaRemoteHost:      struct{}{},
	MetaVariable:        struct{}{},
	MetaComposed:        struct{}{},
	MetaUsageDifference: struct{}{},
	MetaPrefix:          struct{}{},
	MetaSuffix:          struct{}{},
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
	CacheInstanceToArg = make(map[string]string)
	for k, v := range ArgCacheToInstance {
		CacheInstanceToArg[v] = k
	}
}

func init() {
	buildCacheInstRevPrefixes()
}
