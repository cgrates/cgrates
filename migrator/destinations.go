// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package migrator

import (
	"fmt"
	"strings"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func (m *Migrator) migrateCurrentDestinations() (err error) {
	var ids []string
	ids, err = m.dmIN.DataManager().DataDB().GetKeysForPrefix(utils.DestinationPrefix, utils.EmptyString)
	if err != nil {
		return err
	}
	for _, id := range ids {
		idg := strings.TrimPrefix(id, utils.DestinationPrefix)
		dst, err := m.dmIN.DataManager().GetDestination(idg, false, true, utils.NonTransactional)
		if err != nil {
			return err
		}
		if dst == nil || m.dryRun {
			continue
		}
		if err := m.dmOut.DataManager().SetDestination(dst, utils.NonTransactional); err != nil {
			return err
		}
		m.stats[utils.Destinations]++
	}
	return
}

func (m *Migrator) migrateDestinations() (err error) {
	var vrs engine.Versions
	current := engine.CurrentDataDBVersions()
	if vrs, err = m.getVersions(utils.Destinations); err != nil {
		return
	}
	migrated := true
	for {
		version := vrs[utils.Destinations]
		for {
			switch version {
			default:
				return fmt.Errorf("Unsupported version %v", version)
			case current[utils.Destinations]:
				migrated = false
				if m.sameDataDB {
					break
				}
				if err = m.migrateCurrentDestinations(); err != nil {
					return
				}
			}
			if version == current[utils.Destinations] || err == utils.ErrNoMoreData {
				break
			}
		}
		if err == utils.ErrNoMoreData || !migrated {
			break
		}

		// if !m.dryRun  {
		// 		if err = m.dmIN.DataManager().SetDestination(v2, true); err != nil {
		// 	return
		// }
		// }
		m.stats[utils.Destinations]++
	}
	// All done, update version wtih current one
	if err = m.setVersions(utils.Destinations); err != nil {
		return
	}
	return
}

func (m *Migrator) migrateCurrentReverseDestinations() (err error) {
	var ids []string
	ids, err = m.dmIN.DataManager().DataDB().GetKeysForPrefix(utils.ReverseDestinationPrefix, utils.EmptyString)
	if err != nil {
		return err
	}
	for _, id := range ids {
		id := strings.TrimPrefix(id, utils.ReverseDestinationPrefix)
		rdst, err := m.dmIN.DataManager().GetReverseDestination(id, false, true, utils.NonTransactional)
		if err != nil {
			return err
		}
		if rdst == nil {
			continue
		}
		for _, rdid := range rdst {
			rdstn, err := m.dmIN.DataManager().GetDestination(rdid, false, true, utils.NonTransactional)
			if err != nil {
				return err
			}
			if rdstn == nil || m.dryRun {
				continue
			}
			if err := m.dmOut.DataManager().SetDestination(rdstn, utils.NonTransactional); err != nil {
				return err
			}
			if err := m.dmOut.DataManager().SetReverseDestination(rdstn.Id, rdstn.Prefixes, utils.NonTransactional); err != nil {
				return err
			}
			m.stats[utils.ReverseDestinations]++
		}
	}
	return
}

func (m *Migrator) migrateReverseDestinations() (err error) {
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
	switch vrs[utils.ReverseDestinations] {
	case current[utils.ReverseDestinations]:
		if m.sameDataDB {
			break
		}
		if err = m.migrateCurrentReverseDestinations(); err != nil {
			return err
		}
	}
	return m.ensureIndexesDataDB(engine.ColDst, engine.ColRds)
}
