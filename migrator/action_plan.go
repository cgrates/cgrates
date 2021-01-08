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

type v1ActionPlan struct {
	Uuid       string // uniquely identify the timing
	Id         string // informative purpose only
	AccountIds []string
	Timing     *engine.RateInterval
	Weight     float64
	ActionsId  string
	actions    v1Actions
	stCache    time.Time // cached time of the next start
}

type v1ActionPlans []*v1ActionPlan

func (at *v1ActionPlan) IsASAP() bool {
	if at.Timing == nil {
		return false
	}
	return at.Timing.Timing.StartTime == utils.MetaASAP
}

func (m *Migrator) migrateCurrentActionPlans() (err error) {
	var ids []string
	ids, err = m.dmIN.DataManager().DataDB().GetKeysForPrefix(utils.ACTION_PLAN_PREFIX)
	if err != nil {
		return
	}
	for _, id := range ids {
		idg := strings.TrimPrefix(id, utils.ACTION_PLAN_PREFIX)
		var acts *engine.ActionPlan
		if acts, err = m.dmIN.DataManager().GetActionPlan(idg, true, utils.NonTransactional); err != nil {
			return
		}
		if acts == nil || m.dryRun {
			continue
		}
		if err = m.dmOut.DataManager().SetActionPlan(idg, acts, true, utils.NonTransactional); err != nil {
			return
		}
		if err = m.dmIN.DataManager().RemoveActionPlan(idg, utils.NonTransactional); err != nil {
			return
		}
		m.stats[utils.ActionPlans]++
	}
	return
}
func (m *Migrator) removeV1ActionPlans() (err error) {
	var v1 *v1ActionPlans
	for {
		if v1, err = m.dmIN.getV1ActionPlans(); err != nil && err != utils.ErrNoMoreData {
			return
		}
		if v1 == nil {
			return nil
		}
		if err = m.dmIN.remV1ActionPlans(v1); err != nil {
			return
		}
	}
}

func (m *Migrator) migrateV1ActionPlans() (v2 []*engine.ActionPlan, err error) {
	var v1APs *v1ActionPlans
	v1APs, err = m.dmIN.getV1ActionPlans()
	if err != nil {
		return nil, err
	}
	for _, v1ap := range *v1APs {
		v2 = append(v2, v1ap.AsActionPlan())
	}
	return
}

func (m *Migrator) migrateActionPlans() (err error) {
	var vrs engine.Versions
	current := engine.CurrentDataDBVersions()
	if vrs, err = m.getVersions(utils.ActionPlans); err != nil {
		return
	}
	if m.dmIN.DataManager().DataDB().GetStorageType() == utils.Redis { // if redis rebuild action plans indexes
		redisDB, can := m.dmIN.DataManager().DataDB().(*engine.RedisStorage)
		if !can {
			return fmt.Errorf("Storage type %s could not be casted to <*engine.RedisStorage>", m.dmIN.DataManager().DataDB().GetStorageType())
		}
		if err = redisDB.RebbuildActionPlanKeys(); err != nil {
			return
		}
	}
	migrated := true
	migratedFrom := 0
	var v3 []*engine.ActionPlan
	for {
		version := vrs[utils.ActionPlans]
		migratedFrom = int(version)
		for {
			switch version {
			default:
				return fmt.Errorf("Unsupported version %v", version)
			case current[utils.ActionPlans]:
				migrated = false
				if m.sameDataDB {
					break
				}
				if err = m.migrateCurrentActionPlans(); err != nil && err != utils.ErrNoMoreData {
					return
				}
			case 1:
				if v3, err = m.migrateV1ActionPlans(); err != nil && err != utils.ErrNoMoreData {
					return
				}
				version = 3
			case 2: // neded to rebuild action plan indexes for redis
				// All done, update version wtih current one
				vrs := engine.Versions{utils.ActionPlans: engine.CurrentDataDBVersions()[utils.ActionPlans]}
				if err = m.dmOut.DataManager().DataDB().SetVersions(vrs, false); err != nil {
					return utils.NewCGRError(utils.Migrator,
						utils.ServerErrorCaps,
						err.Error(),
						fmt.Sprintf("error: <%s> when updating ActionPlans version into dataDB", err.Error()))
				}
				version = 3
			}
			if version == current[utils.ActionPlans] || err == utils.ErrNoMoreData {
				break
			}
		}
		if err == utils.ErrNoMoreData || !migrated {
			break
		}

		if !m.dryRun {
			//set action plan
			for _, ap := range v3 {
				if err = m.dmOut.DataManager().SetActionPlan(ap.Id, ap, true, utils.NonTransactional); err != nil {
					return
				}
			}
		}
		m.stats[utils.ActionPlans]++
	}
	if m.dryRun || !migrated {
		return nil
	}
	// remove old action plans
	if !m.sameDataDB {
		if migratedFrom == 1 {
			if err = m.removeV1ActionPlans(); err != nil {
				return
			}
		}
	}

	// All done, update version wtih current one
	if err = m.setVersions(utils.ActionPlans); err != nil {
		return err
	}
	return m.ensureIndexesDataDB(engine.ColApl)
}

func (v1AP v1ActionPlan) AsActionPlan() (ap *engine.ActionPlan) {
	for idx, actionID := range v1AP.AccountIds {
		idElements := strings.Split(actionID, "_")
		if len(idElements) != 2 {
			continue
		}
		v1AP.AccountIds[idx] = idElements[1]
	}
	ap = &engine.ActionPlan{
		Id:         v1AP.Id,
		AccountIDs: make(utils.StringMap),
	}
	if x := v1AP.IsASAP(); !x {
		for _, accID := range v1AP.AccountIds {
			if _, exists := ap.AccountIDs[accID]; !exists {
				ap.AccountIDs[accID] = true
			}
		}
	}
	ap.ActionTimings = append(ap.ActionTimings, &engine.ActionTiming{
		Uuid:      utils.GenUUID(),
		Timing:    v1AP.Timing,
		ActionsID: v1AP.ActionsId,
		Weight:    v1AP.Weight,
	})
	return
}
