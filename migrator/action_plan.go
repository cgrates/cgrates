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

	"gopkg.in/mgo.v2/bson"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// const (
// 	v1AccountDBPrefix = "ubl_"
// )

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
	switch m.dataDBType {
	case utils.REDIS:
		var apsv1keys []string
		apsv1keys, err = m.dataDB.GetKeysForPrefix(utils.ACTION_PLAN_PREFIX)
		if err != nil {
			return
		}
		for _, apsv1key := range apsv1keys {
			v1aps, err := m.getV1ActionPlansFromDB(apsv1key)
			if err != nil {
				return err
			}
			aps := v1aps.AsActionPlan()
			if err = m.dataDB.SetActionPlan(aps.Id, aps, true, utils.NonTransactional); err != nil {
				return err
			}
		}
		// All done, update version wtih current one
		vrs := engine.Versions{utils.ACTION_PLAN_PREFIX: engine.CurrentStorDBVersions()[utils.ACTION_PLAN_PREFIX]}
		if err = m.dataDB.SetVersions(vrs, false); err != nil {
			return utils.NewCGRError(utils.Migrator,
				utils.ServerErrorCaps,
				err.Error(),
				fmt.Sprintf("error: <%s> when updating ActionPlans version into StorDB", err.Error()))
		}
		return
	case utils.MONGO:
		dataDB := m.dataDB.(*engine.MongoStorage)
		mgoDB := dataDB.DB()
		defer mgoDB.Session.Close()
		var acp v1ActionPlan
		iter := mgoDB.C(utils.ACTION_PLAN_PREFIX).Find(nil).Iter()
		for iter.Next(&acp) {
			aps := acp.AsActionPlan()
			if err = m.dataDB.SetActionPlan(aps.Id, aps, true, utils.NonTransactional); err != nil {
				return err
			}
		}
		// All done, update version wtih current one
		vrs := engine.Versions{utils.Accounts: engine.CurrentStorDBVersions()[utils.Accounts]}
		if err = m.dataDB.SetVersions(vrs, false); err != nil {
			return utils.NewCGRError(utils.Migrator,
				utils.ServerErrorCaps,
				err.Error(),
				fmt.Sprintf("error: <%s> when updating Accounts version into StorDB", err.Error()))
		}
		return
	default:
		return utils.NewCGRError(utils.Migrator,
			utils.ServerErrorCaps,
			utils.UnsupportedDB,
			fmt.Sprintf("error: unsupported: <%s> for migrateActionPlans method", m.dataDBType))
	}
}

func (m *Migrator) getV1ActionPlansFromDB(key string) (v1aps *v1ActionPlan, err error) {
	switch m.dataDBType {
	case utils.REDIS:
		dataDB := m.dataDB.(*engine.RedisStorage)
		if strVal, err := dataDB.Cmd("GET", key).Bytes(); err != nil {
			return nil, err
		} else {
			v1aps := &v1ActionPlan{Id: key}
			if err := m.mrshlr.Unmarshal(strVal, v1aps); err != nil {
				return nil, err
			}
			return v1aps, nil
		}
	case utils.MONGO:
		dataDB := m.dataDB.(*engine.MongoStorage)
		mgoDB := dataDB.DB()
		defer mgoDB.Session.Close()
		v1aps := new(v1ActionPlan)
		if err := mgoDB.C(utils.ACTION_PLAN_PREFIX).Find(bson.M{"id": key}).One(v1aps); err != nil {
			return nil, err
		}
		return v1aps, nil
	default:
		return nil, utils.NewCGRError(utils.Migrator,
			utils.ServerErrorCaps,
			utils.UnsupportedDB,
			fmt.Sprintf("error: unsupported: <%s> for getV1ActionPlansFromDB method", m.dataDBType))
	}
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
