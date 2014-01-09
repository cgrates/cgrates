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
	"sort"
	"strconv"
)

type CsvCdrWriter struct {
	writer        *csv.Writer
	roundDecimals int      // Round floats like Cost using this number of decimals
	extraFields   []string // Extra fields to append after primary ones, order important
}

func NewCsvCdrWriter(writer io.Writer, roundDecimals int, extraFields []string) *CsvCdrWriter {
	return &CsvCdrWriter{csv.NewWriter(writer), roundDecimals, extraFields}
}

func (dcw *CsvCdrWriter) Write(cdr *utils.RatedCDR) error {
	primaryFields := []string{cdr.CgrId, cdr.MediationRunId, cdr.AccId, cdr.CdrHost, cdr.ReqType, cdr.Direction, cdr.Tenant, cdr.TOR, cdr.Account, cdr.Subject,
		cdr.Destination, cdr.AnswerTime.String(), strconv.Itoa(int(cdr.Duration)), strconv.FormatFloat(cdr.Cost, 'f', dcw.roundDecimals, 64)}
	if len(dcw.extraFields) == 0 {
		dcw.extraFields = utils.MapKeys(cdr.ExtraFields)
		sort.Strings(dcw.extraFields) // Controlled order in case of dynamic extra fields
	}
	lenPrimary := len(primaryFields)
	row := make([]string, lenPrimary+len(dcw.extraFields))
	for idx, fld := range primaryFields { // Add primary fields
		row[idx] = fld
	}
	for idx, fldKey := range dcw.extraFields { // Add extra fields
		row[lenPrimary+idx] = cdr.ExtraFields[fldKey]
	}
	return dcw.writer.Write(row)
}

func (dcw *CsvCdrWriter) Close() {
	dcw.writer.Flush()
}
