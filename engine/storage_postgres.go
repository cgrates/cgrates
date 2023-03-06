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

package engine

import (
	"github.com/cgrates/cgrates/utils"
)

type PostgresStorage struct {
	SQLStorage
}

func (poS *PostgresStorage) SetVersions(vrs Versions, overwrite bool) (err error) {
	tx := poS.db.Begin()
	if overwrite {
		tx.Table(utils.TBLVersions).Delete(nil)
	}
	for key, val := range vrs {
		vrModel := &TBLVersion{Item: key, Version: val}
		if !overwrite {
			if err = tx.Model(&TBLVersion{}).Where(
				TBLVersion{Item: vrModel.Item}).Delete(TBLVersion{Version: val}).Error; err != nil {
				tx.Rollback()
				return
			}
		}
		if err = tx.Save(vrModel).Error; err != nil {
			tx.Rollback()
			return
		}
	}
	tx.Commit()
	return
}

func (poS *PostgresStorage) GetStorageType() string {
	return utils.MetaPostgres
}
