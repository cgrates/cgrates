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
	"bytes"
	"encoding/csv"
	"fmt"

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

// exportItems for TPResources will implement the imethod for tpExporter interface
func (tpAttr TPResources) exportItems(ctx *context.Context, tnt string, itmIDs []string) (expContent []byte, err error) {
	expContent = make([]byte, len(itmIDs))

	for _, attrID := range itmIDs {
		resPrf, err := tpAttr.dm.GetResourceProfile(ctx, tnt, attrID, true, true, utils.NonTransactional)
		if err != nil {
			if err.Error() == utils.ErrNotFound.Error() {
				utils.Logger.Warning(fmt.Sprintf("<%s> cannot find ResourceProfile with id: <%v>", utils.TPeS, attrID))
				continue
			}
			return nil, err
		}
		resMdl := engine.APItoModelResource(engine.ResourceProfileToAPI(resPrf))
		// for every profile, convert it into model to be writable in csv format
		buff := new(bytes.Buffer) // the info will be stored into a buffer
		csvWriter := csv.NewWriter(buff)
		csvWriter.Comma = utils.CSVSep
		for _, tpItem := range resMdl {
			record, err := engine.CsvDump(tpItem)
			if err != nil {
				return nil, err
			}
			// record is a line of a csv file
			if err := csvWriter.Write(record); err != nil {
				return nil, err
			}
		}
		csvWriter.Flush()
		// append our bytes stored in buffer for every profile
		expContent = append(expContent, buff.Bytes()...)
	}
	return
}
