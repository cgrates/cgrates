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
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
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
func populateStoredCdrField(cdr *engine.StoredCdr, fieldId, fieldVal string) error {
	var err error
	switch fieldId {
	case utils.TOR:
		cdr.TOR = fieldVal
	case utils.ACCID:
		cdr.AccId = fieldVal
	case utils.REQTYPE:
		cdr.ReqType = fieldVal
	case utils.DIRECTION:
		cdr.Direction = fieldVal
	case utils.TENANT:
		cdr.Tenant = fieldVal
	case utils.CATEGORY:
		cdr.Category = fieldVal
	case utils.ACCOUNT:
		cdr.Account = fieldVal
	case utils.SUBJECT:
		cdr.Subject = fieldVal
	case utils.DESTINATION:
		cdr.Destination = fieldVal
	case utils.SETUP_TIME:
		if cdr.SetupTime, err = utils.ParseTimeDetectLayout(fieldVal); err != nil {
			return fmt.Errorf("Cannot parse answer time field with value: %s, err: %s", fieldVal, err.Error())
		}
	case utils.PDD:
		if cdr.Pdd, err = utils.ParseDurationWithSecs(fieldVal); err != nil {
			return fmt.Errorf("Cannot parse answer time field with value: %s, err: %s", fieldVal, err.Error())
		}
	case utils.ANSWER_TIME:
		if cdr.AnswerTime, err = utils.ParseTimeDetectLayout(fieldVal); err != nil {
			return fmt.Errorf("Cannot parse answer time field with value: %s, err: %s", fieldVal, err.Error())
		}
	case utils.USAGE:
		if cdr.Usage, err = utils.ParseDurationWithSecs(fieldVal); err != nil {
			return fmt.Errorf("Cannot parse duration field with value: %s, err: %s", fieldVal, err.Error())
		}
	case utils.SUPPLIER:
		cdr.Supplier = fieldVal
	case utils.DISCONNECT_CAUSE:
		cdr.DisconnectCause = fieldVal
	default: // Extra fields will not match predefined so they all show up here
		cdr.ExtraFields[fieldId] = fieldVal
	}
	return nil
}

func NewPartialFlatstoreRecord(record []string) (*PartialFlatstoreRecord, error) {
	if len(record) < 7 {
		return nil, errors.New("MISSING_IE")
	}
	pr := &PartialFlatstoreRecord{Method: record[0], AccId: record[3] + record[1] + record[2], Values: record}
	var err error
	if pr.Timestamp, err = utils.ParseTimeDetectLayout(record[6]); err != nil {
		return nil, err
	}
	return pr, nil
}

// This is a partial record received from Flatstore, can be INVITE or BYE and it needs to be paired in order to produce duration
type PartialFlatstoreRecord struct {
	Method    string    // INVITE or BYE
	AccId     string    // Copute here the AccId
	Timestamp time.Time // Timestamp of the event, as written by db_flastore module
	Values    []string  // Can contain original values or updated via UpdateValues
}

// Pairs INVITE and BYE into final record containing as last element the duration
func pairToRecord(part1, part2 *PartialFlatstoreRecord) ([]string, error) {
	var invite, bye *PartialFlatstoreRecord
	if part1.Method == "INVITE" {
		invite = part1
	} else if part2.Method == "INVITE" {
		invite = part2
	} else {
		return nil, errors.New("MISSING_INVITE")
	}
	if part1.Method == "BYE" {
		bye = part1
	} else if part2.Method == "BYE" {
		bye = part2
	} else {
		return nil, errors.New("MISSING_BYE")
	}
	if len(invite.Values) != len(bye.Values) {
		return nil, errors.New("INCONSISTENT_VALUES_LENGTH")
	}
	record := invite.Values
	for idx := range record {
		switch idx {
		case 0, 1, 2, 3, 6: // Leave these values as they are
		case 4, 5:
			record[idx] = bye.Values[idx] // Update record with status from bye
		default:
			if bye.Values[idx] != "" { // Any value higher than 6 is dynamically inserted, overwrite if non empty
				record[idx] = bye.Values[idx]
			}

		}
	}
	callDur := bye.Timestamp.Sub(invite.Timestamp)
	record = append(record, strconv.FormatFloat(callDur.Seconds(), 'f', -1, 64))
	return record, nil
}

/*
One instance  of CDRC will act on one folder.
Common parameters within configs processed:
 * cdrS, cdrFormat, cdrInDir, cdrOutDir, runDelay
Parameters specific per config instance:
 * duMultiplyFactor, cdrSourceId, cdrFilter, cdrFields
*/
func NewCdrc(cdrcCfgs map[string]*config.CdrcConfig, httpSkipTlsCheck bool, cdrServer *engine.CdrServer, exitChan chan struct{}) (*Cdrc, error) {
	var cdrcCfg *config.CdrcConfig
	for _, cdrcCfg = range cdrcCfgs { // Take the first config out, does not matter which one
		break
	}
	cdrc := &Cdrc{cdrsAddress: cdrcCfg.Cdrs, CdrFormat: cdrcCfg.CdrFormat, cdrInDir: cdrcCfg.CdrInDir, cdrOutDir: cdrcCfg.CdrOutDir,
		runDelay: cdrcCfg.RunDelay, csvSep: cdrcCfg.FieldSeparator,
		httpSkipTlsCheck: httpSkipTlsCheck, cdrServer: cdrServer, exitChan: exitChan, maxOpenFiles: make(chan struct{}, cdrcCfg.MaxOpenFiles),
		partialRecords: make(map[string]map[string]*PartialFlatstoreRecord), guard: engine.NewGuardianLock()}
	var processCsvFile struct{}
	for i := 0; i < cdrcCfg.MaxOpenFiles; i++ {
		cdrc.maxOpenFiles <- processCsvFile // Empty initiate so we do not need to wait later when we pop
	}
	cdrc.cdrSourceIds = make([]string, len(cdrcCfgs))
	cdrc.duMultiplyFactors = make([]float64, len(cdrcCfgs))
	cdrc.cdrFilters = make([]utils.RSRFields, len(cdrcCfgs))
	cdrc.cdrFields = make([][]*config.CfgCdrField, len(cdrcCfgs))
	idx := 0
	for _, cfg := range cdrcCfgs {
		if idx == 0 { // Steal the config from just one instance since it should be the same for all
			cdrc.partialRecordCache = cfg.PartialRecordCache
			cdrc.failedCallsPrefix = cfg.FailedCallsPrefix
		}
		cdrc.cdrSourceIds[idx] = cfg.CdrSourceId
		cdrc.duMultiplyFactors[idx] = cfg.DataUsageMultiplyFactor
		cdrc.cdrFilters[idx] = cfg.CdrFilter
		cdrc.cdrFields[idx] = cfg.CdrFields
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
	cdrsAddress,
	CdrFormat,
	cdrInDir,
	cdrOutDir string
	failedCallsPrefix  string   // Configured failedCallsPrefix, used in case of flatstore CDRs
	cdrSourceIds       []string // Should be in sync with cdrFields on indexes
	runDelay           time.Duration
	csvSep             rune
	duMultiplyFactors  []float64
	cdrFilters         []utils.RSRFields       // Should be in sync with cdrFields on indexes
	cdrFields          [][]*config.CfgCdrField // Profiles directly connected with cdrFilters
	httpSkipTlsCheck   bool
	cdrServer          *engine.CdrServer // Reference towards internal cdrServer if that is the case
	httpClient         *http.Client
	exitChan           chan struct{}
	maxOpenFiles       chan struct{}                                 // Maximum number of simultaneous files processed
	partialRecords     map[string]map[string]*PartialFlatstoreRecord // [FileName"][AccId]*PartialRecord
	partialRecordCache time.Duration                                 // Duration to cache partial records for
	guard              *engine.GuardianLock
}

// When called fires up folder monitoring, either automated via inotify or manual by sleeping between processing
func (self *Cdrc) Run() error {
	if self.runDelay == time.Duration(0) { // Automated via inotify
		return self.trackCDRFiles()
	}
	// Not automated, process and sleep approach
	for {
		select {
		case exitChan := <-self.exitChan: // Exit, reinject exitChan for other CDRCs
			self.exitChan <- exitChan
			engine.Logger.Info(fmt.Sprintf("<Cdrc> Shutting down CDRC on path %s.", self.cdrInDir))
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
	engine.Logger.Info(fmt.Sprintf("<Cdrc> Monitoring %s for file moves.", self.cdrInDir))
	for {
		select {
		case exitChan := <-self.exitChan: // Exit, reinject exitChan for other CDRCs
			self.exitChan <- exitChan
			engine.Logger.Info(fmt.Sprintf("<Cdrc> Shutting down CDRC on path %s.", self.cdrInDir))
			return nil
		case ev := <-watcher.Events:
			if ev.Op&fsnotify.Create == fsnotify.Create && (self.CdrFormat != FS_CSV || path.Ext(ev.Name) != ".csv") {
				go func() { //Enable async processing here
					if err = self.processCsvFile(ev.Name); err != nil {
						engine.Logger.Err(fmt.Sprintf("Processing file %s, error: %s", ev.Name, err.Error()))
					}
				}()
			}
		case err := <-watcher.Errors:
			engine.Logger.Err(fmt.Sprintf("Inotify error: %s", err.Error()))
		}
	}
}

// One run over the CDR folder
func (self *Cdrc) processCdrDir() error {
	engine.Logger.Info(fmt.Sprintf("<Cdrc> Parsing folder %s for CDR files.", self.cdrInDir))
	filesInDir, _ := ioutil.ReadDir(self.cdrInDir)
	for _, file := range filesInDir {
		if self.CdrFormat != FS_CSV || path.Ext(file.Name()) != ".csv" {
			go func() { //Enable async processing here
				if err := self.processCsvFile(path.Join(self.cdrInDir, file.Name())); err != nil {
					engine.Logger.Err(fmt.Sprintf("Processing file %s, error: %s", file, err.Error()))
				}
			}()
		}
	}
	return nil
}

// Processe file at filePath and posts the valid cdr rows out of it
func (self *Cdrc) processCsvFile(filePath string) error {
	if cap(self.maxOpenFiles) != 0 { // 0 goes for no limit
		processCsvFile := <-self.maxOpenFiles // Queue here for maxOpenFiles
		defer func() { self.maxOpenFiles <- processCsvFile }()
	}
	_, fn := path.Split(filePath)
	engine.Logger.Info(fmt.Sprintf("<Cdrc> Parsing: %s", filePath))
	file, err := os.Open(filePath)
	defer file.Close()
	if err != nil {
		engine.Logger.Crit(err.Error())
		return err
	}
	csvReader := csv.NewReader(bufio.NewReader(file))
	csvReader.Comma = self.csvSep
	procRowNr := 0
	timeStart := time.Now()
	for {
		record, err := csvReader.Read()
		if err != nil && err == io.EOF {
			break // End of file
		}
		procRowNr += 1 // Only increase if not end of file
		if err != nil {
			engine.Logger.Err(fmt.Sprintf("<Cdrc> Row %d - csv error: %s", procRowNr, err.Error()))
			continue // Other csv related errors, ignore
		}
		if utils.IsSliceMember([]string{utils.KAM_FLATSTORE, utils.OSIPS_FLATSTORE}, self.CdrFormat) { // partial records for flatstore CDRs
			if record, err = self.processPartialRecord(record, fn); err != nil {
				engine.Logger.Err(fmt.Sprintf("<Cdrc> Failed processing partial record, row: %d, error: %s", procRowNr, err.Error()))
				continue
			} else if record == nil {
				continue
			}
			// Record was overwriten with complete information out of cache
		}
		if err := self.processRecord(record, procRowNr); err != nil {
			engine.Logger.Err(fmt.Sprintf("<Cdrc> Failed processing CDR, row: %d, error: %s", procRowNr, err.Error()))
			continue
		}
	}
	// Finished with file, move it to processed folder
	newPath := path.Join(self.cdrOutDir, fn)
	if err := os.Rename(filePath, newPath); err != nil {
		engine.Logger.Err(err.Error())
		return err
	}
	engine.Logger.Info(fmt.Sprintf("Finished processing %s, moved to %s. Total records processed: %d, run duration: %s",
		fn, newPath, procRowNr, time.Now().Sub(timeStart)))
	return nil
}

// Processes a single partial record for flatstore CDRs
func (self *Cdrc) processPartialRecord(record []string, fileName string) ([]string, error) {
	if strings.HasPrefix(fileName, self.failedCallsPrefix) { // Use the first index since they should be the same in all configs
		record = append(record, "0") // Append duration 0 for failed calls flatstore CDR and do not process it further
		return record, nil
	}
	pr, err := NewPartialFlatstoreRecord(record)
	if err != nil {
		return nil, err
	}
	// Retrieve and complete the record from cache
	var cachedFilename string
	var cachedPartial *PartialFlatstoreRecord
	cachedFNames := []string{fileName} // Higher probability to match as firstFileName
	for fName := range self.partialRecords {
		if fName != fileName {
			cachedFNames = append(cachedFNames, fName)
		}
	}
	for _, fName := range cachedFNames { // Need to lock them individually
		self.guard.Guard(func() (interface{}, error) {
			var hasPartial bool
			if cachedPartial, hasPartial = self.partialRecords[fName][pr.AccId]; hasPartial {
				cachedFilename = fName
			}
			return nil, nil
		}, fName)
		if cachedPartial != nil {
			break
		}
	}

	if cachedPartial == nil { // Not cached, do it here and stop processing
		self.guard.Guard(func() (interface{}, error) {
			if fileMp, hasFile := self.partialRecords[fileName]; !hasFile {
				self.partialRecords[fileName] = map[string]*PartialFlatstoreRecord{pr.AccId: pr}
				if self.partialRecordCache != 0 { // Schedule expiry/dump of the just created entry in cache
					go func() {
						time.Sleep(self.partialRecordCache)
						self.dumpUnpairedRecords(fileName)
					}()
				}
			} else if _, hasAccId := fileMp[pr.AccId]; !hasAccId {
				self.partialRecords[fileName][pr.AccId] = pr
			}
			return nil, nil
		}, fileName)
		return nil, nil
	}

	pairedRecord, err := pairToRecord(cachedPartial, pr)
	if err != nil {
		return nil, err
	}
	self.guard.Guard(func() (interface{}, error) {
		delete(self.partialRecords[cachedFilename], pr.AccId) // Remove the record out of cache
		return nil, nil
	}, fileName)
	return pairedRecord, nil
}

// Dumps the cache into a .unpaired file in the outdir and cleans cache after
func (self *Cdrc) dumpUnpairedRecords(fileName string) error {
	_, err := self.guard.Guard(func() (interface{}, error) {
		if len(self.partialRecords[fileName]) != 0 { // Only write the file if there are records in the cache
			unpairedFilePath := path.Join(self.cdrOutDir, fileName+UNPAIRED_SUFFIX)
			fileOut, err := os.Create(unpairedFilePath)
			if err != nil {
				engine.Logger.Err(fmt.Sprintf("<Cdrc> Failed creating %s, error: %s", unpairedFilePath, err.Error()))
				return nil, err
			}
			csvWriter := csv.NewWriter(fileOut)
			csvWriter.Comma = self.csvSep
			for _, pr := range self.partialRecords[fileName] {
				if err := csvWriter.Write(pr.Values); err != nil {
					engine.Logger.Err(fmt.Sprintf("<Cdrc> Failed writing unpaired record %v to file: %s, error: %s", pr, unpairedFilePath, err.Error()))
					return nil, err
				}
			}
			csvWriter.Flush()
		}
		delete(self.partialRecords, fileName)
		return nil, nil
	}, fileName)
	return err
}

// Takes the record from a slice and turns it into StoredCdrs, posting them to the cdrServer
func (self *Cdrc) processRecord(record []string, srcRowNr int) error {
	recordCdrs := make([]*engine.StoredCdr, 0) // More CDRs based on the number of filters and field templates
	for idx := range self.cdrFields {          // cdrFields coming from more templates will produce individual storCdr records
		// Make sure filters are matching
		filterBreak := false
		for _, rsrFilter := range self.cdrFilters[idx] {
			if rsrFilter == nil { // Nil filter does not need to match anything
				continue
			}
			if cfgFieldIdx, _ := strconv.Atoi(rsrFilter.Id); len(record) <= cfgFieldIdx {
				return fmt.Errorf("Ignoring record: %v - cannot compile filter %+v", record, rsrFilter)
			} else if !rsrFilter.FilterPasses(record[cfgFieldIdx]) {
				filterBreak = true
				break
			}
		}
		if filterBreak { // Stop importing cdrc fields profile due to non matching filter
			continue
		}
		if storedCdr, err := self.recordToStoredCdr(record, idx); err != nil {
			engine.Logger.Err(fmt.Sprintf("<Cdrc> Row %d - failed converting to StoredCdr, error: %s", srcRowNr, err.Error()))
			continue
		} else {
			recordCdrs = append(recordCdrs, storedCdr)
		}
	}
	for _, storedCdr := range recordCdrs {
		if self.cdrsAddress == utils.INTERNAL {
			if err := self.cdrServer.ProcessCdr(storedCdr); err != nil {
				engine.Logger.Err(fmt.Sprintf("<Cdrc> Failed posting CDR, row: %d, error: %s", srcRowNr, err.Error()))
				continue
			}
		} else { // CDRs listening on IP
			if _, err := self.httpClient.PostForm(fmt.Sprintf("http://%s/cdr_post", self.cdrsAddress), storedCdr.AsHttpForm()); err != nil {
				engine.Logger.Err(fmt.Sprintf("<Cdrc> Failed posting CDR, row: %d, error: %s", srcRowNr, err.Error()))
				continue
			}
		}
	}
	return nil
}

// Takes the record out of csv and turns it into storedCdr which can be processed by CDRS
func (self *Cdrc) recordToStoredCdr(record []string, cfgIdx int) (*engine.StoredCdr, error) {
	storedCdr := &engine.StoredCdr{CdrHost: "0.0.0.0", CdrSource: self.cdrSourceIds[cfgIdx], ExtraFields: make(map[string]string), Cost: -1}
	var err error
	var lazyHttpFields []*config.CfgCdrField
	for _, cdrFldCfg := range self.cdrFields[cfgIdx] {
		if utils.IsSliceMember([]string{utils.KAM_FLATSTORE, utils.OSIPS_FLATSTORE}, self.CdrFormat) { // Hardcode some values in case of flatstore
			switch cdrFldCfg.CdrFieldId {
			case utils.ACCID:
				cdrFldCfg.Value = utils.ParseRSRFieldsMustCompile("3;1;2", utils.INFIELD_SEP) // in case of flatstore, accounting id is made up out of callid, from_tag and to_tag
			case utils.USAGE:
				cdrFldCfg.Value = utils.ParseRSRFieldsMustCompile(strconv.Itoa(len(record)-1), utils.INFIELD_SEP) // in case of flatstore, last element will be the duration computed by us
			}

		}
		var fieldVal string
		if utils.IsSliceMember([]string{CSV, FS_CSV, utils.KAM_FLATSTORE, utils.OSIPS_FLATSTORE}, self.CdrFormat) {
			if cdrFldCfg.Type == utils.CDRFIELD {
				for _, cfgFieldRSR := range cdrFldCfg.Value {
					if cfgFieldRSR.IsStatic() {
						fieldVal += cfgFieldRSR.ParseValue("")
					} else { // Dynamic value extracted using index
						if cfgFieldIdx, _ := strconv.Atoi(cfgFieldRSR.Id); len(record) <= cfgFieldIdx {
							return nil, fmt.Errorf("Ignoring record: %v - cannot extract field %s", record, cdrFldCfg.Tag)
						} else {
							fieldVal += cfgFieldRSR.ParseValue(record[cfgFieldIdx])
						}
					}
				}
			} else if cdrFldCfg.Type == utils.HTTP_POST {
				lazyHttpFields = append(lazyHttpFields, cdrFldCfg) // Will process later so we can send an estimation of storedCdr to http server
			} else {
				return nil, fmt.Errorf("Unsupported field type: %s", cdrFldCfg.Type)
			}
		} else { // Modify here when we add more supported cdr formats
			return nil, fmt.Errorf("Unsupported CDR file format: %s", self.CdrFormat)
		}
		if err := populateStoredCdrField(storedCdr, cdrFldCfg.CdrFieldId, fieldVal); err != nil {
			return nil, err
		}
	}
	storedCdr.CgrId = utils.Sha1(storedCdr.AccId, storedCdr.SetupTime.String())
	if storedCdr.TOR == utils.DATA && self.duMultiplyFactors[cfgIdx] != 0 {
		storedCdr.Usage = time.Duration(float64(storedCdr.Usage.Nanoseconds()) * self.duMultiplyFactors[cfgIdx])
	}
	for _, httpFieldCfg := range lazyHttpFields { // Lazy process the http fields
		var outValByte []byte
		var fieldVal, httpAddr string
		for _, rsrFld := range httpFieldCfg.Value {
			httpAddr += rsrFld.ParseValue("")
		}
		if outValByte, err = utils.HttpJsonPost(httpAddr, self.httpSkipTlsCheck, storedCdr); err != nil && httpFieldCfg.Mandatory {
			return nil, err
		} else {
			fieldVal = string(outValByte)
			if len(fieldVal) == 0 && httpFieldCfg.Mandatory {
				return nil, fmt.Errorf("MandatoryIeMissing: Empty result for http_post field: %s", httpFieldCfg.Tag)
			}
			if err := populateStoredCdrField(storedCdr, httpFieldCfg.CdrFieldId, fieldVal); err != nil {
				return nil, err
			}
		}
	}
	return storedCdr, nil
}
