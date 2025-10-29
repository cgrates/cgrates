/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
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

	"github.com/cgrates/cgrates/agents"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func NewFWVFileERER(cfg *config.CGRConfig, cfgIdx int,
	rdrEvents chan *erEvent, rdrErr chan error,
	fltrS *engine.FilterS, rdrExit chan struct{}) (er EventReader, err error) {
	srcPath := cfg.ERsCfg().Readers[cfgIdx].SourcePath
	if strings.HasSuffix(srcPath, utils.Slash) {
		srcPath = srcPath[:len(srcPath)-1]
	}
	fwvER := &FWVFileER{
		cgrCfg:    cfg,
		cfgIdx:    cfgIdx,
		fltrS:     fltrS,
		rdrDir:    srcPath,
		rdrEvents: rdrEvents,
		rdrError:  rdrErr,
		rdrExit:   rdrExit,
		conReqs:   make(chan struct{}, cfg.ERsCfg().Readers[cfgIdx].ConcurrentReqs)}
	var processFile struct{}
	for i := 0; i < cfg.ERsCfg().Readers[cfgIdx].ConcurrentReqs; i++ {
		fwvER.conReqs <- processFile // Empty initiate so we do not need to wait later when we pop
	}
	return fwvER, nil
}

// XMLFileER implements EventReader interface for .xml files
type FWVFileER struct {
	sync.RWMutex
	cgrCfg    *config.CGRConfig
	cfgIdx    int // index of config instance within ERsCfg.Readers
	fltrS     *engine.FilterS
	rdrDir    string
	rdrEvents chan *erEvent // channel to dispatch the events created to
	rdrError  chan error
	rdrExit   chan struct{}
	conReqs   chan struct{} // limit number of opened files
}

type fileVars struct {
	offset        int64 // index of the next byte to process
	lineLength    int64 // length of a line in the file
	headerOffset  int64
	trailerOffset int64 // index where trailer starts, to be used as boundary when reading cdrs
	trailerLength int64
	path          string // absolute path of the file
	headerDP      utils.DataProvider
	trailerDP     utils.DataProvider
	file          *os.File
}

func (rdr *FWVFileER) Config() *config.EventReaderCfg {
	return rdr.cgrCfg.ERsCfg().Readers[rdr.cfgIdx]
}

func (rdr *FWVFileER) Serve() (err error) {
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
				time.Sleep(rdr.Config().RunDelay)
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

	fileVars := &fileVars{file: file, path: absPath}
	rowNr := 0     // This counts the rows in the file, not really number of CDRs
	evsPosted := 0 // Number of CDRs successfully processed
	timeStart := time.Now()
	reqVars := utils.NavigableMap2{utils.MetaFileName: utils.NewNMData(fName)}

	for {
		var hasHeader, hasTrailer bool
		var headerFields, trailerFields []*config.FCTemplate
		if fileVars.offset == 0 { // First time, set the necessary offsets
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

			if err = rdr.setLineLen(fileVars, hasHeader, hasTrailer); err != nil {
				utils.Logger.Err(fmt.Sprintf("<%s> Row 0, error: cannot set lineLen: %s", utils.ERs, err.Error()))
				break
			}
			if hasTrailer {
				if err = rdr.processTrailer(fileVars, trailerFields); err != nil {
					utils.Logger.Err(fmt.Sprintf("<%s> Read trailer error: %s ", utils.ERs, err.Error()))
					return
				}
			}
			if hasHeader {
				if err = rdr.processHeader(fileVars, headerFields); err != nil {
					utils.Logger.Err(fmt.Sprintf("<%s> Row 0, error reading header: %s", utils.ERs, err.Error()))
					return
				}
				continue
			}
		}
		if fileVars.offset >= fileVars.trailerOffset {
			break
		}

		buf := make([]byte, fileVars.lineLength)
		if nRead, err := file.Read(buf); err != nil {
			if err == io.EOF {
				break
			}
			return err
		} else if nRead != len(buf) && int64(nRead) != fileVars.trailerLength {
			utils.Logger.Err(fmt.Sprintf("<%s> Could not read complete line, have instead: %s", utils.ERs, string(buf)))
			fileVars.offset += fileVars.lineLength // increase the offset when exit
			continue
		}
		rowNr++ // increment the rowNr after checking if it's not the end of file
		record := string(buf)
		reqVars[utils.MetaFileLineNumber] = utils.NewNMData(rowNr)
		agReq := agents.NewAgentRequest(
			config.NewFWVProvider(record), reqVars,
			nil, nil, rdr.Config().Tenant,
			rdr.cgrCfg.GeneralCfg().DefaultTenant,
			utils.FirstNonEmpty(rdr.Config().Timezone,
				rdr.cgrCfg.GeneralCfg().DefaultTimezone),
			rdr.fltrS, fileVars.headerDP, fileVars.trailerDP) // create an AgentRequest
		if pass, err := rdr.fltrS.Pass(agReq.Tenant, rdr.Config().Filters,
			agReq); err != nil || !pass {
			continue
		}
		if err := agReq.SetFields(rdr.Config().Fields); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> reading file: <%s> row <%d>, ignoring due to error: <%s>",
					utils.ERs, absPath, rowNr, err.Error()))
			fileVars.offset += fileVars.lineLength // increase the offset when exit
			continue
		}
		fileVars.offset += fileVars.lineLength // increase the offset
		rdr.rdrEvents <- &erEvent{
			cgrEvent: config.NMAsCGREvent(agReq.CGRRequest, agReq.Tenant, utils.NestingSep),
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

// Sets the line length based on first line, sets offset back to initial after reading
func (rdr *FWVFileER) setLineLen(fileVars *fileVars, hasHeader, hasTrailer bool) error {
	buff := bufio.NewReader(fileVars.file)
	// in case we have header we take the length of first line and add it as headerOffset
	i := 0
	lastLineSize := 0
	for {
		readBytes, err := buff.ReadBytes('\n')
		if err != nil {
			break
		}
		if hasHeader && i == 0 {
			fileVars.headerOffset = int64(len(readBytes))
			i++
			continue
		}
		if (hasHeader && i == 1) || (!hasHeader && i == 0) {
			fileVars.lineLength = int64(len(readBytes))
			i++
			continue
		}
		lastLineSize = len(readBytes)
	}
	if hasTrailer {
		if fi, err := fileVars.file.Stat(); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> cannot retrieve stats for file: %s, %s", utils.ERs, fileVars.path, err.Error()))
			return err
		} else {
			fileVars.trailerOffset = fi.Size() - int64(lastLineSize)
			fileVars.trailerLength = int64(lastLineSize)
		}
	}

	// reset the cursor
	if _, err := fileVars.file.Seek(0, 0); err != nil {
		return err
	}
	return nil
}

func (rdr *FWVFileER) processTrailer(fileVars *fileVars, trailerFields []*config.FCTemplate) (err error) {
	buf := make([]byte, fileVars.trailerLength)
	if nRead, err := fileVars.file.ReadAt(buf, fileVars.trailerOffset); err != nil && err != io.EOF {
		return err
	} else if nRead != len(buf) {
		return fmt.Errorf("in trailer, offset: %d, have read: %d instead of: %d", fileVars.trailerOffset, nRead, len(buf))
	}
	record := string(buf)
	fileVars.trailerDP = config.NewFWVProvider(record)
	agReq := agents.NewAgentRequest(
		utils.NavigableMap2{}, nil, nil, nil, rdr.Config().Tenant,
		rdr.cgrCfg.GeneralCfg().DefaultTenant,
		utils.FirstNonEmpty(rdr.Config().Timezone,
			rdr.cgrCfg.GeneralCfg().DefaultTimezone),
		rdr.fltrS, nil, fileVars.trailerDP) // create an AgentRequest
	if pass, err := rdr.fltrS.Pass(agReq.Tenant, rdr.Config().Filters,
		agReq); err != nil || !pass {
		return nil
	}
	if err := agReq.SetFields(trailerFields); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> reading file: <%s> trailer row, ignoring due to error: <%s>",
				utils.ERs, fileVars.path, err.Error()))
		return err
	}
	// reset the cursor after process the trailer
	_, err = fileVars.file.Seek(0, 0)
	return
}

func (rdr *FWVFileER) processHeader(fileVars *fileVars, hdrFields []*config.FCTemplate) error {
	buf := make([]byte, fileVars.headerOffset)
	if nRead, err := fileVars.file.Read(buf); err != nil {
		return err
	} else if nRead != len(buf) {
		return fmt.Errorf("in header, offset: %d, have read: %d", fileVars.headerOffset, nRead)
	}
	return rdr.createHeaderMap(string(buf), fileVars, hdrFields)
}

func (rdr *FWVFileER) createHeaderMap(record string, fileVars *fileVars, hdrFields []*config.FCTemplate) (err error) {
	fileVars.offset += fileVars.headerOffset // increase the offset
	fileVars.headerDP = config.NewFWVProvider(record)
	agReq := agents.NewAgentRequest(
		utils.NavigableMap2{}, nil, nil, nil,
		rdr.Config().Tenant,
		rdr.cgrCfg.GeneralCfg().DefaultTenant,
		utils.FirstNonEmpty(rdr.Config().Timezone,
			rdr.cgrCfg.GeneralCfg().DefaultTimezone),
		rdr.fltrS, fileVars.headerDP, nil) // create an AgentRequest
	if pass, err := rdr.fltrS.Pass(agReq.Tenant, rdr.Config().Filters,
		agReq); err != nil || !pass {
		return nil
	}
	if err := agReq.SetFields(hdrFields); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> reading file: <%s> header row, ignoring due to error: <%s>",
				utils.ERs, fileVars.path, err.Error()))
		return err
	}
	return
}
