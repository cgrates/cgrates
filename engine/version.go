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
	"fmt"

	"github.com/cgrates/cgrates/utils"
)

var (
	dataDBVers = map[string]string{
		utils.Accounts:       "cgr-migrator -migrate=*accounts",
		utils.Attributes:     "cgr-migrator -migrate=*attributes",
		utils.Actions:        "cgr-migrator -migrate=*actions",
		utils.ActionTriggers: "cgr-migrator -migrate=*action_triggers",
		utils.ActionPlans:    "cgr-migrator -migrate=*action_plans",
		utils.SharedGroups:   "cgr-migrator -migrate=*shared_groups",
		utils.Thresholds:     "cgr-migrator -migrate=*thresholds",
	}
	storDBVers = map[string]string{
		utils.CostDetails:   "cgr-migrator -migrate=*cost_details",
		utils.SessionSCosts: "cgr-migrator -migrate=*sessions_costs",
	}
	allVers map[string]string // init will fill this with a merge of data+stor
)

func init() {
	allVers = make(map[string]string)
	for k, v := range dataDBVers {
		allVers[k] = v
	}
	for k, v := range storDBVers {
		allVers[k] = v
	}
}

// Versions will keep trac of various item versions
type Versions map[string]int64 // map[item]versionNr

func CheckVersions(storage Storage) error {
	// get current db version
	storType := storage.GetStorageType()
	isDataDB := isDataDB(storage)

	x := CurrentDBVersions(storType, isDataDB)
	dbVersion, err := storage.GetVersions("")
	if err == utils.ErrNotFound {
		empty, err := storage.IsDBEmpty()
		if err != nil {
			return err
		}
		if !empty {
			return fmt.Errorf("Migration needed: please backup cgrates data and run : <cgr-migrator>")
		}
		// no data, safe to write version
		if err := OverwriteDBVersions(storage); err != nil {
			return err
		}
	} else {
		// comparing versions
		message := dbVersion.Compare(x, storType, isDataDB)
		if message != "" {
			return fmt.Errorf("Migration needed: please backup cgr data and run : <%s>", message)
		}
	}
	return nil
}

// relevant only for mongoDB
func isDataDB(storage Storage) bool {
	conv, ok := storage.(*MongoStorage)
	if !ok {
		return false
	}
	return conv.IsDataDB()
}

func setDBVersions(storage Storage, overwrite bool) (err error) {
	x := CurrentDBVersions(storage.GetStorageType(), isDataDB(storage))
	// no data, write version
	if err = storage.SetVersions(x, overwrite); err != nil {
		utils.Logger.Warning(fmt.Sprintf("Could not write current version to db: %v", err))
		return err
	}
	return
}

func SetDBVersions(storage Storage) (err error) {
	return setDBVersions(storage, false)
}

func OverwriteDBVersions(storage Storage) (err error) {
	return setDBVersions(storage, true)
}

func (vers Versions) Compare(curent Versions, storType string, isDataDB bool) string {
	var message map[string]string
	switch storType {
	case utils.MONGO:
		if isDataDB {
			message = dataDBVers
		} else {
			message = storDBVers
		}
	case utils.MAPSTOR:
		message = allVers
	case utils.POSTGRES, utils.MYSQL:
		message = storDBVers
	case utils.REDIS:
		message = dataDBVers
	}
	for subsis, reason := range message {
		if vers[subsis] != curent[subsis] {
			return reason
		}
	}
	return ""
}

func CurrentDataDBVersions() Versions {
	return Versions{
		utils.StatS:               2,
		utils.Accounts:            3,
		utils.Actions:             2,
		utils.ActionTriggers:      2,
		utils.ActionPlans:         2,
		utils.SharedGroups:        2,
		utils.Thresholds:          3,
		utils.Suppliers:           1,
		utils.Attributes:          2,
		utils.Timing:              1,
		utils.RQF:                 1,
		utils.Resource:            1,
		utils.ReverseAlias:        1,
		utils.Alias:               1,
		utils.User:                1,
		utils.Subscribers:         1,
		utils.DerivedChargersV:    1,
		utils.CdrStats:            1,
		utils.Destinations:        1,
		utils.ReverseDestinations: 1,
		utils.LCR:                 1,
		utils.RatingPlan:          1,
		utils.RatingProfile:       1,
		utils.Chargers:            1,
	}
}

func CurrentStorDBVersions() Versions {
	return Versions{
		utils.CostDetails:        2,
		utils.SessionSCosts:      3,
		utils.CDRs:               2,
		utils.TpRatingPlans:      1,
		utils.TpFilters:          1,
		utils.TpDestinationRates: 1,
		utils.TpActionTriggers:   1,
		utils.TpAccountActionsV:  1,
		utils.TpActionPlans:      1,
		utils.TpActions:          1,
		utils.TpDerivedCharges:   1,
		utils.TpThresholds:       1,
		utils.TpSuppliers:        1,
		utils.TpStats:            1,
		utils.TpSharedGroups:     1,
		utils.TpRatingProfiles:   1,
		utils.TpResources:        1,
		utils.TpRates:            1,
		utils.TpTiming:           1,
		utils.TpResource:         1,
		utils.TpAliases:          1,
		utils.TpUsers:            1,
		utils.TpDerivedChargersV: 1,
		utils.TpCdrStats:         1,
		utils.TpDestinations:     1,
		utils.TpRatingPlan:       1,
		utils.TpRatingProfile:    1,
		utils.TpChargers:         1,
	}
}

func CurrentAllDBVersions() Versions {
	dataDbVersions := CurrentDataDBVersions()
	storDbVersions := CurrentStorDBVersions()
	allVersions := make(Versions)
	for k, v := range dataDbVersions {
		allVersions[k] = v
	}
	for k, v := range storDbVersions {
		allVersions[k] = v
	}
	return allVersions
}

func CurrentDBVersions(storType string, isDataDB bool) Versions {
	switch storType {
	case utils.MONGO:
		if isDataDB {
			return CurrentDataDBVersions()
		}
		return CurrentStorDBVersions()
	case utils.MAPSTOR:
		return CurrentAllDBVersions()
	case utils.POSTGRES, utils.MYSQL:
		return CurrentStorDBVersions()
	case utils.REDIS:
		return CurrentDataDBVersions()
	}
	return nil
}
