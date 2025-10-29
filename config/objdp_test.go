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
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/utils"
)

func TestObjDPNewObjectDP(t *testing.T) {

	type args struct {
		obj     any
		prfxSlc []string
	}

	tests := []struct {
		name string
		args args
		exp  utils.DataProvider
	}{
		{
			name: "new objectDP",
			args: args{obj: "test", prfxSlc: []string{}},
			exp:  &ObjectDP{obj: "test", cache: make(map[string]any), prfxSls: []string{}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			rcv := NewObjectDP(tt.args.obj, tt.args.prfxSlc)

			if rcv.String() != tt.exp.String() {
				t.Errorf("recived %+s, expected %+s", rcv, tt.exp.String())
			}
		})
	}
}

func TestObjDPSetCache(t *testing.T) {

	type args struct {
		path string
		val  any
	}

	tests := []struct {
		name string
		args args
	}{
		{
			name: "set cache",
			args: args{"test", "val1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			odp := ObjectDP{"test", map[string]any{}, []string{}}

			odp.setCache(tt.args.path, tt.args.val)

			if odp.cache["test"] != "val1" {
				t.Error("didn't set cache")
			}
		})
	}
}

func TestObjDPGetCache(t *testing.T) {

	type exp struct {
		val any
		has bool
	}

	tests := []struct {
		name string
		arg  string
		exp  exp
	}{
		{
			name: "get cache",
			arg:  "test",
			exp:  exp{val: "val1", has: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			odp := ObjectDP{"test", map[string]any{}, []string{}}

			odp.setCache(tt.arg, tt.exp.val)
			rcv, has := odp.getCache(tt.arg)

			if !has {
				t.Error("didn't get cache")
			}

			if rcv != tt.exp.val {
				t.Errorf("recived %s, expected %s", rcv, tt.exp.val)
			}

		})
	}
}

func TestObjDPFieldAsInterface(t *testing.T) {

	odp := ObjectDP{"test", map[string]any{}, []string{}}

	odp.setCache("test", "val1")

	type exp struct {
		data any
		err  bool
	}

	tests := []struct {
		name    string
		arg     []string
		exp     exp
		slcPrfx []string
	}{
		{
			name:    "found in cache",
			arg:     []string{"test"},
			exp:     exp{data: "val1", err: false},
			slcPrfx: []string{"!", "."},
		},
		{
			name:    "object has prefix slice and length of field pat his smaller than lenght of prefix slice",
			arg:     []string{"test1"},
			exp:     exp{data: nil, err: true},
			slcPrfx: []string{"!", "."},
		},
		{
			name:    "has slice prefix different from field path",
			arg:     []string{"test1", "test2"},
			exp:     exp{data: nil, err: true},
			slcPrfx: []string{"!", "."},
		},
		{
			name:    "has slice prefix",
			arg:     []string{"!", "."},
			exp:     exp{data: nil, err: false},
			slcPrfx: []string{"!", "."},
		},
		{
			name:    "has selector with error",
			arg:     []string{"test[0", "."},
			exp:     exp{data: nil, err: true},
			slcPrfx: []string{},
		},
		{
			name:    "has selector",
			arg:     []string{"test[0]", "."},
			exp:     exp{data: nil, err: true},
			slcPrfx: []string{},
		},
	}

	for _, tt := range tests {

		odp.prfxSls = tt.slcPrfx

		data, err := odp.FieldAsInterface(tt.arg)

		if !tt.exp.err {
			if err != nil {
				t.Fatal("was not expecting an error")
			}
		} else {
			if err == nil {
				t.Fatal("was expecting an error")
			}
		}

		if !reflect.DeepEqual(data, tt.exp.data) {
			t.Errorf("recived %v, expected %v", data, tt.exp.data)
		}

	}
}

func TestObjDPFieldAsString(t *testing.T) {

	type exp struct {
		data string
		err  bool
	}

	tests := []struct {
		name string
		args []string
		exp  exp
	}{
		{
			name: "error",
			args: []string{"123"},
			exp:  exp{data: "", err: true},
		},
		{
			name: "error",
			args: []string{},
			exp:  exp{data: "", err: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			odp := ObjectDP{"test", map[string]any{}, []string{}}

			data, err := odp.FieldAsString(tt.args)

			if tt.exp.err {
				if err == nil {
					t.Error("was expecting an error")
				}
			} else {
				if err != nil {
					t.Error("was not expecting an error")
				}
			}

			if data != tt.exp.data {
				t.Errorf("recived %s, expected %s", data, tt.exp.data)
			}
		})
	}
}

func TestObjDPRemoteHost(t *testing.T) {

	odp := ObjectDP{"test", map[string]any{}, []string{}}

	rcv := odp.RemoteHost()

	if rcv.String() != "local" {
		t.Errorf("recived %s", rcv.String())
	}
}

func TestOBJDPFieldAsInterface(t *testing.T) {
	str := "test"
	type test struct {
		Field map[string][]string
	}
	tst := test{
		Field: map[string][]string{str: {str}},
	}
	objDP := &ObjectDP{
		obj:     tst,
		cache:   map[string]any{},
		prfxSls: []string{},
	}

	rcv, err := objDP.FieldAsInterface([]string{"Field", "test[0]"})

	if err != nil {
		t.Error(err)
	}

	if rcv != str {
		t.Error(rcv)
	}

	exp := map[string]any{"Field": map[string][]string{str: {str}}, "Field.test": []string{str}, "Field.test[0]": str}

	if !reflect.DeepEqual(exp, objDP.cache) {
		t.Errorf("\nexpected: %s\nreceived: %s\n", utils.ToJSON(exp), utils.ToJSON(objDP.cache))
	}
}
