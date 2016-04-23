package engine

import (
	"fmt"

	"github.com/cgrates/cgrates/utils"
)

func init() {
	// get current db version
	dbRsv, err := ratingStorage.GetRatingStructuresVersion()
	if err != nil {
		utils.Logger.Warning(fmt.Sprintf("Could not retrive current version from db: %v", err))
		return
	}
	// comparing versions
	if currentRsv.CompareAndMigrate(dbRsv) {
		// write the new values
		if err := ratingStorage.SetRatingStructuresVersion(currentRsv); err != nil {
			utils.Logger.Warning(fmt.Sprintf("Could not write current version to db: %v", err))
		}
	}
	dbAsv, err := accountingStorage.GetAccountingStructuresVersion()
	if err != nil {
		utils.Logger.Warning(fmt.Sprintf("Could not retrive current version from db: %v", err))
		return
	}
	// comparing versions
	if currentAsv.CompareAndMigrate(dbAsv) {
		// write the new values
		if err := accountingStorage.SetAccountingStructuresVersion(currentAsv); err != nil {
			utils.Logger.Warning(fmt.Sprintf("Could not write current version to db: %v", err))
		}
	}
	dbCsv, err := cdrStorage.GetCdrStructuresVersion()
	if err != nil {
		utils.Logger.Warning(fmt.Sprintf("Could not retrive current version from db: %v", err))
		return
	}
	// comparing versions
	if currentCsv.CompareAndMigrate(dbCsv) {
		// write the new values
		if err := cdrStorage.SetCdrStructuresVersion(currentCsv); err != nil {
			utils.Logger.Warning(fmt.Sprintf("Could not write current version to db: %v", err))
		}
	}
}

var (
	currentRsv = &RatingStructuresVersion{
		Destinations:    "1",
		RatingPlans:     "1",
		RatingProfiles:  "1",
		Lcrs:            "1",
		DerivedChargers: "1",
		Actions:         "1",
		ActionPlans:     "1",
		ActionTriggers:  "1",
		SharedGroups:    "1",
	}

	currentAsv = &AccountingStructuresVersion{
		Accounts:    "1",
		CdrStats:    "1",
		Users:       "1",
		Alias:       "1",
		PubSubs:     "1",
		LoadHistory: "1",
	}

	currentCsv = &CdrStructuresVersion{
		Cdrs:    "1",
		SMCosts: "1",
	}
)

type RatingStructuresVersion struct {
	Destinations    string
	RatingPlans     string
	RatingProfiles  string
	Lcrs            string
	DerivedChargers string
	Actions         string
	ActionPlans     string
	ActionTriggers  string
	SharedGroups    string
}

func (rsv *RatingStructuresVersion) CompareAndMigrate(dbRsv *RatingStructuresVersion) bool {
	migrationPerformed := false
	if rsv.Destinations != dbRsv.Destinations {
		migrationPerformed = true

	}
	if rsv.RatingPlans != dbRsv.RatingPlans {
		migrationPerformed = true

	}
	if rsv.RatingProfiles != dbRsv.RatingPlans {
		migrationPerformed = true

	}
	if rsv.Lcrs != dbRsv.Lcrs {
		migrationPerformed = true

	}
	if rsv.DerivedChargers != dbRsv.DerivedChargers {
		migrationPerformed = true

	}
	if rsv.Actions != dbRsv.Actions {
		migrationPerformed = true

	}
	if rsv.ActionPlans != dbRsv.ActionPlans {
		migrationPerformed = true

	}
	if rsv.ActionTriggers != dbRsv.ActionTriggers {
		migrationPerformed = true

	}
	if rsv.SharedGroups != dbRsv.SharedGroups {
		migrationPerformed = true

	}
	return migrationPerformed
}

type AccountingStructuresVersion struct {
	Accounts    string
	CdrStats    string
	Users       string
	Alias       string
	PubSubs     string
	LoadHistory string
}

func (asv *AccountingStructuresVersion) CompareAndMigrate(dbAsv *AccountingStructuresVersion) bool {
	migrationPerformed := false
	if asv.Accounts != dbAsv.Accounts {
		migrationPerformed = true
	}
	if asv.CdrStats != dbAsv.CdrStats {
		migrationPerformed = true
	}
	if asv.Users != dbAsv.Users {
		migrationPerformed = true
	}
	if asv.Alias != dbAsv.Alias {
		migrationPerformed = true
	}
	if asv.PubSubs != dbAsv.PubSubs {
		migrationPerformed = true
	}
	if asv.LoadHistory != dbAsv.LoadHistory {
		migrationPerformed = true
	}
	return migrationPerformed
}

type CdrStructuresVersion struct {
	Cdrs    string
	SMCosts string
}

func (csv *CdrStructuresVersion) CompareAndMigrate(dbCsv *CdrStructuresVersion) bool {
	migrationPerformed := false
	if csv.Cdrs != dbCsv.Cdrs {
		migrationPerformed = true
	}
	if csv.SMCosts != dbCsv.SMCosts {
		migrationPerformed = true
	}
	return migrationPerformed
}
