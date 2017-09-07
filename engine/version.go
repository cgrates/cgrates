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
	"errors"
	"fmt"
	"github.com/cgrates/cgrates/utils"
)

func CheckVersions(storage Storage) error {
	// get current db version
	if storage == nil {
		storage = dataStorage
	}
	storType := storage.GetStorageType()
	x := CurrentDBVersions(storType)
	dbVersion, err := storage.GetVersions(utils.TBLVersions)
	if err != nil {
		// no data, write version
		if err := storage.SetVersions(x, false); err != nil {
			utils.Logger.Warning(fmt.Sprintf("Could not write current version to db: %v", err))
		}

	} else {
		// comparing versions
		message := dbVersion.Compare(x, storType)
		if len(message) > 2 {
			// write the new values
			msg := "Migration needed: please backup cgr data and run : <" + message + ">"
			return errors.New(msg)
		}
	}
	return nil
}

func (vers Versions) Compare(curent Versions, storType string) string {
	var x map[string]string
	m := map[string]string{
		utils.Accounts:       "cgr-migrator -migrate=*accounts",
		utils.Actions:        "cgr-migrator -migrate=*actions",
		utils.ActionTriggers: "cgr-migrator -migrate=*action_triggers",
		utils.ActionPlans:    "cgr-migrator -migrate=*action_plans",
		utils.SharedGroups:   "cgr-migrator -migrate=*shared_groups",
		utils.COST_DETAILS:   "cgr-migrator -migrate=*cost_details",
	}
	data := map[string]string{
		utils.Accounts:       "cgr-migrator -migrate=*accounts",
		utils.Actions:        "cgr-migrator -migrate=*actions",
		utils.ActionTriggers: "cgr-migrator -migrate=*action_triggers",
		utils.ActionPlans:    "cgr-migrator -migrate=*action_plans",
		utils.SharedGroups:   "cgr-migrator -migrate=*shared_groups",
	}
	stor := map[string]string{
		utils.COST_DETAILS: "cgr-migrator -migrate=*cost_details",
	}
	switch storType {
	case utils.MONGO:
		x = m
	case utils.POSTGRES:
		x = stor
	case utils.MYSQL:
		x = stor
	case utils.REDIS:
		x = data
	case utils.MAPSTOR:
		x = m
	}
	for y, val := range x {
		if vers[y] != curent[y] {
			return val
		}
	}
	return ""
}

func CurrentDBVersions(storType string) Versions {
	switch storType {
	case utils.MONGO:
		return Versions{utils.Accounts: 2, utils.Actions: 2, utils.ActionTriggers: 2, utils.ActionPlans: 2, utils.SharedGroups: 2, utils.COST_DETAILS: 2}
	case utils.POSTGRES:
		return Versions{utils.COST_DETAILS: 2}
	case utils.MYSQL:
		return Versions{utils.COST_DETAILS: 2}
	case utils.REDIS:
		return Versions{utils.Accounts: 2, utils.Actions: 2, utils.ActionTriggers: 2, utils.ActionPlans: 2, utils.SharedGroups: 2}
	case utils.MAPSTOR:
		return Versions{utils.Accounts: 2, utils.Actions: 2, utils.ActionTriggers: 2, utils.ActionPlans: 2, utils.SharedGroups: 2, utils.COST_DETAILS: 2}
	}
	return nil
}

func CurrentStorDBVersions() Versions {
	return Versions{utils.COST_DETAILS: 2}
}

func CurrentDataDBVersions() Versions {
	return Versions{utils.Accounts: 2, utils.Actions: 2, utils.ActionTriggers: 2, utils.ActionPlans: 2, utils.SharedGroups: 2}
}

// Versions will keep trac of various item versions
type Versions map[string]int64 // map[item]versionNr
