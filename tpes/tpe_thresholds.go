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

type TPThresholds struct {
	dm *engine.DataManager
}

// newTPThresholds is the constructor for TPThresholds
func newTPThresholds(dm *engine.DataManager) *TPThresholds {
	return &TPThresholds{
		dm: dm,
	}
}

// exportItems for TPThresholds will implement the method for tpExporter interface
func (tpThd TPThresholds) exportItems(ctx *context.Context, wrtr io.Writer, tnt string, itmIDs []string) (err error) {
	csvWriter := csv.NewWriter(wrtr)
	csvWriter.Comma = utils.CSVSep
	// before writing the profiles, we must write the headers
	if err = csvWriter.Write([]string{"#Tenant", "ID", "FilterIDs", "Weights", "MaxHits", "MinHits", "MinSleep", "Blocker", "ActionProfileIDs", "Async", "EeIDs"}); err != nil {
		return
	}
	for _, thdID := range itmIDs {
		var thdPrf *engine.ThresholdProfile
		thdPrf, err = tpThd.dm.GetThresholdProfile(ctx, tnt, thdID, true, true, utils.NonTransactional)
		if err != nil {
			if err.Error() == utils.ErrNotFound.Error() {
				return fmt.Errorf("<%s> cannot find ThresholdProfile with id: <%v>", err, thdID)
			}
			return err
		}
		thdMdls := engine.APItoModelTPThreshold(engine.ThresholdProfileToAPI(thdPrf))
		if len(thdMdls) == 0 {
			return
		}
		// for every profile, convert it into model to be compatible in csv format
		for _, tpItem := range thdMdls {
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
