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

func (v1AP v1ActionPlan) AsActionPlan() (ap engine.ActionPlan) {
	for idx, actionId := range v1AP.AccountIds {
		idElements := strings.Split(actionId, utils.CONCATENATED_KEY_SEP)
		if len(idElements) != 3 {
			continue
		}
		v1AP.AccountIds[idx] = fmt.Sprintf("%s:%s", idElements[1], idElements[2])
	}
	ap = engine.ActionPlan{
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
