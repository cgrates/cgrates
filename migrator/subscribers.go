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

// func (m *Migrator) migrateCurrentSubscribers() (err error) {
// 	subs, err := m.dmIN.DataManager().GetSubscribers()
// 	if err != nil {
// 		return err
// 	}
// 	for id, sub := range subs {
// 		if sub != nil {
// 			if m.dryRun != true {
// 				if err := m.dmOut.DataManager().SetSubscriber(id, sub); err != nil {
// 					return err
// 				}
// 				m.stats[utils.Subscribers] += 1
// 			}
// 		}
// 	}
// 	return
// }

func (m *Migrator) migrateSubscribers() (err error) {
	var vrs engine.Versions
	current := engine.CurrentDataDBVersions()
	if vrs, err = m.getVersions(utils.Subscribers); err != nil {
		return
	}
	switch vrs[utils.Subscribers] {
	case current[utils.Subscribers]:
		if m.sameDataDB {
			return
		}
		return utils.ErrNotImplemented
		// return  m.migrateCurrentSubscribers()
	}
	return
}
