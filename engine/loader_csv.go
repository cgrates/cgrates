/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can Storagetribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITH*out ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package engine

import (
	"encoding/csv"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cgrates/cgrates/utils"
)

type CSVReader struct {
	sep               rune
	dataStorage       RatingStorage
	accountingStorage AccountingStorage
	readerFunc        func(string, rune, int) (*csv.Reader, *os.File, error)
	tp                *TPData
	// file names
	destinationsFn, ratesFn, destinationratesFn, timingsFn, destinationratetimingsFn, ratingprofilesFn,
	sharedgroupsFn, lcrFn, actionsFn, actiontimingsFn, actiontriggersFn, accountactionsFn, derivedChargersFn, cdrStatsFn string
}

func NewFileCSVReader(dataStorage RatingStorage, accountingStorage AccountingStorage, sep rune,
	destinationsFn, timingsFn, ratesFn, destinationratesFn, destinationratetimingsFn, ratingprofilesFn, sharedgroupsFn, lcrFn,
	actionsFn, actiontimingsFn, actiontriggersFn, accountactionsFn, derivedChargersFn, cdrStatsFn string) *CSVReader {
	c := new(CSVReader)
	c.sep = sep
	c.dataStorage = dataStorage
	c.accountingStorage = accountingStorage
	c.tp = NewTPData()
	c.readerFunc = openFileCSVReader
	c.destinationsFn, c.timingsFn, c.ratesFn, c.destinationratesFn, c.destinationratetimingsFn, c.ratingprofilesFn,
		c.sharedgroupsFn, c.lcrFn, c.actionsFn, c.actiontimingsFn, c.actiontriggersFn, c.accountactionsFn, c.derivedChargersFn, c.cdrStatsFn = destinationsFn, timingsFn,
		ratesFn, destinationratesFn, destinationratetimingsFn, ratingprofilesFn, sharedgroupsFn, lcrFn, actionsFn, actiontimingsFn, actiontriggersFn, accountactionsFn, derivedChargersFn, cdrStatsFn
	return c
}

func NewStringCSVReader(dataStorage RatingStorage, accountingStorage AccountingStorage, sep rune,
	destinationsFn, timingsFn, ratesFn, destinationratesFn, destinationratetimingsFn, ratingprofilesFn, sharedgroupsFn, lcrFn,
	actionsFn, actiontimingsFn, actiontriggersFn, accountactionsFn, derivedChargersFn, cdrStatsFn string) *CSVReader {
	c := NewFileCSVReader(dataStorage, accountingStorage, sep, destinationsFn, timingsFn, ratesFn, destinationratesFn, destinationratetimingsFn,
		ratingprofilesFn, sharedgroupsFn, lcrFn, actionsFn, actiontimingsFn, actiontriggersFn, accountactionsFn, derivedChargersFn, cdrStatsFn)
	c.readerFunc = openStringCSVReader
	return c
}

func openFileCSVReader(fn string, comma rune, nrFields int) (csvReader *csv.Reader, fp *os.File, err error) {
	fp, err = os.Open(fn)
	if err != nil {
		return
	}
	csvReader = csv.NewReader(fp)
	csvReader.Comma = comma
	csvReader.Comment = utils.COMMENT_CHAR
	csvReader.FieldsPerRecord = nrFields
	csvReader.TrailingComma = true
	return
}

func openStringCSVReader(data string, comma rune, nrFields int) (csvReader *csv.Reader, fp *os.File, err error) {
	csvReader = csv.NewReader(strings.NewReader(data))
	csvReader.Comma = comma
	csvReader.Comment = utils.COMMENT_CHAR
	csvReader.FieldsPerRecord = nrFields
	csvReader.TrailingComma = true
	return
}

func (csvr *CSVReader) ShowStatistics() {
	csvr.tp.ShowStatistics()
}

func (csvr *CSVReader) IsDataValid() bool {
	return csvr.tp.IsValid()
}

func (csvr *CSVReader) WriteToDatabase(flush, verbose bool) (err error) {
	return csvr.tp.WriteToDatabase(csvr.dataStorage, csvr.accountingStorage, flush, verbose)
}

func (csvr *CSVReader) LoadDestinations() (err error) {
	csvReader, fp, err := csvr.readerFunc(csvr.destinationsFn, csvr.sep, utils.DESTINATIONS_NRCOLS)
	if err != nil {
		log.Print("Could not load destinations file: ", err)
		// allow writing of the other values
		return nil
	}
	if fp != nil {
		defer fp.Close()
	}
	var tpDests []*TpDestination
	for record, err := csvReader.Read(); err == nil; record, err = csvReader.Read() {
		if tpDest, err := csvLoad(TpDestination{}, record); err != nil {
			return err
		} else {
			tpd := tpDest.(TpDestination)
			tpDests = append(tpDests, &tpd)
		}
		//log.Printf("%+v\n", tpDest)
	}
	csvr.tp.LoadDestinations(tpDests)
	return
}

func (csvr *CSVReader) LoadTimings() (err error) {
	csvReader, fp, err := csvr.readerFunc(csvr.timingsFn, csvr.sep, utils.TIMINGS_NRCOLS)
	if err != nil {
		log.Print("Could not load timings file: ", err)
		// allow writing of the other values
		return nil
	}
	if fp != nil {
		defer fp.Close()
	}
	for record, err := csvReader.Read(); err == nil; record, err = csvReader.Read() {
		tag := record[0]
		if _, exists := csvr.tp.timings[tag]; exists {
			log.Print("Warning: duplicate timing found: ", tag)
		}
		csvr.tp.timings[tag] = NewTiming(record...)
	}
	return
}

func (csvr *CSVReader) LoadRates() (err error) {
	csvReader, fp, err := csvr.readerFunc(csvr.ratesFn, csvr.sep, utils.RATES_NRCOLS)
	if err != nil {
		log.Print("Could not load rates file: ", err)
		// allow writing of the other values
		return nil
	}
	if fp != nil {
		defer fp.Close()
	}

	for record, err := csvReader.Read(); err == nil; record, err = csvReader.Read() {
		tag := record[0]
		var r *utils.TPRate
		r, err = NewLoadRate(record[0], record[1], record[2], record[3], record[4], record[5])
		if err != nil {
			return err
		}
		// same tag only to create rate groups
		_, exists := csvr.tp.rates[tag]
		if exists {
			csvr.tp.rates[tag].RateSlots = append(csvr.tp.rates[tag].RateSlots, r.RateSlots[0])
		} else {
			csvr.tp.rates[tag] = r
		}
	}
	return
}

func (csvr *CSVReader) LoadDestinationRates() (err error) {
	csvReader, fp, err := csvr.readerFunc(csvr.destinationratesFn, csvr.sep, utils.DESTINATION_RATES_NRCOLS)
	if err != nil {
		log.Print("Could not load destination_rates file: ", err)
		// allow writing of the other values
		return nil
	}
	if fp != nil {
		defer fp.Close()
	}
	for record, err := csvReader.Read(); err == nil; record, err = csvReader.Read() {
		tag := record[0]
		r, exists := csvr.tp.rates[record[2]]
		if !exists {
			return fmt.Errorf("Could not get rates for tag %v", record[2])
		}
		roundingDecimals, err := strconv.Atoi(record[4])
		if err != nil {
			log.Printf("Error parsing rounding decimals: %s", record[4])
			return err
		}
		maxCost, err := strconv.ParseFloat(ValueOrDefault(record[5], "0"), 64)
		if err != nil {
			log.Printf("Error parsing max cost from: %v", record[5])
			return err
		}
		destinationExists := record[1] == utils.ANY
		if !destinationExists {
			_, destinationExists = csvr.tp.destinations[record[1]]
		}
		if !destinationExists && csvr.dataStorage != nil {
			if destinationExists, err = csvr.dataStorage.HasData(DESTINATION_PREFIX, record[1]); err != nil {
				return err
			}
		}
		if !destinationExists {
			return fmt.Errorf("Could not get destination for tag %v", record[1])
		}
		dr := &utils.TPDestinationRate{
			DestinationRateId: tag,
			DestinationRates: []*utils.DestinationRate{
				&utils.DestinationRate{
					DestinationId:    record[1],
					Rate:             r,
					RoundingMethod:   record[3],
					RoundingDecimals: roundingDecimals,
					MaxCost:          maxCost,
					MaxCostStrategy:  record[6],
				},
			},
		}
		existingDR, exists := csvr.tp.destinationRates[tag]
		if exists {
			existingDR.DestinationRates = append(existingDR.DestinationRates, dr.DestinationRates[0])
		} else {
			existingDR = dr
		}
		csvr.tp.destinationRates[tag] = existingDR
	}
	return
}

func (csvr *CSVReader) LoadRatingPlans() (err error) {
	csvReader, fp, err := csvr.readerFunc(csvr.destinationratetimingsFn, csvr.sep, utils.DESTRATE_TIMINGS_NRCOLS)
	if err != nil {
		log.Print("Could not load rate timings file: ", err)
		// allow writing of the other values
		return nil
	}
	if fp != nil {
		defer fp.Close()
	}
	for record, err := csvReader.Read(); err == nil; record, err = csvReader.Read() {
		tag := record[0]
		t, exists := csvr.tp.timings[record[2]]
		if !exists {
			return fmt.Errorf("Could not get timing for tag %v", record[2])
		}
		drs, exists := csvr.tp.destinationRates[record[1]]
		if !exists {
			return fmt.Errorf("Could not find destination rate for tag %v", record[1])
		}
		rpl := NewRatingPlan(t, record[3])
		plan, exists := csvr.tp.ratingPlans[tag]
		if !exists {
			plan = &RatingPlan{Id: tag}
			csvr.tp.ratingPlans[tag] = plan
		}
		for _, dr := range drs.DestinationRates {
			plan.AddRateInterval(dr.DestinationId, GetRateInterval(rpl, dr))
		}
	}
	return
}

func (csvr *CSVReader) LoadRatingProfiles() (err error) {
	csvReader, fp, err := csvr.readerFunc(csvr.ratingprofilesFn, csvr.sep, utils.RATE_PROFILES_NRCOLS)
	if err != nil {
		log.Print("Could not load rating profiles file: ", err)
		// allow writing of the other values
		return nil
	}
	if fp != nil {
		defer fp.Close()
	}
	for record, err := csvReader.Read(); err == nil; record, err = csvReader.Read() {
		direction, tenant, tor, subject, fallbacksubject := record[0], record[1], record[2], record[3], record[6]
		at, err := utils.ParseDate(record[4])
		if err != nil {
			return fmt.Errorf("Cannot parse activation time from %v", record[4])
		}
		// extract aliases from subject
		aliases := strings.Split(subject, utils.INFIELD_SEP)
		csvr.tp.dirtyRpAliases = append(csvr.tp.dirtyRpAliases, &TenantRatingSubject{Tenant: tenant, Subject: aliases[0]})
		if len(aliases) > 1 {
			subject = aliases[0]
			for _, alias := range aliases[1:] {
				csvr.tp.rpAliases[utils.RatingSubjectAliasKey(tenant, alias)] = subject
			}
		}
		key := utils.ConcatenatedKey(direction, tenant, tor, subject)
		rp, ok := csvr.tp.ratingProfiles[key]
		if !ok {
			rp = &RatingProfile{Id: key}
			csvr.tp.ratingProfiles[key] = rp
		}
		_, exists := csvr.tp.ratingPlans[record[5]]
		if !exists && csvr.dataStorage != nil {
			if exists, err = csvr.dataStorage.HasData(RATING_PLAN_PREFIX, record[5]); err != nil {
				return err
			}
		}
		if !exists {
			return fmt.Errorf("Could not load rating plans for tag: %v", record[5])
		}
		rpa := &RatingPlanActivation{
			ActivationTime:  at,
			RatingPlanId:    record[5],
			FallbackKeys:    utils.FallbackSubjKeys(direction, tenant, tor, fallbacksubject),
			CdrStatQueueIds: strings.Split(record[7], utils.INFIELD_SEP),
		}
		rp.RatingPlanActivations = append(rp.RatingPlanActivations, rpa)
		csvr.tp.ratingProfiles[rp.Id] = rp
	}
	return
}

func (csvr *CSVReader) LoadSharedGroups() (err error) {
	csvReader, fp, err := csvr.readerFunc(csvr.sharedgroupsFn, csvr.sep, utils.SHARED_GROUPS_NRCOLS)
	if err != nil {
		log.Print("Could not load shared groups file: ", err)
		// allow writing of the other values
		return nil
	}
	if fp != nil {
		defer fp.Close()
	}
	for record, err := csvReader.Read(); err == nil; record, err = csvReader.Read() {
		tag := record[0]
		sg, found := csvr.tp.sharedGroups[tag]
		if found {
			sg.AccountParameters[record[1]] = &SharingParameters{
				Strategy:      record[2],
				RatingSubject: record[3],
			}
		} else {
			sg = &SharedGroup{
				Id: tag,
				AccountParameters: map[string]*SharingParameters{
					record[1]: &SharingParameters{
						Strategy:      record[2],
						RatingSubject: record[3],
					},
				},
			}
		}
		csvr.tp.sharedGroups[tag] = sg
	}
	return
}

func (csvr *CSVReader) LoadLCRs() (err error) {
	csvReader, fp, err := csvr.readerFunc(csvr.lcrFn, csvr.sep, utils.LCRS_NRCOLS)
	if err != nil {
		log.Print("Could not load LCR rules file: ", err)
		// allow writing of the other values
		return nil
	}
	if fp != nil {
		defer fp.Close()
	}
	for record, err := csvReader.Read(); err == nil; record, err = csvReader.Read() {
		direction, tenant, category, account, subject := record[0], record[1], record[2], record[3], record[4]
		id := utils.LCRKey(direction, tenant, category, account, subject)
		lcr, found := csvr.tp.lcrs[id]
		activationTime, err := utils.ParseTimeDetectLayout(record[9])
		if err != nil {
			return fmt.Errorf("Could not parse LCR activation time: %v", err)
		}
		weight, err := strconv.ParseFloat(record[10], 64)
		if err != nil {
			return fmt.Errorf("Could not parse LCR weight: %v", err)
		}
		if !found {
			lcr = &LCR{
				Tenant:    tenant,
				Category:  category,
				Direction: direction,
				Account:   account,
				Subject:   subject,
			}
		}
		var act *LCRActivation
		for _, existingAct := range lcr.Activations {
			if existingAct.ActivationTime.Equal(activationTime) {
				act = existingAct
				break
			}
		}
		if act == nil {
			act = &LCRActivation{
				ActivationTime: activationTime,
			}
			lcr.Activations = append(lcr.Activations, act)
		}
		act.Entries = append(act.Entries, &LCREntry{
			DestinationId:  record[5],
			RPCategory:     record[6],
			Strategy:       record[7],
			StrategyParams: record[8],
			Weight:         weight,
		})
		csvr.tp.lcrs[id] = lcr
	}
	return
}

func (csvr *CSVReader) LoadActions() (err error) {
	csvReader, fp, err := csvr.readerFunc(csvr.actionsFn, csvr.sep, utils.ACTIONS_NRCOLS)
	if err != nil {
		log.Print("Could not load action file: ", err)
		// allow writing of the other values
		return nil
	}
	if fp != nil {
		defer fp.Close()
	}
	for record, err := csvReader.Read(); err == nil; record, err = csvReader.Read() {
		tag := record[ACTSCSVIDX_TAG]
		var units float64
		if len(record[ACTSCSVIDX_UNITS]) == 0 { // Not defined
			units = 0.0
		} else {
			units, err = strconv.ParseFloat(record[ACTSCSVIDX_UNITS], 64)
			if err != nil {
				return fmt.Errorf("Could not parse action units: %v", err)
			}
		}
		var balanceWeight float64
		if len(record[ACTSCSVIDX_BALANCE_WEIGHT]) == 0 { // Not defined
			balanceWeight = 0.0
		} else {
			balanceWeight, err = strconv.ParseFloat(record[ACTSCSVIDX_BALANCE_WEIGHT], 64)
			if err != nil {
				return fmt.Errorf("Could not parse action balance weight: %v", err)
			}
		}
		weight, err := strconv.ParseFloat(record[ACTSCSVIDX_WEIGHT], 64)
		if err != nil {
			return fmt.Errorf("Could not parse action weight: %v", err)
		}
		a := &Action{
			Id:               tag,
			ActionType:       record[ACTSCSVIDX_ACTION],
			BalanceType:      record[ACTSCSVIDX_BALANCE_TYPE],
			Direction:        record[ACTSCSVIDX_DIRECTION],
			Weight:           weight,
			ExpirationString: record[ACTSCSVIDX_EXPIRY_TIME],
			ExtraParameters:  record[ACTSCSVIDX_EXTRA_PARAMS],
			Balance: &Balance{
				Uuid:           utils.GenUUID(),
				Id:             record[ACTSCSVIDX_BALANCE_TAG],
				Value:          units,
				Weight:         balanceWeight,
				DestinationIds: record[ACTSCSVIDX_DESTINATION_TAG],
				TimingIDs:      record[ACTSCSVIDX_TIMING_TAGS],
				RatingSubject:  record[ACTSCSVIDX_RATING_SUBJECT],
				Category:       record[ACTSCSVIDX_CATEGORY],
				SharedGroup:    record[ACTSCSVIDX_SHARED_GROUP],
			},
		}
		// load action timings from tags
		if a.Balance.TimingIDs != "" {
			timingIds := strings.Split(a.Balance.TimingIDs, utils.INFIELD_SEP)
			for _, timingID := range timingIds {
				if timing, found := csvr.tp.timings[timingID]; found {
					a.Balance.Timings = append(a.Balance.Timings, &RITiming{
						Years:     timing.Years,
						Months:    timing.Months,
						MonthDays: timing.MonthDays,
						WeekDays:  timing.WeekDays,
						StartTime: timing.StartTime,
						EndTime:   timing.EndTime,
					})
				} else {
					return fmt.Errorf("Could not find timing: %v", timingID)
				}
			}
		}
		if _, err := utils.ParseDate(a.ExpirationString); err != nil {
			return fmt.Errorf("Could not parse expiration time: %v", err)
		}
		// update Id
		idx := 0
		if previous, ok := csvr.tp.actions[tag]; ok {
			idx = len(previous)
		}
		a.Id = a.Id + strconv.Itoa(idx)
		csvr.tp.actions[tag] = append(csvr.tp.actions[tag], a)
	}
	return
}

func (csvr *CSVReader) LoadActionTimings() (err error) {
	csvReader, fp, err := csvr.readerFunc(csvr.actiontimingsFn, csvr.sep, utils.ACTION_PLANS_NRCOLS)
	if err != nil {
		log.Print("Could not load action plans file: ", err)
		// allow writing of the other values
		return nil
	}
	if fp != nil {
		defer fp.Close()
	}
	for record, err := csvReader.Read(); err == nil; record, err = csvReader.Read() {
		tag := record[0]
		_, exists := csvr.tp.actions[record[1]]
		if !exists {
			return fmt.Errorf("ActionPlan: Could not load the action for tag: %v", record[1])
		}
		t, exists := csvr.tp.timings[record[2]]
		if !exists {
			return fmt.Errorf("ActionPlan: Could not load the timing for tag: %v", record[2])
		}
		weight, err := strconv.ParseFloat(record[3], 64)
		if err != nil {
			return fmt.Errorf("ActionTiming: Could not parse action timing weight: %v", err)
		}
		at := &ActionTiming{
			Uuid:   utils.GenUUID(),
			Id:     record[0],
			Weight: weight,
			Timing: &RateInterval{
				Timing: &RITiming{
					Years:     t.Years,
					Months:    t.Months,
					MonthDays: t.MonthDays,
					WeekDays:  t.WeekDays,
					StartTime: t.StartTime,
				},
			},
			ActionsId: record[1],
		}
		csvr.tp.actionsTimings[tag] = append(csvr.tp.actionsTimings[tag], at)
	}
	return
}

func (csvr *CSVReader) LoadActionTriggers() (err error) {
	csvReader, fp, err := csvr.readerFunc(csvr.actiontriggersFn, csvr.sep, utils.ACTION_TRIGGERS_NRCOLS)
	if err != nil {
		log.Print("Could not load action triggers file: ", err)
		// allow writing of the other values
		return nil
	}
	if fp != nil {
		defer fp.Close()
	}
	for record, err := csvReader.Read(); err == nil; record, err = csvReader.Read() {
		tag := record[ATRIGCSVIDX_TAG]
		value, err := strconv.ParseFloat(record[ATRIGCSVIDX_THRESHOLD_VALUE], 64)
		if err != nil {
			return fmt.Errorf("Could not parse action trigger threshold value (%v): %v", record[ATRIGCSVIDX_THRESHOLD_VALUE], err)
		}
		recurrent, err := strconv.ParseBool(record[ATRIGCSVIDX_RECURRENT])
		if err != nil {
			return fmt.Errorf("Could not parse action trigger recurrent flag (%v): %v", record[ATRIGCSVIDX_RECURRENT], err)
		}

		minSleep, err := time.ParseDuration(record[ATRIGCSVIDX_MIN_SLEEP])
		if err != nil {
			if record[ATRIGCSVIDX_MIN_SLEEP] == "" {
				minSleep = 0
			} else {
				return fmt.Errorf("Could not parse action trigger MinSleep (%v): %v", record[ATRIGCSVIDX_MIN_SLEEP], err)
			}
		}
		balanceWeight, err := strconv.ParseFloat(record[ATRIGCSVIDX_BAL_WEIGHT], 64)
		if record[ATRIGCSVIDX_BAL_WEIGHT] != "" && err != nil {
			return fmt.Errorf("Could not parse action trigger BalanceWeight (%v): %v", record[ATRIGCSVIDX_BAL_WEIGHT], err)
		}
		balanceExp, err := utils.ParseTimeDetectLayout(record[ATRIGCSVIDX_BAL_EXPIRY_TIME])
		if record[ATRIGCSVIDX_BAL_EXPIRY_TIME] != "" && err != nil {
			return fmt.Errorf("Could not parse action trigger BalanceExpirationDate (%v): %v", record[ATRIGCSVIDX_BAL_EXPIRY_TIME], err)
		}
		minQI, err := strconv.Atoi(record[ATRIGCSVIDX_STATS_MIN_QUEUED_ITEMS])
		if record[ATRIGCSVIDX_STATS_MIN_QUEUED_ITEMS] != "" && err != nil {
			return fmt.Errorf("Could not parse action trigger MinQueuedItems (%v): %v", record[ATRIGCSVIDX_STATS_MIN_QUEUED_ITEMS], err)
		}
		weight, err := strconv.ParseFloat(record[ATRIGCSVIDX_WEIGHT], 64)
		if err != nil {
			return fmt.Errorf("Could not parse action trigger weight (%v): %v", record[ATRIGCSVIDX_WEIGHT], err)
		}

		at := &ActionTrigger{
			Id:                    record[ATRIGCSVIDX_UNIQUE_ID],
			ThresholdType:         record[ATRIGCSVIDX_THRESHOLD_TYPE],
			ThresholdValue:        value,
			Recurrent:             recurrent,
			MinSleep:              minSleep,
			BalanceId:             record[ATRIGCSVIDX_BAL_TAG],
			BalanceType:           record[ATRIGCSVIDX_BAL_TYPE],
			BalanceDirection:      record[ATRIGCSVIDX_BAL_DIRECTION],
			BalanceDestinationIds: record[ATRIGCSVIDX_BAL_DESTINATION_TAG],
			BalanceWeight:         balanceWeight,
			BalanceExpirationDate: balanceExp,
			BalanceTimingTags:     record[ATRIGCSVIDX_BAL_TIMING_TAGS],
			BalanceRatingSubject:  record[ATRIGCSVIDX_BAL_RATING_SUBJECT],
			BalanceCategory:       record[ATRIGCSVIDX_BAL_CATEGORY],
			BalanceSharedGroup:    record[ATRIGCSVIDX_BAL_SHARED_GROUP],
			MinQueuedItems:        minQI,
			ActionsId:             record[ATRIGCSVIDX_ACTIONS_TAG],
			Weight:                weight,
		}
		if at.Id == "" {
			at.Id = utils.GenUUID()
		}
		csvr.tp.actionsTriggers[tag] = append(csvr.tp.actionsTriggers[tag], at)
	}
	return
}

func (csvr *CSVReader) LoadAccountActions() (err error) {
	csvReader, fp, err := csvr.readerFunc(csvr.accountactionsFn, csvr.sep, utils.ACCOUNT_ACTIONS_NRCOLS)
	if err != nil {
		log.Print("Could not load account actions file: ", err)
		// allow writing of the other values
		return nil
	}
	if fp != nil {
		defer fp.Close()
	}
	for record, err := csvReader.Read(); err == nil; record, err = csvReader.Read() {
		tenant, account, direction := record[0], record[1], record[2]
		// extract aliases from subject
		aliases := strings.Split(account, utils.INFIELD_SEP)
		csvr.tp.dirtyAccAliases = append(csvr.tp.dirtyAccAliases, &TenantAccount{Tenant: tenant, Account: aliases[0]})
		if len(aliases) > 1 {
			account = aliases[0]
			for _, alias := range aliases[1:] {
				csvr.tp.accAliases[utils.AccountAliasKey(tenant, alias)] = account
			}
		}
		tag := utils.ConcatenatedKey(direction, tenant, account)
		if _, alreadyDefined := csvr.tp.accountActions[tag]; alreadyDefined {
			return fmt.Errorf("Duplicate account action found: %s", tag)
		}
		aTriggers, exists := csvr.tp.actionsTriggers[record[4]]
		if record[4] != "" && !exists {
			// only return error if there was something there for the tag
			return fmt.Errorf("Could not get action triggers for tag %s", record[4])
		}
		ub := &Account{
			Id:             tag,
			ActionTriggers: aTriggers,
		}
		csvr.tp.accountActions[tag] = ub
		aTimings, exists := csvr.tp.actionsTimings[record[3]]
		if !exists {
			log.Printf("Could not get action plan for tag %s", record[3])
			// must not continue here
		}
		for _, at := range aTimings {
			at.AccountIds = append(at.AccountIds, tag)
		}
	}
	return nil
}

func (csvr *CSVReader) LoadDerivedChargers() (err error) {
	csvReader, fp, err := csvr.readerFunc(csvr.derivedChargersFn, csvr.sep, utils.DERIVED_CHARGERS_NRCOLS)
	if err != nil {
		log.Print("Could not load derivedChargers file: ", err)
		// allow writing of the other values
		return nil
	}
	if fp != nil {
		defer fp.Close()
	}
	for record, err := csvReader.Read(); err == nil; record, err = csvReader.Read() {
		if _, err = utils.ParseRSRFields(record[6], utils.INFIELD_SEP); err != nil { // Make sure rules are OK before loading in db
			return err
		}
		tag := utils.DerivedChargersKey(record[0], record[1], record[2], record[3], record[4])
		if _, found := csvr.tp.derivedChargers[tag]; found {
			if csvr.tp.derivedChargers[tag], err = csvr.tp.derivedChargers[tag].Append(&utils.DerivedCharger{
				RunId:                ValueOrDefault(record[5], "*default"),
				RunFilters:           record[6],
				ReqTypeField:         ValueOrDefault(record[7], "*default"),
				DirectionField:       ValueOrDefault(record[8], "*default"),
				TenantField:          ValueOrDefault(record[9], "*default"),
				CategoryField:        ValueOrDefault(record[10], "*default"),
				AccountField:         ValueOrDefault(record[11], "*default"),
				SubjectField:         ValueOrDefault(record[12], "*default"),
				DestinationField:     ValueOrDefault(record[13], "*default"),
				SetupTimeField:       ValueOrDefault(record[14], "*default"),
				AnswerTimeField:      ValueOrDefault(record[15], "*default"),
				UsageField:           ValueOrDefault(record[16], "*default"),
				SupplierField:        ValueOrDefault(record[17], "*default"),
				DisconnectCauseField: ValueOrDefault(record[18], "*default"),
			}); err != nil {
				return err
			}
		} else {
			if record[5] == utils.DEFAULT_RUNID {
				return errors.New("Reserved RunId")
			}
			csvr.tp.derivedChargers[tag] = utils.DerivedChargers{&utils.DerivedCharger{
				RunId:                ValueOrDefault(record[5], "*default"),
				RunFilters:           record[6],
				ReqTypeField:         ValueOrDefault(record[7], "*default"),
				DirectionField:       ValueOrDefault(record[8], "*default"),
				TenantField:          ValueOrDefault(record[9], "*default"),
				CategoryField:        ValueOrDefault(record[10], "*default"),
				AccountField:         ValueOrDefault(record[11], "*default"),
				SubjectField:         ValueOrDefault(record[12], "*default"),
				DestinationField:     ValueOrDefault(record[13], "*default"),
				SetupTimeField:       ValueOrDefault(record[14], "*default"),
				AnswerTimeField:      ValueOrDefault(record[15], "*default"),
				UsageField:           ValueOrDefault(record[16], "*default"),
				SupplierField:        ValueOrDefault(record[17], "*default"),
				DisconnectCauseField: ValueOrDefault(record[18], "*default"),
			}}
		}
	}
	return
}

func (csvr *CSVReader) LoadCdrStats() (err error) {
	csvReader, fp, err := csvr.readerFunc(csvr.cdrStatsFn, csvr.sep, utils.CDR_STATS_NRCOLS)
	if err != nil {
		log.Print("Could not load cdr stats file: ", err)
		// allow writing of the other values
		return nil
	}
	if fp != nil {
		defer fp.Close()
	}
	for record, err := csvReader.Read(); err == nil; record, err = csvReader.Read() {
		tag := record[CDRSTATIDX_TAG]
		var cs *CdrStats
		var exists bool
		if cs, exists = csvr.tp.cdrStats[tag]; !exists {
			cs = &CdrStats{Id: tag}
		}
		triggerTag := record[CDRSTATIDX_ATRIGGER]
		triggers, exists := csvr.tp.actionsTriggers[triggerTag]
		if triggerTag != "" && !exists {
			// only return error if there was something there for the tag
			return fmt.Errorf("Could not get action triggers for cdr stats id %s: %s", cs.Id, triggerTag)
		}
		tpCs := &utils.TPCdrStat{
			QueueLength:         record[CDRSTATIDX_QLENGHT],
			TimeWindow:          record[CDRSTATIDX_TIMEWINDOW],
			Metrics:             record[CDRSTATIDX_METRICS],
			SetupInterval:       record[CDRSTATIDX_SETUPTIME],
			TORs:                record[CDRSTATIDX_TOR],
			CdrHosts:            record[CDRSTATIDX_CDRHOST],
			CdrSources:          record[CDRSTATIDX_CDRSRC],
			ReqTypes:            record[CDRSTATIDX_REQTYPE],
			Directions:          record[CDRSTATIDX_DIRECTION],
			Tenants:             record[CDRSTATIDX_TENANT],
			Categories:          record[CDRSTATIDX_CATEGORY],
			Accounts:            record[CDRSTATIDX_ACCOUNT],
			Subjects:            record[CDRSTATIDX_SUBJECT],
			DestinationPrefixes: record[CDRSTATIDX_DSTPREFIX],
			UsageInterval:       record[CDRSTATIDX_USAGE],
			Suppliers:           record[CDRSTATIDX_SUPPLIER],
			DisconnectCauses:    record[CDRSTATIDX_DISCONNECT_CAUSE],
			MediationRunIds:     record[CDRSTATIDX_MEDRUN],
			RatedAccounts:       record[CDRSTATIDX_RTACCOUNT],
			RatedSubjects:       record[CDRSTATIDX_RTSUBJECT],
			CostInterval:        record[CDRSTATIDX_COST],
			ActionTriggers:      record[CDRSTATIDX_ATRIGGER],
		}
		UpdateCdrStats(cs, triggers, tpCs)
		csvr.tp.cdrStats[tag] = cs
	}
	return
}

// Automated loading
func (csvr *CSVReader) LoadAll() error {
	var err error
	if err = csvr.LoadDestinations(); err != nil {
		return err
	}
	if err = csvr.LoadTimings(); err != nil {
		return err
	}
	if err = csvr.LoadRates(); err != nil {
		return err
	}
	if err = csvr.LoadDestinationRates(); err != nil {
		return err
	}
	if err = csvr.LoadRatingPlans(); err != nil {
		return err
	}
	if err = csvr.LoadRatingProfiles(); err != nil {
		return err
	}
	if err = csvr.LoadSharedGroups(); err != nil {
		return err
	}
	if err = csvr.LoadLCRs(); err != nil {
		return err
	}
	if err = csvr.LoadActions(); err != nil {
		return err
	}
	if err = csvr.LoadActionTimings(); err != nil {
		return err
	}
	if err = csvr.LoadActionTriggers(); err != nil {
		return err
	}
	if err = csvr.LoadAccountActions(); err != nil {
		return err
	}
	if err = csvr.LoadDerivedChargers(); err != nil {
		return err
	}
	if err = csvr.LoadCdrStats(); err != nil {
		return err
	}
	return nil
}

// Returns the identities loaded for a specific category, useful for cache reloads
func (csvr *CSVReader) GetLoadedIds(categ string) ([]string, error) {
	return csvr.tp.GetLoadedIds(categ)
}
