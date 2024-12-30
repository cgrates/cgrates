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
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ericlagergren/decimal"
)

var (
	DecimalNaN     = &Decimal{}
	DecimalContext decimal.Context
)

func init() {
	d, _ := decimal.WithContext(DecimalContext).SetString("NaN")
	DecimalNaN = &Decimal{d}
}

func NewRoundingMode(rnd string) (decimal.RoundingMode, error) {
	switch rnd {
	case ToNearestEven:
		return decimal.ToNearestEven, nil
	case ToNearestAway:
		return decimal.ToNearestAway, nil
	case ToZero:
		return decimal.ToZero, nil
	case AwayFromZero:
		return decimal.AwayFromZero, nil
	case ToNegativeInf:
		return decimal.ToNegativeInf, nil
	case ToPositiveInf:
		return decimal.ToPositiveInf, nil
	case ToNearestTowardZero:
		return decimal.ToNearestTowardZero, nil
	default:
		return 7, fmt.Errorf("usupoorted rounding: <%q>", rnd)
	}
}
func RoundingModeToString(rnd decimal.RoundingMode) string {
	switch rnd {
	case decimal.ToNearestEven:
		return ToNearestEven
	case decimal.ToNearestAway:
		return ToNearestAway
	case decimal.ToZero:
		return ToZero
	case decimal.AwayFromZero:
		return AwayFromZero
	case decimal.ToNegativeInf:
		return ToNegativeInf
	case decimal.ToPositiveInf:
		return ToPositiveInf
	case decimal.ToNearestTowardZero:
		return ToNearestTowardZero
	default:
		return EmptyString
	}
}
func DivideBig(x, y *decimal.Big) *decimal.Big {
	if x == nil || y == nil {
		return nil
	}
	return decimal.WithContext(DecimalContext).Quo(x, y)
}

func DivideBigWithReminder(x, y *decimal.Big) (q *decimal.Big, r *decimal.Big) {
	if x == nil || y == nil {
		return
	}
	return decimal.WithContext(DecimalContext).QuoRem(x, y, decimal.WithContext(DecimalContext))
}

func MultiplyBig(x, y *decimal.Big) *decimal.Big {
	if x == nil || y == nil {
		return nil
	}
	return decimal.WithContext(DecimalContext).Mul(x, y)
}

func SumBig(x, y *decimal.Big) *decimal.Big {
	if x == nil {
		return y
	}
	if y == nil {
		return x
	}
	return decimal.WithContext(DecimalContext).Add(x, y)
}

func SubstractBig(x, y *decimal.Big) *decimal.Big {
	if x == nil || y == nil {
		return x
	}
	return decimal.WithContext(DecimalContext).Sub(x, y)
}

// MultiplyDecimal multiples two Decimals and returns the result
func MultiplyDecimal(x, y *Decimal) *Decimal {
	return &Decimal{MultiplyBig(x.Big, y.Big)}
}

// DivideDecimal divides two Decimals and returns the result
func DivideDecimal(x, y *Decimal) *Decimal {
	return &Decimal{DivideBig(x.Big, y.Big)}
}

// sumDecimal adds two Decimals and returns the result
func SumDecimal(x, y *Decimal) *Decimal {
	if x == nil {
		return y
	}
	if y == nil {
		return x
	}
	return &Decimal{SumBig(x.Big, y.Big)}
}

func SubstractDecimal(x, y *Decimal) *Decimal {
	return &Decimal{SubstractBig(x.Big, y.Big)}
}

// NewDecimalFromFloat64 is a constructor for Decimal out of float64
// passing through string is necessary due to differences between decimal and binary representation of float64
func NewDecimalFromFloat64(f float64) *Decimal {

	// Might want to use SetFloat here.
	d, _ := decimal.WithContext(DecimalContext).SetString(strconv.FormatFloat(f, 'f', -1, 64))
	return &Decimal{d}
}

// NewDecimalFromUsage is a constructor for Decimal out of unit represents as string
func NewDecimalFromUsage(u string) (d *Decimal, err error) {
	switch {
	// There was no duration present, equivalent of 0 decimal
	case u == EmptyString:
		d = NewDecimal(0, 0)
	//"ns", "us" (or "µs"), "ms", "s", "m", "h"
	case strings.HasSuffix(u, SSuffix),
		strings.HasSuffix(u, MSuffix),
		strings.HasSuffix(u, HSuffix):
		var tm time.Duration
		if tm, err = time.ParseDuration(u); err != nil {
			return
		}
		d = NewDecimal(int64(tm), 0)
	default:
		d, err = NewDecimalFromString(u)
	}
	return
}

// NewDecimalFromUsage is a constructor for Decimal out of unit represents as string
func NewDecimalFromUsageIgnoreErr(u string) (d *Decimal) {
	d, _ = NewDecimalFromUsage(u)
	return
}

// NewDecimal is a constructor for Decimal, following the one of decimal.Big
func NewDecimal(value int64, scale int) *Decimal {
	return &Decimal{decimal.WithContext(DecimalContext).SetMantScale(value, scale)}
}

type Decimal struct {
	*decimal.Big
}

// UnmarshalBinary implements the method for binaryUnmarshal interface for Msgpack encoding
func (d *Decimal) UnmarshalBinary(data []byte) (err error) {
	if d == nil {
		d = &Decimal{decimal.WithContext(DecimalContext)}
	}
	if d.Big == nil {
		d.Big = decimal.WithContext(DecimalContext)
	}
	return d.Big.UnmarshalText(data)
}

// MarshalBinary implements the method for binaryMarshal interface for Msgpack encoding
func (d *Decimal) MarshalBinary() ([]byte, error) {
	if d.Big == nil {
		d.Big = decimal.WithContext(DecimalContext)
	}
	return d.Big.MarshalText()
}

// UnmarshalJSON implements the method for jsonUnmarshal for JSON encoding
func (d *Decimal) UnmarshalJSON(data []byte) (err error) {
	if d == nil {
		d = &Decimal{decimal.WithContext(DecimalContext)}
	}
	if d.Big == nil {
		d.Big = decimal.WithContext(DecimalContext)
	}
	// json Unmarshal does not support NaN
	if bytes.Equal(data, []byte(DecNaN)) {
		*d = *DecimalNaN
		return
	}
	return d.Big.UnmarshalText(data)
}

// MarshalJSON implements the method for jsonMarshal for JSON encoding
func (d *Decimal) MarshalJSON() ([]byte, error) {
	if d.IsNaN(0) { // json Unmarshal does not support NaN
		return []byte(DecNaN), nil
	}
	return d.MarshalText()
}

// Clone returns a copy of the Decimal
func (d *Decimal) Clone() *Decimal {
	return &Decimal{decimal.WithContext(DecimalContext).Copy(d.Big)}
}

// Compare wraps the decimal.Big.Cmp function. It does not handle nil d2
func (d *Decimal) Compare(d2 *Decimal) int {
	if d.IsNaN(0) && !d2.IsNaN(0) {
		return -1
	}
	if !d.IsNaN(0) && d2.IsNaN(0) {
		return 1
	}
	return d.Big.Cmp(d2.Big)
}

// NewDecimalFromString converts a string to decimal
func NewDecimalFromString(value string) (*Decimal, error) {
	z, err := StringAsBig(value)
	if err != nil {
		return nil, err
	}
	return &Decimal{z}, nil
}

// NewDecimalFromStringIgnoreError same as above but ignore error( for test only)
func NewDecimalFromStringIgnoreError(v string) (d *Decimal) {
	d, _ = NewDecimalFromString(v)
	return
}

// Round rounds d down to the Context's precision and returns Decimal. The result is
// undefined if d is not finite. The result of Round will always be within the
// interval [⌊10**x⌋, d] where x = the precision of d.
func (d *Decimal) Round(rndDec int) *Decimal {
	ctx := d.Big.Context
	ctx.Precision = rndDec
	return &Decimal{ctx.Round(d.Big)}
}

// Duration returns the decimal as duration or !ok otherwise
func (d *Decimal) Duration() (dur time.Duration, ok bool) {
	var i64 int64
	if i64, ok = d.Big.Int64(); !ok {
		return
	}
	dur = time.Duration(i64)
	return
}

func CloneDecimalBig(in *decimal.Big) *decimal.Big {
	return decimal.WithContext(DecimalContext).Copy(in)
}
