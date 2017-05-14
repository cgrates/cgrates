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

// ChargedTiming represents one timing attached to a charge
type ChargedTiming struct {
	Years     *utils.Years
	Months    *utils.Months
	MonthDays *utils.MonthDays
	WeekDays  *utils.WeekDays
	StartTime string
}

// RatingUnit represents one unit out of RatingPlan bounded to event
type RatingUnit struct {
	ConnectFee       float64
	RoudingMethod    string
	RoundingDecimals int
	MaxCost          float64
	MaxCostStrategy  string
	Timing           *ChargedTiming // This RatingUnit is bounded to specific timing profile
	Rates            []*Rate
}

// BalanceCharge represents one unit charged to a balance
type BalanceCharge struct {
	UUID        string // balance charged
	AccountID   string
	Rating      *RatingUnit    // special price applied on this balance
	Units       float64        // number of units charged
	ExtraCharge *BalanceCharge // used in cases when paying *voice with *monetary
}

// ChargingIncrement represents one unit charged inside an interval
type ChargingIncrement struct {
	Usage   time.Duration
	Cost    float64
	Rating  *RatingUnit
	Balance *BalanceCharge
}

// ChargingInterval represents one interval out of Usage providing charging info
// eg: PEAK vs OFFPEAK
type ChargingInterval struct {
	StartTime  *time.Time
	Increments []*ChargingIncrement
}

// EventCost holds cost for an event
type EventCost struct {
	CGRID   string
	RunID   string
	Cost    float64
	Usage   time.Duration
	Charges []*ChargingInterval
}

// EventCostDigest is an optimized EventCost with smaller footprint to be sent over network or stored
// not that human friendly as EventCost
type EventCostDigest struct {
	CGRID string
	RunID string
	Cost  float64
	Usage time.Duration
}
