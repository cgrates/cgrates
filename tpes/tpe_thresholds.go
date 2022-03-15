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
	if len(itmIDs) == 0 {
		prfx := utils.ThresholdProfilePrefix + tnt + utils.ConcatenatedKeySep
		// dbKeys will contain the full name of the key, but we will need just the IDs e.g. "acn_cgrates.org:THD_1" -- just THD_1
		var dbKeys []string
		if dbKeys, err = tpThd.dm.DataDB().GetKeysForPrefix(ctx, prfx); err != nil {
			return err
		}
		profileIDs := make([]string, 0, len(dbKeys))
		for _, key := range dbKeys {
			profileIDs = append(profileIDs, key[len(prfx):])
		}
		// if there are not any profiles in db, we do not write in our zip
		if len(profileIDs) == 0 {
			return
		}
		// the map e.g. : *filters: {"THD_1", "THD_1"}
		itmIDs = profileIDs
	}
	csvWriter := csv.NewWriter(wrtr)
	csvWriter.Comma = utils.CSVSep
	// before writing the profiles, we must write the headers
	if err = csvWriter.Write([]string{"#Tenant", "ID", "FilterIDs", "Weights", "Schedule", "TargetType", "TargetIDs", "ActionID", "ActionFilterIDs", "ActionBlocker", "ActionTTL", "ActionType", "ActionOpts", "ActionPath", "ActionValue"}); err != nil {
		return
	}
	for _, thdID := range itmIDs {
		var thdPrf *engine.ThresholdProfile
		thdPrf, err = tpThd.dm.GetThresholdProfile(ctx, tnt, thdID, true, true, utils.NonTransactional)
		if err != nil {
			if err.Error() == utils.ErrNotFound.Error() {
				return fmt.Errorf("<%s> cannot find Actions id: <%v>", err, thdID)
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
