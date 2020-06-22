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

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
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

func (m *Migrator) migrateCurrentActions() (err error) {
	var ids []string
	ids, err = m.dmIN.DataManager().DataDB().GetKeysForPrefix(utils.ACTION_PREFIX)
	if err != nil {
		return err
	}
	for _, id := range ids {
		idg := strings.TrimPrefix(id, utils.ACTION_PREFIX)
		acts, err := m.dmIN.DataManager().GetActions(idg, true, utils.NonTransactional)
		if err != nil {
			return err
		}
		if acts == nil || m.dryRun {
			continue
		}
		if err := m.dmOut.DataManager().SetActions(idg, acts, utils.NonTransactional); err != nil {
			return err
		}
		m.stats[utils.Actions] += 1
	}
	return
}

func (m *Migrator) removeV1Actions() (err error) {
	var v1 *v1Actions
	for {
		if v1, err = m.dmIN.getV1Actions(); err != nil && err != utils.ErrNoMoreData {
			return err
		}
		if v1 == nil {
			return nil
		}
		if err = m.dmIN.remV1Actions(*v1); err != nil {
			return err
		}
	}
}

func (m *Migrator) migrateV1Actions() (acts engine.Actions, err error) {
	var v1ACs *v1Actions

	if v1ACs, err = m.dmIN.getV1Actions(); err != nil {
		return nil, err
	}
	if *v1ACs == nil {
		return
	}
	for _, v1ac := range *v1ACs {
		act := v1ac.AsAction()
		acts = append(acts, act)

	}

	return
}

func (m *Migrator) migrateActions() (err error) {
	var vrs engine.Versions
	current := engine.CurrentDataDBVersions()
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
			"version number is not defined for ActionTriggers model")
	}
	migrated := true
	var acts engine.Actions
	for {
		version := vrs[utils.Actions]
		for {
			switch version {
			case current[utils.Actions]:
				migrated = false
				if m.sameDataDB {
					break
				}
				if err = m.migrateCurrentActions(); err != nil {
					return err
				}
			case 1:
				if acts, err = m.migrateV1Actions(); err != nil && err != utils.ErrNoMoreData {
					return err
				}
				version = 2
			}
			if version == current[utils.Actions] || err == utils.ErrNoMoreData {
				break
			}
		}
		if err == utils.ErrNoMoreData || !migrated {
			break
		}
		if !m.dryRun && migrated {
			if err := m.dmOut.DataManager().SetActions(acts[0].Id, acts, utils.NonTransactional); err != nil {
				return err
			}
		}
		m.stats[utils.Actions] += 1
	}

	if m.dryRun || !migrated {
		return nil
	}
	// remove old actions

	// All done, update version wtih current one
	vrs = engine.Versions{utils.Actions: engine.CurrentStorDBVersions()[utils.Actions]}
	if err = m.dmOut.DataManager().DataDB().SetVersions(vrs, false); err != nil {
		return utils.NewCGRError(utils.Migrator,
			utils.ServerErrorCaps,
			err.Error(),
			fmt.Sprintf("error: <%s> when updating Actions version into dataDB", err.Error()))
	}

	return m.ensureIndexesDataDB(engine.ColAct)
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
