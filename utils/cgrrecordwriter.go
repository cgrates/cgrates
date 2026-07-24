// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package utils

import (
	"io"
)

// CgrRecordWriter is a writer for one line, compatible with csv.Writer interface on Write
// Used in TP exporter
type CgrRecordWriter interface {
	Write([]string) error
	Flush()
}

// NewCgrIORecordWriter return CgrRecordWriter that will replace csv.Writer
func NewCgrIORecordWriter(w io.Writer) *CgrIORecordWriter {
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
func (*CgrIORecordWriter) Flush() {
	return
}
