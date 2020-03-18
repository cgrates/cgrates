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
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/cgrates/cgrates/utils"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/sheets/v4"
)

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
	suppProfilesFn           []string
	attributeProfilesFn      []string
	chargerProfilesFn        []string
	dispatcherProfilesFn     []string
	dispatcherHostsFn        []string
}

func NewCSVStorage(sep rune,
	destinationsFn, timingsFn, ratesFn, destinationratesFn,
	destinationratetimingsFn, ratingprofilesFn, sharedgroupsFn,
	actionsFn, actiontimingsFn, actiontriggersFn, accountactionsFn,
	resProfilesFn, statsFn, thresholdsFn,
	filterFn, suppProfilesFn, attributeProfilesFn,
	chargerProfilesFn, dispatcherProfilesFn, dispatcherHostsFn []string) *CSVStorage {
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
		suppProfilesFn:           suppProfilesFn,
		attributeProfilesFn:      attributeProfilesFn,
		chargerProfilesFn:        chargerProfilesFn,
		dispatcherProfilesFn:     dispatcherProfilesFn,
		dispatcherHostsFn:        dispatcherHostsFn,
	}
}

func NewFileCSVStorage(sep rune, dataPath string) *CSVStorage {
	allFoldersPath, err := getAllFolders(dataPath)
	if err != nil {
		log.Fatal(err)
	}
	destinations_paths := appendName(allFoldersPath, utils.DestinationsCsv)
	timings_paths := appendName(allFoldersPath, utils.TimingsCsv)
	rates_paths := appendName(allFoldersPath, utils.RatesCsv)
	destination_rates_paths := appendName(allFoldersPath, utils.DestinationRatesCsv)
	rating_plans_paths := appendName(allFoldersPath, utils.RatingPlansCsv)
	rating_profiles_paths := appendName(allFoldersPath, utils.RatingProfilesCsv)
	shared_groups_paths := appendName(allFoldersPath, utils.SharedGroupsCsv)
	actions_paths := appendName(allFoldersPath, utils.ActionsCsv)
	action_plans_paths := appendName(allFoldersPath, utils.ActionPlansCsv)
	action_triggers_paths := appendName(allFoldersPath, utils.ActionTriggersCsv)
	account_actions_paths := appendName(allFoldersPath, utils.AccountActionsCsv)
	resources_paths := appendName(allFoldersPath, utils.ResourcesCsv)
	stats_paths := appendName(allFoldersPath, utils.StatsCsv)
	thresholds_paths := appendName(allFoldersPath, utils.ThresholdsCsv)
	filters_paths := appendName(allFoldersPath, utils.FiltersCsv)
	suppliers_paths := appendName(allFoldersPath, utils.SuppliersCsv)
	attributes_paths := appendName(allFoldersPath, utils.AttributesCsv)
	chargers_paths := appendName(allFoldersPath, utils.ChargersCsv)
	dispatcherprofiles_paths := appendName(allFoldersPath, utils.DispatcherProfilesCsv)
	dispatcherhosts_paths := appendName(allFoldersPath, utils.DispatcherHostsCsv)
	return NewCSVStorage(sep,
		destinations_paths,
		timings_paths,
		rates_paths,
		destination_rates_paths,
		rating_plans_paths,
		rating_profiles_paths,
		shared_groups_paths,
		actions_paths,
		action_plans_paths,
		action_triggers_paths,
		account_actions_paths,
		resources_paths,
		stats_paths,
		thresholds_paths,
		filters_paths,
		suppliers_paths,
		attributes_paths,
		chargers_paths,
		dispatcherprofiles_paths,
		dispatcherhosts_paths,
	)
}

func NewStringCSVStorage(sep rune,
	destinationsFn, timingsFn, ratesFn, destinationratesFn,
	destinationratetimingsFn, ratingprofilesFn, sharedgroupsFn,
	actionsFn, actiontimingsFn, actiontriggersFn,
	accountactionsFn, resProfilesFn, statsFn,
	thresholdsFn, filterFn, suppProfilesFn,
	attributeProfilesFn, chargerProfilesFn,
	dispatcherProfilesFn, dispatcherHostsFn string) *CSVStorage {
	c := NewCSVStorage(sep, []string{destinationsFn}, []string{timingsFn},
		[]string{ratesFn}, []string{destinationratesFn}, []string{destinationratetimingsFn},
		[]string{ratingprofilesFn}, []string{sharedgroupsFn}, []string{actionsFn},
		[]string{actiontimingsFn}, []string{actiontriggersFn}, []string{accountactionsFn},
		[]string{resProfilesFn}, []string{statsFn}, []string{thresholdsFn}, []string{filterFn},
		[]string{suppProfilesFn}, []string{attributeProfilesFn}, []string{chargerProfilesFn},
		[]string{dispatcherProfilesFn}, []string{dispatcherHostsFn})
	c.generator = NewCsvString
	return c
}

func NewGoogleCSVStorage(sep rune, spreadsheetId, cfgPath string) (*CSVStorage, error) {
	sht, err := newSheet(cfgPath)
	if err != nil {
		return nil, err
	}
	sheetNames, err := getSpreatsheetTabs(spreadsheetId, sht)
	if err != nil {
		return nil, err
	}
	getIfExist := func(name string) []string {
		if _, has := sheetNames[name]; has {
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
		getIfExist(utils.Suppliers),
		getIfExist(utils.Attributes),
		getIfExist(utils.Chargers),
		getIfExist(utils.DispatcherProfiles),
		getIfExist(utils.DispatcherHosts))
	c.generator = func() csvReaderCloser {
		return &csvGoogle{
			spreadsheetId: spreadsheetId,
			srv:           sht,
		}
	}
	return c, nil
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
	for i, path_ := range paths {
		out[i] = path.Join(path_, fileName)
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
	c.csvReader.Comment = utils.COMMENT_CHAR
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
	c.csvReader.Comment = utils.COMMENT_CHAR
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
func getClient(config *oauth2.Config, configPath string) (*http.Client, error) {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := path.Join(configPath, utils.GoogleConfigDirName, utils.GoogleTokenFileName)
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok, err = getTokenFromWeb(config)
		if err != nil {
			return nil, err
		}
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok), nil
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

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
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

func newSheet(configPath string) (sht *sheets.Service, err error) { //*google_api
	b, err := ioutil.ReadFile(path.Join(configPath, utils.GoogleConfigDirName, utils.GoogleCredentialsFileName))
	if err != nil {
		err = fmt.Errorf("Unable to read client secret file: %v", err)
		return
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/spreadsheets.readonly")
	if err != nil {
		err = fmt.Errorf("Unable to parse client secret file to config: %v", err)
		return
	}
	client, err := getClient(config, configPath)
	if err != nil {
		return nil, err
	}
	sht, err = sheets.New(client)
	if err != nil {
		err = fmt.Errorf("Unable to retrieve Sheets client: %v", err)
	}
	return
}

func getSpreatsheetTabs(spreadsheetId string, srv *sheets.Service) (sheetsName map[string]struct{}, err error) {
	sheetsName = make(map[string]struct{})
	sht, err := srv.Spreadsheets.Get(spreadsheetId).Do()
	if err != nil {
		err = fmt.Errorf("Unable get the information about spreadsheet because: %v", err)
		return
	}
	for _, sheet := range sht.Sheets {
		sheetsName[sheet.Properties.Title] = struct{}{}
	}
	return
}

type csvGoogle struct {
	spreadsheetId string
	srv           *sheets.Service
	response      *sheets.ValueRange
	indx          int
	nrFields      int
}

func (c *csvGoogle) Open(data string, sep rune, nrFields int) (err error) {
	c.response, err = c.srv.Spreadsheets.Values.Get(c.spreadsheetId, data).Do()
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
