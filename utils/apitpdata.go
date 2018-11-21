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
	return strings.Join(tpdi, ",")
}

// Paginate stuff around items returned
type Paginator struct {
	Limit      *int   // Limit the number of items returned
	Offset     *int   // Offset of the first item returned (eg: use Limit*Page in case of PerPage items)
	SearchTerm string // Global matching pattern in items returned, partially used in some APIs
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
	if offset != 0 {
		limit = limit + offset
	}
	if limit == 0 {
		limit = len(in[offset:])
	} else if limit > len(in) {
		limit = len(in)
	}
	ret := in[offset:limit]
	out = make([]string, len(ret))
	for i, itm := range ret {
		out[i] = itm
	}
	return
}

// TPDestination represents one destination in storDB
type TPDestination struct {
	TPid     string   // Tariff plan id
	ID       string   // Destination id
	Prefixes []string // Prefixes attached to this destination
}

func (v1TPDst *TPDestination) AsTPDestination() *TPDestination {
	return &TPDestination{TPid: v1TPDst.TPid, ID: v1TPDst.ID, Prefixes: v1TPDst.Prefixes}
}

// This file deals with tp_* data definition

type TPRate struct {
	TPid      string      // Tariff plan id
	ID        string      // Rate id
	RateSlots []*RateSlot // One or more RateSlots
}

// Needed so we make sure we always use SetDurations() on a newly created value
func NewRateSlot(connectFee, rate float64, rateUnit, rateIncrement, grpInterval string) (*RateSlot, error) {
	rs := &RateSlot{ConnectFee: connectFee, Rate: rate, RateUnit: rateUnit, RateIncrement: rateIncrement,
		GroupIntervalStart: grpInterval}
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
func (self *RateSlot) SetDurations() error {
	var err error
	if self.rateUnitDur, err = ParseDurationWithNanosecs(self.RateUnit); err != nil {
		return err
	}
	if self.rateIncrementDur, err = ParseDurationWithNanosecs(self.RateIncrement); err != nil {
		return err
	}
	if self.groupIntervalStartDur, err = ParseDurationWithNanosecs(self.GroupIntervalStart); err != nil {
		return err
	}
	return nil
}
func (self *RateSlot) RateUnitDuration() time.Duration {
	return self.rateUnitDur
}
func (self *RateSlot) RateIncrementDuration() time.Duration {
	return self.rateIncrementDur
}
func (self *RateSlot) GroupIntervalStartDuration() time.Duration {
	return self.groupIntervalStartDur
}

type TPDestinationRate struct {
	TPid             string             // Tariff plan id
	ID               string             // DestinationRate profile id
	DestinationRates []*DestinationRate // Set of destinationid-rateid bindings
}

type DestinationRate struct {
	DestinationId    string // The destination identity
	RateId           string // The rate identity
	Rate             *TPRate
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

func NewTiming(timingInfo ...string) (rt *TPTiming) {
	rt = &TPTiming{}
	rt.ID = timingInfo[0]
	rt.Years.Parse(timingInfo[1], INFIELD_SEP)
	rt.Months.Parse(timingInfo[2], INFIELD_SEP)
	rt.MonthDays.Parse(timingInfo[3], INFIELD_SEP)
	rt.WeekDays.Parse(timingInfo[4], INFIELD_SEP)
	times := strings.Split(timingInfo[5], INFIELD_SEP)
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

// Used to rebuild a TPRatingProfile (empty RatingPlanActivations) out of it's key in nosqldb
func NewTPRatingProfileFromKeyId(tpid, loadId, keyId string) (*TPRatingProfile, error) {
	s := strings.Split(keyId, ":")
	if len(s) != 4 {
		return nil, fmt.Errorf("Cannot parse key %s into RatingProfile", keyId)
	}
	return &TPRatingProfile{TPid: tpid, LoadId: loadId, Tenant: s[1], Category: s[2], Direction: s[0], Subject: s[3]}, nil
}

type TPRatingProfile struct {
	TPid                  string                // Tariff plan id
	LoadId                string                // Gives ability to load specific RatingProfile based on load identifier, hence being able to keep history also in stordb
	Direction             string                // Traffic direction, OUT is the only one supported for now
	Tenant                string                // Tenant's Id
	Category              string                // TypeOfRecord
	Subject               string                // Rating subject, usually the same as account
	RatingPlanActivations []*TPRatingActivation // Activate rate profiles at specific time
}

// Used as key in nosql db (eg: redis)
func (self *TPRatingProfile) KeyId() string {
	return fmt.Sprintf("%s:%s:%s:%s", self.Direction, self.Tenant, self.Category, self.Subject)
}

func (self *TPRatingProfile) KeyIdA() string {
	return fmt.Sprintf("%s:%s:%s:%s:%s", self.LoadId, self.Direction, self.Tenant, self.Category, self.Subject)
}

func (rpf *TPRatingProfile) GetRatingProfilesId() string {
	return fmt.Sprintf("%s%s%s%s%s%s%s%s%s", rpf.LoadId, CONCATENATED_KEY_SEP, rpf.Direction, CONCATENATED_KEY_SEP, rpf.Tenant, CONCATENATED_KEY_SEP, rpf.Category, CONCATENATED_KEY_SEP, rpf.Subject)
}

func (rpf *TPRatingProfile) SetRatingProfilesId(id string) error {
	ids := strings.Split(id, CONCATENATED_KEY_SEP)
	if len(ids) != 5 {
		return fmt.Errorf("Wrong TPRatingProfileId: %s", id)
	}
	rpf.LoadId = ids[0]
	rpf.Direction = ids[1]
	rpf.Tenant = ids[2]
	rpf.Category = ids[3]
	rpf.Subject = ids[4]
	return nil
}

type AttrSetRatingProfile struct {
	Tenant                string                // Tenant's Id
	Category              string                // TypeOfRecord
	Direction             string                // Traffic direction, OUT is the only one supported for now
	Subject               string                // Rating subject, usually the same as account
	Overwrite             bool                  // Overwrite if exists
	RatingPlanActivations []*TPRatingActivation // Activate rating plans at specific time
}

type AttrGetRatingProfile struct {
	Tenant    string // Tenant's Id
	Category  string // TypeOfRecord
	Direction string // Traffic direction, OUT is the only one supported for now
	Subject   string // Rating subject, usually the same as account
}

func (self *AttrGetRatingProfile) GetID() string {
	return ConcatenatedKey(self.Direction, self.Tenant, self.Category, self.Subject)
}

type TPRatingActivation struct {
	ActivationTime   string // Time when this profile will become active, defined as unix epoch time
	RatingPlanId     string // Id of RatingPlan profile
	FallbackSubjects string // So we follow the api
	CdrStatQueueIds  string
}

// Helper to return the subject fallback keys we need in dataDb
func FallbackSubjKeys(direction, tenant, tor, fallbackSubjects string) []string {
	var sslice sort.StringSlice
	if len(fallbackSubjects) != 0 {
		for _, fbs := range strings.Split(fallbackSubjects, string(FALLBACK_SEP)) {
			newKey := fmt.Sprintf("%s:%s:%s:%s", direction, tenant, tor, fbs)
			i := sslice.Search(newKey)
			if i < len(sslice) && sslice[i] != newKey {
				// not found so insert it
				sslice = append(sslice, "")
				copy(sslice[i+1:], sslice[i:])
				sslice[i] = newKey
			} else {
				if i == len(sslice) {
					// not found and at the end
					sslice = append(sslice, newKey)
				}
			} // newKey was foundfound
		}
	}
	return sslice
}

type AttrSetDestination struct { //ToDo
	Id        string
	Prefixes  []string
	Overwrite bool
}

type AttrTPRatingProfileIds struct {
	TPid      string // Tariff plan id
	Tenant    string // Tenant's Id
	Category  string // TypeOfRecord
	Direction string // Traffic direction
	Subject   string // Rating subject, usually the same as account
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
	Directions      string // Balance direction
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

type TpAlias struct {
	Direction string
	Tenant    string
	Category  string
	Account   string
	Subject   string
	Group     string
	Values    []*AliasValue
}

type AliasValue struct {
	DestinationId string
	Alias         string
	Weight        float64
}

type TPAliases struct {
	TPid      string
	Direction string
	Tenant    string
	Category  string
	Account   string
	Subject   string
	Context   string
	Values    []*TPAliasValue
}

type TPAliasValue struct {
	DestinationId string
	Target        string
	Original      string
	Alias         string
	Weight        float64
}

func (a *TPAliases) GetId() string {
	return ConcatenatedKey(a.Direction, a.Tenant, a.Category, a.Account, a.Subject, a.Context)
}

func (a *TPAliases) SetId(id string) error {
	vals := strings.Split(id, CONCATENATED_KEY_SEP)
	if len(vals) != 6 {
		return ErrInvalidKey
	}
	a.Direction = vals[0]
	a.Tenant = vals[1]
	a.Category = vals[2]
	a.Account = vals[3]
	a.Subject = vals[4]
	a.Context = vals[5]
	return nil
}

type TPUsers struct {
	TPid     string
	Tenant   string
	UserName string
	Masked   bool
	Weight   float64
	Profile  []*TPUserProfile
}

type TPUserProfile struct {
	AttrName  string
	AttrValue string
}

func (u *TPUsers) GetId() string {
	return ConcatenatedKey(u.Tenant, u.UserName)
}

func (tu *TPUsers) SetId(id string) error {
	vals := strings.Split(id, CONCATENATED_KEY_SEP)
	if len(vals) != 2 {
		return ErrInvalidKey
	}
	tu.Tenant = vals[0]
	tu.UserName = vals[1]
	return nil
}

type TPDerivedChargers struct {
	TPid            string
	LoadId          string
	Direction       string
	Tenant          string
	Category        string
	Account         string
	Subject         string
	DestinationIds  string
	DerivedChargers []*TPDerivedCharger
}

type TPDerivedCharger struct {
	RunId                string
	RunFilters           string
	ReqTypeField         string
	DirectionField       string
	TenantField          string
	CategoryField        string
	AccountField         string
	SubjectField         string
	DestinationField     string
	SetupTimeField       string
	PddField             string
	AnswerTimeField      string
	UsageField           string
	SupplierField        string
	DisconnectCauseField string
	CostField            string
	RatedField           string
}

// Key used in dataDb to identify DerivedChargers set
func (tpdc *TPDerivedChargers) GetDerivedChargersKey() string {
	return DerivedChargersKey(tpdc.Direction, tpdc.Tenant, tpdc.Category, tpdc.Account, tpdc.Subject)

}

func (tpdc *TPDerivedChargers) GetDerivedChargesId() string {
	return tpdc.LoadId +
		CONCATENATED_KEY_SEP +
		tpdc.Direction +
		CONCATENATED_KEY_SEP +
		tpdc.Tenant +
		CONCATENATED_KEY_SEP +
		tpdc.Category +
		CONCATENATED_KEY_SEP +
		tpdc.Account +
		CONCATENATED_KEY_SEP +
		tpdc.Subject
}

func (tpdc *TPDerivedChargers) SetDerivedChargersId(id string) error {
	ids := strings.Split(id, CONCATENATED_KEY_SEP)
	if len(ids) != 6 {
		return fmt.Errorf("Wrong TP Derived Charge Id: %s", id)
	}
	tpdc.LoadId = ids[0]
	tpdc.Direction = ids[1]
	tpdc.Tenant = ids[2]
	tpdc.Category = ids[3]
	tpdc.Account = ids[4]
	tpdc.Subject = ids[5]
	return nil
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
	BalanceDirections     string  // Traffic direction
	BalanceDestinationIds string  // filter for balance
	BalanceWeight         string  // filter for balance
	BalanceExpirationDate string  // filter for balance
	BalanceTimingTags     string  // filter for balance
	BalanceRatingSubject  string  // filter for balance
	BalanceCategories     string  // filter for balance
	BalanceSharedGroups   string  // filter for balance
	BalanceBlocker        string  // filter for balance
	BalanceDisabled       string  // filter for balance
	MinQueuedItems        int     // Trigger actions only if this number is hit (stats only)
	ActionsId             string  // Actions which will execute on threshold reached
	Weight                float64 // weight
}

// Used to rebuild a TPAccountActions (empty ActionTimingsId and ActionTriggersId) out of it's key in nosqldb
func NewTPAccountActionsFromKeyId(tpid, loadId, keyId string) (*TPAccountActions, error) {
	// *out:cgrates.org:1001
	s := strings.Split(keyId, ":")
	if len(s) != 2 {
		return nil, fmt.Errorf("Cannot parse key %s into AccountActions", keyId)
	}
	return &TPAccountActions{TPid: tpid, LoadId: loadId, Tenant: s[0], Account: s[1]}, nil
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
	return fmt.Sprintf("%s:%s", aa.Tenant, aa.Account)
}

func (aa *TPAccountActions) GetId() string {
	return aa.LoadId +
		CONCATENATED_KEY_SEP +
		aa.Tenant +
		CONCATENATED_KEY_SEP +
		aa.Account
}

func (aa *TPAccountActions) SetAccountActionsId(id string) error {
	ids := strings.Split(id, CONCATENATED_KEY_SEP)
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
	AccountIds []string
	Offset     int // Set the item offset
	Limit      int // Limit number of items retrieved
}

type ArgsCache struct {
	DestinationIDs        *[]string
	ReverseDestinationIDs *[]string
	RatingPlanIDs         *[]string
	RatingProfileIDs      *[]string
	ActionIDs             *[]string
	ActionPlanIDs         *[]string
	AccountActionPlanIDs  *[]string
	ActionTriggerIDs      *[]string
	SharedGroupIDs        *[]string
	DerivedChargerIDs     *[]string
	AliasIDs              *[]string
	ReverseAliasIDs       *[]string
	ResourceProfileIDs    *[]string
	ResourceIDs           *[]string
	StatsQueueIDs         *[]string
	StatsQueueProfileIDs  *[]string
	ThresholdIDs          *[]string
	ThresholdProfileIDs   *[]string
	FilterIDs             *[]string
	SupplierProfileIDs    *[]string
	AttributeProfileIDs   *[]string
	ChargerProfileIDs     *[]string
}

// Data used to do remote cache reloads via api
type AttrReloadCache struct {
	ArgsCache
	FlushAll bool // If provided, cache flush will be executed before any action
}

type ArgsCacheKeys struct {
	ArgsCache
	Paginator
}

type CacheKeys struct {
}

type AttrCacheStats struct { // Add in the future filters here maybe so we avoid counting complete cache
}

type CacheStats struct {
	Destinations        int
	ReverseDestinations int
	RatingPlans         int
	RatingProfiles      int
	Actions             int
	ActionPlans         int
	AccountActionPlans  int
	SharedGroups        int
	DerivedChargers     int
	Users               int
	Aliases             int
	ReverseAliases      int
	ResourceProfiles    int
	Resources           int
	StatQueues          int
	StatQueueProfiles   int
	Thresholds          int
	ThresholdProfiles   int
	Filters             int
	SupplierProfiles    int
	AttributeProfiles   int
	ChargerProfiles     int
}

type AttrExpFileCdrs struct {
	CdrFormat                  *string  // Cdr output file format <CdreCdrFormats>
	FieldSeparator             *string  // Separator used between fields
	ExportId                   *string  // Optional exportid
	ExportDir                  *string  // If provided it overwrites the configured export directory
	ExportFileName             *string  // If provided the output filename will be set to this
	ExportTemplate             *string  // Exported fields template  <""|fld1,fld2|*xml:instance_name>
	DataUsageMultiplyFactor    *float64 // Multiply data usage before export (eg: convert from KBytes to Bytes)
	SmsUsageMultiplyFactor     *float64 // Multiply sms usage before export (eg: convert from SMS unit to call duration for some billing systems)
	MmsUsageMultiplyFactor     *float64 // Multiply mms usage before export (eg: convert from MMS unit to call duration for some billing systems)
	GenericUsageMultiplyFactor *float64 // Multiply generic usage before export (eg: convert from GENERIC unit to call duration for some billing systems)
	CostMultiplyFactor         *float64 // Multiply the cost before export, eg: apply VAT
	CgrIds                     []string // If provided, it will filter based on the cgrids present in list
	MediationRunIds            []string // If provided, it will filter on mediation runid
	TORs                       []string // If provided, filter on TypeOfRecord
	CdrHosts                   []string // If provided, it will filter cdrhost
	CdrSources                 []string // If provided, it will filter cdrsource
	ReqTypes                   []string // If provided, it will fiter reqtype
	Tenants                    []string // If provided, it will filter tenant
	Categories                 []string // If provided, it will filter çategory
	Accounts                   []string // If provided, it will filter account
	Subjects                   []string // If provided, it will filter the rating subject
	DestinationPrefixes        []string // If provided, it will filter on destination prefix
	OrderIdStart               *int64   // Export from this order identifier
	OrderIdEnd                 *int64   // Export smaller than this order identifier
	TimeStart                  string   // If provided, it will represent the starting of the CDRs interval (>=)
	TimeEnd                    string   // If provided, it will represent the end of the CDRs interval (<)
	SkipErrors                 bool     // Do not export errored CDRs
	SkipRated                  bool     // Do not export rated CDRs
	SuppressCgrIds             bool     // Disable CgrIds reporting in reply/ExportedCgrIds and reply/UnexportedCgrIds
	Paginator
}

func (self *AttrExpFileCdrs) AsCDRsFilter(timezone string) (*CDRsFilter, error) {
	cdrFltr := &CDRsFilter{
		CGRIDs:              self.CgrIds,
		RunIDs:              self.MediationRunIds,
		NotRunIDs:           []string{MetaRaw}, // In exportv1 automatically filter out *raw CDRs
		ToRs:                self.TORs,
		OriginHosts:         self.CdrHosts,
		Sources:             self.CdrSources,
		RequestTypes:        self.ReqTypes,
		Tenants:             self.Tenants,
		Categories:          self.Categories,
		Accounts:            self.Accounts,
		Subjects:            self.Subjects,
		DestinationPrefixes: self.DestinationPrefixes,
		OrderIDStart:        self.OrderIdStart,
		OrderIDEnd:          self.OrderIdEnd,
		Paginator:           self.Paginator,
	}
	if len(self.TimeStart) != 0 {
		if answerTimeStart, err := ParseTimeDetectLayout(self.TimeStart, timezone); err != nil {
			return nil, err
		} else {
			cdrFltr.AnswerTimeStart = &answerTimeStart
		}
	}
	if len(self.TimeEnd) != 0 {
		if answerTimeEnd, err := ParseTimeDetectLayout(self.TimeEnd, timezone); err != nil {
			return nil, err
		} else {
			cdrFltr.AnswerTimeEnd = &answerTimeEnd
		}
	}
	if self.SkipRated {
		cdrFltr.MaxCost = Float64Pointer(-1.0)
	} else if self.SkipRated {
		cdrFltr.MinCost = Float64Pointer(0.0)
		cdrFltr.MaxCost = Float64Pointer(-1.0)
	}
	return cdrFltr, nil
}

type ExportedFileCdrs struct {
	ExportedFilePath          string            // Full path to the newly generated export file
	TotalRecords              int               // Number of CDRs to be exported
	TotalCost                 float64           // Sum of all costs in exported CDRs
	FirstOrderId, LastOrderId int64             // The order id of the last exported CDR
	ExportedCgrIds            []string          // List of successfuly exported cgrids in the file
	UnexportedCgrIds          map[string]string // Map of errored CDRs, map key is cgrid, value will be the error string
}

type AttrGetCdrs struct {
	CgrIds          []string // If provided, it will filter based on the cgrids present in list
	MediationRunIds []string // If provided, it will filter on mediation runid

	TORs                []string // If provided, filter on TypeOfRecord
	CdrHosts            []string // If provided, it will filter cdrhost
	CdrSources          []string // If provided, it will filter cdrsource
	ReqTypes            []string // If provided, it will fiter reqtype
	Directions          []string // If provided, it will fiter direction
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

func (self *AttrGetCdrs) AsCDRsFilter(timezone string) (*CDRsFilter, error) {
	cdrFltr := &CDRsFilter{
		CGRIDs:              self.CgrIds,
		RunIDs:              self.MediationRunIds,
		ToRs:                self.TORs,
		OriginHosts:         self.CdrHosts,
		Sources:             self.CdrSources,
		RequestTypes:        self.ReqTypes,
		Tenants:             self.Tenants,
		Categories:          self.Categories,
		Accounts:            self.Accounts,
		Subjects:            self.Subjects,
		DestinationPrefixes: self.DestinationPrefixes,
		OrderIDStart:        self.OrderIdStart,
		OrderIDEnd:          self.OrderIdEnd,
		Paginator:           self.Paginator,
		OrderBy:             self.OrderBy,
	}
	if len(self.TimeStart) != 0 {
		if answerTimeStart, err := ParseTimeDetectLayout(self.TimeStart, timezone); err != nil {
			return nil, err
		} else {
			cdrFltr.AnswerTimeStart = &answerTimeStart
		}
	}
	if len(self.TimeEnd) != 0 {
		if answerTimeEnd, err := ParseTimeDetectLayout(self.TimeEnd, timezone); err != nil {
			return nil, err
		} else {
			cdrFltr.AnswerTimeEnd = &answerTimeEnd
		}
	}
	if self.SkipRated {
		cdrFltr.MaxCost = Float64Pointer(-1.0)
	} else if self.SkipRated {
		cdrFltr.MinCost = Float64Pointer(0.0)
		cdrFltr.MaxCost = Float64Pointer(-1.0)
	}
	return cdrFltr, nil
}

type AttrRemCdrs struct {
	CgrIds []string // List of CgrIds to remove from storeDb
}

type AttrRateCdrs struct {
	CgrIds              []string // If provided, it will filter based on the cgrids present in list
	MediationRunIds     []string // If provided, it will filter on mediation runid
	TORs                []string // If provided, filter on TypeOfRecord
	CdrHosts            []string // If provided, it will filter cdrhost
	CdrSources          []string // If provided, it will filter cdrsource
	ReqTypes            []string // If provided, it will fiter reqtype
	Tenants             []string // If provided, it will filter tenant
	Categories          []string // If provided, it will filter çategory
	Accounts            []string // If provided, it will filter account
	Subjects            []string // If provided, it will filter the rating subject
	DestinationPrefixes []string // If provided, it will filter on destination prefix
	OrderIdStart        *int64   // Export from this order identifier
	OrderIdEnd          *int64   // Export smaller than this order identifier
	TimeStart           string   // If provided, it will represent the starting of the CDRs interval (>=)
	TimeEnd             string   // If provided, it will represent the end of the CDRs interval (<)
	RerateErrors        bool     // Rerate previous CDRs with errors (makes sense for reqtype rated and pseudoprepaid
	RerateRated         bool     // Rerate CDRs which were previously rated (makes sense for reqtype rated and pseudoprepaid)
	SendToStats         bool     // Set to true if the CDRs should be sent to stats server
}

func (attrRateCDRs *AttrRateCdrs) AsCDRsFilter(timezone string) (*CDRsFilter, error) {
	cdrFltr := &CDRsFilter{
		CGRIDs:              attrRateCDRs.CgrIds,
		RunIDs:              attrRateCDRs.MediationRunIds,
		OriginHosts:         attrRateCDRs.CdrHosts,
		Sources:             attrRateCDRs.CdrSources,
		ToRs:                attrRateCDRs.TORs,
		RequestTypes:        attrRateCDRs.ReqTypes,
		Tenants:             attrRateCDRs.Tenants,
		Categories:          attrRateCDRs.Categories,
		Accounts:            attrRateCDRs.Accounts,
		Subjects:            attrRateCDRs.Subjects,
		DestinationPrefixes: attrRateCDRs.DestinationPrefixes,
		OrderIDStart:        attrRateCDRs.OrderIdStart,
		OrderIDEnd:          attrRateCDRs.OrderIdEnd,
	}
	if aTime, err := ParseTimeDetectLayout(attrRateCDRs.TimeStart, timezone); err != nil {
		return nil, err
	} else if !aTime.IsZero() {
		cdrFltr.AnswerTimeStart = &aTime
	}
	if aTimeEnd, err := ParseTimeDetectLayout(attrRateCDRs.TimeEnd, timezone); err != nil {
		return nil, err
	} else if !aTimeEnd.IsZero() {
		cdrFltr.AnswerTimeEnd = &aTimeEnd
	}
	if attrRateCDRs.RerateErrors {
		cdrFltr.MinCost = Float64Pointer(-1.0)
		if !attrRateCDRs.RerateRated {
			cdrFltr.MaxCost = Float64Pointer(0.0)
		}
	} else if attrRateCDRs.RerateRated {
		cdrFltr.MinCost = Float64Pointer(0.0)
	}
	if attrRateCDRs.RerateErrors || attrRateCDRs.RerateRated {
		cdrFltr.NotRunIDs = append(cdrFltr.NotRunIDs, MetaRaw)
	}
	return cdrFltr, nil
}

type AttrLoadTpFromFolder struct {
	FolderPath string // Take files from folder absolute path
	DryRun     bool   // Do not write to database but parse only
	FlushDb    bool   // Flush previous data before loading new one
	Validate   bool   // Run structural checks on data
}

type AttrImportTPFromFolder struct {
	TPid         string
	FolderPath   string
	RunId        string
	CsvSeparator string
}

type AttrGetDestination struct {
	Id string
}

type AttrDerivedChargers struct {
	Direction, Tenant, Category, Account, Subject, Destination string
}

func NewTAFromAccountKey(accountKey string) (*TenantAccount, error) {
	accountSplt := strings.Split(accountKey, CONCATENATED_KEY_SEP)
	if len(accountSplt) != 2 {
		return nil, fmt.Errorf("Unsupported format for TenantAccount: %s", accountKey)
	}
	return &TenantAccount{accountSplt[0], accountSplt[1]}, nil
}

type TenantAccount struct {
	Tenant, Account string
}

func NewDTCSFromRPKey(rpKey string) (*DirectionTenantCategorySubject, error) {
	rpSplt := strings.Split(rpKey, CONCATENATED_KEY_SEP)
	if len(rpSplt) != 4 {
		return nil, fmt.Errorf("Unsupported format for DirectionTenantCategorySubject: %s", rpKey)
	}
	return &DirectionTenantCategorySubject{rpSplt[0], rpSplt[1], rpSplt[2], rpSplt[3]}, nil
}

type DirectionTenantCategorySubject struct {
	Direction, Tenant, Category, Subject string
}

type AttrCDRStatsReloadQueues struct {
	StatsQueueIds []string
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

// RPCCDRsFilter is a filter used in Rpc calls
// RPCCDRsFilter is slightly different than CDRsFilter by using string instead of Time filters
type RPCCDRsFilter struct {
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
	MinCost                *float64          // Start of the cost interval (>=)
	MaxCost                *float64          // End of the usage interval (<)
	OrderBy                string            // Ascendent/Descendent
	Paginator                                // Add pagination
}

func (self *RPCCDRsFilter) AsCDRsFilter(timezone string) (*CDRsFilter, error) {
	cdrFltr := &CDRsFilter{
		CGRIDs:                 self.CGRIDs,
		NotCGRIDs:              self.NotCGRIDs,
		RunIDs:                 self.RunIDs,
		NotRunIDs:              self.NotRunIDs,
		OriginIDs:              self.OriginIDs,
		NotOriginIDs:           self.NotOriginIDs,
		ToRs:                   self.ToRs,
		NotToRs:                self.NotToRs,
		OriginHosts:            self.OriginHosts,
		NotOriginHosts:         self.NotOriginHosts,
		Sources:                self.Sources,
		NotSources:             self.NotSources,
		RequestTypes:           self.RequestTypes,
		NotRequestTypes:        self.NotRequestTypes,
		Tenants:                self.Tenants,
		NotTenants:             self.NotTenants,
		Categories:             self.Categories,
		NotCategories:          self.NotCategories,
		Accounts:               self.Accounts,
		NotAccounts:            self.NotAccounts,
		Subjects:               self.Subjects,
		NotSubjects:            self.NotSubjects,
		DestinationPrefixes:    self.DestinationPrefixes,
		NotDestinationPrefixes: self.NotDestinationPrefixes,
		Costs:          self.Costs,
		NotCosts:       self.NotCosts,
		ExtraFields:    self.ExtraFields,
		NotExtraFields: self.NotExtraFields,
		OrderIDStart:   self.OrderIDStart,
		OrderIDEnd:     self.OrderIDEnd,
		MinUsage:       self.MinUsage,
		MaxUsage:       self.MaxUsage,
		MinCost:        self.MinCost,
		MaxCost:        self.MaxCost,
		Paginator:      self.Paginator,
		OrderBy:        self.OrderBy,
	}
	if len(self.SetupTimeStart) != 0 {
		if sTimeStart, err := ParseTimeDetectLayout(self.SetupTimeStart, timezone); err != nil {
			return nil, err
		} else {
			cdrFltr.SetupTimeStart = &sTimeStart
		}
	}
	if len(self.SetupTimeEnd) != 0 {
		if sTimeEnd, err := ParseTimeDetectLayout(self.SetupTimeEnd, timezone); err != nil {
			return nil, err
		} else {
			cdrFltr.SetupTimeEnd = &sTimeEnd
		}
	}
	if len(self.AnswerTimeStart) != 0 {
		if aTimeStart, err := ParseTimeDetectLayout(self.AnswerTimeStart, timezone); err != nil {
			return nil, err
		} else {
			cdrFltr.AnswerTimeStart = &aTimeStart
		}
	}
	if len(self.AnswerTimeEnd) != 0 {
		if aTimeEnd, err := ParseTimeDetectLayout(self.AnswerTimeEnd, timezone); err != nil {
			return nil, err
		} else {
			cdrFltr.AnswerTimeEnd = &aTimeEnd
		}
	}
	if len(self.CreatedAtStart) != 0 {
		if tStart, err := ParseTimeDetectLayout(self.CreatedAtStart, timezone); err != nil {
			return nil, err
		} else {
			cdrFltr.CreatedAtStart = &tStart
		}
	}
	if len(self.CreatedAtEnd) != 0 {
		if tEnd, err := ParseTimeDetectLayout(self.CreatedAtEnd, timezone); err != nil {
			return nil, err
		} else {
			cdrFltr.CreatedAtEnd = &tEnd
		}
	}
	if len(self.UpdatedAtStart) != 0 {
		if tStart, err := ParseTimeDetectLayout(self.UpdatedAtStart, timezone); err != nil {
			return nil, err
		} else {
			cdrFltr.UpdatedAtStart = &tStart
		}
	}
	if len(self.UpdatedAtEnd) != 0 {
		if tEnd, err := ParseTimeDetectLayout(self.UpdatedAtEnd, timezone); err != nil {
			return nil, err
		} else {
			cdrFltr.UpdatedAtEnd = &tEnd
		}
	}
	return cdrFltr, nil
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
	ActionPlanId     string
	ActionTriggersId string
	AllowNegative    *bool
	Disabled         *bool
	ReloadScheduler  bool
}

type AttrRemoveAccount struct {
	Tenant          string
	Account         string
	ReloadScheduler bool
}

type AttrGetSMASessions struct {
	SessionManagerIndex int // Index of the session manager queried, defaults to first in the list
}

type AttrGetCallCost struct {
	CgrId string // Unique id of the CDR
	RunId string // Run Id
}

type AttrRateCDRs struct {
	RPCCDRsFilter
	StoreCDRs     *bool
	SendToStatS   *bool // Set to true if the CDRs should be sent to stats server
	ReplicateCDRs *bool // Replicate results
}

type AttrSetBalance struct {
	Tenant         string
	Account        string
	BalanceType    string
	BalanceUUID    *string
	BalanceID      *string
	Directions     *string
	Value          *float64
	ExpiryTime     *string
	RatingSubject  *string
	Categories     *string
	DestinationIds *string
	TimingIds      *string
	Weight         *float64
	SharedGroups   *string
	Blocker        *bool
	Disabled       *bool
}

type TPResource struct {
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
	ActivationTime,
	ExpiryTime string
}

type AttrRLsCache struct {
	LoadID      string
	ResourceIDs []string
}

type ArgRSv1ResourceUsage struct {
	CGREvent
	UsageID  string // ResourceUsage Identifier
	UsageTTL *time.Duration
	Units    float64
}

func (args *ArgRSv1ResourceUsage) TenantID() string {
	return ConcatenatedKey(args.CGREvent.Tenant, args.UsageID)
}

type ArgsComputeFilterIndexes struct {
	Tenant       string
	Context      string
	AttributeIDs *[]string
	ResourceIDs  *[]string
	StatIDs      *[]string
	SupplierIDs  *[]string
	ThresholdIDs *[]string
	ChargerIDs   *[]string
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
	ActivationTime, ExpiryTime time.Time
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

// TPStats is used in APIs to manage remotely offline Stats config
type TPStats struct {
	TPid               string
	Tenant             string
	ID                 string
	FilterIDs          []string
	ActivationInterval *TPActivationInterval
	QueueLength        int
	TTL                string
	Metrics            []*MetricWithParams
	Blocker            bool // blocker flag to stop processing on filters matched
	Stored             bool
	Weight             float64
	MinItems           int
	ThresholdIDs       []string
}

type MetricWithParams struct {
	MetricID   string
	Parameters string
}

type TPThreshold struct {
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

type TPFilterProfile struct {
	TPid               string
	Tenant             string
	ID                 string
	Filters            []*TPFilter
	ActivationInterval *TPActivationInterval // Time when this limit becomes active and expires
}

type TPFilter struct {
	Type      string   // Filter type (*string, *timing, *rsr_filters, *cdr_stats)
	FieldName string   // Name of the field providing us the Values to check (used in case of some )
	Values    []string // Filter definition
}

type TPSupplier struct {
	ID                 string // SupplierID
	FilterIDs          []string
	AccountIDs         []string
	RatingPlanIDs      []string // used when computing price
	ResourceIDs        []string // queried in some strategies
	StatIDs            []string // queried in some strategies
	Weight             float64
	Blocker            bool
	SupplierParameters string
}

type TPSupplierProfile struct {
	TPid               string
	Tenant             string
	ID                 string
	FilterIDs          []string
	ActivationInterval *TPActivationInterval // Time when this limit becomes active and expires
	Sorting            string
	SortingParameters  []string
	Suppliers          []*TPSupplier
	Weight             float64
}

type TPAttribute struct {
	FieldName  string
	Initial    string
	Substitute string
	Append     bool
}

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
