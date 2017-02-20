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
	"github.com/cgrates/cgrates/utils"
)

func (m *Migrator) migrateAccounts() (err error) {
	return
}

type v1Account struct {
	Id             string
	BalanceMap     map[string]v1BalanceChain
	UnitCounters   []*v1UnitsCounter
	ActionTriggers v1ActionTriggers
	AllowNegative  bool
	Disabled       bool
}

type v1BalanceChain []*v1Balance

type v1Balance struct {
	Uuid           string //system wide unique
	Id             string // account wide unique
	Value          float64
	ExpirationDate time.Time
	Weight         float64
	DestinationIds string
	RatingSubject  string
	Category       string
	SharedGroup    string
	Timings        []*engine.RITiming
	TimingIDs      string
	Disabled       bool
}

func (b *v1Balance) IsDefault() bool {
	return (b.DestinationIds == "" || b.DestinationIds == utils.ANY) &&
		b.RatingSubject == "" &&
		b.Category == "" &&
		b.ExpirationDate.IsZero() &&
		b.SharedGroup == "" &&
		b.Weight == 0 &&
		b.Disabled == false
}

type v1UnitsCounter struct {
	Direction   string
	BalanceType string
	//	Units     float64
	Balances v1BalanceChain // first balance is the general one (no destination)
}

type v1ActionTriggers []*v1ActionTrigger

type v1ActionTrigger struct {
	Id                    string
	ThresholdType         string
	ThresholdValue        float64
	Recurrent             bool
	MinSleep              time.Duration
	BalanceId             string
	BalanceType           string
	BalanceDirection      string
	BalanceDestinationIds string
	BalanceWeight         float64
	BalanceExpirationDate time.Time
	BalanceTimingTags     string
	BalanceRatingSubject  string
	BalanceCategory       string
	BalanceSharedGroup    string
	BalanceDisabled       bool
	Weight                float64
	ActionsId             string
	MinQueuedItems        int
	Executed              bool
}
