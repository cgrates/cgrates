// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package migrator

import (
	"fmt"

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
