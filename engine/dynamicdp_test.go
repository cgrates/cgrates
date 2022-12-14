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
	var pInt int32 = 2
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
	var pInt int32 = 2
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
	var pInt int32 = 2
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
	var pInt int32 = 2
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

func TestNewLibPhoneNumberDPErr(t *testing.T) {

	number := "badNum"

	if _, err := newLibPhoneNumberDP(number); err != phonenumbers.ErrNotANumber {

		t.Error(err)
	}
}

func TestNewLibPhoneNumberDP(t *testing.T) {

	number := "+355123456789"

	num, err := phonenumbers.ParseAndKeepRawInput(number, utils.EmptyString)
	if err != nil {
		t.Error(err)
	}
	exp := utils.DataProvider(&libphonenumberDP{pNumber: num, cache: make(utils.MapStorage)})

	if err != nil {
		t.Error(err)
	}
	if rcv, err := newLibPhoneNumberDP(number); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected \n<%+v>, Received \n<%+v>", exp, rcv)
	}

}

func TestFieldAsInterfacelibphonenumberDP(t *testing.T) {

	dDP := &libphonenumberDP{
		pNumber: &phonenumbers.PhoneNumber{},
		cache: utils.MapStorage{
			"testField": "testValue",
		},
	}

	expErr := `invalid field path <[fld1 fld2 fld3]> for libphonenumberDP`
	if _, err := dDP.FieldAsInterface([]string{"fld1", "fld2", "fld3"}); err.Error() != expErr {
		t.Error(err)
	}
}

// unfinished do with buffer
func TestLibphonenumberDPfieldAsInterfaceGeoLocationErr(t *testing.T) {

	dDP := &libphonenumberDP{
		pNumber: &phonenumbers.PhoneNumber{},
		cache: utils.MapStorage{
			"testField": "testValue",
		},
	}

	if rcv, err := dDP.fieldAsInterface([]string{"GeoLocation"}); err != nil {
		t.Error(rcv, err)
	}
}

// unfinished do with buffer
func TestLibphonenumberDPfieldAsInterfaceCarrierErr(t *testing.T) {
	var pInt int32 = 49444444
	var nNum uint64 = 49233333333333
	dDP := &libphonenumberDP{
		pNumber: &phonenumbers.PhoneNumber{
			CountryCode:    &pInt,
			NationalNumber: &nNum,
		},
		cache: utils.MapStorage{
			"testField": "testValue",
		},
	}

	if rcv, err := dDP.fieldAsInterface([]string{"Carrier"}); err != nil {
		t.Error(rcv, err)
	}
}

func TestLibphonenumberDPfieldAsInterface(t *testing.T) {

	var pInt int32 = 49
	var nNum uint64 = 49
	var leadingZeros int32 = 0

	dDP := &libphonenumberDP{
		pNumber: &phonenumbers.PhoneNumber{
			CountryCode:                  &pInt,
			NationalNumber:               &nNum,
			Extension:                    utils.StringPointer("+"),
			RawInput:                     utils.StringPointer("+4917642092123"),
			ItalianLeadingZero:           utils.BoolPointer(true),
			NumberOfLeadingZeros:         &leadingZeros,
			CountryCodeSource:            phonenumbers.PhoneNumber_FROM_DEFAULT_COUNTRY.Enum(),
			PreferredDomesticCarrierCode: utils.StringPointer("262 02"),
		},
		cache: utils.MapStorage{
			"testField": "testValue",
		},
	}

	exp := interface{}(pInt)
	if rcv, err := dDP.fieldAsInterface([]string{"CountryCode"}); err != nil {
		t.Error(err)
	} else if rcv != exp {
		t.Errorf("Expected <%+v>, Received <%v>", exp, rcv)
	}

	exp = interface{}(nNum)
	if rcv, err := dDP.fieldAsInterface([]string{"NationalNumber"}); err != nil {
		t.Error(err)
	} else if rcv != exp {
		t.Errorf("Expected <%+v>, Received <%v>", exp, rcv)
	}

	exp = interface{}("DE")
	if rcv, err := dDP.fieldAsInterface([]string{"Region"}); err != nil {
		t.Error(err)
	} else if rcv != exp {
		t.Errorf("Expected <%+v>, Received <%v>", exp, rcv)
	}
	exp = interface{}(phonenumbers.PhoneNumberType(11))
	if rcv, err := dDP.fieldAsInterface([]string{"NumberType"}); err != nil {
		t.Error(err)
	} else if rcv != exp {
		t.Errorf("Expected <%+v>, Received <%v><%t>", exp, rcv, rcv)
	}

	exp = interface{}("Deutschland")
	if rcv, err := dDP.fieldAsInterface([]string{"GeoLocation"}); err != nil {
		t.Error(err)
	} else if rcv != exp {
		t.Errorf("Expected <%+v>, Received <%v>", exp, rcv)
	}
	// exp = interface{}(nNum)
	// if rcv, err := dDP.fieldAsInterface([]string{"Carrier"}); err != nil {
	// 	t.Error(err)
	// } else if rcv != exp {
	// 	t.Errorf("Expected <%+v>, Received <%v>", exp, rcv)
	// }
	exp = interface{}(0)
	if rcv, err := dDP.fieldAsInterface([]string{"LengthOfNationalDestinationCode"}); err != nil {
		t.Error(err)
	} else if rcv != exp {
		t.Errorf("Expected <%+v>, Received <%v>", exp, rcv)
	}
	exp = interface{}("+4917642092123")
	if rcv, err := dDP.fieldAsInterface([]string{"RawInput"}); err != nil {
		t.Error(err)
	} else if rcv != exp {
		t.Errorf("Expected <%+v>, Received <%v>", exp, rcv)
	}
	exp = interface{}("+")
	if rcv, err := dDP.fieldAsInterface([]string{"Extension"}); err != nil {
		t.Error(err)
	} else if rcv != exp {
		t.Errorf("Expected <%+v>, Received <%v>", exp, rcv)
	}
	exp = interface{}(leadingZeros)
	if rcv, err := dDP.fieldAsInterface([]string{"NumberOfLeadingZeros"}); err != nil {
		t.Error(err)
	} else if rcv != exp {
		t.Errorf("Expected <%+v>, Received <%v>", exp, rcv)
	}
	exp = interface{}(true)
	if rcv, err := dDP.fieldAsInterface([]string{"ItalianLeadingZero"}); err != nil {
		t.Error(err)
	} else if rcv != exp {
		t.Errorf("Expected <%+v>, Received <%v>", exp, rcv)
	}
	exp = interface{}("262 02")
	if rcv, err := dDP.fieldAsInterface([]string{"PreferredDomesticCarrierCode"}); err != nil {
		t.Error(err)
	} else if rcv != exp {
		t.Errorf("Expected <%+v>, Received <%v>", exp, rcv)
	}
	exp = interface{}(phonenumbers.PhoneNumber_FROM_DEFAULT_COUNTRY)
	if rcv, err := dDP.fieldAsInterface([]string{"CountryCodeSource"}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected <%+v>, Received <%v>", exp, rcv)
	}

}

func TestDynamicDPfieldAsInterfaceMetaLibPhoneNumber(t *testing.T) {

	dDP := &dynamicDP{
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

	dp, _ := newLibPhoneNumberDP("+4917642092123")
	expAsField, _ := dp.FieldAsInterface([]string{})
	exp := interface{}(expAsField)
	if rcv, err := dDP.fieldAsInterface([]string{utils.MetaLibPhoneNumber, "+4917642092123"}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected \n<%v>, Received \n<%v>", exp, rcv)
	}
}

func TestDynamicDPfieldAsInterfaceMetaLibPhoneNumberErr(t *testing.T) {

	dDP := &dynamicDP{
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

	if _, err := dDP.fieldAsInterface([]string{utils.MetaLibPhoneNumber, "inexistentNum"}); err != phonenumbers.ErrNotANumber {
		t.Error(err)
	}
}

// func TestDynamicDPfieldAsInterfaceNotFound(t *testing.T) {
// 	Cache.Clear(nil)

// 	ms := utils.MapStorage{}
// 	dDp := newDynamicDP(context.Background(), []string{}, []string{utils.ConcatenatedKey(utils.MetaInternal, utils.StatSConnsCfg)}, []string{}, "cgrates.org", ms)
// 	exp := ""
// 	if rcv, err := dDp.fieldAsInterface([]string{utils.MetaStats, "val", "val3"}); err == nil || err != utils.ErrNotFound {
// 		t.Error(err)
// 	} else if !reflect.DeepEqual(rcv, exp) {
// 		t.Errorf("Expected \n<%v>, Received \n<%v>", exp, rcv)
// 	}
// }
