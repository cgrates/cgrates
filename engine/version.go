package engine

import (
	"fmt"

	"github.com/cgrates/cgrates/utils"
)

func init() {
	// get current db version
	dbVersion, err := accountingStorage.GetStructVersion()
	if err != nil {
		utils.Logger.Warning(fmt.Sprintf("Could not retrive current version from db: %v", err))
		return
	}
	// comparing versions
	if currentVersion.CompareAndMigrate(dbVersion) {
		// write the new values
		if err := accountingStorage.SetStructVersion(currentVersion); err != nil {
			utils.Logger.Warning(fmt.Sprintf("Could not write current version to db: %v", err))
		}
	}
}

var (
	currentVersion = &StructVersion{
		Destinations:    "1",
		RatingPlans:     "1",
		RatingProfiles:  "1",
		Lcrs:            "1",
		DerivedChargers: "1",
		Actions:         "1",
		ActionPlans:     "1",
		ActionTriggers:  "1",
		SharedGroups:    "1",
		Accounts:        "1",
		CdrStats:        "1",
		Users:           "1",
		Alias:           "1",
		PubSubs:         "1",
		LoadHistory:     "1",
		Cdrs:            "1",
		SMCosts:         "1",
	}
)

type StructVersion struct {
	//  rating
	Destinations    string
	RatingPlans     string
	RatingProfiles  string
	Lcrs            string
	DerivedChargers string
	Actions         string
	ActionPlans     string
	ActionTriggers  string
	SharedGroups    string
	// accounting
	Accounts    string
	CdrStats    string
	Users       string
	Alias       string
	PubSubs     string
	LoadHistory string
	// cdr
	Cdrs    string
	SMCosts string
}

func (sv *StructVersion) CompareAndMigrate(dbVer *StructVersion) bool {
	migrationPerformed := false
	if sv.Destinations != dbVer.Destinations {
		migrationPerformed = true

	}
	if sv.RatingPlans != dbVer.RatingPlans {
		migrationPerformed = true

	}
	if sv.RatingProfiles != dbVer.RatingPlans {
		migrationPerformed = true

	}
	if sv.Lcrs != dbVer.Lcrs {
		migrationPerformed = true

	}
	if sv.DerivedChargers != dbVer.DerivedChargers {
		migrationPerformed = true

	}
	if sv.Actions != dbVer.Actions {
		migrationPerformed = true

	}
	if sv.ActionPlans != dbVer.ActionPlans {
		migrationPerformed = true

	}
	if sv.ActionTriggers != dbVer.ActionTriggers {
		migrationPerformed = true

	}
	if sv.SharedGroups != dbVer.SharedGroups {
		migrationPerformed = true

	}
	if sv.Accounts != dbVer.Accounts {
		migrationPerformed = true
	}
	if sv.CdrStats != dbVer.CdrStats {
		migrationPerformed = true
	}
	if sv.Users != dbVer.Users {
		migrationPerformed = true
	}
	if sv.Alias != dbVer.Alias {
		migrationPerformed = true
	}
	if sv.PubSubs != dbVer.PubSubs {
		migrationPerformed = true
	}
	if sv.LoadHistory != dbVer.LoadHistory {
		migrationPerformed = true
	}
	if sv.Cdrs != dbVer.Cdrs {
		migrationPerformed = true
	}
	if sv.SMCosts != dbVer.SMCosts {
		migrationPerformed = true
	}
	return migrationPerformed
}
