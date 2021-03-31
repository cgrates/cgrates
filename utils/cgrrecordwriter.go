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

package utils

import (
	"io"
)

// NopFlushWriter is a writer for one line, compatible with csv.Writer interface on Write
// Used in TP exporter
type NopFlushWriter interface {
	Write([]string) error
	Flush()
}

// NewNopFlushWriter return CgrRecordWriter that will replace csv.Writer
func NewNopFlushWriter(w io.Writer) *CgrIORecordWriter {
	return &CgrIORecordWriter{w: w}
}

// CgrIORecordWriter implements CgrRecordWriter
type CgrIORecordWriter struct {
	w io.Writer
}

// Write into the Writer the record
func (rw *CgrIORecordWriter) Write(record []string) error {
	for _, fld := range append(record, "\n") { // Postpend the new line char and write record in the writer
		if _, err := io.WriteString(rw.w, fld); err != nil {
			return err
		}
	}
	return nil
}

// Flush only to implement CgrRecordWriter
// ToDo: make sure we properly handle this method
func (*CgrIORecordWriter) Flush() {}
