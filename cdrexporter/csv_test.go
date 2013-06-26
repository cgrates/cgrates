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
	"bytes"
	"strings"
	"testing"
)

func TestCsv(t *testing.T) {
	writer := &bytes.Buffer{}
	csvCdrWriter := NewCsvCdrWriter(writer)
	csvCdrWriter.Write(map[string]string{"First": "test", "Second": "the", "Third": "cdr"})
	csvCdrWriter.Close()
	expected := "test,the,cdr"
	result := strings.TrimSpace(writer.String())
	if result != expected {
		t.Errorf("Expected %s was %s.", expected, result)
	}
}
