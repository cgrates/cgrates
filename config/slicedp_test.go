/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package config

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/utils"
)

func TestSliceDPNewSliceDP(t *testing.T) {

	slc := []string{"test", "test2"}
	rcv := NewSliceDP(slc)
	exp := &SliceDP{req: slc, cache: utils.MapStorage{}}

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("recived %v, expected %v", rcv, exp)
	}
}

func TestSliceDPString(t *testing.T) {

	slc := []string{"test", "test2"}
	cP := NewSliceDP(slc)
	rcv := cP.String()
	exp := "{}"

	if rcv != exp {
		t.Errorf("recived %v, expected %v", rcv, exp)
	}
}

func TestSliceDPFieldAsInterface(t *testing.T) {

	slc := []string{"0", "1", "2"}
	slc2 := []string{"0"}
	slc3 := []string{"test"}
	slc4 := []string{"4"}
	cp := SliceDP{req: slc2, cache: utils.MapStorage{"0": "val1"}}

	type exp struct {
		data any
		err  error
	}

	tests := []struct {
		name string
		arg  []string
		exp  exp
	}{
		{
			name: "empty field path",
			arg:  []string{},
			exp:  exp{data: nil, err: nil},
		},
		{
			name: "invalid field path",
			arg:  slc,
			exp:  exp{data: nil, err: fmt.Errorf("Invalid fieldPath %+v", slc)},
		},
		{
			name: "item found in cache",
			arg:  slc2,
			exp:  exp{data: "val1", err: nil},
		},
		{
			name: "strings.Atoi error",
			arg:  slc3,
			exp:  exp{data: nil},
		},
		{
			name: "error not found",
			arg:  slc4,
			exp:  exp{data: nil, err: utils.ErrNotFound},
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := cp.FieldAsInterface(tt.arg)

			if i == 3 {
				if err == nil {
					t.Fatal("was expecting an error")
				}
			} else {
				if !reflect.DeepEqual(err, tt.exp.err) {
					t.Fatalf("recived %s, expected %s", err, tt.exp.err)
				}
			}

			if !reflect.DeepEqual(data, tt.exp.data) {
				t.Errorf("recived %v, expected %v", data, tt.exp.data)
			}
		})
	}

	t.Run("no error", func(t *testing.T) {

		slcRec := []string{"0", "1"}
		slcArg := []string{"0"}
		cp := SliceDP{req: slcRec, cache: utils.MapStorage{"test": "val1"}}

		data, err := cp.FieldAsInterface(slcArg)

		if err != nil {
			t.Fatal("was not expecting an error:", err)
		}

		if !reflect.DeepEqual(data, "0") {
			t.Errorf("recived %v, expected %v", data, "0")
		}
	})

}

func TestSliceDPFieldAsString(t *testing.T) {

	slc := []string{"0", "1"}
	cP := NewSliceDP(slc)

	type exp struct {
		data string
		err  error
	}

	tests := []struct {
		name string
		arg  []string
		exp  exp
	}{
		{
			name: "err",
			arg:  []string{"2"},
			exp:  exp{"", utils.ErrNotFound},
		},
		{
			name: "no err",
			arg:  []string{"1"},
			exp:  exp{"1", nil},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := cP.FieldAsString(tt.arg)

			if err != tt.exp.err {
				t.Fatalf("recived %s, expected %s", err, tt.exp.err)
			}

			if !reflect.DeepEqual(data, tt.exp.data) {
				t.Errorf("recived %v, expected %v", data, tt.exp.data)
			}
		})
	}
}

func TestSliceDPRemoteHost(t *testing.T) {

	slc := []string{"test", "test2"}
	cP := NewSliceDP(slc)

	rcv := cP.RemoteHost()
	exp := utils.LocalAddr()

	if rcv.String() != exp.String() {
		t.Errorf("recived %s, expected %s", rcv, exp)
	}
}
