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
	"fmt"
	"strconv"
	"strings"
	"time"
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
	var canCast bool
	out, canCast = CastFieldIfToString(outIface)
	if !canCast {
		return "", NewErrStringCast(outIface)
	}
	return
}

// DataConverter represents functions which should convert input into output
type DataConverter interface {
	Convert(interface{}) (interface{}, error)
}

// NewDataConverter is a factory of converters
func NewDataConverter(params string) (
	conv DataConverter, err error) {
	switch {
	case params == MetaUsageSeconds:
		return NewDurationSecondsConverter("")
	case strings.HasPrefix(params, MetaRound):
		if len(params) == len(MetaRound) { // no extra params, defaults implied
			return NewRoundConverter("")
		}
		return NewRoundConverter(params[len(MetaRound)+1:])
	default:
		return nil,
			fmt.Errorf("unsupported converter definition: <%s>",
				params)
	}
}

func NewDurationSecondsConverter(params string) (
	hdlr DataConverter, err error) {
	return new(DurationSecondsConverter), nil
}

// UsageSecondsDataConverter converts duration into seconds encapsulated in float64
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

func NewRoundConverter(params string) (hdlr DataConverter, err error) {
	rc := new(RoundConverter)
	var paramsSplt []string
	if params != "" {
		paramsSplt = strings.Split(params, InInFieldSep)
	}
	switch len(paramsSplt) {
	case 2:
		rc.Method = paramsSplt[1]
		fallthrough
	case 1:
		if rc.Decimals, err = strconv.Atoi(paramsSplt[0]); err != nil {
			return nil, fmt.Errorf("%s converter needs integer as decimals, have: <%s>",
				MetaRound, paramsSplt[0])
		}
		fallthrough
	case 0:
		rc.Method = ROUNDING_MIDDLE
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

func (rnd *RoundConverter) Convert(in interface{}) (
	out interface{}, err error) {
	var inFloat float64
	if inFloat, err = IfaceAsFloat64(in); err != nil {
		return
	}
	out = Round(inFloat, rnd.Decimals, rnd.Method)
	return
}
