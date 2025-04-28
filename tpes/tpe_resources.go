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

package tpes

import (
	"encoding/csv"
	"fmt"
	"io"

	"github.com/cgrates/birpc/context"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

type TPResources struct {
	dm *engine.DataManager
}

// newTPResources is the constructor for TPResources
func newTPResources(dm *engine.DataManager) *TPResources {
	return &TPResources{
		dm: dm,
	}
}

// exportItems for TPResources will implement the method for tpExporter interface
func (tpRes TPResources) exportItems(ctx *context.Context, wrtr io.Writer, tnt string, itmIDs []string) (err error) {
	csvWriter := csv.NewWriter(wrtr)
	csvWriter.Comma = utils.CSVSep
	// before writing the profiles, we must write the headers
	if err = csvWriter.Write([]string{"#Tenant", "ID", "FIlterIDs", "Weights", "TTL", "Limit", "AlocationMessage", "Blocker", "Stored", "ThresholdIDs"}); err != nil {
		return
	}
	for _, resID := range itmIDs {
		var resPrf *utils.ResourceProfile
		resPrf, err = tpRes.dm.GetResourceProfile(ctx, tnt, resID, true, true, utils.NonTransactional)
		if err != nil {
			if err.Error() == utils.ErrNotFound.Error() {
				return fmt.Errorf("<%s> cannot find ResourceProfile with id: <%v>", err, resID)
			}
			return err
		}
		resMdl := engine.APItoModelResource(engine.ResourceProfileToAPI(resPrf))
		if len(resMdl) == 0 {
			return
		}
		// for every profile, convert it into model to be compatible in csv format
		for _, tpItem := range resMdl {
			// transform every record into a []string
			var record []string
			record, err = engine.CsvDump(tpItem)
			if err != nil {
				return err
			}
			// record is a line of a csv file
			if err = csvWriter.Write(record); err != nil {
				return err
			}
		}
	}
	csvWriter.Flush()
	return
}
