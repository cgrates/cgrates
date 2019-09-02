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
	"fmt"
	"io/ioutil"
	"strings"
	"sync"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func NewCSVFileER(cfg *config.EventReaderCfg,
	rdrExit chan struct{}, appExit chan bool) (er EventReader, err error) {
	srcPath := cfg.SourcePath
	if strings.HasSuffix(srcPath, utils.Slash) {
		srcPath = srcPath[:len(srcPath)-1]
	}
	return &CSVFileER{erCfg: cfg, rdrDir: srcPath,
		rdrExit: rdrExit, appExit: appExit}, nil
}

// CSVFileER implements EventReader interface for .csv files
type CSVFileER struct {
	sync.RWMutex
	erCfg   *config.EventReaderCfg
	rdrDir  string
	rdrExit chan struct{}
	appExit chan bool
}

func (rdr *CSVFileER) Config() *config.EventReaderCfg {
	return rdr.erCfg
}

func (rdr *CSVFileER) Init() (err error) {
	switch rdr.erCfg.RunDelay {
	case time.Duration(0): // 0 disables the automatic read, maybe per API
		return
	case time.Duration(-1):
		return watchDir(rdr.rdrDir, rdr.processFile,
			utils.ERs, rdr.rdrExit)
	default:
		// Not automated, process and sleep approach
		for {
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
				go func() {
					if err := rdr.processFile(rdr.rdrDir, file.Name()); err != nil {
						utils.Logger.Warning(
							fmt.Sprintf("<%s> processing file %s, error: %s",
								utils.ERs, file, err.Error()))
					}
				}()
			}
			time.Sleep(rdr.erCfg.RunDelay)
		}
	}
}

func (rdr *CSVFileER) Read() (ev *utils.CGREvent, err error) {
	return
}

// processFile is called for each file in a directory and dispatches erEvents from it
func (rdr *CSVFileER) processFile(itmPath, itmID string) (err error) {
	return
}
