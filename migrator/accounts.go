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
	"errors"
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
	ids, err = m.dmIN.DataManager().DataDB().GetKeysForPrefix(utils.AccountPrefix)
	if err != nil {
		return err
	}
	for _, id := range ids {
		idg := strings.TrimPrefix(id, utils.AccountPrefix)
		acc, err := m.dmIN.DataManager().GetAccount(idg)
		if err != nil {
			return err
		}
		if acc == nil || m.dryRun {
			continue
		}
		if err := m.dmOut.DataManager().SetAccount(acc); err != nil {
			return err
		}
		if err := m.dmIN.DataManager().RemoveAccount(idg); err != nil {
			return err
		}
		m.stats[utils.Accounts]++
	}
	return
}

func (m *Migrator) removeV1Accounts() (err error) {
	var v1Acnt *v1Account
	for {
		v1Acnt, err = m.dmIN.getv1Account()
		if err != nil && err != utils.ErrNoMoreData {
			return err
		}
		if err == utils.ErrNoMoreData {
			break
		}
		if err = m.dmIN.remV1Account(v1Acnt.Id); err != nil {
			return err
		}
	}
	return
}

func (m *Migrator) migrateV1Accounts() (v3Acnt *engine.Account, err error) {
	var v1Acnt *v1Account
	v1Acnt, err = m.dmIN.getv1Account()
	if err != nil {
		return nil, err
	} else if v1Acnt == nil {
		return nil, errors.New("Account is nil")
	}
	v3Acnt = v1Acnt.V1toV3Account()
	return
}

func (m *Migrator) removeV2Accounts() (err error) {
	var v2Acnt *v2Account
	for {
		v2Acnt, err = m.dmIN.getv2Account()
		if err != nil && err != utils.ErrNoMoreData {
			return err
		}
		if err == utils.ErrNoMoreData {
			break
		}
		if err = m.dmIN.remV2Account(v2Acnt.ID); err != nil {
			return
		}
	}
	return
}

func (m *Migrator) migrateV2Accounts() (v3Acnt *engine.Account, err error) {
	var v2Acnt *v2Account
	v2Acnt, err = m.dmIN.getv2Account()
	if err != nil {
		return nil, err
	} else if v2Acnt == nil {
		return nil, errors.New("Account is nil")
	}
	v3Acnt = v2Acnt.V2toV3Account()
	return
}

func (m *Migrator) migrateAccounts() (err error) {
	var vrs engine.Versions
	current := engine.CurrentDataDBVersions()
	if vrs, err = m.getVersions(utils.Accounts); err != nil {
		return
	}
	migrated := true
	migratedFrom := 0
	var v3Acnt *engine.Account
	for {
		version := vrs[utils.Accounts]
		migratedFrom = int(version)
		for {
			switch version {
			default:
				return fmt.Errorf("Unsupported version %v", version)
			case current[utils.Accounts]:
				migrated = false
				if m.sameDataDB {
					break
				}
				if err = m.migrateCurrentAccounts(); err != nil {
					return
				}
				version = 3
			case 1: //migrate v1 to v3
				if v3Acnt, err = m.migrateV1Accounts(); err != nil && err != utils.ErrNoMoreData {
					return err
				} else if err == utils.ErrNoMoreData {
					break
				}
				version = 3
			case 2: //migrate v2 to v3
				if v3Acnt, err = m.migrateV2Accounts(); err != nil && err != utils.ErrNoMoreData {
					return
				} else if err == utils.ErrNoMoreData {
					break
				}
				version = 3
			}
			if version == current[utils.Accounts] || err == utils.ErrNoMoreData {
				break
			}
		}
		if err == utils.ErrNoMoreData || !migrated {
			break
		}

		if !m.dryRun {
			if err = m.dmOut.DataManager().SetAccount(v3Acnt); err != nil {
				return
			}
		}
		m.stats[utils.Accounts]++

	}
	if m.dryRun || !migrated {
		return nil
	}
	// Remove old accounts from dbIn (only if dbIn != dbOut )
	if !m.sameDataDB {
		switch migratedFrom {
		case 1:
			if err = m.removeV1Accounts(); err != nil && err != utils.ErrNoMoreData {
				return
			}
		case 2:
			if err = m.removeV2Accounts(); err != nil && err != utils.ErrNoMoreData {
				return
			}
		}
	}

	// All done, update version wtih current one
	if err = m.setVersions(utils.Accounts); err != nil {
		return err
	}
	return m.ensureIndexesDataDB(engine.ColAcc)
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
	ID             string
	BalanceMap     map[string]engine.Balances
	UnitCounters   engine.UnitCounters
	ActionTriggers engine.ActionTriggers
	AllowNegative  bool
	Disabled       bool
}

func (b *v1Balance) IsDefault() bool {
	return (b.DestinationIds == "" || b.DestinationIds == utils.MetaAny) &&
		b.RatingSubject == "" &&
		b.Category == "" &&
		b.ExpirationDate.IsZero() &&
		b.SharedGroup == "" &&
		b.Weight == 0 &&
		!b.Disabled
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
	idElements := strings.Split(ac.ID, utils.ConcatenatedKeySep)
	if len(idElements) != 3 {
		log.Printf("Malformed account ID %s", v1Acc.Id)
	}
	ac.ID = utils.ConcatenatedKey(idElements[1], idElements[2])
	// balances
	for oldBalKey, oldBalChain := range v1Acc.BalanceMap {
		keyElements := strings.Split(oldBalKey, utils.Meta)
		newBalKey := utils.Meta + keyElements[1]
		ac.BalanceMap[newBalKey] = make(engine.Balances, len(oldBalChain))
		for index, oldBal := range oldBalChain {
			balVal := oldBal.Value
			if newBalKey == utils.MetaVoice {
				balVal = utils.Round(balVal/float64(time.Second),
					config.CgrConfig().GeneralCfg().RoundingDecimals,
					utils.MetaRoundingMiddle)
			}
			// check default to set new id
			ac.BalanceMap[newBalKey][index] = &engine.Balance{
				Uuid:           oldBal.Uuid,
				ID:             oldBal.Id,
				Value:          balVal,
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
			if oldUcBal.Disabled {
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
			if balType == utils.MetaVoice {
				balVal = utils.Round(balVal*float64(time.Second),
					config.CgrConfig().GeneralCfg().RoundingDecimals,
					utils.MetaRoundingMiddle)
			}
			// check default to set new id
			ac.BalanceMap[balType][index] = &engine.Balance{
				Uuid:           oldBal.Uuid,
				ID:             oldBal.ID,
				Value:          balVal,
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
