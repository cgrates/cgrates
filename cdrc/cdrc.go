/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2013 ITsysCOM

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

	"github.com/cgrates/cgrates/cdrs"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/howeyc/fsnotify"
)

const (
	CSV    = "csv"
	FS_CSV = "freeswitch_csv"
)

func NewCdrc(config *config.CGRConfig, cdrServer *cdrs.CDRS) (*Cdrc, error) {
	cdrc := &Cdrc{cgrCfg: config, cdrServer: cdrServer}
	// Before processing, make sure in and out folders exist
	for _, dir := range []string{cdrc.cgrCfg.CdrcCdrInDir, cdrc.cgrCfg.CdrcCdrOutDir} {
		if _, err := os.Stat(dir); err != nil && os.IsNotExist(err) {
			return nil, fmt.Errorf("Folder %s does not exist", dir)
		}
	}
	if err := cdrc.parseFieldsConfig(); err != nil {
		return nil, err
	}
	cdrc.httpClient = new(http.Client)
	return cdrc, nil
}

type Cdrc struct {
	cgrCfg       *config.CGRConfig
	cdrServer    *cdrs.CDRS
	cfgCdrFields map[string]string // Key is the name of the field
	httpClient   *http.Client
}

// When called fires up folder monitoring, either automated via inotify or manual by sleeping between processing
func (self *Cdrc) Run() error {
	if self.cgrCfg.CdrcRunDelay == time.Duration(0) { // Automated via inotify
		return self.trackCDRFiles()
	}
	// No automated, process and sleep approach
	for {
		self.processCdrDir()
		time.Sleep(self.cgrCfg.CdrcRunDelay)
	}
}

// Loads all fields (primary and extra) into cfgCdrFields, do some pre-checks (eg: in case of csv make sure that values are integers)
func (self *Cdrc) parseFieldsConfig() error {
	var err error
	self.cfgCdrFields = map[string]string{
		utils.ACCID:       self.cgrCfg.CdrcAccIdField,
		utils.REQTYPE:     self.cgrCfg.CdrcReqTypeField,
		utils.DIRECTION:   self.cgrCfg.CdrcDirectionField,
		utils.TENANT:      self.cgrCfg.CdrcTenantField,
		utils.Category:    self.cgrCfg.CdrcTorField,
		utils.ACCOUNT:     self.cgrCfg.CdrcAccountField,
		utils.SUBJECT:     self.cgrCfg.CdrcSubjectField,
		utils.DESTINATION: self.cgrCfg.CdrcDestinationField,
		utils.SETUP_TIME:  self.cgrCfg.CdrcSetupTimeField,
		utils.ANSWER_TIME: self.cgrCfg.CdrcAnswerTimeField,
		utils.DURATION:    self.cgrCfg.CdrcDurationField,
	}

	// Add extra fields here, config extra fields in the form of []string{"fieldName1:indxInCsv1","fieldName2: indexInCsv2"}
	for _, fieldWithIdx := range self.cgrCfg.CdrcExtraFields {
		splt := strings.Split(fieldWithIdx, ":")
		if len(splt) != 2 {
			return errors.New("Cannot parse cdrc.extra_fields")
		}
		if utils.IsSliceMember(utils.PrimaryCdrFields, splt[0]) {
			return errors.New("Extra cdrc.extra_fields overwriting primary fields")
		}
		self.cfgCdrFields[splt[0]] = splt[1]
	}
	// Fields populated, do some sanity checks here
	for cdrField, cfgVal := range self.cfgCdrFields {
		if utils.IsSliceMember([]string{CSV, FS_CSV}, self.cgrCfg.CdrcCdrType) && !strings.HasPrefix(cfgVal, utils.STATIC_VALUE_PREFIX) {
			if _, err = strconv.Atoi(cfgVal); err != nil {
				return fmt.Errorf("Cannot parse configuration field %s into integer", cdrField)
			}
		}
	}
	return nil
}

// Takes the record out of csv and turns it into http form which can be posted
func (self *Cdrc) recordAsStoredCdr(record []string) (*utils.StoredCdr, error) {
	ratedCdr := &utils.StoredCdr{CdrSource: self.cgrCfg.CdrcSourceId, ExtraFields: map[string]string{}, Cost: -1}
	var err error
	for cfgFieldName, cfgFieldVal := range self.cfgCdrFields {
		var fieldVal string
		if strings.HasPrefix(cfgFieldVal, utils.STATIC_VALUE_PREFIX) {
			fieldVal = cfgFieldVal[1:]
		} else if utils.IsSliceMember([]string{CSV, FS_CSV}, self.cgrCfg.CdrcCdrType) {
			if cfgFieldIdx, err := strconv.Atoi(cfgFieldVal); err != nil { // Should in theory never happen since we have already parsed config
				return nil, err
			} else if len(record) <= cfgFieldIdx {
				return nil, fmt.Errorf("Ignoring record: %v - cannot extract field %s", record, cfgFieldName)
			} else {
				fieldVal = record[cfgFieldIdx]
			}
		} else { // Modify here when we add more supported cdr formats
			fieldVal = "UNKNOWN"
		}
		switch cfgFieldName {
		case utils.ACCID:
			ratedCdr.AccId = fieldVal
		case utils.REQTYPE:
			ratedCdr.ReqType = fieldVal
		case utils.DIRECTION:
			ratedCdr.Direction = fieldVal
		case utils.TENANT:
			ratedCdr.Tenant = fieldVal
		case utils.Category:
			ratedCdr.Category = fieldVal
		case utils.ACCOUNT:
			ratedCdr.Account = fieldVal
		case utils.SUBJECT:
			ratedCdr.Subject = fieldVal
		case utils.DESTINATION:
			ratedCdr.Destination = fieldVal
		case utils.SETUP_TIME:
			if ratedCdr.SetupTime, err = utils.ParseTimeDetectLayout(fieldVal); err != nil {
				return nil, fmt.Errorf("Cannot parse answer time field, err: %s", err.Error())
			}
		case utils.ANSWER_TIME:
			if ratedCdr.AnswerTime, err = utils.ParseTimeDetectLayout(fieldVal); err != nil {
				return nil, fmt.Errorf("Cannot parse answer time field, err: %s", err.Error())
			}
		case utils.DURATION:
			if ratedCdr.Duration, err = utils.ParseDurationWithSecs(fieldVal); err != nil {
				return nil, fmt.Errorf("Cannot parse duration field, err: %s", err.Error())
			}
		default: // Extra fields will not match predefined so they all show up here
			ratedCdr.ExtraFields[cfgFieldName] = fieldVal
		}

	}
	ratedCdr.CgrId = utils.Sha1(ratedCdr.AccId, ratedCdr.SetupTime.String())
	return ratedCdr, nil
}

// One run over the CDR folder
func (self *Cdrc) processCdrDir() error {
	engine.Logger.Info(fmt.Sprintf("<Cdrc> Parsing folder %s for CDR files.", self.cgrCfg.CdrcCdrInDir))
	filesInDir, _ := ioutil.ReadDir(self.cgrCfg.CdrcCdrInDir)
	for _, file := range filesInDir {
		if self.cgrCfg.CdrcCdrType != FS_CSV || path.Ext(file.Name()) != ".csv" {
			if err := self.processFile(path.Join(self.cgrCfg.CdrcCdrInDir, file.Name())); err != nil {
				return err
			}
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
	err = watcher.Watch(self.cgrCfg.CdrcCdrInDir)
	if err != nil {
		return
	}
	engine.Logger.Info(fmt.Sprintf("<Cdrc> Monitoring %s for file moves.", self.cgrCfg.CdrcCdrInDir))
	for {
		select {
		case ev := <-watcher.Event:
			if ev.IsCreate() && (self.cgrCfg.CdrcCdrType != FS_CSV || path.Ext(ev.Name) != ".csv") {
				if err = self.processFile(ev.Name); err != nil {
					engine.Logger.Err(fmt.Sprintf("Processing file %s, error: %s", ev.Name, err.Error()))
				}
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
	for {
		record, err := csvReader.Read()
		if err != nil && err == io.EOF {
			break // End of file
		} else if err != nil {
			engine.Logger.Err(fmt.Sprintf("<Cdrc> Error in csv file: %s", err.Error()))
			continue // Other csv related errors, ignore
		}
		rawCdr, err := self.recordAsStoredCdr(record)
		if err != nil {
			engine.Logger.Err(fmt.Sprintf("<Cdrc> Error in csv file: %s", err.Error()))
			continue
		}
		if self.cgrCfg.CdrcCdrs == utils.INTERNAL {
			if err := self.cdrServer.ProcessRawCdr(rawCdr); err != nil {
				engine.Logger.Err(fmt.Sprintf("<Cdrc> Failed posting CDR, error: %s", err.Error()))
				continue
			}
		} else { // CDRs listening on IP
			if _, err := self.httpClient.PostForm(fmt.Sprintf("http://%s/cgr", self.cgrCfg.HTTPListen), rawCdr.AsRawCdrHttpForm()); err != nil {
				engine.Logger.Err(fmt.Sprintf("<Cdrc> Failed posting CDR, error: %s", err.Error()))
				continue
			}
		}
	}
	// Finished with file, move it to processed folder
	newPath := path.Join(self.cgrCfg.CdrcCdrOutDir, fn)
	if err := os.Rename(filePath, newPath); err != nil {
		engine.Logger.Err(err.Error())
		return err
	}
	engine.Logger.Info(fmt.Sprintf("Finished processing %s, moved to %s", fn, newPath))
	return nil
}
