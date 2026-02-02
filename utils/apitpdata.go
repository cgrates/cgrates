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
	"strconv"
	"strings"
	"time"
)

type ArgsItemIDs struct {
	Tenant      string
	APIOpts     map[string]any
	ItemsSearch string // Search for items containing this string
}

type ArgsSubItemIDs struct {
	Tenant      string
	ProfileID   string
	APIOpts     map[string]any
	ItemsPrefix string
}

type AttrGetCdrs struct {
	//CgrIds          []string // If provided, it will filter based on the cgrids present in list
	MediationRunIds []string // If provided, it will filter on mediation runid

	TORs                []string // If provided, filter on TypeOfRecord
	CdrHosts            []string // If provided, it will filter cdrhost
	CdrSources          []string // If provided, it will filter cdrsource
	ReqTypes            []string // If provided, it will fiter reqtype
	Tenants             []string // If provided, it will filter tenant
	Categories          []string // If provided, it will filter çategory
	Accounts            []string // If provided, it will filter account
	Subjects            []string // If provided, it will filter the rating subject
	DestinationPrefixes []string // If provided, it will filter on destination prefix
	RatedAccounts       []string // If provided, it will filter ratedaccount
	RatedSubjects       []string // If provided, it will filter the ratedsubject
	OrderIdStart        *int64   // Export from this order identifier
	OrderIdEnd          *int64   // Export smaller than this order identifier
	TimeStart           string   // If provided, it will represent the starting of the CDRs interval (>=)
	TimeEnd             string   // If provided, it will represent the end of the CDRs interval (<)
	SkipErrors          bool     // Do not export errored CDRs
	SkipRated           bool     // Do not export rated CDRs
	OrderBy             string   // Ascendent/Descendent
	Paginator
}

type AttrLoadTpFromFolder struct {
	FolderPath string // Take files from folder absolute path
	DryRun     bool   // Do not write to database but parse only
	APIOpts    map[string]any
	Caching    *string
}

type AttrImportTPFromFolder struct {
	TPid         string
	FolderPath   string
	RunId        string
	CsvSeparator string
	APIOpts      map[string]any
}

type AttrDirExportTP struct {
	TPid           *string
	FileFormat     *string // Format of the exported file <csv>
	FieldSeparator *string // Separator used between fields
	ExportPath     *string // If provided it overwrites the configured export path
	Compress       *bool   // If true the folder will be compressed after export performed
}

type ExportedTPStats struct {
	ExportPath    string   // Full path to the newly generated export file
	ExportedFiles []string // List of exported files
	Compressed    bool
}

// RPCCDRsFilter is a filter used in Rpc calls
// RPCCDRsFilter is slightly different than CDRsFilter by using string instead of Time filters
type RPCCDRsFilter struct {
	//CGRIDs                 []string               // If provided, it will filter based on the cgrids present in list
	//NotCGRIDs              []string               // Filter specific CgrIds out
	RunIDs                 []string          // If provided, it will filter on mediation runid
	NotRunIDs              []string          // Filter specific runIds out
	OriginIDs              []string          // If provided, it will filter on OriginIDs
	NotOriginIDs           []string          // Filter specific OriginIDs out
	OriginHosts            []string          // If provided, it will filter cdrhost
	NotOriginHosts         []string          // Filter out specific cdr hosts
	Sources                []string          // If provided, it will filter cdrsource
	NotSources             []string          // Filter out specific CDR sources
	ToRs                   []string          // If provided, filter on TypeOfRecord
	NotToRs                []string          // Filter specific TORs out
	RequestTypes           []string          // If provided, it will fiter reqtype
	NotRequestTypes        []string          // Filter out specific request types
	Tenants                []string          // If provided, it will filter tenant
	NotTenants             []string          // If provided, it will filter tenant
	Categories             []string          // If provided, it will filter çategory
	NotCategories          []string          // Filter out specific categories
	Accounts               []string          // If provided, it will filter account
	NotAccounts            []string          // Filter out specific Accounts
	Subjects               []string          // If provided, it will filter the rating subject
	NotSubjects            []string          // Filter out specific subjects
	DestinationPrefixes    []string          // If provided, it will filter on destination prefix
	NotDestinationPrefixes []string          // Filter out specific destination prefixes
	Costs                  []float64         // Query based on costs specified
	NotCosts               []float64         // Filter out specific costs out from result
	ExtraFields            map[string]string // Query based on extra fields content
	NotExtraFields         map[string]string // Filter out based on extra fields content
	SetupTimeStart         string            // Start of interval, bigger or equal than configured
	SetupTimeEnd           string            // End interval, smaller than setupTime
	AnswerTimeStart        string            // Start of interval, bigger or equal than configured
	AnswerTimeEnd          string            // End interval, smaller than answerTime
	CreatedAtStart         string            // Start of interval, bigger or equal than configured
	CreatedAtEnd           string            // End interval, smaller than
	UpdatedAtStart         string            // Start of interval, bigger or equal than configured
	UpdatedAtEnd           string            // End interval, smaller than
	MinUsage               string            // Start of the usage interval (>=)
	MaxUsage               string            // End of the usage interval (<)
	OrderBy                string            // Ascendent/Descendent
	ExtraArgs              map[string]any    // it will contain optional arguments like: OrderIDStart,OrderIDEnd,MinCost and MaxCost
	Paginator                                // Add pagination
}

// TPResourceProfile is used in APIs to manage remotely offline ResourceProfile
type TPResourceProfile struct {
	TPid              string
	Tenant            string
	ID                string // Identifier of this limit
	FilterIDs         []string
	UsageTTL          string
	Limit             string // Limit value
	AllocationMessage string
	Blocker           bool // blocker flag to stop processing on filters matched
	Stored            bool
	Weights           string   // Weight to sort the ResourceLimits
	ThresholdIDs      []string // Thresholds to check after changing Limit
}
type TPIPPool struct {
	ID        string
	FilterIDs []string
	Type      string
	Range     string
	Strategy  string
	Message   string
	Weights   string
	Blockers  string
}

type TPIPProfile struct {
	TPid      string
	Tenant    string
	ID        string
	FilterIDs []string
	TTL       string
	Stored    bool
	Weights   string
	Pools     []*TPIPPool
}
type ArgsComputeFilterIndexIDs struct {
	Tenant           string
	APIOpts          map[string]any
	AttributeIDs     []string
	ResourceIDs      []string
	IPIDs            []string
	StatIDs          []string
	RouteIDs         []string
	ThresholdIDs     []string
	ChargerIDs       []string
	RateProfileIDs   []string
	AccountIDs       []string
	ActionProfileIDs []string
}

type ArgsComputeFilterIndexes struct {
	Tenant     string
	APIOpts    map[string]any
	AttributeS bool
	ResourceS  bool
	IPs        bool
	StatS      bool
	RouteS     bool
	ThresholdS bool
	ChargerS   bool
	RateS      bool
	AccountS   bool
	ActionS    bool
}

// TPActivationInterval represents an activation interval for an item
type TPActivationInterval struct {
	ActivationTime string
	ExpiryTime     string
}

// AsActivationTime converts TPActivationInterval into ActivationInterval
func (tpAI *TPActivationInterval) AsActivationInterval(timezone string) (ai *ActivationInterval, err error) {
	var at, et time.Time
	if at, err = ParseTimeDetectLayout(tpAI.ActivationTime, timezone); err != nil {
		return
	}
	if et, err = ParseTimeDetectLayout(tpAI.ExpiryTime, timezone); err != nil {
		return
	}
	return &ActivationInterval{ActivationTime: at, ExpiryTime: et}, nil
}

type ActivationInterval struct {
	ActivationTime time.Time
	ExpiryTime     time.Time
}

func (ai *ActivationInterval) IsActiveAtTime(atTime time.Time) bool {
	return (ai.ActivationTime.IsZero() || ai.ActivationTime.Before(atTime)) &&
		(ai.ExpiryTime.IsZero() || ai.ExpiryTime.After(atTime))
}

func (aI *ActivationInterval) Equals(actInt *ActivationInterval) (eq bool) {
	if aI.ActivationTime.IsZero() && !actInt.ActivationTime.IsZero() ||
		!aI.ActivationTime.IsZero() && actInt.ActivationTime.IsZero() ||
		aI.ExpiryTime.IsZero() && !actInt.ExpiryTime.IsZero() ||
		!aI.ExpiryTime.IsZero() && actInt.ExpiryTime.IsZero() {
		return
	}
	return aI.ActivationTime.Equal(actInt.ActivationTime) &&
		aI.ExpiryTime.Equal(actInt.ExpiryTime)
}

// Attributes to send on SessionDisconnect by SMG
type AttrDisconnectSession struct {
	EventStart map[string]any
	Reason     string
}

// MetricWithFilters is used in TPStatProfile
type MetricWithFilters struct {
	FilterIDs []string
	MetricID  string
	Blockers  string
}

// TPStatProfile is used in APIs to manage remotely offline StatProfile
type TPStatProfile struct {
	TPid         string
	Tenant       string
	ID           string
	FilterIDs    []string
	QueueLength  int
	TTL          string
	MinItems     int
	Weights      string
	Blockers     string // blocker flag to stop processing on filters matched
	Stored       bool
	ThresholdIDs []string
	Metrics      []*MetricWithFilters
}

// TPRankingProfile is used in APIs to manage remotely offline RankingProfile
type TPRankingProfile struct {
	TPid              string
	Tenant            string
	ID                string
	Schedule          string
	StatIDs           []string
	MetricIDs         []string
	Sorting           string
	SortingParameters []string
	Stored            bool
	ThresholdIDs      []string
}

// TPTrendProfile is used in APIs to manage remotely offline TrendProfile
type TPTrendsProfile struct {
	TPid            string
	Tenant          string
	ID              string
	Schedule        string
	StatID          string
	Metrics         []string
	TTL             string
	QueueLength     int
	MinItems        int
	CorrelationType string
	Tolerance       float64
	Stored          bool
	ThresholdIDs    []string
}

// TPThresholdProfile is used in APIs to manage remotely offline ThresholdProfile
type TPThresholdProfile struct {
	TPid             string
	Tenant           string
	ID               string
	FilterIDs        []string
	MaxHits          int
	MinHits          int
	MinSleep         string
	Blocker          bool   // blocker flag to stop processing on filters matched
	Weights          string // Weight to sort the thresholds
	ActionProfileIDs []string
	EeIDs            []string
	Async            bool
}

// TPFilterProfile is used in APIs to manage remotely offline FilterProfile
type TPFilterProfile struct {
	TPid    string
	Tenant  string
	ID      string
	Filters []*TPFilter
}

// TPFilter is used in TPFilterProfile
type TPFilter struct {
	Type    string   // Filter type (*string, *rsr_filters, *cdr_stats)
	Element string   // Name of the field providing us the Values to check (used in case of some )
	Values  []string // Filter definition
}

// TPRoute is used in TPRouteProfile
type TPRoute struct {
	ID              string // RouteID
	FilterIDs       []string
	AccountIDs      []string
	RateProfileIDs  []string // used when computing price
	ResourceIDs     []string // queried in some strategies
	StatIDs         []string // queried in some strategies
	Weights         string
	Blockers        string
	RouteParameters string
}

// TPRouteProfile is used in APIs to manage remotely offline RouteProfile
type TPRouteProfile struct {
	TPid              string
	Tenant            string
	ID                string
	FilterIDs         []string
	Weights           string
	Blockers          string
	Sorting           string
	SortingParameters []string
	Routes            []*TPRoute
}

// TPAttribute is used in TPAttributeProfile
type TPAttribute struct {
	FilterIDs []string
	Blockers  string
	Path      string
	Type      string
	Value     string
}

// TPAttributeProfile is used in APIs to manage remotely offline AttributeProfile
type TPAttributeProfile struct {
	TPid       string
	Tenant     string
	ID         string
	FilterIDs  []string
	Weights    string
	Blockers   string
	Attributes []*TPAttribute
}

// TPChargerProfile is used in APIs to manage remotely offline ChargerProfile
type TPChargerProfile struct {
	TPid         string
	Tenant       string
	ID           string
	FilterIDs    []string
	Weights      string
	Blockers     string
	RunID        string
	AttributeIDs []string
}

type TPTntID struct {
	TPid   string
	Tenant string
	ID     string
}

type AttrRemoteLock struct {
	ReferenceID string        // reference ID for this lock if available
	LockIDs     []string      // List of IDs to obtain lock for
	Timeout     time.Duration // Automatically unlock on timeout
}

type RPCCDRsFilterWithAPIOpts struct {
	*RPCCDRsFilter
	APIOpts map[string]any
	Tenant  string
}

type ArgsGetCacheItemIDsWithAPIOpts struct {
	APIOpts map[string]any
	Tenant  string
	ArgsGetCacheItemIDs
}

type ArgsGetCacheItemWithAPIOpts struct {
	APIOpts map[string]any
	Tenant  string
	ArgsGetCacheItem
}

// NewAttrReloadCacheWithOpts returns the ArgCache populated with nil
func NewAttrReloadCacheWithOpts() *AttrReloadCacheWithAPIOpts {
	return &AttrReloadCacheWithAPIOpts{
		ResourceProfileIDs:   []string{MetaAny},
		ResourceIDs:          []string{MetaAny},
		IPProfileIDs:         []string{MetaAny},
		IPIDs:                []string{MetaAny},
		StatsQueueIDs:        []string{MetaAny},
		StatsQueueProfileIDs: []string{MetaAny},
		ThresholdIDs:         []string{MetaAny},
		ThresholdProfileIDs:  []string{MetaAny},
		TrendIDs:             []string{MetaAny},
		RankingIDs:           []string{MetaAny},
		TrendProfileIDs:      []string{MetaAny},
		FilterIDs:            []string{MetaAny},
		RouteProfileIDs:      []string{MetaAny},
		AttributeProfileIDs:  []string{MetaAny},
		ChargerProfileIDs:    []string{MetaAny},
		RateProfileIDs:       []string{MetaAny},
		ActionProfileIDs:     []string{MetaAny},
		AccountIDs:           []string{MetaAny},
		RankingProfileIDs:    []string{MetaAny},

		AttributeFilterIndexIDs:      []string{MetaAny},
		ResourceFilterIndexIDs:       []string{MetaAny},
		IPFilterIndexIDs:             []string{MetaAny},
		StatFilterIndexIDs:           []string{MetaAny},
		ThresholdFilterIndexIDs:      []string{MetaAny},
		RouteFilterIndexIDs:          []string{MetaAny},
		ChargerFilterIndexIDs:        []string{MetaAny},
		RateProfilesFilterIndexIDs:   []string{MetaAny},
		RateFilterIndexIDs:           []string{MetaAny},
		FilterIndexIDs:               []string{MetaAny},
		ActionProfilesFilterIndexIDs: []string{MetaAny},
		AccountsFilterIndexIDs:       []string{MetaAny},
	}
}

func NewAttrReloadCacheWithOptsFromMap(arg map[string][]string, tnt string, opts map[string]any) *AttrReloadCacheWithAPIOpts {
	return &AttrReloadCacheWithAPIOpts{
		Tenant:  tnt,
		APIOpts: opts,

		ResourceProfileIDs:           arg[CacheResourceProfiles],
		ResourceIDs:                  arg[CacheResources],
		IPProfileIDs:                 arg[CacheIPProfiles],
		IPIDs:                        arg[CacheIPAllocations],
		StatsQueueProfileIDs:         arg[CacheStatQueueProfiles],
		StatsQueueIDs:                arg[CacheStatQueues],
		ThresholdProfileIDs:          arg[CacheThresholdProfiles],
		ThresholdIDs:                 arg[CacheThresholds],
		RankingProfileIDs:            arg[CacheRankingProfiles],
		RankingIDs:                   arg[CacheRankings],
		FilterIDs:                    arg[CacheFilters],
		RouteProfileIDs:              arg[CacheRouteProfiles],
		AttributeProfileIDs:          arg[CacheAttributeProfiles],
		ChargerProfileIDs:            arg[CacheChargerProfiles],
		RateProfileIDs:               arg[CacheRateProfiles],
		ActionProfileIDs:             arg[CacheActionProfiles],
		AccountIDs:                   arg[CacheAccounts],
		ResourceFilterIndexIDs:       arg[CacheResourceFilterIndexes],
		IPFilterIndexIDs:             arg[CacheIPFilterIndexes],
		StatFilterIndexIDs:           arg[CacheStatFilterIndexes],
		ThresholdFilterIndexIDs:      arg[CacheThresholdFilterIndexes],
		RouteFilterIndexIDs:          arg[CacheRouteFilterIndexes],
		AttributeFilterIndexIDs:      arg[CacheAttributeFilterIndexes],
		ChargerFilterIndexIDs:        arg[CacheChargerFilterIndexes],
		RateProfilesFilterIndexIDs:   arg[CacheRateProfilesFilterIndexes],
		ActionProfilesFilterIndexIDs: arg[CacheActionProfilesFilterIndexes],
		AccountsFilterIndexIDs:       arg[CacheAccountsFilterIndexes],
		RateFilterIndexIDs:           arg[CacheRateFilterIndexes],
		FilterIndexIDs:               arg[CacheReverseFilterIndexes],
		TrendProfileIDs:              arg[CacheTrendProfiles],
		TrendIDs:                     arg[CacheTrends],
	}
}

type AttrReloadCacheWithAPIOpts struct {
	APIOpts map[string]any `json:",omitempty"`
	Tenant  string         `json:",omitempty"`

	ResourceProfileIDs   []string `json:",omitempty"`
	ResourceIDs          []string `json:",omitempty"`
	IPProfileIDs         []string `json:",omitempty"`
	IPIDs                []string `json:",omitempty"`
	StatsQueueIDs        []string `json:",omitempty"`
	StatsQueueProfileIDs []string `json:",omitempty"`
	ThresholdIDs         []string `json:",omitempty"`
	ThresholdProfileIDs  []string `json:",omitempty"`
	TrendIDs             []string `json:",omitempty"`
	TrendProfileIDs      []string `json:",omitempty"`
	RankingProfileIDs    []string `json:",omitempty"`
	RankingIDs           []string `json:",omitempty"`
	FilterIDs            []string `json:",omitempty"`
	RouteProfileIDs      []string `json:",omitempty"`
	AttributeProfileIDs  []string `json:",omitempty"`
	ChargerProfileIDs    []string `json:",omitempty"`
	RateProfileIDs       []string `json:",omitempty"`
	ActionProfileIDs     []string `json:",omitempty"`
	AccountIDs           []string `json:",omitempty"`

	AttributeFilterIndexIDs      []string `json:",omitempty"`
	ResourceFilterIndexIDs       []string `json:",omitempty"`
	IPFilterIndexIDs             []string `json:",omitempty"`
	StatFilterIndexIDs           []string `json:",omitempty"`
	ThresholdFilterIndexIDs      []string `json:",omitempty"`
	RouteFilterIndexIDs          []string `json:",omitempty"`
	ChargerFilterIndexIDs        []string `json:",omitempty"`
	RateProfilesFilterIndexIDs   []string `json:",omitempty"`
	RateFilterIndexIDs           []string `json:",omitempty"`
	FilterIndexIDs               []string `json:",omitempty"`
	ActionProfilesFilterIndexIDs []string `json:",omitempty"`
	AccountsFilterIndexIDs       []string `json:",omitempty"`
}

func (a *AttrReloadCacheWithAPIOpts) Map() map[string][]string {
	return map[string][]string{
		CacheResourceProfiles:            a.ResourceProfileIDs,
		CacheResources:                   a.ResourceIDs,
		CacheIPProfiles:                  a.IPProfileIDs,
		CacheIPAllocations:               a.IPIDs,
		CacheStatQueueProfiles:           a.StatsQueueProfileIDs,
		CacheStatQueues:                  a.StatsQueueIDs,
		CacheThresholdProfiles:           a.ThresholdProfileIDs,
		CacheThresholds:                  a.ThresholdIDs,
		CacheTrendProfiles:               a.TrendProfileIDs,
		CacheTrends:                      a.TrendIDs,
		CacheFilters:                     a.FilterIDs,
		CacheRouteProfiles:               a.RouteProfileIDs,
		CacheAttributeProfiles:           a.AttributeProfileIDs,
		CacheChargerProfiles:             a.ChargerProfileIDs,
		CacheRankingProfiles:             a.RankingProfileIDs,
		CacheRankings:                    a.RankingIDs,
		CacheRateProfiles:                a.RateProfileIDs,
		CacheActionProfiles:              a.ActionProfileIDs,
		CacheAccounts:                    a.AccountIDs,
		CacheResourceFilterIndexes:       a.ResourceFilterIndexIDs,
		CacheIPFilterIndexes:             a.IPFilterIndexIDs,
		CacheStatFilterIndexes:           a.StatFilterIndexIDs,
		CacheThresholdFilterIndexes:      a.ThresholdFilterIndexIDs,
		CacheRouteFilterIndexes:          a.RouteFilterIndexIDs,
		CacheAttributeFilterIndexes:      a.AttributeFilterIndexIDs,
		CacheChargerFilterIndexes:        a.ChargerFilterIndexIDs,
		CacheRateProfilesFilterIndexes:   a.RateProfilesFilterIndexIDs,
		CacheActionProfilesFilterIndexes: a.ActionProfilesFilterIndexIDs,
		CacheAccountsFilterIndexes:       a.AccountsFilterIndexIDs,
		CacheRateFilterIndexes:           a.RateFilterIndexIDs,
		CacheReverseFilterIndexes:        a.FilterIndexIDs,
	}
}

type AttrCacheIDsWithAPIOpts struct {
	APIOpts  map[string]any
	Tenant   string
	CacheIDs []string
}

type ArgsGetGroupWithAPIOpts struct {
	APIOpts map[string]any
	Tenant  string
	ArgsGetGroup
}

type ArgsGetCacheItemIDs struct {
	CacheID      string
	ItemIDPrefix string
}

type ArgsGetCacheItem struct {
	CacheID string
	ItemID  string
}

type ArgsGetGroup struct {
	CacheID string
	GroupID string
}

type SessionFilter struct {
	Limit   *int
	Filters []string
	Tenant  string
	APIOpts map[string]any
}

type SessionFilterWithEvent struct {
	*SessionFilter
	Event map[string]any
}

type SessionIDsWithAPIOpts struct {
	IDs     []string
	Tenant  string
	APIOpts map[string]any
}

type ArgExportToFolder struct {
	Path  string
	Items []string
}

// DPRArgs are the arguments used by dispatcher to send a Disconnect-Peer-Request
type DPRArgs struct {
	OriginHost      string
	OriginRealm     string
	RemoteAddr      string
	DisconnectCause int
}

type ArgCacheReplicateSet struct {
	Tenant   string
	APIOpts  map[string]any
	CacheID  string
	ItemID   string
	Value    any
	GroupIDs []string
}

// Compiler are objects that need post compiling
type Compiler interface {
	Compile() error
}

type ArgCacheReplicateRemove struct {
	CacheID string
	ItemID  string
	APIOpts map[string]any
	Tenant  string
}

type TPRateProfile struct {
	TPid            string
	Tenant          string
	ID              string
	FilterIDs       []string
	Weights         string
	MinCost         float64
	MaxCost         float64
	MaxCostStrategy string
	Rates           map[string]*TPRate
}

type TPRate struct {
	ID              string   // RateID
	FilterIDs       []string // RateFilterIDs
	ActivationTimes string
	Weights         string // RateWeights will decide the winner per interval start
	Blocker         bool   // RateBlocker will make this rate recurrent, deactivating further intervals
	IntervalRates   []*TPIntervalRate
}

type TPIntervalRate struct {
	IntervalStart string
	FixedFee      float64
	RecurrentFee  float64 // RateValue
	Unit          string
	Increment     string
}

type ArgExportCDRs struct {
	ExporterIDs []string // exporterIDs is used to said which exporter are using to export the cdrs
	Verbose     bool     // verbose is used to inform the user about the positive and negative exported cdrs
	RPCCDRsFilter
}

type TPActionProfile struct {
	TPid      string
	Tenant    string
	ID        string
	FilterIDs []string
	Weights   string
	Blockers  string
	Schedule  string
	Targets   []*TPActionTarget
	Actions   []*TPAPAction
}

type TPActionTarget struct {
	TargetType string
	TargetIDs  []string
}

type TPAPAction struct {
	ID        string
	FilterIDs []string
	TTL       string
	Type      string
	Opts      string
	Weights   string
	Blockers  string
	Diktats   []*TPAPDiktat
}

type TPAPDiktat struct {
	ID        string
	FilterIDs []string
	Opts      string
	Weights   string
	Blockers  string
}

type TPAccount struct {
	TPid         string
	Tenant       string
	ID           string
	FilterIDs    []string
	Weights      string
	Blockers     string
	Balances     map[string]*TPAccountBalance
	ThresholdIDs []string
}

type TPAccountBalance struct {
	ID             string
	FilterIDs      []string
	Weights        string
	Blockers       string
	Type           string
	Opts           string
	CostIncrement  []*TPBalanceCostIncrement
	AttributeIDs   []string
	RateProfileIDs []string
	UnitFactors    []*TPBalanceUnitFactor
	Units          string
}

func NewTPBalanceCostIncrement(filtersStr, incrementStr, fixedFeeStr, recurrentFeeStr string) (costIncrement *TPBalanceCostIncrement, err error) {
	costIncrement = &TPBalanceCostIncrement{
		Increment: incrementStr,
	}
	if filtersStr != EmptyString {
		costIncrement.FilterIDs = strings.Split(filtersStr, ANDSep)
	}
	if fixedFeeStr != EmptyString {
		fixedFee, err := strconv.ParseFloat(fixedFeeStr, 64)
		if err != nil {
			return nil, err
		}
		costIncrement.FixedFee = Float64Pointer(fixedFee)
	}
	if recurrentFeeStr != EmptyString {
		recFee, err := strconv.ParseFloat(recurrentFeeStr, 64)
		if err != nil {
			return nil, err
		}
		costIncrement.RecurrentFee = Float64Pointer(recFee)
	}
	return
}

type TPBalanceCostIncrement struct {
	FilterIDs    []string
	Increment    string
	FixedFee     *float64
	RecurrentFee *float64
}

func (costIncr *TPBalanceCostIncrement) AsString() (s string) {
	if len(costIncr.FilterIDs) != 0 {
		s += strings.Join(costIncr.FilterIDs, ANDSep)
	}
	s += InfieldSep
	if costIncr.Increment != EmptyString {
		s += costIncr.Increment
	}
	s += InfieldSep
	if costIncr.FixedFee != nil {
		s += strconv.FormatFloat(*costIncr.FixedFee, 'f', -1, 64)
	}
	s += InfieldSep
	if costIncr.RecurrentFee != nil {
		s += strconv.FormatFloat(*costIncr.RecurrentFee, 'f', -1, 64)
	}
	return
}

func NewTPBalanceUnitFactor(filtersStr, factorStr string) (unitFactor *TPBalanceUnitFactor, err error) {
	unitFactor = &TPBalanceUnitFactor{
		FilterIDs: strings.Split(filtersStr, ANDSep),
	}
	unitFactor.Factor, err = strconv.ParseFloat(factorStr, 64)
	return
}

type TPBalanceUnitFactor struct {
	FilterIDs []string
	Factor    float64
}

func (unitFactor *TPBalanceUnitFactor) AsString() string {
	return strings.Join(unitFactor.FilterIDs, ANDSep) + InfieldSep + strconv.FormatFloat(unitFactor.Factor, 'f', -1, 64)
}

type ArgScheduleTrendQueries struct {
	TenantIDWithAPIOpts
	TrendIDs []string
}
type ArgScheduledTrends struct {
	TenantIDWithAPIOpts
	TrendIDPrefixes []string
}

type ArgGetTrend struct {
	TenantWithAPIOpts
	ID            string
	RunIndexStart int
	RunIndexEnd   int
	RunTimeStart  string
	RunTimeEnd    string
}

type ScheduledTrend struct {
	TrendID  string
	Next     time.Time
	Previous time.Time
}

type ArgScheduleRankingQueries struct {
	TenantIDWithAPIOpts
	RankingIDs []string
}

type ArgScheduledRankings struct {
	TenantIDWithAPIOpts
	RankingIDPrefixes []string
}

type ScheduledRanking struct {
	RankingID string
	Next      time.Time
	Previous  time.Time
}
