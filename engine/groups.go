/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012-2014 ITsysCOM

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
	"sort"
)

type GroupLink struct {
	Id     string
	Weight float64
}

type GroupLinks []*GroupLink

func (gls GroupLinks) Len() int {
	return len(gls)
}

func (gls GroupLinks) Swap(i, j int) {
	gls[i], gls[j] = gls[j], gls[i]
}

func (gls GroupLinks) Less(j, i int) bool {
	return gls[i].Weight < gls[j].Weight
}

func (gls GroupLinks) Sort() {
	sort.Sort(gls)
}
