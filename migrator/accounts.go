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
	"log"
	"strings"
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"gopkg.in/mgo.v2/bson"
)

const (
	v1AccountDBPrefix = "ubl_"
	v1AccountTBL      = "userbalances"
)

func (m *Migrator) migrateAccounts() (err error) {
	switch m.dataDBType {
	case utils.REDIS:
		var acntV1Keys []string
		acntV1Keys, err = m.dataDB.GetKeysForPrefix(v1AccountDBPrefix)
		if err != nil {
			return
		}
		for _, acntV1Key := range acntV1Keys {
			v1Acnt, err := m.getV1AccountFromDB(acntV1Key)
			if err != nil {
				return err
			}
			if v1Acnt != nil {
				acnt := v1Acnt.AsAccount()
				if err = m.dataDB.SetAccount(acnt); err != nil {
					return err
				}
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
	case utils.MONGO:
		dataDB := m.dataDB.(*engine.MongoStorage)
		mgoDB := dataDB.DB()
		defer mgoDB.Session.Close()
		var accn v1Account
		iter := mgoDB.C(v1AccountDBPrefix).Find(nil).Iter()
		for iter.Next(&accn) {
			if acnt := accn.AsAccount(); acnt != nil {
				if err = m.dataDB.SetAccount(acnt); err != nil {
					return err
				}
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
			fmt.Sprintf("error: unsupported: <%s> for migrateAccounts method", m.dataDBType))
	}
}

func (m *Migrator) getV1AccountFromDB(key string) (*v1Account, error) {
	switch m.dataDBType {
	case utils.REDIS:
		dataDB := m.dataDB.(*engine.RedisStorage)
		if strVal, err := dataDB.Cmd("GET", key).Bytes(); err != nil {
			return nil, err
		} else {
			v1Acnt := &v1Account{Id: key}
			if err := m.mrshlr.Unmarshal(strVal, v1Acnt); err != nil {
				return nil, err
			}
			return v1Acnt, nil
		}
	case utils.MONGO:
		dataDB := m.dataDB.(*engine.MongoStorage)
		mgoDB := dataDB.DB()
		defer mgoDB.Session.Close()
		v1Acnt := new(v1Account)
		if err := mgoDB.C(v1AccountTBL).Find(bson.M{"id": key}).One(v1Acnt); err != nil {
			return nil, err
		}
		return v1Acnt, nil
	default:
		return nil, utils.NewCGRError(utils.Migrator,
			utils.ServerErrorCaps,
			utils.UnsupportedDB,
			fmt.Sprintf("error: unsupported: <%s> for getV1AccountFromDB method", m.dataDBType))
	}
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
type v1UnitsCounter struct {
	Direction   string
	BalanceType string
	//	Units     float64
	Balances v1BalanceChain // first balance is the general one (no destination)
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

func (v1Acc v1Account) AsAccount() (ac *engine.Account) {
	// transfer data into new structure
	ac = &engine.Account{
		ID:             v1Acc.Id,
		BalanceMap:     make(map[string]engine.Balances, len(v1Acc.BalanceMap)),
		UnitCounters:   make(engine.UnitCounters, len(v1Acc.UnitCounters)),
		ActionTriggers: make(engine.ActionTriggers, len(v1Acc.ActionTriggers)),
		AllowNegative:  v1Acc.AllowNegative,
		Disabled:       v1Acc.Disabled,
	}
	idElements := strings.Split(ac.ID, utils.CONCATENATED_KEY_SEP)
	if len(idElements) != 3 {
		log.Printf("Malformed account ID %s", v1Acc.Id)
	}
	ac.ID = fmt.Sprintf("%s:%s", idElements[1], idElements[2])
	// balances
	for oldBalKey, oldBalChain := range v1Acc.BalanceMap {
		keyElements := strings.Split(oldBalKey, "*")
		newBalKey := "*" + keyElements[1]
		newBalDirection := "*" + idElements[0]
		ac.BalanceMap[newBalKey] = make(engine.Balances, len(oldBalChain))
		for index, oldBal := range oldBalChain {
			// check default to set new id
			ac.BalanceMap[newBalKey][index] = &engine.Balance{
				Uuid:           oldBal.Uuid,
				ID:             oldBal.Id,
				Value:          oldBal.Value,
				Directions:     utils.ParseStringMap(newBalDirection),
				ExpirationDate: oldBal.ExpirationDate,
				Weight:         oldBal.Weight,
				DestinationIDs: utils.ParseStringMap(oldBal.DestinationIds),
				RatingSubject:  oldBal.RatingSubject,
				Categories:     utils.ParseStringMap(oldBal.Category),
				SharedGroups:   utils.ParseStringMap(oldBal.SharedGroup),
				Timings:        oldBal.Timings,
				TimingIDs:      utils.ParseStringMap(oldBal.TimingIDs),
				Disabled:       oldBal.Disabled,
				Factor:         engine.ValueFactor{},
			}
		}
	}
	// unit counters
	for _, oldUc := range v1Acc.UnitCounters {
		newUc := &engine.UnitCounter{Counters: make(engine.CounterFilters, len(oldUc.Balances))}
		for index, oldUcBal := range oldUc.Balances {
			bf := &engine.BalanceFilter{}
			if oldUcBal.Uuid != "" {
				bf.Uuid = utils.StringPointer(oldUcBal.Uuid)
			}
			if oldUcBal.Id != "" {
				bf.ID = utils.StringPointer(oldUcBal.Id)
			}
			if oldUc.BalanceType != "" {
				bf.Type = utils.StringPointer(oldUc.BalanceType)
			}
			if oldUc.Direction != "" {
				bf.Directions = utils.StringMapPointer(utils.ParseStringMap(oldUc.Direction))
			}
			if !oldUcBal.ExpirationDate.IsZero() {
				bf.ExpirationDate = utils.TimePointer(oldUcBal.ExpirationDate)
			}
			if oldUcBal.Weight != 0 {
				bf.Weight = utils.Float64Pointer(oldUcBal.Weight)
			}
			if oldUcBal.DestinationIds != "" {
				bf.DestinationIDs = utils.StringMapPointer(utils.ParseStringMap(oldUcBal.DestinationIds))
			}
			if oldUcBal.RatingSubject != "" {
				bf.RatingSubject = utils.StringPointer(oldUcBal.RatingSubject)
			}
			if oldUcBal.Category != "" {
				bf.Categories = utils.StringMapPointer(utils.ParseStringMap(oldUcBal.Category))
			}
			if oldUcBal.SharedGroup != "" {
				bf.SharedGroups = utils.StringMapPointer(utils.ParseStringMap(oldUcBal.SharedGroup))
			}
			if oldUcBal.TimingIDs != "" {
				bf.TimingIDs = utils.StringMapPointer(utils.ParseStringMap(oldUcBal.TimingIDs))
			}
			if oldUcBal.Disabled != false {
				bf.Disabled = utils.BoolPointer(oldUcBal.Disabled)
			}
			bf.Timings = oldUcBal.Timings
			cf := &engine.CounterFilter{
				Value:  oldUcBal.Value,
				Filter: bf,
			}
			newUc.Counters[index] = cf
		}
		ac.UnitCounters[oldUc.BalanceType] = append(ac.UnitCounters[oldUc.BalanceType], newUc)
	}
	//action triggers
	for index, oldAtr := range v1Acc.ActionTriggers {
		at := &engine.ActionTrigger{
			UniqueID:       oldAtr.Id,
			ThresholdType:  oldAtr.ThresholdType,
			ThresholdValue: oldAtr.ThresholdValue,
			Recurrent:      oldAtr.Recurrent,
			MinSleep:       oldAtr.MinSleep,
			Weight:         oldAtr.Weight,
			ActionsID:      oldAtr.ActionsId,
			MinQueuedItems: oldAtr.MinQueuedItems,
			Executed:       oldAtr.Executed,
		}
		bf := &engine.BalanceFilter{}
		if oldAtr.BalanceId != "" {
			bf.ID = utils.StringPointer(oldAtr.BalanceId)
		}
		if oldAtr.BalanceType != "" {
			bf.Type = utils.StringPointer(oldAtr.BalanceType)
		}
		if oldAtr.BalanceRatingSubject != "" {
			bf.RatingSubject = utils.StringPointer(oldAtr.BalanceRatingSubject)
		}
		if oldAtr.BalanceDirection != "" {
			bf.Directions = utils.StringMapPointer(utils.ParseStringMap(oldAtr.BalanceDirection))
		}
		if oldAtr.BalanceDestinationIds != "" {
			bf.DestinationIDs = utils.StringMapPointer(utils.ParseStringMap(oldAtr.BalanceDestinationIds))
		}
		if oldAtr.BalanceTimingTags != "" {
			bf.TimingIDs = utils.StringMapPointer(utils.ParseStringMap(oldAtr.BalanceTimingTags))
		}
		if oldAtr.BalanceCategory != "" {
			bf.Categories = utils.StringMapPointer(utils.ParseStringMap(oldAtr.BalanceCategory))
		}
		if oldAtr.BalanceSharedGroup != "" {
			bf.SharedGroups = utils.StringMapPointer(utils.ParseStringMap(oldAtr.BalanceSharedGroup))
		}
		if oldAtr.BalanceWeight != 0 {
			bf.Weight = utils.Float64Pointer(oldAtr.BalanceWeight)
		}
		if oldAtr.BalanceDisabled != false {
			bf.Disabled = utils.BoolPointer(oldAtr.BalanceDisabled)
		}
		if !oldAtr.BalanceExpirationDate.IsZero() {
			bf.ExpirationDate = utils.TimePointer(oldAtr.BalanceExpirationDate)
		}
		at.Balance = bf
		ac.ActionTriggers[index] = at
		if ac.ActionTriggers[index].ThresholdType == "*min_counter" ||
			ac.ActionTriggers[index].ThresholdType == "*max_counter" {
			ac.ActionTriggers[index].ThresholdType = strings.Replace(ac.ActionTriggers[index].ThresholdType, "_", "_event_", 1)
		}
	}
	return
}
