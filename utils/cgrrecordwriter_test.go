// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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

type writer2 struct{}

func (*writer2) Write(p []byte) (n int, err error) { return 0, ErrNoMoreData }

func TestWrite(t *testing.T) {
	//empty check
	args := new(bytes.Buffer)
	rw := NewCgrIORecordWriter(args)
	record := []string{"test1", "test2"}
	err := rw.Write(record)
	if err != nil {
		t.Errorf("Expecting nil, received %+v", err)
	}
	eOut := "test1test2\n"
	if !reflect.DeepEqual(eOut, args.String()) {
		t.Errorf("Expected %q, received: %q", eOut, args.String())
	}
	//err check
	args2 := &writer2{}
	rw = NewCgrIORecordWriter(args2)
	record = []string{"test1", "test2"}
	err = rw.Write(record)
	if err != ErrNoMoreData {
		t.Errorf("Expecting %+v, received %v", ErrNoMoreData, err)
	}

}
func TestCgrIORecordWriterFlush(t *testing.T) {
	rw := new(CgrIORecordWriter)
	rw.Flush()
}
