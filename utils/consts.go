/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
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
		CacheDiameterMessages, CacheRadiusPackets, CacheRPCResponses, CacheClosedSessions,
		CacheCDRIDs, CacheRPCConnections, CacheUCH, CacheSTIR, CacheEventCharges, MetaAPIBan, MetaSentryPeer,
		CacheCapsEvents, CacheReplicationHosts})

	// DBPartitions excluding Resources, Thresholds, Trends, Rankings, IPs, Stats
	StatelessDBPartitions = NewStringSet([]string{
		CacheFilters, CacheRouteProfiles, CacheAttributeProfiles,
		CacheChargerProfiles, CacheActionProfiles, CacheRouteFilterIndexes,
		CacheAttributeFilterIndexes, CacheChargerFilterIndexes, CacheLoadIDs,
		CacheRateProfiles, CacheRateProfilesFilterIndexes,
		CacheRateFilterIndexes, CacheActionProfilesFilterIndexes,
		CacheAccountsFilterIndexes, CacheReverseFilterIndexes, CacheAccounts,
	})

	DBPartitions = NewStringSet([]string{
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
	CachePartitions = JoinStringSet(extraDBPartition, DBPartitions)

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
		CacheRankingProfiles:             RankingProfilePrefix,
		CacheRankings:                    RankingPrefix,
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
	AtChar                   = "@"
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
	MetaCGRid                = "*cgrID"
	CGRidCharSize            = 40
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

	CustomValue       = "CustomValue"
	Value             = "Value"
	Rules             = "Rules"
	Metrics           = "Metrics"
	RunTimes          = "RunTimes"
	CompressedMetrics = "CompressedMetrics"
	TrendGrowth       = "TrendGrowth"
	TrendLabel        = "TrendLabel"
	MetricID          = "MetricID"
	LastUsed          = "LastUsed"
	PDD               = "PDD"
	RouteStr          = "Route"
	RunID             = "RunID"
	MetaRunID         = "*runID"

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
	Mongo                   = "mongo"
	MetaRedis               = "*redis"
	Redis                   = "redis"
	MetaPostgres            = "*postgres"
	MetaInternal            = "*internal"
	Internal                = "internal"
	MetaLocalHost           = "*localhost"
	MetaBiJSONLocalHost     = "*bijsonLocalhost"
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
	CDRsPrefix                = "cdr_"
	CDRsIndexes               = "cdi_"
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
	CreateDBTablesSQL         = "create_db_tables.sql"
	CreateTariffPlanTablesSQL = "create_tariffplan_tables.sql"

	MetaAsc             = "*asc"
	MetaDesc            = "*desc"
	MetaAscending       = "*ascending"
	MetaDescending      = "*descending"
	MetaConstant        = "*constant"
	MetaPositive        = "*positive"
	MetaNegative        = "*negative"
	MetaLast            = "*last"
	MetaPassword        = "*password"
	MetaFiller          = "*filler"
	MetaHTTPPost        = "*httpPost"
	JanusAdminSubProto  = "janus-admin-protocol"
	MetaHTTPjsonMap     = "*httpJSONMap"
	MetaAMQPjsonMap     = "*amqpJSONMap"
	MetaAMQPV1jsonMap   = "*amqpv1JSONMap"
	MetaSQSjsonMap      = "*sqsJSONMap"
	MetaKafkajsonMap    = "*kafkaJSONMap"
	MetaNATSJSONMap     = "*natsJSONMap"
	MetaSQL             = "*sql"
	MetaCgrcdr          = "*cgrcdr"
	MetaMySQL           = "*mysql"
	MetaS3jsonMap       = "*s3JSONMap"
	ConfigPath          = "/etc/cgrates/"
	DisconnectCause     = "DisconnectCause"
	MetaRating          = "*rating"
	MetaAccounting      = "*accounting"
	NotAvailable        = "N/A"
	Call                = "call"
	ExtraFields         = "ExtraFields"
	MetaDynamic         = "*dynamic"
	MetaCounterEvent    = "*event"
	MetaBalance         = "*balance"
	MetaAccount         = "*account"
	EventName           = "EventName"
	HierarchySep        = ">"
	MetaComposed        = "*composed"
	MetaUsageDifference = "*usageDifference"
	MetaDifference      = "*difference"
	MetaVariable        = "*variable"
	MetaCCUsage         = "*ccUsage"
	MetaSIPCID          = "*sipcid"
	MetaValueExponent   = "*valueExponent"
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
	MetaMaskedDestination    = "*maskedDestination"
	MetaUnixTimestamp        = "*unixTimestamp"
	MetaPostCDR              = "*postCDR"
	MetaDumpToFile           = "*dumpToFile"
	MetaDumpToJSON           = "*dumpToJSON"
	NonTransactional         = ""
	DB                       = "db"
	StorDB                   = "StorDB"
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
	MetaReplication          = "*replication"
	MetaRerate               = "*rerate"
	MetaRefund               = "*refund"
	MetaStats                = "*stats"
	MetaTrends               = "*trends"
	MetaRankings             = "*rankings"
	MetaCores                = "*cores"
	MetaServiceManager       = "*servicemanager"
	MetaChargers             = "*chargers"
	MetaConfig               = "*config"
	MetaTpes                 = "*tpes"
	MetaFilters              = "*filters"
	MetaCDRs                 = "*cdrs"
	MetaEM                   = "*em"
	MetaCaches               = "*caches"
	MetaUCH                  = "*uch"
	MetaGuardian             = "*guardians"
	MetaEEs                  = "*ees"
	MetaEEsIDs               = "*eesIDs"
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
	RatingPlans              = "RatingPlans"
	RatingProfiles           = "RatingProfiles"
	AccountActions           = "AccountActions"
	ResourcesStr             = "Resources"
	Stats                    = "Stats"
	Rankings                 = "Rankings"
	Trends                   = "Trends"
	Filters                  = "Filters"
	RateProfiles             = "RateProfiles"
	ActionProfiles           = "ActionProfiles"
	AccountsString           = "Accounts"

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

	Thresholds            = "Thresholds"
	Routes                = "Routes"
	Attributes            = "Attributes"
	Chargers              = "Chargers"
	StatS                 = "StatS"
	LoadIDsVrs            = "LoadIDs"
	GlobalVarS            = "GlobalVarS"
	CostSource            = "CostSource"
	ExtraInfo             = "ExtraInfo"
	Meta                  = "*"
	MetaSysLog            = "*syslog"
	MetaStdLog            = "*stdout"
	MetaKafkaLog          = "*kafkaLog"
	Kafka                 = "Kafka"
	EventSource           = "EventSource"
	AccountID             = "AccountID"
	AccountIDs            = "AccountIDs"
	ResourceID            = "ResourceID"
	TotalUsage            = "TotalUsage"
	StatID                = "StatID"
	BalanceType           = "BalanceType"
	BalanceID             = "BalanceID"
	BalanceCostIncrements = "BalanceCostIncrements"
	BalanceAttributeIDs   = "BalanceAttributeIDs"
	BalanceRateProfileIDs = "BalanceRateProfileIDs"
	BalanceWeights        = "BalanceWeights"
	BalanceBlockers       = "BalanceBlockers"
	BalanceDisabled       = "BalanceDisabled"
	Units                 = "Units"
	CDRs                  = "CDRs"
	UsageRecord           = "UsageRecord"
	MetaUsageRecord       = "*usageRecord"
	MetaUsageRecordID     = "*usageRecordID"
	ExpiryTime            = "ExpiryTime"
	EventID               = "EventID"
	AllowNegative         = "AllowNegative"
	Disabled              = "Disabled"
	Action                = "Action"

	SessionSCosts      = "SessionSCosts"
	RQF                = "RQF"
	ResourceStr        = "Resource"
	User               = "User"
	Subscribers        = "Subscribers"
	MetaSubscribers    = "*subscribers"
	MetaDataDB         = "*datadb"
	MetaStorDB         = "*stordb"
	MetaWeight         = "*weight"
	MetaLC             = "*lc"
	MetaHC             = "*hc"
	MetaQOS            = "*qos"
	MetaReas           = "*reas"
	MetaReds           = "*reds"
	Weight             = "Weight"
	Limit              = "Limit"
	UsageTTL           = "UsageTTL"
	Usages             = "Usages"
	TTLIdx             = "TTLIdx"
	AllocationMessage  = "AllocationMessage"
	Pools              = "Pools"
	Allocations        = "Allocations"
	TTLIndex           = "TTLIndex"
	Allocation         = "Allocation"
	Range              = "Range"
	Stored             = "Stored"
	RatingSubject      = "RatingSubject"
	Categories         = "Categories"
	Blocker            = "Blocker"
	Blockers           = "Blockers"
	Params             = "Params"
	StartTime          = "StartTime"
	EndTime            = "EndTime"
	ProcessingTime     = "ProcessingTime"
	ReplyState         = "ReplyState"
	RequestProcessorID = "RequestProcessorID"
	EventReaderID      = "EventReaderID"

	// Event types
	EventType                   = "EventType"
	MetaEventType               = "*eventType"
	CDRKey                      = "CDR"
	ThresholdHit                = "ThresholdHit"
	RankingUpdate               = "RankingUpdate"
	ResourceUpdate              = "ResourceUpdate"
	StatUpdate                  = "StatUpdate"
	TrendUpdate                 = "TrendUpdate"
	EventPerformanceReport      = "PerformanceReport"
	EventConnectionStatusReport = "ConnectionStatusReport"

	// Connection status event fields.
	ConnLocalAddr  = "LocalAddr"
	ConnRemoteAddr = "RemoteAddr"
	ConnStatus     = "ConnectionStatus" // sum metric: UP=1, DOWN=-1
	ConnStatusUp   = "UP"
	ConnStatusDown = "DOWN"

	// ReplyState error constants
	ErrReplyStateAuthorize = "ERR_AUTHORIZE"
	ErrReplyStateInitiate  = "ERR_INITIATE"
	ErrReplyStateUpdate    = "ERR_UPDATE"
	ErrReplyStateTerminate = "ERR_TERMINATE"
	ErrReplyStateMessage   = "ERR_MESSAGE"
	ErrReplyStateEvent     = "ERR_EVENT"
	ErrReplyStateCDRs      = "ERR_CDRS"
	ErrReplyStateExport    = "ERR_EXPORT"
	ErrReplyStateRadauth   = "ERR_RADAUTH"

	AccountSummary     = "AccountSummary"
	Charging           = "Charging"
	Accounting         = "Accounting"
	Rating             = "Rating"
	Charges            = "Charges"
	CompressFactor     = "CompressFactor"
	Increments         = "Increments"
	BalanceField       = "Balance"
	Type               = "Type"
	Element            = "Element"
	Values             = "Values"
	YearsFieldName     = "Years"
	MonthsFieldName    = "Months"
	MonthDaysFieldName = "MonthDays"
	WeekDaysFieldName  = "WeekDays"
	GroupIntervalStart = "GroupIntervalStart"
	RateIncrement      = "RateIncrement"
	RateUnit           = "RateUnit"
	BalanceUUID        = "BalanceUUID"
	ChargingID         = "ChargingID"
	RatingID           = "RatingID"
	JoinedChargeIDs    = "JoinedChargeIDs"
	UnitFactorID       = "UnitFactorID"

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
	MetaSession              = "*session"
	MetaDefault              = "*default"
	MetaPrimary              = "*primary"
	Error                    = "Error"
	MetaCgreq                = "*cgreq"
	MetaCgrep                = "*cgrep"
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
	MetaRound                = "*round"
	Pong                     = "Pong"
	MetaEventCost            = "*eventCost"
	MetaBuffer               = "*buffer"

	Freeswitch             = "freeswitch"
	Kamailio               = "kamailio"
	Opensips               = "opensips"
	Asterisk               = "asterisk"
	SchedulerS             = "SchedulerS"
	MetaMultiply           = "*multiply"
	MetaDivide             = "*divide"
	MetaUrl                = "*url"
	MetaZip                = "*zip"
	MetaXml                = "*xml"
	MetaOReq               = "*oreq"
	MetaReq                = "*req"
	MetaVars               = "*vars"
	MetaRep                = "*rep"
	MetaExp                = "*exp"
	MetaHdr                = "*hdr"
	MetaTrl                = "*trl"
	MetaTmp                = "*tmp"
	MetaOpts               = "*opts"
	MetaCfg                = "*cfg"
	MetaDynReq             = "~*req"
	MetaInitiate           = "*initiate"
	MetaUpdate             = "*update"
	MetaDelete             = "*delete"
	MetaTerminate          = "*terminate"
	MetaEvent              = "*event"
	MetaMessage            = "*message"
	MetaDAStats            = "*daStats"
	MetaDAThresholds       = "*daThresholds"
	MetaRAStats            = "*raStats"
	MetaRAThresholds       = "*raThresholds"
	MetaDNSStats           = "*dnsStats"
	MetaDNSThresholds      = "*dnsThresholds"
	MetaHAStats            = "*haStats"
	MetaHAThresholds       = "*haThresholds"
	MetaERsStats           = "*ersStats"
	MetaERsThresholds      = "*ersThresholds"
	MetaDryRun             = "*dryRun"
	Event                  = "Event"
	APIOpts                = "APIOpts"
	EventLowCase           = "event"
	EmptyString            = ""
	DynamicDataPrefix      = "~"
	AttrValueSep           = "="
	ANDSep                 = "&"
	PipeSep                = "|"
	RSRConstSep            = "`"
	RSRConstChar           = '`'
	RSRDataConverterPrefix = "{*"
	RSRDataConverterSufix  = "}"
	RSRDynStartChar        = '<'
	RSRDynEndChar          = '>'
	MetaApp                = "*app"
	MetaAppID              = "*appid"
	MetaCmd                = "*cmd"
	MetaEnv                = "*env:" // use in config for describing enviormant variables
	MetaTemplate           = "*template"
	MetaCCA                = "*cca"
	MetaErr                = "*err"
	OriginRealm            = "OriginRealm"
	ProductName            = "ProductName"
	IdxStart               = "["
	IdxEnd                 = "]"
	IdxCombination         = "]["

	MetaMemory              = "*memory"
	RemoteHost              = "RemoteHost"
	Local                   = "local"
	TCP                     = "tcp"
	UDP                     = "udp"
	VersionName             = "Version"
	MetaTenant              = "*tenant"
	ResourceUsageStr        = "ResourceUsage"
	MetaDuration            = "*duration"
	MetaLibPhoneNumber      = "*libphonenumber"
	MetaTimeString          = "*timeString"
	MetaIP2Hex              = "*ip2hex"
	MetaString2Hex          = "*string2hex"
	MetaUnixTime            = "*unixtime"
	MetaLen                 = "*len"
	MetaSlice               = "*slice"
	MetaSIPURIMethod        = "*sipuriMethod"
	MetaSIPURIHost          = "*sipuriHost"
	MetaSIPURIUser          = "*sipuriUser"
	MetaConnStatus          = "*connStatus"
	MetaConnID              = "*connID"
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
	Meta3GPPULI             = "*3gppULI"
	LoadIDs                 = "loadIDs"
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
	MetaText                = "*text"
	MetaAuthorize           = "*authorize"
	MetaSTIRAuthenticate    = "*stirAuthenticate"
	MetaSTIRInitiate        = "*stirInitiate"
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
	Synchronous             = "Synchronous"
	Attempts                = "Attempts"
	FieldSeparator          = "FieldSeparator"
	ExportPath              = "ExportPath"
	ExporterIDs             = "ExporterIDs"
	TimeNow                 = "TimeNow"
	ExportFileName          = "ExportFileName"
	GroupID                 = "GroupID"
	Recurrent               = "Recurrent"
	Executed                = "Executed"
	MinSleep                = "MinSleep"
	ExpirationDate          = "ExpirationDate"
	OrderIDStart            = "OrderIDStart"
	OrderIDEnd              = "OrderIDEnd"
	MinCost                 = "MinCost"
	MaxCost                 = "MaxCost"
	EeIDs                   = "EeIDs"
	MetaLoaders             = "*loaders"
	TmpSuffix               = ".tmp"
	MetaDiamreq             = "*diamreq"
	MetaRadDAReq            = "*radDAReq"
	MetaRadCoATemplate      = "*radCoATemplate"
	MetaRadDMRTemplate      = "*radDMRTemplate"
	MetaCost                = "*cost"
	MetaRateSCost           = "*rateSCost"
	MetaAccountsCost        = "*accountsCost"
	MetaGroup               = "*group"
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
	PoolID                  = "PoolID"
	PoolFilterIDs           = "PoolFilterIDs"
	PoolType                = "PoolType"
	PoolRange               = "PoolRange"
	PoolStrategy            = "PoolStrategy"
	PoolMessage             = "PoolMessage"
	PoolWeights             = "PoolWeights"
	PoolBlockers            = "PoolBlockers"
	SortedRoutes            = "SortedRoutes"
	MetaMonthly             = "*monthly"
	MetaYearly              = "*yearly"
	MetaDaily               = "*daily"
	MetaWeekly              = "*weekly"
	RateS                   = "RateS"
	Underline               = "_"
	MetaBusy                = "*busy"
	MetaQueue               = "*queue"
	MetaMonthEnd            = "*monthEnd"
	APIKey                  = "ApiKey"
	RouteID                 = "RouteID"
	MetaMonthlyEstimated    = "*monthlyEstimated"
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
	ActionWeights           = "ActionWeights"
	ActionBlockers          = "ActionBlockers"
	ActionDiktatsID         = "ActionDiktatsID"
	ActionDiktatsFilterIDs  = "ActionDiktatsFilterIDs"
	ActionDiktatsOpts       = "ActionDiktatsOpts"
	ActionDiktatsWeights    = "ActionDiktatsWeights"
	ActionDiktatsBlockers   = "ActionDiktatsBlockers"
	TPid                    = "TPid"
	LoadId                  = "LoadId"
	LoadID                  = "LoadID"
	RatingLoadID            = "RatingLoadID"
	AccountingLoadID        = "AccountingLoadID"
	LoadTime                = "LoadTime"
	LoadHistory             = "LoadHistory"
	ActionPlanId            = "ActionPlanId"
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
	MetaChargedUsage        = "*chargedUsage"
	MetaReservedUsage       = "*reservedUsage"
	MetaTotalUsage          = "*totalUsage"
	MetaInterimConsumed     = "*interimConsumed"
	MetaInterimUsage        = "*interimUsage"
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
	NodeID         = "nodeID"
	GoVersion      = "go_version"
	OSThreadsInUse = "osThreadsInUse"
	RunningSince   = "runningSince"
	OpenFiles      = "openFiles"
	ActiveMemory   = "activeMemory"
	SystemMemory   = "systemMemory"

	FieldVersion         = "version"
	FieldMemStats        = "memStats"
	FieldGCDurationStats = "gcDurationStats"
	FieldProcStats       = "procStats"
	FieldCapsStats       = "capsStats"

	MetricRuntimeGoroutines = "goroutines"
	MetricRuntimeThreads    = "threads"
	MetricRuntimeMaxProcs   = "maxprocs"

	MetricMemAlloc        = "alloc"
	MetricMemTotalAlloc   = "totalAlloc"
	MetricMemSys          = "sys"
	MetricMemMallocs      = "mallocs"
	MetricMemFrees        = "frees"
	MetricMemHeapAlloc    = "heapAlloc"
	MetricMemHeapSys      = "heapSys"
	MetricMemHeapIdle     = "heapIdle"
	MetricMemHeapInuse    = "heapInuse"
	MetricMemHeapReleased = "heapReleased"
	MetricMemHeapObjects  = "heapObjects"
	MetricMemStackInuse   = "stackInuse"
	MetricMemStackSys     = "stackSys"
	MetricMemMSpanSys     = "mSpanSys"
	MetricMemMSpanInuse   = "mSpanInuse"
	MetricMemMCacheInuse  = "mCacheInuse"
	MetricMemMCacheSys    = "mCacheSys"
	MetricMemBuckHashSys  = "buckHashSys"
	MetricMemGCSys        = "gcSys"
	MetricMemOtherSys     = "otherSys"
	MetricMemNextGC       = "nextGC"
	MetricMemLastGC       = "lastGC"
	MetricMemLimit        = "memLimit"

	MetricProcCPUTime              = "cpuTime"
	MetricProcMaxFDs               = "maxFDs"
	MetricProcOpenFDs              = "openFDs"
	MetricProcResidentMemory       = "residentMemory"
	MetricProcStartTime            = "startTime"
	MetricProcVirtualMemory        = "virtualMemory"
	MetricProcMaxVirtualMemory     = "maxVirtualMemory"
	MetricProcNetworkReceiveTotal  = "networkReceiveTotal"
	MetricProcNetworkTransmitTotal = "networkTransmitTotal"

	MetricGCQuantiles = "quantiles"
	MetricGCQuantile  = "quantile"
	MetricGCValue     = "value"
	MetricGCSum       = "sum"
	MetricGCCount     = "count"
	MetricGCPercent   = "gcPercent"

	MetricCapsAllocated = "capsAllocated"
	MetricCapsPeak      = "capsPeak"
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
	MetaResourceProfiles  = "*resourceProfiles"
	MetaIPProfiles        = "*ipProfiles"
	MetaStatQueueProfiles = "*statQueueProfiles"
	MetaStatQueues        = "*statQueues"
	MetaRankingProfiles   = "*rankingProfiles"
	MetaTrendProfiles     = "*trendProfiles"
	MetaThresholdProfiles = "*thresholdProfiles"
	MetaRouteProfiles     = "*routeProfiles"
	MetaAttributeProfiles = "*attributeProfiles"
	MetaRateProfiles      = "*rateProfiles"
	MetaRateProfileRates  = "*rateProfileRates"
	MetaChargerProfiles   = "*chargerProfiles"
	MetaIPAllocations     = "*ipAllocations"
	MetaThresholds        = "*thresholds"
	MetaRoutes            = "*routes"
	MetaAttributes        = "*attributes"
	MetaActionProfiles    = "*actionProfiles"
	MetaLoadIDs           = "*loadIDs"
	MetaNodeID            = "*nodeID"
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
	MetaREPSC    = "*repsc"
	MetaREPFC    = "*repfc"
	MetaAverage  = "*average"
	MetaDistinct = "*distinct"
	MetaHighest  = "*highest"
	MetaLowest   = "*lowest"
)

// Diameter/Radius request types
const (
	MetaRAR = "*rar"
	MetaDMR = "*dmr"
	MetaCoA = "*coa"
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
	MetaLog            = "*log"
	MetaTopUpReset     = "*topUpReset"
	MetaTopUp          = "*topup"
	MetaDebit          = "*debit"
	MetaEnableAccount  = "*enableAccount"
	MetaDisableAccount = "*disableAccount"
	MetaUnlimited      = "*unlimited"
	CDRLog             = "*cdrlog"
	MetaRpc            = "*rpc"
	MetaResetThreshold = "*resetThreshold"
	MetaResetStatQueue = "*resetStatQueue"
	ActionID           = "ActionID"
	ActionType         = "ActionType"
	BalanceValue       = "BalanceValue"
	BalanceUnits       = "BalanceUnits"
	BalanceUnitFactors = "BalanceUnitFactors"
	ExtraParameters    = "ExtraParameters"

	MetaAddBalance            = "*addBalance"
	MetaSetBalance            = "*setBalance"
	MetaRemBalance            = "*remBalance"
	DynaprepaidActionplansCfg = "dynaprepaidActionProfile"
	MetaDynamicThreshold      = "*dynamicThreshold"
	MetaDynamicStats          = "*dynamicStats"
	MetaDynamicAttribute      = "*dynamicAttribute"
	MetaDynamicResource       = "*dynamicResource"
	MetaDynamicTrend          = "*dynamicTrend"
	MetaDynamicRanking        = "*dynamicRanking"
	MetaDynamicFilter         = "*dynamicFilter"
	MetaDynamicRoute          = "*dynamicRoute"
	MetaDynamicRate           = "*dynamicRate"
	MetaDynamicIP             = "*dynamicIP"
	MetaDynamicAction         = "*dynamicAction"

	// Diktats Opts Fields
	MetaBalancePath  = "*balancePath"
	MetaBalanceValue = "*balanceValue"
	MetaURL          = "*url"
)

// Migrator Metas
const (
	MetaSetVersions         = "*set_versions"
	MetaEnsureIndexes       = "*ensureIndexes"
	MetaDurationSeconds     = "*durationSeconds"
	MetaDurationNanoseconds = "*durationNanoseconds"
	CapAttributes           = "Attributes"
	CapResourceAllocation   = "ResourceAllocation"
	AllocatedIPField        = "AllocatedIP"
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
	MetaRatio          = "*ratio"
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
	ReplicatorS                      = "ReplicatorS"
	ReplicatorSv1                    = "ReplicatorSv1"
	ReplicatorSv1Ping                = "ReplicatorSv1.Ping"
	ReplicatorSv1GetStatQueue        = "ReplicatorSv1.GetStatQueue"
	ReplicatorSv1GetFilter           = "ReplicatorSv1.GetFilter"
	ReplicatorSv1GetThreshold        = "ReplicatorSv1.GetThreshold"
	ReplicatorSv1GetThresholdProfile = "ReplicatorSv1.GetThresholdProfile"
	ReplicatorSv1GetStatQueueProfile = "ReplicatorSv1.GetStatQueueProfile"
	ReplicatorSv1GetRanking          = "ReplicatorSv1.GetRanking"
	ReplicatorSv1GetRankingProfile   = "ReplicatorSv1.GetRankingProfile"
	ReplicatorSv1GetTrendProfile     = "ReplicatorSv1.GetTrendProfile"
	ReplicatorSv1GetTrend            = "ReplicatorSv1.GetTrend"
	ReplicatorSv1GetResource         = "ReplicatorSv1.GetResource"
	ReplicatorSv1GetResourceProfile  = "ReplicatorSv1.GetResourceProfile"
	ReplicatorSv1GetIPAllocations    = "ReplicatorSv1.GetIPAllocations"
	ReplicatorSv1GetIPProfile        = "ReplicatorSv1.GetIPProfile"
	ReplicatorSv1GetRouteProfile     = "ReplicatorSv1.GetRouteProfile"
	ReplicatorSv1GetAttributeProfile = "ReplicatorSv1.GetAttributeProfile"
	ReplicatorSv1GetChargerProfile   = "ReplicatorSv1.GetChargerProfile"
	ReplicatorSv1GetRateProfile      = "ReplicatorSv1.GetRateProfile"
	ReplicatorSv1GetActionProfile    = "ReplicatorSv1.GetActionProfile"
	ReplicatorSv1GetAccount          = "ReplicatorSv1.GetAccount"
	ReplicatorSv1GetItemLoadIDs      = "ReplicatorSv1.GetItemLoadIDs"
	ReplicatorSv1SetThresholdProfile = "ReplicatorSv1.SetThresholdProfile"
	ReplicatorSv1SetThreshold        = "ReplicatorSv1.SetThreshold"
	ReplicatorSv1SetStatQueue        = "ReplicatorSv1.SetStatQueue"
	ReplicatorSv1SetFilter           = "ReplicatorSv1.SetFilter"
	ReplicatorSv1SetStatQueueProfile = "ReplicatorSv1.SetStatQueueProfile"
	ReplicatorSv1SetRankingProfile   = "ReplicatorSv1.SetRankingProfile"
	ReplicatorSv1SetRanking          = "ReplicatorSv1.SetRanking"
	ReplicatorSv1SetTrendProfile     = "ReplicatorSv1.SetTrendProfile"
	ReplicatorSv1SetTrend            = "ReplicatorSv1.SetTrend"
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
	AdminSv1ReplayFailedReplications          = "AdminSv1.ReplayFailedReplications"
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
	AdminSv1DumpDB                    = "AdminSv1.DumpDB"
	AdminSv1RewriteDB                 = "AdminSv1.RewriteDB"
	AdminSv1BackupDB                  = "AdminSv1.BackupDB"
	AdminSv1RestoreDB                 = "AdminSv1.RestoreDB"
	AdminSv1SnapshotDB                = "AdminSv1.SnapshotDB"
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
	IPsV1ClearIPAllocations      = "IPsV1.ClearIPAllocations"
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
	SessionSv1AlterSession               = "SessionSv1.AlterSession"
	SessionSv1DisconnectPeer             = "SessionSv1.DisconnectPeer"
	SessionSv1STIRAuthenticate           = "SessionSv1.STIRAuthenticate"
	SessionSv1STIRIdentity               = "SessionSv1.STIRIdentity"
	SessionSv1Sleep                      = "SessionSv1.Sleep"
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
	CacheSv1GetStats          = "CacheSv1.GetStats"
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
	EeSv1ResetExporterMetrics = "EeSv1.ResetExporterMetrics"
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
	CGRRoute           = "cgrRoute"
	CGRDisconnectCause = "cgrDisconnectCause"
	CGRFlags           = "cgrFlags"
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
	TBLTPIPs             = "tp_ips"
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
	TBLAccounts          = "accounts"
	TBLIPProfiles        = "ip_profiles"
	TBLIPAllocations     = "ip_allocations"
	TBLActionProfiles    = "action_profiles"
	TBLChargerProfiles   = "charger_profiles"
	TBLAttributeProfiles = "attribute_profiles"
	TBLResourceProfiles  = "resource_profiles"
	TBLResources         = "resources"
	TBLStatQueueProfiles = "stat_queue_profiles"
	TBLStatQueues        = "stat_queues"
	TBLThresholdProfiles = "threshold_profiles"
	TBLThresholds        = "thresholds"
	TBLFilters           = "filters"
	TBLRouteProfiles     = "route_profiles"
	TBLRateProfiles      = "rate_profiles"
	TBLRates             = "rates"
	TBLRankingProfiles   = "ranking_profiles"
	TBLRankings          = "rankings"
	TBLTrendProfiles     = "trend_profiles"
	TBLTrends            = "trends"
	TBLLoadIDs           = "load_ids"
	TBLIndexes           = "indexes"
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
	CacheResourceProfiles            = "*resourceProfiles"
	CacheEventResources              = "*eventResources"
	CacheIPAllocations               = "*ipAllocations"
	CacheIPProfiles                  = "*ipProfiles"
	CacheEventIPs                    = "*eventIPs"
	CacheStatQueueProfiles           = "*statQueueProfiles"
	CacheStatQueues                  = "*statQueues"
	CacheRankingProfiles             = "*rankingProfiles"
	CacheRankings                    = "*rankings"
	CacheTrendProfiles               = "*trendProfiles"
	CacheTrends                      = "*trends"
	CacheThresholdProfiles           = "*thresholdProfiles"
	CacheThresholds                  = "*thresholds"
	CacheFilters                     = "*filters"
	CacheRouteProfiles               = "*routeProfiles"
	CacheAttributeProfiles           = "*attributeProfiles"
	CacheChargerProfiles             = "*chargerProfiles"
	CacheRateProfiles                = "*rateProfiles"
	CacheActionProfiles              = "*actionProfiles"
	CacheAccounts                    = "*accounts"
	CacheResourceFilterIndexes       = "*resourceFilterIndexes"
	CacheIPFilterIndexes             = "*ipFilterIndexes"
	CacheStatFilterIndexes           = "*statFilterIndexes"
	CacheThresholdFilterIndexes      = "*thresholdFilterIndexes"
	CacheRouteFilterIndexes          = "*routeFilterIndexes"
	CacheAttributeFilterIndexes      = "*attributeFilterIndexes"
	CacheChargerFilterIndexes        = "*chargerFilterIndexes"
	CacheDiameterMessages            = "*diameterMessages"
	CacheRadiusPackets               = "*radiusPackets"
	CacheRPCResponses                = "*rpcResponses"
	CacheClosedSessions              = "*closedSessions"
	CacheRateProfilesFilterIndexes   = "*rateProfileFilterIndexes"
	CacheActionProfilesFilterIndexes = "*actionProfileFilterIndexes"
	CacheAccountsFilterIndexes       = "*accountFilterIndexes"
	CacheRateFilterIndexes           = "*rateFilterIndexes"
	MetaPrecaching                   = "*precaching"
	MetaReady                        = "*ready"
	CacheLoadIDs                     = "*loadIDs"
	CacheRPCConnections              = "*rpcConnections"
	CacheCDRIDs                      = "*cdrIDs"
	CacheUCH                         = "*uch"
	CacheSTIR                        = "*stir"
	CacheEventCharges                = "*eventCharges"
	CacheReverseFilterIndexes        = "*reverseFilterIndexes"
	CacheVersions                    = "*versions"
	CacheCapsEvents                  = "*capsEvents"
	CacheReplicationHosts            = "*replicationHosts"
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
	NodeIDCfg               = "nodeID"
	LoggerCfg               = "logger"
	LogLevelCfg             = "log_level"
	RoundingDecimalsCfg     = "roundingDecimals"
	DBDataEncodingCfg       = "dbDataEncoding"
	TpExportPathCfg         = "tpExportDir"
	PosterAttemptsCfg       = "posterAttempts"
	FailedPostsDirCfg       = "failedPostsDir"
	FailedPostsTTLCfg       = "failedPostsTTL"
	FailedPostsStaticTTLCfg = "failedPostsStaticTTL"
	DefaultReqTypeCfg       = "defaultRequestType"
	DefaultCategoryCfg      = "defaultCategory"
	DefaultTenantCfg        = "defaultTenant"
	DefaultTimezoneCfg      = "defaultTimezone"
	DefaultCachingCfg       = "defaultCaching"
	CachingDlayCfg          = "cachingDelay"
	ConnectAttemptsCfg      = "connectAttempts"
	ReconnectsCfg           = "reconnects"
	MaxReconnectIntervalCfg = "maxReconnectInterval"
	AriWebSocketCfg         = "ariWebsocket"
	ConnectTimeoutCfg       = "connectTimeout"
	ReplyTimeoutCfg         = "replyTimeout"
	LockingTimeoutCfg       = "lockingTimeout"
	DigestSeparatorCfg      = "digestSeparator"
	DigestEqualCfg          = "digestEqual"
	MaxParallelConnsCfg     = "maxParallelConns"
	EEsConnsCfg             = "eesConns"
	DecimalMaxScaleCfg      = "decimalMaxScale"
	DecimalMinScaleCfg      = "decimalMinScale"
	DecimalPrecisionCfg     = "decimalPrecision"
	DecimalRoundingModeCfg  = "decimalRoundingMode"
)

const (
	LevelCfg         = "level"
	KafkaConnCfg     = "kafkaConn"
	KafkaTopicCfg    = "kafkaTopic"
	KafkaAttemptsCfg = "kafkaAttempts"
)

const (
	TypeCfg                   = "type"
	SQLMaxOpenConnsCfg        = "sqlMaxOpenConns"
	SQLMaxIdleConnsCfg        = "sqlMaxIdleConns"
	SQLLogLevelCfg            = "sqlLogLevel"
	SQLConnMaxLifetimeCfg     = "sqlConnMaxLifetime"
	StringIndexedFieldsCfg    = "stringIndexedFields"
	PrefixIndexedFieldsCfg    = "prefixIndexedFields"
	SuffixIndexedFieldsCfg    = "suffixIndexedFields"
	ExistsIndexedFieldsCfg    = "existsIndexedFields"
	NotExistsIndexedFieldsCfg = "notExistsIndexedFields"
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
	ConnsCfg                  = "conns"
	Tenants                   = "tenants"
	MysqlLocation             = "mysqlLocation"
)

// DbCfg
const (
	DbConnsCfg                   = "dbConns"
	DbTypeCfg                    = "dbType"
	DbHostCfg                    = "dbHost"
	DbPortCfg                    = "dbPort"
	DbNameCfg                    = "dbName"
	DbUserCfg                    = "dbUser"
	DbPassCfg                    = "dbPassword"
	InternalDBDumpPathCfg        = "internalDBDumpPath"
	InternalDBBackupPathCfg      = "internalDBBackupPath"
	InternalDBStartTimeoutCfg    = "internalDBStartTimeout"
	InternalDBDumpIntervalCfg    = "internalDBDumpInterval"
	InternalDBRewriteIntervalCfg = "internalDBRewriteInterval"
	InternalDBFileSizeLimitCfg   = "internalDBFileSizeLimit"
	RedisBatchSizeCfg            = "redisBatchSize"
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
	ReplicationFilteredCfg       = "replicationFiltered"
	ReplicationCache             = "replicationCache"
	RemoteConnIDCfg              = "remoteConnID"
	ReplicationFailedDirCfg      = "replicationFailedDir"
	ReplicationIntervalCfg       = "replicationInterval"
)

// ItemOpt
const (
	APIKeyCfg    = "apiKey"
	RouteIDCfg   = "routeID"
	RemoteCfg    = "remote"
	ReplicateCfg = "replicate"
	TTLCfg       = "ttl"
	LimitCfg     = "limit"
	StaticTTLCfg = "staticTTL"
	DBConnCfg    = "dbConn"
)

// Tls
const (
	ServerCerificateCfg = "serverCertificate"
	ServerKeyCfg        = "serverKey"
	ServerPolicyCfg     = "serverPolicy"
	ServerNameCfg       = "serverName"
	ClientCerificateCfg = "clientCertificate"
	ClientKeyCfg        = "clientKey"
	CaCertificateCfg    = "caCertificate"
)

// ListenCfg
const (
	RPCJSONListenCfg    = "rpcJSON"
	RPCGOBListenCfg     = "rpcGOB"
	HTTPListenCfg       = "http"
	RPCJSONTLSListenCfg = "rpcJSONtls"
	RPCGOBTLSListenCfg  = "rpcGOBtls"
	HTTPTLSListenCfg    = "httpTLS"
)

// HTTPCfg
const (
	HTTPJsonRPCURLCfg        = "jsonRPCurl"
	RegistrarSURLCfg         = "registrarsURL"
	HTTPWSURLCfg             = "wsURL"
	HTTPFreeswitchCDRsURLCfg = "freeswitchCDRsURL"
	HTTPCDRsURLCfg           = "httpCDRs"
	PprofPathCfg             = "pprofPath"
	HTTPUseBasicAuthCfg      = "useBasicAuth"
	HTTPAuthUsersCfg         = "authUsers"
	HTTPClientOptsCfg        = "clientOpts"

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

const (
	EnabledCfg      = "enabled"
	CacheSConnsCfg  = "cachesConns"
	ScheduledIDsCfg = "scheduledIDs"
)

// Efs
const (
	EFsConnsCfg = "efsConns"
)

// CdrsCfg
const (
	FiltersCfg             = "filters"
	ExtraFieldsCfg         = "extraFields"
	SMCostRetriesCfg       = "sessionCostRetries"
	RetransmissionTimerCfg = "retransmissionTimer"
	OnlineCDRExportsCfg    = "onlineCDRExports"
	SessionCostRetires     = "sessionCostRetries"
)

// SessionSCfg
const (
	ListenBijsonCfg        = "listenBiJSON"
	ListenBigobCfg         = "listenBiGob"
	ReplicationConnsCfg    = "replicationConns"
	RemoteConnsCfg         = "remoteConns"
	DebitIntervalCfg       = "debitInterval"
	StoreSCostsCfg         = "storeSessionCosts"
	SessionTTLCfg          = "sessionTTL"
	SessionTTLMaxDelayCfg  = "sessionTTLMaxDelay"
	SessionTTLLastUsedCfg  = "sessionTTLLastUsed"
	SessionTTLUsageCfg     = "sessionTTLUsage"
	SessionIndexesCfg      = "sessionIndexes"
	ClientProtocolCfg      = "clientProtocol"
	ChannelSyncIntervalCfg = "channelSyncInterval"
	TerminateAttemptsCfg   = "terminateAttempts"
	AlterableFieldsCfg     = "alterableFields"
	MinDurLowBalanceCfg    = "minDurLowBalance"
	DefaultUsageCfg        = "defaultUsage"
	STIRCfg                = "stir"

	AllowedAtestCfg       = "allowedAttest"
	PayloadMaxdurationCfg = "payloadMaxduration"
	DefaultAttestCfg      = "defaultAttest"
	PublicKeyPathCfg      = "publicKeyPath"
	PrivateKeyPathCfg     = "privateKeyPath"
)

// FsAgentCfg
const (
	SubscribeParkCfg          = "subscribePark"
	CreateCdrCfg              = "createCDR"
	LowBalanceAnnFileCfg      = "lowBalanceAnnFile"
	EmptyBalanceContextCfg    = "emptyBalanceContext"
	EmptyBalanceAnnFileCfg    = "emptyBalanceAnnFile"
	MaxWaitConnectionCfg      = "maxWaitConnection"
	ActiveSessionDelimiterCfg = "activeSessionDelimiter"
	EventSocketConnsCfg       = "eventSocketConns"
	EmptyBalanceContext       = "emptyBalanceContext"
)

// From Config
const (
	AddressCfg = "address"
	Password   = "password"
	AliasCfg   = "alias"

	// KamAgentCfg
	EvapiConnsCfg = "evapiConns"
	TimezoneCfg   = "timezone"

	// AsteriskConnCfg
	UserCf = "user"

	// AsteriskAgentCfg
	AsteriskConnsCfg = "asteriskConns"

	// DiameterAgentCfg
	ListenNetCfg                  = "listenNet"
	NetworkCfg                    = "network"
	ListenersCfg                  = "listeners"
	ListenCfg                     = "listen"
	DictionariesPathCfg           = "dictionariesPath"
	DictionariesAppendDefaultsCfg = "dictionariesAppendDefaults"
	CEApplicationsCfg             = "ceApplications"
	OriginHostCfg                 = "originHost"
	OriginRealmCfg                = "originRealm"
	VendorIDCfg                   = "vendorID"
	ProductNameCfg                = "productName"
	SyncedConnReqsCfg             = "syncedConnRequests"
	ASRTemplateCfg                = "asrTemplate"
	RARTemplateCfg                = "rarTemplate"
	ForcedDisconnectCfg           = "forcedDisconnect"
	ConnStatusStatQueueIDsCfg     = "connStatusStatQueueIDs"
	ConnStatusThresholdIDsCfg     = "connStatusThresholdIDs"
	ConnHealthCheckIntervalCfg    = "connHealthCheckInterval"
	TemplatesCfg                  = "templates"
	RequestProcessorsCfg          = "requestProcessors"

	// RequestProcessor
	RequestFieldsCfg = "requestFields"
	ReplyFieldsCfg   = "replyFields"

	// RadiusAgentCfg
	AuthAddrCfg           = "authAddress"
	AcctAddrCfg           = "acctAddress"
	ClientSecretsCfg      = "clientSecrets"
	ClientDictionariesCfg = "clientDictionaries"
	ClientDaAddressesCfg  = "clientDaAddresses"
	RequestsCacheKeyCfg   = "requestsCacheKey"
	DMRTemplateCfg        = "dmrTemplate"
	CoATemplateCfg        = "coaTemplate"
	HostCfg               = "host"
	PortCfg               = "port"

	// JanusAgentCfg
	JanusConnsCfg    = "janusConns"
	AdminAddressCfg  = "adminAddress"
	AdminPasswordCfg = "adminPassword"

	// PrometheusAgentCfg
	CollectGoMetricsCfg      = "collectGoMetrics"
	CollectProcessMetricsCfg = "collectProcessMetrics"
	CacheIDsCfg              = "cacheIDs"
	StatQueueIDsCfg          = "statQueueIDs"

	// AttributeSCfg
	IndexedSelectsCfg  = "indexedSelects"
	ProfileRunsCfg     = "profileRuns"
	NestedFieldsCfg    = "nestedFields"
	MetaProcessRunsCfg = "*processRuns"
	MetaProfileRunsCfg = "*profileRuns"

	// ChargerSCfg
	StoreIntervalCfg = "storeInterval"

	// StatSCfg
	StoreUncompressedLimitCfg = "storeUncompressedLimit"
	EEsExporterIDsCfg         = "eesExporterIDs"

	// Cache
	PartitionsCfg = "partitions"
	PrecacheCfg   = "precache"

	// EEsCfg
	ExportPathCfg           = "exportPath"
	SynchronousCfg          = "synchronous"
	AttemptsCfg             = "attempts"
	AttributeContextCfg     = "attributeContext"
	AttributeIDsCfg         = "attributeIDs"
	ConcurrentRequestsCfg   = "concurrentRequests"
	MetricsResetScheduleCfg = "metricsResetSchedule"

	//LoaderSCfg
	DryRunCfg       = "dryRun"
	LockFilePathCfg = "lockfilePath"
	TpInPathCfg     = "tpInPath"
	TpOutPathCfg    = "tpOutPath"
	DataCfg         = "data"

	DefaultRatioCfg   = "defaultRatio"
	ReadersCfg        = "readers"
	ExportersCfg      = "exporters"
	PoolSize          = "poolSize"
	Conns             = "conns"
	FilenameCfg       = "fileName"
	RequestPayloadCfg = "requestPayload"
	ReplyPayloadCfg   = "replyPayload"
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
	RateIndexedSelectsCfg         = "rateIndexedSelects"
	RateNestedFieldsCfg           = "rateNestedFields"
	RateStringIndexedFieldsCfg    = "rateStringIndexedFields"
	RatePrefixIndexedFieldsCfg    = "ratePrefixIndexedFields"
	RateSuffixIndexedFieldsCfg    = "rateSuffixIndexedFields"
	RateExistsIndexedFieldsCfg    = "rateExistsIndexedFields"
	RateNotExistsIndexedFieldsCfg = "rateNotExistsIndexedFields"
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
	MetaResourcesUsageID          = "*resourcesUsageID"
	MetaIPsAuthorizeCfg           = "*ipsAuthorize"
	MetaAccountsAuthorizeCfg      = "*accountsAuthorize"
	MetaAccountsDebitCfg          = "*accountsDebit"
	MetaIPsAllocateCfg            = "*ipsAllocate"
	MetaIPsReleaseCfg             = "*ipsRelease"
	MetaRoutesDerivedReplyCfg     = "*routesDerivedReply"
	MetaStatsDerivedReplyCfg      = "*statsDerivedReply"
	MetaThresholdsDerivedReplyCfg = "*thresholdsDerivedReply"
	MetaMaxUsageCfg               = "*maxUsage"
	MetaForceUsageCfg             = "*forceUsage"
	MetaTTLCfg                    = "*ttl"
	MetaChargeableCfg             = "*chargeable"
	MetaAutoChargeIntervalCfg     = "*autoChargeInterval"
	MetaTTLLastUsageCfg           = "*ttlLastUsage"
	MetaTTLLastUsedCfg            = "*ttlLastUsed"
	MetaTTLMaxDelayCfg            = "*ttlMaxDelay"
	MetaTTLUsageCfg               = "*ttlUsage"
	MetaAccountsForceUsage        = "*accountsForceUsage"

	// AnalyzerSCfg
	CleanupIntervalCfg = "cleanupInterval"
	IndexTypeCfg       = "indexType"
	DBPathCfg          = "dbPath"

	// CoreSCfg
	CapsCfg              = "caps"
	CapsStrategyCfg      = "capsStrategy"
	CapsStatsIntervalCfg = "capsStatsInterval"
	ShutdownTimeoutCfg   = "shutdownTimeout"

	// AccountSCfg
	MaxIterations = "maxIterations"
	MaxUsage      = "maxUsage"
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
	NewBranchCfg       = "newBranch"
	BlockerCfg         = "blocker"
	LayoutCfg          = "layout"
	CostShiftDigitsCfg = "costShiftDigits"
	MaskDestIDCfg      = "maskDestinationdID"
	MaskLenCfg         = "maskLength"
)

// SureTax
const (
	RootDirCfg              = "rootDir"
	URLCfg                  = "url"
	ClientNumberCfg         = "clientNumber"
	ValidationKeyCfg        = "validationKey"
	BusinessUnitCfg         = "businessUnit"
	IncludeLocalCostCfg     = "includeLocalCost"
	ReturnFileCodeCfg       = "returnFileCode"
	ResponseGroupCfg        = "responseGroup"
	ResponseTypeCfg         = "responseType"
	RegulatoryCodeCfg       = "regulatoryCode"
	ClientTrackingCfg       = "clientTracking"
	CustomerNumberCfg       = "customerNumber"
	OrigNumberCfg           = "origNumber"
	TermNumberCfg           = "termNumber"
	BillToNumberCfg         = "billToNumber"
	ZipcodeCfg              = "zipcode"
	Plus4Cfg                = "plus4"
	P2PZipcodeCfg           = "p2pzipcode"
	P2PPlus4Cfg             = "p2pplus4"
	UnitsCfg                = "units"
	UnitTypeCfg             = "unitType"
	TaxIncludedCfg          = "taxIncluded"
	TaxSitusRuleCfg         = "taxSitusRule"
	TransTypeCodeCfg        = "transTypeCode"
	SalesTypeCodeCfg        = "salesTypeCode"
	TaxExemptionCodeListCfg = "taxExemptionCodeList"
)

// LoaderCgrCfg
const (
	TpIDCfg            = "tpid"
	DataPathCfg        = "dataPath"
	DisableReverseCfg  = "disableReverse"
	CachesConnsCfg     = "cachesConns"
	ActionSConnsCfg    = "actionsConns"
	GapiCredentialsCfg = "gapiCredentials"
	GapiTokenCfg       = "gapiToken"
)

// MigratorCgrCfg
const (
	OutDBRedisSentinel = "outRedisSentinel"
	OutDBOptsCfg       = "outDBOpts"
	UsersFiltersCfg    = "usersFilters"
	FromItemsCfg       = "fromItems"
)

// MailerCfg
const (
	MailerServerCfg = "server"
)

// EventReaderCfg
const (
	IDCfg                  = "id"
	CacheCfg               = "cache"
	FieldSepCfg            = "fieldSeparator"
	RunDelayCfg            = "runDelay"
	StartDelayCfg          = "startDelay"
	SourcePathCfg          = "sourcePath"
	ProcessedPathCfg       = "processedPath"
	TenantCfg              = "tenant"
	FilterIDsCfg           = "filterIDs"
	EEsSuccessIDsCfg       = "eesSuccessIDs"
	EEsFailedIDsCfg        = "eesFailedIDs"
	FlagsCfg               = "flags"
	FieldsCfg              = "fields"
	CacheDumpFieldsCfg     = "cacheDumpFields"
	PartialCommitFieldsCfg = "partialCommitFields"
	PartialCacheTTLCfg     = "partialCacheTTL"
	ActionCfg              = "action"
)

// RegistrarCCfg
const (
	RPCCfg             = "rpc"
	DispatcherCfg      = "dispatchers"
	RegistrarsConnsCfg = "registrarsConns"
	HostsCfg           = "hosts"
	RefreshIntervalCfg = "refreshInterval"
)

// APIBanCfg
const (
	KeysCfg = "keys"
)

const (
	ClientIDCfg     = "clientID"
	ClientSecretCfg = "clientSecret"
	TokenUrlCfg     = "tokenUrl"
	IpsUrlCfg       = "ipsUrl"
	NumbersUrlCfg   = "numbersUrl"
	AudienceCfg     = "audience"
	GrantTypeCfg    = "grantType"
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
	OptsSesMaxUsage, OptsSesForceUsage, MetaInitiate, MetaUpdate, MetaTerminate,
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
	OptsSesMessage                = "*sesMessage"

	// Accounts
	OptsAccountsUsage      = "*accountsUsage"
	OptsAccountsForceUsage = "*accountsForceUsage"
	OptsAccountsProfileIDs = "*accountsProfileIDs"

	// Actions
	OptsActionsProfileIDs = "*actionsProfileIDs"

	// Attributes
	OptsAttributesProfileIDs  = "*attributesProfileIDs"
	OptsAttributesProfileRuns = "*attributesProfileRuns"
	OptsAttributesProcessRuns = "*attributesProcessRuns"

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
	OptsRatesProfileIDs    = "*ratesProfileIDs"
	OptsRatesStartTime     = "*ratesStartTime"
	OptsRatesUsage         = "*ratesUsage"
	OptsRatesIntervalStart = "*ratesIntervalStart"

	// Resources
	OptsResourcesUnits    = "*resourcesUnits"
	OptsResourcesUsageID  = "*resourcesUsageID"
	OptsResourcesUsageTTL = "*resourcesUsageTTL"

	// IPs
	OptsIPsAllocationID = "*ipsAllocationID"
	OptsIPsTTL          = "*ipsTTL"
	MetaAllocationID    = "*allocationID"

	// Routes
	OptsRoutesProfilesCount = "*routesProfilesCount"
	OptsRoutesLimit         = "*routesLimit"
	OptsRoutesOffset        = "*routesOffset"
	OptsRoutesMaxItems      = "*routesMaxItems"
	OptsRoutesIgnoreErrors  = "*routesIgnoreErrors"
	OptsRoutesMaxCost       = "*routesMaxCost"
	OptsRoutesUsage         = "*routesUsage"

	// Stats
	OptsStatsProfileIDs  = "*statsProfileIDs"
	OptsRoundingDecimals = "*roundingDecimals"
	SQItems              = "SQItems"
	SQMetrics            = "SQMetrics"
	Compressed           = "Compressed"

	// Thresholds
	Hits                     = "Hits"
	Snooze                   = "Snooze"
	ThresholdConfig          = "Config"
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
	MetaDerivedReply = "*derivedReply"

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
	DefaultQueueID = "cgratesCDRs"

	// sqs and s3
	AWSRegion = "awsRegion"
	AWSKey    = "awsKey"
	AWSSecret = "awsSecret"
	AWSToken  = "awsToken"

	// sqs
	SQSQueueID        = "sqsQueueID"
	SQSForcePathStyle = "sqsForcePathStyle"
	SQSSkipTlsVerify  = "sqsSkipTlsVerify"

	// s3
	S3Bucket         = "s3BucketID"
	S3FolderPath     = "s3FolderPath"
	S3ForcePathStyle = "s3ForcePathStyle"
	S3SkipTlsVerify  = "s3SkipTlsVerify"

	// sql
	SQLDefaultDBName    = "cgrates"
	SQLDefaultPgSSLMode = "disable"

	SQLDBNameOpt              = "sqlDBName"
	SQLUpdateIndexedFieldsOpt = "sqlUpdateIndexedFields"
	SQLTableNameOpt           = "sqlTableName"
	SQLBatchSize              = "sqlBatchSize"
	SQLDeleteIndexedFieldsOpt = "sqlDeleteIndexedFields"

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

	KafkaTopic           = "kafkaTopic"
	KafkaLinger          = "kafkaLinger"
	KafkaTLS             = "kafkaTLS"
	KafkaCAPath          = "kafkaCAPath"
	KafkaSkipTLSVerify   = "kafkaSkipTLSVerify"
	KafkaDeliveryTimeout = "kafkaDeliveryTimeout"
	KafkaGroupID         = "kafkaGroupID"
	KafkaMaxWait         = "kafkaMaxWait"

	// partial
	PartialOpt      = "*partial"
	PartialRatesOpt = "*partialRates"

	PartialOrderFieldOpt       = "partialOrderField"
	PartialCacheActionOpt      = "partialCacheAction"
	PartialPathOpt             = "partialPath"
	PartialCSVFieldSepartorOpt = "partialcsvFieldSeparator"

	IgnoreErroredItemsOpt = "ignoreErroredItems"

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
	DataDBTypeCgr   = "dataDBType"
	DataDBHostCgr   = "dataDBHost"
	DataDBPortCgr   = "dataDBPort"
	DataDBNameCgr   = "dataDBName"
	DataDBUserCgr   = "dataDBUser"
	DataDBPasswdCgr = "dataDBPasswd"
	//Cgr console
	CgrConsole     = "cgr-console"
	HomeCgr        = "HOME"
	HistoryCgr     = "/.cgr_history"
	RpcEncodingCgr = "rpcEncoding"
	CertPathCgr    = "crtPath"
	KeyPathCgr     = "keyPath"
	CAPathCgr      = "caPath"
	HelpCgr        = "help"
	SepCgr         = " "
	//Cgr engine
	CgrEngine            = "cgr-engine"
	PrintCfgCgr          = "printConfig"
	CheckCfgCgr          = "checkConfig"
	PidCgr               = "pid"
	CpuProfDirCgr        = "cpuProfDir"
	MemProfDirCgr        = "memProfDir"
	MemProfIntervalCgr   = "memProfInterval"
	MemProfMaxFilesCgr   = "memProfMaxFiles"
	MemProfTimestampCgr  = "memProfTimestamp"
	ScheduledShutdownCgr = "scheduledShutdown"
	SingleCpuCgr         = "singleCPU"
	PreloadCgr           = "preload"
	SetVersionsCgr       = "setVersions"
	MemProfFinalFile     = "mem_final.prof"
	CpuPathCgr           = "cpu.prof"
	//Cgr loader
	CgrLoader         = "cgr-loader"
	CachingArgCgr     = "caching"
	FieldSepCgr       = "fieldSep"
	ImportIDCgr       = "importID"
	DisableReverseCgr = "disableReverseMappings"
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
