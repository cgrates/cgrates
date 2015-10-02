/*
Real-time Charging System for Telecom & ISP environments
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

package cdrc

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"gopkg.in/fsnotify.v1"
)

const (
	CSV             = "csv"
	FS_CSV          = "freeswitch_csv"
	UNPAIRED_SUFFIX = ".unpaired"
)

// Populates the
func populateStoredCdrField(cdr *engine.StoredCdr, fieldId, fieldVal, timezone string) error {
	var err error
	switch fieldId {
	case utils.TOR:
		cdr.TOR += fieldVal
	case utils.ACCID:
		cdr.AccId += fieldVal
	case utils.REQTYPE:
		cdr.ReqType += fieldVal
	case utils.DIRECTION:
		cdr.Direction += fieldVal
	case utils.TENANT:
		cdr.Tenant += fieldVal
	case utils.CATEGORY:
		cdr.Category += fieldVal
	case utils.ACCOUNT:
		cdr.Account += fieldVal
	case utils.SUBJECT:
		cdr.Subject += fieldVal
	case utils.DESTINATION:
		cdr.Destination += fieldVal
	case utils.RATED_FLD:
		cdr.Rated, _ = strconv.ParseBool(fieldVal)
	case utils.SETUP_TIME:
		if cdr.SetupTime, err = utils.ParseTimeDetectLayout(fieldVal, timezone); err != nil {
			return fmt.Errorf("Cannot parse answer time field with value: %s, err: %s", fieldVal, err.Error())
		}
	case utils.PDD:
		if cdr.Pdd, err = utils.ParseDurationWithSecs(fieldVal); err != nil {
			return fmt.Errorf("Cannot parse answer time field with value: %s, err: %s", fieldVal, err.Error())
		}
	case utils.ANSWER_TIME:
		if cdr.AnswerTime, err = utils.ParseTimeDetectLayout(fieldVal, timezone); err != nil {
			return fmt.Errorf("Cannot parse answer time field with value: %s, err: %s", fieldVal, err.Error())
		}
	case utils.USAGE:
		if cdr.Usage, err = utils.ParseDurationWithSecs(fieldVal); err != nil {
			return fmt.Errorf("Cannot parse duration field with value: %s, err: %s", fieldVal, err.Error())
		}
	case utils.SUPPLIER:
		cdr.Supplier += fieldVal
	case utils.DISCONNECT_CAUSE:
		cdr.DisconnectCause += fieldVal
	case utils.COST:
		if cdr.Cost, err = strconv.ParseFloat(fieldVal, 64); err != nil {
			return fmt.Errorf("Cannot parse cost field with value: %s, err: %s", fieldVal, err.Error())
		}
	default: // Extra fields will not match predefined so they all show up here
		cdr.ExtraFields[fieldId] += fieldVal
	}
	return nil
}

// Understands and processes a specific format of cdr (eg: .csv or .fwv)
type RecordsProcessor interface {
	ProcessNextRecord() ([]*engine.StoredCdr, error) // Process a single record in the CDR file, return a slice of CDRs since based on configuration we can have more templates
	ProcessedRecordsNr() int64
}

/*
One instance  of CDRC will act on one folder.
Common parameters within configs processed:
 * cdrS, cdrFormat, cdrInDir, cdrOutDir, runDelay
Parameters specific per config instance:
 * duMultiplyFactor, cdrSourceId, cdrFilter, cdrFields
*/
func NewCdrc(cdrcCfgs map[string]*config.CdrcConfig, httpSkipTlsCheck bool, cdrs engine.Connector, closeChan chan struct{}, dfltTimezone string) (*Cdrc, error) {
	var cdrcCfg *config.CdrcConfig
	for _, cdrcCfg = range cdrcCfgs { // Take the first config out, does not matter which one
		break
	}
	cdrc := &Cdrc{cdrFormat: cdrcCfg.CdrFormat, cdrInDir: cdrcCfg.CdrInDir, cdrOutDir: cdrcCfg.CdrOutDir,
		runDelay: cdrcCfg.RunDelay, csvSep: cdrcCfg.FieldSeparator,
		httpSkipTlsCheck: httpSkipTlsCheck, cdrcCfgs: cdrcCfgs, dfltCdrcCfg: cdrcCfg, timezone: utils.FirstNonEmpty(cdrcCfg.Timezone, dfltTimezone), cdrs: cdrs,
		closeChan: closeChan, maxOpenFiles: make(chan struct{}, cdrcCfg.MaxOpenFiles),
	}
	var processFile struct{}
	for i := 0; i < cdrcCfg.MaxOpenFiles; i++ {
		cdrc.maxOpenFiles <- processFile // Empty initiate so we do not need to wait later when we pop
	}
	cdrc.cdrSourceIds = make([]string, len(cdrcCfgs))
	cdrc.duMultiplyFactors = make([]float64, len(cdrcCfgs))
	cdrc.cdrFilters = make([]utils.RSRFields, len(cdrcCfgs))
	cdrc.cdrFields = make([][]*config.CfgCdrField, len(cdrcCfgs))
	idx := 0
	var err error
	for _, cfg := range cdrcCfgs {
		if idx == 0 { // Steal the config from just one instance since it should be the same for all
			cdrc.failedCallsPrefix = cfg.FailedCallsPrefix
			if cdrc.partialRecordsCache, err = NewPartialRecordsCache(cdrcCfg.PartialRecordCache, cdrcCfg.CdrOutDir, cdrcCfg.FieldSeparator); err != nil {
				return nil, err
			}
		}
		cdrc.cdrSourceIds[idx] = cfg.CdrSourceId
		cdrc.duMultiplyFactors[idx] = cfg.DataUsageMultiplyFactor
		cdrc.cdrFilters[idx] = cfg.CdrFilter
		cdrc.cdrFields[idx] = cfg.ContentFields
		idx += 1
	}
	// Before processing, make sure in and out folders exist
	for _, dir := range []string{cdrc.cdrInDir, cdrc.cdrOutDir} {
		if _, err := os.Stat(dir); err != nil && os.IsNotExist(err) {
			return nil, fmt.Errorf("Nonexistent folder: %s", dir)
		}
	}
	cdrc.httpClient = new(http.Client)
	return cdrc, nil
}

type Cdrc struct {
	cdrFormat,
	cdrInDir,
	cdrOutDir string
	failedCallsPrefix   string   // Configured failedCallsPrefix, used in case of flatstore CDRs
	cdrSourceIds        []string // Should be in sync with cdrFields on indexes
	runDelay            time.Duration
	csvSep              rune
	duMultiplyFactors   []float64
	cdrFilters          []utils.RSRFields       // Should be in sync with cdrFields on indexes
	cdrFields           [][]*config.CfgCdrField // Profiles directly connected with cdrFilters
	httpSkipTlsCheck    bool
	cdrcCfgs            map[string]*config.CdrcConfig // All cdrc config profiles attached to this CDRC (key will be profile instance name)
	dfltCdrcCfg         *config.CdrcConfig
	timezone            string
	cdrs                engine.Connector
	httpClient          *http.Client
	closeChan           chan struct{}        // Used to signal config reloads when we need to span different CDRC-Client
	maxOpenFiles        chan struct{}        // Maximum number of simultaneous files processed
	partialRecordsCache *PartialRecordsCache // Shared between all files in the folder we process
}

// When called fires up folder monitoring, either automated via inotify or manual by sleeping between processing
func (self *Cdrc) Run() error {
	if self.runDelay == time.Duration(0) { // Automated via inotify
		return self.trackCDRFiles()
	}
	// Not automated, process and sleep approach
	for {
		select {
		case <-self.closeChan: // Exit, reinject closeChan for other CDRCs
			utils.Logger.Info(fmt.Sprintf("<Cdrc> Shutting down CDRC on path %s.", self.cdrInDir))
			return nil
		default:
		}
		self.processCdrDir()
		time.Sleep(self.runDelay)
	}
}

// Watch the specified folder for file moves and parse the files on events
func (self *Cdrc) trackCDRFiles() (err error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return
	}
	defer watcher.Close()
	err = watcher.Add(self.cdrInDir)
	if err != nil {
		return
	}
	utils.Logger.Info(fmt.Sprintf("<Cdrc> Monitoring %s for file moves.", self.cdrInDir))
	for {
		select {
		case <-self.closeChan: // Exit, reinject closeChan for other CDRCs
			utils.Logger.Info(fmt.Sprintf("<Cdrc> Shutting down CDRC on path %s.", self.cdrInDir))
			return nil
		case ev := <-watcher.Events:
			if ev.Op&fsnotify.Create == fsnotify.Create && (self.cdrFormat != FS_CSV || path.Ext(ev.Name) != ".csv") {
				go func() { //Enable async processing here
					if err = self.processFile(ev.Name); err != nil {
						utils.Logger.Err(fmt.Sprintf("Processing file %s, error: %s", ev.Name, err.Error()))
					}
				}()
			}
		case err := <-watcher.Errors:
			utils.Logger.Err(fmt.Sprintf("Inotify error: %s", err.Error()))
		}
	}
}

// One run over the CDR folder
func (self *Cdrc) processCdrDir() error {
	utils.Logger.Info(fmt.Sprintf("<Cdrc> Parsing folder %s for CDR files.", self.cdrInDir))
	filesInDir, _ := ioutil.ReadDir(self.cdrInDir)
	for _, file := range filesInDir {
		if self.cdrFormat != FS_CSV || path.Ext(file.Name()) != ".csv" {
			go func() { //Enable async processing here
				if err := self.processFile(path.Join(self.cdrInDir, file.Name())); err != nil {
					utils.Logger.Err(fmt.Sprintf("Processing file %s, error: %s", file, err.Error()))
				}
			}()
		}
	}
	return nil
}

// Processe file at filePath and posts the valid cdr rows out of it
func (self *Cdrc) processFile(filePath string) error {
	if cap(self.maxOpenFiles) != 0 { // 0 goes for no limit
		processFile := <-self.maxOpenFiles // Queue here for maxOpenFiles
		defer func() { self.maxOpenFiles <- processFile }()
	}
	_, fn := path.Split(filePath)
	utils.Logger.Info(fmt.Sprintf("<Cdrc> Parsing: %s", filePath))
	file, err := os.Open(filePath)
	defer file.Close()
	if err != nil {
		utils.Logger.Crit(err.Error())
		return err
	}
	var recordsProcessor RecordsProcessor
	switch self.cdrFormat {
	case CSV, FS_CSV, utils.KAM_FLATSTORE, utils.OSIPS_FLATSTORE:
		csvReader := csv.NewReader(bufio.NewReader(file))
		csvReader.Comma = self.csvSep
		recordsProcessor = NewCsvRecordsProcessor(csvReader, self.cdrFormat, self.timezone, fn, self.failedCallsPrefix,
			self.cdrSourceIds, self.duMultiplyFactors, self.cdrFilters, self.cdrFields, self.httpSkipTlsCheck, self.partialRecordsCache)
	case utils.FWV:
		recordsProcessor = NewFwvRecordsProcessor(file, self.cdrcCfgs, self.dfltCdrcCfg, self.httpClient, self.httpSkipTlsCheck, self.timezone)
	default:
		return fmt.Errorf("Unsupported CDR format: %s", self.cdrFormat)
	}
	rowNr := 0 // This counts the rows in the file, not really number of CDRs
	cdrsPosted := 0
	timeStart := time.Now()
	for {
		cdrs, err := recordsProcessor.ProcessNextRecord()
		if err != nil && err == io.EOF {
			break
		}
		if err != nil {
			utils.Logger.Err(fmt.Sprintf("<Cdrc> Row %d, error: %s", rowNr, err.Error()))
			continue
		}
		for _, storedCdr := range cdrs { // Send CDRs to CDRS
			var reply string
			if self.dfltCdrcCfg.DryRun {
				utils.Logger.Info(fmt.Sprintf("<Cdrc> DryRun CDR: %+v", storedCdr))
				continue
			}
			if err := self.cdrs.ProcessCdr(storedCdr, &reply); err != nil {
				utils.Logger.Err(fmt.Sprintf("<Cdrc> Failed sending CDR, %+v, error: %s", storedCdr, err.Error()))
			} else if reply != "OK" {
				utils.Logger.Err(fmt.Sprintf("<Cdrc> Received unexpected reply for CDR, %+v, reply: %s", storedCdr, reply))
			}
			cdrsPosted += 1
		}
	}
	// Finished with file, move it to processed folder
	newPath := path.Join(self.cdrOutDir, fn)
	if err := os.Rename(filePath, newPath); err != nil {
		utils.Logger.Err(err.Error())
		return err
	}
	utils.Logger.Info(fmt.Sprintf("Finished processing %s, moved to %s. Total records processed: %d, CDRs posted: %d, run duration: %s",
		fn, newPath, recordsProcessor.ProcessedRecordsNr(), cdrsPosted, time.Now().Sub(timeStart)))
	return nil
}
