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
	"log"
	"os"

	"github.com/cgrates/cgrates/utils"
)

// Import tariff plan from csv into storDb
type TPCSVImporter struct {
	TPid     string     // Load data on this tpid
	StorDb   LoadWriter // StorDb connection handle
	DirPath  string     // Directory path to import from
	Sep      rune       // Separator in the csv file
	Verbose  bool       // If true will print a detailed information instead of silently discarding it
	ImportId string     // Use this to differentiate between imports (eg: when autogenerating fields like RatingProfileID
	csvr     LoadReader
}

// Maps csv file to handler which should process it. Defined like this since tests on 1.0.3 were failing on Travis.
// Change it to func(string) error as soon as Travis updates.
var fileHandlers = map[string]func(*TPCSVImporter, string) error{
	utils.TimingsCsv:            (*TPCSVImporter).importTimings,
	utils.DestinationsCsv:       (*TPCSVImporter).importDestinations,
	utils.RatesCsv:              (*TPCSVImporter).importRates,
	utils.DestinationRatesCsv:   (*TPCSVImporter).importDestinationRates,
	utils.RatingPlansCsv:        (*TPCSVImporter).importRatingPlans,
	utils.RatingProfilesCsv:     (*TPCSVImporter).importRatingProfiles,
	utils.SharedGroupsCsv:       (*TPCSVImporter).importSharedGroups,
	utils.ActionsCsv:            (*TPCSVImporter).importActions,
	utils.ActionPlansCsv:        (*TPCSVImporter).importActionTimings,
	utils.ActionTriggersCsv:     (*TPCSVImporter).importActionTriggers,
	utils.AccountActionsCsv:     (*TPCSVImporter).importAccountActions,
	utils.ResourcesCsv:          (*TPCSVImporter).importResources,
	utils.StatsCsv:              (*TPCSVImporter).importStats,
	utils.ThresholdsCsv:         (*TPCSVImporter).importThresholds,
	utils.FiltersCsv:            (*TPCSVImporter).importFilters,
	utils.RoutesCsv:             (*TPCSVImporter).importRoutes,
	utils.AttributesCsv:         (*TPCSVImporter).importAttributeProfiles,
	utils.ChargersCsv:           (*TPCSVImporter).importChargerProfiles,
	utils.DispatcherProfilesCsv: (*TPCSVImporter).importDispatcherProfiles,
	utils.DispatcherHostsCsv:    (*TPCSVImporter).importDispatcherHosts,
	utils.RateProfilesCsv:       (*TPCSVImporter).importRateProfiles,
	utils.AccountProfilesCsv:    (*TPCSVImporter).importAccountProfiles,
}

func (tpImp *TPCSVImporter) Run() error {
	tpImp.csvr = NewFileCSVStorage(tpImp.Sep, tpImp.DirPath)
	files, _ := os.ReadDir(tpImp.DirPath)
	var withErrors bool
	for _, f := range files {
		fHandler, hasName := fileHandlers[f.Name()]
		if !hasName {
			continue
		}
		if err := fHandler(tpImp, f.Name()); err != nil {
			withErrors = true
			utils.Logger.Err(fmt.Sprintf("<TPCSVImporter> Importing file: %s, got error: %s", f.Name(), err.Error()))
		}
	}
	if withErrors {
		return utils.ErrPartiallyExecuted
	}
	return nil
}

// Handler importing timings from file, saved row by row to storDb
func (tpImp *TPCSVImporter) importTimings(fn string) error {
	if tpImp.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
	tps, err := tpImp.csvr.GetTPTimings(tpImp.TPid, "")
	if err != nil {
		return err
	}
	for i := 0; i < len(tps); i++ {
		tps[i].TPid = tpImp.TPid
	}

	return tpImp.StorDb.SetTPTimings(tps)
}

func (tpImp *TPCSVImporter) importDestinations(fn string) error {
	if tpImp.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
	tps, err := tpImp.csvr.GetTPDestinations(tpImp.TPid, "")
	if err != nil {
		return err
	}
	for i := 0; i < len(tps); i++ {
		tps[i].TPid = tpImp.TPid
	}

	return tpImp.StorDb.SetTPDestinations(tps)
}

func (tpImp *TPCSVImporter) importRates(fn string) error {
	if tpImp.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
	tps, err := tpImp.csvr.GetTPRates(tpImp.TPid, "")
	if err != nil {
		return err
	}
	for i := 0; i < len(tps); i++ {
		tps[i].TPid = tpImp.TPid
	}

	return tpImp.StorDb.SetTPRates(tps)
}

func (tpImp *TPCSVImporter) importDestinationRates(fn string) error {
	if tpImp.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
	tps, err := tpImp.csvr.GetTPDestinationRates(tpImp.TPid, "", nil)
	if err != nil {
		return err
	}

	for i := 0; i < len(tps); i++ {
		tps[i].TPid = tpImp.TPid
	}

	return tpImp.StorDb.SetTPDestinationRates(tps)
}

func (tpImp *TPCSVImporter) importRatingPlans(fn string) error {
	if tpImp.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
	tps, err := tpImp.csvr.GetTPRatingPlans(tpImp.TPid, "", nil)
	if err != nil {
		return err
	}
	for i := 0; i < len(tps); i++ {
		tps[i].TPid = tpImp.TPid
	}

	return tpImp.StorDb.SetTPRatingPlans(tps)
}

func (tpImp *TPCSVImporter) importRatingProfiles(fn string) error {
	if tpImp.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
	tps, err := tpImp.csvr.GetTPRatingProfiles(&utils.TPRatingProfile{TPid: tpImp.TPid})
	if err != nil {
		return err
	}
	loadId := utils.CSVLoad //Autogenerate rating profile id
	if tpImp.ImportId != "" {
		loadId += "_" + tpImp.ImportId
	}

	for i := 0; i < len(tps); i++ {
		tps[i].TPid = tpImp.TPid
		tps[i].LoadId = loadId

	}
	return tpImp.StorDb.SetTPRatingProfiles(tps)
}

func (tpImp *TPCSVImporter) importSharedGroups(fn string) error {
	if tpImp.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
	tps, err := tpImp.csvr.GetTPSharedGroups(tpImp.TPid, "")
	if err != nil {
		return err
	}
	for i := 0; i < len(tps); i++ {
		tps[i].TPid = tpImp.TPid
	}

	return tpImp.StorDb.SetTPSharedGroups(tps)
}

func (tpImp *TPCSVImporter) importActions(fn string) error {
	if tpImp.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
	tps, err := tpImp.csvr.GetTPActions(tpImp.TPid, "")
	if err != nil {
		return err
	}
	for i := 0; i < len(tps); i++ {
		tps[i].TPid = tpImp.TPid
	}

	return tpImp.StorDb.SetTPActions(tps)
}

func (tpImp *TPCSVImporter) importActionTimings(fn string) error {
	if tpImp.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
	tps, err := tpImp.csvr.GetTPActionPlans(tpImp.TPid, "")
	if err != nil {
		return err
	}
	for i := 0; i < len(tps); i++ {
		tps[i].TPid = tpImp.TPid
	}

	return tpImp.StorDb.SetTPActionPlans(tps)
}

func (tpImp *TPCSVImporter) importActionTriggers(fn string) error {
	if tpImp.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
	tps, err := tpImp.csvr.GetTPActionTriggers(tpImp.TPid, "")
	if err != nil {
		return err
	}
	for i := 0; i < len(tps); i++ {
		tps[i].TPid = tpImp.TPid
	}

	return tpImp.StorDb.SetTPActionTriggers(tps)
}

func (tpImp *TPCSVImporter) importAccountActions(fn string) error {
	if tpImp.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
	tps, err := tpImp.csvr.GetTPAccountActions(&utils.TPAccountActions{TPid: tpImp.TPid})
	if err != nil {
		return err
	}
	loadId := utils.CSVLoad //Autogenerate rating profile id
	if tpImp.ImportId != "" {
		loadId += "_" + tpImp.ImportId
	}
	for i := 0; i < len(tps); i++ {
		tps[i].TPid = tpImp.TPid
		tps[i].LoadId = loadId
	}
	return tpImp.StorDb.SetTPAccountActions(tps)
}

func (tpImp *TPCSVImporter) importResources(fn string) error {
	if tpImp.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
	rls, err := tpImp.csvr.GetTPResources(tpImp.TPid, "", "")
	if err != nil {
		return err
	}
	return tpImp.StorDb.SetTPResources(rls)
}

func (tpImp *TPCSVImporter) importStats(fn string) error {
	if tpImp.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
	sts, err := tpImp.csvr.GetTPStats(tpImp.TPid, "", "")
	if err != nil {
		return err
	}
	return tpImp.StorDb.SetTPStats(sts)
}

func (tpImp *TPCSVImporter) importThresholds(fn string) error {
	if tpImp.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
	sts, err := tpImp.csvr.GetTPThresholds(tpImp.TPid, "", "")
	if err != nil {
		return err
	}
	return tpImp.StorDb.SetTPThresholds(sts)
}

func (tpImp *TPCSVImporter) importFilters(fn string) error {
	if tpImp.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
	sts, err := tpImp.csvr.GetTPFilters(tpImp.TPid, "", "")
	if err != nil {
		return err
	}
	return tpImp.StorDb.SetTPFilters(sts)
}

func (tpImp *TPCSVImporter) importRoutes(fn string) error {
	if tpImp.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
	rls, err := tpImp.csvr.GetTPRoutes(tpImp.TPid, "", "")
	if err != nil {
		return err
	}
	return tpImp.StorDb.SetTPRoutes(rls)
}

func (tpImp *TPCSVImporter) importAttributeProfiles(fn string) error {
	if tpImp.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
	rls, err := tpImp.csvr.GetTPAttributes(tpImp.TPid, "", "")
	if err != nil {
		return err
	}
	return tpImp.StorDb.SetTPAttributes(rls)
}

func (tpImp *TPCSVImporter) importChargerProfiles(fn string) error {
	if tpImp.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
	rls, err := tpImp.csvr.GetTPChargers(tpImp.TPid, "", "")
	if err != nil {
		return err
	}
	return tpImp.StorDb.SetTPChargers(rls)
}

func (tpImp *TPCSVImporter) importDispatcherProfiles(fn string) error {
	if tpImp.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
	dpps, err := tpImp.csvr.GetTPDispatcherProfiles(tpImp.TPid, "", "")
	if err != nil {
		return err
	}
	return tpImp.StorDb.SetTPDispatcherProfiles(dpps)
}

func (tpImp *TPCSVImporter) importDispatcherHosts(fn string) error {
	if tpImp.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
	dpps, err := tpImp.csvr.GetTPDispatcherHosts(tpImp.TPid, "", "")
	if err != nil {
		return err
	}
	return tpImp.StorDb.SetTPDispatcherHosts(dpps)
}

func (tpImp *TPCSVImporter) importRateProfiles(fn string) error {
	if tpImp.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
	rpps, err := tpImp.csvr.GetTPRateProfiles(tpImp.TPid, "", "")
	if err != nil {
		return err
	}
	return tpImp.StorDb.SetTPRateProfiles(rpps)
}

func (tpImp *TPCSVImporter) importAccountProfiles(fn string) error {
	if tpImp.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
	rpps, err := tpImp.csvr.GetTPAccountProfiles(tpImp.TPid, "", "")
	if err != nil {
		return err
	}
	return tpImp.StorDb.SetTPAccountProfiles(rpps)
}
