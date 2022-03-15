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

type TPStats struct {
	dm *engine.DataManager
}

// newTPStats is the constructor for TPStats
func newTPStats(dm *engine.DataManager) *TPStats {
	return &TPStats{
		dm: dm,
	}
}

// exportItems for TPStats will implement the method for tpExporter interface
func (tpSts TPStats) exportItems(ctx *context.Context, wrtr io.Writer, tnt string, itmIDs []string) (err error) {
	if len(itmIDs) == 0 {
		prfx := utils.StatQueueProfilePrefix + tnt + utils.ConcatenatedKeySep
		// dbKeys will contain the full name of the key, but we will need just the IDs e.g. "acn_cgrates.org:STAT_1" -- just STAT_1
		var dbKeys []string
		if dbKeys, err = tpSts.dm.DataDB().GetKeysForPrefix(ctx, prfx); err != nil {
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
		// the map e.g. : *filters: {"STAT_1", "STAT_1"}
		itmIDs = profileIDs
	}
	csvWriter := csv.NewWriter(wrtr)
	csvWriter.Comma = utils.CSVSep
	// before writing the profiles, we must write the headers
	if err = csvWriter.Write([]string{"#Tenant", "ID", "FilterIDs", "Weights", "QueueLength", "TTL", "MinItems", "Metrics", "MetricFilterIDs", "Stored", "Blocker", "ThresholdIDs"}); err != nil {
		return
	}
	for _, statsID := range itmIDs {
		var statPrf *engine.StatQueueProfile
		statPrf, err = tpSts.dm.GetStatQueueProfile(ctx, tnt, statsID, true, true, utils.NonTransactional)
		if err != nil {
			if err.Error() == utils.ErrNotFound.Error() {
				return fmt.Errorf("<%s> cannot find StatQueueProfile with id: <%v>", err, statsID)
			}
			return err
		}
		statsMdls := engine.APItoModelStats(engine.StatQueueProfileToAPI(statPrf))
		if len(statsMdls) == 0 {
			return
		}
		// for every profile, convert it into model to be compatible in csv format
		for _, tpItem := range statsMdls {
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
