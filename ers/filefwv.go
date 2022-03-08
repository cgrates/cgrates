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
	"bufio"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/agents"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func NewFWVFileER(cfg *config.CGRConfig, cfgIdx int,
	rdrEvents, partialEvents chan *erEvent, rdrErr chan error,
	fltrS *engine.FilterS, rdrExit chan struct{}) (er EventReader, err error) {
	srcPath := cfg.ERsCfg().Readers[cfgIdx].SourcePath
	if strings.HasSuffix(srcPath, utils.Slash) {
		srcPath = srcPath[:len(srcPath)-1]
	}
	fwvER := &FWVFileER{
		cgrCfg:        cfg,
		cfgIdx:        cfgIdx,
		fltrS:         fltrS,
		rdrDir:        srcPath,
		rdrEvents:     rdrEvents,
		partialEvents: partialEvents,
		rdrError:      rdrErr,
		rdrExit:       rdrExit,
		conReqs:       make(chan struct{}, cfg.ERsCfg().Readers[cfgIdx].ConcurrentReqs)}
	var processFile struct{}
	for i := 0; i < cfg.ERsCfg().Readers[cfgIdx].ConcurrentReqs; i++ {
		fwvER.conReqs <- processFile // Empty initiate so we do not need to wait later when we pop
	}
	return fwvER, nil
}

// FWVFileER implements EventReader interface for .fwv files
type FWVFileER struct {
	sync.RWMutex
	cgrCfg        *config.CGRConfig
	cfgIdx        int // index of config instance within ERsCfg.Readers
	fltrS         *engine.FilterS
	rdrDir        string
	rdrEvents     chan *erEvent // channel to dispatch the events created to
	partialEvents chan *erEvent // channel to dispatch the partial events created to
	rdrError      chan error
	rdrExit       chan struct{}
	conReqs       chan struct{} // limit number of opened files
	lineLen       int64         // Length of the line in the file
	offset        int64         // Index of the next byte to process
	headerOffset  int64
	trailerOffset int64 // Index where trailer starts, to be used as boundary when reading cdrs
	trailerLenght int64
	headerDP      utils.DataProvider
	trailerDP     utils.DataProvider
}

func (rdr *FWVFileER) Config() *config.EventReaderCfg {
	return rdr.cgrCfg.ERsCfg().Readers[rdr.cfgIdx]
}

func (rdr *FWVFileER) Serve() (err error) {
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
					if !strings.HasSuffix(file.Name(), utils.FWVSuffix) { // hardcoded file extension for xml event reader
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
func (rdr *FWVFileER) processFile(fPath, fName string) (err error) {
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

	rowNr := 0 // This counts the rows in the file, not really number of CDRs
	evsPosted := 0
	timeStart := time.Now()
	reqVars := &utils.DataNode{Type: utils.NMMapType, Map: map[string]*utils.DataNode{utils.FileName: utils.NewLeafNode(fName)}}

	for {
		var hasHeader, hasTrailer bool
		var headerFields, trailerFields []*config.FCTemplate
		if rdr.offset == 0 { // First time, set the necessary offsets
			// preprocess the fields for header and trailer
			for _, fld := range rdr.Config().Fields {
				if strings.HasPrefix(fld.Value[0].Rules, utils.DynamicDataPrefix+utils.MetaHdr) {
					hasHeader = true
					headerFields = append(headerFields, fld)
				}
				if strings.HasPrefix(fld.Value[0].Rules, utils.DynamicDataPrefix+utils.MetaTrl) {
					hasTrailer = true
					trailerFields = append(trailerFields, fld)
				}
			}

			if err = rdr.setLineLen(file, hasHeader, hasTrailer); err != nil {
				utils.Logger.Err(fmt.Sprintf("<%s> Row 0, error: cannot set lineLen: %s", utils.ERs, err.Error()))
				break
			}
			if hasTrailer {
				// process trailer here
				if err = rdr.processTrailer(file, rowNr, evsPosted, absPath, trailerFields); err != nil {
					utils.Logger.Err(fmt.Sprintf("<%s> Read trailer error: %s ", utils.ERs, err.Error()))
					return
				}
			}
			if hasHeader {
				if err = rdr.processHeader(file, rowNr, evsPosted, absPath, headerFields); err != nil {
					utils.Logger.Err(fmt.Sprintf("<%s> Row 0, error reading header: %s", utils.ERs, err.Error()))
					return
				}
				continue
			}
		}

		buf := make([]byte, rdr.lineLen)
		if nRead, err := file.Read(buf); err != nil {
			if err == io.EOF {
				break
			}
			return err
		} else if nRead != len(buf) && int64(nRead) != rdr.trailerLenght {
			utils.Logger.Err(fmt.Sprintf("<%s> Could not read complete line, have instead: %s", utils.ERs, string(buf)))
			rdr.offset += rdr.lineLen // increase the offset when exit
			continue
		}
		rowNr++ // increment the rowNr after checking if it's not the end of file
		record := string(buf)
		agReq := agents.NewAgentRequest(
			config.NewFWVProvider(record),
			reqVars, nil, nil, nil,
			rdr.Config().Tenant,
			rdr.cgrCfg.GeneralCfg().DefaultTenant,
			utils.FirstNonEmpty(rdr.Config().Timezone,
				rdr.cgrCfg.GeneralCfg().DefaultTimezone),
			rdr.fltrS, map[string]utils.DataProvider{utils.MetaHdr: rdr.headerDP, utils.MetaTrl: rdr.trailerDP}) // create an AgentRequest
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
			rdr.offset += rdr.lineLen // increase the offset when exit
			return
		}
		rdr.offset += rdr.lineLen // increase the offset
		cgrEv := utils.NMAsCGREvent(agReq.CGRRequest, agReq.Tenant, utils.NestingSep, agReq.Opts)
		rdrEv := rdr.rdrEvents
		if _, isPartial := cgrEv.APIOpts[utils.PartialOpt]; isPartial {
			rdrEv = rdr.partialEvents
		}
		rdrEv <- &erEvent{
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

// Sets the line length based on first line, sets offset back to initial after reading
func (rdr *FWVFileER) setLineLen(file *os.File, hasHeader, hasTrailer bool) error {
	buff := bufio.NewReader(file)
	// in case we have header we take the length of first line and add it as headerOffset
	i := 0
	lastLineSize := 0
	for {
		readBytes, err := buff.ReadBytes('\n')
		if err != nil {
			break
		}
		if hasHeader && i == 0 {
			rdr.headerOffset = int64(len(readBytes))
			i++
			continue
		}
		if (hasHeader && i == 1) || (!hasHeader && i == 0) {
			rdr.lineLen = int64(len(readBytes))
			i++
			continue
		}
		lastLineSize = len(readBytes)
	}
	if hasTrailer {
		fi, err := file.Stat()
		if err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> Row 0, error: cannot get file stats: %s", utils.ERs, err.Error()))
			return err
		}
		rdr.trailerOffset = fi.Size() - int64(lastLineSize)
		rdr.trailerLenght = int64(lastLineSize)
	}

	if _, err := file.Seek(0, 0); err != nil {
		return err
	}
	return nil
}

func (rdr *FWVFileER) processTrailer(file *os.File, rowNr, evsPosted int, absPath string, trailerFields []*config.FCTemplate) (err error) {
	buf := make([]byte, rdr.trailerLenght)
	if nRead, err := file.ReadAt(buf, rdr.trailerOffset); err != nil && err != io.EOF {
		return err
	} else if nRead != len(buf) {
		return fmt.Errorf("In trailer, line len: %d, have read: %d instead of: %d", rdr.trailerOffset, nRead, len(buf))
	}
	record := string(buf)
	rdr.trailerDP = config.NewFWVProvider(record)
	agReq := agents.NewAgentRequest(
		nil, nil, nil, nil, nil,
		rdr.Config().Tenant,
		rdr.cgrCfg.GeneralCfg().DefaultTenant,
		utils.FirstNonEmpty(rdr.Config().Timezone,
			rdr.cgrCfg.GeneralCfg().DefaultTimezone),
		rdr.fltrS, map[string]utils.DataProvider{utils.MetaTrl: rdr.trailerDP}) // create an AgentRequest
	if pass, err := rdr.fltrS.Pass(context.TODO(), agReq.Tenant, rdr.Config().Filters,
		agReq); err != nil || !pass {
		return nil
	}
	if err := agReq.SetFields(trailerFields); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> reading file: <%s> row <%d>, ignoring due to error: <%s>",
				utils.ERs, absPath, rowNr, err.Error()))
		return err
	}
	cgrEv := utils.NMAsCGREvent(agReq.CGRRequest, agReq.Tenant, utils.NestingSep, agReq.Opts)
	rdrEv := rdr.rdrEvents
	if _, isPartial := cgrEv.APIOpts[utils.PartialOpt]; isPartial {
		rdrEv = rdr.partialEvents
	}
	rdrEv <- &erEvent{
		cgrEvent: cgrEv,
		rdrCfg:   rdr.Config(),
	}
	evsPosted++
	// reset the cursor after process the trailer
	_, err = file.Seek(0, 0)
	return
}

func (rdr *FWVFileER) processHeader(file *os.File, rowNr, evsPosted int, absPath string, hdrFields []*config.FCTemplate) error {
	buf := make([]byte, rdr.headerOffset)
	if nRead, err := file.Read(buf); err != nil {
		return err
	} else if nRead != len(buf) {
		return fmt.Errorf("In header, line len: %d, have read: %d", rdr.headerOffset, nRead)
	}
	return rdr.createHeaderMap(string(buf), rowNr, evsPosted, absPath, hdrFields)
}

func (rdr *FWVFileER) createHeaderMap(record string, rowNr, evsPosted int, absPath string, hdrFields []*config.FCTemplate) (err error) {
	rdr.headerDP = config.NewFWVProvider(record)
	agReq := agents.NewAgentRequest(
		nil, nil, nil, nil, nil,
		rdr.Config().Tenant,
		rdr.cgrCfg.GeneralCfg().DefaultTenant,
		utils.FirstNonEmpty(rdr.Config().Timezone,
			rdr.cgrCfg.GeneralCfg().DefaultTimezone),
		rdr.fltrS, map[string]utils.DataProvider{utils.MetaHdr: rdr.headerDP}) // create an AgentRequest
	if pass, err := rdr.fltrS.Pass(context.TODO(), agReq.Tenant, rdr.Config().Filters,
		agReq); err != nil || !pass {
		return nil
	}
	if err := agReq.SetFields(hdrFields); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> reading file: <%s> row <%d>, ignoring due to error: <%s>",
				utils.ERs, absPath, rowNr, err.Error()))
		rdr.offset += rdr.lineLen // increase the offset when exit
		return err
	}
	rdr.offset += rdr.headerOffset // increase the offset
	cgrEv := utils.NMAsCGREvent(agReq.CGRRequest, agReq.Tenant, utils.NestingSep, agReq.Opts)
	rdrEv := rdr.rdrEvents
	if _, isPartial := cgrEv.APIOpts[utils.PartialOpt]; isPartial {
		rdrEv = rdr.partialEvents
	}
	rdrEv <- &erEvent{
		cgrEvent: cgrEv,
		rdrCfg:   rdr.Config(),
	}
	evsPosted++
	return
}
