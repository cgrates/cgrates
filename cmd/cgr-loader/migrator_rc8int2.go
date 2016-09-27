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
package main

import (
	"log"
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

type BalanceFilter2 struct {
	Uuid           *string
	ID             *string
	Type           *string
	Value          *float64
	Directions     *utils.StringMap
	ExpirationDate *time.Time
	Weight         *float64
	DestinationIDs *utils.StringMap
	RatingSubject  *string
	Categories     *utils.StringMap
	SharedGroups   *utils.StringMap
	TimingIDs      *utils.StringMap
	Timings        []*engine.RITiming
	Disabled       *bool
	Factor         *engine.ValueFactor
	Blocker        *bool
}

type Action2 struct {
	Id               string
	ActionType       string
	ExtraParameters  string
	Filter           string
	ExpirationString string // must stay as string because it can have relative values like 1month
	Weight           float64
	Balance          *BalanceFilter2
}

type Actions2 []*Action2

func (mig MigratorRC8) migrateActionsInt2() error {
	keys, err := mig.db.Cmd("KEYS", utils.ACTION_PREFIX+"*").List()
	if err != nil {
		return err
	}
	newAcsMap := make(map[string]engine.Actions, len(keys))
	for _, key := range keys {
		log.Printf("Migrating action: %s...", key)
		var oldAcs Actions2
		var values []byte
		if values, err = mig.db.Cmd("GET", key).Bytes(); err == nil {
			if err := mig.ms.Unmarshal(values, &oldAcs); err != nil {
				return err
			}
		}
		newAcs := make(engine.Actions, len(oldAcs))
		for index, oldAc := range oldAcs {
			a := &engine.Action{
				Id:               oldAc.Id,
				ActionType:       oldAc.ActionType,
				ExtraParameters:  oldAc.ExtraParameters,
				ExpirationString: oldAc.ExpirationString,
				Filter:           oldAc.Filter,
				Weight:           oldAc.Weight,
				Balance: &engine.BalanceFilter{
					Uuid:           oldAc.Balance.Uuid,
					ID:             oldAc.Balance.ID,
					Type:           oldAc.Balance.Type,
					Directions:     oldAc.Balance.Directions,
					ExpirationDate: oldAc.Balance.ExpirationDate,
					Weight:         oldAc.Balance.Weight,
					DestinationIDs: oldAc.Balance.DestinationIDs,
					RatingSubject:  oldAc.Balance.RatingSubject,
					Categories:     oldAc.Balance.Categories,
					SharedGroups:   oldAc.Balance.SharedGroups,
					TimingIDs:      oldAc.Balance.TimingIDs,
					Timings:        oldAc.Balance.Timings,
					Disabled:       oldAc.Balance.Disabled,
					Factor:         oldAc.Balance.Factor,
					Blocker:        oldAc.Balance.Blocker,
				},
			}
			if oldAc.Balance.Value != nil {
				a.Balance.Value = &utils.ValueFormula{Static: *oldAc.Balance.Value}
			}
			newAcs[index] = a
		}
		newAcsMap[key] = newAcs
	}
	// write data back
	for key, acs := range newAcsMap {
		result, err := mig.ms.Marshal(&acs)
		if err != nil {
			return err
		}
		if err = mig.db.Cmd("SET", key, result).Err; err != nil {
			return err
		}
	}
	return nil
}
