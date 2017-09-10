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
	return at.Timing.Timing.StartTime == utils.ASAP
}

func (m *Migrator) migrateActionPlans() (err error) {
	var v1APs *v1ActionPlans
	for {
		v1APs, err = m.oldDataDB.getV1ActionPlans()
		if err != nil && err != utils.ErrNoMoreData {
			return err
		}
		if err == utils.ErrNoMoreData {
			break
		}
		if *v1APs != nil {
			for _, v1ap := range *v1APs {
				ap := v1ap.AsActionPlan()
				if err = m.dataDB.SetActionPlan(ap.Id, ap, true, utils.NonTransactional); err != nil {
					return err
				}
			}
		}
	}
	// All done, update version wtih current one
	vrs := engine.Versions{utils.ActionPlans: engine.CurrentDataDBVersions()[utils.ActionPlans]}
	if err = m.dataDB.SetVersions(vrs, false); err != nil {
		return utils.NewCGRError(utils.Migrator,
			utils.ServerErrorCaps,
			err.Error(),
			fmt.Sprintf("error: <%s> when updating ActionPlans version into dataDB", err.Error()))
	}
	return
}

func (v1AP v1ActionPlan) AsActionPlan() (ap *engine.ActionPlan) {
	for idx, actionId := range v1AP.AccountIds {
		idElements := strings.Split(actionId, "_")
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
