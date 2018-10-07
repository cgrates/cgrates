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

package engine

import (
	"encoding/csv"

	"github.com/cgrates/cgrates/config"
)

// TPcsvReader reads data from csv file based on template
type TPcsvReader struct {
	fldsTpl    []*config.CfgCdrField
	srcType    string
	srcPath    string
	csvReaders map[string]*csv.Reader // map[fileName]*csv.Reader for each name in fldsTpl
}

// Read implements TPReader interface
func (csv *TPcsvReader) Read() (itm interface{}, err error) {
	return
}
