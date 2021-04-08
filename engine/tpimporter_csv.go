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
	StorDB   LoadWriter // StorDb connection handle
	DirPath  string     // Directory path to import from
	Sep      rune       // Separator in the csv file
	Verbose  bool       // If true will print a detailed information instead of silently discarding it
	ImportID string     // Use this to differentiate between imports (eg: when autogenerating fields like RatingProfileID
	csvr     LoadReader
}

// Maps csv file to handler which should process it. Defined like this since tests on 1.0.3 were failing on Travis.
// Change it to func(string) error as soon as Travis updates.
var fileHandlers = map[string]func(*TPCSVImporter, string) error{
	utils.TimingsCsv:            (*TPCSVImporter).importTimings,
	utils.DestinationsCsv:       (*TPCSVImporter).importDestinations,
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
	utils.ActionProfilesCsv:     (*TPCSVImporter).importActionProfiles,
	utils.AccountsCsv:           (*TPCSVImporter).importAccounts,
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

	return tpImp.StorDB.SetTPTimings(tps)
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

	return tpImp.StorDB.SetTPDestinations(tps)
}

func (tpImp *TPCSVImporter) importResources(fn string) error {
	if tpImp.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
	rls, err := tpImp.csvr.GetTPResources(tpImp.TPid, "", "")
	if err != nil {
		return err
	}
	return tpImp.StorDB.SetTPResources(rls)
}

func (tpImp *TPCSVImporter) importStats(fn string) error {
	if tpImp.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
	sts, err := tpImp.csvr.GetTPStats(tpImp.TPid, "", "")
	if err != nil {
		return err
	}
	return tpImp.StorDB.SetTPStats(sts)
}

func (tpImp *TPCSVImporter) importThresholds(fn string) error {
	if tpImp.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
	sts, err := tpImp.csvr.GetTPThresholds(tpImp.TPid, "", "")
	if err != nil {
		return err
	}
	return tpImp.StorDB.SetTPThresholds(sts)
}

func (tpImp *TPCSVImporter) importFilters(fn string) error {
	if tpImp.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
	sts, err := tpImp.csvr.GetTPFilters(tpImp.TPid, "", "")
	if err != nil {
		return err
	}
	return tpImp.StorDB.SetTPFilters(sts)
}

func (tpImp *TPCSVImporter) importRoutes(fn string) error {
	if tpImp.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
	rls, err := tpImp.csvr.GetTPRoutes(tpImp.TPid, "", "")
	if err != nil {
		return err
	}
	return tpImp.StorDB.SetTPRoutes(rls)
}

func (tpImp *TPCSVImporter) importAttributeProfiles(fn string) error {
	if tpImp.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
	rls, err := tpImp.csvr.GetTPAttributes(tpImp.TPid, "", "")
	if err != nil {
		return err
	}
	return tpImp.StorDB.SetTPAttributes(rls)
}

func (tpImp *TPCSVImporter) importChargerProfiles(fn string) error {
	if tpImp.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
	rls, err := tpImp.csvr.GetTPChargers(tpImp.TPid, "", "")
	if err != nil {
		return err
	}
	return tpImp.StorDB.SetTPChargers(rls)
}

func (tpImp *TPCSVImporter) importDispatcherProfiles(fn string) error {
	if tpImp.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
	dpps, err := tpImp.csvr.GetTPDispatcherProfiles(tpImp.TPid, "", "")
	if err != nil {
		return err
	}
	return tpImp.StorDB.SetTPDispatcherProfiles(dpps)
}

func (tpImp *TPCSVImporter) importDispatcherHosts(fn string) error {
	if tpImp.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
	dpps, err := tpImp.csvr.GetTPDispatcherHosts(tpImp.TPid, "", "")
	if err != nil {
		return err
	}
	return tpImp.StorDB.SetTPDispatcherHosts(dpps)
}

func (tpImp *TPCSVImporter) importRateProfiles(fn string) error {
	if tpImp.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
	rpps, err := tpImp.csvr.GetTPRateProfiles(tpImp.TPid, "", "")
	if err != nil {
		return err
	}
	return tpImp.StorDB.SetTPRateProfiles(rpps)
}

func (tpImp *TPCSVImporter) importActionProfiles(fn string) error {
	if tpImp.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
	rpps, err := tpImp.csvr.GetTPActionProfiles(tpImp.TPid, "", "")
	if err != nil {
		return err
	}
	return tpImp.StorDB.SetTPActionProfiles(rpps)
}

func (tpImp *TPCSVImporter) importAccounts(fn string) error {
	if tpImp.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
	rpps, err := tpImp.csvr.GetTPAccounts(tpImp.TPid, "", "")
	if err != nil {
		return err
	}
	return tpImp.StorDB.SetTPAccounts(rpps)
}
