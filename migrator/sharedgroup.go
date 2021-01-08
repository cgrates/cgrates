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
	ids, err = m.dmIN.DataManager().DataDB().GetKeysForPrefix(utils.SharedGroupPrefix)
	if err != nil {
		return err
	}
	for _, id := range ids {
		idg := strings.TrimPrefix(id, utils.SharedGroupPrefix)
		sgs, err := m.dmIN.DataManager().GetSharedGroup(idg, true, utils.NonTransactional)
		if err != nil {
			return err
		}
		if sgs == nil || m.dryRun {
			continue
		}
		if err := m.dmOut.DataManager().SetSharedGroup(sgs, utils.NonTransactional); err != nil {
			return err
		}
		m.stats[utils.SharedGroups]++
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
		if v1SG == nil || m.dryRun {
			continue
		}
		acnt := v1SG.AsSharedGroup()
		if err = m.dmOut.DataManager().SetSharedGroup(acnt, utils.NonTransactional); err != nil {
			return err
		}
		m.stats[utils.SharedGroups]++
	}
	// All done, update version wtih current one
	if err = m.setVersions(utils.SharedGroups); err != nil {
		return err
	}
	return
}

func (m *Migrator) migrateSharedGroups() (err error) {
	var vrs engine.Versions
	current := engine.CurrentDataDBVersions()
	if vrs, err = m.getVersions(utils.SharedGroups); err != nil {
		return
	}
	switch version := vrs[utils.SharedGroups]; version {
	default:
		return fmt.Errorf("Unsupported version %v", version)
	case current[utils.SharedGroups]:
		if m.sameDataDB {
			break
		}
		if err = m.migrateCurrentSharedGroups(); err != nil {
			return err
		}
	case 1:
		if err = m.migrateV1SharedGroups(); err != nil {
			return err
		}
	}
	return m.ensureIndexesDataDB(engine.ColShg)
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
