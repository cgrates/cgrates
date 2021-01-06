/*
Real-time Online/Offline Charging System (OerS) for Telecom & ISP environments
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

	"github.com/ericlagergren/decimal"
)

func DivideBig(x, y *decimal.Big) *decimal.Big {
	return new(decimal.Big).Quo(x, y)
}

func MultiplyBig(x, y *decimal.Big) *decimal.Big {
	return new(decimal.Big).Mul(x, y)
}

func AddBig(x, y *decimal.Big) *decimal.Big {
	return new(decimal.Big).Add(x, y)
}

func SubstractBig(x, y *decimal.Big) *decimal.Big {
	return new(decimal.Big).Sub(x, y)
}

func NewDecimalFromFloat64(f float64) (*Decimal, error) {
	d, canSet := new(decimal.Big).SetString(strconv.FormatFloat(f, 'f', -1, 64))
	if !canSet {
		return nil, fmt.Errorf("cannot convert float64 to decimal.Big")
	}
	return &Decimal{d}, nil
}

type Decimal struct {
	*decimal.Big
}

func (d *Decimal) UnmarshalBinary(data []byte) (err error) {
	if d == nil {
		d = &Decimal{new(decimal.Big)}
	}
	if d.Big == nil {
		d.Big = new(decimal.Big)
	}
	return d.Big.UnmarshalText(data)
}

func (d *Decimal) MarshalBinary() ([]byte, error) {
	if d.Big == nil {
		d.Big = new(decimal.Big)
	}
	return d.Big.MarshalText()
}

func (d *Decimal) UnmarshalJSON(data []byte) (err error) {
	return d.UnmarshalBinary(data)
}

func (d *Decimal) MarshalJSON() ([]byte, error) {
	x, err := d.MarshalText()
	return bytes.Trim(x, `"`), err
}

func (d *Decimal) Copy(d2 *Decimal) *Decimal {
	d.Big = new(decimal.Big)
	d.Big.Copy(d2.Big)
	return d
}
