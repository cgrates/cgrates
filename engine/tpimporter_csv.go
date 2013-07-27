/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2013 ITsysCOM

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
	"io"
	"io/ioutil"
	"log"
	"github.com/cgrates/cgrates/utils"
)


// Import tariff plan from csv into storDb
type TPCSVImporter struct {
	TPid     string // Load data on this tpid
	StorDb DataStorage // StorDb connection handle
	DirPath   string // Directory path to import from
	Sep 		rune // Separator in the csv file
	Verbose		bool // If true will print a detailed information instead of silently discarding it
}

var fileHandlers = map[string]func(*TPCSVImporter,string) error{
		utils.TIMINGS_CSV: (*TPCSVImporter).importTimings,
		utils.DESTINATIONS_CSV: (*TPCSVImporter).importDestinations,
		utils.RATES_CSV: (*TPCSVImporter).importRates,
		utils.DESTINATION_RATES_CSV: (*TPCSVImporter).importDestinationRates,
		utils.DESTRATE_TIMINGS_CSV: (*TPCSVImporter).importDestRateTimings,
		utils.RATE_PROFILES_CSV: (*TPCSVImporter).importRatingProfiles,
		utils.ACTIONS_CSV: (*TPCSVImporter).importActions,
		utils.ACTION_TIMINGS_CSV: (*TPCSVImporter).importActionTimings,
		utils.ACTION_TRIGGERS_CSV: (*TPCSVImporter).importActionTriggers,
		utils.ACCOUNT_ACTIONS_CSV: (*TPCSVImporter).importAccountActions,
		}

func (self *TPCSVImporter) Run() error {

	// Maps csv file to handler which should process it
	

	files, _ := ioutil.ReadDir(self.DirPath)
	for _, f := range files {
		fHandler,hasName := fileHandlers[f.Name()]
		if !hasName {
			continue
		}
		fHandler( self, f.Name() )
	}
	return nil
}

// Handler importing timings from file, saved row by row to storDb
func (self *TPCSVImporter) importTimings(fn string) error {
	fParser, err := NewTPCSVFileParser( self.DirPath, fn )
	if err!=nil {
		return err
	}
	lineNr := 0
	for {
		lineNr++
		record, err := fParser.ParseNextLine()
		if err == io.EOF { // Reached end of file
			break
		} else if err != nil {
			if self.Verbose {
				log.Printf("Ignoring line %d, warning: <%s> ", lineNr, err.Error())
			}
			continue
		}
		tm := NewTiming( record... )
		if err := self.StorDb.SetTPTiming(self.TPid, tm); err != nil {
			log.Printf("Ignoring line %d, storDb operational error: <%s> ", lineNr, err.Error())
		}
	}
	return nil
}

func (self *TPCSVImporter) importDestinations(fPath string) error {
	return nil
}

func (self *TPCSVImporter) importRates(fPath string) error {
	return nil
}

func (self *TPCSVImporter) importDestinationRates(fPath string) error {
	return nil
}

func (self *TPCSVImporter) importDestRateTimings(fPath string) error {
	return nil
}

func (self *TPCSVImporter) importRatingProfiles(fPath string) error {
	return nil
}

func (self *TPCSVImporter) importActions(fPath string) error {
	return nil
}

func (self *TPCSVImporter) importActionTimings(fPath string) error {
	return nil
}

func (self *TPCSVImporter) importActionTriggers(fPath string) error {
	return nil
}

func (self *TPCSVImporter) importAccountActions(fPath string) error {
	return nil
}


