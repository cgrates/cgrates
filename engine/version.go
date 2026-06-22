/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package engine

import (
	"fmt"
	"maps"

	"github.com/cgrates/cgrates/utils"
)

// Versions will keep trac of various item versions
type Versions map[string]int64 // map[item]versionNr

// CheckVersions returns an error if the db needs migration.
func CheckVersions(storage Storage) error {

	// Retrieve the current DB versions.
	storType := storage.GetStorageType()
	currentVersions := CurrentDBVersions(storType)

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
	if subsys := dbVersions.Compare(currentVersions); subsys != "" {
		return fmt.Errorf("datadb version mismatch for %s (have %d, want %d): back up your data, flush the datadb and reload",
			subsys, dbVersions[subsys], currentVersions[subsys])
	}
	return nil
}

func setDBVersions(storage Storage, overwrite bool) (err error) {
	x := CurrentDBVersions(storage.GetStorageType())
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

// Compare returns the name of the first subsystem whose stored version differs
// from the current one, or "" when every version matches.
func (vers Versions) Compare(current Versions) string {
	for subsys, curVer := range current {
		if vers[subsys] != curVer {
			return subsys
		}
	}
	return ""
}

// CurrentDataDBVersions returns the needed DataDB versions
func CurrentDataDBVersions() Versions {
	return Versions{
		utils.Stats:          1,
		utils.AccountsStr:    1,
		utils.Actions:        1,
		utils.Thresholds:     1,
		utils.Routes:         1,
		utils.Attributes:     1,
		utils.RQF:            1,
		utils.ResourceStr:    1,
		utils.Subscribers:    1,
		utils.Chargers:       1,
		utils.LoadIDsVrs:     1,
		utils.RateProfiles:   1,
		utils.ActionProfiles: 1,
	}
}

func CurrentStorDBVersions() Versions {
	return Versions{
		utils.CostDetails:      1,
		utils.SessionSCosts:    1,
		utils.CDRs:             1,
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
	maps.Copy(allVersions, dataDBVersions)
	return allVersions
}

// CurrentDBVersions returns versions based on dbType
func CurrentDBVersions(storType string) Versions {
	switch storType {
	case utils.MetaMongo, utils.MetaPostgres, utils.MetaMySQL, utils.MetaInternal:
		return CurrentAllDBVersions()
	case utils.MetaRedis:
		return CurrentDataDBVersions()
	}
	return nil
}
