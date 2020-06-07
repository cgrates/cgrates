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

import "time"

type ChargedInterval struct {
	Increments     []*ChargingIncrement // specific increments applied to this interval
	CompressFactor int
	ecUsageIdx     *time.Duration // computed value of totalUsage at the starting of the interval
	usage          *time.Duration // cache usage computation for this interval
	cost           *float64       // cache cost calculation on this interval // #ToDo: replace here with decimal.Big
}
