/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) 2012-2014 ITsysCOM GmbH

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
	"unicode/utf8"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/howeyc/fsnotify"
)

const (
	CSV    = "csv"
	FS_CSV = "freeswitch_csv"
)

func NewCdrc(cdrcCfg *config.CdrcConfig, httpSkipTlsCheck bool, cdrServer *engine.CDRS) (*Cdrc, error) {
	if len(cdrcCfg.FieldSeparator) != 1 {
		return nil, fmt.Errorf("Unsupported csv separator: %s", cdrcCfg.FieldSeparator)
	}
	csvSepRune, _ := utf8.DecodeRune([]byte(cdrcCfg.FieldSeparator))
	cdrc := &Cdrc{cdrsAddress: cdrcCfg.CdrsAddress, cdrType: cdrcCfg.CdrType, cdrInDir: cdrcCfg.CdrInDir, cdrOutDir: cdrcCfg.CdrOutDir,
		cdrSourceId: cdrcCfg.CdrSourceId, runDelay: cdrcCfg.RunDelay, csvSep: csvSepRune, cdrFields: cdrcCfg.CdrFields, httpSkipTlsCheck: httpSkipTlsCheck, cdrServer: cdrServer}
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
	cdrType,
	cdrInDir,
	cdrOutDir,
	cdrSourceId string
	runDelay         time.Duration
	csvSep           rune
	cdrFields        []*config.CfgCdrField
	httpSkipTlsCheck bool
	cdrServer        *engine.CDRS // Reference towards internal cdrServer if that is the case
	httpClient       *http.Client
}

// When called fires up folder monitoring, either automated via inotify or manual by sleeping between processing
func (self *Cdrc) Run() error {
	if self.runDelay == time.Duration(0) { // Automated via inotify
		return self.trackCDRFiles()
	}
	// Not automated, process and sleep approach
	for {
		self.processCdrDir()
		time.Sleep(self.runDelay)
	}
}

// Takes the record out of csv and turns it into http form which can be posted
func (self *Cdrc) recordToStoredCdr(record []string) (*utils.StoredCdr, error) {
	storedCdr := &utils.StoredCdr{CdrHost: "0.0.0.0", CdrSource: self.cdrSourceId, ExtraFields: make(map[string]string), Cost: -1}
	var err error
	for _, cdrFldCfg := range self.cdrFields {
		var fieldVal string
		if utils.IsSliceMember([]string{CSV, FS_CSV}, self.cdrType) {
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
				var outValByte []byte
				var httpAddr string
				for _, rsrFld := range cdrFldCfg.Value {
					httpAddr += rsrFld.ParseValue("")
				}
				if outValByte, err = utils.HttpJsonPost(httpAddr, self.httpSkipTlsCheck, record); err == nil {
					fieldVal = string(outValByte)
					if len(fieldVal) == 0 && cdrFldCfg.Mandatory {
						return nil, fmt.Errorf("MandatoryIeMissing: thEmpty result for http_post field: %s", cdrFldCfg.Tag)
					}
				}
			} else {
				return nil, fmt.Errorf("Unsupported field type: %s", cdrFldCfg.Type)
			}
		} else { // Modify here when we add more supported cdr formats
			return nil, fmt.Errorf("Unsupported CDR file format: %s", self.cdrType)
		}
		switch cdrFldCfg.CdrFieldId {
		case utils.TOR:
			storedCdr.TOR = fieldVal
		case utils.ACCID:
			storedCdr.AccId = fieldVal
		case utils.REQTYPE:
			storedCdr.ReqType = fieldVal
		case utils.DIRECTION:
			storedCdr.Direction = fieldVal
		case utils.TENANT:
			storedCdr.Tenant = fieldVal
		case utils.CATEGORY:
			storedCdr.Category = fieldVal
		case utils.ACCOUNT:
			storedCdr.Account = fieldVal
		case utils.SUBJECT:
			storedCdr.Subject = fieldVal
		case utils.DESTINATION:
			storedCdr.Destination = fieldVal
		case utils.SETUP_TIME:
			if storedCdr.SetupTime, err = utils.ParseTimeDetectLayout(fieldVal); err != nil {
				return nil, fmt.Errorf("Cannot parse answer time field with value: %s, err: %s", fieldVal, err.Error())
			}
		case utils.ANSWER_TIME:
			if storedCdr.AnswerTime, err = utils.ParseTimeDetectLayout(fieldVal); err != nil {
				return nil, fmt.Errorf("Cannot parse answer time field with value: %s, err: %s", fieldVal, err.Error())
			}
		case utils.USAGE:
			if storedCdr.Usage, err = utils.ParseDurationWithSecs(fieldVal); err != nil {
				return nil, fmt.Errorf("Cannot parse duration field with value: %s, err: %s", fieldVal, err.Error())
			}
		default: // Extra fields will not match predefined so they all show up here
			storedCdr.ExtraFields[cdrFldCfg.CdrFieldId] = fieldVal
		}

	}
	storedCdr.CgrId = utils.Sha1(storedCdr.AccId, storedCdr.SetupTime.String())
	return storedCdr, nil
}

// One run over the CDR folder
func (self *Cdrc) processCdrDir() error {
	engine.Logger.Info(fmt.Sprintf("<Cdrc> Parsing folder %s for CDR files.", self.cdrInDir))
	filesInDir, _ := ioutil.ReadDir(self.cdrInDir)
	for _, file := range filesInDir {
		if self.cdrType != FS_CSV || path.Ext(file.Name()) != ".csv" {
			go func() { //Enable async processing here
				if err := self.processFile(path.Join(self.cdrInDir, file.Name())); err != nil {
					engine.Logger.Err(fmt.Sprintf("Processing file %s, error: %s", file, err.Error()))
				}
			}()
		}
	}
	return nil
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
		case ev := <-watcher.Event:
			if ev.IsCreate() && (self.cdrType != FS_CSV || path.Ext(ev.Name) != ".csv") {
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
			engine.Logger.Err(fmt.Sprintf("<Cdrc> Error in csv file: %s", err.Error()))
			continue // Other csv related errors, ignore
		}
		storedCdr, err := self.recordToStoredCdr(record)
		if err != nil {
			engine.Logger.Err(fmt.Sprintf("<Cdrc> Error in csv file: %s", err.Error()))
			continue
		}
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
