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
package engine

import (
	"time"

	"github.com/cgrates/cgrates/utils"
)

// EventCost
type EventCost struct {
	CGRID           string
	RunID           string
	Cost            float64
	Usage           time.Duration
	Charges         []*ChargingInterval
	IntervalDetails map[string]*ChrgIntervDetail
	RatingUnits     map[string]*RatingUnit
	Rates           map[string][]*Rate
	Timings         map[string]*ChargedTiming
}

// ChargingInterval represents one interval out of Usage providing charging info
// eg: PEAK vs OFFPEAK
type ChargingInterval struct {
	StartTime           *time.Time
	IntervalDetailsUUID string               // reference to CIntervDetails
	RatingUUID          string               // reference to RatingUnit
	Increments          []*ChargingIncrement // specific increments applied to this interval
	CompressFactor      int
}

// ChargingIncrement represents one unit charged inside an interval
type ChargingIncrement struct {
	Usage             time.Duration
	Cost              float64
	BalanceChargeUUID string
	CompressFactor    int
}

// BalanceCharge represents one unit charged to a balance
type BalanceCharge struct {
	AccountID       string  // keep reference for shared balances
	BalanceUUID     string  // balance charged
	RatingUUID      string  // special price applied on this balance
	Units           float64 // number of units charged
	ExtraChargeUUID string  // used in cases when paying *voice with *monetary
}

type ChrgIntervDetail struct {
	Subject           string // matched subject
	DestinationPrefix string // matched destination prefix
	DestinationID     string // matched destinationID
	RatingPlanID      string // matched ratingPlanID

}

// ChargedTiming represents one timing attached to a charge
type ChargedTiming struct {
	Years     *utils.Years
	Months    *utils.Months
	MonthDays *utils.MonthDays
	WeekDays  *utils.WeekDays
	StartTime string
}

// RatingUnit represents one unit out of RatingPlan matching for an event
type RatingUnit struct {
	ConnectFee       float64
	RoudingMethod    string
	RoundingDecimals int
	MaxCost          float64
	MaxCostStrategy  string
	TimingUUID       string // This RatingUnit is bounded to specific timing profile
	RatesUUID        string
}
