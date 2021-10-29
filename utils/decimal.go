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

func DivideBig(x, y *decimal.Big) *decimal.Big {
	return new(decimal.Big).Quo(x, y)
}

func DivideBigWithReminder(x, y *decimal.Big) (q *decimal.Big, r *decimal.Big) {
	return new(decimal.Big).QuoRem(x, y, new(decimal.Big))
}

func MultiplyBig(x, y *decimal.Big) *decimal.Big {
	return new(decimal.Big).Mul(x, y)
}

func SumBig(x, y *decimal.Big) *decimal.Big {
	return new(decimal.Big).Add(x, y)
}

func SubstractBig(x, y *decimal.Big) *decimal.Big {
	return new(decimal.Big).Sub(x, y)
}

// MultiplyDecimal multiples two Decimals and returns the result
func MultiplyDecimal(x, y *Decimal) *Decimal {
	return &Decimal{new(decimal.Big).Mul(x.Big, y.Big)}
}

func SubstractDecimal(x, y *Decimal) *Decimal {
	return &Decimal{new(decimal.Big).Sub(x.Big, y.Big)}
}

// NewDecimalFromFloat64 is a constructor for Decimal out of float64
// passing through string is necessary due to differences between decimal and binary representation of float64
func NewDecimalFromFloat64(f float64) *Decimal {
	d, _ := new(decimal.Big).SetString(strconv.FormatFloat(f, 'f', -1, 64))
	return &Decimal{d}
}

// NewDecimalFromUsage is a constructor for Decimal out of unit represents as string
func NewDecimalFromUsage(u string) (d *Decimal, err error) {
	switch {
	//"ns", "us" (or "µs"), "ms", "s", "m", "h"
	case strings.HasSuffix(u, NsSuffix), strings.HasSuffix(u, UsSuffix), strings.HasSuffix(u, µSuffix), strings.HasSuffix(u, MsSuffix),
		strings.HasSuffix(u, SSuffix), strings.HasSuffix(u, MSuffix), strings.HasSuffix(u, HSuffix):
		var tm time.Duration
		if tm, err = time.ParseDuration(u); err != nil {
			return
		}
		d = NewDecimal(int64(tm), 0)
		return
	default:
		var i int64
		if i, err = strconv.ParseInt(u, 10, 64); err != nil {
			return
		}
		d = NewDecimal(i, 0)
		return
	}

}

// NewDecimal is a constructor for Decimal, following the one of decimal.Big
func NewDecimal(value int64, scale int) *Decimal {
	return &Decimal{decimal.New(value, scale)}
}

type Decimal struct {
	*decimal.Big
}

// UnmarshalBinary implements the method for binaryUnmarshal interface for Msgpack encoding
func (d *Decimal) UnmarshalBinary(data []byte) (err error) {
	if d == nil {
		d = &Decimal{new(decimal.Big)}
	}
	if d.Big == nil {
		d.Big = new(decimal.Big)
	}
	return d.Big.UnmarshalText(data)
}

// MarshalBinary implements the method for binaryMarshal interface for Msgpack encoding
func (d *Decimal) MarshalBinary() ([]byte, error) {
	if d.Big == nil {
		d.Big = new(decimal.Big)
	}
	return d.Big.MarshalText()
}

// UnmarshalJSON implements the method for jsonUnmarshal for JSON encoding
func (d *Decimal) UnmarshalJSON(data []byte) (err error) {
	if d == nil {
		d = &Decimal{new(decimal.Big)}
	}
	if d.Big == nil {
		d.Big = new(decimal.Big)
	}
	return d.Big.UnmarshalText(data)
}

// MarshalJSON implements the method for jsonMarshal for JSON encoding
func (d *Decimal) MarshalJSON() ([]byte, error) {
	x, err := d.MarshalText()
	return bytes.Trim(x, `"`), err
}

// Clone returns a copy of the Decimal
func (d *Decimal) Clone() *Decimal {
	return &Decimal{new(decimal.Big).Copy(d.Big)}
}

// Compare wraps the decimal.Big.Cmp function. It does not handle nil d2
func (d *Decimal) Compare(d2 *Decimal) int {
	return d.Big.Cmp(d2.Big)
}

// NewDecimalFromString converts a string to decimal
func NewDecimalFromString(value string) (*Decimal, error) {
	z, ok := new(decimal.Big).SetString(value)
	// verify ok and check if the value was converted successfuly
	// and the big is a valid number
	if !ok || z.IsNaN(0) {
		return nil, fmt.Errorf("can't convert <%+v> to decimal", value)
	}
	return &Decimal{z}, nil
}
