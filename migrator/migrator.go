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
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func NewMigrator(storDB engine.StorDB, storDBType string) *Migrator {
	return &Migrator{storDB: storDB, storDBType: storDBType}
}

type Migrator struct {
	storDB     engine.StorDB
	storDBType string // Useful to convert back to real
}

func (m *Migrator) Migrate(taskID string) (err error) {
	switch taskID {
	default: // unsupported taskID
		err = utils.NewCGRError(utils.Migrator,
			utils.MandatoryIEMissingCaps,
			utils.UnsupportedMigrationTask,
			fmt.Sprintf("task <%s> is not a supported migration task", taskID))
	case utils.MetaCostDetails:
		err = m.migrateCostDetails()
	}
	return
}

func (m *Migrator) migrateCostDetails() (err error) {
	if m.storDB == nil {
		return utils.NewCGRError(utils.Migrator,
			utils.MandatoryIEMissingCaps,
			utils.NoStorDBConnection,
			"no connection to StorDB")
	}
	if !utils.IsSliceMember([]string{utils.MYSQL, utils.POSTGRES}, m.storDBType) {
		return // CostDetails are migrated only for MySQL and Postgres
	}
	vrs, err := m.storDB.GetVersions(utils.COST_DETAILS)
	if err != nil {
		return utils.NewCGRError(utils.Migrator,
			utils.ServerErrorCaps,
			err.Error(),
			fmt.Sprintf("error: <%s> when querying storDB for versions", err.Error()))
	} else if len(vrs) == 0 {
		return utils.NewCGRError(utils.Migrator,
			utils.MandatoryIEMissingCaps,
			utils.UndefinedVersion,
			"version number is not defined for CostDetails model")
	}
	if vrs[utils.COST_DETAILS] != 1 {
		return
	}
	storSQL := m.storDB.(*engine.SQLStorage)
	rows, err := storSQL.Db.Query("SELECT id, tor, direction, tenant, category, account, subject, destination, cost, cost_details FROM cdrs WHERE run_id!= '*raw' and cost_details NOT NULL")
	if err != nil {
		return utils.NewCGRError(utils.Migrator,
			utils.ServerErrorCaps,
			err.Error(),
			fmt.Sprintf("error: <%s> when querying storDB for cdrs", err.Error()))
	}
	defer rows.Close()
	for cnt := 0; rows.Next(); cnt++ {
		var id int64
		var ccDirection, ccCategory, ccTenant, ccSubject, ccAccount, ccDestination, ccTor sql.NullString
		var ccCost sql.NullFloat64
		var tts []byte
		if err := rows.Scan(&id, &ccTor, &ccDirection, &ccTenant, &ccCategory, &ccAccount, &ccSubject, &ccDestination, &ccCost, &tts); err != nil {
			return utils.NewCGRError(utils.Migrator,
				utils.ServerErrorCaps,
				err.Error(),
				fmt.Sprintf("error: <%s> when scanning at count: <%d>", err.Error(), cnt))
		}
		var v1tmsps v1TimeSpans
		if err := json.Unmarshal(tts, &v1tmsps); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<Migrator> Unmarshalling timespans at CDR with id: <%d>, error: <%s>", id, err.Error()))
			continue
		}
		v1CC := &v1CallCost{Direction: ccDirection.String, Category: ccCategory.String, Tenant: ccTenant.String,
			Subject: ccSubject.String, Account: ccAccount.String, Destination: ccDestination.String, TOR: ccTor.String,
			Cost: ccCost.Float64, Timespans: v1tmsps}
		cc, err := v1CC.AsCallCost()
		if err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<Migrator> Error: <%s> when converting into CallCost CDR with id: <%d>", err.Error(), id))
			continue
		}
		if _, err := storSQL.Db.Exec(fmt.Sprintf("UPDATE cdrs SET cost_details='%s' WHERE id=%d", cc.AsJSON(), id)); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<Migrator> Error: <%s> updating CDR with id <%d> into StorDB", err.Error(), id))
			continue
		}
	}
	// All done, update version wtih current one
	vrs = engine.Versions{utils.COST_DETAILS: engine.CurrentStorDBVersions()[utils.COST_DETAILS]}
	if err := m.storDB.SetVersions(vrs); err != nil {
		return utils.NewCGRError(utils.Migrator,
			utils.ServerErrorCaps,
			err.Error(),
			fmt.Sprintf("error: <%s> when updating CostDetails version into StorDB", err.Error()))
	}
	return
}
