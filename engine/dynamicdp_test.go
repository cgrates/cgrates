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
	"bytes"
	"log"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/nyaruka/phonenumbers"
)

func TestDynamicDpFieldAsInterface(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	ms := utils.MapStorage{}
	dDp := newDynamicDP([]string{}, []string{utils.ConcatenatedKey(utils.MetaInternal, utils.StatSConnsCfg)}, []string{}, []string{}, "cgrates.org", ms)
	clientconn := make(chan birpc.ClientConnector, 1)
	clientconn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.StatSv1GetQueueFloatMetrics: func(ctx *context.Context, args, reply any) error {
				rpl := &map[string]float64{
					"stat1": 31,
				}
				*reply.(*map[string]float64) = *rpl
				return nil
			},
		},
	}
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.StatSConnsCfg): clientconn,
	})
	SetConnManager(connMgr)
	if _, err := dDp.fieldAsInterface([]string{utils.MetaStats, "val", "val3"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if _, err := dDp.fieldAsInterface([]string{utils.MetaLibPhoneNumber, "+402552663", "val3"}); err != nil {
		t.Error(err)
	} else if _, err := dDp.fieldAsInterface([]string{utils.MetaLibPhoneNumber, "+402552663", "val3"}); err != nil {
		t.Error(err)
	} else if _, err := dDp.fieldAsInterface([]string{utils.MetaAsm, "+402552663", "val3"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}
}

func TestDDPFieldAsInterface(t *testing.T) {
	utils.Logger.SetLogLevel(4)
	utils.Logger.SetSyslog(nil)
	buf := new(bytes.Buffer)
	log.SetOutput(buf)
	defer func() {
		utils.Logger.SetLogLevel(0)
		log.SetOutput(os.Stderr)
	}()

	libDP, err := newLibPhoneNumberDP("+447975777666")
	if err != nil {
		t.Error(err)
	}
	phoneNm, canCast := libDP.(*libphonenumberDP)
	if !canCast {
		t.Error("can't convert interface")
	}
	dDP := &libphonenumberDP{
		pNumber: phoneNm.pNumber,
		cache: utils.MapStorage{
			"field": "val",
		},
	}
	dDP.pNumber.Extension = utils.StringPointer("+")
	dDP.pNumber.PreferredDomesticCarrierCode = utils.StringPointer("49 172")

	if val, err := dDP.fieldAsInterface([]string{"CountryCode"}); err != nil {
		t.Error(err)
	} else if val.(int32) != int32(44) {
		t.Errorf("expected %v,reveived %v", 44, val.(int32))
	}
	if val, err := dDP.fieldAsInterface([]string{"NationalNumber"}); err != nil {
		t.Error(err)
	} else if val.(uint64) != uint64(7975777666) {
		t.Errorf("expected %v,reveived %v", 7975777666, val)
	}
	if val, err := dDP.fieldAsInterface([]string{"Region"}); err != nil {
		t.Error(err)
	} else if val.(string) != "GB" {
		t.Errorf("expected %v,reveived %v", "GB", val)
	}
	if val, err := dDP.fieldAsInterface([]string{"NumberType"}); err != nil {
		t.Error(err)
	} else if val.(phonenumbers.PhoneNumberType) != 1 {
		t.Errorf("expected %v,reveived %v", 1, val)
	}
	if val, err := dDP.fieldAsInterface([]string{"GeoLocation"}); err != nil {
		t.Error(err)
	} else if val.(string) != "United Kingdom" {
		t.Errorf("expected %v,reveived %v", "United Kingdom", val)
	}
	if val, err := dDP.fieldAsInterface([]string{"Carrier"}); err != nil {
		t.Error(err)
	} else if val.(string) != "Orange" {
		t.Errorf("expected %v,reveived %v", "Orange", val)
	}
	if val, err := dDP.fieldAsInterface([]string{"LengthOfNationalDestinationCode"}); err != nil {
		t.Error(err)
	} else if val.(int) != 4 {
		t.Errorf("expected %v,reveived %v", 0, val)
	}
	if val, err := dDP.fieldAsInterface([]string{"RawInput"}); err != nil {
		t.Error(err)
	} else if val.(string) != "+447975777666" {
		t.Errorf("expected %v,reveived %v", "+447975777666", val)
	}
	if val, err := dDP.fieldAsInterface([]string{"Extension"}); err != nil {
		t.Error(err)
	} else if val.(string) != "+" {
		t.Errorf("expected %v,reveived %v", "+", val)
	}
	if val, err := dDP.fieldAsInterface([]string{"NumberOfLeadingZeros"}); err != nil {
		t.Error(err)
	} else if val.(int32) != int32(1) {
		t.Errorf("expected %v,reveived %v", int32(1), val)
	}
	if val, err := dDP.fieldAsInterface([]string{"ItalianLeadingZero"}); err != nil {
		t.Error(err)
	} else if val.(bool) != false {
		t.Errorf("expected %v,reveived %v", false, val)
	}
	if val, err := dDP.fieldAsInterface([]string{"PreferredDomesticCarrierCode"}); err != nil {
		t.Error(err)
	} else if val.(string) != "49 172" {
		t.Errorf("expected %v,reveived %v", "49 172", val)
	}
	if val, err := dDP.fieldAsInterface([]string{"CountryCodeSource"}); err != nil {
		t.Error(err)
	} else if val.(phonenumbers.PhoneNumber_CountryCodeSource) != phonenumbers.PhoneNumber_CountryCodeSource(1) {
		t.Errorf("expected %v,reveived %v", 1, val)
	}

	dDP = &libphonenumberDP{
		cache:   utils.MapStorage{},
		pNumber: &phonenumbers.PhoneNumber{},
	}
	expLog := `when getting GeoLocation for number`
	if _, err := dDP.fieldAsInterface([]string{"GeoLocation"}); err != nil {
		t.Error(err)
	} else if rcvLog := buf.String(); !strings.Contains(rcvLog, expLog) {
		t.Errorf("Logger %v doesn't contain %v", rcvLog, expLog)
	}
	dDP = &libphonenumberDP{
		cache:   utils.MapStorage{},
		pNumber: &phonenumbers.PhoneNumber{},
	}
	dDP.setDefaultFields()
	utils.Logger.SetLogLevel(0)
	log.SetOutput(os.Stderr)
	utils.Logger.SetLogLevel(4)
	utils.Logger.SetSyslog(nil)
	buf2 := new(bytes.Buffer)
	log.SetOutput(buf2)
	expLog = `when getting GeoLocation for number`
	if rcvLog := buf2.String(); strings.Contains(rcvLog, expLog) {
		t.Errorf("Logger %v doesn't contain %v", rcvLog, expLog)
	}
}
func TestLibphonenumberDPString(t *testing.T) {
	pInt := int32(2)
	LDP := &libphonenumberDP{
		pNumber: &phonenumbers.PhoneNumber{
			CountryCode: &pInt,
		},
	}
	exp2 := "country_code:2"
	rcv2 := LDP.String()
	if !reflect.DeepEqual(rcv2, exp2) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(exp2), utils.ToJSON(rcv2))
	}
}

func TestLibphonenumberDPFieldAsString(t *testing.T) {
	pInt := int32(2)
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
		t.Errorf("expected: <%+v>, received: <%+v>",
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
		t.Errorf("expected: <%v>, received: <%v>",
			"WRONG_PATH", err)
	}
}

func TestLibphonenumberDPFieldAsInterfaceLen0(t *testing.T) {
	pInt := int32(2)
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
		t.Errorf("expected: %+v, received: %+v",
			exp2.cache, rcv2)
	}
}

func TestNewLibPhoneNumberDPErr(t *testing.T) {
	num := "errNum"
	if _, err := newLibPhoneNumberDP(num); err != phonenumbers.ErrNotANumber {
		t.Error(err)
	}
}
