/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2013 ITsysCOM

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
	"fmt"
	"time"
)

type LCR struct {
	Tenant    string
	Customer  string
	Direction string
	LCRs      []*LCREntry
}

type LCREntry struct {
	Destination    string
	TOR            string
	Strategy       string
	Suppliers      string
	ActivationTime time.Time
	Weight         float64
}

type LCRCost struct {
	Supplier string
	Cost     float64
}

func (lcr *LCR) GetId() string {
	return fmt.Sprintf("%s:%s:%s", lcr.Direction, lcr.Tenant, lcr.Customer)
}
