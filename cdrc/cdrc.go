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
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/howeyc/fsnotify"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

const (
	CSV = "csv"
)

func NewCdrc(config *config.CGRConfig) (*Cdrc, error) {
	cdrc := &Cdrc{cgrCfg: config}
	// Before processing, make sure in and out folders exist
	for _, dir := range []string{cdrc.cgrCfg.CdrcCdrInDir, cdrc.cgrCfg.CdrcCdrOutDir} {
		if _, err := os.Stat(dir); err != nil && os.IsNotExist(err) {
			return nil, fmt.Errorf("Folder %s does not exist", dir)
		}
	}
	if cdrc.cgrCfg.CdrcCdrType == CSV {
		if err := cdrc.parseFieldIndexesFromConfig(); err != nil {
			return nil, err
		}
	}
	cdrc.httpClient = new(http.Client)
	return cdrc, nil
}

type Cdrc struct {
	cgrCfg      *config.CGRConfig
	fieldIndxes map[string]int // Key is the name of the field, int is the position in the csv file
	httpClient  *http.Client
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
	return nil
}

// Parses fieldIndex strings into fieldIndex integers needed
func (self *Cdrc) parseFieldIndexesFromConfig() error {
	var err error
	self.fieldIndxes = make(map[string]int)
	fieldKeys := []string{utils.ACCID, utils.REQTYPE, utils.DIRECTION, utils.TENANT, utils.TOR, utils.ACCOUNT, utils.SUBJECT, utils.DESTINATION, utils.ANSWER_TIME, utils.DURATION}
	fieldIdxStrs := []string{self.cgrCfg.CdrcAccIdField, self.cgrCfg.CdrcReqTypeField, self.cgrCfg.CdrcDirectionField, self.cgrCfg.CdrcTenantField, self.cgrCfg.CdrcTorField,
		self.cgrCfg.CdrcAccountField, self.cgrCfg.CdrcSubjectField, self.cgrCfg.CdrcDestinationField, self.cgrCfg.CdrcAnswerTimeField, self.cgrCfg.CdrcDurationField}
	for i, strVal := range fieldIdxStrs {
		if self.fieldIndxes[fieldKeys[i]], err = strconv.Atoi(strVal); err != nil {
			return fmt.Errorf("Cannot parse configuration field %s into integer", fieldKeys[i])
		}
	}
	// Add extra fields here, extra fields in the form of []string{"indxInCsv1:fieldName1","indexInCsv2:fieldName2"}
	for _, fieldWithIdx := range self.cgrCfg.CdrcExtraFields {
		splt := strings.Split(fieldWithIdx, ":")
		if len(splt) != 2 {
			return errors.New("Cannot parse cdrc.extra_fields")
		}
		if utils.IsSliceMember(utils.PrimaryCdrFields, splt[0]) {
			return errors.New("Extra cdrc.extra_fields overwriting primary fields")
		}
		if self.fieldIndxes[splt[1]], err = strconv.Atoi(splt[0]); err != nil {
			return fmt.Errorf("Cannot parse configuration cdrc extra field %s into integer", splt[1])
		}
	}
	return nil
}

// Takes the record out of csv and turns it into http form which can be posted
func (self *Cdrc) cdrAsHttpForm(record []string) (url.Values, error) {
	v := url.Values{}
	v.Set(utils.CDRSOURCE, self.cgrCfg.CdrcSourceId)
	for fldName, idx := range self.fieldIndxes {
		if len(record) <= idx {
			return nil, fmt.Errorf("Ignoring record: %v - cannot extract field %s", record, fldName)
		}
		v.Set(fldName, record[idx])
	}
	return v, nil
}

// One run over the CDR folder
func (self *Cdrc) processCdrDir() error {
	engine.Logger.Info(fmt.Sprintf("Parsing folder %s for CDR files.", self.cgrCfg.CdrcCdrInDir))
	filesInDir, _ := ioutil.ReadDir(self.cgrCfg.CdrcCdrInDir)
	for _, file := range filesInDir {
		if err := self.processFile(path.Join(self.cgrCfg.CdrcCdrInDir, file.Name())); err != nil {
			return err
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
	engine.Logger.Info(fmt.Sprintf("Monitoring %s for file moves.", self.cgrCfg.CdrcCdrInDir))
	for {
		select {
		case ev := <-watcher.Event:
			if ev.IsCreate() && path.Ext(ev.Name) != ".csv" {
				if err = self.processFile(ev.Name); err != nil {
					return err
				}
			}
		case err := <-watcher.Error:
			engine.Logger.Err(fmt.Sprintf("Inotify error: %s", err.Error()))
		}
	}
	return
}

// Processe file at filePath and posts the valid cdr rows out of it
func (self *Cdrc) processFile(filePath string) error {
	_, fn := path.Split(filePath)
	engine.Logger.Info(fmt.Sprintf("Parsing: %s", filePath))
	file, err := os.Open(filePath)
	defer file.Close()
	if err != nil {
		engine.Logger.Crit(err.Error())
		return err
	}
	csvReader := csv.NewReader(bufio.NewReader(file))
	for record, ok := csvReader.Read(); ok == nil; record, ok = csvReader.Read() {
		cdrAsForm, err := self.cdrAsHttpForm(record)
		if err != nil {
			engine.Logger.Err(err.Error())
			continue
		}
		if _, err := self.httpClient.PostForm(fmt.Sprintf("http://%s/cgr", self.cgrCfg.CdrcCdrs), cdrAsForm); err != nil {
			engine.Logger.Err(fmt.Sprintf("Failed posting CDR, error: %s", err.Error()))
			continue
		}
	}
	// Finished with file, move it to processed folder
	return os.Rename(filePath, path.Join(self.cgrCfg.CdrcCdrOutDir, fn))
}
