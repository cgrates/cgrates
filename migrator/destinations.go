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

func (m *Migrator) migrateCurrentDestinations() (err error) {
	var ids []string
	ids, err = m.dmIN.DataManager().DataDB().GetKeysForPrefix(utils.DESTINATION_PREFIX)
	if err != nil {
		return err
	}
	for _, id := range ids {
		idg := strings.TrimPrefix(id, utils.DESTINATION_PREFIX)
		dst, err := m.dmIN.DataManager().DataDB().GetDestination(idg, true, utils.NonTransactional)
		if err != nil {
			return err
		}
		if dst != nil {
			if m.dryRun != true {
				if err := m.dmOut.DataManager().DataDB().SetDestination(dst, utils.NonTransactional); err != nil {
					return err
				}
				m.stats[utils.Destinations] += 1
			}
		}
	}
	return
}

func (m *Migrator) migrateDestinations() (err error) {
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
	switch vrs[utils.Destinations] {
	case current[utils.Destinations]:
		if m.sameDataDB {
			return
		}
		if err := m.migrateCurrentDestinations(); err != nil {
			return err
		}
		return
	}
	return
}

func (m *Migrator) migrateCurrentReverseDestinations() (err error) {
	var ids []string
	ids, err = m.dmIN.DataManager().DataDB().GetKeysForPrefix(utils.REVERSE_DESTINATION_PREFIX)
	if err != nil {
		return err
	}
	for _, id := range ids {
		id := strings.TrimPrefix(id, utils.REVERSE_DESTINATION_PREFIX)
		rdst, err := m.dmIN.DataManager().DataDB().GetReverseDestination(id, true, utils.NonTransactional)
		if err != nil {
			return err
		}
		if rdst != nil {
			for _, rdid := range rdst {
				rdstn, err := m.dmIN.DataManager().DataDB().GetDestination(rdid, true, utils.NonTransactional)
				if err != nil {
					return err
				}
				if rdstn != nil {
					if m.dryRun != true {
						if err := m.dmOut.DataManager().DataDB().SetDestination(rdstn, utils.NonTransactional); err != nil {
							return err
						}
						if err := m.dmOut.DataManager().DataDB().SetReverseDestination(rdstn, utils.NonTransactional); err != nil {
							return err
						}
						m.stats[utils.ReverseDestinations] += 1
					}
				}
			}
		}
	}
	return
}

func (m *Migrator) migrateReverseDestinations() (err error) {
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
	switch vrs[utils.ReverseDestinations] {
	case current[utils.ReverseDestinations]:
		if m.sameDataDB {
			return
		}
		if err := m.migrateCurrentReverseDestinations(); err != nil {
			return err
		}
		return
	}
	return
}
