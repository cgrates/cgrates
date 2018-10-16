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

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

const (
	v1AccountDBPrefix = "ubl_"
	v1AccountTBL      = "userbalances"
)

func (m *Migrator) migrateCurrentAccounts() (err error) {
	var ids []string
	ids, err = m.dmIN.DataManager().DataDB().GetKeysForPrefix(utils.ACCOUNT_PREFIX)
	if err != nil {
		return err
	}
	for _, id := range ids {
		idg := strings.TrimPrefix(id, utils.ACCOUNT_PREFIX)
		acc, err := m.dmIN.DataManager().DataDB().GetAccount(idg)
		if err != nil {
			return err
		}
		if acc != nil {
			if m.dryRun != true {
				if err := m.dmOut.DataManager().DataDB().SetAccount(acc); err != nil {
					return err
				}
				if err := m.dmIN.DataManager().DataDB().RemoveAccount(idg); err != nil {
					return err
				}
				m.stats[utils.Accounts] += 1
			}
		}
	}
	return
}

func (m *Migrator) migrateV1Accounts() (err error) {
	var v1Acnt *v1Account
	for {
		v1Acnt, err = m.dmIN.getv1Account()
		if err != nil && err != utils.ErrNoMoreData {
			return err
		}
		if err == utils.ErrNoMoreData {
			break
		}
		if v1Acnt != nil {
			acnt := v1Acnt.V1toV3Account()
			if m.dryRun != true {
				if err = m.dmOut.DataManager().DataDB().SetAccount(acnt); err != nil {
					return err
				}
				if err = m.dmIN.remV1Account(v1Acnt.Id); err != nil {
					return err
				}
				m.stats[utils.Accounts] += 1
			}
		}
	}
	if m.dryRun != true {
		// All done, update version wtih current one
		vrs := engine.Versions{utils.Accounts: engine.CurrentDataDBVersions()[utils.Accounts]}
		if err = m.dmOut.DataManager().DataDB().SetVersions(vrs, false); err != nil {
			return utils.NewCGRError(utils.Migrator,
				utils.ServerErrorCaps,
				err.Error(),
				fmt.Sprintf("error: <%s> when updating Accounts version into StorDB", err.Error()))
		}
	}
	return
}

func (m *Migrator) migrateV2Accounts() (err error) {
	var v2Acnt *v2Account
	for {
		v2Acnt, err = m.dmIN.getv2Account()
		if err != nil && err != utils.ErrNoMoreData {
			return err
		}
		if err == utils.ErrNoMoreData {
			break
		}
		if v2Acnt != nil {
			acnt := v2Acnt.V2toV3Account()
			if m.dryRun != true {
				if err = m.dmOut.DataManager().DataDB().SetAccount(acnt); err != nil {
					return err
				}
				if err = m.dmIN.remV2Account(v2Acnt.ID); err != nil {
					return err
				}
				m.stats[utils.Accounts] += 1
			}
		}
	}
	if m.dryRun != true {
		// All done, update version wtih current one
		vrs := engine.Versions{utils.Accounts: engine.CurrentDataDBVersions()[utils.Accounts]}
		if err = m.dmOut.DataManager().DataDB().SetVersions(vrs, false); err != nil {
			return utils.NewCGRError(utils.Migrator,
				utils.ServerErrorCaps,
				err.Error(),
				fmt.Sprintf("error: <%s> when updating Accounts version into StorDB", err.Error()))
		}
	}
	return
}

func (m *Migrator) migrateAccounts() (err error) {
	var vrs engine.Versions
	vrs, err = m.dmIN.DataManager().DataDB().GetVersions("")
	if err != nil {
		return utils.NewCGRError(utils.Migrator,
			utils.ServerErrorCaps,
			err.Error(),
			fmt.Sprintf("error: <%s> when querying oldDataDB for versions", err.Error()))
	} else if len(vrs) == 0 {
		return utils.NewCGRError(utils.Migrator,
			utils.MandatoryIEMissingCaps,
			utils.UndefinedVersion,
			"version number is not defined for Actions")
	}
	current := engine.CurrentDataDBVersions()
	switch vrs[utils.Accounts] {

	case 1:
		return m.migrateV1Accounts()
	case 2:
		if err := m.migrateV2Accounts(); err != nil {
			return err
		}
		fallthrough
	case current[utils.Accounts]:
		if m.sameDataDB {
			return
		}
		return m.migrateCurrentAccounts()
	}
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

type v1UnitsCounter struct {
	Direction   string
	BalanceType string
	Balances    v1BalanceChain // first balance is the general one (no destination)
}

type v2Account struct {
	ID                string
	BalanceMap        map[string]engine.Balances
	UnitCounters      engine.UnitCounters
	ActionTriggers    engine.ActionTriggers
	AllowNegative     bool
	Disabled          bool
	executingTriggers bool
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

func (v1Acc v1Account) V1toV3Account() (ac *engine.Account) {
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
		newBalDirection := idElements[0]
		ac.BalanceMap[newBalKey] = make(engine.Balances, len(oldBalChain))
		for index, oldBal := range oldBalChain {
			balVal := oldBal.Value
			if newBalKey == utils.VOICE {
				balVal = utils.Round(balVal/float64(time.Second),
					config.CgrConfig().GeneralCfg().RoundingDecimals,
					utils.ROUNDING_MIDDLE)
			}
			// check default to set new id
			ac.BalanceMap[newBalKey][index] = &engine.Balance{
				Uuid:           oldBal.Uuid,
				ID:             oldBal.Id,
				Value:          balVal,
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

func (v2Acc v2Account) V2toV3Account() (ac *engine.Account) {
	ac = &engine.Account{
		ID:             v2Acc.ID,
		BalanceMap:     make(map[string]engine.Balances, len(v2Acc.BalanceMap)),
		UnitCounters:   make(engine.UnitCounters, len(v2Acc.UnitCounters)),
		ActionTriggers: make(engine.ActionTriggers, len(v2Acc.ActionTriggers)),
		AllowNegative:  v2Acc.AllowNegative,
		Disabled:       v2Acc.Disabled,
	}
	// balances
	for balType, oldBalChain := range v2Acc.BalanceMap {
		ac.BalanceMap[balType] = make(engine.Balances, len(oldBalChain))
		for index, oldBal := range oldBalChain {
			balVal := oldBal.Value
			if balType == utils.VOICE {
				balVal = utils.Round(balVal*float64(time.Second),
					config.CgrConfig().GeneralCfg().RoundingDecimals,
					utils.ROUNDING_MIDDLE)
			}
			// check default to set new id
			ac.BalanceMap[balType][index] = &engine.Balance{
				Uuid:           oldBal.Uuid,
				ID:             oldBal.ID,
				Value:          balVal,
				Directions:     oldBal.Directions,
				ExpirationDate: oldBal.ExpirationDate,
				Weight:         oldBal.Weight,
				DestinationIDs: oldBal.DestinationIDs,
				RatingSubject:  oldBal.RatingSubject,
				Categories:     oldBal.Categories,
				SharedGroups:   oldBal.SharedGroups,
				Timings:        oldBal.Timings,
				TimingIDs:      oldBal.TimingIDs,
				Disabled:       oldBal.Disabled,
				Factor:         oldBal.Factor,
			}
		}
	}
	// unit counters
	ac.UnitCounters = v2Acc.UnitCounters
	//action triggers
	ac.ActionTriggers = v2Acc.ActionTriggers
	return
}
