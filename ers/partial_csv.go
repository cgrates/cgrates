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
	"io/ioutil"
	"os"
	"path"
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
		rdrExit:   rdrExit}

	var function func(itmID string, value interface{})
	if cfg.ERsCfg().Readers[cfgIdx].PartialCacheExpiryAction == utils.MetaDumpToFile {
		function = pCSVFileER.dumpToFile
	} else {
		function = pCSVFileER.postCDR
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
		return watchDir(rdr.rdrDir, rdr.processFile,
			utils.ERs, rdr.rdrExit)
	default:
		go func() {
			for {
				// Not automated, process and sleep approach
				select {
				case <-rdr.rdrExit:
					utils.Logger.Info(
						fmt.Sprintf("<%s> stop monitoring path <%s>",
							utils.ERs, rdr.rdrDir))
					return
				default:
				}
				filesInDir, _ := ioutil.ReadDir(rdr.rdrDir)
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
				time.Sleep(rdr.Config().RunDelay)
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
	csvReader.Comma = ','
	if len(rdr.Config().FieldSep) > 0 {
		csvReader.Comma = rune(rdr.Config().FieldSep[0])
	}
	csvReader.Comment = '#'
	rowNr := 0 // This counts the rows in the file, not really number of CDRs
	evsPosted := 0
	timeStart := time.Now()
	reqVars := make(map[string]interface{})
	for {
		var record []string
		if record, err = csvReader.Read(); err != nil {
			if err == io.EOF {
				break
			}
			return
		}
		rowNr++ // increment the rowNr after checking if it's not the end of file
		agReq := agents.NewAgentRequest(
			config.NewSliceDP(record, utils.EmptyString), reqVars,
			nil, nil, rdr.Config().Tenant,
			rdr.cgrCfg.GeneralCfg().DefaultTenant,
			utils.FirstNonEmpty(rdr.Config().Timezone,
				rdr.cgrCfg.GeneralCfg().DefaultTimezone),
			rdr.fltrS) // create an AgentRequest
		if pass, err := rdr.fltrS.Pass(agReq.Tenant, rdr.Config().Filters,
			agReq); err != nil || !pass {
			continue
		}
		navMp, err := agReq.AsNavigableMap(rdr.Config().ContentFields)
		if err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> reading file: <%s> row <%d>, ignoring due to error: <%s>",
					utils.ERs, absPath, rowNr, err.Error()))
			continue
		}

		// take OriginID and OriginHost to compose CGRID
		orgId, err := navMp.FieldAsString([]string{utils.OriginID})
		if err == utils.ErrNotFound {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> Missing <OriginID> field for row <%d> , <%s>",
					utils.ERs, rowNr, record))
			continue
		}
		orgHost, err := navMp.FieldAsString([]string{utils.OriginHost})
		if err == utils.ErrNotFound {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> Missing <OriginHost> field for row <%d> , <%s>",
					utils.ERs, rowNr, record))
			continue
		}
		cgrID := utils.Sha1(orgId, orgHost)
		// take Partial field from NavigableMap
		partial, _ := navMp.FieldAsString([]string{utils.Partial})
		if val, has := rdr.cache.Get(cgrID); !has {
			if utils.IsSliceMember([]string{"false", utils.EmptyString}, partial) { // complete CDR
				rdr.rdrEvents <- &erEvent{cgrEvent: navMp.AsCGREvent(agReq.Tenant, utils.NestingSep),
					rdrCfg: rdr.Config()}
				evsPosted++
			} else {
				rdr.cache.Set(cgrID,
					[]*engine.CGRSafEvent{engine.NewCGRSafEventFromCGREvent(navMp.AsCGREvent(agReq.Tenant, utils.NestingSep))}, nil)
			}
		} else {
			origCgrSafEvs := val.([]*engine.CGRSafEvent)

			if utils.IsSliceMember([]string{"false", utils.EmptyString}, partial) { // complete CDR
				rdr.rdrEvents <- &erEvent{cgrEvent: origCgrSafEv.AsCGREvent(),
					rdrCfg: rdr.Config()}
				evsPosted++
				rdr.cache.Remove(cgrID)
			} else {

				// overwrite the cache value with merged NavigableMap
				rdr.cache.Set(cgrID, origCgrSafEv, nil)
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

const (
	PartialRecordsSuffix = "partial"
)

func (rdr *PartialCSVFileER) dumpToFile(itmID string, value interface{}) {
	cgrSafEv := value.(*engine.CGRSafEvent)
	// complete CDR are handling in processFile function
	if partial, _ := cgrSafEv.Event.FieldAsString([]string{utils.Partial}); utils.IsSliceMember([]string{"false", utils.EmptyString}, partial) {
		return
	}
	cdr, err := cgrSafEv.Event.AsCDR(nil, cgrSafEv.Tenant, rdr.Config().Timezone)
	if err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> Converting Event : <%s> to cdr , ignoring due to error: <%s>",
				utils.ERs, utils.ToJSON(cgrSafEv), err.Error()))
		return
	}
	record, err := cdr.AsExportRecord(rdr.Config().CacheDumpFields, false, nil, rdr.fltrS)
	if err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> Converting CDR with CGRID: <%s> to record , ignoring due to error: <%s>",
				utils.ERs, cdr.CGRID, err.Error()))
		return
	}
	dumpFilePath := path.Join(rdr.Config().ProcessedPath, fmt.Sprintf("%s.%s.%d", cdr.OriginID, PartialRecordsSuffix, time.Now().Unix()))
	fileOut, err := os.Create(dumpFilePath)
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> Failed creating %s, error: %s", utils.ERs, dumpFilePath, err.Error()))
		return
	}
	csvWriter := csv.NewWriter(fileOut)
	csvWriter.Comma = utils.CSV_SEP
	if err := csvWriter.Write(record); err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> Failed writing partial record %v to file: %s, error: %s", utils.ERs, record, dumpFilePath, err.Error()))
		return
	}
	csvWriter.Flush()

}

func (rdr *PartialCSVFileER) postCDR(itmID string, value interface{}) {
	cgrSafEv := value.(*engine.CGRSafEvent)
	// complete CDR are handling in processFile function
	if partial, _ := cgrSafEv.Event.FieldAsString([]string{utils.Partial}); utils.IsSliceMember([]string{"false", utils.EmptyString}, partial) {
		return
	}
	// how to post incomplete CDR
	rdr.rdrEvents <- &erEvent{cgrEvent: cgrSafEv.AsCGREvent(), rdrCfg: rdr.Config()}
}
