// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package migrator

import (
	"errors"
	"fmt"
	"strings"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func (m *Migrator) migrateCurrentCharger() (err error) {
	var ids []string
	ids, err = m.dmIN.DataManager().DataDB().GetKeysForPrefix(utils.ChargerProfilePrefix, utils.EmptyString)
	if err != nil {
		return err
	}
	for _, id := range ids {
		tntID := strings.SplitN(strings.TrimPrefix(id, utils.ChargerProfilePrefix), utils.InInFieldSep, 2)
		if len(tntID) < 2 {
			return fmt.Errorf("Invalid key <%s> when migrating chargers", id)
		}
		cpp, err := m.dmIN.DataManager().GetChargerProfile(tntID[0], tntID[1], false, false, utils.NonTransactional)
		if err != nil {
			return err
		}
		if cpp == nil || m.dryRun {
			continue
		}
		if err := m.dmOut.DataManager().SetChargerProfile(cpp, true); err != nil {
			return err
		}
		if err := m.dmIN.DataManager().RemoveChargerProfile(tntID[0],
			tntID[1], false); err != nil {
			return err
		}
		m.stats[utils.Chargers]++
	}
	return
}

func (m *Migrator) migrateChargers() (err error) {
	var vrs engine.Versions
	current := engine.CurrentDataDBVersions()
	if vrs, err = m.getVersions(utils.Chargers); err != nil {
		return
	}
	migrated := true
	var v2 *engine.ChargerProfile
	for {
		version := vrs[utils.Chargers]
		for {
			switch version {
			default:
				return fmt.Errorf("Unsupported version %v", version)
			case current[utils.Chargers]:
				migrated = false
				if m.sameDataDB {
					break
				}
				if err = m.migrateCurrentCharger(); err != nil {
					return
				}
			case 1:
				if v2, err = m.migrateV1ToV2Chargers(); err != nil && err != utils.ErrNoMoreData {
					return
				} else if err == utils.ErrNoMoreData {
					break
				}
				version = 2
			}
			if version == current[utils.Chargers] || err == utils.ErrNoMoreData {
				break
			}
		}
		if err == utils.ErrNoMoreData || !migrated {
			break
		}

		if !m.dryRun {
			//set action plan
			if err = m.dmOut.DataManager().SetChargerProfile(v2, true); err != nil {
				return
			}
		}
		m.stats[utils.Chargers]++
	}
	// All done, update version wtih current one
	if err = m.setVersions(utils.Chargers); err != nil {
		return
	}
	return m.ensureIndexesDataDB(engine.ColCpp)
}

func (m *Migrator) migrateV1ToV2Chargers() (v4Cpp *engine.ChargerProfile, err error) {
	v4Cpp, err = m.dmIN.getV1ChargerProfile()
	if err != nil {
		return nil, err
	} else if v4Cpp == nil {
		return nil, errors.New("Charger NIL")
	}
	if v4Cpp.FilterIDs, err = migrateInlineFilterV4(v4Cpp.FilterIDs); err != nil {
		return nil, err
	}
	return
}
