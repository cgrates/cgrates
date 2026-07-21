// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package migrator

import (
	"github.com/cgrates/cgrates/utils"
)

func (m *Migrator) migrateLoadIDs() error {
	vrs, err := m.getVersions(utils.LoadIDsVrs)
	if err != nil {
		return err
	}
	if vrs[utils.LoadIDs] != 1 {
		dataDB, _, err := m.dmTo.DBConns().GetConn(utils.MetaLoadIDs)
		if err != nil {
			return err
		}
		if err = dataDB.RemoveLoadIDsDrv(); err != nil {
			return err
		}
		if err = m.setVersions(utils.LoadIDsVrs); err != nil {
			return err
		}
	}
	return nil
}
