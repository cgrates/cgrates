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
	"strconv"
	"time"
	"github.com/cgrates/cgrates/utils"
)


// Import tariff plan from csv into storDb
type TPCSVImporter struct {
	TPid     string // Load data on this tpid
	StorDb DataStorage // StorDb connection handle
	DirPath   string // Directory path to import from
	Sep 		rune // Separator in the csv file
	Verbose		bool // If true will print a detailed information instead of silently discarding it
	ImportId	string // Use this to differentiate between imports (eg: when autogenerating fields like RatingProfileId
}

// Maps csv file to handler which should process it. Defined like this since tests on 1.0.3 were failing on Travis. 
// Change it to func(string) error as soon as Travis updates.
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
	if self.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
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

func (self *TPCSVImporter) importDestinations(fn string) error {
	if self.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
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
		dst := &Destination{record[0], []string{record[1]}}
		if err := self.StorDb.SetTPDestination(self.TPid, dst); err != nil {
			log.Printf("Ignoring line %d, storDb operational error: <%s> ", lineNr, err.Error())
		}
	}
	return nil
}

func (self *TPCSVImporter) importRates(fn string) error {
	if self.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
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
		rt, err := NewRate(record[0], record[1], record[2], record[3], record[4], record[5], record[6], record[7], record[8])
		if err != nil {
			return err
		}
		if err := self.StorDb.SetTPRates( self.TPid, map[string][]*Rate{ record[0]: []*Rate{rt} } ); err != nil {
			log.Printf("Ignoring line %d, storDb operational error: <%s> ", lineNr, err.Error())
		}
	}
	return nil
}

func (self *TPCSVImporter) importDestinationRates(fn string) error {
	if self.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
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
		dr := &DestinationRate{record[0], record[1], record[2], nil} 
		if err := self.StorDb.SetTPDestinationRates( self.TPid, 
			map[string][]*DestinationRate{ dr.Tag: []*DestinationRate{dr} } ); err != nil {
			log.Printf("Ignoring line %d, storDb operational error: <%s> ", lineNr, err.Error())
		}
	}
	return nil
}

func (self *TPCSVImporter) importDestRateTimings(fn string) error {
	if self.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
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
		weight, err := strconv.ParseFloat(record[3], 64)
		if err != nil {
			log.Printf("Ignoring line %d, warning: <%s> ", lineNr, err.Error())
			continue
		}
		drt := &DestinationRateTiming{Tag: record[0], 
					DestinationRatesTag: record[1], 
					Weight: weight, 
					TimingsTag: record[2],
					}
		if err := self.StorDb.SetTPDestRateTimings( self.TPid, map[string][]*DestinationRateTiming{drt.Tag:[]*DestinationRateTiming{drt}}); err != nil {
			log.Printf("Ignoring line %d, storDb operational error: <%s> ", lineNr, err.Error())
		}
	}
	return nil
}

func (self *TPCSVImporter) importRatingProfiles(fn string) error {
	if self.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
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
		tenant, tor, direction, subject, destRatesTimingTag, fallbacksubject := record[0], record[1], record[2], record[3], record[5], record[6]
		at, err := time.Parse(time.RFC3339, record[4])
		if err != nil {
			log.Printf("Ignoring line %d, warning: <%s> ", lineNr, err.Error())
		}
		rpTag := "TPCSV" //Autogenerate rating profile id
		if self.ImportId != "" {
			rpTag += "_"+self.ImportId
		}
		rp := &RatingProfile{Tag: rpTag, 
				Tenant: tenant, 
				TOR: tor, 
				Direction: direction,
				Subject: subject,
				ActivationTime: at.Unix(),
				DestRatesTimingTag: destRatesTimingTag,
				RatesFallbackSubject: fallbacksubject,
				}
		if err := self.StorDb.SetTPRatingProfiles( self.TPid, map[string][]*RatingProfile{rpTag:[]*RatingProfile{rp}}); err != nil {
			log.Printf("Ignoring line %d, storDb operational error: <%s> ", lineNr, err.Error())
		}
	}
	return nil
}

func (self *TPCSVImporter) importActions(fn string) error {
	return nil
}

func (self *TPCSVImporter) importActionTimings(fn string) error {
	return nil
}

func (self *TPCSVImporter) importActionTriggers(fn string) error {
	return nil
}

func (self *TPCSVImporter) importAccountActions(fn string) error {
	return nil
}


