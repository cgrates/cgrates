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
	destinationsFn           []string
	ratesFn                  []string
	destinationratesFn       []string
	timingsFn                []string
	destinationratetimingsFn []string
	ratingprofilesFn         []string
	sharedgroupsFn           []string
	actionsFn                []string
	actiontimingsFn          []string
	actiontriggersFn         []string
	accountactionsFn         []string
	resProfilesFn            []string
	statsFn                  []string
	thresholdsFn             []string
	filterFn                 []string
	suppProfilesFn           []string
	attributeProfilesFn      []string
	chargerProfilesFn        []string
	dispatcherProfilesFn     []string
	dispatcherHostsFn        []string
}

func NewFileCSVStorage(sep rune,
	destinationsFn, timingsFn, ratesFn, destinationratesFn,
	destinationratetimingsFn, ratingprofilesFn, sharedgroupsFn,
	actionsFn, actiontimingsFn, actiontriggersFn, accountactionsFn,
	resProfilesFn, statsFn, thresholdsFn,
	filterFn, suppProfilesFn, attributeProfilesFn,
	chargerProfilesFn, dispatcherProfilesFn, dispatcherHostsFn []string) *CSVStorage {
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
	c := NewFileCSVStorage(sep, []string{destinationsFn}, []string{timingsFn},
		[]string{ratesFn}, []string{destinationratesFn}, []string{destinationratetimingsFn},
		[]string{ratingprofilesFn}, []string{sharedgroupsFn}, []string{actionsFn},
		[]string{actiontimingsFn}, []string{actiontriggersFn}, []string{accountactionsFn},
		[]string{resProfilesFn}, []string{statsFn}, []string{thresholdsFn}, []string{filterFn},
		[]string{suppProfilesFn}, []string{attributeProfilesFn}, []string{chargerProfilesFn},
		[]string{dispatcherProfilesFn}, []string{dispatcherHostsFn})
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

func (csvs *CSVStorage) proccesData(listType interface{}, fns []string, process func(interface{})) error {
	collumnCount := getColumnCount(listType)
	for _, fileName := range fns {
		csvReader, fp, err := csvs.readerFunc(fileName, csvs.sep, collumnCount)
		if err != nil {
			// maybe a log to view if failed to open file
			continue // try read the rest
		}
		if err = func() error { // to execute defer corectly
			if fp != nil {
				defer fp.Close()
			}
			for record, err := csvReader.Read(); err != io.EOF; record, err = csvReader.Read() {
				if err != nil {
					log.Printf("bad line in %s, %s\n", fileName, err.Error())
					return err
				}
				if item, err := csvLoad(listType, record); err != nil {
					log.Printf("error loading %s: %v", "", err)
					return err
				} else {
					process(item)
				}
			}
			return nil
		}(); err != nil {
			return err
		}
	}
	return nil
}

func (csvs *CSVStorage) GetTPTimings(tpid, id string) ([]*utils.ApierTPTiming, error) {
	var tpTimings TpTimings
	if err := csvs.proccesData(TpTiming{}, csvs.timingsFn, func(tp interface{}) {
		tm := tp.(TpTiming)
		tm.Tpid = tpid
		tpTimings = append(tpTimings, tm)
	}); err != nil {
		return nil, err
	}
	return tpTimings.AsTPTimings(), nil
}

func (csvs *CSVStorage) GetTPDestinations(tpid, id string) ([]*utils.TPDestination, error) {
	var tpDests TpDestinations
	if err := csvs.proccesData(TpDestination{}, csvs.destinationsFn, func(tp interface{}) {
		d := tp.(TpDestination)
		d.Tpid = tpid
		tpDests = append(tpDests, d)
	}); err != nil {
		return nil, err
	}
	return tpDests.AsTPDestinations(), nil
}

func (csvs *CSVStorage) GetTPRates(tpid, id string) ([]*utils.TPRate, error) {
	var tpRates TpRates
	if err := csvs.proccesData(TpRate{}, csvs.ratesFn, func(tp interface{}) {
		r := tp.(TpRate)
		r.Tpid = tpid
		tpRates = append(tpRates, r)
	}); err != nil {
		return nil, err
	}
	return tpRates.AsTPRates()
}

func (csvs *CSVStorage) GetTPDestinationRates(tpid, id string, p *utils.Paginator) ([]*utils.TPDestinationRate, error) {
	var tpDestinationRates TpDestinationRates
	if err := csvs.proccesData(TpDestinationRate{}, csvs.destinationratesFn, func(tp interface{}) {
		dr := tp.(TpDestinationRate)
		dr.Tpid = tpid
		tpDestinationRates = append(tpDestinationRates, dr)
	}); err != nil {
		return nil, err
	}
	return tpDestinationRates.AsTPDestinationRates()
}

func (csvs *CSVStorage) GetTPRatingPlans(tpid, id string, p *utils.Paginator) ([]*utils.TPRatingPlan, error) {
	var tpRatingPlans TpRatingPlans
	if err := csvs.proccesData(TpRatingPlan{}, csvs.destinationratetimingsFn, func(tp interface{}) {
		rp := tp.(TpRatingPlan)
		rp.Tpid = tpid
		tpRatingPlans = append(tpRatingPlans, rp)
	}); err != nil {
		return nil, err
	}
	return tpRatingPlans.AsTPRatingPlans()
}

func (csvs *CSVStorage) GetTPRatingProfiles(filter *utils.TPRatingProfile) ([]*utils.TPRatingProfile, error) {
	var tpRatingProfiles TpRatingProfiles
	if err := csvs.proccesData(TpRatingProfile{}, csvs.ratingprofilesFn, func(tp interface{}) {
		rpf := tp.(TpRatingProfile)
		if filter != nil {
			rpf.Tpid = filter.TPid
			rpf.Loadid = filter.LoadId
		}
		tpRatingProfiles = append(tpRatingProfiles, rpf)
	}); err != nil {
		return nil, err
	}
	return tpRatingProfiles.AsTPRatingProfiles()
}

func (csvs *CSVStorage) GetTPSharedGroups(tpid, id string) ([]*utils.TPSharedGroups, error) {
	var tpSharedGroups TpSharedGroups
	if err := csvs.proccesData(TpSharedGroup{}, csvs.sharedgroupsFn, func(tp interface{}) {
		sg := tp.(TpSharedGroup)
		sg.Tpid = tpid
		tpSharedGroups = append(tpSharedGroups, sg)
	}); err != nil {
		return nil, err
	}
	return tpSharedGroups.AsTPSharedGroups()
}

func (csvs *CSVStorage) GetTPActions(tpid, id string) ([]*utils.TPActions, error) {
	var tpActions TpActions
	if err := csvs.proccesData(TpAction{}, csvs.actionsFn, func(tp interface{}) {
		a := tp.(TpAction)
		a.Tpid = tpid
		tpActions = append(tpActions, a)
	}); err != nil {
		return nil, err
	}
	return tpActions.AsTPActions()
}

func (csvs *CSVStorage) GetTPActionPlans(tpid, id string) ([]*utils.TPActionPlan, error) {
	var tpActionPlans TpActionPlans
	if err := csvs.proccesData(TpActionPlan{}, csvs.actiontimingsFn, func(tp interface{}) {
		ap := tp.(TpActionPlan)
		ap.Tpid = tpid
		tpActionPlans = append(tpActionPlans, ap)
	}); err != nil {
		return nil, err
	}
	return tpActionPlans.AsTPActionPlans()
}

func (csvs *CSVStorage) GetTPActionTriggers(tpid, id string) ([]*utils.TPActionTriggers, error) {
	var tpActionTriggers TpActionTriggers
	if err := csvs.proccesData(TpActionTrigger{}, csvs.actiontriggersFn, func(tp interface{}) {
		at := tp.(TpActionTrigger)
		at.Tpid = tpid
		tpActionTriggers = append(tpActionTriggers, at)
	}); err != nil {
		return nil, err
	}
	return tpActionTriggers.AsTPActionTriggers()
}

func (csvs *CSVStorage) GetTPAccountActions(filter *utils.TPAccountActions) ([]*utils.TPAccountActions, error) {
	var tpAccountActions TpAccountActions
	if err := csvs.proccesData(TpAccountAction{}, csvs.accountactionsFn, func(tp interface{}) {
		aa := tp.(TpAccountAction)
		if filter != nil {
			aa.Tpid = filter.TPid
			aa.Loadid = filter.LoadId
		}
		tpAccountActions = append(tpAccountActions, aa)
	}); err != nil {
		return nil, err
	}
	return tpAccountActions.AsTPAccountActions()
}

func (csvs *CSVStorage) GetTPResources(tpid, tenant, id string) ([]*utils.TPResourceProfile, error) {
	var tpResLimits TpResources
	if err := csvs.proccesData(TpResource{}, csvs.resProfilesFn, func(tp interface{}) {
		tpLimit := tp.(TpResource)
		tpLimit.Tpid = tpid
		tpResLimits = append(tpResLimits, &tpLimit)
	}); err != nil {
		return nil, err
	}
	return tpResLimits.AsTPResources(), nil
}

func (csvs *CSVStorage) GetTPStats(tpid, tenant, id string) ([]*utils.TPStatProfile, error) {
	var tpStats TpStats
	if err := csvs.proccesData(TpStat{}, csvs.statsFn, func(tp interface{}) {
		tPstats := tp.(TpStat)
		tPstats.Tpid = tpid
		tpStats = append(tpStats, &tPstats)
	}); err != nil {
		return nil, err
	}
	return tpStats.AsTPStats(), nil
}

func (csvs *CSVStorage) GetTPThresholds(tpid, tenant, id string) ([]*utils.TPThresholdProfile, error) {
	var tpThreshold TpThresholds
	if err := csvs.proccesData(TpThreshold{}, csvs.thresholdsFn, func(tp interface{}) {
		tHresholdCfg := tp.(TpThreshold)
		tHresholdCfg.Tpid = tpid
		tpThreshold = append(tpThreshold, &tHresholdCfg)
	}); err != nil {
		return nil, err
	}
	return tpThreshold.AsTPThreshold(), nil
}

func (csvs *CSVStorage) GetTPFilters(tpid, tenant, id string) ([]*utils.TPFilterProfile, error) {
	var tpFilter TpFilterS
	if err := csvs.proccesData(TpFilter{}, csvs.filterFn, func(tp interface{}) {
		fIlterCfg := tp.(TpFilter)
		fIlterCfg.Tpid = tpid
		tpFilter = append(tpFilter, &fIlterCfg)
	}); err != nil {
		return nil, err
	}
	return tpFilter.AsTPFilter(), nil
}

func (csvs *CSVStorage) GetTPSuppliers(tpid, tenant, id string) ([]*utils.TPSupplierProfile, error) {
	var tpSPPs TpSuppliers
	if err := csvs.proccesData(TpSupplier{}, csvs.suppProfilesFn, func(tp interface{}) {
		suppProfile := tp.(TpSupplier)
		suppProfile.Tpid = tpid
		tpSPPs = append(tpSPPs, &suppProfile)
	}); err != nil {
		return nil, err
	}
	return tpSPPs.AsTPSuppliers(), nil
}

func (csvs *CSVStorage) GetTPAttributes(tpid, tenant, id string) ([]*utils.TPAttributeProfile, error) {
	var tpAls TPAttributes
	if err := csvs.proccesData(TPAttribute{}, csvs.attributeProfilesFn, func(tp interface{}) {
		attributeProfile := tp.(TPAttribute)
		attributeProfile.Tpid = tpid
		tpAls = append(tpAls, &attributeProfile)
	}); err != nil {
		return nil, err
	}
	return tpAls.AsTPAttributes(), nil
}

func (csvs *CSVStorage) GetTPChargers(tpid, tenant, id string) ([]*utils.TPChargerProfile, error) {
	var tpCPPs TPChargers
	if err := csvs.proccesData(TPCharger{}, csvs.chargerProfilesFn, func(tp interface{}) {
		cpp := tp.(TPCharger)
		cpp.Tpid = tpid
		tpCPPs = append(tpCPPs, &cpp)
	}); err != nil {
		return nil, err
	}
	return tpCPPs.AsTPChargers(), nil
}

func (csvs *CSVStorage) GetTPDispatcherProfiles(tpid, tenant, id string) ([]*utils.TPDispatcherProfile, error) {
	var tpDPPs TPDispatcherProfiles
	if err := csvs.proccesData(TPDispatcherProfile{}, csvs.dispatcherProfilesFn, func(tp interface{}) {
		dpp := tp.(TPDispatcherProfile)
		dpp.Tpid = tpid
		tpDPPs = append(tpDPPs, &dpp)
	}); err != nil {
		return nil, err
	}
	return tpDPPs.AsTPDispatcherProfiles(), nil
}

func (csvs *CSVStorage) GetTPDispatcherHosts(tpid, tenant, id string) ([]*utils.TPDispatcherHost, error) {
	var tpDDHs TPDispatcherHosts
	if err := csvs.proccesData(TPDispatcherHost{}, csvs.dispatcherHostsFn, func(tp interface{}) {
		dpp := tp.(TPDispatcherHost)
		dpp.Tpid = tpid
		tpDDHs = append(tpDDHs, &dpp)
	}); err != nil {
		return nil, err
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
