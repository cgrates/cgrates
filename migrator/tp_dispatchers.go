// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package migrator

import (
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func (m *Migrator) migrateCurrentTPDispatchers() (err error) {
	tpids, err := m.storDBIn.StorDB().GetTpIds(utils.TBLTPDispatchers)
	if err != nil {
		return err
	}

	for _, tpid := range tpids {
		ids, err := m.storDBIn.StorDB().GetTpTableIds(tpid, utils.TBLTPDispatchers,
			utils.TPDistinctIds{"id"}, map[string]string{}, nil)
		if err != nil {
			return err
		}
		for _, id := range ids {
			dispatchers, err := m.storDBIn.StorDB().GetTPDispatcherProfiles(tpid, "", id)
			if err != nil {
				return err
			}
			if dispatchers == nil || m.dryRun {
				continue
			}
			if err := m.storDBOut.StorDB().SetTPDispatcherProfiles(dispatchers); err != nil {
				return err
			}
			for _, dispatcher := range dispatchers {
				if err := m.storDBIn.StorDB().RemTpData(utils.TBLTPDispatchers, dispatcher.TPid,
					map[string]string{"id": dispatcher.ID}); err != nil {
					return err
				}
			}
			m.stats[utils.TpDispatchers]++
		}
	}
	return
}

func (m *Migrator) migrateCurrentTPDispatcherHosts() (err error) {
	tpids, err := m.storDBIn.StorDB().GetTpIds(utils.TBLTPDispatcherHosts)
	if err != nil {
		return err
	}

	for _, tpid := range tpids {
		ids, err := m.storDBIn.StorDB().GetTpTableIds(tpid, utils.TBLTPDispatcherHosts,
			utils.TPDistinctIds{"id"}, map[string]string{}, nil)
		if err != nil {
			return err
		}
		for _, id := range ids {
			dispatchers, err := m.storDBIn.StorDB().GetTPDispatcherHosts(tpid, "", id)
			if err != nil {
				return err
			}
			if dispatchers == nil || m.dryRun {
				continue
			}
			if err := m.storDBOut.StorDB().SetTPDispatcherHosts(dispatchers); err != nil {
				return err
			}
			for _, dispatcher := range dispatchers {
				if err := m.storDBIn.StorDB().RemTpData(utils.TBLTPDispatcherHosts, dispatcher.TPid,
					map[string]string{"id": dispatcher.ID}); err != nil {
					return err
				}
			}
		}
	}
	return
}

func (m *Migrator) migrateTPDispatchers() (err error) {
	var vrs engine.Versions
	current := engine.CurrentStorDBVersions()
	if vrs, err = m.getVersions(utils.TpDispatchers); err != nil {
		return
	}
	switch vrs[utils.TpDispatchers] {
	case current[utils.TpDispatchers]:
		if m.sameStorDB {
			break
		}
		if err := m.migrateCurrentTPDispatchers(); err != nil {
			return err
		}
		if err := m.migrateCurrentTPDispatcherHosts(); err != nil {
			return err
		}
	}
	return m.ensureIndexesStorDB(utils.TBLTPDispatchers, utils.TBLTPDispatcherHosts)
}
