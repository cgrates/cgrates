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
)

var ConReqs *ConcReqs

type ConcReqs struct {
	limit    int
	strategy string
	aReqs    chan struct{}
}

func NewConReqs(reqs int, strategy string) *ConcReqs {
	cR := &ConcReqs{
		limit:    reqs,
		strategy: strategy,
		aReqs:    make(chan struct{}, reqs),
	}
	for i := 0; i < reqs; i++ {
		cR.aReqs <- struct{}{}
	}
	return cR
}

var errDeny = fmt.Errorf("denying request due to maximum active requests reached")

func (cR *ConcReqs) Allocate() (err error) {
	if cR.limit == 0 {
		return
	}
	switch cR.strategy {
	case MetaBusy:
		if len(cR.aReqs) == 0 {
			return errDeny
		}
		fallthrough
	case MetaQueue:
		<-cR.aReqs // get from channel
	}
	return
}

func (cR *ConcReqs) Deallocate() {
	if cR.limit == 0 {
		return
	}
	cR.aReqs <- struct{}{}
	return
}
