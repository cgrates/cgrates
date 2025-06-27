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
	MainCDRFields = NewStringSet([]string{Source, OriginHost, OriginID, ToR, RequestType, Tenant, Category,
		AccountField, Subject, Destination, SetupTime, AnswerTime, Usage, Cost, Rated, Partial, RunID,
		PreRated, CostSource, CostDetails, ExtraInfo, OrderID})
	PostPaidRatedSlice = []string{MetaPostpaid, MetaRated}

	GitCommitDate string // If set, it will be processed as part of versioning
	GitCommitHash string // If set, it will be processed as part of versioning

	extraDBPartition = NewStringSet([]string{
		CacheDiameterMessages, CacheRPCResponses, CacheClosedSessions,
		CacheCDRIDs, CacheRPCConnections, CacheUCH, CacheSTIR, CacheEventCharges, MetaAPIBan, MetaSentryPeer,
		CacheCapsEvents, CacheReplicationHosts})

	// DataDBPartitions excluding Resources, Thresholds, Trends, Rankings, IPs, Stats
	StatelessDataDBPartitions = NewStringSet([]string{
		CacheFilters, CacheRouteProfiles, CacheAttributeProfiles,
		CacheChargerProfiles, CacheActionProfiles, CacheRouteFilterIndexes,
		CacheAttributeFilterIndexes, CacheChargerFilterIndexes, CacheLoadIDs,
		CacheRateProfiles, CacheRateProfilesFilterIndexes,
		CacheRateFilterIndexes, CacheActionProfilesFilterIndexes,
		CacheAccountsFilterIndexes, CacheReverseFilterIndexes, CacheAccounts,
	})

	DataDBPartitions = NewStringSet([]string{
		CacheResourceProfiles, CacheResources, CacheEventResources, CacheIPProfiles, CacheIPAllocations,
		CacheEventIPs, CacheStatQueueProfiles, CacheStatQueues, CacheThresholdProfiles,
		CacheThresholds, CacheFilters, CacheRouteProfiles, CacheAttributeProfiles,
		CacheTrendProfiles, CacheChargerProfiles, CacheActionProfiles, CacheRankingProfiles,
		CacheRankings, CacheTrends, CacheResourceFilterIndexes, CacheIPFilterIndexes, CacheStatFilterIndexes,
		CacheThresholdFilterIndexes, CacheRouteFilterIndexes, CacheAttributeFilterIndexes,
		CacheChargerFilterIndexes, CacheLoadIDs, CacheRateProfiles, CacheRateProfilesFilterIndexes,
		CacheRateFilterIndexes, CacheActionProfilesFilterIndexes, CacheAccountsFilterIndexes,
		CacheReverseFilterIndexes, CacheAccounts,
	})

	// CachePartitions enables creation of cache partitions
	CachePartitions = JoinStringSet(extraDBPartition, DataDBPartitions)

	CacheInstanceToPrefix = map[string]string{
		CacheResourceProfiles:            ResourceProfilesPrefix,
		CacheResources:                   ResourcesPrefix,
		CacheIPProfiles:                  IPProfilesPrefix,
		CacheIPAllocations:               IPAllocationsPrefix,
		CacheStatQueueProfiles:           StatQueueProfilePrefix,
		CacheStatQueues:                  StatQueuePrefix,
		CacheTrendProfiles:               TrendProfilePrefix,
		CacheTrends:                      TrendPrefix,
		CacheThresholdProfiles:           ThresholdProfilePrefix,
		CacheThresholds:                  ThresholdPrefix,
		CacheFilters:                     FilterPrefix,
		CacheRouteProfiles:               RouteProfilePrefix,
		CacheRankingProfiles:             RankingPrefix,
		CacheRankings:                    RankingProfilePrefix,
		CacheAttributeProfiles:           AttributeProfilePrefix,
		CacheChargerProfiles:             ChargerProfilePrefix,
		CacheRateProfiles:                RateProfilePrefix,
		CacheActionProfiles:              ActionProfilePrefix,
		CacheAccounts:                    AccountPrefix,
		CacheResourceFilterIndexes:       ResourceFilterIndexes,
		CacheIPFilterIndexes:             IPFilterIndexes,
		CacheStatFilterIndexes:           StatFilterIndexes,
		CacheThresholdFilterIndexes:      ThresholdFilterIndexes,
		CacheRouteFilterIndexes:          RouteFilterIndexes,
		CacheAttributeFilterIndexes:      AttributeFilterIndexes,
		CacheChargerFilterIndexes:        ChargerFilterIndexes,
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
		CacheIPFilterIndexes:             IPProfilesPrefix,
		CacheStatFilterIndexes:           StatQueueProfilePrefix,
		CacheRouteFilterIndexes:          RouteProfilePrefix,
		CacheAttributeFilterIndexes:      AttributeProfilePrefix,
		CacheChargerFilterIndexes:        ChargerProfilePrefix,
		CacheRateProfilesFilterIndexes:   RateProfilePrefix,
		CacheActionProfilesFilterIndexes: ActionProfilePrefix,
		CacheAccountsFilterIndexes:       AccountPrefix,
		CacheReverseFilterIndexes:        FilterPrefix,
	}

	CacheInstanceToCacheIndex = map[string]string{
		CacheThresholdProfiles: CacheThresholdFilterIndexes,
		CacheResourceProfiles:  CacheResourceFilterIndexes,
		CacheIPProfiles:        CacheIPFilterIndexes,
		CacheStatQueueProfiles: CacheStatFilterIndexes,
		CacheRouteProfiles:     CacheRouteFilterIndexes,
		CacheAttributeProfiles: CacheAttributeFilterIndexes,
		CacheChargerProfiles:   CacheChargerFilterIndexes,
		CacheRateProfiles:      CacheRateProfilesFilterIndexes,
		CacheActionProfiles:    CacheActionProfilesFilterIndexes,
		CacheFilters:           CacheReverseFilterIndexes,
		CacheAccounts:          CacheAccountsFilterIndexes,
	}

	// ProtectedSFlds are the fields that sessions should not alter
	ProtectedSFlds = NewStringSet([]string{OriginHost, OriginID, Usage})

	ConcurrentReqsLimit    int
	ConcurrentReqsStrategy string
)

const (
	CGRateS                  = "CGRateS"
	CGRateSorg               = "cgrates.org"
	Version                  = "v1.0~dev"
	DiameterFirmwareRevision = 918
	CGRateSLwr               = "cgrates"
	Postgres                 = "postgres"
	MySQL                    = "mysql"
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
	RSRSep                   = ";"
	MetaPipe                 = "*|"
	FieldsSep                = ","
	InInFieldSep             = ":"
	StaticHDRValSep          = "::"
	FilterValStart           = "("
	FilterValEnd             = ")"
	PlusChar                 = "+"
	DecNaN                   = `"NaN"`
	JSON                     = "json"
	JSONCaps                 = "JSON"
	GOBCaps                  = "GOB"
	MsgPack                  = "msgpack"
	CSVLoad                  = "CSVLOAD"
	MetaCDRID                = "*cdrID"
	MetaOriginID             = "*originID"
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

	CustomValue = "CustomValue"
	Value       = "Value"
	Rules       = "Rules"
	Metrics     = "Metrics"
	MetricID    = "MetricID"
	LastUsed    = "LastUsed"
	PDD         = "PDD"
	RouteStr    = "Route"
	RunID       = "RunID"
	MetaRunID   = "*runID"

	AttributeIDs            = "AttributeIDs"
	MetaOptsRunID           = "*opts.*runID"
	MetaReqRunID            = "*req.RunID"
	Cost                    = "Cost"
	CostDetails             = "CostDetails"
	Rated                   = "rated"
	Partial                 = "Partial"
	PreRated                = "PreRated"
	StaticValuePrefix       = "^"
	CSV                     = "csv"
	FWV                     = "fwv"
	MetaMongo               = "*mongo"
	MetaRedis               = "*redis"
	MetaPostgres            = "*postgres"
	MetaInternal            = "*internal"
	MetaLocalHost           = "*localhost"
	MetaBiJSONLocalHost     = "*bijson_localhost"
	MetaRatingSubjectPrefix = "*zero"
	OK                      = "OK"
	MetaFileXML             = "*fileXML"
	MetaFileJSON            = "*fileJSON"
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
	ConfigPrefix              = "cfg_"
	ResourcesPrefix           = "res_"
	ResourceProfilesPrefix    = "rsp_"
	IPAllocationsPrefix       = "ips_"
	IPProfilesPrefix          = "ipa_"
	ThresholdPrefix           = "thd_"
	FilterPrefix              = "ftr_"
	CDRsStatsPrefix           = "cst_"
	VersionPrefix             = "ver_"
	StatQueueProfilePrefix    = "sqp_"
	RouteProfilePrefix        = "rpp_"
	AttributeProfilePrefix    = "alp_"
	ChargerProfilePrefix      = "cpp_"
	RateProfilePrefix         = "rtp_"
	ActionProfilePrefix       = "acp_"
	AccountPrefix             = "acn_"
	ThresholdProfilePrefix    = "thp_"
	StatQueuePrefix           = "stq_"
	RankingProfilePrefix      = "rgp_"
	TrendProfilePrefix        = "trp_"
	TrendPrefix               = "trd_"
	LoadIDPrefix              = "lid_"
	LoadInstKey               = "load_history"
	CreateCDRsTablesSQL       = "create_cdrs_tables.sql"
	CreateTariffPlanTablesSQL = "create_tariffplan_tables.sql"
	TestSQL                   = "TEST_SQL"
	MetaAsc                   = "*asc"
	MetaDesc                  = "*desc"
	MetaAscending             = "*ascending"
	MetaDescending            = "*descending"
	MetaConstant              = "*constant"
	MetaPositive              = "*positive"
	MetaNegative              = "*negative"
	MetaLast                  = "*last"
	MetaPassword              = "*password"
	MetaFiller                = "*filler"
	MetaHTTPPost              = "*httpPost"
	JanusAdminSubProto        = "janus-admin-protocol"
	MetaHTTPjsonCDR           = "*http_json_cdr"
	MetaHTTPjsonMap           = "*httpJSONMap"
	MetaAMQPjsonCDR           = "*amqp_json_cdr"
	MetaAMQPjsonMap           = "*amqpJSONMap"
	MetaAMQPV1jsonMap         = "*amqpv1JSONMap"
	MetaSQSjsonMap            = "*sqsJSONMap"
	MetaKafkajsonMap          = "*kafkaJSONMap"
	MetaNATSJSONMap           = "*natsJSONMap"
	MetaSQL                   = "*sql"
	MetaMySQL                 = "*mysql"
	MetaS3jsonMap             = "*s3JSONMap"
	ConfigPath                = "/etc/cgrates/"
	DisconnectCause           = "DisconnectCause"
	MetaRating                = "*rating"
	MetaAccounting            = "*accounting"
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
	MetaUsageDifference    = "*usageDifference"
	MetaDifference         = "*difference"
	MetaVariable           = "*variable"
	MetaCCUsage            = "*ccUsage"
	MetaSIPCID             = "*sipcid"
	MetaValueExponent      = "*valueExponent"
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
	//ExportLogger
	Message   = "Message"
	Severity  = "Severity"
	Timestamp = "Timestamp"

	XML                      = "xml"
	MetaGOB                  = "*gob"
	MetaJSON                 = "*json"
	MetaDateTime             = "*datetime"
	MetaMaskedDestination    = "*masked_destination"
	MetaUnixTimestamp        = "*unixTimestamp"
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
	UnsupportedServiceIDCaps = "UNSUPPORTED_SERVICE_ID"
	ServiceManager           = "ServiceManager"
	ServiceAlreadyRunning    = "service already running"
	RunningCaps              = "RUNNING"
	StoppedCaps              = "STOPPED"
	MetaAdminS               = "*admins"
	MetaReplicator           = "*replicator"
	MetaRerate               = "*rerate"
	MetaRefund               = "*refund"
	MetaStats                = "*stats"
	MetaTrends               = "*trends"
	MetaRankings             = "*rankings"
	MetaCore                 = "*core"
	MetaServiceManager       = "*servicemanager"
	MetaChargers             = "*chargers"
	MetaConfig               = "*config"
	MetaTpes                 = "*tpes"
	MetaFilters              = "*filters"
	MetaCDRs                 = "*cdrs"
	MetaDC                   = "*dc"
	MetaCaches               = "*caches"
	MetaUCH                  = "*uch"
	MetaGuardian             = "*guardians"
	MetaEEs                  = "*ees"
	MetaEFs                  = "*efs"
	MetaERs                  = "*ers"
	MetaRates                = "*rates"
	MetaRateSOverwrite       = "*rateSOverwrite"
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
	FileLockPrefix           = "file_"
	ActionsPoster            = "act"
	MetaFileCSV              = "*fileCSV"
	MetaVirt                 = "*virt"
	MetaElastic              = "*els"
	MetaFileFWV              = "*fileFWV"
	MetaFile                 = "*file"
	AccountsStr              = "Accounts"
	AccountS                 = "AccountS"
	Actions                  = "Actions"
	BalanceMap               = "BalanceMap"
	UnitCounters             = "UnitCounters"
	UpdateTime               = "UpdateTime"
	Rates                    = "Rates"
	RateField                = "Rate"
	Format                   = "Format"
	Conn                     = "Conn"
	Level                    = "Level"
	FailedPostsDir           = "FailedPostsDir"
	//DestinationRates         = "DestinationRates"
	RatingPlans          = "RatingPlans"
	RatingProfiles       = "RatingProfiles"
	AccountActions       = "AccountActions"
	ResourcesStr         = "Resources"
	Stats                = "Stats"
	Rankings             = "Rankings"
	Trends               = "Trends"
	Filters              = "Filters"
	RateProfiles         = "RateProfiles"
	ActionProfiles       = "ActionProfiles"
	AccountsString       = "Accounts"
	MetaEveryMinute      = "*every_minute"
	MetaHourly           = "*hourly"
	ID                   = "ID"
	Address              = "Address"
	Addresses            = "Addresses"
	Transport            = "Transport"
	ClientKey            = "ClientKey"
	ClientCertificate    = "ClientCertificate"
	CaCertificate        = "CaCertificate"
	ConnectAttempts      = "ConnectAttempts"
	Reconnects           = "Reconnects"
	MaxReconnectInterval = "MaxReconnectInterval"
	ConnectTimeout       = "ConnectTimeout"
	ReplyTimeout         = "ReplyTimeout"
	TLS                  = "TLS"
	Strategy             = "Strategy"
	StrategyParameters   = "StrategyParameters"
	ConnID               = "ConnID"
	ConnFilterIDs        = "ConnFilterIDs"
	ConnWeight           = "ConnWeight"
	ConnBlocker          = "ConnBlocker"
	ConnParameters       = "ConnParameters"

	Thresholds   = "Thresholds"
	Routes       = "Routes"
	Attributes   = "Attributes"
	Chargers     = "Chargers"
	StatS        = "StatS"
	LoadIDsVrs   = "LoadIDs"
	GlobalVarS   = "GlobalVarS"
	CostSource   = "CostSource"
	ExtraInfo    = "ExtraInfo"
	Meta         = "*"
	MetaSysLog   = "*syslog"
	MetaStdLog   = "*stdout"
	MetaKafkaLog = "*kafkaLog"
	Kafka        = "Kafka"
	EventSource  = "EventSource"
	AccountID    = "AccountID"
	AccountIDs   = "AccountIDs"
	ResourceID   = "ResourceID"
	TotalUsage   = "TotalUsage"
	StatID       = "StatID"
	BalanceType  = "BalanceType"
	BalanceID    = "BalanceID"
	//BalanceDestinationIds = "BalanceDestinationIds"
	BalanceCostIncrements = "BalanceCostIncrements"
	BalanceAttributeIDs   = "BalanceAttributeIDs"
	BalanceRateProfileIDs = "BalanceRateProfileIDs"
	BalanceWeights        = "BalanceWeights"
	BalanceRatingSubject  = "BalanceRatingSubject"
	BalanceCategories     = "BalanceCategories"
	BalanceBlockers       = "BalanceBlockers"
	BalanceDisabled       = "BalanceDisabled"
	Units                 = "Units"
	AccountUpdate         = "AccountUpdate"
	BalanceUpdate         = "BalanceUpdate"
	StatUpdate            = "StatUpdate"
	TrendUpdate           = "TrendUpdate"
	RankingUpdate         = "RankingUpdate"
	ResourceUpdate        = "ResourceUpdate"
	ProcessTime           = "ProcessTime"
	CDRKey                = "CDR"
	CDRs                  = "CDRs"
	ExpiryTime            = "ExpiryTime"
	AllowNegative         = "AllowNegative"
	Disabled              = "Disabled"
	Action                = "Action"

	SessionSCosts = "SessionSCosts"
	RQF           = "RQF"
	ResourceStr   = "Resource"
	User          = "User"
	Subscribers   = "Subscribers"
	//Destinations             = "Destinations"
	MetaSubscribers          = "*subscribers"
	MetaDataDB               = "*datadb"
	MetaStorDB               = "*stordb"
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
	AddressPool              = "AddressPool"
	Pools                    = "Pools"
	Allocation               = "Allocation"
	Range                    = "Range"
	Stored                   = "Stored"
	RatingSubject            = "RatingSubject"
	Categories               = "Categories"
	Blocker                  = "Blocker"
	Blockers                 = "Blockers"
	Params                   = "Params"
	StartTime                = "StartTime"
	EndTime                  = "EndTime"
	ProcessingTime           = "ProcessingTime"
	AccountSummary           = "AccountSummary"
	RatingFilters            = "RatingFilters"
	RatingFilter             = "RatingFilter"
	Charging                 = "Charging"
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
	ChargingID               = "ChargingID"
	RatingID                 = "RatingID"
	JoinedChargeIDs          = "JoinedChargeIDs"
	UnitFactorID             = "UnitFactorID"
	ExtraChargeID            = "ExtraChargeID"
	BalanceLimit             = "BalanceLimit"
	ConnectFee               = "ConnectFee"
	RoundingMethod           = "RoundingMethod"
	RoundingDecimals         = "RoundingDecimals"
	MaxCostStrategy          = "MaxCostStrategy"
	RateID                   = "RateID"
	RateIDs                  = "RateIDs"
	RateFilterIDs            = "RateFilterIDs"
	RateActivationStart      = "RateActivationStart"
	RateWeights              = "RateWeights"
	RateIntervalStart        = "RateIntervalStart"
	RateFixedFee             = "RateFixedFee"
	RateRecurrentFee         = "RateRecurrentFee"
	RateBlocker              = "RateBlocker"
	RatesID                  = "RatesID"
	RatingFiltersID          = "RatingFiltersID"
	RateProfileID            = "RateProfileID"
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
	RouteRateProfileIDs      = "RouteRateProfileIDs"
	RouteStatIDs             = "RouteStatIDs"
	StatIDs                  = "StatIDs"
	SortedStatIDs            = "SortedStatIDs"
	LastUpdate               = "LastUpdate"
	TrendID                  = "TrendID"
	RankingID                = "RankingID"
	RouteWeights             = "RouteWeights"
	RouteParameters          = "RouteParameters"
	RouteBlockers            = "RouteBlockers"
	RouteResourceIDs         = "RouteResourceIDs"
	ResourceIDs              = "ResourceIDs"
	RouteFilterIDs           = "RouteFilterIDs"
	AttributeFilterIDs       = "AttributeFilterIDs"
	AttributeBlockers        = "AttributeBlockers"
	QueueLength              = "QueueLength"
	QueryInterval            = "QueryInterval"
	CorrelationType          = "CorrelationType"
	Tolerance                = "Tolerance"
	TTL                      = "TTL"
	PurgeFilterIDs           = "PurgeFilterIDs"
	TrendStr                 = "Trend"
	MinItems                 = "MinItems"
	MetricIDs                = "MetricIDs"
	MetricFilterIDs          = "MetricFilterIDs"
	MetricBlockers           = "MetricBlockers"
	FieldName                = "FieldName"
	Path                     = "Path"
	Hosts                    = "Hosts"
	StrategyParams           = "StrategyParams"
	MetaRound                = "*round"
	Pong                     = "Pong"
	MetaEventCost            = "*event_cost"
	MetaPositiveExports      = "*positive_exports"
	MetaNegativeExports      = "*negative_exports"
	MetaBuffer               = "*buffer"
	MetaRoutesEventCost      = "*routesEventCost"
	Freeswitch               = "freeswitch"
	Kamailio                 = "kamailio"
	Opensips                 = "opensips"
	Asterisk                 = "asterisk"
	SchedulerS               = "SchedulerS"
	MetaMultiply             = "*multiply"
	MetaDivide               = "*divide"
	MetaUrl                  = "*url"
	MetaZip                  = "*zip"
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
	MetaDAStats              = "*daStats"
	MetaDAThresholds         = "*daThresholds"
	MetaRAStats              = "*raStats"
	MetaRAThresholds         = "*raThresholds"
	MetaDNSStats             = "*dnsStats"
	MetaDNSThresholds        = "*dnsThresholds"
	MetaHAStats              = "*haStats"
	MetaHAThresholds         = "*haThresholds"
	MetaSAStats              = "*saStats"
	MetaSAThresholds         = "*saThresholds"
	MetaERsStats             = "*ersStats"
	MetaERsThresholds        = "*ersThresholds"
	MetaDryRun               = "*dryRun"
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

	MetaMemory              = "*memory"
	RemoteHost              = "RemoteHost"
	Local                   = "local"
	TCP                     = "tcp"
	UDP                     = "udp"
	CGRDebitInterval        = "CGRDebitInterval"
	VersionName             = "Version"
	MetaTenant              = "*tenant"
	ResourceUsageStr        = "ResourceUsage"
	MetaDuration            = "*duration"
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
	MetaJoin                = "*join"
	MetaSplit               = "*split"
	MetaStrip               = "*strip"
	MetaReload              = "*reload"
	MetaLoad                = "*load"
	MetaFloat64             = "*float64"
	MetaRemove              = "*remove"
	MetaRemoveAll           = "*removeall"
	MetaStore               = "*store"
	MetaParse               = "*parse"
	MetaClear               = "*clear"
	MetaExport              = "*export"
	MetaGigawords           = "*gigawords"
	MetaExportID            = "*export_id"
	LoadIDs                 = "load_ids"
	DNSAgent                = "DNSAgent"
	TLSNoCaps               = "tls"
	UsageID                 = "UsageID"
	AllocationID            = "AllocationID"
	Replacement             = "Replacement"
	Regexp                  = "Regexp"
	Order                   = "Order"
	Preference              = "Preference"
	Flags                   = "Flags"
	Service                 = "Service"
	MetaAnalyzerS           = "*analyzers"
	CGREventString          = "CGREvent"
	MetaTextPlain           = "*text_plain"
	MetaRelease             = "*release"
	MetaAllocate            = "*allocate"
	MetaAuthorize           = "*authorize"
	MetaSTIRAuthenticate    = "*stir_authenticate"
	MetaSTIRInitiate        = "*stir_initiate"
	MetaInit                = "*init"
	ERs                     = "ERs"
	EEs                     = "EEs"
	EFs                     = "EFs"
	Ratio                   = "Ratio"
	Load                    = "Load"
	Slash                   = "/"
	UUID                    = "UUID"
	ActionsID               = "ActionsID"
	MetaAct                 = "*act"
	ExportTemplate          = "ExportTemplate"
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
	MetaCost                = "*cost"
	MetaRateSCost           = "*rateSCost"
	MetaAccountSCost        = "*accountSCost"
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
	ProfileID               = "ProfileID"
	SortedRoutes            = "SortedRoutes"
	MetaMonthly             = "*monthly"
	MetaYearly              = "*yearly"
	MetaDaily               = "*daily"
	MetaWeekly              = "*weekly"
	RateS                   = "RateS"
	Underline               = "_"
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
	Abstracts               = "Abstracts"
	Concretes               = "Concretes"
	ActionS                 = "ActionS"
	Schedule                = "Schedule"
	ActionFilterIDs         = "ActionFilterIDs"
	ActionTTL               = "ActionTTL"
	ActionOpts              = "ActionOpts"
	ActionPath              = "ActionPath"
	TPid                    = "TPid"
	LoadId                  = "LoadId"
	ActionPlanId            = "ActionPlanId"
	AccountActionsId        = "AccountActionsId"
	Loadid                  = "loadid"
	ActionPlan              = "ActionPlan"
	ActionsId               = "ActionsId"
	Prefixes                = "Prefixes"
	RateSlots               = "RateSlots"
	RatingPlanBindings      = "RatingPlanBindings"
	RatingPlanActivations   = "RatingPlanActivations"
	Time                    = "Time"
	TargetIDs               = "TargetIDs"
	TargetType              = "TargetType"
	MetaRow                 = "*row"
	BalanceFilterIDs        = "BalanceFilterIDs"
	BalanceOpts             = "BalanceOpts"
	MetaConcrete            = "*concrete"
	MetaAbstract            = "*abstract"
	MetaMockAbstract        = "*mockabstract"
	MetaBalanceLimit        = "*balanceLimit"
	MetaBalanceUnlimited    = "*balanceUnlimited"
	MetaTemplateID          = "*templateID"
	MetaCdrLog              = "*cdrLog"
	MetaCDR                 = "*cdr"
	MetaExporterIDs         = "*exporterIDs"
	MetaExporterID          = "*exporterID"
	MetaChargeID            = "*chargeID"
	MetaAsync               = "*async"
	MetaUsage               = "*usage"
	MetaDestination         = "*destination"
	MetaStartTime           = "*startTime"
	Weights                 = "Weights"
	ActivationTimes         = "ActivationTimes"
	IntervalRates           = "IntervalRates"
	IntervalStart           = "IntervalStart"
	Unit                    = "Unit"
	Targets                 = "Targets"
	Balances                = "Balances"
	UnitFactorField         = "UnitFactor"
	UnitFactors             = "UnitFactors"
	JoinedCharge            = "JoinedCharge"
	CostIncrements          = "CostIncrements"
	Factor                  = "Factor"
	Increment               = "Increment"
	FixedFee                = "FixedFee"
	RecurrentFee            = "RecurrentFee"
	IncrementStart          = "IncrementStart"
	RateIntervalIndex       = "RateIntervalIndex"
	Diktats                 = "Diktats"
	BalanceIDs              = "BalanceIDs"
	MetaCostIncrement       = "*costIncrement"
	Length                  = "Length"
	V1Prfx                  = "V1"
	Ping                    = "Ping"

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
	MetaTntID             = "*tntID"
)

// CoreS metric constants
const (

	// condensed result fields
	PID            = "pid"
	NodeID         = "node_id"
	GoVersion      = "go_version"
	OSThreadsInUse = "os_threads_in_use"
	RunningSince   = "running_since"
	OpenFiles      = "open_files"
	ActiveMemory   = "active_memory"
	SystemMemory   = "system_memory"

	FieldVersion         = "version"
	FieldMemStats        = "mem_stats"
	FieldGCDurationStats = "gc_duration_stats"
	FieldProcStats       = "proc_stats"
	FieldCapsStats       = "caps_stats"

	MetricRuntimeGoroutines = "goroutines"
	MetricRuntimeThreads    = "threads"
	MetricRuntimeMaxProcs   = "maxprocs"

	MetricMemAlloc        = "alloc"
	MetricMemTotalAlloc   = "total_alloc"
	MetricMemSys          = "sys"
	MetricMemMallocs      = "mallocs"
	MetricMemFrees        = "frees"
	MetricMemHeapAlloc    = "heap_alloc"
	MetricMemHeapSys      = "heap_sys"
	MetricMemHeapIdle     = "heap_idle"
	MetricMemHeapInuse    = "heap_inuse"
	MetricMemHeapReleased = "heap_released"
	MetricMemHeapObjects  = "heap_objects"
	MetricMemStackInuse   = "stack_inuse"
	MetricMemStackSys     = "stack_sys"
	MetricMemMSpanSys     = "mspan_sys"
	MetricMemMSpanInuse   = "mspan_inuse"
	MetricMemMCacheInuse  = "mcache_inuse"
	MetricMemMCacheSys    = "mcache_sys"
	MetricMemBuckHashSys  = "buckhash_sys"
	MetricMemGCSys        = "gc_sys"
	MetricMemOtherSys     = "other_sys"
	MetricMemNextGC       = "next_gc"
	MetricMemLastGC       = "last_gc"
	MetricMemLimit        = "mem_limit"

	MetricProcCPUTime              = "cpu_time"
	MetricProcMaxFDs               = "max_fds"
	MetricProcOpenFDs              = "open_fds"
	MetricProcResidentMemory       = "resident_memory"
	MetricProcStartTime            = "start_time"
	MetricProcVirtualMemory        = "virtual_memory"
	MetricProcMaxVirtualMemory     = "max_virtual_memory"
	MetricProcNetworkReceiveTotal  = "network_receive_total"
	MetricProcNetworkTransmitTotal = "network_transmit_total"

	MetricGCQuantiles = "quantiles"
	MetricGCQuantile  = "quantile"
	MetricGCValue     = "value"
	MetricGCSum       = "sum"
	MetricGCCount     = "count"
	MetricGCPercent   = "gc_percent"

	MetricCapsAllocated = "caps_allocated"
	MetricCapsPeak      = "caps_peak"
)

// Migrator Action
const (
	Move    = "move"
	Migrate = "migrate"
)

// Meta Items
const (
	MetaAccounts          = "*accounts"
	MetaActions           = "*actions"
	MetaResourceProfiles  = "*resource_profiles"
	MetaIPProfiles        = "*ip_profiles"
	MetaStatQueueProfiles = "*statqueue_profiles"
	MetaStatQueues        = "*statqueues"
	MetaRankingProfiles   = "*ranking_profiles"
	MetaTrendProfiles     = "*trend_profiles"
	MetaThresholdProfiles = "*threshold_profiles"
	MetaRouteProfiles     = "*route_profiles"
	MetaAttributeProfiles = "*attribute_profiles"
	MetaRateProfiles      = "*rate_profiles"
	MetaRateProfileRates  = "*rate_profile_rates"
	MetaChargerProfiles   = "*charger_profiles"
	MetaIPAllocations     = "*ip_allocations"
	MetaThresholds        = "*thresholds"
	MetaRoutes            = "*routes"
	MetaAttributes        = "*attributes"
	MetaActionProfiles    = "*action_profiles"
	MetaLoadIDs           = "*load_ids"
	MetaNodeID            = "*node_id"
	MetaIPv4              = "*ipv4"
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
	SessionS        = "SessionS"
	AttributeS      = "AttributeS"
	RouteS          = "RouteS"
	ResourceS       = "ResourceS"
	IPs             = "IPs"
	StatService     = "StatS"
	FilterS         = "FilterS"
	ThresholdS      = "ThresholdS"
	TrendS          = "TrendS"
	RankingS        = "RankingS"
	RegistrarC      = "RegistrarC"
	LoaderS         = "LoaderS"
	ChargerS        = "ChargerS"
	TPeS            = "TPeS"
	CacheS          = "CacheS"
	AnalyzerS       = "AnalyzerS"
	CDRServer       = "CDRServer"
	GuardianS       = "GuardianS"
	ServiceManagerS = "ServiceManager"
	CommonListenerS = "CommonListenerS"
	ConnManager     = "ConnManager"
	LoggerS         = "LoggerS"
	CapS            = "CapS"
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
	MetaRpc                = "*rpc"
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
	BalanceUnitFactors     = "BalanceUnitFactors"
	ExtraParameters        = "ExtraParameters"

	MetaAddBalance            = "*add_balance"
	MetaSetBalance            = "*set_balance"
	MetaRemBalance            = "*rem_balance"
	DynaprepaidActionplansCfg = "dynaprepaid_actionprofile"
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
	TpFilters        = "TpFilters"
	TpThresholds     = "TpThresholds"
	TpRoutes         = "TpRoutes"
	TpAttributes     = "TpAttributes"
	TpStats          = "TpStats"
	TpResources      = "TpResources"
	TpResource       = "TpResource"
	TpChargers       = "TpChargers"
	TpRateProfiles   = "TpRateProfiles"
	TpActionProfiles = "TpActionProfiles"
	TpAccounts       = "TpAccounts"
)

// Dispatcher Const
const (
	MetaFirst          = "*first"
	MetaRandom         = "*random"
	MetaRoundRobin     = "*round_robin"
	MetaRatio          = "*ratio"
	MetaDefaultRatio   = "*default_ratio"
	ThresholdSv1       = "ThresholdSv1"
	TrendSv1           = "TrendSv1"
	RankingSv1         = "RankingSv1"
	StatSv1            = "StatSv1"
	ResourceSv1        = "ResourceSv1"
	IPsV1              = "IPsV1"
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
	MetaContains           = "*contains"
	MetaPrefix             = "*prefix"
	MetaSuffix             = "*suffix"
	MetaBoth               = "*both"
	MetaEmpty              = "*empty"
	MetaExists             = "*exists"
	MetaCronExp            = "*cronexp"
	MetaRSR                = "*rsr"
	MetaLessThan           = "*lt"
	MetaLessOrEqual        = "*lte"
	MetaGreaterThan        = "*gt"
	MetaGreaterOrEqual     = "*gte"
	MetaResources          = "*resources"
	MetaIPs                = "*ips"
	MetaEqual              = "*eq"
	MetaIPNet              = "*ipnet"
	MetaAPIBan             = "*apiban"
	MetaSentryPeer         = "*sentrypeer"
	MetaToken              = "*token"
	MetaIP                 = "*ip"
	MetaNumber             = "*number"
	MetaActivationInterval = "*ai"
	MetaRegex              = "*regex"
	MetaNever              = "*never"

	MetaNotString             = "*notstring"
	MetaNotContains           = "*notcontains"
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
	MetaNotSentryPeer         = "*notsentrypeer"
	MetaNotActivationInterval = "*notai"
	MetaNotRegex              = "*notregex"
	MetaHTTP                  = "*http"

	MetaEC = "*ec"
)

// ReplicatorSv1 APIs
const (
	ReplicatorS                       = "ReplicatorS"
	ReplicatorSv1                     = "ReplicatorSv1"
	ReplicatorSv1Ping                 = "ReplicatorSv1.Ping"
	ReplicatorSv1GetStatQueue         = "ReplicatorSv1.GetStatQueue"
	ReplicatorSv1GetFilter            = "ReplicatorSv1.GetFilter"
	ReplicatorSv1GetThreshold         = "ReplicatorSv1.GetThreshold"
	ReplicatorSv1GetThresholdProfile  = "ReplicatorSv1.GetThresholdProfile"
	ReplicatorSv1GetStatQueueProfile  = "ReplicatorSv1.GetStatQueueProfile"
	ReplicatorSv1GetRanking           = "ReplicatorSv1.GetRanking"
	ReplicatorSv1GetRankingProfile    = "ReplicatorSv1.GetRankingProfile"
	ReplicatorSv1GetTrendProfile      = "ReplicatorSv1.GetTrendProfile"
	ReplicatorSv1GetTrend             = "ReplicatorSv1.GetTrend"
	ReplicatorSv1GetResource          = "ReplicatorSv1.GetResource"
	ReplicatorSv1GetResourceProfile   = "ReplicatorSv1.GetResourceProfile"
	ReplicatorSv1GetIPAllocations     = "ReplicatorSv1.GetIPAllocations"
	ReplicatorSv1GetIPProfile         = "ReplicatorSv1.GetIPProfile"
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
	ReplicatorSv1SetRankingProfile    = "ReplicatorSv1.SetRankingProfile"
	ReplicatorSv1SetRanking           = "ReplicatorSv1.SetRanking"
	ReplicatorSv1SetTrendProfile      = "ReplicatorSv1.SetTrendProfile"
	ReplicatorSv1SetTrend
	ReplicatorSv1SetResource         = "ReplicatorSv1.SetResource"
	ReplicatorSv1SetResourceProfile  = "ReplicatorSv1.SetResourceProfile"
	ReplicatorSv1SetIPAllocations    = "ReplicatorSv1.SetIPAllocations"
	ReplicatorSv1SetIPProfile        = "ReplicatorSv1.SetIPProfile"
	ReplicatorSv1SetRouteProfile     = "ReplicatorSv1.SetRouteProfile"
	ReplicatorSv1SetAttributeProfile = "ReplicatorSv1.SetAttributeProfile"
	ReplicatorSv1SetChargerProfile   = "ReplicatorSv1.SetChargerProfile"
	ReplicatorSv1SetRateProfile      = "ReplicatorSv1.SetRateProfile"
	ReplicatorSv1SetActionProfile    = "ReplicatorSv1.SetActionProfile"
	ReplicatorSv1SetAccount          = "ReplicatorSv1.SetAccount"
	ReplicatorSv1SetLoadIDs          = "ReplicatorSv1.SetLoadIDs"
	ReplicatorSv1RemoveThreshold     = "ReplicatorSv1.RemoveThreshold"

	ReplicatorSv1RemoveStatQueue        = "ReplicatorSv1.RemoveStatQueue"
	ReplicatorSv1RemoveFilter           = "ReplicatorSv1.RemoveFilter"
	ReplicatorSv1RemoveThresholdProfile = "ReplicatorSv1.RemoveThresholdProfile"
	ReplicatorSv1RemoveStatQueueProfile = "ReplicatorSv1.RemoveStatQueueProfile"
	ReplicatorSv1RemoveRankingProfile   = "ReplicatorSv1.RemoveRankingProfile"
	ReplicatorSv1RemoveRanking          = "ReplicatorSv1.RemoveRanking"
	ReplicatorSv1RemoveTrendProfile     = "ReplicatorSv1.RemoveTrendProfile"
	ReplicatorSv1RemoveTrend            = "ReplicatorSv1.RemoveTrend"
	ReplicatorSv1RemoveResource         = "ReplicatorSv1.RemoveResource"
	ReplicatorSv1RemoveResourceProfile  = "ReplicatorSv1.RemoveResourceProfile"
	ReplicatorSv1RemoveIPAllocations    = "ReplicatorSv1.RemoveIPAllocations"
	ReplicatorSv1RemoveIPProfile        = "ReplicatorSv1.RemoveIPProfile"
	ReplicatorSv1RemoveRouteProfile     = "ReplicatorSv1.RemoveRouteProfile"
	ReplicatorSv1RemoveAttributeProfile = "ReplicatorSv1.RemoveAttributeProfile"
	ReplicatorSv1RemoveChargerProfile   = "ReplicatorSv1.RemoveChargerProfile"
	ReplicatorSv1RemoveRateProfile      = "ReplicatorSv1.RemoveRateProfile"
	ReplicatorSv1RemoveActionProfile    = "ReplicatorSv1.RemoveActionProfile"
	ReplicatorSv1RemoveAccount          = "ReplicatorSv1.RemoveAccount"
	ReplicatorSv1GetIndexes             = "ReplicatorSv1.GetIndexes"
	ReplicatorSv1SetIndexes             = "ReplicatorSv1.SetIndexes"
	ReplicatorSv1RemoveIndexes          = "ReplicatorSv1.RemoveIndexes"
)

// AdminSv1 APIs
const (
	//AdminSv1ReplayFailedPosts                 = "AdminSv1.ReplayFailedPosts"
	AdminSv1GetRateRatesIndexesHealth         = "AdminSv1.GetRateRatesIndexesHealth"
	AdminSv1GetChargerProfilesCount           = "AdminSv1.GetChargerProfilesCount"
	AdminSv1GetAccountsIndexesHealth          = "AdminSv1.GetAccountsIndexesHealth"
	AdminSv1GetDispatcherProfilesCount        = "AdminSv1.GetDispatcherProfilesCount"
	AdminSv1GetRouteProfilesCount             = "AdminSv1.GetRouteProfilesCount"
	AdminSv1GetActionsIndexesHealth           = "AdminSv1.GetActionsIndexesHealth"
	AdminSv1GetDispatcherHostsCount           = "AdminSv1.GetDispatcherHostsCount"
	AdminSv1GetRateProfilesIndexesHealth      = "AdminSv1.GetRateProfilesIndexesHealth"
	AdminSv1ComputeFilterIndexes              = "AdminSv1.ComputeFilterIndexes"
	AdminSv1ComputeFilterIndexIDs             = "AdminSv1.ComputeFilterIndexIDs"
	AdminSv1GetAccountActionPlansIndexHealth  = "AdminSv1.GetAccountActionPlansIndexHealth"
	AdminSv1GetReverseDestinationsIndexHealth = "AdminSv1.GetReverseDestinationsIndexHealth"
	AdminSv1GetReverseFilterHealth            = "AdminSv1.GetReverseFilterHealth"
	AdminSv1GetThresholdsIndexesHealth        = "AdminSv1.GetThresholdsIndexesHealth"
	AdminSv1GetResourcesIndexesHealth         = "AdminSv1.GetResourcesIndexesHealth"
	AdminSv1GetIPsIndexesHealth               = "AdminSv1.GetIPsIndexesHealth"
	AdminSv1GetStatsIndexesHealth             = "AdminSv1.GetStatsIndexesHealth"
	AdminSv1GetRoutesIndexesHealth            = "AdminSv1.GetRoutesIndexesHealth"
	AdminSv1GetChargersIndexesHealth          = "AdminSv1.GetChargersIndexesHealth"
	AdminSv1GetAttributesIndexesHealth        = "AdminSv1.GetAttributesIndexesHealth"
	AdminSv1GetDispatchersIndexesHealth       = "AdminSv1.GetDispatchersIndexesHealth"
	AdminSv1Ping                              = "AdminSv1.Ping"
	AdminSv1SetDispatcherProfile              = "AdminSv1.SetDispatcherProfile"
	AdminSv1GetDispatcherProfile              = "AdminSv1.GetDispatcherProfile"
	AdminSv1GetDispatcherProfiles             = "AdminSv1.GetDispatcherProfiles"
	AdminSv1GetDispatcherProfileIDs           = "AdminSv1.GetDispatcherProfileIDs"
	AdminSv1RemoveDispatcherProfile           = "AdminSv1.RemoveDispatcherProfile"
	// APIerSv1SetBalances                       = "APIerSv1.SetBalances"
	AdminSv1SetDispatcherHost    = "AdminSv1.SetDispatcherHost"
	AdminSv1GetDispatcherHost    = "AdminSv1.GetDispatcherHost"
	AdminSv1GetDispatcherHosts   = "AdminSv1.GetDispatcherHosts"
	AdminSv1GetDispatcherHostIDs = "AdminSv1.GetDispatcherHostIDs"
	AdminSv1RemoveDispatcherHost = "AdminSv1.RemoveDispatcherHost"
	// APIerSv1GetEventCost                      = "APIerSv1.GetEventCost"
	// APIerSv1LoadTariffPlanFromFolder          = "APIerSv1.LoadTariffPlanFromFolder"
	// APIerSv1ExportToFolder                    = "APIerSv1.ExportToFolder"
	// APIerSv1GetCost                           = "APIerSv1.GetCost"
	AdminSv1GetFilter           = "AdminSv1.GetFilter"
	AdminSv1GetFilterIndexes    = "AdminSv1.GetFilterIndexes"
	AdminSv1RemoveFilterIndexes = "AdminSv1.RemoveFilterIndexes"
	AdminSv1RemoveFilter        = "AdminSv1.RemoveFilter"
	AdminSv1SetFilter           = "AdminSv1.SetFilter"
	AdminSv1GetFilterIDs        = "AdminSv1.GetFilterIDs"
	AdminSv1GetFiltersCount     = "AdminSv1.GetFiltersCount"
	AdminSv1GetFilters          = "AdminSv1.GetFilters"
	AdminSv1FiltersMatch        = "AdminSv1.FiltersMatch"
	// APIerSv1SetDataDBVersions   = "APIerSv1.SetDataDBVersions"

	// APIerSv1GetActions          = "APIerSv1.GetActions"

	// APIerSv1GetDataDBVersions        = "APIerSv1.GetDataDBVersions"

	// APIerSv1GetCDRs                  = "APIerSv1.GetCDRs"
	// APIerSv1GetTPActions             = "APIerSv1.GetTPActions"
	// APIerSv1GetTPAttributeProfile    = "APIerSv1.GetTPAttributeProfile"
	// APIerSv1SetTPAttributeProfile    = "APIerSv1.SetTPAttributeProfile"
	// APIerSv1GetTPAttributeProfileIds = "APIerSv1.GetTPAttributeProfileIds"
	// APIerSv1RemoveTPAttributeProfile = "APIerSv1.RemoveTPAttributeProfile"
	// APIerSv1GetTPCharger             = "APIerSv1.GetTPCharger"
	// APIerSv1SetTPCharger             = "APIerSv1.SetTPCharger"
	// APIerSv1RemoveTPCharger          = "APIerSv1.RemoveTPCharger"
	// APIerSv1GetTPChargerIDs          = "APIerSv1.GetTPChargerIDs"
	// APIerSv1SetTPFilterProfile       = "APIerSv1.SetTPFilterProfile"
	// APIerSv1GetTPFilterProfile       = "APIerSv1.GetTPFilterProfile"
	// APIerSv1GetTPFilterProfileIds    = "APIerSv1.GetTPFilterProfileIds"
	// APIerSv1RemoveTPFilterProfile    = "APIerSv1.RemoveTPFilterProfile"

	// APIerSv1GetTPResource             = "APIerSv1.GetTPResource"
	// APIerSv1SetTPResource             = "APIerSv1.SetTPResource"
	// APIerSv1RemoveTPResource          = "APIerSv1.RemoveTPResource"
	// APIerSv1SetTPRate                 = "APIerSv1.SetTPRate"
	// APIerSv1GetTPRate                 = "APIerSv1.GetTPRate"
	// APIerSv1RemoveTPRate              = "APIerSv1.RemoveTPRate"
	// APIerSv1GetTPRateIds              = "APIerSv1.GetTPRateIds"
	// APIerSv1SetTPThreshold            = "APIerSv1.SetTPThreshold"
	// APIerSv1GetTPThreshold            = "APIerSv1.GetTPThreshold"
	// APIerSv1GetTPThresholdIDs         = "APIerSv1.GetTPThresholdIDs"
	// APIerSv1RemoveTPThreshold         = "APIerSv1.RemoveTPThreshold"
	// APIerSv1SetTPStat                 = "APIerSv1.SetTPStat"
	// APIerSv1GetTPStat                 = "APIerSv1.GetTPStat"
	// APIerSv1RemoveTPStat              = "APIerSv1.RemoveTPStat"
	// APIerSv1SetTPRouteProfile         = "APIerSv1.SetTPRouteProfile"
	// APIerSv1GetTPRouteProfile         = "APIerSv1.GetTPRouteProfile"
	// APIerSv1GetTPRouteProfileIDs      = "APIerSv1.GetTPRouteProfileIDs"
	// APIerSv1RemoveTPRouteProfile      = "APIerSv1.RemoveTPRouteProfile"
	// APIerSv1GetTPDispatcherProfile    = "APIerSv1.GetTPDispatcherProfile"
	// APIerSv1SetTPDispatcherProfile    = "APIerSv1.SetTPDispatcherProfile"
	// APIerSv1RemoveTPDispatcherProfile = "APIerSv1.RemoveTPDispatcherProfile"
	// APIerSv1GetTPDispatcherProfileIDs = "APIerSv1.GetTPDispatcherProfileIDs"
	// APIerSv1ExportCDRs                = "APIerSv1.ExportCDRs"
	// APIerSv1SetTPRatingPlan           = "APIerSv1.SetTPRatingPlan"
	// APIerSv1SetTPActions              = "APIerSv1.SetTPActions"
	// APIerSv1GetTPActionIds            = "APIerSv1.GetTPActionIds"
	// APIerSv1RemoveTPActions           = "APIerSv1.RemoveTPActions"
	// APIerSv1SetActionPlan             = "APIerSv1.SetActionPlan"
	// APIerSv1ExecuteAction             = "APIerSv1.ExecuteAction"
	// APIerSv1SetTPRatingProfile        = "APIerSv1.SetTPRatingProfile"
	// APIerSv1GetTPRatingProfile        = "APIerSv1.GetTPRatingProfile"

	// APIerSv1ImportTariffPlanFromFolder = "APIerSv1.ImportTariffPlanFromFolder"
	// APIerSv1ExportTPToFolder           = "APIerSv1.ExportTPToFolder"
	// APIerSv1SetActions                 = "APIerSv1.SetActions"

	// APIerSv1GetDataCost              = "APIerSv1.GetDataCost"
	// APIerSv1ReplayFailedPosts        = "APIerSv1.ReplayFailedPosts"
	// APIerSv1GetCacheStats            = "APIerSv1.GetCacheStats"
	// APIerSv1ReloadCache              = "APIerSv1.ReloadCache"
	// APIerSv1RemoveActions            = "APIerSv1.RemoveActions"
	// APIerSv1GetLoadHistory           = "APIerSv1.GetLoadHistory"
	// APIerSv1GetLoadIDs               = "APIerSv1.GetLoadIDs"
	// APIerSv1GetLoadTimes             = "APIerSv1.GetLoadTimes"
	AdminSv1GetAttributeProfilesCount = "AdminSv1.GetAttributeProfilesCount"
	AdminSv1SetAccount                = "AdminSv1.SetAccount"
	AdminSv1GetAccount                = "AdminSv1.GetAccount"
	AdminSv1GetAccounts               = "AdminSv1.GetAccounts"
	AdminSv1GetAccountIDs             = "AdminSv1.GetAccountIDs"
	AdminSv1RemoveAccount             = "AdminSv1.RemoveAccount"
	AdminSv1GetAccountsCount          = "AdminSv1.GetAccountsCount"
	AdminSv1GetCDRs                   = "AdminSv1.GetCDRs"
	AdminSv1RemoveCDRs                = "AdminSv1.RemoveCDRs"
)

const (
	ServiceManagerV1              = "ServiceManagerV1"
	ServiceManagerV1StartService  = "ServiceManagerV1.StartService"
	ServiceManagerV1StopService   = "ServiceManagerV1.StopService"
	ServiceManagerV1ServiceStatus = "ServiceManagerV1.ServiceStatus"
	ServiceManagerV1Ping          = "ServiceManagerV1.Ping"
)

// TPeSv1 APIs
const (
	TPeSv1                 = "TPeSv1"
	TPeSv1Ping             = "TPeSv1.Ping"
	TPeSv1ExportTariffPlan = "TPeSv1.ExportTariffPlan"
)

// EfSv1 APIs
const (
	EfSv1             = "EfSv1"
	EfSv1Ping         = "EfSv1.Ping"
	EfSv1ProcessEvent = "EfSv1.ProcessEvent"
	EfSv1ReplayEvents = "EfSv1.ReplayEvents"
)

// ERs
const (
	ErS            = "ErS"
	ErSv1          = "ErSv1"
	ErSv1Ping      = "ErSv1.Ping"
	ErSv1RunReader = "ErSv1.RunReader"
)

// ConfigSv1 APIs
const (
	ConfigS                    = "ConfigS"
	ConfigSv1                  = "ConfigSv1"
	ConfigSv1ReloadConfig      = "ConfigSv1.ReloadConfig"
	ConfigSv1GetConfig         = "ConfigSv1.GetConfig"
	ConfigSv1SetConfig         = "ConfigSv1.SetConfig"
	ConfigSv1GetConfigAsJSON   = "ConfigSv1.GetConfigAsJSON"
	ConfigSv1SetConfigFromJSON = "ConfigSv1.SetConfigFromJSON"
	ConfigSv1StoreCfgInDB      = "ConfigSv1.StoreCfgInDB"
	ConfigSv1Ping              = "ConfigSv1.Ping"
)

const (
	RateSv1                         = "RateSv1"
	RateSv1CostForEvent             = "RateSv1.CostForEvent"
	RateSv1RateProfilesForEvent     = "RateSv1.RateProfilesForEvent"
	RateSv1RateProfileRatesForEvent = "RateSv1.RateProfileRatesForEvent"
	RateSv1Ping                     = "RateSv1.Ping"
)

const (
	AccountSv1                    = "AccountSv1"
	AccountSv1Ping                = "AccountSv1.Ping"
	AccountSv1AccountsForEvent    = "AccountSv1.AccountsForEvent"
	AccountSv1MaxAbstracts        = "AccountSv1.MaxAbstracts"
	AccountSv1DebitAbstracts      = "AccountSv1.DebitAbstracts"
	AccountSv1MaxConcretes        = "AccountSv1.MaxConcretes"
	AccountSv1DebitConcretes      = "AccountSv1.DebitConcretes"
	AccountSv1RefundCharges       = "AccountSv1.RefundCharges"
	AccountSv1ActionSetBalance    = "AccountSv1.ActionSetBalance"
	AccountSv1ActionRemoveBalance = "AccountSv1.ActionRemoveBalance"
	AccountSv1GetAccount          = "AccountSv1.GetAccount"
)

const (
	CoreS                       = "CoreS"
	CoreSv1                     = "CoreSv1"
	CoreSv1Status               = "CoreSv1.Status"
	CoreSv1Ping                 = "CoreSv1.Ping"
	CoreSv1Panic                = "CoreSv1.Panic"
	CoreSv1Shutdown             = "CoreSv1.Shutdown"
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
	AdminSv1GetRouteProfile          = "AdminSv1.GetRouteProfile"
	AdminSv1GetRouteProfiles         = "AdminSv1.GetRouteProfiles"
	AdminSv1GetRouteProfileIDs       = "AdminSv1.GetRouteProfileIDs"
	AdminSv1RemoveRouteProfile       = "AdminSv1.RemoveRouteProfile"
	AdminSv1SetRouteProfile          = "AdminSv1.SetRouteProfile"
)

// AttributeS APIs
const (
	AdminSv1SetAttributeProfile      = "AdminSv1.SetAttributeProfile"
	AdminSv1GetAttributeProfile      = "AdminSv1.GetAttributeProfile"
	AdminSv1GetAttributeProfiles     = "AdminSv1.GetAttributeProfiles"
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
	AdminSv1GetChargerProfile     = "AdminSv1.GetChargerProfile"
	AdminSv1GetChargerProfiles    = "AdminSv1.GetChargerProfiles"
	AdminSv1RemoveChargerProfile  = "AdminSv1.RemoveChargerProfile"
	AdminSv1SetChargerProfile     = "AdminSv1.SetChargerProfile"
	AdminSv1GetChargerProfileIDs  = "AdminSv1.GetChargerProfileIDs"
)

// ThresholdS APIs
const (
	ThresholdSv1ProcessEvent          = "ThresholdSv1.ProcessEvent"
	ThresholdSv1GetThreshold          = "ThresholdSv1.GetThreshold"
	ThresholdSv1ResetThreshold        = "ThresholdSv1.ResetThreshold"
	ThresholdSv1GetThresholdIDs       = "ThresholdSv1.GetThresholdIDs"
	ThresholdSv1Ping                  = "ThresholdSv1.Ping"
	ThresholdSv1GetThresholdsForEvent = "ThresholdSv1.GetThresholdsForEvent"
	AdminSv1GetThresholdProfileIDs    = "AdminSv1.GetThresholdProfileIDs"
	AdminSv1GetThresholdProfilesCount = "AdminSv1.GetThresholdProfilesCount"
	AdminSv1GetThresholdProfile       = "AdminSv1.GetThresholdProfile"
	AdminSv1GetThresholdProfiles      = "AdminSv1.GetThresholdProfiles"
	AdminSv1RemoveThresholdProfile    = "AdminSv1.RemoveThresholdProfile"
	AdminSv1SetThresholdProfile       = "AdminSv1.SetThresholdProfile"
)

// StatS APIs
const (
	StatSv1ProcessEvent               = "StatSv1.ProcessEvent"
	StatSv1GetQueueIDs                = "StatSv1.GetQueueIDs"
	StatSv1GetQueueStringMetrics      = "StatSv1.GetQueueStringMetrics"
	StatSv1GetQueueFloatMetrics       = "StatSv1.GetQueueFloatMetrics"
	StatSv1GetQueueDecimalMetrics     = "StatSv1.GetQueueDecimalMetrics"
	StatSv1Ping                       = "StatSv1.Ping"
	StatSv1GetStatQueuesForEvent      = "StatSv1.GetStatQueuesForEvent"
	StatSv1GetStatQueue               = "StatSv1.GetStatQueue"
	StatSv1ResetStatQueue             = "StatSv1.ResetStatQueue"
	AdminSv1GetStatQueueProfile       = "AdminSv1.GetStatQueueProfile"
	AdminSv1RemoveStatQueueProfile    = "AdminSv1.RemoveStatQueueProfile"
	AdminSv1SetStatQueueProfile       = "AdminSv1.SetStatQueueProfile"
	AdminSv1GetStatQueueProfiles      = "AdminSv1.GetStatQueueProfiles"
	AdminSv1GetStatQueueProfileIDs    = "AdminSv1.GetStatQueueProfileIDs"
	AdminSv1GetStatQueueProfilesCount = "AdminSv1.GetStatQueueProfilesCount"
)

// RankingS APIs
const (
	AdminSv1GetRankingProfile       = "AdminSv1.GetRankingProfile"
	AdminSv1RemoveRankingProfile    = "AdminSv1.RemoveRankingProfile"
	AdminSv1SetRankingProfile       = "AdminSv1.SetRankingProfile"
	AdminSv1GetRankingProfiles      = "AdminSv1.GetRankingProfiles"
	AdminSv1GetRankingProfileIDs    = "AdminSv1.GetRankingProfileIDs"
	AdminSv1GetRankingProfilesCount = "AdminSv1.GetRankingProfilesCount"
	RankingSv1Ping                  = "RankingSv1.Ping"
	RankingSv1GetRanking            = "RankingSv1.GetRanking"
	RankingSv1GetSchedule           = "RankingSv1.GetSchedule"
	RankingSv1ScheduleQueries       = "RankingSv1.ScheduleQueries"
	RankingSv1GetRankingSummary     = "RankingSv1.GetRankingSummary"
)

// TrendS APIs
const (
	AdminSv1GetTrendProfile       = "AdminSv1.GetTrendProfile"
	AdminSv1RemoveTrendProfile    = "AdminSv1.RemoveTrendProfile"
	AdminSv1SetTrendProfile       = "AdminSv1.SetTrendProfile"
	AdminSv1GetTrendProfiles      = "AdminSv1.GetTrendProfiles"
	AdminSv1GetTrendProfileIDs    = "AdminSv1.GetTrendProfileIDs"
	AdminSv1GetTrendProfilesCount = "AdminSv1.GetTrendProfilesCount"
	TrendSv1Ping                  = "TrendSv1.Ping"
	TrendSv1ScheduleQueries       = "TrendSv1.ScheduleQueries"
	TrendSv1GetTrend              = "TrendSv1.GetTrend"
	TrendSv1GetScheduledTrends    = "TrendSv1.GetScheduledTrends"
	TrendSv1GetTrendSummary       = "TrendSv1.GetTrendSummary"
)

// ResourceS APIs
const (
	ResourceSv1Ping                  = "ResourceSv1.Ping"
	ResourceSv1GetResource           = "ResourceSv1.GetResource"
	ResourceSv1GetResourceWithConfig = "ResourceSv1.GetResourceWithConfig"
	ResourceSv1GetResourcesForEvent  = "ResourceSv1.GetResourcesForEvent"
	ResourceSv1AuthorizeResources    = "ResourceSv1.AuthorizeResources"
	ResourceSv1AllocateResources     = "ResourceSv1.AllocateResources"
	ResourceSv1ReleaseResources      = "ResourceSv1.ReleaseResources"
	AdminSv1SetResourceProfile       = "AdminSv1.SetResourceProfile"
	AdminSv1GetResourceProfiles      = "AdminSv1.GetResourceProfiles"
	AdminSv1RemoveResourceProfile    = "AdminSv1.RemoveResourceProfile"
	AdminSv1GetResourceProfile       = "AdminSv1.GetResourceProfile"
	AdminSv1GetResourceProfileIDs    = "AdminSv1.GetResourceProfileIDs"
	AdminSv1GetResourceProfilesCount = "AdminSv1.GetResourceProfilesCount"
)

// IPs APIs
const (
	IPsV1Ping                    = "IPsV1.Ping"
	IPsV1GetIPAllocations        = "IPsV1.GetIPAllocations"
	IPsV1GetIPAllocationForEvent = "IPsV1.GetIPAllocationForEvent"
	IPsV1AuthorizeIP             = "IPsV1.AuthorizeIP"
	IPsV1AllocateIP              = "IPsV1.AllocateIP"
	IPsV1ReleaseIP               = "IPsV1.ReleaseIP"
	AdminSv1SetIPProfile         = "AdminSv1.SetIPProfile"
	AdminSv1GetIPProfiles        = "AdminSv1.GetIPProfiles"
	AdminSv1RemoveIPProfile      = "AdminSv1.RemoveIPProfile"
	AdminSv1GetIPProfile         = "AdminSv1.GetIPProfile"
	AdminSv1GetIPProfileIDs      = "AdminSv1.GetIPProfileIDs"
	AdminSv1GetIPProfilesCount   = "AdminSv1.GetIPProfilesCount"
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
	DispatcherSv1                    = "DispatcherSv1"
	DispatcherSv1Ping                = "DispatcherSv1.Ping"
	DispatcherSv1GetProfilesForEvent = "DispatcherSv1.GetProfilesForEvent"
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

// RateProfile APIs
const (
	AdminSv1SetRateProfile           = "AdminSv1.SetRateProfile"
	AdminSv1GetRateProfile           = "AdminSv1.GetRateProfile"
	AdminSv1GetRateProfiles          = "AdminSv1.GetRateProfiles"
	AdminSv1GetRateProfileRates      = "AdminSv1.GetRateProfileRates"
	AdminSv1GetRateProfileIDs        = "AdminSv1.GetRateProfileIDs"
	AdminSv1GetRateProfilesCount     = "AdminSv1.GetRateProfilesCount"
	AdminSv1GetRateProfileRatesCount = "AdminSv1.GetRateProfileRatesCount"
	AdminSv1GetRateProfileRateIDs    = "AdminSv1.GetRateProfileRateIDs"
	AdminSv1SetRateProfileRates      = "AdminSv1.SetRateProfileRates"
	AdminSv1RemoveRateProfile        = "AdminSv1.RemoveRateProfile"
	AdminSv1RemoveRateProfileRates   = "AdminSv1.RemoveRateProfileRates"
)

// AnalyzerS APIs
const (
	AnalyzerSv1            = "AnalyzerSv1"
	AnalyzerSv1Ping        = "AnalyzerSv1.Ping"
	AnalyzerSv1StringQuery = "AnalyzerSv1.StringQuery"
)

// LoaderS APIs
const (
	LoaderSv1          = "LoaderSv1"
	LoaderSv1Run       = "LoaderSv1.Run"
	LoaderSv1Ping      = "LoaderSv1.Ping"
	LoaderSv1ImportZip = "LoaderSv1.ImportZip"
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
	CDRsV1                    = "CDRsV1"
	CDRsV1GetCDRsCount        = "CDRsV1.GetCDRsCount"
	CDRsV1RateCDRs            = "CDRsV1.RateCDRs"
	CDRsV1GetCDRs             = "CDRsV1.GetCDRs"
	CDRsV1ProcessCDR          = "CDRsV1.ProcessCDR"
	CDRsV1ProcessExternalCDR  = "CDRsV1.ProcessExternalCDR"
	CDRsV1StoreSessionCost    = "CDRsV1.StoreSessionCost"
	CDRsV1ProcessEvent        = "CDRsV1.ProcessEvent"
	CDRsV1ProcessEventWithGet = "CDRsV1.ProcessEventWithGet"
	CDRsV1ProcessStoredEvents = "CDRsV1.ProcessStoredEvents"
	CDRsV1Ping                = "CDRsV1.Ping"
	CDRsV2                    = "CDRsV2"
	CDRsV2StoreSessionCost    = "CDRsV2.StoreSessionCost"
	CDRsV2ProcessEvent        = "CDRsV2.ProcessEvent"
)

// EEs
const (
	EeS                       = "EeS"
	EeSv1                     = "EeSv1"
	EeSv1Ping                 = "EeSv1.Ping"
	EeSv1ProcessEvent         = "EeSv1.ProcessEvent"
	EeSv1ArchiveEventsInReply = "EeSv1.ArchiveEventsInReply"
)

// ActionProfile APIs
const (
	AdminSv1SetActionProfile       = "AdminSv1.SetActionProfile"
	AdminSv1GetActionProfile       = "AdminSv1.GetActionProfile"
	AdminSv1GetActionProfiles      = "AdminSv1.GetActionProfiles"
	AdminSv1GetActionProfileIDs    = "AdminSv1.GetActionProfileIDs"
	AdminSv1GetActionProfilesCount = "AdminSv1.GetActionProfilesCount"
	AdminSv1RemoveActionProfile    = "AdminSv1.RemoveActionProfile"
)

// AdminSv1
const (
	AdminS   = "AdminS"
	AdminSv1 = "AdminSv1"
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
	ResourcesCsv  = "Resources.csv"
	IPsCsv        = "IPs.csv"
	StatsCsv      = "Stats.csv"
	RankingsCsv   = "Rankings.csv"
	TrendsCsv     = "Trends.csv"
	ThresholdsCsv = "Thresholds.csv"
	FiltersCsv    = "Filters.csv"
	RoutesCsv     = "Routes.csv"
	AttributesCsv = "Attributes.csv"
	ChargersCsv   = "Chargers.csv"
	RatesCsv      = "Rates.csv"
	ActionsCsv    = "Actions.csv"
	AccountsCsv   = "Accounts.csv"
)

// Table Name
const (
	TBLTPResources       = "tp_resources"
	TBLTPStats           = "tp_stats"
	TBLTPRankings        = "tp_rankings"
	TBLTPTrends          = "tp_trends"
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
	CacheConfig                      = "*config"
	CacheResources                   = "*resources"
	CacheResourceProfiles            = "*resource_profiles"
	CacheEventResources              = "*event_resources"
	CacheIPAllocations               = "*ip_allocations"
	CacheIPProfiles                  = "*ip_profiles"
	CacheEventIPs                    = "*event_ips"
	CacheStatQueueProfiles           = "*statqueue_profiles"
	CacheStatQueues                  = "*statqueues"
	CacheRankingProfiles             = "*ranking_profiles"
	CacheRankings                    = "*rankings"
	CacheTrendProfiles               = "*trend_profiles"
	CacheTrends                      = "*trends"
	CacheThresholdProfiles           = "*threshold_profiles"
	CacheThresholds                  = "*thresholds"
	CacheFilters                     = "*filters"
	CacheRouteProfiles               = "*route_profiles"
	CacheAttributeProfiles           = "*attribute_profiles"
	CacheChargerProfiles             = "*charger_profiles"
	CacheRateProfiles                = "*rate_profiles"
	CacheActionProfiles              = "*action_profiles"
	CacheAccounts                    = "*accounts"
	CacheResourceFilterIndexes       = "*resource_filter_indexes"
	CacheIPFilterIndexes             = "*ip_filter_indexes"
	CacheStatFilterIndexes           = "*stat_filter_indexes"
	CacheThresholdFilterIndexes      = "*threshold_filter_indexes"
	CacheRouteFilterIndexes          = "*route_filter_indexes"
	CacheAttributeFilterIndexes      = "*attribute_filter_indexes"
	CacheChargerFilterIndexes        = "*charger_filter_indexes"
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
	CacheCDRsTBL = "*cdrs"
)

// Prefix for indexing
const (
	ResourceFilterIndexes         = "rfi_"
	IPFilterIndexes               = "ifi_"
	StatFilterIndexes             = "sfi_"
	ThresholdFilterIndexes        = "tfi_"
	AttributeFilterIndexes        = "afi_"
	ChargerFilterIndexes          = "cfi_"
	DispatcherFilterIndexes       = "dfi_"
	ActionPlanIndexes             = "api_"
	RouteFilterIndexes            = "rti_"
	RateProfilesFilterIndexPrfx   = "rpi_"
	RateFilterIndexPrfx           = "rri_"
	RankingPrefix                 = "rnk_"
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
	JanusAgent      = "JanusAgent"
	PrometheusAgent = "PrometheusAgent"
)

// Google_API
const (
	MetaGoogleAPI             = "*gapi"
	GoogleCredentialsFileName = "credentials.json"
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
	MaxParallelConnsCfg     = "max_parallel_conns"
	EEsConnsCfg             = "ees_conns"
	DecimalMaxScaleCfg      = "decimal_max_scale"
	DecimalMinScaleCfg      = "decimal_min_scale"
	DecimalPrecisionCfg     = "decimal_precision"
	DecimalRoundingModeCfg  = "decimal_rounding_mode"
)

const (
	LevelCfg         = "level"
	KafkaConnCfg     = "kafka_conn"
	KafkaTopicCfg    = "kafka_topic"
	KafkaAttemptsCfg = "kafka_attempts"
)

const (
	TypeCfg                   = "type"
	SQLMaxOpenConnsCfg        = "sqlMaxOpenConns"
	SQLMaxIdleConnsCfg        = "sqlMaxIdleConns"
	SQLLogLevelCfg            = "sqlLogLevel"
	SQLConnMaxLifetimeCfg     = "sqlConnMaxLifetime"
	StringIndexedFieldsCfg    = "string_indexed_fields"
	PrefixIndexedFieldsCfg    = "prefix_indexed_fields"
	SuffixIndexedFieldsCfg    = "suffix_indexed_fields"
	ExistsIndexedFieldsCfg    = "exists_indexed_fields"
	NotExistsIndexedFieldsCfg = "notexists_indexed_fields"
	MongoQueryTimeoutCfg      = "mongoQueryTimeout"
	MongoConnSchemeCfg        = "mongoConnScheme"
	PgSSLModeCfg              = "pgSSLMode"
	PgSSLCertCfg              = "pgSSLCert"
	PgSSLKeyCfg               = "pgSSLKey"
	PgSSLPasswordCfg          = "pgSSLPassword"
	PgSSLCertModeCfg          = "pgSSLCertMode"
	PgSSLRootCertCfg          = "pgSSLRootCert"
	ItemsCfg                  = "items"
	OptsCfg                   = "opts"
	Tenants                   = "tenants"
	MysqlLocation             = "mysqlLocation"
)

// DataDbCfg
const (
	DataDbTypeCfg                = "db_type"
	DataDbHostCfg                = "db_host"
	DataDbPortCfg                = "db_port"
	DataDbNameCfg                = "db_name"
	DataDbUserCfg                = "db_user"
	DataDbPassCfg                = "db_password"
	InternalDBDumpPathCfg        = "internalDBDumpPath"
	InternalDBBackupPathCfg      = "internalDBBackupPath"
	InternalDBStartTimeoutCfg    = "internalDBStartTimeout"
	InternalDBDumpIntervalCfg    = "internalDBDumpInterval"
	InternalDBRewriteIntervalCfg = "internalDBRewriteInterval"
	InternalDBFileSizeLimitCfg   = "internalDBFileSizeLimit"
	RedisMaxConnsCfg             = "redisMaxConns"
	RedisConnectAttemptsCfg      = "redisConnectAttempts"
	RedisSentinelNameCfg         = "redisSentinel"
	RedisClusterCfg              = "redisCluster"
	RedisClusterSyncCfg          = "redisClusterSync"
	RedisClusterOnDownDelayCfg   = "redisClusterOndownDelay"
	RedisConnectTimeoutCfg       = "redisConnectTimeout"
	RedisReadTimeoutCfg          = "redisReadTimeout"
	RedisWriteTimeoutCfg         = "redisWriteTimeout"
	RedisPoolPipelineWindowCfg   = "redisPoolPipelineWindow"
	RedisPoolPipelineLimitCfg    = "redisPoolPipelineLimit"
	RedisTLSCfg                  = "redisTLS"
	RedisClientCertificateCfg    = "redisClientCertificate"
	RedisClientKeyCfg            = "redisClientKey"
	RedisCACertificateCfg        = "redisCACertificate"
	ReplicationFilteredCfg       = "replication_filtered"
	ReplicationCache             = "replication_cache"
	RemoteConnIDCfg              = "remote_conn_id"
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
	PprofPathCfg             = "pprof_path"
	HTTPUseBasicAuthCfg      = "use_basic_auth"
	HTTPAuthUsersCfg         = "auth_users"
	HTTPClientOptsCfg        = "client_opts"

	HTTPClientSkipTLSVerificationCfg   = "skipTLSVerification"
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
	TrendSConnsCfg    = "trends_conns"
	RankingSConnsCfg  = "rankings_conns"
)

const (
	EnabledCfg         = "enabled"
	ThresholdSConnsCfg = "thresholds_conns"
	CacheSConnsCfg     = "caches_conns"
	ScheduledIDsCfg    = "scheduled_ids"
)

// Efs
const (
	EFsConnsCfg = "efs_conns"
)

// CdrsCfg
const (
	CDRsConnsCfg           = "cdrs_conns"
	FiltersCfg             = "filters"
	ExtraFieldsCfg         = "extra_fields"
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
	IPsConnsCfg            = "ips_conns"
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
	SessionSConnsCfg          = "sessions_conns"
	SubscribeParkCfg          = "subscribe_park"
	CreateCdrCfg              = "create_cdr"
	LowBalanceAnnFileCfg      = "low_balance_ann_file"
	EmptyBalanceContextCfg    = "empty_balance_context"
	EmptyBalanceAnnFileCfg    = "empty_balance_ann_file"
	MaxWaitConnectionCfg      = "max_wait_connection"
	ActiveSessionDelimiterCfg = "active_session_delimiter"
	EventSocketConnsCfg       = "event_socket_conns"
	EmptyBalanceContext       = "empty_balance_context"
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

	// PrometheusAgentCfg
	CoreSConnsCfg            = "cores_conns"
	CollectGoMetricsCfg      = "collect_go_metrics"
	CollectProcessMetricsCfg = "collect_process_metrics"
	StatQueueIDsCfg          = "stat_queue_ids"

	// RequestProcessor
	RequestFieldsCfg = "request_fields"
	ReplyFieldsCfg   = "reply_fields"

	// RadiusAgentCfg
	ListenAuthCfg         = "listen_auth"
	ListenAcctCfg         = "listen_acct"
	ClientSecretsCfg      = "client_secrets"
	ClientDictionariesCfg = "client_dictionaries"

	// JanusAgentCfg
	JanusConnsCfg    = "janus_conns"
	AdminAddressCfg  = "admin_address"
	AdminPasswordCfg = "admin_password"

	// AttributeSCfg
	IndexedSelectsCfg  = "indexed_selects"
	ProfileRunsCfg     = "profile_runs"
	NestedFieldsCfg    = "nested_fields"
	MetaProcessRunsCfg = "*processRuns"
	MetaProfileRunsCfg = "*profileRuns"

	// ChargerSCfg
	StoreIntervalCfg = "store_interval"

	// StatSCfg
	StoreUncompressedLimitCfg = "store_uncompressed_limit"
	EEsExporterIDsCfg         = "ees_exporter_ids"

	// Cache
	PartitionsCfg = "partitions"
	PrecacheCfg   = "precache"

	// CdrsCfg
	ExportPathCfg         = "export_path"
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

	// RouteSCfg
	MetaProfileCountCfg = "*profileCount"
	MetaIgnoreErrorsCfg = "*ignoreErrors"
	MetaMaxCostCfg      = "*maxCost"
	MetaLimitCfg        = "*limit"
	MetaOffsetCfg       = "*offset"
	MetaMaxItemsCfg     = "*maxItems"

	// RateSCfg
	MetaIntervalStartCfg          = "*intervalStart"
	RateIndexedSelectsCfg         = "rate_indexed_selects"
	RateNestedFieldsCfg           = "rate_nested_fields"
	RateStringIndexedFieldsCfg    = "rate_string_indexed_fields"
	RatePrefixIndexedFieldsCfg    = "rate_prefix_indexed_fields"
	RateSuffixIndexedFieldsCfg    = "rate_suffix_indexed_fields"
	RateExistsIndexedFieldsCfg    = "rate_exists_indexed_fields"
	RateNotExistsIndexedFieldsCfg = "rate_notexists_indexed_fields"
	Verbosity                     = "verbosity"

	// ResourceSCfg
	MetaUsageIDCfg  = "*usageID"
	MetaUsageTTLCfg = "*usageTTL"
	MetaUnitsCfg    = "*units"

	// SessionsCfg
	MetaAttributesDerivedReplyCfg = "*attributesDerivedReply"
	MetaBlockerErrorCfg           = "*blockerError"
	MetaCDRsDerivedReplyCfg       = "*cdrsDerivedReply"
	MetaResourcesAuthorizeCfg     = "*resourcesAuthorize"
	MetaResourcesAllocateCfg      = "*resourcesAllocate"
	MetaResourcesReleaseCfg       = "*resourcesRelease"
	MetaResourcesDerivedReplyCfg  = "*resourcesDerivedReply"
	MetaIPsAuthorizeCfg           = "*ipsAuthorize"
	MetaIPsAllocateCfg            = "*ipsAllocate"
	MetaIPsReleaseCfg             = "*ipsRelease"
	MetaRoutesDerivedReplyCfg     = "*routesDerivedReply"
	MetaStatsDerivedReplyCfg      = "*statsDerivedReply"
	MetaThresholdsDerivedReplyCfg = "*thresholdsDerivedReply"
	MetaMaxUsageCfg               = "*maxUsage"
	MetaForceUsageCfg             = "*forceUsage"
	MetaTTLCfg                    = "*ttl"
	MetaChargeableCfg             = "*chargeable"
	MetaDebitIntervalCfg          = "*debitInterval"
	MetaTTLLastUsageCfg           = "*ttlLastUsage"
	MetaTTLLastUsedCfg            = "*ttlLastUsed"
	MetaTTLMaxDelayCfg            = "*ttlMaxDelay"
	MetaTTLUsageCfg               = "*ttlUsage"
	MetaAccountsForceUsage        = "*accountsForceUsage"

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
	FieldSepCfg            = "field_separator"
	RunDelayCfg            = "run_delay"
	StartDelayCfg          = "start_delay"
	SourcePathCfg          = "source_path"
	ProcessedPathCfg       = "processed_path"
	TenantCfg              = "tenant"
	EEsSuccessIDsCfg       = "ees_success_ids"
	EEsFailedIDsCfg        = "ees_failed_ids"
	FlagsCfg               = "flags"
	FieldsCfg              = "fields"
	CacheDumpFieldsCfg     = "cache_dump_fields"
	PartialCommitFieldsCfg = "partial_commit_fields"
	PartialCacheTTLCfg     = "partial_cache_ttl"
	ActionCfg              = "action"
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

const (
	ClientIDCfg     = "client_id"
	ClientSecretCfg = "client_secret"
	TokenUrlCfg     = "token_url"
	IpsUrlCfg       = "ips_url"
	NumbersUrlCfg   = "numbers_url"
	AudienceCfg     = "audience"
	GrantTypeCfg    = "grant_type"
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
var CGROptionsSet = NewStringSet([]string{OptsRatesProfileIDs, OptsRatesStartTime, OptsRatesUsage, OptsSesTTL,
	OptsSesTTLMaxDelay, OptsSesTTLLastUsed, OptsSesTTLLastUsage, OptsSesTTLUsage,
	OptsSesDebitInterval, OptsStirATest, OptsStirPayloadMaxDuration, OptsStirIdentity,
	OptsStirOriginatorTn, OptsStirOriginatorURI, OptsStirDestinationTn, OptsStirDestinationURI,
	OptsStirPublicKeyPath, OptsStirPrivateKeyPath, OptsAPIKey, OptsRouteID, OptsContext, OptsAttributesProfileIDs,
	OptsAttributesProcessRuns, OptsAttributesProfileRuns, OptsRoutesLimit, OptsRoutesOffset, OptsRoutesMaxItems,
	OptsSesChargeable, RemoteHostOpt, MetaCache, OptsThresholdsProfileIDs, OptsRoutesProfilesCount,
	OptsSesAttributeSDerivedReply, OptsSesBlockerError, OptsRoutesUsage,
	MetaCDRs, OptsSesCDRsDerivedReply, MetaResources, MetaIPs, OptsSesResourceSAuthorize,
	OptsSesResourceSAllocate, OptsSesResourceSRelease, OptsSesResourceSDerivedReply, MetaRoutes,
	OptsSesRouteSDerivedReply, OptsSesStatSDerivedReply, OptsSesSTIRAuthenticate, OptsSesSTIRDerivedReply,
	OptsSesSTIRInitiate, OptsSesThresholdSDerivedReply,
	OptsSesMaxUsage, OptsSesForceUsage, OptsSesInitiate, OptsSesUpdate, OptsSesTerminate,
	OptsSesMessage, MetaAttributes, MetaChargers, OptsCDRsExport, OptsCDRsRefund,
	OptsCDRsRerate, MetaStats, OptsCDRsStore, MetaThresholds, MetaRates, MetaAccounts,
	OptsAccountsUsage, OptsStatsProfileIDs, OptsActionsProfileIDs, MetaProfileIgnoreFilters,
	OptsRoundingDecimals})

// Event Opts
const (
	// Subsystems boolean opts

	// SessionS
	OptsSesTTL           = "*sesTTL"
	OptsSesChargeable    = "*sesChargeable"
	OptsSesDebitInterval = "*sesDebitInterval"
	OptsSesTTLLastUsage  = "*sesTTLLastUsage"
	OptsSesTTLLastUsed   = "*sesTTLLastUsed"
	OptsSesTTLMaxDelay   = "*sesTTLMaxDelay"
	OptsSesTTLUsage      = "*sesTTLUsage"
	OptsSesForceUsage    = "*sesForceUsage"

	OptsSesAttributeSDerivedReply = "*sesAttributeSDerivedReply"
	OptsSesBlockerError           = "*sesBlockerError"
	OptsSesCDRsDerivedReply       = "*sesCDRsDerivedReply"
	OptsSesResourceSAuthorize     = "*sesResourceSAuthorize"
	OptsSesResourceSAllocate      = "*sesResourceSAllocate"
	OptsSesResourceSRelease       = "*sesResourceSRelease"
	OptsSesResourceSDerivedReply  = "*sesResourceSDerivedReply"
	OptsSesRouteSDerivedReply     = "*sesRouteSDerivedReply"
	OptsSesStatSDerivedReply      = "*sesStatSDerivedReply"
	OptsSesSTIRAuthenticate       = "*sesSTIRAuthenticate"
	OptsSesSTIRDerivedReply       = "*sesSTIRDerivedReply"
	OptsSesSTIRInitiate           = "*sesSTIRInitiate"
	OptsSesThresholdSDerivedReply = "*sesThresholdSDerivedReply"
	OptsSesMaxUsage               = "*sesMaxUsage"
	OptsSesInitiate               = "*sesInitiate"
	OptsSesUpdate                 = "*sesUpdate"
	OptsSesTerminate              = "*sesTerminate"
	OptsSesMessage                = "*sesMessage"

	// Accounts
	OptsAccountsUsage      = "*acntUsage"
	OptsAccountsForceUsage = "*accountSForceUsage"
	OptsAccountsProfileIDs = "*acntProfileIDs"

	// Actions
	OptsActionsProfileIDs = "*actProfileIDs"

	// Attributes
	OptsAttributesProfileIDs  = "*attrProfileIDs"
	OptsAttributesProfileRuns = "*attrProfileRuns"
	OptsAttributesProcessRuns = "*attrProcessRuns"

	// CDRs
	OptsCDRsExport = "*cdrsExport"
	OptsCDRsRefund = "*cdrsRefund"
	OptsCDRsRerate = "*cdrsRerate"
	OptsCDRsStore  = "*cdrsStore"

	// DispatcherS
	OptsAPIKey                   = "*apiKey"
	OptsRouteID                  = "*routeID"
	OptsDispatchersProfilesCount = "*dispatchersProfilesCount"

	// EEs
	OptsEEsVerbose = "*eesVerbose"

	// Rates
	OptsRatesProfileIDs    = "*rtsProfileIDs"
	OptsRatesStartTime     = "*rtsStartTime"
	OptsRatesUsage         = "*rtsUsage"
	OptsRatesIntervalStart = "*rtsIntervalStart"

	// Resources
	OptsResourcesUnits    = "*rsUnits"
	OptsResourcesUsageID  = "*rsUsageID"
	OptsResourcesUsageTTL = "*rsUsageTTL"

	// IPs
	OptsIPsAllocationID = "*ipAllocationID"
	OptsIPsTTL          = "*ipTTL"
	MetaAllocationID    = "*allocationID"

	// Routes
	OptsRoutesProfilesCount = "*rouProfilesCount"
	OptsRoutesLimit         = "*rouLimit"
	OptsRoutesOffset        = "*rouOffset"
	OptsRoutesMaxItems      = "*rouMaxItems"
	OptsRoutesIgnoreErrors  = "*rouIgnoreErrors"
	OptsRoutesMaxCost       = "*rouMaxCost"
	OptsRoutesUsage         = "*rouUsage"

	// Stats
	OptsStatsProfileIDs  = "*statsProfileIDs"
	OptsRoundingDecimals = "*roundingDecimals"

	// Thresholds
	OptsThresholdsProfileIDs = "*thdProfileIDs"

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

	// Others
	OptsContext              = "*context"
	MetaSubsys               = "*subsys"
	MetaMethod               = "*reqMethod"
	MetaEventType            = "*eventType"
	EventType                = "EventType"
	SchedulerInit            = "SchedulerInit"
	MetaProfileIgnoreFilters = "*profileIgnoreFilters"
	MetaPosterAttempts       = "*posterAttempts"

	RemoteHostOpt = "*rmtHost"
	MetaCache     = "*cache"

	MetaWithIndex   = "*withIndex"
	MetaForceLock   = "*forceLock"
	MetaStopOnError = "*stopOnError"
)

// Event Flags
const (
	MetaDerivedReply = "*derived_reply"

	MetaIDs        = "*IDs"
	MetaProfileIDs = "*profileIDs"

	TrueStr  = "true"
	FalseStr = "false"
)

// ArgCache constats
const (
	ThresholdIDs     = "ThresholdIDs"
	FilterIDs        = "FilterIDs"
	RateProfileIDs   = "RateProfileIDs"
	ActionProfileIDs = "ActionProfileIDs"
)

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
	SQLDefaultDBName    = "cgrates"
	SQLDefaultPgSSLMode = "disable"

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
	PartialOpt      = "*partial"
	PartialRatesOpt = "*partial_rates"

	PartialOrderFieldOpt       = "partialOrderField"
	PartialCacheActionOpt      = "partialCacheAction"
	PartialPathOpt             = "partialPath"
	PartialCSVFieldSepartorOpt = "partialcsvFieldSeparator"

	// EEs Elasticsearch options
	ElsIndex                    = "elsIndex"
	ElsRefresh                  = "elsRefresh"
	ElsOpType                   = "elsOpType"
	ElsPipeline                 = "elsPipeline"
	ElsRouting                  = "elsRouting"
	ElsTimeout                  = "elsTimeout"
	ElsWaitForActiveShards      = "elsWaitForActiveShards"
	ElsCAPath                   = "elsCAPath"
	ElsDiscoverNodesOnStart     = "elsDiscoverNodesOnStart"
	ElsDiscoverNodeInterval     = "elsDiscoverNodeInterval"
	ElsCloud                    = "elsCloud"
	ElsAPIKey                   = "elsAPIKey"
	ElsCertificateFingerprint   = "elsCertificateFingerprint"
	ElsServiceToken             = "elsServiceToken"
	ElsUsername                 = "elsUsername"
	ElsPassword                 = "elsPassword"
	ElsEnableDebugLogger        = "elsEnableDebugLogger"
	ElsLogger                   = "elsLogger"
	ElsCompressRequestBody      = "elsCompressRequestBody"
	ElsCompressRequestBodyLevel = "elsCompressRequestBodyLevel"
	ElsRetryOnStatus            = "elsRetryOnStatus"
	ElsMaxRetries               = "elsMaxRetries"
	ElsDisableRetry             = "elsDisableRetry"

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

// Paginator options
const (
	PageLimitOpt    = "*pageLimit"
	PageOffsetOpt   = "*pageOffset"
	PageMaxItemsOpt = "*pageMaxItems"
	ItemsPrefixOpt  = "*itemsPrefix"
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
	CheckCfgCgr          = "check_config"
	PidCgr               = "pid"
	CpuProfDirCgr        = "cpuprof_dir"
	MemProfDirCgr        = "memprof_dir"
	MemProfIntervalCgr   = "memprof_interval"
	MemProfMaxFilesCgr   = "memprof_maxfiles"
	MemProfTimestampCgr  = "memprof_timestamp"
	ScheduledShutdownCgr = "scheduled_shutdown"
	SingleCpuCgr         = "single_cpu"
	PreloadCgr           = "preload"
	SetVersionsCgr       = "set_versions"
	MemProfFinalFile     = "mem_final.prof"
	CpuPathCgr           = "cpu.prof"
	//Cgr loader
	CgrLoader         = "cgr-loader"
	CachingArgCgr     = "caching"
	FieldSepCgr       = "field_sep"
	ImportIDCgr       = "import_id"
	DisableReverseCgr = "disable_reverse_mappings"
	RemoveCgr         = "remove"
	CacheSAddress     = "caches_address"
	SchedulerAddress  = "scheduler_address"
	//Cgr migrator
	CgrMigrator = "cgr-migrator"
	ExecCgr     = "exec"
)

var AnzIndexType = StringSet{ // AnzIndexType are the analyzers possible index types
	MetaScorch:   {},
	MetaBoltdb:   {},
	MetaLeveldb:  {},
	MetaMoss:     {},
	MetaInternal: {},
}

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

// rounding strings
const (
	ToNearestEven       = "*toNearestEven"
	ToNearestAway       = "*toNearestAway"
	ToZero              = "*toZero"
	AwayFromZero        = "*awayFromZero"
	ToNegativeInf       = "*toNegativeInf"
	ToPositiveInf       = "*toPositiveInf"
	ToNearestTowardZero = "*toNearestTowardZero"
)

const (
	StateServiceUP   = "SERVICE_UP"
	StateServiceDOWN = "SERVICE_DOWN"
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
