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

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"gopkg.in/mgo.v2/bson"
)

type v1Action struct {
	Id               string
	ActionType       string
	BalanceType      string
	Direction        string
	ExtraParameters  string
	ExpirationString string
	Weight           float64
	Balance          *v1Balance
}

type v1Actions []*v1Action

func (m *Migrator) migrateActions() (err error) {
	switch m.dataDBType {
	case utils.REDIS:
		var acts engine.Actions
		var actv1keys []string
		actv1keys, err = m.dataDB.GetKeysForPrefix(utils.ACTION_PREFIX)
		if err != nil {
			return
		}
		for _, actv1key := range actv1keys {
			v1act, err := m.getV1ActionFromDB(actv1key)
			if err != nil {
				return err
			}
			act := v1act.AsAction()
			acts = append(acts, act)
		}
		if err := m.dataDB.SetActions(acts[0].Id, acts, utils.NonTransactional); err != nil {
			return err
		}

		// All done, update version wtih current one
		vrs := engine.Versions{utils.ACTION_PREFIX: engine.CurrentStorDBVersions()[utils.ACTION_PREFIX]}
		if err = m.dataDB.SetVersions(vrs, false); err != nil {
			return utils.NewCGRError(utils.Migrator,
				utils.ServerErrorCaps,
				err.Error(),
				fmt.Sprintf("error: <%s> when updating Accounts version into StorDB", err.Error()))
		}

		return
	case utils.MONGO:
		dataDB := m.dataDB.(*engine.MongoStorage)
		mgoDB := dataDB.DB()
		defer mgoDB.Session.Close()
		var acts engine.Actions
		var v1act v1Action
		iter := mgoDB.C(utils.ACTION_PREFIX).Find(nil).Iter()
		for iter.Next(&v1act) {
			act := v1act.AsAction()
			acts = append(acts, act)
		}
		if err := m.dataDB.SetActions(acts[0].Id, acts, utils.NonTransactional); err != nil {
			return err
		}
		// All done, update version wtih current one
		vrs := engine.Versions{utils.ACTION_PREFIX: engine.CurrentStorDBVersions()[utils.ACTION_PREFIX]}
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
			fmt.Sprintf("error: unsupported: <%s> for migrateActions method", m.dataDBType))
	}
}

func (m *Migrator) getV1ActionFromDB(key string) (v1act *v1Action, err error) {
	switch m.dataDBType {
	case utils.REDIS:
		dataDB := m.dataDB.(*engine.RedisStorage)
		if strVal, err := dataDB.Cmd("GET", key).Bytes(); err != nil {
			return nil, err
		} else {
			v1act := &v1Action{Id: key}
			if err := m.mrshlr.Unmarshal(strVal, v1act); err != nil {
				return nil, err
			}
			return v1act, nil
		}
	case utils.MONGO:
		dataDB := m.dataDB.(*engine.MongoStorage)
		mgoDB := dataDB.DB()
		defer mgoDB.Session.Close()
		v1act := new(v1Action)
		if err := mgoDB.C(utils.ACTION_PREFIX).Find(bson.M{"id": key}).One(v1act); err != nil {
			return nil, err
		}
		return v1act, nil
	default:
		return nil, utils.NewCGRError(utils.Migrator,
			utils.ServerErrorCaps,
			utils.UnsupportedDB,
			fmt.Sprintf("error: unsupported: <%s> for getV1ActionPlansFromDB method", m.dataDBType))
	}
}

func (v1Act v1Action) AsAction() (act *engine.Action) {
	act = &engine.Action{
		Id:               v1Act.Id,
		ActionType:       v1Act.ActionType,
		ExtraParameters:  v1Act.ExtraParameters,
		ExpirationString: v1Act.ExpirationString,
		Weight:           v1Act.Weight,
		Balance:          &engine.BalanceFilter{},
	}
	bf := act.Balance
	if v1Act.Balance.Uuid != "" {
		bf.Uuid = utils.StringPointer(v1Act.Balance.Uuid)
	}
	if v1Act.Balance.Id != "" {
		bf.ID = utils.StringPointer(v1Act.Balance.Id)
	}
	if v1Act.BalanceType != "" {
		bf.Type = utils.StringPointer(v1Act.BalanceType)
	}
	if v1Act.Balance.Value != 0 {
		bf.Value = &utils.ValueFormula{Static: v1Act.Balance.Value}
	}
	if v1Act.Balance.RatingSubject != "" {
		bf.RatingSubject = utils.StringPointer(v1Act.Balance.RatingSubject)
	}
	if v1Act.Balance.DestinationIds != "" {
		bf.DestinationIDs = utils.StringMapPointer(utils.ParseStringMap(v1Act.Balance.DestinationIds))
	}
	if v1Act.Balance.TimingIDs != "" {
		bf.TimingIDs = utils.StringMapPointer(utils.ParseStringMap(v1Act.Balance.TimingIDs))
	}
	if v1Act.Balance.Category != "" {
		bf.Categories = utils.StringMapPointer(utils.ParseStringMap(v1Act.Balance.Category))
	}
	if v1Act.Balance.SharedGroup != "" {
		bf.SharedGroups = utils.StringMapPointer(utils.ParseStringMap(v1Act.Balance.SharedGroup))
	}
	if v1Act.Balance.Weight != 0 {
		bf.Weight = utils.Float64Pointer(v1Act.Balance.Weight)
	}
	if v1Act.Balance.Disabled != false {
		bf.Disabled = utils.BoolPointer(v1Act.Balance.Disabled)
	}
	if !v1Act.Balance.ExpirationDate.IsZero() {
		bf.ExpirationDate = utils.TimePointer(v1Act.Balance.ExpirationDate)
	}
	bf.Timings = v1Act.Balance.Timings
	return
}
