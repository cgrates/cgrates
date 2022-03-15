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

type TPActions struct {
	dm *engine.DataManager
}

// newTPActions is the constructor for TPActions
func newTPActions(dm *engine.DataManager) *TPActions {
	return &TPActions{
		dm: dm,
	}
}

// exportItems for TPActions will implement the method for tpExporter interface
func (tpActs TPActions) exportItems(ctx *context.Context, wrtr io.Writer, tnt string, itmIDs []string) (err error) {
	if len(itmIDs) == 0 {
		prfx := utils.ActionProfilePrefix + tnt + utils.ConcatenatedKeySep
		// dbKeys will contain the full name of the key, but we will need just the IDs e.g. "acn_cgrates.org:ACCTS_1" -- just ACCTS_1
		var dbKeys []string
		if dbKeys, err = tpActs.dm.DataDB().GetKeysForPrefix(ctx, prfx); err != nil {
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
		// the map e.g. : *filters: {"ACCTS_1", "ACCTS_1"}
		itmIDs = profileIDs
	}
	csvWriter := csv.NewWriter(wrtr)
	csvWriter.Comma = utils.CSVSep
	// before writing the profiles, we must write the headers
	if err = csvWriter.Write([]string{"#Tenant", "ID", "FilterIDs", "Weights", "Schedule", "TargetType", "TargetIDs", "ActionID", "ActionFilterIDs", "ActionBlocker", "ActionTTL", "ActionType", "ActionOpts", "ActionPath", "ActionValue"}); err != nil {
		return
	}
	for _, actsID := range itmIDs {
		var actsPrf *engine.ActionProfile
		actsPrf, err = tpActs.dm.GetActionProfile(ctx, tnt, actsID, true, true, utils.NonTransactional)
		if err != nil {
			if err.Error() == utils.ErrNotFound.Error() {
				return fmt.Errorf("<%s> cannot find Actions id: <%v>", err, actsID)
			}
			return err
		}
		actsMdls := engine.APItoModelTPActionProfile(engine.ActionProfileToAPI(actsPrf))
		if len(actsMdls) == 0 {
			return
		}
		// for every profile, convert it into model to be compatible in csv format
		for _, tpItem := range actsMdls {
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
