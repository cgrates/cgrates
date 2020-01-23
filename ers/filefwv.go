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
	"io/ioutil"
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
	return &FWVFileER{
		cgrCfg:    cfg,
		cfgIdx:    cfgIdx,
		fltrS:     fltrS,
		rdrDir:    srcPath,
		rdrEvents: rdrEvents,
		rdrError:  rdrErr,
		rdrExit:   rdrExit}, nil
}

// XMLFileER implements EventReader interface for .xml files
type FWVFileER struct {
	sync.RWMutex
	cgrCfg        *config.CGRConfig
	cfgIdx        int // index of config instance within ERsCfg.Readers
	fltrS         *engine.FilterS
	headerMap     *config.NavigableMap
	rdrDir        string
	rdrEvents     chan *erEvent // channel to dispatch the events created to
	rdrError      chan error
	rdrExit       chan struct{}
	conReqs       chan struct{} // limit number of opened files
	lineLen       int64         // Length of the line in the file
	offset        int64         // Index of the next byte to process
	headerOffset  int64
	trailerOffset int64 // Index where trailer starts, to be used as boundary when reading cdrs
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
				filesInDir, _ := ioutil.ReadDir(rdr.rdrDir)
				for _, file := range filesInDir {
					if !strings.HasSuffix(file.Name(), utils.XMLSuffix) { // hardcoded file extension for xml event reader
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

	rowNr := 0 // This counts the rows in the file, not really number of CDRs
	evsPosted := 0
	timeStart := time.Now()
	reqVars := make(map[string]interface{})

	for {
		if rdr.offset == 0 { // First time, set the necessary offsets
			if err := rdr.setLineLen(file); err != nil {
				utils.Logger.Err(fmt.Sprintf("<%s> Row 0, error: cannot set lineLen: %s", utils.ERs, err.Error()))
				break
			}
			if len(rdr.Config().TrailerFields) != 0 {
				if fi, err := file.Stat(); err != nil {
					utils.Logger.Err(fmt.Sprintf("<%s> Row 0, error: cannot get file stats: %s", utils.ERs, err.Error()))
					return err
				} else {
					rdr.trailerOffset = fi.Size() - rdr.lineLen
				}
			}
			if len(rdr.Config().HeaderFields) != 0 {
				if err = rdr.processHeader(file, rowNr, evsPosted, absPath); err != nil {
					utils.Logger.Err(fmt.Sprintf("<%s> Row 0, error reading header: %s", utils.ERs, err.Error()))
					return
				}
				continue
			}
		}

		buf := make([]byte, rdr.lineLen)
		nRead, err := file.Read(buf)
		if err != nil {
			rdr.offset += rdr.lineLen // increase the offset when exit
			return err
		} else if nRead != len(buf) {
			utils.Logger.Err(fmt.Sprintf("<%s> Could not read complete line, have instead: %s", utils.ERs, string(buf)))
			rdr.offset += rdr.lineLen // increase the offset when exit
			break
		}
		rowNr++ // increment the rowNr after checking if it's not the end of file
		record := string(buf)
		agReq := agents.NewAgentRequest(
			config.NewFWVProvider(record), reqVars,
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
			rdr.offset += rdr.lineLen // increase the offset when exit
			continue
		}
		//backwards compatible with CDRC
		if rdr.headerMap != nil {
			navMp.Merge(rdr.headerMap)
		}
		rdr.offset += rdr.lineLen // increase the offset
		rdr.rdrEvents <- &erEvent{cgrEvent: navMp.AsCGREvent(
			agReq.Tenant, utils.NestingSep),
			rdrCfg: rdr.Config()}
		evsPosted++
		if rdr.trailerOffset != 0 && rdr.offset >= rdr.trailerOffset {
			if err := rdr.processTrailer(file, rowNr, evsPosted, absPath); err != nil && err != io.EOF {
				utils.Logger.Err(fmt.Sprintf("<%s> Read trailer error: %s ", utils.ERs, err.Error()))
			}
			break
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

// Sets the line length based on first line, sets offset back to initial after reading
func (rdr *FWVFileER) setLineLen(file *os.File) error {
	buff := bufio.NewReader(file)
	// in case we have header we take the length of first line and add it as headerOffset
	if len(rdr.Config().HeaderFields) != 0 {
		readBytes, err := buff.ReadBytes('\n')
		if err != nil {
			return err
		}
		rdr.headerOffset = int64(len(readBytes))
	}
	readBytes, err := buff.ReadBytes('\n')
	if err != nil {
		return err
	}
	rdr.lineLen = int64(len(readBytes))

	if _, err := file.Seek(0, 0); err != nil {
		return err
	}
	return nil
}

func (rdr *FWVFileER) processTrailer(file *os.File, rowNr, evsPosted int, absPath string) (err error) {
	buf := make([]byte, rdr.trailerOffset)
	if nRead, err := file.ReadAt(buf, rdr.trailerOffset); err != nil {
		return err
	} else if nRead != len(buf) {
		return fmt.Errorf("In trailer, line len: %d, have read: %d", rdr.trailerOffset, nRead)
	}
	record := string(buf)
	reqVars := make(map[string]interface{})
	agReq := agents.NewAgentRequest(
		config.NewFWVProvider(record), reqVars,
		nil, nil, rdr.Config().Tenant,
		rdr.cgrCfg.GeneralCfg().DefaultTenant,
		utils.FirstNonEmpty(rdr.Config().Timezone,
			rdr.cgrCfg.GeneralCfg().DefaultTimezone),
		rdr.fltrS) // create an AgentRequest
	if pass, err := rdr.fltrS.Pass(agReq.Tenant, rdr.Config().Filters,
		agReq); err != nil || !pass {
		return nil
	}
	navMp, err := agReq.AsNavigableMap(rdr.Config().TrailerFields)
	if err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> reading file: <%s> row <%d>, ignoring due to error: <%s>",
				utils.ERs, absPath, rowNr, err.Error()))
		return err
	}
	rdr.rdrEvents <- &erEvent{cgrEvent: navMp.AsCGREvent(
		agReq.Tenant, utils.NestingSep),
		rdrCfg: rdr.Config()}
	evsPosted++
	return
}

func (rdr *FWVFileER) processHeader(file *os.File, rowNr, evsPosted int, absPath string) error {
	buf := make([]byte, rdr.headerOffset)
	if nRead, err := file.Read(buf); err != nil {
		return err
	} else if nRead != len(buf) {
		return fmt.Errorf("In header, line len: %d, have read: %d", rdr.headerOffset, nRead)
	}
	return rdr.createHeaderMap(string(buf), rowNr, evsPosted, absPath)
}

func (rdr *FWVFileER) createHeaderMap(record string, rowNr, evsPosted int, absPath string) (err error) {
	reqVars := make(map[string]interface{})
	agReq := agents.NewAgentRequest(
		config.NewFWVProvider(record), reqVars,
		nil, nil, rdr.Config().Tenant,
		rdr.cgrCfg.GeneralCfg().DefaultTenant,
		utils.FirstNonEmpty(rdr.Config().Timezone,
			rdr.cgrCfg.GeneralCfg().DefaultTimezone),
		rdr.fltrS) // create an AgentRequest
	if pass, err := rdr.fltrS.Pass(agReq.Tenant, rdr.Config().Filters,
		agReq); err != nil || !pass {
		return nil
	}
	navMp, err := agReq.AsNavigableMap(rdr.Config().HeaderFields)
	if err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> reading file: <%s> row <%d>, ignoring due to error: <%s>",
				utils.ERs, absPath, rowNr, err.Error()))
		rdr.offset += rdr.lineLen // increase the offset when exit
		return err
	}
	rdr.headerMap = navMp
	rdr.offset += rdr.headerOffset // increase the offset
	rdr.rdrEvents <- &erEvent{cgrEvent: navMp.AsCGREvent(
		agReq.Tenant, utils.NestingSep),
		rdrCfg: rdr.Config()}
	evsPosted++
	return
}
