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
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/cgrates/sipingo"
	"github.com/nyaruka/phonenumbers"
)

// DataConverters groups together multiple converters,
// executing optimized conversions
type DataConverters []DataConverter

// ConvertString converts from and to string
func (dcs DataConverters) ConvertString(in string) (out string, err error) {
	outIface := any(in)
	for _, cnv := range dcs {
		if outIface, err = cnv.Convert(outIface); err != nil {
			return
		}
	}
	return IfaceAsString(outIface), nil
}

// DataConverter represents functions which should convert input into output
type DataConverter interface {
	Convert(any) (any, error)
}

// NewDataConverter is a factory of converters
func NewDataConverter(params string) (conv DataConverter, err error) {
	switch {
	case params == MetaDurationSeconds:
		return NewDurationSecondsConverter(EmptyString)
	case params == MetaDurationNanoseconds:
		return NewDurationNanosecondsConverter(EmptyString)
	case params == MetaDurationMinutes:
		return new(DurationMinutesConverter), nil
	case strings.HasPrefix(params, MetaRound):
		if len(params) == len(MetaRound) { // no extra params, defaults implied
			return NewRoundConverter(EmptyString)
		}
		return NewRoundConverter(params[len(MetaRound)+1:])
	case strings.HasPrefix(params, MetaMultiply):
		if len(params) == len(MetaMultiply) { // no extra params, defaults implied
			return NewMultiplyConverter(EmptyString)
		}
		return NewMultiplyConverter(params[len(MetaMultiply)+1:])
	case strings.HasPrefix(params, MetaDivide):
		if len(params) == len(MetaDivide) { // no extra params, defaults implied
			return NewDivideConverter(EmptyString)
		}
		return NewDivideConverter(params[len(MetaDivide)+1:])
	case params == MetaJSON:
		return new(JSONConverter), nil
	case params == MetaDuration:
		return NewDurationConverter(EmptyString)
	case strings.HasPrefix(params, MetaDurationFormat):
		if params == MetaDurationFormat {
			return NewDurationFormatConverter(EmptyString)
		}
		return NewDurationFormatConverter(params[len(MetaDurationFormat)+1:])
	case params == MetaIP2Hex:
		return new(IP2HexConverter), nil
	case params == MetaString2Hex:
		return new(String2HexConverter), nil
	case params == MetaSIPURIHost:
		return new(SIPURIHostConverter), nil
	case params == MetaSIPURIUser:
		return new(SIPURIUserConverter), nil
	case params == MetaSIPURIMethod:
		return new(SIPURIMethodConverter), nil
	case params == MetaUnixTime:
		return new(UnixTimeConverter), nil
	case params == MetaLen:
		return new(LengthConverter), nil
	case params == MetaSlice:
		return new(SliceConverter), nil
	case params == MetaFloat64:
		return new(Float64Converter), nil
	case params == E164DomainConverter:
		return new(e164DomainConverter), nil
	case params == E164Converter:
		return new(e164Converter), nil
	case params == URLDecConverter:
		return new(URLDecodeConverter), nil
	case params == URLEncConverter:
		return new(URLEncodeConverter), nil
	case strings.HasPrefix(params, MetaLibPhoneNumber):
		if len(params) == len(MetaLibPhoneNumber) {
			return NewPhoneNumberConverter(EmptyString)
		}
		return NewPhoneNumberConverter(params[len(MetaLibPhoneNumber)+1:])
	case strings.HasPrefix(params, MetaTimeString):
		var paramsStr string
		if len(params) > len(MetaTimeString) {
			paramsStr = params[len(MetaTimeString)+1:]
		}
		return NewTimeStringConverter(paramsStr)
	case strings.HasPrefix(params, MetaRandom):
		if len(params) == len(MetaRandom) { // no extra params, defaults implied
			return NewRandomConverter(EmptyString)
		}
		return NewRandomConverter(params[len(MetaRandom)+1:])
	case strings.HasPrefix(params, MetaStrip):
		return NewStripConverter(params)
	case strings.HasPrefix(params, MetaGigawords):
		return new(GigawordsConverter), nil
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

func (mS *DurationSecondsConverter) Convert(in any) (
	out any, err error) {
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

func (mS *DurationNanosecondsConverter) Convert(in any) (
	out any, err error) {
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
		rc.Method = MetaRoundingMiddle
	case 1:
		if rc.Decimals, err = strconv.Atoi(paramsSplt[0]); err != nil {
			return nil, fmt.Errorf("%s converter needs integer as decimals, have: <%s>",
				MetaRound, paramsSplt[0])
		}
		rc.Method = MetaRoundingMiddle
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

func (rnd *RoundConverter) Convert(in any) (out any, err error) {
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

func (m *MultiplyConverter) Convert(in any) (out any, err error) {
	var inFloat64 float64
	if inFloat64, err = IfaceAsFloat64(in); err != nil {
		return nil, err
	}
	out = inFloat64 * m.Value
	return
}

func NewDivideConverter(constructParams string) (hdlr DataConverter, err error) {
	if constructParams == EmptyString {
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

func (m *DivideConverter) Convert(in any) (out any, err error) {
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

func (mS *DurationConverter) Convert(in any) (
	out any, err error) {
	return IfaceAsDuration(in)
}

func NewDurationFormatConverter(params string) (hdlr DataConverter, err error) {
	return &DurationFormatConverter{Layout: params}, nil
}

// DurationFormatConverter formats duration in the same way Time is formatted as string
type DurationFormatConverter struct {
	Layout string
}

func (dfc *DurationFormatConverter) Convert(in any) (
	out any, err error) {
	z := time.Unix(0, 0).UTC()
	dur, err := IfaceAsDuration(in)
	if err != nil {
		return nil, err
	}
	if dfc.Layout == EmptyString {
		out = z.Add(time.Duration(dur)).Format("15:04:05")
		return
	}
	out = z.Add(time.Duration(dur)).Format(dfc.Layout)
	return
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

func (lc *PhoneNumberConverter) Convert(in any) (out any, err error) {
	num, err := phonenumbers.Parse(IfaceAsString(in), lc.CountryCode)
	if err != nil {
		return nil, err
	}
	return phonenumbers.Format(num, lc.Format), nil
}

// IP2HexConverter will transform ip to hex
type IP2HexConverter struct{}

// Convert implements DataConverter interface
func (*IP2HexConverter) Convert(in any) (out any, err error) {
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

// SIPURIHostConverter will return the
type SIPURIHostConverter struct{}

// Convert implements DataConverter interface
func (*SIPURIHostConverter) Convert(in any) (out any, err error) {
	return sipingo.HostFrom(IfaceAsString(in)), nil
}

// SIPURIUserConverter will return the
type SIPURIUserConverter struct{}

// Convert implements DataConverter interface
func (*SIPURIUserConverter) Convert(in any) (out any, err error) {
	return sipingo.UserFrom(IfaceAsString(in)), nil
}

// SIPURIMethodConverter will return the
type SIPURIMethodConverter struct{}

// Convert implements DataConverter interface
func (*SIPURIMethodConverter) Convert(in any) (out any, err error) {
	return sipingo.MethodFrom(IfaceAsString(in)), nil
}

// String2HexConverter will transform the string to hex
type String2HexConverter struct{}

// Convert implements DataConverter interface
func (*String2HexConverter) Convert(in any) (o any, err error) {
	var out string
	if out = hex.EncodeToString([]byte(IfaceAsString(in))); len(out) == 0 {
		o = out
		return
	}
	o = "0x" + out
	return
}

// UnixTimeConverter converts the interface in the unix time
type UnixTimeConverter struct{}

// Convert implements DataConverter interface
func (tS *UnixTimeConverter) Convert(in any) (
	out any, err error) {
	var tm time.Time
	if tm, err = ParseTimeDetectLayout(in.(string), EmptyString); err != nil {
		return
	}
	out = tm.Unix()
	return
}

func NewRandomConverter(params string) (dflr DataConverter, err error) {
	randConv := &RandomConverter{}
	if params == EmptyString {
		dflr = randConv
		return
	}
	sls := strings.Split(params, InInFieldSep)
	switch len(sls) {
	case 2:
		if sls[0] != EmptyString {
			if randConv.begin, err = strconv.Atoi(sls[0]); err != nil {
				return
			}
		}
		if sls[1] != EmptyString {
			if randConv.end, err = strconv.Atoi(sls[1]); err != nil {
				return
			}
		}
	case 1:
		if randConv.begin, err = strconv.Atoi(sls[0]); err != nil {
			return
		}
	}
	dflr = randConv
	return
}

type RandomConverter struct {
	begin, end int
}

// Convert implements DataConverter interface
func (rC *RandomConverter) Convert(in any) (
	out any, err error) {
	if rC.begin == 0 {
		if rC.end == 0 {
			return rand.Int(), nil
		} else {
			return rand.Intn(rC.end), nil
		}
	} else {
		if rC.end == 0 {
			return rand.Int() + rC.begin, nil
		} else {
			return int(RandomInteger(int64(rC.begin), int64(rC.end))), nil
		}
	}
}

// LengthConverter returns the lenght of the slice
type LengthConverter struct{}

// Convert implements DataConverter interface
func (LengthConverter) Convert(in any) (out any, err error) {
	switch val := in.(type) {
	case string:
		return len(val), nil
	case []string:
		return len(val), nil
	case []any:
		return len(val), nil
	case []bool:
		return len(val), nil
	case []int:
		return len(val), nil
	case []int8:
		return len(val), nil
	case []int16:
		return len(val), nil
	case []int32:
		return len(val), nil
	case []int64:
		return len(val), nil
	case []uint:
		return len(val), nil
	case []uint8:
		return len(val), nil
	case []uint16:
		return len(val), nil
	case []uint32:
		return len(val), nil
	case []uint64:
		return len(val), nil
	case []uintptr:
		return len(val), nil
	case []float32:
		return len(val), nil
	case []float64:
		return len(val), nil
	case []complex64:
		return len(val), nil
	case []complex128:
		return len(val), nil
	default:
		return len(IfaceAsString(val)), nil
	}
}

// SliceConverter converts the interface in the unix time
type SliceConverter struct{}

// Convert implements DataConverter interface
func (SliceConverter) Convert(in any) (out any, err error) {
	switch val := in.(type) {
	case []string,
		[]any,
		[]bool,
		[]int,
		[]int8,
		[]int16,
		[]int32,
		[]int64,
		[]uint,
		[]uint8,
		[]uint16,
		[]uint32,
		[]uint64,
		[]uintptr,
		[]float32,
		[]float64,
		[]complex64,
		[]complex128:
		return val, nil
	default:
		src := IfaceAsString(in)
		if strings.HasPrefix(src, IdxStart) &&
			strings.HasSuffix(src, IdxEnd) { // it has a similar structure to a json marshaled slice
			var slice []any
			if err := json.Unmarshal([]byte(src), &slice); err == nil { // no error when unmarshal safe to asume that this is a slice
				return slice, nil
			}
		}
		return src, nil
	}
}

type Float64Converter struct{}

// Convert implements DataConverter interface
func (Float64Converter) Convert(in any) (any, error) {
	return IfaceAsFloat64(in)
}

// e164DomainConverter extracts the domain part out of a NAPTR name record
type e164DomainConverter struct{}

func (e164DomainConverter) Convert(in any) (any, error) {
	name := IfaceAsString(in)
	if i := strings.Index(name, ".e164."); i != -1 {
		name = name[i:]
	}
	return strings.Trim(name, "."), nil
}

// e164Converter extracts the E164 address out of a NAPTR name record
type e164Converter struct{}

func (e164Converter) Convert(in any) (any, error) {
	name := IfaceAsString(in)
	i := strings.Index(name, ".e164.")
	if i == -1 {
		return nil, errors.New("unknown format")
	}
	return ReverseString(
		strings.Replace(name[:i], ".", "", -1)), nil
}

// JSONConverter converts an object to json string
type JSONConverter struct{}

func (jsnC JSONConverter) Convert(in any) (any, error) {
	b, err := json.Marshal(in)
	if err != nil {
		return EmptyString, err
	}
	return string(b), nil
}

// StripConverter strips the prefix, the suffix or both from a string.
type StripConverter struct {
	side   string // side represents which part of the string to strip: prefix, suffix, or both.
	substr string // substr represents the substring to be removed from the string.
	amount int    // amount represents the number of characters to be removed from the string.
}

// NewStripConverter initializes and returns a new StripConverter with configurations
// based on the provided parameters in the input string. Each parameter in the input
// string should be separated by ':'.
//
// The input string must follow one of the following formats:
//  1. "*strip:<side>:<amount>"
//  2. "*strip:<side>:<substring>[:<amount>]"
//  3. "*strip:<side>:*char:<substring>[:<amount>]"
//
// Explanation of placeholders:
//   - <side>: Specifies which part of the string to strip. Must be one of "*prefix", "*suffix", or "*both".
//   - <substring>: Identifies the substring to remove. It can be a specific string, "*nil" for null characters,
//     "*space" for spaces, or any other character.
//   - <amount> (optional): Determines the number of characters to remove. If omitted, all instances of <substring>
//     are removed.
//
// Examples:
//   - "*strip:*prefix:5": Removes the first 5 characters from the string's prefix.
//   - "*strip:*suffix:*nil": Eliminates all trailing null characters in the string.
//   - "*strip:*both:*space:2": Clears 2 spaces from both the prefix and suffix of the string.
//   - "*strip:*suffix:*char:abc": Removes the substring "abc" from the suffix of the string.
//   - "*strip:*prefix:*char:abc:2": Strips the substring "abc" from the prefix of the string, repeated 2 times.
func NewStripConverter(params string) (DataConverter, error) {
	paramSlice := strings.Split(params, InInFieldSep)
	paramCount := len(paramSlice)
	if paramCount < 3 || paramCount > 5 {
		return nil, errors.New("strip converter: invalid number of parameters (should have 3, 4 or 5)")
	}
	sc := StripConverter{
		side:   paramSlice[1],
		substr: paramSlice[2],
		amount: -1,
	}
	var err error
	switch sc.substr {
	case EmptyString:
		return nil, errors.New("strip converter: substr parameter cannot be empty")
	case MetaNil, MetaSpace:
		if paramCount == 5 {
			return nil, errors.New("strip converter: cannot have 5 params in *nil/*space case")
		}
		if sc.substr == MetaNil {
			sc.substr = "\u0000"
		} else {
			sc.substr = " "
		}
		if paramCount == 4 {
			sc.amount, err = strconv.Atoi(paramSlice[3])
			if err != nil {
				return nil, fmt.Errorf("strip converter: invalid amount parameter (%w)", err)
			}
			sc.substr = strings.Repeat(sc.substr, sc.amount)
		}
	case MetaChar:
		if paramCount < 4 || paramSlice[3] == EmptyString {
			return nil, errors.New("strip converter: usage of *char implies the need of 4 or 5 non-empty params")
		}
		sc.substr = paramSlice[3]
		if paramCount == 5 {
			sc.amount, err = strconv.Atoi(paramSlice[4])
			if err != nil {
				return nil, fmt.Errorf("strip converter: invalid amount parameter (%w)", err)
			}
			sc.substr = strings.Repeat(sc.substr, sc.amount)
		}
	default:
		if paramCount > 3 {
			return nil, errors.New("strip converter: just the amount specified, cannot have more than 3 params")
		}
		sc.amount, err = strconv.Atoi(paramSlice[2])
		if err != nil {
			return nil, fmt.Errorf("strip converter: invalid amount parameter (%w)", err)
		}
		sc.substr = ""
	}
	return sc, nil
}

// Convert trims the input string based on the StripConverter's configuration.
// It returns a CAST_FAILED error if the input is not a string.
func (sc StripConverter) Convert(in any) (any, error) {
	str, ok := in.(string)
	if !ok {
		return nil, fmt.Errorf("strip converter: %w", ErrCastFailed)
	}
	if sc.amount <= 0 && sc.amount != -1 {
		return str, nil
	}
	switch sc.side {
	case MetaPrefix:
		if sc.substr == EmptyString {
			if sc.amount < len(str) {
				return str[sc.amount:], nil
			}
			return EmptyString, nil
		}
		if sc.amount != -1 {
			return strings.TrimPrefix(str, sc.substr), nil
		}
		return strings.TrimLeft(str, sc.substr), nil
	case MetaSuffix:
		if sc.substr == EmptyString {
			if sc.amount < len(str) {
				return str[:len(str)-sc.amount], nil
			}
			return EmptyString, nil
		}
		if sc.amount != -1 {
			return strings.TrimSuffix(str, sc.substr), nil
		}
		return strings.TrimRight(str, sc.substr), nil
	case MetaBoth:
		if sc.substr == EmptyString {
			if sc.amount*2 < len(str) {
				return str[sc.amount : len(str)-sc.amount], nil
			}
			return EmptyString, nil
		}
		if sc.amount != -1 {
			str = strings.TrimPrefix(str, sc.substr)
			return strings.TrimSuffix(str, sc.substr), nil
		}
		return strings.Trim(str, sc.substr), nil
	default:
		return EmptyString, errors.New("strip converter: invalid side parameter")
	}
}

// URLDecodeConverter converts an URL with encoded special characters back to original string.
type URLDecodeConverter struct{}

func (URLDecodeConverter) Convert(in any) (any, error) {
	urlStr := IfaceAsString(in)
	query, err := url.QueryUnescape(urlStr)
	if err != nil {
		return nil, err
	}
	return query, nil
}

// URLEncodeConverter converts a string with special characters to a valid URL format.
type URLEncodeConverter struct{}

func (URLEncodeConverter) Convert(in any) (any, error) {
	urlStr := IfaceAsString(in)
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}
	if parsedURL.Host == EmptyString {
		return url.QueryEscape(parsedURL.Path), nil
	}
	if len(parsedURL.Query()) != 0 {
		parsedURL.RawQuery = parsedURL.Query().Encode()
	}
	return parsedURL.String(), nil
}

// GigawordsConverter converts a value in Gigawords to octects
type GigawordsConverter struct{}

func (GigawordsConverter) Convert(in any) (any, error) {
	gigawordsValue, err := IfaceAsInt64(in)
	if err != nil {
		return nil, err
	}
	totalOctects := (gigawordsValue * int64(math.Pow(2, 32))) // 2^32
	return totalOctects, nil
}

type DurationMinutesConverter struct{}

func (mS *DurationMinutesConverter) Convert(in any) (
	out any, err error) {
	var inDur time.Duration
	if inDur, err = IfaceAsDuration(in); err != nil {
		return nil, err
	}
	out = inDur.Minutes()
	return
}

type TimeStringConverter struct {
	loc    *time.Location
	layout string
}

func NewTimeStringConverter(params string) (DataConverter, error) {
	locStr := "Local"
	layout := time.DateTime
	if len(params) != 0 {
		values := strings.SplitN(params, InInFieldSep, 2)
		if values[0] != EmptyString {
			locStr = values[0]
		}
		if len(values) > 1 {
			if values[1] != EmptyString {
				layout = values[1]
			}
		}
	}
	loc, err := time.LoadLocation(locStr)
	if err != nil {
		return nil, err
	}
	return TimeStringConverter{loc: loc, layout: layout}, nil
}

func (ts TimeStringConverter) Convert(in any) (out any, err error) {
	var tm time.Time
	switch val := in.(type) {
	case string:
		tm, err = ParseTimeDetectLayout(val, EmptyString)
		if err != nil {
			return nil, err
		}
	case time.Time:
		tm = val
	default:
		return nil, fmt.Errorf("*localtime converter: unsupported input")
	}
	tm = tm.In(ts.loc)
	return tm.Format(ts.layout), nil
}
