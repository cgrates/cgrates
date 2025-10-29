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

func TestFVWDPNewFWVProvider(t *testing.T) {

	dP := NewFWVProvider("test")

	if dP == nil {
		t.Error("didn't recive the FWVProvider", dP)
	}
}

func TestFwvdpString(t *testing.T) {

	dP := NewFWVProvider("test")

	rcv := dP.String()

	if rcv != "{}" {
		t.Errorf("recived %s, expected {}", rcv)
	}
}

func TestFWVDPFieldAsInterface(t *testing.T) {

	dP := FWVProvider{req: "test", cache: utils.MapStorage{"test": "test"}}

	tests := []struct {
		name string
		arg  []string
		exp  any
		err  bool
	}{
		{
			name: "empty field path",
			arg:  []string{},
			exp:  nil,
			err:  false,
		},
		{
			name: "empty field path",
			arg:  []string{"a-b"},
			exp:  nil,
			err:  true,
		},
		{
			name: "empty field path",
			arg:  []string{"5-6"},
			exp:  "",
			err:  true,
		},
		{
			name: "empty field path",
			arg:  []string{"0-a"},
			exp:  nil,
			err:  true,
		},
		{
			name: "empty field path",
			arg:  []string{"0-6"},
			exp:  "",
			err:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			data, err := dP.FieldAsInterface(tt.arg)

			if tt.err {
				if err == nil {
					t.Fatal("was expecting an error")
				}
			} else {
				if err != nil {
					t.Fatal("was not expecting an error")
				}
			}

			if !reflect.DeepEqual(data, tt.exp) {
				t.Errorf("recived %v, expected %v", data, tt.exp)
			}
		})
	}
}

func TestFVWDPFieldAsString(t *testing.T) {

	dP := NewFWVProvider("test")

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
			name: "no err",
			arg:  []string{"0-1"},
			exp:  exp{"t", nil},
		},
		{
			name: "err",
			arg:  []string{"test"},
			exp:  exp{"", nil},
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			data, err := dP.FieldAsString(tt.arg)

			if i < 1 {
				if err != tt.exp.err {
					t.Fatalf("recived %s, expected %s", err, tt.exp.err)
				}
			} else {
				if err == nil {
					t.Fatalf("was expecting an error")
				}
			}

			if data != tt.exp.data {
				t.Errorf("recived %s, expected %s", data, tt.exp.data)
			}
		})
	}
}

func TestFVWDPRemoteHost(t *testing.T) {

	dP := NewFWVProvider("test")

	rcv := dP.RemoteHost()

	if rcv == nil {
		t.Error("didn't recive")
	}
}

func TestFWVProviderFieldAsInterface(t *testing.T) {
	fP := &FWVProvider{
		req:   "test",
		cache: utils.MapStorage{"test": 1},
	}
	fldPath := []string{"test"}

	rcv, err := fP.FieldAsInterface(fldPath)
	if err != nil {
		t.Error(err)
	}

	if rcv != 1 {
		t.Error(rcv)
	}
}
