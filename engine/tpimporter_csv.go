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
	"github.com/cgrates/cgrates/utils"
	"io"
	"io/ioutil"
	"log"
	"strconv"
)

// Import tariff plan from csv into storDb
type TPCSVImporter struct {
	TPid     string      // Load data on this tpid
	StorDb   LoadStorage // StorDb connection handle
	DirPath  string      // Directory path to import from
	Sep      rune        // Separator in the csv file
	Verbose  bool        // If true will print a detailed information instead of silently discarding it
	ImportId string      // Use this to differentiate between imports (eg: when autogenerating fields like RatingProfileId
}

// Maps csv file to handler which should process it. Defined like this since tests on 1.0.3 were failing on Travis.
// Change it to func(string) error as soon as Travis updates.
var fileHandlers = map[string]func(*TPCSVImporter, string) error{
	utils.TIMINGS_CSV:           (*TPCSVImporter).importTimings,
	utils.DESTINATIONS_CSV:      (*TPCSVImporter).importDestinations,
	utils.RATES_CSV:             (*TPCSVImporter).importRates,
	utils.DESTINATION_RATES_CSV: (*TPCSVImporter).importDestinationRates,
	utils.DESTRATE_TIMINGS_CSV:  (*TPCSVImporter).importDestRateTimings,
	utils.RATE_PROFILES_CSV:     (*TPCSVImporter).importRatingProfiles,
	utils.ACTIONS_CSV:           (*TPCSVImporter).importActions,
	utils.ACTION_TIMINGS_CSV:    (*TPCSVImporter).importActionTimings,
	utils.ACTION_TRIGGERS_CSV:   (*TPCSVImporter).importActionTriggers,
	utils.ACCOUNT_ACTIONS_CSV:   (*TPCSVImporter).importAccountActions,
}

func (self *TPCSVImporter) Run() error {
	files, _ := ioutil.ReadDir(self.DirPath)
	for _, f := range files {
		fHandler, hasName := fileHandlers[f.Name()]
		if !hasName {
			continue
		}
		fHandler(self, f.Name())
	}
	return nil
}

// Handler importing timings from file, saved row by row to storDb
func (self *TPCSVImporter) importTimings(fn string) error {
	log.Printf("Processing file: <%s> ", fn)
	fParser, err := NewTPCSVFileParser(self.DirPath, fn)
	if err != nil {
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
		tm := NewTiming(record...)
		if err := self.StorDb.SetTPTiming(self.TPid, tm); err != nil {
			if self.Verbose {
				log.Printf("Ignoring line %d, storDb operational error: <%s> ", lineNr, err.Error())
			}
		}
	}
	return nil
}

func (self *TPCSVImporter) importDestinations(fn string) error {
	log.Printf("Processing file: <%s> ", fn)
	fParser, err := NewTPCSVFileParser(self.DirPath, fn)
	if err != nil {
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
			if self.Verbose {
				log.Printf("Ignoring line %d, storDb operational error: <%s> ", lineNr, err.Error())
			}
		}
	}
	return nil
}

func (self *TPCSVImporter) importRates(fn string) error {
	log.Printf("Processing file: <%s> ", fn)
	fParser, err := NewTPCSVFileParser(self.DirPath, fn)
	if err != nil {
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
		rt, err := NewLoadRate(record[0], record[1], record[2], record[3], record[4], record[5], record[6], record[7], record[8])
		if err != nil {
			return err
		}
		if err := self.StorDb.SetTPRates(self.TPid, map[string][]*LoadRate{record[0]: []*LoadRate{rt}}); err != nil {
			if self.Verbose {
				log.Printf("Ignoring line %d, storDb operational error: <%s> ", lineNr, err.Error())
			}
		}
	}
	return nil
}

func (self *TPCSVImporter) importDestinationRates(fn string) error {
	log.Printf("Processing file: <%s> ", fn)
	fParser, err := NewTPCSVFileParser(self.DirPath, fn)
	if err != nil {
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
		if err := self.StorDb.SetTPDestinationRates(self.TPid,
			map[string][]*DestinationRate{dr.Tag: []*DestinationRate{dr}}); err != nil {
			if self.Verbose {
				log.Printf("Ignoring line %d, storDb operational error: <%s> ", lineNr, err.Error())
			}
		}
	}
	return nil
}

func (self *TPCSVImporter) importDestRateTimings(fn string) error {
	log.Printf("Processing file: <%s> ", fn)
	fParser, err := NewTPCSVFileParser(self.DirPath, fn)
	if err != nil {
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
			if self.Verbose {
				log.Printf("Ignoring line %d, warning: <%s> ", lineNr, err.Error())
			}
			continue
		}
		drt := &DestinationRateTiming{Tag: record[0],
			DestinationRatesTag: record[1],
			Weight:              weight,
			TimingTag:           record[2],
		}
		if err := self.StorDb.SetTPDestRateTimings(self.TPid, map[string][]*DestinationRateTiming{drt.Tag: []*DestinationRateTiming{drt}}); err != nil {
			if self.Verbose {
				log.Printf("Ignoring line %d, storDb operational error: <%s> ", lineNr, err.Error())
			}
		}
	}
	return nil
}

func (self *TPCSVImporter) importRatingProfiles(fn string) error {
	log.Printf("Processing file: <%s> ", fn)
	fParser, err := NewTPCSVFileParser(self.DirPath, fn)
	if err != nil {
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
		_, err = utils.ParseDate(record[4])
		if err != nil {
			if self.Verbose {
				log.Printf("Ignoring line %d, warning: <%s> ", lineNr, err.Error())
			}
			continue
		}
		rpTag := "TPCSV" //Autogenerate rating profile id
		if self.ImportId != "" {
			rpTag += "_" + self.ImportId
		}
		rp := &RatingProfile{Tag: rpTag,
			Tenant:               tenant,
			TOR:                  tor,
			Direction:            direction,
			Subject:              subject,
			ActivationTime:       record[4],
			DestRatesTimingTag:   destRatesTimingTag,
			RatesFallbackSubject: fallbacksubject,
		}
		if err := self.StorDb.SetTPRatingProfiles(self.TPid, map[string][]*RatingProfile{rpTag: []*RatingProfile{rp}}); err != nil {
			if self.Verbose {
				log.Printf("Ignoring line %d, storDb operational error: <%s> ", lineNr, err.Error())
			}
		}
	}
	return nil
}

func (self *TPCSVImporter) importActions(fn string) error {
	log.Printf("Processing file: <%s> ", fn)
	fParser, err := NewTPCSVFileParser(self.DirPath, fn)
	if err != nil {
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
		actId, actionType, balanceType, direction, destTag, rateSubject := record[0], record[1], record[2], record[3], record[6], record[7]
		units, err := strconv.ParseFloat(record[4], 64)
		if err != nil {
			if self.Verbose {
				log.Printf("Ignoring line %d, warning: <%s> ", lineNr, err.Error())
			}
			continue
		}
		balanceWeight, _ := strconv.ParseFloat(record[9], 64)
		weight, err := strconv.ParseFloat(record[10], 64)
		if err != nil {
			if self.Verbose {
				log.Printf("Ignoring line %d, warning: <%s> ", lineNr, err.Error())
			}
			continue
		}
		act := &Action{
			ActionType:       actionType,
			BalanceId:        balanceType,
			Direction:        direction,
			ExpirationString: record[5],
			ExtraParameters:  record[8],
			Balance: &Balance{
				Value:         units,
				DestinationId: destTag,
				RateSubject:   rateSubject,
				Weight:        balanceWeight,
			},
			Weight: weight,
		}
		if err := self.StorDb.SetTPActions(self.TPid, map[string][]*Action{actId: []*Action{act}}); err != nil {
			if self.Verbose {
				log.Printf("Ignoring line %d, storDb operational error: <%s> ", lineNr, err.Error())
			}
		}
	}
	return nil
}

func (self *TPCSVImporter) importActionTimings(fn string) error {
	log.Printf("Processing file: <%s> ", fn)
	fParser, err := NewTPCSVFileParser(self.DirPath, fn)
	if err != nil {
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
		tag, actionsTag, timingTag := record[0], record[1], record[2]
		weight, err := strconv.ParseFloat(record[3], 64)
		if err != nil {
			if self.Verbose {
				log.Printf("Ignoring line %d, warning: <%s> ", lineNr, err.Error())
			}
			continue
		}
		at := &ActionTiming{
			Tag:        tag,
			ActionsTag: actionsTag,
			TimingsTag: timingTag,
			Weight:     weight,
		}
		if err := self.StorDb.SetTPActionTimings(self.TPid, map[string][]*ActionTiming{tag: []*ActionTiming{at}}); err != nil {
			if self.Verbose {
				log.Printf("Ignoring line %d, storDb operational error: <%s> ", lineNr, err.Error())
			}
		}
	}
	return nil
}

func (self *TPCSVImporter) importActionTriggers(fn string) error {
	log.Printf("Processing file: <%s> ", fn)
	fParser, err := NewTPCSVFileParser(self.DirPath, fn)
	if err != nil {
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
		tag, balanceType, direction, thresholdType, destinationTag, actionsTag := record[0], record[1], record[2], record[3], record[5], record[6]
		threshold, err := strconv.ParseFloat(record[4], 64)
		if err != nil {
			if self.Verbose {
				log.Printf("Ignoring line %d, warning: <%s> ", lineNr, err.Error())
			}
			continue
		}
		weight, err := strconv.ParseFloat(record[7], 64)
		if err != nil {
			if self.Verbose {
				log.Printf("Ignoring line %d, warning: <%s> ", lineNr, err.Error())
			}
			continue
		}
		at := &ActionTrigger{
			BalanceId:      balanceType,
			Direction:      direction,
			ThresholdType:  thresholdType,
			ThresholdValue: threshold,
			DestinationId:  destinationTag,
			Weight:         weight,
			ActionsId:      actionsTag,
		}
		if err := self.StorDb.SetTPActionTriggers(self.TPid, map[string][]*ActionTrigger{tag: []*ActionTrigger{at}}); err != nil {
			if self.Verbose {
				log.Printf("Ignoring line %d, storDb operational error: <%s> ", lineNr, err.Error())
			}
		}
	}
	return nil
}

func (self *TPCSVImporter) importAccountActions(fn string) error {
	log.Printf("Processing file: <%s> ", fn)
	fParser, err := NewTPCSVFileParser(self.DirPath, fn)
	if err != nil {
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
		tenant, account, direction, actionTimingsTag, actionTriggersTag := record[0], record[1], record[2], record[3], record[4]
		tag := "TPCSV" //Autogenerate account actions profile id
		if self.ImportId != "" {
			tag += "_" + self.ImportId
		}
		aa := map[string]*AccountAction{
			tag: &AccountAction{Tenant: tenant, Account: account, Direction: direction,
				ActionTimingsTag: actionTimingsTag, ActionTriggersTag: actionTriggersTag},
		}
		if err := self.StorDb.SetTPAccountActions(self.TPid, aa); err != nil {
			if self.Verbose {
				log.Printf("Ignoring line %d, storDb operational error: <%s> ", lineNr, err.Error())
			}
		}
	}
	return nil
}
