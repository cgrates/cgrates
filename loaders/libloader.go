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

package loaders

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

type LoaderData map[string]interface{}

// UpdateFromCSV will update LoaderData with data received from fileName,
// contained in record and processed with cfgTpl
func (ld LoaderData) UpdateFromCSV(fileName string, record []string,
	cfgTpl []*config.CfgCdrField) (err error) {
	for _, cfgFld := range cfgTpl {
		var valStr string
		for _, rsrFld := range cfgFld.Value {
			if rsrFld.IsStatic() {
				var val string
				if val, err = rsrFld.Parse(""); err != nil {
					return err
				}
				valStr += val
				continue
			}
			idxStr := rsrFld.Id // default to Id in the rsrField
			spltSrc := strings.Split(rsrFld.Id, utils.InInFieldSep)
			if len(spltSrc) == 2 { // having field name inside definition, compare here with our source
				if spltSrc[0] != fileName {
					continue
				}
				idxStr = spltSrc[1] // will have index at second position in the rule definition
			}
			var cfgFieldIdx int
			if cfgFieldIdx, err = strconv.Atoi(idxStr); err != nil {
				return
			} else if len(record) <= cfgFieldIdx {
				return fmt.Errorf("Ignoring record: %v - cannot extract field %s", record, cfgFld.Tag)
			}
			valStr += rsrFld.ParseValue(record[cfgFieldIdx])
		}
		switch cfgFld.Type {
		case utils.META_COMPOSED:
			if valOrig, canCast := ld[cfgFld.FieldId].(string); canCast {
				valOrig += valStr
				ld[cfgFld.FieldId] = valOrig
			}
		}
	}
	return
}
