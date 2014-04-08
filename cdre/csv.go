/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2013 ITsysCOM

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

package cdre

import (
	"encoding/csv"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"io"
)

type CsvCdrWriter struct {
	writer                         *csv.Writer
	costShiftDigits, roundDecimals int // Round floats like Cost using this number of decimals
	maskDestId                     string
	maskLen                        int
	exportedFields                 []*utils.RSRField // The fields exported, order important
	firstExpOrderId, lastExpOrderId                 int64
}

func NewCsvCdrWriter(writer io.Writer, costShiftDigits, roundDecimals int, maskDestId string, maskLen int, exportedFields []*utils.RSRField) *CsvCdrWriter {
	return &CsvCdrWriter{writer:csv.NewWriter(writer), costShiftDigits: costShiftDigits, roundDecimals:roundDecimals, maskDestId:maskDestId, maskLen:maskLen, exportedFields:exportedFields}
}

// Return the first exported Cdr OrderId
func (csvwr *CsvCdrWriter) FirstOrderId() int64 {
	return csvwr.firstExpOrderId
}

func (csvwr *CsvCdrWriter) LastOrderId() int64 {
	return csvwr.lastExpOrderId
}

func (csvwr *CsvCdrWriter) WriteCdr(cdr *utils.StoredCdr) error {
	row := make([]string, len(csvwr.exportedFields))
	for idx, fld := range csvwr.exportedFields {
		var fldVal string
		if fld.Id == utils.COST {
			fldVal = cdr.FormatCost(csvwr.costShiftDigits, csvwr.roundDecimals)
		} else if fld.Id == utils.DESTINATION {
			fldVal = cdr.ExportFieldValue(utils.DESTINATION)
			if len(csvwr.maskDestId) != 0 && csvwr.maskLen > 0 && engine.CachedDestHasPrefix(csvwr.maskDestId, fldVal) {
				fldVal = MaskDestination(fldVal, csvwr.maskLen)
			}
		} else {
			fldVal = cdr.ExportFieldValue(fld.Id)
		}
		row[idx] = fld.ParseValue(fldVal)
	}
	if csvwr.firstExpOrderId > cdr.OrderId || csvwr.firstExpOrderId == 0 {
		csvwr.firstExpOrderId = cdr.OrderId
	}
	if csvwr.lastExpOrderId < cdr.OrderId {
		csvwr.lastExpOrderId = cdr.OrderId
	}
	return csvwr.writer.Write(row)

	
}

func (csvwr *CsvCdrWriter) Close() {
	csvwr.writer.Flush()
}
