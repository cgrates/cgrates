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
	"strings"
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

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

func (v1Act v1ActionTriggers) AsActionTriggers() (at engine.ActionTriggers, err error) {
	at = make(engine.ActionTriggers, len(v1Act))
	for index, oldAct := range v1Act {
		atr := &engine.ActionTrigger{
			UniqueID:       oldAct.Id,
			ThresholdType:  oldAct.ThresholdType,
			ThresholdValue: oldAct.ThresholdValue,
			Recurrent:      oldAct.Recurrent,
			MinSleep:       oldAct.MinSleep,
			Weight:         oldAct.Weight,
			ActionsID:      oldAct.ActionsId,
			MinQueuedItems: oldAct.MinQueuedItems,
			Executed:       oldAct.Executed,
		}

		bf := &engine.BalanceFilter{}
		if oldAct.BalanceId != "" {
			bf.ID = utils.StringPointer(oldAct.BalanceId)
		}
		if oldAct.BalanceType != "" {
			bf.Type = utils.StringPointer(oldAct.BalanceType)
		}
		if oldAct.BalanceRatingSubject != "" {
			bf.RatingSubject = utils.StringPointer(oldAct.BalanceRatingSubject)
		}
		if oldAct.BalanceDirection != "" {
			bf.Directions = utils.StringMapPointer(utils.ParseStringMap(oldAct.BalanceDirection))
		}
		if oldAct.BalanceDestinationIds != "" {
			bf.DestinationIDs = utils.StringMapPointer(utils.ParseStringMap(oldAct.BalanceDestinationIds))
		}
		if oldAct.BalanceTimingTags != "" {
			bf.TimingIDs = utils.StringMapPointer(utils.ParseStringMap(oldAct.BalanceTimingTags))
		}
		if oldAct.BalanceCategory != "" {
			bf.Categories = utils.StringMapPointer(utils.ParseStringMap(oldAct.BalanceCategory))
		}
		if oldAct.BalanceSharedGroup != "" {
			bf.SharedGroups = utils.StringMapPointer(utils.ParseStringMap(oldAct.BalanceSharedGroup))
		}
		if oldAct.BalanceWeight != 0 {
			bf.Weight = utils.Float64Pointer(oldAct.BalanceWeight)
		}
		if oldAct.BalanceDisabled != false {
			bf.Disabled = utils.BoolPointer(oldAct.BalanceDisabled)
		}
		if !oldAct.BalanceExpirationDate.IsZero() {
			bf.ExpirationDate = utils.TimePointer(oldAct.BalanceExpirationDate)
		}
		atr.Balance = bf
		at[index] = atr
		if at[index].ThresholdType == "*min_counter" ||
			at[index].ThresholdType == "*max_counter" {
			at[index].ThresholdType = strings.Replace(at[index].ThresholdType, "_", "_event_", 1)
		}

	}
	return at, nil
}
