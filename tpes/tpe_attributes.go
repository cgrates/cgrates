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

type TPAttributes struct {
	dm *engine.DataManager
}

// newTPAttributes is the constructor for TPAttributes
func newTPAttributes(dm *engine.DataManager) *TPAttributes {
	return &TPAttributes{
		dm: dm,
	}
}

// exportItems for TPAttributes will implement the method for tpExporter interface
func (tpAttr TPAttributes) exportItems(ctx *context.Context, wrtr io.Writer, tnt string, itmIDs []string) (err error) {
	csvWriter := csv.NewWriter(wrtr)
	csvWriter.Comma = utils.CSVSep
	for _, attrID := range itmIDs {
		var attrPrf *engine.AttributeProfile
		attrPrf, err = tpAttr.dm.GetAttributeProfile(ctx, tnt, attrID, true, true, utils.NonTransactional)
		if err != nil {
			if err.Error() == utils.ErrNotFound.Error() {
				utils.Logger.Warning(fmt.Sprintf("<%s> cannot find AttributeProfile with id: <%v>", utils.TPeS, attrID))
				continue
			}
			return err
		}
		attrMdl := engine.APItoModelTPAttribute(engine.AttributeProfileToAPI(attrPrf))
		if len(attrMdl) == 0 {
			return
		}
		// for every profile, convert it into model to be writable in csv format
		for _, tpItem := range attrMdl {
			// transform every record into a []string
			record, err := engine.CsvDump(tpItem)
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
