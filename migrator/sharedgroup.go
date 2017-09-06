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
)

type v1SharedGroup struct {
	Id                string
	AccountParameters map[string]*engine.SharingParameters
	MemberIds         []string
}

func (m *Migrator) migrateSharedGroups() (err error) {
	var v1SG *v1SharedGroup
	for {
		v1SG, err = m.oldDataDB.getV1SharedGroup()
		if err != nil && err != utils.ErrNoMoreData {
			return err
		}
		if err == utils.ErrNoMoreData {
			break
		}
		if v1SG != nil {
			acnt := v1SG.AsSharedGroup()
			if err = m.dataDB.SetSharedGroup(acnt, utils.NonTransactional); err != nil {
				return err
			}
		}
	}
	// All done, update version wtih current one
	vrs := engine.Versions{utils.SharedGroups: engine.CurrentStorDBVersions()[utils.SharedGroups]}
	if err = m.dataDB.SetVersions(vrs, false); err != nil {
		return utils.NewCGRError(utils.Migrator,
			utils.ServerErrorCaps,
			err.Error(),
			fmt.Sprintf("error: <%s> when updating SharedGroups version into dataDB", err.Error()))
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
