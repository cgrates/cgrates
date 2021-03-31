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
			[]string{"id"}, map[string]string{}, nil)
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
			[]string{"id"}, map[string]string{}, nil)
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
