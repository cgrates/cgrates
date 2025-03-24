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
		utils.AccountsStr: "cgr-migrator -exec=*accounts",
		utils.Attributes:  "cgr-migrator -exec=*attributes",
		utils.Actions:     "cgr-migrator -exec=*actions",
		utils.Thresholds:  "cgr-migrator -exec=*thresholds",
		utils.LoadIDsVrs:  "cgr-migrator -exec=*load_ids",
		utils.RQF:         "cgr-migrator -exec=*filters",
		utils.Routes:      "cgr-migrator -exec=*routes",
	}
	allVers map[string]string // init will fill this with a merge of data+stor
)

func init() {
	allVers = make(map[string]string)
	for k, v := range dataDBVers {
		allVers[k] = v
	}
}

// Versions will keep trac of various item versions
type Versions map[string]int64 // map[item]versionNr

// CheckVersions returns an error if the db needs migration.
func CheckVersions(storage Storage) error {

	// Retrieve the current DB versions.
	storType := storage.GetStorageType()
	isDataDB := isDataDB(storage)
	currentVersions := CurrentDBVersions(storType, isDataDB)

	dbVersions, err := storage.GetVersions("")
	if err == utils.ErrNotFound {
		empty, err := storage.IsDBEmpty()
		if err != nil {
			return err
		}
		if !empty {
			return fmt.Errorf("No versions defined: please backup cgrates data and run: <cgr-migrator -exec=*set_versions>")
		}
		// No data, safe to set the versions.
		return OverwriteDBVersions(storage)
	} else if err != nil {
		return err
	}
	// Compare db versions with current versions.
	message := dbVersions.Compare(currentVersions, storType, isDataDB)
	if message != "" {
		return fmt.Errorf("Migration needed: please backup cgr data and run: <%s>", message)
	}
	return nil
}

// relevant only for mongoDB
func isDataDB(storage Storage) bool {
	conv, ok := storage.(*MongoStorage)
	return ok && conv.IsDataDB()
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
	case utils.MetaMongo:
		message = dataDBVers
	case utils.MetaInternal:
		message = allVers
	case utils.MetaRedis:
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
		utils.Stats:          4,
		utils.AccountsStr:    3,
		utils.Actions:        2,
		utils.Thresholds:     4,
		utils.Routes:         2,
		utils.Attributes:     7,
		utils.RQF:            5,
		utils.Resource:       1,
		utils.Subscribers:    1,
		utils.Chargers:       2,
		utils.LoadIDsVrs:     1,
		utils.RateProfiles:   1,
		utils.ActionProfiles: 1,
	}
}

func CurrentStorDBVersions() Versions {
	return Versions{
		utils.CostDetails:      2,
		utils.SessionSCosts:    3,
		utils.CDRs:             2,
		utils.TpFilters:        1,
		utils.TpThresholds:     1,
		utils.TpRoutes:         1,
		utils.TpStats:          1,
		utils.TpResources:      1,
		utils.TpResource:       1,
		utils.TpChargers:       1,
		utils.TpRateProfiles:   1,
		utils.TpActionProfiles: 1,
	}
}

// CurrentAllDBVersions returns the both DataDB
func CurrentAllDBVersions() Versions {
	dataDBVersions := CurrentDataDBVersions()
	allVersions := make(Versions)
	for k, v := range dataDBVersions {
		allVersions[k] = v
	}
	return allVersions
}

// CurrentDBVersions returns versions based on dbType
func CurrentDBVersions(storType string, isDataDB bool) Versions {
	switch storType {
	case utils.MetaMongo:
		if isDataDB {
			return CurrentDataDBVersions()
		}
		return CurrentStorDBVersions()
	case utils.MetaInternal:
		return CurrentAllDBVersions()
	case utils.MetaRedis:
		return CurrentDataDBVersions()
	case utils.MetaPostgres, utils.MetaMySQL:
		return CurrentStorDBVersions()
	}
	return nil
}
