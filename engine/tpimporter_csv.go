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
	utils.RATING_PLANS_CSV:      (*TPCSVImporter).importRatingPlans,
	utils.RATING_PROFILES_CSV:   (*TPCSVImporter).importRatingProfiles,
	utils.SHARED_GROUPS_CSV:     (*TPCSVImporter).importSharedGroups,
	utils.ACTIONS_CSV:           (*TPCSVImporter).importActions,
	utils.ACTION_PLANS_CSV:      (*TPCSVImporter).importActionTimings,
	utils.ACTION_TRIGGERS_CSV:   (*TPCSVImporter).importActionTriggers,
	utils.ACCOUNT_ACTIONS_CSV:   (*TPCSVImporter).importAccountActions,
	utils.DERIVED_CHARGERS_CSV:  (*TPCSVImporter).importDerivedChargers,
	utils.CDR_STATS_CSV:         (*TPCSVImporter).importCdrStats,
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
	if self.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
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
	if self.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
	fParser, err := NewTPCSVFileParser(self.DirPath, fn)
	if err != nil {
		return err
	}
	lineNr := 0
	dests := make(map[string]*Destination) // Key:destId, value: listOfPrefixes
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
		} else {
			if dst, hasIt := dests[record[0]]; hasIt {
				dst.Prefixes = append(dst.Prefixes, record[1])
			} else {
				dests[record[0]] = &Destination{record[0], []string{record[1]}}
			}
		}
	}
	for _, dst := range dests {
		if err := self.StorDb.SetTPDestination(self.TPid, dst); err != nil {
			if self.Verbose {
				log.Printf("Ignoring line %d, storDb operational error: <%s> ", lineNr, err.Error())
			}
		}
	}
	return nil
}

func (self *TPCSVImporter) importRates(fn string) error {
	if self.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
	fParser, err := NewTPCSVFileParser(self.DirPath, fn)
	if err != nil {
		return err
	}
	lineNr := 0
	rates := make(map[string][]*utils.RateSlot)
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
		newRt, err := NewLoadRate(record[0], record[1], record[2], record[3], record[4], record[5])
		if err != nil {
			return err
		}
		if _, hasIt := rates[record[0]]; !hasIt {
			rates[record[0]] = make([]*utils.RateSlot, 0)
		}
		rates[record[0]] = append(rates[record[0]], newRt.RateSlots...)
	}
	if err := self.StorDb.SetTPRates(self.TPid, rates); err != nil {
		if self.Verbose {
			log.Printf("Ignoring line %d, storDb operational error: <%s> ", lineNr, err.Error())
		}
	}
	return nil
}

func (self *TPCSVImporter) importDestinationRates(fn string) error {
	if self.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
	fParser, err := NewTPCSVFileParser(self.DirPath, fn)
	if err != nil {
		return err
	}
	lineNr := 0
	drs := make(map[string][]*utils.DestinationRate)
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
		roundingDecimals, err := strconv.Atoi(record[4])
		if err != nil {
			log.Printf("Error parsing rounding decimals: %s", record[4])
			return err
		}
		if _, hasIt := drs[record[0]]; !hasIt {
			drs[record[0]] = make([]*utils.DestinationRate, 0)
		}
		drs[record[0]] = append(drs[record[0]], &utils.DestinationRate{
			DestinationId:    record[1],
			RateId:           record[2],
			RoundingMethod:   record[3],
			RoundingDecimals: roundingDecimals,
		})
	}

	if err := self.StorDb.SetTPDestinationRates(self.TPid, drs); err != nil {
		if self.Verbose {
			log.Printf("Ignoring line %d, storDb operational error: <%s> ", lineNr, err.Error())
		}
	}

	return nil
}

func (self *TPCSVImporter) importRatingPlans(fn string) error {
	if self.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
	fParser, err := NewTPCSVFileParser(self.DirPath, fn)
	if err != nil {
		return err
	}
	lineNr := 0
	rpls := make(map[string][]*utils.TPRatingPlanBinding)
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
		if _, hasIt := rpls[record[0]]; !hasIt {
			rpls[record[0]] = make([]*utils.TPRatingPlanBinding, 0)
		}
		rpls[record[0]] = append(rpls[record[0]], &utils.TPRatingPlanBinding{
			DestinationRatesId: record[1],
			Weight:             weight,
			TimingId:           record[2],
		})
	}
	if err := self.StorDb.SetTPRatingPlans(self.TPid, rpls); err != nil {
		if self.Verbose {
			log.Printf("Ignoring line %d, storDb operational error: <%s> ", lineNr, err.Error())
		}
	}

	return nil
}

func (self *TPCSVImporter) importRatingProfiles(fn string) error {
	if self.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
	fParser, err := NewTPCSVFileParser(self.DirPath, fn)
	if err != nil {
		return err
	}
	lineNr := 0
	rpfs := make(map[string]*utils.TPRatingProfile)
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
		direction, tenant, tor, subject, ratingPlanTag, fallbacksubject := record[0], record[1], record[2], record[3], record[5], record[6]
		_, err = utils.ParseDate(record[4])
		if err != nil {
			if self.Verbose {
				log.Printf("Ignoring line %d, warning: <%s> ", lineNr, err.Error())
			}
			continue
		}
		loadId := utils.CSV_LOAD //Autogenerate rating profile id
		if self.ImportId != "" {
			loadId += "_" + self.ImportId
		}
		newRp := &utils.TPRatingProfile{
			TPid:      self.TPid,
			LoadId:    loadId,
			Tenant:    tenant,
			Category:  tor,
			Direction: direction,
			Subject:   subject,
			RatingPlanActivations: []*utils.TPRatingActivation{
				&utils.TPRatingActivation{ActivationTime: record[4], RatingPlanId: ratingPlanTag, FallbackSubjects: fallbacksubject}},
		}
		if rp, hasIt := rpfs[newRp.KeyId()]; hasIt {
			rp.RatingPlanActivations = append(rp.RatingPlanActivations, newRp.RatingPlanActivations...)
		} else {
			rpfs[newRp.KeyId()] = newRp
		}
	}
	if err := self.StorDb.SetTPRatingProfiles(self.TPid, rpfs); err != nil {
		if self.Verbose {
			log.Printf("Ignoring line %d, storDb operational error: <%s> ", lineNr, err.Error())
		}
	}

	return nil
}

func (self *TPCSVImporter) importSharedGroups(fn string) error {
	if self.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
	fParser, err := NewTPCSVFileParser(self.DirPath, fn)
	if err != nil {
		return err
	}
	shgs := make(map[string][]*utils.TPSharedGroup)
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
		if _, hasIt := shgs[record[0]]; !hasIt {
			shgs[record[0]] = make([]*utils.TPSharedGroup, 0)
		}
		shgs[record[0]] = append(shgs[record[0]], &utils.TPSharedGroup{Account: record[1], Strategy: record[2], RatingSubject: record[3]})
	}
	if err := self.StorDb.SetTPSharedGroups(self.TPid, shgs); err != nil {
		if self.Verbose {
			log.Printf("Ignoring line %d, storDb operational error: <%s> ", lineNr, err.Error())
		}
	}
	return nil
}

func (self *TPCSVImporter) importActions(fn string) error {
	if self.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
	fParser, err := NewTPCSVFileParser(self.DirPath, fn)
	if err != nil {
		return err
	}
	fieldIndex := CSV_FIELD_INDEX[utils.ACTIONS_CSV]
	acts := make(map[string][]*utils.TPAction)
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
		actId, actionType, balanceType, direction, destTag, rateSubject, category, sharedGroup := record[fieldIndex[utils.CSVFLD_ACTIONS_TAG]], record[fieldIndex[utils.CSVFLD_ACTION]],
			record[fieldIndex[utils.CSVFLD_BALANCE_TYPE]], record[fieldIndex[utils.CSVFLD_DIRECTION]], record[fieldIndex[utils.CSVFLD_DESTINATION_TAG]], record[fieldIndex[utils.CSVFLD_RATING_SUBJECT]],
			record[fieldIndex[utils.CSVFLD_CATEGORY]], record[fieldIndex[utils.CSVFLD_SHARED_GROUP]]
		units, err := strconv.ParseFloat(record[fieldIndex[utils.CSVFLD_UNITS]], 64)
		if err != nil && record[fieldIndex[utils.CSVFLD_UNITS]] != "" {
			if self.Verbose {
				log.Printf("Ignoring line %d, warning: <%s> ", lineNr, err.Error())
			}
			continue
		}
		balanceWeight, _ := strconv.ParseFloat(record[fieldIndex[utils.CSVFLD_BALANCE_WEIGHT]], 64)
		weight, err := strconv.ParseFloat(record[fieldIndex[utils.CSVFLD_WEIGHT]], 64)
		if err != nil {
			if self.Verbose {
				log.Printf("Ignoring line %d, warning: <%s> ", lineNr, err.Error())
			}
			continue
		}
		if _, hasIt := acts[actId]; !hasIt {
			acts[actId] = make([]*utils.TPAction, 0)
		}
		acts[actId] = append(acts[actId], &utils.TPAction{
			Identifier:      actionType,
			BalanceType:     balanceType,
			Direction:       direction,
			Units:           units,
			ExpiryTime:      record[5],
			DestinationId:   destTag,
			RatingSubject:   rateSubject,
			Category:        category,
			SharedGroup:     sharedGroup,
			BalanceWeight:   balanceWeight,
			ExtraParameters: record[fieldIndex[utils.CSVFLD_EXTRA_PARAMS]],
			Weight:          weight,
		})
	}
	if err := self.StorDb.SetTPActions(self.TPid, acts); err != nil {
		if self.Verbose {
			log.Printf("Ignoring line %d, storDb operational error: <%s> ", lineNr, err.Error())
		}

	}
	return nil
}

func (self *TPCSVImporter) importActionTimings(fn string) error {
	if self.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
	fParser, err := NewTPCSVFileParser(self.DirPath, fn)
	if err != nil {
		return err
	}
	lineNr := 0
	aplns := make(map[string][]*utils.TPActionTiming)
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
		if _, hasIt := aplns[tag]; !hasIt {
			aplns[tag] = make([]*utils.TPActionTiming, 0)
		}
		aplns[tag] = append(aplns[tag], &utils.TPActionTiming{
			ActionsId: actionsTag,
			TimingId:  timingTag,
			Weight:    weight,
		})
	}
	if err := self.StorDb.SetTPActionTimings(self.TPid, aplns); err != nil {
		if self.Verbose {
			log.Printf("Ignoring line %d, storDb operational error: <%s> ", lineNr, err.Error())
		}
	}

	return nil
}

func (self *TPCSVImporter) importActionTriggers(fn string) error {
	if self.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
	fParser, err := NewTPCSVFileParser(self.DirPath, fn)
	if err != nil {
		return err
	}
	lineNr := 0
	atrs := make(map[string][]*utils.TPActionTrigger)
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
		tag, balanceType, direction, thresholdType, destinationTag, balanceExpirationDate, balanceRatingSubject, balanceCategory, balanceSharedGroup, actionsTag := record[0], record[1], record[2], record[3], record[7], record[9], record[10], record[10], record[12], record[14]
		threshold, err := strconv.ParseFloat(record[4], 64)
		if err != nil {
			if self.Verbose {
				log.Printf("Ignoring line %d, warning: <%s> ", lineNr, err.Error())
			}
			continue
		}
		recurrent, err := strconv.ParseBool(record[5])
		if err != nil {
			log.Printf("Ignoring line %d, warning: <%s>", lineNr, err.Error())
			continue
		}
		minSleep, err := time.ParseDuration(record[6])
		if err != nil && record[6] != "" {
			log.Printf("Ignoring line %d, warning: <%s>", lineNr, err.Error())
			continue
		}
		balanceWeight, err := strconv.ParseFloat(record[8], 64)
		if err != nil && record[8] != "" {
			if self.Verbose {
				log.Printf("Ignoring line %d, warning: <%s> ", lineNr, err.Error())
			}
			continue
		}
		minQueuedItems, err := strconv.Atoi(record[13])
		if err != nil && record[12] != "" {
			log.Printf("Ignoring line %d, warning: <%s>", lineNr, err.Error())
			continue
		}
		weight, err := strconv.ParseFloat(record[15], 64)
		if err != nil {
			if self.Verbose {
				log.Printf("Ignoring line %d, warning: <%s> ", lineNr, err.Error())
			}
			continue
		}
		if _, hasIt := atrs[tag]; !hasIt {
			atrs[tag] = make([]*utils.TPActionTrigger, 0)
		}
		atrs[tag] = append(atrs[tag], &utils.TPActionTrigger{
			BalanceType:           balanceType,
			Direction:             direction,
			ThresholdType:         thresholdType,
			ThresholdValue:        threshold,
			Recurrent:             recurrent,
			MinSleep:              minSleep,
			DestinationId:         destinationTag,
			BalanceWeight:         balanceWeight,
			BalanceExpirationDate: balanceExpirationDate,
			BalanceRatingSubject:  balanceRatingSubject,
			BalanceCategory:       balanceCategory,
			BalanceSharedGroup:    balanceSharedGroup,
			MinQueuedItems:        minQueuedItems,
			Weight:                weight,
			ActionsId:             actionsTag,
		})
	}
	if err := self.StorDb.SetTPActionTriggers(self.TPid, atrs); err != nil {
		if self.Verbose {
			log.Printf("Ignoring line %d, storDb operational error: <%s> ", lineNr, err.Error())
		}
	}

	return nil
}

func (self *TPCSVImporter) importAccountActions(fn string) error {
	if self.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
	fParser, err := NewTPCSVFileParser(self.DirPath, fn)
	if err != nil {
		return err
	}
	loadId := utils.CSV_LOAD //Autogenerate account actions profile id
	if self.ImportId != "" {
		loadId += "_" + self.ImportId
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

		tpaa := &utils.TPAccountActions{TPid: self.TPid, LoadId: loadId, Tenant: tenant, Account: account, Direction: direction,
			ActionPlanId: actionTimingsTag, ActionTriggersId: actionTriggersTag}
		aa := map[string]*utils.TPAccountActions{tpaa.KeyId(): tpaa}
		if err := self.StorDb.SetTPAccountActions(self.TPid, aa); err != nil {
			if self.Verbose {
				log.Printf("Ignoring line %d, storDb operational error: <%s> ", lineNr, err.Error())
			}
		}
	}
	return nil
}

func (self *TPCSVImporter) importDerivedChargers(fn string) error {
	if self.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
	fParser, err := NewTPCSVFileParser(self.DirPath, fn)
	if err != nil {
		return err
	}
	loadId := utils.CSV_LOAD //Autogenerate account actions profile id
	if self.ImportId != "" {
		loadId += "_" + self.ImportId
	}
	dcs := make(map[string][]*utils.TPDerivedCharger)
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
		newDcs := utils.TPDerivedChargers{TPid: self.TPid,
			Loadid:    loadId,
			Direction: record[0],
			Tenant:    record[1],
			Category:  record[2],
			Account:   record[3],
			Subject:   record[4]}
		dcsId := newDcs.GetDerivedChargesId()

		if _, hasIt := dcs[dcsId]; !hasIt {
			dcs[dcsId] = make([]*utils.TPDerivedCharger, 0)
		}
		dcs[dcsId] = append(dcs[dcsId], &utils.TPDerivedCharger{
			RunId:            ValueOrDefault(record[5], "*default"),
			RunFilters:       record[6],
			ReqTypeField:     ValueOrDefault(record[7], "*default"),
			DirectionField:   ValueOrDefault(record[8], "*default"),
			TenantField:      ValueOrDefault(record[9], "*default"),
			CategoryField:    ValueOrDefault(record[10], "*default"),
			AccountField:     ValueOrDefault(record[11], "*default"),
			SubjectField:     ValueOrDefault(record[12], "*default"),
			DestinationField: ValueOrDefault(record[13], "*default"),
			SetupTimeField:   ValueOrDefault(record[14], "*default"),
			AnswerTimeField:  ValueOrDefault(record[15], "*default"),
			UsageField:       ValueOrDefault(record[16], "*default"),
		})
	}
	if err := self.StorDb.SetTPDerivedChargers(self.TPid, dcs); err != nil {
		if self.Verbose {
			log.Printf("Ignoring line %d, storDb operational error: <%s> ", lineNr, err.Error())
		}
	}
	return nil
}

func (self *TPCSVImporter) importCdrStats(fn string) error {
	if self.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
	fParser, err := NewTPCSVFileParser(self.DirPath, fn)
	if err != nil {
		return err
	}
	css := make(map[string][]*utils.TPCdrStat)
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
		if len(record[1]) == 0 {
			record[1] = "0" // Empty value will be translated to 0 as QueueLength
		}
		if _, err = strconv.Atoi(record[1]); err != nil {
			log.Printf("Ignoring line %d, warning: <%s>", lineNr, err.Error())
			continue
		}
		if _, hasIt := css[record[0]]; !hasIt {
			css[record[0]] = make([]*utils.TPCdrStat, 0)
		}
		css[record[0]] = append(css[record[0]], &utils.TPCdrStat{
			QueueLength:       record[1],
			TimeWindow:        ValueOrDefault(record[2], "0"),
			Metrics:           record[3],
			SetupInterval:     record[4],
			TOR:               record[5],
			CdrHost:           record[6],
			CdrSource:         record[7],
			ReqType:           record[8],
			Direction:         record[9],
			Tenant:            record[10],
			Category:          record[11],
			Account:           record[12],
			Subject:           record[13],
			DestinationPrefix: record[14],
			UsageInterval:     record[15],
			MediationRunIds:   record[16],
			RatedAccount:      record[17],
			RatedSubject:      record[18],
			CostInterval:      record[19],
			ActionTriggers:    record[20],
		})
	}
	if err := self.StorDb.SetTPCdrStats(self.TPid, css); err != nil {
		if self.Verbose {
			log.Printf("Ignoring line %d, storDb operational error: <%s> ", lineNr, err.Error())
		}
	}
	return nil
}
