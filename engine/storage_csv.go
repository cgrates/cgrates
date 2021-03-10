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
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/sheets/v4"
)

// CSVStorage the basic csv storage
type CSVStorage struct {
	sep       rune
	generator func() csvReaderCloser
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
	routeProfilesFn          []string
	attributeProfilesFn      []string
	chargerProfilesFn        []string
	dispatcherProfilesFn     []string
	dispatcherHostsFn        []string
	rateProfilesFn           []string
	actionProfilesFn         []string
	accountProfilesFn        []string
}

// NewCSVStorage creates a CSV storege that takes the data from the paths specified
func NewCSVStorage(sep rune,
	destinationsFn, timingsFn, ratesFn, destinationratesFn,
	destinationratetimingsFn, ratingprofilesFn, sharedgroupsFn,
	actionsFn, actiontimingsFn, actiontriggersFn, accountactionsFn,
	resProfilesFn, statsFn, thresholdsFn, filterFn, routeProfilesFn,
	attributeProfilesFn, chargerProfilesFn, dispatcherProfilesFn, dispatcherHostsFn,
	rateProfilesFn, actionProfilesFn, accountProfilesFn []string) *CSVStorage {
	return &CSVStorage{
		sep:                      sep,
		generator:                NewCsvFile,
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
		routeProfilesFn:          routeProfilesFn,
		attributeProfilesFn:      attributeProfilesFn,
		chargerProfilesFn:        chargerProfilesFn,
		dispatcherProfilesFn:     dispatcherProfilesFn,
		dispatcherHostsFn:        dispatcherHostsFn,
		rateProfilesFn:           rateProfilesFn,
		actionProfilesFn:         actionProfilesFn,
		accountProfilesFn:        accountProfilesFn,
	}
}

// NewFileCSVStorage returns a csv storage that uses all files from the folder
func NewFileCSVStorage(sep rune, dataPath string) *CSVStorage {
	allFoldersPath, err := getAllFolders(dataPath)
	if err != nil {
		log.Fatal(err)
	}
	destinationsPaths := appendName(allFoldersPath, utils.DestinationsCsv)
	timingsPaths := appendName(allFoldersPath, utils.TimingsCsv)
	ratesPaths := appendName(allFoldersPath, utils.RatesCsv)
	destinationRatesPaths := appendName(allFoldersPath, utils.DestinationRatesCsv)
	ratingPlansPaths := appendName(allFoldersPath, utils.RatingPlansCsv)
	ratingProfilesPaths := appendName(allFoldersPath, utils.RatingProfilesCsv)
	sharedGroupsPaths := appendName(allFoldersPath, utils.SharedGroupsCsv)
	actionsPaths := appendName(allFoldersPath, utils.ActionsCsv)
	actionPlansPaths := appendName(allFoldersPath, utils.ActionPlansCsv)
	actionTriggersPaths := appendName(allFoldersPath, utils.ActionTriggersCsv)
	accountActionsPaths := appendName(allFoldersPath, utils.AccountActionsCsv)
	resourcesPaths := appendName(allFoldersPath, utils.ResourcesCsv)
	statsPaths := appendName(allFoldersPath, utils.StatsCsv)
	thresholdsPaths := appendName(allFoldersPath, utils.ThresholdsCsv)
	filtersPaths := appendName(allFoldersPath, utils.FiltersCsv)
	routesPaths := appendName(allFoldersPath, utils.RoutesCsv)
	attributesPaths := appendName(allFoldersPath, utils.AttributesCsv)
	chargersPaths := appendName(allFoldersPath, utils.ChargersCsv)
	dispatcherprofilesPaths := appendName(allFoldersPath, utils.DispatcherProfilesCsv)
	dispatcherhostsPaths := appendName(allFoldersPath, utils.DispatcherHostsCsv)
	rateProfilesFn := appendName(allFoldersPath, utils.RateProfilesCsv)
	actionProfilesFn := appendName(allFoldersPath, utils.ActionProfilesCsv)
	accountProfilesFn := appendName(allFoldersPath, utils.AccountProfilesCsv)
	return NewCSVStorage(sep,
		destinationsPaths,
		timingsPaths,
		ratesPaths,
		destinationRatesPaths,
		ratingPlansPaths,
		ratingProfilesPaths,
		sharedGroupsPaths,
		actionsPaths,
		actionPlansPaths,
		actionTriggersPaths,
		accountActionsPaths,
		resourcesPaths,
		statsPaths,
		thresholdsPaths,
		filtersPaths,
		routesPaths,
		attributesPaths,
		chargersPaths,
		dispatcherprofilesPaths,
		dispatcherhostsPaths,
		rateProfilesFn,
		actionProfilesFn,
		accountProfilesFn,
	)
}

// NewStringCSVStorage creates a csv storage from strings
func NewStringCSVStorage(sep rune,
	destinationsFn, timingsFn, ratesFn, destinationratesFn,
	destinationratetimingsFn, ratingprofilesFn, sharedgroupsFn,
	actionsFn, actiontimingsFn, actiontriggersFn, accountactionsFn,
	resProfilesFn, statsFn, thresholdsFn, filterFn, routeProfilesFn,
	attributeProfilesFn, chargerProfilesFn, dispatcherProfilesFn, dispatcherHostsFn,
	rateProfilesFn, actionProfilesFn, accountProfilesFn string) *CSVStorage {
	c := NewCSVStorage(sep, []string{destinationsFn}, []string{timingsFn},
		[]string{ratesFn}, []string{destinationratesFn}, []string{destinationratetimingsFn},
		[]string{ratingprofilesFn}, []string{sharedgroupsFn}, []string{actionsFn},
		[]string{actiontimingsFn}, []string{actiontriggersFn}, []string{accountactionsFn},
		[]string{resProfilesFn}, []string{statsFn}, []string{thresholdsFn}, []string{filterFn},
		[]string{routeProfilesFn}, []string{attributeProfilesFn}, []string{chargerProfilesFn},
		[]string{dispatcherProfilesFn}, []string{dispatcherHostsFn}, []string{rateProfilesFn},
		[]string{actionProfilesFn}, []string{accountProfilesFn})
	c.generator = NewCsvString
	return c
}

// NewGoogleCSVStorage creates a csv storege from google sheets
func NewGoogleCSVStorage(sep rune, spreadsheetID string) (*CSVStorage, error) {
	sht, err := newSheet()
	if err != nil {
		return nil, err
	}
	sheetNames, err := getSpreatsheetTabs(spreadsheetID, sht)
	if err != nil {
		return nil, err
	}
	getIfExist := func(name string) []string {
		if sheetNames.Has(name) {
			return []string{name}
		}
		return []string{}
	}
	c := NewCSVStorage(sep,
		getIfExist(utils.Destinations),
		getIfExist(utils.Timings),
		getIfExist(utils.Rates),
		getIfExist(utils.DestinationRates),
		getIfExist(utils.RatingPlans),
		getIfExist(utils.RatingProfiles),
		getIfExist(utils.SharedGroups),
		getIfExist(utils.Actions),
		getIfExist(utils.ActionPlans),
		getIfExist(utils.ActionTriggers),
		getIfExist(utils.AccountActions),
		getIfExist(utils.Resources),
		getIfExist(utils.Stats),
		getIfExist(utils.Thresholds),
		getIfExist(utils.Filters),
		getIfExist(utils.Routes),
		getIfExist(utils.Attributes),
		getIfExist(utils.Chargers),
		getIfExist(utils.DispatcherProfiles),
		getIfExist(utils.DispatcherHosts),
		getIfExist(utils.RateProfiles),
		getIfExist(utils.ActionProfiles),
		getIfExist(utils.AccountProfilesString))
	c.generator = func() csvReaderCloser {
		return &csvGoogle{
			spreadsheetID: spreadsheetID,
			srv:           sht,
		}
	}
	return c, nil
}

// NewURLCSVStorage returns a CSVStorage that can parse URLs
func NewURLCSVStorage(sep rune, dataPath string) *CSVStorage {
	var destinationsPaths []string
	var timingsPaths []string
	var ratesPaths []string
	var destinationRatesPaths []string
	var ratingPlansPaths []string
	var ratingProfilesPaths []string
	var sharedGroupsPaths []string
	var actionsPaths []string
	var actionPlansPaths []string
	var actionTriggersPaths []string
	var accountActionsPaths []string
	var resourcesPaths []string
	var statsPaths []string
	var thresholdsPaths []string
	var filtersPaths []string
	var routesPaths []string
	var attributesPaths []string
	var chargersPaths []string
	var dispatcherprofilesPaths []string
	var dispatcherhostsPaths []string
	var rateProfilesPaths []string
	var actionProfilesPaths []string
	var accountProfilesPaths []string

	for _, baseURL := range strings.Split(dataPath, utils.InfieldSep) {
		if !strings.HasSuffix(baseURL, utils.CSVSuffix) {
			destinationsPaths = append(destinationsPaths, joinURL(baseURL, utils.DestinationsCsv))
			timingsPaths = append(timingsPaths, joinURL(baseURL, utils.TimingsCsv))
			ratesPaths = append(ratesPaths, joinURL(baseURL, utils.RatesCsv))
			destinationRatesPaths = append(destinationRatesPaths, joinURL(baseURL, utils.DestinationRatesCsv))
			ratingPlansPaths = append(ratingPlansPaths, joinURL(baseURL, utils.RatingPlansCsv))
			ratingProfilesPaths = append(ratingProfilesPaths, joinURL(baseURL, utils.RatingProfilesCsv))
			sharedGroupsPaths = append(sharedGroupsPaths, joinURL(baseURL, utils.SharedGroupsCsv))
			actionsPaths = append(actionsPaths, joinURL(baseURL, utils.ActionsCsv))
			actionPlansPaths = append(actionPlansPaths, joinURL(baseURL, utils.ActionPlansCsv))
			actionTriggersPaths = append(actionTriggersPaths, joinURL(baseURL, utils.ActionTriggersCsv))
			accountActionsPaths = append(accountActionsPaths, joinURL(baseURL, utils.AccountActionsCsv))
			resourcesPaths = append(resourcesPaths, joinURL(baseURL, utils.ResourcesCsv))
			statsPaths = append(statsPaths, joinURL(baseURL, utils.StatsCsv))
			thresholdsPaths = append(thresholdsPaths, joinURL(baseURL, utils.ThresholdsCsv))
			filtersPaths = append(filtersPaths, joinURL(baseURL, utils.FiltersCsv))
			routesPaths = append(routesPaths, joinURL(baseURL, utils.RoutesCsv))
			attributesPaths = append(attributesPaths, joinURL(baseURL, utils.AttributesCsv))
			chargersPaths = append(chargersPaths, joinURL(baseURL, utils.ChargersCsv))
			dispatcherprofilesPaths = append(dispatcherprofilesPaths, joinURL(baseURL, utils.DispatcherProfilesCsv))
			dispatcherhostsPaths = append(dispatcherhostsPaths, joinURL(baseURL, utils.DispatcherHostsCsv))
			rateProfilesPaths = append(rateProfilesPaths, joinURL(baseURL, utils.RateProfilesCsv))
			actionProfilesPaths = append(actionProfilesPaths, joinURL(baseURL, utils.ActionProfilesCsv))
			accountProfilesPaths = append(accountProfilesPaths, joinURL(baseURL, utils.AccountProfilesCsv))
			continue
		}
		switch {
		case strings.HasSuffix(baseURL, utils.DestinationsCsv):
			destinationsPaths = append(destinationsPaths, baseURL)
		case strings.HasSuffix(baseURL, utils.TimingsCsv):
			timingsPaths = append(timingsPaths, baseURL)
		case strings.HasSuffix(baseURL, utils.RatesCsv):
			ratesPaths = append(ratesPaths, baseURL)
		case strings.HasSuffix(baseURL, utils.DestinationRatesCsv):
			destinationRatesPaths = append(destinationRatesPaths, baseURL)
		case strings.HasSuffix(baseURL, utils.RatingPlansCsv):
			ratingPlansPaths = append(ratingPlansPaths, baseURL)
		case strings.HasSuffix(baseURL, utils.RatingProfilesCsv):
			ratingProfilesPaths = append(ratingProfilesPaths, baseURL)
		case strings.HasSuffix(baseURL, utils.SharedGroupsCsv):
			sharedGroupsPaths = append(sharedGroupsPaths, baseURL)
		case strings.HasSuffix(baseURL, utils.ActionsCsv):
			actionsPaths = append(actionsPaths, baseURL)
		case strings.HasSuffix(baseURL, utils.ActionPlansCsv):
			actionPlansPaths = append(actionPlansPaths, baseURL)
		case strings.HasSuffix(baseURL, utils.ActionTriggersCsv):
			actionTriggersPaths = append(actionTriggersPaths, baseURL)
		case strings.HasSuffix(baseURL, utils.AccountActionsCsv):
			accountActionsPaths = append(accountActionsPaths, baseURL)
		case strings.HasSuffix(baseURL, utils.ResourcesCsv):
			resourcesPaths = append(resourcesPaths, baseURL)
		case strings.HasSuffix(baseURL, utils.StatsCsv):
			statsPaths = append(statsPaths, baseURL)
		case strings.HasSuffix(baseURL, utils.ThresholdsCsv):
			thresholdsPaths = append(thresholdsPaths, baseURL)
		case strings.HasSuffix(baseURL, utils.FiltersCsv):
			filtersPaths = append(filtersPaths, baseURL)
		case strings.HasSuffix(baseURL, utils.RoutesCsv):
			routesPaths = append(routesPaths, baseURL)
		case strings.HasSuffix(baseURL, utils.AttributesCsv):
			attributesPaths = append(attributesPaths, baseURL)
		case strings.HasSuffix(baseURL, utils.ChargersCsv):
			chargersPaths = append(chargersPaths, baseURL)
		case strings.HasSuffix(baseURL, utils.DispatcherProfilesCsv):
			dispatcherprofilesPaths = append(dispatcherprofilesPaths, baseURL)
		case strings.HasSuffix(baseURL, utils.DispatcherHostsCsv):
			dispatcherhostsPaths = append(dispatcherhostsPaths, baseURL)
		case strings.HasSuffix(baseURL, utils.RateProfilesCsv):
			rateProfilesPaths = append(rateProfilesPaths, baseURL)
		case strings.HasSuffix(baseURL, utils.ActionProfilesCsv):
			actionProfilesPaths = append(actionProfilesPaths, baseURL)
		case strings.HasSuffix(baseURL, utils.AccountProfilesCsv):
			accountProfilesPaths = append(accountProfilesPaths, baseURL)

		}
	}

	c := NewCSVStorage(sep,
		destinationsPaths,
		timingsPaths,
		ratesPaths,
		destinationRatesPaths,
		ratingPlansPaths,
		ratingProfilesPaths,
		sharedGroupsPaths,
		actionsPaths,
		actionPlansPaths,
		actionTriggersPaths,
		accountActionsPaths,
		resourcesPaths,
		statsPaths,
		thresholdsPaths,
		filtersPaths,
		routesPaths,
		attributesPaths,
		chargersPaths,
		dispatcherprofilesPaths,
		dispatcherhostsPaths,
		rateProfilesPaths,
		actionProfilesPaths,
		accountProfilesPaths,
	)
	c.generator = func() csvReaderCloser {
		return &csvURL{}
	}
	return c
}

func joinURL(baseURL, fn string) string {
	u, err := url.Parse(baseURL)
	if err != nil {
		return baseURL + fn
	}
	u.Path = path.Join(u.Path, fn)
	return u.String()
}

func getAllFolders(inPath string) (paths []string, err error) {
	err = filepath.Walk(inPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			paths = append(paths, path)
		}
		return nil
	})
	return
}

func appendName(paths []string, fileName string) (out []string) {
	out = make([]string, len(paths))
	for i, basePath := range paths {
		out[i] = path.Join(basePath, fileName)
	}
	return
}

func (csvs *CSVStorage) proccesData(listType interface{}, fns []string, process func(interface{})) error {
	collumnCount := getColumnCount(listType)
	for _, fileName := range fns {
		csvReader := csvs.generator()
		err := csvReader.Open(fileName, csvs.sep, collumnCount)
		if err != nil {
			// maybe a log to view if failed to open file
			continue // try read the rest
		}
		if err = func() error { // to execute defer corectly
			defer csvReader.Close()
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
	var tpTimings TimingMdls
	if err := csvs.proccesData(TimingMdl{}, csvs.timingsFn, func(tp interface{}) {
		tm := tp.(TimingMdl)
		tm.Tpid = tpid
		tpTimings = append(tpTimings, tm)
	}); err != nil {
		return nil, err
	}
	return tpTimings.AsTPTimings(), nil
}

func (csvs *CSVStorage) GetTPDestinations(tpid, id string) ([]*utils.TPDestination, error) {
	var tpDests DestinationMdls
	if err := csvs.proccesData(DestinationMdl{}, csvs.destinationsFn, func(tp interface{}) {
		d := tp.(DestinationMdl)
		d.Tpid = tpid
		tpDests = append(tpDests, d)
	}); err != nil {
		return nil, err
	}
	return tpDests.AsTPDestinations(), nil
}

func (csvs *CSVStorage) GetTPRates(tpid, id string) ([]*utils.TPRateRALs, error) {
	var tpRates RateMdls
	if err := csvs.proccesData(RateMdl{}, csvs.ratesFn, func(tp interface{}) {
		r := tp.(RateMdl)
		r.Tpid = tpid
		tpRates = append(tpRates, r)
	}); err != nil {
		return nil, err
	}
	return tpRates.AsTPRates()
}

func (csvs *CSVStorage) GetTPDestinationRates(tpid, id string, p *utils.Paginator) ([]*utils.TPDestinationRate, error) {
	var tpDestinationRates DestinationRateMdls
	if err := csvs.proccesData(DestinationRateMdl{}, csvs.destinationratesFn, func(tp interface{}) {
		dr := tp.(DestinationRateMdl)
		dr.Tpid = tpid
		tpDestinationRates = append(tpDestinationRates, dr)
	}); err != nil {
		return nil, err
	}
	return tpDestinationRates.AsTPDestinationRates(), nil
}

func (csvs *CSVStorage) GetTPRatingPlans(tpid, id string, p *utils.Paginator) ([]*utils.TPRatingPlan, error) {
	var tpRatingPlans RatingPlanMdls
	if err := csvs.proccesData(RatingPlanMdl{}, csvs.destinationratetimingsFn, func(tp interface{}) {
		rp := tp.(RatingPlanMdl)
		rp.Tpid = tpid
		tpRatingPlans = append(tpRatingPlans, rp)
	}); err != nil {
		return nil, err
	}
	return tpRatingPlans.AsTPRatingPlans(), nil
}

func (csvs *CSVStorage) GetTPRatingProfiles(filter *utils.TPRatingProfile) ([]*utils.TPRatingProfile, error) {
	var tpRatingProfiles RatingProfileMdls
	if err := csvs.proccesData(RatingProfileMdl{}, csvs.ratingprofilesFn, func(tp interface{}) {
		rpf := tp.(RatingProfileMdl)
		if filter != nil {
			rpf.Tpid = filter.TPid
			rpf.Loadid = filter.LoadId
		}
		tpRatingProfiles = append(tpRatingProfiles, rpf)
	}); err != nil {
		return nil, err
	}
	return tpRatingProfiles.AsTPRatingProfiles(), nil
}

func (csvs *CSVStorage) GetTPSharedGroups(tpid, id string) ([]*utils.TPSharedGroups, error) {
	var tpSharedGroups SharedGroupMdls
	if err := csvs.proccesData(SharedGroupMdl{}, csvs.sharedgroupsFn, func(tp interface{}) {
		sg := tp.(SharedGroupMdl)
		sg.Tpid = tpid
		tpSharedGroups = append(tpSharedGroups, sg)
	}); err != nil {
		return nil, err
	}
	return tpSharedGroups.AsTPSharedGroups(), nil
}

func (csvs *CSVStorage) GetTPActions(tpid, id string) ([]*utils.TPActions, error) {
	var tpActions ActionMdls
	if err := csvs.proccesData(ActionMdl{}, csvs.actionsFn, func(tp interface{}) {
		a := tp.(ActionMdl)
		a.Tpid = tpid
		tpActions = append(tpActions, a)
	}); err != nil {
		return nil, err
	}
	return tpActions.AsTPActions(), nil
}

func (csvs *CSVStorage) GetTPActionPlans(tpid, id string) ([]*utils.TPActionPlan, error) {
	var tpActionPlans ActionPlanMdls
	if err := csvs.proccesData(ActionPlanMdl{}, csvs.actiontimingsFn, func(tp interface{}) {
		ap := tp.(ActionPlanMdl)
		ap.Tpid = tpid
		tpActionPlans = append(tpActionPlans, ap)
	}); err != nil {
		return nil, err
	}
	return tpActionPlans.AsTPActionPlans(), nil
}

func (csvs *CSVStorage) GetTPActionTriggers(tpid, id string) ([]*utils.TPActionTriggers, error) {
	var tpActionTriggers ActionTriggerMdls
	if err := csvs.proccesData(ActionTriggerMdl{}, csvs.actiontriggersFn, func(tp interface{}) {
		at := tp.(ActionTriggerMdl)
		at.Tpid = tpid
		tpActionTriggers = append(tpActionTriggers, at)
	}); err != nil {
		return nil, err
	}
	return tpActionTriggers.AsTPActionTriggers(), nil
}

func (csvs *CSVStorage) GetTPAccountActions(filter *utils.TPAccountActions) ([]*utils.TPAccountActions, error) {
	var tpAccountActions AccountActionMdls
	if err := csvs.proccesData(AccountActionMdl{}, csvs.accountactionsFn, func(tp interface{}) {
		aa := tp.(AccountActionMdl)
		if filter != nil {
			aa.Tpid = filter.TPid
			aa.Loadid = filter.LoadId
		}
		tpAccountActions = append(tpAccountActions, aa)
	}); err != nil {
		return nil, err
	}
	return tpAccountActions.AsTPAccountActions(), nil
}

func (csvs *CSVStorage) GetTPResources(tpid, tenant, id string) ([]*utils.TPResourceProfile, error) {
	var tpResLimits ResourceMdls
	if err := csvs.proccesData(ResourceMdl{}, csvs.resProfilesFn, func(tp interface{}) {
		tpLimit := tp.(ResourceMdl)
		tpLimit.Tpid = tpid
		tpResLimits = append(tpResLimits, &tpLimit)
	}); err != nil {
		return nil, err
	}
	return tpResLimits.AsTPResources(), nil
}

func (csvs *CSVStorage) GetTPStats(tpid, tenant, id string) ([]*utils.TPStatProfile, error) {
	var tpStats StatMdls
	if err := csvs.proccesData(StatMdl{}, csvs.statsFn, func(tp interface{}) {
		tPstats := tp.(StatMdl)
		tPstats.Tpid = tpid
		tpStats = append(tpStats, &tPstats)
	}); err != nil {
		return nil, err
	}
	return tpStats.AsTPStats(), nil
}

func (csvs *CSVStorage) GetTPThresholds(tpid, tenant, id string) ([]*utils.TPThresholdProfile, error) {
	var tpThreshold ThresholdMdls
	if err := csvs.proccesData(ThresholdMdl{}, csvs.thresholdsFn, func(tp interface{}) {
		tHresholdCfg := tp.(ThresholdMdl)
		tHresholdCfg.Tpid = tpid
		tpThreshold = append(tpThreshold, &tHresholdCfg)
	}); err != nil {
		return nil, err
	}
	return tpThreshold.AsTPThreshold(), nil
}

func (csvs *CSVStorage) GetTPFilters(tpid, tenant, id string) ([]*utils.TPFilterProfile, error) {
	var tpFilter FilterMdls
	if err := csvs.proccesData(FilterMdl{}, csvs.filterFn, func(tp interface{}) {
		fIlterCfg := tp.(FilterMdl)
		fIlterCfg.Tpid = tpid
		tpFilter = append(tpFilter, &fIlterCfg)
	}); err != nil {
		return nil, err
	}
	return tpFilter.AsTPFilter(), nil
}

func (csvs *CSVStorage) GetTPRoutes(tpid, tenant, id string) ([]*utils.TPRouteProfile, error) {
	var tpRoutes RouteMdls
	if err := csvs.proccesData(RouteMdl{}, csvs.routeProfilesFn, func(tp interface{}) {
		suppProfile := tp.(RouteMdl)
		suppProfile.Tpid = tpid
		tpRoutes = append(tpRoutes, &suppProfile)
	}); err != nil {
		return nil, err
	}
	return tpRoutes.AsTPRouteProfile(), nil
}

func (csvs *CSVStorage) GetTPAttributes(tpid, tenant, id string) ([]*utils.TPAttributeProfile, error) {
	var tpAls AttributeMdls
	if err := csvs.proccesData(AttributeMdl{}, csvs.attributeProfilesFn, func(tp interface{}) {
		attributeProfile := tp.(AttributeMdl)
		attributeProfile.Tpid = tpid
		tpAls = append(tpAls, &attributeProfile)
	}); err != nil {
		return nil, err
	}
	return tpAls.AsTPAttributes(), nil
}

func (csvs *CSVStorage) GetTPChargers(tpid, tenant, id string) ([]*utils.TPChargerProfile, error) {
	var tpCPPs ChargerMdls
	if err := csvs.proccesData(ChargerMdl{}, csvs.chargerProfilesFn, func(tp interface{}) {
		cpp := tp.(ChargerMdl)
		cpp.Tpid = tpid
		tpCPPs = append(tpCPPs, &cpp)
	}); err != nil {
		return nil, err
	}
	return tpCPPs.AsTPChargers(), nil
}

func (csvs *CSVStorage) GetTPDispatcherProfiles(tpid, tenant, id string) ([]*utils.TPDispatcherProfile, error) {
	var tpDPPs DispatcherProfileMdls
	if err := csvs.proccesData(DispatcherProfileMdl{}, csvs.dispatcherProfilesFn, func(tp interface{}) {
		dpp := tp.(DispatcherProfileMdl)
		dpp.Tpid = tpid
		tpDPPs = append(tpDPPs, &dpp)
	}); err != nil {
		return nil, err
	}
	return tpDPPs.AsTPDispatcherProfiles(), nil
}

func (csvs *CSVStorage) GetTPDispatcherHosts(tpid, tenant, id string) ([]*utils.TPDispatcherHost, error) {
	var tpDDHs DispatcherHostMdls
	if err := csvs.proccesData(DispatcherHostMdl{}, csvs.dispatcherHostsFn, func(tp interface{}) {
		dpp := tp.(DispatcherHostMdl)
		dpp.Tpid = tpid
		tpDDHs = append(tpDDHs, &dpp)
	}); err != nil {
		return nil, err
	}
	return tpDDHs.AsTPDispatcherHosts(), nil
}

func (csvs *CSVStorage) GetTPRateProfiles(tpid, tenant, id string) ([]*utils.TPRateProfile, error) {
	var tpDPPs RateProfileMdls
	if err := csvs.proccesData(RateProfileMdl{}, csvs.rateProfilesFn, func(tp interface{}) {
		dpp := tp.(RateProfileMdl)
		dpp.Tpid = tpid
		tpDPPs = append(tpDPPs, &dpp)
	}); err != nil {
		return nil, err
	}
	return tpDPPs.AsTPRateProfile(), nil
}

func (csvs *CSVStorage) GetTPActionProfiles(tpid, tenant, id string) ([]*utils.TPActionProfile, error) {
	var tpDPPs ActionProfileMdls
	if err := csvs.proccesData(ActionProfileMdl{}, csvs.actionProfilesFn, func(tp interface{}) {
		dpp := tp.(ActionProfileMdl)
		dpp.Tpid = tpid
		tpDPPs = append(tpDPPs, &dpp)
	}); err != nil {
		return nil, err
	}
	return tpDPPs.AsTPActionProfile(), nil
}

func (csvs *CSVStorage) GetTPAccountProfiles(tpid, tenant, id string) ([]*utils.TPAccountProfile, error) {
	var tpDPPs AccountProfileMdls
	if err := csvs.proccesData(AccountProfileMdl{}, csvs.accountProfilesFn, func(tp interface{}) {
		dpp := tp.(AccountProfileMdl)
		dpp.Tpid = tpid
		tpDPPs = append(tpDPPs, &dpp)
	}); err != nil {
		return nil, err
	}
	return tpDPPs.AsTPAccountProfile()
}

func (csvs *CSVStorage) GetTpIds(colName string) ([]string, error) {
	return nil, utils.ErrNotImplemented
}

func (csvs *CSVStorage) GetTpTableIds(tpid, table string,
	distinct utils.TPDistinctIds, filters map[string]string, p *utils.PaginatorWithSearch) ([]string, error) {
	return nil, utils.ErrNotImplemented
}

type csvReaderCloser interface {
	Open(data string, sep rune, nrFields int) (err error)
	Read() (record []string, err error)
	Close()
}

func NewCsvFile() csvReaderCloser {
	return &csvFile{}
}

type csvFile struct {
	csvReader *csv.Reader
	fp        *os.File
}

func (c *csvFile) Open(fn string, sep rune, nrFields int) (err error) {
	c.fp, err = os.Open(fn)
	if err != nil {
		return
	}
	c.csvReader = csv.NewReader(c.fp)
	c.csvReader.Comma = sep
	c.csvReader.Comment = utils.CommentChar
	c.csvReader.FieldsPerRecord = nrFields
	c.csvReader.TrailingComma = true
	return
}

func (c *csvFile) Read() (record []string, err error) {
	return c.csvReader.Read()
}

func (c *csvFile) Close() {
	if c.fp != nil {
		c.fp.Close()
	}
}

func NewCsvString() csvReaderCloser {
	return &csvString{}
}

type csvString struct {
	csvReader *csv.Reader
}

func (c *csvString) Open(data string, sep rune, nrFields int) (err error) {
	c.csvReader = csv.NewReader(strings.NewReader(data))
	c.csvReader.Comma = sep
	c.csvReader.Comment = utils.CommentChar
	c.csvReader.FieldsPerRecord = nrFields
	c.csvReader.TrailingComma = true
	return
}

func (c *csvString) Read() (record []string, err error) {
	return c.csvReader.Read()
}

func (c *csvString) Close() { // no need for close
}

// Google

// Retrieve a token, saves the token, then returns the generated client.
func getClient(cfg *oauth2.Config, configPath string) (*http.Client, error) {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tok := &oauth2.Token{}
	raw, err := getCfgJSONData(config.CgrConfig().LoaderCgrCfg().GapiToken)
	if err != nil {
		tok, err = getTokenFromWeb(cfg)
		if err != nil {
			return nil, err
		}
		path2TokFileb := config.CgrConfig().LoaderCgrCfg().GapiToken
		path2TokFile := string(path2TokFileb[1 : len(path2TokFileb)-1])
		if err := os.MkdirAll(filepath.Dir(path2TokFile), os.FileMode(0777)); err != nil { // create the directory if not exists
			return nil, err
		}
		saveToken(path2TokFile, tok)
	} else if err = json.Unmarshal(raw, tok); err != nil {
		return nil, err
	}
	return cfg.Client(context.Background(), tok), nil
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) (*oauth2.Token, error) {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		err = fmt.Errorf("Unable to read authorization code: %v", err)
		return nil, err
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		err = fmt.Errorf("Unable to retrieve token from web: %v", err)
	}
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Printf("Unable to cache oauth token: %v\n", err)
		return
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func getCfgJSONData(raw json.RawMessage) (data []byte, err error) {
	if len(raw) == 0 {
		return
	}
	if raw[0] == '{' && raw[len(raw)-1] == '}' {
		data = raw
		return
	}
	dataPath := string(raw[1 : len(raw)-1])
	if !strings.HasSuffix(dataPath, utils.JSNSuffix) {
		dataPath = path.Join(dataPath, utils.GoogleCredentialsFileName)
	}
	return ioutil.ReadFile(dataPath)
}

func newSheet() (sht *sheets.Service, err error) { //*google_api
	var cred []byte
	var cfgPathDir string
	if cred, err = getCfgJSONData(config.CgrConfig().LoaderCgrCfg().GapiCredentials); err != nil {
		err = fmt.Errorf("Unable to read client secret file: %v", err)
		return
	}
	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(cred, "https://www.googleapis.com/auth/spreadsheets.readonly")
	if err != nil {
		err = fmt.Errorf("Unable to parse client secret file to config: %v", err)
		return
	}
	client, err := getClient(config, cfgPathDir)
	if err != nil {
		return nil, err
	}
	sht, err = sheets.New(client)
	if err != nil {
		err = fmt.Errorf("Unable to retrieve Sheets client: %v", err)
	}
	return
}

func getSpreatsheetTabs(spreadsheetID string, srv *sheets.Service) (sheetsName utils.StringSet, err error) {
	sheetsName = make(utils.StringSet)
	sht, err := srv.Spreadsheets.Get(spreadsheetID).Do()
	if err != nil {
		err = fmt.Errorf("Unable get the information about spreadsheet because: %v", err)
		return
	}
	for _, sheet := range sht.Sheets {
		sheetsName.Add(sheet.Properties.Title)
	}
	return
}

type csvGoogle struct {
	spreadsheetID string
	srv           *sheets.Service
	response      *sheets.ValueRange
	indx          int
	nrFields      int
}

func (c *csvGoogle) Open(data string, sep rune, nrFields int) (err error) {
	c.response, err = c.srv.Spreadsheets.Values.Get(c.spreadsheetID, data).Do()
	if err != nil {
		return
	}

	if len(c.response.Values) == 0 {
		return utils.ErrNotFound
	}
	c.nrFields = nrFields
	return
}

func (c *csvGoogle) getNextRow() (row []interface{}, err error) {
	if len(c.response.Values) <= c.indx {
		return nil, io.EOF
	}
	row = c.response.Values[c.indx]
	c.indx++
	if len(row) == 0 {
		return c.getNextRow()
	}
	return
}

func (c *csvGoogle) Read() (record []string, err error) {
	row, err := c.getNextRow()
	if err != nil {
		return
	}
	record = make([]string, c.nrFields)
	for i := 0; i < c.nrFields; i++ {
		if i < len(row) {
			record[i] = utils.IfaceAsString(row[i])
			if i == 0 && strings.HasPrefix(record[i], "#") {
				return c.Read() // ignore row if starts with #
			}
		} else {
			record[i] = ""
		}
	}
	return
}

func (c *csvGoogle) Close() { // no need for close
}

type csvURL struct {
	csvReader *csv.Reader
	page      io.ReadCloser
}

func (c *csvURL) Open(fn string, sep rune, nrFields int) (err error) {
	if _, err = url.ParseRequestURI(fn); err != nil {
		return
	}
	var myClient = &http.Client{
		Timeout: config.CgrConfig().GeneralCfg().ReplyTimeout,
	}
	var req *http.Response
	req, err = myClient.Get(fn)
	if err != nil {
		return utils.ErrPathNotReachable(fn)
	}
	if req.StatusCode != http.StatusOK {
		return utils.ErrNotFound
	}
	c.page = req.Body

	c.csvReader = csv.NewReader(c.page)
	c.csvReader.Comma = sep
	c.csvReader.Comment = utils.CommentChar
	c.csvReader.FieldsPerRecord = nrFields
	c.csvReader.TrailingComma = true
	return
}

func (c *csvURL) Read() (record []string, err error) {
	return c.csvReader.Read()
}

func (c *csvURL) Close() {
	if c.page != nil {
		c.page.Close()
	}
}
