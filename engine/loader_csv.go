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
	"encoding/csv"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/cgrates/cgrates/utils"
)

type CSVReader struct {
	sep               rune
	dataStorage       RatingStorage
	accountingStorage AccountingStorage
	readerFunc        func(string, rune, int) (*csv.Reader, *os.File, error)
	actions           map[string][]*Action
	actionsTimings    map[string][]*ActionTiming
	actionsTriggers   map[string][]*ActionTrigger
	aliases           map[string]string
	accountActions    map[string]*Account
	dirtyAliases      []string // used to clean aliases that might have changed
	destinations      []*Destination
	timings           map[string]*utils.TPTiming
	rates             map[string]*utils.TPRate
	destinationRates  map[string]*utils.TPDestinationRate
	ratingPlans       map[string]*RatingPlan
	ratingProfiles    map[string]*RatingProfile
	sharedGroups      map[string]*SharedGroup
	// file names
	destinationsFn, ratesFn, destinationratesFn, timingsFn, destinationratetimingsFn, ratingprofilesFn,
	sharedgroupsFn, actionsFn, actiontimingsFn, actiontriggersFn, accountactionsFn string
}

func NewFileCSVReader(dataStorage RatingStorage, accountingStorage AccountingStorage, sep rune, destinationsFn, timingsFn, ratesFn, destinationratesFn, destinationratetimingsFn, ratingprofilesFn, sharedgroupsFn, actionsFn, actiontimingsFn, actiontriggersFn, accountactionsFn string) *CSVReader {
	c := new(CSVReader)
	c.sep = sep
	c.dataStorage = dataStorage
	c.accountingStorage = accountingStorage
	c.actions = make(map[string][]*Action)
	c.actionsTimings = make(map[string][]*ActionTiming)
	c.actionsTriggers = make(map[string][]*ActionTrigger)
	c.accountActions = make(map[string]*Account)
	c.rates = make(map[string]*utils.TPRate)
	c.destinationRates = make(map[string]*utils.TPDestinationRate)
	c.timings = make(map[string]*utils.TPTiming)
	c.ratingPlans = make(map[string]*RatingPlan)
	c.ratingProfiles = make(map[string]*RatingProfile)
	c.sharedGroups = make(map[string]*SharedGroup)
	c.readerFunc = openFileCSVReader
	c.aliases = make(map[string]string)
	c.destinationsFn, c.timingsFn, c.ratesFn, c.destinationratesFn, c.destinationratetimingsFn, c.ratingprofilesFn,
		c.sharedgroupsFn, c.actionsFn, c.actiontimingsFn, c.actiontriggersFn, c.accountactionsFn = destinationsFn, timingsFn,
		ratesFn, destinationratesFn, destinationratetimingsFn, ratingprofilesFn, sharedgroupsFn, actionsFn, actiontimingsFn, actiontriggersFn, accountactionsFn
	return c
}

func NewStringCSVReader(dataStorage RatingStorage, accountingStorage AccountingStorage, sep rune, destinationsFn, timingsFn, ratesFn, destinationratesFn, destinationratetimingsFn, ratingprofilesFn, sharedgroupsFn, actionsFn, actiontimingsFn, actiontriggersFn, accountactionsFn string) *CSVReader {
	c := NewFileCSVReader(dataStorage, accountingStorage, sep, destinationsFn, timingsFn, ratesFn, destinationratesFn, destinationratetimingsFn, ratingprofilesFn, sharedgroupsFn, actionsFn, actiontimingsFn, actiontriggersFn, accountactionsFn)
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
	// destinations
	destCount := len(csvr.destinations)
	log.Print("Destinations: ", destCount)
	prefixDist := make(map[int]int, 50)
	prefixCount := 0
	for _, d := range csvr.destinations {
		prefixDist[len(d.Prefixes)] += 1
		prefixCount += len(d.Prefixes)
	}
	log.Print("Avg Prefixes: ", prefixCount/destCount)
	log.Print("Prefixes distribution:")
	for k, v := range prefixDist {
		log.Printf("%d: %d", k, v)
	}
	// rating plans
	rplCount := len(csvr.ratingPlans)
	log.Print("Rating plans: ", rplCount)
	destRatesDist := make(map[int]int, 50)
	destRatesCount := 0
	for _, rpl := range csvr.ratingPlans {
		destRatesDist[len(rpl.DestinationRates)] += 1
		destRatesCount += len(rpl.DestinationRates)
	}
	log.Print("Avg Destination Rates: ", destRatesCount/rplCount)
	log.Print("Destination Rates distribution:")
	for k, v := range destRatesDist {
		log.Printf("%d: %d", k, v)
	}
	// rating profiles
	rpfCount := len(csvr.ratingProfiles)
	log.Print("Rating profiles: ", rpfCount)
	activDist := make(map[int]int, 50)
	activCount := 0
	for _, rpf := range csvr.ratingProfiles {
		activDist[len(rpf.RatingPlanActivations)] += 1
		activCount += len(rpf.RatingPlanActivations)
	}
	log.Print("Avg Activations: ", activCount/rpfCount)
	log.Print("Activation distribution:")
	for k, v := range activDist {
		log.Printf("%d: %d", k, v)
	}
	// actions
	log.Print("Actions: ", len(csvr.actions))
	// action timings
	log.Print("Action plans: ", len(csvr.actionsTimings))
	// account actions
	log.Print("Account actions: ", len(csvr.accountActions))
}

func (csvr *CSVReader) WriteToDatabase(flush, verbose bool) (err error) {
	dataStorage := csvr.dataStorage
	accountingStorage := csvr.accountingStorage
	if dataStorage == nil {
		return errors.New("No database connection!")
	}
	if flush {
		dataStorage.(Storage).Flush()
	}
	if verbose {
		log.Print("Destinations")
	}
	for _, d := range csvr.destinations {
		err = dataStorage.SetDestination(d)
		if err != nil {
			return err
		}
		if verbose {
			log.Print(d.Id, " : ", d.Prefixes)
		}
	}
	if verbose {
		log.Print("Rating plans")
	}
	for _, rp := range csvr.ratingPlans {
		err = dataStorage.SetRatingPlan(rp)
		if err != nil {
			return err
		}
		if verbose {
			log.Print(rp.Id)
		}
	}
	if verbose {
		log.Print("Rating profiles")
	}
	for _, rp := range csvr.ratingProfiles {
		err = dataStorage.SetRatingProfile(rp)
		if err != nil {
			return err
		}
		if verbose {
			log.Print(rp.Id)
		}
	}
	if verbose {
		log.Print("Action plans")
	}
	for k, ats := range csvr.actionsTimings {
		err = accountingStorage.SetActionTimings(k, ats)
		if err != nil {
			return err
		}
		if verbose {
			log.Println(k)
		}
	}
	if verbose {
		log.Print("Shared groups")
	}
	for k, sg := range csvr.sharedGroups {
		err = accountingStorage.SetSharedGroup(k, sg)
		if err != nil {
			return err
		}
		if verbose {
			log.Println(k)
		}
	}
	if verbose {
		log.Print("Actions")
	}
	for k, as := range csvr.actions {
		err = accountingStorage.SetActions(k, as)
		if err != nil {
			return err
		}
		if verbose {
			log.Println(k)
		}
	}
	if verbose {
		log.Print("Account actions")
	}
	for _, ub := range csvr.accountActions {
		err = accountingStorage.SetAccount(ub)
		if err != nil {
			return err
		}
		if verbose {
			log.Println(ub.Id)
		}
	}
	if verbose {
		log.Print("Aliases")
	}
	if err := dataStorage.RemoveAccountAliases(csvr.dirtyAliases); err != nil {
		return err
	}
	for key, alias := range csvr.aliases {
		err = dataStorage.SetAlias(key, alias)
		if err != nil {
			return err
		}
		if verbose {
			log.Print(key)
		}
	}
	return
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
	for record, err := csvReader.Read(); err == nil; record, err = csvReader.Read() {
		tag := record[0]
		var dest *Destination
		for _, d := range csvr.destinations {
			if d.Id == tag {
				dest = d
				break
			}
		}
		if dest == nil {
			dest = &Destination{Id: tag}
			csvr.destinations = append(csvr.destinations, dest)
		}
		dest.AddPrefix(record[1])
	}
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
		if _, exists := csvr.timings[tag]; exists {
			log.Print("Warning: duplicate timing found: ", tag)
		}
		csvr.timings[tag] = NewTiming(record...)
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
		r, err = NewLoadRate(record[0], record[1], record[2], record[3], record[4], record[5], record[6], record[7])
		if err != nil {
			return err
		}
		// same tag only to create rate groups
		existingRates, exists := csvr.rates[tag]
		if exists {
			rss := existingRates.RateSlots
			if err := ValidNextGroup(rss[len(rss)-1], r.RateSlots[0]); err != nil {
				return err
			}
			csvr.rates[tag].RateSlots = append(csvr.rates[tag].RateSlots, r.RateSlots[0])
		} else {
			csvr.rates[tag] = r
		}
	}
	return
}

func (csvr *CSVReader) LoadDestinationRates() (err error) {
	csvReader, fp, err := csvr.readerFunc(csvr.destinationratesFn, csvr.sep, utils.DESTINATION_RATES_NRCOLS)
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
		r, exists := csvr.rates[record[2]]
		if !exists {
			return fmt.Errorf("Could not get rates for tag %v", record[2])
		}
		destinationExists := false
		for _, d := range csvr.destinations {
			if d.Id == record[1] {
				destinationExists = true
				break
			}
		}
		var err error
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
					DestinationId: record[1],
					Rate:          r,
				},
			},
		}
		existingDR, exists := csvr.destinationRates[tag]
		if exists {
			existingDR.DestinationRates = append(existingDR.DestinationRates, dr.DestinationRates[0])
		} else {
			existingDR = dr
		}
		csvr.destinationRates[tag] = existingDR
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
		t, exists := csvr.timings[record[2]]
		if !exists {
			return fmt.Errorf("Could not get timing for tag %v", record[2])
		}
		drs, exists := csvr.destinationRates[record[1]]
		if !exists {
			return fmt.Errorf("Could not find destination rate for tag %v", record[1])
		}
		rpl := NewRatingPlan(t, record[3])
		plan, exists := csvr.ratingPlans[tag]
		if !exists {
			plan = &RatingPlan{Id: tag}
			csvr.ratingPlans[tag] = plan
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
		tenant, tor, direction, subject, fallbacksubject := record[0], record[1], record[2], record[3], record[6]
		at, err := utils.ParseDate(record[4])
		if err != nil {
			return fmt.Errorf("Cannot parse activation time from %v", record[4])
		}
		csvr.dirtyAliases = append(csvr.dirtyAliases, subject)
		// extract aliases from subject
		aliases := strings.Split(subject, ";")
		if len(aliases) > 1 {
			subject = aliases[0]
			for _, alias := range aliases[1:] {
				csvr.aliases[RATING_PROFILE_PREFIX+alias] = subject
			}
		}
		key := fmt.Sprintf("%s:%s:%s:%s", direction, tenant, tor, subject)
		rp, ok := csvr.ratingProfiles[key]
		if !ok {
			rp = &RatingProfile{Id: key}
			csvr.ratingProfiles[key] = rp
		}
		_, exists := csvr.ratingPlans[record[5]]
		if !exists && csvr.dataStorage != nil {
			if exists, err = csvr.dataStorage.HasData(RATING_PLAN_PREFIX, record[5]); err != nil {
				return err
			}
		}
		if !exists {
			return fmt.Errorf("Could not load rating plans for tag: %v", record[5])
		}
		rpa := &RatingPlanActivation{
			ActivationTime: at,
			RatingPlanId:   record[5],
			FallbackKeys:   utils.FallbackSubjKeys(direction, tenant, tor, fallbacksubject),
		}
		rp.RatingPlanActivations = append(rp.RatingPlanActivations, rpa)
		csvr.ratingProfiles[rp.Id] = rp
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
		sg, found := csvr.sharedGroups[tag]
		if found {
			sg.AccountParameters[record[1]] = &SharingParameters{
				Strategy:    record[2],
				RateSubject: record[3],
			}
		} else {
			sg = &SharedGroup{
				Id: tag,
				AccountParameters: map[string]*SharingParameters{
					record[1]: &SharingParameters{
						Strategy:    record[2],
						RateSubject: record[3],
					},
				},
			}
		}
		csvr.sharedGroups[tag] = sg
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
		tag := record[0]
		var units float64
		if len(record[4]) == 0 { // Not defined
			units = 0.0
		} else {
			units, err = strconv.ParseFloat(record[4], 64)
			if err != nil {
				return fmt.Errorf("Could not parse action units: %v", err)
			}
		}
		var balanceWeight float64
		if len(record[8]) == 0 { // Not defined
			balanceWeight = 0.0
		} else {
			balanceWeight, err = strconv.ParseFloat(record[8], 64)
			if err != nil {
				return fmt.Errorf("Could not parse action balance weight: %v", err)
			}
		}
		weight, err := strconv.ParseFloat(record[11], 64)
		if err != nil {
			return fmt.Errorf("Could not parse action weight: %v", err)
		}
		a := &Action{
			Id:               utils.GenUUID(),
			ActionType:       record[1],
			BalanceType:      record[2],
			Direction:        record[3],
			Weight:           weight,
			ExpirationString: record[5],
			ExtraParameters:  record[10],
			Balance: &Balance{
				Uuid:          utils.GenUUID(),
				Value:         units,
				Weight:        balanceWeight,
				DestinationId: record[6],
				RateSubject:   record[7],
				SharedGroup:   record[9],
			},
		}
		if _, err := utils.ParseDate(a.ExpirationString); err != nil {
			return fmt.Errorf("Could not parse expiration time: %v", err)
		}
		csvr.actions[tag] = append(csvr.actions[tag], a)
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
		_, exists := csvr.actions[record[1]]
		if !exists {
			return fmt.Errorf("ActionPlan: Could not load the action for tag: %v", record[1])
		}
		t, exists := csvr.timings[record[2]]
		if !exists {
			return fmt.Errorf("ActionPlan: Could not load the timing for tag: %v", record[2])
		}
		weight, err := strconv.ParseFloat(record[3], 64)
		if err != nil {
			return fmt.Errorf("ActionTiming: Could not parse action timing weight: %v", err)
		}
		at := &ActionTiming{
			Id:     utils.GenUUID(),
			Tag:    record[0],
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
		csvr.actionsTimings[tag] = append(csvr.actionsTimings[tag], at)
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
		tag := record[0]
		value, err := strconv.ParseFloat(record[4], 64)
		if err != nil {
			return fmt.Errorf("Could not parse action trigger value: %v", err)
		}
		weight, err := strconv.ParseFloat(record[7], 64)
		if err != nil {
			return fmt.Errorf("Could not parse action trigger weight: %v", err)
		}
		at := &ActionTrigger{
			Id:             utils.GenUUID(),
			BalanceType:    record[1],
			Direction:      record[2],
			ThresholdType:  record[3],
			ThresholdValue: value,
			DestinationId:  record[5],
			ActionsId:      record[6],
			Weight:         weight,
		}
		csvr.actionsTriggers[tag] = append(csvr.actionsTriggers[tag], at)
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
		csvr.dirtyAliases = append(csvr.dirtyAliases, account)
		// extract aliases from subject
		aliases := strings.Split(account, ";")
		if len(aliases) > 1 {
			account = aliases[0]
			for _, alias := range aliases[1:] {
				csvr.aliases[ACCOUNT_PREFIX+alias] = account
			}
		}
		tag := fmt.Sprintf("%s:%s:%s", direction, tenant, account)
		if _, alreadyDefined := csvr.accountActions[tag]; alreadyDefined {
			return fmt.Errorf("Duplicate account action found: %s", tag)
		}
		aTriggers, exists := csvr.actionsTriggers[record[4]]
		if record[4] != "" && !exists {
			// only return error if there was something ther for the tag
			return fmt.Errorf("Could not get action triggers for tag %s", record[4])
		}
		ub := &Account{
			Id:             tag,
			ActionTriggers: aTriggers,
		}
		csvr.accountActions[tag] = ub
		aTimings, exists := csvr.actionsTimings[record[3]]
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
	return nil
}

// Returns the identities loaded for a specific category, useful for cache reloads
func (csvr *CSVReader) GetLoadedIds(categ string) ([]string, error) {
	switch categ {
	case DESTINATION_PREFIX:
		ids := make([]string, len(csvr.destinations))
		for idx, dst := range csvr.destinations {
			ids[idx] = dst.Id
		}
		return ids, nil
	case RATING_PLAN_PREFIX:
		keys := make([]string, len(csvr.ratingPlans))
		i := 0
		for k := range csvr.ratingPlans {
			keys[i] = k
			i++
		}
		return keys, nil
	case RATING_PROFILE_PREFIX:
		keys := make([]string, len(csvr.ratingProfiles))
		i := 0
		for k := range csvr.ratingProfiles {
			keys[i] = k
			i++
		}
		return keys, nil
	case ACTION_PREFIX: // actionsTimings
		keys := make([]string, len(csvr.actions))
		i := 0
		for k := range csvr.actions {
			keys[i] = k
			i++
		}
		return keys, nil
	case ACTION_TIMING_PREFIX: // actionsTimings
		keys := make([]string, len(csvr.actionsTimings))
		i := 0
		for k := range csvr.actionsTimings {
			keys[i] = k
			i++
		}
		return keys, nil
	case ALIAS_PREFIX: // aliases
		keys := make([]string, len(csvr.aliases))
		i := 0
		for k := range csvr.aliases {
			keys[i] = k
			i++
		}
		return keys, nil
	}
	return nil, errors.New("Unsupported category")
}
