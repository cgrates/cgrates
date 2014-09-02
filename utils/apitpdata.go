/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) 2012-2014 ITsysCOM GmbH

This program is free software: you can Storagetribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITH*out ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package utils

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

// Used to extract ids from stordb
type TPDistinctIds []string

func (tpdi TPDistinctIds) String() string {
	return strings.Join(tpdi, ",")
}

// To paginate stuff from stordb (e.g. ids)
type TPPagination struct {
	Page         int
	ItemsPerPage int
	SearchTerm   string
}

func (pag *TPPagination) GetLimit() (low, high int) {
	if pag.ItemsPerPage == 0 {
		return 0, math.MaxInt32
	}
	return pag.Page * pag.ItemsPerPage, pag.ItemsPerPage
}

type TPDestination struct {
	TPid          string   // Tariff plan id
	DestinationId string   // Destination id
	Prefixes      []string // Prefixes attached to this destination
}

// This file deals with tp_* data definition

type TPRate struct {
	TPid      string      // Tariff plan id
	RateId    string      // Rate id
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
}

// Used to set the durations we need out of strings
func (self *RateSlot) SetDurations() error {
	var err error
	if self.rateUnitDur, err = ParseDurationWithSecs(self.RateUnit); err != nil {
		return err
	}
	if self.rateIncrementDur, err = ParseDurationWithSecs(self.RateIncrement); err != nil {
		return err
	}
	if self.groupIntervalStartDur, err = ParseDurationWithSecs(self.GroupIntervalStart); err != nil {
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
	TPid              string             // Tariff plan id
	DestinationRateId string             // DestinationRate profile id
	DestinationRates  []*DestinationRate // Set of destinationid-rateid bindings
}

type DestinationRate struct {
	DestinationId    string // The destination identity
	RateId           string // The rate identity
	Rate             *TPRate
	RoundingMethod   string
	RoundingDecimals int
}

type ApierTPTiming struct {
	TPid      string // Tariff plan id
	TimingId  string // Timing id
	Years     string // semicolon separated list of years this timing is valid on, *any supported
	Months    string // semicolon separated list of months this timing is valid on, *any supported
	MonthDays string // semicolon separated list of month's days this timing is valid on, *any supported
	WeekDays  string // semicolon separated list of week day names this timing is valid on *any supported
	Time      string // String representing the time this timing starts on
}

type TPTiming struct {
	Id        string
	Years     Years
	Months    Months
	MonthDays MonthDays
	WeekDays  WeekDays
	StartTime string
}

type TPRatingPlan struct {
	TPid               string                 // Tariff plan id
	RatingPlanId       string                 // RatingPlan profile id
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
	// *out:cgrates.org:call:*any
	s := strings.Split(keyId, ":")
	// [*out cgrates.org call *any]
	if len(s) != 4 {
		return nil, fmt.Errorf("Cannot parse key %s into RatingProfile", keyId)
	}
	return &TPRatingProfile{TPid: tpid, LoadId: loadId, Tenant: s[1], Category: s[2], Direction: s[0], Subject: s[3]}, nil
}

type TPRatingProfile struct {
	TPid                  string                // Tariff plan id
	LoadId                string                // Gives ability to load specific RatingProfile based on load identifier, hence being able to keep history also in stordb
	Tenant                string                // Tenant's Id
	Category              string                // TypeOfRecord
	Direction             string                // Traffic direction, OUT is the only one supported for now
	Subject               string                // Rating subject, usually the same as account
	RatingPlanActivations []*TPRatingActivation // Activate rate profiles at specific time
}

// Used as key in nosql db (eg: redis)
func (self *TPRatingProfile) KeyId() string {
	return fmt.Sprintf("%s:%s:%s:%s", self.Direction, self.Tenant, self.Category, self.Subject)
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

type TPRatingActivation struct {
	ActivationTime   string // Time when this profile will become active, defined as unix epoch time
	RatingPlanId     string // Id of RatingPlan profile
	FallbackSubjects string // So we follow the api
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

type AttrTPRatingProfileIds struct {
	TPid      string // Tariff plan id
	Tenant    string // Tenant's Id
	Category  string // TypeOfRecord
	Direction string // Traffic direction
	Subject   string // Rating subject, usually the same as account
}

type TPActions struct {
	TPid      string      // Tariff plan id
	ActionsId string      // Actions id
	Actions   []*TPAction // Set of actions this Actions profile will perform
}

type TPAction struct {
	Identifier      string  // Identifier mapped in the code
	BalanceType     string  // Type of balance the action will operate on
	Direction       string  // Balance direction
	Units           float64 // Number of units to add/deduct
	ExpiryTime      string  // Time when the units will expire
	DestinationId   string  // Destination profile id
	RatingSubject   string  // Reference a rate subject defined in RatingProfiles
	Category        string  // category filter for balances
	SharedGroup     string  // Reference to a shared group
	BalanceWeight   float64 // Balance weight
	ExtraParameters string
	Weight          float64 // Action's weight
}

type TPSharedGroups struct {
	TPid           string
	SharedGroupsId string
	SharedGroups   []*TPSharedGroup
}

type TPSharedGroup struct {
	Account       string
	Strategy      string
	RatingSubject string
}

type TPLcrRules struct {
	TPid       string
	LcrRulesId string
	LcrRules   []*TPLcrRule
}

type TPLcrRule struct {
	Direction     string
	Tenant        string
	Customer      string
	DestinationId string
	Category      string
	Strategy      string
	Suppliers     string
	ActivatinTime string
	Weight        float64
}

type TPCdrStats struct {
	TPid       string
	CdrStatsId string
	CdrStats   []*TPCdrStat
}

type TPCdrStat struct {
	QueueLength       int
	TimeWindow        int64
	Metrics           string
	SetupInterval     string
	TOR               string
	CdrHost           string
	CdrSource         string
	ReqType           string
	Direction         string
	Tenant            string
	Category          string
	Account           string
	Subject           string
	DestinationPrefix string
	UsageInterval     string
	MediationRunIds   string
	RatedAccount      string
	RatedSubject      string
	CostInterval      string
	ActionTriggers    string
}

type TPDerivedChargers struct {
	TPid            string
	Loadid          string
	Direction       string
	Tenant          string
	Category        string
	Account         string
	Subject         string
	DerivedChargers []*TPDerivedCharger
}

func (tpdc TPDerivedChargers) GetDerivedChargesId() string {
	return tpdc.Loadid +
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
	tpdc.Loadid = ids[0]
	tpdc.Direction = ids[1]
	tpdc.Tenant = ids[2]
	tpdc.Category = ids[3]
	tpdc.Account = ids[4]
	tpdc.Subject = ids[5]
	return nil
}

type TPDerivedCharger struct {
	RunId            string
	RunFilters       string
	ReqTypeField     string
	DirectionField   string
	TenantField      string
	CategoryField    string
	AccountField     string
	SubjectField     string
	DestinationField string
	SetupTimeField   string
	AnswerTimeField  string
	UsageField       string
}

type TPActionPlan struct {
	TPid       string            // Tariff plan id
	Id         string            // ActionPlan id
	ActionPlan []*TPActionTiming // Set of ActionTiming bindings this profile will group
}

type TPActionTiming struct {
	ActionsId string  // Actions id
	TimingId  string  // Timing profile id
	Weight    float64 // Binding's weight
}

type TPActionTriggers struct {
	TPid             string             // Tariff plan id
	ActionTriggersId string             // Profile id
	ActionTriggers   []*TPActionTrigger // Set of triggers grouped in this profile

}

type TPActionTrigger struct {
	BalanceType           string        // Type of balance this trigger monitors
	Direction             string        // Traffic direction
	ThresholdType         string        // This threshold type
	ThresholdValue        float64       // Threshold
	Recurrent             bool          // reset executed flag each run
	MinSleep              time.Duration // Minimum duration between two executions in case of recurrent triggers
	DestinationId         string        // filter for balance
	BalanceWeight         float64       // filter for balance
	BalanceExpirationDate string        // filter for balance
	BalanceRatingSubject  string        // filter for balance
	BalanceCategory       string        // filter for balance
	BalanceSharedGroup    string        // filter for balance
	MinQueuedItems        int           // Trigger actions only if this number is hit (stats only)
	ActionsId             string        // Actions which will execute on threshold reached
	Weight                float64       // weight

}

// Used to rebuild a TPAccountActions (empty ActionTimingsId and ActionTriggersId) out of it's key in nosqldb
func NewTPAccountActionsFromKeyId(tpid, loadId, keyId string) (*TPAccountActions, error) {
	// *out:cgrates.org:1001
	s := strings.Split(keyId, ":")
	// [*out cgrates.org 1001]
	if len(s) != 3 {
		return nil, fmt.Errorf("Cannot parse key %s into AccountActions", keyId)
	}
	return &TPAccountActions{TPid: tpid, LoadId: loadId, Tenant: s[1], Account: s[2], Direction: s[0]}, nil
}

type TPAccountActions struct {
	TPid             string // Tariff plan id
	LoadId           string // LoadId, used to group actions on a load
	Tenant           string // Tenant's Id
	Account          string // Account name
	Direction        string // Traffic direction
	ActionPlanId     string // Id of ActionPlan profile to use
	ActionTriggersId string // Id of ActionTriggers profile to use
}

// Returns the id used in some nosql dbs (eg: redis)
func (self *TPAccountActions) KeyId() string {
	return fmt.Sprintf("%s:%s:%s", self.Direction, self.Tenant, self.Account)
}

func (aa *TPAccountActions) GetAccountActionsId() string {
	return aa.LoadId +
		CONCATENATED_KEY_SEP +
		aa.Direction +
		CONCATENATED_KEY_SEP +
		aa.Tenant +
		CONCATENATED_KEY_SEP +
		aa.Account
}

func (aa *TPAccountActions) SetAccountActionsId(id string) error {
	ids := strings.Split(id, CONCATENATED_KEY_SEP)
	if len(ids) != 4 {
		return fmt.Errorf("Wrong TP Account Action Id: %s", id)
	}
	aa.LoadId = ids[0]
	aa.Direction = ids[1]
	aa.Tenant = ids[2]
	aa.Account = ids[3]
	return nil
}

type AttrGetAccount struct {
	Tenant    string
	Account   string
	Direction string
}

// Data used to do remote cache reloads via api
type ApiReloadCache struct {
	DestinationIds   []string
	RatingPlanIds    []string
	RatingProfileIds []string
	ActionIds        []string
	SharedGroupIds   []string
	RpAliases        []string
	AccAliases       []string
	LCRIds           []string
	DerivedChargers  []string
}

type AttrCacheStats struct { // Add in the future filters here maybe so we avoid counting complete cache
}

type CacheStats struct {
	Destinations    int64
	RatingPlans     int64
	RatingProfiles  int64
	Actions         int64
	SharedGroups    int64
	RatingAliases   int64
	AccountAliases  int64
	DerivedChargers int64
}

type AttrCachedItemAge struct {
	Category string // Item's category, same name as .csv files without extension
	ItemId   string // Item's identity tag
}

type CachedItemAge struct {
	Destination     time.Duration
	RatingPlan      time.Duration
	RatingProfile   time.Duration
	Action          time.Duration
	SharedGroup     time.Duration
	RatingAlias     time.Duration
	AccountAlias    time.Duration
	DerivedChargers time.Duration
}

type AttrExpFileCdrs struct {
	CdrFormat               *string  // Cdr output file format <utils.CdreCdrFormats>
	FieldSeparator          *rune    // Separator used between fields
	ExportId                *string  // Optional exportid
	ExportDir               *string  // If provided it overwrites the configured export directory
	ExportFileName          *string  // If provided the output filename will be set to this
	ExportTemplate          *string  // Exported fields template  <""|fld1,fld2|*xml:instance_name>
	DataUsageMultiplyFactor *float64 // Multiply data usage before export (eg: convert from KBytes to Bytes)
	CostMultiplyFactor      *float64 // Multiply the cost before export, eg: apply VAT
	CostShiftDigits         *int     // If defined it will shift cost digits before applying rouding (eg: convert from Eur->cents), -1 to use general config ones
	RoundDecimals           *int     // Overwrite configured roundDecimals with this dynamically, -1 to use general config ones
	MaskDestinationId       *string  // Overwrite configured MaskDestId
	MaskLength              *int     // Overwrite configured MaskLength, -1 to use general config ones
	CgrIds                  []string // If provided, it will filter based on the cgrids present in list
	MediationRunIds         []string // If provided, it will filter on mediation runid
	TORs                    []string // If provided, filter on TypeOfRecord
	CdrHosts                []string // If provided, it will filter cdrhost
	CdrSources              []string // If provided, it will filter cdrsource
	ReqTypes                []string // If provided, it will fiter reqtype
	Directions              []string // If provided, it will fiter direction
	Tenants                 []string // If provided, it will filter tenant
	Categories              []string // If provided, it will filter çategory
	Accounts                []string // If provided, it will filter account
	Subjects                []string // If provided, it will filter the rating subject
	DestinationPrefixes     []string // If provided, it will filter on destination prefix
	RatedAccounts           []string // If provided, it will filter ratedaccount
	RatedSubjects           []string // If provided, it will filter the ratedsubject
	OrderIdStart            int64    // Export from this order identifier
	OrderIdEnd              int64    // Export smaller than this order identifier
	TimeStart               string   // If provided, it will represent the starting of the CDRs interval (>=)
	TimeEnd                 string   // If provided, it will represent the end of the CDRs interval (<)
	SkipErrors              bool     // Do not export errored CDRs
	SkipRated               bool     // Do not export rated CDRs
	SuppressCgrIds          bool     // Disable CgrIds reporting in reply/ExportedCgrIds and reply/UnexportedCgrIds
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
	CgrIds              []string // If provided, it will filter based on the cgrids present in list
	MediationRunIds     []string // If provided, it will filter on mediation runid
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
	OrderIdStart        int64    // Export from this order identifier
	OrderIdEnd          int64    // Export smaller than this order identifier
	TimeStart           string   // If provided, it will represent the starting of the CDRs interval (>=)
	TimeEnd             string   // If provided, it will represent the end of the CDRs interval (<)
	SkipErrors          bool     // Do not export errored CDRs
	SkipRated           bool     // Do not export rated CDRs
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
	Directions          []string // If provided, it will fiter direction
	Tenants             []string // If provided, it will filter tenant
	Categories          []string // If provided, it will filter çategory
	Accounts            []string // If provided, it will filter account
	Subjects            []string // If provided, it will filter the rating subject
	DestinationPrefixes []string // If provided, it will filter on destination prefix
	RatedAccounts       []string // If provided, it will filter ratedaccount
	RatedSubjects       []string // If provided, it will filter the ratedsubject
	OrderIdStart        int64    // Export from this order identifier
	OrderIdEnd          int64    // Export smaller than this order identifier
	TimeStart           string   // If provided, it will represent the starting of the CDRs interval (>=)
	TimeEnd             string   // If provided, it will represent the end of the CDRs interval (<)
	RerateErrors        bool     // Rerate previous CDRs with errors (makes sense for reqtype rated and pseudoprepaid
	RerateRated         bool     // Rerate CDRs which were previously rated (makes sense for reqtype rated and pseudoprepaid)
	SendToStats         bool     // Set to true if the CDRs should be sent to stats server
}

type AttrLoadTpFromFolder struct {
	FolderPath string // Take files from folder absolute path
	DryRun     bool   // Do not write to database but parse only
	FlushDb    bool   // Flush previous data before loading new one
}

type AttrGetDestination struct {
	Id string
}

type AttrDerivedChargers struct {
	Direction, Tenant, Category, Account, Subject string
}

func NewDTAFromAccountKey(accountKey string) (*DirectionTenantAccount, error) {
	accountSplt := strings.Split(accountKey, CONCATENATED_KEY_SEP)
	if len(accountSplt) != 3 {
		return nil, fmt.Errorf("Unsupported format for DirectionTenantAccount: %s", accountKey)
	}
	return &DirectionTenantAccount{accountSplt[0], accountSplt[1], accountSplt[2]}, nil
}

type DirectionTenantAccount struct {
	Direction, Tenant, Account string
}

type AttrCDRStatsReloadQueues struct {
	StatsQueueIds []string
}
