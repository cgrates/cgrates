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
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cgrates/ltcache"

	"github.com/cgrates/cgrates/agents"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

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

func (rdr *FlatstoreER) serveDefault() {
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
}

func (rdr *FlatstoreER) Serve() (err error) {
	switch rdr.Config().RunDelay {
	case time.Duration(0): // 0 disables the automatic read, maybe done per API
		return
	case time.Duration(-1):
		return utils.WatchDir(rdr.rdrDir, rdr.processFile,
			utils.ERs, rdr.rdrExit)
	default:
		go rdr.serveDefault()
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
	failCallPrfx := utils.IfaceAsString(rdr.Config().Opts[utils.FstFailedCallsPrefixOpt])
	for {
		var record []string
		if record, err = csvReader.Read(); err != nil {
			if err == io.EOF {
				err = nil //If it reaches the end of the file, return nil
				break
			}
			return
		}
		if strings.HasPrefix(fName, failCallPrfx) { // Use the first index since they should be the same in all configs
			record = append(record, "0") // Append duration 0 for failed calls flatstore CDR
		} else {
			pr, err := NewUnpairedRecord(record, utils.FirstNonEmpty(rdr.Config().Timezone,
				rdr.cgrCfg.GeneralCfg().DefaultTimezone), fName)
			if err != nil {
				utils.Logger.Warning(
					fmt.Sprintf("<%s> Converting row : <%s> to unpairedRecord , ignoring due to error: <%s>",
						utils.ERs, record, err.Error()))
				continue
			}
			if val, has := rdr.cache.Get(pr.OriginID); !has {
				rdr.cache.Set(pr.OriginID, pr, nil)
				continue
			} else {
				pair := val.(*UnpairedRecord)
				record, err = pairToRecord(pair, pr)
				if err != nil {
					utils.Logger.Warning(
						fmt.Sprintf("<%s> Merging unpairedRecords : <%s> and <%s> to record , ignoring due to error: <%s>",
							utils.ERs, utils.ToJSON(pair), utils.ToJSON(pr), err.Error()))
					continue
				}
				rdr.cache.Set(pr.OriginID, nil, nil)
				rdr.cache.Remove(pr.OriginID)
			}
		}

		// build Usage from Fields based on record lenght
		for i, cntFld := range rdr.Config().Fields {
			if cntFld.Path == utils.MetaCgreq+utils.NestingSep+utils.Usage {
				rdr.Config().Fields[i].Value = config.NewRSRParsersMustCompile("~*req."+strconv.Itoa(len(record)-1), utils.InfieldSep) // in case of flatstore, last element will be the duration computed by us
			}
		}
		rowNr++ // increment the rowNr after checking if it's not the end of file
		agReq := agents.NewAgentRequest(
			config.NewSliceDP(record, nil), reqVars,
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
			utils.ERs, absPath, rowNr, evsPosted, time.Since(timeStart)))
	return
}

func NewUnpairedRecord(record []string, timezone string, fileName string) (*UnpairedRecord, error) {
	if len(record) < 7 {
		return nil, errors.New("MISSING_IE")
	}
	pr := &UnpairedRecord{Method: record[0], OriginID: record[3] + record[1] + record[2], Values: record, FileName: fileName}
	var err error
	if pr.Timestamp, err = utils.ParseTimeDetectLayout(record[6], timezone); err != nil {
		return nil, err
	}
	return pr, nil
}

// UnpairedRecord is a partial record received from Flatstore, can be INVITE or BYE and it needs to be paired in order to produce duration
type UnpairedRecord struct {
	Method    string    // INVITE or BYE
	OriginID  string    // Copute here the OriginID
	Timestamp time.Time // Timestamp of the event, as written by db_flastore module
	Values    []string  // Can contain original values or updated via UpdateValues
	FileName  string
}

// Pairs INVITE and BYE into final record containing as last element the duration
func pairToRecord(part1, part2 *UnpairedRecord) ([]string, error) {
	var invite, bye *UnpairedRecord
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

func (rdr *FlatstoreER) dumpToFile(itmID string, value interface{}) {
	if value == nil {
		return
	}
	unpRcd := value.(*UnpairedRecord)

	dumpFilePath := path.Join(rdr.Config().ProcessedPath, unpRcd.FileName+utils.TmpSuffix)
	fileOut, err := os.Create(dumpFilePath)
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> Failed creating %s, error: %s",
			utils.ERs, dumpFilePath, err.Error()))
		return
	}
	csvWriter := csv.NewWriter(fileOut)
	csvWriter.Comma = rune(utils.IfaceAsString(rdr.Config().Opts[utils.FlatstorePrfx+utils.FieldSepOpt])[0])
	if err = csvWriter.Write(unpRcd.Values); err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> Failed writing partial record %v to file: %s, error: %s",
			utils.ERs, unpRcd.Values, dumpFilePath, err.Error()))
		return
	}

	csvWriter.Flush()
}
