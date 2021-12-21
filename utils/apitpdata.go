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
	"sort"
	"strconv"
	"strings"
	"time"
)

type PaginatorWithSearch struct {
	*Paginator
	Search string // Global matching pattern in items returned, partially used in some APIs
}

// Paginate stuff around items returned
type Paginator struct {
	Limit  *int // Limit the number of items returned
	Offset *int // Offset of the first item returned (eg: use Limit*Page in case of PerPage items)
}

func (pgnt *Paginator) PaginateStringSlice(in []string) (out []string) {
	if len(in) == 0 {
		return
	}
	var limit, offset int
	if pgnt.Limit != nil && *pgnt.Limit > 0 {
		limit = *pgnt.Limit
	}
	if pgnt.Offset != nil && *pgnt.Offset > 0 {
		offset = *pgnt.Offset
	}
	if limit == 0 && offset == 0 {
		return in
	}
	if offset > len(in) {
		return
	}
	if offset != 0 && limit != 0 {
		limit = limit + offset
	}
	if limit == 0 || limit > len(in) {
		limit = len(in)
	}
	return CloneStringSlice(in[offset:limit])
}

// Clone creates a clone of the object
func (pgnt Paginator) Clone() Paginator {
	var limit *int
	if pgnt.Limit != nil {
		limit = new(int)
		*limit = *pgnt.Limit
	}

	var offset *int
	if pgnt.Offset != nil {
		offset = new(int)
		*offset = *pgnt.Offset
	}
	return Paginator{
		Limit:  limit,
		Offset: offset,
	}
}

type AttrGetCdrs struct {
	CgrIds          []string // If provided, it will filter based on the cgrids present in list
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

func (fltr *AttrGetCdrs) AsCDRsFilter(timezone string) (cdrFltr *CDRsFilter, err error) {
	if fltr == nil {
		return
	}
	cdrFltr = &CDRsFilter{
		CGRIDs:              fltr.CgrIds,
		RunIDs:              fltr.MediationRunIds,
		ToRs:                fltr.TORs,
		OriginHosts:         fltr.CdrHosts,
		Sources:             fltr.CdrSources,
		RequestTypes:        fltr.ReqTypes,
		Tenants:             fltr.Tenants,
		Categories:          fltr.Categories,
		Accounts:            fltr.Accounts,
		Subjects:            fltr.Subjects,
		DestinationPrefixes: fltr.DestinationPrefixes,
		OrderIDStart:        fltr.OrderIdStart,
		OrderIDEnd:          fltr.OrderIdEnd,
		Paginator:           fltr.Paginator,
		OrderBy:             fltr.OrderBy,
	}
	if len(fltr.TimeStart) != 0 {
		if answerTimeStart, err := ParseTimeDetectLayout(fltr.TimeStart, timezone); err != nil {
			return nil, err
		} else {
			cdrFltr.AnswerTimeStart = &answerTimeStart
		}
	}
	if len(fltr.TimeEnd) != 0 {
		if answerTimeEnd, err := ParseTimeDetectLayout(fltr.TimeEnd, timezone); err != nil {
			return nil, err
		} else {
			cdrFltr.AnswerTimeEnd = &answerTimeEnd
		}
	}
	if fltr.SkipRated {
		cdrFltr.MaxCost = Float64Pointer(-1.0)
	} else if fltr.SkipErrors {
		cdrFltr.MinCost = Float64Pointer(0.0)
		cdrFltr.MaxCost = Float64Pointer(-1.0)
	}
	return
}

type AttrLoadTpFromFolder struct {
	FolderPath string // Take files from folder absolute path
	DryRun     bool   // Do not write to database but parse only
	APIOpts    map[string]interface{}
	Caching    *string
}

type AttrImportTPFromFolder struct {
	TPid         string
	FolderPath   string
	RunId        string
	CsvSeparator string
	APIOpts      map[string]interface{}
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

// CDRsFilter is a filter used to get records out of storDB
type CDRsFilter struct {
	CGRIDs                 []string          // If provided, it will filter based on the cgrids present in list
	NotCGRIDs              []string          // Filter specific CgrIds out
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
	OrderIDStart           *int64            // Export from this order identifier
	OrderIDEnd             *int64            // Export smaller than this order identifier
	SetupTimeStart         *time.Time        // Start of interval, bigger or equal than configured
	SetupTimeEnd           *time.Time        // End interval, smaller than setupTime
	AnswerTimeStart        *time.Time        // Start of interval, bigger or equal than configured
	AnswerTimeEnd          *time.Time        // End interval, smaller than answerTime
	CreatedAtStart         *time.Time        // Start of interval, bigger or equal than configured
	CreatedAtEnd           *time.Time        // End interval, smaller than
	UpdatedAtStart         *time.Time        // Start of interval, bigger or equal than configured
	UpdatedAtEnd           *time.Time        // End interval, smaller than
	MinUsage               string            // Start of the usage interval (>=)
	MaxUsage               string            // End of the usage interval (<)
	MinCost                *float64          // Start of the cost interval (>=)
	MaxCost                *float64          // End of the usage interval (<)
	Unscoped               bool              // Include soft-deleted records in results
	Count                  bool              // If true count the items instead of returning data
	OrderBy                string            // Can be ordered by OrderID,AnswerTime,SetupTime,Cost,Usage
	Paginator
}

// Prepare will sort all the slices in order to search more faster
func (fltr *CDRsFilter) Prepare() {
	sort.Strings(fltr.CGRIDs)
	sort.Strings(fltr.NotCGRIDs)
	sort.Strings(fltr.RunIDs)
	sort.Strings(fltr.NotRunIDs)
	sort.Strings(fltr.OriginIDs)
	sort.Strings(fltr.NotOriginIDs)
	sort.Strings(fltr.OriginHosts)
	sort.Strings(fltr.NotOriginHosts)
	sort.Strings(fltr.Sources)
	sort.Strings(fltr.NotSources)
	sort.Strings(fltr.ToRs)
	sort.Strings(fltr.NotToRs)
	sort.Strings(fltr.RequestTypes)
	sort.Strings(fltr.NotRequestTypes)
	sort.Strings(fltr.Tenants)
	sort.Strings(fltr.NotTenants)
	sort.Strings(fltr.Categories)
	sort.Strings(fltr.NotCategories)
	sort.Strings(fltr.Accounts)
	sort.Strings(fltr.NotAccounts)
	sort.Strings(fltr.Subjects)
	sort.Strings(fltr.NotSubjects)
	// sort.Strings(fltr.DestinationPrefixes)
	// sort.Strings(fltr.NotDestinationPrefixes)

	sort.Float64s(fltr.Costs)
	sort.Float64s(fltr.NotCosts)

}

// RPCCDRsFilter is a filter used in Rpc calls
// RPCCDRsFilter is slightly different than CDRsFilter by using string instead of Time filters
type RPCCDRsFilter struct {
	CGRIDs                 []string               // If provided, it will filter based on the cgrids present in list
	NotCGRIDs              []string               // Filter specific CgrIds out
	RunIDs                 []string               // If provided, it will filter on mediation runid
	NotRunIDs              []string               // Filter specific runIds out
	OriginIDs              []string               // If provided, it will filter on OriginIDs
	NotOriginIDs           []string               // Filter specific OriginIDs out
	OriginHosts            []string               // If provided, it will filter cdrhost
	NotOriginHosts         []string               // Filter out specific cdr hosts
	Sources                []string               // If provided, it will filter cdrsource
	NotSources             []string               // Filter out specific CDR sources
	ToRs                   []string               // If provided, filter on TypeOfRecord
	NotToRs                []string               // Filter specific TORs out
	RequestTypes           []string               // If provided, it will fiter reqtype
	NotRequestTypes        []string               // Filter out specific request types
	Tenants                []string               // If provided, it will filter tenant
	NotTenants             []string               // If provided, it will filter tenant
	Categories             []string               // If provided, it will filter çategory
	NotCategories          []string               // Filter out specific categories
	Accounts               []string               // If provided, it will filter account
	NotAccounts            []string               // Filter out specific Accounts
	Subjects               []string               // If provided, it will filter the rating subject
	NotSubjects            []string               // Filter out specific subjects
	DestinationPrefixes    []string               // If provided, it will filter on destination prefix
	NotDestinationPrefixes []string               // Filter out specific destination prefixes
	Costs                  []float64              // Query based on costs specified
	NotCosts               []float64              // Filter out specific costs out from result
	ExtraFields            map[string]string      // Query based on extra fields content
	NotExtraFields         map[string]string      // Filter out based on extra fields content
	SetupTimeStart         string                 // Start of interval, bigger or equal than configured
	SetupTimeEnd           string                 // End interval, smaller than setupTime
	AnswerTimeStart        string                 // Start of interval, bigger or equal than configured
	AnswerTimeEnd          string                 // End interval, smaller than answerTime
	CreatedAtStart         string                 // Start of interval, bigger or equal than configured
	CreatedAtEnd           string                 // End interval, smaller than
	UpdatedAtStart         string                 // Start of interval, bigger or equal than configured
	UpdatedAtEnd           string                 // End interval, smaller than
	MinUsage               string                 // Start of the usage interval (>=)
	MaxUsage               string                 // End of the usage interval (<)
	OrderBy                string                 // Ascendent/Descendent
	ExtraArgs              map[string]interface{} // it will contain optional arguments like: OrderIDStart,OrderIDEnd,MinCost and MaxCost
	Paginator                                     // Add pagination
}

func (fltr *RPCCDRsFilter) AsCDRsFilter(timezone string) (cdrFltr *CDRsFilter, err error) {
	if fltr == nil {
		cdrFltr = new(CDRsFilter)
		return
	}
	cdrFltr = &CDRsFilter{
		CGRIDs:                 fltr.CGRIDs,
		NotCGRIDs:              fltr.NotCGRIDs,
		RunIDs:                 fltr.RunIDs,
		NotRunIDs:              fltr.NotRunIDs,
		OriginIDs:              fltr.OriginIDs,
		NotOriginIDs:           fltr.NotOriginIDs,
		ToRs:                   fltr.ToRs,
		NotToRs:                fltr.NotToRs,
		OriginHosts:            fltr.OriginHosts,
		NotOriginHosts:         fltr.NotOriginHosts,
		Sources:                fltr.Sources,
		NotSources:             fltr.NotSources,
		RequestTypes:           fltr.RequestTypes,
		NotRequestTypes:        fltr.NotRequestTypes,
		Tenants:                fltr.Tenants,
		NotTenants:             fltr.NotTenants,
		Categories:             fltr.Categories,
		NotCategories:          fltr.NotCategories,
		Accounts:               fltr.Accounts,
		NotAccounts:            fltr.NotAccounts,
		Subjects:               fltr.Subjects,
		NotSubjects:            fltr.NotSubjects,
		DestinationPrefixes:    fltr.DestinationPrefixes,
		NotDestinationPrefixes: fltr.NotDestinationPrefixes,
		Costs:                  fltr.Costs,
		NotCosts:               fltr.NotCosts,
		ExtraFields:            fltr.ExtraFields,
		NotExtraFields:         fltr.NotExtraFields,
		MinUsage:               fltr.MinUsage,
		MaxUsage:               fltr.MaxUsage,
		Paginator:              fltr.Paginator,
		OrderBy:                fltr.OrderBy,
	}
	if len(fltr.SetupTimeStart) != 0 {
		var sTimeStart time.Time
		if sTimeStart, err = ParseTimeDetectLayout(fltr.SetupTimeStart, timezone); err != nil {
			return
		}
		cdrFltr.SetupTimeStart = TimePointer(sTimeStart)
	}
	if len(fltr.SetupTimeEnd) != 0 {
		var sTimeEnd time.Time
		if sTimeEnd, err = ParseTimeDetectLayout(fltr.SetupTimeEnd, timezone); err != nil {
			return
		}
		cdrFltr.SetupTimeEnd = TimePointer(sTimeEnd)
	}
	if len(fltr.AnswerTimeStart) != 0 {
		var aTimeStart time.Time
		if aTimeStart, err = ParseTimeDetectLayout(fltr.AnswerTimeStart, timezone); err != nil {
			return
		}
		cdrFltr.AnswerTimeStart = TimePointer(aTimeStart)
	}
	if len(fltr.AnswerTimeEnd) != 0 {
		var aTimeEnd time.Time
		if aTimeEnd, err = ParseTimeDetectLayout(fltr.AnswerTimeEnd, timezone); err != nil {
			return
		}
		cdrFltr.AnswerTimeEnd = TimePointer(aTimeEnd)
	}
	if len(fltr.CreatedAtStart) != 0 {
		var tStart time.Time
		if tStart, err = ParseTimeDetectLayout(fltr.CreatedAtStart, timezone); err != nil {
			return
		}
		cdrFltr.CreatedAtStart = TimePointer(tStart)
	}
	if len(fltr.CreatedAtEnd) != 0 {
		var tEnd time.Time
		if tEnd, err = ParseTimeDetectLayout(fltr.CreatedAtEnd, timezone); err != nil {
			return
		}
		cdrFltr.CreatedAtEnd = TimePointer(tEnd)
	}
	if len(fltr.UpdatedAtStart) != 0 {
		var tStart time.Time
		if tStart, err = ParseTimeDetectLayout(fltr.UpdatedAtStart, timezone); err != nil {
			return
		}
		cdrFltr.UpdatedAtStart = TimePointer(tStart)
	}
	if len(fltr.UpdatedAtEnd) != 0 {
		var tEnd time.Time
		if tEnd, err = ParseTimeDetectLayout(fltr.UpdatedAtEnd, timezone); err != nil {
			return
		}
		cdrFltr.UpdatedAtEnd = TimePointer(tEnd)
	}
	if oIDstart, has := fltr.ExtraArgs[OrderIDStart]; has {
		var oID int64
		if oID, err = IfaceAsTInt64(oIDstart); err != nil {
			return
		}
		cdrFltr.OrderIDStart = Int64Pointer(oID)
	}
	if oIDend, has := fltr.ExtraArgs[OrderIDEnd]; has {
		var oID int64
		if oID, err = IfaceAsTInt64(oIDend); err != nil {
			return
		}
		cdrFltr.OrderIDEnd = Int64Pointer(oID)
	}
	if mcost, has := fltr.ExtraArgs[MinCost]; has {
		var mc float64
		if mc, err = IfaceAsFloat64(mcost); err != nil {
			return
		}
		cdrFltr.MinCost = Float64Pointer(mc)
	}
	if mcost, has := fltr.ExtraArgs[MaxCost]; has {
		var mc float64
		if mc, err = IfaceAsFloat64(mcost); err != nil {
			return
		}
		cdrFltr.MaxCost = Float64Pointer(mc)
	}
	return
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
	Weight            float64  // Weight to sort the ResourceLimits
	ThresholdIDs      []string // Thresholds to check after changing Limit
}

type ArgsComputeFilterIndexIDs struct {
	Tenant           string
	APIOpts          map[string]interface{}
	AttributeIDs     []string
	ResourceIDs      []string
	StatIDs          []string
	RouteIDs         []string
	ThresholdIDs     []string
	ChargerIDs       []string
	DispatcherIDs    []string
	RateProfileIDs   []string
	AccountIDs       []string
	ActionProfileIDs []string
}

type ArgsComputeFilterIndexes struct {
	Tenant      string
	APIOpts     map[string]interface{}
	AttributeS  bool
	ResourceS   bool
	StatS       bool
	RouteS      bool
	ThresholdS  bool
	ChargerS    bool
	DispatcherS bool
	RateS       bool
	AccountS    bool
	ActionS     bool
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
	EventStart map[string]interface{}
	Reason     string
}

//MetricWithFilters is used in TPStatProfile
type MetricWithFilters struct {
	FilterIDs []string
	MetricID  string
}

// TPStatProfile is used in APIs to manage remotely offline StatProfile
type TPStatProfile struct {
	TPid         string
	Tenant       string
	ID           string
	FilterIDs    []string
	QueueLength  int
	TTL          string
	Metrics      []*MetricWithFilters
	Blocker      bool // blocker flag to stop processing on filters matched
	Stored       bool
	Weight       float64
	MinItems     int
	ThresholdIDs []string
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
	Blocker          bool    // blocker flag to stop processing on filters matched
	Weight           float64 // Weight to sort the thresholds
	ActionProfileIDs []string
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
	Blocker         bool
	RouteParameters string
}

// TPRouteProfile is used in APIs to manage remotely offline RouteProfile
type TPRouteProfile struct {
	TPid              string
	Tenant            string
	ID                string
	FilterIDs         []string
	Sorting           string
	SortingParameters []string
	Routes            []*TPRoute
	Weights           string
}

// TPAttribute is used in TPAttributeProfile
type TPAttribute struct {
	FilterIDs []string
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
	Attributes []*TPAttribute
	Blocker    bool
	Weight     float64
}

// TPChargerProfile is used in APIs to manage remotely offline ChargerProfile
type TPChargerProfile struct {
	TPid         string
	Tenant       string
	ID           string
	FilterIDs    []string
	RunID        string
	AttributeIDs []string
	Weight       float64
}

type TPTntID struct {
	TPid   string
	Tenant string
	ID     string
}

// TPDispatcherProfile is used in APIs to manage remotely offline DispatcherProfile
type TPDispatcherProfile struct {
	TPid           string
	Tenant         string
	ID             string
	FilterIDs      []string
	Strategy       string
	StrategyParams []interface{} // ie for distribution, set here the pool weights
	Weight         float64
	Hosts          []*TPDispatcherHostProfile
}

// TPDispatcherHostProfile is used in TPDispatcherProfile
type TPDispatcherHostProfile struct {
	ID        string
	FilterIDs []string
	Weight    float64       // applied in case of multiple connections need to be ordered
	Params    []interface{} // additional parameters stored for a session
	Blocker   bool          // no connection after this one
}

// TPDispatcherHost is used in APIs to manage remotely offline DispatcherHost
type TPDispatcherHost struct {
	TPid   string
	Tenant string
	ID     string
	Conn   *TPDispatcherHostConn
}

// TPDispatcherHostConn is used in TPDispatcherHost
type TPDispatcherHostConn struct {
	Address           string
	Transport         string
	ConnectAttempts   int
	Reconnects        int
	ConnectTimeout    time.Duration
	ReplyTimeout      time.Duration
	TLS               bool
	ClientKey         string
	ClientCertificate string
	CaCertificate     string
}

type AttrRemoteLock struct {
	ReferenceID string        // reference ID for this lock if available
	LockIDs     []string      // List of IDs to obtain lock for
	Timeout     time.Duration // Automatically unlock on timeout
}

type RPCCDRsFilterWithAPIOpts struct {
	*RPCCDRsFilter
	APIOpts map[string]interface{}
	Tenant  string
}

type ArgsGetCacheItemIDsWithAPIOpts struct {
	APIOpts map[string]interface{}
	Tenant  string
	ArgsGetCacheItemIDs
}

type ArgsGetCacheItemWithAPIOpts struct {
	APIOpts map[string]interface{}
	Tenant  string
	ArgsGetCacheItem
}

// NewAttrReloadCacheWithOpts returns the ArgCache populated with nil
func NewAttrReloadCacheWithOpts() *AttrReloadCacheWithAPIOpts {
	return &AttrReloadCacheWithAPIOpts{
		ResourceProfileIDs:   []string{MetaAny},
		ResourceIDs:          []string{MetaAny},
		StatsQueueIDs:        []string{MetaAny},
		StatsQueueProfileIDs: []string{MetaAny},
		ThresholdIDs:         []string{MetaAny},
		ThresholdProfileIDs:  []string{MetaAny},
		FilterIDs:            []string{MetaAny},
		RouteProfileIDs:      []string{MetaAny},
		AttributeProfileIDs:  []string{MetaAny},
		ChargerProfileIDs:    []string{MetaAny},
		DispatcherProfileIDs: []string{MetaAny},
		DispatcherHostIDs:    []string{MetaAny},
		RateProfileIDs:       []string{MetaAny},
		ActionProfileIDs:     []string{MetaAny},
		AccountIDs:           []string{MetaAny},

		AttributeFilterIndexIDs:      []string{MetaAny},
		ResourceFilterIndexIDs:       []string{MetaAny},
		StatFilterIndexIDs:           []string{MetaAny},
		ThresholdFilterIndexIDs:      []string{MetaAny},
		RouteFilterIndexIDs:          []string{MetaAny},
		ChargerFilterIndexIDs:        []string{MetaAny},
		DispatcherFilterIndexIDs:     []string{MetaAny},
		RateProfilesFilterIndexIDs:   []string{MetaAny},
		RateFilterIndexIDs:           []string{MetaAny},
		FilterIndexIDs:               []string{MetaAny},
		ActionProfilesFilterIndexIDs: []string{MetaAny},
		AccountsFilterIndexIDs:       []string{MetaAny},
	}
}

func NewAttrReloadCacheWithOptsFromMap(arg map[string][]string, tnt string, opts map[string]interface{}) *AttrReloadCacheWithAPIOpts {
	return &AttrReloadCacheWithAPIOpts{
		Tenant:  tnt,
		APIOpts: opts,

		ResourceProfileIDs:           arg[CacheResourceProfiles],
		ResourceIDs:                  arg[CacheResources],
		StatsQueueProfileIDs:         arg[CacheStatQueueProfiles],
		StatsQueueIDs:                arg[CacheStatQueues],
		ThresholdProfileIDs:          arg[CacheThresholdProfiles],
		ThresholdIDs:                 arg[CacheThresholds],
		FilterIDs:                    arg[CacheFilters],
		RouteProfileIDs:              arg[CacheRouteProfiles],
		AttributeProfileIDs:          arg[CacheAttributeProfiles],
		ChargerProfileIDs:            arg[CacheChargerProfiles],
		DispatcherProfileIDs:         arg[CacheDispatcherProfiles],
		DispatcherHostIDs:            arg[CacheDispatcherHosts],
		RateProfileIDs:               arg[CacheRateProfiles],
		ActionProfileIDs:             arg[CacheActionProfiles],
		AccountIDs:                   arg[CacheAccounts],
		ResourceFilterIndexIDs:       arg[CacheResourceFilterIndexes],
		StatFilterIndexIDs:           arg[CacheStatFilterIndexes],
		ThresholdFilterIndexIDs:      arg[CacheThresholdFilterIndexes],
		RouteFilterIndexIDs:          arg[CacheRouteFilterIndexes],
		AttributeFilterIndexIDs:      arg[CacheAttributeFilterIndexes],
		ChargerFilterIndexIDs:        arg[CacheChargerFilterIndexes],
		DispatcherFilterIndexIDs:     arg[CacheDispatcherFilterIndexes],
		RateProfilesFilterIndexIDs:   arg[CacheRateProfilesFilterIndexes],
		ActionProfilesFilterIndexIDs: arg[CacheActionProfilesFilterIndexes],
		AccountsFilterIndexIDs:       arg[CacheAccountsFilterIndexes],
		RateFilterIndexIDs:           arg[CacheRateFilterIndexes],
		FilterIndexIDs:               arg[CacheReverseFilterIndexes],
	}
}

type AttrReloadCacheWithAPIOpts struct {
	APIOpts map[string]interface{} `json:",omitempty"`
	Tenant  string                 `json:",omitempty"`

	ResourceProfileIDs   []string `json:",omitempty"`
	ResourceIDs          []string `json:",omitempty"`
	StatsQueueIDs        []string `json:",omitempty"`
	StatsQueueProfileIDs []string `json:",omitempty"`
	ThresholdIDs         []string `json:",omitempty"`
	ThresholdProfileIDs  []string `json:",omitempty"`
	FilterIDs            []string `json:",omitempty"`
	RouteProfileIDs      []string `json:",omitempty"`
	AttributeProfileIDs  []string `json:",omitempty"`
	ChargerProfileIDs    []string `json:",omitempty"`
	DispatcherProfileIDs []string `json:",omitempty"`
	DispatcherHostIDs    []string `json:",omitempty"`
	RateProfileIDs       []string `json:",omitempty"`
	ActionProfileIDs     []string `json:",omitempty"`
	AccountIDs           []string `json:",omitempty"`

	AttributeFilterIndexIDs      []string `json:",omitempty"`
	ResourceFilterIndexIDs       []string `json:",omitempty"`
	StatFilterIndexIDs           []string `json:",omitempty"`
	ThresholdFilterIndexIDs      []string `json:",omitempty"`
	RouteFilterIndexIDs          []string `json:",omitempty"`
	ChargerFilterIndexIDs        []string `json:",omitempty"`
	DispatcherFilterIndexIDs     []string `json:",omitempty"`
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
		CacheStatQueueProfiles:           a.StatsQueueProfileIDs,
		CacheStatQueues:                  a.StatsQueueIDs,
		CacheThresholdProfiles:           a.ThresholdProfileIDs,
		CacheThresholds:                  a.ThresholdIDs,
		CacheFilters:                     a.FilterIDs,
		CacheRouteProfiles:               a.RouteProfileIDs,
		CacheAttributeProfiles:           a.AttributeProfileIDs,
		CacheChargerProfiles:             a.ChargerProfileIDs,
		CacheDispatcherProfiles:          a.DispatcherProfileIDs,
		CacheDispatcherHosts:             a.DispatcherHostIDs,
		CacheRateProfiles:                a.RateProfileIDs,
		CacheActionProfiles:              a.ActionProfileIDs,
		CacheAccounts:                    a.AccountIDs,
		CacheResourceFilterIndexes:       a.ResourceFilterIndexIDs,
		CacheStatFilterIndexes:           a.StatFilterIndexIDs,
		CacheThresholdFilterIndexes:      a.ThresholdFilterIndexIDs,
		CacheRouteFilterIndexes:          a.RouteFilterIndexIDs,
		CacheAttributeFilterIndexes:      a.AttributeFilterIndexIDs,
		CacheChargerFilterIndexes:        a.ChargerFilterIndexIDs,
		CacheDispatcherFilterIndexes:     a.DispatcherFilterIndexIDs,
		CacheRateProfilesFilterIndexes:   a.RateProfilesFilterIndexIDs,
		CacheActionProfilesFilterIndexes: a.ActionProfilesFilterIndexIDs,
		CacheAccountsFilterIndexes:       a.AccountsFilterIndexIDs,
		CacheRateFilterIndexes:           a.RateFilterIndexIDs,
		CacheReverseFilterIndexes:        a.FilterIndexIDs,
	}
}

type AttrCacheIDsWithAPIOpts struct {
	APIOpts  map[string]interface{}
	Tenant   string
	CacheIDs []string
}

type ArgsGetGroupWithAPIOpts struct {
	APIOpts map[string]interface{}
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
	APIOpts map[string]interface{}
}

type SessionIDsWithAPIOpts struct {
	IDs     []string
	Tenant  string
	APIOpts map[string]interface{}
}

type ArgExportToFolder struct {
	Path  string
	Items []string
}

// DPRArgs are the arguments used by dispatcher to send a Disconnect-Peer-Request
type DPRArgs struct {
	OriginHost      string
	OriginRealm     string
	DisconnectCause int
}

type ArgCacheReplicateSet struct {
	CacheID string
	ItemID  string
	Value   interface{}
	APIOpts map[string]interface{}
	Tenant  string
}

// Compiler are objects that need post compiling
type Compiler interface {
	Compile() error
}

type ArgCacheReplicateRemove struct {
	CacheID string
	ItemID  string
	APIOpts map[string]interface{}
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
	Weight    float64
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
	Blocker   bool
	TTL       string
	Type      string
	Opts      string
	Diktats   []*TPAPDiktat
}

type TPAPDiktat struct {
	Path  string
	Value string
}

type TPAccount struct {
	TPid         string
	Tenant       string
	ID           string
	FilterIDs    []string
	Weights      string
	Balances     map[string]*TPAccountBalance
	ThresholdIDs []string
}

type TPAccountBalance struct {
	ID             string
	FilterIDs      []string
	Weights        string
	Blocker        bool
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
