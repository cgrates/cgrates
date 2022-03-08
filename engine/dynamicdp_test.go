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
package engine

import (
	"reflect"
	"testing"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
	"github.com/nyaruka/phonenumbers"
)

func TestDynamicDPnewDynamicDP(t *testing.T) {

	expDDP := &dynamicDP{
		resConns:  []string{"conn1"},
		stsConns:  []string{"conn2"},
		actsConns: []string{"conn3"},
		tenant:    "cgrates.org",
		initialDP: utils.StringSet{
			"test": struct{}{},
		},
		cache: utils.MapStorage{},
		ctx:   context.Background(),
	}

	if rcv := newDynamicDP(context.Background(), []string{"conn1"}, []string{"conn2"},
		[]string{"conn3"}, "cgrates.org",
		utils.StringSet{"test": struct{}{}}); !reflect.DeepEqual(rcv, expDDP) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(expDDP), utils.ToJSON(rcv))
	}
}

func TestDynamicDPString(t *testing.T) {

	rcv := &dynamicDP{
		resConns:  []string{"conn1"},
		stsConns:  []string{"conn2"},
		actsConns: []string{"conn3"},
		tenant:    "cgrates.org",
		initialDP: utils.StringSet{
			"test": struct{}{},
		},
		cache: utils.MapStorage{},
		ctx:   context.Background(),
	}
	exp2 := "[\"test\"]"
	rcv2 := rcv.String()
	if !reflect.DeepEqual(rcv2, exp2) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(exp2), utils.ToJSON(rcv2))
	}
}

func TestDynamicDPFieldAsInterfaceErrFilename(t *testing.T) {

	rcv := &dynamicDP{
		resConns:  []string{"conn1"},
		stsConns:  []string{"conn2"},
		actsConns: []string{"conn3"},
		tenant:    "cgrates.org",
		initialDP: utils.StringSet{
			"test": struct{}{},
		},
		cache: utils.MapStorage{},
		ctx:   context.Background(),
	}
	_, err := rcv.FieldAsInterface([]string{""})
	if err == nil || err.Error() != "invalid fieldname <[]>" {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			"invalid fieldname <[]>", err)
	}
}

func TestDynamicDPFieldAsInterfaceErrLenFldPath(t *testing.T) {

	rcv := &dynamicDP{
		resConns:  []string{"conn1"},
		stsConns:  []string{"conn2"},
		actsConns: []string{"conn3"},
		tenant:    "cgrates.org",
		initialDP: utils.StringSet{
			"test": struct{}{},
		},
		cache: utils.MapStorage{},
		ctx:   context.Background(),
	}
	_, err := rcv.FieldAsInterface([]string{})
	if err == nil || err != utils.ErrNotFound {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ErrNotFound, err)
	}
}

func TestDynamicDPFieldAsInterface(t *testing.T) {

	DDP := &dynamicDP{
		resConns:  []string{"conn1"},
		stsConns:  []string{"conn2"},
		actsConns: []string{"conn3"},
		tenant:    "cgrates.org",
		initialDP: utils.StringSet{
			"test": struct{}{},
		},
		cache: utils.MapStorage{
			"testField": "testValue",
		},
		ctx: context.Background(),
	}
	result, err := DDP.FieldAsInterface([]string{"testField"})
	if err != nil {
		t.Error(err)
	}
	exp := "testValue"
	if !reflect.DeepEqual(result, exp) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(exp), utils.ToJSON(result))
	}
}

func TestLibphonenumberDPString(t *testing.T) {
	var pInt int32
	pInt = 2
	LDP := &libphonenumberDP{
		pNumber: &phonenumbers.PhoneNumber{
			CountryCode: &pInt,
		},
	}
	exp2 := "country_code:2 "
	rcv2 := LDP.String()
	if !reflect.DeepEqual(rcv2, exp2) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(exp2), utils.ToJSON(rcv2))
	}
}

func TestLibphonenumberDPFieldAsString(t *testing.T) {
	var pInt int32
	pInt = 2
	LDP := &libphonenumberDP{
		pNumber: &phonenumbers.PhoneNumber{
			CountryCode: &pInt,
		},
		cache: utils.MapStorage{
			"testField": "testValue",
		},
	}
	exp2 := "testValue"
	rcv2, err := LDP.FieldAsString([]string{"testField"})
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(rcv2, exp2) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(exp2), utils.ToJSON(rcv2))
	}
}

func TestLibphonenumberDPFieldAsStringError(t *testing.T) {
	var pInt int32
	pInt = 2
	LDP := &libphonenumberDP{
		pNumber: &phonenumbers.PhoneNumber{
			CountryCode: &pInt,
		},
		cache: utils.MapStorage{
			"testField": "testValue",
		},
	}
	_, err := LDP.FieldAsString([]string{"testField", "testField2"})
	if err == nil || err.Error() != "WRONG_PATH" {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			"WRONG_PATH", err)
	}
}

func TestLibphonenumberDPFieldAsInterfaceLen0(t *testing.T) {
	var pInt int32
	pInt = 2
	LDP := &libphonenumberDP{
		pNumber: &phonenumbers.PhoneNumber{
			CountryCode: &pInt,
		},
		cache: utils.MapStorage{
			"testField": "testValue",
		},
	}
	exp2 := &libphonenumberDP{
		pNumber: &phonenumbers.PhoneNumber{
			CountryCode: &pInt,
		},
		cache: utils.MapStorage{
			"testField": "testValue",
		}}
	exp2.setDefaultFields()

	rcv2, err := LDP.FieldAsInterface([]string{})
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(rcv2, exp2.cache) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			exp2.cache, rcv2)
	}
}
