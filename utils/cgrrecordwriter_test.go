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
	"bytes"
	"io"
	"reflect"
	"testing"
)

func TestNewCgrIORecordWriter(t *testing.T) {
	var args io.Writer
	rcv := NewCgrIORecordWriter(args)
	eOut := &CgrIORecordWriter{w: args}
	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting %+v, received %+v", eOut, rcv)
	}
}

func TestWrite(t *testing.T) {
	//empty check
	args := new(bytes.Buffer)
	rw := NewCgrIORecordWriter(args)
	record := []string{"test1", "test2"}
	rcv := rw.Write(record)
	if rcv != nil {
		t.Errorf("Expecting nil, received %+v", rcv)
	}
	eOut := "test1test2\n"
	if !reflect.DeepEqual(eOut, args.String()) {
		t.Errorf("Expected %q, received: %q", eOut, args.String())
	}

}
func TestCgrIORecordWriterFlush(t *testing.T) {
	rw := new(CgrIORecordWriter)
	rw.Flush()
}
