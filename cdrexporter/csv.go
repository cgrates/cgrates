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
	"io"
)

type CsvCdrWriter struct {
	writer *csv.Writer
}

func NewCsvCdrWriter(writer io.Writer) *CsvCdrWriter {
	return &CsvCdrWriter{csv.NewWriter(writer)}
}

func (dcw *CsvCdrWriter) Write(cdr map[string]string) error {
	var row []string
	for _, v := range cdr {
		row = append(row, v)
	}
	return dcw.writer.Write(row)
}

func (dcw *CsvCdrWriter) Close() {
	dcw.writer.Flush()
}
