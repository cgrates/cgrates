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
	"github.com/howeyc/fsnotify"
)

const (
	CSV    = "csv"
	FS_CSV = "freeswitch_csv"
)

// Populates the
func populateStoredCdrField(cdr *utils.StoredCdr, fieldId, fieldVal string) error {
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
	case utils.ANSWER_TIME:
		if cdr.AnswerTime, err = utils.ParseTimeDetectLayout(fieldVal); err != nil {
			return fmt.Errorf("Cannot parse answer time field with value: %s, err: %s", fieldVal, err.Error())
		}
	case utils.USAGE:
		if cdr.Usage, err = utils.ParseDurationWithSecs(fieldVal); err != nil {
			return fmt.Errorf("Cannot parse duration field with value: %s, err: %s", fieldVal, err.Error())
		}
	default: // Extra fields will not match predefined so they all show up here
		cdr.ExtraFields[fieldId] = fieldVal
	}
	return nil
}

func NewCdrc(cdrcCfgs map[string]*config.CdrcConfig, httpSkipTlsCheck bool, cdrServer *engine.CDRS, exitChan chan struct{}) (*Cdrc, error) {
	var cdrcCfg *config.CdrcConfig
	for _, cdrcCfg = range cdrcCfgs { // Take the first config out, does not matter which one
		break
	}
	cdrc := &Cdrc{cdrsAddress: cdrcCfg.CdrsAddress, CdrFormat: cdrcCfg.CdrFormat, cdrInDir: cdrcCfg.CdrInDir, cdrOutDir: cdrcCfg.CdrOutDir,
		cdrSourceId: cdrcCfg.CdrSourceId, runDelay: cdrcCfg.RunDelay, csvSep: cdrcCfg.FieldSeparator, duMultiplyFactor: cdrcCfg.DataUsageMultiplyFactor,
		httpSkipTlsCheck: httpSkipTlsCheck, cdrServer: cdrServer, exitChan: exitChan}
	cdrc.cdrFilters = make([]utils.RSRFields, len(cdrcCfgs))
	cdrc.cdrFields = make([][]*config.CfgCdrField, len(cdrcCfgs))
	idx := 0
	for _, cfg := range cdrcCfgs {
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
	cdrOutDir,
	cdrSourceId string
	runDelay         time.Duration
	csvSep           rune
	duMultiplyFactor float64
	cdrFilters       []utils.RSRFields       // Should be in sync with cdrFields on indexes
	cdrFields        [][]*config.CfgCdrField // Profiles directly connected with cdrFilters
	httpSkipTlsCheck bool
	cdrServer        *engine.CDRS // Reference towards internal cdrServer if that is the case
	httpClient       *http.Client
	exitChan         chan struct{}
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
	err = watcher.Watch(self.cdrInDir)
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
		case ev := <-watcher.Event:
			if ev.IsCreate() && (self.CdrFormat != FS_CSV || path.Ext(ev.Name) != ".csv") {
				go func() { //Enable async processing here
					if err = self.processFile(ev.Name); err != nil {
						engine.Logger.Err(fmt.Sprintf("Processing file %s, error: %s", ev.Name, err.Error()))
					}
				}()
			}
		case err := <-watcher.Error:
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
				if err := self.processFile(path.Join(self.cdrInDir, file.Name())); err != nil {
					engine.Logger.Err(fmt.Sprintf("Processing file %s, error: %s", file, err.Error()))
				}
			}()
		}
	}
	return nil
}

// Processe file at filePath and posts the valid cdr rows out of it
func (self *Cdrc) processFile(filePath string) error {
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
		recordCdrs := make([]*utils.StoredCdr, 0) // More CDRs based on the number of filters and field templates
		for idx, cdrFieldsInst := range self.cdrFields {
			// Make sure filters are matching
			filterBreak := false
			for _, rsrFilter := range self.cdrFilters[idx] {
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
			if storedCdr, err := self.recordToStoredCdr(record, cdrFieldsInst); err != nil {
				engine.Logger.Err(fmt.Sprintf("<Cdrc> Row %d - failed converting to StoredCdr, error: %s", procRowNr, err.Error()))
				continue
			} else {
				recordCdrs = append(recordCdrs, storedCdr)
			}
		}
		for _, storedCdr := range recordCdrs {
			if self.cdrsAddress == utils.INTERNAL {
				if err := self.cdrServer.ProcessCdr(storedCdr); err != nil {
					engine.Logger.Err(fmt.Sprintf("<Cdrc> Failed posting CDR, row: %d, error: %s", procRowNr, err.Error()))
					continue
				}
			} else { // CDRs listening on IP
				if _, err := self.httpClient.PostForm(fmt.Sprintf("http://%s/cgr", self.cdrsAddress), storedCdr.AsHttpForm()); err != nil {
					engine.Logger.Err(fmt.Sprintf("<Cdrc> Failed posting CDR, row: %d, error: %s", procRowNr, err.Error()))
					continue
				}
			}
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

// Takes the record out of csv and turns it into http form which can be posted
func (self *Cdrc) recordToStoredCdr(record []string, cdrFields []*config.CfgCdrField) (*utils.StoredCdr, error) {
	storedCdr := &utils.StoredCdr{CdrHost: "0.0.0.0", CdrSource: self.cdrSourceId, ExtraFields: make(map[string]string), Cost: -1}
	var err error
	var lazyHttpFields []*config.CfgCdrField
	for _, cdrFldCfg := range cdrFields {
		var fieldVal string
		if utils.IsSliceMember([]string{CSV, FS_CSV}, self.CdrFormat) {
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
	if storedCdr.TOR == utils.DATA && self.duMultiplyFactor != 0 {
		storedCdr.Usage = time.Duration(float64(storedCdr.Usage.Nanoseconds()) * self.duMultiplyFactor)
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
