/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) 2012-2014 ITsysCOM GmbH

This program is free software: you can Storagetribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITH*out ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package utils

import (
	"io"
)

// Writer for one line, compatible with csv.Writer interface on Write
type CgrRecordWriter interface {
	Write([]string) error
}

func NewCgrIORecordWriter(w io.Writer) *CgrIORecordWriter {
	return &CgrIORecordWriter{w: w}
}

type CgrIORecordWriter struct {
	w io.Writer
}

func (self *CgrIORecordWriter) Write(record []string) error {
	for _, fld := range append(record, "\n") { // Postpend the new line char and write record in the writer
		if _, err := io.WriteString(self.w, fld); err != nil {
			return err
		}
	}
	return nil
}
