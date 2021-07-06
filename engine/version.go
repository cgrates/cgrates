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
		utils.Accounts:       "cgr-migrator -exec=*accounts",
		utils.Attributes:     "cgr-migrator -exec=*attributes",
		utils.Actions:        "cgr-migrator -exec=*actions",
		utils.ActionTriggers: "cgr-migrator -exec=*action_triggers",
		utils.ActionPlans:    "cgr-migrator -exec=*action_plans",
		utils.SharedGroups:   "cgr-migrator -exec=*shared_groups",
		utils.Thresholds:     "cgr-migrator -exec=*thresholds",
		// utils.LoadIDsVrs:     "cgr-migrator -exec=*load_ids",
		utils.RQF:         "cgr-migrator -exec=*filters",
		utils.Routes:      "cgr-migrator -exec=*routes",
		utils.Dispatchers: "cgr-migrator -exec=*dispatchers",
		utils.Chargers:    "cgr-migrator -exec=*chargers",
		utils.StatS:       "cgr-migrator -exec=*stats",
	}
	storDBVers = map[string]string{
		utils.CostDetails:   "cgr-migrator -exec=*cost_details",
		utils.SessionSCosts: "cgr-migrator -exec=*sessions_costs",
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

// CheckVersions returns error if the db needs migration
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
			return fmt.Errorf("No versions defined: please backup cgrates data and run : <cgr-migrator -exec=*set_versions>")
		}
		// no data, safe to write version
		return OverwriteDBVersions(storage)
	} else if err != nil {
		return err
	}
	// comparing versions
	message := dbVersion.Compare(x, storType, isDataDB)
	if message != "" {
		return fmt.Errorf("Migration needed: please backup cgr data and run : <%s>", message)
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

// SetDBVersions sets the version without overwriting them
func SetDBVersions(storage Storage) (err error) {
	return setDBVersions(storage, false)
}

// OverwriteDBVersions sets the version overwriting them
func OverwriteDBVersions(storage Storage) (err error) {
	return setDBVersions(storage, true)
}

// Compare returns the migration message if the versions are not the latest
func (vers Versions) Compare(curent Versions, storType string, isDataDB bool) string {
	var message map[string]string
	switch storType {
	case utils.Mongo:
		if isDataDB {
			message = dataDBVers
		} else {
			message = storDBVers
		}
	case utils.INTERNAL:
		message = allVers
	case utils.Postgres, utils.MySQL:
		message = storDBVers
	case utils.Redis:
		message = dataDBVers
	}
	for subsis, reason := range message {
		if vers[subsis] != curent[subsis] {
			return reason
		}
	}
	return ""
}

// CurrentDataDBVersions returns the needed DataDB versions
func CurrentDataDBVersions() Versions {
	return Versions{
		utils.StatS:               4,
		utils.Accounts:            3,
		utils.Actions:             2,
		utils.ActionTriggers:      2,
		utils.ActionPlans:         3,
		utils.SharedGroups:        2,
		utils.Thresholds:          4,
		utils.Routes:              2,
		utils.Attributes:          6,
		utils.Timing:              1,
		utils.RQF:                 5,
		utils.Resource:            1,
		utils.Subscribers:         1,
		utils.Destinations:        1,
		utils.ReverseDestinations: 1,
		utils.RatingPlan:          1,
		utils.RatingProfile:       1,
		utils.Chargers:            2,
		utils.Dispatchers:         2,
		utils.LoadIDsVrs:          1,
	}
}

// CurrentStorDBVersions returns the needed StorDB versions
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
		utils.TpThresholds:       1,
		utils.TpRoutes:           1,
		utils.TpStats:            1,
		utils.TpSharedGroups:     1,
		utils.TpRatingProfiles:   1,
		utils.TpResources:        1,
		utils.TpRates:            1,
		utils.TpTiming:           1,
		utils.TpResource:         1,
		utils.TpDestinations:     1,
		utils.TpRatingPlan:       1,
		utils.TpRatingProfile:    1,
		utils.TpChargers:         1,
		utils.TpDispatchers:      1,
	}
}

// CurrentAllDBVersions returns the both DataDB and StorDB versions
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

// CurrentDBVersions returns versions based on dbType
func CurrentDBVersions(storType string, isDataDB bool) Versions {
	switch storType {
	case utils.Mongo:
		if isDataDB {
			return CurrentDataDBVersions()
		}
		return CurrentStorDBVersions()
	case utils.INTERNAL:
		return CurrentAllDBVersions()
	case utils.Postgres, utils.MySQL:
		return CurrentStorDBVersions()
	case utils.Redis:
		return CurrentDataDBVersions()
	}
	return nil
}
