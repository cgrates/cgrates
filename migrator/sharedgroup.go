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

type v1SharedGroup struct {
	Id                string
	AccountParameters map[string]*engine.SharingParameters
	MemberIds         []string
}

func (m *Migrator) migrateCurrentSharedGroups() (err error) {
	var ids []string
	ids, err = m.dmIN.DataManager().DataDB().GetKeysForPrefix(utils.SHARED_GROUP_PREFIX)
	if err != nil {
		return err
	}
	for _, id := range ids {
		idg := strings.TrimPrefix(id, utils.SHARED_GROUP_PREFIX)
		sgs, err := m.dmIN.DataManager().GetSharedGroup(idg, true, utils.NonTransactional)
		if err != nil {
			return err
		}
		if sgs != nil {
			if m.dryRun != true {
				if err := m.dmOut.DataManager().SetSharedGroup(sgs, utils.NonTransactional); err != nil {
					return err
				}
			}
		}
	}
	return
}

func (m *Migrator) migrateV1SharedGroups() (err error) {
	var v1SG *v1SharedGroup
	for {
		v1SG, err = m.dmIN.getV1SharedGroup()
		if err != nil && err != utils.ErrNoMoreData {
			return err
		}
		if err == utils.ErrNoMoreData {
			break
		}
		if v1SG != nil {
			acnt := v1SG.AsSharedGroup()
			if m.dryRun != true {
				if err = m.dmOut.DataManager().SetSharedGroup(acnt, utils.NonTransactional); err != nil {
					return err
				}
				m.stats[utils.SharedGroups] += 1
			}
		}
	}
	// All done, update version wtih current one
	vrs := engine.Versions{utils.SharedGroups: engine.CurrentStorDBVersions()[utils.SharedGroups]}
	if err = m.dmOut.DataManager().DataDB().SetVersions(vrs, false); err != nil {
		return utils.NewCGRError(utils.Migrator,
			utils.ServerErrorCaps,
			err.Error(),
			fmt.Sprintf("error: <%s> when updating SharedGroups version into dataDB", err.Error()))
	}
	return
}

func (m *Migrator) migrateSharedGroups() (err error) {
	var vrs engine.Versions
	current := engine.CurrentDataDBVersions()
	vrs, err = m.dmOut.DataManager().DataDB().GetVersions("")
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
	switch vrs[utils.SharedGroups] {
	case current[utils.SharedGroups]:
		if m.sameDataDB {
			return
		}
		if err := m.migrateCurrentSharedGroups(); err != nil {
			return err
		}
		return

	case 1:
		if err := m.migrateV1SharedGroups(); err != nil {
			return err
		}
	}
	return
}

func (v1SG v1SharedGroup) AsSharedGroup() (sg *engine.SharedGroup) {
	sg = &engine.SharedGroup{
		Id:                v1SG.Id,
		AccountParameters: v1SG.AccountParameters,
		MemberIds:         make(utils.StringMap),
	}
	for _, accID := range v1SG.MemberIds {
		sg.MemberIds[accID] = true
	}
	return
}
