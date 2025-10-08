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

type TPAccounts struct {
	dm *engine.DataManager
}

// newTPAccounts is the constructor for TPAccounts
func newTPAccounts(dm *engine.DataManager) *TPAccounts {
	return &TPAccounts{
		dm: dm,
	}
}

// exportItems for TPAccounts will implement the method for tpExporter interface
func (tpAcc TPAccounts) exportItems(ctx *context.Context, wrtr io.Writer, tnt string, itmIDs []string) (err error) {
	csvWriter := csv.NewWriter(wrtr)
	csvWriter.Comma = utils.CSVSep
	var accMdls engine.AccountMdls
	// before writing the profiles, we must write the headers
	if err = csvWriter.Write(accMdls.CSVHeader()); err != nil {
		return
	}
	for _, accID := range itmIDs {
		var acc *utils.Account
		acc, err = tpAcc.dm.GetAccount(ctx, tnt, accID)
		if err != nil {
			if err.Error() == utils.ErrNotFound.Error() {
				return fmt.Errorf("<%s> cannot find Account with id: <%v>", err, accID)
			}
			return err
		}
		accMdls = engine.APItoModelTPAccount(engine.AccountToAPI(acc))
		if len(accMdls) == 0 {
			return
		}
		// for every profile, convert it into model to be compatible in csv format
		for _, tpItem := range accMdls {
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
