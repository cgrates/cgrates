package engine

import (
	"encoding/csv"
	"errors"
	"io"
	"log"
	"os"
	"strings"

	"github.com/cgrates/cgrates/utils"
)

type CSVStorage struct {
	sep        rune
	readerFunc func(string, rune, int) (*csv.Reader, *os.File, error)
	// file names
	destinationsFn, ratesFn, destinationratesFn, timingsFn, destinationratetimingsFn, ratingprofilesFn,
	sharedgroupsFn, lcrFn, actionsFn, actiontimingsFn, actiontriggersFn, accountactionsFn, derivedChargersFn, cdrStatsFn string
}

func NewFileCSVStorage(sep rune,
	destinationsFn, timingsFn, ratesFn, destinationratesFn, destinationratetimingsFn, ratingprofilesFn, sharedgroupsFn, lcrFn,
	actionsFn, actiontimingsFn, actiontriggersFn, accountactionsFn, derivedChargersFn, cdrStatsFn string) *CSVStorage {
	c := new(CSVStorage)
	c.sep = sep
	c.readerFunc = openFileCSVStorage
	c.destinationsFn, c.timingsFn, c.ratesFn, c.destinationratesFn, c.destinationratetimingsFn, c.ratingprofilesFn,
		c.sharedgroupsFn, c.lcrFn, c.actionsFn, c.actiontimingsFn, c.actiontriggersFn, c.accountactionsFn, c.derivedChargersFn, c.cdrStatsFn = destinationsFn, timingsFn,
		ratesFn, destinationratesFn, destinationratetimingsFn, ratingprofilesFn, sharedgroupsFn, lcrFn, actionsFn, actiontimingsFn, actiontriggersFn, accountactionsFn, derivedChargersFn, cdrStatsFn
	return c
}

func NewStringCSVStorage(sep rune,
	destinationsFn, timingsFn, ratesFn, destinationratesFn, destinationratetimingsFn, ratingprofilesFn, sharedgroupsFn, lcrFn,
	actionsFn, actiontimingsFn, actiontriggersFn, accountactionsFn, derivedChargersFn, cdrStatsFn string) *CSVStorage {
	c := NewFileCSVStorage(sep, destinationsFn, timingsFn, ratesFn, destinationratesFn, destinationratetimingsFn,
		ratingprofilesFn, sharedgroupsFn, lcrFn, actionsFn, actiontimingsFn, actiontriggersFn, accountactionsFn, derivedChargersFn, cdrStatsFn)
	c.readerFunc = openStringCSVStorage
	return c
}

func openFileCSVStorage(fn string, comma rune, nrFields int) (csvReader *csv.Reader, fp *os.File, err error) {
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

func openStringCSVStorage(data string, comma rune, nrFields int) (csvReader *csv.Reader, fp *os.File, err error) {
	csvReader = csv.NewReader(strings.NewReader(data))
	csvReader.Comma = comma
	csvReader.Comment = utils.COMMENT_CHAR
	csvReader.FieldsPerRecord = nrFields
	csvReader.TrailingComma = true
	return
}

func (csvs *CSVStorage) GetTpTimings(string, string) ([]TpTiming, error) {
	csvReader, fp, err := csvs.readerFunc(csvs.timingsFn, csvs.sep, getColumnCount(TpTiming{}))
	if err != nil {
		log.Print("Could not load timings file: ", err)
		// allow writing of the other values
		return nil, nil
	}
	if fp != nil {
		defer fp.Close()
	}
	var tpTimings []TpTiming
	for record, err := csvReader.Read(); err != io.EOF; record, err = csvReader.Read() {
		if err != nil {
			log.Print("bad line in timings csv: ", err)
			return nil, err
		}
		if tpTiming, err := csvLoad(TpTiming{}, record); err != nil {
			return nil, err
		} else {
			tpTimings = append(tpTimings, tpTiming.(TpTiming))
		}
	}
	return tpTimings, nil
}

func (csvs *CSVStorage) GetTpDestinations(tpid, tag string) ([]TpDestination, error) {
	csvReader, fp, err := csvs.readerFunc(csvs.destinationsFn, csvs.sep, getColumnCount(TpDestination{}))
	if err != nil {
		log.Print("Could not load destinations file: ", err)
		// allow writing of the other values
		return nil, nil
	}
	if fp != nil {
		defer fp.Close()
	}
	var tpDests []TpDestination
	for record, err := csvReader.Read(); err != io.EOF; record, err = csvReader.Read() {
		if err != nil {
			log.Print("bad line in destinations csv: ", err)
			return nil, err
		}
		if tpDest, err := csvLoad(TpDestination{}, record); err != nil {
			return nil, err
		} else {
			tpDests = append(tpDests, tpDest.(TpDestination))
		}
	}
	return tpDests, nil
}

func (csvs *CSVStorage) GetTpRates(tpid, tag string) ([]TpRate, error) {
	csvReader, fp, err := csvs.readerFunc(csvs.ratesFn, csvs.sep, getColumnCount(TpRate{}))
	if err != nil {
		log.Print("Could not load rates file: ", err)
		// allow writing of the other values
		return nil, nil
	}
	if fp != nil {
		defer fp.Close()
	}
	var tpRates []TpRate
	for record, err := csvReader.Read(); err != io.EOF; record, err = csvReader.Read() {
		if err != nil {
			log.Print("bad line in rates csv: ", err)
			return nil, err
		}
		if tpRate, err := csvLoad(TpRate{}, record); err != nil {
			return nil, err
		} else {
			tpRates = append(tpRates, tpRate.(TpRate))
		}
	}
	return tpRates, nil
}

func (csvs *CSVStorage) GetTpDestinationRates(tpid, tag string, p *utils.Paginator) ([]TpDestinationRate, error) {
	csvReader, fp, err := csvs.readerFunc(csvs.destinationratesFn, csvs.sep, getColumnCount(TpDestinationRate{}))
	if err != nil {
		log.Print("Could not load destination_rates file: ", err)
		// allow writing of the other values
		return nil, nil
	}
	if fp != nil {
		defer fp.Close()
	}
	var tpDestinationRates []TpDestinationRate
	for record, err := csvReader.Read(); err != io.EOF; record, err = csvReader.Read() {
		if err != nil {
			log.Print("bad line in destinationrates csv: ", err)
			return nil, err
		}
		if tpRate, err := csvLoad(TpDestinationRate{}, record); err != nil {
			return nil, err
		} else {
			tpDestinationRates = append(tpDestinationRates, tpRate.(TpDestinationRate))
		}
	}
	return tpDestinationRates, nil
}

func (csvs *CSVStorage) GetTpRatingPlans(tpid, tag string, p *utils.Paginator) ([]TpRatingPlan, error) {
	csvReader, fp, err := csvs.readerFunc(csvs.destinationratetimingsFn, csvs.sep, getColumnCount(TpRatingPlan{}))
	if err != nil {
		log.Print("Could not load rate plans file: ", err)
		// allow writing of the other values
		return nil, nil
	}
	if fp != nil {
		defer fp.Close()
	}
	var tpRatingPlans []TpRatingPlan
	for record, err := csvReader.Read(); err != io.EOF; record, err = csvReader.Read() {
		if err != nil {
			log.Print("bad line in rating plans csv: ", err)
			return nil, err
		}
		if tpRate, err := csvLoad(TpRatingPlan{}, record); err != nil {
			return nil, err
		} else {
			tpRatingPlans = append(tpRatingPlans, tpRate.(TpRatingPlan))
		}
	}
	return tpRatingPlans, nil
}

func (csvs *CSVStorage) GetTpRatingProfiles(filter *TpRatingProfile) ([]TpRatingProfile, error) {
	csvReader, fp, err := csvs.readerFunc(csvs.ratingprofilesFn, csvs.sep, getColumnCount(TpRatingProfile{}))
	if err != nil {
		log.Print("Could not load rating profiles file: ", err)
		// allow writing of the other values
		return nil, nil
	}
	if fp != nil {
		defer fp.Close()
	}
	var tpRatingProfiles []TpRatingProfile
	for record, err := csvReader.Read(); err != io.EOF; record, err = csvReader.Read() {
		if err != nil {
			log.Print("bad line rating profiles csv: ", err)
			return nil, err
		}
		if tpRate, err := csvLoad(TpRatingProfile{}, record); err != nil {
			return nil, err
		} else {
			tpRatingProfiles = append(tpRatingProfiles, tpRate.(TpRatingProfile))
		}
	}
	return tpRatingProfiles, nil
}

func (csvs *CSVStorage) GetTpSharedGroups(tpid, tag string) ([]TpSharedGroup, error) {
	csvReader, fp, err := csvs.readerFunc(csvs.sharedgroupsFn, csvs.sep, getColumnCount(TpSharedGroup{}))
	if err != nil {
		log.Print("Could not load shared groups file: ", err)
		// allow writing of the other values
		return nil, nil
	}
	if fp != nil {
		defer fp.Close()
	}

	var tpSharedGroups []TpSharedGroup
	for record, err := csvReader.Read(); err != io.EOF; record, err = csvReader.Read() {
		if err != nil {
			log.Print("bad line in shared groups csv: ", err)
			return nil, err
		}
		if tpRate, err := csvLoad(TpSharedGroup{}, record); err != nil {
			return nil, err
		} else {
			tpSharedGroups = append(tpSharedGroups, tpRate.(TpSharedGroup))
		}
	}
	return tpSharedGroups, nil
}

func (csvs *CSVStorage) GetTpLCRs(tpid, tag string) ([]TpLcrRule, error) {
	csvReader, fp, err := csvs.readerFunc(csvs.lcrFn, csvs.sep, getColumnCount(TpLcrRule{}))
	if err != nil {
		log.Print("Could not load LCR rules file: ", err)
		// allow writing of the other values
		return nil, nil
	}
	if fp != nil {
		defer fp.Close()
	}
	var tpLCRs []TpLcrRule
	for record, err := csvReader.Read(); err != io.EOF; record, err = csvReader.Read() {
		if tpRate, err := csvLoad(TpLcrRule{}, record); err != nil {
			if err != nil {
				log.Print("bad line in lcr rules csv: ", err)
				return nil, err
			}
			return nil, err
		} else {
			tpLCRs = append(tpLCRs, tpRate.(TpLcrRule))
		}
	}
	return tpLCRs, nil
}

func (csvs *CSVStorage) GetTpActions(tpid, tag string) ([]TpAction, error) {
	csvReader, fp, err := csvs.readerFunc(csvs.actionsFn, csvs.sep, getColumnCount(TpAction{}))
	if err != nil {
		log.Print("Could not load action file: ", err)
		// allow writing of the other values
		return nil, nil
	}
	if fp != nil {
		defer fp.Close()
	}
	var tpActions []TpAction
	for record, err := csvReader.Read(); err != io.EOF; record, err = csvReader.Read() {
		if err != nil {
			log.Print("bad line in actions csv: ", err)
			return nil, err
		}
		if tpAction, err := csvLoad(TpAction{}, record); err != nil {
			log.Print("error loading action: ", err)
			return nil, err
		} else {
			tpActions = append(tpActions, tpAction.(TpAction))
		}
	}
	return tpActions, nil
}

func (csvs *CSVStorage) GetTpActionPlans(tpid, tag string) ([]TpActionPlan, error) {
	csvReader, fp, err := csvs.readerFunc(csvs.actiontimingsFn, csvs.sep, getColumnCount(TpActionPlan{}))
	if err != nil {
		log.Print("Could not load action plans file: ", err)
		// allow writing of the other values
		return nil, nil
	}
	if fp != nil {
		defer fp.Close()
	}
	var tpActionPlans []TpActionPlan
	for record, err := csvReader.Read(); err != io.EOF; record, err = csvReader.Read() {
		if tpRate, err := csvLoad(TpActionPlan{}, record); err != nil {
			return nil, err
		} else {
			tpActionPlans = append(tpActionPlans, tpRate.(TpActionPlan))
		}
	}
	return tpActionPlans, nil
}

func (csvs *CSVStorage) GetTpActionTriggers(tpid, tag string) ([]TpActionTrigger, error) {
	csvReader, fp, err := csvs.readerFunc(csvs.actiontriggersFn, csvs.sep, getColumnCount(TpActionTrigger{}))
	if err != nil {
		log.Print("Could not load action triggers file: ", err)
		// allow writing of the other values
		return nil, nil
	}
	if fp != nil {
		defer fp.Close()
	}
	var tpActionTriggers []TpActionTrigger
	for record, err := csvReader.Read(); err != io.EOF; record, err = csvReader.Read() {
		if err != nil {
			log.Print("bad line in action triggers csv: ", err)
			return nil, err
		}
		if tpRate, err := csvLoad(TpActionTrigger{}, record); err != nil {
			return nil, err
		} else {
			tpActionTriggers = append(tpActionTriggers, tpRate.(TpActionTrigger))
		}
	}
	return tpActionTriggers, nil
}

func (csvs *CSVStorage) GetTpAccountActions(filter *TpAccountAction) ([]TpAccountAction, error) {
	csvReader, fp, err := csvs.readerFunc(csvs.accountactionsFn, csvs.sep, getColumnCount(TpAccountAction{}))
	if err != nil {
		log.Print("Could not load account actions file: ", err)
		// allow writing of the other values
		return nil, nil
	}
	if fp != nil {
		defer fp.Close()
	}
	var tpAccountActions []TpAccountAction
	for record, err := csvReader.Read(); err != io.EOF; record, err = csvReader.Read() {
		if err != nil {
			log.Print("bad line in account actions csv: ", err)
			return nil, err
		}
		if tpRate, err := csvLoad(TpAccountAction{}, record); err != nil {
			return nil, err
		} else {
			tpAccountActions = append(tpAccountActions, tpRate.(TpAccountAction))
		}
	}
	return tpAccountActions, nil
}

func (csvs *CSVStorage) GetTpDerivedChargers(filter *TpDerivedCharger) ([]TpDerivedCharger, error) {
	csvReader, fp, err := csvs.readerFunc(csvs.derivedChargersFn, csvs.sep, getColumnCount(TpDerivedCharger{}))
	if err != nil {
		log.Print("Could not load derivedChargers file: ", err)
		// allow writing of the other values
		return nil, nil
	}
	if fp != nil {
		defer fp.Close()
	}
	var tpDerivedChargers []TpDerivedCharger
	for record, err := csvReader.Read(); err != io.EOF; record, err = csvReader.Read() {
		if err != nil {
			log.Print("bad line in derived chargers csv: ", err)
			return nil, err
		}
		if tpRate, err := csvLoad(TpDerivedCharger{}, record); err != nil {
			return nil, err
		} else {
			tpDerivedChargers = append(tpDerivedChargers, tpRate.(TpDerivedCharger))
		}
	}
	return tpDerivedChargers, nil
}

func (csvs *CSVStorage) GetTpCdrStats(tpid, tag string) ([]TpCdrStat, error) {
	csvReader, fp, err := csvs.readerFunc(csvs.derivedChargersFn, csvs.sep, getColumnCount(TpCdrStat{}))
	if err != nil {
		log.Print("Could not load derivedChargers file: ", err)
		// allow writing of the other values
		return nil, nil
	}
	if fp != nil {
		defer fp.Close()
	}
	var tpCdrStats []TpCdrStat
	for record, err := csvReader.Read(); err != io.EOF; record, err = csvReader.Read() {
		if err != nil {
			log.Print("bad line in cdr stats csv: ", err)
			return nil, err
		}
		if tpRate, err := csvLoad(TpCdrStat{}, record); err != nil {
			return nil, err
		} else {
			tpCdrStats = append(tpCdrStats, tpRate.(TpCdrStat))
		}
	}
	return tpCdrStats, nil
}

func (csvs *CSVStorage) GetTpIds() ([]string, error) {
	return nil, errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (csvs *CSVStorage) GetTpTableIds(tpid, table string, distinct utils.TPDistinctIds, filters map[string]string, p *utils.Paginator) ([]string, error) {
	return nil, errors.New(utils.ERR_NOT_IMPLEMENTED)
}
