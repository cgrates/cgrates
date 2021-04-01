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
	"fmt"
	"strings"
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

type v1ActionTrigger struct {
	Id                    string // for visual identification
	ThresholdType         string //*min_counter, *max_counter, *min_balance, *max_balance
	ThresholdValue        float64
	Recurrent             bool          // reset eexcuted flag each run
	MinSleep              time.Duration // Minimum duration between two executions in case of recurrent triggers
	BalanceId             string
	BalanceType           string
	BalanceDirection      string
	BalanceDestinationIds string    // filter for balance
	BalanceWeight         float64   // filter for balance
	BalanceExpirationDate time.Time // filter for balance
	BalanceTimingTags     string    // filter for balance
	BalanceRatingSubject  string    // filter for balance
	BalanceCategory       string    // filter for balance
	BalanceSharedGroup    string    // filter for balance
	Weight                float64
	ActionsId             string
	MinQueuedItems        int // Trigger actions only if this number is hit (stats only)
	Executed              bool
}

type v1ActionTriggers []*v1ActionTrigger

func (m *Migrator) migrateCurrentActionTrigger() (err error) {
	var ids []string
	ids, err = m.dmIN.DataManager().DataDB().GetKeysForPrefix(utils.ActionTriggerPrefix)
	if err != nil {
		return err
	}
	for _, id := range ids {
		idg := strings.TrimPrefix(id, utils.ActionTriggerPrefix)
		acts, err := m.dmIN.DataManager().GetActionTriggers(idg, true, utils.NonTransactional)
		if err != nil {
			return err
		}
		if acts == nil || m.dryRun {
			continue
		}
		if err := m.dmOut.DataManager().SetActionTriggers(idg, acts, utils.NonTransactional); err != nil {
			return err
		}
		m.stats[utils.ActionTriggers]++

	}
	return
}

func (m *Migrator) migrateV1ActionTrigger() (acts engine.ActionTriggers, err error) {
	var v1ACTs *v1ActionTriggers
	v1ACTs, err = m.dmIN.getV1ActionTriggers()
	if err != nil {
		return nil, err
	}
	if v1ACTs == nil {
		return nil, nil
	}
	for _, v1ac := range *v1ACTs {
		act := v1ac.AsActionTrigger()
		acts = append(acts, act)
	}
	if m.dryRun {
		return
	}
	return
}

func (m *Migrator) removeV1ActionTriggers() (err error) {
	var v1ACTs *v1ActionTriggers
	for {
		if v1ACTs, err = m.dmIN.getV1ActionTriggers(); err != nil && err != utils.ErrNoMoreData {
			return err
		}
		if v1ACTs == nil {
			return nil
		}
		if err = m.dmIN.remV1ActionTriggers(v1ACTs); err != nil {
			return err
		}
	}
}

func (m *Migrator) migrateActionTriggers() (err error) {
	var vrs engine.Versions
	current := engine.CurrentDataDBVersions()
	if vrs, err = m.getVersions(utils.ActionTriggers); err != nil {
		return
	}
	migrated := true
	migratedFrom := 0
	var v2 engine.ActionTriggers
	for {
		version := vrs[utils.ActionTriggers]
		migratedFrom = int(version)
		for {
			switch version {
			default:
				return fmt.Errorf("Unsupported version %v", version)
			case current[utils.ActionTriggers]:
				migrated = false
				if m.sameDataDB {
					break
				}
				if err = m.migrateCurrentActionTrigger(); err != nil {
					return
				}
			case 1:
				if v2, err = m.migrateV1ActionTrigger(); err != nil && err != utils.ErrNoMoreData {
					return
				}
				version = 2
			}
			if version == current[utils.ActionTriggers] || err == utils.ErrNoMoreData {
				break
			}
		}
		if err == utils.ErrNoMoreData || !migrated {
			break
		}
		if !m.dryRun {
			//set action triggers
			if err = m.dmOut.DataManager().SetActionTriggers(v2[0].ID, v2, utils.NonTransactional); err != nil {
				return
			}
		}
		m.stats[utils.ActionTriggers]++
	}
	if m.dryRun || !migrated {
		return nil
	}
	// remove old action triggers
	if !m.sameDataDB {
		if migratedFrom == 1 {
			if err = m.removeV1ActionTriggers(); err != nil {
				return
			}
		}
	}

	// All done, update version wtih current one
	if err = m.setVersions(utils.ActionTriggers); err != nil {
		return
	}

	return m.ensureIndexesDataDB(engine.ColAtr)
}

func (v1Act v1ActionTrigger) AsActionTrigger() (at *engine.ActionTrigger) {
	at = &engine.ActionTrigger{
		ID:             v1Act.Id,
		ThresholdType:  v1Act.ThresholdType,
		ThresholdValue: v1Act.ThresholdValue,
		Recurrent:      v1Act.Recurrent,
		MinSleep:       v1Act.MinSleep,
		Weight:         v1Act.Weight,
		ActionsID:      v1Act.ActionsId,
		MinQueuedItems: v1Act.MinQueuedItems,
		Executed:       v1Act.Executed,
	}
	bf := &engine.BalanceFilter{}
	if v1Act.BalanceId != "" {
		bf.ID = utils.StringPointer(v1Act.BalanceId)
	}
	if v1Act.BalanceType != "" {
		bf.Type = utils.StringPointer(v1Act.BalanceType)
	}
	if v1Act.BalanceRatingSubject != "" {
		bf.RatingSubject = utils.StringPointer(v1Act.BalanceRatingSubject)
	}
	if v1Act.BalanceDestinationIds != "" {
		bf.DestinationIDs = utils.StringMapPointer(utils.ParseStringMap(v1Act.BalanceDestinationIds))
	}
	if v1Act.BalanceTimingTags != "" {
		bf.TimingIDs = utils.StringMapPointer(utils.ParseStringMap(v1Act.BalanceTimingTags))
	}
	if v1Act.BalanceCategory != "" {
		bf.Categories = utils.StringMapPointer(utils.ParseStringMap(v1Act.BalanceCategory))
	}
	if v1Act.BalanceSharedGroup != "" {
		bf.SharedGroups = utils.StringMapPointer(utils.ParseStringMap(v1Act.BalanceSharedGroup))
	}
	if v1Act.BalanceWeight != 0 {
		bf.Weight = utils.Float64Pointer(v1Act.BalanceWeight)
	}
	if !v1Act.BalanceExpirationDate.IsZero() {
		bf.ExpirationDate = utils.TimePointer(v1Act.BalanceExpirationDate)
		at.ExpirationDate = v1Act.BalanceExpirationDate
		at.LastExecutionTime = v1Act.BalanceExpirationDate
		at.ActivationDate = v1Act.BalanceExpirationDate
	}
	at.Balance = bf
	if at.ThresholdType == "*min_counter" ||
		at.ThresholdType == "*max_counter" {
		at.ThresholdType = strings.Replace(at.ThresholdType, "_", "_event_", 1)
	}
	return
}
