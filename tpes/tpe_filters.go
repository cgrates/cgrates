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

package tpes

import (
	"encoding/csv"
	"fmt"
	"io"

	"github.com/cgrates/birpc/context"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

type TPFilters struct {
	dm *engine.DataManager
}

// newTPFilters is the constructor for TPFilters
func newTPFilters(dm *engine.DataManager) *TPFilters {
	return &TPFilters{
		dm: dm,
	}
}

// exportItems for TPFilters will implement the method for tpExporter interface
func (tpFltr TPFilters) exportItems(ctx *context.Context, wrtr io.Writer, tnt string, itmIDs []string) (err error) {
	csvWriter := csv.NewWriter(wrtr)
	csvWriter.Comma = utils.CSVSep
	// before writing the profiles, we must write the headers
	if err = csvWriter.Write([]string{"#Tenant", "ID", "Type", "Path", "Values"}); err != nil {
		return
	}
	for _, fltrID := range itmIDs {
		var fltr *engine.Filter
		fltr, err = tpFltr.dm.GetFilter(ctx, tnt, fltrID, true, true, utils.NonTransactional)
		if err != nil {
			if err.Error() == utils.ErrNotFound.Error() {
				return fmt.Errorf("<%s> cannot find Filters with id: <%v>", err, fltrID)
			}
			return err
		}
		fltrMdls := engine.APItoModelTPFilter(engine.FilterToTPFilter(fltr))
		if len(fltrMdls) == 0 {
			return
		}
		// for every profile, convert it into model to be compatible in csv format
		for _, tpItem := range fltrMdls {
			// transform every record into a []string
			var record []string
			record, err = engine.CsvDump(tpItem)
			if err != nil {
				return err
			}
			// record is a line of a csv file
			if err := csvWriter.Write(record); err != nil {
				return err
			}
		}
	}
	csvWriter.Flush()
	return
}
