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

package config

import (
	"fmt"

	"github.com/cgrates/cgrates/utils"
)

type ConcReqs struct {
	aReqs    chan struct{}
	nAReqs   int
	strategy string
}

func NewConReqs(reqs int, strategy string) *ConcReqs {
	return &ConcReqs{
		aReqs:    make(chan struct{}, reqs),
		nAReqs:   reqs,
		strategy: strategy,
	}
}

func (cR *ConcReqs) VerifyAndGet() (err error) {
	if cR.nAReqs == 0 {
		return
	}
	switch cR.strategy {
	case utils.MetaBusy:
		if len(cR.aReqs) == 0 {
			return fmt.Errorf("denying request due to maximum active requests reached")
		}
		fallthrough
	case utils.MetaQueue:
		<-cR.aReqs // get from channel
	}
	return
}

func (cR *ConcReqs) Putback() {
	if cR.nAReqs == 0 {
		return
	}
	cR.aReqs <- struct{}{}
	return
}
