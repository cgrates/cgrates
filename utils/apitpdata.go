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

// This file deals with tp_* data definition

type TPRate struct {
	TPid      string     // Tariff plan id
	RateId    string     // Rate id
	RateSlots []RateSlot // One or more RateSlots
}

type RateSlot struct {
	ConnectFee       float64 // ConnectFee applied once the call is answered
	Rate             float64 // Rate applied
	RatedUnits       int     //  Number of billing units this rate applies to
	RateIncrements   int     // This rate will apply in increments of duration
	GroupInterval    int     // Group position
	RoundingMethod   string  // Use this method to round the cost
	RoundingDecimals int     // Round the cost number of decimals
	Weight           float64 // Rate's priority when dealing with grouped rates
}

type TPDestinationRate struct {
	TPid              string            // Tariff plan id
	DestinationRateId string            // DestinationRate profile id
	DestinationRates  []DestinationRate // Set of destinationid-rateid bindings
}

type DestinationRate struct {
	DestinationId string // The destination identity
	RateId        string // The rate identity
}

type TPDestRateTiming struct {
	TPid             string           // Tariff plan id
	DestRateTimingId string           // DestinationRate profile id
	DestRateTimings  []DestRateTiming // Set of destinationid-rateid bindings
}

type DestRateTiming struct {
	DestRatesId string  // The DestinationRate identity
	TimingId    string  // The timing identity
	Weight      float64 // Binding priority taken into consideration when more DestinationRates are active on a time slot
}

type TPRatingProfile struct {
	TPid                 string             // Tariff plan id
	RatingProfileId      string             // RatingProfile id
	Tenant               string             // Tenant's Id
	TOR                  string             // TypeOfRecord
	Direction            string             // Traffic direction, OUT is the only one supported for now
	Subject              string             // Rating subject, usually the same as account
	RatesFallbackSubject string             // Fallback on this subject if rates not found for destination
	RatingActivations    []RatingActivation // Activate rate profiles at specific time
}

type RatingActivation struct {
	ActivationTime   int64  // Time when this profile will become active, defined as unix epoch time
	DestRateTimingId string // Id of DestRateTiming profile
}

type AttrTPRatingProfileIds struct {
	TPid      string // Tariff plan id
	Tenant    string // Tenant's Id
	TOR       string // TypeOfRecord
	Direction string // Traffic direction
	Subject   string // Rating subject, usually the same as account
}

type TPActions struct {
	TPid      string   // Tariff plan id
	ActionsId string   // Actions id
	Actions   []Action // Set of actions this Actions profile will perform
}

type Action struct {
	Identifier     string  // Identifier mapped in the code
	BalanceType    string  // Type of balance the action will operate on
	Direction      string  // Balance direction
	Units          float64 // Number of units to add/deduct
	ExpiryTime int64   // Time when the units will expire
	DestinationId  string  // Destination profile id
	RateType       string  // Type of rate <*absolute|*percent>
	Rate           float64 // Price value
	MinutesWeight  float64 // Minutes weight
	Weight         float64 // Action's weight
}

type ApiTPActionTimings struct {
	TPid            string            // Tariff plan id
	ActionTimingsId string            // ActionTimings id
	ActionTimings   []ApiActionTiming // Set of ActionTiming bindings this profile will group
}

type ApiActionTiming struct {
	ActionsId string  // Actions id
	TimingId  string  // Timing profile id
	Weight    float64 // Binding's weight
}

type ApiTPActionTriggers struct {
	TPid             string             // Tariff plan id
	ActionTriggersId string             // Profile id
	ActionTriggers   []ApiActionTrigger // Set of triggers grouped in this profile

}

type ApiActionTrigger struct {
	BalanceType      string  // Type of balance this trigger monitors
	Direction      string  // Traffic direction
	ThresholdType  string  // This threshold type
	ThresholdValue float64 // Threshold
	DestinationId  string  // Id of the destination profile
	ActionsId      string  // Actions which will execute on threshold reached
	Weight         float64 // weight
}

type ApiTPAccountActions struct {
	TPid             string // Tariff plan id
	AccountActionsId string // AccountActions id
	Tenant           string // Tenant's Id
	Account          string // Account name
	Direction        string // Traffic direction
	ActionTimingsId  string // Id of ActionTimings profile to use
	ActionTriggersId string // Id of ActionTriggers profile to use
}
