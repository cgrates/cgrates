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
	"time"
)

// This file deals with tp_* data definition

type TPRate struct {
	TPid      string      // Tariff plan id
	RateId    string      // Rate id
	RateSlots []*RateSlot // One or more RateSlots
}

// Needed so we make sure we always use SetDurations() on a newly created value
func NewRateSlot( connectFee, rate float64, rateUnit, rateIncrement, grpInterval, rndMethod string, rndDecimals int ) (*RateSlot, error) {
	rs := &RateSlot{ ConnectFee: connectFee, Rate: rate, RateUnit: rateUnit, RateIncrement: rateIncrement, 
			GroupIntervalStart: grpInterval, RoundingMethod: rndMethod, RoundingDecimals: rndDecimals }
	if err := rs.SetDurations(); err != nil {
		return nil, err
	}
	return rs, nil
}
		

type RateSlot struct {
	ConnectFee         float64       // ConnectFee applied once the call is answered
	Rate               float64       // Rate applied
	RateUnit           string //  Number of billing units this rate applies to
	RateIncrement      string // This rate will apply in increments of duration
	GroupIntervalStart string        // Group position
	RoundingMethod     string        // Use this method to round the cost
	RoundingDecimals   int           // Round the cost number of decimals
	rateUnitDur        time.Duration
	rateIncrementDur   time.Duration
	groupIntervalStartDur   time.Duration
}

// Used to set the durations we need out of strings
func(self *RateSlot) SetDurations() error {
	var err error
	if self.rateUnitDur, err = time.ParseDuration(self.RateUnit); err != nil {
		return err
	}
	if self.rateIncrementDur, err = time.ParseDuration(self.RateIncrement); err != nil {
		return err
	}
	if self.groupIntervalStartDur, err = time.ParseDuration(self.GroupIntervalStart); err != nil {
		return err
	}
	return nil
}
func(self *RateSlot) RateUnitDuration() time.Duration {
	return self.rateUnitDur
}
func(self *RateSlot) RateIncrementDuration() time.Duration {
	return self.rateIncrementDur
}
func(self *RateSlot) GroupIntervalStartDuration() time.Duration {
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

type TPTiming struct {
	Id        string
	Years     Years
	Months    Months
	MonthDays MonthDays
	WeekDays  WeekDays
	StartTime string
}

type TPRatingPlan struct {
	TPid         string        // Tariff plan id
	RatingPlanId string        // RatingPlan profile id
	RatingPlans  []*RatingPlan // Set of destinationid-rateid bindings
}

type RatingPlan struct {
	DestinationRatesId string  // The DestinationRate identity
	TimingId    string  // The timing identity
	Weight      float64 // Binding priority taken into consideration when more DestinationRates are active on a time slot
	timing      *TPTiming // Not exporting it via JSON
}

func(self *RatingPlan) SetTiming(tm *TPTiming) {
	self.timing = tm
}

func(self *RatingPlan) Timing() *TPTiming {
	return self.timing
}

type TPRatingProfile struct {
	TPid                  string // Tariff plan id
	RatingProfileId       string // RatingProfile id
	Tenant                string   // Tenant's Id
	TOR                   string   // TypeOfRecord
	Direction             string   // Traffic direction, OUT is the only one supported for now
	Subject               string   // Rating subject, usually the same as account
	RatingPlanActivations []*TPRatingActivation // Activate rate profiles at specific time
}

type TPRatingActivation struct {
	ActivationTime   string // Time when this profile will become active, defined as unix epoch time
	RatingPlanId string // Id of RatingPlan profile
	FallbackSubjects     string // So we follow the api
}

type AttrTPRatingProfileIds struct {
	TPid      string // Tariff plan id
	Tenant    string // Tenant's Id
	TOR       string // TypeOfRecord
	Direction string // Traffic direction
	Subject   string // Rating subject, usually the same as account
}

type TPActions struct {
	TPid      string    // Tariff plan id
	ActionsId string    // Actions id
	Actions   []*TPAction // Set of actions this Actions profile will perform
}

type TPAction struct {
	Identifier      string  // Identifier mapped in the code
	BalanceType     string  // Type of balance the action will operate on
	Direction       string  // Balance direction
	Units           float64 // Number of units to add/deduct
	ExpiryTime      string  // Time when the units will expire\
	DestinationId   string  // Destination profile id
	RatingSubject   string  // Reference a rate subject defined in RatingProfiles
	BalanceWeight   float64 // Balance weight
	ExtraParameters string
	Weight          float64 // Action's weight
}

type TPActionTimings struct {
	TPid            string             // Tariff plan id
	ActionTimingsId string             // ActionTimings id
	ActionTimings   []*TPActionTiming // Set of ActionTiming bindings this profile will group
}

type TPActionTiming struct {
	ActionsId string  // Actions id
	TimingId  string  // Timing profile id
	Weight    float64 // Binding's weight
}

type TPActionTriggers struct {
	TPid             string              // Tariff plan id
	ActionTriggersId string              // Profile id
	ActionTriggers   []*TPActionTrigger // Set of triggers grouped in this profile

}

type TPActionTrigger struct {
	BalanceType    string  // Type of balance this trigger monitors
	Direction      string  // Traffic direction
	ThresholdType  string  // This threshold type
	ThresholdValue float64 // Threshold
	DestinationId  string  // Id of the destination profile
	ActionsId      string  // Actions which will execute on threshold reached
	Weight         float64 // weight
}

type TPAccountActions struct {
	TPid             string // Tariff plan id
	AccountActionsId string // AccountActions id, used to group actions on a load
	Tenant           string // Tenant's Id
	Account          string // Account name
	Direction        string // Traffic direction
	ActionTimingsId  string // Id of ActionTimings profile to use
	ActionTriggersId string // Id of ActionTriggers profile to use
}


// Data used to do remote cache reloads via api
type ApiReloadCache struct {
	DestinationIds      []string
	RatingPlanIds       []string
}

