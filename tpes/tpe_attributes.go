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

type TPAttributes struct {
	dm *engine.DataManager
}

// newTPAttributes is the constructor for TPAttributes
func newTPAttributes(dm *engine.DataManager) *TPAttributes {
	return &TPAttributes{
		dm: dm,
	}
}

// exportItems for TPAttributes will implement the imethod for tpExporter interface
func (tpAttr TPAttributes) exportItems(ctx *context.Context, tnt string, itmIDs []string) (expContent []byte, err error) {
	//attrBts := make(map[string][]byte)
	for _, attrID := range itmIDs {
		attrPrf, err := tpAttr.dm.GetAttributeProfile(ctx, tnt, attrID, true, true, utils.NonTransactional)
		if err != nil {
			if err.Error() == utils.ErrNotFound.Error() {
				utils.Logger.Warning(fmt.Sprintf("<%s> cannot find AttributeProfile with id: <%v>", utils.TPeS, attrID))
				continue
			}
			return nil, err
		}

		attrMdl := engine.APItoModelTPAttribute(engine.AttributeProfileToAPI(attrPrf))
		if err := writeOut(utils.AttributesCsv, attrMdl); err != nil {
			return nil, err
		}

	}
	return
}

func writeOut(fileName string, tpData engine.AttributeMdls) error {
	buff := new(bytes.Buffer)

	csvWriter := csv.NewWriter(buff)
	for _, tpItem := range tpData {
		record, err := engine.CsvDump(tpItem)
		if err != nil {
			return err
		}
		if err := csvWriter.Write(record); err != nil {
			return err
		}
	}

	return nil
}
