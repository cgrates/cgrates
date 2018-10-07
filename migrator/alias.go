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
	//"log"
	"strings"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func (m *Migrator) migrateCurrentAlias() (err error) {
	var ids []string
	ids, err = m.dmIN.DataManager().DataDB().GetKeysForPrefix(utils.ALIASES_PREFIX)
	if err != nil {
		return err
	}
	for _, id := range ids {
		idg := strings.TrimPrefix(id, utils.ALIASES_PREFIX)
		usr, err := m.dmIN.DataManager().DataDB().GetAlias(idg, true, utils.NonTransactional)
		if err != nil {
			return err
		}
		if usr != nil {
			if m.dryRun != true {
				if err := m.dmOut.DataManager().DataDB().SetAlias(usr, utils.NonTransactional); err != nil {
					return err
				}
				m.stats[utils.Alias] += 1
			}
		}
	}
	return
}

// func (m *Migrator) migrateCurrentReverseAlias() (err error) {
// 	var ids []string
// 	ids, err = m.dmIN.DataDB().GetKeysForPrefix(utils.REVERSE_ALIASES_PREFIX)
// 	if err != nil {
// 		return err
// 	}
// 	log.Print("ids: ", ids)
// 	for _, id := range ids {
// 		idg := strings.TrimPrefix(id, utils.REVERSE_ALIASES_PREFIX)
// 		usrs, err := m.dmIN.DataDB().GetReverseAlias(idg, true, utils.NonTransactional)
// 		if err != nil {
// 			return err
// 		}
// 		log.Print(id)
// 		log.Print(idg)
// 		log.Print("values: ", usrs)

// 		for _, usr := range usrs {
// 			log.Print(usr)

// 			alias, err := m.dmIN.DataDB().GetAlias(usr, true, utils.NonTransactional)
// 			if err != nil {
// 				log.Print("erorr")
// 				return err
// 			}
// 			if alias != nil {
// 				if m.dryRun != true {
// 					if err := m.dmOut.DataDB().SetReverseAlias(alias, utils.NonTransactional); err != nil {
// 						return err
// 					}
// 					m.stats[utils.ReverseAlias] += 1
// 				}
// 			}
// 		}
// 	}
// 	return
// }

func (m *Migrator) migrateAlias() (err error) {
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
	switch vrs[utils.Alias] {
	case current[utils.Alias]:
		if m.sameDataDB {
			return
		}
		if err := m.migrateCurrentAlias(); err != nil {
			return err
		}
		return
	}
	return
}

func (m *Migrator) migrateReverseAlias() (err error) {
	// var vrs engine.Versions
	// current := engine.CurrentDataDBVersions()
	// vrs, err = m.dmOut.DataDB().GetVersions("")
	// if err != nil {
	// 	return utils.NewCGRError(utils.Migrator,
	// 		utils.ServerErrorCaps,
	// 		err.Error(),
	// 		fmt.Sprintf("error: <%s> when querying oldDataDB for versions", err.Error()))
	// } else if len(vrs) == 0 {
	// 	return utils.NewCGRError(utils.Migrator,
	// 		utils.MandatoryIEMissingCaps,
	// 		utils.UndefinedVersion,
	// 		"version number is not defined for ActionTriggers model")
	// }
	// switch vrs[utils.ReverseAlias] {
	// case current[utils.ReverseAlias]:
	// 	if m.sameDataDB {
	// 		return
	// 	}
	// 	if err := m.migrateCurrentReverseAlias(); err != nil {
	// 		return err
	// 	}
	// 	return
	// }
	return
}
