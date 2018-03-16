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
)

// DataConverter represents functions which should convert input into output
type DataConverter interface {
	Convert(interface{}) (interface{}, error)
	ConvertAsString(interface{}) (string, error)
}

// NewDataConverter is a factory of converters
func NewDataConverter(params string) (
	conv DataConverter, err error) {
	switch {
	case params == MetaUsageSeconds:
		return NewUsageSecondsConverter(params[:len(MetaUsageSeconds)])
	case strings.HasPrefix(params, MetaRound):
		return NewRoundConverter(params[len(MetaRound)+1:])
	default:
		return nil,
			fmt.Errorf("unsupported converter definition: <%s>",
				params)
	}
}

func NewUsageSecondsConverter(params string) (
	hdlr DataConverter, err error) {
	return new(UsageSecondsConverter), nil
}

// UsageSecondsDataConverter transforms
type UsageSecondsConverter struct{}

func (mS *UsageSecondsConverter) Convert(in interface{}) (
	out interface{}, err error) {
	return
}

func (mS *UsageSecondsConverter) ConvertAsString(in interface{}) (
	out string, err error) {
	outIface, err := mS.Convert(in)
	if err != nil {
		return "", err
	}
	var canCast bool
	out, canCast = CastFieldIfToString(outIface)
	if !canCast {
		return "", NewErrStringCast(outIface)
	}
	return
}

func NewRoundConverter(params string) (hdlr DataConverter, err error) {
	fmt.Printf("NewRoundConverter, params: <%s>", params)
	rc := new(RoundConverter)
	paramsSplt := strings.Split(params, InInFieldSep)
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

// UsageSecondsDataConverter transforms
type RoundConverter struct {
	Decimals int
	Method   string
}

func (mS *RoundConverter) Convert(in interface{}) (
	out interface{}, err error) {
	return
}

func (mS *RoundConverter) ConvertAsString(in interface{}) (
	out string, err error) {
	outIface, err := mS.Convert(in)
	if err != nil {
		return "", err
	}
	var canCast bool
	out, canCast = CastFieldIfToString(outIface)
	if !canCast {
		return "", NewErrStringCast(outIface)
	}
	return
}
