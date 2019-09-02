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
	"strings"
	"sync"

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

func (csv *CSVFileER) Config() *config.EventReaderCfg {
	return csv.erCfg
}

func (csv *CSVFileER) Init() (err error) {
	if err := watchDir(csv.rdrDir, csv.processDir,
		utils.ERs, csv.rdrExit); err != nil {
		utils.Logger.Crit(
			fmt.Sprintf("<%s> watching directory <%s> got error: <%s>",
				utils.ERs, csv.rdrDir, err.Error()))
		csv.appExit <- true
	}
	return
}

func (csv *CSVFileER) Read() (ev *utils.CGREvent, err error) {
	return
}

// processDir is called for each file in a directory and dispatches erEvents from it
func (csv *CSVFileER) processDir(itmPath, itmID string) (err error) {

	return
}
