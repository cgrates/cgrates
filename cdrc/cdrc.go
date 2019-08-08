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

package cdrc

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
	"github.com/fsnotify/fsnotify"
)

const (
	UNPAIRED_SUFFIX = ".unpaired"
)

// Understands and processes a specific format of cdr (eg: .csv or .fwv)
type RecordsProcessor interface {
	ProcessNextRecord() ([]*engine.CDR, error) // Process a single record in the CDR file, return a slice of CDRs since based on configuration we can have more templates
	ProcessedRecordsNr() int64
}

/*
One instance  of CDRC will act on one folder.
Common parameters within configs processed:
 * cdrS, cdrFormat, CDRInPath, CDROutPath, runDelay
Parameters specific per config instance:
 * cdrSourceId, cdrFilter, cdrFields
*/
func NewCdrc(cdrcCfgs []*config.CdrcCfg, httpSkipTlsCheck bool, cdrs rpcclient.RpcClientConnection,
	closeChan chan struct{}, dfltTimezone string, filterS *engine.FilterS) (cdrc *Cdrc, err error) {
	cdrcCfg := cdrcCfgs[0]
	cdrc = &Cdrc{
		httpSkipTlsCheck: httpSkipTlsCheck,
		cdrcCfgs:         cdrcCfgs,
		dfltCdrcCfg:      cdrcCfg,
		timezone:         utils.FirstNonEmpty(cdrcCfg.Timezone, dfltTimezone),
		cdrs:             cdrs,
		closeChan:        closeChan,
		maxOpenFiles:     make(chan struct{}, cdrcCfg.MaxOpenFiles),
	}
	// Before processing, make sure in and out folders exist
	if utils.CDRCFileFormats.Has(cdrcCfg.CdrFormat) {
		for _, dir := range []string{cdrcCfg.CDRInPath, cdrcCfg.CDROutPath} {
			if _, err := os.Stat(dir); err != nil && os.IsNotExist(err) {
				return nil, fmt.Errorf("<CDRC> nonexistent folder: %s", dir)
			}
		}

		var processFile struct{}
		for i := 0; i < cdrcCfg.MaxOpenFiles; i++ {
			cdrc.maxOpenFiles <- processFile // Empty initiate so we do not need to wait later when we pop
		}
	}
	cdrc.unpairedRecordsCache = NewUnpairedRecordsCache(cdrcCfg.PartialRecordCache,
		cdrcCfg.CDROutPath, cdrcCfg.FieldSeparator)
	cdrc.partialRecordsCache = NewPartialRecordsCache(cdrcCfg.PartialRecordCache,
		cdrcCfg.PartialCacheExpiryAction, cdrcCfg.CDROutPath,
		cdrcCfg.FieldSeparator, cdrc.timezone, httpSkipTlsCheck, cdrs, filterS)
	cdrc.filterS = filterS
	return
}

type Cdrc struct {
	httpSkipTlsCheck     bool
	cdrcCfgs             []*config.CdrcCfg // All cdrc config profiles attached to this CDRC (key will be profile instance name)
	dfltCdrcCfg          *config.CdrcCfg
	timezone             string
	cdrs                 rpcclient.RpcClientConnection
	closeChan            chan struct{} // Used to signal config reloads when we need to span different CDRC-Client
	maxOpenFiles         chan struct{} // Maximum number of simultaneous files processed
	filterS              *engine.FilterS
	unpairedRecordsCache *UnpairedRecordsCache // Shared between all files in the folder we process
	partialRecordsCache  *PartialRecordsCache
}

// When called fires up folder monitoring, either automated via inotify or manual by sleeping between processing
func (self *Cdrc) Run() error {
	if self.dfltCdrcCfg.RunDelay == time.Duration(0) { // Automated via inotify
		return self.trackCDRFiles()
	}
	// Not automated, process and sleep approach
	for {
		select {
		case <-self.closeChan: // Exit, reinject closeChan for other CDRCs
			utils.Logger.Info(fmt.Sprintf("<Cdrc> Shutting down CDRC on path %s.", self.dfltCdrcCfg.CDRInPath))
			return nil
		default:
		}
		self.processCdrDir()
		time.Sleep(self.dfltCdrcCfg.RunDelay)
	}
}

// Watch the specified folder for file moves and parse the files on events
func (self *Cdrc) trackCDRFiles() (err error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return
	}
	defer watcher.Close()
	err = watcher.Add(self.dfltCdrcCfg.CDRInPath)
	if err != nil {
		return
	}
	utils.Logger.Info(fmt.Sprintf("<Cdrc> Monitoring %s for file moves.", self.dfltCdrcCfg.CDRInPath))
	for {
		select {
		case <-self.closeChan: // Exit, reinject closeChan for other CDRCs
			utils.Logger.Info(fmt.Sprintf("<Cdrc> Shutting down CDRC on path %s.", self.dfltCdrcCfg.CDRInPath))
			return nil
		case ev := <-watcher.Events:
			if ev.Op&fsnotify.Create == fsnotify.Create && (self.dfltCdrcCfg.CdrFormat != utils.MetaFScsv || path.Ext(ev.Name) != ".csv") {
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
	utils.Logger.Info(fmt.Sprintf("<Cdrc> Parsing folder %s for CDR files.", self.dfltCdrcCfg.CDRInPath))
	filesInDir, _ := ioutil.ReadDir(self.dfltCdrcCfg.CDRInPath)
	for _, file := range filesInDir {
		if self.dfltCdrcCfg.CdrFormat != utils.MetaFScsv || path.Ext(file.Name()) != ".csv" {
			go func() { //Enable async processing here
				if err := self.processFile(path.Join(self.dfltCdrcCfg.CDRInPath, file.Name())); err != nil {
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
	switch self.dfltCdrcCfg.CdrFormat {
	case utils.MetaFileCSV, utils.MetaFScsv, utils.MetaKamFlatstore, utils.MetaOsipsFlatstore, utils.MetaPartialCSV:
		csvReader := csv.NewReader(bufio.NewReader(file))
		csvReader.Comma = self.dfltCdrcCfg.FieldSeparator
		csvReader.Comment = '#'
		recordsProcessor = NewCsvRecordsProcessor(csvReader, self.timezone, fn, self.dfltCdrcCfg,
			self.cdrcCfgs, self.httpSkipTlsCheck,
			self.dfltCdrcCfg.CacheDumpFields, self.filterS, self.cdrs,
			self.unpairedRecordsCache, self.partialRecordsCache)
	case utils.MetaFileFWV:
		recordsProcessor = NewFwvRecordsProcessor(file, self.dfltCdrcCfg, self.cdrcCfgs,
			self.httpSkipTlsCheck, self.timezone, self.filterS)
	case utils.MetaFileXML:
		if recordsProcessor, err = NewXMLRecordsProcessor(file, self.dfltCdrcCfg.CDRRootPath,
			self.timezone, self.httpSkipTlsCheck, self.cdrcCfgs, self.filterS); err != nil {
			return err
		}
	default:
		return fmt.Errorf("Unsupported CDR format: %s", self.dfltCdrcCfg.CdrFormat)
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
			if err := self.cdrs.Call(utils.CDRsV1ProcessEvent,
				&engine.ArgV1ProcessEvent{CGREvent: *storedCdr.AsCGREvent()}, &reply); err != nil {
				utils.Logger.Err(fmt.Sprintf("<Cdrc> Failed sending CDR, %+v, error: %s", storedCdr, err.Error()))
			} else if reply != "OK" {
				utils.Logger.Err(fmt.Sprintf("<Cdrc> Received unexpected reply for CDR, %+v, reply: %s", storedCdr, reply))
			}
			cdrsPosted += 1
		}
	}
	// Finished with file, move it to processed folder
	newPath := path.Join(self.dfltCdrcCfg.CDROutPath, fn)
	if err := os.Rename(filePath, newPath); err != nil {
		utils.Logger.Err(err.Error())
		return err
	}
	utils.Logger.Info(fmt.Sprintf("Finished processing %s, moved to %s. Total records processed: %d, CDRs posted: %d, run duration: %s",
		fn, newPath, recordsProcessor.ProcessedRecordsNr(), cdrsPosted, time.Now().Sub(timeStart)))
	return nil
}
