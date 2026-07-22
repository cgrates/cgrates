// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package migrator

import (
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func (m *Migrator) migrateLoadIDs() (err error) {
	var vrs engine.Versions
	if vrs, err = m.getVersions(utils.LoadIDsVrs); err != nil {
		return
	}
	if vrs[utils.LoadIDs] != 1 {
		if err = m.dmOut.DataManager().DataDB().RemoveLoadIDsDrv(); err != nil {
			return
		}
		if err = m.setVersions(utils.LoadIDsVrs); err != nil {
			return err
		}
	}

	return
}
