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
	"github.com/ericlagergren/decimal"
)

func NewDecimalFromFloat64(x float64) *Decimal {
	return &Decimal{new(decimal.Big).SetFloat64(x)}
}

func NewDecimalFromUint64(x uint64) *Decimal {
	return &Decimal{new(decimal.Big).SetUint64(x)}
}

func NewDecimal() *Decimal {
	return &Decimal{new(decimal.Big)}
}

// Decimal extends the decimal.Big with additional methods
type Decimal struct {
	*decimal.Big
}

func (d *Decimal) Float64() (f float64) {
	f, _ = d.Big.Float64()
	return
}

func (d *Decimal) MarshalJSON() ([]byte, error) {
	return d.Big.MarshalText()
}

func (d *Decimal) UnmarshalJSON(data []byte) error {
	return d.Big.UnmarshalJSON(data)
}

func (d *Decimal) Divide(x, y *Decimal) *Decimal {
	d.Big.Quo(x.Big, y.Big)
	return d
}

func (d *Decimal) Multiply(x, y *Decimal) *Decimal {
	d.Big.Mul(x.Big, y.Big)
	return d
}

func (d *Decimal) Add(x, y *Decimal) *Decimal {
	d.Big.Add(x.Big, y.Big)
	return d
}

func (d *Decimal) Compare(y *Decimal) int {
	return d.Big.Cmp(y.Big)
}
