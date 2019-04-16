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
	"encoding/csv"
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
	destinationsFn           string
	ratesFn                  string
	destinationratesFn       string
	timingsFn                string
	destinationratetimingsFn string
	ratingprofilesFn         string
	sharedgroupsFn           string
	actionsFn                string
	actiontimingsFn          string
	actiontriggersFn         string
	accountactionsFn         string
	resProfilesFn            string
	statsFn                  string
	thresholdsFn             string
	filterFn                 string
	suppProfilesFn           string
	attributeProfilesFn      string
	chargerProfilesFn        string
	dispatcherProfilesFn     string
	dispatcherHostsFn        string
}

func NewFileCSVStorage(sep rune,
	destinationsFn, timingsFn, ratesFn, destinationratesFn,
	destinationratetimingsFn, ratingprofilesFn, sharedgroupsFn,
	actionsFn, actiontimingsFn, actiontriggersFn, accountactionsFn,
	resProfilesFn, statsFn, thresholdsFn,
	filterFn, suppProfilesFn, attributeProfilesFn,
	chargerProfilesFn, dispatcherProfilesFn, dispatcherHostsFn string) *CSVStorage {
	return &CSVStorage{
		sep:                      sep,
		readerFunc:               openFileCSVStorage,
		destinationsFn:           destinationsFn,
		timingsFn:                timingsFn,
		ratesFn:                  ratesFn,
		destinationratesFn:       destinationratesFn,
		destinationratetimingsFn: destinationratetimingsFn,
		ratingprofilesFn:         ratingprofilesFn,
		sharedgroupsFn:           sharedgroupsFn,
		actionsFn:                actionsFn,
		actiontimingsFn:          actiontimingsFn,
		actiontriggersFn:         actiontriggersFn,
		accountactionsFn:         accountactionsFn,
		resProfilesFn:            resProfilesFn,
		statsFn:                  statsFn,
		thresholdsFn:             thresholdsFn,
		filterFn:                 filterFn,
		suppProfilesFn:           suppProfilesFn,
		attributeProfilesFn:      attributeProfilesFn,
		chargerProfilesFn:        chargerProfilesFn,
		dispatcherProfilesFn:     dispatcherProfilesFn,
		dispatcherHostsFn:        dispatcherHostsFn,
	}
}

func NewStringCSVStorage(sep rune,
	destinationsFn, timingsFn, ratesFn, destinationratesFn,
	destinationratetimingsFn, ratingprofilesFn, sharedgroupsFn,
	actionsFn, actiontimingsFn, actiontriggersFn,
	accountactionsFn, resProfilesFn, statsFn,
	thresholdsFn, filterFn, suppProfilesFn,
	attributeProfilesFn, chargerProfilesFn,
	dispatcherProfilesFn, dispatcherHostsFn string) *CSVStorage {
	c := NewFileCSVStorage(sep, destinationsFn, timingsFn,
		ratesFn, destinationratesFn, destinationratetimingsFn,
		ratingprofilesFn, sharedgroupsFn, actionsFn,
		actiontimingsFn, actiontriggersFn, accountactionsFn,
		resProfilesFn, statsFn, thresholdsFn, filterFn,
		suppProfilesFn, attributeProfilesFn, chargerProfilesFn,
		dispatcherProfilesFn, dispatcherHostsFn)
	c.readerFunc = openStringCSVStorage
	return c
}

func openFileCSVStorage(fn string, comma rune,
	nrFields int) (csvReader *csv.Reader, fp *os.File, err error) {
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

func openStringCSVStorage(data string, comma rune,
	nrFields int) (csvReader *csv.Reader, fp *os.File, err error) {
	csvReader = csv.NewReader(strings.NewReader(data))
	csvReader.Comma = comma
	csvReader.Comment = utils.COMMENT_CHAR
	csvReader.FieldsPerRecord = nrFields
	csvReader.TrailingComma = true
	return
}

func (csvs *CSVStorage) GetTPTimings(tpid, id string) ([]*utils.ApierTPTiming, error) {
	csvReader, fp, err := csvs.readerFunc(csvs.timingsFn, csvs.sep, getColumnCount(TpTiming{}))
	if err != nil {
		//log.Print("Could not load timings file: ", err)
		// allow writing of the other values
		return nil, nil
	}
	if fp != nil {
		defer fp.Close()
	}
	var tpTimings TpTimings
	for record, err := csvReader.Read(); err != io.EOF; record, err = csvReader.Read() {
		if err != nil {
			log.Printf("bad line in %s, %s\n", csvs.timingsFn, err.Error())
			return nil, err
		}
		if tpTiming, err := csvLoad(TpTiming{}, record); err != nil {
			log.Print("error loading timing: ", err)
			return nil, err
		} else {
			tm := tpTiming.(TpTiming)
			tm.Tpid = tpid
			tpTimings = append(tpTimings, tm)
		}
	}
	return tpTimings.AsTPTimings(), nil
}

func (csvs *CSVStorage) GetTPDestinations(tpid, id string) ([]*utils.TPDestination, error) {
	csvReader, fp, err := csvs.readerFunc(csvs.destinationsFn, csvs.sep, getColumnCount(TpDestination{}))
	if err != nil {
		//log.Print("Could not load destinations file: ", err)
		// allow writing of the other values
		return nil, nil
	}
	if fp != nil {
		defer fp.Close()
	}
	var tpDests TpDestinations
	for record, err := csvReader.Read(); err != io.EOF; record, err = csvReader.Read() {
		if err != nil {
			log.Printf("bad line in %s, %s\n", csvs.destinationsFn, err.Error())
			return nil, err
		}
		if tpDest, err := csvLoad(TpDestination{}, record); err != nil {
			log.Print("error loading destination: ", err)
			return nil, err
		} else {
			d := tpDest.(TpDestination)
			d.Tpid = tpid
			tpDests = append(tpDests, d)
		}
	}
	return tpDests.AsTPDestinations(), nil
}

func (csvs *CSVStorage) GetTPRates(tpid, id string) ([]*utils.TPRate, error) {
	csvReader, fp, err := csvs.readerFunc(csvs.ratesFn, csvs.sep, getColumnCount(TpRate{}))
	if err != nil {
		//log.Print("Could not load rates file: ", err)
		// allow writing of the other values
		return nil, nil
	}
	if fp != nil {
		defer fp.Close()
	}
	var tpRates TpRates
	for record, err := csvReader.Read(); err != io.EOF; record, err = csvReader.Read() {
		if err != nil {
			log.Printf("bad line in %s, %s\n", csvs.ratesFn, err.Error())
			return nil, err
		}
		if tpRate, err := csvLoad(TpRate{}, record); err != nil {
			log.Print("error loading rate: ", err)
			return nil, err
		} else {
			r := tpRate.(TpRate)
			r.Tpid = tpid
			tpRates = append(tpRates, r)
		}
	}
	if rs, err := tpRates.AsTPRates(); err != nil {
		return nil, err
	} else {
		return rs, nil
	}
}

func (csvs *CSVStorage) GetTPDestinationRates(tpid, id string, p *utils.Paginator) ([]*utils.TPDestinationRate, error) {
	csvReader, fp, err := csvs.readerFunc(csvs.destinationratesFn, csvs.sep, getColumnCount(TpDestinationRate{}))
	if err != nil {
		//log.Print("Could not load destination_rates file: ", err)
		// allow writing of the other values
		return nil, nil
	}
	if fp != nil {
		defer fp.Close()
	}
	var tpDestinationRates TpDestinationRates
	for record, err := csvReader.Read(); err != io.EOF; record, err = csvReader.Read() {
		if err != nil {
			log.Printf("bad line in %s, %s\n", csvs.destinationratesFn, err.Error())
			return nil, err
		}
		if tpRate, err := csvLoad(TpDestinationRate{}, record); err != nil {
			log.Print("error loading destination rate: ", err)
			return nil, err
		} else {
			dr := tpRate.(TpDestinationRate)
			dr.Tpid = tpid
			tpDestinationRates = append(tpDestinationRates, dr)
		}
	}
	if drs, err := tpDestinationRates.AsTPDestinationRates(); err != nil {
		return nil, err
	} else {
		return drs, nil
	}
}

func (csvs *CSVStorage) GetTPRatingPlans(tpid, id string, p *utils.Paginator) ([]*utils.TPRatingPlan, error) {
	csvReader, fp, err := csvs.readerFunc(csvs.destinationratetimingsFn, csvs.sep, getColumnCount(TpRatingPlan{}))
	if err != nil {
		//log.Print("Could not load rate plans file: ", err)
		// allow writing of the other values
		return nil, nil
	}
	if fp != nil {
		defer fp.Close()
	}
	var tpRatingPlans TpRatingPlans
	for record, err := csvReader.Read(); err != io.EOF; record, err = csvReader.Read() {
		if err != nil {
			log.Printf("bad line in %s, %s\n", csvs.destinationratetimingsFn, err.Error())
			return nil, err
		}
		if tpRate, err := csvLoad(TpRatingPlan{}, record); err != nil {
			log.Print("error loading rating plan: ", err)
			return nil, err
		} else {
			rp := tpRate.(TpRatingPlan)
			rp.Tpid = tpid
			tpRatingPlans = append(tpRatingPlans, rp)
		}
	}
	if rps, err := tpRatingPlans.AsTPRatingPlans(); err != nil {
		return nil, err
	} else {
		return rps, nil
	}
}

func (csvs *CSVStorage) GetTPRatingProfiles(filter *utils.TPRatingProfile) ([]*utils.TPRatingProfile, error) {
	csvReader, fp, err := csvs.readerFunc(csvs.ratingprofilesFn, csvs.sep, getColumnCount(TpRatingProfile{}))
	if err != nil {
		//log.Print("Could not load rating profiles file: ", err)
		// allow writing of the other values
		return nil, nil
	}
	if fp != nil {
		defer fp.Close()
	}
	var tpRatingProfiles TpRatingProfiles
	for record, err := csvReader.Read(); err != io.EOF; record, err = csvReader.Read() {
		if err != nil {
			log.Printf("bad line in %s, %s\n", csvs.ratingprofilesFn, err.Error())
			return nil, err
		}
		if tpRate, err := csvLoad(TpRatingProfile{}, record); err != nil {
			log.Print("error loading rating profile: ", err)
			return nil, err
		} else {
			rpf := tpRate.(TpRatingProfile)
			if filter != nil {
				rpf.Tpid = filter.TPid
				rpf.Loadid = filter.LoadId
			}
			tpRatingProfiles = append(tpRatingProfiles, rpf)
		}
	}
	if rps, err := tpRatingProfiles.AsTPRatingProfiles(); err != nil {
		return nil, err
	} else {
		return rps, nil
	}
}

func (csvs *CSVStorage) GetTPSharedGroups(tpid, id string) ([]*utils.TPSharedGroups, error) {
	csvReader, fp, err := csvs.readerFunc(csvs.sharedgroupsFn, csvs.sep, getColumnCount(TpSharedGroup{}))
	if err != nil {
		//log.Print("Could not load shared groups file: ", err)
		// allow writing of the other values
		return nil, nil
	}
	if fp != nil {
		defer fp.Close()
	}
	var tpSharedGroups TpSharedGroups
	for record, err := csvReader.Read(); err != io.EOF; record, err = csvReader.Read() {
		if err != nil {
			log.Printf("bad line in %s, %s\n", csvs.sharedgroupsFn, err.Error())
			return nil, err
		}
		if tpRate, err := csvLoad(TpSharedGroup{}, record); err != nil {
			log.Print("error loading shared group: ", err)
			return nil, err
		} else {
			sg := tpRate.(TpSharedGroup)
			sg.Tpid = tpid
			tpSharedGroups = append(tpSharedGroups, sg)
		}
	}
	if sgs, err := tpSharedGroups.AsTPSharedGroups(); err != nil {
		return nil, err
	} else {
		return sgs, nil
	}
}

func (csvs *CSVStorage) GetTPActions(tpid, id string) ([]*utils.TPActions, error) {
	csvReader, fp, err := csvs.readerFunc(csvs.actionsFn, csvs.sep, getColumnCount(TpAction{}))
	if err != nil {
		//log.Print("Could not load action file: ", err)
		// allow writing of the other values
		return nil, nil
	}
	if fp != nil {
		defer fp.Close()
	}
	var tpActions TpActions
	for record, err := csvReader.Read(); err != io.EOF; record, err = csvReader.Read() {
		if err != nil {
			log.Printf("bad line in %s, %s\n", csvs.actionsFn, err.Error())
			return nil, err
		}
		if tpAction, err := csvLoad(TpAction{}, record); err != nil {
			log.Print("error loading action: ", err)
			return nil, err
		} else {
			a := tpAction.(TpAction)
			a.Tpid = tpid
			tpActions = append(tpActions, a)
		}
	}
	if as, err := tpActions.AsTPActions(); err != nil {
		return nil, err
	} else {
		return as, nil
	}
}

func (csvs *CSVStorage) GetTPActionPlans(tpid, id string) ([]*utils.TPActionPlan, error) {
	csvReader, fp, err := csvs.readerFunc(csvs.actiontimingsFn, csvs.sep, getColumnCount(TpActionPlan{}))
	if err != nil {
		//log.Print("Could not load action plans file: ", err)
		// allow writing of the other values
		return nil, nil
	}
	if fp != nil {
		defer fp.Close()
	}
	var tpActionPlans TpActionPlans
	for record, err := csvReader.Read(); err != io.EOF; record, err = csvReader.Read() {
		if tpRate, err := csvLoad(TpActionPlan{}, record); err != nil {
			log.Print("error loading action plan: ", err)
			return nil, err
		} else {
			ap := tpRate.(TpActionPlan)
			ap.Tpid = tpid
			tpActionPlans = append(tpActionPlans, ap)
		}
	}
	if aps, err := tpActionPlans.AsTPActionPlans(); err != nil {
		return nil, err
	} else {
		return aps, nil
	}
}

func (csvs *CSVStorage) GetTPActionTriggers(tpid, id string) ([]*utils.TPActionTriggers, error) {
	csvReader, fp, err := csvs.readerFunc(csvs.actiontriggersFn, csvs.sep, getColumnCount(TpActionTrigger{}))
	if err != nil {
		//log.Print("Could not load action triggers file: ", err)
		// allow writing of the other values
		return nil, nil
	}
	if fp != nil {
		defer fp.Close()
	}
	var tpActionTriggers TpActionTriggers
	for record, err := csvReader.Read(); err != io.EOF; record, err = csvReader.Read() {
		if err != nil {
			log.Printf("bad line in %s, %s\n", csvs.actiontriggersFn, err.Error())
			return nil, err
		}
		if tpAt, err := csvLoad(TpActionTrigger{}, record); err != nil {
			log.Print("error loading action trigger: ", err)
			return nil, err
		} else {
			at := tpAt.(TpActionTrigger)
			at.Tpid = tpid
			tpActionTriggers = append(tpActionTriggers, at)
		}
	}
	if ats, err := tpActionTriggers.AsTPActionTriggers(); err != nil {
		return nil, err
	} else {
		return ats, nil
	}
}

func (csvs *CSVStorage) GetTPAccountActions(filter *utils.TPAccountActions) ([]*utils.TPAccountActions, error) {
	csvReader, fp, err := csvs.readerFunc(csvs.accountactionsFn, csvs.sep, getColumnCount(TpAccountAction{}))
	if err != nil {
		//log.Print("Could not load account actions file: ", err)
		// allow writing of the other values
		return nil, nil
	}
	if fp != nil {
		defer fp.Close()
	}
	var tpAccountActions TpAccountActions
	for record, err := csvReader.Read(); err != io.EOF; record, err = csvReader.Read() {
		if err != nil {
			log.Printf("bad line in %s, %s\n", csvs.accountactionsFn, err.Error())
			return nil, err
		}
		if tpAa, err := csvLoad(TpAccountAction{}, record); err != nil {
			log.Print("error loading account action: ", err)
			return nil, err
		} else {
			aa := tpAa.(TpAccountAction)
			if filter != nil {
				aa.Tpid = filter.TPid
				aa.Loadid = filter.LoadId
			}
			tpAccountActions = append(tpAccountActions, aa)
		}
	}
	if ats, err := tpAccountActions.AsTPAccountActions(); err != nil {
		return nil, err
	} else {
		return ats, nil
	}
}

func (csvs *CSVStorage) GetTPResources(tpid, tenant, id string) ([]*utils.TPResourceProfile, error) {
	csvReader, fp, err := csvs.readerFunc(csvs.resProfilesFn, csvs.sep, getColumnCount(TpResource{}))
	if err != nil {
		//log.Print("Could not load resource limits file: ", err)
		// allow writing of the other values
		return nil, nil
	}
	if fp != nil {
		defer fp.Close()
	}
	var tpResLimits TpResources
	for record, err := csvReader.Read(); err != io.EOF; record, err = csvReader.Read() {
		if err != nil {
			log.Printf("bad line in %s, %s\n", csvs.resProfilesFn, err.Error())
			return nil, err
		}
		if tpResLimit, err := csvLoad(TpResource{}, record); err != nil {
			log.Print("error loading resourceprofiles: ", err)
			return nil, err
		} else {
			tpLimit := tpResLimit.(TpResource)
			tpLimit.Tpid = tpid
			tpResLimits = append(tpResLimits, &tpLimit)
		}
	}
	return tpResLimits.AsTPResources(), nil
}

func (csvs *CSVStorage) GetTPStats(tpid, tenant, id string) ([]*utils.TPStatProfile, error) {
	csvReader, fp, err := csvs.readerFunc(csvs.statsFn, csvs.sep, getColumnCount(TpStat{}))
	if err != nil {
		//log.Print("Could not load stats file: ", err)
		// allow writing of the other values
		return nil, nil
	}
	if fp != nil {
		defer fp.Close()
	}
	var tpStats TpStats
	for record, err := csvReader.Read(); err != io.EOF; record, err = csvReader.Read() {
		if err != nil {
			log.Printf("bad line in %s, %s\n", csvs.statsFn, err.Error())
			return nil, err
		}
		if tpstats, err := csvLoad(TpStat{}, record); err != nil {
			log.Print("error loading TPStats: ", err)
			return nil, err
		} else {
			tPstats := tpstats.(TpStat)
			tPstats.Tpid = tpid
			tpStats = append(tpStats, &tPstats)
		}
	}
	return tpStats.AsTPStats(), nil
}

func (csvs *CSVStorage) GetTPThresholds(tpid, tenant, id string) ([]*utils.TPThresholdProfile, error) {
	csvReader, fp, err := csvs.readerFunc(csvs.thresholdsFn, csvs.sep, getColumnCount(TpThreshold{}))
	if err != nil {
		//log.Print("Could not load threshold file: ", err)
		// allow writing of the other values
		return nil, nil
	}
	if fp != nil {
		defer fp.Close()
	}
	var tpThreshold TpThresholds
	for record, err := csvReader.Read(); err != io.EOF; record, err = csvReader.Read() {
		if err != nil {
			log.Printf("bad line in %s, %s\n", csvs.thresholdsFn, err.Error())
			return nil, err
		}
		if thresholdCfg, err := csvLoad(TpThreshold{}, record); err != nil {
			log.Print("error loading TPThreshold: ", err)
			return nil, err
		} else {
			tHresholdCfg := thresholdCfg.(TpThreshold)
			tHresholdCfg.Tpid = tpid
			tpThreshold = append(tpThreshold, &tHresholdCfg)
		}
	}
	return tpThreshold.AsTPThreshold(), nil
}

func (csvs *CSVStorage) GetTPFilters(tpid, tenant, id string) ([]*utils.TPFilterProfile, error) {
	csvReader, fp, err := csvs.readerFunc(csvs.filterFn, csvs.sep, getColumnCount(TpFilter{}))
	if err != nil {
		//log.Print("Could not load filter file: ", err)
		// allow writing of the other values
		return nil, nil
	}
	if fp != nil {
		defer fp.Close()
	}
	var tpFilter TpFilterS
	for record, err := csvReader.Read(); err != io.EOF; record, err = csvReader.Read() {
		if err != nil {
			log.Printf("bad line in %s, %s\n", csvs.filterFn, err.Error())
			return nil, err
		}
		if filterCfg, err := csvLoad(TpFilter{}, record); err != nil {
			log.Print("error loading TPFilter: ", err)
			return nil, err
		} else {
			fIlterCfg := filterCfg.(TpFilter)
			fIlterCfg.Tpid = tpid
			tpFilter = append(tpFilter, &fIlterCfg)
		}
	}
	return tpFilter.AsTPFilter(), nil
}

func (csvs *CSVStorage) GetTPSuppliers(tpid, tenant, id string) ([]*utils.TPSupplierProfile, error) {
	csvReader, fp, err := csvs.readerFunc(csvs.suppProfilesFn, csvs.sep, getColumnCount(TpSupplier{}))
	if err != nil {
		//log.Print("Could not load lcrProfile file: ", err)
		// allow writing of the other values
		return nil, nil
	}
	if fp != nil {
		defer fp.Close()
	}
	var tpSPPs TpSuppliers
	for record, err := csvReader.Read(); err != io.EOF; record, err = csvReader.Read() {
		if err != nil {
			log.Printf("bad line in %s, %s\n", csvs.suppProfilesFn, err.Error())
			return nil, err
		}
		if suppProfile, err := csvLoad(TpSupplier{}, record); err != nil {
			log.Print("error loading tpSupplier: ", err)
			return nil, err
		} else {
			suppProfile := suppProfile.(TpSupplier)
			suppProfile.Tpid = tpid
			tpSPPs = append(tpSPPs, &suppProfile)
		}
	}
	return tpSPPs.AsTPSuppliers(), nil
}

func (csvs *CSVStorage) GetTPAttributes(tpid, tenant, id string) ([]*utils.TPAttributeProfile, error) {
	csvReader, fp, err := csvs.readerFunc(csvs.attributeProfilesFn, csvs.sep, getColumnCount(TPAttribute{}))
	if err != nil {
		//log.Print("Could not load AttributeProfile file: ", err)
		// allow writing of the other values
		return nil, nil
	}
	if fp != nil {
		defer fp.Close()
	}
	var tpAls TPAttributes
	for record, err := csvReader.Read(); err != io.EOF; record, err = csvReader.Read() {
		if err != nil {
			log.Printf("bad line in %s, %s\n", csvs.attributeProfilesFn, err.Error())
			return nil, err
		}
		if attributeProfile, err := csvLoad(TPAttribute{}, record); err != nil {
			log.Print("error loading tpAttributeProfile: ", err)
			return nil, err
		} else {
			attributeProfile := attributeProfile.(TPAttribute)
			attributeProfile.Tpid = tpid
			tpAls = append(tpAls, &attributeProfile)
		}
	}
	return tpAls.AsTPAttributes(), nil
}

func (csvs *CSVStorage) GetTPChargers(tpid, tenant, id string) ([]*utils.TPChargerProfile, error) {
	csvReader, fp, err := csvs.readerFunc(csvs.chargerProfilesFn, csvs.sep, getColumnCount(TPCharger{}))
	if err != nil {
		//log.Print("Could not load AttributeProfile file: ", err)
		// allow writing of the other values
		return nil, nil
	}
	if fp != nil {
		defer fp.Close()
	}
	var tpCPPs TPChargers
	for record, err := csvReader.Read(); err != io.EOF; record, err = csvReader.Read() {
		if err != nil {
			log.Printf("bad line in %s, %s\n", csvs.chargerProfilesFn, err.Error())
			return nil, err
		}
		if cpp, err := csvLoad(TPCharger{}, record); err != nil {
			log.Print("error loading tpChargerProfile: ", err)
			return nil, err
		} else {
			cpp := cpp.(TPCharger)
			cpp.Tpid = tpid
			tpCPPs = append(tpCPPs, &cpp)
		}
	}
	return tpCPPs.AsTPChargers(), nil
}

func (csvs *CSVStorage) GetTPDispatcherProfiles(tpid, tenant, id string) ([]*utils.TPDispatcherProfile, error) {
	csvReader, fp, err := csvs.readerFunc(csvs.dispatcherProfilesFn, csvs.sep, getColumnCount(TPDispatcherProfile{}))
	if err != nil {
		// allow writing of the other values
		return nil, nil
	}
	if fp != nil {
		defer fp.Close()
	}
	var tpDPPs TPDispatcherProfiles
	for record, err := csvReader.Read(); err != io.EOF; record, err = csvReader.Read() {
		if err != nil {
			log.Printf("bad line in %s, %s\n", csvs.dispatcherProfilesFn, err.Error())
			return nil, err
		}
		if dpp, err := csvLoad(TPDispatcherProfile{}, record); err != nil {
			log.Print("error loading tpDispatcherProfile: ", err)
			return nil, err
		} else {
			dpp := dpp.(TPDispatcherProfile)
			dpp.Tpid = tpid
			tpDPPs = append(tpDPPs, &dpp)
		}
	}
	return tpDPPs.AsTPDispatcherProfiles(), nil
}

func (csvs *CSVStorage) GetTPDispatcherHosts(tpid, tenant, id string) ([]*utils.TPDispatcherHost, error) {
	csvReader, fp, err := csvs.readerFunc(csvs.dispatcherHostsFn, csvs.sep, getColumnCount(TPDispatcherHost{}))
	if err != nil {
		// allow writing of the other values
		return nil, nil
	}
	if fp != nil {
		defer fp.Close()
	}
	var tpDDHs TPDispatcherHosts
	for record, err := csvReader.Read(); err != io.EOF; record, err = csvReader.Read() {
		if err != nil {
			log.Printf("bad line in %s, %s\n", csvs.dispatcherHostsFn, err.Error())
			return nil, err
		}
		if dpp, err := csvLoad(TPDispatcherHost{}, record); err != nil {
			log.Print("error loading tpDispatcherHost: ", err)
			return nil, err
		} else {
			dpp := dpp.(TPDispatcherHost)
			dpp.Tpid = tpid
			tpDDHs = append(tpDDHs, &dpp)
		}
	}
	return tpDDHs.AsTPDispatcherHosts(), nil
}

func (csvs *CSVStorage) GetTpIds(colName string) ([]string, error) {
	return nil, utils.ErrNotImplemented
}

func (csvs *CSVStorage) GetTpTableIds(tpid, table string,
	distinct utils.TPDistinctIds, filters map[string]string, p *utils.PaginatorWithSearch) ([]string, error) {
	return nil, utils.ErrNotImplemented
}
