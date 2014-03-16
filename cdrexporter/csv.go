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

package cdrexporter

import (
	"encoding/csv"
	"github.com/cgrates/cgrates/utils"
	"io"
)

type CsvCdrWriter struct {
	writer         *csv.Writer
	roundDecimals  int               // Round floats like Cost using this number of decimals
	exportedFields []*utils.RSRField // The fields exported, order important
}

func NewCsvCdrWriter(writer io.Writer, roundDecimals int, exportedFields []*utils.RSRField) *CsvCdrWriter {
	return &CsvCdrWriter{csv.NewWriter(writer), roundDecimals, exportedFields}
}

func (csvwr *CsvCdrWriter) Write(cdr *utils.StoredCdr) error {
	row := make([]string, len(csvwr.exportedFields))
	for idx, fld := range csvwr.exportedFields { // Add primary fields
		var fldVal string
		if fld.Id == utils.COST {
			fldVal = cdr.FormatCost(csvwr.roundDecimals)
		} else {
			fldVal = cdr.ExportFieldValue(fld.Id)
		}
		row[idx] = fld.ParseValue(fldVal)
	}
	return csvwr.writer.Write(row)
}

func (csvwr *CsvCdrWriter) Close() {
	csvwr.writer.Flush()
}
