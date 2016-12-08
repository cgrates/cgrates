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
package migrator

import (
	"time"

	"github.com/cgrates/cgrates/engine"
)

type CallCostMigrator interface {
	AsCallCost() (*engine.CallCost, error)
}

type v1CallCost struct {
	Direction, Category, Tenant, Subject, Account, Destination, TOR string
	Cost                                                            float64
	Timespans                                                       v1TimeSpans
}

type v1TimeSpans []*v1TimeSpan

type v1TimeSpan struct {
	TimeStart, TimeEnd                                         time.Time
	Cost                                                       float64
	RateInterval                                               *engine.RateInterval
	DurationIndex                                              time.Duration
	Increments                                                 v1Increments
	MatchedSubject, MatchedPrefix, MatchedDestId, RatingPlanId string
}

type v1Increments []*v1Increment

type v1Increment struct {
	Duration            time.Duration
	Cost                float64
	BalanceRateInterval *engine.RateInterval
	BalanceInfo         *v1BalanceInfo
	UnitInfo            *v1UnitInfo
	CompressFactor      int
}

type v1BalanceInfo struct {
	UnitBalanceUuid  string
	MoneyBalanceUuid string
	AccountId        string // used when debited from shared balance
}

type v1UnitInfo struct {
	DestinationId string
	Quantity      float64
	TOR           string
}

func (v1cc *v1CallCost) AsCallCost() (cc *engine.CallCost, err error) {
	cc = new(engine.CallCost)
	return
}
