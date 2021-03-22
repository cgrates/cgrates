/*
Real-time Online/Offline Charging System (OerS) for Telecom & ISP environments
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

package ers

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/cgrates/cgrates/agents"

	"github.com/cgrates/ltcache"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func NewPartialCSVFileER(cfg *config.CGRConfig, cfgIdx int,
	rdrEvents chan *erEvent, rdrErr chan error,
	fltrS *engine.FilterS, rdrExit chan struct{}) (er EventReader, err error) {
	srcPath := cfg.ERsCfg().Readers[cfgIdx].SourcePath
	if strings.HasSuffix(srcPath, utils.Slash) {
		srcPath = srcPath[:len(srcPath)-1]
	}

	pCSVFileER := &PartialCSVFileER{
		cgrCfg:    cfg,
		cfgIdx:    cfgIdx,
		fltrS:     fltrS,
		rdrDir:    srcPath,
		rdrEvents: rdrEvents,
		rdrError:  rdrErr,
		rdrExit:   rdrExit,
		conReqs:   make(chan struct{}, cfg.ERsCfg().Readers[cfgIdx].ConcurrentReqs)}

	var function func(itmID string, value interface{})
	if cfg.ERsCfg().Readers[cfgIdx].PartialCacheExpiryAction == utils.MetaDumpToFile {
		function = pCSVFileER.dumpToFile
	} else {
		function = pCSVFileER.postCDR
	}
	var processFile struct{}
	for i := 0; i < cfg.ERsCfg().Readers[cfgIdx].ConcurrentReqs; i++ {
		pCSVFileER.conReqs <- processFile // Empty initiate so we do not need to wait later when we pop
	}
	pCSVFileER.cache = ltcache.NewCache(ltcache.UnlimitedCaching, cfg.ERsCfg().Readers[cfgIdx].PartialRecordCache, false, function)
	return pCSVFileER, nil
}

// CSVFileER implements EventReader interface for .csv files
type PartialCSVFileER struct {
	sync.RWMutex
	cgrCfg    *config.CGRConfig
	cfgIdx    int // index of config instance within ERsCfg.Readers
	fltrS     *engine.FilterS
	cache     *ltcache.Cache
	rdrDir    string
	rdrEvents chan *erEvent // channel to dispatch the events created to
	rdrError  chan error
	rdrExit   chan struct{}
	conReqs   chan struct{} // limit number of opened files
}

func (rdr *PartialCSVFileER) Config() *config.EventReaderCfg {
	return rdr.cgrCfg.ERsCfg().Readers[rdr.cfgIdx]
}

func (rdr *PartialCSVFileER) Serve() (err error) {
	switch rdr.Config().RunDelay {
	case time.Duration(0): // 0 disables the automatic read, maybe done per API
		return
	case time.Duration(-1):
		return utils.WatchDir(rdr.rdrDir, rdr.processFile,
			utils.ERs, rdr.rdrExit)
	default:
		go func() {
			tm := time.NewTimer(0)
			for {
				// Not automated, process and sleep approach
				select {
				case <-rdr.rdrExit:
					tm.Stop()
					utils.Logger.Info(
						fmt.Sprintf("<%s> stop monitoring path <%s>",
							utils.ERs, rdr.rdrDir))
					return
				case <-tm.C:
				}
				filesInDir, _ := os.ReadDir(rdr.rdrDir)
				for _, file := range filesInDir {
					if !strings.HasSuffix(file.Name(), utils.CSVSuffix) { // hardcoded file extension for csv event reader
						continue // used in order to filter the files from directory
					}
					go func(fileName string) {
						if err := rdr.processFile(rdr.rdrDir, fileName); err != nil {
							utils.Logger.Warning(
								fmt.Sprintf("<%s> processing file %s, error: %s",
									utils.ERs, fileName, err.Error()))
						}
					}(file.Name())
				}
				tm.Reset(rdr.Config().RunDelay)
			}
		}()
	}
	return
}

// processFile is called for each file in a directory and dispatches erEvents from it
func (rdr *PartialCSVFileER) processFile(fPath, fName string) (err error) {
	if cap(rdr.conReqs) != 0 { // 0 goes for no limit
		processFile := <-rdr.conReqs // Queue here for maxOpenFiles
		defer func() { rdr.conReqs <- processFile }()
	}
	absPath := path.Join(fPath, fName)
	utils.Logger.Info(
		fmt.Sprintf("<%s> parsing <%s>", utils.ERs, absPath))
	var file *os.File
	if file, err = os.Open(absPath); err != nil {
		return
	}
	defer file.Close()
	csvReader := csv.NewReader(bufio.NewReader(file))
	csvReader.FieldsPerRecord = rdr.cgrCfg.ERsCfg().Readers[rdr.cfgIdx].RowLength
	csvReader.Comma = ','
	if len(rdr.Config().FieldSep) > 0 {
		csvReader.Comma = rune(rdr.Config().FieldSep[0])
	}
	csvReader.Comment = '#'
	var indxAls map[string]int
	rowNr := 0 // This counts the rows in the file, not really number of CDRs
	evsPosted := 0
	timeStart := time.Now()
	reqVars := utils.NavigableMap{utils.FileName: utils.NewNMData(fName)}
	for {
		var record []string
		if record, err = csvReader.Read(); err != nil {
			if err == io.EOF {
				break
			}
			return
		}
		if rowNr == 0 && len(record) > 0 &&
			strings.HasPrefix(record[0], rdr.cgrCfg.ERsCfg().Readers[rdr.cfgIdx].HeaderDefineChar) {
			record[0] = strings.TrimPrefix(record[0], rdr.cgrCfg.ERsCfg().Readers[rdr.cfgIdx].HeaderDefineChar)
			// map the templates
			indxAls = make(map[string]int)
			for i, hdr := range record {
				indxAls[hdr] = i
			}
			continue
		}
		rowNr++ // increment the rowNr after checking if it's not the end of file
		agReq := agents.NewAgentRequest(
			config.NewSliceDP(record, indxAls), reqVars,
			nil, nil, nil, rdr.Config().Tenant,
			rdr.cgrCfg.GeneralCfg().DefaultTenant,
			utils.FirstNonEmpty(rdr.Config().Timezone,
				rdr.cgrCfg.GeneralCfg().DefaultTimezone),
			rdr.fltrS, nil, nil) // create an AgentRequest
		if pass, err := rdr.fltrS.Pass(agReq.Tenant, rdr.Config().Filters,
			agReq); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> reading file: <%s> row <%d>, ignoring due to filter error: <%s>",
					utils.ERs, absPath, rowNr, err.Error()))
			return err
		} else if !pass {
			continue
		}
		if err = agReq.SetFields(rdr.Config().Fields); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> reading file: <%s> row <%d>, ignoring due to error: <%s>",
					utils.ERs, absPath, rowNr, err.Error()))
			return
		}

		// take OriginID and OriginHost to compose CGRID
		orgID, err := agReq.CGRRequest.FieldAsString([]string{utils.OriginID})
		if err == utils.ErrNotFound {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> Missing <OriginID> field for row <%d> , <%s>",
					utils.ERs, rowNr, record))
			continue
		}
		orgHost, err := agReq.CGRRequest.FieldAsString([]string{utils.OriginHost})
		if err == utils.ErrNotFound {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> Missing <OriginHost> field for row <%d> , <%s>",
					utils.ERs, rowNr, record))
			continue
		}
		cgrID := utils.Sha1(orgID, orgHost)
		// take Partial field from NavigableMap
		partial, _ := agReq.CGRRequest.FieldAsString([]string{utils.Partial})
		cgrEv := config.NMAsCGREvent(agReq.CGRRequest, agReq.Tenant, utils.NestingSep, agReq.Opts)
		if val, has := rdr.cache.Get(cgrID); !has {
			if utils.IsSliceMember([]string{"false", utils.EmptyString}, partial) { // complete CDR
				rdr.rdrEvents <- &erEvent{
					cgrEvent: cgrEv,
					rdrCfg:   rdr.Config(),
				}
				evsPosted++
			} else {
				rdr.cache.Set(cgrID,
					[]*utils.CGREvent{cgrEv}, nil)
			}
		} else {
			origCgrEvs := val.([]*utils.CGREvent)
			origCgrEvs = append(origCgrEvs, cgrEv)
			if utils.IsSliceMember([]string{"false", utils.EmptyString}, partial) { // complete CDR
				//sort CGREvents based on AnswertTime and SetupTime
				sort.Slice(origCgrEvs, func(i, j int) bool {
					aTime, err := origCgrEvs[i].FieldAsTime(utils.AnswerTime, agReq.Timezone)
					if err != nil && err == utils.ErrNotFound {
						sTime, _ := origCgrEvs[i].FieldAsTime(utils.SetupTime, agReq.Timezone)
						sTime2, _ := origCgrEvs[j].FieldAsTime(utils.SetupTime, agReq.Timezone)
						return sTime.Before(sTime2)
					}
					aTime2, _ := origCgrEvs[j].FieldAsTime(utils.AnswerTime, agReq.Timezone)
					return aTime.Before(aTime2)
				})
				// compose the CGREvent from slice
				cgrEv := new(utils.CGREvent)
				cgrEv.ID = utils.UUIDSha1Prefix()
				cgrEv.Time = utils.TimePointer(time.Now())
				cgrEv.APIOpts = map[string]interface{}{}
				for i, origCgrEv := range origCgrEvs {
					if i == 0 {
						cgrEv.Tenant = origCgrEv.Tenant
					}
					for key, value := range origCgrEv.Event {
						cgrEv.Event[key] = value
					}
					for key, val := range origCgrEv.APIOpts {
						cgrEv.APIOpts[key] = val
					}
				}
				rdr.rdrEvents <- &erEvent{
					cgrEvent: cgrEv,
					rdrCfg:   rdr.Config(),
				}
				evsPosted++
				rdr.cache.Set(cgrID, nil, nil)
				rdr.cache.Remove(cgrID)
			} else {
				// overwrite the cache value with merged NavigableMap
				rdr.cache.Set(cgrID, origCgrEvs, nil)
			}
		}

	}
	if rdr.Config().ProcessedPath != "" {
		// Finished with file, move it to processed folder
		outPath := path.Join(rdr.Config().ProcessedPath, fName)
		if err = os.Rename(absPath, outPath); err != nil {
			return
		}
	}

	utils.Logger.Info(
		fmt.Sprintf("%s finished processing file <%s>. Total records processed: %d, events posted: %d, run duration: %s",
			utils.ERs, absPath, rowNr, evsPosted, time.Now().Sub(timeStart)))
	return
}

func (rdr *PartialCSVFileER) dumpToFile(itmID string, value interface{}) {
	tmz := utils.FirstNonEmpty(rdr.Config().Timezone,
		rdr.cgrCfg.GeneralCfg().DefaultTimezone)
	origCgrEvs := value.([]*utils.CGREvent)
	for _, origCgrEv := range origCgrEvs {
		// complete CDR are handling in processFile function
		if partial, _ := origCgrEv.FieldAsString(utils.Partial); utils.IsSliceMember([]string{"false", utils.EmptyString}, partial) {
			return
		}
	}
	// Need to process the first event separate to take the name for the file
	cdr, err := engine.NewMapEvent(origCgrEvs[0].Event).AsCDR(rdr.cgrCfg, origCgrEvs[0].Tenant, tmz)
	if err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> Converting Event : <%s> to cdr , ignoring due to error: <%s>",
				utils.ERs, utils.ToJSON(origCgrEvs[0].Event), err.Error()))
		return
	}
	record, err := cdr.AsExportRecord(rdr.Config().CacheDumpFields, nil, rdr.fltrS)
	if err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> Converting CDR with CGRID: <%s> to record , ignoring due to error: <%s>",
				utils.ERs, cdr.CGRID, err.Error()))
		return
	}
	dumpFilePath := path.Join(rdr.Config().ProcessedPath, fmt.Sprintf("%s%s.%d",
		cdr.OriginID, utils.TmpSuffix, time.Now().Unix()))
	fileOut, err := os.Create(dumpFilePath)
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> Failed creating %s, error: %s",
			utils.ERs, dumpFilePath, err.Error()))
		return
	}
	csvWriter := csv.NewWriter(fileOut)
	csvWriter.Comma = rune(rdr.Config().FieldSep[0])
	if err = csvWriter.Write(record); err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> Failed writing partial record %v to file: %s, error: %s",
			utils.ERs, record, dumpFilePath, err.Error()))
		return
	}
	if len(origCgrEvs) > 1 {
		for _, origCgrEv := range origCgrEvs[1:] {
			// Need to process the first event separate to take the name for the file
			cdr, err = engine.NewMapEvent(origCgrEv.Event).AsCDR(rdr.cgrCfg, origCgrEv.Tenant, tmz)
			if err != nil {
				utils.Logger.Warning(
					fmt.Sprintf("<%s> Converting Event : <%s> to cdr , ignoring due to error: <%s>",
						utils.ERs, utils.ToJSON(origCgrEv.Event), err.Error()))
				return
			}
			record, err = cdr.AsExportRecord(rdr.Config().CacheDumpFields, nil, rdr.fltrS)
			if err != nil {
				utils.Logger.Warning(
					fmt.Sprintf("<%s> Converting CDR with CGRID: <%s> to record , ignoring due to error: <%s>",
						utils.ERs, cdr.CGRID, err.Error()))
				return
			}
			if err = csvWriter.Write(record); err != nil {
				utils.Logger.Err(fmt.Sprintf("<%s> Failed writing partial record %v to file: %s, error: %s",
					utils.ERs, record, dumpFilePath, err.Error()))
				return
			}
		}
	}

	csvWriter.Flush()
}

func (rdr *PartialCSVFileER) postCDR(itmID string, value interface{}) {
	if value == nil {
		return
	}
	tmz := utils.FirstNonEmpty(rdr.Config().Timezone,
		rdr.cgrCfg.GeneralCfg().DefaultTimezone)
	origCgrEvs := value.([]*utils.CGREvent)
	for _, origCgrEv := range origCgrEvs {
		// complete CDR are handling in processFile function
		if partial, _ := origCgrEv.FieldAsString(utils.Partial); utils.IsSliceMember([]string{"false", utils.EmptyString}, partial) {
			return
		}
	}

	// how to post incomplete CDR
	//sort CGREvents based on AnswertTime and SetupTime
	sort.Slice(origCgrEvs, func(i, j int) bool {
		aTime, err := origCgrEvs[i].FieldAsTime(utils.AnswerTime, tmz)
		if err != nil && err == utils.ErrNotFound {
			sTime, _ := origCgrEvs[i].FieldAsTime(utils.SetupTime, tmz)
			sTime2, _ := origCgrEvs[j].FieldAsTime(utils.SetupTime, tmz)
			return sTime.Before(sTime2)
		}
		aTime2, _ := origCgrEvs[j].FieldAsTime(utils.AnswerTime, tmz)
		return aTime.Before(aTime2)
	})
	// compose the CGREvent from slice
	cgrEv := &utils.CGREvent{
		ID:      utils.UUIDSha1Prefix(),
		Time:    utils.TimePointer(time.Now()),
		Event:   make(map[string]interface{}),
		APIOpts: make(map[string]interface{}),
	}
	for i, origCgrEv := range origCgrEvs {
		if i == 0 {
			cgrEv.Tenant = origCgrEv.Tenant
		}
		for key, value := range origCgrEv.Event {
			cgrEv.Event[key] = value
		}
		for key, val := range origCgrEv.APIOpts {
			cgrEv.APIOpts[key] = val
		}
	}
	rdr.rdrEvents <- &erEvent{
		cgrEvent: cgrEv,
		rdrCfg:   rdr.Config(),
	}
}
