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

package rates

import (
	"sort"
	"time"

	"github.com/cgrates/cgrates/engine"
)

// orderRatesOnIntervals will order the rates based on intervalStart
// there can be only one winning Rate for each interval, prioritized by the Weight
func orderRatesOnIntervals(allRts map[time.Duration][]*engine.Rate) (ordRts []*engine.Rate) {
	var idxOrdRts int
	ordRts = make([]*engine.Rate, len(allRts))
	for _, rts := range allRts {
		sort.Slice(rts, func(i, j int) bool {
			return rts[i].Weight > rts[j].Weight
		})
		ordRts[idxOrdRts] = rts[0]
		idxOrdRts++
	}
	sort.Slice(ordRts, func(i, j int) bool {
		return ordRts[i].IntervalStart < ordRts[j].IntervalStart
	})
	return
}
