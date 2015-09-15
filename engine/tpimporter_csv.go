/*
Real-time Charging System for Telecom & ISP environments
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
	"io/ioutil"
	"log"
	"path"

	"github.com/cgrates/cgrates/utils"
)

// Import tariff plan from csv into storDb
type TPCSVImporter struct {
	TPid     string     // Load data on this tpid
	StorDb   LoadWriter // StorDb connection handle
	DirPath  string     // Directory path to import from
	Sep      rune       // Separator in the csv file
	Verbose  bool       // If true will print a detailed information instead of silently discarding it
	ImportId string     // Use this to differentiate between imports (eg: when autogenerating fields like RatingProfileId
	csvr     LoadReader
}

// Maps csv file to handler which should process it. Defined like this since tests on 1.0.3 were failing on Travis.
// Change it to func(string) error as soon as Travis updates.
var fileHandlers = map[string]func(*TPCSVImporter, string) error{
	utils.TIMINGS_CSV:           (*TPCSVImporter).importTimings,
	utils.DESTINATIONS_CSV:      (*TPCSVImporter).importDestinations,
	utils.RATES_CSV:             (*TPCSVImporter).importRates,
	utils.DESTINATION_RATES_CSV: (*TPCSVImporter).importDestinationRates,
	utils.RATING_PLANS_CSV:      (*TPCSVImporter).importRatingPlans,
	utils.RATING_PROFILES_CSV:   (*TPCSVImporter).importRatingProfiles,
	utils.SHARED_GROUPS_CSV:     (*TPCSVImporter).importSharedGroups,
	utils.ACTIONS_CSV:           (*TPCSVImporter).importActions,
	utils.ACTION_PLANS_CSV:      (*TPCSVImporter).importActionTimings,
	utils.ACTION_TRIGGERS_CSV:   (*TPCSVImporter).importActionTriggers,
	utils.ACCOUNT_ACTIONS_CSV:   (*TPCSVImporter).importAccountActions,
	utils.DERIVED_CHARGERS_CSV:  (*TPCSVImporter).importDerivedChargers,
	utils.CDR_STATS_CSV:         (*TPCSVImporter).importCdrStats,
	utils.LCRS_CSV:              (*TPCSVImporter).importLcrs,
	utils.USERS_CSV:             (*TPCSVImporter).importUsers,
	utils.ALIASES_CSV:           (*TPCSVImporter).importAliases,
}

func (self *TPCSVImporter) Run() error {
	self.csvr = NewFileCSVStorage(self.Sep,
		path.Join(self.DirPath, utils.DESTINATIONS_CSV),
		path.Join(self.DirPath, utils.TIMINGS_CSV),
		path.Join(self.DirPath, utils.RATES_CSV),
		path.Join(self.DirPath, utils.DESTINATION_RATES_CSV),
		path.Join(self.DirPath, utils.RATING_PLANS_CSV),
		path.Join(self.DirPath, utils.RATING_PROFILES_CSV),
		path.Join(self.DirPath, utils.SHARED_GROUPS_CSV),
		path.Join(self.DirPath, utils.LCRS_CSV),
		path.Join(self.DirPath, utils.ACTIONS_CSV),
		path.Join(self.DirPath, utils.ACTION_PLANS_CSV),
		path.Join(self.DirPath, utils.ACTION_TRIGGERS_CSV),
		path.Join(self.DirPath, utils.ACCOUNT_ACTIONS_CSV),
		path.Join(self.DirPath, utils.DERIVED_CHARGERS_CSV),
		path.Join(self.DirPath, utils.CDR_STATS_CSV),
		path.Join(self.DirPath, utils.USERS_CSV),
		path.Join(self.DirPath, utils.ALIASES_CSV),
	)
	files, _ := ioutil.ReadDir(self.DirPath)
	for _, f := range files {
		fHandler, hasName := fileHandlers[f.Name()]
		if !hasName {
			continue
		}
		if err := fHandler(self, f.Name()); err != nil {
			Logger.Err(fmt.Sprintf("<TPCSVImporter> Importing file: %s, got error: %s", f.Name(), err.Error()))
		}
	}
	return nil
}

// Handler importing timings from file, saved row by row to storDb
func (self *TPCSVImporter) importTimings(fn string) error {
	if self.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
	tps, err := self.csvr.GetTpTimings(self.TPid, "")
	if err != nil {
		return err
	}
	for i := 0; i < len(tps); i++ {
		tps[i].Tpid = self.TPid
	}

	return self.StorDb.SetTpTimings(tps)
}

func (self *TPCSVImporter) importDestinations(fn string) error {
	if self.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
	tps, err := self.csvr.GetTpDestinations(self.TPid, "")
	if err != nil {
		return err
	}
	for i := 0; i < len(tps); i++ {
		tps[i].Tpid = self.TPid
	}

	return self.StorDb.SetTpDestinations(tps)
}

func (self *TPCSVImporter) importRates(fn string) error {
	if self.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
	tps, err := self.csvr.GetTpRates(self.TPid, "")
	if err != nil {
		return err
	}
	for i := 0; i < len(tps); i++ {
		tps[i].Tpid = self.TPid
	}

	return self.StorDb.SetTpRates(tps)
}

func (self *TPCSVImporter) importDestinationRates(fn string) error {
	if self.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
	tps, err := self.csvr.GetTpDestinationRates(self.TPid, "", nil)
	if err != nil {
		return err
	}

	for i := 0; i < len(tps); i++ {
		tps[i].Tpid = self.TPid
	}

	return self.StorDb.SetTpDestinationRates(tps)
}

func (self *TPCSVImporter) importRatingPlans(fn string) error {
	if self.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
	tps, err := self.csvr.GetTpRatingPlans(self.TPid, "", nil)
	if err != nil {
		return err
	}
	for i := 0; i < len(tps); i++ {
		tps[i].Tpid = self.TPid
	}

	return self.StorDb.SetTpRatingPlans(tps)
}

func (self *TPCSVImporter) importRatingProfiles(fn string) error {
	if self.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
	tps, err := self.csvr.GetTpRatingProfiles(&TpRatingProfile{Tpid: self.TPid})
	if err != nil {
		return err
	}
	loadId := utils.CSV_LOAD //Autogenerate rating profile id
	if self.ImportId != "" {
		loadId += "_" + self.ImportId
	}

	for i := 0; i < len(tps); i++ {
		tps[i].Tpid = self.TPid
		tps[i].Loadid = loadId

	}
	return self.StorDb.SetTpRatingProfiles(tps)
}

func (self *TPCSVImporter) importSharedGroups(fn string) error {
	if self.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
	tps, err := self.csvr.GetTpSharedGroups(self.TPid, "")
	if err != nil {
		return err
	}
	for i := 0; i < len(tps); i++ {
		tps[i].Tpid = self.TPid
	}

	return self.StorDb.SetTpSharedGroups(tps)
}

func (self *TPCSVImporter) importActions(fn string) error {
	if self.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
	tps, err := self.csvr.GetTpActions(self.TPid, "")
	if err != nil {
		return err
	}
	for i := 0; i < len(tps); i++ {
		tps[i].Tpid = self.TPid
	}

	return self.StorDb.SetTpActions(tps)
}

func (self *TPCSVImporter) importActionTimings(fn string) error {
	if self.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
	tps, err := self.csvr.GetTpActionPlans(self.TPid, "")
	if err != nil {
		return err
	}
	for i := 0; i < len(tps); i++ {
		tps[i].Tpid = self.TPid
	}

	return self.StorDb.SetTpActionPlans(tps)
}

func (self *TPCSVImporter) importActionTriggers(fn string) error {
	if self.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
	tps, err := self.csvr.GetTpActionTriggers(self.TPid, "")
	if err != nil {
		return err
	}
	for i := 0; i < len(tps); i++ {
		tps[i].Tpid = self.TPid
	}

	return self.StorDb.SetTpActionTriggers(tps)
}

func (self *TPCSVImporter) importAccountActions(fn string) error {
	if self.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
	tps, err := self.csvr.GetTpAccountActions(&TpAccountAction{Tpid: self.TPid})
	if err != nil {
		return err
	}
	loadId := utils.CSV_LOAD //Autogenerate rating profile id
	if self.ImportId != "" {
		loadId += "_" + self.ImportId
	}
	for i := 0; i < len(tps); i++ {
		tps[i].Tpid = self.TPid
		tps[i].Loadid = loadId
	}
	return self.StorDb.SetTpAccountActions(tps)
}

func (self *TPCSVImporter) importDerivedChargers(fn string) error {
	if self.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
	tps, err := self.csvr.GetTpDerivedChargers(nil)
	if err != nil {
		return err
	}
	loadId := utils.CSV_LOAD //Autogenerate rating profile id
	if self.ImportId != "" {
		loadId += "_" + self.ImportId
	}
	for i := 0; i < len(tps); i++ {
		tps[i].Tpid = self.TPid
		tps[i].Loadid = loadId
	}
	return self.StorDb.SetTpDerivedChargers(tps)
}

func (self *TPCSVImporter) importCdrStats(fn string) error {
	if self.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
	tps, err := self.csvr.GetTpCdrStats(self.TPid, "")
	if err != nil {
		return err
	}
	for i := 0; i < len(tps); i++ {
		tps[i].Tpid = self.TPid
	}

	return self.StorDb.SetTpCdrStats(tps)
}

func (self *TPCSVImporter) importLcrs(fn string) error {
	if self.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
	tps, err := self.csvr.GetTpLCRs(nil)
	if err != nil {
		return err
	}
	for i := 0; i < len(tps); i++ {
		tps[i].Tpid = self.TPid
	}

	return self.StorDb.SetTpLCRs(tps)
}

func (self *TPCSVImporter) importUsers(fn string) error {
	if self.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
	tps, err := self.csvr.GetTpUsers(nil)
	if err != nil {
		return err
	}
	for i := 0; i < len(tps); i++ {
		tps[i].Tpid = self.TPid
	}

	return self.StorDb.SetTpUsers(tps)
}

func (self *TPCSVImporter) importAliases(fn string) error {
	if self.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
	tps, err := self.csvr.GetTpAliases(nil)
	if err != nil {
		return err
	}
	for i := 0; i < len(tps); i++ {
		tps[i].Tpid = self.TPid
	}
	return self.StorDb.SetTpAliases(tps)
}
