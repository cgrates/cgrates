package engine

import (
	"encoding/csv"
	"errors"
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

func (csvs *CSVStorage) GetTpTimings(string, string) ([]*TpTiming, error) {
	csvReader, fp, err := csvs.readerFunc(csvs.timingsFn, csvs.sep, utils.TIMINGS_NRCOLS)
	if err != nil {
		log.Print("Could not load timings file: ", err)
		// allow writing of the other values
		return nil, nil
	}
	if fp != nil {
		defer fp.Close()
	}
	var tpTimings []*TpTiming
	for record, err := csvReader.Read(); err == nil; record, err = csvReader.Read() {
		if tpTiming, err := csvLoad(TpTiming{}, record); err != nil {
			return nil, err
		} else {
			tp := tpTiming.(TpTiming)
			tpTimings = append(tpTimings, &tp)
		}
	}
	return nil, nil
}

func (csvs *CSVStorage) GetTpDestinations(tpid, tag string) ([]*TpDestination, error) {
	csvReader, fp, err := csvs.readerFunc(csvs.destinationsFn, csvs.sep, utils.DESTINATIONS_NRCOLS)
	if err != nil {
		log.Print("Could not load destinations file: ", err)
		// allow writing of the other values
		return nil, nil
	}
	if fp != nil {
		defer fp.Close()
	}
	var tpDests []*TpDestination
	for record, err := csvReader.Read(); err == nil; record, err = csvReader.Read() {
		if tpDest, err := csvLoad(TpDestination{}, record); err != nil {
			return nil, err
		} else {
			tp := tpDest.(TpDestination)
			tpDests = append(tpDests, &tp)
		}
		//log.Printf("%+v\n", tpDest)
	}
	return tpDests, nil
}

func (csvs *CSVStorage) GetTpRates(tpid, tag string) ([]*TpRate, error) {
	csvReader, fp, err := csvs.readerFunc(csvs.ratesFn, csvs.sep, utils.RATES_NRCOLS)
	if err != nil {
		log.Print("Could not load rates file: ", err)
		// allow writing of the other values
		return nil, nil
	}
	if fp != nil {
		defer fp.Close()
	}
	var tpRates []*TpRate
	for record, err := csvReader.Read(); err == nil; record, err = csvReader.Read() {
		if tpRate, err := csvLoad(TpRate{}, record); err != nil {
			return nil, err
		} else {
			tp := tpRate.(TpRate)
			tpRates = append(tpRates, &tp)
		}
		//log.Printf("%+v\n", tpRate)
	}
	return tpRates, nil
}

func (csvs *CSVStorage) GetTpDestinationRates(string, string, *utils.Paginator) (map[string]*utils.TPDestinationRate, error) {
	return nil, nil
}

func (csvs *CSVStorage) GetTpRatingPlans(string, string, *utils.Paginator) (map[string][]*utils.TPRatingPlanBinding, error) {
	return nil, nil
}

func (csvs *CSVStorage) GetTpRatingProfiles(*utils.TPRatingProfile) (map[string]*utils.TPRatingProfile, error) {
	return nil, nil
}

func (csvs *CSVStorage) GetTpSharedGroups(string, string) (map[string][]*utils.TPSharedGroup, error) {
	return nil, nil
}

func (csvs *CSVStorage) GetTpCdrStats(string, string) (map[string][]*utils.TPCdrStat, error) {
	return nil, nil
}

func (csvs *CSVStorage) GetTpDerivedChargers(*utils.TPDerivedChargers) (map[string]*utils.TPDerivedChargers, error) {
	return nil, nil
}

func (csvs *CSVStorage) GetTpLCRs(string, string) (map[string]*LCR, error) { return nil, nil }

func (csvs *CSVStorage) GetTpActions(string, string) (map[string][]*utils.TPAction, error) {
	return nil, nil
}

func (csvs *CSVStorage) GetTPActionTimings(string, string) (map[string][]*utils.TPActionTiming, error) {
	return nil, nil
}

func (csvs *CSVStorage) GetTpActionTriggers(string, string) (map[string][]*utils.TPActionTrigger, error) {
	return nil, nil
}

func (csvs *CSVStorage) GetTpAccountActions(*utils.TPAccountActions) (map[string]*utils.TPAccountActions, error) {
	return nil, nil
}

func (csvs *CSVStorage) GetTPIds() ([]string, error) {
	return nil, errors.New(utils.ERR_NOT_IMPLEMENTED)
}

func (csvs *CSVStorage) GetTPTableIds(string, string, utils.TPDistinctIds, map[string]string, *utils.Paginator) ([]string, error) {
	return nil, errors.New(utils.ERR_NOT_IMPLEMENTED)
}
