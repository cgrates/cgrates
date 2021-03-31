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
	"fmt"
	"sort"
	"strings"
	"time"
)

// Used to extract ids from stordb
type TPDistinctIds []string

func (tpdi TPDistinctIds) String() string {
	return strings.Join(tpdi, FieldsSep)
}

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

// TPDestination represents one destination in storDB
type TPDestination struct {
	TPid     string   // Tariff plan id
	ID       string   // Destination id
	Prefixes []string // Prefixes attached to this destination
}

// This file deals with tp_* data definition
// TPRateRALs -> TPRateRALs
type TPRateRALs struct {
	TPid      string      // Tariff plan id
	ID        string      // Rate id
	RateSlots []*RateSlot // One or more RateSlots
}

// Needed so we make sure we always use SetDurations() on a newly created value
func NewRateSlot(connectFee, rate float64, rateUnit, rateIncrement, grpInterval string) (*RateSlot, error) {
	rs := &RateSlot{
		ConnectFee:         connectFee,
		Rate:               rate,
		RateUnit:           rateUnit,
		RateIncrement:      rateIncrement,
		GroupIntervalStart: grpInterval,
	}
	if err := rs.SetDurations(); err != nil {
		return nil, err
	}
	return rs, nil
}

type RateSlot struct {
	ConnectFee            float64 // ConnectFee applied once the call is answered
	Rate                  float64 // Rate applied
	RateUnit              string  //  Number of billing units this rate applies to
	RateIncrement         string  // This rate will apply in increments of duration
	GroupIntervalStart    string  // Group position
	rateUnitDur           time.Duration
	rateIncrementDur      time.Duration
	groupIntervalStartDur time.Duration
	tag                   string // load validation only
}

// Used to set the durations we need out of strings
func (rs *RateSlot) SetDurations() error {
	var err error
	if rs.rateUnitDur, err = ParseDurationWithNanosecs(rs.RateUnit); err != nil {
		return err
	}
	if rs.rateIncrementDur, err = ParseDurationWithNanosecs(rs.RateIncrement); err != nil {
		return err
	}
	if rs.groupIntervalStartDur, err = ParseDurationWithNanosecs(rs.GroupIntervalStart); err != nil {
		return err
	}
	return nil
}
func (rs *RateSlot) RateUnitDuration() time.Duration {
	return rs.rateUnitDur
}
func (rs *RateSlot) RateIncrementDuration() time.Duration {
	return rs.rateIncrementDur
}
func (rs *RateSlot) GroupIntervalStartDuration() time.Duration {
	return rs.groupIntervalStartDur
}

type TPDestinationRate struct {
	TPid             string             // Tariff plan id
	ID               string             // DestinationRate profile id
	DestinationRates []*DestinationRate // Set of destinationid-rateid bindings
}

type DestinationRate struct {
	DestinationId    string // The destination identity
	RateId           string // The rate identity
	Rate             *TPRateRALs
	RoundingMethod   string
	RoundingDecimals int
	MaxCost          float64
	MaxCostStrategy  string
}

type ApierTPTiming struct {
	TPid      string // Tariff plan id
	ID        string // Timing id
	Years     string // semicolon separated list of years this timing is valid on, *any supported
	Months    string // semicolon separated list of months this timing is valid on, *any supported
	MonthDays string // semicolon separated list of month's days this timing is valid on, *any supported
	WeekDays  string // semicolon separated list of week day names this timing is valid on *any supported
	Time      string // String representing the time this timing starts on
}

type TPTiming struct {
	ID        string
	Years     Years
	Months    Months
	MonthDays MonthDays
	WeekDays  WeekDays
	StartTime string
	EndTime   string
}

// TPTimingWithAPIOpts is used in replicatorV1 for dispatcher
type TPTimingWithAPIOpts struct {
	*TPTiming
	Tenant  string
	APIOpts map[string]interface{}
}

func NewTiming(ID, years, mounths, mounthdays, weekdays, time string) (rt *TPTiming) {
	rt = &TPTiming{}
	rt.ID = ID
	rt.Years.Parse(years, InfieldSep)
	rt.Months.Parse(mounths, InfieldSep)
	rt.MonthDays.Parse(mounthdays, InfieldSep)
	rt.WeekDays.Parse(weekdays, InfieldSep)
	times := strings.Split(time, InfieldSep)
	rt.StartTime = times[0]
	if len(times) > 1 {
		rt.EndTime = times[1]
	}
	return
}

type TPRatingPlan struct {
	TPid               string                 // Tariff plan id
	ID                 string                 // RatingPlan profile id
	RatingPlanBindings []*TPRatingPlanBinding // Set of destinationid-rateid bindings
}

type TPRatingPlanBinding struct {
	DestinationRatesId string    // The DestinationRate identity
	TimingId           string    // The timing identity
	Weight             float64   // Binding priority taken into consideration when more DestinationRates are active on a time slot
	timing             *TPTiming // Not exporting it via JSON
}

func (self *TPRatingPlanBinding) SetTiming(tm *TPTiming) {
	self.timing = tm
}

func (self *TPRatingPlanBinding) Timing() *TPTiming {
	return self.timing
}

type TPRatingProfile struct {
	TPid                  string                // Tariff plan id
	LoadId                string                // Gives ability to load specific RatingProfile based on load identifier, hence being able to keep history also in stordb
	Tenant                string                // Tenant's Id
	Category              string                // TypeOfRecord
	Subject               string                // Rating subject, usually the same as account
	RatingPlanActivations []*TPRatingActivation // Activate rate profiles at specific time
}

// Used as key in nosql db (eg: redis)
func (rpf *TPRatingProfile) KeyId() string {
	return ConcatenatedKey(MetaOut,
		rpf.Tenant, rpf.Category, rpf.Subject)
}

func (rpf *TPRatingProfile) GetId() string {
	return ConcatenatedKey(rpf.LoadId, MetaOut,
		rpf.Tenant, rpf.Category, rpf.Subject)
}

func (rpf *TPRatingProfile) SetRatingProfileID(id string) error {
	ids := strings.Split(id, ConcatenatedKeySep)
	if len(ids) != 4 {
		return fmt.Errorf("Wrong TPRatingProfileId: %s", id)
	}
	rpf.LoadId = ids[0]
	rpf.Tenant = ids[1]
	rpf.Category = ids[2]
	rpf.Subject = ids[3]
	return nil
}

type AttrSetRatingProfile struct {
	Tenant                string                // Tenant's Id
	Category              string                // TypeOfRecord
	Subject               string                // Rating subject, usually the same as account
	Overwrite             bool                  // Overwrite if exists
	RatingPlanActivations []*TPRatingActivation // Activate rating plans at specific time
}

type AttrGetRatingProfile struct {
	Tenant   string // Tenant's Id
	Category string // TypeOfRecord
	Subject  string // Rating subject, usually the same as account
}

func (self *AttrGetRatingProfile) GetID() string {
	return ConcatenatedKey(MetaOut, self.Tenant, self.Category, self.Subject)
}

type TPRatingActivation struct {
	ActivationTime   string // Time when this profile will become active, defined as unix epoch time
	RatingPlanId     string // Id of RatingPlan profile
	FallbackSubjects string // So we follow the api
}

// Helper to return the subject fallback keys we need in dataDb
func FallbackSubjKeys(tenant, tor, fallbackSubjects string) []string {
	var sslice sort.StringSlice
	if len(fallbackSubjects) != 0 {
		for _, fbs := range strings.Split(fallbackSubjects, string(FallbackSep)) {
			newKey := ConcatenatedKey(MetaOut, tenant, tor, fbs)
			i := sslice.Search(newKey)
			if i < len(sslice) && sslice[i] != newKey {
				// not found so insert it
				sslice = append(sslice, "")
				copy(sslice[i+1:], sslice[i:])
				sslice[i] = newKey
			} else if i == len(sslice) {
				// not found and at the end
				sslice = append(sslice, newKey)
			} // newKey was found
		}
	}
	return sslice
}

type AttrSetDestination struct {
	Id        string
	Prefixes  []string
	Overwrite bool
}

type AttrTPRatingProfileIds struct {
	TPid     string // Tariff plan id
	Tenant   string // Tenant's Id
	Category string // TypeOfRecord
	Subject  string // Rating subject, usually the same as account
}

type TPActions struct {
	TPid    string      // Tariff plan id
	ID      string      // Actions id
	Actions []*TPAction // Set of actions this Actions profile will perform
}

type TPAction struct {
	Identifier      string // Identifier mapped in the code
	BalanceId       string // Balance identification string (account scope)
	BalanceUuid     string // Balance identification string (global scope)
	BalanceType     string // Type of balance the action will operate on
	Units           string // Number of units to add/deduct
	ExpiryTime      string // Time when the units will expire
	Filter          string // The condition on balances that is checked before the action
	TimingTags      string // Timing when balance is active
	DestinationIds  string // Destination profile id
	RatingSubject   string // Reference a rate subject defined in RatingProfiles
	Categories      string // category filter for balances
	SharedGroups    string // Reference to a shared group
	BalanceWeight   string // Balance weight
	ExtraParameters string
	BalanceBlocker  string
	BalanceDisabled string
	Weight          float64 // Action's weight
}

type TPSharedGroups struct {
	TPid         string
	ID           string
	SharedGroups []*TPSharedGroup
}

type TPSharedGroup struct {
	Account       string
	Strategy      string
	RatingSubject string
}

type TPActionPlan struct {
	TPid       string            // Tariff plan id
	ID         string            // ActionPlan id
	ActionPlan []*TPActionTiming // Set of ActionTiming bindings this profile will group
}

type TPActionTiming struct {
	ActionsId string  // Actions id
	TimingId  string  // Timing profile id
	Weight    float64 // Binding's weight
}

type TPActionTriggers struct {
	TPid           string             // Tariff plan id
	ID             string             // action trigger id
	ActionTriggers []*TPActionTrigger // Set of triggers grouped in this profile
}

type TPActionTrigger struct {
	Id                    string  // group id
	UniqueID              string  // individual id
	ThresholdType         string  // This threshold type
	ThresholdValue        float64 // Threshold
	Recurrent             bool    // reset executed flag each run
	MinSleep              string  // Minimum duration between two executions in case of recurrent triggers
	ExpirationDate        string  // Trigger expiration
	ActivationDate        string  // Trigger activation
	BalanceId             string  // The id of the balance in the account
	BalanceType           string  // Type of balance this trigger monitors
	BalanceDestinationIds string  // filter for balance
	BalanceWeight         string  // filter for balance
	BalanceExpirationDate string  // filter for balance
	BalanceTimingTags     string  // filter for balance
	BalanceRatingSubject  string  // filter for balance
	BalanceCategories     string  // filter for balance
	BalanceSharedGroups   string  // filter for balance
	BalanceBlocker        string  // filter for balance
	BalanceDisabled       string  // filter for balance
	ActionsId             string  // Actions which will execute on threshold reached
	Weight                float64 // weight
}

type TPAccountActions struct {
	TPid             string // Tariff plan id
	LoadId           string // LoadId, used to group actions on a load
	Tenant           string // Tenant's Id
	Account          string // Account name
	ActionPlanId     string // Id of ActionPlan profile to use
	ActionTriggersId string // Id of ActionTriggers profile to use
	AllowNegative    bool
	Disabled         bool
}

// Returns the id used in some nosql dbs (eg: redis)
func (aa *TPAccountActions) KeyId() string {
	return ConcatenatedKey(aa.Tenant, aa.Account)
}

func (aa *TPAccountActions) GetId() string {
	return aa.LoadId + ConcatenatedKeySep + aa.Tenant + ConcatenatedKeySep + aa.Account
}

func (aa *TPAccountActions) SetAccountActionsId(id string) error {
	ids := strings.Split(id, ConcatenatedKeySep)
	if len(ids) != 3 {
		return fmt.Errorf("Wrong TP Account Action Id: %s", id)
	}
	aa.LoadId = ids[0]
	aa.Tenant = ids[1]
	aa.Account = ids[2]
	return nil
}

type AttrGetAccount struct {
	Tenant  string
	Account string
}

type AttrGetAccounts struct {
	Tenant     string
	AccountIDs []string
	Offset     int // Set the item offset
	Limit      int // Limit number of items retrieved
	Filter     map[string]bool
}

type AttrGetAccountsCount struct {
	Tenant string
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
	Validate   bool   // Run structural checks on data
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

func NewTAFromAccountKey(accountKey string) (*TenantAccount, error) {
	accountSplt := strings.Split(accountKey, ConcatenatedKeySep)
	if len(accountSplt) != 2 {
		return nil, fmt.Errorf("Unsupported format for TenantAccount: %s", accountKey)
	}
	return &TenantAccount{accountSplt[0], accountSplt[1]}, nil
}

type TenantAccount struct {
	Tenant, Account string
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

type AttrSetActions struct {
	ActionsId string      // Actions id
	Overwrite bool        // If previously defined, will be overwritten
	Actions   []*TPAction // Set of actions this Actions profile will perform
}

type AttrExecuteAction struct {
	Tenant    string
	Account   string
	ActionsId string
}

type AttrSetAccount struct {
	Tenant           string
	Account          string
	ActionPlanID     string
	ActionTriggersID string
	ExtraOptions     map[string]bool
	ReloadScheduler  bool
}

type AttrRemoveAccount struct {
	Tenant          string
	Account         string
	ReloadScheduler bool
}

type AttrGetCallCost struct {
	CgrId string // Unique id of the CDR
	RunId string // Run Id
}

type AttrSetBalance struct {
	Tenant          string
	Account         string
	BalanceType     string
	Value           float64
	Balance         map[string]interface{}
	ActionExtraData *map[string]interface{}
	Cdrlog          bool
}

type AttrSetBalances struct {
	Tenant   string
	Account  string
	Balances []*AttrBalance
}

type AttrBalance struct {
	BalanceType     string
	Value           float64
	Balance         map[string]interface{}
	ActionExtraData *map[string]interface{}
	Cdrlog          bool
}

// TPResourceProfile is used in APIs to manage remotely offline ResourceProfile
type TPResourceProfile struct {
	TPid               string
	Tenant             string
	ID                 string // Identifier of this limit
	FilterIDs          []string
	ActivationInterval *TPActivationInterval // Time when this limit becomes active/expires
	UsageTTL           string
	Limit              string // Limit value
	AllocationMessage  string
	Blocker            bool // blocker flag to stop processing on filters matched
	Stored             bool
	Weight             float64  // Weight to sort the ResourceLimits
	ThresholdIDs       []string // Thresholds to check after changing Limit
}

// TPActivationInterval represents an activation interval for an item
type TPActivationInterval struct {
	ActivationTime string
	ExpiryTime     string
}

type ArgRSv1ResourceUsage struct {
	*CGREvent
	UsageID  string // ResourceUsage Identifier
	UsageTTL *time.Duration
	Units    float64
	clnb     bool //rpcclonable
}

// SetCloneable sets if the args should be clonned on internal connections
func (attr *ArgRSv1ResourceUsage) SetCloneable(rpcCloneable bool) {
	attr.clnb = rpcCloneable
}

// RPCClone implements rpcclient.RPCCloner interface
func (attr *ArgRSv1ResourceUsage) RPCClone() (interface{}, error) {
	if !attr.clnb {
		return attr, nil
	}
	return attr.Clone(), nil
}

// Clone creates a clone of the object
func (attr *ArgRSv1ResourceUsage) Clone() *ArgRSv1ResourceUsage {
	var usageTTL *time.Duration
	if attr.UsageTTL != nil {
		usageTTL = DurationPointer(*attr.UsageTTL)
	}
	return &ArgRSv1ResourceUsage{
		UsageID:  attr.UsageID,
		UsageTTL: usageTTL,
		Units:    attr.Units,
		CGREvent: attr.CGREvent.Clone(),
	}
}

type ArgsComputeFilterIndexIDs struct {
	Tenant        string
	Context       string
	AttributeIDs  []string
	ResourceIDs   []string
	StatIDs       []string
	RouteIDs      []string
	ThresholdIDs  []string
	ChargerIDs    []string
	DispatcherIDs []string
}

type ArgsComputeFilterIndexes struct {
	Tenant      string
	Context     string
	AttributeS  bool
	ResourceS   bool
	StatS       bool
	RouteS      bool
	ThresholdS  bool
	ChargerS    bool
	DispatcherS bool
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
	TPid               string
	Tenant             string
	ID                 string
	FilterIDs          []string
	ActivationInterval *TPActivationInterval
	QueueLength        int
	TTL                string
	Metrics            []*MetricWithFilters
	Blocker            bool // blocker flag to stop processing on filters matched
	Stored             bool
	Weight             float64
	MinItems           int
	ThresholdIDs       []string
}

// TPThresholdProfile is used in APIs to manage remotely offline ThresholdProfile
type TPThresholdProfile struct {
	TPid               string
	Tenant             string
	ID                 string
	FilterIDs          []string
	ActivationInterval *TPActivationInterval // Time when this limit becomes active and expires
	MaxHits            int
	MinHits            int
	MinSleep           string
	Blocker            bool    // blocker flag to stop processing on filters matched
	Weight             float64 // Weight to sort the thresholds
	ActionIDs          []string
	Async              bool
}

// TPFilterProfile is used in APIs to manage remotely offline FilterProfile
type TPFilterProfile struct {
	TPid               string
	Tenant             string
	ID                 string
	Filters            []*TPFilter
	ActivationInterval *TPActivationInterval // Time when this limit becomes active and expires
}

// TPFilter is used in TPFilterProfile
type TPFilter struct {
	Type    string   // Filter type (*string, *timing, *rsr_filters, *cdr_stats)
	Element string   // Name of the field providing us the Values to check (used in case of some )
	Values  []string // Filter definition
}

// TPRoute is used in TPRouteProfile
type TPRoute struct {
	ID              string // RouteID
	FilterIDs       []string
	AccountIDs      []string
	RatingPlanIDs   []string // used when computing price
	ResourceIDs     []string // queried in some strategies
	StatIDs         []string // queried in some strategies
	Weight          float64
	Blocker         bool
	RouteParameters string
}

// TPRouteProfile is used in APIs to manage remotely offline RouteProfile
type TPRouteProfile struct {
	TPid               string
	Tenant             string
	ID                 string
	FilterIDs          []string
	ActivationInterval *TPActivationInterval // Time when this limit becomes active and expires
	Sorting            string
	SortingParameters  []string
	Routes             []*TPRoute
	Weight             float64
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
	TPid               string
	Tenant             string
	ID                 string
	FilterIDs          []string
	ActivationInterval *TPActivationInterval // Time when this limit becomes active and expires
	Contexts           []string              // bind this TPAttribute to multiple context
	Attributes         []*TPAttribute
	Blocker            bool
	Weight             float64
}

// TPChargerProfile is used in APIs to manage remotely offline ChargerProfile
type TPChargerProfile struct {
	TPid               string
	Tenant             string
	ID                 string
	FilterIDs          []string
	ActivationInterval *TPActivationInterval // Time when this limit becomes active and expires
	RunID              string
	AttributeIDs       []string
	Weight             float64
}

type TPTntID struct {
	TPid   string
	Tenant string
	ID     string
}

// TPDispatcherProfile is used in APIs to manage remotely offline DispatcherProfile
type TPDispatcherProfile struct {
	TPid               string
	Tenant             string
	ID                 string
	Subsystems         []string
	FilterIDs          []string
	ActivationInterval *TPActivationInterval // Time when this limit becomes active and expires
	Strategy           string
	StrategyParams     []interface{} // ie for distribution, set here the pool weights
	Weight             float64
	Hosts              []*TPDispatcherHostProfile
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
	Address   string
	Transport string
	TLS       bool
}

type UsageInterval struct {
	Min *time.Duration
	Max *time.Duration
}

type TimeInterval struct {
	Begin *time.Time
	End   *time.Time
}

type AttrRemoteLock struct {
	ReferenceID string        // reference ID for this lock if available
	LockIDs     []string      // List of IDs to obtain lock for
	Timeout     time.Duration // Automatically unlock on timeout
}

type SMCostFilter struct { //id cu litere mare
	CGRIDs         []string
	NotCGRIDs      []string
	RunIDs         []string
	NotRunIDs      []string
	OriginHosts    []string
	NotOriginHosts []string
	OriginIDs      []string
	NotOriginIDs   []string
	CostSources    []string
	NotCostSources []string
	Usage          UsageInterval
	CreatedAt      TimeInterval
}

func AppendToSMCostFilter(smcFilter *SMCostFilter, fieldType, fieldName string,
	values []string, timezone string) (smcf *SMCostFilter, err error) {
	switch fieldName {
	case MetaScPrefix + CGRID:
		switch fieldType {
		case MetaString:
			smcFilter.CGRIDs = append(smcFilter.CGRIDs, values...)
		case MetaNotString:
			smcFilter.NotCGRIDs = append(smcFilter.NotCGRIDs, values...)
		default:
			err = fmt.Errorf("FilterType: %q not supported for FieldName: %q", fieldType, fieldName)
		}
	case MetaScPrefix + RunID:
		switch fieldType {
		case MetaString:
			smcFilter.RunIDs = append(smcFilter.RunIDs, values...)
		case MetaNotString:
			smcFilter.NotRunIDs = append(smcFilter.NotRunIDs, values...)
		default:
			err = fmt.Errorf("FilterType: %q not supported for FieldName: %q", fieldType, fieldName)
		}
	case MetaScPrefix + OriginHost:
		switch fieldType {
		case MetaString:
			smcFilter.OriginHosts = append(smcFilter.OriginHosts, values...)
		case MetaNotString:
			smcFilter.NotOriginHosts = append(smcFilter.NotOriginHosts, values...)
		default:
			err = fmt.Errorf("FilterType: %q not supported for FieldName: %q", fieldType, fieldName)
		}
	case MetaScPrefix + OriginID:
		switch fieldType {
		case MetaString:
			smcFilter.OriginIDs = append(smcFilter.OriginIDs, values...)
		case MetaNotString:
			smcFilter.NotOriginIDs = append(smcFilter.NotOriginIDs, values...)
		default:
			err = fmt.Errorf("FilterType: %q not supported for FieldName: %q", fieldType, fieldName)
		}
	case MetaScPrefix + CostSource:
		switch fieldType {
		case MetaString:
			smcFilter.CostSources = append(smcFilter.CostSources, values...)
		case MetaNotString:
			smcFilter.NotCostSources = append(smcFilter.NotCostSources, values...)
		default:
			err = fmt.Errorf("FilterType: %q not supported for FieldName: %q", fieldType, fieldName)
		}
	case MetaScPrefix + Usage:
		switch fieldType {
		case MetaGreaterOrEqual:
			var minUsage time.Duration
			minUsage, err = ParseDurationWithNanosecs(values[0])
			if err != nil {
				err = fmt.Errorf("Error when converting field: %q  value: %q in time.Duration ", fieldType, fieldName)
				break
			}
			smcFilter.Usage.Min = &minUsage
		case MetaLessThan:
			var maxUsage time.Duration
			maxUsage, err = ParseDurationWithNanosecs(values[0])
			if err != nil {
				err = fmt.Errorf("Error when converting field: %q  value: %q in time.Duration ", fieldType, fieldName)
				break
			}
			smcFilter.Usage.Max = &maxUsage
		default:
			err = fmt.Errorf("FilterType: %q not supported for FieldName: %q", fieldType, fieldName)
		}
	case MetaScPrefix + CreatedAt:
		switch fieldType {
		case MetaGreaterOrEqual:
			var start time.Time
			start, err = ParseTimeDetectLayout(values[0], timezone)
			if err != nil {
				err = fmt.Errorf("Error when converting field: %q  value: %q in time.Time ", fieldType, fieldName)
				break
			}
			if !start.IsZero() {
				smcFilter.CreatedAt.Begin = &start
			}
		case MetaLessThan:
			var end time.Time
			end, err = ParseTimeDetectLayout(values[0], timezone)
			if err != nil {
				err = fmt.Errorf("Error when converting field: %q  value: %q in time.Time ", fieldType, fieldName)
				break
			}
			if !end.IsZero() {
				smcFilter.CreatedAt.End = &end
			}
		default:
			err = fmt.Errorf("FilterType: %q not supported for FieldName: %q", fieldType, fieldName)
		}
	default:
		err = fmt.Errorf("FieldName: %q not supported", fieldName)
	}
	return smcFilter, err
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
		ArgsCache: map[string][]string{
			DestinationIDs:           nil,
			ReverseDestinationIDs:    nil,
			RatingPlanIDs:            nil,
			RatingProfileIDs:         nil,
			ActionIDs:                nil,
			ActionPlanIDs:            nil,
			AccountActionPlanIDs:     nil,
			ActionTriggerIDs:         nil,
			SharedGroupIDs:           nil,
			ResourceProfileIDs:       nil,
			ResourceIDs:              nil,
			StatsQueueIDs:            nil,
			StatsQueueProfileIDs:     nil,
			ThresholdIDs:             nil,
			ThresholdProfileIDs:      nil,
			FilterIDs:                nil,
			RouteProfileIDs:          nil,
			AttributeProfileIDs:      nil,
			ChargerProfileIDs:        nil,
			DispatcherProfileIDs:     nil,
			DispatcherHostIDs:        nil,
			TimingIDs:                nil,
			AttributeFilterIndexIDs:  nil,
			ResourceFilterIndexIDs:   nil,
			StatFilterIndexIDs:       nil,
			ThresholdFilterIndexIDs:  nil,
			RouteFilterIndexIDs:      nil,
			ChargerFilterIndexIDs:    nil,
			DispatcherFilterIndexIDs: nil,
			RateFilterIndexIDs:       nil,
			FilterIndexIDs:           nil,
		},
	}
}

type AttrReloadCacheWithAPIOpts struct {
	APIOpts   map[string]interface{}
	Tenant    string
	ArgsCache map[string][]string
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

type RatingPlanCostArg struct {
	RatingPlanIDs []string
	Destination   string
	SetupTime     string
	Usage         string
	APIOpts       map[string]interface{}
}
type SessionIDsWithArgsDispatcher struct {
	IDs     []string
	Tenant  string
	APIOpts map[string]interface{}
}

type GetCostOnRatingPlansArgs struct {
	Account       string
	Subject       string
	Destination   string
	Tenant        string
	SetupTime     time.Time
	Usage         time.Duration
	RatingPlanIDs []string
}

type GetMaxSessionTimeOnAccountsArgs struct {
	Subject     string
	Destination string
	Tenant      string
	SetupTime   time.Time
	Usage       time.Duration
	AccountIDs  []string
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

type AttrsExecuteActions struct {
	ActionPlanID string
	TimeStart    time.Time
	TimeEnd      time.Time // replay the action timings between the two dates
	APIOpts      map[string]interface{}
	Tenant       string
}

type AttrsExecuteActionPlans struct {
	ActionPlanIDs []string
	Tenant        string
	AccountID     string
	APIOpts       map[string]interface{}
}

type ArgExportCDRs struct {
	ExporterIDs []string // exporterIDs is used to said which exporter are using to export the cdrs
	Verbose     bool     // verbose is used to inform the user about the positive and negative exported cdrs
	RPCCDRsFilter
}
