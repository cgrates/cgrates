/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2013 ITsysCOM

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
func NewRateSlot(connectFee, rate float64, rateUnit, rateIncrement, grpInterval, rndMethod string, rndDecimals int) (*RateSlot, error) {
	rs := &RateSlot{ConnectFee: connectFee, Rate: rate, RateUnit: rateUnit, RateIncrement: rateIncrement,
		GroupIntervalStart: grpInterval, RoundingMethod: rndMethod, RoundingDecimals: rndDecimals}
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
	RoundingMethod        string  // Use this method to round the cost
	RoundingDecimals      int     // Round the cost number of decimals
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
	DestinationId string // The destination identity
	RateId        string // The rate identity
	Rate          *TPRate
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
	return &TPRatingProfile{TPid: tpid, LoadId: loadId, Tenant: s[1], TOR: s[2], Direction: s[0], Subject: s[3]}, nil
}

type TPRatingProfile struct {
	TPid                  string                // Tariff plan id
	LoadId                string                // Gives ability to load specific RatingProfile based on load identifier, hence being able to keep history also in stordb
	Tenant                string                // Tenant's Id
	TOR                   string                // TypeOfRecord
	Direction             string                // Traffic direction, OUT is the only one supported for now
	Subject               string                // Rating subject, usually the same as account
	RatingPlanActivations []*TPRatingActivation // Activate rate profiles at specific time
}

// Used as key in nosql db (eg: redis)
func (self *TPRatingProfile) KeyId() string {
	return fmt.Sprintf("%s:%s:%s:%s", self.Direction, self.Tenant, self.TOR, self.Subject)
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
	TOR       string // TypeOfRecord
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
	SharedGroup     string  // Reference to a shared group
	BalanceWeight   float64 // Balance weight
	ExtraParameters string
	Weight          float64 // Action's weight
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
	BalanceType    string  // Type of balance this trigger monitors
	Direction      string  // Traffic direction
	ThresholdType  string  // This threshold type
	ThresholdValue float64 // Threshold
	Recurrent      bool    // reset executed flag each run
	DestinationId  string  // Id of the destination profile
	ActionsId      string  // Actions which will execute on threshold reached
	Weight         float64 // weight
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

// Data used to do remote cache reloads via api
type ApiReloadCache struct {
	DestinationIds   []string
	RatingPlanIds    []string
	RatingProfileIds []string
	ActionIds        []string
	SharedGroupIds   []string
	RpAliases        []string
	AccAliases       []string
	DerivedChargers  []string
}

type AttrCacheStats struct { // Add in the future filters here maybe so we avoid counting complete cache
}

type CacheStats struct {
	Destinations    int
	RatingPlans     int
	RatingProfiles  int
	Actions         int
	SharedGroups    int
	RatingAliases   int
	AccountAliases  int
	DerivedChargers int
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
	CdrFormat         string   // Cdr output file format <utils.CdreCdrFormats>
	ExportId          string   // Optional exportid
	ExportDir         string   // If provided it overwrites the configured export directory
	ExportFileName    string   // If provided the output filename will be set to this
	ExportTemplate    string   // Exported fields template  <""|fld1,fld2|*xml:instance_name>
	CostShiftDigits   int      // If defined it will shift cost digits before applying rouding (eg: convert from Eur->cents), -1 to use general config ones
	RoundDecimals     int      // Overwrite configured roundDecimals with this dynamically, -1 to use general config ones
	MaskDestinationId string   // Overwrite configured MaskDestId
	MaskLength        int      // Overwrite configured MaskLength, -1 to use general config ones
	CgrIds            []string // If provided, it will filter based on the cgrids present in list
	MediationRunId    []string // If provided, it will filter on mediation runid
	CdrHost           []string // If provided, it will filter cdrhost
	CdrSource         []string // If provided, it will filter cdrsource
	ReqType           []string // If provided, it will fiter reqtype
	Direction         []string // If provided, it will fiter direction
	Tenant            []string // If provided, it will filter tenant
	Tor               []string // If provided, it will filter tor
	Account           []string // If provided, it will filter account
	Subject           []string // If provided, it will filter the rating subject
	DestinationPrefix []string // If provided, it will filter on destination prefix
	OrderIdStart      int64    // Export from this order identifier
	OrderIdEnd        int64    // Export smaller than this order identifier
	TimeStart         string   // If provided, it will represent the starting of the CDRs interval (>=)
	TimeEnd           string   // If provided, it will represent the end of the CDRs interval (<)
	SkipErrors        bool     // Do not export errored CDRs
	SkipRated         bool     // Do not export rated CDRs
}

type ExportedFileCdrs struct {
	ExportedFilePath          string            // Full path to the newly generated export file
	TotalRecords              int               // Number of CDRs to be exported
	TotalCost                 float64           // Sum of all costs in exported CDRs
	FirstOrderId, LastOrderId int64             // The order id of the last exported CDR
	ExportedCgrIds            []string          // List of successfuly exported cgrids in the file
	UnexportedCgrIds          map[string]string // Map of errored CDRs, map key is cgrid, value will be the error string

}

type AttrRemCdrs struct {
	CgrIds []string // List of CgrIds to remove from storeDb
}

type AttrRateCdrs struct {
	TimeStart    string // Cdrs time start
	TimeEnd      string // Cdrs time end
	RerateErrors bool   // Rerate previous CDRs with errors (makes sense for reqtype rated and pseudoprepaid
	RerateRated  bool   // Rerate CDRs which were previously rated (makes sense for reqtype rated and pseudoprepaid)
}

type AttrLoadTpFromFolder struct {
	FolderPath string // Take files from folder absolute path
	DryRun     bool   // Do not write to database but parse only
	FlushDb    bool   // Flush previous data before loading new one
}

type AttrGetDestination struct {
	Id string
}
