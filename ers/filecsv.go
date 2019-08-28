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
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func NewCSVFileER(cfg *config.EventReaderCfg) (er EventReader, err error) {
	return new(CSVFileER), nil
}

// CSVFileER implements EventReader interface for .csv files
type CSVFileER struct {
}

func (csv *CSVFileER) ID() (id string) {
	return
}

func (csv *CSVFileER) Config() (rdrCfg *config.EventReaderCfg) {
	return
}

func (csv *CSVFileER) Init(args interface{}) (err error) {
	return
}

func (csv *CSVFileER) Read() (ev *utils.CGREvent, err error) {
	return
}

func (csv *CSVFileER) Processed() (nrItms int64) {
	return
}

func (csv *CSVFileER) Close() (err error) {
	return
}
