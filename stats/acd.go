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

package stats

import (
	"github.com/cgrates/cgrates/engine"
)

func NewACD() (StatsMetric, error) {
	return new(ACD), nil
}

// ACD implements AverageCallDuration metric
type ACD struct{}

func (acd *ACD) GetStringValue(fmtOpts string) (val string) {
	return
}

func (acd *ACD) AddEvent(ev engine.StatsEvent) (err error) {
	return
}

func (acd *ACD) RemEvent(ev engine.StatsEvent) (err error) {
	return
}

func (acd *ACD) GetMarshaled(ms engine.Marshaler) (vals []byte, err error) {
	return
}

func (acd *ACD) SetFromMarshaled(vals []byte, ms engine.Marshaler) (err error) {
	return
}
