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
	actions           map[string][]*Action
	actionsTimings    map[string][]*ActionTiming
	actionsTriggers   map[string][]*ActionTrigger
	rpAliases         map[string]string
	accAliases        map[string]string
	accountActions    map[string]*Account
	dirtyRpAliases    []*TenantRatingSubject // used to clean aliases that might have changed
	dirtyAccAliases   []*TenantAccount       // used to clean aliases that might have changed
	destinations      map[string]*Destination
	timings           map[string]*utils.TPTiming
	rates             map[string]*utils.TPRate
	destinationRates  map[string]*utils.TPDestinationRate
	ratingPlans       map[string]*RatingPlan
	ratingProfiles    map[string]*RatingProfile
	sharedGroups      map[string]*SharedGroup
	lcrs              map[string]*LCR
	derivedChargers   map[string]utils.DerivedChargers
	cdrStats          map[string]*CdrStats
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
	c.actions = make(map[string][]*Action)
	c.actionsTimings = make(map[string][]*ActionTiming)
	c.actionsTriggers = make(map[string][]*ActionTrigger)
	c.accountActions = make(map[string]*Account)
	c.rates = make(map[string]*utils.TPRate)
	c.destinationRates = make(map[string]*utils.TPDestinationRate)
	c.timings = make(map[string]*utils.TPTiming)
	c.destinations = make(map[string]*Destination)
	c.ratingPlans = make(map[string]*RatingPlan)
	c.ratingProfiles = make(map[string]*RatingProfile)
	c.sharedGroups = make(map[string]*SharedGroup)
	c.lcrs = make(map[string]*LCR)
	c.derivedChargers = make(map[string]utils.DerivedChargers)
	c.cdrStats = make(map[string]*CdrStats)
	c.readerFunc = openFileCSVReader
	c.rpAliases = make(map[string]string)
	c.accAliases = make(map[string]string)
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
	// action plans
	log.Print("Action plans: ", len(csvr.actionsTimings))
	// action trigers
	log.Print("Action trigers: ", len(csvr.actionsTriggers))
	// account actions
	log.Print("Account actions: ", len(csvr.accountActions))
	// derivedChargers
	log.Print("Derived Chargers: ", len(csvr.derivedChargers))
	// lcr rules
	log.Print("LCR rules: ", len(csvr.lcrs))
	// cdr stats
	log.Print("CDR stats: ", len(csvr.cdrStats))
}

func (csvr *CSVReader) WriteToDatabase(flush, verbose bool) (err error) {
	dataStorage := csvr.dataStorage
	accountingStorage := csvr.accountingStorage
	if dataStorage == nil {
		return errors.New("No database connection!")
	}
	if flush {
		dataStorage.Flush("")
	}
	if verbose {
		log.Print("Destinations:")
	}
	for _, d := range csvr.destinations {
		err = dataStorage.SetDestination(d)
		if err != nil {
			return err
		}
		if verbose {
			log.Print("\t", d.Id, " : ", d.Prefixes)
		}
	}
	if verbose {
		log.Print("Rating Plans:")
	}
	for _, rp := range csvr.ratingPlans {
		err = dataStorage.SetRatingPlan(rp)
		if err != nil {
			return err
		}
		if verbose {
			log.Print("\t", rp.Id)
		}
	}
	if verbose {
		log.Print("Rating Profiles:")
	}
	for _, rp := range csvr.ratingProfiles {
		err = dataStorage.SetRatingProfile(rp)
		if err != nil {
			return err
		}
		if verbose {
			log.Print("\t", rp.Id)
		}
	}
	if verbose {
		log.Print("Action Plans:")
	}
	for k, ats := range csvr.actionsTimings {
		err = accountingStorage.SetActionTimings(k, ats)
		if err != nil {
			return err
		}
		if verbose {
			log.Println("\t", k)
		}
	}
	if verbose {
		log.Print("Shared Groups:")
	}
	for k, sg := range csvr.sharedGroups {
		err = accountingStorage.SetSharedGroup(sg)
		if err != nil {
			return err
		}
		if verbose {
			log.Println("\t", k)
		}
	}
	if verbose {
		log.Print("LCR Rules:")
	}
	for k, lcr := range csvr.lcrs {
		err = dataStorage.SetLCR(lcr)
		if err != nil {
			return err
		}
		if verbose {
			log.Println("\t", k)
		}
	}
	if verbose {
		log.Print("Actions:")
	}
	for k, as := range csvr.actions {
		err = accountingStorage.SetActions(k, as)
		if err != nil {
			return err
		}
		if verbose {
			log.Println("\t", k)
		}
	}
	if verbose {
		log.Print("Account Actions:")
	}
	for _, ub := range csvr.accountActions {
		err = accountingStorage.SetAccount(ub)
		if err != nil {
			return err
		}
		if verbose {
			log.Println("\t", ub.Id)
		}
	}
	if verbose {
		log.Print("Rating Profile Aliases:")
	}
	if err := dataStorage.RemoveRpAliases(csvr.dirtyRpAliases); err != nil {
		return err
	}
	for key, alias := range csvr.rpAliases {
		err = dataStorage.SetRpAlias(key, alias)
		if err != nil {
			return err
		}
		if verbose {
			log.Print("\t", key)
		}
	}
	if verbose {
		log.Print("Account Aliases:")
	}
	if err := accountingStorage.RemoveAccAliases(csvr.dirtyAccAliases); err != nil {
		return err
	}
	for key, alias := range csvr.accAliases {
		err = accountingStorage.SetAccAlias(key, alias)
		if err != nil {
			return err
		}
		if verbose {
			log.Print("\t", key)
		}
	}
	if verbose {
		log.Print("Derived Chargers:")
	}
	for key, dcs := range csvr.derivedChargers {
		err = accountingStorage.SetDerivedChargers(key, dcs)
		if err != nil {
			return err
		}
		if verbose {
			log.Print("\t", key)
		}
	}
	if verbose {
		log.Print("CDR Stats Queues:")
	}
	for _, sq := range csvr.cdrStats {
		err = dataStorage.SetCdrStats(sq)
		if err != nil {
			return err
		}
		if verbose {
			log.Print("\t", sq.Id)
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
		var found bool
		if dest, found = csvr.destinations[tag]; !found {
			dest = &Destination{Id: tag}
			csvr.destinations[tag] = dest
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
		r, err = NewLoadRate(record[0], record[1], record[2], record[3], record[4], record[5])
		if err != nil {
			return err
		}
		// same tag only to create rate groups
		existingRates, exists := csvr.rates[tag]
		if exists {
			rss := existingRates.RateSlots
			if err := ValidNextGroup(rss[len(rss)-1], r.RateSlots[0]); err != nil {
				return fmt.Errorf("RatesTag: %s, error: <%s>", tag, err.Error())
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
		log.Print("Could not load destination_rates file: ", err)
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
		roundingDecimals, err := strconv.Atoi(record[4])
		if err != nil {
			log.Printf("Error parsing rounding decimals: %s", record[4])
			return err
		}
		destinationExists := record[1] == utils.ANY
		if !destinationExists {
			_, destinationExists = csvr.destinations[record[1]]
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
		direction, tenant, tor, subject, fallbacksubject := record[0], record[1], record[2], record[3], record[6]
		at, err := utils.ParseDate(record[4])
		if err != nil {
			return fmt.Errorf("Cannot parse activation time from %v", record[4])
		}
		// extract aliases from subject
		aliases := strings.Split(subject, ";")
		csvr.dirtyRpAliases = append(csvr.dirtyRpAliases, &TenantRatingSubject{Tenant: tenant, Subject: aliases[0]})
		if len(aliases) > 1 {
			subject = aliases[0]
			for _, alias := range aliases[1:] {
				csvr.rpAliases[utils.RatingSubjectAliasKey(tenant, alias)] = subject
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
		csvr.sharedGroups[tag] = sg
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
		direction, tenant, customer := record[0], record[1], record[2]
		id := fmt.Sprintf("%s:%s:%s", direction, tenant, customer)
		lcr, found := csvr.lcrs[id]
		activationTime, err := utils.ParseTimeDetectLayout(record[7])
		if err != nil {
			return fmt.Errorf("Could not parse LCR activation time: %v", err)
		}
		weight, err := strconv.ParseFloat(record[8], 64)
		if err != nil {
			return fmt.Errorf("Could not parse LCR weight: %v", err)
		}
		if !found {
			lcr = &LCR{
				Direction: direction,
				Tenant:    tenant,
				Customer:  customer,
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
			DestinationId: record[3],
			Category:      record[4],
			Strategy:      record[5],
			Suppliers:     record[6],
			Weight:        weight,
		})
		csvr.lcrs[id] = lcr
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
			Id:               utils.GenUUID(),
			ActionType:       record[ACTSCSVIDX_ACTION],
			BalanceType:      record[ACTSCSVIDX_BALANCE_TYPE],
			Direction:        record[ACTSCSVIDX_DIRECTION],
			Weight:           weight,
			ExpirationString: record[ACTSCSVIDX_EXPIRY_TIME],
			ExtraParameters:  record[ACTSCSVIDX_EXTRA_PARAMS],
			Balance: &Balance{
				Uuid:          utils.GenUUID(),
				Value:         units,
				Weight:        balanceWeight,
				DestinationId: record[ACTSCSVIDX_DESTINATION_TAG],
				RatingSubject: record[ACTSCSVIDX_RATING_SUBJECT],
				Category:      record[ACTSCSVIDX_CATEGORY],
				SharedGroup:   record[ACTSCSVIDX_SHARED_GROUP],
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
			Id:                    utils.GenUUID(),
			BalanceType:           record[ATRIGCSVIDX_BAL_TYPE],
			Direction:             record[ATRIGCSVIDX_BAL_DIRECTION],
			ThresholdType:         record[ATRIGCSVIDX_THRESHOLD_TYPE],
			ThresholdValue:        value,
			Recurrent:             recurrent,
			MinSleep:              minSleep,
			DestinationId:         record[ATRIGCSVIDX_BAL_DESTINATION_TAG],
			BalanceWeight:         balanceWeight,
			BalanceExpirationDate: balanceExp,
			BalanceRatingSubject:  record[ATRIGCSVIDX_BAL_RATING_SUBJECT],
			BalanceCategory:       record[ATRIGCSVIDX_BAL_CATEGORY],
			BalanceSharedGroup:    record[ATRIGCSVIDX_BAL_SHARED_GROUP],
			MinQueuedItems:        minQI,
			ActionsId:             record[ATRIGCSVIDX_ACTIONS_TAG],
			Weight:                weight,
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
		// extract aliases from subject
		aliases := strings.Split(account, ";")
		csvr.dirtyAccAliases = append(csvr.dirtyAccAliases, &TenantAccount{Tenant: tenant, Account: aliases[0]})
		if len(aliases) > 1 {
			account = aliases[0]
			for _, alias := range aliases[1:] {
				csvr.accAliases[utils.AccountAliasKey(tenant, alias)] = account
			}
		}
		tag := fmt.Sprintf("%s:%s:%s", direction, tenant, account)
		if _, alreadyDefined := csvr.accountActions[tag]; alreadyDefined {
			return fmt.Errorf("Duplicate account action found: %s", tag)
		}
		aTriggers, exists := csvr.actionsTriggers[record[4]]
		if record[4] != "" && !exists {
			// only return error if there was something there for the tag
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
		if _, found := csvr.derivedChargers[tag]; found {
			if csvr.derivedChargers[tag], err = csvr.derivedChargers[tag].Append(&utils.DerivedCharger{
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
			}); err != nil {
				return err
			}
		} else {
			if record[5] == utils.DEFAULT_RUNID {
				return errors.New("Reserved RunId")
			}
			csvr.derivedChargers[tag] = utils.DerivedChargers{&utils.DerivedCharger{
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
		tag := record[0]
		var cs *CdrStats
		var exists bool
		if cs, exists = csvr.cdrStats[tag]; !exists {
			cs = &CdrStats{Id: tag}
		}
		triggerTag := record[20]
		triggers, exists := csvr.actionsTriggers[triggerTag]
		if triggerTag != "" && !exists {
			// only return error if there was something there for the tag
			return fmt.Errorf("Could not get action triggers for cdr stats id %s: %s", cs.Id, triggerTag)
		}
		tpCs := &utils.TPCdrStat{
			QueueLength:       record[1],
			TimeWindow:        record[2],
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
		}
		UpdateCdrStats(cs, triggers, tpCs)
		csvr.cdrStats[tag] = cs
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
	switch categ {
	case DESTINATION_PREFIX:
		ids := make([]string, len(csvr.destinations))
		i := 0
		for k := range csvr.destinations {
			ids[i] = k
			i++
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
	case RP_ALIAS_PREFIX: // aliases
		keys := make([]string, len(csvr.rpAliases))
		i := 0
		for k := range csvr.rpAliases {
			keys[i] = k
			i++
		}
		return keys, nil
	case ACC_ALIAS_PREFIX: // aliases
		keys := make([]string, len(csvr.accAliases))
		i := 0
		for k := range csvr.accAliases {
			keys[i] = k
			i++
		}
		return keys, nil
	case DERIVEDCHARGERS_PREFIX: // derived chargers
		keys := make([]string, len(csvr.derivedChargers))
		i := 0
		for k := range csvr.derivedChargers {
			keys[i] = k
			i++
		}
		return keys, nil
	case CDR_STATS_PREFIX: // cdr stats
		keys := make([]string, len(csvr.cdrStats))
		i := 0
		for k := range csvr.cdrStats {
			keys[i] = k
			i++
		}
		return keys, nil
	}
	return nil, errors.New("Unsupported category")
}
