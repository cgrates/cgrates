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
	"encoding/hex"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/nyaruka/phonenumbers"
)

// DataConverters groups together multiple converters,
// executing optimized conversions
type DataConverters []DataConverter

// ConvertString converts from and to string
func (dcs DataConverters) ConvertString(in string) (out string, err error) {
	outIface := interface{}(in)
	for _, cnv := range dcs {
		if outIface, err = cnv.Convert(outIface); err != nil {
			return
		}
	}
	return IfaceAsString(outIface), nil
}

// DataConverter represents functions which should convert input into output
type DataConverter interface {
	Convert(interface{}) (interface{}, error)
}

// NewDataConverter is a factory of converters
func NewDataConverter(params string) (conv DataConverter, err error) {
	switch {
	case params == MetaDurationSeconds:
		return NewDurationSecondsConverter("")
	case params == MetaDurationNanoseconds:
		return NewDurationNanosecondsConverter("")
	case strings.HasPrefix(params, MetaRound):
		if len(params) == len(MetaRound) { // no extra params, defaults implied
			return NewRoundConverter("")
		}
		return NewRoundConverter(params[len(MetaRound)+1:])
	case strings.HasPrefix(params, MetaMultiply):
		if len(params) == len(MetaMultiply) { // no extra params, defaults implied
			return NewMultiplyConverter("")
		}
		return NewMultiplyConverter(params[len(MetaMultiply)+1:])
	case strings.HasPrefix(params, MetaDivide):
		if len(params) == len(MetaDivide) { // no extra params, defaults implied
			return NewDivideConverter("")
		}
		return NewDivideConverter(params[len(MetaDivide)+1:])
	case params == MetaDuration:
		return NewDurationConverter("")
	case params == MetaIP2Hex:
		return new(IP2HexConverter), nil
	case params == MetaString2Hex:
		return new(String2HexConverter), nil
	case strings.HasPrefix(params, MetaLibPhoneNumber):
		if len(params) == len(MetaLibPhoneNumber) {
			return NewPhoneNumberConverter("")
		}
		return NewPhoneNumberConverter(params[len(MetaLibPhoneNumber)+1:])
	default:
		return nil, fmt.Errorf("unsupported converter definition: <%s>", params)
	}
}

func NewDataConverterMustCompile(params string) (conv DataConverter) {
	var err error
	if conv, err = NewDataConverter(params); err != nil {
		panic(fmt.Sprintf("parsing: <%s>, error: %s", params, err.Error()))
	}
	return
}

func NewDurationSecondsConverter(params string) (hdlr DataConverter, err error) {
	return new(DurationSecondsConverter), nil
}

// DurationSecondsConverter converts duration into seconds encapsulated in float64
type DurationSecondsConverter struct{}

func (mS *DurationSecondsConverter) Convert(in interface{}) (
	out interface{}, err error) {
	var inDur time.Duration
	if inDur, err = IfaceAsDuration(in); err != nil {
		return nil, err
	}
	out = inDur.Seconds()
	return
}

func NewDurationNanosecondsConverter(params string) (
	hdlr DataConverter, err error) {
	return new(DurationNanosecondsConverter), nil
}

// DurationNanosecondsConverter converts duration into nanoseconds encapsulated in int64
type DurationNanosecondsConverter struct{}

func (mS *DurationNanosecondsConverter) Convert(in interface{}) (
	out interface{}, err error) {
	var inDur time.Duration
	if inDur, err = IfaceAsDuration(in); err != nil {
		return nil, err
	}
	out = inDur.Nanoseconds()
	return
}

func NewRoundConverter(params string) (hdlr DataConverter, err error) {
	rc := new(RoundConverter)
	var paramsSplt []string
	if params != EmptyString {
		paramsSplt = strings.Split(params, InInFieldSep)
	}
	switch len(paramsSplt) {
	case 0:
		rc.Method = ROUNDING_MIDDLE
	case 1:
		if rc.Decimals, err = strconv.Atoi(paramsSplt[0]); err != nil {
			return nil, fmt.Errorf("%s converter needs integer as decimals, have: <%s>",
				MetaRound, paramsSplt[0])
		}
		rc.Method = ROUNDING_MIDDLE
	case 2:
		rc.Method = paramsSplt[1]
		if rc.Decimals, err = strconv.Atoi(paramsSplt[0]); err != nil {
			return nil, fmt.Errorf("%s converter needs integer as decimals, have: <%s>",
				MetaRound, paramsSplt[0])
		}
	default:
		return nil, fmt.Errorf("unsupported %s converter parameters: <%s>",
			MetaRound, params)
	}
	return rc, nil
}

// RoundConverter will round floats
type RoundConverter struct {
	Decimals int
	Method   string
}

func (rnd *RoundConverter) Convert(in interface{}) (out interface{}, err error) {
	var inFloat float64
	if inFloat, err = IfaceAsFloat64(in); err != nil {
		return
	}
	out = Round(inFloat, rnd.Decimals, rnd.Method)
	return
}

func NewMultiplyConverter(constructParams string) (hdlr DataConverter, err error) {
	if constructParams == EmptyString {
		return nil, ErrMandatoryIeMissingNoCaps
	}
	var val float64
	if val, err = strconv.ParseFloat(constructParams, 64); err != nil {
		return
	}
	return &MultiplyConverter{Value: val}, nil
}

// MultiplyConverter multiplies input with value in params
// encapsulates the output as float64 value
type MultiplyConverter struct {
	Value float64
}

func (m *MultiplyConverter) Convert(in interface{}) (out interface{}, err error) {
	var inFloat64 float64
	if inFloat64, err = IfaceAsFloat64(in); err != nil {
		return nil, err
	}
	out = inFloat64 * m.Value
	return
}

func NewDivideConverter(constructParams string) (hdlr DataConverter, err error) {
	if constructParams == "" {
		return nil, ErrMandatoryIeMissingNoCaps
	}
	var val float64
	if val, err = strconv.ParseFloat(constructParams, 64); err != nil {
		return
	}
	return &DivideConverter{Value: val}, nil
}

// DivideConverter divides input with value in params
// encapsulates the output as float64 value
type DivideConverter struct {
	Value float64
}

func (m *DivideConverter) Convert(in interface{}) (out interface{}, err error) {
	var inFloat64 float64
	if inFloat64, err = IfaceAsFloat64(in); err != nil {
		return nil, err
	}
	out = inFloat64 / m.Value
	return
}

func NewDurationConverter(params string) (hdlr DataConverter, err error) {
	return new(DurationConverter), nil
}

// DurationConverter converts duration into seconds encapsulated in float64
type DurationConverter struct{}

func (mS *DurationConverter) Convert(in interface{}) (
	out interface{}, err error) {
	return IfaceAsDuration(in)
}

// NewPhoneNumberConverter create a new phoneNumber converter
// If the format isn't specify by default we use NATIONAL
// Possible fromats are : E164(0) , INTERNATIONAL(1) , NATIONAL(2) ,RFC3966(3)
// Also ContryCode needs to be specified
func NewPhoneNumberConverter(params string) (
	pbDC DataConverter, err error) {
	lc := new(PhoneNumberConverter)
	var paramsSplt []string
	if params != EmptyString {
		paramsSplt = strings.Split(params, InInFieldSep)
	}
	switch len(paramsSplt) {
	case 2:
		lc.CountryCode = paramsSplt[0]
		frm, err := strconv.Atoi(paramsSplt[1])
		if err != nil {
			return nil, err
		}
		lc.Format = phonenumbers.PhoneNumberFormat(frm)
	case 1:
		lc.CountryCode = paramsSplt[0]
		lc.Format = 2
	default:
		return nil, fmt.Errorf("unsupported %s converter parameters: <%s>",
			MetaLibPhoneNumber, params)
	}
	return lc, nil
}

// PhoneNumberConverter converts
type PhoneNumberConverter struct {
	CountryCode string
	Format      phonenumbers.PhoneNumberFormat
}

func (lc *PhoneNumberConverter) Convert(in interface{}) (out interface{}, err error) {
	num, err := phonenumbers.Parse(IfaceAsString(in), lc.CountryCode)
	if err != nil {
		return nil, err
	}
	return phonenumbers.Format(num, lc.Format), nil
}

// HexConvertor will round floats
type IP2HexConverter struct{}

func (_ *IP2HexConverter) Convert(in interface{}) (out interface{}, err error) {
	var ip net.IP
	switch val := in.(type) {
	case string:
		ip = net.ParseIP(val)
	case net.IP:
		ip = val
	default:
		src := IfaceAsString(in)
		ip = net.ParseIP(src)
	}

	hx := hex.EncodeToString([]byte(ip))
	if len(hx) < 8 {
		return hx, nil
	}
	return "0x" + string([]byte(hx)[len(hx)-8:]), nil
}

// String2HexConverter will transform the string to hex
type String2HexConverter struct{}

// Convert implements DataConverter interface
func (*String2HexConverter) Convert(in interface{}) (o interface{}, err error) {
	var out string
	if out = hex.EncodeToString([]byte(IfaceAsString(in))); len(out) == 0 {
		o = out
		return
	}
	o = "0x" + out
	return
}
