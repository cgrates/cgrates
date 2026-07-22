// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
