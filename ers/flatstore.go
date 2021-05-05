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

package ers

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/ltcache"

	"github.com/cgrates/cgrates/agents"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

type fstRecord struct {
	method   string
	values   []string
	fileName string
}

func NewFlatstoreER(cfg *config.CGRConfig, cfgIdx int,
	rdrEvents chan *erEvent, rdrErr chan error,
	fltrS *engine.FilterS, rdrExit chan struct{}) (er EventReader, err error) {
	srcPath := cfg.ERsCfg().Readers[cfgIdx].SourcePath
	if strings.HasSuffix(srcPath, utils.Slash) {
		srcPath = srcPath[:len(srcPath)-1]
	}
	flatER := &FlatstoreER{
		cgrCfg:    cfg,
		cfgIdx:    cfgIdx,
		fltrS:     fltrS,
		rdrDir:    srcPath,
		rdrEvents: rdrEvents,
		rdrError:  rdrErr,
		rdrExit:   rdrExit,
		conReqs:   make(chan struct{}, cfg.ERsCfg().Readers[cfgIdx].ConcurrentReqs),
	}
	var processFile struct{}
	for i := 0; i < cfg.ERsCfg().Readers[cfgIdx].ConcurrentReqs; i++ {
		flatER.conReqs <- processFile // Empty initiate so we do not need to wait later when we pop
	}
	var ttl time.Duration
	if ttlOpt, has := flatER.Config().Opts[utils.FstPartialRecordCacheOpt]; has {
		if ttl, err = utils.IfaceAsDuration(ttlOpt); err != nil {
			return
		}
	}
	flatER.cache = ltcache.NewCache(ltcache.UnlimitedCaching, ttl, false, flatER.dumpToFile)
	return flatER, nil
}

// FlatstoreER implements EventReader interface for Flatstore CDR
type FlatstoreER struct {
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

func (rdr *FlatstoreER) Config() *config.EventReaderCfg {
	return rdr.cgrCfg.ERsCfg().Readers[rdr.cfgIdx]
}

func (rdr *FlatstoreER) Serve() (err error) {
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
func (rdr *FlatstoreER) processFile(fPath, fName string) (err error) {
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
	var csvReader *csv.Reader
	if csvReader, err = newCSVReader(file, rdr.Config().Opts, utils.FlatstorePrfx); err != nil {
		utils.Logger.Err(
			fmt.Sprintf("<%s> failed creating flatStore reader for <%s>, due to option parsing error: <%s>",
				utils.ERs, rdr.Config().ID, err.Error()))
		return
	}
	rowNr := 0 // This counts the rows in the file, not really number of CDRs
	evsPosted := 0
	timeStart := time.Now()
	reqVars := &utils.DataNode{Type: utils.NMMapType, Map: map[string]*utils.DataNode{utils.FileName: utils.NewLeafNode(fName)}}
	faildCallPrfx := utils.IfaceAsString(rdr.Config().Opts[utils.FstFailedCallsPrefixOpt])
	failedCallsFile := len(faildCallPrfx) != 0 && strings.HasPrefix(fName, faildCallPrfx)
	var methodTmp config.RSRParsers
	if methodTmp, err = config.NewRSRParsers(utils.IfaceAsString(rdr.Config().Opts[utils.FstMethodOpt]), rdr.cgrCfg.GeneralCfg().RSRSep); err != nil {
		return
	}
	var originTmp config.RSRParsers
	if originTmp, err = config.NewRSRParsers(utils.IfaceAsString(rdr.Config().Opts[utils.FstOriginIDOpt]), rdr.cgrCfg.GeneralCfg().RSRSep); err != nil {
		return
	}
	var mandatoryAcK bool
	if mandatoryAcK, err = utils.IfaceAsBool(rdr.Config().Opts[utils.FstMadatoryACKOpt]); err != nil {
		return
	}
	for {
		var record []string
		if record, err = csvReader.Read(); err != nil {
			if err == io.EOF {
				break
			}
			return
		}
		req := config.NewSliceDP(record, nil)
		tmpReq := utils.MapStorage{utils.MetaReq: req}
		var method string
		if method, err = methodTmp.ParseDataProvider(tmpReq); err != nil {
			return
		} else if method != utils.FstInvite &&
			method != utils.FstBye &&
			method != utils.FstAck {
			return fmt.Errorf("unsupported method: <%q>", method)
		}

		var originID string
		if originID, err = originTmp.ParseDataProvider(tmpReq); err != nil {
			return
		}

		cacheKey := utils.ConcatenatedKey(originID, method)
		if rdr.cache.HasItem(cacheKey) {
			utils.Logger.Warning(fmt.Sprintf("<%s> Overwriting the %s method for record <%s>", utils.ERs, method, originID))
			rdr.cache.Set(cacheKey, &fstRecord{method: method, values: record, fileName: fName}, []string{originID})
			continue
		}
		records := rdr.cache.GetGroupItems(originID)

		if lrecords := len(records); !failedCallsFile && // do not set in cache if we know that the calls are failed
			(lrecords == 0 ||
				(mandatoryAcK && lrecords != 2) ||
				(!mandatoryAcK && lrecords != 1)) {
			rdr.cache.Set(cacheKey, &fstRecord{method: method, values: record, fileName: fName}, []string{originID})
			continue
		}
		extraDP := map[string]utils.DataProvider{utils.FstMethodToPrfx[method]: req}
		for _, record := range records {
			req := record.(*fstRecord)
			rdr.cache.Set(utils.ConcatenatedKey(originID, req.method), nil, []string{originID})
			extraDP[utils.FstMethodToPrfx[req.method]] = config.NewSliceDP(req.values, nil)
		}
		rdr.cache.RemoveGroup(originID)

		rowNr++ // increment the rowNr after checking if it's not the end of file
		agReq := agents.NewAgentRequest(
			req, reqVars,
			nil, nil, nil, rdr.Config().Tenant,
			rdr.cgrCfg.GeneralCfg().DefaultTenant,
			utils.FirstNonEmpty(rdr.Config().Timezone,
				rdr.cgrCfg.GeneralCfg().DefaultTimezone),
			rdr.fltrS, nil) // create an AgentRequest
		if pass, err := rdr.fltrS.Pass(context.TODO(), agReq.Tenant, rdr.Config().Filters,
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

		cgrEv := utils.NMAsCGREvent(agReq.CGRRequest, agReq.Tenant, utils.NestingSep, agReq.Opts)
		rdr.rdrEvents <- &erEvent{
			cgrEvent: cgrEv,
			rdrCfg:   rdr.Config(),
		}
		evsPosted++
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

func (rdr *FlatstoreER) dumpToFile(itmID string, value interface{}) {
	if value == nil {
		return
	}
	unpRcd := value.(*fstRecord)
	dumpFilePath := path.Join(rdr.Config().ProcessedPath, unpRcd.fileName+utils.TmpSuffix)
	fileOut, err := os.Create(dumpFilePath)
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> Failed creating %s, error: %s",
			utils.ERs, dumpFilePath, err.Error()))
		return
	}
	csvWriter := csv.NewWriter(fileOut)
	csvWriter.Comma = rune(utils.IfaceAsString(rdr.Config().Opts[utils.FlatstorePrfx+utils.FieldSepOpt])[0])
	if err = csvWriter.Write(unpRcd.values); err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> Failed writing partial record %v to file: %s, error: %s",
			utils.ERs, unpRcd.values, dumpFilePath, err.Error()))
		// return // let it close the opened file
	}
	csvWriter.Flush()
	fileOut.Close()
}
