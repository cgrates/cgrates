// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package migrator

import (
	"fmt"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func (m *Migrator) migrateLoadIDs() (err error) {
	var vrs engine.Versions
	if vrs, err = m.dmIN.DataManager().DataDB().GetVersions(""); err != nil {
		return utils.NewCGRError(utils.Migrator,
			utils.ServerErrorCaps,
			err.Error(),
			fmt.Sprintf("error: <%s> when querying oldDataDB for versions", err.Error()))
	}
	if vrs[utils.LoadIDs] != 1 {
		if err = m.dmOut.DataManager().DataDB().RemoveLoadIDsDrv(); err != nil {
			return
		}
		// All done, update version with current one
		vrs := engine.Versions{utils.LoadIDsVrs: engine.CurrentDataDBVersions()[utils.LoadIDsVrs]}
		if err = m.dmOut.DataManager().DataDB().SetVersions(vrs, false); err != nil {
			return utils.NewCGRError(utils.Migrator,
				utils.ServerErrorCaps,
				err.Error(),
				fmt.Sprintf("error: <%s> when updating LoadIDs version into dataDB", err))
		}
	}

	return
}
