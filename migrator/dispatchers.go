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

func (m *Migrator) migrateCurrentDispatcher() (err error) {
	var ids []string
	ids, err = m.dmIN.DataManager().DataDB().GetKeysForPrefix(utils.DispatcherProfilePrefix, utils.EmptyString)
	if err != nil {
		return
	}
	for _, id := range ids {
		tntID := strings.SplitN(strings.TrimPrefix(id, utils.DispatcherProfilePrefix), utils.InInFieldSep, 2)
		if len(tntID) < 2 {
			return fmt.Errorf("Invalid key <%s> when migrating dispatcher profiles", id)
		}
		dpp, err := m.dmIN.DataManager().GetDispatcherProfile(tntID[0], tntID[1], false, false, utils.NonTransactional)
		if err != nil {
			return err
		}
		if dpp == nil || m.dryRun {
			continue
		}
		if err := m.dmOut.DataManager().SetDispatcherProfile(dpp, true); err != nil {
			return err
		}
		if err := m.dmIN.DataManager().RemoveDispatcherProfile(tntID[0],
			tntID[1], false); err != nil {
			return err
		}
		m.stats[utils.Dispatchers]++
	}
	return
}

func (m *Migrator) migrateCurrentDispatcherHost() (err error) {
	var ids []string
	ids, err = m.dmIN.DataManager().DataDB().GetKeysForPrefix(utils.DispatcherHostPrefix, utils.EmptyString)
	if err != nil {
		return err
	}
	for _, id := range ids {
		tntID := strings.SplitN(strings.TrimPrefix(id, utils.DispatcherHostPrefix), utils.InInFieldSep, 2)
		if len(tntID) < 2 {
			return fmt.Errorf("Invalid key <%s> when migrating dispatcher hosts", id)
		}
		dpp, err := m.dmIN.DataManager().GetDispatcherHost(tntID[0], tntID[1], false, false, utils.NonTransactional)
		if err != nil {
			return err
		}
		if dpp == nil || m.dryRun {
			continue
		}
		if err := m.dmOut.DataManager().SetDispatcherHost(dpp); err != nil {
			return err
		}
		if err := m.dmIN.DataManager().RemoveDispatcherHost(tntID[0],
			tntID[1]); err != nil {
			return err
		}
	}
	return
}

func (m *Migrator) migrateDispatchers() (err error) {
	var vrs engine.Versions
	current := engine.CurrentDataDBVersions()
	if vrs, err = m.getVersions(utils.Dispatchers); err != nil {
		return
	}
	migrated := true
	var v2 *engine.DispatcherProfile
	for {
		version := vrs[utils.Dispatchers]
		for {
			switch version {
			default:
				return fmt.Errorf("Unsupported version %v", version)
			case current[utils.Dispatchers]:
				migrated = false
				if m.sameDataDB {
					break
				}
				if err = m.migrateCurrentDispatcher(); err != nil {
					return
				}
				if err = m.migrateCurrentDispatcherHost(); err != nil {
					return
				}
			case 1:
				if v2, err = m.migrateV1ToV2Dispatchers(); err != nil && err != utils.ErrNoMoreData {
					return
				} else if err == utils.ErrNoMoreData {
					break
				}
				version = 2
			}
			if version == current[utils.Dispatchers] || err == utils.ErrNoMoreData {
				break
			}
		}
		if err == utils.ErrNoMoreData || !migrated {
			break
		}

		if !m.dryRun {
			//set action plan
			if err = m.dmOut.DataManager().SetDispatcherProfile(v2, true); err != nil {
				return
			}
		}
		m.stats[utils.Dispatchers]++
	}
	// All done, update version wtih current one
	if err = m.setVersions(utils.Dispatchers); err != nil {
		return
	}
	return m.ensureIndexesDataDB(engine.ColDpp, engine.ColDph)
}

func (m *Migrator) migrateV1ToV2Dispatchers() (v4Cpp *engine.DispatcherProfile, err error) {
	v4Cpp, err = m.dmIN.getV1DispatcherProfile()
	if err != nil {
		return nil, err
	} else if v4Cpp == nil {
		return nil, errors.New("Dispatcher NIL")
	}
	if v4Cpp.FilterIDs, err = migrateInlineFilterV4(v4Cpp.FilterIDs); err != nil {
		return nil, err
	}
	return
}
